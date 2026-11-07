# Verifikation slice-mock-lesen (Teil 2: Lesen aus dem Entity-State mit Coverage-Prüfung)

**Rolle:** Verifier (`Bauen wir es richtig?`, gegen Plan, DoD und Spezifikation). **Datum:** 2026-11-07.
**Gegenstand:** Zweig `slice-mock-lesen`, HEAD `4568ab8` (Fix-Runde nach dem Review `774249f`), Basis `473fe8f`, Abzweig von `main` bei `8d498c6`.
**Eingang:** DoD-Bestätigung und Bericht des Implementers in §3 des Plans (Tabelle L01 bis L46, Mengen M4 bis M20, Absätze *Mutationen und Raten* und (4a)), Review `docs/reviews/2026-11-07-slice-mock-lesen-review.md` (F-1 bis F-8).
**Bezug:** `LH-FA-06`, `LH-FA-08`, `LH-FA-11`, `LH-QA-05`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
**Spezifikation:** `SPEC-195` Nr. 10 bis 12, `SPEC-196` bis `SPEC-199`, `SPEC-109` Nr. 2 und 5, `SPEC-230` Nr. 8, `SPEC-190` und `SPEC-191` (Pflichten), `SPEC-128` Nr. 9.

## Urteil

**Bestanden mit Auflagen.** Kein HIGH und kein Widerspruch des Produktcodes zum Vertrag: die Sätze von `SPEC-196` bis `SPEC-198` stimmen Satz für Satz mit `lesen.go`, `deuten.go` und `schluessel.go` überein, `make gates` ist grün, die Sendungen, die Rohbytes, `ODR-E6008` und der Katalog halten. Die Fixes F-1 bis F-6 decken ihre Klasse. Gefunden habe ich Lücken der **Fänger** an Elementen, die der Plan und die Fix-Runde nicht nannten: drei MEDIUM (V-1 bis V-3), fünf LOW (V-4 bis V-8) und Plan-Reste (V-9). Neun von 53 Stichproben-Mutationen im Dienst blieben im ganzen Paket grün. Keine verlangt Produktcode; alle verlangen einen Fall in einem Test. Auflagen stehen am Ende.

## Eigene Sensoren (nicht die Behauptung, die Belege)

| Sensor | Ergebnis |
|---|---|
| `make gates` (einmal, eigenes Log `scratchpad/verif-ml/gates.log`, unter `flock /tmp/odata-recorder-heavy.lock`) | Exit 0. `a-check`, `d-check`, `lint-profil`, `lint`, `test`, `test-race`, `build`, `katalog-check`, `schema-check`, `migration-check`, `zeichen-check`, `baseline-verify` in der Kette; `groessenmarke slice-mock-lesen.md: ok, Messung 4159 von Marke 7300` |
| Größenmarke | 4159 von 7300, Auslastung **0,57** (unter 0,60): `SPEC-128` Nr. 9 verlangt die Meldung an den Architect in der Closure (Auflage A-3) |
| Abdeckung unter dem Detektor, selbst gemessen (`go test -race -count=1 -coverprofile`, Image `odata-recorder:race`, Baumkopie per `docker cp`, `--network none --memory=4g --cpus=4`) | `TestZweigeGleichzeitig` allein: **82,2 %**, wie im Plan. Zu 100 % laufen die dort genannten Funktionen; nicht gelaufene Blöcke wie im Plan benannt, dazu ein unbenannter (V-8). Ganzes Paket: **98,7 %** (der Plan nennt 98,4 %, V-9). Im Paket ungelaufen nur: `zerlege` hinter `}`, `abgeleitet` bei unauflösbarem `BaseType`, `entity` bei mehr als einem Ergebnis, `literalPasst` bei unbekannter Property und bei Typ außerhalb der acht (alles Verteidigungszweige) |
| Zeitgrenze der Nullprobe von `src/bootstrap` | Auf meiner Baumkopie läuft die Nullprobe mit den drei Ende-zu-Ende-Tests in 9 Sekunden (Last 20 bis 30). Die 117 Sekunden des Plans konnte ich nicht nachstellen; die Handläufe selbst sind ehrlich (V-10) |
| Rohbytes (`AGENTS.md` §3.18): `LC_ALL=C grep -P` auf Backslash-u-Folge, BOM, NBSP, Zeilentrenner, Nullbreiten- und Steuerzeichen über alle 29 Dateien von `8d498c6..HEAD` unter `src/` und `harness/` | kein Treffer. Das Kelvin-Zeichen steht als `"\xe2\x84\xaaunden"` in `lesen_ablauf_test.go:124`, das kyrillische `е` als `"Acc\xd0\xb5pt"` in `lesen_ablauf_test.go:329`; die Nicht-ASCII-Treffer sind Umlaute in Kommentaren. Dieser Report ist gegengelesen (siehe Ende) |
| `//nolint`, Gate-Kette | keine Direktive im Diff; `git diff --name-only 8d498c6..HEAD` nennt weder `.golangci.yml` noch `.a-check.yml` noch `Makefile`, `Dockerfile`, `harness/mk/`, `harness/README.md`, `tools/` |
| `ODR-E6008` | 38. Code in `model/fehler.go` (404, `ZielMock`), Zeile in `harness/katalog-ausnahmen.txt` mit Grund, `SPEC-199` in §4; `make katalog-check` im Gate-Lauf grün. Die Stellen 37 auf 38: `fehler_test.go` (Soll 38), `audit_test.go` (Orakel `codeTexte` mit `ODR-E6008`, 38, `66-38`, 28), `regeln_test.go` (`alleCodes` mit `ODR-E6008`), `httpodata/fehler_test.go` (Paare gegen `model.AlleCodes()`), `telemetry_test.go` zählt gegen `model.AlleCodes()` (von `main`); keine feste 37 mehr im Code |
| Merge-Stand von `main` | `main` liegt bei `dd38c4b` (`slice-telemetry-handler` gemerged, der Zweig hat ihn noch nicht). `git merge-tree` zeigt **einen** Konflikt: die Historie-Tabelle in `spec/spezifikation.md` (zwei neue Zeilen an derselben Stelle, beide behalten). Kein Konflikt in `src/` |

## Plan gegen Code

- Dateien: `services/mock/lesen.go` (Ablauf, Coverage, Antwort), `deuten.go` (Payload, Maske, Pflicht 3 g), `schluessel.go`, die Verdrahtung `holeWorkspace` in `mock.go`, `model/fehler.go` und `ressource_mengen.go` (`AlleRessourcenArten()`, `AlleCoverageStatus()`), `src/bootstrap/mock_lesen_ende_zu_ende_test.go`, acht `lesen_*_test.go`, `fakes_test.go`: wie in §3. Der Platzhalter von Teil 1 (Z18) entfällt; `TestPlatzhalterLesen` ist weg.
- Die Abweichung des Implementers (Pflichten 3 e und 3 f prüft `model.FuehreAus`, der Dienst bildet `ODR-E9001`) ist mit dem Wortlaut von `SPEC-196` Nr. 4 c und d vereinbar („`FuehreAus` liefert `ErrUngueltig`“); L09 und L10 fangen Mutationen des Dienstes (V11 unten rot).
- Sätze gegen Code, ohne Befund: Art außerhalb der vier (`Mock`, `AlleRessourcenArten`), Modell und Strom (Nr. 2, `ladeModell` schließt auf jedem Weg), Set vor Coverage, Typ nicht aufgelöst, Deuten (Nr. 3 a bis d), Maske (Nr. 4 a), Anzahl allein mit `Filter` (Nr. 4 b), `redigiert` an jeder Stelle des Baums (die sechs Knotenarten von `filter_ast.go`, zwei tragen Namen), Lookup (Nr. 7 a bis e, Tombstone ohne Coverage, `ErrNichtGefunden` über die Coverage), `ListSet` und `Get` je Antwort genau einmal, Header der Coverage nur auf `200` von Collection und Anzahl, `Wurzel` aus `PublicBaseURL`, `Accept` über `werteVon`, Darstellung und leeres Kennzeichen (Nr. 8).

## Tabelle Zusage · Fänger gegen den Vertrag

Jeder Satz von `SPEC-196` Nr. 1 bis 10, `SPEC-197` Nr. 1 bis 6 und `SPEC-198` Nr. 1 bis 8 hat eine Zeile L01 bis L46 oder ein benanntes Negativ (L39, F-8). Ich habe die Zeilen mit eigenen Mutationen aufgelöst, **an Elementen, die Plan und Fix-Runde nicht nannten**. Die Fixes F-1 bis F-6 sind gegen die Klasse aufgelöst.

### Eigene Mutationen (70 Läufe)

Liste 1: 53 Mutationen der Klasse *andere* (`mutationslauf.sh`, Image `odata-recorder:test`, `--memory=4g`, Last 24 bis 27 von 4 CPUs): **43 rot** (3 davon `rot-paket`, ein anderer Test des Pakets wird rot), **10 grün**. Liste 2: 9 Mutationen der Klasse *nebenläufig* (`TestZweigeGleichzeitig`, Image `odata-recorder:race`, 20 Prozesse mit Detektor, `-test.count=1`): 5 rot (20 von 20), 4 grün (0 von 20). Liste 3: 2 Mutationen von `TestLesesicht`. Handläufe: 6 Mutationen in `src/bootstrap` auf einer Baumkopie. Die neun grünen der Liste 1 habe ich zusätzlich gegen das **ganze** Paket `services/mock` gefahren (`go test ./...` des Pakets): alle neun bleiben grün, kein anderer Test fängt sie.

| Klasse (Befund) | Mutation an einem nicht genannten Element | Ergebnis |
|---|---|---|
| F-1 Typen des Sets | Wurzeltyp ohne Bezug zum Set angenommen (`c.BaseType == ""`) | rot, `TestTypenDesSets` |
| F-2 nicht unterstützte Property | `Unterstuetzt()` ersetzt durch Ausschluss eines Typnamens | rot |
| F-3 bis F-5 | Gegenstand `Workspace` an S4, S5, S6, S9, S11 (je eine Stelle) | alle rot (`TestPortFehler`, `TestPortFehlerBlob`) |
| Kontext je Stelle (S4 bis S11) | `context.Background()` statt `ctx` an allen acht Stellen | alle rot (`TestKontextLesen`, S4 durch `TestModellLesen`) |
| Pflicht 3 e/3 f | leerer `Schluessel` in der Anzahl weggelassen | rot |
| Pflicht 3 g | Verknüpfung links, zweites Element von `$orderby`, letzte Entity ungeprüft | alle rot |
| Pflicht 2 | `Select`, `Skip`, `Top` an `FuehreAus` der Anzahl | rot; `Skip` und `Top` fängt `TestOptionenJeArt` (L01), nicht `TestAnzahlNurFilter` (L08) |
| Coverage-Rangfolge | strikt nur bei `unknown`; Prüfung nach `ListSet`; Lookup `partial` strikt als `404`; Schlüssel nach `$filter` | alle rot |
| Lookup/Aufrufzahl | `Get` und `ListSet` mit Baseline-Kennung; `Get` zweimal; `ListSet` zweimal in der Collection | alle rot |
| Lesesicht (L38) | zweiter `ListSet` in Anzahl und in Collection | rot in **20 von 20 mit und 20 von 20 ohne Detektor** (Last 24, Messender der Verifier) |
| Darstellung | `Accept` nil an Entity und Collection; `Auswahl` nil an der Entity; Kennzeichen-Wert oder -Name ungeprüft; `HEAD` ohne Coverage-Header; `Count.Wert` aus `len` statt `Anzahl` | alle rot |
| Schlüssel | `Edm.Guid` nimmt Zeichenkette | rot |
| Ende-zu-Ende (Handläufe, Baumkopie) | Join `;` statt `,`; `Top`, `Filter`, `Skip`, `Select` der Collection entfernt | alle rot in `TestSchluesselEndeZuEnde`, `TestAbfrageEndeZuEnde` |
| Nebenläufig, M14 und M19 | Schreibzugriff auf die Konfig in `ohneTreffer` (E6003), Anzahl-Fehler, Tombstone, `ListSet`-Fehler, Lookup-Treffer | 20 von 20 rot |

**Grün (Fänger-Lücken und ein Negativ):**

| Nr. | Mutation | Ergebnis im ganzen Paket |
|---|---|---|
| G1 | `@odata.type` ohne Prüfung auf `#` (`xN.Special` wird angenommen) | grün (V-4) |
| G2 | `ohneRedigiert` bereinigt nur die erste Entity | grün (V-1) |
| G3 | `$select=*,B` zählt als „nur B“ (`sel.Alle` mit Namen) | grün (V-2) |
| G4 | Entity: `Wurzel` ohne Abschneiden des `/` | grün (V-5) |
| G5 | Boolean aus JSON `1` | grün (V-6) |
| G6 | Schlüssel `Edm.Boolean` nimmt `LiteralZahl` | grün (V-3) |
| G7 | Schlüssel `Edm.Int64` mit 63 Bit statt 64 | grün (V-3) |
| G8 | ein unbenannter Teil bei zwei Schlüsselteilen erlaubt (`>= 1`) | grün (V-3) |
| G9 | `DeadlineExceeded` nicht als Kontextfehler | grün (V-7) |
| G10 | Gegenstand `Workspace` an S14 (`FuehreAus`) | grün: das akzeptierte Negativ F-8 ist gleichwertig, bestätigt |
| N1 | Schreibzugriff im Zweig „Coverage: `ErrNichtGefunden` ist `unknown`“ | **0 von 20** (V-8) |
| N3, N4, N7 | Schreibzugriff in `ohneTreffer` (Fehler der Coverage), bei ungültigem Status, bei „Typ nicht aufgelöst“ | 0 von 20: die im Plan benannten Negative, **ehrlich** |

## Befunde

**V-1 (MEDIUM) `redigiert` wird nur an der ersten Entity entfernt: kein Fänger.** `SPEC-196` Nr. 4 e: die Werte der Art `redigiert` entfernt der Dienst „aus den Entities des Ergebnisses (nie ausgeliefert, `LH-FA-14.a`)“. `TestRedigiertNichtAusgeliefert` (L12) führt je Lauf **eine** Entity (Typ 6 × Art 2); die Mutation G2 bleibt im ganzen Paket grün. Gemeint ist die Menge „jede Entity der Collection“. Wirkung im Betrieb: der echte Schreiber lehnt den Wert ab (500), er wird nicht ausgeliefert, aber die Zusage des Dienstes ist ungefangen. Fix: Fall mit drei Entities, der Wert in der zweiten und der dritten (Schleife über alle Positionen).

**V-2 (MEDIUM) `$select=*,B` am Lookup ohne Fänger.** `SPEC-196` Nr. 7 e: bei `Select` gleich nil **oder `Alle` wahr** zählen alle unterstützten Properties als angefordert. M12 führt fünf Formen (kein, `*`, bekannte Namen, unbekannter Name, nur der Schlüssel), nicht `*` mit Namen; L28 führt `*,B` nur für `Auswahl`. G3 bleibt grün. Fix: die Form `*,B` (und `*,Unbekannt`) in M12, mit `partial`-Eintrag, der B kennt und anderes nicht.

**V-3 (MEDIUM) Schlüssel: drei Elemente von `SPEC-196` Nr. 7 a ohne Fänger.** M10 (Verstoß 8 × Typ 8, 42 erreichbare Zellen) führt nicht jede Zelle der Literal-Art × `EdmType`: G6 (`Edm.Boolean` mit `LiteralZahl`) bleibt grün, `Edm.Guid` mit Zeichenkette ist rot, die Matrix ist also lückenhaft statt vollständig. G7: die Grenze von `Edm.Int64` liegt nur bei „nicht ganzzahlig“, nicht bei 2 hoch 63 minus 1 und 2 hoch 63 (`AGENTS.md` §3.15, beide Seiten jeder Grenze; der Plan hat nur die Grenze von `Edm.Int32`). G8: „ein unbenannter Teil nur bei einem Schlüsselteil“, die Fixture „unbenannt bei zwei Teilen“ besteht aus zwei unbenannten Teilen, die schon an der Doppelung scheitern (`AGENTS.md` §3.15: eine Fixture gleicht nie dem Wert der Mutation); ein unbenannter plus ein benannter Teil fehlt. Fix: Schleife über 8 Literal-Arten × 8 Typen (Orakel von Hand), Int64 an beiden Seiten der Grenze, ein unbenannter plus ein benannter Teil.

**V-4 (LOW) `@odata.type` ohne `#`.** Der Fall „ohne #“ in `TestPayloadVerstoesse` nutzt `N.Special`; ohne die Prüfung schneidet `name[1:]` das `N` ab und der Typ ist unbekannt, die Fixture gleicht also dem Wert der Mutation. G1 (`xN.Special`) bleibt grün. Fix: ein Fall mit beliebigem ersten Zeichen statt `#`.

**V-5 (LOW) `Wurzel` je Art.** `TestWurzel` (L16: Root 3 × Basis 2 × Header 3) liest `letzteCollection().Wurzel`; die Entity trägt nur der Ende-zu-Ende-Fall (Root `/svc`). G4 (Root `/` und `/a/b` an der Entity) bleibt grün. Fix: die Art als Achse von L16 (Collection, Entity).

**V-6 (LOW) JSON-Typen je Property.** M8 führt je Typ einen falschen JSON-Typ (String als Zahl, Zahl als String ...), nicht die Matrix 8 Typen × JSON-Arten (Zahl, String, Boolean, Array, Objekt). G5 (Boolean aus der Zahl `1`) bleibt grün. Fix: Schleife über die Matrix.

**V-7 (LOW) `kontextFehler` halb.** Die vier Formen von M6 (Canceled, DeadlineExceeded, verpackt, `errors.Join`) fahren `errors.Join` mit `Canceled`; G9 (`DeadlineExceeded` aus `kontextFehler` entfernt) bleibt grün, weil `kontextFehler` nur neben `ErrNichtGefunden` wirkt. Fix: `Join(ErrNichtGefunden, DeadlineExceeded)` an S6 und S8.

**V-8 (LOW) Zweig ohne Lauf unter dem Detektor nicht benannt.** Der Plan benennt unter den nicht gelaufenen Blöcken `coverage` mit „Fehler und ungültiger Status“. Nicht gelaufen ist auch der Zweig „kein Eintrag ist `unknown`“ (`nichtGefunden`, N1: 0 von 20); M14 erreicht `unknown` nur über einen gespeicherten Status. Der Zweig ist ein Fall von `SPEC-197` Nr. 1 und läuft im Paket (L18). Fix: im Absatz (5) nennen oder in die Gleichzeitigkeit aufnehmen (Register `zweig-ohne-lauf-unter-dem-detektor`, dieser Slice ist ein weiterer Beleg: ein Zweig, den der Plan nicht nannte).

**V-9 (LOW, Plan) Zahlen nach der Fix-Runde nicht nachgezogen (`AGENTS.md` §3.9).** §6 *Stand des Handoffs* (c) nennt „3893 von Marke 7300, Auslastung 0,53“, die Messung ist 4159 (0,57); Absatz (5) nennt 98,4 % für das Paket, gemessen 98,7 %. Die Auslastung 0,57 liegt unter 0,60: Meldung an den Architect nach `SPEC-128` Nr. 9, in die Closure.

**V-10 (INFO) Handnachweise von `src/bootstrap`.** Der Plan listet für L33 bis L35 fünf Mutationen, nennt aber „2 von Hand“. Ich habe fünf unabhängige Handläufe gemacht (Trenner, `Top`, `Filter`, `Skip`, `Select`), alle rot in dem Test, der sie meint. Die Mutation „`Wurzel` aus einer fremden Basis“ (L35) habe ich nicht gefahren; der Test prüft `evil` und `localhost` im Body und die Basis `http://recorder.test/svc/` (aus dem Quelltext gelesen). Die Nullprobe lief bei mir in 9 Sekunden, wohl weil ich mit `-run` auf die drei Tests beschränkte; die Grenze des Werkzeugs (`SPEC-210` Nr. 5 c) bleibt eine Sache des Planner.

**V-11 (INFO) Fänger-Zuordnung.** L08 zählt „5 Optionen“, die Mutationen `Skip` und `Top` an der Anzahl fängt aber `TestOptionenJeArt` (L01), nicht `TestAnzahlNurFilter`; `Select` und `OrderBy` fängt L08. Kein Verlust, nur die Tabelle.

## Auflagen

- **A-1 (Implementer, vor dem Merge nach `main`):** V-1 bis V-3 mit je einem Fall als Schleife über ihre Menge schließen und V-4 bis V-7 ebenso, oder jedes als akzeptiertes Negativ mit Grund und Folge in §3 eintragen; danach die Mutationen G1 bis G9 rot sehen. Produktcode bleibt unverändert. V-8 bis V-9 im Plan nachziehen.
- **A-2 (Merge):** `main` (bei `dd38c4b`) in den Zweig nehmen; Konflikt allein in der Historie von `spec/spezifikation.md` (beide Zeilen behalten); danach `make gates` erneut.
- **A-3 (Planner/Architect, Closure):** Auslastung 0,57 nach `SPEC-128` Nr. 9 an den Architect melden; im Lerneintrag nennen: „Fixture gleicht dem Wert der Mutation“ (V-3 G8, V-4) und „Menge ohne ihre Positionen“ (V-1, V-2) sind die Klassen dieser Runde.

## Gegenlesen des Reports

Dieser Report enthält keine Folge aus Backslash, `u` und vier Hex-Ziffern; die Escape-Folgen `\xe2\x84\xaa` und `\xd0\xb5` stehen als Text. Gegengelesen: `docs/reviews/2026-11-07-slice-mock-lesen-verifikation.md` mit `LC_ALL=C grep -P` auf BOM, NBSP, Zeilentrenner, Nullbreiten- und Steuerzeichen (kein Treffer).
