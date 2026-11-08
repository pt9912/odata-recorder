# Review-Report: slice-bootstrap-m3 (Teil A) — 2026-11-08

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-bootstrap-m3`, HEAD `8e0c1382`, Basis `ada8bd88`; 18 Dateien (1252 Einfügungen, 93 Löschungen): `src/bootstrap/` (Produktcode `konfig_basis.go`, `konfig.go`, `konfig_werte.go`, `wurzel.go`; Tests `m3_konfig_test.go`, `m3_wurzel_test.go`, `konfig_test.go` u. a.), `tests/integration/m3_test.go`, `tests/e2e/stufe_prozess_test.go`, `harness/katalog-ausnahmen.txt`, `docs/user/benutzerhandbuch.md`, `SPEC-201` Nr. 3 e. Plan §3 (Tabelle A01 bis A16, Absatz *Mutationen*), §6 (Raster, Zusatztabelle).

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-08

**Eingangs-Kontext:**

- Plan von `slice-bootstrap-m3` (§1, §3, §6), `SPEC-110` Nr. 1 bis 14, `SPEC-139` Nr. 11, `SPEC-136` Nr. 8, `SPEC-201` Nr. 3 e
- `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.18
- `LH-FA-12`, `LH-FA-06`, `LH-FA-11`, `LH-QA-04`
- ADR: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (nur genannt)

**Nicht geprüft (Rolle des Verifiers oder Planners):** DoD, `make gates`, Größenmarke. Als bekannt gemeldet und hier nur bewertet: Handbuch-Kapitel und Beispiel zum Modus `mock`, Länge der abgeleiteten `PUBLIC_BASE_URL`, `mutationslauf.sh` an `src/bootstrap`, flackernder `urlparser/TestUeberlangeEingabe`.

## Eigene Mutationen

Messender ist der Reviewer-Agent. Container aus dem Image `odata-recorder:race`, ohne Netz, `--memory=2g`, `--cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`; der Baum liegt als Wegwerfkopie im Container, je Mutation ein `sed`, `go test -count=1 -run <Test> ./src/bootstrap` (ein Prozess), danach Rücksetzen. Last des Hosts zu Beginn: 5,8. Einzelläufe, keine Ratenmessung (alle Fänger sind deterministisch).

| Nr. | Zusage | Mutation | Ergebnis |
|---|---|---|---|
| M1 | A04, Klasse *Prozent-Kodierung* (zweite Hex-Ziffer) | `hexZiffer(seg[i+2])` wird `hexZiffer(seg[i+1])` | **grün** (`TestKonfigPublicBaseURL*`) |
| M2 | A04, Form | Port bis 6 Ziffern statt 5 | rot `TestKonfigPublicBaseURLForm` |
| M3 | A04, Form | Dekodiertes Segment `..` nicht abgelehnt | rot `TestKonfigPublicBaseURLForm` |
| M4 | A03 | Default `STRICT_COVERAGE` `false` | rot `TestKonfigStrictCoverage` |
| M5 | A04, Meldung (`SPEC-110` Nr. 14 d, „nennt nie den Wert“) | Meldung `Form ungültig: ` plus Wert | **grün** (`-run Konfig`, alle Tests des Paketteils) |
| M6 | A12 | Hochfahren öffnet Sessions auch im Modus `mock` | rot, aber als `panic: nil pointer dereference` in `TestM3Lebenszyklus` (allein gefahren) und in `TestM3WahlJeModus` |
| M7 | A07 | `h.verarbeiten = h.record.Record` im Modus `mock` | rot, aber als `panic: nil pointer dereference` in `TestM3WahlJeModus` |
| M8 | A08/A11 | Entities des Mock aus einem zweiten `db.Entities()` | rot `TestM3KonfigDesMock` |
| M9 | A11 | `Blobs` des Snapshot-Dienstes auf `store` statt Wächter | rot `TestM3KonfigDesMock` |
| M10 | A11 | `Blobs` des Workspace-Dienstes auf `store` statt Wächter | rot `TestM3KonfigDesMock` |

## Findings

| kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|
| MEDIUM | AGENTS.md §3.17, §3.15 | `src/bootstrap/m3_konfig_test.go` · `"http://h/%#1", "http://[a#]"}` (Generator in `TestKonfigPublicBaseURLZeichen`) | Der Generator variiert in der Prozent-Kodierung nur die erste Hex-Ziffer (`%#1`), die zweite bleibt `1`; die Liste der ungültigen Formen enthält `%`, `%4`, `%zz`, aber keine Form mit gültiger erster und ungültiger zweiter Ziffer (`%4z`). Die Mutation M1 (zweite Ziffer ungeprüft) bleibt grün; Elemente der Klasse *jede Stelle der Prozent-Kodierung* sind nicht gefangen. | ja | Generator deckt eine Stelle einer Mehrstellen-Form |
| MEDIUM | SPEC-110 Nr. 14 d, LH-FA-14, AGENTS.md §3.15 | `src/bootstrap/konfig_test.go` · `func TestKonfigMeldung` | „Die Meldung nennt nie den Wert“ hat für `PUBLIC_BASE_URL` und `STRICT_COVERAGE` keinen Fänger: `TestKonfigMeldung` führt beide Schlüssel nicht, die Tests der Form prüfen nur `HasPrefix`. M5 (Wert in der Meldung) bleibt grün. `PUBLIC_BASE_URL` kann Zugangsdaten tragen (`http://u:pw@h`), die Meldung geht auf stderr, eine diagnostische Senke. Der Code leckt heute nicht. | ja | Zusage ohne Fänger, Geheimnis-Senke |
| MEDIUM | AGENTS.md §3.15, §3.10 (Hinweis `panic`) | `src/bootstrap/m3_wurzel_test.go` · `TestM3Lebenszyklus fängt A12: im Modus mock entsteht beim Hochfahren keine Session` | Die Zusage „keine Session entsteht“ hat den Fänger „Zahl der Sessions der Ablage vor und nach `Hochfahren` gleich null“ nur im Plan (§3, Zeile A12); der Test zählt keine Sessions. M6 wird nur durch die Nil-Dereferenzierung des fehlenden Session-Dienstes rot (`panic`), nicht aus dem benannten Grund; ein Pfad, der Sessions ohne den Dienst anlegt, bliebe grün. | ja | Fänger nur per Panik, Plan nennt Fänger, der nicht existiert |
| LOW | AGENTS.md §3.7, §3.11 | `src/bootstrap/konfig_test.go` · `// bei mock und hybrid wird kein DATA_DIR angelegt` | Der Kommentar nennt `mock`; der Test fährt nur `MODE=hybrid`. Im Modus `mock` wird `DATA_DIR` jetzt angelegt, der Satz stimmt für `mock` nicht mehr. | nein | Kommentar sagt mehr als der Test / veraltet |
| LOW | AGENTS.md §3.9, §3.11 | `docs/plan/planning/in-progress/slice-bootstrap-m3.md` · `Haken-Test: Schleife über die Dienste und die sechs Felder mit Zählprüfung (sechs je Dienst)` | Der Test zählt fünf Typfelder (`len(typen) != 5`), dazu `Version` und `NurLesend` einzeln; Plan A09 nennt „sechs“. | nein | Plan-Zeile und Test weichen ab |
| LOW | AGENTS.md §3.15 | `tests/integration/m3_test.go` · `d.ref.srv.Close()` und `nachDemEnde` (A13) | Die Prüfung „keine Verbindung zum Referenzdienst“ kann nicht rot werden: der Mock-Lauf bekommt keine `UPSTREAM_URL` (`konfig(…, "mock", dir, "")`), der Dienst ist geschlossen. Die Mutation „der Mock ruft den Upstream“ der Plan-Zeile A13 ist ohne Upstream-Adresse nicht ausführbar; die Zusage trägt `verboten` in `TestM3WahlJeModus`. | nein | Fänger ohne mögliche Wirkung |
| LOW | Maintainability | `src/bootstrap/wurzel_test.go` · `_ = os.MkdirAll(km.DataDir, 0o700)` | Der Fehler von `MkdirAll` wird verworfen; ein Fehlschlag zeigt sich erst als Fehler von `lauf`. | nein | Fehler im Test verworfen |
| INFO | AGENTS.md §3.15 | `src/bootstrap/wurzel_test.go` · `"StrictCoverage":       len(mh.mock) == 1 && mh.mock[0].StrictCoverage` | `TestWerteErreichenVerbraucher` führt `StrictCoverage: true` gleich dem Default; die Mutation „fest `true`“ fängt allein `TestM3KonfigDesMock` (beide Werte). Kein Befund, nur der Hinweis, dass die Zeile für sich keinen Fänger trägt. | ja | Fixture gleicht dem Default |
| INFO | Plan §3 *Ungeprüft*, §6 | `src/bootstrap/wurzel.go` · `Coverage: db.Coverage()` | Als akzeptiertes Negativ mit Grund und Folge geführt (Typ statt Identität); bewertet, Grund trägt (zustandsloses Repository). | nein | akzeptiertes Negativ |
| INFO | SPEC-110 Nr. 14 c, SPEC-200 Nr. 6 d | Plan §6 (Beispiel, abgeleitete Länge) | Die zwei als bekannt gemeldeten offenen Punkte tragen Adresse Architect und Ausgang *offen*; ihr Fehlen im Code ist gemäß Plan kein Befund dieses Reviews. Das Handbuch trägt die acht Codes, die Ausnahmen sind um acht Zeilen kürzer, `SPEC-201` Nr. 3 e ist nachgezogen. | nein | Bekanntes, Adresse benannt |

Keine HIGH. Der Konflikt-Pfad (HIGH mit Rollen-Widerspruch) entfällt.

## Zeilen der Tabelle (A01 bis A16) und Zellen des Rasters

A01 bis A06, A08 bis A11, A13 bis A15: gelesen und mit den Mutationen M2 bis M4, M8 bis M10 sowie den Mutationen des Plans (nicht neu gefahren) abgeglichen; Fänger sind Schleifen über die Mengen (Modus × Wert × Quelle, 256 Bytes × 5 Stellen, 9 Hosts × 3 Ports × 5 Modi), Fakes halten Argumente fest und zählen Aufrufe. Befunde: M1, M5 (oben). A07 und A12: rot nur per Panik (oben). A16: `make katalog-check` ist Fänger der Wahl; die acht Zeilen sind aus `harness/katalog-ausnahmen.txt` entfernt, das akzeptierte Negativ (Erreichbarkeit in den Ausnahmen) trägt Grund und Folge. Raster: jede Zelle nennt eine Nummer; die Zeile *Request am Listener im Modus `mock`* (`ODR-E5003`, `ODR-E7001`, `ODR-E5004`, `ODR-E6009`) hat in `m3_test.go` keinen Fall außer `ODR-E6008`; die Fänger liegen beim Dienst (Geber), die Verdrahtung ist modusfrei. Kein Befund.

## Negativbefund (geprüft, ohne Befund)

- `src/bootstrap/konfig_basis.go`, `konfig.go`, `konfig_werte.go`: Ableitung, Länge (`>` an 2048, Root `/`, längster Root), Pflicht je Modus, ein Grund je Schlüssel, `none`-freie Meldungen: geprüft, ohne Befund außer M1/M5 (Testlücken, nicht Code).
- `src/bootstrap/wurzel.go`: ein Wächter, eine Ablage, Satz der Version 4 mit `Schreiber`, `Eingabe`, `WithFilter(NewFilterLeser())`; kein Upstream, keine Session im Modus `mock`; `PublicBaseURL` leer ⇒ `ErrUngueltig` vor dem Öffnen: geprüft, ohne Befund.
- `src/bootstrap/lebenszyklus.go` (unverändert): Sessions und `CloseAll` hängen am Modus `record`: geprüft, ohne Befund.
- `tests/integration/m3_test.go`, `tests/e2e/stufe_prozess_test.go`: geprüft, nur LOW oben.
- `harness/katalog-ausnahmen.txt`, `docs/user/benutzerhandbuch.md`, `harness/vertraege/SPEC-201-handbuch-katalog.md`: Codes, Zeilenzahl und Wortlaut stimmen mit §3 A16 überein: geprüft, ohne Befund.
- Suppression (`//nolint`, §3.2) und Gate-Lockerung (§3.6): keine im Diff.
- Gegenlesen von Rohbytes (§3.18): Tests bilden Werte aus `\x..`-Literalen in interpretierten Strings; geprüft in `m3_konfig_test.go`, ohne Befund.

Zählung: HIGH 0, MEDIUM 3, LOW 3, INFO 3.
