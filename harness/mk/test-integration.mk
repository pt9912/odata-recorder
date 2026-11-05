# harness/mk/test-integration.mk — Integrationstests des gebauten Binary im Container ohne Netz
# (SPEC-037, ADR-0005). EIN KOMMANDO, KEIN GATE: das Ziel baut ein Image und startet je Stufe einen
# Container; es hängt nicht an GATE_CHECKS. Die Aufnahme in die Gate-Kette ist die eine Zeile
# `GATE_CHECKS += test-integration` in dieser Datei.
DOCKER ?= docker
VERSION ?= dev

.PHONY: test-integration

test-integration: ## Record, Upstream stoppen, Replay mit dem gebauten Binary im Container ohne Netz (SPEC-037) — KEIN Gate
	DOCKER="$(DOCKER)" VERSION="$(VERSION)" sh tools/harness/test-integration.sh
