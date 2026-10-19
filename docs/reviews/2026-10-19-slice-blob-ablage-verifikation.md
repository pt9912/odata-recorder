# Verifikations-Report: slice-blob-ablage — 2026-10-19

**Art:** Verifikation (Modul 11) — gegen DoD, Closure-Trigger, Spezifikation und Entscheidungen. Nicht Review, nicht Validierung.

**Gegenstand:** [slice-blob-ablage](../plan/planning/in-progress/slice-blob-ablage.md), Diff `e45f1f7` (Move nach `in-progress/`) bis `6e44210` (HEAD).
Bezug: [LH-FA-02](../../spec/lastenheft.md), [LH-FA-14](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); Vertrag `SPEC-106`; Sicht `ARC-004`, `ARC-010`.
Vorgänger: [Review-Report](2026-10-19-slice-blob-ablage-review.md) (0 HIGH, 6 MEDIUM, 4 LOW, 3 INFO).

**Rolle:** Verifier · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

## Urteil: BESTANDEN MIT AUFLAGEN

Der Code erfüllt `SPEC-106` Kopf und Nr. 1 bis 10, `ARC-004` und `ARC-010`; alle Gates sind grün.
Die Review-Befunde F-1 bis F-13 sind behoben oder in `SPEC-106` ehrlich als ungeprüft, nicht unterstützt oder unentschieden geführt; die zugehörigen Mutationen werden aus dem benannten Grund rot (selbst nachgefahren).
Kein HIGH.
Offen ist ein MEDIUM: Der Verweis (Symlink) als Nicht-Datei steht an vier Stellen im Vertrag (Kopf, Nr. 3, 5, 8), der Code verhält sich richtig, aber keine Mutation daran wird rot.
Dazu vier LOW und drei INFO.
Der Produktcode ändert sich dafür nicht; es fehlen Fänger und eine Entscheidung über zwei Randformen.

**Anzahl:** HIGH 0 · MEDIUM 1 · LOW 4 · INFO 3.

## Läufe (Verifier selbst)

| Lauf | Ergebnis |
|---|---|
| `make gates` (HEAD, Arbeitsbaum sauber) | Exit 0: a-check 0 Befunde, d-check 85 Dateien 0 Befunde, baseline-verify, build, test (`gofmt`, `go vet`, Tests), lint, schema-check (alle Fälle `ok`), migration-check ok; danach `git status` sauber |
| Mutationen | 56 im Wegwerf-Klon (`git archive HEAD` im Scratchpad), Container `golang:1.27-alpine@sha256:8a5910f3…`, `--network none`, `GOFLAGS=-mod=mod`, `CGO_ENABLED=0`; je Mutation Datei gesichert und zurückgesetzt; Arbeitsbaum unberührt |
| Sonden im Klon (als `--user 1000`, damit Rechte wirken) | schreibgeschütztes Verzeichnis mit und ohne Unterverzeichnisse; Verweis auf gültige Bytes an der Referenz (`Open`, `List`, `Commit`, `Remove`); unlesbare Datei; Unterverzeichnis nach `New` entfernt; `Write` nach `Commit` bei abgebrochenem Kontext; Leser nach `Remove`; `Write(nil)` |
| `-race` | nicht verfügbar (`CGO_ENABLED=0`, [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)); Gleichzeitigkeit stützt sich auf den Test mit acht Writern und die Mutation „Umbenennen als `Link`“ |

## DoD-Punkte 1–4 und Closure-Trigger (§5)

| Punkt | Befund |
|---|---|
| 1 Port `BlobRepository` mit `CaptureWriter` und Adapter: temporäre Datei, atomares Umbenennen, `0600`, Hash und Länge | erfüllt: `blob_repository.go`, `blobs.go`, `writer.go`; temporäre Datei in `tmp/` desselben Dateisystems, `Rename`, Rechte per `OpenFile` und `Chmod` (unter `umask 0277` Mutationen rot), `Commit` liefert Referenz und Länge |
| 2 Lesen prüft den Hash (`ODR-E3001`), Auflisten, Entfernen, gleiche Inhalte teilen eine Datei | erfüllt: `ErrBlobFehlt`/`ErrBlobBeschaedigt` (die Schicht darüber meldet `ODR-E3001`, `SPEC-018`), `List` sortiert, `Remove` idempotent, eine Datei je Inhalt auch gleichzeitig |
| 3 Tests der Randformen aus §6 gegen ein temporäres Verzeichnis | erfüllt mit Lücken (V-1 bis V-5): `TestBlobVertrag` (14 Fälle) am Port, gefahren gegen den Adapter, dazu drei Adapter-Tests; Verweis, „Leser derselben Datei“ und das nicht beschreibbare Verzeichnis ohne Fänger |
| 4 `make gates` grün | erfüllt (siehe Läufe) |
| Closure-Trigger „DoD vollständig · Gates · eine Mutation je Zusage rot · Closure-Notiz mit Lerneintrag“ | Gates: ja. Mutation je Zusage: ja, außer Verweis (V-1), „Leser derselben Datei“ (V-2) und `fsync` (im Plan und in `SPEC-106` ausdrücklich ohne Fänger). DoD vollständig: nein, erwartungsgemäß offen (Häkchen, Report-Haken, Closure-Notiz, Register, Ausgang des zweiten Risikos, Paarungen) |

## Plan-vs-Code-Diff (`e45f1f7..HEAD`)

| Plan §3 | Code | Abweichung |
|---|---|---|
| `src/hexagon/ports/outbound/` update | `blob_repository.go` (neu), `errors.go` (vier Fehler) | keine |
| `src/adapters/driven/blobs/` neu | `blobs.go`, `writer.go`, `blobs_test.go` | keine; nur Standardbibliothek plus `model` und `outbound` (a-check grün) |
| `spec/spezifikation.md` update | `SPEC-106` vor dem Code (`a127f5c`), danach geschärft (`6e44210`) | keine |
| `outboundtest/` update | `blob_vertrag.go` (`RunBlobVertrag`, `BlobAblage`) | keine |
| nicht im Plan | Roadmap („Nichts in Arbeit.“ entfernt), `harness/gegenproben.stempel`, Adresse in `slice-workspace-state-persistenz` (§3.13: nennt den Abgleich, nimmt ihn an), Review-Report | harmlos (I-3) |

§1 Ausschlüsse gewahrt: keine Redaction (der Adapter speichert Bytes, `SPEC-106` Nr. 10), kein Abgleich gegen die Referenzmenge, keine Verschlüsselung.
Weder `.golangci.yml`, `.a-check.yml`, `Dockerfile`, `go.mod` noch `Makefile` wurden berührt; kein `//nolint` (`AGENTS.md` §3.2, §3.6 gewahrt).
Marke §4: gemessen Summe aus `git diff --numstat e45f1f7..HEAD` über `src/hexagon/ports`, `src/adapters/driven/blobs`, `spec/spezifikation.md` 951, mit dem Plan 985; der Plan nennt „rund 980“ und die Überschreitung der 500 (§3.14 gewahrt, I-2).

## Konformität mit `ARC-004`, `ARC-010`, `AGENTS.md` §3

| Quelle | Urteil |
|---|---|
| `ARC-004` Outbound Ports | `BlobRepository` liegt in `src/hexagon/ports/outbound`; Fehler `ErrUngueltig` aus `model`, wie bei den Persistenz-Ports |
| `ARC-010` Bytegetreue, content-adressierte Ablage, Integritätsprüfung | erfüllt; Dateisystemzugriff nur in `src/adapters/driven/blobs` (Architektur §Regeln: „Dateisystemzugriffe für Blobs in `ARC-010`“) |
| §3.2, §3.6 | keine Suppression, keine Lockerung |
| §3.7 Kommentare | tragen Zusage, Kopplung oder Abgrenzung im Indikativ; keine Chronik; eine Überdehnung siehe V-4 |
| §3.9 Plan folgt der Korrektur | erfüllt: Kopf, §1, §3, §4, §6 und die Tabelle nach `6e44210` nachgezogen; alle Fänger-Namen der Tabelle lösen auf (17 Teiltests und Tests geprüft) |
| §3.10, §3.15 Mutation je Zusage | siehe Tabelle unten; Ausnahmen V-1, V-2 |
| §3.12 Randformen vor dem Code | Rest siehe V-3, I-1 |

## SPEC-106 Kopf und Nr. 1–10 einzeln gegen den Code

| Nr. | Zusage | Code | Urteil |
|---|---|---|---|
| Kopf: Referenzform | `hexVon`: Präfix, Länge, `[0-9a-f]`, vor dem Kontext | erfüllt; Präfix, Länge, Großbuchstaben, Kontextreihenfolge je Mutation rot |
| Kopf: Verzeichnis | `Stat`, `MkdirAll`, `Chmod`; leere Pfade scheitern an `Stat` | erfüllt für Datei, fehlend, leer, blockiertes Unterverzeichnis; „nicht beschreibbar“ weicher als der Text (V-5) |
| Kopf: nie sichtbar vor Vollständigkeit; `fsync` | `tmp/` + `Rename` + Verzeichnis-`Sync` | erfüllt; `fsync` ungefangen, im Vertrag unter *Ungeprüft* |
| Kopf: Rechte `0600`/`0700` unabhängig von `umask` | `OpenFile`, `f.Chmod`, `os.Chmod` | erfüllt; unter `umask 0277` rot (F-2 behoben) |
| Kopf: Gültigkeit vor Kontext, Kontext beendet jede Operation | `pfadVon`, `Begin` | erfüllt |
| 1 | `Begin`, Grenze ≥ 1, Writer an Kontext gebunden | erfüllt; Rest der Datei bis `Abort` oder Neustart ist im Vertrag genannt (F-11 behoben) |
| 2 | `Write` über Grenze `ErrZuGross`, Datei gelöscht, genau `grenze` zulässig | erfüllt |
| 3 | `Commit` Referenz und Länge, leer zulässig, gleiche Bytes eine Datei, heilt Beschädigtes, räumt Nicht-Datei | erfüllt; Verweis ungefangen (V-1) |
| 4 | genau ein Abschluss, `Abort` wiederholbar, lässt Commit bestehen | erfüllt |
| 5 | `Open` prüft vor dem Leser, fehlend/beschädigt/Nicht-Datei, Leser derselben Datei | erfüllt; Verweis (V-1) und „dieselbe Datei“ (V-2) ungefangen |
| 6 | `Verify` wie `Open` | erfüllt |
| 7 | `List` aufsteigend, ohne Temporäres, Fremdes, Nicht-Dateien | erfüllt (`IsRegular` jetzt gefangen, F-7 behoben); Verweis ungefangen (V-1) |
| 8 | `Remove`, Nicht-Datei, unbekannt kein Fehler | erfüllt; Verweis ungefangen (V-1) |
| 9 | Erzeugen räumt Temporäres | erfüllt (`TestErzeugenLoeschtReste`) |
| 10 | keine Redaction-Prüfung | erfüllt (Nicht-Zusage) |

## Zusage · Fänger gegen die Tests, Mutationen je Zusage

„rot“ nennt, dass der Test aus dem benannten Grund rot wurde (Meldung geprüft); „grün“ ist eine überlebende Mutation. Alle Mutationen im Wegwerf-Klon.

| Plan-Zeile | Rot gesehen (Teiltest) | Grün geblieben |
|---|---|---|
| Referenzform | Präfix entfernt · Länge entfernt · Großbuchstaben zugelassen · Kontext vor Referenz (je `ungültige Referenzen`) | — |
| `Begin` Grenze | `< 1` zu `< 0` · Kontext zuerst (`Begin mit Grenze 0 und -1`) | — |
| Grenze / `ErrZuGross` | `>` zu `>=` · Datei bleibt · Writer bleibt offen (`Grenze genau erreicht …`) | — |
| `Commit` Hash und Länge | Länge +1 (`Commit liefert Hash und Länge`, weitere über `lege`) · leerer Body abgelehnt · `List` `nil` statt leer | — |
| Gleicher Inhalt, gleichzeitig | `Link` statt `Rename` · `Commit` scheitert bei vorhandener Datei (`gleicher Inhalt teilt eine Datei, auch gleichzeitig`) | — |
| heilt / Nicht-Datei | vorhandene Datei bleibt (`Commit ersetzt eine beschädigte Datei`) · Entfernen der Nicht-Datei entfernt (`Open und Verify prüfen den Hash`) | Verweis: `Lstat` zu `Stat` in `Commit` (V-1) |
| Genau ein Abschluss | Zustandsprüfung in `Write`, in `Commit` entfernt · `Commit` setzt `committet` nicht (`genau ein Abschluss`) | — |
| `Abort` | zweiter Aufruf Fehler · löscht Blob nach `Commit` · Datei bleibt | — |
| Kontext gebunden | in `Write`, in `Commit` ignoriert · `Write` verwirft nicht (`Abbruch des Kontexts …`) | — |
| sichtbar erst nach `Commit` | temporäre Datei mit 64-Hexname im Blob-Verzeichnis: `Datei erst nach Commit sichtbar` rot (F-13 behoben), daneben `List sortiert …` und `TestRechteUnabhaengigVonUmask` | — |
| Rechte / `umask` | Datei `0644` · Verzeichnis `0755` · beide `Chmod` entfernt: unter `umask 0277` Blob `0400`, Verzeichnis `0500` (`TestRechteUnabhaengigVonUmask`) | — |
| `Open` prüft Hash | Hashprüfung entfernt · Dateiprüfung entfernt · Fehler vertauscht · Leser nicht am Anfang (`Open und Verify prüfen den Hash`) | Zweitöffnung statt derselben Datei (V-2); Verweis: `Lstat` zu `Stat` in `Open` (V-1) |
| `Verify` | ignoriert `ErrBlobBeschaedigt` | — |
| `List` | umgekehrt · Fremdes · `IsRegular` entfernt (`List sortiert und ignoriert Fremdes`) | Verweis als Datei gelistet (V-1) |
| `Remove` | Fehler bei unbekannt · ohne `RemoveAll` (gefülltes Verzeichnis) | Verweis folgen und Ziel löschen (V-1) |
| Erzeugen räumt Reste | Aufräumen entfernt (`TestErzeugenLoeschtReste`) | — |
| Verzeichnis ungültig | `New` legt Verzeichnis an (`TestErzeugenMitUngueltigemVerzeichnis`) | — (Sonde V-5) |
| Abgebrochener Kontext | Prüfung in `Begin`, `List`, `pfadVon` entfernt | — |
| `fsync` | — | Datei-`Sync` entfernt, Verzeichnis-`Sync` (Plan und Vertrag: kein Fänger, ehrlich geführt) |
| Hash über `p` statt `p[:n]` | — | grün, nur bei fehlschlagendem Teil-`Write` des Dateisystems erreichbar (kein Befund) |

## Findings

| ID | Stufe | Befund | Quelle | Pfad | Verifizierbar |
|---|---|---|---|---|---|
| V-1 | MEDIUM | Der Verweis (Symlink) als Nicht-Datei ist Teil des Vertrags (`SPEC-106` Kopf/Nr. 3 „einen Verweis selbst und nie sein Ziel“, Nr. 5 `ErrBlobBeschaedigt`, Nr. 7, Nr. 8), hat aber keinen Fänger: `BlobAblage` bietet nur `ErsetzeDurchVerzeichnis`, und die Plan-Zeilen führen nur „Verzeichnis“. Mutationen grün: `Lstat` zu `Stat` in `Open` (ein Verweis auf Bytes mit passendem Hash würde gelesen), `List` zählt Verweise, `Lstat` zu `Stat` in `Commit` (der Verweis bliebe), `Remove` folgt dem Verweis und löscht das Ziel. Gelesen und per Sonde bestätigt: der Code ist richtig (`Open` `Blob beschädigt`, `List` leer, `Commit` ersetzt durch eine Datei und lässt das Ziel, `Remove` lässt das Ziel). Review F-5 behob die Nicht-Datei für Verzeichnisse, nicht für Verweise. | `AGENTS.md` §3.15, §3.10 · `SPEC-106` Kopf, Nr. 3/5/7/8 | `src/hexagon/ports/outbound/outboundtest/blob_vertrag.go` · `BlobAblage`; `src/adapters/driven/blobs/blobs_test.go` · `neueAblage` | ja, vier Mutationen (`os.Lstat(pfad)` zu `os.Stat`, `os.Lstat(ziel)` zu `os.Stat`, `IsRegular` um Symlink erweitern, `EvalSymlinks` vor `RemoveAll`) |
| V-2 | LOW | „liefert Leser derselben geöffneten Datei“ (`SPEC-106` Nr. 5) hat keinen Fänger: `Open` gibt die Datei nach der Prüfung frei und öffnet sie neu (`_ = f.Close(); return os.Open(pfad)`), alle Tests grün. Die Zusage schließt die Lücke zwischen Prüfung und Lesen; ohne Hook im Test schwer zu fangen, aber der Plan führt sie nicht und benennt sie nicht als ungeprüft. | `AGENTS.md` §3.15, §3.11 | `src/adapters/driven/blobs/blobs.go` · `pruefeDatei` | ja, Mutation |
| V-3 | LOW | Zwei Randformen sind im Vertrag ausdrücklich nicht entschieden: `Remove` gegen `Commit` gleicher Bytes („Unentschieden bleibt …“) und die Nutzung eines Writers aus mehreren Goroutinen (der Mutex entfiel, der Plan sagt „ein Writer gehört einer Goroutine“, Vertrag und Port-Kommentar nicht). Plan §6 führt den Ausgang der Randformen als „entfallen: entschieden in `SPEC-106`“ und trifft das für diese beiden nicht. Der Vertrag trägt sie ehrlich als Rest; entscheiden gehört dem Architect (`AGENTS.md` §3.12), nicht dem Code. | `AGENTS.md` §3.12, §3.13 | `spec/spezifikation.md` · `SPEC-106` Schluss; `docs/plan/planning/in-progress/slice-blob-ablage.md` · §6 | ja |
| V-4 | LOW | Der Kopf von `RunBlobVertrag` sagt zu: „prüft SPEC-106 Nr. 1 bis 8 und den Kopf des Vertrags“. Der Kopf enthält Rechte, `umask`, Verzeichnis-Anforderungen und `fsync`; sie prüfen `blobs_test.go` und kein Vertragsfall, und der Verweis-Anteil von Nr. 3/5/7/8 wird gar nicht geprüft (V-1). Ein zweiter Adapter, der `RunBlobVertrag` fährt, bekäme damit weniger Prüfung, als der Kommentar sagt. | `AGENTS.md` §3.11 | `src/hexagon/ports/outbound/outboundtest/blob_vertrag.go` · `RunBlobVertrag` | ja |
| V-5 | LOW | „ein nicht beschreibbares Verzeichnis lässt das Erzeugen scheitern“ (`SPEC-106` Kopf) gilt nur, wenn ein Unterverzeichnis neu anzulegen ist. Sonde als Nicht-Root: ist das Datenverzeichnis `0500`, aber `sha256/` und `tmp/` bestehen, gelingt `New`, `Begin` und `Commit` laufen durch (die Unterverzeichnisse sind beschreibbar). Der Text ist weiter als das Verhalten; der Test ersetzt die Rechte durch eine Datei an Stelle des Unterverzeichnisses und kann als Root keine Rechte-Verweigerung provozieren. Der Plan benennt „Unterverzeichnis nicht anlegbar“ ehrlich. | `AGENTS.md` §3.11, §3.12 | `spec/spezifikation.md` · `SPEC-106` Kopf; `src/adapters/driven/blobs/blobs_test.go` · `TestErzeugenMitUngueltigemVerzeichnis` | ja, Sonde (`--user 1000`) |
| I-1 | INFO | Randformen, die `SPEC-106` nicht nennt (für den Architect, `AGENTS.md` §3.12): eine unlesbare Blob-Datei (`EACCES`) liefert von `Open` den rohen Fehler des Betriebssystems, weder `ErrBlobFehlt` noch `ErrBlobBeschaedigt`; ein Unterverzeichnis, das nach `New` verschwindet, lässt `Begin` und `List` mit rohen Fehlern enden (`Open` meldet dann `ErrBlobFehlt`); bei `Write` nach `Commit` mit abgebrochenem Kontext gewinnt `ErrAbgeschlossen` vor dem Fehler des Kontexts (Reihenfolge Zustand/Kontext nicht entschieden). Keine Zusage gebrochen. | `AGENTS.md` §3.12 | `src/adapters/driven/blobs/blobs.go` · `pruefeDatei`; `writer.go` · `Write` | ja, Sonden |
| I-2 | INFO | Der Entscheid, den Slice trotz überschrittener Marke (§4: „über 500“, gemessen 951 bis 985) ungeteilt zu lassen, steht im Plan, geschrieben vom Implementer; die Zahl „rund 980“ stimmt mit der Messung über Pfade samt Plan überein. Der Entscheid gehört dem Planner (wie V-8 im Report zu `slice-session-recording-persistenz`). | `AGENTS.md` §3.14 | `docs/plan/planning/in-progress/slice-blob-ablage.md` · §4 | ja, `git diff --numstat` |
| I-3 | INFO | Änderungen außerhalb von Plan §3: `docs/plan/planning/in-progress/roadmap.md` (Zeile „Nichts in Arbeit.“ entfernt), `harness/gegenproben.stempel`, der Link in `slice-workspace-state-persistenz` (Move-Nachzug). Harmlos; die Roadmap-Änderung gehört zum Lifecycle-Übergang. | Plan-vs-Code | siehe Pfade | ja, `git diff --stat` |

## Review-Befunde: Nachweis der Behebung

| Review | Zustand |
|---|---|
| F-1 (Präfix) | behoben, selbst nachgefahren: `HasPrefix(ref, "")` rot in `ungültige Referenzen` (Fall `sha257:`) |
| F-2 (`umask`) | behoben, selbst nachgefahren: Test mit `umask 0` und `0277`; beide `Chmod`-Entfernungen rot (`Verzeichnis sha256 500`, `temporäre Datei 400`) |
| F-3 (`fsync`) | als Zusage zurückgenommen: `SPEC-106` führt es unter *Ungeprüft*, der Plan „kein Fänger“; beide Mutationen grün, wie dort gesagt (§3.11 gewahrt) |
| F-4 (Erzeugen legt an) | durch den Vertrag behoben: „ein davor angelegtes leeres Unterverzeichnis bleibt“; Fall `tmp blockiert` im Test; der Code blieb, der Text folgt dem Verhalten |
| F-5 (Nicht-Datei) | behoben für Verzeichnisse (`Commit` entfernt, `Remove` mit `RemoveAll`, Mutationen rot); Verweis ungefangen (V-1) |
| F-6 (ein Adapter je Verzeichnis) | als „Nicht unterstützt“ in den Vertrag genommen; Kommentar an `Store` entspricht dem; kein Fänger nötig |
| F-7 (`IsRegular`) | behoben: `LegeFremdesAb` legt ein hexförmig benanntes Verzeichnis an; Mutation rot |
| F-8 (Mutex, Gleichzeitigkeit) | behoben: Mutex entfällt, Plan-Mutation „`Link` statt `Rename`“ rot in `gleicher Inhalt teilt eine Datei, auch gleichzeitig`; Nutzung eines Writers aus mehreren Goroutinen offen (V-3) |
| F-9 (Fänger-Namen) | behoben: die Tabelle nennt Teiltests, die auflösen |
| F-10 (Marke) | im Plan vermerkt (I-2) |
| F-11 | behoben: in `SPEC-106` Nr. 1 entschieden |
| F-12 | (a) entschieden in `SPEC-106` Kopf; (b) als „unentschieden“ geführt (V-3) |
| F-13 | behoben, selbst nachgefahren: Name der Datei im Blob-Verzeichnis (64 Hexziffern) macht `Datei erst nach Commit sichtbar` rot |

## Offen und erwartet (kein Befund)

Häkchen der DoD, Closure-Notiz mit Steering-Loop-Lerneintrag, Beobachtungs-Register (die Klassen „Zusage ohne rot werdende Mutation“ und „Randform nicht entschieden“ treten mit diesem Slice erneut auf; das Register zählt, der Anker folgt der Welle-Closure), Ausgang des zweiten Risikos in §6 (steht „weiter offen“), die drei Paarungen, das Häkchen für den Review-Report.

## Übergabe

Der Implementer schließt V-1 (Verweis-Hilfe `ErsetzeDurchVerweis` in `BlobAblage`, Fänger je Operation, Plan-Zeile ergänzen) und nimmt V-2, V-4, V-5 mit (Fänger oder Eintrag unter *Ungeprüft*; Kommentare auf das Geprüfte kürzen).
V-3 und I-1 gehen an den Architect (Randformen, `AGENTS.md` §3.12), I-2 an den Planner.
Danach genügt eine erneute Verifikation der betroffenen Zeilen.
