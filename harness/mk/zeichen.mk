# harness/mk/zeichen.mk — unsichtbare Zeichen in den Dateien des Arbeitsbaums (SPEC-233).
# `zeichen-check` hängt an GATE_CHECKS. Der Host wählt die Dateien mit `git` und übergibt sie an den Container
# (tools/harness/zeichen.sh); das Programm der Stufe zeichen des Dockerfile läuft ohne Netz und ohne Mount.
# Netz braucht nur die Stufe deps des Dockerfile.
DOCKER ?= docker
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: zeichen-check
zeichen-check: ## Gate: kein BOM, NBSP, Zeilentrenner, Nullbreiten- oder Steuerzeichen und kein ungültiges UTF-8 in den Dateien des Arbeitsbaums außer .harness/baseline (SPEC-233) · seit welle-sensoren
	$(DOCKER_BUILD) --target zeichen -t odata-recorder:zeichen .
	bash tools/harness/zeichen.sh

GATE_CHECKS += zeichen-check
