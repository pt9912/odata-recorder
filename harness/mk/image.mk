# harness/mk/image.mk — Laufzeit-Image (SPEC-322) und sein Neustarttest (SPEC-323). `test-image` hängt an GATE_CHECKS, `image` nicht.
# Der Neustarttest startet Container im Docker-Daemon des Hosts; Namen und Label trägt jeder Lauf selbst (SPEC-323 Nr. 2 b). Dass die
# Zeile GATE_CHECKS steht, fängt kein Fall (SPEC-323 Nr. 9 d).
DOCKER ?= docker
VERSION ?= dev
DOCKER_BUILD ?= $(DOCKER) build --progress=plain

.PHONY: image test-image

image: ## Laufzeit-Image odata-recorder:image bauen (Dockerfile, Stufe image; SPEC-322) — KEIN Gate
	$(DOCKER_BUILD) --target image --build-arg VERSION=$(VERSION) -t odata-recorder:image .

test-image: ## Gate: Laufzeit-Image prüfen, aufnehmen, Container stoppen und starten, alles wieder lesen, zwei Bauten byte-gleich (SPEC-323)
	DOCKER="$(DOCKER)" VERSION="$(VERSION)" sh tools/harness/test-image.sh

GATE_CHECKS += test-image
