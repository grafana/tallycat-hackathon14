DOCKER_IMAGE_NAME := tallycat-hackathon
DOCKER_IMAGE_TAG := 0.0.1


.PHONY: build
build:
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
