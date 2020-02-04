package podmetrics_test

import (
	"strings"
	"testing"

	"github.com/openshift/network-metrics/pkg/podmetrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPodMetrics(t *testing.T) {
	podmetrics.UpdateNetAttachDefInstanceMetrics("pod", "namespace", "eth0", "firstNAD", podmetrics.Adding)
	podmetrics.UpdateNetAttachDefInstanceMetrics("pod", "namespace", "eth1", "firstNAD", podmetrics.Adding)

	err := testutil.CollectAndCompare(podmetrics.NetAttachDefPerPod, strings.NewReader("hello"))
	if err != nil {
		t.Error("Failed to collect metrics", err)
	}
}
