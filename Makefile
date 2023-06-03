GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_CLEAN=$(GO_CMD) clean
GO_TEST=$(GO_CMD) test

TARGET_KUBELET=kubelet
TARGET_APISERVER=apiserver
TARGET_KUBECTL=kubectl
TARGET_SCHEDULER=scheduler
TARGET_PROXY=kubeproxy
TARGET_AUTO=autoScaler
TARGET_RS=replicaSet
TARGET_SERVERLESS=serverless
.DEFAULT_GOAL := default

# as there is a dir named "test" too, so we need .PHONY to specify this target.
.PHONY:test

all: test master node

master: kubectl apiserver scheduler autoScaler replicaSet serverless

node: kubelet kubeproxy

default: master node

test1:
	go test -v ./test/yaml_test/yaml_test.go
	go test -v ./test/etcd_test/etcd_test.go
	go test -v ./test/container_test/container_test.go
	go test -v ./test/node_test/node_test.go
test:
	go test -v ./test/pod_test/pod_test.go
	go test -v ./test/node_test/node1_test.go
	go test -v ./test/service_test/service_test.go

kubectl:
	$(GO_BUILD) -o ./build/$(GO_KUBECTL) ./cmd/kubectl/kubectl.go

apiserver:
	$(GO_BUILD) -o ./build/$(GO_APISERVER) ./cmd/apiserver/apiserver.go

scheduler:
	$(GO_BUILD) -o ./build/$(GO_SCHEDULER) ./cmd/scheduler/scheduler.go

kubelet:
	$(GO_BUILD) -o ./build/$(GO_KUBELET) ./cmd/kubelet/kubelet.go

kubeproxy:
	$(GO_BUILD) -o ./build/$(GO_PROXY) ./cmd/kubeproxy/kubeproxy.go

autoScaler:
	$(GO_BUILD) -o ./build/$(GO_AUTO) ./cmd/controller/autoScaler/autoScalerController.go

replicaSet:
	$(GO_BUILD) -o ./build/$(GO_RS) ./cmd/controller/replicaSet/replica_controller.go

serverless:
	$(GO_BUILD) -o ./build/$(GO_SERVERLESS) ./cmd/serverless/serverless.go

clean:
	rm -rf ./build

# test-only.
master_start:
#	sudo ./build/apiserver &
#	sudo ./build/scheduler &
#	sudo ./build/kubectl
	sudo /bin/bash -c './build/apiserver &'
	sudo /bin/bash -c './build/scheduler &'
	sudo /bin/bash -c './build/autoScaler &'
	sudo /bin/bash -c './build/replicaSet &'
	sudo /bin/bash -c './build/serverless &'
#	sudo sh -c './build/kubectl &'

node_start:
#	sudo ./build/kubeproxy &
#	sudo ./build/kubelet
	sudo /bin/bash -c './build/kubeproxy &'
#	sudo ./build/kubelet -f ./utils/templates/node_template.yaml
	sudo /bin/bash -c './build/kubelet -f /builds/520021910279/mini-k8s-2023/utils/templates/node_template.yaml &'
