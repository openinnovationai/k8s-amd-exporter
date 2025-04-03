package pods

import (
	"fmt"
	"maps"
	"slices"
)

// Resources contains data about pods using gpus and the node name.
type Resources struct {
	Pods     map[string][]PodInfo
	NodeName string
}

// PodInfo contains information related to pods.
type PodInfo struct {
	Name      string
	Namespace string
	Container string
	NodeName  string
	Labels    Labels
}

type NamespacedName struct {
	Name      string
	Namespace string
}

// Label contains pod labels.
type Labels map[string]string

// PodPerDevices map of pods where device id is the key.
type PodPerDevices map[string]PodInfo

func NewResources() *Resources {
	return &Resources{
		Pods: make(map[string][]PodInfo),
	}
}

func NewNamespacedName(name, namespace string) NamespacedName {
	if namespace == "" {
		namespace = "default"
	}

	return NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

func (p PodInfo) NamespacedName() string {
	if p.Namespace == "" {
		p.Namespace = "default"
	}

	return fmt.Sprintf("%s/%s", p.Namespace, p.Name)
}

func (n NamespacedName) String() string {
	return fmt.Sprintf("%s/%s", n.Namespace, n.Name)
}

func (p PodPerDevices) Pods() []PodInfo {
	podList := make([]PodInfo, 0, len(p))

	for apod := range maps.Values(p) {
		podList = append(podList, apod)
	}

	return podList
}

// SortKeys sort label keys.
func (l Labels) SortKeys() []string {
	keys := make([]string, 0, len(l))

	for key := range l {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}
