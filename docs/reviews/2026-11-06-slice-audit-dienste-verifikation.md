# Verifikationsbericht: slice-audit-dienste — 2026-11-06

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig — gegen Plan, DoD und `SPEC-234`? Nicht die Frage des Validators, nicht die des Reviewers.

**Gegenstand:** `slice-audit-dienste`, Zweig `slice-audit-dienste`, HEAD `303e09b`, Basis `1543b02`. Bezug: `LH-RB-04`, `LH-FA-13`, `LH-QA-04`, [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md).

**Eingang:** Plan `docs/plan/planning/in-progress/slice-audit-dienste.md` (DoD, §1, §3, §4, §6), `SPEC-234` Nr. 1 bis 11 und *Ungeprüft*, Review `2026-11-06-slice-audit-dienste-review.md`, `AGENTS.md` §3.10, §3.15 bis §3.18. Der Bericht des Implementers ist nicht übernommen; die Sensoren sind selbst gefahren.

**Modell:** Sonnet 5.5 (Verifier-Agent, Messender aller Zahlen unten) · **Methode:** einmal gefahren, keine Wiederholungsschleife.

## Ergebnis in einem Satz

Der Code erfüllt `SPEC-234` Nr. 2 bis 10 an allen vier Aufrufstellen, `make gates` ist grün, die Fix-Runde deckt die Klassen von M-1 bis M-4 (Probe-Mutationen an nicht genannten Elementen rot), aber neun eigene Probe-Mutationen an Zweigen und Formen, die kein Fänger führt, blieben grün (V-1, V-2, V-4 MEDIUM; V-3, V-5, V-6 LOW; V-7 für den Planner); kein HIGH.

## Belege der Sensoren (selbst gefahren)

| Sensor | Ergebnis |
|---|---|
| `make gates` auf `303e09b` (unter `flock /tmp/odata-recorder-heavy.lock`, ohne Mutationslauf zugleich) | Exit 0; darin `baseline-verify` OK (54 Dateien), `d-check` 681 Dateien 0 Befunde, `pruefung test`, `pruefung race`, `migration-check` ok, `pruefung handbuch` (Stufe `handbuch`), Lint, `zeichen` ok (1380 Dateien) |
| Größenmarke | `git diff --numstat 1543b02..303e09b` über die fünf `pfad`-Zeilen: 2034 Einfügungen + 24 Löschungen = **2058**, Marke 3000, nachgerechnet; `src/bootstrap/` (111 + 6 = 117 Zeilen) liegt außerhalb der Pfade und ist nicht gezählt |
| Gate-Kette | `git diff --name-only 1543b02..303e09b` nennt weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, `.a-check.yml` noch `.golangci.yml` |
| `//nolint` | keine neue Direktive im Diff (`git diff` auf `+.*nolint`: nur Zitate im Review-Bericht) |
| Rohbytes (`AGENTS.md` §3.18) | gegengelesen mit `LC_ALL=C grep -P`: `src/hexagon/ports/support/supporttest/audit_test.go` (22 Zeilen mit Bytes ab `0x80`: Umlaute in Kommentaren, keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP), außerdem `src/hexagon/services/*/audit*_test.go` und `src/bootstrap/audit_bruecke_test.go` auf BOM und NBSP (kein Treffer). Dieser Bericht trägt keine solche Folge |
| Aufräumen | nach allen Läufen kein Container mit dem Label `odata-recorder.mutationslauf=1`; die Wegwerfkopien liegen im Container; die Listen liegen im Scratchpad der Sitzung, nicht im Repo; der Baum des Zweigs blieb bis auf diesen Bericht unberührt |

## Konformität je Nummer von SPEC-234 gegen den Code

| Nr. | Zusage | Befund am Code |
|---|---|---|
| 1 | Port `Append`, kein `ctx`, sicher für gleichzeitigen Aufruf | Port und Adapter gehören `slice-audit-log`; der Fake ist sicher (zL1 bis zL4 rot unter dem Detektor) |
| 2 | Arten und Felder je Art | `loeschen.go` (Session): `ServiceID` aus `Sessions.Get`, `SessionID`; Aufbewahrung: `dienstID`, `sessionID`; Workspace: nur `WorkspaceID`; Import: `ServiceID`, `ImportID`, `ErrorCode`, keine Quell-Session. Verbotene Felder bleiben leer |
| 3 | Wertformen, kein Geheimnis | kein Feld nimmt Text eines Fehlers, einer URL oder eines Headers; Probe vI4 rot, vI5 rot-paket (INFO unten) |
| 4 a bis d | genau eine Stelle je Art, erst nach der Tatsache | gelesen: Session `Delete` nach `Sessions.Delete` und vor `CleanupBlobs`; Aufbewahrung `loesche` im Zweig `err == nil`; Workspace nach `Workspaces.Delete`; Import nach `neuerImport` bei Fehler ohne Kontextfehler. *Kein Eintrag*-Liste: bis auf die unten genannten Lücken gefangen |
| 5 a | Eintrag folgt der Löschung | vA4, vA5, vP7, vW1 (Eintrag in Ablehnungszweigen) rot |
| 5 b | Ergebnis hängt nicht am Eintrag | `_ =` an allen Stellen; vA3, vW3, vB3 rot; Import: vI1 grün (V-3) |
| 5 c | beendeter Kontext ändert den Eintrag nicht | vA9, vW7 (Eintrag nur ohne `Canceled`) rot |
| 5 d | synchron, ohne Frist, ohne Goroutine | kein `go` und kein `time` in den Aufrufstellen; Fänger E19 durch Reviewer und Implementer gefahren, hier nicht wiederholt |
| 6 | Blobs ohne Eintrag | kein Eintrag in `CleanupBlobs` oder `protokolliere`; vA6/vW4 (Eintrag per `defer` nach der Bereinigung) rot |
| 7 | `KorrelationsID` leer | in keinem Eintrag gesetzt |
| 8 | Feld `Audit` in vier Konfigs, `New` weist nil ab und ruft `Append` nie | vA8, vB5, vW6, vI6 (`New` ruft `Append`) rot; E17 durch Implementer (Prüfung entfernt) nicht wiederholt |
| 9 | Fake: sieben Felder, Zähler, Kopie, `Vor` | vF1 bis vF4, vF5b, vF6, vF7 rot |
| 10 | Wiederholung und Nebenläufigkeit | die Dienste halten keinen Zustand; Zähler-Proben unten |
| 11 | Katalog unverändert | kein Schlüssel, kein Code, kein Kommando im Diff (`git diff` auf `spec/` nur `SPEC-234` und Nachzug in `SPEC-123`, `SPEC-124`, `SPEC-127`, `SPEC-129`) |

## DoD und Plan gegen Code

| Punkt der DoD | Stand |
|---|---|
| Konfig und Fake | erfüllt am Code (Nr. 8, 9), Tests E01 bis E04, E17 |
| Vier Aufrufstellen | erfüllt am Code; drei Lücken im Fänger (V-1, V-2, V-4) |
| Kein Geheimnis | erfüllt am Code; Fänger doppelt gedeckt (INFO I-1) |
| `make gates` grün | **bestätigt** (eigener Lauf, Exit 0) |
| Review durchgeführt | Report liegt vor (`2026-11-06-slice-audit-dienste-review.md`), Fix-Runde `13d9885`, `303e09b` |
| Doku-Update | `spec/spezifikation.md` trägt `SPEC-234` und den Nachzug in vier Verträgen |
| Handbuch | `make handbuch-check` grün; das Handbuch nennt weder Aufbewahrung, `DATA_DIR` noch Session (grep), keine Aussage wird falsch; §7 muss bei der Closure die Datei `audit.jsonl` nennen, solange die Brücke bleibt |
| Closure-Notiz, Register, Ausgänge in §6, Paarungen | offen: §7 steht auf *Wird bei der Closure gefüllt*, mehrere Risiken tragen *offen* / *weiter offen* (Closure-Pflicht, keine Abweichung) |

Plan gegen Code: §3 nennt 19 Zeilen und 7 Mengen; der Code führt jede. M-5 des Reviews (Plan widerspricht sich zur Brücke) ist behoben: §1 Ziel, §2 Handbuch-Punkt, §3 Tabelle und §4 nennen die Brücke. Zwei Reste: §6 *Randformen* trägt noch den Ausgang *offen (bis zum Review)*, das Review liegt vor; §4 nennt `slice-bootstrap-m2` als laufend, der Plan liegt in `done/`.

## Eigene Probe-Mutationen (Fix-Runde gegen die Klasse, `AGENTS.md` §3.17)

Werkzeug: `tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, `--memory 2g` (Vorgabe des Treibers), `--cpus=4`, je Liste unter `flock /tmp/odata-recorder-heavy.lock`, ein schwerer Lauf zugleich (ein anderer Agent hielt den `flock` zwischen meinen Läufen). Messender: der Verifier-Agent (Sonnet 5.5). Last = erster Wert von `/proc/loadavg` des Hosts je Mutation: Listen A, A2, B, C, D zwischen 3,4 und 7,1 bei 4 CPUs des Containers (steht in jeder Ergebniszeile, nicht beschönigt). Ein Rot unter Last ist kein Rot auf ruhigem Host, ein Grün wird durch Last nicht Rot. Insgesamt 57 Mutationen (34 + 5 + 14 + 4), eine davon (vF5) übersetzte nicht und ist durch vF5b ersetzt. Alle Mutationen betreffen Zweige oder Elemente, die weder der Plan noch das Review nannten.

### Die elf Zeilen, die das Review als nicht aufgelöst führte

| Zeile | Eigene Mutation (rot = Test der Zeile oder Paket rot) | Ergebnis |
|---|---|---|
| E01 | Feld `ServiceID`, `SessionID`, `WorkspaceID` im Fake nicht festgehalten (vF1 bis vF3; der Plan nannte `ImportID`, `KorrelationsID`, `ErrorCode`, `Art`) | 3 von 3 rot |
| E02 | `lesungen` um 2 statt 1 (vF4); `Eintraege` liefert den eigenen Slice statt einer Kopie (vF5b); `appends` um 2 (vF6) | 3 von 3 rot |
| E06 | Session: Eintrag bei `ctx` nil (vA4), im Zweig `Sessions.Get`-Fehler (vA5), im Zweig offen (vP7) | 3 von 3 rot; leere Kennung (vP2) **grün** (V-5) |
| E08 | Workspace: Eintrag bei `ctx` nil (vW1) rot; bei Fehler von `Workspaces.Delete` nur für die leere Kennung (vW2) **grün** (V-5) | 1 von 2 |
| E10 | Eintrag per `defer` nach der Bereinigung (vA6 Session, vW4 Workspace) | 2 von 2 rot |
| E12 | Session/Workspace: Fehler von `Append` leert den Bericht (vA3, vW3); Aufbewahrung: Fehler zählt `SessionsGeloescht--` (vB3); Import: Fehler wird mit `errors.Join` an den Rückgabefehler gehängt (vI1) | 3 von 4 rot; vI1 **grün** (V-3) |
| E15 | Import: Eintrag bei erfolgreicher Wiederholung mit gleichem Hash (vI7); Session: Eintrag im Zweig `Get`-Fehler (vA5, rot) | vI7 **grün** (V-1) |
| E17 | `New` ruft `Append` (vA8 Session, vB5 Wartung, vW6 Workspace, vI6 Import) | 4 von 4 rot |
| M2 | Aufbewahrung: Eintrag im Zweig Kontextfehler (vB1, rot-paket); Eintrag für eine junge Session (vB2); Eintrag im Fehler von `Dienste.List` (vB4) und von `Sessions.List` (vP6) | vB1 rot-paket; vB2, vB4, vP6 **grün** (V-4) |
| M4 | Import: nur ein unverpackter `DeadlineExceeded` als Kontextfehler erkannt (vI3); nur ein unverpackter `Canceled` (vI8) | vI8 rot; vI3 **grün** (V-2) |
| M6 | Import: `UpstreamURL` im Feld `SessionID` (vI4); Fehlertext im Feld `ImportID` (vI5) | vI4 rot; vI5 rot-paket (INFO I-1) |

### Fix-Runde M-1 bis M-4, L-1, L-2 gegen die Klasse

| Befund | Klasse | Probe an nicht genanntem Element | Ergebnis |
|---|---|---|---|
| M-1 (Dienst-Achse) | Dienst × Folge der Konfiguration | mittlerer Dienst `[len/2]` in Aufbewahrung (vD1) und Import (vD2); genannt waren erster und letzter | 2 von 2 rot |
| M-2 (Status-Achse) | Status 3 + `open` | doppelter Eintrag nur bei `incomplete` (vA1), nur bei Status außerhalb der Menge (vA2) | 2 von 2 rot |
| M-3 (Fake nebenläufig) | Zweige × Zugriffe | Zähler `appends` ohne Sperre (zL1), Festhalten ohne Sperre (zL2), `lesungen` vor der Sperre (zL3), `Aufrufe` liest `lesungen` ohne Sperre (zL4) | Detektor **20 von 20** je Mutation; ohne Detektor 14, 18, 2, 0 von 20 (die Zusage gilt über den Detektor, wie der Plan sagt) |
| M-4 (Brücke) | Instanz, Pfad, Uhr/Generator/Logger | Pfad im Elternverzeichnis von `DATA_DIR` (vK2); `Build` legt die Datei an (vK4); eigener Generator (vK3); `slog.Default()` als Logger (vK1) | vK2, vK4 rot; vK3 grün (im Plan als akzeptiertes Negativ geführt); **vK1 grün** (V-6) |
| L-1 (Paare Fake) | Fehler × Zähler × `Vor` | `Vor` erhält einen leeren Eintrag (vF7) | rot |
| L-2 (E19 Aufbewahrung) | Position × `ctx` | nicht erneut gefahren (Implementer 20 von 20 je Position, Plan §3); Stichprobe über die Mengen M7 nicht wiederholt | INFO |

### Zähler in nicht genannten Zweigen mit echtem Schreiben, 20 Prozesse mit Detektor

Klasse *nebenläufig*, je Mutation 20 Prozesse mit Detektor (`-race`, `-test.count=1`) und 20 ohne. Geschrieben wird ein Feld der Konfiguration (`Frist++`, `MaxAge++`, `MaxCaptureBody++`) oder `nil` in ein Port-Feld, nie `x = x`.

| Nr. | Zweig | Detektor | ohne | Ergebnis |
|---|---|---|---|---|
| zC4 | Session, vor `Sessions.Get` (alle Aufrufer) | **20/20** | 0/20 | flatternd, Detektor fängt |
| zC5 | Workspace, vor `Workspaces.Delete` (alle Aufrufer) | **20/20** | 0/20 | flatternd, Detektor fängt |
| zL1 | Fake, `appends++` ohne Sperre | **20/20** | 14/20 | flatternd |
| zL2 | Fake, Festhalten ohne Sperre | **20/20** | 18/20 | flatternd |
| zL3 | Fake, `lesungen++` vor der Sperre | **20/20** | 2/20 | flatternd |
| zL4 | Fake, `Aufrufe` liest `lesungen` ohne Sperre | **20/20** | 0/20 | flatternd |
| zC1 | Session, Zweig `Get`-Fehler | 0/20 | 0/20 | grün (der Zweig läuft im Paar nicht; einziger Schreiber) |
| zC2 | Session, Zweig offen | 0/20 | 0/20 | grün (Zweig läuft nicht) |
| zC3 | Session, nach `CleanupBlobs` | 0/20 | 0/20 | grün (nur der Gewinner des Paares schreibt) |
| zC6 | Workspace, Zweig Fehler von `Workspaces.Delete` | 0/20 | 0/20 | rot-paket (anderer Test rot, nicht das Rennen) |
| zC7 | Workspace, vor `CleanupBlobs` | 0/20 | 0/20 | rot-paket |
| zC8 | Aufbewahrung, im Löschzweig | 0/20 | 0/20 | grün (kein Test ruft die Wartung gleichzeitig) |
| zC9 | Import, Fehlerzweig | 0/20 | 0/20 | grün (kein gleichzeitiger Import) |
| zC10 | Import, am Kopf | 0/20 | 0/20 | rot-paket |

Zehn Zähler in Zweigen, die weder Plan noch Review nannten (zC1 bis zC10), dazu vier am Fake; Dienste mit gleichzeitigem Fänger (Session, Workspace) fangen das Rennen, wo zwei Aufrufer den Pfad ausführen. Die grünen Zeilen sind das benannte Negativ von `SPEC-234` *Ungeprüft* (ein Datenrennen, das kein Test auslöst; die Dienste halten keinen Zustand): kein Befund, nur sichtbar gemacht. Die Raten sind Größenordnungen unter Last des Hosts (3,4 bis 5,0), keine festen Werte.

## Befunde

**HIGH: keine.** Keine Verletzung eines ADR, kein Geheimnis in einer Senke, kein `//nolint`, keine Datei der Gate-Kette berührt, `make gates` grün.

### V-1 (MEDIUM): Eine gelungene Wiederholung mit gleichem Hash schreibt keinen Eintrag, ohne Fänger

- `quelle`: `SPEC-234` Nr. 4 *Kein Eintrag* („bei der Wiederholung mit gleichem Hash“), Nr. 10 (ein gelungener Import danach schreibt keinen); `AGENTS.md` §3.15; Plan E09, E15
- `pfad`: `src/hexagon/services/baselineimport/importer.go` · `i.wiederhole(ctx, a, hash)` mit `fertig && err == nil`
- `befund`: vI7 (Eintrag `import_fehlgeschlagen` bei erfolgreicher Wiederholung) bleibt grün, auch im ganzen Paket. `TestAuditKeinEintrag` führt die scheiternde Wiederholung (zwei Stellen × zehn Fehler) und einen ersten Erfolg, `TestAuditWiederholung` zwei gescheiterte Versuche und einen Erfolg danach; die gelungene Wiederholung der Vertragszeile (200, `Neu` falsch) zählt kein Test am Audit.
- `klasse`: Fall des Vertrags ohne Zeile in der Menge (Zusage ohne Fänger)

### V-2 (MEDIUM): Kontextfehler beim Import: verpackter `DeadlineExceeded` ungefangen

- `quelle`: `SPEC-109` Nr. 6, `SPEC-234` Nr. 4 d; `AGENTS.md` §3.17 (Menge Art × Form); Plan M4
- `pfad`: `importer.go` · `err != nil && !kontextFehler(err)`
- `befund`: vI3 (nur der unverpackte `DeadlineExceeded` zählt als Kontextfehler) bleibt grün. `kontextFehler()` im Test führt `Canceled`, `DeadlineExceeded` unverpackt, `Join(Canceled, ErrBlobFehlt)` und `fmt.Errorf("%w", Canceled)`: die Form „verpackt“ gibt es nur für `Canceled`. vI8 (nur unverpackter `Canceled`) wird rot. Die Menge hat also vier Formen, die Klasse Art (2) × Verpackung (3) sechs Elemente. Folge: ein Umbau, der verpackte Fristüberschreitungen nicht erkennt, schriebe `import_fehlgeschlagen` für den Abbruch des Aufrufers.
- `klasse`: Menge schmaler als die Klasse

### V-3 (LOW): E12 im Import vergleicht nur Code und Status

- `quelle`: `SPEC-234` Nr. 5 b („Rückgabe, Bericht, Fehler … dieselben“); `AGENTS.md` §3.15
- `pfad`: `baselineimport/audit_test.go` · `TestAuditFehlerAendertNichts` (`a.Code != b.Code || a.Status != b.Status`)
- `befund`: vI1 (Fehler von `Append` per `errors.Join` an den Rückgabefehler gehängt) bleibt grün. Der Fehlertext der Ablage (im Test `Ablage ` + Marker) darf damit im Rückgabefehler stehen. Session, Workspace und Aufbewahrung sind gefangen (vA3, vW3, vB3 rot); nur der Import trägt einen Fehler als Rückgabe.
- `klasse`: Fänger schmaler als Zusage

### V-4 (MEDIUM): Aufbewahrung: Ablehnungen vor `Sessions.Delete` nur an den Zweigen der Menge M2

- `quelle`: `SPEC-234` Nr. 4 *Kein Eintrag* („jedem Fehler eines Ports“, „je Session, deren `Delete` `nil` lieferte“); `AGENTS.md` §3.15 (Allaussage), §3.17; Plan M2 (5 Elemente)
- `pfad`: `maintenance/startup.go` · `aufbewahrung`, Zweige vor `loesche`
- `befund`: vB2 (Eintrag für eine Session, die nicht alt genug ist), vB4 (Eintrag im Fehlerzweig von `Dienste.List`) und vP6 (im Fehlerzweig von `Sessions.List`) bleiben grün, auch im Paket. M2 nennt für die Aufbewahrung `ErrReferenziert`, `ErrNichtGefunden`, anderer Fehler, `MaxAge` 0, ohne Session. Die junge Session und die beiden Listenfehler stehen nicht in der Menge; die Allaussage des Vertrags („jeder Fehler eines Ports“) trägt sie.
- `klasse`: Menge lässt Elemente der Allaussage aus

### V-5 (LOW): Raster-Zelle „leere Kennung“ ohne Fänger für Session und Workspace

- `quelle`: Plan §6 Raster (Spalte *leere Kennung* mit Ort `SPEC-124` Nr. 7 b, `SPEC-129` Nr. 9 b); `AGENTS.md` §3.16
- `befund`: vP2 (Session) und vW2 (Workspace): ein Eintrag im Ablehnungszweig, nur wenn `id == ""`, bleibt grün. Der Fänger der Ablehnung führt `ErrNichtGefunden` mit einer nichtleeren Kennung. Der Code verzweigt nicht nach der Kennung; die Mutation ist künstlich, die Zelle ist trotzdem ohne Fall.
- `klasse`: Raster-Zelle mit Ort, ohne Fänger

### V-6 (LOW): Brücke: der eine Logger ist ungefangen

- `quelle`: `SPEC-139` Nr. 3 („mit der einen Uhr, dem einen Generator und dem einen Logger“), Plan §3 *Akzeptierte Negative der Brücke* nennt Uhr und Generator, nicht den Logger
- `pfad`: `src/bootstrap/wurzel.go` · `Log: h.log`
- `befund`: vK1 (`slog.Default()`) bleibt grün. Folge: der Adapter protokolliert Fehler der Ablage (`SPEC-235` Nr. 5) an einem anderen Ziel als der Prozess, ohne dass ein Test es sieht. Der Plan führt das nicht als Negativ.
- `klasse`: Zusage ohne Fänger, nicht als Negativ benannt

### V-7 (MEDIUM, Planner): `slice-audit-verdrahtung` kennt die Brücke nicht

- `quelle`: `AGENTS.md` §3.13, §3.9
- `pfad`: `docs/plan/planning/open/slice-audit-verdrahtung.md` (kein Treffer für „Brücke“)
- `befund`: die Brücke setzt den Kern von `slice-audit-verdrahtung` um (Instanz, Pfad, Konfigs); dessen §1 Ziel („Der Composition Root erzeugt … `audit.New` …“) und V04 (dieselbe Instanz) sind schon erfüllt und durch `TestAuditEineInstanz`, `TestAuditPfadUndZeile` doppelt belegt. Der Plan von `slice-audit-dienste` führt die Entscheidung als §6-Zeile *offen*, Adresse Planner (richtig nach §3.13), aber der Nehmer ist nicht nachgezogen. Die Entscheidung ist nicht meine; Sachlage unten.

### INFO

- **I-1:** Der Fänger `TestAuditKeinGeheimnis` des Imports bleibt bei vI5 (Fehlertext im Feld `ImportID`) grün; ein anderer Test des Pakets wird rot, weil der Wert der Kennung nicht stimmt. Gleiche Klasse wie I-3 des Reviews: Der Schutz ist doppelt, E18 allein schmaler als die Zusage.
- **I-2:** zC1 bis zC3, zC8, zC9: Datenrennen in Zweigen, die kein Test mit zwei Aufrufern ausführt, bleiben ungesehen. Das ist das Negativ von `SPEC-234` *Ungeprüft* (Datenrennen, Dienste ohne Zustand), mit Grund und Folge dort; kein Befund.
- **I-3:** Die Nachfahrt der Mengen M7 und des Fensters F3 (E19, Aufbewahrung, Position) hat dieser Bericht nicht wiederholt.

## Sachlage zur Brücke (`src/bootstrap/wurzel.go`, Commits `5b7636e`, `d210747`) — bewertet, nicht entschieden

- **Wirkung auf das Binary:** `Build` erzeugt `audit.New` über `<DATA_DIR>/audit.jsonl` und gibt die eine Instanz allen vier Konfigs (record: vier Träger, replay: zwei). `Build` legt die Datei nicht an (`TestAuditPfadUndZeile`, vK4 rot). Beobachtbar ändert sich das Binary beim ersten Eintrag: `start` ruft `h.wartung.Startup` (`lebenszyklus.go:42`); löscht die Aufbewahrung (`RETENTION_MAX_AGE` größer 0) eine Session, entsteht `audit.jsonl` im Datenverzeichnis. Eine HTTP-Löschung gibt es im Binary noch nicht (kein Admin-Listener in diesem Slice).
- **Ohne die Brücke:** `New` weist ein nil-`Audit` ab, `Build` scheitert, 26 Tests unter `src/bootstrap/` und 19 unter `tests/integration/` sind rot (Plan §6, hier nicht einzeln nachgezählt); `make gates` wäre rot. Der Revert von `5b7636e` allein lässt `Build` rot (der zweite Commit trägt nur die Tabellen der Konfig-Felder).
- **Handbuch-Pflicht:** `make handbuch-check` ist grün. Das Handbuch sagt nichts über Aufbewahrung, `DATA_DIR`, Sessions oder Löschen (grep ohne Treffer), keine Aussage wird falsch. Die DoD-Zeile verlangt, dass §7 bei Brücke die Datei `audit.jsonl` im Datenverzeichnis nennt; ein Handbuch-Kapitel dazu legt `slice-audit-verdrahtung` §1 in `slice-admin-sessions` (erster Slice, dessen Binary eine Löschung auslöst).
- **Verhältnis zu `slice-audit-verdrahtung`:** Kern (Instanz, Pfad, Konfigs) ist durch die Brücke geliefert und getestet; offen bleiben dessen V01 (Zeilen in der Datei über das `Handle`), V02 (Neustart), V03 (Canary), V05 teilweise (durch `TestAuditPfadUndZeile` schon belegt). Der Slice verliert den Kern seines Ziels, behält die Fänger. `src/bootstrap/` ist in beiden Plänen aufgeführt; `slice-bootstrap-m2` liegt in `done/`, auf `main` hat sich seit `1543b02` unter `src/bootstrap/` nichts geändert (`git diff` leer), ein Merge-Konflikt entsteht heute nicht.
- **Größenmarke:** die Brücke zählt nicht in die Messung (117 Zeilen außerhalb der `pfad`-Zeilen); mit ihr wären es 2175, noch unter 3000.
- **Optionen laut Plan §6:** Brücke bleibt und `slice-audit-verdrahtung` wird verkleinert (V-7); Brücke entfällt und beide Slices werden gemeinsam geschnitten; oder `slice-audit-verdrahtung` läuft vor diesem Slice.

## Nicht aufgelöst

- E05 (Dienst-Achse und Zeichen-Achse von M5, 37 und 65 Zeichen) und M5 Länge nicht neu gefahren (Review: Länge 32 und Dienst; hier nur Fix-Runde M-1).
- E09 (63 Teilmengen) und E14 (Fehlerstellen × Codes, Orakel) nur über vI2, vI3, vI8, nicht über die Teilmengen.
- E13 für Aufbewahrung und Import nur über vA9, vW7 (Session, Workspace) und vI3/vI8.
- E16 (32 Paare mit Barriere) nur über Zähler zC4/zC5, nicht über die Barriere selbst; E19 (Fenster F1 bis F3) nicht neu gefahren; die Raten der 104 Mutationen des Implementers (§3) und die 25 der Fix-Runde nicht nachgefahren.
- `L-2` aus dem Review (E19 an der dritten Session der Aufbewahrung, Rate) nicht neu gemessen.
- Last: alle Läufe bei Host-Last 3,4 bis 7,1; kein Lauf auf ruhigem Host.

## Empfehlung Closure

**Ja, nach Nachzug** — kein Wiederholungslauf der Verifikation nötig, wenn der Implementer die Fänger nachzieht und seine Tests selbst rot sieht.

Der Code erfüllt `SPEC-234` an allen Nummern, `make gates` ist grün, keine Fehlfunktion gefunden. Die Befunde sind Fänger-Lücken an Randformen. Closure bleibt blockiert, bis die Punkte unten stehen.

### Nachzugsliste

**Implementer:**
1. V-1: Fall „gelungene Wiederholung mit gleichem Hash schreibt keinen Eintrag“ in `TestAuditKeinEintrag` oder `TestAuditWiederholung` (Mutation vI7 rot).
2. V-2: die Menge der Kontextfehler des Imports als Art (`Canceled`, `DeadlineExceeded`) × Verpackung (unverpackt, `%w`, `errors.Join` neben Fehler mit Code) = 6 Formen statt 4 (Mutation vI3 rot, vI8 bleibt rot); Zahlen in §3 und M4 nachziehen (`AGENTS.md` §3.9).
3. V-4: Elemente „junge Session“, „Fehler von `Dienste.List`“, „Fehler von `Sessions.List`“ in die Menge M2 der Aufbewahrung und in `TestAuditPositionUndErgebnis` (vB2, vB4, vP6 rot); M2 in §3 von 5 auf 8 Elemente.
4. V-3: `TestAuditFehlerAendertNichts` des Imports vergleicht den Fehlertext (oder Identität) mit dem Lauf ohne Fehler (vI1 rot).
5. V-5 und V-6: entweder Fänger (leere Kennung für `Delete` von Session und Workspace; Logger-Zeiger der Instanz über das Handle) oder akzeptiertes Negativ mit Grund und Folge in §3.
6. §6: Ausgang für *Randformen* (Review liegt vor), §4: `slice-bootstrap-m2` als erledigt führen; Closure-Notiz §7 samt Zeile zur Datei `audit.jsonl` und Register.

**Planner:**
1. Entscheidung über die Brücke (V-7): Brücke bleibt, entfällt oder Reihenfolge der Slices; danach `slice-audit-verdrahtung` §1, V04 und V05 nachziehen oder die beiden Slices zusammen schneiden (`AGENTS.md` §3.13).
2. §6-Ausgang der Brücke und der Hinweis „weiter offen“ der übrigen Zeilen bei der Closure auflösen (Planner, nicht Implementer).
3. Handbuch-Kapitel zur Datei bleibt bei `slice-admin-sessions` (Planner bestätigt oder verlegt).
