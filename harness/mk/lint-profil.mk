# harness/mk/lint-profil.mk — Lint-Profil .golangci.yml (ADR-0003, SPEC-101).
# `lint-profil` prüft das Profil gegen das Schema von golangci-lint, `lint` den Code (SPEC-101).
# Kein Bind-Mount: das Profil geht per `docker cp` in den Container, der ohne Netz läuft
# und immer entfernt wird. Die Hebung des Pins ist ein bewusster Commit.
GOLANGCI_LINT_IMAGE ?= golangci/golangci-lint:v2.14.0@sha256:ad862ba6b3798cbe0fd9fd7408d498fd74fbd2623a92406b2fd3898faf0bf98f

.PHONY: lint-profil
lint-profil: ## Gate: .golangci.yml gegen das Schema von golangci-lint prüfen (netzlos, ohne Mount)
	@c=$$(docker create --network none $(GOLANGCI_LINT_IMAGE) golangci-lint config verify -c /tmp/.golangci.yml) && \
	trap 'docker rm -f $$c >/dev/null' EXIT && \
	docker cp "$(CURDIR)/.golangci.yml" $$c:/tmp/.golangci.yml && \
	docker start -a $$c

DOCKER ?= docker
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: lint
lint: ## Gate: Go-Code nach .golangci.yml und eigene Prüfungen (Dockerfile, Stufe lint; netzlos ausser deps)
	$(DOCKER_BUILD) --build-arg LINT_IMAGE=$(GOLANGCI_LINT_IMAGE) --target lint -t odata-recorder:lint .

GATE_CHECKS += lint-profil lint
