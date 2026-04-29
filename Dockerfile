# https://www.bytesizego.com/blog/production-go-docker-image
# https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/

FROM --platform=$BUILDPLATFORM golang:1.26.2-trixie AS builder

COPY . /usr/src/subsonic-sanitizer
WORKDIR /usr/src/subsonic-sanitizer

ARG TARGETOS TARGETARCH

ENV CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2 -fstack-protector-all"
ENV GOFLAGS="-buildmode=pie"

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-s -w" -trimpath -o=/bin/subsonic-sanitizer ./cmd/server.go

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=builder /bin/subsonic-sanitizer /bin/subsonic-sanitizer

USER 65534

CMD ["/bin/subsonic-sanitizer"]
