# Verifikations-Report: slice-v2-payload-lesen — 2026-11-05

**Verifikations-Art:** DoD- und Entscheidungs-Konformität, Plan-vs-Code-Diff — geprüft gegen `docs/plan/planning/in-progress/slice-v2-payload-lesen.md` (DoD §2, Tabelle *Zusage · Fänger* mit 57 Zeilen, Raster §6, Block §4), `SPEC-172` Nr. 1 bis 5 (mit dem neuen Absatz zu Kreisen in Nr. 3), `SPEC-125` Nr. 9 a, `SPEC-126`, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) und `AGENTS.md` §3.10, §3.15, §3.17, §3.18.

**Gegenstand:** `slice-v2-payload-lesen` (Welle `welle-odata-v2`), Zweig `slice-v2-payload-lesen`, HEAD `6b9b332` (Review `ead28d1`, Handoff `617585f`, Architect `6b9b332`), Basis `20b0d1a`. Bezug: `LH-FA-17`, `LH-RB-01`, ADR-0011.

**Rolle / Modell / Datum:** Verifier · claude-sonnet-5-5 · 2026-11-05. Frischer Kontext, einmalige Verifikation, kein Push. Die Session wurde einmal unterbrochen; dieser Report ist der Zwischenstand und nennt, was gemessen ist und was nicht.

**Methode.** Der Bericht des Implementers lag nicht vor; geprüft sind Belege. `make gates` selbst gefahren (Exit 0). Mutationen auf Wegwerfkopien von `src/` (Scratchpad `scratchpad/ver-lesen/`), Image `odata-recorder:test`, `--network none --memory=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`, je Mutante eine Ersetzung; Probe-Paket für die Kreis-Klasse.

---

## Ergebnis (Zwischenstand)

Kein HIGH. Die zwei MEDIUM des Reviews (Lint rot, Plan ohne Absatz) sind behoben und halten gegen die Klasse. Offen sind ein LOW (Kreis: Ringe ab Länge 4 und Ring mit dem Typ des Sets nur durch Probe gefangen, die Suite ist dort grün; akzeptiertes Negativ im Plan, Probe belegt Grenze) und zwei INFO (S1 unadressiert im Nehmer, Lücke der Raten C2 bis C4). **Nicht gemessen wegen Unterbrechung:** Raten von C2, C3, C4 (G20/G21), Mutanten `batch6` (Nr. 1, 2, 3, H6) und `batch7`/`batch8` (G14 `tiefe`/`kodierung` aufgeschoben, Ring mit Typ des Sets in der Suite). Diese stehen unten als *offen*; sie sind kein Befund, sondern ungeprüft.

## Belege (selbst gemessen)

| Beleg | Messung |
|---|---|
| `make gates` | Exit 0; darin `make lint` grün (keine Meldung des Profils), `make test`, `make test-race`, `a-check` 0 Befunde, `d-check` 0 Befunde, `zeichen` ok (1181 Dateien), `groessenmarke slice-v2-payload-lesen.md: ok, Messung 4085 von Marke 8100` |
| Lint-Klasse | `lint.sh` fährt das ganze Profil über alle Module; kein `//nolint` im Diff (`git diff 20b0d1a..HEAD`, Treffer nur Plan-Prosa); `.golangci.yml`, `.a-check.yml`, `Dockerfile` unberührt |
| Größenmarke | `git diff --numstat 20b0d1a..HEAD` über die Pfade des Blocks: **4085**. Der Review nannte 7323: er zählte die ganzen Verzeichnisse `jsoncodecv2` und `outboundtest`; darin liegen 3245 Zeilen aus `slice-v2-metadaten` (`metadata_codec_v2_*.go`, drei `v2_*.xml`), die mit dem Merge `9650388` hinter der Basis `20b0d1a` in den Zweig kamen. 7323 − 3245 = 4078 (Stand des Reviews), +7 = 4085. Ursache geklärt, kein Fehler des Code. Mit den drei V4-Dateien der Suite (97 Zeilen) wären es 4182, ebenfalls unter 8100 |
| J4 | `git diff --numstat --diff-filter=MDRT 20b0d1a..HEAD` über `jsoncodec`, `csdl`, `urlparser`, `model`, `ports/outbound`: genau `odata_codec_kodierung.go` 12/6, `odata_codec_strom.go` 37/32, `odata_codec_werte.go` 7/3 (Summe 56/41, wie im Plan) |
| V4-Suite unverändert | `go test -v -run TestODataCodecVertrag ./src/bootstrap/` gegen `git archive 20b0d1a` und gegen HEAD: alle 127 Zeilen gleich bis auf Zeilennummern in Logzeilen, Heap-Spitzen und Laufzeit; gleiche Unterfälle, gleiche Zählungen (`A3 M2: 9552 Läufe`, 22195 Bodys), alles PASS |
| Echtdateien | SHA-256 und Größe aller 12 Dateien unter `testdata/v2_echt_*` und `v4_fremd_*` gleich den Werten des Plans; kein Zeilenende am Ende |
| Rohbytes (§3.18) | gegengelesen (`LC_ALL=C grep -P`, `iconv`): die 24 geänderten Dateien (`jsoncodecv2/*`, `odata_codec_v2_*.go`, drei V4-Suite-Dateien, Plan, `spec/spezifikation.md`): kein BOM, kein NBSP, kein Steuerzeichen, kein CR, gültiges UTF-8; `\u`+4 Hex-Ziffern nur in bereits vorhandener Spezifikationsprosa (Zeilen 963, 999, 1006, 1409, 1423), nicht in Suite, Plan oder Slice-Code. Die Echtdateien tragen `\/Date(` als Rohbytes |
| Paketvariablen | keine `var` auf Paketebene in `jsoncodecv2`, kein `sync` |
| Tabelle | 57 Zeilen (3 + G1–G21 + H1–H10 + N1–N6 + W1–W11 + K1–K3 + R1, R2 + J4), davon 2 nebenläufig |

## Mutationen (selbst gefahren)

Basis grün: `TestVertrag` (V2) und `TestODataCodecVertrag` (V4). Rot heißt: der genannte Unterfall meldet `--- FAIL`.

| Klasse (Menge) | Mutanten | Ergebnis |
|---|---|---|
| G9 Grammatik, jenseits des Tabulators (Menge: Leerraum, Escapes, Steuerzeichen, Zahlform, Kommas, Literale, Hex, Wurzelrest, Doppelpunkt, Klammerart) | 21: LF, CR, Leerzeichen aus Leerraum; Formfeed zugelassen; `\v` zugelassen; `0x1F` im String; `+` vor Zahl; führende Null; `.` nach Exponent; `E` fehlt; Vorzeichen `+` im Exponenten; `1e` mit Bruch; Komma vor `}` und `]`; Hex `g`; drei Hex-Ziffern; Literal-Zeichen; Literal-Präfix; Müll hinter der Wurzel; `=` als Doppelpunkt; gemischte Klammern | **21 von 21 rot in G9** |
| G14 je Paar von Z2 | Dublette erst am Ende gemeldet; `ende()` ohne Syntaxprüfung; Müll hinter der Wurzel als `kodierung`; Fehler von `src` vor den Daten des Reads | **4 von 4 rot in G14** |
| G14, nicht erreicht | `tiefe_dup` (Dublettenprüfung nur unter Tiefe 4) und `kod_spaeter` (Fortsetzungsbyte-Bereich) | grün in G14, **weil die Mutation das Paar nicht berührt** (die Dublette steht in Tiefe 3, das Paar `kodierung`/`tiefe` nutzt das Führungsbyte `0xFF`, nicht die Fortsetzung); keine Aussage über Fänger. Geplant, nicht gefahren: `batch7` (`tiefe` und `kodierung` je aufgeschoben, Dublette ab Tiefe 3) |
| H8, H9, K1, R2 (Plan: „nicht gefahren“) | `__count`/`__next` an der Entity ignoriert; `results` an der Entity ignoriert; Schlüssel in Body- statt Modellreihenfolge; Fehlerobjekt als leere Collection; Fehlerobjekt als leere Entity | **5 von 5 rot im eigenen Unterfall** (H8, H9, K1, R2, R2) |
| Geteilte Helfer, jede Aufrufstelle V4 und V2 (§3.17: Menge = alle Stellen × alle Elemente der Klasse) | UTF-8-Bereiche an allen 7 Grenzen (`C0`, `E0`, `ED`, `E1..EF`, `F0`, `F4`, `F1..F3`) für `uPruefer`/`b3Alphabet`; für `a3Quellfehler`: Fehler nach vollständigem Dokument verschluckt, `errors.Is(io.EOF)` statt `==`, Daten im Read mit Fehler verworfen, Fehler mitten in einer Zahl verschluckt; für `int32Werte`: `ParseInt(…, bits-1)`, `-0` als `-0`, Int32-Grenze 32 zu 33 | **28 von 28 rot**, je Mutation sowohl in der V4-Suite (B3, A3, F3) als auch in der V2-Suite (G10, G6, W3); das heißt alle sechs Aufrufstellen für jedes Element, darunter 11 Elemente, die der Befund nicht nannte |
| Kreisschutz, Probe über die Klasse (Ringe der Länge 1 bis 12 × Typ des Sets außerhalb/im Ring × Anhänge 0 bis 4 × beide Methoden; Orakel: Kette erreicht den Typ des Sets) | Probe gegen den Code | **grün, 3456 Läufe** |
| Kreisschutz, Mutanten | Markierung `besucht` entfernt; nur am ersten Typ gesetzt; Ring ab Länge 4 nicht abgebrochen (Markierung wird bei 4 Einträgen geleert) | Probe: 3 von 3 rot (der Leser kehrt nicht zurück). Suite (`N1`): Markierung entfernt rot, nur am ersten Typ rot, **Ring ab Länge 4 grün** |

## Raten G20/G21 (Image `odata-recorder:race`, 30 Prozesse je Zeile, `-test.count=1`)

| Mutante | G20 (Prozesse rot) | G21 (Prozesse rot) |
|---|---|---|
| Basis (ohne Mutation) | 0 von 30 | 0 von 30; `-count=20` ohne Detektor: 0 von 20 und 0 von 20 |
| C1 Paketvariable `zaehlerM` | **30 von 30** (letzter Lauf 7 Meldungen) | **30 von 30** (79 Meldungen) |
| C2 Lesepuffer als Paketvariable | offen (unterbrochen) | offen |
| C3 `letzter` mit dem Fehler von `lies` | offen | offen |
| C4 Vermischung unter Sperre, Wirkungs-Fänger | offen | offen |

## DoD, Zeile für Zeile

| DoD-Punkt | Befund |
|---|---|
| Lesen (Hülle `d`, Typen, `__count`, `__next`, übersprungen; Modell anderer Version `ErrUngueltig`; V4 unberührt; `Edm.DateTimeOffset(V2)` als `typ`; `versionstrennung_test.go`) | erfüllt: V2-Suite und V4-Suite grün, V4-Verhalten gleich (oben), `versionstrennung_test.go` kommt per Merge und läuft in `make test` |
| Randformen des Lesers fremder Bytes | erfüllt, soweit gemessen: Raster ohne leere Zelle, Spezifikation führt die Orte; Mutationen oben rot, Prüfgrenze des Kreises als akzeptiertes Negativ |
| `make gates` grün | erfüllt, selbst gefahren |
| Review | Report liegt vor (`2026-11-05-slice-v2-payload-lesen-review.md`) |
| Doku-Update | Architect: `SPEC-172` Nr. 3 Absatz zu Kreisen und Historie vom selben Tag; Binary unverändert, `jsoncodecv2` ist nirgends verdrahtet (Handbuch bleibt, §7 muss es bei der Closure nennen) |
| Closure-Notiz, Register, Ausgänge der Risiken, Paarungen | **Sache der Closure**: §7 lautet „Wird bei der Closure gefüllt“; §6 trägt zwei Ausgänge *offen* mit Adresse (S1, Register) |

## Plan-vs-Code-Diff

| Zusage des Plans | Code | Urteil |
|---|---|---|
| Paket `jsoncodecv2`, eigener Scanner, kein V4-Import, kein Feld im Leser | `type Leser struct{}`, 10 Dateien (1389 Zeilen), `scan.go` und `zahl.go` sind Kopien der V4-Dateien bis auf den Paketnamen; kein Paketzustand | gleich |
| `make lint` ohne Ausnahme grün durch Umbau; drei geteilte Helfer, `b3Alphabet` ohne `bis` | Diff zeigt `uPruefer`, `a3Quellfehler`, `int32Werte`; `lint` im Gate grün | gleich |
| J4: nur drei V4-Dateien der Suite geändert | numstat gleich 56/41; V4-Verhalten gleich | gleich; Entscheidung des Architect (ADR-0011 §Konsequenzen) trägt |
| Kreis: der Leser kehrt zurück, `odata_typ` | `abgeleitet` markiert jeden Typ, einzige Stelle im Leser, die `BaseType` folgt (`grep`) | gleich `SPEC-172` Nr. 3 |
| Mutationsabsatz und Raten im Plan (§3.10) | Absatz und Tabelle stehen im Plan; C1 bis C3 mit 30 von 30 genannt; C1 von mir bestätigt | teilweise bestätigt (Rest offen) |
| Plan-Zeile *Nr. 3* der Tabelle (Import von `jsoncodec`) | in der Mutationstabelle des Plans fehlt die Zeile *Nr. 3* | kleine Lücke, siehe LOW-2 |

## Findings

| Nr. | Stufe | Befund |
|---|---|---|
| V-1 | LOW | **Kreis, Klasse nicht von der Suite gedeckt.** `SPEC-172` Nr. 3 sagt zu: „Länge des Rings ohne Wirkung, auch ab 4“ und „erreicht die Kette den Typ des Sets, ist der Typ abgeleitet“. Die Suite führt Ringe der Länge 1 bis 3 und kein Modell mit dem Typ des Sets im Ring; die Mutante *Ring ab Länge 4 nicht abgebrochen* bleibt in `N1` grün, meine Probe über 3456 Läufe fängt sie. Der Plan führt das als akzeptiertes Negativ mit Grund und Folge; es ist zulässig, aber billig zu schließen (zwei Ringlängen und ein Ring mit dem Set-Typ im Modell `Y`). Der Code ist gegen die Klasse korrekt (Probe grün). Adresse: Planner. |
| V-2 | LOW | Die Mutationstabelle des Plans hat keine Zeile *Nr. 3* (Gegenprobe: `jsoncodecv2` importiert `jsoncodec`), obwohl die Tabelle in §3 sie nennt; der Fänger `versionstrennung_test.go` ist gesetzt, die Mutation von mir nicht mehr gefahren (siehe offen). |
| V-3 | INFO | **S1 an `slice-v2-bootstrap`** (Schlüsselliteral) steht nicht in dessen §1 oder DoD auf `main` (`grep` auf `Schlüsselliteral` und `SPEC-172 Nr. 5`: kein Treffer). Der Plan führt sie mit Ausgang *offen*, Adresse Planner, was `AGENTS.md` §3.13 für eine Rückgabe erlaubt; vor der Closure zu tragen. |
| V-4 | INFO | In `v2G6` trägt der Aufruf von `a3Quellfehler` das Label „mitten in einer Zahl“, schneidet aber vor der Ziffer (`"ID":1` mit Versatz 5); die Klasse deckt `v2G6Zeitpunkte` über jede Position (Mutante *Fehler mitten in einer Zahl verschluckt* rot), die Stelle selbst ist nur ein Label. |
| V-5 | INFO | Die Abweichung der Messung (4085 gegen 7323) ist geklärt (siehe Belege); der Plan §4 nennt die Ursache noch „nicht geklärt“. Planner: Zeile nachziehen. |

## Offen (nicht gemessen)

- Raten C2, C3, C4 für G20 und G21 (Skript `scratchpad/ver-lesen/rate.sh`, Mutanten in `batch5.sh`); C4 (Vermischung unter Sperre) damit weder bestätigt noch widerlegt.
- Mutanten `batch6.sh` (Nr. 1, Nr. 2, Nr. 3, H6) und `batch7.sh` (G14 `tiefe`/`kodierung` aufgeschoben, Dublette in Tiefe 3), `batch8.sh` (Ring mit dem Typ des Sets, Zyklus-Vorprüfung).
