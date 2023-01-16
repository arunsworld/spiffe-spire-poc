IMAGE_NAME := arunsworld/spiffe-poc
LD_FLAGS := -s -w

PLATFORMS ?= linux/amd64,linux/arm64

binaries := client server webserver

.PHONY: build
build: $(addprefix bin/,$(binaries))

go_build := go build -ldflags '$(LD_FLAGS)' -o

bin/%: ./%
	@echo Building $@...
	$(go_build) $@ ./$<

.PHONY: container-builder
container-builder:
	docker buildx create --platform $(PLATFORMS) --name container-builder --node container-builder0 --use

.PHONY: images
images: Dockerfile container-builder
			echo Building docker image ennovation $(PLATFORM)…
			docker buildx build \
				--platform $(PLATFORMS) \
				--target ennovation \
				-o type=registry -t $(IMAGE_NAME) \
				-f Dockerfile \
				.

.PHONY: pack-build
pack-build:
	pack build ${IMAGE_NAME}:latest \
		--default-process server \
		--env "BP_GO_TARGETS=${TARGETS}" \
		--env "BP_GO_BUILD_LDFLAGS=${LD_FLAGS}" \
		--buildpack gcr.io/paketo-buildpacks/go \
		--builder paketobuildpacks/builder:tiny

