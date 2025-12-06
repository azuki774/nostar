FROM golang:1.25.4 AS builder

WORKDIR /workspace

RUN --mount=type=bind,source=.,target=/workspace \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -a -tags netgo \
    -ldflags "-extldflags '-static' -s -w \
    -X main.version=$(git describe --tag --abbrev=0 2>/dev/null || echo 'dev') \
    -X main.revision=$(git rev-list -1 HEAD 2>/dev/null || echo 'unknown') \
    -X main.build=$(git describe --tags 2>/dev/null || echo 'dev')" \
    -o /bin/nostar ./

FROM gcr.io/distroless/static-debian12:nonroot
EXPOSE 9999

COPY --from=builder --chmod=555 /bin/nostar /nostar
COPY ./bin/config.toml /config.toml
USER nonroot:nonroot

# ENV DATABASE_URL=postgres://postgres:postgres@localhost:5432/nostar

ENTRYPOINT ["/nostar", "serve"]
CMD ["-c", "/config.toml"]
