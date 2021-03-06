package component

import "github.com/mesh-operator/pkg/adapter/types"

type ConfigurationCenter interface {
	Start() error

	Events() <-chan *types.ConfigEvent

	FindConfiguratorConfig(serviceName string) *types.ConfiguratorConfig

	Stop() error
}
