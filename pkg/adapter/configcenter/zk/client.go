package zk

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/mesh-operator/pkg/adapter/component"
	"github.com/mesh-operator/pkg/adapter/configcenter"
	"github.com/mesh-operator/pkg/adapter/options"
	"github.com/mesh-operator/pkg/adapter/types"
	"github.com/mesh-operator/pkg/adapter/utils"
	"github.com/mesh-operator/pkg/adapter/zookeeper"
	zkClient "github.com/samuel/go-zookeeper/zk"
	"k8s.io/klog"
	"time"
)

func init() {
	configcenter.Registry("zk", New)
}

type ConfigClient struct {
	conn          *zkClient.Conn
	out           chan *types.ConfigEvent
	configEntries map[string]*types.ConfiguratorConfig
	rootPathCache *zookeeper.PathCache
}

func New(opt options.Configuration) (component.ConfigurationCenter, error) {
	conn, _, err := zkClient.Connect(opt.Address, time.Duration(opt.Timeout)*time.Second)
	if err != nil {
		klog.Errorf("Get zookeeper client has an error: %v", err)
	}

	if err != nil || conn == nil {
		return nil, fmt.Errorf("get zookeeper client fail or client is nil, err:%+v", err)
	}

	return &ConfigClient{
		conn:          conn,
		out:           make(chan *types.ConfigEvent),
		configEntries: make(map[string]*types.ConfiguratorConfig),
		rootPathCache: nil,
	}, nil
}

func (cc *ConfigClient) Start() error {
	// Initializing a configuration for the service without a configurator
	// cc.configEntries[constant.DefaultConfigName] = defaultConfig

	rpc, err := zookeeper.NewPathCache(cc.conn, zookeeper.ConfiguratorPath, "CONFIGURATION", true)
	if err != nil {
		return err
	}
	cc.rootPathCache = rpc
	go cc.eventLoop()

	// FIXME just for debug
	var enablePrint = false
	if enablePrint {
		go func() {
			tick := time.Tick(10 * time.Second)
			for {
				select {
				case <-tick:
					klog.Infof("Observing cache of configuration client\n  flags: %v\n  configs: %v",
						cc.rootPathCache.Cached, cc.configEntries)
				}
			}
		}()
	}

	return nil
}

// eventLoop
func (cc *ConfigClient) eventLoop() {
	for event := range cc.rootPathCache.Events() {
		var data []byte
		var ce *types.ConfigEvent
		switch event.EventType {
		case zookeeper.PathCacheEventAdded:
			data = cc.getData(event.Path)
			config := &types.ConfiguratorConfig{}
			err := yaml.Unmarshal([]byte(data), config)
			if err != nil {
				klog.Errorf("Parsing the configuration data to a defined struct has an error: %v", err)
				continue
			}

			cc.configEntries[config.Key] = config
			ce = &types.ConfigEvent{
				EventType:   types.ConfigEntryAdded,
				Path:        event.Path,
				ConfigEntry: config,
			}
			go cc.notify(ce)
			break
		case zookeeper.PathCacheEventChanged:
			data = cc.getData(event.Path)
			config := &types.ConfiguratorConfig{}
			err := yaml.Unmarshal([]byte(data), config)
			if err != nil {
				klog.Errorf("Parsing the configuration data to a defined struct has an error: %v", err)
				continue
			}
			cc.configEntries[config.Key] = config
			ce = &types.ConfigEvent{
				EventType:   types.ConfigEntryChanged,
				Path:        event.Path,
				ConfigEntry: config,
			}
			go cc.notify(ce)
			break
		case zookeeper.PathCacheEventDeleted:
			// TODO Deleting configurations about this service in the CR
			cc.rootPathCache.Cached[event.Path] = false
			delete(cc.configEntries, utils.ResolveServiceName(event.Path))
			ce = &types.ConfigEvent{
				EventType: types.ConfigEntryDeleted,
				Path:      event.Path,
			}
			go cc.notify(ce)
			break
		default:
			klog.Warningf("can not support event type yet: %v", event.EventType)
		}
	}
}

func (cc *ConfigClient) Events() <-chan *types.ConfigEvent {
	return cc.out
}

// getData
func (cc *ConfigClient) getData(path string) []byte {
	data, _, err := cc.conn.Get(path)
	if err != nil {
		klog.Errorf("Get data with path %s has an error: %v", path, err)
		return data
	}

	// klog.Infof("Get data with path %s: \n%v", path, data)
	return data
}

// Find the configurator from the caches for this service,
// return a nil value if there is no result matches this service.
func (cc *ConfigClient) FindConfiguratorConfig(serviceName string) *types.ConfiguratorConfig {
	return cc.configEntries[serviceName]
}

func (cc *ConfigClient) Stop() error {
	return nil
}

// notify
func (cc *ConfigClient) notify(event *types.ConfigEvent) {
	cc.out <- event
}
