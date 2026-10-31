# Verifikations-Report 5: slice-handbuch-geruest — 2026-10-31

**Rolle:** Verifier, Nachverifikation (fünfte Runde, gezielt auf das Delta der Runde 7) in frischem Kontext nach den Fixes zu [Y-1 bis Y-9](2026-10-30-slice-handbuch-geruest-verifikation-4.md). Frage: „Bauen wir es richtig?“ gegen Plan, DoD, `SPEC-200` (Nummer für Nummer) und `AGENTS.md` §3.9 bis §3.18. Kein Review, keine Reparatur: Code, Plan und Vertrag sind von mir unverändert; alle Mutationen liefen in Kopien im Scratchpad.

**Bezug:** [LH-RB-01](../../spec/lastenheft.md), [LH-QA-05](../../spec/lastenheft.md), [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Ort der Verträge), [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Größenmarke).

**Gegenstand:** Zweig `slice-handbuch-geruest`, HEAD `ad7275b` (Merge von `main`), Stand der Arbeit `1c9f1ee`. Gelesen: `harness/vertraege/SPEC-200-handbuch.md` (154 Zeilen, ganz), `tools/pruefung/handbuch.go`, `handbuch_kopf.go`, `handbuch_mengen.go`, `handbuch_beispiel.go`, `handbuch_runde7_test.go` (ganz), der Plan (DoD, §3 Zeilen Z21 und Z90 bis Z93, Mutationen der Runde 7, §4 Messung), der Report der vierten Runde und die Diffs `ebb62c0..1c9f1ee` (`SPEC-200`, ein Zweig in `liesHandbuch`, Tests, Plan). **Der Merge `ad7275b` berührt `tools/pruefung` und `harness/vertraege` nicht** (`git diff 1c9f1ee..ad7275b --stat` über beide Pfade: leer); er bringt nur `main`-Inhalt, in den Pfaden der Größenmarke eine Zeile (V-7).

**Modell:** claude-sonnet-5-5 · **Last:** Load Average beim Start 20,9 / 21,6 / 19,8 (20 Kerne), beim Probe-Paket 4,9 bis 12; schwere Läufe unter `/tmp/odata-recorder-heavy.lock`, `--memory=4g --memory-swap=4g --cpus=4`, netzlos.

**Methode.** Scratchpad `verif-geruest-5/` (flüchtig). (1) Sensoren selbst gefahren (Abschnitt 6). (2) Ein **Probe-Paket** `probe5_test.go` mit sieben Testfunktionen in einer Kopie von `tools/pruefung`, jedes Orakel von Hand aus dem Wortlaut des Vertrags (nie der Parser): P1 Ausgangswert, P2 Pfade, P3 Bindungszeile, P4 Leerraum, P5 Ordnung, P6 Version und Stand, P7 Spans. Auf dem unmutierten Code sind alle sieben grün. (3) **45 eigene Mutationen** (alle übersetzbar, alle an Stellen, die weder Implementer, Architect noch die vierte Verifikation nannten), je gegen die bestehende Suite (`go test -count=1 -failfast -skip Probe`) und gegen das Probe-Paket; dazu **20 Mutationen aus der Tabelle der Runde 7** zur Nachmessung der Zahlen im Plan (Abschnitt 5).

---

## 1. Verdikt

**NICHT BESTANDEN.** Es gibt kein HIGH. **Der Code weicht an keiner Stelle vom Wortlaut des Vertrags ab**: das Probe-Paket (137 694 Ausgangswert-Formen, 1 112 063 Zeichen, 3 000 zufällige Ordnungen, 400 Span-Absätze und weitere) findet keine Abweichung, und die 13 geänderten Nummern sind mit den Nachbarnummern widerspruchsfrei (Abschnitt 3). Die Fixes zu Y-1 bis Y-5 sind vollständig und gegen die Klasse aufgelöst (Abschnitt 3). Alle Sensoren sind grün, Z83 ist stabil (20 von 20).

Das Urteil beruht auf **drei Zusagen ohne Fänger** (`AGENTS.md` §3.15: MEDIUM), die dieselbe Klasse wie Y-1 und Y-2 tragen (die Menge der Zusage ist größer als die Elemente, die der Bestand führt), und die erst die Mutation an einer Stelle zeigt, die niemand nannte: V-1 (Binary „keine reguläre Datei“), V-2 (Version „ohne Leerraum“ jeder Art), V-3 (Span nach einem Span mit zwei oder mehr Backticks). Jede ist eine Ergänzung einer bestehenden Schleife; der Code ist in allen drei Fällen richtig.

## 2. Befund-Tabelle

| ID | Schwere | Befund | Beleg | Zuständig |
|---|---|---|---|---|
| V-1 | MEDIUM | **Binary „keine reguläre Datei“ hat nur für das Verzeichnis einen Fänger** (`SPEC-200` Nr. 2 c: „oder keine reguläre Datei“; Nr. 2 f). Die Mutation `err != nil \|\| !fi.Mode().IsRegular()` → `err != nil \|\| fi.IsDir()` (`handbuch.go` Zeile 51) bleibt im Bestand **grün** (`TestHandbuchBinary` führt `verz`, Z90 E7 ein Verzeichnis). Mit ihr startet ein Gerät oder eine FIFO als Binary und endet in `Binary nicht startbar: … permission denied` statt `Binary fehlt`. Mein Probe-Paket (Menge: Eintragsarten außer der regulären Datei, 5 Arten; geführt 3 davon: Zeichengerät `/dev/null`, FIFO, Verweis auf das Gerät; dazu Verzeichnis, Komponente davor ist eine Datei, kein Eintrag, 6 Elemente) macht sie rot in 3 von 6. | `g1/bin-nondir`: Bestand `rc=0`, Probe 3 Meldungen | Implementer (Fänger in Z90 als Schleife über die Arten; nicht geführt: Blockgerät und Socket) |
| V-2 | MEDIUM | **`odata-recorder <V>`: „V ohne Leerraum“ hat nur für Leerzeichen und Tabulator einen Fänger** (Nr. 4 c mit Nr. 3 b: Leerraum ist jedes Zeichen, das `unicode.IsSpace` trifft). Die Mutation `strings.ContainsFunc(v, unicode.IsSpace)` → nur ASCII (`handbuch_kopf.go` Zeile 175) bleibt im Bestand **grün**; V mit U+00A0, U+0085, U+2028 usw. gilt dann als Version. Probe: 24 Leerraum-Zeichen × 3 Stellungen {innen, hinten, vorn} = 72 Elemente, die Mutation ist in 57 rot (die 15 grünen sind die 5 ASCII-Zeichen × 3 Stellungen, dort wirkungsgleich). | `g3/ver-space`: Bestand `rc=0`, Probe 57 Meldungen | Implementer (Schleife über die 24 Zeichen bei `TestHandbuchVersionForm`) |
| V-3 | MEDIUM | **Spans nach einem Span mit n ≥ 2 Backticks haben keinen Fänger** (Nr. 9 b: „die nächste Folge von genau *n* Backticks schließt“; die schließende Folge wird ganz verbraucht). Die Mutation `i = k + n` → `i = k + 1` in `codeSpans` (`handbuch_mengen.go` Zeile 185) bleibt im Bestand **grün**; mit ihr ist `` ``xyz``, `make t1` `` blind für `make t1`, weil der Rest der schließenden Folge einen neuen Span öffnet. Probe: 400 Absätze mit 1 bis 5 Spans, n = 1 bis 4 je Span, Orakel durch Konstruktion; die Mutation ist rot. Der Bestand führt Spans mit zwei Backticks einzeln (`` ``make nix`` ``, `` `make ``build`` nix` ``), nie vor einem weiteren Span. | `g4/span-weiter`: Bestand `rc=0`, Probe rot | Implementer (Fänger in der Span-Schleife von Z85/Z84: Menge n × Folgeposition) |
| V-4 | LOW | **Stand: das Jahr 0000 und die Jahre vor 1582 sind weder entschieden noch gefangen** (Nr. 4 d: „vier Ziffern … ein Tag des Gregorianischen Kalenders“). Der Code nimmt `0000-01-01` und `0000-02-29` an (proleptisch); die Mutation `j >= 0` → `j > 0` (`handbuch_kopf.go` Zeile 119) bleibt im Bestand grün, die Generatoren von Z11 laufen nur über 2026 und wenige Schaltjahre. Probe: 5 Jahresgrenzen (`0000`, `0001`, `1581`, `9999` und der 29. 2. des Jahres 0000), die Mutation ist in 2 rot. | `g2/stand-j0`: Bestand `rc=0`, Probe 2 Meldungen | Architect (ein Satz: Jahr 0000 bis 9999 gelten, proleptisch, oder nicht), dann Implementer (Fänger) |
| V-5 | LOW | **Ein Symbolischer Verweis auf eine lesbare Datei in harness/mk/ zählt nicht als Datei, an allen anderen Stellen schon** (Nr. 2 f: „vorhanden, aber keine lesbare Datei“ gilt als nicht vorhanden; Nr. 9 a: „Dateien mit der Endung `.mk`“). `makeDateien` prüft `e.Type().IsRegular()`; ein Verweis `harness/mk/x.mk` fällt heraus, während `Makefile`, `d-check.mk`, `repo.mk`, `Dockerfile` (gelesen mit `os.ReadFile`, folgt dem Verweis) und ein Binary (`os.Stat`) ihn auflösen. Nr. 5 f entscheidet Verweise nur für docs/user/beispiele. Die Mutation „Verweis zählt“ (`!e.IsDir()`) und `os.Stat` → `os.Lstat` am Binary bleiben im Bestand grün. Das Repo hat keinen Verweis in harness/mk, das `Dockerfile` erzeugt keinen; die Folge im Betrieb ist ein Ziel, das `make help` kennt und das Handbuch nicht nennen darf. Der Code folgt einer Lesart; entschieden ist sie nicht (§3.16: im Code entschiedene Zelle). | `g2/mk-symlink`, `g2/bin-lstat`: Bestand `rc=0` | Architect (Satz in Nr. 2 f: Verweis auf eine Datei gilt als Datei, oder nicht), dann Implementer |
| V-6 | INFO | **Merge und Zahl der Messung:** das Gate meldet am HEAD `ad7275b` `ok, Messung 6500 von Marke 6900`; der Plan nennt 6499 (Stand `1c9f1ee`). Die eine Zeile mehr ist die Aufnahme von `main` in `harness/mk/ai-harness-init.md` (`git diff --numstat 975228d..ad7275b` über die sechs Pfade des Blocks = 6500; `..1c9f1ee` = 6499). Marke fest, 94,2 %; kein Befund, nur die Zahl im Plan gilt für den Stand davor (`SPEC-128` Nr. 4 g, 10 g). | `gm.out`; `git diff --numstat` | Planner (beim Nachziehen der Messung) |
| V-7 | INFO | **Plan Z91, Spalte der Mutationen, nennt „eine Datei nennt nur der erste Zaun (L6)“.** Dieselbe Mutation steht in den Abweichungen unter der Tabelle der Runde 7 als äquivalent (bei einem Zaun je Fixture grün); als Fänger von L6 gilt „ein gültiger Zaun nennt nach einer alleinstehenden Zeile nicht“ (Tabelle der Mutationen, Zeile Z91). §3.9: die Spalte ist nachzuziehen. Die Nachmessung der Zeile bestätigt das: `Bindungszeile ohne Zaun` entfällt ⇒ 4 Meldungen. | Plan Zeilen 237, 348, 363 | Planner / Implementer |
| V-8 | INFO | **`AGENTS.md` §3.14 nennt die Faktoren 40 und 65, `SPEC-128` Nr. 1 rechnet 80 und lässt den Zuschlag für Leser fremder Bytes entfallen;** der Plan rechnet nach `SPEC-128` (9655 und 5085 stimmen mit 80). Die höher rangierende Quelle gilt; `AGENTS.md` ist anzupassen (nur nach Diff-Vorlage an den User). | `SPEC-128` Nr. 1, `AGENTS.md` §3.14 | Planner (Diff-Vorlage) |

Einordnung. V-1 bis V-3 sind Klasse-Lücken im Sinn von `AGENTS.md` §3.17 (Y-1 und Y-2 der Vorrunde in der Form: die Menge der Zusage, jede nicht reguläre Eintragsart, jede Art Leerraum, jede Position nach einem Span, ist größer als das Element, das der Bestand führt). Sie liegen **nicht im Delta** der Runde 7, sondern in Nachbarnummern (2 c, 4 c, 9 b), die das Probe-Paket aus dem Wortlaut gebildet hat; die Fixes der Runde 7 selbst sind lückenlos (Abschnitt 3).

## 3. Das Delta der Runde 7, gegen die Klasse aufgelöst

| Befund | Menge (Wortlaut des Vertrags) | Elemente | Mein Beleg |
|---|---|---|---|
| Y-1 Ausgangswert (Nr. 6 c, 13 Teilaussagen) | jede Folge von Zeichen hinter `--- Ausgangswert`: Zahl der Zeilen, Dezimalzahl, Wertebereich, führende Null, Rest, Schluss-LF | P1: alle Folgen bis Länge 4 über 19 Zeichen (ASCII-Ziffern, Buchstaben, Vorzeichen, Punkt, Leerraum aller Art in Auswahl, `\n`, `\r`, U+00A0, U+FF11, U+0663) = 137 561, dazu 120 Ziffernfolgen der Längen 1 bis 40 und 13 Überläufe (2^32, 2^63, j·2^64 + r) = **137 694** | Orakel von Hand (Länge 1 bis 3, nur ASCII-Ziffern, keine führende Null außer `0`, Wert ≤ 255, Schluss-LF): **0 Abweichungen**, 100 gültig. Z21 führt 50 (7 + 43, Zählprüfung im Test, Rest 6 + 9 + 9 + 5 + 13 + 1 = 43 nachgezählt). Mutationen an Stellen, die der Plan nicht nannte: `len(z) > 3` entfällt, `len(z) > 20` (die 20-stelligen Überläufe), `n > 259`, `len(zeilen) < 1`, `pos[0] < 0`, `\r` nur am Ende, jede **rot**; die Prüfung `len(z) > 3` ist für 4 bis 19 Stellen äquivalent und für 20 Stellen tragend (Überlauf auf 0 bis 5: `18446744073709551616`, `…621`), und Z21 führt diese zwei. |
| Y-2 Gleichstand (Nr. 2 d, > 12 Befunde) | jede Gruppe gleicher {Nummer, Zeile, Name} in einer Ausgabe, die nach der Erzeugung umgeordnet wird | P5: 3 000 zufällige Listen von 2 bis 151 Befunden (1 500 davon mit 13 bis 132), Schlüssel Nummer 3 bis 5, Zeile 0 bis 2, Name 4 Werte; dazu 12 Ende-zu-Ende-Dokumente mit bis zu 7 Absätzen, je 44 Muster in zufälliger Reihenfolge | Orakel: Einfügesortierung mit Index als letztem Schlüssel; **0 Abweichungen**. `sort.Slice` statt `SliceStable` rot in Z93 (**31 von 31**, 27 U + 3 E + K; Meldungen aus `handbuch_runde7_test.go` Zeile 360, 415 und 436) und in P5 (beide Funktionen). `<=` im letzten Schlüssel 31, Schlüssel `name` / `nr` entfällt 9 / 12 Meldungen, `zeile` entfällt 14 und Zeile absteigend 17 Meldungen (verteilt auf die drei Tests Z90, Z91 und Z93 des Deltas; der Plan zählt 9 und 12 allein für Z93), Text als letzter Schlüssel 4, Zeile absteigend 17: wie im Plan. |
| Y-3 ENOTDIR (Nr. 2 f, 3 a, 2 c, 2 a, 5 d, 5 f) | jede Stelle unter *Wurzel*, an der ein Eintrag fehlt, eine Datei statt eines Verzeichnisses oder ein Verzeichnis statt einer Datei steht | P2: 12 Stellen {Handbuch, docs, docs/user, beispiele, aufruf.txt, `Makefile`, `d-check.mk`, `repo.mk`, `Dockerfile`, harness, harness/mk, harness/mk/m.mk} × 2 Zustände (fehlt, Art vertauscht) = **24**, dazu Binary 6 = **30** | Orakel von Hand (Exit, Befundliste in Reihenfolge, Zahl der Starts): **0 Abweichungen**. Mutationen: ENOTDIR-Zweig entfällt 4 (E1, E2, E3, E8), jeder Lesefehler `fehlt` 1 (E4), Binary nur `ErrNotExist` 2: wie im Plan. Lücke nur bei der Art des Binary (V-1). |
| Y-4 Bindungszeile allein (Nr. 5 c, 5 f, 7 a) | Lage der Zeile {Dateiende, vor Prosa, Leerzeile, NBSP-Zeile, doppelt, vor offenem Zaun, vor Tilde-, Vier-Backtick-, eingerücktem Zaun, vor Zaun mit Info, mit Backtick in der Info, ungültige Form, im ungebundenen Zaun, nach der Schlusszeile, zweiter Zaun derselben Datei} | P3: **16** Fälle (Z91: 6) | Orakel von Hand (Befundliste, Starts): **0 Abweichungen**. Mutationen: `Bindungszeile ohne Zaun` entfällt 4; nur geschlossene Zäune nehmen ein (`bz-nur-geschlossen`), `Contains` statt `HasPrefix` (zwei Stellen), Block-Vergleich getrimmt: alle rot. |
| Y-5 „leer“ (Nr. 3 b, 13 a; 4 a, 4 b, 5 a) | jedes Zeichen von U+0000 bis U+10FFFF ohne `\n` und Surrogate | P4: **1 112 063** Zeichen, davon 24 Leerraum, an fünf Stellen {Absatztrenner, Zeile des Kopfes, Wert des Schlüssels, Info-Zeichenkette, Zeichen hinter der Schlusszeile} | Orakel: Liste der 24 Zeichen von Hand (U+0009 bis U+000D, U+0020, U+0085, U+00A0, U+1680, U+2000 bis U+200A, U+2028, U+2029, U+202F, U+205F, U+3000): **0 Abweichungen**. Mutationen: leer ist nur `""`, nur Teilmenge (Leerzeichen, Tabulator, U+00A0), Kopfzeile roh, Wert nur Leerzeichen, Info nur Leerzeichen, Schlusszeile nur Leerzeichen: alle rot. Nachbarnummer 4 c fällt heraus (V-2). |
| 14 g (neu: Waise, Arbeitsverzeichnis) | Satz der Folge | – | gelesen; Verhalten bei Signal ist ein akzeptiertes Negativ ohne Fänger (Y-7 der Vorrunde); nichts zu prüfen. |

**Konsistenz der Nachbarnummern** nach den 13 geänderten Nummern (2 a, 2 c, 2 d, 2 f neu, 3 a, 3 b, 5 b, 5 d, 5 f, 6 c, 7 a, 13 a, 14 g), Stelle für Stelle gelesen: 3 b zählt 4 a, 4 b, 5 a und 13 a als Verwender von „leer“ (4 e und 14 a sind keine); Leerraum = `unicode.IsSpace` gilt in 4 a, 4 b, 4 c, 5 a (Schlusszeile rechts getrimmt), 9 b (Wörter), 13 a und stimmt mit 3 c überein (Zs, Zl, Zp sind zulässig, die Cc-Zeichen U+000B, U+000C, U+000D, U+0085 sind Leerraum **und** Befund; Z92 führt das als Spalte `cc`). 2 a / 2 c / 2 f / 3 a nennen dieselben Fälle (Wurzel, docs, docs/user, Handbuch als Datei oder Verzeichnis, Verweis ins Leere) mit demselben Ergebnis `Handbuch fehlt` bzw. `nicht lesbar`; 2 c lässt für das Binary „kein Eintrag, Komponente davor ist eine Datei, keine reguläre Datei“, 2 f verweist zurück. 5 d / 5 f / 7 a stimmen in „genannt“ überein (gültige Bindungszeile unmittelbar vor einem geschlossenen Zaun, Block nicht gebunden darf nennen, Verzeichnis statt Datei ist `fehlt`, aber nennt); 5 c schließt die Zeilen in geschlossenen Zäunen und vor offenen Zäunen richtig aus. 14 d bis 14 g stimmen mit 2 e, 7 c bis 7 e und 9 b überein. **Ein Widerspruch besteht nicht;** offen sind nur V-4 und V-5 (Randformen ohne Satz).

## 4. SPEC-200 Nummer für Nummer

| Nr. | Ergebnis |
|---|---|
| 1 | Handbuch gegen die Regel gelesen: Kopf, Beispiele nur als gebundene Zäune, `make` nur in Inline-Code; Stufe `handbuch` ohne Cache: `ok` |
| 2 a bis f | P2 (30), Z90 (12), P3: Eingabefehler in Reihenfolge, kein Start vor dem Eingabefehler (Starts 0), ENOTDIR wie „kein Eintrag“; V-1, V-5 |
| 3 a bis c | P4 (Leerraum, Zeichen), Bestand Z92 (30); keine Abweichung |
| 4 a bis e | P4 und P6: Kopfzeilen aus Leerraum, Wert leer, V mit 72 Leerraum-Formen (V-2), Stand mit 5 Jahresgrenzen (V-4); sonst gleich Vorrunde |
| 5 a bis f | P3 (16), P2 (Binary, beispiele), Bestand Z91; keine Abweichung |
| 6 a bis d | P1 (137 694); keine Abweichung |
| 7 a bis e | Z83 gefahren (Abschnitt 6); Starts in P2 und P3 gezählt; keine Abweichung |
| 8 | P2: Image nur aus `Makefile`, Pfad nur aus `Dockerfile`, Verzeichnis statt Datei deckt nichts; keine Abweichung |
| 9 a, b | P2 (Zielliste aus vier Quellen), P7 (400 Absätze); V-3, V-5 |
| 10 bis 12 | Mengen unverändert seit der Vorrunde (44 Muster, 104 Formen); P5b nutzt alle 44 Muster |
| 13 a bis d | P4 (Absatz), P5b (Reihenfolge), Bestand Z92; keine Abweichung |
| 14 | a bis h gelesen; g (neu) und e (Zeit) siehe oben und Abschnitt 6 |
| 15 a bis d | `Dockerfile` Stufe `handbuch`, `.dockerignore`, `harness/mk/handbuch.mk`, Sensors-Zeile unverändert seit der Vorrunde; Verdrahtung läuft im Gate (`pruefung handbuch: ok`) |
| Gegenprobe | nicht gefahren (Auftrag) |

## 5. Mutationen, Äquivalente, Nachmessung der Zahlen im Plan

**Eigene Mutationen (45, alle übersetzbar).** Rot im Bestand: 38. Grün im Bestand: 7. Davon **Lücken** (nicht äquivalent, Probe rot oder Verhalten verschieden): `bin-nondir` (V-1), `ver-space` (V-2), `span-weiter` (V-3), `stand-j0` (V-4), `bin-lstat` und `mk-symlink` (V-5). **Äquivalent** (belegt): `span-nurspace` (das Entfernen der Bedingung `Trim(s, " ") != ""` in `spanText` lässt einen Span aus Leerzeichen gekürzt; ein Text nur aus Leerzeichen beginnt nie mit `make `, und nur dieser Beginn wird gelesen; Probe P7 ebenfalls grün). Die 38 roten sind aus dem benannten Grund rot (die Meldung nennt den Test der Klasse: z. B. `TestHandbuchTrefferRand` für den Neustart der Suche, `TestHandbuchNormalisierung` für die einmalige Faltung der Leerzeichen, `TestHandbuchZiellisteZeilenformen` für den Anker der Zielzeile, `TestHandbuchReihenfolgeNamenBytefolge` für die Ordnung der Namen). Zwei Mutationen übersetzten beim ersten Versuch nicht (`eintrag-leer`, `ver-space`: ungenutzte Variable oder Import) und liefen in veränderter Schreibung neu; eine weitere (`zaun-backtick`) ebenso.

**Nachmessung der Tabelle der Runde 7** (20 Mutationen, jede nur gegen die fünf Tests des Deltas; Meldungen = Zeilen `_test.go:n:` der Ausgabe):

| Zeile | Mutation | Plan | gemessen |
|---|---|---|---|
| Z21 | `z == ""` entfällt / `len(zeilen) != 1` → `== 0` / `n < 0` entfällt / führende Null entfällt | 1 / 2 / 20 / 4 | 1 / 2 / 20 / 4 |
| Z21 | `n > 256` / `n >= 255` / `TrimSpace` vor der Prüfung / Schluss-LF-Prüfung entfällt | 1 / 1 / 6 / 1 | 1 / 1 / 6 / 1 |
| Z21 | Variante von „`007` gilt“ (Prüfung nur bei Länge 2) | 1 | 2 (andere Mutation, rot) |
| Z90 | ENOTDIR-Zweig entfällt / jeder Lesefehler `fehlt` / Binary nur `ErrNotExist` | 4 / 1 / 2 | 4 / 1 / 2 |
| Z91 | `Bindungszeile ohne Zaun` entfällt | 4 | 4 |
| Z92 | leer ist nur `""` | 25 | 25 |
| Z93 | `<=` im letzten Schlüssel / `name` / `nr` entfällt / Text als letzter Schlüssel | 31 / 9 / 12 / 4 | 31 / 9 / 12 / 4 |
| Z93 | `zeile` entfällt / Zeile absteigend | 9 / 12 | 14 / 17 in fünf Tests (9 / 12 in `TestHandbuchOrdnungStabil` laut Plan, der Rest sind Folgemeldungen aus Z90 und Z91) |

Nicht nachgemessen: die übrigen Zeilen der Runde 7 (Z90 „ENOTDIR nur bei Wurzel“, „Reihenfolge der Eingabefehler“, „Lesefehler von beispiele / harness/mk / Makefile“, E10; Z91 die vier Mutationen der Lage; Z92 „nur ASCII“, „nur Zs“, „jede Zeile aus einem Zeichen“, U+2029; Z93 M2). Die Zahlen des Plans (Elemente 50 / 12 / 6 / 30 / 31 und 103 Zeilen) stehen als Zählprüfung im Test (`Fatalf` bei Abweichung; Z93: `n != 27 \|\| anders != 27`, 3 E, 1 K) und im Gate (`zeilen=103`).

## 6. Läufe

- **`make gates`** im Worktree unter der Sperre: `EXIT=0` (Datei `gates.rc`, nicht über `| tail` gelesen); darin `a-check`, `d-check`, `baseline-verify`, `groessenmarke ok`, `migration-check: ok`, `schema-check`. Die Stufen liefen **nicht** aus dem Cache: `pruefung test` 109,2 s ok, `pruefung race` 136,9 s ok, `pruefung build` ok, Lint (`lint.sh`, 0 issues je Lauf), `pruefung handbuch: ok` (0,142 s).
- **Stufe `pruefung`** (`gofmt`, `go vet`, `go test ./...` in `tools/pruefung`) stand im Lauf von `make gates` `CACHED`; deshalb einzeln: `docker build --target pruefung --no-cache-filter pruefung`, Exit 0, 36,9 s. Go 1.27.1 (Image `golang@sha256:8a5910f3…`, Stufe `deps`).
- **`make groessenmarke`:** `ok, Messung 6500 von Marke 6900`; Messpunkt `zeilen=103 nebenlaeufig=3 fakes=2 produktcode=1000 schaetzung=9655 marke=6900 messung=6500`; Hinweis `Tabelle ergibt 9655, Schätzung 5085` (kein Befund). **Nachgerechnet:** Tabelle 103 Zeilen (3 nebenläufig: Z26, Z49, Z83; 100 andere); 85·3 + 80·100 + 200·2 + 1000 = 255 + 8000 + 400 + 1000 = **9655**; `git diff --numstat 975228d..1c9f1ee` über `tools/pruefung/ tools/gegenproben/ Dockerfile .dockerignore harness/mk/ harness/README.md` = **6499** (30 Dateien), `..ad7275b` = **6500** (31 Dateien, die eine Zeile mehr ist `harness/mk/ai-harness-init.md`, V-6).
- **Z83 als Wirkungs-Fänger** (`TestHandbuchZeitgrenzeUndWarten`, 16 Zellen), `-test.count=1` je Prozess, ein gebautes Testbinary, **20 Prozesse nacheinander: 20 von 20 grün**, ohne Mutation, Load Average (Host) 6,13 / 7,77 / 12,45 vor und 4,91 / 6,64 / 11,31 nach dem Lauf; Container auf 4 CPUs begrenzt. Keine Flakiness.
- **Z93, Kopplung (Auftrag 3):** Der Kommentar am Test (`handbuch_runde7_test.go` Zeilen 304 bis 310) trägt die Klasse **Kopplung** („Kopplung: sort.Slice ordnet bis 12 Elemente durch Einfügen stabil und erst ab 13 anders; die Schwelle und die Eingaben … gehören zur Go-Version des Images. Ändert Go das Verfahren, meldet die Zählprüfung `anders` …“) und sagt nur zu, was der Test prüft (`n != 27 || anders != 27`, Meldung nennt „Kopplung an die Go-Version“; §3.7, §3.11). **M1 ist beim Image rot:** `sort.Slice` statt `SliceStable` im Image `golang@sha256:8a5910f3…` (Go 1.27.1): 31 von 31 Meldungen (27 U, 3 E, K), Grund die Reihenfolge der Gruppen, nicht die Zählprüfung; dieselbe Mutation ist in P5 rot.
- **Rohbytes (§3.18):** `LC_ALL=C grep -nP '[\x80-\xFF]'` auf `handbuch_runde7_test.go`: Treffer nur die deutschen Umlaute in Kommentaren und Meldungen (Zeilen 14, 19, 35, 38, 67, 73, 74, 187, 190, 194, 215, 234, 268, 305 bis 309, 375, 377, 427); `\\u[0-9A-Fa-f]{4}` 0 Treffer; BOM, U+00A0 (`\xC2\xA0`) 0 Treffer — die Werte jenseits von ASCII entstehen aus `rune1(…)` und der Konstanten `nbsp` (Bytewerte). Dieser Report ist nach dem Schreiben auf Rohbytes gegengelesen (Ergebnis in der Übergabe).
- **Nicht gefahren, wie beauftragt:** `make gegenproben` und der Stempel.

## 7. DoD: Stand

| DoD-Zeile | Stand |
|---|---|
| Handbuch mit Kopf und Kapiteln, Beispiele als Dateien, Lauf in Stufe `handbuch` | erfüllt |
| Anpassung des Standards; Diff-Vorschlag für `AGENTS.md` | wie in Runde 4 (Vorschlag liegt außerhalb des Repos, Ausgang *offen* im Plan §6, Adresse User) |
| Gate `make handbuch-check`; je Zeile der Tabelle eine Mutation rot, je Muster eine | Verdrahtung erfüllt; Gegenproben vorhanden, nicht gefahren; die Zeilen Z21, Z90 bis Z93 sind rot gesehen (Abschnitt 5); **V-1 bis V-3** zeigen drei Zusagen in Nachbarnummern ohne Fänger |
| `make gates` grün; `make docs-check` ohne Befund | erfüllt (Exit 0; `make docs-check` mit diesem Report siehe Übergabe) |
| Review, Report unter `docs/reviews/` | erfüllt (zwei Reviews, vier Verifikationen davor, dieser Report) |
| Doku-Update (`SPEC-200` Stand, `SPEC-102`, `harness/README.md`) | erfüllt |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen, Sache des Planner nach dem Fix dieser Runde (nicht Auftrag dieser Runde) |

## 8. Ungeprüft

- `make gegenproben`, der Stempel; die Fälle H1 bis H10 sind nicht gefahren.
- Fehlende Berechtigung als Nicht-root (der Container läuft als root; Z90 nennt das Negativ), SIGINT, volles TMPDIR.
- Die Mutationen der Runde 7 außer den 20 in Abschnitt 5; die übrigen Zeilen der Tabelle (Z1 bis Z89) in dieser Runde nicht erneut aufgelöst; die vollen Differentiale der Vorrunden (Mengen A, B, C über Formen × Stellungen) nicht wiederholt, weil Code und Vertrag dort unverändert sind.
- Blockgerät und Socket als Binary (V-1: Menge von 5 Arten, geprüft 3).
- Verständlichkeit und Vollständigkeit des Handbuchs (K6, Nr. 14 a, b); Planner-Punkte (Stand `2026-10-30` im Handbuch, Closure) sind nicht Auftrag.

## 9. Register-Kandidaten (nur genannt, nicht geschrieben; Abgleich und Zähler sind Sache des Planner)

- `BEO-REPO/menge-der-zusage-nur-an-einem-element-gefuehrt`: V-1 (Verzeichnis statt jeder nicht regulären Art), V-2 (Leerzeichen und Tabulator statt der 24 Zeichen), V-3 (ein Span statt jeder Position nach einem Span). Dieselbe Klasse wie Y-1 und Y-2 der Vorrunde, jetzt an Nachbarnummern des Deltas.
- `BEO-REPO/randform-erst-im-review-entschieden`: V-4, V-5 (Jahr 0000, Verweis in harness/mk): wieder Zellen, die der Code entschied.

## 10. Zusammenfassung

| Schwere | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 (V-1, V-2, V-3) |
| LOW | 2 (V-4, V-5) |
| INFO | 3 (V-6 bis V-8) |

**Übergabe an den Planner:** NICHT BESTANDEN, ohne Abweichung des Codes vom Vertrag. V-1 bis V-3 an den Implementer (je eine Ergänzung einer bestehenden Schleife; der Bericht nennt je Befund die Menge, die Zahl ihrer Elemente und eine Mutation, rot gesehen, an einem Element, das dieser Report nicht nannte; die Probe-Paket-Funktionen P2, P6 und P7 sind als Orakel verwendbar). V-4 und V-5 an den Architect (je ein Satz in `SPEC-200`: Jahr 0000 bis 9999; Verweis auf eine Datei in harness/mk), danach Implementer. V-6 und V-7 an den Planner beim Nachziehen der Messung und der Spalte von Z91, V-8 als Diff-Vorlage `AGENTS.md` §3.14 an den User. Nach diesen Änderungen genügt eine Teilprüfung der vier Fänger (V-1 bis V-4) und der Sätze (V-4, V-5); Code und Vertrag bleiben im Übrigen unverändert.
