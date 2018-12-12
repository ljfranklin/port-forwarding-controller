package controller

import (
	"github.com/ljfranklin/port-forwarding-controller/pkg/controller/service"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, service.PortForwardingReconciler) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, pfr service.PortForwardingReconciler) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, pfr); err != nil {
			return err
		}
	}
	return nil
}
