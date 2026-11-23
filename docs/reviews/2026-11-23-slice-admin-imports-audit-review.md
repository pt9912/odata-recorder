# Review-Report: slice-admin-imports-audit (Teil 2: Eintrag `import_veroeffentlicht` nach `POST /__admin/imports`, Warten und Abbruch, Nebenläufigkeit) — 2026-11-23

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules (Zeilen IM18 bis IM24, IN1, IN2, das Raster aus §6, die Mengen und die Sendungen (a) bis (e) aus Teil 1; Teil 1 `slice-admin-imports` liegt in `done/`)
**Gegenstand:** `b199f9af..e19858c1` (Handoff-Commit `e19858c1`, Messbasis der Marke `b199f9af`; `main` steht unverändert auf `6b6b03ae`), Zweig `slice-admin-imports-audit`
**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-23

**Eingangs-Kontext:** Plan `slice-admin-imports-audit` (§1 Sendungen (a) bis (e), §2 DoD, §3 Tabelle *Zusage · Fänger* IM18 bis IM24, IN1, IN2, Mengen, Absatz *Mutationen und Raten*, akzeptierte Negative (3), (4) und IM24, §4 Block der Größenmarke, §6 Raster und Risiken, §8); Plan `slice-admin-imports` (§3 Mengen, §7 Closure, akzeptierte Negative (1) bis (11)); `SPEC-297` Nr. 9 bis 12, `SPEC-234` Nr. 2, 4 d, 5 und 7, `SPEC-238` Nr. 5 und 15 a bis d, `SPEC-136` Nr. 7 a; `LH-FA-18`, `LH-FA-13`, `LH-RB-04`; `AGENTS.md` §3.7 und §3.9 bis §3.18; Reviews der Imports, der Workspaces, der Inspektion und des Replay als Folie für die Suche.

Kein Review gegen die DoD (Aufgabe des Verifiers). Der Bericht des Implementers (Modell-Text) ist nicht übernommen: jede Zahl unten ist nachgemessen oder als nicht nachgemessen gekennzeichnet. Eigene Mutationen liefen auf Wegwerfkopien im Scratchpad, je Mutation ein `go test -count=1` über die Pakete `admin`, `support` und `driven/audit` im Image `odata-recorder:test` (`--memory=4g --cpus=4`, ohne Netz, Mount nur lesend, unter `flock /tmp/odata-recorder-heavy.lock`, nie zwei Container zugleich, nie neben `make gates`); die Raten der Klasse `nebenläufig` mit einem Testbinary je Prozess (`-test.count=1`, Image `odata-recorder:race`, Container `--cpus=4`, `GOMAXPROCS` gleich 4 und gleich 1). Unter `make gegenproben` und am Stempel wurde nichts angefasst.

## Findings

Befunde ohne Lösungsvorschlag nach dem Schema des Skills; die Klassen stehen unter *Summary*. Kein HIGH: der Produktcode (`imports.go`, `support/audit.go`) stimmt mit `SPEC-297` Nr. 9 bis 12 und `SPEC-234` Nr. 2 überein (Abschnitt *Negativbefunde*); die drei MEDIUM betreffen Fänger, die eine Form der Eingabe nicht führen, und ein akzeptiertes Negativ, das der Plan nennt und nicht trägt.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `SPEC-297` Nr. 10 a sagt den Eintrag „bei 201 und bei 200, auch wenn der `ctx` inzwischen endete“ zu; die 40 Zellen von `TestIM19Kontext` führen jeden Erfolg mit `Neu` wahr (Standard von `importsUmgebung`, Status 201), keine mit `Neu` falsch (200) bei beendetem `ctx`. Vier Mutationen bleiben grün (4 von 4): kein Eintrag bei beendetem `ctx` und `Neu` falsch (`x04`), nur bei `Canceled` und `Neu` falsch (`x06`), nur bei abgelaufener Frist und `Neu` falsch (`x06b`), Abbruch nach dem Erfolg bei beendetem `ctx` und `Neu` falsch (`x26`); die Gegenseite mit `Neu` wahr (`x05`) ist rot. | `AGENTS.md` §3.15 (Achse `Neu` × Ende des `ctx`), `SPEC-297` Nr. 9 d, 10 a | `src/adapters/driving/admin/imports_warten_test.go` · `{"Erfolg", nil, true},` | ja — `go test` über `admin` mit `x04`, `x06`, `x06b` oder `x26` bleibt grün | Fänger führt nur eine Form der Eingabe (Achse `Neu` × Ende des `ctx`) |
| F-2 | MEDIUM | Die Zusage „ein Kontextfehler beim Aufruf des Dienstes bricht ab, ohne Eintrag“ gilt für `Canceled` und `DeadlineExceeded` bloß, mit `%w` verpackt und in `errors.Join` (6 Formen je Route); die Fänger führen 3 davon (beide bloß, `Canceled` im Verbund mit `ODR-E7003`, `allgemeineAusgaenge` aus Teil 1; `TestIM19Kontext` nur bloß). Fünf von sieben Mutationen der Klasse bleiben grün: Eintrag bei mit `%w` verpacktem `Canceled` (`xf07w` bei `POST`, `xf08i` bei `GET`), bei mit `%w` verpacktem `DeadlineExceeded` (`xf08w`, `xf08g`) und bei `errors.Join` mit `DeadlineExceeded` (`xf08j`); die zwei mit `errors.Join` und `Canceled` (`xf07j`, `xf08h`) sind rot (`TestIM13Fehler`, `TestIM21KeinEintrag`). | `AGENTS.md` §3.15 und §3.17 (Allaussage über Arten als Schleife über beide Mengen), `SPEC-297` Nr. 9 c, 10 c | `src/adapters/driving/admin/imports_warten_test.go` · `{"Canceled", context.Canceled, false}, {"Canceled mit Wert", context.Canceled, true},` | ja — `go test` über `admin` mit `xf07w`, `xf08w`, `xf08j`, `xf08g` oder `xf08i` bleibt grün | Allaussage über Arten führt nicht jede Form (Kontextfehler × Verpackung) |
| F-3 | MEDIUM | Der Plan nennt zehn Mutationen an Elementen, die kein Fänger führt („q01 bis q07, n02, n03 und n01 als Gegenprobe der Schwelle; Maße unten“); unten steht allein das Maß von (3) für n01 bis n03. Für die Kennungen des Eintrags (Länge über 65 Bytes, Mischung von Bytes in einem Wert) und für die Art des Fehlers von `Append` (`q05`) steht kein akzeptiertes Negativ mit Grund, Folge und Maß; die Prüfgrenze steht nur im Kommentar von `TestIM22KennungenUnveraendert`. Eigene Mutationen an Elementen jenseits der Prüfgrenze bleiben grün: Kennung gekürzt auf 66 (`k03b`) und 128 Bytes (`k03`), `--` aus der `ServiceID` (`k05`) und `..` aus der `ImportID` entfernt (`k06`), 4 von 7 Kennungsformen (`k01`, `k02`, `k04` rot); `os.ErrClosed` (`r16a`) und der nil-Zeiger `*inbound.Fehler` (`r16b`) von `Append` als Fehler der Antwort, 2 von 3 Fehlerformen (`r16c` mit Status 404 rot). | `AGENTS.md` §3.15 (akzeptiertes Negativ nennt Grund, Folge und Maß n von m) | `docs/plan/planning/in-progress/slice-admin-imports-audit.md` · `q01 bis q07, n02, n03 und n01 als Gegenprobe der Schwelle; Maße unten` | ja — `go test` über `admin` mit `k03`, `k03b`, `k05`, `k06`, `r16a` oder `r16b` bleibt grün | Akzeptiertes Negativ ohne Grund, Folge und Maß (Verweis ohne Ziel) |
| F-4 | LOW | Das Maß von (4) nennt für das Fenster ohne Fuge „(g06) 1 von 1 ungefangen“; die Route hat sechs solche Fenster (Eintritt bis Ende des Lesens, Ende des Lesens bis `Import`, nach `Import` vor `Append`, nach `Append`, vor und nach `Get`), und bei einem Prozessor bleiben alle sechs grün: 6 von 6 (je 0 von 100 Prozessen, `s05b`, `s05`, `s04`, `s05c`, `s07b`, `s07`). Bei vier Prozessoren sind sie bei Last 4,5 bis 12 rot (je 100 von 100), bei Last 15 bis 19 18 von 20 (`s04`). | `AGENTS.md` §3.15 (Maß n von m der ungefangenen Klasse) | `docs/plan/planning/in-progress/slice-admin-imports-audit.md` · `das Fenster ohne Fuge (g06) 1 von 1 ungefangen` | ja — Treiber im Scratchpad, Raten in *Auflösung* | Maß des akzeptierten Negativs schmaler als die Klasse |
| F-5 | LOW | Die Abweichungen (a) bis (c) von den Zahlen des Plans stehen in §3; §6 trägt keine Zeile mit Ausgang *offen* und Adresse. (c) weicht von der Zeile IN1 ab („Zelle unter einer Sperre in jedem der 6 Fenster (6 Mutationen)“; geliefert sind 5 Mutationen mit 4 Platzierungen), ohne dass der Architect sie entschieden hat. | `AGENTS.md` §3.13 | `docs/plan/planning/in-progress/slice-admin-imports-audit.md` · `Maß: 2 Paare von 2 untrennbar` | nein — Durchsicht des Plans | Abweichung vom Plan ohne Zeile in §6 |
| F-6 | LOW | Die Zählprüfung „6 gleich der Zahl der Haltepunkte des Fakes und des Audit“ (Plan, M-Fenster) vergleicht die sechs Einträge der Karte im Test mit der Zahl 6; weder `impAnzahl` noch `vor`, `nach`, `auditNach` oder `ende` gehen ein, ein weiterer Haltepunkt bliebe ohne Fenster, ohne dass die Prüfung rot wird. | `AGENTS.md` §3.15 (Zählprüfung gegen die Quelle) | `src/adapters/driving/admin/imports_nebenlaeufig_test.go` · `gleichAuf(t, "Zahl der Fenster", len(fenster), 6)` | nein — Durchsicht des Tests | Zählprüfung vergleicht Literal mit Literal |
| F-7 | INFO | IM24 nennt „2 von 2 Fenstern ohne Fänger“ (Kontextfehler in `SetMetadata`, Prozessende zwischen `Import` und `Append`) und trägt sie ehrlich; ein drittes Fenster steht in der Tabelle von `SPEC-297` Nr. 10 c, nicht in der *Folge im Betrieb*: ein ungültiger Report nach gelungenem `Import` ist 500 `ODR-E9001`, die Baseline steht, es gibt keinen Eintrag (Fänger IM21, rot mit `x42a`). Keine Aktion erwartet; Rolle: Architect. | Maintainability | `docs/plan/planning/in-progress/slice-admin-imports-audit.md` · `Maß: 2 von 2 Fenstern ohne Fänger` | nein | Folge im Betrieb nennt nicht jedes Fenster |
| F-8 | INFO | Zwei grüne Mutationen liegen auf den Prüfgrenzen, die der Plan nennt: Semaphor von 32 um `Import` (`t02c`, gleich der Zahl der gleichzeitigen Requests je Route von (3); 16 und 31 sind rot) und Eintrag bei einem ungültigen Report über 4 MiB (`x42b`; über 64 KiB ist rot, `x42a`; Prüfgrenze 1 MiB + 1 aus (1) von Teil 1). Das Maß von (3) trägt n01 bis n03 (Zahl der Requests), nicht die Obergrenze der Gleichzeitigkeit. Keine Aktion erwartet. | Maintainability | `docs/plan/planning/in-progress/slice-admin-imports-audit.md` · `Prüfgrenze 32 zugleich je Route in 150 Runden` | ja — `t02c`, `x42b` | Prüfgrenze ohne Maß je Achse |

## Auflösung der Tabelle *Zusage · Fänger* (IM18 bis IM24, IN1, IN2, je Zeile eigene Mutationen)

Skripte, Listen und Ausgaben: `/tmp/claude-1000/-Development-odata-recorder/efb0f904-f3aa-4d9d-a147-ca0b12313d6a/scratchpad/admin-imports-audit-reviewer/` (`specA.txt` bis `specF.txt`, `specAll.txt`, `specAll2.txt`, Ausgaben `outAll_1.txt`, `outAll2.txt`, `outAll3.txt`, Raten `outRaten.txt`, `outRaten100.txt`, Abdeckung `cover.out`). **67 Mutationen an Elementen, die der Implementer nicht nannte; 50 rot, 17 grün.** Von den 17 grünen sind 15 die Lücken von F-1 bis F-3 (F-1 vier, F-2 fünf, F-3 sechs) und 2 Prüfgrenzen (F-8); zum Vergleich fanden die Vorgänger-Reviews 31 von 55, 38 von 86, 20 von 52 und 4 von 47. Dreizehn frühere Läufe (`r04` bis `r12` und `x07w` bis `x08g` der ersten Fassung) sind verworfen: sie riefen `Err` auf dem `ctx` (verboten von `TestIM15QuelleDerRouten`), zählten eine zweite Stelle von `Append` (`TestAppendEineAufrufstelle`) oder stürzten selbst auf dem nil-Zeiger `*inbound.Fehler` ab; die zweite Fassung (`x…`, `xf…`) ruft eine Hilfsmethode und eine Hilfsdatei, sodass allein das Verhalten rot oder grün wird.

| Zeile | Meine Mutation (Wegwerfkopie) | Ergebnis |
|---|---|---|
| IM18 | `r22` (Frist von 24 h um `Import`), `r22b` (um `Get`), `r20` (abgeleiteter `ctx` um `Import`), `r21`, `r21b` (Wert im `ctx` um `Import`, um `Get`) | rot (`TestIM18Warten`, `TestIM3AufrufeUndKontext`) |
| IM19 | `x05` (kein Eintrag bei beendetem `ctx` und `Neu` wahr), `x25`, `x25b` (Abbruch nach dem Erfolg nur bei Frist, `POST` und `GET`), `x27` (nur bei `Canceled`, `GET`), `x12` (`GET` schreibt bei beendetem `ctx`) | rot (`TestIM19Kontext`) |
| IM19 | `x04`, `x06`, `x06b`, `x26` (siehe F-1); `xf07w`, `xf08w`, `xf08j`, `xf08g`, `xf08i` (siehe F-2) | grün (4 von 4, 5 von 7) |
| IM20 | `r01` (`ServiceID` aus dem Body), `r03` (Eintrag nur bei 200); `k01` (`ServiceID` getrimmt), `k02` (`ImportID` ohne ungültiges UTF-8), `k04` (`ab` aus der `ImportID`) | rot (`TestIM20Eintrag`, `TestIM22KennungenUnveraendert`, `TestIM10Status`) |
| IM21 | `x10` (Eintrag bei Fehler mit Wert), `x11` (bei Status 422), `x42a` (bei ungültigem Report über 64 KiB), `x43` (`GET` bei Report `null`), `x45` (Lesefehler `ErrUnexpectedEOF`), `x46` (Lesefehler außer `ErrUngueltig`), `x47` (`GET` bei 404), `x48` (`GET` mit Wert und Fehler), `xf07j`, `xf08h` (`errors.Join` mit `Canceled`) | rot (`TestIM21KeinEintrag`, `TestIM13Fehler`, `TestIM13Lesen`, `TestIM12Stelle`, `TestIM11Report`) |
| IM22 | `r16c` (Fehler von `Append` mit Status 404 als Fehler der Antwort) | rot (`TestIM22AppendFehler`) |
| IM22 | `r16a` (`os.ErrClosed`), `r16b` (nil-Zeiger `*inbound.Fehler`) als Fehler der Antwort; `k03`, `k03b`, `k05`, `k06` (siehe F-3) | grün (2 von 3, 4 von 7) |
| IM23 | `u01` (Pflichtfelder vertauscht), `u02` (Tippfehler im Text der Art), `u04` (`service_id` nicht Pflicht), `u05` (`workspace_id` Pflicht), `u06` (Zeile doppelt), `u07` (`error_code` Pflicht) | rot (`TestArtenGleichTabelle`, `TestArtenLiefertKopie`, `TestArtTexte`, `TestFeldBelegung`, `TestZeileImportVeroeffentlicht`, `TestKennungsform`, `TestLogBeiFehler`) |
| IN1 | `s01` (`atomic.Value`, Start bis nach `Import`), `s02` (`sync.Map`, vor `Import` bis nach `Import`), `s03` (Kanal), `s06` (`GET`, `atomic.Value` vor bis nach `Get`), `s04`, `s05`, `s05b`, `s05c`, `s07`, `s07b` (Zelle ohne Fuge in den sechs Fenstern), `s08` (geteilter Puffer in `reportAntwort`), `s09` (Paketvariable für die `ImportID`) | rot (`TestIN1GleichzeitigeRequests`, Raten unten) |
| IN1 | `s10` (`GET` aus geteiltem Zwischenspeicher je Pfadwert), `s12` (`POST` aus Zwischenspeicher je `import_id`), `s13` (kein Eintrag je Paar `import_id` und Report) | rot (`TestIM11…`, `TestIM10Status`; Zustand zwischen Requests trägt Teil 1) |
| IN2 | `t01` (Versuchssperre um `Append`), `t02` (Semaphor 16 um `Import`), `t02b` (31), `t03` (eine Sperre um `Import` und `Get`) | rot (`TestIN2KeineSperreUeberDemAufruf`, `TestIN1GleichzeitigeRequests`, `TestIN1ZweiGleichzeitigeGleicheImportID`) |
| IN2 | `t02c` (Semaphor 32 um `Import`) | grün: Prüfgrenze von (3), F-8 |

**Raten der Klasse `nebenläufig`** (eigene Messung, ein Prozess je Lauf, `-test.count=1`; Messender ein Shell-Treiber im Scratchpad, ein Testbinary je Prozess nacheinander; Container `--cpus=4`, `GOMAXPROCS` gleich 4 und gleich 1; Last des Hosts je Zeile als 1-Minuten-Mittel vor und nach der Zeile, Host 20 Prozessoren; die Zahl steht als rot / Prozesse und gilt als Größenordnung). Gegenprobe ohne Mutation: 0/20 ohne und 0/20 mit Detektor bei vier Prozessoren (Last 9,6 bis 14,8).

| Mutation | Fenster | 4 Prozessoren ohne Detektor | 1 Prozessor ohne Detektor | Mit Detektor (4 und 1) |
|---|---|---|---|---|
| `s05b` Zelle ohne Fuge | Eintritt bis Ende des Lesens | 100/100 (Last 4,6 bis 5,9) | **0/100** (Last 5,9 bis 5,7) | — |
| `s05` Zelle ohne Fuge | Ende des Lesens bis `Import` | 100/100 (Last 5,8 bis 6,3) | **0/100** (Last 5,8 bis 4,6) | — |
| `s04` Zelle ohne Fuge | nach `Import`, vor `Append` | 100/100 (Last 8,3 bis 8,0); 18/20 bei Last 15 bis 19 | **0/100** (Last 8,0 bis 6,3), 0/20 | — |
| `s05c` Zelle ohne Fuge | nach `Append` | 100/100 (Last 5,7 bis 6,2) | **0/100** (Last 6,2 bis 6,4) | — |
| `s07` `GET`, Zelle ohne Fuge | nach `Get` | 100/100 (Last 6,4 bis 7,9) | **0/100** (Last 7,9 bis 7,7) | — |
| `s07b` `GET`, Zelle ohne Fuge | vor `Get` | 100/100 (Last 7,7 bis 9,4) | **0/100** (Last 9,4 bis 11,9) | — |
| `s01`, `s02`, `s03`, `s06` Zelle über einen Treffpunkt | Start bis nach `Import`; `Get` | je einmal rot in der Suite (vier Prozessoren) | je 20/20 (Last 6 bis 12; `s06` bis 48 am Ende) | — |
| `s08` geteilter Puffer | `reportAntwort` | 20/20 (Last 9 bis 10) | 20/20 (Last 9) | 20/20 und 20/20 |
| `s09` Paketvariable | nach `Import`, vor `Append` | 6/20 (Last 9 bis 11) | **0/20** (Last 11) | 20/20 und 20/20 |

Lesart: Über einen Treffpunkt (`s01`, `s02`, `s03`, `s06`) ist die Zelle auch bei einem Prozessor in jedem Lauf rot; jede Zelle, die zwischen zwei Treffpunkten schreibt und zurückliest, ist bei einem Prozessor in keinem Lauf rot und bei vier Prozessoren bei Last unter 12 in jedem. Die Behauptung „ein Treffpunkt in jedem Fenster“ (Plan, Zeile IN1) gilt für Paare, die ein Treffpunkt trennt; die Fenster zwischen dem Portaufruf und dem nächsten Schritt der Route im Einzelnen:

| Fenster der Route | Treffpunkt | Zelle über das Fenster hinweg | Zelle innerhalb des Fensters |
|---|---|---|---|
| Eintritt bis Ende des Lesens (`POST`) | Ende des Bodys (1) | rot (`s01`) | ungefangen bei einem Prozessor (`s05b`) |
| Ende des Lesens bis Aufruf von `Import` | im Aufruf (2) | rot (`s02`, `s03`) | ungefangen bei einem Prozessor (`s05`) |
| im Aufruf von `Import` | beide Treffpunkte im Dienst (2, 3) | nicht Code der Route; keine Sperre: IN2 | — |
| nach `Import`, vor `Append` (auch in `Append`) | nach der Rückkehr (3), nach `Append` (4); keiner vor `Append` | rot, sobald ein Treffpunkt zwischen Schreiben und Lesen liegt (`s09` mit Detektor) | ungefangen bei einem Prozessor (`s04`) |
| nach `Append`, vor der Rückkehr des Handlers | nach `Append` (4) | rot (`s02`) | ungefangen bei einem Prozessor (`s05c`) |
| Rückkehr des Handlers bis Schreiben der Antwort | Rahmen, nicht dieser Teil | — | — |
| Eintritt bis `Get` und `Get` bis Rückkehr (`GET`) | im Aufruf (5), nach der Rückkehr (6) | rot (`s06`) | ungefangen bei einem Prozessor (`s07b`, `s07`) |

Der Detektor von `make test-race` fängt `s08` und `s09` in jedem Prozess (20/20 und 20/20); ohne Detektor ist `s09` bei vier Prozessoren 6 von 20 und bei einem 0 von 20.

**Abdeckung** (eigener Lauf `go test -race -count=1 -coverprofile`, Image `odata-recorder:race`): Paket `admin` 99,5 %, in `imports.go` 20 Blöcke, null mit Zähler 0; Paket `support` 100,0 %, null Blöcke mit Zähler 0. Die vier Blöcke von `driven/audit` (95,7 %) stehen im Altcode und wurden nicht nachgemessen.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driving/admin/imports.go` (Handler `importiere` und `lies`) | geprüft gegen `SPEC-297` Nr. 9 bis 12, ohne Befund: ein Aufruf von `Append` nach dem Erfolg und dem gültigen Report, `ServiceID` und `ImportID` aus dem Ergebnis des Dienstes, `KorrelationsID` der Antwort, Ergebnis ignoriert, kein `ctx` für `Append`, keine Frist, keine Ableitung, keine Goroutine (`r20`, `r21`, `r22`, `x05`, `x12` rot) |
| `src/hexagon/ports/support/audit.go`, Zählstellen der Arten | geprüft, ohne Befund: achte Zeile mit `service_id` und `import_id`, sechs Mutationen rot; die Zahl der Arten steht in vier Testdateien (`audit_test.go` des Ports und des Adapters, `log_sperre_test.go`, `audit_betrieb_test.go`), `grep` nach `siebe`, `sieben`, `7 Arten` und `!= 7` findet in `src/`, `tools/` und `harness/` keine weitere |
| `src/bootstrap` | `git diff --stat b199f9af..e19858c1 -- src/bootstrap` nennt allein `audit_betrieb_test.go` (9 Einfügungen, 9 Löschungen); das Binary bleibt unverändert |
| Sendungen (a) bis (e) aus Teil 1 | geprüft, ohne Befund: die Mengen sind Funktionen im Testpaket und werden von `TestIM21KeinEintrag` mit Zählprüfung gelesen (86, 512 mit 8 gültigen, 24, 305, 44 × 3 × 6 und 7 + 44 × 3 × 4 Reports in 3 Zuständen); die 208 Zellen von `TestIM13Lesen` sind nicht extrahiert und zählen die Einträge am Ort (`x45`, `x46` rot in `TestIM13Lesen`); die vier `POST`-Zellen von IM3 zählen 1; `imports_quelltext_test.go` bleibt unverändert und grün, `Append` braucht keine verbotene Stelle |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft, ohne Befund: der Diff der Quellen enthält keinen Konjunktiv über verworfene Alternativen und keine Chronik; die Köpfe der neuen Tests nennen Zellen und Mengen, die sie führen |
| Rohbytes und Escape-Folgen (`AGENTS.md` §3.18) | geprüft: die neuen Zeilen tragen keine Folge Backslash, `u` und vier Hex-Ziffern; die Zeichen der Mengen entstehen aus `\x`-Literalen oder `[]byte{…}` |
| Messung (`SPEC-128`) | nachgemessen: `git diff --numstat b199f9af..e19858c1` über die 14 Dateien des Blocks 776 (Summe der Spalten), Marke 1100, fremde Zeilen 0; Produktcode 17 hinzugefügt, 5 ersetzt |
| `make a-check` | Exit 0, 0 Befunde (der Hinweis auf 97 Dateien unter `tools/` ohne Schicht steht schon auf `main`) |
| Tests auf dem unveränderten Baum | `go test -count=1` über `admin` (20,7 s), `support` und `driven/audit` grün; `make gates`, `make test-race` (mit Last der anderen Linien) und die Gegenproben nicht gefahren |
| Plan-Nachzug (`AGENTS.md` §3.9) | Kopf (`Bezug`, `Berührte Spec-Stellen`), §1, §3, §4 und §6 tragen die Änderungen; offen allein F-5 |
| IM24 (kein Fänger, 2 von 2) | geprüft: das Maß ist ehrlich, die dritte Lage (Fehler von `Append`) fängt IM22; siehe F-7 |
| Vertrag `SPEC-297` Nr. 12 (Geheimnisse) | geprüft, ohne Befund: der Eintrag trägt allein die Felder von Nr. 10; `r01` (Body statt Dienst) und `x45`, `x46` (Eintrag bei Verstoß vor dem Aufruf) sind rot |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 3 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe (Achse `Neu` × Ende des `ctx`) · Allaussage über Arten führt nicht jede Form (Kontextfehler × Verpackung) · Akzeptiertes Negativ ohne Grund, Folge und Maß (Verweis ohne Ziel) · Maß des akzeptierten Negativs schmaler als die Klasse · Abweichung vom Plan ohne Zeile in §6 · Zählprüfung vergleicht Literal mit Literal · Folge im Betrieb nennt nicht jedes Fenster · Prüfgrenze ohne Maß je Achse

## Verdikt

**Merge-blockierend:** ja — die drei MEDIUM (F-1 bis F-3) sind Fänger oder Maße, die der Fix-Slice nachzieht; kein HIGH, der Produktcode ist vertragstreu und die Mutationen an Eintrag, Warten, Abbruch und Nebenläufigkeit sind bis auf die benannten Formen rot (50 von 67).

**Übergabe:** Findings gehen an den Implementer; die Finding-Klassen gehen in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ersetzt keine Verifikation — DoD- und Spec-Konformität prüft der Verifier.
