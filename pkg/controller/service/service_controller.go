package service

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/ljfranklin/port-forwarding-controller/pkg/forwarding"
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
	CreateAddresses([]forwarding.Address) error
	DeleteAddresses([]forwarding.Address) error
}

// Add creates a new Service Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
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
	if r.isAnnotatedLB(instance) {
		addresses := []forwarding.Address{}
		var targetIP string
		var sourceRange string
		if instance.Spec.Type == "LoadBalancer" {
			targetIP = instance.Spec.LoadBalancerIP
			if len(instance.Spec.LoadBalancerSourceRanges) > 0 {
				sourceRange = instance.Spec.LoadBalancerSourceRanges[0]
			}
		} else {
			targetIP = instance.Spec.ExternalIPs[0]
		}
		for _, port := range instance.Spec.Ports {
			addresses = append(addresses, forwarding.Address{
				Name:        fmt.Sprintf("%s-%s", instance.ObjectMeta.Namespace, instance.Name),
				Port:        int(port.Port),
				IP:          targetIP,
				SourceRange: sourceRange,
			})
		}

		finalizerName := "finalizer.port-forwarding.lylefranklin.com/v1"

		if instance.ObjectMeta.DeletionTimestamp.IsZero() {
			err = r.pfr.CreateAddresses(addresses)
			if err != nil {
				return reconcile.Result{}, err
			}
			if !containsString(instance.ObjectMeta.Finalizers, finalizerName) {
				instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
				if err = r.Update(context.Background(), instance); err != nil {
					return reconcile.Result{}, err
				}
			}
		} else {
			if containsString(instance.ObjectMeta.Finalizers, finalizerName) {
				err = r.pfr.DeleteAddresses(addresses)
				if err != nil {
					return reconcile.Result{}, err
				}

				instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizerName)
				if err = r.Update(context.Background(), instance); err != nil {
					return reconcile.Result{}, err
				}
			}
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileService) isAnnotatedLB(instance *corev1.Service) bool {
	if instance.Spec.Type == "LoadBalancer" || (instance.Spec.Type == "NodePort" && len(instance.Spec.ExternalIPs) > 0) {
		for key, value := range instance.ObjectMeta.Annotations {
			if key == "port-forward.lylefranklin.com/enable" && value == "true" {
				return true
			}
		}
	}
	return false
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
