# Review-Report: slice-mock-lesen (Teil 2) — 2026-11-07

**Review-Art:** Code — gegen Plan, Spezifikation (SPEC-195 bis SPEC-198, SPEC-109, SPEC-230) und Hard Rules

**Gegenstand:** Zweig `slice-mock-lesen`, HEAD `c9fb0a5` (Basis `473fe8f`/`158d70e`), Plan `slice-mock-lesen` (§3 Tabelle L01 bis L46, Mengen M4 bis M20, Raster, Absatz *Mutationen und Raten*, §6)

**Skill:** `.harness/skills/reviewer.md` im Zweig · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-07

**Eingangs-Kontext:**

- Plan `slice-mock-lesen` (in-progress), Spezifikation SPEC-195 bis SPEC-199, SPEC-109 Nr. 2 und 5, SPEC-230 Nr. 8, SPEC-190 Nr. 3
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Sprache), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) (keine `//nolint`)
- `LH-FA-06`, `LH-FA-11`, `LH-QA-05`
- `AGENTS.md` (Hard Rules, besonders 3.9, 3.10, 3.15, 3.16, 3.17, 3.18)

**Methode:** 76 eigene Mutationen mit `tools/harness/mutationslauf.sh` (Image `odata-recorder:race`, Kopie des Baums im Container, `flock`, `MUTATIONSLAUF_SPEICHER=4g`, Last des Hosts 7 bis 15, Messender der Reviewer): 65 der Klasse `andere` im Paket `mock` (davon 4 zu den Pflichten 3 e und 3 f), 3 im Paket `src/bootstrap`, 4 der Klasse `nebenläufig` (30 Prozesse mit Detektor, `-test.count=1`). Ergebnis: 64 rot oder rot-paket, 11 grün und 1 Übersetzungsfehler (m35, als n06 neu gefasst: rot). Die 11 grünen: sechs Befunde (m11, m30, m36, m19, n03, n13 = F-1 bis F-6), vier gleichwertige Mutationen oder ohne beobachtbare Wirkung (m18 = F-8, m40, n14, p03) und m12 (`Edm.Int64` mit 32 Bit: im Paket `mock` grün, in `src/bootstrap` rot, wie der Plan sagt). Die Listen liegen im Scratchpad des Reviewers, nicht im Repo.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Der Typ `N.Abstrakt` der Fixture hat keinen `BaseType`; `@odata.type: #N.Abstrakt` wird schon als fremder Typ abgewiesen. Die Mutation „abstrakter Typ erlaubt“ (`!ok \|\| t.Abstract` zu `!ok`) bleibt in Test und Paket grün (m11); die Zusage SPEC-196 Nr. 3 a („nicht abstrakt“) hat keinen Fänger. | `AGENTS.md` §3.15 (Fixture gleicht der Mutation) | `src/hexagon/services/mock/deuten.go` · `if !ok \|\| t.Abstract {`; `src/hexagon/services/mock/lesen_helfer_test.go` · `{Name: "N.Abstrakt", Abstract: true` | ja — Mutationslauf | Fänger führt den abstrakten Typ nicht isoliert (Fixture maskiert die Mutation) |
| F-2 | MEDIUM | Ein Payload mit einer Property, die der Typ kennt, aber nicht unterstützt (`Blob`, `Edm.Binary`, im Modell als `N.Seltsam`), ist nach SPEC-196 Nr. 3 b und d `ODR-E9001`. Die Mutation „`&& p.Unterstuetzt()` entfernt“ bleibt grün (m30): M8 führt nur die unbekannte Property, kein Fall trägt eine nicht unterstützte im Payload. | `AGENTS.md` §3.15 (eine Form der Eingabe) | `src/hexagon/services/mock/deuten.go` · `p.Name == m.name && p.Unterstuetzt()` | ja — Mutationslauf | Fänger führt nur eine Form der Eingabe (unbekannt statt nicht unterstützt) |
| F-3 | MEDIUM | „Typ des Sets nicht aufgelöst“ ist nach SPEC-196 Nr. 2 `ODR-E9001`; die Mutation auf `ODR-E6008` bleibt im ganzen Paket grün (m36). Der Plan (Absatz *Abdeckungsprofil*) nennt den Zweig unter den nicht gelaufenen Blöcken und behauptet, die Fehlerzweige seien „durch L13, L18, L40, L41, L43, L46 gefangen“; für diesen Zweig trifft das nicht zu. | `AGENTS.md` §3.15 (Zusage ohne Fänger); `AGENTS.md` §3.9 | `src/hexagon/services/mock/lesen.go` · `typ, ok := modell.EntityType(set.EntityType)`; Plan · „`lese` (Typ des Sets nicht aufgelöst)“ | ja — Mutationslauf | Zusage ohne Fänger; Plan behauptet einen Fänger, den es nicht gibt |
| F-4 | MEDIUM | Die Aufrufstelle `Blobs.Open` im Lesepfad (`ladeModell`, S4) hat keinen Fall für den Gegenstand: M18 führt „Stelle 7 (S5 bis S11)“, Teil 1 führt S4 nur für `$metadata` (anderer Code). Die Mutation `GegenstandSonstiges` zu `GegenstandWorkspace` an diesem `Open` bleibt grün (m19); die anderen Stellen (m15 bis m17, m20) werden rot. | `AGENTS.md` §3.15 (Aufrufstellen sind eine Menge, ein Fall je Stelle) | `src/hexagon/services/mock/lesen.go` · `strom, err := m.k.Blobs.Open(ctx, s.MetadataBlob)` | ja — Mutationslauf | Fänger nicht je Aufrufstelle |
| F-5 | MEDIUM | Das Raster nennt „Schlüsselname doppelt: SPEC-196 Nr. 7 a“ für den Lookup; die Mutation „`gesetzt[i]` entfernt“ (zwei Teile mit demselben Namen bei einem Typ mit zwei Schlüsseln) bleibt grün (n03). | `AGENTS.md` §3.16 (Zelle des Rasters ohne Fänger) | `src/hexagon/services/mock/schluessel.go` · `i < 0 \|\| gesetzt[i] \|\|` | ja — Mutationslauf | Zelle des Rasters ohne Fänger |
| F-6 | MEDIUM | Beim Lookup ohne `$select` zählen nur Properties mit `Unterstuetzt()` zu den angeforderten (SPEC-196 Nr. 7 e). Die Mutation „`if p.Unterstuetzt()` zu `if true`“ bleibt grün (n13): der Eintrag der Fixture `Seltsam` führt `Blob` in `KnownProperties`, so dass die Mutation `ODR-E6003` nie auslöst. | `AGENTS.md` §3.15 (Fixture gleicht der Mutation) | `src/hexagon/services/mock/lesen.go` · `if p.Unterstuetzt() {`; `src/hexagon/services/mock/lesen_fehler_test.go` · `eintrag("Seltsam", "1", ` | ja — Mutationslauf | Fixture maskiert die Mutation |
| F-7 | LOW | Zeile L25 der Tabelle nennt als Fänger `TestLookupOhneTreffer`; den Test gibt es nicht, L25 hält `TestCoverageMatrix` (Kopf der Funktion und Mutationstabelle sagen es). | `AGENTS.md` §3.9 (Plan folgt dem Code) | Plan · `L25` · `TestLookupOhneTreffer` | ja — Abgleich der Testnamen des Plans gegen `func Test` | Plan nennt Fänger mit falschem Namen |
| F-8 | LOW | Der Plan führt für L13 den Gegenstand „S14 bis S16“ als Fall je Aufrufstelle; `model.FuehreAus` liefert nie einen Fehler eines Ports, die Mutation `GegenstandSonstiges` zu `GegenstandWorkspace` an S14 ist gleichwertig und bleibt grün (m18). Weder Plan noch Test nennen das als akzeptiertes Negativ mit Grund und Folge. | `AGENTS.md` §3.15 (Negativ ohne Grund und Folge) | `src/hexagon/services/mock/lesen.go` · `return model.Ergebnis{}, abbilden(err, outbound.GegenstandSonstiges)` | nein — gleichwertige Mutation | Aussage über Fänger ohne Wirkung, Negativ nicht benannt |
| F-9 | INFO | Abweichung (b) des Handoffs (Pflichten 3 e und 3 f nicht im Dienst) ist begründet: SPEC-196 Nr. 4 c schreibt „`FuehreAus` liefert `ErrUngueltig`“, `pruefeEntities` in `model/abfrage.go` prüft Schlüssel und Key-Properties vor jeder Auswertung, eine zweite Prüfung wäre ein zweiter Ort. Die Fänger L09 und L10 greifen Mutationen des Dienstes (Dubletten zusammengeführt p01, leerer Schlüssel weggelassen p02: rot) und die Prüfung im Modell (p04: rot); p03 („Key-Property `null` wie fehlend“) ist gleichwertig, beide Formen ergeben `ODR-E9001`. | `AGENTS.md` §3.15 | `src/hexagon/services/mock/deuten.go` · `deuteListe` | ja — Mutationslauf | — |
| F-10 | INFO | `httpodata/fehler_test.go` ersetzt die feste Zahl 36 durch `len(model.AlleCodes()) - 2 + 1`. Das schwächt den Test nicht: die Rechnung hängt an der Struktur (zwei Codes nur mit Status 0, ein Code mit zwei Stati), und `TestCodeTabelle` hält 38 fest. Ein neuer Code, der nur den Status 0 trägt, macht den Test rot. | `AGENTS.md` §3.15 | `src/adapters/driving/httpodata/fehler_test.go` · `soll := len(model.AlleCodes()) - 2 + 1` | ja | — |
| F-11 | INFO | Das akzeptierte Negativ „nicht gelaufene Blöcke unter dem Detektor“ (82,2 % von `TestZweigeGleichzeitig` allein, 98,4 % des Pakets) trägt Grund und Folge und ist tragfähig, weil der Dienst keinen Zustand hält (`TestPaketOhneZustand` am Quelltext) und die Blöcke nur lokale Werte berühren. Der Lauf fängt Schreibzugriffe in `anzahl`, `ohneTreffer`, `liefere` und `deuteListe` je 30 von 30 mit Detektor (r01 bis r04); in einem nicht gelaufenen Block habe ich nicht mutiert, dort fängt allein der Quelltext-Test. Das Muster steht damit im Register `zweig-ohne-lauf-unter-dem-detektor` zum zehnten Mal. | `AGENTS.md` §3.10 | Plan · „Nicht gelaufene Blöcke einzeln“ | ja — Abdeckungsprofil unter `make test-race` | — |
| F-12 | INFO | Größenmarke: 3960 von 7300 (Auslastung 0,54, unter 0,60 nach SPEC-128 Nr. 9) steht im Plan mit Adresse; die Messung zählt die Zeilen des Merge von `main` mit, wie der Handoff sagt. Ich habe die Messung nicht neu gefahren. Die Entscheidung über die Neukalibrierung liegt beim User. | `AGENTS.md` §3.14 | Plan · `marke: 7300` | ja — `make groessenmarke` | — |

## Antworten auf die Schwerpunkte

- **Pflichten 3 e und 3 f im Dienst:** die Abweichung ist begründet (F-9); die Adresse „Reviewer“ in §6 ist damit beantwortet: der Dienst muss sie nicht sehen.
- **L11, unbekannte Property je Stelle:** ja, die Prüfung des Dienstes wird gefangen. Entfällt sie (`false && redigiertFrei`, m07), wird der Test rot; ebenso ohne `FilterNicht`, ohne rechte Seite des Vergleichs, ohne `$orderby` (m32 bis m34).
- **Coverage-Reihenfolge:** `ListSet` vor der strikten Prüfung (m01), Status ungeprüft (m02), Lookup `complete` und `partial` vertauscht (m03, m39), Tombstone ohne Treffer (m04): rot. Nicht jedes Paar der Rangfolge habe ich einzeln mutiert.
- **Fehlerabbildung:** Kontext je Aufrufstelle (n15 bis n24, alle Stellen: rot; n18 und n24 als rot-paket aus `TestModellLesen` und `TestCoverageLesen`, nicht aus `TestKontextLesen`); Gegenstand S6 bis S11 rot, S4 grün (F-4), S14 gleichwertig (F-8).
- **Fakes, `SchreibeFehler` nie gerufen:** der Zähler prüft in `t.Cleanup` jeden Test des Pakets; `TestKeinFehlerobjektLesen` führt die Schleife über die Fehler. Ohne Befund.
- **`ODR-E6008`:** in `model/fehler.go` an Position 38 hinter `ODR-E6007`; Status 410 statt 404 wird rot (m38, rot-paket aus `TestCodeTabelle`, nicht aus dem benannten Test); `harness/katalog-ausnahmen.txt` trägt den Grund (mock erreicht das Binary erst mit slice-bootstrap-m3). Alle Stellen mit 37 sind auf 38 oder auf `AlleCodes()` gezogen.
- **Zählprüfungen gegen Quellen:** 13 erreichbare und 25 übrige Codes gegen `AlleCodes()`, L13 mit `3 × (AlleAbfrageUrsachen + AlleFilterTypUrsachen)` und der Map der nicht erreichbaren Paare mit Grund: tragfähig. Die Gründe stimmen mit dem Code überein (`/$count` übergibt allein `Filter`; `$filter` an der Entity endet vor dem Interpreter in `ODR-E6001`; die Maske der Entity ist `nil`). Die Prüfung `erreichbar == (nicht != "")` verlangt jedes Paar genau einmal.
- **Nebenläufigkeit:** L36 mit 4 Mutationen des Reviewers, je 30 von 30 Prozessen mit Detektor, Last 10 bis 13 (r01 bis r04). L38 und L17 habe ich nicht neu gemessen.
- **Ende-zu-Ende L33 bis L35:** `Edm.Int64` mit 32 Bit, `$skip` nicht übergeben und `Wurzel` aus fremdem Host werden in `src/bootstrap` rot (e01 bis e03); die Nullprobe des Pakets lief bei Last 14 durch, die Zwei-Minuten-Grenze des Werkzeugs war an diesem Tag kein Hindernis (133 s für 3 Mutationen).
- **Rohbytes (§3.18):** die Dateien des Diffs tragen kein Muster aus Backslash, `u` und vier Hex-Ziffern; das Kelvin-Zeichen steht als Bytewerte (`"\xe2\x84\xaa"` in `lesen_ablauf_test.go`). Dieser Report enthält keine solche Folge. Keine `//nolint`; `.golangci.yml` und `.a-check.yml` unverändert.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/services/mock/lesen.go` (Ablauf, Coverage, Lookup, Antwort, Header) | geprüft; Befunde F-3, F-4, F-6, F-8 |
| `src/hexagon/services/mock/deuten.go` (Deuten, Maske, Pflichten 3 g) | geprüft; Befunde F-1, F-2 |
| `src/hexagon/services/mock/schluessel.go` | geprüft; Befund F-5 |
| `src/hexagon/services/mock/mock.go` (Übergang zu `lese`, `holeWorkspace`) | geprüft, ohne Befund |
| `src/hexagon/model` (`ODR-E6008`, `ressource_mengen.go`) | geprüft, ohne Befund |
| `src/adapters/driving/httpodata`, `src/adapters/driven/audit`, `harness/katalog-ausnahmen.txt` | geprüft, ohne Befund (F-10 nur Hinweis) |
| `src/bootstrap/mock_lesen_ende_zu_ende_test.go` | geprüft, ohne Befund |
| Fakes und Hilfen (`fakes_test.go`, `lesen_helfer_test.go`) | geprüft; Befunde F-1, F-6 betreffen die Fixture |
| Plan §3, §6 (Tabelle, Mengen, Raster, Absatz *Mutationen und Raten*) | geprüft; Befunde F-3, F-5, F-7, F-8 |
| `.golangci.yml`, `.a-check.yml`, `//nolint` | geprüft, ohne Befund |

## Summary

Kein HIGH. Die Fachlogik des Lesens folgt SPEC-196 bis SPEC-198 an allen Stellen, die ich mutiert habe (64 von 76 Mutationen rot, 11 grün, davon 6 als Befund; dazu die Raten der Gleichzeitigkeit 30 von 30). Sechs MEDIUM-Befunde (F-1 bis F-6) sind Zusagen oder Zellen ohne Fänger, jeweils mit einer grünen Mutation als Beleg; zwei LOW betreffen den Plan (F-7, F-8). Die Abweichung der Pflichten 3 e und 3 f ist begründet (F-9). Die Klassen der MEDIUM-Befunde sind „Fixture oder Fänger führt eine Form der Eingabe“ (F-1, F-2, F-6) und „Aufrufstelle oder Zelle ohne Fall“ (F-3 bis F-5); beide sind nach `AGENTS.md` §3.17 gegen die Klasse aufzulösen, nicht gegen die genannten Mutationen.
