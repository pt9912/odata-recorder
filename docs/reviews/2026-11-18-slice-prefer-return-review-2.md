# Review-Report: slice-prefer-return, Runde 2 — 2026-11-18

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules; nicht gegen die DoD, das ist die Frage des Verifiers)

**Gegenstand:** `slice-prefer-return`, Zweig `slice-prefer-return`, Fix-Runde `5b09dd6a..9a6ee7dc` (Commits `099c6ee4`, `f329146b`, `9a6ee7dc`; `5b09dd6a` ist die Entscheidung des Architect). Pfade: `src/hexagon/model/rueckgabe.go` und `rueckgabe_test.go`, `src/hexagon/services/mock/schreiben_antwort.go` und `prefer_test.go`, `src/hexagon/services/hybrid/prefer_test.go`, `src/adapters/driving/httpodata/prefer_test.go`, `src/bootstrap/prefer_ende_zu_ende_test.go`, `spec/spezifikation.md` (`SPEC-285` *Ungeprüft*), Plan `slice-prefer-return` (DoD, §3 Tabellen, §4 Marke, §6 Raster und Ausgänge). Erste Runde: `docs/reviews/2026-11-18-slice-prefer-return-review.md`.

**Skill:** `.harness/skills/reviewer.md` @ `9a6ee7dc` <!-- d-check:ignore (Zitat des Standes im Lauf) -->
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext** (die Verträge, gegen die geprüft wurde):

- Slice-Plan `slice-prefer-return` (Marke 3500, `make groessenmarke`: Messung 2123)
- `SPEC-285` bis `SPEC-289` (Vertrag), `SPEC-194`, `SPEC-195` Nr. 11, `SPEC-192`, `SPEC-200` Nr. 1
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-07.b`
- `AGENTS.md` §3.7, §3.9 bis §3.18; Befunde F-1 bis F-13 der ersten Runde

**Vorgehen und Belege.** Jeder Fix wurde gegen die Klasse der Zusage aufgelöst (`AGENTS.md` §3.17): die Menge aus dem Wortlaut von `SPEC-285` bis `SPEC-289`, die Mutationen an Elementen, die weder der Befund noch der Bericht des Implementers nannten (die Kennungen sind die des Reviewers). Gefahren (Image `odata-recorder:race`, `--network none`, `--memory=4g --cpus=4`, `sh tools/harness/mutationslauf.sh`, Klasse `andere` ein Prozess je Stufe, unter `flock` auf der Sperrdatei der schweren Läufe; Last des Hosts laut Spalte `last` 6,7 bis 22 von 4 CPUs, gemessen vom Reviewer):

| Liste | Gegenstand | Mutationen | rot | rot-paket | grün |
|---|---|---|---|---|---|
| modell | F-1: Backslash im Wort (BA1 bis BA13) und im fremden Element (EA1 bis EA4) | 15 | 12 | 2 (BA11, EA2) | 1 (EA4) |
| body | F-3: `Body` nil an der `204`, Bedingung je Dimension (N1 bis N9 ohne N2) | 8 | 5 | 3 (N4, N6, N8) | 0 |
| zeit | F-4: quadratische Stellen (QA, QD, QF, QK, QG, QW) und QR | 7 | 4 | 1 (QR) | 2 (QG, QW) |
| hybrid | F-2: Form von `vorbereiten` × Methode × Stufe | 35 | 34 | 0 | 1 (HU-U6) |
| gegenstand | F-11: Gegenstand je Stelle S24, S25, S26 (je drei nicht genannte) | 9 | 9 | 0 | 0 |
| rate | F-4: Rate von `TestLiesRueckgabeZeit` (nebenläufig, `n_ohne` 8 bis 20) | 4 | 3 | 0 | 1 (Z0, Nullprobe) |

Dazu zwei Läufe im Container ohne das Werkzeug (`src/bootstrap` läuft dort nicht, `SPEC-210` Nr. 5 e): E4 (Antwort mit dem Body der vorigen Mutation, Paketvariable je Schlüssel) rot in `TestPreferWiederholungGleichGet` und `TestPreferEntityGleichGet`; eine Probe `TestProbeBackslashFolge` (nicht im Repo) gegen EA4: der unveränderte Stand besteht, EA4 liefert für Folgen von 3, 5 und 7 Backslashes vor dem Schlusszeichen `Minimal` statt `Keine`. Gate-Läufe: `make groessenmarke` (Hinweis: Tabelle ergibt 2905, Schätzung 2585; ok, Messung 2123 von Marke 3500), `make docs-check`, `make zeichen-check`, `make a-check` grün. Die Nullprobe des Werkzeugs bestätigt `model`, `mock` und `hybrid` auf dem Stand grün. Rohbytes gegengelesen (`LC_ALL=C grep -P` auf die Folge aus Backslash, `u` und vier Hex-Ziffern): die Listen des Reviewers und dieser Report, kein Treffer.

**Rate des Zeit-Fängers** (`TestLiesRueckgabeZeit`, je Prozess `-test.count=1`, ohne Detektor, Last 6,7 bis 8,1 von 4 CPUs, Messender: der Reviewer): Nullprobe Z0 (nur ein Kommentar) 0 von 20 Prozessen rot; QD 8 von 8, QF 8 von 8, QK 8 von 8. Eine Einzelzahl gilt als Größenordnung; sie stimmt mit den Raten im Plan überein (Z0 0 von 60, Q1 bis Q5 alle rot).

---

## Auflösung der Befunde der ersten Runde

| Befund | Stand | Beleg |
|---|---|---|
| F-1 Backslash im Wort | Wort-Stelle behoben; die Klasse „Backslash“ bleibt an der Stelle „Paar im fremden Element“ offen (N-2) | BA1 bis BA7, BA10, BA12, BA13 rot in `TestLiesRueckgabeBackslashWort` (je ein nicht genanntes Element: Position, Großbuchstabe, Buchstabe nur in `representation`, Zusammenhang, Länge der Zeile); BA11 gleichwertig (`elementEnde` vom Backslash aus liefert in jedem Wort das Zeilenende), rot-paket durch `TestLiesRueckgabeZeit` |
| F-2 `vorbereiten` mit Body | behoben | 34 von 35 rot: Anfrage an R in Zweig 1 und 6 je Methode (10), Anfrage an M in Zweig 1, 2, 3, 4, 5, 6 je Methode (21), erste Zeile von `Prefer` (2), kanonischer Name (1); HU-U6 grün, gleichwertig (`stufen.upstream.Header` wird nirgends gelesen, der Upstream erhält `upstreamHeader(a.Header)`) |
| F-3 `Body` nil an der `204` | behoben | N1, N3, N5, N7, N9 rot in `TestPreferAntwortTabelle`; N4, N6, N8 rot-paket (`TestPreferApplied`, `TestPreferLesendOhneWirkung`, `TestPreferAdresseNieAusHost`); der Haken sitzt in `preferSende` und trifft jeden Aufruf der Prefer-Tests |
| F-4 Zeit | teilweise; die Stellen des langen Laufs sind nicht gezählt (N-3) | QA, QD, QF, QK rot; QG, QW grün; QR nur durch das Zeitlimit eines anderen Tests |
| F-5 Raster und Wiederholung | behoben | Spalten *Abbruch* und *Wiederholung* im Raster, Zeilen 27 und 28 in §3; E4 rot in beiden Tests; POST fehlt in Zeile 27, weil ein zweites `POST` mit demselben Schlüssel kein zweiter Eintrag ist, sondern ein Konflikt (kein Befund) |
| F-6 Kommentar `kopf204` | behoben | Wortlaut nennt „jede 204“ |
| F-7 „Stelle 13“ | behoben | Kopf nennt 14 × 7, Test zählt `14*7` |
| F-8 Prüfgrenze `tchar` | behoben | Menge nachgerechnet: 6 Trenner und 24 Buchstaben (zwölf, je in zwei Schreibweisen) sind die 30 Bytes mit Wirkung, 226 gleichwertig; der Grund (jedes andere Byte macht Name oder Wort unlesbar, ob Token oder nicht) trägt für Name, Wert und Rest |
| F-9 Roadmap | behoben | `welle-etag` führt `slice-prefer-return` als „in Arbeit“ |
| F-10 Handbuch | behoben mit Rest (N-5) | Architect Weg B, `SPEC-200` unverändert; DoD-Punkt entfallen, Zeile 26 „gilt nicht“ mit Grund, §4 umgeschrieben |
| F-11 Gegenstand | behoben | je Stelle drei nicht genannte Gegenstände (S24: Dienst, Snapshotname, Exchange; S25: Dienst, Import, Workspace; S26: Dienst, Import, Session) rot in `TestPreferSchreiberFehler` |
| F-12 „byte-gleich“ | behoben | der Kopf nennt Werte und kanonischen Namen und sagt, was offen bleibt |
| F-13 `tokenZeichen` | behoben | `istTokenByte` hat dieselbe Menge (Ziffer, ASCII-Buchstabe, 15 Zeichen); die Zusammenführung ändert kein Verhalten |

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| N-1 | HIGH | Ein Kommentar in neuem Testcode beschreibt die Folge der fehlenden Alternative im Konjunktiv („sonst bliebe ein Gegenstand je Stelle ungefangen“), nicht den Zustand; der erste Halbsatz ist Zusage und Kopplung, der zweite die verworfene Abwägung. | `AGENTS.md` §3.7; Reviewer-Skill (HIGH: Kommentar trägt keine der Kommentar-Klassen) | `src/hexagon/services/mock/prefer_test.go` · `sonst bliebe ein Gegenstand je Stelle ungefangen` | nein — Lektüre; kein Gate fängt Konjunktive | Kommentar nennt die verworfene Alternative |
| N-2 | MEDIUM | Die Zusage „ein Backslash im Paar schützt das folgende Zeichen“ (`SPEC-285` Nr. 2) hat an der Stelle „Paar in einem fremden Element“ (`elementEnde`) nur die drei Fälle der Zeile 6 und keinen Fänger für Folgen von Backslashes: EA4 (ein Backslash nach einem Backslash schützt nicht) bleibt in `TestLiesRueckgabeAnfuehrung` und im ganzen Paket `model` grün, obwohl `x="a` + drei Backslashes + `", return=minimal` dann `Minimal` statt `Keine` ergibt; EA2 (ein Backslash außerhalb eines Paares schützt) fängt allein `TestLiesRueckgabeFolgen` (rot-paket), nicht der benannte Test. Der Fix zu F-1 deckt die Klasse „Backslash“ an der Wort-Stelle, nicht an der Element-Stelle. | `AGENTS.md` §3.15, §3.17 (Fänger führt nur eine Stelle der Menge) | `src/hexagon/model/rueckgabe.go` · `if z[i] == '\\' {` (in `elementEnde`); Tests `rueckgabe_test.go` · `TestLiesRueckgabeAnfuehrung` | ja — Mutation EA4 (Liste modell), Probe `TestProbeBackslashFolge` rot gegen EA4 | Fix deckt die Klasse nicht (Backslash im Paar des fremden Elements, Folgen) |
| N-3 | MEDIUM | Der Fix zu F-4 führt sieben Formen langer Zeilen und nennt die Menge der übrigen „unbeschränkt“ (`SPEC-285` *Ungeprüft*, Plan §3 *Prüfgrenzen*); die Achse, die die Formen trennt, ist die Stelle des langen Laufs in der Grammatik, und sie ist endlich: Wert als Token (QW), Wort in Anführungszeichen ohne Backslash (QG) und der Rest außerhalb eines Paares in einem fremden Element (QR) sind ungezählt. QW und QG bleiben in `TestLiesRueckgabeZeit` und im Paket grün; QR macht allein das Zeitlimit von `TestLiesRueckgabeLeerraum` rot (`panic: test timed out`, rot, aber nicht aus dem benannten Test). | `AGENTS.md` §3.15, §3.17 (Menge ohne Achsenliste; Auswahl, die dem Fix einfiel) | `src/hexagon/model/rueckgabe_test.go` · `func zeitFormen()`; `spec/spezifikation.md` · `Das lineare Wachstum der Zeit (Nr. 1) für jede Form langer Zeilen außer sieben` | ja — Mutationen QW, QG (grün), QR (rot-paket durch Zeitlimit) | Negativ nennt die Menge unbeschränkt, die Achse (Stelle des langen Laufs) ist endlich und nur zu sieben Stellen geführt |
| N-4 | LOW | Der Eintrag zur Rückfrage (a) in §6 trägt zwei Ausgänge in einem Feld: „Ausgang: offen“ am Satz der Rückfrage und „Ausgang: entschieden“ am Ende der Messung; ein Zustandsfeld nennt einen Zustand, nicht den Weg dorthin. | `AGENTS.md` §3.7 (Zustandsfelder), §3.13 | `docs/plan/planning/done/slice-prefer-return.md` · `Ausgang:** offen; zeigt Zeile 18 einen Unterschied` | nein — Lektüre | Zustandsfeld trägt Chronik (zwei Ausgänge) |
| N-5 | LOW | Die Zeile `ETAG_ENABLED` der Schlüsseltabelle des Handbuchs weicht von `SPEC-200` Nr. 1 ab; der Plan weist ihre Berichtigung dem Planner zu (kein Slice als Nehmer) und führt die Zuweisung mit „Ausgang: entschieden“, obwohl die Berichtigung bei niemandem eingetragen ist. | `AGENTS.md` §3.13 (offene Rückgabe an den Planner steht mit Ausgang *offen* und Adresse) | `docs/plan/planning/done/slice-prefer-return.md` · `Adresse** der Zeile `ETAG_ENABLED`: Planner` | nein — Lektüre | Zuweisung ohne Nehmer mit Ausgang entschieden |
| N-6 | INFO | Die Grenze der Grammatik, die `EA2` trifft („ein Backslash außerhalb eines Paares hat keine Wirkung“), steht in keiner Zeile der Tabelle §3; der Fänger ist der Orakel-Test `TestLiesRueckgabeFolgen` (Folgen bis fünf Symbole). Gefangen, aber nicht vom benannten Fänger. | `AGENTS.md` §3.15 | `slice-prefer-return.md` · Zeile 6 der Tabelle *Zusage · Fänger* | ja — Mutation EA2 | Zusage gefangen durch den Orakel-Test, nicht durch den der Zeile |
| N-7 | INFO | `make groessenmarke` meldet „Tabelle ergibt 2905, Schätzung 2585“; die Marke 3500 bleibt fest (§4 nennt es mit `SPEC-128` Nr. 5 c), die Messung 2123 liegt darunter. Kein Befund, ein Hinweis des Gates. | `AGENTS.md` §3.14 | `slice-prefer-return.md` · `Nach der Fix-Runde hat die Tabelle 30 Zeilen` | ja — `make groessenmarke` | Hinweis des Gates, im Plan benannt |
| N-8 | INFO | `stufen.upstream.Header` (HU-U6) ist ungelesen; die Mutation daran ist gleichwertig. Der Plan nennt sie (H06, H12, H13); ein Leser, der später aus `stufen.upstream` statt aus `a` liest, ändert die Aussage. | Maintainability | `src/hexagon/services/hybrid/puffer.go` · `m, r, u := a, a, a` | nein | Totes Feld, gleichwertige Mutation benannt |

---

## Auflösung der Tabelle (Zeilen 6, 6a, 9a, 10, 21, 24, 27) gegen die Klasse

| Zeile | Menge aus dem Wortlaut | Eigene Mutationen an nicht genannten Elementen | Ergebnis |
|---|---|---|---|
| 6, 6a | Backslash an jeder Stelle, an der `SPEC-285` Nr. 2 ihn nennt: im Wort (Auflösen), im Paar eines fremden Elements (Schützen), außerhalb eines Paares (keine Wirkung) | BA1 (nur Positionen unter 4), BA2 (nicht an Position 0), BA3 (nicht vor Großbuchstaben), BA4, BA5, BA6, BA12 (nicht vor `a`, `p`, `o`, `i`), BA7 (nur Zeilen unter 30 Byte), BA10 (nicht nach einem Leerzeichen), BA13 (nicht an der ersten Stelle des Paares); EA1 (nur vor `"`), EA2, EA3 (nicht vor Backslash), EA4 (nicht nach Backslash) | Wort-Stelle: 10 rot, BA11 gleichwertig; Element-Stelle: EA1, EA3 rot, EA2 rot-paket, **EA4 grün (N-2)** |
| 9a | Stellen des langen Laufs in der Grammatik | QA (Zeilen), QD (Paar), QF (Leerraum), QK (Kommas), QW (Wert als Token), QG (Wort ohne Backslash), QR (Rest) | QA, QD, QF, QK rot, Raten 8 von 8 und Nullprobe 0 von 20; **QW, QG grün, QR rot-paket (N-3)** |
| 10 | Zellen der `204` | N1, N3, N4, N5, N6, N7, N8, N9 (je eine Bedingung: Etag, Accept, mehrere Zeilen `Prefer`, PATCH, DELETE, PUT ohne Wahl, POST ohne Accept, PATCH ohne Wahl mit Etag) | alle rot oder rot-paket |
| 21 | Gegenstand je Stelle | 9 Mutationen (3 je Stelle) | alle rot |
| 24 | Form von `vorbereiten` × Methode × Stufe | 35 Mutationen | 34 rot, 1 gleichwertig |
| 27 | Wiederholung | E4 | rot |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/rueckgabe.go` (nach F-13, `istTokenByte`) | geprüft, Befunde N-2, N-3; die Zusammenführung ändert kein Verhalten |
| `src/hexagon/services/mock/schreiben_antwort.go` | geprüft, ohne Befund (Kommentar `kopf204` nennt die Verwendung) |
| `src/hexagon/services/mock/prefer_test.go` | geprüft, Befund N-1 |
| `src/hexagon/services/hybrid/` (`puffer.go`, `prefer_test.go`) | geprüft, INFO N-8 |
| `src/adapters/driving/httpodata/prefer_test.go` | geprüft, ohne Befund |
| `src/bootstrap/prefer_ende_zu_ende_test.go` | geprüft, ohne Befund (E4 rot) |
| `spec/spezifikation.md` (`SPEC-285` *Ungeprüft*) | geprüft, Befund N-3 (Grund „unbeschränkt“); Wortlaut stimmt mit dem Test (sieben Formen, Verhältnis 32, Grenze 128) |
| Plan `slice-prefer-return` (DoD, §3, §4, §6) | geprüft, Befunde N-4, N-5, INFO N-6, N-7; DoD und Marke nachgezogen, Mengen-Tabelle trägt Zeilen 6a und 9a |
| `docs/plan/planning/in-progress/roadmap.md` | geprüft, ohne Befund |
| Hard Rules §3.1, §3.2, §3.18 | geprüft, ohne Befund (keine `//nolint`-Direktive, kein `d-check:ignore` im Diff, Rohbytes gegengelesen) |
| Sicherheit und Geheimnisse | geprüft, ohne Befund (kein Header der Anfrage im Text eines Fehlers, Adresse nie aus `Host`) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 1 |
| MEDIUM | 2 |
| LOW | 2 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Kommentar nennt die verworfene Alternative · Fix deckt die Klasse nicht (Stelle der Menge) · Negativ nennt die Menge unbeschränkt, die Achse ist endlich · Zustandsfeld trägt Chronik · Zuweisung ohne Nehmer mit Ausgang entschieden

## Verdikt

**Merge-blockierend:** ja — ein HIGH (N-1, ein Halbsatz in einem Testkommentar) und zwei MEDIUM (N-2, N-3). Die Fixes zu F-2, F-3, F-5 bis F-13 halten gegen die Klasse; F-1 und F-4 decken ihre Klasse an je einer Stelle nicht. Die Klasse bei N-2 ist „Backslash an jeder Stelle der Grammatik“ (Folgen von Backslashes im Paar des fremden Elements, Länge gerade und ungerade, ab drei), bei N-3 „Stelle des langen Laufs“ (Wert als Token, Wort ohne Backslash, Rest). Findings gehen an den Implementer; die Finding-Klassen gehen zusätzlich in die Slice-Closure §7. Dieser Report ist ein Lauf-Beleg und ersetzt keine Verifikation (DoD und Spec prüft der Verifier separat).
