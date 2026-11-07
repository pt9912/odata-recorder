# Verifikations-Bericht: slice-v2-payload-schreiben — 2026-11-07

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan, DoD, Spezifikation und ADRs. Nicht die Frage des Reviewers (Diff gegen Plan und Hard Rules) und nicht die des Validators.

**Gegenstand:** `slice-v2-payload-schreiben`, Zweig `slice-v2-payload-schreiben`, HEAD `a1f83ea` (nach der Fix-Runde), Basis `5f235be`, Diff `5f235be..a1f83ea`: 15 Dateien (13 des Blocks `groessenmarke`, Plan, Review-Bericht).

**Geprüft gegen:** Plan (§3 Tabelle mit 43 Zeilen, Mengen mit Achsen, Pflicht-Mengen, Raten; §4 Block; §6 Raster und Zusatztabelle), `SPEC-172` Nr. 6 a bis o und Nr. 7, `SPEC-168` Nr. 1 bis 9, `SPEC-169` Nr. 2, `SPEC-170` Nr. 4, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md), `AGENTS.md` §3.9 bis §3.18, Review-Bericht `2026-11-07-slice-v2-payload-schreiben-review.md` (F-1 bis F-6).

**Methode.** Einmal gefahren. Belege des Implementers nicht übernommen. Eigener `make gates` mit Log im Scratchpad (`verif-v2ps/gates.log`), eigene Mutationen nur über `tools/harness/mutationslauf.sh` (unter `flock /tmp/odata-recorder-heavy.lock`, `--memory 2g`, `--cpus=4`, Image `odata-recorder:race`), eigene Messung von Anweisungs-Abdeckung und Laufzeit in einem Container mit dem Baum per `docker cp` (ohne Netz, ohne Mount). Der Host war zeitweise stark belastet (Load bis 52); die Last steht je Zeile in der Ergebniszeile des Werkzeugs.

---

## Urteil

**Bestanden, ohne Befund der Klassen HIGH oder MEDIUM.** Die DoD-Punkte, die der Verifier prüfen kann (Liefer-Punkte 1 und 2, `make gates`, V4 unberührt, Größenmarke, Rundlauf), sind belegt. Die Fix-Runde deckt die Klassen von F-1, F-3 und F-6 (35 von 36 gefahrenen Mutationen an nicht genannten Elementen rot, die eine grüne ist äquivalent und begründet); F-2 ist gemessen reproduziert. Offen sind nur Closure-Handlungen des Planners (siehe *Für die Closure*), kein Mangel am Code.

---

## 1. `make gates` (eigener Lauf, `exit=0`)

| Gate | Ausgabe im Log | Anmerkung |
|---|---|---|
| `a-check` | `gesamt: 0 Befund(e)` | frisch gelaufen |
| `baseline-verify` | `v6.18.1 OK — 54 Dateien` | frisch gelaufen |
| `docs-check` | `d-check: 714 Datei(en) geprüft, 0 Befund(e)` | frisch gelaufen |
| `groessenmarke` | `slice-v2-payload-schreiben.md: ok, Messung 3293 von Marke 5900` | frisch gelaufen; eigene Nachmessung `git diff --numstat 5f235be..HEAD` über die 13 Pfade des Blocks: 3293 |
| `migration-check`, `schema-check` | `migration-check: ok` | frisch gelaufen |
| `zeichen-check` | `zeichen: ok, 1455 Datei(en) geprüft` | frisch gelaufen |
| `build` | Stufe gelaufen | — |
| `test` (`pruefung test`), `test-race` (`pruefung race`), `lint` | Docker-Schicht `CACHED` | **nicht neu ausgeführt**: der Baum ist inhaltsgleich zu einem früheren erfolgreichen Lauf (Rolle davor), Docker nahm das Ergebnis aus dem Cache |
| `handbuch-check` | gelaufen | — |

**Behauptung ohne eigene Ausführung, deshalb nachgefahren:** Weil `test`, `test-race` und `lint` aus dem Cache kamen, habe ich das Paket `jsoncodecv2` selbst unter dem Detektor gefahren (Container mit dem Baum von `docker cp`, `CGO_ENABLED=1`, `go test -count=1 -race`): `ok`, 29 s. Die Nullprobe der Mutationsläufe (unten) fuhr das Paket mit demselben Baum je einmal unverändert. `gofmt`/`go vet`/`golangci-lint` am HEAD habe ich nicht neu erzeugt; ihr Ergebnis hängt an einer inhaltsgleichen Schicht (grün).

**Beobachtung (INFO).** Das Image-Tag `odata-recorder:race` wird von allen Worktrees geteilt; ein `docker run` mit dem darin gebackenen `/src` fährt den Baum des zuletzt bauenden Zweigs (mein erster Abdeckungslauf lief gegen fremden Code und meldete „no tests to run“). Das Werkzeug `mutationslauf.sh` ist davon nicht betroffen, es schickt den Baum per `docker cp`. Wer die Abdeckung misst, kopiert den Baum selbst hinein.

## 2. Plan gegen Code

- **Dateien.** Der Block `groessenmarke` nennt 5 Quell- und 8 Testdateien; der Diff trägt genau diese 13 unter `src/adapters/driven/odata/jsoncodecv2/` plus Plan und Review-Bericht. Der Plan nennt die Dateien so, wie sie liegen (`schreiber.go`, `_bau`, `_kanon`, `_adresse`, `_ausgabe`).
- **Ungeänderte Dateien (Zeile 40).** `git diff --stat 5f235be..HEAD -- src/adapters/driven/odata/jsoncodec src/hexagon src/bootstrap src/adapters/driving` ist leer; ebenso `leser*.go`, `.a-check.yml`, `spec/`, `harness/`. Gegen `main` (Merge-Basis) ändert `spec/spezifikation.md` nur die Stellen von `SPEC-172` (Zeilen 2600, 2605) und eine Zeile der Historie (3040), keine andere Nummer. `model.AlleCodes()` unverändert (kein Diff in `src/hexagon`).
- **Tabelle Zusage · Fänger gegen den Vertrag.** Jede der 43 Zeilen nennt einen Test; alle 41 genannten Testnamen existieren im Paket (43 `func Test…`). Die Zeilen 41 bis 43 sind akzeptierte Negative mit Grund und Folge, Zeile 40 ist die benannte Ausgabe (von mir gefahren, s. o.). Gegen `SPEC-172` Nr. 6 h bis o und `SPEC-169` Nr. 2 habe ich `wurzelGueltig`, `segmentGueltig`, `kodiere`, `kanonGanz`, `kanonDezimal`, `kanonGUID`, `zeitpunktMillis`, `text` und `stringWert` Zeile für Zeile gelesen: kein Abweichen vom Wortlaut (Schema klein, Host-Zeichen `-._~`, IPv6 nichtleer aus Hex, `:`, `.`, Port 1 bis 5 Ziffern ohne Wertprüfung, Segment ohne `.`/`..` auch als `%2E`, Länge 2048; Dezimal höchstens 1024 Zeichen; Bruchteil 1 bis 3 Ziffern, letzte nicht `0`; `[REDACTED]` mit Escape für `[`).
- **Raster §6.** Jede Zelle trägt einen Ort; die Zusatztabelle nennt für jede Allaussage einen Fänger. Der Ausgang „offen“ steht bei den Punkten, die die Closure schließt (unten).
- **Rohbytes (`AGENTS.md` §3.18).** `LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}' schreiber*.go` über alle 13 Dateien: leer. BOM, NBSP, U+2028/2029, Nullbreiten-Zeichen: 0 Treffer je Datei. Nicht-ASCII-Zeichen stehen nur in Kommentaren und Meldungstexten als UTF-8 (`Schlüssel`, `gültig`); Fixtures mit Rohbytes sind `\x…`. `make zeichen-check` bestätigt es (1455 Dateien).

## 3. Fixes gegen die Klasse (`AGENTS.md` §3.17)

Eigene Mutationen an Elementen, die weder der Review-Bericht noch der Plan nennt. 27 der Klasse `andere`, 9 der Klasse `nebenläufig` (`n_detektor` 20, `n_ohne` 0: ein Zähler ohne Detektor hat keine Wirkung), jede ersetzt eine Textstelle des Produktcodes genau einmal. Messender: Verifier. Listen: `verif-v2ps/andere.json`, `verif-v2ps/neben.json` im Scratchpad.

### 3.1 F-1 (Port, `TestWurzel`/`TestWurzelPort`)

Menge: Port der `Wurzel`, ein bis fünf Ziffern, Wert nicht geprüft; Elemente 10 gültige Werte und 3 ungültige Formen je 4 Operationen (80 und 12 Zellen); zusätzlich 256 Bytes je Stelle im Port (`TestWurzelZeichenMenge`).

| Mutation (nicht genannt) | Ergebnis | roter Test |
|---|---|---|
| v01 kein Port (`i == 0`) wird angenommen | rot | `TestWurzel` |
| v02 `-` als Portzeichen | rot | `TestWurzelZeichenMenge` |
| v03 Länge 2048 abgelehnt (`>=`) | rot | `TestWurzel` |
| v04 `_` im Host entfernt | rot | `TestWurzel` |
| v05 Segment `.` angenommen | rot | `TestWurzel` |
| v06 IPv6 `[]` angenommen | rot | `TestWurzel` |

Die genannten Mutationen des Reviewers (fünfstelliger Port über 65535 abgelehnt) bleiben mit `99999`, `65536`, `00000` abgedeckt; ein Port, dessen Stellen nur aus Ziffern 0, 1, 3, 5, 6, 9 bestehen, ist die Grenze der gültigen Werte, die kein Code-Pfad unterscheidet (`ziffer` prüft einen Bereich, keine Einzelwerte).

### 3.2 F-2 (a-check, Zeile 39)

Auf einer Wegwerfkopie im Scratchpad (`src`, `tests`, `go.mod`, `.a-check.yml`) mit einer Datei `zz_probe.go`, die den Adapter `sqlite` importiert: `lateral-adapter: 1`, `gesamt: 1 Befund(e)`; dieselbe Kopie ohne die Datei: `gesamt: 0 Befund(e)`. Die Aussage der Zeile 39 trägt. Die Wegwerfkopie liegt im Scratchpad, nicht im Zweig.

### 3.3 F-3 (Zone der Maschine, `TestZeitpunktZone`)

Die Menge ist 3 Zonen × 31 Werte, der Test startet sich dreimal mit `TZ`. Gefahren: v27 (`time.UTC` zu `time.Local` in der Kalenderprüfung `zeitFeldGueltig`, ein Element, das der Plan nicht nennt): **grün**, und zwar zu Recht: die Prüfung vergleicht nur Jahr, Monat und Tag um 00:00 (in keiner der drei Zonen eine Lücke der Sommerzeit zu dieser Stunde); sie ist so äquivalent wie die drei Kalendervergleiche. Der Plan führt diese eine Stelle nicht unter *Äquivalente Mutationen* (INFO, Nachtrag bei der Closure). Die zwei roten Zonen-Mutationen des Plans (t1, t2) sind die Stellen, an denen die Zone wirkt.

**Robustheit und Laufzeit des Kindprozesses.** Allein: `TestZeitpunktZone` 3,0 s unter dem Detektor (`-race -v`), das ganze Paket unter `-race` 29 s. Der Test hängt von `os.Args[0]`, `time/tzdata` (importiert) und der Variablen `TZ`, nicht von der Zone des Hosts; er prüft vor den 31 Werten, dass `TZ` wirkt (Versatz ungleich 0). Er lief ohne Hänger in der Nullprobe jedes Mutationslaufs (ganzes Paket, ohne Detektor), im Stufe-2-Lauf von v27 und in meinem Lauf unter dem Detektor. Restrisiko: eine Umgebung, in der der Test seine eigene Binärdatei nicht starten darf, fiele mit einer klaren Meldung (`Zone <Name>: <Fehler>`), nicht still.

### 3.4 F-6 (Aufrufart 3 über alle 36 Codes, `TestGleichzeitig`)

Menge: `len(model.AlleCodes())` = 36; die Schleife `for k := range codes` läuft über alle, der Beginn hängt von `g` ab. Eigene Zähler (`zk++` auf einer Variable des Pakets) in `SchreibeFehler` für drei Codes, die weder der Review-Bericht (`ODR-E9001`) noch der Plan (`ODR-E5005`) nennen:

| Mutation | mit Detektor | Last |
|---|---|---|
| z-code-E2401 | 20/20 | 4,02/4 |
| z-code-E5008 | 20/20 | 2,39/4 |
| z-code-E7006 | 20/20 | 2,33/4 |

### 3.5 Weitere Zweige unter dem Detektor (Register-Muster `zweig-ohne-lauf-unter-dem-detektor`)

Zählprüfung gegen die Modell-Menge: Aufrufart 3 deckt alle 36 Codes, Aufrufart 4 alle sieben Arten (`Edm.String`, `Edm.Boolean`, `Edm.Int32`, `Edm.Int64`, `Edm.Decimal`, `Edm.Guid`, `Edm.DateTimeOffset`) und `null`, die Aufrufarten 0 bis 2 die drei übrigen Operationen mit Ziel; die Ablehnungen laufen als Aufrufart 10 (Wurzel, Art, Text), 6 (Argumente), 7 (Kontext, `dst`). **Anweisungs-Abdeckung gemessen:** `go test -run '^TestGleichzeitig$' -coverpkg=<Paket> -coverprofile`: 275 Blöcke in den Quelldateien des Schreibers, **0 mit Zähler 0**. Zweig-Zähler in sechs Rückgaben, die der Plan nicht nennt:

| Mutation (Zweig, in dem der Zähler steht) | mit Detektor | Last |
|---|---|---|
| z-guid-laenge (Länge ungleich 36 in `kanonGUID`) | 20/20 | 1,62/4 |
| z-zeitform (Form des Zeitpunkts) | 20/20 | 1,58/4 |
| z-ipv6-leer (`[]` in `hostEnde`) | 20/20 | 1,35/4 |
| z-set-unbekannt (Set unbekannt in `setTyp`) | 20/20 | 9,86/4 |
| z-kurzer-write (`io.ErrShortWrite` in `ausgeben`) | 20/20 | 13,58/4 |
| z-typ-abstrakt (abstrakter Typ in `typ`) | 20/20 | 7,04/4 |

Alle neun Zähler sind in jedem Prozess rot (9 von 9 Mutationen, 180 von 180 Prozessen). Die Rate gilt als Größenordnung: 20 statt 50 Prozesse, der Host war belastet (Last bis 13,6), Messender Verifier. Offen bleibt, was der Plan als akzeptiert nennt: ein Zweig nach Wert innerhalb einer Anweisung (etwa ein Teilausdruck einer `||`-Bedingung) ist durch Block-Abdeckung nicht unterscheidbar.

### 3.6 Andere Klassen der Fix-Runde und des Vertrags (Stichprobe an nicht genannten Elementen)

Alle `rot` (der zuständige Test wird rot, Nullprobe grün): v07 `=` und v08 `'` aus der Kodierung genommen, v09 Hex-Ziffern des unteren Halbbytes klein (`TestAdresseKodierung`); v10 `Edm.Int32` auf 31 Bit (`TestTextKanonisch`); v11 `-0` als `Edm.Decimal` (`TestZifferKlasse`); v12 Bindestrich an Stelle 18 nicht geprüft, v13 Akzent vor `a` als Hex-Ziffer (`TestGuidStellen`); v14 Trennzeichen an Stelle 7, v15 an Stelle 16 des Zeitpunkts nicht geprüft, v16 Vorzeichen in Zeitpunkt-Feldern (`TestZeitpunktKlasse`); v17 U+007F escaped (`TestStringKlasse`); v18 `[redacted]` mit Escape (`TestMarkerText`); v19 `Edm.Date` nicht abgelehnt (`TestPropertyNamen`); v20 Auswahl in `SchreibeEntity`, v21 in `SchreibeCollection` nicht geprüft (`TestAuswahl`); v22 `target` nicht verglichen (`TestFehlerobjektTabelle`); v23 `Edm.Boolean` als String (`TestWerteBytes`); v24 `Edm.DateTimeOffset` gehört zu `Edm.String` (`TestArtGegenTyp`); v25 U+0000 im Schlüssel nicht geprüft (`TestSchluessel`); v26 leeres Set besteht die Suche (`TestSetUngueltig`).

**Zahl:** 36 Mutationen, 35 rot, 1 grün (v27, äquivalent, begründet). Zusammenfassungszeilen des Werkzeugs: `mutationen=27 rot=26 gruen=1 abbruch=0 uebersetzungsfehler=0` und `mutationen=9 rot=9 flatternd=0 abbruch=0 uebersetzungsfehler=0`.

## 4. F-4 (Kommentar)

Der Kommentar von `TestGleichzeitig` sagt jetzt „erst alle 11 Aufrufarten gemischt in 50 Läufen, dann jede der 11 Aufrufarten einzeln in 30 Läufen“; im Code gilt `const varianten = 11` und die Schleife `for _, i := range alle`. Der Kommentar nennt, was der Test fährt.

## 5. DoD

| Punkt | Stand |
|---|---|
| Schreiben nach `SPEC-172` Nr. 6, Erfüllung von `outbound.ODataSchreiber` und `SPEC-168` Nr. 1 bis 9, Zeitpunkt ohne ganze Millisekunde `ErrUngueltig`, Rundlauf (Nr. 7), V4 unberührt | erfüllt (Gates, Tests, Diff leer in `jsoncodec`) |
| Je Zelle des Rasters ein Ort und ein Fänger; Schleifen mit Zählprüfung | erfüllt (Stichproben s. o.; Zählprüfungen in den Tests, z. B. `Menge … will 13`, `3*31`) |
| `make gates` grün | erfüllt (`exit=0`; `test`/`test-race`/`lint` aus dem Cache, nachgefahren s. §1) |
| Review durchgeführt, Report liegt vor | erfüllt (`docs/reviews/2026-11-07-slice-v2-payload-schreiben-review.md`); Haken im Plan noch offen |
| Doku-Update falls öffentlicher Vertrag berührt | erfüllt (Architect: `SPEC-172`); Handbuch unberührt, Binary unverändert |
| Closure-Notiz (§7), Beobachtungs-Register, Ausgang je Risiko (§6), drei Paarungen | **Closure-Handlungen**, noch nicht getan (siehe unten) |

## 6. Für die Closure (Adresse: Planner)

1. Haken in §2 setzen; §7 füllen. Dabei den Wortlaut „das Binary blieb unverändert“ (ein V2-Dienst erreicht es erst mit `slice-v2-bootstrap`) tragen.
2. §6, Ausgänge: *Zeit- und Zahlenschreibweisen*, *Fakes der Ports*, *Randformen*, *Nebenläufige Zusage* stehen auf *offen* und sind nach dieser Verifikation zu schließen (eingetreten/entfallen); *Linie B* und *Integration mit `ODR-E6007`* bleiben *weiter offen* bis zum Merge, *Gate-Kette* bleibt *weiter offen (nicht zu erwarten)*.
3. Register: die Beobachtung „Schreiber und Leser nicht als Paar geprüft“ (§6 Rückgabe 1) in `BEO-REPO/` anlegen; das Muster `zweig-ohne-lauf-unter-dem-detektor` erhält einen weiteren Beleg: hier mit Block-Abdeckung gemessen (275 Blöcke, 0 ohne Lauf), das Muster ist am Schreiber geschlossen, ein Zweig nach Wert innerhalb einer Anweisung bleibt benannt.
4. Plan *Äquivalente Mutationen*: eine Zeile für `time.Local` in `zeitFeldGueltig` (v27, grün, Grund: nur der Kalendertag um 00:00 wird verglichen; Folge: keine).
5. Die Ergebnisse der Fix-Runde dieses Berichts (36 Mutationen, 20 Prozesse je Zähler) gehören in den Lerneintrag, soweit die Closure ihn braucht; die Raten gelten als Größenordnung.

## 7. Grenzen dieser Verifikation (ehrlich)

- Die Mutationen der Spalte *Mutation* der 39 Zeilen des Plans habe ich nicht neu gefahren; die Stichprobe bestand aus eigenen Mutationen an anderen Elementen derselben Klassen (36 Läufe).
- `test`, `test-race` und `lint` stammen im `make gates` dieses Laufs aus dem Docker-Cache; nachgefahren ist das Paket unter dem Detektor, nicht der Gesamtlauf von `go vet ./...` und `golangci-lint`.
- Die Raten der Zähler (20 Prozesse mit Detektor, 0 ohne) sind Größenordnungen bei schwankender Last des Hosts.
- Nicht behauptet: die Verdrahtung des Schreibers im Binary (liegt in `slice-v2-bootstrap`), das Verhalten gegen einen echten V2-Client, und ob `make a-check` den Import von `jsoncodec` meldet (gemessen ist nur der Import von `sqlite`; für `jsoncodec` nennt der Plan zusätzlich `versionstrennung_test.go`).
