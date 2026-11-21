# Verifikationsbericht: slice-listener-tls (Teil 1: Baustein `ladeTLS` und Hilfspaket `tlstest`) — 2026-11-21

**Rolle:** Verifier (Frage: bauen wir es richtig, gegen Plan und DoD) · **Gegenstand:** Zweig `slice-listener-tls`, HEAD `3e7ce69b` (Messbasis der Marke `de7589e5`, Basis des Zweigs `main` `6add6959`); Commits `314b2343` (Handoff), `a97f9f00` (Klärungen des Architect), `a9263970` (Review-Report), `3e7ce69b` (Fix-Runde F-1 bis F-7, ohne zweites Review) · **Datum:** 2026-11-21

**Eingangs-Kontext:** Plan `slice-listener-tls` (§2 DoD, §3 Tabelle A01 bis A17, Mengen, Absatz *Mutationen, Raten und Messende*, *Klassen der Fix-Runde*, akzeptierte Negative, §4 Block der Größenmarke, §6 Risiken und Raster); `SPEC-291` (Nr. 1 bis 13, *Ungeprüft*); `SPEC-110` Nr. 2, 7, 11, 12; `SPEC-136`, `SPEC-139`, `SPEC-055`, `SPEC-088` (Teil 2 trägt die Verdrahtung, hier nur auf Widerspruch gelesen); `SPEC-128` Nr. 10 g; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-RB-02`, `LH-QA-04` (Bezug); `AGENTS.md` §3.7, §3.9 bis §3.18; Review `2026-11-21-slice-listener-tls-review` (0 HIGH, 2 MEDIUM, 4 LOW, 3 INFO).

Alle Mutationen liefen in Wegwerfkopien im Container (Worktree schreibgeschützt eingehängt, Kopie nach `/tmp/c`), nie im Zweig. `make gegenproben` wurde nicht gefahren, der Stempel nicht angefasst (bekannt rot, liegt bei `slice-gegenproben-eichung`). Kein Code, kein Plan geändert, kein Push, kein Merge; geschrieben ist allein dieser Report.

## Ergebnis

**Freigegeben**, mit Auflagen an den Planner für die Closure (V-1 bis V-3, alle LOW: je ein akzeptiertes Negativ nachtragen oder den Fänger schärfen; V-6, V-7: Plan-Verwaltung). Kein HIGH, kein MEDIUM. Die DoD ist, soweit sie vor der Closure prüfbar ist, belegt; `make gates` ist grün (die Stufe `race` frisch gefahren); der Baustein folgt `SPEC-291` Nr. 1 bis 6 und 10 bis 12 Zeile für Zeile; die fünf Klassen der Fix-Runde (F-1 bis F-5) sind gegen die Klasse aufgelöst und halten bis auf die in V-1 bis V-3 genannten Ränder. Findings: 0 HIGH, 0 MEDIUM, 3 LOW, 5 INFO.

## Sensor-Belege (selbst gefahren, nicht aus der Behauptung übernommen)

Last des Hosts: Load Average (1 Minute) laut `/proc/loadavg` im Container oder davor; eine zweite Linie (Closure `slice-telemetry-einbau-mock`) lief parallel, daher 2 bis 20 von 4 CPUs (Angaben je Zeile).

| Sensor | Ergebnis |
|---|---|
| `make gates` (Worktree, HEAD `3e7ce69b`, sauberer Baum) | Exit 0, 56 s. Die Docker-Schichten von `test`, `build`, `lint` und `race` kamen bei gleicher Eingabe aus dem Cache (nicht frisch gefahren); deshalb die Stufe `race` frisch, nächste Zeile. `a-check` 0 Befunde, `docs-check` 989 Dateien 0 Befunde, `groessenmarke` `ok, Messung 1511 von Marke 2000`, `migration-check` ok, `test-contract`, `test-determinism` (100 Läufe, 0 abweichend), `test-integration` grün, `zeichen-check` ok (2013 Dateien) |
| Stufe `race` frisch (`docker build --no-cache-filter race`, gleiche Argumente wie das Ziel `test-race`, anderer Tag, Image danach entfernt) | `pruefung race: ok`, 316 s; Last 12,8 bis 19,6 von 4 CPUs (vor und nach dem Lauf) |
| Messung `git diff --numstat de7589e5..HEAD` über die drei Pfade des Blocks | 113 (`tlsdatei.go`) + 1095 (`tlsdatei_test.go`) + 259 (`tlstest.go`) + 44 (`tlstest_test.go`) = **1511**, Marke 2000, `schaetzung` 1465; gleich der Ausgabe von `make groessenmarke` |
| Nullblöcke: `go test -race -coverprofile -coverpkg=./src/bootstrap,./src/hexagon/ports/support/tlstest -run 'TestLadeTLS\|TestDatei'`, beide Pakete, Zähler je Block summiert | `tlsdatei.go` 33 Blöcke, `tlstest.go` 47 Blöcke, **0 von 80** mit Zähler 0 (gleich dem Plan; Last 3,4) |
| Rohbytes (`LC_ALL=C grep -P` über alle hinzugefügten Zeilen von `git diff de7589e5..HEAD`) | Backslash-u-vier-Hex: 0 Treffer; BOM, NBSP, Nullbreite, Zeilentrenner: 0; Steuerzeichen: 0. Nicht-ASCII nur `ä ö ü ß § × „ “ … ≤` in Kommentaren und Texten. Gegengelesen: `src/bootstrap/tlsdatei.go`, `src/bootstrap/tlsdatei_test.go`, `src/hexagon/ports/support/tlstest/tlstest.go`, `src/hexagon/ports/support/tlstest/tlstest_test.go`; das NUL im Pfad der Menge L steht als Go-Escape `\x00` im Quelltext, das Werkzeug setzt es nicht um |
| Wiedergabe der Mutationsliste des Implementers (die 86 Einträge der Fix-Runde, Kopie des Werkzeugs `tools/mutationslauf` mit `testZeit` 10 Minuten, `MUTATIONSLAUF_SPEICHER=4g`, unter Sperre) | `zusammenfassung mutationen=86 rot=86 rot-paket=0 gruen=0 flatternd=0 uebersetzungsfehler=0 abbruch=0 dauer=370.8`, Last 3,3 bis 3,8 von 4; gleich der Zusage des Plans |
| `make docs-check`, `make zeichen-check` nach dem Schreiben dieses Reports | Ergebnis im Bericht an den Aufrufer |

## DoD (§2) Punkt für Punkt

| Punkt | Stand | Beleg |
|---|---|---|
| Baustein lädt und prüft (leer, genau ein leerer Pfad, nicht lesbar, kein PEM-Zertifikat oder -Schlüssel, Schlüssel passt nicht, Blatt außerhalb; je ein Grund des kleinsten Ranges, mit dem Namen des Aufrufers, nie Pfad, Inhalt oder Text der Bibliothek; zugesagte Formen und fünf Schlüsselformen gültig; A01 bis A09, A14) | erfüllt | `src/bootstrap/tlsdatei.go` (113 Zeilen): Ränge 1 bis 4 in `lies`, `liesBlatt` und `ladeTLS`, fünf feste Literale (`pruefeLiterale` zählt sie mit `go/parser`), Meldung nach Schlüssel sortiert (`slices.SortStableFunc`); Mengen E 4, L 6, Z 7, S 4, P 6, G 6, Q 25, F 8, K 5, B 16, R 12, H 2, T, X 25, Y 4 im Test mit Zählprüfung gegen die Quelle; 86 von 86 Mutationen rot (selbst wiedergegeben) |
| Baustein liefert die Konfiguration (Mindestversion TLS 1.2, ALPN allein `http/1.1`, keine Client-Zertifikate, Kette der Datei; A10 bis A13) | erfüllt | `MinVersion: tls.VersionTLS12, NextProtos: {"http/1.1"}`, Paar aus `tls.X509KeyPair`; Handshakes im Prozess mit `tlstest.Handshake` (Loopback): Client bis TLS 1.1 scheitert, bis 1.2 verhandelt 1.2, ohne Grenze 1.3; `h2` und `http/1.1` verhandelt `http/1.1`, ohne ALPN nichts, allein `h2` scheitert; Kette mit nur der Wurzel im Pool gelingt, das Blatt allein scheitert |
| Ohne Änderung für den Admin-Listener nutzbar, zustandslos, ohne Datenrennen (A15 bis A17) | erfüllt | kein `LISTEN_` in `tlsdatei.go` (Test A17 mit `ADMIN_TLS_*`, 2 × 11 Elemente); Zustandslosigkeit: Bytes und Änderungszeit vor und nach dem Aufruf; Datenrennen: der Detektor von `make test-race` (frisch, ok) und die Rate 50 von 50 an der Mutation (siehe *Raten*) |
| `make gates` grün | erfüllt (selbst gefahren) | siehe Sensor-Belege |
| Review durchgeführt, Report unter `docs/reviews/` | erfüllt für den Handoff; die Fix-Runde `3e7ce69b` hat kein zweites Review | `docs/reviews/2026-11-21-slice-listener-tls-review.md`; die Fixes löst dieser Bericht gegen die Klasse auf (unten) |
| Doku-Update, falls ein öffentlicher Vertrag berührt ist | erfüllt | `SPEC-291` steht vor dem Code (Architect); die Klarstellungen aus `a97f9f00` stehen in Nr. 2, Nr. 4 und *Ungeprüft* und in der Änderungshistorie; Roadmap trägt die Zeile *In Arbeit*; `spec/architecture.md` unberührt |
| Handbuch: das Binary bleibt unverändert, kein Code, Schlüssel oder Kommando kommt dazu | erfüllt im Code; die Aussage in §7 steht aus (Closure) | `git diff --name-only de7589e5..HEAD`: neun Dateien, davon keine `src/bootstrap/konfig*.go`, `wurzel.go`, `lebenszyklus.go`, `start.go`, kein `src/cmd/`, kein `Makefile`, `Dockerfile`, `.a-check.yml`, `docs/user/`; `make katalog-check` und `make handbuch-check` in `make gates` grün |
| Closure-Notiz, Beobachtungs-Register, Ausgänge in §6, drei Paarungen | stehen aus (Planner nach diesem Bericht) | §7 *Wird bei der Closure gefüllt*; §6 trägt drei Ausgänge *offen* mit Adresse Planner (V-6); alle DoD-Häkchen sind noch leer |

## Plan gegen Code (Diff)

| Plan (§3 Tabelle der Dateien) | Diff `de7589e5..HEAD` | Befund |
|---|---|---|
| `src/bootstrap/tlsdatei.go` neu | 113 Zeilen, nur Standardbibliothek; Signatur `ladeTLS(d tlsDateien, jetzt time.Time) (*tls.Config, []verstoss)` wie `SPEC-291` Kopf | gleich |
| `src/bootstrap/tlsdatei_test.go` neu | 1095 Zeilen (Handoff 911, Fix-Runde +184), eine Datei | gleich |
| `src/bootstrap/export_test.go`, 14 Zeilen, nicht im Block | +14 Zeilen mitten in der Datei (`TLSVerstoss`, `LadeTLS`) | gleich dem Plan; Abweichung vom ursprünglichen Plan, in §6 mit Ausgang *offen* (V-6) |
| `src/hexagon/ports/support/tlstest/` neu | `tlstest.go` 259, `tlstest_test.go` 44; nur Standardbibliothek, kein Schlüssel im Repo | gleich |
| `konfig*.go`, `wurzel.go`, `lebenszyklus.go`, `httpodata/` keine Änderung | keine Änderung | gleich |
| `spec/spezifikation.md` update (Architect) | `SPEC-291` Nr. 2, 4, *Ungeprüft*, Historie; weiter nichts | gleich |
| `.a-check.yml`, Handbuch unverändert | unverändert | gleich |
| Handshakes über `net.Pipe` (Tabelle des Architect) | Loopback (`127.0.0.1:0`, ein Listener nur im Helfer, kein Listener des Produkts) | Abweichung, in §6 benannt (V-6) |
| Tabelle A01 bis A17 | jede Zeile hat ihren Test (siehe unten); Zeilenzahl 17, *a* = 1, *b* = 16 | gleich |

Die Klärungen aus `a97f9f00` sind Klarstellungen: die hinzugefügten Zeilen der Spezifikation nennen keinen Slice, keine Welle und keinen Teil (geprüft: `git diff de7589e5..HEAD -- spec/` auf `slice-`, `welle-`, `Teil <n>`: kein Treffer; allein das Wort *Implementers* in der Zeile der Änderungshistorie, eine Rolle). Die Sortierung nach Bytefolge steht schon in `SPEC-110` Nr. 12 und ist dort jeder Meldung zugesagt; den Anfang des Blocks am Zeilenanfang habe ich an `encoding/pem` des Images gemessen (Probe im Container, Go 1.27.1): Block am Dateianfang gefunden, hinter LF gefunden, hinter CRLF gefunden, Text in derselben Zeile davor nicht, hinter einem einzelnen CR nicht. Wortlaut (`am Anfang der Datei oder hinter einem LF`) und Verhalten stimmen.

## Zusagen von `SPEC-291` gegen Fänger (A01 bis A17)

| Nr. | Zusage der Spezifikation | Fänger und Beleg | Urteil |
|---|---|---|---|
| A01 | Nr. 2: beide leer `(nil, nil)`, nichts gelesen | `TestLadeTLSKeinTLS` (Dateien `cert.pem`, `key.pem` im Arbeitsverzeichnis); Mutationen `a01-*` rot | gefangen |
| A02 | Nr. 2 Rang 1, der Name des anderen Schlüssels | `TestLadeTLSEinerLeer` (Menge E), `TestLadeTLSZweiteVerwendung`; `w08` (Text ohne Namen) rot | gefangen |
| A03 | Rang 2 `Datei nicht lesbar`, nie mit Pfad | Menge L (6) und `TestLadeTLSPfadWieUebergeben`; `v28` (leere Datei als nicht lesbar) rot | gefangen; fehlendes Leserecht ungeprüft (Spec *Ungeprüft*) |
| A04, A05 | Rang 3 Zertifikat und Schlüssel (Nr. 4: Typen, erster Block) | Mengen Z (7), S (4), B (16), D (275); Mutationen `v05`, `v07`, `v09` rot | gefangen bis auf die Ränder V-1 |
| A06 | Rang 4 Schlüssel (`tls.X509KeyPair`, Zwischenzertifikat davor, verschlüsselt) | Menge P (6), R (12), H; `w13` (Schlüssel der Hilfe wiederverwendet) rot | gefangen |
| A07 | Nr. 3: beide Grenzen gültig, keine Toleranz, allein das Blatt, keine Uhr | G (6) und G+ (16, Nanosekunden, Zeit null, Jahr 9999); `v27` (Sekunden), `v29` (Zeit null gültig) rot | gefangen |
| A08 | ein Grund je Schlüssel, unabhängig, Rang n nach Rang n-1 | Menge Q (25, Orakel ohne Code des Bausteins); `w05` (Config bei Verstoß) rot | gefangen |
| A09 | Nr. 4: die zugesagten Formen; Nr. 10: Rechte nicht geprüft | Menge F (8, Element acht `0666`), T (jede Länge), X (25); `v11` rot | gefangen |
| A10 | fünf Schlüsselformen tragen einen Handshake | Menge K (5) mit Handshake; Typen B | gefangen |
| A11 | Nr. 5: `MinVersion` 1.2, Höchstversion Standard, kein mTLS, kein Nachladen, nichts außer den drei Feldern | `TestLadeTLSVersion`, `TestLadeTLSFelder` (Menge C, jedes Feld des Typs per Reflexion); `v18`, `v19` rot | gefangen auf der Ebene von `tls.Config`; eine Ebene tiefer V-2 |
| A12 | Nr. 6: ALPN allein `http/1.1` | `TestLadeTLSALPN` (drei Clients); `w07` (doppelter Eintrag) rot | gefangen |
| A13 | Kette der Datei wird gesendet | `TestLadeTLSKette` (Gegenprobe: das Blatt allein scheitert) | gefangen |
| A14 | Nr. 11: kein Geheimnis in der Meldung | `TestLadeTLSKeinGeheimnis` (33 Elemente, 31 Zeilen, fünf Literale); `v26` (Ablaufdatum im Grund) rot | gefangen |
| A15 | Nr. 12: keine Datei geändert, kein Zustand | `TestLadeTLSZustandslos`, `TestLadeTLSRechte` (Menge M, 56); `v11` rot | gefangen; Ränder V-4 |
| A16 | Nr. 12: gleichzeitige Aufrufe teilen nichts | Detektor je Prozess, `TestLadeTLSGleichzeitig`, `TestLadeTLSTeiltNichts`; Raten unten | gefangen mit Detektor; ohne Detektor Rand V-3 |
| A17 | Nr. 2 Kopf: zweite Verwendung mit anderen Namen | `TestLadeTLSZweiteVerwendung` (2 × 11); `w08` rot | gefangen |

Nr. 1 (Ort im Composition Root, keine neue Kante) trägt `make a-check` (0 Befunde) und A17; Nr. 7, 8, 9, 13 gehören zu Teil 2 und sind nicht Gegenstand.

## Klassen der Fix-Runde gegen die Klasse aufgelöst (`AGENTS.md` §3.17)

Je Befund die Menge aus dem Wortlaut der Zusage, ihre Zahl (vom Test gezählt) und mindestens eine Mutation an einem Element, das weder Review noch Implementer nannten. Die Treiber und Listen liegen im Scratchpad unter `scratchpad/tls-verifier/` (`m1.txt` bis `m4.txt`, `mr.txt`, `mr3.txt`, Ausgaben `m1.out` bis `m4.out`, `mr.out`, `mr3.out`, `cov.out`, `impl-liste.out`).

| Befund | Menge (Wortlaut) | Elemente | Eigene Mutation an einem nicht genannten Element | Ergebnis |
|---|---|---|---|---|
| F-1 (MEDIUM) | `SPEC-291` Nr. 4: „Blöcke anderen Typs … überspringt das Lesen“: Dateiform × Art des fremden Blocks × Zahl davor (Menge D) | 5 × 5 × 11 = 275, vom Test gezählt | `v05` Suche endet bei zwei fremden Blöcken gleichen Typs hintereinander; `v07` endet bei fremdem Block über 200 Bytes; `v09` endet bei Typ `CERTIFICATE REQUEST` | rot, jeweils in `TestLadeTLSBloeckeDavor` (Grund geprüft: `erwartet gültig`, Meldung `kein PEM-Zertifikat` bzw. `keinen PEM-Schlüssel`); `v01` bis `v03` (höchstens 10, 11, 12 Blöcke gelesen) rot; **`w01` (13 Blöcke), `w02` (14) und `v04` (64) grün**: die Prüfgrenze liegt bei 13 gelesenen Blöcken (zehn fremde plus Kette und Schlüssel); `v06` (fremder Block mit PEM-Header beendet die Suche) grün, `v08` (fremder Block über 1000 Bytes) grün: V-1 |
| F-2 (MEDIUM) | Nr. 10 und 12: Rechtemuster × Datei × Ergebnis (Menge M) | 14 × 2 × 2 = 56 | `v11` Rechte auf `0600`, wenn Gruppe oder Welt schreiben dürfen | rot in `TestLadeTLSRechte` (`Dateien geändert`, Muster `0660`, `0664`); `v12` (nur Weltschreibrecht ohne Gruppenschreibrecht) grün, `v13`, `v14` (Verzeichnis: Rechte, Änderungszeit) grün: V-4 |
| F-3 (LOW) | Nr. 2: Bytefolge des Namens, Paare verschiedener Namen (Menge O) | 8 × 7 = 56 | `v16` Vergleich nur des ersten Bytes; `v17` ohne Beachtung der Schreibweise (Kontrolle zu `c08`) | rot in `TestLadeTLSSortiertJeNamenpaar` (`AB` und `A` falsch herum); `v15` (nur die ersten 8 Bytes) grün: V-1 |
| F-4 (LOW) | Nr. 5: das Config trägt `Certificates`, `MinVersion`, `NextProtos` und sonst nichts: jedes Feld von `tls.Config` (Menge C) | Zahl der Felder des Typs per Reflexion gezählt | `v18` `ServerName`, `v19` `InsecureSkipVerify` | rot in `TestLadeTLSFelder` (`Feld ServerName: Nullwert false, zugesagt false`); **eine Ebene tiefer grün**: `v20` `OCSPStaple`, `v21` `SignedCertificateTimestamps`, `v23` `SupportedSignatureAlgorithms` im Paar (`v22` `Leaf` ist gleichwertig): V-2 |
| F-5 (LOW) | Nr. 12: das gelieferte Config wird nach der Rückgabe von keinem Aufrufer geändert; zwei Aufrufe teilen nichts | Liste, Paar, Bytes des Zertifikats | `w03` Vorlage mit `Clone()` (teilt `NextProtos`) | rot in `TestLadeTLSTeiltNichts`; **`v25`** (ein gemeinsames Feld der Paare, je Aufruf an Ort und Stelle überschrieben, gleicher Inhalt) grün: V-3 |
| F-6 (LOW) | Kommentar von `TestLadeTLSGleichzeitig` sagt nur zu, was der Fänger fängt | zwei Zwischenspeicher unter Sperre (`r3`, `r4`) | siehe *Raten* | der Kommentar trifft zu (V-8) |
| F-7 (INFO) | Maß im Negativ 4 | von drei Grenzen eine gefangen | nicht neu gefahren (Review `d01` bis `d03`) | Maß steht im Plan |

## Raten (`AGENTS.md` §3.10)

Messender: das Skript `inner_rate.sh` des Verifiers (Scratchpad `tls-verifier/`), je Mutation zwei Binaries (`go test -c`, mit `-race` bei `CGO_ENABLED=1` und ohne), je Binary **50 Prozesse** `-test.run '^TestLadeTLSGleichzeitig$' -test.count=1 -test.timeout=2m`; Container `--memory=4g --cpus=4` unter `flock`; Last des Hosts je Arm (Load Average 1 Minute zu Beginn, 4 CPUs, zweite Linie parallel) in der Spalte. Eine Rate ist eine Größenordnung unter dieser Last, kein fester Wert.

| Mutation | mit Detektor rot | ohne Detektor rot | Last 1 min |
|---|---|---|---|
| unverändert (Kontrolle) | 0 von 50 | 0 von 50 | 5,8 und 3,2 |
| `r1` Paketvariable hält das letzte Config, ohne Sperre | 50 von 50 | 0 von 50 | 3,1 |
| `r2` gemeinsames Feld der Paare, je Aufruf überschrieben, ohne Sperre (`v25`) | 50 von 50 | 0 von 50 | 3,4 |
| `r3` Zwischenspeicher unter Sperre liefert das Ergebnis des vorigen Aufrufs | 50 von 50 | 50 von 50 | 3,8 und 3,7 |
| `r4` Zwischenspeicher unter Sperre liefert immer das erste Ergebnis | 50 von 50 | 50 von 50 | 3,8 |

Das Rennen (r1, r2) fängt allein der Detektor (50 von 50), das deckt sich mit dem Absatz des Plans und mit dem Negativ 1. Die Zwischenspeicher unter Sperre (r3, r4) färben `TestLadeTLSGleichzeitig` in jedem Prozess rot, auch ohne Detektor; die Mutation `c19` des Reviews (49 von 50 ohne Detektor) war eine andere Form.

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| V-1 | LOW | Die Klassen der Fix-Runde mit unbeschränkter Menge nennen keine Prüfgrenze mit Grund und Folge: Zahl der Blöcke (Menge D endet bei 13 gelesenen Blöcken: `w01`, `w02`, `v04` grün, `v01` bis `v03` rot), Länge des gemeinsamen Präfixes der Namen (Menge O: Namen höchstens 2 Bytes; `v15` mit 8 Bytes grün; die echten Namen `LISTEN_TLS_CERT`/`…KEY` und `ADMIN_TLS_CERT`/`…KEY` unterscheiden sich erst ab Byte 10), PEM-Header an einem fremden Block (`v06` grün) und Größe eines fremden Blocks (`v08` grün). Der Kommentar von `TestLadeTLSBloeckeDavor` sagt „an jeder Stelle“ und nennt danach „0 bis 10“. Folge im Betrieb: keine Datei eines Betreibers erreicht eine dieser Grenzen; das Negativ fehlt trotzdem im Plan | `AGENTS.md` §3.17 (unbeschränkte Klasse: Prüfgrenze mit Grund und Folge), §3.15 (akzeptiertes Negativ mit Maß), §3.11 | `docs/plan/planning/in-progress/slice-listener-tls.md` · „Akzeptierte Negative“; `src/bootstrap/tlsdatei_test.go` · „TestLadeTLSBloeckeDavor“ | ja (Mutationen `w01`, `w02`, `v04`, `v06`, `v08`, `v15`) | akzeptiertes Negativ fehlt: Prüfgrenze unbeschränkter Klassen |
| V-2 | LOW | `TestLadeTLSFelder` prüft jedes Feld von `tls.Config` (Menge C), aber keines von `tls.Certificate` im einen Paar: `OCSPStaple` (`v20`, schickte jedem Client eine feste Antwort), `SignedCertificateTimestamps` (`v21`) und `SupportedSignatureAlgorithms` (`v23`, Liste mit SHA-1) bleiben grün (3 von 3, `Leaf` gleichwertig). `SPEC-291` Nr. 5 sagt „genau ein Paar mit der ganzen Kette der Datei“ und lässt die übrigen Felder des Paares offen; die Klasse „Positivliste statt Zusage“ (Review F-4) gilt eine Ebene tiefer weiter | `AGENTS.md` §3.15, §3.17; `SPEC-291` Nr. 5 | `src/bootstrap/tlsdatei_test.go` · „func TestLadeTLSFelder“ | ja (Mutationen `v20`, `v21`, `v23`) | Fänger prüft eine Positivliste, eine Ebene tiefer |
| V-3 | LOW | `TestLadeTLSTeiltNichts` führt eine Datei und prüft das frühere Config erst nach dem dritten Aufruf; ein gemeinsames Feld der Paare, das jeder Aufruf mit gleichem Inhalt überschreibt (`v25`), heilt sich dabei selbst und bleibt grün (der Unit-Test ist grün, der Test von A16 ohne Detektor in 0 von 50 Prozessen rot, mit Detektor in 50 von 50). Der Kommentar „teilen keine Liste und keinen Wert“ ist breiter als der Fänger; das Negativ 1 des Plans nennt allein „A16 ohne Detektor“, nicht diesen Rand. Folge: fällt der Detektor aus (kein cgo), bleibt dieses Teilen ungesehen | `AGENTS.md` §3.15 (Negativ mit Grund, Folge, Maß), §3.11, §3.10 (Nebenläufigkeit) | `src/bootstrap/tlsdatei_test.go` · „func TestLadeTLSTeiltNichts“; Plan · „Akzeptierte Negative“ Nr. 1 | ja (Mutation `v25`/`r2`) | Fänger schmaler als die Zusage „teilen nichts“ (Reihenfolge der Prüfung) |
| V-4 | INFO | Plan §3 (Klasse M) sagt, „das Verzeichnis“ sei nach dem Aufruf unverändert; die Momentaufnahme hält die Namen im Verzeichnis, nicht dessen Rechte und Änderungszeit (`v13`, `v14` grün). Die Zusage der Spezifikation spricht von Dateien (Nr. 12), das Verzeichnis ist nicht Gegenstand. Ebenso `v12` (nur Weltschreibrecht): kein Muster der 14 trägt Bit `0o002` ohne `0o020`; jedes Einzelbit der Rechte ist in beiden Zuständen geführt | `AGENTS.md` §3.11 (Plantext breiter als der Fänger) | Plan · „Klassen der Fix-Runde“ (M) | ja (Mutationen `v12` bis `v14`) | Plantext breiter als der Fänger; ohne Aktion des Implementers |
| V-5 | INFO | Der Wrapper `LadeTLS` in `export_test.go` bildet eine leere Liste der Verstöße auf `nil` ab; `(nil, []verstoss{})` statt `(nil, nil)` bei „beide leer“ ist für den Test nicht sichtbar (`w06` grün). Wirkung für den Aufrufer: keine (`len(vs) == 0`), `SPEC-291` Nr. 2 sagt `(nil, nil)` | `AGENTS.md` §3.15 (Fänger sieht Fake-Normalisierung) | `src/bootstrap/export_test.go` · „func LadeTLS“ | ja (Mutation `w06`) | Naht normalisiert (äquivalent in der Wirkung) |
| V-6 | INFO | Drei Risiken in §6 tragen den Ausgang *offen*, Adresse Planner; sie sind vor `done/` zu schließen (DoD: jedes Risiko trägt einen Ausgang). Bewertung: (a) `export_test.go` +14 Zeilen außerhalb des Blocks: Summe der Änderung 1511 + 14 = 1525 von 2000, die Marke ist mit und ohne die Datei eingehalten; nähme der Planner die Datei in den Block, zählte die Messung (`SPEC-128` Nr. 10 g) jede Änderung der Datei zwischen `basis` und `HEAD`, auch die nach einem `git merge main` aus fremden Slices — das spricht für „außerhalb lassen“; (b) Loopback statt `net.Pipe`: keine Zusage ändert sich, `test`, `race` und `integration` laufen netzlos im Container mit Loopback grün; (c) `tools/mutationslauf` fährt `src/bootstrap` nicht (Nullprobe-Frist 2 Minuten): von mir nicht neu gemessen; die Liste der 86 lief mit der Kopie des Implementers (`testZeit` 10 Minuten) durch, Gesamtdauer 370,8 s | Plan §6; `AGENTS.md` §3.13; `SPEC-128` Nr. 10 g | Plan · §6 „Naht von `ladeTLS` im Testpaket“, „Handshakes im Prozess über den Loopback“, „Werkzeug `tools/mutationslauf` fährt dieses Paket nicht“ | nein | Abweichung vom Plan mit offenem Ausgang (Adresse Planner) |
| V-7 | INFO | Die Messung 1511 steht nicht im Plan (der Block trägt `schaetzung` und `marke`, `make groessenmarke` meldet sie); sie liegt 489 unter der Marke und 46 über der Schätzung 1465 (3 %). Die Zeile `Messung` und der Name dieses Reports gehören bei der Closure in Plan und Evidence (Kennung ohne Pfad und Endung) | `SPEC-128` Nr. 10 a; `AGENTS.md` §3.14 | Plan · §4 „Größenmarke“ | ja (`make groessenmarke`) | Planner-Aufgabe der Closure |
| V-8 | INFO | Der Kommentar von `TestLadeTLSGleichzeitig` („ein Zwischenspeicher unter einer Sperre … färbt die Gleichheit rot“) trifft für beide gemessenen Formen zu (`r3`, `r4`: je 50 von 50 mit und ohne Detektor); die Mutation `c19` des Reviews war ein dritter Fall (49 von 50 ohne Detektor). Der Plan führt für die Form unter Sperre keine Rate | `AGENTS.md` §3.11, §3.10 | `src/bootstrap/tlsdatei_test.go` · „TestLadeTLSGleichzeitig“ | ja (Mutationen `r3`, `r4`) | Kommentar mit Rate belegt; ohne Aktion |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/tlsdatei.go` gegen `SPEC-291` Nr. 2 bis 5, 10 bis 12 (Ränge, Reihenfolge der Meldung, feste Literale, Gültigkeit mit beiden Grenzen, kein eigener Leser, keine Uhr, keine Datei geschrieben) | geprüft, ohne Befund am Code; Befunde nur an Fängern (V-1 bis V-3); jede eigene Mutation, die rot sein muss (26 von 41), wird aus dem benannten Grund rot (Ausgaben `m4.out`: Zeile des Tests und Text der Meldung) |
| Rechte der Schlüsseldatei nicht geprüft (Entscheidung des Users, `SPEC-291` Nr. 10) | geprüft, ohne Befund: weder `os.Stat` noch `Mode` im Code, A09 Element acht gültig |
| Binary, Konfiguration, Verdrahtung, Handbuch, Katalog | geprüft: unverändert (Diff-Namen, `make katalog-check`, `make handbuch-check` grün); Teil 2 trägt sie |
| `tlstest` (nur Standardbibliothek, kein Schlüssel im Repo, `t.Fatalf` mit Präfix `tlstest: `, Rechte unabhängig von `umask`, `Handshake` schließt die Gegenseite) | geprüft, ohne Befund; fünf Mutationen am Helfer rot: `w09` (kein `Chmod`), `w10` (Fehler des Clients verschluckt), `w11` (Protokoll leer), `w12` (Präfix entfällt), `w13` (Schlüssel wiederverwendet: die Fixture „anderer Schlüssel“ wird gleich) |
| Plan-Nachzug nach `AGENTS.md` §3.9 (Commit `3e7ce69b`: §3 Mengen, Klassen, Mutationen, Negative; Loopback und Naht in §3 und §6) | geprüft, ohne Befund: die Zahlen 86 Mutationen (je Befund der Fix-Runde 12, 5, 2, 10, 2 = 31), 80 Nullblöcke, 275 und 56 Elemente stimmen mit meinen Läufen und den Zählungen der Tests |
| Fixtures nach `AGENTS.md` §3.18 | geprüft, ohne Befund (Rohbytes oben); PEM-Varianten entstehen aus Bytewerten (`bytes.ReplaceAll`, `slices.Concat`, `pem.EncodeToMemory`) |
| Kommentare nach `AGENTS.md` §3.7 und §3.11 (Stichprobe: `tlsdatei.go` ganz, `tlstest.go` ganz, die Köpfe der Tests A01 bis A17 und der neuen Tests `…BloeckeDavor`, `…Rechte`, `…Felder`, `…TeiltNichts`, `…SortiertJeNamenpaar`) | geprüft: jeder Kopf trägt Zusage, Kopplung oder Rang-Zeiger, keiner beschreibt eine verworfene Alternative; breiter als der Fänger sind die Köpfe von `…BloeckeDavor` („an jeder Stelle“, V-1) und `…TeiltNichts` („keine Liste und keinen Wert“, V-3); der Kopf von A16 ist belegt (V-8) |
| Größenmarke: `export_test.go` +14 außerhalb des Blocks und Loopback gegen `SPEC-128` Nr. 10 g | geprüft, siehe V-6: keine Überschreitung in beiden Zählweisen, kein Befund an der Marke |
| Zeichen, Nicht-ASCII, Steuerzeichen, Backslash-u-Folgen | geprüft, ohne Befund (Rohbytes oben, `make zeichen-check` in `make gates` und nach diesem Report) |

## Nicht geprüft (benannt)

Die Nullprobe des Werkzeugs `tools/mutationslauf` im Repo (Frist 2 Minuten, ganzes Paket `src/bootstrap`) habe ich nicht neu gemessen. Die Stufen `test`, `build` und `lint` von `make gates` kamen aus dem Docker-Cache bei gleicher Eingabe; frisch gefahren ist allein `race`. `make gegenproben` ist nicht gefahren (Auftrag). Die Zahl der Eingabeformen einer unbeschränkten Klasse ist nicht erschöpfbar; die Ränder stehen in V-1 bis V-4.

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 3 |
| INFO | 5 |

**Eigene Mutationen:** 41 verschiedene (Listen `m1.txt` bis `m3.txt` im Scratchpad), davon 26 rot und 15 grün; von den 15 grünen ist 1 gleichwertig (`v22`), 1 ohne Wirkung für den Aufrufer (`w06`), und 13 stehen als Ränder in V-1 bis V-4 (`w01`, `w02`, `v04`, `v06`, `v08`, `v15` in V-1; `v20`, `v21`, `v23` in V-2; `v25` in V-3; `v12`, `v13`, `v14` in V-4); dazu zwei Kontrollen mit gleichwertiger Änderung (`v00`, `w14`), beide grün, der Treiber unterscheidet also. Zwei Läufe waren ungültig und sind wiederholt: `v24` (Aufbaufehler im Treiber, wiederholt als `w03`) und `w04` (ohne cgo, wiederholt als `w04b` mit Detektor); die Raten des ersten Anlaufs (ohne cgo) sind verworfen und neu gefahren.

## Verdikt

**Freigegeben** (Verifier, 2026-11-21). Übergabe an den Planner: V-1 bis V-3 als akzeptierte Negative mit Grund, Folge und Maß in §3 nachtragen oder an den Implementer zurückgeben (je ein Fall in `TestLadeTLSBloeckeDavor`, `TestLadeTLSFelder` und `TestLadeTLSTeiltNichts`, im letzten das frühere Config vor dem dritten Aufruf zu prüfen); V-6 und V-7 in der Closure (drei Ausgänge schließen, Messung 1511 eintragen); Closure-Notiz, Register und die drei Paarungen sind Sache des Planners.
