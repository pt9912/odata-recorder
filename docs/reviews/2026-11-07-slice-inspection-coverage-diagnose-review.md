# Review: slice-inspection-coverage-diagnose — 2026-11-07

**Art:** Review (einmal), Code gegen Plan, `SPEC-236` Nr. 1, 2, 5 bis 8, `SPEC-109` Nr. 5, `SPEC-110` Nr. 2, `SPEC-230` Nr. 9, `LH-FA-13`, `LH-FA-11`, `LH-QA-04`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und die Hard Rules; nicht gegen die DoD (Verifier).

**Gegenstand:** Zweig `slice-inspection-coverage-diagnose`, HEAD `331d3d8`, Basis `9a9180d`/`eae2d8e`. Plan: `slice-inspection-coverage-diagnose` (§3 Entscheidung 5 a bis m, Tabelle K2 bis K5, A2, C1 bis C6, D1 bis D8, N2, W3, W4, Absatz *Ergebnis der Implementierung*, Mengen, §6).

**Rolle/Modell:** reviewer · claude-sonnet-5-5.

## Urteil

Kein HIGH, kein MEDIUM. Der Produktcode (109 Zeilen) setzt `SPEC-236` Nr. 1, 2, 6, 7, 8 um; Allowlist, Gegenstand `Snapshot`, Kontextverhalten und Nullwerte stimmen mit dem Vertrag. Von 22 eigenen Mutationen blieben 3 grün (alle unten als LOW oder INFO), 16 rot, 1 nur im Paket rot, 2 `flatternd` wie behauptet. Gefunden sind vier LOW (Reihenfolge von `Erlaubt` ohne Fänger, Kennung länger als 64, Negativ „kein Zähler im Dienst" fehlt in N2, Nachzug in §6 noch offen geführt) und sechs INFO zu den Abweichungen, zu K5 und zur Prüfgrenze von m.

## 1. Eigene Belege

| Prüfung | Beleg |
|---|---|
| Diff | 9 Dateien, Port +26, `inspektion.go` 83 Zeilen geändert (109 Zeilen Produktcode zusammen mit dem Port), Tests; `.golangci.yml`, `.a-check.yml`, `spec/`, `harness/` im Diff nicht berührt (`git diff 9a9180d..HEAD --stat`). Keine `//nolint` im Paket. |
| Imports des Dienstes | `maps`, `errors`, `context`, Modell, Ports (`inbound`, `outbound`, `support`): keine Adapter, kein anderer Dienst, kein `os`/`net`/`log`. |
| Mutationen | `tools/harness/mutationslauf.sh`, `flock`, `MUTATIONSLAUF_SPEICHER=4g`, Image `odata-recorder:race`, Liste im Scratchpad (22 Mutationen, Last des Hosts 18 bis 20 bei 4 CPUs). Ergebnis: `rot=16 rot-paket=1 gruen=3 flatternd=2`. |
| Rohbytes (`AGENTS.md` §3.18) | Canary und Umlaut in `coverage_test.go`/`diagnose_test.go` entstehen aus `string([]byte{0xc3, 0xa4})`; die Fixtures nutzen keine Escape-Folge. |
| Plan folgt Code (`AGENTS.md` §3.9) | Tabelle, Mengen, Wege (17 in 5 Arten), Ergebnis-Absatz und Messung stimmen mit dem Code überein; Befund LOW 4 unten. |
| Abdeckung | `go test -race -covermode=atomic -coverpkg` im Container aus Baumkopie (`odata-recorder:race`, `CGO_ENABLED=1`, `--memory=4g`): `TestGleichzeitig` allein 94,1 % (`New` 60,0 %, alle anderen Funktionen 100 %), alle Tests 100,0 %; stimmt mit dem Plan. |
| Größenmarke | Messung 1725 gegen 3000 (Auslastung 0,58) wie berichtet. |

## 2. Findings

| # | Kategorie | Quelle | Pfad · Anker | Befund | verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| 1 | LOW | `AGENTS.md` §3.11 | `src/hexagon/services/inspection/inspektion.go` · `liefert die Schlüssel der Einstellungen, die die Diagnose zeigt, in der Folge von` | Der Kommentar von `Erlaubt` sagt die Folge von `SPEC-236` Nr. 7 zu; die Mutation `R-erlaubt-vertauscht` (die ersten zwei Schlüssel getauscht) bleibt grün, `TestErlaubtGleichWortlaut` vergleicht als Menge. | ja (Mutationslauf) | Kommentar-Zusage ohne Fänger |
| 2 | LOW | `AGENTS.md` §3.15 | `src/hexagon/services/inspection/coverage_test.go` · `strings.Repeat("a", 64)` | Die Kennungen von `TestCoverageKennung` reichen bis 64 Zeichen; die Mutation `R-id-laenge-64` (Kennung über 64 Zeichen ist `ErrUngueltig`) bleibt grün, eine längere Kennung führt kein Fall. | ja (Mutationslauf) | Fänger führt nur eine Längenform |
| 3 | LOW | `AGENTS.md` §3.15 | `spec/spezifikation.md` · `kein Schreiben in einen Port, kein Audit-Eintrag, kein Log, kein Zähler` (SPEC-236 Kopf; Plan Zeile N2) | Die Zusage „kein Zähler" des Vertragskopfs hat für einen Zähler im Dienst selbst (Kanal oder Paketvariable ohne Datenrennen und ohne Wirkung, Mutation `R-diag-kanalzaehler`) keinen Fänger, und N2 nennt dieses Negativ nicht mit Grund und Folge. | ja (Mutationslauf) | Zusage ohne Fänger, Negativ nicht benannt |
| 4 | LOW | `AGENTS.md` §3.9, §3.13 | `docs/plan/planning/done/slice-inspection-coverage-diagnose.md` · `Die Tests des Vorgängers bauen `Konfig` mit `Exchanges` allein` | §6 führt den Nachzug von Teil 1 weiter als *offen* (Adresse Implementer), obwohl er im Zweig getan ist (`konfigMit`, `neuerDienst`, 17 Wege); der Plan nennt die Zeile nicht erledigt. | nein | Plan folgt dem Code nicht |
| 5 | INFO | `AGENTS.md` §3.13 | Plan §6 · Abweichung 1 (`11 + n`) | Die Zahl ist nachgerechnet: fünf Zweige ohne Portfehler, fünf Kontextformen, n + 1 Fehler mit Code. Die Zählung gegen `EintraegeAbbildung()` ist eine Schleife, kein Festwert. Bestätigung gehört dem Architect; kein Fehler im Code. | nein | Abweichung mit Adresse |
| 6 | INFO | `SPEC-236` Nr. 7 | `diagnose_test.go` · `D8` / Plan §6 Abweichung 2 | `Fehler` ist durch `AlleCodes()` nie leer; das Literal `[]` ist unerreichbar, der Test hält `nil`-frei und JSON-Liste. Die Abweichung ist sachlich richtig, der Wortlaut „`{}` und `[]`" in Nr. 7 trifft nur `Einstellungen`; Entscheidung beim Architect. | nein | Vertragswortlaut ungenau |
| 7 | INFO | `AGENTS.md` §3.11 | `diagnose_test.go` · `ausgeschlossenV sind die zehn Namen der Nr. 7` | Der Kommentar nennt „zehn Namen" und zählt dann eine Liste von 19; die Zahl 19 ergibt sich als 10 + 9 Vertreter, nicht als die 18 der Aufzählung des Plans. Die Abweichung 3 ist sachlich unschädlich (mehr Vertreter, `ADMIN_*` mit zwei); die Zählprüfung `len(ausgeschlossenV()) * 3` hält die 19. | ja | Kommentar zählt nicht mit |
| 8 | INFO | `AGENTS.md` §3.15 | `coverage_test.go` · `tabelleSnapshot` | Abweichung 4 (Orakel `Snapshot` aus `tabelleSession` abgeleitet): `tabelleSession` ist in `port_test.go` handgeschrieben und teilt keinen Code mit dem Dienst, die drei geänderten Zeilen sind hier ausgeschrieben; das Orakel ist vom Dienst unabhängig. Es hängt aber an der Richtigkeit der 24 unveränderten Zeilen von Teil 1; die Mutation `C3-gegenstand-session` des Implementers belegt den Unterschied. | nein | Orakel abgeleitet |
| 9 | INFO | `AGENTS.md` §3.10 | Plan · *Nachweis der Bindung (K5)* | K5 ist als Übersetzungsmeldung nachgewiesen (`port_test.go`: `var _ inbound.InspectionUseCase`), nicht als Rot im Mutationslauf. Der Nachweis ist ehrlich benannt (Meldung, Mutationen, Wortlaut) und im Plan als Ausnahme begründet; ein Gate-Lauf zeigt ihn nur über `make test`. | ja (`make test` mit einer Mutation) | Nachweis über Übersetzungsmeldung |
| 10 | INFO | `AGENTS.md` §3.17 | `coverage_test.go` · `ms := []int{0, 1, 2, m101, m1001}` | Das Kürzen auf `MaxLimit` ist bis 1001 Einträge gefangen; ein Kürzen auf eine Zahl über 1001 (Menge unbeschränkt) bleibt ungefangen. Das ist die Prüfgrenze der Achse m; sie steht nicht ausdrücklich als Negativ in N2. | nein | Prüfgrenze nicht benannt |

## 3. Eigene Mutationen (Auszug, Klasse `andere`, sofern nicht anders genannt)

| Zusage | Mutation | Ergebnis |
|---|---|---|
| Reihenfolge von `Erlaubt` | erste zwei Schlüssel getauscht | gruen (Finding 1) |
| `Erlaubt` frisch je Aufruf | Slice als Paketvariable geteilt | rot (`TestErlaubtGleichWortlaut`) |
| Schlüssel der Allowlist | erster, mittlerer, letzter entfernt; `LISTEN` ergänzt | rot, rot, rot, rot |
| vier Modi | je einer entfernt (record, replay, mock), leerer Modus erlaubt | rot (`TestNewModus`) |
| Kontext Diagnose | `context.Cause` statt `ctx.Err()` | rot-paket (nicht im benannten Test; INFO, kein Befund) |
| `ctx` unverändert an `List` | `WithoutCancel`, `WithValue` | rot (`TestCoverageKontext`) |
| Evidenz byte-gleich | auf 16 Bytes gekürzt | rot |
| Zähler `uint64` | über 32 Bit abgeschnitten; auf `MaxInt64` gekappt | rot, rot |
| Zähler 0 steht | Nullzähler übersprungen | rot |
| Wert unverändert | auf 32 Bytes gekürzt | rot |
| Kennung unverändert | über 64 Zeichen abgewiesen | gruen (Finding 2) |
| kein Zähler im Dienst | Kanalzähler in `Diagnose` | gruen (Finding 3) |
| kein Zustand (`nebenläufig`) | Schreibzugriff auf Paketvariable am Anfang von `Coverage`; `defer` mit Schreibzugriff in `Diagnose` | `flatternd`, mit Detektor 6/6, ohne 0/6 (Last 18 bis 20, 4 CPUs; stimmt mit den 20/20 und 0/20 des Plans überein, kleinere Stichprobe) |

Die Schreibzugriffe am Anfang von `Coverage` und am Ende von `Diagnose` ergänzen die Stellen des Plans (Ende von `Coverage`, Anfang von `Diagnose`).
