# Verifikation: slice-handbuch-meldungskatalog — 2026-11-07

**Rolle:** Verifier (Modul 8/11) — Frage „Bauen wir es richtig?" gegen Plan, DoD, Vertrag und ADRs.

**Gegenstand:** `slice-handbuch-meldungskatalog`, Zweig `slice-handbuch-meldungskatalog`, Basis `567b8ba`, geprüfter Stand `b420961` (die Belege des Implementers stehen auf `f877db4`; seither zwei reine Plan-Commits `ed2cf06`, `b420961`).

**Quellen:** Plan `docs/plan/planning/done/slice-handbuch-meldungskatalog.md`, `harness/vertraege/SPEC-201-handbuch-katalog.md` (mit Nr. 2 e), `harness/vertraege/SPEC-102-gegenproben.md` (Fälle 174 bis 188), das Review `2026-11-07-slice-handbuch-meldungskatalog-review.md`, [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), `LH-QA-05`, `LH-RB-01`.

## Urteil

**Bestanden mit zwei MEDIUM-Auflagen (Zusage ohne Fänger, V-1 und V-2), kein HIGH, keine DoD-Verletzung der Liefer-Punkte.** `make gates` ist grün, die Marke hält, der Vertrag und der Code stimmen in jeder geprüften Stelle überein. Zwei Mutationen an Elementen der Klassen, die der Fix nicht nannte, blieben grün (V-1, V-2); beide Lücken liegen in Zusagen des Vertrags ohne Fänger und sind klein. Sie blockieren die Closure nicht, gehören aber vor `done/` als Fänger nachgeliefert oder als akzeptiertes Negativ mit Grund und Folge in den Plan.

## 1. Belege gegen Behauptungen — selbst gefahren

| Behauptung des Implementers | Eigener Lauf | Ergebnis |
|---|---|---|
| `make gates` grün | `make gates` einmal selbst, unter `flock /tmp/odata-recorder-heavy.lock`, Log im Scratchpad (`ver-mk/gates.log`), Exit 0, 2 min 19 s Wanduhr (Images warm, Last des Hosts fremd) | grün; `katalog-check` steht in der Kette (Stufe `katalog` gebaut, Schritte aus dem Cache); `a-check` 0 Befunde, `d-check` 715 Dateien 0 Befunde, `zeichen` ok, `handbuch` grün |
| Messung 2658 von Marke 3500 | `groessenmarke` in `make gates` | `ok, Messung 2658 von Marke 3500`; dazu der Hinweis `Tabelle ergibt 3210, Schätzung 2570` (siehe V-4) |
| Stand nach den Plan-Commits grün | `make docs-check`, `make zeichen-check` auf `b420961` | `0 Befund(e)`, `zeichen: ok` |
| Gegenproben G1 bis G15 (Fälle 174 bis 188) laufen, 188 Fälle 0 rot | `gegenproben-eingabe.sh` mit den Pins der Kette gegen `harness/gegenproben.stempel` | Prüfsumme `e8549540…` gleich dem Stempel. Den Lauf selbst habe ich **nicht** wiederholt (Stempel belegt nach `SPEC-102` Grenze nur, dass ein Lauf zu dieser Prüfsumme grün war; der Lauf kostet 2 bis 3 min unter der Sperre). Stand der Belege: Behauptung des Implementers, vom Stempel gestützt, nicht selbst gesehen |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -nP` auf Backslash-u-Folge, BOM, NBSP und Steuerzeichen über alle 32 geänderten Textdateien des Diffs; Zeichen jenseits von ASCII in `tools/katalog/*_test.go` | kein Treffer; Nicht-ASCII nur Umlaute, `§` und das Malzeichen in Kommentaren. Die Fixture von `TestKatGetrimmt` entsteht aus Bytewerten (gelesen in `nachtrag_test.go`) |

## 2. Plan-vs-Code-Diff

Der Diff `567b8ba..b420961` deckt die Tabelle von §3 des Plans:

| Zeile des Plans §3 | Im Diff | Befund |
|---|---|---|
| Vertrag, Ausnahmen-Datei | `SPEC-201` (Nr. 2 d, 2 e, 3 a bis 3 d, 6 c, 6 d, 7, 8, 10 a nach dem Review), `harness/katalog-ausnahmen.txt` (18 Zeilen; `ODR-E6007` fehlt bewusst) | gleich |
| `tools/katalog/` | `main.go`, `pruefen.go`, `bericht.go`, `quellen.go`, `markdown.go` (zusammen 639 Zeilen Produktcode) und acht Testdateien | gleich; Code in der Fix-Runde unverändert, wie der Plan sagt |
| `Dockerfile`, `.dockerignore` | Stufen `katalog-bin` und `katalog` (`--network=none`, `gofmt`, `go vet`, `go test`, `go build -trimpath`), `COPY` in `lint` und `gegenproben` aus `katalog-bin`; Allowlist um `tools/katalog`, `spec/spezifikation.md`, `harness/katalog-ausnahmen.txt` | gleich; die Stufe `katalog` stellt genau die vier Dateien von `SPEC-201` Nr. 10 a bereit |
| `harness/mk/katalog.mk` | Ziel `katalog-check`, `GATE_CHECKS += katalog-check`, `docker build --target katalog` ohne Tag | gleich |
| `tools/gegenproben/`, `gegenproben-eingabe.sh`, `.golangci.yml` | Fälle 174 bis 188 (`katalog.go`, `katalog_test.go`), `tools/katalog` in den Eingaben, Ausnahme von `os.Std(out|err)` für `tools/katalog/main.go` mit `# Why:` | gleich; `make gegenproben` bleibt außerhalb von `GATE_CHECKS` ([ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)), die Regel in `.golangci.yml` ist ergänzt, nicht gelockert ([ADR-0003](../plan/adr/0003-lint-profil-golangci.md), `AGENTS.md` §3.6) |
| `SPEC-102` | Fälle 174 bis 188, Programmliste, Randformen 2 und 13, Stand, Pflicht zum Fahren, Messung | gleich |
| Handbuch | Abschnitt `Fehlercodes`, Tabelle mit 18 Zeilen = H (36 Codes in `fehler.go` minus 18 Ausnahmen) | gleich; `katalog: ok` und `handbuch-check` grün |
| `harness/README.md` | Sensors-Zeile mit Bindung [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), Gegenproben-Zeile, „Nicht behauptet" um `katalog-check` ergänzt (Verdrahtung, Bedeutung, Erreichbarkeit, Literale außerhalb, Schlüssel und Kommandos) | gleich; Zellen unter 200 Zeichen (`d-check` grün) |

Nicht im Diff und nicht verlangt: Schlüssel und Kommandos (Folge-Slice `slice-handbuch-katalog-schluessel-kommandos`, `SPEC-201` Stand nennt es), `AGENTS.md` (Diff-Vorschlag an den User bei der Closure), Closure-Notiz §7 (Planner, „wird bei der Closure gefüllt").

Programm gegen Vertrag: `pruefen.go`, `bericht.go`, `quellen.go`, `markdown.go` gelesen gegen `SPEC-201` Nr. 2, 3, 6, 7, 8 — Aufruf, Ausgabe, Reihenfolge (Art, Name, Zeile als Zahl, Text), die 25 Zeichen von *getrimmt* (`unicode.IsSpace` ist genau die Menge), die Entdopplung der Form-Befunde je Quelle, Art 8 und 9 an der Zeile der ersten Nennung, Zaun- und Abschnittsregeln, die Reihenfolge der Quellen bei den Eingabefehlern. Keine Abweichung gefunden.

## 3. Tabelle Zusage · Fänger gegen den Vertrag

Alle 32 Zeilen des Plans haben Fänger und Mutation; jede Allaussage von `SPEC-201` (Nr. 1 bis 3, 6 bis 8) findet sich in einer Zeile. Gleichwertig und akzeptiert (Plan §3, Absatz *Gleichwertig*): (1) Kürzung bei genau 200 Zeichen, (2) Entfernen des `\r` in `zerlege`, (3) `[0-9]` als `\d` nicht bildbar. Grund und Folge stehen jeweils im Plan; (2) benennt ausdrücklich „Zusage ohne Fänger, solange kein Ausgabewert `\r` trägt". Ich habe (2) nachgerechnet: Überschrift, Zaun-Schluss, Zellen und Ausnahmen trimmen rechts, das Scanner-Verhalten für Raw-Strings entfernt `\r` selbst; das Negativ ist belegt.

Die Verdrahtung (Zeile 22, `SPEC-201` Nr. 9 f) ist als Negativ geführt und mit Beleg versehen (Handbuch ohne Zeile rot, Ausnahmen ohne Zeile rot). Ich habe sie nicht erneut rot gefahren; das Gate steht in der Kette und die Stufe lief in meinem `make gates`.

## 4. Fixes gegen die Klasse (`AGENTS.md` §3.17) — eigene Stichproben-Mutationen

Fahrer: ein Wegwerf-Treiber im Scratchpad (`ver-mk/`), ein Container aus `odata-recorder:race`, ohne Netz und ohne Mount, `--memory 2g --cpus 4`, unter `flock /tmp/odata-recorder-heavy.lock`; je Mutation eine frische Kopie von `tools/katalog`, genau eine Ersetzung, `go test -count=1 ./...`. Ein Prozess je Mutation (Klasse `andere`, keine Raten). Gewählt habe ich 14 Mutationen an Elementen, die weder der Review noch die Listen des Plans (94 + 42 Mutationen) nennen.

| Id | Klasse der Zusage | Mutation (Element außerhalb der genannten) | Ergebnis |
|---|---|---|---|
| A | Orte einer Nennung (`SPEC-201` Nr. 3 d: „jedes Vorkommen" von `ODR-E`) | Vorkommen mit einem Buchstaben unmittelbar davor überspringen | **grün** — V-1 |
| B | Ende von §4 (Nr. 3 b) | Ende schon bei `##` statt `## ` (also an `###`-Überschrift) | rot (`TestKatKAbschnittEnde`) |
| C | Tabelle = Folge von Zeilen mit `|` (Nr. 6 c) | Leerzeile trennt die Tabelle nicht | rot (`TestKatZaunFormen`) |
| D | Zellen trennen (Nr. 6 c) | Schlussstrich nach `\|` nicht gegen Maskierung geschützt | rot (`TestKatZellen`) |
| E | Trennzeile (Nr. 6 c) | leere Zelle der Trennzeile zugelassen | rot (`TestKatZellen`) |
| F | Kopf bytegleich (Nr. 6 c) | letzte Spalte des Kopfs nicht verglichen | rot (`TestKatHandbuchEingabe`) |
| H | Zeile der Ausnahmen (Nr. 3 c, 7) | Zeilenzahl um eins zu klein | rot (`TestKatAusgabe`) |
| I | Ausnahme: Seiten getrimmt (Nr. 3 c) | Grund nicht getrimmt | grün, **gleichwertig**: die Zeile ist vorher als Ganzes rechts getrimmt, ein Grund nur aus Leerraum kann nicht entstehen |
| J | Form: Wert in Go-Schreibweise (Nr. 7 Art 1) | `QuoteToASCII` statt `Quote` | rot (`TestKatAusnahmenPruefen`) |
| K | Überschrift von §4 (Nr. 3 b) | jede Ebene ab 2 statt genau 2 | rot (`TestKatUeberschriftFormen`) |
| Q | Überschrift, Text rechts getrimmt (Nr. 6 b) | beidseitig getrimmt (zwei Leerzeichen nach `#` gelten) | rot (`TestKatUeberschriftFormen`) |
| W | Abschnitt „mit genau diesem Text" (Nr. 6 b) | Vergleich ohne Beachtung der Schreibung | rot (`TestKatUeberschriftFormen`) |
| X | Zeile der Datenzeile (Nr. 6 c) | Zeilenzahl um eins zu groß | rot (`TestKatAusgabe`) |
| Z | Anfangsstrich steht (Nr. 6 c) | ein zweiter Anfangsstrich wird abgeschnitten (`||x|` verliert die leere erste Zelle) | **grün** — V-2 |

Ergebnis: 14 gefahren, 11 rot, 1 gleichwertig (I), 2 grün (A, Z). Die Fix-Runde (F-1 bis F-7) trägt also für die dort genannten Klassen — Ebenen, Leerraum, Zählung der Form, Zaun, Trennzeile —; die zwei Lücken liegen in Klassen, die der Vertrag zusagt und die der Plan nicht als Menge führt.

## 5. Findings

| ID | Schwere | Befund | Ort | Folge / Auflage |
|---|---|---|---|---|
| V-1 | MEDIUM | Zusage ohne Fänger (`AGENTS.md` §3.15): `SPEC-201` Nr. 3 d sagt „jedes Vorkommen von `ODR-E` mit der folgenden Folge Ziffern" in Prosa, Tabelle, Überschrift und Zaun. Die Menge „Orte im Handbuch" des Plans (12 Elemente) variiert den Ort, nicht das Zeichen **vor** dem Vorkommen. Mutation A (Vorkommen hinter einem Buchstaben überspringen) bleibt grün; ein Code in `xODR-E9999` oder `fooODR-E1234` fiele aus W und aus Art 8 und 9. | `tools/katalog/quellen.go` (`nenne`) | Fänger als Schleife über die Zeichen davor (Buchstabe, Ziffer, Unterstrich, Bindestrich, Anführungszeichen, Klammer, Zeilenanfang) × Klassen (nicht in C, in den Ausnahmen); oder akzeptiertes Negativ mit Grund und Folge im Plan |
| V-2 | MEDIUM | Zusage ohne Fänger: `SPEC-201` Nr. 6 c („der Anfangsstrich steht"; jede Zelle getrimmt) — die leere erste Zelle ohne Leerzeichen (`||x|`) hat keinen Fänger. Mutation Z (zweiten Anfangsstrich abschneiden) bleibt grün; die Zeile gäbe eine Zelle statt zwei (Eingabefehler statt Befund *Zelle*). | `tools/katalog/markdown.go` (`zellen`) | Fänger: Zeilen mit leerer Zelle an jeder Stelle (`||x|`, `|x||`, `| |x|`) für beide Tabellen; oder Negativ mit Folge |
| V-3 | LOW | Laufzeit: Die Messungen des Slice (`make gegenproben` bei geänderter Eingabe 137,3 s, 123,4 s, 132,5 s; Schritt `RUN gegenproben` 80,1 s; Last am Ende über dem Dreifachen der Kernzahl) stehen in `SPEC-102` (Zeitbudget-Absatz zu den Fällen 174 bis 188), alle drei über dem Budget von 120 s. Plan §6 (Risiko *Laufzeit*) nennt noch die alte Zahl 56,9 s und für [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) den Trigger „über 30 s bei geänderter Eingabe"; die neuen Zahlen und die Überschreitung des Budgets fehlen dort, die DoD verlangt „Zahlen in §7". Eine Messung vor der Änderung fehlt (SPEC-102: „Zusatzzeit liegt unter der Streuung"). `SPEC-102` lässt das Budget „höchstens 120 s" stehen, obwohl es in dieser Messung 3 von 3 Mal gerissen ist; „kein Gate prüft es" ist benannt. Die ADR-Frage ist dem User vorgelegt (Ausgang *offen*, Adresse User), nicht entschieden — wie verlangt. | Plan §6, §7; `SPEC-102` Zeitbudget | Planner: bei der Closure die Zahlen in §7 und im Risiko *Laufzeit* nachziehen; der User entscheidet über die ADR. Kein Fehler des Slice |
| V-4 | LOW | Plan §4 (Block `groessenmarke`) führt `schaetzung: 2570` aus 24 Zeilen; die Tabelle hat 32 Zeilen, `make groessenmarke` meldet `Tabelle ergibt 3210, Schätzung 2570` als Hinweis. Marke 3500 und Messung 2658 sind grün; die Schnittfrage ändert sich nicht (beide Werte über der Schnittgrenze, `schnitt-grund` steht). Der Text von §4 („24 Zeilen") ist veraltet gegenüber §3 (`AGENTS.md` §3.9). | Plan §4 | Planner: Schätzung in eigenem Commit nachziehen oder den Hinweis in §4 begründen |
| V-5 | INFO | `ODR-E6007` fehlt in `harness/katalog-ausnahmen.txt` bewusst. Bestätigt: der Plan führt die Sendung dreifach — §1 Nr. 12, §6 (Zeile *Zählstelle*, Ausgang *weiter offen*, Adresse Integrationsschritt) und seit `ed2cf06` in der DoD. Der Nehmer auf `main`, `done/slice-odata-json-eingabe.md`, nennt `katalog-ausnahmen`; `slice-v2-payload-schreiben` auf seinem Zweig ebenso. Der Zweig `slice-odata-json-eingabe` selbst trägt in `docs/plan` keinen Verweis (der Slice liegt auf `main` in `done/`, dort steht er). Bis der Integrationsschritt die Zeile einträgt, ist `make katalog-check` auf dem gemergten Stand mit dem Code `ODR-E6007` in `fehler.go` rot. | Plan §1, §6, DoD | keine Auflage; Integrationsschritt des Planners im ersten Verifier-Pass nach dem Merge |

## 6. DoD

| Punkt | Stand |
|---|---|
| Programm, Stufen, Ziel in `GATE_CHECKS`, Katalog der Fehlercodes, Handbuch-Abschnitt, jede Zusage mit Fänger, vom Implementer rot gesehen | erfüllt, mit V-1 und V-2 (zwei Zusagen ohne Fänger) |
| Regel über beide Richtungen und alle Elemente durch Fänger | erfüllt für die Mengen des Plans; V-1 und V-2 sind Elemente, die die Mengen nicht führen |
| Gegenproben G1 bis G15 laufen in `make gegenproben` | erfüllt laut Stempel (Prüfsumme gleich), Lauf nicht selbst wiederholt |
| Zusatzzeit gemessen, ADR-Frage dem User vorgelegt | gemessen (in `SPEC-102`), Zahlen noch nicht in §7 (V-3); ADR-Frage offen beim User |
| `make gates` grün | erfüllt, selbst gefahren |
| Review durchgeführt | erfüllt (`2026-11-07-slice-handbuch-meldungskatalog-review.md`; kein HIGH, F-1 bis F-7 in der Fix-Runde behandelt) |
| Doku-Update (`harness/README.md`, Handbuch; `AGENTS.md` als Diff-Vorschlag) | README und Handbuch erfüllt; Diff-Vorschlag steht bei der Closure aus |
| Closure-Notiz, Register, Ausgang je Risiko, drei Paarungen | Sache der Closure; §6 trägt Ausgänge (zwei *weiter offen*, Adressen genannt) |

## 7. Akzeptierte Negative — Grund und Folge

Geprüft gegen `AGENTS.md` §3.15: `SPEC-201` Nr. 9 a bis g und die drei Einträge des Plans nennen je Grund und Folge im Betrieb. Eine Ergänzung: Nr. 9 c („Erreichbarkeit und die Ausnahmen") hat ausdrücklich „ein Fänger fehlt"; die Gegenmaßnahme (Pflicht der Folgewellen, Nr. 11) ist als Vorschlag an den User in der Closure vorgemerkt. Das reicht.

## 8. Grenzen dieser Verifikation

- Die Gegenproben habe ich nicht erneut gefahren (Stempel gleich der Prüfsumme der Eingaben).
- Die 14 Mutationen sind eine Stichprobe über die Klassen, keine Vollständigkeit; die zwei grünen Fälle zeigen, dass die Klassen „Zeichen vor einer Nennung" und „leere Zelle ohne Leerzeichen" fehlen, nicht dass keine weitere Lücke besteht.
- Der Beleg für die Zeit von `make gates` (2 min 19 s) ist eine Einzelmessung bei fremder Last und kein Vergleich vorher/nachher.
- Eine Verifikation ersetzt nicht den Probelauf der Abnahme (Bedeutung der Zeilen, `SPEC-201` Nr. 9 a).
