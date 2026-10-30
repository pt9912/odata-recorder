# Slice slice-bootstrap-m3: Composition Root für den Modus mock und der M3-Test

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-12`, `LH-FA-06`, `LH-FA-07`, `LH-FA-08`, `LH-FA-09`, `LH-FA-10`, `LH-FA-11`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `SPEC-110` Nr. 2 und 8 (Schlüssel dieser Version, Modus) · `SPEC-011`, `SPEC-055`, `SPEC-075` · `SPEC-052` (Adressierung) · `ARC-005`, `ARC-012` · `architecture.md` §2.5 (Modus-Routing).
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

**Ziel:** Der Composition Root bindet im Modus `mock` genau `MockUseCase` an den OData-Listener, der Start verlangt keinen Upstream, die Startkonfiguration liefert `MODE=mock` (auch als Default), `STRICT_COVERAGE` und `PUBLIC_BASE_URL`, und der M3-Test aus §1 der Welle läuft grün in `make test`.

**Übernimmt:** den Nachzug von `SPEC-110` Nr. 2 und 8, den `slice-bootstrap-m2` für die Modi `record` und `replay` hinterlassen hat (der Modus `mock` ist dort „nicht unterstützt“).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Verdrahtung des Modus `hybrid` und der Schlüssel `HYBRID_*` und `FALLBACK_*` — `slice-bootstrap-m4`; sie bleiben bis dahin `ODR-E2001`, damit keine Sicherheitseinstellung still wirkungslos ist (`SPEC-110` Nr. 2).
- TLS des Listeners (`LISTEN_TLS_*`) und die Admin-Schlüssel — `welle-betrieb` und `welle-admin-api`.
- Neue Fachlogik — der Slice verdrahtet, was `slice-mock-lesen`, `slice-mock-schreiben` und `slice-snapshot-use-case` liefern.

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

- [ ] Verdrahtung und Konfiguration: Modus `mock` bindet `MockUseCase`, der Start verlangt keinen Upstream, `STRICT_COVERAGE` und `PUBLIC_BASE_URL` (Pflicht bei nicht-Loopback-Host) kommen aus dem Wert des Starts; Fänger, dass der Root sie weiterreicht und den Nullwert nicht stehen lässt (Muster aus `slice-upstream-tls`).
- [ ] M3-Test aus §3 der Welle, Referenzdienst beendet, keine Verbindung zu ihm (`LH-FA-12`, Happy Path).
- [ ] Nicht gelieferte Schlüssel bleiben `ODR-E2001` (`HYBRID_*`, `FALLBACK_*`, `ADMIN_*`, `LISTEN_TLS_*`), und die Spezifikation führt den Nachzug von `SPEC-110`.
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
| `src/bootstrap/` | update | Wahl des Use Case je Modus, Start im Modus `mock`, Schlüssel `STRICT_COVERAGE` und `PUBLIC_BASE_URL` |
| `src/adapters/driving/httpodata/` | update | Mock-Anfragen aller Methoden an den Port; der Adapter kennt den Modus nicht |
| `tests/` | neu | M3-Test mit dem Referenzdienst aus `slice-bootstrap-m2` |
| `spec/spezifikation.md` | update | Nachzug von `SPEC-110` Nr. 2 und 8 (Architect) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Default `MODE=mock` lässt den Start heute mit `ODR-E2001` scheitern (`SPEC-110` Nr. 8); wer auf dieses Verhalten testet, ist mit zu ändern.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-mock-abnahme` liegt in `done/`. Die Spezifikation trägt den Nachzug von `SPEC-110`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): der Wechsel des Defaults `MODE` bricht Tests, die den Start ohne `MODE` erwarten, und der Nachzug von `SPEC-110` liegt nicht vor (siehe §6).

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

- Der Wechsel des Defaults `MODE=mock` macht aus einem heute abgewiesenen Start einen gültigen; Tests, die den alten Fehler erwarten, sind mit zu ändern — **Ausgang:** offen; Liste der Tests im Vertrag des Architect.
- `PUBLIC_BASE_URL` ist bei Wildcard-Host (Container) Pflicht und prägt `Location` und `@odata.id` (`SPEC-055`); Randformen (fehlt, leer, mit Pfad, mit abschließendem `/`) — **Ausgang:** offen; Raster, Ort `SPEC-055` und `SPEC-110`.
- Der M3-Test läuft mit Referenzdienst, Recorder und Client im selben Prozess unter dem Detektor (`make test-race`) — **Ausgang:** offen; Muster und Ordnung der Fänger wie in `slice-bootstrap-m2` (Kanäle statt Schlaf).
- `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×): der Mock-Pfad im Composition Root erweitert die Pakete, die Fälle der Gegenproben mitfahren — **Ausgang:** weiter offen → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` im Register.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Offen und für diesen Slice einschlägig: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, der Mock-Pfad erweitert die Pakete, die Fälle der Gegenproben mitfahren; Ausgang in §6), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes im Root-Test. Verkörpert: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `AGENTS.md` §3.9) für den Nachzug von `SPEC-110`. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
