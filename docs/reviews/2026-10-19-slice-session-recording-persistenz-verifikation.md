# Verifikations-Report: slice-session-recording-persistenz — 2026-10-19

**Art:** Verifikation (Modul 11) — gegen DoD, Closure-Trigger, Spezifikation und Entscheidungen. Nicht Review, nicht Validierung.

**Gegenstand:** [slice-session-recording-persistenz](../plan/planning/done/slice-session-recording-persistenz.md), Diff `7b7e877` (Move nach `in-progress/`) bis `161b4b3` (HEAD).
Bezug: [LH-FA-02](../../spec/lastenheft.md), [LH-FA-03](../../spec/lastenheft.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md).
Vorgänger: [Review-Report](2026-10-19-slice-session-recording-persistenz-review.md) (2 HIGH, 2 MEDIUM, 2 LOW, 4 INFO).

**Rolle:** Verifier · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

## Urteil: BESTANDEN MIT AUFLAGEN

Der Code erfüllt `SPEC-105` Nr. 1–13, `SPEC-001`/`SPEC-004` und `spec/architecture.md` §2.3; alle Gates sind grün, die Gleichzeitigkeit hält.
Die Befunde F-1 bis F-4, F-6 des Reviews sind behoben und halten den Mutationen stand (F-2 und F-3 selbst nachgefahren).
Kein HIGH.
Offen sind drei MEDIUM-Lücken der Fänger (Prüfreihenfolge, Jahresbereich bei Zeitzone und `finished_at`, Löschen fremder Exchanges), die der Plan als gefangen führt oder deren Zusage `SPEC-105` ausdrücklich trägt.
Sie gehören in einen weiteren Implementer-Durchgang vor der Closure; sie sind keine Fehler des Produktcodes.

**Anzahl:** HIGH 0 · MEDIUM 3 · LOW 5 · INFO 5.

## Läufe (Verifier selbst)

| Lauf | Ergebnis |
|---|---|
| `make gates` (HEAD) | Exit 0: a-check 0 Befunde (drei bekannte Dateien in `tools/gegenproben/` ohne Schicht), d-check, baseline-verify, build, test (`gofmt`, `go vet`, Tests), lint, schema-check (alle Fälle `ok`, darunter „Session loeschen nimmt ihre Cursor mit“), migration-check |
| `rm -f harness/gegenproben.stempel; make gegenproben` | Exit 0, „24 Fälle, 0 rot“; Stempel danach unverändert zu HEAD, am Ende `git checkout` (Baum sauber) |
| `go test -count=200 -run 'RecordingVertrag\|SessionVertrag' ./src/adapters/driven/sqlite/` (golang-Container laut Dockerfile-Digest, `--user`, Netz, Wegwerf-Klon) | `ok`, 598 s, kein Flackern |
| Zusatz-Sonde im Klon: 30 Runden × 6 gleichzeitige `Append` + `SetStatus` + `Delete` + `FindOpen` + `ListBySession` auf dieselbe Session, 5 Läufe | nur erwartete Fehler, keine verwaisten Exchanges, kein `SQLITE_BUSY` nach außen |
| Mutationen | 150 im Wegwerf-Klon (`git archive HEAD`), Arbeitsbaum unberührt |

## DoD-Punkte 1–4 und Closure-Trigger (§5)

| Punkt | Befund |
|---|---|
| 1 Modell, Ports, Fehler, Atomarität | erfüllt: `Session`/`Exchange` mit den Zuständen aus `SPEC-001`, beide Ports und sieben Fehler in `outbound/errors.go`; Transaktionsgrenze über mehrere Ports ausdrücklich nicht hier, jede Operation atomar (`inTx`, `BEGIN IMMEDIATE`, Rollback; Mutation „COMMIT trotz Fehler“ rot) |
| 2 SQLite-Repositories | erfüllt: Create, Status, `FindOpen`, append-only Append/Get/List, Delete samt Exchanges |
| 3 Vertragstest-Suite, Randformen aus §6 | erfüllt mit Lücken (V-1 bis V-5): Suite am Port, gegen SQLite gefahren; leer/fehlend/unbekannt, doppelt, Rollback, Gleichzeitigkeit getestet; Abbruch zwischen zwei Schritten einer Operation nicht |
| 4 `make gates` grün | erfüllt (siehe Läufe) |
| Closure-Trigger „DoD vollständig · Gates · Suite gegen SQLite · eine Mutation je Zusage rot“ | Gates und Suite: ja. Mutation je Zusage: ja, außer der Prüfreihenfolge von Nr. 1 und Nr. 8 (V-1). DoD vollständig: nein, erwartungsgemäß offen — Häkchen, Closure-Notiz mit Lerneintrag, Register, Ausgänge §6 |

## Plan-vs-Code-Diff (`7b7e877..HEAD`)

| Plan §3 | Code | Abweichung |
|---|---|---|
| `src/hexagon/model/` neu | `session.go`, `exchange.go`, `model_test.go` | keine |
| `ports/outbound/`, `outboundtest/` neu | `errors.go`, beide Ports, `outboundtest/vertrag.go` | keine |
| `src/adapters/driven/sqlite/` update | `sessions.go`, `recording.go`, `tx.go`, `repositories_test.go` | keine; `lesend()` entfernt (F-6) |
| `spec/spezifikation.md` update | `SPEC-105` vor dem Code (`7336fec`), danach geschärft (`fa5f18d`) | keine |
| nicht im Plan | `harness/gegenproben.stempel`, Roadmap, `slice-baseline-persistenz` (Adresse nachgezogen), Review-Report (Link-Form von F-9) | harmlos (I-4) |

Weder `.golangci.yml`, `.a-check.yml`, `Dockerfile` noch `go.mod` wurden berührt; kein `//nolint` im Diff (`AGENTS.md` §3.2, §3.6 gewahrt).
Die Marke aus §4 (700 Zeilen) ist überschritten und benannt; der genannte Messwert ist veraltet (V-8).

## SPEC-105 Nr. 1–13 einzeln gegen den Code

| Nr. | Zusage | Code | Urteil |
|---|---|---|---|
| 1 | Create: open, Validierung, dann Kennung, dann Service | `Validate`, dann `INSERT`; Primärschlüssel vor Fremdschlüssel (Sonde: doppelt + fremder Service → `ErrBereitsVorhanden`) | erfüllt; Reihenfolge ungefangen (V-1) |
| 2 | Get unbekannt → `ErrNichtGefunden` | ja | erfüllt |
| 3 | nur `open` → `complete`/`incomplete`, sonst `ErrUngueltigerUebergang`, unbekannt `ErrNichtGefunden` | `KannWechselnNach` + Prüfung in der Transaktion | erfüllt |
| 4 | List nach `created_at` (Zeit, nicht Text), dann Kennung; leer; leerer Service → `ErrUngueltig` | neun feste Nachkommastellen, UTC, `ORDER BY created_at, id` | erfüllt; F-2 behoben, 11 Mutationen rot |
| 5 | FindOpen über alle Services, gleiche Ordnung | ja | erfüllt |
| 6 | Delete samt Exchanges und Cursors, `ErrReferenziert`, `ErrNichtGefunden` | Exchanges explizit, Cursors per `CASCADE` (Schema-Gate), Rollback bei Fremdschlüssel | erfüllt; Isolation gegen fremde Exchanges ungefangen (V-3) |
| 7 | Append nur an offene, bekannte Session | ja | erfüllt; `incomplete` ungefangen (V-4) |
| 8 | Sequenz lückenlos, gleichzeitig genau einer; Prüfreihenfolge | `BEGIN IMMEDIATE`, `max(sequence_no)+1`; Reihenfolge Gültigkeit → Session → Kennung → Sequenz | erfüllt; Reihenfolge ungefangen (V-1) |
| 9 | Kennung vorhanden → `ErrBereitsVorhanden`, unverändert; keine ändernde Operation | Vorabprüfung + Primärschlüssel; Port ohne Update/Delete | erfüllt |
| 10 | Gültigkeit (Pflichtfelder, Jahre 1–9999, `finished_at`, Status 0/100–599, UTF-8, Aufzählungen, Body/Blob, `match_key`, discarded, beide `unsafe`) | `Validate` in fünf Gruppen, `zeitGueltig` über UTC | erfüllt; einzelne Regeln schwach gefangen (V-2, V-5) |
| 11 | Get/ListBySession: unbekannt, nach `sequence_no`, leer | ja | erfüllt |
| 12 | Header-Randformen, Status 0 = keiner | Schreib- und Leseseite normalisieren | erfüllt |
| 13 | abgebrochener Kontext beendet jede Operation ohne Änderung | alle neun Operationen getestet; ganze Operation ohne Abbruch → rot | erfüllt; Vorrang gegenüber `ErrUngueltig` offen (I-2) |

`SPEC-001`/`SPEC-004`: Zustände, Blob-Kopplung, `match_key` nur bei `failed`, beide `unsafe`-Regeln und die Feldnamen entsprechen dem Schema; das Modell lässt keinen Zustand zu, den die CHECK-Constraints ablehnen.
`spec/architecture.md` §2.3: beide Ports tragen die dort genannten Verantwortungen; die fachliche Transaktionsgrenze über mehrere Ports ist laut Plan und Architektur der Folge-Slice-Gegenstand.
[ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): nur `modernc.org/sqlite` (`CGO_ENABLED=0`), kein neuer Treiber.
[ADR-0003](../plan/adr/0003-lint-profil-golangci.md): Lint-Profil unverändert, `make lint` 0 Issues.

## Zusage · Fänger gegen die Tests, Mutationen je Zusage

„rot“ nennt, dass der Test aus dem benannten Grund rot wurde (Fehlermeldung des Falls geprüft, nicht Aufbau); „grün“ ist eine überlebende Mutation.
Alle Mutationen im Wegwerf-Klon.

| Plan-Zeile | Rot gesehen | Grün geblieben |
|---|---|---|
| Nr. 1 | Service-Zweig entfernt · Kennungszweig entfernt · `Validate` in `Create` entfernt | **Service vor Kennung geprüft** (V-1) |
| Nr. 2/11 | Get ohne `ErrNoRows`-Abbildung (Session, Exchange) · `ListBySession` unbekannt → leer | — |
| Nr. 3 | Übergang immer erlaubt · open→open erlaubt · aus geschlossen erlaubt · Prüfung im Adapter entfernt · unbekannte Kennung ohne Fehler · schreibt immer `complete` | — |
| Nr. 4/5 | ohne `ORDER BY` · `DESC` · ohne Kennungsordnung · Kennung vor Zeit · variable Nachkommastellen · drei Nachkommastellen · ohne `.UTC()` · Filter umgekehrt · ohne Servicefilter · leerer Service ohne Fehler · `nil` statt leer | — |
| Nr. 6 | Exchanges nicht gelöscht · `ErrReferenziert` entfernt · Unbekannt ohne Fehler · löscht nichts · COMMIT trotz Fehler (kein Rollback) · kein Rollback (Verbindung bleibt in Transaktion) | **löscht die Exchanges aller Sessions** (V-3) |
| Nr. 7 | Zustandsprüfung entfernt · Session-Prüfung entfernt · nur `complete` zählt als offen | **nur `complete` zählt als geschlossen** (V-4) |
| Nr. 8 | Prüfung entfernt · nur Lücke · nur Duplikat · `BEGIN` statt `BEGIN IMMEDIATE` · Sequenz global statt je Session | `count(*)` statt `max` (äquivalent, I-3) |
| Nr. 9 | Kennungsprüfung entfernt · mit `INSERT OR REPLACE` · Prüfung nur je Session | `INSERT OR REPLACE` allein (äquivalent, vom Plan so deklariert) · **Sequenz vor Kennung** · **Kennung vor Session** (V-1) |
| Nr. 10 | 57 Mutationen einzelner Regeln von `Validate` (entfernt oder abgeschwächt) | 9 überlebende (V-2, V-5) |
| Nr. 11 | Ordnung nach Kennung · `nil` statt leer | `ORDER BY` ganz entfernt (äquivalent, I-3) |
| Nr. 12 | Schreib-Normalisierung entfernt · Response-`nil` als `{}` · leer als `nil` gelesen · leer als NULL · Request-Header als `null` | Status 0 gespeichert (äquivalent, I-3) |
| Nr. 13 | je Operation ganz ohne Abbruch (Create, SetStatus, Delete, Append, Get, FindOpen) · Lesepfade `List`, `ListBySession`, `Get` ohne Kontext | Conn/`BEGIN`/`COMMIT`/einzelnes Statement ohne Kontext (redundant, V-7) |
| atomar | COMMIT trotz Fehler · Rollback entfernt | — |

## Findings

| ID | Stufe | Befund | Quelle | Pfad | Verifizierbar |
|---|---|---|---|---|---|
| V-1 | MEDIUM | Die Prüfreihenfolge, die `SPEC-105` Nr. 1 (Gültigkeit → Kennung → Service) und Nr. 8 (Gültigkeit → Session → Kennung → Sequenz) festlegt und die der Plan in der Zeile Nr. 1 „in dieser Reihenfolge“ zusagt, hat keinen Fänger. Mutationen grün: Service vor Kennung geprüft; Sequenz vor Kennung (`vorhanden > 0 && SequenceNo == letzte+1`); Kennung vor Session. Die Tests bauen nie zwei Verstöße zugleich (doppelte Session mit fremdem Service; doppelte Exchange-Kennung mit falscher Sequenz; doppelte Kennung an unbekannter Session). Die Reihenfolge wurde im Review (F-7) als Randform an den Architect gegeben und in `fa5f18d` entschieden, aber nicht mit einem Test belegt. | `AGENTS.md` §3.15, §3.10, §3.12 · `SPEC-105` Nr. 1/8 | `src/hexagon/ports/outbound/outboundtest/vertrag.go` · `sessionFall2`, `recordingFall5` | ja, Mutation `if vorhanden > 0 {` → `if vorhanden > 0 && e.SequenceNo == l0+1 {` |
| V-2 | MEDIUM | Der Jahresbereich (F-3) ist nur an einer Seite gefangen. Grün: `zeitGueltig` ohne `.UTC()` (eine Zeit `9999-12-31T23:00:00-05:00` hat Ortsjahr 9999, UTC-Jahr 10000 und würde als `10000-…` geschrieben und unlesbar); `finished_at` mit Jahr 10000 nicht abgelehnt (`!zeitGueltig(e.FinishedAt)` entfernt). Gelesen am Code: richtig; gefangen: nein. Die Plan-Zeile Nr. 10 führt „Jahre 1 bis 9999“ und „je eine Regel entfernt oder abgeschwächt“. | `AGENTS.md` §3.15 · `SPEC-105` Vorspann, Nr. 10 | `src/hexagon/model/session.go` · `zeitGueltig`; `src/hexagon/model/exchange.go` · `pruefeWerte` | ja, beide Mutationen |
| V-3 | MEDIUM | `Delete` begrenzt das Löschen der Exchanges nicht auf die eigene Session, soweit ein Test es sieht: Die Mutation `DELETE FROM exchanges WHERE ? != ''` (löscht die Exchanges aller Sessions) bleibt grün. `sessionFall6` prüft nach dem erfolgreichen Löschen von `s` nicht, dass `e2` der Session `q` noch lebt (`q` wurde zuvor abgelehnt, ihr Exchange wird danach nicht mehr gelesen). Die Zusage „samt Exchanges“ meint die eigenen. | `AGENTS.md` §3.15 · `SPEC-105` Nr. 6 | `src/hexagon/ports/outbound/outboundtest/vertrag.go` · `sessionFall6` | ja |
| V-4 | LOW | `Append` an eine `incomplete`-Session ist ungetestet: Mutation „nur `complete` gilt als geschlossen“ bleibt grün. Nr. 7 und Nr. 3 sprechen von jedem Zustand außer `open`. | `SPEC-105` Nr. 7 | `src/hexagon/ports/outbound/outboundtest/vertrag.go` · `recordingFall2` | ja |
| V-5 | LOW | Schwache Randfälle von `Validate` (Mutation grün): Response-Body-Aufzählung, Request-Redaction-Aufzählung; Request-Body `present` ohne Blob; Response-Blob ohne `present`; `unsafe` Response ohne verworfenen Body (Test setzt zugleich `replay_equivalence=byte`, trifft also die zweite Bedingung); Status `100` und `599` als gültige Grenzwerte (getestet sind nur `99` und `600`). | `AGENTS.md` §3.10 · `SPEC-105` Nr. 10 | `src/hexagon/model/model_test.go` · `TestExchangeValidate` | ja, je eine Mutation (`V-E8b`, `V-E8c`, `V-E9b`, `V-E9d`, `V-E13b`, `V-E6d`, `V-E6e`) |
| V-6 | LOW | `RunSessionVertrag` führt die Fälle „List ordnet nach Zeit, nicht nach Text“, „Zeiten außerhalb der Jahre 1 bis 9999 sind ungültig“ und „abgebrochener Kontext beendet jede Operation ohne Änderung“ doppelt auf (Kopierfehler im Fallverzeichnis, die Läufe heißen `…#01`). Kein Gutfall wird verdeckt, aber drei Fälle laufen zweimal, und jede Zählung der Fälle (Plan, Bericht) stimmt nicht. | Wartbarkeit | `src/hexagon/ports/outbound/outboundtest/vertrag.go` · `RunSessionVertrag` | ja, `go test -v -run SessionVertrag` nennt `#01` |
| V-7 | LOW | Der Kommentar von `inTx` sagt zu: „ein abgebrochener Kontext beendet jeden Schritt mit seinem Fehler“. Ein Test fängt das nur für den ersten Schritt: `Conn`, `BEGIN`, `COMMIT` und jedes einzelne Statement ohne Kontext bleiben grün, weil der erste Schritt schon scheitert. Der Plan sagt ehrlich „keine Mutation“, der Kommentar sagt mehr zu. Zugleich bleibt die Randform „Abbruch mitten in der Transaktion“ aus §6 (Abbruch zwischen zwei Schritten) ungetestet; getestet ist nur der Abbruch durch einen Fehler (Rollback). | `AGENTS.md` §3.11, §3.7 | `src/adapters/driven/sqlite/tx.go` · `inTx` | ja |
| V-8 | LOW | Plan §4 nennt „1248 Zeilen“; gemessen über die Pfade aus §3 (`src/hexagon/model`, `src/hexagon/ports`, `src/adapters/driven/sqlite`, `spec/spezifikation.md`) mit `git diff --numstat 7b7e877..HEAD` sind es 1399, mit dem Plan selbst 1426. Der Plan wird mit jeder Änderung nachgezogen (`AGENTS.md` §3.9), die Zahl blieb stehen. Die Entscheidung über das Weiterführen trägt weiterhin der Implementer (Review F-8), nicht der Planner. | `AGENTS.md` §3.14, §3.9 | `docs/plan/planning/done/slice-session-recording-persistenz.md` · „Marke überschritten“ | ja |
| I-1 | INFO | Die Nullzeit von Go (`0001-01-01T00:00:00Z`) gilt als „nicht gesetzt“ und wird abgelehnt, obwohl `SPEC-105` die Jahre „1 bis 9999“ nennt (Sonde: `0001-01-01T00:00:01Z` läuft hin und zurück). Die Obergrenze 9999 mit Nanosekunden (`23:59:59.999999999`) ist belegt. Randform für den Architect. | `AGENTS.md` §3.12 | `src/hexagon/model/session.go` · `zeitGueltig` | ja, Sonde |
| I-2 | INFO | Ein ungültiges Argument mit abgebrochenem Kontext liefert `ErrUngueltig`, nicht den Fehler des Kontexts (`Create(ctx, Session{})`, `List(ctx, "")`). `SPEC-105` Nr. 13 sagt „jede Operation“; der Vorrang ist nicht entschieden. Randform für den Architect. | `AGENTS.md` §3.12 | `src/adapters/driven/sqlite/sessions.go` · `Create`, `List` | ja, Sonde |
| I-3 | INFO | Äquivalente Mutationen, die keine Zusage brechen: `count(*)` statt `max(sequence_no)`; `ORDER BY sequence_no` in `ListBySession` entfernt (SQLite liefert in Einfügereihenfolge, und die ist die Sequenz); Status `0` als Zahl statt `NULL` gespeichert (gleich gelesen); `INSERT OR REPLACE` allein (vom Plan so deklariert). Die Ordnung von `ListBySession` hängt damit nur an der Einfügereihenfolge und am Fänger „nach Kennung“. | `AGENTS.md` §3.15 | `src/adapters/driven/sqlite/recording.go` | ja |
| I-4 | INFO | Der Implementer hat im Fix-Commit `161b4b3` den Review-Report angefasst (nur die Link-Form der Zitierung in F-9, damit das Doku-Gate grün bleibt). Inhalt und Findings blieben gleich; ein Report des Reviewers ist trotzdem sein Werkstück. | Rollentrennung | `docs/reviews/2026-10-19-slice-session-recording-persistenz-review.md` | ja, `git show 161b4b3` |
| I-5 | INFO | `-count=200` braucht 598 s (rund 3 s je Durchlauf, vor allem Dateisystem); `-race` ist mit `CGO_ENABLED=0` nicht verfügbar. Die Gleichzeitigkeit stützt sich damit auf Ausgang und Sonden, nicht auf den Race-Detektor. | [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) | — | — |

## Review-Befunde: Nachweis der Behebung

| Review | Zustand |
|---|---|
| F-1 (kein Kontext-Test) | behoben: beide Fälle, alle neun Operationen; ganze Operation ohne Abbruch rot (Einschränkung V-7) |
| F-2 (Zeitordnung) | behoben, selbst nachgefahren: neun Stellen, UTC; Fall mit `.1`, `.12`, ganzer Sekunde und Zeitzone; vier Mutationen rot |
| F-3 (Jahresbereich) | behoben im Code; gefangen für `started_at`/`created_at`, nicht für Zone und `finished_at` (V-2) |
| F-4 (Fänger) | behoben für `ListBySession`-Ordnung, Request-`unsafe`, UTC, Nil-Normalisierung; Cursors: Schema-Gate, nicht Go-Test (Plan nennt das ehrlich) |
| F-5 | behoben: Plan nennt `INSERT OR REPLACE` als äquivalent |
| F-6 | behoben: `lesend()` entfernt |
| F-7 (Randformen) | behoben in `SPEC-105` vor dem Fix; zwei Reste als I-1, I-2 |
| F-8 (Marke) | offen, V-8; Entscheid steht beim Planner |

## Offen und erwartet (kein Befund)

Häkchen der DoD, Closure-Notiz mit Steering-Loop-Lerneintrag, Beobachtungs-Register (`BEO-REPO/randform-erst-im-review-entschieden`, `…/nebenlaeufiger-lauf-nur-im-glueck-gruen`), Ausgänge der Risiken in §6 (der zweite steht „weiter offen“, die anderen „entfallen“), die drei Paarungen.

## Übergabe

Der Implementer schließt V-1 bis V-3 (Fänger nachziehen, kein Produktcode), nimmt V-4 bis V-8 mit; I-1 und I-2 gehen an den Architect (Randformen, `AGENTS.md` §3.12), V-8 und die Marke an den Planner.
Danach genügt eine erneute Verifikation der betroffenen Zeilen; der Produktcode ändert sich dafür voraussichtlich nicht.
