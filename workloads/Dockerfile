FROM --platform=${BUILDPLATFORM} golang:alpine AS builder
ARG TARGETOS TARGETARCH
RUN apk --no-cache --update add make
WORKDIR /workspace
COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH \
    make build

FROM --platform=${BUILDPLATFORM} alpine AS ennovation
WORKDIR /workspace
RUN apk --no-cache --update add dumb-init
RUN apk --no-cache --update add ca-certificates
COPY --from=builder /workspace/bin/* /bin/
