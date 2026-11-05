# Review-Report: slice-v2-payload-lesen — 2026-11-05

**Review-Art:** Code — geprüft gegen Plan (`slice-v2-payload-lesen` §1, §3 mit der Tabelle *Zusage · Fänger* (57 Zeilen), §4 mit Block, Marke 8100 und Basis `20b0d1a`, §6 mit Raster und Zusatztabellen Z1 bis Z3), gegen `SPEC-172` Nr. 1 bis 5 mit E1 bis E6, `SPEC-126`, `SPEC-170`, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) und `AGENTS.md` §3.

**Gegenstand:** `slice-v2-payload-lesen` (Welle `welle-odata-v2`), Zweig `slice-v2-payload-lesen`, HEAD `d314da6` (Zwischenstand-Commits `5b11f3f` und `d314da6` über dem Merge `9650388`). Pfade: `src/adapters/driven/odata/jsoncodecv2/` (Produktcode, 1365 Zeilen in zehn Dateien), `src/hexagon/ports/outbound/outboundtest/odata_codec_v2_*.go` (neun Dateien) mit zwölf Dateien unter `testdata/`.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `d314da6` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-05

**Methode.** Einmaliger Lauf, kein Bericht des Implementers übernommen. Der Produktcode ist vollständig gelesen (alle zehn Dateien), die Suite stichprobenartig (Vertrag, Modelle, Nebenläufigkeit). `make gates` einmal im Worktree gefahren; `make a-check`, `make zeichen-check` und `make handbuch-check` danach einzeln (die Kette brach bei `lint` ab). Mutationen laufen auf einer Kopie von `src/` im Scratchpad, im Image `odata-recorder:test` (`--network none --memory=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`), je Mutante eine Ersetzung, `go test -count=1` über das Paket `jsoncodecv2`; im Worktree liegt nichts außer diesem Report. Es sind 13 Mutanten gefahren, kein Probe-Paket. Die Raten von G20/G21 (Nebenläufigkeit) habe ich **nicht** gemessen (siehe INFO-3).

**Eingangs-Kontext:**

- Slice-Plan `slice-v2-payload-lesen` in `d314da6`
- `SPEC-170`, `SPEC-172` Nr. 1 bis 5 mit E1 bis E6, `SPEC-126`, `SPEC-128`
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md)
- `LH-FA-17`, `LH-RB-01`
- `AGENTS.md` (Hard Rules §3.2, §3.7, §3.9 bis §3.18)

---

## Findings

| Nr. | kategorie | quelle | pfad | befund | verifizierbar | klasse |
|---|---|---|---|---|---|---|
| 1 | MEDIUM | Maintainability; `AGENTS.md` §3.6 (Gate bleibt, wird nicht gelockert) | `src/hexagon/ports/outbound/outboundtest/odata_codec_v2_vertrag.go` · `func v2G18(t *testing.T, s *jsuite) {` und weitere Stellen (siehe Befund) | `make lint` ist rot (Exit 1), die Kette bricht dort ab. Gemeldet sind: `cyclop` und `gocyclo` für `v2G18` (20, Grenze 15), `gocognit` für `v2G6` (29) und `v2FaelleG9Bytes` (22, Grenze 20), `dupl` in vier Paaren (`odata_codec_kodierung.go:163-180` gegen `odata_codec_v2_lesen.go:175-192`, `odata_codec_strom.go:219-244` gegen `odata_codec_v2_geruest.go:144-169`, `odata_codec_werte.go:173-184` gegen `odata_codec_v2_werte.go:141-152`) und `unparam` in `odata_codec_kodierung.go:183` (`b3Alphabet`: `bis` erhält immer 0xFF, weil der neue Aufrufer denselben Wert übergibt). Eine `//nolint`-Direktive ist nicht gesetzt (geprüft: keine im Diff). `make test`, `make test-race`, `make groessenmarke`, `make a-check`, `make zeichen-check` und `make handbuch-check` sind grün; `make lint-profil` und die Gates hinter `lint` in der Kette habe ich nicht einzeln gefahren. | ja (`make lint`) | Gate rot im Handoff; V2-Suite dupliziert Hilfen der V4-Suite |
| 2 | MEDIUM | `AGENTS.md` §3.10, §3.9, §3.13; Plan §3 *Reihenfolge für den Implementer* Nr. 4 | `docs/plan/planning/done/slice-v2-payload-lesen.md` · `**Tabelle *Zusage · Fänger***` | `git diff 9650388..HEAD -- docs spec` ist leer: die beiden Commits tragen nichts im Plan. Es fehlen der Absatz in §3 *Zusage · Mutation · roter Test* mit den Raten von G20/G21 (Zahl der Prozesse, Messender), die Messung in §4 (Größenmarke) und die Ausgänge der in §6 offenen Risiken. Reviewer und Verifier haben damit keine Tabelle der Implementer-Mutationen und keine Rate, gegen die sie messen; die Commit-Nachrichten nennen beide Stände „Zwischenstand“, der Handoff-Commit steht aus. | ja (Diff des Plans) | Plan folgt dem Code nicht (Handoff unvollständig) |
| 3 | LOW | Maintainability; `AGENTS.md` §3.15 (Zusage ohne Fänger) | `src/adapters/driven/odata/jsoncodecv2/entity.go` · `besucht[t.typ.Name] = true` | `abgeleitet` schützt die Suche entlang `BaseType` gegen einen Kreis. Entfernt man die Markierung, bleibt `TestVertrag` grün (Mutante „abgeleitet ohne Besucht“); weder Plan noch Spezifikation nennen die Zusage, und `model` prüft keinen Kreis (kein Treffer für Zyklus in `src/hexagon/model` und `csdlv2`). Ob ein Modell mit Kreis den Leser erreichen kann, habe ich nicht geprüft. | ja (Mutation) | Schutzcode ohne Zusage und ohne Fänger |
| 4 | INFO | Maintainability | `src/adapters/driven/odata/jsoncodecv2/zahl.go` · `int64(len(z.ziffern))+z.exp > 19` | Die Mutante `> 20` bleibt grün, ist aber äquivalent: die Grenze verhindert nur das Ausschreiben langer Zeichenketten, das Ergebnis entscheidet `strconv.ParseInt`. Kein Finding. | nein | äquivalente Mutante |
| 5 | INFO | `SPEC-128` Nr. 5 | `docs/plan/planning/done/slice-v2-payload-lesen.md` · `marke: 8100` | Gemessen (Summe aus Einfügungen und Löschungen, `git diff --numstat 20b0d1a..HEAD` über `jsoncodecv2` und `outboundtest`): 7323, unter der Marke 8100, aber 1,23 mal *S* (5970) und damit über dem erwarteten Bereich 0,8 bis 1,2 des Plans. Die Spanne gilt für das Ende; bei Closure ist ein Grund zu führen, falls die Messung ihn verlangt. Zur Messung: `20b0d1a` ist Basis, die Pfade stammen aus dem Block. | ja (`make groessenmarke`, grün) | Größenmarke |
| 6 | INFO | `AGENTS.md` §3.10 | `src/hexagon/ports/outbound/outboundtest/odata_codec_v2_vertrag.go` · `func v2G20(t *testing.T, s *jsuite) {` | G20 und G21 sind in jeder meiner Mutanten als zusätzlich rot gemeldet (G21 läuft über die ganze Suite); ob sie eine Mutation *nur* der Nebenläufigkeit (Paketvariable, geteilter Puffer, gemeinsamer Fehlerzeiger) in jedem Prozess fangen, ist **nicht** gemessen. Der Leser hat kein Feld (G19 grün) und der Code hat keine Paketvariable (gelesen). Die Messung (30 Prozesse je Mutation mit `make test-race`) gehört in Finding 2 und in die Verifikation. | nein | Rate nicht gemessen |
| 7 | INFO | [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) | Merge nach `main` | `git merge-tree --write-tree HEAD main` meldet keinen Konflikt; `main` hat `slice-v2-query` und `slice-m2-referenzdienst` gemergt (77 Dateien, davon keine unter `jsoncodecv2`). Der Plan nennt den Merge-Konflikt in `spec/spezifikation.md` als offen: er tritt hier nicht auf. | ja | Konfliktpunkte |
| 8 | INFO | Maintainability | `src/adapters/driven/odata/jsoncodecv2/leser.go` · `func vorbereiten` | Der Leser ist Zustandslos (`type Leser struct{}`), der Scanner und der Deuter gehören einem Aufruf; nirgends eine Paketvariable. Die Prüffolge folgt Z1 (Argumente, Version, Kontext, `entity_set`), die Zusage der Zerlegung des Stroms (G7, G12) steht in `lies` und wird von den Mutanten unten gefangen. | ja | — |

## Mutationen (stichprobenartig, durch den Reviewer)

Je Mutante eine Ersetzung im Produktcode, Treffer der Suite benannt (alle in `TestVertrag`, sofern nicht anders angegeben). Das Lauf-Setup ist in *Methode* genannt; zwei Mutanten bauten zuerst nicht und sind neu gefahren (Skriptfehler des Reviewers), sie zählen nicht doppelt.

| Zusage | Mutation | Ergebnis |
|---|---|---|
| G12 (nie mehr als `grenze+1` angefordert) | `want = int(rest) + 2` | rot: G8 (Strom gehört dem Aufrufer), weitere Zeilen nicht aufgelistet |
| G12 (`zu_gross` am Byte `grenze+1`) | `ueber := int64(n) > rest+1` | rot: G8, G12, G16, G21 |
| G10 (UTF-8 streng) | Führungsbyte-Bereich `0xE1..0xEF` zu `0xE1..0xEE` | rot: G10 |
| W8 (Obergrenze) | `ms > zeitMax+1` | rot: G21 (die Aufzählung der Zeilen war auf sechs gekürzt; W8 selbst ist nicht bestätigt) |
| W8 (Bruchteil ohne Nullen) | `TrimRight` entfernt | rot: W8, W9 |
| W6 (Grenze 1024) | `laenge >= maxDezimalLaenge` | rot: W6, G21 |
| H6 (`__count` nicht negativ) | Prüfung `v < 0` entfernt | rot: H6, G21 |
| H7 (`__next` leer ist `wert`) | Prüfung entfernt | rot: H7, G21 |
| N4 (`Edm.Date` übersprungen) | Ausnahme entfernt | rot: N4, G21 |
| K2 (Schlüssel `null`) | Prüfung entfernt | rot: K2, G21 |
| K2 (Schlüssel mit U+0000) | Prüfung entfernt | rot: K2, G21 |
| G13 (doppelter Name über 16 Namen) | Mengenprüfung `return false` | rot: G13, G21 |
| W3 (`Edm.Int32` als String ist `typ`) | `alsString` wahr | rot: W3, W1, W11, H10, G15, G21 |
| G5 (Kontext nach dem letzten Read) | zweite Kontextprüfung entfernt | rot: G5 |
| N1 (abgeleiteter Typ, Kreis) | `besucht` nicht markiert | **grün** (Finding 3) |
| Ganzzahl-Vorprüfung | `> 19` zu `> 20` | grün, äquivalent (Finding 4) |

## Einordnung gegen Plan, ADR und Hard Rules

- **[ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) (versionierter Adapter, keine Vermischung):** `jsoncodecv2` hat einen eigenen Scanner und importiert kein V4-Paket (`versionstrennung_test.go` grün, J4: `git diff --numstat --diff-filter=MDRT 20b0d1a..HEAD` über `jsoncodec`, `csdl`, `urlparser`, `src/hexagon/model`, `src/hexagon/ports/outbound` ist leer). Kein Verstoß.
- **Randformen im Code entschieden (`AGENTS.md` §3.16):** nicht gefunden; `huelle`, `mitResults`, `metadata`, `property` und `schluessel` folgen den Ersatzwortlauten E1 bis E5 (gelesen, nicht gegen jeden Wortlaut der Spezifikation Zeile für Zeile verglichen).
- **Kommentare (`AGENTS.md` §3.7):** im Produktcode stichprobenweise gelesen, keiner beschreibt eine verworfene Alternative oder abwesenden Text. Die Treffer des Suchmusters „früher“ in der Suite sind Namen von Fällen (Rangfolge), keine Kommentare.
- **Suppression (`AGENTS.md` §3.2):** keine `//nolint`-Direktive im Diff; `.golangci.yml`, `.a-check.yml`, `Dockerfile` und `harness/` unberührt.
- **Geheimnisse und kritischer Pfad:** der Leser liest Antworten für den Import; er schreibt kein Log, keine Senke und keinen Wert in den Fehlertext (`Error()` mit dem Ort in `strconv.Quote`-Schreibweise laut G15; nicht durch eine Mutante geprüft).
- **§3.18 (Rohbytes):** die V2-Datumswerte entstehen in der Suite aus Bytewerten (`v2Dt` mit `\x5c`); die Echtdateien sind gegen die SHA-256 des Plans verglichen (Stichprobe: Datei 14, Datei 24 und `v4_fremd_net-21` stimmen; die übrigen neun nicht einzeln verglichen, ihre Größen stimmen mit dem Plan).

## geprüft, ohne Befund

- `src/adapters/driven/odata/jsoncodecv2/` (Scanner, Baum, Deuter, Zahl, Zeit, Guid, Leser): gelesen, ohne Befund über Finding 3 und 4 hinaus.
- `src/adapters/driven/odata/jsoncodec/`, `csdl/`, `csdlv2/`, `urlparser/`: unverändert (J4, `git diff` leer).
- `src/hexagon/model/`, `src/hexagon/ports/outbound/` (außer `outboundtest/`): unverändert.
- `src/hexagon/ports/outbound/outboundtest/odata_codec_v2_*.go`: Findings 1 und 6; sonst stichprobenartig (Vertrag, Modelle, G20/G21, Zählprüfung der Zeit) ohne Befund.
- `src/hexagon/ports/outbound/outboundtest/testdata/`: Größen und drei SHA-256 geprüft, ohne Befund.
- `spec/`, `harness/`, `.golangci.yml`, `.a-check.yml`, `Dockerfile`: im Diff nicht berührt.
