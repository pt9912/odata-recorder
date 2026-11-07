# Slice slice-mock-zuordnung: MockUseCase, Teil Port, Zuordnung und `$metadata` (Teil 1 des Schnitts)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Teil 1 des Schnitts:** Der Plan `slice-mock-lesen` war zu groß (§4). Der Architect hat ihn vor dem Code geschnitten (Rückgabe R1), der Planner hat den Schnitt vollzogen: dieser Slice trägt Port, Konfig, Zuordnung und `$metadata`; [slice-mock-lesen](../open/slice-mock-lesen.md) (Teil 2) trägt das Lesen aus dem Entity-State mit der Coverage-Prüfung und braucht diesen Slice in `done/`.

**Bezug:** `LH-FA-06`, `LH-FA-09`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `LH-FA-06.a` Nr. 5 · `LH-FA-04.c` (Adressierung) · `SPEC-195` (Vertrag des Mock-Dienstes, vom Architect geschrieben; Nr. 1 bis 9, 12 ganz, Nr. 10 und 11 mit den Stellen S1 bis S4, S12 und S13) · `SPEC-198` Nr. 4 und 5 (`$metadata`, `HEAD`) · `SPEC-160` Nr. 1 und 9 (berichtigt) · `SPEC-175` Nr. 3 und 7 · `SPEC-174` Nr. 3 und 4 · `SPEC-109` Nr. 5 und 6 · `SPEC-070` · `SPEC-168` Nr. 10 (Nachzug, R2) · `ARC-002`, `ARC-003` · `architecture.md` §4.4, §2.4 (Zustand *Mock*).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-mock-zuordnung)

**Autor:** pt9912. **Datum:** 2026-10-30 (Plan `slice-mock-lesen`); geschnitten 2026-11-07 (Planner).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `MockUseCase` (Inbound Port in `src/hexagon/ports/inbound/`, Dienst `mock.Mocker` in `src/hexagon/services/mock`) ordnet einen Request Service und Workspace zu und liefert `$metadata` aus den gespeicherten Metadaten des Service; Collection, Entity und Anzahl enden bis Teil 2 mit `ODR-E6001` (Zeile Z18). Der Slice liefert den Port, auf dem `slice-mock-schreiben` und `HybridUseCase` aufbauen, die Konfig mit `Codecs`, die Rangfolge der Schritte, die Fehlerbildung als `*inbound.Fehler` und die Fakes der Ports, die Teil 2 erweitert.

**Übernimmt (Sendung von `slice-v2-bootstrap`, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted`):** die Konfiguration von `MockUseCase` nimmt je Dienst `Codecs *outbound.CodecSatz` (URL-Leser und Schreiber) entgegen; `nil` bedeutet die V4-Implementierungen der Konfig, und der Dienst verzweigt nie nach der Version (`SPEC-175` Nr. 3; ein Feld des Satzes, das der Dienst nicht braucht, darf `nil` sein). Kein Ausschluss dieses Slice trifft sie.

**Übernimmt (Erwartung von `slice-hybrid-service`, dessen §4 *Code-Abhängigkeit, Port*; Geber: `slice-hybrid-service`; Eintrag beim Start, Planner, 2026-11-07):** `MockUseCase` liegt als Inbound Port in `src/hexagon/ports/inbound/` und trägt den Lesepfad so, dass `slice-mock-schreiben` ihn erweitert und `HybridUseCase` ihn ruft; der Fehler `ODR-E6001` (nicht unterstützte Abfrage) kommt als Fehler des Ports an, nach dem der Aufrufer entscheidet; `ODR-E6003` (Coverage) liefert Teil 2. Kein Ausschluss dieses Slice trifft den Punkt.

**Übernimmt (Sendung von `slice-odata-url-pfad`, §6 und §7 dort, `SPEC-160`; Geber: `slice-odata-url-pfad`; Eintrag beim Start, Planner, 2026-11-07):** (a) der Pfad kommt relativ zum Service-Root beim Parser an, das Abschneiden des Präfixes ist Sache des Dienstes (`SPEC-160` Nr. 1; `/` für den Root, auch ohne abschließenden `/`); (c) ein eigener Fake von `URLCodec` in den Tests des Dienstes (a-check verbietet den Import des Adapters); (e) für `/` liefert der Parser die Ursache `service_dokument` (`501`), die der Dienst abbildet, solange keine Mindestfassung entschieden ist (deckt sich mit dem Ausschluss des Service-Dokuments unten); dazu der Fall von (b) für `$metadata` (`/$metadata?$top=1`, Z15). Die Punkte (b) für die übrigen Arten und (d) sowie der Verdrahtungs-Fall des Parsers tragen Teil 2 (`slice-mock-lesen`, §1 dort). Der Planner hat (c) beim Schnitt hierher gelegt: der Fake von `URLCodec` steht in den Zeilen Z05, Z08 und Z09 dieses Slice. Kein Ausschluss dieses Slice trifft diese Punkte.

**Übernimmt (Teil der Sendung von `slice-odata-json-schreiben`, `SPEC-168` Nr. 7, `SPEC-175` Nr. 7):** das Kennzeichen und den `Content-Type` der Ausgabe von `$metadata` setzt der Dienst unverändert aus der `Darstellung` eines verworfenen Aufrufs von `SchreibeAnzahl` (`SPEC-198` Nr. 4, Z14) und kennt den Namen des Kennzeichens nicht; `Wurzel`, `Accept` und die `Darstellung` von Collection, Entity und Anzahl trägt Teil 2. Kein Ausschluss dieses Slice trifft den Punkt.

**Übernimmt (Rückgabe R2 des Architect, `SPEC-168` Nr. 10; Eintrag beim Schnitt, Planner, 2026-11-07):** der letzte Satz von `SPEC-168` Nr. 10 („Der Vertrag von `MockUseCase` selbst steht noch aus und übernimmt diese Nummer.“) ist mit `SPEC-195` überholt; sein Nachzug (Verweis auf `SPEC-195` Nr. 1 und 6 bis 8) gehört in den Merge dieses Slice, denn mit ihm gelangt `SPEC-195` nach `main`. Kein Ausschluss dieses Slice trifft den Punkt; der Wortlaut anderer Nummern von `SPEC-168`, `SPEC-169` und `SPEC-172` bleibt unberührt (Nachbarn, Auftrag 6 in §3).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Collection, Einzel-Lookup und `/$count` (Coverage, Entity-State, Interpreter, Pflichten des Aufrufers, `Wurzel`, `Accept`, `ODR-E6003`, `ODR-E6008`) — `slice-mock-lesen` (Teil 2); der Dienst liefert bis dahin `ODR-E6001` (Z18), und Teil 2 nimmt den Punkt in §1 und DoD an.
- Mutationen — `slice-mock-schreiben`; Konstruktor, Routing und Zuordnung sind gemeinsam, die Zustandsänderung nicht (`SPEC-195` Nr. 7).
- Fallback auf Replay oder Upstream — `welle-hybrid`; `MockUseCase` kennt weder `ReplayUseCase` noch `UpstreamClient`.
- Spans und Metriken (`SPEC-029`) — `welle-betrieb`; es gibt noch keinen Port `Telemetry`.
- Das Service-Dokument (`GET` auf den Service-Root) — nicht im Muss-Umfang von `welle-mock` (Entscheidung des Users, 2026-10-30); kein Vertrag in der Spezifikation, es bleibt `501`. **Beleg des Architect (2026-11-07):** der M3-Test fordert es nicht an (§6); damit entfällt die Mindestfassung und die Rückgabe an den Planner.
- Die Version in der Fehlerantwort des Mock (`OData-Version`, `DataServiceVersion`) — `slice-bootstrap-m3` (Rückgabe R3 des Architect, §6); der Dienst liefert den Fehlerwert, das Fehlerobjekt schreibt der Adapter (`SPEC-195` Nr. 10), und `SchreibeFehler` ruft der Dienst nie (Z24).
- Fehlercodes im Katalog des Binary — der Slice bringt keinen neuen Code (`ODR-E6008` gehört Teil 2); die Codes, die er auslöst, sind angelegt.

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

- [ ] Port und Konfig: `MockUseCase` mit der einen Methode `Mock` in `src/hexagon/ports/inbound/`, `mock.New` mit `Codecs` (`nil` gilt als Satz der Konfig), Rangfolge der Schritte, Fehlerbildung als `*inbound.Fehler` nach `SPEC-109` Nr. 5, Kontext und Fakes der Ports; die Sendungen von `slice-v2-bootstrap`, `slice-hybrid-service` und `slice-odata-url-pfad` (§1: Präfix, Fake von `URLCodec`, `service_dokument` als `501`).
- [ ] Zuordnung und `$metadata`: Service und Workspace aus dem Request (`ODR-E5005`, `ODR-E5003`, `ODR-E7001`, `ODR-E5004`), `$metadata` aus den gespeicherten Metadaten byte-gleich mit dem Kennzeichen der `Darstellung` des Schreibers, `HEAD` gleich `GET` ohne Body; ein EntitySet ohne Instanzen bleibt im Modell, weil der Blob unverändert ausgeliefert wird (`LH-FA-06`, Boundary).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist; dazu der Nachzug von `SPEC-168` Nr. 10 (R2, §1).
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Modus `mock` erreicht es erst mit `slice-bootstrap-m3`); dieser Slice bringt keinen Code.
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
| `src/hexagon/ports/inbound/` | update | `mock.go`: `MockUseCase` (Z23) |
| `src/hexagon/services/mock/` | neu | Konfig und `New`, Zuordnung, URL, Methode, Metadaten, Fehlerbildung, Platzhalter für Collection, Entity und Anzahl, Antwort von `$metadata` (`SPEC-195`, `SPEC-198` Nr. 4 und 5) |
| `spec/spezifikation.md`, `spec/architecture.md` | update (Architect, getan; Planner: Nachzug R2) | `SPEC-195` bis `SPEC-199`, `SPEC-109`, `SPEC-160` Nr. 9, `SPEC-230` Nr. 8, Historie; Reihenfolge der Lesevorgänge in §4.4; `SPEC-168` Nr. 10 |
| `docs/plan/planning/open/slice-mock-schreiben.md` | update (Architect, getan) | §1 *Übernimmt* (Erwartung von diesem Slice und von Teil 2, `AGENTS.md` §3.13) |
| Tests | neu | die Fänger der Tabelle (`TestNewKonfig` bis `TestKeinFehlerobjekt`) mit Fakes der acht Ports (Metadaten, Workspaces, Blobs, `URLCodec`, Schreiber mit Verhalten; Entities, Coverage, `MetadatenCodec` als Zähl-Fakes, die Teil 2 erweitert); der Dienst importiert keinen Upstream-Typ und kein Paket eines anderen Service |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Die Fakes der Ports (Repositories, Codec) halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15).

**Entscheidungen des Architect** (2026-11-07; die Gründe und die Alternativen stehen in der Spezifikation, nicht hier; Aufträge und Entscheidungen des ungeteilten Plans stehen in [slice-mock-lesen](../open/slice-mock-lesen.md) §3, soweit sie Teil 2 betreffen):

1. **Schnitt:** in zwei Teile, Zeilen Z01 bis Z24 hier, Zeilen L01 bis L45 in Teil 2 (Rückgabe R1, vom Planner vollzogen).
2. **Fehlerwert, nicht Fehlerantwort:** der Dienst liefert `*inbound.Fehler`, das Fehlerobjekt schreibt der Adapter (`SPEC-195` Nr. 10); die Sendung „Fehlerobjekt mit `Content-Type` und Kennzeichen“ aus `slice-odata-json-schreiben` gilt damit nicht für den Mock, sondern bleibt eine Erwartung an den Adapter (Rückgabe R3, an `slice-bootstrap-m3`).
3. **`$metadata`** ist der Strom des Blobs mit `application/xml`; das Kennzeichen der Version kommt aus der `Darstellung` eines verworfenen Aufrufs von `SchreibeAnzahl` (`SPEC-198` Nr. 4).
4. **Service-Dokument:** Beleg und Ergebnis in §6; es bleibt `501`.
5. **Kennungen und Dateien der Nachbarn** (Linie B neben `slice-v4-schreiber-marker-text`): `SPEC-195` bis `SPEC-199` sind genommen, `SPEC-192` bis `SPEC-194` bleiben frei; kein `ARC`, kein `ADR`; dieser Slice ändert weder `codec_satz.go` noch den Wortlaut von `SPEC-168`, `SPEC-169` oder `SPEC-172` (außer dem Nachzug R2 in `SPEC-168` Nr. 10 beim Merge) und schreibt in `src/adapters/driven/odata/` keine Datei; er schreibt in `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/`, die Tests und `spec/spezifikation.md` (neue Nummern ab `SPEC-195`). Findet er in den Nachbarn eine Lücke, geht sie als Rückgabe an den Planner. Die Merge-Reihenfolge entscheidet der erste Verifier-Pass; der zweite Zweig nimmt `main` per Merge auf.
6. **Kein Code im Katalog:** der Slice legt keinen Fehlercode an; die 11 Codes, die er auslöst, sind in `src/hexagon/model/fehler.go` angelegt (`ODR-E6003` und `ODR-E6008` kommen mit Teil 2). Berührt eine Änderung dennoch `harness/katalog-ausnahmen.txt`, `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, `.a-check.yml` oder `.golangci.yml`, geht das als Rückgabe an den Planner.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3 a; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz *Mutationen und Raten* darunter ein, `AGENTS.md` §3.10). Die Zeilen Z01 bis Z24 stammen vom Architect; der Planner hat beim Schnitt Z11, Z18, Z19, Z20 und Z24 auf das beschränkt, was dieser Slice erreicht, und die Reste als L43 bis L45 in Teil 2 gelegt. Jede Zusage und jede Sendung hat eine Zeile; die Sendungen sind in §6 den Zeilen zugeordnet. Die Mengen M1 bis M6 stehen unter der Tabelle. Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (Z14) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z01 | `New` weist ab: kein Dienst, leere Kennung, leerer Root, nil in einem der acht Felder, leeres `PublicBaseURL`, `MaxBytes` unter 1, Satz mit nil in `Metadaten`, `URL` oder `Schreiber`; `model.ErrUngueltig`, kein Port gerufen, Text ohne Wert | `TestNewKonfig`: Schleife über M1, alle Fakes zählen 0 Aufrufe; die Zahl der Interface-Felder der Konfig per `reflect` gegen die acht der Menge (ein neues Feld macht den Zähler rot) | Prüfung von `Coverage` fehlt | andere |
| Z02 | `Codecs` nil gilt als Satz der Konfig und nicht nur lesend; ein Satz ersetzt `Metadaten`, `URL` und `Schreiber`, `Daten` und `Eingabe` dürfen nil sein; der Dienst liest `Codecs.Version` nie und kennt keinen Namen eines Kennzeichens | `TestSatzWahl`: Dienst mit und ohne Satz, je Fake ein Zähler (der Fake des Satzes gerufen, der der Konfig 0, und umgekehrt); `TestQuelltextOhneVersion`: im Quelltext des Pakets ohne Tests kommen `Version`, `OData-Version` und `DataServiceVersion` nicht vor | Satz ignoriert, Fake der Konfig gerufen; `Codecs.Version` verzweigt | andere |
| Z03 | Argumente: `ctx` nil, `Methode` leer, `Laenge` -2, `Body` nil bei `Laenge` 1 sind `ErrUngueltig` ohne Code, kein Port, `Body` nicht gelesen; `Laenge` -1 und 0 sind gültig | `TestArgumente`: Schleife über die vier Fälle und die zwei gültigen Grenzen; Zähler aller Fakes 0, Lesezähler des Bodys 0 | `Laenge` -1 als ungültig | andere |
| Z04 | Dienstwahl: Segmentgrenze, längster Root, byteweise, die Schreibweise zählt; kein Dienst, leerer Pfad, Pfad ohne `/`: `ODR-E5005`/404 ohne Port | `TestDienstWahl`: M2 (5 Roots × 12 Pfade) gegen eine Tabelle im Test; Zähler 0 | Segmentgrenze nicht geprüft (`/ab` trifft `/a`); kürzester statt längster Root | andere |
| Z05 | OData-Pfad: Rest nach dem Root, `/` bei leerem Rest, beim Root `/` der Pfad; `ParseURL` genau einmal mit (OData-Pfad, `RawQuery`) unverändert (Sendung (a) von `slice-odata-url-pfad`) | `TestODataPfad`: M2; der Fake von `URLCodec` hält beide Argumente fest | Präfix nicht abgeschnitten; `/` bei leerem Rest fehlt; `RawQuery` verändert | andere |
| Z06 | Zuordnung: genau ein nichtleerer Wert von `X-Recorder-Workspace`, der Name ohne Beachtung der ASCII-Schreibweise, der Wert byte-gleich an `Workspaces.Get`; sonst `ODR-E5003`/400 ohne Port (fehlt, leere Liste, leerer Wert, zwei gleiche, zwei verschiedene, zwei Schreibweisen des Namens) | `TestZuordnung`: M3 (10 Formen × 6 Werte); Fake hält den Wert fest | Wert gekürzt; zwei gleiche Werte angenommen; Name nur klein verglichen | andere |
| Z07 | `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session` ändern nichts (fehlend, vorhanden, mehrfach, leer) | `TestSessionHeaderOhneWirkung`: Schleife über vier Formen, Antwort und Zähler gleich dem Lauf ohne | Dienst verlangt `X-Recorder-Replay-Session` | andere |
| Z08 | Fehler des URL-Lesers: Klasse `ungueltig` ist `ODR-E6005`/400, `nicht_unterstuetzt` `ODR-E6001`/501, auch der Pfad `/` (`service_dokument`, Sendung (e)); die Ursache über `Unwrap`, `Error()` ohne Wert | `TestURLUrsachen`: Schleife über `model.AlleURLUrsachen()` (16), Orakel eine Tabelle im Test, Zählprüfung gegen die Menge; Canary im `Ort`; Zähler `Workspaces` 0 | Klasse einer Ursache vertauscht (`literal_typ` als `ODR-E6005`) | andere |
| Z09 | `*KompatFehler` jeder Ursache ist `ODR-E6006`/501; `model.ErrUngueltig` ohne `URLFehler` und jeder andere Fehler des Lesers `ODR-E9001`/500; eine `Ressource.Art` außerhalb der vier `ODR-E9001` | `TestKompatUndRest`: Schleife über `model.AlleKompatUrsachen()` (6) und drei Fälle (Art 0, Art 99, `errors.New`) | `KompatFehler` als `ODR-E6001` abgebildet | andere |
| Z10 | Methode: `GET` und `HEAD` laufen weiter; jede andere ist `ODR-E6002`/501, bei `NurLesend` `ODR-E6006`; nach dem Lesen der URL und vor dem Workspace (Zähler `Workspaces` 0), der Body wird nicht gelesen | `TestMethode`: Methode 8 (`POST`, `PATCH`, `PUT`, `DELETE`, `OPTIONS`, `get`, `Get`, `FOO`) × `NurLesend` 2, dazu `GET` und `HEAD` | `NurLesend` nach `Workspaces.Get` geprüft; `get` angenommen | andere |
| Z11 | Rangfolge (`SPEC-195` Nr. 3) der acht Stufen dieses Slice: der erste Verstoß entscheidet, die Ports dahinter bleiben ungerufen; die Paare mit „Set unbekannt“, „Coverage im strikten Modus“ und „Fehler des Interpreters“ trägt L44 in Teil 2 | `TestRangfolge`: M4 (8 Stufen, 27 baubare Paare), der Fake löst beide Verstöße aus; Soll der Code der früheren Stufe, Zähler der späteren Ports 0 | Zuordnung nach dem Lesen der URL; `methode` nach dem Workspace | andere |
| Z12 | Workspace: `Workspaces.Get` genau einmal mit dem Wert unverändert, am Port nur `Get` (der Fake zählt `Create`, `List`, `Advance`, `Delete` mit 0); `ErrNichtGefunden` ist `ODR-E7001`/404, ein fremder Service `ODR-E5004`/409 ohne Ursache, ein anderer Fehler nach `SPEC-109` Nr. 5 | `TestWorkspace`: Schleife über Fehler (`ErrNichtGefunden`, `ErrBeschaedigt` als `ODR-E3001`, anderer als `ODR-E9001`) und eigen/fremd; Werte mit Rand-Leerzeichen und Großbuchstaben | Vergleich von `ServiceID` fehlt; `ErrNichtGefunden` als `ODR-E9001` | andere |
| Z13 | Service: `Metadaten.Get(ctx, Dienst.ID)` genau einmal, am Port nur `Get`; `ErrNichtGefunden` ist `ODR-E5005`, ein leerer `MetadataBlob` `ODR-E6004`/422 ohne Ursache, ein anderer Fehler nach `SPEC-109` Nr. 5 | `TestService`: Schleife über vier Zustände (Eintrag, leerer Blob, `ErrNichtGefunden`, anderer Fehler) | leerer Blob liefert `200` mit leerem Body | andere |
| Z14 | `$metadata`: `200`, Body byte-gleich dem Strom von `Blobs.Open(ctx, MetadataBlob)` (Inhalt kein UTF-8, mit BOM, 1 MiB; aus Bytewerten gebaut), `Content-Type: application/xml`, das Kennzeichen unverändert aus der `Darstellung` von `SchreibeAnzahl(ctx, dst, 0)` (der Dienst kennt keinen Namen), `Content-Type` und Body der Probe nie in der Antwort, kein Aufruf von `MetadatenCodec` | `TestMetadaten`: Schleife über drei Inhalte × zwei Sätze; der Fake des Schreibers liefert `X-Fake-Kennzeichen` und `text/plain`; Zähler `ParseMetadata` 0, `SchreibeAnzahl` 1 mit `n` 0 | Kennzeichen fest `OData-Version`; Body durch den Puffer der Probe ersetzt; Body gekürzt | andere |
| Z15 | Optionen ohne Wirkung an `$metadata` (`$filter`, `$orderby`, `$top`, `$skip`, `$count`, `$select`, `$format`; Sendung (b) für diese Art) | `TestMetadatenOptionen`: Schleife über sieben Optionen aus dem Fake des Parsers, Antwort gleich der ohne Option | `$top=0` leert den Body | andere |
| Z16 | Fehler von `$metadata`: `ErrBlobFehlt` und `ErrBlobBeschaedigt` sind `ODR-E3001`/500, ein anderer Fehler `ODR-E9001`; bei einem Fehler der Probe wird der Strom genau einmal geschlossen, bei Erfolg schließt ihn der Adapter | `TestMetadatenFehler`: Schleife über drei Fehler von `Open` und zwei der Probe (`ErrUngueltig`, sonstiger); der Fake des Stroms zählt `Close` | Strom bei Fehler der Probe nicht geschlossen; zweimal geschlossen | andere |
| Z17 | `HEAD` auf `$metadata`: Status und Header gleich `GET`, Body ein leerer Strom (nicht nil), der geöffnete Strom geschlossen, Fehler gleich | `TestHeadMetadaten`: Schleife über `GET`/`HEAD` × Erfolg/Fehler | `HEAD` liefert den Body | andere |
| Z18 | Platzhalter bis Teil 2: Collection, Entity und Anzahl enden mit `ODR-E6001`/501, ohne Port nach Schritt (g); kein Vertragsinhalt, Teil 2 ersetzt die Zeile | `TestPlatzhalterLesen`: Schleife über drei Arten, Zähler von `Entities` und `Coverage` 0 | der Platzhalter ruft `Entities` | andere |
| Z19 | Jeder Fehler mit Code ist `*inbound.Fehler{Code, Status, Ursache}` mit dem einen Status des Codes, die Ursache über `Unwrap`, `Error()` ohne Wert (Canary in Pfad, Header, Kennung, Fehlertext des Ports), `Antwort` der Nullwert; ohne Code bleiben `ErrUngueltig` und die Kontextfehler; dieser Slice löst 11 der 13 Codes von `SPEC-195` Nr. 10 aus (ohne `ODR-E6003` und `ODR-E6008`, die Teil 2 auslöst) | `TestFehlerWert`: M5 (11 Codes, je ein Auslöser; Zählprüfung gegen die Liste in `SPEC-195` Nr. 10 ohne die beiden Codes von Teil 2), `Code.Stati()`, Canary | Status 0; `Error()` nennt den Pfad | andere |
| Z20 | Kontext: jeder Aufruf eines Ports mit `ctx` erhält das `ctx` des Aufrufs unverändert; ein Kontextfehler (`Canceled`, `DeadlineExceeded`, verpackt, in `errors.Join` neben einem Fehler mit Code) kehrt ohne Code unverändert zurück, danach kein weiterer Port, der Strom geschlossen; dieser Slice trägt die Stellen S2, S3, S4 und S12 (S5 bis S11 trägt L43 in Teil 2) | `TestKontext`: M6 (Form 4 × Stelle 4 = 16); der Fake hält `ctx` fest (Wert im Kontext) | Kontextfehler nur mit `==` erkannt; die Probe ruft `context.Background()` | andere |
| Z21 | Die Fakes halten jedes Argument fest und zählen jede Operation ihres Ports, auch die, die der Dienst nie rufen darf | `TestFakesZaehlenAlles`: je Fake eine Schleife über alle Methoden des Interfaces (`reflect`), Zähler 1 und Argument festgehalten; Zählprüfung gegen `NumMethod`; die Schleife läuft über alle acht Fakes, auch die drei Zähl-Fakes, die Teil 2 erweitert | Fake zählt `List` nicht | andere |
| Z22 | Der Dienst hält keinen Zustand, keine Sperre, keine Goroutine und keinen Cache; gleichzeitige Aufrufe (64 Goroutinen × 8 Requests, verschiedene Workspaces) geben je Request die Antwort des Einzelaufrufs | `TestPaketOhneZustand` (Quelltext, in jedem Lauf); `TestGleichzeitig` unter `make test-race`, je Prozess | Paketvariable mit Zähler; Cache des Modells im Dienst | nebenläufig |
| Z23 | Der Port `MockUseCase` in `ports/inbound/mock.go` hat die eine Methode `Mock`; `*Mocker` erfüllt ihn (der Hybrid-Dienst ruft ihn); er wächst nur additiv | Übersetzung (`var _ inbound.MockUseCase = (*Mocker)(nil)`); `TestMockPort`: Methodenmenge per `reflect` gleich {`Mock`} | Methode umbenannt | andere |
| Z24 | Der Dienst ruft `SchreibeFehler` nie: die Fehlerantwort schreibt der Adapter (`SPEC-195` Nr. 10); die Fehlerzeilen des Lesens trägt L45 in Teil 2 | `TestKeinFehlerobjekt`: über alle Fehlerzeilen Z08 bis Z16 der Zähler `SchreibeFehler` 0 | der Dienst schreibt das Fehlerobjekt selbst | andere |

**Mengen** der Fänger (`implement-slice` Schritt 14 und 19, `AGENTS.md` §3.17): der Architect nennt vor dem Code Menge, Achsen und Zahl der Elemente je Achse; die Fänger laufen als Schleifen über alle Elemente, mit Orakel (nie der Dienst selbst) und Zählprüfung. Die Mengen des Modells, die dieser Slice zählt, sind `model.AlleURLUrsachen()` (16), `AlleKompatUrsachen()` (6) und `AlleCodes()` (37, mit der Zählprüfung gegen die elf Codes dieses Slice); `AlleRessourcenArten()` und `AlleCoverageStatus()` legt Teil 2 an. Der Bericht des Implementers nennt je Menge die Zahl, die der Test zählt, und je Achse eine Mutation an einem Element, das die Zeile nicht nennt, rot gesehen; die Spalte *Mutation vorab* nennt eine solche. Die Nummern M4 bis M6 gelten in beiden Plänen für dieselbe Menge: dieser Slice führt den Anteil, den er erreicht, Teil 2 den Rest. **Prüfgrenze der unbeschränkten Mengen** (akzeptiertes Negativ, `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze`): die Länge einer Kennung des Headers (M3 führt 1, 64 und 200 Bytes) und die Größe des `$metadata` (Z14 führt 1 MiB); darüber ist im Dienst nichts geprüft, Folge im Betrieb: eine Grenze, die der Dienst nicht kennt, gibt es nicht, und eine Regression auf großen Mengen bliebe bei grünen Gates unbemerkt (der Lasttest, `LH-QA-01`).

| Nr. | Menge | Achsen und Elemente | Fänger | Mutation vorab |
|---|---|---|---|---|
| M1 | Fehler der Konfiguration | Feld nil 8 (`Metadaten`, `Workspaces`, `Entities`, `Coverage`, `Blobs`, `MetadatenCodec`, `URLCodec`, `Schreiber`) + kein Dienst, leere Kennung, leerer Root, leeres `PublicBaseURL` je 1 + `MaxBytes` 0 und -1 + Satz mit nil in `Metadaten`, `URL`, `Schreiber` je 1 = 17 | Z01 | `MaxBytes` 0 angenommen |
| M2 | Root × Pfad | Root 5 (`/`, `/a`, `/a/b`, `/ab`, `/a%2Fb`) × Pfad 12 (gleich dem Root, Root und `/`, Root und `/x`, Root und `x`, Root in anderer Schreibweise, leer, ohne `/`, nur `/`, mit `%2F`, mit `//`, kürzer als der Root, Root als Präfix eines Segments) = 60 | Z04, Z05 | Pfad `/a/` mit Root `/a` ergibt `//` |
| M3 | Header der Zuordnung | Form 10 (fehlt, leere Liste, leerer Wert, ein Wert, zwei gleiche, zwei verschiedene, zwei Schreibweisen des Namens mit je einem Wert, Name klein, Name `x-RECORDER-workspace`, Wert mit Rand-Leerzeichen) × Wert 6 (1 Byte, 64 Bytes, 200 Bytes, mit Großbuchstaben, mit `/`, mit NUL) = 60 | Z06 | Name nur in der Schreibweise `X-Recorder-Workspace` gesucht |
| M4 | Stufen der Rangfolge | Stufe 8 dieses Slice (Argumente, kein Service, Zuordnung, URL, Methode, Workspace unbekannt, Workspace fremd, Service ohne Metadaten) → 28 Paare, davon 27 baubar („Workspace unbekannt“ und „fremd“ schließen sich aus); die 27 Paare mit den drei Stufen des Lesens (Set unbekannt, Coverage im strikten Modus, Fehler des Interpreters) trägt L44 | Z11 | Zuordnung nach dem Lesen der URL |
| M5 | Erreichbare Codes | 11 dieses Slice (`ODR-E3001`, `ODR-E5003`, `ODR-E5004`, `ODR-E5005`, `ODR-E6001`, `ODR-E6002`, `ODR-E6004`, `ODR-E6005`, `ODR-E6006`, `ODR-E7001`, `ODR-E9001`) von 13 (`SPEC-195` Nr. 10); `ODR-E6003` und `ODR-E6008` trägt L37 | Z19 | Code 12 (`ODR-E5001`) kommt vor |
| M6 | Kontextfehler | Form 4 × Stelle 4 dieses Slice (S2, S3, S4, S12) = 16 von 44 (S2 bis S12); S5 bis S11 trägt L43 | Z20 | `errors.Join` neben einem Code: der Code gewinnt |

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Schleifen aus dem Wortlaut der Zusage, nie eine Liste der Fälle, die einfielen (`AGENTS.md` §3.17).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-query-interpreter` und `slice-odata-json-schreiben` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Stand beim Start** (Planner, 2026-11-07): erfüllt, `slice-query-interpreter`, `slice-odata-json-schreiben`, `slice-handbuch-meldungskatalog`, `welle-record-replay` (mit M2) und `slice-odata-json-eingabe` liegen in `done/`; `slice-v2-bootstrap` ist nicht Voraussetzung (die Konfig nimmt `Codecs` mit `nil` entgegen). Der Slice liegt in der Linie B neben `slice-v4-schreiber-marker-text` (Entscheidung 5 in §3) und berührt keine Datei der Gate-Kette (Übergangsregeln Nr. 7; kein Slice der Linie C in `in-progress/`; `src/hexagon/services/**` ist in `.a-check.yml` als Schicht `services` schon geführt); ändert Vertrag oder Code eine Datei der Gate-Kette, geht das als Rückgabe an den Planner (§6).
- **Größenmarke** (`SPEC-128`; Faktoren allein dort; Stand des Planner nach dem Schnitt, 2026-11-07). Der Block unten trägt den Teil des Architect (`pfad`, `produktcode`, `fakes`, `schnitt-grund`). `basis` ist der letzte Commit des Architect, der Vertrag, Tabelle und Raster abschließt, und `schaetzung` und `marke` folgen der Formel; die drei Zeilen setzt der Planner im Zweig `slice-mock-zuordnung` in eigenem Commit vor dem ersten Code-Commit, weil der Commit des Architect nur dort liegt (auf `main` steht der Block ohne sie und `make groessenmarke` meldet *Marke steht aus*). Die Tabelle in §3 hat 24 Zeilen (1 nebenläufig: Z22; 23 andere); Port-Fakes 8 (`Workspaces`, `Metadaten`, `Blobs`, `URLCodec`, `Schreiber` mit Verhalten, `Entities`, `Coverage`, `MetadatenCodec` als Zähl-Fakes: der Architect zählte die letzten drei in Teil 2, aber Z01, Z18 und Z21 brauchen sie hier); Produktcode erwartet 740 (Port 30, Konfig und `New` 130, Dienstwahl und Pfad 90, Zuordnung 70, Fehlerwert und Abbildung 150, `$metadata` und Probe 130, Ablauf 140). Richtwert nach `SPEC-128` Nr. 1, das Werkzeug rechnet: *S* = 85·1 + 80·23 + 200·8 + 740 = 4265 und *M* = 5800 (die Zählung des Architect mit fünf Fakes ergäbe 3665 und 5000). Das Register führt die Überschätzung (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`, 13×); die Marke folgt trotzdem der Formel.
- **Schnittfrage** (Architect, 2026-11-07, Rückgabe R1; vollzogen vom Planner): der ungeteilte Plan hatte *S* = 8715 und ist geschnitten. Dieser Teil liegt mit *S* = 4265 weiter über der Schnittgrenze (*S* ≥ 2223) und wird nicht weiter geschnitten; der Grund steht in `schnitt-grund`. Der Schnitt, der sich anböte, trennte die Rangfolge, Zuordnung und Fehlerbildung (Z01 bis Z13, Z19 bis Z24) von `$metadata` (Z14 bis Z17); er scheitert am geteilten Mechanismus (der Ablauf bis Schritt (g), die Fehlerbildung nach `SPEC-109` Nr. 5, der Dienst, `Workspaces`, `Metadaten`, `Blobs` und der Schreiber als Fakes) und am Lieferwert: ohne `$metadata` antwortet der Dienst auf jede Anfrage mit einem Fehler und liefert dem Client nichts.

```groessenmarke
pfad: src/hexagon/services/mock
pfad: src/hexagon/ports/inbound/mock.go
pfad: src/hexagon/ports/inbound/mock_test.go
produktcode: 740
fakes: 8
schnitt-grund: Kein weiterer Schnitt von Teil 1 (Planner, 2026-11-07, nach dem Schnitt des Architect, Rückgabe R1): die Zuordnung (Dienstwahl, Header, Workspace, Service), die Fehlerbildung und die Rangfolge trennten sich von $metadata nur so, dass ein Slice nichts liefert, was ein Client abruft, denn der Dienst endet in jedem Zweig mit einem Fehler; beide Hälften teilen den Ablauf bis Schritt (g), die Fehlerbildung nach SPEC-109 Nr. 5, den Dienst und die Fakes von Workspaces, Metadaten, Blobs und Schreiber. Collection, Entity und Anzahl liegen in Teil 2, der diesen Teil braucht.
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Die Tabelle *Zusage · Fänger* liegt vor (§3), der Block steht auf `main` ohne die Zahlen des Planner (siehe oben).
- `in-progress` → `open` (blockiert — Carveout?): die Zuordnung von Service und Workspace im Modus `mock` (welche Header, welcher Fehler) sind nicht entschieden, oder der Referenz-Client des M3-Tests fordert das Service-Dokument an und dessen Mindestfassung steht nicht in der Spezifikation (siehe §6).

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

Stand beim Start (Architect, 2026-11-07; Planner, 2026-11-07): der Vertrag steht in `SPEC-195` bis `SPEC-198` und `SPEC-199`; Tabelle (§3) und Raster (unten) stehen vor dem ersten Code-Commit. Jeder Ausgang unten bleibt *offen* bis zur Closure, außer R1; was der Architect entschieden hat, steht mit Ort dabei. Die Risiken des Lesens (Lesesicht, Einzel-Lookup bei `partial`, Zweig- und Rückgaben-Menge, Verdrahtung des Parsers, Fehlercodes und Katalog) führt Teil 2 ([slice-mock-lesen](../open/slice-mock-lesen.md)).

- Zuordnung im Modus `mock` (Header fehlt, leer, mehrfach, gehört zu einem anderen Service): entschieden in `SPEC-195` Nr. 3 bis 5 und 8 (allein `X-Recorder-Workspace`, `ODR-E5003`, `ODR-E7001`, `ODR-E5004`); das Raster unten hat keine Zelle ohne Ort; Fänger Z04, Z06, Z07, Z11, Z12 — **Ausgang:** offen; eine Zelle ohne Ort, die der Implementer findet, geht an den Architect.
- Das Service-Dokument (`GET` auf den Service-Root) hat in der Spezifikation keinen Vertrag und gehört nicht zum Muss-Umfang (Entscheidung des Users, 2026-10-30). **Beleg des Architect (2026-11-07, vor dem ersten Code-Commit):** der M3-Test (`welle-mock` §3, Closure-Trigger; `slice-bootstrap-m3` §1 und DoD) startet den Recorder im Modus `mock`, legt einen Workspace an, fragt `$metadata` und eine kombinierte Abfrage (`$filter`, `$orderby`, `$top`, `$skip`, `$count`), mutiert, erzeugt einen Snapshot, stellt wieder her und setzt zurück; sein Client ist ein `net/http`-Client im selben Prozess wie Recorder und Referenzdienst (`slice-bootstrap-m3` §6), und kein Wort in `welle-mock`, `slice-bootstrap-m3` oder `slice-m2-referenzdienst` verlangt das Service-Dokument. Das Fehlen des Requests ist der Beleg: es bleibt bei `501` mit `ODR-E6001` (der Parser liefert die Ursache `service_dokument`, `SPEC-195` Nr. 6), ohne Mindestfassung, ohne Rückgabe an den Planner; Fänger Z08, *Ungeprüft* in `SPEC-195` — **Ausgang:** offen bis zur Closure; Adresse: Implementer (Z08).
- Schnittfrage (Auftrag 5 des ungeteilten Plans): *S* = 4265 liegt über 2223 — **Ausgang:** eingetreten für den ungeteilten Plan (geschnitten, R1); für diesen Teil entschieden: kein weiterer Schnitt, Grund in `schnitt-grund` (§4); eine Überschreitung der Marke zeigt `make groessenmarke`, der Eintrag mit Grund ist der Ausweg (`SPEC-128` Nr. 5 d).
- Die Sendungen: `Codecs` mit Satz Z02; `slice-hybrid-service` (Port) Z23; `slice-odata-url-pfad` (a) Z05, (c) Z08 und Z09 (Fake von `URLCodec`), (e) Z08, (b) für `$metadata` Z15; Kennzeichen der Ausgabe von `$metadata` Z14 — **Ausgang:** offen; Adresse: Implementer (Fänger).
- Zweig- und Rückgaben-Menge mit Zählprüfung (Auftrag 4 des ungeteilten Plans): die Stufen, Codes und Stellen, die dieser Slice erreicht, zählen M4, M5 und M6 gegen die Liste in `SPEC-195` Nr. 3, 10 und 11; die Mengen des Modells `AlleRessourcenArten()` und `AlleCoverageStatus()` und die Zweige unter dem Detektor (L36) gehören Teil 2, denn erst dort gibt es mehr als eine Art der Ressource; Z22 fährt `TestGleichzeitig` unter `make test-race` — **Ausgang:** offen; Adresse: Implementer (Test).
- Z19 nennt 11 statt 13 Codes (Planner, beim Schnitt): `ODR-E6003` und `ODR-E6008` löst erst das Lesen aus, `SPEC-195` Nr. 10 bleibt mit 13 Codes der Vertrag beider Teile, und L37 zählt alle 13 — **Ausgang:** offen; Adresse: Reviewer (Zählprüfung gegen die Liste in der Spezifikation).
- Linie B neben `slice-v4-schreiber-marker-text` (Entscheidung 5 in §3): genommen sind `SPEC-195` bis `SPEC-198` (Verträge) und `SPEC-199` (die Zeile des Codes `ODR-E6008`); `SPEC-192` bis `SPEC-194` bleiben frei. Weder `ARC` noch `ADR` nötig. `SPEC-160` Nr. 9 ist berichtigt (Wortlaut, Historie). Dateien der Nachbarn bleiben unberührt (`codec_satz.go`, `jsoncodec`, `jsoncodecv2`, `urlparser`) — **Ausgang:** offen; Adresse: Planner bei einer Berührung, Reihenfolge der Merges der erste Verifier-Pass.
- Gate-Kette (Stand beim Start in §4): geprüft vom Architect, 2026-11-07: `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel, `.a-check.yml` und `.golangci.yml` bleiben unberührt (`services` darf `model` und die Ports, der Dienst importiert keinen Adapter); ein Bedarf an einer Änderung der Gate-Kette — **Ausgang:** weiter offen (nicht zu erwarten); Adresse: Planner.
- `Codecs` und Satz: `nil` heißt die V4-Implementierungen der Konfig (`SPEC-175` Nr. 3, `SPEC-195` Nr. 1); wie die Konfig sie bildet, bleibt `slice-bootstrap-m3`; Fänger Z02 — **Ausgang:** offen; Raster, Zelle „`Codecs` gleich `nil`“ in Z02.
- R1, Schnitt (Rückgabe an den Planner, `AGENTS.md` §3.13): die Zeilen Z01 bis Z24 der Tabelle sind **Teil 1** (dieser Slice), die Zeilen L01 bis L42 **Teil 2**; der Planner legte diesen Slice an, behielt in `slice-mock-lesen` Teil 2 und teilte Block, Tabelle, Raster und DoD — **Ausgang:** eingetreten (geschnitten, Planner, 2026-11-07). Die Zeilen Z11, Z19, Z20 und Z24 reichen in Teil 2 hinein; ihre Reste stehen dort als L44, L37, L43 und L45.
- R2, `SPEC-168` Nr. 10 (Rückgabe an den Planner): der letzte Satz („Der Vertrag von `MockUseCase` selbst steht noch aus und übernimmt diese Nummer.“) ist mit `SPEC-195` überholt; der Nachzug (Verweis auf `SPEC-195` Nr. 1 und 6 bis 8) gehört in den Commit des Merge dieses Slice, denn mit ihm gelangt `SPEC-195` nach `main`; Teil 2 ändert `SPEC-168` nicht — **Ausgang:** offen; Adresse: Implementer, im Merge-Commit (DoD: Doku-Update).
- R3, Version der Fehlerantwort (Rückgabe an den Planner): die Fehlerantwort des Mock trägt das Fehlerobjekt der Form von `SPEC-136` Nr. 8 ohne `OData-Version` und ohne `DataServiceVersion`, auch für einen V2-Service (`SPEC-195` Nr. 10, Entscheidung); `SchreibeFehler` des Schreibers (`SPEC-168` Nr. 7, `SPEC-172` Nr. 6) hat damit in dieser Welle keinen Aufrufer. Ein Weg wäre ein additives Feld am Fehlerwert oder ein Aufruf des Schreibers im Composition Root; beides berührt den Adapter und `SPEC-136` — **Ausgang:** entfallen für diesen Slice (der Dienst ruft `SchreibeFehler` nie, Z24); die Entscheidung und Umsetzung liegen bei `slice-bootstrap-m3` (§1 dort, Übernimmt, Geber `slice-mock-zuordnung`, `AGENTS.md` §3.13).
- R4, Sendungen an `slice-mock-schreiben` (Eintrag in dessen §1): die Mutationen ersetzen `SPEC-195` Nr. 7 für vier Methoden (Punkt (a), Geber dieser Slice); die Punkte (b) bis (d) tragen Teil 2 — **Ausgang:** offen; Adresse: `slice-mock-schreiben` (§1 *Übernimmt*, Erwartung von `slice-mock-zuordnung` und `slice-mock-lesen`); der Planner hat geprüft, dass der Nehmer den Punkt annimmt (kein Ausschluss trifft ihn).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Das Argument der Zuordnung ist der Header `X-Recorder-Workspace` (die Kennung) und der Pfad (der Service); „Header fehlt“ meint das Fehlen der Zeile, „leer“ den leeren Wert oder die leere Liste, „mehrfach“ zwei oder mehr Werte. Die Spalte *Coverage* führt die Achsen Zustand (`complete`, `partial`, `unknown`) × Methode (`GET`, `HEAD`) × Modus von `STRICT_COVERAGE`; die Zeilen für Collection, Einzel-Lookup und `/$count` stehen in Teil 2.

| Operation | Header fehlt | leer | mehrfach | fremd (anderer Service) | unbekannt | Coverage (Zustand × Methode × Modus) | Abbruch (`ctx`) | Wiederholung |
|---|---|---|---|---|---|---|---|---|
| Zuordnung des Service (Pfad) | `SPEC-195` Nr. 4: ein leerer Pfad oder einer ohne führenden `/` ist `ODR-E5005` | `SPEC-195` Nr. 4 (derselbe Fall) | `SPEC-195` Nr. 4: passen mehrere Roots, gilt der längste | gilt nicht: der Service ist der des Pfads, ein fremder Workspace steht in der nächsten Zeile | `SPEC-195` Nr. 4 (`ODR-E5005`) | gilt nicht: die Wahl liest keine Coverage | gilt nicht: die Wahl ruft keinen Port (`SPEC-195` Nr. 2) | `SPEC-195` Nr. 12 (gleiche Anfrage, gleiche Antwort) |
| Zuordnung des Workspace (Header) | `SPEC-195` Nr. 5 (`ODR-E5003`) | `SPEC-195` Nr. 5 (leerer Wert, leere Liste) | `SPEC-195` Nr. 5 (auch zwei gleiche, auch zwei Schreibweisen des Namens) | `SPEC-195` Nr. 8 (`ODR-E5004`) | `SPEC-195` Nr. 8 (`ODR-E7001`) | gilt nicht: die Zuordnung liegt vor der Ressource (`SPEC-195` Nr. 3 c, f) | `SPEC-195` Nr. 2 und 11 (S2) | `SPEC-195` Nr. 12 |
| `$metadata` | wie Zuordnung des Workspace: sie gilt vor jeder Ressource (`SPEC-195` Nr. 3 c) | wie dort (`SPEC-195` Nr. 5) | wie dort (`SPEC-195` Nr. 5) | `SPEC-195` Nr. 8 | `SPEC-195` Nr. 8; Service unbekannt oder ohne Metadaten: `SPEC-195` Nr. 9 | gilt nicht: `$metadata` liest keine Coverage (`SPEC-197` Nr. 4) | `SPEC-195` Nr. 2 und 11 (S3, S4, S12), Strom geschlossen: `SPEC-198` Nr. 4 | `SPEC-198` Nr. 4 (byte-gleich), `SPEC-195` Nr. 12; `HEAD`: `SPEC-198` Nr. 5 |

**Zusatztabelle** (`AGENTS.md` §3.16: mehrere Verstöße am selben Schlüssel und Abhängigkeit vom Modus), mit der Spalte *Fänger*:

| Fall | Ort | Fänger |
|---|---|---|
| Je zwei Verstöße der ersten acht Stufen von `SPEC-195` Nr. 3 (Argumente, kein Service, Zuordnung, URL, Methode, Workspace unbekannt, Workspace fremd, Service ohne Metadaten) | `SPEC-195` Nr. 3, `SPEC-174` Nr. 4 | Z11 (M4: 8 Stufen, 27 baubare Paare; „Workspace unbekannt“ und „fremd“ schließen sich aus; die 27 Paare mit den Stufen des Lesens trägt L44 in Teil 2) |
| Fehler eines Ports neben einem Kontextfehler | `SPEC-195` Nr. 2, `SPEC-109` Nr. 6 | Z20 für S2, S3, S4, S12 (M6: Form `errors.Join` an jeder Stelle; S5 bis S11 trägt L43 in Teil 2) |
| Fehler des URL-Lesers neben einer anderen Methode | `SPEC-195` Nr. 3 und 7 | Z10 und Z11 (die Methode folgt dem Lesen der URL) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-07; Zähler = Zahl der Dateien in `evidence/`). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, `AGENTS.md` §3.15), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.15 und §3.16), `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (28×, `AGENTS.md` §3.17) und `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (21×, `AGENTS.md` §3.9), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14), `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×, Reviewer-Skill), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (8×, beide `AGENTS.md` §3.10) für Z22, `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (19×, `AGENTS.md` §3.14). Offen und für diesen Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) für die Fakes der Ports; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (5×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure von `welle-mock`) für die Zählprüfungen gegen eine Menge (M4 bis M6); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (13×, `geplant`: Neukalibrierung) für die Marke, die trotzdem der Formel folgt; `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (5×) für die Gegenproben bei wachsendem Produktcode. Mit dem Slice erreicht kein offener Eintrag ohne Ausgang 3×, außer den genannten.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
