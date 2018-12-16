package forwarding

import (
	"fmt"
)

type Address struct {
	Name        string
	Port        int
	IP          string
	SourceRange string
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

func (r Reconciler) CreateAddresses(addresses []Address) error {
	for i := range addresses {
		addresses[i].Name = fmt.Sprintf("%s%s", r.RulePrefix, addresses[i].Name)
	}

	existingAddresses, err := r.RouterClient.ListAddresses()
	if err != nil {
		return err
	}

	missingAddresses := r.missingAddresses(addresses, existingAddresses)

	for _, address := range missingAddresses {
		r.Logger.Info("adding port forwarding rule", "name", address.Name, "port", address.Port, "ip", address.IP)
		if err := r.RouterClient.CreateAddress(address); err != nil {
			return err
		}
	}

	return nil
}

func (r Reconciler) DeleteAddresses(addresses []Address) error {
	for i := range addresses {
		addresses[i].Name = fmt.Sprintf("%s%s", r.RulePrefix, addresses[i].Name)
	}

	existingAddresses, err := r.RouterClient.ListAddresses()
	if err != nil {
		return err
	}

	addressesToDelete := r.addressesToDelete(addresses, existingAddresses)

	for _, address := range addressesToDelete {
		r.Logger.Info("deleting port forwarding rule", "name", address.Name, "port", address.Port, "ip", address.IP)
		if err := r.RouterClient.DeleteAddress(address); err != nil {
			return err
		}
	}
	return nil
}

func (r Reconciler) missingAddresses(desiredAddresses, existingAddresses []Address) []Address {
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

	return missingAddresses
}

func (r Reconciler) addressesToDelete(removedAddresses, existingAddresses []Address) []Address {
	addressesToDelete := []Address{}
	for _, removedAddress := range removedAddresses {
		for _, existingAddress := range existingAddresses {
			if removedAddress == existingAddress {
				addressesToDelete = append(addressesToDelete, removedAddress)
			}
		}
	}

	return addressesToDelete
}
