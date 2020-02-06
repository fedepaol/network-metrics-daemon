package controller

import (
	"strings"
	"testing"
	"time"

	promtestutil "github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/openshift/network-metrics/pkg/podmetrics"
	"github.com/openshift/network-metrics/pkg/podnetwork"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	k8sfake "k8s.io/client-go/kubernetes/fake"
)

const metadata = `
	# HELP network_attachment_definition_per_pod Metric to identify clusters with network attachment definition enabled instances.
	# TYPE network_attachment_definition_per_pod gauge
	`

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t               *testing.T
	kubeclient      *k8sfake.Clientset
	podsLister      []*v1.Pod
	kubeobjects     []runtime.Object
	expectedMetrics string
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.kubeobjects = []runtime.Object{}
	return f
}

func newPod(name, namespace string, networkAnnotation string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				podnetwork.Status: networkAnnotation,
			},
		},
	}
}

func (f *fixture) newController() (*Controller, kubeinformers.SharedInformerFactory) {
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())
	c := New(f.kubeclient, k8sI.Core().V1().Pods())

	c.podsSynced = alwaysReady

	for _, p := range f.podsLister {
		k8sI.Core().V1().Pods().Informer().GetIndexer().Add(p)
	}

	return c, k8sI
}

func (f *fixture) run(podName string) {
	f.runController(podName, true, false)
}

func (f *fixture) runController(podName string, startInformers bool, expectError bool) {
	c, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.podHandler(podName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing foo: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing foo, got nil")
	}

}

func TestPublishesMetric(t *testing.T) {
	f := newFixture(t)
	pod := newPod("podname", "namespace", `[{
		"name": "kindnet",
		"interface": "eth0",
		"ips": [
			"10.244.0.10"
		],
		"mac": "4a:e9:0b:e2:63:67",
		"default": true,
		"dns": {}
	}]`)
	f.podsLister = append(f.podsLister, pod)
	f.kubeobjects = append(f.kubeobjects, pod)
	f.expectedMetrics = `
	network_attachment_definition_per_pod{interface="eth0",nad="kindnet",namespace="namespace",pod="podname"} 0
	`
	f.run(getKey(pod, t))

	err := promtestutil.CollectAndCompare(podmetrics.NetAttachDefPerPod, strings.NewReader(metadata+f.expectedMetrics))
	if err != nil {
		t.Error("Failed to collect metrics", err)
	}
	podmetrics.NetAttachDefPerPod.Reset()
}

func getKey(pod *v1.Pod, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(pod)
	if err != nil {
		t.Errorf("Unexpected error getting key for foo %v: %v", pod.Name, err)
		return ""
	}
	return key
}
