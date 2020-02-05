package podmetrics

import (
	"sync"

	"github.com/openshift/network-metrics/pkg/podnetwork"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricStoreInitSize int = 330
	initialMetricsCount int = 0
	metricsIncVal       int = 1
)

type podKey struct {
	name      string
	namespace string
}

var podNetworks = make(map[podKey][]podnetwork.Network)
var mtx sync.Mutex

var (
	// NetAttachDefPerPod represent the network attachment definitions bound to a given
	// pod
	NetAttachDefPerPod = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "network_attachment_definition_per_pod",
			Help: "Metric to identify clusters with network attachment definition enabled instances.",
		}, []string{"pod",
			"namespace",
			"interface",
			"nad"})
)

//UpdateForPod ...
func UpdateForPod(podName, namespace string, networks []podnetwork.Network) {
	for _, n := range networks {
		labels := prometheus.Labels{
			"pod":       podName,
			"namespace": namespace,
			"interface": n.Interface,
			"nad":       n.NetworkName,
		}
		NetAttachDefPerPod.With(labels).Add(0)
	}
	mtx.Lock()
	defer mtx.Unlock()
	podNetworks[podKey{podName, namespace}] = networks
}

func DeleteAllForPod(podName, namespace string) {
	mtx.Lock()
	defer mtx.Unlock()
	nets, ok := podNetworks[podKey{podName, namespace}]
	if !ok {
		return
	}

	delete(podNetworks, podKey{podName, namespace})

	for _, n := range nets {
		labels := prometheus.Labels{
			"pod":       podName,
			"namespace": namespace,
			"interface": n.Interface,
			"nad":       n.NetworkName,
		}
		NetAttachDefPerPod.Delete(labels)
	}
}
