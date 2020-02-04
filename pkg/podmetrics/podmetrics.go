package podmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricStoreInitSize int = 330
	initialMetricsCount int = 0
	metricsIncVal       int = 1
)

// Action represents the action we want to do
// with a given pod's network metric.
type Action int

const (
	Adding Action = iota
	Deleting
)

var (
	// NetAttachDefPerPod represent the network attachment definitions bound to a given
	// pod
	NetAttachDefPerPod = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "network_attachment_definition_enabled_instance_up",
			Help: "Metric to identify clusters with network attachment definition enabled instances.",
		}, []string{"pod",
			"namespace",
			"interface",
			"nad"})
)

//UpdateNetAttachDefInstanceMetrics ...
func UpdateNetAttachDefInstanceMetrics(podName, namespace, nicName, networkName string, action Action) {
	labels := prometheus.Labels{
		"pod":       podName,
		"namespace": namespace,
		"interface": nicName,
		"nad":       networkName,
	}

	if action == Adding {
		NetAttachDefPerPod.With(labels).Add(0)
	} else {
		NetAttachDefPerPod.Delete(labels)
	}
}
