[![slack](https://img.shields.io/badge/slack-brigade-brightgreen.svg?logo=slack)](https://app.slack.com/client/T076JPMCTDM/C08C871TRGB)

# AMD SMI Exporter 

The AMD SMI Exporter is an application that exports AMD CPU & GPU metrics to the Prometheus server. It employs the [E-SMI In-Band C library](https://github.com/amd/esmi_ib_library.git) & [ROCm SMI Library](https://github.com/RadeonOpenCompute/rocm_smi_lib.git) for its data acquisition. The exporter and the E-SMI/ROCm-SMI library have a
[GO binding](https://github.com/amd/go_amd_smi.git) that provides an interface between the e-smi,rocm-smi C,C++ library and the GO exporter code.

This project is an enhancement of the [AMD SMI Exporter](https://github.com/amd/amd_smi_exporter.git) project.

## Dependencies

* [E-SMI In-Band C library](https://github.com/amd/esmi_ib_library.git).
* [ROCm SMI Library](https://github.com/RadeonOpenCompute/rocm_smi_lib.git).
* [GO binding](https://github.com/amd/go_amd_smi.git).

Please refer to the below links for the build and installation instructions.
<https://github.com/amd/esmi_ib_library/blob/master/docs/README.md>,
<https://github.com/amd/go_amd_smi/blob/master/README.md>, and
<https://github.com/RadeonOpenCompute/rocm_smi_lib/blob/master/README.md>

This project provides a base [Dockerfile](deploy/Dockerfile.base) to show how an environment could be built from a container.

## E-SMI In-Band C library

ESMI requires AMD HSMP.
The AMD HSMP driver is now part of the Linux kernel upstream starting in v5.18-rc1, however it might not be detected in the system. If there is an error related to amd_hsmp.h header

try the following, git clone and copy amd_hsmp.h to usr/include/asm-generic

```sh
$ git clone https://github.com/amd/amd_hsmp
$ cd amd_hsmp
$ mv amd_hsmp.h /usr/include/asm-generic
```

or 

```sh
RUN git clone https://github.com/amd/amd_hsmp.git \
&& mkdir -p /usr/include/x86_64-linux-gnu/asm \
&& cp ./amd_hsmp/amd_hsmp.h /usr/include/x86_64-linux-gnu/asm/amd_hsmp.h
```
	
then proceed with build as provide in <https://github.com/amd/esmi_ib_library/blob/master/docs/README.md>

## How to test

You can run unit tests using this Make command

```sh
make test
```

## How to build

* The binary file can be created by running the following command. Binary will be created in `./bin` directory within this project. Make sure you have installed the required dependencies for [GO binding](https://github.com/amd/go_amd_smi.git), otherwise you will get compilation erros.

```sh
make build-linux
```

## How to build container images.

You could set a different container builder by setting the `CONTAINERTOOL` environment variable (`docker` by default).

! Before building the exporter image, make sure you have the base images to build exporter binary and create final exporter image.

* Create a base image to build the exporter by running these commands

```sh
make build-base-dev-image
make tag-base-dev-image
```

* Create a base image to run the exporter by running these commands

```sh
make build-base-image
make tag-base-image
```

* Create exporter image

```sh
make build-image
make tag-image
```

Once you verified all of them were created without any issues, you could push this into the registry.

```sh
make push-image
```

## Environment Variables

Besides the required dependencies, this application needs the following environment variables.

```sh
AMD_EXPORTER_LOG_LEVEL=development
AMD_EXPORTER_WEB_SERVER_PORT=2021
AMD_EXPORTER_KUBELET_SOCKET_PATH=/var/lib/kubelet/pod-resources/kubelet.sock
AMD_EXPORTER_RESOURCE_NAMES=resourcename1,resourcename2
AMD_EXPORTER_WITH_KUBERNETES=true
AMD_EXPORTER_NODE_NAME=oi-wn-gpu-amd-01.test.oiai.corp
AMD_EXPORTER_POD_LABELS=label_oip_tenant_id,label_oip_author_username,label_oip_workspace_id
```

* **AMD_EXPORTER_LOG_LEVEL**: could be `development` or `production`. development shows `debug` logs and production from `info` ones.
* **AMD_EXPORTER_WEB_SERVER_PORT**: http server port.
* **AMD_EXPORTER_KUBELET_SOCKET_PATH**: kubelet pod resources api socket path.
* **AMD_EXPORTER_RESOURCE_NAMES**: additional resource names to `amd.com/gpu`.
* **AMD_EXPORTER_WITH_KUBERNETES**: flag to indicates the exporter that scanning pods is required.
* **AMD_EXPORTER_NODE_NAME**: if you are using kubernetes environment, this contains the cluster node name.
* **AMD_EXPORTER_POD_LABELS**: pod labels to be added to exporter labels.

Regarding the `AMD_EXPORTER_NODE_NAME` environment variable, you can get its value by adding this setting to your manifest.

```yaml
spec:
  containers:
  - image: registry.gitlab.com/openinnovationai/platform/infra/amd/amd_smi_exporter_v2/amd-smi-exporter:0.1.0
    env:
    - name: AMD_EXPORTER_NODE_NAME
      valueFrom:
        fieldRef:
          fieldPath: spec.nodeName
```

For the environment variable `AMD_EXPORTER_POD_LABELS`, you are adding there the list of labels as you have them inside the pods. let's assume you have these labels

```yaml
    env:
    - name: AMD_EXPORTER_POD_LABELS                                                                                                                                                                           
      value: oip/author-username,oip/tenant-id,oip/workspace-id
```

so for each pod you should have these labels

```yaml
     oip/author-username: gpu-user-1                                                                                                                                                                                
     oip/tenant-id: amdexporter                                                                                                                                                                                
     oip/workspace-id: 7a12749b-e9a7-47a7-b75b-7eb994d66e6c     
```

## How to deploy for testing purposes

There is a pod manifest at `./deploy/amd-gpu-pod-2.yaml` that you could use to deploy this exporter to your cluster. It contains the configurations required to allow this object to read GPU information.
