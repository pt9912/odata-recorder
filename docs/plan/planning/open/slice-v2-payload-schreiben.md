# Slice slice-v2-payload-schreiben: Schreiber für V2-Antworten (JSON mit der Hülle `d`) für den Mock

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-QA-05`, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-172` Nr. 6 und 7 (Schreiben, Gegenstücke) · `SPEC-170` Nr. 4 (Typen) · `SPEC-006` (Fehlerobjekt) · `SPEC-178` · `ARC-004`, `ARC-011`.
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

**Ziel:** Der Schreiber `jsoncodecv2` implementiert die Schreib-Schnittstelle der Codec-Gruppe für Version 2 (`SPEC-172` Nr. 6): Entity, Collection (mit `__count` bei verlangter Anzahl), das leere Ergebnis in der Form `{"d":{"results":[]}}`, die Anzahl als Ziffern, das Fehlerobjekt im V2-Format und den Darstellungskopf (Content-Type und Versionskennzeichen); was der Schreiber ausgibt, liest der Leser aus `slice-v2-payload-lesen` mit demselben Modell zu denselben Entities zurück (`SPEC-172` Nr. 7), der Rundlauf ist der Test dieses Slice.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Das Lesen der V2-Antworten — `slice-v2-payload-lesen`; es teilt mit dem Schreiben keinen Mechanismus im Code, nur die Tabelle der Typen in der Spezifikation.
- Query-Semantik und der Kompatibilitätsfehler `ODR-E6006` — `slice-v2-query`; das Fehlerobjekt hier schreibt, was `SPEC-006` liefert.
- Wahl des Adapters je Service, Konfiguration und das Setzen des Darstellungskopfs in der Antwort des Mock — `slice-v2-bootstrap`.
- V2-Mutationen und das Lesen von Client-Bodies für V2 — `welle-odata-v2` §6.

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

- [ ] Schreiben: Entity, Collection (`__count`), leeres Ergebnis, Anzahl, Fehlerobjekt und Darstellungskopf nach `SPEC-172` Nr. 6; ein Zeitpunkt, der keine ganze Millisekunde ist, ist `ErrUngueltig`; Rundlauf Schreiben → Lesen liefert dieselben Entities (`SPEC-172` Nr. 7); V4-Schreiber unberührt.
- [ ] Je Zelle des Rasters (§6) ein Ort in der Spezifikation und ein Fänger; Schleife über die sieben Typen und die Escapes der Zeichenketten.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag; §7 hält fest, dass das Binary unverändert blieb (ein V2-Dienst erreicht es erst mit `slice-v2-bootstrap`).
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
| `src/adapters/driven/odata/jsoncodecv2/` | update | Schreiber der V2-Antworten (`SPEC-172` Nr. 6 und 7) neben dem Leser aus `slice-v2-payload-lesen` |
| Tests | neu | Format je Ausgabe, Werte je Typ, Escapes, Rundlauf mit dem Leser |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) füllt der Architect beim Start vor dem ersten Code-Commit; den Vertrag der Schreib-Schnittstelle (aus `slice-odata-json-schreiben`) trägt er vorher in die Spezifikation ein und nennt dessen Nummer in `SPEC-172` Nr. 6.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (V2 schreibt Zeitpunkte als `\/Date(…)\/`) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Schleifen aus dem Wortlaut der Zusage, nie eine Liste der Fälle, die einfielen (`AGENTS.md` §3.17).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted`, `slice-odata-json-schreiben` liegt in `done/`, und der Vertrag der Schreib-Schnittstelle steht in der Spezifikation (`SPEC-172` Nr. 6 trägt dann nur das Format). Der Schreib-Teil ist deshalb nicht vor `slice-odata-json-schreiben` startbar.

- **Code-Abhängigkeit, mit Grund:** `slice-odata-json-schreiben` (die Schreib-Schnittstelle der Codec-Gruppe und der Darstellungskopf, Sendung von `slice-v2-bootstrap`); `slice-v2-payload-lesen` ist nicht Voraussetzung, der Rundlauf-Test braucht den Leser aber im Paket.
- **Größenmarke:** noch nicht gesetzt; der Planner setzt sie beim Start nach `SPEC-128` aus der Tabelle des Architect.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): der Vertrag der Schreib-Schnittstelle fehlt in der Spezifikation.

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

- Der Vertrag der Schreib-Schnittstelle (Argumente, Kontext, Fehler) steht erst mit `slice-odata-json-schreiben` in der Spezifikation — **Ausgang:** offen; Adresse: Architect, vor dem Start dieses Slice.
- Zeit- und Zahlenschreibweisen von V2 (`\/Date(<ms>)\/`, `Edm.Int64` und `Edm.Decimal` als String): Rundlauf mit dem Leser — **Ausgang:** offen; Raster, eine Zelle je Typ (Schleife über die sieben Typen, `SPEC-172` Nr. 6 f).
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten; jede Zelle nennt ihren Ort, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Orte des Architect; „Vertrag der Schreib-Schnittstelle“ erbt den Vertrag des Schreibers der Codec-Gruppe, dessen Nummer der Architect vor dem Start einträgt):

| Operation | Body leer oder keine Menge | falsche Hülle | unbekannte Property, nicht unterstützter Typ | Typ-Schreibweisen | doppelter Name, über Grenze | Wiederholung, Abbruch |
|---|---|---|---|---|---|---|
| Schreiben einer Entity | Vertrag der Schreib-Schnittstelle; Format: `SPEC-172` Nr. 6 a | gilt nicht: der Schreiber erzeugt die Hülle (`SPEC-172` Nr. 6 a) | Vertrag der Schreib-Schnittstelle | `SPEC-172` Nr. 6 f (Werte), Millisekunde: `ErrUngueltig` | Vertrag der Schreib-Schnittstelle | Vertrag der Schreib-Schnittstelle |
| Schreiben einer Collection, Anzahl, leeres Ergebnis | leeres Ergebnis: `SPEC-172` Nr. 6 c; Anzahl: Nr. 6 d | gilt nicht (wie oben) | Vertrag der Schreib-Schnittstelle | `SPEC-172` Nr. 6 f | Vertrag der Schreib-Schnittstelle | Vertrag der Schreib-Schnittstelle |
| Schreiben des Fehlerobjekts | gilt nicht: das Fehlerobjekt hat immer `code` und `message` (`SPEC-006`) | gilt nicht (wie oben) | gilt nicht: kein Typ im Fehlerobjekt | `SPEC-172` Nr. 6 e (Escapes nach Nr. 6 f) | Vertrag der Schreib-Schnittstelle | Vertrag der Schreib-Schnittstelle |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-01); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für Fakes und Hüllen von Strömen; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
