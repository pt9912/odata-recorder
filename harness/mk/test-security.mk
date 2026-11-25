# harness/mk/test-security.mk — Sicherheitsnachweis im Container ohne Netz (SPEC-041, SPEC-211): Canary-Werte am
# gebauten Binary und Netzwerkzugriff auf den Admin-Listener. `test-security` hängt an GATE_CHECKS: das Ziel baut die
# Stufe security und startet je Test-Binary einen Container. Dass diese Zeile steht, fängt kein Fall (SPEC-041 Nr. 10 h).
DOCKER ?= docker
VERSION ?= dev

.PHONY: test-security

test-security: ## Gate: Canary-Werte in keinem Artefakt, Admin-Endpunkte ohne Token abgewiesen, Admin-Listener nur über Loopback oder TLS (SPEC-041, SPEC-211)
	DOCKER="$(DOCKER)" VERSION="$(VERSION)" sh tools/harness/test-security.sh

GATE_CHECKS += test-security
