package configcenter

import (
	"fmt"

	"github.com/mesh-operator/pkg/adapter/component"
	"github.com/mesh-operator/pkg/adapter/options"
	"k8s.io/klog"
)

type constructor func(regOpt options.Configuration) (component.ConfigurationCenter, error)

var (
	configInstance = make(map[string]constructor)
)

func Registry(typ string, f constructor) {
	if _, ok := configInstance[typ]; ok {
		klog.Fatalln("repeat registry [config center instance]: %s", typ)
	}
	configInstance[typ] = f
}

func GetRegistry(opt options.Configuration) (component.ConfigurationCenter, error) {
	if f, ok := configInstance[opt.Type]; ok {
		return f(opt)
	}
	return nil, fmt.Errorf("config center {%s} was not implemented", opt.Type)
}
