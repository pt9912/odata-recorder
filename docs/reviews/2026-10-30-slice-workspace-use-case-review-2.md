# Review-Report 2: slice-workspace-use-case, Fix-Runde — 2026-10-30

**Review-Art:** Re-Review der Fix-Runde, geprüft gegen Plan (`slice-workspace-use-case` §3 Z01 bis Z70, Nachträge, §6), `SPEC-129` und die Hard Rules in `AGENTS.md` (§3.7 bis §3.18). Nicht geprüft wird gegen die DoD (Aufgabe des Verifiers). Erster Review: `2026-10-30-slice-workspace-use-case-review.md` (Commit `1e1ea2d`).

**Gegenstand und gemessener Stand:** Zweig `slice-workspace-use-case`, **gemessen auf `9e61be6`** (HEAD; Fix-Commits `fcda8de` = SPEC-129 und `9e61be6` = Tests, Code, Plan; Basis des Slice `6cf9202`). Die Wegwerfkopie `rev2-workspace/base` ist aus `git archive 9e61be6` und per `diff -r` byte-gleich mit `src/` im Worktree. Mutationen des ersten Reviews gelten hier nur, weil die Produktdateien (`anlegen.go`, `reset.go`, `loeschen.go`, `manifest.go`, `protokoll.go`, `fehler.go`) seit `1e1ea2d` unverändert sind (Diff: nur `verwaltung.go`, `idGueltig`).

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-30

**Methode.** Der Bericht des Implementers ist gelesen, keine Zahl daraus übernommen. Werkzeuge im Scratchpad `.../scratchpad/rev2-workspace/` (nichts im Repo außer diesem Report): Mutationsläufer (`inner.sh`, `run.sh`; Quelle `:ro`, netzlos, `--memory=4g --memory-swap=4g --cpus=4`, `flock /tmp/odata-recorder-heavy.lock`, jede Mutation auf frischer Kopie, `go test -count=1` über das Paket, Image `odata-recorder:deps`), Ratenläufer (`inner-race.sh`: ein `-race`-Binary, 50 Prozesse mit `-test.count=1`, ein Prozess mit `-test.count=50`, Image `odata-recorder:race`; Schleife A = Detektor allein mit `-test.skip OhneZustand`, Schleife B = `TestPaketOhneZustand`), Probe-Paket `alt/probe2_test.go` (Saat 100000). **Last:** Host-Last beim Start 50, zeitweise 130 bis 190 durch Fremdlast; die Raten sind Raten gegen das Image unter dieser Last. Messender: der Reviewer. Der Container `rev-busy` stammt nicht von diesem Lauf.

Gefahren, mit Exit-Code gelesen: `make test` EXIT=0 (gofmt, vet, Unit-Tests), `make a-check` EXIT=0 (0 Befunde), `make groessenmarke` EXIT=0 (`slice-workspace-use-case.md`: Messung 5424 von Marke 10600, wie der Plan), `make docs-check` ist gestartet, Ergebnis bei Abgabe offen (siehe *Ungelaufen*).

**Ungelaufen bei Abgabe** (Warteschlange des Hosts): (a) die Wiederholung der Mutationen Paketvariable, `go`-Anweisung in `protokoll.go`, `sync`-Alias/Punkt-/Raw-Import, `time.AfterFunc`, `var _ = f()` (`mz4.txt`, die erste Fassung scheiterte am `sed` von BusyBox: `NOCHANGE` bzw. Aufbaufehler, **nicht gewertet**); (b) die Detektor-Raten der Paketvariablen-Mutanten `n57b/c/d` und `n47`-Wiederholung (`mr5` läuft, Zeilen bis `n57` fertig; `mr6` für die drei neuen Mutanten wartet); (c) `make docs-check`, `make lint`, `make test-race`, `make gates`. Zeilen dazu stehen unten nur, wo gemessen.

---

## Antworten auf die Fragen des Auftrags

**F-1 (`TestPaketOhneZustand`, Klasse „Zustand im Dienst“).** Gemessen (Mutation, Ergebnis `go test`):

| Mutation | Test rot |
|---|---|
| `func init()` in `manifest.go`, in `fehler.go` | `TestPaketOhneZustand` (2 von 2) |
| `go func(){}()` in `leseManifest`, in `Create` | `TestPaketOhneZustand` (2 von 2) |
| `_ "sync"` in `verwaltung.go` | `TestPaketOhneZustand` |
| Feld `n int` nach `k`, vor `k`, `_ int`, `k, n Konfig`, eingebettetes `sync.Mutex`, `Verwaltung` in `type (…)`-Gruppe | `TestPaketOhneZustand` und `TestVerwaltungOhneZustand` (6 von 6) |
| `k *Konfig` | beide |
| zehntes Feld in `Konfig` | `TestNewNilFeld`, `TestPortUndKonfigFestgeschrieben` (nicht `TestPaketOhneZustand`) |
| `type Verwaltung verwaltung` mit zweitem Feld im Typ `verwaltung` | nur `TestVerwaltungOhneZustand` (Reflexion; so zugesagt) |
| neue Datei mit Paketvariable, neue Datei ohne Zustand | `TestPaketOhneZustand` (Dateizahl 7) |
| Kontrollen: lokale Variable, ohne Änderung | grün (kein Fehlalarm) |
| **Zustand als Schreibzugriff auf `v.k` (`v.k.Uhr = u`) in `Create`, `Reset`, `Delete`** | **grün, alle Tests (3 von 3)** |
| **eigener Typ mit Zähler, in `New` vor `Konfig.Workspaces` bzw. `Konfig.Blobs` gehängt** | **grün, alle Tests (2 von 2, `rc=0`)** |

Raten des Detektors allein (Schleife A, 50 Prozesse, ohne die zwei `*OhneZustand`-Tests) für die Schreibzugriffe auf `v.k.Uhr` nach dem ersten Port-Aufruf: `Create` **12 von 50**, `Reset` **0 von 50**, `Delete` **0 von 50** (ein Prozess mit `-test.count=50`: 3, 1, 0 Fehlerzeilen). Normalstand 0 von 50. Die Zahlen des Reviewers (21, 3, 0 von 50 für den Zähler in `Create`, `Reset`, `Delete`) sind in dieser Runde nicht wiederholt (`n57b/c/d` ungelaufen); die Zahl 12 von 50 für `Create` liegt unter 21, weil die Mutation eine andere ist (Schreibzugriff statt Zähler) — die Größenordnung („meist unentdeckt“) trägt. `n57` (Paketvariable am Anfang) 50 von 50 Detektor, `TestPaketOhneZustand` 50 von 50 Prozessen und 50 von 50 Läufen; `n47` (Einträge in Paketvariable nach dem Lesen) Detektor allein **1 von 50**, `TestPaketOhneZustand` 50 von 50 und 50 von 50.

Satz in `SPEC-129` *Ungeprüft*, Plan Z57 und Test stimmen für Paketvariablen, `init`, `go`-Anweisung, `sync`-Import und das Feld der `Verwaltung` überein; sie stimmen **nicht** für „Zustand außerhalb der Felder des Dienstes“ als Rest: der Rest ist größer als „fremdes Paket oder Port“ (F-5). Plan Z57 *Raten*: „jede Mutation 50 von 50“ gilt für die dort genannten Formen; für Schreibzugriffe und eigene Typen gilt sie nicht.

**F-2 (Leerraum).** Gemessen, `manifest.go`:

| Mutation | Ergebnis |
|---|---|
| Tab, CR, Leerzeichen, LF je aus `firstNonSpace` entfernt | rot (`TestManifestLeerraum`; Leerzeichen und LF auch `TestManifestGueltigeFormen`) |
| FF ergänzt, `json.Valid` bleibt | grün (äquivalent, `json.Valid` weist zuerst ab, wie im Plan) |
| FF bzw. VT ergänzt, `json.Valid` entfernt | rot |
| `bytes.TrimSpace(b)` am Anfang | rot |
| BOM getrimmt | rot (3 Tests) |
| NBSP nur davor, nur dahinter getrimmt | rot (je 1) |
| Leerzeichen innerhalb der Strings entfernt | rot (3 Tests) |
| **`TrimFunc` U+2028 / U+0085 / 0x1f / U+3000** | **grün (4 von 4)** |
| **Leerraum davor auf 4 Zeichen begrenzt (`i < 4`)** | **grün** |

Probe (Saat 100000, Orakel: angenommen genau dann, wenn alle Bytes davor und dahinter in {20, 09, 0A, 0D}): jedes der 256 Bytes davor, dahinter und beidseitig, alle 254 Runen der Klassen White_Space, Z, Cc, Cf, dazu 20 000 zufällige Läufe aus Leerraum (Länge 0 bis 39) mit Fremdzeichen an zufälliger Stelle: **21 276 Fälle (5 060 angenommen, 16 216 abgewiesen), 0 Abweichungen** (Code `ODR-E3001`, Status 500, `ErrBeschaedigt`). Der Dienst erfüllt die Klasse; der Fänger führt eine Auswahl (F-6).

**F-3 (Log-Fake).** Dienst-Mutationen `Log.WithGroup("g")` (beide Aufrufe, nur der zweite), `.With("x",1)`, `.With().WithGroup()`, `.WithGroup().With()`, `.WithGroup("g").WithGroup("h")`, Attribut als `slog.Group`: je rot in `TestProtokollNachDemLoeschen` (7 von 7); `WithGroup("")` grün (äquivalent: `slog` behandelt es als leer). Fake-Mutanten: Attribute aus `WithAttrs` ohne Präfix und Gruppen nicht kumuliert: **grün (2 von 2)**; Gruppe teilt `pre` nicht und `Handle` ohne Präfix: rot in `TestLogHandlerLeitetAb` (2 von 2). Die Schleife über die Methoden von `*slog.Logger` (2, mit Zahl) trägt die Klasse der Ableitungen; offen bleibt die Kombination (F-7).

**F-4, I-3, I-4.** Z02, Z07, Z09, Z61 und die Bericht-Zeilen sind nachgezogen (Diff gelesen). I-4: der tote Zweig ist entfernt, `idGueltig` hat einen Aufrufer (`anlegen.go`, `id != ""`), der Kommentar nennt die Voraussetzung. Rest: F-8.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-5 | MEDIUM | Die Zusage „der Dienst hält keinen Zustand“ hat für zwei Formen weiter keinen Fänger in jedem Lauf: (a) ein Schreibzugriff auf ein Feld von `v.k` nach `New` und (b) ein Typ des Pakets mit Zustand, den `New` in ein Feld von `Konfig` hängt. `TestPaketOhneZustand` und `TestVerwaltungOhneZustand` prüfen Paketvariablen, `init`, `go`, `sync`-Import und die Zahl der Felder, nicht deren Beschreiben; alle fünf Mutationen blieben in allen Tests grün. Der Detektor allein meldet (a) in `Create` nach dem ersten Port-Aufruf in 12 von 50, in `Reset` und `Delete` in 0 von 50 Prozessen (gleiche Ursache wie F-1: die gemeinsame Sperre des Fakes ordnet die Zugriffe). Szenario: ein später eingeführter Cache (`v.k.Uhr`, `v.k.Log`, ein zählender Wrapper um einen Port) bleibt in den meisten Läufen von `make test-race` unentdeckt und ist im Betrieb ein Datenrennen. `SPEC-129` *Ungeprüft* und Plan Z57 benennen als Negativ „Zustand in einem fremden Paket oder in einem Port“, nicht diese Formen; der Satz „ein Feld des Dienstes prüft die Reflexion“ meint nur das Vorhandensein. | `AGENTS.md` §3.10 (Zusage über Nebenläufigkeit gilt nur bei Rot in jedem Lauf; Wirkungs-Fänger oder Negativ mit Grund und Folge), §3.17 (Fix deckt die Klasse); Skill *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt* | `src/hexagon/services/workspace/zustand_test.go` · `func TestPaketOhneZustand(t *testing.T) {`; `spec/spezifikation.md` · `ein Zustand außerhalb der Felder des Dienstes`; Plan · `Akzeptiertes Negativ: Zustand in einem fremden Paket oder in einem Port.` | ja — Mutationen `y38`/`y39`/`y40` (Schreibzugriff), `z27`/`z28` (Wrapper), Raten in `out-mr2/res.txt` | Zustandsfänger deckt nur neue Felder und Paketvariablen, nicht Schreibzugriff und eigene Typen |
| F-6 | MEDIUM | Der Fix zu F-2 führt für „kein anderes Zeichen“ eine Auswahl von fünf Zeichen und für die Länge des Leerraums einen Lauf von höchstens vier Zeichen davor. Die Menge ist generierbar (256 Bytes, die Runen der Klassen Z/Cc/Cf; die Probe führt sie in 21 276 Fällen) und im Wortlaut der Zusage unbeschränkt („kein anderes Zeichen“); der Plan nennt weder Schleife noch Prüfgrenze mit Grund und Folge. Fünf Mutationen an Elementen, die weder Reviewer noch Implementer nannten, bleiben grün: Beschneiden von U+2028, U+0085, 0x1f, U+3000 und die Begrenzung des Leerraums davor auf 4 Zeichen. Der Dienst ist korrekt (Probe 0 von 21 276); die Lücke liegt im Fänger. | `AGENTS.md` §3.17 (Klasse, nicht genannte Mutation; unbeschränkte Menge: Prüfgrenze mit Grund und Folge), §3.15; Skill *Klasse generierbar benannt, im Handoff nur als Liste* | `src/hexagon/services/workspace/leerraum_test.go` · `keinLeerraum := [][]byte{{0x0C}, {0x0B}, {0xC2, 0xA0}, {0xEF, 0xBB, 0xBF}, {0x00}}`; Plan Z20 · `Nicht-Leerraum {FF, VT, NBSP, BOM, NUL}` | ja — Mutationen `l08` bis `l11`, `l14` (`out-ml/res.txt`), Probe `alt/probe2_test.go` | Fänger führt Auswahl statt Klasse der Zeichen und Längen |
| F-7 | LOW | Der Kopf des Log-Fakes sagt, er qualifiziere die Namen „wie ein Handler des Originals“; `With` nach `WithGroup` und verschachtelte Gruppen prüft kein Test: die Mutanten „`WithAttrs` ohne Präfix“ und „Gruppen nicht kumuliert“ bleiben grün. Eine Mutation im Dienst bleibt dadurch nicht grün (der Dienst-Fänger vergleicht die Namen und sieht den Unterschied schon an der Gruppe), das Versprechen im Kommentar ist größer als der Fänger. | `AGENTS.md` §3.11 (ein Kommentar sagt nur zu, was ein Test prüft) | `src/hexagon/services/workspace/fakes_log_test.go` · `qualifiziert die Namen wie ein Handler des Originals` | ja — Mutanten `g09`, `g10` (`out-mg/res.txt`) | Kommentar des Fakes sagt mehr zu als sein Test |
| F-8 | LOW | F-4 ist nicht vollständig nachgezogen: die Zeile *Generierbarkeit je Klasse* (3) in §3 nennt weiter „20 Stellen“ für die Fehler-Injektion; Z07, Z09 und der Bericht sagen 19 (S17 steht in Z36 und Z37). | `AGENTS.md` §3.9 (jede Stelle, die die Aussage trägt) | Plan · `generierbar: 20 Stellen mal die Menge der Variablen` | nein | Tabelle und Bericht im selben Plan auseinander |
| F-9 | LOW | Der Quelltext-Test fängt „Sperren und Goroutinen“ nur in drei Schreibweisen, `SPEC-129` sagt zu, Sperren und Goroutinen im Paket fehlten. Nach dem Quelltext gelesen, nicht gemessen (die Messung `mz4` ist ungelaufen): ein Import mit Raw-String-Pfad (`` `sync` ``, `strings.Trim(…, "\"")` lässt die Backticks stehen), eine Goroutine über `time.AfterFunc` oder `context.AfterFunc` und Code in `var _ = f()` (Ersatz für `init`) sind keine Befunde. `SPEC-127` nennt die Goroutine über die Standardbibliothek ausdrücklich als *Ungeprüft*, `SPEC-129` nicht. | `AGENTS.md` §3.11, §3.15 | `src/hexagon/services/workspace/zustand_test.go` · `strings.Trim(imp.Path.Value, `; `spec/spezifikation.md` · `dass Paketvariablen, Sperren, init-Funktionen und Goroutinen im Paket fehlen` | ja (Messung offen) | Quelltext-Test kennt nicht jede Schreibweise der Zusage |
| I-6 | INFO | Derselbe Satz („ein Zustand, den eine Änderung einführt, meldet der Detektor, sobald ein Test ihn aus zwei Goroutinen berührt“) steht unberichtigt in *Ungeprüft* von `SPEC-124` (Nr. 10) und `SPEC-127` (Nr. 18). Ob die Fakes dort dieselbe gemeinsame Sperre tragen, ist nicht gemessen. Adresse: Architect. | `AGENTS.md` §3.9 | `spec/spezifikation.md` · `meldet der Detektor, sobald ein Test ihn aus zwei Goroutinen berührt (Nr. 10)` | nein | – |
| I-7 | INFO | Betriebsregel dieser Runde: die zwei Fix-Commits (`fcda8de`, `9e61be6`) zeigen im Diff keinen Stil von `sed -i`/`perl -pi`/`cp` ins Repo (Dateien vollständig geschrieben, Plan zeilenweise geändert); ob die Auflagen eingehalten wurden, ist dem Diff nicht abzulesen, kein Verstoß erkennbar. Rohbytes `LC_ALL=C grep -P` auf NBSP, BOM, U+2000 bis U+206F, kombinierendes `cc 88`, Folge Backslash-`u`-vier-Hex in `zustand_test.go`, `leerraum_test.go`, `logfake_test.go`, `fakes_log_test.go`, `verwaltung.go`, Plan, Spec: 0 Treffer; `leerraum_test.go` baut die Werte aus Bytewerten. | `AGENTS.md` §3.18 | `src/hexagon/services/workspace/leerraum_test.go` · `leerraum := [][]byte{{0x20}, {0x09}, {0x0A}, {0x0D}}` | ja | – |
| I-8 | INFO | `TestPaketOhneZustand` bindet an die Zahl 7 der Dateien ohne `_test.go` (Schutz gegen leeres Lesen); jede legitime neue Quelldatei macht ihn rot. Kopplung bewusst, im Testtext genannt. | Maintainability | `zustand_test.go` · `erwartet 7` | ja | – |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `spec/spezifikation.md` SPEC-129 Nr. 6 a | geprüft, ohne Befund (Leerraum nach RFC 8259 stimmt mit Dienst, Probe und Test überein) |
| `src/hexagon/services/workspace/verwaltung.go` (`idGueltig`) | geprüft, ohne Befund |
| `src/hexagon/services/workspace/` Produktcode (`anlegen.go`, `reset.go`, `loeschen.go`, `manifest.go`, `protokoll.go`, `fehler.go`) | geprüft, ohne Befund am Verhalten (Probe 0 von 21 276; Mutationen des Leerraums an Tab, CR, Leerzeichen, LF, BOM, NBSP rot) |
| Nebenläufigkeits-Fänger Z46 bis Z55 (Schleife A) | geprüft, ohne Befund: 50 von 50 Prozessen und 50 von 50 Läufen rot für `n46a`, `n51`, `n48`, `n49`, `n50`, `n52`, `n53`, `n54`, `n55`; Normalstand 0 von 50 und 0 Läufe (Messung `mr5`; Zeilen `n57b/c/d` ungelaufen) |
| `make a-check` (Z65) | geprüft, ohne Befund (0 Befunde, Exit 0) |
| Größenmarke | geprüft, ohne Befund (`make groessenmarke`: 5424 von 10600) |
| `make test` | geprüft, ohne Befund (Exit 0) |
| `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/` | im Fix-Diff unverändert, ohne Befund |
| Kommentare (§3.7, §3.11) in den neuen Dateien | geprüft; Befund F-7; sonst ohne Konjunktiv über verworfene Alternativen, ohne Chronik |
| Fehler-Injektions-Fänger, Manifest-Verstöße | in dieser Runde nicht erneut mutiert (Produktcode und Tests unverändert gegenüber dem ersten Review) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 3 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Zustandsfänger deckt nur neue Felder und Paketvariablen · Fänger führt Auswahl statt Klasse der Zeichen und Längen · Kommentar des Fakes sagt mehr zu als sein Test · Tabelle und Bericht im selben Plan auseinander · Quelltext-Test kennt nicht jede Schreibweise der Zusage

## Verdikt

**Merge-blockierend:** ja, wegen F-5 und F-6 (beide MEDIUM; kein HIGH, kein Verhaltensfehler des Dienstes). F-1 ist für Paketvariablen, `init`, Goroutinen, `sync`-Import und Felder der `Verwaltung` gelöst (deterministisch rot), F-2 für die vier Zeichen und die fünf genannten Nicht-Zeichen, F-3 für die Ableitungen von `*slog.Logger`, F-4 bis auf F-8. Offen sind die Klasse „Schreibzugriff und eigener Typ“ (Entscheidung zwischen Fänger, etwa einem Test, der Zuweisungen an `v.`-Ziele im Quelltext zählt, und einem Negativ mit Grund und Folge im Plan und in `SPEC-129`; der Wortlaut des Negativs liegt beim Architect) und die Klasse der Zeichen (Schleife über die Menge oder Prüfgrenze mit Grund und Folge).

**Übergabe:** F-5, F-6, F-7, F-8, F-9 an den Implementer; der Satz in *Ungeprüft* zu F-5 und I-6 zusätzlich an den Architect. Nicht gemessen und im Verifier-Lauf nachzuholen: `make gates`, `make docs-check`, `make lint`, `make test-race`, die Zeilen `n57b/c/d` und die Wiederholung `mz4`/`mr6`.
