# SPEC-122 — Race-Detektor (`make test-race`)

**Werkzeug:** `make test-race` über das `Dockerfile` (Stufe `race`); die Prüfung ist `pruefung race` aus `tools/pruefung/` (`SPEC-036`); Image per Digest gepinnt, die Stufe läuft ohne Netz.

**Festlegung:** Das Ziel fährt `go test -race` mit cgo über alle Pakete, die `make test` fährt, und schlägt fehl, wenn `go test` einen Befund meldet: ein Datenrennen, einen roten Test, einen Fehler beim Bau, den Ablauf der Zeitgrenze oder das Ende durch den Speicher.
Es hängt an `GATE_CHECKS`, `make gates` fährt es.
Die Stufe liefert ein Urteil und kein Artefakt; das Binary entsteht in der Stufe `build`, ohne cgo.

1. **Aufruf.** `pruefung race <Wurzel> [Muster...]` arbeitet im Verzeichnis `Wurzel`; ohne Muster gilt `./...`.
   Es führt `go test -race -count=1 -trimpath -buildvcs=false <Muster>` aus, in der Umgebung des Aufrufers mit `CGO_ENABLED=1` für diesen einen Aufruf: die Stufen des `Dockerfile` setzen `CGO_ENABLED=0`, und ohne cgo endet `go test -race` mit der Meldung `-race requires cgo` und dem Exit-Code 2 (gemessen am 2026-10-22).
   `gofmt` und `go vet` führt es nicht aus; das tut `pruefung test`.
   Argumente, die mit `-` beginnen, gehen unverändert an `go` (`SPEC-036` Randform 13); sie stehen hinter den festen Argumenten, ein Muster `-timeout=…` stellt also die Zeitgrenze um (Nr. 5).
   `-count=1` steht fest, damit der Testcache von `go` keinen Lauf überspringt; die Stufe startet mit leerem `GOCACHE`, so dass es heute ohne Wirkung ist, und ein künftiger Cache-Mount ließe sonst Läufe aus.
   Endet `go test` mit einem Exit-Code ungleich 0, ist das der Befund (Exit 1): Text `go test -race:` und dessen Ausgabe, stdout gefolgt von stderr, nach stderr.
   Ohne Befund: Text `pruefung race: ok` nach stdout, Exit 0.
   `pruefung race` unterscheidet keine Ursache: es reicht Exit-Code und Ausgabe von `go test` durch.
2. **Pakete.** Alle Pakete des Quellstands der Stufe `source`, also dieselben, die `make test` fährt (Muster `./...`, Nr. 1).
   Es gibt keine Liste und keine Ausnahme: ein Paket, das hinzukommt, fährt das Ziel ohne Eintrag, und was die Stufe `source` an Quellen aufnimmt (heute `go.mod`, `go.sum` und `src/`; ein künftiges `tests/`, sobald die Stufe es aufnimmt), fahren `make test` und `make test-race` beide.
   Fehlt `src/`, scheitert schon die Stufe `source`.
   Ein Paket ohne Testdatei ist kein Befund (`go` meldet `[no test files]`), ein Quellstand ohne Paket ist einer (`go` meldet `no packages to test` und endet mit 1; beides gemessen am 2026-10-22).
   Die Module unter `tools/` gehören nicht dazu (Nr. 9).
   Die Pakete von Fremdcode, die `go test` mitbaut (etwa `modernc.org/sqlite` und `modernc.org/libc`), stehen unter dem Detektor und unter den Prüfungen von `checkptr`, die `-race` einschaltet; meldet er dort etwas, ist das Ziel rot, und der Befund geht an den Planner.
   Der Vertrag kennt kein Flag, das den Detektor oder `checkptr` für ein Paket abschaltet; eine befristete Ausnahme wäre ein Carveout mit Trigger, keine Zeile dieses Vertrags.
3. **Befunde.** Jede Ursache, die `go test -race` mit Exit-Code ungleich 0 enden lässt, ist ein Befund: ein Datenrennen (die Ausgabe nennt `WARNING: DATA RACE`), ein roter Test, ein Fehler beim Bau, der Ablauf der Zeitgrenze (Nr. 5), das Ende durch den Speicher (`signal: killed`).
   `go test` fährt alle Pakete zu Ende (`-failfast` steht nicht in den Argumenten): jedes Paket mit Befund steht in der Ausgabe mit einer Zeile `FAIL` und seinem Importpfad, auch bei Rennen in mehreren Paketen und bei einem roten Test ohne Rennen in einem weiteren.
   Treffen in einem Test Rennen und Zeitgrenze zusammen, stehen beide Meldungen in der Ausgabe, `WARNING: DATA RACE` zuerst, weil der Detektor beim Zugriff meldet, `panic: test timed out` am Ende; der Exit-Code ist 1.
4. **Image und Go-Version.** Die Stufe `race` läuft auf `golang:1.27.1-trixie@sha256:8f58fd67ea075142d947a60e0caa4317746a55118d312f027793d382c7741734` (Debian 13, Go 1.27.1, gcc 14.2.0).
   Herkunft: am 2026-10-22 mit `docker buildx imagetools inspect` aus dem Tag `1.27.1-trixie` des offiziellen Images `golang` aufgelöst; der Digest ist der des Index über alle Plattformen, nicht der einer einzelnen.
   Der Pin steht als `RACE_IMAGE` in `harness/mk/go-gates.mk` und geht der Stufe als Build-Argument zu; ein fehlendes oder leeres `RACE_IMAGE` lässt den Bau scheitern, ein Image ohne Compiler lässt `go` mit `C compiler "gcc" not found` enden.
   Die Go-Version stimmt mit der von `deps` überein (`golang:1.27-alpine`, Go 1.27.1; `/usr/local/go/VERSION` beider Images war am 2026-10-22 gleich): die Stufe vergleicht sie mit der von `deps` und bricht bei einem Unterschied ab, Text `Go-Version von race … weicht von deps … ab`.
   `deps` und `RACE_IMAGE` werden gemeinsam angehoben, eine Anhebung ist ein bewusster Commit.
   Die Module kommen aus `deps` (`/go/pkg/mod`), die Stufe läuft mit `RUN --network=none`, `-mod=readonly` und `GOTOOLCHAIN=local`; ein Modul, das in `deps` fehlt, lässt den Lauf mit der Meldung von `go` scheitern, das Ziel ist rot.
   Die Stufen `deps`, `source`, `pruefung`, `test` und `build` bleiben ohne Compiler und mit `CGO_ENABLED=0`: ein `CGO_ENABLED=1` dort lässt `go` mit `cgo: C compiler "gcc" not found` enden, `make build` und `make test` würden rot (gemessen am 2026-10-22 in den Images von `build` und `test`).
   Die Stufen `lint` und `gegenproben` laufen auf dem Image von golangci-lint, das einen Compiler trägt; ein `CGO_ENABLED=1` dort bliebe grün.
   Grund, das hinzunehmen: beide Stufen liefern kein Artefakt, das ausgeliefert wird, und ihr Urteil hängt nicht von cgo ab; Folge im Betrieb: keine, das Binary entsteht allein in `build`.
5. **Zeitgrenze, Speicher, Abbruch.** `pruefung race` setzt kein `-timeout`; es gilt die Standardgrenze von `go test`, 10 Minuten je Paket.
   Überschreitet ein Test sie, bricht `go test` das Paket ab (`panic: test timed out after …` mit den Goroutinen), und das ist ein Befund (Nr. 3).
   Ein Muster `-timeout=<Dauer>` stellt die Grenze für einen Aufruf um; nur die Gegenprobe tut es (Fall 36 von `SPEC-102`).
   Unter `-race` laufen Tests langsamer und brauchen mehr Speicher (der Faktor ist nicht gemessen; der Implementer misst ihn in der Stufe).
   Kippt ein Test mit einer Zeitgrenze seines Codes (eine Frist von 20 ms, ein Fenster von 250 ms) nur unter dem Detektor und ohne Rennen, ist das ein Befund an den Planner und keine Lockerung der Grenze im Gate.
   Bricht der Aufrufer den Bau ab, endet die Stufe ohne Ergebnis; `pruefung race` hält keinen Zustand und schreibt keine Datei (`SPEC-036` Randform 10).
6. **Wiederholung und Cache.** Bei unveränderter Eingabe der Stufe (der Quellstand der Stufe `source`, `go.mod`, `go.sum`, das Programm `pruefung`, der Pin des Images, das `Dockerfile`) fährt Docker sie nicht erneut, und `make test-race` meldet das Ergebnis des letzten Laufs, wie `make test`.
   Wer die Stufe erneut fahren will, baut mit `--no-cache-filter race`.
   Folge: ein Rennen, das ein Lauf nicht auslöste (Nr. 8a), bleibt bis zur nächsten Änderung der Eingabe unentdeckt.
   Derselbe Digest und dieselben Module ergeben dasselbe Ergebnis, soweit der Lauf deterministisch ist (Nr. 8).
7. **Plattformen.** Die Stufe läuft in einem Linux-Container auf der Plattform des Bau-Hosts (`--platform=$BUILDPLATFORM`, wie `deps`); gefahren ist sie auf `linux/amd64`.
   Das Image führt `386`, `amd64`, `arm64`, `arm/v7`, `ppc64le`, `riscv64` und `s390x`; `runtime/race` bringt Quellen für `amd64`, `arm64`, `ppc64le`, `riscv64`, `s390x` und `loong64` mit, nicht für `386` und `arm`.
   Auf einer Plattform ohne Detektor endet `go test -race` mit einer Meldung von `go` und dem Exit-Code ungleich 0, das Ziel ist rot; die Stufe verzweigt nicht nach der Plattform und überspringt nichts.
   Folge: ein Host mit `386` oder `arm` fährt `make gates` nicht grün, und nie grün ohne Detektor.
   Auf `arm64` fährt die Stufe nach dieser Regel; dieser Vertrag belegt es dort nicht, weil kein Host dafür vorliegt.
8. **Was der Detektor meldet.** Der Detektor meldet ein Datenrennen, wenn zwei Goroutinen im selben Prozess auf dieselbe Speicherstelle zugreifen, mindestens eine schreibt und keine Synchronisation (Sperre, Kanal, `WaitGroup`, atomare Operation) die Zugriffe ordnet.
   Die Reihenfolge, in der der Scheduler die Goroutinen laufen lässt, ist ohne Belang, und die Zugriffe müssen nicht gleichzeitig stattfinden.
   Gemessen am 2026-10-22 im Image von `RACE_IMAGE`, in einem Wegwerfmodul ohne Abhängigkeiten: zwei Goroutinen, die ohne Sperre denselben Zähler erhöhen und auf die der Test erst danach wartet, wurden in 100 von 100 Prozessläufen gemeldet (`-test.count=1`, je Lauf ein Prozess).
   Innerhalb eines Prozesses meldet der Detektor dasselbe Rennen nur einmal (ein Testbinary mit `-test.count=20` ergab 1 roten Lauf von 20); eine Rate gilt darum je Prozess, nicht je Wiederholung.
   Er meldet nicht:
   - **a) einen Zugriff, den kein Lauf ausführt.** Grund: der Detektor sieht nur ausgeführte Zugriffe; ein Pfad, den kein Test aus zwei Goroutinen durchläuft, und Code in einem Paket ohne Testdatei, den kein Test eines anderen Pakets aufruft, bleiben ohne Befund. Folge im Betrieb: ein Datenrennen dort ist ein Fehler im Produkt (ein Wert, der zur falschen Anfrage gehört; ein Panic nach dem gleichzeitigen Schreiben einer Map, das den Prozess beendet) und wird sichtbar, sobald ein Test beide Zugriffe ausführt;
   - **b) ein Rennen ohne ungeordneten Zugriff.** Jeder Zugriff liegt unter einer Sperre oder hinter einem Kanal, nur die Abfolge aus Prüfen und Handeln ist nicht atomar (eine Sperre, die zwischen den Schritten freigegeben wird). Grund: das ist kein Datenrennen im Sinn des Detektors, alle Zugriffe sind geordnet. Folge im Betrieb: der Fehler der Abfolge bleibt ohne Meldung, etwa ein Exchange, der nach dem Beginn von `Drain` angenommen wird (`SPEC-121` Nr. 2e); eine solche Zusage braucht in ihrem Vertrag einen eigenen Fänger oder ein benanntes Negativ.
9. **Nicht Teil des Ziels.**
   - Die Module unter `tools/` (`pruefung`, `gegenproben`; eigene `go.mod`, eigene Stufen ohne Compiler). Grund: sie gehören nicht zum Wurzelmodul, dessen Pakete `./...` trifft. Folge im Betrieb: ein Datenrennen im nebenläufigen Programm der Gegenproben könnte ein falsches Urteil einer Gegenprobe liefern, ein Produkt ist nicht betroffen; ein wiederholter Lauf zeigt eine unstete Gegenprobe.
   - Die Verdrahtung: dass die Stufe `race` `pruefung race /src` ruft, den Quellstand der Stufe `source` liest, mit `RUN --network=none` läuft und dass `test-race` an `GATE_CHECKS` hängt, fängt kein dauerhafter Fall (wie `SPEC-036` Grenze); das Doku-Gate fängt nur die Zeile des Ziels in `harness/README.md`. Folge im Betrieb: fehlte die Verdrahtung oder läse die Stufe einen anderen Quellstand, bliebe `make gates` grün, ohne dass ein Rennen gemeldet würde. Beleg: der Implementer fährt vor der Übergabe ein Rennen in einer Wegwerfkopie des Quellstands durch `make test-race` und nennt das Rot.
10. **Randformen von `pruefung race`.** Die Randformen 1, 2, 7, 10, 11 und 14 von `pruefung test` (`SPEC-036`) gelten für `race` wörtlich, mit `race` an der Stelle von `test`: ein Aufruf ohne `Wurzel`, mit leerem Argument oder mit falscher Zahl von Argumenten ist `Aufruf:` mit den drei Formen und Exit 2; `go` ist das einzige Programm, das `race` aufruft.
    Randform 8 (kein Befund bei einem Verzeichnis ohne Go-Dateien) gilt für `race` nicht: dort endet `go test` mit 1 (Nr. 2), und das ist der Befund.

**Gegenprobe:** Die Fälle 34 bis 36 von `SPEC-102` fahren `pruefung race` in einem Wegwerfmodul: ein sauberer nebenläufiger Fall, Datenrennen in je einem Paket unter `src/hexagon/`, `src/adapters/` und `src/bootstrap/` samt einem roten Test ohne Rennen, und ein Rennen vor einem Test, der die Zeitgrenze überschreitet.
Das Rennen dieser Fälle ist nicht zeitabhängig (Nr. 8): zwei Goroutinen ändern denselben Zähler ohne Kante zwischen ihnen, und der Test wartet erst danach auf beide.

**Stand:** entsteht mit dem Slice, der den Code liefert.

**Grenze:** Das Ziel belegt, dass `go test` unter dem Detektor in den Paketen des Quellstands keinen Befund meldet; was es nicht sieht, steht in Nr. 8 und Nr. 9.
Es senkt keine Schwelle der übrigen Gates und ersetzt keinen Test: ein Wirkungs-Fänger bleibt dort, wo die Wirkung benennbar ist, und ein Rennen ohne ungeordneten Zugriff fängt nur er.
