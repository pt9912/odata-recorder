# harness/mk/test-contract.mk — Referenzanfragen und Negativfälle am HTTP-Eingang im Container ohne Netz
# (SPEC-042, ADR-0005). EIN KOMMANDO, KEIN GATE: das Ziel baut ein Image und startet je Test-Binary einen
# Container; es hängt nicht an GATE_CHECKS. Die Aufnahme in die Gate-Kette ist die eine Zeile
# `GATE_CHECKS += test-contract` in dieser Datei.
DOCKER ?= docker
VERSION ?= dev

.PHONY: test-contract

test-contract: ## 36 Fixtures ODV4-* über HTTP und je Fehlerklasse von LH-QA-05 ein Negativfall (SPEC-042) — KEIN Gate
	DOCKER="$(DOCKER)" VERSION="$(VERSION)" sh tools/harness/test-contract.sh
