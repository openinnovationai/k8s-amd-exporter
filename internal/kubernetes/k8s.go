package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"

	"github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/gpus"
	podbus "github.com/openinnovationai/k8s-amd-exporter/internal/exporters/domain/pods"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Setup parameters required to create k8s client.
type Setup struct {
	KubeletConn *grpc.ClientConn
	K8SClient   kubernetes.Interface
	Logger      *slog.Logger
	// CustomResourceNames contains additional names for amd resource. e.g. custom.amd.com/gpu
	CustomResourceNames []string
	// Contains the name of the node.
	NodeName string
	// Contains the pod name where this app is running.
	PodName string
	// Contains the pod namespace where this app is running.
	PodNamespace string
}

// Client implements k8s client behaviour.
type Client struct {
	kubeletConn         *grpc.ClientConn
	k8sClient           kubernetes.Interface
	api                 podresourcesapi.PodResourcesListerClient
	logger              *slog.Logger
	customResourceNames []string
	nodeName            string
	podName             string
	podNamespace        string
}

const (
	unixProtocol             = "unix"
	socketPathDefault string = "/var/lib/kubelet/pod-resources/kubelet.sock"
)

func NewClient(settings *Setup) *Client {
	newClient := Client{
		kubeletConn:         settings.KubeletConn,
		api:                 podresourcesapi.NewPodResourcesListerClient(settings.KubeletConn),
		logger:              settings.Logger,
		customResourceNames: settings.CustomResourceNames,
		nodeName:            settings.NodeName,
		k8sClient:           settings.K8SClient,
	}

	return &newClient
}

// CreateGRPCConn creates a grpc connection to call kubelet pod resources api.
// The podresources API is served by the kubelet locally, on the same node on
// which is running. On unix flavors, the endpoint is served over a unix
// domain socket.
func CreateGRPCConn(socketPath string) (*grpc.ClientConn, error) {
	if socketPath == "" {
		socketPath = socketPathDefault
	}

	slog.Info("using kubelet socket path", slog.String("path", socketPath))

	resolver.SetDefaultScheme("passthrough")

	conn, err := grpc.NewClient(socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{}

			return d.DialContext(ctx, unixProtocol, addr)
		}),
	)
	if err != nil {
		slog.Error(
			"creating kubelet pod resources api grpc client",
			slog.String("socketpath", socketPath),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("unable to connect to pod resources api: %w", err)
	}

	return conn, nil
}

// NewK8SClient creates a new kubernetes client with the given config object.
func NewK8SClient(kubeConfig *rest.Config) (*kubernetes.Clientset, error) {
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate k8s client: %w", err)
	}

	return kubeClient, nil
}

// NewKubeConfig get kubernetes config.
func NewKubeConfig() (*rest.Config, error) {
	kubeConfig, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get k8s config: %w", err)
	}

	return kubeConfig, nil
}

// GetPodsUsingDevices gets all pods that use AMD GPU devices and indexes them by device ID.
func (c *Client) GetPodsUsingDevices(ctx context.Context) (podbus.PodPerDevices, error) {
	devicePods, err := c.api.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		slog.Error("listing pod resources", slog.String("error", err.Error()))

		return nil, fmt.Errorf("unable to get pod resources: %w", err)
	}

	deviceToPodMap := make(podbus.PodPerDevices)

	for _, pod := range devicePods.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, device := range container.GetDevices() {
				if !c.containsRequestedDevices(device.GetResourceName()) {
					continue
				}

				podInfo := podbus.PodInfo{
					Name:      pod.GetName(),
					Namespace: pod.GetNamespace(),
					Container: container.GetName(),
					NodeName:  c.nodeName,
				}

				for _, deviceID := range device.GetDeviceIds() {
					c.logger.Debug("pod device info",
						slog.String("pod", podInfo.Name),
						slog.String("container", podInfo.Container),
						slog.String("namespace", podInfo.Namespace),
						slog.String("device-id", deviceID),
						slog.String("node-name", c.nodeName),
					)

					deviceToPodMap[deviceID] = podInfo
				}
			}
		}
	}

	return deviceToPodMap, nil
}

// GetPodsLabels search given pods and return pod labels.
// The key in the result is the namespaced name of the pod.
func (c *Client) GetPodsLabels(ctx context.Context, pods []podbus.PodInfo, customLabels []string) (map[string]podbus.Labels, error) {
	podList, err := c.getPodLabelsOnNode(ctx, pods, customLabels)
	if err != nil {
		c.logger.Error("getting pods within node",
			slog.String("error", err.Error()),
			slog.String("node", c.nodeName))
	}

	if err == nil && len(podList) > 0 {
		return podList, nil
	}

	c.logger.Info("trying pod by pod")

	podList = c.getPodsLabelsWithList(ctx, pods, customLabels)

	return podList, nil
}

// getPodsLabelsWithList search given pods and return pod data plus required labels.
func (c *Client) getPodsLabelsWithList(ctx context.Context, pods []podbus.PodInfo, customLabels []string) map[string]podbus.Labels {
	var result corev1.PodList

	for index := range pods {
		pod, err := c.k8sClient.CoreV1().Pods(pods[index].Namespace).Get(ctx, pods[index].Name, metav1.GetOptions{})
		if err != nil {
			c.logger.Error("getting pod labels",
				slog.String("error", err.Error()),
				slog.String("pod", pods[index].NamespacedName()))

			continue
		}

		if pod == nil {
			continue
		}

		result.Items = append(result.Items, *pod)
	}

	return toPodsMap(result.Items, customLabels)
}

// getPodLabelsOnNode search pods within the preconfigured node and filter them with given pods.
func (c *Client) getPodLabelsOnNode(ctx context.Context, pods []podbus.PodInfo, customLabels []string) (map[string]podbus.Labels, error) {
	if c.nodeName == "" {
		return nil, nil
	}

	podList, err := c.k8sClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", c.nodeName),
	})
	if err != nil {
		c.logger.Error("getting pods labels",
			slog.String("error", err.Error()),
			slog.String("node", c.nodeName))

		return nil, fmt.Errorf("unable to get pods from node: %w", err)
	}

	if podList == nil {
		return nil, nil
	}

	var items []corev1.Pod

	for index := range podList.Items {
		exists := slices.ContainsFunc(pods, func(item podbus.PodInfo) bool {
			name := podbus.NewNamespacedName(podList.Items[index].Name, podList.Items[index].Namespace)

			return name.String() == item.NamespacedName()
		})
		if exists {
			items = append(items, podList.Items[index])
		}
	}

	return toPodsMap(items, customLabels), nil
}

func toPodsMap(podList []corev1.Pod, requiredLabels []string) map[string]podbus.Labels {
	result := make(map[string]podbus.Labels)

	for index := range podList {
		key := podbus.NewNamespacedName(podList[index].Name, podList[index].Namespace)
		result[key.String()] = selectLabels(podList[index].Labels, requiredLabels)
	}

	return result
}

func selectLabels(podLabels map[string]string, requiredLabels []string) map[string]string {
	selectedLabels := make(map[string]string)

	for labelKey, labelValue := range podLabels {
		if slices.Contains(requiredLabels, strings.ToLower(labelKey)) {
			selectedLabels[labelKey] = labelValue
		}
	}

	return selectedLabels
}

func (c *Client) PodName() string {
	return c.podName
}

func (c *Client) PodNamespace() string {
	return c.podNamespace
}

func (c *Client) NodeName() string {
	return c.nodeName
}

// containsRequestedDevices return true if the given resource names match amd resources
// or provided custom resource names by configuration.
func (c *Client) containsRequestedDevices(resourceName string) bool {
	return resourceName == gpus.AMDResourceName ||
		slices.Contains(c.customResourceNames, resourceName)
}

// Close closes kubelet grpc connection.
func (c *Client) Close() {
	if c.kubeletConn == nil {
		return
	}

	err := c.kubeletConn.Close()
	if err != nil {
		c.logger.Error("closing pod resources grpc connection", slog.String("error", err.Error()))
	}
}
