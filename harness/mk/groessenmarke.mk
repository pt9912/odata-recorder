# harness/mk/groessenmarke.mk — Größenmarke jedes Slice-Plans in Arbeit (SPEC-128, ADR-0009).
# `groessenmarke` hängt an GATE_CHECKS. Der Host ruft `git` (tools/harness/groessenmarke.sh), das
# Programm der Stufe groessenmarke des Dockerfile läuft im Container ohne Netz und ohne Mount
# (tools/harness/groessenmarke-docker.sh). Netz braucht nur die Stufe deps des Dockerfile.
DOCKER ?= docker
DOCKER_BUILD ?= $(DOCKER) build --progress=plain
GROESSENMARKE_PLAENE ?= docs/plan/planning/in-progress

.PHONY: groessenmarke
groessenmarke: ## Gate: Größenmarke der Pläne unter in-progress/ mit Block (Marke gleich Formel, Messung auf oder unter der Marke); grün ohne Slice in Arbeit (SPEC-128)
	$(DOCKER_BUILD) --target groessenmarke -t odata-recorder:groessenmarke .
	sh tools/harness/groessenmarke.sh "$(CURDIR)/tools/harness/groessenmarke-docker.sh" $(GROESSENMARKE_PLAENE)

GATE_CHECKS += groessenmarke
