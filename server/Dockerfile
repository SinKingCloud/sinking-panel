# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS build

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache \
    binutils \
    build-base \
    file \
    libseccomp-dev \
    libseccomp-static \
    linux-headers \
    pkgconf

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    test "$TARGETOS" = "linux" && \
    CGO_ENABLED=1 GOOS="$TARGETOS" GOARCH="$TARGETARCH" go build \
        -buildvcs=false \
        -mod=readonly \
        -trimpath \
        -tags seccomp,netgo,osusergo \
        -ldflags='-s -w -linkmode external -extldflags "-static"' \
        -o /out/server .

# Fail the build if a dependency reintroduces a dynamic runtime library.
RUN file /out/server | grep -q "statically linked" && \
    readelf -l /out/server > /tmp/program-headers && \
    ! grep -q "INTERP" /tmp/program-headers && \
    readelf -d /out/server > /tmp/dynamic-section && \
    ! grep -q "NEEDED" /tmp/dynamic-section && \
    /out/server > /dev/null 2>&1

FROM scratch AS artifact

COPY --from=build /out/server /server
