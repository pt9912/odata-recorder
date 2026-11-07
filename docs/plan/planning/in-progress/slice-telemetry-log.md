# Slice slice-telemetry-log: Logzeile, Stufen, Senke und Sperre des Adapters von `Telemetry`

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

**Berührte Spec-Stellen:** `SPEC-231` Nr. 1 bis 3 (Zeile, Zeit und Stufe, Text und Zahlen) · `SPEC-232` Nr. 2 (sensible Inhalte, Weg Log), Nr. 3 (Ausfall der Senke), Nr. 5 c, d, f und g (Sperre und Reihenfolge der Zeile, keine Sperre über `Clock.Now()`, Sperre allein um `Write`; Seite von `Log`) · `SPEC-230` Nr. 1, 3, 4 (Regeln der Attribute, geliefert von `slice-telemetry-port`, hier angewandt) · die Logfelder am Ende von spezifikation.md §4 · `ARC-020`. Der Handler (`SPEC-231` Nr. 4 und 5) und die Einträge der Dienste (`SPEC-121` Nr. 12, `SPEC-123` Nr. 11, `SPEC-124` Nr. 9) gehören `slice-telemetry-handler`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-telemetry-log)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Adapter von `Telemetry` (`src/adapters/driven/telemetry`, `ARC-020`) schreibt, was `SPEC-231` Nr. 1 bis 3 und `SPEC-232` an Logs zusagen: `Log` schreibt je Aufruf eine JSON-Zeile mit den 13 Schlüsseln des Katalogs in fester Reihenfolge, Zeitstempel aus der `Clock`, vier Stufen mit `MinLevel`, `msg` bis 256 Bytes, `duration_ms` mit drei Nachkommastellen; fällt die Senke aus, ist die Zeile verloren, `senke` zählt und der Aufrufer merkt nichts; jede Zeile geht in einem `Write` unter einer Sperre hinaus, die Zeilen einer Goroutine stehen in der Reihenfolge ihrer Aufrufe; ein Canary in Header, Body, Query-Literal und Konfigurationswert erreicht über `Log` keine Zeile und keinen Fehlerwert (`LH-QA-04`). Der Slice liefert den Zeilenschreiber, den `slice-telemetry-handler` (Handler) und `slice-telemetry-spans` (Span-Zeile) benutzen. Er liefert den Adapter, nicht den Einbau: kein Dienst und kein Composition Root ruft ihn, bevor die Einbau-Slices laufen.

**Teil des Schnitts vom 2026-10-30 (Planner, vor dem Code):** Teil `L` der Kette `P` → `L` → `S`, den der Architect in `slice-telemetry-port` §4 vorgeschlagen hat. Der Vertrag `SPEC-230` bis `SPEC-232` gehört zu allen Slices der Kette; dieser liefert `SPEC-231` Nr. 1 bis 3, `SPEC-232` Nr. 2 und 3 für den Weg Log und Nr. 5 c, d, f und g (Seite von `Log`). `P` liefert Ports, Regeln der Attribute, Zähler, `Snapshot`, `New` und `FakeTelemetry`, `S` die Spans samt Span-Zeile (`SPEC-231` Nr. 6). Die Entscheidungen des Architect (1 bis 9, Alternativen, Gründe, Folgen) stehen in `slice-telemetry-port` §3 und gelten hier unverändert; die Entscheidungen 10 bis 13 und 15 bis 17 zu Zeile und Senke stehen in §3 dieses Plans.

**Schnitt vom 2026-11-07 (Planner, vor dem Code, nach dem Vorschlag des Architect; vollzogen):** Nach der vollständigen Tabelle war *S* des ungeschnittenen Plans 3095 (*M* 4200) und lag über der Schnittgrenze 2223 (`SPEC-128` Nr. 1). Der Planner hat vor dem Code in zwei Teile geschnitten, wie der Architect vorgeschlagen hat; die Tabelle trug die Zeilen schon nach Teil getrennt (Präfix `L` und `H` der Nummer), die Nummern bleiben die des Architect:

- **Teil `L` — Zeile, Senke und Sperre** (dieser Slice, Kennung `slice-telemetry-log`, Zweig bleibt): `Log`, Zeitstempel, Stufen, `MinLevel`, `msg`, Zahlen, Senke mit Ausfall und Blockade, Sperre und Reihenfolge, Canary über `Log`; Zeilen `L01` bis `L20` (5 nebenläufig, 15 andere), `FakeSenke` (1), Produktcode 260; *S* 2085, *M* 2900. Die Sendung F-10 (Ersatz von `TestKernSchreibtNichts`) und der Zeilenschreiber, den `slice-telemetry-spans` für die Span-Zeile und `slice-telemetry-handler` für `Handle` braucht, liegen hier.
- **Teil `H` — Handler und Abbildung der Dienste** (neuer Plan `slice-telemetry-handler` in `open/`): `Handler()`, `Handle`, `Enabled`, die zehn Einträge der Dienste, die Arten von `slog.Attr`, `WithAttrs`, `WithGroup`, Canary über den Handler, Sperre und Uhr auf der Seite von `Handle`; Zeilen `H01` bis `H11` (2 nebenläufig, 9 andere), kein Fake (`FakeSenke` liefert dieser Slice), Produktcode 120; *S* 1010, *M* 1400. Abhängigkeit: Code von `L` (Zeilenschreiber, Senke, Sperre, `FakeSenke`, Hilfen der Mengen M1 bis M9), kein geteilter Mechanismus, der einen Schnitt nach dem Code bräuchte, weil vor dem Code geschnitten wird. Entscheidung 14, die Mengen M11 bis M15, der Anteil `H` von M16, die Zweige Z38 bis Z65, das Raster des Handlers und die Zeilen `H01` bis `H11` stehen seitdem dort.
- Folge für die Nachbarn (Planner, nachgezogen): `slice-telemetry-spans` hängt nur an `L` (Zeilenschreiber und Sperre), nicht an `H`; die Einbau-Slices brauchen `L`, `H` und `S`; die Sendungen `SPEC-121` Nr. 12 und die Umstellung der Dienste (§6) gehen an `slice-telemetry-einbau-gateway`.

**Übernimmt (Sendung von `slice-telemetry-port`, Teil `L`):** Zeile, Stufen, `msg`, Ausfall der Senke, Sperre der Zeile und der Canary über den Weg `Log`. Die Hälfte der Sendung für `Handler`, Abbildung der zehn Ereignisse der Dienste und den Canary über den Weg `Handler` hat der Planner beim Schnitt an `slice-telemetry-handler` gegeben (dessen §1 nimmt sie an).

**Übernimmt (Sendung von `slice-telemetry-port`, Review F-10):** Ersatz des Teils „`Log` schreibt keine Zeile“ von `TestKernSchreibtNichts`: sobald `Log` eine Zeile schreibt, hält dieser Teil des Tests im Geber nichts mehr fest, und dieser Slice ersetzt ihn durch die Fänger der Zeile (Zeile `L13`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ports, Attribut, Regeln der Schlüssel und Wertformen, Verwurf, `CountError`, `Snapshot`, `New` und `FakeTelemetry` — `slice-telemetry-port` (liegt beim Start dieses Slice in `done/`); dieser Slice baut auf diesem Code auf und ändert seine Regeln nicht. Die Zählprüfungen der Tests aus `slice-telemetry-port` ändert er nur dort, wo sie eine feste Zahl der Quelle tragen (Auftrag an den Implementer in §3).
- `StartSpan`, `Set`, `End`, Eltern-Kind, Verteilung, Ring, Span-Zeile (`SPEC-231` Nr. 6) und der Canary über die Wege der Spans — `slice-telemetry-spans`; die Span-Zeile nutzt den Zeilenschreiber und die Sperre dieses Slice und kommt danach. Der Teil „`StartSpan`, `Set`, `End` ohne Wirkung“ von `TestKernSchreibtNichts` bleibt dort.
- `Handler()`, `Handle`, `Enabled`, die Arten von `slog.Attr`, `WithAttrs`, `WithGroup`, die zehn Einträge der Dienste und der Canary über den Handler (`SPEC-231` Nr. 4 und 5) — `slice-telemetry-handler` (Schnitt vom 2026-11-07, nach diesem Slice); der Handler ruft den Zeilenschreiber dieses Slice, und dieser Slice legt ihn nicht an.
- Einbau in Dienste und Adapter des Produkts, Verdrahtung im Composition Root und Umstellung der Dienste `record`, `session` und `maintenance` auf `slog.New(Handler())` (`SPEC-231` Nr. 7) — `slice-telemetry-einbau-gateway` (nimmt die Sendung in seinem §1 an, Eintrag des Planners vom 2026-11-07); dieser Slice ändert `src/bootstrap` und die Dienste nicht.
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

- [ ] Zeile: `Log` schreibt je Aufruf ein JSON-Objekt mit den Feldern des Katalogs in fester Reihenfolge, Zeitstempel aus der `Clock` mit neun Nachkommastellen, vier Stufen mit `MinLevel`, `msg` bis 256 Bytes, `duration_ms` mit drei Nachkommastellen; bei doppeltem Schlüssel gilt der erste (`SPEC-231` Nr. 1 bis 3, `LH-QA-05`, `LH-FA-01`).
- [ ] Senke und Geheimnisse: fällt die Senke aus, ist die Zeile verloren, `senke` zählt und der Aufrufer merkt nichts; jede Zeile geht in einem `Write` unter einer Sperre hinaus, die Zeilen einer Goroutine bleiben in Reihenfolge, keine Sperre liegt über `Clock.Now()`, unter `make test-race` ohne Befund; ein Canary erreicht über `Log` keine Zeile, den Stand oder einen Fehlerwert (`SPEC-232` Nr. 2, 3, 5 c, d, f und g, `LH-QA-04`).
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
| `src/adapters/driven/telemetry/` | update | Zeile, Stufen, `msg`, Zeilenschreiber, Senke mit Sperre; `Log` aus `slice-telemetry-port` prüft und zählt bis dahin, schreibt aber keine Zeile (§6); die Senke `FakeSenke` liegt in den Tests des Pakets; der Zuschnitt des Schreibers trägt `Handle` (`slice-telemetry-handler`) und die Span-Zeile (`slice-telemetry-spans`) mit, `Handler()` liefert dieser Slice nicht |
| `src/adapters/driven/telemetry/*_test.go` aus `slice-telemetry-port` | update | die festen Zahlen 37, 7 und 5 in den Zählprüfungen (`reihenFest` und fünf weitere Stellen in `telemetry_test.go`, `TestFehlercodeForm` in `regeln_test.go`) hängen an der Quelle (`model.AlleCodes()`, `support.Verwuerfe()`, `support.Spans()`); der Teil „`Log` schreibt keine Zeile“ von `TestKernSchreibtNichts` entfällt (Zeile `L13`) |
| `spec/spezifikation.md` | update (Architect, erledigt) | `SPEC-231` Nr. 1 bis 3 und 5, `SPEC-232` Nr. 3, 5 d, 5 g und *Ungeprüft* vor dem Code entschieden, eine Zeile in §7 Historie; keine neue Nummer; eine Berichtigung zieht der Implementer im selben Commit nach (`AGENTS.md` §3.9) |
| `spec/architecture.md`, `.a-check.yml`, `go.mod`, `go.sum`, `docs/plan/adr/` | unverändert | wie in `slice-telemetry-port` §3 |
| `src/hexagon/ports/support/`, `src/bootstrap/`, die Dienste | unverändert | der Port steht; der Einbau ist `slice-telemetry-einbau-gateway` |
| Handbuch | unverändert | Binary unverändert (Entscheidung 7 des Architect) |
| Tests | neu | je Zusage ein Fall (Tabelle unten); Canary-Test; nebenläufige Tests laufen unter `make test-race` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation und wurden vom Architect geschrieben (Quelle: `slice-telemetry-port` bis zum Schnitt, Zeilen mit dem Präfix `L`, dann am 2026-11-07 vollständig gemacht); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidungen des Architect zu Zeile und Senke** (2026-11-07, vor dem Code; Fortsetzung der Entscheidungen 1 bis 9 aus `slice-telemetry-port` §3; der Wortlaut steht in der Spezifikation, `AGENTS.md` §3.8; Entscheidung 14 zum Handler steht in `slice-telemetry-handler`):

10. **Maskierung der Zeile** (`SPEC-231` Nr. 1): maskiert werden allein `"`, Backslash, die Zeichen unter U+0020 und die Zeilentrenner U+2028 und U+2029; `<`, `>`, `&` und DEL stehen unverändert. *Alternativen:* die HTML-Maskierung der Standardbibliothek (verworfen: die Spezifikation verlangt Zeilen ohne sie, und `&` in `msg` bliebe lesbar); U+2028 und U+2029 unmaskiert (verworfen: Leser, die an Zeilentrennern trennen, sähen zwei Zeilen). *Folge:* der Fänger prüft Rohbytes der Zeile, nicht nur den gelesenen Wert.
11. **Bildung von `msg`** (`SPEC-231` Nr. 3): erst je ungültigem Byte ein U+FFFD, dann der längste Anfang bis 256 Bytes an einer Zeichengrenze. *Alternative:* erst kürzen (verworfen: „Zeichengrenze“ ist in ungültigem UTF-8 nicht definiert, und das Ersetzen könnte die Grenze danach überschreiten); Läufe ungültiger Bytes zu einem U+FFFD (verworfen: die Länge hinge von der Nachbarschaft ab). *Folge:* ein Byte 0x80 am Offset 255 belegt danach drei Bytes (Menge M2).
12. **Zahlen** (`SPEC-231` Nr. 3): `duration_ms` wie `strconv.FormatFloat(v, 'f', 3, 64)`, ohne Exponenten; `befunde` als Ganzzahl. *Folge:* ein Wert wie `1e21` steht mit 22 Ziffern, ein `MaxFloat64` mit 309; die Länge der Zeile ist dafür nicht begrenzt (die Dauer ist endlich und ab 0, `SPEC-230` Nr. 3).
13. **`MinLevel`** (`SPEC-231` Nr. 2): der Vergleich nimmt den Wert des Levels; für einen Eintrag darunter wird `Clock.Now()` nicht gerufen, und seine Verwürfe zählen weiter. *Alternative:* erst nach `MinLevel` prüfen (verworfen: ein Verstoß einer Stelle bliebe bei `MinLevel` `ERROR` unsichtbar, und die Span-Zeile hält dasselbe Muster, `SPEC-231` Nr. 6: die Zeile entfällt, die Aufzeichnung bleibt; `TestLogOhneMinLevel` aus `slice-telemetry-port` hält es schon fest).
14. *(Handler, `SPEC-231` Nr. 5: steht mit Alternativen und Folge in §3 von `slice-telemetry-handler`; die Nummer bleibt hier frei, damit die Querverweise des Architect bestehen.)*
15. **Fehler der Senke** (`SPEC-232` Nr. 3): Fehler ist `err` oder weniger Bytes als die Zeile; `senke` zählt einmal je verlorener Zeile; mehr Bytes ohne Fehler ist kein Fehler. *Alternative:* auch mehr Bytes als Fehler (verworfen: die Zeile ging vollständig hinaus, ein Zähler `senke` meldete einen Verlust, der nicht eintrat).
16. **Zwei Sperren** (`SPEC-232` Nr. 5 g): die Sperre der Zeile liegt allein um `Write`; Zähler, `Snapshot` und `CountError` nehmen nur die Sperre des Zustands, und diese wird nie über das Warten auf die Sperre der Zeile gehalten. *Alternative:* eine Sperre für beides (verworfen: eine blockierende Senke hielte `Snapshot` und damit die Diagnose, und das Gebot „Aufzeichnung und Zähler hängen nicht an der Senke“, `SPEC-232` Nr. 3, wäre gebrochen).
17. **Synchron** (`SPEC-232` Nr. 5 d; Entscheidung 6 des Architect in `slice-telemetry-port` gilt): `Write` läuft im Aufruf; keine Warteschlange und keine Goroutine. *Folge:* eine Senke, die nie zurückkehrt, hält den Aufrufer (*Ungeprüft*, §6); eine Senke, die den Adapter selbst ruft, blockiert (neu in *Ungeprüft*).

**Mengen für die Fänger** (`AGENTS.md` §3.15, §3.17; Achsen, Elemente als Produkt, Generator, Orakel und Prüfgrenze; die Zählprüfung vergleicht die Zahl der Elemente mit der **Quelle** der Menge, nie mit einer festen Zahl im Test — `slice-mock-lesen` (Teil 2 des Schnitts von `slice-mock-zuordnung`) fügt `ODR-E6008` hinzu, 38 statt 37 Codes). Das Orakel ist nie der Adapter: die Zeile liest ein eigener Leser im Test (`json.Decoder`, für die Folge der Schlüssel `Token`; für Rohbytes `bytes`), und ein eigenes Modell der Zusage (Schleifen über `utf8.DecodeRuneInString`, `strconv`, `time.Format` mit der Zeichenkette der Spezifikation) liefert den erwarteten Wert. Die Mengen M1 bis M5, M7 bis M9 wiederholt `slice-telemetry-handler` über den Weg `Handle` mit denselben Generatoren: der Implementer legt Leser, Modelle und Listen so an, dass sie aus einem späteren Test desselben Pakets nutzbar sind. Die Nummern M11 bis M15 (Handler) stehen seit dem Schnitt in `slice-telemetry-handler`; die Nummern M1 bis M10 und M16 bleiben die des Architect.

| Menge | Achsen | Elemente (Produkt) | Generator und Quelle der Zählprüfung | Orakel | Prüfgrenze (akzeptiertes Negativ: Grund, Folge) |
|---|---|---|---|---|---|
| M1 Felder der Zeile (L01, L08; über `Handle` H04, H05 in `slice-telemetry-handler`) | Folge, Stelle | 2 hoch 13 Teilmengen × 2 Folgen des Aufrufs (Katalogfolge, umgekehrt) = 16384 | alle Teilmengen der Orakel-Liste der 13 Schlüssel; Zählprüfung: die Liste gleich den Schlüsseln, die `Filtern` mit gültigem Wert annimmt, aus den Kandidaten Orakel-Liste, Pflicht-Attribute aus `support.Spans()` (11) und die 12 fremden aus `TestReihenFest` | `Token`-Leser: Schlüssel in Dateifolge und Werte gleich dem Modell (Katalogfolge, gefiltert auf die Teilmenge) | je Schlüssel ein gültiger Wert: die Wertformen fängt `slice-telemetry-port` (P01 bis P05); Folge: eine Form, die nur in der Zeile, nicht im Filter bricht, fiele hier nur für diesen Wert auf |
| M2 Länge von `msg` (L05; H04) | Länge, Zeichen | 301 (Länge 0 bis 300, ASCII) + 160 (Breite 1 bis 4 des Zeichens an Offset 250 bis 257 × Folgezeichen keines oder Breite 1 bis 4: 4 × 8 × 5) + 8 (ein Byte 0x80 an Offset 250 bis 257) = 469 | Schleifen über Länge, Offset, Breite; Zählprüfung: die Schleifengrenzen (`256` ist die Konstante der Spezifikation, die Offsets sind `grenze-6` bis `grenze+1`) | Modell: je ungültigem Byte U+FFFD, dann längster Anfang bis 256 Bytes an Rune-Grenze | Texte bis 300 Bytes und die Grenze bei 256; Folge: ein Fehler des Kürzens, der erst bei einer anderen Länge wirkt, bliebe ungefangen; der Code kürzt an einer Stelle mit einer Konstante |
| M3 Zeichen von `msg` (L06; H04) | Zeichen, Stelle | 256 Einzelbytes + 3840 Paare (erstes Byte 0 bis 255 × zweites aus B) + 3600 Tripel (erstes Byte E0 bis EF × zweites und drittes aus B) + 1200 Quadrupel (erstes Byte F0 bis F4 × zweites aus B × drittes und viertes aus 0x80, 0x8F, 0x90, 0xBF) + 13 benannte Zeichen (U+0080, U+07FF, U+0800, U+D7FF, U+E000, U+FFFF, U+10000, U+10FFFF, U+2028, U+2029, BOM, NBSP, U+FFFD) = 8909 | B = 0x00, 0x22, 0x5C, 0x61, 0x7F, 0x80, 0x8F, 0x90, 0x9F, 0xA0, 0xBF, 0xC0, 0xC2, 0xF5, 0xFF (15 Bytes); jedes Element aus Bytewerten gebaut (`AGENTS.md` §3.18); Zählprüfung: 256 gleich der Schleife `0` bis `255`, die übrigen Zahlen gleich dem Produkt der Längen der Listen | Rohbytes der Zeile: ein `\n` am Ende und kein weiteres Byte unter 0x20, keine Folge E2 80 A8 oder E2 80 A9, `<`, `>`, `&` und 0x7F als Rohbyte, wenn sie im Text stehen; gelesener `msg` gleich dem Modell | Folgen aus mehr als vier Bytes und Bytes außerhalb von B in zweiter bis vierter Stelle; Grund: die Dekodierung liegt in `unicode/utf8`, der Code ersetzt je Byte in einer Schleife; Folge: eine Sonderbehandlung für eine Bytefolge außerhalb der Menge bliebe ungefangen |
| M4 Stufe × `MinLevel` (L03, L04; H07) | Seite, Zeichen der Stufe | 9 × 9 = 81 Paare | die 9 Werte −8, −4, −1, 0, 3, 4, 7, 8, 12; Zählprüfung: für jede Grenze g aus `slog.LevelInfo`, `slog.LevelWarn`, `slog.LevelError` stehen g−1 und g in der Liste | Stufe gleich der Tabelle der Spezifikation (Name nach Wert), Zeile genau dann, wenn Stufe ≥ Minimum, `Now()` einmal je Zeile und nie sonst | Werte zwischen den neun (etwa 1, 2, 5, 6); Grund: die Abbildung vergleicht an drei Grenzen, beide Seiten jeder Grenze stehen in der Liste; Folge: ein Fehler, der nur bei einem Wert zwischen den Grenzen wirkt, bliebe ungefangen |
| M5 Zeitstempel (L02; H07) | Zustand | 8: 0 ns, 1 ns, 10 ns, 123456789 ns, 999999999 ns, Zone +05:30, Zone −08:00, `time.Time{}` | Liste; Zählprüfung: Länge der Liste | `time.Format` mit `2006-01-02T15:04:05.000000000Z` auf `t.UTC()` | Jahre außerhalb 0000 bis 9999: Format nicht zugesagt (`SPEC-231` Nr. 2); Folge: keine, die Systemuhr liefert sie nicht |
| M6 Zahlen (L07) | Zustand, Grenze | 9 Werte `duration_ms` (0, 5e-324, 0.0005, 0.9995, 1.0005, 12.345, 5000, 1e21, `MaxFloat64`) + 5 Werte `befunde` (0, 1, 9, 10, `MaxInt64`) = 14 | Listen; Zählprüfung: Länge | `strconv.FormatFloat(f, 'f', 3, 64)` und `strconv.FormatInt`; kein `e` und kein `E` in der Zeile | andere Gleitkommawerte; Grund: die Formatierung ist ein Aufruf der Standardbibliothek; Folge: gering |
| M7 Doppelter Schlüssel (L08; H05) | Folge, Stelle | 13 Schlüssel × 14 Positionen des zweiten Werts (vor dem ersten Attribut bis hinter dem letzten) = 182 | Schleife über die Orakel-Liste (M1) und die Positionen 0 bis 13 | der Wert, der in der Liste des Aufrufs zuerst steht; `schluessel` zählt 1 | — |
| M8 Argumente von `Log` (L09; H08) | Abbruch, Argumentart | 3 `ctx` (nil, beendet, abgelaufene Frist) × 6 Arten (nil-Slice, leerer Schlüssel, NaN, Art 3, Art −1, Wert mit ungültigem UTF-8) = 18 | Schleifen | eine Zeile mit `timestamp`, `level`, `msg` und den gültigen Feldern; `Dropped` gleich dem Modell; ein Fall, der in Panik gerät, färbt seinen Namen (`recover` je Fall) | — |
| M9 Senke (L10, L12; H08) | Zustand | 7 Formen der Rückgabe (n, err): (0, Fehler), (kleiner als Zeile, Fehler), (Zeile, Fehler), (0, nil), (kleiner als Zeile, nil), (größer als Zeile, nil), (Zeile, nil) × 2 Zeilen (nur `timestamp`/`level`/`msg`, alle 13 Felder) = 14, dazu Blockade (L11) | Liste von Funktionen der `FakeSenke`; Zählprüfung: Länge der Liste gleich 7 aus den vier Zeilen der Tabelle der Fehlerformen in §6 | `senke` gleich dem Modell (1 für die ersten fünf, 0 für die letzten zwei), ein `Write` je Zeile, die nächste Zeile ganz in der Senke | `n` zwischen 0 und Zeile mit jedem Wert: der Code vergleicht `n` mit der Länge; Folge: gering |
| M10 Canary über `Log` (L14) | Stelle, Ausgabe | 13 Werte der Schlüssel + 1 fremder Schlüssel + 1 Fehlertext der Senke = 15 Stellen × 4 Einbringungswege (Header `Bearer …`, Body-Fragment mit Anführungszeichen, Query-Literal `$filter=Name eq '…'`, Konfigurationswert `https://user:…@host`) × 2 Ausgaben (Bytes der Senke, `%+v` des Stands) = 120 | Schleifen; Zählprüfung 13 gegen M1, 4 gegen die Liste der Wege | weder das Canary noch sein Präfix von 4 Bytes in der Ausgabe; das Präfix steht in der Zeile eines Aufrufs mit gültigen Werten nicht (Kontrolle) | der Inhalt von `msg` (L15); der Fehlerwert von `New` und die Wege der Spans (`slice-telemetry-spans`); weitere Canary-Formen; Folge: siehe L15 |
| M16 Rückrufe der Uhr (L19, L20; H11) | Seite, Stelle | `Log`: Rückrufe `CountError`, `Snapshot` (L19) und `Log` (L20) = 3; die Seite `Handle` (4 Rückrufe) führt `slice-telemetry-handler` | Liste von Funktionen in `Now()` der `RueckrufUhr` (ruft den Rückruf beim ersten `Now()`) | der Aufruf kehrt innerhalb 2 s zurück; bei Rückruf `Log` stehen zwei Zeilen in der Senke | — |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Namen der Tests sind Vorschläge, der Fänger ist die beschriebene Schleife mit Orakel. Ein akzeptiertes Negativ zählt `andere`. Die Nummer trägt den Teil des Schnitts in §1: `L` Zeile, Senke und Sperre (dieser Plan); die Zeilen `H` des Handlers stehen in `slice-telemetry-handler`. `FakeSenke` (Hülle eines `io.Writer`) hält jedes Argument von `Write` als Kopie, zählt jeden `Write`, hält den Höchstwert gleichzeitiger `Write`, kann je Aufruf eine Rückgabe (n, err) liefern und blockieren (meldet den Eintritt, wartet auf die Freigabe des Tests); sie verdeckt weder Blockieren noch Fehler noch die Zahl der Aufrufe.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| L01 | Eine Zeile ist ein JSON-Objekt mit genau einem `\n` am Ende, die Felder stehen in der Folge des Katalogs, ein nicht gesetztes Feld fehlt, jeder Aufruf schreibt eine Zeile (`SPEC-231` Nr. 1) | `TestZeilenform`: Menge M1 (16384 Aufrufe); zusätzlich ruft jeder Aufruf denselben Eintrag zweimal: zwei Zeilen; Orakel der `Token`-Leser gegen das Modell | `level` vor `timestamp`; nicht gesetztes Feld als leerer Text; Felder in der Folge des Aufrufs; `duration_ms` vor `error_code` | andere |
| L02 | `timestamp` ist `Clock.Now()` in UTC mit neun Nachkommastellen, `Now()` wird je geschriebener Zeile einmal gerufen (`SPEC-231` Nr. 2) | `TestTimestamp`: Menge M5 (8 Fälle) über `Log`; `FestUhr` zählt die Aufrufe von `Now()` | `time.RFC3339Nano` (schneidet Nullen ab); lokale Zone; `Now()` zweimal gerufen | andere |
| L03 | Die Stufe wird auf vier Namen abgebildet, an drei Grenzen (`SPEC-231` Nr. 2) | `TestStufen`: die 9 Werte aus M4, beide Seiten jeder Grenze; Orakel die Tabelle der Spezifikation | Text von `slog.Level.String` (`INFO+2`); `<=` statt `<` an einer Grenze | andere |
| L04 | Einträge unter `MinLevel` entfallen, die Gleichheit wird geschrieben; verglichen wird der Wert; `Now()` wird für sie nicht gerufen; ihre Verwürfe zählen (`SPEC-231` Nr. 2) | `TestMinLevel`: Menge M4 (81 Paare), jedes Paar mit einem gültigen und einem ungültigen Attribut (`mode` mit Leerzeichen); Orakel „Stufe ≥ Minimum“, `wert` 1 je Aufruf auch unter dem Minimum, `Now()` einmal oder nie; insbesondere MinLevel 1 mit Stufe 0 (entfällt) und 3 (geschrieben) | `>` statt `>=`; Vergleich über den Namen; Verwurf nur bei geschriebener Zeile gezählt; `Now()` immer gerufen | andere |
| L05 | `msg` wird nach dem Ersetzen ungültiger Bytes auf den längsten Anfang bis 256 Bytes an einer Zeichengrenze gekürzt, leer ist zulässig (`SPEC-231` Nr. 3) | `TestMsgLaenge`: Menge M2 (469 Fälle); Orakel das Modell | Schnitt bei 255; Kürzen mitten im Zeichen; Kürzen vor dem Ersetzen (Länge des Rohtextes) | andere |
| L06 | Maskierung und ungültiges UTF-8: maskiert werden allein `"`, Backslash, Zeichen unter U+0020, U+2028 und U+2029; ein ungültiges Byte steht als je ein U+FFFD (`SPEC-231` Nr. 1 und 3) | `TestMsgZeichen`: Menge M3 (8909 Fälle); Rohbytes und gelesener Wert; Fixtures aus Bytewerten, die Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18) | HTML-Maskierung an (`<` wird zur Folge mit Hex); U+2028 unmaskiert; `strings.ToValidUTF8` (ein U+FFFD je Lauf); Formatierung mit `%q` | andere |
| L07 | `duration_ms` steht wie `FormatFloat(v, 'f', 3, 64)` ohne Exponenten, `befunde` als Ganzzahl (`SPEC-231` Nr. 3) | `TestZahlen`: Menge M6 (14 Werte); Orakel `strconv`, kein `e` in der Zeile | Ausgabe mit `%v` oder `'g'`; Rundung abgeschnitten; `befunde` als `7.0` | andere |
| L08 | Bei doppeltem Schlüssel steht der erste Wert in der Zeile, `schluessel` zählt 1 (`SPEC-230` Nr. 4) | `TestDoppelterSchluessel`: Menge M7 (182 Fälle), beide Werte gültig und verschieden | der letzte Wert gewinnt; Sonderfall Position 0 | andere |
| L09 | `Log` kehrt immer normal zurück und schreibt, bei nil-`ctx`, beendetem `ctx`, abgelaufener Frist und jeder Argumentart (`SPEC-230` Nr. 1, `SPEC-232` Nr. 3) | `TestLogArgumente`: Menge M8 (18 Fälle); Orakel eine Zeile und `Dropped` gleich dem Modell | `Log` prüft `ctx.Err()`; `ctx.Value` ohne nil-Prüfung; Zugriff auf `attrs[0]` ohne Länge | andere |
| L10 | Fällt die Senke aus, ist die Zeile verloren, `senke` zählt einmal je Zeile, die nächste wird geschrieben, der Aufrufer merkt nichts; mehr Bytes ohne Fehler ist keiner; Verwürfe und `senke` zählen unabhängig (`SPEC-232` Nr. 3) | `TestSenkeAusfall`: Menge M9 (14 Fälle), dazu ein Eintrag mit einem ungültigen Attribut und defekter Senke (`wert` 1 und `senke` 1); Orakel `senke` aus dem Modell, ein `Write` je Zeile (kein Wiederholen), nächste Zeile ganz | der Fehler löst eine Wiederholung aus; `n` gleich der Zeile mit Fehler gilt als Erfolg; `senke` zweimal je Zeile; Rest nachgeschrieben; `n` größer als die Zeile zählt `senke` | andere |
| L11 | Zähler, `Snapshot` und `CountError` warten nicht auf eine blockierte Senke (`SPEC-232` Nr. 5 g) | `TestBlockierteSenke`: `FakeSenke` blockiert im `Write`; Goroutine A ruft `Log` und hängt; B ruft `CountError`, `Snapshot` und ein `Log` unter `MinLevel` mit ungültigem Attribut (3 Operationen); Orakel jede kehrt innerhalb 2 s zurück, `wert` zählt, nach der Freigabe kehrt A zurück, genau eine Zeile; deterministisch, Haltepunkt am `Writer` | `Write` unter der Sperre des Zustands; Zähler hinter der Sperre der Zeile | nebenläufig |
| L12 | Ein `Write` je Zeile mit der ganzen Zeile samt `\n` (`SPEC-231` Nr. 1, `SPEC-232` Nr. 3) | `TestEinWrite`: drei verschiedene Zeilen hintereinander (nur Pflichtfelder, alle 13 Felder, `msg` mit Maskierung); Orakel Zahl der `Write` gleich Zahl der Aufrufe, jedes Argument endet auf `\n` und trägt genau eines, die Kopien sind unverändert und verschieden | Zeile und `\n` in zwei `Write`; zwei Zeilen in einem `Write`; Puffer nicht zurückgesetzt (Rest der ersten Zeile in der zweiten) | andere |
| L13 | Ersatz des Teils „`Log` schreibt keine Zeile“ von `TestKernSchreibtNichts` (Sendung F-10): `Log` schreibt je Aufruf über `MinLevel` eine Zeile, darunter keine | `TestLogSchreibtZeile`: die drei `Log`-Aufrufe des alten Tests (Stufe `ERROR` mit `mode`, Stufe `DEBUG`, nil-`ctx` mit `INFO`) gegen `neu(t)` (`MinLevel` `INFO`): zwei Zeilen in dieser Folge, der Eintrag `DEBUG` keine; aus `TestKernSchreibtNichts` entfallen die `Log`-Aufrufe, der Rest (Spans ohne Wirkung, Puffer leer, Stand fest) bleibt für `slice-telemetry-spans` | `Log` schreibt nichts; `Log` schreibt auch unter `MinLevel` | andere |
| L14 | Ein Canary erreicht über `Log` keine Zeile und keinen Stand (`SPEC-232` Nr. 2, `LH-QA-04`) | `TestCanary`: Menge M10 (120 Prüfungen) | ein ungültiger Wert wird gekürzt statt verworfen; der Fehlertext der Senke steht im Stand; der Wert eines verworfenen Attributs steht in der Zeile | andere |
| L15 | Akzeptiertes Negativ: ein wohlgeformter Wert in einem zulässigen Schlüssel und der Inhalt von `msg` gehen durch (`SPEC-232` Nr. 2, *Ungeprüft*) | kein Fänger; Grund: die Form ist prüfbar, der Inhalt nicht; Folge: eine Stelle, die einen Request-Wert in `session_id` oder `msg` legt, gibt ihn aus; Fänger je Stelle im Einbau, Review | — | andere |
| L16 | Akzeptierte Negative der Senke und der Uhr: Panik von `Writer` oder `Clock`, eine Senke, die nie zurückkehrt, eine Senke, die den Adapter ruft, ein Jahr außerhalb 0000 bis 9999 (`SPEC-232` *Ungeprüft*, `SPEC-231` Nr. 2) | kein Fänger; Grund: die Ports paniken nie, und die Senke des Binary ist der Standard-Fehlerstrom, der den Adapter nicht ruft; Folge: der Aufrufer wartet mit, eine Senke, die den Adapter ruft, blockiert alle Schreiber, ein fernes Jahr hat ein anderes Format | — | andere |
| L17 | Die Senke sieht nie zwei gleichzeitige `Write`, Zeilen vermischen sich nicht (`SPEC-232` Nr. 5 c) | `TestKeineZweiWrite`: (1) Haltepunkt: A blockiert im `Write`, B ruft `Log`, das zweite `Write` darf innerhalb 200 ms nicht eintreten (Höchstwert 1), nach der Freigabe stehen beide Zeilen da; (2) Masse: 32 Goroutinen × 200 Einträge, Orakel 6400 Zeilen, jede ein gültiges Objekt, Höchstwert 1; Fenster der Sperre in §6 | `Write` ohne Sperre (der Haltepunkt wird in jedem Lauf rot); Sperre je Aufruf statt je Adapter | nebenläufig |
| L18 | Die Zeilen einer Goroutine stehen in der Reihenfolge ihrer Aufrufe, `Write` läuft im Aufruf (`SPEC-232` Nr. 5 d) | `TestZeilenReihenfolge`: (1) Haltepunkt: `Log` darf nicht zurückkehren, solange `Write` blockiert (100 ms Beobachtung), nach der Freigabe kehrt es zurück; (2) Masse: 32 Goroutinen, jede schreibt `befunde` 1 bis 200 mit eigener `session_id`, Orakel je `session_id` streng steigend | `Write` in einer eigenen Goroutine; Warteschlange mit Schreiber | nebenläufig |
| L19 | Keine Sperre des Zustands liegt über `Clock.Now()` (`SPEC-232` Nr. 5 f, Seite `Log`) | `TestUhrRuftZustand`: `RueckrufUhr` ruft beim ersten `Now()` `CountError` oder `Snapshot` (M16, 2 Fälle); Orakel „kehrt zurück“ innerhalb 2 s; deterministisch | `Now()` unter der Sperre des Zustands | nebenläufig |
| L20 | Keine Sperre der Zeile liegt über `Clock.Now()` (`SPEC-232` Nr. 5 f, Seite `Log`) | `TestUhrRuftZeile`: `RueckrufUhr` ruft beim ersten `Now()` `Log` (M16); Orakel „kehrt zurück“ innerhalb 2 s, zwei Zeilen, die innere zuerst; deterministisch | `Now()` unter der Sperre der Zeile | nebenläufig |
Zählung dieser Tabelle: 20 Zeilen, davon 5 `nebenläufig` (*a*: L11, L17, L18, L19, L20) und 15 `andere` (*b*, darunter die akzeptierten Negative L15 und L16). Die Zeilen `H01` bis `H11` (2 und 9) führt seit dem Schnitt vom 2026-11-07 `slice-telemetry-handler`.

**Bericht des Implementers zum Handoff** (`AGENTS.md` §3.10, §3.15, §3.17; Zusage · Mutation · roter Test). Alle Mutationen fuhr `tools/harness/mutationslauf.sh` in einem Container ohne Netz (`--memory=2g`, `--cpus=4`, Image `odata-recorder:race`, Messende: das Werkzeug; Last des Hosts 1,4 bis 2,8 von 4 CPUs zu Beginn der Mutation); 55 Mutationen, jede am Code oder an der Sperre, nie an der Fixture. Klasse `andere`: ein Prozess, `rot` heißt, der benannte Test endet mit Exit ungleich 0; 43 von 43 rot, 0 `uebersetzungsfehler` (zwei erste Fassungen, die nicht übersetzten, wurden berichtigt und neu gefahren).

| Nr. | Zusage | Mutation (Name in der Liste) | roter Test |
|---|---|---|---|
| L01 | Zeile, Folge, `\n` | Sortierung entfernt (`l01-ohne-sortierung`); `duration_ms` vor `error_code` (`l01-duration-vor-error-code`); ohne `\n` (`l01-ohne-zeilenende`) | `TestZeilenform` |
| L02 | Zeitstempel | `time.RFC3339Nano` (`l02-rfc3339nano`); lokale Zone (`l02-lokale-zone`); `Now()` zweimal (`l02-now-zweimal`) | `TestTimestamp` |
| L03 | Stufen | `<=` an jeder der drei Grenzen (`l03-grenze-info`, `-warn`, `-error`); `Level.String` (`l03-level-string`) | `TestStufen` |
| L04 | `MinLevel` | `<=` (`l04-gleich-entfaellt`); Vergleich über den Namen (`l04-vergleich-name`); Verwurf nur bei geschriebener Zeile (`l04-verwurf-nur-geschrieben`); `Now()` immer (`l04-now-immer`) | `TestMinLevel` |
| L05 | Länge von `msg` | Konstante 255 (`l05-schnitt-255`); Schnitt mitten im Zeichen (`l05-mitten-im-zeichen`); Kürzen nach den Bytes des Rohtextes (`l05-kuerzen-vor-ersetzen`) | `TestMsgLaenge` |
| L06 | Maskierung, UTF-8 | `<` maskiert (`l06-html-maskierung`); U+2028 offen (`l06-2028-offen`); U+2029 offen (`l06-2029-offen`); DEL maskiert (`l06-del-maskiert`); 0x1F offen (`l06-1f-offen`); ein U+FFFD je Lauf (`l06-fffd-je-lauf`) | `TestMsgZeichen` |
| L07 | Zahlen | `'g'` (`l07-format-g`); abgeschnitten (`l07-abgeschnitten`); zwei Stellen (`l07-zwei-stellen`); `befunde` als `7.0` (`l07-befunde-als-float`) | `TestZahlen` |
| L08 | Doppelter Schlüssel | zweiter Wert angenommen (`l08-doppelter-schluessel-gilt`, in `regeln.go`) | `TestDoppelterSchluessel` |
| L09 | `Log` kehrt immer zurück | `ctx.Err()` geprüft (`l09-ctx-err`); `attrs[0]` ohne Länge (`l09-attrs-null`) | `TestLogArgumente` |
| L10 | Ausfall der Senke | Fehler mit allen Bytes gilt als Erfolg (`l10-fehler-mit-allen-bytes-ok`); weniger Bytes ohne Fehler gilt als Erfolg (`l10-kurz-ohne-fehler-ok`); mehr Bytes zählt (`l10-mehr-bytes-zaehlt`); `senke` doppelt (`l10-senke-doppelt`); Wiederholung (`l10-wiederholung`); Rest nachgeschrieben (`l10-rest-nachgeschrieben`) | `TestSenkeAusfall` |
| L12 | Ein `Write` je Zeile | zwei `Write` (`l12-zwei-write`); Puffer wiederverwendet (`l12-puffer-wiederverwendet`); Puffer nicht zurückgesetzt (`l12-puffer-nicht-zurueckgesetzt`) | `TestEinWrite` |
| L13 | Ersatz F-10 | `Log` schreibt nichts (`l13-log-schreibt-nichts`); auch unter `MinLevel` (`l13-auch-unter-minlevel`) | `TestLogSchreibtZeile` |
| L14 | Canary | ungültiger Wert bleibt in der Zeile (`l14-ungueltiger-wert-bleibt`, in `regeln.go`); Fehlertext der Senke im Stand (`l14-fehlertext-im-stand`) | `TestCanary` |
| L11 | Zähler warten nicht | `Write` unter der Sperre des Zustands; `CountError`, `Snapshot`, Zählung der Verwürfe je hinter der Sperre der Zeile (vier Mutationen `l11-*`) | `TestBlockierteSenke` |
| L17 | Nie zwei `Write` | ohne Sperre; Sperre je Aufruf statt je Adapter (`l17-*`, Haltepunkt; `l17m-*`, Masse) | `TestKeineZweiWriteHaltepunkt`, `TestKeineZweiWriteMasse` |
| L18 | Reihenfolge, `Write` im Aufruf | `Write` in eigener Goroutine (`l18-write-in-goroutine`, `l18m-write-in-goroutine`) | `TestLogKehrtNachWriteZurueck`, `TestZeilenReihenfolgeMasse` |
| L19 | Keine Sperre des Zustands über `Now()` | `Now()` unter `mu` (`l19-now-unter-zustand`) | `TestUhrRuftZustand` |
| L20 | Keine Sperre der Zeile über `Now()` | `Now()` unter `zeilenMu` (`l20-now-unter-zeile`) | `TestUhrRuftZeile` |

**Raten der Klasse `nebenläufig`** (12 Mutationen; je 50 Prozesse mit `-race` und `-test.count=1`, 50 ohne Detektor; vier Prozesse zugleich; Messende: `mutationslauf.sh`; Last des Hosts 1,4 bis 2,5 von 4 CPUs): die Haltepunkt-Fänger sind in jedem Lauf rot, `l11-*` (4), `l17-*` (2), `l18-write-in-goroutine`, `l19`, `l20` je 50 von 50 mit und 50 von 50 ohne Detektor; die Masse-Fänger daneben `l17m-ohne-sperre` und `l17m-sperre-je-aufruf` ebenfalls 50 von 50 und 50 von 50 (die Senke ruft `runtime.Gosched()` zwischen Eintritt und Austritt und weitet das Fenster); `l18m-write-in-goroutine` ist `flatternd`: 24 von 50 mit Detektor, 49 von 50 ohne. Die Zusage „`Write` läuft im Aufruf“ trägt deshalb allein der Haltepunkt (`TestLogKehrtNachWriteZurueck`, 100 von 100); die Masse (`TestZeilenReihenfolgeMasse`) zählt als zusätzlicher Fänger, nicht als der der Zusage. Die Zusage „unter einer Sperre, ohne sie dazwischen freizugeben“ hat keine Fenster-Mutation außer den in §6 benannten (die Sperre der Zeile liegt allein um `Write`, die Fenster tragen Haltepunkte am `Writer` oder an `Now()`).

**Mengen** (Generator · Orakel · Zahl der Elemente, die der Test zählt): M1 alle 2^13 Teilmengen × 2 Folgen = 16384 Aufrufe, je zweimal (32768 Zeilen), Orakel `json.Decoder.Token` gegen das Modell, Zählprüfung gegen den Filter (`logSchluessel`: aus 13 + 11 Pflicht-Attributen der Spans + 12 fremden nimmt `Filtern` genau die Orakel-Liste an); M2 469 (301 + 160 + 8); M3 8909 (256 + 3840 + 3600 + 1200 + 13, aus Bytewerten, Zählprüfung als Produkt der Längen der Listen); M4 81 Paare (+ 4 Stufen bei `MinLevel` 1), Zählprüfung: an jeder Grenze stehen g−1 und g in der Liste; M5 8; M6 14; M7 182; M8 18; M9 14 (7 Formen × 2 Zeilen, dazu der Fall Verwurf und defekte Senke); M10 120 (15 Stellen × 4 Wege × 2 Ausgaben) mit Kontrolle (Canary in `session_id` steht in der Zeile; das Präfix steht in der Zeile gültiger Werte nicht); M16 (L-Anteil) 3 Rückrufe. Die Zahlen 37, 7 und 5 der Tests von `slice-telemetry-port` hängen an `model.AlleCodes()`, `support.Verwuerfe()`, `support.Spans()` und `support.Eimer()`; die Orakel-Listen `alleCodes` (Codes), die Liste der Gründe in `TestSnapshotVollstaendig` und `spanNamen` bleiben Textlisten, ihre Länge wird gegen die Quelle gehalten und die Meldung sagt „Orakel-Liste nachziehen“. Prüfgrenzen: wie in der Tabelle der Mengen (akzeptierte Negative mit Grund und Folge); dazu: die Zählung der Aufrufe der Senke ist die der Kopien, die der Fake hält.

**Zweig-Menge Z01 bis Z37** (Abdeckung unter dem Detektor: `go test -race -covermode=atomic -coverprofile` über das Paket in der Stufe `race`, Container ohne Netz, `--memory=4g`): Anweisungen des Pakets 98,7 %; alle Funktionen der neuen Datei `zeile.go` (`logFolge`, `eintrag`, `zeileSchreiben`, `stufe`, `msgText`, `anfuehren`) und `Log`, `CountError`, `Snapshot`, `verwerfen` 100,0 %. Blöcke ohne Treffer im Paket: `Set` und `End` des Spans ohne Wirkung (Gegenstand von `slice-telemetry-spans`, nicht dieses Slice), `regeln.go` Zeile 71 (Rückgabe von `spanRegel` für einen Schlüssel außerhalb; Bestand von `slice-telemetry-port`, der Zweig ist von `schluessel` aus nicht erreichbar) und Zeile 136 (`wertGueltig` hinter der Form, Bestand, nicht erreichbar); in den Dateien dieses Slice kein Block ohne Treffer. Grenze der Zählung: eine Anweisungsabdeckung sieht die zwei Seiten einer Bedingung nicht (etwa `err != nil || n < len(z)`); diese Seiten tragen die Mutationen `l10-*` und die sieben Formen der Menge M9.

**Gegengelesen** (`AGENTS.md` §3.18): `hilfen_test.go`, `zeile_test.go`, `sperre_test.go` und `zeile.go` im Verzeichnis `src/adapters/driven/telemetry` auf eine Folge aus Backslash, `u` und vier Hex-Ziffern mit `grep -P` durchsucht: kein Treffer; Werte jenseits von ASCII entstehen aus Zahlen (`string(rune(0x2028))`, `[]byte{...}`).

**Zweige als Menge** (`AGENTS.md` §3.15; Register `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`): die Modell-Menge der Wege von `Log` und der Senke steht in §6 (Zweig-Menge, Z01 bis Z37). Die Zählprüfung ist die Abdeckung unter dem Detektor: der Implementer fährt `go test -race -coverprofile` über das Paket in der Stufe `race` (unter `flock`, mit `--memory`) und nennt im Bericht jeden Block der neuen Dateien ohne Treffer; null, oder je Block ein benanntes Negativ mit Grund und Folge. Eine zweite Zählung im Test selbst entfällt: ein Test kann die Zweige des Codes nicht zählen, ohne den Code zu lesen, und die Abdeckung der Blöcke tut es ohne neues Werkzeug.

**Aufträge an den Implementer aus dieser Antwort:** (1) Die festen Zahlen der Tests aus `slice-telemetry-port` (`reihenFest` und die Stellen mit `37`, `41`, `3101`, `7`, `5` in `telemetry_test.go`; die Prüfung der Texte in `regeln_test.go`) hängen an der Quelle: `len(model.AlleCodes())`, `len(support.Verwuerfe())`, die Zahl der Namen aus `support.Spans()`. Die Liste `alleCodes` im Test bleibt eine Liste der Texte aus dem Wortlaut von `SPEC-109` Nr. 2 (das unabhängige Orakel); ihre Länge wird gegen `len(model.AlleCodes())` gehalten, und die Meldung nennt „Orakel-Liste nachziehen“, wenn ein Code hinzukommt; den Text `ODR-E6008` trägt `slice-mock-lesen` in die Liste ein (Sendung in dessen §1, Planner, 2026-11-07), nicht dieser Slice (§6). (2) Fixtures der Menge M3 entstehen aus Bytewerten; der Bericht nennt die Testdatei, die auf Rohbytes gegengelesen wurde (`AGENTS.md` §3.18). (3) Eine Zelle des Rasters ohne Ort in der Spezifikation geht an den Architect zurück (`AGENTS.md` §3.12).

**Antworten des Architect auf die Aufträge vom 2026-11-07:** (1) Tabelle vollständig: oben, mit L13 für F-10, den Zeilen für die Teile, die der Satz nur nannte (keine HTML-Maskierung in L06, `\n` in L01, `Handle` gibt Text und Stufe unverändert in H01 und H04, ein `Write` je Zeile in L12, Gleichheit an jeder Grenze in L02 bis L05 und H07). (2) Raster und Fehlerformen der Senke: §6. (3) Mengen mit Achsen: oben, Zählprüfung gegen die Quellen. (4) Zweig-Menge: §6, Zählprüfung als Abdeckung. (5) Schnitt: §1 und §4. (6) Kennungen: keine neue Nummer, alle Änderungen in `SPEC-231` und `SPEC-232`, eine Zeile in der Historie der Spezifikation. (7) Katalog: kein neuer Code, Schlüssel oder Kommando, kein Logfeld am Binary sichtbar (`SPEC-232` Nr. 6); `harness/katalog-ausnahmen.txt` und das Handbuch bleiben unberührt, §7 sagt, dass das Binary unverändert blieb.

**Betriebsregeln für alle Rollen:**

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts (Entscheidung des Users vom 2026-10-21).
- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4` (höchstens zwei Container je Linie zugleich) unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen. Die Image-Falle: das Image `odata-recorder:race` trägt `/src/src` eingebacken; ein Lauf, der Quellen hineinkopiert, leert den Pfad zuvor oder nutzt die Stufe `deps`. Ein Rot unter Last wird auf ruhigem Host wiederholt, bevor es als Befund gilt.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`: eine weitere Verletzung ist ein Befund des Reviewers); keine Hilfsdateien im Wurzelverzeichnis. Commit-Nachrichten gehen als Datei (`git commit -F`), nicht als Heredoc.
- Eine Hülle um ein Argument (ein Fake, der `ctx`, einen Strom oder einen Writer umhüllt, ein Test-Writer der Senke) darf keine Eigenschaft des Originals verdecken (Abbruch, Fehler, Blockieren, Zahl der Aufrufe; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`): `FakeSenke` hält jede Zeile fest, zählt jeden `Write`, hält den Höchstwert gleichzeitiger `Write`, kann Fehler, Kurzschreiben und Blockieren wie ein echter `io.Writer`, und eine Panik des Fakes färbt den benannten Test, nicht einen anderen.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Die Klassen der Tabelle sind generierbar (hier: jedes Feld des Katalogs, jede Stufe, jedes Zeichen von `msg`): schon der Handoff führt den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben wird nur im Integrationsschritt gesetzt, nicht im Slice-Commit.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13); der Bericht allein ist keine Adresse.
- Der Implementer hält an einer Randform ohne Ort in der Spezifikation an und gibt sie an den Architect zurück (`AGENTS.md` §3.12); er entscheidet sie nicht im Code.
- Wert mit Escape-Folge oder Rohbytes: der Plan und die Berichte schreiben eine solche Folge nicht als Zeichen, sondern beschrieben (Werkzeug-Falle Backslash-u, `AGENTS.md` §3.18).

**Aufträge je Rolle** (Planner, beim Schnitt vom 2026-10-30, nachgezogen beim Schnitt vom 2026-11-07; Nummern nach der Tabelle vom 2026-11-07):

- *Implementer:* L11, L17, L18, L19 und L20 sind nebenläufig; die Zusage „unter einer Sperre“ gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird: der Bericht nennt die Rate je Mutation mit Zahl der Prozesse (50 Prozesse mit `-test.count=1` unter `make test-race`, dazu `-count=50` für die Wirkung) und das Messende, und der Implementer trägt sie mit der Tabelle in einen Absatz von §3 ein (`AGENTS.md` §3.10). Die Haltepunkt-Fänger (L11, L17, L18, L19, L20) sind deterministisch in Richtung der Mutation (ein zweites `Write`, eine Rückkehr trotz Blockade, ein Hängen innerhalb 2 s); die Masse-Fänger laufen daneben. Die Canary-Schleife L14 läuft als Schleife über 15 Stellen × 4 Einbringungswege × 2 Ausgaben (120 Prüfungen), nicht als Liste einzelner Fälle. `FakeSenke` und jeder Test-Writer halten die Regel der Hüllen (Betriebsregeln). Die Zusatzzeit der Stufen `test` und `race` (vorher und nachher, Zahl der Läufe, Zustand des Hosts; die Mengen M1 und M3 sind die teuersten) steht in §7; Leser, Modelle und Listen der Mengen M1 bis M9 liegen so im Paket, dass `slice-telemetry-handler` sie nutzt (§6, Zuschnitt des Zeilenschreibers); die Entscheidung über den Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) liegt beim User, nicht im Slice.
- *Reviewer:* löst jede Zeile der Tabelle und jede Zelle des Rasters mit einer Mutation auf, fährt die nebenläufigen Zeilen selbst mit Rate, prüft den Canary gegen die Klasse (alle Stellen × 4 Wege × alle Ausgaben), den Zuschnitt des Zeilenschreibers gegen die Nutzung durch `Handle` und die Span-Zeile (§6) und die Hülle `FakeSenke` auf verdeckte Eigenschaften, löst die Zweig-Menge (§6) gegen den Bericht der Abdeckung auf; eine im Code entschiedene Randform ist ein MEDIUM (`AGENTS.md` §3.16).
- *Verifier:* prüft die Tabelle gegen `SPEC-231` und `SPEC-232`, dass dieser Slice nichts aus `slice-telemetry-port`, `slice-telemetry-handler` oder `slice-telemetry-spans` mitnimmt, dass das Binary unverändert blieb, und die Messpunkt-Zeile (Handoff und Ende, `SPEC-128` Nr. 9).
- *Planner (Closure):* setzt die Messpunkt-Zeile, die Auslastung und meldet dem Architect, liegt sie über 1,00 oder unter 0,60.

**Rollenfolge** (`.claude/agents/`, kein Self-Review): Architect (Vertrag, Tabelle, Raster, Block `groessenmarke`) → Planner (Schnitt, Marke beim Start, §4) → Implementer (Code, Mutationen selbst rot, Absatz zu Zusagen und Raten in §3, ein Handoff-Commit) → Reviewer (Report unter `docs/reviews/`) → Fix-Runden (Menge je Befund, `AGENTS.md` §3.17) → Verifier → Planner (Closure mit Lerneintrag; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorlage an den User) → Integrationsschritt (Merge, Register, Roadmap, Stempel, `make gates` im Hauptbaum).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): `slice-telemetry-port` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand beim Start** (Planner, 2026-11-07): erfüllt. `slice-telemetry-port` liegt in `done/` (Teil `P`), die Handbuch-Slices `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` ebenfalls (die Bedingung greift nur bei sichtbarer Änderung am Binary, nicht erwartet). Der Slice berührt keine Datei der Gate-Kette (Übergangsregeln Nr. 7: weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` noch den Stempel; in `in-progress/` liegen `slice-audit-verdrahtung` (Linie D, nur Tests unter `src/bootstrap/`) und `slice-mock-zuordnung` (Linie B), beide keine Gate-Kette). Er berührt keine Datei der Zweige: weder `src/adapters/driven/odata/**` noch `src/hexagon/services/mock/` noch `src/hexagon/ports/inbound/mock.go` (`slice-mock-zuordnung`) noch `src/bootstrap/` (`slice-audit-verdrahtung`: `audit_betrieb_test.go`, `audit_betrieb_hilfen_test.go`; `slice-telemetry-log` ändert `src/bootstrap` nicht). Gemeinsam bleiben nur `spec/spezifikation.md` und diese Pläne (Aufträge 6 in §3). Ändert Vertrag oder Code eine Datei der Gate-Kette, `src/bootstrap/` oder der Linie B (auch `.a-check.yml`, `.golangci.yml`), geht das als Rückgabe an den Planner (§6). Zweig `slice-telemetry-log`, Worktree `/Development/odata-recorder-wt-telemetry-log`.

**Abhängigkeiten — Grund je Eintrag:**

- `slice-telemetry-port`: *Code-Abhängigkeit*. Der Slice erweitert den Adapter aus `src/adapters/driven/telemetry`: `New` und `Config`, die Regeln der Attribute (Schlüssel, Wertformen, Verwurf), `Dropped` samt Grund `senke`, `CountError` und `Snapshot`, dazu die Typen in `src/hexagon/ports/support`. Der `Log` des ersten Teils prüft und zählt, schreibt aber keine Zeile; ohne diesen Code gibt es nichts, was die Zeile aufnimmt.
- Handbuch-Punkt der DoD (nur wenn er greift): `slice-handbuch-geruest` und `slice-handbuch-meldungskatalog` aus `welle-benutzerhandbuch` liegen dann in `done/`; ohne sichtbare Änderung am Binary entfällt die Bedingung.

**Trigger der Welle:** Der Start-Trigger von [welle-betrieb](../welle-betrieb.md) §2 (`welle-mock` in `done/`, M3 erreicht) ist für diesen Slice wie für `slice-telemetry-port` gelockert (Entscheidung des Planners nach Auftrag des Users vom 2026-10-30, dort §4): der Slice braucht keinen Code von Mock oder Workspace und ändert `src/bootstrap` nicht. Die Lockerung gilt nur, solange der Slice ohne Wirkung am Binary bleibt; sonst wartet er auf die beiden Handbuch-Slices (Rückführung `in-progress` → `open`, unten).

**Größenmarke** (Planner, 2026-11-07, nach dem Schnitt in §1; `basis`, `schaetzung` und `marke` stehen im Block, gesetzt in einem eigenen Commit vor dem ersten Code-Commit; `basis` ist der Schnitt-Commit, der Vertrag, Tabelle und Raster dieses Teils abschließt). Die Rechnung nach `SPEC-128` Nr. 1 (Schätzung *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke *M* = 100·⌈135·*S*/10000⌉; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen):

| Plan | *a* | *b* | *f* | *p* | *S* | *M* | Schnittfrage |
|---|---|---|---|---|---|---|---|
| ungeschnitten (L01 bis L20, H01 bis H11; Stand vor dem Schnitt) | 7 | 24 | 1 | 380 | 3095 | 4200 | ja (*S* ≥ 2223) |
| Teil `L` (L01 bis L20, dieser Plan) | 5 | 15 | 1 | 260 | 2085 | 2900 | nein (138 unter der Schwelle) |
| Teil `H` (H01 bis H11, `slice-telemetry-handler`) | 2 | 9 | 0 | 120 | 1010 | 1400 | nein |

*S* und *M* der Teile liegen beide unter der Schnittgrenze; die Summe der Schätzungen (3095) gleich dem ungeschnittenen Plan, die Produktcodes (260 und 120) zusammen 380. Der Architect nannte keinen `schnitt-grund`, weil sich der Plan vor dem Code ohne geteilten Mechanismus schneiden lässt (Zeilenschreiber, Senke und Sperre in `L`, die Abbildung von `slog` darüber in `H`); der Planner hat geschnitten (2026-11-07), also steht kein `schnitt-grund` im Block. Dieser Plan behält die Zeilen `L01` bis `L20` (Produktcode 260, Fakes 1), der Plan von Teil `H` die Zeilen `H01` bis `H11` (Produktcode 120, Fakes 0; dort steht die Marke in Prosa, der Block folgt beim Start). Pfad: allein der Adapter, `src/adapters/driven/telemetry`; der Slice ändert `src/hexagon/ports/support` nicht, und eine Änderung dort wäre eine Abweichung vom Plan (§6), die den Pfad in den Block nimmt. `fakes` zählt `FakeSenke` (1); `FakeTelemetry` liefert `slice-telemetry-port`. Die Zeilen der Tests aus `slice-telemetry-port`, die der Auftrag (1) in §3 ändert, fallen in die Messung (`SPEC-128` Nr. 4 g, Nr. 10 g) und sind im Faktor 80 enthalten, weil es wenige Zeilen sind.

```groessenmarke
pfad: src/adapters/driven/telemetry
produktcode: 260
fakes: 1
basis: 47b97169f594779f8c6e793edd5ba6e6124226c2
schaetzung: 2085
marke: 2900
```

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

- Auslegung des Planners (`Log` in P ohne Zeile): bestätigt vom Architect; `MinLevel` wendet erst dieser Slice an (L04) — **Ausgang:** entschieden; Ort: §3 von `slice-telemetry-port` und `SPEC-232` Nr. 1; die Lesart der Verwürfe unter `MinLevel` steht in `SPEC-231` Nr. 2.
- Sperre über `Clock.Now()` (`SPEC-232` Nr. 5 f), Seite von `Log` — **Ausgang:** entschieden; Fänger L19 und L20 (die Seite von `Handle` fängt H11 in `slice-telemetry-handler`).
- Der Span-Teil nutzt Zeilenschreiber und Sperre dieses Slice für die Span-Zeile (`SPEC-231` Nr. 6), der Handler für `Handle` (`SPEC-231` Nr. 5); ein Zuschnitt des Schreibers, der die Zeile eines Spans oder die Attribute eines Handlers nicht aufnimmt, träfe `slice-telemetry-spans` und `slice-telemetry-handler` — **Ausgang:** weiter offen; Adresse: Implementer dieses Slice (der Schreiber nimmt eine Zeile aus Feldern entgegen, nicht nur die Attribute von `Log`, und stellt Sperre, Uhr und Zählung von `senke` so bereit, dass `Handle` und `End` sie ohne zweite Fassung nutzen; der Reviewer prüft es), Folge-Slices `slice-telemetry-spans` und `slice-telemetry-handler`. Ist-Zustand des Zuschnitts (Implementer): `zeileSchreiben(t, level, msg, felder)` in `zeile.go` nimmt Zeit, Stufe, Text und die Felder in der Reihenfolge der Ausgabe (`support.Attr` der drei Arten) entgegen, bildet die Zeile, schickt sie in einem `Write` unter `zeilenMu` hinaus und zählt `senke`; `eintrag(level, msg, attrs)` ist der Weg von `Log` (Filter, Verwürfe, `MinLevel`, `Now()` vor der Sperre, Reihenfolge der 13 Felder). `Handle` ruft nach der Abbildung der Attribute `eintrag`, die Span-Zeile ruft `zeileSchreiben` mit ihren Feldern (`ereignis`, `span_id`, `parent_id`, `duration_ms`, Attribute) in der Reihenfolge von `SPEC-231` Nr. 6; die Hilfen der Tests (`FakeSenke`, `RueckrufUhr`, `leseZeile`, `modellMsg`, `modellStufe`, `logSchluessel`, `gueltigFuer`, `andererWert`, `mengeZeichen`) liegen in `hilfen_test.go` und `zeile_test.go` des Pakets. Die Aussage, dass `Handle` und die Span-Zeile ohne zweite Fassung auskommen, bleibt bis zu deren Slices ungeprüft.
- Akzeptiertes Negativ L15 (wohlgeformter Wert in zulässigem Schlüssel, Inhalt von `msg`): Folge ist eine Ausgabe, wo eine Stelle einen Request-Wert dort ablegt; der Schutz liegt im Fänger je Stelle des Einbaus — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` (Fänger je Stelle in deren §1 und DoD, Eintrag des Planners vom 2026-11-07).
- Akzeptiertes Negativ L16 (Panik von `Writer` oder `Clock`, nie zurückkehrende Senke, Senke ruft den Adapter, Jahr außerhalb 0000 bis 9999): Grund und Folge in der Tabelle und in `SPEC-232` *Ungeprüft* — **Ausgang:** weiter offen; Adresse: User (Folge im Betrieb, wenn der Log-Treiber des Containers blockiert, und ein Zeitstempel eines fernen Jahres hat ein anderes Format; kein Fänger, solange kein Betrieb diesen Fall zeigt; der Planner führt den Punkt in den Hinweisen für den User in der Roadmap und in der Closure-Notiz, bis der User entschieden hat).
- Verdrahtung des Adapters und Umstellung der Dienste `record`, `session`, `maintenance` auf `slog.New(Handler())` (`SPEC-231` Nr. 7): `slice-telemetry-einbau-gateway` hat die Sendung angenommen (Eintrag des Planners vom 2026-11-07 in dessen §1 und DoD; es nennt `session` und `maintenance` jetzt) — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway`.
- `SPEC-121` Nr. 12: der Satz „einen Port `Telemetry` gibt es noch nicht“ wird mit dem Einbau veraltet; der Wortlaut gehört dem Einbau-Slice — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway` (§1 dort, Sendung (b)).
- Handbuch: die Logfelder werden am Binary sichtbar, sobald der Einbau den Adapter verdrahtet; die Handbuch-Pflicht greift dort, nicht hier — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway` (Abhängigkeiten `slice-handbuch-geruest`, `slice-handbuch-meldungskatalog`).
- Eine Senke, die nie zurückkehrt, hält den Aufrufer (`SPEC-232` Ungeprüft, Entscheidung 6 und 17 des Architect: synchrones Schreiben) — **Ausgang:** weiter offen; Adresse: User (siehe L16).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten ist gefüllt, jede Zelle trägt ihren Ort, die Lücken (Maskierung, Reihenfolge von Ersetzen und Kürzen, Zahlenformat, Wert gegen Name bei `MinLevel`, Rangfolge im Handler, Fehlerformen der Senke, Sperre der Zeile) sind am 2026-11-07 in `SPEC-231` und `SPEC-232` entschieden — **Ausgang:** entschieden; Ort: `SPEC-231` Nr. 1 bis 3 und 5, `SPEC-232` Nr. 3, 5 d, 5 g; eine Zelle ohne Ort, die der Implementer findet, geht an den Architect und steht dann hier.
- Die Größenmarke ist die Schätzung nach `SPEC-128`; `AGENTS.md` §3.14 trägt noch die ältere Formel (Faktoren 85 und 40, Zuschlag 1,25) — **Ausgang:** weiter offen; Adresse: User (Diff-Vorlage im Closure von `slice-telemetry-port`, §6 dort).
- Tabelle, Raster, Pflicht-Mengen mit Achsen, Zweig-Menge, Zeile für den Ersatz von `TestKernSchreibtNichts` (Aufträge 1 bis 4 in §3) — **Ausgang:** entschieden; Ort: §3 (Tabelle L01 bis L20, Mengen M1 bis M10 und der Anteil `L` von M16, Zählprüfung der Zweige als Abdeckung) und dieser Abschnitt (Raster, Fehlerformen, Fenster, Zweig-Menge Z01 bis Z37). Der Teil des Handlers (H01 bis H11, M11 bis M15, Z38 bis Z65) steht in `slice-telemetry-handler`. Der Implementer liefert die Zählprüfungen gegen die Quellen und den Bericht der Abdeckung (`AGENTS.md` §3.17).
- Schnittfrage und Marke (Auftrag 5 in §3, §4): *S* des ungeschnittenen Plans 3095, über der Schnittgrenze 2223; der Architect schlug den Schnitt in `L` (*S* 2085) und `H` (*S* 1010) vor — **Ausgang:** entschieden (Planner, 2026-11-07): geschnitten, Datei `slice-telemetry-handler` in `open/` mit den Zeilen `H01` bis `H11`, den Mengen M11 bis M15 und dem Anteil `H` von M16, den Zweigen Z38 bis Z65, der Entscheidung 14 und der Hälfte `H` der Sendung; `slice-telemetry-spans` hängt nur an diesem Slice, die Einbau-Slices an diesem, an `slice-telemetry-handler` und an `slice-telemetry-spans`; Marke dieses Teils im Block gesetzt (2085, 2900), die von `slice-telemetry-handler` steht dort in Prosa (1010, 1400); Roadmap und `welle-betrieb` nachgezogen.
- Zählprüfungen der Tests aus `slice-telemetry-port` mit festen Zahlen (37 Codes, 7 Gründe, 5 Namen; Liste `alleCodes` in `regeln_test.go`): `slice-mock-lesen` (Teil 2 des Schnitts von `slice-mock-zuordnung`, Linie B) fügt `ODR-E6008` hinzu und macht daraus 38 Codes; die Tests von `slice-telemetry-port` werden nach dem Merge rot, solange die Zahl fest steht — **Ausgang:** entschieden (Planner, 2026-11-07, Rückgabe 2 des Architect): der Implementer dieses Slice zieht die festen Zahlen an die Quellen (Auftrag (1) in §3, `model.AlleCodes()`, `support.Verwuerfe()`, `support.Spans()`); den Text `ODR-E6008` in die Orakel-Liste `alleCodes` trägt `slice-mock-lesen` ein (Sendung in dessen §1; seine Tabelle in §3 nennt `regeln_test.go` und jede Stelle mit 37 Codes schon), und hält die Meldung „Orakel-Liste nachziehen“ die Lücke sichtbar. Merge-Reihenfolge: wer als zweiter nach `main` geht, zieht `main` nach und lässt `make gates` laufen; geht `slice-mock-lesen` vor diesem Slice, ändert es die Zahl 37 auf 38 (seine Zeile), und dieser Slice ersetzt sie durch die Quelle. In den Plan eines anderen Zweigs schreibt dieser Zweig nichts.
- Kennungen und Katalog (Aufträge 6 und 7 in §3): keine neue `SPEC`-Nummer, kein Fehlercode, Schlüssel oder Kommando; der Zusammenstoß in `spezifikation.md` mit `slice-audit-verdrahtung` und `slice-mock-zuordnung` liegt in der Historie (§7 der Spezifikation, jede Zeile oben angefügt) und in der Tabelle der Metriken (37 werden 38, dort nicht von diesem Slice berührt) — **Ausgang:** entschieden (Rückgabe 3 des Architect, Planner, 2026-11-07): die Auflösung des Zusammenstoßes beim Merge ist der Integrationsschritt; er behält beide oder alle Zeilen der Historie in `spec/spezifikation.md` §7 (jede Zeile oben angefügt, nach Datum und Merge-Reihenfolge), löst nur den Konflikt der Historie und der Metriken-Tabelle auf und ändert keinen Wortlaut von `SPEC-230` bis `SPEC-232`; wer als zweiter nach `main` geht, löst ihn.
- Gate-Kette, `src/bootstrap/` und Linie B (Stand beim Start in §4): ein Bedarf an einer Änderung dort — **Ausgang:** weiter offen (nicht zu erwarten); Adresse: Planner.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Dieser Teil trägt die Zeilen `Log`, `Senke` und `Uhr` (Seite `Log`); die Zeile `Handler` und die Seite `Handle` der Uhr stehen in `slice-telemetry-handler`, `CountError` und `Snapshot` in `slice-telemetry-port`, die Zeilen der Spans in `slice-telemetry-spans`. Die Nummern stehen in spezifikation.md §5.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | sensibler Wert | Abbruch (`ctx`) | Wiederholung |
|---|---|---|---|---|---|---|
| Log schreiben (`Log`) | `SPEC-231` Nr. 1 (nicht gesetztes Feld fehlt; Pflicht je Ereignis liegt bei der Stelle) | `SPEC-230` Nr. 3 (leerer Wert ist Formverstoß, `wert`); `SPEC-231` Nr. 3 (leerer `msg` zulässig) | `SPEC-230` Nr. 2 und 4 (Schlüssel außerhalb: `schluessel`); `SPEC-231` Nr. 2 (Level außerhalb der vier Namen; Vergleich mit `MinLevel` nach Wert); `SPEC-231` Nr. 3 (ungültiges UTF-8 in `msg`) | `SPEC-232` Nr. 2 | `SPEC-230` Nr. 1 (wirkt auch bei beendetem `ctx`) | `SPEC-231` Nr. 1 (jeder Aufruf eine Zeile); `SPEC-230` Nr. 4 (doppelter Schlüssel im Aufruf) |
| Senke schreibt (`Write` des Adapters) | gilt nicht: der Adapter reicht die Zeile, kein Argument fehlt | gilt nicht: eine Zeile ist nie leer, sie trägt mindestens `timestamp` und `level` (`SPEC-231` Nr. 1) | `SPEC-232` Nr. 3 (Fehler, weniger Bytes, Fehler mit allen Bytes, mehr Bytes ohne Fehler: Tabelle der Fehlerformen unten); Panik der Senke: `SPEC-232` *Ungeprüft* | `SPEC-232` Nr. 2 (nie ein Fehlertext der Senke) | gilt nicht: kein `ctx` an der Senke (`SPEC-230` Nr. 1); eine Senke, die blockiert: `SPEC-232` Nr. 5 g (Zähler und `Snapshot` warten nicht), nie zurückkehrend: *Ungeprüft* | `SPEC-232` Nr. 3 (kein Wiederholen, die nächste Zeile wird geschrieben); `SPEC-232` Nr. 5 c (ein `Write` je Zeile, nie zwei zugleich) |
| Zeit der Uhr (`Clock.Now()` in `Log`; `Handle` in `slice-telemetry-handler`) | gilt nicht: die Uhr liefert immer eine Zeit; die Nullzeit ist eine Zeit (`SPEC-231` Nr. 2) | gilt nicht: wie fehlender Wert | `SPEC-231` Nr. 2 (andere Zone: UTC; Jahr außerhalb 0000 bis 9999: nicht zugesagt) | gilt nicht: die Zeit trägt keinen Inhalt eines Requests | gilt nicht: kein `ctx` an der Uhr; eine Uhr, die nie zurückkehrt: `SPEC-232` *Ungeprüft* | `SPEC-231` Nr. 2 (einmal je geschriebener Zeile, nie für einen Eintrag unter `MinLevel`); `SPEC-232` Nr. 5 f (Rückruf in `Now()` blockiert nicht) |

**Fehlerformen der Senke** (Zeile „Senke schreibt“; `SPEC-232` Nr. 3; jede Form hat einen Fänger oder ein Negativ mit Grund und Folge):

| Form der Rückgabe von `Write` | Wirkung | Fänger oder Negativ |
|---|---|---|
| Fehler, 0 Bytes | Zeile verloren, `senke` 1, kein Wiederholen | L10 (Menge M9) |
| Fehler, weniger Bytes als die Zeile | Zeile verloren, `senke` 1, Rest nicht nachgeschrieben | L10 |
| Fehler, alle Bytes | Zeile verloren, `senke` 1 (der Fehler zählt, nicht die Menge der Bytes) | L10 |
| kein Fehler, weniger Bytes als die Zeile (auch 0) | Zeile verloren, `senke` 1 | L10 |
| kein Fehler, mehr Bytes als die Zeile | kein Fehler, `senke` 0 (Verstoß der Senke gegen `io.Writer`) | L10 |
| kein Fehler, alle Bytes | Zeile geschrieben, `senke` 0 | L10, L12 |
| blockierend (kehrt später zurück) | Zeile geschrieben, andere Schreiber warten; `CountError`, `Snapshot` und Zählung der Verwürfe warten nicht | L11 (Haltepunkt am `Writer`), L17 (zweiter Schreiber), L18 (Aufruf kehrt erst nach `Write` zurück) |
| blockierend (kehrt nie zurück) | der Aufrufer wartet mit, mit ihm jeder Schreiber | Negativ L16; Grund: die Senke des Binary blockiert erst bei vollem Puffer des Log-Treibers; Folge: Schreiber hängen, Zähler und `Snapshot` nicht |
| Panik oder Aufruf des Adapters aus `Write` | nicht zugesagt | Negativ L16; Grund: die Ports paniken nie, die Senke des Binary ruft den Adapter nicht; Folge: ein Schreiber stürzt oder alle blockieren |

**Fenster der Sperre** (`AGENTS.md` §3.10, Zusage „unter einer Sperre, ohne sie dazwischen freizugeben“, `SPEC-232` Nr. 5 c, d, f, g; je Fenster Schritte, Seite und Fänger oder benanntes Negativ; die Seite `Handle` und die abgeleiteten Handler stehen in `slice-telemetry-handler` §6):

| Fenster | Seite | Fänger oder Negativ |
|---|---|---|
| Zeile bilden und `Now()` rufen, vor der Sperre der Zeile | `Log` | L20 (Rückruf `Log`: ein `Now()` unter der Sperre der Zeile blockiert); L19 (Rückruf `CountError`, `Snapshot`: ein `Now()` unter der Sperre des Zustands blockiert) |
| zwischen Nehmen der Sperre der Zeile und `Write` | `Log` | Negativ ohne Fenster: dazwischen steht kein Schritt, den ein Test erreicht; ein Aufruf eines übergebenen Arguments liegt nicht dazwischen; die Wirkung („zwei `Write` zugleich“) fängt der Haltepunkt am `Writer` in L17, in jedem Lauf |
| `Write` selbst (ein Aufruf mit der ganzen Zeile) | `Log` | L12 (ein `Write` je Zeile), L17 (nie zwei zugleich), L18 (kehrt erst nach `Write` zurück) |
| zwischen `Write` und Freigabe, Zählung von `senke` | `Log` | L10 (nach der Rückkehr des Aufrufs steht `senke` im Stand); ob die Zählung vor oder nach der Freigabe liegt, ist nicht zugesagt, Negativ; Grund: die Zählung nimmt nur die Sperre des Zustands, nie die der Zeile (`SPEC-232` Nr. 5 g); Folge: keine beobachtbare |
| zwischen gleichzeitigen Aufrufen (Reihenfolge der Wartenden) | `Log` | Negativ; Grund: zugesagt ist die Reihenfolge je Goroutine, nicht zwischen Goroutinen (`SPEC-232` Nr. 5 d); Folge: keine |
| Sperre der Zeile während `Write` blockiert | `CountError`, `Snapshot`, Zählung | L11 (Haltepunkt am `Writer`, deterministisch) |

**Zweig-Menge** (Modell-Menge der Wege von `Log` und der Senke; die Zweige Z38 bis Z65 des Handlers stehen in `slice-telemetry-handler`; `AGENTS.md` §3.15; Register `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`; die Zählprüfung ist die Abdeckung unter dem Detektor, §3). Ein Zweig ohne Treffer in `go test -race -coverprofile` ist ein benanntes Negativ mit Grund und Folge im Bericht.

| Zweig | Fänger |
|---|---|
| Z01 Eintrag unter `MinLevel`: keine Zeile, `Now()` nicht gerufen, Verwürfe zählen | L04 |
| Z02 Eintrag ab `MinLevel`: eine Zeile | L01, L04 |
| Z03 bis Z06 Stufe `DEBUG`, `INFO`, `WARN`, `ERROR`, an drei Grenzen | L03 |
| Z07 `msg` leer; Z08 `msg` bis 256 Bytes unverändert | L05 |
| Z09 bis Z12 Kürzung bei einem Zeichen der Breite 1, 2, 3, 4 an der Grenze | L05 |
| Z13 ungültiges Byte wird zu U+FFFD (je Byte, auch in der Kürzung) | L05, L06 |
| Z14 Maskierung von `"` und Backslash; Z15 Zeichen unter U+0020; Z16 U+2028 und U+2029; Z17 unmaskierte Zeichen (`<`, `>`, `&`, DEL, übrige) | L06 |
| Z18 Feld nicht gesetzt (fehlt); Z19 Feld der Art `String`; Z20 Feld der Art `Float` (`duration_ms`); Z21 Feld der Art `Int` (`befunde`) | L01, L07 |
| Z22 doppelter Schlüssel: der erste gilt | L08 |
| Z23 Verwurf `schluessel`; Z24 Verwurf `wert` (Regeln aus `slice-telemetry-port`, hier über die Zeile) | L04, L08, L09 |
| Z25 Zeit in anderer Zone wird UTC; Z26 Zeit in UTC; Z27 Nullzeit | L02 |
| Z28 `Write` ohne Fehler, alle Bytes; Z29 mehr Bytes ohne Fehler | L10, L12 |
| Z30 Fehler, 0 Bytes; Z31 Fehler, weniger Bytes; Z32 Fehler, alle Bytes; Z33 kein Fehler, weniger Bytes | L10 |
| Z34 Zählung `senke` einmal je Zeile | L10 |
| Z35 Senke blockiert, Sperre der Zeile gehalten, Zähler frei | L11, L17 |
| Z36 Rückruf in `Now()` (Zustand, Zeile) | L19, L20 |
| Z37 `Log` ohne Argumente (nil-`ctx`, nil-Slice) | L09 |

**Zusatztabelle: Verstöße am selben Schlüssel** (`AGENTS.md` §3.16; `SPEC-230` Nr. 4, Rangfolge Schlüssel vor Wert, genau ein Verwurf je Attribut): die Rangfolge und die Zählung fängt `TestVerwurfRangfolge` (P17) in `slice-telemetry-port`; dieser Slice fängt den Fall „doppelter Schlüssel, beide gültig und verschieden“ in der Zeile für alle 13 Schlüssel in jeder Position (L08, Menge M7); über den Handler mit `WithAttrs` (H05) und die Rangfolge Schlüssel vor Art im Handler (H03, Menge M12) fängt ihn `slice-telemetry-handler`.

**Zusatztabelle: Modus** (`AGENTS.md` §3.16; die Zusage hängt am Modus):

| Modus | Zusage | Fänger |
|---|---|---|
| `MinLevel` über `INFO` (jeder Wert von −8 bis 12) | keine Zeile unter dem Minimum, `Now()` nicht gerufen, Verwürfe zählen | L04 |
| Senke defekt (sechs Formen) | keine Zeile, `senke` 1 außer bei mehr Bytes ohne Fehler und bei heiler Senke, die nächste wird geschrieben (`SPEC-232` Nr. 3) | L10 |
| Senke blockiert | Zähler und `Snapshot` warten nicht (`SPEC-232` Nr. 5 g) | L11 |

Die Seite `Handle` dieser Modi (H07, H08), der Modus „Weg `Log` oder `Handler`“ (H04) und der Modus „Handler abgeleitet“ (H10, H05) stehen in `slice-telemetry-handler`. Die Modi „`MinLevel` bis `INFO`“ und „Senke defekt“ für die Span-Zeile fängt S13 in `slice-telemetry-spans`.

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

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: die Messpunkt-Zeile von `make groessenmarke` (Handoff und Ende, von Hand gemessener Produktcode, Auslastung), die Raten der Wirkungs-Fänger L11, L17, L18, L19 und L20 mit Zahl der Prozesse und Messendem, die Zusatzzeit der Stufen `test` und `race`, den Bericht der Abdeckung (Zweig-Menge), und dass das Binary unverändert blieb.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-07; Zähler = Zahl der Dateien in `evidence/`). Offen und einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (7×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure der Welle oder des nächsten Slice mit Nebenläufigkeit — dieser Slice hat fünf nebenläufige Zeilen, `slice-telemetry-handler` zwei) gilt für die Zweig-Menge (Auftrag 4) und die Fehlerformen der Senke; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“ (dieser Zweig setzt keinen Zähler); `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (15×) für die Marke. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×) für `FakeSenke`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (8×) für L17 und L18, `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×, `AGENTS.md` §3.10), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.16) für die Allaussage „kein Credential in keinem Log“, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14) für die Mengen der Fänger, `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (28×, `AGENTS.md` §3.17), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, beide `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16). Mit diesem Slice erreicht kein offener Eintrag ohne Ausgang 3×, außer den genannten.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF.
