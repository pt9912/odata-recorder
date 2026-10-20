# Verifikation: slice-redaction-match-key

* Datum: 2026-10-20
* Rolle: verifier (frischer Kontext; kein Code, keine Spec, kein Plan geändert)
* Gegenstand: `git diff 749a5d4..HEAD`, Stand `f984c1f` (Commits `c9c20ae`, `760f1be`, `5e188dd`, `b16cb1e`, `aa5fe58`, `f984c1f`)
* Plan: `docs/plan/planning/done/slice-redaction-match-key.md` (Welle `welle-record-replay`)
* Maßstab: DoD (§2), `SPEC-115`, `SPEC-116`, `SPEC-117`, `SPEC-045`, Tabelle Zusage·Fänger (§3), Raster (§6), `AGENTS.md` §3.7, §3.9 bis §3.16, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md); Bezug [LH-FA-05](../../spec/lastenheft.md), [LH-FA-14](../../spec/lastenheft.md), [LH-QA-02](../../spec/lastenheft.md)
* Eingang: Review `docs/reviews/2026-10-20-slice-redaction-match-key-review.md` (F-1 bis F-13), Nachlauf in `aa5fe58` (Architect) und `f984c1f` (Implementer)
* Belege selbst gefahren: `make test` (Basis grün, Exit 0); 14 Mutationen über `make test` im Docker (netzlos), jede mit `git checkout -- src` zurückgenommen, Arbeitsbaum danach sauber; eine Wegwerf-Testdatei (gelöscht); `git diff --numstat`; `make docs-check` und `make gates` (siehe Abschnitt 8).

## Urteil

**Nicht freigegeben, ein Befund (V1, MEDIUM, Zuständigkeit Implementer).** Code und Spezifikation stimmen überein; alle Findings des Reviews sind behandelt und ihre Mutationen rot. Offen ist eine Zusage ohne Fänger, die erst die Mutation „Sortierung der Query-Paare ignoriert den Schlüssel“ aufdeckt (V1), sowie Closure-Pflichten des Planners (Abschnitt 7). Am Code ist nichts zu ändern; V1 ist durch einen Testfall zu schließen. Danach ist der Slice aus Sicht des Verifiers freigegeben.

## 1. DoD (§2) gegen Code, Spec und Tests

| DoD-Punkt | Befund |
|---|---|
| Redaction von Headern (Default plus `REDACT_HEADERS`), `raw_query`, `Location`/`Content-Location`; Marker `[REDACTED]`; Parser der Pfadmuster, auch vom Start gerufen; `key_redacted` | erfüllt: `redaction.go` (`RedigiereHeader`, `RedigiereQuery`, `OhneRedigierte`, `ParseJSONPfad`, `Marker`), `BildeSchluessel` liefert `KeyRedacted`; `istJSONPfad` in `src/bootstrap/konfig_werte.go` ruft `model.ParseJSONPfad` (Mutation M12 rot) |
| Match-Key: Tupelkodierung, `normalized_path`, sortierte Multimenge der Query-Paare, `literal`/`reject`, Vary-Header, Body-Hash `bytes`/`json-canonical`, volatile und redigierte Header ausgenommen | erfüllt in `schluessel.go`, `schluessel_json.go`; `TestKeyKodierungHandgebaut` baut die Folge aus `SPEC-116` Nr. 1 von Hand nach. Lücke bei der Sortierung der Paare: V1 |
| Replayfähigkeit nach `SPEC-045` als Funktion (`REPLAY_REDACTED_BODIES`) | erfüllt: `replay.go` entspricht `SPEC-117` Nr. 1 bis 3 und der Matrix in `SPEC-045` (`complete` und `verified`/`absent` und `none`, oder `applied` bei `true`; alles andere falsch) |
| `make gates` grün | selbst gefahren, siehe Abschnitt 8 |
| Review durchgeführt, Report liegt vor | erfüllt: `docs/reviews/2026-10-20-slice-redaction-match-key-review.md`; der Report gilt für `5e188dd`, die Änderungen danach (`f984c1f`) prüft dieser Bericht per Mutation (Abschnitt 4); das Häkchen setzt der Planner |
| Doku-Update (öffentlicher Vertrag) | erfüllt: `SPEC-115` bis `SPEC-117` samt Änderungsverzeichnis (Zeilen 1240 bis 1242), Zeiger in `LH-FA-05.a` (`spec/spezifikation.md` Zeile 126) und `LH-FA-14.a` (Zeile 377) |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen, Sache des Planners (Abschnitt 7) |

Die Schichtenrichtung ist eingehalten: `src/bootstrap/` importiert `src/hexagon/model/`, nicht umgekehrt ([ADR-0002](../plan/adr/0002-architektur-gate-a-check.md); `make a-check` im Gate-Lauf grün).

## 2. SPEC-115 bis SPEC-117 gegen Code (alle Nummern)

| Stelle | Zusage | Code und Fänger |
|---|---|---|
| `SPEC-115` Nr. 1 | Marker `[REDACTED]` einziger Platzhalter | `Marker`; `TestMarker` |
| Nr. 2 | Header: genau ein Wert, ASCII-Abgleich, Zusammenführen, Nicht-Token-Namen, Kopie | `RedigiereHeader`, `fasseNamenZusammen`; `TestRedigiereHeaderWerte/Namen/Unveraendert/Schreibweise/NichtToken/Kopie`; Kelvin-Fall: M5 rot |
| Nr. 3 | `Location`/`Content-Location`: Query ersetzt, `#` vor `?`, ganz redigierter Header gewinnt | `redigiereURLWert`; `TestRedigiereLocation` |
| Nr. 4 | Query-Zerlegung, Meldung „geändert“, Idempotenz, scheitert nie | `RedigiereQuery`; `TestRedigiereQueryRandformen/Geaendert/UngueltigeKodierung`, `TestRedigiereIdempotent` (jetzt auch über `locationFaelle`) |
| Nr. 5 | `OhneRedigierte` | `TestOhneRedigierte` |
| Nr. 6 | `$..name`, Bedeutung, Start nutzt den Parser | `ParseJSONPfad`, `Passt`; `TestParseJSONPfad`, `TestJSONPfadPasst`, `TestKonfigJSONPfadWieModell` (M12 rot) |
| Nr. 7 | `key_redacted` | `TestKeyRedacted` (M7 rot) |
| `SPEC-116` Einleitung | ein Aufruf, redigiert selbst; Grenze des Bodys | `BildeSchluessel`; Kommentar an `SchluesselEingabe.Body` zeigt auf die Grenze; kein Fänger, als akzeptiertes Negativ im Plan benannt |
| Nr. 1 | Kodierung | `tupel`; `TestKeyKodierungHandgebaut`, `TestKeyEindeutig` |
| Nr. 2, 3 | Methode, Service, Pfad | `TestKeyKopf`, `TestKeyPfad` |
| Nr. 4 | Paare, Dekodierung, Art 2, Sortierung | `zerlegePaare`; `TestKeyQueryReihenfolge/Zerlegung/Dekodierung`, `TestKeyUngueltigeKodierung`, `TestKeyRedigiertVorKodierung`. Sortierung: V1 |
| Nr. 5 | Vary: Menge, Ausnahmen, redigierte Werte, Anzahl der vorhandenen Namen | `vary` mit `header` aus `RedigiereHeader`; `TestKeyVary`, `TestKeyVaryAusnahme`, `TestKeyVaryRedigiert` (M1 rot), `TestKeyVaryMengeReihenfolge` (M2 rot), Anzahl (M4 rot) |
| Nr. 6 | Body | `schluessel_json.go`; `TestKeyBodyBytes/Kanonisch/Rueckfall/GleicheNamenReihenfolge` (M3 rot) |
| Nr. 7, 8 | `KeyRedacted`, redigierte Darstellung, Idempotenz | `TestKeyRedacted`, `TestKeyIdempotent`, `TestKeyGeheimnisUnsichtbar`, `TestKeinGeheimnisImErgebnis` |
| Nr. 9 | Fehler und Reihenfolge der Prüfungen | `pruefeEingabe`; `TestKeyReihenfolgeVerstoesse` (M9 rot) |
| Nr. 10 | Gleichwertigkeit als Allaussage | `TestKeyGleichwertigkeit`, `TestKeyDeterministisch`, `TestKeyEingabeUnveraendert` |
| `SPEC-117` Nr. 1 bis 3 | Matrix, nicht befragte Felder, Nullwert | `TestReplayfaehigMatrix` (27 Zeilen jetzt ausgeschrieben, `len == 27` gesichert), `…Unbekannt`, `…Unabhaengig`, `…Nullwert` (M6, M8, M14 rot) |

Eine Abweichung zwischen Spezifikation und Code gibt es nicht. `SPEC-045` und `SPEC-117` stimmen in allen Zeilen überein. Der Satz „Fragment nach dem Query-Teil bleibt unredigiert“ (Review F-12) folgt `SPEC-115` Nr. 3 und widerspricht [LH-FA-14](../../spec/lastenheft.md) nicht: ein Fragment erreicht den Upstream nie, und `LH-FA-14.a` nennt als Geltungsbereich Header, `raw_query` und die Pfadmuster des Bodys.

## 3. Zusage·Fänger-Tabelle (§3) und Raster (§6)

Jeder in der Tabelle genannte Fänger existiert unter diesem Namen (`grep ^func Test` über `src/hexagon/model/*_test.go` und `src/bootstrap/konfig_test.go`): `TestMarker`, `TestRedigiereHeaderWerte`, `…Namen`, `…Unveraendert`, `…Schreibweise`, `…NichtToken`, `…Kopie`, `TestRedigiereLocation`, `TestRedigiereQueryRandformen`, `…Geaendert`, `TestRedigiereIdempotent`, `TestRedigiereQueryUngueltigeKodierung`, `TestKeinGeheimnisImErgebnis`, `TestOhneRedigierte`, `TestParseJSONPfad`, `TestJSONPfadPasst`, `TestKonfigJSONPfadWieModell`, `TestKeyBeispieleLH`, `TestKeyEindeutig`, `TestKeyKodierungHandgebaut`, `TestKeyKopf`, `TestKeyPfad`, `TestKeyQueryReihenfolge`, `…Zerlegung`, `…Dekodierung`, `TestKeyUngueltigeKodierung`, `TestKeyRedigiertVorKodierung`, `TestKeyVary`, `TestKeyVaryRedigiert`, `TestKeyVaryMengeReihenfolge`, `TestKeyVaryAusnahme`, `TestKeyBodyBytes`, `…Kanonisch`, `…GleicheNamenReihenfolge`, `…Rueckfall`, `TestKeyRedacted`, `TestKeyIdempotent`, `TestKeyGeheimnisUnsichtbar`, `TestKeyRegelnUngueltig`, `TestKeyReihenfolgeVerstoesse`, `TestKeyGleichwertigkeit`, `TestKeyEingabeUnveraendert`, `TestKeyDeterministisch`, `TestReplayfaehigMatrix`, `…Unbekannt`, `…Unabhaengig`, `…Nullwert`. Keine Zeile nennt einen Namen ohne Test.

Raster §6: keine leere Zelle; jede Zelle nennt eine Nummer der drei Stellen oder „gilt nicht“ mit Grund. Die Zusatzzeilen (Vary × URL-Header, Menge der Vary-Namen, Body unredigiert) haben je einen Ort in `SPEC-116` Nr. 5, 8 und der Einleitung. Die Zusage „Reihenfolge der Query-Paare ohne Wirkung“ ist eine Allaussage über alle Paare; ihr Fänger fängt sie nicht über alle Fälle (V1).

## 4. Mutationen (selbst gefahren)

Jede Mutation ersetzt eine Stelle in `src/`, `make test` läuft, danach `git checkout -- src`. „rot“ nennt die Tests, die aus dem benannten Grund fehlschlugen.

| # | Zusage | Mutation | Ergebnis |
|---|---|---|---|
| M1 | Vary-Werte aus der redigierten Darstellung (F-1) | `b.vary(in.Header, regeln)` statt `header` | rot: `TestKeyVaryRedigiert` |
| M2 | Vary-Namen aufsteigend, Reihenfolge der Konfiguration ohne Wirkung (F-2) | `slices.Sort(namen)` entfernt | rot: `TestKeyVaryMengeReihenfolge` |
| M3 | gleiche Namen in Eingabereihenfolge (F-7) | `SortStableFunc` zu `SortFunc` | rot: `TestKeyBodyGleicheNamenReihenfolge` |
| M4 | Anzahl der vorhandenen Vary-Namen (F-6) | `len(namen)` (konfigurierte) statt `len(vorhanden)` | rot: `TestKeyKodierungHandgebaut`, `TestKeyBodyRueckfall`, `TestKeyBodyGleicheNamenReihenfolge` |
| M5 | Header-Namen nur nach ASCII-Schreibweise (F-4) | `asciiKlein` mit `strings.ToLower` | rot: `TestRedigiereHeaderNamen`, `TestJSONPfadPasst` |
| M6 | Capture-Status wirkt | Prüfung auf `complete` entfernt | rot: `TestReplayfaehigMatrix`, `TestReplayfaehigUnbekannt` |
| M7 | `applied` ergibt `KeyRedacted` | nur `geaendert` | rot: `TestKeyRedacted` |
| M8 | `applied` nur bei `RedigierteBodies=true` | `return true` | rot: `TestReplayfaehigMatrix` |
| M9 | Form des Pfades vor seiner Kodierung (F-3) | Formprüfung hinter die Kodierung des Pfades | rot: `TestKeyReihenfolgeVerstoesse` |
| M12 | Start nutzt den Parser des Modells | `istJSONPfad` akzeptiert zusätzlich `$.a…` | rot: `TestKonfigJSONPfadWieModell`, `TestKonfigListe` |
| M14 | `invalid` und unbekannte Integrität schließen aus | Prüfung auf `== IntegrityInvalid` | rot: `TestReplayfaehigUnbekannt` |
| M10 | Query-Paare sortiert nach (Art, Schlüssel, Wert) | Vergleich des Schlüssels abgeschaltet (`false && c != 0`) | **grün (Exit 0), V1** |

Zwei weitere Versuche (Unicode-Faltung als erste Fassung, Fragment-Mutation in `redigiereURLWert`) brachen ohne Testbefund ab (Vet-Gate, Build) und zählen nicht; das Review hat beide Zellen mit gültigen Mutationen rot gesehen. M10 ist nicht äquivalent: mit einer Wegwerf-Testdatei (nicht committet, gelöscht) ergeben `a=1&b=1` gegen `b=1&a=1` und `a&b` gegen `b&a` unter M10 verschiedene Keys (`gleich=false`, erwartet `true`). Der Code ohne Mutation ist richtig.

## 5. Review-Findings F-1 bis F-13 (Commits `aa5fe58`, `f984c1f`)

| Finding | Behandlung | Beleg |
|---|---|---|
| F-1 MEDIUM Vary × URL-Header | Architect: `SPEC-116` Nr. 5 und 8; Code: `vary(header, …)` mit den redigierten Headern | `TestKeyVaryRedigiert`; M1 rot; Zeile im Raster |
| F-2 MEDIUM Reihenfolge der Vary-Menge | Test ergänzt | `TestKeyVaryMengeReihenfolge`; M2 rot |
| F-3 MEDIUM Pfadform vor Pfadkodierung | Fall `/a%zz#` ergänzt | `TestKeyReihenfolgeVerstoesse`; M9 rot |
| F-4 MEDIUM Kelvin-Fall | Fall auf `Coo` + U+212A + `ie` gegen `Cookie` umgestellt; Kommentar sagt nur noch, was der Test fängt | M5 rot |
| F-5 MEDIUM sechs Zellen im Code entschieden | Prozessbefund; die Zellen stehen in `SPEC-115`/`SPEC-116`, Tests in `5e188dd`; keine Code-Aktion | siehe Abschnitt 6 |
| F-6 MEDIUM Anzahl der Vary-Namen | Architect: `SPEC-116` Nr. 5 nennt „Anzahl der Namen, die der Request mit mindestens einem Wert trägt“ | M4 rot |
| F-7 LOW Stabilität der Sortierung | `TestKeyBodyGleicheNamenReihenfolge` (je 40 Mitglieder) | M3 rot |
| F-8 LOW Allaussage nicht über alle Fälle | `TestRedigiereIdempotent` läuft jetzt auch über `locationFaelle`; die Plan-Zeile zu `TestKeinGeheimnisImErgebnis` nennt `queryFaelle` und `headerFaelle` und nimmt `locationFaelle` und die `TestKey*`-Tabellen ausdrücklich aus | Plan §3 |
| F-9 LOW Matrix berechnet | 27 Zeilen ausgeschrieben, `len(erwartung) == 27`, Kommentar berichtigt | M6, M8, M14 rot |
| F-10 LOW Body-Grenze | `SPEC-116` Einleitung benennt die Grenze; Kommentar an `SchluesselEingabe.Body`; akzeptiertes Negativ im Plan; Nehmer tragen die Zusage (Abschnitt 6) | kein Fänger möglich, benannt |
| F-11, F-12, F-13 INFO | keine Aktion erwartet; F-12 siehe Abschnitt 2 | — |

## 6. Übergaben (§3.13) und Plan-vs-Code (§3.9)

| Nehmer | Sendung | Befund |
|---|---|---|
| `slice-streaming-redaction` | Pfadmuster (`ParseJSONPfad`, Bedeutung) und Marker | §1 „Übernimmt (Pfadmuster und Marker)“ trägt die Sendung mit der Kennung des Gebers; sein Ausschluss „Redaction der Header und der Query“ nennt `slice-redaction-match-key`, trifft die Sendung nicht |
| `slice-record-service` | `BildeSchluessel` mit unredigiertem Request und **bereits redigiertem** Body, `ErrKodierung` (`ODR-E5006` vor der Weiterleitung), `unsafe` ergibt `null`-Key | §1 „Übernimmt (Redaction und Match-Key)“; Ausschlüsse (HTTP-Server, Sessions, Link-Umschreiben, Fallback) treffen die Sendung nicht |
| `slice-replay-service` | `BildeSchluessel` mit durch den Redaktor geleitetem Body, `ErrKodierung` ohne Cursor-Fortschritt, `Replayfaehig`, `OhneRedigierte`, `Location` mit Marker | §1 „Übernimmt …“ trägt jetzt die Body-Pflicht (Änderung in `aa5fe58`); Ausschlüsse treffen sie nicht |
| `slice-baseline-import` | `BildeSchluessel` für `@odata.nextLink`, `ErrKodierung` ergibt nicht auflösbar | §1 „Übernimmt (Match-Key)“ und der Satz in seinem Abschnitt zur Kettenauflösung (Zeile 94); Ausschlüsse treffen sie nicht |

Alle vier Nehmer liegen in `open/` (keiner in `done/`), alle nennen den Geber `slice-redaction-match-key` in §1 und im Start-Trigger. Die Reihenfolge der Welle (`welle-record-replay`) passt dazu.

Plan-vs-Code (§3.9): Kopf (`Berührte Spec-Stellen` mit `SPEC-046`), §1, §3 (Dateitabelle, Zusage·Fänger), §4 (Marke, gemessene Größe) und §6 (Raster, Risiko-Ausgänge) tragen den Stand von `f984c1f`. Die Dateitabelle nennt `redaction.go`, `schluessel.go`, `schluessel_json.go`, `replay.go` und `src/bootstrap/` (`istJSONPfad`); der Diff berührt genau diese und die Tests, dazu `harness/gegenproben.stempel` (Stempel folgt der geänderten Eingabe) und `spec/spezifikation.md`. Eine Abweichung zwischen Plan und Code fand sich nicht.

Kommentare (§3.7, §3.11): im neuen Quellcode kein Konjunktiv über verworfene Alternativen, keine Chronik, keine `//nolint`-Direktive (`grep`-Lauf ohne Treffer). Die Kopf- und Testkommentare nach `f984c1f` sagen zu, was die Tests fangen (F-4, F-9 berichtigt); Rest: V1.

## 7. Größe (§4) und Closure-Pflichten

`git diff --numstat 749a5d4..HEAD` über `src` und `spec/spezifikation.md`: Summe aus Einfügungen und Löschungen **1954** (Produktcode `redaction.go` 235, `schluessel.go` 281, `schluessel_json.go` 331, `replay.go` 21 und `konfig_werte.go` 8; Tests 1051 mit `konfig_test.go`; Spezifikation 27, davon 16 Einfügungen und 11 Löschungen). Die Marke in §4 (3500) ist nicht erreicht, die Schätzung (rund 2400) nicht überschritten; der Ausgang des Größenrisikos in §6 kann „entfallen“ lauten. Die Zahl entspricht der im Plan genannten.

Für die Closure bleibt dem Planner (nicht dem Verifier):

* DoD-Häkchen in §2 und §7 (Closure-Notiz mit Steering-Loop-Lerneintrag, Finding-Klassen des Reviews).
* Ausgänge in §6, die „offen“ stehen: die Auslegung von `SPEC-045` für `Location` mit Marker (fachlicher Owner), die Größe, der Rückfall `json-canonical` zu Art 2 und `reject` im Record (liegt bei `slice-record-service` und `slice-replay-service`, „weiter offen bis zu deren Closure“ ist ein zulässiger Ausgang).
* Beobachtungs-Register fortschreiben oder „keine Beobachtung“ in §7 notieren; V1 (Zusage ohne Fänger, vgl. `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`) und F-5 (vgl. `BEO-REPO/randform-erst-im-review-entschieden`) sind Kandidaten, die Zuordnung entscheidet der Planner.
* Der Review gilt für `5e188dd`; die Änderung danach am Produktcode ist eine Zeile (`vary(header, …)`), von M1 gefangen. Ob ein Re-Review nötig ist, entscheidet der Planner.

## 8. Gate-Lauf

`make docs-check` und `make gates` nach dem Commit dieses Berichts: Ergebnis im Handback an den Planner. Die Basis (`make test`) lief vor jeder Mutation grün (Exit 0).

## Befunde

| ID | Kategorie | Befund | Zuständig |
|---|---|---|---|
| V1 | MEDIUM | Zusage ohne Fänger: „Reihenfolge der Query-Paare ohne Wirkung“ (`SPEC-116` Nr. 4, Nr. 10; Plan-Zeile *Query-Paare: Reihenfolge ohne Wirkung* mit Fänger `TestKeyQueryReihenfolge`). Die Mutation „Vergleich des Schlüssels in der Sortierung von `zerlegePaare` abschalten“ bleibt grün, weil `TestKeyQueryReihenfolge` (`a=1&b=2`, `a=1&a=2`) und `TestKeyKodierungHandgebaut` nur Paare mit verschiedenen Werten umstellen und die Reihenfolge nach Wert mit der nach Schlüssel zusammenfällt. Unter der Mutation ergeben `a=1&b=1` und `b=1&a=1` (Art 1, gleicher Wert) sowie `a&b` und `b&a` (Art 0) verschiedene Keys; ein Replay mit umgestellten Parametern fände den Exchange nicht, genau das Beispiel aus [LH-FA-05](../../spec/lastenheft.md) (Happy Path). Der Code ist richtig; es fehlt ein Fall. Ort: `src/hexagon/model/schluessel_test.go`, `TestKeyQueryReihenfolge` | Implementer: Fälle mit gleichem Wert (Art 1) und mit Art 0 ergänzen, Mutation rot sehen; Plan-Zeile nachziehen (`AGENTS.md` §3.9) |

## Zuständigkeiten

* Implementer: V1.
* Planner: Closure-Pflichten aus Abschnitt 7, Entscheidung über ein Re-Review, Zuordnung der Finding-Klassen (V1: Zusage ohne Fänger).
* Architect: keine Spec-Lücke gefunden; keine Aktion.
