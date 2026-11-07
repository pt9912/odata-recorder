# Slice slice-telemetry-handler: `slog.Handler` des Adapters von `Telemetry`, Arten, Gruppen und Abbildung der Dienste

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-FA-01`, `LH-QA-05`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-231` Nr. 4 (Abbildung der Dienste), Nr. 5 (Handler) und die Nr. 1 bis 3 für den Weg `Handle` (dieselbe Zeile wie `Log`) · `SPEC-232` Nr. 2 (sensible Inhalte, Weg Handler), Nr. 3 (`Handle` liefert `nil`, Ausfall der Senke), Nr. 5 c, d, f und g (Sperre und Reihenfolge, Uhr, Sperre der Zeile, Seite von `Handle` und der abgeleiteten Handler) · `SPEC-230` Nr. 1, 3, 4 (Regeln der Attribute, geliefert von `slice-telemetry-port`, hier über den Handler angewandt) · `SPEC-121` Nr. 12, `SPEC-123` Nr. 11, `SPEC-124` Nr. 9 (Einträge der Dienste `record`, `session`, `maintenance`) · die Logfelder am Ende von spezifikation.md §4 · `ARC-020`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-telemetry-handler`, Worktree `/Development/odata-recorder-wt-telemetry-handler`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-11-07.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `Handler()` des Adapters von `Telemetry` (`src/adapters/driven/telemetry`, `ARC-020`) liefert einen `slog.Handler`, mit dem die Dienste `record`, `session` und `maintenance` ohne Änderung ihres Codes im Format des Katalogs loggen: `Enabled` folgt `MinLevel`, `Handle` bildet die Arten von `slog.Attr` ab (zuerst der Schlüssel, dann die Art; eine Gruppe ist ein Attribut), schreibt über den Zeilenschreiber von `slice-telemetry-log` dieselbe Zeile wie `Log`, wendet `MinLevel` selbst an, ignoriert `Record.Time` und liefert immer `nil`; `WithAttrs` und `WithGroup` leiten Handler ab, die Sperre und Stand teilen; die zehn Einträge der Dienste erscheinen mit Text, Stufe und Attributen unverändert; ein Canary in Header, Body, Query-Literal und Konfigurationswert erreicht über den Handler keine Zeile und keinen Fehlerwert (`LH-QA-04`). Der Slice liefert den Handler, nicht den Einbau: kein Dienst und kein Composition Root ruft ihn, bevor `slice-telemetry-einbau-gateway` läuft.

**Teil des Schnitts vom 2026-11-07 (Planner, vor dem Code):** Teil `H` des Plans `slice-telemetry-log`, den der Architect in dessen §1 vorgeschlagen und der Planner vollzogen hat (Schätzung des ungeschnittenen Plans 3095, über der Schnittgrenze 2223, `SPEC-128` Nr. 1). Teil `L` (`slice-telemetry-log`: Zeile, Stufen, Senke, Sperre, Canary über `Log`; Zeilen `L01` bis `L20`) liegt vor diesem Slice, die Kette heißt jetzt `P` → `L` → `H` und `P` → `L` → `S`; `H` und `S` hängen beide an `L` und nicht aneinander. Der Vertrag `SPEC-230` bis `SPEC-232` gehört zu allen Slices der Kette; dieser liefert `SPEC-231` Nr. 4 und 5 und die Seite von `Handle` in `SPEC-232` Nr. 2, 3 und 5. Die Entscheidungen 1 bis 9 des Architect stehen in `slice-telemetry-port` §3, die Entscheidungen 10 bis 13 und 15 bis 17 in `slice-telemetry-log` §3 und gelten hier unverändert; die Entscheidung 14 (Handler) steht in §3 dieses Plans.

**Übernimmt (Sendung von `slice-telemetry-port` über `slice-telemetry-log`, Teil `H`; Geber: `slice-telemetry-log` §1, Schnitt des Planners vom 2026-11-07):** `Handler`, die Abbildung der zehn Ereignisse der Dienste, die Arten von `slog.Attr`, `WithAttrs`, `WithGroup`, die Sperre der Zeile und die Uhr auf der Seite von `Handle` und der Canary über den Weg `Handler`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Zeile, Stufen, `MinLevel`, `msg`, Zahlen, Zeilenschreiber, Senke samt Ausfall, Sperre der Zeile, `FakeSenke`, Canary über `Log` und der Ersatz von `TestKernSchreibtNichts` (Sendung F-10) — `slice-telemetry-log` (liegt beim Start dieses Slice in `done/`); dieser Slice baut auf diesem Code auf, ändert die Form der Zeile nicht und legt keinen zweiten Schreiber an. Eine Änderung dort ist eine Abweichung vom Plan (§6).
- Ports, Attribut, Regeln der Schlüssel und Wertformen, Verwurf, `CountError`, `Snapshot`, `New` und `FakeTelemetry` — `slice-telemetry-port`.
- `StartSpan`, `Set`, `End`, Verteilung, Ring und Span-Zeile — `slice-telemetry-spans`; die Span-Zeile nutzt den Zeilenschreiber von `slice-telemetry-log`, nicht den Handler.
- Einbau in Dienste und Adapter des Produkts, Verdrahtung im Composition Root und Umstellung der Dienste `record`, `session` und `maintenance` auf `slog.New(Handler())` (`SPEC-231` Nr. 7), der Wortlaut von `SPEC-121` Nr. 12 zum Port — `slice-telemetry-einbau-gateway` (nimmt die Sendung in seinem §1 an, Eintrag des Planners vom 2026-11-07); dieser Slice ändert `src/bootstrap` und die Dienste nicht.
- Schlüssel für Stufe, Format oder Senke — Entscheidung 7 des Architect: Stufe, Format und Senke sind fest.
- Endpunkt, Netz-Port und Export an einen Collector — kein Lastenheft-Bedarf (`slice-telemetry-port` §1).
- Audit-Einträge — `slice-audit-log`; ein Audit-Eintrag ist kein Log-Eintrag.

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

- [ ] Handler: `Handler()` bildet die Arten von `slog.Attr` ab (zuerst der Schlüssel, dann die Art; eine Gruppe ist ein Attribut), trägt `WithAttrs` und `WithGroup`, wendet `MinLevel` in `Enabled` und `Handle` an, ignoriert `Record.Time`, liefert dieselbe Zeile wie `Log` und gibt die zehn Einträge der Dienste `record`, `session` und `maintenance` mit Text, Stufe und Attributen unverändert aus (`SPEC-231` Nr. 4 und 5, `LH-FA-01`, `LH-QA-05`).
- [ ] Senke, Sperre und Geheimnisse über den Handler: `Handle` liefert immer `nil` und gerät nie in Panik, auch bei defekter Senke; Handler und alle abgeleiteten teilen die Sperre der Zeile und den Stand, die Zeilen einer Goroutine bleiben in Reihenfolge, keine Sperre liegt über `Clock.Now()`, unter `make test-race` ohne Befund; ein Canary erreicht über keinen der Einbringungswege eine Zeile, den Stand oder einen Fehlerwert (`SPEC-231` Nr. 5, `SPEC-232` Nr. 2, 3, 5 c, d, f und g, `LH-QA-04`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary, oder macht er Logfelder am Binary sichtbar: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Handler wird erst vom Einbau verdrahtet; die Logfelder werden dort sichtbar).
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
| `src/adapters/driven/telemetry/` | update | `Handler()` mit `Enabled`, `Handle`, `WithAttrs`, `WithGroup` über den Zeilenschreiber, die Sperre und die Senke aus `slice-telemetry-log`; kein zweiter Schreiber |
| `src/adapters/driven/telemetry/*_test.go` | neu | je Zusage ein Fall (Tabelle unten); Canary-Test über den Handler; nebenläufige Tests laufen unter `make test-race`; `FakeSenke`, `FestUhr` und `RueckrufUhr` stammen aus `slice-telemetry-log` bzw. `supporttest`, dieser Slice legt keinen Fake an |
| `spec/spezifikation.md` | unverändert (Architect, erledigt) | `SPEC-231` Nr. 4 und 5 und `SPEC-232` Nr. 3, 5 d, 5 g sind vor dem Code entschieden (Historie 2026-11-07); eine Berichtigung zieht der Implementer im selben Commit nach (`AGENTS.md` §3.9) |
| `spec/architecture.md`, `.a-check.yml`, `go.mod`, `go.sum`, `docs/plan/adr/` | unverändert | wie in `slice-telemetry-port` §3 |
| `src/hexagon/ports/support/`, `src/bootstrap/`, die Dienste | unverändert | der Port steht; der Einbau ist `slice-telemetry-einbau-gateway` |
| Handbuch | unverändert | Binary unverändert (Entscheidung 7 des Architect) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation und wurden vom Architect geschrieben (Zeilen `H01` bis `H11` aus `slice-telemetry-log` vom 2026-11-07, vom Planner beim Schnitt hierher getragen); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidung des Architect zum Handler** (2026-11-07, vor dem Code; Fortsetzung der Entscheidungen 1 bis 9 aus `slice-telemetry-port` §3 und 10 bis 13, 15 bis 17 aus `slice-telemetry-log` §3; der Wortlaut steht in der Spezifikation, `AGENTS.md` §3.8):

14. **Handler** (`SPEC-231` Nr. 5): zuerst der Schlüssel, dann die Art; eine Gruppe ist ein Attribut und wird nicht aufgelöst; `WithAttrs` wird bei jedem Eintrag wie dessen erste Attribute geprüft und gezählt; `WithGroup` setzt das Präfix auch vor spätere `WithAttrs`; abgeleitete Handler teilen Sperre und Stand. *Alternative:* die `slog`-Konvention (Gruppen auflösen, leere Gruppe und Nullwert ignorieren; verworfen: ein aufgelöstes Gruppenelement erreichte die Menge unter zusammengesetztem Schlüssel, und die Regeln der Attribute hätten zwei Zähler für dieselbe Stelle). *Folge:* wer `slog.Group` übergibt, bekommt einen Verwurf `schluessel` oder `wert`, nie eine Zeile mit dem Inhalt.

**Mengen für die Fänger** (`AGENTS.md` §3.15, §3.17; Achsen, Elemente als Produkt, Generator, Orakel und Prüfgrenze; die Zählprüfung vergleicht die Zahl der Elemente mit der **Quelle** der Menge, nie mit einer festen Zahl im Test). Das Orakel ist nie der Adapter: die Zeile liest ein eigener Leser im Test (`json.Decoder`, für die Folge der Schlüssel `Token`; für Rohbytes `bytes`), und ein eigenes Modell der Zusage liefert den erwarteten Wert. **Mengen aus `slice-telemetry-log` §3**, die dieser Slice über den Weg `Handle` wiederholt, mit Generator, Orakel und Zählprüfung von dort: M1 (Felder der Zeile: H04, H05), M2 (Länge von `msg`: H04), M3 (Zeichen von `msg`: H04), M4 (Stufe × `MinLevel`: H04, H07), M5 (Zeitstempel: H07), M7 (doppelter Schlüssel: H05), M8 (Argumente von `Log`, hier die 3 `ctx`: H08), M9 (Senke, 7 Formen der Rückgabe: H08). Die Mengen M6 und M10 gehören allein `slice-telemetry-log`. Die Nummern M11 bis M15 sind hier unverändert geführt, damit die Querverweise des Architect bestehen bleiben.

| Menge | Achsen | Elemente (Produkt) | Generator und Quelle der Zählprüfung | Orakel | Prüfgrenze (akzeptiertes Negativ: Grund, Folge) |
|---|---|---|---|---|---|
| M11 Arten von `slog.Attr` (H02, H03) | Stelle, Zeichen | 10 Arten × 3 Schlüssel (`service_id` String, `duration_ms` Float, `befunde` Int) = 30 + 18 Grenzfälle (Uint64 `MaxInt64`, `MaxInt64`+1, `MaxUint64`; Int64 −1, 0, `MaxInt64`; Duration 0, 1 ns, −1 ns; Float64 NaN, +Inf, −0, 0, 1.5; Any `error`, `fmt.Stringer`, nil, String) + 3 `LogValuer` (löst zu String auf, zu Bool, Panik) = 51 | `for k := slog.KindAny; k <= slog.KindLogValuer; k++`; Zählprüfung: die Zahl der Arten gleich `slog.KindLogValuer`+1 und `(slog.KindLogValuer+1).String()` beginnt mit `Kind(` (eine neue Art der Standardbibliothek färbt den Test) | Modelltabelle Art × Schlüssel → angenommen oder verworfen; der Text eines `error` fehlt in der Zeile | `Any` mit weiteren Typen; Grund: jeder `Any` wird verworfen; Folge: keine |
| M12 Rangfolge (H03) | Stelle, Art | 4 Schlüssel (gültig, fremd, leer, mit Gruppen-Präfix) × 10 Arten = 40 | Schleifen | genau ein Verwurf je Attribut: fremd, leer, Präfix → `schluessel`, gültig mit falscher Art → `wert`; eine Gruppe mit drei Elementen zählt 1 | — |
| M13 Einträge der Dienste (H01) | Stelle | 10 Ereignisse: `record` 3, `session` 3, `maintenance` 4 | Datentabelle aus dem Wortlaut von `SPEC-121` Nr. 12, `SPEC-124` Nr. 9, `SPEC-123` Nr. 11; Zählprüfung: Summe der drei Teillisten gleich der Zahl der Zeilen von `SPEC-231` Nr. 4 (3 + 3 + 4) | Zeile mit genau den Schlüsseln der Stelle, `msg` und `level` gleich dem Aufruf, `Dropped` überall 0 | die Werte der Attribute im Betrieb; Folge: L15 in `slice-telemetry-log` |
| M14 `WithAttrs` und `WithGroup` (H05, H06) | Folge, Stelle | `WithAttrs`: 13 Schlüssel × 3 Fälle (gültig, ungültiger Wert, doppelt zum Aufruf) + 1 fremder Schlüssel = 40; `WithGroup`: 4 Folgen von Namen (leer; `g`; `a`, dann `b`; leer, dann `g`) × 2 Lagen von `WithAttrs` (vor, nach der Gruppe) + der Fall „leerer Name gibt denselben Handler“ = 9 | Schleifen | Zeile gleich `Log(vorgelegte + aufgerufene Attribute)`; Zählung je Eintrag; Empfänger unverändert | tiefere Verschachtelung als 2; Folge: gering, das Präfix ist eine Verkettung |
| M15 Canary über den Handler (H09) | Stelle, Ausgabe | 13 Werte über den Aufruf + 13 Werte über `WithAttrs` + 5 (Text eines `error`, Text eines `Stringer`, Name einer Gruppe, Name von `WithGroup`, fremder Schlüssel) + 1 Fehlertext der Senke = 32 Stellen × 4 Wege (Header `Bearer …`, Body-Fragment mit Anführungszeichen, Query-Literal `$filter=Name eq '…'`, Konfigurationswert `https://user:…@host`) × 2 Ausgaben (Bytes der Senke, `%+v` des Stands) = 256 | Schleifen; Zählprüfung 13 gegen M1, 4 gegen die Liste der Wege | weder das Canary noch sein Präfix von 4 Bytes in der Ausgabe; das Präfix steht in der Zeile eines Aufrufs mit gültigen Werten nicht (Kontrolle) | der Inhalt von `msg`; der Fehlerwert von `New` und die Wege der Spans; weitere Canary-Formen; Folge: siehe L15 in `slice-telemetry-log` |
| M16 Rückrufe der Uhr, Seite `Handle` (H11) | Seite, Stelle | `Handle`: Rückrufe `CountError`, `Snapshot`, `Log`, `Handle` = 4 | Liste von Funktionen in `Now()` der `RueckrufUhr` (ruft den Rückruf beim ersten `Now()`) | der Aufruf kehrt innerhalb 2 s zurück; bei Rückruf `Log`/`Handle` stehen zwei Zeilen in der Senke | — |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Namen der Tests sind Vorschläge, der Fänger ist die beschriebene Schleife mit Orakel. Ein akzeptiertes Negativ zählt `andere`. Die Nummern `H01` bis `H11` bleiben die des Architect. `FakeSenke` (Hülle eines `io.Writer`, geliefert von `slice-telemetry-log`) hält jedes Argument von `Write` als Kopie, zählt jeden `Write`, hält den Höchstwert gleichzeitiger `Write`, kann je Aufruf eine Rückgabe (n, err) liefern und blockieren; sie verdeckt weder Blockieren noch Fehler noch die Zahl der Aufrufe.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| H01 | Die zehn Einträge der Dienste erscheinen mit Text, Stufe und Attributen unverändert (`SPEC-231` Nr. 4) | `TestDienstEintraege`: Menge M13 (10), über `slog.New(Handler())` mit `Error`, `Warn`, `Info`; dazu der Eintrag `capture` bytegleich der Beispielzeile in Nr. 4 bis auf `timestamp` | `Handle` schreibt `INFO` statt der Stufe; ein Schlüssel der Stelle fehlt in der Menge | andere |
| H02 | Der Handler bildet die Arten ab: `String`, `Int64`, `Uint64` bis `MaxInt64`, `Float64`, `Duration` ja, jede andere nicht, auch kein `error` (`SPEC-231` Nr. 5) | `TestSlogArten`: Menge M11 (51); Orakel die Modelltabelle, der Text des `error` fehlt in der Zeile | `Any` wird mit `fmt.Sprint` ausgegeben; `Uint64` ohne Obergrenze; `Duration` in Nanosekunden | andere |
| H03 | Zuerst der Schlüssel, dann die Art; eine Gruppe ist ein Attribut; ein leerer Schlüssel hat keine Sonderregel (`SPEC-231` Nr. 5) | `TestHandlerRangfolge`: Menge M12 (40); Orakel genau ein Verwurf je Attribut | Art zuerst geprüft; Gruppe aufgelöst; leerer Schlüssel ignoriert | andere |
| H04 | `Log` und `Handler` ergeben für dieselben Attribute, Texte und Stufen dieselbe Zeile (`SPEC-231` Nr. 5) | `TestHandlerGleichLog`: Mengen M1 (16384 Aufrufe), M2 (469) und M3 (8909) über beide Wege, dazu die 9 Stufen aus M4; Orakel die Zeilen beider Wege bytegleich | der Handler schreibt die Felder in der Folge des Aufrufs; `msg` nicht gekürzt; Stufe über `Record.Level.String()` | andere |
| H05 | `WithAttrs` legt Attribute vor, die bei jedem Eintrag geprüft und gezählt werden; Empfänger unverändert; Kopie (`SPEC-231` Nr. 5) | `TestWithAttrs`: Menge M14 (40); drei Einträge hintereinander mit ungültigem vorgelegtem Attribut: `wert` 3; `WithAttrs` selbst zählt 0; Elternhandler mit Kapazität über der Länge und zwei abgeleitete Geschwister beeinflussen sich nicht; der Aufrufer überschreibt den Slice nach `WithAttrs`, die Zeile bleibt; doppelter Schlüssel zwischen vorgelegtem und aufgerufenem Attribut (M7), der erste gilt | `WithAttrs` ohne Kopie (gemeinsames Array der Geschwister); `WithAttrs` zählt beim Aufruf; Zählung nur beim ersten Eintrag; der Empfänger wird geändert | andere |
| H06 | `WithGroup` mit nichtleerem Namen setzt das Präfix `name.` vor jeden folgenden Schlüssel, auch vor spätere `WithAttrs`; mit leerem Namen derselbe Handler (`SPEC-231` Nr. 5) | `TestWithGroup`: Menge M14 (9); Orakel `schluessel` für Attribute nach der Gruppe, gültig davor, Identität bei leerem Namen | `WithGroup` ignoriert; Präfix auch vor früheren `WithAttrs`; leerer Name erzeugt das Präfix `.` | andere |
| H07 | `Enabled` folgt `MinLevel`; `Handle` wendet `MinLevel` selbst an; der Zeitstempel kommt von der `Clock`, nicht von `Record.Time` (`SPEC-231` Nr. 5) | `TestHandlerMinLevelZeit`: Menge M4 (81 Paare) für `Enabled` und für `Handle` (direkter Aufruf: keine Zeile, `Now()` nie, Verwürfe zählen); `Record.Time` verschieden von `Clock.Now()`, ein Fall mit `time.Time{}`; Menge M5 (8) | `Enabled` immer wahr; `Handle` ohne Prüfung; `Record.Time` verwendet | andere |
| H08 | `Handle` liefert immer `nil` und gerät nie in Panik (`SPEC-231` Nr. 5, `SPEC-232` Nr. 3) | `TestHandleOhneFehler`: Menge M9 (7 Formen der Senke) × Menge M8 (3 `ctx`) = 21 Aufrufe von `Handle`; Orakel `nil`, `senke` gemäß Modell | `Handle` gibt den Schreibfehler zurück; `Handle` prüft `ctx.Err()` | andere |
| H09 | Ein Canary erreicht über den Handler keine Zeile und keinen Stand (`SPEC-232` Nr. 2, `LH-QA-04`) | `TestCanaryHandler`: Menge M15 (256 Prüfungen) | der Text eines `error` steht in der Zeile; der Name einer Gruppe steht in der Zeile; `Stringer` wird aufgerufen und ausgegeben | andere |
| H10 | Handler und alle abgeleiteten teilen die Sperre der Zeile; die Zeilen einer Goroutine bleiben in Reihenfolge (`SPEC-232` Nr. 5 c, d und g) | `TestHandlerSperre`: (1) Haltepunkt: A schreibt über den Wurzelhandler und blockiert im `Write`, B über `WithAttrs`, C über `WithGroup`: kein zweites `Write` innerhalb 200 ms; (2) Masse: 32 Goroutinen × 200 Einträge verteilt auf die vier Handler (Wurzel, `WithAttrs`, `WithGroup`, `slog.New`), Höchstwert 1, je Goroutine streng steigend; (3) `Handle` kehrt nicht zurück, solange `Write` blockiert | der abgeleitete Handler kopiert die Sperre (`sync.Mutex` als Wert); eigener Stand je abgeleitetem Handler; `Write` in eigener Goroutine über `Handle` | nebenläufig |
| H11 | Keine Sperre liegt über `Clock.Now()`, Seite `Handle` (`SPEC-232` Nr. 5 f) | `TestUhrRuftHandle`: `RueckrufUhr` ruft beim ersten `Now()` `CountError`, `Snapshot`, `Log` oder `Handle` eines abgeleiteten Handlers (M16, 4 Fälle); Orakel „kehrt zurück“ innerhalb 2 s, bei `Log` und `Handle` zwei Zeilen | `Handle` ruft `Now()` unter der Sperre der Zeile oder des Zustands | nebenläufig |

Zählung dieser Tabelle: 11 Zeilen, davon 2 `nebenläufig` (*a*: H10, H11) und 9 `andere` (*b*).

**Zweige als Menge** (`AGENTS.md` §3.15; Register `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`): die Modell-Menge der Wege von `Handle` steht in §6 (Zweig-Menge, Z38 bis Z65). Die Zählprüfung ist die Abdeckung unter dem Detektor: der Implementer fährt `go test -race -coverprofile` über das Paket in der Stufe `race` (unter `flock`, mit `--memory`) und nennt im Bericht jeden Block der neuen Dateien ohne Treffer; null, oder je Block ein benanntes Negativ mit Grund und Folge. Eine zweite Zählung im Test selbst entfällt: ein Test kann die Zweige des Codes nicht zählen, ohne den Code zu lesen, und die Abdeckung der Blöcke tut es ohne neues Werkzeug.

**Betriebsregeln für alle Rollen:**

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird im Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt der Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt der Zweig nichts (Entscheidung des Users vom 2026-10-21).
- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4` (höchstens zwei Container je Linie zugleich) unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Die Image-Falle: das Image `odata-recorder:race` trägt `/src/src` eingebacken; ein Lauf, der Quellen hineinkopiert, leert den Pfad zuvor oder nutzt die Stufe `deps`. Ein Rot unter Last wird auf ruhigem Host wiederholt, bevor es als Befund gilt.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); keine Hilfsdateien im Wurzelverzeichnis. Commit-Nachrichten gehen als Datei (`git commit -F`), nicht als Heredoc.
- Eine Hülle um ein Argument (ein Fake, der `ctx`, einen Strom oder einen Writer umhüllt, ein Test-Writer der Senke) darf keine Eigenschaft des Originals verdecken (Abbruch, Fehler, Blockieren, Zahl der Aufrufe; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`): die Tests dieses Slice nutzen `FakeSenke` aus `slice-telemetry-log` und fügen keine eigene Hülle hinzu, die Abbruch, Fehler oder die Zahl der Aufrufe verdeckt; eine Panik eines Fakes färbt den benannten Test.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Die Klassen der Tabelle sind generierbar (hier: jedes Feld des Katalogs, jedes Ereignis der Dienste, jede Art von `slog.Kind`): schon der Handoff führt den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben wird nur im Integrationsschritt gesetzt, nicht im Slice-Commit.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13); der Bericht allein ist keine Adresse.
- Der Implementer hält an einer Randform ohne Ort in der Spezifikation an und gibt sie an den Architect zurück (`AGENTS.md` §3.12); er entscheidet sie nicht im Code.
- Wert mit Escape-Folge oder Rohbytes: der Plan und die Berichte schreiben eine solche Folge nicht als Zeichen, sondern beschrieben (Werkzeug-Falle Backslash-u, `AGENTS.md` §3.18).

**Aufträge je Rolle** (Planner, beim Schnitt vom 2026-11-07):

- *Implementer:* H10 und H11 sind nebenläufig; die Zusage „unter einer Sperre“ gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird: der Bericht nennt die Rate je Mutation mit Zahl der Prozesse (50 Prozesse mit `-test.count=1` unter `make test-race`, dazu `-count=50` für die Wirkung) und das Messende, und der Implementer trägt sie mit der Tabelle in einen Absatz von §3 ein (`AGENTS.md` §3.10). Die Haltepunkt-Fänger (H10 (1) und (3), H11) sind deterministisch in Richtung der Mutation; die Masse-Fänger laufen daneben. Die Canary-Schleife H09 läuft als Schleife über 32 Stellen × 4 Wege × 2 Ausgaben (256 Prüfungen), nicht als Liste einzelner Fälle. Die Mengen M1 bis M5, M7 bis M9 wiederholt der Slice aus `slice-telemetry-log` mit denselben Generatoren: der Implementer nutzt deren Hilfen (Leser, Modelle, Listen) und legt keine zweite Fassung an. Die Zusatzzeit der Stufen `test` und `race` (vorher und nachher, Zahl der Läufe, Zustand des Hosts; M1 und M3 über den Handler sind die teuersten Läufe) steht in §7; die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) liegt beim User, nicht im Slice.
- *Reviewer:* löst jede Zeile der Tabelle und jede Zelle des Rasters mit einer Mutation auf, fährt die nebenläufigen Zeilen selbst mit Rate, prüft den Canary gegen die Klasse (alle Stellen × 4 Wege × alle Ausgaben), löst die Zweig-Menge (§6) gegen den Bericht der Abdeckung auf; eine im Code entschiedene Randform ist ein MEDIUM (`AGENTS.md` §3.16).
- *Verifier:* prüft die Tabelle gegen `SPEC-231` Nr. 4 und 5 und `SPEC-232`, dass dieser Slice nichts aus `slice-telemetry-port`, `slice-telemetry-log` oder `slice-telemetry-spans` ändert, dass das Binary unverändert blieb, und die Messpunkt-Zeile (Handoff und Ende, `SPEC-128` Nr. 9).
- *Planner (Closure):* setzt die Messpunkt-Zeile, die Auslastung und meldet dem Architect, liegt sie über 1,00 oder unter 0,60.

**Aufträge an den Architect zum Start** (Start, Planner, 2026-11-07, `main` = `69269ec`; vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Architect arbeitet im Zweig `slice-telemetry-handler`; Ausgang jeweils *offen* in §6). Tabelle, Raster und Mengen sind am 2026-11-07 geschrieben; die Aufträge prüfen sie mit den Regeln, die `slice-telemetry-log` gelernt hat. Wo etwas vollständig ist, genügt die Bestätigung mit Nummer; eine neue Zeile rechnet der Planner in die Schätzung ein.

- **A. Zuschnitt des Zeilenschreibers (Prüfauftrag, §6).** `zeile.go` liefert `eintrag(level, msg, attrs []support.Attr)` (prüft über `filtern(opLog, …)`, zählt, wendet `MinLevel` an, ruft die Uhr erst für eine geschriebene Zeile) und `zeileSchreiben(t, level, msg, felder)` (Zeile bilden, ein `Write` unter der Sperre der Zeile, `senke` zählen). Der Handler soll **ohne zweite Fassung** darüber schreiben. Der Architect liest den Code auf `main` und entscheidet schriftlich, in welcher Form `Handle` die Zeile übergibt: Abbildung `slog.Attr` → `support.Attr` (Schlüssel vor Art, Gruppe als Attribut, Präfix von `WithGroup`, vorgelegte Attribute von `WithAttrs` bei jedem Eintrag) vor dem Aufruf von `eintrag`, oder ein schmaler Zugang am Schreiber. Er nennt, ob `filtern` die Verwürfe `schluessel` und `wert` für die Fälle des Handlers (Präfix, Gruppe, Art außerhalb) selbst liefert oder ob der Handler eine eigene Prüfung bräuchte (das wäre die zweite Fassung der Regeln der Attribute). Nimmt der Schreiber das nicht auf, ist es eine Änderung an `slice-telemetry-log` und eine Rückführung nach §4; der Architect schreibt dann den Zuschnitt in §3 und in die Spezifikation, nicht der Implementer.
- **B. Sätze einer Vertragsnummer (Lerneintrag `slice-telemetry-log`).** Jeder Satz von `SPEC-231` Nr. 4 und 5 und `SPEC-232` Nr. 2, 3 und 5 c, d, f und g, auch ein Nebensatz neben der Hauptzusage (zum Beispiel „Empfänger unverändert“, „Kopie“, „`Handle` liefert auch bei beendetem `ctx` `nil`“, „zählt `WithAttrs` selbst 0“), hat eine Zeile der Tabelle oder ein benanntes Negativ mit Grund und Folge. Der Architect geht die Nummern Satz für Satz durch und trägt fehlende Zeilen als `H12` ff. ein.
- **C. Mengen und Zählprüfung.** M11 bis M16 und die aus `slice-telemetry-log` wiederholten Mengen haben Achsen, Generator, Orakel und Prüfgrenze; die Zählprüfung vergleicht gegen die **Quelle** (`slog.Kind`, `model.AlleCodes()`, die Liste der Felder des Katalogs, die Liste der Ereignisse der Dienste), nie gegen eine feste Zahl. Der Architect prüft M13 (10 Einträge) gegen den Wortlaut der drei Stellen `SPEC-121` Nr. 12, `SPEC-123` Nr. 11 und `SPEC-124` Nr. 9 und M15 (13 Werte) gegen M1.
- **D. Zweige in Runden unter dem Detektor** (Register `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, 9×). Die Zweig-Menge Z38 bis Z65 hat einen Fänger je Zweig; für jeden Zweig, den eine Goroutine nur einmal erreicht (Panik von `LogValue`, `Handle` unter `MinLevel`, Senke), nennt der Architect die Runde (ein Lauf je Zweig unter `go test -race -coverprofile`), und die Abdeckung der neuen Blöcke ist die Zählprüfung (null Blöcke ohne Treffer oder ein Negativ mit Grund und Folge).
- **E. Gegenstand der Fehlerabbildung je Aufrufstelle.** Was der Vertrag je Aufruf zusagt (`ctx`, Verwurf, Kontextfehler), hat einen Fall je Aufrufstelle: `Handle`, `Enabled`, `WithAttrs`, `WithGroup` und die abgeleiteten Handler. Der Architect bestätigt M8 (3 `ctx`) für `Handle` und `Enabled` und nennt, ob `Enabled` das `ctx` liest.
- **F. Fakes zählen jede Operation.** `FakeSenke` hält jedes Argument von `Write` und zählt jeden `Write`; der Architect nennt, ob der Handler einen Fake neben `FakeSenke` und `RueckrufUhr` braucht (dann steht er in der Tabelle `f`), und dass die Tests keine Hülle um Senke oder `ctx` legen, die Abbruch, Fehler oder Zahl der Aufrufe verdeckt.
- **G. Haltepunkt-Fänger gegen Masse.** H10 (1) und (3) und H11 sind Haltepunkt-Fänger; (2) ist Masse. Der Architect nennt je nebenläufiger Zusage, welche Mutation der Haltepunkt in jedem Lauf rot macht und welche nur die Masse mit Rate fängt, und die Negative der Tabelle *Fenster der Sperre* (§6).
- **H. Schnittfrage und Marke.** Der Planner hat vor dem Code *S* = 1010 und *M* = 1400 gerechnet (§4, Prosa); der Architect bestätigt *a*, *b*, *f* und *p* oder nennt neue Werte und legt keinen Block mit `basis`, `schaetzung` und `marke` an. Den Block mit den drei Zeilen setzt der Planner nach der Bestätigung in eigenem Commit vor dem ersten Code-Commit (`AGENTS.md` §3.14). Schnitt-Grund nur, wenn *S* ≥ 2223 würde.
- **I. Kennungen und Katalog-Pflicht.** Der Slice vergibt keine neue Kennung: `SPEC-230` bis `SPEC-232` genügen; `SPEC-236` gehört `slice-inspection-use-case` und wird nicht berührt, `SPEC-195` bis `SPEC-199` gehören `slice-mock-lesen`. Der Architect bestätigt, dass kein Code, Schlüssel oder Kommando in den Katalog des Binary kommt (`SPEC-201` Nr. 3 und 11); die Tests ändern weder `telemetry_test.go` noch `regeln_test.go` noch `audit_test.go` (dort zieht `slice-mock-lesen` die Zahl der Codes nach); die Hilfen (Zählprüfung gegen `model.AlleCodes()`) werden genutzt, nicht geändert.
- **J. Reports der Rollen.** Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (Register `BEO-REPO/report-der-rolle-faellt-docs-check`, 5×); der Architect nimmt den Satz in seinen Bericht auf, der Planner reicht ihn an Reviewer und Verifier weiter.

**Rollenfolge** (`.claude/agents/`, kein Self-Review): Architect (Vertrag, Tabelle, Raster, Block `groessenmarke`; erledigt) → Planner (Marke beim Start, §4) → Implementer (Code, Mutationen selbst rot, Absatz zu Zusagen und Raten in §3, ein Handoff-Commit) → Reviewer (Report unter `docs/reviews/`) → Fix-Runden (Menge je Befund, `AGENTS.md` §3.17) → Verifier → Planner (Closure mit Lerneintrag; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorlage an den User) → Integrationsschritt (Merge, Register, Roadmap, Stempel, `make gates` im Hauptbaum).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): `slice-telemetry-log` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Beim Start prüft der Planner die Übergangsregeln (Linie D, [`README.md`](../README.md) §Parallelbetrieb) und den Stand der Zweige gegen den Pfad des Slice.

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-log`: *Code-Abhängigkeit*. `Handle` schreibt über den Zeilenschreiber, die Stufen mit `MinLevel`, die Senke mit Ausfall und die Sperre der Zeile aus diesem Slice; die Tests bauen auf `FakeSenke`, die Hilfen der Mengen M1 bis M9 und `RueckrufUhr` auf. Ohne ihn gibt es nichts, was der Handler aufnimmt; die Kette `L` → `H` ist die Reihenfolge des Codes, nicht eine Wahl.
- `slice-telemetry-port`: *Code-Abhängigkeit* über `slice-telemetry-log` (Regeln der Attribute, `Dropped`, `CountError`, `Snapshot`, `New`); liegt in `done/`.
- Handbuch-Punkt der DoD (nur wenn er greift): `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` liegen in `done/`; die Bedingung greift nur bei sichtbarer Änderung am Binary (nicht erwartet).

**Unabhängig von** `slice-telemetry-spans` (beide hängen an `slice-telemetry-log`, nicht aneinander; sie laufen in beliebiger Folge, schreiben aber beide in `src/adapters/driven/telemetry`, und nach dem ersten Merge zieht der zweite Zweig main nach).

**Trigger der Welle:** Der Start-Trigger von [welle-betrieb](../welle-betrieb.md) §2 (`welle-mock` in `done/`, M3 erreicht) ist für diesen Slice wie für `slice-telemetry-port` und `slice-telemetry-log` gelockert (Ausnahme je Slice, Planner, 2026-11-07, Linie D): der Slice braucht keinen Code von Mock oder Workspace und ändert `src/bootstrap` nicht. Die Lockerung gilt nur, solange der Slice ohne Wirkung am Binary bleibt; sonst wartet er auf die beiden Handbuch-Slices (Rückführung `in-progress` → `open`, unten).

**Größenmarke** (Planner, 2026-11-07; Prosa-Form, der Block mit `basis`, `schaetzung` und `marke` folgt beim Start, vor dem ersten Code-Commit, in eigenem Commit): Nach `SPEC-128` Nr. 1 (Schätzung *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke *M* = 100·⌈135·*S*/10000⌉; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen) und der Tabelle in §3: *a* = 2 (H10, H11), *b* = 9, *f* = 0 (`FakeSenke` liefert `slice-telemetry-log`), *p* = 120 (Teil des Architect, Pfad allein `src/adapters/driven/telemetry`, der Slice ändert `src/hexagon/ports/support` nicht und eine Änderung dort wäre eine Abweichung vom Plan, §6), also *S* = 170 + 720 + 0 + 120 = 1010 und *M* = 1400; die Schnittfrage (*S* ≥ 2223) stellt sich nicht. Der Planner rechnet beim Start neu, falls die Tabelle oder die Schnittfläche zu `slice-telemetry-log` sich bis dahin änderte; `basis` ist dann der Commit, der Vertrag, Tabelle und Raster abschließt.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (Messung nach `SPEC-128`).
- `in-progress` → `open` (blockiert): der Slice bringt einen Code, Schlüssel oder ein Kommando in den Katalog des Binary oder macht Logfelder am Binary sichtbar, und `slice-handbuch-geruest` oder `slice-handbuch-meldungskatalog` liegen nicht in `done/` (nicht erwartet); oder der Zeilenschreiber aus `slice-telemetry-log` nimmt die Zeile des Handlers nicht auf (§6, Adresse Architect).

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

- Sperre über `Clock.Now()` (`SPEC-232` Nr. 5 f), Seite von `Handle` — **Ausgang:** entschieden; Fänger H11 (die Seite von `Log` fängt L19 und L20 in `slice-telemetry-log`).
- Zuschnitt des Zeilenschreibers: der Handler braucht einen Schreiber, der eine Zeile aus Feldern und Attributen aufnimmt, mit der Sperre der Zeile und der Zählung von `senke`, ohne eigene Kopie der Regeln der Attribute; nimmt der Schreiber aus `slice-telemetry-log` das nicht auf, ist das eine Abweichung — **Ausgang:** weiter offen; Adresse: Architect (der Implementer hält an und gibt zurück, `AGENTS.md` §3.12).
- Aufträge A bis J an den Architect zum Start (§3): Zuschnitt des Zeilenschreibers, Sätze je Vertragsnummer, Zählprüfung gegen die Quellen, Runden für die Zweige, gezählte Aufrufstellen, Fakes, Haltepunkt gegen Masse, Marke, Kennungen, Reports — **Ausgang:** offen; Adresse: Architect (Bestätigung in diesem Abschnitt oder Ergänzung), Implementer (Test), Reviewer und Verifier (Auftrag J).
- Block `groessenmarke` (Auftrag H in §3, §4): der Block steht auf `main` nicht; die Prosa-Marke (*S* 1010, *M* 1400) gilt, bis der Planner den Block setzt — **Ausgang:** offen; Adresse: Planner (Block mit `basis`, `schaetzung` und `marke` nach der Bestätigung des Architect, vor dem ersten Code-Commit).
- Pfade neben den laufenden Zweigen (Planner, Start 2026-11-07): `slice-mock-lesen` schreibt `telemetry_test.go`, `regeln_test.go` und `audit_test.go` (Zahl der Codes), `slice-inspection-use-case` schreibt `ports/inbound/inspection.go` und `services/inspection/`; dieser Slice schreibt nur neue Dateien in `src/adapters/driven/telemetry` — **Ausgang:** offen; Adresse: Integrationsschritt (der zweite Merge zieht `main` nach; berührt dieser Slice eine der drei Testdateien, ist das eine Abweichung an den Planner).
- Verdrahtung des Adapters und Umstellung der Dienste `record`, `session`, `maintenance` auf `slog.New(Handler())` (`SPEC-231` Nr. 7): angenommen von `slice-telemetry-einbau-gateway` (Eintrag des Planners vom 2026-11-07 in dessen §1) — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway`.
- `SPEC-121` Nr. 12: der Satz „einen Port `Telemetry` gibt es noch nicht“ wird mit dem Einbau veraltet; der Wortlaut gehört dem Einbau-Slice — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway`.
- Handbuch: die Logfelder werden am Binary sichtbar, sobald der Einbau den Adapter verdrahtet; die Handbuch-Pflicht greift dort, nicht hier — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway` (Abhängigkeiten `slice-handbuch-geruest`, `slice-handbuch-meldungskatalog`).
- Akzeptiertes Negativ L15 aus `slice-telemetry-log` (wohlgeformter Wert in zulässigem Schlüssel, Inhalt von `msg`) gilt auch für den Handler: die Dienste legen ihre Werte über `slog.New(Handler())` ab; der Schutz liegt im Fänger je Stelle des Einbaus — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` (Fänger je Stelle in der DoD, Eintrag des Planners vom 2026-11-07).
- Akzeptiertes Negativ L16 aus `slice-telemetry-log` (Panik von `Writer` oder `Clock`, nie zurückkehrende Senke, Senke ruft den Adapter, Jahr außerhalb 0000 bis 9999) trifft `Handle` ebenso; `Handle` liefert `nil`, wartet aber mit einer Senke, die nie zurückkehrt — **Ausgang:** weiter offen; Adresse: User (siehe Roadmap-Hinweis dort).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten ist gefüllt, jede Zelle trägt ihren Ort; die Lücken des Handlers (Rangfolge Schlüssel vor Art, Gruppe als Attribut, `WithAttrs` je Eintrag, Sperre geteilt) sind am 2026-11-07 in `SPEC-231` Nr. 5 und `SPEC-232` Nr. 5 entschieden — **Ausgang:** entschieden; Ort: `SPEC-231` Nr. 5, `SPEC-232` Nr. 3, 5 c, d, f, g; eine Zelle ohne Ort, die der Implementer findet, geht an den Architect und steht dann hier.
- Die Größenmarke ist die Schätzung nach `SPEC-128`; `AGENTS.md` §3.14 trägt noch die ältere Formel (Faktoren 85 und 40, Zuschlag 1,25) — **Ausgang:** weiter offen; Adresse: User (Diff-Vorlage im Closure von `slice-telemetry-port`, §6 dort).
- Zählprüfungen: die Hilfen der Mengen M1 bis M9 liegen in den Tests von `slice-telemetry-log`; ändert dieser Slice sie (statt sie zu nutzen), ist das eine Abweichung — **Ausgang:** weiter offen; Adresse: Architect.
- Gate-Kette, `src/bootstrap/` und Linie B (Stand beim Start): ein Bedarf an einer Änderung dort — **Ausgang:** weiter offen (nicht zu erwarten); Adresse: Planner.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Dieser Teil trägt die Zeile `Handler` und die Seite `Handle` der Uhr; die Zeilen `Log`, `Senke` und die Seite `Log` der Uhr stehen in `slice-telemetry-log`, `CountError` und `Snapshot` in `slice-telemetry-port`, die der Spans in `slice-telemetry-spans`. Die Nummern stehen in spezifikation.md §5.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | sensibler Wert | Abbruch (`ctx`) | Wiederholung |
|---|---|---|---|---|---|---|
| Log über den `Handler` | `SPEC-231` Nr. 1 | `SPEC-230` Nr. 3, `SPEC-231` Nr. 5 (leerer Schlüssel ohne Sonderregel; `WithGroup` leer: derselbe Handler) | `SPEC-231` Nr. 5 (Arten, Gruppen und Rangfolge Schlüssel vor Art; Präfix: `schluessel`) | `SPEC-231` Nr. 5 (kein `error`-Text), `SPEC-232` Nr. 2 | `SPEC-231` Nr. 5 (`Handle` liefert `nil`, auch bei beendetem `ctx`), `SPEC-230` Nr. 1 | `SPEC-231` Nr. 1 und 5 (`WithAttrs` bei jedem Eintrag geprüft, doppelter Schlüssel, abgeleiteter Handler) |
| Zeit der Uhr (`Clock.Now()` in `Handle`) | gilt nicht: die Uhr liefert immer eine Zeit; die Nullzeit ist eine Zeit (`SPEC-231` Nr. 2); `Record.Time` wird nicht gelesen (`SPEC-231` Nr. 5) | gilt nicht: wie fehlender Wert | `SPEC-231` Nr. 2 (andere Zone: UTC; Jahr außerhalb 0000 bis 9999: nicht zugesagt) | gilt nicht: die Zeit trägt keinen Inhalt eines Requests | gilt nicht: kein `ctx` an der Uhr; eine Uhr, die nie zurückkehrt: `SPEC-232` *Ungeprüft* | `SPEC-231` Nr. 2 (einmal je geschriebener Zeile, nie für einen Eintrag unter `MinLevel`); `SPEC-232` Nr. 5 f (Rückruf in `Now()` blockiert nicht) |

**Fenster der Sperre** (`AGENTS.md` §3.10, Zusage „unter einer Sperre, ohne sie dazwischen freizugeben“, `SPEC-232` Nr. 5 c, d, f, g; je Fenster Schritte, Seite und Fänger oder benanntes Negativ; die übrigen Fenster der Zeile, die nicht am Handler hängen, stehen in `slice-telemetry-log` §6):

| Fenster | Seite | Fänger oder Negativ |
|---|---|---|
| Zeile bilden und `Now()` rufen, vor der Sperre der Zeile | `Handle` | H11 (Rückruf `CountError`, `Snapshot`: ein `Now()` unter der Sperre des Zustands blockiert; Rückruf `Log`, `Handle`: ein `Now()` unter der Sperre der Zeile blockiert) |
| zwischen Nehmen der Sperre der Zeile und `Write` | `Handle`, abgeleitete Handler | Negativ ohne Fenster: dazwischen steht kein Schritt, den ein Test erreicht; die Wirkung („zwei `Write` zugleich“) fängt der Haltepunkt am `Writer` in H10, in jedem Lauf |
| `Write` selbst (ein Aufruf mit der ganzen Zeile) | `Handle` | H10 (nie zwei zugleich; kehrt erst nach `Write` zurück); der Fänger „ein `Write` je Zeile“ ist L12 in `slice-telemetry-log` (derselbe Schreiber) |
| zwischen gleichzeitigen Aufrufen (Reihenfolge der Wartenden) | alle Handler | Negativ; Grund: zugesagt ist die Reihenfolge je Goroutine, nicht zwischen Goroutinen (`SPEC-232` Nr. 5 d); Folge: keine |
| Ableiten eines Handlers (`WithAttrs`, `WithGroup`) | abgeleitete Handler | H10 (Sperre geteilt: Haltepunkt über Wurzel, `WithAttrs`, `WithGroup`), H05 (Stand geteilt: die Zählung je Eintrag steht im `Snapshot` des Adapters) |

**Zweig-Menge** (Modell-Menge der Wege von `Handle`; `AGENTS.md` §3.15; Register `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`; die Zählprüfung ist die Abdeckung unter dem Detektor, §3). Ein Zweig ohne Treffer in `go test -race -coverprofile` ist ein benanntes Negativ mit Grund und Folge im Bericht. Die Nummern Z38 bis Z65 sind die des Architect; Z01 bis Z37 (Zeile und Senke) stehen in `slice-telemetry-log`.

| Zweig | Fänger |
|---|---|
| Z38 `Enabled` wahr, Z39 `Enabled` falsch | H07 |
| Z40 `Handle` unter `MinLevel` (direkt aufgerufen) | H07 |
| Z41 bis Z50 je eine Art von `slog.Kind` (`Any`, `Bool`, `Duration`, `Float64`, `Int64`, `String`, `Time`, `Uint64`, `Group`, `LogValuer`) | H02 |
| Z51 `Uint64` bis `MaxInt64`; Z52 `Uint64` darüber | H02 |
| Z53 `LogValuer` löst auf; Z54 `LogValue` gerät in Panik (`Resolve` fängt) | H02 |
| Z55 Schlüssel außerhalb vor der Art (`schluessel`), Z56 Schlüssel gültig, Art falsch (`wert`) | H03 |
| Z57 `WithAttrs` leer; Z58 `WithAttrs` nichtleer; Z59 vorgelegtes Attribut ungültig, je Eintrag gezählt | H05 |
| Z60 `WithGroup` leer (derselbe Handler); Z61 `WithGroup` nichtleer; Z62 verschachtelt; Z63 `WithAttrs` nach `WithGroup` | H06 |
| Z64 `Record.Time` ignoriert | H07 |
| Z65 `Handle` liefert `nil` bei jeder Form der Senke | H08 |

**Zusatztabelle: Verstöße am selben Schlüssel** (`AGENTS.md` §3.16; `SPEC-230` Nr. 4, Rangfolge Schlüssel vor Wert, genau ein Verwurf je Attribut): die Rangfolge und die Zählung fängt `TestVerwurfRangfolge` (P17) in `slice-telemetry-port`; der Fall „doppelter Schlüssel, beide gültig und verschieden“ in der Zeile fängt L08 in `slice-telemetry-log`, über den Handler mit `WithAttrs` fängt ihn H05 (Menge M7); die Rangfolge Schlüssel vor Art im Handler fängt H03 (Menge M12).

**Zusatztabelle: Modus** (`AGENTS.md` §3.16; die Zusage hängt am Modus):

| Modus | Zusage | Fänger |
|---|---|---|
| `MinLevel` über `INFO` (jeder Wert von −8 bis 12) | keine Zeile unter dem Minimum, `Now()` nicht gerufen, Verwürfe zählen | H07 (Seite `Log`: L04 in `slice-telemetry-log`) |
| Senke defekt (sechs Formen) | `Handle` liefert `nil`, `senke` gemäß Modell, die nächste Zeile wird geschrieben (`SPEC-232` Nr. 3) | H08 (Seite `Log`: L10 in `slice-telemetry-log`) |
| Weg `Log` oder `Handler` | dieselbe Zeile (`SPEC-231` Nr. 5) | H04 |
| Handler abgeleitet (`WithAttrs`, `WithGroup`) | dieselbe Sperre, derselbe Stand | H10, H05 |

Der Modus „Senke blockiert“ (Zähler und `Snapshot` warten nicht, `SPEC-232` Nr. 5 g) hängt am Weg der Zeile und steht mit L11 in `slice-telemetry-log`; H10 (3) fängt ihn für `Handle`.

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

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: die Messpunkt-Zeile von `make groessenmarke` (Handoff und Ende, von Hand gemessener Produktcode, Auslastung), die Raten der Wirkungs-Fänger H10 und H11 mit Zahl der Prozesse und Messendem, die Zusatzzeit der Stufen `test` und `race`, den Bericht der Abdeckung (Zweig-Menge Z38 bis Z65) und dass das Binary unverändert blieb.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet und beim Start nachgezogen (Planner, 2026-11-07, `main` = `69269ec`; Zähler = Zahl der Dateien in `evidence/`), Stand und Treffer wie in `slice-telemetry-log` §8. Zähler beim Start: `zweig-ohne-lauf-unter-dem-detektor` 9×, `rate-ohne-detektor-schwankt-mit-der-last` 8×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `betriebsregel-fuer-repo-dateien-verletzt` 9×, `schaetzformel-ueberschaetzt-den-slice` 18×, `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3× (Zählprüfung der Mengen gegen die Quelle), `report-der-rolle-faellt-docs-check` 5× (Auftrag J), `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 43× (Auftrag B). Offen und einschlägig für diesen Teil: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure der Welle oder des nächsten Slice mit Nebenläufigkeit — dieser Slice hat zwei nebenläufige Zeilen, H10 und H11) gilt für die Zweig-Menge Z38 bis Z65; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` für die Betriebsregel „nur `Edit` und `Write`“ (dieser Slice setzt keinen Zähler); `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` für die Marke. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` für den Gebrauch von `FakeSenke`, `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (`AGENTS.md` §3.10) für H10 und H11, `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (`AGENTS.md` §3.16) für die Allaussage „kein Credential in keinem Log“, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` für die Mengen der Fänger, `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (`AGENTS.md` §3.17), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (beide `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (`AGENTS.md` §3.16). Dieser Schnitt legt kein neues Verzeichnis an und setzt keinen Zähler.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF.
