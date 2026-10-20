# Verifikation: slice-fehlermodell-konfiguration

* Datum: 2026-10-20
* Rolle: verifier (frischer Kontext)
* Frage: Bauen wir es richtig? (DoD, SPEC-109/SPEC-110, LH-QA-05, LH-FA-12, LH-FA-04)
* Gegenstand: `git diff 89db02c..HEAD` (HEAD bf8e9bd), Plan `docs/plan/planning/done/slice-fehlermodell-konfiguration.md`
* Eingang: Reviewer-Report `2026-10-20-slice-fehlermodell-konfiguration-review.md`
* Sensor selbst gefahren: `make gates`, Exit 0 (gofmt, vet, Tests, a-check, lint, schema-check, migration-check, docs-check). Arbeitsbaum danach sauber.

## Urteil

**Bestanden mit einem MEDIUM-Befund und drei Hinweisen.** Der Code erfüllt die Liefer-Punkte der DoD. Die Closure-Pflichten stehen noch aus (siehe unten).

## DoD gegen Code und Beleg

| DoD-Punkt | Befund |
|---|---|
| Fehlermodell, Abbildung, Test über alle Port-/Adapter-Fehler | erfüllt: `model/fehler.go`, `outbound/abbildung.go`, `TestAbbildungVollstaendig`, `TestSqliteFehlerVollstaendig` vorhanden |
| Konfiguration der Schlüssel, Defaults, `ODR-E2001` | erfüllt: `konfig*.go`; Spezifikation Nr. 1 bis 13 im Code wiedergefunden |
| Tests der Randformen aus §6 | erfüllt: alle 37 Fänger-Namen der Tabelle §3 existieren als `func` (per Suche geprüft) |
| `make gates` grün | selbst gefahren, Exit 0 |
| Review mit Report | Report liegt vor (Verdikt dort: merge-blockierend, 3 MEDIUM); F-1 bis F-5 sind im Code erledigt (siehe unten), es gibt aber keinen Nachweis eines Re-Reviews |
| Closure-Notiz §7, Register, DoD-Häkchen, Risiko-Ausgänge | offen: §7 "Wird bei der Closure gefüllt", alle Häkchen leer, drei Risiken "weiter offen" (zulässig bis Closure) |

## Reviewer-Findings F-1 bis F-5

* F-1 erledigt: `TestKonfigDefaults` setzt `DATA_DIR` nicht und prüft `DataDir == "data"` sowie das Verzeichnis im Arbeitsverzeichnis. Plan-Zeile nennt den Fänger.
* F-2 erledigt: SPEC-110 Nr. 1, 6, 7, 9, 12 und Nr. 6 von SPEC-109 tragen die Randformen (Diff der Spezifikation, Historie 2026-10-20); Plan §6 Raster und Tabelle sind nachgezogen.
* F-3 erledigt: §4 nennt Metrik und Zahl. Nachgemessen: Summe `git diff --numstat 89db02c..HEAD` über Pfade aus §3 (Spezifikation eingeschlossen) = 2288, Produktcode ohne Tests = 1107 minus 19 Zeilen Spezifikation = 1088. Beide Zahlen stimmen mit §4 überein; Marke 800 überschritten, Begründung steht dort.
* F-4 erledigt: der Kommentar vor dem `DATA_DIR`-Default-Test ist jetzt zutreffend.
* F-5 erledigt: `konfig.go` sortiert Verstöße als Struktur (Schlüssel, dann Grund); der Test prüft Schlüsselreihenfolge.

## Plan gegen Code

* §3 Zusage und Fänger: jede Zeile hat einen existierenden Test; der Inhalt (z. B. `none`, Port 1 bis 65535, `[10]` vor `[2]`, Grund-Folge bei `id: 1`) ist im Test wiederzufinden. Der Plan-Eintrag zu `TestKonfigMeldungSortierung` stimmt mit dem Test überein.
* §6 Raster: alle Zellen gefüllt, keine leere Zelle; jede nennt Nummer oder "gilt nicht" mit Grund.
* §4: Überschreitung 2288 gegen 800 mit Metrik genannt (AGENTS.md §3.14), Zahlen nachgemessen und richtig.
* Spec-Straten: Die Änderungen an `spec/spezifikation.md` stehen im Diff vor oder mit dem Code; Erwartung aus §3.12 (Vertrag vor Code) für die Erstfassung erfüllt, die Randformen wurden nachträglich nach dem Review entschieden (Plan §6 deklariert das als "eingetreten").

## Befund

### V-1 MEDIUM — `"id":1` erzeugt zwei Zeilen für denselben Schlüssel; die zweite ist unwahr und im Vertrag nicht entschieden

Ort: `src/bootstrap/konfig_dienste.go`, Funktion `dateiEintrag`, Hilfsfunktion `text`; `konfig_test.go` Zeilen 516 und 750.

Verhalten: Bei `{"id":1,"service_root":"/a"}` meldet `text("id")` "muss eine Zeichenkette sein" und liefert danach `""`. `pflichtfeld` wertet `""` als "Pflichtfeld fehlt" und meldet denselben Schlüssel ein zweites Mal. Ausgabe:

```
ODR-E2001 SERVICES_FILE[0].id: Pflichtfeld fehlt
ODR-E2001 SERVICES_FILE[0].id: muss eine Zeichenkette sein
```

Einordnung gegen SPEC-110:

* Nr. 9 sagt: "ein Feld des falschen Typs … ist `ODR-E2001`" und: nach Nr. 1 gilt nur ein leerer Text oder `null` als fehlend. Ein Feld falschen Typs ist damit nicht "fehlend" laut Vertrag; die Zeile "Pflichtfeld fehlt" widerspricht der Aussage der Meldung (das Feld steht in der Datei).
* Nr. 12 ordnet Zeilen "bei gleichem Schlüssel nach Grund" und fasst gleiche Zeilen zusammen. Das regelt die Reihenfolge mehrerer Gründe, nicht, ob ein falscher Typ zugleich "fehlend" meldet. Ein Satz, der zwei Gründe je falschem Typ vorschreibt, steht nicht in der Spezifikation.
* Der Plan hält die Doppelmeldung als Testziel fest (§3 Zeile Meldungssortierung: "`id: 1` meldet `Pflichtfeld fehlt` vor `muss eine Zeichenkette sein`"). Die Antwort wurde also im Plan und im Code getroffen, nicht in SPEC-110.

Urteil: keine Verletzung des Wortlauts von SPEC-110, aber eine **ungeregelte Randform, im Code entschieden** (AGENTS.md §3.12, §3.16: MEDIUM). Sie ist zudem irreführend und überflüssig; sie dient im Test nur dazu, die Grund-Sortierung beobachtbar zu machen. Der Fall gilt gleichermaßen für `service_root` und `upstream_url` (bei `record`).

Auflösung (Entscheidung des Architect, nicht des Verifiers): entweder Nr. 9 ergänzen ("ein Feld falschen Typs meldet nur den Typ, nicht zusätzlich das Fehlen") und den Code mit Test ändern, oder die Doppelmeldung ausdrücklich in Nr. 9/12 festschreiben. Die Sortierung nach Grund bleibt mit anderem Testfall beobachtbar (Verstöße gleicher Schlüssel aus anderer Quelle, z. B. unbekanntes Feld und Typ).

## Hinweise (kein Mangel)

* H-1: Es gibt keinen Nachweis eines Re-Reviews nach den Korrekturen; der Reviewer-Report trägt weiter "merge-blockierend: ja" und keine Erledigungs-Spalte. Der Planner soll entscheiden, ob für die Korrekturen ein kurzer Nachlauf nötig ist; V-1 ist in beiden Fällen offen.
* H-2: Folgen aus Reviewer F-6 (Mutationen, die grün bleiben: `LINK_REWRITE`-Wert `rewrite`, Dateigröße, leeres `id`) sind als INFO akzeptiert; unverändert.
* H-3: Der Start-Befehl ruft `LadeOderMelde` nicht (Plan §1, ausdrücklich ausgeschlossen); "Start scheitert" ist daher nur auf Ebene der Funktion belegt. Der Nehmer (HTTP-/Start-Slice) muss die Verdrahtung zusagen; vor der Closure prüfen, dass er den Punkt in §1 oder DoD trägt (AGENTS.md §3.13).

## Ergebnis

Bestanden. Vor `done/` offen: V-1 (MEDIUM, Entscheidung des Architect), Closure-Notiz §7, DoD-Häkchen, Register-Notiz, Risiko-Ausgänge.
