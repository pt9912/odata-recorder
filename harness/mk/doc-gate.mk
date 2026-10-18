# harness/mk/doc-gate.mk — Doc-Gate-Fragment, emittiert von ai-harness-init.
# Bindet das tool-generierte d-check.mk ein (Befund-Gate docs-check) und haengt
# docs-check an GATE_CHECKS an; der Root-Aggregator faehrt es via make gates.
include d-check.mk

# VORLAUF-WAECHTER fuer die zwei history-lesenden Targets: ueber einer aufloesbaren,
# aber LEEREN Commit-Range melden doc-immutable und doc-commits am emittierten
# .d-check.yml "0 Befund(e)", Exit 0 — gruen ueber leerem Pruefbereich. Beide haengen
# darum an history-range-guard, der VOR dem Modul-Lauf mit einer Meldung abbricht; das
# Rezept der beiden Targets bleibt das aus d-check.mk, und die Emission prueft, dass
# diese Datei die zwei Targets fuehrt.
#
# FAIL-CLOSED, je Ziel: die Vorbindung setzt voraus, dass das eingebundene d-check.mk das
# Ziel MIT REZEPT fuehrt. Die Probe steht EINMAL (DOC_GATE_ZIEL) und wird je Ziel mit seinem
# Namen aufgerufen; nur der belegte Ausgang waehlt die Bindung, jeder andere — kein Treffer,
# eine Ziel-Zeile ohne Rezept, kein Probe-Werkzeug (awk) oder eine leere Ausgabe — faellt in
# den Abbruch mit Exit 2. Eine Vorbindung ueber einem Ziel ohne Rezept laesst make mit Exit 0
# enden und den Waechter allein laufen; dagegen steht diese Bedingung. Gelesen wird dieselbe
# Datei, die das include oben einbindet (awk, kein Bild, kein Netz).
#
# KEIN GATE: die Range setzt der Aufrufer, ohne sie ist der Pruefbereich nicht
# hermetisch — das Ziel steht darum nicht in GATE_CHECKS.
.PHONY: history-range-guard

history-range-guard: ## Vorlauf-Waechter: RANGE muss aufloesbar UND nicht leer sein (STAGED=1 prueft den Index; den STAGED-Zweig fuehrt nur doc-immutable)
	@bash tools/harness/history-range-guard.sh "$(if $(STAGED),--staged,$(RANGE))"

# DOC_GATE_ZIEL <ziel> — die EINE Probe: "da", wenn d-check.mk das Ziel mit einer
# Rezept-Zeile fuehrt, "rezeptlos", wenn es die Ziel-Zeile ohne Rezept traegt. Kein Treffer,
# ein fehlendes Probe-Werkzeug und jede leere Ausgabe lassen den Vergleich scheitern.
# Die Zuweisung traegt override: die Kommandozeile setzt diese Variable nicht, ein Aufruf wie
#   make DOC_GATE_ZIEL=da <ziel>
# aendert die Entscheidung also nicht. Rekursiv zugewiesen (nicht :=), sonst expandiert $(1)
# schon hier.
override DOC_GATE_ZIEL = $(shell awk '/^$(1):/{f=1;next} f&&/^[[:space:]]*$$/{next} f{print (substr($$0,1,1)=="\t" ? "da" : "rezeptlos");exit}' d-check.mk 2>/dev/null)

ifeq ($(call DOC_GATE_ZIEL,doc-immutable),da)
doc-immutable: history-range-guard
else
doc-immutable:
	@echo "harness/mk/doc-gate.mk: d-check.mk fuehrt 'doc-immutable' nicht als Ziel mit Rezept (oder awk fehlt) — die Vorbindung des Vorlauf-Waechters haette dort kein Rezept." >&2
	@exit 2
endif

ifeq ($(call DOC_GATE_ZIEL,doc-commits),da)
doc-commits: history-range-guard
else
doc-commits:
	@echo "harness/mk/doc-gate.mk: d-check.mk fuehrt 'doc-commits' nicht als Ziel mit Rezept (oder awk fehlt) — die Vorbindung des Vorlauf-Waechters haette dort kein Rezept." >&2
	@exit 2
endif

GATE_CHECKS += docs-check
