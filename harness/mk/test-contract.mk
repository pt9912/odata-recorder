# harness/mk/test-contract.mk — Referenzanfragen und Negativfälle am HTTP-Eingang im Container ohne Netz
# (SPEC-042, SPEC-211). `test-contract` hängt an GATE_CHECKS: das Ziel baut die Stufe contract und startet
# je Test-Binary einen Container. Dass diese Zeile steht, fängt kein Fall (SPEC-211 Nr. 5 a).
DOCKER ?= docker
VERSION ?= dev

.PHONY: test-contract

test-contract: ## Gate: 36 Fixtures ODV4-* über HTTP und je Fehlerklasse von LH-QA-05 ein Negativfall (SPEC-042, SPEC-211)
	DOCKER="$(DOCKER)" VERSION="$(VERSION)" sh tools/harness/test-contract.sh

GATE_CHECKS += test-contract
