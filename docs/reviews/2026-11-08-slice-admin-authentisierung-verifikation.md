# Verifikation: slice-admin-authentisierung — 2026-11-08

**Art:** Verifikation (einmal) — gegen DoD, Tabelle Zusage · Fänger, `SPEC-238` Nr. 1 bis 5 und 13, `SPEC-240`, `LH-FA-13.a`, `LH-QA-04` und die Hard Rules; nicht Diff-Review (siehe `2026-11-08-slice-admin-authentisierung-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-admin-authentisierung`, HEAD `f468858c` (Implementer `c4c06ed1`, Review `e115c53d`, Architect-Entscheid F-3 `3f754815`, Fix-Runde F-1 `4c3cc788`, Plan `f468858c`), Basis `c469ef1e`. Plan: `docs/plan/planning/done/slice-admin-authentisierung.md`. Bezug: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden mit Auflagen an den Planner; kein HIGH, keine DoD-Verletzung im Code.** Der Produktcode (`admin.go`) setzt `SPEC-238` Nr. 1 (`New` kopiert, `Vergleich` nil), Nr. 3 (a, c, d, e, f), Nr. 4 (a bis e) und Nr. 5 vollständig um; die Fix-Runde F-1 deckt die Klasse „Byte am Rand des Kandidaten“ gegen vier eigene Mutationen an nicht genannten Elementen (alle rot). Die Nachbarklasse „Zahl der führenden Leerzeichen“ hat keine Prüfgrenze im Plan, eine Mutation bleibt grün (V-1). Der Plan zieht die Fix-Runde in drei Stellen von §3 nicht nach (V-2). Beides sind Auflagen an den Planner, kein Blocker. `make gates` ist grün.

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| Nullprobe des Pakets | `go test -count=1 -skip TestWettlaufNebenlaeufig ./src/adapters/driving/admin/` in einem eigenen Container (`odata-recorder:test`, netzlos, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`; Baum des Zweigs per `git archive` hineinkopiert, nichts vom Repo eingehängt): `ok`. Der Flake `TestWettlaufNebenlaeufig` (Teil 1) ist ausgenommen, wie vorgegeben. |
| Zweige | eigenes `go test -coverprofile`: Paket 98,6 %, genau drei Blöcke mit Zähler 0 (`route.go:128`, `tokendatei.go:108`, `tokendatei.go:132`), keiner in `admin.go`; das deckt sich mit dem Plan (§3, „Mengen“). Der Lauf ist ohne `-race` (das Image `test` hat keinen C-Compiler); den Lauf unter dem Detektor fährt `make gates` (`test-race`). |
| Größenmarke | `git diff --numstat c469ef1e..HEAD -- src/adapters/driving/admin/`: Summe 1537, Marke 2600 (0,59), die Schnittgrenze 2223 ist nicht berührt; `make groessenmarke` im Gate-Lauf: `ok, Messung 1537 von Marke 2600`. |
| Plan-vs-Code-Diff | `git diff --stat main...HEAD`: 12 Dateien. Produktcode allein `admin.go` (109 Zeilen), sonst Tests, `export_test.go`, `spec/spezifikation.md` (4 Zeilen), zwei Pläne und der Review-Report. Kein Pfad der Gate-Kette; `src/bootstrap` und `src/hexagon` unberührt (das Binary bleibt unverändert, wie §2 zusagt). |
| Spec-Änderung | `spec/spezifikation.md`: nur `SPEC-238` Nr. 4 e (Entscheid F-3) und der Absatz *Ungeprüft* zum Hashen; der Kommentar von `vergleiche` folgt dem Wortlaut („Der Vergleich der Summen hängt weder vom Inhalt noch von der Länge ab …“). Kein Widerspruch zu `LH-FA-13.a` gefunden. |
| Merge-Nähe | `git merge-tree --write-tree main HEAD`: Rückgabe 0, Baum `c48552fe`, kein Konflikt. Zählstellen der Fehlercodes: `src/hexagon/model/fehler.go` trägt auf `main` und auf dem Zweig je 40 Konstanten `Code…`; der Zweig ändert die Datei nicht und führt keinen neuen Code (`ODR-E2101` steht in `harness/katalog-ausnahmen.txt`), also ist kein Nachzug nötig. |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -nP '\\u[0-9A-Fa-f]{4}\|\xEF\xBB\xBF\|\xC2\xA0' src/adapters/driving/admin/*.go`: 0 Treffer. Das Token mit BOM entsteht in `bom()` aus Bytewerten, die 256 Bytes aus einer Schleife über `byte(i)`. |
| Betriebsregeln | Docker nur mit `--memory=4g --memory-swap=4g --cpus=4`, ein Container, jeder Lauf unter `flock`; Mutationen in einer Wegwerfkopie im Scratchpad, das Repo blieb unverändert (`git status`: allein dieser Report). |
| `make gates` | einmal am Ende, unter `flock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/gates.log`: `EXIT=0`; darin `d-check: 880 Datei(en) geprüft, 0 Befund(e)` (mit diesem Report im Baum), `groessenmarke … ok`, `zeichen: ok`, die Stufe `race` gebaut und grün. Der Flake `TestWettlaufNebenlaeufig` trat in diesem Lauf nicht auf. |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Ablauf (A1, A3 bis A6; 26 Requestarten) | im Code erfüllt, Haken offen | `ServeHTTP` gelesen: `KorrelationsID` genau einmal als erste Operation, `GET` plus `URL.Path == "/healthz"` frei, Authentisierung vor `ordne`, `rufe` ruft den Handler einmal. Eigene Mutationen M9, M10, M12 rot. |
| Authentisierung (K4, K5, U1 bis U11) | im Code erfüllt, Haken offen | `authentisiert`: genau eine Zeile, Schema in jeder ASCII-Schreibweise, Rest nichtleer, `Vergleich` genau einmal, allein sein Ergebnis gilt; `vergleiche` = `ConstantTimeCompare` auf `Sum256`; 401 mit `WWW-Authenticate: Bearer` und festem Fehlerobjekt. Eigene Mutationen M8, M11 rot; M1 grün (V-1). |
| Sendung `slice-admin-rahmen` (A7) | erfüllt, Haken offen | `anfrage` filtert über `strings.EqualFold` und kopiert die übrigen Header (`slices.Clone`); (a) 65536 + 384 Fälle direkt, (b) 8192 + 96 über `ServeHTTP`. Eigene Mutation M7 (Header ohne Kopie) rot. |
| Antwort (E2 bis E4) | im Code erfüllt, Haken offen | `schreibeErfolg` und `schreibeFehler` gelesen: `Allow` nur bei 405, `WWW-Authenticate` nur bei 401, `X-Correlation-Id` vor jeder Verzweigung gesetzt; `Content-Type` und `Content-Length` nur bei nichtleerem Body. |
| `make gates` grün | erfüllt | Abschnitt 1. |
| Review, Doku-Update, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen (Closure) | Der Review-Report liegt vor; §7 steht auf „Wird bei der Closure gefüllt“. §6 trägt für jedes Risiko einen Ausgang; „weiter offen“ nur bei Register-Zeilen und der Lücke des Lastenhefts (Adresse User). Kein Fehlercode, kein Handbuch-Nachzug nötig. |

Eine Lücke gegen `LH-FA-13.a` ist offen benannt und kein Verstoß dieses Slice: das Abweisen schreibt noch keinen Audit-Eintrag und zählt nicht (`slice-admin-abweisen-audit`, `slice-admin-fehlerabbildung`, §6); der Listener ist nirgends verdrahtet. `LH-QA-04`: weder Token noch Kandidat erreichen Body oder Header (U11, 22 Fälle, Teilstücke ab 8 Bytes); `Anfrage.Header` trägt kein `Authorization`.

## 3. Tabelle Zusage · Fänger gegen den Vertrag (22 Zeilen)

Jede Zeile gegen den Wortlaut von `SPEC-238` Nr. 1 bis 5 und 13 gelesen. Alle 22 Zeilen decken die Zusagen; keine Zusage des Vertrags (Nr. 3 a, c, d, e, f; Nr. 4 a bis e; Nr. 5 Header und Erfolgsantwort; Nr. 13 a, b) ist ohne Zeile, keine Zeile ohne Fänger im Test. Eigene Mutationen, die der Implementer nicht nannte (ein Container, ein Lauf je Mutation, Paket ohne den Flake):

| Zeile | Eigene Mutation | Ergebnis |
|---|---|---|
| K4 | Kopie des Tokens auf 64 Bytes gekappt (`M8`) | rot: `TestKandidatAlleBytes`, `TestTokenLaengen` |
| A1 | `KorrelationsID` ein zweites Mal im Zweig 404 (`M12`) | rot: `TestKorrelationsHeader`, `TestKorrelationsID`, `TestKorrelationsIDJeRequestart` |
| A3, A4 | `/healthz/` frei (`TrimSuffix`, `M9`); `HEAD /healthz` frei (`M10`) | rot: `TestHealthzMethodeUndPfad` (M10 zusätzlich `TestAuthentisierungVorZuordnung`, `TestMethodeUndPfad`) |
| U3, U4 | Kandidat der Länge 1 als leer behandelt (`len(rest) < 2`, `M11`) | rot: `TestVergleichEntscheidetAllein` |
| A7 | `Header` ohne Kopie, Slices geteilt (`M7`) | rot: `TestHeaderKopie` |
| U3 (Rand) | Abschnitt 4 | rot |
| U2 | Zahl der führenden Leerzeichen auf 5 begrenzt (`M1`) | **grün** (V-1) |
| A5, A6, U5 bis U11, E2, E3, E4 | gelesen; die Fänger sind Schleifen über die Mengen mit Literal-Orakel und Zählprüfung; die Fakes zählen jede Operation (`zaehlendeIDs` fünf, `vergleichFake` hält beide Argumente, `routeFake` `Anfrage` und `ctx`, `useCaseFake` beide Operationen, Audit und Telemetrie jede) | ohne Befund; keine eigene Mutation (Mutationen des Implementers und des Reviewers nicht als Beleg übernommen) |

Raster (§6) und Zusatztabelle: jede Zelle trägt eine Nummer von `SPEC-238` oder „gilt nicht“ mit Grund; die Zusatztabelle nennt zu jedem Paar einen Fänger aus §3. Keine im Code entschiedene Zelle gefunden (die Lesart `URL.Path` für `/healthz` ist N-1). Die akzeptierten Negative N-1 bis N-4 tragen Grund und Folge.

## 4. Fix-Runde F-1 gegen die Klasse (`AGENTS.md` §3.17)

Menge aus dem Wortlaut von `SPEC-238` Nr. 4 b („der Rest hinter den führenden Leerzeichen ist der Kandidat, Byte für Byte unverändert“): Rand des Rests = jedes Byte am Anfang (hinter einem und hinter drei Leerzeichen; das Leerzeichen selbst gehört zu den führenden und entfällt, 255 × 2), am Ende (256) und an beiden Rändern (255) = 1021 Elemente. `TestKandidatRaender` zählt sie (`255*2+256+255`), baut die Werte aus `byte(b)` (keine Escape-Folge), ruft `ServeHTTP` direkt und prüft, dass `Vergleich` genau einmal den Rest unverändert erhält. Die Menge stimmt mit dem Wortlaut überein und ist über die Bytewerte vollständig.

Eigene Mutationen an Elementen, die weder der Befund (0xA0, 0x01, `\r\n\v\f`) noch die Fix-Runde (`\xff`, NUL, Tab, Leerzeichen) nannten:

| Mutation | Element | Ergebnis |
|---|---|---|
| `TrimLeft(…, " \x7f")` (`M2`) | Anfang 0x7f | rot: `TestKandidatRaender` |
| `TrimLeft(…, " \xfe")` (`M3`) | Anfang 0xfe | rot: `TestKandidatRaender` |
| `TrimRight(…, "\x7f")` (`M4`) | Ende 0x7f | rot: `TestKandidatRaender` |
| `TrimSuffix(…, "\xfe")` (`M5`) | Ende 0xfe | rot: `TestKandidatRaender` |
| letztes Byte gestrichen, wenn gleich dem ersten (`M6`) | beide Ränder gleich | rot, aber in vielen Tests (Fixture-Token beginnt und endet mit `t`): kein reiner Beleg für den Fänger |

Damit ist die Klasse „Byte am Rand des Kandidaten“ für Bytewerte geschlossen. Nicht geschlossen ist die Nachbarklasse „Zahl der führenden Leerzeichen“ (V-1).

## 5. Sendung A7 und Abgabe aus `slice-admin-rahmen`

- **A7 (Sendung, Befund 1 des Rahmens):** In §1, in der DoD (dritter Punkt) und als Zeile A7 geführt, der Geber ist genannt. Die Menge (8192 Schreibweisen von „authorization“, 2 hoch 13, mal 8 Wertformen = 65536; 48 Nachbarn × 8 = 384) folgt dem Wortlaut „in keiner Schreibweise … bei keiner Zahl und Form der Werte“; die Zählprüfung steht im Test. Die Erreichbarkeit ist zutreffend begründet: ein Request mit zwei Zeilen endet im 401 (U1), also beobachtet allein der direkte Aufruf von `anfrage` über `export_test.go` den Filter; (b) fährt dieselben Schlüssel mit einer gültigen kanonischen Zeile über `ServeHTTP`. Mutation M7 und die Mutationen des Implementers (`A7-nur-einzeilig` nur in (a) rot) tragen das.
- **N-8 bis N-11 des Rahmens:** Der Rahmen gibt allein Befund 1 (mehrzeiliges `Authorization`) an diesen Slice ab; N-8 und N-9 (Zuordnung, `route.go`) und N-11 (Länge und Zeichen eines Parameterwertes, Name und Wert eines Headers) sind akzeptierte Negative ohne Sendung, N-9 sichtet der Planner beim Start der Endpunkt-Slices. Der Plan führt das zutreffend: die DoD nennt nur die Sendung, N-4 übernimmt den Rest von N-10 (Namen ab 20 Zeichen und jenseits der 48 Nachbarn bleiben ohne Fänger) und sagt zutreffend, dass die Wertformen (leerer erster Wert, leere Liste) jetzt A7 fängt. Für den Wert eines Headers (Länge, Zeichen) bleibt keine obere Grenze im Test, und N-4 nennt es nicht (V-3, INFO).

## 6. Befunde und Auflagen

| ID | Kategorie | Befund | Auflage |
|---|---|---|---|
| V-1 | LOW | Die Zahl der führenden Leerzeichen hinter dem Schema ist unbeschränkt; U2 führt 1, 2 und 5, `TestKandidatRaender` 1 und 3. Die Mutation „höchstens fünf führende Leerzeichen entfernen“ (`for n := 0; n < 5 && r[0] == ' '`) bleibt grün (`M1`, Paket `ok`); ein Client mit mehr als fünf Leerzeichen hinter dem Schema würde bei einer solchen Änderung abgewiesen, ohne dass ein Test rot wird. Quelle: `SPEC-238` Nr. 4 b („mindestens einem Leerzeichen“), `AGENTS.md` §3.15, Beobachtung `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze`. | Planner: entweder die Prüfgrenze benennen und ergänzen (z. B. U2 um 6, 64 und 1024 Leerzeichen) oder in §3 als akzeptiertes Negativ N-5 mit Grund und Folge führen (Menge unbeschränkt, Kopfgröße begrenzt der `http.Server`; Folge: eine Begrenzung der Leerzeichen bliebe grün). Kein Code-Fehler, kein Blocker. |
| V-2 | LOW | Der Plan zieht die Fix-Runde in drei Stellen von §3 nicht nach (`AGENTS.md` §3.9): die Zeile U3 der Tabelle nennt `TestKandidatRaender` weder als Fänger der Klasse „Rand des Kandidaten“ noch gibt sie die Mutationen an; die Mutationstabelle („Zeile · Mutation · roter Test“) hat zu U3 keine der neun Mutationen der Fix-Runde; der Absatz „Mengen“ führt die 1021 Fälle nicht auf (sie stehen allein im Absatz „Fix-Runde nach dem Review“). Der Absatz „Ergebnis“ nennt weiter „109 Mutationen, 108 rot“, der Fix-Absatz davor „118, 117“: stimmig (109 + 9), aber die Absätze stehen in umgekehrter Reihenfolge. | Planner (Closure): U3-Zeile, Mutationstabelle und Mengen-Absatz um `TestKandidatRaender`, die neun Mutationen und die Zahl 1021 ergänzen; keine Codeänderung. |
| V-3 | INFO | N-4 nimmt aus N-10 des Rahmens die Namen ab 20 Zeichen auf, nicht Länge und Zeichen eines Header-Wertes aus N-11. Die unbeschränkte Menge hat keine benannte Prüfgrenze im Plan. | Planner: in N-4 mitführen oder auf N-11 verweisen. |
| V-4 | INFO | Zwei Mutationen des Reviews bleiben als akzeptiert bestätigt: `return a == b` neben dem Aufruf (N-3, F-2 des Reviews) und die äquivalente Mutation `E2-www-authenticate-bearer-an-jedem-fehler`. Die Parameter `allow, www` von `schreibeFehler` (F-4 des Reviews) sind gleichen Typs; E2 und U9 fangen eine Vertauschung. | keine |

Die Befunde des Reviews F-1 (MEDIUM) und F-3 (INFO) sind aufgelöst: F-1 durch `TestKandidatRaender` und die Klasse in Abschnitt 4, F-3 durch den Entscheid des Architect (`SPEC-238` Nr. 4 e, Kommentar von `vergleiche`) und den Nachtrag in §6 des Plans.

## 7. Summary

HIGH 0 · MEDIUM 0 · LOW 2 (V-1, V-2) · INFO 2. DoD im Code erfüllt; die Haken der DoD, §7 und die Closure-Pflichten gehören dem Closure-Schritt. `make gates` grün, Merge gegen `main` ohne Konflikt, Zahl der Fehlercodes 40 unverändert (kein Nachzug).
