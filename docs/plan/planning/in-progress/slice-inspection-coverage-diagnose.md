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

**Berührte Spec-Stellen:** `LH-FA-13.a` (Coverage, Diagnose) · `SPEC-236` (Vertrag, Nr. 1, 2, 6, 7, 8, 10; vom Architect im Zweig von `slice-inspection-use-case` geschrieben, im Zweig dieses Slice um einen Eintrag unter *Ungeprüft* ergänzt) · `SPEC-110` Nr. 2 (die 27 Schlüssel, aus denen die Allowlist der Diagnose 18 wählt) · `SPEC-005` (Zeilen `GET /__admin/coverage`, `GET /__admin/diagnostics`) · `LH-FA-11.a` · `SPEC-044`, `SPEC-043` (Coverage und Evidenz) · `SPEC-107` (`CoverageRepository`) · `SPEC-109` (Fehlerabbildung, Anwendung in `SPEC-236` Nr. 8) · `SPEC-230` Nr. 8 und 9 (`TelemetryReader`, `Snapshot().Errors`) · `SPEC-071`, `SPEC-112` · `ARC-002`, `ARC-003` · `architecture.md` §2.2 (Zeile `InspectionUseCase`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-inspection-coverage-diagnose`, Worktree `/Development/odata-recorder-wt-inspection-coverage-diagnose`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07.

**Herkunft:** Teil 2 des Schnitts von [slice-inspection-use-case](../done/slice-inspection-use-case.md) (Entscheidung des Architect, Schnitt vor dem Code, `SPEC-128` Nr. 1; vollzogen vom Planner am 2026-11-07). Der Plan trägt die Zeilen K2 bis K4, A2, C1 bis C5, D1 bis D8, N2, W3 und W4 der Tabelle dieses Slice; der Architect hat beim Start K5 und C6 ergänzt (Entscheidung 5).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `InspectionUseCase` (Inbound Port `src/hexagon/ports/inbound/inspection.go`, Dienst `src/hexagon/services/inspection`) liest zusätzlich die Coverage einer Baseline und die Diagnose (Konfigurationszusammenfassung ohne Geheimnisse, Modus, Zähler je Fehlercode) ohne Geheimnisse; er verändert nichts und kennt kein HTTP. Der Slice erweitert Port und `Konfig` des Dienstes aus `slice-inspection-use-case` additiv um die Methoden `Coverage` und `Diagnose` und um die Felder `Coverage`, `Zaehler`, `Modus`, `Einstellungen`; die Signatur von `New` und die bestehende Methode `Exchanges` bleiben.

**Übernimmt (Sendung von `slice-inspection-use-case`, Schnitt des Architect vom 2026-11-07):** Coverage und Diagnose mit den Zeilen K2 bis K5, A2, C1 bis C6, D1 bis D8, N2, W3 und W4 der Tabelle; das Doku-Update zu `SPEC-109` Nr. 5, Absatz *Anwendung*, soweit es `Coverage` betrifft (Gegenstand `Snapshot`, `SPEC-236` Nr. 8), entsteht im Integrationsschritt dieses Slice durch den Planner, nicht im Zweig (siehe §6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Exchanges, die Seitenbildung, die Allowlist der Felder einer Exchange-Zeile, das Gerüst (Paket, `Konfig`, `New`, `mitCode`) und `FakeRecording` — `slice-inspection-use-case`; dieser Slice setzt es als `done/` voraus und ändert es nur additiv. Die Tests des Vorgängers zieht er nach, soweit `New` mit dem Zusatz strenger wird (Nachzug, §3).
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
| `src/hexagon/services/inspection/` | update | Lesen der Coverage und der Diagnose; `Konfig` um `Coverage`, `Zaehler`, `Modus`, `Einstellungen`; `Erlaubt`; Nachzug der Tests des Vorgängers (siehe *Nachzug*) |
| `src/hexagon/ports/outbound/`, `src/adapters/driven/sqlite/` | unverändert | `CoverageRepository.List` genügt (`SPEC-236` Nr. 6); beide Pfade liegen nicht im Block |
| `spec/spezifikation.md` | `SPEC-236` besteht (Architect, im Zweig von `slice-inspection-use-case`), im Zweig dieses Slice um einen Eintrag unter *Ungeprüft* ergänzt (Architect); Nachzug `SPEC-109` Nr. 5 im Integrationsschritt (Planner) | siehe §6 |
| Tests | neu | Fälle je Operation; `FakeCoverage` und `FakeTelemetry` (beide Ports, aus `slice-telemetry-port`) halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service |

Der Vertrag steht als `SPEC-236` in der Spezifikation (Architect, 2026-11-07, im Zweig von `slice-inspection-use-case`; auf `main` mit dessen Merge, der diesem Slice vorausgeht), die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) hier und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind entschieden, der Implementer entscheidet keine im Code.
Die Spalte `Nr.` trägt den Buchstaben der Art: `K` Konfiguration, `A` Kontext nil, `C` Coverage, `D` Diagnose, `N` akzeptiertes Negativ, `W` Zählprüfung der Stellen und Nebenläufigkeit. Die Nummern bleiben die der Tabelle in `slice-inspection-use-case` (dort K1, A1, X1 bis X13, N1, W1, W2); K5 und C6 sind neu. Die Spalte *Fänger* nennt Fälle und die Mengen des Abschnitts *Pflicht-Mengen* darunter; die Namen der Tests wählt der Implementer. Jede Zeile ist ein Weg des Dienstes (*Modell-Menge der Wege* unten); ein Test, der die Wege im Quelltext zählt, hält Zeilen und Wege gleich (W3).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| K2 | `New` weist ein nil in `Coverage` und in `Zaehler` ab (Nr. 1) | Schleife über beide Felder, je ein nil, sonst gültig: `ErrUngueltig`, kein Dienst | die Prüfung eines der beiden Felder entfernt | andere |
| K3 | `New` nimmt genau die vier Modi an, jeder andere ist `ErrUngueltig` (Nr. 1) | Schleife über `record`, `replay`, `mock`, `hybrid` (gültig) und `""`, `Record`, `record ` (Leerzeichen dahinter), ` record`, `recorder`, `live` (ungültig); die Literale stehen im Test (Wortlaut Nr. 1), keine Konstante des Dienstes | Modus-Prüfung entfernt; Prüfung ohne `hybrid`; Vergleich mit `EqualFold`, `TrimSpace` oder als Präfix | andere |
| K4 | `Einstellungen` darf nil sein; `New` kopiert, eine spätere Änderung der Abbildung wirkt nicht auf den Dienst (Nr. 1) | Fälle: nil gültig; nach `New` einen Wert ändern, einen Schlüssel hinzufügen und einen löschen, dann `Diagnose` | Abbildung ohne Kopie gespeichert | andere |
| K5 | `*inspection.Inspektion` erfüllt `inbound.InspectionUseCase` mit `Exchanges`, `Coverage` und `Diagnose` (Lerneintrag `BEO-REPO/dienst-erfuellt-den-port-ohne-bindung`) | die Zeile `var _ inbound.InspectionUseCase = (*inspection.Inspektion)(nil)` in `port_test.go` besteht und wächst mit dem Port, ohne Zusatz; Nachweis ist die Übersetzungsmeldung (*Nachweis der Bindung* unten); eine Zählprüfung der Methoden per `reflect` gehört nicht dazu (Entscheidung 5 f) | Signatur von `Coverage` (zusätzlicher Parameter) oder Rückgabetyp von `Diagnose` im Dienst geändert; `Diagnose` im Dienst entfernt | andere |
| A2 | `ctx` nil bei `Coverage` und `Diagnose` ist `ErrUngueltig` ohne `Fehler`-Wert, kein Port wird gerufen (Nr. 2) | Schleife über beide Operationen (bei `Coverage` auch mit zugleich leerer Kennung); `FakeCoverage` und `FakeTelemetry` zählen jede Operation: 0 | Prüfung bei einer der beiden entfernt | andere |
| C1 | Eine leere `baselineID` ist `ErrUngueltig` ohne Code und ohne Port, auch bei beendetem `ctx`; jede andere geht unverändert und genau einmal an `List` (Nr. 2, 6) | Schleife über Kennung × `ctx`-Zustand {lebend, beendet}: `""` (ungültig, `List` 0, `ErrUngueltig` und nicht der Kontextfehler), `b1`, `B1`, ein Leerzeichen, ` b1`, `b1 `, `w1` (fremd), 64 Zeichen, zwei Bytes jenseits von ASCII (gültig, unverändert, `List` 1); `FakeCoverage` hält Argument und Zahl der Aufrufe | `TrimSpace`; `ToLower`; leere Kennung an den Port; Kontextprüfung vor der Argumentprüfung | andere |
| C2 | Das Ergebnis hat einen `CoverageEintrag` je Eintrag des Ports in dessen Reihenfolge, `EntitySet`, `Status` und `Evidenz` byte-gleich; eine Baseline ohne Einträge liefert eine leere Liste (nie `nil`); keine Seitenbildung, kein abgeleiteter Status, kein Zwischenspeicher (Nr. 6, 10) | Generator M-Coverage mit Orakel (die Fixture-Liste in der Reihenfolge des Fakes, nicht aufsteigend nach `entity_set`); darin m gleich 0 (der Fake liefert `nil` und eine leere Liste) und m gleich 101 und 1001; zwei Aufrufe nacheinander mit geänderter Fixture dazwischen: das zweite Ergebnis ist neu, `List` 2 | Evidenz geparst und neu kodiert; Status aus der Evidenz abgeleitet; Liste auf 100 oder 1000 Einträge gekürzt; Liste aufsteigend sortiert; Liste im Dienst zwischengespeichert | andere |
| C3 | Jeder Fehler von `List` wird nach `SPEC-109` Nr. 5 mit dem Gegenstand `Snapshot` abgebildet, `Status` ist der eine Status des Codes: `ErrNichtGefunden` (auch für eine Kennung anderer Art) ist `ODR-E7002`; erreichbar sind `ODR-E7002`, `ODR-E3001` und `ODR-E9001` (darunter `ErrUngueltig` des Ports) (Nr. 6, 8) | Schleife über Menge M-Fehler (Aufrufstelle `List`): jeder Eintrag von `outbound.EintraegeAbbildung()` und ein anderer Fehler, je verpackt; Tabelle des Tests aus `SPEC-109` Nr. 5 für den Gegenstand `Snapshot` abgeschrieben (`ErrNichtGefunden` gleich `ODR-E7002`, `ErrBereitsVorhanden` und `ErrUngueltig` des Modells gleich `ODR-E9001`, weil die Zeile `Snapshot` nicht nennt); erwartet Code, `errors.Is` auf die Ursache und den Status (Literal 404 für `ODR-E7002`, 500 für `ODR-E3001` und `ODR-E9001`, bei den übrigen `StatusErlaubt`) | Gegenstand `Session` (ergäbe `ODR-E7003` und für `ErrBereitsVorhanden` `ODR-E7010`); Gegenstand `Sonstiges`; Status 0 | andere |
| C4 | Kontextfehler unverändert ohne Code; `ctx` an `List` unverändert; Abbruch nach der Rückkehr ändert das Ergebnis nicht; der Dienst prüft den Kontext selbst nicht (Nr. 2) | Fälle: Identität des `ctx` im Fake; vor dem Aufruf abgebrochen (Fake liefert `ctx.Err()`, `List` 1, `Get` 0); der Fake bricht ab und liefert Daten: Ergebnis ohne Fehler; Schleife über Form {`Canceled`, `DeadlineExceeded`, verpackt, `errors.Join` mit `ErrNichtGefunden` in beiden Reihenfolgen}; `errors.Is` und `==`-Gleichheit, kein `*Fehler` | Fehlerabbildung vor der Prüfung des Kontexts; `context.Background()` an den Port; `ctx.Err()` vor oder nach dem Aufruf geprüft | andere |
| C5 | Lesen schreibt nichts: `Coverage` ruft von `CoverageRepository` nur `List`, genau einmal (Nr. 1) | `FakeCoverage` zählt `Get` und `List` bei Erfolg, Fehler, Parameterfehler, `ctx` nil und Kontextfehler: `Get` 0, `List` höchstens 1 | `Get` je Eintrag; zweiter `List` | andere |
| C6 | Ein Ergebnis ist bei einem Fehler der Nullwert, bei `Coverage` `nil`, bei `Diagnose` `Diagnose{}` (Absatz vor Nr. 1) | Schleife über jeden Fehlerzweig von `Coverage` (`ctx` nil, leere Kennung, 5 Kontextfehler-Formen, jeder Eintrag von M-Fehler) mit einem Fake, der Daten und Fehler zugleich liefert: Ergebnis `== nil` (nicht leer); `Diagnose` mit `ctx` nil, `Canceled`, `DeadlineExceeded`: `reflect.DeepEqual(d, inbound.Diagnose{})`, also `Einstellungen` und `Fehler` nil (im Erfolgsfall nie nil, D8) | im Fehlerfall `[]CoverageEintrag{}` oder die gelesene Liste neben dem Fehler; `Diagnose` mit gesetztem `Modus` neben dem Kontextfehler | andere |
| D1 | `Diagnose` mit beendetem `ctx` ist dessen Fehler, unverändert und ohne Code, ohne Aufruf von `Snapshot`; ein lebender liefert das Ergebnis, auch wenn er während `Snapshot` endet (Nr. 2, 7) | Fälle: lebend, `Canceled`, `DeadlineExceeded`; `ctx` endet im Aufruf von `Snapshot` (Haken des Fakes, siehe *Fakes*): Ergebnis ohne Fehler; `FakeTelemetry.Anzahl(OpSnapshot)` 0 bei beendetem, 1 sonst | `ctx.Err()` nicht geprüft; geprüft nach `Snapshot` | andere |
| D2 | Der Dienst ruft `Telemetry` nie; `Snapshot` genau einmal je Aufruf; kein Log, kein Zähler (Nr. 1, 7) | nach Erfolg, beendetem `ctx` und `ctx` nil: jeder Eintrag von `FakeTelemetry.Aufrufe()` hat `Op == OpSnapshot`, ihre Zahl gleicht der Zahl der Aufrufe, die `Snapshot` erreichen (Quelle: die Aufzeichnung des Fakes, nicht eine Liste der Operationen); dazu Schleife über die Konstanten von `supporttest.Operation` (`OpLog` bis `OpSnapshot`): `Anzahl` 0 außer `OpSnapshot`, Summe der `Anzahl` gleich der Länge von `Aufrufe()` | ein `CountError` oder `Log` im Dienst; zweiter `Snapshot` | andere |
| D3 | `Fehler` hat einen Eintrag je Code von `model.AlleCodes()` in dieser Reihenfolge, `Anzahl` gleich dem Zähler (Nr. 7) | Generator M-Codes mit Orakel (Fixture-Abbildung, je Code ein eigener Wert); Zählprüfung gegen `len(model.AlleCodes())`, nie gegen eine feste Zahl | Schleife über die Schlüssel der Abbildung statt der Tabelle (Zählerstand 0 fehlt); Reihenfolge der Abbildung; ein Code ausgelassen; Wert des Nachbarcodes | andere |
| D4 | Ein Code ohne Eintrag hat `Anzahl` 0; ein Eintrag außerhalb der Tabelle wird verworfen; ein nil `Errors` ist gültig (Nr. 7) | Fälle: `Errors` nil, leer, nur ein Code; Schlüssel 0, -1, größter Code + 1, `math.MaxInt` | Eintrag außerhalb durchgereicht; nil-Abbildung ohne Behandlung | andere |
| D5 | `Einstellungen` enthält genau die Einträge, deren Schlüssel byte-gleich in `Erlaubt()` stehen, mit unverändertem Wert, in jedem Modus gleich; ein fehlender fehlt (Nr. 7) | Schleife über `Erlaubt()` (jeder Schlüssel, eigener Wert); Zählprüfung: `Erlaubt()` gleich der Literal-Liste im Test (Wortlaut der Nr. 7); Abwandlungen jedes erlaubten Schlüssels (Kleinschreibung, Leerraum davor und dahinter, Präfix, Suffix) fehlen; ein erlaubter ohne Eintrag fehlt, kein leerer Wert; ein erlaubter mit leerem Wert bleibt mit leerem Wert; dieselbe Konfiguration in allen vier Modi liefert gleiche `Einstellungen` | Vergleich ohne Beachtung der Schreibweise; `TrimSpace` auf den Schlüssel; leerer Wert verworfen; ein Schlüssel zu viel in `Erlaubt()`; Filter je Modus | andere |
| D6 | Kein Geheimnis in der Diagnose: kein Wert unter einem ausgeschlossenen Schlüssel erscheint (Nr. 7) | Allaussage als Schleife über Menge M-Ausgeschlossen (die Namen der Nr. 7, je ein Vertreter der Muster `ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`, `PUBLIC_BASE_URL`, `STRICT_COVERAGE`, ein unbekannter Name, der leere Name) × Kodierung {json, `%+v`, Gang über alle Zeichenketten}: der Canary unter dem Schlüssel erscheint nirgends; Gegenprobe: derselbe Canary unter einem erlaubten Schlüssel erscheint | Allowlist zur Sperrliste umgekehrt; `DATA_DIR` oder `UPSTREAM_URL` in `Erlaubt()` | andere |
| D7 | `Modus` ist der `Modus` der Konfiguration, unverändert (Nr. 7) | Schleife über die vier Modi (Literale) | Modus aus `Einstellungen["MODE"]`; fester Wert | andere |
| D8 | Das Ergebnis ist eine Kopie und hat im Erfolgsfall nie nil: `Einstellungen` und `Fehler`; jeder Aufruf liest den Stand neu (Nr. 7, 10) | Fälle: nil-Konfiguration und nil `Errors`: `json.Marshal` je Feldwert gleich `{}` bzw. `[]` (Literale); Änderung des Ergebnisses (Schlüssel hinzufügen und löschen, Wert und `Fehler[0].Anzahl` ändern) wirkt weder auf den Dienst noch auf das nächste Ergebnis; Änderung des Standes im Fake zwischen zwei Aufrufen zeigt das zweite Ergebnis neu | Abbildung der Konfiguration ohne Kopie zurückgegeben; Ergebnis des ersten Aufrufs gespeichert; `Fehler` aus einem gespeicherten Feld | andere |
| N2 | Akzeptierte Negative dieses Teils: Schlüsselmenge und Modi des Root nicht gegen den Dienst gezählt; Zählen der Fehler des Dienstes; Verzweigungen ohne Aufruf oder Stelle zählt W3 nicht; Erreichbarkeit der Codes gegen den echten Adapter; Inhalt der Evidenz (Grund und Folge unter *Akzeptierte Negative*) | kein Fänger; die ersten, zweiten und letzten stehen in `SPEC-236` *Ungeprüft*, die übrigen unten; D5 hält `Erlaubt()` gegen den Wortlaut | – | andere |
| W3 | Zählprüfung: die Wege von `New`, `Coverage` und `Diagnose` im Quelltext sind die Zeilen K1 bis K4, A1 und A2, C1 bis C5, D1 bis D8 (Tabelle *Wege des Tests* unten); die Importe des Pakets sind die erlaubten (`SPEC-236` Nr. 1: keine Umgebung, kein HTTP, kein Log, kein Adapter, kein Dienst) | Test mit `go/parser` über die Dateien des Pakets ohne `_test`: Zahl der Stellen, die `ErrUngueltig` bilden, der Aufrufe von `ctxNil`, der Aufrufe über Felder von `Konfig`, der Aufrufe von `mitCode` und der Aufrufe von `Err` auf `ctx`, je gleich der Zahl der Wege der Tabelle des Tests (Name, Art, Nr. dieser Tabelle); Importe: jeder Pfad liegt unter `model`, `ports/inbound`, `ports/outbound` oder `ports/support` des Moduls oder ist Standardbibliothek (erstes Segment ohne Punkt) außerhalb von `os`, `os/…`, `net`, `net/…`, `log`, `log/…`, `syscall`, `database/sql`; der Test von W1 aus `slice-inspection-use-case` zählt die Wege von `New` nach diesem Slice mit K1 bis K4 und `Exchanges` mit X1 bis X13 (Nachzug im selben Commit, `AGENTS.md` §3.9) | eine Stelle `ErrUngueltig`, ein Port-Aufruf, ein Aufruf von `ctxNil` oder ein `ctx.Err()` in `Coverage` ohne Zeile; `import "os"`, `"net/http"` oder `"log/slog"` im Dienst | andere |
| W4 | Gleichzeitige Aufrufe von `Exchanges`, `Coverage` und `Diagnose` stören sich nicht, der Dienst hält keinen Zustand und keinen Zwischenspeicher (Nr. 10) | Schleife über die Zweig-Menge (Abschnitt *Zweig-Menge*): je Zweig 8 Goroutinen mit je 20 Runden gegen einen Dienst, ein Test der gemischten Folge, Wirkungs-Fänger (nummerierte Antworten), `make test-race` je Prozess (`-test.count=1`), Abdeckungsprofil unter dem Detektor; jedes Ergebnis gleich dem sequentiellen | Zwischenspeicher der Coverage oder der Zusammenfassung im Dienst, mit und ohne Sperre: Wirkungs-Fänger rot in jedem Lauf, Detektor rot in N von N Prozessläufen (Zahl nennt der Bericht, Ziel 20 von 20); ein Schreibzugriff auf eine geteilte Variable in einem Zweig | nebenläufig |

**Nachweis der Bindung (K5).** Die Bindung prüft zur Übersetzungszeit, ein Übersetzungsfehler ist hier kein Bruch des Aufbaus, sondern der Weg des Fängers. Der Bericht nennt je Mutation die Meldung von `make test` mit der Zeile von `port_test.go` und dem Wortlaut (`does not implement inbound.InspectionUseCase`, `wrong type for method Coverage` bzw. `missing method Diagnose`); eine Meldung an anderer Stelle gilt nicht. Die Zählprüfung der Methoden per `reflect` gehört nicht dazu: die Bindung prüft, dass der Dienst den Port erfüllt, und der Port ist über `SPEC-236` auf drei Methoden festgelegt; eine zusätzliche Methode des Dienstes (etwa eine schreibende) ändert den Port nicht, und ihren Aufruf eines Ports zählt W3 über die Felder von `Konfig`.

**Pflicht-Mengen** (`AGENTS.md` §3.15, §3.17, `implement-slice` Schritt 14): je Menge Achsen, Zahl der Elemente als Produkt, Generator, Orakel (nie der Dienst und nie sein Code), Zählprüfung gegen die Quelle der Menge (keine feste Zahl im Test), Prüfgrenze der unbeschränkten Menge.

| Menge | Achsen | Elemente | Orakel | Zählprüfung gegen die Quelle | Prüfgrenze (akzeptiertes Negativ) |
|---|---|---|---|---|---|
| M-Coverage (C1, C2) | `Status` {`complete`, `partial`, `unknown`, ein fremder Text} × `Evidenz` {ein JSON-Text, leer, ein Text mit Leerraum und Zeichen jenseits von ASCII aus Bytewerten} × Zahl der Einträge m {0, 1, 2, 101, 1001}; Kennung (C1) {leer, `b1`, `B1`, Leerzeichen, Leerzeichen davor, Leerzeichen dahinter, fremd (`w1`), 64 Zeichen, Bytes jenseits von ASCII} × `ctx`-Zustand {lebend, beendet} | 4 × 3 × 5 = 60; C1: 9 × 2 = 18 | die Fixture-Liste in der Reihenfolge, in der der Fake sie liefert (nicht aufsteigend); die Status als Literale `complete`, `partial`, `unknown`, nie `model.Coverage*` als Erwartung | die Literale sind gleich den Konstanten `model.CoverageComplete`, `model.CoveragePartial`, `model.CoverageUnknown`; der Test prüft für jede `Coverage{...}.Validate()` gleich nil und für den fremden Text ungleich nil, damit „fremd“ fremd bleibt | Evidenz-Inhalte sind unbeschränkt: drei Formen; Folge: ein Dienst, der die Evidenz liest und umformt, würde bei einer weiteren Form auffallen, nicht bei diesen; der Dienst parst sie nie. m: 101 und 1001 liegen je eins über `StandardLimit` und `MaxLimit` von `Exchanges`, den Grenzen, auf die eine Seitenbildung kürzte; eine Kürzung auf eine Zahl ab 1001 bliebe unentdeckt, Folge: keine, eine Baseline hat so viele EntitySets nicht (`SPEC-127` Nr. 12) |
| M-Codes (D3, D4) | Code {jeder von `model.AlleCodes()`} × Stand im Fake {fehlt, 0, 1, 2 hoch 32, größter `uint64`}; außerhalb der Tabelle {0, -1, größter Code + 1, `math.MaxInt`} | Länge von `model.AlleCodes()` × 5 (zurzeit 37 × 5, mit `ODR-E6008` aus `slice-mock-zuordnung` 38 × 5), dazu 4 | die Fixture-Abbildung, je Code ein eigener Wert (abhängig vom Index des Codes); erwartet je Code der Tabelle ein Eintrag mit dem Fixture-Wert, fehlend 0; die Reihenfolge ist die von `AlleCodes()` (Quelle des Vertrags in `SPEC-109` Nr. 2, nicht des Dienstes) | `len(model.AlleCodes())` und die Reihenfolge von `AlleCodes()`, nie die Zahl im Test | keine; die Menge der Codes ist beschränkt und kommt aus der Quelle |
| M-Erlaubt (D5) | erlaubter Schlüssel {jeder von `Erlaubt()`} × Abwandlung {unverändert, Kleinschreibung, Leerraum davor, Leerraum dahinter, Präfix, Suffix} × Modus {4} | Länge von `Erlaubt()` × 6 × 4 | Fixture: jeder erlaubte Schlüssel mit eigenem Wert; erwartet nur die unveränderten; die 18 Schlüssel als Literale aus dem Wortlaut der Nr. 7 | `Erlaubt()` gleich der Literal-Liste als Menge und in der Länge (zurzeit 18); die Literal-Liste ist frei von Doppelten | die Menge der Schlüssel der Startkonfiguration liegt im Root (nicht importierbar, `SPEC-236` *Ungeprüft*): ein Schlüssel, den der Root neu einführt, ist ausgeschlossen (Standard ist Verwerfen) und wird nie gezeigt; Folge: unschädlich, die Allowlist ist die Zusage |
| M-Ausgeschlossen (D6) | ausgeschlossener Name {die zehn Namen der Nr. 7 (`MODE`, `LISTEN`, `DATA_DIR`, `SERVICE_ID`, `SERVICE_ROOT_PATH`, `SERVICE_ODATA_VERSION`, `UPSTREAM_URL`, `SERVICES_FILE`, `UPSTREAM_CA_FILE`, `ADMIN_TOKEN_FILE`), je ein Vertreter der Muster `ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*` und `PUBLIC_BASE_URL`, `STRICT_COVERAGE`, ein unbekannter Name, der leere Name} × Kodierung {json, `%+v`, Gang über alle Zeichenketten} | 19 × 3 = 57 | der Canary je Name, ein eigener Wert; erwartet: kommt in keiner Kodierung vor | die Literal-Liste ist zu `Erlaubt()` disjunkt (Quelle); die 18 erlaubten und die 9 Einzelnamen der Nr. 7 zusammen sind die 27 Schlüssel von `SPEC-110` Nr. 2 (vom Architect von Hand gezählt, kein Test liest die Spezifikation) | die Muster sind unbeschränkt (jeder Name `ADMIN_*`): ein Vertreter je Muster; Folge: ein Name, den die Allowlist nicht nennt, ist verworfen, das gilt für jeden Namen ohne Unterschied, und die Gegenprobe zeigt es am erlaubten |
| M-Fehler (C3, C4, C6, D1) | Aufrufstelle `List` × Fehler {jeder Eintrag von `outbound.EintraegeAbbildung()`, ein anderer Fehler}; `List` × Form des Kontextfehlers {`Canceled`, `DeadlineExceeded`, verpackt, Join vorn, Join hinten}; `List` × `ctx`-Zustand {lebend, vor dem Aufruf abgebrochen, vom Fake abgebrochen mit Daten}; `Diagnose` × {lebend, `Canceled`, `DeadlineExceeded`, in `Snapshot` abgebrochen} | (Zahl der Einträge + 1) + 5 + 3 + 4 | `outbound.Abbilden` ist die Funktion unter Test in `SPEC-109`; das Orakel hier ist die Tabelle des Tests (Name, Wert, erwarteter Code und Status je Gegenstand `Snapshot`), die aus `SPEC-109` Nr. 5 abgeschrieben ist, nicht der Aufruf von `Abbilden` | die Tabelle des Tests hat genau die Namen von `outbound.EintraegeAbbildung()` (Quelle), nie eine feste Zahl | keine für die Fehler; die Form der Kontextfehler ist unbeschränkt (jede Verpackung): fünf geprüft, Folge: eine Verpackung, die `errors.Is` nicht trifft, ist ein Fehler des Ports. Die Aufrufstelle `ListBySession` führt `slice-inspection-use-case` |

**Literale des Vertrags im Orakel** (Lerneintrag `slice-inspection-use-case` F-1; Architect): die Orakel dieser Zeilen tragen die Werte als Literale aus dem Wortlaut von `SPEC-236` und `SPEC-109`, nie die Konstante des Dienstes; eine Konstante dient allein der Zählprüfung gegen das Literal.

| Zeilen | Literale im Orakel | Konstante der Zählprüfung |
|---|---|---|
| K3, D7 | `record`, `replay`, `mock`, `hybrid` | keine; der Dienst führt sie im Code, der Test hält sie im Wortlaut |
| D5, D6 | die 18 Schlüssel: `REPLAY_STRATEGY`, `REPLAY_ASSIGNMENT`, `REPLAY_REDACTED_BODIES`, `MATCH_VARY_HEADERS`, `MATCH_INVALID_ENCODING`, `MATCH_BODY_MODE`, `REDACT_HEADERS`, `REDACT_QUERY_PARAMS`, `REDACT_JSON_PATHS`, `MAX_CAPTURE_BODY`, `UPSTREAM_TIMEOUT`, `UPSTREAM_TLS_VERIFY`, `UPSTREAM_TLS_INSECURE_SKIP_VERIFY`, `SHUTDOWN_TIMEOUT`, `RETENTION_MAX_AGE`, `SQLITE_BUSY_TIMEOUT`, `VOLATILE_HEADERS`, `LINK_REWRITE`; die ausgeschlossenen Namen aus M-Ausgeschlossen | `inspection.Erlaubt()` (gleich der Liste als Menge und in der Länge) |
| C2 | `complete`, `partial`, `unknown`, ein fremder Text; Evidenz als eigener Text | `model.CoverageComplete`, `model.CoveragePartial`, `model.CoverageUnknown` (Wert gleich Literal) |
| C3 | Code und Status je Fehler aus `SPEC-109` Nr. 5 und `SPEC-018`, `SPEC-025`, `SPEC-113`: `ODR-E7002` mit 404, `ODR-E3001` mit 500, `ODR-E9001` mit 500 | `outbound.EintraegeAbbildung()` für die Namen |
| D3, D4 | je Code der Wert der Fixture; Reihenfolge von `AlleCodes()` | `model.AlleCodes()` |
| D8 | `{}` und `[]` als Ergebnis von `json.Marshal` je Feldwert | keine |

**Fixtures gleichen nie dem Wert einer Mutation** (`BEO-REPO/fixture-gleicht-dem-wert-der-mutation`): je Zeile nennt die Fänger-Spalte beide Seiten, etwa den Schlüssel in der Schreibweise der Allowlist neben seinen Abwandlungen (D5), `b1` neben `B1` und `b1 ` (C1, gegen `ToLower` und `TrimSpace`), eine Fixture-Liste, die nicht aufsteigend liegt (C2, gegen das Sortieren), je Code einen eigenen Zählerwert (D3, gegen Vertauschen), m gleich 101 und 1001 statt 20 (C2, gegen das Kürzen auf `StandardLimit` oder `MaxLimit`).

**Modell-Menge der Wege** (der Test W3 zählt sie gegen den Quelltext): jeder Verwurf nach Grund, jede Fehlerabbildung je Port, jedes Feld der Allowlist.

| Operation | Wege | Zeilen |
|---|---|---|
| `New` (Anteil dieses Slice) | nil `Coverage` oder `Zaehler`; Modus ungültig; Einstellungen kopiert; Erfolg | K2, K3, K4 |
| `Coverage` | `ctx` nil; Kennung leer; `List`: Kontextfehler, Fehler mit Code, Erfolg (je Eintrag eine Zeile, leer eine leere Liste) | A2, C1 bis C6 |
| `Diagnose` | `ctx` nil; `ctx` beendet; `Snapshot`; Eintrag der Tabelle (gesetzt, fehlend); Eintrag außerhalb; Schlüssel erlaubt, verworfen, fehlend; Modus | A2, D1 bis D8 |

Jeder Weg läuft in einem Test des Pakets und damit unter `make test-race`; ein Weg ohne Test ist nach W3 rot. Ein benanntes Negativ gibt es nur für die Fehler des Ports, die der echte Adapter nie liefert (etwa `ErrBlobFehlt`): der Fake liefert sie, der Weg ist derselbe wie für jeden Fehler (`mitCode`), Folge keine.

**Wege des Tests (W3, Tabelle `wege()` nach dem Nachzug).** Jede Stelle im Quelltext hat genau eine Zeile; der Implementer schreibt je Verwurf eine eigene Stelle `ErrUngueltig` (keine gemeinsame Hilfsfunktion außer `ctxNil`) und ruft `ctxNil` je Operation, damit die Zählung jede Operation sieht. Der Test des Vorgängers zählte die Stelle in `ctxNil` und hätte den Aufruf in `Coverage` und `Diagnose` nicht gesehen; deshalb kommt die Art *Aufruf von `ctxNil`* hinzu.

| Art | Wege (Name · Zeile) | Zahl |
|---|---|---|
| Stelle, die `ErrUngueltig` bildet | `ctxNil` (A1, A2) · `New`: `Exchanges` nil (K1) · `New`: `Coverage` nil (K2) · `New`: `Zaehler` nil (K2) · `New`: Modus (K3) · `Exchanges`: `Nach` unter 0 (X1) · `Exchanges`: `Limit` unter 0 (X1) · `Coverage`: Kennung leer (C1) | 8 |
| Aufruf von `ctxNil` | `Exchanges` (A1) · `Coverage` (A2) · `Diagnose` (A2) | 3 |
| Aufruf eines Ports über ein Feld von `Konfig` | `Exchanges.ListBySession` (X7) · `Coverage.List` (C1, C5) · `Zaehler.Snapshot` (D2) | 3 |
| Aufruf von `mitCode` | `Exchanges` (X9) · `Coverage` (C3) | 2 |
| Aufruf von `Err` auf dem `ctx` | `Diagnose` (D1); in `Coverage` und `Exchanges` keiner (C4) | 1 |

Die Zählung fängt eine Verzweigung nicht, die weder einen Port ruft noch `ErrUngueltig` bildet (in `Diagnose` der Schlüssel, der Eintrag außerhalb, der fehlende Eintrag, der Modus): sie hält D3 bis D7 und das Abdeckungsprofil (benanntes Negativ in N2, Folge dort).

**Zweig-Menge von `TestGleichzeitig` (W4; Architect).** Der Test des Vorgängers (`TestGleichzeitig`, `zweige()`, `wegeExchanges()`, `TestZweigeSindDieWege`) wächst zur Schleife über die Zweige aller drei Operationen, nicht zu zwei zusätzlichen Aufrufen. Die Namen der Wege tragen die Operation (`Coverage: …`, `Diagnose: …`), damit die Zählprüfung sie trennt.

| Operation | Zweige (Weg der Modell-Menge · Erwartung) | Zahl |
|---|---|---|
| `Coverage` | `ctx` nil (ungültig) · Kennung leer (ungültig) · Erfolg mit `nil` vom Fake (leere Liste) · Erfolg mit leerer Liste · Erfolg mit 101 Einträgen, nicht aufsteigend · Kontextfehler in je einer der 5 Formen · Fehler mit Code je Eintrag von `outbound.EintraegeAbbildung()` und ein anderer | 10 + n, n = Länge von `outbound.EintraegeAbbildung()` |
| `Diagnose` | `ctx` nil · `ctx` beendet (`Canceled`, `DeadlineExceeded`) · `Errors` nil · `Errors` mit Eintrag je Code und einem Eintrag außerhalb · `Einstellungen` nil · `Einstellungen` mit erlaubtem, verworfenem und fehlendem Schlüssel je in einem der vier Modi | 1 + 2 + 1 + 1 + 1 + 4 = 10 |
| `Exchanges` | die Zweige des Vorgängers | wie dort |

- **Runden.** Je Zweig 8 Goroutinen mit je 20 Runden gegen einen Dienst (wie `slice-inspection-use-case` W2); jede Goroutine führt den Zweig in jeder Runde aus und prüft das Ergebnis gegen die Erwartung des Zweigs. Ein Zweig, den eine Goroutine nur einmal erreichen würde (`Errors` nil, ein Eintrag außerhalb, ein Verwurf, `ErrNichtGefunden`), läuft deshalb 160-mal je Prozess. Dazu ein Test der gemischten Folge: dieselben 8 Goroutinen rufen je Runde alle drei Operationen in wechselnder Reihenfolge gegen einen Dienst, jedes Ergebnis gleich dem sequentiellen.
- **Zählprüfung der Zweige.** `TestZweigeSindDieWege` hält die Wege je Operation (Listen von Hand, ein Eintrag je Weg der Modell-Menge) gegen die Zweige, die Fehler mit Code einen je Eintrag von `outbound.EintraegeAbbildung()` und einen anderen, die Kontextfehler 5 je Operation; keine feste Zahl der Einträge.
- **Abdeckungsprofil als Messung des Implementers.** `go test -race -run '^TestGleichzeitig$' -coverpkg=<Paket> -coverprofile` im Container (Betriebsregeln unten): kein Block von `Coverage`, `Diagnose`, `Erlaubt` und `New` ohne Treffer, oder ein benanntes Negativ je Block mit Grund und Folge im Bericht. Das Profil ist eine einmalige Messung vor der Übergabe und kein Gate; was darüber hinaus bleibt, hält die Zählprüfung der Zweige.
- **Wirkungs-Fänger, wo der Detektor schweigt.** Die Fakes ordnen jeden Aufruf eines Ports über eine gemeinsame Sperre; ein Zwischenspeicher, der nach dem ersten Port-Aufruf geschrieben wird, entgeht dem Detektor meist (`SPEC-236` Nr. 10, *Ungeprüft* in `SPEC-124`). Deshalb liefert `FakeCoverage` je Aufruf eine fortlaufende Nummer in der `Evidenz` des ersten Eintrags und der Fake von `Snapshot` eine in `Anzahl`: bei N Aufrufen sind die N Antworten paarweise verschieden, lückenlos 1 bis N, und die Zahl der Port-Aufrufe ist N. Ein Zwischenspeicher liefert eine Nummer doppelt und ist in jedem Lauf rot, auch ohne Detektor. Die Goroutinen schreiben außerdem in ihr Ergebnis (Schlüssel in `Einstellungen`, `Fehler[0].Anzahl`): ein geteilter Speicher bricht unter dem Detektor und zählt in der Wirkung (D8 sequentiell).

**Fakes** (`AGENTS.md` §3.15; Architect). `FakeCoverage` liegt in einer eigenen Datei des Testpakets neben `helfer_test.go` (`helfer_coverage_test.go`), weil die Datei des Vorgängers `FakeRecording` trägt und dieser Slice sie nicht umschreibt. Er trägt beide Operationen von `CoverageRepository`: `List` hält die Kennung und den `ctx` (als Funktion gebunden wie in `supporttest.Aufruf`, damit kein Feld vom Typ `context.Context` entsteht), zählt den Aufruf, liefert die Fixture oder den gesetzten Fehler und kennt zwei Haken (den `ctx` beenden, die Nummer des Aufrufs in die Evidenz legen); `Get` zählt und liefert `ErrNichtGefunden`, keine Panik (`BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`). Für `FakeTelemetry` (`supporttest`) genügen `SetzeStand`, `Aufrufe` und `Anzahl`; D1 (der `ctx` endet in `Snapshot`) und die Nummern des Wirkungs-Fängers brauchen einen Haken: ein Typ im Test bettet `*supporttest.FakeTelemetry` ein, überschreibt allein `Snapshot`, ruft darin den eingebetteten (damit dieser zählt und aufzeichnet) und führt danach den Haken aus. Das ist eine Einbettung und keine Hülle um den Port: kein Argument, keine Zahl und keine Operation wird verdeckt, und die Operationen von `Telemetry` bleiben erreichbar und zählbar (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`).

**Nachzug der Tests des Vorgängers (Auftrag J; Architect, im selben Commit wie der Code, `AGENTS.md` §3.9).**

1. `New` weist mit diesem Slice ein nil in `Coverage` und `Zaehler` und einen ungültigen `Modus` ab. Jede Stelle der Tests des Vorgängers, die `Konfig{Exchanges: …}` allein baut (`neuerDienst`, der gültige Fall von K1), erhält einen Helfer mit `FakeCoverage`, `FakeTelemetry` und Modus `record`; sonst scheitert `New` in jedem Test des Vorgängers.
2. `TestWegeImQuelltext` (W1) zählt nach dem Zusatz die fünf Arten der Tabelle *Wege des Tests* mit 17 Wegen (8, 3, 3, 2, 1). Ohne die Art *Aufruf von `ctxNil`* sähe er den Aufruf in `Coverage` und `Diagnose` nicht, weil die Stelle `ErrUngueltig` einmal in `ctxNil` steht; mit ihr fängt er jeden Aufruf, jede Stelle `ErrUngueltig`, jeden Port-Aufruf über ein Feld von `Konfig`, jeden Aufruf von `mitCode` und jedes `ctx.Err()`.
3. `wegeExchanges()` (13 Wege von Hand), `zweige()` und `TestZweigeSindDieWege` wachsen um die Wege von `Coverage` und `Diagnose` (Abschnitt *Zweig-Menge*): die Wege der Modell-Menge je Operation, Fehler mit Code und Kontextfehler je Operation gezählt.
4. Grenze, die bleibt (Teil 1, `BEO-REPO/zusage-ueber-verzweigungen-reicher-als-die-funktionsliste`): die Zählung zählt Aufrufe und Stellen und keine Verzweigungen; die Verzweigungen von `Diagnose` ohne Aufruf halten D3 bis D7 und das Abdeckungsprofil.

**Akzeptierte Negative** (`AGENTS.md` §3.15; Grund und Folge):

- Verzweigungen ohne Aufruf oder Stelle (siehe oben). Grund: die Zählung ist über den Quelltext des Pakets gebaut und zählt, was ein Ausdruck benennt. Folge: ein Zweig von `Diagnose`, den kein Fall führt (etwa ein dritter Weg für Schlüssel), fiele erst der Messung des Abdeckungsprofils und der Hand-Liste der Zweige auf.
- Die Erreichbarkeit von `ODR-E3001` und `ODR-E9001` gegen den echten SQLite-Adapter. Grund: der Fänger C3 führt jeden Eintrag der Abbildung durch den Fake und nicht den Adapter; ob `List` des Adapters `ErrBeschaedigt` oder `ErrUngueltig` liefert, gehört dem Vertrag des Adapters (`SPEC-107`). Folge: keine für den Dienst; liefert der Adapter einen Fehler nicht, ist der Code in `Coverage` unerreichbar, aber die Abbildung bleibt richtig.
- Der Inhalt der Evidenz (Nr. 6: nur Zahlen und Gründe, keinen Wert aus Request oder Response). Grund: der Import schreibt die Evidenz (`SPEC-127` Nr. 13), der Dienst reicht sie byte-gleich weiter und prüft sie nicht; ein Test des Dienstes könnte nur seine eigene Fixture lesen. Folge: ein Import, der einen Wert in die Evidenz schriebe, zeigte ihn über die Admin-API; die Zusage prüft der Fänger des Imports. Der Eintrag steht in `SPEC-236` *Ungeprüft*.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts.
- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.
- Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (`BEO-REPO/report-der-rolle-faellt-docs-check`, 5×); der Planner reicht den Satz an beide weiter.

**Entscheidungen des Architect für diesen Teil** (2026-11-07; Vertrag `SPEC-236`, Ausgang der Aufträge 1, 4, 5, 8 bis 10 und 13 von `slice-inspection-use-case`; eine ADR entsteht nicht):

1. **Coverage:** Parameter ist die Baseline-Kennung (HTTP `baseline_id`, Pflicht), nicht Workspace oder Service; keine Seitenbildung; unbekannt oder fremde Art ist `ODR-E7002` (`SPEC-236` Nr. 6).
2. **Diagnose:** `Modus` und `Einstellungen` kommen als fertiger Wert aus dem Root; Allowlist von 18 Schlüsseln, jeder andere wird verworfen; die Zähler sind ein Eintrag je Code von `model.AlleCodes()` aus `Snapshot().Errors` (fehlend 0); `TelemetryReader` reicht, die Anwendung weicht von `SPEC-230` Nr. 8 und 9 nicht ab; der Dienst prüft `ctx.Err()` einmal, weil kein Port der Diagnose einen Kontext nimmt (`SPEC-236` Nr. 7).
3. **Katalog:** kein neuer Fehlercode, Schlüssel oder Kommando, kein Eintrag in `harness/katalog-ausnahmen.txt`, kein Handbuch-Kapitel; das Binary bleibt unverändert (`SPEC-201` Nr. 3 und 11).
4. **Kennungen:** keine eigene; `SPEC-236` besteht und ändert dieser Slice nicht (eine Änderung des Vertrags geht an den Architect, eine an `SPEC-109` oder `SPEC-230` an den Planner).
5. **Prüfung beim Start** (Architect, 2026-11-07, Ausgang der Aufträge A bis M; Plan und Vertrag gegen die Entscheidungslage, bevor Code besteht; kein Konflikt mit einer ADR, die Schichtregeln des Architektur-Gates erlauben die Importe von `services` nach `model`, `ports-inbound`, `ports-outbound` und `ports-support`):
   - a) **Sätze (A).** Jeder Satz von `SPEC-236` Nr. 1, 2, 6, 7, 8 und des Absatzes vor Nr. 1 hat eine Zeile oder ein Negativ (Tabelle *Sätze der Nummern* unten). Neu: **C6** (Nullwert im Fehlerfall; der Planner hatte C3 und C4 als Fänger genannt, sie prüfen den Fehler, nicht das Ergebnis) und **K5** (Bindung); erweitert: C3 um Status und Gegenstand je Eintrag der Abbildung (die Zeile `ErrBereitsVorhanden` unterscheidet `Session` von `Snapshot`), C1 um die Reihenfolge Argument vor Kontext, C4 und C5 um die Zahl der Aufrufe bei Kontextfehler, D1 um das Ende des `ctx` in `Snapshot`, D2 um die Aufzeichnung statt einer Liste von Operationen, D5 um den leeren Wert und die Modus-Unabhängigkeit, W3 um Importe und Aufrufe von `ctxNil` und `ctx.Err()`. „`Diagnose` liefert keinen Fehler mit Code“ fangen A2 und D1 (`ErrUngueltig` und Kontextfehler ohne `*Fehler`); „ein Fehler von `Snapshot` gibt es nicht“ ist Typ des Ports (kein Fehler in der Signatur) und braucht keinen Test.
   - b) **Raster (B).** Die Zellen tragen Orte; die Zelle *Abbruch* von `Diagnose` sagte „gilt nicht“ für die Zeit nach `Snapshot`, obwohl `SPEC-236` Nr. 2 sie für jede Operation entscheidet: sie nennt jetzt den Ort, und D1 fängt ihn. Die Zusatztabelle *Fänger* ist für `Coverage` nötig (leere Kennung mit `ctx` nil oder beendet), für `Diagnose` nicht (kein Paar von Verstößen am selben Schlüssel, die Allowlist kennt eine Regel; der Modus ändert nichts, D5 hält es).
   - c) **Mengen (C).** M-Coverage führt m gleich 101 und 1001 statt 20 (20 fing weder das Kürzen auf `StandardLimit` noch auf `MaxLimit`); M-Schlüssel ist in M-Erlaubt und M-Ausgeschlossen geteilt; M-Codes rechnet mit `math.MaxInt` (der Typ `Code` ist `int`); die Zählprüfungen vergleichen gegen `AlleCodes()`, `EintraegeAbbildung()`, `Erlaubt()` gegen die Literal-Liste und die Literale gegen `model.Coverage*`, nie gegen 37 oder 38.
   - d) **Literale (D).** Die Tabelle *Literale des Vertrags* nennt je Zeile die Werte; kein Orakel nutzt `Erlaubt()` oder `model.Coverage*` als Erwartung.
   - e) **Nebenläufigkeit (E).** Die Zweig-Menge, die Runden (8 × 20 je Zweig, Ziel 20 von 20 je Prozess), der Wirkungs-Fänger (nummerierte Antworten) und die Messung des Abdeckungsprofils stehen oben; ein Gate für das Profil gibt es nicht (kein Ziel in `make`), Folge unter *Akzeptierte Negative*.
   - f) **Bindung (F).** K5; Nachweis ist die Übersetzungsmeldung. Die Zählprüfung der Methoden per `reflect` gehört nicht dazu (Begründung unter *Nachweis der Bindung*).
   - g) **Zeitfelder (G).** Gilt nicht: `CoverageEintrag{EntitySet, Status, Evidenz}` und `Diagnose{Modus, Einstellungen, Fehler}` tragen keine Zeitfelder, und `model.Coverage` auch nicht; die Fixtures brauchen keine Zeiten.
   - h) **Allowlist und Zähler (H).** Die 18 Schlüssel stehen alle in `SPEC-110` Nr. 2 (27 Schlüssel: 18 erlaubt, 9 einzeln ausgeschlossen); keiner fehlt dort, und jeder trägt Aufzählung, Wahrheitswert, Dauer, Größe oder Liste von Namen (`REDACT_JSON_PATHS` Pfade der Form `$..name`, `SPEC-110` Nr. 6), nie eine Adresse, einen Pfad oder ein Geheimnis. Die Zähler sind ein Eintrag je Code von `model.AlleCodes()`, die Zählprüfung hält mit dem 38. Code ohne Änderung.
   - i) **Fakes (I).** `FakeCoverage` liegt in `helfer_coverage_test.go`; Einbettung statt Hülle für `FakeTelemetry` (Abschnitt *Fakes*); der Gegenstand `Snapshot` hat eine Aufrufstelle (`List`), der Fall steht in C3.
   - j) **Nachzug (J).** Abschnitt *Nachzug*: ohne die Art *Aufruf von `ctxNil`* fängt die Zählung nach dem Zusatz nicht jede Stelle; `TestZweigeSindDieWege` wächst um die Wege beider Operationen.
   - k) **Marke und Block (K).** Siehe §4: 21 andere Zeilen und 1 nebenläufige, *a* gleich 1, *b* gleich 21, *f* gleich 1, *p* gleich 190, *S* gleich 2155, *M* gleich 3000, unter der Schnittgrenze. Der Block im Zweig steht ohne `basis`, `schaetzung` und `marke`.
   - l) **Katalog und Kennungen (L).** Bestätigt: kein Code, Schlüssel oder Kommando im Katalog des Binary, kein Handbuch-Kapitel, keine neue Kennung; der Slice schreibt weder `model/fehler.go` noch `ressource_mengen.go` noch `ports/inbound/mock.go` noch `services/mock/` noch `handler*.go` in `src/adapters/driven/telemetry`.
   - m) **Reports (M).** Der Satz steht unter den Betriebsregeln.

**Sätze der Nummern von `SPEC-236`** (Auftrag A; Architect): Satz in Kurzform, Zeile oder Negativ.

| Stelle | Satz | Zeile oder Negativ |
|---|---|---|
| Kopf | importiert Domain, Ports, Standardbibliothek, keinen Adapter, keinen Dienst; kennt weder HTTP noch Umgebung | W3 (Importe), `make a-check` für Adapter und Dienst |
| Kopf | verändert nichts: kein Schreiben in einen Port | C5, D2, X13 des Vorgängers |
| Kopf | kein Audit-Eintrag, kein Log, kein Zähler | D2 (Aufzeichnung), W3 (Importe, Aufruf über ein neues Feld von `Konfig`) |
| Kopf | Typen im Port, Konstanten und `Erlaubt` im Dienst | K5 (Übersetzung), D5 |
| Kopf | Ergebnis im Fehlerfall der Nullwert, bei `Coverage` `nil` | C6 |
| 1 | `New` weist nil in `Exchanges`, `Coverage`, `Zaehler` ab; Modus außerhalb der vier; `Einstellungen` darf nil sein | K1 (Vorgänger), K2, K3, K4 |
| 1 | `New` kopiert `Einstellungen` | K4 |
| 1 | ruft nur `ListBySession`, `List`, `Snapshot`; nie `Telemetry`, nie den Adapter | X13, C5, D2 |
| 1 | `Modus` und `Einstellungen` als fertiger Wert, nie die Umgebung | W3 (Import `os`) |
| 1 | Grund der Modus-Prüfung | Begründung ohne Verhalten; K3 hält die Folge |
| 2 | `ctx` nil ist `ErrUngueltig` ohne `Fehler`, kein Port | A1, A2 |
| 2 | Dienst prüft den Kontext nicht selbst außer in `Diagnose` | C4 (`List` 1 bei beendetem `ctx`), W3 (ein `ctx.Err()`) |
| 2 | Kontextfehler unverändert, ohne Code, auch neben einem Fehler der Tabelle | C4 |
| 2 | beendet an der Stelle, danach kein Port | C4, C5, D1 |
| 2 | `ctx` unverändert an den Port; Ende nach der Rückkehr ändert nichts | C4, D1 |
| 2 | Argumentprüfung geht dem Port voraus, `ErrUngueltig` ohne Wert im Text | C1; der Text trägt keinen Wert: gilt nicht, der einzige ungültige Wert der Kennung ist leer |
| 2 | HTTP-Antwort bestimmt der Adapter | `SPEC-236` *Ungeprüft* |
| 6 | Parameter ist die Baseline, kein Workspace, kein „alle“ | C1 (`w1` unverändert), C3 |
| 6 | leere Kennung `ErrUngueltig`; jede andere unverändert, genau einmal | C1 |
| 6 | Ergebnis je Eintrag, Reihenfolge, byte-gleich, nichts abgeleitet, Status nicht geprüft | C2 |
| 6 | Evidenz enthält nur Zahlen und Gründe | N2 (Import, `SPEC-127` Nr. 13) |
| 6 | Baseline ohne Einträge: leere Liste, nie `nil` | C2 |
| 6 | nicht seitenweise | C2 (101, 1001) |
| 6 | unbekannte Baseline und Snapshot anderer Art: `ODR-E7002` | C3 |
| 7 | `ctx` nil, dann `ctx.Err()` einmal, `Snapshot` nicht gerufen | A2, D1 |
| 7 | `Snapshot` genau einmal und sonst nichts | D2 |
| 7 | „ein Fehler dieses Aufrufs gibt es nicht“ | Typ von `Snapshot()` (keine Fehler-Rückgabe); kein Test nötig |
| 7 | `Modus` der Konfiguration | D7 |
| 7 | `Einstellungen` genau die erlaubten Schlüssel, unverändert, verworfen ohne Meldung, fehlend ohne Eintrag | D5, D6 |
| 7 | `Fehler` ein Eintrag je Code in der Reihenfolge, `Anzahl` oder 0, außerhalb verworfen, Zahl der Codes | D3, D4 |
| 7 | Kopie, nie `nil`, jeder Aufruf liest neu | D8 |
| 8 | `*inbound.Fehler` nach `SPEC-109` Nr. 5, Gegenstand `Snapshot`; Anwendung der Tabelle | C3 |
| 8 | `Status` ist der eine Status des Codes | C3 |
| 8 | erreichbar `ODR-E7002`, `ODR-E3001`, `ODR-E9001` samt `ErrUngueltig` des Ports | C3, N2 (Adapter) |
| 8 | `Diagnose` liefert keinen Fehler mit Code | A2, D1 |
| 8 | neue Codes führt der Vertrag nicht ein | Kataloge-Gate (DoD) |
| 10 | kein Zustand, keine Pufferung, kein Fenster | C2, D8, W4 |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-inspection-use-case` (Gerüst, Port, `Konfig`, `New`, `FakeRecording`, `SPEC-236` auf `main`) und `slice-telemetry-port` (`welle-betrieb`, `TelemetryReader`) liegen in `done/` (erfüllt: `slice-inspection-use-case` ist mit der Closure vom 2026-11-07 auf `main`, Messung 1408 von 2500, Auslastung 0,56). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-236`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** Der Slice erweitert Paket, Port und `Konfig` aus `slice-inspection-use-case` (*Code-Abhängigkeit*: `New`, `mitCode`, die Typen `Seite` und `Fehler`-Abbildung liegen dort) und liest die Zähler über `TelemetryReader` aus `slice-telemetry-port` (*Code-Abhängigkeit*, `src/hexagon/ports/support/`). `CoverageRepository` liegt seit `welle-persistenz` in `src/hexagon/ports/outbound/` und braucht nichts davon.

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`** (Mock, `WorkspaceUseCase`, HTTP-Adapter, Composition Root): der Dienst kennt weder HTTP noch einen Modus als Verhalten, nur als Wert der Konfiguration.

**Größenmarke** (Stand beim Start, Architect, 2026-11-07, `SPEC-128` Nr. 1): Die Rechnung aus der Tabelle in §3 ergibt 22 Zeilen, davon 1 nebenläufig (W4) und 21 andere (K2 bis K5, A2, C1 bis C6, D1 bis D8, N2, W3), dazu `fakes` 1 (`FakeCoverage`; `FakeTelemetry` besteht seit `slice-telemetry-port`) und `produktcode` 190 (Methoden und Typen des Ports, `Coverage`, `Diagnose`, `Erlaubt`, Felder von `Konfig`): *S* = 85·1 + 80·21 + 200·1 + 190 = 2155, *M* = 3000; keine Schnittfrage (*S* unter 2223, Abstand 68). Die Prosa-Marke des Planner vom Schnitt (*S* 1995, *M* 2700, 19 andere Zeilen) wuchs um K5 und C6 (Architect, Entscheidung 5 a); *a* = 1, *b* = 21, *f* = 1, *p* = 190. Der Architect hat den Block im Zweig angelegt, ohne `basis`, `schaetzung` und `marke` (Zustand *Marke ausstehend*, `SPEC-128` Nr. 3 p); der Planner setzt die drei Zeilen nach seiner Bestätigung in eigenem Commit vor dem ersten Code-Commit (`AGENTS.md` §3.14): `basis` ist der Commit, der Stand und Tabelle trägt, `schaetzung` 2155, `marke` 3000. Die Marke gilt für die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Teil 1 endete mit Auslastung 0,56 (Messung 1408, Schätzung 1845); der Architect legt keine Neukalibrierung fest (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`, 19×; ein Vorschlag geht an den User).

```groessenmarke
pfad: src/hexagon/ports/inbound/inspection.go
pfad: src/hexagon/services/inspection/
produktcode: 190
fakes: 1
```

**Stand der Bedingungen beim Start** (Planner, 2026-11-07, `main` = `6ff4af7`): `slice-inspection-use-case` und `slice-telemetry-port` liegen in `done/` (`TelemetryReader` mit `Snapshot().Errors` besteht in `src/hexagon/ports/support/telemetry.go`); `SPEC-236` steht in der Spezifikation; Nr. 7 der Übergangsregeln (Gate-Kette) ist nicht berührt, der Slice ändert `Makefile`, `harness/mk/`, `Dockerfile`, `.a-check.yml` und `.golangci.yml` nicht. Speicher beim Start 13 GB verfügbar (`free -g`), Last 19,7 auf dem Host; schwere Läufe laufen unter `flock`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); die Rechnung des Architect oben ist der Stand vor dem Start).
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
| Coverage lesen (Baseline) | `ctx` nil: `ErrUngueltig` (Nr. 2); es gibt keinen Standard und kein „alle Baselines“ (Nr. 6) | `baselineID` leer: `ErrUngueltig` ohne Port (Nr. 6); Baseline ohne Einträge: leere Liste (Nr. 6) | unbekannte Baseline und Kennung eines Snapshots anderer Art: `ODR-E7002` (Nr. 6, 8); Status oder Evidenz: unverändert (Nr. 6) | vor dem Aufruf des Ports: sein Kontextfehler unverändert; nach der Rückkehr: Ergebnis, keine Prüfung (Nr. 2) | zustandslos, kein Zwischenspeicher, jeder Aufruf ruft `List` neu (Nr. 10); eine Baseline ist unveränderlich (`SPEC-107`), das Ergebnis gleich |
| Diagnose lesen (Konfiguration, Modus, Zähler) | `ctx` nil: `ErrUngueltig` (Nr. 2); `Einstellungen` nil: `{}` (Nr. 1, 7); Code ohne Zähler: `Anzahl` 0 (Nr. 7) | `Errors` leer oder nil: jeder Code mit 0 (Nr. 7); `Einstellungen` leer: wie nil (Nr. 1, 7); erlaubter Schlüssel mit leerem Wert: bleibt, unverändert (Nr. 7); `Modus` leer: `New` scheitert (Nr. 1) | Schlüssel außerhalb der Allowlist: verworfen (Nr. 7); Eintrag außerhalb der Tabelle: verworfen (Nr. 7); ungültiger Modus: `New` scheitert (Nr. 1) | beendeter `ctx` vor `Snapshot`: sein Fehler, kein `Snapshot` (Nr. 2, 7); endet der `ctx` in oder nach `Snapshot`: Ergebnis, keine Prüfung danach (Nr. 2) | jeder Aufruf liest den Stand neu, die Zähler wachsen (Nr. 7, 10) |

**Zusatztabelle *Fänger*** (`AGENTS.md` §3.16; Paare von Verstößen und Modus):

| Paar oder Modus | Entscheidung | Fänger |
|---|---|---|
| `Coverage`: `ctx` nil und leere Kennung | beide sind `ErrUngueltig` ohne Code, nicht zu unterscheiden; kein Port (Nr. 2) | A2 (Fall mit leerer Kennung) |
| `Coverage`: leere Kennung und beendeter `ctx` | `ErrUngueltig`, nicht der Kontextfehler, `List` 0: die Prüfung der Argumente geht jedem Aufruf eines Ports voraus (Nr. 2) | C1 (Kennung × `ctx`-Zustand) |
| `Coverage`: Kontextfehler und Fehler der Tabelle in einem Fehler | der Kontextfehler, unverändert und ohne Code (Nr. 2) | C4 (Join in beiden Reihenfolgen) |
| `Diagnose`: Modus × `Einstellungen` | die Allowlist ist vom Modus unabhängig; der Modus ist ein eigenes Feld (Nr. 7) | D5 (vier Modi), D7 |
| `Diagnose`: Paare von Verstößen am selben Schlüssel | gilt nicht: ein Schlüssel hat genau eine Regel (byte-gleich in `Erlaubt()` oder verworfen), `Diagnose` meldet keinen Verstoß | – |

- Lesen der Zähler: `slice-telemetry-port` (`done/`) hat entschieden, dass `TelemetryReader` mit `Snapshot().Errors` liest, nicht `Telemetry` (`SPEC-230` Nr. 9, Sendung des Gebers an `slice-inspection-use-case`, von dort an diesen Slice weitergegeben); dieser Slice trägt sie in Ziel, DoD und §4 (`AGENTS.md` §3.13) — **Ausgang:** entfallen (die Frage nach einer Leseoperation ist entschieden, Form und Fehlerfall stehen in `SPEC-236` Nr. 7).
- Die Parameter von `GET /__admin/coverage` und der Code für eine unbekannte Baseline — **Ausgang:** entfallen; entschieden in `SPEC-236` Nr. 6, 8 und in `SPEC-005`.
- Geheimnisse in der Diagnose: die Allowlist der Schlüssel (`SPEC-236` Nr. 7) ist der Schutz, der Fänger ist D6; der Fänger über alle Artefakte ist der Canary-Test in `slice-test-security` — **Ausgang:** weiter offen bis zur Closure; Adresse: Implementer (D6), danach Reviewer.
- Die Zählprüfungen (W3 und die der Mengen) und der Bericht je Befund mit Menge, Zahl der Elemente und einer Mutation an einem Element, das der Befund nicht nannte (`AGENTS.md` §3.17) — **Ausgang:** weiter offen; Adresse: Implementer.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): `FakeCoverage` und `FakeTelemetry` tragen jede Schnittstelle des Originals und zählen jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `SPEC-109` Nr. 5, Absatz *Anwendung*, nennt `InspectionUseCase` nicht; der Satz für `Coverage` (Gegenstand `Snapshot`, `SPEC-236` Nr. 8) gehört zu diesem Teil, aber der Zweig ändert `SPEC-109` nicht (`slice-mock-zuordnung`, Linie B, ändert dieselbe Stelle) — **Ausgang:** offen; Adresse: Planner im Integrationsschritt (Merge-Commit), nach dem Satz für `Exchanges` aus `slice-inspection-use-case`.
- `SPEC-236` steht mit dem Merge von `slice-inspection-use-case` auf `main` für den ganzen Vertrag; bis zur Closure dieses Slice liefert der Code `Coverage` und `Diagnose` nicht — **Ausgang:** weiter offen bis zur Closure dieses Slice; Adresse: Implementer (kein Gate prüft die Lücke zwischen Vertrag und Code; sie steht allein hier und in §1, `SPEC-236` *Ungeprüft* nennt sie nicht).
- Kennung `SPEC-236`: `slice-telemetry-log` (Zweig) vergibt in `spezifikation.md` bisher `SPEC-230` bis `SPEC-232` und keine `SPEC-236`; vergibt es später doch `SPEC-236`, stößt es beim zweiten Merge zusammen: der zweite Merger nummeriert seine Kennung auf die nächste freie Nummer der Linie D um (ab `SPEC-237`) und zieht alle Stellen nach (`AGENTS.md` §3.9); derselbe Hinweis steht in `slice-inspection-use-case` §6 — **Ausgang:** weiter offen bis zum Merge von `slice-inspection-use-case`; Adresse: Planner (Integrationsschritt).
- Größe: die Rechnung des Architect steht in §4 (*S* 2155, *M* 3000; die Prosa-Marke des Planner vom Schnitt war *S* 1995, *M* 2700, der Unterschied sind K5 und C6); der Block steht im Zweig ohne `basis`, `schaetzung` und `marke` — **Ausgang:** offen; Adresse: Planner (die drei Zeilen setzen, `schaetzung` 2155 und `marke` 3000 nach `SPEC-128` Nr. 1, in eigenem Commit vor dem ersten Code-Commit).
- Die Tests des Vorgängers bauen `Konfig` mit `Exchanges` allein; mit dem Zusatz weist `New` sie ab (Abschnitt *Nachzug*, Punkt 1). Der Nachzug gehört in den Commit des Codes — **Ausgang:** weiter offen bis zur Closure; Adresse: Implementer.
- Aufträge A bis M an den Architect zum Start (Start-Commit): Sätze je Vertragsnummer, Raster, Mengen und Zählprüfung gegen die Quellen, Literale des Vertrags im Orakel, `TestGleichzeitig` als Schleife mit Runden, Bindung an den Port, Zeitfelder, Allowlist und Zähler, zählende Fakes, Nachzug von W1, Marke, Katalog, Reports — **Ausgang:** entfallen (Architect, Entscheidung 5; die Ergebnisse stehen in §3, die Rückgabe der Marke an den Planner in der Zeile *Größe*).
- Pfade neben den laufenden Zweigen (Planner, Start 2026-11-07): `slice-mock-lesen` schreibt `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/mock.go`, `src/hexagon/model/fehler.go`, `src/hexagon/model/ressource_mengen.go` und Testdateien in `src/adapters/driven/telemetry` und `src/adapters/driven/audit`; `slice-telemetry-handler` schreibt neue Dateien `handler*.go` in `src/adapters/driven/telemetry`; dieser Slice schreibt `src/hexagon/ports/inbound/inspection.go` und `src/hexagon/services/inspection/` additiv (geprüft: `git diff --name-only main...<Zweig>` für beide Zweige berührt keinen dieser zwei Pfade). Der Dienst liest `model.AlleCodes()`: die Zählprüfung gegen die Quelle hält mit dem 38. Code (`ODR-E6008`) ohne Änderung der Tests. Berührt dieser Slice eine Datei der Nachbarn, ist das eine Abweichung an den Planner — **Ausgang:** offen; Adresse: Integrationsschritt (der zweite Merge zieht `main` nach).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet und beim Start nachgezogen (Planner, 2026-11-07, `main` = `6ff4af7`; Zähler = Zahl der Dateien in `evidence/`). Zähler beim Start: `zweig-ohne-lauf-unter-dem-detektor` 10×, `betriebsregel-fuer-repo-dateien-verletzt` 9×, `schaetzformel-ueberschaetzt-den-slice` 19×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `fix-deckt-die-genannte-mutation-nicht-die-klasse` 31×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 43×, `akzeptiertes-negativ-ohne-folge` 14×, `randform-erst-im-review-entschieden` 34×, `universalaussage-im-vertrag-ungeprueft` 14×, `suite-ohne-generator-fuer-klassen-der-spezifikation` 6×, `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11×, `report-der-rolle-faellt-docs-check` 5×; neu seit `slice-inspection-use-case`: `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (verkörpert, Auftrag D), `BEO-REPO/zusage-ueber-verzweigungen-reicher-als-die-funktionsliste` (2×, für W3: die Zählprüfung zählt Aufrufe und Stellen, keine Verzweigungen; benanntes Negativ in Teil 1) und `BEO-REPO/dienst-erfuellt-den-port-ohne-bindung` (1×, Auftrag F). Offen und einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (10×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure der Welle, Diff-Vorlage an den User) für die Zweig-Menge (W3, W4, Auftrag E); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (19×) für die Marke. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×) für `FakeCoverage` und `FakeTelemetry`; `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (29×, `AGENTS.md` §3.17), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14) für die Mengen; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, beide `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.16, für „ohne Geheimnisse“). Mit diesem Slice erreicht kein offener Eintrag mit weniger als 3× die Schwelle.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
