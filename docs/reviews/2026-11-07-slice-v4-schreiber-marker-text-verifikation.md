# Verifikation: slice-v4-schreiber-marker-text — 2026-11-07

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-169` Nr. 4, 8 und *Ungeprüft*, `SPEC-168` Nr. 4, `SPEC-126` Nr. 10 f und die Hard Rules; nicht Diff-Review (siehe `2026-11-07-slice-v4-schreiber-marker-text-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-v4-schreiber-marker-text`, HEAD `956e1c3` (Implementer-Stand `c2f296c`, Review-Commits `bb5e935`, `956e1c3`), Basis `94dafe9`. Plan: `docs/plan/planning/in-progress/slice-v4-schreiber-marker-text.md`.

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden, mit Auflagen an den Planner (Closure), kein Befund MEDIUM oder höher, keine Fix-Runde.** Der Produktcode setzt die Zusage von `SPEC-169` Nr. 4 um, der Leser ist unverändert, `make gates` ist grün (eigener Lauf), die Größenmarke hält (771 von 1500), und von 20 eigenen Mutationen an Elementen, die weder Plan noch Review nannten, blieb keine grün. Die DoD-Punkte sind unten einzeln aufgelöst; offen sind nur die Punkte, die dem Planner bei der Closure gehören.

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/verif-v4/gates.log`: `exit=0`; `a-check` (Hinweis auf `tools/gegenproben` vorbestehend), `baseline-verify` OK, `d-check` 725 Dateien 0 Befunde, `race`-Stufe gebaut und grün, `migration-check: ok`, `zeichen: ok`, Lint-Stufe durchgelaufen. |
| `make groessenmarke` | im selben Lauf: `slice-v4-schreiber-marker-text.md: ok, Messung 771 von Marke 1500`; eigene Rechnung `git diff --numstat 94dafe9` über die vier Pfade des Blocks: 14 + 2 + 424 + 331 = 771. |
| Plan-vs-Code-Diff | `git diff 94dafe9..HEAD -- src`: genau vier Dateien. Produktcode: `schreiber_ausgabe.go` +14 (Konstante `markerText`, Methode `ausgabe.stringWert`) und `schreiber_bau.go` 1/1 (`a.text(w.Text)` wird `a.stringWert(w.Text)` im `case model.ArtString`). Das ist die Datei-Tabelle des Plans (§3), nicht mehr und nicht weniger. `scan.go`, `leser*.go`, `wert.go`, `jsoncodecv2/` unverändert. `spec/` seit `94dafe9` unverändert (Spezifikation vom Architect vor dem Code geschrieben). |
| Code gegen Vertrag | `stringWert` schreibt `"`, Backslash, `u`, `005b` (Kleinbuchstaben), `REDACTED]`, `"` genau bei `s == markerText` (zehn ASCII-Zeichen), sonst `text`; das ist der Wortlaut von `SPEC-169` Nr. 4 Absatz *Der Text `[REDACTED]`*. Andere Arten und Typen und `redigiert`: `ErrUngueltig` über den vorbestehenden `pruefeWert` (`SPEC-168` Nr. 4). Ein Fehlerpfad, Kontext und Kurzschreiben bleiben in `ausgeben` unverändert; Z07 fängt die Länge mit den fünf Escape-Bytes. |
| Abdeckung (Register `zweig-ohne-lauf-unter-dem-detektor`, 6×) | eigenes `go test -coverprofile` (Image `odata-recorder:race`, `--memory=4g`, netzlos): Paket 94,1 %; `stringWert`, `text`, `wert`, `pruefeWert`, `artPruefer`, `ausgeben` je 100,0 %. Kein Zweig der berührten Funktionen ohne Lauf. Der Schreibpfad ist nicht nebenläufig (kein Zustand, `SPEC-168` Nr. 9); der Detektor hat hier nichts zu melden, der Zweig-Fänger ist die Zählprüfung `TestMarkerZweige`. |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -P '\\u[0-9a-fA-F]{4}'` und BOM, NBSP, U+2028 über beide Quelldateien, beide Testdateien, den Plan und den Review-Report: je 0 Treffer. Escape und Texte der Tests entstehen aus Bytewerten (`esc`, `bs`, `mkRune`). Der Bericht des Implementers nennt die Gegenlesung ausdrücklich. |
| Merge-Hinweis (Historie `spec/spezifikation.md`) | Der Zweig steht 50 Commits hinter `main`. Probe-Merge ohne Eingriff in den Baum (`git merge-tree --write-tree HEAD main`): Rückgabe 0, kein Konflikt; die Historie im Ergebnis trägt die Zeile `2026-11-07` dieses Slice über den Zeilen `2026-11-06` aus `main`, beide bleiben. Ob `slice-v2-payload-schreiben` schon in `main` steht, habe ich nicht geprüft; die Probe ist gegen den Stand von `main` zum Zeitpunkt der Verifikation. Auflage: `main` in den Zweig mergen und `make gates` erneut vor dem Merge nach `main`. |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Vertrag: Form für `[REDACTED]` und Randformen vom Architect vor dem ersten Code-Commit | erfüllt (abgehakt) | `SPEC-169` Nr. 4, 8, *Ungeprüft*, `SPEC-168` Nr. 4; Commit `94dafe9` liegt vor allen Code-Commits; Randformen sind M1, Raster in §6 mit Ort oder „gilt nicht“. |
| Schreiber setzt den Vertrag um; Rundlauf als Schleife über die Klasse; Fänger rot ohne die Änderung | **erfüllt, Haken fehlt** | Rundlauf `TestMarkerRundlauf` über M3 (2002 Elemente); Fänger ohne die Änderung rot: Mutation `a.text(w.Text)` statt `a.stringWert` (Datei `schreiber_bau.go`) rot in `TestMarkerBytes`; Zweig entfernt (Review) rot. Der Planner hakt ab. |
| Vertragstext nennt keine Ausnahme mehr | erfüllt (abgehakt) | Nr. 8 sagt ausdrücklich, dass der Rundlauf den Text meint; ein Lauf über den Abschnitt findet kein Wort Ausnahme. |
| `make gates` grün | **erfüllt, Haken fehlt** | eigener Lauf, `exit=0`. Nach dem Merge von `main` einmal wiederholen. |
| Review durchgeführt, Report liegt vor | erfüllt (Haken fehlt) | `docs/reviews/2026-11-07-slice-v4-schreiber-marker-text-review.md`, kein HIGH/MEDIUM. |
| Doku-Update | zu prüfen vom Planner | Der öffentliche Vertrag steht in `spec/` (vom Architect); `make handbuch-check` ist grün; ob `docs/user/` den Text berührt, habe ich nicht gesondert geprüft. |
| Closure-Notiz, Register, Ausgänge der Risiken, drei Paarungen | offen, Sache des Planners | siehe Auflagen. |

Die Liefer-Punkte (zwei) bleiben unter dem Limit von drei.

## 3. Tabelle *Zusage · Fänger* gegen den Vertrag

Ich habe jede Zeile Z01 bis Z13 gegen den Wortlaut von `SPEC-169` Nr. 4 und 8 und `SPEC-168` Nr. 4 gelegt. Jede Aussage des Vertrags hat eine Zeile: Escape-Form und Stelle (Z01), Allaussage über den Rohtext (Z02), Nachbarn byte-gleich (Z03), Rundlauf (Z04), andere Arten und `redigiert` (Z05), UTF-8 (Z06), Fehler von `dst` (Z07), Kontext und Vorrang des Arguments (Z08), keine Änderung des Arguments, Wiederholung (Z09), Zweige (Z10), Fehlerobjekt (Z11), zwei Negative mit Grund und Folge (Z12, Z13). Randformen: das Raster in §6 hat in jeder Zelle einen Ort oder „gilt nicht“ mit Grund; die Zelle *Text in Collection-Property* nennt `SPEC-168` Nr. 3 c (`ErrUngueltig`), das ich gegen den Spec-Text bestätigt habe.

**Stichproben-Mutationen an nicht genannten Elementen** (`tools/harness/mutationslauf.sh`, Klasse `andere`, `flock`, `MUTATIONSLAUF_SPEICHER=4g`, Image `odata-recorder:race`, Liste `.../scratchpad/verif-v4/liste.json`, Ausgabe `.../scratchpad/verif-v4/mut.out`; 20 Mutationen, 43,6 s, Host-Last 2,4 bis 2,5 von 4, Messender das Werkzeug; keine Wahrscheinlichkeit im Spiel, die Fänger sind deterministisch):

| Gruppe | Mutationen (nicht im Plan und nicht in den 121 des Reviews) | Ergebnis |
|---|---|---|
| Escape-Bytes | Ziffern `006b`; großes `U`; Rest ab Stelle 2; Rest `REDACTED}`; Schlussquote fehlt; Escape schließt `]` (`]`) | 6 rot; `]` ist im Rundlauf wertgleich (zu Recht grün dort) und in den Byte-Tests rot (`rot-paket`) |
| Konstante, Vergleich | Konstante mit Leerzeichen; `[Redacted]`; `!=` statt `==`; `ungueltig` statt Escape; `a.text` statt `a.stringWert` | 5 rot |
| Nachbarn (Klasse erweitert) | Escape auch für leeren Text; auch für den Marker doppelt | 2 rot |
| Andere Arten | Marker als kanonisch für `int64`, `decimal`, `date`; `null` mit Marker-Text | 4 rot |
| Länge | `n < len(body)-6` | rot |
| Außerhalb von M1 | Escape auch für den Text `REDACTED` (ohne Klammern); auch für Texte über 70 000 Zeichen | **rot-paket**, nicht im benannten Test: nur `TestMarkerZweige` wird rot (Wortlaut von W06) |

Ergebnis: 20 Mutationen, 17 `rot`, 3 `rot-paket`, 0 `gruen`. Die beiden Mutationen außerhalb von M1 sind genau die akzeptierte Prüfgrenze Z13 (der Text `REDACTED` ohne beide Klammern und ein Text über 65 547 Bytes sind keine Elemente von M1); die Zeile Z13 nennt diese Grenze mit Grund, und die Mutationen zeigen zugleich, dass der Wortlaut-Fänger von Z10 an der Stelle eine Zusatzwirkung hat (eine geänderte Bedingung in `stringWert` ist rot). Das Gegenstück siehe F-V2.

## 4. Befunde zu den Findings des Reviews (Auflösung, keine Fix-Runde)

Beide LOW-Findings bleiben **LOW**: nach meiner Prüfung liegt keine Wirkung im Produkt, kein Vertragssatz ist ohne Fänger, und die Lücken sind Wortlaute des Plans. Ich löse sie so auf.

### F-1 (Z06, UTF-8-Wächter doppelt) — akzeptiertes Negativ, Wortlaut zu berichtigen

Bestätigt gegen den Code: `artPruefer` gibt für `ArtString` `utf8.ValidString` zurück, `text` prüft dasselbe noch einmal und setzt `a.err`. Für jeden Wert, den `artPruefer` abweist, weist `text` denselben Fehler ab; die Prüfung in `artPruefer` ist also für das Verhalten äquivalent (äquivalente Mutation, kein Fänger möglich). Die Prüfung in `text` ist von `TestUngueltigesUTF8AnVierStellen` und `TestMarkerZweige` gefangen (Review: `rot-paket`). Z06 sagt „Fänger“, der Fänger ist aber nur das Paar. Grund und Folge, wie `AGENTS.md` §3.15 sie verlangt:

- **Grund:** zwei Wächter für denselben Fehler (Defense in Depth, `artPruefer` für die Prüfung vor dem Schreiben, `text` für die Eigenschaft von `text`); ein isolierter Fänger für den ersten existiert nicht, weil der zweite den Fehler abfängt.
- **Folge im Betrieb:** entfällt `artPruefer`'s Prüfung allein, ändert sich kein Byte und kein Fehler des Schreibers; entfällt die in `text`, fängt sie ein älterer Test (`TestUngueltigesUTF8AnVierStellen`); entfallen beide, fängt `TestMarkerUTF8` (Z06).
- **Auflage an den Planner:** in Z06 die Wendung „jeder Wächter allein ist gleichwertig (grün, …)“ durch Grund und Folge ersetzen, den Fänger als **Paar** benennen und die Zeile als akzeptiertes Negativ für die Isolation führen (Wortlaut-Berichtigung, `AGENTS.md` §3.11); kein Code und keine neue Implementer-Runde.

### F-2 (Z13/Z10: „meldet jede neue Verzweigung“) — Wortlaut zu berichtigen

Bestätigt: `mkZweige` führt fünf Funktionen (`wert`, `stringWert`, `text`, `pruefeWert`, `ausgeben`), `TestMarkerZweige` vergleicht je Funktion die Liste der Verzweigungen. Eine Verzweigung in `artPruefer` (acht `case`) oder `zahl` oder in einer neuen Hilfsfunktion fängt er nicht. Die Zusage in Z13 ist damit weiter als der Fänger. Keine Wirkung im Produkt: die Funktionen, in denen die Entscheidung über den Text fällt (`stringWert`, `text`, `wert`), sind gezählt, und meine Stichprobe oben zeigt, dass ein geänderter Operand in `stringWert` rot wird (Wortlaut je Zeile).

- **Auflage an den Planner:** Z13 *Folge im Betrieb* und Z10 berichtigen: „Z10 meldet jede neue Verzweigung **in den fünf Funktionen von W** (und jede Änderung ihrer Bedingung); eine Verzweigung in `artPruefer`, `zahl` oder einer neuen Hilfsfunktion meldet er nicht“. Optional im Closure-Lerneintrag: Funktionsliste der Zählprüfung ist eine Setzung, kein Aufruf-Graph.

### I-1 (M1: 143 Einträge, 139 verschiedene Texte)

Bestätigt und im Plan offen genannt (§3 Absatz *Bericht*, §6, Ausgang *weiter offen*, Adresse Planner). Kein Element der Klasse fehlt: die vier Doppelten sind zwei Namen desselben Textes; meine 20 Mutationen am Rand der Klasse (Nr. 1, 10 bis 13 oben) und die 121 des Reviews sind rot. **Auflage an den Planner:** bei der Closure den Ausgang setzen — *eingetreten, akzeptiert*: die Zählprüfung zählt Einträge der Tabelle, nicht verschiedene Texte; Folge: keine, weil die Doppelten dieselben Bytes zweimal prüfen; Grund: die Tabelle der Achsen in §3 ist so zugesagt. Kein Code.

### I-2 (`null` an nicht nullbarer Property nur im Paket rot)

Gehört zu `SPEC-168` Nr. 4 und ist von `TestArtGegenTyp` gefangen. Kein Handlungsbedarf.

## 5. Neue Befunde dieser Verifikation

- **F-V1 — INFO — Zusage „18 Zeilen, Fänger im Paket“ prüft Namen, nicht Ausführung.** `TestMarkerZweige` prüft, dass ein Test mit dem genannten Namen im Paket steht, nicht, dass er die Zeile erreicht; der Plan sagt „im Paket steht“ und nichts mehr (§3.11 gewahrt). Die Abdeckung (100 % der fünf Funktionen, siehe oben) belegt die Ausführung einmalig; kein Gate hält sie. Kein Handlungsbedarf, nur benannt.
- **F-V2 — INFO — Prüfgrenze Z13 an zwei konkreten Elementen belegt.** Der Text `REDACTED` (beide Klammern gestrichen) und ein Text über 65 547 Bytes sind außerhalb von M1; eine Mutation, die dort den Escape schreibt, wird nur durch den Wortlaut-Fänger von Z10 rot, nicht durch einen Wert-Fänger. Das ist Z13 wie geschrieben (unbeschränkte Menge, Prüfgrenze mit Grund und Folge); kein Befund, die Zeile verlangt nur die Berichtigung aus F-2.

Keine Spec-Lücke gefunden: `SPEC-169` Nr. 4 trägt die Entscheidung, Nr. 8 den Rundlauf, *Ungeprüft* das eine Negativ (Z12); der Wortlaut stimmt mit dem Code überein.

## 6. ADRs und Hard Rules

- `AGENTS.md` §3.4 (Architektur meilensteinfrei) und §3.8 (keine Regeldetails in ADRs): keine Datei in `spec/architecture.md` oder `docs/plan/adr/` berührt; keine ADR nötig (Plan §3 Nr. 6).
- §3.9 (Plan folgt Code): Datei-Tabelle, Z-Tabelle, Bericht und §6 stimmen mit dem Diff; die Wortlaute von W05, W06, W07, W15 im Test nennen etwas mehr Fänger als der Plan (nicht weniger), laut Review.
- §3.10, §3.15, §3.17: je Zusage Mutation und roter Test im Plan (Tabelle in §3); die Mengen M1 bis M8 laufen als Schleifen; Mutationen an nicht genannten Elementen siehe Abschnitt 3 und das Review.
- §3.12 und §3.16: Raster in §6 gefüllt; keine Zelle im Code entschieden.
- §3.14: Marke von Planner vor dem ersten Code-Commit gesetzt (`9aa2098`), Messung 771 gegen 1500, keine Rückführung.
- §3.18: Rohbytes gegengelesen, siehe oben.
- Gates unverändert (kein Gate gelockert); keine `//nolint`.

## 7. Auflagen an den Planner (Closure), nicht an den Implementer

1. Z06 und Z13/Z10 im Plan §3 im Wortlaut berichtigen (F-1, F-2); keine Code- und keine Teständerung.
2. Ausgang für den M1-Eintrag in §6 setzen (I-1), Ausgang der Randform-Zeile und der Pflicht-Mengen-Zeile (weiter offen bis zu diesem Report, jetzt *entfallen* mit Verweis auf Review und diesen Report).
3. Haken in §2: Schreiber, `make gates`, Review; Closure-Notiz, Register (Beobachtung: „äquivalente Mutation hinter doppeltem Wächter“ und „Zusage über Verzweigungen reicher als die Funktionsliste des Zählers“ als neue Dateien oder weitere Evidenz im Register), Lerneintrag.
4. Vor dem Merge nach `main`: `main` in den Zweig mergen (Probe sauber), `make gates` einmal wiederholen.

Keine Aktion an `main`, kein Push.
