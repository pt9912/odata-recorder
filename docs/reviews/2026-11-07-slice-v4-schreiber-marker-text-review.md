# Review: slice-v4-schreiber-marker-text — 2026-11-07

**Art:** Review (einmal) — gegen Plan, `SPEC-169`/`SPEC-168`/`SPEC-126`, `LH-FA-07`, `LH-QA-05` und die Hard Rules; nicht gegen die DoD (Verifier).

**Gegenstand:** Zweig `slice-v4-schreiber-marker-text`, HEAD `c2f296c`, Basis `94dafe9`. Plan: `docs/plan/planning/in-progress/slice-v4-schreiber-marker-text.md` (§3 Z01 bis Z13, Mengen M1 bis M8, Zweig-Menge W01 bis W18, Absatz *Bericht des Implementers*, §6).

**Rolle/Modell:** reviewer · claude-sonnet-5-5.

## Urteil

Kein HIGH. Der Produktcode (`ausgabe.stringWert`, eine Methode und eine geänderte Zeile) setzt `SPEC-169` Nr. 4 um; der Leser ist unverändert. Von 121 eigenen Mutationen blieb keine grün, die das Verhalten ändert; eine Mutation ist nur im Paket rot (kein Marker-Bezug), zwei brachen die Übersetzung und sind unten einzeln wiederholt. Gefunden sind ein Fänger, der nur das Paar zweier Wächter fängt (Z06, LOW), eine überzogene Zusage der Zweig-Menge (Z10/Z13, LOW) und zwei INFO.

## 1. Eigene Belege

| Prüfung | Beleg |
|---|---|
| Leser unverändert | `git diff 94dafe9..HEAD --stat -- src`: nur `schreiber_ausgabe.go` (+14), `schreiber_bau.go` (1/1) und die zwei neuen Testdateien; `scan.go`, `leser*.go`, `wert.go` nicht berührt. |
| Größenmarke | numstat über die vier Pfade des Blocks: 14 + 2 + 424 + 331 = 771 gegen 1500 (Block `basis 94dafe9`). |
| Abdeckung | `go test -coverprofile` (Image `odata-recorder:race`, `--memory=4g`): Paket 94,1 %; `stringWert`, `text`, `wert`, `pruefeWert`, `artPruefer`, `ausgeben` je 100,0 %. Der Befund der Abdeckung aus dem Plan stimmt. |
| `make a-check` | 0 Befunde (der Hinweis auf `tools/gegenproben` ist vorbestehend). `make lint`: Stufe lief durch (Schritt `CACHED`). |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -P` auf Backslash, `u` und vier Hex-Ziffern über beide Testdateien, `schreiber_ausgabe.go`, `schreiber_bau.go`: kein Treffer; die Treffer auf `\x00`/`\xff` sind Go-Hex-Escapes im Quelltext, gewollt. Escape und Texte entstehen aus `esc`, `bs`, `mkRune`. |
| Orakel | `mkEnc` (byteweise), `mkAdresse`, `mkStrings`, `encoding/json`, erwartete Entity aus dem Text: kein Aufruf von Schreiber-Interna, der Leser nur im Rundlauf Z04 als Prüfling. `mkEnc` schreibt den Escape für den Marker als `esc("005b")` fest, nicht aus dem Schreiber. |
| Adresse Z01 | `Alles('%5BREDACTED%5D')` stimmt mit `SPEC-169` Nr. 2 (`'` unkodiert, `[` `]` als `%5B` `%5D`); `mkAdresse` ist ein eigener Kodierer. |

## 2. Mutationen (`tools/harness/mutationslauf.sh`, Klasse `andere`, `flock`, `--memory=4g`, Liste und Generator im Scratchpad)

121 Mutationen in einem Lauf (121 s, Host-Last 3,2 von 4): 118 `rot` im benannten Test, 1 `rot-paket`, 2 Übersetzungsfehler (Mutation, nicht Code), 0 `gruen`. Die Mengen aus Z03/M1 habe ich als Schleifen über ihre Achsen mutiert, nicht über die Elemente des Plans:

| Achse | Mutationen | Ergebnis |
|---|---|---|
| Abweichung: Zeichen an Stelle 0 bis 9 beliebig (Ersetzung, auch mehrbyte) | 10 | alle rot (`TestMarkerNachbarn`) |
| Abweichung: Buchstabe an Stelle 1 bis 8 klein | 8 | alle rot |
| Kürzung Präfix/Suffix 1 bis 9 | 16 + 2 (Länge 9) | alle rot |
| leer, doppelt, enthält, Rest davor/dahinter beliebig | 5 | alle rot |
| Rest 12 Zeichen × Seite 3 (davor, dahinter, beide) | 36 | alle rot |
| Rest-Länge 4096/65 537 × 2 Seiten | 4 | alle rot |
| Escape-Folge für `[` im Text (Hex-Ziffern klein und groß) und Backslash davor | 3 | alle rot |
| Escape des Markers: Hex `B`, `5a`, `5c`, `4b`; zweites, letztes Zeichen escapet; ohne Escape; `EDACTED`-Verlust; Zweig weg (Bytes, Rohtext, Rundlauf) | 11 | alle rot |
| `dst`: `n < len(body)-k` für k = 1, 3, 4, 5, 6 und `+1`; `ctx` nach dem `Write` | 7 | alle rot |
| Art: `kanon` der 7 Arten nimmt den Marker an | 7 | alle rot (`TestMarkerAndereArten`) |
| Typ: Art `string` an jeder der 7 fremden Properties durchgelassen | 7 | alle rot |
| `redigiert` durchgelassen; `null` mit Text | 2 | rot |
| `null` an nicht nullbarer Property durchgelassen | 1 | `rot-paket` (siehe I-2) |
| Z09: Argument verändert (alle Texte; nur der Marker) | 2 | rot (`TestMarkerArgumente`) |

Elemente, die der Plan nicht nannte und die ich rot sah: Ersetzung an jeder Stelle 1 bis 8, Rest `NBSP`, `BOM`, `NUL`, `[`, `]`, `"`, `\` einzeln je Seite, Länge 65 537 auf beiden Seiten, Hex-Ziffer `4b` und `5c`, `n < len(body)-3/-4/-6`, `kanon` von `int32`, `int64`, `decimal`, `date`.

## 3. Findings

### F-1 — LOW — Z06: der Fänger fängt nur beide Wächter zusammen

- `kategorie`: LOW (nahe MEDIUM „Akzeptiertes Negativ ohne Grund und Folge“; der Grund steht, die Folge nicht)
- `quelle`: `AGENTS.md` §3.15
- `pfad`: `docs/plan/planning/in-progress/slice-v4-schreiber-marker-text.md` · „jeder Wächter allein ist gleichwertig (grün, siehe Absatz *Bericht des Implementers*)“
- `befund`: Selbst gefahren (Wurzel des Baums, einzeln): Prüfung der Art in `artPruefer` entfällt = `gruen` in jedem Test des Pakets (`TestMarkerUTF8` 2 Prozesse, kein Fänger); Prüfung in `text` entfällt = `rot-paket` (`TestMarkerUTF8` grün, `TestUngueltigesUTF8AnVierStellen` und `TestMarkerZweige` rot). Beide zusammen, und der Wächter der Art zusammen mit der Bedingung „Präfix 9“ oder „Suffix 9“ in `stringWert`, sind rot in `TestMarkerUTF8` (je in einer Kopie gefahren). M5 erreicht den Zweig des Escapes mit keinem Text (kein Text von M5 ist der Marker), die Zeile Z06 hat also für keinen Wächter einen isolierenden Fänger; sie steht als Fänger-Zeile in der Tabelle, ist aber ein Negativ ohne Folge. `artPruefer` trägt eine zweite, redundante Prüfung (`utf8.ValidString`), die kein Test unterscheidet.
- `verifizierbar`: ja (Mutationslauf, Listen 3 und 4 im Scratchpad)
- `klasse`: „Doppelter Wächter: Fänger rot nur für beide, Einzelfall Negativ ohne Folge“

### F-2 — LOW — Z13 und Z10: „meldet jede neue Verzweigung“ ist zu weit

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.11
- `pfad`: Plan §3 · „Z10 meldet jede neue Verzweigung“ (Z13, Folge); Z10 · „ein zweiter `if` in `stringWert` ohne Zeile in W“
- `befund`: `TestMarkerZweige` zählt Knoten in genau fünf Funktionen (`wert`, `stringWert`, `text`, `pruefeWert`, `ausgeben`). Eine Verzweigung in `artPruefer` (acht `case` je Art, auf dem Pfad jedes Wertes), in `zahl` oder in einer neuen Hilfsfunktion, die `stringWert` ruft, ändert die Zahl nicht. Das Paar der Zusage („jede neue Verzweigung“) und des Fängers ist damit schmaler als gesagt; ein Verhaltensfehler dort fiele nur über M4/M1 auf, nicht über Z10.
- `verifizierbar`: ja (Lesen von `mkZweige`/`TestMarkerZweige`)
- `klasse`: „Zusage über Verzweigungen reicher als die Funktionsliste des Zählers“

### I-1 — INFO — M1: 143 Einträge, 139 verschiedene Texte

Die Zählprüfung (`len(texte) != 143`, 4004, 3976, 2002, 286) zählt Einträge, nicht Texte; Plan §3 und §6 sagen das offen (Ausgang *weiter offen*, Adresse Planner). Kein Element der Klasse fehlt: die vier Doppelten sind je zwei Namen desselben Textes (Streichung an Stelle 0/9 gleich Suffix/Präfix 9, `x` davor/dahinter gleich Einfügung vor Stelle 0/10), und jede meiner 100 Mutationen an den Achsen von M1 ist rot. Die Zählprüfung ist als Zählung der Tabelle ehrlich; sie belegt nicht die Zahl verschiedener Elemente. Zuständig: Planner.

### I-2 — INFO — `Z05`-Zeile deckt `null` an nicht nullbarer Property nicht

Die Mutation „`null` an nicht nullbarer Property durchgelassen“ (`w.Text != ""` allein) ist nur im Paket rot (`TestArtGegenTyp`, `TestMarkerZweige`), nicht in `TestMarkerAndereArten`. Das gehört zu `SPEC-168` Nr. 4 und nicht zum Marker; M4 nennt die Zelle „`null` mit dem Text an nullbarer Property“ und lässt sie aus. Kein Handlungsbedarf, nur Zuordnung. Zudem bricht `TestMarkerZweige` auch bei einer wortgleichen, wirkungsgleichen Umformulierung der Bedingung (`s == markerText && len(s) == 10 && s[0] == '['` = `rot-paket`): so gewollt (Wortlaut je Zeile), der Autor einer Umformulierung ergänzt W.

## 4. Zu den Schwerpunkten ohne Befund

- **`TestMarkerZweige`, Robustheit:** `os.ReadDir(".")` im Paketverzeichnis (Arbeitsverzeichnis von `go test`), läuft im Container (Stufen `test`, `race`) und im Mutationslauf; kein Netz, kein Schreibzugriff, `go/parser` ohne Typprüfung. Entfernt oder verschiebt eine Änderung eine Verzweigung der fünf Funktionen, ist er rot (Mutationen `zweig-weg`, `pos*` ändern den Wortlaut und färben ihn rot; eine ersatzlose Entfernung der Prüfung in `text` ist in `nur-text-weg` rot). Eine Funktion der fünf kann nicht ungezählt umgezogen werden (`gefunden[fn]` fehlt, Fehler). Grenze: siehe F-2. Der Test prüft von den Fängernamen nur, dass ein Test dieses Namens im Paket steht, nicht, dass er die Zeile ausführt (der Plan sagt „im Paket steht“).
- **Z01-Adresse, Z02, Z04, Z07 bis Z09, Z11:** Mutationen oben; Z11 (Code-Katalog) nicht eigens mutiert, die Position `raw[2+2*i]` stimmt mit der Reihenfolge `code`, `message`, `target`.
- **Plan folgt Code (§3.9):** §3 (Datei-Tabelle, Z-Tabelle, Absatz *Bericht*), §6 und Kopf stimmen mit dem Diff; die Wortlaute von W05/W06/W07/W15 im Code (`mkZweige`) nennen etwas mehr Fänger als die Plan-Zeilen (`TestMarkerAndereArten` bei W05, `TestMarkerUTF8` bei W07), nicht weniger. DoD §2 Zeile 2 (Schreiber) und `make gates` sind noch nicht abgehakt, das ist Sache des Planners nach dem Verifier.
- **Merge-Hinweis zur Historie in `spec/spezifikation.md`:** Plan §6 nennt Ort, Richtung und Auflösung („beide Zeilen bleiben“); ich habe `main` nicht gemerged geprüft. Der Zweig trägt eine Spec-Änderung gegenüber `main` (`10`/`4` Zeilen); der V2-Zweig kommt voraussichtlich zuerst, dann muss dieser Zweig `main` aufnehmen, bevor der Verifier den Merge bestätigt.

## 5. Geprüft, ohne Befund

- `src/adapters/driven/odata/jsoncodec/schreiber_ausgabe.go` · `schreiber_bau.go`: geprüft, ohne Befund (Kommentare tragen Zusage/Kopplung; kein Konjunktiv, keine abwesende Fassung).
- `src/adapters/driven/odata/jsoncodec/schreiber_marker_test.go`: geprüft, ohne Befund außer I-1.
- `src/adapters/driven/odata/jsoncodec/schreiber_marker_zweige_test.go`: geprüft, Befund F-1, F-2.
- `spec/`, `harness/`, `.golangci.yml`, `.a-check.yml`: vom Diff nicht berührt (Spec-Änderungen stammen aus dem Commit des Architect, Basis `94dafe9`).
- Leser (`scan.go`, `leser*.go`, `wert.go`) und `jsoncodecv2`: unverändert.
