# harness/mk/katalog.mk — Kataloge der Fehlercodes, Schlüssel und Kommandos gegen Spezifikation und Handbuch (SPEC-201).
# `katalog-check` hängt an GATE_CHECKS und baut die Stufe katalog des Dockerfile; ihr RUN fährt das Programm
# aus tools/katalog ohne Netz und ohne Mount. Netz braucht nur die Stufe deps des Dockerfile.
DOCKER ?= docker
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: katalog-check
katalog-check: ## Gate: katalog hält Fehlercodes, Schlüssel und Kommandos des Quelltexts gegen die Spezifikation, die Ausnahmen und das Handbuch (SPEC-201 · seit slice-handbuch-meldungskatalog; Dockerfile, Stufe katalog; netzlos ausser deps)
	$(DOCKER_BUILD) --target katalog .

GATE_CHECKS += katalog-check
