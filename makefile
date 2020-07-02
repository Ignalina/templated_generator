include TopMakefile

.PHONY: dockerize push dockerize-and-push

SHELL=/bin/bash -o pipefail
IMAGE = GoWebHDFS-monolith
SHA = $(shell git rev-parse --short HEAD)
DOCKER = podman

ifeq ($(origin PRIVATE_REGISTRY),undefined)
PRIVATE_REGISTRY := $(shell minikube ip 2>/dev/null):5000
endif

ifneq ($(PRIVATE_REGISTRY),)
	PREFIX:=${PRIVATE_REGISTRY}/
endif

dockerize-and-push: dockerize push

dockerize:
	@echo "[${DOCKER} build] building ${IMAGE} (tags: ${PREFIX}${IMAGE}:latest, ${PREFIX}${IMAGE}:${SHA})"
	@${DOCKER} build --file ./Dockerfile \
		--tag ${PREFIX}${IMAGE}:latest \
		--tag ${PREFIX}${IMAGE}:${SHA} \
		../../ 2>&1 | sed -e "s/^/ | /g"

push:
	@echo "[${DOCKER} push] pushing ${PREFIX}${IMAGE}:latest"
	@${DOCKER} push ${PREFIX}${IMAGE}:latest 2>&1 | sed -e "s/^/ | /g"
	@echo "[${DOCKER} push] pushing ${PREFIX}${IMAGE}:${SHA}"
	@${DOCKER} push ${PREFIX}${IMAGE}:${SHA} 2>&1 | sed -e "s/^/ | /g"
