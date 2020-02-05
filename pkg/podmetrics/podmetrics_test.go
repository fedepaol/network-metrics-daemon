package podmetrics_test

import (
	"strings"
	"testing"

	"github.com/openshift/network-metrics/pkg/podmetrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

var podMetricsTests = []struct {
	testName        string
	setMetrics      func()
	expectedMetrics string
}{
	{
		"twonetworks same nad",
		func() {
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname", "namespacename", "eth0", "firstNAD", podmetrics.Adding)
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname", "namespacename", "eth1", "firstNAD", podmetrics.Adding)
		},
		`
			network_attachment_definition_per_pod{interface="eth0",nad="firstNAD",namespace="namespacename",pod="podname"} 0
			network_attachment_definition_per_pod{interface="eth1",nad="firstNAD",namespace="namespacename",pod="podname"} 0
			`,
	},
	{
		"twonetworks different nad",
		func() {
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname", "namespacename", "eth0", "firstNAD", podmetrics.Adding)
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname", "namespacename", "eth1", "secondNAD", podmetrics.Adding)
		},
		`
			network_attachment_definition_per_pod{interface="eth0",nad="firstNAD",namespace="namespacename",pod="podname"} 0
			network_attachment_definition_per_pod{interface="eth1",nad="secondNAD",namespace="namespacename",pod="podname"} 0
			`,
	},
	{
		"add and delete",
		func() {
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname", "namespacename", "eth0", "firstNAD", podmetrics.Adding)
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname", "namespacename", "eth0", "firstNAD", podmetrics.Deleting)
		},
		`
		`,
	},
	{
		"two pods and delete one",
		func() {
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname1", "namespacename", "eth0", "firstNAD", podmetrics.Adding)
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname2", "namespacename", "eth0", "firstNAD", podmetrics.Adding)
			podmetrics.UpdateNetAttachDefInstanceMetrics("podname1", "namespacename", "eth0", "firstNAD", podmetrics.Deleting)

		},
		`
			network_attachment_definition_per_pod{interface="eth0",nad="firstNAD",namespace="namespacename",pod="podname2"} 0

		`,
	},
}

func TestPodMetrics(t *testing.T) {

	const metadata = `
	# HELP network_attachment_definition_per_pod Metric to identify clusters with network attachment definition enabled instances.
	# TYPE network_attachment_definition_per_pod gauge
	`

	for _, tst := range podMetricsTests {
		tst.setMetrics()
		err := testutil.CollectAndCompare(podmetrics.NetAttachDefPerPod, strings.NewReader(metadata+tst.expectedMetrics))
		if err != nil {
			t.Error("Failed to collect metrics", tst.testName, err)
		}
		podmetrics.NetAttachDefPerPod.Reset()
	}

}
