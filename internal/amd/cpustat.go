/*
 * MIT-X11 Open Source License
 *
 * Copyright (c) 2022, Advanced Micro Devices, Inc.
 * All rights reserved.
 *
 * Developed by:
 *
 *                 AMD Research and AMD Software Development
 *
 *                 Advanced Micro Devices, Inc.
 *
 *                 www.amd.com
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or
 * sellcopies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 *  - The above copyright notice and this permission notice shall be included in
 *    all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * Except as contained in this notice, the name of the Advanced Micro Devices,
 * Inc. shall not be used in advertising or otherwise to promote the sale, use
 * or other dealings in this Software without prior written authorization from
 * the Advanced Micro Devices, Inc.
 *
 */

// Package cpustat provides an example parser for Linux CPU utilization statistics.
package amd

import (
	"log/slog"

	goamdsmi "github.com/amd/go_amd_smi"
	// goamdsmi "gitlab.com/openinnovationai/platform/infra/amd/amd_smi_exporter_v2/internal/amd/goamdsmi"
	"gitlab.com/openinnovationai/platform/infra/amd/amd_smi_exporter_v2/internal/exporters/domain/gpus"
)

var UINT16_MAX = uint16(0xFFFF)
var UINT32_MAX = uint32(0xFFFFFFFF)
var UINT64_MAX = uint64(0xFFFFFFFFFFFFFFFF)

type Scanner struct {
	logger *slog.Logger
}

func NewScanner(logger *slog.Logger) *Scanner {
	newScanner := Scanner{
		logger: logger,
	}

	return &newScanner
}

func (s *Scanner) Scan() gpus.AMDParams {
	s.logger.Debug("scanning metrics")

	var stat gpus.AMDParams
	stat.Init()

	value64 := uint64(0)
	value32 := uint32(0)
	value16 := uint16(0)

	s.logger.Debug("GO_cpu_init", slog.Bool("value", goamdsmi.GO_cpu_init()))
	if true == goamdsmi.GO_cpu_init() {
		num_sockets := int(goamdsmi.GO_cpu_number_of_sockets_get())
		num_threads := int(goamdsmi.GO_cpu_number_of_threads_get())
		num_threads_per_core := int(goamdsmi.GO_cpu_threads_per_core_get())

		stat.Sockets = uint(num_sockets)
		stat.Threads = uint(num_threads)
		stat.ThreadsPerCore = uint(num_threads_per_core)

		for i := 0; i < num_threads; i++ {
			value64 = uint64(goamdsmi.GO_cpu_core_energy_get(i))
			if UINT64_MAX != value64 {
				stat.CoreEnergy[i] = float64(value64)
			}
			value64 = 0

			value32 = uint32(goamdsmi.GO_cpu_core_boostlimit_get(i))
			if UINT32_MAX != value32 {
				stat.CoreBoost[i] = float64(value32)
			}
			value32 = 0
		}

		for i := 0; i < num_sockets; i++ {
			value64 = uint64(goamdsmi.GO_cpu_socket_energy_get(i))
			if UINT64_MAX != value64 {
				stat.SocketEnergy[i] = float64(value64)
			}
			value64 = 0

			value32 = uint32(goamdsmi.GO_cpu_socket_power_get(i))
			if UINT32_MAX != value32 {
				stat.SocketPower[i] = float64(value32)
			}
			value32 = 0

			value32 = uint32(goamdsmi.GO_cpu_socket_power_cap_get(i))
			if UINT32_MAX != value32 {
				stat.PowerLimit[i] = float64(value32)
			}
			value32 = 0

			value32 = uint32(goamdsmi.GO_cpu_prochot_status_get(i))
			if UINT32_MAX != value32 {
				stat.ProchotStatus[i] = float64(value32)
			}
			value32 = 0
		}
	}

	s.logger.Debug("GO_gpu_init", slog.Bool("value", goamdsmi.GO_gpu_init()))
	if true == goamdsmi.GO_gpu_init() {

		num_gpus := int(goamdsmi.GO_gpu_num_monitor_devices())
		stat.NumGPUs = uint(num_gpus)

		for i := 0; i < num_gpus; i++ {
			value16 = uint16(goamdsmi.GO_gpu_dev_id_get(i))
			if UINT16_MAX != value16 {
				stat.GPUDevID[i] = float64(value16)
			}
			value16 = 0

			value64 = uint64(goamdsmi.GO_gpu_dev_power_cap_get(i))
			if UINT64_MAX != value64 {
				stat.GPUPowerCap[i] = float64(value64)
			}
			value64 = 0

			value64 = uint64(goamdsmi.GO_gpu_dev_power_get(i))
			if UINT64_MAX != value64 {
				stat.GPUPower[i] = float64(value64)
			}
			value64 = 0

			//Get the value for GPU current temperature. Sensor = 0(GPU), Metric = 0(current)
			value64 = uint64(goamdsmi.GO_gpu_dev_temp_metric_get(i, 0, 0))
			if UINT64_MAX == value64 {
				//Sensor = 1 (GPU Junction Temp)
				value64 = uint64(goamdsmi.GO_gpu_dev_temp_metric_get(i, 1, 0))
			}
			if UINT64_MAX != value64 {
				stat.GPUTemperature[i] = float64(value64)
			}
			value64 = 0

			value64 = uint64(goamdsmi.GO_gpu_dev_gpu_clk_freq_get_sclk(i))
			if UINT64_MAX != value64 {
				stat.GPUSCLK[i] = float64(value64)
			}
			value64 = 0

			value64 = uint64(goamdsmi.GO_gpu_dev_gpu_clk_freq_get_mclk(i))
			if UINT64_MAX != value64 {
				stat.GPUMCLK[i] = float64(value64)
			}
			value64 = 0

			value32 = uint32(goamdsmi.GO_gpu_dev_gpu_busy_percent_get(i))
			if UINT32_MAX != value32 {
				stat.GPUUsage[i] = float64(value32)
			}
			value32 = 0

			value64 = uint64(goamdsmi.GO_gpu_dev_gpu_memory_busy_percent_get(i))
			if UINT64_MAX != value64 {
				stat.GPUMemoryUsage[i] = float64(value64)
			}
			value64 = 0
		}
	}

	return stat
}
