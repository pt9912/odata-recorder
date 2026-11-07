# Slice slice-inspection-coverage-diagnose: InspectionUseCase, Coverage und Diagnose lesen

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

**Berührte Spec-Stellen:** `LH-FA-13.a` (Coverage, Diagnose) · `SPEC-236` (Vertrag, Nr. 1, 2, 6, 7, 8, 10; vom Architect im Zweig von `slice-inspection-use-case` geschrieben) · `SPEC-005` (Zeilen `GET /__admin/coverage`, `GET /__admin/diagnostics`) · `LH-FA-11.a` · `SPEC-044`, `SPEC-043` (Coverage und Evidenz) · `SPEC-107` (`CoverageRepository`) · `SPEC-109` (Fehlerabbildung, Anwendung in `SPEC-236` Nr. 8) · `SPEC-230` Nr. 8 und 9 (`TelemetryReader`, `Snapshot().Errors`) · `SPEC-071`, `SPEC-112` · `ARC-002`, `ARC-003` · `architecture.md` §2.2 (Zeile `InspectionUseCase`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07.

**Herkunft:** Teil 2 des Schnitts von [slice-inspection-use-case](../done/slice-inspection-use-case.md) (Entscheidung des Architect, Schnitt vor dem Code, `SPEC-128` Nr. 1; vollzogen vom Planner am 2026-11-07). Der Plan trägt die Zeilen K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3 und W4 der Tabelle dieses Slice.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `InspectionUseCase` (Inbound Port `src/hexagon/ports/inbound/inspection.go`, Dienst `src/hexagon/services/inspection`) liest zusätzlich die Coverage einer Baseline und die Diagnose (Konfigurationszusammenfassung ohne Geheimnisse, Modus, Zähler je Fehlercode) ohne Geheimnisse; er verändert nichts und kennt kein HTTP. Der Slice erweitert Port und `Konfig` des Dienstes aus `slice-inspection-use-case` additiv um die Methoden `Coverage` und `Diagnose` und um die Felder `Coverage`, `Zaehler`, `Modus`, `Einstellungen`; die Signatur von `New` und die bestehende Methode `Exchanges` bleiben.

**Übernimmt (Sendung von `slice-inspection-use-case`, Schnitt des Architect vom 2026-11-07):** Coverage und Diagnose mit den Zeilen K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3 und W4 der Tabelle; das Doku-Update zu `SPEC-109` Nr. 5, Absatz *Anwendung*, soweit es `Coverage` betrifft (Gegenstand `Snapshot`, `SPEC-236` Nr. 8), entsteht im Integrationsschritt dieses Slice durch den Planner, nicht im Zweig (siehe §6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Exchanges, die Seitenbildung, die Allowlist der Felder einer Exchange-Zeile, das Gerüst (Paket, `Konfig`, `New`, `mitCode`) und `FakeRecording` — `slice-inspection-use-case`; dieser Slice setzt es als `done/` voraus und ändert es nur additiv.
- Die Endpunkte, Pfade und Statuscodes, auch die Antwort auf einen ungültigen Parameter — `slice-admin-sessions` und `slice-admin-listener-token`; der Use Case kennt kein HTTP, der Adapter bildet ab.
- Die Ports `Telemetry` und `TelemetryReader` und die Zähler selbst — `welle-betrieb` (`slice-telemetry-port`, `done/`); der Dienst liest die Zähler über `TelemetryReader`, er erzeugt sie nicht und erhält nie `Telemetry` oder den Adapter (`SPEC-230` Nr. 9).
- Das Zählen der Fehler mit Code, die der Dienst liefert (`ODR-E7002`), und aller anderen — der Dienst zählt sie nicht; die Stelle, die den Fehler ausliefert, zählt (`SPEC-230` Nr. 8): der Admin-Adapter in `slice-admin-listener-token`.
- Die Zusammenfassung der Konfiguration aus der Umgebung bilden und die Schlüsselmenge des Starts gegen `Erlaubt()` zählen — `slice-bootstrap-m6` (der Root bildet den typisierten Wert; der Dienst sieht nie die Umgebung); der Dienst nimmt `Modus` und `Einstellungen` als Wert in `Konfig`.
- Export und Abruf der Metriken und Spans — `welle-betrieb`.
- Audit des Lesens — `SPEC-005` auditiert die Schreibendpunkte; ein Lesezugriff schreibt keinen Eintrag.

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

- [ ] Coverage: die Einträge je EntitySet einer Baseline (Parameter ist die Baseline-Kennung) mit Status und Evidenz aus `CoverageRepository`; eine unbekannte Baseline `ODR-E7002` (`SPEC-236` Nr. 6).
- [ ] Diagnose: Konfigurationszusammenfassung aus der Allowlist von Schlüsseln (nie Token, Zugangsdaten, Adressen oder Pfade, `SPEC-236` Nr. 7), Modus und Zähler je Fehlercode über den Lese-Port `TelemetryReader` (`Snapshot().Errors`, ein Eintrag je Code, auch mit 0; `SPEC-230` Nr. 9).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist; dazu `SPEC-109` Nr. 5, Absatz *Anwendung*: der Satz zu `InspectionUseCase` ist vollständig (Exchanges mit dem Gegenstand `Session` stammt aus `slice-inspection-use-case`, hier `Coverage` mit dem Gegenstand `Snapshot`), eingetragen im Integrationsschritt durch den Planner.
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
| `src/hexagon/ports/inbound/inspection.go` | update | `Coverage` und `Diagnose` als zusätzliche Methoden, Typen der Ergebnisse (additiv) |
| `src/hexagon/services/inspection/` | update | Lesen der Coverage und der Diagnose; `Konfig` um `Coverage`, `Zaehler`, `Modus`, `Einstellungen`; `Erlaubt` |
| `src/hexagon/ports/outbound/`, `src/adapters/driven/sqlite/` | unverändert | `CoverageRepository.List` genügt (`SPEC-236` Nr. 6); beide Pfade liegen nicht im Block |
| `spec/spezifikation.md` | `SPEC-236` besteht (Architect, im Zweig von `slice-inspection-use-case`); Nachzug `SPEC-109` Nr. 5 im Integrationsschritt (Planner) | siehe §6 |
| Tests | neu | Fälle je Operation; `FakeCoverage` und `FakeTelemetry` (beide Ports, aus `slice-telemetry-port`) halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service |

Der Vertrag steht als `SPEC-236` in der Spezifikation (Architect, 2026-11-07, im Zweig von `slice-inspection-use-case`; auf `main` mit dessen Merge, der diesem Slice vorausgeht), die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) hier und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind entschieden, der Implementer entscheidet keine im Code.
Die Spalte `Nr.` trägt den Buchstaben der Art: `K` Konfiguration, `A` Kontext nil, `C` Coverage, `D` Diagnose, `N` akzeptiertes Negativ, `W` Zählprüfung der Stellen und Nebenläufigkeit. Die Nummern bleiben die der Tabelle in `slice-inspection-use-case` (dort K1, A1, X1 bis X13, N1, W1, W2). Die Spalte *Fänger* nennt Fälle und die Mengen des Abschnitts *Pflicht-Mengen* darunter; die Namen der Tests wählt der Implementer. Jede Zeile ist ein Weg des Dienstes (*Modell-Menge der Wege* unten); ein Test, der die Wege im Quelltext zählt, hält Zeilen und Wege gleich (W3).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| K2 | `New` weist ein nil in `Coverage` und in `Zaehler` ab (Nr. 1) | Schleife über beide Felder, je ein nil, sonst gültig: `ErrUngueltig`, kein Dienst | die Prüfung eines der beiden Felder entfernt | andere |
| K3 | `New` nimmt genau die vier Modi an, jeder andere ist `ErrUngueltig` (Nr. 1) | Schleife über `record`, `replay`, `mock`, `hybrid` (gültig) und `""`, `Record`, `record ` (Leerzeichen dahinter), ` record`, `live` (ungültig) | Modus-Prüfung entfernt; Prüfung ohne `hybrid` | andere |
| K4 | `Einstellungen` darf nil sein; `New` kopiert tief, eine spätere Änderung der Abbildung wirkt nicht auf den Dienst (Nr. 1) | Fälle: nil gültig; nach `New` einen Wert ändern, einen Schlüssel hinzufügen und einen löschen, dann `Diagnose` | Abbildung ohne Kopie gespeichert | andere |
| A2 | `ctx` nil bei `Coverage` und `Diagnose` ist `ErrUngueltig` ohne `Fehler`-Wert, kein Port wird gerufen (Nr. 2) | Schleife über beide Operationen; `FakeCoverage` und `FakeTelemetry` zählen jede Operation: 0 | Prüfung bei einer der beiden entfernt | andere |
| C1 | Eine leere `baselineID` ist `ErrUngueltig` ohne Code und ohne Port; jede andere geht unverändert und genau einmal an `List` (Nr. 6) | Schleife über `""` (ungültig), `b1`, ein Leerzeichen, `w1` (fremd), 64 Zeichen, zwei Bytes jenseits von ASCII; `FakeCoverage` hält Argument und Zahl der Aufrufe | `TrimSpace`; leere Kennung an den Port | andere |
| C2 | Das Ergebnis hat einen `CoverageEintrag` je Eintrag des Ports in dessen Reihenfolge, `EntitySet`, `Status` und `Evidenz` byte-gleich; eine Baseline ohne Einträge liefert eine leere Liste (nie `nil`); keine Seitenbildung, kein abgeleiteter Status (Nr. 6) | Generator M-Coverage mit Orakel (die Fixture-Liste in Reihenfolge); darin m gleich 0 und m gleich 20 | Evidenz geparst und neu kodiert; Status aus der Evidenz abgeleitet; die Liste auf `StandardLimit` gekürzt | andere |
| C3 | Jeder Fehler von `List` wird nach `SPEC-109` Nr. 5 mit dem Gegenstand `Snapshot` abgebildet: `ErrNichtGefunden` (auch für eine Kennung anderer Art) ist `ODR-E7002` (Nr. 6, 8) | Schleife über Menge M-Fehler (Aufrufstelle `List`): jeder Eintrag von `outbound.EintraegeAbbildung()`, ein anderer Fehler (`ODR-E9001`); erwartet `ODR-E7002` für `ErrNichtGefunden` | Gegenstand `Session` (ergäbe `ODR-E7003`) | andere |
| C4 | Kontextfehler unverändert ohne Code; `ctx` an `List` unverändert; Abbruch nach der Rückkehr ändert das Ergebnis nicht (Nr. 2) | Fälle: Identität des `ctx` im Fake; vor dem Aufruf abgebrochen (Fake liefert `ctx.Err()`); der Fake bricht ab und liefert Daten: Ergebnis ohne Fehler; Schleife über Form {`Canceled`, `DeadlineExceeded`, verpackt, `errors.Join` mit `ErrNichtGefunden` in beiden Reihenfolgen}; `errors.Is` und `==`-Gleichheit, kein `*Fehler` | Fehlerabbildung vor der Prüfung des Kontexts; `context.Background()` an den Port; `ctx.Err()` nach dem Aufruf geprüft | andere |
| C5 | Lesen schreibt nichts: `Coverage` ruft von `CoverageRepository` nur `List`, genau einmal (Nr. 1) | `FakeCoverage` zählt `Get` und `List` bei Erfolg, Fehler und Parameterfehler: `Get` 0, `List` höchstens 1 | `Get` je Eintrag; zweiter `List` | andere |
| D1 | `Diagnose` mit beendetem `ctx` ist dessen Fehler, unverändert und ohne Code, ohne Aufruf von `Snapshot`; ein lebender liefert das Ergebnis (Nr. 7) | Fälle: lebend, `Canceled`, `DeadlineExceeded`; `FakeTelemetry.Anzahl(OpSnapshot)` | `ctx.Err()` nicht geprüft; geprüft nach `Snapshot` | andere |
| D2 | Der Dienst ruft `Telemetry` nie; `Snapshot` genau einmal je Aufruf (Nr. 1, 7) | Schleife über jede Operation des `FakeTelemetry` (Quelle: die Konstanten von `supporttest.Operation`, Zählprüfung gegen ihre Menge): außer `Snapshot` 0, `Snapshot` 1 je Aufruf | ein `CountError` oder `Log` im Dienst; zweiter `Snapshot` | andere |
| D3 | `Fehler` hat einen Eintrag je Code von `model.AlleCodes()` in dieser Reihenfolge, `Anzahl` gleich dem Zähler (Nr. 7) | Generator M-Codes mit Orakel (Fixture-Abbildung); Zählprüfung gegen `len(model.AlleCodes())`, nie 37 | Schleife über die Schlüssel der Abbildung statt der Tabelle (Zählerstand 0 fehlt); Reihenfolge der Abbildung; ein Code ausgelassen | andere |
| D4 | Ein Code ohne Eintrag hat `Anzahl` 0; ein Eintrag außerhalb der Tabelle wird verworfen; ein nil `Errors` ist gültig (Nr. 7) | Fälle: `Errors` nil, leer, nur ein Code; Schlüssel 0, -1, letzter + 1, größter `int32` | Eintrag außerhalb durchgereicht; nil-Abbildung ohne Behandlung | andere |
| D5 | `Einstellungen` enthält genau die Einträge, deren Schlüssel byte-gleich in `Erlaubt()` stehen, mit unverändertem Wert; ein fehlender fehlt (Nr. 7) | Schleife über `Erlaubt()` (jeder Schlüssel, eigener Wert); Zählprüfung: `Erlaubt()` gleich der Literal-Liste im Test (Wortlaut der Nr. 7); Abwandlungen jedes erlaubten Schlüssels (Kleinschreibung, Leerraum davor und dahinter, Präfix, Suffix) fehlen; ein erlaubter ohne Eintrag fehlt, kein leerer Wert | Vergleich ohne Beachtung der Schreibweise; `TrimSpace` auf den Schlüssel; ein Schlüssel zu viel in `Erlaubt()` | andere |
| D6 | Kein Geheimnis in der Diagnose: kein Wert unter einem ausgeschlossenen Schlüssel erscheint (Nr. 7) | Allaussage als Schleife über Menge M-Schlüssel: Canary als Wert unter jedem ausgeschlossenen Schlüssel (Liste der Nr. 7 und `ADMIN_TOKEN`) × Kodierung {json, `%+v`, Gang über alle Zeichenketten}; Gegenprobe: derselbe Canary unter einem erlaubten Schlüssel erscheint | Allowlist zur Sperrliste umgekehrt; `DATA_DIR` oder `UPSTREAM_URL` in `Erlaubt()` | andere |
| D7 | `Modus` ist der `Modus` der Konfiguration, unverändert (Nr. 7) | Schleife über die vier Modi | Modus aus `Einstellungen["MODE"]`; fester Wert | andere |
| D8 | Das Ergebnis ist eine Kopie und hat nie nil: `Einstellungen` und `Fehler`; jeder Aufruf liest den Stand neu (Nr. 7) | Fälle: nil-Konfiguration ergibt `{}` und `[]` (JSON); Änderung des Ergebnisses wirkt nicht auf den Dienst oder das nächste Ergebnis; Änderung des Standes im Fake zwischen zwei Aufrufen zeigt das zweite Ergebnis neu | Abbildung der Konfiguration ohne Kopie zurückgegeben; Ergebnis des ersten Aufrufs gespeichert | andere |
| N2 | Akzeptierte Negative: Schlüsselmenge und Modi des Root nicht gegen den Dienst gezählt; Zählen der Fehler des Dienstes (`SPEC-236` *Ungeprüft*) | kein Fänger; Grund und Folge in `SPEC-236` *Ungeprüft*; D5 hält `Erlaubt()` gegen den Wortlaut | – | andere |
| W3 | Zählprüfung: die Wege von `Coverage`, `Diagnose` und `New` im Quelltext sind die Zeilen K1 bis K4 (Tabelle des Vorgängers, hier K2 bis K4), A2, C1 bis C5, D1 bis D8 | Test mit `go/parser` über die Dateien des Pakets ohne `_test`: Zahl der Aufrufe über Felder von `Konfig`, der Aufrufe von `mitCode` und der Stellen, die `ErrUngueltig` bilden, gleich der Zahl der Wege der Tabelle des Tests (Name, Art, Nr. dieser Tabelle); der Test von W1 aus `slice-inspection-use-case` zählt die Wege von `New` nach diesem Slice mit der Tabelle K1 bis K4 und `Exchanges` mit X1 bis X13 (Nachzug im selben Commit, `AGENTS.md` §3.9) | eine Stelle `ErrUngueltig` oder ein Port-Aufruf ohne Zeile | andere |
| W4 | Gleichzeitige Aufrufe von `Coverage` und `Diagnose` stören sich nicht, der Dienst hält keinen Zustand (Nr. 10) | 8 Goroutinen mit je 50 Aufrufen gegen einen Dienst, für beide Operationen, auch gemischt mit `Exchanges`; jedes Ergebnis gleich dem sequentiellen; `make test-race` je Prozess (`-test.count=1`) | Zwischenspeicher der Coverage oder der Zusammenfassung im Dienst: Detektor rot in N von N Prozessläufen (Zahl nennt der Bericht) | nebenläufig |

**Pflicht-Mengen** (`AGENTS.md` §3.15, §3.17, `implement-slice` Schritt 14): je Menge Achsen, Zahl der Elemente als Produkt, Generator, Orakel (nie der Dienst und nie sein Code), Zählprüfung gegen die Quelle der Menge (keine feste Zahl im Test), Prüfgrenze der unbeschränkten Menge.

| Menge | Achsen | Elemente | Orakel | Zählprüfung gegen die Quelle | Prüfgrenze (akzeptiertes Negativ) |
|---|---|---|---|---|---|
| M-Coverage (C2) | `Status` {`complete`, `partial`, `unknown`, ein fremder Text} × `Evidenz` {ein JSON-Text, leer, ein Text mit Leerraum und Zeichen jenseits von ASCII aus Bytewerten} × Zahl der Einträge m {0, 1, 2, 20}; Kennung (C1) {leer, bekannt, Leerraum, fremd (Kennung eines Workspace-Snapshots), 64 Zeichen, Bytes jenseits von ASCII} | 4 × 3 × 4 = 48; C1: 6 | die Fixture-Liste in der Reihenfolge, in der der Fake sie liefert | die drei Status sind die Konstanten `model.CoverageComplete`, `model.CoveragePartial`, `model.CoverageUnknown`; der Test prüft für jede `Coverage{...}.Validate()` gleich nil und für den fremden Text ungleich nil, damit „fremd“ fremd bleibt | Evidenz-Inhalte sind unbeschränkt: drei Formen; Folge: ein Dienst, der die Evidenz liest und umformt, würde bei einer weiteren Form auffallen, nicht bei diesen; der Dienst parst sie nie |
| M-Codes (D3, D4) | Code {jeder von `model.AlleCodes()`} × Stand im Fake {fehlt, 0, 1, 2 hoch 32, größter `uint64`}; außerhalb der Tabelle {0, -1, letzter + 1, größter `int32`} | Länge von `model.AlleCodes()` × 5 (zurzeit 37 × 5, mit `ODR-E6008` aus `slice-mock-zuordnung` 38 × 5), dazu 4 | die Fixture-Abbildung; erwartet je Code der Tabelle ein Eintrag mit dem Fixture-Wert, fehlend 0 | `len(model.AlleCodes())` und die Reihenfolge von `AlleCodes()`, nie die Zahl im Test | keine; die Menge der Codes ist beschränkt und kommt aus der Quelle |
| M-Schlüssel (D5, D6) | erlaubter Schlüssel {jeder von `Erlaubt()`} × Abwandlung {unverändert, Kleinschreibung, Leerraum davor, Leerraum dahinter, Präfix, Suffix}; ausgeschlossener Schlüssel {die Liste der Nr. 7 und `ADMIN_TOKEN`} × Kodierung {json, `%+v`, Gang über alle Zeichenketten}; Modus {4} | Länge von `Erlaubt()` × 6; Länge der Liste × 3; 4 | Fixture: jeder erlaubte Schlüssel mit eigenem Wert; erwartet nur die unveränderten | die Literal-Liste im Test gleich `Erlaubt()` (Wortlaut Nr. 7); die Liste der ausgeschlossenen gleich den Schlüsseln der Nr. 7 | die Menge der Schlüssel der Startkonfiguration liegt im Root (nicht importierbar, `SPEC-236` *Ungeprüft*): ein Schlüssel, den der Root neu einführt, ist ausgeschlossen (Standard ist Verwerfen) und wird nie gezeigt; Folge: unschädlich, die Allowlist ist die Zusage |
| M-Fehler (C3, C4, D1) | Aufrufstelle `List` × Fehler {jeder Eintrag von `outbound.EintraegeAbbildung()`, ein anderer Fehler}; `List` × Form des Kontextfehlers {`Canceled`, `DeadlineExceeded`, verpackt, Join vorn, Join hinten}; `List` × `ctx`-Zustand {lebend, vor dem Aufruf abgebrochen, vom Fake abgebrochen mit Daten}; `Diagnose` × {lebend, `Canceled`, `DeadlineExceeded`} | (Zahl der Einträge + 1) + 5 + 3 + 3 | `outbound.Abbilden` ist die Funktion unter Test in `SPEC-109`; das Orakel hier ist die Tabelle des Tests (Name, Wert, erwarteter Code je Gegenstand), die aus `SPEC-109` Nr. 5 abgeschrieben ist, nicht der Aufruf von `Abbilden` | die Tabelle des Tests hat genau die Namen von `outbound.EintraegeAbbildung()` (Quelle), nie eine feste Zahl | keine für die Fehler; die Form der Kontextfehler ist unbeschränkt (jede Verpackung): fünf geprüft, Folge: eine Verpackung, die `errors.Is` nicht trifft, ist ein Fehler des Ports. Die Aufrufstelle `ListBySession` führt `slice-inspection-use-case` |

**Modell-Menge der Wege** (der Test W3 zählt sie gegen den Quelltext): jeder Verwurf nach Grund, jede Fehlerabbildung je Port, jedes Feld der Allowlist.

| Operation | Wege | Zeilen |
|---|---|---|
| `New` (Anteil dieses Slice) | nil `Coverage` oder `Zaehler`; Modus ungültig; Einstellungen kopiert; Erfolg | K2, K3, K4 |
| `Coverage` | `ctx` nil; Kennung leer; `List`: Kontextfehler, Fehler mit Code, Erfolg (je Eintrag eine Zeile, leer eine leere Liste) | A2, C1 bis C5 |
| `Diagnose` | `ctx` nil; `ctx` beendet; `Snapshot`; Eintrag der Tabelle (gesetzt, fehlend); Eintrag außerhalb; Schlüssel erlaubt, verworfen, fehlend; Modus | A2, D1 bis D8 |

Jeder Weg läuft in einem Test des Pakets und damit unter `make test-race`; ein Weg ohne Test ist nach W3 rot. Ein benanntes Negativ gibt es nur für die Fehler des Ports, die der echte Adapter nie liefert (etwa `ErrBlobFehlt`): der Fake liefert sie, der Weg ist derselbe wie für jeden Fehler (`mitCode`), Folge keine.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts.
- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Entscheidungen des Architect für diesen Teil** (2026-11-07; Vertrag `SPEC-236`, Ausgang der Aufträge 1, 4, 5, 8 bis 10 und 13 von `slice-inspection-use-case`; eine ADR entsteht nicht):

1. **Coverage:** Parameter ist die Baseline-Kennung (HTTP `baseline_id`, Pflicht), nicht Workspace oder Service; keine Seitenbildung; unbekannt oder fremde Art ist `ODR-E7002` (`SPEC-236` Nr. 6).
2. **Diagnose:** `Modus` und `Einstellungen` kommen als fertiger Wert aus dem Root; Allowlist von 18 Schlüsseln, jeder andere wird verworfen; die Zähler sind ein Eintrag je Code von `model.AlleCodes()` aus `Snapshot().Errors` (fehlend 0); `TelemetryReader` reicht, die Anwendung weicht von `SPEC-230` Nr. 8 und 9 nicht ab; der Dienst prüft `ctx.Err()` einmal, weil kein Port der Diagnose einen Kontext nimmt (`SPEC-236` Nr. 7).
3. **Katalog:** kein neuer Fehlercode, Schlüssel oder Kommando, kein Eintrag in `harness/katalog-ausnahmen.txt`, kein Handbuch-Kapitel; das Binary bleibt unverändert (`SPEC-201` Nr. 3 und 11).
4. **Kennungen:** keine eigene; `SPEC-236` besteht und ändert dieser Slice nicht (eine Änderung des Vertrags geht an den Architect, eine an `SPEC-109` oder `SPEC-230` an den Planner).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-inspection-use-case` (Gerüst, Port, `Konfig`, `New`, `FakeRecording`, `SPEC-236` auf `main`) und `slice-telemetry-port` (`welle-betrieb`, `TelemetryReader`) liegen in `done/` (erfüllt: `slice-inspection-use-case` ist mit der Closure vom 2026-11-07 auf `main`, Messung 1408 von 2500, Auslastung 0,56). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-236`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** Der Slice erweitert Paket, Port und `Konfig` aus `slice-inspection-use-case` (*Code-Abhängigkeit*: `New`, `mitCode`, die Typen `Seite` und `Fehler`-Abbildung liegen dort) und liest die Zähler über `TelemetryReader` aus `slice-telemetry-port` (*Code-Abhängigkeit*, `src/hexagon/ports/support/`). `CoverageRepository` liegt seit `welle-persistenz` in `src/hexagon/ports/outbound/` und braucht nichts davon.

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`** (Mock, `WorkspaceUseCase`, HTTP-Adapter, Composition Root): der Dienst kennt weder HTTP noch einen Modus als Verhalten, nur als Wert der Konfiguration.

**Größenmarke** (Stand vor dem Start, Planner, 2026-11-07, `SPEC-128` Nr. 1): Die Rechnung aus der Tabelle in §3 ergibt 20 Zeilen, davon 1 nebenläufig (W4) und 19 andere, dazu `fakes` 1 (`FakeCoverage`; `FakeTelemetry` besteht seit `slice-telemetry-port`) und `produktcode` 190 (Methoden und Typen des Ports, `Coverage`, `Diagnose`, `Erlaubt`, Felder von `Konfig`): *S* = 85·1 + 80·19 + 200·1 + 190 = 1995, *M* = 2700; keine Schnittfrage (*S* unter 2223). Der Block mit `pfad` (`src/hexagon/ports/inbound/inspection.go`, `src/hexagon/services/inspection/`), `produktcode`, `fakes`, `basis`, `schaetzung` und `marke` steht erst beim Start: `basis` ist der Commit, der den Stand von `slice-inspection-use-case` und diese Tabelle trägt, und ein Block ohne diese Werte wäre ein Platzhalter. Die Marke gilt für die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); die Prosa-Marke oben ist der Stand vor dem Start).
- `in-progress` → `open` (blockiert — Carveout?): `TelemetryReader` reicht für die Zähler der Diagnose nicht (Form oder Fehlerfall; Rückgabe an den Planner, der Port gehört `slice-telemetry-port` in `done/`).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07): jede Zelle nennt ihren Ort als Nummer von `SPEC-236` oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle an, die der Code anders entscheiden müsste (`AGENTS.md` §3.12). Die Zeile *Exchanges einer Session lesen* führt `slice-inspection-use-case`.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Coverage lesen (Baseline) | `ctx` nil: `ErrUngueltig` (Nr. 2); es gibt keinen Standard und kein „alle Baselines“ (Nr. 6) | `baselineID` leer: `ErrUngueltig` ohne Port (Nr. 6); Baseline ohne Einträge: leere Liste (Nr. 6) | unbekannte Baseline und Kennung eines Snapshots anderer Art: `ODR-E7002` (Nr. 6, 8); Status oder Evidenz: unverändert (Nr. 6) | vor dem Aufruf des Ports: sein Kontextfehler unverändert; nach der Rückkehr: Ergebnis, keine Prüfung (Nr. 2) | zustandslos; eine Baseline ist unveränderlich (`SPEC-107`), das Ergebnis gleich (Nr. 10) |
| Diagnose lesen (Konfiguration, Modus, Zähler) | `ctx` nil: `ErrUngueltig` (Nr. 2); `Einstellungen` nil: `{}` (Nr. 1, 7); Code ohne Zähler: `Anzahl` 0 (Nr. 7) | `Errors` leer oder nil: jeder Code mit 0 (Nr. 7) | Schlüssel außerhalb der Allowlist: verworfen (Nr. 7); Eintrag außerhalb der Tabelle: verworfen (Nr. 7); ungültiger Modus: `New` scheitert (Nr. 1) | beendeter `ctx` vor `Snapshot`: sein Fehler, kein `Snapshot` (Nr. 2, 7); nach `Snapshot`: gilt nicht, es folgt kein blockierender Schritt | jeder Aufruf liest den Stand neu, die Zähler wachsen (Nr. 7, 10) |

- Lesen der Zähler: `slice-telemetry-port` (`done/`) hat entschieden, dass `TelemetryReader` mit `Snapshot().Errors` liest, nicht `Telemetry` (`SPEC-230` Nr. 9, Sendung des Gebers an `slice-inspection-use-case`, von dort an diesen Slice weitergegeben); dieser Slice trägt sie in Ziel, DoD und §4 (`AGENTS.md` §3.13) — **Ausgang:** entfallen (die Frage nach einer Leseoperation ist entschieden, Form und Fehlerfall stehen in `SPEC-236` Nr. 7).
- Die Parameter von `GET /__admin/coverage` und der Code für eine unbekannte Baseline — **Ausgang:** entfallen; entschieden in `SPEC-236` Nr. 6, 8 und in `SPEC-005`.
- Geheimnisse in der Diagnose: die Allowlist der Schlüssel (`SPEC-236` Nr. 7) ist der Schutz, der Fänger ist D6; der Fänger über alle Artefakte ist der Canary-Test in `slice-test-security` — **Ausgang:** weiter offen bis zur Closure; Adresse: Implementer (D6), danach Reviewer.
- Die Zählprüfungen (W3 und die der Mengen) und der Bericht je Befund mit Menge, Zahl der Elemente und einer Mutation an einem Element, das der Befund nicht nannte (`AGENTS.md` §3.17) — **Ausgang:** weiter offen; Adresse: Implementer.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): `FakeCoverage` und `FakeTelemetry` tragen jede Schnittstelle des Originals und zählen jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `SPEC-109` Nr. 5, Absatz *Anwendung*, nennt `InspectionUseCase` nicht; der Satz für `Coverage` (Gegenstand `Snapshot`, `SPEC-236` Nr. 8) gehört zu diesem Teil, aber der Zweig ändert `SPEC-109` nicht (`slice-mock-zuordnung`, Linie B, ändert dieselbe Stelle) — **Ausgang:** offen; Adresse: Planner im Integrationsschritt (Merge-Commit), nach dem Satz für `Exchanges` aus `slice-inspection-use-case`.
- `SPEC-236` steht mit dem Merge von `slice-inspection-use-case` auf `main` für den ganzen Vertrag; bis zur Closure dieses Slice liefert der Code `Coverage` und `Diagnose` nicht — **Ausgang:** weiter offen bis zur Closure dieses Slice; Adresse: Implementer (kein Gate prüft die Lücke zwischen Vertrag und Code; sie steht allein hier und in §1, `SPEC-236` *Ungeprüft* nennt sie nicht).
- Kennung `SPEC-236`: `slice-telemetry-log` (Zweig) vergibt in `spezifikation.md` bisher `SPEC-230` bis `SPEC-232` und keine `SPEC-236`; vergibt es später doch `SPEC-236`, stößt es beim zweiten Merge zusammen: der zweite Merger nummeriert seine Kennung auf die nächste freie Nummer der Linie D um (ab `SPEC-237`) und zieht alle Stellen nach (`AGENTS.md` §3.9); derselbe Hinweis steht in `slice-inspection-use-case` §6 — **Ausgang:** weiter offen bis zum Merge von `slice-inspection-use-case`; Adresse: Planner (Integrationsschritt).
- Größe: die Prosa-Marke in §4 steht (*S* 1995, *M* 2700); der Block folgt beim Start — **Ausgang:** offen; Adresse: Planner beim Start.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-07; Zähler = Zahl der Dateien in `evidence/`); der Planner zieht Treffer und Zähler beim Start nach. Offen und einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (7×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure) für die Zweig-Menge (W3); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (16×) für die Marke. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×) für `FakeCoverage` und `FakeTelemetry`; `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (29×, `AGENTS.md` §3.17), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14) für die Mengen; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, beide `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.16, für „ohne Geheimnisse“). Mit diesem Slice erreicht kein offener Eintrag mit weniger als 3× die Schwelle.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
