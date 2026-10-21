# Slice slice-http-adapter: HTTP-Driving-Adapter für Record und Replay

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-FA-04`, `LH-FA-02`, `LH-QA-05`, `LH-RB-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-01.a` · `LH-FA-04.b`, `LH-FA-04.c` · `SPEC-052` · `LH-FA-02.b`, `LH-FA-02.c` · `LH-FA-13.a` (Korrelations-ID) · `SPEC-006` · `SPEC-059` · `ARC-005` · `architecture.md` §2.2, §2.5.
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

**Ziel:** Der Driving Adapter `src/adapters/driving/httpodata` nimmt HTTP-Requests des OData-Listeners an, bildet Methode, Pfad, `raw_query`, Header und Body-Strom unverändert auf Domain-DTOs ab, ruft den Inbound Port des Betriebsmodus, streamt die Antwort, schließt die Ströme, vergibt die Korrelations-ID und schreibt Fehler als Fehlerobjekt.

**Übernimmt:** Die Korrelations-ID kommt aus `IDGenerator.KorrelationsID()` (`SPEC-114` Nr. 6), nicht aus eigenem Code des Adapters; die Tests setzen `supporttest.FestIDs` (Sendung von `slice-clock-idgenerator`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Entscheidungen über Cursor, Match-Key, Coverage oder CRUD — der Adapter entscheidet sie nicht (`architecture.md` §2.2).
- Die Wahl des Use Case je Modus — Composition Root (`slice-bootstrap-m2`); der Adapter kennt den Modus nicht.
- Span `odata.gateway.request` und Metriken — es gibt noch keinen `Telemetry`-Port (Welle §6); der Adapter vergibt die Korrelations-ID, die der Port später trägt.
- TLS des Listeners — Folgewelle.

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

- [ ] Abbildung und Streaming: Request und Response ohne Veränderung der Bytes, geliehene kontextgebundene Ströme, Schließen durch den Adapter; Fehlerobjekt nach `SPEC-006` mit Code, Status und Korrelations-ID aus der Tabelle der Codes.
- [ ] Lebenszyklus: Listener öffnen und schließen, bei SIGTERM-Ablauf neue Requests mit 503 abweisen, laufende bis `SHUTDOWN_TIMEOUT` abwarten; Transportfehler vor dem ersten Byte als Gateway-Status, danach keine Änderung des Status (`LH-FA-02.b`).
- [ ] Tests über `httptest` gegen Fakes der Inbound Ports, ohne SQLite.
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
| `src/adapters/driving/httpodata/` | neu | Handler, Mapping, Fehlerausgabe, Lebenszyklus |
| `spec/spezifikation.md` | update | Vertrag mit Randformen: Header-Behandlung, Hop-by-Hop, Abbruch, Größe des Requests |
| Tests des Pakets | neu | `httptest` mit Fake-Ports; Fälle je Fehlercode aus `LH-QA-05.a` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

- Der Adapter importiert nur Inbound Ports, Querschnitts-Ports und Domain (`a-check`); die Ports `RecordUseCase` und `ReplayUseCase` liefern die beiden Slices davor.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-record-service` und `slice-replay-service` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 1000 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 700 Zeilen in derselben Metrik.
- `in-progress` → `open` (blockiert): der Rohrequest lässt sich nicht ohne Typ der HTTP-Bibliothek an den Port geben (siehe §6).

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

- Header-Behandlung: mehrfach vorkommende Header, Groß-/Kleinschreibung, `Host`, `Connection`-Liste, `Transfer-Encoding` — **Ausgang:** offen; Raster *Header-Art × Behandlung*, Entscheidung in der Spezifikation.
- Client bricht ab, bevor der Use Case antwortet; Antwort bricht ab nach dem ersten Byte — **Ausgang:** offen; das Gate fährt kein `-race` (`CGO_ENABLED=0`): die nebenläufigen Fänger ordnen mit Kanälen, und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50`.
- Pfad ohne Service-Root, Pfad mit Prozent-Kodierung, `raw_query` mit ungültiger Kodierung — **Ausgang:** offen; Raster-Zelle, der Adapter reicht unverändert, der Use Case entscheidet (`ODR-E5005`).
- Request-Größe und Zeitgrenzen des Servers (`ReadHeaderTimeout`, Idle) — **Ausgang:** offen; die Spezifikation nennt nur `UPSTREAM_TIMEOUT` und `SHUTDOWN_TIMEOUT`; Architect.
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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (über der Schwelle, ohne Ausgang); der Slice ist nebenläufig (Abbrüche, Herunterfahren). Der Slice ist nebenläufig: das Gate fährt kein `-race` (`CGO_ENABLED=0`), die nebenläufigen Fänger ordnen mit Kanälen, jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3), und ein drittes Auftreten dieser Beobachtung in dieser Welle ist kein Befund mehr, sondern ein eigener Slice.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
