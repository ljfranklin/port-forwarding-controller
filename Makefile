
IMG ?= ljfranklin/port-forwarding-controller
IMG_TAG ?= latest

all: dev-setup test manager

# Install deps
dev-setup:
	./scripts/dev-setup.sh

# Run tests
test: generate fmt vet manifests
	KUBEBUILDER_ASSETS=$(shell pwd)/bin go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/ljfranklin/port-forwarding-controller/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	./bin/kustomize build config/default | kubectl apply -f -

# Delete controller in the configured Kubernetes cluster in ~/.kube/config
delete-deployment: manifests
	./bin/kustomize build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build:
	IMAGE=$(IMG) ./scripts/docker-build.sh

# Push the docker image
docker-push:
	docker push $(IMG):amd64-$(IMG_TAG)
	docker push $(IMG):arm32v6-$(IMG_TAG)
	docker push $(IMG):arm64v8-$(IMG_TAG)
	docker manifest create --amend $(IMG):$(IMG_TAG) $(IMG):amd64-$(IMG_TAG) $(IMG):arm32v6-$(IMG_TAG) $(IMG):arm64v8-$(IMG_TAG)
	docker manifest annotate $(IMG):$(IMG_TAG) $(IMG):arm32v6-$(IMG_TAG) --os linux --arch arm
	docker manifest annotate $(IMG):$(IMG_TAG) $(IMG):arm64v8-$(IMG_TAG) --os linux --arch arm64 --variant armv8
	docker manifest push --purge $(IMG):$(IMG_TAG)
