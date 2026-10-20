# Review: slice-fehlermodell-konfiguration

* Datum: 2026-10-20
* Rolle: reviewer (frischer Kontext, kein Self-Review)
* Review-Art: Code (gegen Plan, Spezifikation und Hard Rules)
* Gegenstand: `git diff 89db02c..HEAD` (Commits 908f4ca, c168c76, 5b17da7)
* Plan: `slice-fehlermodell-konfiguration` (§3 Zusage · Fänger, §6 Raster)
* Verträge: `spec/spezifikation.md` SPEC-109 und SPEC-110 (Stand Commit 3c359b1), §3 Defaults, §4 Fehler-Codes, SPEC-006
* Maßstab: Plan, `AGENTS.md` §3.7, §3.9, §3.10, §3.11, §3.12, §3.14, §3.15, §3.16, Reviewer-Skill. Nicht geprüft: DoD (Verifier).
* Methode: Mutationen im Arbeitsbaum, je Lauf `make test` (Docker, netzlos), danach `git checkout` der Datei; `git status` am Ende leer. Etwa 100 Mutationen, alle Zeilen der Tabelle *Zusage · Fänger* und alle Zeilen des Rasters aufgelöst (siehe unten).

## Zusammenfassung

Keine HIGH. Drei MEDIUM, zwei LOW, vier INFO.
Das Fehlermodell, die Abbildung und der SQLite-Adapter sind gut gefangen; die Konfiguration ebenfalls bis auf die unten genannten Stellen.
Die Code-Tabelle des Modells stimmt mit §4 und SPEC-109 Nr. 2 überein: alle 33 Codes (`ODR-E2001` bis `ODR-E9001`, einschließlich `E7011`, `E7012`, `E9001`), Status und Zielbereich gleich, `E6004` mit `422` und `0`, `E4001` mit `502` und `504`.
Der SQLite-Adapter ändert sein Verhalten nicht: Namen, Texte und Identität der vier Variablen bleiben, `errors.Is` trifft zusätzlich das Ziel; die Tests des Adapters sind unverändert (Diff zeigt nur `sqlite.go` und die neue `fehler_test.go`).

## Findings

### F-1 MEDIUM — Default von `DATA_DIR` ohne Fänger
* kategorie: MEDIUM
* quelle: AGENTS.md §3.15 (Zusage ohne Fänger), Plan §3 Zeile „Defaults aus §3 für jeden Schlüssel“, SPEC-077
* pfad: `src/bootstrap/konfig_werte.go` · `case !ok:\n\t\treturn "data"`
* befund: Mutation `return "data"` → `return "daten"` (beide Stellen) lässt `make test` grün. `TestKonfigDefaults` setzt über `basis(t)` immer `DATA_DIR` und vergleicht nur diesen Wert; der Default `data` aus §3 wird von keinem Test geprüft, obwohl die Plan-Zeile „alle Werte gegen §3“ zusagt. Ein falscher Default legte beim Start still ein falsches Verzeichnis an.
* verifizierbar: ja (`make test` mit der Mutation)
* klasse: Default aus §3 ohne Fänger

### F-2 MEDIUM — Randformen im Code entschieden; Plan §6 sagt „kein Test legt eine Antwort fest“, einige Tests tun es
* kategorie: MEDIUM
* quelle: AGENTS.md §3.12, §3.16 („eine im Code entschiedene Zelle ist ein MEDIUM“)
* pfad: Plan §6 · „kein Test legt eine dieser Antworten fest“; `src/bootstrap/konfig_werte.go` · `func headerMenge`; `src/bootstrap/konfig_test.go` · `"Odata-Version"`, `"X-Api-Key"`, `"Etag"`
* befund: Der Plan gibt acht Randformen an den Architect zurück und behauptet, kein Test lege sie fest. Gemessen: `TestKonfigDefaults` und `TestKonfigListenErgaenzen` legen die kanonische Schreibweise der Header fest (`Odata-Version` statt `OData-Version` aus §3 SPEC-061, `X-Api-Key` statt `X-API-Key` aus der Default-Liste der Spezifikation, `Etag`), und `TestKonfigNone` legt fest, dass `REDACT_HEADERS=none` ein Header-Name `None` ist. Die Mutation `none` bei `MATCH_VARY_HEADERS` entfernen wird dagegen von `TestKonfigNone` gefangen, die Randformen selbst stehen weiter weder in SPEC-110 noch im Raster.
* Weitere im Code entschiedene, nicht im Plan genannte Formen (aus dem Code gelesen, nicht per Test belegt): ein ungültiger `MODE` lässt `record` falsch auswerten, sodass `UPSTREAM_URL`-Pflicht dann nicht gemeldet wird; eine formal ungültige `UPSTREAM_URL` im Modus `replay` ist ein Verstoß, obwohl SPEC-110 Nr. 9 sie dort „zulässig und ohne Wirkung“ nennt.
* Mutation gefahren: `null` als `id` in `SERVICES_FILE` wie gültig behandeln (`return "x"`) bleibt grün; die Antwort „wie fehlend“ ist damit nur im Code, nicht in Vertrag oder Test.
* verifizierbar: ja für die Test-Festlegung (Header-Schreibweise in `konfig_test.go`), nein für die Vertragslücke
* klasse: Randform im Code statt im Vertrag entschieden

### F-3 MEDIUM — Rückführungs-Marke des Plans überschritten, Plan nicht nachgezogen
* kategorie: MEDIUM
* quelle: AGENTS.md §3.14, §3.9; Plan §4 „über 800 Zeilen (Summe aus `git diff --numstat` …)“
* pfad: Plan §4 · „die Schätzung liegt bei rund 600 Zeilen“
* befund: `git diff --numstat 89db02c..HEAD` über die Pfade aus §3 ergibt 2081 (Einfügungen plus Löschungen, Tests eingeschlossen), der Produktcode allein 1053. Die Marke 800 ist deutlich überschritten, Rückführung nach `next` ist nicht erfolgt und §4 nennt die Überschreitung nicht; ein Review dieses Umfangs ist die Folge, die die Marke verhindern sollte.
* verifizierbar: ja (`git diff --numstat`)
* klasse: Rückführungs-Marke überschritten ohne Plan-Nachzug

### F-4 LOW — Kommentar im Test sagt zu, was der Test nicht prüft
* kategorie: LOW (Konventions-Anker: AGENTS.md §3.7, §3.11)
* quelle: AGENTS.md §3.11
* pfad: `src/bootstrap/konfig_test.go` · `// Ohne DATA_DIR gilt `data`, relativ zum Arbeitsverzeichnis`
* befund: Der Kommentar nennt den Default `data` und Modus `replay`; der folgende Code prüft allein, dass `MODE=record` ohne `UPSTREAM_URL` einen Verstoß liefert. Der Satz ist zudem ein abgebrochener Gedanke („Modus replay genügt nicht, daher Verstoß erzwingen“).
* verifizierbar: nein
* klasse: Kommentar sagt Prüfung zu, die fehlt

### F-5 LOW — Meldung sortiert die ganze Zeile, der Vertrag sagt: nach Schlüsselname
* kategorie: LOW
* quelle: SPEC-110 Nr. 12, Maintainability
* pfad: `src/bootstrap/konfig.go` · `slices.Sort(zeilen)`; `src/bootstrap/konfig_test.go` · `slices.IsSorted(zeilen)`
* befund: Sortiert wird die Zeichenfolge `ODR-E2001 <Schlüssel>: <Grund>`, der Test prüft dieselbe Sortierung. Für die heutigen Schlüssel fällt das mit der Sortierung nach Schlüsselname zusammen; ein Schlüssel, der Präfix eines anderen ist und danach ein Zeichen unter `:` trägt, würde abweichen. Die Fänger-Zeile nennt „sortiert“ ohne Bezug auf den Schlüssel.
* verifizierbar: ja
* klasse: Test spiegelt die Implementierung statt den Vertrag

### F-6 INFO — Überlebende Mutationen ohne Befund
* `LimitReader` bei `maxServicesFile+1` macht die Prüfung `len(roh) > maxServicesFile` für die Verletzung redundant: die Mutation `false` bleibt grün, weil die gekürzte Datei als kein JSON-Array auch ein `ODR-E2001` ergibt. Der Test „über 1 MiB“ prüft damit das Ergebnis, nicht den Grund.
* Die Mutation, `Pflichtfeld fehlt` für leeres `id` abzuschalten, bleibt grün (gleichwertig: die Formprüfung meldet denselben Schlüssel).
* Die Mutation, bei `LINK_REWRITE` zusätzlich `rewrite` zuzulassen, bleibt grün; `TestKonfigEnum` probiert für den Schlüssel nur `absolute`. Zuständig: Implementer, kein Vertragsbruch.
* verifizierbar: ja
* klasse: Fall prüft Ergebnis statt Grund

### F-7 INFO — Abgleich der Test-Tabelle mit der Spezifikation (Plan §3, Zeile 2)
`TestCodeTabelle` führt den Spiegel von SPEC-109 Nr. 2; der Reviewer hat ihn gegen die Spezifikation gelesen, alle 33 Zeilen stimmen. Die feste `message` von `ODR-E5001` stimmt mit dem Beispiel in SPEC-006.

### F-8 INFO — `REDACT_HEADERS`-Default aus der Spezifikation
Die Default-Liste im Code (`Authorization`, `Proxy-Authorization`, `Cookie`, `Set-Cookie`, `X-API-Key`, `X-Auth-Token`, `X-CSRF-Token`) stimmt mit der Spezifikation (Default-Redaction der Header) überein; die Schreibweise siehe F-2.

### F-9 INFO — Verzahnung mit Folge-Slices
Die Abbildung kennt keine Zeile für den Upstream-Transport (SPEC-109 „Ungeprüft“); jeder Slice, der einen Code auslöst, trägt seine Zeile. Zuständig: Planner bei der Zuweisung (AGENTS.md §3.13), kein Befund am Diff.

## Auflösung der Tabelle *Zusage · Fänger* (§3 des Plans)

Jede Mutation wurde gefahren; „rot“ nennt den ausgelösten Test.

| Zeile (Zusage) | Mutation | Ergebnis |
|---|---|---|
| Fehlerobjekt nur `code`, `message`, `target` | Feld `Status` im Objekt ergänzt | rot, `TestFehlerobjektJSON` |
| Status und Ziel je Code, `E4001` nur 502/504, alle Codes | `E4001` um 500 erweitert; `E5009` auf `replay`; `E7012` entfernt; `E6004` ohne `0` | rot, `TestCodeTabelle` (je Fall) |
| `message` fest, nichtleer, ohne Platzhalter | `%s` in `message`; `message` leer | rot, `TestMessageFest` |
| Code außerhalb §4 nicht darstellbar | Prüfung in `NeuesFehlerobjekt` abgeschaltet | rot, `TestFehlerobjektUngueltigerCode` |
| Abbildung je Zeile und Gegenstand, `Sonstiges` → `E9001` | `ErrPfad` → `E3001`; Dienst → `E7001`; Import-Duplikat → `E6101`; `Exchange` → `E7002`; `E7011` → `E9001`; Rückfall bei fremdem Gegenstand → `E3001`; Rückfall unbekannt → `E3001`; `Sonstiges` bei `ErrNichtGefunden` belegt | rot, `TestAbbildungTabelle` |
| Verpacken, `errors.Join`, erster Treffer der Tabelle | `errors.Is` → `==`; Zeile `ErrUngueltig` an den Anfang der Tabelle | rot, `TestAbbildungVerpackt` |
| `nil` und Kontextfehler: kein Code | `nil`-Prüfung, `Canceled`, `DeadlineExceeded` je einzeln entfernt | rot, `TestAbbildungKeinCode` (je Fall) |
| Adapter verpackt Port-Fehler | `ErrSchemaNeuer` → `ErrPfad`; `ErrPfad` → `ErrBeschaedigt`; `Unwrap` liefert `nil`; `ErrFremdeDatenbank` ohne Verpacken | rot, `TestSqliteFehlerAbbildung` (je Unterfall) |
| Jede Variable `Err…` hat Eintrag | Eintrag `ErrRevisionVeraltet` entfernt; neue Variable im Paket `outbound`; neue Variable im Adapter ohne `verpacke` | rot, `TestAbbildungVollstaendig`, `TestSqliteFehlerVollstaendig` |
| Leer = ungesetzt, Leerraum entfernt | `TrimSpace` entfernt; „leer“ als gesetzt; Leerraum bei „nicht geliefert“ | rot, `TestKonfigLeer` (und Folgetests) |
| `REDACT_*` ergänzen; `none` nur bei zwei Mengen | `REDACT_QUERY_PARAMS` ersetzt den Default; `none` nicht mehr leere Menge | rot, `TestKonfigListenErgaenzen`, `TestKonfigNone` |
| Defaults aus §3 | `LISTEN`-Default geändert | rot, `TestKonfigDefaults`; **`DATA_DIR`-Default grün** (F-1) |
| Nicht gelieferte Schlüssel; fremde Variablen | Präfix `FALLBACK_` entfernt; jede Variable `FOO…` gesperrt | rot, `TestKonfigNichtGeliefert`, `TestKonfigFremdeVariable` |
| Aufzählung und Wahrheitswert nur §3-Schreibweise | `EqualFold`; `1` als wahr | rot, `TestKonfigEnum`, `TestKonfigBool` |
| Dauer | Einheit `w`; `0` zulässig; Maximum +1; 63-Bit-Prüfung abgeschaltet | rot, `TestKonfigDauer` |
| Größe | Grenze 2 GiB; `0` zulässig; zwei Leerzeichen zulässig | rot, `TestKonfigGroesse` |
| Liste | leerer Eintrag übersprungen; Leerraum im Token; `$.a` als Pfad | rot, `TestKonfigListe` |
| `LISTEN`, Service-ID, Root, URL | Port 0 / 65536; leerer Host; `65`-Zeichen-ID; Root mit `/`-Segment und `..`; `user:pass@`; `ftp`; Query | rot, `TestKonfigLISTEN`, `TestKonfigServiceID`, `TestKonfigRoot`, `TestKonfigURL` |
| `mock`/`hybrid` gültig, Start scheitert; Default `mock` | `mock` zulässig; `hybrid` zulässig; ungültiger Wert zusätzlich als „nicht unterstützt“ | rot, `TestKonfigModus`, `TestKonfigLeer`, `TestKonfigModusUngueltig` |
| `SERVICES_FILE` | doppelte `id`; unbekanntes Feld; Anhang nach dem Array; leeres Array; Doppelkonfiguration; kein UTF-8 | rot, `TestKonfigServicesFile` (je Unterfall); Dateigröße und leeres `id`: F-6 |
| Überlappende Roots | Zeichenpräfix statt Segmentgrenze; nur gleiche Roots | rot, `TestKonfigRootUeberlappung` |
| `UPSTREAM_URL` Pflicht in `record` | Pflicht abgeschaltet | rot, `TestKonfigRecordOhneUpstream` |
| TLS-Widersprüche | beide Widerspruchsfälle einzeln abgeschaltet; Wirkung von `SKIP_VERIFY` entfernt | rot, `TestKonfigTLSWiderspruch` |
| Reihenfolge, `DATA_DIR`, CA-Datei | `DATA_DIR` immer anlegen; Rechte `0755`; PEM-Prüfung abgeschaltet | rot, `TestKonfigDatenverzeichnis`, `TestKonfigCADatei` |
| Meldung: alle Verstöße, sortiert, ohne Wert, Exit 1 | nur erster Verstoß; ohne Sortierung; Wert in der Meldung; Exit 2; Code `E3001` | rot, `TestKonfigMeldung` |
| Wiederholung | nicht mutiert (die Eingabe ist rein, die Sortierung ist durch `TestKonfigMeldung` gefangen) | — |

## Auflösung des Rasters (§6 des Plans)

Zeilen *Aufzählung/Wahrheitswert*, *Dauer und Größe*, *Liste*, *`LISTEN`, Roots, URL, `SERVICE_ID`*, *nicht gelieferte Schlüssel*, *`MODE`*, *`SERVICES_FILE`*, *Dateien prüfen*, *Fehler abbilden* und *Fehlerobjekt bilden und serialisieren* sind über die Mutationen der Tabelle oben aufgelöst. Zellen „gilt nicht“ (Abbruch und Wiederholung der reinen Funktionen und der synchronen Konfiguration) sind begründet und nicht mutierbar. Im Code entschiedene Zellen: siehe F-2.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/fehler.go` | geprüft, ohne Befund (Tabelle gegen §4 und SPEC-109 Nr. 2 gelesen) |
| `src/hexagon/ports/outbound/abbildung.go`, `errors.go` | geprüft, ohne Befund |
| `src/adapters/driven/sqlite/sqlite.go`, `fehler_test.go` | geprüft, ohne Befund; keine Verhaltensänderung |
| `src/bootstrap/konfig.go`, `konfig_werte.go`, `konfig_dienste.go` | geprüft, Befunde F-1, F-2, F-5 |
| `src/bootstrap/konfig_test.go` | geprüft, Befunde F-1, F-4, F-6 |
| Geheimnis in Meldung (Skill HIGH „Senke“) | geprüft, ohne Befund: Meldung nennt Schlüssel und Grund, nie den Wert (Mutation „Wert in Meldung“ rot) |
| Kommentare in Code und Tests (§3.7) | geprüft; F-4 |
| Suppression (`//nolint`) | geprüft, ohne Befund |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 2 |
| INFO | 4 |

**Finding-Klassen dieses Laufs:** Default aus §3 ohne Fänger · Randform im Code statt im Vertrag entschieden · Rückführungs-Marke überschritten ohne Plan-Nachzug · Kommentar sagt Prüfung zu, die fehlt · Test spiegelt die Implementierung statt den Vertrag · Fall prüft Ergebnis statt Grund

## Verdikt

**Merge-blockierend:** ja (drei MEDIUM). Übergabe an den Implementer; F-2 berührt Spezifikation und geht an den Architect (jede Antwort ändert SPEC-110 oder SPEC-109, kein Herunterstufen durch Widerspruch). Kein Eintrag dieses Reports bewertet die DoD.
