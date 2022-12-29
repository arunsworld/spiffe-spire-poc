IMAGE_NAME := arunsworld/spiffe-poc
LD_FLAGS := -s -w
TARGETS := ./client:./server

build:
	pack build ${IMAGE_NAME}:latest \
		--default-process server \
		--env "BP_GO_TARGETS=${TARGETS}" \
		--env "BP_GO_BUILD_LDFLAGS=${LD_FLAGS}" \
		--buildpack gcr.io/paketo-buildpacks/go \
		--builder paketobuildpacks/builder:tiny

