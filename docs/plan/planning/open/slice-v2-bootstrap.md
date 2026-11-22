# Slice slice-v2-bootstrap: Wahl des Satzes je Service-Version im Composition Root (Teil A der Schlussverdrahtung von OData V2)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-FA-12`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-170` Nr. 6 · `SPEC-174` Nr. 3 · `SPEC-175` Nr. 1 bis 6 · `SPEC-176` · `SPEC-110` Nr. 7, 9, 12, 17 (neu, vom Architect) · `SPEC-139` Nr. 2, 4, 6, 11 b, 12 a · `SPEC-107` Nr. 14 · `SPEC-127` Nr. 1 · `LH-FA-17.a` (versionierter Adapter, keine Vermischung von V2 und V4) · `LH-FA-12.a` (Betriebsarten) · `SPEC-054` (`SERVICE_ROOT_PATH`, Adressierung je Service) · `SPEC-091` (`SERVICES_FILE`: `id`, `service_root`, `upstream_url`, `odata_version`) · `SPEC-201` Nr. 11 (Kataloge nachziehen) · `ARC-005`, `ARC-012` · `architecture.md` §2.5 (Modus-Routing), §7 (Erweiterungspunkt OData V2).
Der Verweis zeigt **aufwärts**: Die Spezifikation nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30; Vertrag und Zuschnitt vom Architect am 2026-11-22.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Zuschnitt (Architect, 2026-11-22, vor dem Code, `AGENTS.md` §3.14):** Die Tabelle *Zusage · Fänger* des ganzen Gegenstands ergab *S* = 4125 (Teil A 2120, Teil B 2005, beide aus der Tabelle nach `SPEC-128` Nr. 1 gerechnet); *S* ≥ 2223 ist die Schnittfrage, und die zwei Teile teilen keinen Code: Teil A ist Produktcode im Composition Root mit Tests am Root, Teil B ist ein Test mit Prozessen, ein Referenzdienst als Test-Server und das Kapitel zu V2 und ändert keine Zeile Produktcode. **Dieser Slice ist Teil A.** Teil B ist ein eigener Slice, `slice-v2-rundlauf` (Name Vorschlag des Architect); seinen Plan legt der Planner aus dem Abschnitt *Zuschnitt Teil B* in §3 an, und er startet nach diesem Slice (Start-Trigger in §4).

**Ziel:** Der Composition Root wählt je Service den Satz der Implementierungen seiner OData-Version: die Version steht je Service in der Startkonfiguration (`SERVICE_ODATA_VERSION` bzw. `odata_version` je Eintrag der `SERVICES_FILE`, ohne Angabe `4.0`), der Root legt den Service mit dieser Version an, und Import und Mock erhalten Metadatenleser, Payload-Codec, URL-Parser, Schreiber und `NurLesend` dieser Version; der Modus `hybrid` mit einem Service `2.0` verhindert den Start (`ODR-E2001`); der V4-Pfad bleibt unverändert, und V2 und V4 mischen sich in keinem Service. Der Vertrag steht vor dem Code in der Spezifikation (`SPEC-110` Nr. 17, `SPEC-139` Nr. 2, 4, 6, 11 b, 12 a, `SPEC-175`).

**Übernimmt (Konfiguration und Anlage des Service):** die Angabe der Version je Service (`SERVICES_FILE` kennt heute keine, `SPEC-091`) und die Anlage mit der konfigurierten statt der festen Version (`SPEC-107` Nr. 14); die Ableitung der Version aus dem Modell trägt `slice-v2-metadaten` (Auftrag 6 dort, geliefert).

**Übernimmt (Wahl des Satzes, Sendungen von `slice-v2-import-codecs`, `slice-v2-metadaten`, `slice-v2-query` und `slice-v2-query-filter`, `AGENTS.md` §3.13):** die Tabelle Version → Satz im Root und die Verdrahtung in Import und Mock (`SPEC-175` Nr. 3 und 4, `SPEC-139` Nr. 11 b); darin die Verdrahtung `CodecSatz{URL: urlparserv2.New(urlparserv2.WithFilter(urlparserv2.NewFilterLeser()))}` für Version 2.0 (`NewFilterLeser` liefert `slice-v2-query-filter`; ohne ihn wäre jeder `$filter` Kompatibilität `filter`), im Fänger A17.

**Übernimmt (Kataloge, Sendung von `slice-handbuch-katalog-schluessel-kommandos` und `slice-handbuch-meldungskatalog`, `SPEC-201` Nr. 11):** der Schlüssel `SERVICE_ODATA_VERSION` bekommt seinen Leser und damit seine Zeile in der Tabelle *Einstellungen*; `ODR-E6006` wird erreichbar, verlässt `harness/katalog-ausnahmen.txt` und bekommt seine Zeile in der Tabelle *Fehlercodes* (Fänger A22). Das Kapitel zu V2 und das Beispiel gegen das Binary trägt der Rundlauf-Slice (Teil B), weil seine Aussagen der Rundlauf belegt (`SPEC-200` Nr. 1).

**Mock-Ergänzungen (Sendung des Architect, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)): entfallen, schon geliefert.** Der Mock liest `NurLesend` und den Darstellungskopf des Schreibers seit `slice-mock-zuordnung` und `slice-mock-schreiben` (`SPEC-195` Nr. 7, `SPEC-192` Nr. 3 Punkt 1, `SPEC-198` Nr. 8; `src/hexagon/services/mock/mock.go`, `schreiben.go`, `lesen.go`); der Slice ändert den Mock nicht. Die Wirkung am Binary belegt der Rundlauf (Teil B, Zeilen B04, B09).

**Import je Version:** entschieden in `SPEC-175` Nr. 4 (derselbe Importer, Satz je Dienst). `slice-v2-import-codecs` liefert `baselineimport.Dienst.Codecs`; dieser Slice trägt die Tabelle Version → Satz im Root und die Verdrahtung.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Metadaten-Leser, Payload-Codec und URL-Parser samt Leser von `$filter` für V2 — `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-v2-payload-schreiben`, `slice-v2-query` und `slice-v2-query-filter`; dieser Slice verdrahtet sie und ändert keinen von ihnen.
- Der Rundlauf (Aufzeichnen, Importieren, Workspace, Abfragen im Modus `mock` gegen einen Referenzdienst V2), der Referenzdienst V2 als Test-Server, das Kapitel zu V2 im Handbuch samt gebundenem Beispiel, der Fänger „Versions-Kopfzeilen ändern nichts“ (`SPEC-174` Nr. 6), das Schlüsselliteral-Paar von Leser und URL-Leser (Sendung von `slice-v2-payload-lesen`), die Wirkung von `NurLesend` und des Darstellungskopfs am Binary — `slice-v2-rundlauf` (Teil B, Zuschnitt vom Architect, siehe oben und §3). Begründung: Teil B ist ein Test ohne Produktcode, er braucht diesen Slice in `done/`, und beide zusammen überschritten die Schnittgrenze (`SPEC-128` Nr. 1). **Adresse:** die Datei von Teil B besteht noch nicht; bis sie besteht, ist die Adresse der Planner (`AGENTS.md` §3.13, §6 Zeile 1). Die Pläne in `done/`, die als Nehmer `slice-v2-bootstrap` nennen, sind Zeitdokumente und bleiben stehen; der Planner trägt die Umadressierung in `welle-odata-v2` ein.
- Aufzeichnen im Modus `record`: die Aufzeichnung ist von der Version unabhängig (`welle-odata-v2` §1); kein Teil des Record-Pfads ändert sich, und `record.Dienst`, `replay.Dienst` und `hybrid.Dienst` kennen keine Version (`SPEC-139` Nr. 4).
- V2-Mutationen und der Modus `hybrid` für einen V2-Dienst — `welle-odata-v2` §6 bzw. `welle-hybrid`; für einen V2-Dienst im Modus `hybrid` gilt `ODR-E2001` beim Start (`SPEC-175` Nr. 5, `SPEC-110` Nr. 17 b).
- Änderungen am V4-Pfad (Mock, Replay, Record) über die Ports hinaus, hinter denen der Adapter steht — `welle-odata-v2` §6.
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

- [ ] Konfiguration, Anlage und Wahl im Root: `SERVICE_ODATA_VERSION` und `odata_version` (`SPEC-110` Nr. 17, `SPEC-175` Nr. 1, `SPEC-176`; fehlend `4.0`, fremd `ODR-E2001`, `hybrid` mit `2.0` `ODR-E2001` an `MODE`), Anlage des Service mit der Version des Dienstes und Vergleich mit ihr (`SPEC-107` Nr. 14, `SPEC-139` Nr. 6), die Tabelle Version → `outbound.CodecSatz` mit zwei Einträgen und der Satz je Dienst in Import und Mock (`SPEC-175` Nr. 3, `SPEC-139` Nr. 11 b), `Build` weist eine unbekannte Version und `hybrid` mit `2.0` als `model.ErrUngueltig` ab (`SPEC-139` Nr. 2); ein V4-Dienst nutzt unverändert die V4-Adapter und alle Tests, die einen `Dienst` ohne Version bilden, laufen unverändert grün; je Zeile A01 bis A21 der Tabelle in §3 ein Fänger, je Zelle des Rasters in §6 ein Ort in der Spezifikation und ein Fänger.
- [x] Mock-Ergänzungen: **entfallen**, schon geliefert (§1; Beleg: `SPEC-195` Nr. 7, `SPEC-192` Nr. 3, `SPEC-198` Nr. 8 und die Tests von `slice-mock-zuordnung` und `slice-mock-schreiben`; die Wirkung am Binary belegt Teil B). Grund beim Geber: die Sendung des Architect an diesen Slice war vor der Lieferung des Mock-Kerns geschrieben.
- [ ] Kataloge und Zeilen des Handbuchs (`SPEC-201` Nr. 11): die Zeile `SERVICE_ODATA_VERSION` und die geänderte Zeile `SERVICES_FILE` (nennt `odata_version`) in der Tabelle *Einstellungen*, die Zeile `ODR-E6006` in der Tabelle *Fehlercodes*, die Zeile `ODR-E6006` aus `harness/katalog-ausnahmen.txt` entfernt; `make katalog-check` und `make handbuch-check` grün (A22). Ohne Kennung der Planung im Text (`SPEC-200` Nr. 1); das Kapitel zu V2 folgt mit Teil B.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen: hier als die Zeilen des Punktes 3 (das Binary liest `SERVICE_ODATA_VERSION` und kann `ODR-E6006` erreichen; das Kataloge-Gate verlangt beide Zeilen); das Kapitel zu V2 mit gebundenem Beispiel trägt Teil B, weil erst sein Rundlauf die Aussagen belegt (Grund: `SPEC-200` Nr. 1; Negativ in §6).
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
| `src/bootstrap/konfig.go` | update | Feld `Version` im Typ `Dienst`; die Konfiguration trägt den wirksamen Wert |
| `src/bootstrap/konfig_dienste.go` | update | `odata_version` als Feld der `SERVICES_FILE`, der Einzelschlüssel in `einzelDienst`, `SERVICE_ODATA_VERSION` in der Doppelkonfiguration |
| `src/bootstrap/konfig_version.go` | neu | Lesen und Aufzählung der Version (`SPEC-110` Nr. 17 a), Zusammenhang mit dem Modus (Nr. 17 b); das Literal `SERVICE_ODATA_VERSION` steht in einer Datei `konfig*.go` (`SPEC-201`) |
| `src/bootstrap/version.go` | neu | Tabelle Version → Satz, die Wahl, die wirksame Version eines Dienstes; die eine Stelle, an der der Root nach der Version wählt |
| `src/bootstrap/wurzel.go` | update | `satzVersion4` geht in die Tabelle; Mock und Import erhalten den Satz je Dienst; `Build` prüft Version und Modus vor dem Öffnen der Ablage |
| `src/bootstrap/lebenszyklus.go` | update | Anlage mit der Version des Dienstes, Vergleich mit ihr (`abweichungen`); die Zeilen des Listeners (Telemetrie) berührt der Slice nicht |
| Tests in `src/bootstrap/` | neu, update | `konfig_version_test.go`, `version_wahl_test.go`, `version_anlage_test.go` (neu); `haken_test.go` (`protDienste` hält den `model.Service` fest), `export_test.go` (Zugriff auf Tabelle und Wahl) |
| `harness/katalog-ausnahmen.txt`, `docs/user/benutzerhandbuch.md` | update | Zeilen nach `SPEC-201` Nr. 11 (Punkt 3 der DoD) |
| `src/hexagon/services/` (Mock, Import), die V2-Pakete unter `src/adapters/driven/odata/` | keine Änderung | geliefert (§1) |

**Dateien in `src/bootstrap/` (für den Block des Planners, `SPEC-128` Nr. 10 g):** angelegt `konfig_version.go`, `version.go`, `konfig_version_test.go`, `version_wahl_test.go`, `version_anlage_test.go`; geändert `konfig.go`, `konfig_dienste.go`, `wurzel.go`, `lebenszyklus.go`, `haken_test.go`, `export_test.go`. Kollision mit laufenden Slices: `lebenszyklus.go` ändert auch `slice-telemetry-einbau-gateway` (Funktion `starteListener`, andere Stelle als `legeDiensteAn` und `abweichungen`, ein Merge ohne Konflikt ist zu erwarten); `haken_test.go` und `export_test.go` sind gemeinsame Testhilfen, jede Änderung dort ist additiv. **Gate-Kette (Nr. 7 der Übergangsregeln): berührt der Slice nicht** (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel bleiben); `harness/katalog-ausnahmen.txt` ist eine Eingabe von `make katalog-check` und gehört nicht zu Nr. 7; ob sie die Prüfsumme der Gegenproben ändert, prüft der Integrationsschritt (`tools/harness/gegenproben-eingabe.sh`, Stempel nach dem letzten Commit unter den Eingabepfaden).

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind vom Architect entschieden, nicht vom Implementer.

- Der Slice liefert keinen Test-Server und keinen neuen Fake: der Port-Fake `protDienste` des Bestands (`haken_test.go`) hält künftig den ganzen `model.Service` jedes `Create` fest und zählt jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `AGENTS.md` §3.15); er zählt im Block als ein Fake (`SPEC-128`: Port-Fakes).

**Auftrag an den Architect:** erledigt, alle sechs *entschieden* (Orte in der Spezifikation; [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted`). Der Text der ursprünglichen Fragen bleibt zur Nachprüfung stehen:

1. Konfiguration: Schlüssel der Version je Service, Werte, Default, Prüfung; Verhalten bei einer Version, die der Speicher anders kennt. — **entschieden:** `SPEC-175` Nr. 1 und `SPEC-176`; die Prüfung steht in `SPEC-110` Nr. 7, 9 und 17 a (Aufzählung `4.0`, `2.0`, Grund `Wert nicht in der Aufzählung: 4.0, 2.0`, Randleerraum entfernt, fehlend, leer und `null` gleich `4.0`, falscher Typ `muss eine Zeichenkette sein`, zugleich mit `SERVICES_FILE` allein die Zeile der Doppelkonfiguration); der Widerspruch zum Speicher in `SPEC-107` Nr. 14 und `SPEC-139` Nr. 6 (`ODR-E2001 odata_version: …`, der Vergleich nie gegen eine Konstante). Die offene Rückgabe der Welle (`SPEC-110` Nr. 9 nenne `SERVICE_ODATA_VERSION` im ersten Satz nicht) war nur halb erledigt: Nr. 7 und der zweite Satz von Nr. 9 führten den Schlüssel, der erste Satz nicht; der Satz ist mit dieser Änderung nachgezogen.
2. Wahl: wo der Root die Adapter bündelt und wie Mock-Kern und Import sie erhalten; keine Kreuzung. — **entschieden:** `SPEC-175` Nr. 3 und 6, `SPEC-139` Nr. 11 b: eine Tabelle mit zwei Einträgen im Root, die Wahl an einer Stelle (Fänger A21), jeder Dienst trägt den Satz seiner Version in `Codecs` (Fänger A16, A19), der Rückfall der Konfig ist der Satz `4.0` (A18).
3. Import je Version: Auswahl der Antworten und Ketten für V2. — **entschieden:** `SPEC-175` Nr. 4 (derselbe Importer; `__next`, `__count`, Atom `kein_json`, `$inlinecount` `query`, Dokument anderer Version `ODR-E6004`) und `SPEC-127` Nr. 8; die Verdrahtung ist A19, die Antworten des Referenzdienstes und ihr Ausgang stehen im Zuschnitt Teil B (B02).
4. Modi: welche Modi ein V2-Dienst kennt. — **entschieden:** `record`, `replay`, `mock` wie `SPEC-170` Nr. 6; `hybrid` mit `2.0` ist `ODR-E2001` an `MODE` (`SPEC-110` Nr. 17 b, `SPEC-175` Nr. 5), in `Build` `model.ErrUngueltig` (`SPEC-139` Nr. 2); Fänger A07 bis A10 und A12.
5. Kennzeichnung der V2-Antwort im Header. — **entschieden:** `SPEC-172` Nr. 6 g und `SPEC-175` Nr. 7; geliefert vom Mock (§1); der Fänger am Binary ist B04.
6. Der Referenzdienst V2 des Tests. — **entschieden:** im Zuschnitt Teil B (Abschnitt *Der Referenzdienst V2*): Metadaten, Collection, ein leeres Ergebnis, eine nicht unterstützte Eigenschaft, die sieben Typen und ein Atom-Austausch; Ort: dieser Plan, bis der Planner den Plan von Teil B anlegt.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15, Teil A). Zusage-Mengen tragen ihre Elemente; jede Menge ist ein Generator aus dem Wortlaut der Zusage, nie eine Liste der Fälle, die einfielen (Betriebsregeln unten). Die Mengen: **W** = {`4.0`, `2.0`} (2); **R** = Randleerraum um W: 4 Zeichen (Leerzeichen, Tabulator, LF, CR) × 3 Seiten (links, rechts, beidseitig) × 2 Werte = 24; **F** = fremde Werte, 27: je Wert von W neun Abwandlungen (Präfix `v`, `V` und `0`; Suffix `0`, `.0` und `x`; Komma statt Punkt; Punkt entfernt; nur der Major `4` bzw. `2`; 18), fünf Nachbarn (`1.0`, `2.1`, `3.0`, `4.01`, `5.0`) und NBSP vor und nach jedem Wert von W (4); **T** = falsche Typen eines Feldes der Datei: Zahl, Wahrheitswert, Array, Objekt (4); **Belegungen** = die Folgen aus W der Länge N = 1, 2, 3 (2 + 4 + 8 = 14). Ein Wert mit NBSP oder anderen Zeichen jenseits von ASCII entsteht aus Bytewerten und wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A01 | Der Einzelschlüssel `SERVICE_ODATA_VERSION` mit `4.0` oder `2.0` erreicht `Dienste[0].Version` ohne Leerraum am Rand (`SPEC-110` Nr. 17 a). Menge W und R, 26 Eingaben | `TestVersionEinzelGueltig`: Schleife über W und R, Zählprüfung 2 und 24, Erwartung aus der Menge, nicht aus dem Code | Wert nicht gekürzt (24 rot); Wert fest `4.0` (13 rot); `Version` nicht gesetzt (26 rot) | andere |
| A02 | Fehlend oder leer gilt `4.0`, die Konfiguration trägt nie einen leeren Wert (Nr. 17 a). Menge 6: fehlend, leer, nur Leerraum je Zeichen (4) | `TestVersionEinzelFehlt`: Schleife über 6 Eingaben, Zählprüfung 6, `Version` gleich `4.0` | `Version` bleibt leer (6 rot); Default `2.0` (6 rot) | andere |
| A03 | Ein fremder Wert ist genau eine Zeile `ODR-E2001 SERVICE_ODATA_VERSION: Wert nicht in der Aufzählung: 4.0, 2.0`; die Zeile nennt den Wert nicht; `DATA_DIR` wird nicht angelegt (Nr. 17 a und c, Nr. 11). Menge F, 27 | `TestVersionEinzelFremd`: Schleife über F, Zählprüfung 27 und Disjunktheit gegen W und R; Zeile, fehlender Wert in der Meldung, fehlendes Verzeichnis | Vergleich mit `HasPrefix` (die Formen mit Suffix rot); Vergleich mit `strings.Contains` (die Formen mit Präfix oder Suffix rot); Kürzen mit `strings.TrimSpace` (Unicode, die 4 Formen mit NBSP rot); Meldung nennt den Wert (27 rot); Wert fällt auf `4.0` zurück (27 rot) | andere |
| A04 | Datei: `odata_version` je Eintrag wirkt für diesen Eintrag und für keinen anderen (Nr. 17 a, Nr. 9). Menge 30: Datei mit 1, 2 oder 3 Einträgen (6 Positionen) × 5 Wertklassen (`2.0`, `2.0` mit Randleerraum, fehlend, `null`, leer); die Nachbarn tragen den anderen Wert | `TestVersionDatei`: Schleife, Zählprüfung 6 und 30; je Eintrag `Version` | Feld nur für Eintrag 0 gelesen (jeder Eintrag hinter dem ersten mit Wert `2.0` rot); `odata_version` als unbekanntes Feld abgewiesen (jede Zelle mit dem Feld rot); Wert des Nachbarn übernommen (jede Datei mit zwei Einträgen rot) | andere |
| A05 | Datei, fremder Wert und falscher Typ: fremd ist `SERVICES_FILE[<i>].odata_version: Wert nicht in der Aufzählung: 4.0, 2.0` mit dem Index des Eintrags, ein Feld falschen Typs allein `muss eine Zeichenkette sein`, keine zweite Zeile (Nr. 17 a, Nr. 9). Menge: F und T × 6 Positionen (186) | `TestVersionDateiFremd`: Schleife, Zählprüfung 27, 4 und 6; die Meldung als Ganzes | Typ nach der Form geprüft, zwei Zeilen (24 rot); Index fest `[0]` (3 Positionen, die Einträge hinter dem ersten, rot); Eintrag ohne Meldung | andere |
| A06 | `SERVICE_ODATA_VERSION` zugleich mit `SERVICES_FILE` ist allein die Zeile `ODR-E2001 SERVICE_ODATA_VERSION: zugleich mit SERVICES_FILE gesetzt`, auch bei fremdem Wert (Nr. 9, 17 a). Menge 8: die vier Einzelschlüssel (`SERVICE_ID`, `SERVICE_ROOT_PATH`, `UPSTREAM_URL`, `SERVICE_ODATA_VERSION`) × {gültig, fremd} | `TestDoppelkonfiguration`: Schleife über 4 × 2, Zählprüfung 4 und 2; je Fall genau diese Zeile | Liste der Einzelschlüssel ohne den vierten (2 rot); zusätzliche Formzeile bei fremdem Wert (1 rot) | andere |
| A07 | `hybrid` mit mindestens einem Service `2.0` ist genau eine Zeile `ODR-E2001 MODE: Modus hybrid für Version 2.0 nicht unterstützt`, nicht eine je Service, und legt `DATA_DIR` nicht an (Nr. 17 b und c). Menge 12: der Einzelschlüssel mit `2.0` (1) und die Datei mit N = 1, 2, 3 Einträgen, jede Belegung mit mindestens einem `2.0` (1 + 3 + 7 = 11) | `TestHybridMitV2`: Schleife über 12 Konfigurationen, Zählprüfung 1, 3, 7; `LadeKonfiguration` und `Start` (Exit 1, Zeile in `stderr`, Verzeichnis fehlt) | Prüfung nur für den ersten Dienst (4 rot: die Belegungen, deren erster Dienst `4.0` ist); eine Zeile je V2-Dienst (5 rot: die Belegungen mit mindestens zwei `2.0`); Schlüssel `SERVICES_FILE[i].odata_version` statt `MODE` (11 rot) | andere |
| A08 | Nur `hybrid` mit `2.0` verletzt; `record`, `replay` und `mock` nehmen jede Version, auch gemischt; bei ungültigem `MODE` meldet allein `MODE` seine Form (Nr. 17 b, Nr. 8). Menge 10: {`record`, `replay`, `mock`, `hybrid`, ungültig `Mock`} × {`4.0`, `2.0`} | `TestVersionModus`: Tabelle, Zählprüfung 5 und 2; je Zelle die Meldung als Ganzes | Prüfung in jedem Modus (4 rot: `record`, `replay`, `mock` und der ungültige Wert mit `2.0`); bei ungültigem `MODE` zusätzlich die Zeile (1 rot); `hybrid` mit `4.0` verletzt (1 rot) | andere |
| A09 | Ein fremder Wert zählt weder als `2.0` noch als `4.0`: mit `hybrid` meldet er allein seine Form (Nr. 17 b). Menge F, 27, am Einzelschlüssel mit `MODE=hybrid` | `TestVersionFremdHybrid`: Schleife über F, Zählprüfung 27 | fremder Wert als `2.0` gewertet, etwa durch `!= "4.0"` (27 rot) | andere |
| A10 | Alle Verstöße stehen in einer Meldung, nach Schlüssel in Bytefolge; `MODE` vor `SERVICE_ODATA_VERSION` und vor `SERVICES_FILE[…]` (Nr. 17 c, Nr. 12). Menge 4 Begleitverstöße zu `hybrid` mit `2.0`: `SERVICE_ID`, `LISTEN` und `HYBRID_FALLBACK` ungültig, fremde Version eines zweiten Eintrags der Datei | `TestVersionMeldung`: Schleife über 4, Zählprüfung 4; Zeilen und Reihenfolge als Ganzes | Prüfung bricht beim ersten Verstoß ab (4 rot); Meldung ungeordnet (4 rot) | andere |
| A11 | `Build`: eine leere `Version` gilt `4.0`; `Create` erhält `4.0`, der Dienst trägt den Satz `4.0`, und die 37 Stellen der bestehenden Tests, die einen `Dienst` ohne Version bilden, laufen unverändert (`SPEC-139` Nr. 2) | `TestBuildVersionLeer` (Argument von `Create`, Satz) und die bestehende Suite ohne Änderung an diesen Stellen | `Create` mit leerer Version (Fehler des Speichers); leere Version als Fehler abgewiesen (die 37 Stellen rot) | andere |
| A12 | `Build` weist eine unbekannte Version und `hybrid` mit `2.0` als `model.ErrUngueltig` ab, ohne Code und ohne die Ablage zu öffnen (`SPEC-139` Nr. 2). Menge 82: F × {`record`, `replay`, `mock`} (81) und (`hybrid`, `2.0`) (1) | `TestBuildVersionUngueltig`: Schleife, Zählprüfung 27, 3 und 1; `errors.Is`, kein Code, kein `odata-recorder.db` | unbekannte Version fällt auf den Satz `4.0` zurück (81 rot); `hybrid` baut den Mock mit dem Satz `2.0` (1 rot) | andere |
| A13 | Die Anlage ruft je Dienst `Create(id, root, Version, "", "")` einmal, mit der Version des Dienstes, in der Reihenfolge der Dienste (`SPEC-139` Nr. 6, `SPEC-107` Nr. 14). Menge 14: die Belegungen für N = 1, 2, 3 | `TestAnlageVersion`: Schleife über 14, Zählprüfung 2, 4 und 8; der Port-Fake `protDienste` hält jeden `model.Service` fest und zählt jede Operation | `Create` fest mit `4.0` (11 rot: jede Belegung mit einem `2.0`); Version des ersten Dienstes für alle (8 rot: die gemischten Belegungen); Reihenfolge vertauscht (je Belegung mit zwei Diensten rot); `Create` nur für den ersten Dienst (12 rot: N = 2 und 3) | andere |
| A14 | Ein bekannter Dienst: gleiche Werte tun nichts (kein zweites `Create`); jede Abweichung verhindert den Start mit je einer Zeile `ODR-E2001 service_root: …` und `ODR-E2001 odata_version: …` in dieser Reihenfolge und je einem Logeintrag; der Vergleich geht gegen die Version des Dienstes (`SPEC-139` Nr. 6, 8 c). Menge 9: gespeichert {`4.0`, `2.0`} × konfiguriert {`4.0`, `2.0`, fehlend} (6) und, bei anderem Root, Version gleich, `4.0` zu `2.0`, `2.0` zu `4.0` (3) | `TestAnlageBekannt`: Tabelle, Zählprüfung 6 und 3; Meldung, Logeinträge und Zahl der `Create` | Vergleich gegen die Konstante `4.0` (die Zellen mit gespeichert oder konfiguriert `2.0` rot); Vergleich nur von `service_root` (jede Zelle mit anderer Version rot); Reihenfolge der zwei Zeilen vertauscht (die Zelle mit beiden Abweichungen rot); Logeintrag nur für das erste Feld (dieselbe Zelle rot) | andere |
| A15 | Abbruch und Wiederholung: bricht der Start zwischen zwei Anlagen ab, legt der nächste Start nur an, was fehlt, mit der Version des Dienstes, und ändert keine gespeicherte Version (`SPEC-107` Nr. 12 und 14). Menge 12: Abbruchstelle {vor dem ersten, zwischen den beiden, nach dem zweiten `Create`} × die vier Belegungen für N = 2 | `TestAnlageWiederholung`: Schleife, Zählprüfung 3 und 4; Zustand des Speichers nach dem zweiten Start | zweiter Start legt den ersten Dienst mit der Version des zweiten neu an (die gemischten Belegungen rot); Abbruch lässt einen Dienst ohne Version zurück | andere |
| A16 | Der Satz je Dienst im Mock: jedes Feld ist das der Version des Dienstes (`Version`, `Metadaten`, `Daten`, `URL`, `Schreiber`, `Eingabe` bei `2.0` `nil`, `NurLesend`); kein Dienst trägt den Satz eines anderen; ein V4-Dienst trägt den Satz des Bestands (`SPEC-139` Nr. 11 b, `SPEC-175` Nr. 3). Menge 14 Zellen (7 Felder × 2 Versionen) und 14 Belegungen | `TestWahlSatz`: Tabelle 7 × 2 und Schleife über 14 Belegungen an `mock.Konfig.Dienste[i].Codecs`, Typen mit `reflect.TypeOf`; Orakel: die Paketnamen `csdl`, `jsoncodec`, `urlparser` und die mit Endung `v2`, aus der Zusage, nicht aus dem Code | je Zelle ein Feld vertauscht (14 rot); alle Dienste tragen den Satz des ersten (8 rot: die gemischten Belegungen); `Codecs` `nil` bei `2.0` (11 rot: jede Belegung mit einem `2.0`); `NurLesend` falsch bei `2.0` (die Zelle rot) | andere |
| A17 | `WithFilter`: der URL-Leser beider Sätze trägt den Leser von `$filter` (`SPEC-139` Nr. 11 b, `SPEC-160` Nr. 8, `SPEC-173` Nr. 5). Menge 4: (`4.0`, `Id eq 1`), (`4.0`, `contains(Name,'a')`), (`2.0`, `Id eq 1`), (`2.0`, `substringof('a',Name)`) | `TestWahlFilterLeser`: Tabelle über `Satz.URL.ParseURL`; Erwartung: Ausdruck, Ausdruck, Ausdruck, `*model.KompatFehler` mit Ursache `filter` | `urlparserv2.New()` ohne `WithFilter` (Zeile 3 rot); `urlparser.New()` ohne `WithFilter` (Zeilen 1 und 2 rot) | andere |
| A18 | Rückfall: `MetadatenCodec`, `URLCodec` und `Schreiber` der Konfig des Mock und `MetadatenCodec` und `DatenCodec` der Konfig des Imports sind in jeder Belegung die Werte der Version `4.0`, auch bei Diensten, die alle `2.0` sind (`SPEC-139` Nr. 4, 11 b). Menge 5 Felder × 14 Belegungen | `TestWahlRueckfall`: Schleife, Zählprüfung 5 und 14 | Felder aus dem Satz des ersten Dienstes (7 Belegungen rot: der erste Dienst ist `2.0`); Felder `nil` bei nur `2.0` (`Build` scheitert, 3 Belegungen rot) | andere |
| A19 | Import: jeder Dienst des Imports trägt in `Codecs` den Satz seiner Version, auch bei `4.0`; `record.Dienst`, `replay.Dienst` und `hybrid.Dienst` tragen keine Version und gleichen für jede Belegung denen von V4; in `replay`, `mock` und `hybrid` ist `Import()` `nil` und `baselineimport.New` wird nicht gerufen (`SPEC-139` Nr. 4 und 11 b, `SPEC-170` Nr. 6). Menge 14 Belegungen × 4 Modi | `TestWahlImport`: der Haken `NeuerImport` hält `baselineimport.Konfig`, der Zähler der verbotenen Konstruktoren wie im Test des Modus `mock`; Zählprüfung 14 und 4 | `Codecs` `nil` bei `2.0` (11 rot: jede Belegung mit einem `2.0`); Satz des Nachbarn (die gemischten Belegungen rot); Import im Modus `mock` gebaut (rot); `Version` in `record.Dienst` (rot) | andere |
| A20 | `Etag` ist je Dienst wahr genau bei `ETAG_ENABLED` und dem Satz `4.0` (`SPEC-139` Nr. 11 a, `SPEC-280` Nr. 2). Menge 28: 14 Belegungen × {wahr, falsch} im Modus `mock` | `TestWahlEtag`: Schleife, Zählprüfung 14 und 2; `mock.Dienst.Etag` je Index | `Etag` für `2.0` an (11 rot: jede Belegung mit einem `2.0`, bei `ETAG_ENABLED` wahr); `Etag` aus dem Satz des ersten Dienstes (die gemischten Belegungen rot) | andere |
| A21 | Die Wahl steht an einer Stelle: im Quelltext des Pakets vergleicht allein `version.go` und die Prüfung der Konfiguration (`konfig_version.go`) einen Wert mit `"4.0"` oder `"2.0"`; kein Dienst verzweigt nach der Version (`SPEC-139` Nr. 11 b, `SPEC-175` Nr. 3). Menge: die Dateien des Pakets ohne Tests (Verzeichnis gelesen, Zählprüfung gegen die Zahl der Dateien) | `TestVersionLiteraleAnEinerStelle`: AST über jede Datei; erwartet genau die zwei Dateien | `wurzel.go` vergleicht `d.Version == "2.0"` (rot); `lebenszyklus.go` vergleicht mit `"4.0"` (rot); neue Datei mit dem Literal (rot) | andere |
| A22 | Kataloge: `SERVICE_ODATA_VERSION` steht als Zeile der Tabelle *Einstellungen*, `SERVICES_FILE` nennt `odata_version`, `ODR-E6006` steht als Zeile der Tabelle *Fehlercodes* und fehlt in `harness/katalog-ausnahmen.txt` (`SPEC-201` Nr. 11) | `make katalog-check` und `make handbuch-check`, Ausgabe `ok` (benannte Ausgabe des Gates) | Zeile `SERVICE_ODATA_VERSION` fehlt (Befund Art 7); Ausnahme `ODR-E6006` bleibt (Befund Art 10) | andere |

**Aufrufstellen des Roots, an denen die Version oder der Satz weitergegeben wird** (Menge, `AGENTS.md` §3.15; jede hat einen Fall je Belegung, nicht einen am ersten Dienst): (S1) `MetadataRepository.Create` mit der Version (Modi alle vier; A13, A14, A15); (S2) `mock.Dienst.Codecs` (Modi `mock`, `hybrid`; A16, A17); (S3) `baselineimport.Dienst.Codecs` (Modus `record`; A19); (S4) `mock.Dienst.Etag` (Modi `mock`, `hybrid`; A20); (K1) `mock.Konfig.MetadatenCodec`, `URLCodec`, `Schreiber` (Modi `mock`, `hybrid`; A18); (K2) `baselineimport.Konfig.MetadatenCodec`, `DatenCodec` (Modus `record`; A18). Das sind 12 Paare aus Stelle und Modus (4 + 2 + 1 + 2 + 2 + 1); in `hybrid` ist jeder Dienst `4.0`. Zählprüfung gegen die Quelle: A21 (Vergleiche mit dem Literal) und ein Fall je Paar in A13 bis A20.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (hier: NBSP in F) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen; der Bericht nennt je Menge die Zählprüfung.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

### Zuschnitt Teil B (`slice-v2-rundlauf`, Vorschlag): Rundlauf, Referenzdienst und Kapitel zu V2

Dieser Abschnitt ist der Vertrag von Teil B, vom Architect vor dem Code geschrieben. **Der Planner legt daraus den Plan `slice-v2-rundlauf` an** (§1 mit den Sendungen unten, DoD, §3 mit der Tabelle ohne die Spalte *Teil*, §6 mit dem Raster, den Block *groessenmarke* ohne Produktcode) und nimmt diesen Abschnitt dann aus diesem Plan heraus. Die Kopfzelle der Tabelle ist hier `Nr. B`, damit `make groessenmarke` genau eine Tabelle dieses Plans liest (`SPEC-128` Nr. 3 b); im Plan von Teil B heißt sie `Nr.`.

**Gegenstand, Trigger, Größe.** Der Rundlauf der Welle (Closure-Trigger von `welle-odata-v2`): ein Test zeichnet einen Referenzdienst V2 auf, importiert die Session, legt einen Workspace an, beantwortet im Modus `mock` die Metadaten und eine Standardabfrage nach dem dokumentierten V2-Umfang; ein leeres Ergebnis bleibt im V2-Format, eine nicht unterstützte Eigenschaft ergibt den Kompatibilitätsfehler `ODR-E6006`, ein V4-Dienst verhält sich wie vorher, das Handbuch führt V2. **Start:** dieser Slice (`slice-v2-bootstrap`) liegt in `done/`, dazu die Trigger des Gebers (`slice-v2-import-codecs`, `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-v2-payload-schreiben`, `slice-v2-query`, `slice-v2-query-filter`, `welle-mock`, `slice-bootstrap-m3`, `slice-handbuch-meldungskatalog` in `done/`). **Gate-Kette:** berührt Teil B nicht (nur `tests/integration/`, eine Datei in `src/bootstrap/`, `docs/user/`). **Größe** (Block, Architect): *S* = 85·1 + 80·19 + 200·2 + 0 = 2005 (eine nebenläufige Zeile, 19 andere, zwei Fakes: der Referenzdienst V2 und der Wächter um den Mock-Port in B10; kein Produktcode), *M* = 2800 nach `SPEC-128` Nr. 1; keine Schnittfrage (*S* < 2223). Pfade des Blocks: `tests/integration/v2_referenzdienst_test.go`, `tests/integration/v2_rundlauf_test.go`, `tests/integration/v2_mehrere_test.go`, `src/bootstrap/v2_ursache_test.go`; das Handbuch zählt nicht (`SPEC-128` Nr. 3 m).

**Übernimmt (Sendungen, `AGENTS.md` §3.13; jede der Quellen nennt als Nehmer `slice-v2-bootstrap`, der Planner trägt um):**

1. Von `slice-v2-query` und `slice-v2-query-filter`: der Rundlauf-Fänger „Versions-Kopfzeilen ändern nichts“ (`SPEC-174` Nr. 6, `SPEC-173` Nr. 7; B12); der Test über beide Seiten (Parser plus Interpreter, B06 und B14); die Kopie der Grammatik gegen V4 (Fänger: B06 und B14 gegen Literale in der Form von `SPEC-173` Nr. 3); der Handbuch-Eintrag von `ODR-E6006` ist mit diesem Slice (A22) geliefert.
2. Von `slice-v2-payload-lesen` (Sendung S1): das Schlüsselliteral des V2-Payload-Lesers (`SPEC-172` Nr. 5) gleicht dem Literal, das der V2-URL-Leser (`SPEC-173` Nr. 3) für denselben Schlüssel bildet, Schleife über die sieben Typen (B14).
3. Von `slice-v2-payload-schreiben` und `slice-mock-zuordnung`: der Darstellungskopf der Antwort des Mock und `NurLesend` sind am Binary belegt (B04, B09).
4. Von `slice-v2-payload-lesen`, `slice-v2-metadaten`, `slice-v2-query`, `slice-v2-payload-schreiben`: das Kapitel zu V2 im Handbuch mit Aufgabe, Voraussetzung, Konfiguration (`SERVICE_ODATA_VERSION`, `odata_version`), Umfang (Tabelle der Bereiche und ihrer Folge `ODR-E6006`), Betriebsarten (`hybrid` startet nicht), Aufzeichnen mit JSON (`Accept: application/json` oder `$format=json`; ein Atom-Austausch bleibt Replay-fähig und liefert beim Import `kein_json`), Antwortformat und Kennzeichen, `Edm.DateTime` als UTC, `Accept` und die Versions-Kopfzeilen ohne Wirkung, Funktionsimport wie ein unbekanntes Set, Fehlerantworten im Format des Dienstes; ohne Kennung der Planung (`SPEC-200` Nr. 1); die Aussagen belegt der Rundlauf (B20).
5. Das gebundene Beispiel zu V2 im Handbuch (Start mit `MODE=hybrid` und `SERVICE_ODATA_VERSION=2.0`, Ausgangswert 1, in der Fehlerausgabe die Zeile `ODR-E2001 MODE: Modus hybrid für Version 2.0 nicht unterstützt`) — **bedingt:** das Format von `SPEC-200` kennt heute keine Umgebung (Nr. 6 d); liegt `slice-handbuch-mock` (liefert die Umgebung im Beispiel-Format, nimmt die Sendung nicht an) beim Start von Teil B in `done/`, trägt Teil B das Beispiel; sonst bleibt das Kapitel Prosa und Tabelle, und das Beispiel ist ein akzeptiertes Negativ mit Grund (Format ohne Umgebung) und Folge (die Aussage zu `hybrid` mit V2 ist nur ein Test des Roots belegt, A07; keine Wiederkehr zu erwarten, weil jedes Kapitel mit Umgebung dasselbe Negativ trägt, `slice-listener-tls-verdrahtung`). Es entsteht kein Eintrag im Register und kein Folge-Slice.

**Der Referenzdienst V2** (Test-Server, Auftrag 6; Port-Fake im Sinne von `SPEC-128`; eine Datei `tests/integration/v2_referenzdienst_test.go`). Er ist ein `httptest`-Server mit festen Bytes und **importiert keines der Pakete `csdlv2`, `jsoncodecv2`, `urlparserv2`**: das Orakel ist nicht der Code (B19; `AGENTS.md` §3.15 „Eine Fixture gleicht nie dem Wert der Mutation“). Der Dienst liegt unter `/odata`; er zählt Verbindungen und Requests wie der Referenzdienst von V4 und liefert JSON nur, wenn `Accept` `application/json` nennt oder `$format=json` in der Query steht, sonst Atom.

| Nr. | Anfrage an den Referenzdienst (Aufnahme-Folge) | Antwort | Ausgang beim Import (`SPEC-127` Nr. 8) |
|---|---|---|---|
| E1 | `GET /odata/$metadata` | EDMX 1.0 (CSDL 2008/09), `Content-Type: application/xml;charset=utf-8`: Entity-Typ `Product` mit `Id` (Int32, Schlüssel), `Name` (String, nullbar), `Active` (Boolean), `Qty` (Int64), `Price` (Decimal), `Created` (DateTime), `Ref` (Guid) und `Weight` (**`Edm.Double`**, nicht unterstützt), Sets `Products` und `Archive` (beide `Product`), dazu sieben Typen `KString`, `KBoolean`, `KInt32`, `KInt64`, `KDecimal`, `KGuid`, `KDateTime` mit je einem Schlüssel des Typs im Namen und `Name`, je ein Set gleichen Namens | gelesen (Metadaten) |
| E2 | `GET /odata/Products`, `Accept: application/json` | `{"d":{"results":[<drei Entities mit __metadata, alle Properties inklusive Weight>]}}`, `application/json;charset=utf-8`, `DataServiceVersion: 2.0` | gelesen (Collection) |
| E3 | `GET /odata/Products(1)`, `Accept: application/json` | `{"d":{…}}` | gelesen (Entity) |
| E4 | `GET /odata/Products` ohne `Accept` | Atom-Feed, `application/atom+xml;charset=utf-8` | übersprungen: `kein_json` |
| E5 | `GET /odata/Products?$inlinecount=allpages`, `Accept: application/json` | `{"d":{"__count":"3","results":[…]}}` | übersprungen: `query` |
| E6 | `GET /odata/Archive`, `Accept: application/json` | `{"d":{"results":[]}}` | gelesen (Collection, leer) |
| E7 bis E13 | je `GET /odata/K<Typ>(<Literal der URL>)`, `Accept: application/json` | `{"d":{…}}` mit dem Schlüssel des Typs | gelesen (Entity) ×7 |

Zählprüfung gegen die Quelle: 13 Austausche; gelesen 11 (Metadaten 1, Collection 2, Entity 8), übersprungen 2 (`kein_json` 1, `query` 1), Summe 13 (`SPEC-127` Nr. 14). Die drei Entities von `Products` tragen die sieben Typen in der Schreibweise von V2 (`Edm.Int64` und `Edm.Decimal` als Zeichenkette, `Edm.DateTime` als `\/Date(<ms>)\/`, eine nullbare Property mit `null` in einer Entity); die URL-Literale der Schlüssel (E7 bis E13) sind je Typ mindestens eines der Formen aus `SPEC-173` Nr. 3 (Menge 11 Formen: String `'a''b'`, Boolean `true`, Int32 `1`, Int64 `7` und `7L`, Decimal `1.5` und `1.5M`, Guid `guid'<36 Zeichen>'`, DateTime `datetime'2020-01-02T03:04:05'`, dazu `l`- und `m`-Suffix klein: `7l`, `1.5m`; Zählprüfung 11).

**Tabelle *Zusage · Fänger* (Teil B).** Mengen als Generatoren; **Ursachen** = {`format`, `methode`, `pfad`, `option`, `filter`, `literal_typ`} (6, `SPEC-174` Nr. 2); **Antwortarten** = {`$metadata`, Collection, Entity, Anzahl (`/$count`), Fehler `ODR-E6006`} (5).

| Nr. B | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| B01 | Aufzeichnen gegen den Referenzdienst V2 schreibt jeden der 13 Austausche; die Bytes von Request und Response in den Blobs gleichen den Konstanten der Fixture, unabhängig von der Version (`SPEC-170` Nr. 6, Modus `record`) | `TestV2Aufzeichnung`: Schleife über E1 bis E13, Zählprüfung 13; Bytes aus den Konstanten | Root lässt den V2-Dienst aus `record.Dienste` (13 rot, `ODR-E5009`); Aufnahme verwirft Antworten mit `application/json;charset=utf-8` (11 rot: E2, E3, E5, E6, E7 bis E13) | andere |
| B02 | Beenden und Import: `Neu`; der Report zählt 11 gelesen (Metadaten 1, Collection 2, Entity 8) und 2 übersprungen (`kein_json` 1, `query` 1), Summe 13 (`SPEC-127` Nr. 8 und 14, `SPEC-175` Nr. 4 b und c) | `TestV2Import`: Zahlen aus der Fixture (Zählprüfung 13, 11, 2) | Import mit dem Satz `4.0` (`ODR-E6004`, alle rot); `kein_json` als gelesen gezählt (Fehler des Codecs); `$inlinecount` gelesen | andere |
| B03 | Coverage nach dem Import: `Products` ist `complete` (ungefilterte Abfrage ohne Folgelink), `Archive` ist `complete` (leere Kette), jeder der sieben Sätze `K<Typ>` ist `partial` (nur Einzelabfrage), je Set aus `SPEC-127` Nr. 12 abgeleitet und gegen den Report gelesen, nicht gegen den Code | `TestV2Coverage`: Schleife über 9 Sets, Zählprüfung 9 | Kette nach `__next` statt `@odata.nextLink` gesucht; Status für ein leeres Set `partial` | andere |
| B04 | Der Mock beantwortet im Modus `mock` nach Workspace `$metadata`, Collection, Entity und Anzahl mit Body, `Content-Type` und Kennzeichen der Version; Body, Typ und Kopfzeile gleichen den Konstanten (`SPEC-172` Nr. 6, `SPEC-198` Nr. 4, `SPEC-175` Nr. 7): `application/json;charset=utf-8` mit `DataServiceVersion: 2.0`, bei der Anzahl `text/plain;charset=utf-8` mit denselben Ziffern, bei `$metadata` das aufgezeichnete Dokument. Menge 12 Zellen: 4 Arten × {Body, `Content-Type`, Kennzeichen} | `TestV2Antwortarten`: Tabelle 4 × 3, Zählprüfung 4 und 3; Orakel Konstanten | Kennzeichen `OData-Version: 4.0` (4 rot); Typ `application/json;odata.metadata=minimal` (3 rot); Body in der Form von V4 (3 rot) | andere |
| B05 | Die sieben Typen stehen in Collection und Entity in der Schreibweise von V2; die nicht unterstützte Eigenschaft `Weight` erscheint nie (`SPEC-170` Nr. 4, `SPEC-172` Nr. 3 und 6 f). Menge 14: 7 Typen × {Wert, `null` bei nullbarer Property}, dazu `Weight` in beiden Antwortarten | `TestV2Typen`: Schleife über 7 und 2, Zählprüfung 14; Orakel Konstanten | `Edm.Int64` als Zahl (2 rot); `Edm.DateTime` ohne Hülle `Date` (2 rot); `Weight` geschrieben (2 rot) | andere |
| B06 | Standardabfrage nach dem Umfang: `$filter`, `$orderby`, `$top`, `$skip`, `$select`, `$inlinecount` (`allpages` mit `__count` als Zeichenkette, `none`) ergeben die erwartete Menge in der erwarteten Reihenfolge; `$filter` mit `eq`, `ne`, `gt`, `ge`, `lt`, `le`, `and`, `or`, `not`, Klammern (`SPEC-170` Nr. 2). Menge 16 Anfragen: 6 Optionen und 10 Filterformen | `TestV2Abfragen`: Schleife, Zählprüfung 6 und 10; Orakel aus der Fixture | `$skip` ignoriert (1 rot); `__count` als Zahl (1 rot); `lt` und `le` vertauscht (2 rot); `not` ohne Wirkung (1 rot) | andere |
| B07 | Ein leeres Ergebnis bleibt im V2-Format: `{"d":{"results":[]}}`, mit `$inlinecount=allpages` `{"d":{"__count":"<n>","results":[]}}` mit der Gesamtzahl ohne `$top` und `$skip`, nie `null`, nie ein fehlendes `results` (`SPEC-172` Nr. 6 c und l). Menge 8: Wege {Filter ohne Treffer (n = 0), `$skip` über das Ende (n = 3), `$top=0` (n = 3), leeres Set `Archive` (n = 0)} × {ohne, mit `$inlinecount`} | `TestV2Leer`: Schleife, Zählprüfung 4 und 2 | `results` fehlt (8 rot); `__count` zählt die Entities der Seite (2 rot: `$skip`, `$top=0`); `null` statt `[]` (8 rot) | andere |
| B08 | Eine Anfrage außerhalb des Umfangs ist `501` mit `ODR-E6006`, ohne einen Wert der Anfrage im Body, mit `X-Correlation-Id`; der Zustand des Workspace (Entities, Revisionen, Coverage) bleibt (`SPEC-174` Nr. 1 und 2, `SPEC-170` Nr. 3). Menge 12: 6 Ursachen × 2 Repräsentanten (`$format=atom`, `$format=xml`; `POST`, `MERGE`; `/` und `/Products(1)/Category`; `$expand=Category`, `$skiptoken=x`; `substringof('a',Name)`, `Price add 1 gt 2`; `Price gt 1.5f`, `Id eq X'AA'`) | `TestV2Kompat`: Schleife, Zählprüfung 6 und 2; Status, Code, Body, Zustand vor und nach | V2-Dienst mit dem Satz `4.0` (`ODR-E6001`/`ODR-E6005`, alle rot); Antwort `404` statt `501` (12 rot); Zustand durch `POST` geändert (2 rot) | andere |
| B09 | Jede Methode außer `GET` und `HEAD` ist an einem V2-Dienst `ODR-E6006` mit der Ursache `methode`, der Body wird nicht gelesen, nichts wird geschrieben; `HEAD` ist `GET` ohne Body (`SPEC-174` Nr. 3, `SPEC-195` Nr. 7, `SPEC-192` Nr. 3). Menge 24: {`POST`, `PUT`, `PATCH`, `MERGE`, `DELETE`, `OPTIONS`} × {Collection, Entity, `/$count`, `$metadata`}; dazu `HEAD` an den vier | `TestV2Methoden`: Schleife, Zählprüfung 6, 4 und 4; Status, Code, Zustand | `NurLesend` falsch im Satz (`ODR-E6002` oder eine Ausführung, 24 rot); `HEAD` mit Body (4 rot) | andere |
| B10 | Ursache und Ort des Fehlers je Fall und die Rangfolge der Paare am Port des Mock im Root: `errors.As` auf `*model.KompatFehler`; Menge 12 Fälle aus B08 und 4 Paare: `ODR-E5003` vor Kompatibilitätsfehler, `ODR-E6005` vor Kompatibilitätsfehler, Kompatibilitätsfehler des URL-Lesers vor `methode`, `methode` vor `ODR-E7001` (`SPEC-174` Nr. 4, `SPEC-195` Nr. 3) | `TestV2Ursache` in `src/bootstrap`: der Haken `NeuerMock` umhüllt `mock.New`, hält jeden Fehler fest (Port-Fake, zählt jeden Aufruf), Schleife, Zählprüfung 12 und 4 | `option` und `filter` vertauscht (4 rot); `methode` hinter dem Lesen des Workspace (1 Paar rot); `ODR-E6005` hinter `ODR-E6006` (1 Paar rot) | andere |
| B11 | Der V4-Pfad ist byte-gleich: ein V4-Dienst neben einem V2-Dienst im selben Prozess antwortet wie ein V4-Dienst allein, ein V2-Dienst wie ein V2-Dienst allein, mit der Kopfzeile seiner Version (`OData-Version: 4.0` bzw. `DataServiceVersion: 2.0`); kein Service beeinflusst den anderen (`SPEC-175` Nr. 6, `LH-FA-17.a`). Menge 12: 6 Antwortarten (`$metadata`, Collection, Entity, Anzahl, `$filter`, Fehler `ODR-E6008`) × 2 Dienste | `TestV2NebenV4`: ein Prozess mit `SERVICES_FILE` und zwei Upstreams, Import beider, zwei Vergleichsläufe mit je einem Dienst; Zählprüfung 6 und 2 | ein Satz für beide Dienste, der letzte gewinnt (12 rot); `ETAG_ENABLED` für den V2-Dienst wirksam (1 rot); Metadaten-Cache je Prozess | andere |
| B12 | Die Versions-Kopfzeilen der Anfrage ändern nichts: dieselbe Antwort (Status, Header außer `X-Correlation-Id`, Body) mit und ohne Kopfzeile (`SPEC-174` Nr. 6, `SPEC-173` Nr. 7). Menge 85: {`DataServiceVersion`, `MaxDataServiceVersion`, `MinDataServiceVersion`, `OData-Version`} × {`2.0`, `4.0`, `1.0`, `3.0`} und alle vier zugleich (17 Formen) × 5 Antwortarten | `TestV2Kopfzeilen`: Schleife, Zählprüfung 4, 4, 17 und 5 | `MaxDataServiceVersion: 1.0` ändert die Form (5 rot); die Antwort trägt die Version der Anfrage (5 rot); je Name eine Mutation | andere |
| B13 | `Accept` ändert nichts: auch `application/atom+xml` ergibt JSON (`SPEC-174` Nr. 5). Menge 15: {fehlt, `application/json`, `application/atom+xml`, `application/xml`, `*/*`} × {Collection, Entity, Anzahl} | `TestV2Accept`: Schleife, Zählprüfung 5 und 3 | Atom bei `Accept: application/atom+xml` (3 rot) | andere |
| B14 | Das Schlüsselliteral des V2-Payload-Lesers gleicht dem des V2-URL-Lesers: der Mock findet jede importierte Entity über `GET /<Set>(<Literal>)` für jeden der sieben Typen (`SPEC-172` Nr. 5, `SPEC-173` Nr. 3). Menge 11 Formen der Literale über 7 Typen | `TestV2SchluesselLiteral`: Schleife über die 11 Formen, Zählprüfung 7 und 11; Status 200 und die Entity | Leser schreibt die Guid in Großbuchstaben (`404`, die Form `guid'…'` rot); Zeitpunkt ohne Normalisierung auf UTC (die Form `datetime'…'` rot); Suffix `L` nicht gelesen (`7L` rot) | andere |
| B15 | Ein Dokument anderer Version als der deklarierten ist `ODR-E6004`, nichts ist geschrieben: ein V2-Dokument unter einem Service `4.0` und ein V4-Dokument unter einem Service `2.0` (`SPEC-175` Nr. 4 d, `SPEC-171` Nr. 2). Menge 2 Richtungen | `TestV2Kreuzdokument`: Schleife über 2; Zustand vor und nach (kein Import, keine Baseline, `metadata_blob` leer) | Prüfung der Version am Leser entfernt (2 rot); der Importer nimmt den Satz eines anderen Dienstes | andere |
| B16 | Replay: ein V2-Service im Modus `replay` beantwortet jeden der 13 Austausche byte-gleich, auch den Atom-Austausch (`SPEC-170` Nr. 6, `SPEC-175` Nr. 5). Menge 13 | `TestV2Replay`: Schleife über E1 bis E13, Zählprüfung 13 | Root verdrahtet `replay` für `2.0` nicht (13 rot); Atom wird verworfen (1 rot) | andere |
| B17 | Eine Antwort des Mock für einen Dienst `4.0` bleibt, wie sie war: die Tests von `welle-mock` laufen unverändert grün (`LH-FA-17.a`, Risiko des Plans) | die Suite des Bestands ohne Änderung, `make test` | Satz `2.0` für einen V4-Dienst (die Suite rot) | andere |
| B18 | Gleichzeitige Anfragen an einen V2- und einen V4-Dienst desselben Prozesses beeinflussen sich nicht: jede Antwort gleicht der Einzelantwort, kein Datenrennen (`SPEC-175` Nr. 6; die Implementierungen haben keinen Zustand, `SPEC-125` Nr. 14). Menge: 16 Goroutinen (je 8 je Dienst) × 5 Antwortarten, Start an einer Schranke | `TestV2Gleichzeitig`: Wirkungs-Fänger (jede Antwort gegen die Einzelantwort, Zählprüfung 16 und 5) und `make test-race` je Prozess (`-test.count=1`) | der Root hält den zuletzt gewählten Satz in einem Feld des Handle und nutzt ihn je Anfrage (Wirkung rot in jedem Lauf; der Detektor meldet das Rennen mit einer Rate, die der Bericht je Prozesslauf nennt); Tabelle der Sätze beim ersten Zugriff gefüllt | nebenläufig |
| B19 | Das Orakel ist nicht der Code: die Dateien des Referenzdienstes und der Erwartungen importieren keines der Pakete `csdlv2`, `jsoncodecv2`, `urlparserv2` (`AGENTS.md` §3.15, Fixture gleicht nie dem Wert der Mutation) | `TestV2OrakelOhneCode`: AST über die Testdateien von Teil B | ein Test baut eine Erwartung mit `jsoncodecv2.NewSchreiber()` (rot) | andere |
| B20 | Das Handbuch führt V2 als Kapitel (Inhalt oben, Sendung 4): Ist-Zustand, ohne Kennung der Planung, die Aussagen belegt der Rundlauf; das Beispiel nach Sendung 5, wenn die Bedingung gilt; die Kataloge bleiben grün | `make handbuch-check` und `make katalog-check`, Ausgabe `ok` (benannte Ausgabe) | Kapitel nennt `SPEC-175` (Befund der Menge von Kennungen); Aussage über Geplantes (Befund) | andere |

**Raster *Operation × Argumentart* (Teil B).** Der Dienst `Mock` an einem V2-Dienst; jede Zelle nennt ihren Ort oder „gilt nicht“ mit Grund.

| Operation | Methode außerhalb | Kopfzeile der Version | `Accept` | Ressource, Option, Filter, Literal, Format außerhalb | leeres Ergebnis | Funktionsimport | Abbruch, Wiederholung |
|---|---|---|---|---|---|---|---|
| Anfrage lesen (URL, Methode) | `ODR-E6006` `methode`, nach dem URL-Leser, vor Workspace (`SPEC-174` Nr. 3, `SPEC-195` Nr. 7) | ohne Wirkung (`SPEC-174` Nr. 6) | ohne Wirkung (`SPEC-174` Nr. 5) | `ODR-E6006` mit Ursache (`SPEC-174` Nr. 2, `SPEC-170` Nr. 3) | gilt nicht: die Antwort kennt keinen Fehler | wie ein unbekanntes Set, akzeptiertes Negativ (`SPEC-174` Nr. 7) | Abbruch: `SPEC-195` Nr. 2 (Kontextfehler ohne Code); Wiederholung: die Anfrage hat keinen Zustand (`SPEC-195`, `SPEC-192` Nr. 7 für Mutationen, hier nie ausgeführt) |
| Antwort schreiben (Schreiber des Satzes) | gilt nicht: nichts wird geschrieben | gilt nicht: der Schreiber sieht keine Kopfzeile (`SPEC-172` Nr. 6 g) | gilt nicht (`SPEC-172` Nr. 6 m) | gilt nicht: der Fehler entsteht vorher | `{"d":{"results":[]}}` (`SPEC-172` Nr. 6 c) | gilt nicht: kein Set | Abbruch: `SPEC-168` Nr. 5; Wiederholung: gleiche Bytes (`SPEC-172` Nr. 6) |
| Import, Aufzeichnung lesen | gilt nicht: der Import liest nur `GET` (`SPEC-127` Nr. 8 a) | gilt nicht: der Import liest keine Anfrage-Kopfzeile außer `Accept` im Key | Atom: `kein_json` (`SPEC-127` Nr. 8 g, `SPEC-175` Nr. 4 b) | `$inlinecount`: `query` (`SPEC-175` Nr. 4 c); Property außerhalb: übersprungen mit Grund `typ` (`SPEC-172` Nr. 3) | leere Collection: gelesen (`SPEC-172` Nr. 2 a, `SPEC-127` Nr. 12) | gilt nicht: der Pfad ist kein Set (`SPEC-127` Nr. 8 b) | Abbruch: `SPEC-127` Nr. 2; Wiederholung: `SPEC-127` Nr. 5 |

**Zusatztabelle (Teil B).**

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| V2-Service in `record` | wie V4, von der Version unabhängig (`SPEC-170` Nr. 6) | B01 |
| V2-Service in `replay` | wie V4 (`SPEC-170` Nr. 6) | B16 |
| V2-Service in `mock` | `SPEC-170` bis `SPEC-174`; Methode außer `GET` und `HEAD`: `ODR-E6006` `methode` (`SPEC-174` Nr. 3) | B04 bis B10 |
| V4-Service neben einem V2-Service | unverändert (`SPEC-175` Nr. 6) | B11, B18 |
| Kopfzeilen und `Accept` | ohne Wirkung (`SPEC-174` Nr. 5 und 6) | B12, B13 |
| Funktionsimport eines V2-Dienstes im Mock | unbekanntes Set statt `ODR-E6006` (`SPEC-174` Nr. 7, akzeptiertes Negativ) | benannt, kein Fänger: der Referenzdienst führt keinen Funktionsimport (`csdlv2` meldet ihn als Fund `konstrukt`, `SPEC-171` Nr. 4); Folge: der Client sieht den Fehler eines unbekannten Sets; Maß: 1 von 6 Ursachenklassen des Rasters |

**DoD von Teil B (Entwurf für den Planner, drei Liefer-Punkte):** (1) der Rundlauf in `make test`: B01 bis B19 mit je einer roten Mutation, der Bericht nennt je Zusage Zusage · Mutation · roter Test und die Laufzeit des Rundlaufs (Risiko *Laufzeit*, `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`); (2) das Kapitel zu V2 im Handbuch (B20, Sendungen 4 und 5); (3) die Sendungen 1 bis 3 in den Fängern B04, B06, B09, B12, B14, B17. Dazu die üblichen Punkte (`make gates`, Review, Closure, Register, Risiken, Paarungen). Die Zeit des Rundlaufs nennt der Bericht gemessen, mit der Last des Hosts vor und nach jedem Lauf (`AGENTS.md` §3.10).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-v2-import-codecs`, `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-v2-payload-schreiben`, `slice-v2-query` und `slice-v2-query-filter` liegen in `done/`, `welle-mock` liegt in `done/` (Trigger der Welle), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` (erfüllt). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-110` Nr. 17, `SPEC-139`, `SPEC-170` bis `SPEC-177`), die Tabelle *Zusage · Fänger* und das Raster stehen in §3 und §6 (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** die V2-Slices davor (die Adapter und der Import, die der Root verdrahtet; `slice-v2-query-filter` liefert `NewFilterLeser`, den der Satz für `WithFilter` braucht), `slice-bootstrap-m2` (der Composition Root, in dem Record, Import und die Anlage der Services verdrahtet sind, `SPEC-107` Nr. 14), `slice-bootstrap-m3` (der Modus `mock` im Root, `satzVersion4`, `Dienst.Etag`), `slice-handbuch-meldungskatalog` und `slice-handbuch-katalog-schluessel-kommandos` (Kataloge-Gate und Abgleich der Schlüssel, die die Zeilen von DoD Punkt 3 verlangen). Der Slice ist **nicht** vor `welle-mock` und nicht ohne M2 startbar.
- **Was der Slice nicht braucht:** `welle-etag`, `welle-expand-batch`, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`; die Reihenfolge nach den Muss-Wellen ist Priorisierung.
- **Was nach diesem Slice kommt:** `slice-v2-rundlauf` (Teil B, §3) startet, wenn dieser Slice in `done/` liegt.
- **Größenmarke** (`AGENTS.md` §3.14; vom Architect nach `SPEC-128` Nr. 1 aus der Tabelle gerechnet, Planner setzt `basis`, `schaetzung`, `marke` vor dem ersten Code-Commit): 22 Zeilen der Klasse `andere`, keine nebenläufige, ein Fake (`protDienste`), Produktcode erwartet 160 Zeilen (`konfig_version.go` rund 55, `version.go` rund 45, `wurzel.go` rund 30, `konfig_dienste.go` rund 13, `lebenszyklus.go` und `konfig.go` rund 8, `Dienst.Version` und Kommentare): *S* = 85·0 + 80·22 + 200·1 + 160 = 2120, *M* = 100 · ⌈135·2120 / 10000⌉ = 2900; keine Schnittfrage (*S* < 2223, 103 darunter). Das Handbuch zählt nicht (`SPEC-128` Nr. 3 m). Der Slice liest keine fremden Bytes im Sinne eines Formats; die Konfiguration ist Text des Betreibers.

```groessenmarke
pfad: src/bootstrap/konfig.go
pfad: src/bootstrap/konfig_dienste.go
pfad: src/bootstrap/konfig_version.go
pfad: src/bootstrap/version.go
pfad: src/bootstrap/wurzel.go
pfad: src/bootstrap/lebenszyklus.go
pfad: src/bootstrap/konfig_version_test.go
pfad: src/bootstrap/version_wahl_test.go
pfad: src/bootstrap/version_anlage_test.go
pfad: src/bootstrap/haken_test.go
pfad: src/bootstrap/export_test.go
produktcode: 160
fakes: 1
```

**Rückführungen — vorab benannt, nicht erst im Nachhinein begründet:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Der Schnitt vor dem Code ist gemacht (Teil B); ein weiterer Schnitt ist nicht vorgesehen, weil Konfiguration, Anlage und Wahl denselben Wert `Dienst.Version` teilen (ein Satz ohne Konfiguration oder eine Konfiguration ohne Satz lieferte einen V2-Dienst mit den Implementierungen der Version 4, die Vermischung, die `LH-FA-17.a` verbietet).
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`, oder `slice-v2-import-codecs` liegt nicht in `done/`.

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

- Der Import eines V2-Dienstes braucht mehr als einen Codec je Version: Auswahl der Antworten, Ketten und Folgelinks sind V4-Regeln in `done/` (`SPEC-127` Nr. 8 und 12) — **Ausgang:** eingetreten: `slice-v2-import-codecs` (`SPEC-175` Nr. 4: derselbe Importer, Satz je Dienst).
- Die Tests von `welle-mock` müssen unverändert grün laufen (`welle-odata-v2` §3); die 37 Stellen, die einen `Dienst` ohne Version bilden, bleiben textgleich (A11) — **Ausgang:** offen; Fänger A11 und die Suite des Bestands; die byte-gleichen V4-Antworten neben V2 liegen im Rundlauf (Teil B, B11).
- Mehrere Services verschiedener Versionen im selben Prozess dürfen sich nicht beeinflussen (Roots, Metadaten, Caches) — **Ausgang:** offen; hier die Wahl je Dienst (A13, A16, A19, A20, A21), im Prozess Teil B (B11, B18).
- Der Port-Fake `protDienste` hält jedes Argument fest und zählt jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Der Rundlauf ist ein Test mit Prozessen, Fristen und einem Referenzdienst; er erbt die Zeitgrenzen von `make test` und `make gegenproben` — **Ausgang:** entschieden: ein Risiko von Teil B (der Zuschnitt nennt es in der DoD dort, `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster steht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-22): alle Zellen gefüllt; offen ist allein, ob der Implementer eine ungenannte Randform findet (Rückgabe an den Architect).
- **Rückgabe an den Planner (Zuschnitt, `AGENTS.md` §3.13):** (1) Plan `slice-v2-rundlauf` aus dem Abschnitt *Zuschnitt Teil B* in §3 anlegen und den Abschnitt danach hier streichen; (2) die Sendungen der Pläne in `done/` an `slice-v2-bootstrap` umadressieren: Rundlauf, Referenzdienst, Fänger der Kopfzeilen, Schlüsselliteral-Paar und das Kapitel zu V2 gehen an `slice-v2-rundlauf`, Verdrahtung, Konfiguration und Anlage bleiben hier; die Zeilen *Slices in dieser Welle* und *Abhängigkeiten* von `welle-odata-v2` (§4, §5), die Roadmap und `rueckverfolgung.md` tragen den neuen Slice; (3) den Satz „Offene Rückgabe an den Architect: `SPEC-110` Nr. 9“ in `welle-odata-v2` §5 streichen: erledigt, `SPEC-110` Nr. 9 und 17 führen `SERVICE_ODATA_VERSION`; (4) die Größenmarke beider Teile setzen (Block oben; Teil B ohne Produktcode, zwei Fakes). — **Ausgang:** offen; Adresse: Planner.
- Das Handbuch zu V2 ist zwischen diesem Slice und Teil B nur als Zeilen der Tabellen vorhanden (der Slice ändert das Binary, belegt es aber erst im Rundlauf) — **Ausgang:** akzeptiertes Negativ: Grund: die Aussagen des Kapitels belegt erst der Rundlauf (`SPEC-200` Nr. 1), und das Kataloge-Gate verlangt nur die Zeilen; Folge: ein Anwender, der V2 zwischen den zwei Slices nutzt, findet nur die Zeile des Schlüssels und die des Fehlercodes; Maß: 2 von 3 Teilen der Handbuch-Pflicht (Zeilen) hier, das Kapitel in Teil B.
- `SPEC-201` Nr. 1 nennt `SERVICE_ODATA_VERSION` als Beispiel für einen Namen der Spezifikation ohne Leser, und `tools/katalog` führt ihn in seinen Fixtures; mit dem Leser trifft das Beispiel das Repo nicht mehr — **Ausgang:** akzeptiertes Negativ: Grund: die Fixtures bilden eigene Bäume und prüfen kein Repo; Folge: ein Satz des Vertrags nennt einen Fall, den es im Repo nicht mehr gibt, ohne Wirkung auf das Gate; keine Wiederkehr, der Satz ist ein Beispiel (Vertrag und Werkzeug anzupassen ist ein Slice der Gate-Kette, `slice-gegenproben-eichung` oder ein späterer, und es entsteht keine eigene Folgepflicht).

**Raster *Operation × Argumentart* (Teil A)** (alle Zellen vom Architect gefüllt; „gilt nicht“ steht mit Grund):

| Operation | Version fehlt | Version leer | unbekannte oder fremde Version | Widerspruch (zum gespeicherten Service, zum Modus) | Services verschiedener Versionen nebeneinander | Abbruch (Signal im Start), Wiederholung (Neustart) |
|---|---|---|---|---|---|---|
| Prüfung der Startkonfiguration, Einzelschlüssel `SERVICE_ODATA_VERSION` | Default `4.0` (`SPEC-110` Nr. 17 a, `SPEC-175` Nr. 1) | wie fehlend (`SPEC-110` Nr. 1, 17 a) | `ODR-E2001`, Grund `Wert nicht in der Aufzählung: 4.0, 2.0` (`SPEC-110` Nr. 17 a, Nr. 3) | zum Speicher: gilt nicht, die Prüfung kennt den Speicher nicht (die Anlage prüft); zum Modus: `SPEC-110` Nr. 17 b | gilt nicht: der Einzelschlüssel nennt genau einen Service (`SPEC-110` Nr. 9); zugleich mit `SERVICES_FILE`: `ODR-E2001` (`SPEC-110` Nr. 9, 17 a) | gilt nicht: die Prüfung hat keinen Kontext; Signale: `SPEC-138`, `SPEC-110` Nr. 13 |
| Prüfung der Startkonfiguration, Feld `odata_version` der `SERVICES_FILE` | Default `4.0` (`SPEC-110` Nr. 9, 17 a) | wie fehlend, `null` ebenso (`SPEC-110` Nr. 1, 9) | `ODR-E2001` mit Index des Eintrags (`SPEC-110` Nr. 17 a); falscher Typ: „muss eine Zeichenkette sein“ (`SPEC-110` Nr. 9) | zum Modus: `SPEC-110` Nr. 17 b, eine Zeile | zulässig, je Eintrag eine Version (`SPEC-175` Nr. 1 und 6, `SPEC-110` Nr. 17 a) | gilt nicht (wie oben) |
| Prüfung Version gegen Modus | Version `4.0`: kein Verstoß (`SPEC-110` Nr. 17 b) | wie fehlend | nur die Form, kein Verstoß am Modus (`SPEC-110` Nr. 17 b) | `hybrid` mit `2.0`: `ODR-E2001` an `MODE` (`SPEC-110` Nr. 17 b, `SPEC-175` Nr. 5); ungültiger `MODE`: keine Prüfung (`SPEC-110` Nr. 8) | eine Zeile, gleich wie viele Services `2.0` sind (`SPEC-110` Nr. 17 b) | gilt nicht (Prüfung ohne Kontext) |
| `Build` (von Hand gebaute Konfiguration) | gilt `4.0` (`SPEC-139` Nr. 2) | wie fehlend (`SPEC-139` Nr. 2) | `model.ErrUngueltig`, Ablage nie geöffnet (`SPEC-139` Nr. 2) | `hybrid` mit `2.0`: `model.ErrUngueltig` (`SPEC-139` Nr. 2) | je Dienst sein Satz (`SPEC-139` Nr. 11 b) | Abbruch: `Build` ruft nur Konstruktoren, `ctx` (`SPEC-139` Nr. 2, Ablage bei Fehler geschlossen); Wiederholung: gilt nicht, `Build` läuft einmal je `Handle` (`SPEC-139` Nr. 10) |
| Anlage des Service (`Hochfahren` Schritt 2) | `Create` mit `4.0` (`SPEC-139` Nr. 6, `SPEC-107` Nr. 14) | wie fehlend | gilt nicht: Prüfung und `Build` wiesen sie ab | Start verhindert, `ODR-E2001 odata_version: …` mit Logeintrag (`SPEC-107` Nr. 14, `SPEC-139` Nr. 6 und 8 c) | je Dienst eine Anlage mit seiner Version, in der Reihenfolge der Dienste (`SPEC-139` Nr. 6, `SPEC-175` Nr. 2 und 6) | Abbruch: `SPEC-107` Nr. 12, `SPEC-139` Nr. 6 (`ctx` vor jedem Schritt); Wiederholung: ein bekannter Service mit gleichen Werten tut nichts (`SPEC-107` Nr. 14) |
| Wahl des Satzes im Mock | Satz `4.0` (`SPEC-175` Nr. 3, `SPEC-139` Nr. 11 b) | wie fehlend | gilt nicht: die Wahl kennt nur zwei Werte | gilt nicht: der Satz folgt der Konfiguration, die die Anlage gegen den Speicher bestätigt (`SPEC-139` Nr. 6) | je Dienst sein Satz, keine Kreuzung (`SPEC-175` Nr. 3 und 6, `SPEC-139` Nr. 11 b) | Wiederholung: die Wahl hat keinen Zustand (`SPEC-175` Nr. 3); Abbruch: gilt nicht, die Wahl ist eine Tabelle |
| Wahl des Satzes im Import | Satz `4.0` (`SPEC-139` Nr. 4 und 11 b) | wie fehlend | gilt nicht (wie oben) | gilt nicht (wie oben); ein Dokument anderer Version: `ODR-E6004` (`SPEC-175` Nr. 4 d, Teil B) | je Dienst sein Satz (`SPEC-127` Nr. 1, `SPEC-139` Nr. 4) | wie die Wahl im Mock |

**Zusatztabelle** (je Modus eine Zeile; `AGENTS.md` §3.16): V2-Dienst in `record`, `replay`, `mock` und `hybrid`; V4-Dienst unverändert; die Paare von Verstößen am selben Schlüssel.

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| V2-Service in `record` | Aufnahme von der Version unabhängig; der Import trägt den Satz `2.0` (`SPEC-170` Nr. 6, `SPEC-139` Nr. 4) | A19; am Binary B01, B02 |
| V2-Service in `replay` | wie V4, kein Satz (`SPEC-170` Nr. 6, `SPEC-139` Nr. 11 b); die Konfiguration nimmt ihn (`SPEC-110` Nr. 17 b) | A08, A19; am Binary B16 |
| V2-Service in `mock` | Satz `2.0`, `NurLesend` (`SPEC-139` Nr. 11 b, `SPEC-174` Nr. 3) | A16, A20; am Binary B04 bis B10 |
| V2-Service in `hybrid` | `ODR-E2001` an `MODE` beim Start (`SPEC-110` Nr. 17 b); in `Build` `model.ErrUngueltig` (`SPEC-139` Nr. 2) | A07, A12 |
| V4-Service neben einem V2-Service | unverändert (`SPEC-175` Nr. 6) | A13, A16, A19, A20; am Binary B11, B18 |
| `MODE` ungültig und `hybrid` mit `2.0` | nur die Form von `MODE` (`SPEC-110` Nr. 8, 17 b) | A08 |
| fremde Version und `hybrid` | nur die Form der Version (`SPEC-110` Nr. 17 b) | A09 |
| `SERVICE_ODATA_VERSION` fremd und `SERVICES_FILE` gesetzt | allein die Zeile der Doppelkonfiguration (`SPEC-110` Nr. 9, 17 a) | A06 |
| Feld falschen Typs und fremder Wert desselben Feldes | allein „muss eine Zeichenkette sein“ (`SPEC-110` Nr. 9) | A05 |
| `service_root` und `odata_version` weichen beide ab | zwei Zeilen und zwei Logeinträge, in der Reihenfolge `service_root`, `odata_version` (`SPEC-139` Nr. 6) | A14 |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für den Port-Fake `protDienste`; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) für den Rundlauf von Teil B. Der Slice ist nicht nebenläufig im Sinne einer Zusage über Sperren; die Nebenläufigkeit über zwei Dienste liegt in Teil B (B18). Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
