# Verifikation slice-hybrid-service-upstream, Teil 2 (Upstream-Fallback)

**Rolle:** Verifier (einmal). **Datum:** 2026-11-18. **Stand:** Zweig `slice-hybrid-service-upstream`, HEAD `7eff3500`. **Bezug:** [LH-FA-12](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), `SPEC-284`, `SPEC-261`, `SPEC-120`, `SPEC-121`.
**Plan:** [slice-hybrid-service-upstream](../plan/planning/done/slice-hybrid-service-upstream.md). **Modell:** claude-sonnet-5-5.
Die beiden Review-Reports (Runde 1 und 2) sind nicht als Beleg gelesen worden; alles unten ist selbst gefahren.

## Urteil

**PASS mit einer Auflage (MEDIUM) und drei LOW.** Alle Gates sind frisch grün, der Code folgt `SPEC-284` Nr. 1 bis 13, und von 63 eigenen Mutationen an Stellen, die ich selbst wählte, sind 56 rot (Zahlen siehe unten). Eine Zusage hat auf dem Weg U keinen Fänger: „Der Dienst schließt den Body der Anfrage nie“ (`SPEC-284` Nr. 5 d, 10); ihre Mutation bleibt grün (V-1, MEDIUM nach `AGENTS.md` §3.15). Das ist eine Lücke des Tests, nicht des Codes: der Code schließt den Body nicht. Die Auflage geht an den Implementer vor der Closure; sie ändert den Code nicht.

## Findings

| Nr. | Schwere | Ort | Befund |
|---|---|---|---|
| V-1 | MEDIUM | `src/hexagon/services/hybrid/upstream.go:54` (Aufruf `Senden`); Fänger fehlt in `body_test.go:243` (`TestBodyNieGeschlossen`) und `upstream_test.go` | Die Zusage „der Dienst schließt den Body der Anfrage nie“ (`SPEC-284` Nr. 5 d, Nr. 10) hat auf dem Weg U keinen Fänger. `TestBodyNieGeschlossen` (Z16, Teil 1) läuft mit `neuesRig` ohne `Ziel`, U wird dort nie erreicht; die Tests der Teile U (`upstream_test.go:263-290`, `:543`) zählen `reads`, nie `closes` des Bodys der Anfrage. Mutation x05 (vor `Senden` `a.Body` als `Close() error` schließen) bleibt **grün** (`gruen`, 2 Prozesse). Menge: Pfade zu U (nach M, nach R) × Body (gepuffert, `Laenge` 0, über der Grenze, Lesefehler). Auflage: Fall je Pfad zu U mit Zählstrom; die Zeile steht heute in keiner Tabellenzeile von Teil 2 (Z16 läuft „unverändert weiter“, deckt aber U nicht). |
| V-2 | LOW | Plan §3, Zeilen Z12 (U) (Plan Z. 90, 119) und Z19 (Z. 91) gegen akzeptiertes Negativ (a) (Z. 162) | Die Zeilen nennen als Mutation „Probe ohne den Code `ODR-E5001`“ und „Probe mit weniger als vier Aufrufen“, die ihr Fänger fange; das Negativ (a) sagt, ein einzelner Aufruf der Probe ohne `ODR-E5001` ändere nie ein Ergebnis. Gemessen: `ODR-E5001` aus der Probe (v30) und `ODR-E6001` bzw. `ODR-E6002` allein (v31, w03, w04) bleiben grün, weil jede Policy, die `ODR-E5001` freigibt, auch `ODR-E6003` freigibt und `ODR-E6001`/`ODR-E6002` dieselbe Kategorie tragen: die Mutationen sind äquivalent. Die Mutationen „ohne `ODR-E6003`“ (w01) und „ohne beide `unsupported`-Codes“ (w02) sind rot, aber über `TestPufferung` (Teil 1, `rot-paket`), nicht über `TestUpstreamProbe`, den die Zeile nennt. Folge: die Zeilen versprechen einen Fänger, den nur das Negativ ehrlich beschreibt. Der Planner zieht Zeile und Negativ gleich. |
| V-3 | LOW | `docs/plan/planning/done/slice-hybrid-service.md:295` gegen Plan §1 und §6 | Das Negativ „Zulassungsliste der Importe“ (V-2 von Teil 1) hat als Ausgang „weiter offen → akzeptiert, bis Teil 2 … einführt; Adresse: Teil 2“. Teil 2 führt weder in §1 noch in §6 noch in der DoD eine Zeile dazu (`AGENTS.md` §3.13: der Nehmer führt die Sendung). Der Code ändert die Liste (`quelltext_test.go`: `strings` und `outbound` zugelassen), sie ist jetzt eine Zulassungsliste, und das Negativ gilt weiter (cgo, `unsafe`, Reflexion). Adresse: Planner (Ausgang in §6 nennen). |
| V-4 | LOW | Plan §3 (akzeptierte Negative) | `SPEC-284` Nr. 3 f („ohne weiteren Aufruf von … `Entscheide`“ nach U) hat keinen Fänger und kein Negativ: `Entscheide` ist rein (`SPEC-261` Nr. 9), ein zusätzlicher Aufruf ist am konkreten Typ `*policy.Policy` nicht beobachtbar. Die Teile „kein M“ und „kein R“ sind gefangen (x01, x02, x04 rot). Das Negativ mit Grund und Folge (keine Folge: rein) fehlt im Plan. |
| V-5 | INFO | Closure | Alle DoD-Häkchen offen, §7 `Wird bei der Closure gefüllt`, Marke nicht gesetzt (bekannt, Planner): `closure-note-reviewer` ist noch nicht anwendbar. Die Zeile zum unveränderten Binary (DoD Punkt 7) steht bei der Closure aus; belegt ist sie (siehe DoD). |
| V-6 | INFO | `244827e8` nach `ffc59a19` | Die Randform „beide Fehler des Ports“ (`SPEC-284` Nr. 11) kam nach dem Code in die Spezifikation. §6 des Plans führt sie mit Ausgang *entschieden*, der Code stimmte mit der Entscheidung überein. Verfahren nach `AGENTS.md` §3.12/§3.13 gewahrt (Rückgabe, Zeile in §6); kein Befund. |

## Belege der Sensoren (selbst gefahren)

**`make gates`, einmal vollständig, mit frischen Läufen von build, test und test-race.** Aufruf: `flock /home/builder/work/odata-recorder/.wt/heavy.lock env DOCKER_BUILD="docker build --progress=plain --no-cache" make gates`. `DOCKER_BUILD` ist in den Fragmenten `?=`; die Umgebung gilt damit für jedes Image der Kette. Im Log tragen alle `docker build`-Zeilen `--no-cache`; als `CACHED` stehen nur die Stufen `WORKDIR /src` (Stufe `deps`/`race`/`lint`), kein `pruefung`-Lauf: `pruefung build: ok`, `pruefung test: ok` (333,9 s im Container), `pruefung race: ok` (334,4 s) liefen frisch. Host-Last 8 bis 12 von 4 CPUs. Exit 0, `git status` danach leer.

| Gate | Ergebnis |
|---|---|
| `a-check` | 0 Befunde (Hinweis: 94 Dateien in keiner Schicht, wie zuvor) |
| `docs-check` | `d-check: 864 Datei(en) geprüft, 0 Befund(e)` |
| `baseline-verify` | v6.18.1 OK, 54 Dateien |
| `build` | `pruefung build: ok` |
| `test` | `pruefung test: ok` |
| `test-race` | `pruefung race: ok` |
| `groessenmarke` | Hinweis `Marke steht aus: Schätzung 2220, Marke 3000`, grün (bekannt, nicht gemeldet) |
| `handbuch-check` | `pruefung handbuch: ok` |
| `katalog-check` | `katalog: ok` |
| `migration-check`, `schema-check` | ok |
| `lint-profil`, `lint` | grün (Exit 0 der Kette) |
| `zeichen-check` | ok, 1843 Dateien |

**Größenmarke.** `git diff --numstat 75249f56..HEAD` über `src/hexagon/ports/inbound/` und `src/hexagon/services/hybrid/`, Tests eingeschlossen: 1995 Einfügungen + 49 Löschungen = **2044**; der Plan nennt 2044 (§6). Über alle Pfade (mit `record/ohneantwort_test.go`, Plan, Spezifikation, Reports): 2389. Gegen die vorläufige Marke 3000 und die Schätzung 2220: unter beiden; Abstand zur Schnittgrenze 2223 nicht überschritten (Messung 2044). Metrik nach `SPEC-128` Nr. 1, wie §4 sie nennt; die fehlende gesetzte Marke ist bekannt (Planner).

**Rohbytes** (`AGENTS.md` §3.18): `LC_ALL=C grep -rPn '\\u[0-9A-Fa-f]{4}'` über `src/hexagon/services/hybrid/` und `src/hexagon/services/record/ohneantwort_test.go`: kein Treffer.

## Eigene Mutationen (`tools/harness/mutationslauf.sh`, Image `odata-recorder:race` frisch aus dem Gate-Lauf, `MUTATIONSLAUF_SPEICHER=4g`, `--cpus=4`, unter der Sperre; Listen `liste-v`, `liste-w`, `liste-x` im Scratchpad)

63 Mutationen an Stellen, die ich selbst wählte (Elemente, die weder Plan noch Review nannten: Kopfzeilen einzeln, Tokens von `Connection`, Wert-Struktur, Status je Fehlerart, Segmentgrenze, kürzester Root, Wurzel `/`, Pfad ohne Dienst, Aufrufstellen H3 und H5, Rückfall, Wiederholung, Body-Schließen):

| Liste | Mutationen | rot | rot-paket | grün | Übersetzung |
|---|---|---|---|---|---|
| v (40) | 40 | 36 | 0 | 3 | 1 (v35, unbenutzter Import, nicht gewertet; als w05 neu: rot) |
| w (16) | 16 | 11 | 3 | 2 | 0 |
| x (7) | 7 | 6 | 0 | 1 | 0 |

Die Klasse `nebenläufig` (x06 Schreibzugriff über `copy`, x07 über Zeiger auf ein Element, beide nicht an `TestQuelltext` sichtbar): `TestGleichzeitigUpstream` 20 von 20 Prozessen mit Detektor und 20 von 20 ohne, je Mutation 40 Prozesse; Host-Last 6,7 von 4, Messender: ich. Die Zusage (Nr. 13) gilt als gefangen; zwei Formen, die der Plan nicht nannte.

Grüne Mutationen und ihre Würdigung:

| Mutation | Ergebnis | Würdigung |
|---|---|---|
| v27 `Pfad` an H5 gleich OData-Pfad | grün | akzeptiertes Negativ (a), Grund und Folge im Plan; bestätigt (Entscheide liest den Pfad nur mutierend, mutierend mit `ODR-E5001` ergibt `Bleiben`) |
| v30 Probe ohne `ODR-E5001`; v31, w03, w04 Probe nur mit einem der beiden `unsupported`-Codes | grün | äquivalent (V-2) |
| x05 Body der Anfrage vor `Senden` geschlossen | **grün** | V-1 (MEDIUM) |

Rot und aus dem benannten Grund gesehen unter anderem: Query leer (v01), Header ungefiltert (v02, v03), Kennzeichnung `replay` statt `upstream` (v04), 504 ohne Ursache (v05), 500 als 502 (v06, w12), nur `Canceled` als Kontextfehler (v07, w10), Segmentgrenze fehlt (v08), `trailer` und `keep-alive` bleiben (v09, v10), Tab am Rand (v11), Token und `Connection` mit Beachtung der Schreibweise (v12, v13), leere Werte entfernt (v14), `expect`, `host`, `content-length`, `x-method-override`, `accept-encoding` je einzeln (v15 bis v18, w14), Präfix `x-recorder-s` (v19), `context.WithoutCancel` an H6 (v20), Root gleich Pfad abgewiesen (v21), kürzester Root (v22), `Ziel` des ersten Dienstes (v23), `verwirf` nicht bei Kontextfehler (v24), Leser von M an H3 und von R an H5 (v25, v26), `Methode` fest an H3 und H5 (v28, v29), Leser leer (v32), `Laenge` bei fehlendem Body (v33), Grenze `>=` (v34), Wertliste geteilt (w05), Case-Vergleich des Kennzeichnungs-Headers (v36), Vorrang von `ErrUpstreamZeit` im Record-Dienst (v37), `Upstream` nil zugelassen (v38), `Ziel` an H3 leer (v39), Rest ohne `/` (v40), `Connection`-Tokens in der Antwort ignoriert (w06), `Laenge` der Anfrage an U (w07), `Methode` fest an `Senden` (w08), 3xx zu 200 (w09), `ErrUngueltig` ohne Code (w11), Body bei `Laenge` 0 gelesen (w13), Root `/` ohne führenden `/` (w15), `Ziel` ohne Dienst (w16), Mock nach U, Replay nach U, Wiederholung, Rückfall auf R (x01 bis x04).

## DoD gegen Code und Spezifikation

| DoD-Punkt (Plan §2) | Ergebnis |
|---|---|
| 1. Upstream-Fallback nur nach Freigabe; `X-Recorder-*` entfernt; Body aus dem Puffer; `Entscheide` vor jedem Senden mit demselben `Ziel`; Dienstwahl; drei Override-Header nie | gehalten: `upstream.go` und `verteiler.go` stimmen mit `SPEC-284` Nr. 3 c bis f, 4, 5, 8 überein; die Mutationen v02, v15 bis v19, v21 bis v23, v25 bis v29, v32 bis v34, v39, v40 und w07, w08, w13, w15, w16 sind rot. Offen: V-1 (Body der Anfrage wird nie geschlossen: Code gehalten, Fänger fehlt). |
| 2. Antwort und Ströme; Fehler als `ODR-E4001` (504, 502) und `ODR-E9001` (500), ohne Wiederholen und Rückfall | gehalten: Nr. 9 bis 11 gegen `upstreamFallback`, `upstreamFehler`, `gekennzeichnet`; Mutationen v03 bis v07, v24, v35/w05, v36, w06, w09 bis w12, x01 bis x04 rot; Vorrang bei `errors.Join` (Nr. 11) im Code und im Record-Dienst (v37 rot). |
| 3. `make gates` grün | belegt (oben, frisch) |
| 4. Review, Report liegt vor | belegt (zwei Runden unter `docs/reviews/`); Inhalt nicht als Beleg gewertet |
| 5. Doku-Update | `SPEC-284` Nr. 11 und `SPEC-121` Nr. 11 nachgezogen; Plan §1, §3, §6 und Kopf stimmen mit dem Code überein (bis auf V-2, V-3, V-4); `architecture.md` unberührt, Nachzug lag bei Teil 1 |
| 6. Handbuch/Katalog | Binary unverändert: der Diff berührt weder `src/bootstrap`, `src/adapters` noch `src/hexagon/model`, `katalog-check` und `handbuch-check` grün; die Zeile in §7 steht bei der Closure aus |
| 7. bis 10. Closure-Notiz, Register, Ausgang je Risiko, Paarungen | offen (Closure, Planner); §6 trägt je Risiko einen Ausgang oder *offen* mit Adresse (V-3 fehlt dort) |

## Tabelle Zusage · Fänger gegen den Vertrag (§3 des Plans)

Jede Nummer von `SPEC-284` mit U hat eine Zeile: Nr. 1 (Z01 U, D), Nr. 3 c bis f (Z07, Z08, Z09, Z10), Nr. 4 (Z19, Z01 D), Nr. 5 a, c (Z12 U, Z24), Nr. 8 a bis c (Z20 bis Z24), Nr. 9 (Z26, Z27), Nr. 10 (Z28), Nr. 11 (Z25), Nr. 12 (Z11), Nr. 13 (Zeile „Nr. 13“), *Ungeprüft* (Z35). Ohne Zeile oder Negativ bleiben: „Body der Anfrage nie geschlossen“ auf dem Weg U (V-1) und „kein `Entscheide` nach U“ (V-4). Die Fänger laufen als Schleifen über die Mengen mit Zählprüfung (Header: 60 Namen in vier Schreibweisen, Trenner 131, Pfade 32 und 16), und die Fixtures gleichen nicht dem Wert der Mutation (an Stichproben v08, v22, v34, w15 geprüft).

## Raster (§6)

Das Raster steht in §6 von `slice-hybrid-service` (Teil 1), von Teil 2 in §6 (Z. 203) als gehörig benannt. Geprüft: 4 Zeilen × 5 Spalten, keine leere Zelle, jede Zelle mit `SPEC-284 Nr.`, `SPEC-261 Nr.` oder `gilt nicht` mit Grund. Die Zusatztabelle (Ergebnis des Mock × Policy-Zustand) trägt je Zeile einen Fänger. Die Randform „beide Fehler des Ports“ steht als eigene Zeile in §6 mit Ort `SPEC-284` Nr. 11.

## Plan-vs-Code-Diff (`git diff main...HEAD`, 23 Dateien)

Berührt: `src/hexagon/services/hybrid/` (Produkt: `upstream.go` neu, `verteiler.go`, `puffer.go`, `antwort.go`, `fehler.go`; Tests: 9 Dateien), `src/hexagon/services/record/ohneantwort_test.go` (nur Test), `spec/spezifikation.md` (zwei Zeilen: Nr. 11 beider Verträge), Plan, `welle-hybrid.md`, `roadmap.md` (Pfade), `done/slice-hybrid-service.md` (reine Linkberichtigung nach `in-progress/`), zwei Review-Reports. Kein Eingriff in `Makefile`, `Dockerfile`, `harness/`, `tools/`, `.golangci.yml`, `.a-check.yml`, `src/bootstrap`, `src/adapters`, `src/hexagon/model` und `src/hexagon/ports`. Produktcode: `upstream.go` 137 Zeilen plus rund 60 in `verteiler.go` und `puffer.go`, passend zur Schätzung von rund 300. Das Paket importiert zusätzlich `outbound`, `model` und `strings`; kein Adapter, kein anderer Dienst (`TestQuelltext`, `TestImportErlaubt`).

## Akzeptierte Negative

(a) `Pfad` an H5 und einzelne Aufrufe der Probe: Grund und Folge stehen, die Äquivalenz ist gemessen (v27, v30, v31). (b) `ODR-E5002` zu U: Grund (`Bleiben` immer), Folge keine; plausibel. (c) bis (e) und `Z35`: Grund und Folge stehen. Nebenläufiges Negativ (e): Grund und Folge stehen. Ergänzungsbedarf: V-4.

## Pfad und Commit

Report: `/home/builder/work/odata-recorder/.wt/hybrid/docs/reviews/2026-11-18-slice-hybrid-service-upstream-verifikation.md`.
