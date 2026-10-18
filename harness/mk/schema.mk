# harness/mk/schema.mk — neutrales Schema der Persistenz (tools/schema/schema.yaml) mit
# d-migrate. schema-validate und schema-generate sind Werkzeuge; schema-check ist das Gate
# (GATE_CHECKS) und braucht beide Images lokal. Die Hebung eines Pins ist ein bewusster Commit.
# Kein Bind-Mount: das Schema geht per `docker cp` in den Container, das Ergebnis kommt als
# Tar-Strom auf stdout heraus; der Container läuft ohne Netz und wird immer entfernt.
D_MIGRATE_IMAGE ?= ghcr.io/pt9912/d-migrate@sha256:af9d3eb323a6cfd13eb63f012823788e510f764509918933d2ff98c8e68e4212
SCHEMA_CHECK_IMAGE ?= python@sha256:a6e34c598f2467ed0e9a8d349809fcd8b5c603269512df273a0bb1784edc11b1

.PHONY: schema-validate schema-generate schema-check
schema-validate: ## d-migrate: neutrales Schema prüfen (netzlos, ohne Mount, kein Gate)
	@c=$$(docker create --network none $(D_MIGRATE_IMAGE) schema validate --source /work/schema.yaml) && \
	trap 'docker rm -f $$c >/dev/null' EXIT && \
	docker cp "$(CURDIR)/tools/schema/schema.yaml" $$c:/work/schema.yaml && \
	docker start -a $$c

schema-generate: ## d-migrate: SQLite-DDL und Report als Tar-Strom auf stdout (netzlos, ohne Mount, kein Gate)
	@c=$$(docker create --network none $(D_MIGRATE_IMAGE) schema generate --source /work/schema.yaml --target sqlite --deterministic --output /tmp/out/schema.sql) && \
	trap 'docker rm -f $$c >/dev/null' EXIT && \
	docker cp "$(CURDIR)/tools/schema/schema.yaml" $$c:/work/schema.yaml && \
	docker start -a $$c >&2 && \
	docker cp $$c:/tmp/out -

# Das generierte DDL läuft in SQLite (Python-Standardbibliothek, netzlos); tools/schema/check.py
# hält die Constraint-Zusagen aus SPEC-004 als Negativtests und prüft foreign_key_check und
# integrity_check. Die Migrationskette ist kein Teil davon: sie braucht eine zweite Schemaversion.
schema-check: ## Gate: Schema validieren, DDL generieren, in SQLite laden, Negativtests und Integritaetspruefung
	@$(MAKE) --no-print-directory schema-validate && \
	d=$$(mktemp -d) && c= && trap 'rm -rf "$$d"; [ -z "$$c" ] || docker rm -f $$c >/dev/null' EXIT && \
	$(MAKE) --no-print-directory schema-generate | tar -x -C "$$d" && \
	c=$$(docker create --network none $(SCHEMA_CHECK_IMAGE) python -I /tmp/check.py /tmp/schema.sql) && \
	docker cp "$$d/out/schema.sql" $$c:/tmp/schema.sql && \
	docker cp "$(CURDIR)/tools/schema/check.py" $$c:/tmp/check.py && \
	docker start -a $$c

GATE_CHECKS += schema-check
