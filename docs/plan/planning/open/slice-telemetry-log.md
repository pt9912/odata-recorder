# Slice slice-telemetry-log: Logzeile, Stufen und `slog.Handler` des Adapters von `Telemetry`

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-05`, `LH-QA-04`, `LH-FA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-231` Nr. 1 bis 5 (Zeile, Zeit und Stufe, Text und Zahlen, Abbildung der Dienste, Handler) · `SPEC-232` Nr. 2 (sensible Inhalte, Weg Log), Nr. 3 (Ausfall der Senke), Nr. 5 c, d und f (Sperre und Reihenfolge der Zeile, keine Sperre über `Clock.Now()`) · `SPEC-230` Nr. 1, 3, 4 (Regeln der Attribute, geliefert von `slice-telemetry-port`, hier angewandt) · `SPEC-121` Nr. 12, `SPEC-123` Nr. 11, `SPEC-124` Nr. 9 (Einträge der Dienste `record`, `session`, `maintenance`) · die Logfelder am Ende von spezifikation.md §4 · `ARC-020`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Adapter von `Telemetry` (`src/adapters/driven/telemetry`, `ARC-020`) schreibt, was `SPEC-231` und `SPEC-232` an Logs zusagen: `Log` schreibt je Aufruf eine JSON-Zeile mit den 13 Schlüsseln des Katalogs in fester Reihenfolge, Zeitstempel aus der `Clock`, vier Stufen mit `MinLevel`, `msg` bis 256 Bytes, `duration_ms` mit drei Nachkommastellen; `Handler()` liefert einen `slog.Handler`, mit dem die Dienste `record`, `session` und `maintenance` ohne Änderung ihres Codes im Format des Katalogs loggen (die zehn Einträge der Dienste); fällt die Senke aus, ist die Zeile verloren, `senke` zählt und der Aufrufer merkt nichts; jede Zeile geht in einem `Write` unter einer Sperre hinaus, die Zeilen einer Goroutine stehen in der Reihenfolge ihrer Aufrufe; ein Canary in Header, Body, Query-Literal und Konfigurationswert erreicht keine Zeile und keinen Fehlerwert (`LH-QA-04`). Der Slice liefert den Adapter, nicht den Einbau: kein Dienst und kein Composition Root ruft ihn, bevor die Einbau-Slices laufen.

**Teil des Schnitts vom 2026-10-30 (Planner, vor dem Code):** Teil `L` der Kette `P` → `L` → `S`, den der Architect in `slice-telemetry-port` §4 vorgeschlagen hat. Der Vertrag `SPEC-230` bis `SPEC-232` gehört zu allen drei Slices; dieser liefert `SPEC-231` Nr. 1 bis 5, `SPEC-232` Nr. 2 und 3 für den Weg Log und Nr. 5 c, d und f (Seite von `Log` und `Handle`). `P` liefert Ports, Regeln der Attribute, Zähler, `Snapshot`, `New` und `FakeTelemetry`, `S` die Spans samt Span-Zeile (`SPEC-231` Nr. 6). Die Entscheidungen des Architect (1 bis 9, Alternativen, Gründe, Folgen) stehen in `slice-telemetry-port` §3 und gelten hier unverändert.

**Übernimmt (Sendung von `slice-telemetry-port`, Teil `L`):** Zeile, Stufen, `msg`, `Handler`, Abbildung der zehn Ereignisse der Dienste, Ausfall der Senke, Sperre der Zeile und der Canary über die Wege `Log` und `Handler`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ports, Attribut, Regeln der Schlüssel und Wertformen, Verwurf, `CountError`, `Snapshot`, `New` und `FakeTelemetry` — `slice-telemetry-port` (liegt beim Start dieses Slice in `done/`); dieser Slice baut auf diesem Code auf und ändert seine Regeln nicht.
- `StartSpan`, `Set`, `End`, Eltern-Kind, Verteilung, Ring, Span-Zeile (`SPEC-231` Nr. 6) und der Canary über die Wege der Spans — `slice-telemetry-spans`; die Span-Zeile nutzt den Zeilenschreiber und die Sperre dieses Slice und kommt danach.
- Einbau in Dienste und Adapter des Produkts, Verdrahtung im Composition Root und Umstellung der Dienste `record`, `session` und `maintenance` auf `slog.New(Handler())` (`SPEC-231` Nr. 7) — Empfehlung des Architect: `slice-telemetry-einbau-gateway`; die Annahme steht aus (§6); dieser Slice ändert `src/bootstrap` und die Dienste nicht.
- Schlüssel für Stufe, Format oder Senke — Entscheidung 7 des Architect: Stufe, Format und Senke sind fest; jeder Schlüssel wäre ein Nachzug von `SPEC-110` und ein Handbuch-Kapitel ohne belegten Bedarf.
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

- [ ] Zeile: `Log` schreibt je Aufruf ein JSON-Objekt mit den Feldern des Katalogs in fester Reihenfolge, Zeitstempel aus der `Clock` mit neun Nachkommastellen, vier Stufen mit `MinLevel`, `msg` bis 256 Bytes, `duration_ms` mit drei Nachkommastellen; bei doppeltem Schlüssel gilt der erste (`SPEC-231` Nr. 1 bis 3, `LH-QA-05`).
- [ ] Handler: `Handler()` bildet die Arten von `slog.Attr` ab, trägt `WithAttrs` und `WithGroup`, liefert dieselbe Zeile wie `Log` und gibt die zehn Einträge der Dienste `record`, `session` und `maintenance` mit Text, Stufe und Attributen unverändert aus (`SPEC-231` Nr. 4 und 5, `LH-FA-01`).
- [ ] Senke und Geheimnisse: fällt die Senke aus, ist die Zeile verloren, `senke` zählt und der Aufrufer merkt nichts; jede Zeile geht in einem `Write` unter einer Sperre hinaus, die Zeilen einer Goroutine bleiben in Reihenfolge, keine Sperre liegt über `Clock.Now()`, unter `make test-race` ohne Befund; ein Canary erreicht über keinen der Einbringungswege eine Zeile, den Stand oder einen Fehlerwert (`SPEC-232` Nr. 2, 3, 5 c, d und f, `LH-QA-04`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary, oder macht er Logfelder am Binary sichtbar: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Adapter wird erst vom Einbau verdrahtet; die Logfelder werden dort sichtbar).
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
| `src/adapters/driven/telemetry/` | update | Zeile, Stufen, `msg`, `Handler()`, Senke mit Sperre; `Log` aus `slice-telemetry-port` prüft und zählt bis dahin, schreibt aber keine Zeile (§6); die Senke `FakeSenke` liegt in den Tests des Pakets |
| `spec/spezifikation.md` | unverändert | `SPEC-231` Nr. 1 bis 5 und `SPEC-232` stehen; eine Berichtigung zieht der Implementer im selben Commit nach (`AGENTS.md` §3.9) |
| `spec/architecture.md`, `.a-check.yml`, `go.mod`, `go.sum`, `docs/plan/adr/` | unverändert | wie in `slice-telemetry-port` §3 |
| `src/hexagon/ports/support/`, `src/bootstrap/`, die Dienste | unverändert | der Port steht; der Einbau ist `slice-telemetry-einbau-gateway` |
| Handbuch | unverändert | Binary unverändert (Entscheidung 7 des Architect) |
| Tests | neu | je Zusage ein Fall (Tabelle unten); Canary-Test; nebenläufige Tests laufen unter `make test-race` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation und wurden vom Architect geschrieben (Quelle: `slice-telemetry-port` bis zum Schnitt, Zeilen mit dem Präfix `L`); die Randformen entscheidet der Architect, nicht der Implementer.

**Mengen für die Fänger** (`AGENTS.md` §3.15, §3.17; die Zahl liefert jeweils die Tabelle oder der Wortlaut, eine Zählprüfung im Test vergleicht sie):

- Felder des Logkatalogs: zehn; `timestamp` und `level` setzt der Adapter, acht setzt der Aufrufer, dazu fünf weitere (`ereignis`, `richtung`, `grund`, `blob`, `befunde`): 13 Schlüssel der Menge `Log` (`SPEC-230` Nr. 3).
- Stufen: 4 und 9 Werte des Levels; Ereignisse der Dienste: 10 (`record` 3, `session` 3, `maintenance` 4); Arten von `slog.Attr`: 10 (`slog.Kind`); Operationen der Ports, die der Slice anfasst: `Log`, `CountError`, `Handle`.
- Einbringungswege des Canary: 4 (Header, Body, Query-Literal, Konfigurationswert); Eingabestellen dieses Slice: 13 Schlüssel von `Log`; Ausgabewege dieses Slice: 4 (Zeile über `Log`, Zeile über den `Handler`, `Errors`, `Durations` und `Dropped` des Stands, Fehlerwert von `New`); die Wege Span-Zeile und `Spans` und `Attrs` des Stands und die 11 Span-Schlüssel prüft `slice-telemetry-spans` (6 Wege, 24 Schlüssel).
- Beide Seiten jeder Grenze: `msg` 255, 256, 257 Bytes, `MinLevel` bei Gleichheit, Dauer 0 (zwei gleiche Zeiten), Zeitstempel 0 ns, 1 ns, 999999999 ns.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Namen der Tests sind Vorschläge, der Fänger ist die beschriebene Schleife mit Orakel. Ein akzeptiertes Negativ zählt `andere`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| L01 | Eine Zeile ist ein JSON-Objekt mit `\n` am Ende, in fester Reihenfolge; nicht gesetzte Felder fehlen; jeder Aufruf schreibt eine Zeile (`SPEC-231` Nr. 1) | `TestZeilenform`: Schleife über 15 Teilmengen (leer, jeder der 13 Schlüssel einzeln, alle); Orakel ein Token-Parser, der Reihenfolge und Zahl der Zeilen prüft; derselbe Eintrag zweimal ergibt zwei Zeilen | Reihenfolge von `level` und `timestamp` getauscht; nicht gesetztes Feld als leerer Text | andere |
| L02 | `timestamp` kommt aus der `Clock` in UTC mit neun Nachkommastellen (`SPEC-231` Nr. 2) | `TestTimestamp`: `FestUhr` mit 0 ns, 1 ns, 999999999 ns und einer Zeit in anderer Zone (4 Fälle) | `time.RFC3339Nano` (schneidet Nullen ab) | andere |
| L03 | Die Stufe wird auf vier Namen abgebildet (`SPEC-231` Nr. 2) | `TestStufen`: Schleife über die 9 Werte −8, −4, −1, 0, 3, 4, 7, 8, 12; Orakel die Abbildung der Spezifikation | Text von `slog.Level.String` (`INFO+2`) | andere |
| L04 | Einträge unter `MinLevel` entfallen ohne Zählung, die Gleichheit wird geschrieben (`SPEC-231` Nr. 2) | `TestMinLevel`: Schleife über 4 Stufen × 4 Minima (16 Paare); Orakel „Stufe ab Minimum“; `Dropped` bleibt 0 | `>` statt `>=` | andere |
| L05 | `msg` wird auf 256 Bytes an einer Zeichengrenze gekürzt, leer ist zulässig, ungültiges UTF-8 steht als U+FFFD (`SPEC-231` Nr. 3) | `TestMsg`: 6 Fälle (255, 256, 257 Bytes; 255 Bytes plus ein Zeichen aus zwei Bytes; ungültige Bytefolge aus Bytewerten; leer) | Kürzen mitten im Zeichen | andere |
| L06 | `duration_ms` steht mit genau drei Nachkommastellen (`SPEC-231` Nr. 3) | `TestDauerFormat`: 0, 1 ns, 12.345 ms, 1,0005 ms, 5 s; Orakel `strconv.FormatFloat(f, 'f', 3, 64)` | Ausgabe ohne feste Stellen (`%v`) | andere |
| L07 | Bei doppeltem Schlüssel im Aufruf steht der erste Wert in der Zeile (`SPEC-230` Nr. 4) | `TestDoppelterSchluessel`: Schleife über 13 Schlüssel (Zählprüfung 13), zweiter Wert gültig und verschieden; Zeile zeigt den ersten, `schluessel` 1 | der letzte Wert gewinnt | andere |
| L08 | Die zehn Einträge der Dienste erscheinen mit ihren Attributen, Text und Stufe unverändert (`SPEC-231` Nr. 4) | `TestDienstEintraege`: Schleife über 10 Ereignisse (`record` 3, `session` 3, `maintenance` 4; Zählprüfung 10), die Attribute als Daten aus dem Wortlaut von `SPEC-121` Nr. 12, `SPEC-123` Nr. 11, `SPEC-124` Nr. 9, über `slog.New(Handler())`; Orakel genau diese Schlüssel, `Dropped` überall 0 | `blob` fehlt in der Menge `Log` | andere |
| L09 | Der Handler bildet die Arten ab: `String`, `Int64`, `Uint64` bis `MaxInt64`, `Float64`, `Duration` ja, jede andere nicht, auch kein `error` (`SPEC-231` Nr. 5) | `TestSlogArten`: Schleife über die 10 Werte von `slog.Kind` (Zählprüfung 10) × 2 Schlüssel (`service_id`, `duration_ms`) und `Uint64` über `MaxInt64`; `Any` mit einem `error` und einem `fmt.Stringer`, `Group`, `LogValuer`; Orakel nur zulässige Paare in der Zeile, der Text des `error` fehlt | `Any` wird mit `fmt.Sprint` ausgegeben | andere |
| L10 | `Log` und `Handler` ergeben für dieselben Attribute dieselbe Zeile; `WithAttrs` und `WithGroup` gelten (`SPEC-231` Nr. 5) | `TestHandlerGleichLog`: Schleife über 13 Schlüssel, beide Wege, gleiche Zeile; `WithAttrs` gültig, mit unbekanntem Schlüssel, doppelt zum Aufruf; `WithGroup` nichtleer (Schlüssel entfällt), leer (derselbe Handler) | `WithAttrs` ohne Prüfung | andere |
| L11 | `Handle` liefert immer `nil`, `Enabled` folgt `MinLevel` (`SPEC-231` Nr. 5) | `TestHandleOhneFehler`: Senke in 3 Zuständen (heil, Fehler, Kurzschreiben); `Enabled` über 4 Stufen | `Handle` gibt den Schreibfehler zurück | andere |
| L12 | Der Zeitstempel kommt von der `Clock`, nicht von `Record.Time` (`SPEC-231` Nr. 5) | `TestHandlerZeit`: `Record.Time` verschieden von `Clock.Now()` | `Record.Time` verwendet | andere |
| L13 | `Log` und der Handler schreiben bei nil-`ctx`, beendetem `ctx` und abgelaufener Frist (`SPEC-230` Nr. 1) | `TestLogCtx`: 3 `ctx`-Formen × 2 Wege (6 Fälle); Orakel eine Zeile | `Log` prüft `ctx.Err()` | andere |
| L14 | Fällt die Senke aus, ist die Zeile verloren, `senke` zählt, die nächste wird geschrieben, der Aufrufer merkt nichts (`SPEC-232` Nr. 3) | `TestSenkeAusfall` mit `FakeSenke`: Schleife über 3 Fehlerformen (Fehler ohne Bytes, weniger Bytes als die Zeile, Fehler mit allen Bytes) × 2 Wege (`Log`, `Handler`); Orakel `senke` 1, nächste Zeile da, keine Panik | der Fehler löst eine Wiederholung aus | andere |
| L15 | Keine Operation gerät in Panik (`SPEC-232` Nr. 3) | `TestKeinePanik`: Schleife über `Log`, `Handle`, `CountError` × 5 Argumentarten (nil-`ctx`, nil-Slice, Art außerhalb der drei, leerer Schlüssel, NaN) | `ctx.Value` ohne nil-Prüfung | andere |
| L16 | Ein Canary erreicht keine Ausgabe (`SPEC-232` Nr. 2, `LH-QA-04`) | `TestCanary`: Schleife über 4 Einbringungswege (Header `Bearer …`, Body-Fragment mit Anführungszeichen, Query-Literal `$filter=Name eq '…'`, Konfigurationswert `https://user:…@host`) × 13 Schlüssel von `Log` (Zählprüfung 4 und 13), Orakel: weder das Canary noch sein Präfix von 4 Bytes steht in der Senke oder im Text des Stands (`%+v`); die 11 Span-Schlüssel und die Wege Span-Zeile und `Spans` laufen mit S13 in `slice-telemetry-spans`, dort steht die Menge 4 × 24 über 6 Wege | ein ungültiger Wert wird gekürzt statt verworfen | andere |
| L17 | Akzeptiertes Negativ: ein wohlgeformter Wert in einem zulässigen Schlüssel und der Inhalt von `msg` gehen durch (`SPEC-232` Nr. 2, Ungeprüft) | kein Fänger; Grund: die Form ist prüfbar, der Inhalt nicht; Folge: eine Stelle, die einen Request-Wert in `session_id` oder `msg` legt, gibt ihn aus; Fänger je Stelle im Einbau, Review | — | andere |
| L18 | Eine Zeile geht in einem `Write` unter einer Sperre hinaus (`SPEC-232` Nr. 5 c) | `TestZeilenGleichzeitig`: 32 Goroutinen × 200 Einträge über `Log` und `Handler`; `FakeSenke` hält jede Zeile und den Höchstwert gleichzeitiger `Write` (1); Orakel 6400 Zeilen, jede ein gültiges Objekt | `Write` ohne Sperre | nebenläufig |
| L19 | Die Zeilen einer Goroutine stehen in der Reihenfolge ihrer Aufrufe (`SPEC-232` Nr. 5 d) | `TestZeilenReihenfolge`: 32 Goroutinen, jede schreibt `befunde` 1 bis 200 mit eigener `session_id`; Orakel je `session_id` streng steigend | `Write` in einer eigenen Goroutine | nebenläufig |
| L20 | Keine Sperre des Adapters liegt über `Clock.Now()` (`SPEC-232` Nr. 5 f) | `TestUhrRuftAdapter`: Schleife über 2 Wege (`Log`, `Handle`) × 3 Rückrufe in `Now()` (`CountError`, `Snapshot`, `Log`; 6 Fälle, Zählprüfung 2 und 3); der Aufruf läuft in einer Goroutine mit Frist von 2 s, Orakel „kehrt zurück“, bei `Log` im Rückruf zwei Zeilen; Deadlock in jedem Lauf, kein Haltepunkt nötig | `Log` ruft `Now()` unter der Sperre der Zeile oder des Zustands | nebenläufig |

Zählung dieses Teils: 20 Zeilen, davon 3 `nebenläufig` (*a*, L18, L19, L20) und 17 `andere` (*b*, darunter das akzeptierte Negativ L17).

**Betriebsregeln für alle Rollen:**

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts (Entscheidung des Users vom 2026-10-21).
- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4` (höchstens zwei Container je Linie zugleich) unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Die Image-Falle: das Image `odata-recorder:race` trägt `/src/src` eingebacken; ein Lauf, der Quellen hineinkopiert, leert den Pfad zuvor oder nutzt die Stufe `deps`. Ein Rot unter Last wird auf ruhigem Host wiederholt, bevor es als Befund gilt.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`: eine weitere Verletzung ist ein Befund des Reviewers); keine Hilfsdateien im Wurzelverzeichnis. Commit-Nachrichten gehen als Datei (`git commit -F`), nicht als Heredoc.
- Eine Hülle um ein Argument (ein Fake, der `ctx`, einen Strom oder einen Writer umhüllt, ein Test-Writer der Senke) darf keine Eigenschaft des Originals verdecken (Abbruch, Fehler, Blockieren, Zahl der Aufrufe; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`): `FakeSenke` hält jede Zeile fest, zählt jeden `Write`, hält den Höchstwert gleichzeitiger `Write`, kann Fehler, Kurzschreiben und Blockieren wie ein echter `io.Writer`, und eine Panik des Fakes färbt den benannten Test, nicht einen anderen.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Die Klassen der Tabelle sind generierbar (hier: jedes Feld des Katalogs, jedes Ereignis der Dienste, jede Art von `slog.Kind`): schon der Handoff führt den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben wird nur im Integrationsschritt gesetzt, nicht im Slice-Commit.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13); der Bericht allein ist keine Adresse.
- Der Implementer hält an einer Randform ohne Ort in der Spezifikation an und gibt sie an den Architect zurück (`AGENTS.md` §3.12); er entscheidet sie nicht im Code.
- Wert mit Escape-Folge oder Rohbytes: der Plan und die Berichte schreiben eine solche Folge nicht als Zeichen, sondern beschrieben (Werkzeug-Falle Backslash-u, `AGENTS.md` §3.18).

**Aufträge je Rolle** (Planner, beim Schnitt vom 2026-10-30):

- *Implementer:* L18, L19 und L20 sind nebenläufig; die Zusage „unter einer Sperre“ gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird: der Bericht nennt die Rate je Mutation mit Zahl der Prozesse (50 Prozesse mit `-test.count=1` unter `make test-race`, dazu `-count=50` für die Wirkung) und das Messende, und der Implementer trägt sie mit der Tabelle in einen Absatz von §3 ein (`AGENTS.md` §3.10). Die Canary-Schleife L16 läuft als Schleife über 4 Einbringungswege × 13 Schlüssel (52 Paare) über die 4 Ausgabewege dieses Slice, nicht als Liste einzelner Fälle. `FakeSenke` und jeder Test-Writer halten die Regel der Hüllen (Betriebsregeln). Die Zusatzzeit der Stufen `test` und `race` (vorher und nachher, Zahl der Läufe, Zustand des Hosts) steht in §7; die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) liegt beim User, nicht im Slice.
- *Reviewer:* löst jede Zeile der Tabelle und jede Zelle des Rasters mit einer Mutation auf, fährt L18, L19 und L20 selbst mit Rate, prüft den Canary gegen die Klasse (alle 4 Wege × 13 Schlüssel × 4 Ausgabewege) und die Hülle `FakeSenke` auf verdeckte Eigenschaften; eine im Code entschiedene Randform ist ein MEDIUM (`AGENTS.md` §3.16).
- *Verifier:* prüft die Tabelle gegen `SPEC-231` und `SPEC-232`, dass dieser Slice nichts aus `slice-telemetry-port` oder `slice-telemetry-spans` mitnimmt, dass das Binary unverändert blieb, und die Messpunkt-Zeile (Handoff und Ende, `SPEC-128` Nr. 9).
- *Planner (Closure):* setzt die Messpunkt-Zeile, die Auslastung und meldet dem Architect, liegt sie über 1,00 oder unter 0,60.

**Rollenfolge** (`.claude/agents/`, kein Self-Review): Architect (Vertrag, Tabelle, Raster, Block `groessenmarke`) → Planner (Schnitt, Marke beim Start, §4) → Implementer (Code, Mutationen selbst rot, Absatz zu Zusagen und Raten in §3, ein Handoff-Commit) → Reviewer (Report unter `docs/reviews/`) → Fix-Runden (Menge je Befund, `AGENTS.md` §3.17) → Verifier → Planner (Closure mit Lerneintrag; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorlage an den User) → Integrationsschritt (Merge, Register, Roadmap, Stempel, `make gates` im Hauptbaum).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): `slice-telemetry-port` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-port`: *Code-Abhängigkeit*. Der Slice erweitert den Adapter aus `src/adapters/driven/telemetry`: `New` und `Config`, die Regeln der Attribute (Schlüssel, Wertformen, Verwurf), `Dropped` samt Grund `senke`, `CountError` und `Snapshot`, dazu die Typen in `src/hexagon/ports/support`. Der `Log` des ersten Teils prüft und zählt, schreibt aber keine Zeile; ohne diesen Code gibt es nichts, was die Zeile aufnimmt.
- Handbuch-Punkt der DoD (nur wenn er greift): `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` aus `welle-benutzerhandbuch` liegen dann in `done/`; ohne sichtbare Änderung am Binary entfällt die Bedingung.

**Trigger der Welle:** Der Start-Trigger von [welle-betrieb](../welle-betrieb.md) §2 (`welle-mock` in `done/`, M3 erreicht) ist für diesen Slice wie für `slice-telemetry-port` gelockert (Entscheidung des Planners nach Auftrag des Users vom 2026-10-30, dort §4): der Slice braucht keinen Code von Mock oder Workspace und ändert `src/bootstrap` nicht. Die Lockerung gilt nur, solange der Slice ohne Wirkung am Binary bleibt; sonst wartet er auf die beiden Handbuch-Slices (Rückführung `in-progress` → `open`, unten).

**Größenmarke:** `basis`, `schaetzung` und `marke` setzt der Planner beim Start, vor dem ersten Code-Commit, in eigenem Commit (`SPEC-128`: Schätzung *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke *M* = 100·⌈135·*S*/10000⌉; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Erwartet nach der Tabelle dieses Teils: *a* = 3 (L18, L19, L20), *b* = 17, *f* = 1 (`FakeSenke`), *p* = 380 (Teil des Architect, 1310 für alle drei Teile), also *S* = 85·3 + 80·17 + 200·1 + 380 = 2195 und *M* = 3000; die Schnittfrage (*S* ≥ 2223) stellt sich nicht, liegt aber knapp darunter: wächst die Tabelle um eine Zeile (+80 oder +85), liegt *S* über der Schwelle und der Planner schneidet vor dem Code. Der Planner rechnet beim Start neu, falls die Tabelle sich bis dahin änderte. Pfad: allein der Adapter; der Slice ändert `src/hexagon/ports/support` nicht, und eine Änderung dort wäre eine Abweichung vom Plan (§6), die den Pfad in den Block nimmt.

```groessenmarke
pfad: src/adapters/driven/telemetry
produktcode: 380
fakes: 1
```

`fakes` zählt `FakeSenke`; `FakeTelemetry` liefert `slice-telemetry-port`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (Messung nach `SPEC-128`).
- `in-progress` → `open` (blockiert): der Slice bringt einen Code, Schlüssel oder ein Kommando in den Katalog des Binary oder macht Logfelder am Binary sichtbar, und `slice-handbuch-geruest` oder `slice-handbuch-meldungskatalog` liegen nicht in `done/` (nicht erwartet); oder `Log` aus `slice-telemetry-port` ist nach Randform oder Fänger nicht als Grundlage der Zeile tauglich (§6, Adresse Architect).

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

- Auslegung des Planners (`Log` in P ohne Zeile): bestätigt vom Architect; `MinLevel` wendet erst dieser Slice an (L04) — **Ausgang:** entschieden; Ort: §3 von `slice-telemetry-port` und `SPEC-232` Nr. 1.
- Sperre über `Clock.Now()` (`SPEC-232` Nr. 5 f), Seite von `Log` und `Handle` — **Ausgang:** entschieden; Fänger L20.
- Der Span-Teil nutzt Zeilenschreiber und Sperre dieses Slice für die Span-Zeile (`SPEC-231` Nr. 6); ein Zuschnitt des Schreibers, der die Zeile eines Spans nicht aufnimmt, träfe `slice-telemetry-spans` — **Ausgang:** weiter offen; Adresse: Implementer dieses Slice (der Schreiber nimmt eine Zeile aus Feldern entgegen, nicht nur die Attribute von `Log`; der Reviewer prüft es), Folge-Slice `slice-telemetry-spans`.
- Akzeptiertes Negativ L17 (wohlgeformter Wert in zulässigem Schlüssel, Inhalt von `msg`): Folge ist eine Ausgabe, wo eine Stelle einen Request-Wert dort ablegt; der Schutz liegt im Fänger je Stelle des Einbaus — **Ausgang:** weiter offen; Adresse: Planner (Fänger je Stelle in die DoD von `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock`, `AGENTS.md` §3.13; in `open/` schreibt dieser Zweig nichts).
- Verdrahtung des Adapters und Umstellung der Dienste `record`, `session`, `maintenance` auf `slog.New(Handler())` (`SPEC-231` Nr. 7): Empfehlung des Architect `slice-telemetry-einbau-gateway`, dessen Ziel die Verdrahtung nennt, `session` und `maintenance` aber nicht; die Annahme steht aus — **Ausgang:** weiter offen; Adresse: Planner (Datei von `slice-telemetry-einbau-gateway` ergänzen, bevor dieser Slice sie nennt, `AGENTS.md` §3.13).
- `SPEC-121` Nr. 12: der Satz „einen Port `Telemetry` gibt es noch nicht“ wird mit dem Einbau veraltet; der Wortlaut gehört dem Einbau-Slice — **Ausgang:** weiter offen; Adresse: Planner (`slice-telemetry-einbau-gateway`).
- Handbuch: die Logfelder werden am Binary sichtbar, sobald der Einbau den Adapter verdrahtet; die Handbuch-Pflicht greift dort, nicht hier — **Ausgang:** weiter offen; Adresse: Planner (`slice-telemetry-einbau-gateway`, Abhängigkeiten `slice-handbuch-geruest`, `slice-handbuch-meldungskatalog`).
- Eine Senke, die nie zurückkehrt, hält den Aufrufer (`SPEC-232` Ungeprüft, Entscheidung 6 des Architect: synchrones Schreiben) — **Ausgang:** weiter offen; Adresse: User (Folge im Betrieb, wenn der Log-Treiber des Containers blockiert; kein Fänger, solange kein Betrieb diesen Fall zeigt).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten ist gefüllt, jede Zelle trägt ihren Ort — **Ausgang:** entfallen; eine Zelle ohne Ort, die der Implementer findet, geht an den Architect und steht dann hier.
- Die Größenmarke ist die Schätzung nach `SPEC-128`; `AGENTS.md` §3.14 trägt noch die ältere Formel (Faktoren 85 und 40, Zuschlag 1,25) — **Ausgang:** weiter offen; Adresse: User (Diff-Vorlage im Closure von `slice-telemetry-port`, §6 dort).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Dieser Teil trägt die Zeilen `Log` und `Handler`; `CountError` und `Snapshot` stehen in `slice-telemetry-port`, die Zeilen der Spans in `slice-telemetry-spans`. Die Nummern stehen in spezifikation.md §5.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | sensibler Wert | Abbruch (`ctx`) | Wiederholung |
|---|---|---|---|---|---|---|
| Log schreiben (`Log`) | `SPEC-231` Nr. 1 (nicht gesetztes Feld fehlt; Pflicht je Ereignis liegt bei der Stelle) | `SPEC-230` Nr. 3 (leerer Wert ist Formverstoß, `wert`); `SPEC-231` Nr. 3 (leerer `msg` zulässig) | `SPEC-230` Nr. 2 und 4 (Schlüssel außerhalb: `schluessel`); `SPEC-231` Nr. 2 (Level außerhalb der vier Namen) | `SPEC-232` Nr. 2 | `SPEC-230` Nr. 1 (wirkt auch bei beendetem `ctx`) | `SPEC-231` Nr. 1 (jeder Aufruf eine Zeile); `SPEC-230` Nr. 4 (doppelter Schlüssel im Aufruf) |
| Log über den `Handler` | `SPEC-231` Nr. 1 | `SPEC-230` Nr. 3, `SPEC-231` Nr. 5 | `SPEC-231` Nr. 5 (Arten und Gruppen: `wert`, Präfix: `schluessel`) | `SPEC-231` Nr. 5 (kein `error`-Text), `SPEC-232` Nr. 2 | `SPEC-231` Nr. 5 (`Handle` liefert `nil`), `SPEC-230` Nr. 1 | `SPEC-231` Nr. 1 und 5 (`WithAttrs`, doppelter Schlüssel) |
| Senke schreibt (`Write` des Adapters) | gilt nicht: der Adapter reicht die Zeile, kein Argument fehlt | gilt nicht: eine Zeile ist nie leer, sie trägt mindestens `timestamp` und `level` (`SPEC-231` Nr. 1) | `SPEC-232` Nr. 3 (Fehler oder weniger Bytes als die Zeile: verloren, `senke`) | `SPEC-232` Nr. 2 (nie ein Fehlertext der Senke) | gilt nicht: kein `ctx` an der Senke (`SPEC-230` Nr. 1) | `SPEC-232` Nr. 3 (kein Wiederholen, die nächste Zeile wird geschrieben); `SPEC-232` Nr. 5 c (ein `Write` je Zeile, nie zwei zugleich) |

**Zusatztabelle: Verstöße am selben Schlüssel** (`AGENTS.md` §3.16; `SPEC-230` Nr. 4, Rangfolge Schlüssel vor Wert, genau ein Verwurf je Attribut): die Rangfolge und die Zählung fängt `TestVerwurfRangfolge` (P17) in `slice-telemetry-port`; dieser Slice fängt den Fall „doppelter Schlüssel, beide gültig und verschieden“ in der Zeile für alle 13 Schlüssel (L07) und über den Handler mit `WithAttrs` (L10).

**Zusatztabelle: Modus** (`AGENTS.md` §3.16; die Zusage hängt am Modus):

| Modus | Zusage | Fänger |
|---|---|---|
| `MinLevel` über `INFO` | keine Zeile unter dem Minimum, kein Verwurf | L04 |
| Senke defekt | keine Zeile, `senke` 1, die nächste wird geschrieben (`SPEC-232` Nr. 3) | L14 |
| Weg `Log` oder `Handler` | dieselbe Zeile (`SPEC-231` Nr. 5) | L10 |

Die Modi „`MinLevel` bis `INFO`“ und „Senke defekt“ für die Span-Zeile fängt S13 in `slice-telemetry-spans`.

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

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: die Messpunkt-Zeile von `make groessenmarke` (Handoff und Ende, von Hand gemessener Produktcode, Auslastung), die Raten der Wirkungs-Fänger L18, L19 und L20 mit Zahl der Prozesse und Messendem, die Zusatzzeit der Stufen `test` und `race`, und dass das Binary unverändert blieb.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, Zähler aus den Dateien unter `evidence/`; der Planner zieht Treffer und Zähler beim Start nach). Einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für die Pfade, die kein Lauf ausführt (Ausfall der Senke), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für `FakeSenke`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) für L18 und L19, `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (7×, `AGENTS.md` §3.10) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `AGENTS.md` §3.16) für die Allaussage „kein Credential in keinem Log“, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `implement-slice` Schritt 14) für die Mengen der Fänger, dazu `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (offen, Schwelle im Hauptbaum erreicht; dieser Zweig setzt keinen Zähler).

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF.
