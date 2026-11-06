# Slice slice-query-interpreter: Query-Interpreter mit Filter, Sortierung, Paging und Projektion

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-08`, `LH-FA-11`, `LH-QA-02`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-08.a` (Pipeline) · `LH-FA-08.b` · `LH-QA-02.a` (stabile Sortierung) · `SPEC-044` (Property-Maske) · `SPEC-022`, `SPEC-024` · `ARC-001`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-query-interpreter, Worktree /Development/odata-recorder-wt-query-interpreter)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Interpreter führt einen geprüften Query-AST auf einer Menge typisierter Entities aus: Filter, stabile Sortierung (bei Gleichstand nach Entity-Key, ohne `$orderby` nach Key), Zählung vor dem Paging, `$skip` und `$top` (`$top=0` ergibt eine leere Liste) und Projektion mit Property-Maske, und liefert Elemente und Count für die Serialisierung.

**Übernimmt (Hinweis, keine Pflicht; Sendung von `slice-bootstrap-m2`, `AGENTS.md` §3.13):** Längen-Differenz zwischen Filter und Schlüssel (`SPEC-161`, `SPEC-164`; 1641 Abweichungen in 167634 Läufen): alle sind `literal` im Filter gegen `zu_lang` im Schlüssel, erklärt durch die Regel „der Rest hinter dem schließenden `'` zählt nicht zur Länge“ (`SPEC-164` Nr. 4 a); entschieden und vertraglich getragen, kein Fehler. Eine Angleichung wäre eine Änderung des Vertrags des Lesers; wer sie will, entscheidet sie in diesem Slice, dessen Lauf beide Pfade verwendet. Ohne Beschluss bleibt die Regel.

**Übernimmt (Sendung von `slice-edm-ausdruecke`, `AGENTS.md` §3.13):** (a) der Interpreter ruft `model.PruefeFilter` mit der Property-Maske vor der Auswertung des Filters und bildet den Fehler über die Abbildung ab (`SPEC-167` Nr. 6); (b) `$orderby` nach einer Property, für die `model.HatOrdnung` falsch ist (`Edm.Guid`), ist `501` mit `ODR-E6001` (`SPEC-166` Nr. 3 f); (c) `desc` kehrt die Ordnung von `model.OrdneWerte` um, `null` steht dann zuletzt (`SPEC-166` Nr. 6). Die Auswertung der Wahrheitswerte, die Vergleiche und die Ordnung selbst liefert `slice-edm-ausdruecke`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Coverage-Prüfung vor der Ausführung — `slice-mock-lesen`; der Interpreter bekommt einen Bestand und fragt nicht, wie vollständig er ist.
- Lesen aus dem Repository und Serialisierung — `slice-mock-lesen` und `slice-odata-json-schreiben`; der Interpreter hat keinen Port.
- `$expand` und Navigationen — `welle-expand-batch`.

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

- [ ] Filter und Count: die Zählung nach dem Filter und vor dem Paging, `@odata.count` und `/$count` mit derselben Zahl.
- [ ] Sortierung und Paging: stabil, Gleichstand nach Entity-Key, `$top=0` leer, `$skip` über das Ende leer (`LH-FA-08`, Boundary).
- [ ] Projektion mit Property-Maske: `$select` einer nicht beobachteten Property ergibt `ODR-E6003`; `null` und fehlend bleiben unterscheidbar (`LH-FA-18`, Teilprojektion).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/model/` | update | Ausführung des geprüften AST auf einer Menge typisierter Entities |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-08.a`: kombinierte Abfrage, `$top=0`, Count vor Paging, Gleichstand, Maske |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Reine Funktionen auf einer Liste im Speicher; wie der Interpreter bei 10.000 Entities pro Workspace bleibt (`LH-QA-01.a`), ist eine Randform des Vertrags, keine Entscheidung im Code.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-edm-ausdruecke` liegt in `done/` (eingetreten, Beleg [slice-edm-ausdruecke](../done/slice-edm-ausdruecke.md)); der Start als Teil 1 von `welle-mock` folgt dem Trigger aus deren §2 (Linie B). Zusammenhang mit der Regel „höchstens eine Linie von Teil 1 zugleich“ (`welle-mock` §4): `slice-odata-json-schreiben` aus Teil 1 läuft bereits in einem eigenen Worktree; die Regel bleibt im Wortlaut der Welle stehen und gilt hier als Hinweis, denn der User hat höchstens drei Worktrees zugleich entschieden. Beide Slices brauchen nur das Modell (der Interpreter zusätzlich `slice-edm-ausdruecke`, in `done/`) und teilen keinen Code. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): der Ort der Entscheidung über negative oder nicht ganzzahlige Werte von `$top` und `$skip` (Parser oder Interpreter) ist nicht entschieden (siehe §6).

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Metrik wie oben; Faktoren, Formel und Schnittfrage stehen allein in `SPEC-128` Nr. 1, hier steht keine Zahl davon. Der Block `groessenmarke` steht erst im Plan, wenn der Architect `pfad`, `produktcode`, `fakes` liefert und der Planner `basis`, `schaetzung`, `marke` mit `make groessenmarke` setzt, in eigenem Commit vor dem ersten Code-Commit im Slice-Zweig (damit `make groessenmarke` auf `main` bis zum Merge grün bleibt). Trifft die Schnittfrage zu, wird vor dem Code geschnitten; sie stellt der Architect mit der Tabelle (§6).

**Aufträge an den Architect** (vor dem ersten Code-Commit, `AGENTS.md` §3.12): (1) Vertrag der Ausführung in der Spezifikation mit Randformen, Argumenten, Fehlern und der Eingabe (geprüfter Query-AST, Menge typisierter Entities, Property-Maske); die Sendung (a) bis (c) aus §1 ist Teil des Vertrags (`SPEC-166` Nr. 3 f und 6, `SPEC-167` Nr. 6, `SPEC-125` Nr. 10 und 11). (2) Die Entscheidung über negative und nicht ganzzahlige Werte von `$top` und `$skip` samt Ort (§6); fällt sie dem Parser zu, ist `slice-odata-url-pfad` kein Nehmer (liegt in `done/`), und der Architect gibt die Änderung als Rückgabe an den Planner (`AGENTS.md` §3.13). (3) Tabelle *Zusage · Fänger* in §3. (4) Raster *Operation × Argumentart* in §6, darin die offenen Risiken (gemischte Sortierschlüssel, Teilprojektion). (5) Erwartete Größe für die Marke. (6) Schnittfrage nach `SPEC-128` Nr. 1. (7) Kennungen: `slice-odata-json-schreiben` läuft zugleich in Linie B; zur Vermeidung stiller Kollisionen vergibt dieser Slice aus dem hinteren Teil des Bereichs: `SPEC-190` bis `SPEC-199`, `ARC-035` bis `ARC-039`, ADR `0025` bis `0029`, `CO-030` bis `CO-039`; `slice-odata-json-schreiben` behält `SPEC-176` bis `SPEC-189`, `ARC-030` bis `ARC-034`, ADR `0020` bis `0024`, `CO-020` bis `CO-029`. Ein erschöpfter Teil geht als Frage an den Planner.

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7): Der Slice berührt laut §3 keine Datei der Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel). Muss der Architect oder der Implementer eine ändern (etwa die Regeln von `.a-check.yml` für ein neues Paket), hält er an und gibt es als Rückgabe an den Planner zurück (§6, Ausgang *offen*, `AGENTS.md` §3.13); der Slice ändert sie nicht selbst.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Negative oder nicht ganzzahlige Werte von `$top` und `$skip` (`400`): der Ort der Entscheidung (Parser oder Interpreter) ist offen — **Ausgang:** offen; Architect, eine Zelle des Rasters.
- Ordnung von `null` im `$orderby` — **Ausgang:** entfallen als Lücke: entschieden in `SPEC-166` Nr. 6 (`null` zuerst aufsteigend, zuletzt absteigend). Gemischte Sortierschlüssel — **Ausgang:** offen; Architect.
- Laufzeit bei 10.000 Entities pro Workspace (`LH-QA-01.a`): dieser Slice sagt Korrektheit zu, nicht Latenz — **Ausgang:** weiter offen; die Messung trägt der Closure-Trigger von `welle-betrieb` (§3 dort, Lasttest).
- Teilprojektionen: `$select` einer Property, die nur teilweise beobachtet ist (`SPEC-044`) — **Ausgang:** offen; Raster *Operation × Argumentart*.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start am 2026-11-06 erneut; der Zähler folgt aus den Dateien unter `../observations/BEO-REPO/<slug>/evidence/`, hier steht keine Zahl. Reine Funktionen mit Allaussagen: `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16). Beim Start zusätzlich einschlägig für Fänger und Raster (verkörpert in `AGENTS.md` §3.17 bzw. §3.14): `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (10.000 Entities, Mengen von Sortierschlüsseln), `BEO-REPO/zustand-des-arguments-als-dimension` (fehlend, `null`, leer) und `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`. Offen: `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (Stand: geplant in `slice-guard-schreibwege`); die Betriebsregel „nur `Edit` und `Write`“ gilt für jede Rolle dieses Slice (Übergangsregeln Nr. 8).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
