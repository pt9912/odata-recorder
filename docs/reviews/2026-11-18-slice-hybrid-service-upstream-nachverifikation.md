# Nachverifikation slice-hybrid-service-upstream: Auflage V-1

Rolle: Verifier, frischer Kontext. Gegenstand: Fix von V-1 in `caaa546d` (`src/hexagon/services/hybrid/upstream_body_test.go`, Plan §3), Stand `93eee5ec`. Bezug: `LH-FA-12`.

**Ergebnis: PASS**

## (1) Menge aus dem Wortlaut von SPEC-284

Zusage: Nr. 5 d und Nr. 10, der Dienst schließt den Body der Anfrage nie. Wege zu U nach Nr. 3 c und d: nach M (Ergebnis `Upstream` von `Entscheide`) und nach R (`ODR-E5001`, H5 gibt `Upstream`); mehr Wege gibt es nicht (Nr. 3 f). Formen des Bodys nach Nr. 5 b und c: gepuffert mit `Laenge` -1 und 4, `Laenge` 0 mit Leser, über `MaxBytes`, Lesefehler. Ausgänge nach Nr. 8 und 11: Erfolg, die sieben Fehler der Tabelle, Kontextfehler von `Senden`, Abbruch vor U (Nr. 3 e).

`TestBodyNieGeschlossenUpstream` läuft als Dreifach-Schleife über 2 Wege (`wege`), 5 Formen und 10 Ausgänge, prüft je Fall U genau einmal bzw. nie und `closes == 0`, und rechnet die Zahl 100 gegen die Mengen (`2*5*(1+len(fehlerTabelle())+2)`). Jedes Element ist geführt, die Mengen stimmen mit meiner Bildung überein.

## (2) Eigene Mutationen (`tools/harness/mutationslauf.sh`, unter `flock`, Last 7 von 4, Messender: ich)

Alle schließen `a.Body` per Typ-Assertion auf `Close() error`. Die erste Fassung der Liste scheiterte an der Übersetzung (`io.Reader` ohne `Close`) und zählt nicht. Die zweite Fassung lief ohne Übersetzungsfehler.

| Mutation | Stelle | Ergebnis |
|---|---|---|
| n1 Body vor `Senden` geschlossen | `upstream.go`, vor `Senden` | rot |
| n2 Body nach `Senden` geschlossen (Erfolg und Fehler) | `upstream.go`, vor `if err != nil` | rot |
| n3 Body im Fehlerpfad geschlossen | `upstream.go`, nach `verwirf` | rot |
| n4 Body im Abbruchpfad vor U geschlossen | `upstream.go`, Zweig `ctx.Err()` | rot |
| n5 Body an H5 nach Replay-Miss geschlossen | `verteiler.go`, vor `upstreamFallback` | rot |
| n6 Body an H3 bei `policy.Upstream` geschlossen | `verteiler.go`, `case policy.Upstream` | rot |
| n7 Body nach dem Puffern geschlossen | `puffer.go`, nach `io.ReadAll` | rot |

7 von 7 rot, 0 grün, 0 flatternd, 0 Übersetzungsfehler, je ein Prozess ohne Detektor. n6 und n7 sind Elemente, die der Befund V-1 und die Fix-Mutationen im Plan (Zeile Z16) nicht nannten (H3-Weg, Puffern). Der Grund ist der benannte: der einzige Fehlerpfad des Tests bei geschlossenem Body ist die Meldung „Body N-mal geschlossen“ (der Lauf trennt `uebersetzungsfehler` von `rot`).

## (3) Test ohne Mutation

Die Nullprobe des Treibers (unveränderter Baum, derselbe Test) war in jedem der drei Läufe grün, sonst bricht der Treiber mit Exit 2 ab. Der Test scheitert also nicht am Aufbau. Eine Mutation, die den Text nicht ändert, weist der Treiber ab (n0 „neu ändert nichts“).

## (4) V-2 bis V-4 im Plan (`docs/plan/planning/done/slice-hybrid-service-upstream.md`)

- V-2: Zeile Z12 (U) nennt den Fänger der Probe nur als „ohne `ODR-E6003`“ und „ohne beide `unsupported`-Codes“, rot über `Pufferung` aus Teil 1, und verweist für die Probe ohne `ODR-E5001`, `ODR-E6001`, `ODR-E6002` auf das Negativ (a) als äquivalent. Das stimmt mit der Messung überein (v30, v31, w03, w04 grün).
- V-3: §6 trägt das Negativ „Zulassungsliste der Importe“ mit Ausgang akzeptiert; §1 (Übernommenes Negativ) führt es mit Grund und Folge. Stimmt.
- V-4: Negativ (f) nennt Grund (`Entscheide` rein, am konkreten Typ nicht zählbar), Folge (ohne Wirkung) und die gefangenen Teile (kein M, kein R). Stimmt mit der Messung (x01, x02, x04 rot).
- Z16: Zeile und V-1-Raster nennen 2 · 5 · 10 = 100; der Test führt dieselbe Zahl.

## Offen

Nichts. Es wurde kein Produkt- oder Plan-Inhalt geändert; die Mutationsliste lag im Scratchpad, nicht im Repo.
