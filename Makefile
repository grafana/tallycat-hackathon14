DOCKER_HUB_USER = shantanua
IMAGE_NAME = tallycat
TAG ?= latest
FULL_IMAGE_NAME = $(DOCKER_HUB_USER)/$(IMAGE_NAME):$(TAG)

.PHONY: help build push build-push clean

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build:
	docker build --platform linux/amd64 -t $(FULL_IMAGE_NAME) .

push:
	docker push $(FULL_IMAGE_NAME)

build-push: build push

run:
	docker run -p 4000:4000 -p 8080:8080 $(FULL_IMAGE_NAME)

clean: ## Remove local Docker image
	docker rmi $(FULL_IMAGE_NAME) 2>/dev/null || true