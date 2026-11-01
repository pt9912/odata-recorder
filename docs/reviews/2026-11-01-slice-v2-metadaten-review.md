# Review-Report: slice-v2-metadaten — 2026-11-01

**Review-Art:** Code — geprüft gegen Plan (`slice-v2-metadaten` §1, §3 mit der Tabelle *Zusage · Fänger* A1 bis J4 (57 Zeilen), dem Absatz *Mutationen vom Implementer rot gesehen*, den *Raten* und der *Zählprüfung*, §4 mit Block, Marke 8200 und Basis `b5ac4a5`, §6 mit Raster, Zusatztabelle und den drei Rückgaben), gegen `SPEC-171` Nr. 1 bis 10 samt *Ungeprüft*, `SPEC-170`, `SPEC-172` *Ungeprüft*, `SPEC-125` (Mechanik wortgleich), [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) und `AGENTS.md` §3.

**Gegenstand:** `slice-v2-metadaten` (Welle `welle-odata-v2`), `git diff b5ac4a5..dd68d95` (HEAD `dd68d95`, Zweig `slice-v2-metadaten`). Pfade: `src/adapters/driven/odata/csdlv2/` (Produktcode, 1492 Zeilen in acht Dateien), `src/adapters/driven/odata/versionstrennung_test.go`, `src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_*.go` (sechs Dateien) mit `testdata/v2_*.xml`, dazu `spec/spezifikation.md` und der Plan.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `dd68d95` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-01

**Methode.** Der Bericht des Implementers lag nicht vor; Tabelle, Raten und Zählprüfung sind aus Plan und Diff rekonstruiert und nicht übernommen. Alle Mutationen laufen auf Kopien im Scratchpad (`/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/rev-v2m/`; im Worktree liegt nichts außer diesem Report), im Image `odata-recorder:race` (`--memory=4g --memory-swap=4g --cpus=4`, ohne Netz, unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container zugleich). Ein Treiber im Container wendet je Mutante genau eine Ersetzung an (die Zahl der Fundstellen ist geprüft), fährt `go test -count=1 -run 'TestVertrag/^(<Zeile>)_'` (bei J den jeweiligen Test) und stellt die Datei wieder her. Exit-Codes sind nicht maskiert; eine Mutante zählt als rot, wenn der Treiber ein benanntes `--- FAIL` der Zielzeile meldet, die Ausnahmen stehen in der Tabelle. Acht Mutanten bauten nicht und sind neu geschrieben; zwei Mutanten waren ohne Wirkung (Schreibfehler des Reviewers) und sind nicht gezählt; ein Lauf des Treibers wurde durch einen zweiten gestört (Setup-Fehler), die betroffenen Mutanten sind erneut gefahren.

**Eingangs-Kontext:**

- Slice-Plan `slice-v2-metadaten` in `dd68d95`
- `SPEC-170`, `SPEC-171` Nr. 1 bis 10 und *Ungeprüft*, `SPEC-172` *Ungeprüft*, `SPEC-125` Nr. 1 bis 14 und *Ungeprüft*, `SPEC-128`
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md)
- `LH-FA-17`, `LH-RB-01`
- `AGENTS.md` (Hard Rules, besonders §3.2, §3.7, §3.9 bis §3.18); Reviewer-Skill mit den Zeilen *Fänger führt nur eine Form der Eingabe*, *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*, *Akzeptiertes Negativ ohne Grund und Folge*

---

## Antworten auf die Fragen des Auftrags

**(1) Tabelle A1 bis J4 (57 Zeilen) und Raster, je Zeile mindestens eine eigene Mutation, gegen die Klasse.** 195 verschiedene Mutanten in 209 Läufen (zwölf Läufe wiederholen eine Mutante über die ganze Suite), dazu 7 Mutanten der Ratenmessung. Von den 195 sind **182 rot und 13 grün**. Von den 13 grünen sind 6 äquivalent, zu schwach oder in der falschen Zielzeile (A5, J3d, E1c, B4e, B5b, N5b; Einzelheiten in der Tabelle und in I-3), **7 sind Lücken des Beweises** (F-2 bis F-5). Jede der 57 Zeilen ist mit mindestens einer eigenen, roten Mutation aufgelöst (Tabelle unten); I1 bis I3 mit Rate in 30 Prozessen (Absatz *Raten*), J4 mit der benannten Ausgabe. Gegen die **Klasse** (`AGENTS.md` §3.17): Mutanten an Elementen, die der Plan nicht nannte, sind in der Tabelle mit „+“ gekennzeichnet; vollständig über die Menge aufgelöst sind D7 (9 von 9 Doppelungen), F4 (6 von 6 Referenzarten), G1 (14 von 14 Stellen), D2 (alle vier Referenzarten), N7 (6 von 6), teilweise D4 (5 von 8 Typen), D5 (4 von 21 Namen, 2 Familien), D6 (11 von 20 Pflichtattributen), D9 (5 von 7 Nachbarpaaren). Die Zählprüfung der Mengen stimmt mit dem Protokoll von `go test -v` überein (Absatz *Zählprüfung*). Ein Probe-Paket (17 496 Mutationen des Referenzdokuments: jedes Byte gelöscht oder durch `<`, `&`, `"`, `>`, `/`, `\xff`, `\x00`, `\xc3` ergänzt, je fünf Zerlegungen des Stroms) fand keine zerlegungsabhängige Abweichung und keinen Fehler ohne `MetadatenFehler`; ein Northwind-Dokument mit Fremdattributen, `ReferentialConstraint`, `FunctionImport` und zwei Schemas ergibt das Modell und die Funde, die `SPEC-171` verlangt.

**(2) Rasterzellen.** Alle Zellen des Rasters in §6 (5 Operationen × 8 Argumentarten) haben einen Ort in `SPEC-171` oder „gilt nicht“ mit Grund; die Zusatztabelle hat je Paar einen Fänger (C5, D9, N3/N7, E1, D6/D7). Aufgelöst mit Mutationen: C5 (Müll vor der Wurzel), D9 (fünf der sieben Nachbarpaare der Durchläufe von mir, zwei vom Plan), N3/N7, E1. Eine im Code entschiedene Zelle (`AGENTS.md` §3.16, MEDIUM) habe ich nicht gefunden; die beiden Rückgaben stehen mit Ausgang *offen* und Adresse in §6 (`AGENTS.md` §3.13).

**(3) Nebenläufige Zeilen I2 und I3 mit Rate je Mutation.** Siehe Absatz *Raten*; 30 Prozesse, `go test -c`-Binary (ohne und mit `-race`), `-test.count=1`, Docker `--cpus=4`, Messender: der Reviewer.

**(4) J4 (V4-Dateien byte-gleich).** `git diff --numstat --diff-filter=MDRT b5ac4a5..dd68d95` über `src/adapters/driven/odata/csdl`, `src/adapters/driven/odata/jsoncodec`, `src/adapters/driven/odata/urlparser`, `src/hexagon/model` und `src/hexagon/ports/outbound`: **leer**. Über das ganze Repo erscheinen unter `--diff-filter=MDRT` nur `docs/plan/planning/done/slice-v2-metadaten.md` (235/18) und `spec/spezifikation.md` (8/6); `src/` trägt 19 Hinzufügungen und keine Änderung; `.golangci.yml`, `.a-check.yml`, `Dockerfile`, `Makefile` und `harness/` sind unberührt.

**(5) `versionstrennung_test` (J2, J3) samt Gegenprobe.** Sieben Mutanten, sechs rot: V2-Leser nimmt die V4-Wurzel nicht als `version` (J2a), V4-Leser nimmt die V2-Wurzel nicht als `version` (J2b), V4-Leser akzeptiert `Version="1.0"` (J2c); `csdlv2` importiert `csdl` (J3a), `csdl` importiert `csdlv2` (J3b), der Scanner meldet nur eine Richtung (J3c, rot in der Gegenprobe `TestImportScannerErkenntVerstoss`). Grün: J3d (das neutrale Paket wird mitgescannt), äquivalent, weil das neutrale Paket der Gegenprobe selbst nichts importiert.

**(6) `make a-check`, Hexagon.** `make a-check`: 0 Befunde. `csdlv2` importiert nur `src/hexagon/model` und die Standardbibliothek, die Suite nur das Modell und `ports/outbound`. `.a-check.yml` kennt nur Schichten; Kanten zwischen Driven Adaptern verbietet es nicht (INFO I-2).

**(7) Größenmarke.** `make groessenmarke`: „slice-v2-metadaten.md: ok, Messung 4525 von Marke 8200“. Eigene Messung `git diff --numstat b5ac4a5..dd68d95` über die elf Pfade des Blocks: 4525 Einfügungen, 0 Löschungen, Summe 4525 (Produktcode `csdlv2` 1492 Zeilen gegen `produktcode: 1500`; Schätzung 6070, Marke 8200 = aufgerundet 1,35 × 6070). Das Ende liegt bei 0,75 × Schätzung und 55 % der Marke; ein Eintrag `ueberschreitung` ist nicht nötig.

**(8) Rohbytes (§3.18).** Gegengelesen auf Rohbytes (`LC_ALL=C grep`, `od -c`): die sechs Dateien `metadata_codec_v2_*.go`, die acht Dateien `csdlv2/*.go`, `versionstrennung_test.go`, `leser_test.go` und die drei `testdata/v2_*.xml`. Keine Folge aus Backslash, `u` und vier Hex-Ziffern als Text, kein BOM, kein NBSP, kein Zeilentrenner; die Fixtures sind reines ASCII. Nicht-ASCII steht als Buchstaben (`Größe`, `ä`, `٣`) und als Rohbytes `EF BF BD` (U+FFFD) in `csdlv2/lesen.go` (2 Zeilen, `od -c` bestätigt) und `metadata_codec_v2_lesen.go` (5 Zeilen); das ist der gewollte Wert, `make zeichen-check` ist grün (1107 Dateien, mit diesem Report), der V4-Leser schreibt dieselbe Form.

**(9) Kommentarregeln §3.7, §3.11.** Keine Kommentarzeile beschreibt eine verworfene Alternative oder abwesenden Text; kein `nolint`, kein `d-check:ignore`. Ein Kopf sagt mehr zu, als sein Test fängt: `zeileV2G4` nennt „ein Element vor seinen Attributen, `property_typ` an der Stelle seiner Property“, die beiden Mutanten dazu bleiben grün (F-3).

**(10) Lint, Format, Läufe.** `make lint`: Exit 0 (die Stufe `lint` kam aus dem Build-Cache, also mit demselben Inhalt schon einmal grün; der Inhalt von `src` ist `dd68d95`); `gofmt -l src`: leer; `go vet` über `src/adapters/driven/odata/...` und `src/hexagon/ports/outbound/...`: ohne Befund; `go test -race -count=1 ./src/adapters/driven/odata/...`: ok (`csdlv2` 70,5 s).

**(11) Die beiden Rückgaben an den Architect (nicht als Befund gezählt).**

- *`Target` eines Bindings.* Der Code bildet `<qualifizierter Container>/<Name>` (`aufloesen.go`, `bindings`). Das ist die **kleinste Lesart**: `SPEC-125` Nr. 11 sagt „`Target` ist immer `<qualifizierter Container>/<Name>`“, `SPEC-171` Nr. 8 übernimmt Nr. 11 mit vier Abweichungen, `Target` ist keine davon, und Nr. 6 („gleich dem `EntitySet` des `End`“) widerspricht dem nicht; „nur der Name“ wäre eine fünfte Abweichung. Fänger: vorhanden und mehrfach; `Target` ohne Container ist rot in **drei** Zeilen (F1, N4, D11), `Set` und `Target` vertauscht rot in N4 (Mutanten N4a, N4b).
- *`AssociationSet` und Rollen.* Der Code prüft die Rollen allein gegen `FromRole` und `ToRole` der verweisenden Navigation (`assocSetAufloesen`); ohne Navigation bleibt eine `AssociationSet` mit fremden Rollen gültig. Das ist der Wortlaut von `SPEC-171` Nr. 10 f („sobald eine Navigation auf sie verweist“) und die **kleinste Lesart**; die Zusage der Zeile N6 im Plan („deren `End`-Rolle nicht zur Association passt“) ist weiter gefasst als die Spezifikation. Fänger in beide Richtungen: die strengere Lesart (Rollen der `AssociationSet` auch ohne Navigation gegen die `Association` geprüft, Mutante N6x) ist rot in N5 (Fall „AssociationSet ohne Navigation, Rollen der Association nicht getragen“); die schwächere (nur die Mengen geprüft, nur eine Rolle oder keine) ist rot in N6 (Mutanten N6a, N6b, N6c).

---

## Aufgelöste Tabelle (eine Zeile je Zeile des Plans)

„+“ bezeichnet ein Element der Klasse, das die Tabelle des Plans nicht nannte. Rot heißt: die Zielzeile meldet ein benanntes `--- FAIL`.

| Zeile | Eigene Mutation(en) | Ergebnis |
|---|---|---|
| A1 | `grenze < 1` zu `< 2`; Argumentprüfung nur bei unbeendetem Kontext (Reihenfolge) | 2 von 2 rot |
| A2 | Kontextprüfung im Strom nur im ersten Read; Prüfung vor dem Auflösen entfernt; Kontextfehler als `struktur` | 3 von 3 rot |
| A3 | Fehler von `src` als `syntax`; Fehler nach vollständigem Dokument verschluckt | 2 von 2 rot |
| A4 | BOM-Randprüfung (EF) entfernt; UTF-16-Randprüfung (FF/FE) entfernt+; unvollständiges Mehrbytezeichen als Verstoß+; leerer Read als Ende | 4 von 4 rot |
| A5 | Grenze `grenze+2`; Wächter „kein Read nach Ende oder Fehler“ ganz entfernt; Wächter nur für `io.EOF` entfernt | erste zwei rot; die dritte **grün** auch in der ganzen Suite, äquivalent (der Decoder liest nach `io.EOF` nie wieder, I-3) |
| B1 | leeres Dokument gültig | rot |
| B2 | Grenze 0x20 zu 0x1F; U+FFFF gültig; U+FFFE gültig+; CR ungültig+; TAB ungültig+; U+007F ungültig+; echtes U+FFFD ungültig+; alles über U+FFFD ungültig+ | 8 von 8 rot |
| B3 | `DOCTYPE` zu `DOCTYP` | rot |
| B4 | Obergrenze DFFF (`<` statt `<=`); Untergrenze D800 (`>` statt `>=`); nur die erste Referenz je Wert; nur das erste Attribut je Element+; Referenz außerhalb der Wurzel nicht geprüft | 4 von 4 rot; die fünfte grün in B4, rot in B5, D10, G3 (andere Zielzeile) |
| B5 | Namensraum beim Attributvergleich ignoriert; Nachlauf nach der Wurzel erlaubt; NBSP als Leerraum+; „Deklaration nur am Anfang“ entfernt | 3 von 3 rot; die vierte grün in B5, rot in B2, D10, G3 (andere Zielzeile) |
| B6 | undeklariertes Präfix `edm` als EDM | rot |
| C1 | V4-Namensraum mit beliebigem Namen als `version`; Namensraum nicht verglichen; Name nicht verglichen | 3 von 3 rot |
| C2 | Präfixvergleich `1.0`; `TrimSpace`; `1.00` gültig+ | 3 von 3 rot |
| C3 | `2.0.1` gültig+; zweites `DataServices` bis 2 erlaubt; „keines“ nicht geprüft; Namensraum des Attributs ignoriert | 4 von 4 rot |
| C4 | Namensraum 2006/04 nicht akzeptiert+; 2007/05 nicht akzeptiert+; CSDL 3.0 nicht `version`; V4 nicht `version` | 4 von 4 rot |
| C5 | Prüfung „Müll vor der Wurzel“ entfernt | rot |
| D1 | `Abstract` nicht beachtet; geerbter Schlüssel nicht gezählt+ | 2 von 2 rot |
| D2 | `TrimSpace`; Ort `@BaseType`, `@Type`, `@EntityType`, `@Name` je ohne Zusatz+ (4); Präfix `Edm` statt `Edm.` | `TrimSpace` und vier Orte rot; **Präfix `Edm` grün in der ganzen Suite** (F-4) |
| D3 | Grenze 64 zu 65; `maxKette` 63+; Zyklus nur gegen den ersten Typ+; Reihenfolge der Fälle vertauscht+ | 4 von 4 rot |
| D4 | `Edm.Time`, `Edm.SByte`, `Edm.Binary`, `Edm.Double`, `Edm.DateTimeOffset(V2)` als Schlüssel erlaubt+; Prüfung des `/` entfernt | 6 von 6 rot |
| D5 | `Edm.Stream`, `Edm.GeometryCollection`, `Edm.GeographyMultiPolygon` bekannt+; Präfix `Edm.Geo` bekannt+; `Collection(…)` angenommen | 5 von 5 rot |
| D6 | Pflicht entfernt für `Relationship`, `AssociationSet.Association`, `SetEnd.EntitySet`, `EntitySet.EntityType`, `Property.Type`, `FromRole`, `ToRole`, `End.Type`, `End.Role`, `SetEnd.Role` und leeren `PropertyRef.Name`+ (11 von 20); `True` gültig; `maxName` 129; `maxNamensraum` 257; `Edm.X`-Präfix | 15 von 15 rot |
| D7 | neun Doppelungen je ohne Prüfung: Entity-Typ, Member, `EntitySet`, `PropertyRef`, `Key` (>2), Container, `Association`, `AssociationSet`, Alias-Widerspruch | 9 von 9 rot |
| D8 | Schlüssel in Ketten nicht geprüft; geerbter Name nur gegen Properties; geerbte Navigationen nicht gesammelt; nur das direkte Elternglied+ | 4 von 4 rot |
| D9 | fünf Nachbarpaare der acht Durchläufe vertauscht (basen/ketten, ketten/summe, vererbung/typen, typen/container, container/schluessel) | 5 von 5 rot; die ersten zwei durch den Speicherwächter der Suite („eine Schleife des Lesers wächst ohne Ende“, Zeile D9), nicht durch eine Erwartung (I-3) |
| D10 | Teilbericht bei Fehler im Auflösen; Teilbericht bei Fehler im Lesen+ | 2 von 2 rot |
| D11 | erstes Schema abgelehnt; ab dem dritten Schema abgelehnt+ | erste rot in D11; zweite rot nur in I1 (D11 führt 0 bis 2 Schemas, I1 sechs) |
| N1 | `FromRole == ToRole` zugelassen; Rolle `FromRole` nicht geprüft; `Target` aus dem F-`End` | 3 von 3 rot |
| N2 | `Collection` bei `!= "1"`; `Collection` auch bei `1`; `Target` ohne Auflösung des Alias | 3 von 3 rot |
| N3 | `0..*` gültig; nur das erste `End` geprüft; `1..*` gültig+ | erste zwei rot; **`1..*` grün in der ganzen Suite** (F-5) |
| N4 | `Set` und `Target` vertauscht; `Target` ohne Container; zweite Navigation übersprungen | 3 von 3 rot (die zweite in drei Zeilen) |
| N5 | Binding ohne Navigation erzeugt; strengere Lesart der Rollen (N6x) | 2 von 2 rot |
| N6 | nur das F-`End` auf `EntitySet` geprüft; Rolle F nicht geprüft; Rolle G nicht geprüft | 3 von 3 rot |
| N7 | drittes `End` der `Association` und der `AssociationSet` ignoriert (2); gleiche Rolle in beiden (2); zu wenige `End` nur bei 0 in beiden (2) | 6 von 6 rot |
| E1 | `>` zu `>=`; Überlauf bei `MaxInt64` (`rest < MaxInt64` entfernt); BOM nicht gezählt+ (zwei Formen) | 3 von 3 rot, eine schwächere Form des BOM-Falls (nur spätere Reads) grün in E1, rot in A4 (I-3) |
| E2 | Grenze `>=`; `maxTiefe` 65 | 2 von 2 rot |
| E3 | `>=`; nur Properties gezählt; `maxMembers` 65537 | 3 von 3 rot |
| F1 | `Version()` liefert `4.0` | rot |
| F2 | Schlüssel nicht geerbt; geerbte Navigationen ausgelassen | 2 von 2 rot |
| F3 | Schlüssel sortiert statt in Reihenfolge der `PropertyRef` | rot |
| F4 | je Referenzart ohne Auflösung des Alias: `Association` der `AssociationSet`, `Relationship`, `End`-`Type`, `BaseType`, `EntityType` des Sets, Property-`Type` | 6 von 6 rot |
| F5 | `Edm.Int16` auf `Edm.Int32` abgebildet+ (ein nicht unterstützter Typ erschiene als unterstützt) | rot |
| F6 | `Nullable`-Vorgabe falsch; `Nullable` immer wahr; `Open` immer wahr; `Namespaces()` ohne Kopie | 4 von 4 rot |
| G1 | `attributFunde` je Stelle entfernt: 14 Stellen (Edmx, DataServices, Schema, EntityType, Key, PropertyRef, Property, NavigationProperty, Association, End, EntityContainer, EntitySet, AssociationSet, End der AssociationSet; Property und Navigation getrennt); Ort im Container und im Schema vertauscht; Ort der Erweiterung ohne Namen des Kindes; dazu die Aufrufstelle im Konstrukt und sein Ort | 17 von 17 rot; die Aufrufstelle im **Konstrukt** und sein Ort **grün** in der ganzen Suite (F-2) |
| G2 | `Unicode` als Fund; `m:DataServiceVersion` als Fund+; `xml:`-Attribute als Fund | erste zwei rot; **`xml:` grün in der ganzen Suite** (F-2) |
| G3 | Bericht ohne Fund `nil` | rot |
| G4 | Funde nach Art gruppiert; Reihenfolge umgekehrt+; Attribute vor ihrem Konstrukt; `property_typ` hinter den Attributen seiner Property | die ersten zwei rot; **die letzten zwei grün** in der ganzen Suite (F-2, F-3) |
| G5 | deklarierter Typ als unterstützt (`ComplexType` wird `Edm.String`); `FunctionImport` als `EntitySet` | 2 von 2 rot |
| G6 | `attributFunde` am `End` der `Association` und am `End` der `AssociationSet` je entfernt | 2 von 2 rot |
| H1 | Ort ohne `strconv.Quote` (Kopie von `model`) | rot |
| I1 | Bindings, Typen, Namensräume je aus einer Map | siehe *Raten* |
| I2 | Zähler als Paketvariable (nur Datenrennen); Typenliste als Paketvariable | siehe *Raten* |
| I3 | gemeinsamer Fehlerzeiger; Funde über Paketvariable | siehe *Raten* |
| J1 | Feld `cache` | rot (`TestLeserOhneZustand`) |
| J2 | siehe (5) | 3 von 3 rot |
| J3 | siehe (5) | 3 von 4 rot, 1 äquivalent |
| J4 | siehe (4) | Ausgabe leer |

---

## Raten (I1; I2 und I3 nebenläufig)

Je Mutation 30 Prozesse, Binary mit `go test -c` (ohne und mit `-race`), `-test.run 'TestVertrag/^(<Zeile>)_' -test.count=1`, `--cpus=4`, Messender: der Reviewer. „rot“ zählt die Prozesse mit einer `--- FAIL`-Zeile der Zielzeile; „DATA RACE“ zählt die Prozesse mit einer Meldung des Detektors.

| Mutation | Zeile | ohne Detektor (rot von 30) | mit `-race` (rot von 30 / Prozesse mit DATA RACE) |
|---|---|---|---|
| keine (Basis) | I1, I2, I3 | 0, 0, 0 | 0, 0, 0 / kein DATA RACE |
| I1: Bindings je `AssociationSet` aus einer Map | I1 | 30 | nicht gemessen |
| I1: Typen aus einer Map | I1 | 30 | nicht gemessen |
| I1: Namensräume aus einer Map | I1 | **0**; in der ganzen Suite rot in A4, D11, F4 (INFO I-4) | nicht gemessen |
| I2: Zähler als Paketvariable (Datenrennen ohne Wirkung) | I2 | 0 | 30 / 30 |
| I2: Typenliste als Paketvariable, geschrieben in `basen`, gelesen in `ohneSchluessel` | I2 | 0 | 0 / **30**: der Prozess endet mit der Panik des Speicherwächters, ohne `--- FAIL`-Zeile (INFO I-5) |
| I3: gemeinsamer Fehlerzeiger im Paket | I3 | 30 | 30 / 30 |
| I3: Funde über eine Paketvariable (Zuweisung und Rückgabe in `defer`) | I3 | 0 | 30 / 30 |

I2 und I3 gelten nach `AGENTS.md` §3.10 nur als gefangen, wenn die Mutation in jedem Lauf rot wird: **mit dem Detektor von `make test-race` ist jede der vier Mutationen in 30 von 30 Prozessen rot** (DATA RACE in 30 von 30). Ohne Detektor fängt der Wirkungs-Fänger nur den gemeinsamen Fehlerzeiger (30 von 30); Zähler, Typenliste und Funde-Variable sind dort in 0 von 30 rot. Der Plan nennt als akzeptiertes Negativ die Form mit kurzem Fenster (Zuweisung und Rückgabe in aufeinanderfolgenden Anweisungen: bei mir wie im Plan 0 von 30); der Zähler ohne Wirkung und die Typenliste bei gleichförmiger Last (I-5) sind dieselbe Art Negativ, ihre Zusage trägt allein `make test-race`. Die Basis- und Map-Raten des Plans (0 und 30 von 30) habe ich bestätigt; die weiteren Raten des Plans stehen für andere Textstellen und sind von mir nicht wiederholt.

---

## Zählprüfung der Mengen

Das Protokoll von `go test -v ./src/adapters/driven/odata/csdlv2/` (Zeile `Z Zählprüfung der Mengen`) nennt: A1 30, B1 12, B2 112, B3 11, B4 305, C1 43, C2 13, C3 19, C4 7, C5 11, D1 26, D2 50, D3 36, D4 35, D5 62, D6 142, D7 18, D8 14, D9 60, D11 156, E1 33, E2 11, E3 7, G1 104, N1 36, N3 54, N6 29, N7 11 und alle Dokumente der Suite 1521; A4 fährt 1500 Dokumente bis 5000 Bytes mit 2 261 618 Zweiteilungen. Das stimmt mit den Zahlen in §3 des Plans überein (A1 30, B4 305, C1 43, D1 26, D2 50, D3 36, D4 35, D5 62, D6 142, D9 60, D11 156, G1 104, N1 36, N3 54, N6 29, 1521). Die Mindestzahlen im Test liegen bei fast allen Mengen auf dem Wert oder knapp darunter (D6 140 gegen 142, D7 17 gegen 18, C3 17 gegen 19); weit ist nur die Schranke „alle Dokumente der Suite“ (mindestens 1000 gegen 1521), sie ersetzt die Einzelschranken nicht.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Zu jeder `AssociationSet` und jeder Navigation-Property derselben `Association` entsteht ein Binding: die Zahl der Bindings wächst mit dem Produkt zweier Mengen, nicht mit der Größe des Dokuments. Gemessen (Scratchpad-Probe, ein Dokument aus n Typen mit je einer Navigation auf `Ns.R` und n `AssociationSet` auf `Ns.R`): n = 250 (74 KiB) ergibt 62 500 Bindings und 30 MiB, n = 1000 (299 KiB) 1 000 000 Bindings und 443 MiB, n = 2000 (599 KiB) 4 000 000 Bindings und 2121 MiB in 2,0 s; der Import übergibt `MAX_CAPTURE_BODY` (einstellbar bis 1 GiB) als `grenze`. Der Plan sagt im Raster (Auflösen der Assoziationen, Spalte „über Grenze“) „Bindings ohne Grenze, durch `grenze` begrenzt (`SPEC-125` Nr. 5)“, und `SPEC-125` Nr. 5 begründet die fehlende Grenze mit „Modell und Bericht wachsen linear mit dem Dokument“; für die Bindings von V2 trifft das nicht zu, und weder ein akzeptiertes Negativ mit richtigem Grund und richtiger Folge noch ein Fänger noch eine Grenze steht dort. Ausfallszenario: ein Dokument unter 1 MiB, das der Import liest, belegt über 2 GiB (quadratisch hochgerechnet bei 8 MiB ein Vielfaches des Speichers); der Prozess bricht ab. | `AGENTS.md` §3.15 (akzeptiertes Negativ ohne Grund und Folge); `SPEC-171` Nr. 6, `SPEC-125` Nr. 5 *Keine Grenze* | `src/adapters/driven/odata/csdlv2/aufloesen.go` · `for _, m := range d.navigationen(s.ziel) {`; Plan §6 Raster · „Bindings ohne Grenze, durch `grenze` begrenzt“ | ja — ein Test mit dem Dokument der Probe (n = 2000) und einer Grenze für Speicher oder Zahl der Bindings | Zusage der Grenze falsch begründet, Wachstum quadratisch |
| F-2 | MEDIUM | Zwei Zusagen von `SPEC-171` Nr. 4 haben keinen Fänger: (a) ein Attribut mit Namensraum an einem als `konstrukt` gemeldeten Element ist ein Fund `erweiterung` mit dem Ort des Konstrukts, (b) Attribute des Präfixes `xml` (`xml:lang`, `xml:space`) sind kein Fund (so `SPEC-125` Nr. 7, auf das Nr. 4 verweist). Die Mutanten „Aufruf von `attributFunde` im Konstrukt entfernt“, „Ort des Konstrukts durch den Ort des Elternelements ersetzt“, „Attribute vor ihrem Konstrukt gemeldet“ und „`xml:`-Attribute als Fund“ bleiben in **allen** Zeilen von `TestVertrag` grün. Die Form ist real: das Northwind-Beispiel der Probe mit `<FunctionImport … m:HttpMethod="GET">` meldet zwei Funde, `konstrukt` `FunctionImport` und `erweiterung` `{…/metadata}HttpMethod` am Ort des Funktionsimports. | `AGENTS.md` §3.15 (Zusage ohne Fänger) | `src/adapters/driven/odata/csdlv2/elemente.go` · `l.fund(model.FundKonstrukt, ort, se.Name.Local)` und `case a.Name.Space == "", a.Name.Space == "xmlns", a.Name.Space == nsXML:`; Suite · `metadata_codec_v2_funde.go` `g1FaelleV2` | ja — Mutation im Scratchpad | Fänger führt nur eine Form der Eingabe (Attribute an ausgewerteten, nicht an gemeldeten Elementen) |
| F-3 | MEDIUM | Die Zusage „`property_typ` an der Stelle seiner Property“ (G4) ist nicht gefangen, wenn die Property zugleich ein Attribut mit Namensraum trägt: der Fund `property_typ` darf hinter die Attributfunde derselben Property rücken, ohne dass eine Zeile rot wird (Mutante in der ganzen Suite grün). Das Dokument von G4 hat die Property mit nicht unterstütztem Typ (`Wide`) ohne Attribut; der Kopf von `zeileV2G4` sagt „ein Element vor seinen Attributen, `property_typ` an der Stelle seiner Property“ zu. | `AGENTS.md` §3.15; `AGENTS.md` §3.11 (Kopf sagt mehr zu, als der Test fängt) | `src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_funde.go` · `func g4DocV2() []byte`; `src/adapters/driven/odata/csdlv2/elemente.go` · `l.slots = append(l.slots, slot{fund: model.Fund{Art: model.FundPropertyTyp` | ja — Mutation im Scratchpad | Fänger führt nur eine Form der Eingabe |
| F-4 | MEDIUM | Ein Property-Typ aus einem Namensraum, der mit `Edm` beginnt, aber nicht `Edm` oder `Edm.…` ist (gültig: `Edm2`, `EdmX`, `Edmonton`; D6 prüft das für `Namespace` und `Alias`), kommt in keinem Dokument der Suite als Typ einer Property vor. Die Mutante `strings.HasPrefix(s, "Edm")` statt `"Edm."` in `property()` bleibt in der ganzen Suite grün; sie ließe `Edm2.Adresse` als `nicht_aufloesbar` ablehnen und einen Dienst mit einem solchen Namensraum unimportierbar machen (`ODR-E6004`). | `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe); `SPEC-171` Nr. 5 und Nr. 10 a (`Edm` und `Edm.`-Anfang reserviert, `Edm2` nicht) | `src/adapters/driven/odata/csdlv2/aufloesen.go` · `if strings.HasPrefix(s, "Edm.") {` | ja — Mutation im Scratchpad | Fänger führt nur eine Form der Eingabe |
| F-5 | LOW | Die Menge der ungültigen `Multiplicity`-Werte ist unbeschränkt, N3 führt zwölf; `1..*` als gültig zugelassen bleibt in der ganzen Suite grün. Der Plan nennt dazu keine Prüfgrenze mit Grund und Folge. | `AGENTS.md` §3.17 (unbeschränkte Menge ohne Prüfgrenze) | `src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_assoziation.go` · `"1..1", "*..1", "* ", "1.", "0..1 ", " 1", "2", "0", "many", "0..1..*"` | ja — Mutation im Scratchpad | Unbeschränkte Menge ohne Prüfgrenze im Plan |
| F-6 | LOW | Die Handoff-Regel des Plans („ein Commit für Code und Plan“, `AGENTS.md` §3.9) ist nicht eingehalten: drei Commits ändern Produktcode, Suite und Test (`a89be28`, `eabb6a9`, `cea5627`), ohne §3 oder §6 des Plans anzufassen; den Plan zieht erst der vierte Commit (`dd68d95`) nach. | `AGENTS.md` §3.9; Plan §3 Betriebsregeln | `docs/plan/planning/done/slice-v2-metadaten.md` · „Der Handoff ist **ein Commit für Code und Plan**“ | ja — `git log --stat b5ac4a5..dd68d95` | Plan folgt der Änderung nicht im selben Commit |
| F-7 | LOW | Vier Stellen des Plans tragen ihre Berichtigung nicht oder widersprechen einander: (a) „Mutationen, Doubles, Aufbau (Architect)“ nennt „D5 (19 + Klammerformen)“, Tabelle und Suite führen 21 V4-Namen; (b) die dritte Rückgabe („19 gezählt, 21“) steht mit Ausgang „offen“, Adresse Planner, obwohl die Tabelle im selben Commit nachgezogen ist; (c) §8 sagt „Der Slice ist nicht nebenläufig“, während I2 und I3 `nebenläufig` sind und eine Rate tragen; (d) die Zeile N6 nennt die Zusage „deren `End`-Rolle nicht zur Association passt“, `SPEC-171` Nr. 10 f die engere „Rollen der Navigation, sobald eine Navigation auf sie verweist“. | `AGENTS.md` §3.9 (alle Stellen, die eine Aussage tragen) | `docs/plan/planning/done/slice-v2-metadaten.md` · `D5 (19 + Klammerformen)`; `Der Slice ist nicht nebenläufig` | ja — `grep` | Berichtigung nicht an allen Stellen nachgezogen |

*Nicht gezählt (Auftrag):* die beiden offenen Rückgaben an den Architect (Absatz (11)); der Plan führt beide mit Adresse.

### Hinweise ohne erwartete Aktion

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| I-1 | INFO | Ab 100 aufeinanderfolgenden leeren Reads (`0, nil`) liefert der Leser `io.ErrNoProgress` statt des Ergebnisses (Probe: 99 leere Reads vor den Daten ergeben das Modell, 100 den Fehler, der kein `MetadatenFehler` ist); `SPEC-125` Nr. 2 sagt ein zerlegungsunabhängiges Ergebnis auch mit leeren Reads zu und nennt unter *Ungeprüft* nur den endlosen Strom. Der V4-Leser trägt denselben Zweig (`csdl/lesen.go`, Zeile 69); Zuständigkeit: Architect. | `SPEC-125` Nr. 2 (geerbt) | `src/adapters/driven/odata/csdlv2/lesen.go` · `case errors.Is(err, io.ErrNoProgress):` | ja — Probe mit 100 leeren Reads | Zusage über Zerlegung ohne Grenze für leere Reads |
| I-2 | INFO | Die Trennung der Versionen (J3) hat als einziges Gate den Go-Test; `.a-check.yml` kennt nur Schichten, `adapters-driven` hat keine Kanten zwischen Adaptern. Die Fitness Function trägt ihre Grenze im Namensschema (`…v2`): ein V2-Paket ohne diese Endung zählt als V4. | [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) Fitness Function | `src/adapters/driven/odata/versionstrennung_test.go` · `strings.HasSuffix(von, "v2") != strings.HasSuffix(nach, "v2")` | nein | Grenze einer Fitness Function im Namensschema |
| I-3 | INFO | Vier Mutanten sind keine Lücke: der Wächter „kein Read nach `io.EOF`“ in `strom.Read` ist äquivalent (der Decoder liest nach `io.EOF` nie wieder; ohne den Wächter für Fehler wird A5 rot); die Vertauschung von `basen` mit `ketten` und von `ketten` mit `summe` ist rot durch den Speicherwächter der Suite (Panik mit Zeile D9), nicht durch eine Erwartung; „BOM nicht gezählt“ in der Form „spätere Reads“ trifft nur Dokumente über 4096 Bytes (die Form „im ersten Read“ ist rot in E1); J3d ist äquivalent. | Maintainability | `src/adapters/driven/odata/csdlv2/strom.go` · `if s.ende != nil {` | ja | äquivalente und indirekt rote Mutanten |
| I-4 | INFO | Zeile I1 vergleicht die Namensräume nur nach der Zahl (`len(m.Namespaces()) != 6`), ihr Kopf sagt „gegen die Erwartung der Erzeugung“; die Mutante „Namensräume aus einer Map“ ist in I1 0 von 30 rot, in der ganzen Suite rot in A4, D11 und F4 (zwei Schemas, Wahrscheinlichkeit 1/2 je Dokument). | `AGENTS.md` §3.11 | `src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_funde.go` · `len(m.Namespaces()) != 6` | ja | Fänger der Zeile schmaler als ihr Kopf, von anderen Zeilen gedeckt |
| I-5 | INFO | Der Wirkungs-Fänger von I2 lässt alle 16 Goroutinen dieselben Dokumente in derselben Reihenfolge lesen; eine Paketvariable, die beim Auflösen geschrieben und später im selben Aufruf gelesen wird, zeigt deshalb meist dasselbe Dokument und hat keine sichtbare Wirkung (0 von 30 ohne Detektor). Die Zusage trägt dort allein der Detektor (30 von 30); der Plan nennt als akzeptiertes Negativ nur die Form mit kurzem Fenster. | `AGENTS.md` §3.10 | `src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_funde.go` · `func i2LaufV2` | ja | Wirkungs-Fänger mit gleichförmiger Last |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/odata/csdlv2/` (Produktcode, acht Dateien) | geprüft; Befunde nur über Fänger und Zusage (F-1 bis F-4), kein Korrektheitsfehler gefunden: 17 496 Mutationen des Referenzdokuments bei fünf Zerlegungen liefern dasselbe Ergebnis wie bei einem Read, kein Fehler ohne `MetadatenFehler`, kein Absturz; das Northwind-Dokument ergibt Modell und Funde nach `SPEC-171` |
| `src/adapters/driven/odata/versionstrennung_test.go`, `csdlv2/leser_test.go` (J1 bis J3) | geprüft, ohne Befund (INFO I-2) |
| `src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_*.go` und `testdata/` | geprüft; Lücken F-2 bis F-5; sonst gleicht keine Fixture dem Wert einer Mutation, die Orakel (`wurzelUrsache`, `d11Dokument`, `g1FaelleV2`, `i1DokV2`) lesen die Parameter der Erzeugung, nicht den Leser |
| V4-Pfad (`csdl`, `jsoncodec`, `urlparser`, `model`, `ports/outbound`) | geprüft, ohne Befund (J4: `git diff --numstat --diff-filter=MDRT` leer) |
| `spec/spezifikation.md` (Diff 8/6) und Reihenfolge Spezifikation vor Code | geprüft, ohne Befund: der Planner-Commit `dfff71f` liegt vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Straten widersprechen sich nicht |
| Gates `make a-check`, `make groessenmarke`, `make lint`, `make zeichen-check`, `gofmt`, `go vet` | geprüft, ohne Befund |
| Kommentare (§3.7, §3.11), Suppressionen (§3.2) | geprüft; eine Aussage über das Prüfbare hinaus: F-3 (und I-4); kein `nolint`, kein `d-check:ignore` |
| Hard Rules §3.12, §3.13, §3.16 (Randformen, Adressen, Raster) | geprüft, ohne Befund: jede Zelle hat Ort oder „gilt nicht“; die zwei Rückgaben und die Zählabweichung stehen mit Ausgang *offen* und Adresse im selben Commit wie ihr Bericht |
| Harness-Dateien (`.golangci.yml`, `.a-check.yml`, `Dockerfile`, `Makefile`, `harness/`) | geprüft, ohne Befund: unverändert |

## Summary

Kein HIGH, kein ADR-Verstoß, kein Korrektheitsfehler im Produktcode gefunden. Vier MEDIUM: ein quadratisches Wachstum der Bindings mit einer falsch begründeten Zusage (F-1) und drei Fänger-Lücken bei den Funden und Typnamen (F-2 bis F-4). Drei LOW (F-5 bis F-7) und fünf INFO. 182 von 195 Mutanten rot; die 13 grünen: 6 äquivalent, zu schwach oder in der falschen Zielzeile, 7 Lücken. I2 und I3 mit Detektor in 30 von 30 Prozessen rot (je vier Mutationen), J4 leer, Marke 4525 von 8200.
