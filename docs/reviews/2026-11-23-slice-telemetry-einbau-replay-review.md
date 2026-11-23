# Review-Report: slice-telemetry-einbau-replay (Teil 2) — 2026-11-23

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD)

**Gegenstand:** Zweig `slice-telemetry-einbau-replay`, HEAD `2eef966f` (Handoff-Commit, Code und Plan zusammen), Messbasis `0c644a15`; Teil 2 des Schnitts in drei Teile, Zeilen W1 bis W13: Match-Span `odata.replay.match` im Replay-Dienst (`src/hexagon/services/replay/{replay.go,span.go}`) und ein Adapter der Telemetrie je Prozess in `replay` und `hybrid` (`src/bootstrap/wurzel.go`), dazu die Tests im Paket `replay` und in `src/bootstrap/`. Teil 3 (Capture-Span) ist nicht Gegenstand.

**Skill:** `.harness/skills/reviewer.md` @ `c3741274` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-23

**Eingangs-Kontext:**

- Plan `slice-telemetry-einbau-replay`: §1 (Ziel, Sendungen), §2 (DoD, nur zur Abgrenzung), §3 (Tabelle W1 bis W13, Mengen mit Zählprüfung, Fundstellen der Adapter-Zählung, Absätze *Mutationen und Raten* und *Raten von W13*, akzeptierte Negative (1) bis (7)), §4 (Block, Marke 1800, Schätzung 1290, Basis `0c644a15`), §6 (Abweichungen (a) bis (c) mit Ausgang *offen*, Adresse Architect; Raster), §8.
- `SPEC-296` Nr. 1 bis 9 und *Ungeprüft*, `SPEC-135` Nr. 1, 2, 3, 9 bis 11, 15 bis 18, `SPEC-139` Nr. 3, 11, 13 c, `SPEC-230`, `SPEC-231` Nr. 1 bis 7, `SPEC-121` Nr. 12, `SPEC-195` Nr. 10 und 11, `SPEC-136` Nr. 7 a und b, `SPEC-108` Nr. 1, `SPEC-109` Nr. 6.
- `AGENTS.md` §3.7, §3.9 bis §3.18.
- `LH-QA-01`, `LH-FA-02`, `LH-QA-05`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Bezug des Plans, nur genannt).
- Vorherige Reviews am Modul: `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` (Klassen *Fänger führt nur eine Form der Eingabe*, *Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt*).

**Unverändert (geprüft mit `git diff --name-only 0c644a15..HEAD`):** `Makefile`, `harness/`, `Dockerfile`, `.golangci.yml`, `.a-check.yml`, `spec/`, `tools/`, `src/adapters/`, `src/bootstrap/m3_wurzel_test.go` und alle übrigen Dateien von `src/bootstrap/` außer den sechs des Blocks. Keine Direktive `//nolint`, kein `t.Skip` im neuen Code (`git diff 0c644a15..HEAD -- src` durchsucht), `git diff --check` ohne Befund.

**Größenmarke:** `git diff --numstat 0c644a15..HEAD` über die Pfade des Blocks: 1283 Einfügungen, 52 Löschungen, Summe 1335, wie berichtet, gegen die Marke 1800; Produktcode `replay.go` 27 und 8, `span.go` 44, `wurzel.go` 2 und 6; keine fremde Zeile (der Zweig ist der Handoff-Commit auf dem Start-Commit `b7604f22` von `main`).

**Sensoren dieses Laufs:** `make a-check`: 0 Befunde, Exit 0 (der Hinweis auf 97 ungeprüfte Dateien unter `tools/gegenproben` ist Bestand). `make gates`, `make test` und `make test-race` fuhr der Reviewer nicht (der Implementer berichtet sie grün auf `5a30a2e3`); die Abdeckung (100,0 % im Paket `replay`, 97,5 % in `src/bootstrap`) ist nicht nachgemessen.

---

## Eigene Mutationen

Ein eigener Treiber im Scratchpad (`telemetry-einbau-replay-reviewer/`, außerhalb des Repos, aus dem Treiber des Implementers abgeleitet, eigene Mutationsliste): Container `odata-recorder:race` ohne Netz und Mount, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`, nacheinander, nie neben `make gates`; je Mutation eine Wegwerfkopie des Baums und ein Prozess `go test -count=1` über das **ganze** Paket `replay` (rund 4 s, damit auch fremde Tests fangen), im Root mit den Tests dieses Slice per `-run`. `tools/harness/mutationslauf.sh` lief nicht (Mutationen mit Test in `src/bootstrap` halten die Zeitgrenze der Nullprobe nicht). Host mit 20 Prozessoren, Container `--cpus=4` (Go setzt `GOMAXPROCS` auf 4), für die Arme mit einem Prozessor `GOMAXPROCS=1` per Umgebung; Last des Hosts (1-Minuten-Wert, gelesen vom Treiber im Container je Prozess) im Mittel 5,6 bis 5,9, höchstens 7,1; Messender: der Treiber.

Zahlen: 54 Einträge; 5 davon gehören zu W13 und stehen mit ihren Raten unten; 2 sind äquivalent (kein `End` bei `ODR-E5003` und `ODR-E5005`: beide Wege enden vor dem Beginn des Spans); von den übrigen 47 sind **43 rot und 4 grün** (3 im Paket `replay`, 1 im Root). Mutationen, die die Hilfsdatei des Treibers (`mut_hilfen.go`) brauchten, färben `TestPaketOhneZustand` zusätzlich rot (Paketvariable und Typ); gezählt ist dort nur, was andere Tests rot machen, und die drei grünen Mutationen sind ohne Hilfsdatei wiederholt. Die Nullprobe (Baum ohne Änderung) ist grün: Paket `replay` 20 von 20 Prozessen je Arm, Root 1 von 1.

### Auflösung der Tabelle *Zusage · Fänger*

| Zeile | Eigene Mutationen (Zusage, Klasse) | Ergebnis |
|---|---|---|
| W1 | `StartSpan` nach `dienstFuer`, vor der Zuordnung (`x-w1-start-nach-dienst`) | rot: `TestKeinSpanOhneZuordnung` |
| W2 | `StartSpan` nach `bildeKey`, vor `ListBySession` (`s-start-nach-key`) | rot: `TestSpanFolge`, `TestSpanPlatzhalter`, `TestSpanJederRueckweg`, `TestKeinRequestWertImSpan` |
| W3 | kein `End` bei genau einem der 13 Codes (13 Mutationen, `w3-kein-end-bei-<Code>`); `End` nur, wenn `ctx.Err() == nil` (`s-end-nicht-bei-ctx-beendet`); `Set` nur, wenn sich ein Wert vom Platzhalter unterscheidet; `Set` nicht bei Fehler außerhalb von `f`; `Set` nicht bei Fehler in `f` | 11 rot aus `TestSpanJederRueckweg`, `TestSpanPlatzhalter`, `TestSpanNachDerAuswahl`, `TestSpanJedesPaar` (2 äquivalent); `Set`-Mutationen 3 rot; **`End` nur bei unbeendetem `ctx`: grün** (F-4) |
| W4 | `anderer_key` und `erschoepft` vertauscht; `ODR-E5007` und `ODR-E5002` je als `nicht_replayfaehig`; je Strategie (4) das Ergebnis `treffer` bei einem Fehler der Auswahl | 7 rot: `TestSpanJedesPaar` (die Mutation je Strategie auch `TestSpanGleichzeitig`) |
| W5 | Strategie der Konfig schon vor `Cursors.Get` (`c-strategie-vor-cursorget`) | rot: `TestSpanPlatzhalter` |
| W6 | Strategie des Cursors nur bei leerem `Consumed` (`c-consumed-konfig`), nur bei `Position` 0 (`c-position-konfig`), nicht bei `last` (`c-strategie-last-cursor-konfig`), fehlender Cursor ohne Strategie (`c-fehlender-cursor-platzhalter`) | **2 grün** (F-1), 2 rot: `TestSpanCursorStrategie`, `TestSpanJedesPaar` |
| W7 | `treffer` erst nach `Blobs.Open`; `nicht_gewaehlt` nach Fehler von `Open`; `nicht_gewaehlt` nach Fehler von `Put` | 3 rot: `TestSpanNachDerAuswahl` |
| W8 | `CountError` in `f` (`x-w8-counterror-in-f`) | rot: `TestSpanNurInF`, `TestNurDieAchtOperationen`, `TestTelemetrieImQuelltext` und weitere |
| W9 | abgeleitetes `ctx` an `Sessions.Get`, `Exchanges.ListBySession`, `WithWorkspaceTransaction`, `Workspaces.Get`, `Cursors.Get`, `Blobs.Open`, `Cursors.Put` (7 Mutationen, je Stelle eine) | 7 rot: `TestKontextJederStelle` |
| W10 | zweiter Adapter nur im Modus `hybrid`; zweiter Adapter nur im Modus `replay` | 2 rot: `TestWurzelGesamtspanJedeAnfrage`, `TestWurzelReplayMatchLiegtImGesamtspan`, `TestM4WahlJeModus` und `TestWurzelSpanZeile` (hybrid), `TestEinAdapterJeProzessMitDerSenke` und `TestWerteErreichenVerbraucher` (replay) |
| W11 | `CountError(ODR-E5001)` bei `kein_treffer` (2 Einträge: Paket, Root) | im Paket rot (7 Tests, darunter `TestKeinZaehlerImDienst`), **im Root grün** (F-3) |
| W12 | `recorder.match_key` als drittes Attribut von `Set`; `recorder.path` in `StartSpan`; `key` an das Ergebnis angehängt | 3 rot, nur die zweite aus `TestKeinRequestWertImSpan`, die beiden anderen aus den Tests mit Literal (`TestSpanJedesPaar` und weitere) |
| W13 | siehe *Raten* | siehe *Raten* |

### Raten von W13 (Treffpunkt vor und nach)

Je Prozess `go test -count=1 -run ^TestSpanGleichzeitig$` (20 Läufe zu 16 Goroutinen), 20 Prozesse je Arm; „rot“ heißt: der Test wird rot (bei den Armen mit Detektor: Wirkung oder Meldung `DATA RACE`).

| Mutation | geteilter Zustand | `GOMAXPROCS` 4, ohne Detektor | `GOMAXPROCS` 1, ohne Detektor | `GOMAXPROCS` 4, mit Detektor | `GOMAXPROCS` 1, mit Detektor |
|---|---|---|---|---|---|
| `w13v-atomar-ergebnis-in-f` | in `f` geschrieben, nach dem Treffpunkt gelesen (atomar) | 20 von 20 | 20 von 20 | 20 von 20 | nicht gefahren |
| `w13n-atomar-ergebnis` | nach dem Treffpunkt geschrieben und gelesen (atomar) | **6 von 20** | **0 von 20** | 20 von 20 | **0 von 20** |
| `w13n-atomar-strategie` | wie vorige, Strategie | **8 von 20** | **0 von 20** | nicht gefahren | nicht gefahren |
| `w13n-plain-ergebnis` | nach dem Treffpunkt, nicht atomar | nicht gefahren | nicht gefahren | 20 von 20 (20 Meldungen `DATA RACE`) | 20 von 20 (20 Meldungen) |
| `w13n-plain-strategie` | wie vorige, Strategie | nicht gefahren | nicht gefahren | 20 von 20 (20 Meldungen) | nicht gefahren |
| Nullprobe | — | 0 von 20 | 0 von 20 | 0 von 20 | nicht gefahren |

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Menge M-Cursor führt Cursor nur mit leerem `Consumed` und `Position` 0 und nennt die Achse Cursor-Zustand nicht. Zwei eigene Mutationen, die im Span die Strategie der Konfig statt der des Cursors zeigen, sobald der Cursor `Consumed` trägt (`c-consumed-konfig`) oder `Position` ungleich 0 hat (`c-position-konfig`), bleiben im ganzen Paket grün (beide in 1 von 1 Prozess, deterministisch). Der Fall, den die Zusage trägt, ist ein Cursor nach dem ersten Aufruf bei geänderter `REPLAY_STRATEGY`; die Fälle mit `Consumed` (`erschoepft`) setzen die Strategie der Konfig gleich der des Cursors. | `AGENTS.md` §3.15, §3.17, `SPEC-296` Nr. 4 a und 7 d | `src/hexagon/services/replay/telemetrie_span_test.go` · `func TestSpanCursorStrategie`, `u.cursor(strat)` ohne `Consumed`; Plan §3 *Mengen* · `M-Cursor (W6)` | ja — `make test` mit den zwei Mutationen | Fänger führt nur eine Form der Eingabe: Menge ohne Achse Cursor-Zustand |
| F-2 | MEDIUM | Das akzeptierte Negativ (7) nennt als Maß „3 von 3 Mutationen der Zeile W13 liegen vor dem Treffpunkt“, aber nicht die Rate der ungefangenen Klasse. Gemessen fängt der Wirkungs-Fänger geteilten Zustand, der allein nach dem Treffpunkt geschrieben und gelesen wird, ohne Detektor in 6 und 8 von 20 Prozessen (4 Prozessoren) und in 0 von 20 (1 Prozessor), mit Detektor ohne Rennen in 20 von 20 (4 Prozessoren) und 0 von 20 (1 Prozessor); der Satz „der Wirkungs-Fänger fängt es nicht“ stimmt für 1 Prozessor, nicht für 4. Der Satz „die Rate hängt nicht an der Prozessorzahl“ gilt für Mutationen vor dem Treffpunkt (20 von 20 in beiden Armen) und für die Klasse nach ihm nicht. | `AGENTS.md` §3.10, §3.15 | Plan §3 · `(7) Der Treffpunkt von W13 liegt am Ende der Transaktion`; Plan §3 · `Die Rate hängt nicht an der Prozessorzahl, weil der Treffpunkt am Ende jeder Transaktion` | ja — `go test -count=1 -run ^TestSpanGleichzeitig$` mit `w13n-atomar-ergebnis`, 20 Prozesse je Arm | Akzeptiertes Negativ ohne Maß der ungefangenen Klasse; Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt |
| F-3 | LOW | Die Zeile der Zusatztabelle für den Modus `hybrid` (Replay-Versuch mit `ODR-E5001`, dann Fallback: „der Fehler zählt nicht“) nennt W10 als Fänger. `TestWurzelReplayMatchLiegtImGesamtspan` liest den Zähler je Fehlercode nicht: die Mutation `CountError(ODR-E5001)` bei `kein_treffer` bleibt dort grün (1 von 1 Prozess); die Zusage hält `TestKeinZaehlerImDienst` (rot in 7 Tests des Pakets). | `AGENTS.md` §3.15 | Plan §3 *Zusatztabelle* · `Replay-Versuch mit`; `src/bootstrap/telemetrie_replay_test.go` · `func TestWurzelReplayMatchLiegtImGesamtspan` | ja — Mutation mit `-run TestWurzelReplayMatchLiegtImGesamtspan` | Fänger-Zuordnung der Zusatztabelle nennt den Test, der die Zusage nicht prüft |
| F-4 | LOW | Die Mutation `End` nur bei `ctx.Err() == nil` bleibt im ganzen Paket grün (1 von 1 Prozess). Die Fänger der Zusage „`End` wirkt auch bei beendetem `ctx`“ speisen den Kontextfehler als Rückgabe eines Fakes ein; `u.replay` läuft in allen Span-Tests mit `context.Background()`, kein Test beendet das `ctx` des Aufrufs und zählt dann `End`. Der Code verzweigt nicht nach dem Zustand von `ctx`. | `AGENTS.md` §3.15, `SPEC-296` Nr. 3 f und 9 b | `src/hexagon/services/replay/telemetrie_span_test.go` · `func kontextAn`; `src/hexagon/services/replay/telemetrie_hilfen_test.go` · `resp, err := u.replay(context.Background(), a)` | ja — Mutation `s-end-nicht-bei-ctx-beendet` | Abbruch nur als Rückgabe eines Ports nachgebildet, kein beendetes `ctx` |
| F-5 | LOW | Die Zeile `M-Stellen-Replay (W12)` der Tabelle *Mengen mit Zählprüfung* nennt „gleich 18“; der Test und der Absatz *Abweichungen* (c) in §6 nennen 46 Elemente (7 Stellen des Requests und die Ursache an 8 Ports in 2 Formen, je in 2 Arten). Die Zeilen W3, W5, W7 der Mengen-Tabelle sind auf 52, 26, 12 nachgezogen, diese eine nicht. | `AGENTS.md` §3.9 | Plan §3 · `| M-Stellen-Replay (W12) | 8 Stellen, die achte in zwei Formen, × 2 Arten gleich 18 |` | nein — `make docs-check` prüft das nicht; Handlesung | Plan-Nachzug unvollständig: eine Zählprüfung der Mengen-Tabelle |
| F-6 | LOW | Die Menge der 46 Elemente in `TestKeinRequestWertImSpan` zählt für die Art *Treffer* 16 Läufe der Ursache (8 Ports × 2 Formen), die dieselbe Szene fahren: das Closure ignoriert `p.op` und speist den Text am fehlenden Cursor (S6) ein; 15 davon wiederholen sie. Verschieden sind 31 Elemente (14 des Requests, 16 der Ursache mit Fehler, 1 der Ursache mit Treffer). Die Aussage „an jedem der 8 Ports in 2 Formen × 2 Arten“ gilt für die Art *Fehler*; die Art *Treffer* gibt es strukturell nur an S6. | `AGENTS.md` §3.11, §3.15 | `src/hexagon/services/replay/telemetrie_dienst_test.go` · `stellen["Ursache "+p.name+" "+form] = func(u *umg, k string, treffer bool)` | nein | Zählprüfung zählt Wiederholungen derselben Szene |
| F-7 | INFO | Das akzeptierte Negativ (6) nennt als Grund, der Text der Ursache eines Fehlers eines Ports sei über HTTP nicht erreichbar. Der Haken `NeuerReplay` des Root-Tests (`w.replay = &k`) erlaubt, einen Port mit Fehler und Canary im Text vor dem Dienst einzusetzen; das Maß „1 von 8 Stellen an der Senke“ stimmt. Welche Stelle fehlt und warum: bestätigt (die achte, `Text der Ursache`). | `AGENTS.md` §3.15 | Plan §3 · `(6) Die Senke des echten Adapters prüft der Root-Test für 7 der 8 Stellen` | nein | Grund eines akzeptierten Negativs enger gefasst als die Lage |
| F-8 | INFO | Die Zelle *Wiederholung* des Rasters (`SPEC-296` Nr. 7 c: zwei Aufrufe mit `sequential` ergeben zwei Spans mit `treffer`, dann `erschoepft`) hat keinen Test mit zwei Aufrufen am selben `Replayer` und denselben Stores; am nächsten kommt W13 mit 16 verschiedenen Cursorn. Eine Mutation, die der zweite Aufruf allein zeigte, braucht Zustand im Dienst, den `TestPaketOhneZustand` ausschließt; deshalb keine Mutation gefunden, die grün bliebe. | `AGENTS.md` §3.16 | `spec/spezifikation.md` · `SPEC-296` Nr. 7 c; Plan §6 *Raster* · `Wiederholung` | nein | Zelle des Rasters ohne eigenen Fall, durch Bauart gedeckt |
| F-9 | INFO | `ergebnisVon` liest `f.Code` nach `errors.As` ohne Prüfung auf einen nil-Zeiger `*inbound.Fehler`. Erreichbar ist das nicht: der Eingang ist der Fehler von `waehle`, der allein aus `mitCode` stammt; ein Kontextfehler oder ein Fehler eines Ports erreicht die Funktion nicht, der Span prüft sonst nie einen Fehlerwert. | Maintainability | `src/hexagon/services/replay/span.go` · `switch f.Code {` | nein | Zeiger-Dereferenzierung ohne Prüfung, unerreichbar |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/services/replay/replay.go`, `span.go` (Produktcode) | geprüft, ohne Befund: Beginn unmittelbar nach der Zuordnung und vor `Sessions.Get`, `End` mit `defer` genau einmal je Rückweg nach dem Beginn und nach `Log`, ein `Set` nach der Rückkehr der Transaktion, auch wenn `f` nie lief, kein Aufruf von `Telemetrie` in `f`, `ctx` des Aufrufs an `StartSpan` und jede Stelle unverändert (7 eigene Mutationen und `StartSpan` durch `TestKontextJederStelle`), `CountError` nirgends, im Span nur Werte aus Konfig, Cursor und Auswahl; Mengen und Werte stimmen mit `SPEC-296` Nr. 3 bis 6 überein, die 15 von 20 Paaren stimmen mit `waehle` (die fünf fehlenden Paare sind unmöglich: `anderer_key` allein bei `strict`, `erschoepft` bei `first` und `last` nie). |
| Kontextfehler, beide Arten, auch in `errors.Join` | geprüft, ohne Befund: der Span prüft keinen Fehlerwert außer dem der Auswahl; `End` und `Set` hängen nicht an der Art; je 8 Stellen × 2 Arten sowie das Paar mit Fehler der Tabelle laufen in `TestSpanJederRueckweg` (52), `TestSpanPlatzhalter` (26) und `TestSpanNachDerAuswahl` (12); die Zählungen 52, 26, 12 und 46 sind am Test nachgezählt. |
| Rückwege von `Sessions.Get` und `Blobs.Open`, jeder der 13 Codes | geprüft, ohne Befund: `End` bei jedem der 11 Codes mit Span je Code ausgeschaltet, rot (13 Mutationen, 2 äquivalent). |
| `src/bootstrap/wurzel.go`, Adapter je Prozess in `replay` und `hybrid` | geprüft, ohne Befund: `baueReplay` erhält `h.tel`; `grep` nach `NeueTelemetrie`, `alleTel`, `len(w.adapter)`, `io.Discard`, `"telemetrie":` und `.adapter` über `src/`, `tests/`, `harness/` findet keine weitere Stelle, die zwei Adapter festhält; die Zahl zwei stand an den sechs genannten Stellen in vier Dateien, alle sind auf 1 umgezogen; `m3_wurzel_test.go` und alle weiteren Dateien von `src/bootstrap/` sind unverändert, `m3` zählt im Modus `mock` einen. Ein zweiter Adapter nur in `hybrid` und nur in `replay` ist rot (2 Mutationen). Die Schleifen `1-min(i, 1)` in `telemetrie_wurzel_test.go` und `telemetrie_mutation_test.go` sind mit einem Adapter weiter richtig. |
| Zeile im Gesamtspan (W10) | geprüft, ohne Befund: `TestWurzelReplayMatchLiegtImGesamtspan` prüft die Zeile als Literal mit `parent_id` gleich `span_id` des Gesamtspans, der nach ihr steht, `Dropped` 0 und den Adapter des Dienstes gleich dem des Prozesses, in `replay` und `hybrid` (`kein_treffer`, ein Aufruf des Upstream). |
| Kein Wert des Requests in einer Senke (W12) | geprüft, ohne Befund: Canary aus Bytewerten (`kanarie()`), 7 Stellen an der Senke des echten Adapters, 8 Stellen am Fake, 3 eigene Mutationen mit Wert des Requests im Attribut rot; die Ursache am Port wird an allen 8 Ports in zwei Formen geführt (F-6, F-7). |
| Abweichung (a), Treffpunkt von W13 | geprüft: die Begründung stimmt. Der Fake reiht die Transaktionen in einer Schlange (`einreihen`, `SPEC-108` Nr. 1), eine Schranke mit 16 Teilnehmern in `f` oder an `Blobs.Open` würde nie voll; ein Treffpunkt am Ende jeder Transaktion hält alle 16 an, bevor einer `Set` ruft, und fängt Zustand, der vor ihm geschrieben wird, in 20 von 20 Prozessen bei 4 und bei 1 Prozessor. Das Fenster nach ihm: F-2. |
| Abweichung (b), Senke des echten Adapters | geprüft, ohne Befund: 7 von 8 Stellen im Root-Test, 8 von 8 am Fake; fehlende Stelle benannt; Grund enger gefasst als die Lage (F-7). |
| Abweichung (c), verbreiterte Mengen | geprüft, ohne Befund: je Art des Kontextfehlers und je Port sind die Mengen der Zeilen W3, W5, W7, W11 und W12 Obermengen des Wortlauts des Architect; Adresse und Ausgang *offen* stehen in §6 (`AGENTS.md` §3.13); die Zeilen der Tabelle behalten den alten Wortlaut bis zur Antwort. |
| Akzeptierte Negative (1) bis (5) | geprüft, ohne Befund: Grund, Folge und Maß je Negativ vorhanden; (7) siehe F-2, (6) siehe F-7. |
| Plan-Nachzug (`AGENTS.md` §3.9) | geprüft: Kopf, §3, §4 (Block, Messpunkt 1335), §6 nachgezogen; die Zahl „sechs Stellen in vier Dateien“ stimmt am Quelltext; eine Zeile bleibt (F-5). |
| `make a-check` | Exit 0, 0 Befunde. |
| Kommentare (`AGENTS.md` §3.7, §3.11) | geprüft, ohne Befund: die Kommentare in `replay.go`, `span.go` und `wurzel.go` tragen Zusage, Kopplung oder Rang-Zeiger im Indikativ; kein Kommentar beschreibt eine verworfene Alternative. |
| Handbuch, Kataloge | geprüft, ohne Befund: kein Fehlercode, Schlüssel oder Kommando dazu; keine Datei von `docs/user` und nicht `harness/katalog-ausnahmen.txt` im Diff. |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 4 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe: Menge ohne Achse Cursor-Zustand · Akzeptiertes Negativ ohne Maß der ungefangenen Klasse; Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt · Fänger-Zuordnung der Zusatztabelle nennt den Test, der die Zusage nicht prüft · Abbruch nur als Rückgabe eines Ports nachgebildet · Plan-Nachzug unvollständig · Zählprüfung zählt Wiederholungen derselben Szene

## Verdikt

**Merge-blockierend:** ja — F-1 und F-2 sind MEDIUM; beide betreffen Tests und Plan, kein Befund trifft den Produktcode (4 von 47 eigenen Mutationen grün, davon 1 nur im Root, kein HIGH). Die Entscheidungen der Abweichungen (a) bis (c) bleiben beim Architect (Ausgang *offen*, Adresse im Plan).

**Übergabe:** Findings gehen an den Implementer; die **Finding-Klassen** gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. Der Report heißt `2026-11-23-slice-telemetry-einbau-replay-review` (Kennung für Plan und Evidence, ohne Pfad und Endung). Dieser Report ersetzt keine Verifikation — DoD- und Spec-Konformität prüft der Verifier.
