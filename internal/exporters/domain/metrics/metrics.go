package metrics

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"

	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/gpus"
	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/pods"
	"github.com/prometheus/client_golang/prometheus"
)

// CustomMetric defines data required to build a metric.
type CustomMetric struct {
	Name      string
	Namespace string
	Subsystem string
	HelpText  string
	Type      prometheus.ValueType
	// Labels The metric's variable label dimensions.
	Labels []string
	// Divide indicates that the metric value should be divided by the given divisor.
	Divide  bool
	Divisor float64
}

// AMDMetrics set of prometheus metrics to be collected from amd resources.
type AMDMetrics struct {
	DataDesc       *CustomMetric
	CoreEnergy     *CustomMetric
	SocketEnergy   *CustomMetric
	BoostLimit     *CustomMetric
	SocketPower    *CustomMetric
	PowerLimit     *CustomMetric
	ProchotStatus  *CustomMetric
	Sockets        *CustomMetric
	Threads        *CustomMetric
	ThreadsPerCore *CustomMetric
	NumGPUs        *CustomMetric
	GPUDevID       *CustomMetric
	GPUPowerCap    *CustomMetric
	GPUPower       *CustomMetric
	GPUTemperature *CustomMetric
	GPUSCLK        *CustomMetric
	GPUMCLK        *CustomMetric
	GPUUsage       *CustomMetric
	GPUMemoryUsage *CustomMetric
	CardsInfo      [gpus.MaxNumGPUDevices]gpus.Card
	K8SResources   *pods.Resources
	Data           gpus.AMDParamsHandler // This is the Scan() function handle
	logger         *slog.Logger
	withKubernetes bool
}

// Setup contains objects required to process metrics.
type Setup struct {
	AMDParamsHandler gpus.AMDParamsHandler
	Logger           *slog.Logger
	WithKubernetes   bool
}

// metric labels.
const (
	podNameLabel       string = "exported_pod"
	namespaceNameLabel string = "exported_namespace"
	containerNameLabel string = "exported_container"
	nodeNameLabel      string = "exported_node"
	productNameLabel   string = "productname"
	deviceNameLabel    string = "device"

	deviceIDPrefix           string = "amd"
	amdMetricHelpTextDefault string = "AMD Params" // The metric's help text.
)

// metric common values.
const amdNamespace string = "amd"

// labelPrefix in case you want prefix your labels with "label" word.
const labelPrefixPattern = "label_%s"

// Prometheus label naming convention regex.
var validLabelRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Replace invalid characters with underscores in labels.
var replacer = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// NewAMDMetrics creates AMD metrics based on given handler and a flag to indicate
// if k8s resources are needed.
func NewAMDMetrics(settings *Setup) *AMDMetrics {
	newAMDMetrics := &AMDMetrics{
		withKubernetes: settings.WithKubernetes,
		Data:           settings.AMDParamsHandler,
		logger:         settings.Logger,
	}

	return newAMDMetrics.initializeMetrics()
}

// initializeMetrics initializes prometheus metric descriptions.
func (a *AMDMetrics) initializeMetrics() *AMDMetrics {
	a.DataDesc = &CustomMetric{
		Name:     "amd_data",
		HelpText: amdMetricHelpTextDefault,
		Labels:   []string{"socket"},
	}
	a.CoreEnergy = a.newAMDCounterMetric("core_energy", "thread")
	a.SocketEnergy = a.newAMDCounterMetric("socket_energy", "socket")
	a.BoostLimit = a.newAMDGaugeMetric("boost_limit", "thread")
	a.SocketPower = a.newAMDGaugeMetric("socket_power", "socket")
	a.PowerLimit = a.newAMDGaugeMetric("power_limit", "power_limit")
	a.ProchotStatus = a.newAMDGaugeMetric("prochot_status", "prochot_status")
	a.Sockets = a.newAMDGaugeMetric("num_sockets", "num_sockets")
	a.Threads = a.newAMDGaugeMetric("num_threads", "num_threads")
	a.ThreadsPerCore = a.newAMDGaugeMetric("num_threads_per_core", "num_threads_per_core")
	a.NumGPUs = a.newAMDGaugeMetric("num_gpus", "num_gpus")
	a.GPUDevID = a.newAMDGPUGaugeMetric("gpu_dev_id")
	a.GPUPowerCap = a.newAMDGPUGaugeMetric("gpu_power_cap").
		WithDivisor(1e6)
	a.GPUPower = a.newAMDGPUCounterMetric("gpu_power").
		WithDivisor(1e6)
	a.GPUTemperature = a.newAMDGPUGaugeMetric("gpu_current_temperature").
		WithDivisor(1e3)
	a.GPUSCLK = a.newAMDGPUGaugeMetric("gpu_SCLK").
		WithDivisor(1e6)
	a.GPUMCLK = a.newAMDGPUGaugeMetric("gpu_MCLK").
		WithDivisor(1e6)
	a.GPUUsage = a.newAMDGPUGaugeMetric("gpu_use_percent")
	a.GPUMemoryUsage = a.newAMDGPUGaugeMetric("gpu_memory_use_percent")

	return a
}

func (a *AMDMetrics) newAMDGaugeMetric(name string, label ...string) *CustomMetric {
	return a.newAMDMetric(name, prometheus.GaugeValue, label...)
}

func (a *AMDMetrics) newAMDCounterMetric(name string, label ...string) *CustomMetric {
	return a.newAMDMetric(name, prometheus.CounterValue, label...)
}

func (a *AMDMetrics) newAMDGPUGaugeMetric(name string) *CustomMetric {
	return a.newAMDGPUMetric(name, prometheus.GaugeValue)
}

func (a *AMDMetrics) newAMDGPUCounterMetric(name string) *CustomMetric {
	return a.newAMDGPUMetric(name, prometheus.CounterValue)
}

func (a *AMDMetrics) newAMDGPUMetric(name string, mType prometheus.ValueType) *CustomMetric {
	return a.newAMDMetric(name, mType, name, productNameLabel, deviceNameLabel)
}

func (a *AMDMetrics) newAMDMetric(name string, mType prometheus.ValueType, label ...string) *CustomMetric {
	if a.withKubernetes {
		label = append(label, nodeNameLabel)
	}

	return &CustomMetric{
		Name:      name,
		Namespace: amdNamespace,             // metric namespace
		HelpText:  amdMetricHelpTextDefault, // The metric's help text.
		Labels:    label,                    // The metric's variable label dimensions.
		Type:      mType,
	}
}

// WithDivisor enable dividing metric value by given divisor.
func (c *CustomMetric) WithDivisor(divisor float64) *CustomMetric {
	c.Divide = true
	c.Divisor = divisor

	return c
}

// buildPrometheusMetric builds prometheus metric based on given value and metric configuration.
func (c *CustomMetric) buildPrometheusMetric(value float64, labelValues ...string) prometheus.Metric {
	value = c.transformValue(value)

	return prometheus.MustNewConstMetric(
		c.NewDesc(),
		c.Type,
		value,
		labelValues...,
	)
}

// transformValue transform metric value to the desired format.
func (c *CustomMetric) transformValue(value float64) float64 {
	if c.divisionRequired() {
		value /= c.Divisor
	}

	return value
}

// divisionRequired checks if metric value should be divided.
func (c *CustomMetric) divisionRequired() bool {
	return c.Divide && c.Divisor > 0
}

// NewDesc allocates and initializes a new prometheus Desc.
func (c *CustomMetric) NewDesc(additionalLabels ...string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(c.Namespace, c.Subsystem, c.Name),
		c.HelpText,                            // The metric's help text.
		append(c.Labels, additionalLabels...), // The metric's variable label dimensions.
		nil,                                   // The metric's constant label dimensions.
	)
}

// CollectAndBuildMetrics scans amd data and build a collection of metrics.
func (a *AMDMetrics) CollectAndBuildMetrics() []prometheus.Metric {
	data := a.Data() // Scan AMD metrics

	a.logger.Debug("scanning", slog.Any("amd-params", data))

	metrics := make([]prometheus.Metric, 0)

	metrics = append(metrics, a.buildMetrics(data.CoreEnergy[:data.Threads], data.Threads, a.CoreEnergy)...)
	metrics = append(metrics, a.buildMetrics(data.CoreBoost[:data.Threads], data.Threads, a.BoostLimit)...)
	metrics = append(metrics, a.buildMetrics(data.SocketEnergy[:data.Sockets], data.Sockets, a.SocketEnergy)...)
	metrics = append(metrics, a.buildMetrics(data.SocketPower[:data.Sockets], data.Sockets, a.SocketPower)...)
	metrics = append(metrics, a.buildMetrics(data.PowerLimit[:data.Sockets], data.Sockets, a.PowerLimit)...)
	metrics = append(metrics, a.buildMetrics(data.ProchotStatus[:data.Sockets], data.Sockets, a.ProchotStatus)...)

	// GPU metrics
	metrics = append(metrics, a.buildGPUMetrics(data.GPUDevID[:data.NumGPUs], data.NumGPUs, a.GPUDevID)...)
	metrics = append(metrics, a.buildGPUMetrics(data.GPUPowerCap[:data.NumGPUs], data.NumGPUs, a.GPUPowerCap)...)
	metrics = append(metrics, a.buildGPUMetrics(data.GPUPower[:data.NumGPUs], data.NumGPUs, a.GPUPower)...)
	metrics = append(metrics, a.buildGPUMetrics(data.GPUTemperature[:data.NumGPUs], data.NumGPUs, a.GPUTemperature)...)
	metrics = append(metrics, a.buildGPUMetrics(data.GPUSCLK[:data.NumGPUs], data.NumGPUs, a.GPUSCLK)...)
	metrics = append(metrics, a.buildGPUMetrics(data.GPUMCLK[:data.NumGPUs], data.NumGPUs, a.GPUMCLK)...)
	metrics = append(metrics, a.buildGPUMetrics(data.GPUUsage[:data.NumGPUs], data.NumGPUs, a.GPUUsage)...)
	metrics = append(metrics, a.buildGPUMetrics(data.GPUMemoryUsage[:data.NumGPUs], data.NumGPUs, a.GPUMemoryUsage)...)

	metrics = append(metrics, a.resourceGroupMetrics(&data)...)

	return metrics
}

// buildMetrics builds prometheus metric based on given amd metric.
func (a *AMDMetrics) buildMetrics(
	data []float64,
	attrValue uint,
	metric *CustomMetric,
) []prometheus.Metric {
	if attrValue == 0 {
		return nil
	}

	metrics := make([]prometheus.Metric, attrValue)

	for i := range data {
		if a.withKubernetes {
			metrics[i] = metric.buildPrometheusMetric(data[i], strconv.Itoa(i), a.K8SResources.NodeName)

			continue
		}

		metrics[i] = metric.buildPrometheusMetric(data[i], strconv.Itoa(i))
	}

	return metrics
}

// buildGPUMetrics builds prometheus metric based on given amd gpu metric.
func (a *AMDMetrics) buildGPUMetrics(
	data []float64,
	attrValue uint,
	metric *CustomMetric,
) []prometheus.Metric {
	if attrValue == 0 {
		return nil
	}

	var metrics []prometheus.Metric

	for i := range data {
		metrics = append(metrics, a.newMetricWithResources(metric, data[i], i)...)
	}

	return metrics
}

// resourceGroupMetrics build global metrics such as sockets, thread and number of GPUs.
func (a *AMDMetrics) resourceGroupMetrics(params *gpus.AMDParams) []prometheus.Metric {
	return []prometheus.Metric{
		a.Sockets.buildPrometheusMetric(float64(params.Sockets), "", a.K8SResources.NodeName),
		a.Threads.buildPrometheusMetric(float64(params.Threads), "", a.K8SResources.NodeName),
		a.ThreadsPerCore.buildPrometheusMetric(float64(params.ThreadsPerCore), "", a.K8SResources.NodeName),
		a.NumGPUs.buildPrometheusMetric(float64(params.NumGPUs), "", a.K8SResources.NodeName),
	}
}

// newMetricWithResources map given GPU card metric with pod
// using it. If there is no any pod using this card then
// a prometheus metric is created with pod labels.
func (a *AMDMetrics) newMetricWithResources(
	metric *CustomMetric,
	value float64, cardIndex int,
) []prometheus.Metric {
	labelValues := a.commonGPULabelValues(cardIndex)

	if !a.withKubernetes {
		return []prometheus.Metric{
			metric.buildPrometheusMetric(value, labelValues...),
		}
	}

	podsInfo, exist := a.K8SResources.Pods[a.CardsInfo[cardIndex].PCIBus]
	if !exist {
		return []prometheus.Metric{
			// metric.buildPrometheusMetric(value, labelValues...),
			prometheus.MustNewConstMetric(
				metric.NewDesc(),
				metric.Type,
				metric.transformValue(value),
				append(labelValues, a.K8SResources.NodeName)...),
		}
	}

	metrics := make([]prometheus.Metric, 0, len(podsInfo))

	for _, p := range podsInfo {
		newLabels, newLabelValues := buildK8SPodLabelValues(p, labelValues)

		metrics = append(metrics,
			prometheus.MustNewConstMetric(
				metric.NewDesc(newLabels...),
				metric.Type,
				metric.transformValue(value),
				newLabelValues...),
		)
	}

	return metrics
}

// commonGPULabelValues returns common GPU labels.
func (a *AMDMetrics) commonGPULabelValues(cardIndex int) []string {
	return []string{
		strconv.Itoa(cardIndex),
		a.CardsInfo[cardIndex].Cardseries,
		buildDeviceLabelValue(cardIndex),
	}
}

// buildDeviceIDLabelValue build device label name.
func buildDeviceLabelValue(cardIndex int) string {
	return fmt.Sprintf("%s%d", deviceIDPrefix, cardIndex)
}

// buildK8SPodLabelValues return 2 slices of pod labels and its respective values.
func buildK8SPodLabelValues(pod pods.PodInfo, existingLabelValues []string) ([]string, []string) {
	labels := k8sVariableLabels()
	values := slices.Concat(
		existingLabelValues,
		[]string{
			pod.NodeName,
			pod.Name,
			pod.Container,
			pod.Namespace,
		},
	)

	// adding existing labels within pod
	for _, key := range pod.Labels.SortKeys() {
		labels = append(labels, formatLabel(key, true))
		values = append(values, pod.Labels[key])
	}

	return labels, values
}

// k8sVariableLabels return list of kubernetes labels required in metrics.
func k8sVariableLabels() []string {
	return []string{podNameLabel, containerNameLabel, namespaceNameLabel}
}

// formatLabel format label to follow prometheus conventions.
func formatLabel(label string, withPrefix bool) string {
	// Prometheus label naming convention regex.
	if validLabelRegex.MatchString(label) {
		return label // Label is already valid.
	}

	// Replace invalid characters with underscores.
	sanitized := replacer.ReplaceAllString(label, "_")

	// Ensure the label starts with a letter or underscore.
	if sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "_" + sanitized
	}

	if withPrefix {
		sanitized = fmt.Sprintf(labelPrefixPattern, sanitized)
	}

	return sanitized
}
