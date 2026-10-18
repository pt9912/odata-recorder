# a-check.mk — Architektur-Gate via a-check, zum `include` aus
# harness/mk/arch-gate.mk. Erzeugt von `a-check --print-mk`; A_CHECK_IMAGE ist
# manuell auf den Release-Digest gepinnt (Pin-Hebung = bewusster Commit).
#
# Benutzerhandbuch (aufgabenorientiert, deutsch):
#   https://github.com/pt9912/a-check/blob/main/docs/user/benutzerhandbuch.md
#
A_CHECK_IMAGE ?= ghcr.io/pt9912/a-check@sha256:e8208764b119c606c92f82722813386277a65b12812d23b6107ea7a14dc25da1

# Container-Runtime ueber eine Indirektion (podman/nerdctl/docker); wer eine
# eigene Runtime nutzt, definiert sie VOR dem `include`.
DOCKER ?= docker

.PHONY: a-check a-check-graph
a-check: ## Architektur: Hexagon-Regeln via a-check (netzlos, read-only).
	$(DOCKER) run --rm --network none -v "$(CURDIR)":/src:ro $(A_CHECK_IMAGE) /src

a-check-graph: ## Architektur-Graph (Mermaid) aus .a-check.yml auf stdout (read-only, kein Scan).
	$(DOCKER) run --rm --network none -v "$(CURDIR)":/src:ro $(A_CHECK_IMAGE) --print-graph /src
