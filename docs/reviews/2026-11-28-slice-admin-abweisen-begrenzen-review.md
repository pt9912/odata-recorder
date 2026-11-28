# Review-Report: slice-admin-abweisen-begrenzen (höchstens 20 Einträge `admin_abgewiesen` je Fenster von 60 s) — 2026-11-28

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf.

**Gegenstand:** `git diff fd960822..d5a86f0a` über die Pfade des Blocks und das Handbuch; der eigene Commit ist `f1269fd1`. Die Änderungen aus `main` sind nicht Gegenstand. Der Merge `ec48ad3e` trägt nach `git show --remerge-diff` keine eigene Änderung, `d5a86f0a` allein die Auflösung eines Konflikts im Slice-Plan.

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`d5a86f0a`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-28

**Eingangs-Kontext:** Plan `slice-admin-abweisen-begrenzen` (§1, §3 mit W01 bis W18, Fenster der Sperre, Mutationen und Raten, §4, §6); `SPEC-301`; `SPEC-238` Nr. 8 und 12; `SPEC-300` Nr. 5 und 12; `SPEC-041` Nr. 7; `SPEC-114`; `LH-FA-13` samt `LH-FA-13.a`, `LH-QA-04`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `AGENTS.md` §3.

**Urteil: nicht freigegeben bis F-1 bis F-3.** Kein HIGH. Der Produktcode (`abweisen_fenster.go`, die Zeile in `schreibeFehler`, die Zeile im Root) hält in diesem Lauf `SPEC-301`; die Befunde betreffen Fänger und Plan.

**Vorgehen.** Wegwerfkopie (`git archive HEAD`) per Tar-Strom in einen Container aus dem Image `odata-recorder:race` (`--network none --memory=4g --memory-swap=4g --cpus=4`, kein Mount), unter `flock /tmp/odata-recorder-heavy.lock`. Zwei Läufe nacheinander mit je einem Container, nie zwei zugleich; beide entfernt. Beim Start 14 GiB verfügbar, Last 6,1 und 4,4; während der Messungen Last 2,7 bis 4,0. Messend: Reviewer.

- Grundlauf grün: Paket `admin`; in `bootstrap` die Tests `Admin|M6`; `TestFensterGleichzeitig` mit Detektor in 0 von 10 Prozessen rot (1200 Fenster).
- Das ganze Paket `bootstrap` lief in diesem Review nicht zu Ende (Zeitgrenze 300 s des Skripts); kein Ergebnis daraus.
- Zähler an der Stelle der Drosselung: im Paket `admin` drosseln außer den Tests des Fensters allein `TestSchemaJedesByte`, `TestKandidatAlleBytes` (beide prüfen keine Einträge) und die Kontrolle von `TestSicherheitKeineAenderung`; in `bootstrap` unter `Admin|M6` allein `TestAdminAbweisenBegrenzt`. Kein Bestandstest, der Einträge prüft, ist still verengt.
- Nicht gefahren: `make gates`, `make gegenproben`, `make test-race`, `make handbuch-check`, `make test-integration`, `tests/e2e`.

**Eigene Mutationen** an `abweisen_fenster.go`:

| Nr. | Mutation | Ergebnis |
|---|---|---|
| X1 | `beginn` wird bei jeder Zulassung auf `jetzt` gesetzt | **grün:** Paket `admin` ohne und mit Detektor; eine Probe (1 Abweisung, nach 50 s 19, nach weiteren 10 s eine) ist am Original zugelassen, an X1 gedrosselt |
| X2 | zweiter Aufruf von `uhr.Now()` je Abweisung | **grün:** Paket `admin` ohne und mit Detektor; die Probe zählt 42 Aufrufe bei 21 Abweisungen |
| X3 | `beginn` aus einer zweiten Lesung der Uhr statt aus `jetzt` | **grün:** wie X2 |
| X4 | Sperre zwischen `uhr.Now()` und dem Vergleich freigegeben und neu genommen | **grün:** Paket `admin` ohne Detektor; `Fenster` und `Gedrosselt` ohne und mit Detektor |
| X5 | Vergleich und Öffnen in zwei Abschnitten | rot: `TestFensterGleichzeitig`, ohne und mit Detektor |
| X6 | eine gedrosselte Abweisung verschiebt `beginn` | rot: `TestFensterGrenze`, `TestFensterBeginn` |
| X7 | Öffnen und Prüfen in zwei Abschnitten (§6 Negativ 6) | grün, wie im Plan benannt |
| X8 | Prüfen und Erhöhen in zwei Abschnitten, nur `TestFensterGleichzeitig` | `--cpus=4`: rot 20 von 20 ohne und 20 von 20 mit Detektor; `GOMAXPROCS=1`: rot 0 von 10 ohne und 0 von 5 mit Detektor |

## Findings

Befunde ohne Lösungsvorschlag nach dem Schema des Skills.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Alle 20 Zulassungen eines Fensters liegen in jedem Test auf demselben Stand der Uhr (oder jede in einem eigenen Fenster); ein Beginn, der mit jeder Zulassung mitwandert, bleibt grün (X1), obwohl W04 den Beginn der eröffnenden Abweisung zusagt. Versagen: eine solche Änderung drosselt 60 s nach der eröffnenden Abweisung weiter und besteht die Suite. | `SPEC-301` Nr. 2 b, 5 c; `AGENTS.md` §3.15 | `src/adapters/driving/admin/abweisen_fenster_test.go` · `func pruefeFensterVoll` | ja, `make test` bliebe mit X1 grün | Fixture verdeckt, was der Test prüft (Zeitpunkte der Zulassungen gleich) |
| F-2 | MEDIUM | Die Zusage `Uhr.Now()` genau einmal je Zulassung hat weder eine Zeile in der Tabelle noch einen Fänger: gezählt wird die Uhr allein bei `New` und den Antworten ohne 401, jeweils auf 0 (X2, X3). Versagen: der Beginn oder der Vergleich stammt aus einer anderen Lesung als zugesagt, und die Suite bleibt grün. | `SPEC-301` Nr. 2 a; `AGENTS.md` §3.15 | `docs/plan/planning/` Plan §3 · `W09` / `abweisen_fenster.go` · `jetzt := uhr.Now()` | ja, `make test` bliebe mit X2 grün | Zusage ohne Fänger |
| F-3 | MEDIUM | Die Zeile *Uhr vor der Sperre (Lesen – Öffnen)* nennt allein `Now()` vor `Lock`; ein Freigeben der Sperre nach `Now()` und vor dem Vergleich bleibt in jedem Lauf grün (X4) und steht auch nicht als Negativ. Das Maß von §6 Negativ 6 (1 von 4 Fenstern ohne Fänger) trifft damit nicht zu. Die Wirkung gleicht der von Negativ 6: die Zahl je Fenster bleibt höchstens 20, eine Abweisung zählt in ein Fenster, das nach ihrer Zeit beginnt. | `AGENTS.md` §3.10, §3.15; `SPEC-301` Nr. 2 | Plan §3 · `Uhr vor der Sperre (Lesen – Öffnen)` | ja, `make test-race` bliebe mit X4 grün | Fenster der Sperre ohne Fänger und ohne Negativ |
| F-4 | LOW | §6 verweist auf eine *Mutation D7a in §3*; §3 führt keine Mutation dieser Bezeichnung (die Zeile heißt dort *Nachzug von `TestJedeAbweisungEinEintrag`*). | `AGENTS.md` §3.9 | Plan §6 · `Mutation D7a in §3` | nein | Verweis ins Leere |
| F-5 | INFO | Nach einem Rücksprung der Systemzeit hält die Uhr des Prozesses den Höchstwert (`SPEC-114` Nr. 2): das offene Fenster lässt höchstens 20 Einträge zu und bleibt bis zum Einholen oder Neustart offen, bei einer zuvor weit vorgestellten Uhr also beliebig lange ohne Eintrag einer Abweisung; allein der Zähler steigt. `SPEC-301` Nr. 5 c entscheidet das so, das Handbuch nennt es nicht. Zuständig: Architect. | `SPEC-301` Nr. 5 c | `src/adapters/driven/clock/clock.go` · `return u.hoechst` | nein | Folge im Betrieb nicht im Handbuch |
| F-6 | INFO | Der Wirkungs-Fänger für Prüfen und Erhöhen in zwei Abschnitten wirkt nur mit mehreren Prozessoren (X8: 0 von 15 Prozessen rot bei `GOMAXPROCS=1`); die Raten des Plans gelten bei `--cpus=4`, wie dort genannt. Ein Gate-Host mit einem Prozessor sähe die Mutation nicht. | `AGENTS.md` §3.10 | `abweisen_fenster_gleichzeitig_test.go` · `const goroutinen, fenster = 64, 1200` | ja, Lauf mit `GOMAXPROCS=1` | Rate hängt an der Zahl der Prozessoren |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driving/admin/abweisen_fenster.go`, `admin.go` (Sperre, Atomarität, Reihenfolge, Verklemmung) | geprüft, ohne Befund: ein Abschnitt unter der Sperre, frei vor `Append`; unter ihr allein `Uhr.Now()`, keine Gegenrichtung der Sperren |
| Ausnutzung durch einen Client ohne Token | geprüft, ohne Befund über `SPEC-301` Nr. 9 hinaus: andere Arten des Audit unberührt, gedrosselte schreiben weder Eintrag noch Log, eine gedrosselte verschiebt das Fenster nicht (X6) |
| Abweisung selbst (Status, Body, Header, Zähler, beide Wege zu 401) | geprüft, ohne Befund: allein `Append` hängt am Fenster |
| Bestandstests (`TestJedeAbweisungEinEintrag`, `konfig_test.go`, `abweisen_geheimnisse_test.go`, `admin_start_test.go`, Netzwerkzugriffstest) | geprüft, ohne Befund: Erwartungen je zugelassener Abweisung unverändert, Mengen erweitert; die Grenze 1100 und 1200 gilt über alle Fenster, wie der Kommentar sagt |
| `sessions_quelle_test.go` (`rahmenDateien`) | geprüft, ohne Befund: ausgenommen ist allein die neue Datei; Goroutine, Timer und Systemuhr dort hält `TestFensterOhneZeitgeber` |
| Gates (`.golangci.yml`, `Makefile`, `harness/mk/`, `Dockerfile`, `.a-check.yml`, `.d-check.yml`, `tools/`, `//nolint`) | geprüft, ohne Befund: keine Änderung im eigenen Commit |
| `src/bootstrap/` (Verdrahtung der einen Uhr, W17) | geprüft, ohne Befund |
| `docs/user/benutzerhandbuch.md` (Grenze, Beginn, Zähler, Token, 256 Bit; weitere Stellen zum Audit) | geprüft, ohne Befund außer F-5: die Sätze treffen den Code; kein zweiter Satz behauptet je Abweisung eine Zeile |
| Raster *Operation × Argumentart* und Zusatztabelle in Plan §6; akzeptierte Negative 1, 2, 4, 5, 7 | geprüft, ohne Befund: je Zelle ein Ort, je Negativ Grund, Folge und Maß |
| `tests/integration`, `tests/e2e` (Zählung je Abweisung) | nur gelesen: die Mengen je Prozess stehen im Plan unter 20; nicht gefahren |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 1 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fixture verdeckt, was der Test prüft · Zusage ohne Fänger · Fenster der Sperre ohne Fänger und ohne Negativ · Verweis ins Leere · Folge im Betrieb nicht im Handbuch · Rate hängt an der Zahl der Prozessoren

## Verdikt

**Merge-blockierend:** ja — F-1 bis F-3 (MEDIUM).

**Übergabe:** Findings an den Implementer, F-5 an den Architect; die Finding-Klassen in die Slice-Closure §7. Die DoD-Konformität prüft der Verifier getrennt.
