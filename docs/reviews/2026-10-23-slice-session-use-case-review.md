# Review-Report: slice-session-use-case — 2026-10-23

**Review-Art:** Code — geprüft gegen Plan (`slice-session-use-case` §1, §3 mit der Tabelle *Zusage · Fänger* V1 bis V20 und dem Absatz *Zusage · Mutation · roter Test* samt Raten, §4 Marke 2300, §6 mit Raster und Zusatztabelle), Vertrag (`SPEC-124` Nr. 1 bis 10 und *Ungeprüft*, die Nachbarn `SPEC-105` Nr. 1, 3, 6, 7, `SPEC-109` Nr. 5 und 6, `SPEC-123` Nr. 2, 8, 10) und Hard Rules (`AGENTS.md` §3.7 bis §3.16). Nicht gegen die DoD: das ist die Frage des Verifiers.

**Gegenstand:** `slice-session-use-case` (Welle `welle-record-replay`), `git diff eae82b8..9bdd4bd`. Implementer: `bf83ddc`, `c1fdc2d`, `9bdd4bd`. Der Diff berührt `src/hexagon/ports/inbound/session.go`, `src/hexagon/services/session/` (Produktcode `verwaltung.go`, `loeschen.go`, `protokoll.go`, `fehler.go`; Tests `helfer_test.go`, `anlegen_test.go`, `konfig_test.go`, `kontext_test.go`, `lesen_test.go`, `loeschen_test.go`, `nebenlaeufig_test.go`) und den Plan (§3, 39 Zeilen).

**Skill:** `.harness/skills/reviewer.md` @ HEAD `9bdd4bd` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-23

**Methode.** Jede Zeile V1 bis V20, jede Zelle des Rasters und jede Zeile der Zusatztabelle ist mit eigenen Mutationen aufgelöst; was der Plan an Mutationen nennt, habe ich nachgefahren, nicht übernommen, und um Mutationen ergänzt, die der Plan nicht nennt. 272 Läufe (2 ungültig: ein Baufehler, ein Muster ohne Treffer), je eine Änderung in einer Wegwerfkopie aus `git archive 9bdd4bd -- go.mod go.sum src` (kein `.git`, nie im Repo), gefahren in `odata-recorder:race` ohne Netz (`docker run --network none`, `go test -count=1` der Pakete `services/session`, für V19 mit `-race`). Ein Go-Programm in der Wegwerf-Ablage setzt die Änderungen ein (Textersetzung mit genau einem Treffer) und liest die roten Tests; kein `sed`, `awk`, `perl` und kein Interpreter berührt das Repo. Die Raten habe ich mit übersetzten Testprogrammen gemessen (`go test -c`, mit und ohne `-race`), je Prozess `-test.count=1 -test.run '^<Fänger>$'`. Der Bericht des Implementers lag mir nicht vor; sein Absatz in §3 des Plans schon.

**Eingangs-Kontext:**

- Slice-Plan `slice-session-use-case` in `9bdd4bd` (§1 bis §4, §6, §8); `SPEC-124` in `spec/spezifikation.md`
- [ADR-0008](../plan/adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)
- [LH-FA-13](../../spec/lastenheft.md#lh-fa-13--management-und-diagnose-muss), [LH-RB-04](../../spec/lastenheft.md#lh-rb-04--schutz-von-daten-und-zugangsdaten), [LH-FA-03](../../spec/lastenheft.md#lh-fa-03--persistenz-und-wiederverwendung-muss), [LH-QA-03](../../spec/lastenheft.md#lh-qa-03--isolation-und-konsistenz)
- Vorheriger Report am Modul: `2026-10-23-slice-session-wartung-review` (Klassen „Allaussage über eine Menge ohne Fänger je Element“, „Gegenstand ohne Wirkung“)
- `AGENTS.md` (Hard Rules)

---

## Antworten auf die Fragen des Auftrags

**(1) Tabelle V1 bis V20.** Jede Zeile hat einen Test unter dem Namen der Tabelle (20 Tests, alle vorhanden). Jede Mutation, die der Plan nennt, ist aus dem genannten Grund rot, mit zwei Anmerkungen: V2 *Prüfung in `Create` entfernt* wird beim Lauf des ganzen Pakets durch eine Panik in `TestOpenInitialSession` rot, nicht in `TestCtxNil` (F-7); V19 ohne Detektor bleibt grün, wie der Plan sagt. Lücken gegenüber dem Wortlaut der Zusagen stehen als F-1 bis F-4 und F-8 (jeweils eine Mutation, die grün bleibt). Die Einzelergebnisse stehen in der Tabelle *Mutationen je Zeile*.

**(2) Raster (§6) und Zusatztabelle (27 Fälle).** Jede Zelle nennt einen Ort, dessen Wortlaut die Entscheidung trägt; der Code trifft sie so, und es gibt **keine im Code entschiedene Zelle**. Zellen mit Lücke im Fänger: *jede Operation · `ctx` · fremd* (Weitergabe an die Ports, F-1), *Delete · Bereinigung · leer* (Bericht aus Nullen, F-3), *Create · sonst kein Port* und *`New` · `Dienste` dient allein `List`* (F-2). Von den 27 Zeilen der Zusatztabelle sind 20 mit mindestens einer roten Mutation aufgelöst, 1 (Nr. 3) bleibt grün (F-2) (siehe die Tabelle *Zusatztabelle*); 2 tragen laut Plan keinen Fänger hier (*`OpenInitialSession` im Modus außer `record`*, *Konfig wächst additiv*: Träger `slice-bootstrap-m2` und `slice-session-beenden`, ihre §1 und DoD tragen die Kennung dieses Slice), 1 ist „gilt nicht“, 3 sind akzeptierte Negative mit Grund und Folge (Nr. 14, 15, 17). Die Allaussage *alle übrigen Operationen schreiben nichts* hat keinen Fänger über alle Pfade (F-4).

**(3) Nebenläufige Fänger V12, V13, V17, V18, V19.** Deterministisch und unabhängig vom Detektor sind V12, V13, V17, V18: jede Mutation rot in jedem Prozess; V19 ist rot genau unter dem Detektor. Die Zahlen unten bestätigen die des Implementers.

**(4) Vertrag gegen Code.** `SPEC-124` Nr. 1 bis 10 gleichen dem Code: nil-Prüfung der sechs Felder; `ctx` nil vor jedem Port, ohne `Now()` und `SessionID()`; der Kontextfehler wird nicht selbst geprüft, `outbound.Abbilden` liefert für ihn kein Code (`X_NurCanceledErkannt` ist rot); Prüfreihenfolge nil-`ctx`, `serviceID`, `id`, dann Kennung und Zeit je einmal vor `Sessions.Create`, auch bei dessen Fehler; Form der Client-`id` (alle 256 Bytewerte an Anfang, Mitte, Ende; 64 und 65); `OpenInitialSession` ist `Create(…, "")`; `Get` reicht die Kennung unverändert; `List` mit `Dienste.List` einmal und `Sessions.List` je Service einmal, `[]` statt `nil`, erster Fehler ohne Teilliste, `ErrNichtGefunden` aus `Sessions.List` als `ODR-E9001`; `Delete` mit der Rangfolge unbekannt, offen, referenziert, `CleanupBlobs` genau einmal nach `Sessions.Delete` mit dem `ctx` unverändert und synchron, Bericht unverändert und Fehler `nil`; Protokoll mit festen Attributen, ohne Fehlertext, nicht bei Kontextfehler; kein Zustand im Dienst. Abweichungen vom Wortlaut: keine. Ob die Reihenfolge von `Now()` und `SessionID()` zugesagt ist: der Vertrag sagt „Kennung und Zeit werden vor `Sessions.Create` bezogen“ und nennt keine Reihenfolge; die Vertauschung ist grün (I-3).

**(5) Kommentare (§3.7, §3.11).** Produktcode und Tests tragen Zusage, Kopplung oder Abgrenzung. Suche nach `wäre`, `hätte`, `würde`, `früher`, `bisher`, `ursprünglich`, `stattdessen`, `nicht mehr`, `jetzt`, `TODO`, `FIXME` in `services/session` und `session.go`: kein Treffer. Kommentare, die mehr zusagen, als ein Fänger prüft: `// Dienste dient allein List` (`verwaltung.go`, F-2) und `Aufrufzähler aller Fakes 0` in `TestCtxNil` und `TestCreateArgumente` (zählt `Dienste.Get`, `Dienste.Create`, `Dienste.SetMetadata` nicht, F-2).

**(6) Imports (V20).** `session` importiert Domain, Ports (`inbound`, `outbound`, `support`) und die Standardbibliothek. Die Mutationen *neuer Import* von `services/maintenance`, `services/record`, dem SQLite-Adapter, `os`, `os/exec`, `net`, `net/http`, `crypto/tls`, `net/url`, `modernc.org/sqlite`, `supporttest`, in einer zweiten und in einer dritten Datei: alle rot in `TestImporte`. Die Gegenprobe des Tests fängt zwei der vier Elemente der Standardbibliothek (F-5).

**(7) Marke und Metrik (§3.14).** `git diff --numstat eae82b8..HEAD -- src/hexagon/ports/inbound src/hexagon/services/session` ergibt 2107 Einfügungen und 0 Löschungen, Summe **2107**; die Pfade des Plans sind vollständig (außerhalb: nur der Plan). Marke 2300, Schätzung 1825: der Handoff liegt 15,5 % über der Schätzung und bei 91,6 % der Marke (I-1).

**(8) Bestand.** `git diff eae82b8..HEAD` außerhalb der Pfade des Plans berührt nur den Plan. `make gates` läuft nach dem Commit dieses Reports; das Ergebnis steht im Hand-back, nicht hier.

**(9) Adressaten.** F-1 bis F-4: Implementer (Fänger); F-2 und F-4 zugleich Planner (Zeile der Tabelle); F-6: Planner; F-5, F-7: Implementer; I-4: Architect.

---

## Mutationen je Zeile

„rot“ nennt den Test, an dem die Mutation rot wird; „grün“ heißt: kein Test wurde rot. Kennungen sind die Namen der Mutationen in der Wegwerfkopie. „Alle“ heißt: jede Mutation der Zelle rot.

| Zeile | Mutationen | Ergebnis |
|---|---|---|
| V1 | je eine der sechs Prüfungen entfernt; `New` liefert Dienst trotz Fehler; Fehler mit Code | rot `TestNewArgumente` (8 von 8) |
| V2 | Prüfung entfernt in `Create`, `Get`, `List`, `Delete`; Prüfung von `Create` hinter `Now()`; hinter `SessionID()`; nil-`ctx` als `ODR-E9001` | rot `TestCtxNil` (7 von 7); `V2_Create` im ganzen Paket zuerst durch Panik in `TestOpenInitialSession` (F-7), mit `-run TestCtxNil` in `TestCtxNil` |
| V3 | Kontextfehler als `ODR-E9001`; `ODR-E7004`, `ODR-E7010` vor dem Kontextfehler; `DeadlineExceeded` mit Code; `Canceled` an `Sessions.Create`, `Sessions.Get`, `Dienste.List`, `Sessions.List` (zweiter Service), `Sessions.Get` und `Sessions.Delete` von `Delete` übergangen; `outbound.Abbilden` erkennt nur `Canceled` (`X_NurCanceledErkannt`) | rot `TestKontext` (alle 11); die beiden übergangenen Stellen von `Delete` zusätzlich `TestBereinigungNachCommit` |
| V4 | Prüfung der `id` oder von `serviceID` entfernt; je hinter `Now()`/`SessionID()`; `id` im Text; je mit Code | rot `TestCreateArgumente` (alle 8) |
| V5 | `Now()` zweimal, nach Fehler, nach Erfolg; `SessionID()` zweimal, bei gelieferter `id`; `CreatedAt` plus 1 ns; `ServiceID` vertauscht; Status `complete`, `incomplete`, ungültig; Ergebnis mit anderer `ID`; `Get` vor `Create`; `SetStatus` auf offene Session; `CleanupBlobs` gerufen; `Dienste.List` gerufen | rot `TestCreate` (15 von 17) |
| V5 | `Dienste.Get` vor `Create` gerufen (`V5_DiensteGetGerufen`, `R_CreateDienstGetVorAnlegenUnbekannt`) | **grün** (F-2) |
| V5 | `Now()` vor `SessionID()` bezogen | grün, kein Vertragsbruch (I-3) |
| V5 | Status `complete`, `incomplete`, `CreatedAt` plus 1 ns **mit** `Validate` aus dem Fake entfernt | rot `TestCreate` durch den Vergleich mit der erwarteten Session, ohne den Fake (I-2) |
| V6 | `MaxIDLaenge` 63, 65; `>=`; Untergrenze 2; leere `id` ungültig; nur erstes, nur letztes, nur mittlere Bytes; Alphabet um `0x3a`, `0x2f`, `0x60`, `0x7b`, `0x2e`, `0x40`, `0x5b`, `0x2c`, `0x5f`, `0x20`, `0x00`, `0x7f`, `0x80`, `0xff`, `A` bis `Z`, nur `Z` erweitert; ohne `-`, `0`, `9`, `a`, `z` | rot `TestIDForm` (oder `TestCreateArgumente`), 29 von 30 |
| V6 | Prüfung über `range` der Runen mit `byte(r)` | grün (I-7) |
| V7 | Gegenstand `Sonstiges` bei `Create`; `ErrDienstUnbekannt` als `ODR-E7003`; `ErrUngueltig` ohne Code; `ErrBeschaedigt` als `ODR-E9001`; zweites `Create` überschreibt; `ErrBereitsVorhanden` als `nil`; Status 500 bei `ODR-E7010`; Jahr 10000 ohne `Validate` im Fake | rot `TestCreateFehler` (alle 8) |
| V8 | anderer Status; zweite Kennung; ohne Prüfung von `serviceID`; feste `id`; ohne `Now()`; anderer Fehler bei nil-`ctx` | rot `TestOpenInitialSession` (6 von 6); „Formprüfung der leeren `id`“ gleichwertig (leere `id` wird nie geprüft), grün |
| V9 | Formprüfung der Kennung; nur `open` geliefert; Status immer `open`; Gegenstand `Sonstiges`; `Now()`; `SessionID()`; Wartung; zweimal gelesen; `strings.ToLower(id)` | rot `TestGet` (alle 9) |
| V10 | nur der erste Service; nach Zeit über Services sortiert; ein Service ohne Session beendet die Liste; Teilliste bei Fehler; `nil` statt `[]`; Gegenstand `Session` bei `Sessions.List`; `Dienste.List` zweimal; `Sessions.List` zweimal; Sessions `incomplete` ausgelassen; Services rückwärts; Wartung gerufen | rot `TestList` (11 von 12) |
| V10 | Gegenstand `Session` bei `Dienste.List` | **grün** (F-8) |
| V11 | Statusprüfung entfernt; hinter `Sessions.Delete`; `!= complete`; unbekannter Status abgewiesen; `ODR-E7004` vor `ODR-E7009`; `ErrReferenziert` als `ODR-E9001`; Bericht bei `open`, bei `Get`-Fehler, bei `Delete`-Fehler nicht der Nullwert; Formprüfung der Kennung; leere Kennung ungültig; Gegenstand `Sonstiges` bei `Get` und bei `Delete`; `ODR-E7009` mit Ursache, mit Status 500; zweites `Delete` liefert `nil`; `ToLower` an `Get` und an `Delete`; leere Kennung bei `Get` ungültig | rot `TestDelete` oder `TestGet` (alle 19); `s.ID+""` an `Sessions.Delete` gleichwertig, grün |
| V12 | `CleanupBlobs` vor `Sessions.Delete`; nach `Delete`-Fehler; bei `open`; nach `Get`-Fehler; zweimal; `Startup` in `Delete`, `Create`, `List`, `Get`, in jedem Fehlerzweig, im Zweig nil-`ctx`, in `New`, hinter dem Protokoll; `CleanupBlobs` in `Create`-Fehler, `List`-Fehler, Argumentfehler, Zweig nil-`ctx` | rot (alle 21): `TestBereinigungNachCommit`, `TestDelete`, `TestKontext` oder durch den Wächter am Ende von `neu` (`Startup n-mal gerufen`) |
| V13 | `context.Background()`; `context.WithoutCancel(ctx)`; eigener `WithTimeout`; Goroutine ohne Warten; keine Bereinigung bei beendetem `ctx`; `ctx` durch einen Wert ersetzt | rot `TestBereinigungKontext` (alle 6) |
| V14 | Fehler weitergegeben (allgemein; nur `ODR-E3001`, nur `ODR-E9001`, nur ohne `Fehler`-Wert, nur Kontextfehler); Bericht bei Fehler genullt; `Zurueckgestellt` als Fehler; `Entfernt` plus 1; `Zurueckgestellt` verloren | rot `TestBereinigungErgebnis` (alle 9) |
| V14 | Bericht aus Nullen in `{Entfernt: 1}` oder `{Zurueckgestellt: true}` umgeschrieben | **grün** (F-3) |
| V15 | Attribut fehlt, ein Attribut mehr, doppelt, vertauscht, `service_id` falsch, `session_id` falsch; Fehlertext im Attribut und im Text; Stufe `WARN` statt `ERROR`, `ERROR` statt `INFO`; Eintrag bei Erfolg, bei Kontextfehler; nur `Canceled`, nur `DeadlineExceeded` als Kontextfehler; `error_code` bei `Zurueckgestellt`; Fehler ohne `*Fehler` ohne Eintrag; Code immer `ODR-E9001`; Eintrag für `ODR-E7004`, `ODR-E7009`, `ODR-E7010` (`Create`-Fehler), `ODR-E9001` (`Get`-Fehler) | rot `TestProtokoll` (alle 21) |
| V15 | Eintrag bei Erfolg von `Create`, von `Get`; Eintrag in `List` (Erfolg, Fehler von `Dienste.List`, Fehler von `Sessions.List`); Eintrag bei Argumentfehler von `Create`; Eintrag bei nil-`ctx` | **grün** (F-4, 7 von 7) |
| V15 | `ctx` des Protokolls `context.Background()` oder `nil` | grün, `ctx` des Eintrags ist nicht zugesagt (I-3) |
| V16 | Status 500 für jeden Code; `Ursache` nicht gesetzt; `ODR-E7009` mit Ursache, mit Status 500; Ursache durch fremden Fehler ersetzt | rot `TestFehlerWert` (alle 5) |
| V17 | `Get` vor dem Anlegen; Sperre um das Anlegen; `ErrBereitsVorhanden` als `nil`; zweiter Versuch mit anderer `id` | rot `TestAnlegenGleichzeitig` (alle 4; die Sperre nach 5 s Barriere) |
| V18 | der Verlierer ruft die Bereinigung; liefert `ODR-E9001`; der Gewinner ruft sie nicht; Sperre um `Delete`; Löschen ohne `Get` | rot `TestLoeschenGleichzeitig` (alle 5; die Sperre nach 5 s Barriere) |
| V19 | Feld mit der zuletzt angelegten Kennung (`Create`); Zähler in `Delete`, in `Get`, in `List` | unter `-race` rot `TestGleichzeitig` (`WARNING: DATA RACE`, alle 4); das Feld ohne Detektor grün, wie der Plan sagt |
| V20 | Importe wie in Antwort (6) (13 Mutationen) | rot `TestImporte` (alle 13) |
| V20 | Mutationen des Matchers im Test: `net` (genau) entfernt; `crypto/tls` entfernt; Präfix `os/` entfernt | grün (F-5); Präfix `net/` entfernt und `ports/support` aus der erlaubten Liste rot |

## Raster

| Zelle | Mutation | Ergebnis |
|---|---|---|
| `New` · fehlend (6 Felder) | V1 | rot |
| `New` · Wiederholung (zwei Dienste über denselben Ports) | `New` verweigert den zweiten Aufruf (Paketvariable) | rot (`TestNewArgumente` und alle Tests, die zweimal bauen) |
| `ctx` · fehlend | V2 | rot |
| `ctx` · fremd (Wert, Frist unverändert an Ports und Bereinigung) | `Background()` an `Sessions.Create`, `Sessions.Get`: rot; an `Dienste.List`, `Sessions.List` (auch `WithoutCancel`), `Sessions.Get` von `Delete`, `Sessions.Delete` (auch `WithoutCancel`): **grün** | Bereinigung rot (V13), vier der sechs Port-Aufrufe grün (F-1) |
| `ctx` · Abbruch | V3 | rot |
| `Create` · `serviceID` leer, unbekannt, zweimal derselbe Service | V4, V7, `Create` weist bei offener Session des Service ab | rot |
| `Create` · `id` leer, gültig, außerhalb der Form, Server-Form, zweimal | V4, V5, V6, V7, V17 | rot |
| `OpenInitialSession` · alle Zellen wie `Create`; Modus | V8; Modus: `Konfig` mit Pflichtfeld `Modus` rot, ein Verbot im Modus `replay` nicht ausdrückbar (die Konfig des Tests hat kein Feld; kein Fänger hier, Träger `slice-bootstrap-m2`) | rot / benannt |
| `Get` · `id` leer, unbekannt, außerhalb der Form, jeder Status | V9, V11 (leere Kennung ungültig) | rot |
| `List` · leer, Service ohne Session, alle Status, erster Fehler | V10 | rot |
| `Delete` · `id` leer, unbekannt, außerhalb der Form, zweites `Delete`, gleichzeitig | V11, V18 | rot |
| `Delete` · Status (`open`, unbekannter Statuswert, `complete`, `incomplete`, offen und referenziert) | V11 | rot |
| `Delete` · Bereinigung (leer, `Zurueckgestellt`, Fehler mit Code, ohne `Fehler`-Wert, Kontextfehler, beendeter `ctx`) | V13, V14, V15; **leer** (Bericht aus Nullen) | rot; Bericht aus Nullen **grün** (F-3) |

## Zusatztabelle (27 Fälle)

| # | Fall | Mutation | Ergebnis |
|---|---|---|---|
| 1 | Anlegen in den Modi | `Konfig` mit Pflichtfeld `Modus` | rot (alle Tests); Verbot im Modus `replay` nicht ausdrückbar |
| 2 | `OpenInitialSession` im Modus außer `record` | — | kein Fänger hier, Träger `slice-bootstrap-m2` |
| 3 | Doppelte `id` × unbekannter Service | `Create` ruft `Dienste.Get` vorab | **grün** (F-2); die Rangfolge selbst fängt `outboundtest` |
| 4 | Ungültige `id` × leerer `serviceID` | V4 je Fall | rot (Folge nicht unterscheidbar, wie der Vertrag sagt) |
| 5 | Offen × Baseline-Quelle | `ODR-E7004` vor `ODR-E7009` | rot `TestDelete` |
| 6 | Unbekannt × offen | — | gilt nicht |
| 7 | Fehler der Bereinigung × Ergebnis | V14 | rot |
| 8 | `Zurueckgestellt` × Ergebnis | V14 | rot |
| 9 | Kontextfehler × Fehler der Tabelle | V3 `ODR-E7004`, `ODR-E7010` | rot `TestKontext` (c) |
| 10 | Kontext nach dem Commit beendet | V13 `WithoutCancel`, keine Bereinigung bei beendetem `ctx` | rot `TestBereinigungKontext` (c) |
| 11 | Fenster *Löschen → Bereinigung* | V12 | rot, deterministisch |
| 12 | Fenster *zwei Anlegen* | V17 | rot, deterministisch |
| 13 | Fenster *zwei Löschen* | V18 | rot, deterministisch |
| 14 | Lesen des Status → Löschen | — | akzeptiertes Negativ mit Grund und Folge (*V* *Ungeprüft*, Plan §3) |
| 15 | Löschen × Exchange | — | akzeptiertes Negativ, Grund: kein Fenster im Dienst |
| 16 | Löschen × Import | V11 `ErrReferenziert` als `ODR-E9001` | rot; Atomarität bei `outboundtest` |
| 17 | Anlegen × Bereinigung | — | `slice-session-wartung` (W15, W24), akzeptiert |
| 18 | Abbruch × Warten auf die Bereinigung | V13 Goroutine ohne Warten | rot (d); das Warten selbst benanntes Negativ |
| 19 | Wiederholung | V7 zweites `Create` überschreibt, V11 zweites `Delete` liefert `nil` | rot |
| 20 | `Startup` nie | `Startup` auf 12 Pfaden (siehe V12) | rot (der Wächter am Ende von `neu` zählt jeden Pfad) |
| 21 | `CleanupBlobs` nie nach Fehler des Löschens | V12 | rot |
| 22 | Ein `Delete` löscht nie eine offene Session | V11 Statusprüfung entfernt, hinter `Sessions.Delete` | rot |
| 23 | Jede Session jedes Services | V10 | rot |
| 24 | Kein Fehlertext, kein Wert der `id` im Text | V15 Fehlertext, V4 `id` im Text | rot |
| 25 | Jeder `*Fehler` mit erlaubtem Status | V16 | rot |
| 26 | Der Dienst hält keinen Zustand | V19, vier Formen | rot unter `-race` |
| 27 | Konfig wächst additiv | — | kein Fänger hier, Träger `slice-session-beenden` |

Die Allaussage *alle übrigen Operationen schreiben nichts* (`SPEC-124` Nr. 9) hat in der Zusatztabelle keine eigene Zeile (F-4).

## Raten der nebenläufigen Fänger

Messender: ich, Docker auf einem Host mit 20 CPUs, Image `golang:1.27.3` (ohne Netz, Testprogramme aus `odata-recorder:race` übersetzt, Quelle `9bdd4bd`). Je Prozess `-test.count=1` und `-test.run '^<Fänger>$'`; rot heißt Rückgabecode ungleich 0. Ungedrosselt 3 Läufe zu je 100 Prozessen (300 je Zelle, 4 und 6 Prozesse gleichzeitig), gedrosselt mit `--cpus=0.05` 1 Lauf zu 20 Prozessen (4 gleichzeitig).

| Fänger · Mutation | unmutiert, ohne / mit Detektor | mutiert ohne Detektor | mutiert mit Detektor |
|---|---|---|---|
| V12 `TestBereinigungNachCommit` · Bereinigung vor `Sessions.Delete` | 0 von 300 / 0 von 300; gedrosselt 0 von 20 / 0 von 20 | 300 von 300; gedrosselt 20 von 20 | 300 von 300; gedrosselt 20 von 20 |
| V13 `TestBereinigungKontext` · Goroutine ohne Warten | 0 von 300 / 0 von 300; gedrosselt 0 von 20 / 0 von 20 | 300 von 300; gedrosselt 20 von 20 | 300 von 300; gedrosselt 20 von 20 |
| V17 `TestAnlegenGleichzeitig` · `Get` vor dem Anlegen | 0 von 300 / 0 von 300; gedrosselt 0 von 20 / 0 von 20 | 300 von 300; gedrosselt 20 von 20 | 300 von 300; gedrosselt 20 von 20 |
| V18 `TestLoeschenGleichzeitig` · der Verlierer ruft die Bereinigung | 0 von 300 / 0 von 300; gedrosselt 0 von 20 / 0 von 20 | 300 von 300; gedrosselt 20 von 20 | 300 von 300; gedrosselt 20 von 20 |
| V19 `TestGleichzeitig` · Feld ohne Sperre | 0 von 300 / 0 von 300; gedrosselt 0 von 20 / 0 von 20 | **0 von 300 rot**; gedrosselt 0 von 20 | 300 von 300; gedrosselt 20 von 20 |

Spanne zwischen den drei ungedrosselten Läufen: 0 (jede Zelle in jedem Lauf 0 oder 100 von 100). Unmutiert laufen alle fünf Fänger in je einem Prozess mit `-test.count=50` grün, ohne und mit Detektor (Rückgabecode 0, 10 Prozesse); die V19-Mutation meldet unter `-test.count=50` in 3 von 3 Prozessen genau ein `WARNING: DATA RACE` je Prozess (`SPEC-122` Nr. 8). Die Sperr-Mutationen von V17 und V18 sind je einmal gelaufen (rot nach dem Auslaufen der Barriere von 5 s), nicht als Rate. **Die Zahlen des Implementers (150 von 150, V19 ohne Detektor 0 von 150, mit Detektor 150 von 150, gedrosselt 60 von 60) bestätigen sich;** eine Streuung habe ich nicht gesehen. V12, V13, V17, V18 sind nicht wahrscheinlichkeitsabhängig; V19 ohne Detektor ist blind (0 von 300 plus 0 von 20), mit Detektor 320 von 320.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Der Dienst gibt den `ctx` des Aufrufs an vier der sechs Port-Aufrufe weiter, ohne dass ein Test es merkt: `Dienste.List`, `Sessions.List`, `Sessions.Get` von `Delete` und `Sessions.Delete` mit `context.Background()` (auch `context.WithoutCancel(ctx)`) bleiben grün; nur `Sessions.Create` und `Sessions.Get` werden rot, und nur, weil `(a)` den beendeten `ctx` am ersten Port der Kette sieht. Das Raster nennt die Zelle *ctx · fremd: unverändert an die Ports*, V13 prüft ihn nur an der Bereinigung. Szenario: ein Umbau gibt `Sessions.Delete` einen eigenen `ctx`, die Frist des Aufrufers gilt nicht mehr, und alle Tests bleiben grün. | `AGENTS.md` §3.15, §3.16; `SPEC-124` Nr. 2 | `loeschen.go` · `if err := v.k.Sessions.Delete(ctx, id); err != nil {`; `verwaltung.go` · `ss, err := v.k.Sessions.List(ctx, d.ID)`; `kontext_test.go` · `(a) ctx schon beendet` | ja — Mutation `R_CtxDeleteDeleteBackground`, `R_CtxSessionsListBackground`, `R_CtxDiensteListBackground`, `R_CtxDeleteGetBackground` grün | Fänger führt nur eine Form der Eingabe (Port-Kette nur am ersten Glied) |
| F-2 | MEDIUM | `Dienste dient allein List` und *sonst kein Aufruf eines Ports* haben für `Dienste.Get`, `Dienste.Create` und `Dienste.SetMetadata` keinen Fänger: die Methoden des Fakes zählen nicht und schreiben kein Ereignis, `gesamt()` und das Ereignisprotokoll sehen sie nie. `Create` mit vorgeschaltetem `Dienste.Get`, `Delete` mit `Dienste.Get`, `List` mit `Dienste.Get` und `SetMetadata` bleiben grün. Szenario: `Create` prüft den Service vorab über `Dienste.Get` (ODR-E5005 vor der Doppelung, das Fenster zwischen Prüfen und Anlegen, das `SPEC-124` Nr. 3 ausschließt), und der Fake meldet „gefunden“. Die Kommentare `Dienste dient allein List` und `Aufrufzähler aller Fakes 0` sagen mehr zu, als ein Test prüft. | `AGENTS.md` §3.15, §3.11; `SPEC-124` Nr. 1, 3 | `helfer_test.go` · `func (d *fakeDienste) Get(context.Context, string) (model.Service, error) {`; `verwaltung.go` · `Dienste dient allein List, Wartung` | ja — Mutation `V5_DiensteGetGerufen`, `R_DeleteDiensteGet`, `R_ListeFehlerBeiDienstenGet` grün | Allaussage über eine Menge ohne Fänger je Element |
| F-3 | MEDIUM | Die Raster-Zelle *Delete · Bereinigung · leer: Bericht aus Nullen unverändert* hat keinen Fänger: die Fälle von `TestBereinigungErgebnis` tragen `Entfernt` 1 bis 4 oder `Zurueckgestellt` allein, der Nullbericht kommt in `TestProtokoll` vor, wird dort aber nicht verglichen. Der Dienst darf einen Nullbericht in `{Entfernt: 1}` oder `{Zurueckgestellt: true}` umschreiben, ohne dass etwas rot wird. Szenario: eine Sonderbehandlung des leeren Berichts („nichts gelöscht“) ändert, was der Aufrufer über die Blobs erfährt. | `AGENTS.md` §3.16; `SPEC-124` Nr. 7 f | `loeschen_test.go` · `{"(a) Entfernt", inbound.Loeschbericht{Entfernt: 3}, nil},` | ja — Mutation `R_BerichtNullenBefuellt`, `R_BerichtNullenZurueckgestellt` grün | Raster-Zelle ohne Fänger (leere Form der Eingabe) |
| F-4 | MEDIUM | Die Allaussage *Eine gelungene Bereinigung und alle übrigen Operationen schreiben nichts* (`SPEC-124` Nr. 9) hat keinen Fänger über alle Pfade: `TestProtokoll (g)` prüft sechs Fehlerfälle von `Delete`, `Create` und `Get`. Ein Eintrag beim Erfolg von `Create` und `Get`, in `List` (Erfolg, beide Fehlerstellen), beim Argumentfehler von `Create` und bei nil-`ctx` bleibt grün (7 von 7 Mutationen). Szenario: eine neue Logzeile in `Create` mit der gelieferten `id` oder dem Fehlertext des Ports erreicht das Protokoll, das `LH-QA-04` auf feste Texte und Codes begrenzt. | `AGENTS.md` §3.15, §3.16 (Allaussage); `LH-QA-04` | `loeschen_test.go` · `// (g) Die Fehler der Operationen selbst erhält der Aufrufer als Code; ins Protokoll gelangt nichts.` | ja — Mutation `V15_EintragBeiErfolgreichemCreate`, `V15_EintragBeiList`, `X_LogBeiGetErfolg`, `X_LogBeiListFehlerDienste`, `X_LogBeiListFehlerSessions`, `X_LogBeiCreateArgumentFehler`, `X_LogBeiCtxNil` grün | Allaussage über eine Menge ohne Fänger je Element |
| F-5 | LOW | Die Gegenprobe von `TestImporte` führt vier Quelltexte (`net/http`, SQLite-Adapter, `os`, `services/maintenance`) und zählt `len(v) != 4`; drei Elemente des Matchers haben keinen Fall: `net` (genau), `crypto/tls`, das Präfix `os/`. Entfernt eine Änderung eines davon aus `verboteneImporte`, bleibt die Gegenprobe grün (das heutige `session` ist unberührt: ein Import von `net`, `crypto/tls`, `os/exec` macht `TestImporte` rot). | `AGENTS.md` §3.15 (ein Fall je Element einer Menge) | `konfig_test.go` · `if v := verboteneImporte(t, gegen); len(v) != 4 {` | ja — Mutation `TEST_ImportTestMutationNetAlsPraefix_gegenprobe`, `TEST_Crypto`, `TEST_OsPraefix` grün | Fänger führt nur eine Form der Eingabe (Gegenprobe) |
| F-6 | LOW | Die Zeile V10 der Tabelle nennt `c` als den Service ohne Session, die Fixture von `TestList` legt den Service ohne Session in die Mitte (`b`) und die älteste Session an `c`; der Absatz in §3 sagt „die Zeile bleibt unverändert“. Die Korrektur der Fixture ist berechtigt (mit `c` ohne Session an letzter Stelle bliebe *ein Service ohne Session beendet die Liste* grün; mit der jetzigen Fixture rot), die Zeile trägt sie nicht nach. | `AGENTS.md` §3.9 | `slice-session-use-case.md` · `die Zeile bleibt unverändert` | nein — Plan-Text, kein Gate | Plan folgt der Korrektur nicht |
| F-7 | LOW | Eine Panik eines Fakes in einem Test ohne `recover` bricht das ganze Testprogramm ab und verdeckt den benannten Fänger: `V2_Create` (Prüfung in `Create` entfernt) wird im ganzen Paket durch den Fall `ctx nil` von `TestOpenInitialSession` rot und bricht ab, bevor `TestCtxNil` (Datei `konfig_test.go`, nach `anlegen_test.go`) läuft; erst `-run TestCtxNil` zeigt den benannten Test. Dasselbe gilt für `Startup` im Fake (Panik). Das Rot kommt aus dem richtigen Grund, aber nicht aus dem benannten Test. | `AGENTS.md` §3.10 (ein Test zählt nur aus dem benannten Grund) | `konfig_test.go` · `Ein Port-Aufruf panict auf dem nil-ctx und macht den Fall rot.` | ja — `V2_Create` ohne `-run` | Rot aus dem benannten Grund in einem anderen Test |
| F-8 | LOW | Der Gegenstand von `Dienste.List` in `List` hat keinen Fänger: `outbound.GegenstandSession` statt `GegenstandSonstiges` bei `Dienste.List` bleibt grün, weil `ErrNichtGefunden` und `ErrBereitsVorhanden`, die den Gegenstand unterscheiden, nur an `Sessions.List` injiziert werden. `SPEC-124` Nr. 8 sagt `Sonstiges` bei `List` ohne Einschränkung auf einen Port. Szenario: `Dienste.List` meldet `ErrNichtGefunden` und der Aufrufer erhält `ODR-E7003` („Recording session unknown“) statt `ODR-E9001`. | `AGENTS.md` §3.15 (ein Fall je Aufrufstelle) | `verwaltung.go` · `dienste, err := v.k.Dienste.List(ctx)` | ja — Mutation `V10_GegenstandSessionBeiDiensteList` grün | Gegenstand der Abbildung ohne Wirkung an einer Stelle |

Mit F-8 sind es 4 MEDIUM, 4 LOW; die Zählung in der Summary folgt dieser Tabelle.

### INFO

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| I-1 | INFO | Marke: gemessen 2107 beim Handoff, Schätzung 1825 (+15,5 %), Marke 2300: der Handoff liegt bei 91,6 %, 193 Zeilen darunter. Die Regel nimmt für den Stand nach dem Review 17 bis 25 % über dem Handoff an (2465 bis 2634); behebt der Implementer F-1 bis F-5, liegt der Stand über der Marke, und die Closure berichtet die Messung mit Grund. Der Plan nennt in §4 noch keinen gemessenen Wert. | `AGENTS.md` §3.14 | `slice-session-use-case.md` · `über **2300** Zeilen nach dem Review` | ja — `git diff --numstat eae82b8..<Stand> -- src/hexagon/ports/inbound src/hexagon/services/session` | Marke gegen Handoff |
| I-2 | INFO | Die Aussage im Absatz in §3, die Mutation *Status ≠ `open`* werde nur über `Validate` im Fake rot, trifft nicht: ohne `Validate` im Fake bleiben `complete`, `incomplete` und `CreatedAt` plus 1 ns rot, weil `TestCreate` Ergebnis und abgelegte Session mit der erwarteten vergleicht. Der Vergleich ist der Wächter, `Validate` ein zweiter. | Maintainability | `slice-session-use-case.md` · `der Wächter ist der Vergleich mit der erwarteten Session` | ja — `V5_StatusComplete_FakeOhneValidate` rot | Plan-Aussage über den Fänger ungenau |
| I-3 | INFO | Reihenfolge `Now()` und `SessionID()` nicht gepinnt: der Vertrag nennt keine Reihenfolge, `TestCreate` sortiert die beiden Ereignisse (`vorCreate`); die Vertauschung ist grün. Ebenso ist der `ctx` des Protokolleintrags nicht zugesagt (`LogAttrs(context.Background()` und `nil` grün). Kein Vertragsbruch. | `SPEC-124` Nr. 3, 9 | `anlegen_test.go` · `slices.Sort(kopf)` | ja — `R_ReihenfolgeNowSessionID`, `R_ProtokollCtxBackground` grün | Reihenfolge ohne Zusage |
| I-4 | INFO | Der Wortlaut der Protokolltexte (`session: Bereinigung nach dem Löschen fehlgeschlagen`, `… zurückgestellt`) steht im Test, nicht im Vertrag; `SPEC-124` Nr. 9 sagt „mit festem Text“ ohne Wortlaut. Die Mutation des Textes wäre rot, eine harmlose Umformulierung auch. Ob der Wortlaut zum Vertrag gehört, entscheidet der Architect. | `AGENTS.md` §3.8, §3.12 | `loeschen_test.go` · `"session: Bereinigung nach dem Löschen fehlgeschlagen", attrs("bereinigung_fehlgeschlagen", "ODR-E3001")` | nein | Wortlaut im Test, nicht im Vertrag |
| I-5 | INFO | Ein `error_code` außer `ODR-E3001` und `ODR-E9001` aus `CleanupBlobs` hat keinen Test, und braucht keinen: `SPEC-123` Nr. 10 schließt ihn aus, der Dienst gibt den Code des Fehlers durch (`V15_CodeImmerE9001` ist rot). Eine Einschränkung auf die beiden Codes wäre im Vertrag gleichwertig. | `SPEC-123` Nr. 10 | `protokoll.go` · `c = f.Code` | ja — `V15_CodeImmerE9001` rot | Zelle außerhalb des Vertrags |
| I-6 | INFO | Das Warten von `Delete` auf die Bereinigung ist ein benanntes Negativ mit Grund und Folge (`SPEC-124` *Ungeprüft*, Plan §3); V13 (d) hält die Synchronität (Goroutine ohne Warten rot in 320 von 320 Prozessen). Die übrigen vier Negative der Zusatztabelle tragen Grund und Folge ebenfalls. | `AGENTS.md` §3.15 | `spec/spezifikation.md` · `Das Warten von `Delete` auf die Bereinigung.` | ja — `V13_GoroutineOhneWarten` rot | — |
| I-7 | INFO | Eine Prüfung über die Runen von `id` mit `byte(r)` bleibt grün: das gültige Mehrbyte-Zeichen der Fixture ist `ä` (Bytes `0xc3 0xa4`), kein Zeichen, dessen Rune beim Abschneiden auf ein Byte in `0-9a-z-` fällt (etwa `š`, `0x161`). Wortlaut der Tabelle V6 ist erfüllt. | `AGENTS.md` §3.15 | `anlegen_test.go` · `"ä": string([]rune{0xe4})` | ja — `V6_RunenStattBytes` grün | Fixture gleicht dem Wert der Mutation nicht |
| I-8 | INFO | Offengelegt zu meinem Lauf: ein Aufruf von `python3 --version` wurde vom Wächter des Repos blockiert (kein Lauf); `sed -i` nur auf eigenen Dateien in der Wegwerf-Ablage; Mutationen und Raten per `docker run` mit Mount der eigenen Wegwerf-Ablage statt über `make` (Auftrag). Das Repo blieb unberührt. | `AGENTS.md` §3.1 | — | nein | — |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/inbound/session.go` | geprüft, ohne Befund (fünf Operationen wie `SPEC-124`, Kommentare Zusage, `Loeschbericht` aus `maintenance.go`) |
| `src/hexagon/services/session/verwaltung.go` | geprüft, Mutation je Zeile, ohne Befund am Produktcode (F-8 betrifft den Fänger) |
| `src/hexagon/services/session/loeschen.go` | geprüft, ohne Befund am Produktcode (Rangfolge, einmal nach dem Commit, Bericht unverändert) |
| `src/hexagon/services/session/protokoll.go`, `fehler.go` | geprüft, ohne Befund am Produktcode (kein Fehlertext, kein Wert der `id`, Kontextfehler ohne Eintrag) |
| `src/hexagon/services/session/*_test.go` | geprüft; Befunde F-1 bis F-8 und I-2 bis I-7 betreffen Fänger, nicht Fehlverhalten des Codes |
| [ADR-0008](../plan/adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md) (Bereinigung unmittelbar nach dem Commit, synchron) | geprüft, ohne Befund; [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md) (Detektor im Gate) ohne Befund (V19 unter `-race`) |
| Sicherheit und Geheimnisse (HIGH-Liste) | geprüft, ohne Befund: die Einträge tragen `service_id`, `session_id`, `error_code` und `ereignis`, nie einen Fehlertext, nie einen Wert aus der Eingabe außer der Kennung der gelöschten Session |
| Spec-Straten | geprüft, ohne Befund (`SPEC-124`, `SPEC-109` Nr. 5, `SPEC-123` Nr. 8 und 10 sagen dasselbe; der Diff ändert keine Spezifikation) |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft, ohne Befund bis auf die zwei Zusagen in F-2 |
| `docs/plan/planning/done/slice-session-use-case.md` §3-Absatz | geprüft, Zahlen und Namen stimmen (20 Fänger-Namen, Raten bestätigt); F-6, I-2 |
| Importe (`.a-check.yml`) | geprüft, ohne Befund (V20) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 4 |
| LOW | 4 |
| INFO | 8 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe (Port-Kette, leerer Bericht, Gegenprobe) · Allaussage über eine Menge ohne Fänger je Element (Dienste-Methoden, Protokoll) · Plan folgt der Korrektur nicht · Rot aus dem benannten Grund in einem anderen Test · Gegenstand der Abbildung ohne Wirkung an einer Stelle

## Verdikt

**Merge-blockierend:** ja — vier MEDIUM (F-1 bis F-4), je eine Mutation, die grün bleibt, und je ein Szenario; kein HIGH, der Produktcode gleicht dem Vertrag in allen zehn Nummern. Die Findings verlangen Fänger (Tests und Fake), keine Änderung am Produktcode; LOW und INFO blockieren nicht.

**Übergabe:** Findings gehen an den Implementer (F-1 bis F-5, F-7, F-8; F-2 und F-4 zugleich an den Planner für die Zeile der Tabelle, F-6 an den Planner); I-4 an den Architect. Die **Finding-Klassen** gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ist ein Lauf-Beleg; er ersetzt keine Verifikation — DoD und Spec-Konformität prüft der Verifier separat. Referenzen: `LH-FA-13`, `LH-RB-04`, [ADR-0008](../plan/adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md).
