package forwarding

import (
	"fmt"
	"strings"
)

type Address struct {
	Name string
	Port int
	IP   string
}

//go:generate counterfeiter . RouterClient
type RouterClient interface {
	ListAddresses() ([]Address, error)
	CreateAddress(Address) error
	DeleteAddress(Address) error
}

//go:generate counterfeiter . InfoLogger
type InfoLogger interface {
	Info(msg string, keysAndValues ...interface{})
}

type Reconciler struct {
	RulePrefix   string
	RouterClient RouterClient
	Logger       InfoLogger
}

// TODO: also remove extra rules that are not longer needed that
// start with given RulePrefix

func (r Reconciler) Reconcile(desiredAddresses []Address) error {
	for i := range desiredAddresses {
		desiredAddresses[i].Name = fmt.Sprintf("%s%s", r.RulePrefix, desiredAddresses[i].Name)
	}

	existingAddresses, err := r.RouterClient.ListAddresses()
	if err != nil {
		return err
	}

	missingAddresses, err := r.missingAddresses(desiredAddresses, existingAddresses)
	if err != nil {
		return err
	}

	for _, address := range missingAddresses {
		r.Logger.Info("adding port forwarding rule", "name", address.Name, "port", address.Port, "ip", address.IP)
		if err := r.RouterClient.CreateAddress(address); err != nil {
			return err
		}
	}

	extraAddresses, err := r.extraAddresses(desiredAddresses, existingAddresses)
	if err != nil {
		return err
	}

	for _, address := range extraAddresses {
		r.Logger.Info("deleting port forwarding rule", "name", address.Name, "port", address.Port, "ip", address.IP)
		if err := r.RouterClient.DeleteAddress(address); err != nil {
			return err
		}
	}
	return nil
}

func (r Reconciler) missingAddresses(desiredAddresses, existingAddresses []Address) ([]Address, error) {
	missingAddresses := []Address{}

	for _, address := range desiredAddresses {
		alreadyExists := false
		for _, a := range existingAddresses {
			if address == a {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			missingAddresses = append(missingAddresses, address)
		}
	}

	return missingAddresses, nil
}

func (r Reconciler) extraAddresses(desiredAddresses, existingAddresses []Address) ([]Address, error) {
	extraAddresses := []Address{}

	for _, address := range existingAddresses {
		if !strings.HasPrefix(address.Name, r.RulePrefix) {
			continue
		}
		matchFound := false
		for _, a := range desiredAddresses {
			if address == a {
				matchFound = true
				break
			}
		}
		if !matchFound {
			extraAddresses = append(extraAddresses, address)
		}
	}

	return extraAddresses, nil
}
