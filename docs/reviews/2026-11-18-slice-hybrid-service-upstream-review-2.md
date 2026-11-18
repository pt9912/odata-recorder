# Review-Report: slice-hybrid-service-upstream (Teil 2: Upstream-Fallback), zweite Runde — 2026-11-18

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** `slice-hybrid-service-upstream`, Zweig `slice-hybrid-service-upstream`, Commits seit dem ersten Review (`c6637c66`): `244827e8` (`SPEC-284` Nr. 11, `SPEC-121` Nr. 11), `80a26331` (Fänger für F-1 bis F-5, F-7, F-8, Plan nachgezogen) und `54f75aea` (nur Roadmap). Gelesen: `git diff main...HEAD` im Ganzen, `git diff c6637c66..HEAD` im Einzelnen.

**Skill:** `.harness/skills/reviewer.md` @ `54f75aea` <!-- d-check:ignore (Zitat des Standes im Lauf) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext:** Slice-Plan §1, §3, §6; `SPEC-284` Nr. 3, 4, 5, 8, 9, 11, 13; `SPEC-121` Nr. 11; `SPEC-120` Nr. 8; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-12`; `AGENTS.md` §3.9 bis §3.18; der Report der ersten Runde.

**Vorgehen und Belege.**

1. Je Befund der ersten Runde die Menge aus dem Wortlaut der Spezifikation gebildet (nicht aus dem Befund) und 29 eigene Mutationen an Elementen gefahren, die weder der Befund noch der Plan nannten (`tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, `--network none`, `--memory` 2g, `--cpus=4`, je Aufruf ein Container, unter `flock /home/builder/work/odata-recorder/.wt/heavy.lock`). Zwei Listen: 23 Mutationen (Namen `q01` bis `q23`, davon `q20` ohne Wirkung und nicht gewertet) und 6 Mutationen (`w01` bis `w06`). Messender: der Reviewer; Last des Hosts laut Ausgabe des Werkzeugs 5,9 bis 6,2 von 4 (Liste 1) und 21,9 bis 23,1 von 4 (Liste 2, andere Läufe auf dem Host).
2. Nebenläufigkeit (F-5): `n_detektor` und `n_ohne` je 20, das heißt je Mutation 20 Prozesse mit Detektor (`-race`, `CGO_ENABLED=1`) und 20 ohne; Prozesse werden einzeln gezählt (`-test.count=1`).
3. `make gates` einmal vollständig (siehe unten).
4. Rohbytes (`AGENTS.md` §3.18): `LC_ALL=C grep -nP '\\u[0-9A-Fa-f]{4}'` über `src/hexagon/services/hybrid/*.go`, `src/hexagon/services/record/ohneantwort_test.go` und den Plan: kein Treffer. Die Kelvin-, NBSP-, EM-SPACE- und NEL-Werte in `upstream_klassen_test.go` entstehen aus `bytesVon(0x..)`. Dieser Report ist nach dem Schreiben auf demselben Weg gegengelesen.

---

## Ergebnis von `make gates`

Gefahren als `make -k gates` unter der Sperre (damit ein rotes Gate die übrigen nicht verdeckt); der Lauf begann auf `80a26331` plus der Roadmap-Fassung davor, `docs-check` danach einzeln auf `54f75aea` wiederholt.

| Gate | Ergebnis |
|---|---|
| `a-check` | grün, 0 Befunde (Hinweis: 94 Dateien unter `tools/` und weitere liegen in keiner Schicht, bekannt) |
| `baseline-verify` | grün, `v6.18.1`, 54 Dateien |
| `docs-check` | im Lauf rot, 4 Befunde, alle in `docs/plan/planning/in-progress/roadmap.md` (bekannt, nicht gemeldet); auf `54f75aea` nach dem Roadmap-Commit: grün, 0 Befunde |
| `schema-check` | grün (alle Zeilen `ok`, `foreign_key_check`, `integrity_check`) |
| `migration-check` | grün (`migration-check: ok`) |
| `build` | grün (Stufe `build`, Schicht aus dem Cache: die Eingabe ist byte-gleich dem Lauf des Implementers) |
| `test` | grün (`pruefung test /src`, Schicht aus dem Cache, wie `build`) |
| `test-race` | grün (`pruefung race /src`, Schicht aus dem Cache); unabhängig davon fuhr die Nullprobe von `mutationslauf.sh` das ganze Paket `hybrid` und `record` ohne Mutation grün |
| `groessenmarke` | grün mit Hinweis „Marke steht aus: Schätzung 2220, Marke 3000“ (bekannt, Plan §6, Adresse Planner) |
| `handbuch-check` | grün |
| `zeichen-check` | grün (1841 Dateien) |
| `katalog-check` | grün |
| `lint-profil`, `lint` | grün |

Hinweis zum Cache: dass `build`, `test` und `test-race` aus dem Schicht-Cache kamen, heißt gleiche Eingabe, nicht einen frischen Lauf. Die frischen Läufe von Test und Detektor liefert die Mutationsliste (Nullprobe, Stufe 2).

---

## Auflösung der Befunde der ersten Runde gegen die Klasse (`AGENTS.md` §3.17)

| Befund | Menge (aus dem Wortlaut) | Eigene Mutationen an Elementen, die der Befund nicht nannte | Ergebnis |
|---|---|---|---|
| F-1 Leser für U | Aufrufstellen von U × Methoden × Deklaration der `Laenge` (Nr. 5 c, 8 b) | `q01` `Laenge` des Puffers durch `a.Laenge` ersetzt; `q02` U ohne Body erhält den Leser der Anfrage; `q03` R teilt den Leser mit M; `q04` U über der Grenze ohne Body | `q01`, `q02` rot (`TestUpstreamLeserJederWeg`, `TestUpstreamArgumente`); `q03` rot im Paket (`TestPufferung`, Teil 1, der benannte Test bleibt grün); `q04` grün und äquivalent: `fallback` ist über der Grenze falsch, U läuft nie. **Geschlossen.** |
| F-2 Root `/` | Listen × Pfade ohne führenden `/` (Nr. 4) | `q05` Prüfung `pfad != "*"`, `q06` Prüfung `pfad != "" && pfad != "odata"`, `q07` Rest leer bleibt leer | alle rot (`TestDienstwahlKeinDienst`, `TestDienstwahl`). **Geschlossen.** |
| F-3 `UpstreamURL` leer | gemischte Listen × Pfade × Reihenfolge | `q09` `Ziel` ohne Schlusszeichen `/`, `q10` `Ziel` klein | rot (`TestZielAnStellen`). **Geschlossen.** |
| F-4 Kontext beendet vor U | Aufrufstellen von U × Art des Endes | `q13` Prüfung nur auf `DeadlineExceeded`, `q15` Kontextfehler erst nach `ErrUpstreamZeit` geprüft, `q14` `context.Cause(ctx)` statt `ctx.Err()` | `q13`, `q15` rot; `q14` grün (F-14). **Geschlossen bis auf F-14 (INFO).** |
| F-5 Nebenläufigkeit, Weg U | Zugriffe auf den Empfänger: Feld, Index, Inkrement, Alias, Aufruf | `w02` `v.dienste[0].Root = ziel`, `w03` `v.maxBytes++` in `Hybrid`, `w04` Alias `d := v.dienste; d[0].UpstreamURL = ziel`, `w06` `copy(v.dienste, …)`, `q22` und `w05` `slices.Reverse(v.dienste)` in U und in `waehle` | `w02`, `w03` rot in `TestQuelltext`; `w04` und `w06` rot 20 von 20 mit Detektor und 20 von 20 ohne (nur durch `TestGleichzeitigUpstream`); `q22` mit Detektor 20 von 20, ohne 14 von 20; `w05` mit Detektor 20 von 20, ohne 12 von 20 (Ergebnis `flatternd` des Werkzeugs, weil die Mutation ohne Detektor nur in einem Teil der Läufe rot wird). **Geschlossen:** die Zusage gilt je Prozess als gefangen, wo der Detektor läuft (20 von 20); ohne Detektor nur eine Größenordnung (12 bis 14 von 20, Last 5,9 bis 23). Der Plan nennt dieselbe Rate für Alias (20 von 20, 20 von 20) und nennt den Detektor je Prozess; die Zahlen decken sich mit meiner Messung. |
| F-6 Rangfolge `ErrUpstreamZeit` | beide Fehler des Ports × beide Reihenfolgen des `errors.Join` × beide Dienste (`hybrid`, `record`) | `q16` (`hybrid`) und `q21` (`record`): `ErrUpstreamZeit` nur ohne `ErrUpstreamNichtErreichbar` | beide rot (`TestUpstreamFehler`, `TestOhneAntwort`). Die Spezifikation trägt die Entscheidung in `SPEC-284` Nr. 11 und `SPEC-121` Nr. 11; der Wortlaut stimmt mit `SPEC-120` Nr. 8 (Frist vor „jeder andere Fehler“) überein. **Geschlossen.** |
| F-7 Rand | Randzeichen × Seite; Namen mit Faltung | `q17` Trenner `;` innerhalb eines Tokens; `q18`, `q19` Namen `Upgrade`, `Proxy-Authenticate` aus der Liste entfernt | `q18`, `q19` rot; `q17` grün (F-13). **Offen als F-13.** |
| F-8 Zahl im Plan | `git diff --numstat 75249f56..HEAD` über `src/hexagon/services/hybrid` | nachgerechnet: 1791 und 49, Summe 1840 | stimmt mit dem Plan (§6) überein. **Geschlossen.** |

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-12 | MEDIUM | Die Zusage „der Dienst wird byteweise gewählt“ (`SPEC-284` Nr. 4) ist für Groß- und Kleinschreibung des Pfades ungefangen. Die Mutation, die Pfad und Root vor dem Vergleich in Kleinbuchstaben wandelt (`strings.HasPrefix(strings.ToLower(pfad), strings.ToLower(root)+"/")`), bleibt grün, auch im ganzen Paket: kein Fall führt einen Pfad, der sich vom Root nur durch die Schreibweise unterscheidet (`/A/x` gegen `/a`), weder `TestDienstwahl` noch `TestDienstwahlKeinDienst` noch `TestDienstwahlOhneURL`. Folge: ein Request auf `/A/x` erhielte `Ziel` und OData-Pfad des Dienstes `/a`, und die Policy gäbe den Upstream dieses Dienstes für einen Pfad frei, den die Spezifikation keinem Dienst zuordnet. | `AGENTS.md` §3.15 (ein Fall je Form der Eingabe, beide Seiten der Grenze), §3.17 (Klasse „Dienstwahl“, Elemente des Wortlauts „byteweise“) | `upstream.go` · `case strings.HasPrefix(pfad, root+"/"):`; Test `upstream_test.go` · `func TestDienstwahl` (`faelle := []struct{ pfad, ziel, odata string }{`) | ja — Mutation `q08-root-ohne-gross-klein`, Ergebnis `gruen` (zwei Prozesse, Stufe 1 und 2) | Fänger führt nur eine Form der Eingabe (Schreibweise des Pfades fehlt) |
| F-13 | LOW | Die Zusage „an Kommas geteilt“ (`SPEC-284` Nr. 8 c) ist für andere Trenner innerhalb eines Tokens ungefangen: `strings.FieldsFunc(w, … r == ',' \|\| r == ';')` statt `strings.Split(w, ",")` bleibt grün; `TestHeaderFilter` und `TestHeaderRand` führen kein Token mit `;` (oder Leerraum in der Mitte) im Wert von `Connection`. Ein Wert wie `close;X-Edge` würde mit der Mutation den Header `X-Edge` entfernen. Der Fix von F-7 deckt die Randzeichen und die Faltung, nicht die Trenner. | `AGENTS.md` §3.15, §3.17 | `upstream.go` · `strings.Split(w, ",")`; Test `upstream_klassen_test.go` · `func TestHeaderRand` | ja — Mutation `q17-connection-token-semikolon`, Suite grün | Fänger schmaler als die Zusage (Trenner innerhalb des Tokens) |
| F-14 | LOW | Zwei Zusagen sind ohne Fall, in dem der Wert vom Original abweicht: (a) `Methode` „der Anfrage“ unverändert (`SPEC-284` Nr. 8 a): `strings.ToUpper(a.Methode)` bleibt grün, weil jede Methode in den Tests schon in Großbuchstaben steht (`q11-methode-gross`); (b) „ist `ctx` beendet, ist das Ergebnis `ctx.Err()`“ (Nr. 3 e): `context.Cause(ctx)` bleibt grün (`q14-ctx-cause`), weil kein Fall einen Kontext mit Ursache (`WithCancelCause`) führt; ein HTTP-Server beendet den Kontext einer Verbindung mit einer Ursache, die von `ctx.Err()` abweicht. Beides ist am Eingang des Dienstes unwahrscheinlich (der Adapter liefert Großbuchstaben; ein Fehler mit anderer Ursache hat dieselbe Art), die Zusagen haben aber keinen Fänger. | `AGENTS.md` §3.15 (gleiche Werte in Fixture und Mutation) | `upstream.go` · `Methode:  a.Methode,` und `return inbound.Antwort{}, ctxErr`; Tests `upstream_test.go` · `func TestUpstreamArgumente`, `upstream_klassen_test.go` · `func TestUpstreamAbbruchKlassen` | ja — Mutationen `q11`, `q14`, Suite grün | Fixture gleicht dem Wert der Mutation (Methode groß; Kontext ohne Ursache) |
| F-15 | LOW | Der Kopf des Plans trägt `SPEC-121` nur mit „Nr. 4 und 7“; `244827e8` ändert `SPEC-121` Nr. 11, und `80a26331` ergänzt in §1 das Paket `record` und den Test dort, zieht aber die Zeile `Berührte Spec-Stellen` nicht nach. | `AGENTS.md` §3.9 (der Kopf `Berührte Spec-Stellen` folgt jeder Korrektur) | `docs/plan/planning/in-progress/slice-hybrid-service-upstream.md` · `**Berührte Spec-Stellen:**` · `SPEC-121` Nr. 4 und 7 | ja — Lesen des Kopfes gegen `git show 244827e8 --stat` | Kopf des Plans nicht nachgezogen |
| F-16 | INFO | Die Rate des Fängers für einen Schreibzugriff ohne Wirkung auf das Ziel ist ohne Detektor niedrig: `slices.Reverse(v.dienste)` ist ohne Detektor in 14 von 20 (in `upstreamFallback`) und 12 von 20 (in `waehle`) Prozessen rot, mit Detektor in 20 von 20 (Last 5,9 bis 23 von 4, Messender: der Reviewer). Der Plan nennt als akzeptiertes Negativ nur den Schreibzugriff, der nie ein Ziel ändert; der Reverse ändert die Reihenfolge, und der Detektor fängt ihn je Prozess. Das Negativ trägt Grund und Folge; die Zusage gilt wegen des Detektors als gefangen. | `AGENTS.md` §3.10 | Plan §3 · `Nebenläufigkeit (`AGENTS.md` §3.10; Messender: der Implementer` | ja — `q22`, `w05` | — |
| F-17 | INFO | `make test`, `make test-race` und `make build` kamen im Gate-Lauf aus dem Schicht-Cache (gleiche Eingabe wie beim Implementer). Der Verifier fährt sie, wenn er sie frisch braucht, mit `make -B` oder nach einer Änderung der Eingabe. Zuständig: Verifier. | Maintainability | `gates-Lauf` · `pruefung test /src`, `pruefung race /src` | nein | — |

---

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `spec/spezifikation.md` (`SPEC-284` Nr. 11, `SPEC-121` Nr. 11) | geprüft, ohne Befund: beide Stellen sagen dasselbe (`ErrUpstreamZeit` vor `ErrUpstreamNichtErreichbar`, 504), `SPEC-120` Nr. 8 trägt die Rangfolge, kein Widerspruch zum Lastenheft, keine Wellen- oder Slice-Bezüge |
| `src/hexagon/services/hybrid/upstream.go`, `puffer.go`, `verteiler.go`, `fehler.go`, `antwort.go` | geprüft, ohne Befund am Code; Befunde F-12 bis F-14 liegen im Fänger |
| `src/hexagon/services/hybrid/upstream_klassen_test.go`, `nebenlaeufig_upstream_test.go`, `quelltext_test.go` (`schreibtEmpfaenger`) | geprüft: Köpfe nennen, was die Tests führen; die Grenze von `schreibtEmpfaenger` (Aufruf, Alias) steht im Kopf und ist durch `w04`, `w06`, `q22`, `w05` über den Wirkungs-Fänger und den Detektor gedeckt (§3.11); Zählprüfungen (15, 32, 16, 4, 19) stimmen mit den Mengen überein |
| `src/hexagon/services/record/ohneantwort_test.go`, `protokoll.go` | geprüft, ohne Befund (`q21` rot; beide Reihenfolgen des `errors.Join`, `errors.Is` auf beide Fehler in der Ursache); kein Produktcode geändert |
| Kommentare in Code und Tests | geprüft, ohne Befund (keine verworfene Alternative, kein abwesender Text, keine Chronik) |
| `AGENTS.md` §3.2 (Suppression) | geprüft, ohne Befund (keine `//nolint`-Direktive im Diff) |
| `AGENTS.md` §3.4 (Architektur meilensteinfrei) | geprüft, ohne Befund (`spec/architecture.md` nicht berührt) |
| `AGENTS.md` §3.13 (Adresse) | geprüft, ohne Befund: die Randform von F-6 steht mit Ausgang *entschieden* in §6; die Zeilen mit Adresse Planner (Größenmarke) bleiben *offen* |
| `AGENTS.md` §3.14 (Größenmarke) | geprüft: 1840 (1791 und 49) stimmt mit `git diff --numstat 75249f56..HEAD` über `src/hexagon/services/hybrid` überein; die Marke steht aus (Adresse Planner, bekannt) |
| `AGENTS.md` §3.18 (Rohbytes) | geprüft, ohne Befund (siehe Vorgehen 4) |
| `docs/plan/planning/in-progress/roadmap.md` | nicht gemeldet (Integrationsschritt); `docs-check` auf `54f75aea` grün |

---

## Summary

0 HIGH, 1 MEDIUM, 3 LOW, 2 INFO.

Die Fixes der ersten Runde decken ihre Klassen bis auf zwei Randelemente: F-1 bis F-6 und F-8 sind geschlossen, F-5 ist je Prozess mit Detektor 20 von 20 gefangen und ohne Detektor mit 12 bis 14 von 20 als Größenordnung gemessen. Neu offen ist die Schreibweise des Pfades bei der Dienstwahl (F-12, MEDIUM, Fänger führt nur eine Form der Eingabe) und die schmalere Menge „Trenner“ in F-7 (F-13, LOW). Am Code selbst fand ich kein Fehlverhalten.

**Konflikt-Pfad:** kein HIGH, kein Rollen-Widerspruch; F-12 bis F-15 gehen als Annahme oder Begründung an den Implementer, keine Sequenz über den Architect.
