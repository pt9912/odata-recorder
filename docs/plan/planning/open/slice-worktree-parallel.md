# Slice slice-worktree-parallel: Gates im Parallelbetrieb mehrerer Arbeitsbäume (Image-Tags je Arbeitsbaum, Sperre, Stempel)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-sensoren](../welle-sensoren.md) — Teil 0, **kein Produkt-Slice**: Er liefert keine Produktfunktion, liegt auf keinem Pfad eines Meilensteins und blockiert keinen produktiven Slice und keine produktive Welle. Er läuft in einem eigenen Worktree (Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)); sein Start-Trigger ist `slice-groessenmarke-werkzeug` in `done/`.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe: Container, netzlos; das Lastenheft führt keine Anforderung an Gates), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Rahmen der Gegenproben; dieser Slice entscheidet ihn nicht), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Verträge der Werkzeuge unter `harness/vertraege/`), [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-102` (Gegenproben: Image-Tag der Stufe `gegenproben`, Stempel) · `SPEC-122` (`test-race`) · `SPEC-128` (`make groessenmarke`) · `harness/vertraege/README.md` §Kennungen (der Architect legt einen Vertrag an oder schärft die genannten, Kennung aus dem Bereich der Linie C: `SPEC-200` bis `SPEC-229`) · `harness/README.md` §Sensors, *Nicht behauptet*.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Zwei Arbeitsbäume desselben Repos fahren die Gates und Prüfläufe auf einem Docker-Daemon, ohne einander zu stören, und die Übergangsregeln des Parallelbetriebs (`planning/README.md`, Nr. 1 und 2) werden Code: (a) die Image-Tags der Gate-Stufen (`odata-recorder:test`, `:race`, `:lint`, `:build`, `:gegenproben`, `:groessenmarke`) tragen in einem Arbeitsbaum außer dem Hauptbaum ein Präfix aus dem Verzeichnisnamen, der Hauptbaum behält die Namen; (b) die schweren Ziele der Gate-Kette (`gates`, `gegenproben`, `test-race`, `groessenmarke` und jeder `docker build` der Fragmente) laufen unter einer Sperre über die Arbeitsbäume, ein zweiter Lauf wartet; (c) der Stempel `harness/gegenproben.stempel` wird nur dort geschrieben, wo die Regel es zulässt (Hauptbaum oder ausdrücklich angefordert), die Regel „erst nach dem letzten Eingabe-Commit“ bleibt; (d) `harness/README.md` und der Vertrag nennen es, die Übergangsregeln entfallen.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Jede Produktfunktion und jede Änderung an `src/`: der Slice ist Werkzeug; sein Nutzen ist Parallelität der Entwicklung, nicht ein Verhalten des Recorders.
- Ein Ersatz für `welle-dev-container` ([welle-dev-container](../welle-dev-container.md)): Dort trennt ein eigener Docker-Daemon je Arbeitscontainer die Läufe vollständig, mit Limits und Zugang; hier teilen sich die Arbeitsbäume weiter einen Daemon und einen Host, und die Sperre serialisiert. Die Welle bleibt unverändert nach `welle-record-replay`.
- Die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Lauf über 30 s) und die Fristen der Gegenproben (120 s je Fall): Der Slice hält die Wartezeit auf die Sperre aus der Messung der Fristen heraus (Zusage in §3) und ändert weder Frist noch Trigger; beides entscheidet der User.
- Ein Gate lockern (`AGENTS.md` §3.6): Keine Prüfung entfällt oder wird seltener gefahren; die Sperre macht Läufe langsamer, nie weniger.
- Eine Lösung für Konflikte im Inhalt der Zweige (Merge von `spec/spezifikation.md`, Roadmap): Sie liegen bei der Reihenfolge der Linien und beim Integrationsschritt der Übergangsregeln (Kennungsbereiche, Nr. 4 bis 6), nicht in einem Werkzeug.
- Änderungen an `AGENTS.md`, `.claude/commands/`, `.claude/agents/` und `.harness/skills/`: Sie gehen als Diff-Vorschlag an den User (Betriebsregel der Welle).
- Ein Parallelbetrieb der Gate-Kette selbst (zwei Slices der Linie C zugleich): Sie berühren dieselben Dateien und laufen nacheinander (Übergangsregeln Nr. 7); der Slice lockert das nicht.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Image-Tags je Arbeitsbaum: ein Lauf aus einem zweiten Arbeitsbaum überschreibt keinen Tag des Hauptbaums; der Hauptbaum behält die Namen; belegt durch Gegenprobe (zwei Verzeichnisse, ein Daemon) und eine benannte Ausgabe.
- [ ] Sperre über die Arbeitsbäume in den schweren Zielen: ein zweiter Lauf wartet, der Lauf im Lauf (`make gates` ruft Fragmente) blockiert sich nicht selbst, eine abgebrochene Sperre bleibt nicht stehen; belegt durch Gegenprobe.
- [ ] Stempel-Regel: ein Lauf im Arbeitsbaum außer dem Hauptbaum schreibt den Stempel nicht, solange es nicht verlangt wird; die Lesart „unverändert seit dem letzten grünen Lauf“ bleibt; belegt durch Gegenprobe; `harness/README.md` (Sensors, Werkzeuge, *Nicht behauptet*) und der Vertrag nennen die Wirkung und ihre Grenzen.
- [ ] `make gates` grün; `make gegenproben` grün (Eingabe geändert, Lauf ausgeführt, Stempel erst nach dem letzten Eingabe-Commit).
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `harness/README.md` und der Vertrag; die Übergangsregeln Nr. 1 und 2 in `planning/README.md` werden vom Integrationsschritt nach dem Merge durch einen Zeiger auf den Vertrag ersetzt (Aufgabe des Planners, nicht des Zweigs).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der Welle-Closure.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/mk/go-gates.mk`, `harness/mk/lint-profil.mk`, `harness/mk/groessenmarke.mk`, `harness/mk/gegenproben.mk` | update | die festen Tags `odata-recorder:<stufe>` werden aus einem Präfix gebildet; Sperre um die `docker build`-Rezepte |
| `Makefile`, `harness/mk/` (neues Fragment oder Hilfsskript unter `tools/harness/`) | update / neu | Bildung des Präfixes aus dem Verzeichnisnamen (Hauptbaum: kein Präfix), Sperre mit Schutz gegen Selbstblockade beim Aufruf aus `gates` |
| `tools/harness/gegenproben-eingabe.sh`, `harness/mk/gegenproben.mk` | update | Stempel-Regel; die Tags der Stufe `gegenproben` |
| `tools/gegenproben/` | update | Fälle zu den Zusagen (zwei Verzeichnisse, ein Daemon; Sperre; Stempel), Treiber auf dem Host wie bei `groessenmarke` und `schema.mk` ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); die Wartezeit auf die Sperre liegt vor dem Zeitpunkt, ab dem die Frist je Fall läuft |
| `Dockerfile` | update, falls nötig | nur wenn ein Tag im `Dockerfile` selbst steht; sonst unverändert |
| `harness/vertraege/` | neu oder update (Architect) | Vertrag der drei Zusagen mit Randformen; Kennung aus `SPEC-200` bis `SPEC-229` |
| `harness/README.md` | update | Sensors/Werkzeuge/*Nicht behauptet*; die Zusage „zwei Arbeitsbäume stören sich nicht“ nennt ihre Grenze (ein Daemon, ein Host) |
| der Plan dieses Slice (`slice-worktree-parallel.md` im Verzeichnis seines Zustands) | update | §3 bis §7 nach `AGENTS.md` §3.9 |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Aufträge 1 bis 6, Tabelle *Zusage · Fänger*, Raster, Nachzüge, erwartete Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Closure (Planner). Worktree: `/Development/odata-recorder-wt-werkzeug`, Zweig `slice-worktree-parallel` (Linie C).

**Betriebsregeln für alle Rollen:** Wortlaut und Gründe in den [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) und in [welle-sensoren](../welle-sensoren.md) §4. Dieser Slice ändert die Gate-Kette selbst: Seine Gate-Läufe laufen unter der Sperre der Übergangsregeln (`flock /tmp/odata-recorder-heavy.lock`), bis der Slice sie ersetzt; sein erster Lauf mit der neuen Sperre läuft ohne die äußere, damit er sich nicht selbst blockiert, und der Bericht nennt das.

**Aufträge an den Architect** (Entscheidungen, die kein Plan trifft; der Vertrag steht in `harness/vertraege/`, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), `AGENTS.md` §3.8; eine Entscheidung über Host-Abhängigkeiten oder die Bindung eines Gates ist eine ADR aus dem Bereich 0030 bis 0039 der Übergangsregeln):

1. **Wie der Arbeitsbaum erkannt wird** und woraus das Präfix entsteht: Verzeichnisname, `git rev-parse --git-common-dir` gegen `--git-dir`, eine Umgebungsvariable zum Übersteuern; Verhalten in einem Klon ohne Worktree, bei einem Verzeichnisnamen mit Zeichen, die ein Docker-Tag nicht trägt (Großbuchstaben, Länge), bei zwei Verzeichnissen mit demselben Namen. Der Hauptbaum behält `odata-recorder:<stufe>`.
2. **Die Sperre**: Mittel (`flock` aus util-linux ist eine Host-Abhängigkeit über Docker, GNU `make`, `bash` und `git` hinaus — `AGENTS.md` §3.1; die Alternative ist eine Sperre in `bash` ohne neues Programm), Ort der Sperrdatei (gemeinsam über Arbeitsbäume, Pfad des Gate-Fragments), Wiedereintritt (`make gates` ruft schwere Ziele; ein Zähler oder eine Umgebungsvariable), Verhalten bei Abbruch, bei `make -j`, bei Zeitlimit des Aufrufers, und ob ein Lauf auf die Sperre wartet oder nach einer Frist abbricht (und mit welchem Fehlertext). Die Wartezeit zählt nicht in die Fristen der Gegenproben.
3. **Die Stempel-Regel**: Wer schreibt, wer liest; Verhalten im Arbeitsbaum außer dem Hauptbaum (nicht schreiben, schreiben auf Anforderung, in eine Datei außerhalb des Baums); wie der Integrationsschritt den Stempel nach dem Merge setzt; Verhältnis zu `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`.
4. **Die Grenze der Zusage**: Was „stören sich nicht“ trägt und was nicht — der geteilte Daemon (Build-Cache, Speicher, Platte), feste Namen außerhalb der Gate-Stufen (zum Beispiel `docker run --name`), Läufe der Gegenproben, die selbst Docker rufen, Tags in den Treibern von `tools/gegenproben/`. Eine Stelle mit festem Namen ist hier beseitigt oder als benannter Rest mit Folge im Vertrag genannt (`AGENTS.md` §3.15).
5. **Tabelle *Zusage · Fänger*** (unten), **Raster** (§6), **Nachzüge** nach `AGENTS.md` §3.9 (`SPEC-102`, `SPEC-122`, `SPEC-128`, `harness/README.md`, Gegenproben-Eingaben) und ob die Sperre eine ADR braucht.
6. **Erwartete Größe** (Zeilen Make und Skript, Fälle der Gegenprobe; die Sperre und ihr Fänger sind nebenläufig und zählen × 85) für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10; Kandidaten der Zusagen: Tags je Arbeitsbaum, Hauptbaum unverändert, Sperre wartet, Sperre ohne Selbstblockade, Sperre nach Abbruch frei, Wartezeit nicht in der Frist, Stempel im Arbeitsbaum nicht geschrieben, Stempel im Hauptbaum unverändert geschrieben):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 5) | — | — |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-groessenmarke-werkzeug` liegt in `done/` (er berührt `Makefile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und den Stempel; solange er in Arbeit ist, hat dieser Slice keine feste Basis), und kein anderer Slice der Gate-Kette liegt in `in-progress/` (Übergangsregeln Nr. 7). Ansonsten frei: kein produktiver Slice und keine produktive Welle wartet auf ihn, und er wartet auf keinen. Reihenfolge: parallel zu den Linien A und B, niedrige Priorität, kein Blocker; ein Slice von `welle-sensoren` Teil 1 oder 2, der zuerst startet, geht vor. Die Architect-Aufträge 1 bis 6 sind entschieden, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Ein Schnitt ist vorab benannt: die Stempel-Regel (Zusage 3) teilt keinen Mechanismus mit Tags und Sperre und lässt sich als eigener Teil zuletzt schreiben.
- `in-progress` → `open` (blockiert): der Architect findet, dass die Sperre eine Host-Abhängigkeit braucht, die der User nicht freigibt (Auftrag 2); der Slice geht zurück, bis die Entscheidung des Users vorliegt.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

- Die DoD aus §2 ist vollständig, `make gates` und `make gegenproben` sind grün (im Hauptbaum, nach dem Merge, auf ruhigem Host), der Verifier hat bestanden.
- Der Lauf aus einem zweiten Arbeitsbaum gegen einen laufenden Gate-Lauf des Hauptbaums ist beobachtet: kein Tag überschrieben, der zweite Lauf hat gewartet; die Beobachtung steht in §7 mit Host-Zustand.
- Der Lerneintrag nennt, was der Parallelbetrieb tatsächlich gestört hat (Übergangsregeln gegen Beobachtung) und was davon der Slice trägt.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* füllt der Architect vor dem ersten Code-Commit; Operationen sind *Präfix bilden*, *Sperre nehmen*, *Sperre freigeben*, *Stempel schreiben*; Argumentarten sind fehlender, leerer, unbekannter und fremder Wert (Verzeichnisname mit Sonderzeichen, Sperrdatei nicht beschreibbar, Stempel fehlt), Abbruch und Wiederholung. Jede Zelle nennt ihren Ort im Vertrag oder „gilt nicht“ mit Grund. **Ausgang:** offen, Adresse Architect.

- Der Slice ändert `Makefile`, `harness/mk/*`, `Dockerfile`, `harness/README.md` und `tools/gegenproben/` und kollidiert damit mit `slice-groessenmarke-werkzeug`, bis dessen Fix-Runde in `main` liegt — **Ausgang:** entfallen mit dem Start-Trigger (`slice-groessenmarke-werkzeug` in `done/`); bis dahin startet er nicht.
- Er kollidiert mit den Slices von `welle-sensoren` und `welle-dev-container` (dieselben Dateien) — **Ausgang:** weiter offen, geregelt durch die Übergangsregeln Nr. 7 (einer zugleich, wer zuerst startet); sein Nachweis ist der Merge ohne Konflikt im Inhalt.
- Eine Host-Abhängigkeit (`flock`) widerspricht `AGENTS.md` §3.1 (Host braucht nur Docker und GNU `make`, dazu `bash` und `git`) — **Ausgang:** offen, Adresse Architect (Auftrag 2), danach User, falls eine Abhängigkeit entsteht.
- Ein fester Name außerhalb der Tags (ein Container, ein Netz, ein Volume) stört trotz Präfix — **Ausgang:** offen, Adresse Architect (Auftrag 4); wird er nicht beseitigt, steht er im Vertrag als benannter Rest mit Folge.
- Die Sperre macht Läufe langsamer, und die Fristen der Gegenproben (120 s je Fall) hängen weiter an der Last des geteilten Hosts (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`) — **Ausgang:** weiter offen → Register; der Slice misst die Wartezeit und entscheidet die Fristen nicht.
- Die Übergangsregeln und der Vertrag sagen Verschiedenes, sobald der Slice liegt — **Ausgang:** offen, Adresse Planner: Der Integrationsschritt ersetzt Nr. 1 und 2 der Übergangsregeln im selben Merge durch einen Zeiger (`AGENTS.md` §3.9).

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

*(Wird bei der Closure gefüllt.)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt wird die Sub-Area `*` (Kürzel `REPO`, Greenfield, `harness/conventions.md` §Modus-Deklaration); `harness/`, `tools/` und `Makefile` gehören zu ihr, eine feinere Sub-Area ist nicht deklariert und nicht nötig.

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Treffer: `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (1×, offen) für die Stempel-Regel; `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, offen) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, offen) für die Sperre und die Fristen; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, verkörpert in `AGENTS.md` §3.15) für die Zusage „stören sich nicht“, die ihre Grenze nennen muss; `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×) für jeden benannten Rest; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, offen) für die Betriebsregel „nur `Edit` und `Write`“. Keiner erreicht mit diesem Slice 3× neu.

**Modus-Begründungsblock:** alle berührten Sub-Areas GF.
