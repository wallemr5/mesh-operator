/*
Copyright 2020 The Symcn Authors.
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package appmeshconfig

import (
	"context"
	"testing"

	meshv1 "github.com/mesh-operator/pkg/apis/mesh/v1"
	"github.com/mesh-operator/pkg/option"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestReconcileAppMeshConfig_reconcileVirtualService(t *testing.T) {
	fakeScheme := GetFakeScheme()
	type fields struct {
		client     client.Client
		scheme     *runtime.Scheme
		opt        *option.ControllerOption
		meshConfig *meshv1.MeshConfig
	}
	type args struct {
		ctx context.Context
		cr  *meshv1.AppMeshConfig
		svc *meshv1.Service
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test-reconcile-virtualservice-create-ok",
			fields: fields{
				client:     GetFakeClient(TestMeshConfig),
				scheme:     fakeScheme,
				opt:        TestOpt,
				meshConfig: TestMeshConfig,
			},
			args: args{
				ctx: context.Background(),
				cr:  amcTestOK,
				svc: testOKService,
			},
			wantErr: false,
		},
		{
			name: "test-reconcile-virtualservice-update-ok",
			fields: fields{
				client:     GetFakeClient(TestMeshConfig, amcTestOK, fakeVirtualService),
				scheme:     fakeScheme,
				opt:        TestOpt,
				meshConfig: TestMeshConfig,
			},
			args: args{
				ctx: context.Background(),
				cr:  amcTestOK,
				svc: testUpdateOKService,
			},
			wantErr: false,
		},
		{
			name: "test-reconcile-virtualservice-delete-ok",
			fields: fields{
				client:     GetFakeClient(TestMeshConfig, amcTestOK, fakeVirtualService, fakeDeleteVirtualService),
				scheme:     fakeScheme,
				opt:        TestOpt,
				meshConfig: TestMeshConfig,
			},
			args: args{
				ctx: context.Background(),
				cr:  amcTestOK,
				svc: testUpdateOKService,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileAppMeshConfig{
				client:     tt.fields.client,
				scheme:     tt.fields.scheme,
				opt:        tt.fields.opt,
				meshConfig: tt.fields.meshConfig,
			}
			if err := r.reconcileVirtualService(tt.args.ctx, tt.args.cr, tt.args.svc); (err != nil) != tt.wantErr {
				t.Errorf("ReconcileAppMeshConfig.reconcileVirtualService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
