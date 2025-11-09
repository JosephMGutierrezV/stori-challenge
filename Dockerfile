ARG GO_VERSION=1.25
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-bookworm AS build

WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache \
    go env -w GOPROXY=https://proxy.golang.org,direct

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download -x

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=unknown
ARG COMMIT=unknown

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath \
      -ldflags "-s -w -X main.version=$VERSION -X main.commit=$COMMIT" \
      -o /out/main ./cmd/lambda_api/main.go


FROM public.ecr.aws/lambda/go:1

COPY --from=build /out/main ${LAMBDA_TASK_ROOT}/main

CMD ["main"]
