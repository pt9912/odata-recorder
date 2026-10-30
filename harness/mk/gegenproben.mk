# harness/mk/gegenproben.mk — Gegenproben zu a-check, lint, migration-check, pruefung (test, build, race, handbuch) und groessenmarke samt Treiber (SPEC-102). Kein Gate: das Ziel
# hängt nicht an GATE_CHECKS. Die Verdrahtung von pruefung und groessenmarke in den Stufen des Dockerfile und der Aufruf von docker im
# Treiber prüfen die Fälle nicht; die zu race laufen auf dem Image der Stufe gegenproben, nicht auf RACE_IMAGE. Eine Prüfsumme der Eingaben (tools/harness/gegenproben-eingabe.sh)
# entscheidet, ob der Lauf nötig ist; der Stempel ist eine versionierte Datei und reist mit dem Commit, auch
# in einen anderen Klon oder Container.
GEGENPROBEN_STEMPEL ?= harness/gegenproben.stempel
DOCKER ?= docker
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: gegenproben
gegenproben: ## Gegenproben zu a-check, lint, migration-check, pruefung (test, build, race, handbuch) und groessenmarke, nur bei geänderter Eingabe (Dockerfile, Stufe gegenproben; netzlos) — KEIN Gate
	@h=$$(bash tools/harness/gegenproben-eingabe.sh "$(GOLANGCI_LINT_IMAGE)" "$(A_CHECK_IMAGE)") && \
	if [ "$$(cat $(GEGENPROBEN_STEMPEL) 2>/dev/null)" = "$$h" ]; then \
	  echo "gegenproben: Eingaben unverändert seit dem letzten grünen Lauf ($$h) — nichts läuft neu"; exit 0; fi && \
	$(DOCKER_BUILD) --no-cache-filter gegenproben --build-arg LINT_IMAGE=$(GOLANGCI_LINT_IMAGE) \
	  --build-arg A_CHECK_IMAGE=$(A_CHECK_IMAGE) --target gegenproben -t odata-recorder:gegenproben . && \
	mkdir -p $(dir $(GEGENPROBEN_STEMPEL)) && printf '%s\n' "$$h" > $(GEGENPROBEN_STEMPEL)
