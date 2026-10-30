# Slice slice-expand-ausfuehrung: Auflösen der Navigation im Query-Interpreter und Ausgabe im OData-JSON

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-FA-11`, `LH-QA-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-16.a` (Executor- und Serializer-Contract-Test für `$expand`) · `LH-FA-11.a` (Coverage vor der Abfrage) · `SPEC-044` (Coverage, Property-Maske) · `SPEC-049` (Zeile GET: Navigation-Expand) · `SPEC-024` (`ODR-E6003`) · `SPEC-126` Nr. 9 (der Leser überspringt eingebettete Navigationen: Grund `navigation`) · `SPEC-002`, `SPEC-003` (Modell, Entity-State) · `ARC-001`, `ARC-004` · `architecture.md` §7 (Erweiterungspunkte).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

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

**Ziel:** Der Query-Interpreter löst die Navigationen des geprüften `$expand`-AST (aus `slice-expand-parser`) gegen die Datenbasis der Navigationen auf, und der Schreiber von OData-JSON stellt die verknüpften Entities eingebettet dar (eine Navigation auf eine Entity als Objekt oder `null`, eine Navigation auf eine Collection als Array in stabiler Reihenfolge nach Key). Eine leere Navigation ergibt die gültige leere Darstellung der unterstützten OData-Version, die Coverage des Navigationsziels wird wie die der Abfrage geprüft, und derselbe Zustand ergibt dieselbe Antwort (`LH-QA-02`). Welche Datenbasis die Beziehungen liefert, entscheidet der Architect vor dem Start (§3, Auftrag 1).

**Übernimmt (Interpreter und Ausgabe):** Navigationen im Query-Interpreter — `slice-query-interpreter` schließt `$expand` und Navigationen aus und nennt `welle-expand-batch` (Sendung von `slice-query-interpreter`, `AGENTS.md` §3.13); eingebettete Entities in der Ausgabe — `slice-odata-json-schreiben` schließt Navigationen und eingebettete Entities aus und nennt `welle-expand-batch` (Sendung von `slice-odata-json-schreiben`). Beide nimmt dieser Slice an.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Parser und Validator von `$expand` — `slice-expand-parser`; dieser Slice liest den geprüften AST und prüft keinen Text.
- Änderungen an Metadatenleser, Import, Leser der Antworten oder Schema, die die Datenbasis verlangen könnte — jeder davon ändert einen Vertrag in `done/` (`SPEC-125`, `SPEC-126`, `SPEC-127`, `SPEC-004`) und ist keine Adresse für eine Sendung (`AGENTS.md` §3.13). Verlangt die Entscheidung des Architect eine solche Änderung, legt der Planner dafür einen Slice vor diesem an (siehe §4 und §6); dieser Slice trägt sie nicht.
- Navigation als Pfadsegment (`Products(1)/Category`), `$ref`, Schreiben von Navigationen (`@odata.bind`) und Deep Insert — kein Gegenstand von `LH-FA-16` (nur `$expand`); der Mock antwortet `501` (`SPEC-049`).
- `$apply`, `$search`, Actions, Functions — bleiben `501` (Lastenheft §5).
- Aufzeichnen und Abspielen von `$expand`-Requests — Record und Replay behandeln sie als HTTP und unabhängig vom Parser (`LH-FA-16.a`, `welle-expand-batch` §6).

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

- [ ] Auflösen im Interpreter: jede Navigation des AST (Single und Collection) gegen die Datenbasis nach Entscheidung des Architect; leere Navigation leer; fehlendes Ziel, Selbstbezug und Tiefe nach dem Raster; stabile Reihenfolge nach Key; Coverage des Ziels vor der Abfrage (`ODR-E6003` bei `STRICT_COVERAGE=true`, sonst nach dem Vertrag).
- [ ] Ausgabe im OData-JSON: eingebettete Entities (Objekt, `null`, Array), `@odata.context` mit dem Fragment der Expansion, Property-Maske der eingebetteten Entities; Rundlauf mit dem Leser (`SPEC-126`).
- [ ] Executor- und Serializer-Contract-Test für `$expand` (`LH-FA-16.a`); ein Test im Modus `mock` über die Abnahme der Welle: verknüpfte Daten sichtbar, leere Navigation leer.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün.
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
| `src/hexagon/` (Paket des Query-Interpreters) | update | Auflösen der Navigationen des AST |
| `src/hexagon/services/mock/` | update | Verdrahtung: Zugriff auf die Datenbasis, Coverage des Ziels |
| `src/adapters/driven/odata/jsoncodec/` | update | Schreiber: eingebettete Entities |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, Datenbasis der Navigationen (Architect) |
| Tests | neu | Fälle von `LH-FA-16` (Happy Path, Boundary) je Navigationsart; Generator über die Klasse der Beziehungen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Auftrag an den Architect** (vor dem Start der Welle; Ergebnis in der Spezifikation, Neues unter einer künftigen Nummer, die der Architect vergibt):

1. **Datenbasis der Navigationen.** Der Import liest eingebettete Navigationen nicht (`SPEC-126` Nr. 9), der Metadatenleser wertet `ReferentialConstraint` nicht aus (Fund `konstrukt`, `SPEC-125` Nr. 7), und der Entity-State (`SPEC-003`) hält Properties, keine Beziehungen. Der Architect nennt die Quelle der Beziehungen (etwa Fremdschlüssel-Properties aus `ReferentialConstraint`, beobachtete Navigationsantworten, eine eigene Ablage) und den Weg dorthin: welche Verträge in `done/` sich dafür ändern und ob dafür ein vorgelagerter Slice nötig ist (Adresse: Planner, `AGENTS.md` §3.13). Nennt er keine Quelle, hat `LH-FA-16` im Mock keine Datenbasis, und die Welle geht an den User zurück.
2. Verhalten bei fehlendem Ziel, Tombstone, Selbstbezug und Tiefe; Reihenfolge der Elemente; abgeleitete Typen im Ziel.
3. Coverage des Ziels: derselbe Maßstab wie für die Abfrage oder ein eigener; Verhalten bei `partial` und `unknown` mit und ohne `STRICT_COVERAGE`.
4. Darstellung: `@odata.context`, `@odata.count` der eingebetteten Collection, Property-Maske der eingebetteten Entities (nur beobachtete Properties), Leerform (`[]`, `null`).
5. Wechselwirkung mit Mutationen: ändert eine Mutation die Beziehung (etwa die Fremdschlüssel-Property per PATCH), und wo steht die Regel.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect mit dem Vertrag, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Trigger der Welle) und `slice-expand-parser` liegt in `done/`. Die Entscheidung des Architect zur Datenbasis steht in der Spezifikation, und ein Slice, den sie verlangt (Reader, Import, Schema), liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-expand-parser` (der geprüfte AST ist die Eingabe), `slice-query-interpreter` (der Interpreter, den dieser Slice erweitert), `slice-odata-json-schreiben` (der Schreiber, der die eingebetteten Entities ausgibt), `slice-mock-lesen` (Verdrahtung in `MockUseCase`, Workspace-Zugriff und Coverage-Prüfung vor der Abfrage). Keiner davon liegt heute in `done/`: der Slice ist **nicht** vor `welle-mock` startbar.
- **Binary und Handbuch:** der Modus `mock` erreicht das Binary mit `slice-bootstrap-m3`, das Kataloge-Gate kommt aus `slice-handbuch-meldungskatalog`; beide liegen mit `welle-mock` in `done/`.
- **Was der Slice nicht braucht:** `slice-mock-schreiben` und `slice-snapshot-use-case` (Lesepfad; ändert die Entscheidung des Architect die Beziehungen über Mutationen, entsteht der Fänger dafür erst hinter `slice-mock-schreiben`, was die Welle ohnehin voraussetzt), `slice-batch-parser`, `slice-batch-changesets`, `welle-etag`, `welle-odata-v2`, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil. Der Slice liest keine fremden Bytes.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt. Der naheliegende Schnitt vor dem Code ist die Datenbasis (Auftrag 1) gegen Auflösen und Ausgabe; ob er ohne geteilten Mechanismus möglich ist, entscheidet der Plan beim Start.
- `in-progress` → `open` (blockiert — Carveout?): die Datenbasis der Navigationen ist nicht entschieden, oder ihre Umsetzung verlangt eine Änderung an Verträgen in `done/`, für die es noch keinen Slice gibt (siehe §6).

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

- Die Datenbasis fehlt (kein Weg zu Beziehungen im heutigen Entity-State, Import und Leser): die Quelle ist Entscheidung des Architect (Auftrag 1), und ihr Weg kann Änderungen an `SPEC-125`, `SPEC-126`, `SPEC-127` oder `SPEC-004` samt Migration nach sich ziehen — **Ausgang:** offen; Adresse: Architect (Entscheidung) und Planner (Slice für Änderungen in `done/`, vor diesem), `AGENTS.md` §3.13.
- Selbstbezug und Zyklen in den Beziehungen (etwa Vorgesetzter) bei Tiefe über eins — **Ausgang:** offen; Architect (Auftrag 2), Raster.
- Eine eingebettete Collection ohne Nachweis der Vollständigkeit darf nicht scheinbar vollständig erscheinen (`SPEC-044`) — **Ausgang:** offen; Architect (Auftrag 3), Fänger mit `STRICT_COVERAGE` an und aus.
- Die Menge der Beziehungsformen ist unbeschränkt (beliebige Modelle) — **Ausgang:** offen; Generator über die Klasse der Modelle (1:1, 1:n, n:m, Selbstbezug, abgeleitete Typen) oder Prüfgrenze als akzeptiertes Negativ (`AGENTS.md` §3.17).
- Fakes der Ports (Repository, Schreiber, Interpreter) halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×; `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Navigation leer | Ziel fehlt oder gelöscht | Ziel nicht nachweislich vollständig | Selbstbezug, Zyklus, Tiefe | Property-Maske, abgeleiteter Typ | Abbruch (Kontext), Wiederholung |
|---|---|---|---|---|---|---|
| Auflösen: Navigation auf eine Entity | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Auflösen: Navigation auf eine Collection | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Ausgabe im OData-JSON | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Lesen einer Entity (Key) und einer Collection mit `$expand` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) gelten für die Fakes von Repository, Schreiber und Interpreter; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
