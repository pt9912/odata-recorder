# Review-Report: slice-betrieb-spans-nachweis (Record, Replay und Mock mit Spans und Zählern) — 2026-11-28

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf.

**Gegenstand:** der eigene Anteil in `a819be8e..4347ffae`: `src/bootstrap/telemetrie_betrieb_test.go` und der Plan; kein Produktcode. Die übrigen Pfade der Spanne stammen aus dem Merge von `main` und sind nicht geprüft.

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`4347ffae`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-28

**Eingangs-Kontext:** Plan `slice-betrieb-spans-nachweis` (BS1 bis BS11, §3, §6); `welle-betrieb` §3; `SPEC-026` bis `SPEC-030`, `SPEC-230`, `SPEC-290`, `SPEC-294` bis `SPEC-296`; `LH-QA-05`, `LH-QA-04`, `LH-FA-01`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `AGENTS.md` §3.

**Urteil: nicht ohne Nacharbeit.** Kein HIGH, 2 MEDIUM, 1 LOW. Der Test belegt das Kriterium der Welle im Wortlaut: eine Testfunktion fährt die drei Modi, findet die fünf Spans mit genau ihren Pflicht-Attributen (ein weiteres Attribut ist rot), und die Zähler sind je Modus exakt gleich den ausgelösten Codes. Beide MEDIUM betreffen, was Zusage und Negative über den Test hinaus sagen.

**Das Kriterium der Welle** (`welle-betrieb` §3):

| Teil im Wortlaut | Belegt durch | Ergebnis |
|---|---|---|
| „Ein Test durchläuft Record, Replay und Mock“ | `TestBetriebSpansUndZaehler` über den Composition Root am OData-Listener | belegt |
| „findet die Spans `SPEC-026` bis `SPEC-030` mit ihren Pflicht-Attributen“ | Gesamtspan je Anfrage, Span des Dienstes über `parent_id`, Vergleich der Attributmenge, Menge der fünf Namen | belegt |
| „ohne Query-Literale und Bodies als Attribut“ | exakte Attributmenge und Suche der elf Werte in der Senke je Modus | belegt |
| „die Zähler je Fehlercode steigen für jeden Code, den der Test auslöst“ | `Snapshot().Errors` gleich den Codes der Fehlerobjekte, mit Anzahl; die Senke nennt außerhalb der Spans allein den Eintrag zu `ODR-E4001`, und der ist gezählt | belegt |

**Vorgehen.** Ein Container (`--network none`, 4 CPU, 4 GiB, ohne Mount, Kopie des Zweigs) unter `flock /tmp/odata-recorder-heavy.lock`; Mutationen am Produktcode allein in der Kopie, vor allem in den Diensten. Messend: Reviewer; Last des Hosts 3,3 beim Start, 5,9 bis 7,4 im 15-Minuten-Mittel; kein OOM. `make gates` und `make gegenproben` nicht gefahren.

| Lauf | Ergebnis |
|---|---|
| unveränderter Test, `-count=300` | grün, 172 s |
| `-cpu=1,2,8 -count=60` | grün, 103 s |
| `-race -count=100` | grün, 82 s, keine Meldung des Detektors |

**Mutationen** (je `-count=3`; rot heißt 3 von 3, aus der Prüfung des Tests, ohne Panic und ohne Baufehler):

| Mutation | Ergebnis |
|---|---|
| Record-Dienst loggt Query, Request-Body, Antwort-Body (3) | rot: „die Senke trägt“ |
| Replay-Dienst loggt Query, Request-Body, Antwort-Body (3) | rot: „die Senke trägt“ |
| Mock-Dienst loggt Query, Request-Body, Antwort-Body; Literal als Wert von `recorder.entity_set` (4) | rot: „die Senke trägt“, beim Attribut auch die Attributmenge |
| Replay-Span ohne Eltern; ohne `End` (2) | rot: Spans unter dem Gesamtspan |
| Record: unbekannter Schlüssel im Log; Span doppelt beendet; `recorder.session_id` fehlt (3) | rot: Verwurf `schluessel`, Verwurf `zustand`, Attributmenge |
| Zähler: `ODR-E5003` nicht gezählt; Zusatzcode im Adapter; Record-Dienst zählt zusätzlich (3) | rot: Stand der Zähler |
| Query erreicht den Upstream nicht (1) | rot: Antwort und Beobachtung des Upstream |
| Record loggt die Query im Eintrag „Upstream ohne Antwort“; bei der Abweisung vor der Annahme (2) | **grün** (F-1); `go test` des Pakets `record` mit derselben Mutation rot |
| Mock loggt die Query bei einem Fehler; den Schlüssel bei der Abweisung `ODR-E6007` (2) | **grün** (F-1); `go test` des Pakets `mock` mit derselben Mutation rot |

## Findings

Befunde ohne Lösungsvorschlag nach dem Schema des Skills.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | In `record` und `mock` trägt allein eine erfolgreiche Anfrage einen Wert aus Query oder Body; keine Anfrage mit Fehlerobjekt trägt einen, und kein akzeptiertes Negativ nennt das (nur `replay` führt den Fehlerweg). Ein Dienst, der die Query oder den Schlüssel auf dem Fehlerweg ausgibt, lässt den Test grün (4 Mutationen); die Tests der Pakete `record` und `mock` fangen diese vier, das Produkt ist also nicht ungeschützt. | `AGENTS.md` §3.15; Reviewer-Skill „Fänger führt nur eine Form der Eingabe“; `LH-QA-04.a` | `src/bootstrap/telemetrie_betrieb_test.go` · `vorher: upServer.Close, status: 502, code: "ODR-E4001"` | ja, `go test` des Tests bleibt mit der Mutation grün | Fänger führt nur den Erfolgsweg, nicht den Fehlerweg |
| F-2 | MEDIUM | Das Negativ zu den Werten der Attribute zählt die ungeführten Werte auf und lässt vier weg: `unbestimmt` (`capture_status`), `nicht_gewaehlt` (`match_result`), `nicht_gelesen` (`coverage_status`) und den Platzhalter `_.` (`entity_set`); ein Maß (n von m) fehlt. Wer es liest, hält diese Wege für geführt. MEDIUM nach der Skill-Regel zur Wiederholung: das Muster war schon mindestens zweimal LOW (Belege der Beobachtung `ein-negativ-benennt-die-ungefangene-klasse-zu-eng`). | `AGENTS.md` §3.15 | `slice-betrieb-spans-nachweis` §3 · `Die Attribute führen nicht jeden Wert ihrer Menge` | nein | Negativ benennt die ungefangene Klasse zu eng |
| F-3 | LOW | Die Abweichung, dass die Mutationen zu BS6 bis BS8 im HTTP-Adapter sitzen statt in den Diensten, steht allein in §3; §6 führt sie nicht, und die Spalte *Mutation* der Tabelle sagt weiter „ein Dienst gibt …“. In der Sache trägt der Fänger: die zehn Mutationen dieses Laufs in den Diensten sind rot. | `AGENTS.md` §3.13, §3.9 | `slice-betrieb-spans-nachweis` §3 · `Die Mutationen zu BS6 bis BS8 sitzen im HTTP-Adapter` | nein | Abweichung vom Plan ohne Zeile in den Risiken |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/telemetrie_betrieb_test.go`: Rot aus dem benannten Grund (Spans, Attribute, Eltern, Zähler, Senke, Verwurf) | geprüft, ohne Befund außer F-1 |
| Ankunft der Werte im Prozess (Upstream und Antwort in `record`; Treffer und zwei Fehltreffer in `replay`; Filter und Antwort in `mock`) | geprüft, ohne Befund |
| Zähler exakt (Gleichheit der Mengen mit Anzahl, kein „mindestens“; eigener Adapter je Modus) | geprüft, ohne Befund |
| Flakiness (Lesen erst nach dem Herunterfahren; `upServer.Close` vor der vierten Anfrage; 580 Läufe) | geprüft, ohne Befund |
| Zeitgrenzen und Lecks (Client mit 10 s ohne Keep-Alive, `ShutdownTimeout` 2 s, `Cleanup` je Prozess, `defer` des Upstream) | geprüft, ohne Befund; Goroutinen nicht gezählt |
| Fixtures gegen den Wert der Mutation (elf verschiedene Werte, Strategie `last` abweichend vom Default) | geprüft, ohne Befund |
| Abweichung vom Planwortlaut BS6 bis BS8 (Mutationen in den Diensten nachgefahren) | geprüft, ohne Befund außer F-3 |
| Kommentare der Testdatei (`AGENTS.md` §3.7, §3.11) | geprüft, ohne Befund |
| Schichten und [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Test im Paket `bootstrap_test`, keine neue Abhängigkeit, keine Suppression) | geprüft, ohne Befund |
| Plan: §1 Abgrenzung, §4 Marke (328 Zeilen der Datei gegen 1200), §6 Risiken | geprüft, ohne Befund außer F-2, F-3 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 1 |
| INFO | 0 |

**Finding-Klassen dieses Laufs:** Fänger führt nur den Erfolgsweg, nicht den Fehlerweg · Negativ benennt die ungefangene Klasse zu eng · Abweichung vom Plan ohne Zeile in den Risiken

## Verdikt

**Merge-blockierend:** ja, wegen F-1 und F-2. Das Kriterium der Welle selbst ist belegt.

**Übergabe:** Findings an den Implementer; die Finding-Klassen zusätzlich in die Closure §7. Der Report ersetzt keine Verifikation.
