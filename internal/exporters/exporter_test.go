package exporters_test

import (
	"testing"

	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters"
	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/gpus"
	"github.com/openinnovationai/k8s-amd-exporter/internal/sdk/unittests/fakekubelet"
	"github.com/openinnovationai/k8s-amd-exporter/internal/sdk/unittests/k8sfixtures"
	"github.com/openinnovationai/k8s-amd-exporter/internal/sdk/unittests/metricfixtures"
	"github.com/openinnovationai/k8s-amd-exporter/internal/sdk/unittests/testlogs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCollectWithoutAdditionalLabels(t *testing.T) {
	t.Parallel()

	k8sClient := fakekubelet.New(t,
		fakekubelet.WithPodResources(k8sfixtures.NewPodResourcesFixture(t)),
		fakekubelet.WithAMDCustomResourceNames([]string{"amd-custom-resource-name"}),
		fakekubelet.WithNodeName("node-1"),
		fakekubelet.WithClientSet(
			fake.NewClientset(
				k8sfixtures.ExistingPodsWithoutLabelsFixture(t)...),
		),
	)

	cardsInfo := [24]gpus.Card{
		0: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:b3:00.0",
			CardGUID:   "63755",
		},
		1: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:8e:00.0",
			CardGUID:   "63756",
		},
		2: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:34:00.0",
			CardGUID:   "63757",
		},
	}

	logger := testlogs.NewLogger()

	settings := exporters.Setup{
		K8SClient:      k8sClient,
		CardsInfo:      cardsInfo,
		Logger:         logger,
		WithKubernetes: true,
		GetMetricsFunc: makeAMDDataFuncFixture(t),
	}

	metricStream := make(chan prometheus.Metric)

	exporter := exporters.NewExporter(&settings)

	want := []prometheus.Metric{
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, []string{"gpu_dev_id", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1"}),

		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, []string{"gpu_power_cap", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1"}),

		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, []string{"gpu_power", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1"}),

		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, []string{"gpu_current_temperature", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1"}),

		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, []string{"gpu_SCLK", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1"}),

		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, []string{"gpu_MCLK", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1"}),

		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, []string{"gpu_use_percent", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1"}),

		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, []string{"gpu_memory_use_percent", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1"}),

		metricfixtures.ConstGaugeMetric("num_sockets", 0, []string{"num_sockets", "exported_node"}, []string{"", "node-1"}),
		metricfixtures.ConstGaugeMetric("num_threads", 0, []string{"num_threads", "exported_node"}, []string{"", "node-1"}),
		metricfixtures.ConstGaugeMetric("num_threads_per_core", 0, []string{"num_threads_per_core", "exported_node"}, []string{"", "node-1"}),
		metricfixtures.ConstGaugeMetric("num_gpus", 3, []string{"num_gpus", "exported_node"}, []string{"", "node-1"}),
	}

	// When
	go func() {
		defer close(metricStream)
		exporter.Collect(metricStream)
	}()

	// Then
	var got []prometheus.Metric
	for metric := range metricStream {
		got = append(got, metric)
	}

	assert.Equal(t, want, got)
}

func TestCollectWithAdditionalLabels(t *testing.T) {
	t.Parallel()

	k8sClient := fakekubelet.New(t,
		fakekubelet.WithPodResources(k8sfixtures.NewPodResourcesFixture(t)),
		fakekubelet.WithAMDCustomResourceNames([]string{"amd-custom-resource-name"}),
		fakekubelet.WithClientSet(
			fake.NewClientset(
				k8sfixtures.ExistingPodsWithLabelsFixture(t)...),
		),
	)

	cardsInfo := [24]gpus.Card{
		0: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:b3:00.0",
			CardGUID:   "63755",
		},
		1: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:8e:00.0",
			CardGUID:   "63756",
		},
		2: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:34:00.0",
			CardGUID:   "63757",
		},
	}

	logger := testlogs.NewLogger()

	settings := exporters.Setup{
		K8SClient:      k8sClient,
		CardsInfo:      cardsInfo,
		Logger:         logger,
		WithKubernetes: true,
		GetMetricsFunc: makeAMDDataFuncFixture(t),
		OIPLabels:      []string{"label_1", "label_2"},
	}

	metricStream := make(chan prometheus.Metric)

	exporter := exporters.NewExporter(&settings)

	want := []prometheus.Metric{
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "", "pod-c", "container-1", "team-2", "value-i", "value-ii"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, []string{"gpu_dev_id", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", ""}),

		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "", "pod-c", "container-1", "team-2", "value-i", "value-ii"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, []string{"gpu_power_cap", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", ""}),

		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "", "pod-c", "container-1", "team-2", "value-i", "value-ii"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, []string{"gpu_power", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", ""}),

		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "", "pod-c", "container-1", "team-2", "value-i", "value-ii"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, []string{"gpu_current_temperature", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", ""}),

		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "", "pod-c", "container-1", "team-2", "value-i", "value-ii"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, []string{"gpu_SCLK", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", ""}),

		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "", "pod-c", "container-1", "team-2", "value-i", "value-ii"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, []string{"gpu_MCLK", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", ""}),

		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "", "pod-c", "container-1", "team-2", "value-i", "value-ii"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, []string{"gpu_use_percent", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", ""}),

		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "", "pod-c", "container-1", "team-2", "value-i", "value-ii"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, []string{"gpu_memory_use_percent", "productname", "device", "exported_node"}, []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", ""}),

		metricfixtures.ConstGaugeMetric("num_sockets", 0, []string{"num_sockets", "exported_node"}, []string{"", ""}),
		metricfixtures.ConstGaugeMetric("num_threads", 0, []string{"num_threads", "exported_node"}, []string{"", ""}),
		metricfixtures.ConstGaugeMetric("num_threads_per_core", 0, []string{"num_threads_per_core", "exported_node"}, []string{"", ""}),
		metricfixtures.ConstGaugeMetric("num_gpus", 3, []string{"num_gpus", "exported_node"}, []string{"", ""}),
	}

	// When
	go func() {
		defer close(metricStream)
		exporter.Collect(metricStream)
	}()

	// Then
	var got []prometheus.Metric
	for metric := range metricStream {
		got = append(got, metric)
	}

	assert.Equal(t, want, got)
}

func TestDescribe(t *testing.T) {
	t.Parallel()

	k8sClient := fakekubelet.New(t,
		fakekubelet.WithPodResources(k8sfixtures.NewPodResourcesFixture(t)),
		fakekubelet.WithAMDCustomResourceNames([]string{"amd-custom-resource-name"}),
	)

	cardsInfo := [24]gpus.Card{
		0: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:b3:00.0",
			CardGUID:   "63755",
		},
		1: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:8e:00.0",
			CardGUID:   "63756",
		},
		2: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:34:00.0",
			CardGUID:   "63757",
		},
	}

	logger := testlogs.NewLogger()

	settings := exporters.Setup{
		K8SClient:      k8sClient,
		CardsInfo:      cardsInfo,
		Logger:         logger,
		WithKubernetes: true,
		GetMetricsFunc: makeAMDDataFuncFixture(t),
	}

	descStream := make(chan *prometheus.Desc)

	exporter := exporters.NewExporter(&settings)

	want := prometheus.NewDesc(
		"amd_data",
		"AMD Params",
		[]string{"socket"},
		nil,
	)

	// When
	go func() {
		defer close(descStream)
		exporter.Describe(descStream)
	}()

	// Then
	got := <-descStream

	require.NotNil(t, got)
	assert.Equal(t, want, got)
}

func makeAMDDataFuncFixture(t *testing.T) func() gpus.AMDParams {
	return func() gpus.AMDParams {
		t.Helper()

		amdParams := gpus.AMDParams{}
		amdParams.Init()

		amdParams.NumGPUs = 3
		amdParams.GPUDevID[0] = float64(0)
		amdParams.GPUPowerCap[0] = float64(300)
		amdParams.GPUPower[0] = float64(301)
		amdParams.GPUTemperature[0] = float64(302)
		amdParams.GPUSCLK[0] = float64(303)
		amdParams.GPUMCLK[0] = float64(304)
		amdParams.GPUUsage[0] = float64(305)
		amdParams.GPUMemoryUsage[0] = float64(306)

		amdParams.GPUDevID[1] = float64(0)
		amdParams.GPUPowerCap[1] = float64(300)
		amdParams.GPUPower[1] = float64(301)
		amdParams.GPUTemperature[1] = float64(302)
		amdParams.GPUSCLK[1] = float64(303)
		amdParams.GPUMCLK[1] = float64(304)
		amdParams.GPUUsage[1] = float64(305)
		amdParams.GPUMemoryUsage[1] = float64(306)

		amdParams.GPUDevID[2] = float64(0)
		amdParams.GPUPowerCap[2] = float64(300)
		amdParams.GPUPower[2] = float64(301)
		amdParams.GPUTemperature[2] = float64(302)
		amdParams.GPUSCLK[2] = float64(303)
		amdParams.GPUMCLK[2] = float64(304)
		amdParams.GPUUsage[2] = float64(305)
		amdParams.GPUMemoryUsage[2] = float64(306)

		return amdParams
	}
}
