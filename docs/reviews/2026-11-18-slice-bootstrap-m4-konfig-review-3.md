# Review-Report: slice-bootstrap-m4-konfig — 2026-11-18 (Runde 3)

**Review-Art:** Code (gegen Plan, `SPEC-110` Nr. 1, 6, 9, 16, `SPEC-115` Nr. 4, `SPEC-261`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und `AGENTS.md` §3); Prüfung der Fixes von Runde 2 gegen die Klasse der Zusage (`AGENTS.md` §3.17)

**Gegenstand:** Diff `6da21441..HEAD` (Zweig `slice-bootstrap-m4`; Fix-Commit `fb889da5`, Merge `b47430cb` von `origin/main`, Plan-Commit `f1fdaf5a`), ohne die von `main` hereingemergten Dateien fremder Slices (`prefer*`, `rueckgabe*`, Spezifikationsnummern `SPEC-285` bis `SPEC-289`); Gesamtstand von `src/bootstrap/konfig*.go`, `m4_konfig_test.go`, `SPEC-110` Nr. 1, 6, 9, 16 und Plan `slice-bootstrap-m4-konfig`

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext:**

- Reports Runde 1 und 2 (`docs/reviews/2026-11-18-slice-bootstrap-m4-konfig-review.md`, `…-review-2.md`)
- Slice-Plan `slice-bootstrap-m4-konfig` (§1, §3, §4 Größenmarke, §6 Raster und Entscheidungen)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-12`, `LH-FA-14`
- `SPEC-110` Nr. 1, 6, 9, 16; `SPEC-115` Nr. 4; `SPEC-128` Nr. 1, 5; `SPEC-201` Nr. 2 e; `SPEC-261`
- `AGENTS.md` (Hard Rules)

**Messung dieses Laufs.** Eigener Container-Treiber im Scratchpad (Image `odata-recorder:test`, ohne Netz, `--memory=4g --cpus=4`, unter `flock` auf `.wt/heavy.lock`, Last zu Beginn der gültigen Läufe 9,95 bis 12,23, erster Wert von `/proc/loadavg`, Messender der Treiber); `tools/harness/mutationslauf.sh` wurde nicht benutzt (Nullprobe von `src/bootstrap` über 2 Minuten, Register `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`). Je Mutation ein deterministischer Prozess an einer Kopie im Container (der Arbeitsbaum blieb unverändert), alle Tests der Familien `TestKonfig*`, `TestHybrid*` und `TestUpstream*` (2 bis 3 s je Lauf); ein Rot zählt nur mit der Meldung eines Tests an der Stelle der Zusage. Rate je Mutation 1 von 1, kein Fänger dieses Diffs ist wahrscheinlichkeitsbehaftet, keine Zusage betrifft Nebenläufigkeit.
Nicht gezählt: ein erster Lauf mit 20 Mutationen, dessen Treiber das Muster der Tests falsch verankerte (alle 20 „grün“, kein Test lief), und drei Mutationen, die auf einem leeren Slice mit `panic` endeten (Rot, aber nicht aus dem Grund der Zusage; sie wurden mit Grenzprüfung neu gefahren). Gezählt sind 23 Mutationen: 12 rot, 11 grün; keine der grünen ist äquivalent, jede ändert das Verhalten für einen Wert der Klasse, den kein Test führt (Zeile *Wert, den kein Test führt* der Tabelle unten). Dazu eine Probe ohne Fänger zu F-1 (`go test -run` auf einer Kopie, Datei im Scratchpad, nicht im Repo).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | HIGH | Ein Eintrag von `REDACT_QUERY_PARAMS` mit gültiger Prozent-Kodierung (`my%20token`, `a%41`, `tok%2Ben`) besteht aus den 90 Zeichen von `istQueryName` und startet ohne Verstoß, trifft aber den Parameter nicht, den ein Client so schreibt: `RedigiereQuery` dekodiert den Schlüssel der Anfrage nach `%XX` und vergleicht dann exakt (`SPEC-115` Nr. 4). Probe: Name `my%20token`, Query `my%20token=GEHEIM` bleibt unverändert (geändert: falsch); Name `a%41`, Query `a%41=GEHEIM` bleibt unverändert, nur `a%2541=GEHEIM` wird redigiert; Name `tok%2Ben`, Query `tok%2Ben=GEHEIM` bleibt unverändert. Der Wert geht unredigiert ins Recording. Das ist die Klasse von F-1 der Runde 2 („ein Name träfe nie einen Parameter“, `SPEC-110` Nr. 6) mit einem Element, das die Zeichenmenge nicht trennt, weil `%` selbst zu den 90 Zeichen gehört; in `SPEC-110` Nr. 6 und im Raster von §6 steht keine Zelle für eine Escape-Folge im Namen, der Code entscheidet sie implizit (`AGENTS.md` §3.16). | `LH-FA-14`, `SPEC-110` Nr. 6, `SPEC-115` Nr. 4, `AGENTS.md` §3.16, §3.17 | `src/bootstrap/konfig_werte.go` · `func istQueryName(s string) bool` | ja — Probe dieses Laufs (Test auf Kopie, 5 Namen × 8 Queries) | Fail-open: akzeptierter Query-Name trifft die Schreibweise der Anfrage nie (Escape-Folge) |
| F-2 | MEDIUM | Die Zusage „jeder Eintrag einer Liste hat die Form seines Namens“ hat keinen Fänger für die Position des ungültigen Eintrags in der Liste: `pruefeEintraege` nur über den ersten Eintrag (`e[:1]`) oder nur über den letzten bleibt grün, nur der Ausschluss des ersten (`e[1:]`) wird rot (`TestKonfigListe`, ein ungültiger erster Eintrag). Die Fänger der Klasse „Name in `REDACT_QUERY_PARAMS`“ (`TestKonfigQueryNameInListe`, `TestKonfigListenFremderLeerraum`) setzen das Zeichen in alle drei Einträge zugleich, so dass jede Teilprüfung rot würde und doch keine der beiden Teilmengen unterscheidet. Szenario: `REDACT_QUERY_PARAMS=token,bad name` oder `REDACT_QUERY_PARAMS=bad name,token` startet bei einer Prüfung nur des ersten bzw. des letzten Eintrags; der ungültige Name trifft nie. | `AGENTS.md` §3.15 (ein Fall je Element einer Menge), §3.17 | `src/bootstrap/konfig_werte.go` · `func (p *pruefer) pruefeEintraege` | ja — Mutationen v-eintraege-nur-erster, v-eintraege-nur-letzter dieses Laufs | Fänger führt nur eine Form der Eingabe (Position des Verstoßes in der Liste) |
| F-3 | MEDIUM | Die Zusage „die Syntax aller Schlüssel wird in jedem Modus geprüft“ (`SPEC-110` Nr. 8) hat für die Einträge der Listen `REDACT_*`, `MATCH_VARY_HEADERS` und `VOLATILE_HEADERS` nur einen Fänger im Modus `record`: die Prüfung der Einträge (`ergaenzendeListe`) entfällt bei `MODE=replay`, `mock`, `hybrid` oder ungesetztem `MODE` (4 Mutationen), die der Header-Token bei `replay` (1): alle fünf grün; entfällt sie bei `record`, wird `TestKonfigListe` rot. Die Schleifen von Runde 2 (`KonfigQueryNameInListe`, `KonfigListenFremderLeerraum`, `KonfigWeissraumAusserhalbBleibt`) laufen mit `hybridBasis(t, "record")`; die Schleife über vier Modi (`TestHybridKonfigJederModus`) führt nur die vier Schlüssel von `hybrid`. Szenario: ein Start mit `MODE=replay` und `REDACT_QUERY_PARAMS=bad name` wird angenommen, der Name geht in die Regeln des Match-Keys und trifft nie. | `AGENTS.md` §3.15, §3.17 (Achse Modus × Zeichen), `SPEC-110` Nr. 8 | `src/bootstrap/konfig_werte.go` · `func (p *pruefer) ergaenzendeListe` | ja — Mutationen v-liste-pruefung-nicht-in-replay, -in-mock, -in-hybrid, -ohne-modus, v-header-token-nicht-in-replay | Fänger führt nur einen Modus der Klasse Syntax × Modus |
| F-4 | MEDIUM | Die Position eines Zeichens im Namen ist eine unbeschränkte Menge, der Fänger führt nur Namen bis 5 Bytes (`a<Zeichen>b`, `p_a`, Anfang, Mitte, Ende eines kurzen Namens), und im Plan steht keine Prüfgrenze mit Grund und Folge: `istQueryName` mit einer Schleife nur über die ersten 32 Bytes, nur über die ersten 256 Bytes oder mit wachsender Schrittweite (`i += 1 + i/8`) bleibt grün. Szenario: ein Name, dessen ungültiges Zeichen hinter einer Prüfgrenze steht, wird angenommen und trifft nie. | `AGENTS.md` §3.15, §3.17 (Menge unbeschränkt: Prüfgrenze als akzeptiertes Negativ), Skill MEDIUM *Fänger schmaler als Zusage* | `src/bootstrap/konfig_werte.go` · `func istQueryName(s string) bool` | ja — Mutationen v-name-nur-32-bytes, v-name-nur-256-bytes, v-name-nur-jedes-zweite-byte | Menge Position im Namen ohne Fänger über die Länge und ohne benannte Prüfgrenze |
| F-5 | LOW | Das Negativ „21 Zeichen mit White_Space außerhalb des Leerraums bleiben im Eintrag“ deckt die endliche Menge, aber `SPEC-110` Nr. 1 sagt „kein anderes Zeichen“; die Menge der übrigen Zeichen ist unbeschränkt, und der Plan nennt dafür keine Prüfgrenze. `kappe` um U+180E erweitert (das Zeichen, das `SPEC-201` Nr. 2 e ausdrücklich als nicht zum Leerraum gehörend nennt) bleibt grün; U+0085, U+00A0, U+200B, U+FEFF, LF und Tabulator in `kappe` werden rot. | `AGENTS.md` §3.17 (akzeptiertes Negativ mit Grund und Folge) | `docs/plan/planning/done/slice-bootstrap-m4-konfig.md` · `Eine Erweiterung von kappe um eines dieser Zeichen wird damit bei unveränderter Konstante leerraum rot` | ja — Mutation v-kappe-mit-u180e | Unbeschränkte Restmenge ohne benannte Prüfgrenze |
| F-6 | LOW | Der Plan folgt dem Marken-Commit nicht in allen Stellen (`AGENTS.md` §3.9): §4 *Messpunkt nach dem Review der Runde 2* nennt „Eintrag `ueberschreitung: 1259`“ und „§6, Ausgang *offen*“, der Block trägt seit `f1fdaf5a` 1500 mit der Messung 1402 (143 Zeilen fremd), und §6 trägt *entschieden (Planner, 2026-11-18)*. Zudem benennt §6 kein Artefakt der Entscheidung des Planners: der Ausgang wechselt im selben Commit wie der Eintrag von *offen* auf *entschieden*, und das vorab benannte Rückführungs-Kriterium (über 900) ist eingetreten (`AGENTS.md` §3.13: kein Pfeil ohne benennbares Artefakt). Gemessen und gedeckt ist die Zahl: `git diff --numstat 7a5170d9..HEAD -- src/bootstrap/` ergibt 1402, davon 143 in `prefer_ende_zu_ende_test.go`; 1500 deckt (`SPEC-128` Nr. 5 d). | `AGENTS.md` §3.9, §3.13, §3.14 | `docs/plan/planning/done/slice-bootstrap-m4-konfig.md` · `Eintrag ueberschreitung: 1259` trägt den Grund | nein — kein Gate liest den Satz | Plan-Satz hinter dem Block zurück; Entscheidung ohne benanntes Artefakt |
| F-7 | INFO | Die Zeile `REDACT_QUERY_PARAMS` des Handbuchs nennt keine Form des Namens; eine ungültige Eintragung endet mit `Eintrag hat eine ungültige Form` ohne Eintrag und ohne Regel (die Meldung nennt nie den Wert). Die Form mit 90 Zeichen ist seit Runde 2 neu und für den Betreiber nur in der Spezifikation lesbar. Das Handbuch liegt nicht im Diff; Adresse: Verifier und Teil 2 (Handbuch-Kapitel). | `SPEC-200` Nr. 1 | `docs/user/benutzerhandbuch.md` · `Zusätzliche Namen von Query-Parametern` | nein | Neue Eingabeform ohne Zeile im Handbuch |
| F-8 | INFO | Die Fixes von Runde 2 decken ihre Klassen, soweit eigene Mutationen an Elementen, die weder Review noch Bericht nannten, sie prüfen: Zeichen im Namen (`;` und `%` als verboten: rot, `TestKonfigQueryNameZeichen`), Zeichen des Leerraums in `kappe` (ohne LF, ohne Tabulator, mit U+0085, U+00A0, U+200B, U+FEFF: rot), Leerraum am Listeneintrag ohne LF (rot, `TestKonfigLeerraumJedeStelle`), `menge` ohne `Compact` (rot, `TestKonfigListenErgaenzen`). Kein Befund; Hinweis für die Verifikation. | `AGENTS.md` §3.17 | `src/bootstrap/m4_konfig_test.go` · `func TestKonfigQueryNameZeichen` | ja | — |

## Mutationen dieses Laufs (Klasse, nicht die genannten Beispiele; §3.17)

23 Mutationen, Klasse `andere` (deterministisch, 1 von 1 je Mutation); 12 rot (aus dem Grund der Zusage, Meldung des Tests in der Spalte), 11 grün.

| Zusage | Mutation (Name der Liste) | Ergebnis |
|---|---|---|
| Form jedes Eintrags einer Liste, Position | `pruefeEintraege` nur über `e[:1]` (v-eintraege-nur-erster), nur über den letzten (v-eintraege-nur-letzter) | **grün** (F-2) |
| dasselbe | ohne den ersten Eintrag (v-eintraege-nicht-erster) | rot: `TestKonfigListe` (`MATCH_VARY_HEADERS=Bad Header`) |
| Syntax in jedem Modus | Prüfung der Einträge entfällt bei `MODE=replay` (v-liste-pruefung-nicht-in-replay), `mock`, `hybrid`, ungesetzt; Header-Token bei `replay` | **grün** (F-3), fünf Mutationen |
| dasselbe | Prüfung der Einträge entfällt bei `MODE=record` (v-liste-pruefung-nicht-in-record) | rot: `TestKonfigListe` (`REDACT_HEADERS=Bad Header`) |
| Query-Name: jedes Zeichen außerhalb der 90 verboten, jedes innerhalb erlaubt | `;` verboten (v-name-zeichen-semikolon-verboten), `%` verboten (v-name-prozent-verboten) | rot: `TestKonfigQueryNameZeichen` (`Name ";": gültig false, erwartet true`; `Name "%"`) |
| Query-Name: jede Position im Namen | Schleife nur über 32 Bytes, nur über 256 Bytes, Schrittweite `1 + i/8` | **grün** (F-4), drei Mutationen; *Wert, den kein Test führt:* ein Name mit mehr als 8 Bytes |
| Leerraum sind genau vier Zeichen | `leerraum` ohne LF (v-kappe-ohne-lf), ohne Tabulator (v-kappe-ohne-tab) | rot: `TestKonfigServicesFile`, `TestKonfigLeer` |
| dasselbe, Verhalten statt Konstante | `kappe` um U+0085, U+00A0, U+200B, U+FEFF erweitert | rot: `TestKonfigListenFremderLeerraum` (drei), `TestKonfigQueryNameInListe` (U+200B) |
| dasselbe | `kappe` um U+180E erweitert (v-kappe-mit-u180e) | **grün** (F-5) |
| Leerraum am Listeneintrag | `liste`: `Trim(e, " \t\r")` statt `kappe` (ohne LF) | rot: `TestKonfigLeerraumJedeStelle` |
| Doppelte haben keine Wirkung | `menge` ohne `Compact` (v-menge-ohne-compact) | rot: `TestKonfigListenErgaenzen` |
| Probe zu F-1 (kein Fänger) | Namen `my%20token`, `a%41`, `tok%2Ben`, `a%zz`, `t+k` gegen acht Queries | `my%20token`, `a%41` (außer Doppel-Kodierung `a%2541`), `tok%2Ben` treffen die übliche Schreibweise nie; `a%zz` und `t+k` treffen ihre eigene Schreibweise |

Die Rate je Mutation ist 1 von 1 (ein Prozess, deterministisch, kein Detektor nötig).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/konfig.go` (`leerraum`, `kappe`, `wert`) | geprüft, ohne Befund am Code; Kommentare tragen Zusage oder Kopplung (Rang-Zeiger auf `SPEC-110` Nr. 1, 6); Fänger siehe F-5 |
| `src/bootstrap/konfig_dienste.go` (`upstreamPflicht`, `upstreamURL`, `dateiEintrag`) | geprüft, ohne Befund; derselbe Grund `Pflichtfeld fehlt` in `record` und `hybrid`, `kappe` an den drei Stellen |
| `src/bootstrap/konfig_hybrid.go`, `konfig_werte.go` (`liste`, `eintraege`, `nichtGeliefert`) | geprüft; Befunde F-1 bis F-4 an `istQueryName`, `pruefeEintraege` und `ergaenzendeListe`; keine verworfene Alternative im Kommentar; kein weiteres `TrimSpace` im Produktcode von `src/bootstrap` |
| Aufrufer von `istQueryName` und `RedigiereQuery` (`model/schluessel.go`, `model/redaction.go`, `record/ende.go`, `wurzel.go`) | geprüft: eine Quelle für die Namen (`Konfiguration.RedactQueryParams`), keine zweite Prüfung eines Namens; der Vergleich dekodiert den Schlüssel der Anfrage, Befund F-1 |
| Header-, JSON-Pfad-, `SERVICE_ID`-, Root-Prüfung (`istHeaderToken`, `ParseJSONPfad`, `zeichenAus`) | geprüft, ohne Befund: byteweise ASCII, kein Unicode-Buchstabe (`unicode.*` nicht im Produktcode), jedes fremde Zeichen ist ein Verstoß |
| Politik der drei Listen von `hybrid` (`policy/host.go`, `policy/freigabe.go`) | nur gelesen, nicht im Diff: byteweise Formprüfung, ein fremdes Zeichen am Rand bleibt im Eintrag und scheitert dort |
| `src/bootstrap/export_test.go`, `m3_konfig_test.go`, `konfig_test.go` | geprüft, ohne Befund (Textanpassung `Pflichtfeld fehlt`, `Leerraum` und `IstQueryName` als Zugang) |
| `src/bootstrap/m4_konfig_test.go` | geprüft; Köpfe nennen ihre Menge und die Zählung (279 gültig und 171 Verstoß von 450; 151 Zeichen, 90 und 38 und 61; 76 und 19), die Mengen stimmen mit der Rechnung; Werte aus Bytewerten und Zahlenwerten, keine Folge Backslash-u im Quelltext; Befunde F-2 bis F-5 |
| `spec/spezifikation.md` (`SPEC-110` Nr. 1, 6, 9, 16 f, Historie) | geprüft gegen Code und Plan: 90 Zeichen, vier Zeichen des Leerraums und derselbe Grund der Pflicht stimmen; die 25 White_Space-Zeichen und die 21 außerhalb stimmen mit `SPEC-201` Nr. 2 e; Historienzeile für Nr. 1 und 6 liegt vor; Lücke F-1 (Escape-Folge) |
| Plan `slice-bootstrap-m4-konfig` (Kopf, §1, §3, §4, §6) | geprüft: Kopf, §1 und §3 tragen die Nachzüge der Runde 2; Raster §6 ohne leere Zelle, Zusatztabelle führt die Zeile für den Query-Namen; Befunde F-1 (Zelle fehlt), F-4, F-5, F-6 |
| `docs/plan/planning/done/slice-bootstrap-m4.md` (Teil 2), Adressen §3.13 | geprüft, ohne Befund: Sendungen mit Kennung des Gebers, Nehmer liegt in `open/`, §1 *Ausdrücklich NICHT* trifft sie nicht |
| Größenmarke | Marke 900 gleich Formel (`SPEC-128` Nr. 1); Messung 1402 reproduziert (1259 eigene, 143 aus `prefer_ende_zu_ende_test.go`); Eintrag 1500 deckt; Befund F-6 nur zu Prosa und Artefakt |
| `harness/`, `.harness/skills/`, Hooks | nicht im Diff; keine Unterdrückung (`//nolint`) im Diff gefunden |
| Gate `make gates` | nicht Gegenstand dieses Reviews (Verifier) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 1 |
| MEDIUM | 3 |
| LOW | 2 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fail-open: akzeptierter Query-Name trifft die Schreibweise der Anfrage nie (Escape-Folge) · Fänger führt nur eine Form der Eingabe (Position des Verstoßes in der Liste; Modus) · Menge Position im Namen ohne Fänger über die Länge und ohne benannte Prüfgrenze · Unbeschränkte Restmenge ohne benannte Prüfgrenze · Plan-Satz hinter dem Block zurück; Entscheidung ohne benanntes Artefakt · Neue Eingabeform ohne Zeile im Handbuch

## Verdikt

**Merge-blockierend:** ja — ein HIGH (F-1) und drei MEDIUM (F-2, F-3, F-4). Die Fixes von Runde 2 decken die genannten Klassen (Zeichenmenge, Zeichen des Leerraums je Stelle, Grund der Pflicht); die neuen Befunde liegen an Elementen, die die Menge nicht trennt (Escape-Folge im Namen) oder die kein Fänger führt (Position in der Liste, Modus, Länge des Namens).

**Übergabe:** F-1 an den Architect (Rang 2: Frage ist, ob ein Name mit `%` und zwei Hex-Ziffern, den der Vergleich nach der Dekodierung nie trifft, in `SPEC-110` Nr. 6 zulässig ist; die Zelle fehlt im Raster, die Entscheidung stünde vor dem Code, `AGENTS.md` §3.12). Ein Widerspruch zum Entscheid der Runde 2 (die 90 Zeichen nennen `%` ausdrücklich) ist möglich; ein HIGH wird nicht herabgestuft, der Konflikt läuft als Sequenz über den Architect mit diesem Report und der Probe als Übergabe-Artefakt. F-2 bis F-4 an den Implementer (Klasse: Position × Liste, Modus × Zeichen, Länge × Zeichen, mit je einer Mutation an einem Element, das dieser Report nicht nannte, oder einer benannten Prüfgrenze mit Grund und Folge, `AGENTS.md` §3.17); F-5, F-6 an den Implementer und den Planner (Artefakt der Entscheidung zur Marke); F-7, F-8 zur Kenntnis. Die Finding-Klassen gehen in die Slice-Closure §7. Der Report ersetzt keine Verifikation.
