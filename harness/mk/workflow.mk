# harness/mk/workflow.mk — Pins der Aktionen im CI-Workflow (SPEC-217 Nr. 8).
# `workflow-check` hängt an GATE_CHECKS; der Root-Aggregator fährt es via make gates.
# Es fährt das gepinnte d-check aus d-check.mk (DCHECK_REF) mit eigener Konfiguration, netzlos,
# mit dem Repo als schreibgeschütztem Mount.
# Grenze: Auslöser, Rechte und Aufruf des Workflows prüft das Ziel nicht (SPEC-217 Nr. 11 c);
# dass die Zeile GATE_CHECKS steht, fängt kein Fall (Nr. 11 d).
.PHONY: workflow-check

workflow-check: ## Gate: Aktionen des Workflows auf voller SHA mit Tag-Kommentar, gültiges YAML, nicht leere Prüfmenge (SPEC-217)
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --config .d-check-workflows.yml

GATE_CHECKS += workflow-check
