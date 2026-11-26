# harness/mk/lasttest.mk — Lasttest nach LH-QA-01.a (SPEC-216): ein Werkzeug, kein Gate; das Ziel hängt nicht an GATE_CHECKS und läuft
# nicht in `make gates`. Der Treiber liest LASTTEST_BETRIEBSART, LASTTEST_PROFIL, LASTTEST_ZIEL_MS und LASTTEST_STARTWERT aus der Umgebung
# (make exportiert Werte der Kommandozeile); das Fragment setzt keine Vorgaben. Netz braucht nur die Stufe deps des Dockerfile.
DOCKER ?= docker

.PHONY: lasttest
lasttest: ## Werkzeug: Lasttest nach LH-QA-01.a im Container mit 2 CPU und 4 GiB ohne Netz (SPEC-216; Profil rauch: LASTTEST_PROFIL=rauch) — KEIN Gate
	DOCKER="$(DOCKER)" sh tools/harness/lasttest.sh
