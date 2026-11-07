# Slice slice-sqlite-interrupt-transaktion: Kontext-Abbruch beendet die Transaktion des SQLite-Adapters nicht mehr

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle (Linie REPO; die Closure-Bedingung ist die DoD, keine weitere).

**Bezug:** `LH-QA-03` (Konsistenz: eine Transaktion bestätigt nur, was gelang; Fehler-Injektion), `LH-QA-06` (gleiche Eingabe, gleiches Testverhalten: der Gate-Lauf ist nicht vom Zeitpunkt eines Interrupts abhängig), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) (bestätigt, unverändert; keine Folge-ADR, siehe §3 Auftrag A), [ADR-0007](../../adr/0007-race-detektor-im-gate.md)

**Berührte Spec-Stellen:** `SPEC-108` Nr. 1 (neuer Absatz *Kontext in der Transaktion* mit den Sätzen a bis f; Nr. 17 gilt in `f` nach diesem Absatz, Nr. 17 selbst bleibt) · `SPEC-105` Nr. 13 und `SPEC-107` Nr. 12 (unverändert; der Absatz in Nr. 1 sagt, wie sie in `f` gelten) · `SPEC-109` Nr. 5 und Nr. 6 (unverändert: Kontextfehler ohne Code, jeder andere Fehler der Datenbank `ODR-E9001`, `ErrBeschaedigt` `ODR-E3001`; kein neuer Code) · `SPEC-004` (Persistenzschema, unverändert).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-sqlite-interrupt-transaktion`, Worktree `/Development/odata-recorder-wt-sqlite-interrupt`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-11-07.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Abbruch des Kontexts mitten in einer Transaktion des SQLite-Adapters (`WithWorkspaceTransaction`) lässt die Transaktion nicht mehr durch den Interrupt des Treibers zurückrollen, sodass `TestTransaktionVertrag` nicht mehr vom Zeitpunkt des Interrupts abhängt: der Abbruch endet die Operation mit dem Fehler des Kontexts nach dem Vertrag, und die Teilwirkung eines Fehlers (auch eines verschluckten) wird nie still durch ein gescheitertes `COMMIT` ersetzt.

**Beleg (Diagnose, fertig):** Zweig `diag-sqlite-interrupt`, Commit `1ad35ba3` (Worktree `/Development/odata-recorder-wt-diag-sqlite-interrupt`, Dateien `src/adapters/driven/sqlite/diag_interrupt_test.go` und `export_test.go`). Mechanismus: Der Treiber `modernc.org/sqlite` v1.60.1 startet bei Abbruch des Kontexts eine Goroutine mit `sqlite3_interrupt`; trifft sie ein schreibendes Statement, rollt SQLite die ganze Transaktion zurück (`autoCommit=1`), das `COMMIT` scheitert mit `cannot commit - no transaction is active`. Erzwungene Reihenfolge Interrupt zuerst 6 von 6 rot, Statement zuerst 6 von 6 grün; natürlicher Ablauf unter Last (8 Prozesse und 24 Busy-Loops, `-race`, `-count=1` je Prozess) 457 von 2400 rot (19 %), ohne Last 0 von 200. Der Adapter verschluckt Fehler von `ROLLBACK TO` und `RELEASE` (`zugriff.schreibe`). Die Rohprotokolle lagen im Scratchpad der Sitzung und sind kein Repo-Beleg; ihre Zusammenfassung steht in der Evidenz des Registers (`../observations/BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last/evidence/slice-sqlite-interrupt-transaktion.md`).

**Entscheidung des Users:** Die Behebung ist freigegeben (Planner-Auftrag, 2026-11-07). Welche der Varianten gilt, hat der Architect entschieden (§3, Auftrag A: Variante A mit B in der kleinen Form).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung der Treiberversion oder ein Patch des Treibers — die Wahl des Treibers ist [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); eine Behebung im Adapter genügt, und ein Wechsel wäre eine neue ADR.
- Die Änderung eines Gates, von `GATE_CHECKS` oder der Zeitgrenze von `make test-race` — die Gate-Kette bleibt unberührt (Nr. 7 der Übergangsregeln); wird die Laufzeit der Diagnose-Tests im Gate zum Problem, kürzt der Architect die Tests, nicht das Gate ([ADR-0007](../../adr/0007-race-detektor-im-gate.md)).
- Andere Wege, auf denen der Kontext eines Aufrufs Zustand beendet (Abbruch außerhalb einer offenen Transaktion, `Open`, Wartung) — hier steht nur die offene Transaktion (`zugriff` mit `conn != nil`) im Beleg; jede Operation außerhalb ist durch `inTx` geschützt und nicht beobachtet. Der Architect hat den Weg außerhalb (eine Operation = eine Transaktion) gelesen und denselben Mechanismus nicht gefunden: ein Interrupt beendet dort die eine Transaktion, und die Operation endet mit dem Fehler ohne Wirkung, was `SPEC-108` Nr. 17 sagt. Der Weg außerhalb teilt mit dem Weg in `f` den `COMMIT`-Pfad (`inTx`); die Änderung am `COMMIT` (Kontext prüfen, dann ohne Abbruch ausführen) wirkt auf beide, das Ergebnis außerhalb bleibt gleich (Fehler ohne Wirkung), und das gilt als Teil dieses Slice, nicht als Erweiterung des Umfangs (Entscheidung des Planners, 2026-11-07). Fänger dort sind die bestehenden Fälle `BrichNachLoeschenAb` der Repositories und `transaktionAbschluss`; eine Zeile der Tabelle kommt dafür nicht hinzu. Nicht Teil des Slice bleibt jede andere Änderung am Weg außerhalb. `Open` und Wartung hat er nicht gelesen (akzeptiertes Negativ, §6).
- Eine Lockerung des Tests `TestTransaktionVertrag` (Variante C der Diagnose: nur den Test anpassen) — der Vertrag (`SPEC-108` Nr. 1) sagt den Fehler des Kontexts und die Atomarität zu, und der Adapter bricht sie; ein angepasster Test verdeckte das Rot, ohne die Ursache zu beheben (`AGENTS.md` §3.6). Der Architect hält C nicht für richtig (§3, Auftrag A); es geht nichts an den Planner zurück.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Der Abbruch des Kontexts in einer offenen Transaktion beendet sie nicht mehr durch den Interrupt des Treibers: Interrupt zuerst und Statement zuerst ergeben dieselbe, im Vertrag entschiedene Wirkung (Fehler des Kontexts nach `SPEC-108` Nr. 1, `COMMIT` bestätigt nur, was gelang), belegt durch einen Regressionstest mit erzwungener Reihenfolge (aus `TestDiagInterruptReihenfolge`, `LH-QA-03`).
- [ ] Jede Zusage der Fehlerabbildung bei Kontext-Abbruch in der Transaktion hat einen Fänger, je Zusage eine Mutation rot gesehen, Raten mit Zahl der Prozesse, Last und Messendem (`AGENTS.md` §3.10, §3.15); `TestTransaktionVertrag` bleibt unter der Last der Diagnose (8 Prozesse, 24 Busy-Loops, `-race`) in jedem Prozesslauf grün (`LH-QA-06`).
- [ ] Die Diagnose-Tests sind umgeformt: was ein Regressionstest ist, bleibt im Gate mit begrenzter Laufzeit, was Messung war (`DIAG_N`, natürlicher Ablauf), geht nicht ins Gate.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: die Fehlerabbildung bei Kontext-Abbruch in der Transaktion steht in der Spezifikation (`SPEC-108` Nr. 1, Absatz *Kontext in der Transaktion*, Sätze a bis f; vom Architect vor dem Code geschrieben); das Binary und der Katalog der Fehlercodes bleiben unverändert (in §7 vermerkt), das Handbuch ist nicht berührt.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben: `BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last/` trägt die belegte Ursache, den Beleg-Pfad in `evidence/` und den Zielort der Behebung in `state.md`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. (Vom Architect geschrieben; der Implementer trägt nach der Behebung den Stand nach.)
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/adapters/driven/sqlite/zugriff.go` | update | neuer Typ für die Verbindung der offenen Transaktion (Vorschlag `txVerbindung`): prüft den Kontext vor jedem `ExecContext`, `QueryContext` und `QueryRowContext` und führt das Statement mit `context.WithoutCancel` aus; `schreibe` führt `SAVEPOINT` über ihn, `ROLLBACK TO` und `RELEASE` ohne Kontext und gibt deren Fehler weiter (Zustand *unbrauchbar*); `zugriff` trägt ihn statt der rohen `*sql.Conn` |
| `src/adapters/driven/sqlite/tx.go` | update | `COMMIT`: Kontext prüfen, dann ohne Kontext ausführen (`inTx`, auch für Operationen außerhalb von `f`: gleiches Ergebnis, `COMMIT` nur nicht mehr unterbrechbar) |
| `src/adapters/driven/sqlite/workspace_tx.go` | update | nach `f` den Zustand *unbrauchbar* als Fehler der Transaktion liefern; `zugriff` mit dem neuen Typ aufbauen |
| `src/adapters/driven/sqlite/diag_interrupt_test.go` → `transaktion_abbruch_test.go` | git mv, dann update | aus dem Diagnose-Commit `1ad35ba3` (cherry-pick, §4), `git mv` in eigenem Commit, dann Umformung nach Auftrag G (Zeilen 5 und 6) |
| `src/adapters/driven/sqlite/transaktion_haltepunkt_test.go`, `transaktion_unbrauchbar_test.go`, `aufrufstellen_test.go` | neu | die Tests der Zeilen 7 bis 9, 12 und 13 (Haltepunkt am Kontext), 15 bis 18 (Zustand *unbrauchbar*) und 19 (Zählung der Aufrufstellen und der Steuer-Statements); drei Dateien statt einer, damit jede unter der Grenze der Funktionslänge des Lint-Profils bleibt |
| `src/adapters/driven/sqlite/export_test.go` | update | `DiagTransaktionAktiv` wird `TransaktionAktiv`; `StatementAufVerbindung` für den Eingriff der Zeilen 15 und 18 |
| `src/hexagon/ports/outbound/outboundtest/transaktion_vertrag.go` | update | Fälle für die Zeilen 1 bis 4, 10 (zwei verschluckte Fehler), 12 und 19 der Tabelle (Menge der 19 Operationen; adapterunabhängig, ohne Zählung von Prüfungen) |
| `src/hexagon/ports/outbound/outboundtest/transaktion_operationen.go` | neu | die 19 Operationen als Tabelle (`Operationen()`, Art, Zahl der Prüfungen, Aufruf, Aufruf mit Fehler der Datenbank), die Arten des Kontextfehlers und die Fälle der Zeilen 1 bis 4 und 19; geteilt zwischen der Suite und den Haltepunkt-Tests im Paket `sqlite` |
| `spec/spezifikation.md` | update (vom Architect geschrieben) | `SPEC-108` Nr. 1: Absatz *Kontext in der Transaktion*, nur eigene Sätze a bis f (sieben neue Zeilen nach dem Absatz zur Atomarität) |
| `docs/plan/planning/observations/BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last/` | update (vom Architect geschrieben) | `state.md` und `evidence/slice-sqlite-interrupt-transaktion.md` (Auftrag I) |

**Aufträge an den Architect zum Start** (Start, Planner, 2026-11-07, `main` = `dcfe4712`; vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Architect arbeitet im Zweig `slice-sqlite-interrupt-transaktion`; Ausgang jeweils *offen* in §6). Wo etwas vollständig ist, genügt die Bestätigung mit Nummer.

- **A. Entscheidung A, B oder C der Diagnose, mit Gründen.** A: in der offenen Transaktion (`zugriff` mit `conn != nil`) vor jedem `ExecContext`, `QueryContext` und `QueryRowContext` den Kontext prüfen und dessen Fehler liefern, das Statement mit `context.WithoutCancel(ctx)` ausführen. B: `zugriff.schreibe` gibt Fehler von `ROLLBACK TO` und `RELEASE` weiter (nur zusammen mit A zu bewerten). C: nur den Test anpassen (nicht empfohlen, siehe §1). Der Architect nennt je Variante, was sie fängt und was nicht (insbesondere: das Fenster zwischen Prüfung und Statement; `rows` einer offenen Abfrage, die beim Abbruch geschlossen werden; `BEGIN`, `COMMIT`, `ROLLBACK` und `SAVEPOINT`; Aufrufe von `QueryRowContext`, die erst beim `Scan` laufen), und belegt die Wahl mit einem Lauf. Eine Variante D, die er findet, steht mit Gründen daneben.
- **B. Vertrag vor dem Code (`AGENTS.md` §3.12).** Die Fehlerabbildung bei Kontext-Abbruch in der Transaktion steht in der Spezifikation (Ort nach `SPEC-108` Nr. 1, `SPEC-109` Nr. 5): Abbruch vor der ersten Operation, zwischen zwei Operationen, mitten in einer Operation, nach der letzten Operation vor dem `COMMIT`, während des `COMMIT`; verschluckt `f` den Fehler des Kontexts und liefert `nil`, was ist das Ergebnis der Transaktion (bestätigt oder zurückgerollt, welcher Fehler)? Das Verhältnis zu Nr. 12 des Vertrags der Repositories (`SPEC-107`, „ohne eine Änderung“) und zur Zusage „scheitert eine Operation, nimmt sie ihre Teilwirkung zurück“ ist entschieden, nicht dem Code überlassen.
- **C. Raster Operation × Argumentart (§6, `AGENTS.md` §3.16).** Operation {jede Operation der Repositories in `f`: Workspace, Entity, Cursor, Snapshot; `COMMIT`; `ROLLBACK TO`; `RELEASE`} × Zustand des Kontexts {nicht abgebrochen, vor der Operation abgebrochen, während der Operation abgebrochen, nach der Operation abgebrochen, `DeadlineExceeded`, `Canceled` mit Ursache} × Wiederholung, jede Zelle mit Ort im Vertrag oder „gilt nicht“ mit Grund. Zusatztabelle mit Fänger je Paar von Verstößen (Kontextfehler und Fehler der Operation in derselben Transaktion; verschluckter Kontextfehler und Fehler des `COMMIT`).
- **D. Tabelle Zusage · Fänger mit Mutation je Zusage (`AGENTS.md` §3.10, §3.15).** Jede Zusage des neuen Vertrags hat eine Zeile mit Nr., Zusage, Fänger, Mutation, Klasse (`nebenläufig` oder `andere`, `SPEC-128` Nr. 3); die Spalten, die das Werkzeug aus `SPEC-128` liest. Jedes Test-Symbol der Tabelle existiert (Register `tabelle-nennt-test-der-nicht-existiert`). Der Fänger der Reihenfolge Interrupt zuerst erzwingt sie über einen Haltepunkt an einem Argument, das der Test übergibt (`ctx`, Fake, Treiber-Hook), und braucht keine Last; die Mutation „Prüfung des Kontexts entfernt“ oder „Statement mit dem Kontext des Aufrufs“ wird in jedem Lauf rot.
- **E. Mengen mit Zählprüfung gegen die Quelle** (Register `menge-im-plan-weicht-vom-orakel-des-tests-ab`, `menge-ohne-ihre-positionen`): die Aufrufstellen von `ExecContext`, `QueryContext` und `QueryRowContext` in den Repositories (gezählt im Quelltext, mit Zahl in der Tabelle, `zugriff.schreibe` und alle Leser eingeschlossen) und die Operationen von `WithWorkspaceTransaction` bilden die Mengen; die Allaussage „an jeder Stelle, in jeder Operation“ ist eine Schleife über beide Mengen, nicht eine Liste der Fälle der Diagnose. Die Position des Abbruchs ist eine Menge (vor, in, nach jeder Operation, vor dem `COMMIT`).
- **F. Raten, Last, Messender** (`AGENTS.md` §3.10; Register `rate-ohne-detektor-schwankt-mit-der-last`, `nebenlaeufiger-lauf-nur-im-glueck-gruen`, `raten-messung-mit-zusatzlast-im-selben-container`). Der Test, der das Rot der Diagnose fängt, wird in jedem Lauf rot (erzwungene Reihenfolge); die natürliche Messung unter Last (8 Prozesse und 24 Busy-Loops, `-race`, `-count=1` je Prozess; Vergleichswert 457 von 2400, 19 %, ohne Last 0 von 200) wird nach der Behebung wiederholt und mit Zahl der Prozesse, Last und Messendem im Plan genannt. Der Absatz in §3 trägt die Tabelle und die Raten (`AGENTS.md` §3.10), der Bericht allein genügt nicht. Je Fenster der Sperre (zwischen Prüfung und Statement; zwischen `RELEASE` und `COMMIT`) eine Zeile mit Wirkungs-Fänger oder benanntem Negativ mit Grund und Folge.
- **G. Umformung der Diagnose-Tests.** Der Architect entscheidet, welcher Teil Regressionstest wird (erzwungene Reihenfolge, beide Seiten), welcher gelöscht wird (Messung), und wie lange der Rest im Gate läuft (heute etwa 5,4 s; Stufe `race` gesondert nennen); `DIAG_N` und jede Umgebungsvariable, die ein Gate-Lauf nicht setzt, gehören nicht in den Regressionstest. Die Dateinamen tragen den Zweck, nicht das Wort „diag“, sobald sie Regressionstest sind (Kommentar-Regeln `AGENTS.md` §3.7, §3.11: der Kopf nennt nur, was der Test fängt).
- **H. Schnitt und Marke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1). Der Architect meldet *a*, *b*, *f* und *p* aus der Tabelle und setzt im Block `groessenmarke` weder `basis` noch `schaetzung` noch `marke`; den Block setzt der Planner nach der Bestätigung in eigenem Commit vor dem ersten Code-Commit. Wird *S* ≥ 2223, wird vor dem Code geschnitten (Kandidat: Behebung gegen Umformung der Diagnose-Tests; der Schnitt muss einen grünen Stand lassen, Register `schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices`) oder `schnitt-grund` in §4. Produktcode wird überschätzt, Tests unterschätzt (Register `schaetzformel-ueberschaetzt-den-slice`); der Architect benennt das in der Meldung, ohne die Formel zu ändern.
- **I. Register nachziehen.** `BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last/`: `state.md` trägt die belegte Ursache (Mechanismus, Zahlen der Diagnose) und den Zielort (dieser Slice, später die Spezifikationsstelle und der Regressionstest); eine Datei `evidence/slice-sqlite-interrupt-transaktion.md` fasst die Rohprotokolle der Diagnose zusammen (Reihenfolge erzwungen 6/6 und 6/6, natürlich 457 von 2400 und 0 von 200, Befehl, Last, Messender), weil das Scratchpad kein Repo-Beleg ist. Keine Zählerzeile; der Zähler folgt aus den Dateien.
- **J. Kennungen und Katalog.** Der Architect nimmt nur freie Nummern; der Vertrag steht als neue Sätze in bestehenden Nummern (`SPEC-108` Nr. 1). Braucht er eine neue Nummer oder einen neuen Fehlercode (`SPEC-201` Nr. 3 und 11), geht die Frage als Rückgabe an den Planner, bevor er schreibt. Er bestätigt, dass kein Fehlercode entsteht und `make katalog-check` unberührt ist.
- **K. Pfade neben den Zweigen.** Der Slice schreibt `src/adapters/driven/sqlite/`, gegebenenfalls `src/hexagon/ports/outbound/outboundtest/` und `spec/spezifikation.md` (nur eigene Sätze; `slice-snapshot-use-case` schreibt dort `SPEC-108` Nr. 12 bis 16 im anderen Zweig) sowie die Evidenz des Registers. Der Architect bestätigt die Liste und nennt jede Datei darüber hinaus (Abweichung an den Planner). Eine Änderung von `Makefile`, `harness/mk/` oder `GATE_CHECKS` ist ausgeschlossen (Nr. 7 der Übergangsregeln); verlangt eine Lösung sie, geht das als Rückgabe an den Planner.
- **L. Reports der Rollen und Betriebsregeln.** Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (Register `report-der-rolle-faellt-docs-check`); der Planner reicht den Satz an Reviewer und Verifier weiter. Der Implementer schreibt Repo-Dateien nur mit `Edit` und `Write` (Register `betriebsregel-fuer-repo-dateien-verletzt`), prüft vor dem Handoff, dass jedes Test-Symbol der Tabelle existiert, und fährt Raten in einem Container mit `--memory` und wenigen parallelen Läufen (Speicherlimit der Läufe).

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Fake eines Ports hält jedes Argument fest und zählt jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15).

### Antworten des Architect (2026-11-07)

Die Spezifikation trägt den Vertrag (`SPEC-108` Nr. 1, Sätze a bis f), bevor Code entsteht. Die Entscheidungen stehen hier; Gründe und Alternativen, nicht die Regeln im Einzelnen (`AGENTS.md` §3.8).

**A. Entscheidung: Variante A, dazu B in der kleinen Form; C und D verworfen.**

- *Ursache und Hebel.* Der Treiber ruft `sqlite3_interrupt` nur für einen Aufruf, dessen Kontext ein `Done` trägt, und nur, solange der Aufruf läuft. Ein Statement, das mit `context.WithoutCancel(ctx)` läuft, löst keinen Interrupt aus; die Reihenfolge von Interrupt und Statement, die das Rot bestimmte, entsteht nicht mehr, statt dass man sie entscheidet. A ist darum die einzige Variante, die die Ursache beseitigt und nicht ihre Wirkung abfängt.
- *A, was sie fängt.* Jedes Statement der offenen Transaktion, gleich ob schreibend (`ExecContext`), lesend (`QueryContext`, `QueryRowContext`) oder `SAVEPOINT`: Der Kontext wird vor dem Statement geprüft, das Statement läuft ohne Abbruch zu Ende. Offene `rows` werden beim Abbruch nicht mehr von `database/sql` geschlossen; eine Liste kommt vollständig (`SPEC-108` Nr. 1 c). `QueryRowContext` läuft erst beim `Scan`: ist der Kontext beendet, gibt die Verbindung ihn an `sql.Conn.QueryRowContext` weiter, und `database/sql` liefert den Fehler des Kontexts aus `Row.Scan`, ohne den Treiber zu rufen (Zeile 3 der Tabelle fängt eine Änderung dieses Verhaltens). `COMMIT` prüft den Kontext der Transaktion und läuft ohne Abbruch; `ROLLBACK`, `ROLLBACK TO` und `RELEASE` liefen schon ohne Abbruch und laufen weiter so.
- *A, was sie nicht fängt, und warum das gilt.* (1) Das Fenster zwischen Prüfung und Statement: ein Abbruch genau dazwischen lässt das Statement laufen; das ist kein Fehler, sondern Satz c (die Operation gelingt, oder die nächste Prüfung bemerkt den Abbruch, Satz b). Zeilen 7 bis 9 fangen es mit einem Haltepunkt am Kontext. (2) `BEGIN IMMEDIATE` behält den Kontext des Aufrufs: ein Interrupt dort beendet keine Transaktion (noch keine offen), und das Warten auf die Sperre ist nicht zugesagt (Satz e). (3) Der Weg außerhalb von `f` (`conn == nil`) bleibt: sein Statement bekommt weiter den Kontext, ein Interrupt beendet die eine Transaktion der Operation, und das Ergebnis ist der Fehler ohne Wirkung. Teilt `inTx` den Weg mit `f`, so ändert sich dort nur das `COMMIT` (nicht mehr unterbrechbar, vorher geprüft); die Fälle `transaktionAbschluss` und die Fälle der Repositories mit `BrichNachLoeschenAb` fangen das Ergebnis dort unverändert.
- *B, in der kleinen Form.* Nach A kann kein Kontext mehr das Zurücknehmen scheitern lassen; bleibt ein Fehler der Datenbank selbst (E/A, Speicher). Ihn zu verschlucken bricht still die Zusage „scheitert eine Operation, nimmt sie ihre Teilwirkung zurück“, denn `f` darf den Fehler verschlucken und die Transaktion bestätigte den Rest. Die kleine Form: der Fehler von `ROLLBACK TO` und `RELEASE` wird geliefert und macht die Transaktion *unbrauchbar* (Satz f): jede weitere schreibende Operation endet mit ihm, die Transaktion wird zurückgerollt, auch bei `nil` von `f`. Kosten etwa zwölf Zeilen (ein Feld und drei Prüfungen); die größere Form (Wiederholen, Neuaufbau der Verbindung) wäre eine eigene Entscheidung ohne Anlass.
- *C verworfen.* Der Vertrag sagt den Fehler des Kontexts und die Atomarität zu (`SPEC-108` Nr. 1 und 17), und der Adapter bricht sie unter Last in 19 % der Durchläufe; ein angepasster Test verdeckte das (`AGENTS.md` §3.6). Der Vertrag gilt, und der Planner bekommt nichts zurück.
- *D, geprüft und verworfen.* (D1) Nur Schreibwege schützen, Lesen mit Kontext: das Interrupt-Flag gilt für die ganze Verbindung (jedes laufende Statement), ein Lesen mit offenem `rows` während einer Schreib-Operation reißt es mit; zwei Regeln statt einer. (D2) Den Treiber über `sql.Conn.Raw` ohne Kontext rufen: gleiche Wirkung wie `WithoutCancel`, mehr Code und Abhängigkeit von der Treiber-Schnittstelle. (D3) Nach dem letzten Statement noch einmal prüfen und dann zurücknehmen, damit ein Abbruch im letzten Statement immer ein Fehler ist: nimmt eine Wirkung zurück, die gelang, und löst nichts, denn ein Abbruch eine Nanosekunde nach der letzten Prüfung bleibt von einem davor nicht zu unterscheiden; Satz c erlaubt beides schon. (D4) Treiberwechsel oder Patch: Gegenstand von [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), ausgeschlossen (§1).
- *Entscheidungs-Bezug.* [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) gilt unverändert (A wählt keinen anderen Treiber und gibt keine seiner Zusagen auf); [ADR-0007](../../adr/0007-race-detektor-im-gate.md) unberührt. Eine Folge-ADR entsteht nicht, weil die Entscheidung im Adapter liegt und ihre Regeln in der Spezifikation stehen (`AGENTS.md` §3.8); die Gründe stehen hier und in der Evidenz des Registers.
- *Beleg mit einem Lauf.* Ein Prototyp der Varianten A und B (Scratchpad der Sitzung, kein Repo-Code, etwa 80 Zeilen in `zugriff.go`, `tx.go` und `workspace_tx.go`) lief im Container `odata-recorder:race` (`--memory 3g`, ohne Netz, unter `flock`, Messender der Architect, 2026-11-07). (1) Erzwungene Reihenfolge *Interrupt zuerst* (500 ms Warten nach dem Abbruch): in beiden Unterfällen `Delete` und `ReplaceAll` Operation `context.Canceled`, Transaktion aktiv, `COMMIT` gelungen, Workspace `w2` erhalten, Teilwirkung zurückgenommen (die Prüfung des Zustands ohne Behebung meldet erwartungsgemäß rot). (2) Alle übrigen Tests des Pakets `sqlite` grün, `TestTransaktionVertrag` eingeschlossen (`go test -race -skip TestDiagInterruptReihenfolge`, 55 s). (3) Natürlicher Ablauf unter der Last der Diagnose (8 gleichzeitige Prozesse und 24 Busy-Loops im selben Container, `-race`, `-test.count=1` je Prozess, 150 Durchläufe je Unterfall und Prozess): **0 von 2400** rot, gegen 457 von 2400 (19 %) vor der Behebung; je Prozess 191 bis 197 s. Das ist ein Lauf des Prototyps, kein Nachweis des Codes im Slice; der Implementer wiederholt ihn nach der Behebung (Absatz zu den Raten).

**B. Vertrag.** Bestätigt, geschrieben und ohne neue Nummer: `SPEC-108` Nr. 1 trägt den Absatz *Kontext in der Transaktion* mit den Sätzen a (vor dem ersten Statement), b (später, vor dem letzten Statement), c (letztes Statement, Liste, nach der Operation), d (Bestätigen und Zurückrollen), e (Beginn), f (unbrauchbar). Verschluckt `f` den Fehler einer Operation und liefert `nil`: bestätigt, wenn der Kontext der Transaktion lebt (die gescheiterte Operation hat nichts geändert, Satz b), sonst zurückgerollt mit dem Fehler des Kontexts (Satz d). Nr. 17 (`SPEC-108`) bleibt wörtlich; `SPEC-105` Nr. 13 und `SPEC-107` Nr. 12 gelten für `f` nach Satz a bis c. `SPEC-109` Nr. 5 und 6 bleiben: der Kontextfehler hat keinen Code (Nr. 6), der Fehler aus dem Zurücknehmen ist ein Fehler der Datenbank (`ErrBeschaedigt` `ODR-E3001`, sonst `ODR-E9001`, Nr. 5).

**C. Raster.** Siehe §6.

**D. Tabelle.** Siehe die Tabelle unten. Die Namen der Typen in der Spalte *Mutation* sind Vorschläge; die Mutation sitzt an der beschriebenen Stelle, gleich wie der Implementer sie benennt. Die Test-Symbole der Spalte *Fänger* sind die Namen, die der Implementer anlegt; vor dem Handoff prüft er, dass jedes existiert.

**E. Mengen mit Zählprüfung.**

| Menge | Zahl | Quelle der Zählung |
|---|---|---|
| Aufrufstellen von `ExecContext`, `QueryContext`, `QueryRowContext`, die in `f` laufen können | 24: 9 `ExecContext`, 5 `QueryContext`, 10 `QueryRowContext` | Quelltext: `workspaces.go` 8 (4/1/3), `entities.go` 4 (2/1/1), `cursors.go` 5 (2/1/2), `workspace_snapshots.go` 3 (1/1/1), `baselines.go` 3 (`gibtEs`, `SnapshotRepository.Get` und `.List`: 0/1/2), `zugriff.go` 1 (`dienstDesWorkspace`: 0/0/1); die Aufrufe von `Coverage` in `baselines.go` laufen über `db.sql` und nicht in `f` |
| Statements der Steuerung | 7: `BEGIN IMMEDIATE`, `ROLLBACK`, `COMMIT` in `tx.go`; `SAVEPOINT`, `ROLLBACK TO`, `RELEASE` (Fehlerpfad), `RELEASE` (Erfolgspfad) in `zugriff.go` | Quelltext |
| Operationen in `f` | 19: 11 lesend (`Workspaces.Get`, `.List`; `Entities.Get`, `.List`, `.ListSet`; `Cursors.Get`, `.List`; `Snapshots.Get`, `.List`, `.GetWorkspace`, `.ListWorkspace`; davon 6 Listen und 5 Einzelne) und 8 schreibend (`Workspaces.Create`, `.Advance`, `.Delete`; `Entities.Put`, `.ReplaceAll`; `Cursors.Put`, `.ResetAll`; `Snapshots.CreateWorkspace`) | Methoden der vier Felder von `outbound.WorkspaceStores` (`reflect`, Summe der `NumMethod`) |
| Arten des Kontextfehlers | 3: `Canceled`, `DeadlineExceeded`, `Canceled` mit Ursache (`WithCancelCause`) | Satz a |
| Position des Abbruchs je Operation | Prüfungen der Operation in einem gelungenen Lauf: `Workspaces.Create` 5, `.Get` 1, `.List` 1 (leer 2), `.Advance` 3, `.Delete` 4; `Entities.Put` 3, `.Get` 1, `.List` 1 (leer 2), `.ListSet` 1 (leer 2), `.ReplaceAll` 3 plus Zahl der Einträge; `Cursors.Put` 4, `.Get` 1, `.List` 1 (leer 2), `.ResetAll` 3; `Snapshots.Get` 1, `.List` 1, `.CreateWorkspace` 5, `.GetWorkspace` 1, `.ListWorkspace` 1 (leer 2) | gelesen am Quelltext der Operationen (jede Prüfung ist ein Statement, einschließlich `SAVEPOINT`); der Test hält die Zahl je Operation fest und bestätigt sie mit einem zählenden Kontext, dessen Abweichung rot ist |

Die Zählprüfung der Aufrufstellen ist ein Test, der den Quelltext der sechs Dateien parst (`go/parser`) und je Datei und Art die Zahl festhält; eine Stelle mehr oder weniger ist rot. Dazu hält er fest, dass in den vier Dateien der Repositories und in `baselines.go` kein Aufruf an der Verbindung vorbei (`conn`, `db.sql`) läuft, ausgenommen `Coverage`. Die Allaussagen („vor jeder Operation“, „in jeder Position“) sind Schleifen über *Operationen × Arten des Kontextfehlers* und *Operationen × Position*, nicht Listen.

**F. Raten, Last, Messender; Fenster der Sperre.**

- *Vergleichswert und Protokoll.* Vor der Behebung: 457 von 2400 Durchläufen rot (19 %), 8 Prozesse, 24 Busy-Loops, `-race`, `-count=1` je Prozess, Container `--memory 3g`, Messender die Diagnose-Sitzung (Evidenz des Registers). Nach der Behebung misst der Implementer in demselben Aufbau (`odata-recorder:race`, `--memory 3g`, ohne Netz, unter `flock`, wenige parallele Läufe) und trägt in §3 die Zahlen ein: (1) die Diagnose-Messung mit den Durchläufen von Commit `1ad35ba3` (nur im Container, nicht im Repo): 8 Prozesse, 24 Busy-Loops, 150 Durchläufe je Unterfall und Prozess, Erwartung 0 von 2400; (2) `TestTransaktionVertrag` und die Tests der Zeilen 1 bis 19 in mindestens 20 Prozessläufen mit `-test.count=1` ohne Last und mindestens 20 unter derselben Last, Erwartung 0 rot; (3) je Mutation der Zeilen 5 und 6 dieselben 20 Prozessläufe ohne und unter Last, Erwartung rot in 20 von 20. Eine Einzelzahl gilt als Größenordnung; ohne Detektor schwankt sie mit der Last (`AGENTS.md` §3.10).
- *Zeilen 5 und 6 sind nur mit einem Zeitfenster erzwungen.* *Interrupt zuerst* wartet 500 ms im Statement; der Treiber meldet nicht, wann der Interrupt angewendet ist. Unter extremer Last kann die Goroutine des Treibers später als 500 ms laufen, die Mutation bliebe dann grün. Die Diagnose maß 6 von 6; die Rate nach (3) steht im Plan. Die Zeilen 7 bis 9 fangen dieselbe Klasse (Statement mit dem Kontext des Aufrufs) ohne Zeitfenster und ohne Goroutine, mit einem Haltepunkt am Kontext, der in jedem Lauf rot wird.
- *Fenster der Sperre* („der Schreiber hält die Sperre für die Dauer von `f`“, `SPEC-108` Nr. 1; jedes Fenster ist ein Schritt, in dem der Kontext enden kann, ohne dass die Sperre frei wird):

| Fenster | Fänger oder Negativ |
|---|---|
| zwischen `BEGIN` und der ersten Operation | Zeilen 1 bis 3 (Kontext der Operation bei Beginn von `f` bereits beendet) |
| zwischen Prüfung und Statement (je Art: `Exec`, `Query`, `QueryRow`) | Zeilen 9, 7, 8 (Haltepunkt am Kontext: der Abbruch folgt dem `Err`-Aufruf der Prüfung, das Statement läuft, der nächste Aufruf scheitert) |
| zwischen zwei Statements einer Operation | Zeile 10 (Position als Menge) |
| zwischen `ROLLBACK TO` und `RELEASE` (Fehlerpfad) | Zeile 10: beide laufen ohne Kontext, eine Mutation mit dem Kontext des Aufrufs macht den Fall rot; ein Abbruch dazwischen hat keine Wirkung, weil beide den Kontext nicht lesen |
| zwischen dem `RELEASE` der letzten Operation und der Prüfung vor `COMMIT` | Zeile 12 (Haltepunkt am Kontext der Transaktion: der `Err`-Aufruf vor `COMMIT` liefert den Fehler, zurückgerollt) |
| zwischen der Prüfung vor `COMMIT` und `COMMIT` | Zeile 13 (der Abbruch folgt der Prüfung, bestätigt, `nil`) |
| zwischen `COMMIT` und der Rückgabe der Verbindung an den Pool | Negativ: nach `COMMIT` ist nichts mehr zu schützen, ein Abbruch ändert die bestätigte Wirkung nicht; allein ein fehlgeschlagenes `ROLLBACK` bei einem Fehler könnte die Verbindung mit offener Transaktion zurückgeben (Zeile 21) |

**G. Umformung der Diagnose-Tests.**

- *Regressionstest.* `TestDiagInterruptReihenfolge` wird `TestTransaktionAbbruchImStatement` in `transaktion_abbruch_test.go` (`git mv` in eigenem Commit vor der Umformung, `AGENTS.md` §3.3). Beide Modi (*Interrupt zuerst*, *Statement zuerst*) und beide Unterfälle (`Delete`, `ReplaceAll`) laufen **einmal** statt dreimal; beide Modi erwarten dasselbe: Operation `context.Canceled`, Transaktion aktiv, `COMMIT` gelungen, `w2` erhalten, Teilwirkung zurückgenommen. Die Erwartung „Rollback der ganzen Transaktion“ des Modus *Interrupt zuerst* entfällt.
- *Gelöscht.* `TestDiagInterruptNatuerlich`, `DIAG_N`, das Protokoll im Speicher (`protokoll`, `ereignis`, `diagBericht` mit Zeitstempeln, `b2i`): sie waren Messung; ihre Zahlen stehen in der Evidenz des Registers, der Quelltext im Commit `1ad35ba3`. Der Test meldet bei Abweichung Operationsfehler, Zustand der Transaktion, `COMMIT`-Fehler und Zustand der Tabellen, ohne Zeitstempel.
- *Namen.* `export_test.go`: `DiagTransaktionAktiv` wird `TransaktionAktiv`; kein Symbol und kein Dateiname trägt „diag“ oder „Diag“ mehr. Der Kopf des Tests nennt, was er fängt (Statement, das der Abbruch des Aufrufs nicht unterbricht; beide Reihenfolgen gleich) und was offen bleibt (das Zeitfenster von 500 ms, Zeilen 5 und 6).
- *Laufzeit.* Heute etwa 5,4 s im Gate (6 Durchläufe *Interrupt zuerst* zu je 0,5 s Warten und 6 *Statement zuerst*, zweimal je Unterfall). Danach 2 Durchläufe mit Warten von 0,5 s (je Unterfall einer) und 2 ohne: etwa 1,2 s je Stufe, `test` wie `race`. Der Implementer misst beide Stufen und trägt die Zahlen in §3 ein; über 2,5 s kürzt er das Warten, nicht den Umfang der Fälle (Untergrenze: die Rate nach (3) bleibt 20 von 20). Das Gate bleibt unberührt.

**H. Meldung für die Marke (`SPEC-128` Nr. 1).** Aus der Tabelle: *a* = 6 (Zeilen 5 bis 9 und 13), *b* = 15, *f* = 1 (der Kontext mit Haltepunkt der Zeilen 7 bis 13), *p* = 90. Daraus *S* = 85·6 + 80·15 + 200·1 + 90 = 2000 und *M* = 2700; keine Schnittfrage (*S* < 2223). Der Block steht unten; `basis`, `schaetzung` und `marke` hat der Planner in eigenem Commit vor dem ersten Code-Commit gesetzt. Produktcode ist überschätzt (der Prototyp umfasst etwa 80 Zeilen in drei Dateien); die Tests sind unterschätzt, denn die Schleifen über die 19 Operationen, die Zählprüfung und die Umformung der Diagnose-Tests wiegen mehr als 80 Zeilen je Zeile der Tabelle. Die Formel bleibt, wie sie ist (Register `schaetzformel-ueberschaetzt-den-slice`); eine Überschreitung trägt der Implementer mit Grund unter `ueberschreitung` nach.

```groessenmarke
pfad: src/adapters/driven/sqlite/
pfad: src/hexagon/ports/outbound/outboundtest/
produktcode: 90
fakes: 1
basis: e82372390145be87c2b4f260eeb3901ba4d63a33
schaetzung: 2000
marke: 2700
```

**Größenmarke** (Planner, 2026-11-07, `SPEC-128` Nr. 1): der Block trägt `pfad`, `produktcode` und `fakes` des Architect und `basis`, `schaetzung` und `marke` des Planner, in eigenem Commit vor dem ersten Code-Commit gesetzt; die Basis ist der Commit des Architect. *a* = 6, *b* = 15, *f* = 1, *p* = 90: *S* = 85·6 + 80·15 + 200·1 + 90 = 2000, *M* = 2700, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. `make groessenmarke` meldet `Schätzung 2000, Marke 2700`.

**I. Register.** Geschrieben: `state.md` (Ursache belegt, Zielort dieser Slice, danach `SPEC-108` Nr. 1 und der Regressionstest) und `evidence/slice-sqlite-interrupt-transaktion.md` (Zusammenfassung der Rohprotokolle: erzwungen 6/6 und 6/6, natürlich 457 von 2400 und 0 von 200, Befehl, Last, Messender). Keine Zählerzeile.

**J. Kennungen und Katalog.** Bestätigt: keine neue Nummer (neue Sätze a bis f in `SPEC-108` Nr. 1), kein neuer Fehlercode; `SPEC-109` Nr. 5 und 6 und `make katalog-check` bleiben unberührt; das Binary ändert sich im Verhalten nur durch die behobene Atomarität, nicht im Katalog.

**K. Pfade.** Bestätigt die Liste der Tabelle oben. Eine Anmerkung zu §1: `inTx` gehört dem Weg in `f` und dem Weg außerhalb; die Änderung am `COMMIT` wirkt auf beide (Ergebnis gleich, `COMMIT` nur nicht mehr unterbrechbar), ohne dass eine Datei dazukommt. Es kommt keine Datei hinzu, `Makefile`, `harness/mk/` und `GATE_CHECKS` bleiben (die Umformung der Tests kürzt keine Gate-Zeitgrenze).

**L. Reports und Betriebsregeln.** Zur Kenntnis genommen; der Satz zu den Links der Kennungen von ADRs geht über den Planner an Reviewer und Verifier. Anmerkung zur Betriebsregel für Repo-Dateien: dem Architect stand in diesem Lauf kein `Edit` zur Verfügung; die sieben Zeilen in der `spec/spezifikation.md` (eine Datei von etwa 1 MB, nicht als Ganzes mit `Write` zu schreiben) sind aus einer mit `Write` geschriebenen Datei zwischen die Zeilen 761 und 762 gesetzt worden (`head` und `tail` in eine neue Datei, kopiert), ohne `sed -i`, `perl -pi` oder Anhängen per Heredoc, danach auf Rohbytes gegengelesen.

**Tabelle Zusage · Fänger** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3 a; *Ort* ist der Satz in `SPEC-108` Nr. 1; ein Fänger der Suite läuft für jeden Adapter, ein Fänger im Paket `sqlite` nur dort):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Satz a, Art `Exec`: ist der Kontext einer schreibenden Operation (8) vor ihrem ersten Statement beendet, liefert sie seinen Fehler (3 Arten) ohne Änderung | Suite `transaktionKontextVorOperation` (Schleife 19 Operationen × 3 Arten; Zustand der Tabellen unverändert, `w2` der gelungenen Operation erhalten) für das Ergebnis; die Prüfung in `ExecContext` selbst fängt `TestTransaktionAbbruchHaltepunkt/position` (jede schreibende Operation führt nach dem Savepoint zuerst ein `QueryRow`, dessen Prüfung denselben Fehler liefert, so dass die Suite die Mutation nicht sieht) | Prüfung in `txVerbindung.ExecContext` entfernt | andere |
| 2 | Satz a, Art `Query`: dasselbe für die 6 Listen | `transaktionKontextVorOperation` | Prüfung in `txVerbindung.QueryContext` entfernt | andere |
| 3 | Satz a, Art `QueryRow`: dasselbe für die 5 Einzelnen; der Fehler kommt aus `Row.Scan` | `transaktionKontextVorOperation` | Prüfung in `txVerbindung.QueryRowContext` entfernt (das Statement läuft ohne Abbruch, die Operation gelingt) | andere |
| 4 | Satz a, Vorrang: der Fehler des Kontexts geht dem Fehler der Datenbank vor (unbekannte Kennung, bekannte Kennung) in jeder Operation, die einen solchen hat | Suite `transaktionKontextVorDatenbankfehler` (Schleife über die Operationen mit `ErrNichtGefunden` oder `ErrBereitsVorhanden` an einem Zustand, der ihn auslöst, mit beendetem Kontext) | Prüfung in `txVerbindung.QueryRowContext` entfernt (das Statement läuft, `Row.Scan` liefert den Fehler der Datenbank); für `Exec` und `Query` gibt es keine beobachtbare Mutation „danach“ (der Savepoint ist das erste Statement jeder schreibenden Operation, und die Fehler der Datenbank der Listen kommen aus einem `QueryRow`), akzeptiertes Negativ | andere |
| 5 | Satz 1 (nie währenddessen) und Sätze b und c: der Abbruch unterbricht ein Statement der offenen Transaktion nicht; Reihenfolge *Interrupt zuerst* (Treiber-Hook im Statement ruft den Abbruch und wartet 500 ms): Operation `context.Canceled`, Transaktion aktiv, `COMMIT` gelungen, `w2` erhalten, Teilwirkung zurück (`Delete`, `ReplaceAll`) | `TestTransaktionAbbruchImStatement/interrupt-zuerst` | `context.WithoutCancel` in `txVerbindung.ExecContext` entfernt (das Statement läuft mit dem Kontext des Aufrufs, der Treiber unterbricht, `COMMIT` scheitert) | nebenläufig |
| 6 | dasselbe, Reihenfolge *Statement zuerst* (`GOMAXPROCS(1)`): gleiches Ergebnis; kein Fehler des Kontexts mit Wirkung | `TestTransaktionAbbruchImStatement/statement-zuerst` | `txVerbindung.ExecContext` rollt nach dem Statement die ganze Transaktion zurück, wenn der Kontext beendet ist (das Bild des Interrupts nach dem Statement) | nebenläufig |
| 7 | Satz c, Art `Query`: der Abbruch nach der Prüfung des Statements einer Liste ändert nichts, die Liste kommt vollständig | `TestTransaktionAbbruchHaltepunkt/query` (Kontext mit Haltepunkt: der `Err`-Aufruf der Prüfung liefert `nil` und beendet danach den Kontext; Schleife über die 6 Listen) | `context.WithoutCancel` in `txVerbindung.QueryContext` entfernt (die Abfrage scheitert mit dem Fehler des Kontexts) | nebenläufig |
| 8 | Satz c, Art `QueryRow`: dasselbe für die 5 Einzelnen | `TestTransaktionAbbruchHaltepunkt/queryrow` | `context.WithoutCancel` in `txVerbindung.QueryRowContext` entfernt | nebenläufig |
| 9 | Sätze b und c, Art `Exec`, Position als Menge: Abbruch nach der Prüfung des k-ten Statements, für k von 1 bis zur Zahl der Prüfungen der Operation (Tabelle der Mengen): bei k kleiner der Zahl Fehler des Kontexts ohne Änderung, bei k gleich der Zahl gelungen samt Wirkung | `TestTransaktionAbbruchHaltepunkt/position` (Schleife 8 schreibende Operationen × k; Zahl je Operation aus der Tabelle der Mengen, bestätigt durch den Probelauf) | `context.WithoutCancel` in `txVerbindung.ExecContext` entfernt (bei k gleich der Zahl scheitert die Operation) | nebenläufig |
| 10 | Satz b und Atomarität: scheitert eine Operation zwischen zwei Statements am Kontext, nimmt sie die Teilwirkung zurück, auch wenn `f` den Fehler verschluckt; `ROLLBACK TO` und `RELEASE` lesen den Kontext nicht; die gelungene Operation derselben Transaktion bleibt; zwei verschluckte Fehler derselben Transaktion (Kontext, Datenbank) | Suite `transaktionTeilwirkung` (bestehend, Unterfälle `Delete` und `ReplaceAll`), Suite `transaktionVerschluckteFehler` (zwei verschluckte Fehler: Kontext, Datenbank) und `TestTransaktionAbbruchHaltepunkt/position` (k kleiner der Zahl: Zustand der Tabellen unverändert) | `ROLLBACK TO` mit dem Kontext des Aufrufs statt ohne; `ROLLBACK TO` weggelassen | andere |
| 11 | Satz c: trifft der Abbruch das letzte Statement nach der Prüfung, gelingt die Operation samt Wirkung, oder sie scheitert ohne sie; nie ein Fehler mit Wirkung | Suite `transaktionAbschluss` (bestehend) und `TestTransaktionAbbruchHaltepunkt/position` (k gleich der Zahl) | Ergebnis der Operation nach dem Statement durch den Fehler des Kontexts ersetzt, Wirkung bleibt | andere |
| 12 | Satz d: ist der Kontext der Transaktion vor dem Bestätigen beendet, wird zurückgerollt und sein Fehler geliefert, auch wenn `f` `nil` liefert; die Operationen mit eigenem, lebendem Kontext gelingen; 3 Arten des Fehlers; die nächste Transaktion ist nutzbar | Suite `transaktionFall4` (bestehend, um die Arten und die Folge-Transaktion erweitert) und `TestTransaktionAbbruchHaltepunkt/bestaetigen-vorher` (Haltepunkt am Kontext der Transaktion beim `Err`-Aufruf vor `COMMIT`) | Prüfung vor `COMMIT` entfernt (bestätigt trotz beendetem Kontext); `ROLLBACK` mit dem Kontext (die Verbindung kehrt mit offener Transaktion in den Pool, die nächste Transaktion scheitert) | andere |
| 13 | Satz d: trifft der Abbruch das Bestätigen nach der Prüfung, wird bestätigt und `nil` geliefert | `TestTransaktionAbbruchHaltepunkt/bestaetigen-nachher` (der `Err`-Aufruf vor `COMMIT` liefert `nil` und beendet danach den Kontext der Transaktion) | `context.WithoutCancel` vor `COMMIT` entfernt (`COMMIT` scheitert mit dem Fehler des Kontexts) | nebenläufig |
| 14 | Satz e: ist der Kontext vor dem Beginn beendet, wird `f` nicht gerufen und sein Fehler geliefert | Suite `transaktionFall4` (bestehend, erster Teil; nun je Art des Kontextfehlers) | `d.sql.Conn` und `BEGIN IMMEDIATE` beide ohne Kontext des Aufrufs (jede der beiden Prüfungen allein genügt, eine Mutation nur an einer bleibt grün: gleichwertige Prüfungen) | andere |
| 15 | Satz f: scheitert `ROLLBACK TO`, liefert die Operation diesen Fehler | `TestTransaktionUnbrauchbar/operation` (ein Eingriff in `f` löst mit `RELEASE operation` den Fehler `no such savepoint` aus, danach scheitert die Operation und mit ihr `ROLLBACK TO`) | Fehler von `ROLLBACK TO` wieder verworfen | andere |
| 16 | Satz f: jede weitere schreibende Operation in `f` endet mit diesem Fehler; eine lesende liefert weiter | `TestTransaktionUnbrauchbar/weitere` (Schleife über die 8 schreibenden Operationen; dazu die 11 lesenden) | Zustand *unbrauchbar* wird von der Prüfung des Statements nicht gelesen | andere |
| 17 | Satz f: die Transaktion wird zurückgerollt und liefert den Fehler, auch wenn `f` `nil` liefert; liefert `f` einen Fehler, gilt dieser unverändert; ein Kontextfehler der Operation geht dem Fehler des Zurücknehmens nicht voraus, sondern wird von ihm abgelöst | `TestTransaktionUnbrauchbar/transaktion` | `workspace_tx.go` liefert den Zustand nicht (bestätigt trotz Fehler); der Fehler von `f` wird durch den Zustand ersetzt | andere |
| 18 | Satz f, Erfolgspfad: scheitert `RELEASE` nach gelungener Operation, ist die Transaktion ebenso unbrauchbar | `TestTransaktionUnbrauchbar/release` (ein Eingriff in `f` löst den Fehler vor dem Abschluss der Operation aus) | Fehler von `RELEASE` auf dem Erfolgspfad verworfen | andere |
| 19 | Mengen: 24 Aufrufstellen (9 / 5 / 10) und 7 Steuer-Statements in den sechs Dateien, keine Stelle an der Prüfung vorbei; 19 Operationen in `WorkspaceStores` | `TestAufrufstellenZaehlung` (parst die Dateien), `TestSteuerStatementsZaehlung` (die 7 Statements der Steuerung als Literale) und Suite `transaktionOperationenMenge` (`reflect` zählt die Methoden der vier Felder) | eine Stelle ruft `conn` oder `db.sql` unmittelbar; ein Port erhält eine Methode, die Suite nicht | andere |
| 20 | Akzeptiertes Negativ, Satz e: ein Abbruch während des Wartens auf die Sperre in `BEGIN IMMEDIATE` wirkt nicht zugesagt. Grund: der Treiber meldet das Ende des Wartens nicht, und die Wartezeit begrenzt `busy_timeout`. Folge: ein abgebrochener Aufruf kann bis zum Ende des Wartens hängen und gelingt oder scheitert dann ohne Transaktion; keine Wirkung auf Daten | kein Fänger | nicht anwendbar | andere |
| 21 | Akzeptiertes Negativ: scheitert das `ROLLBACK` am Ende einer Transaktion (Fehler der Datenbank), bleibt der Fehler verworfen und die Verbindung kehrt womöglich mit offener Transaktion in den Pool. Grund: ohne Fehlerinjektion im Treiber nicht auszulösen, und der Beleg kennt keinen Fall. Folge: die nächste Transaktion auf dieser Verbindung scheitert mit dem Fehler der Datenbank (`cannot start a transaction within a transaction`), nicht still | kein Fänger | nicht anwendbar | andere |
| 22 | Satz f, Reihenfolge Argument vor Zustand: eine weitere schreibende Operation mit ungültigem Argument in einer unbrauchbaren Transaktion liefert `ErrUngueltig` (nicht den Fehler der Transaktion), und der Zustand bleibt gesetzt; Schleife über die 8 schreibenden Operationen, je mit einem ungültigen Argument, das die Operation vor ihrem ersten Statement erkennt | `TestTransaktionUnbrauchbar/argument` (ergänzt `weitere`) | die Prüfung des Zustands in `schreibe` vor die Prüfung der Argumente gezogen oder der Zustand von `ErrUngueltig` überschrieben | andere |
| 23 | Satz f gegen Satz a, Reihenfolge Zustand vor Kontext: eine weitere schreibende Operation mit gültigem Argument und beendetem Kontext in einer unbrauchbaren Transaktion liefert den Fehler der Transaktion, nicht den des Kontexts (3 Arten des Kontextfehlers, 8 Operationen); der Kontext ist beim Aufruf schon beendet, kein Haltepunkt | `TestTransaktionUnbrauchbar/kontext` (Schleife Operationen x Arten) | die Prüfung des Zustands in `schreibe` hinter den `SAVEPOINT` (die Prüfung des Kontexts) gesetzt | andere |

**Mutationen und Raten** (Implementer, 2026-11-07; Messender der Implementer; Container `odata-recorder:race`, `--memory 3g`, ohne Netz, unter `flock`, ein Lauf zur Zeit; der Host war ausgelastet, Load-Average 8 bis 70 während der Läufe; eine Rate gilt als Größenordnung, `AGENTS.md` §3.10). Jede Mutation wurde auf einer frischen Kopie des Arbeitsbaums im Container angewandt und die Tests der Transaktion gefahren (`-race`); die Spalte nennt die roten Tests, die Zahl in Klammern die roten Elemente der Schleife.

| Nr. | Mutation (wie gefahren) | rote Tests |
|---|---|---|
| 1 | Bedingung der Prüfung in `ExecContext` durch `error(nil)` ersetzt | `Haltepunkt/position` (8 von 8 Operationen), `AbbruchImStatement` (beide Modi), `Teilwirkung`, `Unbrauchbar` (4); die Suite `KontextVorOperation` bleibt grün (siehe Zeile 1) |
| 2 | dasselbe in `QueryContext` | Suite `KontextVorOperation` (6 Listen), `Haltepunkt/query` (6) |
| 3 | `if ctx.Err() != nil` in `QueryRowContext` abgeschaltet | `KontextVorOperation` (5 Einzelne), `Vorrang` (5), `Haltepunkt/queryrow` (5), `Haltepunkt/position` (8), `Unbrauchbar/release` (die Zahl der Prüfungen verschiebt sich) |
| 4 | wie 3 (Prüfung nach dem Statement: `Row.Scan` liefert den Fehler der Datenbank) | `Vorrang` (5 Einzelne) |
| 5, 9 | `ExecContext` ohne `context.WithoutCancel` | `AbbruchImStatement/interrupt-zuerst` (`Delete`, `ReplaceAll`), `Haltepunkt/position` (8 von 8) |
| 6 | `ExecContext` rollt nach dem Statement die ganze Transaktion zurück, wenn der Kontext beendet ist | `AbbruchImStatement/statement-zuerst` und `interrupt-zuerst` (je 2), `Teilwirkung` (2), `Abschluss`, `Haltepunkt/position`, `SteuerStatementsZaehlung` |
| 7 | `QueryContext` ohne `context.WithoutCancel` | `Haltepunkt/query` (6 von 6) |
| 8 | `QueryRowContext` ohne `context.WithoutCancel` | `Haltepunkt/queryrow` (5 von 5), `Haltepunkt/position` (8) |
| 10 | a) `ROLLBACK TO` und `RELEASE` mit dem Kontext des Aufrufs; b) `ROLLBACK TO` weggelassen | a) `Teilwirkung` (2), `Abschluss`, `AbbruchImStatement` (4), `Haltepunkt/position`, `Unbrauchbar` (3); b) `Teilwirkung` (2), `AbbruchImStatement` (4), `Haltepunkt/position` (`Delete`, `ReplaceAll`), `SteuerStatementsZaehlung` |
| 11 | die Operation liefert nach dem letzten `RELEASE` den Fehler des Kontexts, die Wirkung bleibt | `Abschluss`, `Haltepunkt/position` (8 von 8, `k` gleich der Zahl) |
| 12 | a) Prüfung vor `COMMIT` abgeschaltet; b) `ROLLBACK` mit dem Kontext | a) `Fall4`, `Haltepunkt/bestaetigen-vorher` und `-nachher`; b) `Fall4`, `Haltepunkt/bestaetigen-vorher` |
| 13 | `COMMIT` mit dem Kontext statt `WithoutCancel` | `Haltepunkt/bestaetigen-nachher` |
| 14 | `d.sql.Conn` und `BEGIN IMMEDIATE` beide ohne Kontext (nur eine von beiden: grün, die zweite Prüfung genügt) | `Fall4` |
| 15 | der Fehler von `ROLLBACK TO` und `RELEASE` wird der Operation nicht geliefert (der Zustand bleibt gesetzt) | `Unbrauchbar/operation`, `Unbrauchbar/weitere` |
| 16 | `schreibe` liest den Zustand *unbrauchbar* nicht | `Unbrauchbar/weitere` (8 schreibende Operationen) |
| 22 | Prüfung des Zustands vor die Prüfung der Argumente (oder Zustand ersetzt `ErrUngueltig`) | `Unbrauchbar/argument` (8 schreibende Operationen) |
| 23 | Prüfung des Zustands hinter den `SAVEPOINT` | `Unbrauchbar/kontext` (8 Operationen x 3 Arten) |
| 17 | a) `WithWorkspaceTransaction` liefert den Zustand nicht; b) der Fehler von `f` wird durch den Zustand ersetzt | a) `Unbrauchbar/transaktion`, `/release`; b) `Unbrauchbar/transaktion` |
| 18 | `RELEASE` auf dem Erfolgspfad verworfen | `Unbrauchbar/release` |
| 19 | a) `Advance` ruft `r.z.db.sql` unmittelbar; b) `r.z.conn.conn`; c) `ReplaceAll` ohne `DELETE`; d) `Operationen()` nennt `Snapshots.Listx` | a) `AufrufstellenZaehlung`, `Haltepunkt/position` (`Advance`); b) `AufrufstellenZaehlung`, `Löschen` (Absturz); c) `AufrufstellenZaehlung`, `Teilwirkung/ReplaceAll`, `Reset`, `Restore`, `AbbruchImStatement/*/ReplaceAll`, `Haltepunkt/position` (`ReplaceAll`); d) Suite `OperationenMenge` |
| (Zahl der Prüfungen) | eine Prüfung mehr in `ExecContext` | `Haltepunkt/position` (8 von 8, Probelauf), `Unbrauchbar` (3) |

**Rate je Prozesslauf** (`-test.count=1`, `-race`, ein Binary je Mutation; rot = Exit-Code ungleich 0; ohne Last 20 Läufe nacheinander, unter Last 24 Läufe in 3 Runden zu 8 Prozessen neben 24 Busy-Loops im selben Container):

| Lauf | ohne Last | unter Last (8 Prozesse, 24 Busy-Loops) |
|---|---|---|
| unverändert: `TestTransaktion*`, `TestAufrufstellenZaehlung`, `TestSteuerStatementsZaehlung` (Zeilen 1 bis 19) | rot 0 von 20 | rot 0 von 24 |
| Zeile 5, Mutation 5: `TestTransaktionAbbruchImStatement` | rot 20 von 20 | rot 24 von 24 |
| Zeile 9, Mutation 5: `Haltepunkt/position` | rot 20 von 20 | rot 24 von 24 |
| Zeile 6, Mutation 6: `AbbruchImStatement/statement-zuerst` | rot 20 von 20 | rot 24 von 24 |
| Zeile 7, Mutation 7: `Haltepunkt/query` | rot 20 von 20 | rot 24 von 24 |
| Zeile 8, Mutation 8: `Haltepunkt/queryrow` | rot 20 von 20 | rot 24 von 24 |
| Zeile 13, Mutation 13: `Haltepunkt/bestaetigen-nachher` | rot 20 von 20 | rot 24 von 24 |

Die Haltepunkt-Tests kennen kein Zeitfenster; das Warten von 500 ms in den Zeilen 5 und 6 (`AbbruchImStatement`) hielt in diesen Läufen unter Last (Rate 24 von 24), ist aber weiter ein Fenster, das der Treiber nicht schließt.

**Natürlicher Ablauf der Diagnose auf dem Code des Slice** (Auftrag F (1)): die Tests von Commit `1ad35ba3` (`TestDiagInterruptNatuerlich`, `DIAG_N=150`, nur im Container, nicht im Repo) mit der Umbenennung `DiagTransaktionAktiv` auf `TransaktionAktiv`; 8 Prozesse zu 150 Durchläufen je Unterfall (`Delete`, `ReplaceAll`), `-race`, `-test.count=1` je Prozess, 24 Busy-Loops, `--memory 3g`: **0 von 2400** rot (Vergleichswert vor der Behebung 457 von 2400, 19 %; je Prozess etwa 157 s). Messender der Implementer.

**Laufzeit von `TestTransaktionAbbruchImStatement`** (Auftrag G), gemessen im Container ohne Last: Stufe `race` 1,9 s (`interrupt-zuerst` 1,4 s, `statement-zuerst` 0,5 s), Stufe `test` ohne `-race` 1,6 s; die Grenze von 2,5 s hält, das Warten von 0,5 s bleibt. Das Paket `sqlite` läuft in der Stufe `race` etwa 62 s (davor etwa 55 s im Prototyp des Architect; `TestTransaktionVertrag` etwa 13 s, `Haltepunkt` etwa 15 s, `Unbrauchbar` etwa 1 s).

**Mengen mit Zählprüfung** (Auftrag E; `AGENTS.md` §3.17, Klasse generierbar): *Operationen × Art des Kontextfehlers* — Generator `Operationen()` × `kontextArten()`, Orakel der Fehler `errors.Is(err, Art)` und der unveränderte Zustand der Tabellen, die Zählprüfung `OperationenMenge` gegen die Methoden der vier Felder von `outbound.WorkspaceStores` (`reflect`): 19 × 3 = 57 Elemente; eine nicht genannte Mutation je Achse, rot gesehen: *Operation* — `Cursors.ResetAll`, `Snapshots.ListWorkspace` und `Entities.Get` nehmen den Kontext nicht (je nur ihre Zeile rot); *Art* — `QueryContext` ohne Prüfung (Mutation 2) und `QueryRowContext` (Mutation 3). *Position* (8 schreibende Operationen, Prüfungen 5, 3, 4, 3, 4, 4, 3, 5): 31 Elemente `k`, Orakel Fehler und unveränderter Zustand für `k` kleiner der Zahl, gelungen samt Wirkung für `k` gleich der Zahl; Probelauf zählt die Prüfungen. *Aufrufstellen*: 24 (9 `Exec`, 5 `Query`, 10 `QueryRow`) und 7 Statements der Steuerung, gezählt durch `go/parser`; die Mutationen 19 a bis c. Die Zahlen der Tabelle der Mengen (E) stimmen mit dem Probelauf überein. Akzeptiert: die Menge der Fehler der Datenbank je Operation ist die eine Form `ErrNichtGefunden` oder `ErrBereitsVorhanden` (Zeile 4); die Fehler `ErrFalscheArt`, `ErrFremderService` und `ErrDienstUnbekannt` der Operationen führt kein Fall der Zeile 4, Folge: ein Vorrang des Kontexts vor diesen Fehlern ist durch die Prüfung vor dem ersten Statement mitgetragen (jeder dieser Fehler entsteht nach einem Statement), aber nicht einzeln gefangen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): Die Diagnose liegt vor (Commit `1ad35ba3` im Zweig `diag-sqlite-interrupt`), und der User hat die Behebung freigegeben (2026-11-07). Ein Vorgänger-Slice ist nicht nötig: der Slice schreibt nur den Adapter `sqlite`, seine Tests und eigene Sätze der Spezifikation. Der Zweig `slice-sqlite-interrupt-transaktion` entsteht nach dem Start-Commit von `main`; der Diagnose-Commit wird per `git cherry-pick` aus `diag-sqlite-interrupt` übernommen, damit die Tests mit ihrer Geschichte im Zweig liegen.
**Stand beim Start** (Planner, 2026-11-07, `main` = `dcfe4712`): erfüllt. §1 *Ausdrücklich NICHT* und DoD der Nachbarn gelesen (`AGENTS.md` §3.13): `slice-snapshot-use-case` (in Arbeit; er schreibt `SPEC-108` Nr. 12 bis 16 und nicht den Adapter), `slice-batch-changesets`, `slice-etag-versionen` und `slice-if-match` (nutzen `WithWorkspaceTransaction` und die Atomarität nach `SPEC-108` Nr. 1; ein geschärfter Vertrag bei Kontext-Abbruch gilt für sie unverändert, keiner schließt eine Sendung dieses Slice aus). Die Gate-Kette bleibt unberührt (Nr. 7 der Übergangsregeln): der Slice ändert weder `GATE_CHECKS` noch ein Gate (Auftrag K). Speicher beim Start: 14 GB verfügbar. **Marke:** beim Start nicht gesetzt; sie steht seit der Tabelle des Architect im Block `groessenmarke` in §3 (*M* = 2700, siehe *Stand nach dem Architect*). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).
**Stand nach dem Architect** (2026-11-07): die Tabelle liegt vor (§3, 21 Zeilen), der Vertrag steht in `SPEC-108` Nr. 1, das Raster in §6. Die Meldung für die Marke steht in §3 Auftrag H (*S* = 2000, *M* = 2700); der Block steht dort vollständig: `pfad`, `produktcode` und `fakes` vom Architect, `basis`, `schaetzung` und `marke` vom Planner in eigenem Commit vor dem ersten Code-Commit.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Marke ist *M* = 2700 (Block in §3).
- `in-progress` → `open` (blockiert — Carveout?): keine der Varianten A bis D (Auftrag A) behebt das Rot, ohne einen Treiberwechsel ([ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), neue ADR) oder eine Änderung der Gate-Kette (Nr. 7 der Übergangsregeln). Der Prototyp zu Variante A (§3) zeigt, dass A trägt; die Rückführung ist damit unwahrscheinlich, nicht ausgeschlossen (die Raten nach der Behebung stehen aus).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10), und `TestTransaktionVertrag` ist unter der Last der Diagnose in keinem der gemessenen Prozessläufe rot (Zahl, Last und Messender im Plan); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Das Fenster zwischen der Prüfung des Kontexts und dem Statement: ein Abbruch genau dazwischen lässt das Statement mit abgebrochenem Kontext laufen, falls es nicht ohne ihn läuft — **Ausgang:** offen bis zur Closure; entschieden: das Statement läuft ohne Abbruch (Variante A), Fänger Zeilen 7 bis 9 der Tabelle (Haltepunkt am Kontext); Adresse Implementer (Test, Raten).
- Lese-Statements und offene `rows` in der Transaktion lösen beim Abbruch denselben Interrupt aus und beenden die Transaktion, falls nur Schreibwege geschützt werden — **Ausgang:** offen bis zur Closure; entschieden: alle drei Arten laufen ohne Abbruch (Menge der 24 Aufrufstellen, Auftrag E), Fänger Zeilen 2, 3, 7, 8 und 19; Adresse Implementer.
- Fehler von `ROLLBACK TO` und `RELEASE` weiterreichen (Variante B) ändert, was `f` sieht, wenn es einen Fehler verschluckt; Aufrufer (`slice-batch-changesets`) verlassen sich auf die Atomarität — **Ausgang:** offen bis zur Closure; entschieden: Satz f (Operation liefert den Fehler, Transaktion unbrauchbar und zurückgerollt); Fänger Zeilen 15 bis 18; Adresse Implementer. Die Aufrufer sehen den Fehler nur bei einem Fehler der Datenbank selbst.
- Laufzeit der umgeformten Diagnose-Tests in der Stufe `race` (heute etwa 5,4 s im Gate) — **Ausgang:** offen bis zur Closure; entschieden: ein Durchlauf je Modus und Unterfall, erwartet etwa 1,2 s, Grenze 2,5 s (Auftrag G); Adresse Implementer (Messung beider Stufen).
- Randformen des Vertrags (Raster, Auftrag C) — **Ausgang:** entschieden vor dem ersten Code-Commit: das Raster unten hat in jeder Zelle einen Ort in `SPEC-108` Nr. 1 oder „gilt nicht“ mit Grund; keine Randform ohne Ort. Der Ausgang (*entfallen*) wird bei der Closure bestätigt.
- Aufträge A bis L an den Architect zum Start (§3) — **Ausgang:** beantwortet (§3 *Antworten des Architect*); die zwei Rückgaben an den Planner (Marke, `COMMIT` außerhalb von `f`) sind erledigt, siehe unten.
- Das Rot lag nur unter Last und im Lauf `race` vor; ein Gate, das den Lauf nicht fährt, sieht den Fehler nicht — **Ausgang:** offen bis zur Closure; die erzwungenen Fänger der Zeilen 7 bis 9 brauchen weder Last noch `-race`, die Rate nach der Behebung (Auftrag F) bestätigt es; Adresse Implementer.
- Die Marke (Adresse Planner): Der Block in §3 ist vollständig (*S* = 2000, *M* = 2700, `basis` = Commit des Architect; Auftrag H). **Ausgang:** entfallen; gesetzt in eigenem Commit vor dem ersten Code-Commit, `make groessenmarke` grün.
- **Offen, Adresse Planner:** `inTx` ist dem Weg in `f` und dem Weg außerhalb gemeinsam, §1 schließt den Weg außerhalb aus. Die Änderung am `COMMIT` (Prüfung, dann ohne Kontext) wirkt auf beide; das Ergebnis außerhalb ist gleich (Fehler ohne Wirkung), die bestehenden Fälle der Repositories fangen es. Der Architect hält das für vertretbar und nimmt es nicht als Erweiterung des Umfangs; bestätigt der Planner nicht, ist die Alternative ein eigener `COMMIT`-Pfad für `f` (mehr Code, zwei Wege). **Ausgang:** entfallen; der Planner bestätigt (2026-11-07): ein gemeinsamer `COMMIT`-Pfad, vertretbar, keine Erweiterung des Umfangs; §1 ist präzisiert, die Fänger außerhalb sind die bestehenden Fälle `BrichNachLoeschenAb` und `transaktionAbschluss`.
- **Akzeptiertes Negativ:** `Open`, Wartung und `Migrate` hat der Architect nicht gelesen. Grund: der Beleg kennt für sie keinen Fall (alle sechs Rot-Meldungen sind `TestTransaktionVertrag`, Weg in `f`), und sie laufen durch `inTx` (eine Transaktion, ein Interrupt beendet sie ganz). Folge: findet ein Lauf dort denselben Mechanismus, ist es eine neue Beobachtung im Register. **Ausgang:** entfallen als Aufgabe dieses Slice.
- **Offen, Adresse Architect** (Implementer, Satz f): „jede weitere schreibende Operation in `f` endet mit ihm“ sagt nicht, ob der Zustand *unbrauchbar* vor oder nach der Prüfung der Argumente einer Operation gilt. Der Adapter prüft ihn in `schreibe`, nach der Prüfung der Argumente: eine schreibende Operation mit ungültigem Argument liefert weiter `ErrUngueltig` (wie bei beendetem Kontext, wo das Argument dem Kontext vorausgeht). Die Tests führen gültige Argumente (`Unbrauchbar/weitere`); die Reihenfolge bei ungültigem Argument hat keinen Fänger und keinen Ort im Vertrag. **Ausgang:** entschieden (Architect, 2026-11-07): die Reihenfolge ist Argumente, dann Zustand *unbrauchbar*, dann Kontext (Satz a), dann Datenbank; Ort `SPEC-108` Nr. 1 f (letzte Sätze); Fänger Zeile 22 (`Unbrauchbar/argument`). Die jetzige Codelage gilt unverändert.
- **Randform Paar beendeter Kontext x unbrauchbare Transaktion** (Review M-2, Adresse Architect). **Ausgang:** entschieden (Architect, 2026-11-07): der Zustand geht dem Kontext vor, denn jede weitere schreibende Operation endet mit dem Fehler der Transaktion (f), mit dem die Transaktion ohnehin endet; Satz a nennt die Ausnahme; Fänger Zeile 23 (`Unbrauchbar/kontext`). Die jetzige Codelage gilt unverändert.
- **Akzeptiertes Negativ** (Implementer): die Zeile 4 trägt als Mutation „Prüfung nach dem Statement“ nur für `QueryRow`; für `Exec` und `Query` gibt es keine beobachtbare Mutation (jede schreibende Operation führt zuerst den Savepoint, dessen Prüfung den Kontext vorrangig meldet, und die Fehler der Datenbank der Listen kommen aus einem `QueryRow`). Folge: ein Vorrang des Kontexts vor einem Fehler der Datenbank, den ein `Exec` oder `Query` selbst meldet, ist nicht einzeln gefangen; die Operationen der Menge haben keinen solchen Fehler an erster Stelle. **Ausgang:** offen bis zur Closure.

**Raster Operation × Zustand des Kontexts** (`AGENTS.md` §3.16; *Ort* ist der Satz in `SPEC-108` Nr. 1, „Nr. 1 Abs.“ der Absatz zur Atomarität davor; jede Zelle trägt Ort oder „gilt nicht“ mit Grund). Spalten: **nicht abgebrochen** · **vor dem Schritt abgebrochen** · **nach der Prüfung, im Schritt abgebrochen** · **zwischen den Schritten abgebrochen** · **`DeadlineExceeded`, `Canceled` mit Ursache** · **Wiederholung**.

| Operation | nicht abgebrochen | vor dem Schritt | nach der Prüfung, im Schritt | zwischen den Schritten | `DeadlineExceeded`, Ursache | Wiederholung |
|---|---|---|---|---|---|---|
| Beginn (`BEGIN IMMEDIATE`) | Nr. 1 (Satz 1) | Nr. 1 e | Nr. 1 e (Warten: nicht zugesagt, Zeile 20) | Nr. 1 a (die erste Operation prüft) | Nr. 1 a, e; `SPEC-109` Nr. 6 | gilt nicht: jede Transaktion beginnt neu; nach dem Zurückrollen Nr. 1 d |
| lesende Operation (11: 6 Listen, 5 Einzelne) | `SPEC-108` Nr. 3, 7, 8, 11, 14 und `SPEC-107` | Nr. 1 a | Nr. 1 c (läuft zu Ende, Liste vollständig) | Nr. 1 b (leere Liste: zweites Statement; keine Wirkung zurückzunehmen) | Nr. 1 a; `SPEC-109` Nr. 6 | Nr. 1 b (dieselbe Operation mit lebendem Kontext gelingt, nichts hat sich geändert) |
| schreibende Operation (8) | `SPEC-108` Nr. 2, 4, 5, 7, 9, 11, 12, 13 | Nr. 1 a (`SAVEPOINT` geprüft) | Nr. 1 b (früheres Statement) oder c (letztes) | Nr. 1 b | Nr. 1 a; `SPEC-109` Nr. 6 | Nr. 1 b (zweiter Aufruf in derselben Transaktion nach dem Zurücknehmen gelingt mit lebendem Kontext) |
| Zurücknehmen und Abschluss (`ROLLBACK TO`, `RELEASE`) | Nr. 1 Abs. (Atomarität) | gilt nicht: sie lesen keinen Kontext (Nr. 1 Satz 1), ein beendeter Kontext ändert sie nicht | gilt nicht, selber Grund | gilt nicht, selber Grund | gilt nicht, selber Grund; scheitert die Datenbank: Nr. 1 f | gilt nicht: nach einem Fehler ist die Transaktion unbrauchbar (Nr. 1 f), eine Wiederholung gibt es nicht |
| Bestätigen (`COMMIT`) | Nr. 1 (Satz 1) | Nr. 1 d | Nr. 1 d (bestätigt, `nil`) | Nr. 1 d (nach der letzten Operation, vor der Prüfung: zurückgerollt) | Nr. 1 d; `SPEC-109` Nr. 6 | Nr. 1 d (die nächste Transaktion ist nutzbar) |
| Zurückrollen (`ROLLBACK`, nach Fehler, Panik, beendetem Kontext) | Nr. 1 (Satz 1) | Nr. 1 d (hängt von keinem Kontext ab) | Nr. 1 d | Nr. 1 d | Nr. 1 d | Nr. 1 d (nächste Transaktion nutzbar); scheitert es selbst: Zeile 21, akzeptiertes Negativ |

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16):

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| Kontextfehler und Fehler der Datenbank in derselben Operation (`ErrNichtGefunden`, `ErrBereitsVorhanden`) | der erste Schritt, der scheitert, bestimmt den Fehler; ein beendeter Kontext vor dem ersten Statement geht vor (Nr. 1 a) | Zeile 4 |
| Kontextfehler einer Operation und Fehler einer anderen Operation derselben Transaktion, beide von `f` verschluckt | beide haben nichts geändert (Nr. 1 b); die gelungenen bleiben; lebt der Kontext der Transaktion, wird bestätigt (Nr. 1 d) | Zeile 10 (zwei verschluckte Fehler) |
| verschluckter Kontextfehler der Operation und beendeter Kontext der Transaktion beim Bestätigen | zurückgerollt, Fehler des Kontexts der Transaktion (Nr. 1 d) | Zeile 12 |
| Kontextfehler der Operation und Fehler beim Zurücknehmen | der Fehler des Zurücknehmens (Nr. 1 f), die Transaktion ist unbrauchbar, der Kontextfehler wird von ihm abgelöst | Zeile 17 |
| Fehler von `f` und unbrauchbare Transaktion | der Fehler von `f` unverändert (Nr. 1 f, letzter Halbsatz) | Zeile 17 |
| beendeter Kontext der Operation und unbrauchbare Transaktion (Satz a gegen Satz f) | der Fehler der Transaktion; der Zustand geht dem Kontext vor (Nr. 1 f, Reihenfolge der Prüfungen; Ausnahme in a) | Zeile 23 |
| ungültiges Argument und unbrauchbare Transaktion | `ErrUngueltig`, der Zustand bleibt; die Argumente gehen dem Zustand vor (Nr. 1 f, Reihenfolge der Prüfungen) | Zeile 22 |
| Modus: Kontext der Operation beendet, Kontext der Transaktion lebt | Operation nach Nr. 1 a bis c; bestätigt, wenn `f` `nil` liefert (Nr. 1 d) | Zeilen 5, 6, 10, 11 (`Teilwirkung` und `Abschluss` übergeben der Operation einen eigenen Kontext) |
| Modus: Kontext der Transaktion beendet, Kontext der Operation lebt | Operationen gelingen; zurückgerollt beim Bestätigen (Nr. 1 d) | Zeile 12 (`transaktionFall4`: `Create` mit `Background`, Abbruch des Kontexts der Transaktion in `f`) |
| Modus: beide Kontexte derselbe | jede weitere Operation endet mit dem Fehler (Nr. 1 a), `f` reicht ihn weiter oder verschluckt ihn, das Ergebnis ist der Fehler des Kontexts (Nr. 1 d) | Zeilen 1 bis 3 und 12 |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-07, Zähler aus den Dateien unter `evidence/`). Gegenstand: `BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last` (6×, Schwelle lange erreicht; der Eintrag trug bisher keinen Zielort, weil die Ursache offen war; sie ist belegt, Zielort ist dieser Slice, Auftrag I). Einschlägig für Raten und Fänger: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (8×) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×) (Auftrag F), `BEO-REPO/raten-messung-mit-zusatzlast-im-selben-container` (1×) und `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last` (1×) (Auftrag F, G), `BEO-REPO/menge-im-plan-weicht-vom-orakel-des-tests-ab` (3×) (Auftrag E), `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (6×), `BEO-REPO/tabelle-nennt-test-der-nicht-existiert` (1×) (Auftrag D), `BEO-REPO/schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices` (1×) und `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (25×) (Auftrag H), `BEO-REPO/report-der-rolle-faellt-docs-check` (5×) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (12×) (Auftrag L). Kein weiterer Eintrag erreicht mit diesem Slice 3× neu.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF.
