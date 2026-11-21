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

- [ ] Funktion und Fehlertyp: jede Navigation des Knotens gegen die wirksamen Navigation-Properties des Typs (auch geerbt); Property ohne Navigation, unbekannter Name und die nicht unterstützte Form mit Ursache, Klasse, Code und Status nach dem Vertrag des Architect; kein Teil der Eingabe im Fehler; kein Zugriff auf Daten.
- [ ] Randformen (Programmfehler des Aufrufers, Wiederholung, Nebenläufigkeit, Reihenfolge mehrerer Verstöße, Subtyp, Navigation ohne Binding): je Zelle des Rasters (§6) ein Ort in der Spezifikation und ein Fänger der Tabelle (§3), als Schleife über alle Navigationen und alle Properties des Fixture-Modells.
- [ ] Validator-Contract-Test (`LH-FA-16.a`): ein eigener Test fährt die Funktion gegen die Fälle von `SPEC-263` Nr. 2 a bis d, 3 und 4 (die Spezifikation führt keine eigene Beispieltabelle, §6; Zeile 18 der Tabelle); `make a-check` hält die Richtung (`model` importiert keinen Port und keinen Dienst).
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
| `src/hexagon/model/` | neu | `PruefeExpand` und der Fehlertyp (Muster `filter_typ.go`, `filter_typ_fehler.go`), technologieunabhängig |
| `src/hexagon/ports/outbound/abbildung.go` | unverändert | `ExpandFehler` trägt die Sentinel der URL-Klassen, die `SPEC-109` Nr. 5 schon abbildet (`SPEC-263` Nr. 3); keine Zeile neu |
| `spec/spezifikation.md` | update | Vertrag `SPEC-263` mit Randformen (Architect, geschrieben) |
| Tests | neu | Fälle je Navigation und je Fehlerklasse; Schleife über alle Navigationen und Properties des Fixture-Modells |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Größenmarke hat der Planner beim Start nach `SPEC-128` aus dieser Tabelle und der Erwartung des Architect gesetzt (§4): `produktcode` rund 120 (die Funktion rund 50 Zeilen, der Fehlertyp rund 70 nach dem Muster von `filter_typ_fehler.go`, 76 Zeilen). Die Tabelle *Zusage · Fänger* hat der Planner aus `SPEC-263` Nr. 1 bis 6 vor dem Start gefüllt (Auftrag des Plans; der Architect hat Vertrag, Entscheidungen V1 bis V7 und Raster geliefert, §6).

**Tabelle *Zusage · Fänger*** (Planner, 2026-11-21, aus `SPEC-263` Nr. 1 bis 6 und dem Raster in §6; Fixture-Modell und Orakel entstehen im Test: eine eigene Tabelle der erwarteten Navigationen und Properties je Typ ohne Code der Funktion, `AGENTS.md` §3.17; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10; Dateien liegen unter `src/hexagon/model/`):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Eine Navigation des Typs, auf eine Entity oder auf eine Collection, auch eine geerbte, ist gültig: Ergebnis `nil` (`SPEC-263` Nr. 2 a, Nr. 1) | `TestPruefeExpandGueltig`: Schleife über alle Typen des Fixture-Modells (Basis, abgeleitet, abstrakt, offen, Kette aus drei Stufen) und alle wirksamen Navigationen je Typ, einzeln und als Liste in aufsteigender und absteigender Folge; Orakel: die erwartete Liste je Typ als Tabelle im Test; Zählprüfung der Menge | Geerbte Navigationen werden nicht gesucht; nur Navigationen auf eine Entity werden gefunden | andere |
| 2 | Ein Name wird bytegleich gesucht: Groß- und Kleinschreibung zählt, keine Normalisierung (`SPEC-263` Nr. 1) | `TestPruefeExpandSchreibweise`: Schleife über alle Navigationen aller Typen × Abwandlungen {erster Buchstabe groß, alles klein, alles groß, angehängtes Leerzeichen, Normalform NFD statt NFC bei einem Namen mit Zeichen jenseits von ASCII}; Fixture aus Bytewerten, die Datei auf Rohbytes gegengelesen (`AGENTS.md` §3.18); Erwartung `navigation_unbekannt` | Vergleich mit `strings.EqualFold`; Normalisierung vor dem Vergleich | andere |
| 3 | Ein Name in `typ.Properties`, jede Property, auch eine geerbte und eine mit nicht unterstütztem Typ, ist `keine_navigation`, im offenen wie im nicht offenen Typ (`SPEC-263` Nr. 2 b, V6) | `TestPruefeExpandProperty`: Schleife über alle wirksamen Properties aller Typen, offen und nicht offen, darunter Properties mit nicht unterstütztem Typ; Zählprüfung | Eine Property mit nicht unterstütztem Typ zählt als unbekannt; die Prüfung von `Open` geht der der Properties vor | andere |
| 4 | Ein Name, der in keiner der beiden Listen steht, in einem offenen Typ, ist `navigation_nicht_unterstuetzt` (501), nicht `navigation_unbekannt` (`SPEC-263` Nr. 2 c, V6) | `TestPruefeExpandOffen`: Schleife über die Namen {fremder Name, Name einer Navigation eines anderen Typs, Name einer Navigation des Subtyps, Name mit anderer Schreibweise einer vorhandenen Navigation} × alle offenen Typen des Modells | `Open` wird nicht gelesen (alle `navigation_unbekannt`); ein offener Typ macht jeden Namen gültig | andere |
| 5 | Ein Name, der in keiner der beiden Listen steht, in einem nicht offenen Typ, ist `navigation_unbekannt` (400): auch der Name einer Navigation oder Property eines anderen Typs, in jedem Typ, auch im abstrakten (`SPEC-263` Nr. 2 d) | `TestPruefeExpandUnbekannt`: Schleife über alle Typen × {fremder Name, jede Navigation und jede Property jedes anderen Typs des Modells, die der Typ nicht trägt} × {abstrakt, nicht abstrakt}; Orakel: Menge der Namen je Typ aus der Tabelle des Tests | Ein abstrakter Typ liefert `nil`; die Namen aller Typen des Modells werden gesucht | andere |
| 6 | Eine Navigation, die nur ein abgeleiteter Typ trägt, ist am Typ des Sets `navigation_unbekannt` und am abgeleiteten Typ gültig; `typ` ist der Typ des Sets (`SPEC-263` Nr. 2 d, V4) | `TestPruefeExpandSubtyp`: Schleife über alle Paare (Typ, abgeleiteter Typ) des Fixture-Modells, bis zur dritten Stufe, und alle Navigationen, die nur der abgeleitete Typ trägt | Die Prüfung sucht auch in den Subtypen | andere |
| 7 | Der erste Verstoß von links entscheidet, gleich welcher Klasse; die Prüfung endet dort (`SPEC-263` Nr. 2, V3) | `TestPruefeExpandRang`: alle geordneten Paare (i ungleich j) aus {gültig, `navigation_unbekannt`, `keine_navigation`} im nicht offenen und aus {gültig, `navigation_nicht_unterstuetzt`, `keine_navigation`} im offenen Typ an den Stellen 0 bis 2 von Listen der Länge 2 und 3 (die vier ersten Paare der Zusatztabelle in §6 als benannte Unterfälle); Orakel: der erste Verstoß der Liste | Die Klassen werden nach Klasse (400 vor 501) statt nach Stelle entschieden; der letzte Verstoß entscheidet | andere |
| 8 | Programmfehler sind `model.ErrUngueltig` ohne Code und kein `ExpandFehler`: `o` gleich `nil`, leere Liste, leerer Name an jeder Stelle, der Name `*`, zwei Elemente mit bytegleichem Namen; `a` und `A` sind zwei Namen (`SPEC-263` Nr. 4, V7) | `TestPruefeExpandProgrammfehler`: Schleife über die Arten {`o` nil, `Elemente` nil, `Elemente` leer, leerer Name an Stelle 0 bis 2, `*` an Stelle 0 bis 2, Doppelter Name mit Abstand 0 bis 2 in gültigen und ungültigen Namen} und den Typ {nicht offen, offen}; `errors.Is(ErrUngueltig)`, `errors.As(*ExpandFehler)` falsch; Gegenprobe `a`, `A` bleibt gültig | Eine leere Liste ist gültig; `*` wird als Name gesucht; der doppelte Name wird nur bei gültigen Namen erkannt; Vergleich mit `EqualFold` | andere |
| 9 | Ein Programmfehler geht jedem Verstoß vor, auch einem früheren; er wird für den ganzen Knoten vor der Prüfung der Namen festgestellt (`SPEC-263` Nr. 4, fünfte Zeile der Zusatztabelle) | `TestPruefeExpandProgrammfehlerVor`: Schleife über die fünf Arten des Programmfehlers × {Verstoß davor, Verstoß danach} × die drei Klassen des Verstoßes (`navigation_unbekannt`, `keine_navigation`, `navigation_nicht_unterstuetzt`) | Die Programmfehler werden in der Schleife über die Namen erkannt (ein früherer Verstoß entscheidet) | andere |
| 10 | Jeder Verstoß ist genau ein `*ExpandFehler` mit genau einer Ursache, Klasse, Code und Status nach der Tabelle von Nr. 3: `errors.Is` trifft genau die Sentinel der Klasse, `Code()` ist `ODR-E6005` bzw. `ODR-E6001`, `Error()` hat den Wortlaut der Spezifikation (`SPEC-263` Nr. 3, V2) | `TestExpandFehler`: Schleife über die drei Ursachen, aus `PruefeExpand` und von Hand gebaut; Orakel: die Tabelle der Spezifikation; `errors.Is` gegen beide Sentinel und `ErrUngueltig`, `errors.As`, Code, Text | Eine Ursache mit der Klasse der anderen; die Codes vertauscht; `Error()` ohne die Ursache | andere |
| 11 | Eine Ursache außerhalb der drei (nur ein von Hand gebauter `ExpandFehler`) hat die Klasse der ersten Zeile: `ErrURLUngueltig`, `ODR-E6005`, `ODR-E6005: Expand ungültig: <Ursache>` (`SPEC-263` Nr. 3) | `TestExpandFehlerFremdeUrsache`: Ursachen {leer, `x`, eine der drei in anderer Schreibweise}; `errors.Is` auf beide Sentinel, Code, Text | Der Standardzweig wählt die Klasse `nicht_unterstuetzt` | andere |
| 12 | Kein Teil der Eingabe steht im Fehler: weder der Name noch die Stelle in der Liste noch der Name des Typs (`SPEC-263` Nr. 3) | `TestPruefeExpandFehlertext`: Schleife über alle Verstöße (die drei Ursachen) × Stelle 0 bis 2 mit unverwechselbarem Namen aus Bytewerten, der in keinem Typ des Modells vorkommt; `Error()`, `%v` und `%+v` enthalten weder Namen noch Stelle noch Typnamen | Der Name steht in der Meldung; die Stelle steht in der Meldung | andere |
| 13 | Eine Navigation ohne `NavigationBinding` am Set, auch mit leerem Ziel, ist gültig: die Funktion kennt Sets und Bindings nicht, die Ausführung entscheidet (`SPEC-263` Nr. 5, V5) | `TestPruefeExpandOhneBinding`: Schleife über Navigationen mit leerem `Target`, mit `Target` auf einen Typ, den das Modell nicht führt, und als Collection; Ergebnis `nil` | Die Prüfung verlangt ein nichtleeres Ziel | andere |
| 14 | Die Funktion ändert `typ` und `o` nicht (`SPEC-263` Nr. 1) | `TestPruefeExpandOhneAenderung`: tiefe Kopie von `typ` und `o` vor dem Aufruf, Vergleich danach; Schleife über alle Fälle der Zeilen 1 bis 9 (gültig und Verstoß) | Die Listen des Typs oder des Knotens werden sortiert | andere |
| 15 | Gleiche Eingabe ergibt gleiches Ergebnis; die Funktion hat keine eigene Grenze: Listen bis 32 Elemente (Grenze des Parsers) sind gültig (`SPEC-263` Nr. 6) | `TestPruefeExpandWiederholung`: jede Eingabe der Zeilen 1 bis 9 dreimal in Folge, tiefe Gleichheit; Listen der Länge 1 bis 32 aus gültigen Navigationen (Typ mit 32 Navigationen) und mit einem Verstoß an Stelle 0, 15 und 31 | Ein Zähler im Paket färbt das zweite Ergebnis; eine Obergrenze von 16 Elementen | andere |
| 16 | Gleichzeitige Aufrufe auf einem Typ und einem Knoten beeinflussen sich nicht; die Funktion ändert keine Variable des Pakets (`SPEC-263` Nr. 6) | `TestPruefeExpandGleichzeitig`: 8 Goroutinen × 50 Runden × alle Eingaben der Zeilen 1 bis 12 auf einem gemeinsamen `typ` und `o`, gleichzeitiger Start über eine Schranke; Ergebnis je Aufruf gegen den Einzelaufruf, verglichen erst nach dem Ende aller Goroutinen (Wirkungs-Fänger, nicht der Detektor); `make test-race` als zweiter Fänger je Prozess | Eine Paketvariable hält das Ergebnis der letzten Suche (ohne Sperre) | nebenläufig |
| 17 | `model` importiert keinen Port, keinen Dienst und keinen Adapter: die Funktion liest nur ihre zwei Argumente, ohne `ctx` und ohne Zugriff auf Daten (`SPEC-263` Kopf, `LH-FA-16.a`) | `make a-check` (Regeln des Hexagons, Ausgabe grün); die Signatur trägt kein `ctx` und kein Interface (Zeile 18) | Ein Import von `ports/outbound` im Paket `model` macht `make a-check` rot | andere |
| 18 | Validator-Contract-Test: die Funktion antwortet auf die Fälle der Spezifikation, wie sie sie nennt (`LH-FA-16.a`, `SPEC-263` Nr. 2 a bis d, 3, 4) | `TestPruefeExpandContract`: Tabelle der Fälle als Fixture (Typ, Liste, Ergebnis oder Ursache), ein Fall je Zeile, aus dem Wortlaut von Nr. 2 a bis d, 3 und 4 (die Spezifikation führt keine Beispieltabelle, §6); keine Eingabe aus dem Generator; Signatur über `reflect` gegen `PruefeExpand(EntityType, *ExpandOption) error` | Die Ursache einer Zeile wird vertauscht; die Signatur nimmt einen `ctx` | andere |
| 19 | Akzeptiertes Negativ: die Abbildung von `ExpandFehler` auf Code und Status ändert sich nicht und wird in diesem Slice nicht geprüft (`SPEC-109` Nr. 5, `SPEC-263` Nr. 3) | kein Fänger hier, der die Abbildung fährt; Grund: `model` darf `ports/outbound` nicht importieren; die Kette Fehler zu Sentinel trägt Zeile 10, die Zeilen von `abbildung_test.go` (Sentinel zu Code) bleiben unverändert grün; Folge: eine Änderung der Abbildung, die einen `ExpandFehler` anders behandelt, wird erst im Test des Aufrufers rot (`slice-expand-ausfuehrung`, dessen DoD trägt die Abbildung); Maß: 1 von 3 Gliedern der Kette (Fehler, Sentinel, Code) liegt außerhalb dieses Slice | — | andere |
| 20 | Akzeptiertes Negativ: die Namen in `typ.Properties` und `typ.NavigationProperties` sind eindeutig; die Funktion prüft es nicht (`SPEC-263` *Ungeprüft*) | kein Fänger; Grund: der Metadatenleser liefert jeden Namen einmal (`SPEC-125` Nr. 8 und 9 c); Folge: steht ein Name in beiden Listen, gilt Nr. 2 a (Navigation zuerst) und der Name ist gültig; Maß: 1 Fall von 1 (Name in beiden Listen) ohne Fänger | — | andere |
| 21 | Akzeptiertes Negativ: ein Datenrennen, das kein Lauf auslöst (`SPEC-263` *Ungeprüft*, `SPEC-122` Nr. 8 und 9) | kein Fänger über Zeile 16 hinaus; Grund: die Funktion hat heute keine Variable des Pakets; Folge: eine später eingeführte Paketvariable, die nur ein Zweig schreibt, den Zeile 16 nicht trifft, bleibt vom Detektor ungemeldet; Maß: Zeile 16 führt die vier Fälle (a) bis (d) von Nr. 2 und die fünf Arten des Programmfehlers von Nr. 4 je mindestens einmal, 0 von 9 Zweigen der Prüfung ohne Lauf | — | andere |

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

- Die Menge der Fälle ist die Menge der Navigationen und Properties aller Typen eines Modells (unbeschränkt) — **Ausgang:** offen; Schleife über die Menge des Fixture-Modells mit Subtypen, Ketten und offenem Typ, Prüfgrenze und Folge als akzeptiertes Negativ (`AGENTS.md` §3.17).
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
