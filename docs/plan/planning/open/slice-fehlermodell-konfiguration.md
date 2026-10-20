# Slice slice-fehlermodell-konfiguration: Fehlermodell mit ODR-Codes und Startkonfiguration

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-QA-05`, `LH-FA-12`, `LH-FA-04`, `LH-RB-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-006` (Fehlerobjekt) · `SPEC-016` bis `SPEC-100` (Fehler-Codes, §4 der Spezifikation) · `SPEC-007` bis `SPEC-098` (Defaults, §3) · `LH-FA-04.c` (Adressierung) · `LH-QA-05.a` · `ARC-012`.
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

**Ziel:** Das Domain-Modell trägt ein Fehlermodell (Code, HTTP-Status, Zielbereich, Fehlerobjekt nach `SPEC-006`), die Fehler der Outbound Ports und des SQLite-Adapters bilden auf diese Codes ab (darunter `ErrBeschaedigt`, `ErrBlobFehlt` und `ErrBlobBeschaedigt` auf `ODR-E3001`), und der Composition Root lädt die Startkonfiguration der Welle aus Umgebung und `SERVICES_FILE` und lehnt Ungültiges mit `ODR-E2001` ab.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Verwendung der Codes in den Services und im HTTP-Adapter — jeder Slice, der einen Code auslöst (`slice-record-service`, `slice-replay-service`, `slice-baseline-import`, `slice-http-adapter`), bildet ihn über dieses Fehlermodell ab; hier entsteht nur die Tabelle und die Abbildung.
- Schlüssel von Admin-API, Hybrid-Fallback und Listener-TLS (`ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`) — diese Bereiche liefert die Welle nicht (Welle §6); die Konfiguration kennt nur, was `record`, `replay` und der Import brauchen.
- Das OData-V4-Fehlerformat des Mock-Dienstes — gehört zum Mock, der nicht in der Welle liegt; hier gilt das Fehlerobjekt `SPEC-006`.

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

- [ ] Fehlermodell im Domain-Modell: jeder Code aus §4 der Spezifikation mit HTTP-Status und Zielbereich, Fehlerobjekt nach `SPEC-006`, Abbildung der Fehler der Outbound Ports auf Codes (Tabelle in der Spezifikation, Test über alle Fehler des Pakets `outbound` und des SQLite-Adapters).
- [ ] Konfiguration der Schlüssel `MODE`, `LISTEN`, `DATA_DIR`, `SERVICE_ID`, `SERVICE_ROOT_PATH`, `UPSTREAM_URL`, `SERVICES_FILE`, `REPLAY_*`, `MATCH_*`, `REDACT_*`, `MAX_CAPTURE_BODY`, `UPSTREAM_*`, `SHUTDOWN_TIMEOUT`, `RETENTION_MAX_AGE`, `SQLITE_BUSY_TIMEOUT`, `VOLATILE_HEADERS`, `LINK_REWRITE` mit den Defaults aus §3 der Spezifikation; `ODR-E2001` bei ungültigem Wert, Doppelkonfiguration (Datei und Einzelschlüssel), überlappenden Roots und fehlendem Pflichtwert.
- [ ] Tests der Randformen aus §6 gegen die Spezifikation.
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
| `src/hexagon/model/` | update | Fehlermodell: Code, Status, Zielbereich, Fehlerobjekt |
| `src/hexagon/ports/outbound/` | update | Abbildung der Port-Fehler auf Codes (`ErrBeschaedigt`, `ErrBlob*` auf `ODR-E3001`) |
| `src/bootstrap/` | update | Konfiguration aus Umgebung und `SERVICES_FILE`, Defaults, Validierung |
| `spec/spezifikation.md` | update | Syntax der Werte (Dauern, Größen, Listen), Abbildungstabelle der Fehler, Verhalten bei `mock` und `hybrid` in dieser Version |
| Tests der Pakete | neu | Tabellentests je Schlüssel und je Fehler nach `LH-QA-05.a` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

- Die Konfiguration liegt im Composition Root, weil nur er alle Pakete kennen darf; die Services erhalten typisierte Werte über ihre Konstruktoren, nie die Umgebung.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` ist eröffnet; kein Vorgänger-Slice. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 800 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 600 Zeilen in derselben Metrik.
- `in-progress` → `open` (blockiert): die Abbildung der Port-Fehler lässt sich unter den Kanten von `a-check` nicht an der geplanten Stelle ablegen; sie wandert dann in die Schicht, die beide Seiten kennen, und der Plan wird nachgezogen.

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

- Die Syntax der Werte ist nicht festgelegt: `MAX_CAPTURE_BODY` heißt `8 MiB`, `SHUTDOWN_TIMEOUT` `10s`, `RETENTION_MAX_AGE` `30d` oder `720h` — ob `d` auch für andere Dauern gilt, ob `8MiB` ohne Leerzeichen und `8388608` zulässig sind, steht nicht da — **Ausgang:** offen; geht vor dem ersten Code-Commit an den Architect (`AGENTS.md` §3.12).
- Was ein Prozess mit `MODE=mock` oder `MODE=hybrid` tut, solange die Welle diese Modi nicht liefert: Die Spezifikation nennt beide gültig, der Composition Root hat keinen Use Case dafür — **Ausgang:** offen; Entscheidung des Architect (Vorschlag: gültiger Wert, Start scheitert mit `ODR-E2001` und Begründung).
- Leere Werte: ein gesetzter, aber leerer Schlüssel gegen einen ungesetzten, eine leere Liste gegen den Default — **Ausgang:** offen; Raster-Zelle in der Spezifikation.
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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (2×); die Randformen der Konfiguration (Wert fehlt, leer, unbekannt, doppelt) stehen als Raster in §6, nicht als Liste.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
