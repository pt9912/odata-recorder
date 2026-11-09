# Verifikation: slice-etag-versionen — 2026-11-08

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-280` bis `SPEC-282` samt Nachzügen, die Tabelle *Zusage · Fänger* E1 bis E24 und die Hard Rules; nicht Diff-Review (siehe `2026-11-08-slice-etag-versionen-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-etag-versionen`, HEAD `a5246418` (Fix-Runde nach dem Review; davor Entscheidungen des Architect `4204a3a0`, Review `05899549`, Code `7783f4f5`), Worktree `/Development/odata-recorder-wt-etag-versionen`, Basis `af1f1bac`. Plan: `docs/plan/planning/done/slice-etag-versionen.md`. Bezug: [`LH-FA-15`](../../spec/lastenheft.md), [`LH-QA-02`](../../spec/lastenheft.md), [`LH-FA-07`](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden, mit Auflagen an Planner und Architect (Plan- und Vertrags-Nachzug, Closure); keine Fix-Runde am Code.** `make gates` ist im eigenen Lauf grün (Exit 0), das Paket-Set der Tests ist im eigenen Lauf grün, die Fix-Runde zu F1 deckt die Klasse (acht von acht Mutationen an Elementen, die weder Befund noch Fix nannten und die in der Menge liegen, sind rot; die zwei grünen sind die benannte Prüfgrenze), die Wirkungs-Fänger E18 und E19 sind in jedem Prozess rot (12 von 12, 10 von 10), und der Merge mit `main` und den beiden anderen Zweigen hat nur die bekannten Textkonflikte. Eine Zusage ohne Träger (V1, MEDIUM) bleibt: der Satz „das Handbuch nennt, dass `If-Match` ohne Wirkung ist“ steht im Vertrag und in E24, aber nicht mehr im Handbuch.

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/vf-gates.txt`: `rc=0`, 10 min 39 s einschließlich Wartezeit auf die Sperre. Darin: `d-check: 924 Datei(en) geprüft, 0 Befund(e)`; `pruefung test` und `pruefung race` laufen; `groessenmarke slice-etag-versionen.md: ok, Messung 1999 von Marke 3000`; die Stufen `handbuch` und `katalog` bauen. Die Meldung von a-check über Dateien unter `tools/` ohne Schicht ist ein Hinweis, kein Befund (Exit 0), und gehört nicht zu diesem Diff. |
| Tests der Pakete des Slice | Container aus dem Image `odata-recorder:race`, Quellen per `docker cp` eingebracht (kein geteiltes Image-Tag überschrieben, kein Mount), `--network none --memory=4g --cpus=4`, unter derselben Sperre: `model`, `services/mock`, `jsoncodec`, `jsoncodecv2`, `csdl`, `csdlv2`, `urlparser*`, `httpodata` und `src/bootstrap` (209,99 s) alle `ok`. |
| Bestehende Abnahme unverändert | `git diff --stat af1f1bac HEAD -- src/bootstrap/abnahme*_test.go tests`: nur `abnahme_etag_test.go` neu (150 Zeilen), keine bestehende Abnahme-Datei berührt; `TestVertragMatrix` mit `TestVertragMatrixZaehlung` (36 IDs, 92 Schritte) läuft in `src/bootstrap` grün, `ETAG_ENABLED` aus (Vorgabe). `TestAbnahmeMatrixMitEtag` prüft die Zählung 36 und 92 selbst (Zeile 144) und ist grün. |
| V2 liefert nie ETag | `TestEtagNieGeschrieben` (`jsoncodecv2`) grün; Mutation an `etagFuerSatz` (E4) rot, siehe §4. |
| Plan-vs-Code-Diff | siehe §3. |
| Merge-Nähe | siehe §6. |
| Größe | `Messpunkt … produktcode=260 schaetzung=2190 marke=3000 messung=1999`: Auslastung 0,67 der Marke, 0,91 der Schätzung; unter der Marke, kein Eintrag `ueberschreitung` nötig. |
| §3.2 / §3.18 | `//nolint` im Diff: 0 Treffer (`git diff af1f1bac HEAD -- src \| grep nolint`). Gegenlesen auf Rohbytes über diesen Bericht: 0 Treffer (Muster im Schluss). |
| Fehlercode-Zahl | `src/hexagon/model/fehler.go` im Diff nicht berührt; 40 verschiedene Codes im Quelltext (`grep -rhoE '"ODR-[A-Z][0-9]{4}"' … \| sort -u \| wc -l`); `katalog` im eigenen Lauf `ok`. |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Bildung und Auslieferung (Header und Darstellung, gleicher Zustand gleicher ETag, auch nach Reset und Restore) | **erfüllt, Haken fehlt** | `EtagFuer` (`model/etag.go`), `etagVon`, `liefere` (ein Lesen, `erg.Entities[0].Etag`), `kopf204`, `antwort`; Fänger E1, E7 bis E15, E21, E22; Mutationen §4. |
| Zähler (Transaktion, Fehlschlag, andere Entities und Workspaces, Neuanlage nach `DELETE`) | **erfüllt, Haken fehlt** | der Zähler selbst ist Bestand (`SPEC-193` Nr. 4); der Slice fügt Fänger E16 bis E20 hinzu; Wirkungs-Fänger E18 und E19 mit Rate in §5. |
| Aktivierung (aus: weder Header noch `@odata.etag`) | **erfüllt, Haken fehlt** | `Dienst.Etag`, `etagFuerSatz` (nur Version `4.0`), `ETAG_ENABLED` Vorgabe `false`; E3, E4, E5, bestehende Abnahme grün. |
| `make gates` grün | erfüllt | eigener Lauf, `rc=0`; nach dem Merge von `main` einmal wiederholen (Auflage A3). |
| Review durchgeführt, Report liegt vor | erfüllt (Haken fehlt) | `docs/reviews/2026-11-08-slice-etag-versionen-review.md`; F1 bis F3 in `a5246418` abgearbeitet (§4). |
| Doku-Update | erfüllt | öffentlicher Vertrag `SPEC-280` bis `SPEC-282` und die sieben Nachzüge (`SPEC-049`, `SPEC-110`, `SPEC-168`, `SPEC-169`, `SPEC-193`, `SPEC-194`, `SPEC-198`) stehen im Diff von `spec/spezifikation.md`; Historie-Zeilen vom 2026-11-08 sind da. |
| Handbuch nachgezogen (Zeile `ETAG_ENABLED`), Kataloge-Gate grün | erfüllt, mit V1 | Zeile steht; `katalog` und `handbuch` im `make gates` grün; ohne die Zeile ist `katalog` rot (§4). |
| Closure-Notiz, Register, Paarungen, Ausgang je Risiko (§6) | **offen, Sache des Planners** | §7 ist „Wird bei der Closure gefüllt“; die §6-Punkte stehen auf *offen* (siehe V2 und Auflage A2). |

Liefer-Punkte: drei, wie der Plan zählt.

## 3. Plan-vs-Code-Diff

- **Unberührt, wie zugesagt:** `Dockerfile`, `Makefile`, `harness/`, `tools/`, `.golangci.yml`, `.a-check.yml`: `git diff --stat af1f1bac HEAD -- Dockerfile Makefile harness tools` ist leer. Kein Gate gelockert (`AGENTS.md` §3.6).
- **Außerhalb von `src/`:** `docs/plan/planning/done/slice-etag-versionen.md`, `docs/plan/planning/done/slice-if-match.md` (Sendung, §1 angenommen), `docs/reviews/…-review.md`, `docs/user/benutzerhandbuch.md` (eine Zeile), `spec/spezifikation.md`. Nichts sonst.
- **Produktcode:** zwölf Dateien, genau die Pfade des Blocks plus `jsoncodec/schreiber_etag.go` (neu, im Pfad `jsoncodec/`): `model/{etag,entity,abfrage}.go`, `services/mock/{deuten,lesen,mock,schreiben}.go`, `jsoncodec/{schreiber_bau,schreiber_etag}.go`, `bootstrap/{konfig,konfig_werte,wurzel}.go`. `src/bootstrap` ist auf die drei Dateien des Blocks beschränkt. Tests außerhalb des Blocks: `bootstrap/wurzel_test.go`, `bootstrap/export_test.go` und `jsoncodecv2/schreiber_etag_test.go`, wie §6 des Plans angibt.
- **Verhalten gegen den Vertrag, gelesen:** `etagFuerSatz` verlangt Version `4.0` (`SPEC-280` Nr. 2); der Dienst verzweigt nicht nach der Version, er liest nur `Dienst.Etag`. `liefere` setzt die Kopfzeile aus derselben Entity, die den Body trägt (ein Lesen, `SPEC-281` Nr. 1); `DELETE` trägt keinen ETag (`kopf204`); der Schreiber schreibt das Mitglied nur in `minimal`, hinter `@odata.id`, und prüft die Form vorher (`pruefeEtag`, `ErrUngueltig`); der Leser setzt `Etag` nie (Feld nur im Schreiber-Pfad).
- **Abweichung vom Plan:** keine am Code. Die Fix-Runde ändert Tests, Plan und die Handbuch-Zeile; zu der Zeile siehe V1.

## 4. Fix-Runde gegen die Klasse aufgelöst (§3.17)

Messender: Verifier-Agent. Je Mutation Wegwerfkopie im Container (`/orig` als unveränderte Kopie, `sed` auf `/src`, danach zurückkopiert), `go test -count=1` des Pakets, `--network none --memory=4g --cpus=4`, unter der Sperre. **Treiber ist ein eigenes Skript im Scratchpad (`vf-mut.sh`), nicht `tools/harness/mutationslauf.sh`:** dessen Nullprobe-Grenze trägt `src/bootstrap` nicht (Vorgabe des Aufrufers). Eine Mutation ohne Änderung der Datei bricht der Treiber ab (`KEINE AENDERUNG`).

### F1 — `TestEtagFuerKlasse`

Menge, aus dem Wortlaut der Zusage (`SPEC-280` Nr. 1: `EtagFuer(n)` ist `W/"` + Dezimalzahl + `"` für jedes `n` von 0 bis 2^63−1, `SPEC-108`: nichtnegativ) selbst gebildet: **die 2^63 nichtnegativen 63-Bit-Zahlen**; unbeschränkt im Sinn von nicht aufzählbar. Die Fix-Runde prüft 301 + 54 + 189 + 1 + bis zu 4000 Elemente (Schleife 0 bis 300, 10^k mit Nachbarn für k = 1 bis 18, 2^k mit Nachbarn für k = 1 bis 62, 2^63−1, 2000 gleichverteilte und 2000 logarithmisch verteilte Zahlen mit festem Seed), Orakel eine Ziffernrechnung ohne `strconv`; die Prüfgrenze steht im Plan §3 mit Grund und Folge. Meine Prüfung: das Orakel ist unabhängig vom Code, der Fänger ist eine Schleife, die Zählprüfung (mindestens 4000) steht.

| Nr. | Mutation an `strconv.FormatInt(version, 10)` in `etag.go` | Element, das weder Befund noch Fix nannte? | Ergebnis |
|---|---|---|---|
| V-F1a | Band 2^40 bis 2^50, Version + 1 | ja (Band, das keine Stichprobe nennt) | rot `TestEtagFuerKlasse` (nur dieser; Aufbau gebaut) |
| V-F1b | Umwandlung `int64(float64(version))` (53 Bit) | ja (jedes Element über 2^53, das keine Zweierpotenz ist) | rot `TestEtagFuer`, `TestEtagFuerKlasse` |
| V-F1c | Vorzeichen bei 2^63−1 | genannt im Auftrag, im Fix als Stichprobe | rot `TestEtagFuer`, `TestEtagFuerKlasse` |
| V-F1d | Vorzeichen ab 3 mal 2^61 (ein Viertel des Bereichs) | ja | rot `TestEtagFuer`, `TestEtagFuerKlasse` |
| V-F1e | Abschneiden auf `int32` ab 2^31 | ja | rot `TestEtagFuer`, `TestEtagFuerKlasse` |
| V-F1f | hexadezimal ab 2^45 | ja | rot `TestEtagFuer`, `TestEtagFuerKlasse` |
| V-F1g | Version + 1, wenn Zehnerstellen gleich 77 und Zahl ab 10^15 (Rest hinter dem Band, Anteil 1 von 100) | ja | rot `TestEtagFuerKlasse` |
| V-F1h | letzte Ziffer weg im Band 10^17 bis 2^58 (Länge der Zahl) | ja | rot |
| V-F1i | Vorzeichen nur bei 2^63−2 | ja, Einzelelement | **grün** (Prüfgrenze) |
| V-F1j | Version + 1 nur in 3 mal 2^61 plus/minus 5 | ja, enges Band | **grün** (Prüfgrenze) |

Urteil: Die Klasse „Bänder, Breiten, Vorzeichen, Umwandlungen, Ziffernstellen“ ist gedeckt; die beiden grünen Mutationen liegen in einem einzelnen Element beziehungsweise einem Band von elf Zahlen zwischen den Stichproben und sind genau die im Plan benannte Prüfgrenze (Grund: Menge nicht aufzählbar; Folge: eine solche Abweichung bliebe ungefangen, die Funktion ist eine Zeile ohne Verzweigung). Ich nehme das Negativ an; es nennt Grund und Folge (`AGENTS.md` §3.15).

### Weitere Mutationen an Stellen, die der Review nicht nannte

| Zusage | Mutation | Ergebnis |
|---|---|---|
| E3 | Vorgabe `ETAG_ENABLED` auf `true` | rot `TestKonfigEtagEnabled` |
| E3 | Wert wird gelesen, aber nicht in die Konfiguration übernommen | rot `TestKonfigEtagEnabled` |
| E4 | `etagFuerSatz` prüft die Version nicht (`return an`) | rot `TestEtagFuerSatz` |
| E4 | `etagFuerSatz` ignoriert `an` (nur Version) | rot `TestEtagFuerSatz`, `TestEtagVerdrahtung` |
| E4 | `etagFuerSatz` immer falsch | rot `TestEtagFuerSatz`, `TestEtagVerdrahtung` |
| E10 | Projektion kopiert `Etag` nicht | rot `TestFuehreAusBehaeltEtag` |
| E13 | Schreiber schreibt das Mitglied nie (`if e.Etag != "" && false`) | rot `TestEtagMitglied` |
| E7 | `HEAD` ohne Kopfzeile `ETag` (`etag != "" && !l.kopf`) | rot `TestEtagAusUndAn`, `TestEtagWiederholung`, `TestEtagEinLesen` |
| E15 | `DELETE` trägt einen ETag (Bedingung `mu.methode != methodeDelete` entfernt) | rot `TestEtagAusUndAn`, `TestEtagMutationen` |

Nicht gefahren (aus E3, Mutationen des Plans): „Prüfung nur im Modus `mock`“ und „Wert in der Meldung“; beide hat der Implementer rot gesehen (Plan §3), ich habe sie nicht wiederholt.

### Offener Punkt des Reviewers: E24 / E3 „Zeile entfernt“

Aus den Quellen von `tools/katalog` und `tools/pruefung` im Container gebaut, das Binary aus `./src/bootstrap/main` mit `-X main.version=dev`, dann beide Prüfungen gegen das Handbuch mit und ohne die Zeile `ETAG_ENABLED`:

| Handbuch | `katalog` | `pruefung handbuch` |
|---|---|---|
| unverändert | `ok`, rc 0 | `ok`, rc 0 |
| Zeile `ETAG_ENABLED` entfernt | **rot**, rc 1: `katalog: Schlüssel: ETAG_ENABLED fehlt im Handbuch` | **grün**, rc 0 |

**`make handbuch-check` fängt das Entfernen der Zeile nicht; `make katalog-check` fängt es.** Das stimmt mit der Spalte *Mutation* der Zeile E24 überein (dort steht allein `make katalog-check` rot), die Spalte *Fänger* nennt beide Ziele; `handbuch-check` prüft Kopf, Regel und Beispiele, nicht den Schlüsselkatalog. Kein Befund, aber die Zelle sollte es sagen (Auflage A2).

## 5. Raten der Wirkungs-Fänger E18 und E19 (`AGENTS.md` §3.10)

Messender: Verifier-Agent. `go test -race -c` (CGO an, Image `race`) beziehungsweise `go test -c`, dann *N* Prozesse des Test-Binary mit `-test.count=1` im Container (`--cpus=4 --memory=4g`); rot heißt Exit ungleich 0 irgendeiner Wiederholung des Tests im Prozess. Die Mutationen sind **eigene Bauarten** (nicht die Skripte des Implementers): E18 liest die Version des Eintrags **vor** der Transaktion aus dem Repository (`m.k.Entities.Get` vor `WithWorkspaceTransaction`, `imWorkspace` verwendet dieses Ergebnis); E19 liest für die Kopfzeile **ein zweites Mal** (`m.k.Entities.Get` nach der Darstellung, `EtagFuer(es2.Version)`). Last des Hosts (Lastmittel der Minute) vor und nach dem Lauf, gemeinsam mit anderen Slices: 7,0 bis 8,5.

| Mutation · Test | mit Detektor | ohne Detektor | Datenrennen gemeldet |
|---|---|---|---|
| ohne Mutation · `TestEtagGleichzeitigeAenderungen`, `TestEtagLesenWaehrendAenderung`, `TestEtagEinLesen` (Fake) | rot 0 von 12 | – | 0 |
| ohne Mutation · `TestEtagGleichzeitigSQLite` | rot 0 von 6 | – | 0 |
| E18 Version vor der Transaktion · `TestEtagGleichzeitigeAenderungen` (Fake) | **rot 12 von 12** | **rot 12 von 12** | 0 |
| E18 Version vor der Transaktion · `TestEtagGleichzeitigSQLite` | **rot 10 von 10** | – | 0 |
| E19 zweites Lesen · `TestEtagLesenWaehrendAenderung` | **rot 12 von 12** | **rot 12 von 12** | 0 |

Das deckt sich mit den Zahlen im Plan (20/20, 20/20, 10/10) und mit der Messung des Reviewers (10/10, 10/10, 5/5). Der Detektor meldet bei diesen Mutationen kein Datenrennen (jeder Zugriff liegt unter einer Sperre des Fakes oder der Transaktion): der Fänger zählt die Wirkung, die Aussage des Plans stimmt. Die zwei Zeilen der Mutation „Antwort liest nach `Put` erneut“ (0 von 20, äquivalent) habe ich nicht neu gemessen; das Argument des Plans (gleiche Sperre, gleicher Eintrag) trägt, und `TestEtagEinLesen` fängt das zweite Lesen ohne Gleichzeitigkeit.

## 6. Merge-Nähe

`git merge-tree --write-tree HEAD <Zweig>` (nur Berechnung, kein Merge):

| Gegenstück | Ergebnis |
|---|---|
| `main` (`c5dd4b2a`) | ein Textkonflikt in `spec/spezifikation.md`, zwei Stellen: die Einfügestelle vor `SPEC-280` (HEAD trägt `SPEC-280`/`281`, `main` trägt `SPEC-261`) und die Änderungstabelle/Historie. Beides ist „beide Seiten behalten“; **keine Kennung kollidiert** (`SPEC-260`/`261` gegen `SPEC-280` bis `282`), `wurzel.go`, `konfig.go` und alle `src/`-Dateien mergen ohne Konflikt. |
| `slice-admin-lebenszyklus` | derselbe Textkonflikt (Änderungstabelle/Historie in `spec/spezifikation.md`); `src/bootstrap/wurzel.go` und `konfig.go` mergen ohne Konflikt (der Zweig berührt `src/adapters/driving/admin/`). |
| `slice-fallback-policy-entscheidung` | derselbe Textkonflikt in `spec/spezifikation.md`; `src/` ohne Konflikt. |

Fehlercodes: 40 in `fehler.go`, in keinem der Zweige geändert. Nach dem Merge der Spezifikation `make gates` einmal wiederholen (A3). Die Merge-Reihenfolge steht im Plan §6 (zuerst `slice-fallback-policy`, dann `slice-admin-lebenszyklus`, dann dieser Slice).

## 7. Tabelle *Zusage · Fänger* gegen den Vertrag

Geprüft Zeile für Zeile gegen `SPEC-280` Nr. 1 bis 5, `SPEC-281` Nr. 1 bis 4, `SPEC-282`, `SPEC-110` Nr. 2 und 15, `SPEC-168` Nr. 3 (e), `SPEC-169` Nr. 5. Jede der Zeilen E1 bis E24 nennt einen Fänger, dessen Symbol im Quelltext steht: `grep -rn "^func <Name>("` über alle 31 genannten Tests ergab 31 Treffer, kein „FEHLT“ (`TestEtagFuer`, `TestEtagFuerKlasse`, `TestEtagNurAusVersion`, `TestEtagWiederholung`, `TestEtagAusUndAn`, `TestKonfigEtagEnabled`, `TestEtagFuerSatz`, `TestEtagVerdrahtung`, `TestEtagPost`, `TestEtagHeaderKeinEinfluss`, `TestAntwortEtag`, `TestEtagEinLesen`, `TestEtagCollection`, `TestEtagAccept`, `TestFuehreAusBehaeltEtag`, `TestEtagMitglied`, `TestEtagKeinETagBeiFehler`, `TestEtagUngueltig`, `TestEtagGenerator`, `TestEtagNieGeschrieben`, `TestEtagMutationen`, `TestEtagFehlschlag`, `TestEtagAndereBleiben`, `TestEtagGleichzeitigeAenderungen`, `TestEtagGleichzeitigSQLite`, `TestEtagLesenWaehrendAenderung`, `TestEtagLoeschenNeuanlage`, `TestEtagResetStelltHer`, `TestEtagRestoreStelltHer`, `TestAbnahmeMatrixMitEtag`, `TestEtagRundlauf`).

- **E6** (`If-Match` nie gelesen): akzeptiertes Negativ in `SPEC-280` Nr. 3, mit Grund (einfache Map, kein Haken) und Folge (nur Wirkung geprüft: 72 Fälle); die Allaussage „liest nie“ ist damit nicht mehr als Zusage ohne Fänger geführt. Angenommen.
- **E16** (Abbruch): die Aussage „Kontext beendet vor … der Transaktion“ ist zurückgenommen; gefangen sind die Kontextfehler als Fehler von `Put` und `Advance`; die Abbruch-Zellen GET Key und GET Collection stehen als Negativ im Plan. Angenommen: der Diff fügt keinen Aufruf eines Ports hinzu (gelesen in `lesen.go`: nur `liefere` bekommt ein Argument), das Negativ nennt Grund und Folge.
- **E20** `DELETE` erhöht die Version nicht: kein Test dieses Slice fängt es, mindestens sechs vorhandene Tests des Pakets tun es (vom Reviewer mit R20 gemessen); der Plan sagt das ausdrücklich.
- **E24:** siehe V1.

### V1 · MEDIUM — Zusage ohne Träger: „das Handbuch nennt, dass `If-Match` ohne Wirkung ist“

- `quelle`: `AGENTS.md` §3.15 (jede Zusage hat einen Fänger oder ein akzeptiertes Negativ mit Grund und Folge), §3.9 (wer eine Aussage berichtigt, zieht alle Stellen nach), `SPEC-280` Nr. 3, `SPEC-200` Nr. 1 (Handbuch-Regel).
- `pfad`: `spec/spezifikation.md` `SPEC-280` Nr. 3 („*Folge im Betrieb:* … das Handbuch nennt es“); Plan §3, Zeile E24 (Zusage: „… und dass `If-Match` bis zur Auswertung ohne Wirkung ist“); `docs/user/benutzerhandbuch.md`, Zeile `ETAG_ENABLED`.
- `befund`: Die Fix-Runde (`a5246418`) hat die Handbuch-Zeile gekürzt und dabei den Satz „Die Header `If-Match` und `If-None-Match` ändern bei `true` wie bei `false` nichts an der Antwort und am Bestand“ gestrichen (`git show a5246418 -- docs/user/benutzerhandbuch.md`). Vertrag und Zeile E24 sagen weiter, das Handbuch nenne es; die Zeile nennt es nicht. Ein Fänger gab es nie (`handbuch-check` prüft keine Prosa, `katalog-check` nur Schlüssel), und der Plan hat die Kürzung nicht in §6 eingetragen (`AGENTS.md` §3.9, §3.13). Ein Tester, der `If-Match` sendet, erfährt aus dem Handbuch nicht, dass der Mock es bis zur Auswertung ignoriert.
- Gründe, die für die Kürzung sprechen: `SPEC-200` Nr. 1 verlangt, dass ein Verhalten des Binary im Handbuch nur als gebundenes Beispiel steht; das Beispiel liegt bei `slice-handbuch-mock` (dort in §1 *Übernimmt* seit `855be54e` angenommen). Die Entscheidung gehört nicht dem Verifier.
- `Auflage an Architect/Planner` (vor der Closure, ein Commit): entweder den Halbsatz „das Handbuch nennt es“ aus `SPEC-280` Nr. 3 und aus E24 streichen und die Aussage als Teil des gebundenen Beispiels an `slice-handbuch-mock` geben (Adresse: dessen §1 und DoD tragen sie, oder die Sendung ist keine), oder den Satz in der Zeile belassen und die Regel von `SPEC-200` Nr. 1 für Aussagen zu Eingaben (Header) im Handbuch klären. Bis dahin ist die Zusage ohne Träger.

### V2 · LOW — der Plan folgt der Fix-Runde nicht an allen Stellen (`AGENTS.md` §3.9)

- `pfad`: Plan §6, Zeile *Größe*: „Messung des Implementers beim Handoff … `messung=1954` … `ok, Messung 1954 von Marke 3000`“; Stand nach der Fix-Runde ist 1999 (eigener Lauf, `make groessenmarke` im `make gates`). Plan §6, Punkt „Abweichung E24“: *Ausgang* „die Sendung ist offen, bis der Nehmer sie annimmt“; der Nehmer hat sie auf `main` angenommen (`855be54e`, `slice-handbuch-mock` §1 *Übernimmt*, „zweite Sendung, von `slice-etag-versionen`“). Plan §3, Tabelle der Mutationen: die Zeilen E3 und E13 nennen die Mutationen des Implementers, die Fix-Runde fügt F1 hinzu (steht im Absatz *Generatoren* und *Prüfgrenzen*, die Tabelle selbst nennt die Klassen-Mutationen von E1 nur dort).
- `Auflage an den Planner` (Closure): die Messung 1999 (Auslastung 0,67) und den Ausgang der Sendung E24 (angenommen) eintragen; die offenen §6-Ausgänge auf einen Ausgang bringen: Neuanlage nach `DELETE` (E20 rot gesehen: *entfallen*), Zähler unter Gleichzeitigkeit (E18, E19 mit Raten: *entfallen*), Reset und Restore (E21, E22: *entfallen*), Fakes (`entities.wachsend`, `zeigeEtag`: *entfallen*), Randformen (Reviewer hat bestätigt: *entfallen*), Größe (*entfallen*), Pfad `src/bootstrap/` (Merge-Reihenfolge: *offen* bis zum Merge), Sendung an `slice-if-match` (*offen*, Adresse Planner dieses Slice).

### V3 · INFO — Zeile E24 trägt zwei Ziele als Fänger, nur eines fängt

Siehe §4: `make handbuch-check` bleibt bei entfernter Zeile `ETAG_ENABLED` grün, `make katalog-check` ist rot. Die Spalte *Mutation* der Zeile E24 nennt allein `katalog-check` und ist damit richtig; die Spalte *Fänger* sollte `handbuch-check` nur für Kopf und Regel nennen, nicht für den Schlüssel. Kein Hindernis.

## 8. Auflagen

- **A1 (Architect/Planner, vor der Closure):** V1 entscheiden und im selben Commit Vertrag, E24 und Handbuch/`slice-handbuch-mock` angleichen.
- **A2 (Planner, bei der Closure):** V2 und V3 nachziehen; §7 Closure-Notiz mit Lerneintrag, Beobachtungs-Register, drei Paarungen (Anker, Folge-Slice, Register). Mögliche Beobachtung: eine Zusage im Vertrag („das Handbuch nennt es“) wird von einer Kürzung der Zeile unterlaufen, ohne dass ein Sensor anschlägt (Klasse „Zusage ohne Träger durch spätere Kürzung“, Zähler aus den Dateien, nicht hier gesetzt).
- **A3 (Planner/Implementer, beim Merge):** nach dem Merge von `main` (Konflikte nur in `spec/spezifikation.md`, beide Seiten behalten) `make gates` einmal wiederholen und die Messung von `make groessenmarke` aus dem Merge-Stand in den Plan übernehmen (`SPEC-128` Nr. 4 g: Zeilen Fremder in den sieben Dateien von `src/bootstrap/` zählen mit; heute bringt `main` keine).

## Grenzen dieser Verifikation

- Mutationen laufen auf einer Kopie mit eigenem Treiber (Skript im Scratchpad), nicht über `tools/harness/mutationslauf.sh` (Nullprobe-Grenze von `src/bootstrap`, Vorgabe des Aufrufers).
- Die Ratenläufe stammen aus eigenen Bauarten der Mutationen, ein Host mit Fremdlast (7 bis 8,5), je 10 bis 12 Prozesse; „12 von 12“ ist eine Größenordnung, kein fester Wert.
- Nicht gefahren: `make test-integration`, `make test-determinism`, `make gegenproben` (nicht Teil der Gate-Kette dieses Slice, Dockerfile und `harness/` unberührt).
- Das Probe-Paket aus Generatoren über die Klasse ist hier durch zehn gezielte Mutationen ersetzt, nicht durch einen eigenen Generator-Lauf; die zwei grünen Mutationen sind das Ergebnis dieser Wahl und decken sich mit der Prüfgrenze des Plans.
- Gegengelesen wurde dieser Bericht (`docs/reviews/2026-11-08-slice-etag-versionen-verifikation.md`) mit `LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}|\xEF\xBB\xBF|\xC2\xA0'`, Ergebnis: 0 Treffer, auch nach dem Schreiben.
