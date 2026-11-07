# Review-Bericht: slice-v2-payload-schreiben — 2026-11-07

**Rolle:** Reviewer (Modul 10) · **Frage:** Ist der Diff gegen Plan, Entscheidungen und Hard Rules wartbar und richtig? — nicht gegen die DoD (Verifier).

**Review-Art:** Code (gegen Plan und Konventionen).

**Gegenstand:** `slice-v2-payload-schreiben`, Zweig `slice-v2-payload-schreiben`, HEAD `55d3a51`, Diff `5f235be..55d3a51` (14 Dateien, 3336 Einfügungen, 13 Löschungen; Pfade des Blocks `groessenmarke`: 3229 Zeilen bei Marke 5900).

**Geprüft gegen:** Plan `slice-v2-payload-schreiben` (§3 Tabelle 43 Zeilen, Mengen mit Achsen, Pflicht-Mengen, Raten, §6 Raster), Vertrag `SPEC-172` Nr. 6 a bis o und Nr. 7, `SPEC-168`, `AGENTS.md` §3.7, §3.9 bis §3.11, §3.15 bis §3.18. Skill: `.harness/skills/reviewer.md`. Modell: Sonnet 5.5.

**Methode.** Einmal gefahren (keine Wiederholungsschleife). Der Bericht des Implementers ist nicht übernommen. Eigene Mutationen nur über `tools/harness/mutationslauf.sh`, unter `flock /tmp/odata-recorder-heavy.lock`, `--memory 2g`, `--cpus=4`; Last des Hosts 1,7 bis 28 (andere Läufe auf demselben Host). 32 Mutationen gelaufen (30 auswertbar, 2 `uebersetzungsfehler` durch ungenutzten Import, nicht gezählt): 25 der Klasse `andere`, 5 `nebenläufig` mit `n_detektor` 20 und `n_ohne` 20 bzw. 0, je Prozess `-test.count=1`. Die Mutationen sind eigene (r*, s*, c*, i*), nicht die der Spalte *Mutation* der Plan-Tabelle.

**Grenze dieses Laufs (ehrlich).** Die Mutationen der Spalte *Mutation* der 39 Zeilen habe ich nur dort gefahren, wo sie mit einer eigenen zusammenfallen (Zeilen 15, 22, 30, 31 Teile); nicht aufgelöst sind die übrigen Zeilen der Spalte. `make a-check` fährt diese Rolle nicht (nur der Mutationslauf ist erlaubt); Aussagen zu `a-check` sind deshalb am Konfigurationstext abgeleitet, nicht gemessen. Die Last des Hosts schwankt; Raten gelten als Größenordnung.

---

## Findings

### F-1 (MEDIUM): Zusage „der Wert des Ports wird nicht geprüft" hat keinen Fänger

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe; Zusage ohne Fänger), §3.17; `SPEC-172` Nr. 6 h (Wurzel: „ein bis fünf Ziffern (der Wert wird nicht geprüft)")
- `pfad`: `src/adapters/driven/odata/jsoncodecv2/schreiber_adresse.go` · `if i == 0 || i > 5 {`; Fänger `schreiber_adresse_test.go` · `"http://h:65535"`
- `befund`: Der einzige gültige Port mit fünf Ziffern in Zeile 25 ist `65535`; die Mutation „fünfstelliger Port mit erster Ziffer über 6 abgelehnt" (`i == 5 && p[0] > '6'`, also `http://h:99999` abgelehnt) bleibt in allen Tests grün (s05, `gruen`, 2 Prozesse). Die Zusage „Wert nicht geprüft" (Werte 65536 bis 99999, auch `00000`) hat weder Fall noch Negativ.
- `verifizierbar`: ja (Mutation s05, `sh tools/harness/mutationslauf.sh`)
- `klasse`: Fänger führt nur eine Form der Eingabe (Grenze der Zusage nicht beidseitig)

### F-2 (MEDIUM): Zeile 39 „keinen Adapter" nennt als Fänger `make a-check`, die Konfiguration trägt ihn nicht erkennbar; Negativ fehlt

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Zusage ohne Fänger; akzeptiertes Negativ ohne Grund und Folge); `SPEC-172` Nr. 6 o („importiert … keinen Adapter … kein Paket der Version 4"); Plan §3 Zeile 39
- `pfad`: Plan · `ein Import von `jsoncodec` (a-check und `versionstrennung_test.go` rot)`; `.a-check.yml` · `adapters-driven:`
- `befund`: Ein Import von `jsoncodec` wird von `TestKeinPaketImportiertEineAndereVersion` rot (i01, 1 Prozess). Ein Import des Adapters `sqlite` bleibt dort grün (i02, `gruen`); `.a-check.yml` fasst `src/adapters/driven/**` zu einer Schicht ohne Kante driven zu driven zusammen, ob `a-check` diesen Import meldet, belegt der Plan nicht (keine Mutation für `a-check`, kein Negativ mit Grund und Folge).
- `verifizierbar`: ja (`make a-check` auf einer Wegwerfkopie mit der Mutation i02; von mir nicht gefahren)
- `klasse`: Fänger benannt, Wirkung nicht belegt (Gate-Fänger ohne Mutation)

### F-3 (LOW): Der Zeitpunkt wird in der Zone des Aufrufers nicht abgesichert (Umgebung des Tests ist UTC)

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15 (Fänger schmaler als Zusage); `SPEC-172` Nr. 6 f, k (Millisekunden seit 1970 aus einem Text mit `Z`)
- `pfad`: `schreiber_kanon.go` · `ms*int(time.Millisecond), time.UTC).UnixMilli()`
- `befund`: Die Mutation `time.UTC` zu `time.Local` bleibt grün (r13, `gruen`), weil der Container in UTC läuft; auf einem Host mit anderer Zone lieferte sie um den Versatz falsche Millisekunden. Die Menge der Zeile 15 (31 Werte) hat keine Achse „Zone der Maschine".
- `verifizierbar`: ja (Mutation r13)
- `klasse`: Fänger hängt an der Umgebung des Laufs

### F-4 (LOW): Kommentar von `TestGleichzeitig` nennt eine Ausnahme, die der Code nicht macht

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.11 (Ein Kommentar sagt nur zu, was ein Test prüft), §3.7
- `pfad`: `schreiber_format_test.go` · `dann jede Aufrufart allein in 30 Läufen (ohne die Zeitpunkte, die fmt und andere Pools zwischen den Goroutinen ordnen)`
- `befund`: Die Schleife `for _, i := range alle` führt alle 11 Aufrufarten einzeln, auch die mit Zeitpunkt (Art 4 mit `Edm.DateTimeOffset`, Art 9); der Klammersatz liest sich als Ausschluss. Gemeint ist offenbar die Wirkung von `fmt` auf die Ordnung im Detektor; der Satz sagt es nicht.
- `verifizierbar`: nein
- `klasse`: Kommentar nennt mehr oder anderes als der Test

### F-5 (INFO): Redundante Prüfungen (äquivalente Mutationen), kein Fänger fehlt

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `schreiber_kanon.go` · `t.Year() == f[0] && int(t.Month()) == f[1] && t.Day() == f[2]`; `schreiber_bau.go` · `if _, doppelt := ts.idx[p.Name]; !doppelt {`; `schreiber_kanon.go` · `return "Edm.String", utf8.ValidString, true`
- `befund`: Jede der drei Kalenderprüfungen ist für sich überflüssig, weil die beiden anderen die Normalisierung von `time.Date` fangen (r07, r08, r09 je `gruen`). Die Auswahl „erstes Vorkommen" im Verzeichnis der Properties ist unerreichbar, wenn das Modell Namen eindeutig führt (r06 `gruen`); die UTF-8-Prüfung der Art `string` wiederholt die im Escape (s02 `gruen`). Verteidigung in der Tiefe, kein Fehler; die Mutationen sind nicht rot zu machen.
- `verifizierbar`: ja (Mutationen r06 bis r09, s02)
- `klasse`: Äquivalente Mutation durch doppelte Prüfung

### F-6 (INFO): `TestGleichzeitig` erreicht datenabhängige Zweige nur für die Eingaben von `aufruf`

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.10
- `pfad`: Plan §3 · `führt jede Anweisung des Produktcodes aus`; `schreiber_format_test.go` · `model.AlleCodes()[g%len(model.AlleCodes())]`
- `befund`: Mit 16 Goroutinen erreicht die Aufrufart 3 nur die ersten 16 der 36 Codes; ein Zähler in `SchreibeFehler` für `ODR-E9001` bleibt unerreicht (c05, `gruen`, 0 von 20 mit Detektor). Die Aussage des Plans gilt für Anweisungen, nicht für Zweige nach Wert. Ein Zähler in jedem Zweig, den die Aufrufe ausführen, wird gefangen (c01 bis c04, siehe unten).
- `verifizierbar`: ja (Mutation c05)
- `klasse`: Abdeckung nach Anweisung statt nach Wert

---

## Aufgelöste Fänger (eigene Mutationen, ohne Befund)

| Zeile | Mutation (nicht in der Tabelle des Plans) | Ergebnis |
|---|---|---|
| 22 | r01 Name der Kopfzeile `DataserviceVersion`; r02 `text/plain;charset=UTF-8` | rot (`TestDarstellungskopf`) |
| 18 | r03 `message` statt Gesamtobjekt verglichen | rot |
| 16 | r04 Auswahl in `SchreibeCollection` nicht geprüft; r05 in `SchreibeEntity` nicht geprüft | rot (`TestAuswahl`) |
| 14 | r10 Länge 25 zugelassen; r15 Trennzeichen der Minute nicht geprüft | rot |
| 12 | r11 `-0` als `Edm.Decimal` angenommen | rot |
| 13 | r14 Bindestrich an Stelle 18 nicht geprüft | rot |
| 23 | r16 Hex-Ziffer des unteren Halbbytes klein | rot |
| 30 | r17 U+007F escaped | rot |
| 31 | r18 Präfix `[REDACTED]` statt gleichem Text; r19 Zeichen hinter dem Marker | rot |
| 26 | r20 Hex-Zeichen hinter `%` eingeschränkt; s01 letztes Zeichen eines IPv6-Hosts nicht geprüft | rot |
| 7 | s03 mittleres Element der Collection ohne Fehler | rot |
| 39 | i01 Import von `jsoncodec` | rot (`TestKeinPaketImportiertEineAndereVersion`) |

**Nebenläufigkeit (Zeilen 37, 38).** Eigene Mutationen mit Wirkung auf geteilten Zustand, Image `odata-recorder:race`, 4 CPUs, Messender Reviewer:

| Mutation | mit Detektor | ohne Detektor | Last |
|---|---|---|---|
| c01 Map des Pakets in `adresse` (ein Aufruf schreibt) | 20/20 | 20/20 | 3,76 |
| c02 Zähler im Zweig `io.ErrShortWrite` von `ausgeben` | 20/20 | 0/0 | 3,76 |
| c03 Zähler im Zweig des beendeten Kontexts | 20/20 | 0/0 | 2,72 |
| c04 gemeinsamer Puffer in `ergebnis` (wie w02, anderer Ort) | 20/20 | 20/20 | 2,33 |
| c05 Zähler im Zweig `ODR-E9001` von `SchreibeFehler` | 0/20 | 0/0 | 1,70 |

c02 und c03 sind Zweige, die der Fake von `dst` und der abgebrochene Kontext in `aufruf` erreichen; der Detektor fängt sie in jedem Prozess. Hänger wie bei w02 des Plans (7 von 50 ohne Detektor) traten in c04 mit 20 Prozessen nicht auf; die Rate ohne Detektor ist mit 20/20 und der Yield im Ziel im Einklang mit der „Lehre der Runde". Zahl der Prozesse ist 20 statt 50, daher Größenordnung.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/odata/jsoncodecv2/schreiber*.go` (Produktcode, 736 Zeilen) | geprüft; Befunde F-1, F-3 bis F-6; keine Funktion ungenutzt |
| Rohbytes (`AGENTS.md` §3.18) | geprüft, ohne Befund: `LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}'` über `schreiber*.go` leer; Fixtures mit Nicht-ASCII-Inhalt als `\x…`; kein BOM, NBSP, Zeilentrenner im Diff |
| `[REDACTED]`-Escape (`SPEC-172` Nr. 6 n), Rundlauf mit dem Leser | geprüft, ohne Befund: r18, r19 rot, `TestRundlauf` trägt den Marker; der Bezug zu `slice-v4-schreiber-marker-text` steht in §6 Rückgabe 1 mit Ausgang und Adresse |
| Ungeänderte Dateien | geprüft, ohne Befund: `git diff --stat 5f235be..HEAD` über `jsoncodec`, `src/hexagon`, `src/bootstrap`, `src/adapters/driving`, `leser.go`, `leser_test.go`, `spec/`, `harness/` und `.a-check.yml` ist leer (Zeile 40) |
| Größenmarke | geprüft, ohne Befund: 3229 Zeilen über die 13 Pfade des Blocks bei Marke 5900 |
| Plan §3, §6 (Raster, Zusatztabelle, Prosa-Pfad der Ausnahmen-Datei) | geprüft, ohne Befund: jede Zelle trägt Ort; der Pfad der Ausnahmen-Datei steht als Prosa; offene Ausgänge tragen Adresse |
| Kommentare im Produktcode (`AGENTS.md` §3.7) | geprüft, ohne Befund: keine Konjunktive über verworfene Alternativen, kein abwesender Text |
| Geteilte Funktionen (`zeitForm`, `zeitFelder`, `zeitFeldGueltig`, `bruchMillis`, `kodiere`, `wurzelGueltig`) | geprüft, ohne Befund außer F-5: Zerlegung deckt die Zweige, keine zweite Kopie im Paket |
| Tests der Mengen mit Zählprüfung (Zeilen 11 bis 15, 26, 30) | geprüft, ohne Befund: Orakel in den Tests sind Tabellen, Literale, reguläre Ausdrücke und Standardbibliothek, nicht der Schreiber (außer `TestGleichzeitig` gegen den einzelnen Aufruf, wie die Zusage es nennt) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 2 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe · Gate-Fänger ohne Mutation benannt · Fänger hängt an der Umgebung des Laufs · Kommentar nennt mehr als der Test · Äquivalente Mutation durch doppelte Prüfung · Abdeckung nach Anweisung statt nach Wert

## Verdikt

**Merge-blockierend:** ja — kein HIGH, aber F-1 und F-2 sind MEDIUM und blockieren nach dem Maßstab des Skills; sie sind klein (ein Fall `http://h:99999` in Zeile 25; ein Negativ mit Grund und Folge oder ein Fall für `a-check` in Zeile 39) und gehen an den Implementer. Der Merge wartet auf deren Auflösung.

**Übergabe:** Findings gehen an den Implementer; die Finding-Klassen gehen in die Slice-Closure §7. Der Bericht ersetzt keine Verifikation gegen die DoD.
