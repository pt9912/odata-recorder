# harness/mk/archivierung.mk — Fragment der Wellen-Archivierung, emittiert von
# ai-harness-init. EIN KOMMANDO, KEIN GATE.
#
# Es liegt im Fragment-Verzeichnis wie die Gate-Fragmente, haengt aber NICHTS an
# GATE_CHECKS und steht in keiner Prerequisite-Kette: die Archivierung prueft nichts
# und faerbt nichts rot.
#
# DER AUFRUF BEWEGT, LOESCHT UND COMMITTET IM VERSIONIERTEN BAUM DIESES REPOS.
# Er zieht die eingesammelten Slice-Dateien und den Welle-Plan nach
# docs/plan/planning/done/<welle-id>/, daneben gekuerzte Stubs aus der vendored
# Vorlage, und packt die Volltexte in ein Archiv.
# Er loescht die Review-Reports dieser Slices und committet zweimal.
# Er laeuft nur auf ausdruecklichen Aufruf, nie nebenbei.
# WELLE=altbestand archiviert die wellenlosen Slices samt ihren Review-Reports
# unter done/altbestand/; er ist die Untergrenze, an der `[untergrenze]` fuer die
# erste Wellen-Archivierung haengt. Ein Traeger ohne Schreibpfad weist ihn ohne
# --vorschau mit `[kein-schreib-pfad]` ab. KENNUNG=<K> nennt eine Kennung dieses
# Repos, die beide Commits tragen und die sein commit-msg-Traeger annimmt; fuer
# WELLE=altbestand ist sie Pflicht, ohne sie bricht der Lauf ab, bevor er etwas
# bewegt. Eine Voreinstellung gibt es nicht.
# Liegt eine Ergebnisnotiz in done/, zieht die Commit-Abstammung die Grenze: ein
# Lauf nimmt nur wellenlose Slices, deren Add-Commit Vorfahr des Add-Commits einer
# Ergebnisnotiz ist (WELLE=<welle-id>: deren frueheste Closure diese Welle ist);
# spaeter geschlossene bleiben flach liegen und stehen in der Vorpruefung unter
# "bleibt liegen (nach der Grenze)". Ein flacher Klon sperrt mit `[flacher-klon]`,
# ein Pfad ohne Add-Commit mit `[add-commit]`.
# Ein Lauf ueber einem unsauberen Arbeitsbaum bricht ab, statt fremden Inhalt in
# den Archivierungs-Commit zu nehmen.
.PHONY: archive-welle

# Der Traeger liegt im gitignorierten Zustands-Bereich: ein frischer Klon hat ihn
# nicht. Beide Namen werden gesucht, weil der Bootstrap die Endung der Plattform
# mitnimmt (auf Windows `.exe`).
ARCHIV_CARRIER ?= .harness/state/bin/ai-harness-init

# Fehlt der Traeger, sagt dieses Ziel das — und faerbt nichts rot: ein fehlender
# Traeger ist kein Fehler des Repos, sondern der Zustand eines frischen Klons. Die
# Faehigkeit liegt, wo der Traeger liegt; ein erneuter Lauf des Werkzeugs legt ihn
# wieder ab.
archive-welle: ## Zeitdokumente einer geschlossenen Welle archivieren (WELLE=<welle-id> | WELLE=altbestand KENNUNG=<K>) — KEIN Gate
	@for c in "$(ARCHIV_CARRIER)" "$(ARCHIV_CARRIER).exe"; do \
		if [ -x "$$c" ]; then exec "$$c" archive-welle $(if $(KENNUNG),--kennung "$(KENNUNG)") "$(WELLE)"; fi; \
	done; \
	echo "archive-welle: der Traeger liegt nicht ($(ARCHIV_CARRIER)) — dieses Repo archiviert seine Wellen nicht."; \
	echo "archive-welle: ein erneuter Lauf des Werkzeugs legt ihn wieder ab."
