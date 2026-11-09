# Review-Report: slice-hybrid-service (Teil 1: Mock und Replay-Fallback) — 2026-11-09

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** `slice-hybrid-service`, Zweig `slice-hybrid-service`, `6de6da87..b8131aba`; Pfade `src/hexagon/ports/inbound/hybrid.go` und `src/hexagon/services/hybrid/{verteiler,puffer,antwort,fehler}.go` samt Tests, Plan-Abschnitte §3 (Tabelle Z01 bis Z06, Z08 bis Z18, Z26, Z28 bis Z34, Z36; *Stand der Umsetzung Teil 1*) und §6 (Raster, Abweichungen a bis d).

**Skill:** `.harness/skills/reviewer.md` @ `b8131aba` <!-- d-check:ignore (Zitat des Standes im Lauf) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-09

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde):

- Slice-Plan `slice-hybrid-service` (Teil 1, Messung 1990 bei Marke 3900)
- `SPEC-284` Nr. 1 bis 13 (Teil 1: Stufen M und R) und `SPEC-261` (Policy, `Entscheide`)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-12`, `LH-FA-04`
- `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.18

**Vorgehen und Belege.** Gelesen: der gesamte Produktcode (234 Zeilen) und die Tests. Gefahren (Image `odata-recorder:race`, `--network none`, `--memory=4g`, `--cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`, Quellen schreibgeschützt eingehängt, Mutationen in einer Kopie im Container):

1. Die Liste des Implementers über `tools/harness/mutationslauf.sh` (83 Mutationen): 79 rot, 4 grün, 0 flatternd, 71 s, Last 3,4 bis 3,6 von 4 — reproduziert; die vier grünen sind die benannten.
2. Rund 30 eigene Mutationen, die nicht in der Liste stehen (Tabelle unten), je mit Prüfung, dass das Muster die Quelle ändert.
3. Ein eigenes Orakel-Probe-Paket `orakel_probe_test.go` für die Pufferung (32220 Fälle, Menge aus dem Wortlaut von `SPEC-284` Nr. 5: Policy {keine, `coverage_insufficient`, `unsupported`} × `MaxBytes` {1, 2, 3, 5} × Länge 0 bis 2·`MaxBytes`+3 × deklarierte `Laenge` {-1, 0, 1, N, N-1, N+1, N+3, `MaxBytes`, `MaxBytes`+1, 2·`MaxBytes`+4} × Fehler des Bodys nach k Bytes × Lesemodus von M {nichts, ein Byte, alles} × Code {`ODR-E6003`, `ODR-E6005`}). Der Code besteht es unverändert; das Paket liegt nur im Scratchpad und ist nicht committet.
4. Nebenläufigkeit: Test-Binaries (`go test -c`, mit und ohne `-race`), je Mutation 30 Prozesse mit `-test.count=1`.
5. Gesamtpaket unter `-race -count=20`: grün (2,8 s).

Gegengelesen auf Rohbytes (`LC_ALL=C grep -P '\\u[0-9A-Fa-f]{4}'`): alle Testdateien des Pakets, kein Treffer.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Zusage „`Laenge` gleich der Länge des Puffers“ (`SPEC-284` Nr. 5 c) ist nur für eine ehrliche oder zu knapp deklarierte `Laenge` gefangen: die Mutation „`Laenge` nur bei -1 auf die Pufferlänge setzen“ bleibt für M wie für R grün in der Suite (rot allein im Orakel-Probe). Die Fälle „gelogen“ in `TestGrenze` setzen `min(n, max)`, bei `n` ≤ `max` ist die Deklaration also wahr; eine deklarierte `Laenge` kleiner oder größer als die gelesene Zahl, beide höchstens `MaxBytes`, fehlt. Folge: M oder R erhielten eine `Laenge`, die nicht zu den Bytes ihres Lesers passt. | `AGENTS.md` §3.15 (ein Fall je Form der Eingabe; Allaussage als Schleife) | `puffer.go` · `m.Body, m.Laenge = bytes.NewReader(puffer), int64(len(puffer))` und `r.Body, r.Laenge = bytes.NewReader(puffer), int64(len(puffer))`; Test `body_test.go` · `laenge = min(n, max)` | ja — Mutation `if a.Laenge < 0 { m.Laenge = … }` (ebenso für `r`), Suite grün | Fänger führt nur eine Form der Eingabe (gelogene Länge unterhalb der Grenze) |
| F-2 | MEDIUM | Die Zusage „`New` weist einen Dienst mit leerem `Root` ab“ (`SPEC-284` Nr. 1) ist für eine Liste mit leerem `Root` nur am einzigen oder am letzten Eintrag gefangen. Die Mutation „nur der letzte Dienst wird geprüft“ bleibt grün; ein leerer `Root` vor einem belegten letzten Eintrag passiert `New`. | `AGENTS.md` §3.15 (ein Fall je Element einer Menge) | `verteiler.go` · `slices.ContainsFunc(k.Dienste, func(d Dienst) bool { return d.Root == "" })`; Test `konfig_test.go` · `Root später leer` | ja — Mutation `k.Dienste[len(k.Dienste)-1].Root == ""`, Suite grün (die Mutation „nur der erste“ ist rot) | Fänger führt nur eine Position der Menge |
| F-3 | MEDIUM | Die Zusage „Code hat ein Fehler genau dann, wenn `errors.As` einen `*inbound.Fehler` findet“ (`SPEC-284` Nr. 3 b) ist nur für direkte Werte gefangen: ein verpackter oder verbundener `*Fehler` mit Auslöser-Code (`fmt.Errorf("…: %w", f)`, `errors.Join(f, andererFehler)` ohne Kontextfehler) fehlt in der Menge der Formen. Die Mutation „Typzusicherung `err.(*inbound.Fehler)` statt `errors.As`“ bleibt grün; in diesem Fall bliebe der Replay-Fallback still aus, und der Client erhielte `ODR-E6003`. | `AGENTS.md` §3.15 | `fehler.go` · `if !errors.As(err, &f) \|\| f.Code == 0 {`; Tests `ablauf_test.go` · `TestFehlerOhneCode`, `TestCodesSchleife` (nur direkte Werte, Verpackung nur um Kontextfehler) | ja — Mutation, Suite grün | Fänger führt nur eine Form der Eingabe (Verpackung des Fehlers mit Code) |
| F-4 | MEDIUM | Nebenläufigkeit Z31/Z32: Die Zusage „kein Zustand, keine Sperre nach `New`“ ist für Datenrennen vom Detektor in jedem Prozess gefangen (30 von 30 je Prozess, Mutation ohne künstliches Fenster, siehe Tabelle unten), nicht vom Wirkungs-Fänger (0 von 30 ohne Detektor). Für Zustand mit geordneten Zugriffen (Feld im `Verteiler`, durch einen Kanal als Sperre geschützt; kein `sync`-Import) sind beide Fänger nicht in jedem Lauf rot (mit Detektor 22 und 25 von 30, ohne 18 und 25 von 30), `TestQuelltext` prüft keine Felder des `Verteiler`, und der Plan nennt die Klasse weder als Fänger noch als akzeptiertes Negativ mit Grund und Folge. | `AGENTS.md` §3.10 (Nebenläufigkeit: „in jedem Lauf rot“, sonst Wirkungs-Fänger oder Negativ; „Prüfen und Handeln … bei lauter geordneten Zugriffen“) | `nebenlaeufig_test.go` · `func TestGleichzeitig(t *testing.T)`; `quelltext_test.go` · `typ != "Verteiler" && typ != "fehlerLeser"` | ja — Mutation n04 (Feld `letzter`, Kanal der Kapazität 1 als Sperre), 30 Prozesse je Variante | Zusage über Nebenläufigkeit ohne Fänger in jedem Lauf und ohne Negativ |
| F-5 | LOW | Die Mutation „`context.WithoutCancel(ctx)` an M“ beendet die Suite nicht mit `--- FAIL`, sondern mit `panic: test timed out after 10m0s`: der Fake in `TestAbbruchZwischenStufen/Frist in M` wartet ohne eigene Grenze auf `c.Done()`. Rot, aber nicht aus dem benannten Test; `TestKontextAnStellen` allein wird für dieselbe Mutation rot (`--- FAIL`). Dasselbe gilt für `Background` an M. | Reviewer-Skill (Hinweis zu `panic:`/Zeitlimit); `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` | `ablauf_test.go` · `<-c.Done()` | ja — Mutation `v.mock.Mock(context.WithoutCancel(ctx), anfrageM)` | Fake blockiert unbegrenzt, Zeitlimit statt FAIL |
| F-6 | LOW | Der Plan nennt „Messung der Pfade aus dem Block 1989 Zeilen“; `git diff --numstat 6de6da87..b8131aba` über die Pfade des Plans ergibt 1990 (der letzte Commit änderte einen Test, die Zahl zog nicht mit). Die Marke 3900 ist weit entfernt, die Zahl ist ein Zustandsfeld. | `AGENTS.md` §3.9 (Plan folgt jeder Korrektur), §3.14 | `docs/plan/planning/done/slice-hybrid-service.md` · `Messung der Pfade aus dem Block 1989 Zeilen` | ja — `git diff --numstat` | Zahl im Plan nicht nachgezogen |
| F-7 | INFO | Abweichung (a), Dienstwahl: begründet. `New` prüft `Dienste` und speichert sie nicht; in Teil 1 hat `SPEC-284` Nr. 4 keine beobachtbare Wirkung (`Ziel` immer leer, `ODR-E5005` liefert der Mock). §6 trägt die Abweichung mit Ausgang *offen* und Adresse Planner, wie `AGENTS.md` §3.13 es verlangt; der Ausschluss in Teil 2 („Dienstwahl ohne `UpstreamURL`“ als Teil 1) macht Teil 2 derzeit nicht zum Nehmer. Zuständig: Planner. | `AGENTS.md` §3.13 | Plan §6 · `Dienstwahl` | nein — Planungsentscheidung | Nehmer der Abweichung fehlt (Adresse Planner, in Arbeit) |
| F-8 | INFO | Abweichungen (b) und (c): bestätigt unbeobachtbar. Mit leerem `Ziel` liefert `Entscheide` für `ODR-E5001` nie `Upstream` (H5 entfällt ohne Wirkung); `Pfad` und die Zahl der Aufrufe von `Entscheide` sind an `*policy.Policy` (konkreter Typ, Replay-Zweig liest `Pfad` nie) nicht zu fangen. Zusätzlich grün und in Teil 1 äquivalent, aber im Plan nicht genannt: `Ziel` ungleich leer am Aufruf H3, und `Upstream` wie `Replay` behandelt. Alle vier gehören zu den Fängern Z19 bis Z21 in Teil 2. Grund und Folge der Negative stehen nur in Plan §6, nicht in `SPEC-284` *Ungeprüft*. | `AGENTS.md` §3.15 (akzeptiertes Negativ mit Grund und Folge) | Plan §6 · `akzeptierte Negative mit Grund und Folge`; `verteiler.go` · `v.policy.Entscheide(policy.Anfrage{Methode: a.Methode, Pfad: a.Pfad, Code: code}) != policy.Replay` | ja — Mutationen `Ziel: "https://h"` an H3 und `== policy.Bleiben` bleiben grün | Negativ ohne Ort in der Spezifikation (Teil 2 nimmt auf) |
| F-9 | INFO | Abweichung (d), Randform ohne Ort: bestätigt und reproduziert. Liefert M `(*inbound.Fehler)(nil)` als `error`, panikt `Hybrid` mit `invalid memory address or nil pointer dereference` (in `errors.Is` über `Fehler.Unwrap`); `SPEC-284` nennt die Form nicht, der Dienst entscheidet sie nicht. Zuständig: Architect (in Arbeit); danach braucht Z04 die Form. | `AGENTS.md` §3.12, §3.16 | `fehler.go` · `if errors.Is(err, context.Canceled) \|\| errors.Is(err, context.DeadlineExceeded) {` | ja — Probe `TestProbeTypedNil` (nicht committet) | Randform ohne Ort (Adresse Architect, in Arbeit) |
| F-10 | INFO | Sicherheitsachse „kein Upstream-Zugriff in Teil 1“ bestätigt (siehe Negativbefunde). Das Gate dafür ist im Wesentlichen strukturell: kein Port, kein Import (`TestQuelltext` verbietet `net/http`, Adapter, andere Dienste), jede `Entscheide`-Stelle gibt leeres `Ziel`, `Upstream` wird wie `Bleiben` behandelt. Die Teil-1-Zusage „jeder andere Wert ist `Bleiben`“ ist als Mutation nicht fangbar (F-8) und hängt damit an der Bauweise. | `LH-FA-12` | `puffer.go` · `ist leer, die Policy gibt es nie frei` | nein | — |

---

## Auflösung der Tabelle (Z01 bis Z36 Teil 1) mit je eigener Mutation

Eigene Mutationen (nicht in der Liste des Implementers), Suite des Implementers ohne mein Probe-Paket beurteilt; „rot“ heißt `--- FAIL` eines Tests, „Zeit“ heißt Zeitlimit.

| Zeile | Eigene Mutation | Ergebnis (roter Test) |
|---|---|---|
| Z01 | nur erster Dienst auf `Root` geprüft / nur letzter Dienst | rot `TestNew` / **grün (F-2)** |
| Z01 | `New` ruft `Mock` | rot (u. a. `TestErfolgUnveraendert`) |
| Z02 | Prüfung der Anfrage nach der Pufferung (Body wird vor `ErrUngueltig` gelesen) | rot (`TestGrenze` u. a.; Prüfung nach dem Lesen) |
| Z02 | `ErrUngueltig`-Text trägt `Methode` und `Pfad` | rot `TestUngueltigeAnfrage` |
| Z02 | Mock wird bei ungültiger Anfrage gerufen | rot `TestUngueltigeAnfrage` |
| Z03/Z04 | `errors.As` durch Typzusicherung ersetzt | **grün (F-3)** |
| Z05 | `strings.ToUpper(Methode)` an `Entscheide` | rot `TestEntscheideArgumente`, `TestReplayMiss` |
| Z05 | `Ziel` ungleich leer an H3; `Upstream` wie `Replay` behandelt | grün, in Teil 1 äquivalent (F-8) |
| Z06 | `Replay` bei Ergebnis `!= Bleiben` | grün, äquivalent in Teil 1 (F-8) |
| Z10 | `ctx.Err()` vor der Entscheidung (Fehler ohne Auslöser wird zum Kontextfehler) | rot `TestAbbruchZwischenStufen`, `TestGleichzeitigerAbbruch` |
| Z10 | feste `context.Canceled` statt `ctx.Err()` | rot `TestAbbruchZwischenStufen` |
| Z11 | `WithoutCancel` an M | rot `TestKontextAnStellen` (Gesamtlauf: Zeit, F-5) |
| Z11 | `Background` an R / `WithValue` an R / `WithCancel` an M | rot `TestKontextAnStellen` (Background auch `TestGleichzeitigerAbbruch`) |
| Z12 | Probe mit festem `Code` `ODR-E6003` / mit fester `Methode` `POST` | rot `TestPufferung` / rot `TestOhnePufferung` |
| Z12 | `Body` nil nur bei `Laenge` ≥ 0 als „ohne Body“ | rot `TestOhnePufferung` (Panik) |
| Z13 | `Laenge` nur bei -1 auf Pufferlänge (M; R) | **grün (F-1)** |
| Z13 | Leser von R halbiert | rot `TestGleichzeitig`, `TestGrenze` u. a. |
| Z14 | Vergleich mit `int` statt `int64`; `Laenge != 0` ergänzt | grün, auf 64 Bit äquivalent |
| Z15 | Lesefehler nur bei gelesenen Bytes > 0 beachtet | rot `TestLesefehler` |
| Z15 | `fehlerLeser` liefert bei Puffer der Länge 0 `(0, nil)` | grün, äquivalent (`io.ReadAll` ruft nie mit leerem Puffer) |
| Z26 | Schreibweise `Replay`; Schlüssel `x-recorder-fallback` | rot `TestKennzeichnung` |
| Z26 | Vergleich mit `strings.EqualFold` statt ASCII-Vergleich (KELVIN SIGN, U+212A) | rot `TestKennzeichnung` |
| Z26 | alle `X-Recorder-*` der Stufe gefiltert | rot `TestKennzeichnung` |
| Z26/Z28 | `Status` auf 200 gesetzt; Body umhüllt (`struct{ io.ReadCloser }`); Wertliste geteilt (statt geklont) | rot (`TestKennzeichnung`, `TestGleichzeitigerAbbruch`, `TestKategorieOrakel`) |
| Z28 | `Close` nur bei Fehler mit Code; `Close` von R nicht bei Kontextfehler | rot `TestStromSchliessen` |
| Z29 | Mock vor R ein zweites Mal gerufen | rot (`TestCodesSchleife`, `TestGleichzeitig` u. a.) |

Zeilen ohne eigene Mutation neben der Liste des Implementers: Z16 bis Z18, Z30 (Liste reproduziert, rot), Z33, Z34 und Z36 (akzeptierte Negative mit Grund und Folge in `SPEC-284` *Ungeprüft*, geprüft).

## Raster (Plan §6, Operation × Argumentart)

Jede Zelle von „Aufruf des Mock“, „Replay-Fallback“ und „Schließen der Ströme“ (Teil 1) verweist auf eine Nummer, deren Wortlaut die Entscheidung trägt; keine Zelle ist im Code entschieden. Aufgelöst mit je einer Mutation: fehlender Body (Z12, Mutation `Body` nil), leerer Body (`Laenge` 0 mit Leser, Liste), unbekannte Methode (`get`, Liste `z08`), Abbruch (Z10/Z11/Z32, eigene und Listenmutationen), Wiederholung (Z30). Die einzige Zelle ohne Ort ist (d) (F-9, Adresse Architect). Die Zusatztabelle (Ergebnis des Mock × Zustand der Policy, Spalten M und R): Fänger Z03, Z06, Z08, Z09 laufen als Schleife über alle Codes mit Zählprüfung gegen `model.AlleCodes()`.

## Nebenläufigkeit Z31/Z32 (Messung, Prozesse mit `-test.count=1`, Last des Hosts 1,4 bis 3,9)

| Mutation (natürliches Fenster, keine künstliche Schleife) | Test | mit Detektor | ohne Detektor |
|---|---|---|---|
| Variable auf Paketebene hält den letzten Fehler | `TestGleichzeitig` / `TestGleichzeitigerAbbruch` | 30/30 / 30/30 | 0/30 / 0/30 |
| gemeinsamer Puffer auf Paketebene | `TestGleichzeitig` | 30/30 | 0/30 |
| Feld `letzter` im `Verteiler`, ohne Sperre | `TestGleichzeitig` / `TestGleichzeitigerAbbruch` | 30/30 / 30/30 | 0/30 / 0/30 |
| Feld `letzter`, Kanal der Kapazität 1 als Sperre (geordnete Zugriffe) | `TestGleichzeitig` / `TestGleichzeitigerAbbruch` | 22/30 / 25/30 | 18/30 / 25/30 |
| dasselbe mit einer Schleife von 300000 Kanalwechseln im Fenster | beide | 30/30 | 30/30 |
| ohne Mutation (Grundlauf) | beide | 0/30 | 0/30 |

Bewertung: Die Aussage des Implementers stimmt. Die Zusage „kein Datenrennen, kein geteilter Zustand“ ist durch den Detektor je Prozess in jedem Lauf gefangen (30 von 30, unabhängig von einem künstlichen Fenster, selbst gemessen); der Wirkungs-Fänger trägt allein dann, wenn die Mutation selbst ein Fenster öffnet, und ist deshalb für natürliche Mutationen kein zweiter Fänger. Das ist nach `AGENTS.md` §3.10 für Datenrennen zulässig. Nicht gedeckt ist die Klasse „Zustand mit geordneten Zugriffen“ (F-4).

## Sicherheitsachse und Bewertung der Fragen aus dem Auftrag

- **Kein Upstream-Zugriff in Teil 1:** geprüft. Die Quellen kennen weder `UpstreamClient` noch `Senden` noch `net/http`; der Dienst hält nur `Mock`, `Replay`, `Policy`. Alle vier Aufrufe von `Entscheide` tragen ein leeres `Ziel`; `hostFrei("")` liefert `false`, `Upstream` entsteht nicht, und der Dienst behandelt jedes Ergebnis außer `Replay` als `Bleiben`. `TestQuelltext` (Importe, Methoden, `go`, Variablen auf Paketebene) und die Architektur-Regeln fangen einen neuen Weg strukturell.
- **Mock genau einmal beim mutierenden Replay-Fallback:** eigene Mutation „Mock vor R ein zweites Mal“ rot (neben der Mutation „nach R“ der Liste); Fänger Z29 zählt Aufrufe von M (1) und R (1) für fünf Methoden.
- **Body-Pufferung:** Orakel-Probe (32220 Fälle) und Suite stimmen mit dem Code überein; Befund F-1 ist die eine Form, die die Suite nicht führt. Geprüft und gefangen: `MaxBytes`+1-Leser (`LimitReader` ohne `+1`, Liste), Überlauf-Rest (`Rest nicht angehängt`, Liste), Lesefehler nach 0, 3, `MaxBytes` Bytes (auch meine Mutation „Lesefehler nur bei Bytes > 0“ rot), `MaxInt64` ohne Überlauf (Liste), `Laenge` größer `MaxBytes` ohne Lesen (Liste).
- **Header `X-Recorder-*` nur im Replay-Fallback / `X-Recorder-Fallback: replay`:** Z26 rot für Schreibweise, Wert, Schlüssel, Status, Wertliste, Umhüllung; R erhält alle `X-Recorder-*` unverändert (Z17).
- **Stromschluss:** verworfene Antwort genau einmal (Mutationen fehlt, doppelt, nur bei Code, nicht bei Kontextfehler rot); ausgelieferte Antwort nie, Body der Anfrage nie (Z16).
- **Kontext je Aufrufstelle:** `Background`, `WithoutCancel`, `WithValue`, `WithCancel` an H2 und H4 sowie `ctx.Err()` vor R rot (F-5 für den Hang bei `WithoutCancel`/`Background` an M).
- **Abweichungen:** (a) F-7, (b) und (c) F-8, (d) F-9. Alle vier stehen im selben Commit in §6 mit Ausgang *offen* und Adresse.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/inbound/hybrid.go` und `hybrid_test.go` | geprüft, ohne Befund |
| `src/hexagon/services/hybrid/verteiler.go`, `antwort.go`, `fehler.go` (Ablauf, Kennzeichnung, Fehlerprüfung gegen `SPEC-284` Nr. 2, 3, 6, 7, 9, 10, 11, 13) | geprüft, ohne Befund im Produktcode; Befunde betreffen Fänger (F-1 bis F-4) |
| `src/hexagon/services/hybrid/puffer.go` (Pufferung gegen `SPEC-284` Nr. 5) | geprüft, ohne Befund im Produktcode (Orakel-Probe 32220 Fälle konform) |
| Kommentare im Produktcode gegen `AGENTS.md` §3.7, §3.11 | geprüft, ohne Befund (Indikativ, Klassen Zusage/Kopplung/Rang-Zeiger) |
| Suppressionen (`//nolint`, Lockerung eines Gates, `AGENTS.md` §3.2, §3.6) | geprüft, ohne Befund (kein Treffer im Diff; `.golangci.yml`, `Makefile`, `harness/` unverändert) |
| Escape-Folgen und Rohbytes in den Tests (`AGENTS.md` §3.18) | geprüft, ohne Befund (KELVIN SIGN aus Bytefolge) |
| `docs/plan/planning/done/slice-hybrid-service.md` (§3 Tabelle und Stand, §6) | geprüft; Befund F-6, Abweichungen siehe F-7 bis F-9 |
| `spec/spezifikation.md` `SPEC-284` (Teil 1) | geprüft, ohne Befund außer der Randform ohne Ort (F-9, Architect) |

## Summary

**0 HIGH · 4 MEDIUM · 2 LOW · 4 INFO.** MEDIUM: F-1 (gelogene `Laenge` unterhalb der Grenze ohne Fänger), F-2 (leerer `Root` nur an einziger/letzter Stelle gefangen), F-3 (verpackter `*Fehler` mit Auslöser-Code ohne Fänger), F-4 (Zustand mit geordneten Zugriffen weder in jedem Lauf gefangen noch als Negativ benannt). Sicherheitsachse (kein Upstream in Teil 1) ohne Befund. Die Aussage des Implementers zu Z31/Z32 stimmt; die Zusage gilt für Datenrennen durch den Detektor (30 von 30 je Prozess, selbst gemessen), nicht durch den Wirkungs-Fänger. Die Liste des Implementers (79 von 83) ist reproduziert. Nicht Gegenstand dieses Reports: DoD (Verifier), Gate-Läufe (`make gates`).

Scratchpad (nicht committet): `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/rev-hy/` mit Mutationsskripten (`muts/`, `nat/`), Orakel-Probe und Ergebnissen.
