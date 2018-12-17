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
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.CreateAddresses(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.CreateAddressCallCount()).To(Equal(1))
	g.Expect(fakeRouter.CreateAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-some-svc-80",
		Port: 80,
		IP:   "1.2.3.4",
	}))
}

func TestCreateAddressesWithRulesAlreadyAdded(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name:        "test-some-svc-80",
			Port:        80,
			IP:          "1.2.3.4",
			SourceRange: "any",
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
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.CreateAddresses(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.CreateAddressCallCount()).To(Equal(0))
	g.Expect(fakeRouter.DeleteAddressCallCount()).To(Equal(0))
}

func TestCreateAddressesWithOptions(t *testing.T) {
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
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
			Options: map[string]string{
				"unifi-site": "some-site",
			},
		},
	}
	err := r.CreateAddresses(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.ListAddressesCallCount()).To(Equal(1))
	g.Expect(fakeRouter.ListAddressesArgsForCall(0)).To(Equal(map[string]string{
		"unifi-site": "some-site",
	}))
	g.Expect(fakeRouter.CreateAddressCallCount()).To(Equal(1))
	g.Expect(fakeRouter.CreateAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-some-svc-80",
		Port: 80,
		IP:   "1.2.3.4",
		Options: map[string]string{
			"unifi-site": "some-site",
		},
	}))
}

func TestDeleteAddresses(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name:        "test-some-svc-80",
			Port:        80,
			IP:          "1.2.3.4",
			SourceRange: "any",
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
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.DeleteAddresses(extraRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.DeleteAddressCallCount()).To(Equal(1))
	g.Expect(fakeRouter.DeleteAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-some-svc-80",
		Port: 80,
		IP:   "1.2.3.4",
	}))
}

func TestCreateAddressesWithUpdatedRules(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name:        "test-some-svc-8080",
			Port:        8080,
			IP:          "5.6.7.8",
			SourceRange: "any",
		},
		{
			Name:        "test-some-svc-443",
			Port:        443,
			IP:          "1.2.3.4",
			SourceRange: "10.0.0.0/16",
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
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
		{
			Name: "some-svc",
			Port: 443,
			IP:   "1.2.3.4",
		},
	}
	err := r.CreateAddresses(desiredRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.DeleteAddressCallCount()).To(Equal(2))
	g.Expect(fakeRouter.DeleteAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-some-svc-8080",
		Port: 8080,
		IP:   "5.6.7.8",
	}))
	g.Expect(fakeRouter.DeleteAddressArgsForCall(1)).To(Equal(forwarding.Address{
		Name:        "test-some-svc-443",
		Port:        443,
		IP:          "1.2.3.4",
		SourceRange: "10.0.0.0/16",
	}))
	g.Expect(fakeRouter.CreateAddressCallCount()).To(Equal(2))
	g.Expect(fakeRouter.CreateAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-some-svc-80",
		Port: 80,
		IP:   "1.2.3.4",
	}))
	g.Expect(fakeRouter.CreateAddressArgsForCall(1)).To(Equal(forwarding.Address{
		Name: "test-some-svc-443",
		Port: 443,
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
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.DeleteAddresses(extraRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.DeleteAddressCallCount()).To(Equal(0))
}

func TestDeleteAddressesWithOptions(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeRouter := &forwardingfakes.FakeRouterClient{}
	fakeRouter.ListAddressesReturns([]forwarding.Address{
		{
			Name:        "test-some-svc-80",
			Port:        80,
			IP:          "1.2.3.4",
			SourceRange: "any",
			Options: map[string]string{
				"unifi-site": "some-site",
			},
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
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
			Options: map[string]string{
				"unifi-site": "some-site",
			},
		},
	}
	err := r.DeleteAddresses(extraRules)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fakeRouter.ListAddressesCallCount()).To(Equal(1))
	g.Expect(fakeRouter.ListAddressesArgsForCall(0)).To(Equal(map[string]string{
		"unifi-site": "some-site",
	}))
	g.Expect(fakeRouter.DeleteAddressCallCount()).To(Equal(1))
	g.Expect(fakeRouter.DeleteAddressArgsForCall(0)).To(Equal(forwarding.Address{
		Name: "test-some-svc-80",
		Port: 80,
		IP:   "1.2.3.4",
		Options: map[string]string{
			"unifi-site": "some-site",
		},
	}))
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
			Name:        "test-some-svc-80",
			Port:        80,
			IP:          "1.2.3.4",
			SourceRange: "any",
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
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}
	err := r.DeleteAddresses(extraRules)
	g.Expect(err).To(MatchError("some-error"))
}
