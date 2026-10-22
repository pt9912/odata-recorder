# Dockerfile — Build und Test des OData Recorder in Stufen.
# Das Eingangs-Image ist per Digest gepinnt; eine Anhebung ist ein bewusster Commit.
# Netz braucht nur die Stufe deps; alle Stufen danach laufen mit `RUN --network=none`.

# Das Image der Stufe lint kommt aus harness/mk/lint-profil.mk (Build-Argument LINT_IMAGE), das
# der Stufe race aus harness/mk/go-gates.mk (RACE_IMAGE).
ARG LINT_IMAGE=scratch
ARG A_CHECK_IMAGE=scratch
ARG RACE_IMAGE=scratch

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

# --- pruefung: das Programm aus tools/pruefung trägt die Prüfungen der Stufen test und build
# (SPEC-036). Es wird geprüft (gofmt, vet, Test) und gebaut.
FROM deps AS pruefung
WORKDIR /pf
COPY tools/pruefung ./
RUN --network=none f="$(gofmt -l .)" && { [ -z "$f" ] || { echo "nicht gofmt-formatiert:"; echo "$f"; exit 1; }; } \
 && go vet ./... && go test ./... && go build -trimpath -buildvcs=false -o /out/pruefung .

# --- test: Formatierung (gofmt nennt die nicht formatierten Dateien), vet und
# Unit-Tests, gefahren von pruefung.
FROM source AS test
COPY --from=pruefung /out/pruefung /usr/local/bin/pruefung
RUN --network=none pruefung test /src

# --- build: das Binary für die Zielplattform, reproduzierbar (-trimpath, ohne
# VCS-Stempel); pruefung prüft Version und -trimpath.
FROM source AS build
COPY --from=pruefung /out/pruefung /usr/local/bin/pruefung
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN --network=none GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -buildvcs=false \
      -ldflags="-s -w -X main.version=$VERSION" \
      -o /out/odata-recorder ./src/bootstrap/main
RUN --network=none pruefung build /out/odata-recorder "$VERSION"

# --- race: go test -race mit cgo über den Quellstand der Stufe source, gefahren von pruefung
# (SPEC-122). RACE_IMAGE bringt den C-Compiler mit; die Stufe bricht ab, wenn seine Go-Version
# von der der Stufe deps abweicht. Umgebung und Netz wie bei deps und lint: cgo schaltet allein
# `pruefung race` für seinen Aufruf ein, die übrigen Stufen bleiben bei CGO_ENABLED=0.
FROM --platform=$BUILDPLATFORM ${RACE_IMAGE} AS race
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly
COPY --from=deps /usr/local/go/VERSION /deps-go-version
RUN --network=none r="$(head -n 1 /usr/local/go/VERSION)" && d="$(head -n 1 /deps-go-version)" \
 && { [ "$r" = "$d" ] || { echo "Go-Version von race ($r) weicht von deps ($d) ab"; exit 1; }; }
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY --from=source /src ./
COPY --from=pruefung /out/pruefung /usr/local/bin/pruefung
RUN --network=none pruefung race /src

# --- lint: golangci-lint nach .golangci.yml und die eigenen Prüfungen von tools/harness/lint.sh
# auf dem Quellstand und auf tools/gegenproben und tools/pruefung; die Module kommen aus deps.
FROM ${LINT_IMAGE} AS lint
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY --from=source /src ./
COPY .golangci.yml ./
COPY tools/gegenproben tools/gegenproben
COPY tools/pruefung tools/pruefung
COPY tools/harness/lint.sh /usr/local/bin/lint.sh
RUN --network=none sh /usr/local/bin/lint.sh

# --- a-check-bin: das Binary von a-check, wird in die Stufe gegenproben kopiert.
FROM ${A_CHECK_IMAGE} AS a-check-bin

# --- gegenproben: das Programm aus tools/gegenproben prüft, dass a-check, lint.sh,
# migration-check.sh und pruefung einen Verstoß rot melden (SPEC-102). Es wird geprüft (gofmt, vet, Test), gebaut und gegen die
# Eingaben gefahren.
FROM ${LINT_IMAGE} AS gegenproben
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly
COPY --from=a-check-bin /a-check /usr/local/bin/a-check
COPY --from=pruefung /out/pruefung /usr/local/bin/pruefung
WORKDIR /gp
COPY tools/gegenproben ./
RUN --network=none f="$(gofmt -l .)" && { [ -z "$f" ] || { echo "nicht gofmt-formatiert:"; echo "$f"; exit 1; }; } \
 && go vet ./... && go test ./... && go build -trimpath -o /usr/local/bin/gegenproben .
WORKDIR /src
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY go.mod go.sum* .a-check.yml .golangci.yml ./
COPY src ./src
COPY tools/harness/lint.sh tools/harness/migration-check.sh tools/harness/
RUN --network=none gegenproben
