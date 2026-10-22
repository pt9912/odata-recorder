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
   Endet `go test` mit einem Exit-Code ungleich 0 oder durch ein Signal, ist das der Befund (Exit 1): Text `go test -race:` und dessen Ausgabe nach stderr; stdout und stderr von `go test` stehen beide vollständig darin, eine Reihenfolge der Ströme sagt dieser Vertrag nicht zu.
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
   Endet `go` selbst durch ein Signal (etwa weil der Kernel es wegen des Speichers beendet), hat es keinen Exit-Code; auch das ist ein Befund: Text `go test -race:` und die Ausgabe bis dahin, Exit 1, ohne das Signal zu nennen.
   `go test` fährt alle Pakete zu Ende (`-failfast` steht nicht in den Argumenten): jedes Paket mit Befund steht in der Ausgabe mit einer Zeile `FAIL` und seinem Importpfad, auch bei Rennen in mehreren Paketen und bei einem roten Test ohne Rennen in einem weiteren.
   Treffen in einem Test Rennen und Zeitgrenze zusammen, stehen beide Meldungen in der Ausgabe (`WARNING: DATA RACE` und `panic: test timed out`); der Exit-Code ist 1.
   Die Reihenfolge der Meldungen ist die des Testbinary und keine Zusage dieses Vertrags: `pruefung race` reicht die Ausgabe durch und ordnet nichts um.
4. **Image und Go-Version.** Die Stufe `race` läuft auf `golang:1.27.1-trixie@sha256:8f58fd67ea075142d947a60e0caa4317746a55118d312f027793d382c7741734` (Debian 13, Go 1.27.1, gcc 14.2.0).
   Herkunft: am 2026-10-22 mit `docker buildx imagetools inspect` aus dem Tag `1.27.1-trixie` des offiziellen Images `golang` aufgelöst; der Digest ist der des Index über alle Plattformen, nicht der einer einzelnen.
   Der Pin steht als `RACE_IMAGE` in `harness/mk/go-gates.mk` und geht der Stufe als Build-Argument zu; ein fehlendes oder leeres `RACE_IMAGE` lässt den Bau scheitern, ein Image ohne Compiler lässt `go` mit `C compiler "gcc" not found` enden.
   Die Go-Version stimmt mit der von `deps` überein (`golang:1.27-alpine`, Go 1.27.1): die Stufe vergleicht die erste Zeile von `/usr/local/go/VERSION` beider Images (`go1.27.1`) und bricht bei einem Unterschied ab, Text `Go-Version von race … weicht von deps … ab`.
   Die zweite Zeile der Datei (`time …`) ist ein Zeitstempel und keine Version und geht nicht in den Vergleich ein; fehlt die Datei, etwa in einem Image ohne Go, scheitert der Bau an der Meldung von `head`.
   `deps` und `RACE_IMAGE` werden gemeinsam angehoben, eine Anhebung ist ein bewusster Commit.
   Die Module kommen aus `deps` (`/go/pkg/mod`).
   Die Stufe setzt `CGO_ENABLED=0`, `GOTOOLCHAIN=local` und `GOFLAGS=-mod=readonly` selbst, wie `deps` und `lint` (sie erbt aus `FROM <RACE_IMAGE>` nichts aus `deps`); jeder `RUN` der Stufe läuft mit `--network=none`, auch der Vergleich der Version.
   Ein Modul, das in `deps` fehlt, lässt den Lauf mit der Meldung von `go` scheitern, das Ziel ist rot.
   Dass diese Umgebung, das Netz und der Digest so in der Stufe stehen, fängt kein Fall (Nr. 9).
   Die Stufen `deps`, `source`, `pruefung`, `test` und `build` bleiben ohne Compiler und mit `CGO_ENABLED=0`.
   Ein `CGO_ENABLED=1` in `deps`, `source`, `test` oder `build` lässt `go` mit `cgo: C compiler "gcc" not found` enden, `make build` und `make test` werden rot (je Stufe gemessen am 2026-10-22; `src/` bindet Pakete der Standardbibliothek mit cgo-Fassung, etwa `net`, und die braucht den Compiler).
   In `pruefung` bliebe es grün; Grund und Folge stehen in Nr. 9.
   Die Stufen `lint` und `gegenproben` laufen auf dem Image von golangci-lint, das einen Compiler trägt; ein `CGO_ENABLED=1` dort bliebe grün.
   Grund, das hinzunehmen: beide Stufen liefern kein Artefakt, das ausgeliefert wird, und ihr Urteil hängt nicht von cgo ab; Folge im Betrieb: keine, das Binary entsteht allein in `build`.
5. **Zeitgrenze, Speicher, Abbruch.** `pruefung race` setzt kein `-timeout`; es gilt die Standardgrenze von `go test`, 10 Minuten je Paket.
   Überschreitet ein Test sie, bricht `go test` das Paket ab (`panic: test timed out after …` mit den Goroutinen), und das ist ein Befund (Nr. 3).
   Ein Muster `-timeout=<Dauer>` stellt die Grenze für einen Aufruf um; nur die Gegenprobe tut es (Fall 36 von `SPEC-102`).
   Unter `-race` laufen Tests langsamer und brauchen mehr Speicher.
   Gemessen am 2026-10-22 auf einem Host mit 20 Kernen, beide Stufen bei kaltem Build-Cache: die Stufe `race` brauchte das 1,1- bis 1,3-fache der Stufe `test`; je Paket weicht der Faktor stark vom Faktor der Stufe ab, und für einen Test mit Zeitgrenze zählt der seines Pakets.
   Der Speicherbedarf ist nicht gemessen; keiner dieser Läufe endete durch den Speicher.
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
   - Die Verdrahtung: dass die Stufe `race` `pruefung race /src` ruft, den Quellstand der Stufe `source` liest und dass `test-race` an `GATE_CHECKS` hängt, fängt kein dauerhafter Fall (wie `SPEC-036` Grenze); das Doku-Gate fängt nur die Zeile des Ziels in `harness/README.md`. Folge im Betrieb: fehlte die Verdrahtung oder läse die Stufe einen anderen Quellstand, bliebe `make gates` grün, ohne dass ein Rennen gemeldet würde. Beleg: der Implementer fährt vor der Übergabe ein Rennen in einer Wegwerfkopie des Quellstands durch `make test-race` und nennt das Rot.
   - Die Umgebung der Stufe (Nr. 4): `GOFLAGS=-mod=readonly`, `GOTOOLCHAIN=local`, `CGO_ENABLED=0`, `RUN --network=none` und der Digest im Pin von `RACE_IMAGE`; kein Fall fängt, dass sie im `Dockerfile` und in `harness/mk/go-gates.mk` stehen.
     Gemessen am 2026-10-22 im Image von `RACE_IMAGE`: `go` hat dort kein `GOFLAGS` und arbeitet ohne `vendor/` mit `-mod=readonly` als Vorgabe (`go help build`), `GOTOOLCHAIN` ist `local`, `CGO_ENABLED` ist 1, und `pruefung race` setzt es selbst (Nr. 1).
     Grund (Gleichwertigkeit) für `GOFLAGS`, `GOTOOLCHAIN` und `CGO_ENABLED`: ohne die Zeilen läuft derselbe Aufruf mit demselben Urteil. Folge im Betrieb: keine.
     Grund für `--network=none`: das Urteil über Rennen hängt nicht am Netz, verloren wäre die Netzlosigkeit des Laufs. Folge im Betrieb: ein fehlendes Modul oder eine andere Toolchain würde geholt, statt den Lauf scheitern zu lassen, und der Lauf bliebe grün; es fiele erst beim Lesen des `Dockerfile` auf, wie bei den Stufen `test`, `build`, `lint` und `gegenproben`, deren `RUN --network=none` kein Gate prüft.
     Grund für den Digest: das Urteil hängt am Detektor von `go`, dessen Version der Vergleich mit `deps` bindet (Nr. 4). Folge im Betrieb: fehlte der Digest, zöge Docker den beweglichen Tag `1.27.1-trixie`, Debian und gcc könnten sich ohne Commit verschieben, und der Pin trüge die Reproduzierbarkeit des Images nicht mehr; die Digests von `deps` und `lint` prüft ebenfalls kein Gate.
   - Die Stufe `pruefung` bei `CGO_ENABLED=1` (Nr. 4): ein `ENV CGO_ENABLED=1` dort lässt `make build` und `make test` grün (gemessen am 2026-10-22: `pruefung test: ok`). Grund: das Programm `pruefung` (`tools/pruefung/`) bindet nur Pakete der Standardbibliothek ohne cgo (`bytes`, `context`, `errors`, `fmt`, `io`, `os`, `os/exec`, `path/filepath`, `strings`, in den Tests `testing`), so dass das fehlende `gcc` nichts trifft, und das Binary des Produkts entsteht in `build` aus `source`, nicht aus dieser Stufe. Folge im Betrieb: keine; das Urteil der Stufen `test`, `build` und `race` ändert sich nicht. Bindet `tools/pruefung` künftig ein Paket mit cgo-Fassung (etwa `net`), wird die Stufe bei `CGO_ENABLED=1` rot; einen Fänger gegen den Wechsel selbst gibt es nicht.
   - Die zweite Zeile von `/usr/local/go/VERSION` (Nr. 4): dass der Vergleich nur die erste Zeile nimmt, fängt kein Fall; ob zwei Images derselben Version in der zweiten Zeile abweichen, ist nicht geprüft, und kein Lauf stellt es her. Grund (Gleichwertigkeit): der Zeitstempel sagt nichts über den Detektor, die Version bindet ihn. Folge im Betrieb: keine für das Urteil über Rennen; ein Vergleich der ganzen Datei machte den Bau bei einem Unterschied allein im Zeitstempel rot.
10. **Randformen von `pruefung race`.** Die Randformen 1, 2, 7, 10, 11 und 14 von `pruefung test` (`SPEC-036`) gelten für `race` wörtlich, mit `race` an der Stelle von `test`: ein Aufruf ohne `Wurzel`, mit leerem Argument oder mit falscher Zahl von Argumenten ist `Aufruf:` mit den drei Formen und Exit 2; `go` ist das einzige Programm, das `race` aufruft.
    Randform 8 (kein Befund bei einem Verzeichnis ohne Go-Dateien) gilt für `race` nicht: dort endet `go test` mit 1 (Nr. 2), und das ist der Befund.

**Gegenprobe:** Die Fälle 34 bis 36 von `SPEC-102` fahren `pruefung race` in einem Wegwerfmodul: ein sauberer nebenläufiger Fall, Datenrennen in je einem Paket unter `src/hexagon/`, `src/adapters/` und `src/bootstrap/` samt einem roten Test ohne Rennen, und ein Test, der ein Rennen erzeugt und danach blockiert, bis die Zeitgrenze abläuft.
Das Rennen dieser Fälle ist nicht zeitabhängig (Nr. 8): zwei Goroutinen ändern denselben Zähler ohne Kante zwischen ihnen, und der Test wartet erst danach auf beide.

**Stand:** vorhanden: `pruefung race` (`tools/pruefung/`), die Stufe `race` des `Dockerfile`, das Ziel in `GATE_CHECKS` und die Fälle 34 bis 36 von `SPEC-102`.

**Grenze:** Das Ziel belegt, dass `go test` unter dem Detektor in den Paketen des Quellstands keinen Befund meldet; was es nicht sieht, steht in Nr. 8 und Nr. 9.
Es senkt keine Schwelle der übrigen Gates und ersetzt keinen Test: ein Wirkungs-Fänger bleibt dort, wo die Wirkung benennbar ist, und ein Rennen ohne ungeordneten Zugriff fängt nur er.
