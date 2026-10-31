# Verifikations-Report 2: slice-groessenmarke-werkzeug — 2026-10-30

**Rolle:** Verifier, frischer Kontext (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-128`, `SPEC-102`, `LH-QA-06`, `AGENTS.md` §3.9 bis §3.18). Kein Review, keine Reparatur. Vorgänger: [Verifikation 1](2026-10-30-slice-groessenmarke-werkzeug-verifikation.md) (NICHT BESTANDEN wegen V-1).

**Gegenstand:** Stand `ac12358` (Fix-Runde `e2a254e`: Tests, Anker, Stempel; `ac12358`: Plan), Basis der Fix-Runde `641fc09`, Basis der Marke `9a0efa5`. `git diff 641fc09..HEAD` ändert unter `tools/` und `harness/` nur Tests (`formel_test.go`, `mengen_befund_test.go`, `mengen_pfad_test.go`, `plan_test.go`), die Target-Zeile in `harness/mk/groessenmarke.mk` und den Stempel; der Produktcode des Programms (`formel.go`, `plan.go`, `messung.go`, `main.go`) ist unverändert, die Proben aus Verifikation 1 über den Code gelten damit weiter.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-30

**Methode.** Keine Zahl und keine Mutation des Implementers übernommen. Scratchpad `…/scratchpad/vf5/` (flüchtig, nichts im Repo außer diesem Report): `git archive HEAD tools/groessenmarke`, je Mutation eine Wegwerfkopie, Perl-Ersetzung (bei Nullwirkung „KEIN EFFEKT“), Container `odata-recorder:deps`, `go test ./...`, `--network none --memory=4g --memory-swap=4g --cpus=4`, Quelle `:ro`, ein Container zugleich, alles unter `flock /tmp/odata-recorder-heavy.lock`. Rot zählt, wenn ein Test unter `--- FAIL` steht (nicht bei Baufehler). Kontrolle: die Kopie mit einem Zeilenumbruch mehr (`ctl`) ist grün. Last vor den Läufen 11 bis 14 auf 20 Kernen, `free -g` verfügbar 17 bis 18 GB.

---

## 1. Verdikt

**BESTANDEN.** V-1 (MEDIUM) ist gegen die Klasse gelöst; V-2 und V-3 sind gegen die Klasse deutlich geschlossen, mit je einem Rest an LOW. Es bleibt kein MEDIUM und kein HIGH offen. Die Zeitgrenze (V-6) und die Wortlaut-Fragen (V-7, §6-Punkte) liegen bei User, Architect und Planner und sind nicht im Code entschieden.

## 2. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| W-1 | LOW | V-1 gelöst: die Menge von Nr. 5 e (Position × Schlüssel × Messung) ist bis drei Einträge `ueberschreitung` geführt; die Menge ist aber nach oben unbeschränkt (jede Anzahl Einträge), und der Bericht nennt die Prüfgrenze 3 nicht als akzeptiertes Negativ mit Grund und Folge (`AGENTS.md` §3.17). Mutation A3 (nur die ersten drei Einträge geprüft, `[:min(3, len)]`) bleibt in allen Tests grün. Die Folge im Betrieb wäre ein stilles Auslassen ab dem vierten Eintrag; ein Plan trägt in der Praxis höchstens zwei bis drei. | A3 grün (Abschnitt 3) | Implementer: `k` bis 5 in `TestGrund` und M6 (d), oder die Grenze 3 als Negativ mit Grund und Folge in §3 |
| W-2 | LOW | V-3 Rest: Nr. 3 d („bytegleich“) hat die Menge „Paare, die bytegleich verschieden sind“; geführt sind drei Paare (Groß/klein in zwei Richtungen, Leerzeichen innen). Normalisierungen an Elementen, die der Befund nicht nannte, bleiben grün: Nullen am Anfang abgeschnitten (`01` gleich `1`, N2), Punkt am Ende abgeschnitten (`Z1.` gleich `Z1`, N3), Leerzeichen innen zusammengezogen (`Z  1` gleich `Z 1`, N7). Folge nur ein Eingabefehler an einem Label oder eine Nachsicht; der Code vergleicht bytegleich. | N2, N3, N7 grün | Implementer: Fänger als Schleife (jede Stelle × Ersatzzeichen), oder Negativ mit Grund |
| W-3 | INFO | Stempel und Code stehen im selben Commit `e2a254e`; die Betriebsregel „Stempel erst nach dem letzten Commit unter den Eingabepfaden“ meint den Inhalt: `make gegenproben` bestätigt „Eingaben unverändert … `ebdb45ef…`“, `ac12358` ändert keinen Eingabepfad. Kein Befund, nur der Hinweis, dass die Reihenfolge der Commits nicht die der Regel ist. | `git log -- harness/gegenproben.stempel`; Lauf | – |
| V-4 | LOW (unverändert) | Die drei offenen Randformen (1) Verzeichnis `slice-x.md`, (2) M3-Zeile „im Block“, (3) M4 (c) „alle ungültig“ stehen jetzt in §6 mit Ausgang *offen* und Adresse (Architect, Planner, Planner und Architect); (4) Marke 0 ist entschieden (Nr. 5 b, 5 i). §3.13 ist erfüllt. Die Entscheidungen liegen nicht bei mir. | §6 Zeile „Randformen, die die Fänger der Runde …“ | Architect (1), Planner (2), Planner und Architect (3) |
| V-5 | geschlossen | §4 trägt die Messpunkt-Zeile des Programms für den Stand der Fix-Runde (`messung=5342`) und die Zahlen 5342, 0,83, 1,49, 67,6; sie gleichen meinem Lauf. | Abschnitt 5 | – |
| V-6 | LOW (offen) | Zeitbudget von `SPEC-102` (120 s): mein erzwungener Lauf von `make gegenproben` brauchte 122,7 s (Schritt `RUN gegenproben` 75,2 s, Rest Bau und Export), Last 13,6 zu Beginn; 120 Fälle, 0 rot. Zusammen mit meinem Lauf in Verifikation 1 (121,0 s) und dem des Implementers (121,3 s) hält das Budget in drei von drei Messungen nicht. Die Zusatzzeit der Fälle 38 bis 120 ist klein (2,8 s); der Rest ist Lage des Laufs. §6 trägt den Punkt mit Adresse. | Abschnitt 6 | User (Trigger 30 s, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)), Architect (Wert des Budgets in `SPEC-102`) |
| V-7 | INFO (offen) | „Getrimmt“ und „Leerraum“ sind in `SPEC-128` nicht auf ASCII festgelegt (`strings.TrimSpace` kürzt auch NBSP und Unicode-Leerraum); §6 trägt den Punkt mit Adresse Architect. Nicht im Code entschieden. | §6 letzter Punkt | Architect |
| V-8 | geschlossen (Anker) / offen (Rest) | Der Anker `· seit welle-record-replay` steht auf der Target-Zeile `groessenmarke:` in `harness/mk/groessenmarke.mk` (`git diff`); der Rest der Closure-Vorbedingungen liegt beim Planner und User (Abschnitt 7). | Abschnitt 7 | Planner, User |

## 3. Die Fixes gegen die Klasse (`AGENTS.md` §3.17)

Die Mengen sind aus dem Wortlaut von `SPEC-128` gebildet, die Elemente aus dem Code der Tests nachgezählt (die Tests führen eigene Zähler, die bei Abweichung rot werden: `TestGrund` 84 = 3 · 7 · 4, M6 (d) 120, M6 (e) 72, Kopfzeilen 70).

**V-1, Nr. 5 e: Stelle × Form des leeren Grundes × Messung.** Geführt: Position (erster bis dritter von einem bis drei Einträgen `ueberschreitung`, jeder einzeln leer; `ersetzt` mit null bis drei Einträgen davor) × Form {leer, Leerraum, Tab} × Messung {0, 1, Marke, Marke + 1}, dazu beide Schlüssel zugleich leer. `ersetzt` ist nicht wiederholbar (`wiederholbar` in `plan.go`), hat also eine Position. Eigene Mutationen, keine vom Reviewer oder Implementer genannte:

| Mutation | Wirkung | Ergebnis |
|---|---|---|
| A1 nur der letzte Eintrag `ueberschreitung` geprüft | Position | rot (`TestGrund`, `TestM6GrundAnJederStelle`) |
| A2 der zweite Eintrag ausgelassen | Position | rot (beide) |
| A4 nur die ersten zwei Einträge | Position | rot (beide) |
| A3 nur die ersten drei Einträge | Position über der Prüfgrenze | **grün** (W-1) |
| A5 `ueberschreitung`-Befund nur bei Messung über 0 | Messung 0 | rot (`TestGrund`, `TestRangfolge`, `TestExitUndStroeme`, …) |
| A6 `ueberschreitung`-Befund nicht bei Messung gleich Marke | Grenze | rot (4 Tests) |
| A7 nicht bei Messung gleich Marke + 1 | Grenze | rot (4 Tests) |
| A8 `ersetzt`-Befund nicht bei Messung gleich Marke | Grenze | rot (3 Tests) |
| A9 `ersetzt`-Befund nicht bei Messung 1 | Messung | rot (beide) |
| A10 `ersetzt`-Befund nur bis zur Marke | Messung | rot (beide) |
| A11 `ersetzt`-Befund nur bei N gleich 1000 (dem Wert der Fixture) | Fixture-Wert | rot (`TestGrund`) |
| A12 `ersetzt`-Befund entfällt bei drei Einträgen davor | Zusammenwirken | rot (M6 (d)) |
| A13 `ueberschreitung`-Befund nicht bei N gleich 1300 (der mittlere Eintrag) | Position | rot (beide) |
| A14 `ueberschreitung`-Befund nur über der Marke oder bei mehr als einem Eintrag | Messung × Anzahl | rot (6 Tests) |
| A15 `ersetzt`-Befund entfällt, sobald ein Eintrag `ueberschreitung` steht | Zusammenwirken | rot (M6 (d)) |
| A16 `ueberschreitung`-Befund entfällt, sobald `ersetzt` steht | Zusammenwirken | rot (M6 (d)) |
| B1 Grund nur links getrimmt (Leerzeichen) | äquivalent: die Blockzeile ist vor dem Schnitt am Strich schon getrimmt | grün, äquivalent |
| B2 Grund nur rechts getrimmt | Grund `   x` | rot (`TestEintraege`) |

15 von 16 Mutationen der Reihe A rot, eine (A3) grün: das ist die Prüfgrenze von W-1, nicht eine Lücke des Codes. Die Mutationen X02 und X10 aus Verifikation 1 entsprechen A1/A4 und A5/A10 und sind rot.

**V-2, Nr. 4 c: Kopfzeilen der Messung.** Geführt: jede der sechs Zeilen (`basis`, `spitze`, drei `pfad`, `numstat`) × elf Abweichungen (Name groß vorn, ganz groß, Präfix, Suffix, Leerzeichen davor, um ein Zeichen gekürzt, Leerzeichen am Ende, CR am Ende, zwei Leerzeichen, Tab, ohne Trennzeichen), dazu „Zeile fehlt“ und „vertauscht“. Eigene Mutationen:

| Mutation | Ergebnis |
|---|---|
| K1 Zeile vor dem Vergleich kleingeschrieben (`basis`, `spitze`) | rot (`TestM5KopfDerMessung`, `TestMessungKopf`) |
| K2 `numstat` ohne Beachtung der Schreibweise | rot |
| K3 `pfad` ohne Trennzeichen gelesen | rot (beide) |
| K4 Wert von `pfad` getrimmt | rot |
| K5 `numstat` mit Leerzeichen am Ende | rot |
| K6 `numstat` nach `TrimSpace` | rot |
| K7 `pfad` ohne Beachtung der Schreibweise | rot |
| K8 `numstat` als Ende der Zeile (Suffix) | rot |
| K9 Leerzeichen vor dem Namen von `basis`/`spitze` gekürzt | rot |
| K10 CR am Ende des Werts von `basis`/`spitze` gekürzt | rot |
| K11 Tab oder Leerzeichen vor `pfad` gekürzt | rot |
| K12 `pfad` mit Leerzeichen oder Tab als Trennzeichen | rot |

12 von 12 rot. K1 bis K12 sind an Elementen, die weder Reviewer noch Implementer nannten (genannt waren U16, V03, Y1 bis Y3). Kein Rest.

**V-3, Nr. 3 d: bytegleich.** Mutationen: N1 (Schreibweise gefaltet auf Großbuchstaben), N5 (alle Leerzeichen entfernt), N6 (Schreibweise gefaltet auf Kleinbuchstaben): rot (`TestNrEindeutig`). N2 (Nullen am Anfang), N3 (Punkt am Ende), N7 (Leerzeichen zusammengezogen): **grün** (W-2). N4 (letztes Zeichen abgeschnitten) ist rot, aber aus einem anderen Grund: es lässt `Z1` und `Z10` kollidieren und bricht damit auch die Tabellen anderer Tests (22 `FAIL`); es zählt nicht als Fänger der Zeile.

**Bilanz.** 37 eigene Mutationen am Programm: 31 rot aus dem benannten Grund, 1 rot aus anderem Grund (N4), 1 äquivalent (B1), 4 grün (A3, N2, N3, N7), alle vier LOW (W-1, W-2). Rate: das Programm ist nicht nebenläufig, jede Mutation färbt `go test` deterministisch; ich habe je Mutation einen Prozess gefahren und keine Streuung gesehen (die unveränderte Kopie `ctl` ist grün).

## 4. Hülle, Treiber, Fälle

Unverändert seit Verifikation 1 (kein Commit der Fix-Runde unter `tools/gegenproben`, `tools/harness`); die Befunde und Raten dort (Signal-Fänger 20 von 20 für drei Mutationen) gelten ohne Änderung. Nicht neu gemessen.

## 5. Größe, §4, Selbstlauf

- `git diff --numstat --no-renames 9a0efa5..HEAD` über `tools/`, `harness/` ohne `harness/vertraege`, `Dockerfile`, `.dockerignore`, `.golangci.yml`: **5342** (5297 Einfügungen, 45 Löschungen) gleich dem Plan (5342 bei Marke 6400). Nach Teilen: `tools/groessenmarke` 3372, `tools/gegenproben` 1776, `tools/harness` 135, `harness` (Fragmente, README, Stempel) 29, die drei Dateien 30. Auslastung 5342 / 6400 = 0,83.
- `make groessenmarke` auf den eigenen Plan (Exit 0): `groessenmarke slice-groessenmarke-werkzeug.md: ok, Messung 5342 von Marke 6400` und `Messpunkt … zeilen=44 nebenlaeufig=0 fakes=2 produktcode=750 schaetzung=4670 marke=6400 messung=5342`; die Zeile steht in §4 (V-5 geschlossen).
- §4 liest 0,83 ausdrücklich als Stand vor dem Review der Runde, nicht als Endwert nach `SPEC-128` Nr. 9; das Wachstum 3589 → 5342 = 1,49 steht neben der Erwartung 1,19 bis 1,33. Das ist ehrlich. Die Marke misst am Ende (nach Review und Verifikation); dieser Report liegt unter `docs/` und zählt nie. Faktor und Formel entscheidet der Architect (E22), nicht ich.
- `make groessenmarke` ist im Lauf von `make gates` ebenfalls ausgegeben (`ok, Messung 5342 von Marke 6400`).

## 6. Gates, Zeit

| Lauf | Ergebnis |
|---|---|
| `make gates` (unter `flock`) | Exit 0, 91,8 s inklusive Wartezeit auf andere Läufe; `docs-check`, `a-check`, `lint`, `test`, `test-race` (Ebene `RUN pruefung race` im Cache, identische Eingabe), `schema-check`, `migration-check`, `groessenmarke` grün |
| `go test ./...` im Container auf der unveränderten Kopie `ctl` (eigener, nicht zwischengespeicherter Lauf, Quelle von HEAD) | grün |
| `make gegenproben` am HEAD | „Eingaben unverändert … `ebdb45ef…`“ |
| `make gegenproben` mit `GEGENPROBEN_STEMPEL` im Scratchpad (erzwungen) | **120 Fälle, 0 rot**, 122,7 s (`time`), Schritt `RUN gegenproben` 75,2 s; der Stempel im Scratchpad gleicht dem committeten (`ebdb45ef…`) |
| Arbeitsbaum nach allen Läufen | sauber (`git status --short` leer) |

V-6: das Budget von 120 s hält in meiner Messung nicht (122,7 s), wie schon in Verifikation 1 (121,0 s) und beim Implementer (121,3 s). Den Wert und den Trigger entscheide ich nicht.

## 7. DoD-Rest und Closure-Vorbedingungen (nur genannt, nicht erledigt)

- **Verifier-Teil der DoD:** `make gates` und `make gegenproben` grün (belegt, Abschnitt 6); je Zusage eine Mutation rot (Tabelle Z1 bis Z40, N1 bis N4; Verifikation 1 und Abschnitt 3); Review-Report liegt vor (`2026-10-30-slice-groessenmarke-werkzeug-review.md`).
- **Offen beim Planner:** der Diff-Vorschlag an den User für `AGENTS.md` §3.14 (Zeiger statt Formel, DoD Punkt 3, zweite Hälfte) liegt nicht im Repo (Closure); `welle-record-replay.md` §4 Auftrag 1 und die weiteren Stellen der Formel erst nach Annahme durch den User; Closure-Notiz (§7 trägt „Wird bei der Closure gefüllt“); Beobachtungs-Register fortschreiben; jedes Risiko in §6 trägt noch keinen Ausgang (alle „offen bis zur Closure“); die Paarungen (Anker · Folge-Slice · Register); die Zelle M3 „im Block“ (2) berichtigen.
- **Offen beim User:** [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) steht auf `Proposed` (Annahme); der Trigger von 30 s in [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md); der Zeiger in `AGENTS.md`.
- **Offen beim Architect:** Randform (1) Verzeichnis `slice-x.md`, (3) M4 (c) zusammen mit dem Planner, V-7 (Leerraum), Wert des Zeitbudgets von `SPEC-128`/`SPEC-102`.
- **Hard Rules:** §3.18 gegengelesen: `LC_ALL=C grep -P` auf Backslash-`u`-Folgen über `tools/`, `harness/mk`, `harness/README.md`, `SPEC-128`, den Plan: 0 Treffer; dieser Report enthält keine solche Folge und ist nach dem Schreiben gegengelesen. §3.2: keine `//nolint` in der Fix-Runde.

## 8. Register-Kandidaten für die Closure (nicht geschrieben)

1. `groessenmarke-ueberschritten-ohne-rueckfuehrung`: Ausgang `geplant` → `verkörpert`; Zielort `harness/mk/groessenmarke.mk:groessenmarke` (Anker `· seit welle-record-replay` steht jetzt auf der Target-Zeile), mit dem Vorbehalt, dass [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) `Proposed` ist.
2. `fix-deckt-die-genannte-mutation-nicht-die-klasse`: weiterer Beleg aus dieser Runde (V-1 zuerst nicht auf Position und Schlüssel geführt, V-2 nur an wenigen Kopfzeilen); die Runde danach löst die Klasse, bis auf die unbeschränkte Anzahl (W-1) und die offenen Normalisierungen (W-2).
3. Neu, 1×: `unbeschraenkte-menge-ohne-benannte-pruefgrenze` (§3.17 verlangt die Prüfgrenze mit Grund und Folge; die Mengen „Einträge“ und „Paare“ nennen sie nicht; W-1, W-2).
4. `gegenprobe-haengt-an-der-laufzeit-des-produkts`: weiterer Beleg (122,7 s gegen 120 s in meinem Lauf; V-6).
5. `universalaussage-im-vertrag-ungeprueft` und `vertrag-sagt-mehr-zu-als-ein-gate-prueft`: je ein weiterer Beleg aus V-1 und V-2 dieses Slice, behoben.
6. Spec-Lücke für den Architect ohne Register-Eintrag vor der Schwelle: „Leerraum“ in `SPEC-128` (V-7).

## 9. Zusammenfassung

Verdikt **BESTANDEN**: 0 HIGH, 0 MEDIUM, 4 LOW (W-1, W-2, V-4 unverändert und adressiert, V-6), 3 INFO (W-3, V-7, Closure-Rest). V-1 ist gegen die Klasse gelöst (15 von 16 Mutationen der Reihe A rot, eine Prüfgrenze); V-2 ist gegen die Klasse gelöst (12 von 12 rot); V-3 ist bis auf drei weitere Normalisierungen gelöst. Gates grün, Gegenproben 120 Fälle 0 rot, Größe 5342 von 6400 gleich dem Plan, `make groessenmarke` auf den eigenen Plan grün. Die Entscheidungen zu Zeitbudget, [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md), Randformen (1) bis (3) und Leerraum liegen bei User, Architect und Planner.
