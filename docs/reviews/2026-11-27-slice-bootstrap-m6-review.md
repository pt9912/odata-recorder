# Review-Report: slice-bootstrap-m6 (Teil D: M6-Test über Admin- und OData-Listener, Zellen Modus × Art, Stufe 11 `TestE2EAdmin`) — 2026-11-27

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf, gesammelt.

**Gegenstand:** `git diff 5dd673ca..702308a2` über die Pfade des Blocks (`tests/integration/m6_test.go`, `m6_hilfen_test.go`, `referenzdienst_test.go`, `tests/e2e/stufe_admin_test.go`, `tools/harness/test-integration.sh`, `src/bootstrap/audit_betrieb_test.go`, `audit_betrieb_hilfen_test.go`, `harness/README.md`) und die Dokumente des Slice (`SPEC-037`, `docs/plan/planning/e2e-abdeckung.md`), Zweig `slice-bootstrap-m6`. Die Commits aus `main` im selben Bereich sind nicht Gegenstand. `main` (`6c1526bc`) ist nur gelesen (`SPEC-300` Nr. 12 a und 14).

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`702308a2`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-27

**Eingangs-Kontext:** Plan `slice-bootstrap-m6` (§1, §3 Tabelle D01 bis D07, Kriterien-Tabelle, Messung der Mutationen, benannte Negative, §4, §6); `SPEC-300` (Nr. 12 bis 14, *Cursor-Test*, *Ungeprüft*), `SPEC-005`, `SPEC-110`, `SPEC-139`; `SPEC-037` Nr. 2, 7, 9, 10; `LH-FA-02`, `LH-FA-09`, `LH-FA-10`, `LH-FA-13`, `LH-FA-14`, `LH-FA-18`, `LH-QA-04`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `AGENTS.md` (Hard Rules). Die Mutationen des Implementers sind nicht übernommen und nicht wiederholt; die Zahlen unten sind eigene Läufe.

**Urteil: nicht freigegeben bis F-1.** Kein HIGH, ein MEDIUM, drei LOW, drei INFO. Am Produkt ist in diesem Lauf kein Fehler gefunden, und die neuen Tests sind in allen Wiederholungen mit frischem `/voll` grün. F-1 ist eine Aussage des Plans über die Deckung eines Kriteriums, kein Befund am Code.

**Vorgehen.** Wegwerfkopie (`git archive HEAD` von `go.mod`, `go.sum`, `src`, `tests`), per `docker cp` in einen Container aus dem Image `odata-recorder:race` (`--network none --memory=4g --memory-swap=4g --cpus=4`, `tmpfs` `/voll` von 1 MiB, kein Mount), die Läufe unter `flock /tmp/odata-recorder-heavy.lock`; Last des Hosts 2,8 bis 9,6; Messender Reviewer. Die Stufe lief am selbst gebauten Binary als Test-Binary, nicht über den Treiber und ohne `--read-only`.

- Basis grün: `TestE2EAdmin` (1,1 s gegen die Frist von 120 s), `TestM6Rundlauf`, `TestM6AbweisungOhneAenderung`, `TestM6Cursor`, `TestAuditAdminZellen`, `TestAuditAufrufstellen`; das ganze Paket `tests/integration` (20 s, die neue Route `POST` des Referenzdienstes bricht keinen Bestandstest).
- Wiederholung: `TestE2EAdmin` 40 von 40 Prozessen grün (`/voll` je geleert); die drei M6-Tests `-count=30` grün und `-race -count=3` grün; `TestAuditAdminZellen` `-count=20` und `-race -count=3` grün. Unter dem Detektor dauern die M6-Tests zusammen rund 7 s. Außer F-7 kein Flackern gesehen.
- Nicht gefahren: `make gates`, `make test-integration`, `make gegenproben`, das ganze Paket `src/bootstrap`.

**Eigene Mutationen** (je ein Lauf, deterministisch):

| Nr. | Mutation | Ergebnis |
|---|---|---|
| X1 | die erste Hälfte des Tokens erst beim Herunterfahren auf Deskriptor 2 | rot: Stufe 11, *nach dem Herunterfahren: das Token steht in stderr* |
| X2 | das Token auf Deskriptor 2 nur bei `DELETE` mit gültigem Token | **grün:** Stufe 11 |
| X3 | `DELETE` ohne Authentisierung | rot: `TestM6AbweisungOhneAenderung` (Status und Ablage); Stufe 11 grün |
| X4 | gültiges Token, Methode ohne Route: vor der 405 läuft der Handler von `reset` | **grün:** die M6-Tests und Stufe 11; rot im Paket `admin` (`TestWS2Zuordnung`, am Fake) |
| X5 | `workspace_angelegt` trägt zusätzlich `SessionID` | rot: `TestAuditAdminZellen`, 4 Zellen |

## Findings

Befunde ohne Lösungsvorschlag nach dem Schema des Skills.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Tabelle *Kriterien · Träger* nennt für `LH-FA-13` Negative („unberechtigte **oder fehlerhafte** Management-Anfrage … ohne Recorder-Zustand zu verändern“) D02 und D04 und sonst nichts. Beide führen allein die unberechtigte Hälfte; mit gültigem Token ist die einzige fehlerhafte Anfrage der Import im Modus `mock` (404). X4 bleibt in M6-Tests und Stufe 11 grün; die fehlerhafte Hälfte tragen allein Tests des Adapters am Fake, die weder die Tabelle noch ein benanntes Negativ dafür nennt, anders als bei `LH-FA-18` Negative mit derselben Lage. Versagen: das Kriterium gilt bei der Closure als über den Admin-Listener auf der echten Ablage belegt, obwohl eine abgewiesene Anfrage mit Token dort nicht gegen den Zustand geprüft wird. | `LH-FA-13` (Negative), `AGENTS.md` §3.11, §3.15 | Plan §3 · `D02 und D04` in der Zeile `LH-FA-13` Negative | ja: Mutation X4 (die M6-Tests bleiben grün) | Kriterium nur zur Hälfte beim benannten Träger |
| F-2 | LOW | Die Stufe 11 ruft mit gültigem Token 6 der 9 schreibenden Routen: `POST /__admin/sessions` und beide `DELETE` fehlen. `SPEC-300` Nr. 12 a verlangt als Anlass „ein gültiges Token auf jeder schreibenden Route“; der Fänger im Prozess (B16) sieht ein Schreiben auf die Deskriptoren nicht, und genau diese Lücke ging an diese Stufe. X2 bleibt grün. Auch die drei Fast-Treffer des Tokens aus Nr. 12 a sendet die Stufe nicht. Kein benanntes Negativ nennt diese Menge. | `SPEC-300` Nr. 12 a, `LH-QA-04`, `AGENTS.md` §3.15 | `tests/e2e/stufe_admin_test.go` · `func adminRundlauf` | ja: Mutation X2 (die Stufe bleibt grün) | Fänger führt nur eine Form der Eingabe (Routen mit gültigem Token) |
| F-3 | LOW | Die Zeile `LH-FA-14` Negative nennt allein D05. D05 liest `capture_status` und `close`; „keine ungeschützten Bytes“ prüft D05 nicht, den Marker `MARKER-ADMIN-UNREDIGIERBAR` sendet der Fall und sucht ihn nie. Diesen Teil trägt `TestE2ESchutz`, den die Tabelle nicht nennt. | `LH-FA-14` (Negative), `AGENTS.md` §3.11 | `tests/e2e/stufe_admin_test.go` · `MARKER-ADMIN-UNREDIGIERBAR` | nein | Kriterium nur zur Hälfte beim benannten Träger |
| F-4 | LOW | Der Absatz *Nicht behauptet*, den der Slice ändert, sagt weiter, „das gebaute Binary fährt kein Lauf in `make gates`“. `make test-integration` hängt an `GATE_CHECKS` und fährt dieses Binary in seinen elf Stufen; der Satz davor („kein Gate fängt diese eine Zeile“ in `main`) steht im selben Zusammenhang. Der Satz ist älter als der Slice. | `AGENTS.md` §3.9 | `harness/README.md` · `das gebaute Binary fährt kein Lauf in` | nein | Berichtigung erreicht nicht jede Stelle |
| F-5 | INFO | Nach dem nächsten Merge von `main` stehen nach: die Zeile `LH-FA-09` Boundary der Kriterien-Tabelle (sie nennt D01, das einen **geänderten** Workspace zurücksetzt; `main` benennt `TestWorkspaceResetAufDerDatei`, den Test gibt es), die zwei Punkte *offen* in §6 und gegebenenfalls *Nicht behauptet* in `harness/README.md`. Den berichtigten Wortlaut von Nr. 12 a erfüllt die Stufe ohne Änderung (Marke unter `DATA_DIR`, Zeile auf `stderr`, `X-Correlation-Id` in den Antworten, `stdout` über `version`). | `SPEC-300` Nr. 12 a, 14; `AGENTS.md` §3.9 | Plan §6 · `Den Ausgang zieht nach, wer den Plan im Zweig führt` | nein | Nachzug nach Merge offen (Adresse eingetragen) |
| F-6 | INFO | Die Spalte *Mutation* von D02 erwartet „rot: 4 Routen“, die Messung darunter nennt 33 Zellen an 9 Routen; die Abweichung ist in der Messung vermerkt, die Tabelle nicht berichtigt. | `AGENTS.md` §3.9 | Plan §3 · `rot: 4 Routen der Zellen mit falschem Token` | nein | Plan-Tabelle hinter der Messung |
| F-7 | INFO | Der Fall *Speicher voll* legt sein `DATA_DIR` unter `/voll` an und räumt es nicht: in einem Prozess mit `-test.count=30` scheitern 25 Läufe am Start (`ODR-E9001`). Der Treiber fährt `-test.count=1` in frischem Container; eine Rate je Lauf (`AGENTS.md` §3.10) lässt sich nur mit geleertem `/voll` messen. Dasselbe Muster steht in `volleKonfig`. | Maintainability | `tests/e2e/stufe_admin_test.go` · `os.MkdirTemp("/voll", "data")` | ja: `-test.count=30` im Container | Test nicht wiederholbar im selben Container |

## Geprüft, ohne eigenes Finding

- **Abweisungs-Matrix:** die Routenliste von D02 gleicht den schreibenden Zeilen von `SPEC-005` (9, ohne Import 8); 9 × 2 + 3 × 8 × 2 = 66, Modi und Anlässe als Schleifen mit Zählprüfung. Je Zelle Status, Code, jede Tabelle, jeder Blob und genau eine Zeile mit der Korrelations-ID; die Kontrolle danach verlangt je Route Erfolg und eine Änderung.
- **Zellen Modus × Art:** 32 Zellen, 29 mit Eintrag und exakter Feldmenge; drei „gilt nicht“. Die Anpassung von `TestAuditAufrufstellen` lockert nichts (§3.6): die Gleichheit der Zellmengen gilt jetzt über 6 + 32, die Ausnahme der acht Arten entfällt.
- **Cursor-Test:** Folge 7, 11, 13, 7 mit einer Replay-Session wie `SPEC-300` Nr. 14; die ungeführte Achse (mehrere Replay-Sessions und Workspaces) steht im Plan.
- **Rundlauf:** zwei Starts auf einem `DATA_DIR` (dazu zwei zum Lesen), 11 + 2 Zeilen in Folge, Token-Datei byte-gleich in vier Starts, sechs Zellen der Lesewege mit Daten der eigenen Route.
- **Stufe 11:** die Schritte von `SPEC-037` Nr. 7 sind da; Abweisung mit Zustandsvergleich über die Lesewege; Suche vor und nach dem Herunterfahren mit Kontrollwerten (X1).
- **Geheimnisse in Ausgaben:** kein Test schreibt das Token im grünen Lauf aus; die Meldungen der Suche nennen Ort oder Länge. Meldungen mit `stderr` oder Antworten geben im Fehlerfall Prozessausgaben aus; ein Token darin ist das eines Wegwerf-`DATA_DIR`.
- **`SPEC-037`:** *Stand* („elf Stufen“) und Nr. 10 a, h, j decken sich mit Treiber und Test; `e2e-abdeckung.md` zeigt auf die Zeile von `stufe 11`.
- **Treiber und Referenzdienst:** drei Zeilen der Stufe in der Form von `SPEC-037` Nr. 3; die Route `POST` ist additiv.
- **Suppression (§3.2), Architektur-Sicht (§3.4), ADR (§3.5, §3.6), Kommentar-Klassen (§3.7):** ohne Befund.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/` (Tests des Slice) | geprüft, ohne Befund |
| `tests/integration/` | geprüft, ohne Befund am Code; F-1 trifft die Zuschreibung im Plan |
| `tests/e2e/` | geprüft; F-2, F-3, F-7 |
| `tools/harness/` | geprüft, ohne Befund |
| `harness/` | geprüft; F-4 |
| `docs/plan/planning/` | geprüft; F-1, F-5, F-6 |
| `spec/` | durch den Slice nicht geändert (Änderungen kommen aus `main`) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 3 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Kriterium nur zur Hälfte beim benannten Träger · Fänger führt nur eine Form der Eingabe · Berichtigung erreicht nicht jede Stelle · Nachzug nach Merge offen · Plan-Tabelle hinter der Messung · Test nicht wiederholbar im selben Container

## Verdikt

**Merge-blockierend:** ja, allein F-1; auflösbar im Plan (Träger oder benanntes Negativ mit Grund und Folge), ohne Änderung am Code. F-2 bis F-4 blockieren nicht.

**Übergabe:** Findings gehen an den Implementer, F-1 mit Rückkante an den Planner (Zuschreibung der Kriterien). Der Report ersetzt keine Verifikation.
