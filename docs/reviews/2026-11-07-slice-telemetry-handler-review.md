# Review-Report: slice-telemetry-handler — 2026-11-07

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-telemetry-handler`, `a6b822b..ae7203c` (Basis bis HEAD); `handler.go` und vier Testdateien, Plan §3, §4, §6.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-07

**Eingangs-Kontext:**

- `docs/plan/planning/in-progress/slice-telemetry-handler.md` (im Zweig): §3 Entscheidungen 18 und 19, Tabelle H01 bis H16, Mengen M1 bis M17, Runden R1 bis R6, Fenster der Sperre, Absatz *Umsetzung*, §6
- `SPEC-231` Nr. 4 und 5, `SPEC-232` Nr. 2, 3, 5 c, d, f, g, `SPEC-121` Nr. 12, `SPEC-123` Nr. 11, `SPEC-124` Nr. 9
- `AGENTS.md` §3.7, §3.9 bis §3.18
- ADR: [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Trigger Zeit/Gate-Kette, nur genannt)

**Unverändert (geprüft mit `git diff a6b822b..HEAD`):** `zeile.go`, `regeln.go`, `telemetry.go`, `telemetry_test.go`, `regeln_test.go`, `audit_test.go`.
Größenmarke: Summe der fünf Pfade 84 + 400 + 129 + 262 + 680 = 1555 gegen Marke 2000, wie berichtet.

## Eigene Mutationen

Werkzeug `tools/harness/mutationslauf.sh` (Container ohne Netz und Mount, `--memory=2g`, unter `flock`), Klasse `andere`, ein Prozess je Mutation; Host-Last 25 bis 35 von 4 CPUs.
20 Mutationen am Code (`handler.go`), ein Probe-Test in einer Kopie des Baums (nicht im Zweig).

| Mutation | Ergebnis |
|---|---|
| `WithGroup` ohne den Punkt (`h.praefix + name`) | **grün**, auch das ganze Paket (siehe Befund 1) |
| `Handle` teilt die Liste `vorgelegt` (`attrs := h.vorgelegt`) | im Zweig grün (wie berichtet, n06); mit Probe-Test `TestZZWiederEintritt` (sequentiell, ohne Detektor) **rot** (siehe Befund 2) |
| `Enabled` gegen festes `LevelInfo` statt `MinLevel` | `TestEnabled` grün (MinLevel dort stets Info), Paket rot-paket |
| nur das erste Attribut des Records (`return false`) | `TestHandlerRangfolge` grün (ein Attribut je Fall), Paket rot-paket |
| `Handle` prüft `MinLevel` vor der Zählung | rot (`TestHandlerMinLevelZeit`) |
| `Uint64`: `<` statt `<=` | rot |
| `WithAttrs` ohne Präfix des Empfängers im Kind; ohne Präfix beim Abbilden | rot (`TestWithGroup`) |
| `msg` leer; `Level+1`; Duration halbiert; Float verschoben; `Bool` zu `Int`; `Resolve` entfernt; `slices.Clone(h.vorgelegt[:0])`; `WithGroup("")` liefert Kopie | je rot |
| `Time` zu `String`; `Uint64` ohne Prüfung (`\|\| true`); leeres `WithAttrs` liefert neuen Handler | grün, **gleichwertig** (siehe Antworten c, f unten): der Text von `Time` enthält Leerzeichen, jeder Schlüssel der Menge `Log` lehnt ihn ab; ein `Uint64` über `MaxInt64` wird zu negativem `Int`, `befunde` lehnt es ab; `SPEC-231` Nr. 5 sagt „ob es derselbe ist, ist nicht zugesagt" |

## Findings

### F1 · MEDIUM

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Zusage ohne Fänger), §3.16
- `pfad`: `src/adapters/driven/telemetry/handler.go` · `praefix: h.praefix + name + "."`
- `befund`: `SPEC-231` Nr. 5 sagt „setzt vor jeden folgenden Schlüssel `name.`“. Entfällt der Punkt, bleiben `TestWithGroup` und das ganze Paket grün; `WithGroup("service")` mit dem Attribut `_id` ergibt dann den gültigen Schlüssel `service_id` und schreibt ein Feld statt `schluessel` zu zählen. Die Begründung des akzeptierten Negativs m22 (jedes Präfix enthalte einen Punkt, kein Schlüssel der Menge einen) stützt sich auf genau diesen Punkt, den kein Test hält; die Fixtures wählen Gruppen (`g`, `a`, `b`, `n`), deren Verkettung nie einen Schlüssel der Menge ergibt.
- `verifizierbar`: ja (Mutation oben, `mutationslauf.sh`)
- `klasse`: Fänger führt nur eine Form der Eingabe (Zerlegung eines Schlüssels in Gruppe und Rest)

### F2 · MEDIUM

- `kategorie`: MEDIUM (Antwort auf (a), Rückgabe an den Architect)
- `quelle`: `AGENTS.md` §3.15, §3.10 (Zusage ohne Fänger, der sie deterministisch fängt), Plan §6 (Eintrag *Anhang durch `Handle`*)
- `pfad`: `slice-telemetry-handler.md` · „fängt allein „`WithAttrs` ohne Kopie“ (Mutation m13)“; `handler.go` · `attrs = append(attrs, h.vorgelegt...)`
- `befund`: Die Berichtigung ist ehrlich, aber die Aussage „sequentiell unsichtbar, nur der Detektor sieht es“ trifft nicht zu. Ein Eintrag, dessen `LogValue` auf demselben Handler `Handle` ruft (Wiedereintritt, `Resolve` läuft in `Handle` zwischen zwei `append`), überschreibt die geteilte Zelle: Kette aus drei `WithAttrs` (Kapazität 4, Länge 3), Record mit `exchange_id=ex-1` und einem `LogValuer`, der `Handle` mit `exchange_id=ex-2` ruft. Mit der Mutation steht in der äußeren Zeile `ex-2`; der Probe-Test ist ohne Detektor in einem Prozess rot (Probe nur in einer Kopie des Baums, nicht im Zweig). Ein Fänger ohne Detektor und ohne Rate ist damit schreibbar; der Zweig führt die Zusage nur über `make test-race`, mit der Folge, dass der Fänger bei jedem Lauf ohne Detektor (Stufe `test`) grün bleibt.
- `verifizierbar`: ja
- `klasse`: Zusage nur mit Detektor getragen, sequentieller Wirkungs-Fänger übersehen

### F3 · MEDIUM

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.13 (Adresse nimmt an, bevor sie genannt wird), §3.15 (Orakel aus dem Code)
- `pfad`: `slice-telemetry-handler.md` · „Fixture der Ereignisse (M13): die Texte der drei Ereignisse von `record` stehen nicht im Wortlaut von `SPEC-121` Nr. 12“
- `befund`: Richtig geführt ist die Feststellung (`SPEC-121` Nr. 12 sagt „festem Text“, die drei Texte `record: Aufzeichnung eines Bodys verworfen`, `record: Upstream ohne Antwort`, `record: Exchange nicht angehängt` stehen nur im Code und, für `capture`, im Beispiel von `SPEC-231` Nr. 4; die Texte von `session` und `maintenance` stehen in der Spezifikation). Die Adresse trägt sie nicht: `slice-telemetry-einbau-gateway` §1 nennt Fänger je Stelle (c) und den Wortlaut zum Port, aber weder diese Lücke noch den Geber `slice-telemetry-handler`; eine Spezifikationslücke gehört zudem dem Architect. Bis dahin ist der Text von `upstream` und `ablage` im Test ein Orakel aus dem Code. Zusätzlich widerspricht der Absatz *Mengen* (M13: Texte „stehen so in `SPEC-121` Nr. 12“) dem Eintrag in §6.
- `verifizierbar`: nein (Plan-Prüfung)
- `klasse`: offene Entscheidung ohne angenommene Adresse

### F4 · LOW

- `kategorie`: LOW
- `quelle`: Maintainability, `AGENTS.md` §3.11
- `pfad`: `handler_test.go` · `func mengeLaenge()` und `zeile_test.go` · `func TestMsgLaenge`
- `befund`: Zweite Fassung des Generators M2 (Duplikat von `TestMsgLaenge`, `zeile_test.go` Zeilen 252 bis 268); gleiche Zählprüfung 469, aber die Grenzen sind in der Kopie als `250..257` fest geschrieben statt aus `grenze` gebildet. Weicht eine Fassung ab, läuft die andere unbemerkt weiter; Plan §6 führt das (Adresse Architect, Hilfe heben) mit Ausgang offen. Bleibt als Folge: ein Gate fängt die Abweichung nur über die Zahl.
- `verifizierbar`: nein
- `klasse`: Hilfe doppelt geführt

### F5 · LOW

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.11
- `pfad`: `handler_test.go` · `func TestEnabled`; Plan *Umsetzung* · „H14 | `Enabled` prüft `ctx.Err()` (m32); immer wahr (m33)“
- `befund`: `TestEnabled` hat nur `MinLevel` Info; „`Enabled` gegen festes Info“ bleibt dort grün (`TestHandlerMinLevelZeit`, `TestAufrufstellen` fangen sie). Die Zeile H14 nennt als Fänger allein `TestEnabled`, ein Paar von MinLevel-Werten fehlt dort.
- `verifizierbar`: ja
- `klasse`: Fänger führt nur eine Form der Eingabe

### F6 · INFO

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `handler.go` · `u <= math.MaxInt64`
- `befund`: Die Prüfung `Uint64 <= MaxInt64` ist beobachtbar nur über die Klemme (m04), nicht über den Wrap: ein Wert darüber würde als negativer `Int` von `befunde` abgelehnt, das Ergebnis `wert` wäre gleich. Kein Fehler; der Code folgt der Zusage der Spezifikation.
- `verifizierbar`: ja
- `klasse`: gleichwertige Mutation

## Antworten zu den Schwerpunkten

- **(a)** H05 fängt den Anhang an `vorgelegt` durch `Handle` nicht; der Bericht sagt es ehrlich (Plan §6, im selben Commit berichtigt). Die Annahme „kein sequentieller Fänger“ ist falsch (F2).
- **(b)** `mengeLaenge` ist ein Duplikat (F4).
- **(c)** m22 ist als Begründung richtig, soweit der Punkt im Präfix steht; sie ist unvollständig, weil der Punkt selbst ungefangen ist (F1). Die Verkettung der Namen (nur letzter Name) ist für die Zeile und die Zähler unbeobachtbar, Grund und Folge stehen.
- **(d)** Stimmt: `slog.Value.Resolve` (Go 1.27.1, `value.go`) setzt `fmt.Errorf("LogValue panicked\n%s", stack(3, 5))`, der Panikwert ist nicht enthalten; die Stelle kann nie rot werden. Akzeptiertes Negativ mit Grund und Folge steht.
- **(e)** n06 flatternd 50/50 mit, 0/50 ohne: die Zahl ist plausibel; der Wirkungs-Fänger ist schreibbar (F2).
- **(f)** „Zählung nur beim ersten Eintrag“ und „`Resolve` bei jedem Eintrag“: im Zuschnitt nicht schreibbar, weil der Handler beim Ableiten abbildet und je Eintrag nichts hält. Stimmt; die Klasse „zu oft aufgelöst“ deckt `TestWithAttrsLogValuerEinmal`.
- **Kopier-Handler:** `WithAttrs` klont `vorgelegt` (Mutation `[:0]`/ohne Kopie rot), `WithGroup` teilt `vorgelegt` nur lesend (Liste unveränderlich nach dem Ableiten, `Handle` baut je Aufruf neu). Präfix bei `WithAttrs` nach `WithGroup` rot gefangen. Uint64 > MaxInt64 und nicht abbildbare Arten: Art 0, Ergebnis `wert`.
- **Zählprüfung gegen `slog.Kind`:** `(Kind+1).String()` ist in Go 1.27.1 `<unknown slog.Kind>`; die Prüfung vergleicht damit. Ein neuer Kind-Wert färbt den Test.
- **§3.18:** `LC_ALL=C grep -P '\\u[0-9A-Fa-f]{4}'` über `handler*.go` und den Plan: kein Treffer, kein BOM, kein NBSP, kein U+2028.
- **§3.9:** §3, §4, §6 und die Zeilen `basis`, `schaetzung`, `marke` sind nachgezogen; die berichtigte Zeile in Antwort G und im Fenster der Sperre stimmt mit §6 überein.
- **M13 gegen SPEC-121 Nr. 12 / SPEC-123 Nr. 11 / SPEC-124 Nr. 9:** zehn Ereignisse, Stufen und Attribute stimmen mit `spezifikation.md` und den Diensten überein; die Texte der drei `record`-Ereignisse stehen nicht im Wortlaut der Spezifikation (F3).
- **Abdeckung (Register `zweig-ohne-lauf-unter-dem-detektor`):** selbst nicht gemessen (Last des Hosts 25 bis 35, Warteschlange an `flock`); ersatzweise Mutationen an jedem Zweig von `handler.go` (jeder Block von `Enabled`, `Handle`, `WithAttrs`, `WithGroup`, `abbilden`: rot, außer den gleichwertigen oben). Die Zählung der Runden R1 bis R6 im Plan wurde nicht nachgemessen.

## Geprüft, ohne Befund

- `src/adapters/driven/telemetry/handler.go` (Kommentare tragen Zusage, Kopplung oder Abgrenzung, keine Chronik)
- `src/adapters/driven/telemetry/handler_sperre_test.go` (Haltepunkte, Fenster der Sperre)
- `src/adapters/driven/telemetry/handler_canary_test.go` (35 × 4 × 2 gegen die Quelle gezählt)
- `src/adapters/driven/telemetry/handler_arten_test.go` (Zählung 53 bzw. 106 und 80 stimmt)
- `zeile.go`, `regeln.go`, `telemetry.go` und die drei fremden Testdateien: unverändert
