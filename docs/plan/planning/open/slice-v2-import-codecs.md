# Slice slice-v2-import-codecs: Der Import nimmt den Satz der Codecs je Dienst entgegen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-FA-18`, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-127` Nr. 1 (`baselineimport.Dienst.Codecs`, Prüfung und Wahl) · `SPEC-175` Nr. 3 und 4 (Satz der Implementierungen, Import je Version) · `ARC-002`, `ARC-004`, `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-01.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `outbound.CodecSatz` (`Version`, `Metadaten`, `Daten`, `NurLesend`) liegt im Port-Paket, `baselineimport.Dienst` trägt das Feld `Codecs *outbound.CodecSatz`, und der Importer nimmt je Aufrufstelle (`ParseMetadata`, `ParseCollection`, `ParseEntity`) den Codec des Dienstes, sonst den der Konfig (`SPEC-127` Nr. 1, `SPEC-175` Nr. 3 und 4). Die Tests des Imports für Version 4 laufen textgleich grün. Der Import ist damit für beide Versionen derselbe Dienst; sein Versionsunterschied liegt allein im Satz seines Dienstes ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Implementierungen für V2 (Metadaten-Leser, Leser und Schreiber der Antworten, URL-Parser) — `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-v2-payload-schreiben`, `slice-v2-query`; dieser Slice ist mit Fakes testbar und braucht keine.
- Die Felder `URL` und `Schreiber` des Satzes — die Slices, die ihre Schnittstellen liefern (`slice-v2-query`, `slice-odata-json-schreiben`), fügen sie hinzu; hier gibt es keine Schnittstelle, die sie tragen könnte.
- Die Tabelle Version → Satz im Composition Root und die Konfiguration der Version — `slice-v2-bootstrap`.
- Der Mock — `slice-mock-lesen` (nimmt `Codecs` je Dienst von Anfang an entgegen, Sendung von `slice-v2-bootstrap`).

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

- [ ] `Dienst.Codecs` mit der Prüfung nach `SPEC-127` Nr. 1 (`nil` ist die Codecs der Konfig; `Metadaten` oder `Daten` `nil` ist `ErrUngueltig`) und der Wahl je Aufrufstelle; die Tests des Imports für Version 4 sind textgleich grün.
- [ ] Fänger: ein Fake-Satz mit zwei Diensten verschiedener Sätze; je Aufrufstelle gehen `ctx`, `grenze`, Modell und Strom unverändert an den Codec des Dienstes und kein Aufruf an den anderen (Menge: die drei Aufrufstellen mal zwei Dienste, als Schleife); `Codecs` `nil` und Teil-`nil` als `ErrUngueltig`.
- [ ] Handoff nach `AGENTS.md` §3.9 (ein Commit für Code und Plan), Tabelle *Zusage · Fänger* und Absatz der Mutationen in §3.
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
| `src/hexagon/ports/outbound/` | update | `CodecSatz{Version, Metadaten, Daten, NurLesend}` (`SPEC-175` Nr. 3); `URL` und `Schreiber` fügen die Slices hinzu, die ihre Schnittstellen liefern |
| `src/hexagon/services/baselineimport/` | update | `Dienst.Codecs`, Prüfung in `New` (`SPEC-127` Nr. 1), Wahl des Codecs je Aufrufstelle |
| Tests | neu / update | Fake-Satz mit zwei Diensten; Schleife über Aufrufstellen mal Dienste; die Tests für Version 4 bleiben textgleich |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster in §6 füllt der Architect beim Start vor dem ersten Code-Commit; die Größenmarke setzt der Planner aus dieser Tabelle (`SPEC-128`).

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18).
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Schleifen aus dem Wortlaut der Zusage, nie eine Liste der Fälle, die einfielen (`AGENTS.md` §3.17).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` und `SPEC-127` Nr. 1 ist nachgezogen (beides erfüllt). Code-Abhängigkeit allein `slice-baseline-import` in `done/`; nicht `welle-mock`, nicht M2.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`), vorläufig aus der Grobschätzung des Architect (etwa 8 Zeilen der Tabelle *Zusage · Fänger*, davon *a* = 1 nebenläufig und *b* = 7 andere; *f* = 1 Port-Fake; *p* = 120 Produktcode): *S* = 85·1 + 80·7 + 200·1 + 120 = **965**, Marke *M* = 100 · ⌈135·965 / 10000⌉ = **1400**. Die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/hexagon/ports/outbound/` und `src/hexagon/services/baselineimport/`, Tests eingeschlossen. Keine Schnittfrage (*S* < 2223). Der Block nach `SPEC-128` Nr. 3 h folgt mit der Tabelle des Architect; bis dahin ist die Marke eine Prosa-Zeile und das Gate misst sie nicht.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`.

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

- Der Slice ändert den Vertrag eines Dienstes in `done/` (`SPEC-127` Nr. 1); die Tests für Version 4 müssen textgleich grün bleiben — **Ausgang:** offen; Fänger: die V4-Tests unverändert, `Codecs` `nil` als Fall.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Operationen `New` und `Import` mal Argumentart (`Codecs` `nil`, `Metadaten` `nil`, `Daten` `nil`, zwei Dienste mit verschiedenen Sätzen, Abbruch); Orte: `SPEC-127` Nr. 1 (Prüfung), Nr. 2 (Kontext), `SPEC-175` Nr. 3 und 4. Das Raster füllt der Architect vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-01); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für die Fakes der Codec-Ports; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
