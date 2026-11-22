# Slice slice-expand-validator: Validator von `$expand` gegen die Navigationen des Modells

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-16.a` (eigener Validator-Contract-Test, unbekannte Operationen `501`) · `SPEC-002`, `SPEC-125` Nr. 9 d und Nr. 11 (Navigation-Properties des Modells, wirksam einschließlich geerbter; `Target` aufgelöst) · `SPEC-160` Nr. 9 (was der Dienst prüft, nicht der Parser) · `SPEC-263` (Vertrag von `PruefeExpand`) · `SPEC-167` (Muster: reine Funktion in `model`, eigener Fehlertyp) · `SPEC-022` (`ODR-E6001`), `SPEC-163` (`ODR-E6005`) · `SPEC-049` (Zeile GET: Navigation-Expand) · `ARC-004`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-expand-validator`, Worktree `/Development/odata-recorder-wt-expand-validator`; der Zweig entsteht von `main` nach dem Start-Commit)

**Autor:** pt9912. **Datum:** 2026-11-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Eine reine Funktion im Paket `src/hexagon/model` (Muster `PruefeFilter`, `SPEC-167`) prüft jede Navigation des Knotens `$expand` (geliefert von `slice-expand-parser`) gegen die Navigation-Properties des Entity-Typs im Modell des Dienstes, die geerbten eingeschlossen: eine vorhandene Navigation ist gültig, ein unbekannter Name oder eine Property ohne Navigation ist `400` (`ODR-E6005`), eine gültige, aber nicht unterstützte Form `501` (`ODR-E6001`), nach `SPEC-263` (§6 V1 bis V7). Die Funktion liest nur den Knoten und das Modell, sie greift auf keine Daten zu und kennt weder Kontext noch Port (`LH-FA-16.a`: Validator-Contract-Test).

**Übernimmt (Schnitt von `slice-expand-parser`, Geber `slice-expand-parser`, `AGENTS.md` §3.13):** den Validator gegen das Modell samt Fehlertyp, Raster und Validator-Contract-Test, die der frühere Plan von `slice-expand-parser` führte (Navigation vorhanden, auch geerbt; Property ist eine Navigation; Zieltyp aufgelöst; unbekannt oder keine Navigation `400`, gültig, aber nicht unterstützt `501`; kein Zugriff auf Daten).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Parser, Grammatik, Knoten und Einhängestelle von `$expand` — `slice-expand-parser`; dieser Slice liest den Knoten und ändert ihn nicht (`SPEC-160` Nr. 3 gilt für den Parser).
- Der Aufruf des Validators im Dienst und die Verdrahtung — `slice-expand-ausfuehrung` (dessen §1 trägt die Sendung); dieser Slice liefert die Funktion, nicht ihren Aufrufer.
- Auflösen gegen Daten, Datenbasis der Navigationen, Ausgabe — `slice-expand-ausfuehrung`.
- Navigation ohne `NavigationBinding` am Set und fehlende Daten der Navigation — `slice-expand-ausfuehrung` (dessen §1 trägt die Sendung, `AGENTS.md` §3.13); die Funktion prüft den Typ, nicht die Bindings (`SPEC-263` Nr. 5).
- Änderungen am Metadatenleser und am Modell (`SPEC-125`): `Path` der `NavigationPropertyBinding` wird nicht aufgelöst (`SPEC-125` *Ungeprüft*), und die Funktion liest, was das Modell liefert; eine Änderung des Modells ist ein Vertrag in `done/` und keine Adresse für eine Sendung (`AGENTS.md` §3.13).
- Herstellerspezifische Navigations-Erweiterungen — Out-of-Scope von `LH-FA-16`.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste.

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [x] Funktion und Fehlertyp: jede Navigation des Knotens gegen die wirksamen Navigation-Properties des Typs (auch geerbt); Property ohne Navigation, unbekannter Name und die nicht unterstützte Form mit Ursache, Klasse, Code und Status nach dem Vertrag des Architect; kein Teil der Eingabe im Fehler; kein Zugriff auf Daten. (Beleg: `src/hexagon/model/expand_typ.go`, `src/hexagon/model/expand_typ_fehler.go`; Zeilen 1 bis 12 der Tabelle in §3, Mutationen dort.)
- [x] Randformen (Programmfehler des Aufrufers, Wiederholung, Nebenläufigkeit, Reihenfolge mehrerer Verstöße, Subtyp, Navigation ohne Binding): je Zelle des Rasters (§6) ein Ort in der Spezifikation und ein Fänger der Tabelle (§3), als Schleife über alle Navigationen und alle Properties des Fixture-Modells. (Beleg: Zeilen 6 bis 9 und 13 bis 16 der Tabelle, Mengen und Raten in §3.)
- [x] Validator-Contract-Test (`LH-FA-16.a`): ein eigener Test fährt die Funktion gegen die Fälle von `SPEC-263` Nr. 2 a bis d, 3 und 4 (die Spezifikation führt keine eigene Beispieltabelle, §6; Zeile 18 der Tabelle); `make a-check` hält die Richtung (`model` importiert keinen Port und keinen Dienst). (Beleg: `TestPruefeExpandContract` in `src/hexagon/model/expand_typ_contract_test.go`; `make a-check` ohne Befund, mit einem Import von `ports/outbound` ein Befund `core-impurity`, §3 Zeile 17.)
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (entschieden: `SPEC-109` bleibt, `ExpandFehler` trägt die Sentinel der URL-Klassen, `SPEC-263` Nr. 3).
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen; das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (die Funktion hat keinen Aufrufer vor `slice-expand-ausfuehrung`).
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
| `src/hexagon/model/expand_typ.go`, `src/hexagon/model/expand_typ_fehler.go` | neu | `PruefeExpand` (102 Zeilen Produktcode mit dem Fehlertyp) und `ExpandFehler` mit `ExpandUrsache` (Muster `filter_typ.go`, `filter_typ_fehler.go`), technologieunabhängig |
| `src/hexagon/ports/outbound/abbildung.go` | unverändert | `ExpandFehler` trägt die Sentinel der URL-Klassen, die `SPEC-109` Nr. 5 schon abbildet (`SPEC-263` Nr. 3); keine Zeile neu |
| `spec/spezifikation.md` | update | Vertrag `SPEC-263` mit Randformen (Architect, geschrieben) |
| `src/hexagon/model/expand_typ_fixture_test.go` (Fixture-Modell, Orakel, Zählung), `expand_typ_test.go` (Zeilen 1 bis 9 und 13), `expand_typ_fehler_test.go` (Zeilen 10 bis 12), `expand_typ_eigenschaft_test.go` (Zeilen 14 und 16), `expand_typ_achsen_test.go` (Typachsen der Zeilen 4 und 5, Namen der Zeile 8, Längen der Zeile 15), `expand_typ_contract_test.go` (Zeile 18) | neu | Fälle je Navigation und je Fehlerklasse; Schleife über alle Navigationen und Properties des Fixture-Modells |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Größenmarke hat der Planner beim Start nach `SPEC-128` aus dieser Tabelle und der Erwartung des Architect gesetzt (§4): `produktcode` rund 120 (die Funktion rund 50 Zeilen, der Fehlertyp rund 70 nach dem Muster von `filter_typ_fehler.go`, 76 Zeilen). Die Tabelle *Zusage · Fänger* hat der Planner aus `SPEC-263` Nr. 1 bis 6 vor dem Start gefüllt (Auftrag des Plans; der Architect hat Vertrag, Entscheidungen V1 bis V7 und Raster geliefert, §6).

**Tabelle *Zusage · Fänger*** (Planner, 2026-11-21, aus `SPEC-263` Nr. 1 bis 6 und dem Raster in §6; Fixture-Modell und Orakel entstehen im Test: eine eigene Tabelle der erwarteten Navigationen und Properties je Typ ohne Code der Funktion, `AGENTS.md` §3.17; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10; Dateien liegen unter `src/hexagon/model/`):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Eine Navigation des Typs, auf eine Entity oder auf eine Collection, auch eine geerbte, ist gültig: Ergebnis `nil` (`SPEC-263` Nr. 2 a, Nr. 1) | `TestPruefeExpandGueltig`: Schleife über alle Typen des Fixture-Modells (Basis, abgeleitet, abstrakt, offen, Kette aus drei Stufen) und alle wirksamen Navigationen je Typ, einzeln und als Liste in aufsteigender und absteigender Folge; Orakel: die erwartete Liste je Typ als Tabelle im Test; Zählprüfung der Menge | Ist-Stand: die Suche lässt die erste, die letzte oder alle bis auf die ersten acht Navigationen aus; sie findet nur Navigationen auf eine Entity oder nur auf eine Collection (`z01a` bis `z01e`; „geerbt nicht gesucht“ ist im Code der Funktion nicht schreibbar, sie sieht nur den wirksamen Typ) | andere |
| 2 | Ein Name wird bytegleich gesucht: Groß- und Kleinschreibung zählt, keine Normalisierung (`SPEC-263` Nr. 1) | `TestPruefeExpandSchreibweise`: Schleife über alle Navigationen aller Typen × Abwandlungen {erster Buchstabe groß, alles klein, alles groß, angehängtes Leerzeichen, Normalform NFD statt NFC bei einem Namen mit Zeichen jenseits von ASCII}; Fixture aus Bytewerten, die Datei auf Rohbytes gegengelesen (`AGENTS.md` §3.18); Erwartung `navigation_unbekannt`; Ist-Stand: auch jede Property, und `é` steht in einer Navigation und in einer Property | Vergleich mit `strings.EqualFold` (Navigationen, Properties); Normalisierung vor dem Vergleich (Navigationen, Properties) | andere |
| 3 | Ein Name in `typ.Properties`, jede Property, auch eine geerbte und eine mit nicht unterstütztem Typ, ist `keine_navigation`, im offenen wie im nicht offenen Typ (`SPEC-263` Nr. 2 b, V6) | `TestPruefeExpandProperty`: Schleife über alle wirksamen Properties aller Typen, offen und nicht offen, darunter Properties mit nicht unterstütztem Typ; Zählprüfung | Eine Property mit nicht unterstütztem Typ zählt als unbekannt; die Prüfung von `Open` geht der der Properties vor | andere |
| 4 | Ein Name, der in keiner der beiden Listen steht, in einem offenen Typ, ist `navigation_nicht_unterstuetzt` (501), nicht `navigation_unbekannt` (`SPEC-263` Nr. 2 c, V6) | `TestPruefeExpandOffen`: Schleife über die Namen {fremder Name, Name einer Navigation eines anderen Typs, Name einer Navigation des Subtyps, Name mit anderer Schreibweise einer vorhandenen Navigation} × alle offenen Typen des Modells; Ist-Stand: dazu `TestPruefeExpandTypachsen` über alle acht Kombinationen aus offen, abstrakt und abgeleitet (offene Typen mit und ohne `Abstract`, mit und ohne Basis) | `Open` wird nicht gelesen (alle `navigation_unbekannt`); ein offener Typ macht jeden Namen gültig; `Open` gilt nicht für abstrakte bzw. abstrakte abgeleitete Typen (`f1-r03`, `f1-r04`) | andere |
| 5 | Ein Name, der in keiner der beiden Listen steht, in einem nicht offenen Typ, ist `navigation_unbekannt` (400): auch der Name einer Navigation oder Property eines anderen Typs, in jedem Typ, auch im abstrakten (`SPEC-263` Nr. 2 d) | `TestPruefeExpandUnbekannt`: Schleife über alle Typen × {fremder Name, jede Navigation und jede Property jedes anderen Typs des Modells, die der Typ nicht trägt} × {abstrakt, nicht abstrakt}; Orakel: Menge der Namen je Typ aus der Tabelle des Tests; Ist-Stand: `Keys` jedes Typs nennt die erste Property und einen Namen, den keine Liste führt (`Schluessel`, ein fremder Name in jeder Menge), und `TestPruefeExpandTypachsen` führt die acht Kombinationen aus offen, abstrakt und abgeleitet | Ist-Stand: ein abstrakter bzw. abgeleiteter Typ lässt fremde Namen zu; unbekannt als 501; ein Name gleich dem `Target` einer Navigation ist gültig; ein Name in `typ.Keys` zählt als Property bzw. als gültig (`z05a` bis `z05d`, `i1-r18`, `i1-n1`; „die Namen aller Typen des Modells werden gesucht“ ist im Code der Funktion nicht schreibbar) | andere |
| 6 | Eine Navigation, die nur ein abgeleiteter Typ trägt, ist am Typ des Sets `navigation_unbekannt` und am abgeleiteten Typ gültig; `typ` ist der Typ des Sets (`SPEC-263` Nr. 2 d, V4) | `TestPruefeExpandSubtyp`: Schleife über alle Paare (Typ, abgeleiteter Typ) des Fixture-Modells, bis zur dritten Stufe, und alle Navigationen, die nur der abgeleitete Typ trägt; Ist-Stand: auch jede Property, die nur der abgeleitete Typ trägt | Ist-Stand: ein Typ ohne Basis bzw. ein abgeleiteter Typ lässt fremde Namen zu (`z06a`, `z06b`; „sucht auch in den Subtypen“ ist im Code der Funktion nicht schreibbar) | andere |
| 7 | Der erste Verstoß von links entscheidet, gleich welcher Klasse; die Prüfung endet dort (`SPEC-263` Nr. 2, V3) | `TestPruefeExpandRang`: alle geordneten Paare (i ungleich j) aus {gültig, `navigation_unbekannt`, `keine_navigation`} im nicht offenen und aus {gültig, `navigation_nicht_unterstuetzt`, `keine_navigation`} im offenen Typ an den Stellen 0 bis 2 von Listen der Länge 2 und 3 (die vier ersten Paare der Zusatztabelle in §6 als benannte Unterfälle); Orakel: der erste Verstoß der Liste | Die Klassen werden nach Klasse (400 vor 501) statt nach Stelle entschieden; der letzte Verstoß entscheidet | andere |
| 8 | Programmfehler sind `model.ErrUngueltig` ohne Code und kein `ExpandFehler`: `o` gleich `nil`, leere Liste, leerer Name an jeder Stelle, der Name `*`, zwei Elemente mit bytegleichem Namen; `a` und `A` sind zwei Namen (`SPEC-263` Nr. 4, V7) | `TestPruefeExpandProgrammfehler`: Schleife über die Arten {`o` nil, `Elemente` nil, `Elemente` leer, leerer Name an Stelle 0 bis 2, `*` an Stelle 0 bis 2, Doppelter Name mit Abstand 0 bis 2 in gültigen und ungültigen Namen} und den Typ {nicht offen, offen}; `errors.Is(ErrUngueltig)`, `errors.As(*ExpandFehler)` falsch; Gegenprobe `a`, `A` bleibt gültig; Ist-Stand: der doppelte Name an jedem Paar von Stellen in Listen von 32 Elementen (Abstand 0 bis 30), Namen mit `*` und aus Leerraum sowie ein Name mit angehängtem Leerraum als Gegenfälle (kein Programmfehler, als fremder Name und als Member eines Typs); in Listen bis 4096 Elementen ein doppelter Name, ein leerer Name und `*` (`TestPruefeExpandLaenge`) | Eine leere Liste ist gültig; `*` wird als Name gesucht; der doppelte Name wird nur benachbart oder nur in einem Fenster von 8 bzw. 31 Elementen erkannt; Vergleich mit gefalteten Namen; `*x`, `**`, `a*`, Präfix `*`, nur Leerraum und `NBSP` sind Programmfehler; der Schlüssel der Dubletten-Prüfung ohne Leerzeichen am Rand | andere |
| 9 | Ein Programmfehler geht jedem Verstoß vor, auch einem früheren; er wird für den ganzen Knoten vor der Prüfung der Namen festgestellt (`SPEC-263` Nr. 4, fünfte Zeile der Zusatztabelle) | `TestPruefeExpandProgrammfehlerVor`: Ist-Stand: Schleife über die drei Arten des Programmfehlers mit Element (leerer Name, `*`, doppelter Name; `nil` und leere Liste haben keine Stelle für einen Verstoß, Zeile 8) × {Verstoß davor, Verstoß danach} × die Klassen des Verstoßes (fremd bzw. im offenen Typ nicht unterstützt, `keine_navigation`) | Die Programmfehler werden in der Schleife über die Namen erkannt (ein früherer Verstoß entscheidet) | andere |
| 10 | Jeder Verstoß ist genau ein `*ExpandFehler` mit genau einer Ursache, Klasse, Code und Status nach der Tabelle von Nr. 3: `errors.Is` trifft genau die Sentinel der Klasse, `Code()` ist `ODR-E6005` bzw. `ODR-E6001`, `Error()` hat den Wortlaut der Spezifikation (`SPEC-263` Nr. 3, V2) | `TestExpandFehler`: Schleife über die drei Ursachen, aus `PruefeExpand` und von Hand gebaut; Orakel: die Tabelle der Spezifikation; `errors.Is` gegen beide Sentinel und `ErrUngueltig`, `errors.As`, Code, Text; Ist-Stand: Zählprüfung gegen die Quelle (`go/parser`) und gegen `NumField`: der Fehlerwert trägt genau das Feld `Ursache`, jedes andere Feld ist der Nullwert, `%#v` und `%T` nennen nur Typ und Ursache | Eine Ursache mit der Klasse der anderen; die Codes vertauscht; `Error()` ohne die Ursache; eine vierte Ursache; ein zusätzliches Feld (exportiert oder nicht) | andere |
| 11 | Eine Ursache außerhalb der drei (nur ein von Hand gebauter `ExpandFehler`) hat die Klasse der ersten Zeile: `ErrURLUngueltig`, `ODR-E6005`, `ODR-E6005: Expand ungültig: <Ursache>` (`SPEC-263` Nr. 3) | `TestExpandFehlerFremdeUrsache`: Ursachen {leer, `x`, eine der drei in anderer Schreibweise}; `errors.Is` auf beide Sentinel, Code, Text | Der Standardzweig wählt die Klasse `nicht_unterstuetzt` | andere |
| 12 | Kein Teil der Eingabe steht im Fehler: weder der Name noch die Stelle in der Liste noch der Name des Typs (`SPEC-263` Nr. 3) | `TestPruefeExpandFehlertext`: Schleife über alle Verstöße (die drei Ursachen) × Stelle 0 bis 2 mit unverwechselbarem Namen aus Bytewerten, der in keinem Typ des Modells vorkommt; `Error()`, `%v` und `%+v` enthalten weder Namen noch Stelle noch Typnamen; Ist-Stand: dazu `%#v`, `%T` und die Felder des Werts über `reflect` (siehe Zeile 10) | Der Name steht in der Meldung; die Stelle steht in der Meldung; ein Feld des Fehlerwerts trägt den Namen | andere |
| 13 | Eine Navigation ohne `NavigationBinding` am Set, auch mit leerem Ziel, ist gültig: die Funktion kennt Sets und Bindings nicht, die Ausführung entscheidet (`SPEC-263` Nr. 5, V5) | `TestPruefeExpandOhneBinding`: Schleife über Navigationen mit leerem `Target`, mit `Target` auf einen Typ, den das Modell nicht führt, und als Collection; Ergebnis `nil` | Die Prüfung verlangt ein nichtleeres Ziel | andere |
| 14 | Die Funktion ändert `typ` und `o` nicht (`SPEC-263` Nr. 1) | `TestPruefeExpandOhneAenderung`: tiefe Kopie von `typ` und `o` vor dem Aufruf, Vergleich danach; Schleife über alle Fälle der Zeilen 1 bis 9 (gültig und Verstoß) | Die Listen des Typs oder des Knotens werden sortiert | andere |
| 15 | Gleiche Eingabe ergibt gleiches Ergebnis; die Funktion hat keine eigene Grenze: Listen bis 32 Elemente (Grenze des Parsers) sind gültig (`SPEC-263` Nr. 6) | `TestPruefeExpandWiederholung`: jede Eingabe der Zeilen 1 bis 9 dreimal in Folge, tiefe Gleichheit; Ist-Stand: `TestPruefeExpandLaenge` führt Listen der Länge 1 bis 70 mit einem Verstoß (fremd, Property) an jeder Stelle und der Längen 100, 255, 256, 257, 1000 und 4096 mit Verstößen an den Stellen um jede Schwelle (erste, zweite, mittlere, zwei letzte, 31 bis 33, 63 bis 65, bis 4095), dazu einen doppelten Namen, einen leeren Namen und `*` (Typ mit 4096 Navigationen, offen und nicht offen) | Ein Zähler im Paket färbt das zweite Ergebnis; eine Obergrenze von 16, 31, 32, 33, 63, 256 oder 2000 Elementen; die Schleife prüft nur die ersten 32, 64, 255, 1000 oder 4095 Elemente; die Prüfung auf doppelte oder leere Namen entfällt über 32 bzw. 100 Elementen | andere |
| 16 | Gleichzeitige Aufrufe auf einem Typ und einem Knoten beeinflussen sich nicht; die Funktion ändert keine Variable des Pakets (`SPEC-263` Nr. 6) | `TestPruefeExpandGleichzeitig`: 8 Goroutinen × 50 Runden × alle Eingaben der Zeilen 1 bis 12 auf einem gemeinsamen `typ` und `o`, gleichzeitiger Start über eine Schranke; Ergebnis je Aufruf gegen den Einzelaufruf, verglichen erst nach dem Ende aller Goroutinen (Wirkungs-Fänger, nicht der Detektor); `make test-race` als zweiter Fänger je Prozess | Eine Paketvariable hält das Ergebnis der letzten Suche (ohne Sperre) | nebenläufig |
| 17 | `model` importiert keinen Port, keinen Dienst und keinen Adapter: die Funktion liest nur ihre zwei Argumente, ohne `ctx` und ohne Zugriff auf Daten (`SPEC-263` Kopf, `LH-FA-16.a`) | `make a-check` (Regeln des Hexagons, Ausgabe grün); die Signatur trägt kein `ctx` und kein Interface (Zeile 18) | Ein Import von `ports/outbound` im Paket `model` macht `make a-check` rot | andere |
| 18 | Validator-Contract-Test: die Funktion antwortet auf die Fälle der Spezifikation, wie sie sie nennt (`LH-FA-16.a`, `SPEC-263` Nr. 2 a bis d, 3, 4) | `TestPruefeExpandContract`: Tabelle der Fälle als Fixture (Typ, Liste, Ergebnis oder Ursache), ein Fall je Zeile, aus dem Wortlaut von Nr. 2 a bis d, 3 und 4 (die Spezifikation führt keine Beispieltabelle, §6); keine Eingabe aus dem Generator; Signatur über `reflect` gegen `PruefeExpand(EntityType, *ExpandOption) error` | Die Ursache einer Zeile wird vertauscht; Ist-Stand: die Signatur wird variadisch (`_ ...any`), denn eine Signatur mit `ctx` ist in jedem Test ein Übersetzungsfehler (`SPEC-210` Nr. 5 d) | andere |
| 19 | Akzeptiertes Negativ: die Abbildung von `ExpandFehler` auf Code und Status ändert sich nicht und wird in diesem Slice nicht geprüft (`SPEC-109` Nr. 5, `SPEC-263` Nr. 3) | kein Fänger hier, der die Abbildung fährt; Grund: `model` darf `ports/outbound` nicht importieren; die Kette Fehler zu Sentinel trägt Zeile 10, die Zeilen von `abbildung_test.go` (Sentinel zu Code) bleiben unverändert grün; Folge: eine Änderung der Abbildung, die einen `ExpandFehler` anders behandelt, wird erst im Test des Aufrufers rot (`slice-expand-ausfuehrung`, dessen DoD trägt die Abbildung); Maß: 1 von 3 Gliedern der Kette (Fehler, Sentinel, Code) liegt außerhalb dieses Slice | — | andere |
| 20 | Akzeptiertes Negativ: die Namen in `typ.Properties` und `typ.NavigationProperties` sind eindeutig; die Funktion prüft es nicht (`SPEC-263` *Ungeprüft*) | kein Fänger; Grund: der Metadatenleser liefert jeden Namen einmal (`SPEC-125` Nr. 8 und 9 c); Folge: steht ein Name in beiden Listen, gilt Nr. 2 a (Navigation zuerst) und der Name ist gültig; Maß: 1 Fall von 1 (Name in beiden Listen) ohne Fänger | — | andere |
| 21 | Akzeptiertes Negativ: ein Datenrennen, das kein Lauf auslöst (`SPEC-263` *Ungeprüft*, `SPEC-122` Nr. 8 und 9) | kein Fänger über Zeile 16 hinaus; Grund: die Funktion hat heute keine Variable des Pakets; Folge: eine später eingeführte Paketvariable, die nur ein Zweig schreibt, den Zeile 16 nicht trifft, bleibt vom Detektor ungemeldet; Maß: Zeile 16 führt die vier Fälle (a) bis (d) von Nr. 2 und die fünf Arten des Programmfehlers von Nr. 4 je mindestens einmal, 0 von 9 Zweigen der Prüfung ohne Lauf | — | andere |

**Abweichungen der Umsetzung vom Entwurf der Tabelle** (Ort: die Zeilen oben tragen den Entwurf, dieser Absatz den Ist-Stand; keine Abweichung ändert eine Zusage oder braucht eine Entscheidung des Architect):

- Zeilen 1, 5, 6: `PruefeExpand` sieht nur den wirksamen Typ, nicht das Modell und nicht die Subtypen; die Mutationen „Geerbte Navigationen werden nicht gesucht“, „die Namen aller Typen des Modells werden gesucht“ und „die Prüfung sucht auch in den Subtypen“ lassen sich im Code der Funktion nicht schreiben. An ihre Stelle treten Mutationen, die dieselbe Zusage brechen (Suche ohne die erste, letzte oder alle bis auf die ersten acht Navigationen; abstrakter und abgeleiteter Typ lassen fremde Namen zu; ein Name gleich dem `Target` einer Navigation ist gültig; ein Typ ohne Basis lässt fremde Namen zu), Tabelle unten.
- Zeile 18: eine Signatur mit `ctx` ist in jedem Test ein Übersetzungsfehler und kein Rot (`SPEC-210` Nr. 5 d); die Mutation ist die variadische Form `_ ...any`, die sich übersetzt und die der Vergleich über `reflect` fängt.
- Zeile 9: `o` gleich `nil` und eine leere Liste haben keine Stelle für einen Verstoß davor oder danach; die Schleife läuft über die drei Arten mit Element (leerer Name, `*`, doppelter Name) × Lage × Verstoß (fremd, im offenen Typ nicht unterstützt, Property, wo der Typ eine hat); die zwei anderen Arten führt Zeile 8.
- Zeilen 7 und 8 sind über den Entwurf hinaus verbreitert, weil die Klasse der Zusage größer ist als die Auswahl des Entwurfs (`AGENTS.md` §3.17): der erste Verstoß von links gilt in Listen bis 32 Elementen (jede Stelle 0 bis 31 und jedes Paar von Stellen, Zählung unten), der doppelte Name gilt für jeden Abstand 0 bis 30 und nicht nur 0 bis 2.
- Zeile 10: die Zählprüfung gegen die Quelle liest `expand_typ_fehler.go` mit `go/parser` und vergleicht die Werte aller Konstanten vom Typ `ExpandUrsache` mit den Zeilen der Tabelle von `SPEC-263` Nr. 3 (Mutation `z10e` fügt eine vierte hinzu); die Spezifikation liegt im Container der Tests nicht vor, die Zahl der Zeilen der Tabelle steht im Test als Tabelle.
- Zeilen 14 bis 16: die Eingaben sind die 8591 Fälle der Zeilen 1 bis 9 und 12 mit den Typachsen (`exvAlleFaelle`); Zeile 13 hat eigene Typen, die Zeilen 7 und 8 in Listen bis 32 Elementen und die Längen der Zeile 15 laufen nur in ihren Tests.
- Nach dem Review (`2026-11-22-slice-expand-validator-review`, Befunde F-1 bis F-5, INFO-1, `AGENTS.md` §3.17; die Mengen sind aus dem Wortlaut von `SPEC-263` gebildet, nicht aus den Mutationen des Reviews): Zeilen 4 und 5 führen das Produkt der Typmerkmale (offen × abstrakt × abgeleitet, 8 Kombinationen, `TestPruefeExpandTypachsen`), Zeile 15 die Länge der Liste weit über jede plausible Schwelle (`TestPruefeExpandLaenge`, Prüfgrenze 4096 in den Akzeptierten Negativen), Zeilen 10 bis 12 den Inhalt des Fehlerwerts über alle Felder (`reflect`, `%#v`, `%T`), Zeile 8 die Namen, die nur fast ein Programmfehler sind, Zeile 2 `é` auch in einer Property, und `Keys` jedes Typs nennt die erste Property und einen fremden Namen (`Schluessel`).

**Mengen** (Generator · Orakel · Zahl, die der Test zählt; Fixture-Modell: 18 Typen in beiden Folgen der Listen, also 36 Fälle des Typs, davon 12 offen und 24 nicht offen: Basis, abgeleiteter Typ, dritte Stufe, abstrakter Typ mit Konkret und einem abstrakten abgeleiteten Typ, offener Typ mit abgeleitetem offenem Typ, ein abstrakter offener Typ mit einem abstrakten offenen abgeleiteten Typ, drei Typen mit Entity-Navigationen, Typ ohne Navigation, leerer und leerer offener Typ, je ein Typ mit 32 Navigationen offen und nicht offen; die acht Kombinationen aus offen, abstrakt und abgeleitet sind alle geführt (Zählprüfung in `TestPruefeExpandFixture` über die Achsen); Orakel: die eigenen Listen jedes Typs und die Kette der Basistypen, in `TestPruefeExpandFixture` gegeneinander geprüft, kein Code der Funktion; wo eine Zählung Member zählt, folgt sie aus den eigenen Listen der Kette (`exvZaehlen`), nicht aus den wirksamen Listen):

| Nr. | Menge | Achsen | Fälle |
|---|---|---|---|
| 1 | jede wirksame Navigation (einzeln), dazu jede Liste aller Navigationen aufsteigend und absteigend | Typ (36) × Navigation × Form (einzeln, aufsteigend, absteigend) | 232 |
| 2 | jede Navigation und jede Property in 4 Abwandlungen (erster Buchstabe umgeschaltet, klein, groß, Leerzeichen angehängt) und, bei `é` (NFC, in einer Navigation und in einer Property), in NFD | Typ × Name × Abwandlung | 1166 |
| 3 | jede wirksame Property, auch geerbt und mit nicht unterstütztem Typ | Typ (offen und nicht offen) × Property | 102 |
| 4 | im offenen Typ jeder Name des Fixture-Modells außerhalb beider Listen (fremder Name, `Schluessel` aus `Keys`, Typname, Name eines anderen Typs, eines Subtyps), dazu die Abwandlungen der Navigationen | offener Typ (6 × 2 Folgen) × Name | 1630 |
| 5 | im nicht offenen Typ, auch im abstrakten, jeder Name des Fixture-Modells außerhalb beider Listen | nicht offener Typ (12 × 2 Folgen) × Name | 2636 |
| 4, 5 (Typachsen) | je Typ der acht Kombinationen offen × abstrakt × abgeleitet ein Name je Art: Navigation, Property, fremd, `*` (Orakel: die Achse offen der Schleife, 501 nur im offenen Typ, `*` Programmfehler) | Kombination (8) × Typ × Art (4) | 134 |
| 6 | jedes Paar (Typ, abgeleiteter Typ) bis zur dritten Stufe × jede Navigation und Property, die nur der abgeleitete Typ führt, am Typ und am abgeleiteten Typ | Paar × Member × Seite | 168 |
| 7 | jede Folge der Länge 1 bis 4 aus gültig, fremd, Property (Stelle 0 bis 3) in den fünf Typen mit mindestens 4 Navigationen und 4 Properties, in beiden Folgen | Länge × Folge × Typ | 1200 |
| 7 (lang) | in den Typen mit 32 Navigationen jede Stelle 0 bis 31 mit fremd oder Property und jedes Paar von Stellen (i < j) mit jedem Paar der zwei Klassen | Typ (4) × (64 + 496 × 4) | 8192 |
| 8 | `nil`, `Elemente` nil und leer, leerer Name und `*` an Stelle 0 bis 2, doppelter Name (Navigation, Property, fremd) an sechs Stellenpaaren in Listen der Länge 4 | Typ (36) × Art | 912 |
| 8 (lang) | doppelter Name an jedem Paar von Stellen (i < j) in Listen der Länge 32, als Navigation, Property und fremder Name | Typ (36) × 496 × Klasse (bis 3) | 48 608 |
| 8 (fast) | die Namen `*x`, `a*`, `**`, `*Name`, `" "`, `"  "`, Tab, `NBSP`, Zeilenumbruch als fremder Name; je Name (Navigation, Property, fremd) vier Paare mit Leerraum an einem Namen (`n` und `n `, `n ` und `n`, ` n` und `n`, `n` und Tab `n`); dazu ein Typ (offen und nicht offen) mit den vier Namen mit `*` als Navigationen und den fünf Leerraum-Namen als Properties, einzeln und als Listen | Typ (36) × (9 + 4 × Klasse) + 26 | 742 |
| 9 | leerer Name, `*`, doppelter Name × Verstoß davor oder danach × fremd oder Property | Typ × Art (3) × Lage (2) × Verstoß | 408 |
| 13 | Ziel leer, Ziel außerhalb des Modells, Ziel im Modell, kurzer Name des Ziels × Collection (2), einzeln und alle zusammen | Typ (offen, nicht offen) × Navigation | 18 |
| 15 | Länge 1 bis 70 (jede) sowie 100, 255, 256, 257, 1000, 4096; je Länge die gültige Liste, ein Verstoß (fremd, Property) an jeder Stelle bis 70, darüber an der ersten, zweiten, mittleren, den zwei letzten Stellen und an 31, 32, 33, 63 bis 65, 127 bis 129, 255 bis 257, 1023 bis 1025, 4095, dazu ein doppelter Name an bis zu vier Stellenpaaren und ein leerer Name und `*` an der ersten, mittleren und letzten Stelle; Typ mit 4096 Navigationen, offen und nicht offen (Zählprüfung gegen die Stellenmengen, deren Eigenschaften der Test prüft) | Länge (76) × Stelle × Art × Typ (2) | 11 930; die Fälle bis 70 Elemente dreimal in `TestPruefeExpandWiederholung` |
| 16 | alle 8591 Eingaben, 8 Goroutinen × 50 Runden, jede ab einer anderen Stelle | Eingabe × Goroutine × Runde | 3 436 400 Aufrufe |

Zeile 10 führt die drei Ursachen auf zwei Wegen (von Hand, aus `PruefeExpand`), Zeile 11 zählt 12 fremde Ursachen (3 leer oder fremd, 3 Ursachen × 3 Schreibweisen), Zeile 12 zählt 12 Fälle (4 Paare aus Ursache und Modus × Stelle 0 bis 2, Zählung im Test); Zeile 18 führt 31 Fälle der Spezifikation, die drei Ursachen der Tabelle von Nr. 3 und die Signatur.

**Mutationen, vom Implementer gefahren** (`sh tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, Klasse `andere`, ein Prozess je Stufe; zwei Läufe zusammen mit 59 Mutationen, Ergebnis 58 `rot`, 0 `rot-paket`, 1 `gruen` (`z20`, akzeptiertes Negativ der Zeile 20), 0 Übersetzungsfehler; Host-Last der Ergebniszeilen 4,3 bis 4,9 bei 4 CPUs im Lauf der ersten fertigen Tests, 8,1 bis 30,6 im allerersten Lauf (dasselbe Ergebnis), Messender der Lauf; derselbe Satz auf den Tests nach der Fix-Runde unten: dasselbe Ergebnis; die Namen tragen die Nr. der Zeile; jeder Test wird aus dem Grund seiner Zeile rot, die Mutation verändert nur die benannte Stelle):

| Nr. | Mutationen (alle `rot`, wo nichts anderes steht) | roter Test |
|---|---|---|
| 1 | `z01a` Suche ohne die erste, `z01b` ohne die letzte, `z01e` nur die ersten acht Navigationen, `z01c` nur Navigationen auf eine Entity, `z01d` nur auf eine Collection | `TestPruefeExpandGueltig` |
| 2 | `z02a` `strings.EqualFold` bei Navigationen, `z02c` ASCII-Faltung bei Properties, `z02b` NFD zu NFC vor dem Vergleich | `TestPruefeExpandSchreibweise` |
| 3 | `z03a` Property mit nicht unterstütztem Typ ist unbekannt, `z03b` `Open` vor den Properties | `TestPruefeExpandProperty` |
| 4 | `z04a` `Open` nicht gelesen, `z04b` offener Typ macht jeden Namen gültig | `TestPruefeExpandOffen` |
| 5 | `z05a` abstrakter Typ lässt fremde Namen zu, `z05b` abgeleiteter Typ, `z05c` unbekannt als 501, `z05d` Name gleich dem `Target` einer Navigation ist gültig | `TestPruefeExpandUnbekannt` |
| 6 | `z06a` Typ ohne Basis lässt fremde Namen zu, `z06b` abgeleiteter Typ | `TestPruefeExpandSubtyp` |
| 7 | `z07a` 501 nach 400, `z07c` unbekannt nach Property, `z07b` letzter Verstoß, `z07d` nur das erste Element, `z07e` die ersten drei, `z07f` die ersten 31, `z07g` ab Stelle 20 entscheidet der letzte Verstoß | `TestPruefeExpandRang` |
| 8 | `z08a` leere Liste gültig, `z08g` nur `nil` ist leer, `z08f` `o` nil gültig, `z08b` `*` gesucht, `z08c` leerer Name gesucht, `z08h` und `z08i` leerer Name bzw. `*` nur an Stelle 0, `z08d` doppelter Name nur neben dem ersten, `z08e` ASCII-gefaltet verglichen, `z08j` Fenster von 8 Elementen, `z08k` von 31 | `TestPruefeExpandProgrammfehler` |
| 9 | `z09a` Programmfehler in der Schleife über die Namen | `TestPruefeExpandProgrammfehlerVor` |
| 10 | `z10a` Ursachen der Klassen vertauscht, `z10b` Codes vertauscht, `z10c` und `z10d` Text ohne Ursache (400, 501), `z10e` vierte Ursache, `z10f` `Is` trifft `ErrUngueltig`, `z10g` `Is` trifft beide Klassen | `TestExpandFehler` |
| 11 | `z11a` Standardzweig wählt `nicht_unterstuetzt` | `TestExpandFehlerFremdeUrsache` |
| 12 | `z12a` Name, `z12b` Stelle, `z12c` Name des Typs im Fehler | `TestPruefeExpandFehlertext` |
| 13 | `z13a` die Prüfung verlangt ein nichtleeres Ziel | `TestPruefeExpandOhneBinding` |
| 14 | `z14a`, `z14b`, `z14c` sortiert die Navigationen des Typs, die Properties des Typs, die Elemente des Knotens | `TestPruefeExpandOhneAenderung` |
| 15 | `z15a` Zähler im Paket färbt das zweite Ergebnis (`TestPruefeExpandWiederholung`), `z15b` Grenze 16 Elemente, `z15c` Grenze 31 (beide auch `TestPruefeExpandLaenge`; Längen darüber: Fix-Runde) | `TestPruefeExpandWiederholung` |
| 16 | siehe Raten | `TestPruefeExpandGleichzeitig` (und `make test-race`) |
| 17 | Import von `ports/outbound` in `expand_typ.go`: `make a-check` meldet `core-impurity` (`src/hexagon/model/expand_typ.go:6: Kern importiert …/ports/outbound`, 1 Befund), ohne den Import 0 Befunde; von Hand gefahren, danach zurückgenommen | `make a-check` |
| 18 | `z18a` Property als unbekannt, `z18b` variadische Signatur, `z18c` offener Typ als Property | `TestPruefeExpandContract` |
| 20 | `z20` Reihenfolge der Prüfung von Navigation und Property vertauscht: `gruen` (2 Prozesse, beide Stufen), wie die Zeile es zusagt | kein Fänger |

**Fix-Runde nach dem Review** (`2026-11-22-slice-expand-validator-review`, Befunde F-1 bis F-5 und INFO-1; `AGENTS.md` §3.17: die Menge ist aus dem Wortlaut von `SPEC-263` Nr. 1 bis 6 gebildet, die Orakel stehen in den Schleifen der Achsen und in den eigenen Listen der Kette, nicht im Code der Funktion). Mutationen: `mutationslauf`, Klasse `andere`, ein Lauf mit 32 Mutationen (31 `rot`, 1 `gruen`: `f2-g1`, die Prüfgrenze von Zeile 15), dazu die vier Mutationen des Befunds F-3, die zwei Dateien ändern, von Hand gefahren (4 `rot`); die 59 Mutationen der Tabelle oben auf den gewachsenen Tests erneut: 58 `rot`, 1 `gruen` (`z20`); Host-Last der Ergebniszeilen 5,2 bis 20,1 bei 4 CPUs, Messender der Lauf. Der Produktcode blieb unverändert. `r…` sind die Mutationen des Reviews, `n…` Mutationen an einem Element, das der Befund nicht nannte:

| Befund | Menge · Achsen · Elementzahl je Achse | Fänger | Mutationen (alle `rot`, wo nichts anderes steht) |
|---|---|---|---|
| F-1 (Zeilen 4, 5) | Typmerkmale offen × abstrakt × abgeleitet = 8 Kombinationen, alle im Fixture (3 Typen neu: abstrakter abgeleiteter, abstrakter offener, abstrakter offener abgeleiteter Typ; Zählprüfung über die Achsen) × Art des Namens (Navigation, Property, fremd, `*`) × 2 Folgen der Listen = 134 Fälle; Orakel: die Achse offen (501 nur im offenen Typ, 400 sonst, `*` Programmfehler) | `TestPruefeExpandTypachsen`; die Mengen der Zeilen 4 und 5 tragen jetzt auch diese Typen (1630, 2636 Fälle) | `f1-r03` (offen und nicht abstrakt), `f1-r04` (offen und nicht abstrakt abgeleitet); `f1-n1` (offen und nicht abstrakt ohne Basis), `f1-n2` (nicht offen, abstrakt, abgeleitet: fremder Name gültig), `f1-n3` (nicht offen, abstrakt, abgeleitet: 501) |
| F-2 (Zeile 15) | Länge der Liste (76: 1 bis 70, 100, 255, 256, 257, 1000, 4096) × Stelle des Verstoßes (bis 70 jede, darüber 5 feste und die Schwellen 2^k - 1, 2^k, 2^k + 1 unter n) × Art (gültig, fremd, Property, doppelter Name, leerer Name, `*`) × Typ (offen, nicht offen) = 11 930 Fälle; Orakel: der erste Verstoß von links, durch die Bildung der Liste | `TestPruefeExpandLaenge` | `f2-r05` (Programmfehler über 32), `f2-r22` (über 33), `f2-r06` (Schleife nur über 32); `f2-n1` (über 63), `f2-n2` (über 256), `f2-n3` (über 2000), `f2-n4`, `f2-n5`, `f2-n6`, `f2-n7` (Schleife nur über 64, 255, 1000, 4095), `f2-r15` (doppelter Name nur bis 32), `f2-n8` (leerer Name nur bis 100); `f2-g1` (über 4096) `gruen`: Prüfgrenze |
| F-3 (Zeilen 10 bis 12) | Felder des Fehlerwerts: alle per `reflect` (Zählung gegen `NumField`, ein zugesagtes Feld), jedes andere Feld der Nullwert, `%#v` und `%T`, an 3 Ursachen × (von Hand, aus `PruefeExpand`, Verstoß an Stelle 0 bis 2 im offenen und nicht offenen Typ, Fall der Tabelle der Contract-Zeile) | `exvPruefeWert` in `TestExpandFehler`, `TestPruefeExpandFehlertext`, `TestPruefeExpandContract`; `TestExpandFehler` zählt `NumField` | `f3-r11` (Feld `Name`, ungenutzt); `f3-n1` (privates Feld `name`); von Hand mit zwei Dateien: Feld `Name` vom Code mit dem Namen belegt (rot in `TestExpandFehler` und `TestPruefeExpandFehlertext`), privates Feld `name` belegt (ebenso) |
| F-4 (Zeile 8) | Namen mit `*` (`*x`, `a*`, `**`, `*Name`), aus Leerraum (zwei Leerzeichen, Leerzeichen, Tab, Zeilenumbruch, `NBSP`), Formen mit Leerraum am Rand je Name × 4 Paare; als fremder Name, als Member (Sondertyp, offen und nicht offen), in 36 Typen = 742 Fälle | `TestPruefeExpandProgrammfehler` | `f4-r07` (`**`, `*x`), `f4-r19` (`a*`), `f4-r08` (Präfix `*`), `f4-r09` (nur Leerraum), `f4-r20` (Leerzeichen wie leer), `f4-r10` (Schlüssel der Dubletten ohne Leerzeichen am Rand); `f4-n1` (`NBSP` wie leer), `f4-n2` (Schlüssel ohne Tab am Anfang) |
| F-5 (Zeile 2) | `é` in einer Property (`Prüfer`, 1 neu) und in einer Navigation (`Café`) × 5 Abwandlungen je Name | `TestPruefeExpandSchreibweise` (1166 Fälle) | `f5-r16` (NFD zu NFC bei Properties); `f5-n1` (`strings.EqualFold` bei Properties) |
| INFO-1 (Zeile 5, 3) | `Keys` jedes Typs: die erste Property und `Schluessel`, ein Name, den keine Liste führt (in der Menge der fremden Namen) | `TestPruefeExpandUnbekannt`, `TestPruefeExpandProperty` | `i1-r18` (Name in `Keys` ist `keine_navigation`); `i1-n1` (Name in `Keys` ist gültig) |

**Raten der Nebenläufigkeit** (Zeile 16, `TestPruefeExpandGleichzeitig`, je Mutation 50 Prozesse mit Detektor und 50 ohne, `-test.count=1`, 8 Goroutinen je Prozess, höchstens 4 Prozesse zugleich, alle 8591 Eingaben; Host-Last 2,9 bis 7,9 bei 4 CPUs vor, zwischen und nach den Armen (Fix-Runde, nach der Erweiterung der Eingaben um die Typachsen; die erste Messung mit 6827 Eingaben bei Last 3,3 bis 7,8 gab dieselben Zahlen), zwei andere Linien teilen den Host; Messender: ein Skript des Implementers, das dasselbe fährt wie `SPEC-210` Nr. 7 b (zwei Test-Binaries aus einer Kopie des Baums, `-test.run '^TestPruefeExpandGleichzeitig$' -test.timeout=2m`, Container ohne Netz mit `--memory=4g --cpus=4`, unter `flock`); jede Zahl ist eine Größenordnung unter Last). **Grenze des Werkzeugs:** `mutationslauf` bricht die Klasse `nebenläufig` im Paket `model` in der Nullprobe ab (`Ausgangsstand rot: src/hexagon/model`): das Paket läuft unter dem Detektor 131 bis 182 s (der Bestandstest `TestLiesRueckgabeErschoepfend` allein 86 s bei Host-Last 11), die Grenze je Prozess ist 2 Minuten (`SPEC-210` Nr. 5 c, 9 c); die Klasse `andere` läuft ohne Detektor meist in 1,5 bis 2,9 s je Mutation. §6 trägt die Rückgabe. Unveränderter Stand: 50/50 grün mit und 50/50 grün ohne Detektor.

| Mutation (Zeile 16 und 21) | mit Detektor | ohne Detektor | Fänger |
|---|---|---|---|
| `z16d` gemeinsamer Zustand mit Fenster zwischen Prüfen und Handeln (Paketvariable `letzte` wird vor der Schleife auf nil gesetzt, in der Schleife belegt, am Ende gelesen) | 50/50 | 50/50 | Wirkung, in jedem Lauf; `make test-race` ebenso |
| `z16b` gemeinsamer Puffer der Elemente (Paketvariable, jeder Aufruf überschreibt ihn und liest ihn in der Schleife) | 50/50 | 50/50 | Wirkung, in jedem Lauf; `make test-race` ebenso |
| `z16a` Paketvariable, die das Ergebnis der letzten Suche hält (`letzte = f(); return letzte`, ein Zugriff unmittelbar hinter dem anderen) | 50/50 | 0/50 | nur `make test-race`: der Übersetzer reicht den geschriebenen Wert weiter, es gibt ein Rennen ohne Wirkung |
| `z16c` Zähler ohne Wirkung | 50/50 | 0/50 | nur `make test-race` |
| `z21` Zähler in einem Zweig, den kein Lauf trifft (Liste mit mehr als 32 Elementen) | 0/50 | 0/50 | keiner (Zeile 21) |

**Akzeptierte Negative** (`AGENTS.md` §3.15, §3.17; Zeilen 19 bis 21 stehen in der Tabelle oben mit Grund, Folge und Maß; gesehen: `z20` grün, `z21` 0/50 und 0/50). Dazu die Prüfgrenzen der Mengen, deren Klasse unbeschränkt ist:

| Zusage | Menge der Zusage | Prüfgrenze (geführt bis) | Grund | Folge im Betrieb | Maß |
|---|---|---|---|---|---|
| Zeile 2: Namen werden bytegleich gesucht | jede Abwandlung eines Namens (unbeschränkt) | 5 Abwandlungen: erster Buchstabe, klein, groß, Leerzeichen, `é` in NFD | die Menge der Normalformen und Foldings ist nicht abzählbar | ein Vergleich, der nur eine andere Normalform (NFKC, andere kombinierende Zeichen) oder ein anderes Zeichen als `é` angleicht, bleibt ungefangen; der Parser liefert nach `SPEC-262` Nr. 4 a nur Buchstaben und Ziffern, ein kombinierendes Zeichen kommt nie an | von 5 gefahrenen Angleichungen (`strings.EqualFold`, ASCII-Faltung und NFD zu NFC an Navigationen oder Properties: `z02a`, `z02b`, `z02c`, `f5-r16`, `f5-n1`) alle rot; 0 von 5 ungefangen, die Menge darüber ist ungeführt |
| Zeilen 1, 4, 5: jeder Name jedes Typs | Namen und Typen aller Modelle (unbeschränkt) | 18 Typen (alle 8 Kombinationen aus offen, abstrakt und abgeleitet), 32 Navigationen je Typ, 4 Properties je Typ mit 32 Navigationen | Laufzeit und kein Anhaltspunkt für eine Schwelle (`SPEC-263` Nr. 6 nennt keine) | eine Regel, die erst ab 33 Navigationen oder 5 Properties greift (und ohne Familie, die sie trifft), bleibt ungefangen; der Parser liefert höchstens 32 Elemente, mehr Navigationen eines Typs sind möglich (Zeile 15 führt Typen mit 4096 Navigationen für die Länge der Liste) | von 1 gefahrenen Schwelle der Suche (nur die ersten 8 Navigationen, `z01e`) 1 rot; Typen mit mehr als 32 Navigationen führt allein Zeile 15 (4096) |
| Zeile 15: keine eigene Grenze (`SPEC-263` Nr. 6) | Länge der Liste (unbeschränkt) | 4096 Elemente; Verstoß an jeder Stelle bis Länge 70, darüber an der ersten, zweiten, mittleren, den zwei letzten Stellen und um jede Schwelle 2^k - 1, 2^k, 2^k + 1 bis 4095 | die Suche je Name ist linear in den Navigationen des Typs; eine gültige Liste von n Namen in einem Typ mit n Navigationen kostet n²/2 Vergleiche (4096: 8,4 Mio.; 65 536: 2,1 Mrd.), und der Test hält den Typ im Speicher | eine Regel, die erst über 4096 Elementen greift (Programmfehler, Prüfschleife oder Prüfung von doppelten und leeren Namen), bleibt ungefangen; der Parser liefert höchstens 32 (`SPEC-262` Nr. 6), ein Aufrufer ohne Parser liefert sie selbst | von 15 gefahrenen Regeln über der Länge (Programmfehler über 16, 31, 32, 33, 63, 256, 2000, 4096 Elementen; Schleife nur über die ersten 32, 64, 255, 1000, 4095; doppelter Name nur bis 32; leerer Name nur bis 100) 14 rot, 1 grün (`f2-g1`, über 4096); 1 von 15 ungefangen |
| Zeile 7: erster Verstoß von links | Folgen von Klassen in Listen bis 32 (3 Klassen) | alle Folgen bis Länge 4, in Listen von 32 jeder Verstoß und jedes Paar | die Menge der Folgen wächst mit 3^Länge | eine Regel, die erst bei drei oder mehr Verstößen in einer Liste von mehr als 4 Elementen greift, bleibt ungefangen | von 7 gefahrenen Regeln (`z07a` bis `z07g`) 7 rot |
| Zeile 8: doppelter Name | Paare von Stellen in Listen (unbeschränkt) | jedes Paar (i < j) in Listen von 32, in längeren Listen bis 4096 vier Paare (erste und letzte Stelle, erste zwei, letzte zwei, mittlere und letzte) | der Parser liefert höchstens 32; darüber gilt die Prüfgrenze der Zeile 15 | eine Regel, die nur ein Paar in einer Liste über 32 Elementen jenseits dieser vier Paare nicht erkennt, bleibt ungefangen | von 6 gefahrenen Regeln (`z08d`, `z08e`, `z08j`, `z08k`, `f4-r10`, `f2-r15`) 6 rot |
| Zeile 8: Namen, die kein Programmfehler sind | Zeichenfolgen eines Namens (unbeschränkt) | `*x`, `a*`, `**`, `*Name`, vier Formen Leerraum (Leerzeichen, zwei Leerzeichen, Tab, Zeilenumbruch), `NBSP` und die Formen eines Namens mit Leerzeichen oder Tab am Rand | die Menge der Zeichenfolgen ist unbeschränkt; `SPEC-263` Nr. 4 nennt allein den leeren Namen und `*` | eine Regel, die ein anderes Zeichen (Nullbreitenzeichen, `BOM`) oder eine andere Folge zum Programmfehler macht, bleibt ungefangen | von 8 gefahrenen Regeln (`f4-r07`, `f4-r08`, `f4-r09`, `f4-r19`, `f4-r20`, `f4-n1`, `f4-r10`, `f4-n2`) 8 rot |
| Zeile 16: Nebenläufigkeit | Zahl der Goroutinen und Runden | 8 × 50 | Laufzeit unter dem Detektor (rund 2 s je Prozess ohne Last; unter `-coverprofile` mit Detektor 585 s für das Paket) | ein Zustand, der erst bei mehr als 8 gleichzeitigen Aufrufen auf dem Fenster wirkt, bleibt ohne Wirkungs-Fänger; der Detektor meldet das Rennen | von 4 Mutationen mit Zustand im Paket sind 2 (`z16b`, `z16d`) mit Wirkung gefangen, in jedem Lauf; 2 (`z16a`, `z16c`) fängt nur der Detektor (50/50, ohne 0/50); `z21` fängt keiner |

**Nullblöcke:** `go test -race -coverprofile` über das Paket `model` (Image `odata-recorder:race`, 584,8 s, Host-Last 5 bis 10 bei 4 CPUs): in `expand_typ.go` und `expand_typ_fehler.go` null Blöcke mit Zähler 0, jede Funktion 100,0 %. Die 36 Blöcke mit Zähler 0 im Paket liegen im Bestand: `abfrage_filter.go` (3), `baseline.go` (5), `exchange.go` (1), `filter_ast.go` (6, leere Blöcke ohne Anweisung), `redaktor.go` (2), `redaktor_text.go` (2), `schluessel_json.go` (17). Der Lauf entstand vor dem Nachtrag der Generatoren „7 (lang)“ und „8 (lang)“ und vor der Fix-Runde nach dem Review; der Produktcode ist seither unverändert (die Datei `expand_typ.go` und `expand_typ_fehler.go` tragen dieselben Bytes), die Tests sind nur gewachsen.

**Gegengelesen** (`AGENTS.md` §3.18): `src/hexagon/model/expand_typ_fixture_test.go` auf Rohbytes (`LC_ALL=C grep -P`): `é` NFC (`\xc3\xa9`) und NFD (`e\xcc\x81`) entstehen aus Bytewerten in `exvEAkut` und `exvEAkutNFD`, im Quelltext steht kein `\u`; `expand_typ_fehler_test.go`: `ä` in `geheim` aus den Bytes `\xc3\xa4`; die Mutation `z02b` trägt die Bytes als `\xcc\x81` und `\xc3\xa9` in der Liste.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen; die Menge der Fälle des Validators sind alle Navigationen und alle Properties jedes Typs des Fixture-Modells, einschließlich der geerbten.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-expand-parser` liegt in `done/` (der Knoten und seine Form, Entscheidung A8 des Architect dort). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-expand-parser` (Typ des Knotens in `src/hexagon/model/`; dieser Slice liest ihn und kann ihn nicht vorwegnehmen). Das Modell mit den Navigation-Properties (`SPEC-002`, `SPEC-125` Nr. 11) liegt in `done/` (`slice-odata-codec-csdl`). Der Welle-Trigger (`welle-mock` in `done/`) ist Reihenfolge-Vorsicht, keine Code-Abhängigkeit.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Query-Interpreter, JSON-Schreiber, Workspace, Composition Root, M2, `slice-bootstrap-m3` (die Funktion hat keinen Aufrufer), `slice-expand-ausfuehrung` (Folge-Slice), `slice-batch-parser`.
**Stand der Bedingungen beim Start** (Planner, 2026-11-21, `main` = `a8d8f720` nach den Moves): `slice-expand-parser` liegt in `done/` (der Knoten `ExpandOption` mit `ExpandElement.Name` besteht in `src/hexagon/model/url_ast_expand.go`; der Fehlertyp `ExpandFehler` und `PruefeExpand` bestehen noch nicht); `welle-mock` liegt in `done/`; `SPEC-263` mit Nr. 1 bis 6, *Ungeprüft* und dem Raster stehen auf `main` (Architect). Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` oder des Stempels. Überschneidung mit dem Slice in Arbeit `slice-admin-sessions` (`src/adapters/driving/admin/`, `src/bootstrap/`): keine gemeinsame Datei, beide laufen zugleich. Sendungen (`AGENTS.md` §3.13): *an* diesen Slice der Validator samt Fehlertyp, Raster und Contract-Test (Geber `slice-expand-parser`, in §1 und DoD, in `done/`); *von* diesem Slice die Behandlung einer Navigation ohne `NavigationBinding`, der Aufruf des Validators und die Abbildung seiner Fehler an `slice-expand-ausfuehrung`, dessen §1 *Übernimmt* sie trägt (am 2026-11-21 gelesen, kein *Ausdrücklich NICHT* dort trifft die Sendungen, der Plan liegt in `open/`).

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect: `produktcode` 120, Teil des Planner 2026-11-21, vor dem ersten Code-Commit; Messung 0 beim Start): die Tabelle in §3 trägt 21 Zeilen, *a* = 1 (Zeile 16), *b* = 20 (die Zeilen 19 bis 21, akzeptierte Negative, zählen als `andere`, `SPEC-128` Nr. 3 e), *f* = 0 (die Funktion hat keinen Port), *p* = 120: *S* = 85·1 + 80·20 + 200·0 + 120 = 1805 gleich dem Block, 135·1805 = 243675, *M* = 100 · ⌈24,37⌉ = 2500 gleich der Formel; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`; der Slice liegt 418 unter der Schnittgrenze. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `basis` ist der Commit des Moves nach `in-progress/` (`a8d8f720`; ein Nachzug von Verweisen war nicht nötig). Pfad: `src/hexagon/model/` als Verzeichnis (Code und Tests von `PruefeExpand` und `ExpandFehler`; kein anderer Slice in Arbeit ändert `model`). Die Zeilen, die ein Merge von `main` fremd in `src/hexagon/model/` trüge, misst der Implementer vor dem Merge (`SPEC-128` Nr. 10 g). Die Änderung der Spezifikation (`spec/`) liegt außerhalb der Metrik (Nr. 3 m).

```groessenmarke
pfad: src/hexagon/model/
produktcode: 120
fakes: 0
basis: a8d8f720
schaetzung: 1805
marke: 2500
```

**Messung nach dem Code** (Implementer, `git diff --numstat a8d8f720` über `src/hexagon/model/`, Tests eingeschlossen, vor dem Handoff-Commit): 1713 Zeilen in acht neuen Dateien (Produktcode 102: `expand_typ.go` 56, `expand_typ_fehler.go` 46; Tests 1611; vor der Fix-Runde nach dem Review 1403), also 787 unter der Marke 2500; kein Pfad außerhalb des Blocks berührt (außer diesem Plan und dem Report unter `docs/reviews/`), kein `merge` von `main` bis zur Messung.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem Code und nur über der Schnittgrenze (`SPEC-128` Nr. 1); die Schätzung liegt darunter, ein weiterer Schnitt ist nicht vorgesehen: Fehlertyp und Funktion teilen den Katalog der Ursachen. Eine Messung über der Marke (2500) nach dem Code führt nicht zurück (`AGENTS.md` §3.14): der Slice schließt mit der Messung und dem Grund im Eintrag `ueberschreitung`.
- `in-progress` → `open` (blockiert — Carveout?): die Form des Knotens aus `slice-expand-parser` trägt die Prüfung nicht, oder das Modell liefert die Navigationen eines Typs nicht so, wie `SPEC-125` Nr. 11 sie beschreibt.

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

- Die Menge der Fälle ist die Menge der Navigationen und Properties aller Typen eines Modells (unbeschränkt) — **Ausgang:** offen; Schleife über die Menge des Fixture-Modells mit Subtypen, Ketten und offenem Typ, Prüfgrenze und Folge als akzeptiertes Negativ (`AGENTS.md` §3.17); Umsetzung: Mengen, Prüfgrenzen und Maß stehen in §3 (Absätze *Mengen* und *Akzeptierte Negative*).
- Umsetzung: `mutationslauf` fährt die Klasse `nebenläufig` im Paket `model` nicht: die Nullprobe bricht mit `Ausgangsstand rot: src/hexagon/model` ab, weil das Paket unter dem Detektor 131 bis 182 s läuft gegen die Grenze von 2 Minuten je Prozess (`SPEC-210` Nr. 5 a, 5 c, 9 c; der Bestandstest `TestLiesRueckgabeErschoepfend` braucht allein 86 s bei Host-Last 11, die neuen Tests rund 3 s). Die Raten der Zeile 16 stehen in §3 aus einem Skript des Implementers, das dasselbe fährt (zwei Test-Binaries, 50 Prozesse je Arm). **Ausgang:** offen. **Adresse:** Planner (Beobachtungs-Register: die Klasse `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` führt dieselbe Ursache schon mit 10 Dateien unter `evidence/`; dieser Slice ist eine weitere Evidenz, die der Planner bei der Closure als Datei dort ablegt, ohne neue Klasse) und Architect (`SPEC-210` Nr. 5 c und 9 c: die Grenze je Prozess gilt dem ganzen Paket der Nullprobe); `make test-race` und `make gates` sind nicht betroffen.
- Umsetzung (Fix-Runde, Zeile 15): `SPEC-263` Nr. 6 sagt „der Aufwand ist linear in der Zahl der Elemente und der Navigationen des Typs“; die Funktion sucht jeden Namen linear in beiden Listen, der Aufwand ist also das Produkt aus Elementen und Member (höchstens 32 Elemente aus dem Parser, bei einer Liste von n Namen in einem Typ mit n Navigationen n²/2 Vergleiche). Der Fänger der Länge geht deshalb bis 4096 (Prüfgrenze in §3). **Ausgang:** offen. **Adresse:** Architect (`SPEC-263` Nr. 6: gemeint ist linear je Element in den Membern des Typs, oder verlangt der Vertrag eine Form ohne Produkt; eine Änderung des Codes entscheidet der Implementer nicht).
- Fakes der Ports: die Funktion hat keinen Port — **Ausgang:** entfallen (`PruefeExpand` hat keinen Port, `SPEC-263`).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Raster unten, `SPEC-263`); die Tabelle *Zusage · Fänger* hat der Planner beim Start gefüllt (21 Zeilen, §3).
- Der Validator-Contract-Test (`LH-FA-16.a`) soll nach der DoD gegen „die Beispiele der Spezifikation“ fahren; `SPEC-263` führt keine Beispieltabelle, nur die Fälle von Nr. 2 a bis d, 3 und 4 — **Ausgang:** entschieden (Planner, beim Start): die Tabelle des Tests (Zeile 18) nimmt die Fälle aus dem Wortlaut dieser Nummern, die DoD ist berichtigt; will der Architect eine Beispieltabelle in `SPEC-263`, trägt er sie vor dem ersten Code-Commit nach und der Test nimmt sie auf (Frage im Bericht des Planners an den Orchestrator).

**Entscheidungen des Architect — Ausgang *entschieden*, Ort `SPEC-263`** (`AGENTS.md` §3.13). Keine ADR: die Entscheidungen schärfen die Spezifikation (`AGENTS.md` §3.8) und ändern keine angenommene ADR.

- **V1 Signatur.** Entschieden (`SPEC-263` Kopf, Nr. 1): `PruefeExpand(typ EntityType, o *ExpandOption) error`; `typ` ist der Typ des Entity-Sets mit den wirksamen Properties und Navigation-Properties, den geerbten eingeschlossen; das Modell wird nicht gebraucht (die Zieltypen sind beim Lesen aufgelöst, `SPEC-125` Nr. 9 d); `o` ist `Optionen.Expand`, ein Aufrufer ohne `$expand` ruft nicht.
- **V2 Fehlerwerte.** Entschieden (`SPEC-263` Nr. 3): `model.ExpandFehler{Ursache}` mit `navigation_unbekannt` und `keine_navigation` (`ODR-E6005`, 400) und `navigation_nicht_unterstuetzt` (`ODR-E6001`, 501), Sentinel `ErrURLUngueltig` bzw. `ErrURLNichtUnterstuetzt`; kein Teil der Eingabe im Fehler; `abbildung.go` und `SPEC-109` bleiben.
- **V3 Reihenfolge.** Entschieden (`SPEC-263` Nr. 2): der erste Verstoß von links entscheidet, gleich welcher Klasse (wie `SPEC-167` Nr. 3); Programmfehler gehen allem vor (Nr. 4).
- **V4 Subtyp und Typ-Cast.** Entschieden (`SPEC-263` Nr. 2 d): eine Navigation, die nur ein abgeleiteter Typ trägt, ist `navigation_unbekannt` (`typ` ist der Typ des Sets); der Typ-Cast im Pfad ist schon im Parser `501`; ein abstrakter Typ des Sets verhält sich wie jeder andere.
- **V5 Navigation ohne Binding.** Entschieden (`SPEC-263` Nr. 5): gültig, die Funktion kennt Sets und Bindings nicht. Die Behandlung in der Ausführung trägt `slice-expand-ausfuehrung` (dessen §1, Sendung von `slice-expand-validator`).
- **V6 Property ohne Navigation.** Entschieden (`SPEC-263` Nr. 2 b, c): jede Property des Typs, auch mit nicht unterstütztem Typ, ist `keine_navigation` (400); ein Name im offenen Typ, der in keiner der beiden Listen steht, ist `navigation_nicht_unterstuetzt` (501, wie `SPEC-167` Nr. 5).
- **V7 Programmfehler.** Entschieden (`SPEC-263` Nr. 4): `o` nil, leere Liste, leerer Name, `*` und doppelter Name sind `model.ErrUngueltig` ohne Code, für den ganzen Knoten vor der Prüfung der Namen.

**Raster *Operation × Argumentart*** (Zellen gefüllt vom Architect; der Ort steht in `SPEC-263`):

| Operation | Knoten fehlt oder leer | unbekannte Navigation | fremder Wert (Property ohne Navigation, anderer Typ, Subtyp) | geerbte Navigation | Wiederholung (Aufruf, Name) | Nebenläufigkeit |
|---|---|---|---|---|---|---|
| `PruefeExpand` | `SPEC-263` Nr. 4 | `SPEC-263` Nr. 2 d und Nr. 3 | `SPEC-263` Nr. 2 b, c, d (Property des Typs, Name im offenen Typ, Navigation eines Subtyps, Name eines anderen Typs) | `SPEC-263` Nr. 1 und Nr. 2 a (wirksame Navigationen, geerbte eingeschlossen) | `SPEC-263` Nr. 6 (Aufruf) und Nr. 4 (doppelter Name) | `SPEC-263` Nr. 6 |

**Zusatztabelle: mehrere Verstöße in einem Aufruf** (je Paar das Ergebnis nach `SPEC-263` Nr. 2 und 4; die Spalte *Fänger* nennt die Zeile der Tabelle in §3, die der Planner aus `SPEC-263` Nr. 1 bis 6 gefüllt hat):

| Paar (links, rechts im Knoten) | Ergebnis | Fänger |
|---|---|---|
| unbekannte Navigation, Property ohne Navigation | `navigation_unbekannt` (der erste von links, Nr. 2) | Zeile 7 (`TestPruefeExpandRang`, Paar im nicht offenen Typ in beiden Folgen) |
| Property ohne Navigation, unbekannte Navigation | `keine_navigation` (der erste von links, Nr. 2) | Zeile 7 (dasselbe Paar in der anderen Folge) |
| Name im offenen Typ, Property ohne Navigation | `navigation_nicht_unterstuetzt` (der erste von links, gleich welcher Klasse) | Zeile 7 (Paar im offenen Typ) |
| Property ohne Navigation, Name im offenen Typ | `keine_navigation` (der erste von links) | Zeile 7 (dasselbe Paar in der anderen Folge) |
| Verstoß, danach doppelter Name oder `*` | `model.ErrUngueltig` (Programmfehler vor jedem Verstoß, Nr. 4) | Zeile 9 (`TestPruefeExpandProgrammfehlerVor`, Verstoß davor und danach) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-18) und beim Start nachgezogen (Planner, 2026-11-21; der Zähler ist die Zahl der Dateien unter `evidence/`). Die Treffer von `slice-expand-parser` §6 gelten für die Schleife über Navigationen und Properties: `BEO-REPO/menge-ohne-ihre-positionen` (11×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/menge-im-plan-weicht-vom-orakel-des-tests-ab` (4×, verkörpert im Skill des Reviewers), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (61×, verkörpert in `AGENTS.md` §3.17). Für die Zeile 16 der Tabelle einschlägig: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×, `AGENTS.md` §3.10: der Wirkungs-Fänger nennt seine Rate mit Last und Messendem); außerdem verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (67×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (19×, `AGENTS.md` §3.15; die Zeilen 19 bis 21 tragen Grund, Folge und Maß) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×, Übergangsregeln Nr. 8). Keine Beobachtung steht mit diesem Slice offen unter 3×. Der Slice ist nicht nebenläufig, bis auf die Zusage der gleichzeitigen Aufrufe (`SPEC-167` Nr. 8 als Muster).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
