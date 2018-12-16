package service_test

import (
	"testing"
	"time"

	"github.com/ljfranklin/port-forwarding-controller/pkg/controller/service"
	"github.com/ljfranklin/port-forwarding-controller/pkg/controller/service/servicefakes"
	"github.com/ljfranklin/port-forwarding-controller/pkg/forwarding"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

const timeout = time.Second * 5

func startManager(g *GomegaWithT, pfr service.PortForwardingReconciler) (chan reconcile.Request, func()) {
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(HaveOccurred())
	c = mgr.GetClient()

	reconciler := service.NewReconciler(mgr, pfr)
	recFn, requests := SetupTestReconcile(reconciler)
	g.Expect(service.AddWithReconciler(mgr, recFn)).NotTo(HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	return requests, func() {
		close(stopMgr)
		mgrStopped.Wait()
	}
}

func createServiceAndWait(g *GomegaWithT, svc *corev1.Service, requests chan reconcile.Request) {
	err := c.Create(context.TODO(), svc)
	g.Expect(err).NotTo(HaveOccurred())

	expectedRequest := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      svc.Name,
			Namespace: "default",
		},
	}
	g.Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
	// annotated LBs get a finalizer added which will trigger an update
	if len(svc.ObjectMeta.Annotations) > 0 {
		g.Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
	}
}

func deleteServiceAndWait(g *GomegaWithT, svc *corev1.Service, requests chan reconcile.Request) {
	err := c.Delete(context.TODO(), svc)
	g.Expect(err).NotTo(HaveOccurred())

	expectedRequest := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      svc.Name,
			Namespace: "default",
		},
	}
	g.Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
}

func TestReconcileWithLoadBalancer(t *testing.T) {
	g := NewGomegaWithT(t)

	fakePFR := &servicefakes.FakePortForwardingReconciler{}
	requests, shutdown := startManager(g, fakePFR)
	defer shutdown()

	instance := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-svc",
			Namespace: "default",
			Annotations: map[string]string{
				"port-forward.lylefranklin.com/enable": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:           "LoadBalancer",
			LoadBalancerIP: "1.2.3.4",
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}
	createServiceAndWait(g, instance, requests)
	defer c.Delete(context.TODO(), instance)

	// Create may get called a second time after finalizer is added
	g.Expect(fakePFR.CreateAddressesCallCount()).To(BeNumerically(">=", 1))
	g.Expect(fakePFR.CreateAddressesArgsForCall(0)).To(Equal([]forwarding.Address{
		{
			Name: "default-some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}))
}

func TestReconcileWithNodePortAndExternalIP(t *testing.T) {
	g := NewGomegaWithT(t)

	fakePFR := &servicefakes.FakePortForwardingReconciler{}
	requests, shutdown := startManager(g, fakePFR)
	defer shutdown()

	instance := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-svc",
			Namespace: "default",
			Annotations: map[string]string{
				"port-forward.lylefranklin.com/enable": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:        "NodePort",
			ExternalIPs: []string{"1.2.3.4"},
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}
	createServiceAndWait(g, instance, requests)
	defer c.Delete(context.TODO(), instance)

	// Create may get called a second time after finalizer is added
	g.Expect(fakePFR.CreateAddressesCallCount()).To(BeNumerically(">=", 1))
	g.Expect(fakePFR.CreateAddressesArgsForCall(0)).To(Equal([]forwarding.Address{
		{
			Name: "default-some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}))
}

func TestReconcileWithNoAnnotation(t *testing.T) {
	g := NewGomegaWithT(t)

	fakePFR := &servicefakes.FakePortForwardingReconciler{}
	requests, shutdown := startManager(g, fakePFR)
	defer shutdown()

	nonAnnotatedInstance := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "some-ignored-svc",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
		Spec: corev1.ServiceSpec{
			Type:           "LoadBalancer",
			LoadBalancerIP: "5.6.7.8",
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}
	createServiceAndWait(g, nonAnnotatedInstance, requests)
	defer c.Delete(context.TODO(), nonAnnotatedInstance)

	g.Expect(fakePFR.CreateAddressesCallCount()).To(Equal(0))
}

func TestReconcileWithDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	fakePFR := &servicefakes.FakePortForwardingReconciler{}
	requests, shutdown := startManager(g, fakePFR)
	defer shutdown()

	instance := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-svc",
			Namespace: "default",
			Annotations: map[string]string{
				"port-forward.lylefranklin.com/enable": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:           "LoadBalancer",
			LoadBalancerIP: "1.2.3.4",
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}
	createServiceAndWait(g, instance, requests)

	deleteServiceAndWait(g, instance, requests)

	g.Expect(fakePFR.DeleteAddressesCallCount()).To(Equal(1))
	g.Expect(fakePFR.DeleteAddressesArgsForCall(0)).To(Equal([]forwarding.Address{
		{
			Name: "default-some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}))
}
