# Slice slice-expand-bausteine: Beziehungen der Navigationen bestimmen und eingebettete Entities schreiben

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-264` (Datenbasis, Angaben des Modells, Bestimmung) · `SPEC-266` (eingebettete Entities schreiben) · `LH-FA-16.a` (eigener Serializer-Contract-Test) · `SPEC-125` Nr. 5, 7, 11 (Metadatenleser, additiv) · `SPEC-171` (Leser der Version 2.0) · `SPEC-168`, `SPEC-169`, `SPEC-126` Nr. 9 c (Schreiber und Gegenstück) · `SPEC-262`, `SPEC-263` (Knoten und Prüfung von `$expand`) · die Zeiger in `SPEC-002`, `SPEC-125`, `SPEC-168`, `SPEC-169` · `ARC-001`, `ARC-004`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-28.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Metadatenleser der Version 4.0 hält `Partner` und die Referenzbedingungen des `$metadata` fest, die reine Funktion `model.Beziehungen` bestimmt daraus je Navigation von `$expand`, welche Entities zueinander gehören (`SPEC-264`), und der Schreiber des OData-JSON gibt `model.Erweiterung` als eingebettete Entities aus (`SPEC-266`). Das Verhalten des Binary bleibt: `$expand` antwortet weiter mit `501`.

**Übernimmt (Schnitt von `slice-expand-ausfuehrung`, Geber [`slice-expand-ausfuehrung`](slice-expand-ausfuehrung.md), `AGENTS.md` §3.13):**

- die Angaben `Partner` und `Bedingungen` im Metadatenleser und `model.Beziehungen` (`SPEC-264`): Zeilen A1 bis A11 der Tabelle *Zusage · Fänger*, die Zeilen „Angaben des `$metadata` lesen“ und „Beziehung bestimmen“ des Rasters; darin die Bestimmung einer Navigation ohne `NavigationBinding` (`SPEC-264` Nr. 3 a, Sendung von `slice-expand-validator` an den Geber; die Antwort `501` des Binary bleibt beim Geber);
- den Schreiber eingebetteter Entities (`SPEC-266`): Zeilen B1 bis B9, die Zeile „Ausgabe im OData-JSON“ des Rasters und den Serializer-Contract-Test für `$expand` (`LH-FA-16.a`; Sendungen von `slice-odata-json-schreiben` und `slice-expand-parser` an den Geber);
- den Anteil dieser Zeilen am Block der Größenmarke.

**Naht zum Nehmer der Typen:** Dieser Slice liefert `model.Beziehung`, `model.BeziehungsPaar`, `model.Erweiterung`, `model.Entity.Erweiterungen` und `model.Beziehungen`; [`slice-expand-ausfuehrung`](slice-expand-ausfuehrung.md) ruft und füllt sie. Die Pflichten, die der Vertrag dem Aufrufer stellt, stehen dort in §1.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ausführen im Mock (`SPEC-265`): Einschalten (`urlparser.WithExpand`), Aufruf von `PruefeExpand` und `Beziehungen` im Dienst, Abbildung der Fehler, Coverage, Lesen und Auflösen gegen Daten, Setzen der `Erweiterungen` — [`slice-expand-ausfuehrung`](slice-expand-ausfuehrung.md) (dessen §1 und DoD tragen es).
- Fixtures mit geänderter Antwort (`ODV4-E-03` c, `SPEC-042` Nr. 5, `TestSatzVersion4OhneExpand`), Abnahme im Modus `mock` und Handbuch — [`slice-expand-ausfuehrung`](slice-expand-ausfuehrung.md); dieser Slice ändert keine Antwort des Binary.
- Parser und Validator von `$expand` — `slice-expand-parser`, `slice-expand-validator` liegen in `done/`; dieser Slice liest den geprüften Knoten.
- Import, Entity-State, Schema, Snapshots — die Datenbasis ändert keinen davon (`SPEC-264` Nr. 1).
- Der Leser der Antworten (`ODataCodec`) — er liest die eingebetteten Entities nicht (`SPEC-266` Nr. 6); Zeile B7 prüft den bestehenden Leser, ohne ihn zu ändern.
- Eingebettete Entities und Referenzbedingungen der Version 2.0 — deren Leser setzt die Angaben nie (`SPEC-264` Nr. 2 f), deren Schreiber lehnt Erweiterungen ab (`SPEC-266` Nr. 5); mehr entsteht nicht.

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

- [ ] Beziehungen (`SPEC-264`, `LH-RB-01`): Angaben `Partner` und `Bedingungen` im Leser der Version 4.0 samt Kopien, der Leser der Version 2.0 setzt sie nie, `model.Beziehungen` mit Ergebnis und Fehlern (Zeilen A1 bis A11).
- [ ] Schreiber und Serializer-Contract-Test für `$expand` (`SPEC-266`, `LH-FA-16.a`): `model.Erweiterung` mit Prüfung, Ausgabe und Rundlauf mit dem Leser, die Version 2.0 lehnt ab (Zeilen B1 bis B9).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist. Ein Handbuch-Kapitel entsteht nicht (das Binary antwortet unverändert, kein neuer Code, Schlüssel oder Befehl); §7 nennt das, und das Kataloge-Gate ist grün.
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
| `src/hexagon/model/service_model.go` | update | `Partner`, `Bedingungen`, `model.Referenzbedingung`, Kopien (`SPEC-264` Nr. 2) |
| `src/hexagon/model/beziehung.go`, `src/hexagon/model/beziehung_test.go` | neu | `model.Beziehungen` mit Typen und Fängern (Zeilen A5 bis A11) |
| `src/adapters/driven/odata/csdl/` | update | Leser der Version 4.0 hält die Angaben fest |
| `src/hexagon/ports/outbound/outboundtest/` | update | Vertrag des `MetadataCodec` (A1 bis A4), Fall der Version 2.0 (A5) |
| `src/hexagon/model/entity.go` | update | `model.Erweiterung`, `Entity.Erweiterungen` (`SPEC-266`) |
| `src/adapters/driven/odata/jsoncodec/`, `src/adapters/driven/odata/jsoncodecv2/` | update | Schreiber und Ablehnung (Zeilen B1 bis B9) |
| `spec/spezifikation.md` | update (Architect, vor dem Start) | Vertrag `SPEC-264`, `SPEC-266` ist geschrieben; offen ist allein die Randform aus §6 |

**Bestehende Tests, die mitziehen** (geprüft am Stand `40546730`): `model.NavigationProperty` ist mit dem Listenfeld nicht mehr vergleichbar; `src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_assoziation.go` vergleicht zwei Listen davon mit `slices.Equal` und übersetzt dann nicht. Die Anpassung prüft mindestens so streng wie vorher (`BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger`). Bricht eine Testdatei außerhalb der Pfade des Blocks, trägt der Plan sie in diese Tabelle ein und meldet sie dem Planner (§6).

**Entscheidungen zum Vertrag** (Architect; Ort in der Spezifikation, keine ADR berührt): Datenbasis sind die Referenzbedingungen des `$metadata` (`SPEC-264` Nr. 1); Navigation auf eine Entity über eigene Bedingungen, auf eine Collection über `Partner` (`SPEC-264` Nr. 3); fehlendes Binding und jede andere Form `501` (Nr. 4); Form der eingebetteten Entities und Kontext (`SPEC-266` Nr. 3, 4). Ein neuer Fehlercode und eine neue Ursache entstehen nicht.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, aus `SPEC-264` und `SPEC-266`; die Nummern sind die des ungeschnittenen Plans, `A` Beziehungen, `B` Schreiber; Testnamen sind Vorgaben, die der Implementer im Plan nachzieht; Orakel und Fixture-Modell entstehen im Test ohne Code der Funktion):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A1 | `Partner` steht unverändert im Modell, auch mit Pfad oder Leerraum, ohne Attribut `""`, auch an der geerbten Navigation (`SPEC-264` Nr. 2 a, e) | `TestNavigationPartner` über den Vertrag des `MetadataCodec`: Schleife über die Werte × deklariert und geerbt | `Partner` wird nicht oder getrimmt gesetzt | andere |
| A2 | Je `ReferentialConstraint` im Namensraum `edm` ein Paar in Dokument-Reihenfolge, fehlendes Attribut `""`; fremder Namensraum, andere Stelle und Inhalt des Elements ungelesen (`SPEC-264` Nr. 2 b) | `TestNavigationBedingungen`: 0, 1, 3 Elemente × Attribute vorhanden, fehlend, leer × Stelle und Namensraum | Nur das erste Element wird übernommen | andere |
| A3 | 32 Elemente stehen, 33 ergeben eine leere Liste (`SPEC-264` Nr. 2 c) | `TestNavigationBedingungenGrenze`: 31, 32, 33 | `>` wird `>=` | andere |
| A4 | Die Angaben lehnen kein Dokument ab und ändern den Bericht nicht (`SPEC-264` Nr. 2 d) | `TestNavigationBedingungenBericht`: je Form aus A1 und A2 gleiche Funde wie vor dem Lesen der Angaben, kein Fehler | Der Fund `konstrukt` entfällt für gelesene Bedingungen | andere |
| A5 | Zugriffe liefern Kopien; der Leser der Version 2.0 setzt beide Angaben nie (`SPEC-264` Nr. 2 e, f) | `TestModellBedingungenKopie`; Fall im Vertrag des V2-Lesers | `kopieTyp` kopiert `Bedingungen` flach | andere |
| A6 | Navigation auf eine Entity: Paare aus eigenen Bedingungen, `Ziel` genau die Schlüssel des Zieltyps, auch zusammengesetzt und in anderer Reihenfolge, auch Selbstbezug und abgeleitetes Zielset (`SPEC-264` Nr. 3 a, b, 4) | `TestBeziehungenEntity`: Generator über Schlüssel mit 1 und 2 Teilen × Reihenfolge × Zielset | `Quelle` und `Ziel` vertauscht | andere |
| A7 | Navigation auf eine Collection: Paare aus den Bedingungen des `Partner`, `Partner.Target` gleich *T* oder Basistyp, Bedingungen an *N* ungelesen (`SPEC-264` Nr. 3 c) | `TestBeziehungenCollection`: dieselben Achsen, dazu Partner auf Basistyp und Bedingung an *N* | Die Paare stammen aus *N* statt aus *P* | andere |
| A8 | Zielmenge nur über genau ein Binding auf ein Entity-Set passenden Typs, sonst `501` (`SPEC-264` Nr. 3 a) | `TestBeziehungenBinding`: kein, zwei gleiche, zwei verschiedene, `Singleton`, unbekanntes Set, fremder Typ, Basistyp des Ziels, abgeleiteter Typ, Pfad mit Cast | Ohne Binding gilt das einzige Set des Zieltyps | andere |
| A9 | Jede untaugliche Bedingung ist `501`: keine, zu wenige, zu viele, doppelter Schlüssel, unbekannte, nicht unterstützte oder verschieden typisierte Property (`SPEC-264` Nr. 3 b, d) | `TestBeziehungenUntauglich`: Schleife über alle Formen × `Quelle` und `Ziel` | Gleichheit der Schlüsselmenge wird Teilmenge | andere |
| A10 | Jeder untaugliche `Partner` ist `501`: leer, mit `/`, unbekannt, Collection, anderes `Target`, ohne Bedingung, widersprechendes Binding; ohne Binding des Partners gültig (`SPEC-264` Nr. 3 c) | `TestBeziehungenPartnerUntauglich` | Die Prüfung des Bindings von *P* entfällt | andere |
| A11 | Das erste nicht auflösbare Element entscheidet mit `ExpandFehler` `navigation_nicht_unterstuetzt`; Programmfehler sind `ErrUngueltig`; Argumente bleiben unverändert (`SPEC-264` Nr. 4) | `TestBeziehungenFehler`: Stellen 1 bis 3 × Verstoß; Programmfehler je Art | Das Ergebnis lässt das untaugliche Element aus | andere |
| B1 | Ein Mitglied je Erweiterung hinter den Properties in der Folge der Navigationen des Typs: Array, `[]`, Objekt, `null`, in Entity und Collection (`SPEC-266` Nr. 3) | `TestSchreiberErweiterungForm`: Sammlung und Einzel × 0, 1, 2 × Reihenfolge der Übergabe | Die Folge ist die der Übergabe | andere |
| B2 | Verknüpfte Entity: `@odata.id` mit ihrem Set, `@odata.type` und `@odata.etag` wie ein Element, kein Kontext; Stufe `none` ohne Annotationen mit den Mitgliedern (`SPEC-266` Nr. 3) | `TestSchreiberErweiterungAnnotationen`: Stufen × abgeleiteter Typ × ETag | `@odata.id` mit dem Set der Anfrage | andere |
| B3 | Zahlform und Escape von `[REDACTED]` gelten in verknüpften Entities (`SPEC-266` Nr. 3) | `TestSchreiberErweiterungWerte` | Die Zahlform wird nicht durchgereicht | andere |
| B4 | `@odata.context` und `@odata.count` wie ohne Erweiterung; die eingebettete Collection trägt weder Zahl noch Folgelink (`SPEC-266` Nr. 3, 4) | `TestSchreiberErweiterungKontext` | Der Kontext nennt die Navigation | andere |
| B5 | Jede ungültige Erweiterung ist `ErrUngueltig`, nichts geschrieben: Name, doppelt, Art, Set, Typ des Sets, zwei Entities bei einer Entity, ungültige oder selbst erweiterte Entity (`SPEC-266` Nr. 2) | `TestSchreiberErweiterungUngueltig`: Schleife über alle Prüfungen × Entity und Collection-Element | Die Prüfung der Tiefe entfällt | andere |
| B6 | Ohne Erweiterungen (`nil`, leer) sind die Bytes unverändert (`SPEC-266` Nr. 1) | Bestehende Golden-Tests; `TestSchreiberOhneErweiterung` | Eine leere Liste schreibt ein Mitglied | andere |
| B7 | `ODataCodec` liest den Body zu den Entities der Anfrage, `Uebersprungen` nennt die Navigationen (`SPEC-266` Nr. 6) | `TestSchreiberErweiterungRundlauf` über alle Typen | Ein Mitglied steht zwischen den Properties unter fremdem Namen | andere |
| B8 | `jsoncodecv2` lehnt nichtleere Erweiterungen ab (`SPEC-266` Nr. 5) | `TestSchreiberV2Erweiterung` | Die Erweiterung wird still ausgelassen | andere |
| B9 | Der Schreiber ändert die Erweiterungen nicht und behält keinen Verweis (`SPEC-266`, `SPEC-168` Nr. 9) | `TestSchreiberErweiterungUnveraendert` | Die Liste wird an Ort und Stelle sortiert | andere |

**Größenmarke, Teil des Architect** (Pfade der Zeilen `A` und `B`; `produktcode` ist der Anteil, den der Planner beim Schnitt aus den 590 Zeilen des Architect gesetzt hat, §4):

```groessenmarke
pfad: src/hexagon/model/service_model.go
pfad: src/hexagon/model/beziehung.go
pfad: src/hexagon/model/beziehung_test.go
pfad: src/hexagon/model/entity.go
pfad: src/adapters/driven/odata/csdl
pfad: src/adapters/driven/odata/jsoncodec
pfad: src/adapters/driven/odata/jsoncodecv2
pfad: src/hexagon/ports/outbound/outboundtest
produktcode: 300
fakes: 0
```

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): Der Vertrag `SPEC-264` bis `SPEC-266` liegt auf `main`, `slice-expand-parser` und `slice-expand-validator` liegen in `done/`, die offene Randform aus §6 (Rang zweier Verstöße von `model.Beziehungen`) ist in der Spezifikation entschieden, und der Planner hat dem User die Frage zum Umfang (§6) vorgelegt.

- **Voraussetzungen, geprüft am Stand `40546730`:** `welle-mock` liegt in `done/` (Trigger der Welle; Metadatenleser und Schreiber vorhanden); Knoten `ExpandOption`, `PruefeExpand` und `ExpandFehler` vorhanden.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Coverage, Workspace, HTTP-Adapter, Composition Root (die Funktionen sehen Modell und Werte, nichts ruft sie); `slice-batch-parser`, `slice-batch-validator`, `slice-batch-changesets`.
- **Schnitt vor dem Code** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1; Planner, 2026-11-28, nach dem Vorschlag des Architect): Der ungeschnittene Plan hatte 40 Zeilen und `produktcode` 590, *S* = 3790, über der Schnittgrenze. Die Zeilen `A` und `B` (dieser Slice) und `C` ([`slice-expand-ausfuehrung`](slice-expand-ausfuehrung.md)) teilen keinen Mechanismus; die Naht sind die Typen `model.Beziehung` und `model.Erweiterung`. Den `produktcode` hat der Planner als 300 und 290 verteilt (Setzung des Planner, der Architect nennt nur die Summe). `A` und `B` bleiben ein Slice: zusammen liegen sie unter der Schnittgrenze. Der Teil ist ein Baustein ohne Aufrufer im Binary; er ist für sich grün und durch seine Contract-Tests prüfbar, sichtbar am Binary wird er erst mit dem Nehmer (§6).
- **Größenmarke:** 20 Zeilen, *a* = 0, *b* = 20, *f* = 0, *p* = 300, *S* = 1900, Marke *M* = **2600**; keine Schnittfrage. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Block trägt `basis`, `schaetzung` und `marke` noch nicht (`make groessenmarke` meldet *Marke steht aus*); der Planner setzt sie beim Start, vor dem ersten Code-Commit.
- **Kollisionen:** kein offener Block nennt eine Datei unter den Pfaden (geprüft am Stand `40546730`: Blöcke unter `open/` und `in-progress/`; `slice-batch-parser` und `slice-batch-validator` schreiben `src/hexagon/model/batch*`). Kein Kettenslice.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 2600 Zeilen (Metrik wie oben). Eine Überschreitung schneidet nicht (`SPEC-128` Nr. 10 c); sie bekommt einen Eintrag `ueberschreitung` mit Grund im Block.
- `in-progress` → `open` (blockiert — Carveout?): eine Randform ohne Ort in `SPEC-264` oder `SPEC-266`, die der Architect nicht im laufenden Slice entscheiden kann.

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

- Randform ohne Ort: `model.Beziehungen` bei einem nicht auflösbaren Element und, an späterer Stelle, einem Namen ohne Navigation von *T* — `SPEC-264` Nr. 4 nennt das erste nicht auflösbare Element und den Programmfehler, nicht den Rang zwischen beiden (`SPEC-263` Nr. 4 stellt seine Programmfehler dem ganzen Knoten voran, den Namen ohne Navigation nennt es nicht darunter) — **Ausgang:** offen; Adresse: Architect, vor dem Start (`AGENTS.md` §3.12, §3.16).
- Der unterstützte Umfang ist eng (Datenbasis `SPEC-264` Nr. 1), und dieser Slice baut ihn; die Frage an den fachlichen Owner, ob er für `LH-FA-16` reicht, steht mit der Empfehlung des Architect in [`slice-expand-ausfuehrung`](slice-expand-ausfuehrung.md) §6 — **Ausgang:** offen; Adresse: User; der Planner legt sie vor dem Start dieses Slice vor.
- Die Formen der Beziehungen sind unbeschränkt (beliebige Modelle) — **Ausgang:** offen; Generator über Schlüsselteile, Reihenfolge, Selbstbezug und abgeleitete Typen (Zeilen `A6`, `A7`); Prüfgrenze zwei Schlüsselteile (die Zahl der Bedingungen führt Zeile `A3`).
- `model.NavigationProperty` erhält ein Feld mit Liste und ist damit nicht mehr mit `==` vergleichbar; `model.Entity` erhält `Erweiterungen` — **Ausgang:** offen; der Compiler meldet jede Stelle (§3 *Bestehende Tests*).
- Die Bausteine haben bis zum Nehmer keinen Aufrufer im Binary; bleibt `slice-expand-ausfuehrung` liegen, steht Code ohne sichtbare Wirkung im Repo — **Ausgang:** offen; Planner beim Start des Nehmers.
- Pflichten des Aufrufers gehen bei der Übergabe verloren (`BEO-REPO/sendung-nennt-nicht-alle-pflichten-des-vertrags`): sie stehen mit den Nummern des Vertrags in §1 des Nehmers — **Ausgang:** offen; der Reviewer hält die gelieferten Signaturen dagegen.
- Randformen (`AGENTS.md` §3.12, §3.16): Raster und Zusatztabelle unten; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*:**

| Operation | Fehlender, leerer, unbekannter, fremder Wert | Navigation leer | Ziel fehlt oder gelöscht | Ziel nicht nachweislich vollständig | Selbstbezug, Zyklus, Tiefe | Property-Maske, abgeleiteter Typ | Abbruch (Kontext), Wiederholung |
|---|---|---|---|---|---|---|---|
| Angaben des `$metadata` lesen | `SPEC-264` Nr. 2 a bis d | gilt nicht: liest keine Daten | gilt nicht: liest keine Daten | gilt nicht: liest keine Daten | geerbte Navigation `SPEC-264` Nr. 2 e; Tiefe gilt nicht: eine Liste | `SPEC-264` Nr. 2 e | `SPEC-125` Nr. 1, 14 |
| Beziehung bestimmen | `SPEC-264` Nr. 3, 4 | gilt nicht: liest keine Daten | gilt nicht: liest keine Daten | gilt nicht: liest keine Daten | Selbstbezug nach `SPEC-264` Nr. 3 wie jedes Paar; Tiefe `SPEC-262` Nr. 4 | `SPEC-263` Nr. 2 d, `SPEC-264` Nr. 3 a | kein `ctx`; `SPEC-264` Nr. 4 |
| Ausgabe im OData-JSON | `SPEC-266` Nr. 2 | `SPEC-266` Nr. 3 | `SPEC-266` Nr. 3 (`null`) | gilt nicht: der Schreiber kennt keine Coverage | `SPEC-266` Nr. 2 (keine zweite Ebene) | `SPEC-266` Nr. 3 | `SPEC-168` Nr. 1, 6, 9 |

**Zusatztabelle** (mehrere Verstöße und Modi):

| Paar oder Modus | Ergebnis | Fänger |
|---|---|---|
| zwei nicht auflösbare Elemente, Stellen 1 bis 3 | das erste entscheidet (`SPEC-264` Nr. 4) | `A11` |
| nicht auflösbares Element vor einem Namen ohne Navigation | offen, Architect (erstes Risiko oben) | `A11` nach der Entscheidung |
| zwei ungültige Erweiterungen | `ErrUngueltig`, nichts geschrieben; ein Rang ist nicht beobachtbar (`SPEC-266` Nr. 2) | `B5` |
| Stufe `none` | Mitglieder der Navigationen ohne Annotationen | `B2` |
| Version 2.0 | Leser setzt keine Angabe, Schreiber lehnt ab | `A5`, `B8` |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (Planner, 2026-11-28; Zähler gleich der Zahl der Dateien unter `evidence/`); der Planner zieht Treffer und Zähler beim Start nach.

- **Offen und einschlägig:** `BEO-REPO/sendung-nennt-nicht-alle-pflichten-des-vertrags` (4×, ohne Ausgang) für die Übergabe an den Nehmer (§1, §6); `BEO-REPO/schreiber-und-leser-nicht-als-paar-geprueft` (2×) für Zeile `B7` (Rundlauf als Schleife über die Klasse der Werte); `BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger` (2×) für den Test aus §3; `BEO-REPO/regel-in-zwei-adaptern-ohne-sensor` (1×) für die zwei Schreiber (`B8`); `BEO-REPO/schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices` (1×): geprüft, die Felder sind additiv, kein Konstruktor verlangt ein neues Pflichtfeld, der Stand ist ohne den Nehmer grün.
- **Verkörpert und zu beachten:** `vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `menge-ohne-ihre-positionen` (`AGENTS.md` §3.15), `fix-deckt-die-genannte-mutation-nicht-die-klasse` (§3.17), `randform-erst-im-review-entschieden` (§3.16), `berichtigung-erreicht-nicht-jede-stelle` (§3.9), `unbeschraenkte-menge-ohne-benannte-pruefgrenze`, `fixture-gleicht-dem-wert-der-mutation`, `betriebsregel-fuer-repo-dateien-verletzt`.

Der Slice ist nicht nebenläufig. Kein offener Eintrag verlangt vor diesem Slice einen eigenen Slice.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
