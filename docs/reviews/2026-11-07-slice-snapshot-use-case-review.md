# Review-Bericht: slice-snapshot-use-case — 2026-11-07

**Rolle:** Reviewer (Modul 10) · **Frage:** Ist der Diff gegen Plan, Entscheidungen und Hard Rules wartbar und richtig? — nicht gegen die DoD (Verifier).

**Review-Art:** Code (gegen Plan und Konventionen).

**Gegenstand:** `slice-snapshot-use-case`, Zweig `slice-snapshot-use-case`, HEAD `833824cd`, Diff `ee8e6b1a..833824cd` (30 Dateien, 5490 Einfügungen, 40 Löschungen). Messung der Marke nach `SPEC-128` (Basis `6e1e52d3`, Pfade des Blocks `groessenmarke`): 5409 Zeilen bei Marke 10600; `src/bootstrap/workspace_sqlite_test.go` (8 plus 7 Zeilen) liegt außerhalb der Pfade.

**Geprüft gegen:** Plan `slice-snapshot-use-case` (§3 Tabelle T01 bis T81, Konkordanz, Mengen M1 bis M17, Z-Liste, Absatz *Mutationen und Raten*, §6 Raster und Rückgaben R1 bis R4), Vertrag `SPEC-129` Nr. 16 bis 26, `SPEC-109` Nr. 5, `AGENTS.md` §3.7, §3.9 bis §3.18. Skill: `.harness/skills/reviewer.md`. Modell: Sonnet 5.5.

**Methode.** Einmal gefahren, keine Wiederholungsschleife. Der Bericht des Implementers ist nicht übernommen; seine Zahlen habe ich nur dort wiedergegeben, wo ich sie selbst gemessen habe. Alle Läufe über `docker run --network none --memory 6g` unter `flock /tmp/odata-recorder-heavy.lock`, auf Baumkopien (`git archive`), nie im Worktree. Die Last des Hosts lag zwischen 45 und 80 (andere Läufe auf demselben Host, der Lauf stand lange in der Schlange der Sperre). Gefahren:

- 75 Läufe eigener Mutationen an `src/hexagon/services/snapshot/` und `src/hexagon/services/workspace/` (Pakettests, je Mutation ein Lauf von `go test`), dazu je ein Lauf der Bootstrap-Tests (`-race`, ein Prozess) für zwei Mutationen am `Restore`. Davon 62 rot, 10 grün (F-1, F-5, F-7, F-9), 3 `build failed` durch ungenutzten Import oder ungenutzte Variable (nicht gezählt; die Varianten M01, V05b und V10b sind nachgefahren und rot). Die Mutationen sind eigene, nicht die Spalte *Mutation* der Plan-Tabelle; sie decken die Klassen: Grenze und Byte-Mengen des Namens, `ctx`-Identität je Aufrufstelle (alle 8 Stellen von `Create` und `List`, alle 8 Stellen von `Restore` einschließlich Transaktion, `Get`, `ReplaceAll`, `Advance`), Gegenstand je Stelle (N1 bis N4, N10, N11, S19, S20, S4, S5, S9, S11), Kennung und Zeit vor dem Blob, Reihenfolge der Liste, Folge und Schlüssel im Manifest, Escape-Zeichen (Hexziffern, `"`, `\`), Doppelte nach Schlüssel allein und nach Set allein, `Abort` (fehlt, doppelt), Strom (`Close` doppelt), Kontextfehler neben `ErrFalscheArt`, Revision (veraltet, nicht übernommen, zweite Transaktion), Wert-Lecks in Verstoß-Texten.
- Abdeckung unter dem Detektor (`go test -race -covermode=atomic -coverprofile`, beide Pakete, einmal): 100,0 % beider Pakete, 0 von 163 Blöcken ohne Treffer in `services/snapshot/` (`dienst.go` 41, `fehler.go` 5, `manifest.go` 20), `wiederherstellen.go` (37) und `manifest_snapshot.go` (60) — wie berichtet.
- Gegenlesung auf Rohbytes (`LC_ALL=C grep -P` auf `\\u[0-9a-fA-F]{4}`, Steuerzeichen, NBSP, BOM) über alle neuen Testdateien der drei Pakete: kein Treffer.
- Katalog und Gates: `make gates`, `make lint` und `make a-check` habe ich nicht gefahren (Aufgabe des Implementers beim Handoff und des Verifiers). `make docs-check` über diesen Bericht: siehe unten.

**Grenze dieses Laufs (ehrlich).** Die Raten T60 bis T71 habe ich nicht mit 50 oder 10 Prozessen nachgemessen. Ich habe je Fenster eine Mutation in einem Prozess rot gesehen; das belegt die Fänger, nicht die Raten. Von den nebenläufigen Zeilen sind T60 (Revision veraltet, Haltepunkt am Fake) und T61 (zweite Transaktion) am Fake und T62, T63, T64 (Revision aus dem äußeren Workspace und `Get` der Transaktion entfernt) gegen den echten SQLite-Adapter in je einem Prozess rot gesehen; T66, T68, T70, T71 nicht selbst mutiert, T65 siehe F-8. Den Flake `TestTransaktionVertrag` des SQLite-Adapters habe ich nicht gesehen (die Bootstrap-Basis lief grün, 55,7 s).

---

## Findings

### F-1 (MEDIUM): Zusage „Ort ohne Wert“ des Manifest-Lesers hat Fänger nur für Schlüssel und Payload

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe; Zusage ohne Fänger); `SPEC-129` Nr. 20 *Lesen* („die Ursache nennt den Ort und keinen Wert“); Plan T45 (Fänger: „ein Canary im Wert eines Gliedes steht weder in `Error()` noch in der Ursache“)
- `pfad`: `src/hexagon/services/workspace/manifest_snapshot.go` · `verstoss("%s passt nicht zum Snapshot", name)`, `verstoss("%s.completeness ist unbekannt", ort)`, `verstoss("%s.version ist keine Ganzzahl im Bereich von int64", ort)`, `verstoss("manifest_version ist nicht die Ganzzahl 1")`, `verstoss("entities[%d] verletzt die Invarianten des Eintrags", i)`; Fänger `wiederherstellen_manifest_test.go` · `kanarienWert`
- `befund`: Fünf Mutationen, die den Wert des Gliedes in den Text des Verstoßes schreiben (V06 `service_id` und `metadata_hash`, Z01 `completeness`, Z02 `version`, Z03 `manifest_version`, Z05 den Fehler von `Validate()` samt `entity_set` und `typed_key`), bleiben in allen Tests des Pakets grün. Rot wird nur die Mutation am Text der Doppelten (V07, Schlüssel); der Canary steht in `kanarienEintraege` nur in Schlüsseln und Payloads, nicht im Kopf und nicht in `completeness` und `version`; der Fall der Invarianten lässt ihn offenbar nicht in den Text des Fehlers von `Validate()` gelangen (Z05 grün).
- `verifizierbar`: ja (Mutationen V06, Z01, Z02, Z03, Z05, je ein Lauf von `go test ./src/hexagon/services/workspace/`)
- `klasse`: Fänger führt nur eine Form der Eingabe (Allaussage „kein Wert“ ohne Schleife über alle Glieder)

### F-2 (MEDIUM): Zählprüfung der Z-Liste von `Restore` gegen den Quelltext fehlt

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Zusage und Fänger als Paar), §3.10 (Zweige unter dem Detektor); Plan Auftrag D und T72 („Z-Liste, Zählprüfung gegen den Quelltext per `go/parser`“)
- `pfad`: `src/hexagon/services/workspace/wiederherstellen_fenster_test.go` · `t.Fatalf("Zählprüfung: %d Zweige", len(zs))`; Gegenstück `src/hexagon/services/snapshot/zweige_test.go` · `fehlerRueckgaben(t)`
- `befund`: Der Dienst `snapshot` zählt die Fehlerrückgaben von `dienst.go` per `go/parser` (13). Für `wiederherstellen.go` und `manifest_snapshot.go` zählt `TestZweigeRestore` nur die Länge der eigenen Liste (17); ein neuer Zweig mit Fehlerrückgabe im Quelltext lässt kein Gate rot werden. Die 100 % Abdeckung bleiben ein einmaliger Lauf von Hand (ich habe ihn wiederholt und bestätigt), kein Gate und kein Test.
- `verifizierbar`: ja (Lesen von `TestZweigeRestore`; ich habe die zusätzliche `return`-Zeile nicht gefahren, der Test zählt keine Quelle)
- `klasse`: Zählprüfung gegen die Quelle nur für einen der beiden Dienste

### F-3 (MEDIUM): Satz zu `ODR-E5004` in `LH-FA-10.a` und `SPEC-129` Nr. 18 widersprechen sich, der Ausgang ist offen

- `kategorie`: MEDIUM
- `quelle`: Maintainability / Spec-Straten (gleiches Stratum); `LH-FA-10.a`; `SPEC-129` Nr. 18 *Entscheidung*; Plan §6 Rückgabe R3 und T78
- `pfad`: `spec/spezifikation.md` · `ein Snapshot eines anderen Service ergibt `ODR-E5004`` (Abschnitt `LH-FA-10.a — Snapshots`); `spec/spezifikation.md` · `der Workspace-Snapshot eines Workspace eines anderen Service ist `404`/`ODR-E7002``
- `befund`: Der Abschnitt `LH-FA-10.a` nennt den Fall ohne Einschränkung, Nr. 18 liefert für den Workspace-Snapshot eines Workspace eines anderen Service `ODR-E7002` statt `ODR-E5004`, und der Code folgt Nr. 18. Der Abschnitt `LH-FA-10.a` trägt den Vorbehalt nicht; Rückgabe R3 steht mit Ausgang *offen* beim Planner (Meldung an den User), das Lastenheft nennt den Code nicht. Ich stufe nicht höher, weil der Widerspruch benannt, entschieden und mit Fänger (T35, T37) und Negativ (T78) geführt ist; bleibt R3 offen, ist der Satz in `LH-FA-10.a` für einen Leser des Abschnitts allein falsch.
- `verifizierbar`: nein (kein Gate liest den Widerspruch zwischen zwei Sätzen derselben Datei)
- `klasse`: Satz einer Nummer in einem Abschnitt geändert, Parallelstelle nicht nachgezogen (`AGENTS.md` §3.9)

### F-4 (LOW): Der bestehende Test zählt eine übersprungene Zelle als gelaufen

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.11 (ein Kommentar sagt nur zu, was ein Test prüft), Plan T77 („keine wird abgeschwächt“)
- `pfad`: `src/hexagon/services/workspace/kontext_test.go` · `if s.nr == "S19" && fe.err == outbound.ErrFalscheArt {` mit `faelle++` und `continue`
- `befund`: Die Zelle S19 × `ErrFalscheArt` ist im Test der Schleife übersprungen, zählt aber in `faelle` (`28*28`); geprüft wird sie in `TestRestorePortFehler`. Die Zählprüfung des bestehenden Tests nennt damit mehr Fälle, als er prüft. Eine Abschwächung sehe ich nicht (alle anderen Änderungen an bestehenden Tests heben Zahlen an der Stelle, die der Vertrag ändert), die Zahl ist aber nicht mehr die der geprüften Zellen.
- `verifizierbar`: ja (Lesen der Stelle)
- `klasse`: Zähler zählt Zelle ohne Prüfung

### F-5 (LOW): Die Kopie der Eingabe im Manifest-Schreiber hat keinen Fänger

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15; `SPEC-129` Nr. 20 („der Dienst stellt die Folge her und glaubt der Reihenfolge des Ports nicht“)
- `pfad`: `src/hexagon/services/snapshot/manifest.go` · `sortiert := slices.Clone(es)`, `props := slices.Clone(e.KnownProperties)`
- `befund`: Die Mutationen „ohne Kopie“ (N12 für die Liste, Y01 für die Properties) bleiben grün: der Dienst sortiert dann die Liste und die Property-Listen, die der Port geliefert hat, an Ort und Stelle. Kein Test hält die Eingabe fest und vergleicht sie nach `Create`; der echte Adapter liefert frische Listen, ein anderer Port könnte Zustand teilen.
- `verifizierbar`: ja (Mutationen N12, Y01)
- `klasse`: Seiteneffekt auf Argument des Ports nicht gefangen

### F-6 (INFO): Plan-Aussage zu T61 trifft für die Mutation „zweite Transaktion“ nicht zu

- `kategorie`: INFO
- `quelle`: Plan *Mutationen und Raten* Zeile T61
- `pfad`: Plan · `„Advance in einer zweiten Transaktion“ selbst nicht gesehen, weil der Fake eine verschachtelte Transaktion einreiht und hängt`
- `befund`: Die Mutation `Advance` in einer zweiten, auf die erste folgenden Transaktion (W20) ist rot mit `--- FAIL` in `TestRestoreFensterTransaktion`, `TestRestoreAblauf`, `TestRestoreAtomar`, `TestStellenRestore` und weiteren, ohne Zeitlimit. Nur die verschachtelte Form hängt; die ist ein anderer Fall. Der Plan darf sie nennen; die Aussage „nicht gesehen“ stimmt für die zweite Transaktion nicht. Keine Aktion außer Nachziehen des Satzes.
- `verifizierbar`: ja (Mutation W20)
- `klasse`: Plan nennt Negativ, das ein Fänger deckt

### F-7 (INFO): Prüfgrenze „Strom bis zum Ende“ liegt bei 64 KiB nicht

- `kategorie`: INFO
- `quelle`: Plan T79 (Größe des Manifests ungeprüft, akzeptiertes Negativ mit Grund und Folge)
- `pfad`: `src/hexagon/services/workspace/wiederherstellen.go` · `io.ReadAll(rc)`
- `befund`: Die Mutation `io.ReadAll(io.LimitReader(rc, 1<<16))` (W11) bleibt grün: kein Test liest ein Manifest über 64 KiB. Das ist T79 (Mengen bis 48 Einträge geprüft); ich führe es als bestätigt, nicht als neuen Befund. Das Negativ nennt Grund und Folge.
- `verifizierbar`: ja
- `klasse`: akzeptiertes Negativ bestätigt

### F-8 (INFO): Rückgabe R4 — meine Auflösung

- `kategorie`: INFO
- `quelle`: Plan §6 Rückgabe R4
- `pfad`: Plan · `Rückgabe R4 an den Architect: Zählfehler der Tabelle und ein Negativ, das der Test nicht fängt`
- `befund`: (a) 8 + 31 = 39 bestätigt: `TestRueckgabenSnapshot` teilt `model.AlleCodes()` in zwei Listen des Tests ohne feste Zahl; die Zeile `ODR-E7007` folgt aus `outbound.Abbilden` (`ErrBereitsVorhanden` mit dem Gegenstand `Workspace`), die Mutation N2 mit dem Gegenstand `Sonstiges` (M08) wird rot. (b) 16 Verstöße der Zeichen vor und hinter dem Manifest in `TestManifestSyntax` gezählt; die Tabelle ist nachgezogen. (c) Die Lücke in `SPEC-129` Nr. 24 (kein Satz zu `ODR-E7007`) bleibt eine Entscheidung des Architect; der Code folgt dem letzten Satz von Nr. 24. (d) T65: Ich fahre den Mutanten selbst, den der Plan nicht findet: `Get` der Transaktion entfernt (W21, Revision und Zeile aus dem äußeren Workspace). `TestRestoreGegenLoeschen` bleibt grün; rot werden die Tests am Fake (Zählung der Operationen, Kontext, Fehlerabbildung) und die Bootstrap-Tests T62 bis T64. T65 selbst hat also keinen Mutanten am Dienst; die Folge „nie ein State ohne Workspace“ trägt der Adapter, und das Negativ ist mit Grund (Adapter) und Folge benannt. Das ist ein tragfähiges akzeptiertes Negativ, ich zähle es nicht als Zusage ohne Fänger. (e) `src/bootstrap/workspace_sqlite_test.go` (Feld `waechter`, additiv, nichts abgeschwächt) liegt außerhalb der Pfade des Blocks, der Plan nennt ihn in §3 und in R4 e; die Messung der Marke ändert das um 15 Zeilen von 5409.
- `verifizierbar`: ja
- `klasse`: Rückgabe aufgelöst

### F-9 (INFO): Mutationen, die ich als äquivalent einstufe

- `kategorie`: INFO
- `quelle`: Maintainability
- `pfad`: `src/hexagon/services/workspace/manifest_snapshot.go` · `if len(m["payload"]) == 0 {`, `if !utf8.Valid(b) || !json.Valid(b)`
- `befund`: Grün bleiben: `payload` fehlt nicht geprüft (W18; `Validate()` lehnt den leeren Payload ab), `json.Valid` entfernt (W19; `Unmarshal` lehnt dieselben Eingaben ab). Beide sind redundante Prüfungen (Vertrag Nr. 20 c und e); die Absicht steht im Plan (Doppelung). Kein Befund.
- `verifizierbar`: ja
- `klasse`: äquivalente Mutation

---

## Zeilen der Tabelle, die ich mit eigenen Mutationen aufgelöst habe

| Zeilen | Eigene Mutation (Klasse über die genannten hinaus) | Ergebnis |
|---|---|---|
| T03 | Länge `> 128` auf `> 129`; `0x20` als verbotenes Byte (`<= 0x20`); UTF-8-Prüfung ersetzt | rot (`TestNameForm`) |
| T07 bis T12, T46 | Hexziffern groß; `"` ohne `\`; `version` auf `int32` gekürzt; Properties und Sets nicht sortiert | rot |
| T15, T17, T23 | `Abort` fehlt nach `Write`, fehlt nach `Commit`, doppelt; `Begin` mit `MaxInt32`; Kennung und Uhr vor dem Blob | rot |
| T21, T49 | Gegenstand je Stelle umgesetzt an allen Stellen, die das Raster nennt (N1, N2, N3, N4, N10, N11, S19, S20, S4, S5, S9) | rot an jeder Stelle |
| T22, T50 | abgeleitetes `ctx` an allen acht Stellen von `Create` und `List`, an allen Stellen von `Restore` einschließlich der Transaktion, ihrer drei Schritte und `Get` | rot an jeder Stelle (Identität per `==`) |
| T26 | Liste umgedreht | rot (`TestListe`) |
| T34 bis T37 | `ServiceID` verglichen mit der Kennung des Snapshots; Ursache an `ODR-E5004` gehängt; `ErrFalscheArt` verloren | rot |
| T38, T39 | `StateHash` gegen leer statt gegen `StateBlob` geprüft | rot |
| T40 | `Close` doppelt | rot |
| T42 bis T45 | `completeness` nur `complete` angenommen; `deleted: null` angenommen; Doppelte nur nach Schlüssel oder nur nach Set; UTF-8-Prüfung ersetzt | rot (Ausnahme: Wert-Lecks, F-1) |
| T32, T60 | Revision nicht aus `Advance` übernommen; Revision aus dem äußeren Workspace | rot (`TestRestoreRevision`, `TestRestoreFensterLesen`) |
| T31 | letzter Eintrag vor `ReplaceAll` abgeschnitten | rot |
| T61 | zweite, nachfolgende Transaktion für `Advance` (F-6) | rot |
| T62 bis T64 | `Get` der Transaktion entfernt (W21); Revision des äußeren Workspace (W22) gegen den echten Adapter, je ein Prozess mit `-race` | rot (`TestRestoreGegenMutation`, `TestRestoreGegenReset`, `TestRestoreGegenRestore`) |
| T65 | siehe F-8 d | grün, Negativ tragfähig |
| T58 | nicht mutiert | Negativ nennt Grund und Folge |
| T72 | Abdeckung beider Pakete unter dem Detektor wiederholt | 0 von 163 Blöcken ohne Treffer |

Nicht gefahren: T67, T69 (Fenster von `Create`) und die Wettläufe T66, T68, T70, T71 mit eigenen Mutationen; für T67 und T69 liegen die Haltepunkte am Argument des Fakes im Test, die Fänger lesen sich wie berichtet.

## Prüfung der Hard Rules am Diff

- §3.7, §3.11 (Kommentare): geprüft an den neuen Quelldateien (`dienst.go`, `fehler.go`, `manifest.go`, `manifest_snapshot.go`, `wiederherstellen.go`): keine verworfene Alternative, kein abwesender Text, keine Chronik. In den Tests nennen Kommentare Plan-Zeilen (`T26`), das ist ein auflösbarer Zeiger.
- §3.9: Plan §1, §3, §6 und der Kopf tragen die Rückgaben R1 bis R4 nach; Ausgang *offen* mit Adresse bei R3 und R4.
- §3.15 und §3.17: Der Fänger führt Mengen als Schleife mit Zählprüfung (M1 40, M2 48, M3 14, M7 128 und 144, M8 32 und 36); die Mutationen an Elementen, die der Plan nicht nannte, sind rot, außer in F-1 und F-5.
- §3.18: Rohbytes gegengelesen, kein Fund.
- Gate-Suppression (`//nolint`): keine neue Direktive im Diff (gesucht).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/inbound/` (`snapshot.go`, `workspace.go`) | geprüft, ohne Befund |
| `src/hexagon/services/snapshot/dienst.go`, `fehler.go`, `manifest.go` | geprüft, Befund F-5 (Kopien der Eingabe); Texte der Verstöße siehe F-1 |
| `src/hexagon/services/workspace/wiederherstellen.go` | geprüft, ohne Befund am Produktcode |
| `src/hexagon/services/workspace/manifest_snapshot.go` | geprüft, Befund F-1 |
| `src/hexagon/services/workspace/` bestehende Tests (T77) | geprüft, Befund F-4 (LOW), keine Abschwächung |
| `src/bootstrap/snapshot_sqlite_test.go`, `workspace_sqlite_test.go` | geprüft, ohne Befund (Feld `waechter` additiv) |
| `spec/spezifikation.md` Nr. 16 bis 26 | geprüft, Befund F-3 |
| Rückgabe R4 und Negative T58, T61, T65, T78 bis T81 | geprüft, F-6 bis F-8; kein Negativ ohne Grund und Folge |

## Summary

Keine HIGH-Findings, drei MEDIUM (F-1 Wert-Lecks in Verstoß-Texten ohne Fänger, F-2 Zählprüfung der Z-Liste von `Restore`, F-3 Satz zu `ODR-E5004` in zwei Nummern), zwei LOW, vier INFO. Die ctx-Identität und die Fehlerabbildung sind an jeder Aufrufstelle gefangen, die Fenster der Sperre T60 und T61 am Fake in einem Prozess rot, die Wettläufe gegen den echten Adapter rot bei den zwei Mutationen, die ich gefahren habe. Die Raten des Implementers (50 von 50, 10 von 10) habe ich nicht nachgemessen.
