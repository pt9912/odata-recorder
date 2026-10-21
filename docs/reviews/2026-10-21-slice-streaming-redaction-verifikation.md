# Verifikation: slice-streaming-redaction

* Datum: 2026-10-21
* Rolle: verifier (frischer Kontext; kein Code, keine Spec, kein Plan geändert)
* Gegenstand: `git diff 401fa86..HEAD`, Stand `ce2240f` (Commits `95156b5` Implementierung, `2618bc7` Review, `ac11962` Architect-Entscheid, `ce2240f` Fänger zu den Findings)
* Plan: `docs/plan/planning/in-progress/slice-streaming-redaction.md` (Welle `welle-record-replay`)
* Maßstab: DoD (§2), `SPEC-118`, `SPEC-046`, `SPEC-115`, Tabelle Zusage·Fänger (§3), Raster (§6), `AGENTS.md` §3.7, §3.9 bis §3.16, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md); Bezug [LH-FA-05](../../spec/lastenheft.md), [LH-FA-14](../../spec/lastenheft.md), [LH-QA-02](../../spec/lastenheft.md)
* Eingang: Review `docs/reviews/2026-10-21-slice-streaming-redaction-review.md` (F-1 bis F-16, I-1 bis I-3), Nachlauf in `ac11962` (Architect) und `ce2240f` (Implementer)
* Belege selbst gefahren: `make test` (Basis grün); 37 Mutationen über `make test` im Docker (netzlos), siehe Abschnitt 4; `make gegenproben` (einmal, Stempel danach mit `git checkout` zurückgesetzt); `git diff --numstat`; `grep`-Läufe über den Diff der Tests; `make docs-check` und `make gates` (Abschnitt 8).

## Urteil

**Freigegeben aus Sicht des Verifiers, mit einer Auflage und einer offenen Entscheidung des Planners.** Code, Spezifikation und Tests stimmen überein; jede Zeile der Tabelle Zusage·Fänger hat einen existierenden Fänger, 37 eigene Mutationen wurden aus dem benannten Grund rot, keine blieb grün. Alle Findings des Reviews sind behandelt. Offen sind (a) ein LOW-Befund V1 (der Stempel der Gegenproben ist nach `ce2240f` veraltet; Zuständigkeit Implementer, mechanisch) und (b) die Größe: 2313 gegen die Marke 2200, Entscheidung Anheben oder Zerlegen beim Planner (Abschnitt 7). DoD-Häkchen und §7 setzt der Planner.

## 1. DoD (§2) gegen Code, Spec und Tests

| DoD-Punkt | Befund |
|---|---|
| `RedigiereBody` nach `SPEC-118`: Typ, strenge JSON-Prüfung, Marker nur für String-Werte unter treffenden Namen, Status `none`/`applied`/`unsafe`, Ausgabe byte-gleich bis auf ersetzte Werte, unabhängig von der Zerlegung | erfüllt: `src/hexagon/model/redaktor.go` (Argumente, Typ, EDMX, Lese- und Schreibschleife), `redaktor_json.go` (Automat); Typ-Fänger `TestBodyTyp` (M20, M32b rot), Marker `TestBodyMarker`, Zerlegung `TestBodyZerlegung` (M10, M11, M34 rot) |
| Fail-closed mit `*UnsicherFehler` (`ODR-E2401`): unbekannter Typ, `multipart/mixed`, `text/*`, ungültiges oder unvollständiges JSON, ungültiges UTF-8, Tiefe über 512, Zahl/Objekt/Array unter treffendem Namen, Überschreiten von `MAX_CAPTURE_BODY`, Fehler von `src` und `dst`; kein geheimes Byte in `dst`; Speicher unabhängig von der Größe | erfüllt: je Grund ein Fall in `bodyFaelle` (`TestBodyFaelleDecken` sichert die Klassen), `TestKeinGeheimnisInAusgabe` über alle Fälle mit Geheimnis, Speicher `TestBodySpeicher` (M12 rot), Puffer `TestBodyPuffer` (M13 bis M15 rot) |
| Sicherheits-Fixtures aus `SPEC-053`, soweit der Redaktor sie fängt; nicht gefangene benannt | erfüllt: `TestBodyFixturesSPEC053` (Fall aus `bodyFaelle` je Fixture); die nicht gefangenen (gzip-Bombe, Content-Encoding-Kette, langsamer Consumer, Queue-Überlauf, Crash zwischen Blob-Rename und DB-Commit, Middleware-Logging) stehen in der DoD mit ihren Trägern `slice-capture-pipeline` und `slice-record-service` |
| `make gates` grün | selbst gefahren, siehe Abschnitt 8 |
| Review durchgeführt, Report liegt vor | erfüllt: `docs/reviews/2026-10-21-slice-streaming-redaction-review.md` (für `95156b5`); die Änderungen danach (`ac11962`, `ce2240f`) prüft dieser Bericht per Mutation; das Häkchen setzt der Planner |
| Doku-Update (öffentlicher Vertrag) | erfüllt: `SPEC-118` samt Zeigern in `SPEC-046`, `SPEC-115` Nr. 6, `LH-FA-14.a` und Änderungsverzeichnis (Zeile 1280); Architect-Entscheid in `ac11962` (Nr. 5, 7, 9, Kopf, *Ungeprüft*) |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen, Sache des Planners (Abschnitt 7) |

Die Schichtenrichtung ist eingehalten: das Modell importiert die Ports nicht (`errVoll` im Test als Stellvertreter für `outbound.ErrZuGross`); `make a-check` im Gate-Lauf grün.

## 2. SPEC-118 gegen Code (alle Nummern)

| Stelle | Zusage | Code und Fänger |
|---|---|---|
| Kopf | `MaxBytes` ab 1 bis `MaxInt64`, kein Überlauf; kein Zustand, keine Goroutine | `leseDaten` (`rest < math.MaxInt64`); `TestBodyGrenze` (`MaxInt64`-Fall), `TestBodyOhneZustand` |
| Nr. 1 | Argumente vor allem anderen | `pruefeBodyArgumente`; `TestBodyArgumente` (M28 rot) |
| Nr. 2 | leere Eingabe `none` vor dem Typ | `lauf` (erster Read); `TestBodyLeer` (Fänger der Zeile; die Ende-Mutationen M35 und M36 machen ihn ebenfalls rot) |
| Nr. 3 | Typ, EDMX mit BOM | `klassifiziereBody`, `edmxPruefer`; `TestBodyTyp` (M20, M32b), `TestBodyEdmx` (M22, M36) |
| Nr. 4 | strenge Grammatik, UTF-8, Tiefe 512, Ende des Stroms | `jsonRedaktor`; `TestBodySyntax`, `TestBodyUTF8` (M25), `TestBodyTiefe` (M17, M18), `TestBodyPraefixe` (M30, M35), Zahlgrammatik (M26b) |
| Nr. 5 | Namen nach Dekodierung, Marker, Wertarten, `applied`/`none` | `passt` über `JSONPfad.Passt` (M21), `nameByte`/`escape`/`hex` (M23, M24), `wertStart` (M02, M02b, M02c, M03, M33), Marker (M04, M05, M27) |
| Nr. 6 | höchstens `MaxBytes + 1` gelesen; Speicher unabhängig von Länge; Namen bis Musterlänge | `leseDaten`, `nameByte`; `TestBodyGrenze` (M16, M29b, M38), `TestBodySpeicher` (M12), `TestBodyNamenLaenge` |
| Nr. 7 | Fehlerform, Gründe, Rangfolge der Ursachen, leere Reads | `UnsicherFehler`, `scheitere`, `lauf`; `TestBodyGrund`, `TestBodyFehlertext` (M19), `TestBodyRangfolge` (M06, M07, M09), `TestBodyEingabeFehler` (M34), `TestBodyAusgabeFehler` (M31b) |
| Nr. 8 | zerlegungsunabhängig, zustandslos, idempotent | `TestBodyZerlegung`, `TestBodyOhneZustand`, `TestBodyIdempotent` |
| Nr. 9 | Präfix bei Fehlern, Marker erst mit dem schließenden Anführungszeichen | `scheitere`, `stringEnde`; `TestBodyTabelle`, `TestBodyPraefixBeiFehler` (M08, M34) |
| Nr. 10 | Pflicht der Aufrufer | kein Fänger möglich; als akzeptiertes Negativ in §3 benannt, Träger tragen die Sendung (Abschnitt 6) |
| *Ungeprüft* | Typed-nil, endloses `0, nil`, Geheimnis unter nicht genannten Namen, Datenrennen, Laufzeit | deckungsgleich mit dem Negativ-Absatz in §3 des Plans |

Eine Abweichung zwischen `SPEC-118` und Code gibt es nicht. Der Widerspruch in Nr. 5 (F-8) ist beseitigt: der Satz nennt `[REDACTED]` ohne Escape als keine Änderung und führt den Inhalt mit einem Escape für die öffnende Klammer, `[REDACTED] ` mit Leerzeichen und den leeren String als Änderung. Nr. 7 und Nr. 9 stimmen jetzt zusammen: der gepufferte Teil geht ein letztes Mal an `dst`, außer bei `ausgabe` (Code: `scheitere`).

## 3. Zusage·Fänger-Tabelle (§3) und Raster (§6)

Jeder in der Tabelle genannte Fänger existiert unter diesem Namen (`grep ^func Test` über `redaktor_test.go`): `TestBodyArgumente`, `TestBodyLeer`, `TestBodyTyp`, `TestBodyEdmx`, `TestBodyMarker`, `TestBodyNamen`, `TestBodyNamenEscapes`, `TestBodyDefaults`, `TestBodyMitglieder`, `TestBodyWertarten`, `TestBodyZahlAnfang`, `TestBodyMarkerIdempotent`, `TestBodyStatus`, `TestBodySyntax`, `TestBodyUTF8`, `TestBodyTiefe`, `TestBodyPraefixe`, `TestBodyZerlegung`, `TestBodyGrenze`, `TestBodySpeicher`, `TestBodyNamenLaenge`, `TestBodyPuffer`, `TestKeinGeheimnisInAusgabe`, `TestBodyFaelleDecken`, `TestBodyGrund`, `TestBodyTabelle`, `TestBodyReihenfolge`, `TestBodyFehlertext`, `TestBodyEingabeFehler`, `TestBodyAusgabeFehler`, `TestBodyAusgabeVoll`, `TestBodyOhneZustand`, `TestBodyIdempotent`, `TestBodyRangfolge`, `TestBodyPraefixBeiFehler`, `TestBodyFixturesSPEC053`. Keine Zeile nennt einen Namen ohne Test.

Die Fänger trennen die Mutationen mit den Eingaben, die die Zeile nennt; das zeigt Abschnitt 4 (jede Mutation ist an einer Zeile der Tabelle ausgerichtet und wird von dem dort genannten Test rot). Die Fixtures stehen in der Tabelle `bodyFaelle` und in den Testfunktionen mit Erwartungen von Hand, der Marker als Text `[REDACTED]`.

Raster §6: keine leere Zelle; jede nennt eine Nummer von `SPEC-118` oder „gilt nicht“ mit Grund. Die Zelle `MaxBytes × fremd` (`math.MaxInt64`) nennt als Ort den Kopf von `SPEC-118`, der den Wert nennt (F-11 erledigt). Die Zusatzzeilen *Fehler von `dst` × Fehler von `src`* und *Fehler von `dst` × Verstoß* haben ihren Ort in Nr. 7 *Rangfolge der Ursachen* (F-6 erledigt) und ihren Fänger in `TestBodyRangfolge` (M06, M07, M09 rot). *Typed-nil* und *endloses `0, nil`* verweisen auf *Ungeprüft*.

## 4. Mutationen (selbst gefahren)

**Verfahren.** Jede Mutation ersetzt eine Stelle in `src/hexagon/model/` mit `perl`, `make test` (Dockerfile, Stufe `test`, netzlos) läuft, danach wird die Datei aus dem Repo zurückkopiert. Die Läufe liefen in vier Kopien des Stands `HEAD` (`git archive`, außerhalb des Repos, im Scratchpad), damit sie parallel laufen; der Baum des Repos blieb unberührt (`git status` leer, `diff -rq src` zu jeder Kopie ohne Unterschied). Gezählt werden nur Mutationen, die kompilieren, `gofmt` und `go vet` bestehen und nicht hängen. „rot“ nennt die Tests, die aus dem benannten Grund fehlschlugen; keine Mutation wurde wegen Aufbaubruchs als rot gewertet.

| # | Zusage | Mutation | Ergebnis |
|---|---|---|---|
| M02 | Zahl unter treffendem Namen `nichtstring`, Anfang `2` | `2` aus `"{[-0123456789"` entfernt | rot: `TestBodyZahlAnfang` |
| M02b | dito, Anfang `5` | `5` entfernt | rot: `TestBodyZahlAnfang` |
| M02c | dito, Anfang `-` | `-` entfernt | rot: `TestBodyWertarten`, `TestBodySyntax`, `TestBodyZahlAnfang` |
| M03 | dito, Anfang `9` | `9` entfernt | rot: `TestBodyZahlAnfang` |
| M33 | Objekt unter treffendem Namen `nichtstring` | `{` aus der Klasse entfernt | rot: `TestBodyWertarten`, `TestBodyStatus`, `TestBodyZerlegung`, `TestKeinGeheimnisInAusgabe`, `TestBodyGrund` |
| M04 | Marker-Vergleich Byte für Byte | `&& c == Marker[s.mPos]` entfernt (nur Länge zählt) | rot: `TestBodyMarkerIdempotent`, `TestBodyStatus`, `TestBodyZerlegung`, `TestBodyGrund` |
| M05 | Marker-Vergleich nicht als Präfix | `s.mAbw \|\| s.mPos != len(Marker)` zu `s.mAbw` | rot: `TestBodyMarker`, `TestBodyMarkerIdempotent`, `TestBodyStatus`, `TestBodyZerlegung` |
| M27 | `applied` genau bei geändertem String | `geaendert` immer gesetzt | rot: `TestBodyMarkerIdempotent`, `TestBodyStatus`, `TestBodyZerlegung`, `TestBodyGrund` |
| M06 | Rangfolge Nr. 7: Verstoß schlägt Fehler von `dst` beim Leeren | `scheitere` meldet den Fehler des Leerens | rot: `TestBodyRangfolge` |
| M07 | Rangfolge: `zu_gross` vor dem Leeren | Leeren vor die Prüfung `zuGross` gesetzt | rot: `TestBodyRangfolge` |
| M09 | Rangfolge: Daten und Leeren vor Lesefehler | Auswertung von `rerr` vor das Leeren gesetzt | rot: `TestBodyRangfolge` |
| M38 | Rangfolge: Verstoß vor der Grenze gewinnt gegen `zu_gross` | `f != nil && !zuGross` | rot: `TestBodyGrenze` |
| M08 | Nr. 9: gepufferter Teil geht nach dem Verstoß an `dst` | Leeren in `scheitere` abgeschaltet (`false &&`) | rot: `TestBodyEdmx`, `TestBodyMitglieder`, `TestBodyWertarten`, `TestBodySyntax` |
| M10 | Zerlegung: Hex-Ziffernzähler des `u`-Escapes überlebt die Chunk-Grenze | `hexN` am Anfang von `schreibe` zurückgesetzt | rot: `TestBodyZerlegung` |
| M11 | Zerlegung: Marker-Vergleichsstand überlebt die Chunk-Grenze | `mPos` am Anfang von `schreibe` zurückgesetzt | rot: `TestBodyZerlegung` |
| M34 | Daten mit Fehler im selben Read werden verarbeitet | `n = 0`, wenn `err != nil` | rot: `TestBodyZerlegung`, `TestBodyGrenze`, `TestBodyEingabeFehler`, `TestBodyRangfolge` |
| M12 | Speicher unabhängig von der Länge eines Namens | Kappung des Namens (`ueberlang`) abgeschaltet | rot: `TestBodySpeicher` (9-MiB-Name) |
| M13 | Eingabepuffer 32 KiB | 128 KiB | rot: `TestBodyPuffer` |
| M14 | Ausgabeschwelle 32 KiB | 128 KiB | rot: `TestBodyPuffer`, `TestBodyAusgabeVoll`, `TestBodyRangfolge` |
| M15 | Füllstand-Flush | `voll()` immer falsch | rot: `TestBodyPuffer`, `TestBodyAusgabeVoll`, `TestBodyRangfolge` |
| M16 | genau `MaxBytes` zulässig | `>` zu `>=` | rot: `TestBodyGrenze` |
| M29b | höchstens `MaxBytes + 1` Bytes gelesen | `rest++` zu `rest += 2` | rot: `TestBodyGrenze` (Zähler) |
| M17 | Tiefe 512 zulässig | Grenze 511 | rot: `TestBodyStatus`, `TestBodyTiefe`, `TestBodyPraefixe`, `TestBodyGrenze` |
| M18 | Tiefe 513 `tiefe` | Grenze 513 | rot: `TestBodyTiefe`, `TestBodyGrenze`, `TestBodyGrund`, `TestBodyTabelle` |
| M19 | Fehlertext nennt Code und Grund | `Error()` liefert `fehler` | rot: `TestBodyFehlertext` |
| M20 | Typ `+json` | Suffix `json` | rot: `TestBodyTyp` |
| M32b | Rand des Media-Types nur Leerzeichen und Tabulator | `strings.TrimSpace` | rot: `TestBodyTyp` |
| M21 | Namensvergleich allein über `JSONPfad.Passt` | `strings.EqualFold` | rot: `TestBodyNamen` (Kelvin) |
| M22 | EDMX: zweites BOM-Byte geprüft | `return true` | rot: `TestBodyEdmx`, `TestBodyStatus`, `TestBodyZerlegung`, `TestBodyGrund` |
| M36 | EDMX: Ende vor `<` ist `mismatch` | Prüfung am Ende abgeschaltet | rot: `TestBodyLeer`, `TestBodyEdmx`, `TestBodyStatus`, `TestBodyZerlegung` |
| M23 | Das `u`-Escape ist der ganze Codepunkt | nur das niederwertige Byte | rot: `TestBodyNamenEscapes` |
| M24 | kurze Escapes sind Steuerzeichen | Buchstabe hinter dem Backslash | rot: `TestBodyNamenEscapes` |
| M25 | UTF-8: Folgebytes bis `BF` | Obergrenze `FF` | rot: `TestBodyUTF8` |
| M26b | Zahl: auf führende `0` keine Ziffer | Prüfung entfernt | rot: `TestBodyStatus`, `TestBodySyntax`, `TestBodyZerlegung`, `TestBodyGrund`, `TestBodyTabelle` |
| M28 | `MaxBytes < 1` ist `ErrUngueltig` | `< 0` | rot: `TestBodyArgumente`, `TestBodyReihenfolge` |
| M30 | Ende bei `-` ist `unvollstaendig` | `jzMinus` als gültiges Ende | rot: `TestBodyPraefixe` |
| M31b | Kurzschreiben ist `ausgabe` | `io.ErrShortWrite` nie gesetzt | rot: `TestBodyAusgabeFehler` |
| M35 | Ende des Stroms vor dem Wert ist `unvollstaendig` | `ende()` liefert `nil` | rot: `TestBodyLeer`, `TestBodyStatus`, `TestBodyPraefixe`, `TestBodyZerlegung` |

Zu den geforderten Schwerpunkten: **Zahl-Anfangsziffern** M02, M02b, M02c, M03, M33 (drei Ziffern, `-` und `{`); **Marker-Vergleich** M04, M05, M27; **Rangfolge Nr. 7** M06, M07, M09, M38; **Präfix Nr. 9** M08 (zusätzlich M34); **Zerlegungs-Unabhängigkeit** M10, M11, M34; **Speicher** M12 bis M15.

**Nicht gezählt** (gemeldet, damit die Zahl nachvollziehbar ist): ein Versuch ohne `rest++` (`M29`) hängt (Exit 124 nach 550 s, ein `Read` mit leerem Puffer liefert `0, nil`) und zählt nicht; der Ersatz M29b ist rot. `M31` (Kurzschreiben, erste Fassung) brach mit „declared and not used“ ab, `M26` (erste Fassung) an `gofmt`; beide wurden in einer Fassung wiederholt, die baut (M31b, M26b). Zwei Versuche änderten nichts (Musterfehler: `M32` erste Fassung, `M37` Klammerart bei leeren Containern) und zählen nicht. **Keine Mutation blieb grün.** Die Klammerart leerer Container (`[}`, `{]`) hat Fänger (die Fälle stehen in `TestBodySyntax`, `grep`), aber der Verifier hat dazu keine kompilierende Mutation rot gesehen; der Review hat sie gesehen.

## 5. Review-Findings F-1 bis F-16, I-1 bis I-3

| Finding | Behandlung | Beleg |
|---|---|---|
| F-1 MEDIUM Marker gleicher Länge | Fälle `S3CR3T-7fX` und `[REDACTEX]` (10 Bytes) in `TestBodyMarkerIdempotent` | M04 rot |
| F-2 MEDIUM Zahl-Anfänge | `TestBodyZahlAnfang` für `0` bis `9` und `-1`, einstellig und mit Ziffern, unter `password` und unter `p` | M02, M02b, M02c, M03 rot |
| F-3 MEDIUM Syntax-Allaussage | Fälle `0.5`, `-0.5`, `0e5`, `1e-5`, `-01`, `1.5.`, `1e5e5`, `[}`, `{]`, `C2 C0`, `C2 FF`, `E1 80 C0`, `F1 80 80 C0` in `TestBodySyntax`/`TestBodyUTF8` | M25, M26b rot; Klammerart: Fall vorhanden, Mutation nicht vom Verifier gesehen |
| F-4 MEDIUM Verstoß vor der Grenze | Fälle „… vor der Grenze, Strom darüber“ in `TestBodyGrenze` (Syntax, `utf8`, `nichtstring`, `mismatch`, Tiefe) | M38 rot |
| F-5 MEDIUM Speicher-Zusage | `TestBodyPuffer`, `TestBodyAusgabeVoll`; Kommentar und `SPEC-118` Nr. 6 sagen „heute“ und „die Zusage ist die Unabhängigkeit von der Länge“; Rest benannt (Name bis Musterlänge plus 64 KiB; `schreibeVoll` im EDMX-Prüfer äquivalent) | M12 bis M15 rot |
| F-6 MEDIUM Daten mit Lesefehler und `dst` | Architect: Nr. 7 *Rangfolge der Ursachen*; `TestBodyRangfolge` | M06, M07, M09 rot |
| F-7 MEDIUM Leeren nach dem Verstoß | Architect: Nr. 7 („Einzig der schon redigierte, noch gepufferte Teil …“) und Nr. 9 | M08 rot |
| F-8 MEDIUM Widerspruch in Nr. 5 | Architect: Satz berichtigt; Code und Tabelle folgen ihm | M05, M27 rot |
| F-9 MEDIUM Escapes im Namen | `TestBodyNamenEscapes` mit Mustern, die den falsch dekodierten Namen treffen | M23, M24 rot |
| F-10 LOW Plan-Zeile gegen Spec | Plan-Zeile *JSON streng* trägt „erstes Byte“; `SPEC-118` Nr. 4 letzter Punkt sagt es ausdrücklich | `grep` im Plan und in Nr. 4 |
| F-11 LOW Ort `MaxInt64` | Kopf von `SPEC-118` nennt den Wert; Raster-Zelle verweist dorthin | Kopf von `SPEC-118` |
| F-12 LOW `ErrZuGross` ohne Nehmer | `slice-capture-pipeline` §1 trägt den Beweis mit dem Echtwert „(Sendung von `slice-streaming-redaction`)“ | Abschnitt 6 |
| F-13 LOW BOM und Rand | `TestBodyEdmx` (`EF 41 BF 3C`), `TestBodyTyp` (CR, LF, FF, VT, NBSP am Rand) | M22, M32b rot |
| F-14 LOW Fehlertext | `TestBodyFehlertext` prüft Code und Grund | M19 rot |
| F-15 LOW Ränder des Aufrufs | `SPEC-118` *Ungeprüft* und Absatz „Nicht gefangen“ in §3 des Plans nennen Typed-nil und endloses `0, nil` | Abschnitt 9 |
| F-16 LOW Kommentar `bodyAusReserve` | berichtigt: „der Ausgabepuffer hat diesen Spielraum über der Schwelle, weil er erst nach dem Byte geleert wird, das sie erreicht“ | `redaktor.go` Zeile 80 bis 82 |
| I-1 INFO äquivalente Mutationen | keine Aktion | — |
| I-2 INFO Name über Musterlänge und Chunk-Grenze | Fall `name-ueber-musterlaenge` (20 Zeichen vor `password`) steht in `bodyFaelle` und in der Pflichtliste von `TestBodyZerlegung` | `redaktor_test.go` Zeile 374 und 851 |
| I-3 INFO Kommentar in `TestBodyGrenze` | „… wäre kein int64“ im Test, Klasse *Grenze*; keine Aktion | Zeile 899 |

## 6. Übergaben (§3.13) und Plan-vs-Code (§3.9)

| Nehmer | Sendung | Befund |
|---|---|---|
| `slice-capture-pipeline` | Aufruf der Funktion, `Metadata`, `dst` bei Fehler verwerfen, Zustände nach Art des Fehlers, Strom entpackt, redigierte Request-Bytes für den Key, Beweis mit `outbound.ErrZuGross` | §1 „Übernimmt (Redaktor)“ trägt alles mit „(Sendung von `slice-streaming-redaction`)“; Ausschluss „Der Redaktor selbst — `slice-streaming-redaction`“ trifft die Sendung nicht (die Pipeline ruft ihn) |
| `slice-record-service` | Request-Body durch den Redaktor, Key aus der Ausgabe, `redaction.request=unsafe` mit `null`-Key | §1 „Übernimmt (Redaktor)“ mit Kennung des Gebers; Ausschlüsse (HTTP-Server, Sessions, Link-Umschreiben, Fallback) treffen die Sendung nicht |
| `slice-replay-service` | Request-Body mit `MAX_CAPTURE_BODY`, Key aus der Ausgabe, `ODR-E5001` ohne Cursor-Fortschritt, Request ohne Body kein Fehler | §1 „Übernimmt (Redaktor)“ mit Kennung des Gebers; Ausschlüsse treffen die Sendung nicht |

Alle drei Nehmer liegen in `open/`, keiner in `done/`.

Plan-vs-Code (§3.9): die Dateitabelle in §3 nennt `redaktor.go`, `redaktor_json.go`, `redaktor_test.go` und `spec/spezifikation.md`; der Diff `401fa86..HEAD` berührt genau diese, dazu den Plan, `slice-capture-pipeline` (Übergabe F-12), den Bericht des Reviews und `harness/gegenproben.stempel`. Die Zahlen im Plan stimmen mit dem Diff überein (Tests 1187 zu 1449, also +262; Produktcode 340 zu 344, also +4 Zeilen Kommentar; die Zeilen *Rangfolge*, *Fehlerform*, *Speicher* nennen die Mutationen, die Abschnitt 4 rot sieht). Der Kopf nennt `Berührte Spec-Stellen`, §1 die drei Übergaben. Eine Abweichung zwischen Plan und Code fand sich nicht; zu `harness/gegenproben.stempel` siehe V1.

## 7. Größe (§4) und Closure-Pflichten

Messung nachgestellt: `git diff --numstat 401fa86..HEAD` über `src/hexagon/model/` und `spec/spezifikation.md` (die Pfade aus §3 des Plans; Tests eingeschlossen):

| Pfad | Einfügungen | Löschungen |
|---|---|---|
| `src/hexagon/model/redaktor.go` | 344 | 0 |
| `src/hexagon/model/redaktor_json.go` | 504 | 0 |
| `src/hexagon/model/redaktor_test.go` | 1449 | 0 |
| `spec/spezifikation.md` | 9 | 7 |

Summe aus Einfügungen und Löschungen: **2313** (Produktcode 848, Tests 1449, Spezifikation 16, davon der Architect-Entscheid `ac11962` mit 9 und 7). Ohne die Spezifikation: 2297. Die Zahl entspricht der im Plan genannten (§4, §6). Die Marke 2200 ist um **113** überschritten (mit Spezifikation; ohne sie um 97); die Schätzung (rund 1500) um 813. Beim Review war der Stand 2031 (Basis `401fa86..95156b5`, bestätigt); der Zuwachs von 282 sind +262 Tests, +4 Produktcode und +16 Spezifikation. Der Satz in §4 des Plans stellt die Überschreitung fest („die Marke ist überschritten“); der Ausgang des Risikos *Größe* in §6 steht „offen“. Ob die Marke angehoben wird oder der Slice nach `next` zurückgeht, entscheidet der Planner; dieser Bericht trifft keine Entscheidung.

Für die Closure bleibt dem Planner (nicht dem Verifier):

* DoD-Häkchen in §2 und die Closure-Notiz §7 mit Steering-Loop-Lerneintrag (Finding-Klassen des Reviews: „Gleichwert-Fall fehlt im Fänger“, „Klasse nur zum Teil durch Fixtures belegt“, „Allaussage ohne Fänger“, „Kommentar sagt mehr zu als der Test fängt“, „Randform erst im Code entschieden“).
* Die Ausgänge in §6, die „offen“ stehen: Speichertest (Messung über `TotalAlloc`), Raster, Übergaben (weiter offen bis zur Closure der drei Nehmer ist ein zulässiger Ausgang), Größe.
* Beobachtungs-Register fortschreiben oder „keine Beobachtung“ notieren; die Register-Einträge `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `BEO-REPO/randform-erst-im-review-entschieden` und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` sind Kandidaten für die Zuordnung der Findings F-1 bis F-9.
* Der Review gilt für `95156b5`; die Änderungen danach (`ce2240f`) sind Tests und vier Zeilen Kommentar im Produktcode, von Abschnitt 4 per Mutation geprüft. Ob ein Re-Review nötig ist, entscheidet der Planner.

## 8. Gate-Lauf

`make test` lief vor jeder Mutation grün. `make docs-check`: 126 Dateien (mit diesem Bericht), 0 Befunde. `make gates` nach dem Commit dieses Berichts: Ergebnis im Handback an den Planner. `make gegenproben` (kein Gate) lief einmal grün, siehe V1.

## 9. Non-ASCII im Diff der Tests, akzeptierte Negative

**Non-ASCII.** Die Datei `redaktor_test.go` enthält 134 Zeichen außerhalb von ASCII. Die Aufstellung (`grep -o`): `ä` 39, `ü` 33, `ö` 19, `ß` 17 (deutsche Kommentare und Meldungen), `é` 9, `€` 8, `😀` 8 (die Fixtures `mehrbyte-ueberall`, `unicode-escape-im-wert`, `TestBodyUTF8`), eine Ellipse in `kurz()`. Es gibt im Quelltext kein NBSP, kein Zeichen der Klasse U+2000 bis U+206F, kein BOM und kein Kelvin-Zeichen **als Rohbyte** (`LC_ALL=C grep -P` auf die Bytefolgen, ohne Treffer): Fixtures mit solchen Zeichen stehen als Escape-Folge, und die Escape-Folgen sind das, was sie sagen. Dieser Bericht beschreibt sie mit Worten (Backslash, `u`, vier Hex-Ziffern), weil die Übertragung seines eigenen Entwurfs eine solche Folge im Eingabetext zum Zeichen umgesetzt hat (aus dem Escape für die öffnende Klammer wurde im ersten Entwurf eine öffnende Klammer, aufgefallen beim Gegenlesen und berichtigt): dieselbe Umsetzung, die der Implementer gemeldet hat, wirkt in der Übertragung von Werkzeug-Eingaben, nicht im Quelltext, den `grep` auf Bytes belegt.

* In **Rohstrings** (Backtick) steht der Escape als Text: Backslash, `u` und vier Hex-Ziffern (etwa `0077`, `0050`, `212A`, `d800`). Der Redaktor sieht den Backslash und dekodiert. Beispiele der Tabelle: `pass` mit `0077` für das `w`, `0050` für das große `P`, `api_` mit `006b` und `0065` (und in Großschreibung `004B`, `0045`) für `k` und `e`, `0170` und `7000` am Anfang von `assword` (Mutation M23: nur das niederwertige Byte würde `p` ergeben), das einzelne Surrogat `d800`.
* In **Strings mit Anführungszeichen** dekodiert Go: der Fall `api_` + Escape `212A` + `ey` ist `api_` + U+212A (Kelvin) + `ey` als Rohbytes `E2 84 AA` (der Fall „Kelvin roh in UTF-8“ der Plan-Zeile), `passw` + Escape `00f6` + `rd` ist `passwörd` roh, und die Fälle mit Escape `00a0` am Rand von `application/json` tragen ein echtes NBSP. Der Fall in `bodyFaelle` (Zeile 344) schreibt über doppelte Backslashes den Escape-Text (Wert `é` als Escape `00e9`, danach `\/`).
* Das Paar steht nebeneinander in Zeile 588 von `redaktor_test.go`: der Kelvin-Fall einmal in Anführungszeichen (roh) und einmal im Rohstring (als Escape), beide auf der richtigen Seite der Anführung; eine Umsetzung durch ein Heredoc (Escape zu Zeichen oder umgekehrt) hat sich im Quelltext nicht niedergeschlagen. Die Mutation M21 (`strings.EqualFold` statt `Passt`) macht `TestBodyNamen` rot und bestätigt, dass die Kelvin-Fälle als Kelvin-Zeichen ankommen (`EqualFold` faltet es auf `k`); M23 und M24 bestätigen die Escape-Fälle von `TestBodyNamenEscapes`.

**Akzeptierte Negative.** Richtig benannt: Typed-nil als `dst` oder `src` und ein `src`, der endlos `0, nil` liefert (`SPEC-118` *Ungeprüft* und Absatz in §3: kein Fänger, ein Aufruf damit ist kein Fall dieses Slice); das Rennen innerhalb gleichzeitiger Aufrufe (kein Paketzustand im Code, nur Konstanten und `Pfade` lesend; das Gate fährt nicht mit dem Race-Detektor; das Register `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` trägt den Fall); die Pflicht der Aufrufer (Nr. 10, die Nehmer tragen sie). Der Verifier fand kein weiteres.

## Befunde

| ID | Kategorie | Befund | Zuständig |
|---|---|---|---|
| V1 | LOW | `harness/gegenproben.stempel` ist nach `ce2240f` veraltet: `src` ist Eingabe der Prüfsumme (`tools/harness/gegenproben-eingabe.sh`), `ce2240f` änderte `redaktor.go` und `redaktor_test.go`, den Stempel aber zuletzt `95156b5`. `make gegenproben` läuft deshalb neu (selbst gefahren, 67 s, grün) und schreibt `21ef739b7718efcdbbc3658c60647e767e37056627ad5f03f32ba9a60673ddb9`; der Verifier hat die Datei danach mit `git checkout` zurückgesetzt und nicht committet. Die Gegenproben sind kein Gate, `make gates` sieht den Unterschied nicht. | Implementer: `make gegenproben` fahren, Stempel committen (wie in `c9c20ae` und `95156b5`) |
| V2 | INFO | Größe 2313 gegen die Marke 2200 (Abschnitt 7); keine Entscheidung des Verifiers. | Planner |

## Zuständigkeiten

* Implementer: V1 (mechanisch, ein Zeilen-Commit).
* Planner: V2 (Anheben oder Zerlegen), Closure-Pflichten aus Abschnitt 7, Entscheidung über ein Re-Review, Zuordnung der Finding-Klassen.
* Architect: keine Spec-Lücke gefunden; keine Aktion.
