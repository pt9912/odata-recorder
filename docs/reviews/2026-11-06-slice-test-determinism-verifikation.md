# Verifikationsbericht: slice-test-determinism - 2026-11-06

**Rolle:** Verifier (Modul 11) - **Frage:** Bauen wir es richtig? - gegen Plan, DoD und Vertrag `SPEC-038`; nicht Reviewer (Diff), nicht Validator.

**Gegenstand:** Zweig `slice-test-determinism`, HEAD `492720d`, Diff `01c547c..492720d` (7 Dateien, 845 Einfuegungen). Review: `docs/reviews/2026-11-06-slice-test-determinism-review.md` (HEAD `f075c69`); Fix-Runde `5b6a9f2`/`65906cd`; Architect `492720d`. Bezug: `LH-QA-02`, `LH-FA-04`, `LH-FA-09.b`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md).

**Methode.** Einmal gefahren, keine Wiederholungsschleife. Belege des Implementers nicht uebernommen: `make gates` selbst gefahren; Mutationen auf Wegwerfkopie im Scratchpad (`ver-td/mut`), je Mutation ein Prozess, Build und Lauf mit `--memory 2g`, unter `flock /tmp/odata-recorder-heavy.lock`, ein schwerer Lauf zugleich; Image `odata-recorder:vertd` danach geloescht. Alle Mutationen am Test oder an `src/` sind deterministisch (ein Prozess, nacheinander gesendete Requests); gemessen hat der Verifier.

## Ergebnis in Kurzform

HIGH: keine. MEDIUM: 1 (V-1). LOW: 1 (V-2). INFO: 6. **Empfehlung: Closure nein, bis V-1 behoben ist** (Fix klein, nur Fall in `TestDeterminismusVergleich`); V-2 darf als benanntes Negativ stehen bleiben.

## DoD und Entscheidungen

| Punkt | Befund |
|---|---|
| `make gates` | gruen, Exit 0 (selbst gefahren im Worktree; `groessenmarke slice-test-determinism.md: ok, Messung 668 von Marke 1800`) |
| Ziel `make test-determinism` liegt daneben | ja: Fragment `harness/mk/test-determinism.mk`, kein Eintrag in `GATE_CHECKS` (`AGENTS.md` §3.6); `Makefile`, `go-gates.mk`, `Dockerfile`, `src/`, `tests/e2e/uebergang/` im Diff unberuehrt |
| Entscheidung Gate-Kette | nirgends als entschieden dargestellt: Plan §6 *offen*, Adresse User; `SPEC-038` Nr. 2 weist sie dem Betreiber zu; README-Zeile, Fragment-Kopf und Commit-Texte nennen nur "die eine Zeile" |
| `make test` und `make test-race` fahren `TestDeterminismus*` ohne Binary | ja. `pruefung` nimmt `./...`; Lauf in den Images `test` und `race`: `TestDeterminismusVergleich` PASS, `TestDeterminismusStand` PASS, `TestE2EDeterminismus` SKIP (kein `E2E_BINARY`) |
| Messung 668 | stimmt: `git diff --numstat 01c547c..492720d` = 11 + 584 + 73 |
| Rohbytes | `LC_ALL=C grep -P` auf Escape-Folgen, BOM, NBSP in Test, Treiber, Fragment: 0 Treffer |
| DoD-Haken, Closure-Notiz, Register, Risiko-Ausgaenge | offen, wie vor der Closure zu erwarten; §6 *Laufzeit* steht auf "offen bis zur Closure" (gemessen 0,74 bis 0,93 s) und braucht dort seinen Ausgang; *Gate-Kette* und *Formelkonflikt* bleiben beim User |
| Roadmap, Welle, Register | nicht angefasst |

## Tabelle gegen Vertrag

Alle 15 Zeilen der Tabelle *Zusage - Faenger* decken Nr. 1 bis 10 von `SPEC-038`; Zeilen 14 und 15 sind Negative mit Grund und Folge (Nr. 10 f, a, b, e). Nr. 8 f (Herunterfahren), "nur lesend" und `k.VolatileHeaders` stehen als akzeptierte Negative mit Grund und Folge (Plan §3 und `SPEC-038` Nr. 6). Nr. 9 (Mock) ist ausgeschlossen und adressiert nicht. Keine Zusage ohne Zeile gefunden ausser V-2 (Teilzusage).

## Probe-Mutationen (Klasse, nicht Befund)

Elemente, die weder Reviewer noch Implementer nannten. "rot" = Mutation gefangen.

| Nr. | Menge / Zusage | Mutation | Ergebnis |
|---|---|---|---|
| T1 | Kontrollwert `Position` groesser 0, Rand 0 | `Position <= 0` zu `< 0` | rot (`Position 0 bei gefuelltem Consumed: Leser gilt als gueltig`) |
| T2 | Status verschieden | Vergleich ueber `status/10` | rot |
| T7 | Stand: jeder Cursor | letzter Cursor nicht verglichen | rot (7 Felder an Cursor 1) |
| E1 | Reset in jedem Lauf | Reset entfaellt in Lauf 2, 73, 99 | rot, je `Lauf <i>: Revision <i-1>, erwartet <i>` |
| E2 | Reset setzt Cursor zurueck, in jedem Lauf | `ResetAll` ab der 60. Rufung ohne Wirkung (`src/`) | rot ab Lauf 8: `Stand nach Reset: Cursor 0: Position 5, erwartet 0` |
| E3 | Fehlerantworten (Request 9, 10) gleich in allen Laeufen | Header `X-Probe` im Fehlerpfad (`fehler.go`) nur bei der 143. und 200. Antwort (Lauf 72 und 100) | rot, `Laeufe: 100, abweichend: 2` |
| E4 | Ablage geschlossen | `Close` entfaellt nur in Lauf 50 | rot (`4 nach Lauf 2, 8 nach Lauf 100`) |
| E5 | dasselbe | `Close` entfaellt nur in Lauf 100 | rot (`4`, `7`) |
| E6b | Orakel Requests 1 bis 8 gleich Aufnahme | Replay liefert fuer Request 3 und 5 den Body `zz` (nur Replay-Requests, `src/`) | rot: `Lauf 1: Orakel, Request 3` und `Request 5` |
| **T3** | Body verschieden (Laenge oder ein Byte) | Vergleich ueber `strings.TrimSpace` | **gruen** |
| **T4** | dasselbe | Vergleich ueber `strings.EqualFold` | **gruen** |
| **T5** | Bytes der Werte eines Namens | Werte ueber `strings.EqualFold` | **gruen** |
| **T6** | dasselbe | Werte ueber `TrimSpace` | **gruen** |
| **E7** | Ablage "geschlossen nach jedem Lesen" | `leseStand` haelt eine Verbindung offen (nie `Close`, wiederverwendet) | **gruen** |

Treiber-Handlaeufe nachvollzogen (Kopie des Treibers mit anderem Image-Namen, Kontext die Wegwerfkopie): Argument `x`, `""`, `a b`: `Aufruf:`, Exit 2, kein Container; `DOCKER=false`: `Aufbau der Stufe integration gescheitert`, Exit 2; normal: gruen, `Laeufe: 100, abweichend: 0`; `t.Skip` im Test: `test-determinism: rot (SKIP)`; `--network none` entfernt: bleibt gruen (akzeptiertes Negativ Nr. 10 f); SIGTERM nach 1 s: Exit 2, kein Container bleibt.

## Befunde

### V-1 (MEDIUM): Die Klasse "Body verschieden" und "Bytes der Werte" ist nur an einer Bitaenderung und am Kopf gedeckt

- `quelle`: `AGENTS.md` §3.15, §3.17; `SPEC-038` Nr. 5 (*Body verschieden (Laenge oder ein Byte)*, *Bytes der Werte eines Namens verschieden*)
- `pfad`: `tests/e2e/determinismus_test.go`, `vergleicheAntworten` und die Faelle in `TestDeterminismusVergleich`
- `befund`: Die Body-Faelle aendern ein Byte per `^= 1` (erste, mittlere, letzte Stelle), kuerzen vorn um ein Byte oder leeren. Die Header-Faelle haengen `x` an oder leeren. Ein Vergleich, der Leerraum abschneidet (T3, T6) oder Gross und Klein gleichsetzt (T4, T5), bleibt gruen (je 1 von 1, deterministisch). Ein abschliessender Zeilenumbruch oder ein Gross/Klein-Wechsel in einer Antwort wuerde so als gleich gelten. Die Menge, die die Zusage meint (jede Byte-Abweichung), ist nur mit Bit 0 und Anhaengen belegt; die Fixture-Werte (`{"Id":1}`, Buchstaben, kein Leerraum) lassen Leerraum und Gross/Klein nicht entstehen.
- `klasse`: Fixture gleicht dem Wert der Mutation nicht, aber die Mutation liegt ausserhalb der erzeugten Abweichungen (§3.15, §3.17); gleiche Bauart wie Review M-3, andere Elemente.
- Fix gegen die Klasse: zusaetzliche Faelle je Art (angehaengter Zeilenumbruch, vorangestellter Leerraum, Gross/Klein-Wechsel eines Buchstabens, Leerraum im Wert eines Headers), in die vorhandene Schleife; `len(faelle) != 30` zaehlt neu.

### V-2 (LOW): "Geschlossen nach jedem Lesen" ist nur als Nicht-Wachstum gefangen

- `quelle`: `AGENTS.md` §3.11, §3.15; `SPEC-038` Nr. 6 (*geschlossen nach jedem Lesen*)
- `befund`: `offeneAblagen` vergleicht die Zahl nach Lauf 2 mit der nach Lauf 100 (beide groesser 0). Eine Verbindung, die ueber alle Laeufe offen bleibt (E7), verandert die Zahl nicht und ist gruen. Plan Zeile 10 (Nachtrag) und die Meldung ("Lesen schliesst jede Verbindung") sagen mehr zu, als der Fanger haelt. Folge im Betrieb: die Ablage ist waehrend des Servers eine zweite Verbindung; der Vertrag verlangt sie untaetig/geschlossen, der Fall "dauerhaft offen" bleibt ungefangen. Weder Plan noch Vertrag nennen es als Negativ.
- Fix: entweder die Zusage auf "kein Wachstum der offenen Dateien" einschraenken (Vertrag Nr. 6, Plan, Meldung) oder als benanntes Negativ mit Grund und Folge fuehren.

### INFO

- I-1: Fehlerantworten (Request 9 und 10) sind gedeckt (E3, an Positionen, die niemand nannte). Reset-Positionen 2, 73, 99 und eine Wirkungsabnahme ab Lauf 8 (E2) sind rot. Orakel Request 3 und 5 sind rot (E6b); eine Mutation, die den Lauf nur im Test-Prozess aendert, war in der Wegwerfkopie nicht noetig, da das Binary im Modus `record` dieselben Quellen baut: eine erste Variante (E6) brach daher die Aufnahme und gilt nicht; E6b bedingt die Aenderung auf den Steuerheader `X-Recorder-Replay-Session`.
- I-2: Der E2E-Test fuehrt per Kontrollwert genau einen Cursor; ein `ResetAll`, das nur einen von mehreren Cursors zuruecksetzt, kann er nicht fangen (nicht gemessen, aus Code gelesen). Das gehoert zu `SPEC-129`, nicht zu diesem Vertrag.
- I-3: Das Handle wird im Test zweimal heruntergefahren (Zeile im Cleanup und nach Lauf 100); der zweite Aufruf ist grueen Laeufen nach idempotent. Nr. 8 f bleibt Negativ ohne Fanger (Plan §3 *Negative und Rueckgabe* a), Grund und Folge stehen dort und in `SPEC-038` Nr. 10.
- I-4: Der Wortlaut zu `LISTEN` (Nr. 3, Platzhalter `127.0.0.1:8080`, dann `Listen` auf Port 0) stimmt mit dem Code ueberein.
- I-5: Ohne Binary laufen die beiden Vergleichstests in `make test` und `make test-race`; `TestE2EDeterminismus` ueberspringt dort und wird nur vom Treiber als gruen gezaehlt.
- I-6: Der Plan nennt fuer Zeile 10 die Zahlen `11` und `403` offene Dateien; meine kleinere Leck-Mutation (E4, E5) zeigt `4` zu `8` bzw. `7`; Abweichung der Zahl, nicht der Wirkung (andere Mutation).

## Nicht aufgeloest

- Nr. 8 f (Fehler beim Herunterfahren): keinen Fanger moeglich ohne Produktcode oder Fake; akzeptiertes Negativ, nicht nachgefahren.
- "nur lesend": Zusage der Testseite, nicht erzwungen (akzeptiertes Negativ, gelesen).
- `k.VolatileHeaders` mit fester Liste gleich dem Default bleibt gruen (akzeptiertes Negativ, nicht erneut gemessen; Reviewer: 10 von 10 rot fuer `nil`).
- Strategien, Binary, Admin-Reset, Mock: ausserhalb des Vertrags (Nr. 9, 10).
- `--network none`, `--memory`, `--read-only` entfernt: nur `--network none` nachgefahren (gruen, wie benannt).
- Formelkonflikt der Groessenmarke und Gate-Kette: Entscheidung beim User.
