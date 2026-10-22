# harness/mk/go-gates.mk — Build und Test über das Dockerfile (ADR-0001, ADR-0007).
# `build`, `test` und `test-race` hängen an GATE_CHECKS; der Root-Aggregator fährt sie via make gates.
# Netz braucht nur die Stufe deps des Dockerfiles, und nur bei leerem Build-Cache oder
# geänderten Abhängigkeiten.
# RACE_IMAGE trägt den Compiler für cgo (SPEC-122 Nr. 4); die Go-Version stimmt mit der der Stufe deps
# überein, beide werden gemeinsam angehoben, eine Anhebung ist ein bewusster Commit.
DOCKER ?= docker
VERSION ?= dev
DOCKER_BUILD ?= $(DOCKER) build --progress=plain
RACE_IMAGE ?= golang:1.27.1-trixie@sha256:8f58fd67ea075142d947a60e0caa4317746a55118d312f027793d382c7741734

.PHONY: build test test-race

build: ## Binary bauen; pruefung build prüft Version aus VERSION und -trimpath (Dockerfile, Stufe build; netzlos ausser deps)
	$(DOCKER_BUILD) --build-arg VERSION=$(VERSION) --target build -t odata-recorder:build .

test: ## pruefung test: gofmt, vet und Unit-Tests (Dockerfile, Stufe test; netzlos ausser deps)
	$(DOCKER_BUILD) --target test -t odata-recorder:test .

test-race: ## pruefung race: go test -race mit cgo über alle Pakete der Stufe source (Dockerfile, Stufe race; netzlos ausser deps)
	$(DOCKER_BUILD) --build-arg RACE_IMAGE=$(RACE_IMAGE) --target race -t odata-recorder:race .

GATE_CHECKS += build test test-race
