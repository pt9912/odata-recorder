# harness/mk/go-gates.mk — Build und Test über das Dockerfile (ADR-0001).
# `build` und `test` hängen an GATE_CHECKS; der Root-Aggregator fährt sie via make gates.
# Netz braucht nur die Stufe deps des Dockerfiles, und nur bei leerem Build-Cache oder
# geänderten Abhängigkeiten.
DOCKER ?= docker
VERSION ?= dev
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: build test

build: ## Binary bauen und prüfen: Version aus VERSION, -trimpath (Dockerfile, Stufe build; netzlos ausser deps)
	$(DOCKER_BUILD) --build-arg VERSION=$(VERSION) --target build -t odata-recorder:build .
	@v="$$($(DOCKER) run --rm odata-recorder:build /out/odata-recorder version)" && [ "$$v" = "odata-recorder $(VERSION)" ] || { echo "build: Version des Binary ist '$$v', erwartet 'odata-recorder $(VERSION)'" >&2; exit 1; }
	@$(DOCKER) run --rm odata-recorder:build go version -m /out/odata-recorder | grep -q -- '-trimpath=true' || { echo "build: das Binary ist nicht mit -trimpath gebaut" >&2; exit 1; }

test: ## gofmt, vet und Unit-Tests (Dockerfile, Stufe test; netzlos ausser deps)
	$(DOCKER_BUILD) --target test -t odata-recorder:test .

GATE_CHECKS += build test
