# harness/mk/schema.mk — neutrales Schema der Persistenz (tools/schema/schema.yaml) mit
# d-migrate. Beide Targets sind Werkzeuge und hängen an keinem GATE_CHECKS-Eintrag: sie
# brauchen das d-migrate-Image lokal. Die Hebung des Pins ist ein bewusster Commit.
# Kein Bind-Mount: das Schema geht per `docker cp` in den Container, das Ergebnis kommt als
# Tar-Strom auf stdout heraus; der Container läuft ohne Netz und wird immer entfernt.
D_MIGRATE_IMAGE ?= ghcr.io/pt9912/d-migrate@sha256:af9d3eb323a6cfd13eb63f012823788e510f764509918933d2ff98c8e68e4212

.PHONY: schema-validate schema-generate
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
