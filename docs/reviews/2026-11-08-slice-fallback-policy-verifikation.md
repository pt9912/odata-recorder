# Verifikation slice-fallback-policy, Teil 1 (Prüfung der Einträge und `New`)

**Rolle:** Verifier (einmal). **Datum:** 2026-11-08. **Stand:** Zweig `slice-fallback-policy`, HEAD `e246094a` (Fix-Runde nach dem [Review](2026-11-08-slice-fallback-policy-review.md)). **Bezug:** [LH-FA-12](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), `SPEC-261`.
**Plan:** [slice-fallback-policy](../plan/planning/in-progress/slice-fallback-policy.md).

## Urteil

Die Zusagen des Teils 1 sind gehalten. Keine Abweichung von Vertrag und Plan, die den Übergang nach `done/` hindert. Zwei Hinweise (kein Befund), drei Auflagen für Closure und Integration (unten). Eine Sicherheitslücke in der Freigabe-Syntax oder im Host-Eintrag fand ich nicht.

## Belege der Sensoren (selbst gefahren)

| Sensor | Ergebnis |
|---|---|
| `make gates` (einmal, unter `flock /tmp/odata-recorder-heavy.lock`, am HEAD, Last 5 bis 11) | Exit 0, alle Gates grün, darunter `lint`, `a-check`, `test-race`, `katalog-check`, `docs-check` |
| `groessenmarke` | `slice-fallback-policy.md: ok, Messung 1311 von Marke 1400`; `Messpunkt: zeilen=9 nebenlaeufig=0 fakes=0 produktcode=300 schaetzung=1020 marke=1400 messung=1311` |
| eigene Messung `git diff --numstat 9310c219..HEAD -- src/hexagon/services/hybrid` | 1311 (Produktcode 358 Zeilen in drei Dateien, 286 ohne Leerzeilen und Kommentare; Tests 953) |
| Last | Hostlast beim Start 11, beim Gate-Lauf 5; Gates warteten auf das Schloss, kein paralleler Lauf |

Die Behauptung des Implementers (Gates grün, 1311) ist damit belegt, nicht nur zugesagt.

## Plan-vs-Code-Diff

- Berührt sind genau `src/hexagon/services/hybrid/policy/` (9 Dateien, davon fünf Tests), `spec/spezifikation.md`, `spec/architecture.md`, der Plan und der Review-Report. Keine Änderung an `Makefile`, `Dockerfile`, `harness/`, `.golangci.yml`, `.a-check.yml`, `tools/`, `GATE_CHECKS`, an `src/hexagon/model/` (die Zahl der Fehlercodes bleibt 40) und an keinem anderen Pfad. Das deckt Plan §3 und §6 (Gate-Kette: entfallen).
- Das Paket importiert `errors`, `fmt`, `net/netip`, `strings`; kein Port, kein Adapter, kein `model`. `TestQuelltext` und `make a-check` bestätigen es.
- Der Plan hat neun Zeilen (P01a, P11, P13, P14, P17 bis P21); die Zahlen der Mengen (30, 19, 37, 169, 10, 19, 5, 8) stimmen mit den Zählprüfungen der Tests überein (gelesen: `len(...) != 10/20`, `5/14`, `9/28`, `13`, `169/6`, `2/8`, `n != 19`, `5`).
- **Produktcode 358 Zeilen gegen Block 300.** Der Block ist eine Erwartung des Architect, keine Grenze; die Marke (1400) gilt für die Summe, und die Messung 1311 liegt darunter. Die Abweichung (+19 % roh) stammt aus der Fix-Runde (`punktEinzel`, Paketkommentar) und der Host-Zerlegung; sie ändert keine Schätzung, die das Werkzeug als Befund meldet. Bewertung: Hinweis, kein Befund. (Die genannten 340 konnte ich nicht reproduzieren: 358 roh, 286 ohne Leerzeilen und Kommentare.)

## Vertrag `SPEC-261` gegen den Code

| Stelle | Ergebnis |
|---|---|
| Nr. 5 a Host-Syntax (Name, `[IPv6]`, Port 1 bis 65535 ohne führende Null, `leer`/`host`) | gehalten |
| Nr. 5 b Schreibweise (ASCII-Kleinschreibung, kanonische IPv6) | gehalten |
| Nr. 5 e (Eintrag gibt nur den genannten Host frei; `[::ffff:1.2.3.4]` gültig, gleich `[::ffff:102:304]`) | gehalten, in P11 mit eigenem Fall |
| Nr. 6 a Rangfolge `leer`, `trenner`, `methode`, `muster` | gehalten |
| Nr. 6 b Muster, Segmente, `*` höchstens einmal, `**` nur als letztes ganzes Segment, Punkt-Segmente | gehalten, siehe Klassenauflösung |
| Nr. 6 c, e (Segment mit `*` neben Punktanteil ungültig, sechs Muster) | gehalten |
| Nr. 7 Kategorien (Schreibweise zählt) | gehalten |
| Nr. 8 a bis d (alle Listen auch bei `Upstream` aus, erster Verstoß je Schlüssel ab 1, `errors.Join` in Reihenfolge der Schlüssel, danach `ohne-hosts`, keine Policy bei Fehler, Zusammenhänge, keine Umgebung) | gehalten |
| `*KonfigFehler`, Text ohne Wert | `Error()` formatiert allein Schlüssel, Nummer, Ursache; das Struct hält keinen Wert; P19 mit Marker `GEHEIM-7f3a` für alle fünf wertbezogenen Ursachen |
| *Ungeprüft* (Nr. 5 e: Name auf unerwartete Adresse) | im Vertrag benannt; kein Fänger nötig, im Code nicht behauptet |

Offen für den Folge-Slice, nicht Teil 1: Nr. 5 c und d (Ziel lesen, Vergleich), Nr. 6 c (Vergleich), Nr. 2 bis 4, Nr. 9 a und b.

## Tabelle Zusage · Fänger gegen den Vertrag

P11, P13, P14, P17, P18, P19, P20, P21 und P01a decken je eine Zusage oder Randform des Vertrags; ich fand keine Zusage der Nr. 5 a/e, 6 a/b/c/e (Syntaxteil), 7, 8 a bis d ohne Zeile. P01a ist ein akzeptiertes Negativ mit Grund und Folge und nennt den Fänger (P01 im zweiten Slice); das genügt `AGENTS.md` §3.15. Die Negative von P21 (kein `init`-Verbot, nur Wurzelverzeichnis, feste Untergrenzen) tragen Grund und Folge. Die Zahlen der Mutationen im Plan (151, davon 150 rot, eine äquivalent) habe ich nicht Stück für Stück neu gefahren; ich habe zehn eigene gefahren (unten) und die Mutationsliste nicht im Repo vorgefunden (der Plan sagt es selbst).

## Fix-Runde gegen die Klasse aufgelöst (`AGENTS.md` §3.17)

**M1 (Segmentmuster).** Die Menge bilde ich aus dem Wortlaut von Nr. 6 b und 6 e, nicht aus dem Befund: alle Segmente, in die ein `*` (ein bis drei Bytes aus W, auch `2`, `e`, `E`, `4`, `1`) so eingesetzt werden kann, dass das Ergebnis nach vollständiger Prozent-Dekodierung (`url.PathUnescape`, also mehr als `%2e`/`%2E`) `.` oder `..` ist; ebenso jedes Segment ohne `*`. Generator: alle Folgen von bis zu vier Teilen aus `.`, `%2e`, `%2E`, `*`, `a`, `%41`, `%2f`, `%2e%2E` (4680 Segmente), jedes an fünf Stellen (`/a/S`, `/S/b`, `/a/S/**`, `/S`, `/S/**`), also 23 400 Muster; Orakel unabhängig vom Code (Einsetzung plus Dekodierung statt Regel je Segment). Ergebnis: **0 unsicher akzeptiert, 0 ohne Unsicherheit abgewiesen** (bei höchstens einem `*`). Kein gültiges Muster deckt ein Segment `.` oder `..`, kodiert oder nicht, außer dem ganzen Segment `*`/`**`, das der Vertrag (Nr. 6 c) dem Vergleich überlässt (Ungeprüft in diesem Slice, Fänger P-Zeilen des zweiten Slice).

Eigene Mutationen an Elementen, die weder Befund noch Fix nannten (Kopie des Pakets samt Tests, Container ohne Netz, `--memory=3g`), alle rot:

| Mutation | rote Tests |
|---|---|
| `%2e` in `ersetzePunkte` entfällt | `TestFreigabeGenerator`, `TestMusterEintraege`, `TestMusterStern` |
| `%2E` entfällt | `TestMusterEintraege`, `TestMusterStern` |
| `punktEinzel` prüft nur `nach` | `TestFreigabeGenerator`, `TestMusterStern` |
| `punktSegment` nur `.` | `TestFreigabeGenerator`, `TestMusterEintraege` |
| `punktSegment` nur `..` | `TestFreigabeGenerator`, `TestMusterEintraege` |
| `punktEinzel` auch bei `..` (zu streng, `..*`) | `TestFreigabeGenerator`, `TestMusterStern` |
| `punktEinzel` auch bei leerem Rest (`*` allein) | `TestFreigabeGenerator`, `TestMusterEintraege`, `TestMusterStern` |
| Zone des Hosts nicht geprüft | `TestHostEintraege`, `TestHostGenerator` |
| `Is6` des Hosts entfällt (`[1.2.3.4]`) | `TestHostEintraege` (der Generator fängt sie nicht; der Golden genügt) |
| Kleinschreibung des Namens entfällt | `TestHostEintraege`, `TestHostGenerator` |

Ohne Mutation: Basis grün. Die Klasse ist gedeckt; der Fänger läuft als Schleife über 13 × 13 Segmente und drei Stellen und ist breiter als Befund und Fix.

**I1 (`[::ffff:1.2.3.4]`).** In P11 als eigener Fall, gleich `[::ffff:102:304]`; Mutation „abgewiesen" rot (Plan). In meinem Host-Generator (7 Grundformen, jedes Byte 0 bis 255 an jeder Stelle einfügen, ersetzen, löschen; 4083 angenommene Einträge): jede angenommene Klammerform trägt genau die genannte Adresse (`netip.ParseAddr`, Gleichheit der kanonischen Form, `Is6`, keine Zone); jeder angenommene Name ist ASCII-klein, ohne `:`, `[`, `]`, `/`, `@`, `%`, Leerraum. **Ein Host-Eintrag gibt keine andere Adresse frei als genannt.** Die Gleichsetzung von `[::ffff:1.2.3.4]` mit dem IPv4-Ziel ist ausgeschlossen (Nr. 5 a); ein Betreiber kann sie dennoch erwarten, das Grenzfeld steht in Nr. 5 e (*Ungeprüft*).

## Sicherheit

- Muster: siehe oben, 23 400 Muster, keine Deckung eines Punkt-Segments.
- Host: siehe oben, keine Verwechslung von Adresse oder Name.
- `*KonfigFehler`: kein Wert im Text und im Struct (P19, fünf Ursachen). Beim Ablehnen schreibt `parseHost` den Eintrag nirgends hin.

## Überschneidung

`git merge-tree --write-tree` gegen `main`, `slice-etag-versionen`, `slice-admin-lebenszyklus`: kein Konflikt in Go-Dateien, in `Makefile` oder im Harness. Konflikte nur in `spec/spezifikation.md`, rein textlich:

- gegen `main` (Snapshot-Dienst wird `SPEC-260`) und `slice-admin-lebenszyklus`: je die letzte Zeile der Änderungstabelle (beide fügen unten an);
- gegen `slice-etag-versionen`: die Stelle der neuen Abschnitte vor `SPEC-280` (beide fügen an derselben Stelle ein) und die letzte Zeile der Änderungstabelle.

Kennungen kollidieren nicht (`SPEC-260` main, `SPEC-261` hier, `SPEC-280` bis `282` etag). Die Auflösung ist „beide behalten"; wer als Zweiter integriert, löst sie im Merge-Commit, nicht durch Kürzen. Fehlercodes: keiner neu (Zahl 40 unverändert), kein Zusammenstoß mit den anderen Zweigen.

## Offene DoD-Punkte (Stand HEAD)

Die Häkchen der DoD stehen alle offen, §7 ist leer: erwartet, sie gehören in die Closure (Planner). Zu belegen sind dort: Gates grün (hier belegt), Review (Report liegt vor), Verifikation (dieser Report), Zeile in §7, dass das Binary unverändert blieb (der Handbuch-Punkt der DoD), Beobachtungs-Register, Ausgang jedes Risikos (alle tragen einen; der Punkt „offene Fragen an den User" ist als weiter offen mit Adresse benannt).

## Auflagen

1. **Closure (Planner):** §7 nennt, dass das Binary unverändert blieb; DoD-Häkchen und Register nachziehen; das Messergebnis 1311 und die Zeile `Messpunkt` eintragen (`SPEC-128` Nr. 9).
2. **Integration:** Reihenfolge nach `main` beachten; der Merge-Konflikt in `spec/spezifikation.md` (Änderungstabelle, Einfügestelle vor `SPEC-280`) ist beim Zweiten von Hand zu lösen, beide Seiten bleiben. `make gates` danach neu.
3. **Folge-Slice `slice-fallback-policy-entscheidung`:** Das Segment, das ganz von `*`/`**` gedeckt wird, darf `.`/`..` nicht sein (Nr. 6 c); diese Zusage hat in Teil 1 keinen Fänger und ist dort zu fangen (P-Zeilen des zweiten Plans), ebenso `Entscheide` für Nr. 5 c und d.

Hinweis ohne Auflage: Der Paketkommentar von `policy.go` ist eine einzige lange Zeile und nennt „braucht es noch nicht"; ein Satz über den Stand eines Imports kann veralten (`AGENTS.md` §3.7, Klasse *Grenze*). Kein Befund.

## Geprüfte Dateien und Gegenlesung

Geschrieben und gegengelesen: nur dieser Report (ASCII mit deutschen Umlauten); kein Wert mit Backslash-u-Folge. Probe-Pakete und Mutationsskript liegen im Scratchpad des Laufs, nicht im Repo.
