# Review-Bericht: slice-test-determinism — 2026-11-06

**Rolle:** Reviewer (Modul 10) · **Frage:** Ist der Diff gegen Plan, Entscheidungen und Hard Rules wartbar und richtig? — nicht gegen die DoD (Verifier).

**Gegenstand:** `slice-test-determinism`, Zweig `slice-test-determinism`, HEAD `f075c69`, Diff `01c547c..f075c69` (6 Dateien, 652 Einfügungen, 6 Löschungen). Bezug: `LH-QA-02`, `LH-FA-04`, `LH-FA-09.b`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md).

**Geprüft gegen:** Plan `slice-test-determinism` (§3 Tabelle mit 15 Zeilen und Mutationsabsatz, §4 Messung 613 gegen Marke 1800, §6 mit Raster), Vertrag `SPEC-038` Nr. 1 bis 10, `AGENTS.md` §3.6, §3.7, §3.10, §3.11, §3.15, §3.16, §3.17, §3.18.

**Methode.** Einmal gefahren (Durchstich vor Härtung, keine Wiederholungsschleife). Der Bericht des Implementers ist nicht übernommen. Selbst gemessen: Docker nur unter `flock /tmp/odata-recorder-heavy.lock`, `--memory 4g` im Aufbau und `2g` im Lauf, je Mutation eine Wegwerfkopie im Scratchpad (`rev-td`), ein Image, `-test.count=1`. Alle Mutationen sind deterministisch (ein Prozess, nacheinander gesendete Requests); gemessen hat der Reviewer.

**Grenze dieses Laufs (ehrlich).** Nicht jede Tabellenzeile ist nachgefahren. Mutiert habe ich 15 Stellen, bewusst an Elementen, die Plan und Implementer nicht nannten (Tabelle unten). Gelesen, nicht mutiert: die Zeilen 1, 2, 4, 5 (Orakel gegen Fehlercodes), 7, 8, 9, 12 und alle Mutationen der Zeile 10 am Standvergleich (Felder einzeln); ihre Mutationen stehen im Plan und sind hier nicht wiederholt. `make gates` nicht gefahren (Aufgabe des Verifiers). Unmutiert habe ich `make test-determinism` einmal gesehen: grün, `Läufe: 100, abweichend: 0`.

| Nr. | Mutation (Wegwerfkopie) | Ergebnis |
|---|---|---|
| A1 | `pruefeKontrollwert` ignoriert `Position` allein | grün (1 von 1) |
| A2 | `pruefeKontrollwert` ignoriert `Consumed` allein | grün |
| A3 | `pruefeStandNachReset` prüft `Position` nur ab Cursor 1 | grün |
| A4 | `pruefeStandNachReset` prüft `Consumed` nur ab Cursor 1 | grün |
| A5 | Vergleich der Werte eines Namens über `strings.Join(w, ",")` | grün |
| B1 | Aufrufstelle `k.VolatileHeaders` durch `nil` ersetzt | rot, 10 von 10 Prozessen (`X-Correlation-Id`, `abweichend: 99`) |
| B2 | `aufnahmeReplay` liefert 6 statt 7 Requests (Aufnahme 7 statt 8 Austausche) | grün |
| B3 | Replay setzt `X-Probe` nur in Lauf 100 Request 1 (Zähler 793), nur Lauf 100 Request 8 (800), nur Lauf 50 Request 5 (397) | rot, je `abweichend: 1` |
| B4 | `X-Probe` nur in Lauf 2 Request 1 | rot |
| B5 | `leseStand` schließt die Ablage nicht | grün |
| T | Treiber mit Argument `x` und `""` | `Aufruf:`, Exit 2; SIGTERM nach 1,5 s: Exit 2, kein Container bleibt |

---

## Befunde

### Findings nach Schwere

**HIGH: keine.** Kein ADR-Verstoß, kein Produktcode berührt (`src/` unverändert), kein Gate gelockert: `GATE_CHECKS` ist unverändert, das Fragment trägt keinen Eintrag, das Root-Makefile ist im Diff nicht berührt (`AGENTS.md` §3.6).

#### M-1 (MEDIUM): Der Kontrollwert des Lesers hat nur einen Fänger für beide Bedingungen zugleich

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe), §3.17; `SPEC-038` Nr. 6 (Kontrollwert: `Position` größer 0 und nichtleeres `Consumed`)
- `pfad`: `tests/e2e/determinismus_test.go` · `case c.Position <= 0 || len(c.Consumed) == 0:` und der Fall `"Position 0": {func() model.ReplayCursor { c := gut()[0]; c.Position, c.Consumed = 0, nil; return c }()}`
- `befund`: Der Fall `Position 0` setzt `Position` und `Consumed` zugleich zurück. Entfällt eine der beiden Bedingungen allein (A1: `Position` ignoriert, A2: `Consumed` ignoriert), bleibt `TestDeterminismusStand/Kontrollwert` grün (je 1 von 1); der Plan nennt als Mutation nur „ignoriert `Position` und `Consumed`“ zusammen. Ein Leser, der `Consumed` nicht liest (leer, `Position` gesetzt) oder `Position` nicht liest, gälte als gültig.
- `verifizierbar`: ja (A1, A2)
- `klasse`: Fänger ändert zwei Felder zugleich, Mutation an einem Feld bleibt grün

#### M-2 (MEDIUM): Der Stand nach Reset prüft `Position` und `Consumed` in den Fängern nur am zweiten Cursor

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.16 (Raster Vergleich der Stände, je Cursor); `SPEC-038` Nr. 6 (jeder Cursor mit `Position` 0 und leerem `Consumed`)
- `pfad`: `tests/e2e/determinismus_test.go` · `"Position": func(s []model.ReplayCursor) []model.ReplayCursor { s[1].Position = 1; return s }` und `"Consumed": … s[1].Consumed = []int64{1}`
- `befund`: Beide Verstöße ändern nur Cursor 1 der Basis (`a`, `b`); Cursor 0 trägt nur die Identitäts-Verstöße. Die Mutation „`Position` wird nur ab Cursor 1 geprüft“ (A3) und die gleiche für `Consumed` (A4) bleiben grün. Das `TestE2EDeterminismus` hat nur einen Cursor und kann es nicht fangen (`Position` und `Consumed` am einzigen Cursor 0 prüft die Schleife, der Fall ist dort aber nicht von Cursor 1 zu trennen).
- `verifizierbar`: ja (A3, A4)
- `klasse`: Allaussage „jeder Cursor“ als Fänger an einem Element der Menge

#### M-3 (MEDIUM): Die Zahl der Werte eines Headers ist nicht von ihrer Verkettung getrennt

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.17; `SPEC-038` Nr. 5 (*Zahl, Reihenfolge oder Bytes der Werte eines Namens verschieden*)
- `pfad`: `tests/e2e/determinismus_test.go` · `case !slices.Equal(s[n], w):`
- `befund`: Mutation A5 (Vergleich der Werte als `strings.Join(.., ",")`) bleibt grün: zwei Zeilen `Vary: Accept` / `Vary: Prefer` gegen eine Zeile `Vary: Accept,Prefer` gelten dann als gleich. Die Fälle der Probe (zusätzlicher Wert, fehlender Wert, Reihenfolge, Bytes) ändern die verkettete Zeichenfolge immer mit; die Zusage „Zahl der Werte“ (die Rahmung der Antwort) hat keinen Fall, in dem die Verkettung gleich bleibt.
- `verifizierbar`: ja (A5)
- `klasse`: Fänger ändert Werte so, dass auch eine gröbere Mutation sie sieht

#### M-4 (MEDIUM): Das Orakel für Lauf 1 prüft nicht, dass die Aufnahme acht Austausche hat

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15; `SPEC-038` Nr. 4 (Orakel: Requests 1 bis 8 gleich der Aufnahme, Folge von `$count` 7, 11, 13; Request 9 `410`)
- `pfad`: `tests/e2e/determinismus_test.go` · `for i, p := range aufnahme {`
- `befund`: Die Schleife läuft über so viele Austausche, wie `l.aufnahme` hält; eine Länge oder die Werte 7, 11, 13 stehen nirgends im Test. Mutation B2 (`aufnahmeReplay` führt einen `$count` weniger, in `tests/e2e/stufe_replay_test.go`) lässt `TestE2EDeterminismus` grün (1 von 1, `abweichend: 0`): Request 8 gibt dann `410`, wird nicht gegen die Aufnahme geprüft, und alle Läufe sind gleich. Eine Aufnahme, die ausfällt oder abbricht, lässt das Orakel stumm schrumpfen. Der Plan (Zeile 5) nennt als Mutation nur den leeren Header und zwei Fehlercodes.
- `verifizierbar`: ja (B2)
- `klasse`: Orakel ohne Vollständigkeitsprüfung seiner Eingabe

#### M-5 (MEDIUM): Drei Zusagen des Vertrags haben keine Tabellenzeile und keinen Fänger

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (jede Zusage eine Zeile; ein Negativ nennt Grund und Folge); `SPEC-038` Nr. 3 und Nr. 8 f (*„ein Fehler dabei ist rot“*, Herunterfahren), Nr. 6 (*„geschlossen nach jedem Lesen“*, *„nur lesend“*)
- `pfad`: `tests/e2e/determinismus_test.go` · `if err := h.Herunterfahren(ctx); err != nil {` und `if cerr := db.Close(); err == nil && cerr != nil {`
- `befund`: (a) Das Raster nennt für das Herunterfahren Nr. 8 f, die Tabelle hat dazu keine Zeile und kein benanntes Negativ; ein Fehler dabei ließe sich nur mit einem Handle provozieren, das beim Herunterfahren scheitert, und der Test hat keinen. (b) Mutation B5 (`leseStand` schließt die Ablage nicht) bleibt grün (`Läufe: 100, abweichend: 0`): 200 offene Verbindungen fallen nicht auf. (c) „nur lesend“ ist nicht erzwungen: `sqlite.Options` kennt keinen Lesemodus und `sqlite.Open` migriert (`SPEC-103`); die Zusage heißt, der Test schreibt nicht, nicht, dass die Verbindung es nicht könnte. Weder Tabelle noch Negativ in §3 nennen (a) bis (c).
- `verifizierbar`: ja für (b), nein für (a), (c)
- `klasse`: Zusage im Vertrag ohne Zeile in der Tabelle Zusage · Fänger

#### L-1 (LOW): Kommentare der Standtests sagen mehr zu, als die Fänger halten

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.11
- `pfad`: `tests/e2e/determinismus_test.go` · `// dem Reset in jedem Verstoß und der Kontrollwert des Lesers.`
- `befund`: Der Kopf von `TestDeterminismusStand` sagt zu „der Vergleich in jedem Feld jedes Cursors, der Stand nach dem Reset in jedem Verstoß und der Kontrollwert“; nach M-1 und M-2 hält der Fänger das für `Position`, `Consumed` und den Kontrollwert nicht an jedem Element. Gleiches gilt für den Kopf von `TestDeterminismusVergleich` („je Art der Abweichung ein Fall“), solange M-3 offen ist.
- `verifizierbar`: nein
- `klasse`: Kommentar sagt mehr zu, als die Prüfung fängt

---

#### I-1 (INFO): Die Aufrufstelle `k.VolatileHeaders` ist weniger ungefangen, als das Negativ sagt

Das akzeptierte Negativ (Plan §3, Absatz *Akzeptiertes Negativ*) nennt als einzige Folge, dass jemand die Liste durch den Default ersetzt. Die Mutation B1 (Liste ersetzt durch `nil`) ist rot, 10 von 10 Prozessen: `X-Correlation-Id` trägt in jedem Lauf einen anderen Wert, ein Datum bleibt unter einer Sekunde nicht gleich. Der Grund (feste Liste gleich dem Default bleibt grün, `VOLATILE_HEADERS` ist laut Nr. 3 ungesetzt) reicht für das Negativ; die Folge ist enger gefasst als nötig. Kein Befund.

#### I-2 (INFO): Klasse „Abweichung in einem einzelnen Lauf oder Request“ ist gedeckt

Mit den Mutationen B3 und B4 (Lauf 2 Request 1, Lauf 50 Request 5, Lauf 100 Request 1 und 8, Elemente, die weder Tabelle noch Implementer nannten) ist die Schleife über Läufe und Requests in beiden Randwerten rot. Request 9 und 10 (Fehlerantworten, nicht über `schreibeAntwort`) habe ich nicht einzeln mutiert.

#### I-3 (INFO): Die Entscheidung Gate-Kette oder daneben steht nirgends als getroffen

Geprüft: Plan §2 und §6, `SPEC-038` Nr. 2, `harness/README.md` (Sensors-Tabelle), das Fragment `harness/mk/test-determinism.mk` und die Commit-Nachrichten. Keine Textstelle stellt die Entscheidung als vom User getroffen dar; der Plan führt sie in §6 als *offen*, Adresse User, `SPEC-038` Nr. 2 weist sie dem Betreiber zu. Die Aussage „wie der User es entschieden hat“ steht in keinem Artefakt des Diffs (sie stand allenfalls im Bericht des Implementers; dort nicht geprüft). `GATE_CHECKS` ist unverändert. Kein Befund.

#### I-4 (INFO): Abweichung `LISTEN=127.0.0.1:0` bestätigt

`bootstrap.LadeKonfiguration` weist Port 0 ab (`src/bootstrap/konfig_werte.go`, Meldung *Port 1 bis 65535*); der Test lädt mit `127.0.0.1:8080` und setzt danach `Listen` auf `127.0.0.1:0`. Der Wortlaut von `SPEC-038` Nr. 3 widerspricht dem Code; die Berichtigung liegt beim Architect (Plan §6, Ausgang *offen*). Kein Befund am Diff.

#### I-5 (INFO): Treiber und Fragment, gelesen und teilweise gefahren

Argument `x` und `""`: `Aufruf:`, Exit 2; SIGTERM nach 1,5 s: Exit 2, kein Container bleibt; ungestört: Exit 0, grün. Die Wertung (`SKIP`, `FAIL`, kein `PASS`, Exit-Code) ist an der Datei gelesen und stimmt mit `SPEC-038` Nr. 2 überein; nicht nachgefahren habe ich `t.Skip` und das Entfernen von `--network none`/`--memory` (akzeptiertes Negativ, `SPEC-038` Nr. 10 f). Rohbytes: `LC_ALL=C grep -P` auf `\\u[0-9A-Fa-f]{4}`, BOM und NBSP in Test, Treiber, Fragment, Plan und Vertrag: 0 Treffer; Nicht-ASCII nur als Umlaute in Kommentaren.

---

**Negativbefund-Zeilen (geprüft, ohne Befund):**

- `tools/harness/` · `test-determinism.sh`: geprüft, ohne Befund (Exit-Codes, Signalbehandlung, Wertung, `--network none`, `--memory 2g`, `--read-only`, tmpfs gelesen; Aufruf und SIGTERM gefahren).
- `harness/mk/` · `test-determinism.mk`: geprüft, ohne Befund (kein `GATE_CHECKS`-Eintrag, Kommentar beschreibt, was da ist).
- `harness/README.md`, `harness/vertraege/SPEC-038-test-determinism.md`: geprüft, ohne Befund (Sensors-Zeile benennt Ziel, Vertrag, Bindung; `Stand` nennt den gelieferten Umfang).
- `docs/plan/planning/done/slice-test-determinism.md`: geprüft, ohne Befund außer M-5 (Tabelle) und I-1 (Negativ); Messung 613 stimmt mit `git diff --numstat` (11 + 529 + 73).
- `src/`, `Dockerfile`, `tests/e2e/uebergang/`: unverändert, nicht berührt.

**Nicht aufgelöst:** Zeilen 1, 2, 4, 5, 7, 8, 9, 12 der Tabelle und die Feld-Mutationen der Zeile 10 (gelesen, nicht nachgefahren); Treiber-Handläufe `t.Skip` und entferntes `--network none`; `make gates`; Lauf 1 gegen Orakel für Request 9 und 10 mit eigenen Mutationen; Verhalten des Handle bei Fehler im Herunterfahren (kein Fänger vorhanden).
