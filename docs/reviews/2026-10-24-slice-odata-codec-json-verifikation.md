# Verifikations-Report: slice-odata-codec-json — 2026-10-24

**Rolle:** Verifier (Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-126`, `AGENTS.md` §3.9 bis §3.17). Kein Review, keine Reparatur.

**Gegenstand:** `slice-odata-codec-json` (Welle `welle-record-replay`), HEAD `56ad5ec`; Basis der Marke `bc05c62`. Geprüft: Review-Report `64b7cc9` (`docs/reviews/2026-10-24-slice-odata-codec-json-review.md`), Entscheidungen des Architect `58cc559`, Fix-Runde `811c730`, Stempel `56ad5ec`. Der Produktcode (`jsoncodec`, `model`, Port) ist seit `0acc49a` unverändert (`git diff 0acc49a..HEAD -- src/adapters src/hexagon/model` leer); die Fix-Runde ändert Suite, Plan, Spezifikation (Architect) und den Stempel.

**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-24

**Methode.** Belege der Behauptungen statt Behauptungen: alle Sensoren frisch gefahren (Docker-Cache der Stufen mit `--no-cache-filter` umgangen, sonst zeigt `make gates` nur den Cache); eigene Mutationen in Wegwerfkopien (Quelle `:ro`, `--network none`, `--memory` 2g oder 4g, Mutationen ohne Grenze mit 2g und Zeitgrenze, nach der Anweisung des Planners später `--cpus=4`, höchstens zwei Container, nacheinander); ein Probe-Paket aus Generatoren (Reviewer-Paket aus dem Scratchpad plus vier neue Klassen des Verifiers). Scratchpad: `/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/vf2/` (flüchtig; `base/src/revtest/verif_test.go` ist das neue Paket, `runner/v1.txt`, `v2.txt`, `v3.txt` die Mutationen). Das Repo ist bis auf diesen Report unverändert. Ein Versehen: ein Kommando hat den Archivbaum kurz nach `/tmp` entpackt; die entpackten Dateien sind wieder entfernt (nur Dateien aus dem Archiv, Verzeichnisse nur wenn leer).

---

## 1. DoD-Konformität

| DoD-Punkt (Plan §2) | Befund |
|---|---|
| 1 Port `ODataCodec`, acht EDM-Typen kanonisch | erfüllt: Code, `SPEC-126` Nr. 10 und die Zeilen F1 bis F11 stimmen; das Probe-Paket (siehe 3) findet im Code keine Abweichung |
| 2 Kanonischer Schlüssel, `ODR-E6101` für ungültigen Schlüssel, Typfehler, widersprüchlichen Inhalt | erfüllt (G1 bis G3, H3, Probe: alle 40 320 Permutationen des Reviewers weiter ohne Abweichung) |
| 3 `@odata.context`, `nextLink`, `count`, `null` gegen fehlend | erfüllt (D1, D2, E1, E2, F9) |
| 4 `make gates` grün | erfüllt, siehe 4 |
| 5 Review durchgeführt, Report liegt vor | erfüllt (`64b7cc9`); die Fix-Runde ist nicht erneut begutachtet, das ist dieser Report |
| 6 Doku-Update bei öffentlichem Vertrag | erfüllt: `SPEC-126` Nr. 1 f, 3, 13 a, 14 tragen die Entscheidungen von `58cc559`; Code und Spezifikation stimmen (gelesen) |
| 7 bis 10 Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | noch nicht: §7 trägt „Wird bei der Closure gefüllt“, jedes Risiko in §6 steht auf „offen bis zur Closure“; das ist Sache des Planners nach diesem Report, kein Mangel des Implementers |

Die Zahl der Liefer-Punkte (drei) ist eingehalten.

## 2. Befunde F-1 bis F-8 gegen die Klasse (`AGENTS.md` §3.17)

Mengen nachgezählt: die Suite zählt sie selbst und schlägt bei Abweichung an (`gZaehle`, `len(…) != …`); ich habe `-v` gefahren und die Zahlen gelesen, sie stimmen mit dem Bericht des Implementers.

| Befund | Menge und gezählte Elemente | Eigene Mutation an einem Element, das weder Review noch Implementer nannten | Urteil |
|---|---|---|---|
| F-1 (B2) | G1 597 870, G2 960 799, G3 266 320 (71 610 + 133 455 + 61 255), G4 269; zusammen 1 825 258; 3 + 1 Körper `doppelter_name` | Zahl endet im Zustand `1e+` (`ende()` nimmt `zExpVz` als vollständige Zahl): **grün** (VM1). Gleiche Mutation für `zPunkt`, `zExpE`, `zMinus`: rot durch B1 (`1.`, `1e`, `-`). Rot außerdem: Komma vor `]` (VM16), `1e5.5` (VM17) | **Klasse nicht gedeckt:** V-1 |
| F-2 (B3) | GZ 76, U 396 772 (256 + 49 284 + 1 632 + 345 600), im Lauf gezählt | innere Fortsetzungsbytes `0xA5` im zweiten Byte einer Drei- oder Vier-Byte-Folge: **grün** (VM5); `0xB7` im letzten Byte einer Zwei-Byte-Folge: rot (VM6) | wie geplant: das akzeptierte Negativ „Fortsetzungsbyte zwischen den Grenzen“ steht mit Grund und Folge im Plan; V-7 (INFO) |
| F-3 (B3) | A2 1 027 (4 + 4 × 255 + 3) | `EF` mit `BC` statt `BB` als zweites Byte angenommen: **grün** (VM18), weil A2 nach dem abweichenden Byte den Strom beendet; `FF` mit `FD` angenommen: rot (VM19) | **Klasse nicht ganz gedeckt:** V-5 |
| F-4 (F5) | 90 (5 × 2 × 3 × 3), gezählt | Länge des reinen Bruchteils um eins zu groß (`3 - z.exp`, VM13), ganzer Teil um eins (VM14), Vorzeichen nur bei `exp <= 0` gezählt (VM15): alle rot | gedeckt |
| F-5 (A1) | 2 Aufrufe (beide Methoden) | (nicht erneut mutiert; der Fall ist im Lauf, `Set X` ohne Typ) | gedeckt |
| F-6 (K1) | 390 Namen (3 × 128 + 6) | `é` wird `~2`: rot (VM7); `U+3042` oder `U+1F600` wird `~2`: **grün** (VM8, VM9) | **Klasse nicht ganz gedeckt:** V-4 |
| F-7 (A5) | 21 819 Fälle plus 2 mit Deckel | `want` ohne `rest + 1` (VA5a) rot mit 5096 Bytes statt 1001; `ueber := false` (VA5b) rot an den Reads mit leerem Puffer (9999), nicht an der Zeitgrenze; `want = int(rest)` rot in A5 (B5 hängt danach bis zur Zeitgrenze, nach A5) | die genannte Mutation gedeckt; die Klasse „nach einem **Fehler** nicht weiter lesen“ nicht: V-2 |
| F-8 | Zahlen und H1/H2 | Plan nachgemessen: A4 21 811 / 4 151 100, H3 15 284, H4 6 535, G-Mengen, 270 und 281 stimmen mit meinem Lauf überein; L1 94 und 37 MiB (Plan 83 und 47: „schwankt um einige MiB“, nicht vergleichbar) | gedeckt |

Zusätzlich aus dem Wortlaut der Spezifikation gebildete Klassen (Mengen vom Verifier):

- **Hex-Ziffer eines `\u`-Escapes** (22 Zeichen, beide Schreibweisen): je eine Mutation „Ziffer wird abgelehnt“, **22 von 22 rot** (`v3.txt`).
- **Grenze der Tiefe, des Limits ohne Grenze** (Speicherwächter): siehe 5.
- Die 243 Mutationen des Reviewers (`m1` bis `m16`) gegen den neuen Stand: 211 rot, 17 `emit` (kein Lauf), 13 grün, 1 Zeitgrenze, 1 Baufehler (eine Mutation des Reviewers gegen ihre Fixture, kein Befund). Die 13 grünen sind die vom Reviewer genannten äquivalenten Mutationen und das akzeptierte Negativ L1 (+1 KB je Element, `fertig`, `MaxInt64`, `namenLinear`, `Properties: nil`, `Replacer` des Pfads, Boolean-aus-String durch F2); keine der 14 überlebenden Mutationen des Reviews (F-1 bis F-4, F-6) lebt noch. Es gibt keinen Rückfall.

### Probe-Paket (Orakel unabhängig vom Leser)

Der Leser importiert `context`, `errors`, `io`, `strconv`, `strings`, `time`, `unicode/utf8` und `model`; `encoding/json` steht allein in der Suite (`odata_codec_generator.go`: `json.Valid` und `json.Decoder` als Orakel), nie im Leser. Alle Läufe ohne Abweichung am Code:

| Paket | Läufe | Abweichungen |
|---|---|---|
| Reviewer-Paket (23 Tests: Zahlen, Guid, Date, DateTimeOffset, Schlüssel, Kontext, Paare, Kodierung, Grammatik, Pointer, Zerlegung, Grenzen), unverändert gegen den neuen Stand, 78 s | rund 8,8 Mio. | 0 |
| V1 alle Drei-Byte-Folgen (51 Lead-Bytes × 222 × 222, ohne `"`, `\` und Steuerzeichen) und alle Vier-Byte-Folgen (`F0` bis `F4`, Fortsetzungsbytes `0x80` bis `0xFF`), Orakel `utf8.Valid` | 12 999 244 (2 513 484 + 10 485 760) | 0 |
| V2 Dezimalgitter: 15 Mantissen (Nullen vorn und hinten) × Exponent −1050 bis 1050 × Vorzeichen × Zahl und String, Orakel `math/big`, Länge 1024 beider Seiten | 126 060 | 0 |
| V3 jedes Unicode-Zeichen U+0080 bis U+10FFFF (ohne Surrogate) als Name einer unbekannten Property, Ort `"/" + Name` | 1 111 936 | 0 |
| V4 Puffergrenze: Bodys der Länge 32 768, 65 536, 98 304 ± 12, `grenze` aus {L−1, L, L+1, 32 767, 32 768, 32 769, 65 535 bis 65 537}, ganz und in Zweiteilungen um die Grenze | 6 750 | 0 |

Der Code ist also in diesen Klassen richtig. Die Befunde unten betreffen Fänger, nicht den Code.

## 3. Befunde des Verifiers

| ID | Schwere | Befund | Zuständig |
|---|---|---|---|
| V-1 | MEDIUM | **Klasse „jeder Verstoß der Grammatik ist `syntax`“ (SPEC Nr. 4) hat eine Lücke am Ende des Stroms.** Die Mutation „`ende()` nimmt `zExpVz` als vollständige Zahl“ (`1e+`, `1e-` als gesamter Body) bleibt in der ganzen Suite grün; der Body wäre `struktur` statt `syntax`. B1 führt `-`, `1e`, `1.`, aber nicht `1e+`; G2 setzt die Zahl in `[` `]`, G1 hat kein `e`, G3 geht von Objekten aus. Die Begründung der Prüfgrenze im Plan (§3 *Akzeptierte Negative*, Grammatik: „jeder Zustand kommt in einem der gültigen Dokumente von G3 vor, so dass eine Einzeländerung dort jeden Übergang trifft“) ist für die Übergänge am Ende des Stroms falsch. Wirkung klein (nur ein Body, der allein aus einer Zahl besteht, beide Ursachen sind `ODR-E6101`), aber die Zusage „jeder“ ist nicht gefangen. Fix: `1e+`, `1e-` in B1 und G2 auch ohne Klammern (Wurzelzahl). | Implementer (Fänger), Architect (Begründung der Prüfgrenze berichtigen) |
| V-2 | MEDIUM | **Zusage „liest nach `io.EOF` **oder einem Fehler** nicht weiter“ (SPEC Nr. 2) ist nur für `io.EOF` gefangen.** Die Mutation „bei einem Fehler von `src` noch einmal lesen“ (`VR1`) bleibt in der ganzen Suite grün (A3 prüft nur den Fehler, nicht die Zahl der Reads; die Zähler `nachEnde` und `a5Strom` stehen in A5, und A5 fährt nur Ströme ohne Fehler). Die Mutation „Fehler von `src` ignorieren und weiterlesen“ (`VM10`) hängt in A3 bis zur Zeitgrenze (60 s bei mir, 10 Minuten ohne Angabe), rot nicht aus dem benannten Grund (wie F-7 im Review). Gegenprobe: das Lesen nach `io.EOF` (VR2) ist rot (A5, „Reads nach dem Ende 1“). Fix: in A3 `rd.nachEnde == 0` je Fall prüfen und dem Fehler-Strom wie `a5Strom` eine Grenze der Reads geben. | Implementer |
| V-3 | MEDIUM | **B5 fängt die Grenze nur bei kleinen Bodys; der Leser hat bei der Puffergröße (32 KiB) eine zweite Grenze (`rest < want`).** Die Mutation `rest <= int64(want)` (VM11) bleibt in der ganzen Suite grün; sie macht bei `grenze − gelesen == 32 768` den Slice `puffer[:32769]` zur Panik, also bei der Standardgrenze von 8 MiB (256 × 32 768) für Bodys ab rund 8 MiB − 32 KiB. Mein Paket V4 (Bodys um 32 768, 65 536, 98 304 mit den Grenzen L−1, L, L+1 und 32 767 bis 32 769) fährt dieselbe Mutation rot und den Code ohne Abweichung. `AGENTS.md` §3.15: „beide Seiten jeder Grenze“. | Implementer |
| V-4 | LOW | **K1 (F-6): die Menge „jedes Zeichen ab `0x80`“ ist mit 6 Zeichen abgetastet, ohne Prüfgrenze mit Grund und Folge im Plan.** `U+3042` und `U+1F600` als `~2` (VM8, VM9) bleiben grün; `é` (VM7) ist rot (E5). Die Menge ist endlich und billig: V3 fährt alle 1 111 936 Zeichen in 3,8 Sekunden. | Implementer (Schleife über die Menge) oder Architect (Negativ mit Grund und Folge) |
| V-5 | LOW | **A2 (F-3): „abweichendes Folgebyte“ ist nur am Ende des Stroms geprüft.** `EF BC` als BOM-Fortsetzung (VM18) bleibt grün, denn erst `EF BC BF` unterscheidet `syntax` von `kodierung`. Fix: je Präfix und abweichendem Byte ein Fall mit angehängtem fortsetzenden Byte. | Implementer |
| V-6 | LOW | **Plan und Bericht stimmen nicht an vier Stellen (`AGENTS.md` §3.9).** (a) Die Tabelle *Ausgang des Architect zum Review*, Zeile F-7, sagt „die Mutation endet am Deckel“; sie endet an den Reads mit leerem Puffer (Bericht Abweichung 1; die Zeile A5 und der Absatz *Speicherwächter* sind richtig). (b) Bericht Abweichung 6 nennt „Der Leser nutzt `encoding/json` nicht“ eine Zusage des Plans; Plan §3 Auftrag 7 (b) und `SPEC-126` Kopf lassen die Wahl offen, es gibt weder eine Zusage noch einen Fänger; `.a-check.yml` hat zu `encoding/json` keine Regel. Die Unabhängigkeit des Orakels `json.Valid` hängt an dieser Eigenschaft und wird von keinem Gate gehalten (siehe INFO V-8). (c) Offene Entscheidungen stehen nur im Bericht: ob ein Wächter auf `HeapAlloc` die richtige Größe misst (Bericht Abweichung 7), und dass G1 bis G4 die Methode mit der Nummer des Körpers wechseln und U nur `ParseCollection` fährt (Abweichung 5). Weder §6 (Ausgang *offen*, Adresse *Architect*) noch *Akzeptierte Negative* tragen sie. (d) `jsuite.drossel` und ihre Wirkung stehen weder in der Dateitabelle noch im Absatz *Speicherwächter* von §3, noch als Risiko in §6. | Planner (a, c), Implementer (b, d), Architect (c) |
| V-7 | INFO | Das akzeptierte Negativ „Fortsetzungsbyte zwischen den Grenzen“ (Drei- und Vier-Byte-Folgen) ist als solches richtig dokumentiert; die Mutation VM5 zeigt, dass es eintritt. Die Drei-Byte-Menge ließe sich ganz fahren (V1: 2,5 Mio. Folgen in rund 14 Sekunden). | Architect (Entscheidung) |
| V-8 | INFO | Die Zusage im Kopf von `leser.go` („nutzt allein die Standardbibliothek“) und im SPEC-Kopf („keine Fremdbibliothek“) hat kein Gate außer der Durchsicht von `go.mod`. Kein Befund, weil der Vertrag nichts anderes zusagt. | — |
| V-9 | INFO | `harness/gegenproben.stempel` ist vom Slice geändert (eine Zeile); der Plan sagt „`harness/` nur, wenn der Slice sie ändert, und dann getrennt gemeldet“, §6 meldet es nicht. Wirkung null (2 Zeilen). | Planner |

## 4. Gates, Laufzeiten, Last

Alle Läufe am Stand `56ad5ec`, Repo unverändert.

| Sensor | Ergebnis | Laufzeit |
|---|---|---|
| `make gates` | Rückgabecode 0 (zweimal; der Docker-Cache der Stufen `build`, `test`, `race`, `lint` ließ ihn in 11 Sekunden enden, darum die frischen Läufe unten) | 11 s |
| `make test` (gleicher Befehl, `--no-cache-filter test`) | `pruefung test: ok` | 71 s (Schritt 59 s) |
| `make test-race` (`--no-cache-filter race`) | grün | 124 s |
| `make lint` (`--no-cache-filter lint`) | grün | 32 s |
| Gegenproben (`docker build --no-cache-filter gegenproben`, die Zeile des Targets; `make gegenproben` selbst meldet „Eingaben unverändert“ und fährt nichts) | `37 Fälle, 0 rot` | 144 s (Schritt 109 s) |
| Vertragssuite allein (`go test ./src/revtest -v`) | grün | 35 s |

**Zu 2 (b), Speicherwächter und Last.** `pruefung test` lief bei einer Last (`/proc/loadavg`) von 19 bis 35 auf 20 Kernen zweimal grün (122 s und 113 s statt 59 s). Die Gegenproben liefen bei einer Last von 67 bis 131 (drei CPU-lastige Mutationscontainer des Verifiers daneben) **rot: 2 von 37 Fällen, `test/basis` und `test/roter-test`: „Zeit überschritten“** (die Zeitgrenze von 120 s je Fall, der Speicherwächter feuerte nicht). Bei mäßiger Last (unter 20) grün. Die Drossel hält also den Heap, aber die Laufzeit der Suite (59 s ohne Last) wächst unter Last auf mehr als das Doppelte und kommt an die Zeitgrenze der Gegenprobe. Das ist die Grenze der Abhilfe, kein Befund gegen den Code; die Last war hier von mir verursacht und höher als im Betrieb zu erwarten. Der Planner sollte wissen, dass `make gegenproben` bei Last jenseits etwa dem Dreifachen der Kernzahl an der Zeit scheitert (Register `rate-ohne-detektor-schwankt-mit-der-last`, Kandidat für einen weiteren Beleg).

Der Wächter fängt weiter die Mutationen „Grenze entfernt“ aus dem benannten Grund: **Exponent ausgeschrieben für `Edm.Int32`** (LF3: „Heap 956 MiB über 512 MiB in der Zeile F3 Int32“), **Ausschreiben vor der Prüfung für `Edm.Decimal`** (LF5: 956 MiB, Zeile F5 Decimal); **Tiefe entfernt** (LB4, B4) endet im Fall (`tiefe` erwartet, Body gelesen), nicht im Wächter (wie im Bericht); **L1**: die Kopie je Element und je Member der Reviewer-Mutationen (L1b, L1e) sind rot, +1 KB je Element (L1c, L1d) bleibt grün (akzeptiertes Negativ des Architect). Alle mit `--memory=2g` und Zeitgrenze.

**Raten (Stichprobe, Detektor `golang:1.27.1-trixie` für J1, sonst Image `golang:1.27.1`-Stufe des Reviewers; jeder Prozess `-test.count=1`; Messender: Verifier):**

| Zeile | Mutation | Prozesse rot |
|---|---|---|
| J1 | Zähler des Pakets in `lies`, Detektor | 10 von 10 |
| J2 | Puffer des Pakets unter Sperre, ohne Detektor | 10 von 10; 3 Prozesse mit `-test.count=20`: 60 von 60 Läufen |
| I1 | `Properties` über Map, ohne Detektor | 10 von 10; 3 Prozesse mit `-test.count=20`: 60 von 60 Läufen |

Die Zahlen des Implementers (30 von 30, 20 von 20, 100 von 100) sind nicht widerlegt.

## 5. A5 und die Abweichungen (Punkt 2 des Auftrags)

(a) **A5, „Größengrenze entfernt“.** Bestätigt: mit `ueber := false` wird `rest` negativ, `want` null, der Leser ruft `Read` mit leerem Puffer ohne Fortschritt; der Wrapper `a5Strom` beendet das nach 10 000 Reads mit einem Fehler (VA5b rot mit „Reads mit leerem Puffer 9999“). **`SPEC-126` ist wahr** (sie nennt weder Deckel noch Wächter für den Leerraum); der Plan ist wahr bis auf die Zeile F-7 der Entscheidungstabelle des Architect (V-6 a). Der Architect muss nur diese Zeile berichtigen oder den Satz „endet am Deckel“ mit dem Zusatz der Abweichung versehen; die Zeile A5 und der Absatz *Speicherwächter* sind richtig.
(b) siehe 4. (c) Die vier Körper `doppelter_name` sind im Plan, Zeile B2, schon so entschieden („3 in G1, 1 in G3“); der Bericht nennt sie „strenger als ausnehmen“, der Plan ist die Quelle: keine Abweichung. Die Zahlen stimmen (3 in G1, 1 in G3 im Lauf). (d) Wechsel der Methode und `U` nur `ParseCollection`: wie im Bericht, Folge klein, weil `vorbereiten` und `lies` beiden Methoden gemeinsam sind; als akzeptiertes Negativ nicht im Plan (V-6 c). (e) Siehe V-6 b und V-8.

## 6. Größe (Punkt 3)

`git diff --numstat bc05c62..HEAD` über die Pfade der Marke (`src/hexagon/model/`, `src/hexagon/ports/outbound/`, `src/adapters/driven/odata/jsoncodec/`, `src/bootstrap/`, `beenden_test.go`, `.a-check.yml`): **6500** Zeilen (6498 Einfügungen, 2 Löschungen), davon Produktcode **1640**, Tests und Suite **4860**; gleich dem Bericht. Die Aufteilung der +888 seit dem Handoff (611 + 67 + 71 + 103 + 28 + 8) summiert sich. Marke 4500: +44,4 %. **§6 weist die Überschreitung ehrlich aus** (Zahl, Aufteilung Produktcode/Tests, Grund, „Marke nicht geändert“); §4 trägt den gemessenen Wert nicht, das ist nach §3.14 zulässig („nach dem Code gemessen und mit Grund berichtet“). `harness/gegenproben.stempel` ist nicht Teil der Metrik, aber vom Slice geändert (V-9).

## 7. Register (Punkt 5, nichts geschrieben)

Unter `docs/plan/planning/observations/BEO-REPO/` gibt es zu diesem Slice **keine** Evidence-Datei und der Plan (§7 „Wird bei der Closure gefüllt“, §8) merkt keine vor; nur der Review-Report (I-6) nennt drei. Für den Lese-Schritt der Closure sind vorzumerken (Planner entscheidet): `suite-ohne-generator-fuer-klassen-der-spezifikation` (Beleg: 14 Überlebende im Review; nach den Generatoren weiter V-1, V-3, V-4, V-5 — damit 2×); `werkzeug-eingabe-setzt-escape-in-zeichen-um` (Review I-6: dritter Beleg, löst die Verkörperung beim Lese-Schritt aus); `fix-deckt-die-genannte-mutation-nicht-die-klasse` (V-1, V-4, V-5: auch dieser Fix deckt die Klasse nicht ganz; 4. Slice); `groessenmarke-ueberschritten-ohne-rueckfuehrung` (zweiter Messpunkt 5612 am Handoff, 6500 nach der Fix-Runde, Marke 4500); `befund-an-rolle-ohne-artefakt` (V-6 c: offene Entscheidungen nur im Bericht; wäre der 3. Beleg); `rate-ohne-detektor-schwankt-mit-der-last` (Abschnitt 4: Gegenprobe an der Zeitgrenze unter Last; wäre der 3. Beleg); `berichtigung-erreicht-nicht-jede-stelle` (V-6 a, F-8); `akzeptiertes-negativ-ohne-folge` (F-3 des Reviews, V-6 c).

---

## Verdikt

**Nicht bestanden (nachzubessern, kein Neubau).** Der Produktcode ist in allen geprüften Klassen richtig (rund 22 Mio. Läufe des Verifiers und 8,8 Mio. des Reviewers ohne Abweichung gegen unabhängige Orakel), alle Gates sind frisch gefahren grün, die DoD-Punkte 1 bis 6 sind sachlich erfüllt. Nicht bestanden, weil die Fix-Runde die Klassen nach `AGENTS.md` §3.17 nicht ganz deckt: **drei MEDIUM** (V-1 Ende des Stroms in einer Zahl, V-2 Lesen nach einem Fehler von `src`, V-3 Grenze an der Puffergröße), **drei LOW** (V-4, V-5, V-6), drei INFO. Keine Verletzung der Spezifikation, kein HIGH.

**Übergabe.** An den **Implementer**: V-1 (B1 und G2 ohne Klammern), V-2 (A3: `nachEnde` und Grenze der Reads), V-3 (B5 und L1: Bodys und `grenze` um 32 768), V-4 (Schleife über alle Zeichen), V-5, V-6 b und d; jeder Fix als Klasse mit einer Mutation an einem Element, das hier nicht genannt ist. An den **Architect**: V-1 (Begründung der Prüfgrenze), V-4 und V-7 (Menge oder Negativ), V-6 c (HeapAlloc-Wächter). An den **Planner**: V-6 a und c, V-9, die Lage der Gegenprobe unter Last, die Register-Merkliste (7) und die Closure.
