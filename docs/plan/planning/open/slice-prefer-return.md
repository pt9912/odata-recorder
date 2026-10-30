# Slice slice-prefer-return: Prefer return=minimal und return=representation für mutierende Requests

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-etag](../welle-etag.md).

**Bezug:** `LH-FA-07`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-07.b` (CRUD-Vertragsmatrix: `Prefer` für POST, PATCH, PUT; ohne `Prefer` gilt das Muss-Verhalten; ein nicht unterstützter Wert ergibt keinen Fehler) · `SPEC-049` (Zeile Header, Spalte Soll: `Prefer: return=…`) · `SPEC-006` · `ARC-002`, `ARC-004`.
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

**Ziel:** Der Mock wertet den Header `Prefer` bei POST, PATCH und PUT aus: `return=minimal` ergibt eine Antwort ohne Entity im Body (Status und Header nach dem Vertrag des Architect), `return=representation` eine Antwort mit der Entity im Body; die Antwort nennt den angewandten Wert im Header `Preference-Applied`, soweit der Vertrag das verlangt. Ohne `Prefer` und bei jedem anderen Wert gilt das Muss-Verhalten ohne Fehler (POST `201` mit Entity, PATCH und PUT `204`). `Prefer` ändert den Zustand des Workspace nie: dieselbe Mutation ergibt mit und ohne Header denselben Zustand, und ein fehlgeschlagener Vorgang antwortet mit demselben Fehlerobjekt.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- ETag, `If-Match`, `If-None-Match` und `412` — `slice-etag-versionen` und `slice-if-match`; die beiden Äste sind unabhängig (`welle-etag` §5). Berührung ohne Kante: trägt eine Antwort mit Entity den ETag (aus `slice-etag-versionen`), prüft der Test des später liegenden Slice den Fall „`return=representation` mit ETag“.
- `Prefer`-Werte außer `return=minimal` und `return=representation` (etwa `odata.maxpagesize`, `respond-async`, `odata.include-annotations`) — Out-of-Scope der Welle (`welle-etag` §6); sie werden ohne Fehler ignoriert (`LH-FA-07.b`).
- `Prefer` bei GET und DELETE — GET liefert immer die Entity, DELETE kennt keinen Body; der Architect nennt den Fall im Raster („gilt nicht“ mit Grund oder ignoriert).
- `Prefer` im Record und Replay — der Header bleibt Teil des Exchange und geht als Vary-Header in den Match-Key (`SPEC-061`); der Mock-Pfad ist der einzige, der ihn auswertet.

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

- [ ] `return=minimal` und `return=representation` für POST, PATCH und PUT: Status, Body, `Location` und `OData-EntityId`, `Preference-Applied` nach dem Vertrag; ohne `Prefer` unverändert das Muss-Verhalten.
- [ ] Randformen des Headers (Groß- und Kleinschreibung, Leerraum, mehrere Präferenzen in einem Header, mehrere Header, widersprüchliche Werte, leerer oder unbekannter Wert): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; ein nicht unterstützter Wert ergibt keinen Fehler.
- [ ] Unveränderter Zustand: mit und ohne `Prefer` derselbe Zustand des Workspace und dieselbe Revision; bei fehlgeschlagener Mutation dasselbe Fehlerobjekt.
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
| `src/hexagon/services/mock/` | update | Wahl der Antwort nach der Präferenz |
| `src/hexagon/model/` | update | Typ der Präferenz im Request und der angewandten Präferenz im Ergebnis |
| `src/adapters/driving/httpodata/` | update | Lesen von `Prefer`, Schreiben von `Preference-Applied` und Status |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-07.b` je Methode und je Wert; Generator über die Klasse der Header-Formen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis in der Spezifikation, Neues unter einer künftigen Nummer, die der Architect vergibt):

1. Antwort je Methode und Präferenz: Status (`201`, `200`, `204`), Body, `Location`, `OData-EntityId`, `Preference-Applied`, Header `Vary`.
2. Grammatik des Headers: mehrere Präferenzen, Parameter, Groß- und Kleinschreibung, Leerraum, mehrere Header, widersprüchliche Werte (erster, letzter, keiner); was „nicht unterstützt“ heißt (ignorieren ohne Fehler, `LH-FA-07.b`).
3. `Prefer` bei GET, DELETE und bei fehlgeschlagener Mutation.
4. Wie die Entity der Antwort entsteht (aus dem geschriebenen Zustand, nicht aus dem Body des Requests) und ob der Schreiber aus `slice-odata-json-schreiben` dafür unverändert genügt.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect mit dem Vertrag, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (etwa Header-Werte mit NBSP) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Trigger der Welle). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-mock-schreiben` (Statuscodes und Antworten der Mutationen, die der Slice umschaltet), `slice-odata-json-schreiben` (die Entity im Body der Antwort), `slice-http-adapter` (der Header erreicht den Use Case als Teil des Requests). Früheste Lage für Code: `slice-mock-schreiben` und `slice-odata-json-schreiben` in `done/`; der Slice kann **nicht** vor ihnen beginnen, weil `MockUseCase` mit Mutation nicht existiert.
- **Binary und Handbuch:** das Beispiel der DoD läuft gegen das Binary im Modus `mock` (`slice-bootstrap-m3`); der Slice schließt damit frühestens nach `slice-bootstrap-m3` und `slice-handbuch-meldungskatalog`. Ein Start vor dem Trigger der Welle ist eine Entscheidung des Users, nicht dieses Plans.
- **Was der Slice nicht braucht:** `slice-etag-versionen` und `slice-if-match` (unabhängiger Ast, `welle-etag` §5), `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-expand-batch`, `welle-odata-v2`.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, die Produktcode-Erwartung des Architect ist ein fester Anteil. Der Header ist kurz; ob der Slice als Leser fremder Bytes gilt, entscheidet die Grammatik, die der Architect festlegt (Auftrag 2), und ändert die Formel nicht.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): Grammatik des Headers oder Status je Methode sind nicht entschieden (siehe §6).

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

- Die Menge der Header-Formen ist unbeschränkt (jede Zeichenfolge) — **Ausgang:** offen; Generator über die Klasse oder akzeptiertes Negativ mit Prüfgrenze und Folge im Betrieb (`AGENTS.md` §3.17).
- `Prefer` darf den Zustand nie ändern; ein Fänger vergleicht den Zustand nach derselben Mutation mit und ohne Header, über alle Methoden und alle Werte (Schleife über beide Mengen, `AGENTS.md` §3.15) — **Ausgang:** offen.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Header fehlt | `return=minimal` | `return=representation` | leerer oder unbekannter Wert | mehrere Präferenzen, mehrere Header, widersprüchlich | fehlgeschlagene Mutation |
|---|---|---|---|---|---|---|
| POST | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| PATCH | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| PUT | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| DELETE | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| GET | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für die Fakes von Repository und Schreiber; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
