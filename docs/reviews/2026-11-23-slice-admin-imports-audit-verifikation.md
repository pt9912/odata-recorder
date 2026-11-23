# Verifikations-Report: slice-admin-imports-audit (Teil 2: Eintrag `import_veroeffentlicht`, Warten und Abbruch, Nebenläufigkeit) — 2026-11-23

**Verifikations-Art:** DoD- und Entscheidungs-Konformität plus Plan-vs-Code-Diff nach Modul 11 (Zeilen IM18 bis IM24, IN1, IN2; die Fix-Runde `ab37b3a2` zu den Befunden F-1 bis F-8 des Reviews, gegen die Klasse aufgelöst, `AGENTS.md` §3.17). Kein Review des Diffs (Aufgabe des Reviewers, Report `2026-11-23-slice-admin-imports-audit-review`) und keine Validierung gegen den Bedarf.
**Gegenstand:** `b199f9af..21bc3f29` (Messbasis der Marke `b199f9af`), Zweig `slice-admin-imports-audit`, Worktree `/Development/odata-recorder-wt-admin-imports-audit`; Commits des Slice: `e19858c1` (Handoff), `d8bcee1d` (Review-Report), `ab37b3a2` (Fix-Runde: Tests und Plan), `521b5fe8` (Merge von `main` `0feeb705`), `21bc3f29` (Nachzug des Plans und ein Kommentar).
**Eingang:** Plan `slice-admin-imports-audit` (alle Abschnitte), Review-Report (F-1 bis F-8), `SPEC-297` Nr. 9 bis 12, `SPEC-234` Nr. 2, 4 d, 7, `SPEC-238` Nr. 9 und 15, `SPEC-136` Nr. 7 a, `AGENTS.md` §3.9 bis §3.18, Bezug [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `git diff b199f9af..HEAD` (Produktcode `imports.go` 12 Einfügungen und 3 Löschungen, `audit.go` des Ports 5 und 2).

Die Behauptungen des Implementers sind nicht übernommen. Gates, Messung, Abdeckung und alle Mutationen unten sind von mir gefahren; was ich nicht nachfuhr, steht unter *Nicht nachgemessen*. **Messender, Host, Last:** Verifier (ein Agent, Modell Sonnet 5.5); Host mit 20 Prozessoren; Mutationen und Raten in Containern `--cpus=4 --memory=4g --memory-swap=4g` ohne Netz, mit einer Wegwerfkopie je Mutation als schreibgeschütztem Mount (`:ro`, Scratchpad `…/scratchpad/admin-imports-audit-verifier/`, Listen `spec*.txt`, Ausgaben `sweep*.log` und `raten*.log`), alle unter dem `flock` `/tmp/odata-recorder-heavy.lock`, nacheinander, nie parallel zu `make gates`; Last des Hosts je Zeile als `last=`: 4,4 bis 26,3 bei den Verhaltens-Mutationen (zwei fremde Linien liefen mit), 3,7 bis 12,0 bei den Raten. `make gates` startete bei Last 7,6 (Load der letzten Minute) und 10,7 GB verfügbarem Speicher (Schwelle: Last unter 15 und mindestens 6 GB).

## Urteil

**Freigegeben unter Auflagen.** Kein HIGH, ein MEDIUM (V-1), zwei LOW, vier INFO. Der Produktcode (`imports.go`, `audit.go` des Ports) folgt `SPEC-297` Nr. 9 bis 11 und `SPEC-234` Nr. 2 in jedem Punkt, den ich gegen den Text gehalten habe (Abschnitt 2); `make gates` ist grün, `pruefung test` (362 s) und `pruefung race` (402 s) sind frisch gelaufen und nicht aus dem Schicht-Cache, die Messung ist 1075 von 1100 mit null fremden Zeilen, in `imports.go` ist kein Block ungelaufen, und `src/bootstrap` trägt allein `audit_betrieb_test.go`.

Die Fix-Runde deckt die Klassen des Reviews: **75 eigene Mutationen an Elementen, die weder Review noch Implementer nannten** (Abschnitt 3), davon 43 aus dem benannten Grund rot, 9 rot allein durch den Quelltest oder durch eine Panik an anderer Stelle (nicht gezählt), 18 grün und 5 Fenster ohne Fuge (Abschnitt 4: bei vier Prozessoren rot, bei einem grün). Die 18 grünen liegen auf zwei Gruppen: (a) 10 auf Elementen jenseits der benannten Prüfgrenzen (3), (5), (6), (8), und die Maße des Plans stimmen (Semaphor 31 rot, 32 und 40 grün; ungültiger Report ab 1 MiB + 1 rot, ab 1048700 grün; Fehler von `Append` per `==` und per Typ grün); (b) 8 auf Elementen, die kein Negativ nennt: die Verpackung eines Kontextfehlers ab Tiefe 3 und im `Join` mit drei Teilen (V-1, MEDIUM), ein Fehler anderer Art und ein Fehler von `Append` bei beendetem `ctx` (V-2, LOW) und ein Ende des `ctx` erst in `Append` (V-5, INFO). Das Fenster ohne Fuge bei einem Prozessor (Negativ (4)) ist mit 5 von 5 weiteren Formen ungefangen ehrlich benannt.

**Auflagen an den Planner** (Plan-Text, kein Code; der Verifier hat den Plan nicht geändert): A-1 bis A-4 am Ende dieses Reports; A-1 ist vor dem Move nach `done/` zu erledigen (`AGENTS.md` §3.15), keine Auflage blockiert die Übergabe.

## Findings

| ID | Kategorie | Befund | Ort | Beleg | Klasse |
|---|---|---|---|---|---|
| V-1 | MEDIUM | Die Menge M-Kontextfehler (24 Formen, 288 Zellen) ist in zwei Achsen unbeschränkt: Tiefe der Verpackung und Breite eines `errors.Join`. Der Plan nennt keine Prüfgrenze (kein Negativ unter (3) bis (8)), `AGENTS.md` §3.17 verlangt sie mit Grund und Folge. Vier Mutationen an Elementen jenseits der 24 Formen bleiben grün (4 von 4): ein Kontextfehler hinter drei `%w` (`v-h03` `POST`, `v-h04` `GET`; die Menge führt bis zwei), in der Mitte eines `Join` aus drei Fehlern (`v-h05`) und der `Join` aus `Canceled` und `DeadlineExceeded` zugleich (`v-h06`; die Menge führt je einen Kontextfehler). Die Kontrolle `v-h07` an einem Element der Menge (zwei `%w`) ist rot (`TestIM19Kontextfehler`). Die Route prüft den Kontextfehler nicht selbst (`SPEC-238` Nr. 9 gehört dem Rahmen), die Folge im Betrieb ist klein; aber die Zusage „jede Form“ steht ohne Grenze | `docs/plan/planning/in-progress/slice-admin-imports-audit.md` §3, Zeile F-2 der Fix-Runde und *Unbeschränkte Mengen* · `src/adapters/driving/admin/imports_warten_test.go` `kontextfehlerFormen` | ja — `v-h03` bis `v-h06`, Abschnitt 3.2 | Unbeschränkte Klasse ohne akzeptiertes Negativ |
| V-2 | LOW | Ein Fehler des Dienstes anderer Art bei beendetem `ctx` und ein Fehler von `Append` bei beendetem `ctx` haben keinen Verhaltens-Fänger: `v-h01` (`POST`) und `v-h02` (`GET`) machen aus einem Fehler mit Code bei beendetem `ctx` einen Kontextfehler, `v-h08` aus dem Fehler von `Append` bei beendetem `ctx` ein 500, alle drei grün (3 von 3). Dieselben Mutationen in `imports.go` selbst (`v-k10`, `v-k11`, `v-a07`) sind rot allein durch `TestIM15QuelleDerRouten` (Verbot von `Err`, `Done`, `Deadline`, `Cause` nach Namen in `imports*.go`); eine Hilfsfunktion in einer anderen Datei des Pakets umgeht den Namen. Die Zusatztabelle führt die Paare (Fehler des Dienstes × beendeter `ctx`) und (Fehler von `Append` × beendeter `ctx`) nicht (`AGENTS.md` §3.16); die Zusage „unverändert“ (`SPEC-297` Nr. 7, 10 e) gilt dort ohne Test mit Wirkung | `…slice-admin-imports-audit.md` §6, Zusatztabelle · `imports_warten_test.go` `TestIM19Kontext` (Ausgänge) | ja — `v-h01`, `v-h02`, `v-h08` | Paar von Verstößen ohne Fänger |
| V-3 | LOW | Plan-Nachzug (`AGENTS.md` §3.9) unvollständig: (a) Zeile IM21 der Tabelle nennt „die 35 von IM19“; die Menge M-Kontext hat 66 Zellen (Abweichung (a)), und die Zellen von IM19 zählt `TestIM19Kontext` mit demselben Orakel, nicht `TestIM21KeinEintrag`: die Zahl ist die der älteren Fassung. (b) §6, Punkt *Fremde Zeilen*, trägt „Stand des Implementers: `main` unverändert auf `6b6b03ae`“; nach dem Merge `521b5fe8` steht `main` auf `0feeb705` (die Messpunkt-Zeile von §4 nennt es richtig). (c) Die benannte Ausgabe von IM23 im *Messpunkt beim Handoff* („`git diff --stat b199f9af -- src/bootstrap` nennt allein `audit_betrieb_test.go`“) gilt nach dem Merge nur gegen den `main`-Elter: gegen `b199f9af` nennt sie 7 Dateien (6 fremde aus `slice-telemetry-einbau-replay`), gegen `0feeb705` allein `audit_betrieb_test.go` (9 und 9; von mir gemessen). Sendung (d) schreibt den `main`-Elter vor, der Messpunkt nach der Fix-Runde nennt die Ausgabe nicht. (d) Der Kopf von `TestIM16BinaryUnveraendert` (Teil 1, unverändert) sagt „`git diff --stat` über `src/bootstrap` ist in diesem Teil leer“: der Satz gilt für Teil 1, in Teil 2 trägt `audit_betrieb_test.go` die Zählstelle | `…slice-admin-imports-audit.md` Tabelle IM21, §6, §4 · `imports_quelltext_test.go` Kopf | nein — Durchsicht, `git diff --stat` | Nachzug nach Berichtigung nicht an jeder Stelle |
| V-4 | INFO | Die Maße von (5) und (6) zählen gewählte Mutationen, nicht die Klasse. Zu (5) (3 von 17 grün): drei weitere Formen jenseits der Zweierfolgen sind grün (3 von 3): `-` zwischen zwei Ziffern entfernt (`v-n05`, eine Dreierfolge), `NUL` nur ab 65 Bytes entfernt (`v-n06`, Länge × Zeichen) und die Länge genau 4000 gekürzt (`v-n07`, eine Länge zwischen den Zweier- und Zehnerpotenzen); `v-n08` (gekürzt auf 1048600) ist grün wie `2 MiB` im Plan; `v-n01` bis `v-n04`, `v-n09` und `v-n10` (Kürzung auf 5000, 131072 und 1048576, Kleinschreibung, `.` zwischen Alphanumerischen, Großschreibung ab 8 Zeichen) sind rot. Die Folge im Betrieb im Plan nennt Dreierfolgen und die Mischung, nicht die Punktlänge. Zu (6): `v-a01` (`== syscall.EPIPE`) und `v-a06` (Typ `syscall.Errno`) sind grün (2 von 2, wie zugesagt); der Text einer Kennung („Text `boom` grün“) hängt an der Bauart der Mutation: `v-a03` ruft `Error()` und wird rot, weil das Element „nil-Zeiger `*inbound.Fehler`“ der Menge dabei in die Panik läuft (der Fänger ist stärker als gesagt, das Maß „grün“ gilt für eine Mutation mit Schutz vor dem nil-Zeiger) | `…slice-admin-imports-audit.md` §3 (5), (6) | ja — Abschnitt 3.3 | Maß schmaler als die Klasse |
| V-5 | INFO | Zu M-Kontext: die drei Zeitpunkte „vor dem Aufruf“, „im Aufruf“ und „nach der Rückkehr“ liegen alle vor der Rückkehr des Dienstes (die Haltepunkte `vor` und `nach` laufen im Fake); für die Route ist „nach der Rückkehr“ von „im Aufruf“ nicht zu trennen. Ein Ende des `ctx` erst in `Append` oder danach kennt die Menge nicht; eine Mutation, die den Übergang auswertet (`v-h09`: Abbruch, wenn der `ctx` zwischen vor und nach `Append` endet), bleibt grün (1 von 1), jede Mutation, die den Zustand am Ort prüft, ist durch die früheren Enden rot (`v-h10` rot). Kein Bruch der Zusage; die Zeile IM19 sagt „nach der Rückkehr (bei `POST` vor `Append`)“, der Plan nennt die Gleichheit mit „im Aufruf“ nicht | `…slice-admin-imports-audit.md` §3 IM19, M-Kontext | ja — `v-h09`, `v-h10` | Zeitpunkt jenseits der Menge |
| V-6 | INFO | `SPEC-297` Nr. 11 Satz 2 („mit anderem Inhalt gewinnt einer, der andere ist `ODR-E6102` mit dem Eintrag des Dienstes“) hat keine Zeile in IN1; der Teil der Route (kein Eintrag bei `ODR-E6102`, der Sieger hat seinen) tragen IM21 (`d02`) und IM20, die Wirkung hat der Dienst (`SPEC-127` Nr. 18 b). Keine Aktion | `…slice-admin-imports-audit.md` §3 IN1 | nein — Durchsicht | Satz des Vertrags ohne eigene Zeile |
| V-7 | INFO | Abdeckung: `go test -race -coverprofile` (Image `odata-recorder:race`, Wegwerfkopie von `HEAD`) gibt für `admin` 99,6 % (der Plan nennt 99,5 %: der Merge von `main` bringt den Code von `workspaces` ins Paket), null Blöcke mit Zähler 0 in `imports.go`, dieselben drei Blöcke im Altcode (`route.go` 128.4, `tokendatei.go` 108.3 und 132.3); `support` 100 %; `driven/audit` 95,7 % mit den vier Blöcken im Altcode (`audit.go` 95.3, 105.3, 114.3, 121.3). Der Plan stimmt bis auf die Dezimalstelle | `…slice-admin-imports-audit.md` §3 *Zweige als Menge* | ja — Abschnitt 6 | Zahl des Plans und Messung weichen um 0,1 ab |
| V-8 | INFO | `make gates` lief in rund 45 Sekunden (Stempel `gates-passed.head` gleich `21bc3f29`): die Stufen `build`, `test`, `race`, `lint`, `handbuch`, `katalog`, `groessenmarke` kamen aus dem Schicht-Cache, `test-contract`, `test-determinism` (100 Läufe, 0 abweichend) und `test-integration` (Stufen 1 bis 10 grün) liefen live. Das ist kein Beleg für `test` und `race`; ich habe beide frisch gefahren (Abschnitt 5) | `harness/README.md` §Sensors | ja — Abschnitt 5 | Gate aus dem Cache |

Die bekannten Flakes (`tools/pruefung`, `TestWettlaufNebenlaeufig`) sind nicht aufgetreten.

## 1. Definition of Done, Punkt für Punkt

| DoD-Punkt (Plan §2) | Stand | Beleg |
|---|---|---|
| Audit: genau ein Eintrag `import_veroeffentlicht` nach `Import`, vor der Rückkehr des Handlers, Felder aus dem Dienst und Korrelations-ID der Antwort; kein Eintrag bei Fehler, Verstoß, Kontextfehler, ungültigem Report und `GET`; Fehler von `Append` ändert nichts; achte Zeile von `Arten()`; Zählstellen 7 → 8 | erfüllt | `imports.go` Zeilen 63 bis 70 (`reportAntwort` vor `Append`, `_ =`, `return antwort, nil`); Mutationen `v-k01` bis `v-k16`, `v-s03`, `v-z01` bis `v-z04` (Abschnitt 3); Zählstellen: `grep` über `src/`, `tests/`: vier Dateien tragen die Zahl (`support/audit_test.go` 8, `driven/audit/audit_test.go` 8 Arten und 21 Paare, `log_sperre_test.go` 3·8 und 8, `bootstrap/audit_betrieb_test.go` 8), keine weitere Stelle (`workspaces_quelltext_test.go` nennt die Art als Verbot für die Dateien von `workspaces`); `v-z01` (neunte Art) macht in allen vier Paketen einen Test rot |
| Sendung (e): IM21 liest die Mengen von Teil 1, die vier `POST`-Zellen von IM3 zählen 1, `TestIM15QuelleDerRouten` liest auch die Dateien dieses Teils | erfüllt | `TestIM21KeinEintrag` liest 86 + 512 + 24 + 305 Bodys, 44 ungültige Elemente × 3 × 6, `reportsUngueltig` 7 + 44 × 3 × 4 in 3 Zuständen (Zählprüfungen im Test); `TestIM3AufrufeUndKontext` rot bei `v-w03`, `v-w04` (abgeleiteter `ctx`); der Quelltest ändert sich nicht (`imports_quelltext_test.go` 0 Zeilen), `Append` ohne `ctx` |
| Warten und Abbruch (IM18, IM19) | erfüllt, mit V-1 und V-2 | `v-w01`, `v-w02` (Frist von 1 h um `Import` und um `Get`) rot in `TestIM18Warten`, `TestIM19Kontext`, `TestIM3AufrufeUndKontext`; `v-k01` bis `v-k04`, `v-k06` bis `v-k09`, `v-k13`, `v-k15`, `v-k16` rot in `TestIM19Kontext`; `v-k12` rot in `TestIM19Kontextfehler` |
| Nebenläufigkeit (IN1, IN2), Pflichten (b) und (c) von `SPEC-238` Nr. 15 | erfüllt | Abschnitt 4 (Raten); `v-m1-semaphor-1-um-append` und `…-um-get` rot in `TestIN2KeineSperreUeberDemAufruf` und `TestIN1GleichzeitigeRequests`; `TestAppendEineAufrufstelle` rot bei `v-k06`, `v-k09` (weiterer Aufruf von `Append`); `admin_abgewiesen` und die vier Arten des Dienstes erscheinen in keiner Datei der Routen (`TestIM15QuelleDerRouten` grün) |
| `make gates` grün | von mir belegt | Abschnitt 5 (Log ohne `***`, Stempel, `pruefung test` und `pruefung race` frisch); das Kästchen im Plan ist noch offen |
| Review durchgeführt, Report liegt vor | Report liegt vor; Kästchen offen | `2026-11-23-slice-admin-imports-audit-review` (0 HIGH, 3 MEDIUM, 3 LOW, 2 INFO); die Fix-Runde folgte ohne zweites Review und ist hier gegen die Klassen aufgelöst |
| Doku-Update, Handbuch, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen (Closure) | §7 trägt *Wird bei der Closure gefüllt*; kein Fehlercode, Schlüssel oder Kommando kommt hinzu (`SPEC-297` Nr. 13; `make katalog-check` und `make handbuch-check` im Lauf von `make gates` grün); §6 trägt je Risiko einen Ausgang oder *weiter offen* mit Adresse |

## 2. Zusagen des Vertrags gegen den Code (Plan-vs-Code-Diff)

| Zusage | Stelle im Code | Befund |
|---|---|---|
| `SPEC-297` Nr. 9 a: `Import` einmal, synchron, mit dem `ctx` des Handlers; keine Frist, keine Goroutine, keine Ablösung; Antwort erst nach der Rückkehr | `imports.go` 55 (`h.imports.Import(ctx, anfrage)`), 76 (`Get(ctx, …)`); kein `go`, kein `context.With*` | stimmt; `TestIM18Warten` (4 Zellen), `v-w01` bis `v-w04` rot |
| Nr. 9 c: Kontextfehler bricht ab, ohne Fehlerobjekt, Zählen, Eintrag | die Route reicht den Fehler unverändert zurück (Zeilen 56 bis 58, 77 bis 79), der Rahmen bricht ab (`SPEC-238` Nr. 9) | stimmt; Menge M-Kontextfehler 288 Zellen; Grenze der Menge V-1 |
| Nr. 9 d: Erfolg bei beendetem `ctx` ist Antwort und Eintrag | `Append` ohne `ctx`, keine Prüfung des Kontexts | stimmt; M-Kontext 66 Zellen, 11 eigene Mutationen rot |
| Nr. 10 a, b, d: genau ein `Append` nach dem gültigen Report, Felder aus `e.Import`, Korrelations-ID der Antwort, vor der Rückkehr | Zeilen 63 bis 70 | stimmt; `ServiceID`, `ImportID`, `KorrelationsID` je eine Stelle, übrige Felder nicht gesetzt (`ErrorCode`, `SessionID`, `WorkspaceID` leer: `TestIM20Eintrag` gegen das Literal mit allen sieben Feldern) |
| Nr. 10 c: kein Eintrag bei Fehler, Verstoß, Kontextfehler, ungültigem Report, `GET` | Rückgaben vor dem `Append` (Zeilen 51 bis 58, 64 bis 66); `lies` ruft nie `Append` | stimmt; `v-k09` (Eintrag bei `GET`) rot |
| Nr. 10 e: Ergebnis von `Append` ändert nichts | `_ =` | stimmt; M-Append 56 Fehler, 112 Zellen; `v-a02`, `v-a03`, `v-a05` rot, `v-a01`, `v-a06` grün (Negativ (6)); V-2 für den Fehler bei beendetem `ctx` |
| Nr. 11: keine Sperre über einem Aufruf, kein Zustand außer den Ports | `importHandler` hält `imports`, `audit`; kein `sync` | stimmt; Abschnitt 4 |
| `SPEC-234` Nr. 2: achte Zeile `import_veroeffentlicht`, Pflicht `service_id`, `import_id`, sonst verboten | `audit.go` des Ports: `ArtImportVeroeffentlicht` und `{… FeldServiceID, FeldImportID}` | stimmt; `v-z02`, `v-z03`, `v-z04` rot |

## 3. Die Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

75 eigene Mutationen, je eine Wegwerfkopie, ein Container je Mutation (`go test -count=1 -run '^(TestIM|TestIN|TestAppendEineAufrufstelle)'` über `admin`, bei `v-z*` die Pakete `support` und `driven/audit`, bei `v-z01` auch `bootstrap`; die Raten in Abschnitt 4). Rot für den benannten Grund zählt nur, wenn ein Test der Verhaltens-Gruppe selbst rot wird (die Listen der roten Tests stehen in `sweep*.log`); Mutationen mit `ctx.Err()` in `imports.go` sind zusätzlich durch `TestIM15QuelleDerRouten` rot, das zählt hier nicht. Summe: 43 rot für den Grund, 9 nicht gezählt (`v-k10`, `v-k11`, `v-k14`, `v-a07` allein durch den Quelltest, `v-f01` bis `v-f05` durch eine Panik oder die Zählstelle an anderer Stelle), 18 grün, 5 Fenster ohne Fuge (Abschnitt 4).

### 3.1 M-Kontext (66 Zellen) und der Rest von IM19

| Mutation (an Elementen, die Review und Implementer nicht nannten) | Ergebnis | Roter Test |
|---|---|---|
| `v-k01` Abbruch nach dem Erfolg bei `cancel` vor dem Aufruf und `Neu` falsch | rot | `TestIM19Kontext` |
| `v-k02` `GET`: Abbruch nach dem Erfolg bei Frist vor dem Aufruf | rot | `TestIM19Kontext` |
| `v-k03` Eintrag ohne `ServiceID` bei Frist und `Neu` falsch | rot | `TestIM19Kontext` |
| `v-k04` Eintrag ohne Korrelations-ID bei `cancel` und `Neu` wahr | rot | `TestIM19Kontext` |
| `v-k06` zweiter Eintrag bei `cancel` | rot | `TestIM19Kontext`, `TestAppendEineAufrufstelle` |
| `v-k07` `GET`-Body leer bei beendetem `ctx` | rot | `TestIM19Kontext` |
| `v-k08` Status 202 bei Frist und `Neu` falsch | rot | `TestIM19Kontext` |
| `v-k09` Eintrag bei `GET` allein bei `cancel` | rot | `TestIM19Kontext`, `TestAppendEineAufrufstelle` |
| `v-k13` Eintrag nur bei aktivem `ctx` und `Neu` wahr oder bei beendetem und `Neu` falsch | rot | 14 Tests (u. a. `TestIM19Kontext`, `TestIM20Eintrag`) |
| `v-k15` Eintrag nur bei aktivem `ctx` oder `Neu` wahr | rot | `TestIM19Kontext` |
| `v-k16` `GET`: Status 202 bei `cancel` | rot | `TestIM19Kontext` |
| `v-k10`, `v-k11` Fehler des Dienstes bei beendetem `ctx` wird Kontextfehler (`POST`, `GET`) | rot allein `TestIM15QuelleDerRouten` | V-2 |
| `v-h01`, `v-h02` dasselbe mit Hilfsfunktion in einer anderen Datei | **grün (2 von 2)** | V-2 |
| `v-k14` Abbruch, wenn der `ctx` erst in `Append` endet | rot allein `TestIM15QuelleDerRouten` | V-5 |
| `v-h09` dasselbe mit Hilfsfunktion | **grün** | V-5 |
| `v-h10` Eintrag nur bei aktivem `ctx` (Hilfsfunktion; Kontrolle) | rot | `TestIM19Kontext` |
| `v-w01`, `v-w02` Frist von 1 h um `Import` bzw. `Get` (Hilfsfunktion) | rot | `TestIM18Warten`, `TestIM19Kontext`, `TestIM3AufrufeUndKontext` |
| `v-w03`, `v-w04` abgeleiteter `ctx` mit Wert um `Import` bzw. `Get` | rot | `TestIM3AufrufeUndKontext` (Teil 1: Identität des `ctx`) |

Die Menge M-Kontext (Route 2 × Art des Endes 2 × Zeitpunkt 3 × Ausgang 6 bzw. 5 = 66) deckt die Achsen, die der Wortlaut von `SPEC-297` Nr. 9 d und 10 a nennt (`Neu`, Art, Zeitpunkt, Route): 11 von 11 Mutationen an anderen Elementen der Menge sind rot. Offen sind die Elemente jenseits der Menge (V-2, V-5).

### 3.2 M-Kontextfehler (24 Formen, 288 Zellen)

| Mutation | Ergebnis |
|---|---|
| `v-k12` bei Frist wird ein Fehler anderer Art Kontextfehler | rot (`TestIM19Kontextfehler`) |
| `v-h03` (`POST`) und `v-h04` (`GET`): Kontextfehler hinter drei `%w` wird 202 | **grün (2 von 2)** |
| `v-h05` Kontextfehler in der Mitte eines `Join` aus drei Fehlern wird 202 | **grün** |
| `v-h06` `Join` aus `Canceled` und `DeadlineExceeded` wird 202 | **grün** |
| `v-h07` Ursache in zweiter Verpackung wird 202 (Kontrolle an einem Element der Menge) | rot (`TestIM19Kontextfehler`) |
| `v-f01` bis `v-f05` (in `imports.go` mit `errors.Is` und `errors.Unwrap` ohne Schutz) | rot, aber allein durch eine Panik in `TestIM13Unveraendert` an einem nil-Zeiger `*inbound.Fehler` oder durch die Zählstelle: **nicht gezählt** (der benannte Grund ist nicht der Grund); ersetzt durch `v-h03` bis `v-h07` mit Schutz vor der Panik |

Die Menge deckt, was sie nennt (Art 2 × Verpackung 12), und die Kontrolle `v-h07` zeigt, dass der Fänger an ihr trennt. Die Achsen Tiefe und Breite sind unbeschränkt: **V-1**.

### 3.3 M-Kennung (1869 Zellen) und M-Append (112 Zellen)

| Mutation | Ergebnis |
|---|---|
| `v-n01` `ID` gekürzt auf 5000, `v-n02` `ServiceID` auf 131072, `v-n09` `ID` auf 1048576 | rot (`TestIM22KennungenUnveraendert`) |
| `v-n03` `ID` kleingeschrieben, `v-n04` `.` zwischen Alphanumerischen entfernt, `v-n10` `ServiceID` ab 8 Zeichen großgeschrieben | rot (`TestIM22KennungenUnveraendert`, `v-n10` auch `TestIM20Eintrag`, `TestIM19Kontext`) |
| `v-n05` `-` zwischen zwei Ziffern entfernt, `v-n06` `NUL` nur ab 65 Bytes entfernt, `v-n07` Länge genau 4000 gekürzt, `v-n08` gekürzt auf 1048600 | **grün (4 von 4)**: jenseits der Zweierfolgen und der Längen der Menge (Negativ (5), V-4) |
| `v-a02` `errors.Is(…, EPIPE)`, `v-a03` Text per `Error()`, `v-a05` `errors.As` auf `*os.PathError` | rot (`TestIM22AppendFehler`; der Fehler für jedes Ziel und der nil-Zeiger tragen) |
| `v-a01` `== syscall.EPIPE`, `v-a06` Typ `syscall.Errno` | **grün (2 von 2)**: Fehler außerhalb der Menge (Negativ (6)) |
| `v-a07`, `v-h08` Fehler von `Append` bei beendetem `ctx` wird 500 | rot allein `TestIM15QuelleDerRouten` bzw. **grün** (V-2) |

### 3.4 Prüfgrenzen (3) und (8), Zustand, Art

| Mutation | Ergebnis |
|---|---|
| `v-m31`, `v-m32`, `v-m40` Semaphor 31, 32, 40 um `Import` | 31 rot (`TestIN1GleichzeitigeRequests`, nach 10 s), 32 und 40 grün: wie (3) |
| `v-t65536`, `v-t1048577`, `v-t1048700`, `v-t3145728` Eintrag bei ungültigem Report ab der Länge | die ersten zwei rot (`TestIM12Stelle`), die letzten zwei grün: wie (8) |
| `v-m1-semaphor-1-um-append`, `…-um-get` | rot (`TestIN2KeineSperreUeberDemAufruf`, `TestIN1GleichzeitigeRequests`) |
| `v-s01` Zwischenspeicher je `id` (`GET`), `v-s02` Zustand je `import_id` (`POST`), `v-s03` Eintrag nur einmal je `import_id` | rot (8 bis 14 Tests, u. a. `TestIM17FolgeOhneZustand`, `TestIM10Status`, `TestIM20Eintrag`) |
| `v-z01` neunte Art in `Arten()` | rot in `support` (`TestArtenGleichTabelle`, `TestArtenLiefertKopie`), `driven/audit` (`TestKennungsform`, `TestLogBeiFehler`) und `bootstrap` (`TestAuditAufrufstellen`; der Test trägt `n != 8` und die Schleife über `Arten()`, die Zeile ist nicht einzeln gemutet) |
| `v-z02`, `v-z03`, `v-z04` Pflicht `session_id` statt `import_id`, Leerzeichen im Text, `error_code` als drittes Pflichtfeld | rot (`TestArtenGleichTabelle`, `TestFeldBelegung`, `TestZeileImportVeroeffentlicht`, `TestArtTexte`) |

## 4. IN1: Raten je Prozess und Fenster (Klasse `nebenläufig`)

Messender das Skript `rate.sh` (Scratchpad) mit dem Test-Binary der Wegwerfkopie: `go test -c`, dann *N* Prozesse nacheinander mit `-test.run '^TestIN1GleichzeitigeRequests$' -test.count=1`; Arm ohne Detektor mit `CGO_ENABLED=0`, Arm mit Detektor `go test -race -c`; Container `--cpus=4`, `GOMAXPROCS` gleich 4 oder 1; Last des Hosts 3,7 bis 12,0 (Host mit 20 Prozessoren). Die Zahl steht als rot / Prozesse. Kontrolle ohne Mutation: 0 von 10 bei 4 und bei 1 Prozessor.

| Mutation | Fenster | 4 Prozessoren | 1 Prozessor |
|---|---|---|---|
| `v-gA1` `atomic.Value` geschrieben vor dem Lesen des Bodys, gelesen für den Eintrag | Lesen | 20/20 | 20/20 |
| `v-gA2` `atomic.Value` um `Import` | Aufruf | 20/20 | 20/20 |
| `v-gA4` `atomic.Value` um `Append`, nach der Rückkehr verglichen | `Append` | 20/20 | 20/20 |
| `v-gM5` `sync.Map` um `Get` | `Get` | 20/20 | 20/20 |
| `v-gA3` `atomic.Value`, geschrieben und sofort gelesen (nach `Import`, vor `Append`) | ohne Fuge | 30/30 | **0/30** |
| `v-gM3` `sync.Map`, ebenso | ohne Fuge | 30/30 | **0/30** |
| `v-gP3` Sperre, ebenso | ohne Fuge | 30/30 | **0/30** |
| `v-gA6` `atomic.Value`, `GET` ohne Fuge | ohne Fuge | 26/30 | **0/30** |
| `v-gR3` Zelle ohne Sperre (Datenrennen), ebenso | ohne Fuge | ohne Detektor 0/20, mit Detektor 20/20 | ohne Detektor 0/20, mit Detektor 20/20 |

Lesart: die vier Zellen mit Fuge (der Treffpunkt hält alle 32 Requests) sind in jedem Lauf rot, auch bei einem Prozessor; die Fenster ohne Fuge sind bei einem Prozessor nicht gefangen (5 von 5 weitere Formen ungefangen, zusammen mit den 6 von 6 des Plans) und bei vier Prozessoren und der Last 4 bis 8 in 26 bis 30 von 30 Läufen rot; das Datenrennen ohne Fuge fängt allein `make test-race` (20 von 20). **Negativ (4) ist ehrlich**: es nennt dieselben Maße und die Abhängigkeit von Prozessorzahl und Last. Die Raten gelten als Größenordnung (`AGENTS.md` §3.10). Die Raten von IN2, IM18 und IM20 (`a01`, `a05`, `c08`, `h01` bis `h05` des Plans) habe ich nicht neu gemessen; die zwei Semaphore (`v-m1`) sind mit einem Prozess je Mutation rot.

## 5. Gates und Messung

| Prüfung | Ergebnis |
|---|---|
| `make gates` im Worktree | Lauf unter `flock`, Last 7,6 und 10,7 GB frei beim Start; Log ohne `***`; Stempel `.harness/state/gates-passed.head` gleich `21bc3f29`; `a-check` 0 Befunde, `baseline-verify` OK, `migration-check` ok, `test-contract` grün, `test-determinism` 100 Läufe 0 abweichend, `test-integration` Stufen 1 bis 10 grün, `zeichen-check` ok (2271 Dateien); Stufen aus dem Schicht-Cache (V-8) |
| `pruefung test` frisch | `docker build --no-cache-filter pruefung,test --target test`: Exit 0, `pruefung test: ok` nach 362,2 s; Modul `tools/pruefung` 33,4 s ok; Last 10,3 bis 12,6 |
| `pruefung race` frisch | `docker build --no-cache-filter race --build-arg RACE_IMAGE=… --target race`: Exit 0, `pruefung race: ok` nach 401,8 s |
| `make groessenmarke` | `ok, Messung 1075 von Marke 1100`; Messpunkt `zeilen=9 nebenlaeufig=2 fakes=0 produktcode=25 schaetzung=755 marke=1100 messung=1075` |
| `git diff --numstat b199f9af..HEAD` über die 14 Pfade des Blocks | 1075 (13 Dateien mit Zeilen, `imports_quelltext_test.go` 0); die Zeilen stimmen mit dem Messpunkt des Plans |
| fremde Zeilen | `git diff --numstat b199f9af..0feeb705` über die 14 Pfade: leer (0) |
| `git diff --stat 0feeb705 -- src/bootstrap` (`main`-Elter des Merge, IM23) | allein `src/bootstrap/audit_betrieb_test.go`, 9 Einfügungen, 9 Löschungen; gegen `b199f9af` 7 Dateien (V-3 c) |
| Produktcode von Hand | `imports.go` 12 Einfügungen und 3 Löschungen, `audit.go` 5 und 2: 17 und 5 wie im Plan |
| `make docs-check` | 1161 Dateien geprüft, 0 Befunde (mit diesem Report in der Fassung vor der letzten Berichtigung der Zahlen; nach dem Schreiben der endgültigen Fassung erneut gefahren) |

## 6. Abdeckung (Nullblöcke)

`go test -race -count=1 -coverprofile` je Paket (Image `odata-recorder:race`, Wegwerfkopie von `HEAD`, Last 12,6 bei Beginn und 9,8 am Ende), Blöcke mit Zähler 0: `admin` 99,6 %: `route.go` 128.4 (`zerlegePfad`), `tokendatei.go` 108.3 (`pruefeUndLies`) und 132.3 (`erzeuge`), **keiner in `imports.go`**; `support` 100 %, keiner; `driven/audit` 95,7 %: `audit.go` 95.3, 105.3 (`schreibe`), 114.3 und 121.3 (`endetOhneZeilenende`), alle im Altcode (V-7).

## 7. Nicht nachgemessen

- `make gegenproben` (nicht zu fahren; Stempel nicht berührt); `make test-integration` und `make test-determinism` nur als Teil von `make gates`, kein eigener Lauf.
- Die Raten von IN2 (`h01` bis `h05`), IM18 (`a01`, `a05`) und IM20 (`c08`) des Plans und die Mutationen `g01` bis `g14` des Implementers; ich habe eigene Mutationen an anderen Elementen gefahren (Abschnitt 4).
- Die Raten bei Last über 15 und mit mehr als 30 Prozessen je Arm (die Messung lag bei Last 3,7 bis 12,0; die Raten des Plans bei Last 11 bis 17 sind nicht wiederholt).
- Die Tests des Rahmens zur Erkennung eines Kontextfehlers (`SPEC-238` Nr. 9, Teil von `slice-admin-fehlerabbildung`): ob sie Tiefe 3 und `Join` aus drei Fehlern führen, habe ich nicht geprüft (V-1).
- Die Zeile `n != 8` in `audit_betrieb_test.go` einzeln: `v-z01` macht `TestAuditAufrufstellen` rot, der Test trägt zwei Prüfungen; die Zeile allein ist nicht gemutet.
- `TestIN1GleichzeitigeRequests` (F-6, Zählprüfung der Haltepunkte aus den Typen): gelesen (`haltepunkte(reflect.Type)` liest Felder und Arrays vom Typ `func`), keine eigene Mutation an einem weiteren Haltepunkt.
- Die Zusagen über das Trennen der Verbindung am Socket (`SPEC-297` *Ungeprüft*) und ein Test-Server mit echtem Client.

## Auflagen an den Planner

- **A-1 (V-1, vor dem Move nach `done/`):** Entweder ein akzeptiertes Negativ zu M-Kontextfehler (Prüfgrenze: Tiefe der `%w`-Kette bis 2, Breite eines `Join` bis 2, ein Kontextfehler je Wert; Grund: die Verpackung ist unbeschränkt, die Route prüft den Kontextfehler nicht selbst, `SPEC-238` Nr. 9 trägt der Rahmen; Folge im Betrieb: eine Route, die allein bei Tiefe 3, in der Mitte eines `Join` aus drei Fehlern oder bei beiden Kontextfehlern zugleich anders handelt, bleibt grün; Maß: 4 von 4 grün, die Kontrolle an der Menge rot) oder die Menge um Tiefe 3 und die Mitte eines `Join` aus drei Fehlern erweitern (zwei Achsen, je eine Schleife, `AGENTS.md` §3.17).
- **A-2 (V-2):** Die Zusatztabelle in §6 um die Paare (Fehler des Dienstes mit Code × beendeter `ctx`) und (Fehler von `Append` × beendeter `ctx`) ergänzen und je eine Zelle als Fänger führen (M-Kontext um den Ausgang „Fehler mit Code 404“, M-Append um ein Ende des `ctx`) oder als Negativ mit Grund (der Quelltest verbietet die Namen in `imports*.go`) und Maß (3 von 3 grün ohne den Namen) führen.
- **A-3 (V-3):** Nachzug von Zeile IM21 („die 35 von IM19“: die Menge von 66 zählt `TestIM19Kontext` mit dem Orakel von IM19), von §6 *Fremde Zeilen* (Stand `0feeb705`) und der benannten Ausgabe von IM23 nach dem Merge (gegen den `main`-Elter, `git diff --stat 0feeb705 -- src/bootstrap`); den Kopf von `TestIM16BinaryUnveraendert` bei der Closure von Teil 2 anpassen (Folge-Commit des Implementers, kein Teil dieses Reports).
- **A-4:** Die Abweichungen (a) und (b) (66 statt 35 Zellen, 32 Requests statt 16 Goroutinen) bestätigt der Planner bei der Closure (die Adresse steht in §6); F-7 (dritte Lage von IM24) bleibt beim Architect. Die Kästchen *`make gates` grün* und *Review durchgeführt* hakt der Planner bei der Closure ab; die Evidenz liefert dieser Report.

**Gegengelesen:** Dieser Report ist nach dem Schreiben mit `LC_ALL=C grep -P` auf Folgen aus Backslash, `u` und vier Hex-Ziffern, auf NBSP, BOM und Zeilentrenner geprüft (`AGENTS.md` §3.18); `make docs-check` und `make zeichen-check` sind nach dem Schreiben gelaufen.
