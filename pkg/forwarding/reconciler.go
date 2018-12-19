package forwarding

import (
	"fmt"
	"reflect"
	"strings"
)

type Address struct {
	Name        string
	Port        int
	IP          string
	SourceRange string
	Options     map[string]string
}

//go:generate counterfeiter . RouterClient
type RouterClient interface {
	ListAddresses(map[string]string) ([]Address, error)
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
	addresses = r.addAddressPrefix(addresses)
	options := r.getListOptions(addresses)

	existingAddresses, err := r.listExistingAddresses(options)
	if err != nil {
		return err
	}

	staleAddresses := r.staleAddresses(addresses, existingAddresses)
	for _, address := range staleAddresses {
		r.Logger.Info("deleting stale port forwarding rule", "name", address.Name, "port", address.Port, "ip", address.IP)
		if err := r.RouterClient.DeleteAddress(address); err != nil {
			return err
		}
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
	addresses = r.addAddressPrefix(addresses)
	options := r.getListOptions(addresses)

	existingAddresses, err := r.listExistingAddresses(options)
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

func (r Reconciler) addAddressPrefix(addresses []Address) []Address {
	updated := make([]Address, len(addresses))
	for i, addr := range addresses {
		addr.Name = fmt.Sprintf("%s%s", r.RulePrefix, addr.Name)
		updated[i] = addr
	}
	return updated
}

func (r Reconciler) getListOptions(addresses []Address) map[string]string {
	options := map[string]string{}
	if len(addresses) > 0 {
		// assumes options are the same for each port
		options = addresses[0].Options
	}
	return options
}

func (r Reconciler) listExistingAddresses(options map[string]string) ([]Address, error) {
	existingAddresses, err := r.RouterClient.ListAddresses(options)
	if err != nil {
		return nil, err
	}
	for i := range existingAddresses {
		if existingAddresses[i].SourceRange == "any" {
			existingAddresses[i].SourceRange = ""
		}
	}
	return existingAddresses, nil
}

func (r Reconciler) missingAddresses(desiredAddresses, existingAddresses []Address) []Address {
	missingAddresses := []Address{}

	for _, address := range desiredAddresses {
		address.Name = fmt.Sprintf("%s-%d", address.Name, address.Port)
		alreadyExists := false
		for _, a := range existingAddresses {
			if reflect.DeepEqual(address, a) {
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

func (r Reconciler) staleAddresses(desiredAddresses, existingAddresses []Address) []Address {
	staleAddresses := []Address{}

	// if existingAddresses exists which:
	// - shares the same name prefix as a desiredAddress
	// - but ports/src do not match any in desired addresses
	// then delete as stale
	desiredRulePrefixes := []string{}
	for _, a := range desiredAddresses {
		desiredRulePrefixes = append(desiredRulePrefixes, a.Name)
	}
	for _, address := range existingAddresses {
		matchesPrefix := false
		for _, prefix := range desiredRulePrefixes {
			if strings.HasPrefix(address.Name, prefix) {
				matchesPrefix = true
				break
			}
		}

		if matchesPrefix {
			foundMatch := false
			for _, a := range desiredAddresses {
				if strings.HasPrefix(address.Name, a.Name) {
					a.Name = fmt.Sprintf("%s-%d", a.Name, a.Port)
					if reflect.DeepEqual(address, a) {
						foundMatch = true
						break
					}
				}
			}
			if !foundMatch {
				staleAddresses = append(staleAddresses, address)
			}
		}
	}

	return staleAddresses
}

func (r Reconciler) addressesToDelete(removedAddresses, existingAddresses []Address) []Address {
	addressesToDelete := []Address{}
	for _, removedAddress := range removedAddresses {
		removedAddress.Name = fmt.Sprintf("%s-%d", removedAddress.Name, removedAddress.Port)
		for _, existingAddress := range existingAddresses {
			if reflect.DeepEqual(removedAddress, existingAddress) {
				addressesToDelete = append(addressesToDelete, removedAddress)
			}
		}
	}

	return addressesToDelete
}
