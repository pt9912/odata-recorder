# Verifikation: slice-admin-listener-token — 2026-11-07

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-239` (Nr. 1 bis 6 und *Ungeprüft*), `SPEC-238` soweit Teil 1, `LH-FA-13.a`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und die Hard Rules; nicht Diff-Review (siehe `2026-11-07-slice-admin-listener-token-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-admin-listener-token`, HEAD `b83a1da1` (Fix-Runde nach Review `ddd20fc5` und Entscheidungen des Architect `61e6483a`), Basis `20c02c5`. Plan: `docs/plan/planning/done/slice-admin-listener-token.md`. Teil 1 des Admin-Adapters: die Token-Datei `admin.LadeToken`.

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden, mit Auflagen an den Planner und den Architect; kein HIGH, drei MEDIUM (Zusage ohne Fänger, `AGENTS.md` §3.15), zwei LOW, ein INFO.** Der Produktcode setzt `SPEC-239` um, ist in der Fix-Runde unverändert (`tokendatei.go` 174 Zeilen), `make gates` ist grün (eigener Lauf), die Größenmarke hält (1280 von 2900), Binary, `src/bootstrap`, `.golangci.yml`, `.a-check.yml` und `ODR-E2102` sind unberührt. Die fünf Fixes der Fix-Runde (F-1 bis F-5 des Reviews) halten gegen die **Klasse**: von 112 eigenen Mutationen blieb keine an einem Element grün, das die Klasse der Fixes trägt. Die offenen Stellen liegen **neben** den Fixes, an Sätzen von `SPEC-239`, die weder die Tabelle T1 bis T26 noch die Negative Z-N1 bis Z-N5 tragen (Umgebung, Zustand, Entropie, ein Zweig, den ein Haltepunkt erreicht). Alle Auflagen sind Tests oder Plan-Zeilen; kein Produktcode ist zu ändern, keine weitere Prüfrunde nötig, wenn der Planner jede Auflage mit einem Test (Probe liegt vor) oder einem Negativ mit Grund und Folge schließt.

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/verifier-admin-token/gates.log`: `EXIT 0`; `a-check` 0 Befunde, `d-check: 821 Datei(en) geprüft, 0 Befund(e)`, `gofmt`/`go vet`/Tests der Stufe `pruefung`, Stufe `race` gebaut, `migration-check: ok`, `zeichen: ok, 1659 Datei(en)`, Lint-Stufe durchgelaufen. |
| `make groessenmarke` | im selben Lauf: `slice-admin-listener-token.md: ok, Messung 1280 von Marke 2900`; eigene Rechnung `git diff --numstat 20c02c5..HEAD` über `src/adapters/driving/admin/`: 13 + 150 + 174 + 943 = 1280. |
| Abdeckung (Register `zweig-ohne-lauf-unter-dem-detektor`, 13×) | eigener Lauf `go test -race -count=1 -coverprofile` im Image `odata-recorder:race` mit Baumkopie (`docker create` + `docker cp`, ohne Netz und Mount, `--memory=4g`): **95,2 %**, genau **drei Blöcke mit Zähler 0**: `tokendatei.go:108.3` (Lesefehler, Z-N1), `:132.3` (Fehler von `schreibe`, Z-N2), `:142.2` (Rückgabe nach `Link`-Fehler außer `fs.ErrExist`, Z-N2). Das sind die drei des Plans. Die Zeile bei `:142` ist aber **erreichbar** (V-4). |
| Flackern | `go test -count=40` ok (61,2 s); `go test -race -count=5` ok (9,0 s); `-race -count=3 -run TestWettlaufNebenlaeufig` ok; `-count=100` über `TestNichtRegulaer`, `TestKeineAusgabe`, `TestZweiterAufruf`, `TestWettlauf*` ok (45,9 s). Kein Test flackert auf dem unveränderten Stand. Kein `t.Parallel` im Paket (`grep`: 0 Treffer); die Umask wird allein in `TestRechteBeiJederUmask` verstellt und per `defer` zurückgesetzt. |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -P` über die vier Dateien des Pakets: je 0 Treffer für Backslash-u-Folgen, BOM, NBSP, Nullbreiten-Zeichen und Steuerzeichen; der Wert für Canary, NBSP und Umlaute entsteht aus `\x`-Folgen und `[]byte`. Der Bericht des Implementers nennt die Gegenlesung. Auch diesen Report habe ich gegengelesen (siehe unten). |
| Hard Rules | `grep nolint` über das Paket: 0; `git diff 20c02c5..HEAD -- .golangci.yml .a-check.yml src/bootstrap Makefile harness`: leer; `grep E2102 src/adapters`: 0 Treffer; Docker-Läufe allein mit Baumkopie, `--memory`, ohne Netz; Repo-Dateien des Zweigs mit `Write`/`Edit` (der Report ist die einzige Datei, die ich schreibe). |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Token-Datei nach `SPEC-239`: Erzeugen (256 Bit, `0600`, nie in einem Log), sieben Gründe in der Rangfolge, Symlink auf Datei gültig, vorhandene Datei unverändert, Gewinner bei Wettlauf | **erfüllt** (Haken fehlt, Sache des Planners), mit den Auflagen V-1 bis V-3 zu Sätzen ohne Fänger | Tabelle T1 bis T26, Abschnitt 4; 112 + 14 + 8 + 10 eigene Mutationen. |
| `make gates` grün | **erfüllt** (Haken fehlt) | eigener Lauf, `EXIT 0`. |
| Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-11-07-slice-admin-listener-token-review.md` (F-1 bis F-8). |
| Doku-Update | entfällt | kein öffentlicher Vertrag berührt: `docs/user/` und Binary unverändert (`src/bootstrap` ohne Diff); die Spezifikation trägt die Berichtigungen des Architect (`SPEC-239` Nr. 1 f, 3 b, 3 f, *Ungeprüft*). |
| Handbuch-Kapitel, Kataloge-Gate | entfällt für diesen Teil | kein neuer Code und kein Schlüssel (`ODR-E2001` besteht). |
| Closure-Notiz, Register, Ausgänge der Risiken, drei Paarungen | offen, Sache des Planners | §6: R7 entschieden, R5 *weiter offen* (User), drei Register-Zeilen *weiter offen*; die Ausgänge sind benannt. |

Liefer-Punkte: einer (unter dem Limit von drei).

## 3. Plan-vs-Code-Diff

`git diff --numstat 20c02c5..HEAD`: `src/adapters/driving/admin/` neu mit `tokendatei.go` (174), `export_test.go` (13), `tokendatei_test.go` (943), `tokendatei_ausgabe_test.go` (150); dazu der Plan (51/6), der Review-Report (49) und `spec/spezifikation.md` (6/4, Berichtigungen des Architect zu `SPEC-239` Nr. 1 f, 3 b, 3 f und den *Ungeprüft*-Zeilen). Das ist die Datei-Tabelle des Plans (§3), nicht mehr und nicht weniger. Der Produktcode des Handoffs und der der Fix-Runde sind gleich (`git diff d8575c3a..HEAD -- tokendatei.go` leer): die Fix-Runde ändert Tests, Plan und Spezifikation, wie der Auftrag des Architect (nur Tests und Plan) es sagte. Importe des Produktcodes: Standardbibliothek und `hexagon/model`; `make a-check` grün. Die Zeilenzahl des Produktcodes (174 gegen 150 geschätzt) liegt innerhalb der Marke.

Plan folgt dem Code (`AGENTS.md` §3.9): Tabelle T1 bis T26 (24 Zeilen: 4 + 20), Bericht *Zusage · Mutation · roter Test*, Raten, Mengen mit Zählprüfung (M-Token-Inhalt 320, M-Rand 520, M-Rechte 13, M-Nicht-regulär 5, M-Gründe 7, M-Umask 4, 40 Aufrufe von `TestKeineAusgabe`), Zweige und Negative Z-N1 bis Z-N5, R7: jede Zahl habe ich im Test gefunden (`len(...) != 320`, `n != 520`, `len(fs) != 13`, `len(fs) != 5`, `len(im) != 7`, `len(umasks) != 4`, `aufrufe != 3+29+5+3`).

## 4. Tabelle *Zusage · Fänger* gegen den Vertrag

Ich habe jeden Satz von `SPEC-239` gegen die Zeilen gelegt und, wo die Zeile einen Fänger nennt, den Fänger mit einer Mutation aufgelöst.

| Satz von `SPEC-239` | Zeile / Fänger | Stand |
|---|---|---|
| Einleitung: Paket, Datei, Standardbibliothek; synchron ohne Kontext | Signatur (Übersetzung), `make a-check` | gedeckt |
| Einleitung: „das Paket liest die Umgebung nie“ | – | **V-1: kein Fänger, kein Negativ** |
| Nr. 1: leerer `pfad` `ErrUngueltig`; relativer Pfad | T12 | gedeckt |
| Nr. 1 a: `Stat` folgt Symlinks; „fehlt“ führt zu Nr. 3, außer `Lstat` gelingt | T13, T4, T9 (Symlink ohne Ziel) | gedeckt; die Wirkung des `Lstat`-Zweigs ist ohne Haltepunkt nicht sichtbar (V-6, INFO) |
| Nr. 1 a: jeder andere Fehler `Datei nicht lesbar` (Pfadelement ist Datei, Schleife, NUL) | T10 (3 Formen) | gedeckt |
| Nr. 1 a: „fehlendes Recht an einem Pfadelement“ | – (Superuser) | **V-5: kein Fänger, Negativ nicht benannt** |
| Nr. 1 b, c | T8, T9 (5 Formen, Frist 10 s) | gedeckt |
| Nr. 1 d: `0400`, `0500`, `0600`, `0700` gültig; jedes Bit von Gruppe und Welt Verstoß; Rechte der aufgelösten Datei | T2 (13), T4 | gedeckt (Maske jedes einzelnen Bits, Abschnitt 6) |
| Nr. 1 e | Z-N1 | akzeptiertes Negativ, Grund und Folge genannt, bestätigt grün |
| Nr. 1 f: genau sechs Bytes Leerraum an beiden Enden, jedes andere Byte bleibt; `\n` oder `\r` im Rest | T1 (320 und M-Rand 520), T7 | gedeckt |
| Nr. 1 g: unter 32 Bytes, leer, nur Leerraum | T5 (12) | gedeckt |
| Nr. 1: Ergebnis als Kopie; innerer Leerraum, Schreibweise, Bytes außerhalb ASCII | T14 (`cap < 512`), T1 | gedeckt; (3) des Handoffs benennt die Abhängigkeit von `os.ReadFile` |
| Nr. 1: Symlink auf Datei und Kette gültig | T13 | gedeckt |
| Nr. 2: Plattform | Z-N3 | akzeptiertes Negativ; der Ausdruck wird auf Linux ausgeführt (Mutation `GOOS != "linux"` rot) |
| Nr. 3 a: 32 Bytes `crypto/rand`, 64 Zeichen `0-9a-f`, Datei = Token + `\n`; 256 Bit | T14, T15 | gedeckt für Länge, Zeichen, Zeilenende, Quelle; **V-3: 256 Bit nicht gefangen** |
| Nr. 3 b: Name `<Dateiname>.tmp-` + 16 Zeichen, `0600` von Anfang an, Teilmenge, bei Umask 000 genau `0600` | T17, T16 | gedeckt (Abschnitt 5) |
| Nr. 3 b: `O_EXCL` | Z-N4 | akzeptiertes Negativ, bestätigt grün |
| Nr. 3 c: `Sync`, Schließen | Z-N5 | akzeptiertes Negativ, bestätigt grün (beide) |
| Nr. 3 d: `Link`, Gewinner, Verlierer lädt nach Nr. 1; andere Fehler `Datei nicht anlegbar` | T18, T19, T20; Z-N2 | gedeckt für den Wettlauf; **V-4: die Begründung von Z-N2 für den `Link`-Fehler trägt nicht** |
| Nr. 3 e: Hilfsdatei auf jedem Rückkehrpfad entfernt, ein Fehler dabei ohne Wirkung | T20 | Entfernen gedeckt; „ein Fehler dabei ohne Wirkung“ ohne Fänger (V-4) |
| Nr. 3 f: Verzeichnis fehlt, Symlink ohne Ziel als Pfadelement: `Datei nicht anlegbar`, nichts entsteht; Pfadelement ist Datei = `Datei nicht lesbar` (R7) | T21, T10 | gedeckt |
| Nr. 4: kein Zustand; zweiter Aufruf gleich; Wettlauf genau eine Datei; Datei nie geändert | T22, T18, T19, T23 | gedeckt außer „hält keinen Zustand“ (**V-2**) |
| Nr. 5: Fehlertext ohne Inhalt und Pfad; kein Log, kein `stdout`, kein `stderr`; Token nur in Rückgabe und Datei | T24, T25 (40 Aufrufe) | gedeckt (Abschnitt 5) |
| Nr. 6: sieben Gründe, Rangfolge | T24 (Zählprüfung 7 gegen den Quelltext), T11 | gedeckt; Permutationen aller drei Arten-Fälle (Abschnitt 6) |

## 5. Die Fixes der Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Die Menge bilde ich aus dem Wortlaut von `SPEC-239`, nicht aus dem Befund.

| Befund | Menge (Wortlaut) | Elemente | Eigene Mutationen an Elementen, die Befund und Implementer nicht nannten | Ergebnis |
|---|---|---|---|---|
| F-1 Allaussage „nach jedem Verstoß keine Ausgabe“ (Nr. 5) | jede Ausgabestelle (Pfad, Token, Inhalt) × jeder Kanal (`println`, `stdout`, `stderr`) in jedem Zweig, der einen der sieben Gründe, den Weg des Verlierers oder das Erzeugen liefert | 20 Stellen, davon 18 erreichbar | 20 Mutationen `P-*`: Verzeichnis, nicht regulär, Rechte, Zeilen, zu kurz, gültiger Rückgabepunkt (`println` und `os.Stdout.Write`), Erzeugen (beide Kanäle), `fehlend`-Zweig, beide Zweige von `statFehler`, `nachAnlegen`, `vorLink`, Verlierer (Aufruf, `Stat`, Prüfung), Anlegefehler | 18 rot; 2 grün = `:108` und `:142`, die zwei Zweige ohne Lauf (Z-N1, V-4) |
| F-2 Rand (Nr. 1 f) | jedes der 256 Bytewerte und jede mehrbytige Folge, links und rechts | 520 | `leerraum` + `0x08`, `0x0e`, `0x1f`, `0x85`, `0xa0`, `0xe0`, `A`, NUL (8); je ein Element des Satzes weggenommen (6); `TrimLeft`, `TrimRight`; Zeilenprüfung vor dem Entfernen; nur `\r`; Rune statt Byte als Maß; `<=`; Länge vor dem Entfernen | alle 21 rot (Ursache gelesen: `L-minus-cr` meldet `mehr als eine Zeile bei gültiger Datei`, `L-runen` die 16-Runen-Fälle von Kern 3) |
| F-3 Hilfsdatei-Name (Nr. 3 b) | Präfix, Länge, Zeichenmenge, Verzeichnis, Zufall | 5 Achsen | `.tmp_`, `tmp-` ohne Punkt, Länge 7 und 12 Bytes, Großbuchstaben, ein Nicht-Hex-Zeichen, Verzeichnis `os.TempDir()`, fester Name | alle 8 rot; der feste Name ist als einziger im Muster (er passt) und wird von T18 durch die Wirkung gefangen (Kollision mit `O_EXCL`) |
| F-5 Rechte der Hilfsdatei (Nr. 3 b) | jedes Bit außerhalb von `0600`; Anlegen und `Chmod` | 9 Bits × 4 Umasks | Anlegen mit `0640`, `0602`, `0300`, `0100`, `0500`, `0701`, `0000`, `0666`, `0610`, `0604` (10); `Chmod` auf `0700`, `0400`, `0640`, `0200`, `0601`, `0000` (6) | alle 16 rot (`R-open-0o300`: `Rechte 0300, Größe 0` bei Umask 0000/0022/0077) |
| F-4 / Z-N5 | `Sync`, Schließen | 2 | `Sync` zu `nil`, `Close` entfernt | beide grün: wie im Plan als Negativ geführt |

Zusatz zu F-2 / Rechteprüfung (Nr. 1 d): Maske jedes einzelnen Bits statt `0o077` (`0o037`, `0o057`, `0o067`, `0o073`, `0o075`, `0o076`), Obermenge (`0o177`, `0o277`, `0o477`), `> 1`, `> 0o10`, nie, `GOOS != "linux"`: alle 13 rot durch `TestRechteMenge`.

## 6. Weitere eigene Mutationen (Messender: der Verifier)

`tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, `MUTATIONSLAUF_SPEICHER=4g`, unter `flock`, Last des Hosts 6,6 bis 15,1 bei 4 CPUs. Listen und Ausgaben liegen im Scratchpad `.../scratchpad/verifier-admin-token/` (`liste-a.json`, `erg-a.txt`, `liste-b20.json`, `erg-b20.txt`, `erg-b50.txt`, `liste-c.json`, `erg-c.txt`, `liste-d.json`, `erg-d.txt`, `ursache.txt`). Erzeugt mit `jq` aus dem Quelltext; jede Mutation trifft ihr `alt` genau einmal (das Werkzeug prüft es).

| Liste | Gegenstand | Ergebnis |
|---|---|---|
| a, 112 | Leerraum (21), Hilfsdatei (8), Rechte der Hilfsdatei (16), Rechteprüfung (13), Reihenfolge (6: alle 5 Permutationen der Fälle Verzeichnis/Art/Rechte, Länge vor Zeilen), Art (5), Stat-Fehler und R7 (5), Ausgabe (20), Token (6), Unverändert (8), Negative (4) | 101 rot, **11 grün**: 5 akzeptierte Negative bestätigt (Z-N1 `:readfile`, Z-N2 `:link`, Z-N4 `O_EXCL`, Z-N5 `Sync`, `Close`); 3 äquivalent (`S-eloop-fehlt`: der `Lstat` weist die Schleife vorab ab; `V-chmod-zeilen`, `V-chmod-kurz`: die Fixture trägt schon `0600`); **3 Befunde**: `E-halbe-entropie`, `E-viertel` (V-3), `N-lstat-weg-F6` (V-6) |
| a, Ursache | 11 rot gelesen: `A-nur-geraet`, `E-31`, `H-suffix-gross`, `L-minus-cr`, `L-runen`, `O-perm-3`, `P-gueltig-stdout`, `R-open-0o300`, `RP-nie`, `S-enotdir-fehlt`, `V-chtimes-gueltig` | alle aus dem benannten Grund rot (Fehlerzeile des erwarteten Tests, kein `panic`, kein Übersetzungsfehler; `ursache.txt`) |
| b, 14 (7 Mutationen je Tests `andere` und `nebenläufig`) | Klasse `nebenläufig`: `Link` zu `Symlink`; fester Name der Hilfsdatei; `Remove(pfad)` statt `Remove(hilfsdatei)`; Verlierer liest ohne Prüfung (`os.ReadFile`); `Stat` dann `Rename`; `WriteFile` auf `pfad`; Link vor dem Schreiben | je 20 Prozesse mit und 20 ohne Detektor: **20/20 und 20/20** für 6 Mutationen; je 50 und 50: **50/50 und 50/50** für 6. Last 7,4 bis 12,9 (n = 20), 8,2 bis 10,4 (n = 50) bei 4 CPUs. Die Zeile *Link vor dem Schreiben* ist `flatternd` in `TestWettlaufNebenlaeufig` (18/20 mit, 19/20 ohne; 45/50 mit, 43/50 ohne), aber ihr deterministischer Fänger `TestVollstaendigOderGarNicht` (Haltepunkt vor dem `Link`, Größe 65) ist in jedem Lauf rot (T17). `Stat` dann `Rename` ist in T18 grün (der Haltepunkt liegt vor dem `Stat`, das Fenster zwischen Prüfen und Handeln erreicht er nicht) und in T19 20/20 und 50/50 rot, dazu `TestHilfsdateiEntfernt` rot. |
| c, 8 | Sätze ohne Zeile: `os.Getenv`; Cache je Pfad; `fs.ErrPermission` als „fehlt“; `Remove`-Fehler mit `panic`; Token im Fehlertext; `MkdirAll`; keine Kopie | rot: Token im Text, `MkdirAll`, keine Kopie; Cache nur als `rot-paket` durch ein *concurrent map write* in `TestWettlaufNebenlaeufig` (ein Zufall der Mutation, kein Fänger der Zusage); **grün**: beide `Getenv`, `ErrPermission`, `Remove`-Fehler (V-1, V-2, V-4, V-5) |
| d, 10 (Probe-Paket, nicht im Repo) | `probe_test.go` in einer Baumkopie: Haltepunkt entfernt die Hilfsdatei; `nachAnlegen` bei Symlink ohne Ziel; Suche nach `os.Getenv`; Entfernen der Datei zwischen zwei Aufrufen; 496 Paare von Byte-Positionen in 100 Token | 9 rot, 1 grün: `println` im `Link`-Fehlerzweig bleibt grün, weil die Probe die Ausgabe nicht erfasst; die Probe zeigt, dass der Zweig erreichbar ist (`D-link-fehler-ok` und `D-remove-fehler-panic` rot). Cache mit Kanal als Sperre, beide Entropie-Mutationen, `b[7]=7`, `Getenv`, `Lstat` entfernt: rot |

## 7. Befunde

| ID | Kategorie | Befund | Auflage |
|---|---|---|---|
| V-1 | MEDIUM | `SPEC-239` Einleitung: „das Paket liest die Umgebung nie“ hat weder Zeile noch Negativ. `os.Getenv("ADMIN_TOKEN_FILE")` am Anfang von `lade` und als Ersatz eines leeren `pfad` bleibt grün (`C-getenv`, `C-getenv-ersatz`). Das ist die Zusage, die verhindert, dass der Adapter `SPEC-110` Nr. 2 (nur der Composition Root liest `ADMIN_*`) umgeht. | Eine Zeile T27 mit Fänger: ein AST-/Text-Test über die Dateien des Pakets ohne `_test.go`, der `os.Getenv`, `os.LookupEnv`, `os.Environ`, `os.ExpandEnv` verwirft (Probe `TestProbeKeineUmgebung`, rot gegen beide Mutationen), oder ein Negativ mit Grund und Folge. |
| V-2 | MEDIUM | Nr. 4 „`LadeToken` hält keinen Zustand“ hat keinen Fänger. Ein Cache je Pfad (Karte hinter einer Sperre) bleibt in `TestZweiterAufruf` und in allen Tests grün, weil jeder Test einen neuen Pfad nimmt; allein die ungeschützte Karte in meiner ersten Fassung fiel durch ein Datenrennen im Wettlauf auf. Folge: ein Token bliebe nach dem Entfernen oder Ändern der Datei im Prozess gültig. | Ein Fall in T22: Datei zwischen zwei Aufrufen entfernen (das Token ist ein neues) und durch einen Verstoß ersetzen (der Verstoß gilt); Probe `TestProbeKeinZustand` (rot gegen Cache mit Kanal als Sperre). |
| V-3 | MEDIUM | Nr. 3 a / `LH-FA-13.a` (mindestens 256 Bit): T14 prüft Muster, 100 verschiedene Werte, feste Werte und Länge, aber nicht die Unabhängigkeit der 32 Bytes. `E-halbe-entropie` (16 Bytes, zweimal geschrieben, 128 Bit) und `E-viertel` (4 Bytes wiederholt, 32 Bit) bleiben grün, jedes Token passt auf `^[0-9a-f]{64}$`. | Fänger als Schleife über die Positionen: 100 Token, kein Byte konstant und keines der 496 Paare in allen 100 gleich (Probe `TestProbeEntropie`, 1,3 s, rot gegen vier Mutationen: Hälfte, Viertel, Spiegel, `b[7]=7`; auf dem Stand grün). Alternative: akzeptiertes Negativ mit Grund (Entropie ist nicht beweisbar) und Folge (ein Fehler in der Erzeugung bleibt unbemerkt). |
| V-4 | LOW | Z-N2 begründet „`Link` außer `fs.ErrExist` ohne Test“ mit „`os.File` ohne Naht“. Der Zweig (`tokendatei.go:142`) ist über die vorhandene Naht erreichbar: ein `vorLink`, der die Hilfsdatei entfernt, lässt `os.Link` mit `ENOENT` scheitern. Der Fall erlaubt zugleich den Fänger zu Nr. 3 e „ein Fehler dabei (Entfernen) ohne Wirkung“ (`C-remove-fehler` grün, Probe `D-remove-fehler-panic` rot) und schließt die Allaussage von T25 für den vierten Weg zu `Datei nicht anlegbar`. Die Begründung des Negativs stimmt nicht. | Z-N2 auf `Write`, `Sync`, `Chmod` kürzen (die Fehlerzweige von `schreibe`, wirklich ohne Naht) und den `Link`-Fehler als Zeile T20/T25 mit dem Haltepunkt führen (Probe `TestProbeLinkFehler`). |
| V-5 | LOW | Nr. 1 a nennt „fehlendes Recht an einem Pfadelement“ als `Datei nicht lesbar`. Im Image läuft der Test als Superuser; `C-eacces-fehlt` (`ErrPermission` als „fehlt“) bleibt grün. Z-N1 deckt nur das Lesen (`os.ReadFile`). | Ein Negativ Z-N6 mit Grund (der Superuser des Images sieht jedes Pfadelement) und Folge (ein Betreiber ohne Recht am Verzeichnis sieht `Datei nicht lesbar`, der Zweig ist ohne Test); in der Z-Liste führen. |
| V-6 | INFO | F-6 des Reviews bleibt: der `Lstat` in `istFehlend` hat keine beobachtbare Wirkung, solange das Verzeichnis beschreibbar ist (`N-lstat-weg-F6` grün: der Symlink ohne Ziel läuft dann durch Erzeugen und `Link` mit `fs.ErrExist` und endet gleich). Nr. 1 a sagt „führt zu Nr. 3 … außer `Lstat` gelingt“, also *nicht* in das Erzeugen. | Ein Fall mit dem Haltepunkt `nachAnlegen` (`t.Error`) bei Symlink ohne Ziel; Probe `TestProbeDanglingOhneHilfsdatei`, rot gegen die Mutation. Oder ein Negativ im Plan. |

Weder die Raten noch die Mengen der Fix-Runde sind im Plan falsch: ich habe sie nachgemessen (Abschnitt 6) und sie bleiben gleich oder besser (6 von 7 nebenläufigen Mutationen 20/20 und 50/50, die siebte durch T17 deterministisch).

## 8. Akzeptierte Negative (Z-N1 bis Z-N5, Handoff 1 bis 5, T26)

| Negativ | Grund und Folge im Plan | Verifier |
|---|---|---|
| Z-N1 Lesefehler | Superuser; Zweig ohne Fänger | bestätigt (`P-readfile-unerreichbar` grün, Block `:108` Zähler 0) |
| Z-N2 `Write`, `Sync`, `Chmod`, `Link` | `os.File` ohne Naht | für `Link` nicht tragend (V-4); für `Write`, `Sync`, `Chmod` bestätigt (Block `:132`) |
| Z-N3 Windows | Image läuft auf Linux | bestätigt (kein Block mit Zähler 0; der Ausdruck wird ausgeführt: `GOOS != "linux"` rot) |
| Z-N4 `O_EXCL` | Wirkung nur bei Namenszusammenstoß, 2^-64 | bestätigt (`N-excl-weg` grün); der feste Name ist durch T18 gefangen |
| Z-N5 `Sync`, Schließen | nicht beobachtbar | bestätigt (beide grün) |
| Handoff (1) Zählung Zeilen von Nr. 6 | `spec/` fehlt in der Stufe `source` | angenommen; ich habe von Hand gezählt: sieben Gründe in acht Zeilen |
| Handoff (2) Standard-Log ohne Mutation | Kontrolle `TestErfassungSiehtAusgabe` | angenommen; eine eigene Mutation mit `log` fuhr ich nicht (Import und Aufruf an zwei Stellen) |
| Handoff (3) `cap < 512` | Abhängigkeit von `os.ReadFile` | angenommen; „keine Kopie“ rot |
| Handoff (4) `Dup3` nur `linux` | Datei trägt `//go:build linux` | bestätigt |
| T26 | `ADMIN_TOKEN_FILE` bei `slice-bootstrap-m6` | `src/bootstrap` ohne Diff, kein Aufruf von `admin` aus dem Binary |

## 9. Hinweise und Grenzen

- **Scratchpad geteilt:** ich habe zu Beginn das Unterverzeichnis `.../scratchpad/verif/` (nicht meines, schon belegt) für `gates.log`, `gen.py`, `gen.jq`, `liste-a.json`, `liste-b20.json`, `liste-b50.json`, `erg-a.txt`, `erg-b20.txt`, `erg-b50.txt` benutzt. Eine dort bereits liegende `gates.log` wurde dabei überschrieben (Zeitstempel 15:47); gelöscht habe ich nichts. Alle weiteren Dateien liegen in `.../scratchpad/verifier-admin-token/` (Kopien der Obengenannten inbegriffen). Fremde Dateien im Verzeichnis `verif/` sind unberührt.
- **Ein Python-Lauf wurde vom Hook abgewiesen** (`AGENTS.md` §3.1); die Listen entstanden mit `jq`.
- **Nicht geprüft:** Windows, die Verdrahtung beim Start (`slice-bootstrap-m6`), `SPEC-238` über Teil 1 hinaus, die Mutation mit `log.Print` (Handoff (2)), der Lauf auf einem Dateisystem ohne Hardlinks.
- **Gegengelesen (Rohbytes, `AGENTS.md` §3.18):** dieser Report auf Backslash-u-Folgen, BOM, NBSP und Nullbreiten-Zeichen mit `LC_ALL=C grep -P`: null Treffer (Lauf nach dem Schreiben).
- **Eine Fix-Runde** ist nach dieser Verifikation nicht nötig: V-1 bis V-3 haben fertige Proben (nicht im Repo, `.../scratchpad/verifier-admin-token/baum/src/adapters/driving/admin/probe_test.go`), V-4 bis V-6 sind Planzeilen oder Negative.

## Auflagen

**An den Planner (vor der Closure):** (1) V-1, V-2, V-3 schließen: je ein Test (die Proben nutzen) oder ein Negativ mit Grund und Folge, im selben Commit in §3 der Tabelle als Zeile T27 ff. und im Bericht; (2) Z-N2 berichtigen und V-5 als Z-N6 führen (V-4, V-5); (3) V-6 als Fall oder als Negativ; (4) DoD abhaken, Closure-Notiz mit Lerneintrag, Register (*Zusage ohne Fänger in Sätzen der Einleitung eines Vertrags* ist eine neue oder weitere Beobachtung), Ausgänge der Risiken; (5) `main` in den Zweig mergen und `make gates` vor dem Merge nach `main` wiederholen (der Zweig läuft hinter `main`).

**An den Architect:** keine Rückgabe offen; R7 ist entschieden und im Code und in `SPEC-239` gleich. Zu prüfen bleibt, ob Nr. 1 a „fehlendes Recht“ ein Satz sein soll, den ein Test im Image nie sehen kann (Superuser); wenn ja, steht er als Negativ im Plan (V-5).
