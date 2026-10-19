# Dockerfile — Build und Test des OData Recorder in Stufen.
# Das Eingangs-Image ist per Digest gepinnt; eine Anhebung ist ein bewusster Commit.
# Netz braucht nur die Stufe deps; alle Stufen danach laufen mit `RUN --network=none`.

# Das Image der Stufe lint kommt aus harness/mk/lint-profil.mk (Build-Argument LINT_IMAGE).
ARG LINT_IMAGE=scratch

# --- deps: Module aus go.mod und go.sum (go.sum entsteht mit der ersten Abhängigkeit),
# `go mod verify` prüft sie dagegen; /go/pkg/mod liegt auch ohne Abhängigkeit. Die Stufe und alle von ihr abgeleiteten laufen auf der Plattform
# des Bau-Hosts.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS deps
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly
COPY go.mod go.sum* ./
RUN go mod download && go mod verify && mkdir -p /go/pkg/mod

# --- source: der Quellstand aus src/ (go.mod liegt aus deps schon da).
FROM deps AS source
COPY src ./src

# --- test: Formatierung (gofmt nennt die nicht formatierten Dateien), vet und
# Unit-Tests.
FROM source AS test
RUN --network=none f="$(gofmt -l .)" && { [ -z "$f" ] || { echo "nicht gofmt-formatiert:"; echo "$f"; exit 1; }; } \
 && go vet ./... \
 && go test -trimpath -buildvcs=false ./...

# --- build: das Binary für die Zielplattform, reproduzierbar (-trimpath, ohne
# VCS-Stempel).
FROM source AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN --network=none GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -buildvcs=false \
      -ldflags="-s -w -X main.version=$VERSION" \
      -o /out/odata-recorder ./src/bootstrap/main

# --- lint: golangci-lint nach .golangci.yml und die eigenen Prüfungen von tools/harness/lint.sh
# auf dem Quellstand; die Module kommen aus deps.
FROM ${LINT_IMAGE} AS lint
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY --from=source /src ./
COPY .golangci.yml ./
COPY tools/harness/lint.sh /usr/local/bin/lint.sh
RUN --network=none sh /usr/local/bin/lint.sh
