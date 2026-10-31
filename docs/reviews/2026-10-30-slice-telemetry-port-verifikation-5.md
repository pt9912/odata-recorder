# Verifikations-Report 5: slice-telemetry-port (Teil P der Kette P, L, S), gezielte Nachverifikation des Deltas nach Fix-Runde 6 - 2026-10-31

**Rolle:** Verifier (Modul 11), fünfte Runde, frischer Kontext, gezielt auf das Delta `5e422f4..9b3dac7`. Nicht Gegenstand: V-14 (gelöst), V-22 (INFO), die Closure, das Register.

**Gegenstand:** `slice-telemetry-port`, Zweig `slice-telemetry-port`, HEAD `9b3dac7` (Fix-Runde 6). Plan: [slice-telemetry-port](../plan/planning/in-progress/slice-telemetry-port.md); Vorbericht: [Verifikation 4](2026-10-30-slice-telemetry-port-verifikation-4.md).

**Gegen was geprüft:** `AGENTS.md` §3.10, §3.11, §3.14, §3.15, §3.17; `SPEC-230` Nr. 9 und 10, `SPEC-232` Nr. 4 in `spec/spezifikation.md`; der Vertrag [SPEC-128](../../harness/vertraege/SPEC-128-groessenmarke.md) (Nr. 3 n, Nr. 5 d, e); der Diff `git diff 5e422f4..9b3dac7` (4 Dateien: 3 Testdateien, der Plan).

**Modell:** claude-sonnet-5-5. **Arbeitsweise:** Gate und Stufen selbst gefahren; Mutationen und Probe-Tests nur auf Kopien im Scratchpad (`/tmp/claude-1000/-Development-odata-recorder/c59f1e78-ff6f-4a6a-8989-2406d85f058d/scratchpad/verif-telemetrie-5/`), Image `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, jede Gruppe unter `flock /tmp/odata-recorder-heavy.lock`. Im Repo geändert ist nur dieser Report. Kein Container und keine wartende Sperre dieser Runde blieb zurück; die drei Stufen-Images `v5-test`, `v5-race`, `v5-lint` sind entfernt.

---

## Verdikt

**BESTANDEN** - mit 0 HIGH, 0 MEDIUM, 2 LOW, 2 INFO. V-23, V-24 und V-25 sind gegen ihre Klassen geschlossen: 30 Mutationsläufe an Elementen, die weder Report 4 noch der Implementer nannten (§2), davon 17 rot; die 13 grünen sind keine Lücke der drei Befunde, sondern gehören zu vier benachbarten Klassen (§3: V-28 bis V-31), jede grüne mit Probe-Test als nicht äquivalent belegt (Probe auf dem Stand grün, unter der Mutation rot), außer wo der Plan sie als N-2 führt. Der Zeit-Fänger ist bei Last stabil (0 von 400). Die Messung 3093 ist nachgerechnet, der Eintrag `ueberschreitung: 3100` formgerecht, die Marke 3000 unberührt. Produktcode unverändert.

Warum nicht NICHT BESTANDEN: V-28 und V-29 haben die Form von V-23 (Zustand oder Art des Arguments, die ein Fänger nicht führt), aber keine geschriebene Zusage nennt sie als Menge: V-23 widersprach dem Wortlaut eines Kommentars ("Abbruch, Frist und Werte bleiben") mit drei Elementen; V-28 und V-29 verlangen eine Auslegung von "dieselbe Identität" und "jede Operation" über Zustand und Span-Art. Jede dazu nötige Mutation ist künstlich bedingt (`if ctx.Err() != nil`, `if s.Eltern != nil`); ein Dienst-Test, der Kind-Spans beendet, würde den Fehler sofort sehen. Der Planner kann sie als MEDIUM lesen (`AGENTS.md` §3.15, "Zusage ohne Fänger"); dann ist das Verdikt NICHT BESTANDEN, und der Fix ist klein (siehe §3).

## 1. Sensor-Belege (selbst gefahren)

| Sensor | Ergebnis | Beleg |
|---|---|---|
| `make gates` im Worktree | **Exit 0** (Exit-Code in `gates.exit` geschrieben und aus der Datei gelesen, nicht über `tail`); Load Average danach 9,68 | `gates.exit`, `gates.log` im Scratchpad: a-check `gesamt: 0 Befund(e)`, d-check `372 Datei(en) geprüft, 0 Befund(e)` |
| `make groessenmarke` (im Gate) | `groessenmarke slice-telemetry-port.md: ok, Messung 3093 von Marke 3000`; `Messpunkt ... zeilen=17 nebenlaeufig=2 fakes=1 produktcode=580 schaetzung=2150 marke=3000 messung=3093` | Gate-Ausgabe |
| Cache im Gate | Die Schichten von `test`, `race`, `lint` standen als `CACHED`; darum einzeln ohne Cache (`--no-cache-filter`, Argumente wie im Makefile, Tags `v5-*` statt der Gate-Tags) | `stage-*.log` |
| Stufe `test` ohne Cache | Exit 0, `pruefung test: ok` (86,1 s) | `stage-test.log` |
| Stufe `race` ohne Cache | Exit 0, `pruefung race: ok` (136,7 s) | `stage-race.log` |
| Stufe `lint` ohne Cache | Exit 0, `0 issues.` (erster Lauf scheiterte an meinem fehlenden `LINT_IMAGE`-Argument, nicht am Code; Wiederholung mit dem Argument des Makefiles) | `stage-lint.log` |
| Last | Load Average 17 bis 29 während der Läufe (viele wartende Linien an der Sperre) | `uptime` |
| Produktcode unverändert | `git diff --name-only 5e422f4..9b3dac7` nennt `src/adapters/driven/telemetry/telemetry_test.go`, `src/hexagon/ports/support/supporttest/telemetrie_test.go`, `src/hexagon/ports/support/telemetry_test.go` und den Plan; keine Datei unter `src/` ohne `_test.go`, nichts unter `spec/`, `harness/`, `go.mod` | `git diff` |
| `make docs-check` auf diesem Report | Exit 0, `d-check: 373 Datei(en) geprüft, 0 Befund(e)` (§10) | `docs.log` |

Nicht gefahren, wie beauftragt: `make gegenproben`, Stempel.

## 2. Klasse gegen Klasse: eigene Mutationen

Mengen aus dem Wortlaut der Kommentare von `FakeTelemetry` und aus `SPEC-230` Nr. 9 und 10 gebildet, nicht aus dem Befund. Läufe: `go test -count=1` über `ports/support`, `supporttest`, `adapters/driven/telemetry`, ohne Detektor (die Fänger sind deterministisch), je Mutation ein Lauf.

**Rot (17 Läufe):**

| Id | Klasse | Mutation | roter Test |
|---|---|---|---|
| H5 | V-23, Zustand abgelaufen | abgelaufener Eingabe-ctx wird `WithoutCancel` | `TestFakeCtx` |
| H6 | V-23, Frist mit Span im ctx | Eingabe-ctx mit Frist und Span wird `WithoutCancel` | `TestFakeCtx` |
| X1 | V-23, Zustand später abgebrochen | ctx mit Frist und ohne Abbruch wird neu aus `WithoutCancel` und Frist gebaut (späterer Abbruch geht verloren) | `TestFakeCtx` |
| X2 | V-23, Ursache | abgebrochener ctx mit Ursache wird neu gebaut, Ursache bleibt, Frist geht verloren | `TestFakeCtx` |
| X3 | V-23, abgelaufen plus Werte | abgelaufener ctx wird auf `Background` mit Span und Frist neu gebaut (Werte gehen verloren) | `TestFakeCtx` |
| H9 | V-23, nil-ctx | `StartSpan` mit nil-ctx gibt nil als ctx zurück | `TestFakeArgumenteUngeprueft`, `TestFakeEltern` |
| H11 | V-24, Richtung hinein | `SetzeStand` kopiert nicht | `TestFakeSnapshotStand`, `TestFakeStandKopie` |
| H12, H13 | V-24, nil und leer wie gesetzt | `Snapshot` macht `Dropped` bzw. `Spans` nicht-nil | `TestFakeSnapshotStand` |
| H14 | Zählen | `Snapshot` mit gesetztem Stand wird nicht gezählt | `TestFakeSnapshotGleichzeitig` |
| H16, H16b | Länge der Liste | `Log` bzw. `Set` kürzen auf 12 Attribute | `TestFakeLaengen` |
| H17, H18 | Argumente ungeprüft | `CountError` maskiert den Code auf 24 Bit; `Log` deckelt die Stufe bei 100000 | `TestFakeArgumenteUngeprueft` |
| H21 | Aufzeichnung vollständig | `Aufrufe()` liefert nur die letzten 5000 | `TestFakeGleichzeitig`, `TestFakeLesenGleichzeitig` |
| K1 | V-25, Position und Länge | `Counts` der Länge 3 bei ungeradem `Count` geteilt | `TestFakeStandKopie`, `TestSnapshotKopieViele` |
| K4 | V-25, Zahl der Einträge | ab 1000 Verteilungen wird `Counts` geteilt | `TestSnapshotKopieViele` |

Gezählt nach Läufen: 17 rot (H5, H6, X1, X2, X3, H9, H11, H12, H13, H14, H16, H16b, H17, H18, H21, K1, K4), 13 grün (unten), zusammen **30 Läufe**; H16 und H16b sind dieselbe Mutation an zwei Aufrufstellen (`Log`, `Set`), H12 und H13 an zwei Sammlungen.

**Grün (13)** - Einordnung in §3:

| Id | Mutation | Klasse | Probe (nicht äquivalent) |
|---|---|---|---|
| H1 | `Log` zeichnet bei abgebrochenem ctx `WithoutCancel(ctx)` auf | V-28 | `TestProbeCtxZustand` rot (P1) |
| H2 | `StartSpan` dasselbe | V-28 | rot (P2) |
| H3 | `StartSpan`: Eltern fehlt bei abgebrochenem ctx | V-28 | rot (P3) |
| H15 | `End` eines Spans mit Eltern wird nicht aufgezeichnet und nicht gezählt | V-29 | `TestProbeKind` rot (Q1) |
| H15b | `Set` eines Spans mit Eltern dasselbe | V-29 | rot (Q2) |
| H24 | `End` eines Kind-Spans ab ID 100 wird nicht gezählt | V-29 | rot (Q5) |
| H20 | Span-ID läuft ab 1000 wieder bei 1 an | V-30 | rot (Q3) |
| H22 | `Log` kürzt `Text` ab Position 1 auf 100 Bytes | V-30 | rot (Q4) |
| K5 | `Kopie` teilt `Errors` ab 40 Einträgen | V-30 | nicht geprüft (siehe unten) |
| H10 | Eltern aus dem ctx eines anderen Fakes wird nicht festgehalten | V-30 | nicht geprüft (siehe unten) |
| A4 | `CountError` maskiert den Code auf 16 Bit (65537 zählt als Code 1) | V-30 (N-2 A02c) | nicht geprüft (siehe unten) |
| K2, K3 | `Kopie` kopiert `Attrs` nicht bei Position 3000 bzw. ab 1024 | N-2 (Position über 1024) | N-2 des Plans, wie dort |

Die Proben P0 und `base` laufen auf dem unveränderten Stand grün (`ok`); P1 bis P3 und Q1 bis Q5 rot unter der jeweiligen Mutation. K5, H10 und A4 habe ich nicht mit einer Probe gegengeprüft: eine Eingabe, die sie unterscheidet, gibt es nach Bauart (ein `Errors` mit 41 Einträgen; ein Eltern-Span eines anderen Fakes im ctx; der Code 65537), die Fänger führen sie nicht.

**Fix-Runde 6 gegen ihre Klassen:**
- **V-23** (Eigenschaft x Zustand x Ausstattung x Aufrufstelle, 80 Fälle am Test gerechnet: 5 x 2 x 2 x 4): geschlossen. Ich habe H5, H6, X1, X2, X3 und H9 an Elementen gefahren, die weder Report 4 noch der Implementer nannten (F01 bis F03 und F05, F06 waren die genannten); alle rot, alle in `TestFakeCtx`. Ob die Menge der Zustände vollständig ist: "abgebrochen und danach abgelaufen" (beide Ursachen zugleich) führt kein Zustand; ein Element, an dem sich kein Fänger-Unterschied konstruieren ließ (Err bleibt, was zuerst eintrat), darum kein Befund.
- **V-24** (3 beschreibbare Maps x 2 Leser): geschlossen; 6 Fälle am Test gezählt. Die Aussage "ein Slice der Länge 0 lässt sich nicht beschreiben, Spans ist äquivalent" gilt für den leeren Slice ohne Kapazität; ein geteilter leerer Slice mit Kapazität wäre über ein erneutes Aufschneiden sichtbar - die Bauart des Fakes (Literal `[]support.SpanRecord{}`) hat sie nicht, kein Befund. Eigene Mutation am leeren Stand habe ich nicht gefahren (der Implementer nannte F08, F08b bis F08d, F08f; H12 und H13 gehören zu "nil oder leer wie gesetzt").
- **V-25** (Position, Länge, Art, Code): geschlossen für die genannten Dimensionen. Eigene rote Mutationen an Elementen außerhalb der Aufzählung: K1, K4 (Position x Länge, Zahl der Einträge). Zwei Mengen der Fix-Runde sind kleiner als die Dimension: V-30.

## 3. Befunde

Zuständigkeit: Implementer = Test im Slice; Planner = Plan und Adresse.

| Nr. | Schwere | Zuständig | Befund |
|---|---|---|---|
| **V-28** | LOW | Implementer | **Zustand des Eingabe-ctx gilt für den Rückgabe-ctx, nicht für den aufgezeichneten ctx und die Eltern (`AGENTS.md` §3.15, §3.17).** `SPEC-230` Nr. 10 sagt, der Fake halte "den `ctx` (dieselbe Identität)" fest und `StartSpan` den Eltern-Span aus dem `ctx`. `TestFakeArgumenteUngeprueft` und `TestFakeEltern` führen den ctx aktiv; ein abgebrochener oder abgelaufener Eingabe-ctx steht in keinem Fänger der Aufzeichnung. Grün: H1, H2 (aufgezeichnet wird `WithoutCancel(ctx)`), H3 (Eltern fehlt bei abgebrochenem ctx); Proben rot. Fix: in `TestFakeArgumenteUngeprueft` und `TestFakeEltern` die drei Zustände (aktiv, schon abgebrochen, schon abgelaufen) als Schleife über `Log` und `StartSpan` (Identität) und `StartSpan` (Eltern). |
| **V-29** | LOW | Implementer | **"Zählt jede Operation, auch `Set`, `End`" ist nur für Wurzel-Spans gefangen (`AGENTS.md` §3.15).** `TestFakeSpanJeOperation` und die übrigen Fänger rufen `Set` und `End` nur an Spans ohne Eltern. Grün: H15, H15b (Kind-Span wird nicht aufgezeichnet), H24 (Kind-Span ab ID 100 nicht gezählt); Proben rot. Fix: ein Kind-Span (Eltern im ctx) in `TestFakeSpanJeOperation` und in einem Zähler über mehr als 100 Spans. |
| V-30 | INFO | Planner / Implementer | **N-2 und die Fänger lassen weitere Dimensionen offen, ohne sie zu nennen (`AGENTS.md` §3.17).** Grün: H20 (Span-ID läuft ab 1000 wieder an; `TestFakeSpanID` zählt wenige Spans), H22 (Produkt aus Länge der Zeichenkette und Position in der Liste: `TestFakeStrings` führt die Position 0, `TestFakeLaengen` kurze Strings), K5 (Zahl der Einträge in `Errors`, `Dropped`; die Fänger gehen nur für `Durations` und `Spans` über 1024), A4 (Maskierung des Codes, 65537 zählt als Code 1; N-2 nennt "Verzweigung" zwischen 3051 und 99997, nicht Aliasing), H10 (Eltern aus einem anderen Fake; die Spezifikation schweigt). Folge im Betrieb: keine, weil der Fake keine dieser Verzweigungen hat (gelesen), der Adapter `CountError` per Map-Zugriff zählt, und die Tabelle 33 Codes mit den Werten 1 bis 33 hat. Entweder je Dimension einen Fall oder N-2 um diese Zeilen ergänzen. |
| V-31 | INFO | Planner | **Die Herleitung in N-2 ist an zwei Stellen ungenau.** (a) "Zeichenketten 70000 gegen die Grenzen 40, 128 und 4096 der Formen": die Formen haben die Grenzen 16, 32, 64 und 128 (und 71 beim Blob, `regeln.go`); 40 und 4096 sind Mutationspunkte aus Report 4 und dem Fix, keine Grenzen einer Form; `spec/` kennt 4096 nicht. (b) "die Tabelle 33 Codes im Bereich bis 9001": 9001 ist der Text von `ODR-E9001`; der Wert des Typs `model.Code` läuft von 1 bis 33 (`iota + 1`), das Prüfen von -50 bis 3050 liegt also um den Wertebereich, nicht um 9001. Beide ändern die Folge im Betrieb nicht. Der Satz "Verteilungen und Spans 1024 gegen ... 65536 Ring-Einträge" sagt offen, dass die Prüfgrenze unter dem Betriebsmaximum liegt; der Plan trägt das als Negativ. |

Befunde: 0 HIGH, 0 MEDIUM, 2 LOW, 2 INFO.

Gelöst gegenüber Report 4: V-23, V-24, V-25 (Fänger), V-26 (Chronik-Zeile V-2 auf die Zustände von P09 gezogen), V-27 (Herleitung steht im Plan, mit V-31 als Genauigkeit).

## 4. N-2 (V-17, V-18, V-5, V-25)

- **Grenzen und Herleitung:** 64, 16, 40, 14, 1024 und 70000 sind gegen Betriebsmaxima hergeleitet (13 Schlüssel, Tiefe 3, 13 plus Dubletten, 12 und 3, 5 Namen, 65536 Ring-Einträge); die Herleitung steht im Plan (V-27 gelöst). V-31 nennt zwei Ungenauigkeiten.
- **Grund stimmt:** `filtern` (Schleife über `attrs`), `Kopie` (`maps.Clone`, `slices.Clone`), der Fake und `CountError` (Map-Zugriff) haben keine Fallunterscheidung nach Länge, Position oder Größe eines Codes; gelesen. Meine grünen Läufe K2 und K3 liegen oberhalb der Prüfgrenze 1024 und werden vom Plan getragen.
- **Folge im Betrieb für A02c und alle Dimensionen:** stimmt, wenn der Code nicht nach Größe verzweigt. Für A02c ("Verzweigung nur zwischen 3051 und 99997 bleibt grün") ist die Folge richtig; das Aliasing durch Maskieren (A4) bleibt grün und ist in V-30 genannt.
- **Lücke:** die Dimensionen aus V-30.

## 5. Messung nach [SPEC-128](../../harness/vertraege/SPEC-128-groessenmarke.md)

- **Metrik:** `git diff --numstat a0396b0 HEAD -- src/hexagon/ports/support src/adapters/driven/telemetry`: Einfügungen 3092, Löschungen 1, Summe **3093** (gleich dem Gate, gleich dem Plan). Anteil des Deltas `5e422f4..9b3dac7` an diesen Pfaden: 221 Zeilen im Delta über alle vier Dateien (9 + 4 + 119 + 70 + 19 = 221 in den Testdateien; hinzu der Plan mit 22 + 10, außerhalb der Pfade).
- **Form des Eintrags:** `ueberschreitung: 3100 | <Grund>` im Block, getrennt am ersten `|` (Nr. 3 n); N gleich 3100 = Messung 3093 aufgerundet auf volle Hundert (Nr. 5 d), N größer oder gleich Messung, Grund nicht leer (Nr. 5 e). Das Gate meldet `ok`, keinen Befund.
- **Grund trägt die Überschreitung:** er nennt jetzt die Fix-Runden 5 und 6, die Befunde der Verifikationen 3 und 4, die betroffenen Tests (darunter die neuen `TestFakeStrings`, `TestSnapshotKopieViele`) und dass der Produktcode unverändert blieb. Das passt zum Diff: kein Produktcode im Delta. Auslastung 3093/3000 = 1,03 und 3093/2150 = 1,44 (1,0310; 1,4386), wie im Plan.
- **Marke:** unverändert 3000 (`git diff 5e422f4..9b3dac7` zeigt keine Änderungszeile `marke:`); bestätigt. Marke gleich Formel (Nr. 5 b): 100 x ceil(135 x 2150 / 10000) = 3000.

## 6. Zeit-Fänger `TestFakeCtx` (Frist 30 ms, Wartegrenze 1 s)

Test-Binary einmal gebaut (`go test -c`), je Prozess `-test.run '^TestFakeCtx$' -test.count=1`, im Container `odata-recorder:race`, keine Mutation. Messender: der Verifier. Load Average (1 Minute, `/proc/loadavg`, ist die des Hosts) am Ende der Gruppe; die Host-Last anderer Linien lief zugleich.

| Gruppe | Prozesse | rot | Load Average am Ende |
|---|---|---|---|
| ohne Zusatzlast (`--cpus=4`) | 100 | **0** | 14,74 |
| 24 Rechenschleifen im selben Container (`--cpus=4`) | 100 | **0** | 13,33 |
| 48 Rechenschleifen (`--cpus=1`) | 100 | **0** | 18,60 |
| 96 Rechenschleifen (`--cpus=1`) | 100 | **0** | 14,47 |

Rate ohne Mutation 0 von 400, bei Last von 14 bis 19 (Host) und 96 Rechenschleifen auf einer CPU. Die Rate unter Mutation habe ich nicht gemessen; der Test wird nur durch Verhalten rot. Kein Befund.

## 7. Plan-Zahlen gegen Gate und Tests

| Aussage des Plans | Prüfung | Ergebnis |
|---|---|---|
| 80 Fälle `TestFakeCtx` (5 Zustände x 2 Ausstattungen x 2 Stellen x 4 Eigenschaften) | gelesen; der Test prüft `faelle != 80` und die vier Faktoren selbst | stimmt |
| 6 Fälle `TestFakeStandKopie` für den leeren Stand (3 Maps x 2 Leser) | gelesen: 3 Läufe `i` mit je drei Leseaufrufen; der Kommentar nennt 6 Fälle, der Test zählt sie nicht | Zahl stimmt (3 x 2 Leser; der dritte Aufruf ist ein frischer Fake); keine Zählprüfung im Test, kein Befund |
| 13 Längen `TestFakeStrings` | gezählt: 0, 1, 39, 40, 41, 127, 128, 129, 1000, 4095, 4096, 4097, 70000 | stimmt (13); `len(a) != 4*len(laengen)` im Test |
| Codes lückenlos -50 bis 3050 (3101) und `7 + 3101 + 25*4 + 8` | gerechnet: 3050 - (-50) + 1 = 3101; 7 + 3101 + 100 + 8 = 3216; der Plan nennt 3216 Aufrufe | stimmt; Test prüft die Zahl |
| Tabelle *Zusage · Fänger* 17 Zeilen, `zeilen=17` | gezählt (`| P01 |` bis `| P17 |` der Zusage-Tabelle), Gate | stimmt |
| Messung 3093, 3093/3000 = 1,03 und 3093/2150 = 1,44 | nachgerechnet | stimmt |
| 29 rote Mutationen in 30 von 30, A02c in 0 von 30; Kontrolle 0 von 100 und 0 von 30 | Tabelle der Fix-Runde 6 gezählt: V-23 5 neu + 6 weiter rot (K6, Ma1 bis Ma5) = 11; V-24 5; V-25 3 + 7 + 1 + 2 = 13; 11 + 5 + 13 = 29 | stimmt; die Raten selbst nicht neu gemessen (§10) |
| §6: V-23 bis V-27 mit Ausgang *entfallen* | gelesen | stimmt; V-28 und V-29 sind neu und stehen noch nicht in §6 |
| Kommentar `TestFakeStrings`: "die Grenzen 40, 128 und 4096 mit beiden Nachbarn" | die Längen 39 bis 41, 127 bis 129, 4095 bis 4097 sind in der Liste | stimmt als Aussage über die Liste; nicht als Aussage über Grenzen der Formen (V-31) |

## 8. Konformität gegen die DoD des Plans

| DoD-Punkt | Stand |
|---|---|
| 1 Ports, Typen, Fake | erfüllt; Fänger-Lücken in V-28 bis V-30 |
| 2 Kern des Adapters | erfüllt; `make test-race` ohne Befund (Stufe `race` ohne Cache) |
| 3 `make gates` grün | erfüllt (Exit 0; `test`, `race`, `lint` zusätzlich ohne Cache) |
| 4 Review, 5 Doku-Update, 6 Handbuch | wie in Report 4; das Delta ändert weder Produktcode noch Spezifikation |
| 7 bis 10 Closure-Notiz, Register, Ausgänge, Paarungen | Aufgabe des Planners nach dieser Verifikation |

Messpunkt-Zeile: Handoff-Wert 3093 nachgerechnet und gleich dem Gate; die Ende-Zeile setzt der Planner.

## 9. Ungeprüft

- `make gegenproben` und `harness/gegenproben.stempel` (beauftragt nicht gefahren).
- Die Raten des Implementers (29 Mutationen in 30 Prozessen): nicht wiederholt; die Fänger sind deterministisch, ich habe je Mutation einen Lauf gefahren.
- Last über 96 Rechenschleifen beim Zeit-Fänger und Rate des Zeit-Fängers unter Mutation.
- Mutationen an `support.Spans()` und an den Fenster-Fängern (`TestSnapshotZeitpunkt`), die Reports 3 und 4 beschreiben; die Wertformen über die Stichproben der Vorrunden hinaus.
- K5, H10, A4 ohne Probe-Test (siehe §2).
- Binary-Hash (Produktcode unverändert, nicht neu verglichen).

## 10. Gegengelesen

Dieser Report auf Rohbytes (`LC_ALL=C grep -P`: keine Folge aus Backslash, `u` und vier Hex-Ziffern, kein BOM, kein NBSP) und mit `make docs-check` im Worktree: Exit 0, `d-check: 373 Datei(en) geprüft, 0 Befund(e)`; Rohbytes-Prüfung: 0 Treffer für Escape-Folge, BOM, NBSP.

## 11. Register-Kandidaten (genannt, nicht geschrieben)

- "Zustand und Art des Arguments" als Dimension jeder Allaussage über den Fake (Zustand des ctx beim Aufruf; Span mit und ohne Eltern): V-23 (Rückgabe-ctx) ist geschlossen, V-28 und V-29 gehören zur selben Form an benachbarten Zusagen; zu `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` gehörig.
- "Produkt zweier unbeschränkter Dimensionen" (Länge x Position, V-30 H22): jede Dimension ist einzeln geprüft, das Produkt nicht.
