# ADR-0011: Versionierter Adapter für OData V2 — versionsneutrale Ports, Implementierungen je Version, ein Satz je Service

**Status:** Proposed

**Datum:** 2026-11-01

**Autor:** pt9912

**Bezug:** [`LH-FA-17`](../../../spec/lastenheft.md#lh-fa-17--weitere-odata-versionen-kann), [`LH-RB-01`](../../../spec/lastenheft.md#lh-rb-01--odata-kompatibilität), [`LH-QA-05`](../../../spec/lastenheft.md#lh-qa-05--bedienbarkeit-und-fehlerdiagnose), [`LH-FA-18`](../../../spec/lastenheft.md#lh-fa-18--recording-daten-in-mock-baseline-importieren-muss), [ADR-0002](0002-architektur-gate-a-check.md)

**Schärft:** [`LH-FA-17.a`](../../../spec/spezifikation.md#lh-fa-17a--odata-v2-kann), [`SPEC-170`](../../../spec/spezifikation.md#spec-170--odata-v2-umfang-version-je-service-und-abbildung-auf-das-modell), [`SPEC-175`](../../../spec/spezifikation.md#spec-175--odata-v2-version-je-service-wahl-der-implementierungen-und-import), [`ARC-004`, `ARC-011`, `ARC-012`](../../../spec/architecture.md#1-komponenten-übersicht)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

`LH-FA-17` verlangt OData V2 als optionalen, eigenen Kompatibilitätsumfang; `LH-FA-17.a` und der Erweiterungspunkt der Architektur verlangen einen versionierten Adapter, der das V4-Domain-Modell nicht verunreinigt und V2 und V4 nie in einem Service-Modell mischt.
Heute gibt es je Darstellung einen Port (`MetadataCodec`, `ODataCodec`, `URLCodec`, die Schnittstelle zum Schreiben von OData-JSON) mit genau einer Implementierung, die Typen des Modells sind versionsfrei benannt (`ServiceModel` trägt ein Feld `version`, der Query-AST kennt nur Formen von Literalen), und der Speicher hält `odata_version` je Service, die der Composition Root fest auf `4.0` setzt.
Import und Mock nehmen heute genau ein Paar von Codecs entgegen.
Offen sind drei Fragen: wie viele Port-Familien es gibt, wo die Version eines Service lebt und wie Import und Mock zu den Codecs ihres Service kommen.
Annahmen: (1) Der V2-Umfang lässt sich auf die vorhandenen Begriffe des Modells abbilden; was dort keine Entsprechung hat, liegt außerhalb des Umfangs und erscheint als Fund oder Kompatibilitätsfehler. (2) Ein Service hat zu jeder Zeit genau eine Version.
Kippt (1), weil ein V2-Konstrukt im Umfang ein neues Modellelement braucht, kippt die Entscheidung gegen Option D; kippt (2), kippt sie ganz.

## Entscheidung

Wir wählen **versionsneutrale Ports, eine Implementierung je Version und einen Satz von Implementierungen je Service**.
Die Ports und das Domain-Modell bleiben, wie sie sind; V2 ist je Darstellung eine zweite Implementierung in einem eigenen Paket unter `src/adapters/driven/odata/`, und kein V2-Paket importiert ein V4-Paket oder umgekehrt.
Die Version gehört dem Service: Die Startkonfiguration deklariert sie, der Speicher hält sie in `odata_version`, der Composition Root wählt daraus den Satz der Implementierungen und gibt jedem Dienst (Import, Mock) den Satz seines Service; kein Dienst verzweigt nach der Version, und jede V2-Implementierung lehnt ein Modell anderer Version ab.
Eine Version wird nie aus einem Dokument, einer Antwort oder einem Header erkannt; ein Dokument anderer Version als der deklarierten ist ein Fehler des Lesers.
Der Import bleibt ein Dienst für beide Versionen, sein Versionsunterschied liegt allein im Satz seines Service.
Die Regeln im Einzelnen (Umfang, Abbildung, Fehlerklasse, Konfiguration) stehen in `SPEC-170` bis `SPEC-177`.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — nichts tun (V2 nicht liefern) | kein Aufwand, kein Risiko für den V4-Pfad | `LH-FA-17` (Kann) bleibt unerfüllt; der Erweiterungspunkt bliebe unbelegt |
| B — eigene Ports je Version (`MetadataCodecV2` usw.) | Version steht im Typ der Schnittstelle | jede Schnittstelle doppelt; Import und Mock bräuchten für jede Version eigene Pfade oder eine Verzweigung; der Mock-Kern müsste V2 kennen |
| C — Verzweigung nach Version in den V4-Paketen | keine neuen Pakete | V4-Pakete lernen V2; jede V2-Änderung berührt den V4-Pfad und seine Tests; `LH-FA-17.a` („ohne Verunreinigung“) verfehlt |
| D — eigenes V2-Modell und eigener Mock-Pfad | maximale Trennung | der Kern (Interpreter, Workspace, Coverage, Snapshot) entstünde doppelt; Aufwand weit über dem Kann-Umfang |
| E — ein Verteiler hinter jedem Port, der nach Version entscheidet | Dienste bleiben unverändert | `ParseMetadata` und `ParseURL` kennen weder Service noch Modell, der Verteiler müsste die Version heimlich über den Kontext erhalten |
| F — Version aus dem Dokument oder der Antwort erkennen statt deklarieren | keine Konfiguration | `LH-FA-17` spricht vom „als V2 deklarierten Dienst“; eine Erkennung mischt Versionen still, sobald ein Dienst beide Formen liefert |
| **G — versionsneutrale Ports, Implementierung je Version, Satz je Service (gewählt)** | V4-Pfad und Modell unberührt; V2 ist ein Verzeichnis und lässt sich entfernen; Dienste bleiben versionsfrei; jede Version einzeln testbar | ein Satz-Typ im Port-Paket und eine Wahl im Composition Root sind neu; V2-Konstrukte ohne Entsprechung im Modell bleiben außerhalb des Umfangs; die Dienste für Import und Mock müssen den Satz je Service entgegennehmen |

## Konsequenzen

- Positiv: Der V4-Pfad (Mock, Replay, Record) bleibt Byte für Byte der heutige; seine Tests laufen unverändert. Ein V2-Service nutzt Interpreter, Workspace, Coverage und Snapshot des Kerns. Mehrere Services verschiedener Versionen laufen in einem Prozess, weil jeder seinen Satz hat und die Implementierungen zustandslos sind.
- Negativ: Das V2-Modell ist ein V4-förmiges Modell mit Version `2.0`; was V4 nicht kennt (komplexe Typen, Funktionsimporte, Atom), ist nicht Teil des V2-Umfangs. Zwei Implementierungen derselben Mechanik (Stromlesen mit Grenze, XML-Rahmen, JSON-Scanner) können doppelt entstehen; ein versionsfreies Paket für das Gemeinsame darf entstehen, wenn die V4-Tests dabei unverändert bleiben, verlangt wird es nicht.
- Folgepflicht: Vertrag und Umfang in `SPEC-170` bis `SPEC-177` und der Nachzug der Architektur-Sicht; der Importer nimmt den Satz je Dienst entgegen (Änderung eines gelieferten Vertrags, eigener Slice vor der Verdrahtung); Import und Mock-Dienste, die noch nicht gebaut sind, nehmen den Satz je Dienst von Anfang an entgegen, damit der Mock-Kern nicht nachträglich geändert wird; der Composition Root hält die Tabelle Version → Satz; das Handbuch nennt den V2-Umfang und die Fehlerklasse.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `go test` über die Importe der Pakete unter `src/adapters/driven/odata/` | kein V2-Paket importiert ein V4-Paket und kein V4-Paket ein V2-Paket; Gemeinsames liegt in einem Paket ohne Versionsnamen | `make test` |
| `go test` der V4-Pakete und des Mock-Kerns ohne Änderung | der V4-Pfad verhält sich nach dem Einbau von V2 wie vorher | `make test` |
| a-check mit `.a-check.yml` | die Richtungen der Schichten gelten für die V2-Pakete wie für alle Driven Adapter | `make a-check` |

## Re-Evaluierungs-Trigger

Wenn ein V2-Konstrukt im Umfang ein neues Element des Domain-Modells verlangt (Annahme 1), wenn eine dritte Version hinzukommt oder ein Service mehr als eine Version zugleich bedienen soll (Annahme 2), oder wenn V2 aufhört, ein Kann-Umfang zu sein.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-01 | Proposed | Entwurf des Architect zum Start der Welle für OData V2; die Annahme steht beim Nutzer aus |
