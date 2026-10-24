# Slice slice-odata-codec-json: OData-Codec, Teil Entities und Collections (JSON)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-18`, `LH-RB-01`, `LH-FA-11`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-18.a` Nr. 2 bis 4 · `SPEC-049` (EDM-Typen) · `SPEC-044` · `SPEC-051` (`ODR-E6101`) · `LH-FA-07.b` (Key-Schreibweise) · `SPEC-125` Nr. 11 (`model.ServiceModel`) · `SPEC-107` Nr. 15 (Aufrufer im Import) · `LH-FA-11` (Vollständigkeitsnachweis, Quelle `@odata.nextLink`/`@odata.count`) · `ARC-004` (`ODataCodec`) · `ARC-011` · `architecture.md` §2 (Zeile `ODataCodec`, Baseline-Import).
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

**Ziel:** Der Codec liest OData-V4-JSON-Antworten auf Einzel-Entity- und Collection-Abfragen gegen das Modell in typisierte Entities mit kanonischem Schlüssel, ausgewählten Properties und Nachweis, ob `@odata.nextLink` die Kette fortsetzt, und unterscheidet `null` von nicht selektiert.

**Übernimmt (Port):** Der Port `ODataCodec` entsteht in diesem Slice als eigene Schnittstelle neben `MetadataCodec` und ändert `MetadataCodec` nicht (Entscheidung des Architect von `slice-odata-codec-csdl`: je Darstellung eine Schnittstelle, `SPEC-125`; `AGENTS.md` §3.13). Sein Modell ist `model.ServiceModel` aus `SPEC-125` Nr. 11 (wirksame Schlüssel und Properties, qualifizierte Namen nach Auflösung der Aliase, Zugriffe als Kopien, unbekannter Name `false`); die Erweiterung des Modells um Wertetypen dieses Slice ist eine Änderung der Spezifikation, kein Implementer-Entscheid.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Serialisierung zurück ins OData-JSON und Query-AST — Mock-Modus, nicht in der Welle.
- Zusammenführen, Deduplizieren und Konflikte — `slice-baseline-import` entscheidet; der Codec liefert pro Antwort die beobachteten Fakten.
- Die Redaktionsmaske (`unknown` statt `null`) — Entscheidung des Imports, der die Redaction-Zustände kennt; der Codec liefert, welche Properties im Body standen.

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

- [ ] Port `ODataCodec` (Entities), eigene Schnittstelle neben `MetadataCodec` (Sendung von `slice-odata-codec-csdl`): Antwort, Modell und Entity-Set rein, typisierte Entities raus; EDM-Typen `Edm.String`, `Boolean`, `Int32`, `Int64`, `Decimal`, `Guid`, `Date`, `DateTimeOffset`.
- [ ] Kanonischer, typisierter Schlüssel (einfach und zusammengesetzt) mit Normalisierung der Literale; ungültiger Schlüssel, Typfehler und widersprüchlicher Inhalt als `ODR-E6101`.
- [ ] `@odata.context`-Prüfung gegen das Modell, `@odata.nextLink`, `@odata.count`; `null` gegen fehlend.
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
| `src/hexagon/model/` | update | typisierte Entity, kanonischer Schlüssel |
| `src/hexagon/ports/outbound/` | update | `ODataCodec` (Entities) |
| `src/adapters/driven/odata/jsoncodec/` | neu | JSON-Leser |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, kanonische Form der Schlüssel und Werte |
| `spec/architecture.md` | update, falls der Vertrag sie ändert | Zeile `ODataCodec` in §2, `ARC-004`, `ARC-011`, Sequenz Baseline-Import (Berichtigung an jeder Stelle, `AGENTS.md` §3.9; ohne Slices und ADR-Bezüge, `AGENTS.md` §3.4) |
| `src/hexagon/ports/outbound/outboundtest/` | update | Vertragssuite des Ports `ODataCodec`, die jeder Adapter fährt (Muster `MetadataCodecVertrag`) |
| `.a-check.yml` | update, falls der Vertrag eine Import-Regel für den Adapter verlangt | Muster: Regel für `encoding/xml` im Kern und ihr Fall in den Gegenproben |
| Tests | neu | Fixtures gegen das CSDL des Referenzdienstes, aus Bytewerten gebaut, wo ein Zeichen oder eine Kodierung die Eingabe ist; Negativfälle nach den Mengen der Tabelle *Zusage · Fänger* (§3 Auftrag 3); Vertragstest am Port |

Der Vertrag, die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen in der Spezifikation (Produkt, kein Werkzeug) vor dem ersten Code-Commit (`AGENTS.md` §3.12), nicht vor dem Start; die Randformen, die sie tragen müssen, nennt §6.

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Vertrag, Modell, Entscheidungen aus Auftrag 1, Tabelle, Raster, Paare, Grenzen, Nachzüge aus Auftrag 6, Sendungen an Nehmer, Nennung der erwarteten Produktcode-Größe, Auftrag 7) → Planner (Marke aus der Tabelle und der Produktcode-Erwartung, §4, eigener Commit; Prüfung der Adressen nach `AGENTS.md` §3.13; Schnittfrage, wenn die Schätzung × 1,25 über 3000 liegt) → Implementer (erster Code-Commit erst nach der Marke) → Reviewer → Fixes → Verifier → Planner (Closure). Der letzte Commit des Architect, der Vertrag, Tabelle und Raster abschließt, ist die `<Basis>` der Marke (§4). Hält der Implementer an einer Zelle ohne Ort an (`AGENTS.md` §3.12), geht sie an den Architect, nicht in den Code.

**Betriebsregel Speicher, für alle Rollen** (am 2026-10-24 gab es auf dem 31-GB-Host mehrfach OOM durch Mutationen ohne Grenze): jeder eigene `docker run` mit `--memory=4g --memory-swap=4g`; höchstens drei Container gleichzeitig; eine Mutation ohne Grenze oder Tiefe läuft mit `--memory=2g` und `timeout`; auf Repo-Dateien kein `sed`, `awk` und kein Python, nur `Edit`/`Write`. Docker sonst nur über `make` (`AGENTS.md` §3.1). Ein Leser, dessen Grenze eine Mutation entfernt, muss früh mit einem Fehler enden und den Host nicht füllen (Auftrag 4).

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; Pfade `spec/spezifikation.md`, `spec/architecture.md`; Kennungen fortlaufend je Datei, `SPEC-<NNN>` neu oder als Ergänzung einer bestehenden; Muster ist `SPEC-125`):

1. **Vertrag des Ports `ODataCodec` (Entities und Collections).** (a) Zuschnitt: `SPEC-125` nennt `ODataCodec` als Schnittstelle für „die Entities im JSON und die URL“; dieser Slice liefert nur die Entities. Entscheide, was `ODataCodec` hier trägt und was die URL-Methoden später sind (dem Mock-Modus, nicht der Welle), und halte `MetadataCodec` unverändert. (b) Signatur: Eingabe (Strom oder Bytes, Größengrenze als Parameter wie `grenze` bei `ParseMetadata`, Verhalten bei Lesefehler nach dem ersten Byte), `ctx` (nil und beendet; an welchen Stellen geprüft: vor dem Lesen, zwischen Elementen, innerhalb einer Collection), Modell und Entity-Set als Argumente (unbekanntes Set, Set ohne Typ), Rückgabe (typisierte Entities, kanonischer Schlüssel, die Menge der im Body vorhandenen Properties, `nextLink` und `count` als Nachweise, `null` gegen nicht selektiert; Teilergebnis im Fehlerfall oder nur Fehler), Fehlerform (`ODR-E6101` mit Ursache und Ort wie `MetadatenFehler`; Kontextfehler ohne Code wie `SPEC-124` Nr. 2; welche Fehler `ODR-E6101` sind und welche nicht, etwa Eingabe, die gar kein JSON ist). Kein Typ von `encoding/json` in Signaturen. (c) Wertetypen im Modell (Erweiterung von `SPEC-125` Nr. 11 / `SPEC-002`, §1 *Übernimmt (Port)*): die typisierte Entity und der Wert je EDM-Typ der acht (`SPEC-125` Nr. 10), die unveränderlich nach der Rückgabe ist; Verhalten für Properties mit nicht unterstütztem Typ (`Unterstuetzt()` falsch): gelesen als Rohwert, übersprungen oder gemeldet. (d) **Kanonische Form** je Typ (Risiko 1 in §6, `LH-FA-18.a` Nr. 4: bestimmt, ob zwei Payloads identisch sind): `Decimal` (`1.0` gegen `1`, Exponent, Vorzeichen, `-0`, Präzision ohne Grenze oder mit), `Int64` (Zahl gegen Zeichenkette, `IEEE754Compatible`, Werte über 2^53), `Int32` (Grenzen, `1.0`, `1e2`), `Guid` (Groß-/Kleinschreibung, Klammern, ohne Bindestriche), `Date`, `DateTimeOffset` (Zone, `Z` gegen `+00:00`, Sekundenbruchteile, Schaltsekunde, Jahre außerhalb 0001 bis 9999), `String` (Normalisierung gilt nicht oder gilt: Unicode-Form, Steuerzeichen, ungültiges UTF-8, ein halbes Surrogatpaar als `\ud800`, `\u0000`), `Boolean`. (e) **Kanonischer Schlüssel**: einfach und zusammengesetzt, Reihenfolge nach dem Modell (nicht nach dem Body), Literalform (`SPEC-116`-Bezug nur, wo der Match-Key den Schlüssel bildet; hier ist es der Entity-Schlüssel nach `LH-FA-07.b`), Anführungszeichen in String-Schlüsseln, Sonderzeichen, Schlüssel-Property fehlt, `null`, doppelt, vom falschen JSON-Typ, nicht im Select. (f) **Struktur**: `value` fehlt, ist kein Array, leer; Einzel-Entity gegen Collection (der Aufruf nennt, was erwartet wird, oder der Body entscheidet: entscheide); Elemente, die kein Objekt sind; doppelte Namen im selben Objekt (`encoding/json` nimmt den letzten still); unbekannte Property, `@odata.type` für abgeleitete Typen (Typ des Modells, Properties des abgeleiteten Typs), Instanz-Annotationen am Objekt und an einer Property (`Prop@odata.type`, `@odata.etag`, `@odata.id`, `@odata.editLink`), Navigation-Properties und `expand` (eingebettete Entities und Collections: gelesen, übersprungen oder `ODR-E6101`), komplexe und Collection-Properties (Wert eines nicht unterstützten Typs). (g) **Nachweise**: `@odata.context` gegen das Modell (Formen `$metadata#Set`, `$metadata#Set/$entity`, `$metadata#Set(Prop,…)`, `$metadata#Set(Prop)/$entity`, absolut und relativ, mit Alias oder qualifiziertem Namen, fehlender Kontext bei `odata.metadata=none`, Kontext eines anderen Sets, eines anderen Dienstes), `@odata.nextLink` (absolut, relativ, leer, kein String; der Codec liefert die Zeichenkette, der Import löst auf, `SPEC-116` und `slice-baseline-import` §1), `@odata.count` (Zahl, Zeichenkette, negativ, größer als `int64`, bei Einzel-Entity), die Unterscheidung `null` gegen nicht selektiert gegen fehlend und was `$select` im Kontext für die Menge der erwarteten Properties bedeutet. (h) Abgrenzung: was dieser Vertrag nicht festlegt (Schreiben, Query-AST, Zusammenführen, Maske; §1).
2. **Raster als §6-Zellen** *Operation × Argumentart* (`AGENTS.md` §3.16), jede Zelle mit Ort in der Spezifikation oder „gilt nicht“ mit Grund, eine leere Zelle ist ein Fehler. Operationen mindestens: Lesen einer Entity, Lesen einer Collection, Bilden des kanonischen Schlüssels, Prüfen des Kontexts, Lesen eines Werts je EDM-Typ. Argumentarten mindestens: `ctx` nil und beendet (vor, während); Eingabe `nil`, leer, nur Leerraum, nur BOM, kein JSON, abgeschnitten mitten im Wert, Lesefehler des Stroms nach dem ersten Byte, Rest nach dem Wurzelwert (zweiter Wert, Leerraum, Müll); Kodierung (UTF-8 mit und ohne BOM, UTF-16, ungültige Bytefolge, `\u`-Escapes, Surrogate, `\u0000`); Wurzel kein Objekt (Array, Zahl, `null`, Zeichenkette); Modell `nil`/leer/Set unbekannt/Set mit Typ ohne Schlüssel (kann nicht vorkommen: `SPEC-125`); Wiederholung (gleiche Eingabe, gleiche Ausgabe, Eingabe unverändert) und gleichzeitige Aufrufe (Auftrag 6). Die Zellen führen den Fall „zwei verglichene Werte gleich“ (`1` gegen `1.0`, gleicher Schlüssel in zwei Elementen derselben Collection: Duplikat, gemeldet oder beobachtet; der Codec liefert pro Antwort die beobachteten Fakten, `slice-baseline-import` entscheidet über Konflikte, §1).
3. **Menge vor dem Fänger, Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15, §3.17): die Mengen werden aus dem Wortlaut der Zusage im Vertrag gebildet, nicht aus dem Fall, der dem Autor einfiel; zu jeder Zeile steht die Menge, die Zahl ihrer Elemente und der Fänger als Schleife über alle (die Fixture gleicht nie dem Wert der Mutation). Mindestens: (i) **8 EDM-Typen × Wertformen** (richtiger Wert in jeder zulässigen Schreibweise, falscher JSON-Typ aus {Zeichenkette, Zahl, `true`/`false`, `null`, Objekt, Array}, Grenzwerte und Grenze + 1, `null` bei `nullable` wahr und falsch) — eine Schleife über Typ × Wertform; (ii) **Schlüssel**: einfach, zusammengesetzt (jede Reihenfolge der Properties im Body gegen die des Modells), je Schlüsseltyp der acht, jede Verletzung (fehlt, `null`, doppelt, falscher Typ, außerhalb des Bereichs) an jeder Position des Schlüssels; (iii) **Kontextformen** (g) einzeln, jede als Fall; (iv) **Elementarten × Positionen** — Lehre aus N-4 von `slice-odata-codec-csdl`: die Zusatztabelle der Paare von Verstößen nennt die Dimension *Art und Position des Elements*, an dem der Verstoß liegt (Wurzel, Collection-Element erstes/mittleres/letztes, Property im Element, Annotation, eingebettete Entity), und die Rangfolge, welcher Verstoß gemeldet wird, wenn zwei am selben Element oder an verschiedenen Elementen liegen; je Paar der Fänger, die Zahl der Paare wird aus der Menge der Verstöße × Positionen gezählt (nicht aus einer Auswahl); die Allaussage „kein Fall außer … ist `ODR-E6101`“ nennt ihren Fänger über alle Fälle; (v) **Annotationsarten**, **abgeleitete Typen**, **unbekannte Properties** und **`expand`** je Element der Menge; (vi) **Stabilität**: Reihenfolge der Entities und Properties ist die des Bodys (nicht die einer Map), Fänger mit Rate je Prozess und Zahl oder durch Konstruktion mit `-count=20`; (vii) **Fixtures** gegen das CSDL des Referenzdienstes (Entity-Sets mit einfachem und zusammengesetztem Schlüssel, Navigation) und echte Antwortformen; Randformen aus Bytewerten gebaut (`BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um`), nach dem Schreiben gegengelesen; (viii) **Port-Vertragstest** im Stil von `outboundtest/` (Suite, die jeder Adapter fährt) und, falls der Slice einen Fake von `ODataCodec` liefert (Zählung jeder Operation, jedes Argument festgehalten), seine Zeile und 200 in der Marke; liefert ihn `slice-baseline-import`, nennt der Vertrag den Träger. **Aufrufstellen des Ports als Menge**: hier nur der Vertragstest; der Aufrufer ist der Import (Sendung, Auftrag 5), jede Aufrufstelle des Ports dort bekommt dort ihren Fall (`AGENTS.md` §3.15: `ctx`, Fehlerabbildung, Kontextfehler je Stelle, kein Aufruf, den der Dienst nie rufen darf). Zu jeder Zeile die Mutation, die sie rot macht, und der Fall, in dem zwei verglichene Werte gleich sind. Für Reviewer und Verifier nennt der Architect je Klasse der Spezifikation, ob sie **generierbar** ist (Wertformen je Typ, Schlüsselpositionen, Kontextformen, Paare), damit das Probe-Paket aus Generatoren (`BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`, 1×) auf den Klassen des Vertrags arbeiten kann; ob die Suite selbst einen Generator führt, entscheidet der Architect mit Grund und Folge.
4. **Prüfgrenzen und akzeptierte Negative** (Lerneintrag von `slice-session-beenden`: eine Klasse mit unbeschränktem Umfang hat kein letztes Mitglied): für Größe des Bodys (Beziehung zu `MAX_CAPTURE_BODY`, `SPEC-013`; die Grenze als Parameter wie `grenze` bei `ParseMetadata`), Tiefe der Verschachtelung (`expand`, komplexe Werte; `encoding/json` hat eine eigene Tiefengrenze, der Vertrag nennt seine), Zahl der Elemente in `value`, Zahl der Properties je Objekt, Länge einer Zeichenkette, Länge einer Zahl (`Decimal` mit Millionen Ziffern), Länge der Annotationsnamen, Zahl der Annotationen: je Klasse die Grenze, die der Leser durchsetzt, oder „keine, weil …“ mit Grund und Folge im Betrieb (Speicher, Laufzeit), den Fehlercode bei Überschreitung (`ODR-E6101` mit Ursache oder eigene Entscheidung), den Fänger **Grenze und Grenze + 1** mit einer Fixture, die ungleich der Mutation ist. **Speicherwächter:** der Leser muss bei einer Mutation ohne Grenze früh mit Fehler enden, nicht den Host füllen (Muster `speicherLimit` in `csdl`, Betriebsregel Speicher oben); die Grenzen werden so gesetzt, dass der Fänger der Mutation in einem Container mit `--memory=2g` mit dem benannten Grund rot wird. Jedes akzeptierte Negativ nennt Grund und Folge (`AGENTS.md` §3.15); jede Fänger-Menge, die nur abgetastet wird, sagt es.
5. **Sendungen nach `AGENTS.md` §3.13.** Der Aufrufer des Ports ist `slice-baseline-import` (Ziel dort: „interpretiert sie über den Codec“; §1 dort nennt bisher nur `MetadataCodec`, nicht `ODataCodec`). Der Planner hat beim Start gelesen: *Ausdrücklich NICHT* dort trifft die Sendung nicht (Endpunkt, Lesen im Mock, mehrere Sessions), die DoD Punkt 1 nimmt Zusammenführen, Dedup, Konflikte (`ODR-E6101`) und die Maske `unknown` an (die drei Ausschlüsse in §1 oben haben damit einen Nehmer); `slice-baseline-import` liegt in `open/`. Zu entscheiden und im selben Commit wie die Entscheidung beim Nehmer in §1 und DoD einzutragen, mit der Kennung dieses Slice: der Aufruf von `ODataCodec` durch den Import (welche Methode je Antwortart, welches Modell, `ctx` und Größengrenze, Fehlerabbildung auf `ODR-E6101` und `ODR-E6004`), der Fake von `ODataCodec` (200 in der Marke des Nehmers), die Auflösung von `@odata.nextLink` (Import, mit `SPEC-116`; der Codec liefert die Zeichenkette), die Maske `unknown` (Import; der Codec liefert, welche Properties im Body standen). Lies vorher dessen §1 *Ausdrücklich NICHT*, DoD und §4 (Marke, vorläufig 2100). Trifft ein Ausschluss die Sendung, geht der Punkt an den Planner.
6. **Nachzug an jeder Stelle, die dieselbe Aussage trägt** (`AGENTS.md` §3.9): `LH-FA-18.a` Nr. 2 bis 4, `SPEC-002`, `SPEC-044`, `SPEC-049` (EDM-Zeile, Teilumfang), `SPEC-051` (`ODR-E6101`, Ursache und Ort im Fehler), `SPEC-068`, `SPEC-107` Nr. 15, `SPEC-125` Nr. 11 (Modell), `SPEC-125` Satz zu `ODataCodec`, die Zeile `ODataCodec` in `architecture.md` §2, `ARC-004`, `ARC-011`, die Sequenz Baseline-Import, `SPEC-006`/Fehlermodell, wenn die Ursache des `ODR-E6101` ein Feld trägt; `src/hexagon/model/fehler.go` ändert der Architect nicht (Code). Ohne Slices, Wellen und ADR-Bezüge in `architecture.md` (`AGENTS.md` §3.4). Ein Nachzug, der einen Satz berichtigt, sucht alle Stellen, die ihn tragen.
7. **Nebenläufigkeit, Bibliothek, Größe.** (a) Ist der Leser ein Wert ohne Zustand, steht die Zusage „gleichzeitige Aufrufe sind erlaubt und beeinflussen sich nicht“ mit Fänger unter `make test-race` (je Prozess, `-test.count=1`) und Wirkungs-Fänger, oder als akzeptiertes Negativ mit Grund und Folge; die Entscheidung bestimmt den Faktor 85 gegen 40 der Marke. Zweige, die der Detektor-Test ausführt, aus dem Abdeckungsprofil (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`). (b) Bibliothek: `encoding/json` genügt oder eine ADR vor dem Code; gilt `encoding/json`, stehen die Zusagen zu doppelten Namen, ungültigem UTF-8 (die Bibliothek ersetzt es still durch U+FFFD), Zahlen (`UseNumber`, Präzision) und Tiefe im Vertrag, nicht im Verhalten der Bibliothek; ob der Leser den Strom mit `Decoder.Token` liest oder den Body in Werte lädt, ist Folge der Speichergrenze (Auftrag 4). (c) **Größe:** die Formel von `AGENTS.md` §3.14 zählt Fänger, nicht den Produktcode (§4 *Hinweis*: csdl Produktcode 1506 neben 2827 Test; Schätzung 1850, Ende 4333). Der Architect nennt zum Vertrag die **erwartete Produktcode-Größe** in Zeilen (Leser, Modell, Port, Abbildung) mit Begründung und, wo es trägt, den Vergleich mit den 1506 Zeilen des CSDL-Lesers, und er benennt einen **Teil, der keinen Mechanismus mit dem Rest teilt** und beim Handoff geschnitten werden dürfte (etwa die Wertetypen und ihre kanonische Form als reine Funktionen gegen das Lesen des Bodys; entscheide, ob es einen gibt), damit §4 ihn vor dem Code benennen kann (`AGENTS.md` §3.14). Er zählt die Zeilen der Tabelle nach Zählregel von `slice-odata-codec-csdl` §4 (nebenläufig, wenn der Code eine Goroutine startet oder der Fänger eine Goroutine, eine Barriere oder einen Halt trägt).

**Der Planner hat nichts davon entschieden.** Die Aufträge benennen die Entscheidungen und die Mengen, die der Architect vor dem Code schließen muss; jede offene Entscheidung steht in §6 mit dem Ausgang *offen*.

**Aufträge an den Implementer** (nach der Marke): (a) der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9; Praxis aus N-8 von `slice-odata-codec-csdl`, in drei Runden war jede Arbeit auf drei Commits verteilt); Gegenprobe und Stempel dürfen folgen. (b) Der Absatz in §3 mit Zusage · Mutation · rotem Test und den Raten (`AGENTS.md` §3.10). (c) Jeder Fix zu einem Befund bildet die Menge aus dem Wortlaut der Zusage, schreibt den Fänger als Schleife und nennt im Bericht je Befund Menge, Zahl der Elemente und eine Mutation an einem Element, das der Befund nicht nannte (`AGENTS.md` §3.17). (d) Messen vor der Übergabe mit `git diff --numstat <Basis>..HEAD` über die Pfade der Marke, Zahl im Bericht. (e) Betriebsregel Speicher oben.

**Aufträge an Reviewer und Verifier** (jeder in frischem Kontext): die Zeilen der Tabelle und das Raster mit je einer Mutation auflösen, Fixes gegen die **Klasse** (`AGENTS.md` §3.17), nicht gegen die genannte Mutation; der Verifier fährt zusätzlich ein **Probe-Paket aus Generatoren** über die Klassen der Spezifikation (`BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`), einmal gegen den Code, dann gegen die Mutationen, und legt es im Report ab (Pfad und Ergebnis). Betriebsregel Speicher oben, besonders für Mutationen, die eine Grenze entfernen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-odata-codec-csdl` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand des Starts (2026-10-24):** Der Start-Trigger ist erfüllt: `slice-odata-codec-csdl` liegt in `done/` (Closure mit Lerneintrag, `AGENTS.md` §3.17), `ODR-E6101` ist als Code im Fehlermodell (`SPEC-051`). Der Vertrag, die Tabelle und das Raster stehen noch aus (Architect, §3); die Marke ist **noch nicht gesetzt**.

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Der Planner setzt sie nach dem Vertrag des Architect und vor dem ersten Code-Commit auf die Schätzung × 1,25, aufgerundet auf volle Hundert, und ersetzt die vorläufige Erwartung unten dabei mit Grund. Die Schätzung nach der Regel: Zeilen der Tabelle *Zusage · Fänger* (× 85 nebenläufig, × 40 jede andere) plus 200 je Port-Fake oder Test-Server; **zusätzlich** nennt der Architect die erwartete Produktcode-Größe (§3 Auftrag 7 (c)), und der Planner führt beide Zahlen getrennt in §4 auf und rechnet die Produktcode-Erwartung als festen Anteil in die Schätzung ein (`AGENTS.md` §3.14: „feste Anteile eingeschlossen“); der Faktor 1,32 bis 1,39 aus dem Messpunkt von `slice-odata-codec-csdl` bleibt ein Hinweis, kein Zuschlag (Entscheidung des Users). Liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten; lässt sich kein Teil ohne geteilten Mechanismus abtrennen, steht der Grund hier. Ein beim Handoff erlaubter Schnitt ist nur der Teil, den der Architect benennt und §4 vor dem Code festhält. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der letzte Commit des Architect, der Vertrag, Tabelle und Raster abschließt (Dokumente zählen nicht mit). Dieser Slice ist der zweite Leser fremder Bytes und liefert den zweiten Messpunkt (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`, 14×).

*Hinweis, keine Änderung der Regel:* Der erste Leser fremder Bytes, `slice-odata-codec-csdl`, hat die Formel von `AGENTS.md` §3.14 weit überschritten: Schätzung 1850, Handoff 3628, Ende 4333 Zeilen (Marke 2400); Produktcode 1506, Tests und Fixtures 2827 (64 je Zeile der Tabelle statt 40). Die Formel zählt Fänger, nicht den Code, den sie fangen. Der Planner setzt die Marke beim Start mit diesem Messpunkt vor Augen, die Closure dieses Slice liefert den zweiten gemessenen Wert; über eine Regel entscheidet der User danach (Entscheidung des Users: kein Zuschlag mit einem Messpunkt). Praxis für den Auftrag an den Implementer (Regel `AGENTS.md` §3.9 unverändert): der Handoff ist ein Commit für Code und Plan; Gegenprobe und Stempel dürfen folgen.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Größenmarke oben (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Erwartung, vorläufig, vor der Tabelle des Architect:** 1400 Zeilen aus der älteren Fassung dieses Plans (Schätzung rund 1000); sie liegt unter dem Messpunkt von `slice-odata-codec-csdl` (4333) und gilt nur, bis die Marke gesetzt ist; der Planner ersetzt sie dann mit Grund (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): die kanonische Form eines Werts (`Decimal`, `Date`, `DateTimeOffset`) ist ohne Entscheidung des Architect nicht festlegbar (siehe §6).

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

- Kanonische Form von `Decimal` (`1.0` gegen `1`), `DateTimeOffset` (Zeitzone, Sekundenbruchteile), `Guid` (Groß-/Kleinschreibung), `Int64` als String oder Zahl (`IEEE754Compatible`) — **Ausgang:** offen; sie bestimmt, ob zwei Payloads „identisch“ sind (`LH-FA-18.a` Nr. 4); Architect entscheidet vor dem Code.
- Schlüssel mit Sonderzeichen, mit Anführungszeichen in String-Schlüsseln, Composite-Schlüssel in beliebiger Reihenfolge — **Ausgang:** offen; Raster *Schlüsselform × Behandlung*.
- Antworten außerhalb des Musters: `value` fehlt, leere Collection, Entity mit unbekannter Property, Annotationen am Wert, `@odata.type` für abgeleitete Typen — **Ausgang:** offen; Raster-Zelle je Form.
- Grenzen: große Collections und tiefe Verschachtelung — **Ausgang:** offen; Architect bestimmt Grenzen oder „gilt nicht“ mit Grund.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- Bibliothek `encoding/json`: ersetzt ungültiges UTF-8 still, nimmt bei doppelten Namen den letzten, rundet Zahlen ohne `UseNumber` über `float64` — **Ausgang:** offen; Architect entscheidet vor dem Code (§3 Auftrag 7 (b)), Zusagen im Vertrag, jede mit Fänger.
- Paare von Verstößen nach Art und Position des Elements (Lehre N-4 von `slice-odata-codec-csdl`, §3 Auftrag 3 (iv)) — **Ausgang:** offen; Zusatztabelle mit der Dimension und Rangfolge vor dem Code.
- Speicher: ein Leser ohne Grenze füllt den Host (am 2026-10-24 mehrfach OOM durch Mutationen ohne Grenze) — **Ausgang:** offen; Grenzen und Speicherwächter nach §3 Auftrag 4, Betriebsregel Speicher in §3.
- Sendung an `slice-baseline-import` (Aufrufer von `ODataCodec`, Fake, Auflösung von `nextLink`, Maske `unknown`; `AGENTS.md` §3.13) — **Ausgang:** offen; Architect trägt sie beim Nehmer ein (§3 Auftrag 5), der Planner prüft vor der Marke.
- Größe (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`, 14×, verkörpert in `AGENTS.md` §3.14): vorläufige Erwartung 1400 liegt weit unter dem Messpunkt 4333 — **Ausgang:** offen; der Planner setzt die Marke nach der Tabelle und der Produktcode-Erwartung des Architect (§4).
- Bericht des Implementers (`BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`, 3×, verkörpert in `AGENTS.md` §3.10): Tabelle *Zusage · Mutation · roter Test* und Raten in §3 — **Ausgang:** offen; fehlt der Absatz im Review, geht der Beleg an den User.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Register unter `../observations/BEO-REPO/` trägt beim Start fünfzehn Einträge (Sub-Area `*`; Stand 2026-10-24 nach der Closure von `slice-odata-codec-csdl`); Zahlen aus den Dateien unter `evidence/`, Ausgänge aus den `state.md`. Der Slice liefert einen neuen Vertrag über ein fremdes Format, den zweiten Leser fremder Bytes:
- `randform-erst-im-review-entschieden` (18×, verkörpert in `AGENTS.md` §3.16) und `universalaussage-im-vertrag-ungeprueft` (7×, §3.16): Raster in §6 vor dem ersten Code-Commit, Allaussagen („kein Fall außer … ist `ODR-E6101`“, „jeder Typ“) nennen ihren Fänger über alle Fälle (§3 Auftrag 2, 3).
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (23×, §3.15) und `akzeptiertes-negativ-ohne-folge` (8×, §3.15): Tabelle mit Menge je Zeile, Aufrufstellen des Ports als Menge, Negative mit Grund und Folge (§3 Auftrag 3, 4).
- `fix-deckt-die-genannte-mutation-nicht-die-klasse` (3×, verkörpert in `AGENTS.md` §3.17): Mengen aus dem Wortlaut der Zusage, Fänger als Schleife; Auftrag an Implementer, Reviewer, Verifier (§3).
- `suite-ohne-generator-fuer-klassen-der-spezifikation` (1×, offen): Probe-Paket aus Generatoren über die Klassen der Spezifikation im Auftrag an Reviewer und Verifier; der Architect nennt die generierbaren Klassen (§3 Auftrag 3).
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (14×, §3.14): Marke nach Tabelle und Produktcode-Erwartung, vor dem ersten Code-Commit (§4).
- `berichtigung-erreicht-nicht-jede-stelle` (9×, §3.9): Stellen in §3 Auftrag 6, Sendung an `slice-baseline-import` (Auftrag 5, `AGENTS.md` §3.13).
- `nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, §3.10, Sensor `make test-race`), `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen) und `zweig-ohne-lauf-unter-dem-detektor` (1×, offen): Bezug über die Zusage zu gleichzeitigen Aufrufen und die Reihenfolge des Bodys (§3 Auftrag 3 (vi), 7 (a)); eine Rate nennt Zahl der Prozesse und Messenden.
- `bericht-der-rolle-liegt-nicht-im-repo` (3×, §3.10): Risiko in §6.
- `befund-an-rolle-ohne-artefakt` (2×, offen): jede offene Entscheidung steht in §6 mit Ausgang *offen* und der Adresse *Architect*.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×, offen): Fixtures mit `\u`-Escapes, BOM, Surrogaten und Kodierungen aus Bytewerten, nach dem Schreiben gegengelesen (§3 Auftrag 3 (vii)).
- `panik-eines-fakes-verdeckt-den-benannten-test` (2×, offen): betrifft Fakes in Goroutinen; trifft zu, falls der Slice einen Fake von `ODataCodec` liefert und der Fänger nebenläufig ist (§3 Auftrag 3 (viii)).

**Stand des Starts (2026-10-24):** Das Register trägt die Zahlen oben. Unter der Schwelle (kein Ausgang zu vergeben): `befund-an-rolle-ohne-artefakt`, `panik-eines-fakes-verdeckt-den-benannten-test`, `rate-ohne-detektor-schwankt-mit-der-last`, `werkzeug-eingabe-setzt-escape-in-zeichen-um` (je 2×), `suite-ohne-generator-fuer-klassen-der-spezifikation` und `zweig-ohne-lauf-unter-dem-detektor` (je 1×); ein dritter Beleg löst beim Lese-Schritt der Closure die Verkörperung aus.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
