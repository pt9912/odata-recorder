# SPEC-039 — Architektur-Gate (`make a-check`)

**Werkzeug:** `make a-check`.

**Festlegung:** Hexagonale Importregeln gemäß `.a-check.yml`; keine Adapter-Imports im Core, keine lateralen Adapter-Imports.

1. **Wurzel.** Der Composition Root ist `src/bootstrap/**` (`composition_root`); nur er ist von den Richtungsregeln und von `tech` ausgenommen.
2. **Prüfpakete.** `tests/**` ist die Schicht `tests` ohne Rolle mit Kanten auf `model`, `ports-inbound` und `adapters-driven`; den Composition Root nennt sie ohne Kante.
   Ein Import von Services, Outbound-Ports oder Driving Adaptern aus `tests/` ist ein `wrong-direction`, ebenso ein Import von `tests/` aus `src/`.
   `tech` gilt für `tests/` wie überall: `modernc.org/sqlite` und `encoding/xml` außerhalb ihres Adapters sind ein `tech-leak`.
   Die Kante auf `adapters-driven` gilt für alle Driven Adapter, nicht für einen namentlich; die Bibliotheksgrenze hält `tech`.
   Eine weitere Kante der Schicht entscheidet eine neue ADR.
3. **Arbeitsbäume.** `.a-check.yml` nimmt `.wt/**` an der Wurzel aus (`exclude`); Entscheidung, Muster, Randformen und Fänger stehen in `SPEC-212`.
4. **Der Root als Ziel.** Der Composition Root ist keine Schicht, und ein Import auf eine Datei ohne Schicht bleibt von den Kanten unbeurteilt (gemessen: der Import des Root aus dem Adapter `admin` meldete 0 Befunde, als Blank-Import und als benannter Import mit Verwendung).
   Darum trägt `.a-check.yml` einen Eintrag im Block `tech`: der Importpfad des Root (`odata-recorder/src/bootstrap`) darf nur in Dateien unter `src/bootstrap/` und `tests/` stehen.
   Jeder andere Pfad, der ihn importiert, ist ein Befund, und zwar `core-impurity` im Modell, `app-impurity` in den Diensten, `port-impurity` in den drei Port-Schichten (Inbound, Outbound, Support) und `tech-leak` in den Adaptern (Driving und Driven) und bei einem Pfad ohne Schicht; das gilt für jede Form des Imports mit Anführungszeichen (ohne Alias, mit Alias, Blank-Import, Punkt-Import).
   Gemessen sind 32 Zellen (acht Quellen mal vier Formen): `tech-leak` 12, `core-impurity` 4, `app-impurity` 4, `port-impurity` 12.
   Zugesagt ist der Befund mit dem Pfad der Datei, nicht die Art: die Art wählt die Schicht der Quelle, nicht diese Regel.
   Die Zone gleicht dem Root selbst und Nr. 2: `tests/` nennt den Root ohne Kante und darf ihn nennen.
   Das Gate verschärft sich damit, keine Kante und keine Schicht ändert sich; eine weitere Zone ist eine Änderung dieser Nummer.

**Stand:** vorhanden.

**Gegenprobe:** Die Fälle 171 bis 173 von `SPEC-102`: Bibliothek in `tests/` (`tech-leak`), Import eines Dienstes aus `tests/` (`wrong-direction`), der erlaubte Weg über `bootstrap`, Inbound-Port und Driven Adapter (Exit 0); zu Nr. 3 die Fälle 205 bis 208.
Offen bleibt der Import von `tests/` aus `src/`: er ist gemessen ein `wrong-direction`, hat aber keinen eigenen Fall, weil keine Zeile der Datei ihn trägt, die jemand entfernen könnte; er entstünde nur durch eine neue Kante.
Offen bleibt Nr. 4: kein Fall der Gegenprobe führt die Regel; sie ist gemessen, an der Kopie des Baums je Schicht und Form des Imports und an den drei Pfaden der Zone, aber ein Entfernen der Regel oder ein Muster, das den Import nicht mehr trifft, färbt `make gegenproben` nicht.
Die Fälle (ein Import des Root aus `admin` meldet `tech-leak`, einer aus `tests/` bleibt ohne Befund) trägt der nächste Eingriff in `tools/gegenproben/` nach, der ohnehin geschieht.

**Grenze:** Prüft Importrichtungen, nicht Verhalten.
Zu Nr. 4: a-check gleicht die Zone als Teilstring des Pfads der Datei, nicht segmentgenau (ein Pfad, der `tests/` oder `src/bootstrap/` an irgendeiner Stelle enthält, zählt als Zone; im Repo liegt keiner außerhalb der Wurzel und von `tests/`), und es liest Importe text-heuristisch: den Import mit Raw-String und die Gruppe in einer Zeile (`import ("fmt"; "…")`) liest es nicht und nennt sie auch nicht als Hinweis (gemessen: 0 Befunde, kein Hinweis, je Form); `gofmt -l` in `make test` nennt beide für `src` und `tests` (gemessen: 2 von 2), nicht für `tools/mutationslauf` (dort fährt `make gates` weder `gofmt` noch `go vet`).
Ob es eine andere Schreibweise, die es nicht liest, als Hinweis nennt, ist nicht gemessen.
