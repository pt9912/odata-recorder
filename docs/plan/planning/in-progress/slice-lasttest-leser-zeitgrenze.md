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

**Berührte Spec-Stellen:** [SPEC-216](../../../../harness/vertraege/SPEC-216-lasttest.md) Nr. 9 e (das Programm entleert die Ströme des Binary ohne Unterbrechung und liest nie langsamer, als das Binary schreibt) und Nr. 10 a (Zeilen bis 1 MiB, letzte Zeile, Fehler des Stroms; nennt `TestLeserHaeltNieAn` als Fänger der zwei letzten); beide Zusagen bleiben, der Vertrag wird nicht geändert.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-lasttest-leser-zeitgrenze`, Worktree `/Development/odata-recorder-wt-leser`; der Zweig entsteht von `main` nach dem Start-Commit, der Start liegt auf `main`)

**Autor:** pt9912 (Planner, Closure von `slice-replay-speicher-und-suche`). **Datum:** 2026-11-27.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice.

**Übernimmt:** aus `slice-replay-speicher-und-suche` (Plan §6, Verifikation V-6): `make test-race` war in einem Lauf von `make gates` rot an `TestLeserHaeltNieAn` (`tests/lasttest/leser_test.go`, „nach 1s kein Wert“) bei hoher Last des Hosts.

**Ziel:** `TestLeserHaeltNieAn` fängt weiter, dass der Leser den Schreiber nie anhält, und wird auf einem belasteten Host nicht allein deshalb rot, weil 20 MiB unter dem Detektor nicht binnen 1 s Echtzeit durch die Pipe gehen (`pipeLauf`, `warteAuf(…, time.Second)`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung am Leser oder an den Zusagen von `SPEC-216` Nr. 9 e und 10 a — der Befund betrifft den Fänger, nicht das Werkzeug; ein Lauf des Lasttests ist nicht nötig.
- Die weiteren lastabhängigen Tests in `urlparser`, `admin` und `httpodata` (`BEO-REPO/test-fremden-pakets-flackert-unter-last`) — andere Pakete, Ursache nicht untersucht; Adresse Planner beim Lese-Schritt des Registers.
- Eine Lockerung eines Gates (`make test-race` bleibt, wie es ist; `AGENTS.md` §3.6).

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Ursache belegt: die Dauer von `pipeLauf` mit 20 MiB unter dem Detektor steht mit Last des Hosts und Messendem im Plan.
- [ ] Der Fänger hängt nicht mehr an der Grenze von 1 s, oder die Grenze ist aus der Messung begründet; die Mutationen der Zeile A1 des Plans von `slice-lasttest-auswertung` (Scanner mit 64 KiB, Pause beim Auswerten, Grenze entfällt, Sammler ohne Sperre) bleiben je rot, mit Rate (`AGENTS.md` §3.10, §3.15).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tests/lasttest/leser_test.go` | update | `pipeLauf` und `TestLeserHaeltNieAn` |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Planner beim Start, 2026-11-27; kein Architect-Lauf, weil der Slice keinen neuen Vertrag liefert). Die Zeilen folgen der Zeile A1 des Plans von [`slice-lasttest-auswertung`](../done/slice-lasttest-auswertung.md) und der DoD; die Spalten sind die von `SPEC-128` Nr. 3 a. Der Implementer trägt Mutationen und Raten mit Zahl der Prozesse, Last und Messendem in einen Absatz darunter ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| LZ1 | (Nr. 9 e) Ein Leser, der beim Auswerten pausiert, macht den Fänger rot | `TestLeserHaeltNieAn` (`pipeLauf`) | 1 ms Pause je Span-Zeile in `nimm` (`a1-pause-beim-auswerten`); der Wert der Mutation bleibt der aus `slice-lasttest-auswertung` | nebenläufig |
| LZ2 | (Nr. 9 e) Ein Lesen, das endet, solange der Schreiber noch schreibt, macht den Fänger rot | `TestLeserHaeltNieAn` (`pipeLauf`) | `bufio.Scanner` mit der Grenze von 64 KiB (`a1-scanner64`) | nebenläufig |
| LZ3 | Am unveränderten Leser wird der Fänger auf einem belasteten Host nicht allein an der Zeit rot | Messung, kein deterministischer Test: Prozesse von `TestLeserHaeltNieAn` unter dem Detektor bei belastetem Host, mit Zahl, Last und Messendem, dazu die Dauer von `pipeLauf` (DoD 1) | die Grenze von 1 s bleibt (Stand vor dem Slice); die Rate hängt an der Last | nebenläufig |
| LZ4 | (Nr. 10 a) Eine Zeile über 1 MiB wird verworfen und gezählt, der Speicher wächst nicht mit der Länge | `TestLeserHaeltNieAn` (1 MiB + 1 B, 64 MiB in einer Zeile) | die Grenze von 1 MiB entfällt (`a1-grenze-entfaellt`) | andere |
| LZ5 | Der Sammler, in den der Leser schreibt, ist unter Nebenläufigkeit sicher | `TestSammlerGleichzeitig` unter `make test-race` (Bestand, `auswertung_test.go`) | Sperre fehlt in `nimm` (`a1-sammler-nimm-ohne-sperre`) | nebenläufig |
| LZ6 | (Nr. 10 a) Eine letzte Zeile ohne Zeilenende ist eine Zeile, ein Fehler des Stroms geht an den Aufrufer | `TestLeserHaeltNieAn` | die letzte Zeile ohne Zeilenende wird verworfen; der Fehler des Stroms wird verschluckt | andere |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-replay-speicher-und-suche` liegt in `done/` und ist nach `main` gemergt. Wird ein Lauf von `make gates` vorher an diesem Test rot, hat der Slice Vorrang.

**Stand beim Start** (Planner, 2026-11-27): Die Bedingung ist erfüllt: `slice-replay-speicher-und-suche` liegt in `done/` und ist nach `main` gemergt (`ce6736cb`). In `in-progress/` liegt daneben allein [`slice-bootstrap-m6`](slice-bootstrap-m6.md); `git diff --name-only main...slice-bootstrap-m6` nennt beim Start elf Pfade, keinen unter `tests/lasttest/`. Kein Kettenslice (Nr. 7 der Übergangsregeln). Keine Frage an den Architect ist offen; er wird erst mit der ersten Rückführung unten gebraucht. Plan vor dem Start ergänzt: Tabelle in §3, Block unten, Risiken in §6.

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

- Ohne Zeitgrenze fängt der Test die Mutation „Pause beim Auswerten“ nicht mehr — **Ausgang:** weiter offen: die Mutationen der Zeile A1 sind die DoD.
- Auch eine bloß angehobene Grenze kann diese Mutation durchlassen (gerechnet, nicht gemessen): der Lauf mit 20 MiB schreibt 5232 Span-Zeilen, die Pause von 1 ms verlängert ihn um etwa 5 s, die 19 Läufe zu 1 MiB um je etwa 0,26 s. Eine Grenze von mehr als etwa 5 s macht den Fänger an der heutigen Form blind — **Ausgang:** weiter offen: DoD 2 verlangt die Mutation mit unverändertem Wert rot; gelingt das nicht, greift die erste Rückführung in §4.
- `SPEC-216` Nr. 14 a lässt allein `uhr.go` die Zeit lesen (`TestPaketOhneZeit`): eine Grenze relativ zu einem Vergleichslauf darf in `leser_test.go` kein `time.Now` nennen — **Ausgang:** entschieden (Vertrag): Wartezeiten laufen über `warteAuf` und `wache`.
- Randformen (`AGENTS.md` §3.12, §3.16): kein Raster, weil der Slice keinen neuen Vertrag liefert und das Verhalten des Lesers unverändert bleibt — **Ausgang:** entfallen; eine Randform, die doch eine Entscheidung braucht, geht an den Architect.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register.

Beim Anlegen entstanden aus der Closure von `slice-replay-speicher-und-suche`; gestartet am 2026-11-27, noch keine Closure.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** Register beim Start gesichtet (2026-11-27; der Zähler ist die Zahl der Dateien unter `evidence/`). Offen: `BEO-REPO/test-fremden-pakets-flackert-unter-last` (3×, Schwelle erreicht, noch ohne Ausgang; dieser Slice ist die Adresse für `tests/lasttest`), `BEO-REPO/zeit-zusage-an-einer-form-und-einer-locale` (1×; eine Grenze der Zeit gilt für die gemessene Form und Last, DoD 1 und 2) und `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×; die Marke steht in §4). Verkörpert und zu beachten: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (15×) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (27×) in `AGENTS.md` §3.10, `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last` (3×, ebendort: mehrere Läufe mit Last), `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (23×, Command-Guard). Geplant: `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×; der Block nennt eine Datei).
