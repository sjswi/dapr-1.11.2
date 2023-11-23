package components

import componentsV1alpha1 "github.com/dapr/dapr/pkg/apis/components/v1alpha1"

// HostComponents loads components from a given directory.
type HostComponents struct {
	componentsManifestLoader ManifestLoader[componentsV1alpha1.Component]
}

// NewHostComponents returns a new HostComponents.
func NewHostComponents(resourcesPaths ...string) *HostComponents {
	return &HostComponents{
		componentsManifestLoader: NewDiskManifestLoader[componentsV1alpha1.Component](resourcesPaths...),
	}
}

// LoadComponents loads dapr components from a given directory.
func (s *HostComponents) LoadComponents() ([]componentsV1alpha1.Component, error) {
	return s.componentsManifestLoader.Load()
}
