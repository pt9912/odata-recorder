# Review-Report: slice-odata-url-pfad — 2026-10-30

**Review-Art:** Code — geprüft gegen Plan (`slice-odata-url-pfad` §1, DoD, §3 mit der Tabelle *Zusage · Fänger* Z1 bis Z61, den Entscheidungen 1 bis 17, den *Mutationsbelegen und Raten*, §4 mit Block, Basis `7535bf7` und Marke 9200, §6 mit Raster und Zusatztabelle), `SPEC-160` bis `SPEC-163`, `SPEC-125`/`SPEC-126` (Beispiele), `SPEC-109` Nr. 2 und 5, `SPEC-127` Nr. 8 i und die Hard Rules `AGENTS.md` §3.7, §3.9 bis §3.18.

**Gegenstand:** `slice-odata-url-pfad` (Linie B, Welle `welle-mock`), `git diff 7535bf7..HEAD` ohne den Merge von `main` (`0edc059`); HEAD `179ca87`, die Commits des Implementers `49f836d` (Code und Plan), `6e74d75` (Spezifikation), `8869637` (Tests und Plan), `179ca87` (Messung). Der Produktcode änderte sich zwischen `49f836d` und `8869637` nicht.

**Skill:** `.harness/skills/reviewer.md` @ HEAD `179ca87` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-30

**Methode.** Der Bericht des Implementers (Plan §3, *Mutationsbelege und Raten*) wurde gelesen; keine Mutation, keine Rate und keine Einschätzung daraus ist übernommen. Eigene Werkzeuge liegen im Scratchpad `rev-urlpfad/` (nichts im Repo außer diesem Report): ein Probe-Paket mit Generatoren und eigenem Orakel (aus `SPEC-161` und `SPEC-162` gelesen, nie der Parser; Saat ab 100000; Dateien `probe_test.go` bis `probe4_test.go` in `rev-urlpfad/wt/src/adapters/driven/odata/urlparser/`), ein Mutationsrahmen (Änderung an einer Kopie, `go test -count=1` über `urlparser`, `model`, `outbound`, `baselineimport`, ohne das Probe-Paket; Kennungen `A1` bis `BZ1`, `T1` bis `X4` stehen nur in diesem Report) und ein Raten-Rahmen (je Prozess ein Aufruf des Testbinaries, Container `odata-recorder:race`). Alle schweren Läufe unter `/tmp/odata-recorder-heavy.lock`. `make gates` auf HEAD: Exit 0 (Datei `rev-urlpfad/gates.out`; `a-check` 0 Befunde, `d-check` 0 Befunde, `make groessenmarke` „Messung 4300 von Marke 9200“, Lint und Race grün).

**Eingangs-Kontext:**

- Slice-Plan `slice-odata-url-pfad` in `179ca87`
- `SPEC-160` bis `SPEC-163`, `SPEC-125`, `SPEC-126` Nr. 10 und 11, `SPEC-109`, `SPEC-127` Nr. 8 i, `SPEC-122`, `SPEC-128`
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)
- `LH-FA-08`, `LH-FA-07.b`, `LH-FA-12.b`, `LH-QA-05`
- `AGENTS.md` (Hard Rules §3.1 bis §3.18); Reviewer-Skill; Report `2026-10-30-slice-groessenmarke-werkzeug-review`
- Nicht doppelt gemeldet (Adresse beim Planner/Architect, Plan §6): (f) `a-check` sieht keinen Import zwischen Adaptern derselben Schicht, (g) Mutationsskript außerhalb des Repos, Aufträge an die Nehmer, Zeitbudget der Gegenproben, Stempel.

---

## Antworten auf die Fragen des Auftrags

**(1) Tabelle Z1 bis Z61, Raster, Mengen gegen die Klasse.** 116 eigene Mutationen am Produktcode und an den Zählungen der Tests (99 + 17: 101 rot, 8 Aufbaufehler ohne Aussage, 7 grün) lösen die Zeilen auf (Tabelle unten). Grün blieben: **L1** (F-3), **AJ3** (F-2), dazu **mW2** (F-4, Nebenläufigkeit, nicht in den 116), vier Mutationen, die ich als äquivalent einstufe (`AZ1` Reihenfolge der disjunkten Leser `zahl`/`guid`/`datum`; `BD1` ein ungültiges zweites Hex-Zeichen ergibt immer das Byte `0xFF`, also ebenfalls `kodierung`; `BP1` `i == 1` vor `$count` — ab `i >= 2` steht schon eine 501; `BW1` eine künstliche Sonderregel von mir) sowie `T7` (das Entfernen der Zählprüfung selbst, gewollt grün). Das Probe-Paket (Generatoren über Pfadformen, Schlüssel mit 1 bis 18 Teilen, Literale aller sechs Arten und der Typen ohne Modell, Prozent-Kodierungen, alle Optionen, Rangfolge Pfad × Query × Segment × Option, Zufallsbytes) lief mit **390 000 Eingaben** gegen mein Orakel: 60 000 Literale, 40 000 Körper, 60 000 Queries, 80 000 Pfad × Query, 150 000 Zufallsbytes — **ohne Abweichung des Parsers von der Spezifikation** (zwei Berichtigungen meines Orakels waren nötig: das leere Literal hinter `Id=` ist `key`; ein Schlüssel mit einem Literal ohne Modell ist hinter `$count` weiter ein Schlüssel; dazu I-2). Zufallsbytes: keine Panik, nie ein Teil der Eingabe in `Error()`, Ergebnis bei Fehler der Nullwert.

**(2) Regel „nacktes Wort ist `key`“.** Im Code vollständig: `inf`, `nan`, `Id`, `yes`, `t` sind `key`; `true`/`false`/`null` in jeder ASCII-Schreibweise sind Literal, `INF`/`NaN` `literal_typ`; `Id=yes` ist `literal`; Unicode-Namen (`Ünï`) sind `key`. Fünf Mutationen an der Regel (`Z1` `inf`/`nan` als Literal, `Z2` `null` aus der Liste, `Z3` nur die Kleinschreibung der Ausnahmen, `Z4` `INF`/`NaN` als `key`, `Z5` ein Wort mehr als `key`) sind alle in `TestLiteralBool` rot. Das Probe-Paket führt dieselben Wörter in 100 000 Körpern.

**Zählungen.** Z33 (12 nackte Wörter, 69 Fälle), Z6 (76 von 78), Z8 (29 von 81), Z9 (63 von 64) zählen; die Mutationen `T1` bis `T6` liegen in den Tests: `T1` (ein Wort aus der Liste gestrichen) rot in `TestLiteralBool`, `T2` (eine Paarung der Stufe 0 weniger) rot in `TestRangStufe0`, `T3` (ein Paar des Pfads weniger) rot in `TestRangPfadPfad`, `T4` (ein Paar der Query zusätzlich nicht bildbar) rot in `TestRangQueryQuery`, `T5` (eine Ursache der Query weniger) rot in drei Tests, `T6` (ein Marker weniger) rot in `TestFehlerOhneEingabe`. Z5 (das fünfte Paar `literal_typ` gegen `literal`) fängt `J1` (`TestRangImSegment`, `TestRangPfadQuery`, `TestRangPfadPfad`). Einschränkung zur Menge von Z5 in F-7.

**(3) Nebenläufigkeit, Raten mit Zahl der Prozesse.** Messender: mein Skript im Container `odata-recorder:race` (Go 1.27.1), Host unter Last durch fremde Läufe; je Zelle 20 Prozesse mit `-test.count=1`, einmal mit Detektor (`-race`), einmal ohne. Unveränderter Code: 0 von 20 in jeder Zelle.

| Mutation | `TestGleichzeitig` Detektor / ohne | `TestVerschraenkt` Detektor / ohne |
|---|---|---|
| m14b `erster501` in der Instanz, je Aufruf zurückgesetzt | 20/20 · 20/20 | 20/20 · 20/20 |
| m62a ein Puffer in der Instanz, den alle Aufrufe teilen | 20/20 · 20/20 | 0/20 · 0/20 |
| m15a Cache in einer Variablen des Pakets | 20/20 · 19/20 | 0/20 · 0/20 |
| mW1 gesperrter Zustand der Instanz, von jedem Aufruf geschrieben, nach dem Leser gelesen | 0/20 · 0/20 | 20/20 · 20/20 |
| mW2 derselbe Zustand, nur von einem Aufruf geschrieben, der durch den Leser läuft | grün (1 Prozess, ganze Suite, `-count=5`, ohne Detektor) | grün (ebenso) |

Der Detektor meldet mW1 nicht (0 von 20); bei mW2 liegt jeder Zugriff unter der Sperre, ein Rennen gibt es nicht (nicht gefahren). `TestVerschraenkt` fängt mW1 in jedem Lauf, mW2 nicht (F-4). Der Wächter ist zählbar und ohne Zeitgrenze (`TotalAlloc`); 80 Läufe von `TestWaechterZuteilung` und `TestUeberlangeEingabe` (`-count=40 -cpu 1,4`) sind stabil, die größte gemessene Zuteilung auf den Wächterfamilien liegt bei 22 % der Grenze von `SPEC-160` Nr. 6 c (`$select` aus 8184 Kommas). Die Grenze ist weit; eine quadratische Zuteilung fängt er, wo die Eingabe die Stelle erreicht (F-3).

**(4) Hüllen und Fakes.** Der Zähl-Fake `zaehlLeser` hält das einzige Argument (`text`), zählt jeden Aufruf und trägt die Schnittstelle des Ports; `haltLeser` hält an Kanälen. Es gibt keine Hülle um `ctx`, Strom oder Fake (der Port hat keinen `ctx`) und keinen Fake von `URLCodec` (Entscheidung 11, Sendung an `slice-mock-lesen`). Offen: `haltLeser` hat keine Grenze der Wartezeit (F-6).

**(5) Fehlerwert und Ort.** Z5: kein Teil der Eingabe in `Error()`/`Ort` — 150 000 Zufallseingaben und die 16 Fixtures bestätigen es; die Menge der Zeile trägt ihre Zahl nur zum Teil (F-7). Z4: Klasse, `Unwrap`, Code, Nullwert — `D1` (Klasse einer Ursache vertauscht) rot in vier Tests. Z61: `TestAbbildungURL` führt 16 Ursachen × 3 Verpackungen × 8 Gegenstände; `AT1` (Status 422) rot in `TestCodeTabelle`. Die Entscheidung des Planners, `ODR-E6005` binde diesen Slice nicht an die Handbuch-Pflicht, ist im Plan (§4, §6) begründet und wird hier nicht beanstandet (Befunde zu `ODR-E6005` nur gegen `SPEC-163`: keiner).

**(6) Äquivalente Mutationen und akzeptierte Negative.** `48d` (`len(t) > 19`) ist **nicht** äquivalent (F-2). `56b` ist aus dem Repo nicht nachzufahren (das Skript liegt außerhalb, Plan §6 (g)); es wird weder bestätigt noch bestritten. Die acht Einträge *Ungeprüft* von `SPEC-160` und die Zeile 58 tragen Grund und Folge. Der Eintrag „Laufzeit ohne Zuteilung … das Review des Diffs fängt sie“ ist hier eingelöst: im Diff steht keine Schleife über eine Schleife über die Eingabe; die größten verschachtelten Vergleiche sind `slices.Contains` über höchstens 256 Namen zu je höchstens 128 Zeichen (`leseSelect`), `enthaeltSortierung` über höchstens 32 und `strukturiere` über höchstens 16 Teile.

**(7) a-check, Lint, Alias.** `.golangci.yml` und `.a-check.yml` sind unverändert, keine `//nolint`-Direktive im Diff, keine neue Ausnahme; `LiteralGUID` steht in Code, Spezifikation und Plan einheitlich. `model.FilterAusdruck` als Alias der anonymen Schnittstelle: die Absicht steht im Kommentar und in §6 (e); `ireturn.allow` führt `anon` seit vor diesem Slice, die Form umgeht also nichts (I-3). Z59: `AU1` (Import von `jsoncodec` im Parser) ist in `TestImporte` rot; dass `make a-check` dabei grün bleibt, sagt der Plan (§6 (f)), ich habe es nicht erneut gefahren.

**(8) Plan folgt Code, Kommentare, Rohbytes, Größe.** §3.9: Code, Tests und Plan stehen im Commit `49f836d`, die Berichtigungen nach Entscheidung 17 in `6e74d75`/`8869637`; Tabelle (61 Zeilen, 2 nebenläufig), Zählungen, Spezifikation und Code stimmen überein (`LiteralGUID`, `'a''`, 12 Wörter). §3.7/§3.11: zwei Kommentare im Konjunktiv (F-1), die Köpfe zu Z5 und Z16 sagen mehr zu, als die Prüfung fängt (F-7, F-3). §3.18: Fixtures mit Nicht-ASCII entstehen aus Bytewerten (`"\xc3\xa4"`, `"e\xcc\x81t"`); `LC_ALL=C grep -P` über `urlparser/*.go`, `slice-odata-url-pfad.md` und `SPEC-160` bis `SPEC-163` auf eine Folge Rückstrich–`u`–vier Hex-Ziffern, NBSP (`\xC2\xA0`), BOM (`\xEF\xBB\xBF`) und U+FFFD: kein Treffer. Größe: `git diff --numstat 7535bf7..HEAD` über die 14 Pfade des Blocks ergibt 4291 + 9 = **4300**, gleich der Messung von `make groessenmarke` (Marke 9200); der Block nennt `url_ast_test.go`, die es nicht gibt (I-4).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | HIGH | Zwei Testkommentare beschreiben die verworfene Alternative im Konjunktiv: „`%5F` ist kein Name, `_` wäre einer“ und „ein Apostroph entstünde nicht“; sie sagen, was eine doppelte Dekodierung ergäbe, nicht, was der Test festhält. | `AGENTS.md` §3.7; Reviewer-Skill (HIGH: Kommentar trägt keine der Kommentar-Klassen) | `src/adapters/driven/odata/urlparser/pfad_test.go` · `das Ergebnis `%5F` ist kein Name, `_` wäre einer` und `der Text `%27` bleibt, ein Apostroph entstünde nicht` | nein — kein Gate liest Kommentare | Konjunktiv über die verworfene Alternative im Kommentar |
| F-2 | MEDIUM | Die Mutation `48d` (`len(t) > 19` statt `> 10`) ist nicht „dieselbe Antwort für jede Eingabe“: mit ihr bleibt die Suite grün und `$top=9223372036854775808` ergibt `Top{Gesetzt: true, Wert: -9223372036854775808}` (19 Ziffern laufen in `int` über; gemessen auch `9223372036854775813` und `9999999999999999999`). Die Menge von Z48 hat 20 und 21 Ziffern (2^64, 2^64 + 5), aber keinen Wert mit 19 Ziffern ab 2^63; `SPEC-160` Nr. 2 sagt „`Top.Wert` und `Skip.Wert` sind nie negativ“. | `AGENTS.md` §3.15 (Fänger führt jede Form), §3.17; Reviewer-Skill (MEDIUM: Fänger führt nur eine Form der Eingabe) | `src/adapters/driven/odata/urlparser/query.go` · `if !nurZiffern(wert) \|\| len(t) > 10`; Plan §3 *Mutationsbelege* · `48d (len(t) > 19 statt > 10: dieselbe Antwort für jede Eingabe` | ja — Mutation `len(t) > 19` an einer Kopie: Suite grün; ein Fall `9223372036854775808` in `TestTopSkip` wird rot | Mutation als äquivalent erklärt, Fänger führt nur eine Form der Eingabe (Überlauf bei 19 Ziffern) |
| F-3 | MEDIUM | Der Wächter von Z16 fängt eine quadratische Zuteilung nur, wo die Eingabe aus Escapes besteht: die Mutation `L1` (`out = append(append([]byte(nil), out...), s[i])` im Zweig der rohen Bytes von `dekodiere`) bleibt grün; für `x=%41` + 8000 × `a` teilt sie 34 325 896 Bytes zu (Grenze 577 984), für einen Pfad aus 2000 × `a` und `%41` 2 158 096 (Grenze 193 792). Die Familie `%` der Menge 12 ist `%41` · n, ein rohes Byte neben einem Escape fehlt; der Kopf von Z16 („auch für eine, die aus lauter … `%` … besteht“) trifft nur diese Form. | `AGENTS.md` §3.15, §3.11; Reviewer-Skill (MEDIUM: Fänger führt nur eine Form der Eingabe); `SPEC-160` Nr. 6 c | `src/adapters/driven/odata/urlparser/dekodieren.go` · `out = append(out, s[i])`; `grenzen_test.go` · `{"%", teile(wiederhole("%41", 140)), "x=" + wiederhole("%41", 2730), true, true}` | ja — `L1` an einer Kopie, Suite grün; Fall `x=%41` + 8000 × `a` rot | Fänger führt nur eine Form der Eingabe (Wächter-Familie ohne gemischten Text) |
| F-4 | MEDIUM | Die Zusage von `SPEC-160` Nr. 7/8 („gleichzeitige Aufrufe beeinflussen sich nicht“, der Leser wird aus mehreren Goroutinen gerufen) hat keinen Fänger für zwei Aufrufe, die beide durch den `FilterLeser` laufen: `TestGleichzeitig` läuft ohne Leser, in `TestVerschraenkt` hat der freie Aufruf kein `$filter`. Die Mutation `mW2` (Feld der Instanz unter einer Sperre vor dem Leser geschrieben und danach als `Filter.Text` gelesen) bleibt in der ganzen Suite grün (`-count=5`); der Detektor hat nichts zu melden (alle Zugriffe unter der Sperre). | `AGENTS.md` §3.10 (Zusage über Nebenläufigkeit gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird), §3.15; Reviewer-Skill (MEDIUM: Fänger führt nur eine Form der Eingabe) | `src/adapters/driven/odata/urlparser/nebenlaeufig_test.go` · `faelle := []struct{ halt, frei [2]string }{` und `p := parserNeu()` in `TestGleichzeitig` | ja — Mutation `mW2` an einer Kopie (`go test -count=5` grün) | Nebenläufigkeits-Fänger deckt nicht jedes Fenster (zwei Aufrufe durch den Leser) |
| F-5 | MEDIUM | Liefert ein `FilterLeser` einen typisierten nil-Zeiger als Fehler (`var e *model.URLFehler; return nil, e`), trifft `errors.As` und `uf.Ursache` dereferenziert nil: `ParseURL` bricht mit einer Panik ab, statt `model.ErrUngueltig` ohne `URLFehler` zu liefern (`SPEC-160` Nr. 8 d: „ein anderer Fehler … ist ein Fehler des Aufrufers“). Das Raster nennt „Leser liefert Fehler, fremden Fehler, nichts“, die Form fehlt in Zelle und Zeile 19. | `AGENTS.md` §3.16 (Randform ohne Zelle), `SPEC-160` Nr. 8; Reviewer-Skill (MEDIUM: unklare Fehlerbehandlung am Rand des Spec-Bereichs) | `src/adapters/driven/odata/urlparser/query.go` · `case errors.As(err, &uf) && uf.Ursache.Gueltig():` | ja — Probe mit einem Leser, der `(*URLFehler)(nil)` liefert: `nil pointer dereference` | Randform eines eingehängten Ports ohne Zelle (typisierter nil-Fehler) |
| F-6 | LOW | `TestVerschraenkt` wartet mit `<-h.erreicht` ohne Grenze. Erreicht der Aufruf den Leser nicht (Mutation `G1`: ein 501 des Pfads beendet vor der Query; `H1`: `$batch` als 400), hängt der Prozess bis zur Zeitgrenze von `go test` (600,1 s und 586,6 s gemessen) und meldet `panic: test timed out`, nicht den benannten Test; die Ergebnisse der übrigen Tests des Pakets gehen verloren. | `AGENTS.md` §3.10 (ein Test zählt nur, wenn er aus dem benannten Grund rot wird) | `src/adapters/driven/odata/urlparser/nebenlaeufig_test.go` · `<-h.erreicht` | ja — Mutation `G1` an einer Kopie: Lauf endet nach 600 s | Fake blockiert ohne Grenze der Wartezeit |
| F-7 | LOW | Die Menge von Z5 („16 Ursachen × 3 Stellen = 48, der Marker steht in der Eingabe“) trägt nur zum Teil: von 48 Zellen haben 18 einen Marker in der Eingabe (`service_dokument` keinen, die meisten Ursachen nur einen von drei); der Kopf des Tests zählt 64 Schleifenläufe. Eine Weitergabe der Eingabe an `Error()`/`Ort` in einer der anderen Zellen fängt nur die feste Form von Ort und Text über den Korpus (`pruefeFehlerform`), nicht der Marker. | `AGENTS.md` §3.11, §3.15 | `src/adapters/driven/odata/urlparser/fehlerwert_test.go` · `for _, marker := range []string{"ZnN", "ZwW", "ZsS", "\xc3\xa4"}`; Plan Z5 · `der Marker steht in der Eingabe` | ja — Zählung der Platzhalter in `ursachenFaelle()` | Menge zählt Schleifenläufe, nicht Fälle |
| I-1 | INFO | Rollen-Achse: die Entscheidung 17 steht im Abschnitt „Entscheidungen des Architect“, die Berichtigung der Spezifikation in `6e74d75` mit dem Präfix „Implementer“; ob der Architect sie schrieb, ist aus dem Repo nicht ablesbar. Zuständig: Planner/Integrationsschritt. | Maintainability (Rollen-Trennung, Modul 8) | `docs/plan/planning/done/slice-odata-url-pfad.md` · `17. **Berichtigungen nach der Rückgabe des Implementers (2026-10-30).**` | nein | Herkunft einer Entscheidung nicht ablesbar |
| I-2 | INFO | `SPEC-161` Nr. 6 g („Buchstaben“) und `SPEC-162` Nr. 3 (f) („Zeichen jenseits von ASCII“) lassen offen, ob Unicode-Buchstaben und -Ziffern im Typpräfix (`duraätion'x'` ergibt `literal_typ`) und im Parameter-Alias (`@ä` ergibt `Sonstige`) gelten; der Code folgt `SPEC-125` Nr. 8 b (`L`, `Nd`). Kein Fänger, keine Zelle. Zuständig: Architect. | `SPEC-161` Nr. 6, `SPEC-162` Nr. 3 | `src/adapters/driven/odata/urlparser/literal.go` · `!unicode.IsLetter(r) && !unicode.IsDigit(r)` | nein | Randform der Spezifikation nicht ausdrücklich entschieden |
| I-3 | INFO | `model.FilterAusdruck` ist ein Alias der anonymen Schnittstelle; Kommentar und §6 (e) nennen den Grund (`ireturn`); `.golangci.yml` führt `anon` in `ireturn.allow`, unverändert seit vor dem Slice. Keine Aktion. | `AGENTS.md` §3.2, §3.6 | `src/hexagon/model/url_ast.go` · `type FilterAusdruck = interface {` | ja — `make lint` grün | — |
| I-4 | INFO | Der Block `groessenmarke` nennt `src/hexagon/model/url_ast_test.go`, die Datei existiert nicht; `export_test.go` und `beenden_test.go` stehen nicht im Block (Plan §6, bekannt). Die Messung 4300 ist davon unberührt (eigene Rechnung 4291 + 9). | `SPEC-128` Nr. 4 | `docs/plan/planning/done/slice-odata-url-pfad.md` · `pfad: src/hexagon/model/url_ast_test.go` | ja — `make groessenmarke` | — |
| I-5 | INFO | `urlparser/zeit.go` und `jsoncodec/zeit.go` führen dieselben Zeilen für Datum, Uhrzeit, Sekunden, Versatz und kanonischen Zeitpunkt (rund 140 Zeilen, getrennt gepflegt); die Gleichheit der Texte ist in `SPEC-160` *Ungeprüft* mit Grund und Folge benannt, der Fänger ist der Ende-zu-Ende-Fall in `slice-mock-lesen`. Zuständig: Architect, falls die Duplizierung bleiben soll. | `SPEC-160` *Ungeprüft*; Maintainability | `src/adapters/driven/odata/urlparser/zeit.go` · `func zeitpunkt(s string) (model.Literal, bool)` | nein | Dupliziertes Kanon-Verfahren zwischen Adaptern |
| I-6 | INFO | Die Kennung `62a` in *Mutationsbelege* (Zeilen 14 und 15) hat keine Zeile 62; die Tabelle endet bei 61. Rückverfolgung nur über den Text. | Maintainability | `docs/plan/planning/done/slice-odata-url-pfad.md` · `62a ein Puffer in der Instanz, den alle Aufrufe teilen` | nein | Kennung ohne Zeile |
| I-7 | INFO | Betriebsregel-Vorfälle des Implementers (Dateien anderer Läufe im geteilten Scratchpad überschrieben oder gelöscht): vom Auftrag genannt, im Repo nicht ablesbar und hier nicht geprüft. Zuständig: Planner (Beobachtungs-Register, `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`, im Plan §8 mit 1× geführt). | Plan §3 *Betriebsregeln für alle Rollen* | — | nein | Betriebsregel verletzt |

## Auflösung der Zeilen (eigene Mutationen, rote Tests)

Kennung · Zeile · roter Test (der aus dem Grund der Zeile). „Zeitgrenze“ bezeichnet einen Hänger bis zur Zeitgrenze von `go test` (F-6).

| Zeile | Mutationen | rot |
|---|---|---|
| 1 | A1 `#` im Pfad erlaubt, A2 DEL, A3 Steuerzeichen in `rawQuery` nur am Anfang | `TestArgumenteFormverstoss` |
| 2 | B1, B2 `>=`, C8 Grenze 2047 | `TestGrenzenLaenge` |
| 3 | C1 bis C7, AW1, BJ1, BY1 | `TestGrenzenAnzahl` |
| 4 | D1 Klasse von `literal_typ` | `TestFehlerwert` |
| 5 | E1b Ort trägt den Namen der Option, T6 | `TestFehlerOhneEingabe` |
| 6 | F1, F2, F3, T2, T5 | `TestRangStufe0` |
| 7 | G1 (Zeitgrenze), I1 | `TestRangPfadQuery`, `TestFilterMitLeser` |
| 8 | H1 (Zeitgrenze), AV1, T3 | `TestRangPfadPfad`, `TestVerstossHinter501` |
| 9, 11 | I1, K1, AQ1, T4 | `TestRangQueryQuery`, `TestOptionDoppelt`, `TestRangInDerOption` |
| 10 | J1, BR1 | `TestRangImSegment`, `TestSegmentForm` |
| 14, 15 | m14b, m62a, m15a, mW1; **mW2 grün** | siehe Tabelle der Raten |
| 16 | **L1 grün** | — |
| 17 | O1, O2 | `TestUeberlangeEingabe` |
| 18, 19 | AX1, P1 bis P3, E1b; typisierter nil-Fehler des Lesers: **ohne Fänger** | `TestFilterOhneLeser`, `TestFilterMitLeser` |
| 20, 41 | Q1, BQ1, AV1 | `TestPfadFormen`, `TestRestSegmente` |
| 24, 25, 26 | S1, U1b, T1 (Batch 1, `IndexByte(s, 0) < 0` zu `true`) | `TestKodierungUngueltig`, `TestUTF8`, `TestNul` |
| 29 bis 33 | V1 bis V3, Y1, BT1, BU1, BX1, Z1 bis Z5 | `TestKeyStruktur`, `TestLiteralString`, `TestLiteralBool` |
| 34 bis 39 | AA1 bis AA3, AB1, AC1b bis AC3, AD1b bis AD7, BA1, BB1, BZ1, X4 | `TestLiteralZahl`, `TestLiteralGuid`, `TestLiteralDatum`, `TestLiteralZeitpunkt`, `TestLiteralTyp` |
| 43 bis 45 | AR1b, AR2 | `TestQuerySegmente`, `TestParserZerlegtWieModell` |
| 46, 47 | AH1, BJ1, BL1, K1 | `TestOptionsklassen`, `TestOptionDoppelt` |
| 48, 49 | AJ1, AJ2, AK1, X1, X2; **AJ3 grün** | `TestTopSkip`, `TestGesetzt` |
| 50 bis 53 | AL1, BM1, AM1b, AM2, AN1, AN2, AO1, AY1, BC1, BK1, X3 | `TestCount`, `TestSelect`, `TestOrderby`, `TestFormat` |
| 55 | AQ1 | `TestWertUngelesen` |
| 59 | AU1 | `TestImporte` |
| 60 | AT1 | `TestCodeTabelle` |

Die Zeilen 12, 13, 21 bis 23, 27, 28, 40, 42, 54, 56 bis 58 und 61 sind über das Probe-Paket und die Mutationen der Nachbarzeilen mitgeprüft (keine Abweichung des Orakels), nicht je mit einer eigenen Mutation aufgelöst. Die Mutation `AF1b` (Rest-Segment ohne Prüfung der Klammer) ließ sich nicht bauen und blieb ungefahren.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/odata/urlparser/` (Produktcode: `parser.go`, `dekodieren.go`, `pfad.go`, `schluessel.go`, `literal.go`, `zeit.go`, `query.go`) | gegen `SPEC-160` bis `SPEC-162` gelesen und mit 390 000 Probe-Eingaben gegen ein eigenes Orakel geprüft; im Verhalten kein Verstoß gegen die Spezifikation gefunden; Befund F-5 (Panik bei einer Randform des Lesers), sonst ohne Befund |
| `src/adapters/driven/odata/urlparser/*_test.go` | Tabellen, Generatoren (Saat 1 bis 2000, Zählprüfung der Läufe), Zählungen geprüft; Befunde F-1, F-2, F-3, F-4, F-6, F-7, sonst ohne Befund |
| `src/hexagon/model/` (`url_ast.go`, `url_fehler.go`, `query_segmente.go`, `fehler.go`) | geprüft, ohne Befund (Klasse, `Unwrap`, Code, Tabelle; `QuerySegmente` hält die Zerlegung, Mutationen AR1b und AR2 rot) |
| `src/hexagon/ports/outbound/` (`url_codec.go`, `abbildung.go`) | geprüft, ohne Befund (Port ohne `ctx` begründet, Abbildung 16 × 3 × 8) |
| `src/hexagon/services/baselineimport/` (`auswahl.go`, `export_test.go`, `zerlegung_test.go`) | geprüft, ohne Befund; Verhalten des Imports unverändert (gleiche Zerlegung, Dekodierung nachsichtig) |
| `src/hexagon/services/session/beenden_test.go` | geprüft, ohne Befund (zwei Namen in der Liste der Tabellenzeilen) |
| `spec/spezifikation.md` (`SPEC-160` bis `SPEC-163`, `SPEC-109`, `SPEC-127` Nr. 8 i, Historie) | geprüft; Bezeichnungen und Beispiele stimmen mit Code und Plan überein; I-2 |
| Kommentare, Rohbytes, Größe (`AGENTS.md` §3.7, §3.11, §3.18, §3.14) | F-1, F-3, F-7; Rohbytes ohne Befund; Größe 4300 von 9200 bestätigt |
| `.golangci.yml`, `.a-check.yml`, `Makefile`, `harness/` | in den Commits des Implementers unverändert; keine Lockerung |
| Stempel `harness/gegenproben.stempel` | von den Commits des Implementers nicht berührt (die Änderung im Gesamtdiff stammt aus dem Merge von `main`) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 1 |
| MEDIUM | 4 |
| LOW | 2 |
| INFO | 7 |

**Finding-Klassen dieses Laufs:** Konjunktiv über die verworfene Alternative im Kommentar · Mutation als äquivalent erklärt, Fänger führt nur eine Form der Eingabe (Überlauf bei 19 Ziffern) · Fänger führt nur eine Form der Eingabe (Wächter-Familie ohne gemischten Text) · Nebenläufigkeits-Fänger deckt nicht jedes Fenster · Randform eines eingehängten Ports ohne Zelle (typisierter nil-Fehler) · Fake blockiert ohne Grenze der Wartezeit · Menge zählt Schleifenläufe, nicht Fälle

## Verdikt

**Merge-blockierend:** ja — ein HIGH (zwei Kommentarzeilen) und vier MEDIUM. Die Befunde liegen in den Fängern und in einem Rand des Ports, nicht im Verhalten des Parsers gegen die Spezifikation: 390 000 Probe-Eingaben ergaben keine Abweichung, `make gates` ist grün, die Messung (4300 von 9200) stimmt.

**Übergabe:** Findings gehen an den Implementer; ein Fix deckt die Klasse (`AGENTS.md` §3.17): zu F-2 die Menge „Werte, bei denen die Ziffernfolge den Bereich von `int` überschreitet“ (19 Ziffern ab 2^63 und jede Länge zwischen 11 und 19), zu F-3 die Menge „Eingaben mit einem Escape und rohem Text an jeder Stelle von `pfad` und `rawQuery`“, zu F-4 die Menge „Paare von Aufrufen, die beide den Leser erreichen, in beiden Reihenfolgen“, zu F-5 jede Art von Fehlerwert, die ein Leser liefern kann (nil-Zeiger eines `*URLFehler`, eingepackt und nicht eingepackt). F-5 braucht vorher eine Zelle im Raster (Architect, `AGENTS.md` §3.12, §3.13: Eintrag in §6 mit Ausgang *offen* im selben Commit wie der Bericht). I-2 und I-5 gehen an den Architect, I-1, I-4, I-6, I-7 an den Planner. Die Finding-Klassen gehen in die Closure §7 und in den Zähler. Dieser Report ist ein Lauf-Beleg und ersetzt keine Verifikation.
