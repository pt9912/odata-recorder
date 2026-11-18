# harness/mk/test-integration.mk — Integrationstests des gebauten Binary im Container ohne Netz
# (SPEC-037, SPEC-211). `test-integration` hängt an GATE_CHECKS: das Ziel baut die Stufe integration und
# startet je Stufe einen Container. Dass diese Zeile steht, fängt kein Fall (SPEC-211 Nr. 5 a).
DOCKER ?= docker
VERSION ?= dev

.PHONY: test-integration

test-integration: ## Gate: Record, Upstream stoppen, Replay mit dem gebauten Binary im Container ohne Netz (SPEC-037, SPEC-211)
	DOCKER="$(DOCKER)" VERSION="$(VERSION)" sh tools/harness/test-integration.sh

GATE_CHECKS += test-integration
