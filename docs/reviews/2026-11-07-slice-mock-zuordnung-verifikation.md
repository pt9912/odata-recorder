# Verifikations-Bericht: slice-mock-zuordnung — 2026-11-07

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan, DoD, Spezifikation und ADRs. Nicht die Frage des Reviewers (Diff gegen Plan und Hard Rules) und nicht die des Validators.

**Gegenstand:** `slice-mock-zuordnung` (Teil 1 des Schnitts von `slice-mock-lesen`), Zweig `slice-mock-zuordnung`, HEAD `c82d4d7` (nach Fix-Runde und Entscheidung R6), Basis `03c9420`, Messung 2692 von Marke 5800.

**Geprüft gegen:** Plan (§3 Tabelle Z01 bis Z24, Mengen M1 bis M6, Absatz *Mutationen und Raten*, §4 Block, §6 Raster, Zusatztabelle, R1 bis R7), `SPEC-195` Nr. 1 bis 12, `SPEC-198` Nr. 4, 5 und 8, `SPEC-109` Nr. 5 und 6, `SPEC-199`, Review-Bericht `2026-11-07-slice-mock-zuordnung-review.md` (F-1 bis F-7).

**Methode.** Einmal gefahren. Belege des Implementers nicht übernommen. Eigener `make gates` (Log im Scratchpad `verif-zuordnung/gates.log`), eigene Mutationen nur über `tools/harness/mutationslauf.sh` (unter `flock /tmp/odata-recorder-heavy.lock`, `MUTATIONSLAUF_SPEICHER=3g`, `--cpus=4`, Image `odata-recorder:race`), eine Abdeckungsmessung unter dem Detektor in einem Container ohne Netz (`--memory=3g`). Listen: `verif-zuordnung/liste2.json` (36 Mutationen) und `liste3.json` (10), Ausgaben `mut.out`, `mut3.out`.

---

## Urteil

**Bestanden mit Auflagen.** Kein HIGH. Die DoD-Punkte, die der Verifier prüfen kann, sind belegt (`make gates` Exit 0, Größenmarke, Gate-Kette unberührt, keine `//nolint`, Rohbytes sauber, Teil 1 ohne `ODR-E6008`, 37 Codes, Katalog-Gate grün). Die Fixes der Fix-Runde decken ihre Klassen (Abschnitt 4). Zwei MEDIUM betreffen die Zusage Z22 (Nebenläufigkeit) und sind Auflagen vor der Closure; fünf LOW sind Lücken von Mengen, die der Fänger nicht ganz abdeckt.

| ID | Klasse | Befund (Kurz) | Auflage an |
|---|---|---|---|
| V-1 | MEDIUM | Der Zweig „Kennzeichen leer“ (R6, `metadaten.go`) läuft nicht unter `TestGleichzeitig`; der Absatz *Abdeckungsprofil* im Plan stimmt nicht mehr | Implementer (Plan §3, Test) |
| V-2 | MEDIUM | Ein Schreibzugriff in einem Zweig, den jede Goroutine nur einmal ausführt, wird vom Detektor nicht in jedem Prozess gemeldet (11/20, 36/50, 18/20, 47/50) | Implementer (Test oder Negativ mit Rate) |
| V-3 | LOW | M4 führt je Stufe nur eine Form: `Art` außerhalb der vier (Stufe URL) und `ctx` nil (Stufe Argumente) stehen in keinem Paar | Implementer |
| V-4 | LOW | M3 „mehr als ein Wert“ führt nur zwei Werte; zwei Werte mit einem leeren führt keine Zelle | Implementer |
| V-5 | LOW | Z10: die Schreibweise zählt für `HEAD` genauso wie für `GET`; die Menge führt nur `get`, `Get` | Implementer |
| V-6 | LOW | „ohne Ursache“ (`SPEC-195` Nr. 10) ist für `ODR-E5003` und `ODR-E5005` (Nr. 4) ungefangen | Implementer |
| V-7 | LOW | Die Reihenfolge von `Dienste` ist kein Faktor von M2: „letzter Treffer gewinnt“ fängt nur zufällig ein anderer Test | Implementer |

---

## 1. `make gates` (eigener Lauf, `EXIT=0`)

| Gate | Ausgabe im Log | Anmerkung |
|---|---|---|
| `a-check` | `gesamt: 0 Befund(e)` | frisch gelaufen |
| `docs-check` | `d-check: 748 Datei(en) geprüft, 0 Befund(e)` | frisch gelaufen |
| `baseline-verify` | `v6.18.1 OK — 54 Dateien` | frisch gelaufen |
| `groessenmarke` | `slice-mock-zuordnung.md: ok, Messung 2692 von Marke 5800` | frisch; Nachmessung `git diff --numstat 03c9420..HEAD` über die drei Pfade des Blocks: 2692; gegen die `basis` des Blocks (`2fc460a`) ebenfalls 2692 |
| `migration-check`, `schema-check`, `zeichen-check` | ok; `zeichen: ok, 1542 Datei(en)` | frisch gelaufen |
| `build`, `handbuch-check`, `katalog-check` | Stufen gelaufen | frisch gelaufen |
| `test` (`pruefung test`), `test-race` (`pruefung race`), `lint` | Docker-Schicht `CACHED` | **nicht neu ausgeführt**: der Baum ist inhaltsgleich zu dem Lauf, den die Rolle davor grün gefahren hat |

**Behauptung ohne eigene Ausführung, deshalb nachgefahren:** Weil `test`, `test-race` und `lint` aus dem Cache kamen, habe ich das Paket `mock` selbst unter dem Detektor gefahren (`go test -race -count=1 ./src/hexagon/services/mock/`, Container ohne Netz): grün, 100,0 % der Anweisungen (Abschnitt 3). Die Nullprobe jeder Liste des Mutationslaufs (unveränderter Stand, ganzes Paket) ist grün. `make lint` ist nicht neu gefahren; Lint-Eingaben (`.golangci.yml`, Quelle) sind inhaltsgleich zum Cache-Treffer, `.golangci.yml` im Diff unberührt.

## 2. Plan gegen Code

- **Dateien.** Der Block nennt `src/hexagon/services/mock`, `ports/inbound/mock.go`, `mock_test.go`. Der Diff `03c9420..HEAD` trägt außerdem Plan, Review-Bericht, `spec/spezifikation.md` (4 Zeilen: `SPEC-195` Nr. 6, `SPEC-198` Nr. 8; Architect) — der Plan nennt sie in §3.
- **Gate-Kette unberührt.** `git diff --stat 03c9420..HEAD` über `Makefile`, `Dockerfile`, `harness/`, `tools/`, `.a-check.yml`, `.golangci.yml`, `docs/user`, `src/hexagon/model` ist leer. Keine `//nolint` in `src/`.
- **Rohbytes (`AGENTS.md` §3.18).** `LC_ALL=C grep -nP '\\u[0-9A-Fa-f]{4}|\xEF\xBB\xBF|\xC2\xA0|\xE2\x80[\xA8\xA9]'` über `src/hexagon/services/mock/*.go`, `ports/inbound/mock*.go` und den Plan: kein Treffer. U+212A und U+017F in `zuordnung_test.go` entstehen aus Bytewerten (`[]byte{0xE2, 0x84, 0xAA}`, `{0xC5, 0xBF}`), gelesen in Zeile 252, 255, 258.
- **Zeilen Z01 bis Z24.** Alle benannten Tests existieren (`TestNewKonfig`, `TestNewAkzeptiert`, `TestNewKopiert`, `TestSatzWahl`, `TestQuelltextOhneVersion`, `TestArgumente`, `TestDienstWahl`, `TestZuordnung`, `TestFremdeKennungen`, `TestSessionHeaderOhneWirkung`, `TestURLUrsachen`, `TestKompatUndRest`, `TestMethode`, `TestRangfolge`, `TestWorkspace`, `TestService`, `TestMetadaten`, `TestMetadatenOptionen`, `TestMetadatenFehler`, `TestKennzeichenLeer`, `TestHeadMetadaten`, `TestPlatzhalterLesen`, `TestFehlerWert`, `TestGegenstandJeStelle`, `TestKontext`, `TestFakesZaehlenAlles`, `TestPaketOhneZustand`, `TestMockerFelder`, `TestKeinZustand`, `TestGleichzeitig`, `TestMockPort`, `TestKeinFehlerobjekt`). Die Zahlen der Mengen stimmen mit dem Code: M1 17, M2 60 (und 60 gegen alle fünf Roots), M3 78, M4 27 Paare, M5 11 Codes, M6 16, `AlleURLUrsachen()` 16, `AlleKompatUrsachen()` 6.
- **Mutations-Absatz.** Die Liste im Plan nennt 111 verschiedene Kennungen (nachgezählt), wie der Kopf des Absatzes sagt.
- **Vertrag gegen Code.** `SPEC-195` Nr. 1 (Konfig, `Codecs` nil, kein Port in `New`), Nr. 2 (Argumente), Nr. 3 (Reihenfolge a bis h im Code von `Mock`), Nr. 4 (Dienstwahl, Segmentgrenze, längster Root), Nr. 5 (Header), Nr. 6 (Abbildung über `outbound.Abbilden`), Nr. 7, 8, 9, 12 stimmen mit dem Code überein. `SPEC-198` Nr. 4/5/8 (`$metadata`, `HEAD`, leeres Kennzeichen) stimmt mit `metadaten.go` überein. Die Liste der 13 Codes in `spec195Nr10()` ist gleich der Liste in `SPEC-195` Nr. 10 (von mir abgeglichen); das Negativ (4) des Plans bleibt damit zutreffend, ist aber kein Fänger.
- **Aufrufstellen S1 bis S16.** Teil 1 trägt S1 bis S4, S12, S13 (Fänger `TestGegenstandJeStelle`, `TestKontext`, `TestMetadatenFehler`, `TestKennzeichenLeer`); S5 bis S11 trägt L43, S14 bis S16 L-Zeilen in Teil 2. `slice-mock-lesen` §1 nimmt die Sendungen an (L43, L44, L45, M4 bis M6 mit dem Rest; Mengen `AlleRessourcenArten`, `AlleCoverageStatus`).
- **R2/R5.** Der Nachzug von `SPEC-168` Nr. 10 und `SPEC-180` Nr. 6 steht noch aus (richtig: er gehört in den Merge-Commit) und ist in der DoD (Zeile *Doku-Update*), im Kopf (*Berührte Spec-Stellen*) und in §6 (R2, R5) genannt.
- **R7.** Teil 1 ohne `ODR-E6008`, `AlleRessourcenArten`, `AlleCoverageStatus`: `fehler.go` hat 37 Codes (nachgezählt), `harness/katalog-ausnahmen.txt` unverändert, `katalog-check` im Gate-Lauf grün. Die Spezifikation führt `ODR-E6008` schon (`SPEC-199`, Tabelle Zeile 819): die Lücke zwischen Spezifikation und Code gehört Teil 2 (Eintrag in `katalog-ausnahmen.txt` bei der Closure von Teil 2 oder Handbuch-Kapitel).
- **Rangfolge, Raster, Zusatztabelle.** Jede Zelle des Rasters nennt einen Ort; geprüft: die Orte tragen die Entscheidung (Nr. 4, 5, 8, 9, 2 und 11). Die drei Zeilen der Zusatztabelle haben Fänger (Z11/M4, Z20/M6, Z10 und Z11).

## 3. Akzeptierte Negative und Register `zweig-ohne-lauf-unter-dem-detektor`

Eigene Messung `go test -race -coverprofile` (Container ohne Netz, `--memory=3g`, `--cpus=4`):

- Ganzes Paket: **100,0 %** der Anweisungen.
- `TestGleichzeitig` allein: **91,9 %** (der Plan sagt 93,5 %). Nicht gelaufen sind **acht** Blöcke: sieben in `New` und `pruefeKonfig` (`mock.go:50`, `66`, `68`, `70`, `74`, `76`, `78`, wie im Plan) und **einer in `metadaten.go:27-29`** (der Zweig des leeren Kennzeichens, mit R6 neu). `metadaten` läuft darin zu 86,7 %, nicht zu 100 %.

Die Negative des Plans im Einzelnen:

- (1) Schreibzugriff in `New`/`pruefeKonfig`: Grund (kein gleichzeitiger Aufruf) und Folge (Composition Root baut einmal) stehen, tragen. Er deckt aber den Zweig in `metadaten.go` nicht.
- (2) `z22c`/`z22e` ohne Detektor 0/50: Grund und Folge stehen. Bestätigt durch mich: ohne Wirkung kein Wirkungs-Fänger; die Zusage hängt am Gate `test-race`.
- (3) Prüfgrenzen der unbeschränkten Mengen: Grund und Folge stehen.
- (4) Handkopie der 13 Codes: Grund und Folge stehen; ich habe die Kopie abgeglichen, sie ist gleich.
- (5) `>=` gleichwertig: zutreffend; `x03` (unten, V-7) zeigt, dass nur der Gleichstand gleichwertig ist, nicht jede Wahl.

## 4. Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Eigene Mutationen an Elementen, die weder Befund noch Plan nennen (Ergebnis `rot`, wenn nichts anderes steht; alle aus dem benannten Grund, geprüft an `test=` der Ergebniszeile):

| Befund | Menge (Elemente) | Eigene Mutation an einem nicht genannten Element | Ergebnis |
|---|---|---|---|
| F-1 Gegenstand je Stelle | 5 Stellen × 6 Fehler = 30 | `v04` S12 mit `Session`, `v05` S1 mit `Session`, `v06` S2 mit `Session`, `v07` S3 mit `Session`, `v08` S4 mit `Import` (der Plan nennt `Dienst`, `Workspace`, `Sonstiges`, an S4 auch `Session`) | alle 5 rot (`TestGegenstandJeStelle`) |
| F-2 kein Zustand | 5 Lesestellen + Feldmenge | im Plan benannt: `c01` bis `c06` decken alle fünf Lesestellen (Blob, Kennzeichen, URL-Leser, Workspace, Service); keine weitere Lesestelle des Dienstes bleibt (`dienstFuer` liest keinen Port). Eigene: `v14` und `v15` (zweiter Aufruf von `Workspaces.Get` / `ParseURL`) | beide rot |
| F-3 Kennungen des fremden Service | 12 Formen | `v12` (`Metadaten.Get` mit dem Root statt der Kennung) | rot |
| F-4 Unicode-Faltung | 3 Formen | Menge ASCII-Faltung gedeckt; `v03` (Rand-Leerzeichen im Namen akzeptiert) grün — **kein Befund**: das Wort „Rand-Leerzeichen im Namen“ steht nicht im Wortlaut von `SPEC-195` Nr. 5, und kein gültiger HTTP-Name trägt Leerzeichen | grün (außerhalb der Klasse) |
| R6 leeres Kennzeichen | 3 Formen × `GET`/`HEAD` | im Plan `z16d` bis `z16h`; eigene: `v50` siehe V-1 | — |
| Kontext je Stelle | 4 Formen × 4 Stellen | `v09` (S4 mit `context.Background()`), `v10` (S2), `v11` (S3); der Plan nennt S12 | alle 3 rot (`TestKontext`) |
| Konfig (M1) | 17 | `v29` (Satz mit `Schreiber` nil nicht geprüft), `v30` (leere Kennung nicht geprüft), `v31` (`MaxBytes` 1 abgelehnt), `v33` (Satz geteilt statt kopiert), `v34` (Standardsatz nur lesend) | alle 5 rot |
| Methode (Z10) | 8 × 3 Modi | `v18` (nur lesend als `ODR-E6001`) rot; `v02` siehe V-5 | |
| Fehlerwert (Z19) | 11 Codes | `v40` (letzter statt erster Status), `v41` (`ErrUngueltig` mit Code), `v42`, `v43`, `v45` | alle 5 rot; `v44` siehe V-6 |
| Dienstwahl (Z04/Z05) | 60 + 60 | `v46` (Pfad gleich Root nicht angenommen), `v47` (Root `/` mit leerem Pfad) | beide rot |
| Rangfolge (M4) | 27 Paare | `x01` (Zuordnung vor Dienstwahl) rot; Paare der Formen siehe V-3 | |
| `HEAD` | | `v25` (Status 204 bei `HEAD`) rot; `v26` Übersetzungsfehler (zählt nicht) | |

## 5. Befunde

### V-1 (MEDIUM) Der Zweig „Kennzeichen leer“ läuft nicht unter `TestGleichzeitig`; das Abdeckungsprofil im Plan ist überholt

Messung: Abschnitt 3 (91,9 %, acht Blöcke, `metadaten.go:27-29`). Mutation `v50` (Schreibzugriff `m.k.MaxBytes = m.k.MaxBytes + 1` im Zweig des leeren Kennzeichens, Klasse `nebenläufig`, 20 Prozesse mit Detektor, `-test.count=1`): **0/20 rot**, ergebnis `gruen`. Der Plan (Absatz *Abdeckungsprofil*) sagt, nicht gelaufen seien sieben Blöcke in `New` und `pruefeKonfig` und „jede andere Funktion von `Mock` und ihrer Helfer läuft darin zu 100 %“; das stimmte vor R6 (Commit `c82d4d7`), jetzt nicht mehr. Das Negativ (1) deckt den Zweig nicht, weil er nicht in `New` liegt und ein Aufruf gleichzeitig möglich ist. Das ist genau die Klasse des Registers `zweig-ohne-lauf-unter-dem-detektor` (§8 des Plans nennt sie). **Auflage:** ein Service mit leerem Kennzeichen in `TestGleichzeitig` (der Fake des Schreibers trägt ein einstellbares Kennzeichen, `kennzeichenFaelle()`), oder das Negativ mit Grund und Folge im Plan; in jedem Fall den Absatz *Abdeckungsprofil* mit der gemessenen Zahl nachziehen (`AGENTS.md` §3.9).

### V-2 (MEDIUM) Ein Schreibzugriff in einem Zweig, den jede Goroutine nur einmal erreicht, meldet der Detektor nicht in jedem Prozess

Der Plan: „Der Detektor fängt jeden Schreibzugriff in jedem Prozess“ (Raten von `z22c` bis `z22e`: 50/50). Das gilt für Zweige, die jeder Request erreicht. Eigene Messung mit demselben Schreibzugriff (`m.k.MaxBytes = m.k.MaxBytes + 1`) an anderen Stellen, `-test.count=1`, nur mit Detektor, Messender der Verifier mit `mutationslauf.sh`, `--cpus=4`, `last` 5,6 bis 9,8 (der Host war beschäftigt; Größenordnung):

| Mutation | Stelle | Requests je Goroutine, die sie erreichen | mit Detektor |
|---|---|---|---|
| `w6` | am Anfang von `Mock` (jeder Request) | 18 | 50/50 |
| `w5` | Zweig ungültiger Argumente | 1 | 50/50 |
| `w3` | kein Dienst (`ODR-E5005`) | 1 | 50/50 |
| `w4` | `Art` außerhalb der vier | 1 | 50/50 |
| `w2` | `HEAD`-Zweig | 1 | 47/50 (und 18/20 im ersten Lauf, `v52`) |
| `w1` | Service ohne Metadaten (`ODR-E6004`) | 1 | **36/50** (und 11/20 im ersten Lauf, `v51`) |
| `v50` | leeres Kennzeichen | 0 | 0/20 (V-1) |

`AGENTS.md` §3.10: eine Zusage über Nebenläufigkeit gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird; sonst Wirkungs-Fänger oder akzeptiertes Negativ mit Grund im Plan. Der Plan führt diese Zeilen nicht (Negativ (1) und (2) betreffen andere Fälle). **Auflage:** `TestGleichzeitig` ruft die selten erreichten Zweige (`ODR-E6004`, `HEAD`, leeres Kennzeichen) mehrfach je Goroutine auf (die Raten der Zweige in der Tabelle steigen mit der Zahl der Läufe), und der Plan trägt die Raten je Zweig mit Zahl der Prozesse und Messendem nach; bleibt eine Rate unter allen Prozessen, steht sie als Negativ mit Grund und Folge (Folge im Betrieb: ein ungeordneter Schreibzugriff in diesem Zweig kann das Gate `test-race` in einem Lauf überstehen).

### V-3 (LOW) M4: Stufen mit mehreren Formen führen im Paar nur eine

`SPEC-195` Nr. 3 (a) und (d) fassen mehrere Verstöße in einer Stufe: Argumente (`ctx` nil, leere Methode, `Laenge` -2, `Body` nil bei `Laenge` 1) und URL (Fehler des Lesers, `Ressource.Art` außerhalb der vier). `stufenM4()` baut „Argumente“ nur mit `Laenge` -2 und „URL“ nur mit einem `*URLFehler`. Mutationen:

- `v01`: die Prüfung der `Art` vor die der Methode (`a.Methode != "POST" && !slices.Contains(…)`): `POST` mit `Art` 0 liefert `ODR-E6002` statt `ODR-E9001`; **grün** (`TestRangfolge`, danach das ganze Paket).
- `v49`: `ctx` nil erst nach der Dienstwahl (`ctx == nil && a.Pfad != "/nirgends"`): **grün**.

Beide Mutationen sind eine Reihenfolge-Umkehr, die der Wortlaut „der erste Verstoß entscheidet“ verbietet. **Auflage:** in M4 je Stufe die Formen als Achse (Argumente 4, URL 2), die Paare als Schleife darüber.

### V-4 (LOW) M3: „mehr als ein Wert“ ist nur mit zwei Werten gefangen

`SPEC-195` Nr. 5: „ein leerer Wert und mehr als ein Wert … sind `ODR-E5003`“. Die Formen führen zwei Werte (`zwei gleiche`, `zwei verschiedene`, `zwei Schreibweisen`). Mutationen: `v19` (`len(w) == 0 || len(w) == 2 || w[0] == ""`, also drei Werte angenommen): **grün**. `v20` (leere Werte vor dem Zählen entfernt, also `["", "x"]` angenommen): **grün**. **Auflage:** eine Form mit drei Werten und eine mit zwei Werten, von denen einer leer ist, je mit allen sechs Werten.

### V-5 (LOW) Z10: die Schreibweise von `HEAD`

`SPEC-195` Nr. 7: „die Schreibweise zählt“; die Menge der Methoden führt `get` und `Get`, keine abweichende Schreibweise von `HEAD`. Mutation `v02` (`head` zusätzlich angenommen): **grün**. **Auflage:** `head` und `Head` in die Menge.

### V-6 (LOW) „ohne Ursache“ für `ODR-E5003` und `ODR-E5005` (Dienstwahl)

`SPEC-195` Nr. 10: „`ODR-E5003`, `ODR-E5005` (Nr. 4), … bildet der Dienst ohne Tabelle, jeweils ohne Ursache“. Gefangen ist es für `ODR-E5004`, `ODR-E6002`, `ODR-E6004`, `ODR-E9001` (Art, Kennzeichen), `ODR-E6001` (Platzhalter). Mutationen `x02` (`ODR-E5003` mit `ErrUngueltig` als Ursache) und `x04` (`ODR-E5005` der Dienstwahl mit Ursache): beide **grün**. **Auflage:** `TestFehlerWert` prüft über die Fehlerzeilen, dass die Codes ohne Tabelle `Ursache == nil` tragen (Menge aus dem Wortlaut der Nummer).

### V-7 (LOW) Z04: die Reihenfolge von `Dienste`

`TestDienstWahl` baut die fünf Roots immer in derselben Reihenfolge (`/`, `/a`, `/a/b`, `/ab`, `/a%2Fb`); in ihr ist der letzte Treffer immer der längste. Mutation `x03` („der letzte Treffer gewinnt“, `if ok {`): `TestDienstWahl` bleibt **grün**; rot wird nur `TestGleichzeitig` („Klassen der Workspaces nicht unterscheidbar“), weil dort der Root `/` zuletzt steht — Ergebnis `rot-paket`, nicht aus dem benannten Grund. **Auflage:** die Reihenfolge der Dienste (aufsteigend, absteigend, umgekehrt zum Treffer) als Achse in M2.

### INFO

- **Merge.** `main` steht 32 Commits vor dem Merge-Basis-Commit des Zweigs (`e22b3c7`); der Zweig nimmt `main` per Merge auf (Entscheidung 5 in §3). Im Merge-Commit stehen der Nachzug R2 (`SPEC-168` Nr. 10) und R5 (`SPEC-180` Nr. 6).
- **Mutation `v48`** war von mir falsch gebaut (gleichwertig) und zählt nicht; `x01` ist die richtige Form und rot.
- **Mutation `v26`** ist ein Übersetzungsfehler und zählt nicht.
- **Zählung meiner Mutationen:** 47 (36 in `liste2.json`, 10 in `liste3.json`, 1 gestrichen): 30 rot, 1 `rot-paket` (`x03`), 11 grün (davon `v48` gleichwertig, `v03` außerhalb des Wortlauts; die neun übrigen sind V-1 bis V-7), 4 flatternd (`v51`, `v52`, `w1`, `w2`), 1 Übersetzungsfehler.

## 6. DoD (Punkt für Punkt)

| Punkt der DoD | Stand |
|---|---|
| Port und Konfig | erfüllt (Z01 bis Z03, Z19 bis Z21, Z23) bis auf V-3, V-6 |
| Zuordnung und `$metadata` | erfüllt (Z04 bis Z18) bis auf V-4, V-5, V-7 |
| `make gates` grün | belegt (Abschnitt 1) |
| Review durchgeführt | Report liegt vor (`2026-11-07-slice-mock-zuordnung-review.md`) |
| Doku-Update (R2, R5) | im Merge-Commit; in DoD, Kopf, §6 genannt |
| Katalog / Handbuch | Teil 1 bringt keinen Code; §7 muss „Binary unverändert“ sagen (bei der Closure) |
| Closure-Notiz, Register, Risiken mit Ausgang, Paarungen | bei der Closure (§7 ist leer) |
