# Verifikationsbericht: slice-test-integration — 2026-11-06

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan, DoD und Vertrag; nicht „das Richtige" (Validator) und nicht „der Diff" (Reviewer).

**Gegenstand:** `slice-test-integration`, Zweig `slice-test-integration`, HEAD `72e5b0d`, Diff `9a6fb92..72e5b0d` (Architect `fa86b70`, Review `44bfb2d`, Fix-Runde `bb13fce` und `72e5b0d`). Bezug: `LH-FA-01`, `LH-FA-02`, `LH-FA-03`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-RB-03`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md).

**Geprüft gegen:** Plan `slice-test-integration` (§2 DoD, §3 Tabelle *Zusage · Fänger* mit 30 Zeilen und Mutationsabsatz, §4 Messung, §6 Entscheidungen und Raster), Vertrag `SPEC-037` Nr. 1 bis 10, Lastenheft `LH-RB-03` (Nachweis), `LH-FA-01` bis `LH-FA-05`, `LH-FA-14`, `AGENTS.md` §3.6, §3.10, §3.13 bis §3.18.

**Methode.** Einmal gefahren, keine Wiederholungsschleife. Der Bericht des Implementers ist nicht übernommen; die Belege im Plan sind gegen eigene Läufe gehalten. **Gemessen hat der Verifier** (weder Implementer noch Reviewer). Docker nur unter `flock /tmp/odata-recorder-heavy.lock`, `docker build --memory 4g`, Container `--memory 2g --network none --read-only` mit `tmpfs` für `/tmp` und `/voll` (1 MiB), je Lauf ein Prozess (`-test.count=1`). Je Mutation eine Wegwerfkopie von `src/`, `tests/`, `go.mod`, `Dockerfile` im Scratchpad (`/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/ver-ti/`), ein Image, N Container. Der Arbeitsbaum bleibt bis auf diesen Bericht unverändert; `df` vorher und nachher 30 GB frei, `docker image prune -f` einmal (3,6 GB ungetaggt), sonst nichts gelöscht.

**Zwei Fehlanläufe, offen genannt.** Eine Mutation der Zeile 18 (Abfrage `Code()` statt des Feldes `Code` des Fehlers) wirkte nicht (0 von 20); das ist ein Fehler der Mutation, nicht des Fängers, und sie ist mit dem richtigen Feld neu gefahren (20 von 20, unten). Das Entfernen von `$..api_key` per `sed` hat beim ersten Mal nichts geändert (Diff leer); es ist mit angepasstem Muster wiederholt. Nur die wiederholten Läufe stehen in den Tabellen.

---

## Ergebnis in einem Satz

**Closure ja, mit Auflagen an Planner und Architect.** Kein HIGH, kein Befund am Produkt, `make gates` grün, `GATE_CHECKS` unverändert. Ein MEDIUM (V-1: die Zusage „als unvollständig markiert" von `LH-FA-14` Negative hat keinen Fänger und kein benanntes Negativ) und ein zweites MEDIUM (V-2: Zeile 30 hat keinen Ort im Vertrag) gehen als offene Zeilen mit Adresse in §6 des Plans, nicht in den Code.

---

## Befunde nach Schwere

**HIGH: keine.**

### V-1 (MEDIUM): `LH-FA-14` Negative sagt „als unvollständig markiert"; kein Fänger, kein akzeptiertes Negativ

- `quelle`: `spec/lastenheft.md` `LH-FA-14` (*Negative:* „… werden keine ungeschützten Bytes abgelegt, **der Austausch als unvollständig markiert** und der Fehler diagnostiziert, ohne den bereits gesendeten Status rückwirkend zu ändern"); `AGENTS.md` §3.15 (Zusage ohne Fänger), §3.10; Plan §2 DoD („`LH-FA-14` (alle drei, Negative: nicht redigierbarer Body und Speicher voll)").
- `befund`: `TestE2ESchutz` prüft zum Negativ den Status des Clients, dass Marker und Geheimnisse in keiner Datei stehen, und die Diagnose (`error_code=ODR-E2401`, `ereignis=ablage`). Die Markierung des Austauschs (`capture_status`) liest keine Stufe: der Übergang `zeige` meldet nur `status=` und `exchanges=` (`SPEC-037` Nr. 6), und kein Fall ruft den Replay des unredigierbaren Requests. `SPEC-037` Nr. 7 (Schutz) und Nr. 10 nennen die Markierung weder als Szenario noch als akzeptiertes Negativ; der Plan führt sie weder in der Tabelle noch in §6 (`grep -in 'unvollst\|incomplete\|capture_status\|markiert' tests/e2e harness/vertraege/SPEC-037-test-integration.md` trifft nur Nr. 10 b, die Wiederherstellung `incomplete` beim Neustart).
- **Probe-Mutation des Verifiers (kein Befund eines Vorgängers):** `captureStatus` in `src/hexagon/services/record/capture/abschluss.go` liefert bei fehlgeschlagenem Request-Body `CaptureComplete` statt `CaptureFailed`. Ergebnis: alle sechs Stufen mit Aufnahme (`TestE2ESchutz`, `Aufzeichnung`, `Replay`, `Weiterleitung`, `Persistenz`, `Matching`) **grün, 0 von 3 Läufen je Stufe**. Die Markierung ist über die Prozessgrenze ungefangen (im Prozess von `go test` fangen sie die Tests von `capture`; nicht Gegenstand dieses Berichts).
- `klasse`: Fänger schmaler als die Zusage des Lastenhefts; der Vertrag trägt weniger als das Kriterium, das er zu tragen vorgibt.
- Adresse: Architect (Vertrag: Szenario in `SPEC-037` Nr. 7 oder Negativ mit Grund und Folge in Nr. 10; ein Weg wäre `zeige` mit `capture_status` je Austausch oder der Replay mit `ODR-E5008`). Kein Code im Slice, bis der Vertrag es entscheidet (`AGENTS.md` §3.12).

### V-2 (MEDIUM): Zeile 30 und die Route `/odata/Konto` haben keinen Ort im Vertrag

- `quelle`: `AGENTS.md` §3.15 (der Verifier prüft die Tabelle gegen den Vertrag), §3.12, §3.16; Plan §6 letzter Punkt.
- `befund`: Die Fix-Runde (`bb13fce`) führt Zeile 30 (7 Felder der Standardmenge von `REDACT_JSON_PATHS`, 14 Werte in Request- und Antwort-Body, 4 Orte), die Route `GET /odata/Konto` im Referenzdienst und einen `POST` mit allen Feldern. `SPEC-037` Nr. 4 (Routenliste) und Nr. 7 (Schutz) nennen weder Route noch Szenario (`grep -c Konto` auf den Vertrag: 0). Nr. 7 verlangt nur das Body-Feld `password` bei vollem Speicher. Der Plan trägt es als offenen Punkt (§6 *Referenzdienst und `/odata/Konto`*, Ausgang *offen*, Adresse Architect); das ist nach §3.13 richtig adressiert, aber der Vertrag sagt zu den sieben Feldern nichts.
- `Folge`: Die Zeile hat einen Fänger (Klassenprobe unten), ihre Zusage steht aber nur im Plan, nicht im Vertrag; ein Leser des Vertrags kennt die Route nicht.
- Adresse: Architect (Nr. 4 und Nr. 7 ergänzen), wie in §6 eingetragen.

### V-3 (LOW): Größenmarke und §4 sind nach Zeile 30 nicht nachgezogen

- `quelle`: `AGENTS.md` §3.9, §3.14; `SPEC-128`.
- `befund`: `make groessenmarke` (selbst gefahren) meldet `Hinweis: Tabelle ergibt 2960, Schätzung 2880`, dann `ok, Messung 1931 von Marke 3900` und `Messpunkt … zeilen=30 nebenlaeufig=6 … schaetzung=2960 marke=3900 messung=1931`. §4 des Plans trägt noch `zeilen=29` und die Schätzung 2880 als Programmausgabe; der Block `groessenmarke` trägt `schaetzung: 2880` (Teil des Planner). Mit 24 anderen Zeilen und Zuschlag 1,35 ergäbe die Formel eine Marke von 4000; die Marke bleibt 3900, die Messung liegt weit darunter (Auslastung 0,50).
- Zahlen: die Messung stimmt (selbst gemessen `git diff --numstat 9a6fb92..72e5b0d` über die Pfade des Blocks: 1931 Einfügungen, 0 Löschungen). Die Aufteilung in §4 ergibt 1930 (`Dockerfile 8` genannt, gemessen 9); `tests/e2e` 1776 stimmt.
- Adresse: Planner (Block und §4 nachziehen; Auslastung 0,50 unter 0,60 geht nach §6 an den Architect).

### V-4 (LOW): `SPEC-037` Nr. 10 f nennt Grund und Folge nicht ausdrücklich

- `quelle`: `AGENTS.md` §3.15 (jedes akzeptierte Negativ nennt Grund und Folge).
- `befund`: Nr. 10 a, b, c, d, e, g tragen *Grund* und *Folge*; Nr. 10 f (Verdrahtung) nennt, was kein Fall fängt, und einen Lauf des Implementers, aber weder Grund noch Folge im Betrieb. Gemessen (Treiber-Tabelle unten): ohne `--read-only`, ohne `--memory 2g` und ohne `-test.count=1` bleibt `make test-integration` grün; ohne `--network none` und ohne `--tmpfs /voll` wird es rot. Die Aussage der Nr. 10 f stimmt also, und der Fänger für `/voll` ist besser als zugesagt.
- `Folge` (die der Vertrag nennen sollte): ein Lauf ohne Schreibschutz oder Speicherlimit fällt nicht auf.
- Adresse: Architect (eine Zeile).

### V-5 (LOW): Mutationstabelle — Reihenfolge

- `befund`: Die Mutationszeile 28 steht zweimal, die zweite (Ströme) hinter Zeile 30; die Tabelle ist nicht nach Nr. geordnet. Beide Zeilen sind in sich richtig (Ströme-Mutation unten wieder 30 von 30). Keine Wirkung.

### INFO

- **I-1** Die Matrix [`rueckverfolgung.md`](../plan/planning/rueckverfolgung.md) führt `LH-FA-14` Negative noch bei `slice-test-security`; Plan §6 (*Fehler-Injektion*) hält fest, dass der Planner sie nachzieht. Offen, Adresse Planner; der Diff `9a6fb92..72e5b0d` berührt die Datei nicht.
- **I-2** `make e2e-abdeckung`: Der Erzeuger vergleicht den Anker als Teilzeichenfolge: `TestE2ESchutz` zu `TestE2ESchutzX` im Treiber bleibt für ihn „unverändert" (Verifier-Lauf); der Treiber meldet das zur Laufzeit als `Test … nicht gefunden` (Zeile 2). Eigenschaft des Werkzeugs, nicht dieses Slice.
- **I-3** DoD-Häkchen in §2 des Plans sind nicht gesetzt; Closure-Notiz und Register schreibt der Planner. `Review durchgeführt` ist durch `docs/reviews/2026-11-06-slice-test-integration-review.md` belegt.

---

## Plan gegen Code (Diff `9a6fb92..72e5b0d`)

| Plan-Zusage (§3, §6) | im Diff | Urteil |
|---|---|---|
| `make test-integration` daneben, Fragment `test-integration.mk`, Root-Makefile unverändert | `harness/mk/test-integration.mk` (11 Zeilen, **keine** Zeile `GATE_CHECKS +=`); `Makefile` und alle übrigen Fragmente unberührt; `repo.mk` trägt die drei Marker von `make e2e-abdeckung` | erfüllt; `GATE_CHECKS` unverändert (`AGENTS.md` §3.6 nicht berührt) |
| Stufe `integration` im `Dockerfile` aus `source` und Binary der Stufe `build`, alle `RUN` ohne Netz | `FROM source AS integration`, `COPY --from=build`, `RUN --network=none go test -c … && go build …` | erfüllt |
| Treiber, sieben Stufen, Deklaration je Stufe | `tools/harness/test-integration.sh` (130 Zeilen), 7 Kopfzeilen, 7 `e2e_abdeckung`, 7 Anker | erfüllt; `make e2e-abdeckung` selbst gefahren: `unveraendert (7 Stufen, 7 Deklarationen, 0 ohne Kennung)`, die Orte der Datei passen zu den Zeilen des Treibers |
| `tests/e2e` mit Referenzdienst, Prozess-Helfer, Datensuche, Übergang; `tests/integration` unberührt | 14 Dateien unter `tests/e2e`; `tests/integration` im Diff nicht berührt | erfüllt |
| Ausnahme `forbidigo` nur für `tests/e2e/uebergang/main.go` | `.golangci.yml`: Muster `^tests/e2e/uebergang/main\.go$`, `# Why:`-Block | erfüllt, keine Lockerung (§3.6) |
| Kein Haken im Produkt (Nr. 8) | `src/` im Diff unberührt | erfüllt |
| Sensors-Tabelle in `harness/README.md` | Zeile `make test-integration` (kein Gate), *Nicht behauptet* ergänzt | erfüllt |
| `SPEC-037`-Stand gesetzt | Zeile *Stand* trägt den gelieferten Umfang | erfüllt |
| Zeilen 1 bis 29 der Tabelle | je Zeile ein Fänger im Code gefunden | erfüllt |
| Zeile 30 | Fänger vorhanden, Vertragsort fehlt | V-2 |

**DoD (Plan §2).** Punkt 1 (Ziel und Vertrag): erfüllt; die Einbindung daneben ist der Rat des Architect, die Entscheidung des Users bleibt offen (§6). Punkt 2 (Szenarien je Kriterium): `LH-FA-01` bis `LH-FA-05` Happy Path, Boundary, Negative je gefangen; `LH-FA-14` alle drei bis auf V-1. Die *zusätzliche verbindliche Abnahmeregel* von `LH-FA-04` (zwei parallele Replay-Sessions, Workspace-Reset) steht nicht in der DoD-Liste; die Sessions laufen hier nacheinander (`Sequenz`), der Reset gehört zu `slice-bootstrap-m3` (dessen DoD nennt ihn, Zeile 67). Punkt 3 (je Szenario eine Mutation): Tabelle vollständig, Raten unten nachgemessen. `make gates`: grün (unten). Review-Report und Doku-Update erfüllt. Closure-Notiz, Register und Risiko-Ausgänge sind Sache des Planner (offen in §6: Gate-Kette, Formelkonflikt, Auslastung, `/odata/Konto`, die akzeptierten Negative).

## Tabelle *Zusage · Fänger* gegen den Vertrag

Jede Zeile 1 bis 29 hat in `SPEC-037` eine Nummer (Nr. 1, 2, 3, 5, 7, 8, 9); Zeile 30 hat keine (V-2). Die Zusatztabelle *Modus* verweist auf die Zeilen 18 und 6: Zeile 18 gefangen (unten 20 von 20); Zeile 6, vierte Umgebung, hat der Verifier nicht mutiert. Das Raster in §6 des Plans: jede Zelle trägt Nummer oder „gilt nicht" mit Grund, keine leere Zelle gefunden. Lastenheft: alle Kriterien von `LH-FA-01` bis `LH-FA-05` und der Nachweis `LH-RB-03` (Container-Build und End-to-End-Test auf Linux) haben eine Stufe; `LH-FA-14` Negative bis auf V-1.

## Raten, vom Verifier gemessen

Je Zeile: Wegwerfkopie, Image, N Container, je Lauf ein Prozess; gemessen hat der Verifier. Unmutiert (Verifier): `TestE2ESchutz` 30 von 30 grün, `TestE2EAufzeichnung` 30 von 30, `TestE2EPersistenz` 20 von 20, `TestE2EReplay` 20 von 20; ein voller Lauf von `make test-integration`: 7 von 7 grün.

### Nebenläufige Zeilen, Element nicht im Plan genannt

| Zeile | Mutation (Datei), Element | roter Test, Ausgabe | Rate |
|---|---|---|---|
| 7 | `starten`: das erste Signal gilt als erzwungenes Ende (`erstes := true`, `cli.go`); der Plan mutiert das Abmelden von `SIGTERM` | `TestE2EProzess/SIGTERM`: `SIGTERM: Exit 143, erwartet 0` | 20 von 20 |
| 13 | `protokolliereAblage` schreibt nur den ersten Eintrag je Prozess (`protokoll.go`); der zweite Request bei vollem Speicher bleibt ohne Eintrag | `TestE2EAufzeichnung/Speicher_voll`: `1 von 2 Zeilen auf stderr mit ["ereignis=ablage" "error_code=ODR-"]` | 30 von 30 |
| 14 | Replay lässt die **erste** Aufnahme der Session aus (`replay.go`) | `TestE2EPersistenz/Neustart`: `/odata/$metadata: Wiedergabe weicht von der Aufnahme ab: Status` | 20 von 20 |
| 14 | Replay lässt die **mittlere** Aufnahme aus (`replay.go`) | `TestE2EPersistenz/Neustart`: `/odata/Products(1): Wiedergabe weicht … ab: Status` | 20 von 20 |
| 15 | leere Session: `ODR-E5002` (410) statt `ODR-E5001` (404) (`auswahl.go`) | `TestE2EPersistenz/leer`: `410 … ODR-E5002, erwartet 404 mit ODR-E5001`, auch für `$metadata` | 20 von 20 |
| 18 | Replay ruft den Upstream nur bei `ODR-E5003` auf einem `GET` (`wurzel.go`; ein Fehlerfall von neun, den die Zeile nicht nannte) | `TestE2EReplay/Fehlerfälle`: `X-Recorder-Workspace fehlend: Zähler (Verbindungen, Requests) [2 9], vor dem Replay [1 8]`; `leer` `[2 10]`; `doppelt` `[2 11]` | 20 von 20 |
| 28 | rohe Bytes des Request-Bodys auf stderr (`println` in `Aufnahme.Read`, `aufnahme.go`) | `TestE2ESchutz/Speicher_voll`: `Geheimnis GEHEIM-E-8f40 in stderr` (zugleich `JSON-Felder`, `unredigierbar`) | 30 von 30 |

Die sechs nebenläufigen Zeilen (7, 13, 14, 15, 18, 28) zählen die Wirkung (Exit-Code, Zähler am Referenzdienst, Zeilen auf stderr, Bytes unter `DATA_DIR`), und jede Mutation wird in jedem Lauf rot (`AGENTS.md` §3.10). Die Zahlen 20 von 20 und 30 von 30 des Plans sind für diese Zeilen bestätigt. Ein Datenrennen sagt keine Zeile zu.

### Klasse der Fix-Runde (§3.17): Zeile 30, `JSON-Felder`

Menge nach dem Wortlaut der Zusage: 7 Pfade der Standardmenge `REDACT_JSON_PATHS` (`SPEC-085`: `$..password`, `$..passwd`, `$..secret`, `$..client_secret`, `$..access_token`, `$..refresh_token`, `$..api_key`) × 2 Bodies (Request, Antwort) × 4 Orte (Dateien vor und nach dem Herunterfahren, stdout, stderr). Der Fänger ist eine Schleife über alle drei Mengen (`sensibleFelder()` und `jsonFelder`).

**Je ein Pfad aus der Standardmenge entfernt (`src/bootstrap/konfig_werte.go`), alle sieben einzeln:**

| Entfernt | roter Test, Ausgabe | Rate |
|---|---|---|
| `$..password` | `TestE2ESchutz/JSON-Felder`: `Feld password: GEHEIM-Q1-anfrage steht in Dateien vor dem Herunterfahren` und `nach`, ebenso `-antwort` | 1 von 1 |
| `$..passwd` | `Feld passwd: GEHEIM-Q2-…` vor und nach, beide Bodies | 1 von 1 |
| `$..secret` | `Feld secret: GEHEIM-Q3-…` | 1 von 1 |
| `$..client_secret` | `Feld client_secret: GEHEIM-Q4-…` | 1 von 1 |
| `$..access_token` | `Feld access_token: GEHEIM-Q5-…` | 1 von 1 |
| `$..refresh_token` | `Feld refresh_token: GEHEIM-Q6-…` | 1 von 1 |
| `$..api_key` | `Feld api_key: GEHEIM-Q7-…` | 1 von 1 |

7 von 7 rot, je nur das Feld `JSON-Felder`, je beide Bodies, je vor und nach dem Herunterfahren (4 Meldungen je Pfad). Klasse „andere" (deterministisch), ein Lauf je Mutation. Der Implementer hat dieselben sieben gefahren; hier ist es unabhängig gemessen.

**Mutationen am Matcher und an der Richtung, die weder Reviewer noch Implementer nannten:**

| Mutation | betroffene Elemente | roter Test | Rate |
|---|---|---|---|
| `JSONPfad.Passt` trifft keinen Namen mit `_` (`redaction.go`) | `client_secret`, `access_token`, `refresh_token`, `api_key`, je Request und Antwort (8 Elemente) | `Feld <name>: GEHEIM-Q<n>-… steht in Dateien nach dem Herunterfahren` (und vor) | 1 von 1 |
| `Passt` trifft keinen Namen über 8 Zeichen (`redaction.go`) | `client_secret`, `access_token`, `refresh_token`, je beide Bodies (6 Elemente); `password`, `passwd`, `secret`, `api_key` ohne Meldung, wie erwartet | wie oben | 1 von 1 |
| Antwort-Bodies ohne Pfadmuster (`aufnahme.go`, Richtung `Response`) | alle sieben, nur `-antwort` | `Feld <name>: GEHEIM-Q<n>-antwort steht in Dateien …`, 7 Felder | 1 von 1 |
| Request-Bodies ohne Pfadmuster (Richtung `Request`) | alle sieben, nur `-anfrage` | wie oben, 7 Felder | 1 von 1 |

Gedeckt sind damit Pfad × Body für die Orte *Dateien*; für *stderr* belegt die Mutation der Zeile 28 (rohe Request-Bytes auf stderr): `Feld <name>: GEHEIM-Q<n>-anfrage steht in stderr` für alle sieben Felder. *stdout*: die Schleife führt den Ort; eine Mutation dorthin habe ich nicht gefahren (`sucheStroeme` durchsucht stdout und stderr in einem Aufruf, `suche_test.go:48`).

### Klasse der Fix-Runde (§3.17): Ströme bei vollem Speicher, Zeile 28

Menge: 3 Geheimnisse (`GEHEIM-A`, `GEHEIM-D`, das Body-Feld `password` `GEHEIM-E`) × {stdout, stderr}. Mutation an einem Element, das der Befund nicht nannte (Body-Feld statt Header, rohe Bytes): `Geheimnis GEHEIM-E-8f40 in stderr`, 30 von 30.

## Treiber, Verdrahtung und akzeptierte Negative

| Mutation (Kopie des Treibers oder eines Tests) | Ergebnis | Zeile / Nr. |
|---|---|---|
| `--network none` durch `--network bridge` | `Stufe 1: TestE2EProzess rot (Exit 1)`, übrige sechs grün, Exit 1 | Zeile 1 |
| `--tmpfs /voll` entfernt | `Stufe 3: TestE2EAufzeichnung rot (Exit 1)`, `Stufe 7: TestE2ESchutz rot (Exit 1)`, Exit 1 | Zeilen 13, 28 (Vorbedingung `ENOSPC`) |
| `--read-only` entfernt | alle sieben grün, Exit 0 | Nr. 10 f (akzeptiert) |
| `--memory 2g` entfernt | alle sieben grün, Exit 0 | Nr. 10 f (akzeptiert) |
| `-test.count=1` entfernt | alle sieben grün, Exit 0 | wirkungslos (Vorgabe von `go test`) |
| `TestE2EReplay` umbenannt, Treiber unverändert | `Stufe 5: Test TestE2EReplay nicht gefunden`, Exit 2, Stufen 6 und 7 nicht gelaufen | Zeile 2 (andere Stufe als im Plan) |
| `t.Skip` in `TestE2EMatching` | `Stufe 6: TestE2EMatching rot (SKIP)`, Exit 1 | Zeile 3 |
| Argumentprüfung entfernt, Aufrufe mit `x` und `''` | beide Läufe: sieben Stufen grün, Exit 0; unmutiert `Aufruf:` und Exit 2 ohne Container (beide Läufe selbst gefahren) | Zeile 4 |
| Deklaration `LH-FA-14` gelöscht | `make e2e-abdeckung` Exit 2, `Stufe ohne Deklaration` | Zeile 5 (andere Deklaration als im Plan) |
| Anker einer Deklaration geändert (`TestE2ESchutz` zu `TestE2EAndere`; `TestE2EAufzeichnung`, `TestE2EPersistenz` ebenso) | `make e2e-abdeckung` Exit 2, `Deklaration ohne Stufe` | Zeile 5 |

**Akzeptierte Negative (`SPEC-037` Nr. 10 a bis g, Nr. 4):** a Grund (kein Eingang im Binary) und Folge (ein Fehler im Eingang ist ungefangen, bis er existiert; `slice-bootstrap-m6` führt den Rundlauf als Stufe, seine DoD nennt ihn, Zeile 64); b Grund (zweites Signal und Wiederherstellung im Prozess von `go test`) und Folge (Verdrahtung nur des ersten Signals ohne Fänger), die Verweise `SPEC-138` Nr. 7 und `SPEC-123` Nr. 5 lösen auf; c Grund (Fristen sind Grenzen) und Folge (keine Messung, `LH-QA-01.a`); d Grund und Folge (`Content-Encoding`, `chunked` und große Bodies prüft der Vergleichstest `LH-QA-07`; `tests/integration/vergleich_test.go` und `vergleichsdienst_test.go` auf `main` nennen `chunked` und `Content-Encoding`); e Grund und Folge (Plattform); **f ohne Grund und Folge (V-4)**; g Grund und Folge (SKIP; der Treiber wertet ihn als rot, Zeile 3 selbst gefahren); Nr. 4 (eigene Fassung des Referenzdienstes) Grund und Folge genannt. Alle stehen in Plan §6 mit Ausgang *weiter offen* (Negative sind Grenzen).

Adressen (`AGENTS.md` §3.13): `slice-bootstrap-m3`, `slice-bootstrap-m4`, `slice-bootstrap-m6` tragen die Folge-Szenarien in ihrer DoD (Zeilen 67, 65, 64); `slice-container-image` führt `SPEC-037` unter den berührten Spec-Stellen und übernimmt nichts von hier.

## Sensoren, vom Verifier gefahren

| Sensor | Ergebnis |
|---|---|
| `make test-integration` | 7 von 7 Stufen grün (warmer Cache); Stufe 7 trägt `JSON-Felder` |
| `make e2e-abdeckung` | unverändert, 7 Stufen, 7 Deklarationen, 0 ohne Kennung; `docs/plan/planning/e2e-abdeckung.md` passt zu den 7 Stufen |
| `make groessenmarke` | `ok, Messung 1931 von Marke 3900`; Hinweis `Tabelle ergibt 2960, Schätzung 2880` (V-3) |
| `make gates` im Arbeitsbaum (Zweig `slice-test-integration`, Commit `d743e62` über `72e5b0d`, kein Code geändert) | **Exit 0**, alle Gates grün, darunter `zeichen-check: ok, 1273 Datei(en)` |
| Rohbytes (`AGENTS.md` §3.18): `LC_ALL=C grep -P` nach Backslash-`u`-Folge, BOM, NBSP, U+2028 in `tests/e2e`, Treiber, `SPEC-037` und Plan | kein Treffer; das wörtliche `\n-` aus Befund L-1 des Reviewers ist behoben |
| `make docs-check` und `make zeichen-check` mit diesem Bericht | `d-check: 629 Datei(en) geprüft, 0 Befund(e)`; `zeichen: ok, 1273 Datei(en)` |

Gegengelesen auf Rohbytes wurde diese Datei (`docs/reviews/2026-11-06-slice-test-integration-verifikation.md`).

## Entscheidungs-Konformität

- **Netzlosigkeit mit zwei Servern (`SPEC-037` Nr. 1):** ein Container je Stufe mit `--network none`; das Fehlen fängt `TestE2EProzess/außen` (rot gesehen). Konform.
- **Eigene Fassung des Referenzdienstes (Nr. 4):** `tests/integration` unberührt, `tests/e2e` eigenes Paket. Konform.
- **Quelle der Session-ID, Übergang (Nr. 6):** Hilfsprogramm über die Schnittstellen des Composition Root, Adapter `sqlite` nur lesend, [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md); `make a-check` ist Teil von `make gates` (grün). Konform.
- **Fehler-Injektion ohne Haken im Produkt (Nr. 8):** `src/` im Diff unberührt, `tmpfs` von 1 MiB. Konform.
- **Spurwert und Geheimnis (Nr. 9):** der Spurwert färbt vor dem Herunterfahren nicht rot (`t.Logf`, im Lauf gesehen: `Spurwert vor dem Herunterfahren: {odata-recorder.db-wal MARKER-VOLL-5a9c}`), nach dem Herunterfahren rot; Geheimnisse vor und nach, in Dateien und Strömen. Konform.
- **Gate-Kette oder daneben ([ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)):** daneben, `GATE_CHECKS` unverändert, keine Schwelle gesenkt (§3.6); die Entscheidung des Users zur Aufnahme in die Kette bleibt offen (§6).
- **Ausnahme `forbidigo`:** eine Datei, `# Why:`, keine Lockerung. Konform.

## Empfehlung

**Closure: ja**, sofern der Planner V-1 und V-2 als offene Zeilen in §6 mit Adresse Architect einträgt (`AGENTS.md` §3.13; V-2 steht dort schon) und bei der Closure V-3 (Block und §4 nach Zeile 30) nachzieht. V-4 und V-5 sind je eine Zeile. Die Closure bindet nichts an die Gate-Kette und ändert keinen Code; ein Befund am Produkt liegt nicht vor.

**Nicht Teil dieses Laufs:** die Mutationen der Zeilen 1 bis 6 (außer den oben genannten), 8 bis 12, 16, 17, 19 bis 27 und 29 — deterministisch, im Plan belegt, vom Reviewer teils gelesen; Zeile 6, vierte Umgebung; Zeile 12 und 16 nicht nachgefahren. Die Zeile 30 ist unter einem Lauf je Mutation gesehen (deterministisch).
