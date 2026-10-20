# Nachtrag: Re-Review slice-fehlermodell-konfiguration

* Datum: 2026-10-20
* Rolle: reviewer (frischer Kontext, kein Self-Review)
* Review-Art: Code, enger Re-Review
* Gegenstand: `git diff 89db02c..HEAD`, Schwerpunkt `bf8e9bd` (Randformen im Code) und `a6f1a4b` (V-1: falscher Typ meldet nur den Typ); Spezifikation seit dem ersten Review in `8732e5e` und `eb7ff59`
* Vorbericht: `docs/reviews/2026-10-20-slice-fehlermodell-konfiguration-review.md` (F-1 bis F-9), `...-verifikation.md`
* Maßstab: `SPEC-109`, `SPEC-110`, Plan §3 (Zusage · Fänger), §6 (Raster), `AGENTS.md` §3.7, §3.11, §3.12, §3.15, §3.16. Nicht geprüft: DoD.
* Methode: Mutationen im Arbeitsbaum, je Lauf `make test` (Docker, netzlos), danach `git checkout` der Datei; eine Sondierung mit einer temporären Testdatei, danach gelöscht. Arbeitsbaum am Ende sauber bis auf diesen Report. Mutationen, die nur wegen `gofmt` oder ungenutztem Import rot wurden, sind wiederholt und hier nicht gezählt.

## Zusammenfassung

Keine HIGH. Zwei MEDIUM, ein LOW, drei INFO.
Die Findings F-1 (Default `DATA_DIR`), F-4 (Kommentar), F-5 (Sortierung nach Schlüssel) und F-3 (Marke im Plan benannt) des Vorberichts sind behoben und belegt. F-2 ist durch die Entscheidungen in `SPEC-110` aufgelöst.
Zwei neue Befunde betreffen Zusagen der neuen Fassung, die Mutationen überleben oder die Wirklichkeit nicht beschreiben (N-1, N-2).

## Findings

### N-1 MEDIUM — Zusage „ein Feld falschen Typs meldet nur den Typ, bei `upstream_url` in jedem Modus“ hat nur im Modus `record` einen Fänger
* kategorie: MEDIUM
* quelle: AGENTS.md §3.15 (Zusage ohne Fänger), `SPEC-110` Nr. 9 („bei `upstream_url` in jedem Modus, weil der Typ vor der Pflicht des Modus geprüft wird“)
* pfad: `src/bootstrap/konfig_test.go` · `func TestKonfigDienstefeldTyp` (alle drei Fälle mit `"MODE=record"`); `src/bootstrap/konfig_dienste.go` · `muss eine Zeichenkette sein`
* befund: Mutation „Typfehler von `upstream_url` bleibt im Modus `replay` ungemeldet“ (in `text` vor dem `melde`: `if feld == "upstream_url" && !record { return "", true }`) lässt `make test` grün. Der Satz „in jedem Modus“ steht in der Spezifikation und im Raster (Zeile `SERVICES_FILE` lesen, Zelle „falscher Typ“), die Tabelle *Zusage · Fänger* nennt nur `record`. Failure-Szenario: `{"upstream_url": 1}` unter `MODE=replay` startet ohne Meldung mit leerer URL, obwohl der Vertrag `ODR-E2001` sagt.
* verifizierbar: ja (`make test` mit der Mutation)
* klasse: Zusage mit Moduszusatz, Fänger deckt einen Modus

### N-2 MEDIUM — `SPEC-110` Nr. 12 behauptet „kein Fall meldet zwei verschiedene Gründe für einen Schlüssel“; `UPSTREAM_TLS_VERIFY=yes` meldet zwei, die zweite führt in die Irre
* kategorie: MEDIUM
* quelle: AGENTS.md §3.12, §3.16 (im Code entschiedene Zelle), §3.11 (Zusage, die keine Prüfung deckt); `SPEC-110` Nr. 3, 10, 12
* pfad: `src/bootstrap/konfig_dienste.go` · `func (p *pruefer) liesTLS`; `spec/spezifikation.md` · „in dieser Version meldet kein Fall zwei verschiedene Gründe für einen Schlüssel“; Plan §3 · „aber in dieser Version nicht beobachtbar“
* befund: Sondierung (temporärer Test, gelöscht): `UPSTREAM_TLS_VERIFY=yes` und `=1` liefern zwei Zeilen: `UPSTREAM_TLS_VERIFY: Wahrheitswert muss true oder false sein` und `UPSTREAM_TLS_VERIFY: false braucht UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true`. `optionalerWahrheitswert` gibt für den ungültigen Wert `false` zurück, und `liesTLS` wertet das als gesetztes `false` ohne `SKIP_VERIFY`. Damit ist die Sortierung nach Grund beobachtbar (die Spezifikation und der Plan sagen das Gegenteil), der Nr. 10-Widerspruch wird für einen Wert gemeldet, den niemand `false` genannt hat, und die zweite Zeile rät, `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` zu setzen, also die Zertifikatsprüfung abzuschalten. Failure-Szenario: Tippfehler `yes` in einem Sicherheitsschlüssel führt den Betreiber zur unsicheren Einstellung. Der Fall steht weder in `TestKonfigTLSWiderspruch` (nur gültige Werte, prüft nur „Fehler ja/nein“) noch im Raster (Zeile „Aufzählung und Wahrheitswert“ × „ungültig“ gegen Nr. 10). Die Mutation `M5b` (Sortierung nach Grund entfernt) bleibt grün, die Zusage der Sortierung nach Grund ist damit weiterhin ohne Fänger, obwohl der Fall, der sie beobachtbar macht, existiert.
* Entscheidung gehört an den Architect (Nr. 10 bei ungültigem Wahrheitswert: eine Zeile oder zwei; Nr. 12 Satz „kein Fall“ stimmt nicht).
* verifizierbar: ja (Sondierung wie beschrieben)
* klasse: Randform im Code entschieden / Vertragsaussage über Beobachtbarkeit falsch

### N-3 LOW — Testkommentar beschreibt die verworfene Alternative
* kategorie: LOW (Konventions-Anker: AGENTS.md §3.7; der Reviewer-Skill führt „Ohne X wäre …“ als HIGH-Muster, hier ein Testkommentar mit einem Satz und ein Fänger vorhanden, deshalb LOW; die Einstufung ist begründet, nicht stillschweigend)
* quelle: AGENTS.md §3.7
* pfad: `src/bootstrap/konfig_test.go` · `obwohl '.' < ':' die ganze Zeile umordnen würde`
* befund: Der Kommentar beschreibt die Sortierung der ganzen Zeile, die der Code nicht hat, im Konjunktiv. Der Indikativ über den Zustand („Schlüssel `SERVICES_FILE[0]` steht vor `SERVICES_FILE[0].id`“) trüge dieselbe Zusage.
* verifizierbar: nein
* klasse: Kommentar beschreibt verworfene Alternative

### N-4 INFO — Aufgelöste Findings des Vorberichts
* F-1: `TestKonfigDefaults` prüft `DataDir == "data"` und `os.Stat("data")` nach `t.Chdir`; Mutation `"data"` → `"daten"` (beide Stellen): rot, `TestKonfigDefaults`.
* F-4: der Kommentar ist durch den Absatz „DATA_DIR ungesetzt: der Default ist `data` …“ ersetzt und deckt, was der Code prüft.
* F-5: Sortierung nach Schlüssel (Bytefolge), dann Grund (`cmp.Or`), `Compact` auf `verstoss`; Mutationen „ohne Sortierung“ und „ohne `Compact`“: rot (`TestKonfigMeldung`, `TestKonfigMeldungSortierung`). Grund-Sortierung: siehe N-2.
* F-3: Plan §4 trägt die Überschreitung und die Begründung. Der genannte Wert 2315 weicht von `git diff --numstat 89db02c..HEAD` über `src/hexagon src/adapters src/bootstrap` (2294, Produktcode 1097) um 21 ab; Produktcode stimmt. Kein Befund, die Zahl wandert mit jedem Commit.
* F-2: durch die Spezifikationsänderungen aufgelöst; der Plan-Satz „kein Test legt eine dieser Antworten fest“ ist durch die Fänger-Zeilen ersetzt.

### N-5 INFO — Überlebende Mutationen ohne Befund
* `M5b` (Sortierung nach Grund entfernt): grün, im Plan als akzeptiertes Negativ benannt; durch N-2 nicht mehr unbeobachtbar.
* Aus dem Vorbericht unverändert (F-6): `LimitReader`-Grenze, leeres `id` doppelt gemeldet.

### N-6 INFO — Aufgabenkreis der Verifikation
`verifizierbar`-Aussagen zu DoD und Plan-Zeilenzahlen gehören dem Verifier.

## Auflösung der neuen und geänderten Zeilen der Tabelle *Zusage · Fänger* und des Rasters

| Zeile / Zelle | Mutation | Ergebnis |
|---|---|---|
| Defaults, `DATA_DIR=data` | `"data"` → `"daten"` | rot, `TestKonfigDefaults` |
| `none` nur allein und klein, nur bei zwei Mengen | `EqualFold` → `==` bei der Erkennung; Exaktheit des Eintrags auf `EqualFold`; Bedingung auf die zwei Mengen entfernt; nur den ersten Eintrag prüfen; `VOLATILE_HEADERS` aus den ersetzenden Listen | rot, `TestKonfigNone` (jeweils) |
| Header kanonisch | Kanonisierung entfernt/verfälscht (`ToLower`) | rot, `TestKonfigDefaults`, `TestKonfigListenErgaenzen`, `TestKonfigHeaderSchreibweise` |
| Menge ohne Doppelte | `Compact` in `menge` entfernt | rot, `TestKonfigHeaderSchreibweise`, `TestKonfigListe`, `TestKonfigListenErgaenzen` |
| Port der URL 1 bis 65535, `host:` | Untergrenze 0; Obergrenze 65536; Prüfung auf `host:` abgeschaltet | rot, `TestKonfigURL` (jeweils) |
| `LISTEN` Port 1 bis 65535 | Untergrenze 0; Obergrenze 65536 | rot, `TestKonfigLISTEN` |
| Dienste-Datei trimmt, `null`/leer wie fehlend | `TrimSpace` entfernt | rot, `TestKonfigServicesFile` (Leerraum, `null`) |
| Element `null` ist kein Objekt | `felder == nil` entfernt | rot, `TestKonfigServicesFile/Element_null_ist_kein_Objekt` |
| Form von `UPSTREAM_URL` in jedem Modus | Formprüfung nur im Modus `record` | rot, `TestKonfigUpstreamModus`, `TestKonfigServicesFile` |
| Pflicht nur bei gültigem Modus | `record` auch bei ungültigem Modus | rot, `TestKonfigModus`, `TestKonfigModusUngueltig`, `TestKonfigUpstreamModus` |
| Falscher Typ meldet nur den Typ („keine Formprüfung“) | Typfehler lässt die Folgeprüfung laufen (`ok=true`); nur bei `id` | rot, `TestKonfigDienstefeldTyp` (alle Felder bzw. `id`) |
| Falscher Typ von `upstream_url` in jedem Modus | Typfehler nur im Modus `record` melden | **grün** (N-1) |
| Meldung: Schlüssel, dann Grund; gleiche Zeilen einmal | ohne Sortierung; ohne `Compact`; ohne Grund-Kriterium | rot (`TestKonfigMeldung`, `TestKonfigMeldungSortierung`); ohne Grund-Kriterium **grün** (akzeptiertes Negativ, siehe N-2) |
| Index als Text (`[10]` vor `[2]`) | durch die Sortierung nach Schlüssel mitgefangen | rot mit „ohne Sortierung“ |

Raster: die Zellen „`none` nur allein und klein“, „Port außerhalb 1 bis 65535“, „`null` und leer wie fehlend“, „falscher Typ“ und „Feld doppelt“ sind mit den Mutationen oben aufgelöst; „Feld doppelt: letzter Wert“ ist von `encoding/json` getragen und nicht mutierbar, ohne den Code zu ersetzen. Zelle „`MODE` ungültig, Pflicht von `UPSTREAM_URL` entfällt“ rot. Nicht im Raster aufgelöst: Wahrheitswert ungültig × Zusammenhang Nr. 10 (N-2).

## Prüfung der Schwerpunkte

* Formprüfung bei falschem Typ: Fänger `TestKonfigDienstefeldTyp` fängt die Weiterprüfung (Pflichtfeld, Form, Pflicht des Modus) in `record`; nicht in `replay` (N-1). Geprüfte Typen im Test: nur Zahl; Wahrheitswert, Array und Objekt teilen dieselbe Ablehnung durch `json.Unmarshal` in `*string` und sind nicht gesondert fangbar.
* Sortierung/Compact: Schlüssel gefangen; Grund nur Totalordnung, im Code beobachtbar durch N-2.
* `none` in allen Listen: alle fünf Listen und die Schreibweisen im Test; Mutationen rot.
* Port-Grenzen: beide Grenzen für `LISTEN` und URL gefangen.
* Kommentare (§3.11): `TestKonfigDienstefeldTyp`-Kopf nennt „weder Pflichtfeld fehlt noch die Form noch im Modus record erforderlich“, der Test prüft genau diese Zeilenmenge in `record`; die Zusage „jeden Modus“ trägt nur die Spezifikation (N-1). Spezifikation Nr. 12 „kein Fall“ stimmt nicht (N-2).
* Neue ungeregelte Randformen: ungültiger Wahrheitswert × Nr. 10 (N-2); sonst keine gefunden (LISTEN mit führenden Nullen im Port und IPv4 in Klammern sind von Nr. 7 gedeckt).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/konfig.go` | geprüft, ohne Befund (Sortierung, `Compact`, Meldung ohne Wert) |
| `src/bootstrap/konfig_werte.go` | geprüft, ohne Befund (`none`, Port, Kanonisierung; Mutationen rot) |
| `src/bootstrap/konfig_dienste.go` | geprüft, Befunde N-1, N-2 |
| `src/bootstrap/konfig_test.go` | geprüft, Befunde N-1, N-3 |
| `spec/spezifikation.md` SPEC-109/110 (Änderungen seit `89db02c`) | geprüft, Befund N-2 (Satz in Nr. 12) |
| `src/hexagon/**`, `src/adapters/**` | nicht geändert seit dem ersten Review, ohne neuen Befund |
| Geheimnis in Meldung | geprüft, ohne Befund |
| Suppression | geprüft, ohne Befund |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 1 |
| INFO | 3 |

**Finding-Klassen:** Zusage mit Moduszusatz, Fänger deckt einen Modus · Randform im Code entschieden / Vertragsaussage über Beobachtbarkeit falsch · Kommentar beschreibt verworfene Alternative

## Verdikt

Der Merge-Block aus dem ersten Review (drei MEDIUM) ist aufgehoben: F-1, F-2, F-3 sind behoben bzw. entschieden und belegt. Die zwei neuen MEDIUM (N-1, N-2) sind keine Wiederkehr der alten Befunde, sondern Befunde an der neuen Fassung; N-1 verlangt einen Test, N-2 eine Entscheidung des Architect (Spezifikation) und danach Code und Test. Ob sie den Merge blockieren, ist nach Skill-Lage ja (MEDIUM), in diesem Report nicht herabgestuft; nach Behebung ist ein erneuter enger Lauf nur für diese zwei Stellen nötig.
