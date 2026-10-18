# harness/mk/traeger.mk — Fragment des Traeger-Fetch, emittiert von ai-harness-init.
# EIN KOMMANDO, KEIN GATE.
#
# Es liegt im Fragment-Verzeichnis wie die Gate-Fragmente, haengt aber NICHTS an
# GATE_CHECKS und steht in keiner Prerequisite-Kette — auch nicht an archive-welle:
# dessen Fehlt-Fall ist die Zusage "Exit 0, nennt das Fehlende, schreibt nichts", und
# ein Fetch davor wuerde sie still aendern.
# Der Fetch ist der AUSDRUECKLICHE Weg aus dem Zustand, den jene Meldung benennt.
#
# Der Traeger liegt im gitignorierten Zustands-Bereich: ein frischer Klon hat ihn
# nicht. Dieses Kommando holt ihn aus dem gepinnten Release nach — das Asset wird
# gegen den SHA256SUMS-Eintrag desselben Releases verifiziert VOR der Ablage, eine
# Abweichung bricht ab, ohne den Traeger zu legen; der
# Transport laeuft im gepinnten Bild, nicht auf dem Host. Netz braucht
# er an genau diesem Aufruf; er laeuft nur auf ausdruecklichem Aufruf.
.PHONY: traeger-fetch

# DER PIN: das
# Fragment fuehrt nur den Release-Tag — keinen Wert, der vom Bau-Ergebnis abhaengt
# (keine Digest-Variable, kein Export).
# Die Verifizierung liest der Fetch aus der SHA256SUMS desselben Releases; einen
# zweiten Kanal mit Einzeldigests fuehrt das Ziel nicht.
# Der Tag ist eine Release-Entscheidung: er wandert mit dem
# Release-Schnitt, nicht mit jedem Bau.
TRAEGER_TAG ?= v0.9.0
TRAEGER_CARRIER ?= .harness/state/bin/ai-harness-init

export TRAEGER_TAG TRAEGER_CARRIER

traeger-fetch: ## Traeger aus dem gepinnten Release nachholen (braucht Netz) — NICHT in gates
	@bash tools/harness/traeger-fetch.sh