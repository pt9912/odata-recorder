# harness/mk/selbstpruefung.mk — Fragment der Selbstpruefung der
# Durchsetzungsschicht, emittiert von ai-harness-init. EIN KOMMANDO, KEIN GATE.
#
# WARUM ES AN KEINER GATE-KETTE HAENGT: das Ziel legt fuer diese Pruefung einen
# eigenen Klon an und faehrt darin das Gate-Kommando. Haengte es an `gates`,
# fuehrte jeder Gate-Lauf einen Klon mit, und ein rotes Ergebnis waere nicht
# mehr vom roten Klon-Lauf zu unterscheiden. Es wird ausdruecklich gerufen.
#
# DIESE DATEI IST KONVERGENT: jeder Lauf des Werkzeugs
# schreibt sie kanonisch neu, und ein Edit an den Belegungen unten ist danach
# still weg. Sie sind darum die VORGABE, nicht der Setz-Ort. Dasselbe gilt fuer
# das Root-Makefile: es ist der generierte Aggregator und wird ebenso neu
# geschrieben. Gesetzt wird am Aufruf
#   make selbstpruefung SELBSTPRUEFUNG_GATE='make baseline-verify'
# oder dauerhaft in repo.mk an der Wurzel, der Datei des Repos:
#   SELBSTPRUEFUNG_GATE = make baseline-verify
# Der Aggregator bindet sie nach den Fragmenten ein (`-include repo.mk`), und
# ein Lauf legt sie nur an, wo sie fehlt. Das einfache `=` dort ueberschreibt
# die Belegung des `?=` unten. Der
# Traeger unter .githooks/ liegt in derselben Klasse wie repo.mk: an einem
# belegten Ort schreibt dieses Werkzeug ihn nicht (skip-if-present),
# er gehoert dem Repo.
#
# ABHAENGIGKEIT. git, make und coreutils; was das Gate-Kommando braucht,
# bringt es selbst mit.
.PHONY: selbstpruefung

# Der Traeger im Klon — der Name gehoert git, das Verzeichnis dem Repo. Der
# Lauf haelt ihn gegen den Hook, den git nach der Aktivierung wirklich ruft.
SELBSTPRUEFUNG_TRAEGER ?= .githooks/commit-msg
# Der eine Schritt zwischen liegendem und wirksamem Traeger.
SELBSTPRUEFUNG_AKTIVIERUNG ?= make hooks-install
# Das Kommando, das im Klon gruen laufen muss.
SELBSTPRUEFUNG_GATE ?= make gates
# Die zwei Commit-Messages. Ein Repo mit EIGENEM Traeger nennt hier je eine,
# die jener aufhaelt bzw. durchlaesst.
SELBSTPRUEFUNG_MSG_ROT ?= Selbstpruefung ohne Kennung
SELBSTPRUEFUNG_MSG_GRUEN ?= Selbstpruefung mit Kennung LH-FA-01

# Fehlt das Skript, sagt dieses Ziel das und bricht ab, statt auf ein
# fehlendes Programm zu zeigen.
selbstpruefung: ## die emittierte Durchsetzungsschicht auf einem frischen Klon selbst pruefen — KEIN Gate
	@test -x tools/harness/selbstpruefung.sh || { echo "selbstpruefung: tools/harness/selbstpruefung.sh liegt nicht ausfuehrbar — das Werkzeug legt sie bei jedem Lauf kanonisch ab; ein erneuter Lauf stellt sie her."; exit 2; }
	@SELBSTPRUEFUNG_TRAEGER="$(SELBSTPRUEFUNG_TRAEGER)" \
	 SELBSTPRUEFUNG_AKTIVIERUNG="$(SELBSTPRUEFUNG_AKTIVIERUNG)" \
	 SELBSTPRUEFUNG_GATE="$(SELBSTPRUEFUNG_GATE)" \
	 SELBSTPRUEFUNG_MSG_ROT="$(SELBSTPRUEFUNG_MSG_ROT)" \
	 SELBSTPRUEFUNG_MSG_GRUEN="$(SELBSTPRUEFUNG_MSG_GRUEN)" \
	 bash tools/harness/selbstpruefung.sh
