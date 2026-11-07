# Verifikation slice-mock-schreiben

**Rolle:** Verifier (Frage: Bauen wir es richtig, gegen Plan, DoD, Spezifikation).
**Gegenstand:** Zweig `slice-mock-schreiben`, HEAD `db95af66` (Fix-Runde nach dem Review `92fdbb0b`, `main` bis `3efc64f5` gemerged), Basis `bad8f76`.
**Datum:** 2026-11-07. **Bezug:** `LH-FA-07`, `LH-FA-09`, `LH-QA-02`, `LH-QA-03`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
**Stand des Baums:** `git status` im Worktree leer, der Implementer schrieb nicht mehr; nichts Uncommittetes.

## Urteil

**Bestanden mit Auflagen.** Alle Gates sind im eigenen Lauf grün, die Tabelle *Zusage · Fänger* trägt gegen den Vertrag, die Fixes des Reviews decken ihre Klassen bis auf zwei Elemente (V-1, V-2), und die Fenster der Sperre sind in jedem Lauf rot. Es gibt keinen Befund über LOW. Die Auflagen (V-1 bis V-5) sind Nachzüge in Tests und Plan; sie blockieren die Closure nicht, gehören aber vor das `done/` an den Planner (V-4, V-5) bzw. an den Implementer (V-1, V-2, V-3).

## 1. Belege gegen Behauptung

| Behauptung des Implementers | Beleg des Verifiers |
|---|---|
| `make gates` grün | eigener Lauf, Log `scratchpad/verif-mockschr/gates.log`: Exit 0; `groessenmarke slice-mock-schreiben.md: ok, Messung 4544 von Marke 11100` |
| Messung 4544 | eigene Summe aus `git diff --numstat bad8f76..HEAD` über die Pfade des Blocks: 4544 |
| keine `//nolint`, `.golangci.yml`/`.a-check.yml` unverändert | `git diff bad8f76..HEAD` über beide Dateien leer; `//nolint` nur in Prosa von Reports, in keiner `.go`-Datei |
| `model.AlleCodes()` = 39, `ODR-E6009` | `fehler.go` (Zeile hinter `ODR-E6008`, 409, `mock`, `Entity key already exists`), `fehler_test.go` 38 auf 39 und Status-Schleife 0 bis 999, `katalog-ausnahmen.txt` mit Grund; Katalog-Gate im Lauf grün |
| Tabelle W01 bis W81: jedes Test-Symbol existiert | 62 verschiedene Symbole der Spalte *Fänger* und alle der Tabelle *Mutationen*: per `grep "^func <Symbol>("` unter `src/` 0 fehlend (F-2 behoben) |
| Binary unverändert | kein Diff unter `src/bootstrap/` außer `mock_schreiben_ende_zu_ende_test.go` und zwei Zeilen im Lesetest; `main`/Katalog nur über die Ausnahme-Zeile |
| Rohbytes (`AGENTS.md` §3.18) | in den neuen `.go`-Dateien keine Folge Backslash-`u`-vier Hex-Ziffern (`LC_ALL=C grep -P`), kein BOM/NBSP/Nullbreite in 19 geprüften Dateien; `make zeichen-check` grün |
| Z-Liste M24 = 36 | `TestZweigeGleichzeitigMutation` zählt `len(zweige) == 36` (4+8+5+13+6), Gruppen stimmen mit M24 überein |

## 2. Plan-gegen-Code und Tabelle gegen den Vertrag

Ich habe `SPEC-192` Nr. 1 bis 9, `SPEC-193` Nr. 1 bis 7 und `SPEC-194` Nr. 1 bis 5 Satz für Satz gegen `schreiben.go` und `zustand_schreiben.go` gelesen. Keine Abweichung des Codes vom Vertrag; die Rangfolge (e) bis (l), die Aufrufe S17 bis S26 mit `ctx` des Aufrufs, die Coverage nur bei fehlendem Eintrag von `PATCH`, `PUT`, `DELETE`, Befund vor Typ, `Version` plus 1 (fehlend 1, Tombstone plus 1), ein `Advance` mit in der Transaktion gelesener Revision, Antwort in der Transaktion (Rollback bei Schreiberfehler), `204` mit nur dem Kennzeichen, `201` mit genau drei Headern, `PATCH {}` schreibt, Tombstone behält `Payload`/`Known`/`Completeness`, nicht genannte nullbare Properties `null` und bekannt, `[REDACTED]` nur bei `Edm.String`, Widerspruch von Body und URL byte-gleich: alle im Code.
`SPEC-180` Nr. 3 (F-15): `kuerzeOrt` liefert den längsten Anfang bis 256 Bytes am Zeichenrand, `...` nur bei Kürzung, Feld `Ort` ungekürzt; die Obergrenze 1280 Bytes folgt (256 Bytes, im ungünstigen Fall `\x00` je 4 Zeichen, plus Rahmen).
Abweichung A2 (vier Lesetests): `rang_test.go` (`POST` nach `OPTIONS`, notwendig, weil `POST` jetzt läuft), `lesen_fehler_test.go` (13 auf 15 Codes, 38 auf 39), `lesen_gleichzeitig_test.go`, `konfig_test.go`. Nichts abgeschwächt. Die engste Lockerung ist `TestQuelltextOhneVersion`: die Schreibweisen `es.Version` und `Version: version` fallen aus der Verbotsprüfung (Entity-State, nicht ETag); jede andere Nennung bleibt Befund. Akzeptiert.
Akzeptierte Negative W66, W79, W80, W81 tragen Grund und Folge im Plan und stimmen mit den *Ungeprüft*-Listen von `SPEC-192` und `SPEC-193` überein.

## 3. Fixes gegen die Klasse (§3.17): eigene Mutationen

Werkzeug: `tools/harness/mutationslauf.sh`, `MUTATIONSLAUF_SPEICHER=4g`, unter `flock /tmp/odata-recorder-heavy.lock`, Image `odata-recorder:race` (Baumkopie im Container). Messender: Verifier auf dem Host; Last des Hosts 8,3 bis 11,1 bei vier CPUs des Containers (`last=` der Ergebniszeilen). Listen: `scratchpad/verif-mockschr/liste1.json` (51), `liste2b.json` (17); Logs `lauf1.log`, `lauf2.log`.

**Lauf 1 (Klasse `andere`, 51 Mutationen, alle an Elementen, die der Befund nicht nannte oder die Menge erweitern):** 47 rot (alle im benannten Test, kein `rot-paket`), 1 Übersetzungsfehler meiner Liste (`declared and not used`, in Lauf 2 korrigiert wiederholt), 3 grün (siehe unten).

| Menge (Befund) | Elemente, an denen ich mutierte | Ergebnis |
|---|---|---|
| ctx-Identität (F-1) S2, S3, S5, S17 bis S24, S26 (`context.WithoutCancel(ctx)` je Stelle) | 12 Stellen; S25 (`Adresse`) hat kein `ctx` | 12 von 12 rot in `TestKontextMutation`; der Befund nannte S18 bis S26, ich habe zusätzlich S2, S3, S5 und S17 gemessen |
| Zustandsmatrix Methode × Eintrag | Coverage-Status `unknown` wie `complete`; Coverage-Lesen bei `POST` über Tombstone und bei Tombstone von PATCH/PUT/DELETE; Version bei `POST` und bei `PUT`; `Version++` entfernt; DELETE leert `KnownProperties` bzw. setzt `partial`; Typabgleich bei `PUT` ausgelassen | 9 von 9 rot |
| Maske / Known / nicht genannte nullbare | nullbare, nicht genannte Property nicht bekannt; `Compact` entfernt; `PATCH {}` verändert Payload | 3 von 3 rot |
| Payload-Kanonform | Sortierung entfernt; `U+001F` nicht escaped (`< 0x1f`); `U+007F` escaped; `int32` als String | 4 von 4 rot |
| Antwort | Fehler der Antwort verschluckt (Rollback), `Content-Type` im `204`, `204` mit `nil`-Body, `X-Recorder-Coverage` im `201`, leerer `ContentType` bzw. leerer Wert des Kennzeichens nicht geprüft | 6 von 6 rot |
| Methode, Paarung, Optionen, Eingabe | `NurLesend` ohne `DELETE`; `$filter` ohne `PUT`; `Eingabe` nil verlangt auch `DELETE`; Paarung `DELETE` an Collection; Marker ohne Beachtung der Schreibweise; Widerspruch nur erste Schlüssel-Property; fehlende Schlüssel-Property widerspricht; `Transaktion` nil angenommen | 8 von 8 rot |
| `erwartet` | `Advance` mit Revision von vor der Transaktion | siehe V-3 |
| F-15 Kürzung | Grenze 257; `...` fehlt; Startindex `n-1`; Schwelle `<= 257` | 4 von 4 rot |
| Schlüssel mit Apostroph (F-3) | `Replace(…, 2)` statt `ReplaceAll`; Schlusszeichen `'` fehlt; Prüfung auf `Edm.String` entfernt | Schlusszeichen rot; `Replace(…, 2)` **grün** (V-1); Prüfung auf `Edm.String` **grün**, gleichwertig, weil nur `Edm.String` die Art `ArtString` trägt |
| Typ-Mitglied | `typ == setTyp` zu `HasPrefix(typ, setTyp)` | **grün** (V-2) |

**Lauf 2 (Klasse `nebenläufig`, 17 Mutationen), Raten je Prozess mit `-test.count=1`, Messender Verifier, Last 8,3 bis 11,1:**

| Zeile | Mutation | Test | mit Detektor | ohne Detektor |
|---|---|---|---|---|
| W69 | Prüfen vor der Transaktion (Lesen von `Entities.Get` über `Konfig.Entities` vor `WithWorkspaceTransaction`, Entscheidung in ihr auf dem alten Stand; mein Textersatz von `transaktion` bis zum Lesen in `imWorkspace`) | `TestPostWettlauf` | 50 von 50 | 50 von 50 |
| W70 | dieselbe Mutation | `TestDeleteWettlauf` | 50 von 50 | 50 von 50 |
| W71 | `erwartet` aus `mu.w.Revision` (vor der Transaktion) | `TestRevisionWettlauf` | 50 von 50 | 50 von 50 |
| W75 | `m.k.MaxBytes++` in 12 Zweigen, die der Implementer nicht nannte: Tombstone, fehlend bei `complete`, reservierter Text, Typ, `ErrRevisionVeraltet`, Fehler von `Put`, Widerspruch, Workspace gelöscht, leerer Schlüssel, Set unbekannt, Fehler des Lesers, Schlüssel der URL ungültig | `TestZweigeGleichzeitigMutation` | je 12 von 12 | nicht gefahren (der Fänger ist der Detektor, `SPEC-122` Nr. 8) |

Damit sind die Wirkungs-Fänger W69, W70, W71 in jedem Lauf rot, auch ohne Detektor (F-4 gilt für W70 mit dem Haltepunkt `sperreAmEingang`). Mein Textersatz für W69/W70 ist ein anderer als der des Implementers (`vorabLesen`), kommt aber zum selben Ergebnis. Die Z-Liste: mit den drei Zweigen des Implementers und meinen zwölf tragen 15 von 36 Zweigen eine eigene Schreib-Mutation; 21 sind nicht gemessen (nur Zweige in freien Funktionen wie `pruefeMethodeUndRessource` können nicht mit `m.k.MaxBytes++` mutiert werden, ohne die Signatur zu ändern; der Zugriff dort hat keinen `m`).
Nicht selbst gefahren: W48, W49, W56, die Positionen im Set (M10/M25) und `Prefer`; deren Mutationen stehen im Plan (*Fix-Runde*, letzte Zeile), ich habe sie nicht wiederholt. W69/W70 vollständig, W71 vollständig.

## 4. Befunde

| Nr. | Klasse | Befund | Auflage |
|---|---|---|---|
| V-1 | LOW | Menge „Apostroph im Schlüssel“ (F-3) führt 0, 1 und 2 Apostrophe je Text (`O'Neil`, `''`, `'`, `x''y`); `strings.Replace(w.Text, "'", "''", 2)` bleibt grün, drei und mehr Apostrophe sind ungefangen. §3.17 verlangt die Menge „jeder Apostroph“. Folge im Betrieb: ein Schlüssel mit drei Apostrophen widerspräche zu Unrecht (`ODR-E6007`). | Implementer: einen Text mit drei Apostrophen (etwa `a'b'c'd`, aus Bytewerten) in die Menge von `TestSchluesselWiderspruch`, Mutation rot sehen. |
| V-2 | LOW | `typMitglied` vergleicht `typ == setTyp`; `HasPrefix(typ, setTyp)` bleibt grün, weil kein abgeleiteter Typ der Fixtures den Namen des Set-Typs als Anfang trägt (Fixture gleicht dem Wert der Mutation, §3.15). Folge: ein abgeleiteter Typ `<Set-Typ>X` verlöre `@odata.type`. | Implementer: abgeleiteten Typ mit dem Set-Typ als Namensanfang in M11/M12 (`TestPayloadKanonform`). |
| V-3 | LOW | Zeile W45 nennt `TestVersionUndRevision` als Fänger der Mutation „`erwartet` vor der Transaktion gelesen“; sie wird dort nicht rot (`rot-paket`: rot erst in `TestRevisionWettlauf` und `TestFensterZuordnung`). Die Tabelle *Mutationen* nennt `TestFensterZuordnung`, die Zeile W45 nicht. | Implementer/Planner: Fänger der Zeile W45 auf `TestRevisionWettlauf`/`TestFensterZuordnung` berichtigen oder die Mutation der Zeile auf „`erwartet` fest 0“ einengen. |
| V-4 | LOW | Plan nicht nachgezogen (`AGENTS.md` §3.9): *Nicht gemessen* (3) nennt 33 von 36 Zweigen ohne eigene Mutation; die Fix-Runde (laut Aufruf acht weitere, 5 von 5) steht nirgends im Plan (`grep` ohne Treffer), meine zwölf weiteren ergeben mit den dreien des Implementers 15 belegte Zweige. | Planner/Implementer: Zahl und Raten in Absatz *Mutationen und Raten* und *Nicht gemessen* (3) auf den Stand ziehen. |
| V-5 | LOW | W61: die Zeile sagt `t.Cleanup` des Fake-Speichers für jeden Test (Vergleich mit dem Stand beim Anlegen); umgesetzt ist nur der `Cleanup` für `SchreibeFehler` (`fakes_test.go`) und `unveraendert` an den Fehlerzeilen. *Nicht gemessen* (4) nennt es, die Zeile selbst sagt weiter mehr als der Fänger (`AGENTS.md` §3.11, §3.15). | Planner: Zeile W61 auf den tatsächlichen Fänger kürzen oder als Negativ mit Grund und Folge führen. |
| V-6 | INFO | `kuerzeOrt` liefert bei einem Ort aus mehr als 256 Folgebytes ohne Zeichenanfang (ungültiges UTF-8) den leeren Anfang; `Quote` schreibt je Byte vier Zeichen, die Grenze 1280 hält. `SPEC-180` Nr. 3 sagt dazu nichts. Nur gelesen, nicht gefahren; der Leser liefert gültiges UTF-8. | Architect: Randform im Vertrag nennen oder als Ungeprüft führen. |
| V-7 | INFO | DoD-Kästchen in §2 sind offen (Closure folgt); Review-Report liegt vor (`docs/reviews/2026-11-07-slice-mock-schreiben-review.md`); Handbuch: das Binary blieb unverändert, `ODR-E6009` steht als Ausnahme mit Grund (`slice-bootstrap-m3`). Beobachtung `BEO-REPO/…` und Closure-Notiz sind Sache der Closure. | Planner bei der Closure. |

## 5. DoD

| Punkt | Zustand |
|---|---|
| POST und DELETE (Key im Body, `201` mit `Location`, `409`, `404`/`ODR-E6003`, `204`) | erfüllt (Code, W16 bis W26, W51 bis W53, Ende-zu-Ende W52/W76/W78) |
| PATCH und PUT (nur übergebene, ersetzt, kein Upsert, `partial`, unbekannte Property `400`) | erfüllt (W23, W24, W31 bis W39) |
| Atomarität (Transaktion, Rollback, Revision nur bei Erfolg, andere Workspaces unberührt) | erfüllt (W59, W60, W69 bis W73, eigene Raten 50/50) |
| `make gates` grün | eigener Lauf grün |
| Review durchgeführt | Report liegt vor |
| Doku/Handbuch | kein Handbuch-Eintrag nötig, Binary unverändert, Ausnahme mit Grund |
| Closure, Register, Ausgänge in §6 | offen, Sache der Closure (Ausgänge: R1 bis R3 und A1 *offen* mit Adresse Planner, R4 entfallen) |

## 6. Rückgaben

- **An den Planner/Architect:** V-4, V-5 (Plan), V-6 (Vertrag), Closure-Pflichten nach V-7.
- **An den Implementer (vor der Closure, keine neue Prüfrunde nötig):** V-1, V-2, V-3, jeweils mit dem Bericht „Menge · Elemente · Mutation an nicht genanntem Element rot gesehen“.
- ADR-Kennungen in diesem Report stehen als auflösender Link ([ADR-0001](../plan/adr/0001-implementierungssprache-go.md)).
