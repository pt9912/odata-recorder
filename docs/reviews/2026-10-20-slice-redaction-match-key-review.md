# Review-Report: slice-redaction-match-key — 2026-10-20

**Review-Art:** Code — geprüft gegen Plan, Spezifikation, ADRs und Hard Rules (`AGENTS.md` §3.7 bis §3.16), nicht gegen die DoD (Verifier).

**Gegenstand:** `slice-redaction-match-key`, Commits nach `749a5d4` bis `5e188dd` (`c9c20ae` Implementierung, `760f1be` Spec-Entscheid des Architect, `5e188dd` Testfälle) über `src/`, `spec/spezifikation.md` (`SPEC-115`, `SPEC-116`, `SPEC-117`) und den Plan; Summe aus `git diff --numstat` über `src` und `spec/spezifikation.md`: 1872 Zeilen (Marke in §4: 3500).

**Skill:** `.harness/skills/reviewer.md` @ HEAD `5e188dd` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-20

**Methode.** Jede Zeile der Tabelle *Zusage · Fänger* (§3 des Plans) und des Rasters (§6) wurde mit mindestens einer Mutation aufgelöst, über `make test` im Docker (netzlos). Eine Mutation ersetzt eine Stelle in `src/`, `make test` läuft, danach `git checkout -- src/`; rund 90 Mutationen, der Arbeitsbaum ist am Ende sauber (`git status` leer). Vier Probeaufrufe mit einer Wegwerf-Testdatei (nicht committet, gelöscht) belegen die Findings F-1 bis F-3. Eine Mutation zählt nur, wenn sie aus dem benannten Grund rot wird; Mutationen mit Build-Fehler wurden wiederholt.

**Eingangs-Kontext:**

- Slice-Plan `slice-redaction-match-key` (§1, §3 Zusage·Fänger, §6 Raster, §4 Marke)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Go), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md) (Schichtenrichtung; `src/bootstrap/` ruft `src/hexagon/model/`, nicht umgekehrt)
- [LH-FA-05](../../spec/lastenheft.md), [LH-FA-14](../../spec/lastenheft.md), [LH-QA-02](../../spec/lastenheft.md), [LH-QA-04](../../spec/lastenheft.md)
- `SPEC-115`, `SPEC-116`, `SPEC-117`, `SPEC-110` Nr. 6, `SPEC-045`
- `AGENTS.md` (Hard Rules)

---

## Findings

Kein Fix wird vorgeschlagen (Skill: Reviewer kategorisiert, Implementer entscheidet). Der Befund zu F-1 ist eine Randform ohne Ort; sie geht an den Architect (`AGENTS.md` §3.12).

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `Location` (oder `Content-Location`) in `MATCH_VARY_HEADERS` geht mit dem unredigierten Wert in den Key ein, während `Schluessel.Header` die Query redigiert: zwei Requests, die sich nur im Wert von `token` in `Location` unterscheiden, ergeben verschiedene Keys (Probe: `gleich=false`). Das widerspricht `SPEC-116` Nr. 8 („ein anderes Geheimnis ergibt denselben Key“); der Key, der persistiert wird, hängt vom Geheimnis ab, und ein Replay mit anderem Token findet den Exchange nicht. Das Raster hat für Vary × redigierter Parameter in URL-Header keine Zelle. | `SPEC-116` Nr. 5 und Nr. 8, `SPEC-115` Nr. 3, `AGENTS.md` §3.16 | `src/hexagon/model/schluessel.go` · `ausgenommen := mengeKlein(append(slices.Clone(regeln.VolatileHeaders), regeln.RedactHeaders...))` | ja — Wegwerf-Test: `VaryHeaders=["Location"]`, `Location: /p?token=eins` gegen `…zwei` | Randform ohne Ort im Raster (Vary × URL-Header) |
| F-2 | MEDIUM | Die Mutation „Namen der Vary-Menge nicht sortieren“ bleibt grün (`make test` rc 0). `SPEC-116` Nr. 5 sagt „Bytefolge aufsteigend“ zu und `SPEC-110` Nr. 6 sagt, die Reihenfolge einer Liste habe keine Wirkung; die Fixtures `schluesselRegeln()` führen `VaryHeaders` bereits sortiert, die Zeile „Reihenfolge der Map“ des Plans prüft nur die Map des Requests. Zwei Betreiber mit derselben Menge in anderer Reihenfolge erhielten verschiedene Keys. | `SPEC-116` Nr. 5, `SPEC-110` Nr. 6, `AGENTS.md` §3.15 | `src/hexagon/model/schluessel.go` · `slices.Sort(namen)` (in `vary`) | ja — Mutation oben; Wegwerf-Test `VaryHeaders` in zwei Reihenfolgen ergibt heute denselben Key | Zusage ohne Fänger |
| F-3 | MEDIUM | Die Mutation „Form des Pfades erst nach der Kodierung des Pfades prüfen“ (Pfad `/a%zz#` bei `reject`) bleibt grün; ungemutiert ergibt der Aufruf `ErrUngueltig`, gemutet `ErrKodierung`. `SPEC-116` Nr. 9 legt „Form des Pfades, dann Kodierung des Pfades“ fest; `TestKeyReihenfolgeVerstoesse` und die Plan-Zeile führen diese Paarung nicht, und die Reihenfolge Pfad-Kodierung gegen Query-Kodierung ist beobachtbar nicht unterscheidbar (beide `ErrKodierung`). | `SPEC-116` Nr. 9, `AGENTS.md` §3.15, §3.16 (Zusatzzeile „Zwei Verstöße in einem Aufruf“) | `src/hexagon/model/schluessel_test.go` · `func TestKeyReihenfolgeVerstoesse` | ja — Mutation oben | Paar von Verstößen ohne Fänger |
| F-4 | MEDIUM | Die Mutation „Namen-Abgleich der Header mit Unicode-Faltung (`strings.ToLower`)“ bleibt grün. Der Test `TestRedigiereHeaderNamen` setzt Header `Coo`+U+212A+`ie` gegen die Menge `Cooxie` (ein `x`, kein `k`); beide Namen gleichen sich unter keiner Faltung, der Fall kann nicht rot werden. Der Kommentar „Kelvin-Zeichen U+212A ist kein ASCII-k und gleicht dem Namen nicht“ sagt mehr zu, als der Test prüft (`AGENTS.md` §3.11). Die Richtung der Abweichung redigiert mehr, nie weniger. | `AGENTS.md` §3.11, §3.15; Plan-Zeile „Name-Abgleich ohne Beachtung der ASCII-Schreibweise“ | `src/hexagon/model/redaction_test.go` · `// Kelvin-Zeichen U+212A ist kein ASCII-k und gleicht dem Namen nicht.` | ja — Mutation oben (rc 0) | Kommentar sagt mehr zu als der Test fängt |
| F-5 | MEDIUM | Sechs Zellen des Vertrags wurden zuerst im Code entschieden: `\u00XX`-Schreibweise, Marke und Text der Art 2 mit Längenpräfix, Nicht-Token-Header-Namen (drei vom Implementer gemeldet), `#` vor dem ersten `?` in `Location`, Escape ohne UTF-8 im Pfad, Hex-Schreibweise im Text der Art 2 (drei erst bei der Prüfung des Codes gegen den Vertrag gefunden). Der Architect hat alle sechs in `760f1be` wie im Code bestätigt, mit Tests in `5e188dd`; jede Zelle hat jetzt einen Ort und eine rote Mutation (siehe unten). Die Regel `AGENTS.md` §3.16 stuft „im Code entschieden“ als MEDIUM ein, unabhängig davon, dass die Spezifikation nachzog. | `AGENTS.md` §3.12, §3.16 | `docs/plan/planning/done/slice-redaction-match-key.md` · „Der Implementer fand drei Zellen“ | nein — Prozessbefund, die Zellen selbst sind in `SPEC-115`/`SPEC-116` entschieden | Randform erst im Code entschieden |
| F-6 | MEDIUM | `SPEC-116` Nr. 1 und Nr. 5 sagen nicht, dass die Vary-Folge mit der Anzahl der **vorhandenen** Namen beginnt; Nr. 5 nennt nur „Eine leere Menge (`none`) ergibt die Anzahl 0“. Der Code schreibt die Anzahl der Namen, die der Request mit mindestens einem Wert trägt (`zahl(len(vorhanden))`); `TestKeyKodierungHandgebaut` übernimmt diese Lesart (`zahl(1)` bei drei konfigurierten Namen, ein vorhanden). Eine Lesart „Anzahl der konfigurierten Namen“ ergäbe andere Keys für dieselbe Aufzeichnung; das Format ist Teil von `SPEC-015`. | `SPEC-116` Nr. 1 und Nr. 5, `AGENTS.md` §3.16 | `spec/spezifikation.md` · „Eine leere Menge (`none`) ergibt die Anzahl 0.“ | nein — Lesen von Nr. 1 gegen den Test (bekanntes akzeptiertes Negativ: eine Quelle; hier fehlt der Satz in der Quelle selbst) | Kodierungs-Zelle nur durch den Code festgelegt |
| F-7 | LOW | Ein Gate fängt „Stabilität der Sortierung der Objektmitglieder“ nicht: `slices.SortStableFunc` gegen `slices.SortFunc` bleibt grün, weil die Tests höchstens zwei Mitglieder mit gleichem Namen führen und der Sortieralgorithmus bis zwölf Elemente stabil verfährt. `SPEC-116` Nr. 6 sagt „bei gleichem Namen in Eingabereihenfolge“ zu. | `AGENTS.md` §3.15, `SPEC-116` Nr. 6 | `src/hexagon/model/schluessel_json.go` · `slices.SortStableFunc(ms, …)` | ja — Mutation oben (rc 0) | Zusage ohne Fänger jenseits der Testgröße |
| F-8 | LOW | Plan-Zeilen sagen eine breitere Allaussage zu, als die Tests prüfen: `TestKeinGeheimnisImErgebnis` läuft nicht über „alle Fälle der Tabellen in `TestRedigiere*` und `TestKey*`“, sondern über `queryFaelle()` und `headerFaelle()` (die Tabelle von `TestRedigiereLocation` und die `TestKey*`-Tabellen fehlen); `TestRedigiereIdempotent` deckt die Fälle `/p#f?token=1`, `http://x?` und `?` der Location-Tabelle nicht ab. Jeder dieser Fälle hat eine eigene Erwartung im jeweiligen Test, die Lücke ist die der Allaussage. | `AGENTS.md` §3.15 (Allaussage nennt Fänger über alle Fälle), §3.16 | `src/hexagon/model/redaction_test.go` · `func TestKeinGeheimnisImErgebnis` | ja — Mutation „Content-Location nicht untersuchen“ wird nur von `TestRedigiereLocation` und zusätzlich von `TestKeinGeheimnisImErgebnis` gefangen, „Fragment-Fall“ nur von `TestRedigiereLocation` | Allaussage ohne Fänger über alle Fälle |
| F-9 | LOW | `TestReplayfaehigMatrix` schreibt die Erwartung nicht für alle 54 Kombinationen aus, wie der Plan sagt („die Erwartung ausgeschrieben und nicht aus derselben Bedingung berechnet“): neun Zeilen sind ausgeschrieben, die Spalte Capture wird mit `if ci != 0 { want = false }` berechnet. Der Kommentar beginnt mit `matrix schreibt …` (ein Name, den es nicht gibt). Die Mutationen auf `Replayfaehig` sind trotzdem alle rot. | `AGENTS.md` §3.7, §3.11; Plan-Zeile „Replayfähigkeit: jede Zeile der Matrix“ | `src/hexagon/model/replay_test.go` · `// matrix schreibt die Erwartung aus` | ja — Lesen | Plan-Zusage weicht vom Test ab |
| F-10 | LOW | Die Allaussage in `SPEC-116` Einleitung („ein Aufrufer kann den Key nicht aus einer unredigierten Form bilden“) gilt für Header und Query; für den Body trägt sie allein der Kommentar an `SchluesselEingabe.Body` („die Bytes des bereits redigierten Request-Bodys“), und kein Test oder Satz der Spezifikation fängt einen unredigierten Body. Die Redaction des Bodys liegt laut Plan bei `slice-streaming-redaction`. | `AGENTS.md` §3.11, `SPEC-116` Einleitung und Nr. 8 | `src/hexagon/model/schluessel.go` · `Body trägt die Bytes des bereits redigierten Request-Bodys.` | nein | Allaussage reicht über den Fänger hinaus |
| F-11 | INFO | Bekannte Lücke (e) „kein roter Gegenbeweis für den leeren Header-Namen als eigener Eintrag“: die Prüfung ergibt keine nicht-äquivalente Mutation, die grün bleibt. „Leeren Namen wie `X-Leer` kanonisieren“ und „leeren Namen in `Namen` aufnehmen“ werden von `TestRedigiereHeaderNichtToken` rot; „Wächter `if name == ""` entfernen“ ist durch Lesen äquivalent (die Schleife läuft über null Bytes und gibt `""` zurück; der Lauf mit entferntem Wächter brach am `gofmt`-Gate ab und ist kein Beleg). Die Lücke betrifft damit keine beobachtbare Zusage. | `SPEC-115` Nr. 2 | `src/hexagon/model/redaction.go` · `if name == "" {` | ja — drei Mutationen | Aufwand: keine Aktion |
| F-12 | INFO | Ein `#`-Fragment nach dem Query-Teil von `Location` bleibt unredigiert (`/p?a=1#f?token=1` behält `token=1` im Fragment); das entspricht `SPEC-115` Nr. 3 („bis zum ersten folgenden `#`“) und dem Test `http://x/p?a=1&api_key=k#frag?token=1`. Ein Fragment erreicht den Server nie; Hinweis für den Verifier gegen `LH-FA-14.a`. | `SPEC-115` Nr. 3 | `src/hexagon/model/redaction_test.go` · `"http://x/p?a=1&api_key=k#frag?token=1"` | nein | Verifier-Hinweis |
| F-13 | INFO | `RedigiereHeader` kopiert die Wertlisten des nicht redigierten Falls zweimal (`fasseNamenZusammen` und `slices.Clone` im Default-Zweig); die Mutation „Clone entfernen“ bleibt grün (äquivalent). Keine Aktion erwartet. | Maintainability | `src/hexagon/model/redaction.go` · `out[name] = slices.Clone(werte)` | ja — Mutation | äquivalente Mutation |

## Auflösung der Zusage · Fänger-Tabelle (Plan §3) und des Rasters (Plan §6)

Je Zeile eine Mutation; „rot“ nennt die Tests, die aus dem benannten Grund fehlschlugen. Unter „Zusage ohne Fänger“ stehen die Zeilen, deren Mutation grün blieb.

| Zusage (Plan-Zeile) | Mutation | Rot |
|---|---|---|
| Marker (`TestMarker`) | `Marker = "[GEHEIM]"` | `TestMarker` und weitere |
| Header-Werte | redigierter Header mit leerer Liste behält die Liste | `TestRedigiereHeaderWerte` |
| Header-Namen ohne Beachtung der ASCII-Schreibweise | Abgleich nur auf exaktem Namen (`slices.Contains`) | `TestRedigiereHeaderNamen`, `…NichtToken`, `TestRedigiereLocation` |
| ähnliche Namen | Präfix `authori` redigiert mit | `TestRedigiereHeaderNamen` |
| nicht redigierte Header unverändert | nur erster Wert bleibt | `TestRedigiereHeaderUnveraendert`, `…Schreibweise` |
| Schreibweisen zusammenführen | Namen nicht kanonisieren | `TestRedigiereHeaderSchreibweise`, `TestKeyVary`, `TestKeyGleichwertigkeit`, `TestKeyKodierungHandgebaut` |
| Nicht-Token-Namen getrennt | Nicht-Token-Name mit Leerzeichen klein schreiben | `TestRedigiereHeaderNichtToken` |
| Kopie, nie `nil`, Eingabe unverändert | `nil` bei leerer Eingabe; `OhneRedigierte` ohne `Clone` | `TestRedigiereHeaderKopie` |
| `Location`: `#` vor `?` | erst nach `#` die Query suchen | `TestRedigiereLocation` |
| `Location`: Fragment bleibt | Fragment verwerfen | `TestRedigiereLocation` |
| `Location`: ganz redigierter Header gewinnt | Header-Redaction vor URL-Zweig aussetzen | `TestRedigiereLocation` |
| `Content-Location` | Namen aus `urlHeader` streichen | `TestRedigiereLocation`, `TestKeinGeheimnisImErgebnis` |
| Meldung „Wert änderte sich“ | immer `true` | `TestRedigiereQueryGeaendert`, `TestKeyRedacted`, `TestRedigiereIdempotent` |
| Query-Zerlegung | `;` trennt mit | `TestRedigiereQueryRandformen`, `TestKeyQueryZerlegung` |
| Segment ohne `=` | redigiert mit | `TestRedigiereQueryRandformen`, `…Geaendert` |
| `%`-dekodierter Name | Dekodierung abschalten | `TestRedigiereQueryRandformen` |
| jedes Vorkommen | nur der erste Treffer | `TestRedigiereQueryRandformen`, `TestKeinGeheimnisImErgebnis`, `TestKeyGeheimnisUnsichtbar` |
| ASCII-Schreibweise der Parameter | Vergleich ohne Faltung | `TestRedigiereQueryRandformen` |
| Allaussage „kein Geheimnis im Ergebnis“ | zweiter Parameter im Key nicht redigiert (`RedactQueryParams[:1]`); Header/Query unredigiert zurückgeben | `TestKeinGeheimnisImErgebnis`, `TestKeyGeheimnisUnsichtbar`, `TestKeyIdempotent` |
| `OhneRedigierte` | „jeder Marker“ statt „alle“; leere Liste mitentfernen | `TestOhneRedigierte` |
| Pfadmuster | leerer Name erlaubt; Leerzeichen erlaubt | `TestParseJSONPfad`, `TestKonfigJSONPfadWieModell`, `TestKonfigListe` |
| Bedeutung des Musters | Groß-/Kleinschreibung zählt; Unicode-Faltung (`strings.ToLower`) | `TestJSONPfadPasst` |
| Start nutzt denselben Parser | eigene Grammatik in `istJSONPfad` mit `.` im Namen | `TestKonfigJSONPfadWieModell`, `TestKonfigListe` |
| Kodierung eindeutig | Längenpräfix weglassen | `TestKeyEindeutig`, `TestKeyKodierungHandgebaut`, `TestKeyBodyRueckfall` |
| Kodierung nach Nr. 1 | Marke `v2`; Hex der Kanonform groß; `\n` als Kurzform; U+007F escapen | `TestKeyKodierungHandgebaut` (und weitere) |
| `Method`, `ServiceID` leer | Prüfung abschalten | `TestKeyKopf`, `TestKeyReihenfolgeVerstoesse` |
| Pfad: Escape ohne UTF-8 gültig | `%FF` als ungültig melden | `TestKeyPfad` |
| Pfad: Gleichheiten und Unterschiede | Pfad klein schreiben; Hex nicht groß; unreserviertes Escape nie ersetzen | `TestKeyPfad`, `TestKeyGleichwertigkeit`, `TestKeyKodierungHandgebaut` |
| Query-Paare: Multimenge, Reihenfolge | Paare nicht sortieren; Doppelte entfernen | `TestKeyQueryReihenfolge`, `TestKeyBeispieleLH`, `TestKeyGleichwertigkeit` |
| Art 2 byteweise | Text der Art 2 groß schreiben | `TestKeyUngueltigeKodierung`, `TestKeyKodierungHandgebaut` |
| `+` bleibt `+` | `+` als Leerzeichen dekodieren | `TestKeyQueryDekodierung` |
| `reject` liefert `ErrKodierung` | `ErrUngueltig` an beiden Stellen | `TestKeyPfad`, `TestKeyUngueltigeKodierung`, `TestKeyRedigiertVorKodierung`, `TestKeyReihenfolgeVerstoesse` |
| Prüfung über die redigierte Darstellung | Paare aus `in.RawQuery` zerlegen | `TestKeyRedigiertVorKodierung`, `TestKeyGeheimnisUnsichtbar`, `TestKeyIdempotent` |
| Vary: fehlend gegen leer | leeren ersten Wert wie fehlend behandeln | `TestKeyVary` |
| Vary: Rand-Leerraum | nicht trimmen | `TestKeyVary`, `TestKeyKodierungHandgebaut` |
| Vary: Ausnahme volatil und redigiert | nur flüchtige; nur redigierte | `TestKeyVaryAusnahme` (und weitere) |
| Vary: Reihenfolge der Menge | Namen nicht sortieren | **grün (F-2)** |
| Vary: Reihenfolge der Map | Werte der Schreibweisen unsortiert zusammenführen | `TestKeyDeterministisch`, `TestKeyVary`, `TestRedigiereHeaderSchreibweise` |
| Body leer gleich `nil`; `bytes` jedes Byte | nur `nil` als leer; Body trimmen | `TestKeyBodyBytes`, `TestKeyBodyRueckfall` |
| `json-canonical` | Mitglieder nicht sortieren; Zahlen normalisieren | `TestKeyBodyKanonisch`, `TestKeyKodierungHandgebaut` |
| Doppelte Schlüssel in Eingabereihenfolge | `SortStableFunc` gegen `SortFunc` | **grün (F-7)** |
| Rückfall auf Art 2 | Tiefe `>=`; Tiefe 600; Low-Surrogat annehmen; UTF-8 nicht prüfen | `TestKeyBodyKanonisch`, `TestKeyBodyRueckfall` |
| `KeyRedacted` | `applied` nicht werten; `unsafe` erlauben | `TestKeyRedacted`, `TestKeyReihenfolgeVerstoesse` |
| redigierte Darstellung im Ergebnis | Header oder Query unredigiert zurückgeben | `TestKeyIdempotent`, `TestKeinGeheimnisImErgebnis` |
| Aufzählungen der Regeln | `MATCH_INVALID_ENCODING` und `MATCH_BODY_MODE` leer erlauben | `TestKeyRegelnUngueltig` |
| Reihenfolge der Prüfungen (Nr. 9) | Pfadform nach Pfadkodierung | **grün (F-3)** |
| Allaussage Gleichwertigkeit (Nr. 10) | Pfad klein schreiben (zu viel); Hex nicht normalisieren (zu wenig) | `TestKeyGleichwertigkeit` |
| Replay-Matrix, Unbekannt, Unabhängig | Capture ignorieren; `absent` streichen; `applied` immer; Unbekannt `true`; `KeyRedacted` wirkt | `TestReplayfaehigMatrix`, `TestReplayfaehigUnbekannt`, `TestReplayfaehigUnabhaengig` |
| Kommentar-Zusage Unicode (Header) | Unicode-Faltung im Namen-Abgleich | **grün (F-4)** |

Nicht mutiert (kein plausibler Eingriff, der sich von einer äquivalenten Mutation unterscheidet): `TestReplayfaehigNullwert`, `TestKeyEingabeUnveraendert` (die Fixtures sind schon sortiert, ein In-place-Sortieren der Konfigurationsliste bliebe unsichtbar; Hinweis, kein eigenes Finding), `TestRedigiereQueryUngueltigeKodierung` (die Signatur trägt keinen Fehler).

## Lesen von `SPEC-116` Nr. 1 gegen `TestKeyKodierungHandgebaut` (Auftrag c)

Unabhängig aus dem Spezifikationstext gelesen: Marke `odr-match-v1`, `Method`, `ServiceID`, Pfad je als Länge (4 Byte Big Endian) und Bytes; Paare als Anzahl, je Paar ein Byte Art, Schlüssel (und bei Art 1 Wert) als Zeichenfolge, bei Art 2 der Text als Zeichenfolge, sortiert nach (Art, Schlüsselbytes, Wertbytes); Body als Art-Byte und 32 Byte Digest (Art 0 ohne Digest); Digest-Präfix `sha256:` mit 64 kleinen Hex-Ziffern. Der Test baut diese Folge für Art 0, 1 und 2 der Paare und Art 0, 1, 2 und 3 des Bodys von Hand nach (`%7e` zu `~`, `%2f` zu `%2F`, `%c3%A9` zu `%C3%A9`, `\u001F` zu `\u001f`, `\n` zu `\u000a`, U+007F unverändert); die Sortierung (Art 0 vor Art 1 vor Art 2, `a` vor `token`) stimmt mit dem Text. Eine Abweichung des Tests vom Text fand sich nicht. Offen ist allein die Anzahl der Vary-Namen: der Test setzt `zahl(1)` für die Anzahl der vorhandenen Namen, der Text von Nr. 1 und Nr. 5 nennt sie nicht ausdrücklich (F-6).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/model/redaction.go` (Redaction, Pfadmuster, `dekodiereProzent`) | geprüft, Befunde F-1, F-4, F-11, F-13; sonst ohne Befund |
| `src/hexagon/model/schluessel.go` | geprüft, Befunde F-1, F-2, F-3, F-6 |
| `src/hexagon/model/schluessel_json.go` (Kanonform, Zahlen, Surrogate, Tiefe) | geprüft, Befund F-7; sonst ohne Befund (Tiefe 512 und 513, Zahlengrammatik, einzelnes Surrogat, UTF-8 durch Mutation gefangen) |
| `src/hexagon/model/replay.go` (`SPEC-117`) | geprüft, ohne Befund |
| `src/bootstrap/konfig_werte.go` (`istJSONPfad` ruft den Parser des Modells) | geprüft, ohne Befund; Schichtenrichtung `bootstrap` nach `model` entspricht [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md) |
| `src/hexagon/model/*_test.go`, `src/bootstrap/konfig_test.go` | geprüft, Befunde F-3, F-4, F-8, F-9 |
| `spec/spezifikation.md` (`SPEC-115` bis `SPEC-117`, Änderungsverzeichnis) | geprüft, Befunde F-5, F-6, F-10; kein Widerspruch zu [LH-FA-05](../../spec/lastenheft.md), [LH-FA-14](../../spec/lastenheft.md) und [LH-QA-02](../../spec/lastenheft.md) gefunden |
| Kommentare im Diff (`AGENTS.md` §3.7, §3.11) | geprüft; kein Konjunktiv über verworfene Alternativen, keine Chronik in Code-Kommentaren; Befunde F-4, F-9, F-10 |
| Plan `slice-redaction-match-key` (§3.9: Plan folgt der Korrektur) | geprüft, ohne Befund: Kopf, §1, §3, §6 tragen den Stand von `760f1be`; Marke in §4 nennt ihre Metrik; Zusatzzeilen des Rasters haben je einen Ort |
| Suppression (`AGENTS.md` §3.2) | geprüft, ohne Befund: keine `//nolint`-Direktive im Diff |
| `harness/gegenproben.stempel` | geprüft, ohne Befund (Stempel folgt dem geänderten Eingabestand) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 6 |
| LOW | 4 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Randform ohne Ort im Raster (Vary × URL-Header) · Zusage ohne Fänger · Paar von Verstößen ohne Fänger · Kommentar sagt mehr zu als der Test fängt · Randform erst im Code entschieden · Kodierungs-Zelle nur durch den Code festgelegt · Allaussage ohne Fänger über alle Fälle

## Verdikt

**Merge-blockierend:** ja — wegen der sechs MEDIUM. F-2, F-3 und F-4 sind durch einen Test zu schließen, der aus dem benannten Grund rot wird; F-1 und F-6 sind Randformen ohne Ort und gehen an den Architect (`AGENTS.md` §3.12), bevor der Code sie entscheidet; F-5 ist ein Prozessbefund ohne weitere Code-Aktion (die Spezifikation trägt die Zellen). Kein HIGH: die Redaction der Header und Query greift in allen gefahrenen Mutationen, und der Key enthält nach den Tests kein Geheimnis in Klartext. F-1 ist die einzige Stelle, an der ein Geheimnis den persistierten Key beeinflusst (als SHA-256, nur bei `Location`/`Content-Location` in `MATCH_VARY_HEADERS`); stuft der Architect diese Konfiguration als unterstützt ein, ist die Einordnung des Findings neu zu bewerten.

**Übergabe:** Findings gehen an den Implementer (F-2, F-3, F-4, F-7 bis F-9) und an den Architect (F-1, F-6, F-10); die **Finding-Klassen** gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. Dieser Report ist ein Lauf-Beleg und ersetzt keine Verifikation: DoD- und Spec-Konformität prüft der Verifier separat.
