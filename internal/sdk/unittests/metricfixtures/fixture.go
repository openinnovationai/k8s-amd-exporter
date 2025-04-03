package metricfixtures

import "github.com/prometheus/client_golang/prometheus"

func ConstCounterMetric(name string, value float64, labels, labelValues []string) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName("amd", "", name),
			"AMD Params", // The metric's help text.
			labels,       // The metric's variable label dimensions.
			nil,          // The metric's constant label dimensions.
		),
		prometheus.CounterValue,
		value,
		labelValues...,
	)
}

func ConstGaugeMetric(name string, value float64, labels, labelValues []string) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName("amd", "", name),
			"AMD Params", // The metric's help text.
			labels,       // The metric's variable label dimensions.
			nil,          // The metric's constant label dimensions.
		),
		prometheus.GaugeValue,
		value,
		labelValues...,
	)
}

func GPULabels(label string, customLabels ...string) []string {
	base := []string{label, "productname", "device", "exported_node", "exported_pod", "exported_container", "exported_namespace"}
	base = append(base, customLabels...)

	return base
}
