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

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
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
		Name: "port-1",
		Port: 80,
		IP:   "1.2.3.4",
	}))
}

func TestReconcileWithRulesAlreadyAdded(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name: "port-1",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}, nil)

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
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

func TestReconcileWithListError(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns(nil, errors.New("some-error"))

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
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

	r := forwarding.Reconciler{
		RouterClient: fakeRouter,
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
