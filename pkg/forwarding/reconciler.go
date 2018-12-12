package forwarding

type Address struct {
	Name       string
	Port       int
	IP         string
	RulePrefix string
}

type Reconciler struct{}

func (r Reconciler) Reconcile(addresses []Address) error {
	return nil
}
