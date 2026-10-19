# Dockerfile — Build und Test des OData Recorder in Stufen.
# Das Eingangs-Image ist per Digest gepinnt; eine Anhebung ist ein bewusster Commit.
# Netz braucht nur die Stufe deps; alle Stufen danach laufen mit `RUN --network=none`.

# Das Image der Stufe lint kommt aus harness/mk/lint-profil.mk (Build-Argument LINT_IMAGE).
ARG LINT_IMAGE=scratch
ARG A_CHECK_IMAGE=scratch

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

# --- a-check-bin: das Binary von a-check, wird in die Stufe gegenproben kopiert.
FROM ${A_CHECK_IMAGE} AS a-check-bin

# --- gegenproben: das Programm aus tools/gegenproben prüft, dass a-check und lint.sh einen
# Verstoß rot melden (SPEC-102). Es wird geprüft (gofmt, vet, Test), gebaut und gegen die
# Eingaben gefahren.
FROM ${LINT_IMAGE} AS gegenproben
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly
COPY --from=a-check-bin /a-check /usr/local/bin/a-check
WORKDIR /gp
COPY tools/gegenproben ./
RUN --network=none f="$(gofmt -l .)" && { [ -z "$f" ] || { echo "nicht gofmt-formatiert:"; echo "$f"; exit 1; }; } \
 && go vet ./... && go test ./... && go build -trimpath -o /usr/local/bin/gegenproben .
WORKDIR /src
COPY go.mod go.sum* .a-check.yml .golangci.yml ./
COPY src ./src
COPY tools/harness/lint.sh tools/harness/lint.sh
RUN --network=none gegenproben
