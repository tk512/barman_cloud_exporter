DOCKER_ARCHS ?= amd64 armv7 arm64

all: vet

include Makefile.common

STATICCHECK_IGNORE =

DOCKER_IMAGE_NAME ?= barman-cloud-exporter

.PHONY: test-docker-single-exporter
test-docker-single-exporter:
	@echo ">> testing docker image for single exporter"
	./test_image.sh "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" 9104

.PHONY: test-docker
