package registry

import (
	"fmt"

	"github.com/mesh-operator/pkg/adapter/component"
	"github.com/mesh-operator/pkg/adapter/options"
	"k8s.io/klog"
)

type constructor func(regOpt options.Registry) (component.Registry, error)

var (
	registryInstance = make(map[string]constructor)
)

func Registry(typ string, f constructor) {
	if _, ok := registryInstance[typ]; ok {
		klog.Fatalln("repeat registry [registry center instance]: %s", typ)
	}
	registryInstance[typ] = f
}

func GetRegistry(opt options.Registry) (component.Registry, error) {
	if f, ok := registryInstance[opt.Type]; ok {
		return f(opt)
	}
	return nil, fmt.Errorf("registry center {%s} was not implemented", opt.Type)
}
