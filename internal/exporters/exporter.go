package exporters

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/gpus"
	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/metrics"
	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/pods"
	"github.com/openinnovationai/k8s-amd-exporter/internal/kubernetes"
	"github.com/prometheus/client_golang/prometheus"
)

type Setup struct {
	K8SClient      *kubernetes.Client
	CardsInfo      [gpus.MaxNumGPUDevices]gpus.Card
	Logger         *slog.Logger
	GetMetricsFunc gpus.AMDParamsHandler
	// list of custom labels required for pods.
	OIPLabels      []string
	WithKubernetes bool
}

// Exporter implements logic about scanning metrics from environment
// and from applications running within the gpu environment.
type Exporter struct {
	k8sClient      *kubernetes.Client
	cardsInfo      [gpus.MaxNumGPUDevices]gpus.Card
	getMetricsFunc gpus.AMDParamsHandler
	amdMetrics     *metrics.AMDMetrics
	oipLabels      []string
	withKubernetes bool
	logger         *slog.Logger
}

var gkeMigDeviceIDRegex = regexp.MustCompile(`^amd([0-9]+)/gi([0-9]+)$`)

func NewExporter(settings *Setup) *Exporter {
	newScanner := Exporter{
		k8sClient:      settings.K8SClient,
		logger:         settings.Logger,
		cardsInfo:      settings.CardsInfo,
		getMetricsFunc: settings.GetMetricsFunc,
		withKubernetes: settings.WithKubernetes,
		oipLabels:      settings.OIPLabels,
	}

	newScanner.makeCollector()

	return &newScanner
}

func (e *Exporter) makeCollector() {
	settings := metrics.Setup{
		AMDParamsHandler: e.getMetricsFunc,
		WithKubernetes:   e.withKubernetes,
		Logger:           e.logger,
	}
	e.amdMetrics = metrics.NewAMDMetrics(&settings)
	e.amdMetrics.CardsInfo = e.cardsInfo
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent. The sent descriptors fulfill the
// consistency and uniqueness requirements described in the Desc
// documentation.
func (e *Exporter) Describe(descStream chan<- *prometheus.Desc) {
	descStream <- e.amdMetrics.DataDesc.NewDesc()
}

// Collect is called by the Prometheus registry when collecting
// metrics.
func (e *Exporter) Collect(metricStream chan<- prometheus.Metric) {
	e.logger.Debug("collecting metrics")

	k8sResources, err := e.scanK8SResources(context.TODO())
	if err != nil {
		e.logger.Error("scanning k8s resources", slog.String("error", err.Error()))
		e.logger.Info("will continue collecting without k8s resources")

		k8sResources = pods.NewResources()
	}

	e.amdMetrics.K8SResources = k8sResources

	metrics := e.amdMetrics.CollectAndBuildMetrics()

	for i := range metrics {
		metricStream <- metrics[i]
	}
}

// scanK8SResources scans k8s resources in order to map pods with gpu metrics.
func (e *Exporter) scanK8SResources(ctx context.Context) (*pods.Resources, error) {
	if !e.withKubernetes {
		return pods.NewResources(), nil
	}
	// Get apps using GPUs
	apps, err := e.k8sClient.GetPodsUsingDevices(ctx)
	if err != nil {
		slog.Error("getting pods using gpu devices", slog.String("error", err.Error()))

		return nil, fmt.Errorf("unable to get pods using gpu devices: %w", err)
	}

	// get required labels found in pods.
	existingLabels, err := e.k8sClient.GetPodsLabels(ctx, apps.Pods(), e.oipLabels)
	if err != nil {
		slog.Error("getting pod labels", slog.String("error", err.Error()))
	}

	for appKey, app := range apps {
		labels, ok := existingLabels[app.NamespacedName()]
		if !ok {
			continue
		}

		app.Labels = labels
		apps[appKey] = app
	}

	deviceToPodMap := make(map[string][]pods.PodInfo)

	for deviceID, podInfo := range apps {
		additionalPodIDs := calculateAdditionalDeviceIDs(deviceID)
		for _, devicePODId := range additionalPodIDs {
			deviceToPodMap[devicePODId] = append(deviceToPodMap[devicePODId], podInfo)
		}
		// Default mapping between deviceID and pod information
		deviceToPodMap[deviceID] = append(deviceToPodMap[deviceID], podInfo)
	}

	result := pods.Resources{
		Pods:     deviceToPodMap,
		NodeName: e.k8sClient.NodeName(),
	}

	return &result, nil
}

// calculateAdditionalDeviceIDs calculate other possible device ids for pods based
// on the device id returned by the kubelete pod resources api.
func calculateAdditionalDeviceIDs(deviceID string) []string {
	var result []string

	if gkeMigDeviceIDMatches := gkeMigDeviceIDRegex.FindStringSubmatch(deviceID); gkeMigDeviceIDMatches != nil {
		var gpuIndex string

		var gpuInstanceID string

		for groupIdx, group := range gkeMigDeviceIDMatches {
			switch groupIdx {
			case 1:
				gpuIndex = group
			case 2:
				gpuInstanceID = group
			}
		}

		giIdentifier := fmt.Sprintf("%s-%s", gpuIndex, gpuInstanceID)

		result = append(result, giIdentifier)

		return result
	}

	if strings.Contains(deviceID, gpus.GKEVirtualGPUDeviceIDSeparator) {
		result = append(result, strings.Split(deviceID, gpus.GKEVirtualGPUDeviceIDSeparator)[0])

		return result
	}

	if strings.Contains(deviceID, gpus.AMDVirtualGPUDeviceIDSeparator) {
		result = append(result, strings.Split(deviceID, gpus.AMDVirtualGPUDeviceIDSeparator)[0])

		return result
	}

	if strings.Contains(deviceID, ":") {
		gpuInstanceID := strings.Split(deviceID, ":")[0]
		result = append(result, gpuInstanceID)

		return result
	}

	return result
}
