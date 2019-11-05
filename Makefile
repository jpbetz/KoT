SHELL = bash

DOCKER_ORG ?= docker.io/$(shell whoami)

all: build
build: build-admission build-conversion
push: push-admission push-conversion

.PHONY: build-conversion
build-conversion:
		docker build . -f Dockerfile-conversion -t $(DOCKER_ORG)/things-conversion-webhook:latest

.PHONY: build-admission
build-admission:
		docker build . -f Dockerfile-admission -t $(DOCKER_ORG)/deepsea-admission-webhook:latest

.PHONY: build-controllers
build-controllers:
		docker build . -f Dockerfile-controllers -t $(DOCKER_ORG)/deepsea-controllers:latest

.PHONY: build-simulator
build-simulator:
		docker build . -f Dockerfile-simulator -t $(DOCKER_ORG)/deepsea-simulator:latest

push-conversion:
		docker push $(DOCKER_ORG)/things-conversion-webhook:latest
		sed 's,image: .*$$,image: $(DOCKER_ORG)/things-conversion-webhook@'$$(docker inspect --format='{{index .RepoDigests 0}}' $(DOCKER_ORG)/things-conversion-webhook | cut -f2 -d@)',' conversion/manifests.yaml > conversion/manifests.yaml.updated && mv conversion/manifests.yaml.updated conversion/manifests.yaml

push-admission:
		docker push $(DOCKER_ORG)/deepsea-admission-webhook:latest
		sed 's,image: .*$$,image: $(DOCKER_ORG)/deepsea-admission-webhook@'$$(docker inspect --format='{{index .RepoDigests 0}}' $(DOCKER_ORG)/deepsea-admission-webhook | cut -f2 -d@)',' admission/manifests.yaml > admission/manifests.yaml.updated && mv admission/manifests.yaml.updated admission/manifests.yaml

push-simulator:
		docker push $(DOCKER_ORG)/deepsea-simulator:latest
		sed 's,image: .*$$,image: $(DOCKER_ORG)/deepsea-simulator@'$$(docker inspect --format='{{index .RepoDigests 0}}' $(DOCKER_ORG)/deepsea-simulator | cut -f2 -d@)',' simulator/manifests.yaml > simulator/manifests.yaml.updated && mv simulator/manifests.yaml.updated simulator/manifests.yaml


clean:
		rm -f tls.key tls.crt
