# Slice slice-lasttest-leser-zeitgrenze: Der Fänger des Lasttest-Lesers hängt nicht an einer Echtzeitgrenze, die unter Last reißt

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung über die DoD hinaus. Er ändert einen Test unter `tests/lasttest/`, nicht die Gate-Kette (Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe), `LH-QA-01` (der Test gehört zum Lasttest).

**Berührte Spec-Stellen:** [SPEC-216](../../../../harness/vertraege/SPEC-216-lasttest.md) Nr. 9 e (das Programm entleert die Ströme des Binary ohne Unterbrechung und liest nie langsamer, als das Binary schreibt) und Nr. 10 a (Zeilen bis 1 MiB, letzte Zeile, Fehler des Stroms; nennt `TestLeserHaeltNieAn` als Fänger der zwei letzten), dazu Nr. 14 a (der Test liest die Uhr allein über `systemuhr` aus `uhr.go`) und Nr. 16 s (benanntes Negativ des Fängers, vom Architect eingetragen). Die Zusagen bleiben; dieser Slice streicht am Vertrag allein den Stand-Vermerk von Nr. 16 s („nicht geliefert“).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-lasttest-leser-zeitgrenze`, Worktree `/Development/odata-recorder-wt-leser`; der Zweig entsteht von `main` nach dem Start-Commit, der Start liegt auf `main`)

**Autor:** pt9912 (Planner, Closure von `slice-replay-speicher-und-suche`). **Datum:** 2026-11-27.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice.

**Übernimmt:** aus `slice-replay-speicher-und-suche` (Plan §6, Verifikation V-6): `make test-race` war in einem Lauf von `make gates` rot an `TestLeserHaeltNieAn` (`tests/lasttest/leser_test.go`, „nach 1s kein Wert“) bei hoher Last des Hosts.

**Ziel:** `TestLeserHaeltNieAn` fängt weiter, dass der Leser den Schreiber nie anhält, und wird auf einem belasteten Host nicht allein deshalb rot, weil 20 MiB unter dem Detektor nicht binnen 1 s Echtzeit durch die Pipe gehen (`pipeLauf`, `warteAuf(…, time.Second)`).

**Form der Lösung:** `pipeLauf` läuft in einer Blase von `testing/synctest`; an die Stelle der Grenze von 1 s Echtzeit tritt, dass die Uhr der Blase beim Lesen und Auswerten stehen bleibt, dazu, dass kein Write des Schreibers scheitert. „Den Schreiber anhalten“ fängt der Test damit als Warten auf die Uhr der Blase und als Ende des Lesens vor dem Strom, nicht als Verlangsamung, während diese Uhr steht (Negativ in §3 und `SPEC-216` Nr. 16 s).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung am Leser oder an den Zusagen von `SPEC-216` Nr. 9 e und 10 a — der Befund betrifft den Fänger, nicht das Werkzeug; ein Lauf des Lasttests ist nicht nötig.
- Die weiteren lastabhängigen Tests in `urlparser`, `admin` und `httpodata` (`BEO-REPO/test-fremden-pakets-flackert-unter-last`) — andere Pakete, Ursache nicht untersucht; Adresse Planner beim Lese-Schritt des Registers.
- Eine Lockerung eines Gates (`make test-race` bleibt, wie es ist; `AGENTS.md` §3.6).

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [x] Ursache belegt: die Dauer von `pipeLauf` mit 20 MiB unter dem Detektor steht mit Last des Hosts und Messendem im Plan.
- [x] Der Fänger hängt nicht mehr an der Grenze von 1 s, oder die Grenze ist aus der Messung begründet; die Mutationen der Zeile A1 des Plans von `slice-lasttest-auswertung` (Scanner mit 64 KiB, Pause beim Auswerten, Grenze entfällt, Sammler ohne Sperre) bleiben je rot, mit Rate (`AGENTS.md` §3.10, §3.15).
- [x] `make gates` grün (Stand `2170d89a`; die Grenze dieses Belegs steht in §7 *Gate*).
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tests/lasttest/leser_test.go` | update | `pipeLauf` (Blase von `testing/synctest`, Uhr der Blase statt Grenze von 1 s), `schreibeStrom` (Schreiber mit Fehler des Write) und der Kopftext von `TestLeserHaeltNieAn` |
| `harness/vertraege/SPEC-216-lasttest.md` | update | allein der Stand-Vermerk von Nr. 16 s entfällt mit der Lieferung; nicht im Block, weil `SPEC-128` Nr. 3 m und 4 a Vertragstext ausschließen |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Planner beim Start, 2026-11-27; kein Architect-Lauf, weil der Slice keinen neuen Vertrag liefert). Die Zeilen folgen der Zeile A1 des Plans von [`slice-lasttest-auswertung`](../done/slice-lasttest-auswertung.md) und der DoD; die Spalten sind die von `SPEC-128` Nr. 3 a. Der Implementer trägt Mutationen und Raten mit Zahl der Prozesse, Last und Messendem in einen Absatz darunter ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| LZ1 | (Nr. 9 e) Ein Leser, der beim Auswerten pausiert, macht den Fänger rot | `TestLeserHaeltNieAn` (`pipeLauf`) | 1 ms Pause je Span-Zeile in `nimm` (`a1-pause-beim-auswerten`); der Wert der Mutation bleibt der aus `slice-lasttest-auswertung` | nebenläufig |
| LZ2 | (Nr. 9 e) Ein Lesen, das endet, solange der Schreiber noch schreibt, macht den Fänger rot | `TestLeserHaeltNieAn` (`pipeLauf`) | `bufio.Scanner` mit der Grenze von 64 KiB (`a1-scanner64`) | nebenläufig |
| LZ3 | Am unveränderten Leser wird der Fänger auf einem belasteten Host nicht allein an der Zeit rot | Messung, kein deterministischer Test: Prozesse von `TestLeserHaeltNieAn` unter dem Detektor bei belastetem Host, mit Zahl, Last und Messendem, dazu die Dauer von `pipeLauf` (DoD 1) | die Grenze von 1 s bleibt (Stand vor dem Slice); die Rate hängt an der Last | nebenläufig |
| LZ4 | (Nr. 10 a) Eine Zeile über 1 MiB wird verworfen und gezählt, der Speicher wächst nicht mit der Länge | `TestLeserHaeltNieAn` (1 MiB + 1 B, 64 MiB in einer Zeile) | die Grenze von 1 MiB entfällt (`a1-grenze-entfaellt`) | andere |
| LZ5 | Der Sammler, in den der Leser schreibt, ist unter Nebenläufigkeit sicher | `TestSammlerGleichzeitig` unter `make test-race` (Bestand, `auswertung_test.go`) | Sperre fehlt in `nimm` (`a1-sammler-nimm-ohne-sperre`) | nebenläufig |
| LZ6 | (Nr. 10 a) Eine letzte Zeile ohne Zeilenende ist eine Zeile, ein Fehler des Stroms geht an den Aufrufer | `TestLeserHaeltNieAn` | die letzte Zeile ohne Zeilenende wird verworfen; der Fehler des Stroms wird verschluckt | andere |

**Messung, Mutationen und Raten** (Implementer, 2026-11-27; `AGENTS.md` §3.10). *Aufbau:* Messender ist der Implementer (ein Agent, Modell Sonnet 5.5), Host mit 20 Kernen; je Reihe ein Container aus dem Image der Stufe `race` (go1.27.1) ohne Netz und Mount mit `--memory=4g --memory-swap=4g --cpus=4`, Test-Binary mit `-race`, je Lauf ein Prozess mit `-test.count=1` (Rate je Prozess gleich Rate je Lauf von `go test`). *Zusatzlast* sind K Endlosschleifen im selben Container; „Last“ ist die 1-Minuten-Last des Hosts, den fremde Prozesse mit belegen (gedrosselte Prozesse zählt sie nur teilweise). *Alt* ist der Stand `3d352b2d`, *neu* der dieses Slice.

*Dauer und Ursache* (DoD 1; je Zelle 10 Prozesse, Sekunden, kleinster bis größter Wert; für diese Messung allein trug der alte Stand 1 min statt 1 s als Grenze):

| Zusatzlast | Last des Hosts | `pipeLauf` 20 MiB alt | `pipeLauf` 20 MiB neu | 19 mal 1 MiB alt | 19 mal 1 MiB neu |
|---|---|---|---|---|---|
| 0 | 3,2 bis 4,4 | 0,29 bis 0,40 | 0,28 bis 0,31 | 0,57 bis 0,64 | 0,58 bis 0,62 |
| 8 | 3,6 bis 7,9 | 0,84 bis 1,19 | 0,81 bis 1,22 | 1,53 bis 2,33 | 1,68 bis 2,40 |
| 16 | 2,5 bis 9,6 | 1,80 bis 2,12 | 1,60 bis 2,69 | 3,61 bis 4,11 | 3,21 bis 4,21 |

*LZ3, unveränderter Leser* (`TestLeserHaeltNieAn`, je Zelle 20 Prozesse, alt und neu abwechselnd):

| Zusatzlast | Last des Hosts vor und nach der Reihe | alt rot | neu rot | Dauer des ganzen Tests neu, s |
|---|---|---|---|---|
| 0 | 3,2 und 3,6 | 0 von 20 | 0 von 20 | 1,35 bis 1,97 |
| 8 | 7,9 und 3,3 | 15 von 20 („nach 1s kein Wert“) | 0 von 20 | 3,90 bis 4,99 |
| 16 | 9,6 und 6,3 | 20 von 20 („nach 1s kein Wert“) | 0 von 20 | 8,91 bis 12,60 |

Es bleibt eine Grenze in Echtzeit: `wache` mit 60 s über den ganzen Test, das 4,8-Fache des größten gemessenen Werts; sie ist auf höherer Last nicht gemessen. Eine angehobene Grenze an der alten Form trennt nicht: die Pause der Zeile LZ1 verlängert den Lauf mit 20 MiB um 5,2 s, der unveränderte Lauf braucht bei 16 Schleifen bis 2,1 s und wächst mit der Last.

*Mutationen der Tabelle* (Last des Hosts 3,8 bis 9,4; neu je 20 Prozesse ohne und mit 8 Schleifen, alt je 20 Prozesse ohne Zusatzlast):

| Zeile | Mutation | roter Test und Meldung (neu) | neu, 0 und 8 | alt |
|---|---|---|---|---|
| LZ1 | `time.Sleep(time.Millisecond)` am Anfang von `nimm` | `TestLeserHaeltNieAn`: „warteten 5.232s auf die Uhr“ | 20 von 20, 20 von 20 | 20 von 20 |
| LZ2 | `lies` über `bufio.Scanner` mit 64 KiB | ebenda: Write „closed pipe“, Leser „token too long“, Spans 0 | 20 von 20, 20 von 20 | 20 von 20 |
| LZ4 | die Grenze von 1 MiB entfällt | ebenda: die Zeile von 1 MiB + 1 B wird gelesen, 383 MB alloziert | 20 von 20, 20 von 20 | 20 von 20 |
| LZ5 | `nimm` ohne Sperre | `TestSammlerGleichzeitig`: `DATA RACE` | 20 von 20, 20 von 20 | Test unverändert |
| LZ6 | letzte Zeile ohne Zeilenende verworfen | `TestLeserHaeltNieAn`: Zeilen `s1 l1 s2` | 20 von 20, 20 von 20 | 20 von 20 |
| LZ6 | Fehler des Stroms verschluckt | ebenda: Fehler `<nil>` statt „Strom gerissen“ | 20 von 20, 20 von 20 | 20 von 20 |

*Weitere Formen derselben Klassen* (Last 2,6 bis 9,1, ohne Zusatzlast; neu 20 Prozesse, alt 5). Achsen der Pause: Stelle (`nimm`, je Zeile in `lies`, hinter der überlangen Zeile), Länge (1 ns, 1 µs, 1 ms), Häufigkeit (je Span-Zeile, je Zeile, einmal), Form (`time.Sleep`, `time.After`). Achsen des frühen Endes: hinter der überlangen Zeile, mitten im Strom; dazu ein Lesen, das nie endet.

| Mutation | neu | alt |
|---|---|---|
| 1 µs je Span-Zeile in `nimm` | 20 von 20 | 0 von 5 |
| einmal 1 ns in `nimm` (beim 101. Gesamtspan) | 20 von 20 | 0 von 5 |
| `<-time.After(time.Millisecond)` in `nimm` | 20 von 20 | 5 von 5 |
| 1 µs je Zeile in `lies` | 20 von 20 | 5 von 5 |
| einmal 1 ms hinter der Zeile über 1 MiB (`haenge`) | 20 von 20 | 0 von 5 |
| `lies` endet mit `nil` hinter der Zeile über 1 MiB | 20 von 20 | 5 von 5 |
| `lies` endet mit `nil` mitten im Strom (`fremd` 50000) | 20 von 20 | 5 von 5 |
| `lies` blockiert mitten im Strom für immer | 20 von 20 (panic von synctest: „deadlock“) | 5 von 5 |

Die Zusicherung über die Uhr trägt LZ1 allein: mit `gewartet < 0` statt `gewartet != 0` und der Mutation LZ1 ist der Test in 5 von 5 Prozessen grün.

**Akzeptiertes Negativ** (`AGENTS.md` §3.15; Wortlaut, Grund, Alternativen und Wirkung im Messlauf in `SPEC-216` Nr. 16 s). *Klasse:* eine Verlangsamung von Leser oder Auswertung, während die Uhr der Blase steht: Rechenzeit und jedes Warten, das `synctest` nicht als dauerhaft blockiert zählt (Systemaufruf, Ein- und Ausgabe, eine Sperre, deren Halter nicht auf diese Uhr wartet, ein Kanal oder Ticker von außerhalb der Blase). *Maß:* 4 von 4 geprüften Formen bleiben grün: `syscall.Nanosleep` mit 1 ms und eine Rechenschleife je Span-Zeile, je 0 von 20 Prozessen rot (Implementer; Dauer des Tests etwa 14 s und 17 s statt 1,4 s, am alten Stand beide in 5 von 5 rot), und 1 ms je Gesamtspan auf einen Ticker von außerhalb der Blase, 0 von 5 rot (Review `2026-11-27-slice-lasttest-leser-zeitgrenze-review`, m3; am alten Stand nicht gefahren), und je Gesamtspan Warten auf eine Sperre, deren Halter 1 ms in `syscall.Nanosleep` steht, 0 von 3 rot bei 13,7 s statt 2,4 s (Verifikation `2026-11-27-slice-lasttest-leser-zeitgrenze-verifikation`; `SPEC-216` Nr. 16 s führt die Sperre noch als nicht gefahren, §7). Ein- und Ausgabe ist nicht gefahren. Die Achsen der Pause in der Tabelle darüber führen allein das Warten auf die Uhr der Blase. *Folge:* ein solcher Leser besteht die Tests, solange die Wache von 60 s hält.

**Ein Hänger endet nicht immer als Fehler des Tests** (Review F-2 und I-1; vom Reviewer gefahren, je 1 oder 3 Prozesse). Ein Senden ohne Empfänger endet in 3 von 3 Läufen als `--- FAIL` mit dem panic „deadlock“ von `synctest`. Eine Warteschleife mit `time.Sleep`, die nie endet (m6), und ein Warten auf eine Sperre, deren Halter 1 ms auf die Uhr der Blase wartet (m4), enden je in 1 von 1 Lauf erst nach 60 s im panic von `wache`: das Test-Binary bricht ab, die Meldung nennt die Frist und nicht den Grund, und die übrigen Tests des Pakets laufen nicht. *Grund:* die Uhr der Blase geht bei der Sperre nicht vor, und die Schleife blockiert nie dauerhaft; eine Grenze je `pipeLauf` in Echtzeit wäre wieder die Grenze, die der Slice entfernt. *Folge:* der Lauf ist rot, der Grund ist aus dem Stapel zu lesen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-replay-speicher-und-suche` liegt in `done/` und ist nach `main` gemergt. Wird ein Lauf von `make gates` vorher an diesem Test rot, hat der Slice Vorrang.

**Stand beim Start** (Planner, 2026-11-27): Die Bedingung ist erfüllt: `slice-replay-speicher-und-suche` liegt in `done/` und ist nach `main` gemergt (`ce6736cb`). In `in-progress/` liegt daneben allein [`slice-bootstrap-m6`](../done/slice-bootstrap-m6.md); `git diff --name-only main...slice-bootstrap-m6` nennt beim Start elf Pfade, keinen unter `tests/lasttest/`. Kein Kettenslice (Nr. 7 der Übergangsregeln). Keine Frage an den Architect ist offen; er wird erst mit der ersten Rückführung unten gebraucht. Plan vor dem Start ergänzt: Tabelle in §3, Block unten, Risiken in §6.

**Größenmarke** (`SPEC-128` Nr. 1, `AGENTS.md` §3.14; Planner beim Start, vor dem ersten Code-Commit): sechs Zeilen der Tabelle (LZ1 bis LZ6), *a* = 4 (LZ1, LZ2, LZ3, LZ5), *b* = 2, *f* = 0, *p* = 0 (nur Test-Code): *S* = 85·4 + 80·2 = 500, Marke 700, keine Schnittfrage. Den Teil des Blocks, den sonst der Architect setzt (`pfad`, `produktcode`, `fakes`), hat der Planner gesetzt. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Ändert der Implementer eine weitere Datei (etwa `uhr_test.go`), nennt er sie im selben Commit im Block.

```groessenmarke
pfad: tests/lasttest/leser_test.go
produktcode: 0
fakes: 0
basis: 679059d0
schaetzung: 500
marke: 700
```

**Rückführungen — vorab benannt:**

- `in-progress` → `open` (blockiert): die Zusage lässt sich ohne Echtzeitgrenze nicht fangen; dann entscheidet der Architect über ein benanntes Negativ.
- `in-progress` → `next` (zu groß): die Marke liegt nach `SPEC-128` über der Schnittgrenze.

## 5. Closure-Trigger

DoD vollständig, Review bestanden, `make gates` grün, Closure-Notiz mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Ohne Zeitgrenze fängt der Test die Mutation „Pause beim Auswerten“ nicht mehr — **Ausgang:** entfallen: die Mutation LZ1 ist mit unverändertem Wert (1 ms) rot, beim Implementer in 20 von 20 Prozessen ohne und mit 8 Schleifen, beim Verifier in 10 von 10 und 5 von 5.
- Auch eine bloß angehobene Grenze kann diese Mutation durchlassen (gerechnet, nicht gemessen): der Lauf mit 20 MiB schreibt 5232 Span-Zeilen, die Pause von 1 ms verlängert ihn um etwa 5 s, die 19 Läufe zu 1 MiB um je etwa 0,26 s. Eine Grenze von mehr als etwa 5 s macht den Fänger an der heutigen Form blind — **Ausgang:** entfallen: der Fänger hebt keine Grenze an, er verlangt, dass die Uhr der Blase steht; LZ1 ist rot (Zeile davor), die erste Rückführung in §4 griff nicht.
- `SPEC-216` Nr. 14 a lässt allein `uhr.go` die Zeit lesen (`TestPaketOhneZeit`): eine Grenze relativ zu einem Vergleichslauf darf in `leser_test.go` kein `time.Now` nennen — **Ausgang:** entfallen: der Slice braucht keinen Vergleichslauf; `pipeLauf` liest die Uhr über `Uhr` mit `systemuhr`, und `TestPaketOhneZeit` läuft in den gestempelten Gates grün (Verifikation).
- Der Fänger sieht eine Verlangsamung, während die Uhr der Blase steht, nicht (Negativ in §3, 4 von 4 geprüften Formen); der alte Stand sah sie nach Rechnung, sobald der Lauf mit 20 MiB damit über 1 s lag. Entschieden hat der Architect am 2026-11-27 (`SPEC-216` Nr. 16 s): das Negativ trägt, erweitert um jedes Warten, bei dem die Uhr der Blase steht; kein weiterer Fänger, keine Rückführung des Slice — **Ausgang:** weiter offen: → `BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger` im Register (die Lücke bleibt, die Anpassung an einen belasteten Host hat sie geöffnet).
- Ein Hänger durch eine Warteschleife auf die Uhr oder durch eine Sperre, deren Halter auf sie wartet, bricht das ganze Test-Binary nach 60 s ab, statt den Test mit seinem Grund rot zu melden (§3; Review `2026-11-27-slice-lasttest-leser-zeitgrenze-review`, I-1 und F-2; der Kopf von `pipeLauf` nennt es) — **Ausgang:** weiter offen: → `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register; kein Fix, der Lauf ist rot, und der Grund ist aus dem Stapel zu lesen.
- Randformen (`AGENTS.md` §3.12, §3.16): kein Raster, weil der Slice keinen neuen Vertrag liefert und das Verhalten des Lesers unverändert bleibt — **Ausgang:** entfallen: keine Randform brauchte eine Entscheidung, das Produktverhalten ist unverändert.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register.

**Closure** (Planner, 2026-11-27; Zweig `slice-lasttest-leser-zeitgrenze`, Messbasis `679059d0`; der Slice entstand aus der Closure von `slice-replay-speicher-und-suche`). Start `3d352b2d`, Code und Plan `a2100d4d`, Review-Report `75fbb584`, Fix-Runde `2170d89a`, Verifikations-Report `fc4b5368`, Merge von `main` vor der Closure `faeefeff`. Review (`2026-11-27-slice-lasttest-leser-zeitgrenze-review`: 0 HIGH, 1 MEDIUM, 1 LOW, 2 INFO) und Verifikation (`2026-11-27-slice-lasttest-leser-zeitgrenze-verifikation`: bestanden, 2 INFO an den Planner) liefen je einmal. Merge nach `main` und `make gates` im Hauptbaum sind der Integrationsschritt (Übergangsregeln Nr. 5).

**Was geliefert ist** (kein Produktcode): `pipeLauf` in `tests/lasttest/leser_test.go` läuft in einer Blase von `testing/synctest`. An die Stelle der Grenze von 1 s Echtzeit tritt, dass die Uhr der Blase beim Lesen und Auswerten stehen bleibt und kein Write scheitert; um den ganzen Test bleibt die Wache von 60 s. Dazu entfiel der Stand-Vermerk von `SPEC-216` Nr. 16 s. Am Binary ändert sich nichts.

**Größe** (`SPEC-128` Nr. 9). Stand *Ende*, Zeile des Programms (`make groessenmarke`, 2026-11-27, Plan in `in-progress/`, nach dem Merge `faeefeff`): `Messpunkt slice-lasttest-leser-zeitgrenze.md: zeilen=6 nebenlaeufig=4 fakes=0 produktcode=0 schaetzung=500 marke=700 messung=76`. Produktcode von Hand: 0; `git diff --numstat 679059d0..faeefeff` über den Pfad des Blocks nennt 49 und 27. Stand *Handoff*: keine Zeile des Programms eingetragen; von Hand `679059d0..a2100d4d` 73. Auslastung 76 ÷ 700 = **0,11**, Ende ÷ *S* (500) = 0,15, Ende ÷ Handoff = 1,04. Fremde Zeilen 0. **Meldung an den Architect** (Auslastung unter 0,45): er rechnet die Faktoren aus allen Messpunkten neu (`SPEC-128` Nr. 2 und 9); der Punkt ist `zeilen=6`, `nebenlaeufig=4`, `fakes=0`, `produktcode=0`, *S* 500, *M* 700, Ende 76. *Ursache:* die sechs Zeilen der Tabelle sind Zusagen eines bestehenden Tests; der Slice baute allein `pipeLauf` um, LZ4 bis LZ6 hatten ihre Fänger schon, und neue Testfälle kamen nicht dazu.

**Mutationen** (§3 und Reports): der Implementer je Zeile LZ1, LZ2, LZ4, LZ5 und LZ6 rot in 20 von 20 Prozessen ohne und mit 8 Schleifen, dazu acht weitere Formen in 20 von 20; das Review sah m3 grün (F-1, benanntes Negativ); der Verifier fuhr die Zeilen erneut und drei eigene Stellen in `lies`, alle rot. LZ3 ist eine Messung: der unveränderte Leser war am neuen Stand in keinem Prozess rot (Implementer 60, Verifier 45), am alten bei 8 und 16 Schleifen in 15 und 20 von 20 (Verifier 4 von 8 und 3 von 3). Last und Messende stehen in §3.

**Akzeptierte Negative:** (1) Verlangsamung von Leser oder Auswertung, während die Uhr der Blase steht (§3; vom Architect als `SPEC-216` Nr. 16 s eingetragen, mit Grund, verworfenen Alternativen und Wirkung im Messlauf); (2) ein Hänger endet nicht immer als Fehler des Tests, sondern erst nach 60 s im panic der Wache (§3); (3) die Wache von 60 s ist oberhalb von 16 Schleifen nicht gemessen (§3); (4) die Zusicherung über die Uhr trägt LZ1 allein (§3).

**Was weder Review noch Verifikation sahen:** (1) der Reviewer sah die Fix-Runde nicht, der Verifier sah sie; (2) die Sätze zum Hänger im Kopf von `pipeLauf` fuhr der Verifier nicht erneut, sie stehen auf m4 und m6 des Reviews; (3) `make lasttest` fuhr niemand (§1: kein Lauf nötig), `make gegenproben` niemand (`tests/` ist keine Eingabe der Gegenproben); (4) die Texte der Closure (Risiko-Ausgänge, Register, Roadmap, Welle-Datei, der Satz zur Sperre in §3) hat niemand geprüft.

**Gate:** `.harness/state/gates-passed.head` im Worktree trägt `2170d89a` (die Verifikation sah den Hash des Arbeitsbaums gleich; welche Stufen aus dem Schichten-Cache kamen, sagt der Stempel nicht). Die Commits danach sind von keinem Lauf der Kette gedeckt, darunter der Merge `faeefeff`, der Testcode und den Treiber von `slice-bootstrap-m6` bringt. Der Planner fuhr `make docs-check` (nach dem Merge grün), `make zeichen-check` und `make groessenmarke`, nicht `make gates` (Vorgabe des Orchestrators). Der Beleg auf dem Stand von `main` ist der Lauf des Integrationsschritts.

**Abweichungen vom Plan:** nach dem Code erweiterte der Architect das Negativ um jedes Warten, bei dem die Uhr der Blase steht (F-1, `SPEC-216` Nr. 16 s), und die Fix-Runde berichtigte den Kopf von `pipeLauf` (F-2). Keine Rückführung aus §4 griff.

**Steering-Loop-Einträge:**

1. **Neuer Sensor:** `TestLeserHaeltNieAn` misst an der Uhr einer Blase von `testing/synctest`; er trennt ein Warten auf diese Uhr von der Last des Hosts.
2. **Lernsignal:** der Test war unter Last rot, weil er die Zusage *der Leser hält den Schreiber nie an* als Dauer in Echtzeit maß: 20 MiB brauchten unter dem Detektor 0,84 bis 1,19 s bei 8 und 1,80 bis 2,12 s bei 16 Schleifen gegen eine Grenze von 1 s. Eine angehobene Grenze trennt nicht, weil die Mutation den Lauf nur um 5,2 s verlängert und der unveränderte Lauf mit der Last wächst.
3. **Architektur-Beobachtung:** eine virtuelle Uhr ersetzt eine Echtzeitgrenze nicht gleichwertig. Sie sieht allein das dauerhafte Blockieren in der Blase; Rechenzeit, Systemaufruf, Sperre und Kanal von außen sieht sie nicht (4 von 4 geprüften Formen grün), und ein Warten auf eine Sperre, deren Halter auf die Uhr wartet, wird in ihr zum Hänger. Wer eine Zeitgrenze so ersetzt, entscheidet über ein Negativ am Vertrag, nicht allein über den Test.
4. **Ein Negativ wurde zuerst zu eng benannt** (F-1: Rechenzeit und Systemaufruf, nicht das Warten, das `synctest` nicht zählt), weil es aus den gefahrenen Formen und nicht aus der Regel des Werkzeugs gebildet war. Evidenz zu `BEO-REPO/ein-negativ-benennt-die-ungefangene-klasse-zu-eng`. Kein neuer Zielort (`AGENTS.md` §3.15 trägt es).
5. **Kein `liegt in`-Feld:** keine Regel ist mit diesem Slice verkörpert.

**Lerneintrag:** neuer Sensor (Eintrag 1) mit dem Lernsignal und der Architektur-Beobachtung (Einträge 2 und 3).

**Reviewer-Skill:** keine Änderung. F-1 trifft die MEDIUM-Zeile *Fänger schmaler als Zusage, Menge unbeschränkt oder ohne Achsenliste*, F-2 die Zeile *Fixture, Hilfe oder Kommentar verdeckt, was der Test prüft*, I-1 den Hinweis zu `panic:` und Zeitlimit. Keine neue Klasse.

**Register** (Evidenz mit der Closure im Zweig; kein Zähler gesetzt, er folgt aus den Dateien): Evidenz `slice-lasttest-leser-zeitgrenze` zu fünf vorhandenen Einträgen: `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (27 → 28 Dateien), `kommentar-nennt-mehr-elemente-als-der-test-prueft` (19 → 20), `panik-eines-fakes-verdeckt-den-benannten-test` (11 → 12), `anpassung-eines-bestehenden-tests-lockert-seinen-faenger` (1 → 2) und `schaetzformel-ueberschaetzt-den-slice` (31 → 32). Kein neuer Eintrag; 99 Einträge, 813 → 818 Evidenz-Dateien. `test-fremden-pakets-flackert-unter-last`: keine neue Evidenz (der Vorgang ist die Behebung, kein Auftreten), der Stand ist nachgezogen; der Eintrag bleibt offen für `urlparser`, `admin` und `httpodata`. **Schwellen:** kein neuer Übertritt.

**Risiko-Ausgänge (§6):** 6 Punkte, je ein Ausgang: entfallen 4, eingetreten 0, weiter offen 2 (beide im Register).

**Sendungen (`AGENTS.md` §3.13):** *an diesen Slice* aus `slice-replay-speicher-und-suche` (§1 *Übernimmt*): für `TestLeserHaeltNieAn` geliefert. *Von diesem Slice:* an keinen Slice. *An den Architect:* die Meldung der Auslastung; `SPEC-216` Nr. 16 s führt die Sperre als nicht gefahren und nennt 3 von 3 Formen, die Verifikation fuhr sie (§3, 4 von 4).

**Folge-Slice:** keiner. Kein Befund und kein Risiko verlangt Arbeit, die an einen Slice gebunden ist; die übrigen lastabhängigen Tests haben ihre Adresse beim Lese-Schritt des Registers (§1).

**Drei Paarungen:** Anker (kein Eintrag trägt ein Zielort-Feld) · Folge-Slice (keiner genannt) · Register (die sechs genannten Verzeichnisse tragen `evidence/`).

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** Register beim Start gesichtet (2026-11-27; der Zähler ist die Zahl der Dateien unter `evidence/`). Offen: `BEO-REPO/test-fremden-pakets-flackert-unter-last` (3×, Schwelle erreicht, noch ohne Ausgang; dieser Slice ist die Adresse für `tests/lasttest`), `BEO-REPO/zeit-zusage-an-einer-form-und-einer-locale` (1×; eine Grenze der Zeit gilt für die gemessene Form und Last, DoD 1 und 2) und `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×; die Marke steht in §4). Verkörpert und zu beachten: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (15×) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (27×) in `AGENTS.md` §3.10, `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last` (3×, ebendort: mehrere Läufe mit Last), `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (23×, Command-Guard). Geplant: `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×; der Block nennt eine Datei).
