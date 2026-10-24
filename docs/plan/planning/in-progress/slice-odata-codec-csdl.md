# Slice slice-odata-codec-csdl: OData-Codec, Teil Metadaten (CSDL)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-06`, `LH-RB-01`, `LH-QA-07` (Sendung von `slice-protokolltreue`), [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-06.a` (Metadatenmodell) · `SPEC-002` (OData-Modell) · `SPEC-049` (EDM-Typen) · `SPEC-068` (`ODR-E6004`) · `SPEC-107` (Service-Metadaten) · `LH-RB-01.a` · `LH-QA-07.a` (zweiter Satz) · `SPEC-032` (CSDL V4.0) · `SPEC-107` Nr. 3 (`SetMetadata`) · `ARC-004` (`ODataCodec`) · `ARC-011` · `architecture.md` §2 (Zeile `ODataCodec`, Baseline-Import).
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

**Ziel:** Der Outbound Port `ODataCodec` (Teil Metadaten) und sein Driven Adapter `src/adapters/driven/odata/csdl` lesen EDMX/CSDL V4 in ein technologieunabhängiges Modell aus Schemas, Entity-Typen, Schlüsseln, Properties, Entity-Sets und Navigation-Properties, lösen Referenzen und Vererbung auf und weisen nicht Unterstütztes im Validierungsbericht aus.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Entities und Collections im JSON — `slice-odata-codec-json`; der Teil braucht dieses Modell.
- OData-URL-Parsing und Query-AST — gehört zum Mock; der Port trägt hier nur die Operationen für Metadaten.
- Ausliefern der Metadaten über HTTP — Mock-Modus, nicht in der Welle.
- Serialisieren des Modells zurück nach CSDL — Mock-Modus (`$metadata`-Auslieferung), Welle §6; hier liest der Codec nur.
- OData V2 und EDMX-Versionen außer `4.0` — `LH-FA-17` (Kann), Welle §6; ein Dokument mit anderer Version ist `ODR-E6004` (`LH-FA-06.a`), kein zweites Modell. Lastenheft und Spezifikation geben V4.0 vor (`LH-RB-01`); eine Kompatibilität zu V2 gehört in keine Zeile dieses Slice.
- Der Aufruf von `MetadataRepository.SetMetadata` und das Anlegen der Services im Speicher — der Codec liest Bytes und legt nichts ab; wer ruft, entscheidet der Architect (§3 Auftrag 1 (e)) und trägt die Sendung beim Nehmer ein (`slice-baseline-import` oder `slice-bootstrap-m2`, beide in `open/`; erst danach nennt dieser Punkt ihn, `AGENTS.md` §3.13).

**Übernimmt (Markierung unbekannter Erweiterungen):** Die Markierung unbekannter OData-Erweiterungen als `unsupported_features` (zweiter Satz von `LH-QA-07.a`, `LH-FA-06.a` Nr. 3) liegt in DoD Punkt 3; der Vergleichstest in `slice-protokolltreue` prüft nur die unveränderte Weiterleitung (Sendung von `slice-protokolltreue`, §1 dort, *Ausdrücklich NICHT* trifft sie nicht).

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

- [ ] Port `ODataCodec` (Metadaten) und das Domain-Modell des OData-Dienstes; kein Parsertyp in Signaturen (`architecture.md` §2).
- [ ] CSDL-Adapter: Parsen, Referenzen, Vererbung, EDM-Typen der Abnahmematrix; `ODR-E6004` bei nicht parsebarem CSDL, EDMX-Version ungleich 4.0, Entity-Typ ohne Schlüssel, nicht auflösbarer Referenz und Schlüssel-Property mit nicht unterstütztem Typ.
- [ ] Validierungsbericht mit `unsupported_features` für Properties und Annotationen, die das Modell gültig lassen; trägt die Sendung von `slice-protokolltreue` (Markierung unbekannter OData-Erweiterungen, `LH-QA-07.a`), mit einem Fall je Art (§3 Auftrag 3 (iv)).
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
| `src/hexagon/model/` | update | Modell des OData-Dienstes |
| `src/hexagon/ports/outbound/` | update | `ODataCodec` |
| `src/adapters/driven/odata/csdl/` | neu | CSDL-Leser |
| `spec/spezifikation.md` | update | Vertrag von `ODataCodec` (Metadaten) mit Randformen, Form des Validierungsberichts |
| `spec/architecture.md` | update, falls der Vertrag sie ändert | Zeile `ODataCodec` in §2, `ARC-004`, `ARC-011`, Sequenz Baseline-Import (Berichtigung an jeder Stelle, `AGENTS.md` §3.9; ohne Slices und ADR-Bezüge, `AGENTS.md` §3.4) |
| `src/hexagon/ports/outbound/outboundtest/` | update | Vertragssuite des Ports `ODataCodec`, die jeder Adapter fährt (Muster der vorhandenen `*_vertrag.go`) |
| Tests | neu | Fixtures echter CSDL-Dokumente und Negativfälle je Fehlerbedingung nach der Menge der Tabelle *Zusage · Fänger* (§3 Auftrag 3), Fixtures der Randformen aus Bytewerten; Vertragstest am Port; Fakes nur, wo die Tabelle sie festlegt (§4) |

Der Vertrag, die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen in der Spezifikation (Produkt, kein Werkzeug) vor dem ersten Code-Commit (`AGENTS.md` §3.12), nicht vor dem Start; die Randformen, die sie tragen müssen, nennt §6. Reihenfolge der Rollen nach dem Start, jede erst, wenn die vorige übergeben hat: Architect (Vertrag, Modell, Entscheidungen aus Auftrag 1, Tabelle, Raster, Nachzüge aus Auftrag 5, Sendungen an Nehmer) → Planner (Marke aus der Tabelle, §4, eigener Commit; Prüfung der Adressen nach `AGENTS.md` §3.13; Schnittfrage, wenn die Schätzung × 1,25 über 3000 liegt) → Implementer (erster Code-Commit erst nach der Marke; trägt Tabelle *Zusage · Mutation · roter Test* und die Raten in einen Absatz von §3 ein, `AGENTS.md` §3.10) → Reviewer → Verifier → Planner (Closure). Der letzte Commit des Architect, der Vertrag, Tabelle und Raster abschließt, ist die `<Basis>` der Marke (§4). Hält der Implementer an einer Zelle ohne Ort an (`AGENTS.md` §3.12), geht sie an den Architect, nicht in den Code.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; Pfad `spec/spezifikation.md`, `spec/architecture.md`; Kennungen nummeriert fortlaufend je Datei, `SPEC-<NNN>` neu oder als Ergänzung einer bestehenden):

1. **Vertrag der Codec-Schnittstelle (Metadaten).** (a) Zuschnitt des Ports: `ODataCodec` führt nach `architecture.md` CSDL, URL und JSON in einer Schnittstelle, `slice-odata-codec-json` fügt Methoden hinzu; entscheide, ob der Port eine Schnittstelle bleibt (dann ändert der JSON-Slice einen Port aus `done/`) oder je Teil eine kleine Schnittstelle (ISP), und schreibe die Folge in `slice-odata-codec-json` §1 und DoD (Sendung, `AGENTS.md` §3.13). (b) Signatur: Eingabe (Bytes oder Strom; bei Strom Verhalten bei Lesefehler, Größengrenze), `ctx` (nil und beendet, an welchen Stellen geprüft: vor dem Lesen, während des Parsens), Rückgabe (Modell und Validierungsbericht als Werte, Verhalten des Berichts im Fehlerfall: nur Fehler oder Fehler samt Teilbericht), Fehlerform (`*Fehler` mit `ODR-E6004` und der Ursache, welche der fünf Bedingungen eintrat, mit Ort im Dokument; Kontextfehler ohne Code wie in `SPEC-124` Nr. 2). Kein Parsertyp in Signaturen (`architecture.md` §2). (c) Modell nach `SPEC-002`: erweitere `ServiceModel`, `EntityType`, `Property` um das, was die Folge-Slices lesen (Entity-Set → Entity-Typ, wirksame Schlüssel und Properties **einschließlich geerbter**, Property-Typ und `nullable`, Navigation-Properties samt Ziel, Basistyp, `abstract`, offene Typen, Composite-Schlüssel mit Reihenfolge, Name qualifiziert und nach Auflösung der Aliase), mit Zugriffen, die `slice-odata-codec-json` und `slice-baseline-import` brauchen (Lese-Sicht, unveränderlich nach der Rückgabe; Verhalten bei unbekanntem Typ oder Set). Typ- und Namensvalidierung aus `SPEC-002` (Form eines Namens, Länge) als Randformen. (d) Form des Validierungsberichts: `unsupported_features` mit Art (Property-Typ, Annotation, Konstrukt), Ort (qualifizierter Name, Property) und Grund, Reihenfolge, Verhalten bei doppelten Funden, Verhalten bei einem Dokument ohne Fund (leerer Bericht, kein `nil`-Unterschied), Bezug zu `ODR-E6001`/`ODR-E6002` der späteren Mock-Zugriffe (nur Verweis, kein Verhalten hier). (e) **`SetMetadata`-Lücke:** wer legt Services und ihre Metadaten im Speicher an und wann, und was ist `metadata_hash` (Hash über die Bytes des Dokuments oder über das Modell, Verfahren, Ort der Berechnung); entscheide den Aufrufer von `MetadataRepository.SetMetadata` (§6 Risiko 2), trage die Sendung beim Nehmer in §1 und DoD ein (Kennung dieses Slice) und lies vorher dessen *Ausdrücklich NICHT* (`AGENTS.md` §3.13). Kandidaten, beide in `open/`: `slice-baseline-import`, `slice-bootstrap-m2`; passt keiner, geht die Entscheidung an den Planner zurück (kein dritter Slice ohne neue Zerlegung). Der Aufrufer bekommt keinen Code in diesem Slice.
2. **Randformen als Raster** *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), jede Zelle mit Ort in der Spezifikation oder „gilt nicht“ mit Grund, eine leere Zelle ist ein Fehler. Operation: das Parsen (ggf. getrennt: Prüfen der Eingabe, Auflösen, Bericht). Argumentarten mindestens: `ctx` nil und beendet (vor und während des Parsens); Eingabe fehlend (`nil`), leer, nur Leerraum, nur BOM; kein XML; abgeschnitten mitten im Element; Lesefehler des Stroms nach dem ersten Byte; Kodierung (UTF-8 mit und ohne Deklaration, UTF-16, ISO-8859-1, unbekannte Deklaration); `DOCTYPE` und Entitäten (interne, externe, Entitätsangriff), Verhalten bei Verarbeitungsanweisungen und Kommentaren; falsches Wurzelelement; EDMX-Namensraum fremd (V2/V3: `http://schemas.microsoft.com/ado/2007/06/edmx`) und `Version` fehlend, leer, `4.0`, `4.01`, `1.0`, `3.0`, `4`, mit Leerraum; mehrere `Schema` (gleicher und verschiedener Namensraum), `Alias`, `edmx:Reference` und `Include` (externe Dokumente werden nicht geladen), `DataServices` fehlend oder doppelt; Wiederholung (gleiche Eingabe, gleiche Ausgabe, Eingabe wird nicht verändert) und gleichzeitige Aufrufe (Auftrag 6). Zusatztabelle mit der Spalte *Fänger* je Paar von Verstößen am selben Dokument (ein Dokument verletzt mehrere der fünf Bedingungen: welche meldet der Fehler, Rangfolge) und je Modus; die Allaussage des Vertrags „nur diese fünf Bedingungen sind `ODR-E6004`, alles andere im Modell gültig“ nennt ihren Fänger über alle Fälle.
3. **Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15) mit einer Zeile je Zusage, Randform und Fall, **die Menge vor dem Fänger**, jede Allaussage als Schleife über beide Mengen, nicht als Liste der Fälle, die einfielen: (i) **Fehlerbedingungen × Eingabeformen**: jede der fünf Bedingungen (nicht parsebar, Version ungleich 4.0, Entity-Typ ohne Schlüssel, nicht auflösbare Referenz, Schlüssel-Property mit nicht unterstütztem Typ) ist `ODR-E6004` mit eigener Ursache, an jeder Stelle, an der sie auftreten kann (ein Entity-Typ ohne Schlüssel als erster, mittlerer und letzter Typ, in zweitem Schema, als abgeleiteter Typ, der den Schlüssel erbt: kein Fehler); **nicht auflösbare Referenz** als Menge der Referenzarten (`BaseType`, Typ einer Property, Typ einer Navigation-Property, `EntityType` eines Entity-Sets, `Target` eines Navigation-Property-Binding, Alias), ein Fall je Art, ungebundener Name mit Alias und ohne; (ii) **EDM-Typen**: die acht der Abnahmematrix (`SPEC-049`) einzeln als unterstützt, **jeder andere Primitivtyp des V4-Standards** (`Edm.Binary`, `Byte`, `SByte`, `Int16`, `Single`, `Double`, `Duration`, `TimeOfDay`, `Stream`, die `Geography*`- und `Geometry*`-Typen) einzeln als nicht unterstützt, dazu `Collection(...)`, komplexe Typen, Aufzählungstypen und `Edm.Untyped`: als Property lässt der Typ das Modell gültig und erscheint im Bericht, als Schlüssel-Property ist er `ODR-E6004`; eine Schleife über alle, die Fixture gleicht nie der Mutation (zwei Typen reichen nicht); (iii) **Vererbung**: wirksame Schlüssel und Properties aus Basistyp, mehrstufig; Zyklus (Selbstbezug, Zyklus über zwei und drei Typen) als `ODR-E6004` (nicht auflösbar) oder eigene Ursache; abstrakter Basistyp; Schlüssel nur im Basistyp (geerbt), Schlüssel im abgeleiteten Typ trotz Schlüssel im Basistyp (nach OData V4 unzulässig: entscheide), Überschreiben einer Property; (iv) **Bericht**: jede Annotation-Art und jedes Konstrukt, das der Leser nicht auswertet (Annotation mit unbekanntem Namensraum, Attribut aus fremdem Namensraum, `Function`, `Action`, `EntityContainer`-Elemente außer Entity-Set und Singleton, `ComplexType`, `EnumType`, `TypeDefinition`, `OpenType`), ein Fall je Element der Menge, ob es im Bericht erscheint oder als „ignoriert, nicht gemeldet“ gilt (dann akzeptiertes Negativ mit Grund und Folge, `LH-RB-01`: nicht Unterstütztes erscheint nie als unterstützt); die Sendung von `slice-protokolltreue` (Markierung unbekannter Erweiterungen) hat hier ihren Fänger; (v) **Stabilität**: Modell und Bericht in der Reihenfolge des Dokuments, Wiederholung desselben Dokuments ergibt dasselbe Ergebnis (Fänger bildet die Rate je Prozess bei Map-Iteration, `AGENTS.md` §3.10, oder die Reihenfolge ist konstruiert und der Fänger fährt `-count=20`); (vi) **Fixtures** aus echten CSDL-Dokumenten (mindestens: ein Referenzdienst mit Entity-Sets und Navigation, ein Dokument mit Vererbung, ein Dokument mit Composite-Schlüssel) und Fixtures der Randformen, **aus Bytewerten gebaut** (`BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um`: BOM, UTF-16, Zeichenreferenzen), Grenzwerte beider Seiten jeder Grenze (Auftrag 4); (vii) **Port-Vertragstest** im Stil von `src/hexagon/ports/outbound/outboundtest/`: eine Suite, die jeder Adapter des Ports fährt, damit `slice-odata-codec-json` und `slice-baseline-import` gegen einen Fake mit demselben Vertrag arbeiten können; nenne den Fake (Zählung jeder Operation, jedes Argument festgehalten) nur, wenn dieser Slice ihn liefert, sonst den Träger. Aufrufstellen des Ports als Menge: in diesem Slice nur der Vertragstest, die Aufrufer folgen in den Folge-Slices (Sendungen, `AGENTS.md` §3.13). Zu jeder Zeile die Mutation, die sie rot macht, und der Fall, in dem zwei verglichene Werte gleich sind (Version `4.0` gegen `4.00`, gleicher Name in zwei Schemas, gleicher Typ in Basis und Ableitung).
4. **Prüfgrenzen und akzeptierte Negative** (Lerneintrag von `slice-session-beenden`: eine Klasse mit unbeschränktem Umfang hat kein letztes Mitglied, die Frage ist, wo die Prüfung endet): für Größe des Dokuments (Beziehung zu `MAX_CAPTURE_BODY`, `SPEC-013`: der Capture begrenzt den Body, der Import liest ihn zurück), Zahl der Typen und Properties, Länge einer Vererbungskette, Tiefe der Verschachtelung der Elemente, Länge von Namen und Werten: je Klasse die Grenze, die der Leser durchsetzt, oder „keine, weil …“ mit Grund und Folge im Betrieb (Speicher, Laufzeit), der Fehlercode bei Überschreitung (`ODR-E6004` mit Ursache oder eigene Entscheidung), der Fänger beider Seiten der Grenze (Grenze und Grenze + 1) und die Zahl der Fixture ungleich der Zahl der Mutation. Jedes akzeptierte Negativ nennt Grund und Folge (`AGENTS.md` §3.15), jede Fänger-Menge, die nur abgetastet wird, sagt es.
5. **Nachzug an jeder Stelle, die dieselbe Aussage trägt** (`AGENTS.md` §3.9): `LH-FA-06.a` Nr. 2 und 3, `SPEC-002`, `SPEC-049` (EDM-Zeile), `SPEC-068` (Zeile `ODR-E6004`), `SPEC-107` Nr. 3, `LH-RB-01.a`, `SPEC-032`, die Zeile `ODataCodec` in `architecture.md` §2 und `ARC-004`, `ARC-011`, die Sequenz Baseline-Import in `architecture.md` (`ImportUseCase` interpretiert über den Codec); der Fehlertext in `SPEC-006`/Fehlermodell, wenn die Ursache des `ODR-E6004` ein Feld trägt; `src/hexagon/model/fehler.go` ändert der Architect nicht (Code). Ohne Slices, Wellen und ADR-Bezüge in `architecture.md` (`AGENTS.md` §3.4).
6. **Nebenläufigkeit und Bibliothek.** (a) Ist der Leser ein Wert ohne Zustand (kein Paket-Zustand, kein gemeinsamer Cache), steht die Zusage „gleichzeitige Aufrufe sind erlaubt und beeinflussen sich nicht“ mit Fänger unter `make test-race` (je Prozess, `-test.count=1`) und Wirkungs-Fänger (gleichzeitige Aufrufe mit verschiedenen Dokumenten, jedes Ergebnis gleich dem sequenziellen), oder sie steht als akzeptiertes Negativ mit Grund und Folge; die Zeile ist dann nebenläufig im Sinn von `AGENTS.md` §3.14, sonst nicht (die Entscheidung bestimmt den Faktor der Marke). Zweige, die der Detektor-Test ausführt: aus dem Abdeckungsprofil, nicht aus der Erinnerung (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`). (b) Bibliothek: `encoding/xml` der Standardbibliothek genügt oder eine ADR vor dem Code (§6 Risiko 1); gilt `encoding/xml`, die Zusagen zu `DOCTYPE`/Entitäten, Kodierungen (`CharsetReader`) und Namensräumen stehen im Vertrag, nicht im Verhalten der Bibliothek. (c) Abgrenzung zu den Teilen URL und JSON: der Vertrag nennt, was dieser Slice nicht festlegt (`slice-odata-codec-json`).

**Der Planner hat nichts davon entschieden.** Die Aufträge benennen die Entscheidungen und die Mengen, die der Architect vor dem Code schließen muss; jede offene Entscheidung steht in §6 mit dem Ausgang *offen*.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-fehlermodell-konfiguration` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice steht an der Spitze des Codec-Asts (`welle-record-replay` §5) und braucht keinen Slice des Record-Asts oder der Sessions.

**Stand des Starts (2026-10-24):** Der Start-Trigger ist erfüllt: `slice-fehlermodell-konfiguration` liegt in `done/` (`ODR-E6004` ist als `CodeE6004` im Fehlermodell, Ziel `ZielImport`, Status 422); die Welle ist offen und trägt keinen weiteren Slice dieses Asts davor. Der Vertrag, die Tabelle und das Raster stehen noch aus (Architect, §3).

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Der Planner setzt sie nach dem Vertrag des Architect und vor dem ersten Code-Commit auf die Schätzung aus der Tabelle × 1,25, aufgerundet auf volle Hundert, und ersetzt die Erwartung unten dabei mit Grund; liegt die Schätzung × 1,25 über 3000, schneidet er vor dem ersten Code-Commit (lässt sich kein Teil ohne geteilten Mechanismus abtrennen, steht der Grund hier). Die Schätzung zählt die Zeilen der Tabelle (× 85 nebenläufig, × 40 jede andere; der Leser ist voraussichtlich eine reine Funktion über Bytes, der Architect sagt in §3 Auftrag 6, ob eine Zeile nebenläufig ist) und 200 je Port-Fake oder Test-Server; Fixtures echter CSDL-Dokumente zählen als Tests mit (Pfade aus §3). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der letzte Commit des Architect, der Vertrag, Tabelle und Raster vor dem Code abschließt (Dokumente zählen nicht mit). Der Slice ist der erste mit einem Parser fremder Bytes: sein gemessener Wert in §4 ist der erste, den die Formel für diese Art Code hat.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Größenmarke oben (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Erwartung des Planner, vorläufig, vor der Tabelle des Architect:** rund 1200 Zeilen in derselben Metrik, Marke 1700 aus der älteren Fassung dieses Plans; beide gelten, bis die Marke gesetzt ist, und ersetzt sie dann mit Grund (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): eine Bibliothek für CSDL wird nötig und hat keine ADR (siehe §6).

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

- Bibliothek: Die Standardbibliothek (`encoding/xml`) reicht für EDMX, eine Drittbibliothek braucht eine ADR vor dem Code (`AGENTS.md` §3.6 gilt sinngemäß, wie bei [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)) — **Ausgang:** offen; Architect entscheidet vor dem Code.
- Wer die Metadaten eines Service ablegt (`MetadataRepository.SetMetadata`, `SPEC-107` Nr. 3) ist nicht festgelegt: Die Spezifikation nennt das Erfassen von `$metadata` im Modus `record` und die Prüfung beim Import, aber keinen Aufrufer und keinen Zeitpunkt; kein Code ruft `SetMetadata` außerhalb der Vertragstests, und `slice-record-service` legt `$metadata` nur als Austausch ab (Kennzeichen `Metadata` der Aufnahme), nicht im Service — **Ausgang:** offen; Architect entscheidet vor dem Code (§3 Auftrag 1 (e)) und weist den Aufruf einem Slice zu, der in `open/` liegt und die Sendung annimmt: Kandidaten `slice-baseline-import` (Metadaten beim Import, dort §6 *Metadaten des Service* hängt an derselben Entscheidung; `ODR-E6004` beim Import) und `slice-bootstrap-m2` (Composition Root, der die Services anlegt, dort §6); `slice-record-service` liegt in `done/` und ist keine Adresse (`AGENTS.md` §3.13). Geprüft beim Start: kein *Ausdrücklich NICHT*-Punkt beider Kandidaten trifft die Sendung. Die Sendung steht beim Nehmer in §1 und DoD, mit der Kennung dieses Slice, im selben Commit wie die Entscheidung.
- Sendung von `slice-protokolltreue` (§1 dort, *Ausdrücklich NICHT*): die Markierung unbekannter OData-Erweiterungen als `unsupported_features` (`LH-QA-07.a` zweiter Satz) liegt hier; trägt dieser Slice sie in DoD Punkt 3, ist sie angenommen — **Ausgang:** offen; der Fänger ist eine Zeile der Tabelle *Zusage · Fänger* (§3 Auftrag 3 (iv)); beim Closure ist sie *eingetreten* oder geht an den Planner zurück.
- Eingabe aus einem fremden Dienst: XML mit DOCTYPE und Entitäten, anderer Kodierung als UTF-8, Byte-Order-Mark, abgeschnittenem Dokument oder einem Reader-Fehler mitten im Strom — **Ausgang:** offen; Zeilen im Raster, Ort in der Spezifikation (§3 Auftrag 2).
- Ausgabe in stabiler Reihenfolge: die Standardbibliothek iteriert Maps zufällig; ein Modell oder Bericht, dessen Reihenfolge von einer Map abhängt, ändert sich zwischen Läufen und damit der Hash der Metadaten, wenn er aus dem Modell gebildet wird — **Ausgang:** offen; Fänger mit der Rate je Prozess und je Lauf, Zahl und Messender (`AGENTS.md` §3.10; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 2×), oder die Reihenfolge ist durch Konstruktion festgelegt (Quellreihenfolge des Dokuments) und der Fänger fährt Wiederholungen (§3 Auftrag 3 (v)).
- Größe (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`, 13×, verkörpert in `AGENTS.md` §3.14): Erwartung des Planner 1200, vorläufig und älter als die Formel (§4) — **Ausgang:** offen; der Planner setzt die Marke nach der Tabelle des Architect.
- Bericht des Implementers (`BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`, 3×, verkörpert in `AGENTS.md` §3.10): der Implementer trägt Zusage, Mutation, roten Test und die Raten (Zahl der Prozesse, Messender) in einen Absatz von §3 ein, damit Reviewer und Verifier ihn haben — **Ausgang:** offen; fehlt der Absatz im Review, ist das ein weiterer Beleg gegen die verkörperte Regel und geht an den User.
- Prüfgrenzen (Lerneintrag von `slice-session-beenden`): die Klassen mit unbeschränktem Umfang dieses Slice (Zahl der Typen, Länge einer Vererbungskette, Größe des Dokuments, Länge der Namen) haben kein letztes Mitglied — **Ausgang:** offen; die Spezifikation benennt je Klasse die geprüfte Grenze mit Grund und Folge im Betrieb, und ein Fänger führt beide Seiten der Grenze (§3 Auftrag 4).
- Unbekannte Namensräume, mehrere Schemas, Aliase, Zyklen in der Vererbung, abstrakte Typen, offene Typen, Composite-Schlüssel, Schlüssel aus Navigation — **Ausgang:** offen; Raster *Konstrukt × Behandlung* in §6.
- Größe und Tiefe: ein Dokument mit sehr vielen Typen oder tiefer Vererbung — **Ausgang:** offen; Grenzen für Eingabegröße und Tiefe benennt die Spezifikation oder die Zelle sagt „gilt nicht“ mit Grund.
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

**Vorgelagert — offene Beobachtungen sichten:** Das Register unter `../observations/BEO-REPO/` trägt beim Start vierzehn Einträge (Sub-Area `*`; Stand 2026-10-24 nach der Closure von `slice-session-beenden`); dieser Plan antwortet auf jeden (Zahlen aus den Dateien unter `evidence/`, Ausgänge aus den `state.md`). Der Slice liefert einen neuen Vertrag über ein fremdes Format; er ist als Leser ohne eigenen Zustand voraussichtlich nicht nebenläufig (der Architect entscheidet, §3 Auftrag 6):
- `randform-erst-im-review-entschieden` (17×, verkörpert in `AGENTS.md` §3.16) und `universalaussage-im-vertrag-ungeprueft` (6×, verkörpert in `AGENTS.md` §3.16): das Raster in §6 vor dem ersten Code-Commit; Allaussagen („jede Bedingung ist `ODR-E6004`“, „jeder nicht unterstützte Typ erscheint im Bericht“) nennen ihren Fänger über alle Fälle (§3 Auftrag 3).
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (22×, verkörpert in `AGENTS.md` §3.15, geschärft in der Closure von `slice-session-use-case`): der Vertrag nennt nur, was ein Fänger der Tabelle prüft; die Aufrufstellen des Ports sind eine Menge mit einem Fall je Stelle (hier: der Vertragstest am Port; Aufrufer im Code gibt es erst in den Folge-Slices), eine Allaussage über Stellen × Arten ist eine Schleife über beide Mengen (§3 Auftrag 3).
- `akzeptiertes-negativ-ohne-folge` (7×, verkörpert in `AGENTS.md` §3.15): was der Vertrag ungeprüft lässt (Konstrukte, die der Leser ignoriert; Grenzen), steht mit Grund und Folge im Betrieb; §3 Auftrag 4.
- `fix-deckt-die-genannte-mutation-nicht-die-klasse` (2×, offen) und `zweig-ohne-lauf-unter-dem-detektor` (1×, offen): unter der Schwelle, kein Ausgang zu vergeben; ein dritter Beleg löst beim Lese-Schritt der Closure die Verkörperung aus. Für diesen Slice: ein Fix zu einem Befund deckt die Klasse (Stellen × Arten: alle nicht unterstützten EDM-Typen, alle Referenzarten, alle Fehlerbedingungen), nicht die genannte Mutation; und nur wenn der Leser nebenläufig ist, bildet der Implementer die Zweige unter dem Detektor aus dem Abdeckungsprofil, nicht aus der Erinnerung (§3 Auftrag 3, 6).
- `panik-eines-fakes-verdeckt-den-benannten-test` (2×, offen): unter der Schwelle; betrifft Fakes in Goroutinen, die dieser Slice voraussichtlich nicht hat (kein Fake, §4); trifft einer zu, hält ein Fake keine Panik in einer Goroutine, die den Test mit dem Namen der Zeile überdeckt.
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (13×, verkörpert in `AGENTS.md` §3.14): die Marke setzt der Planner nach der Tabelle des Architect, vor dem ersten Code-Commit (§4); die ältere Erwartung 1200 ist vorläufig und wird dabei mit Grund ersetzt.
- `berichtigung-erreicht-nicht-jede-stelle` (8×, verkörpert in `AGENTS.md` §3.9): die Liste der Stellen steht in §3 Auftrag 5; die Sendung zu `SetMetadata` steht beim Nehmer in §1 und DoD mit der Kennung dieses Slice, bevor dieser Plan ihn nennt (`AGENTS.md` §3.13).
- `nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, verkörpert in `AGENTS.md` §3.10, Sensor `make test-race`) und `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen): Bezug nur über die Ausgabe in stabiler Reihenfolge (Iteration über eine Map, §3 Auftrag 3 (v)) und eine etwaige Zusage über gleichzeitige Aufrufe (§3 Auftrag 6); ein Fänger, der nur mit Wahrscheinlichkeit rot wird, nennt die Rate je Prozess und je Lauf mit Zahl und Messendem (`AGENTS.md` §3.10).
- `bericht-der-rolle-liegt-nicht-im-repo` (3×, verkörpert in `AGENTS.md` §3.10): Risiko in §6; der Implementer trägt Tabelle *Zusage · Mutation · roter Test* und die Raten in §3 ein.
- `befund-an-rolle-ohne-artefakt` (2×, offen): unter der Schwelle; jede offene Entscheidung steht in §6 mit dem Ausgang *offen* und der Adresse *Architect*, einer Rolle mit Artefakt (dieser Plan).
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×, offen): Bezug über XML-Fixtures mit Zeichenreferenzen, Escapes und Kodierungen (§3 Auftrag 3 (vi)); sie werden aus Bytewerten gebaut, nicht als Text mit Escape-Folgen geschrieben, und nach dem Schreiben gegengelesen.

**Stand des Starts (2026-10-24):** Das Register trägt die Zahlen oben; `slice-session-beenden` ist geschlossen und hat `fix-deckt-die-genannte-mutation-nicht-die-klasse` auf 2×, `zweig-ohne-lauf-unter-dem-detektor` auf 1× und `panik-eines-fakes-verdeckt-den-benannten-test` auf 2× gebracht (Lerneintrag dort).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
