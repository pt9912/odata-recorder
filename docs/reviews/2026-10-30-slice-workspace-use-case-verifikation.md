# Verifikations-Report: slice-workspace-use-case — 2026-10-30

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? (gegen Plan, DoD und `SPEC-129`) · **Zweig:** `slice-workspace-use-case`, Stand `a37d372` · **Basis des Diffs:** `6cf9202` · **Bezug:** `LH-FA-09`, `LH-FA-04`, `ADR-0006`.
Es ist die erste Verifikation dieser Linie; geprüft sind der Slice nach vier Fix-Runden und die vierte Runde (G-1 bis G-5), die noch kein Re-Review sah.

## Verdikt: BESTANDEN, mit drei MEDIUM-Auflagen an den Implementer vor der Closure

- Der Code erfüllt `SPEC-129` und die DoD, soweit sie vor der Closure prüfbar ist: kein Fund an Produktcode (0 Abweichungen in 130 000 generierten Fällen mit eigenem Orakel), 0 HIGH.
- Drei Zusagen haben keinen Fänger über ihre Klasse (M-1 bis M-3). Das sind Lücken der Tests, kein Fehler des Dienstes; nach `AGENTS.md` §3.15 sind sie MEDIUM. Sie sind klein (je ein Test) und gehören vor die Closure, oder der Planner nimmt sie als akzeptierte Negative mit Grund und Folge auf.
- Die offenen Posten der Closure (unten) sind Sache des Planners und kein Befund.

## 1. Sensor-Belege (selbst gefahren, nicht aus der Behauptung übernommen)

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` im Worktree | Exit 0 (aus Datei `gates.exit`); in der Ausgabe stehen viele `CACHED`-Stufen, darum die Stufen unten einzeln ohne Cache | `…/scratchpad/verif-workspace/gates.exit`, `gates.out` |
| Stufe `test` ohne Cache (`docker build --no-cache-filter test --target test`) | Exit 0, `pruefung test: ok` (70,6 s) | `test.out` |
| Stufe `race` ohne Cache, 4 Läufe nacheinander (`--no-cache-filter race`, Befehl des Targets `test-race`) | 4 von 4 Exit 0, je `pruefung race: ok` (107 s); der Befund des Implementers (`TestTransaktionVertrag/…/ReplaceAll`, „cannot commit“) trat in keinem Lauf auf | `race1.out` bis `race4.out` |
| Stufe `lint` ohne Cache | Exit 0, `0 issues` | `lint.out` |
| `make groessenmarke` | Exit 0: `slice-workspace-use-case.md: ok, Messung 6087 von Marke 10600`; Marke unverändert; eigene Nachmessung `git diff --numstat 6cf9202..HEAD` über die drei Pfade des Blocks: 6087 (gleich bei `d4d4561` und `a37d372`) | `gm.out` |
| `make gegenproben`, Stempel | nicht gefahren (Integrationsschritt, Auftrag) | – |

**Der Befund aus dem Gate-Lauf des Implementers (SQLite-Adapter).** Der Diff `6cf9202..HEAD` berührt `src/adapters/` nicht; der Slice liefert nichts im Paket `sqlite`. Wiederholt: 4 uncached `race`-Läufe über alle Pakete 0 rot; zusätzlich `TestTransaktionVertrag` allein unter dem Detektor in 15 Prozessen (`-test.count=1`) 0 von 15 rot (`sqlite.out`). Rate unter Last nicht erreicht; ob der Slice den Befund berührt: nein (kein Diff im Paket, kein gemeinsamer Zustand mit dem neuen Test in `src/bootstrap/`, der in eigenem Temp-Verzeichnis läuft). Der Befund gehört in den Bestand des Adapters und ist ein Register-Kandidat (siehe 6).

## 2. Plan gegen Code (Diff `6cf9202..HEAD`)

Der Code liefert, was der Plan sagt, und nichts darüber:
- Neu: `src/hexagon/ports/inbound/workspace.go` (fünf Operationen), `src/hexagon/services/workspace/` (sieben Dateien Produktcode: `verwaltung`, `anlegen`, `manifest`, `reset`, `loeschen`, `protokoll`, `fehler`; rund 550 Zeilen) samt Tests, `src/bootstrap/workspace_sqlite_test.go`.
- Spezifikation: nur `SPEC-129` (Berichtigungen von *Ungeprüft*, Nr. 6 a, Nr. 14) und die Berichtigungen in `SPEC-124` und `SPEC-127` *Ungeprüft*. Der Diff berührt weder `AGENTS.md`, `.claude/`, `.harness/`, `harness/`, `tools/`, `Dockerfile`, `Makefile` noch den Stempel.
- „Ausdrücklich NICHT“ eingehalten: kein `Restore`, keine Mutation des State, kein Endpunkt, kein Coverage-Feld in der Konfig; `Create` liest den Workspace nicht vorab, `Delete` liest ihn nicht davor.
- Die Zahlen des Plans stimmen mit dem Code und dem Vertrag: 17 Stellen, 24 Paare, 20 mit Fehlerrückgabe, 18 mit `ctx` (aus der Tabelle in `SPEC-129` Nr. 15 nachgezählt); 169 Paare der Zusatztabelle (54, 65, 7, 5, 2, 36 nachgerechnet); 70 Zeilen `Z01` bis `Z70`. Alle Testnamen, die der Plan nennt, bestehen im Quelltext (Skript über alle `Test…` im Plan, 0 fehlende).
- Unsichtbare Zeichen (§3.18): `LC_ALL=C grep -P` auf NBSP, BOM, U+200B bis U+200F, U+2028 bis U+202E, U+2060 bis U+2064 in allen geänderten Dateien von `src/`, `spec/spezifikation.md`, Plan und Reports: kein Treffer.

## 3. DoD Punkt für Punkt

Alle Häkchen der DoD stehen unausgefüllt (`[ ]`); das Setzen gehört dem Planner bei der Closure. Der Stand der Belege:

| DoD-Punkt | Beleg liegt vor? |
|---|---|
| Anlegen (Baseline lesen, `ODR-E7002`/`ODR-E3001`, Manifest, Transaktion, `ODR-E7007`; `ODR-E7008`/`ODR-E5004` nur in der Abbildung) | ja: Code `anlegen.go`/`manifest.go` entspricht `SPEC-129` Nr. 4 bis 7; Probe und Mutationen unten |
| Lesen, Auflisten, Löschen mit Bereinigung (Fehler und `Zurueckgestellt` ändern das Ergebnis nicht) | ja: `verwaltung.go`, `loeschen.go`, `protokoll.go` gegen Nr. 8, 9, 13 gelesen; Mutationen V22 bis V27, V33 rot |
| Reset in einer Transaktion (State, Cursors, Revision; idempotenter Inhalt; Fehler lässt Zustand unverändert) | ja: `reset.go` gegen Nr. 10; Mutationen V30 bis V32, R2, R3 rot |
| `make gates` grün, darin `make groessenmarke` | ja (Abschnitt 1) |
| Review durchgeführt, Report liegt vor | ja: drei Reports (`-review.md`, `-review-2.md`, `-review-3.md`); die vierte Runde hat kein Re-Review, ihre Fixes löst dieser Report gegen die Klasse auf (Abschnitt 5) |
| Doku-Update öffentlicher Vertrag | ja: `SPEC-129` und Nachzüge; das Doku-Gate ist grün |
| Handbuch | `[ ]` offen bis zur Closure; §7 muss festhalten, dass das Binary unverändert blieb (Plan Entscheidung 14); kein Verhalten für Anwender, kein Katalogeintrag |
| Closure-Notiz mit Lerneintrag · Register · Risiko-Ausgänge · drei Paarungen | offen (Planner): §7 steht auf *Wird bei der Closure gefüllt*, §6 trägt für alle Risiken noch den Ausgang *offen* |

## 4. Tabelle Zusage · Fänger gegen den Vertrag; Raster

- Jede Nummer von `SPEC-129` (Nr. 1 bis 15) hat mindestens eine Zeile: Nr. 1 `Z01` bis `Z04`; Nr. 2 `Z05` bis `Z08`; Nr. 3 `Z13`, `Z14`; Nr. 4 `Z14` bis `Z16`, `Z28`; Nr. 5 `Z17` bis `Z19`; Nr. 6 `Z20` bis `Z26`, `Z68`; Nr. 7 `Z27`, `Z29` bis `Z31`, `Z66`; Nr. 8 `Z33`, `Z34`, `Z70`; Nr. 9 `Z35` bis `Z38`, `Z69`; Nr. 10 `Z39` bis `Z45`; Nr. 11 `Z46` bis `Z50`; Nr. 12 `Z09`, `Z10`, `Z32`, `Z45`; Nr. 13 `Z37`; Nr. 14 `Z46` bis `Z57`; Nr. 15 `Z06`, `Z07`, `Z09`, `Z12`. Die Negative `Z66` bis `Z70` nennen Grund und Folge.
- **Eine Zusage hat in der Tabelle keinen tragenden Fänger**: Nr. 6 „die Ursache nennt den Ort (Glied und Stelle) und keinen Wert“ — `Z20` bis `Z23` nennen nur „`ODR-E3001` mit `ErrBeschaedigt`“; siehe M-1.
- Raster (§6): 6 Operationen × 8 Spalten, keine leere Zelle; jede Zelle nennt eine Nummer von `SPEC-129` oder „gilt nicht“ mit Grund. Stichprobe gegen den Wortlaut (zwölf Zellen, darunter `Reset` × *Zustand* „Revision am Maximum gilt nicht“, `Delete` × *Herunterfahren* Nr. 2/9 c/13, `List` × *ctx* Nr. 8): die genannte Nummer trägt jeweils die Entscheidung. Keine im Code entschiedene Zelle gefunden.

## 5. Fix-Klassen gegen die Klasse aufgelöst (`AGENTS.md` §3.17)

**Probe-Paket aus Generatoren mit eigenem Orakel** (`probe_verif_test.go`, Package `workspace_test`, Saat 300002; die Mengen werden aus dem Vertragstext der Spezifikation geparst, nicht aus den Tests des Slice: Länge und Alphabet aus Nr. 3 (`1 bis 64 Bytes`, `0-9`, `a-z`, `-` = 37 Zeichen), Leerraum aus Nr. 6 a (Leerzeichen, Tabulator, Zeilenvorschub, Wagenrücklauf)). Das Orakel für das Manifest ist ein zweiter Prüfer aus dem Wortlaut von Nr. 6 (Decodierung über `any` mit `UseNumber`, Ganzzahl-Form per Regex), unabhängig vom Dienst, der `json.RawMessage` benutzt.

| Klasse | Elemente | Ergebnis |
|---|---|---|
| Kennung `id` (G: Z13/F-Klasse) | 40 000 Kennungen (6 077 gültig, 33 923 ungültig): Längen 0 bis 3, 62 bis 66, bis 99, Nachbarbytes der Bereichsgrenzen, beliebige Bytes | 0 Abweichungen |
| Leerraum vor und hinter dem Manifest (G-2/G-3/F-6) | 60 000 Paare aus Folgen der Längen 0 bis 6 (7 363 angenommen, 52 637 abgewiesen): Leerraum, Seitenvorschub, BOM, NBSP, 29 gewählte Sonderrunen (U+0085, U+1680, U+180E, U+2000 bis U+200B, U+2028/9, U+202F, U+205F, U+3000, U+FEFF, U+10FFFF u. a.), beliebige Runen 0 bis U+10FFFF, beliebige Bytes (auch ungültiges UTF-8), je Seite unabhängig | 0 Abweichungen |
| Manifest-Form (Nr. 6 b bis d) | 30 000 Manifeste (10 953 gültig, 19 047 ungültig): 0 bis 2 zufällige Mutationen aus 39 Werten (u. a. `1.0`, `1e0`, `01`, `-0`, `null`, Typ-Wechsel, leere/doppelte Properties, Dubletten byte-gleich und mit anderer Schreibweise) an Kopf- und Eintragsgliedern, Löschungen; bei Gültigen: Payload byte-gleich zum Manifest (auch mit Innen-Leerraum und Escapes), `Version` 0, `Deleted` falsch, `KnownProperties` in Manifest-Folge, `Validate()` | 0 Abweichungen |
| `entity_count` über alle Ziffern | Anzahl 0 bis 130 (gültig) und je Abweichung +1, +10, +100 (E3001) | 0 Abweichungen am Code; die Mutation V15 (Ziffer 9 abgewiesen) wird von diesem Probe-Test rot (`mutd.out`), von keinem Test des Slice |

**G-1/F-5 (Quelltext-Fänger, Mengen 660/528/32/11).** Die Zählprüfungen im Test (`n != 11*6*10`, `np != 4*6*11*2`, `len(uebrige) != 32`, `len(gut) != 11`) stehen im Code. Zehn eigene Mutationen am echten Quelltext des Dienstes an Formen, die Plan und Reports nicht nennen, alle rot in `TestPaketOhneZustand` (deterministisch): Zeiger auf die Konfig `kp := &v.k` (Z01), Zeiger auf ein Feld `up := &v.k.Uhr; *up = nil` (Z02), gruppierte `var ( a, b = 1, 2 )` (Z03), Kopie `vk := *v` mit Schreibzugriff (Z04), `*v = Verwaltung{}` (Z05, nur mit `-run` auf die Zustandstests sichtbar, weil `TestNurBenannteOperationen` zuvor mit einer Panik abbricht), Methode `kp() *Konfig { return &v.k }` (Z06), Paketvariable hinter einer Funktion (Z07) und in `fehler.go` (Z08), `go` im `defer` (Z09), Methode an `Konfig` mit Schreibzugriff (Z10). Dazu fängt `D1` (Paketvariable `zaehler`) der Test. **Beobachtung (INFO):** ein Abbruch durch Panik in einem früheren Test verdeckt `TestPaketOhneZustand` im selben Lauf (Reihenfolge der Dateien); wirksam bleibt es, weil der Lauf rot ist.

**G-2/G-3/F-6 (Leerraum-Klasse N = 6323).** Das Probe-Paket oben (60 000 Paare, unabhängige Menge, einschließlich ungültiger Bytefolgen und der Mischungen) findet nichts; die Prüfgrenzen (Länge 130, Folgenlänge 4, ein einzelnes Zeichen außerhalb, z. B. U+00E8) sind in Plan (Negative 7 bis 9), `SPEC-129` *Ungeprüft* und im Testkopf mit Grund und Folge benannt; die Folge im Betrieb ist belegt (der Import schreibt die Kanonform ohne Leerraum). Zwei Mutationen am Code im Leerraum-Bereich blieben grün und sind gleichwertig, nicht Lücken: V38 (`firstNonSpace` schneidet zusätzlich NBSP) und V39 (`json.Valid` entfernt, `json.Unmarshal` in `objekt` prüft dieselbe Syntax) — wie Äquivalenz (1) des Plans, solange eine der beiden Prüfungen bleibt.

**F-7 (Log-Fake `TestLogHandlerKetten`) und F-9 (Schreibweisen).** Am Dienst: V22 bis V25 (Bericht, Kontextfehler, `DeadlineExceeded`, Typ-Assertion statt `errors.As`) rot. Am Fake selbst keine eigene Mutation (die Reports 2 und 3 haben sie aufgelöst; hier nur Übernahme, nicht neu gemessen).

**Eigene Mutationen am Dienst (mindestens 30, an Elementen, die Plan, Implementer und Reports nicht nannten).** 67 übersetzende Mutationen (53 in der Reihe `V`, 10 in `Z`, 4 in `R`; `V50` übersetzt nicht, `D2` bis `D5` laufen nur zusammen mit `D1`, Abschnitt 6), 58 rot, 9 grün. Rot unter anderen: Kennungs-Alphabet an drei Rändern (V01 bis V04), `baselineID` mit Längenprüfung (V05), `ServiceID` aus der Snapshot-Kennung (V06), `BaselineSnapshotID` aus dem Service (V07), `CreatedAt` aus dem Snapshot (V08), `ErrDienstUnbekannt` als `Snapshot` (V09), `StateHash` gegen `MetadataHash` verglichen (V10), Kopf-Glieder `service_id`/`baseline_snapshot_id` aus dem falschen Feld (V13, V14), Schlüssel/Menge leer erlaubt (V16, V17), `entities: null` (V18), `selected_properties: null` (V19), Eigenschaften sortiert (V20), Folge umgekehrt (V21), Gegenstand an allen 14 Stellen S3 bis S16 einzeln vertauscht (V26 bis V37, V29, V30 bis V34), Dublette nur benachbart (V41) und nur über Kleinschreibung (V42), `entity_count` kleiner-gleich (V43), Präfix statt Gleichheit (V44), Version `1.0` (V45), `Log == nil` ungeprüft (V46), `ctx` nil ohne `ErrUngueltig` im Text/`Delete` ohne `ctx`-Prüfung (V51, V52).

**Gruene Mutationen (9) — Einordnung:**

| Mutation | Einordnung |
|---|---|
| V12 (`io.ReadAll` auf 1 MiB begrenzt) | akzeptiertes Negativ `Z66` (Größe: keine Grenze, Grund und Folge im Plan und in Nr. 7) |
| V38, V39 | gleichwertig (siehe oben) |
| V53 (`revision + 0*…`) | identische Mutation, bedeutungslos (Kontrollfall) |
| V11 (`Blobs.Open` mit `MetadataHash` statt `StateBlob`) | Lücke, M-3 |
| V15 (`ganzzahl` weist die Ziffer 9 ab) | Lücke, M-2 |
| V47, V48, V49 (Ort fehlt im Text; Wert im Text bei `service_id`/`baseline_snapshot_id`; Wert im Text bei leerem Schlüssel) | Lücke, M-1 |

Dieselben fünf (V11, V15, V47 bis V49) bleiben auch gegen `./src/bootstrap/` grün (`mutb.out`), die Datei fängt sie also nicht.

## 6. Nebenläufigkeit (§3.10)

Zusagen, deren Fänger der Detektor (`make test-race`) ist: keine nach Plan Z46 bis Z57 — die Fenster Z51 bis Z54 und die Rennen Z46 bis Z50 haben Wirkungs-Fänger (FIFO und Kanal im Store-Fake); die Zusage „kein Zustand“ fängt der Quelltext-Test. Eigene Messung (Messender: Verifier; Image `odata-recorder:race`, `go test -race -c`, je Mutation ein Test-Binary, **20 Prozesse** mit `-test.count=1`; Host unter `flock`, ein Container zugleich mit meinem, 4 CPUs, 4 GiB; fremde Läufe anderer Linien liefen zwischen meinen Läufen, nicht gleichzeitig mit ihnen):

| Lauf | Prozesse rot | davon `DATA RACE` | Einordnung |
|---|---|---|---|
| ohne Mutation | 0 von 20 | 0 | Normalstand |
| R1 `ReplaceAll` in zweiter Transaktion (Z46/Z51) | 20 von 20 | 0 | Wirkungs-Fänger, kein Detektor |
| R2 Revision für `Advance` außerhalb gelesen (Z48) | 20 von 20 | 0 | Wirkungs-Fänger |
| R3 `Advance` in zweiter Transaktion (Z54) | 20 von 20 | 0 | Wirkungs-Fänger |
| R4 Bereinigung unabhängig vom Ergebnis (Z55) | 20 von 20 | 0 | Wirkungs-Fänger |
| D2 Zähler (Paketvariable) in `Create` **vor** dem ersten Port-Aufruf, ohne die Quelltext-Tests | 20 von 20 | 20 | Detektor trägt es |
| D3 derselbe Zähler in `Create` nach dem ersten Port-Aufruf | 6 von 20 | 6 | wahrscheinlichkeitsabhängig (Review: 21 von 50) |
| D4 in `Reset` nach dem ersten Port-Aufruf | 0 von 20 | 0 | Detektor sieht es nicht (Review: 3 von 50) |
| D5 in `Delete` nach dem ersten Port-Aufruf | 0 von 20 | 0 | Detektor sieht es nicht (Review: 0 von 50) |

Die Zahlen D3 bis D5 stützen den Wortlaut von `SPEC-129` *Ungeprüft* (der Detektor ist für „kein Zustand“ kein Fänger) und die Entscheidung, den Quelltext-Test zum Fänger zu machen; sie widersprechen den Messungen des Reviewers nicht (kleinere Stichprobe, 20 statt 50). Die Zusage „kein Zustand“ ist mit `TestPaketOhneZustand` deterministisch gefangen (Mutationen Z01 bis Z10 und D1 rot in jedem Lauf; er parst nur, startet keine Goroutine). Die Raten des Implementers „n von n in 50 Prozessen“ für R1 bis R4 bestätigt die Stichprobe von 20 Prozessen; 50 Prozesse habe ich nicht gefahren (Zeit unter fremder Last).

## 7. Befunde

| Nr. | Stufe | Befund | Zuständig |
|---|---|---|---|
| M-1 | MEDIUM | `SPEC-129` Nr. 6 sagt: Der Verstoß ist `ErrBeschaedigt` „die den Ort nennt (Glied und Stelle) und keinen Wert“. `Z20` bis `Z23` fangen nur Code und Ursache; `TestFehlerOhneWert` prüft eine einzige Form (`completeness` mit einem Geheimwert) und gar nicht, dass ein Ort im Text steht. Mutationen an Elementen, die der Test nicht führt, bleiben in beiden Paketen grün: V47 (Ort aus `verstoss` entfernt), V48 (Wert von `service_id`/`baseline_snapshot_id` im Text), V49 (Wert von `entity_set` im Text). Klasse: jede der Verstoßstellen von `manifest.go` (rund 20 Aufrufe von `verstoss`). Ein Wert im Fehlertext ist ein Leck in Logs und Antworten. | Implementer (Fänger als Schleife über die Verstoßformen aus `TestManifestVerstoesse`: Text enthält den Ort, nie den Wert) oder Planner (akzeptiertes Negativ mit Grund/Folge) |
| M-2 | MEDIUM | `ganzzahl` (`entity_count`, Nr. 6 b „Ziffern“): die Mutation „Ziffer 9 abgewiesen“ (V15) bleibt in beiden Paketen grün, weil kein Test ein gültiges Manifest mit einer Anzahl mit der Ziffer 9 führt (Anzahlen 0, 1, 2, 3, 17, 56). Ein Dienst, der Baselines mit 9, 19, 29, … Entities ablehnt, besteht die Suite. Das Probe-Paket (Anzahl 0 bis 130 mit Abweichungen +1/+10/+100) fängt es. Klasse: die zehn Ziffern an Einer- und Zehnerstelle. | Implementer (Schleife über Anzahlen 0 bis 130 in `TestManifestGueltigeFormen` oder `TestAbbildungGenerator`) |
| M-3 | MEDIUM | `SPEC-129` Nr. 5 c „`Blobs.Open(ctx, StateBlob)` genau einmal“: In jeder Fixture ist `MetadataHash` gleich `StateBlob` gleich `StateHash` (`helfer_test.go:76`, `workspace_sqlite_test.go:126`, `:367`), darum bleibt V11 (`Open` mit `MetadataHash`) grün; „Eine Fixture gleicht nie dem Wert der Mutation“ (§3.15). Die Mutation `Open(StateHash)` ist dagegen gleichwertig (der Vergleich davor schützt). Wirkung im Betrieb: der M2-Test von `slice-bootstrap-m2` würde es fangen, kein Test des Slice. | Implementer (`MetadataHash` in der Fixture ungleich `StateBlob` setzen; der Fake für `Open` lehnt andere Referenzen ab) |
| L-1 | LOW | Plan §3 (Zeile Z20) nennt die Ziffer-Menge des Imports von `entity_count` nicht; dieselbe Klasse wie M-2 im Raster ohne Zeile. | Implementer (mit M-2) |
| I-1 | INFO | Eine Panik in einem früheren Test (z. B. `*v = Verwaltung{}` in `Get`) bricht den Lauf ab, bevor `TestPaketOhneZustand` läuft; das Gate bleibt rot, die Meldung des Zustandstests fehlt dann. Kein Fehler. | – |
| I-2 | INFO | Sendungen an die Nehmer (`slice-snapshot-use-case`, `slice-mock-schreiben`, `slice-replay-service`, `slice-bootstrap-m2`) sind in §3 und §6 mit Ausgang *offen* und Adresse „Planner beim Start“ geführt; ihre Pläne in `open/` bleiben nach Entscheidung des Users unverändert. Die §1-Ausschlüsse der Nehmer treffen die Sendungen nicht (gelesen: `slice-snapshot-use-case` führt „Snapshots erzeugen und wiederherstellen“ aus diesem Slice, `slice-replay-service` schließt „Workspace-Anlage und Reset“ nur als Lieferung dieses Slice aus). `slice-snapshot-use-case` trägt weder die Additivität (Nr. 1) noch „Restore ohne `ResetAll`“ (Nr. 11 c) in der DoD: bei dessen Start nachzutragen (Planner). `slice-admin-workspaces` besteht nicht (Folgewelle). | Planner |
| I-3 | INFO | Der Gate-Lauf `make gates` war überwiegend gecacht; die Aussage „grün“ stützt sich auf die uncached Stufen oben. | – |

## 8. Offene Posten der Closure (nur genannt)

- Closure-Notiz mit Steering-Loop-Lerneintrag (§7); Messpunkt *Ende* (nach den Fixes zu M-1 bis M-3 und dem Merge; Stand jetzt 6087 von 10600, Grenze 12500 unberührt); Beobachtungs-Register fortgeschrieben; alle Risiko-Ausgänge in §6 (alle noch *offen*); die drei Paarungen (Anker · Folge-Slice · Register).
- Diff-Vorlage an den User zu `AGENTS.md` §3.14 (Faktoren nach dem Messpunkt: Schätzung 7815, Handoff 4602 + 586, Ende nach vier Runden 6087; Verhältnis Ende ÷ Schätzung 0,78).
- Integrationsschritt: `make gegenproben` auf ruhigem Host und Stempel (`harness/gegenproben.stempel`); Prüfung der Marke gegen den Stand von `main` mit der Closure des Werkzeug-Slice; Roadmap/Welle-Plan; Hinweis, dass `SPEC-129` besteht und `SPEC-130` bis `SPEC-134` frei bleiben; der Nehmer-Eintrag in `slice-bootstrap-m2` und die Wiederholung von `make gegenproben` (Rot unter Last, Plan §6).
- Das Handbuch-Häkchen (Binary unverändert) braucht den Satz in §7.

## 9. Register-Kandidaten (nur genannt, nicht geschrieben)

- `fehlertext-nennt-ort-und-wert-ohne-faenger` (Zusage über den Inhalt einer Fehlermeldung; hier M-1) — nahe an `vertrag-sagt-mehr-zu-als-ein-gate-prueft`.
- `fixture-gleicht-dem-wert-der-mutation` (hier M-3: drei Felder derselben Referenz tragen denselben Wert) — nahe an `fix-deckt-die-genannte-mutation-nicht-die-klasse`.
- `ziffern-klasse-ohne-schleife` (M-2, die Anzahlen decken nicht alle Ziffern).
- `sqlite-transaktionsvertrag-rot-unter-last` (ein Lauf des Implementers, hier 0 von 19 Prozessen und 4 Läufen; Adapter, nicht dieser Slice).
- `panik-im-vorigen-test-verdeckt-den-quelltext-test` (I-1; schwach, 1×).

## Ungeprüft geblieben

- 50 Prozesse je nebenläufige Mutation (gefahren: 20); `-count=50` je Fänger nicht wiederholt.
- Z02, Z03, Z63, Z64 (Reflexionstests der Fakes und Hüllen) und die Fake-Mutationen aus F-7 nicht neu aufgelöst (Übernahme aus Report 3).
- `make gegenproben` und Stempel (Auftrag).

Die Mutations- und Probe-Quellen liegen unter `/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verif-workspace/` (`mut/`, `probe/`, Ausgaben `mut.out`, `mutb.out`, `mutd.out`, `probe2.out`, `probe3.out`, `racemut.out`, `sqlite.out`).
