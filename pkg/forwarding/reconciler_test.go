package forwarding_test

import (
	"errors"
	"testing"

	"github.com/ljfranklin/port-forwarding-controller/pkg/forwarding"
	"github.com/ljfranklin/port-forwarding-controller/pkg/forwarding/forwardingfakes"
	. "github.com/onsi/gomega"
)

func TestReconcileWithNoExistingRules(t *testing.T) {
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
	err := r.Reconcile(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.CreateAddressCallCount()).To(Equal(1))
	g.Expect(fakeRouter.CreateAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-port-1",
		Port: 80,
		IP:   "1.2.3.4",
	}))
}

func TestReconcileWithRulesAlreadyAdded(t *testing.T) {
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
	err := r.Reconcile(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.CreateAddressCallCount()).To(Equal(0))
}

func TestReconcileWithExtraRules(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name: "test-port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
		{
			Name: "non-matching-prefix-port-1",
			Port: 443,
			IP:   "5.6.7.8",
		},
	}, nil)
	fakeLogger := &forwardingfakes.FakeInfoLogger{}

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
		RulePrefix:   "test-",
		Logger:       fakeLogger,
	}

	desiredRules := []forwarding.Address{}
	err := r.Reconcile(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.DeleteAddressCallCount()).To(Equal(1))
	g.Expect(fakeRouter.DeleteAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-port-1",
		Port: 80,
		IP:   "1.2.3.4",
	}))
}

func TestReconcileWithListError(t *testing.T) {
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
	err := r.Reconcile(desiredRules)
	g.Expect(err).To(MatchError("some-error"))
}

func TestReconcileWithCreateError(t *testing.T) {
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
	err := r.Reconcile(desiredRules)
	g.Expect(err).To(MatchError("some-error"))
}
