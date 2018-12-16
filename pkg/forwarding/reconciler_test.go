package forwarding_test

import (
	"errors"
	"testing"

	"github.com/ljfranklin/port-forwarding-controller/pkg/forwarding"
	"github.com/ljfranklin/port-forwarding-controller/pkg/forwarding/forwardingfakes"
	. "github.com/onsi/gomega"
)

func TestCreateAddressesWithNoExistingRules(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{}, nil)
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	desiredRules := []forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.CreateAddresses(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.CreateAddressCallCount()).To(Equal(1))
	g.Expect(fakeRouter.CreateAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-port-1",
		Port: 80,
		IP:   "1.2.3.4",
	}))
}

func TestCreateAddressesWithRulesAlreadyAdded(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name: "test-port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}, nil)
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	desiredRules := []forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.CreateAddresses(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.CreateAddressCallCount()).To(Equal(0))
}

func TestDeleteAddresses(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name: "test-port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}, nil)
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	extraRules := []forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.DeleteAddresses(extraRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.DeleteAddressCallCount()).To(Equal(1))
	g.Expect(fakeRouter.DeleteAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-port-1",
		Port: 80,
		IP:   "1.2.3.4",
	}))
}

func TestDeleteAddressesWithNoExistingRules(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{}, nil)
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	extraRules := []forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.DeleteAddresses(extraRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.DeleteAddressCallCount()).To(Equal(0))
}

func TestCreateAddressesWithListError(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns(nil, errors.New("some-error"))
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	desiredRules := []forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.CreateAddresses(desiredRules)
	g.Expect(err).To(MatchError("some-error"))
}

func TestCreateAddressesWithCreateError(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{}, nil)
	fakeRouter.CreateAddressReturns(errors.New("some-error"))
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	desiredRules := []forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.CreateAddresses(desiredRules)
	g.Expect(err).To(MatchError("some-error"))
}

func TestDeleteAddressesWithListError(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{}, errors.New("some-error"))
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	extraRules := []forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.DeleteAddresses(extraRules)
	g.Expect(err).To(MatchError("some-error"))
}

func TestDeleteAddressesWithDeleteError(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name: "test-port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}, nil)
	fakeRouter.DeleteAddressReturns(errors.New("some-error"))
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	extraRules := []forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.DeleteAddresses(extraRules)
	g.Expect(err).To(MatchError("some-error"))
}
