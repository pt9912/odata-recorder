# Verifikations-Report: slice-bootstrap-m4-konfig — 2026-11-18

**Art:** Verifikation (Modul 11: gegen DoD, Spezifikation und Entscheidungen; `AGENTS.md` §3.15, §3.17), in frischem Kontext, ohne Änderung am Produktcode

**Gegenstand:** Slice `docs/plan/planning/done/slice-bootstrap-m4-konfig.md` (Teil 1 von `slice-bootstrap-m4`), Zweig `slice-bootstrap-m4`, Stand `77915118` (Fix-Commits Runde 4: `49c29b8a`, `77915118`); Reports Runde 1 bis 4 unter `docs/reviews/2026-11-18-slice-bootstrap-m4-konfig-review*.md`

**Maßstab:** [LH-FA-12](../../spec/lastenheft.md), [LH-FA-14](../../spec/lastenheft.md), [LH-QA-02](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `SPEC-110` Nr. 1, 2, 6, 8, 9, 12, 14, 15, 16, `SPEC-261` Nr. 4 e, 8, 10 b, `SPEC-139`, `SPEC-037`, `SPEC-115`, `SPEC-128` Nr. 1, 4, 5; `AGENTS.md` §3.7, §3.9 bis §3.18

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

## Verdikt: PASS

Die technische DoD (Punkte 1 bis 4 und der Teil „Doku-Update“) ist belegt; `make gates` ist auf dem Stand grün, und die Stichproben-Mutationen an Elementen, die weder Review noch Implementer nannten, werden 24 von 24 rot. Kein HIGH, kein MEDIUM. Offen sind die Formalien der Closure (Häkchen, Closure-Notiz, Register, Ausgänge der vier offenen Punkte in §6); sie gehören nach `AGENTS.md` §3.3 vor den Übergang nach `done/` und liegen beim Implementer und beim Planner (Übergaben unten).

## Eigene Läufe

| Lauf | Ergebnis |
|---|---|
| `make docs-check` vor dem Report | 906 Dateien, 0 Befunde |
| `make gates` im Worktree (unter `flock` auf `.wt/heavy.lock`, Last zu Beginn 5,4, keine Gegenproben, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) | Exit 0; im Lauf: `a-check`, `docs-check`, `build`, `test` (`pruefung test`), `test-race`, `groessenmarke` (`slice-bootstrap-m4-konfig.md: ok, Messung 1872 von Marke 900`), `handbuch-check`, `katalog-check`, `lint-profil`, `lint`, `zeichen-check` (ok, 1896 Dateien); Log im Scratchpad (`verif-gates.log`), nicht im Repo |
| Größenmarke von Hand: `git diff --numstat 7a5170d9..HEAD -- src/bootstrap/` | 1872 (Summe aus Einfügungen und Löschungen), davon `prefer_ende_zu_ende_test.go` 143 |
| Rohbytes gegengelesen (`LC_ALL=C grep -P` auf U+0085, U+00A0, U+2000 bis U+202F, U+3000, U+FEFF, U+180E, U+205F, U+1680) | 0 Treffer in `m4_konfig_klassen_test.go`, `m4_konfig_test.go`, `konfig_werte.go`, `konfig_hybrid.go`, `export_test.go`, dem Plan, dem Handbuch (`AGENTS.md` §3.18) |
| Direktive `//nolint` in `src/bootstrap/*.go` | keine (`AGENTS.md` §3.2) |
| Fänger der Tabelle (§3) im Quelltext | alle 31 genannten Tests (`TestHybrid*`, `TestKonfig*`, `TestUpstream*`, `TestKappe*`) existieren; keine Zeile nennt einen Test, den es nicht gibt |

## Beleg je DoD-Punkt

| Nr. | DoD-Punkt | Urteil | Beleg |
|---|---|---|---|
| 1 | Lesen und Prüfen der vier Schlüssel; `*KonfigFehler` der Policy wird `ODR-E2001` ohne Wert; `UPSTREAM_URL` Pflicht bei `enabled`; die Funktion, die die Policy baut | erfüllt | `src/bootstrap/konfig_hybrid.go`: `liesHybrid` (vier Schlüssel, `policy.New`, `meldeKonfigFehler` über den ganzen Join, Grund nennt nie den Wert), `Konfiguration.NeuePolicy` (Rückgabe `*policy.Policy`, Fehler `*KonfigFehler`, Policy bei Fehler `nil`); `upstreamPflicht` in `konfig_dienste.go` (derselbe Text `Pflichtfeld fehlt` in `record` und `hybrid`, nur bei gültigem `enabled`); Fänger M04 bis M06 grün im Gate; Mutationen unten |
| 2 | `hybrid` gültiger Wert von `MODE`; Zusammenhänge, `PUBLIC_BASE_URL`, `STRICT_COVERAGE`, `ETAG_ENABLED` in `hybrid`; `ADMIN_*` und `LISTEN_TLS_*` bleiben `ODR-E2001` | erfüllt | `liesModus` ohne Sonderfall für `hybrid`; `nichtGelieferte` nennt allein `ADMIN_` und `LISTEN_TLS_`; `liesPublicBaseURL` mit `mock || hybrid`, Grund `im Modus mock oder hybrid bei diesem Host erforderlich` (`SPEC-110` Nr. 14); Fänger M07 bis M09 (`TestHybridKonfigMockSchluessel`, `…JederModus`, `…Familien` 4 × 6, `…Modus`); der Spezifikations-Nachzug liegt in `SPEC-110` Nr. 2, 8, 9, 12, 14, 15, 16 (Architect) |
| 3 | Handbuch: vier Zeilen in der Tabelle der Einstellungen, `make katalog-check` grün | erfüllt | `docs/user/benutzerhandbuch.md`: die vier Zeilen und die Zeile `MODE` nennen Wertemenge, Form, Default und den Ist-Zustand (geprüft, sonst nicht verwendet; Entscheidung (g)); `make handbuch-check` und `make katalog-check` laufen im grünen `make gates` |
| 4 | `make gates` grün | erfüllt | eigener Lauf, Exit 0 (Tabelle oben) |
| 5 | Review durchgeführt, Report unter `docs/reviews/` | erfüllt | vier Reports (Runden 1 bis 4); die Befunde der Runde 4 (3 MEDIUM, 2 LOW, 4 INFO) sind in `49c29b8a`/`77915118` behoben und unten gegen die Klasse aufgelöst |
| 6 | Doku-Update, falls ein öffentlicher Vertrag berührt ist | erfüllt | `SPEC-110`, `SPEC-139` Nr. 12, `SPEC-261` Nr. 8 b, 10 b, `SPEC-037` (Architect) und Handbuch im Zweig; `docs-check` und `katalog-check` grün |
| 7 | Closure-Notiz mit Steering-Loop-Lerneintrag | offen | §7 trägt „Wird bei der Closure gefüllt“; Folge der Reihenfolge Review, Verifikation, Closure, kein Fehler des Standes (Übergabe 1) |
| 8 | Beobachtungs-Register fortgeschrieben | offen | kein Eintrag unter `docs/plan/planning/observations/` nennt diesen Slice; auch „keine Beobachtung“ steht noch nicht in §7 (Übergabe 1) |
| 9 | Jedes Risiko aus §6 trägt einen Ausgang | erfüllt mit vier *offen*-Zeilen | jede Zeile trägt einen Ausgang; *offen*/*weiter offen* mit Adresse tragen vier (Reihenfolge zu `slice-bootstrap-m6`, Nullprobe `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`, Datenpunkt `SPEC-128` Nr. 2, `policy.parseMuster`), siehe Übergabe 2 |
| 10 | Drei Paarungen (Anker, Folge-Slice, Register) | offen | Anker: keine Regel mit `liegt in` im Slice; Folge-Slice: `slice-bootstrap-m4` trägt alle Sendungen mit Kennung des Gebers (`done/slice-bootstrap-m4.md` Zeilen 38, 42, 67, 69, 95, gelesen); Register: siehe Nr. 8 |

## Zusage · Fänger gegen den Vertrag (§3.15, §3.16)

Tabelle M04 bis M09 und die Zusatztabelle (§6) gegen `SPEC-110` Nr. 1, 2, 6, 8, 9, 12, 14 bis 16 und `SPEC-261` Nr. 8, 10 b gelesen: Jede Zusage hat eine Zeile mit Fänger und Mutation; jede Zelle des Rasters (3 Operationen × 5 Argumentarten, 15 Zellen) nennt einen Ort oder „gilt nicht“ mit Grund; keine Zelle ist im Code entschieden. Die Rückgaben des Implementers (Handbuch-Zeilen, Feld `Fallback`, `NeuePolicy`, Nicht-ASCII-Host) stehen mit Ausgang und Adresse in §6 und sind in `done/slice-bootstrap-m4.md` angenommen (`AGENTS.md` §3.13). Eine Zeile ohne Fänger fand ich nicht.

## Fixes der Runde 4 gegen die Klasse (§3.17)

Mutationen eines eigenen Treibers im Scratchpad (`r2drv`; Container ohne Netz, `--memory=4g --memory-swap=4g --cpus=4`, `flock`, Baum per `docker cp`, der Arbeitsbaum blieb unverändert, `git status` sauber; Last zu Beginn der Läufe 2,0 bis 6,1, erster Wert von `/proc/loadavg`, Messender der Treiber). `tools/harness/mutationslauf.sh` nicht benutzt (Nullprobe von `src/bootstrap` über 2 Minuten, Plan §6). Nullprobe grün. Klasse `andere`: ein deterministischer Prozess je Mutation, Rate 1 von 1. Die Liste liegt im Scratchpad (`liste-verif.json`), nicht im Repo. **24 Mutationen, alle an Elementen, die die Tabellen des Plans nicht nennen: 24 von 24 rot**; je erster roter Test aus der Familie der Zusage.

| Klasse (Zusage der Runde 4) | Menge, aus dem Wortlaut gebildet | Mutation an einem nicht genannten Element | Ergebnis (roter Test) |
|---|---|---|---|
| F-1 Position und Eigenschaft jedes Eintrags | Position 1 bis 17 × acht Listen × drei Eigenschaften (408) | Leerraum nur an geraden Positionen gekappt (`liste`); Form nur an geraden Positionen geprüft (`pruefeEintraege`); `none` ohne den letzten Eintrag gesucht; die drei Listen von `hybrid` nur an geraden Positionen gekappt; die Folge der Einträge von `hybrid` vertauscht (erster gegen letzten) | rot (5 von 5): `TestKonfigListenErgaenzen`, `TestKonfigListePositionBisN`, `TestHybridKonfigSyntax` und weitere der Familie |
| F-2 Gegenseite der Zeichen, Header-Token | 179 Bytes × drei Listen | zusätzlich erlaubt: `:`, `<`, `\`, `[` (der Plan nannte `(`, `/`, `"`, DEL, 0xA0, `;`, `@`, 0xFF, NUL) | rot (4 von 4) |
| F-2 Gegenseite der Zeichen, JSON-Name | 192 Bytes | zusätzlich erlaubt: `.`, `$`, Leerzeichen (der Plan nannte `"`, `/`, DEL, 0xA0, `:`, `;`, `@`, 0xFF, NUL) | rot (3 von 3) |
| F-3 Verstoß je Prüfgruppe in jedem Modus | 6 Modi × 8 Gruppen (48) | `PUBLIC_BASE_URL` auch im Modus `record` Pflicht; `UPSTREAM_URL` auch im Modus `replay` Pflicht; `LISTEN_TLS_` aus der Ablehnung genommen; `ADMIN_` aus der Ablehnung genommen | rot (4 von 4), `TestKonfigNichtGeliefert` an der Stelle der Zusage; bei `ADMIN_` ohne Ablehnung startet ein Test der Familie einen echten Listener und hängt bis zur Zeitgrenze von `go test` (601,6 s), der Fänger meldet rot, bevor er hängt |
| F-5 Stelle im Namen und Form des Query-Namens | 89 erlaubte Zeichen, 5 verbotene, Länge bis 4096, Stellen | `;` und `?` zusätzlich verboten (beide gehören zu den 89); Prüfung nur über die ersten 1000 Bytes (die Prüfgrenze des Plans nennt `i == 1000`, nicht die Obergrenze 1000) | rot (3 von 3) |
| `kappe` jedes andere Zeichen bleibt (F-5 der Runde 3, Runde 4 gegengelesen) | 1112060 Codepunkte, 128 Rohbytes | `leerraum` um U+0000 und um U+001F erweitert | rot (2 von 2): `TestKappeLaesstJedesAndereZeichenStehen` |
| M04 Weitergabe und Meldung | je Wert der Policy, je Ursache | `ReplayKategorien` und `Schreibfreigaben` in `Fallback` verworfen; Trenner der Meldung `Eintrag %d; %s` | rot (3 von 3) |

Gegen die Prüfgrenzen des Plans: Listen über 17 Einträge und mehrbytige ungültige Folgen bleiben die benannten Negative (§6, Grund und Folge im Betrieb stehen dort); die Obergrenze 1000 der Länge fängt der Fänger an den Stellen 4095 und 4096, das benannte Negativ `i == 1000` bleibt davon unberührt.

## Plan-vs-Code-Diff

| Plan sagt | Code/Stand | Urteil |
|---|---|---|
| §3: `konfig*.go` neu `konfig_hybrid.go`, `upstreamPflicht`, `PUBLIC_BASE_URL` in `hybrid`, `NeuePolicy`; `wurzel.go` unberührt | Diff stimmt: `konfig.go`, `konfig_basis.go`, `konfig_dienste.go`, `konfig_hybrid.go`, `konfig_werte.go`; `wurzel.go` nicht im Diff, `vorbereiten` lehnt `hybrid` weiter ab | erfüllt |
| §1 *Ausdrücklich NICHT*: keine Bindung, `ADMIN_*`/`LISTEN_TLS_*` bleiben abgelehnt, `LINK_REWRITE` bleibt `off` | kein Verstoß im Diff; `nichtGelieferte` hält `ADMIN_`, `LISTEN_TLS_` | erfüllt |
| §3: `tests/e2e/` nimmt `ADMIN_LISTEN` statt `MODE=hybrid` | `tests/e2e/stufe_prozess_test.go` geändert wie beschrieben; `SPEC-037` Nr. 5, 7 führen `ADMIN_LISTEN` | erfüllt |
| Kommentare (§3.7, §3.11) | Köpfe von `liesHybrid`, `eintraege`, `upstreamPflicht`, `istQueryName`, `kappe` nennen Zusage, Rang-Zeiger oder Abgrenzung im Indikativ; die Köpfe der Fänger nennen ihre Prüfgrenze (Runde 4, F-5); kein Konjunktiv über verworfene Alternativen im Produktcode gefunden | erfüllt |
| §4 Größenmarke: Marke 900 vom Planner vor dem Code gesetzt, Block mit `ueberschreitung: 1872` | `make groessenmarke`: `ok, Messung 1872 von Marke 900`; eigene Messung 1872 | erfüllt |

## Größenmarke (SPEC-128 Nr. 1, 4, 5; `AGENTS.md` §3.14)

- Messung 1872 stimmt mit dem Gate und mit der Hand-Messung überein; die Marke 900 (Formel: *S* 630, *M* = 100 · ⌈135 · 630 / 10000⌉) ist unverändert und stimmt mit der Formel (Nr. 5 b).
- **143 fremde Zeilen** (`prefer_ende_zu_ende_test.go`, 143 Einfügungen) kommen mit dem Merge von `main` (`b47430cb`, `7c888289`, Slice `slice-prefer-return`) in die Messung; `SPEC-128` Nr. 4 g und Nr. 10 g zählen sie ausdrücklich mit, Nr. 4 g nennt als Wege `ueberschreitung` mit Grund oder neue Basis mit `ersetzt`. Der Eintrag `ueberschreitung: 1872 | …` nennt die 143 und den Grund (Nr. 5 d: *N* ≥ Messung, 1872 ≥ 1872; ein nicht leerer Grund, Nr. 5 e). **Keine Anhebung der Marke**, keine Basis-Änderung; das Verfahren entspricht dem Vertrag. Ohne die fremden Zeilen wären es 1729 (Planner-Entscheidung deckt Runde 4, §6).
- Schnitt oder Rückführung: nach `AGENTS.md` §3.14 nicht erlaubt, weil kein vorab benannter Teil existiert (Konstante, Funktion und Fänger teilen den Mechanismus); der Grund steht in §4 und im Eintrag; die Entscheidung des Planners (§6) deckt den Mehrumfang bis 1872. Die älteren Zahlen (1402, 1500, 1709) stehen nur noch als Messpunkte, nicht als Zusage (der Befund F-4 der Runde 4 ist behoben).
- Der Datenpunkt Ende ÷ Handoff = 2,58 (1729 ÷ 671) gegen die Spanne 1,19 bis 1,33 von `SPEC-128` Nr. 2 steht als *offen*, Adresse Planner (Übergabe 2).

## Befunde

| ID | Kategorie | Befund | Quelle | Pfad |
|---|---|---|---|---|
| V-1 | INFO | Closure-Formalien stehen noch aus: die Häkchen in §2, §7 (Closure-Notiz, Lerneintrag, Register-Antwort), die drei Paarungen. Die Reihenfolge verlangt sie nach der Verifikation. | DoD Punkte 7, 8, 10 | `docs/plan/planning/done/slice-bootstrap-m4-konfig.md` · `*Wird bei der Closure gefüllt.*` |
| V-2 | INFO | `SPEC-037` Nr. 5 sagt „`MODE=hybrid` startet“ und Nr. 7 führt `TestE2EHybrid` als vereinbart, ohne Test; das Binary lehnt `hybrid` bis Teil 2 ab (§6 (f), akzeptiertes Negativ, dort nur für `SPEC-110` und `SPEC-139` benannt). Kein Fänger fängt die Abweichung zwischen den Teilen, bis `slice-bootstrap-m4` (Zeile M01, M4-Test) sie schließt; beide Slices gehen nacheinander in denselben Zweig. | `SPEC-037` Nr. 5, 7; §6 (f) | `harness/vertraege/SPEC-037-test-integration.md` · `MODE=hybrid` startet |
| V-3 | INFO | Mit der Mutation „`ADMIN_` nicht abgelehnt“ startet ein Test der Familie `TestKonfig*` einen echten Listener und hält den Lauf bis zur Zeitgrenze von `go test` (601,6 s). Der Fänger der Zusage (`TestKonfigNichtGeliefert`) wird vorher rot, der Lauf ist nur langsam. Keine Wirkung auf das Gate (`make test` fährt den unveränderten Code). | `AGENTS.md` §3.10 | `src/bootstrap/konfig_test.go` · `TestKonfigNichtGeliefert` |

Keine HIGH-, keine MEDIUM-, keine LOW-Befunde.

## Offen und Übergaben

1. **Implementer/Planner vor `done/`** (`AGENTS.md` §3.3, Ablauf Inhalt, dann reiner `git mv`): Häkchen setzen, Closure-Notiz mit Lerneintrag in §7, Register-Antwort (Eintrag oder „keine Beobachtung“, Zähler wird nicht gesetzt), `liegt in` nur falls etwas verkörpert wurde.
2. **Planner** (Adressen stehen in §6, Ausgang *offen*): Reihenfolge zu `slice-bootstrap-m6`; Nullprobe `tools/harness/mutationslauf.sh` für Pakete über 2 Minuten (`BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`); Datenpunkt 2,58 für `SPEC-128` Nr. 2; Fänger je Position der Segmente in `policy.parseMuster` (Folge-Slice der Policy, `slice-fallback-policy` liegt in `done/`).
3. **Architect, nur bei bekannter Betreiber-Anforderung:** Namen mit Leerzeichen, Nicht-ASCII oder `%` in `REDACT_QUERY_PARAMS`; Nicht-ASCII-Host in `UPSTREAM_URL` ohne passenden Eintrag von `FALLBACK_ALLOW_HOSTS` (beide fail-closed, akzeptierte Negative in §6).
4. **`slice-bootstrap-m4` (Teil 2):** ruft `Konfiguration.NeuePolicy`, trägt `Fallback` in die Tabelle der Verbraucher von `TestWerteErreichenVerbraucher` ein, zieht die Handbuch-Zeilen `PUBLIC_BASE_URL`, `STRICT_COVERAGE`, `ETAG_ENABLED` und die vier Zeilen der Schlüssel auf den lieferbaren Modus nach; die Sendungen sind dort angenommen.

## Negativbefunde (geprüft, ohne Befund)

`src/bootstrap/konfig.go`, `konfig_basis.go`, `konfig_dienste.go`, `konfig_hybrid.go`, `konfig_werte.go` (Produktcode, 175 Zeilen); `src/bootstrap/export_test.go`; Handbuch (`MODE` und die vier Zeilen, `REDACT_QUERY_PARAMS`); Plan §1, §3, §4, §6 (Kopf, Tabellen, Raster, Entscheidungen (a) bis (g), Sendungen); `done/slice-bootstrap-m4.md` (Annahme der Sendungen); Rohbytes der Tests und des Plans; keine Direktive `//nolint`.
