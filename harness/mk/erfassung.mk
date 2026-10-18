# harness/mk/erfassung.mk — Aufraeum- und Berichts-Fragment der Erfassungsschicht,
# emittiert von ai-harness-init. ZWEI KOMMANDOS, KEIN GATE.
#
# Es liegt im Fragment-Verzeichnis wie die Gate-Fragmente, haengt aber NICHTS an
# GATE_CHECKS: ein Bericht prueft nichts und faerbt nichts rot, und ein Gate ueber ihm
# waere eines ueber leerem Pruefbereich. Und das Aufraeum-Ziel ruft niemand von selbst —
# ein Loeschpfad in einer Kette entfernte fremde Daten ohne Anlass.
.PHONY: span-report span-clean

# Bestand und Traeger liegen im gitignorierten Zustands-Bereich: ein frischer Klon hat
# den Traeger nicht, und span-clean nimmt den Bestand weg.
SPAN_DIR ?= .harness/state/spans
SPAN_CARRIER ?= .harness/state/bin/ai-harness-init

# Der Leser nennt seine ABDECKUNG ZUERST und meldet damit seine eigene Leere. Fehlt der
# Traeger, sagt dieses Ziel das — und sagt zugleich, dass es keine Aussage ueber den
# Bestand ist: ohne Leser wird nicht gelesen. Beide Namen werden gesucht, weil der
# Bootstrap die Endung der Plattform mitnimmt (auf Windows `.exe`).
span-report: ## Token-Bilanz je Rolle aus dem Span-Bestand — KEIN Gate (Bericht, kein Sensor)
	@for c in "$(SPAN_CARRIER)" "$(SPAN_CARRIER).exe"; do \
		if [ -x "$$c" ]; then exec "$$c" span-report "$(SPAN_DIR)"; fi; \
	done; \
	echo "span-report: der Traeger liegt nicht ($(SPAN_CARRIER)) — dieses Repo liest gerade nichts."; \
	echo "span-report: das ist KEINE Aussage ueber den Bestand, sondern ueber den Leser."; \
	echo "span-report: ein erneuter Lauf des Werkzeugs legt ihn wieder ab."

# AUFGERAEUMT WIRD AUSDRUECKLICH, NIE NEBENBEI.
# OHNE DIESEN AUFRUF WAECHST DER BESTAND UNBEGRENZT.
# Eine automatische Rotation ist nicht zugesagt, und das ist keine Luecke: ein
# Loeschpfad, der von selbst laeuft, entfernt fremde Daten ohne Anlass — der teurere
# Fehlerfall. Ob eine andere Sitzung noch in den Bestand schreibt, ist hier nicht
# entscheidbar; also raeumt niemand ausser dem Aufrufer.
#
# Entfernt wird GENAU der Span-Bestand, nichts darueber. Der Traeger daneben bleibt
# liegen: er ist kein Bestand, und ohne ihn erfasst das Repo nichts mehr.
span-clean: ## Span-Bestand entfernen (ausdruecklich, kein Automatismus) — KEIN Gate
	@rm -rf $(SPAN_DIR) && echo "span-clean: $(SPAN_DIR) entfernt"
