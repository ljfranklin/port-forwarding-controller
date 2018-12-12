package service

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/ljfranklin/port-forwarding-controller/pkg/forwarding"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//go:generate counterfeiter . PortForwardingReconciler
type PortForwardingReconciler interface {
	Reconcile([]forwarding.Address) error
}

// Add creates a new Service Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this core.Add(mgr) to install this Controller
func Add(mgr manager.Manager, pfr PortForwardingReconciler) error {
	return add(mgr, NewReconciler(mgr, pfr))
}

func AddWithReconciler(mgr manager.Manager, r reconcile.Reconciler) error {
	return add(mgr, r)
}

// NewReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager, pfr PortForwardingReconciler) reconcile.Reconciler {
	return &ReconcileService{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		pfr:    pfr,
		log:    logf.Log.WithName("service-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("service-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Service - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1.Service{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileService{}

// ReconcileService reconciles a Service object
type ReconcileService struct {
	client.Client
	scheme *runtime.Scheme
	pfr    PortForwardingReconciler
	log    logr.Logger
}

// Reconcile reads that state of the cluster for a Service object and makes changes based on the state read
// and what is in the Service.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Service instance
	instance := &corev1.Service{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	if instance.Spec.Type == "LoadBalancer" {
		for _, port := range instance.Spec.Ports {
			// r.log.Info("Service", "Name", instance.Name, "IP", instance.Spec.LoadBalancerIP, "Ports", ports, "NodePorts", nodePorts)
			r.pfr.Reconcile([]forwarding.Address{
				{
					Name: instance.Name,
					Port: int(port.Port),
					IP:   instance.Spec.LoadBalancerIP,
				},
			})
		}
	}

	return reconcile.Result{}, nil
}
