# Review-Report: slice-session-recording-persistenz — 2026-10-19

**Review-Art:** Code — gegen Plan, Entscheidungen und Hard Rules.

**Gegenstand:** `slice-session-recording-persistenz`, Diff `7b7e877..c62502d` (Vertrag `SPEC-105` in `7336fec`, Umsetzung `35adc0d`, Plan-Nachzug `c62502d`).

**Skill:** `.harness/skills/reviewer.md` @ 424c252
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

**Eingangs-Kontext:**

- Plan `slice-session-recording-persistenz` (inkl. Tabelle *Zusage · Fänger*, `AGENTS.md` §3.15)
- [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md)
- [LH-FA-02](../../spec/lastenheft.md), [LH-FA-03](../../spec/lastenheft.md)
- `SPEC-105` (Nr. 1–13 einzeln gegen den Code), `SPEC-001`, `SPEC-004`, `spec/architecture.md` §2.3, `.a-check.yml`
- `AGENTS.md` (Hard Rules §3.2, §3.6, §3.7, §3.9–§3.15)

**Eigene Läufe:** `make test`, `make lint`, `make a-check` grün (a-check nennt nur die schon bekannten drei Dateien in `tools/gegenproben/` ohne Schicht). Explorative Tests und 22 Mutationen liefen in einer Kopie des Baums im Scratchpad; das Repo blieb unberührt (`git status` sauber, `make gegenproben` nicht gefahren, daher kein Stempel-Rücksetzen nötig).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | HIGH | Die Zeile Nr. 13 der Tabelle *Zusage · Fänger* nennt „beide Tests `abgebrochener Kontext ändert nichts`". Kein solcher Test existiert (`grep` nach „abgebrochen" und „Cancel" findet in den Tests nur `TestWalUmstellenEndetBeiKontextabbruch`, der `stelleWalSicher` prüft, keine Repository-Operation). Die Zusage `SPEC-105` Nr. 13 hat keinen Fänger, der Plan behauptet einen. | `AGENTS.md` §3.15, §3.10, §3.11 | `docs/plan/planning/done/slice-session-recording-persistenz.md` · „beide Tests `abgebrochener Kontext ändert nichts`" | ja, `grep -rn "abgebrochener Kontext" src/` bleibt leer; eine Mutation (z. B. `context.WithoutCancel` an `BEGIN IMMEDIATE`) wäre grün | Fänger im Plan benannt, im Code nicht vorhanden |
| F-2 | HIGH | `List` und `FindOpen` sortieren mit `ORDER BY created_at` über Text im Format RFC3339Nano, das nachgestellte Nullen kürzt; Zeichenvergleich ordnet dann falsch. Beobachtet: `10:00:00Z` landet hinter `10:00:00.5Z`, und `.12Z` vor `.1Z`. `SPEC-105` Nr. 4 und 5 verlangen Ordnung nach `created_at`; der Vertragstest nutzt eine feste Nanosekunden-Endung (`123456789`) und ganze Minuten und sieht den Fehler nicht. `FindOpen` bestimmt die Reihenfolge der Wiederaufnahme offener Sessions. | `SPEC-105` Nr. 4/5 · `LH-FA-02` | `src/adapters/driven/sqlite/sessions.go` · `"ORDER BY created_at, id"` zusammen mit `time.RFC3339Nano` in `zeitText` | ja, Test mit Sessions bei `…:00Z`, `…:00.1Z`, `…:00.12Z` (hier selbst gefahren: Reihenfolge `c e a b` statt `b e c a`) | Sortierung über variabel breiten Zeittext |
| F-3 | MEDIUM | `Session.Validate` und `Exchange.Validate` lassen Zeiten außerhalb des Jahresbereichs 0000–9999 zu. Eine Session mit Jahr 10000 wird geschrieben, dann scheitern `Get` und — für alle Sessions desselben Service — `List`/`FindOpen` am Parse von `created_at`. Ein einziger Schreibvorgang macht die Liste des Service unlesbar. | `SPEC-105` Nr. 1/4, Maintainability | `src/hexagon/model/session.go` · `case s.CreatedAt.IsZero():` und `src/adapters/driven/sqlite/sessions.go` · `func zeitAus` | ja, selbst gefahren: `Create` ok, danach `List` → `parsing time "10000-01-01T00:00:00Z"` | Schreibbarer Wert, der der Leser nicht liest |
| F-4 | MEDIUM | Mehrere Zusagen der Fänger-Tabelle halten einer Mutation nicht stand, obwohl der Plan sie als gefangen führt. Grün geblieben: (a) `ListBySession` ohne `ORDER BY sequence_no` (Nr. 11; die Tests hängen in Reihenfolge an), (b) in `Validate` die Bedingung „Request `unsafe` verlangt verworfenen Request-Body" entfernt (Nr. 10 „je eine Regel entfernt"), (c) `zeitText` ohne `.UTC()` (Vorspann von `SPEC-105`, Tests nutzen nur UTC-Eingaben), (d) `lesend` ohne `BEGIN` (Konsistenz von `ListBySession`, Plan „jede Operation ist atomar"). Zusage „Cursors" aus Nr. 6 hat gar keinen Test (Vertragstest legt keinen Cursor an). | `AGENTS.md` §3.10, §3.15 | `src/hexagon/ports/outbound/outboundtest/vertrag.go` · `recordingFall3`, `recordingFall6`, `sessionFall6`; `src/adapters/driven/sqlite/tx.go` · `"BEGIN"` | ja, je Mutation (Ergebnisse oben, jede grün) | Zusage ohne rot werdende Mutation |
| F-5 | LOW | Zwei Mutationen aus der Tabelle sind äquivalent und werden nie rot: `INSERT OR REPLACE` (Nr. 9) bleibt grün, weil die vorgeschaltete Kennungsprüfung `vorhanden > 0` den Fall abfängt; die „`nil`-Normalisierung" (Nr. 12) steht doppelt (Schreib- und Leseseite), jede allein entfernt bleibt grün. Der tatsächliche Fänger von Nr. 9 ist die Kennungsprüfung (ihre Entfernung wird rot). Die Tabelle nennt den falschen Mutationsort. | `AGENTS.md` §3.15, §3.11 | `docs/plan/planning/done/slice-session-recording-persistenz.md` · `Mutation: \`INSERT OR REPLACE\`` | ja | Mutation in Tabelle unwirksam |
| F-6 | LOW | Der Kommentar von `lesend` sagt zu, dass mehrere Abfragen einen Stand sehen, „das übernimmt BEGIN (aufgeschoben)". Kein Test prüft die Zusage (siehe F-4 d). Der Satz ist zudem schwer lesbar (Zusage und Kopplung in einem Halbsatz). | `AGENTS.md` §3.7, §3.11 | `src/adapters/driven/sqlite/tx.go` · `mehrere Abfragen sehen denselben Stand nur` | nein | Kommentar sagt mehr zu als ein Test prüft |
| F-7 | INFO | Randformen von `Exchange`, die `SPEC-001`/`SPEC-105` nicht entscheiden und die der Code still zulässt: negativer `ResponseStatus` (−5 wird gespeichert), `FinishedAt` vor `StartedAt`, ungültiges UTF-8 in Header-Werten (`json.Marshal` ersetzt durch U+FFFD, Roundtrip ist nicht verlustfrei), `Create` einer Session, die zugleich doppelt und mit unbekanntem Service ist (Reihenfolge der Fehler nicht festgelegt, Code liefert `ErrBereitsVorhanden`). Rückgabe an den Architect (`AGENTS.md` §3.12), nicht im Code zu entscheiden. | `AGENTS.md` §3.12 | `src/hexagon/model/exchange.go` · `func (e Exchange) Validate` | nein | Randform nicht entschieden |
| F-8 | INFO | Plan §4 nennt die Marke „über 700 Zeilen → zurück zur Zerlegung"; gemessen 1248 (Summe `git diff --numstat 7b7e877..HEAD` über `src` und `spec`). Der Slice bleibt, begründet vom Implementer selbst („einzeln nichts liefern"). Die Entscheidung über das Weiterführen gehört dem Planner/Architect, nicht dem Implementer. | `AGENTS.md` §3.14 | `docs/plan/planning/done/slice-session-recording-persistenz.md` · `**Marke überschritten:**` | nein | Marke überschritten, Entscheid nicht von der Rolle |
| F-9 | INFO | Nur Hinweis: `outboundtest` importiert `testing` in einem Nicht-Test-Paket. Das ist hier zulässig und sauber: `.a-check.yml` erlaubt `ports-outbound → model`, `outboundtest` importiert nur `model`, `outbound` und die Standardbibliothek; `make a-check`, `make lint`, `go vet` sind grün, und kein Produktpaket importiert `outboundtest` (nur `repositories_test.go`). Es ist auf diese Lage angewiesen: Sobald Bootstrap oder ein Service es importiert, hängt `testing` im Binary. | `[ADR-0003](../plan/adr/0003-lint-profil-golangci.md)` (Linter-Profil) · Maintainability | `src/hexagon/ports/outbound/outboundtest/vertrag.go` · `"testing"` | ja, `make a-check` | Hinweis |
| F-10 | INFO | Test-Hilfen mit schwacher Form, ohne Konventions-Anker: `repositories_test.go` unterscheidet zwei SQL-Anweisungen über `s[:27] == stmts[1][:27]`; `sessionFall2` trägt `_ = name` für eine ungenutzte Schleifenvariable. | Maintainability | `src/adapters/driven/sqlite/repositories_test.go` · `s[:27] == stmts[1][:27]` | nein | Hilfscode fragil |

## Prüfung der Fragen aus dem Auftrag

- **Modell gegen CHECK-Constraints:** Zustand für Zustand deckungsgleich (Aufzählungen, Blob-Kopplung, `match_key` nur bei `failed`, verworfene Bodies bei `complete`, beide `unsafe`-Regeln, `sequence_no >= 1`). Das Modell ist strenger als die Datenbank (leeres `Method`/`Path`, leerer `match_key` statt `NULL`, Nullzeit) und lässt nie einen Zustand zu, den die Datenbank ablehnt; umgekehrt kann die Datenbank per SQL Zeilen halten (`match_key=''`), die `Validate` ablehnen würde, sie werden aber lesbar geliefert.
- **Schicht-Richtung `outboundtest`:** zulässig (F-9).
- **Transaktionen:** `inTx` rollt in jedem Fehlerpfad zurück (ROLLBACK mit `WithoutCancel`), `conn.Close` über `defer`, kein Leck gefunden. `BEGIN IMMEDIATE` ist tragend: auf `BEGIN` gestellt wird `gleichzeitige Anhänge ordnen sich` rot (selbst gefahren).
- **Gleichzeitigkeit:** 8× gleichzeitig `Append`/`Create`/`Delete` auf dieselbe Session (selbst gefahren): nur erwartete Fehler (`nicht gefunden`, `bereits vorhanden`), keine verwaisten Exchanges, kein `SQLITE_BUSY` nach außen (der `busy_timeout` hängt als `_pragma` an jeder Verbindung).
- **Abbruch:** abgebrochener Kontext liefert für alle sieben Operationen `context canceled` und ändert nichts — gelebt, aber nicht gefangen (F-1).
- **Fehlercodes 787/1555/2067/1811:** richtig zugeordnet; `Delete` und `Create` ordnen sie richtig ein (Mutationen auf beide Zweige wurden rot).
- **Zeitformat:** Roundtrip stimmt für UTC; Ordnung falsch (F-2), Jahresbereich (F-3), Nullzeit wird `NULL` und kommt als Nullzeit zurück.
- **SQL-Injektion:** alle Werte gebunden; Verkettung nur von Konstanten (`exchangeSpalten`, interne `where`-Fragmente). Eine Kennung `x'; DROP TABLE sessions;--` wird wörtlich gespeichert.

## Mutationen (selbst gefahren)

Rot (Fänger wirkt): Kennungsprüfung entfernt · `ORDER BY` der Sessions geändert/entfernt · `DELETE FROM exchanges` entfernt · `BEGIN IMMEDIATE`→`BEGIN` · Sequenzprüfung entfernt · Zustandsprüfung entfernt · Startzeit-, Methoden-, Kennungs-, Request-verworfen-, Response-unsafe-Regel von `Validate` entfernt · FK-Zweig `ErrDienstUnbekannt` und `ErrReferenziert` entfernt.
Grün (Fänger fehlt oder Mutation äquivalent): siehe F-4 und F-5.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/` | geprüft, Befund F-3, F-7; sonst ohne Befund (Standardbibliothek, keine I/O) |
| `src/hexagon/ports/outbound/` | geprüft, ohne Befund (Fehler, Schnittstellen entsprechen `spec/architecture.md` §2.3 und `SPEC-105`) |
| `src/hexagon/ports/outbound/outboundtest/` | geprüft, Befunde F-1, F-2 (Blindstelle), F-4, F-9, F-10 |
| `src/adapters/driven/sqlite/` (`sessions.go`, `recording.go`, `tx.go`) | geprüft, Befunde F-2, F-3, F-6; Transaktionen, Ressourcen, Injektion, Fehlercodes ohne Befund |
| `spec/spezifikation.md` (`SPEC-105`) | geprüft, ohne Befund außer F-7 (Randformen); Nr. 1–13 einzeln gegen den Code: Nr. 1–3, 5–9, 12 umgesetzt wie zugesagt, Nr. 4 verletzt (F-2), Nr. 10 umgesetzt, Nr. 11 umgesetzt (Ordnung ungefangen, F-4), Nr. 13 umgesetzt (ungefangen, F-1) |
| Plan (`done/slice-session-recording-persistenz.md`) | geprüft, Befunde F-1, F-5, F-8; Kopf, §1, §3, §6 nachgezogen (`AGENTS.md` §3.9), Ausgänge der Risiken gesetzt |
| `.golangci.yml`, `//nolint` | geprüft, ohne Befund (keine Suppression im Diff, keine Lockerung des Profils, `AGENTS.md` §3.2/§3.6) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 2 |
| MEDIUM | 2 |
| LOW | 2 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Fänger im Plan benannt, im Code nicht vorhanden · Zusage ohne rot werdende Mutation · Sortierung über variabel breiten Zeittext · Schreibbarer Wert, der der Leser nicht liest · Kommentar sagt mehr zu als ein Test prüft · Randform nicht entschieden

## Verdikt

**Merge-blockierend:** ja — F-1 und F-2 (HIGH), F-3 und F-4 (MEDIUM). F-2 ist ein verifizierter Verstoß gegen `SPEC-105` Nr. 4/5, F-1 eine Plan-Aussage ohne Gegenstück im Code.

**Übergabe:** Findings gehen an den Implementer; F-7 zusätzlich an den Architect (Randformen), F-8 an Planner/Architect. Die Finding-Klassen gehen in die Slice-Closure §7. Dieser Report ersetzt keine Verifikation gegen die DoD.
