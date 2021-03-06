/*
Copyright 2020 The symcn authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"github.com/mesh-operator/pkg/adapter"
	"github.com/mesh-operator/pkg/adapter/options"
	"github.com/mesh-operator/pkg/option"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// NewAdapterCmd ...
func NewAdapterCmd(ropt *option.RootOption) *cobra.Command {

	opt := options.DefaultOption()
	cmd := &cobra.Command{
		Use:     "adapter",
		Aliases: []string{"adapter"},
		Short:   "Adapters configured for different registry center",
		Run: func(cmd *cobra.Command, args []string) {
			PrintFlags(cmd.Flags())
			opt.EventHandlers.Kubeconfig = ropt.Kubeconfig
			opt.EventHandlers.ConfigContext = ropt.ConfigContext
			_, err := adapter.NewAdapter(opt)
			if err != nil {
				klog.Fatalf("unable to NewAdapter err: %v", err)
			}
		},
	}

	cmd.PersistentFlags().StringArrayVar(
		&opt.Registry.Address,
		"raddr",
		//"r",
		opt.Registry.Address,
		"address for registry center, e.g. zk: 127.0.0.1:2181")

	cmd.PersistentFlags().Int64Var(
		&opt.Registry.Timeout,
		"rto",
		opt.Registry.Timeout,
		"the zookeeper session timeout second for registry")

	cmd.PersistentFlags().StringArrayVar(
		&opt.Configuration.Address,
		"caddr",
		opt.Configuration.Address,
		"address for configuration center, e.g. zk: 127.0.0.1:2181")

	cmd.PersistentFlags().Int64Var(
		&opt.Configuration.Timeout,
		"--cto",
		opt.Configuration.Timeout,
		"the zookeeper session timeout second for configuration center")

	cmd.PersistentFlags().StringVar(
		&opt.EventHandlers.ClusterOwner,
		"cluster-owner",
		opt.EventHandlers.ClusterOwner,
		"the labels that multiple cluster manager used for select clusters")

	cmd.PersistentFlags().StringVar(
		&opt.EventHandlers.ClusterNamespace,
		"cluster-namespace",
		opt.EventHandlers.ClusterNamespace,
		"the namespace that multiple cluster manager uses when selecting the cluster config maps")
	return cmd
}
