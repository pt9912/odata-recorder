# Review slice-test-contract (Reviewer, ein Lauf)

* Gegenstand: Zweig `slice-test-contract`, HEAD `a4486595` (Code `aaeb8f71`), Basis `54a1f1b4`
* Gemessen gegen: Plan `slice-test-contract.md` (T01 bis T17, Raster §6, §3 Mutationen), `SPEC-042` Nr. 1 bis 9, `SPEC-049` Nr. 1/11 a, ADR-0005, ADR-0012, `AGENTS.md` §3.2, §3.6, §3.9 bis §3.18
* Summe: HIGH 0 · MEDIUM 2 · LOW 1 · INFO 4

## Findings

### MEDIUM

1. **kategorie** MEDIUM · **quelle** `AGENTS.md` §3.15, `SPEC-042` Nr. 3 b und 6
   **pfad** `src/bootstrap/vertrag_http_test.go` · `req.Header.Set("Authorization", "Bearer GEHEIM-A")`; `tests/integration/vertrag_fehlerklassen_test.go` · `"Authorization": "Bearer GEHEIM-A"` und `?x=GEHEIM-Q`
   **befund** Die Zusage „jede Anfrage trägt `Authorization`, `Cookie` und (bei zwei Klassen) `GEHEIM-Q`“ hat keinen Fänger. Mutation (gefahren, Scratchpad-Kopie): die zwei `Header.Set` der Matrix gelöscht, die Zuordnung der Klassen auf `{}` gesetzt, `GEHEIM-Q` aus den Pfaden entfernt ⇒ `go test -run '^TestVertrag'` beider Pakete grün. Damit ist auch der Fänger von T15 (Mutation „Adapter spiegelt `Authorization`“) stumm, sobald die Sendeseite wegfällt: kein Marker steht in der Anfrage, also kann keine Antwort ihn tragen.
   **verifizierbar** ja (Mutation oben) · **klasse** Zusage der Sendeseite ohne Fänger (Marker-Wirkung nur auf der Empfangsseite geprüft)

2. **kategorie** MEDIUM · **quelle** `AGENTS.md` §3.15/§3.16, `SPEC-042` Nr. 8 d
   **pfad** `tests/integration/m2_test.go` (Hilfe `hochgefahren`) · `t.Cleanup(func() { _ = h.Herunterfahren(context.WithoutCancel(ctx)) })`
   **befund** Nr. 8 d sagt zu: „ein Fehler beim Herunterfahren ist rot“. In `vertrag_http_test.go` gilt das (`t.Errorf` in `vertragServer`), in `TestVertragFehlerklassen` verwirft die Hilfe den Fehler des Handle; kein Fall fängt ihn. Die Randform *Herunterfahren* ist im Raster (Zeile „Fall einer Fehlerklasse“, Spalte Abbruch: Nr. 8 a, d) mit Ort eingetragen, aber nur für die eine Hälfte der Tests belegt. Die Zusage steht ohne Fänger und ohne benanntes Negativ in Nr. 9.
   **verifizierbar** nein (Mutation am Handle nicht gefahren; Lesebefund) · **klasse** Randform nur in einem von zwei Test-Paketen gefangen

### LOW

3. **kategorie** LOW · **quelle** Maintainability / `AGENTS.md` §3.9
   **pfad** `docs/plan/planning/in-progress/slice-test-contract.md` · „Der Fall *kein PASS* (Name in der Liste, Test läuft nicht) ist nicht gefahren.“
   **befund** Der Fall ist gefahren (siehe INFO 5); der Satz im Handlauf-Absatz von §3 ist damit überholt und der Plan zieht die Messung nicht nach.
   **verifizierbar** ja · **klasse** Plan folgt der Messung nicht

### INFO

4. Bekannt (1), `$metadata` bei kaputtem Blob: Der Architect hat in `a4486595` entschieden (Nicht-Fall, `SPEC-042` Nr. 9 i, Plan T10 und Stand, Zelle in `harness/README.md` „Nicht behauptet“). Ich bewerte die Entscheidung als konsistent: Plan, Vertrag, Test und README sagen dasselbe; die Folge (Abbildung nur auf dem Pfad `$metadata` unbemerkt) ist mit Grund benannt. Verweis an Verifier: die Rückmeldung an die Spec-Stelle `SPEC-198` Nr. 4 selbst (soll `$metadata` ein ungültiges Dokument melden?) ist keine Frage dieses Slice.
5. Handlauf T16, der nicht gefahrene Fall „kein PASS“ (gefahren in Scratchpad-Kopie): `integ` um `TestM3` ergänzt (Name steht in `-test.list`, passt nicht auf `^TestVertrag`) ⇒ `test-contract: rot (/out/integration.test: kein PASS (TestM3))`, Exit 1, das erste Binary lief und blieb grün. Wertung des Treibers wie in Nr. 2 zugesagt; Präfix-Fall `TestVertragMatrix` / `TestVertragMatrixZaehlung` ist durch das Leerzeichen im Muster `--- PASS: <Name> ` getrennt, Unterläufe durch den Anker `^`.
6. Bekannt (2), zwei Fassungen von `vertragFehlerobjekt` (und `vertragFelder`, `vertragGeheimnisse`, `vertragMarker`, `vertragCodes`): Im Code Zeile für Zeile gleich bis auf den Typ der Antwort. Der Architect hat Nr. 4 und 9 j so entschieden; jede Fassung hat ihre Schleife (9 Codes × 20 Eigenschaften, Stichwort des Grundes). Bewertung: tragbar, keine ADR-Verletzung (ADR-0012 hält die Kanten). Rest: `vertragCodes()` und die Schleife sind doppelt zu pflegen; ein 10. Code der Fixtures fällt in `bootstrap_test` und `integration_test` getrennt auf (kein Fänger für „Menge `vertragCodes()` deckt alle Codes der Fixtures“, keine Zusage des Vertrags).
7. Geprüft ohne Befund: `Dockerfile` ändert nur durch eine neue Stufe `contract` (reines Anhängen, `--network=none` je `RUN`), keine bestehende Stufe geändert; `make test-contract` hängt nicht an `GATE_CHECKS` (`harness/mk/test-contract.mk`, Kommentar sagt es, kein `GATE_CHECKS +=`); `.a-check.yml` unberührt, `tests/integration` nennt nur `model`, `adapters/driven` (ADR-0012), `bootstrap_test` den Composition Root; Stempel `harness/gegenproben.stempel` gleich der neu gerechneten Eingabe (`tools/harness/gegenproben-eingabe.sh`); `harness/README.md` Zelle *Tut was* 181 und *Bindung* 161 Zeichen (≤ 200); der Satz „Von `make test-contract` ist nicht behauptet“ nennt Verdrahtung, Binary, `$metadata`, `make gates`; er ist nicht erschöpfend gegenüber Nr. 9 b bis e, g, j, sagt das aber auch nicht (§3.11 nicht verletzt). Kein `//nolint`, kein gelockertes Gate.

## Auflösung der Tabelle (T01 bis T17)

Statisch gegen den Code aufgelöst, jeweils mit der Mutation aus dem Plan als Maßstab; selbst gefahren nur die Mutation zu Finding 1 und der Treiberfall (Host-Last 18, ein Lauf je Mutation, `-count=1`).

* T01: `TestVertragMatrixZaehlung` als Schleife über 36 IDs × (fehlt, doppelt) + 92 Schritte; Fänger deckt die Klasse, keine Befunde.
* T02/T03: `TestVertragVergleich`, Schleife über alle Header-Namen der Fixture (3 Fälle je Name) und vier Body-Formen; Gleichheitsfall (`gleich`) ist vorhanden.
* T04: Schleife über 256 Bytes, Soll als eigene Literal-Menge (nicht aus der Implementierung abgeleitet), Zählung 171; keine Befunde.
* T05 bis T07: über `TestVertragMatrix` (E-04 a strikt, `GET-01` lax; CRUD; Revision); T05 hat als Fänger nur einen Fixture-Schritt (`ODV4-E-04` a), die Zusage „alle anderen lax“ fängt jeder lax Schritt; keine Befunde.
* T08: 9 Codes × 20 Eigenschaften je Paket; Reihenfolge-, Leerraum-, Typ-Fälle vorhanden.
* T09 bis T14: Tabelle, Paarschleife, Zählprüfung (Zeile *Tabelle* mit „fehlt“, „fünfte“, „gleicher Code“) vorhanden; T10 nur Collection (Nicht-Fall `$metadata`, siehe INFO 4).
* T15: Empfangsseite gefangen, Sendeseite nicht (Finding 1).
* T16/T17: Negative mit Grund und Folge in `SPEC-042` Nr. 9 f, a bis e, g, h, i, j.

Raster §6: drei Zeilen vollständig mit Ort; die Zelle „Herunterfahren“ in der Zeile *Fall einer Fehlerklasse* trägt Nr. 8 d, der Fänger fehlt dort (Finding 2). Zelle „Wiederholung“ (Nr. 8 f): jede Klasse ruft `m3Aufbau` neu (eigenes `t.TempDir`), kein Befund.

## Negativbefund (geprüft, ohne Befund)

* `Dockerfile` (Stufe `contract`): ohne Befund (INFO 7)
* `harness/mk/test-contract.mk`, `tools/harness/test-contract.sh`: ohne Befund (Wertung, Signal, Präfix-Fall, kein-PASS-Fall gefahren)
* `harness/README.md`, `harness/gegenproben.stempel`: ohne Befund
* `src/bootstrap/vertrag_http_test.go`: ohne Befund außer Finding 1
* `tests/integration/vertrag_fehlerklassen_test.go`: Findings 1 und 2
* `harness/vertraege/SPEC-042-test-contract.md`, Plan: Finding 3; sonst ohne Befund
