# Review-Report: slice-bootstrap-m3-e2e-mock — 2026-11-08

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-bootstrap-m3-e2e-mock`, Code-Commits `d2f848ae`, `ba512b40`, `854817be` (Basis `fc6d1a6b`), dazu `56c413d1` des Architect (Spezifikation und Plan §3, §6) als Eingang; HEAD zu Beginn `854817be`, während des Laufs `56c413d1`.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-08

**Eingangs-Kontext:**

- Plan von `slice-bootstrap-m3-e2e-mock` (§3 Tabelle E01 bis E09, Absatz *Mutationen*, *Raten*, *Mengen*, *Grenzen*; §6 Raster und Ausgänge)
- `SPEC-037` Nr. 5 bis 7 und 10 i, j, k, l; `SPEC-139` Nr. 11 d und f; `SPEC-195` Nr. 5 und 8
- `LH-FA-12`, `LH-FA-09`, `LH-FA-03`
- `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.18
- vorheriger Report am Modul: `2026-11-08-slice-mock-abnahme-review.md` (Form)

**Nicht geprüft (Rolle des Verifiers oder Planners):** DoD, `make gates`, ein voller Lauf von `make test-integration` (das Image `odata-recorder:integration` ist über Worktrees geteilt, ich habe es nicht neu gebaut), die Verdrahtung der Zeilen E07 im Treiber (Mutation *Name der Stufe* und *Anker* nicht gefahren, nur gelesen), Größenauslastung (Summe `git diff --numstat` über `tests/e2e` und den Treiber: 567 von Marke 1100, Tests eingeschlossen).

## Eigene Mutationen

Messender ist der Reviewer-Agent. Jede Mutation lief in einer Wegwerfkopie von `src/`, `tests/`, `go.mod`, `go.sum` (Quelle des Worktrees schreibgeschützt), im Container aus `odata-recorder:deps` ohne Netz, `--memory 4g`, ein Container zur Zeit unter `flock /tmp/odata-recorder-heavy.lock`; gebaut wurden Binary, Übergang und Test-Binary, gefahren die Stufe mit `-test.count=1` und `go test` des Pakets `tests/e2e/uebergang`. Jede Mutation baut (eine Mutation, die nicht baute, ist verworfen und wiederholt). Der Host war belastet (Last 16 bis 134); je Mutation ein Lauf, keine Raten.

Gewählt sind die Mutationen der Tabelle des Implementers als Nachmessung und Elemente der Mengen, die er nicht nannte.

| Nr. | Zeile · Zusage / Menge | Mutation (Wegwerfkopie) | Ergebnis |
|---|---|---|---|
| R1 | E01 `Location` aus `LISTEN` | `basisAusListen` liefert festen Host `localhost` | rot `TestE2EMock/CRUD` |
| R2 | E01, E03 Start ohne `UPSTREAM_URL` | `record := true` in `liesDienste` (Pflicht in jedem Modus) | rot `TestE2EMock` und beide Teile von `TestE2EMockNeustart`, `Binary nicht bereit` |
| R3 | E02 Reset wirkt | `reset` des Übergangs ruft `Workspaces().Get` statt `Reset` | rot `TestE2EMock` und `TestReset` (Revisionen 0, 0) |
| R4 | E02, nicht genannt: der Reset entfernt eine angelegte Entity | `ReplaceAll` der Ablage löscht nicht (`SELECT` statt `DELETE`) | rot `TestE2EMock` (Collection nach dem Reset) |
| R5 | E02 leerer Wert ist `400` | `w[0] == ""` entfällt in `einzelwert` | rot `TestE2EMock/Zuordnung` |
| R6 | E02 zwei Werte sind `400` | `len(w) != 1` wird `len(w) < 1` | rot `TestE2EMock/Zuordnung` und `/CRUD` (der `DELETE` ändert den Zustand) |
| R7 | E02 Klasse *mehr als ein Wert*, nicht genannt: zwei gleiche Werte | `einzelwert` lehnt nur ab, wenn die Werte verschieden sind | rot `TestE2EMock/Zuordnung` und `/CRUD` |
| R8 | E02 Klasse *mehr als ein Wert*, nicht genannt: zwei **verschiedene** Werte | `einzelwert` lehnt nur ab, wenn die ersten zwei Werte gleich sind | **grün** `TestE2EMock` (siehe F1) |
| R9 | E02 *zwei Schreibweisen* (Nachmessung der Aussage in `SPEC-037` Nr. 10 l) | `werteVon` des Mock-Dienstes liest nur die erste Schreibweise (`namen[:1]`) | rot `TestZuordnung` im Prozess (Zahl der roten Zellen nicht gezählt); am Binary nicht fangbar, wie Nr. 10 l sagt |
| R10 | E04 Mutationen überleben das Signal | `Herunterfahren` löscht `-wal` und `-shm`, schließt die Ablage nicht | rot `TestE2EMockNeustart/Mutationen…` (auch `TestReset` des Übergangs, der dieselbe Funktion nutzt) |
| R11 | E05 fehlendes `DATA_DIR` wird angelegt | `MkdirAll` wird `Stat` in `legeDatenverzeichnisAn` | rot `TestE2EMockNeustart/DATA_DIR_besteht_nicht` (`Binary nicht bereit`) |
| R12 | E06 Exit 1 und 2 getrennt | die Konstanten `exitFehler` und `exitAufruf` vertauscht | rot `TestAufrufformen` und `TestReset` |
| R13 | E06 `reset` legt `DATA_DIR` nicht an | die Prüfung auf ein vorhandenes Verzeichnis lässt ein fehlendes zu | rot `TestReset`: `reset hat das DATA_DIR angelegt` |
| R14 | E06 leeres Argument je Stelle | `args[1] == ""` entfällt; getrennt `args[2] == ""` entfällt | rot `TestAufrufformen`, je 3 Elemente (import, zeige, reset) |
| R15 | E06 ein Argument zu viel | `len(args) != 3` wird `len(args) < 3` | rot `TestAufrufformen`, 3 Elemente |
| R16 | E06, nicht genannt: unbekannter Workspace ist Exit 1 | `reset` meldet den Fehler des `Reset` nicht (Exit 0) | rot `TestReset` (unbekannter Workspace, Kennung einer Session) |

Eine erste Fassung der Mutationen *Prüfung entfällt* (nil-Zugriff auf `FileInfo`) und *`os.Stat` statt `MkdirAll`* (Übersetzungsfehler) war ungültig und ist verworfen; R11 und R13 sind die wiederholten Läufe.

Nicht gefahren: E07 (Anker, Name der Stufe im Treiber; braucht den Bau des geteilten Image-Tags), E08 und E09 (Negative), E03 am Zähler (Negativ `SPEC-037` Nr. 10 k, vom Architect entschieden, hier nicht neu gefunden), die Raten der Zeile E04 (nicht nachgemessen; R10 ist ein Lauf, rot).

## Findings

### F1 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15 (ein Fänger führt jede Form der Eingabe, die der Code verschieden behandelt; Allaussage über eine Menge)
- `pfad`: `tests/e2e/stufe_mock_test.go` · `{"zwei gleiche Werte", "ODR-E5003", 400, []string{"X-Recorder-Workspace: " + ws, "X-Recorder-Workspace: " + ws}},`
- `befund`: `SPEC-195` Nr. 5 nennt „mehr als ein Wert"; die fünf Formen am Binary führen nur gleiche Werte (die Form *zwei Schreibweisen* ist am Binary dieselbe, `SPEC-037` Nr. 10 l). Die Mutation R8 (nur zwei gleiche Werte abgelehnt, zwei verschiedene angenommen) bleibt am Binary grün; im Prozess fängt sie `TestZuordnung`, Form *zwei verschiedene*.
- `verifizierbar`: ja — R8 in einer Baumkopie, `TestE2EMock` grün
- `klasse`: Fänger führt von einer Klasse nur die gleichen Werte (Fixture gleicht sich selbst)

### F2 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.9 (der Plan folgt jeder Korrektur), §3.15 (Zusage ohne Fänger), `SPEC-037` Nr. 6
- `pfad`: `tests/e2e/uebergang/uebergang.go` · `upstream := os.Getenv(upstreamVar)` gegen `SPEC-037` Nr. 6 · `` `reset` und `zeige` lesen sie nicht ``
- `befund`: Die Spezifikation sagt, nur `import` lese `E2E_UPSTREAM_URL`; `handle` liest sie für jeden Befehl, auch für `reset` und `zeige` (über `session`). Kein Test setzt die Variable für diese Befehle, die Zusage hat keinen Fänger; ein beobachtbarer Unterschied entsteht nicht, weil der Upstream nie gerufen wird.
- `verifizierbar`: ja — Lesen von `handle`; ein Test mit gesetzter Variable an `reset` zeigte den Widerspruch nur über den Code, nicht über ein Ergebnis
- `klasse`: Spezifikation sagt weniger als der Code tut, Zusage ohne Fänger

### F3 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.16 (jede Zelle des Rasters hat einen Fänger oder ein benanntes Negativ), §3.15
- `pfad`: Plan §6 Raster, Zeile *Start auf vorhandenem DATA_DIR*, Zelle *leerer Wert* · „eine Ablage ohne Baseline und ohne Metadaten ist `ODR-E7001` und `ODR-E6004`“; `SPEC-139` Nr. 11 f
- `befund`: Der Ort steht in der Spezifikation, aber `ODR-E6004` (Service ohne Metadaten) ist am Binary nicht erreichbar, weil `ODR-E7001` (Workspace unbekannt) vorausgeht und es im Modus `mock` keinen Weg zu einem Workspace ohne Service-Metadaten gibt; E05 prüft allein `ODR-E7001`. Weder die Tabelle E01 bis E09 noch `SPEC-037` Nr. 10 nennen für diese Hälfte der Zelle einen Fänger oder ein Negativ.
- `verifizierbar`: nein — Lesen von Raster, `SPEC-195` Nr. 3 (Reihenfolge f vor g) und Tabelle
- `klasse`: Rasterzelle mit Ort, aber ohne Fänger oder Negativ

### F4 · LOW

- `kategorie`: LOW
- `quelle`: Maintainability (Struktur der Vertragsnummer; `AGENTS.md` §3.9 Nachzug aller Stellen)
- `pfad`: `harness/vertraege/SPEC-037-test-integration.md` · `` `E2E_UEBERGANG reset <DATA_DIR> <Workspace-ID>` baut das Handle wie `import` ``
- `befund`: Der Absatz zu `reset` steht ohne Einrückung zwischen `zeige` und dem Satz *„Ein fehlendes oder leeres Argument ist Exit 2 (`Aufruf:`), eine unbekannte Session … `ODR-E7003`"*, der damit nach `reset` steht und für `reset` (unbekannte Session ist dort `ODR-E7001`) gelesen werden kann; die Exit-2-Regel steht in Nr. 6 zweimal.
- `verifizierbar`: nein — Lesen der Nr. 6
- `klasse`: eingefügter Absatz trennt eine Regel von ihrem Gegenstand

### F5 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.11 (ein Kommentar sagt nur zu, was ein Test prüft)
- `pfad`: `tests/e2e/stufe_mock_test.go` · `// zuordnungsFormen sind die fünf Formen des Headers` und `{"zwei Schreibweisen", …}`
- `befund`: Der Kopf nennt fünf Formen, der Name der Form *zwei Schreibweisen* und die Zahl 25 zählen eine Form, die am Binary mit *zwei gleiche Werte* zusammenfällt; Nr. 10 l der Spezifikation trägt das Negativ, der Kopf der Funktion nennt es nicht. Kein Befund gegen das Verhalten, nur gegen den Wortlaut des Kopfes (Rolle: Implementer, falls gewünscht).
- `verifizierbar`: nein
- `klasse`: Name einer Form sagt mehr, als der Test trennt

### F6 · INFO

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.10 (Klasse *nebenläufig*, Wirkungs-Fänger, Raten)
- `pfad`: Plan §3 Absatz *Raten* · `Die Zeile gilt als gefangen.`
- `befund`: Bewertung der Zeile E04: der Test ist sequenziell, `SIGTERM` folgt der letzten Antwort, der Fänger ist ein Wirkungs-Fänger gegen eine Ablage, die der Neustart liest; die Raten 30 von 30 und 10 von 10 bei Last 18,4 bis 19,3 mit Zahl der Prozessläufen und Messendem stehen im Plan und entsprechen §3.10. Meine Nachmessung ist ein Lauf je Mutation (R10 rot unter Last 16 bis 134), keine Rate; ich sehe keinen Anlass, die Zahlen des Implementers zu bezweifeln. Die Zeile fängt kein Datenrennen im Binary, das der Plan ausdrücklich ausnimmt.
- `verifizierbar`: ja — R10
- `klasse`: nebenläufig benannte Zeile mit sequenziellem Wirkungs-Fänger

## Zu den drei gemeldeten Punkten (nicht neu gefunden, bewertet)

1. `E2E_UPSTREAM_URL`: durch `56c413d1` in `SPEC-037` Nr. 6 entschieden; der Code stimmt mit dem Wortlaut bis auf den Satz zu `reset` und `zeige` überein (F2).
2. E03 (Zähler): akzeptiertes Negativ mit Grund und Folge in `SPEC-037` Nr. 10 k; die Zusage trägt der Start ohne `UPSTREAM_URL`, R2 rot am Start. Die Folge ist ehrlich benannt (ein Pfad mit fest eingebauter Adresse bleibt ungefangen). Kein Befund.
3. E02 *zwei Schreibweisen*: Negativ in Nr. 10 l; die Aussage, `TestZuordnung` im Prozess fange eine `werteVon` mit nur einer Schreibweise, habe ich mit R9 bestätigt (rot). F1 betrifft die Klasse *verschiedene Werte*, nicht diesen Punkt; F5 den Wortlaut des Kopfes.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tests/e2e/stufe_mock_test.go` (CRUD, Reset, Neustart, `DATA_DIR`) | geprüft; ohne Befund außer F1, F5 |
| `tests/e2e/lauf_test.go` (Umgebung des Übergangs) | geprüft, ohne Befund |
| `tests/e2e/uebergang/` (`reset`, Argumentformen, `TestReset`) | geprüft; ohne Befund außer F2 |
| `tools/harness/test-integration.sh` (Zeilen der Stufen 8 und 9) | geprüft: Name, Anker, drei Zeilen je Stufe wie Nr. 3; Zeilennummern in `e2e-abdeckung.md` (131, 135) stimmen mit den `stufe`-Aufrufen überein; ohne Befund |
| `docs/plan/planning/e2e-abdeckung.md` | geprüft, ohne Befund |
| `harness/vertraege/SPEC-037-test-integration.md` | geprüft; F4 |
| Plan §1, §3, §6 | geprüft; F3 |
| Hard Rules §3.2 (keine `//nolint` im Diff), §3.6 (kein Gate gelockert, keine Zeile in `Makefile`, `Dockerfile`, `harness/mk/`), §3.18 (Fixtures aus Bytewerten; im Diff keine Escape-Folge als Wert) | geprüft, ohne Befund |
| Größenmarke (§3.14) | 567 von 1100, Metrik wie im Plan, ohne Befund |

## Summary

HIGH 0 · MEDIUM 0 · LOW 4 · INFO 2.

Von den 16 Mutationen R1 bis R16 werden 15 aus dem benannten Grund rot (R8 bleibt am Binary grün); von vier Elementen, die der Implementer nicht nannte (R4, R7, R8, R16), bleibt eines grün (R8, F1, im Prozess gefangen). Keine Rasterzelle ist im Code entschieden; F3 betrifft den Fänger einer entschiedenen Zelle.
