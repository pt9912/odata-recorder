# Verifikation: slice-http-adapter — 2026-10-31

**Rolle:** Verifier (Modul 11) — "Bauen wir es richtig?" gegen Plan, DoD, `SPEC-136` und `AGENTS.md`. Kein Review, keine Reparatur.

**Gegenstand:** `slice-http-adapter` (Welle `welle-record-replay`), HEAD `4ca8f2d`; Implementer-Commits `81d4bf5` bis `bea0afc`, Review `1f1bb25` (F-1 bis F-17), Fix-Runde ab `8d9d78b`. Pfade: `src/adapters/driving/httpodata/`, `.a-check.yml`, `spec/spezifikation.md` (`SPEC-136`), Plan `docs/plan/planning/done/slice-http-adapter.md`.

**Modell:** claude-sonnet-5-5. **Messverfahren:** Docker, `--memory=4g --memory-swap=4g --cpus=4 --network none`, unter `flock /tmp/odata-recorder-heavy.lock`; Exit-Codes aus Datei; Mutationen in Kopien (`.../scratchpad/verify-http/m-*`), Worktree-Code unverändert. Gegenprobe-Läufe von `a-check` am gepinnten Digest (`ac*`, `g1` bis `g5`).

## Urteil

**Bestanden, ohne Blocker.** Kein HIGH, kein falsches Codeverhalten gefunden. Der Code erfüllt `SPEC-136` Nr. 1 bis 10, 12 (Lebenszyklus Nr. 3b, 11, 13 bis 15 liegt nicht im Adapter und gehört zu `slice-http-lebenszyklus`). Die Fix-Runde deckt die Klassen von F-3 bis F-9 und F-12 in der Hauptsache; an Elementen, die weder Befund noch Implementer nannten, bleiben **Fänger-Lücken** (unten, Rest 1 bis 6), alle mit Codeverhalten, das stimmt. Sie sind keine Blocker (Code korrekt, kein HIGH), sondern MEDIUM/LOW nach `AGENTS.md` §3.15 und gehen als Negativ mit Grund und Folge an den Planner.

## Stufe (f): `make gates`

Im Worktree `make gates` gelaufen: **Exit 0** (Datei `.../verify-http/gates.exit`). Darin `docs-check` 0 Befunde, `groessenmarke slice-http-adapter.md: ok, Messung 2067 von Marke 4800`, `a-check` 0 Befunde, `test-race` grün, `migration-check: ok`.

## Stufe (e): Größenmarke

Marke 4800 fest: `git log -p` über den Plan seit `92a5f38` zeigt keine Änderung der Zeilen `marke:`, `schaetzung:`, `basis:` (Commits `b50a4ae` bis `4ca8f2d`). Formel nach `SPEC-128` Nr. 1: S = 85·5 + 80·29 + 200·2 + 380 = 3525, M = 100·⌈135·3525/10000⌉ = 4800. Messung nachgerechnet: `git diff --numstat b50a4ae..HEAD -- src/adapters/driving/httpodata/` = 2067 Einfügungen, 0 Löschungen (Produktcode 268 = 26+103+55+84; Tests 1799), gleich der Messung des Gates und dem Plan (§4, "nach der Fix-Runde"). 2067 ≤ 4800.

Beobachtung INFO für den Planner: `AGENTS.md` §3.14 nennt die Faktoren 40 je andere Zeile und × 1,25, `SPEC-128` Nr. 1 (Rang 2) 80 und 1,35. Der Plan folgt `SPEC-128`, richtig nach Source Precedence; `AGENTS.md` ist hier der nachzuziehende Text.

## Stufe (d): `.a-check.yml`

**Bestätigt: Berichtigung nach `ARC-021`, keine Lockerung nach `AGENTS.md` §3.6.** `ARC-021`: "jede Schicht darf sie nutzen"; die Importtabelle (`architecture.md`) nennt die Querschnitts-Ports in den erlaubten Zellen der Driving und Driven Adapter. Die Kanten `adapters-driving` und `adapters-driven` zu `ports-support` standen schon in `edges`; nur `direction: outbound` widersprach ihnen. Gegenproben (Kopien des Baums, `a-check` am gepinnten Digest):

| Lauf | Ergebnis |
|---|---|
| HEAD-Konfiguration | 0 Befunde |
| Konfiguration von `01a3648` (mit `direction: outbound`) | 3 Befunde `port-direction-mismatch`: `httpodata` → `ports/support` (belegt die Ursache) |
| `httpodata` importiert `ports/outbound/clock` | `port-direction-mismatch` |
| `ports/support` importiert `ports/outbound/clock` | `wrong-direction` |
| `ports/support` importiert `adapters/driven/clock` bzw. `adapters/driving/httpodata` | `port-impurity` |
| `httpodata` importiert `services/record` | `wrong-direction` |

Verbotene Kanten bleiben gemeldet; Schwelle nicht gesenkt, keine ADR nötig ([ADR-0002](../plan/adr/0002-architektur-gate-a-check.md) spiegelt die Sicht). Der Plan führt die Zeile mit Ausgang *offen*, Adresse Architect (§6) und das Negativ "kein Fall der Gegenproben nennt `ports-support`" (F-2) mit Grund und Folge. Dieses Urteil und das des Reviews sind der Beleg für die Entscheidung des Architect.

## Stufe (c): §6 offene Punkte

Im Plan §6 stehen als **eigene Zeilen mit Ausgang *offen* und Adresse Architect**: `.a-check.yml`; `DisableGeneralOptionsHandler` (F-11, F-17); `EnableFullDuplex` samt Folge für die Wiederverwendung der Verbindung (F-11, F-17); `http.Server.ErrorLog` auf `io.Discard` (F-10). Der Code entscheidet nichts darüber im Verhalten neu: `ErrorLog` bleibt unverändert, `SPEC-136` *Ungeprüft* nennt den Ist-Stand ("nicht entschieden"). Bestätigt. Rest: `ErrorLog` hat weiter keinen Fänger (Zeile benennt es selbst; `AGENTS.md` §3.12 ist durch die offene Zeile erfüllt).

## Stufe (g): DoD und Grenzen

- DoD-Haken 1 bis 4 gesetzt und wahr: Abbildung/Streaming (A3 bis A25), Fehlerobjekt je Code (A26 bis A33), Tests über `httptest`/roher Client ohne SQLite (kein Import von `adapters/driven`), `make gates` grün (oben). `-count=50` ohne Mutation: die sieben nebenläufigen Tests (`TestKontextClientAbbruch`, `TestSchreibfehler`, `TestLeihe`, `TestLeiheJederPfad`, `TestStreaming`, `TestGleichzeitig`, `TestAnfrageBodyUngelesen`) in einem Prozess, 0 von 50 rot (Exit 0, Datei `c50.exit`); `-race` läuft in `make gates` grün. Die übrigen Haken (Review, Doku, Closure, Register, Risiko-Ausgänge, Paarungen) sind offen und gehören zur Closure des Planners; kein Haken zu viel.
- Im Adapter keine Grenzen des Servers (`ReadHeaderTimeout`, `IdleTimeout`, `MaxHeaderBytes`) und kein `Anhalten`/`Schliessen`/`ErrFristAbgelaufen`: Suche in `src/adapters/driving/httpodata/*.go` ohne Treffer. Der Lebenszyklus liegt bei `slice-http-lebenszyklus`. Importe: Standardbibliothek, `model`, `ports/inbound`, `ports/support`.

## Stufe (a): Tabelle gegen den Vertrag `SPEC-136`

Zeile für Zeile gegen den Wortlaut gelesen (nicht gegen den Plan). Gedeckt: Nr. 1 (A1, auch "ruft keine Methode des Listeners": `zaehlListener`), 2 (A3, A6, A11), 3 (A10, A23), 4a bis 4i (A3 bis A9, A12), 5a bis 5d (A13 bis A17), 6a bis 6e (A18 bis A24), 7a bis 7c (A26 bis A30), 8 (A26, A31 bis A33), 9a bis 9c (A12, A25, A30), 10 (A34), 12 (A2). Der Code ist mit dem Vertrag ohne Abweichung lesbar (Reihenfolge der Prüfung von `err`, Status-Bereich 200 bis 599, Kontextfehler vor `*Fehler`, `Close` genau einmal per `defer`/vor dem Abbruch, `EnableFullDuplex` am Anfang des Handlers).

**Eine Zusage ohne Fänger (MEDIUM, kein Blocker):** `SPEC-136` Nr. 6c letzter Satz, "Der Fehler von `Close` bleibt ohne Wirkung", hat keine Zeile A1 bis A34; der Fake `koerper.Close` liefert immer `nil`. Mutation `defer func() { if a.Body.Close() != nil { brichAb() } }()` in `schreibeAntwort`: `go test` **Exit 0, nichts rot** (x6c). Folge im Betrieb: gering, weil `SPEC-121` Nr. 7 zusagt, dass `Close` des Dienstes immer `nil` liefert; ein künftiger Dienst mit Fehler von `Close` brächte die Verbindung ab, ohne dass etwas rot wird.

## Stufe (b): Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Je Befund mindestens eine Mutation an einem Element, das weder Befund noch Implementer nannten; Kopie des HEAD-Pakets, `go test -count=1`. "rot" = aus dem benannten Grund (Test nennt die Abweichung).

| Befund | Menge (Elemente) | Mutation an unbenanntem Element | Ergebnis |
|---|---|---|---|
| F-3 HEAD-Header | 33 Paare × 2 | `X-Correlation-Id` bei HEAD fehlt für Status 502 (`f3a2`); `Content-Type` bei HEAD fehlt für Code `ODR-E9001` (`f3b`) | rot: `TestFehlerHeadAlleCodes` (beide) |
| F-4 Body ungelesen/unumhüllt | Formen Body: `Content-Length` klein, groß, chunked | `a.Body` mit `struct{io.Reader}` umhüllt bei chunked (`f4a`); `io.NopCloser` bei `Laenge > 3` (`f4b`) | **grün** (beide): der Typ-Fänger `TestAnfrageBodyUngelesen` fährt nur `Content-Length: 3` |
| F-5 gleiche Werte | 4 Namen × 4 Folgen × 2 | leere Werte übersprungen, wenn der Name mehrere Werte trägt (`f5a`) | **grün**: die Folgen enthalten keinen leeren Wert |
| F-6 `Connection`-Liste | 3 × 4 × 2 = 24 je Bindung | nur die ersten 2 Zeilen von `Connection` (`f6b`); nur die ersten 3 Teile je Wert (`f6c`) | **grün** (beide): die Formen tragen höchstens 2 Zeilen und 2 Namen; `i >= 2` je Wert wird rot (leerer Teil mitgezählt) |
| F-7 Status | 400 × 2 | 451 zu 450 (`f7`) | rot: `TestAntwortStatusAlle` |
| F-8 Pfad/Query | 220 + 512 Pfad, 222 + 512 + 4 Query | `%5c` zu `%5C` im Pfad (`f8a`); `%3B` zu `;` in der Query (`f8b`); `\xC3\xA4` zu `\xC3\xA5` in der Query (`f8e`); `;` am Ende der Query gekürzt (`f8c`); `.` am Ende des Pfads gekürzt (`f8d`) | rot: `f8a`, `f8b`, `f8e`; **grün: `f8c`, `f8d`** (die Schleifen setzen jedes Byte nur *inner*halb von `a…b`; das Ende steht nur für `?`, `&`, `=`-Formen) |
| F-9 Header-Namen | 24 Namen | `X-Forwarded-For` bzw. `If-Range` aus `r.Header` gelöscht (`f9a`, `f9b`) | **grün**: die Menge der Namen ist unbeschränkt; keine Prüfgrenze mit Grund und Folge im Plan |
| F-12 Leihe je Pfad | 6 Pfade | `io.Copy(io.Discard, r.Body)` im Statusverstoß (`f12a`); `r.Header.Del` im Pfad `ErrUngueltig` vor `Close` (`f12b`); `r.Body.Close()` im Kontextfehler vor `Close` (`f12c`); `EnableFullDuplex` entfernt (`f12d`) | rot: `TestLeiheJederPfad` (a bis d), `TestLeihe` (d) |
| F-12 Lage von `EnableFullDuplex` | — | zurück vor `WriteHeader` in `schreibeAntwort` (`f12e`) | **grün**, aus Grund: `net/http` schreibt den Kopf der kleinen Fehlerantwort erst nach der Rückkehr des Handlers, also nach `Close`; das Mittel ist auf den Fehlerpfaden nicht beobachtbar. Äquivalent, kein Verhaltensfehler; der Satz "damit es auch die Fehlerpfade trägt" im Plan ist von keinem Test belegt. |

`TestAntwortHeaderGleicheWerte`, `TestAntwortStatusAlle`, `TestFehlerHeadAlleCodes`, `TestAnfragePfadKlassen`, `TestAnfrageQueryKlassen`, `TestAnfrageHeaderNamen` und `TestLeiheJederPfad` laufen als Schleifen über die Mengen (gelesen); F-3, F-7, F-12 sind damit als Klasse gedeckt, F-8 für das Innere der Bytefolge.

## Reste (Negativ mit Grund und Folge, alle Fänger-Lücken bei korrektem Code)

1. **F-4 Klasse "Body unumhüllt" (MEDIUM):** der Typ wird nur für `Content-Length: 3` geprüft. Grund: eine Hülle ändert nur den Typ, kein Bytevergleich fängt sie. Folge: ein Umhüllen bei chunked oder großem Body bleibt ungefangen (`f4a`, `f4b`). Ein Fall je Form (Länge 0/1/groß, chunked) im Fänger schließt es.
2. **F-5 leere Werte in Wertlisten (LOW):** `f5a` grün.
3. **F-6 Zahl der `Connection`-Zeilen und -Namen (LOW):** `f6b`, `f6c` grün; Formen mit drei Zeilen und drei Namen fehlen.
4. **F-8 Position am Ende von Pfad und Query (LOW):** `f8c`, `f8d` grün; die Schleife über Rohbytes setzt das Byte nur im Inneren.
5. **F-9 unbeschränkte Menge der Namen (LOW):** keine Prüfgrenze im Plan. Folge: ein gezieltes Entfernen eines Namens außerhalb der 24 bleibt ungefangen.
6. **Nr. 6c "Fehler von `Close` ohne Wirkung" (MEDIUM, Zusage ohne Fänger):** siehe Stufe (a).
7. **F-15 (aus dem Review, nicht Gegenstand der Fix-Runde):** Prüfgrenzen mit Grund und Folge stehen im Plan für A9 und A15; für A18 (Antwort bis 1 MiB + 1) und `Content-Length` bei 204/304 (A20) fehlen sie.

Adresse für 1 bis 7: Planner (Plan §6 mit Ausgang *offen* oder *akzeptiertes Negativ* vor der Closure, `AGENTS.md` §3.13) oder Implementer, falls der Planner den Fänger nachfordert. Kein Blocker: kein HIGH, kein falsches Codeverhalten.

## Plan-vs-Code-Diff (Zusammenfassung)

Plan und Code stimmen in Zeilenzahl, Dateien und Mutationen überein (`AGENTS.md` §3.9): Tabelle A1 bis A34, Mengen, Raten (50 von 50 und 20 von 20 laut Review bestätigt; hier nur der Grundlauf 0 von 50 gemessen), §6, Marke 4800 fest, Messung 2067. Keine Abweichung zwischen Plan und Code gefunden. Die Mutationen dieser Verifikation erscheinen nicht im Plan; sie sind die Reste oben.

## Rohbytes (`AGENTS.md` §3.18)

Dieser Bericht enthält kein `\u` mit Hex-Ziffern; Umlaute stehen als UTF-8. Gegengelesen mit `LC_ALL=C grep -P '\\\\u[0-9A-Fa-f]{4}'` auf dieser Datei (kein Treffer).
