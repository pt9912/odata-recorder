# Verifikation slice-prefer-return

**Rolle:** Verifier (einmal). **Datum:** 2026-11-18. **Stand:** Zweig `slice-prefer-return`, HEAD `9a6ee7dc`, Basis `5a98ef2a`. **Bezug:** `LH-FA-07` (`LH-FA-07.b`), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), `SPEC-285` bis `SPEC-289`.
**Plan:** [slice-prefer-return](../plan/planning/done/slice-prefer-return.md). **Modell:** claude-sonnet-5-5.
Die Reviews (Runde 1 und 2) sind nicht als Beleg gelesen; alles unten ist selbst gefahren oder gelesen. Der Lauf teilt den Host mit Reviewer-Runde 2 (Load average 4,7 bis 14 von 4 CPUs).

## Urteil

**PASS mit einer Auflage (MEDIUM, Adresse Planner) und zwei LOW.**
Alle DoD-Punkte, die vor dem Review zu belegen sind, sind belegt: `make gates` ist frisch grün (ohne Cache), die Größenmarke hält (2123 von 3500), der Code folgt `SPEC-285` bis `SPEC-289` und [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), die Tabelle *Zusage · Fänger* deckt den Vertrag, und eigene Probe-Pakete über die Klasse der Header-Formen (5 Mio. Fälle) und sechs Mutationen an Stellen, die ich selbst wählte, fanden keinen Fehler im Code.
Die Auflage: Der Entscheid „kein Handbuch-Kapitel“ (Weg B) ist mit dem Lastenheft und `SPEC-200` vereinbar, widerspricht aber noch dem Welle-Plan `welle-etag` (V-1).

## Findings

| Nr. | Schwere | Ort | Befund |
|---|---|---|---|
| V-1 | MEDIUM | `docs/plan/planning/welle-etag.md` Zeile 47 (Closure-Kriterium) und Zeile 64 (*Pflicht Handbuch*) gegen Plan §6 (Entscheidung Weg B) | `welle-etag` verlangt: „Das Handbuch deckt alles ab, was das Binary dieser Welle kann (ETag, `If-Match`, `Prefer`, …)“ und „In dieser Welle betrifft [die Pflicht, ein Handbuch-Kapitel mit Beispiel nachzuziehen] alle drei Slices (ETag-Header, `412`, `Prefer`)“. Weg B (Architect, `5b09dd6a`) schreibt für `Prefer` kein Kapitel; das Handbuch nennt `Prefer` nur als Vary-Header in `MATCH_VARY_HEADERS` (Zeile 79 von `docs/user/benutzerhandbuch.md`). Die Welle kann ihr Kriterium so nicht erfüllen, und kein Dokument sagt, dass sie es ändert. §6 des Plans adressiert nur die Zeile `ETAG_ENABLED` (Planner), nicht das Welle-Kriterium (`AGENTS.md` §3.9: betroffene Folgedokumente im selben Commit; §3.13: Adresse). Das Lastenheft ist nicht verletzt (siehe unten), die Welle-Datei ist niedriger im Rang; der Widerspruch gehört trotzdem vor die Welle-Closure. **Adresse:** Planner, Auflage: `welle-etag` Zeile 47 und 64 auf „`Prefer` hat kein Handbuch-Kapitel (`SPEC-200` Nr. 1, Entscheidung Architect)“ nachziehen oder den Architect `SPEC-200` um eine Form für HTTP erweitern lassen. Kein Blocker für die Closure des Slice. |
| V-2 | LOW | Plan §6 (Punkt „Handbuch-Kapitel der DoD“) | Der Punkt hat Ausgang *entschieden* und nennt die Adresse Planner nur für die Zeile `ETAG_ENABLED`; die Sendung „Welle-Kriterium und *Pflicht Handbuch*“ (V-1) steht nicht darin. Mit V-1 im selben Commit eintragen. |
| V-3 | LOW | Plan §4, Zeile 249 (Rückführungen) | „Die Tabelle *Zusage · Fänger* liegt vor (26 Zeilen)“; die Tabelle hat 30 Zeilen (1 bis 28 mit 6a und 9a, gezählt). Zeile 231 erklärt den Stand der Marke („nach der Fix-Runde 30 Zeilen“), Zeile 249 nicht. Zustandsangabe nachziehen. |
| V-4 | INFO | Plan §2, §6, §7 | Die Punkte Review, Doku-Update, Closure-Notiz, Register, Ausgang jedes Risikos aus §6 und Paarungen stehen offen; §6 trägt fünf Ausgänge *offen* (Mengen unbeschränkt, Zustand, Fakes, Raster, Body gleich `GET`). Das ist der erwartete Stand vor der Closure und kein Befund. Sie sind bei der Closure aufzulösen; das Raster und die Ausgänge der Zeilen 1, 2 sind belegt (unten). |
| V-5 | INFO | eigene Mutation V7 | Von sieben geplanten eigenen Mutationen lief die siebte (V7, `ETag` bei Minimal fehlt) nicht: der Container lief ab (`sleep 1200`). Sechs sind rot (V1 bis V6). `ETag` bei `Minimal` fängt `TestEtagAusUndAn` und `TestPreferAntwortTabelle` (Zeile 10, Mutation s05 des Plans, vom Implementer rot gesehen); ich habe sie nicht selbst wiederholt. |

## Belege der Sensoren (selbst gefahren)

**`make gates`, einmal vollständig und frisch.** Aufruf: `flock /home/builder/work/odata-recorder/.wt/heavy.lock env DOCKER_BUILD="docker build --progress=plain --no-cache" make gates`, Ausgang 0. (Davor ein Lauf mit warmem Cache, ebenfalls Ausgang 0.)

| Gate | Ergebnis |
|---|---|
| `a-check` | 0 Befunde (Hinweis: 94 Dateien in keiner Schicht, unverändert) |
| `docs-check` | `d-check: 874 Datei(en) geprüft, 0 Befund(e)` |
| `baseline-verify` | v6.18.1 OK, 54 Dateien |
| `build` | `pruefung build: ok` |
| `test` | `pruefung test: ok` |
| `test-race` | `pruefung race: ok` |
| `groessenmarke` | `ok, Messung 2123 von Marke 3500` (Hinweis `Tabelle ergibt 2905, Schätzung 2585`, vom Plan §4 erklärt); eigene Rechnung `git diff --numstat 5a98ef2a..HEAD` über die Pfade des Blocks: 2123 |
| `handbuch-check` | `pruefung handbuch: ok` |
| `katalog-check` | `katalog: ok` |
| `migration-check`, `schema-check`, `lint-profil`, `lint` | grün (Ausgang 0 der Kette) |
| `zeichen-check` | ok, 1861 Dateien |

Nicht gefahren (kein Gate dieses Auftrags): `make gegenproben`, `make test-integration`, `make test-contract`. Der Slice ändert weder Gates noch Gegenproben-Eingaben (`tools/`, `harness/`, `Makefile`, `Dockerfile`, `.golangci.yml` unverändert gegen `f20f709a`).
Rohbytes: alle neun Dateien unter `src/` des Slice mit `LC_ALL=C grep -cP '\\u[0-9A-Fa-f]{4}'` gegengelesen, je 0; keine `//nolint`-Direktive im Diff.

## DoD gegen Beleg

| DoD-Punkt (Plan §2) | Ergebnis | Beleg |
|---|---|---|
| `return=minimal` und `return=representation` für POST, PATCH, PUT: Status, Body, `Location`, `OData-EntityId`, `Preference-Applied` nach `SPEC-286`; ohne `Prefer` unverändert | erfüllt | Code `schreiben_antwort.go` Zeile für Zeile gegen die Tabelle von `SPEC-286` Nr. 1 gelesen (6 Zeilen: Status 201/200/204, Body, `Content-Type`, `Location` nur POST, `OData-EntityId` nur Minimal, `Preference-Applied` nur bei Wahl; kein `Vary`); eigene Mutationen V1 bis V6 rot (unten) |
| Randformen des Headers: je Zelle ein Ort und ein Fänger; nicht unterstützter Wert ohne Fehler | erfüllt | Grammatik (`SPEC-285` Nr. 2 bis 4) gegen `model.LiesRueckgabe` gelesen; Probe-Pakete (unten) 5 Mio. Fälle ohne Abweichung; Raster in §6 hat je Zelle einen Ort (geprüft: jede genannte Nummer existiert in `SPEC-285` bis `SPEC-289`, `SPEC-194`, `SPEC-198`); Zusatztabelle für die Modi vorhanden |
| Unveränderter Zustand, dasselbe Fehlerobjekt bei Fehlschlag | erfüllt | `Prefer` wird erst in `antwort` (Schritt (l)) gelesen, nach Schritt (k): vor dem Lesen entsteht kein Fehler, der am Header hängt (`SPEC-287` Nr. 1, 2); `Put`, `Advance` und `Transaktion` sind in `schreiben.go` unberührt (Diff: nur Header-Map in `mutation`, `antwort` ausgelagert); Tests `TestPreferZustand`, `TestPreferFehlerGleich` vorhanden, durch V4 mitgetragen |
| `make gates` grün | erfüllt | frischer Lauf, Ausgang 0 |
| Review, Doku-Update, Closure-Notiz, Register, Ausgänge, Paarungen | offen, noch nicht fällig | V-4 |

Der Plan führt §2 mit drei Liefer-Punkten (Grenze der Baseline: höchstens 3); der Punkt „Handbuch-Kapitel“ ist bewusst entfallen (V-1).

## Handbuch (Weg B) gegen Lastenheft und `SPEC-200`

- **Lastenheft:** `spec/lastenheft.md` enthält weder `LH-FA-07.b` noch das Wort Handbuch oder Dokumentation (`grep -n -i "handbuch\|benutzerdoku\|dokumentation"`: kein Treffer). `LH-FA-07` (Zustandsbehaftete CRUD) und seine Akzeptanzkriterien nennen `Prefer` nicht. `LH-FA-07.b` ist die Matrix in `spec/spezifikation.md` und führt `Prefer` als Soll-Zeile (`SPEC-049`); sie verlangt kein Handbuch-Beispiel. Das Lastenheft verlangt also keines; Weg B verletzt es nicht.
- **`SPEC-200` Nr. 1:** Das Handbuch beschreibt Verhalten des Binary nur als gebundenes Beispiel (Klasse K1, Aufruf des Binary mit Ausgabe und Ausgangswert); ein HTTP-Request an den Mock ist keine solche Aussage. Weg B ist mit dem Wortlaut vereinbar, `make handbuch-check` bleibt grün.
- **Folge im Betrieb:** Ein Anwender findet `Prefer` im Handbuch nur als Vary-Header. Die Ist-Aussage steht in der Spezifikation und in den Tests; das gilt als akzeptiertes Negativ des Architect (Plan §6) und bleibt gleich dem Bestand (die Zeile `ETAG_ENABLED`).
- **Offen:** der Welle-Plan (V-1, V-2).

## Plan gegen Code (Diff `5a98ef2a..HEAD`, Pfade des Slice)

Der Diff gegen die Basis trägt auch den Merge von `slice-hybrid-service-upstream` (Dateien unter `src/hexagon/services/hybrid/` außer `prefer_test.go`, `record/ohneantwort_test.go`, Roadmap, Register); die eigene Lieferung des Slice steht in `git diff f20f709a..HEAD` und besteht aus neun Dateien unter `src/` und der Spezifikation.

| Plan §3 | Code | Ergebnis |
|---|---|---|
| `mock/`: `schreiben_antwort.go` neu (`antwort`, `antwortMitEntity`, `antwortOhneBody`, `adresse`, `kopf204`), `mutation.kopf` | vorhanden, `schreiben.go` −54 netto | gleich |
| `model/rueckgabe.go` neu (`Rueckgabe`, `LiesRueckgabe`), kein neuer Typ am Port | vorhanden, 131 Zeilen; `ports/` unverändert | gleich |
| `httpodata/`: Code unverändert, `prefer_test.go` neu | Diff nur die Testdatei | gleich |
| `bootstrap/prefer_ende_zu_ende_test.go` neu | vorhanden | gleich |
| `schreiben_antwort_test.go` update (`TestPreferIgnoriert` auf Fälle ohne Wahl) | 7 Zeilen | gleich |
| `hybrid/prefer_test.go` neu, kein neuer Fake | vorhanden; Produktcode des Hybrid-Dienstes für `Prefer` unverändert | gleich |
| `model/rueckgabe_test.go`, `mock/prefer_test.go` neu | vorhanden (613 und 908 Zeilen) | gleich |
| `spec/spezifikation.md`: Vertrag vom Architect | `SPEC-285` bis `SPEC-289` vor dem ersten Code-Commit (Basis `5a98ef2a`); danach nur ein Punkt in *Ungeprüft* von `SPEC-285` (`9a6ee7dc`, Rückfrage (d), im Plan §6 entschieden) | gleich |

Alle in §3 und §6 genannten Testnamen existieren im Baum (Prüfung per `grep "func <Name>("` über alle Namen des Plans: keiner fehlt).
Kein Code außerhalb der Pfade des Plans, keine neue Aufrufstelle am Schreiber über S24 bis S26 hinaus, kein neuer Port.

## Tabelle *Zusage · Fänger* gegen den Vertrag

Jede Nummer der Verträge ist einer Zeile zugeordnet; aufgelöst gegen den Wortlaut:

| Vertrag | Zeile(n) der Tabelle | Befund |
|---|---|---|
| `SPEC-285` Nr. 1 (nie ändern, nie Fehler, nie Panik, Reihenfolge, linear) | 5, 9, 9a | gedeckt; linear nur für sieben Formen, der Rest steht in *Ungeprüft* (eigener Punkt, `9a6ee7dc`) |
| Nr. 2 (OWS, Paare, Backslash, `Rest`, unlesbar) | 3, 6, 6a, 7, 8 | gedeckt |
| Nr. 3 (erste Instanz, Schreibweise, jede andere Präferenz) | 2, 4 | gedeckt |
| Nr. 4 (Beispiele) | 1 | gedeckt (36 Zeilen) |
| `SPEC-286` Nr. 1 (Tabelle, Adresse, `Preference-Applied`, leerer Strom) | 10, 11, 12 | gedeckt |
| Nr. 2 (kein `Vary`) | 10 | gedeckt (Mutation s04) |
| Nr. 3 (wann `Preference-Applied`) | 13 | gedeckt |
| Nr. 4 (Wahl und Zustand) | 15 | gedeckt |
| `SPEC-287` Nr. 1 bis 7 | 17, 14, 16, 15, 21, 24, 27 | gedeckt; Nr. 1 (Zugriff) ist ein benanntes Negativ mit Grund und Folge (`SPEC-289` Nr. 3) |
| `SPEC-288` Nr. 1 bis 3 | 18, 19, 20, 21, 22 | gedeckt; Allaussagen als Schleifen (Stellen × Fehler, 98 und 36 Elemente) |
| `SPEC-289` Nr. 1, 2 | 28, 23 | gedeckt; Zeile 23 als Wirkungs-Fänger, 20 von 20 Prozessen rot |
| Adapter und Handbuch | 25, 26 | 25 gedeckt; 26 „gilt nicht“ mit Grund (siehe oben) |

Akzeptierte Negative (Plan §3): jedes nennt Grund und Folge im Betrieb (Prüfgrenzen der Generatoren, `tchar` an den 226 gleichwertigen Bytes, Zugriff bei `GET`, `HEAD`, `DELETE`, Zeitform ausser den sieben, Zustand innerhalb der Transaktion). Geprüft gegen `SPEC-285` *Ungeprüft* und `SPEC-289` Nr. 3: gleich im Wortlaut, nichts, was der Plan zusagt und die Spezifikation nicht trägt.

## Eigene Prüfung gegen die Klasse (`AGENTS.md` §3.17)

**Klasse 1: jede Zeichenfolge von `Prefer` (unbeschränkt).** Zwei Probe-Pakete in `src/hexagon/model/` auf einer Kopie des Baums im Container (`odata-recorder:race`, `--network none`, `--memory=4g --cpus=4`, Kopie per `docker cp`, nichts im Repo gelassen): ein unabhängiger Orakel-Leser, der die Grammatik von `SPEC-285` Nr. 2 und 3 in anderer Form liest (erst Elemente trennen, dann je Element lesen; keine Code-Teilung mit dem Dienst).
- Ungerichtet: 3.000.000 Folgen aus 23 Atomen (Namen, `=`, `;`, `"`, Backslash, `,`, Leerzeichen, Tabulator, `0xA0`, `0x0B`, `0xC2 0xA0`, Wörter), 0 bis 3 Zeilen: 0 Abweichungen (Verteilung der Antworten: Keine 2.999.162, Minimal 548, Entity 290; dünn bei den Wahlen, deshalb das zweite Paket).
- Gerichtet: 2.000.000 Folgen aus 18 Elementformen, zusammengesetzt mit drei Trennern und bis zu zwei Mutationen (Einfügen, Streichen, Ersetzen eines Zeichens aus 13 Zeichen): 0 Abweichungen (Keine 1.127.754, Minimal 569.213, Entity 303.033).
- Das Probe-Paket fängt: eine Mutation „spätere Instanz zählt“ im Produktcode (nachgestellt in der Kopie) macht das ungerichtete Paket rot (fünf Abweichungen bis zum Abbruch). Die Probe ist damit nicht blind.
Folge: Die Klasse der Zeilenformen ist nach Wortlaut des Vertrags gedeckt; die Länge der Folgen (bis 8 Atome je Zeile, bis 3 Zeilen) bleibt Grenze, wie im Plan benannt.

**Klasse 2: Antwort je Zelle (3 Methoden × 3 Wahlen × Etag).** Sechs eigene Mutationen an `schreiben_antwort.go` (in einer Kopie, per Perl; Pakete `mock` und `bootstrap`), Stellen, die ich wählte, keine der Plan-Kennungen:

| Kennung | Mutation | rot |
|---|---|---|
| V1 | `OData-EntityId` an der `200` von `PATCH`, `PUT` · Entity | ja: `TestPreferAntwortTabelle`, `TestEtagAusUndAn`, `TestEtagPost` |
| V2 | `Preference-Applied` bei `Minimal` fehlt | ja: `TestPreferAntwortTabelle`, `TestPreferApplied`, `TestPreferNamenUndReihenfolge` |
| V3 | `Preference-Applied` bei jeder Wahl außer `Minimal` (auch `POST` · Keine) | ja: `TestPreferAntwortTabelle`, `TestEtagPost`, `TestEtagAusUndAn` |
| V4 | `DELETE` liest `Prefer` | ja: `TestPreferAppliedNurBeiErfolg`, `TestPreferLesendOhneWirkung`, `TestPreferSchreiberAufrufe` |
| V5 | Status der `200` von `PATCH`, `PUT` · Entity als `201` | ja: `TestPreferAntwortTabelle`, `TestPreferGleichzeitig` |
| V6 | `Location` an der `204` von `POST` · Minimal fehlt | ja: `TestPreferAntwortTabelle`, `TestPreferAdresseNieAusHost` |

6 von 6 rot, je durch den benannten Fänger der Zeile 10 oder 13, nicht durch einen Aufbaufehler (die Fehlermeldungen stammen aus den Vergleichen der Tabelle, nicht aus Panik oder Kompilierfehler). V7 (V-5) lief nicht.

## Steering-Loop und Richtung

- Kein Befund verkörpert eine neue Regel; V-1 ist eine Zeigerlücke zwischen Welle und Slice (`AGENTS.md` §3.9, §3.13), ein Register-Eintrag ist nicht geprüft und Sache des Planner.
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Go, Hexagon): Code in `src/hexagon/` ohne Import nach außen; `make a-check` 0 Befunde; keine neue Abhängigkeit (`go.mod`, `go.sum` unverändert).

## Für den Validator (End-Nutzer-Wert)

**Ja, Wert liegt vor.** Ein Client, der `Prefer: return=minimal` oder `return=representation` an `POST`, `PATCH` oder `PUT` des Mock sendet, erhält die Antwort, die er wählt: `204` ohne Body mit `OData-EntityId` (und `Location` bei `POST`) oder `200` beziehungsweise `201` mit der Entity, die dem folgenden `GET` gleicht (ETag eingeschlossen), und `Preference-Applied` nennt die angewandte Wahl. Ohne den Header und bei jedem anderen Wert bleibt das Muss-Verhalten ohne Fehler; Groß- und Kleinschreibung, Leerraum, mehrere Präferenzen und mehrere Header verhalten sich nach RFC 7240 (erste Instanz zählt). `Prefer` ändert den Zustand nie, und bei Fehlern ist die Antwort mit und ohne Header gleich.
Grenzen, die der Validator kennt: Das Handbuch nennt `Prefer` nicht (Weg B, V-1); `Prefer` hat im Replay keine Wirkung und wird im Hybrid-Dienst unverändert an Fallback und Upstream gereicht, `Preference-Applied` kommt dort nur, wenn die Stufe es selbst setzt; andere Präferenzen (`respond-async`, `odata.maxpagesize`) bleiben wirkungslos ohne Fehler.
