## Port Forwarding Controller

A Kubernetes (k8s) controller which watches for new annotated Services and
automatically creates a corresponding port forwarding rule on your router.

Currently Unifi routers are the only supported router model.

#### Use Case

The expected use case is to create a DNS record which points to your router's
public IP and have a port forwarding rule send the traffic into the k8s cluster.
This controller automatically manages the creation and deletion of port forwarding
rules as Services are created and destroyed.

For example, you might deploy an [nginx-ingress](https://github.com/kubernetes/ingress-nginx)
Service with the following annotation:

```
port-forwarding.lylefranklin.com/enable: "true"
```

The port-forwarding-controller will notice this new annotated Service and add
a port forwarding rule on your router to forward traffic on the given ports to
the IP of the new Service.

#### Supported Service types

The controller supports Services of type [LoadBalancer](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer)
and Services with a non-empty [externalIPs property](https://kubernetes.io/docs/concepts/services-networking/service/#external-ips).
Services must be [annotated](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/)
with `port-forwarding.lylefranklin.com/enable: "true"` for the controller to
manage forwarding rules for that Service.

##### LoadBalancer Service

Pros:
- Traffic is balanced across multiple workers

Cons:
- Requires setting up a bare metal Load Balancer, e.g. [MetalLB](https://metallb.universe.tf/concepts/)

Setup steps for MetalLB + Unifi router:
- Configure each k8s worker machine with a static IP
- Deploy MetalLB as described [here](https://metallb.universe.tf/installation/)
- Create a ConfigMap containing your router's IP address as described [here](https://metallb.universe.tf/configuration/#bgp-configuration)
- Configure your Unifi router to accept BGP peers as described [here](https://help.ubnt.com/hc/en-us/articles/205222990-EdgeRouter-Border-Gateway-Protocol)
- To persist your BGP config across reboots create a [config.gateway.json](https://help.ubnt.com/hc/en-us/articles/215458888-UniFi-USG-Advanced-Configuration#2) on your Unifi Controller
- Deploy a LoadBalancer service with the `port-forwarding` annotation
- Verify the controller has created a new port forwarding rule in the Unifi Controller UI

Explanation:

The MetalLB [Concepts page](https://metallb.universe.tf/concepts/bgp/) does a good job
explaining the underlying BGP concepts and the limitations of it.

##### ClusterIP Service with ExternalIP

Pros:
- Minimal additional setup

Cons:
- All traffic goes to a single worker

Setup steps:
- Configure a k8s worker machine with a static IP
- Create a ClusterIP Service with an `externalIPs` set to the IP address of the k8s worker
- Verify the controller has created a new port forwarding rule on your router

Explanation:

Setting the `externalIPs` field will cause the `kube-proxy` job on each worker node to
start listening on the Service's configure `port` on the host.
If a worker receives a packet on that port and the destination IP field matches
the configured `externalIP` of a given Service, `kube-proxy` will forward traffic to that
Service's pod.
If there are multiple pod replicas, traffic will be load balanced between the pods,
but all traffic will be initially received by a single worker node before being routed
to one of the pod replicas.
This is due to port forwarding rules requiring a one-to-one mapping of port to IP address.

#### Deploying the Controller

##### With Helm

- Deploy Helm/Tiller into your k8s cluster
- Create a `./secrets/router.yml` file with the following fields (additional optional props [here](./chart/values.yml)):
  ```
  router:
    url: $YOUR_ROUTER_URL
    username: $YOUR_ROUTER_USERNAME
    password: $YOUR_ROUTER_PASSWORD
  ```
- Run the following command to install or upgrade the controller:
  ```
  helm upgrade \
    --install \
    --values ./secrets/router.yml \
    port-forwarding \
    https://storage.googleapis.com/ansible-assets/port-forwarding-0.1.0.tgz
  ```

##### Without Helm

- Create a `./secrets/router.env` file with the following contents:
  ```
  ROUTER_URL=$YOUR_ROUTER_URL
  ROUTER_USERNAME=$YOUR_ROUTER_USERNAME
  ROUTER_PASSWORD=$YOUR_ROUTER_PASSWORD
  ```
- Run the following command to install or upgrade the controller:
  ```
  make deploy
  ```

#### Additional options

Additional optional Annotations:
- `port-forwarding.lylefranklin.com/unifi-site: SOME_SITE`, defaults to `default` site

#### Contributing

- Run unit tests: `make test`
- Try out controller in a local container: `make run`
- Build docker images: `make docker-build`
- Push docker images: `make docker-push`
- Create a new controller class: `kubebuilder create api --group workloads --version v1beta1 --kind ContainerSet`
