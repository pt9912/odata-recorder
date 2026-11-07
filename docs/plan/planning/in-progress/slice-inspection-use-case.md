# Slice slice-inspection-use-case: InspectionUseCase, Exchanges einer Session lesen (Teil 1)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Exchanges) · `SPEC-236` (Vertrag, neu; Nr. 1 bis 5, 8, 10 für diesen Teil) · `SPEC-005` (Zeile `GET …/exchanges` mit Parametern) · `SPEC-105` (*Ungeprüft*-Satz), `SPEC-107` (Repositories, aus denen gelesen wird) · `SPEC-109` (Fehlerabbildung, Anwendung in `SPEC-236` Nr. 8) · `SPEC-071`, `SPEC-112` (Zeile ergänzt) · `ARC-002`, `ARC-003` · `architecture.md` §2.2 (Zeile `InspectionUseCase`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-inspection-use-case)

**Autor:** pt9912. **Datum:** 2026-10-30.

**Schnitt:** Dieser Plan ist Teil 1 des Schnitts vor dem Code (Architect und Planner, 2026-11-07, `SPEC-128` Nr. 1). Teil 2, Coverage und Diagnose, ist [slice-inspection-coverage-diagnose](../open/slice-inspection-coverage-diagnose.md) (Start nach diesem Slice in `done/`); die Zeilen K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3 und W4 der Tabelle stehen dort.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `InspectionUseCase` (Inbound Port in `src/hexagon/ports/inbound`, Datei `inspection.go`, Dienst in `src/hexagon/services/inspection`) liest die Exchanges einer Session paginiert und ohne Geheimnisse; er verändert nichts und kennt kein HTTP. Der Slice legt das Gerüst an (Paket, `Konfig` mit dem Feld `Exchanges`, `New`, `mitCode`), auf dem `slice-inspection-coverage-diagnose` additiv aufbaut.

**Übernimmt:** das Doku-Update zu `SPEC-109` Nr. 5, Absatz *Anwendung*, soweit es `Exchanges` betrifft (Rückgabe 2 des Architect): der Satz zu `InspectionUseCase` mit dem Gegenstand `Session` bei `Exchanges` (`SPEC-236` Nr. 8) fehlt; der Zweig ändert `SPEC-109` nicht (`slice-mock-zuordnung`, Linie B, ändert dieselbe Stelle), der Planner trägt den Satz im Integrationsschritt (Merge-Commit) ein. Der Satz für `Coverage` (Gegenstand `Snapshot`) gehört zu `slice-inspection-coverage-diagnose`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Endpunkte, Pfade und Statuscodes — `slice-admin-sessions`; der Use Case kennt kein HTTP, der Adapter bildet ab. Dort stehen auch die Parameter `limit` und `after` und die Antwort auf einen ungültigen Wert (der Use Case liefert `model.ErrUngueltig` ohne Code).
- Coverage und Diagnose — `slice-inspection-coverage-diagnose` (Teil 2 des Schnitts): die Methoden `Coverage` und `Diagnose`, die Felder `Coverage`, `Zaehler`, `Modus` und `Einstellungen` von `Konfig`, `Erlaubt`, `FakeCoverage`, die Zeilen K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3 und W4. Der Slice liefert das Gerüst; Teil 2 erweitert Port und `Konfig` additiv, die Signatur von `New` bleibt.
- Anlegen, Beenden und Löschen von Sessions — `SessionUseCase` (`done/`); `InspectionUseCase` liest nur.
- Die Ports `Telemetry` und `TelemetryReader` und die Zähler — `welle-betrieb` (`slice-telemetry-port`, `done/`); dieser Teil braucht sie nicht (Zähler der Diagnose: Teil 2). Das Zählen der Fehler mit Code (`ODR-E7003`) übernimmt nicht der Dienst, sondern der Rahmen des Admin-Listeners (`slice-admin-listener-token`, Sendung vom 2026-11-07; `SPEC-230` Nr. 8).
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
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist; dazu `SPEC-109` Nr. 5, Absatz *Anwendung*: der Satz zu `InspectionUseCase` mit dem Gegenstand `Session` bei `Exchanges`, eingetragen im Integrationsschritt durch den Planner, nicht im Zweig (§1 *Übernimmt*).
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
| `src/hexagon/ports/inbound/inspection.go` | neu | `InspectionUseCase` mit `Exchanges` und den Typen der Seite und der Zeile |
| `src/hexagon/services/inspection/` | neu | Lesen der Exchanges; Seitenbildung, Allowlist der Felder; Gerüst (`Konfig`, `New`, `mitCode`) |
| `src/hexagon/ports/outbound/`, `src/adapters/driven/sqlite/` | unverändert | Entscheidung des Architect (2026-11-07): `ListBySession` bleibt, der Dienst schneidet die Seite im Speicher (`SPEC-236` Nr. 4); beide Pfade liegen nicht im Block |
| `spec/spezifikation.md` | update (Architect, erledigt) | `SPEC-236` mit Seitenform, Feldern der Exchanges, Parametern der Coverage und Allowlist der Diagnose (der Vertrag gilt für beide Teile); `SPEC-005`, `SPEC-105` (*Ungeprüft*-Satz), `SPEC-112`, Historie |
| Tests | neu | Fälle je Operation; `FakeRecording` hält jedes Argument fest und zählt jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service |

Der Vertrag steht als `SPEC-236` in der Spezifikation (Architect, 2026-11-07), die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) hier und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind entschieden, der Implementer entscheidet keine im Code.
Die Spalte `Nr.` trägt den Buchstaben der Art: `K` Konfiguration, `A` Kontext nil, `X` Exchanges, `N` akzeptiertes Negativ, `W` Zählprüfung der Stellen und Nebenläufigkeit. Die Nummern bleiben die des ungeschnittenen Plans; die Zeilen K2 bis K4, A2, C, D, N2, W3 und W4 trägt `slice-inspection-coverage-diagnose`. Die Spalte *Fänger* nennt Fälle und die Mengen des Abschnitts *Pflicht-Mengen* darunter; die Namen der Tests wählt der Implementer. Jede Zeile ist ein Weg des Dienstes (*Modell-Menge der Wege* unten); ein Test, der die Wege im Quelltext zählt, hält Zeilen und Wege gleich (W1).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| K1 | `New` weist ein nil in `Exchanges` ab (`SPEC-236` Nr. 1) | Fall: sonst gültige `Konfig`, `Exchanges` nil: `model.ErrUngueltig`, kein Dienst | Prüfung entfernt: `New` liefert einen Dienst | andere |
| A1 | `ctx` nil bei `Exchanges` ist `ErrUngueltig` ohne `Fehler`-Wert, kein Port wird gerufen, auch zusammen mit einem ungültigen Seitenparameter (Nr. 2, 3) | Schleife über Seite {gültig, `Nach` -1, `Limit` -1}: `errors.As(*Fehler)` falsch; `FakeRecording` zählt jede Operation: 0 | Prüfung entfernt: der Port erhält ein nil-`ctx` | andere |
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
| N1 | Akzeptierte Negative dieses Teils: die Größe einer Session (Schnitt im Speicher) und ein Geheimnis im Pfad (`SPEC-236` *Ungeprüft*) | kein Fänger; Grund und Folge in `SPEC-236` Nr. 4 und 5; die Gegenprobe in X12 hält den Pfad als sichtbar fest | – | andere |
| W1 | Zählprüfung: die Wege von `Exchanges` und der Anteil von `New` (K1) im Quelltext sind die Zeilen dieser Tabelle (Aufrufe der Ports, Aufrufe der Abbildung, Stellen eines Argumentverstoßes) | Test mit `go/parser` über die Dateien des Pakets ohne `_test`: Zahl der Aufrufe über Felder von `Konfig`, der Aufrufe von `mitCode` und der Stellen, die `ErrUngueltig` bilden, gleich der Zahl der Wege der Tabelle des Tests (Name, Art, Nr. dieser Tabelle); `slice-inspection-coverage-diagnose` erweitert die Tabelle des Tests um seine Wege im selben Commit (`AGENTS.md` §3.9) | ein weiterer Port-Aufruf, eine weitere Stelle `ErrUngueltig` ohne Zeile | andere |
| W2 | Gleichzeitige Aufrufe von `Exchanges` stören sich nicht, der Dienst hält keinen Zustand (Nr. 10) | Schleife über die Zweig-Menge W2 (jeder Weg von `Exchanges`, 41 Zweige: Erfolg, Weiter, Port-Fehler je Eintrag von `outbound.EintraegeAbbildung()` und ein anderer, Kappung, Limit 0, Limit unter 0, Nach unter 0, `ctx` nil, 5 Kontextfehler-Formen), je Zweig 8 Goroutinen mit je 20 Runden gegen einen Dienst; jedes Ergebnis gleich der Erwartung des Zweigs; Zählprüfung der Zweige gegen die Wege-Menge (`TestZweigeSindDieWege`); `make test-race` je Prozess (`-test.count=1`) | ein Schreibzugriff auf eine geteilte Variable je Zweig (15 Mutationen): Detektor rot in 20 von 20 Prozessläufen je Mutation (Bericht, §3) | nebenläufig |

**Pflicht-Mengen** (`AGENTS.md` §3.15, §3.17, `implement-slice` Schritt 14): je Menge Achsen, Zahl der Elemente als Produkt, Generator, Orakel (nie der Dienst und nie sein Code), Zählprüfung gegen die Quelle der Menge (keine feste Zahl im Test), Prüfgrenze der unbeschränkten Menge. Die Mengen M-Coverage, M-Codes und M-Schlüssel führt `slice-inspection-coverage-diagnose`.

| Menge | Achsen | Elemente | Orakel | Zählprüfung gegen die Quelle | Prüfgrenze (akzeptiertes Negativ) |
|---|---|---|---|---|---|
| M-Seite (X1 bis X5) | `Limit` {0, 1, `StandardLimit`, `MaxLimit` - 1, `MaxLimit`, `MaxLimit` + 1, größter `int`} × Bestand n {0, 1, L - 1, L, L + 1, 2L, 2L + 1} mit L gleich dem wirksamen `Limit` × `Nach` {0, 1, n - 1, n, n + 1, n + L}; ungültig: `Nach` und `Limit` {-1, -2, kleinster `int`} | 7 × 7 × 6 = 294 vor dem Entfernen gleicher Fälle (L gleich 1 macht n - 1 gleich 0), der Test zählt die erzeugten nach der Formel | Modell der Seiten: der Test bildet die Exchanges 1 bis n und schneidet in einer eigenen Schleife (`sequence_no` größer `Nach`, die ersten L), `Weiter` gleich der Nummer des letzten, wenn sie kleiner n ist; kein Aufruf des Dienstes | Konstanten `StandardLimit` und `MaxLimit` des Pakets, nie 100 oder 1000; Zahl der Fälle gleich der Formel aus den Längen der Achsen | `Limit` und n sind unbeschränkt: die Grenze ist `MaxLimit` und ein Bestand bis 2 · `MaxLimit` + 1; Folge: ein Fehler erst bei Bestand über 2 · `MaxLimit` + 1 bliebe ungefangen, der Dienst rechnet ohne Fallunterscheidung nach Größe |
| M-Bestand (X6) | Anhängen k {0, 1, L, L + 1} × Zeitpunkt {nach der ersten, nach der letzten Seite} × n {0, 1, L, L + 1, 2L + 1} | 4 × 2 × 5 = 40 | Folge der Seiten gelesen bis `Weiter` 0; erwartet jede `sequence_no` von 1 bis n + k genau einmal in Reihenfolge, soweit die Folge sie erreicht (nach der letzten Seite angehängt: nicht erreicht) | Längen der Achsen | Anhängen während des Aufrufs von `ListBySession` gilt nicht (ein Aufruf liest einen Stand, `SPEC-105` Nr. 13, `SPEC-236` Nr. 10) |
| M-Feld (X11, X12) | enthaltene Felder (X11); Einbringungsorte der ausgeschlossenen Felder: `SessionID`, `RawQuery`, `MatchKey`, Name und Wert in `RequestHeaders`, Name und Wert in `ResponseHeaders`, `RequestBlob`, `ResponseBlob` × Kodierung {`encoding/json`, `%+v`, Gang über alle Zeichenketten} (X12) | X11: Zahl der enthaltenen Felder; X12: 9 Orte × 3 = 27 | der Canary und der Wert des Feldes stehen im Test als Bytewerte; erwartet: Gleichheit des Wertes (X11), Abwesenheit des Canary (X12) | per Reflexion: enthaltene + ausgeschlossene Felder gleich den Feldern von `model.Exchange`, ausgeschlossene Orte gleich den ausgeschlossenen Feldern (Header zählen als ein Feld mit zwei Orten) | Kodierungen des Canary sind unbeschränkt (Hash, Teilstück, Umkodierung): drei geprüft; Folge: ein Dienst, der ein ausgeschlossenes Feld umkodiert in die Zeile setzte, bliebe ungefangen; der Dienst kopiert Felder und kodiert nichts. Body und Query als Art: der Dienst hat keinen Blob-Port und kein Parsen, ein Body kommt nur als Blob-Referenz (Ort `RequestBlob`, `ResponseBlob`), eine Query nur als `RawQuery` |
| M-Fehler (X8 bis X10) | Aufrufstelle `ListBySession` × Fehler {jeder Eintrag von `outbound.EintraegeAbbildung()`, ein anderer Fehler}; `ListBySession` × Form des Kontextfehlers {`Canceled`, `DeadlineExceeded`, verpackt, Join vorn, Join hinten}; `ListBySession` × `ctx`-Zustand {lebend, vor dem Aufruf abgebrochen, vom Fake abgebrochen mit Daten} | (Zahl der Einträge + 1) + 5 + 3 | `outbound.Abbilden` ist die Funktion unter Test in `SPEC-109`; das Orakel hier ist die Tabelle des Tests (Name, Wert, erwarteter Code je Gegenstand), die aus `SPEC-109` Nr. 5 abgeschrieben ist, nicht der Aufruf von `Abbilden` | die Tabelle des Tests hat genau die Namen von `outbound.EintraegeAbbildung()` (Quelle), nie eine feste Zahl | keine für die Fehler; die Form der Kontextfehler ist unbeschränkt (jede Verpackung): fünf geprüft, Folge: eine Verpackung, die `errors.Is` nicht trifft, ist ein Fehler des Ports. Die Aufrufstellen `List` und `Diagnose` führt `slice-inspection-coverage-diagnose` |

**Modell-Menge der Wege** (Auftrag 10; der Test W1 zählt sie gegen den Quelltext): jeder Verwurf nach Grund, jede Seitengrenze, jede Fehlerabbildung je Port, jedes Feld der Allowlist.

| Operation | Wege | Zeilen |
|---|---|---|
| `New` (Anteil dieses Teils) | nil `Exchanges`; Erfolg | K1 |
| `Exchanges` | `ctx` nil; `Nach` < 0; `Limit` < 0; Limit 0; Limit über der Grenze; Limit sonst; `ListBySession`: Kontextfehler, Fehler mit Code (je Eintrag der Tabelle), Erfolg; Filter `sequence_no` > `Nach`; Schnitt auf `Limit`; `Weiter` gesetzt oder 0; Feld je Eintrag der Allowlist | A1, X1 bis X13 |

Jeder Weg läuft in einem Test des Pakets und damit unter `make test-race`; ein Weg ohne Test ist nach W1 rot. Ein benanntes Negativ gibt es nur für die Fehler des Ports, die der echte Adapter nie liefert (etwa `ErrBlobFehlt` an `ListBySession`): der Fake liefert sie, der Weg ist derselbe wie für jeden Fehler (`mitCode`), Folge keine.

**Bericht des Implementers** (Handoff an den Reviewer; `AGENTS.md` §3.9, §3.10, §3.15, §3.17; Stand: Dienst, Port und Tests liegen im Zweig, die Fänger heißen wie folgt):

| Zeile | Fänger (Tests in `src/hexagon/services/inspection/`) |
|---|---|
| K1, A1 | `TestNewExchangesNil`, `TestKontextNil` |
| X1 | `TestSeitenparameterNegativ` |
| X2, X3, X4 | `TestSeitenModell` (Menge M-Seite), `TestSeitenGoldens` (Zahlen statt Modell), `TestMengeSeiteZaehlung` (Zählprüfung der Achsen), `TestGrenzenDesVertrags` (Standard 100 und Obergrenze 1000 als Literale des Tests: Limit 0, 1, 99, 100, 101, 500, 999, 1000, 1001, 1002, 1100, größter `int` bei Bestand 1100) |
| X3, X5 | `TestSeitenFolge` (Menge M-Seite, Folge ab jedem `Nach`; zweiter Aufruf gleich erstem), `TestSequenzMitLuecken` (Nummern mit Lücken: der Zeiger ist der Wert der `sequence_no`) |
| X6 | `TestAnhaengenZwischenSeiten` (Menge M-Bestand) |
| X7, X8, X9, X10 | `TestKennungUnveraendert`, `TestKontextfehlerUnveraendert`, `TestFehlerDesPortsMitCode`, `TestKontextDesAufrufs` |
| X11 | `TestAllowlistIstMengeDerFelder` (Reflexion), `TestFeldUnveraendert` (Schleife über die 15 enthaltenen Felder) |
| X12 | `TestKeinGeheimnis` (9 Orte × 3 Kodierungen, Gegenprobe Pfad) |
| X13 | `TestLesenSchreibtNichts` |
| W1, W2 | `TestWegeImQuelltext` (`go/parser`), `TestGleichzeitig` (Schleife über die Zweig-Menge W2), `TestZweigeSindDieWege` (Zählprüfung der Zweige) |

*Mutationen, je Zusage rot gesehen.* Ein Lauf des Werkzeugs `tools/harness/mutationslauf.sh` (`SPEC-210`, Klasse `andere`, ein Prozess je Mutation, 4g, `--cpus=4`) über 65 Mutationen: alle 65 `rot`, keine `rot-paket`, keine `gruen`, keine `uebersetzungsfehler`; Last des Hosts beim Lauf etwa 9,7 bei 4 CPUs (Messender: der Implementer; Last: andere Läufe auf dem Host). Der Fänger ist der benannte Test, der Rest des Pakets lief mit (Stufe 2 nur bei grün). Nach dem Review lief die Liste mit den 30 neuen Mutationen (15 Klasse `andere`: `g1` bis `g6`, `l1`, `t1` bis `t4`, `z1` bis `z4`; 15 Klasse `nebenläufig`: `w-*`) noch einmal, 95 Mutationen: 80 der Klasse `andere` `rot` (zwei, `z3` und `z4`, waren zuerst `gruen`, weil `In` die monotone Ablesung der Fixture verwarf; nach der Änderung der Fixture in einem eigenen Lauf `rot`), 15 `flatternd` (nur mit Detektor rot, siehe Raten). Die Liste im Wortlaut steht am Ende dieses Abschnitts. `x2-null-nicht-standard` wird über ein `panic` des Dienstes rot (Index in `schneide` bei `Limit` 0, das der Vertrag nie liefert); alle anderen über eine Meldung des Tests.

| Zusage | Mutation (Namen der Liste) | roter Test |
|---|---|---|
| K1 | Prüfung entfernt (`k1-pruefung-weg`) | `TestNewExchangesNil` |
| A1 | Prüfung `ctx` nil entfernt (`a1-ctx-nil-pruefung-weg`) | `TestKontextNil` |
| X1 | `<` zu `<=` bei `Nach` und bei `Limit`; Prüfung von `Nach`, von `Limit` entfernt; Text nennt einen Wert (`x1-*`, 5); der Text nennt den Wert -2 oder den kleinsten `int` bei `Nach` und bei `Limit` (`t1` bis `t4`, über ein Map-Literal, ohne Import) | `TestSeitenparameterNegativ` |
| X2 | Kappung entfernt; Kappung erst ab `MaxLimit + 1`, schon ab `MaxLimit - 1`; Standard + 1; 0 nicht auf Standard; `Limit` der Antwort 0 (`x2-*`, 6); die Werte selbst: Standard 101, 99, 50, Obergrenze 1001, 1002, 999 (`g1` bis `g6`) | `TestSeitenModell`; die Werte `TestGrenzenDesVertrags` |
| X3 | `>=` statt `>` im Filter; `Weiter` aus dem ersten Eintrag; Umkehr der Reihenfolge; `Eintraege` nil bei leerer Seite; `Gesamt` - 1 (`x3-*`, 5) | `TestSeitenModell` |
| X4 | `Weiter` gesetzt, wenn die Seite genau bis zum Ende reicht; Fehler bei `Nach` über `Gesamt` (`x4-*`, 2) | `TestSeitenModell` |
| X5 | `Weiter` um 1 zu klein (Doppelung), um 1 zu groß (Lücke) (`x5-*`, 2) | `TestSeitenFolge` |
| X3, X5 | `Weiter` als `Nach` plus Zahl der Einträge statt `sequence_no` des letzten (`l1-weiter-nach-plus-len`) | `TestSequenzMitLuecken` |
| X6 | Liste der ersten Antwort im Dienst gepuffert (Zustand zwischen Aufrufen; Werkzeug nicht möglich, weil zwei Textstellen, Einzellauf in Kopie `mA`) | `TestAnhaengenZwischenSeiten`: rot in 5 von 5 Prozessen ohne und 5 von 5 mit Detektor |
| X7 | Kennung verlängert; Leerzeichen-Kennung zur leeren; zweiter Aufruf von `ListBySession` (`x7-*`, 3) | `TestKennungUnveraendert` |
| X8 | Kontextfehler bekommt einen Code (`x8-kontextfehler-mit-code`) | `TestKontextfehlerUnveraendert` |
| X9 | Gegenstand `Snapshot`; Status 0; Status des letzten statt des ersten; Ursache nil (`x9-*`, 4) | `TestFehlerDesPortsMitCode` |
| X10 | `context.Background()` an den Port; `ctx.Err()` nach dem Aufruf geprüft (`x10-*`, 2) | `TestKontextDesAufrufs` |
| X11 | je Feld der Allowlist ungesetzt (15, `f01` bis `f15`); `ID` aus `Method`; `Method` und `Path`, `StartedAt` und `FinishedAt`, die Bodies, die Redaction-Zustände getauscht; `SequenceNo` aus `ResponseStatus` (`v1` bis `v6`); `StartedAt` und `FinishedAt` nach UTC umgerechnet, die monotone Ablesung abgestreift (`z1` bis `z4`; die Fixture trägt beide Formen: lokale Zeit mit monotoner Ablesung und eine Zone ungleich UTC) | `TestFeldUnveraendert` |
| X12 | je ausgeschlossenes Feld in den Pfad kopiert (7: `SessionID`, `RawQuery`, `MatchKey`, `RequestBlob`, `ResponseBlob`, `RequestHeaders` samt Namen und Werten, `ResponseHeaders` ebenso; `l1` bis `l7`) | `TestKeinGeheimnis` |
| X13 | `Get` je Aufruf; `Append` eines Vermerks (`x13-*`, 2) | `TestLesenSchreibtNichts` |
| W1 | ein weiterer Port-Aufruf, eine weitere Stelle `ErrUngueltig`, ein weiterer Aufruf von `mitCode`, je hinter `if false` (`w1-*`, 3) | `TestWegeImQuelltext` |
| W2 | je Zweig der Menge W2 ein Schreibzugriff auf eine geteilte Variable (15 Mutationen `w-*`: `ctx` nil, Nach unter 0, Limit unter 0, Port-Fehler in `Exchanges`, Erfolg, Kontextfehler in `mitCode`, Code in `mitCode`, Limit 0, Kappung, Limit sonst, Filter, Schnitt, Weiter gesetzt, Weiter 0, `zeile`) | `TestGleichzeitig`: mit Detektor rot in 20 von 20 Prozessen je Mutation (je `-test.count=1`), ohne Detektor 0 von 50 |
| W2 | Liste in einem Feld des Dienstes abgelegt und von dort gelesen (Einzellauf in Kopie `mB`) | `TestGleichzeitig`: mit Detektor 30 von 30, ohne 0 von 30; Ausgangsstand 0 von 30 mit und ohne |
| X6 | Cache der ersten gelesenen Liste (Einzellauf in Kopie `mA`) | `TestAnhaengenZwischenSeiten`: 5 von 5 ohne und 5 von 5 mit Detektor |

Raten: Messender der Implementer; Last des Hosts bei den Läufen der 15 Mutationen `w-*` 7,3 bis 15,1 bei 4 CPUs, bei den Einzelläufen 8,0 bis 9,9 (andere Rollen und `make gates` anderer Zweige liefen auf dem Host). Die 15 Mutationen `w-*` liefen im Werkzeug (Klasse `nebenläufig`, 20 Prozesse mit Detektor, 50 ohne): mit Detektor 20 von 20 je Mutation, ohne Detektor 0 von 50, Ergebnis `flatternd` (rot nur mit Detektor, wie bei einem Datenrennen ohne Wirkung). Der Ausgangsstand ist in 30 von 30 Prozessen mit und in 30 von 30 ohne Detektor grün. Ein Zustand mit Wirkung (der Cache einer Liste) fängt `TestAnhaengenZwischenSeiten` (X6), nicht mehr `TestGleichzeitig`; die Rate steht in der Tabelle oben. Es gibt kein Fenster zwischen Prüfen und Handeln (`SPEC-236` Nr. 10), also keine Zeile je Fenster. Abdeckung des Tests allein (`go test -race -run 'TestGleichzeitig|TestZweigeSindDieWege' -coverprofile`): 97,1 % der Anweisungen; `mitCode`, `Exchanges`, `wirksamesLimit`, `schneide` und `zeile` je 100,0 %, `New` 66,7 % (der Zweig mit nil läuft nicht gleichzeitig: `New` liest keinen Zustand und läuft nur im Aufbau). Der Rest des Pakets (alle Zweige nacheinander) 100,0 %.

*Mengen* (`implement-slice` Schritt 14, `AGENTS.md` §3.17): jede Menge ist eine Schleife, die Zählprüfung hält die Zahl gegen die Quelle der Menge, nie gegen eine feste Zahl; das Orakel ist nie der Dienst.

| Menge | Achsen und Elemente | Generator und Orakel | Zählprüfung | Mutation an einem Element je Achse, rot gesehen |
|---|---|---|---|---|
| M-Seite | `Limit` 7 × Bestand 7 × `Nach` 6 = 294, nach dem Entfernen gleicher Fälle und ungültiger Zeiger (`n - 1` bei `n` gleich 0) 236 | `seitenFaelle`; Orakel `modellSeite` (eigene Schleife, die Grenzen 100 und 1000 als Literale des Tests aus `SPEC-236` Nr. 3, nie aus dem Paket) | `TestMengeSeiteZaehlung`: Produkt der Längen der Achsen gleich der Zahl vor dem Entfernen; der Fall der leeren Session je `Limit` | `Limit`: Kappung erst über `MaxLimit + 1` (`x2-kappung-erst-ueber-max-plus-1`); Bestand: `Eintraege` nil bei n gleich 0 (`x3-eintraege-nil`); `Nach`: `Weiter` um 1 zu klein, die Folge ab `Nach` doppelt (`x5-weiter-minus-1`); Grenzwerte der Achse `Limit` (`TestGrenzenDesVertrags`, 12 Werte): Obergrenze 1002 (`g5-max-1002`, in der Fassung vor der Fix-Runde nicht genannt) |
| M-Lücken | Nummernfolgen 4 × Limit 4 × Zeiger je Folge (0, hinter der letzten, je Nummer davor, auf und hinter ihr) | `TestSequenzMitLuecken`; Orakel `modellListe` (eigene Schleife über die Nummern) | die Zahl der Fälle gleich dem Zähler der Schleifen; mindestens ein Fall | `l1-weiter-nach-plus-len`; Prüfgrenze: der Port liefert keine Lücken (`SPEC-105` Nr. 8), die Menge prüft allein, dass der Dienst nach dem Wert filtert |
| M-Zweig (W2) | 8 feste Zweige + 5 Formen des Kontextfehlers + 28 Fehler mit Code (27 Einträge der Abbildung und ein anderer) = 41; je Zweig 8 Goroutinen × 20 Runden | `zweige`, Erwartung `pruefe` je Art (Seite gegen `modellSeite`, ungültig, Kontextfehler gleicher Wert, Code) | `TestZweigeSindDieWege`: die Wege der Zweige gleich `wegeExchanges` (13 Wege), Fehler mit Code gleich `EintraegeAbbildung()` plus 1, Kontextfehler 5 | 15 Schreibzugriffe `w-*`, je einer in einem Zweig, darunter die im Review nicht genannten: Kontextfehler in `mitCode`, Weiter 0, Filter, Schnitt, Limit sonst, `zeile`; je 20 von 20 mit Detektor |
| M-Bestand | k 4 × Zeitpunkt 2 × n 5 = 40 | Schleife in `TestAnhaengenZwischenSeiten`; Orakel: `bereich(1, n + k)`, die Nachlese holt, was die Folge nicht erreicht | die Zahl der gelaufenen Fälle gleich dem Produkt der Längen der Achsen | nur eine, für alle drei Achsen: Cache der ersten Liste im Dienst, rot in 5 von 5 Prozessen; eine Mutation je Achse gibt es nicht, weil der Dienst jeden Stand neu liest und Zustand zwischen Aufrufen die einzige Art Fehler ist, die das Anhängen betrifft (akzeptiertes Negativ, Folge: ein Fehler allein bei k gleich `L + 1` bliebe ungefangen) |
| M-Feld | X11: 15 enthaltene Felder (Reflexion über `model.Exchange`); X12: 9 Orte × 3 Kodierungen = 27, in 3 Seiten je Ort | `TestFeldUnveraendert` (Orakel `erwartet`: Kopie nach Namen per Reflexion), `orte`, `kodierungen` | enthaltene + ausgeschlossene Felder gleich `NumField` von `model.Exchange`; Felder von `ExchangeZeile` gleich den enthaltenen mit demselben Typ; Orte gleich den ausgeschlossenen Feldern (Header mit zwei Orten) | Feld: alle 15 ungesetzt (`f01` bis `f15`); Ort: alle 7 ausgeschlossenen Felder (`l1` bis `l7`); Kodierung: keine (die Gegenprobe im Pfad hält alle drei als sichtbar fest) |
| M-Fehler | 27 Einträge von `outbound.EintraegeAbbildung()` + 1 anderer Fehler = 28; 5 Formen des Kontextfehlers; 3 Zustände von `ctx` | `tabelleSession` (abgeschrieben aus `SPEC-109` Nr. 5, nicht der Aufruf von `Abbilden`) | Namen der Tabelle des Tests gleich `EintraegeAbbildung()` plus dem anderen Fehler | Fehler: Gegenstand `Snapshot`; Kontextfehler: Code für den Kontextfehler; Zustand: `ctx.Err()` danach |

Prüfgrenzen (akzeptierte Negative, Grund und Folge im Plan oben, Tabelle *Pflicht-Mengen*): Bestand bis 2 · `MaxLimit` + 1; drei Kodierungen des Canary; fünf Formen des Kontextfehlers. Ein Text, der den Wert -2 oder den kleinsten `int` nennt, ist gefangen (`t1` bis `t4`); ein benanntes Negativ dazu gibt es nicht. Benannte Grenze von `TestWegeImQuelltext` (W1): er zählt Aufrufe über Felder von `Konfig`, Aufrufe von `mitCode` und Stellen `ErrUngueltig`; ein Port-Aufruf über eine lokale Variable und ein Aufruf von `Abbilden` ohne `mitCode` zählen nicht. Folge: ein solcher Weg ohne Zeile bliebe vom Zähler ungesehen; `Get` und `Append` fängt `TestLesenSchreibtNichts` zur Laufzeit, ein Fehler ohne Code bliebe nur durch `TestFehlerDesPortsMitCode` gefangen.

*Abdeckung* (Register `zweig-ohne-lauf-unter-dem-detektor`): `go test -race -cover` über das Paket: 100,0 % der Anweisungen, je Funktion 100,0 % (`mitCode`, `New`, `ctxNil`, `Exchanges`, `wirksamesLimit`, `schneide`, `zeile`); jeder Zweig läuft unter dem Detektor, und `TestGleichzeitig` führt ihn gleichzeitig aus (Abdeckung des Tests allein bei den Raten).

*Zeichen* (`AGENTS.md` §3.18): die geschriebenen Dateien `src/hexagon/ports/inbound/inspection.go` und `src/hexagon/services/inspection/*.go` sind auf Rohbytes gegengelesen (`LC_ALL=C grep -P` auf Backslash-u-Folge, BOM, NBSP, Nullbreiten- und Trennzeichen: kein Treffer); Werte jenseits von ASCII entstehen aus Bytewerten (`string([]byte{0xc3, 0xa4})`, der Canary aus `[]byte{0x4b, ...}`).

*Mutationsliste im Wortlaut* (reproduzierbar: eine Zeile je Mutation, `Kennung | Datei | alt => neu`; die Datei ist `inspektion` oder `fehler` im Paket des Dienstes, `⏎` ist ein Zeilenumbruch, Einrückung und die Anführungszeichen des JSON sind weggelassen, rechts von `=>` nichts heißt Löschung; `alt` kommt in der Datei genau einmal vor; die Zeilen werden in die Felder `alt` und `neu` einer Liste nach `SPEC-210` Nr. 2 übernommen, Test und Klasse stehen in der Tabelle oben):

```text
k1-pruefung-weg | inspektion | if k.Exchanges == nil { => if false {
a1-ctx-nil-pruefung-weg | inspektion | if ctx == nil {⏎return inbound.ExchangeSeite{}, ctxNil()⏎}⏎ =>
x1-nach-kleiner-gleich | inspektion | seite.Nach < 0 => seite.Nach <= 0
x1-limit-kleiner-gleich | inspektion | seite.Limit < 0 => seite.Limit <= 0
x1-nach-pruefung-weg | inspektion | if seite.Nach < 0 {⏎return inbound.ExchangeSeite{}, errors.Join(model.ErrUngueltig, errors.New("Nach ist negativ"))⏎}⏎ =>
x1-limit-pruefung-weg | inspektion | if seite.Limit < 0 {⏎return inbound.ExchangeSeite{}, errors.Join(model.ErrUngueltig, errors.New("Limit ist negativ"))⏎}⏎ =>
x1-text-nennt-wert | inspektion | errors.New("Nach ist negativ") => errors.New("Nach ist negativ: -1")
x2-kappung-weg | inspektion | case limit > MaxLimit:⏎return MaxLimit⏎ =>
x2-kappung-erst-ueber-max-plus-1 | inspektion | limit > MaxLimit: => limit > MaxLimit+1:
x2-kappung-ab-max-minus-1 | inspektion | limit > MaxLimit: => limit >= MaxLimit-1:
x2-standard-plus-1 | inspektion | return StandardLimit => return StandardLimit + 1
x2-null-nicht-standard | inspektion | case limit == 0:⏎return StandardLimit⏎ =>
x2-limit-antwort-null | inspektion | Limit: limit, => Limit: 0,
x3-filter-kleiner | inspektion | e.SequenceNo <= nach => e.SequenceNo < nach
x3-weiter-aus-erstem | inspektion | seite.Eintraege[len(seite.Eintraege)-1].SequenceNo => seite.Eintraege[0].SequenceNo
x3-reihenfolge-umgekehrt | inspektion | return schneide(alle, seite.Nach, wirksamesLimit(seite.Limit)), nil => for i, j := 0, len(alle)-1; i < j; i, j = i+1, j-1 {⏎alle[i], alle[j] = alle[j], alle[i]⏎}⏎return schneide(alle, seite.Nach, wirksamesLimit(seite.Limit)), nil
x3-eintraege-nil | inspektion | Eintraege: []inbound.ExchangeZeile{}, => Eintraege: nil,
x3-gesamt-minus-1 | inspektion | Gesamt: len(alle)} => Gesamt: len(alle) - 1}
x4-weiter-bei-gleich | inspektion | if rest { => if rest || len(seite.Eintraege) == limit {
x4-fehler-bei-nach-ueber-gesamt | inspektion | return schneide(alle, seite.Nach, wirksamesLimit(seite.Limit)), nil => if seite.Nach > len(alle) {⏎return inbound.ExchangeSeite{}, mitCode(outbound.ErrNichtGefunden, outbound.GegenstandSession)⏎}⏎return schneide(alle, seite.Nach, wirksamesLimit(seite.Limit)), nil
x5-weiter-minus-1 | inspektion | seite.Weiter = seite.Eintraege[len(seite.Eintraege)-1].SequenceNo => seite.Weiter = seite.Eintraege[len(seite.Eintraege)-1].SequenceNo - 1
x5-weiter-plus-1 | inspektion | seite.Weiter = seite.Eintraege[len(seite.Eintraege)-1].SequenceNo => seite.Weiter = seite.Eintraege[len(seite.Eintraege)-1].SequenceNo + 1
x7-kennung-verlaengert | inspektion | ListBySession(ctx, sessionID) => ListBySession(ctx, sessionID+"x")
x7-leerzeichen-id-leer | inspektion | alle, err := s.k.Exchanges.ListBySession(ctx, sessionID) => if sessionID == " " {⏎sessionID = ""⏎}⏎alle, err := s.k.Exchanges.ListBySession(ctx, sessionID)
x7-zweiter-aufruf | inspektion | alle, err := s.k.Exchanges.ListBySession(ctx, sessionID) => alle, err := s.k.Exchanges.ListBySession(ctx, sessionID)⏎if err == nil {⏎_, err = s.k.Exchanges.ListBySession(ctx, sessionID)⏎}
x13-get-je-aufruf | inspektion | alle, err := s.k.Exchanges.ListBySession(ctx, sessionID) => _, _ = s.k.Exchanges.Get(ctx, sessionID)⏎alle, err := s.k.Exchanges.ListBySession(ctx, sessionID)
x13-append-vermerk | inspektion | alle, err := s.k.Exchanges.ListBySession(ctx, sessionID) => _ = s.k.Exchanges.Append(ctx, model.Exchange{SessionID: sessionID})⏎alle, err := s.k.Exchanges.ListBySession(ctx, sessionID)
x8-kontextfehler-mit-code | fehler | if !ok {⏎return err⏎} => if !ok {⏎c, ok = outbound.Abbilden(outbound.ErrNichtGefunden, g)⏎}
x9-gegenstand-snapshot | inspektion | outbound.GegenstandSession => outbound.GegenstandSnapshot
x9-status-null | fehler | Status: c.Stati()[0], => Status: 0,
x9-status-letzter | fehler | Status: c.Stati()[0], => Status: c.Stati()[len(c.Stati())-1],
x9-ursache-nil | fehler | Ursache: err} => Ursache: nil}
x10-background-an-den-port | inspektion | ListBySession(ctx, sessionID) => ListBySession(context.Background(), sessionID)
x10-ctx-err-danach | inspektion | if err != nil {⏎return inbound.ExchangeSeite{}, mitCode(err, outbound.GegenstandSession)⏎}⏎ => if err == nil {⏎err = ctx.Err()⏎}⏎if err != nil {⏎return inbound.ExchangeSeite{}, mitCode(err, outbound.GegenstandSession)⏎}⏎
f01-id-ungesetzt | inspektion | ID: e.ID,  =>
f02-sequenceno-ungesetzt | inspektion | SequenceNo: e.SequenceNo,  =>
f03-method-ungesetzt | inspektion | Method: e.Method,  =>
f04-path-ungesetzt | inspektion | Path: e.Path, =>
f05-startedat-ungesetzt | inspektion | StartedAt: e.StartedAt,  =>
f06-finishedat-ungesetzt | inspektion | FinishedAt: e.FinishedAt,  =>
f07-keyredacted-ungesetzt | inspektion | KeyRedacted: e.KeyRedacted, =>
f08-requestbody-ungesetzt | inspektion | RequestBody: e.RequestBody,  =>
f09-responsebody-ungesetzt | inspektion | ResponseBody: e.ResponseBody,  =>
f10-responsestatus-ungesetzt | inspektion | ResponseStatus: e.ResponseStatus, =>
f11-capture-ungesetzt | inspektion | Capture: e.Capture,  =>
f12-requestredact-ungesetzt | inspektion | RequestRedact: e.RequestRedact,  =>
f13-responseredact-ungesetzt | inspektion | ResponseRedact: e.ResponseRedact, =>
f14-integrity-ungesetzt | inspektion | Integrity: e.Integrity,  =>
f15-equivalence-ungesetzt | inspektion | Equivalence: e.Equivalence, =>
v1-id-aus-method | inspektion | ID: e.ID,  => ID: e.Method, 
v2-method-path-getauscht | inspektion | Method: e.Method, Path: e.Path, => Method: e.Path, Path: e.Method,
v3-zeiten-getauscht | inspektion | StartedAt: e.StartedAt, FinishedAt: e.FinishedAt, => StartedAt: e.FinishedAt, FinishedAt: e.StartedAt,
v4-bodies-getauscht | inspektion | RequestBody: e.RequestBody, ResponseBody: e.ResponseBody, => RequestBody: e.ResponseBody, ResponseBody: e.RequestBody,
v5-redact-getauscht | inspektion | RequestRedact: e.RequestRedact, ResponseRedact: e.ResponseRedact, => RequestRedact: e.ResponseRedact, ResponseRedact: e.RequestRedact,
v6-sequenceno-aus-status | inspektion | SequenceNo: e.SequenceNo,  => SequenceNo: e.ResponseStatus, 
l1-sessionid-im-pfad | inspektion | Path: e.Path, => Path: e.Path + e.SessionID,
l2-rawquery-im-pfad | inspektion | Path: e.Path, => Path: e.Path + e.RawQuery,
l3-matchkey-im-pfad | inspektion | Path: e.Path, => Path: e.Path + e.MatchKey,
l4-requestblob-im-pfad | inspektion | Path: e.Path, => Path: e.Path + e.RequestBlob,
l5-responseblob-im-pfad | inspektion | Path: e.Path, => Path: e.Path + e.ResponseBlob,
l6-requestheaders-im-pfad | inspektion | Path: e.Path, => Path: func() string { s := e.Path; for k, v := range e.RequestHeaders { s += k; for _, x := range v { s += x } }; return s }(),
l7-responseheaders-im-pfad | inspektion | Path: e.Path, => Path: func() string { s := e.Path; for k, v := range e.ResponseHeaders { s += k; for _, x := range v { s += x } }; return s }(),
w1-weiterer-port-aufruf | inspektion | alle, err := s.k.Exchanges.ListBySession(ctx, sessionID) => if false {⏎_, _ = s.k.Exchanges.Get(ctx, sessionID)⏎}⏎alle, err := s.k.Exchanges.ListBySession(ctx, sessionID)
w1-weitere-stelle-ungueltig | inspektion | alle, err := s.k.Exchanges.ListBySession(ctx, sessionID) => if false {⏎return inbound.ExchangeSeite{}, model.ErrUngueltig⏎}⏎alle, err := s.k.Exchanges.ListBySession(ctx, sessionID)
w1-weiterer-mitcode | inspektion | alle, err := s.k.Exchanges.ListBySession(ctx, sessionID) => if false {⏎return inbound.ExchangeSeite{}, mitCode(nil, outbound.GegenstandSession)⏎}⏎alle, err := s.k.Exchanges.ListBySession(ctx, sessionID)
g1-standard-101 | inspektion | StandardLimit = 100 => StandardLimit = 101
g2-standard-50 | inspektion | StandardLimit = 100 => StandardLimit = 50
g3-max-1001 | inspektion | MaxLimit      = 1000 => MaxLimit      = 1001
g4-max-999 | inspektion | MaxLimit      = 1000 => MaxLimit      = 999
g5-max-1002 | inspektion | MaxLimit      = 1000 => MaxLimit      = 1002
g6-standard-99 | inspektion | StandardLimit = 100 => StandardLimit = 99
l1-weiter-nach-plus-len | inspektion | seite.Weiter = seite.Eintraege[len(seite.Eintraege)-1].SequenceNo => seite.Weiter = nach + len(seite.Eintraege)
t1-nach-text-minus-2 | inspektion | errors.New("Nach ist negativ") => errors.New("Nach ist negativ"+map[int]string{-2: " -2"}[seite.Nach])
t2-nach-text-minint | inspektion | errors.New("Nach ist negativ") => errors.New("Nach ist negativ"+map[int]string{-9223372036854775808: " -9223372036854775808"}[seite.Nach])
t3-limit-text-minus-2 | inspektion | errors.New("Limit ist negativ") => errors.New("Limit ist negativ"+map[int]string{-2: " -2"}[seite.Limit])
t4-limit-text-minint | inspektion | errors.New("Limit ist negativ") => errors.New("Limit ist negativ"+map[int]string{-9223372036854775808: " -9223372036854775808"}[seite.Limit])
z1-startedat-utc | inspektion | StartedAt: e.StartedAt,  => StartedAt: e.StartedAt.UTC(), 
z2-finishedat-utc | inspektion | FinishedAt: e.FinishedAt,  => FinishedAt: e.FinishedAt.UTC(), 
z3-startedat-round0 | inspektion | StartedAt: e.StartedAt,  => StartedAt: e.StartedAt.Round(0), 
z4-finishedat-round0 | inspektion | FinishedAt: e.FinishedAt,  => FinishedAt: e.FinishedAt.Round(0), 
w-* (15, Klasse nebenläufig, 20 Prozesse mit Detektor) | je Zweig eine Zeile `s.k = s.k` (ctx nil, Nach unter 0, Limit unter 0, Port-Fehler, Erfolg in Exchanges) oder `geteilt++` mit der Deklaration `var geteilt int` hinter der Funktion (Kontextfehler und Code in mitCode, Limit 0, Kappung, Limit sonst in wirksamesLimit; Filter, Schnitt, Weiter gesetzt, Weiter 0 in schneide; zeile) in dem Zweig
X6 Zustand: ein Feld puffer []model.Exchange in Inspektion; Exchanges liest ListBySession nur, wenn puffer nil ist, und legt das Ergebnis dort ab
W2 Datenrennen: dasselbe Feld, in jedem Aufruf geschrieben (s.puffer = alle) und von dort geschnitten
```

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts.
- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12; geliefert, sie stehen als Herkunft hier). Seit dem Schnitt gehören die Aufträge 4 (Coverage) und 5 (Diagnose) und die Achsen *Coverage* und *Diagnose* des Auftrags 9 zu `slice-inspection-coverage-diagnose`; hier gelten die übrigen:

1. **Vertrag von `InspectionUseCase`:** Operationen (Exchanges, Coverage, Diagnose), Argumente und Rückgabetypen, Fehlerabbildung als Anwendung von `SPEC-109`, Ort in der Spezifikation.
2. **Paginierung:** `SPEC-005` und die Architektur sagen „paginiert“, kein Vertrag nennt Parameter (Seitengröße, Zeiger oder Versatz), Standard und Obergrenze, Reihenfolge, Verhalten bei ungültigem oder zu großem Wert, Stabilität bei gleichzeitigem Anhängen, Gesamtzahl. `RecordingRepository.ListBySession` liefert alle Exchanges: erweitert der Architect den Port um eine Seitenabfrage (berührt `SPEC-105` und den SQLite-Adapter und damit die Größe) oder schneidet der Dienst im Speicher (Grenze für große Sessions)?
3. **Felder eines Exchange in der Liste:** Kopfdaten (Kennung, Sequenz, Methode, Pfad, Status, `capture_status`, Zeitpunkte, Vermerke der Redaction) und ob Header-Werte, Query und Bodies (oder nur Größe und Hash) erscheinen; „ohne Geheimnisse“ als Allowlist der Felder, nicht als Liste der entfernten.
4. **Coverage:** *(Teil 2)* `SPEC-005` nennt für `GET /__admin/coverage` keinen Parameter; Baseline-Kennung, Antwortform, Code einer unbekannten Baseline.
5. **Diagnose:** *(Teil 2)* Allowlist der Schlüssel der Konfigurationszusammenfassung, Form der Zähler je Fehlercode, Anwendung von `TelemetryReader.Snapshot().Errors`.
6. **`ODR-E7012`:** `SPEC-112` kennt „Exchange unbekannt“ (Admin), `SPEC-005` hat keinen Endpunkt für einen einzelnen Exchange: Zeile ergänzen oder den Code anders verorten.
7. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Aufträge an den Architect, Ergänzung beim Start** (Planner, 2026-11-07; vor dem ersten Code-Commit, `AGENTS.md` §3.12; geliefert):

8. **Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15) gegen den Vertrag aus Auftrag 1 bis 6: je Zusage, Randform und Fall eine Zeile, auch „Lesen schreibt nichts“ (Fake jedes Repositories zählt jede Operation, auch `Create`, `Update`, `Delete`, `Append`), „kein Geheimnis in Exchanges und Diagnose“ (Allaussage, Fänger über alle Felder der Allowlist, `BEO-REPO/universalaussage-im-vertrag-ungeprueft`, 14×), der `ctx` und der Kontextfehler je Aufrufstelle (Exchanges: Session und Exchanges; Coverage; Diagnose: Zusammenfassung und `Snapshot`), Gleichheit zweier verglichener Werte an jeder Grenze der Seite (Seitengröße gleich Obergrenze, Versatz gleich Zahl der Exchanges). `FakeTelemetry` ist `TelemetryReader` und `Telemetry`; der Dienst darf `Telemetry` nie rufen, der Fake zählt es.
9. **Pflicht-Mengen mit Achsen** (`AGENTS.md` §3.15, §3.17, `implement-slice` Schritt 14; Register `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, 29×, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`, 6×, `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze`, 11×): je Klasse die **Achsen**, die Zahl der Elemente als Produkt, der Generator, das **Orakel** (unabhängig vom Dienst: Modell der Seiten aus der Menge der Exchanges, nie der Dienst selbst) und die Prüfgrenze der unbeschränkten Menge als akzeptiertes Negativ mit Grund und Folge im Betrieb. Mindestens die Achsen **Seite** (Größe 0, 1, Standard, Obergrenze minus 1, Obergrenze, Obergrenze plus 1, negativ, sehr groß; Versatz oder Zeiger an jeder Grenze: 0, letzte, hinter der letzten, ungültig), **Bestand** (0, 1, genau eine Seite, eine Seite plus 1, mehrere Seiten; Anhängen zwischen zwei Seiten), **Feld** (jedes Feld der Allowlist der Exchange-Zeile; jede Art, in der ein Geheimnis eingebracht sein kann: Header, Query, Body, Pfad), **Coverage** *(Teil 2)*, **Diagnose** *(Teil 2)* und **Fehler** (jeder Fehler jedes Ports aus `SPEC-109` Nr. 5 an jeder Aufrufstelle, `ctx` abgebrochen vor und nach jedem Aufruf). Die Zählprüfung hält die Zahl der Elemente gegen die Menge der Quelle (`model.AlleCodes()`, die Menge der Statuswerte, die Felder des Typs), nie gegen eine feste Zahl im Test.
10. **Zweige und Rückgaben als Menge mit Zählprüfung** (Register `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, 7×, Schwelle erreicht): der Architect nennt die **Modell-Menge** der Wege (jeder Verwurf nach Grund, jede Seitengrenze, jede Fehlerabbildung je Port, jedes Feld der Allowlist) und die Tabelle führt je Element eine Zeile mit Fänger; der Test zählt die Menge gegen diese Modell-Menge (Zahl der Zweige im Code gleich Zahl der Zeilen), ein Zweig, den kein Lauf unter dem Detektor erreicht, ist ein benanntes Negativ mit Grund und Folge. Trägt der Vertrag eine Zusage über Nebenläufigkeit (Anhängen während des Lesens, Stand der Zähler), gilt `AGENTS.md` §3.10 je Fenster, sonst steht „gilt nicht“ mit Grund.
11. **Schnittfrage** nach `SPEC-128` Nr. 1: der Planner schätzt die Größe nach der Tabelle des Architect; über der Schnittgrenze wird vor dem Code geschnitten (Ausgang: §4, *Schnitt*).
12. **Kennungen und Dateien** (Linie D, Bereich `SPEC` 230 bis 259, `ARC` 040 bis 049, `ADR` 0040 bis 0049, `CO` 060 bis 079; letzte vergebene `SPEC` der Linie: `SPEC-235`): neue Nummern ab `SPEC-236`; `slice-telemetry-log` (Zweig) kann `SPEC-236` ebenfalls brauchen, der Zusammenstoß ist der des Integrationsschritts, der Architect nennt die Nummer im Plan. Der Slice ändert in `spezifikation.md` die Zeilen der Admin-API (`SPEC-005` ff.), `SPEC-105` und `SPEC-107`, soweit der Wortlaut es trägt, `SPEC-112` und hängt eine Zeile an die Änderungstabelle am Dateiende. Neben den Zweigen: `slice-mock-zuordnung` (Linie B) ändert `SPEC-160`, `SPEC-109` Nr. 5 und 6, `SPEC-174`, `SPEC-175`, `SPEC-195` bis `SPEC-199`, die Zeile `odr_errors_total` der Tabelle der Metriken und die Änderungstabelle; `slice-telemetry-log` (Linie D) ändert `SPEC-231`, `SPEC-232`, `SPEC-230` und die Änderungstabelle. Dieser Slice berührt `src/adapters/driven/odata/**`, `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/mock.go`, `src/adapters/driven/telemetry`, `src/hexagon/ports/support/` und `src/bootstrap/` nicht; `src/hexagon/ports/inbound/` nur um eine neue Datei `inspection.go`. Eine Änderung an `SPEC-109` oder `SPEC-230` geht als Rückgabe an den Planner.
13. **Katalog-Pflicht** (`SPEC-201` Nr. 3 und 11): erwartet ist kein Eintrag im Katalog des Binary (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`); bringt der Vertrag doch einen neuen Fehlercode, Schlüssel oder ein Kommando, steht er im selben Slice in `harness/katalog-ausnahmen.txt` und im Handbuch und in §4 des Plans, oder der Architect gibt die Frage an den Planner zurück (Rückführung `in-progress` → `open`, §4).

**Entscheidungen des Architect** (2026-11-07; Vertrag `SPEC-236`, Ausgang der Aufträge 1 bis 13; eine ADR entsteht nicht: die Entscheidungen ändern keinen Port und sind durch einen Folge-Vertrag umkehrbar, Gründe und Folgen stehen in `SPEC-236` wie in `SPEC-124`; die Entscheidungen 4 und 5 tragen den Teil 2):

1. **Vertrag:** `SPEC-236`, hinter `SPEC-124`: drei Operationen `Exchanges(ctx, sessionID, seite)`, `Coverage(ctx, baselineID)`, `Diagnose(ctx)`; Typen im Paket `inbound`, Dienst `inspection.Inspektion` mit `Konfig{Exchanges, Coverage, Zaehler, Modus, Einstellungen}`; Fehlerabbildung als Anwendung von `SPEC-109` Nr. 5 (Gegenstand `Session` bzw. `Snapshot`), ein Argumentverstoß ist `model.ErrUngueltig` ohne Code (wie `SPEC-124` Nr. 3). Dieser Teil liefert `Exchanges` und das Feld `Exchanges` von `Konfig`; die übrigen Operationen und Felder liefert Teil 2.
2. **Paginierung:** Zeiger ist die `sequence_no` (`Nach`, HTTP `after`), `Limit` (HTTP `limit`) mit Standard 100 und Obergrenze 1000; größere Werte werden gekappt, negative sind ungültig; die Antwort trägt `Gesamt` und `Weiter`; stabil beim Anhängen. **Der Port bleibt unverändert, der Dienst schneidet im Speicher** (`SPEC-236` Nr. 4): `SPEC-105` ändert nur den *Ungeprüft*-Satz, `src/hexagon/ports/outbound/` und `src/adapters/driven/sqlite/` liegen nicht im Block.
3. **Felder:** 15 Felder der Allowlist (`SPEC-236` Nr. 5), 7 ausgeschlossen (`SessionID`, `RawQuery`, `MatchKey`, beide Header-Abbildungen, beide Blob-Referenzen); der Pfad steht wie gespeichert (akzeptiertes Negativ, Folge in `SPEC-236`).
4. **Coverage:** *(Teil 2)* Parameter ist die Baseline-Kennung (HTTP `baseline_id`, Pflicht); keine Seitenbildung; unbekannt oder fremde Art ist `ODR-E7002`.
5. **Diagnose:** *(Teil 2)* `Modus` und `Einstellungen` kommen als fertiger Wert aus dem Root; Allowlist von 18 Schlüsseln; die Zähler sind ein Eintrag je Code von `model.AlleCodes()` aus `Snapshot().Errors`; `TelemetryReader` reicht.
6. **`ODR-E7012`:** ohne Verwender in dieser Version, die Zeile bleibt (`SPEC-236` Nr. 9).
7. bis 10. Tabelle, Pflicht-Mengen, Modell-Menge der Wege: oben; Raster: §6.
11. **Schnitt:** §4, *Schnitt*.
12. **Kennungen:** ein Bedarf, `SPEC-236` (eingefügt hinter `SPEC-124`, fern von den Stellen der Zweige `slice-mock-zuordnung` und `slice-telemetry-log`); keine `ARC`, `ADR` oder `CO`. Geändert sind in `spezifikation.md`: `SPEC-236` (neu), die drei Zeilen und der Satz hinter der Tabelle von `SPEC-005`, der *Ungeprüft*-Satz von `SPEC-105`, die Zeile `SPEC-112` und eine Zeile oben in der Historie. Zusammenstoß mit `slice-telemetry-log`: §6.
13. **Katalog:** kein neuer Fehlercode, Schlüssel oder Kommando, kein Eintrag in `harness/katalog-ausnahmen.txt`, kein Handbuch-Kapitel; das Binary bleibt unverändert (`SPEC-201` Nr. 3 und 11).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): keine Code-Abhängigkeit außer den Repositories aus `welle-persistenz` (`done/`). Die Spezifikation trägt den Vertrag samt Randformen, Seitenform und Allowlist, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand beim Start** (Planner, 2026-11-07): erfüllt. Der Slice berührt keine Datei der Gate-Kette (Übergangsregeln Nr. 7: weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` noch den Stempel); in `in-progress/` liegen `slice-telemetry-log` (Linie D, allein `src/adapters/driven/telemetry`), `slice-mock-zuordnung` (Linie B, `src/adapters/driven/odata/**`, `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/mock.go`), beide keine Gate-Kette. Er berührt keine Datei der Zweige: `src/hexagon/ports/inbound/` nur mit einer neuen Datei `inspection.go`, `src/hexagon/services/inspection/` neu; `src/hexagon/ports/outbound/` und `src/adapters/driven/sqlite/` bleiben unverändert. Gemeinsam bleiben nur `spec/spezifikation.md` (Aufträge 12 in §3) und die Pläne. Ändert Vertrag oder Code eine Datei der Gate-Kette, `src/bootstrap/` oder der Linie B (auch `.a-check.yml`, `.golangci.yml`), geht das als Rückgabe an den Planner (§6). Zweig `slice-inspection-use-case`, Worktree `/Development/odata-recorder-wt-inspection-use-case`.

**Warum:** Dieser Teil liest `RecordingRepository`, das seit `welle-persistenz` in `src/hexagon/ports/outbound/` liegt, und braucht keinen Slice von `welle-betrieb`. Die Zähler je Fehlercode über `TelemetryReader` (`slice-telemetry-port`, `done/`) braucht allein Teil 2, `slice-inspection-coverage-diagnose`, das nach diesem Slice startet (*Code-Abhängigkeit*: Gerüst, Port und `Konfig`).

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`** (Mock, `WorkspaceUseCase`, HTTP-Adapter, Composition Root): der Dienst kennt weder HTTP noch einen Modus. Der Schnitt der Diagnose in einen eigenen Slice ist vollzogen (Planner, 2026-11-07).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke dieses Abschnitts (Größenmarke unten), die der Planner vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` gesetzt hat (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): nicht erwartet; die Randformen sind entschieden (`SPEC-236`), eine Seitenabfrage im Port gibt es nicht (`SPEC-236` Nr. 4).

**Schnitt** (Auftrag 11, `SPEC-128` Nr. 1; Architect und Planner, vor dem Code, 2026-11-07): Die Rechnung des Architect aus der Tabelle des ungeschnittenen Plans ergab 38 Zeilen, davon 2 nebenläufig (W2, W4) und 36 andere, *S* = 85·2 + 80·36 + 200·2 + 390 = 3840, *M* = 5200; *S* lag über der Schnittgrenze (2223). Der Planner hat geschnitten, ohne geteilten Mechanismus außer dem Gerüst (Paket, `Konfig`, `New`, `mitCode`, etwa 30 Zeilen):

| Teil | Slice | Zeilen der Tabelle | Inhalt | *S* | *M* |
|---|---|---|---|---|---|
| 1 | `slice-inspection-use-case` (dieser Plan) | K1, A1, X1 bis X13, N1, W1, W2 (18; 1 nebenläufig, 17 andere) | Port `InspectionUseCase` mit `Exchanges`, Dienst, `Konfig` mit `Exchanges`, Seitenbildung, Allowlist der Felder; `FakeRecording`; *produktcode* etwa 200, *fakes* 1 | 85 + 1360 + 200 + 200 = 1845 | 2500 |
| 2 | [slice-inspection-coverage-diagnose](../open/slice-inspection-coverage-diagnose.md) | K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3, W4 (20; 1 nebenläufig, 19 andere) | `Coverage` und `Diagnose` als zusätzliche Methoden des Ports, zusätzliche Felder von `Konfig`, `Erlaubt`; `FakeCoverage`; *produktcode* etwa 190, *fakes* 1 | 85 + 1520 + 200 + 190 = 1995 | 2700 |

Beide Teile liegen unter der Schnittgrenze. Teil 2 folgt Teil 1, erweitert `Konfig` und den Port um Felder und Methoden, ohne bestehende zu ändern (die Signatur von `New` bleibt), und trägt seinen Block beim Start; `slice-admin-sessions` und `slice-bootstrap-m6` hängen an beiden. Eine Rückführung ist das nicht, der Schnitt geschah vor dem Code.

```groessenmarke
pfad: src/hexagon/ports/inbound/inspection.go
pfad: src/hexagon/services/inspection/
produktcode: 200
fakes: 1
basis: 81e960d14d0d53d78aa7b77de482372550a0cad5
schaetzung: 1845
marke: 2500
```

Die Marke dieses Teils (Planner, 2026-11-07, vor dem ersten Code-Commit): *S* = 85·1 + 80·17 + 200·1 + 200 = 1845, *M* = 2500, keine Schnittfrage. `basis` ist der Schnitt-Commit im Zweig, der Vertrag (`SPEC-236`), Tabelle, Raster und Mengen dieses Teils abschließt. Die Metrik steht in `SPEC-128` Nr. 1 und 4; die Marke des Teils 2 steht in dessen §4.

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07): jede Zelle nennt ihren Ort als Nummer von `SPEC-236` oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle an, die der Code anders entscheiden müsste (`AGENTS.md` §3.12). Die Zeilen *Coverage lesen* und *Diagnose lesen* führt `slice-inspection-coverage-diagnose`.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Exchanges einer Session lesen (Session, Seitenparameter) | `ctx` nil: `ErrUngueltig` ohne Code (Nr. 2); `Seite{}`: Anfang, Standard-`Limit` (Nr. 3); `sessionID` leer: unverändert an den Port, `ODR-E7003` (Nr. 4) | Session ohne Exchange und `Nach` hinter der letzten: leere Seite, `Weiter` 0 (Nr. 4); `Limit` 0: `StandardLimit` (Nr. 3) | unbekannte Session: `ODR-E7003` (Nr. 4, 8); `Limit` über `MaxLimit`: gekappt (Nr. 3); `Nach` oder `Limit` negativ: `ErrUngueltig` ohne Code (Nr. 3); Feld außerhalb der Allowlist: erscheint nie (Nr. 5) | vor dem Aufruf des Ports: sein Kontextfehler unverändert; nach der Rückkehr: Ergebnis, keine Prüfung (Nr. 2) | zustandslos; gleiche Argumente, gleiches Ergebnis; Anhängen zwischen zwei Seiten verschiebt nichts (Nr. 4, 10) |

- Die Paginierung (Auftrag 2) und die Seitenabfrage in `RecordingRepository` und SQLite-Adapter — **Ausgang:** entfallen; entschieden in `SPEC-236` Nr. 3 und 4 (Zeiger `sequence_no`, Standard 100, Obergrenze 1000; der Port bleibt, der Dienst schneidet im Speicher); die Größe einer Session ist ein akzeptiertes Negativ mit Grund und Folge (`SPEC-236` Nr. 4 und *Ungeprüft*).
- Lesen der Zähler (`TelemetryReader`, `SPEC-230` Nr. 9) — **Ausgang:** entfallen für diesen Slice; der Gegenstand ist Teil 2, `slice-inspection-coverage-diagnose` trägt die Sendung des Gebers `slice-telemetry-port` in Ziel, DoD und §4 (`AGENTS.md` §3.13).
- Aufträge 8 bis 13 (Tabelle, Pflicht-Mengen mit Achsen, Zweig-Menge mit Zählprüfung, Schnittfrage, Kennungen und Dateien, Katalog-Pflicht) — **Ausgang:** entfallen für den Architect (geliefert: §3 und `SPEC-236`); die Zählprüfung W1 und die Zählprüfungen der Mengen sind umgesetzt (§3, *Bericht des Implementers*); **weiter offen** bis zur Closure ist der Bericht je Befund aus Review und Verifikation mit Menge, Zahl der Elemente und einer Mutation an einem Element, das der Befund nicht nannte (`AGENTS.md` §3.17); Adresse: Implementer.
- Der Code für `ODR-E7012` ohne Endpunkt (Auftrag 6) — **Ausgang:** entfallen; entschieden in `SPEC-236` Nr. 9 und in `SPEC-005`.
- Geheimnisse in Exchanges: die Allowlist der Felder (`SPEC-236` Nr. 5) ist der Schutz, der Fänger ist X12 (`TestKeinGeheimnis`, umgesetzt, 7 Mutationen rot); der Fänger über alle Artefakte ist der Canary-Test in `slice-test-security` — **Ausgang:** weiter offen bis zur Closure; Adresse: Reviewer.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): `FakeRecording` trägt jede Schnittstelle des Originals und zählt jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Schätzung des ungeschnittenen Plans lag über der Schnittgrenze, der Schnitt ist vollzogen (§4, *Schnitt*); die Marke dieses Teils steht im Block (Planner, 2026-11-07, Schnitt-Commit als `basis`) — **Ausgang:** entfallen (Marke gesetzt); die Messung steht bei der Closure.
- Der Vertrag `SPEC-236` gilt für beide Teile, dieser Teil liefert `Exchanges` und `Konfig` mit dem Feld `Exchanges`; `Coverage`, `Diagnose` und die übrigen Felder (`SPEC-236` Nr. 1, 6, 7) fehlen im Code bis zur Closure von `slice-inspection-coverage-diagnose`, und kein Gate prüft diese Lücke (die Lücke steht allein hier und in §1; `SPEC-236` *Ungeprüft* nennt sie nicht, der Zitatort ist berichtigt) — **Ausgang:** weiter offen; Adresse: `slice-inspection-coverage-diagnose`.
- Doku-Update `SPEC-109` Nr. 5, Absatz *Anwendung* (Rückgabe 2 des Architect): der Satz zu `InspectionUseCase` fehlt; dieser Zweig ändert die Stelle nicht, weil `slice-mock-zuordnung` (Linie B) `SPEC-109` Nr. 5 und 6 ändert — **Ausgang:** offen; Adresse: Planner im Integrationsschritt (Merge-Commit): Gegenstand `Session` bei `Exchanges` (`SPEC-236` Nr. 8); der Satz für `Coverage` folgt mit `slice-inspection-coverage-diagnose`.
- Zählen der Fehler mit Code (Rückgabe 4, `SPEC-230` Nr. 8): der Dienst erhält nie `Telemetry` und zählt `ODR-E7003` nicht — **Ausgang:** entfallen für diesen Slice; die Sendung an den Rahmen des Admin-Listeners ist in `slice-admin-listener-token` getragen (§1 *Übernimmt*, DoD, Start-Trigger, Planner, 2026-11-07).
- Parameter `limit`, `after` und die Antwort auf einen ungültigen Wert (Rückgabe 3): der Use Case liefert `model.ErrUngueltig` ohne Code, die Spezifikation hat dafür keinen Code und keinen Status — **Ausgang:** entfallen für diesen Slice; getragen in `slice-admin-sessions` (§1 *Übernimmt*, DoD Punkt 2, Auftrag 4, §6) und im Code für `model.ErrUngueltig` aus `slice-admin-listener-token` (Planner, 2026-11-07).
- Die Kennung `SPEC-236`: `slice-telemetry-log` (Zweig) vergibt in `spezifikation.md` bisher `SPEC-230` bis `SPEC-232` und keine `SPEC-236` (geprüft am Zweig, 2026-11-07); dieser Slice behält die Nummer. Vergibt `slice-telemetry-log` später doch `SPEC-236`, stößt es beim zweiten Merge zusammen: der zweite Merger nummeriert seine Kennung auf die nächste freie Nummer der Linie D um (ab `SPEC-237`) und zieht alle Stellen nach (`AGENTS.md` §3.9). Derselbe Hinweis steht in `slice-inspection-coverage-diagnose` §6 — **Ausgang:** weiter offen bis zum Merge dieses Zweigs; Adresse: Planner (Integrationsschritt).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-07; Zähler = Zahl der Dateien in `evidence/`). Offen und einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (7×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure) für die Zweig-Menge (Auftrag 10); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (16×) für die Marke. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×) für `FakeRecording`; `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (29×, `AGENTS.md` §3.17), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14) für die Mengen; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, beide `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.16, für „ohne Geheimnisse“). Mit diesem Slice erreicht kein offener Eintrag mit weniger als 3× die Schwelle.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
