package servicemeshentry

import (
	"context"

	meshv1 "github.com/mesh-operator/pkg/apis/mesh/v1"
	"github.com/mesh-operator/pkg/option"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	controllerName = "servicemeshentry-controller"
	httpRouteName  = "dubbo-http-route"
	proxyRouteName = "dubbo-proxy-route"
)

var log = logf.Log.WithName("controller_servicemeshentry")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ServiceMeshEntry Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, opt *option.ControllerOption) error {
	return add(mgr, newReconciler(mgr, opt), opt)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, opt *option.ControllerOption) reconcile.Reconciler {
	return &ReconcileServiceMeshEntry{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		opt:    opt,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, opt *option.ControllerOption) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{
		Reconciler:              r,
		MaxConcurrentReconciles: opt.MaxConcurrentReconciles,
	})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ServiceMeshEntry
	err = c.Watch(&source.Kind{Type: &meshv1.ServiceMeshEntry{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{
		Type: &networkingv1beta1.WorkloadEntry{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &meshv1.ServiceMeshEntry{},
		})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{
		Type: &networkingv1beta1.VirtualService{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &meshv1.ServiceMeshEntry{},
		})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{
		Type: &networkingv1beta1.DestinationRule{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &meshv1.ServiceMeshEntry{},
		})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{
		Type: &networkingv1beta1.ServiceEntry{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &meshv1.ServiceMeshEntry{},
		})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileServiceMeshEntry implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileServiceMeshEntry{}

// ReconcileServiceMeshEntry reconciles a ServiceMeshEntry object
type ReconcileServiceMeshEntry struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	opt        *option.ControllerOption
	meshConfig *meshv1.MeshConfig
}

// Reconcile reads that state of the cluster for a ServiceMeshEntry object and makes changes based on the state read
// and what is in the ServiceMeshEntry.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileServiceMeshEntry) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	klog.Infof("Reconciling ServiceMeshEntry: %s/%s", request.Namespace, request.Name)
	ctx := context.TODO()

	// Fetch the MeshConfig
	err := r.getMeshConfig(ctx)
	if err != nil {
		klog.Errorf("Get cluster MeshConfig[%s/%s] error: %+v",
			r.opt.MeshConfigNamespace, r.opt.MeshConfigName, err)
		return reconcile.Result{}, err
	}

	// Fetch the ServiceMeshEntry instance
	instance := &meshv1.ServiceMeshEntry{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request foundect not found, could have been deleted after reconcile request.
			// Owned foundects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.Infof("Can't found ServiceMeshEntry[%s/%s], requeue...", request.Namespace, request.Name)
			return reconcile.Result{}, nil
		}
		// Error reading the foundect - requeue the request.
		return reconcile.Result{}, err
	}

	// Distribute Istio Config
	if err := r.reconcileWorkloadEntry(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.reconcileServiceEntry(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.reconcileDestinationRule(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.reconcileVirtualService(ctx, instance); err != nil {
		return reconcile.Result{}, err
	}

	// Update Status
	klog.Infof("Update ServiceMeshEntry[%s/%s] status...", request.Namespace, request.Name)
	err = r.updateStatus(ctx, request, instance)
	if err != nil {
		klog.Errorf("%s/%s update ServiceMeshEntry failed, err: %+v", request.Namespace, request.Name, err)
		return reconcile.Result{}, err
	}

	klog.Infof("End Reconciliation, ServiceMeshEntry: %s/%s.", request.Namespace, request.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileServiceMeshEntry) getMeshConfig(ctx context.Context) error {
	meshConfig := &meshv1.MeshConfig{}
	err := r.client.Get(
		ctx,
		types.NamespacedName{
			Namespace: r.opt.MeshConfigNamespace,
			Name:      r.opt.MeshConfigName,
		},
		meshConfig,
	)
	if err != nil {
		return err
	}
	r.meshConfig = meshConfig
	klog.V(4).Infof("Get cluster MeshConfig: %+v", meshConfig)
	return nil
}
