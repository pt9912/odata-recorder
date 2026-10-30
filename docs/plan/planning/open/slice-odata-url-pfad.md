# Slice slice-odata-url-pfad: Parser der OData-URL, Teil Ressourcenpfad, Key-Syntax und Systemoptionen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-08`, `LH-FA-07`, `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-08.a` · `LH-FA-08.b` (URL-Key-Syntax) · `LH-FA-07.b` (Adressierung per Key) · `SPEC-049` · `SPEC-022` · `ARC-004`, `ARC-011` · `architecture.md` §1 (Codec-Zugriff), §4.4.
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

**Ziel:** Der Parser im Paket `src/adapters/driven/odata/urlparser` (hinter einem Port der Codec-Gruppe) überführt Pfad und `raw_query` einer OData-URL in einen technologieunabhängigen Query-AST des Domain-Modells: Ressource (EntitySet, Key einfach oder zusammengesetzt mit typisierten Literalen, `$metadata`, `/$count`) und die Systemoptionen `$select`, `$orderby`, `$top`, `$skip`, `$count`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Grammatik von `$filter` — `slice-odata-url-filter`; bis dahin ergibt die Option `$filter` den Fehler `ODR-E6001`, nie eine ungefilterte Antwort (`LH-FA-08`, Negative).
- Die Typverträglichkeit der Literale und die Ausdruckssemantik — `slice-edm-ausdruecke`; der Parser liest die Syntax eines Literals, entscheidet aber nicht, ob es zum Typ der Property passt.
- `$expand`, `$apply`, `$search`, `$batch`, Actions und Functions — nicht im Muss-Umfang (`SPEC-049`); sie ergeben `501`, der Erweiterungspunkt liegt bei `welle-expand-batch`.

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

- [ ] Ressourcenpfad: EntitySet, einfacher und benannter zusammengesetzter Key mit typisierten Literalen und Escaping, `$metadata`, `/$count`; ein ungültiger oder unvollständiger Key ergibt `400`.
- [ ] Systemoptionen: `$select`, `$orderby` (mehrere Schlüssel, Richtung), `$top`, `$skip`, `$count` mit ihren Randformen (`$top=0` gültig, negativ oder nicht ganzzahlig `400`, doppelte und unbekannte Option, `$filter` bis `slice-odata-url-filter` `ODR-E6001`).
- [ ] Query-AST im Domain-Modell ohne Typ einer Parser-Bibliothek; der Parser liegt hinter einem Port der Codec-Gruppe und wird allein vom Composition Root verdrahtet (`make a-check`).
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
| `src/hexagon/model/` | update | Query-AST: Ressource, Key, Optionen |
| `src/hexagon/ports/outbound/` | update | Port der OData-URL in der Codec-Gruppe |
| `src/adapters/driven/odata/urlparser/` | neu | Parser von Pfad und `raw_query` |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-08` und `LH-FA-07.b` je Option und je Key-Form |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Parser liest fremde Bytes (eine URL, die ein Client schreibt): die Formel der Marke nimmt Faktor 65 und die Produktcode-Größe des Architect als festen Anteil (`AGENTS.md` §3.14); die Allaussagen („jede Option“, „jede Key-Form“) prüft ein Test als Schleife über die Menge (`AGENTS.md` §3.15).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/` (Trigger der Welle, M2 erreicht). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt; sie zu schätzen, bevor der Vertrag steht, wäre die freie Schätzung, die §3.14 ersetzt.
- `in-progress` → `open` (blockiert — Carveout?): die Randformen der URL (Escapes, Schreibweise der Optionsnamen, Wiederholung einer Option) sind in der Spezifikation nicht entschieden (siehe §6).

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

- Randformen der URL (Escapes im Pfad, Schreibweise der Optionsnamen, doppelte oder leere Option, `$top` mit Vorzeichen) — **Ausgang:** offen; das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit, jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund (`AGENTS.md` §3.16); eine Zelle ohne Ort geht an den Architect.
- Key-Literale: Escaping des Apostrophs, Prozent-Kodierung, Reihenfolge benannter Key-Teile, Name eines Key-Teils, den das Modell nicht kennt (der Parser kennt das Modell nicht) — **Ausgang:** offen; Architect.
- Allaussagen („jede Option“, „jede Key-Form“) — **Ausgang:** offen; der Fänger läuft als Schleife über die Menge, der Fix deckt die Klasse (`AGENTS.md` §3.15, §3.17).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice liest fremde Bytes: `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14, `.claude/agents/implementer.md` und `.harness/skills/reviewer.md`) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (5×, verkörpert in `AGENTS.md` §3.17) gelten für „jede Option“ und „jede Key-Form“; `BEO-REPO/randform-erst-im-review-entschieden` (20×, verkörpert in `AGENTS.md` §3.16) für das Raster. Offen: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
