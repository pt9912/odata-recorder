# Review-Report: slice-handbuch-meldungskatalog — 2026-11-07

**Review-Art:** Code — gegen Plan, SPEC-201, SPEC-102 und Hard Rules.

**Gegenstand:** `slice-handbuch-meldungskatalog`, Zweig `slice-handbuch-meldungskatalog`, Diff `567b8ba..81869b6`.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-07

**Eingangs-Kontext:**

- `slice-handbuch-meldungskatalog` (§3 Tabelle Zusage · Fänger und Mutationen, §6 Raster und Rückgaben)
- `SPEC-201` (Vertrag des Abgleichs), `SPEC-102` (Fälle 174 bis 188), [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) (Ausnahmen in `.golangci.yml`)
- `LH-QA-05`, `LH-RB-01`
- `AGENTS.md` §3.6, §3.7, §3.9, §3.10, §3.11, §3.12 bis §3.18

**Messung des Reviewers.** 95 eigene Mutationen an `tools/katalog` (ein Wegwerf-Treiber im Scratchpad, ein Container aus `odata-recorder:race`, ohne Netz und Mount, `--memory 2g --cpus 4`, unter `flock /tmp/odata-recorder-heavy.lock`, je Mutation `go test -count=1 ./...`, 168 s; Host mit 20 Kernen, Load Average 3,9 bei Beginn). Ergebnis: 67 rot, 4 Baufehler (ungültig, nicht gezählt), 24 grün. Von den 24 grünen sind 10 gleichwertig oder keine Mutation (Kürzung, `\r`-Entfernen in `zerlege` samt zwei Varianten, Abbruch nach `Unquote`-Fehler, Vergleich trotz Eingabefehler, eingerückter Zaun-Beginn, `TrimSpace` vor `TrimPrefix`, Leerzeile der Trennzeile, zwei Ersetzungen ohne Wirkung); 14 sind echte Lücken (F-1 bis F-6). Die Rate ist 0 oder 1: kein Fänger hängt an Wahrscheinlichkeit.

**Gleichwertigkeit geprüft.** Die zwei Einträge des Plans sind echt gleichwertig: (1) Kürzung bei genau 200 Zeichen: `q[:200]` schneidet bei Länge 200 nichts ab. (2) `\r`-Entfernen in `zerlege`: jeder Leser trimmt rechts mit `unicode.IsSpace` (Überschrift, Zaun-Schluss, Zellen, Ausnahmen), `nenne` hört an `\r` auf, kein Ausgabewert trägt `\r`; die Zeilenzahl ändert sich nicht. Das Negativ nennt Grund und Folge (§3.15 erfüllt). Die Mutation "nicht bildbar" (`[0-9]` als `\d`) ist stimmig: `TestKatCZiffern` führt 1024 Bytewerte.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | "Nach Zeile aufsteigend" hat keinen Fänger: Entfernen des Schlüssels `cmp.Compare(x.zeile, y.zeile)` bleibt grün, weil der Rückfall auf den Text (`Zeile 10` vor `Zeile 9`) in keiner Fixture von 9 auf 10 springt; die Umkehr (absteigend) wird nur von `TestKatCForm` gefangen. Szenario: zwei Befunde Art 1 oder Art 2 am selben Namen mit Zeilen 9 und 10 erscheinen vertauscht, nichts wird rot. | `SPEC-201` Nr. 2 d; `AGENTS.md` §3.15 (beide Seiten jeder Grenze) | `tools/katalog/bericht.go` · `cmp.Compare(x.zeile, y.zeile)` | ja: Mutation `sort-ohne-zeile` grün | Sortierschlüssel ohne Fänger an der Stellengrenze |
| F-2 | MEDIUM | Die Zusage "Befund Form an dieser Zeile" (Nr. 3 d) für die Nennungen im Handbuch hat zwei grüne Mutationen: der Schlüssel der Entdopplung nur `x` (derselbe falsche Name in zwei Zeilen gibt einen Befund statt zwei) und nur `n` (zwei verschiedene falsche Namen in einer Zeile geben einen Befund). `TestKatNennungForm` führt keine der beiden Formen. | `SPEC-201` Nr. 3 d; `AGENTS.md` §3.15 | `tools/katalog/quellen.go` · `k := strconv.Itoa(n) + " " + x` | ja: Mutationen `hb-formkey-x`, `hb-formkey-n` grün | Entdopplungsschlüssel ohne Fänger je Achse |
| F-3 | MEDIUM | Die Rückgabe (e) des Plans behauptet "eine Ausnahme mit Form-Verstoß und leerem Grund trägt beide Befunde (Art 1 und Art 5)"; kein Test führt dieses Paar. Die Mutation `grund == "" && istCode(code)` bleibt grün (die Form-Fälle von `TestKatAusnahmenPruefen` tragen alle den Grund `Grund`, die Fälle mit leerem Grund einen gültigen Code). Szenario: eine Ausnahme `ODR-E123 |` meldet nur die Form, der fehlende Grund bleibt unbemerkt. | `AGENTS.md` §3.16 (Paar von Ursachen am selben Schlüssel); Plan §6 (e) | `tools/katalog/quellen.go` · `if grund == "" {` | ja: Mutation `aus-grund-nur-gueltig` grün | Paar von Verstößen am selben Schlüssel ohne Fänger |
| F-4 | MEDIUM | Die Zusage "bis zur nächsten Überschrift derselben oder einer höheren Ebene" ist nur für die gleiche Ebene gefangen: `e2 == e` statt `e2 <= e` bleibt grün (Abschnitt `###` endet an einer `##`-Überschrift in keiner Fixture; die Zeile 19 des Plans nennt "Ebene fest 2" und "über 3", nicht die höhere Ebene). | `SPEC-201` Nr. 6 b; `AGENTS.md` §3.15 | `tools/katalog/markdown.go` · `ok && e2 <= e` | ja: Mutation `md-abs-eq` grün | Ebenen-Grenze nur auf einer Seite gefangen |
| F-5 | MEDIUM | Eine Tabelle, die nur aus der Kopfzeile besteht, hat keinen Fall: Entfernen von `len(t) < 2 ||` vor `trennzeile(zellen(m.zeilen[t[1]]))` bleibt grün, obwohl `t[1]` dann außerhalb der Folge liegt (Go-Panik, Exit 2 mit Stapelspur statt `Trennzeile … fehlt`). Die Fixtures "Trennzeile fehlt" haben dahinter Datenzeilen. | `SPEC-201` Nr. 6 c, Nr. 8; `AGENTS.md` §3.16 (Randform Operation × Argumentart) | `tools/katalog/markdown.go` · `len(t) < 2 \|\| !trennzeile` | ja: Mutation `md-trennzeile-len` grün | Randform "nur Kopfzeile" ohne Fänger |
| F-6 | MEDIUM | Sechs weitere Wortlaut-Zusagen haben je einen grünen Nachbarn: (a) Kopf "bytegleich" (`strings.EqualFold` statt `Equal`); (b) Überschrift §4 "genau" (`HasPrefix` statt Gleichheit); (c) Präfix `ODR-` (`ODR` ohne Strich: ein Literal `ODRX…` würde Form statt "kein Code"); (d) Trennzeile "ein `:` an einem oder beiden Enden" (`Trim(c, ":")` nimmt `::--`); (e) Zeile einer Tabelle "beginnt mit `\|`" (eingerückte Zeile zählt); (f) Zaun in Zeile 1 ohne Schluss (`m.offen > 1`: die Zeile ist zugleich Zähler und Wächter 0); (g) Grund von `nicht lesbar` ohne Pfad (die Mutation, die den Pfad der Wurzel in den Text nimmt, bleibt grün; `TestKatCEingabe` und `TestKatHandbuchEingabe` prüfen nur das Präfix `nicht lesbar: `). Jede für sich schmal, in Summe Fänger führen eine Form der Eingabe. | `AGENTS.md` §3.15, §3.17; `SPEC-201` Nr. 2 d (gleiche Eingabe, gleiche Ausgabe, gilt für g), Nr. 3 a, 3 b, 6 a, 6 c | `tools/katalog/markdown.go` · `slices.Equal(zellen(m.zeilen[t[0]]), f.kopf)`; `quellen.go` · `t == ueberschriftK`, `!strings.HasPrefix(v, "ODR-")`, `grund = pe.Err.Error()`; `markdown.go` · `strings.Trim(c, "-") != ""`, `m.offen = i + 1` | ja: Mutationen `md-kopf-fold`, `k-prefix`, `c-prefix-ODR`, `md-trenn-trimall`, `md-tab-prefix-trim`, `hb-offen-ignor`, `lies-grund-mit-pfad` grün | Fänger schmaler als Zusage (Gruppe) |
| F-7 | MEDIUM | Zwei Randformen sind im Code entschieden und stehen weder in `SPEC-201` noch in den Rückgaben (a) bis (h) von Plan §6: (a) das Abschnittsende von §4 ignoriert Zeilen `## ` in Zäunen (der Wortlaut "nächste Zeile, die mit `## ` beginnt" sagt es nicht; Mutation `k-ende-ohne-zaun` grün); (b) "getrimmt" meint `unicode.IsSpace` (NBSP, U+0085, U+2028 und weitere werden an Zellen, Ausnahmen, Überschrift-Ende und Zaun-Schluss mit abgeschnitten); `SPEC-201` definiert "getrimmt" nicht, und kein Test führt ein Unicode-Leerzeichen. | `AGENTS.md` §3.12, §3.16 (eine im Code entschiedene Zelle ist ein MEDIUM) | `tools/katalog/quellen.go` · `!m.zaun[j] && strings.HasPrefix(m.zeilen[j], "## ")`; `tools/katalog/markdown.go` · `func rechtsGetrimmt` | nein (Entscheidung ohne Vertragsort); (a) per Mutation grün | Randform ohne Ort im Code entschieden |
| F-8 | INFO | `.golangci.yml`: die Ausnahme der Regel `os.Std(out\|err)` nimmt `tools/katalog/main.go` in die bestehende Pfadliste auf (zwei Zeilen, `# Why:` angepasst, Plan §3 führt die Zeile). Das ist keine Schwellen-Senkung nach §3.6: die Regel bleibt, die Ausnahme folgt dem Muster der vier Programme mit eigenem Einstieg und nennt nur diese Datei. | `AGENTS.md` §3.2, §3.6; [ADR-0003](../plan/adr/0003-lint-profil-golangci.md) | `.golangci.yml` · `^tools/(gegenproben\|groessenmarke\|katalog\|pruefung\|zeichen)/main\.go$` | ja: `make lint` | Ausnahme folgt bestehendem Muster |

Kein HIGH. Keine Verletzung einer HIGH-Regel der Liste: kein ADR-Verstoß, keine Unterdrückung (`//nolint`), kein Kommentar der verworfenen Alternative, kein Zustandsfeld mit Chronik.

## Bestätigungen und Prüfungen ohne Befund

- **Rückgaben an den Architect (Plan §6, (a) bis (h)):** führt §6 als Zeile mit Ausgang *weiter offen* und Adresse *Architect* (`AGENTS.md` §3.13). Nicht entschieden. Hinzu kommen F-7 (a) und (b) als Formen, die dort fehlen.
- **Rückgabe an den Planner (`tools/mutationslauf`, `relInModul`):** bestätigt. `relInModul(m0, m0.Paket)` schneidet bei `Paket == Modul.Dir` nichts ab (`TrimPrefix(p, Dir+"/")` auf `p == Dir`), `nullprobe.go` ruft sie statt `paketArg`; die Nullprobe läuft in einem Verzeichnis, das nicht existiert. Ausgang *weiter offen*, Adresse Planner, steht in §6.
- **Größenmarke:** Summe aus Einfügungen und Löschungen über die Pfade des Blocks (`tools/katalog`, `tools/gegenproben`, `tools/harness/gegenproben-eingabe.sh`, `Dockerfile`, `.dockerignore`, `harness/mk/katalog.mk`, `harness/katalog-ausnahmen.txt`, `harness/README.md`) von `cee8df2` und von `567b8ba` bis `81869b6`: je 2222, unter der Marke 3500. Die Ausnahme-Datei ist unverändert im Diff.
- **Plan folgt Code (§3.9):** `.golangci.yml` und `SPEC-102` stehen als Zeilen in §3, die Mutationen mit Mengen und die Rückgaben in §3 und §6. Sensors-Zeile in `harness/README.md` und Zeile "Nicht behauptet" nennen die Verdrahtung ohne Fänger (`SPEC-201` Nr. 9 f).
- **Rohbytes (§3.18):** keine Folge aus Backslash, `u` und vier Hex-Ziffern in `tools/katalog/*.go` (`LC_ALL=C grep -P`); die Fixtures für Nicht-UTF-8 und NUL entstehen aus `\xff` und `\x00`, die Ziffernmenge aus Bytewerten (`TestKatCZiffern`, 1024).
- **Fixture gleicht Mutation:** `TestKatFormGekuerzt` rechnet die erwartete Kürzung mit derselben Logik wie der Code, pinnt aber die Längen (199, 200, 200, 200, 106, 200) unabhängig davon per `t.Fatalf`; Mutationen auf `QuoteToASCII`, Kürzung vor `Quote`, Bytes statt Zeichen, 199 und 201 sind rot. Kein Befund.
- **Gegenproben G1 bis G15 / Fälle 174 bis 188:** Fälle und Erwartungen stimmen mit `SPEC-201` Gegenprobe und `SPEC-102` überein; Ströme getrennt, `strom_test.go` (40 und 148) nachgezogen. Nicht selbst gefahren (kein Gate-Lauf im Auftrag).
- **Kommentare (§3.7, §3.11):** die Köpfe von `main.go`, `pruefen.go`, `bericht.go`, `markdown.go` und `katalog.go` nennen Zustand und Zusage; die Zusage "schreibt nur nach stdout und stderr" fängt `TestKatSchreibfrei`. Kein Befund.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tools/katalog/` (Leser C, K, Ausnahmen, Handbuch; Vergleich; Ausgabe) | geprüft, Befunde F-1 bis F-7 |
| `tools/gegenproben/katalog.go`, `faelle.go`, `main.go`, `strom_test.go` | geprüft, ohne Befund |
| `Dockerfile`, `.dockerignore`, `harness/mk/katalog.mk`, `harness/mk/gegenproben.mk`, `tools/harness/gegenproben-eingabe.sh` | geprüft, ohne Befund (Verdrahtung ohne Fänger ist als akzeptiertes Negativ `SPEC-201` Nr. 9 f benannt) |
| `docs/user/benutzerhandbuch.md` (Abschnitt Fehlercodes, 18 Zeilen) | geprüft, ohne Befund (Inhalt der Zeilen ist ungeprüft nach Nr. 9 a) |
| `harness/README.md`, `harness/vertraege/SPEC-102-gegenproben.md` | geprüft, ohne Befund |
| `.golangci.yml` | geprüft, F-8 (INFO) |
| `docs/plan/planning/done/slice-handbuch-meldungskatalog.md` | geprüft; Rückgaben mit Ausgang und Adresse, siehe oben |

## Summary

Kein HIGH, 7 MEDIUM, 1 INFO. Der Kern ist belastbar: 67 von 81 gültigen, nicht gleichwertigen Mutationen sind rot, alle Befundarten und Eingabefehler sind gefangen. Die Lücken liegen an Grenzen der Zusagen (Zeile 9 auf 10, Ebene höher, nur Kopfzeile, Paar Form und leerer Grund, Entdopplungsschlüssel, Wortlaut-Nachbarn) und an zwei im Code entschiedenen Randformen. Der Fix zählt erst, wenn er die Klasse deckt (`AGENTS.md` §3.17): die Mengen sind "Stellenzahl der Zeilen in der Sortierung", "Ebenen 1 bis 6 gegen die Ebene des Abschnitts", "Paare aus Form-Verstoß und leerem Grund je Quelle", "Kopf in Schreibung und Leerraum", nicht die oben genannten Mutationen. Findings an `SPEC-201` (F-7) gehen an den Architect.
