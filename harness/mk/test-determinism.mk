# harness/mk/test-determinism.mk — 100 Läufe des Replay nach Reset im Container ohne Netz (SPEC-038,
# ADR-0005). EIN KOMMANDO, KEIN GATE: das Ziel baut ein Image und startet einen Container; es hängt
# nicht an GATE_CHECKS. Die Aufnahme in die Gate-Kette ist die eine Zeile
# `GATE_CHECKS += test-determinism` in dieser Datei.
DOCKER ?= docker
VERSION ?= dev

.PHONY: test-determinism

test-determinism: ## 100 Läufe des Replay nach Reset, gleiche normalisierte Antworten und Cursor-Stände (SPEC-038) — KEIN Gate
	DOCKER="$(DOCKER)" VERSION="$(VERSION)" sh tools/harness/test-determinism.sh
