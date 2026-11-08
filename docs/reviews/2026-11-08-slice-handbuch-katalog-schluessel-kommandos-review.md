# Review: slice-handbuch-katalog-schluessel-kommandos — 2026-11-08

**Art:** Review (einmal), Code — gegen Plan, `SPEC-201` (Nr. 4 bis 8, 9 e und h, 10 a, 11; G16 bis G31), `SPEC-102`, `LH-QA-05`, `LH-RB-01`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md) und die Hard Rules; nicht gegen die DoD (Verifier).

**Gegenstand:** Zweig `slice-handbuch-katalog-schluessel-kommandos`, HEAD `233308a1`, Basis `6dbe5d02` (Implementer `8907ddb3`, `48832016`, `023c06ce`; Architect-Entscheid `233308a1`). Plan: `slice-handbuch-katalog-schluessel-kommandos` (§3 Tabelle 22 Zeilen, *Mutationen und Raten*, *Mengen*, §6).

**Rolle/Modell:** reviewer · claude-sonnet-5-5.

## Urteil

Kein HIGH. Der Leser (`schluessel.go`, `kommandos.go`) liest die Verträge von `SPEC-201` Nr. 4, 5, 7, 8 wortgetreu; keine Zahl der Codes oder Schlüssel ist im Programm verdrahtet, die Fixtures sind Wegwerfbäume, das Gate ist nicht gelockert, `//nolint` kommt nicht vor.
Von 84 eigenen Mutationen (68 gegen die Suite, 16 gegen den echten Baum; Programm in Wegwerfkopie, ein Container zur Zeit unter `flock`) blieben 34 gegen die Suite grün; 7 davon sind gleichwertig oder ohne Wirkung (`A4`, `A6`, `K2`, vier Knotenarten ohne Literal), die übrigen 27 zeigen die Lücken der Findings F-1 bis F-6.
Die Findings sind sieben MEDIUM (Fänger schmaler als die Zusage oder Zusage ohne Fänger, eine Zeitaussage, die ihre Zahlen nicht trägt), drei LOW und zwei INFO.
Zur Budgetfrage der Gegenproben (169 s gegen 120 s) sage ich nichts; F-7 betrifft nur den Wortlaut der Messung.

## 1. Eigene Belege

| Prüfung | Beleg |
|---|---|
| Programm ohne Mutation | `go test -count=1 ./...` in `tools/katalog` (Image `golang:1.27-alpine` des `Dockerfile`, `--network none --memory 3g --cpus 4`): `ok`. |
| Gate am echten Baum | `katalog` gegen eine Kopie der sechs echten Quellen: `katalog: ok`. |
| Stempel | `bash tools/harness/gegenproben-eingabe.sh <3 Pins aus make -n gegenproben>` gibt `1445b173…a6c`, gleich `harness/gegenproben.stempel`. |
| Keine feste Zahl | `grep` über `tools/katalog/*.go` ohne Tests: kein `26`, `27`, `SERVICE_ODATA_VERSION`; die Zahlen stehen in den Tests als `zaehle(...)` über Fixture-Listen, kein Test liest `src/`, `model.AlleCodes()` oder die echte Ausnahmen-Datei. |
| `AGENTS.md` §3.2 | `git diff 6dbe5d02..HEAD` auf `nolint`: kein Treffer; `.golangci.yml`, `.a-check.yml`, `.dockerignore` nicht berührt. |
| `AGENTS.md` §3.6 | kein Gate gelockert: `Dockerfile` nur um zwei `COPY` und Kommentar erweitert, `katalog.mk` nur Hilfetext. |
| Sensors-Zelle | `harness/README.md` Zeile `make katalog-check`: Zelle *Tut was* 190 Zeichen, Zelle *Bindung* 88 (Grenze 200 aus `.d-check.yml`). |
| Rohbytes (§3.18) | `LC_ALL=C grep -P` auf Backslash, `u` und vier Hex-Ziffern über `tools/katalog/*.go`, `tools/gegenproben/katalog*.go`, Handbuch, `SPEC-201`: kein Treffer. `Schlüssel` in `quellen.go` ist NFC (`c3 bc`), der Gedankenstrich der Überschrift `e2 80 94`. |
| Mengen des Plans | Gelesen gegen §3 *Mengen* (`zaehle` im Test): 30 Positionen, 16 Werte, 14 Grenzfälle, 24 Fälle W gegen N, 38 Fälle SPEC-110, 54, 30, 27 + 2 Läufe, 20 × 3, 12, 7 Paare stimmen mit dem Code. Alle 19 im Plan genannten Testnamen existieren. |

## 2. Mutationen (Programm in Wegwerfkopie; je Mutation `go test -count=1`; Klasse `andere`, kein Fänger wahrscheinlichkeitsbehaftet: jede Mutation 1 von 1 Lauf)

*Rot im Paket, im benannten Test:* `M1b` (Ziffer `9`), `M2` (Unterstrich), `M3` (`Z` am Anfang), `M4b` (`a`), `B1` (Länge 2), `N6`, `N7` (Katalog der Tabelle vertauscht), `N8`, `N9` (Ende des Abschnitts SPEC-110, je eine der zwei Überschriftformen), `N10` (`_test.go` mitgelesen), `N15` (Zeile aus dem Schlüssel der Sortierung), `N16`, `N17` (Zeile oder Abschnitt doppelt), `P1` (Zeile des letzten Vorkommens) und 17 Knotenarten des Lesers übersprungen (`T_AssignStmt`, `BinaryExpr`, `BlockStmt`, `CallExpr`, `CaseClause`, `CompositeLit`, `DeclStmt`, `DeferStmt`, `ExprStmt`, `FuncLit`, `GenDecl`, `IndexExpr`, `KeyValueExpr`, `ReturnStmt`, `SendStmt`, `SwitchStmt`, `ValueSpec`).
*Rot nur durch `panic` eines anderen Tests:* `N13` (meine Mutation brach die Typannahme) und der erste Versuch von `N2` (Schnittfehler der Mutation); beide zählen nicht.
*Grün:* siehe die Findings. Nicht aufgeführt sind die gleichwertigen Mutationen `A4` (Rest hinter dem Span ab dem schließenden Backtick: bei Trennern ohne Backtick dieselbe Menge), `A6`, `K2` und die Knotenarten ohne Literal (`T_IncDecStmt`, `StarExpr`, `TypeAssertExpr`, `SelectorExpr`).

Am echten Baum (Leser mit übersprungener Knotenart gegen die sechs echten Quellen, 16 Läufe): rot sind nur `CallExpr`, `AssignStmt` und `IfStmt`; grün bleiben `CaseClause`, `FuncLit`, `GenDecl`, `CompositeLit`, `ReturnStmt` (die fünf, die der Implementer nannte) und dazu `IndexExpr`, `BinaryExpr`, `ExprStmt`, `DeclStmt`, `KeyValueExpr`, `SwitchStmt`, `ForStmt`, `RangeStmt`. Die Meldung des Implementers stimmt (siehe I-1).

## 3. Findings

### F-1 — MEDIUM — Zeile 1: die 30 Positionen tragen die Achse „Anweisungsart und Klammerung“ nicht, die Menge ist unbeschränkt und hat keine Prüfgrenze

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.17
- `pfad`: `tools/katalog/schl_test.go` · „// positionen sind die 30 Stellen, an denen ein Literal in `konfig*.go` stehen kann.“
- `befund`: Selbst gefahren (Leser mit übersprungener Knotenart, gegen die Suite): `IfStmt`, `ForStmt`, `RangeStmt`, `GoStmt`, `LabeledStmt`, `SelectStmt`, `CommClause`, `TypeSwitchStmt`, `ParenExpr`, `UnaryExpr` (`&T{"N"}`), `SliceExpr`, `Field` (Struct-Tag) und `ImportSpec` bleiben grün: keine der 30 Positionen steht im Rumpf von `if`, `for`, `go` oder `select`, in einer Klammer, hinter einem `&` oder in einem Ausschnitt. Am echten Baum fängt von diesen allein `IfStmt` (neben `CallExpr`, `AssignStmt`). Der Plan nennt zur Menge weder die Achsen noch eine Prüfgrenze als akzeptiertes Negativ.
- `verifizierbar`: ja (Mutation `case *ast.<Knoten>: return false` in `literale`, `go test`)
- `klasse`: „Positionsmenge ohne Achsenliste, Menge unbeschränkt ohne Prüfgrenze“

### F-2 — MEDIUM — Zeile 2 und 12: Grenzen der Zeichenklassen nur an einer Seite

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (beide Seiten jeder Grenze)
- `pfad`: `tools/katalog/schluessel.go` · „(c < '0' || c > '9')“; `tools/katalog/kommandos.go` · „v[i] > 'z'“
- `befund`: `c < '0'` zu `c < '1'` (die Ziffer `0` in einem Namen) und `v[i] > 'z'` zu `v[i] > 'y'` (der Buchstabe `z` in einem Kommando) bleiben in der ganzen Suite grün: die Werte der Zeile 2 tragen nur `1` und `9` als Ziffern, die Kommandos der Zeile 12 haben `a`, aber kein `z`; die Gegenseiten (`9`, `Z`, `a`) sind rot. Ein Schlüssel wie `HTTP10_X` oder ein Kommando wie `analyze` würde dann nicht gelesen und bliebe ohne Pflicht im Handbuch.
- `verifizierbar`: ja (Mutationen `M1c`, `M4a`)
- `klasse`: „Grenze einer Zeichenklasse nur an einer Seite gefangen“

### F-3 — MEDIUM — Zeile 5 und 7: Dateiname und Überschrift „beginnt mit“ und „endet auf“ ohne Köder an der Grenze, Verweisart nur Datei und ins Leere

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15
- `pfad`: `tools/katalog/schluessel.go` · „!strings.HasPrefix(n, \"konfig\") || !strings.HasSuffix(n, \".go\") || strings.HasSuffix(n, \"_test.go\")“; `tools/katalog/schluessel.go` · „strings.HasPrefix(l, ueberschriftSchl)“; `tools/katalog/schluessel.go` · „os.Stat(filepath.Join(dir, n))“
- `befund`: Grün bleiben `"konfig"` zu `"konfi"` (Datei `konfi.go` würde gelesen), `".go"` zu `"go"` (Datei `konfig_ergo` würde gelesen), `"_test.go"` zu `"test.go"` (Datei `konfigtest.go` würde übersprungen), `HasPrefix` zu `Contains` bei der Überschrift `### SPEC-110 — ` (eine Zeile mit dem Text mitten darin würde Abschnitt) und `os.Stat` zu `os.Lstat` (ein Verweis auf ein Verzeichnis `konfig_v.go` würde statt übersprungen als Eingabefehler `is a directory` gemeldet; Probelauf am Original: `katalog: ok`). Die nicht gelesenen Dateien der Zeile 5 (`Konfig.go`, `xkonfig.go`, `konfig.go.bak`, …) liegen nicht an der Grenze des Musters; der Verweis auf ein Verzeichnis fehlt in der Menge der Verweise.
- `verifizierbar`: ja (Mutationen `N11`, `N12`, `MT1`, `K1`, `L1`)
- `klasse`: „Muster mit Anfang und Ende ohne Köder an der Grenze“

### F-4 — MEDIUM — Zeile 12: Wertarten der Konstante schmaler als „kein Zeichenketten-Literal“

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.17
- `pfad`: `tools/katalog/kommandos.go` · „if !ok || l.Kind != token.STRING {“; `tools/katalog/kommandos.go` · „if i >= len(v.Values) {“
- `befund`: Ohne die Prüfung `l.Kind != token.STRING` bleibt die Suite grün, obwohl `strconv.Unquote("'a'")` einen Wert liefert und eine Rune-Konstante (`const sep = 'a'`) dann ein Kommando wäre; ebenso mit `i > len(v.Values)` (Index außerhalb bei einer Konstante ohne eigenen Wert in einer Gruppe, `const ( a = "x"; b )`). Die 16 Formen der Zeile 12 führen weder ein Rune-, Zahl- oder Fließkomma-Literal noch eine Konstante ohne Wert; am Original sind Rune- und Zahl-Konstante grün (Probelauf `katalog: ok`), die Zusage „Wert, der kein Zeichenketten-Literal ist, gibt kein Kommando“ (`SPEC-201` Nr. 5 a) hat für sie keinen Fänger.
- `verifizierbar`: ja (Mutationen `M15`, `N14`)
- `klasse`: „Wertart der Konstante: nur einige Nicht-Literal-Formen im Fänger“

### F-5 — MEDIUM — „erste“ in der Zusage ohne Fänger: erstes `Geprüft werden`, Zeile des ersten Fehlers

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15; `SPEC-201` Nr. 4 b, Nr. 8
- `pfad`: `tools/katalog/schluessel.go` · „strings.Cut(zeilen[0], hinterSchl)“; `tools/katalog/schluessel.go` · „zeile = el[0].Pos.Line“
- `befund`: Nr. 4 b sagt „hinter dem ersten `Geprüft werden`“, Nr. 8 „n ist die Zeile des ersten Fehlers“. Die Mutation auf das letzte `Geprüft werden` (`strings.LastIndex`) und die auf die Zeile des letzten Parser-Fehlers (`el[len(el)-1]`) bleiben grün: keine Fixture führt die Zeile mit zweimal `Geprüft werden`, keine ein Go mit zwei Fehlern (die Fälle der Zeile 6 führen je einen Fehler).
- `verifizierbar`: ja (Mutationen `N2`, `N18`)
- `klasse`: „Ordnungsangabe ‚erste‘ ohne Fall mit zwei Kandidaten“

### F-6 — MEDIUM — `SPEC-201` Nr. 4 c: „ein Präfix aus N nimmt keinen Namen aus H“ ohne Fänger; die Randform ist nicht auflösbar

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.16; `SPEC-201` Nr. 4 c
- `pfad`: `harness/vertraege/SPEC-201-handbuch-katalog.md` · „ein Präfix aus N nimmt keinen Namen aus H, der mit ihm beginnt“; `tools/katalog/schluessel.go` · „!tabelle && !s.n[name]“
- `befund`: Der Satz hat keinen Fänger: `s.n[name]` zu `s.abgelehnt(name)` (Präfix wie Name) bleibt grün. Am Original ergibt ein Name in C, der mit dem Präfix beginnt (`ADMIN_TOKEN`, Präfix `ADMIN_`, in SPEC-110): ohne Zeile Art 7, mit Zeile Art 10; kein Handbuch ist grün (Probelauf). Das Raster des Plans und der Vertrag nennen für diese Zelle keinen Ausgang.
- `verifizierbar`: ja (Mutation `N1`; Probeläufe B1, B2)
- `klasse`: „Zusage ohne Fänger, Randform widersprüchlich“

### F-7 — MEDIUM — `SPEC-102` Zeitbudget: „liegt unter der Streuung“ und „Dreifaches der Kernzahl“ tragen die genannten Zahlen nicht

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.7, §3.11
- `pfad`: `harness/vertraege/SPEC-102-gegenproben.md` · „Die Zusatzzeit liegt unter der Streuung der Läufe (die Läufe davor mit 188 Fällen: 123,4 bis 137,3 s); beide Läufe liegen über dem Budget von 120 s, bei einer Last von mehr als dem Dreifachen der Kernzahl.“
- `befund`: Die Läufe mit 188 Fällen streuen über 13,9 s (123,4 bis 137,3), die mit 204 Fällen liegen bei 169,3 und 171,1 s, also mindestens 32 s über dem alten Maximum; der Schritt `RUN gegenproben` stieg von 80,1 s (Lauf mit 137,3 s) auf 102,7 und 103,7 s, während die 16 Fälle je wenige Millisekunden brauchen. „Unter der Streuung“ ist damit nicht belegt. Die Last von mehr als 60 (3 × 20 Kerne) tragen nur die Werte am Ende der neuen Läufe (63,0 und 68,2); zu Beginn waren es 3,5 und 55,2. Der frühere Absatz sagt „am Ende des Laufs“ für Werte von 54,6 bis 57,2, die unter 60 liegen. Die Ursache des Anstiegs ist offen; die Frage nach dem Budget bleibt beim User und beim Planner (§6 des Plans, Ausgang offen).
- `verifizierbar`: nein (Messung; kein Gate prüft das Budget)
- `klasse`: „Messprosa behauptet mehr, als die genannten Zahlen tragen“

### F-8 — LOW — `//line`-Direktive: die Zeile eines Namens ist im Code entschieden

- `kategorie`: LOW (nach §3.16 wäre eine im Code entschiedene Zelle MEDIUM; hier führt das Raster die Zelle nicht, und die Wirkung entsteht nur bei erzeugtem Quelltext)
- `quelle`: `AGENTS.md` §3.12, §3.16
- `pfad`: `tools/katalog/schluessel.go` · „fset.PositionFor(l.Pos(), false)“
- `befund`: `false` nimmt die Zeile der Datei, nicht die einer `//line`-Direktive (Probelauf: Zeile 4); `true` bleibt in der ganzen Suite grün (`N3`), und `SPEC-201` Nr. 7 sagt „Zeile der Quelle“ ohne diese Form.
- `verifizierbar`: ja (Mutation `N3`)
- `klasse`: „Randform im Code entschieden, ohne Ort im Vertrag“

### F-9 — LOW — Slice-Plan folgt der Korrektur nicht überall (`AGENTS.md` §3.9)

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.9
- `pfad`: Plan · „**Berührte Spec-Stellen:** `SPEC-201` (Vertrag unter `harness/vertraege/`, Abgleich der Kataloge) Nr. 4, 5, 6, 7 (Art 4, 8, 10), 8, 9 e und h, 10 a“; Plan §3 · „`tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` | update“
- `befund`: Der Kopf nennt Nr. 1 nicht, die `233308a1` und `fc5de8c0` ändern (H der Schlüssel). Die Zeile in §3 nennt `tools/harness/gegenproben-eingabe.sh` als `update`; die Datei ist im Diff unberührt (`EINGABEN` führt `tools/katalog` schon).
- `verifizierbar`: ja (Lesen von `git diff --stat 6dbe5d02..HEAD`)
- `klasse`: „Plankopf oder Tabelle nennt eine Stelle, die der Diff nicht berührt, und lässt eine berührte aus“

### F-10 — LOW — Wortlaut: ein Satzzeichen fehlt, ein Satz ist verdreht

- `kategorie`: LOW
- `quelle`: Maintainability
- `pfad`: `harness/README.md` · „Literale `ODR-…` außerhalb von `fehler.go` (Nr. 9 b) ein Schlüssel oder ein Kommando“; `harness/vertraege/SPEC-201-handbuch-katalog.md` · „einen Namen mit dem Präfix eines Präfixes aus N (`ADMIN_TOKEN`)“
- `befund`: In der Aufzählung von *Nicht behauptet* fehlt hinter „(Nr. 9 b)“ das Komma, der Satz liest sich als Fortsetzung des Literals. G19 spricht vom „Präfix eines Präfixes“; `SPEC-102` Fall 192 sagt es richtig („ein Name am Präfix `ADMIN_`“).
- `verifizierbar`: nein
- `klasse`: „Verdrehter oder unvollständiger Satz in einer Vertragszeile“

### I-1 — INFO — Fänger am echten Baum: die Meldung des Implementers stimmt, die Deckung ist schmaler, als Zeile 18 klingt

Von den Knotenarten, die der echte Baum nutzt, fangen nur `CallExpr`, `AssignStmt` und `IfStmt` (§2); Zeile 18 des Plans sagt „Leseregel gilt am echten `konfig*.go`“ und „→ Rot, wo das echte `konfig*.go` die Stelle nutzt“, was für diese drei gilt und für alle übrigen leer. Der Plan sagt es in §3 und im ersten Risiko von §6 (Ausgang offen); die Schutzwirkung vor einer Umformung des Lesers tragen allein die Fixtures (siehe F-1).

### I-2 — INFO — Lauf

Der Scratchpad der Sitzung ist mit anderen Rollen geteilt; ich habe darin Dateien gleicher Namen wie fremde (`base`, `real`, `batch1.sh`, `batch2.sh`, `run.sh`) neu angelegt und am Ende meine Arbeitsverzeichnisse und Skripte gelöscht. Kein Repo-Pfad ist betroffen. Ein Lauf eines anderen Agenten (Mutationslauf im Image `odata-recorder:race`) hielt die Sperre `/tmp/odata-recorder-heavy.lock` rund 40 Minuten; meine Läufe warteten darauf.

## 4. Zu den Schwerpunkten ohne Befund

- **`AGENTS.md` §3.2:** kein `//nolint` im Diff. **§3.6:** kein Gate gelockert; die Stufe `katalog` kopiert zusätzlich `konfig*.go` und `cli.go`, die Stufe `gegenproben` bleibt bei `katalog-bin`.
- **`AGENTS.md` §3.11:** Köpfe der Tests nennen die Mengen, die der Test zählt; die Abweichungen von der Tabelle (Köder 9, Köder-Tabelle, Zeile 16) stehen im Plan unter *Abweichungen*. Die Kommentare im Programm tragen Zusage oder Kopplung; keine verworfene Alternative, kein abwesender Text.
- **`AGENTS.md` §3.12/§3.16, Raster:** alle 30 Zellen haben Ort oder „gilt nicht“ mit Grund; die Zeilen des Rasters stimmen in meinen Probeläufen am Original (Verweis auf Verzeichnis übersprungen, Rune- und Zahl-Konstante kein Kommando, Name im Kommentar kein Schlüssel); die übrigen Zellen habe ich nur gelesen. Die beiden im Code entschiedenen Randformen (a, b) hat der Architect mit `233308a1` in Nr. 1, 4 a, 4 c, 8 nachgetragen; ihre Fänger (`TestKatSchlGrenze` Fall g, `TestKatSchlPaare` Paar 2, `TestKatSchlDateien` Verweis ins Leere) habe ich nicht eigens mutiert; `P1` (Zeile des ersten Vorkommens, dieselbe Datei-Zeile) sah ich rot.
- **`AGENTS.md` §3.13:** die offenen Punkte (Zeit der Gegenproben, Merge-Nähe, erstes Risiko) stehen in §6 mit Ausgang *offen* und Adresse.
- **`AGENTS.md` §3.18:** keine Escape-Folge in den geschriebenen Dateien; `ü`, `§` und der Gedankenstrich stehen als NFC-Text oder Bytewerte (`\xe2\x80\x94`, `Schl\xc3\xbcssel`); die Tests lesen Fixtures über `liesZurueck` zurück.
- **Reihenfolge der Befunde (Zeile 16):** `N15` (Zeile aus dem Schlüssel der Sortierung) rot; die Mutation „keine Sortierung“ des Implementers hängt an der Map-Iteration (60 Läufe) und ist von mir nicht wiederholt.
- **Gegenproben `tools/gegenproben/katalog_schluessel.go`:** 16 Fälle nach `SPEC-201` G16 bis G31 mit den Zeilen aus `kaBaum` (konfig.go ab Zeile 4, Tabellen ab 19 und 28 stimmen mit dem Aufbau); Namen und Reihenfolge prüft `TestKatalogFaelleHabenDieNamenAusDemVertrag`; `SPEC-102` Fälle 189 bis 204 stimmen mit dem Code.
- **Handbuch:** die 26 Zeilen *Einstellungen* und 2 *Kommandos* stimmen für Standardwerte, Grenzen und Meldungen mit `konfig_werte.go` und `konfig_dienste.go` (Stichprobe: `DATA_DIR`, `LISTEN`, `MATCH_VARY_HEADERS`, `VOLATILE_HEADERS`, `MAX_CAPTURE_BODY`, `SQLITE_BUSY_TIMEOUT`, `MODE`, `LINK_REWRITE`); den Inhalt prüft der Verifier.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tools/katalog/schluessel.go`, `kommandos.go`, `quellen.go`, `pruefen.go`, `bericht.go`, `main.go` | geprüft; Findings F-1 bis F-6, F-8 |
| `tools/katalog/*_test.go` (neu und geändert) | geprüft; bestehende Tests nicht abgeschwächt, Fixtures ohne Zahl aus dem Repo |
| `tools/gegenproben/katalog.go`, `katalog_schluessel.go`, Tests | geprüft, ohne Befund |
| `Dockerfile`, `harness/mk/katalog.mk` | geprüft, ohne Befund |
| `harness/README.md` (Sensors, Nicht behauptet) | geprüft; F-10 |
| `harness/vertraege/SPEC-201`, `SPEC-102` | geprüft; F-6, F-7, F-10 |
| `docs/user/benutzerhandbuch.md` | geprüft (Stichprobe gegen den Code), ohne Befund |
| Plan `slice-handbuch-katalog-schluessel-kommandos` (§3, §6, Kopf) | geprüft; F-9 |
| `harness/gegenproben.stempel` | geprüft, ohne Befund (Prüfsumme gleich) |

## Summary

Kein HIGH, sieben MEDIUM, drei LOW, zwei INFO.
Der Produktcode setzt die Nummern 4, 5, 7, 8 des Vertrags um; die Suite fängt die Mutationen des Plans, nicht aber die Klassen dahinter: Positionen nach Anweisungsart (F-1), die andere Seite der Zeichen- und Dateinamen-Grenzen (F-2, F-3), die Wertarten der Konstanten (F-4), „erste“ der Zusage (F-5) und der Satz zum Präfix (F-6).
F-7 betrifft den Wortlaut der Zeitmessung in `SPEC-102`, nicht die Budgetfrage.
Offen für den Verifier: Zeile 18 (Fänger am echten Baum) trägt nur drei Knotenarten, und die Risiken in §6 mit Ausgang *offen* sind vor `done/` zu schließen.
