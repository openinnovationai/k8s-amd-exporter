package metrics_test

import (
	"testing"

	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/gpus"
	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/metrics"
	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/pods"
	"github.com/openinnovationai/k8s-amd-exporter/internal/sdk/unittests/metricfixtures"
	"github.com/openinnovationai/k8s-amd-exporter/internal/sdk/unittests/testlogs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewAMDMetrics(t *testing.T) {
	t.Parallel()
	// Given
	settings := metrics.Setup{}

	want := &metrics.AMDMetrics{
		DataDesc: &metrics.CustomMetric{
			Name:     "amd_data",
			HelpText: "AMD Params",
			Labels:   []string{"socket"},
		},
		CoreEnergy: &metrics.CustomMetric{
			Name:      "core_energy",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.CounterValue,
			Labels:    []string{"thread"},
		},
		SocketEnergy: &metrics.CustomMetric{
			Name:      "socket_energy",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.CounterValue,
			Labels:    []string{"socket"},
		},
		BoostLimit: &metrics.CustomMetric{
			Name:      "boost_limit",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"thread"},
		},
		SocketPower: &metrics.CustomMetric{
			Name:      "socket_power",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"socket"},
		},
		PowerLimit: &metrics.CustomMetric{
			Name:      "power_limit",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"power_limit"},
		},
		ProchotStatus: &metrics.CustomMetric{
			Name:      "prochot_status",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"prochot_status"},
		},
		Sockets: &metrics.CustomMetric{
			Name:      "num_sockets",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"num_sockets"},
		},
		Threads: &metrics.CustomMetric{
			Name:      "num_threads",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"num_threads"},
		},
		ThreadsPerCore: &metrics.CustomMetric{
			Name:      "num_threads_per_core",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"num_threads_per_core"},
		},
		NumGPUs: &metrics.CustomMetric{
			Name:      "num_gpus",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"num_gpus"},
		},
		GPUDevID: &metrics.CustomMetric{
			Name:      "gpu_dev_id",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"gpu_dev_id", "productname", "device"},
		},
		GPUPowerCap: &metrics.CustomMetric{
			Name:      "gpu_power_cap",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Divide:    true,
			Divisor:   1e6,
			Labels:    []string{"gpu_power_cap", "productname", "device"},
		},
		GPUPower: &metrics.CustomMetric{
			Name:      "gpu_power",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.CounterValue,
			Divide:    true,
			Divisor:   1e6,
			Labels:    []string{"gpu_power", "productname", "device"},
		},
		GPUTemperature: &metrics.CustomMetric{
			Name:      "gpu_current_temperature",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Divide:    true,
			Divisor:   1e3,
			Labels:    []string{"gpu_current_temperature", "productname", "device"},
		},
		GPUSCLK: &metrics.CustomMetric{
			Name:      "gpu_SCLK",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Divide:    true,
			Divisor:   1e6,
			Labels:    []string{"gpu_SCLK", "productname", "device"},
		},
		GPUMCLK: &metrics.CustomMetric{
			Name:      "gpu_MCLK",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Divide:    true,
			Divisor:   1e6,
			Labels:    []string{"gpu_MCLK", "productname", "device"},
		},
		GPUUsage: &metrics.CustomMetric{
			Name:      "gpu_use_percent",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"gpu_use_percent", "productname", "device"},
		},
		GPUMemoryUsage: &metrics.CustomMetric{
			Name:      "gpu_memory_use_percent",
			Namespace: "amd",
			HelpText:  "AMD Params",
			Type:      prometheus.GaugeValue,
			Labels:    []string{"gpu_memory_use_percent", "productname", "device"},
		},
	}
	// When
	got := metrics.NewAMDMetrics(&settings)
	// Then
	assert.Equal(t, want, got)
}

func TestCollectAndBuildMetricsWithoutGPUs(t *testing.T) {
	t.Parallel()
	// Given
	settings := metrics.Setup{
		AMDParamsHandler: makeAMDDataAllDataFuncFixture(t),
		WithKubernetes:   true,
		Logger:           testlogs.NewLogger(),
	}
	amdMetrics := metrics.NewAMDMetrics(&settings)
	amdMetrics.CardsInfo = makeCardInfoFixture(t)
	amdMetrics.K8SResources = &pods.Resources{
		Pods:     makeK8SResourcesFixture(t),
		NodeName: "workload-node",
	}

	want := []string{
		`Desc{fqName: "amd_core_energy", help: "AMD Params", constLabels: {}, variableLabels: {thread,exported_node}}`,
		`Desc{fqName: "amd_boost_limit", help: "AMD Params", constLabels: {}, variableLabels: {thread,exported_node}}`,
		`Desc{fqName: "amd_socket_energy", help: "AMD Params", constLabels: {}, variableLabels: {socket,exported_node}}`,
		`Desc{fqName: "amd_socket_power", help: "AMD Params", constLabels: {}, variableLabels: {socket,exported_node}}`,
		`Desc{fqName: "amd_power_limit", help: "AMD Params", constLabels: {}, variableLabels: {power_limit,exported_node}}`,
		`Desc{fqName: "amd_prochot_status", help: "AMD Params", constLabels: {}, variableLabels: {prochot_status,exported_node}}`,
		`Desc{fqName: "amd_gpu_dev_id", help: "AMD Params", constLabels: {}, variableLabels: {gpu_dev_id,productname,device,exported_node,exported_pod,exported_container,exported_namespace}}`,
		`Desc{fqName: "amd_gpu_power_cap", help: "AMD Params", constLabels: {}, variableLabels: {gpu_power_cap,productname,device,exported_node,exported_pod,exported_container,exported_namespace}}`,
		`Desc{fqName: "amd_gpu_power", help: "AMD Params", constLabels: {}, variableLabels: {gpu_power,productname,device,exported_node,exported_pod,exported_container,exported_namespace}}`,
		`Desc{fqName: "amd_gpu_current_temperature", help: "AMD Params", constLabels: {}, variableLabels: {gpu_current_temperature,productname,device,exported_node,exported_pod,exported_container,exported_namespace}}`,
		`Desc{fqName: "amd_gpu_SCLK", help: "AMD Params", constLabels: {}, variableLabels: {gpu_SCLK,productname,device,exported_node,exported_pod,exported_container,exported_namespace}}`,
		`Desc{fqName: "amd_gpu_MCLK", help: "AMD Params", constLabels: {}, variableLabels: {gpu_MCLK,productname,device,exported_node,exported_pod,exported_container,exported_namespace}}`,
		`Desc{fqName: "amd_gpu_use_percent", help: "AMD Params", constLabels: {}, variableLabels: {gpu_use_percent,productname,device,exported_node,exported_pod,exported_container,exported_namespace}}`,
		`Desc{fqName: "amd_gpu_memory_use_percent", help: "AMD Params", constLabels: {}, variableLabels: {gpu_memory_use_percent,productname,device,exported_node,exported_pod,exported_container,exported_namespace}}`,
		`Desc{fqName: "amd_num_sockets", help: "AMD Params", constLabels: {}, variableLabels: {num_sockets,exported_node}}`,
		`Desc{fqName: "amd_num_threads", help: "AMD Params", constLabels: {}, variableLabels: {num_threads,exported_node}}`,
		`Desc{fqName: "amd_num_threads_per_core", help: "AMD Params", constLabels: {}, variableLabels: {num_threads_per_core,exported_node}}`,
		`Desc{fqName: "amd_num_gpus", help: "AMD Params", constLabels: {}, variableLabels: {num_gpus,exported_node}}`,
	}

	// When
	metrics := amdMetrics.CollectAndBuildMetrics()

	// Then
	var got []string
	for i := range metrics {
		got = append(got, metrics[i].Desc().String())
	}
	assert.Equal(t, want, got)
}

func TestCollectAndBuildMetricsOnlyGPUs(t *testing.T) {
	t.Parallel()
	// Given
	settings := metrics.Setup{
		AMDParamsHandler: makeAMDDataFuncFixture(t),
		WithKubernetes:   true,
		Logger:           testlogs.NewLogger(),
	}
	amdMetrics := metrics.NewAMDMetrics(&settings)
	amdMetrics.CardsInfo = makeCardInfoFixture(t)
	amdMetrics.K8SResources = &pods.Resources{
		Pods:     makeK8SResourcesFixture(t),
		NodeName: "workload-node",
	}

	want := []prometheus.Metric{
		metricfixtures.ConstCounterMetric("core_energy", -1, []string{"thread", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstGaugeMetric("boost_limit", -1, []string{"thread", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstCounterMetric("socket_energy", -1, []string{"socket", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstGaugeMetric("socket_power", -1, []string{"socket", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstGaugeMetric("power_limit", -1, []string{"power_limit", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstGaugeMetric("prochot_status", -1, []string{"prochot_status", "exported_node"}, []string{"0", "workload-node"}),

		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("num_sockets", 1, []string{"num_sockets", "exported_node"}, []string{"", "workload-node"}),
		metricfixtures.ConstGaugeMetric("num_threads", 1, []string{"num_threads", "exported_node"}, []string{"", "workload-node"}),
		metricfixtures.ConstGaugeMetric("num_threads_per_core", 1, []string{"num_threads_per_core", "exported_node"}, []string{"", "workload-node"}),
		metricfixtures.ConstGaugeMetric("num_gpus", 4, []string{"num_gpus", "exported_node"}, []string{"", "workload-node"}),
	}

	// When
	got := amdMetrics.CollectAndBuildMetrics()

	// Then
	assert.Equal(t, want, got)
}

func TestCollectAndBuildMetricsGPUsAndK8SWithLabels(t *testing.T) {
	t.Parallel()
	// Given
	settings := metrics.Setup{
		AMDParamsHandler: makeAMDDataFuncFixture(t),
		WithKubernetes:   true,
		Logger:           testlogs.NewLogger(),
	}
	amdMetrics := metrics.NewAMDMetrics(&settings)
	amdMetrics.CardsInfo = makeCardInfoFixture(t)
	amdMetrics.K8SResources = &pods.Resources{
		Pods:     makeK8SResourcesWithLabelsFixture(t),
		NodeName: "workload-node",
	}

	want := []prometheus.Metric{
		metricfixtures.ConstCounterMetric("core_energy", -1, []string{"thread", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstGaugeMetric("boost_limit", -1, []string{"thread", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstCounterMetric("socket_energy", -1, []string{"socket", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstGaugeMetric("socket_power", -1, []string{"socket", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstGaugeMetric("power_limit", -1, []string{"power_limit", "exported_node"}, []string{"0", "workload-node"}),
		metricfixtures.ConstGaugeMetric("prochot_status", -1, []string{"prochot_status", "exported_node"}, []string{"0", "workload-node"}),

		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id", "label_1", "label_oip_author_username"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a", "value-1", "gpu-user-1"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_dev_id", 0, metricfixtures.GPULabels("gpu_dev_id"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap", "label_1", "label_oip_author_username"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a", "value-1", "gpu-user-1"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_power_cap", 0.0003, metricfixtures.GPULabels("gpu_power_cap"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2", "value-1", "value-2"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power", "label_1", "label_oip_author_username"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a", "value-1", "gpu-user-1"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstCounterMetric("gpu_power", 0.000301, metricfixtures.GPULabels("gpu_power"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature", "label_1", "label_oip_author_username"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a", "value-1", "gpu-user-1"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_current_temperature", 0.302, metricfixtures.GPULabels("gpu_current_temperature"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK", "label_1", "label_oip_author_username"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a", "value-1", "gpu-user-1"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_SCLK", 0.000303, metricfixtures.GPULabels("gpu_SCLK"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK", "label_1", "label_oip_author_username"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a", "value-1", "gpu-user-1"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_MCLK", 0.000304, metricfixtures.GPULabels("gpu_MCLK"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent", "label_1", "label_oip_author_username"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a", "value-1", "gpu-user-1"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_use_percent", 305, metricfixtures.GPULabels("gpu_use_percent"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent", "label_1", "label_2"), []string{"0", "amdinstinctmi250(mcm)oamacmba", "amd0", "node-1", "pod-ii", "container-1", "team-b", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent", "label_1", "label_2"), []string{"1", "amdinstinctmi250(mcm)oamacmba", "amd1", "node-1", "pod-c", "container-1", "team-2", "value-1", "value-2"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent", "label_1", "label_oip_author_username"), []string{"2", "amdinstinctmi250(mcm)oamacmba", "amd2", "node-1", "pod-1", "container-1", "team-a", "value-1", "gpu-user-1"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-y", "container-1", "team-a"}),
		metricfixtures.ConstGaugeMetric("gpu_memory_use_percent", 306, metricfixtures.GPULabels("gpu_memory_use_percent"), []string{"3", "amdinstinctmi250(mcm)oamacmba", "amd3", "node-1", "pod-z", "container-1", "team-a"}),

		metricfixtures.ConstGaugeMetric("num_sockets", 1, []string{"num_sockets", "exported_node"}, []string{"", "workload-node"}),
		metricfixtures.ConstGaugeMetric("num_threads", 1, []string{"num_threads", "exported_node"}, []string{"", "workload-node"}),
		metricfixtures.ConstGaugeMetric("num_threads_per_core", 1, []string{"num_threads_per_core", "exported_node"}, []string{"", "workload-node"}),
		metricfixtures.ConstGaugeMetric("num_gpus", 4, []string{"num_gpus", "exported_node"}, []string{"", "workload-node"}),
	}

	// When
	got := amdMetrics.CollectAndBuildMetrics()

	// Then
	assert.Equal(t, want, got)
}

func makeAMDDataFuncFixture(t *testing.T) func() gpus.AMDParams {
	return func() gpus.AMDParams {
		t.Helper()

		amdParams := gpus.AMDParams{}
		amdParams.Init()

		amdParams.Threads = 1
		amdParams.Sockets = 1
		amdParams.ThreadsPerCore = 1
		amdParams.NumGPUs = 1

		amdParams.NumGPUs = 4
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

		amdParams.GPUDevID[3] = float64(0)
		amdParams.GPUPowerCap[3] = float64(300)
		amdParams.GPUPower[3] = float64(301)
		amdParams.GPUTemperature[3] = float64(302)
		amdParams.GPUSCLK[3] = float64(303)
		amdParams.GPUMCLK[3] = float64(304)
		amdParams.GPUUsage[3] = float64(305)
		amdParams.GPUMemoryUsage[3] = float64(306)

		return amdParams
	}
}

func makeAMDDataAllDataFuncFixture(t *testing.T) func() gpus.AMDParams {
	return func() gpus.AMDParams {
		t.Helper()

		amdParams := gpus.AMDParams{}
		amdParams.Init()

		amdParams.Threads = 1
		amdParams.Sockets = 1
		amdParams.ThreadsPerCore = 1
		amdParams.NumGPUs = 1

		amdParams.CoreBoost[0] = float64(1)
		amdParams.CoreEnergy[0] = float64(2)
		amdParams.PowerLimit[0] = float64(3)
		amdParams.ProchotStatus[0] = float64(4)
		amdParams.SocketEnergy[0] = float64(5)
		amdParams.SocketPower[0] = float64(6)

		amdParams.GPUDevID[0] = float64(0)
		amdParams.GPUPowerCap[0] = float64(300)
		amdParams.GPUPower[0] = float64(301)
		amdParams.GPUTemperature[0] = float64(302)
		amdParams.GPUSCLK[0] = float64(303)
		amdParams.GPUMCLK[0] = float64(304)
		amdParams.GPUUsage[0] = float64(305)
		amdParams.GPUMemoryUsage[0] = float64(306)

		return amdParams
	}
}

func makeCardInfoFixture(t *testing.T) [24]gpus.Card {
	t.Helper()

	return [24]gpus.Card{
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
		3: {
			Cardseries: "amdinstinctmi250(mcm)oamacmba",
			Cardmodel:  "0x740c",
			Cardvendor: "advancedmicrodevices,inc.[amd/ati]",
			CardSKU:    "d65210v",
			PCIBus:     "0000:11:00.0",
			CardGUID:   "63757",
		},
	}
}

func makeK8SResourcesFixture(t *testing.T) map[string][]pods.PodInfo {
	t.Helper()

	return map[string][]pods.PodInfo{
		"0000:34:00.0": {
			{
				Name:      "pod-1",
				Namespace: "team-a",
				Container: "container-1",
				NodeName:  "node-1",
			},
		},
		"0000:8e:00.0": {
			{
				Name:      "pod-c",
				Namespace: "team-2",
				Container: "container-1",
				NodeName:  "node-1",
			},
		},
		"0000:b3:00.0": {
			{
				Name:      "pod-ii",
				Namespace: "team-b",
				Container: "container-1",
				NodeName:  "node-1",
			},
		},
		"0000:11:00.0": {
			{
				Name:      "pod-y",
				Namespace: "team-a",
				Container: "container-1",
				NodeName:  "node-1",
			},
			{
				Name:      "pod-z",
				Namespace: "team-a",
				Container: "container-1",
				NodeName:  "node-1",
			},
		},
	}
}

func makeK8SResourcesWithLabelsFixture(t *testing.T) map[string][]pods.PodInfo {
	t.Helper()

	return map[string][]pods.PodInfo{
		"0000:34:00.0": { // card2
			{
				Name:      "pod-1",
				Namespace: "team-a",
				Container: "container-1",
				Labels: pods.Labels{
					"label_1":             "value-1",
					"oip/author-username": "gpu-user-1",
				},
				NodeName: "node-1",
			},
		},
		"0000:8e:00.0": { // card1
			{
				Name:      "pod-c",
				Namespace: "team-2",
				Container: "container-1",
				Labels: pods.Labels{
					"label_1": "value-1",
					"label_2": "value-2",
				},
				NodeName: "node-1",
			},
		},
		"0000:b3:00.0": { // card0
			{
				Name:      "pod-ii",
				Namespace: "team-b",
				Container: "container-1",
				Labels: pods.Labels{
					"label_1": "value-1",
					"label_2": "value-2",
				},
				NodeName: "node-1",
			},
		},
		"0000:11:00.0": { // card3
			{
				Name:      "pod-y",
				Namespace: "team-a",
				Container: "container-1",
				NodeName:  "node-1",
			},
			{
				Name:      "pod-z",
				Namespace: "team-a",
				Container: "container-1",
				NodeName:  "node-1",
			},
		},
	}
}
