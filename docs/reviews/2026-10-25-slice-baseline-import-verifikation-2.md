# Verifikations-Report 2: slice-baseline-import — 2026-10-25

**Rolle:** Verifier, zweite Verifikation, frischer Kontext (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-127`, `LH-FA-18`, `LH-FA-11`, `LH-FA-14`, `LH-FA-03`, `AGENTS.md` §3.7, §3.9 bis §3.18). Kein Review, keine Reparatur.

**Gegenstand:** `slice-baseline-import` (Welle `welle-record-replay`), HEAD `70882c4`, Basis der Marke `cf1fb79`. Geprüft: die Verifikation `2e3a250` (V-1 MEDIUM, V-2 bis V-4 LOW, V-5 bis V-10 INFO), die Entscheidungen des Architect `81c52ed` (Entscheidungen 19 bis 23, `SPEC-127` Nr. 8 g, 12 e, 18, *Ungeprüft*), die zweite Fix-Runde `59f8a2e` (Tests und Plan, kein Produktcode) und der Stempel `70882c4`.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-25

**Methode.** Belege statt Behauptungen: keine Zahl und keine Mutation des Implementers ist übernommen. Alle Mutationen laufen in Wegwerfkopien außerhalb des Repos (Scratchpad `…/scratchpad/ver2/`, flüchtig), Quelle `:ro`, `--network none --memory=4g --memory-swap=4g --cpus=4`, ein Container zugleich, Gates nie parallel zu Mutationen. **Welcher Stand geprüft wurde:** `ver2/w0/src` ist ein `cp -r` des Arbeitsbaums `src/` bei `70882c4` (`diff -r` gegen `src/` leer; `git status --short` zeigt allein `.claude/agents/architect.md`, nicht angefasst). Das Image `deps` und das Image `race` tragen keinen Quellstand, den ich nutze: jeder Lauf leert `/src/src` und kopiert die Kopie hinein. Die Gate-Läufe bauen aus dem Repo selbst (stufenweise ohne Cache, siehe 4). Bei Beginn liefen keine fremden Container; Last 3,3 (1 Minute), 19 GB verfügbar.

---

## 1. V-1 gegen die Klasse (`SPEC-127` Nr. 8 g)

**Die Elementzahlen, im Lauf nachgezählt.** `TestZeileE3InhaltstypKlassen` zählt jede Klasse und hält bei einer Abweichung an (`fmt.Sprint(p.zaehl) != fmt.Sprint(soll)`); ich habe die Quelle der Zahlen gelesen und vier Elemente einzeln entfernt (ein Sonderzeichen `\x7f` in (b), ein ähnlicher Name in (c), `;=utf-8` in (h), eine Schreibweise in (c) durch eine Maske): **4 von 4 rot** durch die Zählprüfung.

| Klasse | Soll | Quelle der Zahl (Code gelesen) | Ergebnis |
|---|---|---|---|
| (a) Rand des Medientyps | 147 | 7 Folgen aus {Leerzeichen, Tabulator} der Länge 0 bis 2, links mal rechts mal 3 Medientypen = 7 · 7 · 3 | stimmt |
| (b) Zeichen am Rand, nicht befreit | 204 | 31 Bytes 0x00 bis 0x20 ohne Leerzeichen und Tabulator, `0x7F`, U+0085, U+00A0 = 34; mal 2 Seiten mal 3 Medientypen | stimmt |
| (c) Name des Headers | 4104 | 2048 Schreibweisen (2^11, Bindestrich hat keine) mal 2 (allein; neben abweichender Schreibweise) = 4096; 6 ähnliche Namen; kein Wert; drei Werte | stimmt (4096 + 6 + 1 + 1) |
| (d) ASCII-klein | 266 | 7 Verwendungen; 256 Bytes; 3 Zeichen (U+212A, U+017F, U+0130) | stimmt |
| (e) Leerraum um Name und Wert | 100 | 4 Parameter mal (4 Orte mal 7 Folgen, entdoppelt 25) | stimmt |
| (f) / (g) / (h) / (i) / (j) | 4 / 3 / 9 / 6 / 1 | Aufzählung im Test | stimmt |
| Summe | 4844 | | stimmt |

**Die Zahl in (h): 9 oder 8.** Der Wortlaut von Nr. 8 g führt diese Randformen der Parameter: ein leerer Teil (`;;`, ein abschließendes `;`, `; `), ohne `=` (`charset`, `odata.metadata`), der Wert ist der Rest hinter dem ersten `=` (`charset=utf-8=x`), der leere Wert (`charset=`, `odata.metadata=`) und der leere Name (`=utf-8`). Das sind **neun verschiedene** Formen; keine ist doppelt, keine fehlt. Die Zahl 9 und die Summe 4844 entsprechen dem Wortlaut; die Entscheidung 19 und die Tabelle an ihrem Ende (8 und 4843) sind falsch. Der Plan führt den Widerspruch selbst in §6 (Ausgang *offen*, Adresse Architect); er ist nicht gelöst (V2-3).

**Mutationen.** Alle Läufe gegen die ganze Suite des Pakets (`go test`, 3 Sekunden), Basis HEAD.

| Gruppe | Mutationen | Ergebnis |
|---|---|---|
| Die neun Mutationen des Verifiers 1 (m1 bis m9): nur Leerzeichen gekürzt, nur rechts gekürzt, Tabulator rechts nicht, Tabulator links nicht, Header-Name nur zwei Schreibweisen, ohne `CONTENT-TYPE`, ohne `cOntent-type`, `TrimSpace`, Teilung `;` nach Anführungszeichen | 9 | **9 von 9 rot**, alle in `E3InhaltstypKlassen` (m9 dort allein) |
| Rand des Medientyps, Elemente, die weder Verifier noch Implementer nannten (n1 bis n6): `\v`, `\f`, `0x1F`, U+00A0, U+0085, `\r` mit gekürzt; Randzeichen am Medientyp, Suffix, Präfix, Teilstring (n19, n20, n18) | 9 | **9 von 9 rot** |
| Name des Headers: `EqualFold` (n23), nur Tabulator/`\v`/NBSP gekürzt (n24, p12, p13), `TrimSpace` (p11), `Contains` (n25), Präfix (n21 am Parameter) | 6 | p11, n25 rot; n23 äquivalent (`content-type` hat kein `k` und kein `s`); n24, p12, p13 grün (V2-2) |
| Kleinschreibung weggelassen je Verwendung (l1 Medientyp, l2 Name des Parameters, l3 Wert, l4 Header-Name, l5 nur ein XML-Typ, l6 `none`) | 6 | **6 von 6 rot** |
| Anzahl der Werte, Teilung, Anführungszeichen (m10, m11, n17, n22, n27, p4, p5, p6) | 8 | **8 von 8 rot** |
| Parameter: `TrimSpace` am Namen (p1), am Wert (p2), an beiden (p3); `\v` (n9), U+00A0 (n10) oder `"`/`'` (n8, n11) am Rand des Namens; `\v` (n12), U+00A0 (n13), `\r\n` (n14) am Rand des Wertes | 10 | **10 von 10 grün (V2-1, MEDIUM)** |
| Namen, die nur ähneln: `charset…` als Präfix (p7), Suffix (p8), `odata.metadata…` als Präfix (p9), Suffix (p10) | 4 | **4 von 4 grün (V2-2, LOW)** |
| Anführungszeichen: Paar nicht verlangt, nur vorn geprüft (n15), nur hinten geprüft (n16) | 2 | grün (V2-2) |
| Medientyp: `"` am Rand mitgekürzt (n7) | 1 | grün (V2-2) |
| `ToLower` statt `asciiKlein` am Medientyp, Namen, Wert, Header (p14 bis p17) | 4 | äquivalent: die verglichenen Literale tragen kein `k`, kein `s`, kein `İ`-Bild |

**Eigenes Probe-Paket (Orakel unabhängig vom Dienst).** `zz_ct2_test.go` (flüchtig, `ver2/`): ein Orakel von Nr. 8 g als Zustandsautomat über Bytes (ohne `strings.Cut`, `Trim`, `Split`; Schreibweise und Kürzen von Hand). **Neue Klasse, nicht im Plan:** alle Folgen aus 32 Bausteinen (Leerzeichen, Tabulator, `;`, `=`, `"`, `charset`, `CharSet`, `utf-8`, `UTF-8`, `utf8`, `odata.metadata`, `ODATA.Metadata`, `none`, `NONE`, `minimal`, `x`, `\v`, `\f`, `0x1F`, `0x7F`, U+0085, U+00A0, U+212A, `\r`, `\n`, `,`, `'`, `\`, drei Medientypen in zwei Schreibweisen) der Länge 0 bis 4 hinter `application/json;`: **1 082 401** Folgen, davon 44 mit `ohne_kontext`, 0 Abweichung; dazu 3 000 000 zufällige Header-Tabellen (1 bis 3 Namen aus 10 Schreibweisen und Fehlformen, 0 bis 2 Werte je Name, bis 9 Bausteine) gegen zwei Medienmengen: **4 499 513** Fälle, 141 223 gelesen, 3 `ohne_kontext`, 0 Abweichung. Der Dienst ist richtig; die Befunde V2-1 und V2-2 betreffen allein die Fänger.

**Urteil V-1.** Die Klasse, die Verifier 1 nannte (Rand des Medientyps mit beiden Zeichen, Name in allen Schreibweisen, Zahl der Werte, Leerraum um den Wert), ist **gelöst**: 9 von 9 genannten und 25 von 25 nicht genannten Mutationen daran rot (n1 bis n6, n18 bis n20, p4 bis p6, p11, n17, n22, n25, n27, m10, m11, l1 bis l6). Der Wortlaut sagt „Befreit werden allein das Leerzeichen und der Tabulator“ aber **für drei Stellen** (Medientyp, Name, Wert eines Parameters, „dieselbe Befreiung“); der Fänger der Zeichen, die nicht befreit werden, steht nur an einer (V2-1).

## 2. V-2 bis V-10

| Befund | Prüfung | Urteil |
|---|---|---|
| V-2 (D1 18, E3 (i) 6) | D1: 6 Klassen mal 3 Zahlen = 18 (Test zählt); sieben Mutationen (`json`, `xml`, `%`, leer, `=` abgewiesen; zwei Optionen; `atom` als Negativ): 6 von 6 rot, `atom` grün wie das Negativ sagt. E3 (i): 6 Werte (Test zählt), `charset`-Präfix und -Suffix rot. Die Negative stehen im Plan mit Grund und Folge (*Akzeptierte Negative*, *Nach der Verifikation*) | gelöst; Reste der Klasse: V2-2 |
| V-3 Stempel | `make gegenproben` am HEAD: „Eingaben unverändert seit dem letzten grünen Lauf (56444555…)“, 0,5 s, `git status` danach unverändert. Erzwungen (`GEGENPROBEN_STEMPEL` im Scratchpad): **37 Fälle, 0 rot**, Schritt der Fälle **56,7 s**, 87 s gesamt (Frist 120 s je Fall); der geschriebene Stempel ist **gleich** dem in `harness/gegenproben.stempel` | gelöst |
| V-4 (H6 (E) 82) | Orakel `net/url`; Mutationen `H6-userinfo-erstes` (`IndexByte`), nur bei einem `@` gekürzt, nur bei weniger als drei, nie gekürzt: **4 von 4 rot** (zwei nicht genannte: „bei weniger als drei“, „nur bei genau einem“). Entscheidung 20 steht in `SPEC-127` Nr. 12 e; Plan §6 trägt Ausgang *entfallen* | gelöst |
| V-5 Mengen | C1 180 Paare (Test zählt `paare != 180`), E1 10 und 4 (`Fatalf`), H10 (a) 10 (`Fatalf`), H6 28 und 2 mal 464, E5 90, I2 4: gezählt im Test; Elemente entfernt: E1 (`Orders`) rot, H10 (a) (`{anders, gleich, gleich}`) rot. Die Zeilen H6, E5, C1, I2, E1, H10, D1, B7, I10 der Tabelle tragen die Mengen | gelöst |
| V-6 Barriere | `B7SchrankeSelbsttest`, Menge 3 (Test zählt `n != 3`). Rate unter dem Detektor, ein Testbinary, 20 Prozesse mit `-test.count=1`, Messender: Verifier: ohne Mutation **0 von 20**; `ablauf.Store(true)` entfällt **20 von 20**; `oeffne` nach der Frist entfällt **20 von 20**; `pruefe` ignoriert den Ablauf **20 von 20** (je `B7SchrankeSelbsttest`) | gelöst |
| V-7 H13 | `H13EvidenzMitKette` zählt `n != 12`, `H13FormDerEvidenz` `n != 3`; die Zeile H13 nennt 15 Goldens; `helfer_test.go` hat 961 Zeilen (`wc -l`), der Plan nennt 961. Der Rest an Zeile 397 ist die Wiedergabe des Befunds | gelöst |
| V-8 I10 | `goroutinenIn` liest den Selektor, nicht den Aufruf; 7 Formen mal die Dateien des Pakets (`n != 7*len(dateien)`, heute 11 Dateien = 77; die Zahl 11 hält der Test nicht, V2-4). Eigene Einfügungen in Dateien, die der Implementer nicht nannte: `go func() {}()` in `manifest.go`, `time.AfterFunc` als Variable über einen umbenannten Import in `kette.go`, `WaitGroup.Go` in `kanon.go`, `context.AfterFunc` als Wert in `metadaten.go`: **4 von 4 rot**; `iter.Pull` in `coverage.go` grün, wie das Negativ sagt | gelöst |
| V-9 echte Leser | akzeptiertes Negativ mit Grund und Folge; die Sendung steht in `slice-bootstrap-m2` §1 *Übernimmt (Import)* (`git diff cf1fb79..HEAD -- docs/plan/planning/open/slice-bootstrap-m2.md`, 1 Zeile), dessen *Ausdrücklich NICHT* sie nicht trifft (gelesen) | gelöst |
| V-10 Größe | siehe 5 | Planner |

**Raten (Detektor, 20 Prozesse je Mutation, Messender: Verifier; Basis HEAD, Testbinary `-race`):**

| Mutation | rot | Tests |
|---|---|---|
| ohne Mutation | **0 von 20** | — |
| Mutex um den ganzen Aufruf von `Import` | **20 von 20** | B5, B6, B7 je 20 |
| Mutex um die Wiederholungsprüfung | **20 von 20** | B5, B6, B7 je 20 |
| ein Zähler des Pakets in `Import` | **20 von 20** | B5 |
| Mutex erst nach dem ersten Port (`ladeSession`) | **0 von 20** | benanntes Negativ |
| Selbsttest der Barriere: `Store` entfällt / `oeffne` entfällt / `pruefe` ignoriert | je **20 von 20** | `B7SchrankeSelbsttest` |

(Ein erster Lauf mit einer fehlerhaften Mutation baute nicht und zählte „20 von 20“; ich habe ihn verworfen und die drei Mutationen neu gefahren, die Zahlen oben sind aus den gültigen Läufen.)

## 3. DoD, Plan gegen Code (`AGENTS.md` §3.9)

| DoD-Punkt (Plan §2) | Befund |
|---|---|
| 1 Lesen, Interpretieren, Zusammenführen | erfüllt; Auswahl der Antworten (Nr. 8) über die Klassen gefangen bis auf V2-1 und V2-2 |
| 2 Coverage nur durch Beweis | erfüllt (unverändert seit der ersten Verifikation: Produktcode gleich `c99b178`, `git diff 2e3a250..HEAD -- src` berührt nur `_test.go`) |
| 3 Manifest, Veröffentlichen, Idempotenz, Codes | erfüllt |
| 4 `make gates` grün | erfüllt, siehe 4 |
| 5 Review, Report | erfüllt (`5388d67`, `2e3a250`); dieser Report ist die zweite Verifikation |
| 6 Doku-Update | erfüllt (`SPEC-127` Nr. 8 g, 12 e, 18, *Ungeprüft*, Änderungstabelle; `make docs-check` im Gate 273 Dateien, 0 Befunde) |
| 7 bis 10 Closure-Notiz, Register, Risiken, Paarungen | offen, Sache des Planners (§7 *Wird bei der Closure gefüllt*, §6 Risiken mit Ausgang *offen* bis auf die Zeilen mit *entfallen*) |

**Plan gegen Code.** Der Produktcode ist seit `c99b178` unverändert (`git diff 2e3a250..HEAD --stat -- src`: nur `_test.go`). Die Dateitabelle (§3) und die Zeilen B7, C1, D1, E1, E3, E5, H6, H10, H13, I2, I10 nennen die Mengen des Codes. **Ein Widerspruch bleibt:** Entscheidung 19 und die Tabelle an ihrem Ende nennen (h) 8 und 4843, die Zeile E3, die Mengen und die Tests 9 und 4844 (V2-3). `export_test.go`, drei Hüllen: `AnfrageHash`, `Kanon`, `VergleicheZeilen`, `Inhaltstyp`, `ASCIIKlein` rufen die gleichnamige Funktion ohne Verhalten; die Datei ist neu im Slice (26 Einfügungen, 0 Löschungen gegen `cf1fb79`), der eine `sed -i` des Implementers (Betriebsregel verletzt, im Bericht genannt) hat nichts verfälscht: Inhalt geprüft gegen die Aufrufstellen in `inhaltstyp_test.go`, `auswahl_test.go`, `konfig_test.go`. **Abweichung (2)** `ASCIIKlein` statt `AsciiKlein`: `revive` (`var-naming`) meldet die andere Schreibweise, das Profil bleibt (`make lint`: 0 Befunde); zulässig. **Abweichung (3)** Teilung des Tests nach `maintidx`: zulässig, die Zählprüfung steht in der Summenfunktion (oben mutiert, rot).

**Gegenlesen auf Rohbytes (`AGENTS.md` §3.18).** `inhaltstyp_test.go`, `export_test.go`: `LC_ALL=C grep -nP '[^\x00-\x7F]'` trifft allein Kommentarzeilen (Umlaute, UTF-8); keine Folge aus Backslash, `u` und vier Hex-Ziffern, keine NUL, BOM, NBSP, U+2028 in den Dateien; die Werte ab U+0080 entstehen aus `\x`-Folgen und `rune` im Quelltext (`"\x7f"`, `string(rune(0x85))`, `string(rune(0xa0))`, `"\xe2\x84\xaa"`, `"\xc5\xbf"`, `"\xc4\xb0"`).

## 4. Gates, Laufzeiten, Last

Alle Läufe am Stand `70882c4`, Repo unverändert.

| Sensor | Ergebnis | Laufzeit |
|---|---|---|
| `make gates` | Rückgabecode 0 (Docker-Cache der Stufen; `migration-check: ok`) | 12 s |
| `make test`, Stufe `test` frisch (`--no-cache-filter test`) | `pruefung test: ok` | 69 s im Schritt, 80 s gesamt |
| `make test-race`, Stufe `race` frisch (`--no-cache-filter race`) | `pruefung race: ok` | 106 s im Schritt, 117 s gesamt |
| `make lint`, Stufe `lint` frisch | `0 issues.` | 30 s |
| Paket `baselineimport` unter dem Detektor mit Profil | ok, **98,9 %**; 8 Blöcke ohne Lauf (`auswahl.go` 283, 287; `kanon.go` 68; `link.go` 67, 69, 81, 135, 184), dieselben wie im ersten Report, alle als unerreichbar benannt | 36,7 s |
| `make gegenproben` am HEAD | „Eingaben unverändert“, Stempel nicht geändert | 0,5 s |
| `make gegenproben` erzwungen (Stempel im Scratchpad) | **37 Fälle, 0 rot**; Schritt der Fälle 56,7 s | 87 s gesamt |
| Suite des Pakets ohne Detektor | ok | 3 s |

Last (`/proc/loadavg`): vor dem erzwungenen Lauf 3,2 (1 Minute), danach 46,8; die Fälle laufen mit allen Kernen. Abstand zur Frist von 120 s je Fall: rund 63 s. Ein Rot war nicht.

## 5. Größe (`AGENTS.md` §3.14)

`git diff --numstat cf1fb79..HEAD` über die Pfade der Marke (`src/hexagon/ports/inbound/`, `…/services/baselineimport/`, `…/model/`, `…/ports/outbound/abbildung*.go`, `…/services/session/beenden_test.go`, `.a-check.yml`): **9974** Zeilen (Einfügungen plus Löschungen; 0 Löschungen), davon Produktcode **1802** (unverändert gegen die erste Verifikation), Tests und Fakes **8172** (`_test.go`). `harness/` + `tools/gegenproben`: **2** (der Stempel). Gleich dem Plan (§6 *Stand nach der zweiten Fix-Runde*: 9974, 1802, 8172, 2). Marke **8200**: überschritten um 1774 (121,6 %). §6 ist ehrlich: Zahl, Aufteilung, Grund (Fänger aus dem Wortlaut kosten rund 111 Zeilen je Zeile der Tabelle gegen 40 der Formel) stehen dort, ohne zu kürzen. Die Marke setzt und die Überschreitung bewertet der Planner.

## 6. Befunde

| ID | Schwere | Befund | Zuständig |
|---|---|---|---|
| V2-1 | **MEDIUM** | **Fänger führt nur eine Stelle der „allein Leerzeichen und Tabulator“: Name und Wert eines Parameters.** `SPEC-127` Nr. 8 g: „Befreit werden allein das Leerzeichen (0x20) und der Tabulator (0x09) … Dieselbe Befreiung gilt für Name und Wert eines Parameters“; „ein anderes Zeichen als Leerzeichen und Tabulator am Rand des Namens macht ihn zu einem anderen Namen“. Die Suite führt die nicht befreiten Zeichen (34, links und rechts) nur am **Medientyp** (b); am Parameter führt (e) allein die befreiten Zeichen. Grün in der ganzen Suite (alle 10 Mutationen oben): `strings.TrimSpace` am Namen (p1), am Wert (p2), an beiden (p3); U+00A0 am Rand des Namens (n10); `\v`, U+00A0, `\r\n` am Rand des Wertes (n12 bis n14); `\v`, `"`, `'` am Rand des Namens (n9, n8, n11). Das Negativ „Steuerzeichen am Rand des Namens eines Parameters“ (*Akzeptierte Negative*) nennt allein den Namen und allein Steuerzeichen; für den **Wert** gibt es weder Fänger noch Negativ, für U+00A0 und Anführungszeichen am Namen auch nicht. Die Menge ist endlich und aus dem Wortlaut gebildet (dieselben 34 Zeichen wie in (b); mal {Name, Wert}, {links, rechts}, {`charset`, `odata.metadata`}): gleiche Klasse und gleiche Fehlerart wie V-1 (a) (`AGENTS.md` §3.15 *Fänger führt jede Form der Eingabe, die der Code verschieden behandelt*; §3.17: die Klasse, nicht die genannten Elemente). Failure-Szenario: ein Refactoring ersetzt `strings.Trim(w, " \t")` durch `strings.TrimSpace(w)` (die naheliegende Änderung); `charset=utf-8` mit U+00A0 am Rand würde gelesen statt übersprungen; kein Test wird rot. Folge im Betrieb gering (die Antwort ginge an den Codec, nie ein falsches `complete`); der Wortlaut sagt „allein“ dennoch zu. Fix: ein Fänger als Schleife über die 34 Zeichen mal {Name, Wert} mal {links, rechts} mal 2 Parameter (272 Elemente, in der Zählprüfung von E3), eine Mutation an einem Element, das hier nicht genannt ist (z. B. `strings.TrimSpace` nur am Wert, U+0085 am Namen); danach die Menge in der Zeile E3 und im Bericht. Kein Produktcode. | Implementer (Fänger, Zeile E3, Bericht), Architect (nur wenn er die Klasse lieber als Negativ führt: dann mit Grund und Folge für den Wert) |
| V2-2 | LOW | **Unbeschränkte Reste der Klasse Content-Type ohne benannte Prüfgrenze** (`AGENTS.md` §3.17: ist die Menge unbeschränkt, nennt der Plan Grund und Folge). Grün: ein druckbares Zeichen am Rand des Medientyps außer den 34, etwa `"` (n7; Wortlaut „jedes andere Zeichen am Rand“); Namen, die `charset` oder `odata.metadata` nur ähneln, als Präfix und Suffix (`charsetx=iso-8859-1`, `xcharset=…`, `odata.metadata2=none`, p7 bis p10; Wortlaut „ein Parameter mit anderem Namen ist ohne Wirkung“); ein Anführungszeichen vorn mit Anhängsel hinten (`"utf-8x`) und umgekehrt (n15, n16; Wortlaut „nur als Paar“); ein Tabulator oder U+00A0 am Rand des **Namens des Headers** (p12, p13, äquivalent zu ` content-type`, das (c) führt). Keine bricht eine Zusage der Tabelle in einem Fall, den der Plan nennt; es sind Nachbarn der geführten Mengen. Fix: je ein Element mehr (billig) oder ein akzeptiertes Negativ mit Grund und Folge. | Implementer (Plan, Fänger), Architect (Folge im Wortlaut) |
| V2-3 | LOW | **Entscheidung 19 und die Tabelle an ihrem Ende widersprechen Zeile E3, den Mengen und dem Code** (`AGENTS.md` §3.9: eine berichtigte Aussage wird an allen Stellen nachgezogen). Entscheidung 19 und die Zeile V-1: (h) 8, Summe 4843; die Zeile E3, die Mengen, der Test und der Wortlaut: 9 und 4844 (die Zählung oben: neun verschiedene Formen, keine doppelt). Der Implementer hat angehalten und die Adresse in §6 eingetragen (`AGENTS.md` §3.12 erfüllt); der Architect hat den Punkt noch nicht entschieden. Die Zahl im Wortlaut der Spezifikation fehlt (Nr. 8 g nennt keine Zahl), die Spezifikation widerspricht nicht. Fix: im Text der Entscheidung 19 und in der Tabelle „8“ durch „9“ und „4843“ durch „4844“ ersetzen. | Architect |
| V2-4 | INFO | **Die Zahl der Dateien von I10 (11) hält kein Test.** `n != 7*len(dateien)` rechnet mit der Zahl, die das Verzeichnis liefert; der Plan nennt 77 = 7 mal 11. Wächst das Paket um eine Datei, bleibt der Test grün und der Plan nennt 77 weiter. Die Wirkung auf die Fänger ist keine: jede Datei wird geprüft. | keiner (Zahl im Plan beim nächsten Berühren nachziehen) |
| V2-5 | INFO | **Betriebsregel verletzt (`sed -i` auf `export_test.go`)**, vom Implementer berichtet. Geprüft: nichts verfälscht (siehe 3). | — |
| V2-6 | INFO | Ein **leeres Verzeichnis `w/`** liegt im Repo-Wurzelverzeichnis (Änderungszeit 12:35, vor Beginn meiner Sitzung; leer, git sieht es nicht). Es stammt nicht von mir; vermutlich ein Mount-Punkt eines früheren Laufs. Ich habe es nicht angefasst. | Planner oder User (entfernen) |
| V2-7 | INFO | Größe 9974 gegen Marke 8200 (121,6 %), Aufteilung und Grund in §6 (siehe 5). | Planner (Messung, Bewertung) |

**Gültige Seite:** keine Verletzung der Spezifikation im Produktcode, kein HIGH. Der Produktcode ist unverändert und in allen gefahrenen Klassen richtig (4 499 513 + 1 082 401 Header-Fälle gegen ein unabhängiges Orakel, 0 Abweichung; die Probe-Pakete der ersten Verifikation sind nicht erneut gefahren, der Produktcode ist unverändert).

## 7. Register (nichts geschrieben): Kandidaten für den Lese-Schritt der Closure

Unter `docs/plan/planning/observations/BEO-REPO/` gibt es zu diesem Slice keine Evidence-Datei dieses Reports. Vorzumerken (Planner entscheidet):

1. `fix-deckt-die-genannte-mutation-nicht-die-klasse` (verkörpert in `AGENTS.md` §3.17): zweiter Fall im selben Slice. V-1 selbst ist gegen die Klasse gelöst (34 von 34 Mutationen rot, 9 genannte und 25 nicht genannte); V2-1 ist die **Nachbarstelle derselben Zusage** („allein … dieselbe Befreiung“): der Fix deckte die Stelle, die der Befund nannte, nicht die zwei anderen Stellen derselben Aussage. Gegenbelege: F-1, F-2, F-4, F-5, F-8 bis F-13, V-4, V-6, V-8 hielten.
2. `universalaussage-im-vertrag-ungeprueft`: „allein das Leerzeichen und der Tabulator“ (drei Stellen, V2-1).
3. `akzeptiertes-negativ-ohne-folge` / `unbeschraenkte-menge-ohne-pruefgrenze`: V2-2 (Reste), dritter Fall im Slice.
4. `suite-ohne-generator-fuer-klassen-der-spezifikation` (2×, offen): Gegenbeleg — die Fix-Runden schlossen die Lücken mit Generatoren (E3 4844, H6 4531, H10, H14); mein Orakel über 5,5 Millionen Fälle fand im Produktcode nichts.
5. `berichtigung-erreicht-nicht-jede-stelle`: V2-3 (Entscheidung 19 und Tabelle gegen Zeile E3), V2-4 (die Zahl 11).
6. `stempel-gehoert-zu-keinem-committeten-stand` (1×): **Gegenbeleg** — der Stempel in `70882c4` gehört zum Stand der Eingaben am HEAD (`make gegenproben`: „Eingaben unverändert“, der erzwungene Lauf schrieb denselben Stempel).
7. `nebenlaeufiger-lauf-nur-im-glueck-gruen`: Gegenbeleg — Mutex-Mutationen 20 von 20, Selbsttest der Barriere 20 von 20, benanntes Negativ 0 von 20.
8. `hilfe-ersetzt-argument-verdeckt-eigenschaften` (1×, offen): V-6 gelöst (Selbsttest der Barriere), Beleg der Gegenrichtung.
9. `groessenmarke-ueberschritten-ohne-rueckfuehrung` (15×): nächster Messpunkt, 9974 bei Marke 8200.
10. `betriebsregel-fuer-repo-dateien-verletzt` (neu, 1×, INFO): `sed -i` auf eine Repo-Datei trotz Regel; berichtet, ohne Folge (V2-5).

---

## Verdikt

**Nicht bestanden.** Gelöst und belegt: V-1 in den Teilen, die Verifier 1 nannte (Rand des Medientyps, Name des Headers in 2048 Schreibweisen, Leerraum um den Wert; 9 von 9 genannte und 25 von 25 nicht genannte Mutationen rot; alle Klassen im Lauf gezählt, 4844), V-2, V-3, V-4, V-5, V-6, V-7, V-8, V-9; Gates grün (`make gates` 12 s, Stufen `test` 69 s, `race` 106 s, `lint` 30 s frisch ohne Cache), Gegenproben 37 Fälle ohne Rot (56,7 s), Abdeckung 98,9 %, Raten 20 von 20 beziehungsweise 0 von 20 beim benannten Negativ, Stempel gehört zum Stand. **Offen ist ein MEDIUM: V2-1** — die Zusage „allein Leerzeichen und Tabulator“ hat den Fänger für die nicht befreiten Zeichen nur am Medientyp; am Namen und am Wert eines Parameters bleiben zehn Mutationen, darunter `strings.TrimSpace`, in der ganzen Suite grün.

**Was fehlt, und wer es tut.**
- **Implementer:** (V2-1) ein Fänger als Schleife über die 34 Zeichen mal {Name, Wert} mal {links, rechts} mal 2 Parameter, in die Zählprüfung von E3, Zeile E3 im Plan und Bericht mit Menge, Zahl, einer Mutation an einem nicht genannten Element; (V2-2) je ein Element mehr oder Negativ mit Grund und Folge. Kein Produktcode.
- **Architect:** (V2-3) die Zahlen 9 und 4844 in Entscheidung 19 und ihrer Tabelle; (V2-1) entscheidet nur, wenn er den Wert als Negativ führen will (nicht empfohlen: die Menge ist endlich).
- **Planner:** (V2-6) das leere Verzeichnis `w/`, (V2-7) die Größe, das Register nach 7; die Closure (DoD 7 bis 10) nach der dritten Verifikation.

**Übergabe:** an den **Implementer** (V2-1, V2-2), danach eine kurze dritte Verifikation: die zehn Mutationen p1 bis p3, n9 bis n14 rot, die vier Ähnlichkeits-Mutationen p7 bis p10 rot oder als Negativ benannt, `make gates` grün; der Produktcode bleibt, also entfallen die Probe-Pakete.
