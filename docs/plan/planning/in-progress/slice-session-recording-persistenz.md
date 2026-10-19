# Slice slice-session-recording-persistenz: Session und Recording: Modell, Ports und SQLite-Repositories

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-persistenz](../welle-persistenz.md).

**Bezug:** `LH-FA-02`, `LH-FA-03`, [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)

**Berührte Spec-Stellen:** `SPEC-001` (Fachliche Recording-Struktur) · `SPEC-105` (Vertrag der Repositories) · `SPEC-048` (Crash-Recovery) · `ARC-001` · `ARC-004` · `ARC-009`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-19.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Session und Exchange sind Domain-Modell, `SessionRepository` und `RecordingRepository` sind Outbound Ports, und der SQLite-Adapter erfüllt sie mit append-only Exchanges und den Statusübergängen einer Session.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Bodies und Blobs — der Exchange trägt nur Verweise; die Ablage liefert [slice-blob-ablage](../open/slice-blob-ablage.md).
- Match-Key-Bildung, Redaction und Capture-Pipeline — Fachlogik der Welle welle-record-replay.
- Aufbewahrung und Startwartung als Use Case — `MaintenanceUseCase` kommt mit den Services; hier gibt es nur die Repository-Operationen.

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

- [ ] Domain-Modell `Session` und `Exchange` mit den Zuständen aus `SPEC-001` und Outbound Ports `SessionRepository`, `RecordingRepository` mit ihren Fehlern; die Transaktionsgrenze über mehrere Ports entsteht mit der ersten Operation über mehrere Repositories in `slice-baseline-persistenz`, hier ist jede Operation atomar.
- [ ] SQLite-Repositories: Session anlegen, Status `open` → `complete`/`incomplete`, offene Sessions finden, Exchanges append-only anhängen und lesen, Session samt Exchanges löschen.
- [ ] Vertragstest-Suite am Port, gefahren gegen den SQLite-Adapter; die Randformen aus §6 sind getestet.
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
| `src/hexagon/model/` | neu | `Session`, `Exchange` und ihre Zustände mit `Validate`; Standardbibliothek ohne I/O |
| `src/hexagon/ports/outbound/`, `outboundtest/` | neu | `SessionRepository`, `RecordingRepository`, ihre Fehler; Vertragstests, die gegen jeden Adapter laufen |
| `src/adapters/driven/sqlite/` | update | Implementierung beider Ports |
| `spec/spezifikation.md` | update | Randformen der Repositories (doppelte `sequence_no`, Anhängen an eine geschlossene Session, Statusübergang rückwärts) |

- Die Vertragstest-Suite liegt am Port und läuft gegen jeden Adapter, der ihn erfüllt; heute gegen SQLite, später gegen ein Fake in den Service-Tests.

**Zusage · Fänger** (`AGENTS.md` §3.15):

| Zusage | Fänger |
|---|---|
| Nr. 1: `Create` prüft Gültigkeit, Kennung, Service (in dieser Reihenfolge) | `Create lehnt Ungültiges, Doppeltes und unbekannten Service ab`, `TestSessionValidate`; Mutation: Service-Zweig entfernt |
| Nr. 2/11: Unbekanntes ist `ErrNichtGefunden` | `Get einer unbekannten Kennung`, `Get und ListBySession bei Unbekanntem und Leerem` |
| Nr. 3: nur `open` → `complete`/`incomplete` | `SetStatus erlaubt nur …`, `TestSessionStatusWechsel`; Mutationen: Übergang immer erlaubt, unbekannte Kennung ohne Fehler |
| Nr. 4/5: Ordnung nach Zeit, leere Liste, `FindOpen` | `List und FindOpen ordnen …`, `List ordnet nach Zeit, nicht nach Text`; Mutationen: ohne `ORDER BY`, Filter umgekehrt, Zeitformat mit variabler Nachkommastelle, Zeit ohne UTC |
| Nr. 6: `Delete` nimmt Exchanges mit, achtet Baseline-Quellen | `Delete nimmt Exchanges mit …`; Mutationen: Exchanges nicht gelöscht, alle Exchanges gelöscht, `ErrReferenziert` entfernt, kein Rollback; Cursors löscht der Fremdschlüssel `CASCADE` (`make schema-check`, Fall „Session löschen nimmt ihre Cursor mit“) |
| Nr. 7: `Append` nur an offene, bekannte Session | `Append prüft Session und Zustand`; Mutationen: Zustandsprüfung entfernt, `incomplete` als offen behandelt |
| Nr. 8: `sequence_no` lückenlos, auch gleichzeitig | `sequence_no ist lückenlos`, `gleichzeitige Anhänge ordnen sich`; Mutation: Sequenzprüfung entfernt |
| Nr. 9: append-only, doppelte Kennung ändert nichts | `doppelte Kennung ändert nichts`; es gibt keine ändernde Operation am Port; die Kennungsprüfung und der Primärschlüssel decken sich, `INSERT OR REPLACE` allein ist äquivalent und keine Mutation |
| Nr. 10: Gültigkeit von Exchange und Session (Pflichtfelder, Aufzählungen, Bodies, Zustände, Jahre 1 bis 9999, `finished_at`, Status, UTF-8) | `ungültige Exchanges werden abgelehnt`, `Zeiten außerhalb der Jahre 1 bis 9999 sind ungültig`, `TestExchangeValidate`, `TestSessionValidate`; Mutationen, je rot im benannten Test: Kennungsprüfung aus, Zeit ohne UTC, Jahresprüfung von `finished_at` entfernt, Status 100 und 599 als Grenze, Aufzählung von Response-Body und Request-Redaction, Blob-Regel je Richtung |
| Nr. 11: `ListBySession` nach `sequence_no` | `ListBySession ordnet nach sequence_no, nicht nach Kennung`; Mutation: Ordnung nach Kennung |
| Nr. 12: Header-Randformen | `Header-Randformen`; Mutation: `nil`-Normalisierung im Schreiber entfernt |
| Nr. 13: abgebrochener Kontext beendet jede Operation ohne Änderung | beide Tests `abgebrochener Kontext beendet jede Operation ohne Änderung` (alle neun Operationen); die Prüfung leistet `database/sql`, eine eigene im Adapter gibt es nicht und ist keine Mutation |
| Jede Operation ist atomar | Mutation „ohne Rollback“ färbt `TestSessionVertrag` rot |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-sqlite-adapter-migration](../done/slice-sqlite-adapter-migration.md) liegt in `done/`.

**Marke überschritten:** Die Summe aus `git diff --numstat` über die Pfade aus §3 liegt bei rund 1430 Zeilen, getragen von den Vertragstests (`outboundtest`, rund 450 Zeilen), den Modelltests und dem Adapter mit Tests. Der Slice bleibt ungeteilt, weil Modell, Ports und Adapter einzeln nichts liefern.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 700 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert): das Modell von `Exchange` verlangt eine Entscheidung der Spezifikation, die fehlt.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, die Vertragstest-Suite läuft gegen SQLite und eine Mutation je Zusage wurde rot gesehen; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen der Repositories (`AGENTS.md` §3.12): leerer, fehlender oder unbekannter Wert, doppelter Schlüssel, Abbruch mitten in der Transaktion, gleichzeitiger Zugriff — **Ausgang:** entfallen: entschieden in `spezifikation.md` (Produkt) vor dem ersten Code-Commit.
- Das Modell von `Exchange` ist breit (Redaction-, Integritäts- und Replay-Status) und wächst mit der Fachlogik — **Ausgang:** weiter offen: wird mit welle-record-replay sichtbar und dann als Beobachtung geführt.
- Der Port legt eine SQL-Eigenheit offen — **Ausgang:** entfallen: `a-check` hält SQL im Adapter, die Architektur verlangt eine fachliche Transaktionsgrenze.

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

Wird bei der Closure geschrieben.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (2×), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (1×) und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15); sie fließen als Randformen in §6 und als Tabelle *Zusage · Fänger* in §3 ein.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
