# harness/mk/go-gates.mk — Build und Test über das Dockerfile (ADR-0001).
# `build` und `test` hängen an GATE_CHECKS; der Root-Aggregator fährt sie via make gates.
# Netz braucht nur die Stufe deps des Dockerfiles, und nur bei leerem Build-Cache oder
# geänderten Abhängigkeiten.
DOCKER ?= docker
VERSION ?= dev
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: build test

build: ## Binary bauen; pruefung build prüft Version aus VERSION und -trimpath (Dockerfile, Stufe build; netzlos ausser deps)
	$(DOCKER_BUILD) --build-arg VERSION=$(VERSION) --target build -t odata-recorder:build .

test: ## pruefung test: gofmt, vet und Unit-Tests (Dockerfile, Stufe test; netzlos ausser deps)
	$(DOCKER_BUILD) --target test -t odata-recorder:test .

GATE_CHECKS += build test
