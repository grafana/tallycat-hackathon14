DOCKER_IMAGE_NAME := tallycat-hackathon
DOCKER_IMAGE_TAG := 0.0.1


.PHONY: build
build:
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

.PHONY: docker-run
docker-run:
	docker run --rm -it -p 8080:8080 -p 4317:4317 $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
