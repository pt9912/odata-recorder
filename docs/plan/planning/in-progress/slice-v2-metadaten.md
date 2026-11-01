# Slice slice-v2-metadaten: Metadaten-Leser für V2 (EDMX/CSDL V2) mit Validierungsbericht und Version je Service

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

**Berührte Spec-Stellen:** `SPEC-170` · `SPEC-171` (darin Nr. 1, 2 c, 5, 6 mit der Grenze der Bindings, 7 und die Strukturregeln der Assoziationen Nr. 10) · `SPEC-172` *Ungeprüft* (Typ `Edm.DateTimeOffset` im V2-Dokument) · `SPEC-178` · `SPEC-179` · `LH-FA-17.a` (versionierter Adapter für V2-Metadaten, keine Vermischung von V2 und V4 in einem Service-Modell) · `LH-RB-01` (nicht Unterstütztes erscheint nie als unterstützt) · `SPEC-002` (OData-Modell) · `SPEC-125` (Vertrag des V4-Metadatenlesers als Vorbild; Nr. 6 a: ein EDMX der Versionen 1 bis 3 ist dort `version`; Nr. 2 und *Ungeprüft*: 100 leere Reads in Folge enden mit `io.ErrNoProgress`) · `SPEC-107` Nr. 14 und 15 (`odata_version` des Service, Metadaten durch den Import) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
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

**Ziel:** Ein Leser für das EDMX/CSDL einer V2-Quelle (Teilumfang nach dem dokumentierten V2-Umfang des Architect) liefert das Modell des Dienstes als `model.ServiceModel` mit `Version()` gleich `2.0` (`SPEC-171`, Paket `csdlv2`) und einen Validierungsbericht: ausgewertet wird der Teilumfang, alles andere erscheint als Fund und nie als unterstützt (`LH-RB-01`). Das Modell trägt die Version des Dienstes (die „Version je Service“ des Slice-Titels: die Version, die das Modell nennt, nicht ihre Deklaration in der Konfiguration), und ein V2-Dokument vermischt sich nicht mit dem V4-Modell. Der V4-Leser (`csdl`) und das Verhalten eines V2-Dokuments an ihm (`ODR-E6004`, Ursache `version`) bleiben unverändert.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Payload-Codec für V2 — `slice-v2-payload-lesen` und `slice-v2-payload-schreiben`.
- Query-Semantik und Kompatibilitätsfehler — `slice-v2-query`.
- Deklaration der Version eines Service in der Konfiguration und Wahl des Adapters im Composition Root — `slice-v2-bootstrap`.
- Anpassung des Imports an einen Satz von Codecs je Dienst — `slice-v2-import-codecs` (der Nehmer trägt es zuerst ein, `AGENTS.md` §3.13; die Änderung am Vertrag von `SPEC-127` Nr. 1 liegt dort).
- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`; sie erscheinen im Bericht als Fund.
- V2-Mutationen und ihre Metadaten — `welle-odata-v2` §6; was die Spezifikation nicht nennt, ergibt den Kompatibilitätsfehler.

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

- [ ] Leser und Modell: EDMX/CSDL V2 im Teilumfang des Architect (Wurzel, Schema, Entity-Typen, Schlüssel, Properties, Assoziationen und Navigation-Properties, Container, Entity-Sets) in `model.ServiceModel` mit Version `2.0`; Namen aufgelöst; V4-Modell und V4-Leser unberührt (die Tests des V4-Lesers laufen unverändert grün).
- [ ] Validierungsbericht: jedes nicht ausgewertete Konstrukt (Funktionsimporte, komplexe Typen soweit nicht im Umfang, herstellerspezifische Annotationen und Attribute, nicht unterstützte Property-Typen) ist ein Fund; nichts Unterstütztes wird verschwiegen und nichts Unterstütztes behauptet (`LH-RB-01`).
- [ ] Verträge und Randformen eines Lesers fremder Bytes (Kodierung, Größe, Tiefe, ungültiges XML, falsche Wurzel oder Version, Verweise ohne Ziel, Wiederholung, Zerlegung des Stroms in Reads): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der Dokumente.
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
| `src/adapters/driven/odata/csdlv2/` | neu | Leser und Validierungsbericht (`SPEC-171`); kein eigenes Modell und kein neuer Port ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)) |
| `src/adapters/driven/odata/csdlv2/leser_test.go` | neu | Aufruf der Suite mit `csdlv2.New()` und `TestLeserOhneZustand` (J1) |
| `src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_*.go` (sechs Dateien: Vertrag, Lesen, Auflösen, Assoziation, Funde, Modell) | neu | Suite `MetadataCodecV2Vertrag(t, neu)` mit Generator über die Klasse; nutzt die Doubles von `outboundtest`, die V4-Dateien bleiben byte-gleich |
| `src/hexagon/ports/outbound/outboundtest/testdata/v2_referenz.xml`, `v2_vererbung.xml`, `v2_composite.xml` | neu | Fixtures, gekürzt nach dem Aufbau eines öffentlichen V2-Dokuments (kein Abruf, `harness/README.md` §Safety); Erwartungen als feste Werte im Test |
| `src/adapters/driven/odata/versionstrennung_test.go` | neu | Fitness Function: kein V2-Paket importiert ein V4-Paket und umgekehrt (J2, J3) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Leser liest fremde Bytes (ein Format, das dieses Repo nicht schreibt): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Dokumente mit Escape-Folgen, BOM oder Rohbytes entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Auftrag an den Architect:** erledigt ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), `Accepted`). Auftrag 1 (Umfang) steht in `SPEC-170`; 2 (Modell) in [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) und `SPEC-171` Nr. 8 (derselbe Typ mit Version `2.0`); 3 (Abbildung) in `SPEC-170` Nr. 4, `SPEC-171` Nr. 5 und 6; 4 (Dokument) in `SPEC-171` Nr. 2 und 3; 5 (Bericht) in `SPEC-171` Nr. 4, 9 und `SPEC-125` Nr. 13; 6 (Ort der Version) in `SPEC-170` Nr. 1 und `SPEC-175` Nr. 1.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; vom Architect gefüllt, der Implementer trägt die Raten der Mutationen ein):

Alle Zeilen `andere`, außer I2 und I3 (`nebenläufig`). Mengen sind Schleifen über den Wortlaut der Zusage (`AGENTS.md` §3.15, §3.17); „Schleife“ heißt: ein Fall je Element, beide Seiten jeder Grenze. Ein Dokument mit Rohbytes oder Escape-Folge entsteht aus Bytewerten (`AGENTS.md` §3.18). Jede Fehlerzeile prüft Ursache, Ort, Code, `errors.Is` und Nullwerte von Modell und Bericht (`pruefeFall`).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A1 | Argumente: `ctx` nil, `src` nil, `grenze` unter 1 sind `model.ErrUngueltig` ohne `MetadatenFehler` und ohne Code; `src` bleibt ungelesen; die Prüfung der Argumente geht der des Kontexts voraus (`SPEC-171` Nr. 1, `SPEC-125` Nr. 1 a bis d) | Schleife über jede nichtleere Teilmenge der drei Verstöße (7) mal Kontext beendet oder nicht (2); `grenze` 0, -1, MinInt64; Doppel zählt Reads: 0 | Prüfung `grenze < 1` zu `grenze < 0` | andere |
| A2 | Kontext: Fehler des Kontexts unverändert (`==`), ohne Code; geprüft vor jedem Read aus `src` und einmal vor dem Auflösen; ein Verstoß in schon gelesenen Bytes geht vor (`SPEC-125` Nr. 1) | Schleife über Abbruchstellen (vor Read *k* für jedes *k* eines zerlegten Dokuments, vor dem Auflösen) mal `Canceled` und `DeadlineExceeded`; Fall Verstoß in gelesenen Bytes plus Abbruch | Kontextprüfung vor dem Auflösen entfernt; Kontextfehler in `MetadatenFehler` gewickelt | andere |
| A3 | Fehler von `src` ungleich `io.EOF` beendet das Lesen mit ihm (`errors.Is`, kein `MetadatenFehler`, kein Code), auch nach vollständigem Dokument; Daten mit dem Fehler werden vorher verarbeitet, ein Verstoß in ihnen schlägt den Fehler (`SPEC-125` Nr. 2) | Schleife über Fehlerort (vor, im, nach dem Dokument, zusammen mit Daten) mal Dokument gültig oder ungültig | Fehler von `src` durch `syntax` ersetzt | andere |
| A4 | Ergebnis und Fehler hängen nicht von der Zerlegung in Reads ab: Einzelbytes, jede Zweiteilung an jeder Stelle (BOM, Deklaration, mehrbyteiges Zeichen, Referenz), leere Reads `0, nil`, `n > 0` mit `io.EOF` (`SPEC-125` Nr. 2) | Schleife über jede Zweiteilung von Referenzdokument, BOM plus Deklaration, Umlaut, Referenz und je einem Dokument je Ursache; Vergleich mit dem Ergebnis aus einem Read | Randprüfung des BOM (weniger als 3 Bytes) entfernt; leerer Read als Ende | andere |
| A5 | Strom gehört dem Aufrufer: kein `Close`, kein Read nach `io.EOF` oder Fehler, höchstens `grenze + 1` Bytes aus `src` (`SPEC-125` Nr. 2, 5) | `teilReader` zählt Reads, Bytes, `Close`, Reads nach dem Ende; Fälle gültig, ungültig, über der Grenze, Fehler von `src` | `Close` am Ende gerufen; weiterlesen nach Fehler | andere |
| B1 | Leer, nur Leerraum, nur BOM, kein XML, endet im Element, ohne Wurzel sind `syntax` (`SPEC-125` Nr. 4) | Fall je Element der Aufzählung (8) | leeres Dokument als gültiges leeres Modell | andere |
| B2 | Kodierung UTF-8: ein BOM übergangen, UTF-16-BOM beider Reihenfolgen `kodierung`, Deklaration mit `encoding` ungleich `utf-8` `kodierung`; ungültiges UTF-8 und jedes Zeichen außerhalb des XML-Bereichs `syntax` (`SPEC-125` Nr. 3) | Schleife über die Zeichen an beiden Seiten jeder Bereichsgrenze (0x08/0x09, 0x0B, 0x0C, 0x0D/0x0E, 0x1F/0x20, D7FF/D800, DFFF/E000, FFFD/FFFE/FFFF, 10FFFF) und BOM-Doppelung; Schreibweisen von `utf-8` | untere Grenze 0x20 zu 0x1F | andere |
| B3 | `<!DOCTYPE` in jeder Form ist `doctype`, bevor etwas gelesen wird; keine externe Ressource, keine Expansion (`SPEC-125` Nr. 4 a) | Fall je Form (leer, intern, `SYSTEM`, `PUBLIC`, Parameter-Entität, Entitäts-Angriff mit Zeitgrenze) | Präfixprüfung `DOCTYPE` verkürzt | andere |
| B4 | Referenzen: in jedem Attributwert und Text gelten fünf Entitäten und Zeichenreferenzen im XML-Bereich, jede andere ist `syntax`; Ersatzzeichen-Grenzen beider Seiten; außerhalb der Wurzel ist jede Referenz und jede `CDATA`-Sektion `syntax` (`SPEC-125` Nr. 4 b) | Schleife über Stelle (Attributwert je ausgewertetes Element, Text, vor und nach der Wurzel) mal Referenz (gültig, ungültig, Grenzwerte, neben `U+FFFD`) | nur die erste Referenz je Wert geprüft | andere |
| B5 | Kommentare, Verarbeitungsanweisungen, Deklaration nur am Anfang (`version`, `encoding`), doppelte Attribute (Namensraum und lokaler Name), Nachlauf nach der Wurzel (`SPEC-125` Nr. 4) | Fall je Element der Aufzählung; zwei Präfixe derselben URI doppelt, derselbe lokale Name in zwei Namensräumen nicht doppelt | Namensraum beim Vergleich der Attribute ignoriert | andere |
| B6 | Namensräume: ein Präfix ohne Deklaration bindet keinen bekannten Namensraum, der Name zählt als fremd (`SPEC-125` Nr. 4 c, `SPEC-171` Nr. 4) | Fall Element und Attribut ohne Deklaration; Name im Fund mit Präfix an der Stelle des Namensraums | undeklariertes Präfix als `edm` gelesen | andere |
| C1 | Wurzel: `Edmx` im Namensraum des V2-EDMX; `Edmx` im V4-Namensraum ist `version`; jede andere Wurzel (Name, Namensraum, keiner) ist `wurzel` (`SPEC-171` Nr. 2 a) | Schleife über Name mal Namensraum (V2, V4, fremd, keiner) | Vergleich des Namensraums entfernt; V4-Wurzel als `wurzel` | andere |
| C2 | `Version` der Wurzel ist genau `1.0`; fehlend, leer, jeder andere Wert ist `version`; (a) geht (b) vor (`SPEC-171` Nr. 2 b) | Schleife über `1.0`, fehlend, leer, `1`, `1.00`, `1.0 `, ` 1.0`, `2.0`, `3.0`, `4.0`, `1.0.0`; Fall `Foo` mit `Version="3.0"` ist `wurzel` | `1.0` mit Präfixvergleich | andere |
| C3 | `DataServiceVersion` am Kind `DataServices`: fehlend gilt `1.0`; `1.0` und `2.0` gültig; jeder andere Wert (`3.0`, `2.0.1`, leer) `version`; ein zweites `DataServices` ist `struktur` beim Beginn des zweiten, keines ist `struktur` am Ende der Wurzel; ein Attribut gleichen Namens ohne Namensraum ist ohne Wirkung (`SPEC-171` Nr. 2 c, Abschnitt 5 Änderung 1) | Schleife über die Werte; Fälle zweites, keines, Attribut ohne Namensraum | `3.0` als gültig; zweites `DataServices` ignoriert | andere |
| C4 | `Schema` ist gültig in den drei Namensräumen CSDL 1.0 bis 2.0 mit gleichem Ergebnis; im Namensraum CSDL 3.0 oder im V4-Namensraum `version`; in einem anderen Fund `erweiterung` (`SPEC-171` Nr. 3) | Schleife über fünf Namensräume mit Vergleich der Modelle der drei gültigen | Namensraum 2008/09 nicht akzeptiert | andere |
| C5 | Wurzel und Version stehen fest, sobald das Start-Tag der Wurzel gelesen ist; Müll davor ist `syntax`; (a) geht (b) vor (`SPEC-171` Nr. 2; Zusatztabelle Zeile 1) | Fall je Paar: Müll davor und falsche Version, falsche Version und ungültiges XML dahinter, falsche Wurzel und falsche Version, falsche `DataServiceVersion` und Verstoß im Schema dahinter | Version erst am Ende geprüft | andere |
| D1 | Ein Entity-Typ, der nicht `Abstract` ist, hat mindestens eine `PropertyRef` in seiner Kette, sonst `ohne_schluessel`; ein Schlüssel im abgeleiteten Typ ist zulässig, wenn kein Basistyp einen hat (`SPEC-171` Nr. 7, `SPEC-125` Nr. 9 g) | Schleife `Abstract` mal Schlüssel (eigener, geerbter, keiner) mal Tiefe der Kette (1 bis 3) | `Abstract` nicht beachtet | andere |
| D2 | Eine Referenz ohne Ziel ist `nicht_aufloesbar` mit Ort `<Typ>@BaseType`, `<Typ>/<Name>@Type`, `<Container>/<Set>@EntityType`, `<Typ>/<Name>@Name`; Schreibweise und Leerraum zählen; ein deklarierter, nicht modellierter Name (`ComplexType`) löst nicht auf (`SPEC-171` Nr. 7, `SPEC-125` Nr. 9) | Schleife Referenzart (4) mal Form (ohne Punkt, unbekannter Namensraum, unbekannter Name, Leerraum vorn und hinten, Schreibweise, `ComplexType`-Name, Name aus fremdem Dokument); Relationship und `End`-Typ siehe N1 | `TrimSpace` auf die Referenz | andere |
| D3 | Kette der Basistypen: Selbstbezug, Zyklus, Schwanz in einen Zyklus unter 65 Typen `nicht_aufloesbar` (Ort der erste Typ), ab dem 65. verschiedenen Typ `kette` (auch ein Ring); 64 Typen gültig (`SPEC-171` Nr. 7 a) | Schleife Ringgröße 1, 2, 63, 64, 65, 66 mal Schwanz 0 und 1 mal Reihenfolge im Dokument | Grenze 64 zu 65 | andere |
| D4 | Schlüsseltypen: jeder nicht unterstützte bekannte Typ (Binary, Byte, DateTimeOffset, Double, Int16, SByte, Single, Time), ein deklarierter Typ und ein Name mit `/` ist `schluesseltyp`; die sieben unterstützten (mit `Edm.DateTime`) sind gültig (`SPEC-171` Nr. 5, 7 f) | Schleife über die 8 + 2 + 7 Typen | `Edm.Double` als Schlüssel erlaubt | andere |
| D5 | Formen des Property-Typs: `Edm.<unbekannt>` (darunter die V4-Namen `Date`, `Duration`, `TimeOfDay`, `Stream`, `Geography…`, `Untyped`), `Collection(…)` in jeder Form und jeder nicht auflösbare Name sind `nicht_aufloesbar`; jeder bekannte nicht unterstützte Name und jeder deklarierte Typ (`ComplexType`, Entity-Typ) lässt die Property im Modell mit `Unterstuetzt()` falsch und erzeugt `property_typ` (`SPEC-171` Nr. 5, Abschnitt 5 Änderung 2) | Schleife über alle Namen der Mengen (V4-Namen 21: `Date`, `Duration`, `TimeOfDay`, `Stream`, `Untyped` und je acht `Geography…` und `Geometry…`; `Collection(`-Formen: Klammer fehlt, Rest dahinter, Rest im Namen, verschachtelt) und über die 8 + 2 bekannten und deklarierten | `Edm.Date` als bekannter V2-Typ | andere |
| D6 | Pflichtangaben und Namensform: fehlendes oder leeres Pflichtattribut ist `struktur` (Menge: Abschnitt 5 Nr. 10 a); einfacher Name 1 bis 128 Zeichen, Namensraum 256; `Edm` und `Edm.`-Anfang reserviert; Wahrheitswerte genau `true` oder `false`; Ort wie `SPEC-125` Nr. 8 (`SPEC-171` Nr. 1) | Schleife Attribut (alle der Menge) mal {fehlt, leer, ungültig}; Grenzfälle 127/128/129 und 255/256/257 Zeichen; `edm`, `EdmX`, `Edm2` gültig | `Role` an `End` nicht als Pflicht | andere |
| D7 | Eindeutigkeit: doppelt sind Entity-Typ (auch über zwei `Schema` eines Namensraums), Member, `EntitySet`, `PropertyRef`, `Key`, zweiter `EntityContainer`, `Association` (auch über Schemas), `AssociationSet`; Alias-Widerspruch; gleiche Namen in verschiedenen Namensräumen gültig (`SPEC-171` Nr. 1, Abschnitt 5 Nr. 10 e) | Schleife über die Menge der Doppelungen (8) und über Alias gleich eigener Namensraum, zwei Schemas mit verschiedenen Aliasen | Association-Namen nicht verglichen | andere |
| D8 | Ketten: zwei Typen einer Kette mit je einem `Key` ist `struktur` (Ort der abgeleitetere); ein geerbter Name an Property oder Navigation ist `struktur` (`SPEC-171` Nr. 7 c) | Fall je Kombination Property und Navigation, gleicher Typ, Basis, Basis der Basis | Navigation nicht gegen geerbte Namen geprüft | andere |
| D9 | Rangfolge der Auflösungsschritte: acht Durchläufe in fester Reihenfolge; für je zwei meldet ein Dokument mit beiden den früheren, innerhalb eines Schritts der erste im Dokument (`SPEC-171` Nr. 7, `SPEC-125` Nr. 9) | Schleife über alle 28 Paare der Schritte in beiden Dokumentreihenfolgen | Schritte (b) und (c) vertauscht | andere |
| D10 | Jeder Verstoß ist `ODR-E6004` mit genau einer der 13 Ursachen und dem Ort der Zusage; bei jedem Fehler sind Modell und Bericht Nullwerte (`SPEC-125` Nr. 13, `SPEC-171` Nr. 9) | jede Fehlerzeile der Suite prüft Code, `errors.Is`, Ursache, Ort, Nullwerte; Fall Teilbericht vor `nicht_aufloesbar` | Teilbericht bei Fehler geliefert | andere |
| D11 | Alles andere ist gültig: jedes Dokument der Klasse aus Schemas (0 bis 2), Typen (0 bis 3), Properties der sieben Typen, Schlüsseln, Assoziationen, Sets, fremden Teilen ergibt das erwartete Modell und die erwarteten Funde | Generator über die Klasse (deterministisch, ohne Zufall), Orakel aus den Parametern der Erzeugung (nicht aus dem Lesen), Zählprüfung (Mindestzahl Dokumente und je Parameterwert) | zwei Schemas abgelehnt | andere |
| N1 | Navigation löst: `Relationship` löst (mit Alias) auf eine `Association`, deren `End` die Rollen `FromRole` und `ToRole` tragen (verschieden), und der `Type` des `End` mit `ToRole` löst auf einen Entity-Typ, sonst `nicht_aufloesbar` mit Ort `<Typ>/<Name>@Relationship` (`SPEC-171` Nr. 6) | Schleife Form (Relationship fehlt in der Association, Alias unbekannt, Leerraum, Association ohne Rolle F, ohne Rolle G, F gleich G, Typ G nicht auflösbar, Typ G ein `ComplexType`) mal Stelle (erster Typ, abgeleiteter Typ); Fall: Typ des F-`End` unauflösbar bleibt gültig | `Target` aus dem F-`End` gelesen | andere |
| N2 | `NavigationProperty{Name, Target, Collection}`: `Target` der qualifizierte Name des Typs, `Collection` wahr genau bei `Multiplicity` `*` des `End` mit `ToRole` (`SPEC-171` Nr. 6) | Schleife `Multiplicity` (`0..1`, `1`, `*`) mal Referenzform (Alias, Namensraum) | `Collection` auch bei `0..1` | andere |
| N3 | `Multiplicity` ist `0..1`, `1` oder `*`, sonst `struktur`, Ort `<Association>@Multiplicity`, geprüft beim Lesen jedes `End` in `Association`, auch ohne Verweis (Abschnitt 5 Nr. 10 d) | Schleife ungültiger Werte (leer, `0..*`, `1..1`, `*..1`, `* `, `1.`) mal Association (referenziert, nicht referenziert) | nur referenzierte `End` geprüft | andere |
| N4 | Bindings: zu jeder `AssociationSet` A und jeder Navigation N mit gleicher `Relationship` ein `NavigationBinding{Set, Path, Target}` mit Set aus dem `End` der Rolle F, Path gleich dem Namen von N, Target aus dem `End` der Rolle G; Reihenfolge: `AssociationSet` im Dokument, darin die Navigation-Properties (`SPEC-171` Nr. 6) | feste Erwartung für Dokumente mit zwei Navigationen je Relationship, zwei `AssociationSet` je Association, gemischten Reihenfolgen | Reihenfolge nach Typen statt nach Sets | andere |
| N5 | Gültig ohne Binding: eine Navigation ohne `AssociationSet` bleibt im Modell ohne Binding; eine `AssociationSet` ohne Navigation erzeugt kein Binding (ihre Auflösung gilt, N6) | Fälle beider Richtungen mit Erwartung der Anzahl | Binding mit leerem `Path` erzeugt | andere |
| N6 | Eine `AssociationSet`, deren `Association` nicht auflöst, deren `End`-Rollen nicht die Rollen der Navigation (`FromRole`, `ToRole`) tragen, sobald eine Navigation auf sie verweist, oder deren `EntitySet` kein `EntitySet` des Containers ist, ist `nicht_aufloesbar`, Ort `<Container>/<AssociationSet>`; `EntitySet` und `AssociationSet` in der Reihenfolge der Kinder des Containers (`SPEC-171` Nr. 6, 7 e, Abschnitt 5 Nr. 10 f) | Schleife Verstoß (3) mal `End` (F, G); Fall `Set` ohne Ziel vor `AssociationSet` ohne Ziel und umgekehrt | unbekanntes `EntitySet` ergibt Binding mit leerem `Target` | andere |
| N7 | Eine `Association` und eine `AssociationSet` haben genau zwei `End` mit verschiedener `Role`, sonst `struktur` (Ort der Association bzw. `<Container>/<Name>`, bei gleicher Rolle `@Role`; ein drittes `End` beim Beginn, zu wenige am Ende) (Abschnitt 5 Nr. 10 b, c) | Schleife Zahl der `End` (0, 1, 3) und gleiche Rolle mal {`Association`, `AssociationSet`} | drittes `End` still ignoriert | andere |
| E1 | Größe: höchstens `grenze` Bytes (BOM zählt); das Byte `grenze + 1` ist `zu_gross`; ein Verstoß davor entscheidet, dahinter wird nie gesehen; `MaxInt64` gültig (`SPEC-125` Nr. 5) | Schleife `grenze − 1`, `grenze`, `grenze + 1` mal {gültig, Verstoß davor, Verstoß dahinter} mal BOM | `>` zu `>=` | andere |
| E2 | Tiefe: das 65. offene Element ist `tiefe`, das 64. gültig, in Schema, fremdem Element und gemeldetem Konstrukt gleich (`SPEC-125` Nr. 5) | Schleife Inhalt (Typ, Fremdes, Konstrukt) mal 64 und 65 | Grenze 64 zu 65 | andere |
| E3 | Summe der wirksamen Properties und Navigation-Properties (geerbte mitgezählt) höchstens 65536; 65537 ist `zu_viele_properties`, Ort der Typ (`SPEC-125` Nr. 5, 9 b) | Dokumente mit 65536 und 65537 aus Ableitungen und aus Navigationen (Generator) | Navigationen nicht gezählt | andere |
| F1 | Modell des Referenzdokuments: `Version()` `2.0`, Namensräume, Typen, Properties, Schlüssel, Navigationen, Sets, Bindings in Dokumentreihenfolge (`SPEC-171` Nr. 8) | feste Erwartung (Zahlen im Test) für `v2_referenz.xml` | `Version()` liefert `4.0` | andere |
| F2 | Vererbung im Modell: wirksame Properties und Navigationen mit den geerbten, Schlüssel geerbt, `Abstract` (`SPEC-125` Nr. 11) | feste Erwartung für `v2_vererbung.xml` und eine Kette von 3 | geerbte Navigationen ausgelassen | andere |
| F3 | Composite-Schlüssel in der Reihenfolge der `PropertyRef` | feste Erwartung für `v2_composite.xml` | Reihenfolge nach Properties | andere |
| F4 | Aliase und Schemas: jede Referenzart mit Alias oder Namensraum ergibt denselben qualifizierten Namen (`BaseType`, Property-`Type`, `Relationship`, `End`-`Type`, `EntityType` des Sets, `Association` der `AssociationSet`) | Schleife Referenzart (6) mal {Alias, Namensraum} mal {ein, zwei Schemas}; `Alias` gleich `Namespace` | Alias bei `Association` nicht aufgelöst | andere |
| F5 | Typabbildung: `Edm.DateTime` ergibt `Edm.DateTimeOffset`, die sechs anderen bleiben; die acht nicht unterstützten Typen bleiben im Modell mit `Unterstuetzt()` falsch, `Edm.DateTimeOffset` des Dokuments trägt `EdmType` `Edm.DateTimeOffset(V2)`; `UnterstuetzteEDMTypen()` unverändert (`SPEC-170` Nr. 4, `SPEC-171` Nr. 5, 8, Abschnitt 5 Änderung 2) | Schleife über die 7 + 8 Typen | `Edm.DateTimeOffset` des Dokuments unverändert übernommen | andere |
| F6 | Randformen des Modells: `Open` immer falsch; `Nullable` Vorgabe wahr, `false` gelesen; Zugriffe liefern Kopien (`SPEC-171` Nr. 8, `SPEC-125` Nr. 11) | Fälle; Änderung einer zurückgegebenen Liste ändert das Modell nicht | Kopie durch Verweis ersetzt | andere |
| G1 | Funde nach Art, Name und Ort: `konstrukt` für jedes nicht ausgewertete Kindelement in `edm` oder `edmx`, `erweiterung` für Element oder Attribut mit fremdem Namensraum, `property_typ`; Ort nach `SPEC-125` Nr. 7 mit den V2-Elementen (`SPEC-171` Nr. 4) | Generator über Stelle (jedes ausgewertete Element, 14) mal Art (3) mal Name; Orakel aus der Erzeugung | Ort der `End`-Funde ohne Elternort | andere |
| G2 | Ohne Wirkung und nie ein Fund: Attribute ohne Namensraum, die Nr. 4 nicht nennt (`MaxLength`, `Precision`, `Scale`, `Unicode`, `FixedLength`, `ConcurrencyMode`, `Collation`, `StoreGeneratedPattern`), die drei `m:`-Attribute, Text, Kommentare, Verarbeitungsanweisungen (`SPEC-171` Nr. 4) | Schleife über die 8 + 3 + 3 mal Stelle | `m:IsDefaultEntityContainer` als Fund | andere |
| G3 | Der Bericht ohne Fund ist eine leere Menge, nicht `nil`; bei Fehler `nil` (`SPEC-171` Nr. 9, `SPEC-125` Nr. 12) | Fall gültig ohne Fund, Fall jeder Ursache | `nil` bei leerem Bericht | andere |
| G4 | Reihenfolge der Funde: die der Stellen im Dokument, ein Element vor seinen Attributen, `property_typ` an der Stelle seiner Property | Dokument mit Funden aller Arten, feste Erwartung | Funde nach Art gruppiert | andere |
| G5 | Nichts Gemeldetes erscheint als unterstützt: ein gemeldetes Konstrukt (`ComplexType`, `FunctionImport`, `ReferentialConstraint`, `Documentation`, `ValueTerm`, `Using`, `OnDelete`) steht nie im Modell; eine Property mit nicht unterstütztem Typ hat `Unterstuetzt()` falsch (`LH-RB-01`) | Schleife über die 7 Konstrukte (Modell unverändert gegen ein Dokument ohne sie) und die 8 Typen | `FunctionImport` als EntitySet übernommen | andere |
| G6 | Funde V2-eigener Stellen: `ReferentialConstraint` und `Documentation` an `Association`, `OnDelete` an `End`, `FunctionImport` am Container, herstellerspezifische Attribute (`m:HasStream`, `sap:label`) an Typ, Property, `Association`, `End`, `AssociationSet` | Fall je Stelle mit Erwartung von Art, Name, Ort | `End` ohne Namensraum-Attribut-Prüfung | andere |
| H1 | Form des Fehlers: Code `ODR-E6004`, `errors.Is(err, ErrMetadatenUngueltig)`, `Error()` mit `strconv.Quote` des Ortes (`SPEC-125` Nr. 13) | Fall je Ursache (13); Ort mit Steuerzeichen aus Bytewerten | Ort ohne `Quote` | andere |
| I1 | Gleiche Bytes ergeben gleiche Werte in der Reihenfolge des Dokuments, nie die einer Map (`SPEC-125` Nr. 14) | 30 Läufe über ein Dokument mit 8 Typen, 8 Sets, 6 Bindings, 6 Namensräumen: tiefer Vergleich; Rate je Lauf von `go test` | Bindings aus einer Map gelesen | andere |
| I2 | Gleichzeitige Aufrufe beeinflussen sich nicht; das Ergebnis jedes Aufrufs gleicht dem eines einzelnen; Variablen des Pakets sind nach der Initialisierung unveränderlich (`SPEC-125` Nr. 14) | 16 Goroutinen mit verschiedenen Dokumenten, Vergleich je Ergebnis mit dem Einzelaufruf (Wirkung); `make test-race` je Prozess | Dokumentbaum als Paketvariable; Tabelle der Typen beim Lesen verändert | nebenläufig |
| I3 | Gleichzeitige Aufrufe mit verschiedenen Ausgängen (gültig, jede Ursache, Abbruch) vermischen sich nicht: Fehler, Modell und Bericht jedes Aufrufs gehören seinem Dokument | Goroutinen mit Dokumenten verschiedener Ursachen und mit abbrechbarem Kontext, Zählung der Vermischungen (Wirkung) | gemeinsamer Fehlerwert im Paket | nebenläufig |
| J1 | Der Leser hat kein Feld: `csdlv2.New() *Leser` (`SPEC-171`) | `reflect` zählt die Felder: 0 | Feld `cache` hinzugefügt | andere |
| J2 | Ein V4-Dokument am V2-Leser und ein V2-Dokument am V4-Leser sind `version`, ohne Modell (`SPEC-171` Nr. 9, `SPEC-125` Nr. 6 a) | Schleife über Dokumentmenge (Referenzdokumente beider Versionen, minimale Wurzeln, Wurzel mit `DataServiceVersion` 3.0) mal beide Leser, in `versionstrennung_test.go` | V2-Leser nimmt die V4-Wurzel an | andere |
| J3 | Kein V2-Paket importiert ein V4-Paket und umgekehrt; Gemeinsames liegt in einem Paket ohne Versionsnamen (Fitness Function der Entscheidung zu V2) | Test über die Importe aller Nicht-Test-Dateien unter `src/adapters/driven/odata/` (`go/parser`, nur Importe), Schleife über alle Verzeichnisse; Gegenprobe des Scanners auf einem Wegwerfbaum mit Verstoß (rot) | Import von `csdl` in `csdlv2/lesen.go` | andere |
| J4 | Der V4-Pfad bleibt unberührt (`LH-FA-17.a`): keine Änderung an `csdl`, `model`, `ports/outbound` außer den neuen Dateien; die Tests dort laufen unverändert grün | benannte Ausgabe: `git diff --numstat <Basis>..HEAD` über `src/adapters/driven/odata/csdl`, `src/hexagon/model` und die bestehenden Dateien von `ports/outbound` ist leer, im Bericht genannt; `make test` | eine Zeile in `csdl/lesen.go` geändert: Ausgabe nicht leer | andere |

**Nebenläufige Zusagen** (`AGENTS.md` §3.10): I2 und I3 gelten nur als gefangen, wenn die Mutation in jedem Lauf rot wird. Der Implementer fährt je Mutation 30 Prozesse (`-test.count=1`, `make test-race`) und trägt die Rate mit Zahl der Prozesse und Messendem in einen Absatz von §3 ein; die Wirkungs-Fänger (Vermischungen zählen) sind unabhängig vom Detektor. I1 ist eine Wahrscheinlichkeits-Mutation (Map-Reihenfolge): der Bericht nennt die Rate je Lauf von `go test` mit Zahl.

**Mutationen vom Implementer rot gesehen** (`AGENTS.md` §3.10, §3.17; Scratchpad-Kopien des Quellbaums, je Mutation `go test -run 'TestVertrag/^<Zeile>_'`, alle rot; „+“ nennt ein Element der Klasse, das die Tabelle nicht nannte):

| Zeile | Mutation (rot) | roter Test (`TestVertrag/…`) |
|---|---|---|
| A1 | `grenze < 1` zu `grenze < 0` | `A1_Argumente` |
| A2 | Kontextprüfung vor dem Auflösen entfernt; Kontextfehler als `struktur` gewickelt | `A2_Kontext` |
| A3 | Fehler von `src` durch `syntax` ersetzt | `A3_Fehler_von_src` |
| A4 | BOM-Randprüfung entfernt; leerer Read als Ende | `A4_Zerlegung_in_Reads` |
| A5 | `Close` am Ende gerufen | `A5_Strom_gehört_dem_Aufrufer` |
| B1 | leeres Dokument gültig | `B1_leer_und_unvollständig` |
| B2 | Grenze 0x20 zu 0x1F; + U+FFFF als gültig | `B2_Kodierung` |
| B3 | Präfix `DOCTYPE` zu `DOCTYP` | `B3_DOCTYPE` |
| B4 | nur die erste Referenz je Wert geprüft | `B4_Referenzen` |
| B5 | Namensraum beim Vergleich der Attribute ignoriert | `B5_Kommentare,_PI,_Nachlauf` |
| B6 | undeklariertes Präfix `edm` als Namensraum gelesen | `B6_Namensräume` |
| C1 | Vergleich des Namensraums entfernt; V4-Wurzel als `wurzel` | `C1_Wurzel` |
| C2 | `1.0` mit Präfixvergleich | `C2_Version` |
| C3 | `3.0` gültig; zweites `DataServices` ignoriert | `C3_DataServiceVersion` |
| C4 | Namensraum 2008/09 nicht akzeptiert; + V4-Namensraum nicht `version` | `C4_Namensräume_des_Schemas` |
| C5 | (a) und (b) vertauscht | `C5_Wurzel_steht_fest` |
| D1 | `Abstract` nicht beachtet | `D1_ohne_Schlüssel` |
| D2 | `TrimSpace` auf die Referenz | `D2_nicht_auflösbar` |
| D3 | Kettengrenze 64 zu 65 | `D3_Zyklus_und_Kette` |
| D4 | `Edm.Double` als Schlüssel erlaubt; + `Edm.Time` | `D4_Schlüsseltypen` |
| D5 | `Edm.Date` bekannt; + `Edm.Untyped` bekannt | `D5_Formen_des_Property-Typs` |
| D6 | `Role` an `End` keine Pflicht | `D6_Pflichtangaben_und_Namen` |
| D7 | `Association`-Namen nicht verglichen; + doppelte `PropertyRef` | `D7_Eindeutigkeit` |
| D8 | Navigation nicht gegen geerbte Namen geprüft | `D8_Ketten` |
| D9 | Schritte (b) und (c) vertauscht; + (f) und (g) | `D9_Rangfolge` |
| D10 | Teilbericht bei Fehler | `D10_jeder_Verstoß_ist_ODR-E6004`, `G3_leerer_Bericht` |
| D11 | zweites `Schema` abgelehnt | `D11_alles_andere_ist_gültig` |
| N1 | `Target` aus dem `End` der Rolle F; + F gleich G zugelassen | `N1_Navigation_löst` |
| N2 | `Collection` auch bei `0..1` | `N2_Navigation_im_Modell` |
| N3 | `Multiplicity` nur an referenzierten `End` geprüft | `N3_Multiplicity` |
| N4 | Navigationen in umgekehrter Reihenfolge; + Bindings der `AssociationSet` umgekehrt | `N4_Bindings` |
| N5 | Binding mit leerem `Path` | `N5_gültig_ohne_Binding` |
| N6 | unbekanntes `EntitySet` nicht geprüft; + Rollen nicht gegen die Navigation | `N6_AssociationSet_löst` |
| N7 | drittes `End` der `Association` ignoriert; + der `AssociationSet`; + zu wenige `End` nur bei 0 | `N7_Zahl_der_End` |
| E1 | `>` zu `>=` | `E1_Größe` |
| E2 | Grenze 64 zu 65; + `>` zu `>=` | `E2_Tiefe` |
| E3 | Navigationen nicht gezählt; + `>` zu `>=` | `E3_Summe_der_Members` |
| F1 | `Version()` liefert `4.0` | `F1_Modell_Referenzdienst` |
| F2 | geerbte Navigationen ausgelassen | `F2_Vererbung_im_Modell` |
| F3 | Schlüssel sortiert statt in Reihenfolge der `PropertyRef` | `F3_Composite-Schlüssel` |
| F4 | Alias bei `Association` nicht aufgelöst | `F4_Aliase_und_Schemas` |
| F5 | `Edm.DateTimeOffset` des Dokuments unverändert | `F5_Typabbildung` |
| F6 | `Open` immer wahr; Zugriff auf `Namespaces()` ohne Kopie (nur im Scratchpad, `model` bleibt unverändert) | `F6_Randformen_des_Modells` |
| G1 | Ort der `End`-Funde ohne Elternort; + `AssociationSet` ohne Attributfunde | `G1_Funde` |
| G2 | `m:IsDefaultEntityContainer` als Fund; + `m:MaxDataServiceVersion` | `G2_ohne_Wirkung` |
| G3 | `nil` bei leerem Bericht | `G3_leerer_Bericht` |
| G4 | Funde nach Art gruppiert | `G4_Reihenfolge_der_Funde` |
| G5 | `FunctionImport` als `EntitySet` übernommen | `G5_nie_als_unterstützt` |
| G6 | `End` ohne Namensraum-Attribut-Prüfung | `G6_Funde_V2-eigener_Stellen` |
| H1 | Ort ohne `strconv.Quote` (nur im Scratchpad) | `H1_Fehlerform` |
| J1 | Feld `cache` | `TestLeserOhneZustand` |
| J2 | V2-Leser nimmt die V4-Wurzel an; V4-Leser nimmt die V2-Wurzel nicht als `version` (Scratchpad) | `TestVersionstrennungDokumente` |
| J3 | Import von `csdl` in `src/adapters/driven/odata/csdlv2/lesen.go`; Scanner prüft nur eine Richtung | `TestKeinPaketImportiertEineAndereVersion`, `TestImportScannerErkenntVerstoss` |
| J4 | Ausgabe `git diff --numstat --diff-filter=MDRT b5ac4a5..HEAD` über `csdl`, `model`, `ports/outbound`, `jsoncodec`, `urlparser`: leer (nur Hinzufügungen) | Bericht, `make test` |

**Raten** (Messender: der Implementer, 30 Prozesse je Zeile, `-test.count=1`, Binary mit `go test -c`, Docker `--cpus=4`):

| Mutation | Zeilen | ohne Detektor (rot von 30) | mit `-race` (rot von 30) |
|---|---|---|---|
| keine (Basis) | I1, I2, I3 | 0, 0, 0 | I2: 0, I3: 0 |
| I1: Bindings aus einer Map gelesen | I1, I2, I3 | 30, 30, 30 | I2: 30, I3: 30 |
| I2: Typenliste des Dokuments als Paketvariable (zu Beginn des Auflösens geschrieben, am Ende gelesen) | I1, I2, I3 | 0, 30, 30 | I2: 30, I3: 30 |
| I3: gemeinsamer Fehlerwert als Zeiger im Paket (`fehler` setzt Ursache und Ort eines Paketwerts) | I2, I3 | 30, 30 | I2: 30, I3: 30 |
| I3: Fehlerwert als Paketvariable zugewiesen und sofort zurückgegeben | I1, I2, I3 | 0, 1, 0 | I2: 30, I3: 30 |

Die Rate eines Datenrennens zählt der Detektor je Prozess (`AGENTS.md` §3.10): 30 von 30 für I2 und I3 in allen Zeilen. Der Wirkungs-Fänger (Vergleich jedes Ergebnisses mit dem Einzelaufruf) fängt die Mutationen mit längerem Fenster in 30 von 30 Prozessen, die letzte Mutation (Zuweisung und Rückgabe in aufeinanderfolgenden Anweisungen) nur in 1 von 30 (I2) und 0 von 30 (I3). Das ist ein akzeptiertes Negativ des Wirkungs-Fängers mit Grund und Folge: das Fenster zwischen Zuweisung und Rückgabe ist kürzer als ein Scheduler-Wechsel, die Wirkung tritt in einem Lauf ohne Detektor kaum auf; die Zusage trägt dort allein `make test-race`. I1 (Reihenfolge einer Map) ist eine Wahrscheinlichkeits-Mutation: 30 von 30 Prozessen rot, je Prozess liest der Test 30 Läufe über dasselbe Dokument (zwei `AssociationSet` mit je drei Navigation-Properties: Wahrscheinlichkeit eines gleich geordneten Laufs 1/36 je Lauf).

**Zählprüfung der Mengen** (Zeile `Z_Zählprüfung_der_Mengen` der Suite, `AGENTS.md` §3.15, §3.17; die Zahlen stehen im Protokoll von `go test -v`): A1 30 Läufe (7 Teilmengen mal Kontext mal Grenze), B4 305 Dokumente (14 Stellen mal 20 Referenzen, Text, CDATA), C1 43, D1 26, D2 50, D3 36 (24 Ringe), D4 35, D5 62, D6 142, D9 60 (28 Paare mal beide Reihenfolgen, dazu 4 Fälle), D11 156 Dokumente der Klasse (Orakel aus den Parametern der Erzeugung; jeder Wert jedes Parameters kommt vor), G1 104 (14 Stellen, davon 91 aus dem Generator und 13 feste Fälle von G6), G2 196 (14 mal 14), N1 36, N3 54, N6 29, alle Dokumente der Suite 1521. Die Mengen sind Schleifen und Generatoren über den Wortlaut der Zusage; die Orakel (`wurzelUrsache`, `d11Dokument`, `g1FaelleV2`, `i1DokV2`) lesen die Parameter der Erzeugung und nie den Leser.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

**Mutationen, Doubles, Aufbau** (Architect):

- Mutation je Zeile steht in der Tabelle; der Implementer fährt sie selbst und nennt je Zusage *Zusage · Mutation · roter Test* (`AGENTS.md` §3.10). Die Klassen-Fixes (`AGENTS.md` §3.17) laufen als Schleife über die Menge des Wortlauts: A1 (7 × 2), A4 (alle Zweiteilungen), B2 (Bereichsgrenzen), D2 (4 × 7), D4 (17), D5 (21 + Klammerformen), D6 (alle Pflichtattribute × 3), D7 (8), D9 (28 Paare), F4 (6 × 2 × 2), G1 (14 Stellen × 3 Arten), G2 (14), G5 (7).
- Generator (D11, G1): deterministisch, ohne Zufall (Aufzählung der Parameter), Orakel aus den Parametern der Erzeugung, nicht aus einem zweiten Lauf des Lesers; Zählprüfung nennt die Zahl der erzeugten Dokumente und je Parameterwert mindestens eines. Das schließt die Lücke `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` für diesen Leser.
- Doubles: keine neuen (`fakes: 0`); `teilReader`, `zerlegtReader`, `endlos`, `beendbar` halten jedes Argument fest und zählen jede Operation, auch `Close` und Reads nach dem Ende.
- Nebenläufig: I2 und I3, Rate je Mutation in 30 von 30 Prozessen; I1 (Map-Reihenfolge) mit Rate je Lauf.
- Die Dokumente der Suite entstehen aus Bytewerten, wo sie Rohbytes oder Escape-Folgen tragen; die Ablage der Datei `metadata_codec_v2_funde.go` und der Fixtures wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt die Dateien.


## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` und der V2-Umfang steht in der Spezifikation (`SPEC-170`, `SPEC-171`; beides erfüllt). Die Tabelle *Zusage · Fänger* steht in §3 (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** keine auf einen offenen Slice. Der Leser steht hinter dem Port des Metadatenlesers (`MetadataCodec`, `SPEC-125`), den `slice-odata-codec-csdl` in `done/` geliefert hat, und braucht das Modell und das Fehlermodell (`done/`); die ADR ändert den Port nicht, und das Modell ist derselbe Typ. Er ruft weder `MockUseCase` noch Workspace, Interpreter oder URL-Parser. Der Welle-Trigger (`welle-mock` in `done/`) trägt für diesen Slice nichts; er ist Reihenfolge-Vorsicht (Priorisierung nach den Muss-Wellen).
- **Was der Slice nicht braucht — und warum:** `welle-mock` und M2 (Leser und Bericht arbeiten auf Bytes und Modell; der Rundlauf über Import und Mock gehört `slice-v2-bootstrap`), `slice-v2-payload-lesen`, `slice-v2-payload-schreiben` und die späteren Slices der Welle, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`. Das heißt: **startbar ohne `welle-mock` und ohne M2, sobald [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted` ist.**
**Verdikt zur Schnittfrage und Größenmarke** (Architect; die Marke misst der Block unten, Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen):

**Kein Schnitt, keine Hebung des XML-Rahmens, kein vorab benannter Teil.** Ein Slice, Marke 8200.

Verglichen mit dem Metrik-Maß von `SPEC-128` (Summe aus Einfügungen und Löschungen, Tests eingeschlossen; *S* = 85·*a* + 80·*b* + 200·*f* + *p*, *M* = 1,35·*S* auf volle Hundert). Zeilen der Tabelle: ungeschnitten 57 (*a* = 2, *b* = 55).

| Option | Slices und Größe | Urteil |
|---|---|---|
| O1 ungeschnitten (gewählt) | *S* = 170 + 4400 + 0 + 1500 = **6070**, *M* = **8200** | trägt: ein Review, eine Verifikation, ein Lieferpunkt für den Rundlauf |
| O2 Hebung des Rahmens als Vorlauf | Vorlauf: *S* etwa 1000 bis 1100, *M* etwa 1400 bis 1500; Rest: 48 Zeilen, *p* etwa 1230, *S* etwa 5080, *M* etwa 6900; Summe der Marken etwa 8400 | verworfen, Gründe unten |
| O3 vorab benannter Teil „Assoziationen“ (N1 bis N7, rund 800 Zeilen) | Rest *M* etwa 7200 | verworfen, Gründe unten |
| O4 Schnitt Lesen / Auflösen | zwei Teile über demselben Baum | verworfen (Plan: teilen den Baum) |

**Warum O2 nicht trägt.**
1. Der Rahmen steckt im V4-Leser nicht als trennbares Stück: `lauf` (`lesen.go`) trägt Strom, Token-Schleife, Grenzen und die Dispatch auf V4-Elemente in einem Zustand. Die Hebung von `strom.go`, der Token-Schleife, `refs`/`hatErsatzRef`, `doppelt` und der Namenshilfen (zusammen rund 330 Zeilen) ändert V4-Dateien; Verschieben zählt in der Metrik doppelt (Löschung plus Einfügung, rund 660 Zeilen) und braucht eine Einhängeschnittstelle (Rückrufe für Start- und End-Element), die es heute nicht gibt.
2. Der Gewinn im Rest ist klein: 330 Zeilen Produktcode (rund 450 auf der Marke). Die Fänger für die Zusagen des Gerüsts (A1 bis B6, E1, E2, I1 bis I3) bleiben notwendig, denn `SPEC-171` Nr. 1 sagt „gilt wortgleich“ für den V2-Leser, und eine Verdrahtung, die `ctx`, Grenze oder `Close` falsch weiterreicht, fängt nur ein Fall am V2-Leser. Ersparnis höchstens zehn Zeilen der Tabelle (rund 800).
3. [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verspricht, dass der V4-Pfad „Byte für Byte der heutige“ bleibt. Die Hebung bricht das ohne Not für ein Kann-Produkt; der Preis eines Regressionsfehlers im Pfad der Abnahme (`LH-FA-06.a`) steht in keinem Verhältnis zu 330 gesparten Zeilen. Die ADR erlaubt die Hebung, verlangt sie nicht.
4. Zwei Slices heißen zwei Runden Review und Verifikation (`durchstich-vor-haertung`: je einmal) bei gleicher Gesamtgröße; das Produkt (V2-Rundlauf) kommt später, nicht früher. Der Slice liegt ohnehin nicht auf dem kritischen Pfad: `slice-v2-bootstrap` braucht auch `slice-v2-payload-lesen`, `slice-v2-query` und `slice-v2-import-codecs`, die parallel laufen.

**Warum O3 nicht trägt.** Assoziationen teilen mit dem Rest den Baum, die Dispatch-Tabelle der Elemente, die Reihenfolge der Auflösungsschritte (`SPEC-171` Nr. 7 d, e) und die Namensräume der Members (Navigation-Properties zählen zur Summe und zu den Namen eines Typs). Ohne sie wäre die `NavigationProperty` eines V2-Dokuments entweder ein Fund oder ein unaufgelöster Name; beides widerspricht `SPEC-171`, und der Lese-Teil von `slice-v2-payload-lesen` braucht die Navigation-Properties im Modell (`SPEC-172` Nr. 3: `Uebersprungen` mit Grund `navigation` statt `unbekannte_property`). Der Teil ist nicht lieferbar ohne den Rest, also „teilt Mechanismus“ im Sinn von `AGENTS.md` §3.14.

**Folge für §4 des Plans.** Die Zeile zur Schnittfrage lautet: *Schnittfrage* (*S* = 6070 ≥ 2223): kein Teil lässt sich ohne geteilten Mechanismus abtrennen — Lesen und Auflösen teilen den Baum, Assoziationen teilen Dispatch, Namensräume und Reihenfolge der Auflösungsschritte, der Rahmen steckt im V4-Zustand `lauf` und seine Hebung änderte den V4-Pfad (Verschieben zählt doppelt, Ersparnis höchstens zehn Zeilen der Tabelle); Vorbild sind `csdl` (*M* 6800) und `json` (*M* 7400), beide ungeschnitten geliefert. Ein vorab benannter Schnitt gilt nicht. Liegt die Messung über der Marke, steht ein Eintrag `ueberschreitung` mit Grund im Block (`SPEC-128` Nr. 5 d), keine Rückführung. Die ältere Marke 8300 (vorläufig, *S* ≈ 6100) ersetzt der Planner durch 8200 mit dem Grund „Tabelle gezählt: 57 Zeilen, *a* = 2, *p* = 1500“.

Erwartung für das Ende (Stand nach dem Review): die Enden der drei Messpunkte lagen bei 0,86 bis 1,33 mal *S*, also 5200 bis 8100 bei *S* = 6070; die Marke 8200 hält alle drei.

```groessenmarke
pfad: src/adapters/driven/odata/csdlv2
pfad: src/adapters/driven/odata/versionstrennung_test.go
pfad: src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_vertrag.go
pfad: src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_lesen.go
pfad: src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_aufloesen.go
pfad: src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_assoziation.go
pfad: src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_funde.go
pfad: src/hexagon/ports/outbound/outboundtest/metadata_codec_v2_modell.go
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_referenz.xml
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_vererbung.xml
pfad: src/hexagon/ports/outbound/outboundtest/testdata/v2_composite.xml
produktcode: 1500
fakes: 0
basis: b5ac4a5
schaetzung: 6070
marke: 8200
schnitt-grund: Lesen und Auflösen teilen den Baum, Assoziationen teilen Dispatch, Namensräume und Reihenfolge der Auflösungsschritte, der Rahmen steckt im V4-Zustand `lauf` und seine Hebung änderte den V4-Pfad (Verschieben zählt doppelt, Ersparnis höchstens zehn Zeilen der Tabelle); Vorbild `csdl` und `json`, beide ungeschnitten geliefert
ersetzt: 8300 | Tabelle gezählt: 57 Zeilen, a = 2, p = 1500; S = 6070 statt der vorläufigen 6100
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Zahl der Marke steht im Block der Größenmarke oben.
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

- Der dokumentierte V2-Umfang war weder im Lastenheft noch in der Spezifikation beschrieben — **Ausgang:** entfallen: er steht in `SPEC-170`.
- Die Menge der V2-Dokumente ist unbeschränkt (jede Bytefolge, jede Zerlegung) — **Ausgang:** offen; Generator über die Klasse mit Orakel und Zählprüfung oder Prüfgrenze als akzeptiertes Negativ mit Grund und Folge im Betrieb (`AGENTS.md` §3.17).
- Echte V2-Dienste liefern Konstrukte, die der Umfang nicht kennt (herstellerspezifische Annotationen, Funktionsimporte); sie dürfen den Bericht füllen, nicht das Lesen brechen — **Ausgang:** offen; Fänger je Art des Fundes.
- Das V4-Modell darf durch den Leser nicht berührt werden (`LH-FA-17.a`): die Tests von `slice-odata-codec-csdl` laufen unverändert, und ein V2-Dokument am V4-Leser bleibt `ODR-E6004` — **Ausgang:** offen; Fänger an beiden Lesern.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation, auch `Close` und jeden Read (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster steht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** eingetreten und entschieden für die sieben offenen Randformen (Tabelle unten); weiter offen: was der Implementer beim Schreiben findet (er hält an, `AGENTS.md` §3.12).

Das Raster des Plans wird um zwei Operationen (Lesen und Auflösen der Assoziationen) und die Spalte *Aufruf wiederholt oder gleichzeitig* erweitert. „Nr.“ ohne Spec-Zahl heißt `SPEC-171`; „Abschnitt 5 Nr. 10 x“ in den Zellen heißt `SPEC-171` Nr. 10 x, „Abschnitt 5, Änderung 1 bis 4“ sind die Schärfungen von `SPEC-171` Nr. 2 c, 5, 7 und von `SPEC-172` *Ungeprüft*; beide stehen vor dem ersten Code-Commit in `spec/spezifikation.md`. Jede Zelle ist entschieden; „gilt nicht“ trägt den Grund.

| Operation | Dokument leer oder ohne Schema | falsche Wurzel oder Version (V4 am V2-Leser) | unbekannter oder nicht unterstützter Typ | fremde Erweiterung | Verweis ohne Ziel, Wiederholung (doppelter Name) | über Grenze (Größe, Tiefe, Summe) | Abbruch (Kontext, Strom) | Aufruf wiederholt oder gleichzeitig |
|---|---|---|---|---|---|---|---|---|
| Lesen des Dokuments | leer, nur Leerraum, nur BOM: `syntax` (Nr. 1, `SPEC-125` Nr. 4); ohne `Schema`: gültig, leeres Modell (Nr. 1); ohne `DataServices`: `struktur` (Abschnitt 5, Änderung 1) | Nr. 2 und 3 (`wurzel`, `version`) | gilt nicht beim Lesen: der Typ wird erst beim Auflösen beurteilt (Nr. 7 d) | Fund `erweiterung`, Ausnahme der drei `m:`-Attribute (Nr. 4) | doppelter Name: `struktur` (Nr. 1, `SPEC-125` Nr. 8 e, Abschnitt 5 Nr. 10 e); Verweis: gilt nicht beim Lesen | Nr. 1 (`SPEC-125` Nr. 5: `zu_gross`, `tiefe`, Namenslängen `struktur`) | Nr. 1 (`SPEC-125` Nr. 1, 2) | Nr. 1 (`SPEC-125` Nr. 14) |
| Lesen der Assoziationen (`Association`, `End`, `AssociationSet`, `NavigationProperty`) | Pflichtangabe fehlt oder leer: `struktur` (Abschnitt 5 Nr. 10 a); Zahl der `End`: Nr. 10 b, c | gilt nicht: Wurzel und Version stehen am Start-Tag der Wurzel fest (Nr. 2) | gilt nicht beim Lesen (Auflösen); `Multiplicity` ungültig: `struktur` (Nr. 10 d) | Fund `erweiterung` am ausgewerteten Element (Nr. 4); `ReferentialConstraint`, `OnDelete`: Fund `konstrukt` | gleiche `Role`, doppelte `Association`, doppelte `AssociationSet`: `struktur` (Nr. 10 b, c, e) | Tiefe und Summe: Nr. 1 (`SPEC-125` Nr. 5), Navigationen zählen zur Summe (Nr. 7) | Nr. 1 | Nr. 1 |
| Auflösen der Namen (Typen, Basis, Schlüssel, Sets) | leeres Modell: nichts aufzulösen, gültig (Nr. 1); Typ ohne Schlüssel: `ohne_schluessel` (Nr. 7) | gilt nicht: Wurzel und Version stehen vor dem Auflösen fest (Nr. 2) | `nicht_aufloesbar` für `Edm.<unbekannt>`, `Collection(…)`, unbekannten Namen; Property bleibt mit Fund `property_typ` bei bekannten nicht unterstützten und deklarierten Typen; `schluesseltyp` (Nr. 5, 7; Abschnitt 5 Änderung 2) | gilt nicht: Erweiterungen werden gemeldet, nicht aufgelöst (Nr. 4) | `nicht_aufloesbar`, Zyklus und `kette` (Nr. 7 → `SPEC-125` Nr. 9) | Kette und Summe (Nr. 7 → `SPEC-125` Nr. 5, 9 a, b) | gilt nicht: das Auflösen prüft den Kontext nicht, einmal davor (Nr. 1, *Ungeprüft*) | Nr. 1 |
| Auflösen der Assoziationen (Navigation, Binding) | keine `Association` im Dokument: jede Navigation `nicht_aufloesbar` (Nr. 6); ohne Container: keine Bindings | gilt nicht (Nr. 2) | `End`-Typ nicht auflösbar oder `ComplexType`: `nicht_aufloesbar`, Ort `<Typ>/<Name>@Relationship` (Nr. 6) | gilt nicht: Erweiterungen werden gemeldet (Nr. 4) | `Relationship`, `Role`, `EntitySet`, `Association` ohne Ziel: `nicht_aufloesbar` (Nr. 6, Nr. 10 f); zwei `AssociationSet` je `Association`: je ein Binding (Nr. 6) | Summe: Nr. 7; Bindings: höchstens 65 536, sonst `zu_viele_properties` (`SPEC-171` Nr. 6, 7); Sets, Funde, Schemas: durch `grenze` begrenzt (`SPEC-125` Nr. 5) | gilt nicht: nicht abbrechbar (Nr. 1, *Ungeprüft*) | Nr. 1 |
| Bericht | leere Menge, nicht `nil` (Nr. 9) | Fehler: Bericht Nullwert (Nr. 9) | Fund `property_typ` (Nr. 4, 5) | Fund `erweiterung` (Nr. 4) | gilt nicht: ein Fehler ist kein Fund (Nr. 9) | Fehler: Bericht Nullwert (Nr. 9) | Fehler: Bericht Nullwert (Nr. 9) | gleiche Bytes, gleiche Reihenfolge der Funde (Nr. 1 → `SPEC-125` Nr. 14) |

**Zusatztabelle** (je Paar von Verstößen am selben Dokument; `AGENTS.md` §3.16). Es gibt keinen Modus.

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| ungültiges XML und falsche Version | die Wurzel entscheidet, sobald ihr Start-Tag gelesen ist; Müll davor ist `syntax` (Nr. 2) | C5, ein Fall je Paar |
| falsche `DataServiceVersion` und Verstoß im Schema dahinter | `version` beim Start-Tag von `DataServices` (Nr. 2 c), davor entscheidet das Dokument in seiner Reihenfolge | C5 |
| Verweis ohne Ziel und nicht unterstützter Typ | `nicht_aufloesbar` gewinnt; der Fund `property_typ` erscheint nicht, der Bericht ist Nullwert (Nr. 7, 9) | D9, ein Fall je Paar |
| Navigation ohne Ziel (Schritt d) und `AssociationSet` ohne Ziel (Schritt e) | Schritt d vor Schritt e, unabhängig von der Reihenfolge im Dokument (Nr. 7) | D9, N6 |
| ungültige `Multiplicity` und gleiche `Role` am selben `Association` | der erste im Dokument: `End` für `End` in der Reihenfolge der Bytes (Nr. 10 b, d) | N3, N7 |
| Pflichtangabe fehlt und Doppelname im selben Element | beide `struktur`; welcher, ist nicht zugesagt (`SPEC-125` Nr. 8 f) | D6, D7 (nur Ursache und Ort des Elements) |
| Größe über der Grenze und Verstoß dahinter | `zu_gross` am Byte `grenze + 1`, davor entscheidet der Verstoß (Nr. 1, `SPEC-125` Nr. 5) | E1, ein Fall je Paar |

**Risiken und Ausgänge des Architect:**

| Risiko | Ausgang |
|---|---|
| Schnittfrage (S ≥ 2223, ungeschnitten) | entfallen: Verdikt oben, Grund in §4; die Marke ist 8200 |
| Sieben Randformen von `SPEC-171` waren offen oder widersprüchlich (`Edm.DateTimeOffset` im V2-Dokument, `DataServices`, Typ eines Entity-Typs als Property, Pflichtangaben, Assoziationen, Reihenfolge der Schritte, `Multiplicity`) | eingetreten und entschieden: `SPEC-171` Nr. 2 c, 5, 7, 10 und `SPEC-172` *Ungeprüft*; weiter offen: nur, was der Implementer beim Schreiben findet (er hält an und gibt zurück, `AGENTS.md` §3.12) |
| Das Orakel des Generators teilt Annahmen mit dem Leser | weiter offen: das Orakel liest die Parameter der Erzeugung, nicht das Dokument; der Verifier bildet ein eigenes Probe-Paket über die Klasse (`AGENTS.md` §3.17) |
| Mehrere `EntityContainer` in einem V2-Dienst (`m:IsDefaultEntityContainer`) sind `struktur` | akzeptiertes Negativ mit Grund und Folge: Abschnitt 5 |
| Viele herstellerspezifische Attribute (`sap:`) ergeben große Berichte | akzeptiertes Negativ: der Bericht wächst linear mit dem Dokument, höchstens `grenze` (`SPEC-125` Nr. 5, *Keine Grenze*); Folge: ein Dokument mit tausenden Funden ist lesbar, der Bericht lang |
| Das V4-Modell und der V4-Leser bleiben unberührt (`LH-FA-17.a`) | weiter offen bis zur Closure; Fänger J2 und J4 |
| Der Test der Versionstrennung (J3) kann von zwei V2-Slices zugleich angelegt werden | weiter offen; Adresse: Planner (die Sendung an `slice-v2-payload-lesen` trägt der Nehmer auf `main` zuerst ein, `AGENTS.md` §3.13); bis dahin legt dieser Slice den Test an |
| `Edm.DateTimeOffset(V2)` als Typname im Modell ist ein Sondername | akzeptiertes Negativ mit Grund und Folge: Abschnitt 5 Änderung 2; kein anderer Teil schlägt ihn nach |

**Rückgaben des Implementers** (`AGENTS.md` §3.12, §3.13; der Code folgt der engeren Lesart, der Architect bestätigt oder ändert sie):

| Rückgabe | Ausgang | Adresse |
|---|---|---|
| `Target` eines Bindings: `SPEC-171` Nr. 6 sagt „gleich dem `EntitySet` des `End`“, Nr. 8 nimmt `SPEC-125` Nr. 11 bis auf vier Abweichungen (darunter nicht die Form von `Target`). Der Code bildet die Form von Nr. 11, `<qualifizierter Container>/<Name>` (`Ns.Container/Categories`); F1, F2, N4 und F4 tragen sie. Wäre nur der Name gemeint, ändern sich diese Fänger und `bindings` in `src/adapters/driven/odata/csdlv2/aufloesen.go` | eingetreten und entschieden: `SPEC-171` Nr. 6 und 8 (die Form von `Target`, wie der Code sie bildet) | – |
| `AssociationSet` und Rollen: `SPEC-171` Nr. 10 f nennt „Rollen der Navigation, sobald eine Navigation auf sie verweist“, und die Zelle N6 trägt diesen Wortlaut. Der Code prüft die Rollen allein gegen `FromRole` und `ToRole` der verweisenden Navigation; ohne Navigation bleibt eine `AssociationSet` mit fremden Rollen gültig (Fall in N5) | eingetreten und entschieden: `SPEC-171` Nr. 10 f (die engere Lesart des Codes) | – |
| Die Menge der V4-Namen in D5 hat nach dem Wortlaut (`Date`, `Duration`, `TimeOfDay`, `Stream`, `Untyped`, `Geography…`, `Geometry…`) 21 Elemente; Tabelle, Mutations-Absatz und Fänger zählen 21 | entschieden | – |
| Review F-1 (Bindings wachsen mit dem Produkt aus Sets und Navigationen), F-5 (Menge ungültiger `Multiplicity`) und I-1 (100 leere Reads): die Spezifikation ist nachgezogen (`SPEC-171` Nr. 1, 6, 7; `SPEC-125` Nr. 2 und *Ungeprüft*); die Zelle *Auflösen der Assoziationen* trägt die Grenze | Spezifikation entschieden; offen: Code (`maxBindings`, Zählung vor dem Bauen), die Fänger (Zeile N8, die Schleife von N3 samt Negativ, der Fänger zu `io.ErrNoProgress`) und ihre Zeilen in §3 | Implementer (Code, Tests und §3 im selben Commit, `AGENTS.md` §3.9) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für Fakes und Hüllen von Strömen; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Leser hat keinen Zustand zwischen Aufrufen; nebenläufig sind allein die Zusagen I2 und I3 (gleichzeitige Aufrufe, Rate in §3), sonst nichts. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
