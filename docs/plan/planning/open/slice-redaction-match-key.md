# Slice slice-redaction-match-key: Redaction von Headern und Query, Match-Key, Replayfähigkeit

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-05`, `LH-FA-14`, `LH-QA-02`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-05.a` (Match-Key) · `LH-FA-14.a` (Redaction) · `SPEC-045` (Fähigkeitsmatrix) · `SPEC-060`, `SPEC-061`, `SPEC-063`, `SPEC-081` bis `SPEC-085` · `SPEC-001` (`key_redacted`) · `ARC-001`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Domain-Modell redigiert Header, `raw_query` und Query-Werte in `Location` und `Content-Location` mit dem festen Marker, bildet den Match-Key als SHA-256 über die eindeutig kodierte Tupelfolge aus der redigierten Darstellung, und entscheidet nach `SPEC-045`, ob ein Exchange replayfähig ist.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Redaction der JSON-Bodies im Strom — `slice-streaming-redaction`; hier entsteht die Syntax der Pfadmuster (`$..name`) und der Marker, nicht der Strom.
- Anwendung beim Record und beim Replay — `slice-record-service` und `slice-replay-service` rufen dieselbe Funktion; hier sind es reine Funktionen ohne I/O.
- Semantische Gleichsetzung von `$filter`-Ausdrücken — `LH-FA-05.a` nimmt sie aus dem Muss-Umfang.

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

- [ ] Redaction von Headern (Default-Liste plus `REDACT_HEADERS`), `raw_query` und Query-Werten in `Location`/`Content-Location` (`REDACT_QUERY_PARAMS`); fester Marker `[REDACTED]`; Parser der Pfadmuster aus `REDACT_JSON_PATHS`; `key_redacted`.
- [ ] Match-Key: Tupelkodierung ohne Mehrdeutigkeit, sortierte Multimenge der dekodierten Query-Paare (`+` bleibt `+`, ungültige Kodierung `literal` oder `reject`), Vary-Header, Body-Hash in den Modi `bytes` und `json-canonical`, volatile Header ausgenommen.
- [ ] Replayfähigkeit nach `SPEC-045` als Funktion über Capture-Status, Body-Redaction und Integrität (`REPLAY_REDACTED_BODIES`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/model/` | update | Redaction, Match-Key, Replayfähigkeit als reine Funktionen |
| `spec/spezifikation.md` | update | Tupelkodierung und `normalized_path` als Format, Randformen von Query und Pfad |
| Tests des Pakets `model` | neu | Tabellentests mit Mutationen je Zusage nach `AGENTS.md` §3.10, darunter die Beispiele aus `LH-FA-05.a` und `LH-FA-14.a` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

- Der Match-Key ist persistiert (`exchanges.match_key`, Index `(session_id, match_key)`); seine Kodierung ist damit Teil des Aufzeichnungsformats (`RECORDING_FORMAT_VERSION`) und wird in der Spezifikation fixiert, bevor Code existiert.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` ist eröffnet; kein Vorgänger-Slice. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 1000 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 700 Zeilen in derselben Metrik.
- `in-progress` → `open` (blockiert): die Tupelkodierung lässt sich ohne Entscheidung des Architect nicht festlegen (siehe §6).

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

- `normalized_path` ist nicht definiert: `LH-FA-05.a` sagt, Segmentgrenzen, Groß-/Kleinschreibung und OData-Literale blieben erhalten, nennt aber nicht, was normalisiert wird (Doppelschrägstrich, abschließender Schrägstrich, Prozent-Kodierung im Pfad, Service-Root als Präfix) — **Ausgang:** offen; geht vor dem Code an den Architect.
- Query-Randformen ohne Ort in der Spezifikation: Paar ohne `=`, leerer Schlüssel, leerer Wert, nur `?`, Paare ohne Trennzeichen, gleiche Schlüssel mit verschiedenen Werten in verschiedener Reihenfolge — **Ausgang:** offen; Raster *Operation × Argumentart* in §6.
- Redaction von Headern mit mehreren Werten und von mehrfach vorkommenden Headern; Groß-/Kleinschreibung der Namen — **Ausgang:** offen; Raster-Zelle.
- Ein `Vary`-Header, der im Request fehlt, gegen einen leer gesetzten — **Ausgang:** offen; muss im Key verschieden sein oder nicht, die Spezifikation schweigt.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (2×); der Match-Key ist ein Vertrag mit vielen Randformen, deshalb steht das Raster vor dem Code (`AGENTS.md` §3.16).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
