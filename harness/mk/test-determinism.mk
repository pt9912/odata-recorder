# harness/mk/test-determinism.mk — 100 Läufe des Replay nach Reset im Container ohne Netz (SPEC-038,
# SPEC-211). `test-determinism` hängt an GATE_CHECKS: das Ziel baut die Stufe integration und startet
# einen Container. Dass diese Zeile steht, fängt kein Fall (SPEC-211 Nr. 5 a).
DOCKER ?= docker
VERSION ?= dev

.PHONY: test-determinism

test-determinism: ## Gate: 100 Läufe des Replay nach Reset, gleiche normalisierte Antworten und Cursor-Stände (SPEC-038, SPEC-211)
	DOCKER="$(DOCKER)" VERSION="$(VERSION)" sh tools/harness/test-determinism.sh

GATE_CHECKS += test-determinism
