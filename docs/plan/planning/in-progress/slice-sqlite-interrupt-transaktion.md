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

**Bezug:** `LH-QA-03` (Konsistenz: eine Transaktion bestätigt nur, was gelang; Fehler-Injektion), `LH-QA-06` (gleiche Eingabe, gleiches Testverhalten: der Gate-Lauf ist nicht vom Zeitpunkt eines Interrupts abhängig), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md)

**Berührte Spec-Stellen:** `SPEC-108` Nr. 1 (Transaktion: „ein abgebrochener Kontext“, Atomarität jeder Operation in `f`) · `SPEC-105` und `SPEC-107` (Nr. 12 des Vertrags der Repositories: abgebrochener Kontext beendet jede Operation mit dem Fehler des Kontexts, ohne eine Änderung) · `SPEC-109` Nr. 5 (Fehlerabbildung) · `SPEC-004` (Persistenzschema, unverändert). Der Architect nennt die Stelle, an der die Fehlerabbildung bei Kontext-Abbruch in der Transaktion steht.
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

**Beleg (Diagnose, fertig):** Zweig `diag-sqlite-interrupt`, Commit `1ad35ba3` (Worktree `/Development/odata-recorder-wt-diag-sqlite-interrupt`, Dateien `src/adapters/driven/sqlite/diag_interrupt_test.go` und `export_test.go`). Mechanismus: Der Treiber `modernc.org/sqlite` v1.60.1 startet bei Abbruch des Kontexts eine Goroutine mit `sqlite3_interrupt`; trifft sie ein schreibendes Statement, rollt SQLite die ganze Transaktion zurück (`autoCommit=1`), das `COMMIT` scheitert mit `cannot commit - no transaction is active`. Erzwungene Reihenfolge Interrupt zuerst 6 von 6 rot, Statement zuerst 6 von 6 grün; natürlicher Ablauf unter Last (8 Prozesse und 24 Busy-Loops, `-race`, `-count=1` je Prozess) 457 von 2400 rot (19 %), ohne Last 0 von 200. Der Adapter verschluckt Fehler von `ROLLBACK TO` und `RELEASE` (`zugriff.schreibe`). Die Rohprotokolle liegen im Scratchpad der Sitzung (`/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/diag-sqlite/`) und sind kein Repo-Beleg; der Architect trägt eine Zusammenfassung in die Evidenz des Registers ein (Auftrag R).

**Entscheidung des Users:** Die Behebung ist freigegeben (Planner-Auftrag, 2026-11-07). Welche der Varianten gilt, entscheidet der Architect (§3, Auftrag A).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung der Treiberversion oder ein Patch des Treibers — die Wahl des Treibers ist [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md); eine Behebung im Adapter genügt, und ein Wechsel wäre eine neue ADR.
- Die Änderung eines Gates, von `GATE_CHECKS` oder der Zeitgrenze von `make test-race` — die Gate-Kette bleibt unberührt (Nr. 7 der Übergangsregeln); wird die Laufzeit der Diagnose-Tests im Gate zum Problem, kürzt der Architect die Tests, nicht das Gate ([ADR-0007](../../adr/0007-race-detektor-im-gate.md)).
- Andere Wege, auf denen der Kontext eines Aufrufs Zustand beendet (Abbruch außerhalb einer offenen Transaktion, `Open`, Wartung) — hier steht nur die offene Transaktion (`zugriff` mit `conn != nil`) im Beleg; jede Operation außerhalb ist durch `inTx` geschützt und nicht beobachtet. Findet der Architect dort denselben Mechanismus, geht der Befund als Rückgabe an den Planner.
- Eine Lockerung des Tests `TestTransaktionVertrag` (Variante C der Diagnose: nur den Test anpassen) — der Vertrag (`SPEC-108` Nr. 1) sagt den Fehler des Kontexts und die Atomarität zu, und der Adapter bricht sie; ein angepasster Test verdeckte das Rot, ohne die Ursache zu beheben (`AGENTS.md` §3.6). Hält der Architect C für richtig, weil der Vertrag so nicht gilt, geht das als Rückgabe an den Planner, bevor Code entsteht.

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
- [ ] Doku-Update: die Fehlerabbildung bei Kontext-Abbruch in der Transaktion steht in der Spezifikation (`SPEC-108` Nr. 1 oder die Stelle, die der Architect benennt); das Binary und der Katalog der Fehlercodes bleiben unverändert (in §7 vermerkt), das Handbuch ist nicht berührt.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben: `BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last/` trägt die belegte Ursache, den Beleg-Pfad in `evidence/` und den Zielort der Behebung in `state.md`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/adapters/driven/sqlite/zugriff.go` | update | Pfad `conn != nil`: Prüfung des Kontexts vor jedem Statement der offenen Transaktion und Ausführung ohne den Abbruch des Aufrufs (Variante A); Fehler von `ROLLBACK TO` und `RELEASE` (Variante B, nur zusammen mit A zu bewerten) |
| `src/adapters/driven/sqlite/tx.go`, `workspace_tx.go` | update, falls der Architect es nennt | `BEGIN`, `COMMIT`, `ROLLBACK` der Transaktion: welcher Kontext sie führt, steht im Vertrag |
| `src/adapters/driven/sqlite/diag_interrupt_test.go`, `export_test.go` | update | aus dem Diagnose-Commit `1ad35ba3` (cherry-pick, siehe §4): `TestDiagInterruptReihenfolge` wird nach A an Fall (a) rot und ist als Regressionstest umzuformen (Interrupt zuerst, `COMMIT` ok); `TestDiagInterruptNatuerlich` (`DIAG_N`) und die Laufzeit von etwa 5,4 s im Gate sind zu bewerten |
| `src/hexagon/ports/outbound/outboundtest/transaktion_vertrag.go` | update, falls der Vertrag es verlangt | Fall „Abbruch des Kontexts mitten in `f`“ für jeden Adapter der Suite; der Architect nennt, ob die Suite oder nur das Paket `sqlite` ihn trägt |
| `spec/spezifikation.md` | update | Fehlerabbildung bei Kontext-Abbruch in der Transaktion mit Randformen (Architect), nur eigene Sätze in `SPEC-108` Nr. 1 und der Stelle, die er nennt |
| `docs/plan/planning/observations/BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last/` | update | `state.md` und eine Datei in `evidence/` (Auftrag I) |

**Aufträge an den Architect zum Start** (Start, Planner, 2026-11-07, `main` = `dcfe4712`; vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Architect arbeitet im Zweig `slice-sqlite-interrupt-transaktion`; Ausgang jeweils *offen* in §6). Wo etwas vollständig ist, genügt die Bestätigung mit Nummer.

- **A. Entscheidung A, B oder C der Diagnose, mit Gründen.** A: in der offenen Transaktion (`zugriff` mit `conn != nil`) vor jedem `ExecContext`, `QueryContext` und `QueryRowContext` den Kontext prüfen und dessen Fehler liefern, das Statement mit `context.WithoutCancel(ctx)` ausführen. B: `zugriff.schreibe` gibt Fehler von `ROLLBACK TO` und `RELEASE` weiter (nur zusammen mit A zu bewerten). C: nur den Test anpassen (nicht empfohlen, siehe §1). Der Architect nennt je Variante, was sie fängt und was nicht (insbesondere: das Fenster zwischen Prüfung und Statement; `rows` einer offenen Abfrage, die beim Abbruch geschlossen werden; `BEGIN`, `COMMIT`, `ROLLBACK` und `SAVEPOINT`; Aufrufe von `QueryRowContext`, die erst beim `Scan` laufen), und belegt die Wahl mit einem Lauf. Eine Variante D, die er findet, steht mit Gründen daneben.
- **B. Vertrag vor dem Code (`AGENTS.md` §3.12).** Die Fehlerabbildung bei Kontext-Abbruch in der Transaktion steht in der Spezifikation (Ort nach `SPEC-108` Nr. 1, `SPEC-109` Nr. 5): Abbruch vor der ersten Operation, zwischen zwei Operationen, mitten in einer Operation, nach der letzten Operation vor dem `COMMIT`, während des `COMMIT`; verschluckt `f` den Fehler des Kontexts und liefert `nil`, was ist das Ergebnis der Transaktion (bestätigt oder zurückgerollt, welcher Fehler)? Das Verhältnis zu Nr. 12 des Vertrags der Repositories (`SPEC-107`, „ohne eine Änderung“) und zur Zusage „scheitert eine Operation, nimmt sie ihre Teilwirkung zurück“ ist entschieden, nicht dem Code überlassen.
- **C. Raster Operation × Argumentart (§6, `AGENTS.md` §3.16).** Operation {jede Operation der Repositories in `f`: Workspace, Entity, Cursor, Snapshot; `COMMIT`; `ROLLBACK TO`; `RELEASE`} × Zustand des Kontexts {nicht abgebrochen, vor der Operation abgebrochen, während der Operation abgebrochen, nach der Operation abgebrochen, `DeadlineExceeded`, `Canceled` mit Ursache} × Wiederholung, jede Zelle mit Ort im Vertrag oder „gilt nicht“ mit Grund. Zusatztabelle mit Fänger je Paar von Verstößen (Kontextfehler und Fehler der Operation in derselben Transaktion; verschluckter Kontextfehler und Fehler des `COMMIT`).
- **D. Tabelle Zusage · Fänger mit Mutation je Zusage (`AGENTS.md` §3.10, §3.15).** Jede Zusage des neuen Vertrags hat eine Zeile mit Nr., Zusage, Fänger, Mutation, Klasse (`nebenläufig` oder `andere`, `SPEC-128` Nr. 3); die Spalten, die das Werkzeug aus `SPEC-128` liest. Jedes Test-Symbol der Tabelle existiert (Register `tabelle-nennt-test-der-nicht-existiert`). Der Fänger der Reihenfolge Interrupt zuerst erzwingt sie über einen Haltepunkt an einem Argument, das der Test übergibt (`ctx`, Fake, Treiber-Hook), und braucht keine Last; die Mutation „Prüfung des Kontexts entfernt“ oder „Statement mit dem Kontext des Aufrufs“ wird in jedem Lauf rot.
- **E. Mengen mit Zählprüfung gegen die Quelle** (Register `menge-im-plan-weicht-vom-orakel-des-tests-ab`, `menge-ohne-ihre-positionen`): die Aufrufstellen von `ExecContext`, `QueryContext` und `QueryRowContext` in den Repositories (gezählt im Quelltext, mit Zahl in der Tabelle, `zugriff.schreibe` und alle Leser eingeschlossen) und die Operationen von `WithWorkspaceTransaction` bilden die Mengen; die Allaussage „an jeder Stelle, in jeder Operation“ ist eine Schleife über beide Mengen, nicht eine Liste der Fälle der Diagnose. Die Position des Abbruchs ist eine Menge (vor, in, nach jeder Operation, vor dem `COMMIT`).
- **F. Raten, Last, Messender (`AGENTS.md` §3.10; Register `rate-ohne-detektor-schwankt-mit-der-last`, `nebenlaeufiger-lauf-nur-im-glueck-gruen`, `raten-messung-mit-zusatzlast-im-selben-container`).** Der Test, der das Rot der Diagnose fängt, wird in jedem Lauf rot (erzwungene Reihenfolge); die natürliche Messung unter Last (8 Prozesse und 24 Busy-Loops, `-race`, `-count=1` je Prozess; Vergleichswert 457 von 2400, 19 %, ohne Last 0 von 200) wird nach der Behebung wiederholt und mit Zahl der Prozesse, Last und Messendem im Plan genannt. Der Absatz in §3 trägt die Tabelle und die Raten (`AGENTS.md` §3.10), der Bericht allein genügt nicht. Je Fenster der Sperre (zwischen Prüfung und Statement; zwischen `RELEASE` und `COMMIT`) eine Zeile mit Wirkungs-Fänger oder benanntem Negativ mit Grund und Folge.
- **G. Umformung der Diagnose-Tests.** Der Architect entscheidet, welcher Teil Regressionstest wird (erzwungene Reihenfolge, beide Seiten), welcher gelöscht wird (Messung), und wie lange der Rest im Gate läuft (heute etwa 5,4 s; Stufe `race` gesondert nennen); `DIAG_N` und jede Umgebungsvariable, die ein Gate-Lauf nicht setzt, gehören nicht in den Regressionstest. Die Dateinamen tragen den Zweck, nicht das Wort „diag“, sobald sie Regressionstest sind (Kommentar-Regeln `AGENTS.md` §3.7, §3.11: der Kopf nennt nur, was der Test fängt).
- **H. Schnitt und Marke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1). Der Architect meldet *a*, *b*, *f* und *p* aus der Tabelle und setzt im Block `groessenmarke` weder `basis` noch `schaetzung` noch `marke`; den Block setzt der Planner nach der Bestätigung in eigenem Commit vor dem ersten Code-Commit. Wird *S* ≥ 2223, wird vor dem Code geschnitten (Kandidat: Behebung gegen Umformung der Diagnose-Tests; der Schnitt muss einen grünen Stand lassen, Register `schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices`) oder `schnitt-grund` in §4. Produktcode wird überschätzt, Tests unterschätzt (Register `schaetzformel-ueberschaetzt-den-slice`); der Architect benennt das in der Meldung, ohne die Formel zu ändern.
- **I. Register nachziehen.** `BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last/`: `state.md` trägt die belegte Ursache (Mechanismus, Zahlen der Diagnose) und den Zielort (dieser Slice, später die Spezifikationsstelle und der Regressionstest); eine Datei `evidence/slice-sqlite-interrupt-transaktion.md` fasst die Rohprotokolle der Diagnose zusammen (Reihenfolge erzwungen 6/6 und 6/6, natürlich 457 von 2400 und 0 von 200, Befehl, Last, Messender), weil das Scratchpad kein Repo-Beleg ist. Keine Zählerzeile; der Zähler folgt aus den Dateien.
- **J. Kennungen und Katalog.** Der Architect nimmt nur freie Nummern; der Vertrag steht als neue Sätze in bestehenden Nummern (`SPEC-108` Nr. 1). Braucht er eine neue Nummer oder einen neuen Fehlercode (`SPEC-201` Nr. 3 und 11), geht die Frage als Rückgabe an den Planner, bevor er schreibt. Er bestätigt, dass kein Fehlercode entsteht und `make katalog-check` unberührt ist.
- **K. Pfade neben den Zweigen.** Der Slice schreibt `src/adapters/driven/sqlite/`, gegebenenfalls `src/hexagon/ports/outbound/outboundtest/` und `spec/spezifikation.md` (nur eigene Sätze; `slice-snapshot-use-case` schreibt dort `SPEC-108` Nr. 12 bis 16 im anderen Zweig) sowie die Evidenz des Registers. Der Architect bestätigt die Liste und nennt jede Datei darüber hinaus (Abweichung an den Planner). Eine Änderung von `Makefile`, `harness/mk/` oder `GATE_CHECKS` ist ausgeschlossen (Nr. 7 der Übergangsregeln); verlangt eine Lösung sie, geht das als Rückgabe an den Planner.
- **L. Reports der Rollen und Betriebsregeln.** Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (Register `report-der-rolle-faellt-docs-check`); der Planner reicht den Satz an Reviewer und Verifier weiter. Der Implementer schreibt Repo-Dateien nur mit `Edit` und `Write` (Register `betriebsregel-fuer-repo-dateien-verletzt`), prüft vor dem Handoff, dass jedes Test-Symbol der Tabelle existiert, und fährt Raten in einem Container mit `--memory` und wenigen parallelen Läufen (Speicherlimit der Läufe).

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Fake eines Ports hält jedes Argument fest und zählt jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): Die Diagnose liegt vor (Commit `1ad35ba3` im Zweig `diag-sqlite-interrupt`), und der User hat die Behebung freigegeben (2026-11-07). Ein Vorgänger-Slice ist nicht nötig: der Slice schreibt nur den Adapter `sqlite`, seine Tests und eigene Sätze der Spezifikation. Der Zweig `slice-sqlite-interrupt-transaktion` entsteht nach dem Start-Commit von `main`; der Diagnose-Commit wird per `git cherry-pick` aus `diag-sqlite-interrupt` übernommen, damit die Tests mit ihrer Geschichte im Zweig liegen.
**Stand beim Start** (Planner, 2026-11-07, `main` = `dcfe4712`): erfüllt. §1 *Ausdrücklich NICHT* und DoD der Nachbarn gelesen (`AGENTS.md` §3.13): `slice-snapshot-use-case` (in Arbeit; er schreibt `SPEC-108` Nr. 12 bis 16 und nicht den Adapter), `slice-batch-changesets`, `slice-etag-versionen` und `slice-if-match` (nutzen `WithWorkspaceTransaction` und die Atomarität nach `SPEC-108` Nr. 1; ein geschärfter Vertrag bei Kontext-Abbruch gilt für sie unverändert, keiner schließt eine Sendung dieses Slice aus). Die Gate-Kette bleibt unberührt (Nr. 7 der Übergangsregeln): der Slice ändert weder `GATE_CHECKS` noch ein Gate (Auftrag K). Speicher beim Start: 14 GB verfügbar. **Marke: noch nicht gesetzt** (Prosa, kein Block auf `main`); eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt (Auftrag H): danach setzt der Planner in eigenem Commit vor dem ersten Code-Commit den Block `groessenmarke` mit `basis`, `schaetzung` und `marke`; bis dahin bindet kein Gate. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): keine der Varianten A bis D (Auftrag A) behebt das Rot, ohne einen Treiberwechsel ([ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), neue ADR) oder eine Änderung der Gate-Kette (Nr. 7 der Übergangsregeln).

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

- Das Fenster zwischen der Prüfung des Kontexts und dem Statement: ein Abbruch genau dazwischen lässt das Statement mit abgebrochenem Kontext laufen, falls es nicht ohne ihn läuft — **Ausgang:** offen; Fänger durch Haltepunkt am Argument (Auftrag F), Architect.
- Lese-Statements und offene `rows` in der Transaktion lösen beim Abbruch denselben Interrupt aus und beenden die Transaktion, falls nur Schreibwege geschützt werden — **Ausgang:** offen; Menge der Aufrufstellen (Auftrag E), Architect.
- Fehler von `ROLLBACK TO` und `RELEASE` weiterreichen (Variante B) ändert, was `f` sieht, wenn es einen Fehler verschluckt; Aufrufer (`slice-batch-changesets`) verlassen sich auf die Atomarität — **Ausgang:** offen; Vertrag (Auftrag B), Architect.
- Laufzeit der umgeformten Diagnose-Tests in der Stufe `race` (heute etwa 5,4 s im Gate) — **Ausgang:** offen; Auftrag G.
- Randformen des Vertrags (Raster, Auftrag C) — **Ausgang:** offen; Architect, vor dem ersten Code-Commit.
- Aufträge A bis L an den Architect zum Start (§3) — **Ausgang:** offen; Adresse: Architect (Bestätigung in diesem Abschnitt oder Ergänzung), Implementer (Test), Reviewer und Verifier (Auftrag L).
- Das Rot lag nur unter Last und im Lauf `race` vor; ein Gate, das den Lauf nicht fährt, sieht den Fehler nicht — **Ausgang:** offen; der erzwungene Fänger (Auftrag D) braucht weder Last noch `-race`, die Rate nach der Behebung (Auftrag F) bestätigt es.

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
