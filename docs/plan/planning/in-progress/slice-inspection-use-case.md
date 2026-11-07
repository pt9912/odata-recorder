# Slice slice-inspection-use-case: InspectionUseCase, Exchanges, Coverage und Diagnose lesen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-FA-11`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Exchanges, Coverage, Diagnose) · `SPEC-236` (Vertrag, neu) · `SPEC-005` (Zeilen `GET …/exchanges`, `GET /__admin/coverage`, `GET /__admin/diagnostics` mit Parametern) · `LH-FA-11.a` · `SPEC-044`, `SPEC-043` (Coverage und Evidenz) · `SPEC-105` (*Ungeprüft*-Satz), `SPEC-107` (Repositories, aus denen gelesen wird) · `SPEC-109` (Fehlerabbildung, Anwendung in `SPEC-236` Nr. 8) · `SPEC-071`, `SPEC-112` (Zeile ergänzt) · `ARC-002`, `ARC-003` · `architecture.md` §2.2 (Zeile `InspectionUseCase`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-inspection-use-case)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `InspectionUseCase` (Inbound Port in `src/hexagon/ports/inbound`, Dienst in `src/hexagon/services/inspection`) liest die Exchanges einer Session, die Coverage einer Baseline und die Diagnose (Konfigurationszusammenfassung ohne Geheimnisse, Modus, Zähler je Fehlercode) paginiert und ohne Geheimnisse; er verändert nichts und kennt kein HTTP.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Endpunkte, Pfade und Statuscodes — `slice-admin-sessions`; der Use Case kennt kein HTTP, der Adapter bildet ab.
- Anlegen, Beenden und Löschen von Sessions — `SessionUseCase` (`done/`); `InspectionUseCase` liest nur.
- Die Ports `Telemetry` und `TelemetryReader` und die Zähler selbst — `welle-betrieb` (`slice-telemetry-port`, `done/`); der Dienst liest die Zähler über `TelemetryReader`, er erzeugt sie nicht und erhält nie `Telemetry` oder den Adapter (`SPEC-230` Nr. 9). Das Zählen in den Diensten ist Einbau (`slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-mock`).
- Export und Abruf der Metriken und Spans — `welle-betrieb` (Welle §6).
- Audit des Lesens — `SPEC-005` auditiert die Schreibendpunkte; ein Lesezugriff schreibt keinen Eintrag.
- Den Import-Report lesen — `ImportUseCase.Get` (`done/`); der Endpunkt liegt in `slice-admin-imports`.

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

- [ ] Exchanges: die Exchanges einer Session paginiert (Zeiger `sequence_no`, Standard 100, Obergrenze 1000, `SPEC-236` Nr. 3 und 4); eine Session ohne Exchange ergibt eine leere Seite, eine unbekannte Session `ODR-E7003`; die Zeile führt nur die Felder der Allowlist, nie Header, Query, Body oder Blob-Referenz (`SPEC-236` Nr. 5).
- [ ] Coverage: die Einträge je EntitySet einer Baseline (Parameter ist die Baseline-Kennung) mit Status und Evidenz aus `CoverageRepository`; eine unbekannte Baseline `ODR-E7002` (`SPEC-236` Nr. 6).
- [ ] Diagnose: Konfigurationszusammenfassung aus der Allowlist von Schlüsseln (nie Token, Zugangsdaten, Adressen oder Pfade, `SPEC-236` Nr. 7), Modus und Zähler je Fehlercode über den Lese-Port `TelemetryReader` (`Snapshot().Errors`, ein Eintrag je Code, auch mit 0; `SPEC-230` Nr. 9).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/hexagon/ports/inbound/` | update | `InspectionUseCase` |
| `src/hexagon/services/inspection/` | neu | Lesen der Exchanges, der Coverage und der Diagnose; Seitenbildung |
| `src/hexagon/ports/outbound/`, `src/adapters/driven/sqlite/` | unverändert | Entscheidung des Architect (2026-11-07): `ListBySession` bleibt, der Dienst schneidet die Seite im Speicher (`SPEC-236` Nr. 4); beide Pfade liegen nicht im Block |
| `spec/spezifikation.md` | update (Architect, erledigt) | `SPEC-236` mit Seitenform, Feldern der Exchanges, Parametern der Coverage und Allowlist der Diagnose; `SPEC-005`, `SPEC-105` (*Ungeprüft*-Satz), `SPEC-112`, Historie |
| Tests | neu | Fälle je Operation; die Fakes der Repositories und `FakeTelemetry` (beide Ports, aus `slice-telemetry-port`) halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service |

Der Vertrag steht als `SPEC-236` in der Spezifikation (Architect, 2026-11-07), die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) hier und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind entschieden, der Implementer entscheidet keine im Code.
Die Spalte `Nr.` trägt den Buchstaben der Art: `K` Konfiguration, `A` Kontext nil, `X` Exchanges, `C` Coverage, `D` Diagnose, `N` akzeptiertes Negativ, `W` Zählprüfung der Stellen und Nebenläufigkeit. **Teil 1** des Schnitts (§4, Exchanges) sind K1, A1, X1 bis X13, N1, W1, W2; **Teil 2** (Coverage und Diagnose) K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3, W4. Die Spalte *Fänger* nennt Fälle und die Mengen des Abschnitts *Pflicht-Mengen* darunter; die Namen der Tests wählt der Implementer. Jede Zeile ist ein Weg des Dienstes (*Modell-Menge der Wege* unten); ein Test, der die Wege im Quelltext zählt, hält Zeilen und Wege gleich (W1, W3).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| K1 | `New` weist ein nil in `Exchanges` ab (`SPEC-236` Nr. 1) | Fall: sonst gültige `Konfig`, `Exchanges` nil: `model.ErrUngueltig`, kein Dienst | Prüfung entfernt: `New` liefert einen Dienst | andere |
| K2 | `New` weist ein nil in `Coverage` und in `Zaehler` ab (Nr. 1) | Schleife über beide Felder, je ein nil, sonst gültig: `ErrUngueltig`, kein Dienst | die Prüfung eines der beiden Felder entfernt | andere |
| K3 | `New` nimmt genau die vier Modi an, jeder andere ist `ErrUngueltig` (Nr. 1) | Schleife über `record`, `replay`, `mock`, `hybrid` (gültig) und `""`, `Record`, `record ` (Leerzeichen dahinter), ` record`, `live` (ungültig) | Modus-Prüfung entfernt; Prüfung ohne `hybrid` | andere |
| K4 | `Einstellungen` darf nil sein; `New` kopiert tief, eine spätere Änderung der Abbildung wirkt nicht auf den Dienst (Nr. 1) | Fälle: nil gültig; nach `New` einen Wert ändern, einen Schlüssel hinzufügen und einen löschen, dann `Diagnose` | Abbildung ohne Kopie gespeichert | andere |
| A1 | `ctx` nil bei `Exchanges` ist `ErrUngueltig` ohne `Fehler`-Wert, kein Port wird gerufen, auch zusammen mit einem ungültigen Seitenparameter (Nr. 2, 3) | Schleife über Seite {gültig, `Nach` -1, `Limit` -1}: `errors.As(*Fehler)` falsch; `FakeRecording` zählt jede Operation: 0 | Prüfung entfernt: der Port erhält ein nil-`ctx` | andere |
| A2 | `ctx` nil bei `Coverage` und `Diagnose` ist `ErrUngueltig` ohne `Fehler`-Wert, kein Port wird gerufen (Nr. 2) | Schleife über beide Operationen; `FakeCoverage` und `FakeTelemetry` zählen jede Operation: 0 | Prüfung bei einer der beiden entfernt | andere |
| X1 | Ein `Nach` oder `Limit` unter 0 ist `ErrUngueltig` ohne Code, der Text nennt den Wert nicht, kein Port wird gerufen (Nr. 3) | Schleife über Parameter {`Nach`, `Limit`} × Wert {-1, -2, kleinster `int`}; `FakeRecording` zählt 0; der Text enthält die Zahl nicht; dazu `Nach` 0 und `Limit` 0 gültig | `< 0` zu `<= 0`: 0 wird abgewiesen; Prüfung von `Nach` oder von `Limit` entfernt | andere |
| X2 | Wirksames `Limit`: 0 ist `StandardLimit`, über `MaxLimit` wird auf `MaxLimit` gekappt (kein Fehler), sonst unverändert; die Antwort trägt es (Nr. 3) | Schleife über Menge M-Seite (Achse Limit) bei Bestand über `MaxLimit`; erwartet aus einer Rechnung des Tests (`min`), die Konstanten aus dem Paket, nie 100 oder 1000 im Test | Kappung entfernt; 0 → Standard entfernt (leere Seite); Kappung ab `MaxLimit - 1` | andere |
| X3 | Seite = Modell: `Eintraege`, `Gesamt`, `Weiter`, Reihenfolge (Nr. 4) | Generator M-Seite mit Orakel (Modell der Seiten aus der Menge 1 bis n, Schleife im Test); darin n gleich 0 (leere Liste, nie `nil`, `Weiter` 0) | `>` statt `>=` im Filter; `Weiter` aus dem ersten statt dem letzten Eintrag; Umkehr der Reihenfolge | andere |
| X4 | An jeder Grenze der Seite: `Weiter` ist 0 genau dann, wenn kein Exchange hinter dem letzten Eintrag folgt; `Nach` gleich oder über der letzten `sequence_no` ist kein Fehler (Nr. 4) | Fälle: `Nach` gleich `Gesamt` - 1, `Gesamt`, `Gesamt` + 1; `Nach` + `Limit` gleich `Gesamt` - 1, `Gesamt`, `Gesamt` + 1; `Limit` gleich `MaxLimit` bei n gleich `MaxLimit`; `Gesamt` zählt alle | `Weiter` bei `Nach + Limit == Gesamt` gesetzt; Fehler bei `Nach > Gesamt` | andere |
| X5 | Die Folge der Seiten (`Nach` gleich `Weiter` bis 0) liefert jeden Exchange genau einmal in Reihenfolge; gleiche Argumente zweimal ergeben das gleiche Ergebnis (Nr. 4, 10) | Generator M-Seite, Schleife über die Folge; Vereinigung gleich 1 bis n; zweiter Aufruf gleich erster | `Weiter` um 1 zu klein (Doppelung); um 1 zu groß (Lücke); Zustand zwischen Aufrufen | andere |
| X6 | Anhängen zwischen zwei Seiten verschiebt nichts: kein Exchange doppelt, keiner fehlt, `Gesamt` wächst (Nr. 4 *Stabilität*) | Generator M-Bestand: `FakeRecording` hängt nach dem ersten Aufruf k Exchanges an; Folge gelesen bis `Weiter` 0; Orakel: 1 bis n + k, wenn die Folge die angehängten erreicht | Zeiger als Versatz in der gelesenen Liste statt `sequence_no` (bleibt gleich bei Anhängen; die Mutation: Zeiger als Seitennummer) | andere |
| X7 | `ListBySession` wird genau einmal gerufen, mit `sessionID` unverändert, sonst kein Port (Nr. 4) | Schleife über `s1`, `""`, ein Leerzeichen, 64 Zeichen, `ses-` mit zwei Bytes jenseits von ASCII (aus Bytewerten); der Fake hält Argument und Zahl der Aufrufe fest | `strings.TrimSpace` auf der Kennung; ein zweiter Aufruf (`Get`) | andere |
| X8 | Ein Kontextfehler des Ports bleibt unverändert und ohne Code, auch neben einem Fehler der Tabelle; danach kein Port (Nr. 2) | Schleife über Form {`Canceled`, `DeadlineExceeded`, verpackt, `errors.Join` mit `ErrNichtGefunden` in beiden Reihenfolgen}; `errors.Is` und `==`-Gleichheit, kein `*Fehler` | Fehlerabbildung vor der Prüfung des Kontexts | andere |
| X9 | Jeder Fehler des Ports wird nach `SPEC-109` Nr. 5 mit dem Gegenstand `Session` abgebildet: `Code`, `Status` (der eine Status), `Ursache` erreichbar (Nr. 8) | Schleife über Menge M-Fehler (Aufrufstelle `ListBySession`): jeder Eintrag von `outbound.EintraegeAbbildung()`, ein anderer Fehler (`ODR-E9001`); erwartet `ODR-E7003` für `ErrNichtGefunden` | gegebener Gegenstand `Snapshot`; `Stati()[1]`; Ursache nicht verpackt | andere |
| X10 | Der `ctx` des Aufrufs geht unverändert an den Port; ein vor dem Aufruf beendeter Kontext endet mit seinem Fehler; ein nach der Rückkehr beendeter ändert das Ergebnis nicht (Nr. 2) | Fälle: Identität des `ctx` im Fake; vor dem Aufruf abgebrochen (Fake liefert `ctx.Err()`); der Fake bricht ab und liefert Daten: Ergebnis ohne Fehler | `context.Background()` an den Port; `ctx.Err()` nach dem Aufruf geprüft | andere |
| X11 | `ExchangeZeile` führt genau die Felder der Allowlist (Nr. 5) mit unverändertem Wert; die Allowlist ist die Menge der Felder von `model.Exchange` abzüglich der ausgeschlossenen | Zählprüfung per Reflexion: enthaltene + ausgeschlossene Felder des Tests gleich `NumField` von `model.Exchange`, Felder von `ExchangeZeile` gleich den enthaltenen; Schleife über jedes enthaltene Feld mit eigenem Wert (Fixture ungleich Mutationswert); `FinishedAt` Nullzeit, `ResponseStatus` 0 | ein Feld vertauscht oder ungesetzt; ein neues Feld in `model.Exchange` ohne Entscheidung (Zählprüfung rot) | andere |
| X12 | Kein Geheimnis in Exchanges: kein Name und kein Wert eines Headers, kein Query-Wert, keine Blob-Referenz, kein `match_key` in der Antwort (Nr. 5) | Allaussage als Schleife über Menge M-Feld: Canary (aus Bytewerten) an jedem Einbringungsort der ausgeschlossenen Felder × Kodierung {`encoding/json`, `%+v`, Gang über alle Zeichenketten der Antwort}, über alle Seiten; Gegenprobe: ein Canary im `Path` erscheint | ein ausgeschlossenes Feld in die Zeile kopiert (je Feld eine Mutation, auch ein Header-Wert) | andere |
| X13 | Lesen schreibt nichts: `Exchanges` ruft von `RecordingRepository` nur `ListBySession` (Nr. 1) | `FakeRecording` zählt jede Operation (`Append`, `Get`, `ListBySession`) bei Erfolg, Port-Fehler, Parameterfehler, `ctx` nil: `Append` und `Get` 0, Bestand nach dem Aufruf gleich dem davor | `Get` je Zeile; `Append` eines Vermerks | andere |
| N1 | Akzeptierte Negative von Teil 1: die Größe einer Session (Schnitt im Speicher) und ein Geheimnis im Pfad (`SPEC-236` *Ungeprüft*) | kein Fänger; Grund und Folge in `SPEC-236` Nr. 4 und 5; die Gegenprobe in X12 hält den Pfad als sichtbar fest | – | andere |
| W1 | Zählprüfung Teil 1: die Wege von `Exchanges` im Quelltext sind die Zeilen dieser Tabelle (Aufrufe der Ports, Aufrufe der Abbildung, Stellen eines Argumentverstoßes) | Test mit `go/parser` über die Dateien des Pakets ohne `_test`: Zahl der Aufrufe über Felder von `Konfig`, der Aufrufe von `mitCode` und der Stellen, die `ErrUngueltig` bilden, gleich der Zahl der Wege der Tabelle des Tests (Name, Art, Nr. dieser Tabelle) | ein weiterer Port-Aufruf, eine weitere Stelle `ErrUngueltig` ohne Zeile | andere |
| W2 | Gleichzeitige Aufrufe von `Exchanges` stören sich nicht, der Dienst hält keinen Zustand (Nr. 10) | 8 Goroutinen mit je 50 Aufrufen verschiedener Seiten gegen einen Dienst; jedes Ergebnis gleich dem sequentiellen; `make test-race` je Prozess (`-test.count=1`) | der Dienst puffert die gelesene Liste in einem Feld: Detektor rot in N von N Prozessläufen (Zahl nennt der Bericht) | nebenläufig |
| C1 | Eine leere `baselineID` ist `ErrUngueltig` ohne Code und ohne Port; jede andere geht unverändert und genau einmal an `List` (Nr. 6) | Schleife über `""` (ungültig), `b1`, ein Leerzeichen, `w1` (fremd), 64 Zeichen, zwei Bytes jenseits von ASCII; `FakeCoverage` hält Argument und Zahl der Aufrufe | `TrimSpace`; leere Kennung an den Port | andere |
| C2 | Das Ergebnis hat einen `CoverageEintrag` je Eintrag des Ports in dessen Reihenfolge, `EntitySet`, `Status` und `Evidenz` byte-gleich; eine Baseline ohne Einträge liefert eine leere Liste (nie `nil`); keine Seitenbildung, kein abgeleiteter Status (Nr. 6) | Generator M-Coverage mit Orakel (die Fixture-Liste in Reihenfolge); darin m gleich 0 und m gleich 20 | Evidenz geparst und neu kodiert; Status aus der Evidenz abgeleitet; die Liste auf `StandardLimit` gekürzt | andere |
| C3 | Jeder Fehler von `List` wird nach `SPEC-109` Nr. 5 mit dem Gegenstand `Snapshot` abgebildet: `ErrNichtGefunden` (auch für eine Kennung anderer Art) ist `ODR-E7002` (Nr. 6, 8) | Schleife über Menge M-Fehler (Aufrufstelle `List`) wie in X9; erwartet `ODR-E7002` | Gegenstand `Session` (ergäbe `ODR-E7003`) | andere |
| C4 | Kontextfehler unverändert ohne Code; `ctx` an `List` unverändert; Abbruch nach der Rückkehr ändert das Ergebnis nicht (Nr. 2) | die Fälle von X8 und X10 für `List` | wie X8 und X10 | andere |
| C5 | Lesen schreibt nichts: `Coverage` ruft von `CoverageRepository` nur `List`, genau einmal (Nr. 1) | `FakeCoverage` zählt `Get` und `List` bei Erfolg, Fehler und Parameterfehler: `Get` 0, `List` höchstens 1 | `Get` je Eintrag; zweiter `List` | andere |
| D1 | `Diagnose` mit beendetem `ctx` ist dessen Fehler, unverändert und ohne Code, ohne Aufruf von `Snapshot`; ein lebender liefert das Ergebnis (Nr. 7) | Fälle: lebend, `Canceled`, `DeadlineExceeded`; `FakeTelemetry.Anzahl(OpSnapshot)` | `ctx.Err()` nicht geprüft; geprüft nach `Snapshot` | andere |
| D2 | Der Dienst ruft `Telemetry` nie; `Snapshot` genau einmal je Aufruf (Nr. 1, 7) | Schleife über jede Operation des `FakeTelemetry` (Quelle: die Konstanten von `supporttest.Operation`, Zählprüfung gegen ihre Menge): außer `Snapshot` 0, `Snapshot` 1 je Aufruf | ein `CountError` oder `Log` im Dienst; zweiter `Snapshot` | andere |
| D3 | `Fehler` hat einen Eintrag je Code von `model.AlleCodes()` in dieser Reihenfolge, `Anzahl` gleich dem Zähler (Nr. 7) | Generator M-Codes mit Orakel (Fixture-Abbildung); Zählprüfung gegen `len(model.AlleCodes())`, nie 37 | Schleife über die Schlüssel der Abbildung statt der Tabelle (Zählerstand 0 fehlt); Reihenfolge der Abbildung; ein Code ausgelassen | andere |
| D4 | Ein Code ohne Eintrag hat `Anzahl` 0; ein Eintrag außerhalb der Tabelle wird verworfen; ein nil `Errors` ist gültig (Nr. 7) | Fälle: `Errors` nil, leer, nur ein Code; Schlüssel 0, -1, letzter + 1, größter `int32` | Eintrag außerhalb durchgereicht; nil-Abbildung ohne Behandlung | andere |
| D5 | `Einstellungen` enthält genau die Einträge, deren Schlüssel byte-gleich in `Erlaubt()` stehen, mit unverändertem Wert; ein fehlender fehlt (Nr. 7) | Schleife über `Erlaubt()` (jeder Schlüssel, eigener Wert); Zählprüfung: `Erlaubt()` gleich der Literal-Liste im Test (Wortlaut der Nr. 7); Abwandlungen jedes erlaubten Schlüssels (Kleinschreibung, Leerraum davor und dahinter, Präfix, Suffix) fehlen; ein erlaubter ohne Eintrag fehlt, kein leerer Wert | Vergleich ohne Beachtung der Schreibweise; `TrimSpace` auf den Schlüssel; ein Schlüssel zu viel in `Erlaubt()` | andere |
| D6 | Kein Geheimnis in der Diagnose: kein Wert unter einem ausgeschlossenen Schlüssel erscheint (Nr. 7) | Allaussage als Schleife über Menge M-Schlüssel: Canary als Wert unter jedem ausgeschlossenen Schlüssel (Liste der Nr. 7 und `ADMIN_TOKEN`) × Kodierung {json, `%+v`, Gang über alle Zeichenketten}; Gegenprobe: derselbe Canary unter einem erlaubten Schlüssel erscheint | Allowlist zur Sperrliste umgekehrt; `DATA_DIR` oder `UPSTREAM_URL` in `Erlaubt()` | andere |
| D7 | `Modus` ist der `Modus` der Konfiguration, unverändert (Nr. 7) | Schleife über die vier Modi | Modus aus `Einstellungen["MODE"]`; fester Wert | andere |
| D8 | Das Ergebnis ist eine Kopie und hat nie nil: `Einstellungen` und `Fehler`; jeder Aufruf liest den Stand neu (Nr. 7) | Fälle: nil-Konfiguration ergibt `{}` und `[]` (JSON); Änderung des Ergebnisses wirkt nicht auf den Dienst oder das nächste Ergebnis; Änderung des Standes im Fake zwischen zwei Aufrufen zeigt das zweite Ergebnis neu | Abbildung der Konfiguration ohne Kopie zurückgegeben; Ergebnis des ersten Aufrufs gespeichert | andere |
| N2 | Akzeptierte Negative von Teil 2: Schlüsselmenge und Modi des Root nicht gegen den Dienst gezählt; Zählen der Fehler des Dienstes (`SPEC-236` *Ungeprüft*) | kein Fänger; Grund und Folge in `SPEC-236` *Ungeprüft*; D5 hält `Erlaubt()` gegen den Wortlaut | – | andere |
| W3 | Zählprüfung Teil 2: die Wege von `Coverage`, `Diagnose` und `New` im Quelltext sind die Zeilen dieser Tabelle | wie W1 für die Dateien von Teil 2 und für `New` | eine Stelle `ErrUngueltig` oder ein Port-Aufruf ohne Zeile | andere |
| W4 | Gleichzeitige Aufrufe von `Coverage` und `Diagnose` stören sich nicht, der Dienst hält keinen Zustand (Nr. 10) | wie W2 für beide Operationen, auch gemischt mit `Exchanges` | Zwischenspeicher der Coverage oder der Zusammenfassung im Dienst: Detektor rot | nebenläufig |

**Pflicht-Mengen** (`AGENTS.md` §3.15, §3.17, `implement-slice` Schritt 14): je Menge Achsen, Zahl der Elemente als Produkt, Generator, Orakel (nie der Dienst und nie sein Code), Zählprüfung gegen die Quelle der Menge (keine feste Zahl im Test), Prüfgrenze der unbeschränkten Menge.

| Menge | Achsen | Elemente | Orakel | Zählprüfung gegen die Quelle | Prüfgrenze (akzeptiertes Negativ) |
|---|---|---|---|---|---|
| M-Seite (X1 bis X5) | `Limit` {0, 1, `StandardLimit`, `MaxLimit` - 1, `MaxLimit`, `MaxLimit` + 1, größter `int`} × Bestand n {0, 1, L - 1, L, L + 1, 2L, 2L + 1} mit L gleich dem wirksamen `Limit` × `Nach` {0, 1, n - 1, n, n + 1, n + L}; ungültig: `Nach` und `Limit` {-1, -2, kleinster `int`} | 7 × 7 × 6 = 294 vor dem Entfernen gleicher Fälle (L gleich 1 macht n - 1 gleich 0), der Test zählt die erzeugten nach der Formel | Modell der Seiten: der Test bildet die Exchanges 1 bis n und schneidet in einer eigenen Schleife (`sequence_no` größer `Nach`, die ersten L), `Weiter` gleich der Nummer des letzten, wenn sie kleiner n ist; kein Aufruf des Dienstes | Konstanten `StandardLimit` und `MaxLimit` des Pakets, nie 100 oder 1000; Zahl der Fälle gleich der Formel aus den Längen der Achsen | `Limit` und n sind unbeschränkt: die Grenze ist `MaxLimit` und ein Bestand bis 2 · `MaxLimit` + 1; Folge: ein Fehler erst bei Bestand über 2 · `MaxLimit` + 1 bliebe ungefangen, der Dienst rechnet ohne Fallunterscheidung nach Größe |
| M-Bestand (X6) | Anhängen k {0, 1, L, L + 1} × Zeitpunkt {nach der ersten, nach der letzten Seite} × n {0, 1, L, L + 1, 2L + 1} | 4 × 2 × 5 = 40 | Folge der Seiten gelesen bis `Weiter` 0; erwartet jede `sequence_no` von 1 bis n + k genau einmal in Reihenfolge, soweit die Folge sie erreicht (nach der letzten Seite angehängt: nicht erreicht) | Längen der Achsen | Anhängen während des Aufrufs von `ListBySession` gilt nicht (ein Aufruf liest einen Stand, `SPEC-105` Nr. 13, `SPEC-236` Nr. 10) |
| M-Feld (X11, X12) | enthaltene Felder (X11); Einbringungsorte der ausgeschlossenen Felder: `SessionID`, `RawQuery`, `MatchKey`, Name und Wert in `RequestHeaders`, Name und Wert in `ResponseHeaders`, `RequestBlob`, `ResponseBlob` × Kodierung {`encoding/json`, `%+v`, Gang über alle Zeichenketten} (X12) | X11: Zahl der enthaltenen Felder; X12: 9 Orte × 3 = 27 | der Canary und der Wert des Feldes stehen im Test als Bytewerte; erwartet: Gleichheit des Wertes (X11), Abwesenheit des Canary (X12) | per Reflexion: enthaltene + ausgeschlossene Felder gleich den Feldern von `model.Exchange`, ausgeschlossene Orte gleich den ausgeschlossenen Feldern (Header zählen als ein Feld mit zwei Orten) | Kodierungen des Canary sind unbeschränkt (Hash, Teilstück, Umkodierung): drei geprüft; Folge: ein Dienst, der ein ausgeschlossenes Feld umkodiert in die Zeile setzte, bliebe ungefangen; der Dienst kopiert Felder und kodiert nichts. Body und Query als Art: der Dienst hat keinen Blob-Port und kein Parsen, ein Body kommt nur als Blob-Referenz (Ort `RequestBlob`, `ResponseBlob`), eine Query nur als `RawQuery` |
| M-Coverage (C2) | `Status` {`complete`, `partial`, `unknown`, ein fremder Text} × `Evidenz` {ein JSON-Text, leer, ein Text mit Leerraum und Zeichen jenseits von ASCII aus Bytewerten} × Zahl der Einträge m {0, 1, 2, 20}; Kennung (C1) {leer, bekannt, Leerraum, fremd (Kennung eines Workspace-Snapshots), 64 Zeichen, Bytes jenseits von ASCII} | 4 × 3 × 4 = 48; C1: 6 | die Fixture-Liste in der Reihenfolge, in der der Fake sie liefert | die drei Status sind die Konstanten `model.CoverageComplete`, `model.CoveragePartial`, `model.CoverageUnknown`; der Test prüft für jede `Coverage{...}.Validate()` gleich nil und für den fremden Text ungleich nil, damit „fremd“ fremd bleibt | Evidenz-Inhalte sind unbeschränkt: drei Formen; Folge: ein Dienst, der die Evidenz liest und umformt, würde bei einer weiteren Form auffallen, nicht bei diesen; der Dienst parst sie nie |
| M-Codes (D3, D4) | Code {jeder von `model.AlleCodes()`} × Stand im Fake {fehlt, 0, 1, 2 hoch 32, größter `uint64`}; außerhalb der Tabelle {0, -1, letzter + 1, größter `int32`} | Länge von `model.AlleCodes()` × 5 (zurzeit 37 × 5, mit `ODR-E6008` aus `slice-mock-zuordnung` 38 × 5), dazu 4 | die Fixture-Abbildung; erwartet je Code der Tabelle ein Eintrag mit dem Fixture-Wert, fehlend 0 | `len(model.AlleCodes())` und die Reihenfolge von `AlleCodes()`, nie die Zahl im Test | keine; die Menge der Codes ist beschränkt und kommt aus der Quelle |
| M-Schlüssel (D5, D6) | erlaubter Schlüssel {jeder von `Erlaubt()`} × Abwandlung {unverändert, Kleinschreibung, Leerraum davor, Leerraum dahinter, Präfix, Suffix}; ausgeschlossener Schlüssel {die Liste der Nr. 7 und `ADMIN_TOKEN`} × Kodierung {json, `%+v`, Gang über alle Zeichenketten}; Modus {4} | Länge von `Erlaubt()` × 6; Länge der Liste × 3; 4 | Fixture: jeder erlaubte Schlüssel mit eigenem Wert; erwartet nur die unveränderten | die Literal-Liste im Test gleich `Erlaubt()` (Wortlaut Nr. 7); die Liste der ausgeschlossenen gleich den Schlüsseln der Nr. 7 | die Menge der Schlüssel der Startkonfiguration liegt im Root (nicht importierbar, `SPEC-236` *Ungeprüft*): ein Schlüssel, den der Root neu einführt, ist ausgeschlossen (Standard ist Verwerfen) und wird nie gezeigt; Folge: unschädlich, die Allowlist ist die Zusage |
| M-Fehler (X8 bis X10, C3, C4, D1) | Aufrufstelle {`ListBySession`, `List`} × Fehler {jeder Eintrag von `outbound.EintraegeAbbildung()`, ein anderer Fehler} ; Aufrufstelle × Form des Kontextfehlers {`Canceled`, `DeadlineExceeded`, verpackt, Join vorn, Join hinten}; Aufrufstelle × `ctx`-Zustand {lebend, vor dem Aufruf abgebrochen, vom Fake abgebrochen mit Daten}; `Diagnose` × {lebend, `Canceled`, `DeadlineExceeded`} | 2 × (Zahl der Einträge + 1) + 2 × 5 + 2 × 3 + 3 | `outbound.Abbilden` ist die Funktion unter Test in `SPEC-109`; das Orakel hier ist die Tabelle des Tests (Name, Wert, erwarteter Code je Gegenstand), die aus `SPEC-109` Nr. 5 abgeschrieben ist, nicht der Aufruf von `Abbilden` | die Tabelle des Tests hat genau die Namen von `outbound.EintraegeAbbildung()` (Quelle), nie eine feste Zahl | keine für die Fehler; die Form der Kontextfehler ist unbeschränkt (jede Verpackung): fünf geprüft, Folge: eine Verpackung, die `errors.Is` nicht trifft, ist ein Fehler des Ports |

**Modell-Menge der Wege** (Auftrag 10; der Test W1 und W3 zählt sie gegen den Quelltext): jeder Verwurf nach Grund, jede Seitengrenze, jede Fehlerabbildung je Port, jedes Feld der Allowlist.

| Operation | Wege | Zeilen |
|---|---|---|
| `New` | nil `Exchanges`; nil `Coverage` oder `Zaehler`; Modus ungültig; Einstellungen kopiert; Erfolg | K1, K2, K3, K4 |
| `Exchanges` | `ctx` nil; `Nach` < 0; `Limit` < 0; Limit 0; Limit über der Grenze; Limit sonst; `ListBySession`: Kontextfehler, Fehler mit Code (je Eintrag der Tabelle), Erfolg; Filter `sequence_no` > `Nach`; Schnitt auf `Limit`; `Weiter` gesetzt oder 0; Feld je Eintrag der Allowlist | A1, X1 bis X13 |
| `Coverage` | `ctx` nil; Kennung leer; `List`: Kontextfehler, Fehler mit Code, Erfolg (je Eintrag eine Zeile, leer eine leere Liste) | A2, C1 bis C5 |
| `Diagnose` | `ctx` nil; `ctx` beendet; `Snapshot`; Eintrag der Tabelle (gesetzt, fehlend); Eintrag außerhalb; Schlüssel erlaubt, verworfen, fehlend; Modus | A2, D1 bis D8 |

Jeder Weg läuft in einem Test des Pakets und damit unter `make test-race`; ein Weg ohne Test ist nach W1 und W3 rot. Ein benanntes Negativ gibt es nur für die Fehler des Ports, die der echte Adapter nie liefert (etwa `ErrBlobFehlt` an `ListBySession`): der Fake liefert sie, der Weg ist derselbe wie für jeden Fehler (`mitCode`), Folge keine.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts.
- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Vertrag von `InspectionUseCase`:** Operationen (Exchanges, Coverage, Diagnose), Argumente und Rückgabetypen, Fehlerabbildung als Anwendung von `SPEC-109`, Ort in der Spezifikation.
2. **Paginierung:** `SPEC-005` und die Architektur sagen „paginiert“, kein Vertrag nennt Parameter (Seitengröße, Zeiger oder Versatz), Standard und Obergrenze, Reihenfolge, Verhalten bei ungültigem oder zu großem Wert, Stabilität bei gleichzeitigem Anhängen, Gesamtzahl. `RecordingRepository.ListBySession` liefert alle Exchanges: erweitert der Architect den Port um eine Seitenabfrage (berührt `SPEC-105` und den SQLite-Adapter und damit die Größe) oder schneidet der Dienst im Speicher (Grenze für große Sessions)?
3. **Felder eines Exchange in der Liste:** Kopfdaten (Kennung, Sequenz, Methode, Pfad, Status, `capture_status`, Zeitpunkte, Vermerke der Redaction) und ob Header-Werte, Query und Bodies (oder nur Größe und Hash) erscheinen; „ohne Geheimnisse“ als Allowlist der Felder, nicht als Liste der entfernten.
4. **Coverage:** `SPEC-005` nennt für `GET /__admin/coverage` keinen Parameter. Auf welche Baseline bezieht sich der Aufruf (Baseline-Kennung, Workspace, Service, alle), Form der Antwort je EntitySet mit Evidenz, Fehlerfall einer unbekannten Baseline (Code?).
5. **Diagnose:** Allowlist der Schlüssel der Konfigurationszusammenfassung (nie Token, Pfade von Geheimnisdateien als Inhalt, Zugangsdaten), Form der Zähler je Fehlercode, und die Anwendung von `TelemetryReader.Snapshot().Errors` (entschieden in `SPEC-230` Nr. 8 und 9 und im Satz hinter der Tabelle von `SPEC-005`; der Architect prüft Form, Reihenfolge und Fehlerfall, entscheidet nichts neu); die Zusammenfassung kommt als typisierter Wert vom Composition Root, der Dienst sieht nie die Umgebung.
6. **`ODR-E7012`:** `SPEC-112` kennt „Exchange unbekannt“ (Admin), `SPEC-005` hat keinen Endpunkt für einen einzelnen Exchange: Zeile ergänzen oder den Code anders verorten.
7. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Aufträge an den Architect, Ergänzung beim Start** (Planner, 2026-11-07; vor dem ersten Code-Commit, `AGENTS.md` §3.12; Ausgang jeweils *offen* in §6):

8. **Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15) gegen den Vertrag aus Auftrag 1 bis 6: je Zusage, Randform und Fall eine Zeile, auch „Lesen schreibt nichts“ (Fake jedes Repositories zählt jede Operation, auch `Create`, `Update`, `Delete`, `Append`), „kein Geheimnis in Exchanges und Diagnose“ (Allaussage, Fänger über alle Felder der Allowlist, `BEO-REPO/universalaussage-im-vertrag-ungeprueft`, 14×), der `ctx` und der Kontextfehler je Aufrufstelle (Exchanges: Session und Exchanges; Coverage; Diagnose: Zusammenfassung und `Snapshot`), Gleichheit zweier verglichener Werte an jeder Grenze der Seite (Seitengröße gleich Obergrenze, Versatz gleich Zahl der Exchanges). `FakeTelemetry` ist `TelemetryReader` und `Telemetry`; der Dienst darf `Telemetry` nie rufen, der Fake zählt es.
9. **Pflicht-Mengen mit Achsen** (`AGENTS.md` §3.15, §3.17, `implement-slice` Schritt 14; Register `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, 29×, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`, 6×, `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze`, 11×): je Klasse die **Achsen**, die Zahl der Elemente als Produkt, der Generator, das **Orakel** (unabhängig vom Dienst: Modell der Seiten aus der Menge der Exchanges, nie der Dienst selbst) und die Prüfgrenze der unbeschränkten Menge als akzeptiertes Negativ mit Grund und Folge im Betrieb. Mindestens die Achsen **Seite** (Größe 0, 1, Standard, Obergrenze minus 1, Obergrenze, Obergrenze plus 1, negativ, sehr groß; Versatz oder Zeiger an jeder Grenze: 0, letzte, hinter der letzten, ungültig), **Bestand** (0, 1, genau eine Seite, eine Seite plus 1, mehrere Seiten; Anhängen zwischen zwei Seiten), **Feld** (jedes Feld der Allowlist der Exchange-Zeile; jede Art, in der ein Geheimnis eingebracht sein kann: Header, Query, Body, Pfad), **Coverage** (jeder Status, Evidenz vorhanden und fehlend, 0, 1, viele EntitySets; unbekannte, leere und fremde Baseline), **Diagnose** (jeder Schlüssel der Zusammenfassung gegen die Allowlist; jeder Modus; jeder Code der Tabelle von `SPEC-109` mit Zählerstand 0 und größer 0) und **Fehler** (jeder Fehler jedes Ports aus `SPEC-109` Nr. 5 an jeder Aufrufstelle, `ctx` abgebrochen vor und nach jedem Aufruf). Die Zählprüfung hält die Zahl der Elemente gegen die Menge der Quelle (`model.AlleCodes()`, die Menge der Statuswerte, die Felder des Typs), nie gegen eine feste Zahl im Test: `slice-mock-zuordnung` (Linie B) fügt `ODR-E6008` hinzu (Tabelle 37 auf 38 Codes; der Zähler `Snapshot().Errors` hat dann 38 Einträge).
10. **Zweige und Rückgaben als Menge mit Zählprüfung** (Register `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, 7×, Schwelle erreicht): der Architect nennt die **Modell-Menge** der Wege von `Exchanges`, `Coverage` und `Diagnose` (jeder Verwurf nach Grund, jede Seitengrenze, jede Fehlerabbildung je Port, jedes Feld der Allowlist) und die Tabelle führt je Element eine Zeile mit Fänger; der Test zählt die Menge gegen diese Modell-Menge (Zahl der Zweige im Code gleich Zahl der Zeilen), ein Zweig, den kein Lauf unter dem Detektor erreicht, ist ein benanntes Negativ mit Grund und Folge. Trägt der Vertrag eine Zusage über Nebenläufigkeit (Anhängen während des Lesens, Stand der Zähler), gilt `AGENTS.md` §3.10 je Fenster, sonst steht „gilt nicht“ mit Grund.
11. **Schnittfrage** nach `SPEC-128` Nr. 1: der Planner schätzt die Größe nach der Tabelle des Architect (Block `groessenmarke` im Zweig); über der Schnittgrenze wird vor dem Code geschnitten. Der Plan liefert drei Operationen mit eigener Seiten- und Allowlist-Regel; der Architect nennt vor der Tabelle einen Schnitt (etwa Exchanges mit Seitenabfrage, Coverage, Diagnose; jeder Teil ohne geteilten Mechanismus) oder den `schnitt-grund`, sonst schneidet der Planner nach der Tabelle. Entscheidet der Architect die Seitenabfrage in `RecordingRepository` und im SQLite-Adapter (Auftrag 2), liegen `src/hexagon/ports/outbound/` und `src/adapters/driven/sqlite/` im Block (`pfad`) und der Schnitt trennt diesen Teil zuerst.
12. **Kennungen und Dateien** (Linie D, Bereich `SPEC` 230 bis 259, `ARC` 040 bis 049, `ADR` 0040 bis 0049, `CO` 060 bis 079; letzte vergebene `SPEC` der Linie: `SPEC-235`): neue Nummern ab `SPEC-236`; `slice-telemetry-log` (Zweig) kann `SPEC-236` ebenfalls brauchen, der Zusammenstoß ist der des Integrationsschritts, der Architect nennt die Nummer im Plan. Der Slice ändert in `spezifikation.md` die Zeilen der Admin-API (`SPEC-005` ff.), `SPEC-105` und `SPEC-107`, soweit der Wortlaut es trägt, `SPEC-112` und hängt eine Zeile an die Änderungstabelle am Dateiende. Neben den Zweigen: `slice-mock-zuordnung` (Linie B) ändert `SPEC-160`, `SPEC-109` Nr. 5 und 6, `SPEC-174`, `SPEC-175`, `SPEC-195` bis `SPEC-199`, die Zeile `odr_errors_total` der Tabelle der Metriken und die Änderungstabelle; `slice-telemetry-log` (Linie D) ändert `SPEC-231`, `SPEC-232`, `SPEC-230` und die Änderungstabelle. Dieser Slice berührt `src/adapters/driven/odata/**`, `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/mock.go`, `src/adapters/driven/telemetry`, `src/hexagon/ports/support/` und `src/bootstrap/` nicht; `src/hexagon/ports/inbound/` nur um eine neue Datei `inspection.go`. Eine Änderung an `SPEC-109` oder `SPEC-230` geht als Rückgabe an den Planner.
13. **Katalog-Pflicht** (`SPEC-201` Nr. 3 und 11): erwartet ist kein Eintrag im Katalog des Binary (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`); bringt der Vertrag doch einen neuen Fehlercode, Schlüssel oder ein Kommando (etwa einen Code für Seitenparameter oder eine unbekannte Baseline), steht er im selben Slice in `harness/katalog-ausnahmen.txt` und im Handbuch und in §4 des Plans, oder der Architect gibt die Frage an den Planner zurück (Rückführung `in-progress` → `open`, §4).

**Entscheidungen des Architect** (2026-11-07; Vertrag `SPEC-236`, Ausgang der Aufträge 1 bis 13; eine ADR entsteht nicht: die Entscheidungen ändern keinen Port und sind durch einen Folge-Vertrag umkehrbar, Gründe und Folgen stehen in `SPEC-236` wie in `SPEC-124`):

1. **Vertrag:** `SPEC-236`, hinter `SPEC-124`: drei Operationen `Exchanges(ctx, sessionID, seite)`, `Coverage(ctx, baselineID)`, `Diagnose(ctx)`; Typen im Paket `inbound`, Dienst `inspection.Inspektion` mit `Konfig{Exchanges, Coverage, Zaehler, Modus, Einstellungen}`; Fehlerabbildung als Anwendung von `SPEC-109` Nr. 5 (Gegenstand `Session` bzw. `Snapshot`), ein Argumentverstoß ist `model.ErrUngueltig` ohne Code (wie `SPEC-124` Nr. 3).
2. **Paginierung:** Zeiger ist die `sequence_no` (`Nach`, HTTP `after`), `Limit` (HTTP `limit`) mit Standard 100 und Obergrenze 1000; größere Werte werden gekappt, negative sind ungültig; die Antwort trägt `Gesamt` und `Weiter`; stabil beim Anhängen. **Der Port bleibt unverändert, der Dienst schneidet im Speicher** (`SPEC-236` Nr. 4): `SPEC-105` ändert nur den *Ungeprüft*-Satz, `src/hexagon/ports/outbound/` und `src/adapters/driven/sqlite/` liegen nicht im Block.
3. **Felder:** 15 Felder der Allowlist (`SPEC-236` Nr. 5), 7 ausgeschlossen (`SessionID`, `RawQuery`, `MatchKey`, beide Header-Abbildungen, beide Blob-Referenzen); der Pfad steht wie gespeichert (akzeptiertes Negativ, Folge in `SPEC-236`).
4. **Coverage:** Parameter ist die Baseline-Kennung (HTTP `baseline_id`, Pflicht), nicht Workspace oder Service; keine Seitenbildung; unbekannt oder fremde Art ist `ODR-E7002`.
5. **Diagnose:** `Modus` und `Einstellungen` kommen als fertiger Wert aus dem Root; Allowlist von 18 Schlüsseln, jeder andere wird verworfen; die Zähler sind ein Eintrag je Code von `model.AlleCodes()` aus `Snapshot().Errors` (fehlend 0). `TelemetryReader` reicht; die Anwendung weicht von `SPEC-230` Nr. 8 und 9 nicht ab. Der Dienst prüft `ctx.Err()` einmal, weil kein Port der Diagnose einen Kontext nimmt.
6. **`ODR-E7012`:** ohne Verwender in dieser Version, die Zeile bleibt (`SPEC-236` Nr. 9).
7. bis 10. Tabelle, Pflicht-Mengen, Modell-Menge der Wege: oben; Raster: §6.
11. **Schnitt:** §4, *Größenmarke, Teil des Architect*.
12. **Kennungen:** ein Bedarf, `SPEC-236` (eingefügt hinter `SPEC-124`, fern von den Stellen der Zweige `slice-mock-zuordnung` und `slice-telemetry-log`); keine `ARC`, `ADR` oder `CO`. Geändert sind in `spezifikation.md`: `SPEC-236` (neu), die drei Zeilen und der Satz hinter der Tabelle von `SPEC-005`, der *Ungeprüft*-Satz von `SPEC-105`, die Zeile `SPEC-112` und eine Zeile oben in der Historie. `slice-telemetry-log` kann `SPEC-236` ebenfalls brauchen: der Zusammenstoß ist der des Integrationsschritts, dieser Slice behält die Nummer, bis der Merge entscheidet.
13. **Katalog:** kein neuer Fehlercode, Schlüssel oder Kommando, kein Eintrag in `harness/katalog-ausnahmen.txt`, kein Handbuch-Kapitel; das Binary bleibt unverändert (`SPEC-201` Nr. 3 und 11).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-telemetry-port` (`welle-betrieb`) liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, Seitenform und Allowlist, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand beim Start** (Planner, 2026-11-07): erfüllt. `slice-telemetry-port` liegt in `done/` und liefert `TelemetryReader` mit `Snapshot().Errors` (ein Eintrag je Code der Tabelle von `SPEC-109` Nr. 2, auch mit 0; `SPEC-230` Nr. 9) und `FakeTelemetry`; `slice-telemetry-log` und `slice-telemetry-spans` sind nicht Voraussetzung. Der Slice berührt keine Datei der Gate-Kette (Übergangsregeln Nr. 7: weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` noch den Stempel); in `in-progress/` liegen `slice-telemetry-log` (Linie D, allein `src/adapters/driven/telemetry`), `slice-mock-zuordnung` (Linie B, `src/adapters/driven/odata/**`, `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/mock.go`), beide keine Gate-Kette. Er berührt keine Datei der Zweige: `src/hexagon/ports/inbound/` nur mit einer neuen Datei `inspection.go`, `src/hexagon/services/inspection/` neu; `src/hexagon/ports/outbound/` und `src/adapters/driven/sqlite/` nur, wenn der Architect die Seitenabfrage entscheidet (kein anderer Zweig berührt sie). Gemeinsam bleiben nur `spec/spezifikation.md` (Aufträge 12 in §3) und die Pläne. Ändert Vertrag oder Code eine Datei der Gate-Kette, `src/bootstrap/` oder der Linie B (auch `.a-check.yml`, `.golangci.yml`), geht das als Rückgabe an den Planner (§6). Zweig `slice-inspection-use-case`, Worktree `/Development/odata-recorder-wt-inspection-use-case`.

**Warum, und für welchen Teil:** nur der Teil *Diagnose* liest Zähler je Fehlercode über `TelemetryReader` (`src/hexagon/ports/support/`, geliefert von `slice-telemetry-port`): *Code-Abhängigkeit* für diesen Teil. Die Teile *Exchanges* und *Coverage* lesen `RecordingRepository`, `SessionRepository` und `CoverageRepository`, die seit `welle-persistenz` in `src/hexagon/ports/outbound/` liegen, und brauchen nichts davon.

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`** (Mock, `WorkspaceUseCase`, HTTP-Adapter, Composition Root): der Dienst kennt weder HTTP noch einen Modus. Starteten Exchanges und Coverage vor dem Port, wäre das ein Schnitt der Diagnose in einen eigenen Slice (Umplanung durch den Planner); dieser Plan hält den Slice ganz und nennt den Port als Trigger. Ob `slice-telemetry-port` selbst vor dem Mock starten darf, entscheidet die Planung von `welle-betrieb`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.

**Größenmarke** (Stand beim Start, Planner): Der Block mit `basis`, `schaetzung` und `marke` steht auf `main` nicht (ein Block ohne diese Werte wäre ein Platzhalter). Rollenfolge: Der Architect legt den Block im Zweig an (`pfad`, `produktcode`, `fakes`; `basis` ist der Commit, der Vertrag, Tabelle und Raster abschließt), der Planner setzt `basis`, `schaetzung` und `marke` in eigenem Commit vor dem ersten Code-Commit nach `SPEC-128` Nr. 1 (Schätzung *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke *M* = 100·⌈135·*S*/10000⌉; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Die Schnittfrage (*S* ≥ 2223) stellt der Planner nach der Tabelle; zur Frage an den Architect siehe Auftrag 11 in §3.

**Größenmarke, Teil des Architect** (2026-11-07): Der Block trägt `pfad`, `produktcode` und `fakes`; `basis`, `schaetzung` und `marke` setzt der Planner in eigenem Commit vor dem ersten Code-Commit. `basis` ist der Commit, der Vertrag, Tabelle und Raster abschließt. Der Block gilt für den ganzen Plan (alle Zeilen der Tabelle in §3): `produktcode` 390 (Port etwa 105 Zeilen, Dienst etwa 285), `fakes` 2 (`FakeRecording` und `FakeCoverage`; `FakeTelemetry` besteht seit `slice-telemetry-port`). `src/hexagon/ports/outbound/` und `src/adapters/driven/sqlite/` fehlen unter `pfad`, weil der Port unverändert bleibt (`SPEC-236` Nr. 4).

**Schnitt** (Auftrag 11, `SPEC-128` Nr. 1): Die Rechnung des Architect aus der Tabelle in §3 ergibt für den ganzen Plan 38 Zeilen, davon 2 nebenläufig (W2, W4) und 36 andere, *S* = 85·2 + 80·36 + 200·2 + 390 = 3840, *M* = 5200; *S* liegt über der Schnittgrenze (2223). Der Schnitt, vor dem Code, ohne geteilten Mechanismus außer dem Gerüst:

| Teil | Zeilen der Tabelle | Inhalt | *S* | *M* |
|---|---|---|---|---|
| 1 | K1, A1, X1 bis X13, N1, W1, W2 (18; 1 nebenläufig, 17 andere) | Port `InspectionUseCase` mit `Exchanges`, Dienst, `Konfig` mit `Exchanges`, Seitenbildung, Allowlist der Felder; `FakeRecording`; *produktcode* etwa 200, *fakes* 1 | 85 + 1360 + 200 + 200 = 1845 | 2500 |
| 2 | K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3, W4 (20; 1 nebenläufig, 19 andere) | `Coverage` und `Diagnose` als zusätzliche Methoden des Ports, zusätzliche Felder von `Konfig`, `Erlaubt`; `FakeCoverage`; *produktcode* etwa 190, *fakes* 1 | 85 + 1520 + 200 + 190 = 1995 | 2700 |

Beide Teile liegen unter der Schnittgrenze. Geteilt ist allein das Gerüst (Paket, `Konfig`, `New`, `mitCode`, etwa 30 Zeilen): Teil 1 legt es an, Teil 2 erweitert `Konfig` und den Port um Felder und Methoden, ohne bestehende zu ändern (die Signatur von `New` bleibt), Teil 2 folgt Teil 1. Der Planner legt Teil 2 als eigenen Slice an (Vorschlag `slice-inspection-coverage-diagnose`, Welle `welle-admin-api`, Start-Trigger: Teil 1 in `done/`), verschiebt die Zeilen von Teil 2 aus §3 dorthin und gibt jedem Plan seinen Block (Teil 1: `produktcode` 200, `fakes` 1); `slice-admin-sessions` hängt dann an beiden. Eine Rückführung ist das nicht, der Schnitt geschieht vor dem Code. Schneidet der Planner nicht, braucht der Block ein `schnitt-grund`; der Architect nennt keinen, denn die Operationen teilen nichts außer dem Gerüst.

```groessenmarke
pfad: src/hexagon/ports/inbound/inspection.go
pfad: src/hexagon/services/inspection/
produktcode: 390
fakes: 2
```

- `in-progress` → `open` (blockiert — Carveout?): `TelemetryReader` reicht für die Zähler der Diagnose nicht (Form oder Fehlerfall, Auftrag 5; Rückgabe an den Planner, der Port gehört `slice-telemetry-port` in `done/`), (eine Seitenabfrage im Port gibt es nicht: `SPEC-236` Nr. 4).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07): jede Zelle nennt ihren Ort als Nummer von `SPEC-236` oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle an, die der Code anders entscheiden müsste (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Exchanges einer Session lesen (Session, Seitenparameter) | `ctx` nil: `ErrUngueltig` ohne Code (Nr. 2); `Seite{}`: Anfang, Standard-`Limit` (Nr. 3); `sessionID` leer: unverändert an den Port, `ODR-E7003` (Nr. 4) | Session ohne Exchange und `Nach` hinter der letzten: leere Seite, `Weiter` 0 (Nr. 4); `Limit` 0: `StandardLimit` (Nr. 3) | unbekannte Session: `ODR-E7003` (Nr. 4, 8); `Limit` über `MaxLimit`: gekappt (Nr. 3); `Nach` oder `Limit` negativ: `ErrUngueltig` ohne Code (Nr. 3); Feld außerhalb der Allowlist: erscheint nie (Nr. 5) | vor dem Aufruf des Ports: sein Kontextfehler unverändert; nach der Rückkehr: Ergebnis, keine Prüfung (Nr. 2) | zustandslos; gleiche Argumente, gleiches Ergebnis; Anhängen zwischen zwei Seiten verschiebt nichts (Nr. 4, 10) |
| Coverage lesen (Baseline) | `ctx` nil: `ErrUngueltig` (Nr. 2); es gibt keinen Standard und kein „alle Baselines“ (Nr. 6) | `baselineID` leer: `ErrUngueltig` ohne Port (Nr. 6); Baseline ohne Einträge: leere Liste (Nr. 6) | unbekannte Baseline und Kennung eines Snapshots anderer Art: `ODR-E7002` (Nr. 6, 8); Status oder Evidenz: unverändert (Nr. 6) | wie Exchanges (Nr. 2) | zustandslos; eine Baseline ist unveränderlich (`SPEC-107`), das Ergebnis gleich (Nr. 10) |
| Diagnose lesen (Konfiguration, Modus, Zähler) | `ctx` nil: `ErrUngueltig` (Nr. 2); `Einstellungen` nil: `{}` (Nr. 1, 7); Code ohne Zähler: `Anzahl` 0 (Nr. 7) | `Errors` leer oder nil: jeder Code mit 0 (Nr. 7) | Schlüssel außerhalb der Allowlist: verworfen (Nr. 7); Eintrag außerhalb der Tabelle: verworfen (Nr. 7); ungültiger Modus: `New` scheitert (Nr. 1) | beendeter `ctx` vor `Snapshot`: sein Fehler, kein `Snapshot` (Nr. 2, 7); nach `Snapshot`: gilt nicht, es folgt kein blockierender Schritt | jeder Aufruf liest den Stand neu, die Zähler wachsen (Nr. 7, 10) |

- Die Paginierung (Auftrag 2) und die Seitenabfrage in `RecordingRepository` und SQLite-Adapter — **Ausgang:** entfallen; entschieden in `SPEC-236` Nr. 3 und 4 (Zeiger `sequence_no`, Standard 100, Obergrenze 1000; der Port bleibt, der Dienst schneidet im Speicher); die Größe einer Session ist ein akzeptiertes Negativ mit Grund und Folge (`SPEC-236` Nr. 4 und *Ungeprüft*).
- Lesen der Zähler: `slice-telemetry-port` (`done/`) hat entschieden, dass `TelemetryReader` mit `Snapshot().Errors` liest, nicht `Telemetry` (`SPEC-230` Nr. 9, Sendung des Gebers an diesen Slice); der Nehmer trägt sie seit dem Start in Ziel, DoD, Auftrag 5 und §4 (`AGENTS.md` §3.13) — **Ausgang:** entfallen (die Frage nach einer Leseoperation ist entschieden, Form und Fehlerfall stehen in `SPEC-236` Nr. 7 und tragen).
- Aufträge 8 bis 13 (Tabelle, Pflicht-Mengen mit Achsen, Zweig-Menge mit Zählprüfung, Schnittfrage, Kennungen und Dateien, Katalog-Pflicht) — **Ausgang:** entfallen für den Architect (geliefert: §3 und `SPEC-236`); **weiter offen** für die Umsetzung der Zählprüfungen (W1, W3 und die Zählprüfungen der Mengen) und den Bericht je Befund mit Menge, Zahl der Elemente und einer Mutation an einem Element, das der Befund nicht nannte (`AGENTS.md` §3.17); Adresse: Implementer.
- Die Parameter von `GET /__admin/coverage` und der Code für eine unbekannte Baseline (Auftrag 4), `ODR-E7012` ohne Endpunkt (Auftrag 6) — **Ausgang:** entfallen; entschieden in `SPEC-236` Nr. 6, 8 und 9 und in `SPEC-005`.
- Geheimnisse in Exchanges und Diagnose: die Allowlist der Felder und der Schlüssel (`SPEC-236` Nr. 5 und 7) ist der Schutz, die Fänger sind X12 und D6; der Fänger über alle Artefakte ist der Canary-Test in `slice-test-security` — **Ausgang:** weiter offen bis zur Closure; Adresse: Implementer (X12, D6), danach Reviewer.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): die Fakes der Repositories und `FakeTelemetry` tragen jede Schnittstelle des Originals und zählen jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Schätzung des Architect liegt über der Schnittgrenze, der Schnitt ist benannt (§4, *Schnitt*); die Marken setzt der Planner — **Ausgang:** offen; Adresse: Planner (Rückgabe 1).

**Rückgaben an den Planner** (`AGENTS.md` §3.13; der Architect schreibt weder `open/`-Pläne anderer Slices noch `SPEC-109` oder `SPEC-230` noch Roadmap und Welle-Pläne — Ausgang jeweils *offen*, Adresse Planner):

1. **Schnitt und Marken:** Teil 2 (Coverage und Diagnose) als eigener Slice anlegen (Vorschlag `slice-inspection-coverage-diagnose`, Zeilen K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3, W4 aus §3 dorthin), `basis`, `schaetzung` und `marke` je Plan nach `SPEC-128` Nr. 1 setzen, Welle `welle-admin-api` und Roadmap nachziehen, `slice-admin-sessions` auf beide Teile als Trigger stellen. Berichtigung des Architect in §4 (*Stand beim Start*): `Snapshot().Errors` hat einen Eintrag je Code der Tabelle von `SPEC-109` Nr. 2 (37, mit `slice-mock-zuordnung` 38), nicht 33.
2. **`SPEC-109` Nr. 5, Absatz *Anwendung*:** der Satz zu `InspectionUseCase` fehlt (Gegenstand `Session` bei `Exchanges`, `Snapshot` bei `Coverage`, `SPEC-236` Nr. 8); `SPEC-109` ändert dieser Zweig nicht.
3. **`slice-admin-sessions`** (`open/`, ändert dieser Zweig nicht): §1 oder DoD tragen die Parameter `limit`, `after` und `baseline_id` (`SPEC-005`, `SPEC-236` Nr. 3, 4, 6) und die Antwort auf einen ungültigen Parameter (negativer Wert, nicht lesbare Zahl, fehlendes `baseline_id`): der Use Case liefert `model.ErrUngueltig` ohne Code, ein Code oder Status dafür steht in §4 der Spezifikation nicht; die Entscheidung gehört in den Vertrag von `slice-admin-sessions` (neuer Code samt Katalog und Handbuch dort, oder ein vorhandener).
4. **Zählen der Fehler** (`SPEC-230` Nr. 8): der Dienst erhält nie `Telemetry` und zählt seine Fehler mit Code (`ODR-E7003`, `ODR-E7002`) nicht; die Stelle, die sie ausliefert (der Admin-Adapter in `slice-admin-sessions` oder `slice-telemetry-einbau-*`), trägt das `CountError` je Fehler.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-07; Zähler = Zahl der Dateien in `evidence/`). Offen und einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (7×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure) für die Zweig-Menge (Auftrag 10); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (16×) für die Marke. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×) für die Fakes der Repositories und `FakeTelemetry`; `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (29×, `AGENTS.md` §3.17), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14) für die Mengen; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, beide `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.16, für „ohne Geheimnisse“). Mit diesem Slice erreicht kein offener Eintrag mit weniger als 3× die Schwelle.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
