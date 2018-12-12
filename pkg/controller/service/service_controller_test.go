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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	fakePFR := &servicefakes.FakePortForwardingReconciler{}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(service.NewReconciler(mgr, fakePFR))
	g.Expect(service.AddWithReconciler(mgr, recFn)).NotTo(HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the Service object and expect the Reconcile
	instance := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-svc",
			Namespace: "default",
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
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Fatalf("failed to create object, got an invalid object error: %v", err)
	}
	g.Expect(err).NotTo(HaveOccurred())
	defer c.Delete(context.TODO(), instance)

	expectedRequest := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "some-svc",
			Namespace: "default",
		},
	}
	g.Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

	g.Expect(fakePFR.ReconcileCallCount()).To(Equal(1))
	g.Expect(fakePFR.ReconcileArgsForCall(0)).To(Equal([]forwarding.Address{
		{
			Name: "some-svc",
			Port: 80,
			IP:   "1.2.3.4",
		},
	}))
}
