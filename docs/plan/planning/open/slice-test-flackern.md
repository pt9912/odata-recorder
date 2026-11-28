# Slice slice-test-flackern: Drei Tests, die unter Last rot werden, hängen nicht mehr an Zeit oder Reihenfolge

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung, die von der DoD verschieden wäre.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe), [ADR-0007](../../adr/0007-race-detektor-im-gate.md), Register `BEO-REPO/test-fremden-pakets-flackert-unter-last`.

**Berührte Spec-Stellen:** `SPEC-321` Nr. 9 a (Spans gleichzeitiger Anfragen) · `SPEC-160` Nr. 6 (Längengrenzen der URL) · `SPEC-136` Nr. 4 e (Request-Body an den Dienst). Der Slice ändert keine Stelle.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-test-flackern`, Worktree `/Development/odata-recorder-wt-flackern`; der Zweig entsteht von `main` nach den Start-Commits, der Start liegt auf `main`)

**Autor:** pt9912 (Planner, Lese-Schritt der Closure von `welle-betrieb`). **Datum:** 2026-11-28.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** `TestGleichzeitigeSpans` (`src/hexagon/services/hybrid`), `TestUeberlangeEingabe` (Paket `urlparser`) und `TestAnfrageBodyBytes` (`src/adapters/driving/httpodata`) werden unter Last nicht mehr aus einem Grund rot, den das Produkt nicht hat.

**Übernimmt** (Lese-Schritt, `AGENTS.md` §3.13): die offenen Fälle aus `BEO-REPO/test-fremden-pakets-flackert-unter-last`. Die Belege liegen unter `evidence/` des Eintrags.

**Gelesene Ursachen** (Planner beim Start, 2026-11-28, `main` bei `7c9014bb`; gelesen, nicht nachgestellt, keine davon ist belegt):

- `TestGleichzeitigeSpans`: `gleichzeitigeRunde` ruft `cancel()` für die Anfragen des Szenarios Abbruch und gibt unmittelbar danach die Schranke der Stufe Upstream frei. In `schranke.warte` sind dann für eine Goroutine, die das `select` noch nicht erreicht hat, `frei` und `ctx.Done()` bereit, und die Wahl ist zufällig.
- `TestUeberlangeEingabe`: `gesamtZuteilung` liest `TotalAlloc` und zählt damit die Zuteilungen des ganzen Prozesses; die Grenze ist absolut (4096 Bytes), beobachtet waren 5576. Vermutet ist eine fremde Zuteilung während der Messung; welche, ist nicht bekannt.
- `TestAnfrageBodyBytes`: `u.client()` (`fake_test.go`) setzt `Timeout` 10 s je Request; die beobachtete Meldung (`Client.Timeout exceeded while awaiting headers` bei 8388609 Bytes) passt zu dieser Frist unter dem Detektor bei Last.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `TestLeserHaeltNieAn` — behoben mit [`slice-lasttest-leser-zeitgrenze`](../done/slice-lasttest-leser-zeitgrenze.md).
- `TestWettlaufNebenlaeufig` — behoben mit [`slice-token-datei-wettlauf`](../done/slice-token-datei-wettlauf.md).
- Eine Änderung am Produktcode — ein Fehler des Produkts geht an den Planner.
- Eine höhere Zeitgrenze als Behebung — sie verschiebt den Bruch nur. Als höhere Grenze zählt jede größere Zahl für dieselbe Messung, auch die bloße Streichung der Frist, nach der allein die Zeitgrenze des Pakets bleibt. Eine Schutzfrist gegen Hängen, die statt der Gesamtdauer den Stillstand misst oder an einem Ereignis hängt, ist keine höhere Grenze.
- Eine höhere Grenze der Zuteilung in `TestUeberlangeEingabe` (4096 Bytes) — dieselbe Verschiebung.
- Kleinere oder weniger Eingaben (Größen und Bindungen in `TestAnfrageBodyBytes`, vier Formen in `TestUeberlangeEingabe`, 16 Goroutinen je Szenario und fünf Runden in `TestGleichzeitigeSpans`) — sie schwächen den Fänger.
- Weitere Tests mit derselben Form außerhalb der drei Dateien in §3 — der Implementer nennt sie in §6 (`AGENTS.md` §3.17), der Planner entscheidet über einen Folge-Slice.
- Änderungen an `harness/`, `Makefile`, `Dockerfile` oder `tools/` — [slice-ci-github-actions](../in-progress/slice-ci-github-actions.md) hält die Gate-Kette (Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)); dieser Slice ist kein Kettenslice. Wer es müsste, gibt an den Planner zurück.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] `TestGleichzeitigeSpans` wartet an einem Haltepunkt statt auf die Reihenfolge der Goroutinen; die Ursache ist vorher nachgestellt.
- [ ] Für `TestUeberlangeEingabe` und `TestAnfrageBodyBytes` ist die Ursache untersucht; je Test eine Behebung oder ein akzeptiertes Negativ mit Grund, Folge und Maß.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/services/hybrid/telemetrie_gleichzeitig_test.go` | update | Haltepunkt: auf die Rückkehr der abgebrochenen Aufrufe warten |
| `src/adapters/driven/odata/urlparser/grenzen_test.go` | update | Messung von `TestUeberlangeEingabe` nach Befund |
| `src/adapters/driving/httpodata/anfrage_test.go` | update | Frist von `TestAnfrageBodyBytes` nach Befund; ein eigener Klient steht in dieser Datei, `u.client()` bleibt für die übrigen Tests |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Planner beim Start, 2026-11-28; kein Architect-Lauf, weil der Slice keinen neuen Vertrag liefert und keine Entscheidung offen ist: er ändert Tests, ihre Zusagen stehen in den Spec-Stellen im Kopf). Spalten nach `SPEC-128` Nr. 3 a. Zwei Arten von Zeilen: das Urteil hängt nicht mehr an der Störung (Nachstellung), und der Test fängt weiter, was er vorher fing (Mutation). Nachstellung und Mutation laufen nur in einer Wegwerfkopie im Container, nie als Änderung im Zweig.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| FK1 | (`SPEC-321` Nr. 9 a) Das Urteil von `TestGleichzeitigeSpans` hängt nicht davon ab, welchen bereiten Fall das `select` der Schranke wählt: die Runde gibt die Stufe Upstream erst frei, wenn alle 16 abgebrochenen Aufrufe zurückgekehrt sind | Nachstellung: Verzögerung zwischen Ankunft und `select` in `schranke.warte`; vor der Behebung rot, mit ihr grün | das Warten auf die Rückkehr entfernt, die Verzögerung bleibt | nebenläufig |
| FK2 | (`SPEC-321` Nr. 9 a) Art und Ergebnis je Stufe, darunter `abgebrochen`, je Span ein `Set` und ein `End` und die Summen 48, 48 und 16 prüft der Test unverändert | `TestGleichzeitigeSpans` | im Dienst: Abbruch endet als `antwort`; ein zweites `End`; ein Zähler fehlt | nebenläufig |
| FK3 | Kehrt ein abgebrochener Aufruf nicht zurück, endet der Test mit eigenem Fehler innerhalb von `frist`, nicht mit der Zeitgrenze des Pakets | `TestGleichzeitigeSpans` in der Wegwerfkopie | der Fake des Upstream beachtet `ctx` nicht | nebenläufig |
| FK4 | Eine einmalige fremde Zuteilung während einer Messung macht `TestUeberlangeEingabe` nicht rot | Nachstellung: eine fremde Zuteilung über 4096 Bytes in einer Messung; vor der Behebung rot, mit ihr grün | die Behebung entfernt, die Zuteilung bleibt | andere |
| FK5 | (`SPEC-160` Nr. 6) Liest der Parser eine überlange Eingabe, wird der Test in jeder der vier Formen rot; die Grenze bleibt 4096 Bytes | `TestUeberlangeEingabe` | je Form: Dekodieren oder Zerlegen vor der Prüfung der Länge | andere |
| FK6 | Das Urteil von `TestAnfrageBodyBytes` hängt nicht an der Gesamtdauer der Übertragung (oder akzeptiertes Negativ nach DoD Punkt 2) | Nachstellung: der Dienst liest den größten Body über mehr als 10 s verteilt; vor der Behebung rot, mit ihr grün | die Behebung entfernt, die Verzögerung bleibt | nebenläufig |
| FK7 | Ein Request ohne Antwort endet als Fehler dieses Tests, nicht mit der Zeitgrenze des Pakets | `TestAnfrageBodyBytes` in der Wegwerfkopie | der Dienst kehrt nicht zurück, die Schutzfrist entfernt | nebenläufig |
| FK8 | (`SPEC-136` Nr. 4 e) Alle acht Größen an beiden Bindungen kommen unverändert und ohne Grenze beim Dienst an | `TestAnfrageBodyBytes` | der Adapter kürzt bei 8 MiB; er ändert ein Byte | andere |

Der Bericht nennt je Zeile Mutation und roten Lauf; wird eine nur mit Wahrscheinlichkeit rot, steht die Rate mit Zahl der Prozesse, Last und Messendem (`AGENTS.md` §3.10), und der Absatz steht auch hier in §3. Für die unveränderten drei Tests mit Behebung nennt er die Läufe ohne und mit Detektor (Zahl, Last, Messender).

**Randformen:** Das Raster nach `AGENTS.md` §3.16 entfällt: es gibt keinen neuen Vertrag und keine neue Operation.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein Trigger außer der Priorisierung; nicht Teil der Erstabnahme.

**Stand beim Start** (Planner, 2026-11-28, `main` bei `7c9014bb`, durch Lesen und Suchen, ohne Testlauf): Die Bedingung ist erfüllt. Keine Frage an den Architect ist offen. *Kollision:* in `in-progress/` liegt daneben allein [slice-ci-github-actions](../in-progress/slice-ci-github-actions.md); sein Zweig ändert gegen `main` den Workflow, `harness/` und seinen Plan, keinen Pfad unter `src/`. Dieser Slice ändert keinen davon. Er ist keine Voraussetzung der CI und sie keine von ihm; wer später mergt, nimmt `main` vorher auf (Nr. 5 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)).

**Größenmarke** (`SPEC-128` Nr. 1, `AGENTS.md` §3.14; Planner beim Start, vor dem ersten Code-Commit): acht Zeilen der Tabelle (FK1 bis FK8), *a* = 5 (FK1, FK2, FK3, FK6, FK7), *b* = 3, *f* = 0, *p* = 0 (nur Test-Code): *S* = 665, Marke **900**, keine Schnittfrage. Den Teil des Blocks, den sonst der Architect setzt (`pfad`, `produktcode`, `fakes`), hat der Planner gesetzt. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Ändert der Implementer eine weitere Datei (etwa `fake_test.go`), nennt er sie im selben Commit in §3 und im Block. Für einen Schnitt beim Handoff ist kein Teil vorab benannt; eine Überschreitung steht mit Grund im Block.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß): über 900 Zeilen (Metrik wie oben). Eine Überschreitung schneidet nicht (`SPEC-128` Nr. 10 c); sie bekommt einen Eintrag `ueberschreitung` mit Grund im Block.
- `in-progress` → `open` (blockiert): eine Ursache liegt im Produkt, oder eine Behebung verlangt eine Änderung an Produktcode, Gate oder Spezifikation; Rückgabe an den Planner.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Eine Ursache lässt sich nicht nachstellen — **Ausgang:** weiter offen; dann trägt der zweite Liefer-Punkt ein Negativ mit der gemessenen Zahl der Läufe.
- Eine wiederholte Messung (etwa das Minimum) verdeckt eine Zuteilung des Parsers, die nicht in jedem Aufruf anfällt (Puffer aus einem Pool) — **Ausgang:** offen, Adresse Implementer: FK5 an jeder Form; was bleibt, ist ein Negativ mit Grund und Folge. Dasselbe gilt für fremde Zuteilungen, die in jeder Wiederholung anfallen.
- Für `TestAnfrageBodyBytes` bleibt keine Behebung außer einer höheren Grenze (§1) — **Ausgang:** offen, Adresse Implementer: dann akzeptiertes Negativ mit Grund, Folge und gemessener Rate; FK6 und FK7 stehen als Negativ.
- Dieselbe Form steht in weiteren Tests (Freigabe einer Schranke nach `cancel()` ohne Warten; Zuteilung mit absoluter Grenze; Gesamtfrist je Request über großen Bodies) — **Ausgang:** offen, Adresse Implementer: er nennt die Menge je Form mit Zahl der Elemente (`AGENTS.md` §3.17) und behebt, was in den Dateien aus §3 liegt; der Rest geht an den Planner.
- `TestGleichzeitigeSpans` wird nach der Behebung seltener rot gegen einen geteilten Zustand, weil die Abbrüche nicht mehr mit den Antworten überlappen — **Ausgang:** offen, Adresse Reviewer: der Kommentar des Tests nennt nur, was er fängt (`AGENTS.md` §3.11).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten** (Planner beim Start, 2026-11-28; Zähler gleich der Zahl der Dateien unter `evidence/`, alle im Kürzel `BEO-REPO`):

- **Gegenstand:** `test-fremden-pakets-flackert-unter-last` (4×, geplant in diesem Slice). `sqlite-transaktionsvertrag-rot-unter-last` ist gestrichen (Ursache belegt und behoben) und nicht Teil des Slice.
- **Zu beachten:** `rate-ohne-detektor-schwankt-mit-der-last` (15×, verkörpert in `AGENTS.md` §3.10; jede Rate mit Last und Messendem), `raten-messung-mit-zusatzlast-im-selben-container` (1×; die Zusatzlast einer Messung benennen), `paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (4×, offen; `httpodata`, `hybrid` und `urlparser` bekommen keine teuren Wiederholungen), `fixture-gleicht-dem-wert-der-mutation` (17×), `gate-lauf-stufen-aus-dem-schicht-cache` (21×; die Läufe der drei Tests frisch), `betriebsregel-fuer-repo-dateien-verletzt` (25×; Übergangsregeln Nr. 8).
- **Planung:** `marke-beim-start-nicht-gesetzt` (1×; Marke mit diesem Start gesetzt), `schaetzformel-ueberschaetzt-den-slice` (34×; bei einem reinen Test-Slice liegt das Ende eher weit unter der Marke).

Kein Eintrag verlangt vor diesem Slice einen eigenen Slice.

**Modus-Begründungsblock:** entfällt, alle berührten Sub-Areas GF.
