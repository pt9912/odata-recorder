# harness/mk/handbuch.mk — Benutzerhandbuch gegen das gebaute Binary (SPEC-200).
# `handbuch-check` hängt an GATE_CHECKS und baut die Stufe handbuch des Dockerfile, abgeleitet von
# der Stufe build; ihr RUN läuft ohne Netz (netzlos ausser deps) und taggt nichts Neues. VERSION
# gilt wie in go-gates.mk, damit die Stufe build aus dem Cache kommt.
DOCKER ?= docker
VERSION ?= dev
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: handbuch-check
handbuch-check: ## Gate: pruefung handbuch prüft docs/user/benutzerhandbuch.md und seine Beispiele gegen das gebaute Binary (Dockerfile, Stufe handbuch; netzlos ausser deps)
	$(DOCKER_BUILD) --build-arg VERSION=$(VERSION) --target handbuch .

GATE_CHECKS += handbuch-check
