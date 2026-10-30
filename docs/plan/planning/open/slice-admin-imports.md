# Slice slice-admin-imports: Admin-Endpunkte für Imports und Import-Autorisierung

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-18`, `LH-FA-13`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-047` (Import-API-Vertrag) · `SPEC-005` (Zeilen `POST /__admin/imports`, `GET /__admin/imports/{id}`) · `LH-FA-18.a` · `SPEC-127` Nr. 3, 7, 14, 17, 18 (Anfrage, Antwort, Idempotenz) · `LH-FA-13.a` (Authentisierung) · `SPEC-051`, `SPEC-064`, `SPEC-065`, `SPEC-066`, `SPEC-068`, `SPEC-069`, `SPEC-071`, `SPEC-073`, `SPEC-097` · `SPEC-109` · `ARC-007` · `architecture.md` §4.2.
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

**Ziel:** `POST /__admin/imports` und `GET /__admin/imports/{id}` bilden auf `ImportUseCase` ab (Request und Statuscodes nach `SPEC-047`, Body der Antwort gleich dem `ReportJSON` des Imports, Wiederholung mit gleicher `import_id` mit 200 und byte-gleichem Body), und ein nicht autorisierter Client wird ohne Änderung abgewiesen: kein Aufruf des Use Case, `ODR-E2101`, Audit-Eintrag (`LH-FA-18`, Import-Autorisierung).

**Übernimmt:** den Endpunkt samt Authentisierung aus `slice-baseline-import` (dessen §1 *Ausdrücklich NICHT* ihn der Folgewelle zuweist; `AGENTS.md` §3.13) und die Abnahmeregel *Import-Autorisierung* von `LH-FA-18`, die ohne Admin-Listener nicht belegbar ist ([Abgleichsbericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §8 Nr. 1).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Fachlogik des Imports (Lesen, Zusammenführen, Coverage-Beweis, Konfliktregel) — `ImportUseCase` (`done/`); der Adapter bildet nur HTTP ab.
- Mehrere Quell-Sessions und andere Konfliktstrategien als `reject` — `SPEC-047` lässt im MVP eine Session, `target=new_baseline` und `conflict_policy=reject` zu; der Endpunkt weist jeden anderen Wert mit `400` ab, und kein Folge-Slice nimmt den Punkt an (`slice-baseline-import` §1).
- Rollen oder Berechtigungsstufen — Out-of-Scope der Welle (§6): „autorisiert“ heißt hier „trägt das gültige Token“ (Lastenheft-Lücke in §6).
- Der Rahmen (Listener, Token, Fehlerobjekt) — `slice-admin-listener-token`; der Port `AuditLog` — `welle-betrieb` (`slice-audit-log`).

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

- [ ] `POST /__admin/imports`: Body nach `SPEC-047` (Pflichtfelder, Standardwerte, im MVP nicht zulässige Werte `400`), Abbildung auf `ImportAnfrage`; Statuscodes `201`, `200` (Wiederholung, Body byte-gleich), `400`, `401`, `404`, `409`, `422` und `500` mit den `ODR`-Codes der Tabelle in `SPEC-047`; der Body der Antwort ist der `ReportJSON`.
- [ ] `GET /__admin/imports/{id}`: der Report; ein unbekannter Import ergibt `ODR-E7005` (404).
- [ ] Import-Autorisierung (`LH-FA-18`): ohne oder mit falschem Token weist der Endpunkt ab (`ODR-E2101`), `ImportUseCase` wird nicht gerufen (Zähler 0), der Audit-Eintrag steht; jeder `POST` ist auditiert (Umfang nach Vertrag).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/adapters/driving/admin/` | update | Routen und Abbildung der Imports |
| `spec/spezifikation.md` | update (Architect) | Vertrag der Abbildung: Body, Tabelle Fehler → Status → Code, Antwort; Nachzug `SPEC-047`, `SPEC-005` |
| Tests | neu | Adapter gegen Fakes von `ImportUseCase` und `AuditLog`: jedes Argument festgehalten, jede Operation gezählt, auch die, die der Adapter nie rufen darf (`AGENTS.md` §3.15); ein Fall je Zeile der Statustabelle von `SPEC-047` und je Fehlerart des Ports (Schleife über beide Mengen) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

| Zusage | Fänger |
|---|---|
| *(Architect: eine Zeile je Zusage, Randform und Fall des Vertrags)* | *(ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates)* |

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Body:** JSON-Randformen (Pflichtfelder, `null`, falsche Typen, unbekannte Felder, doppelte Schlüssel, Größe, `Content-Type`); wer `400` liefert, der Adapter oder `ImportUseCase` (`SPEC-127` Nr. 3 prüft die Form der `import_id`); der Code des `400` (`SPEC-047` nennt keinen `ODR`-Code, Übernahme in `slice-admin-listener-token`).
2. **Statusabbildung:** Tabelle Fehler → Status → Code für jede Zeile von `SPEC-047`; Rangfolge mehrerer Verstöße nach `SPEC-127` Nr. 4; `ODR-E3001` und `ODR-E9001` mit `500`.
3. **Antwort:** `201` oder `200` aus `ImportErgebnis.Neu`; der Body ist der `ReportJSON` byte-gleich, ohne Neuserialisierung; Header (`Content-Type`); `GET` ebenso.
4. **Autorisierung:** was „nicht autorisierter Admin-Client“ heißt (nur fehlendes oder falsches Token, Lücke in §6) und die Reihenfolge: Authentisierung vor dem Lesen des Bodys.
5. **Abbruch:** der Client bricht während des Imports ab: Kontext, atomares Rollback (`SPEC-047`), lange Läufe gegen Timeouts des Servers.
6. **Audit:** Umfang (nur erfolgreiche oder jeder Import), Felder ohne Geheimnisse (`import_id`, `service_id`, Quell-Session), Reihenfolge.
7. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* der Endpunkt aus `slice-baseline-import` (`done/`, nicht änderbar): angenommen, §1 *Übernimmt* trägt ihn mit der Kennung des Gebers; die Import-Autorisierung ist die Abnahmeregel von `LH-FA-18`. *Von diesem Slice an:* keine.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token` und `slice-audit-log` (`welle-betrieb`) liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** Rahmen, Authentisierung und Fehlerabbildung kommen aus `slice-admin-listener-token`, der Port `AuditLog` für den Schreibendpunkt aus `slice-audit-log` (beide *Code-Abhängigkeit*). `ImportUseCase` liegt seit `slice-baseline-import` in `done/` (`src/hexagon/ports/inbound/import.go`, Port vorhanden).

**Nicht gebraucht:** `slice-inspection-use-case` (der Endpunkt liest keine Exchanges), `slice-admin-sessions`, `welle-mock`, M2, `slice-http-adapter` und `slice-bootstrap-m6`. Der Slice hängt damit nur am Rahmen und am Audit; die Reihenfolge der drei Endpunkt-Slices ist frei.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): der Code der `400`-Antworten ist nicht entschieden (Übernahme in `slice-admin-listener-token`), oder „nicht autorisiert“ meint mehr als „kein gültiges Token“ (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): „(Vertrag, offen)“ heißt, der Architect trägt den Ort in der Spezifikation vor dem Übergang `next` → `in-progress` ein; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `POST /__admin/imports` (Body: `import_id`, `service_id`, `source_session_ids`, `metadata_hash`, `conflict_policy`, `target`) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| `POST /__admin/imports` ohne oder mit falschem Token (Import-Autorisierung) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| `GET /__admin/imports/{id}` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Audit-Eintrag des `POST` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- `SPEC-047` nennt für `400` (fehlerhaftes JSON, fehlende Pflichtfelder, nicht zulässiger Wert) keinen `ODR`-Code, und `SPEC-109` kennt keinen für ein formal ungültiges Argument — **Ausgang:** offen; Adresse: Architect (Übernahme in `slice-admin-listener-token`, Auftrag 1).
- Die größte zulässige Größe des Request-Bodys ist nicht festgelegt (`source_session_ids`, `import_id`) — **Ausgang:** offen; Adresse: Architect (Auftrag 1).
- `LH-FA-18` fordert „nicht autorisierter Admin-Client“, `LH-FA-13.a` kennt ein Token ohne Rollen: ob „nicht autorisiert“ mehr heißt als „kein gültiges Token“, ist im Lastenheft nicht gesagt — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User; der Slice setzt „gültiges Token“ (§1) und belegt die Regel damit.
- Ein langer Import und der Abbruch des Clients: das Rollback ist atomar (`SPEC-047`), das Verhalten gegen Timeouts des Servers hat keinen Ort (Auftrag 5) — **Ausgang:** offen; Adresse: Architect.
- Der Body der Wiederholung ist byte-gleich (`SPEC-047`); eine Neuserialisierung im Adapter bräche die Zusage — **Ausgang:** offen bis zur Closure; Fänger: zwei Aufrufe, Vergleich der Bytes, Mutation „Neuserialisierung“ rot.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×): die Fakes von `ImportUseCase` und `AuditLog` tragen jede Schnittstelle des Originals und zählen jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** offen; Adresse: Planner beim Start (§4).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes von `ImportUseCase` und `AuditLog`, außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (5×, `AGENTS.md` §3.17; jede Zeile der Statustabelle und jede Fehlerart des Ports als Schleife), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
