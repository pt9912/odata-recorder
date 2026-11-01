# Slice slice-v2-payload-lesen: Leser für V2-Antworten (JSON mit der Hülle `d`) für den Import

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-170` (Umfang, Typen) · `SPEC-172` Nr. 1 bis 5 (Lesen; Nr. 6 und 7 gehören `slice-v2-payload-schreiben`) · `SPEC-178` · `LH-FA-17.a` (versionierter Adapter für V2-Payloads, keine Vermischung von V2 und V4) · `LH-FA-17` Boundary (leeres Ergebnis bleibt im V2-Antwortformat konsistent) · `SPEC-126` (Vertrag des V4-Lesers als Vorbild; Schreibweisen der Typen Nr. 10) · `SPEC-127` Nr. 8 (Auswahl der Antworten durch den Import) · `SPEC-002` · `SPEC-006` (Fehlerobjekt) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
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

**Ziel:** Der Leser `jsoncodecv2` (Paket `src/adapters/driven/odata/jsoncodecv2`, `SPEC-172`) **liest** die JSON-Antworten eines V2-Dienstes (Hülle `d`) auf Entity- und Collection-Abfragen für den Import in typisierte Entities gegen ein `model.ServiceModel` mit Version `2.0`: Typen in der Schreibweise von V2 (`/Date(…)/`, `Edm.Int64` und `Edm.Decimal` als String), Anzahl `__count`, Folgelink `__next`, nicht gedeutete Konstrukte als übersprungen gemeldet; das leere Ergebnis bleibt eine leere Menge (`LH-FA-17`, Boundary). Das Modell baut der Test mit `model.NeuesServiceModel("2.0", …)`; der V2-Metadatenleser ist keine Voraussetzung. Der Schnitt von Lesen und Schreiben folgt `AGENTS.md` §3.14: beide teilen allein die Tabelle der Typen in der Spezifikation (`SPEC-170` Nr. 4), keinen Mechanismus im Code; der Rundlauf beider ist der Test von `slice-v2-payload-schreiben`.

**Übernommen von `slice-v2-metadaten`** (Geber; `AGENTS.md` §3.13, Sendung vor dem Nehmer eingetragen):

- Eine Fixture mit einer Property, deren `Unterstuetzt()` falsch ist und deren `EdmType` `Edm.DateTimeOffset(V2)` lautet (V2 nennt `Edm.DateTimeOffset` außerhalb der sieben Typen, `SPEC-170` Nr. 4): erwartet ist `Uebersprungen` mit Grund `typ` (`SPEC-172` Nr. 3); der Wert des Modells `Edm.DateTimeOffset`, das der V2-Metadatenleser aus `Edm.DateTime` bildet, bleibt `datetimeoffset` (`SPEC-170` Nr. 4). Beide Fälle gehören zum Lesen und stehen in der DoD (Lesen) und in §3 (Nr. 1 und 2).
- Der Test `src/adapters/driven/odata/versionstrennung_test.go` (eine Schleife über alle Verzeichnisse unter `src/adapters/driven/odata/`, kein Eintrag je Paket) kommt aus `slice-v2-metadaten`. Fehlt er beim Start dieses Slice, legt dieser Slice ihn nach demselben Wortlaut an; liegt er vor, ändert dieser Slice ihn nicht und die Schleife erfasst `jsoncodecv2` ohne Eintrag.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Das Schreiben der V2-Antworten (Entity, Collection, Anzahl, leeres Ergebnis, Fehlerobjekt, Darstellungskopf; `SPEC-172` Nr. 6 und 7) — `slice-v2-payload-schreiben`; es braucht `slice-odata-json-schreiben` und teilt keinen Mechanismus mit dem Lesen.
- Metadaten-Leser für V2 — `slice-v2-metadaten`; dieser Slice liest Entities gegen ein Modell, er bildet es nicht, und braucht den Leser nicht (das Modell ist derselbe Typ, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)).
- Query-Semantik und Kompatibilitätsfehler für Abfragen — `slice-v2-query`.
- Wahl des Adapters je Service und Konfiguration — `slice-v2-bootstrap`; die Anpassung des Imports an einen Satz von Codecs je Dienst ist `slice-v2-import-codecs`.
- V2-Mutationen und das Lesen von Client-Bodies für V2 — `welle-odata-v2` §6; was die Spezifikation nicht nennt, ergibt den Kompatibilitätsfehler.
- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`.

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

- [ ] Lesen: Antworten auf Entity- und Collection-Abfrage in typisierte Entities gegen das Modell der Version `2.0` (Hülle `d`, Typen in der Schreibweise von V2, `__count`, `__next`, nicht gedeutete Konstrukte als übersprungen gemeldet); kanonische Schreibweise je Typ; ein Modell anderer Version ist `ErrUngueltig`; V4-Leser und V4-Schreiber unberührt (ihre Tests laufen unverändert grün); eine nicht unterstützte Property (`Unterstuetzt()` falsch, `EdmType` `Edm.DateTimeOffset(V2)`) ist `Uebersprungen` mit Grund `typ`, und der Test `versionstrennung_test.go` über alle Verzeichnisse von `src/adapters/driven/odata/` liegt vor (aus `slice-v2-metadaten`, sonst von diesem Slice angelegt).
- [ ] Randformen des Lesers fremder Bytes (leer, falsche Hülle, V4-Antwort am V2-Leser, Typ-Schreibweisen, Zahlen jenseits von `float64`, doppelte Namen, Größe, Tiefe, Zerlegung des Stroms in Reads): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der Bodies.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (ein V2-Dienst erreicht es erst mit `slice-v2-bootstrap`, das auch das Kapitel zu V2 trägt).
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
| `src/adapters/driven/odata/jsoncodecv2/` | neu | Leser der V2-Antworten (`SPEC-172` Nr. 1 bis 5); kein Port wird geändert ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)) |
| Tests | neu | Fälle von `LH-FA-17` (Boundary, Happy Path) für das Lesen; Generator über die Klasse der Bodies |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Leser liest fremde Bytes (ein Format, das dieses Repo nicht schreibt): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Bodies mit Escape-Folgen oder Rohbytes entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Auftrag an den Architect:** erledigt (Entwurf zu [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), `Accepted`). Auftrag 1 (Format) steht in `SPEC-170` Nr. 2 (JSON, Atom außerhalb), 2 und 3 (Hülle, Mitglieder, leeres Ergebnis) in `SPEC-172` Nr. 2 bis 4, 5 (Grenzen, Fehler) in `SPEC-172` Nr. 1, 6 (V4-Body am V2-Leser) in `SPEC-172` Nr. 2; Auftrag 4 (Fehlerobjekt, Darstellungskopf) betrifft das Schreiben und gehört `slice-v2-payload-schreiben` (`SPEC-172` Nr. 6 e und g).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect mit dem Vertrag, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Eine Property mit `Unterstuetzt()` falsch und `EdmType` `Edm.DateTimeOffset(V2)` ist `Uebersprungen` mit Grund `typ` (`SPEC-172` Nr. 3), nicht `unbekannte_property` und kein Wert; aus `slice-v2-metadaten` | Test mit Fixture, die diese Property trägt (Wert gültig gegen den JSON-Typ, damit nicht `wert` greift) | Prüfung von `Unterstuetzt()` entfernt oder Grund `typ` durch `wert` ersetzt: rot | Sendung |
| 2 | Der Wert des Modells `Edm.DateTimeOffset` aus `Edm.DateTime` bleibt `datetimeoffset` (`SPEC-170` Nr. 4) | Test: Modell mit dieser Property liest `/Date(…)/` zu `datetimeoffset` | Abbildung auf eine andere Art: rot | Sendung |
| 3 | Eine Schleife über alle Verzeichnisse von `src/adapters/driven/odata/` hält die Versionstrennung (kein Eintrag je Paket); aus `slice-v2-metadaten` | `versionstrennung_test.go`; legt ihn dieser Slice an, nach dem Wortlaut des Gebers | `jsoncodecv2` importiert ein V4-Paket des Adapters: rot | Sendung |

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (V2 schreibt Zeitpunkte als `\/Date(…)\/`) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` (erfüllt), und der Vertrag samt Randformen steht in `SPEC-172` Nr. 1 bis 5 (erfüllt); die Tabelle *Zusage · Fänger* füllt der Architect vor dem ersten Code-Commit (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** keine auf einen offenen Slice. Der Leser steht hinter dem Port `ODataCodec` (`done/`) und braucht Modell und Fehlermodell (`done/`); das V2-Modell ist derselbe Typ `model.ServiceModel` ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)), die Tests bauen es mit `model.NeuesServiceModel("2.0", …)`, der V2-Metadatenleser ist keine Voraussetzung. Der Schnitt von Lesen und Schreiben ist ein Teil ohne geteilten Mechanismus (`AGENTS.md` §3.14): sie teilen allein die Tabelle der Typen in der Spezifikation (`SPEC-170` Nr. 4, `SPEC-172` Nr. 4), der Rundlauf beider liegt im Schreiben.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Workspace, Composition Root, M2 und `slice-bootstrap-m3` (der Codec arbeitet auf Bytes, Modell und Entities; der Rundlauf über Import und Mock gehört `slice-v2-bootstrap`), `slice-v2-query`, `slice-v2-metadaten`, `slice-odata-json-schreiben`, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`. Das heißt: **startbar ohne `welle-mock` und ohne M2, sobald [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted` ist.**
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, Leser fremder Bytes haben in der Formel keinen Zuschlag mehr, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil; der V4-Leser `json` (Zeile *json* in `SPEC-128` Nr. 2) ist die Messlatte für die Erwartung des Lesens.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Der Schnitt zwischen Lesen und Schreiben ist vor dem Code gemacht; die Grobschätzung des Architect für das Lesen ist *S* ≈ 4500 (45 Zeilen mal 80 plus Produktcode 900), die Schnittfrage von `SPEC-128` bleibt damit gestellt; die Hebung des JSON-Scanners wäre die Alternative, sie steht beim Start in §4.
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

- Der Import eines V2-Dienstes braucht mehr als den Leser: die Auswahl der Antworten (`SPEC-127` Nr. 8: heute `application/json` mit `odata.metadata`) und die Ketten über Folgelinks und Anzahl (`SPEC-127` Nr. 12) sind V4-Regeln in `done/`; die Anpassung trägt `slice-v2-import-codecs` (`SPEC-175` Nr. 4) — **Ausgang:** eingetreten: `slice-v2-import-codecs`.
- Die Menge der Bodies ist unbeschränkt (jede Bytefolge) — **Ausgang:** offen; Generator über die Klasse mit Orakel und Zählprüfung oder Prüfgrenze als akzeptiertes Negativ mit Grund und Folge im Betrieb (`AGENTS.md` §3.17).
- Das Format der V2-Antworten (JSON oder Atom) steht im Lastenheft nicht — **Ausgang:** entfallen: JSON, Atom außerhalb des Umfangs (`SPEC-170` Nr. 2).
- Zeit- und Zahlenschreibweisen von V2 (Zeitpunkte als Millisekunden, `Edm.Int64` und `Edm.Decimal` als Zeichenketten): jede Abbildung auf die kanonische Form ist verlustbehaftet oder nicht — **Ausgang:** offen; Raster, eine Zelle je Typ (Schleife über die sieben Typen, `SPEC-172` Nr. 4).
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Body leer oder keine Menge | falsche Hülle (V4-Body am V2-Leser) | unbekannte Property, nicht unterstützter Typ | Typ-Schreibweisen, Zahlen jenseits `float64` | doppelter Name, über Grenze (Größe, Tiefe) | Wiederholung, Abbruch (Kontext, Strom) |
|---|---|---|---|---|---|---|
| Lesen einer Entity | leer: `syntax` (`SPEC-172` Nr. 1, `SPEC-126` Nr. 4) | `struktur`, Ort `/d` (`SPEC-172` Nr. 2) | `unbekannte_property` bzw. `Uebersprungen` mit `typ` (`SPEC-172` Nr. 3) | `SPEC-172` Nr. 4 (Tabelle, `/Date(…)/`), `zahl_zu_lang` (`SPEC-172` Nr. 1, `SPEC-126` Nr. 5) | `SPEC-172` Nr. 1 (`SPEC-126` Nr. 4 und 5: `doppelter_name`, `zu_gross`, `tiefe`) | `SPEC-172` Nr. 1 (`SPEC-126` Nr. 1, 2, 15) |
| Lesen einer Collection | `{"d":{"results":[]}}` und `{"d":[]}` gültig, `Entities` leer, nicht `nil` (`SPEC-172` Nr. 2 a); Body leer: `syntax` | `struktur`, Ort `/d` (`SPEC-172` Nr. 2) | wie Entity (`SPEC-172` Nr. 3) | wie Entity; `__count` als String oder Zahl (`SPEC-172` Nr. 2 a) | wie Entity | wie Entity |

Die Zelle *unbekannte Property, nicht unterstützter Typ* führt auch die Property mit `Unterstuetzt()` falsch und `EdmType` `Edm.DateTimeOffset(V2)`: `Uebersprungen` mit `typ`, Ort `SPEC-172` Nr. 3, Fänger Nr. 1 der Tabelle in §3 (Sendung von `slice-v2-metadaten`). Die Zelle *Typ-Schreibweisen* führt den Wert des Modells `Edm.DateTimeOffset` aus `Edm.DateTime`: `datetimeoffset`, Ort `SPEC-170` Nr. 4, Fänger Nr. 2.

- Die Sendungen von `slice-v2-metadaten` (Fixture `Edm.DateTimeOffset(V2)`, Test `versionstrennung_test.go`) — **Ausgang:** offen; Adresse: dieser Slice, §1 *Übernommen*, DoD (Lesen), §3 Nr. 1 bis 3.

Die Zeilen zum Schreiben des Rasters stehen in `slice-v2-payload-schreiben`.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für Fakes der Codec-Ports und Hüllen von Strömen (jede Schnittstelle des Originals, jeder Read gezählt); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
