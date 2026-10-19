# Review-Report: slice-blob-ablage — 2026-10-19

**Review-Art:** Code — gegen Plan, Entscheidungen und Hard Rules.

**Gegenstand:** `slice-blob-ablage`, Diff `a127f5c..bc6348e` (Vertrag `SPEC-106` und Plan in `a127f5c`, Umsetzung in `bc6348e`).

**Skill:** `.harness/skills/reviewer.md` @ 424c252
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-10-19

**Eingangs-Kontext:**

- Plan `slice-blob-ablage` (inkl. Tabelle *Zusage · Fänger*, `AGENTS.md` §3.15)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [LH-FA-02](../../spec/lastenheft.md), [LH-FA-14](../../spec/lastenheft.md)
- `SPEC-106` (Kopf und Nr. 1–10 einzeln gegen den Code), `SPEC-043`, `SPEC-046`, `SPEC-053`
- `AGENTS.md` (Hard Rules §3.2, §3.6, §3.7, §3.9–§3.15)
- vorheriger Report am Nachbarmodul: `2026-10-19-slice-session-recording-persistenz-review.md`

**Eigene Läufe:** `make a-check` und `make lint` Exit 0, `make docs-check` 0 Befunde (a-check nennt nur die bekannten drei Dateien in `tools/gegenproben/`). Mutationen und explorative Tests liefen in einem Wegwerf-Klon (`git archive HEAD`) im Scratchpad, im Container `golang:1.27-alpine@sha256:8a5910f3…`, `--network none`, `GOFLAGS=-mod=mod`, `CGO_ENABLED=0`; der Arbeitsbaum blieb unberührt (`git status` sauber). `-race` war im Container nicht möglich (verlangt cgo).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Zusage „nur `sha256:` und 64 Kleinbuchstaben-Hexziffern" hat für das Präfix keinen Fänger. Mutation: `!strings.HasPrefix(ref, praefix) \|\|` entfernt, alle Tests grün. Ein Aufrufer mit `abcdefg` + 64 Hexziffern bekäme eine gültige Referenz. Der Test `ungültige Referenzen` ändert beim Fall „anderer Hash" (`sha1:`) die Länge und trifft so nur die Längenprüfung. Der Pfad wird aus den Hexziffern gebildet, daher keine Pfad-Injektion, aber `SPEC-106` Kopf („jede andere Zeichenfolge ist `ErrUngueltig`") ist gebrochen. | `AGENTS.md` §3.10, §3.15 · `SPEC-106` Kopf | `src/adapters/driven/blobs/blobs.go` · `!strings.HasPrefix(ref, praefix)` | ja, Mutation grün (selbst gefahren) | Zusage ohne rot werdende Mutation |
| F-2 | MEDIUM | „Rechte `0600`/`0700` unabhängig von der `umask`" ist nur in der Richtung `umask 0` geprüft, in der die `umask` nichts ändern kann. Mutation: beide `Chmod`-Aufrufe entfernt, Suite grün. Bei `umask 0277` liefert der Code ohne `Chmod` Blob `0400` und Verzeichnis `0500` (selbst gefahren); mit `Chmod` `0600`/`0700`. Ein Prozess mit strenger `umask` bekäme ein nicht mehr beschreibbares Verzeichnis. Die Plan-Mutation nennt nur `0644`. | `AGENTS.md` §3.10, §3.15 · `SPEC-106` Kopf | `src/adapters/driven/blobs/blobs_test.go` · `syscall.Umask(0)` | ja | Test prüft den Fall, in dem die Zusage nicht wirkt |
| F-3 | MEDIUM | Die Zusage „nie sichtbar, bevor … auf dem Datenträger gesichert: `fsync`, Umbenennen, `fsync` des Verzeichnisses" hat keinen Fänger. Mutationen: `w.f.Sync()` entfernt, grün; `d.Sync()` entfernt, grün. Die Plan-Tabelle führt keine Zeile dazu; „Ungeprüft bleibt" in `SPEC-106` nimmt die Haltbarkeit aus, nicht den Aufruf. | `AGENTS.md` §3.15, §3.11 | `src/adapters/driven/blobs/writer.go` · `if err := w.f.Sync(); err != nil` | ja | Zusage ohne Fänger |
| F-4 | MEDIUM | „Ein … Verzeichnis … lässt das Erzeugen scheitern, ohne etwas anzulegen" ist verletzt. Liegt an `<dir>/tmp` eine Datei, scheitert `New` am zweiten `MkdirAll`, hat aber `<dir>/sha256` schon angelegt (selbst gefahren: `sha256 dir=true`, `tmp dir=false`). Der Test `Erzeugen mit ungültigem Verzeichnis` blockiert nur das erste Unterverzeichnis und sieht es nicht; die Plan-Zeile behauptet „nichts angelegt". | `SPEC-106` Kopf · `AGENTS.md` §3.15 | `src/adapters/driven/blobs/blobs.go` · `for _, d := range []string{s.blobs, s.tmp}` | ja, Test mit Datei an `tmp` (selbst gefahren) | Zusage durch Test nur halb gefangen, Code bricht sie |
| F-5 | MEDIUM | Eine Nicht-Datei an einer Referenz ist für `Open` ein Fall (`ErrBlobBeschaedigt`), für `Commit` und `Remove` nicht entschieden. Selbst gefahren: `Commit` gleicher Bytes über ein Verzeichnis an der Referenz (leer oder gefüllt) endet mit rohem `rename … file exists`, `Remove` eines gefüllten Verzeichnisses mit `directory not empty`; der Blob ist nicht heilbar. `SPEC-106` Nr. 3 heilt „beschädigt", Nr. 5 zählt Nicht-Datei als beschädigt, und die Randform steht nicht im Plan §6. Der Symlink-Fall ist sauber (siehe Negativbefunde). | `AGENTS.md` §3.12 · `SPEC-106` Nr. 3/5/8 | `src/adapters/driven/blobs/writer.go` · `os.Rename(w.f.Name(), ziel)` | ja | Randform nicht entschieden |
| F-6 | MEDIUM | „Ein Verzeichnis gehört zu einem Adapter zur Zeit" (Nr. 9) hat weder einen Fänger noch eine Wirkung im Code. Selbst gefahren: ein zweites `New` auf demselben Verzeichnis löscht die temporäre Datei des offenen Writers des ersten; dessen `Commit` endet mit rohem `rename … no such file or directory`, ein typisierter Fehler fehlt. Ein Neustart, der den alten Prozess überlappt, verliert so laufende Captures. | `AGENTS.md` §3.15, §3.12 · `SPEC-106` Nr. 9 | `src/adapters/driven/blobs/blobs.go` · `func (s *Store) raeumeRestAuf` | ja | Zusage ohne Fänger |
| F-7 | LOW | `List` filtert mit `e.Type().IsRegular()`; die Mutation ohne diesen Teil bleibt grün. Der Test `List sortiert und ignoriert Fremdes` legt nur `fremd.txt` ab, kein hexadezimal benanntes Verzeichnis und keinen Symlink. Die Wirkung (Nicht-Dateien erscheinen nicht in `List`, obwohl `Open` sie als beschädigt meldet) ist Verhalten ohne Fänger. | `AGENTS.md` §3.10 | `src/adapters/driven/blobs/blobs.go` · `e.Type().IsRegular()` | ja | Zusage ohne rot werdende Mutation |
| F-8 | LOW | Die Zeile „Gleicher Inhalt, eine Datei, auch gleichzeitig" nennt als Mutation „Zähler im Dateinamen". Sie wird rot, aber in sieben Teiltests, weil der Name der Referenz nicht mehr stimmt, nicht aus dem benannten Grund (gleichzeitig). Eine Mutation, die nur die Gleichzeitigkeit trifft, fehlt: der Mutex in `Commit` entfernt bleibt grün. Der Test `gleichzeitige Commits` fährt acht Writer auf acht Dateien; jeder Writer wird nur von einem Goroutine benutzt. | `AGENTS.md` §3.10 | `src/hexagon/ports/outbound/outboundtest/blob_vertrag.go` · `for i := 0; i < 8; i++` | ja | Mutation rot aus anderem Grund |
| F-9 | LOW | Mehrere in der Fänger-Tabelle genannte Tests lösen als Name nicht auf: `Open fehlend`, `Open verändert`, `Open gekürzt`, `Open ist Verzeichnis`, `Write nach Commit`, `Commit nach Abort`, `Rechte`, `Erzeugen mit ungültigem Verzeichnis` sind Teilfälle in Sammeltests (`Open und Verify prüfen den Hash`, `genau ein Abschluss`, `TestRechteUnabhaengigVonUmask`, `TestErzeugenMitUngueltigemVerzeichnis`). Die Fälle sind vorhanden; der Verifier findet den Fänger nur über Lesen, nicht über den Namen. | `AGENTS.md` §3.15 | `docs/plan/planning/done/slice-blob-ablage.md` · `Tests \`Open fehlend\`` | nein | Fänger im Plan unter anderem Namen |
| F-10 | INFO | Plan §4 nennt die Marke „über 500 Zeilen (Summe aus `git diff --numstat`, Tests eingeschlossen)"; gemessen 895 über `src/` (`a127f5c..bc6348e`), ohne die Spec. Der Plan vermerkt die Überschreitung nicht. Dieselbe Klasse wie F-8 im Report zu `slice-session-recording-persistenz`. Der Entscheid über das Weiterführen gehört dem Planner/Architect. | `AGENTS.md` §3.14 | `docs/plan/planning/done/slice-blob-ablage.md` · `über 500 Zeilen` | ja, `git diff --numstat` | Marke überschritten, nicht vermerkt |
| F-11 | INFO | Wird der Kontext von `Begin` abgebrochen und danach weder `Write`, `Commit` noch `Abort` gerufen, bleiben temporäre Datei und Dateikennung bis zum nächsten Aufruf oder Neustart bestehen (selbst gefahren: `tmp/….tmp` vorhanden). `SPEC-106` Nr. 1 sagt „der Writer gilt als abgebrochen"; die Datei wird erst beim nächsten Aufruf gelöscht. Aufrufer müssen `Abort` rufen; der Vertrag sagt es nicht. | `SPEC-106` Nr. 1 | `src/adapters/driven/blobs/writer.go` · `ctxFehler func() error` | nein | Randform nicht entschieden |
| F-12 | INFO | Zwei Randformen, die `SPEC-106` nicht entscheidet und die an den Architect gehen (`AGENTS.md` §3.12): (a) scheitert `fsync` des Verzeichnisses nach dem Umbenennen, meldet `Commit` einen Fehler, der Blob ist aber sichtbar und der Writer abgebrochen (`SPEC-043` deckt verwaiste Blobs, nicht die Meldung); (b) `Remove` und `Commit` gleicher Bytes gleichzeitig: ein `Remove` nach dem Umbenennen löscht den Blob, den `Commit` schon meldete. Für (b) ist der Abgleich mit der Referenzmenge (Folge-Slice) der Ort. | `AGENTS.md` §3.12 | `src/adapters/driven/blobs/writer.go` · `errors.Join(d.Sync(), d.Close())` | nein | Randform nicht entschieden |
| F-13 | LOW | Der in der Tabelle genannte Fänger für „Datei erst nach `Commit` sichtbar" fängt die genannte Mutation nicht: bei Schreiben direkt auf den Zielnamen (Zufallsname im Blob-Verzeichnis) bleibt `Datei erst nach Commit sichtbar` grün, rot wird allein `TestRechteUnabhaengigVonUmask`, aus anderem Grund. Der Test würde nur bei Schreiben unter dem Namen der Referenz rot. | `AGENTS.md` §3.10, §3.15 | `src/hexagon/ports/outbound/outboundtest/blob_vertrag.go` · `offener Writer sichtbar` | ja | Mutation rot aus anderem Grund |

## Auflösung der Tabelle *Zusage · Fänger* (Plan §3)

Jede Zeile mit einer eigenen Mutation; „rot" nennt den Teiltest, aus dem sie rot wurde.

| Zeile | Mutation | Ergebnis |
|---|---|---|
| Referenzform | Längenprüfung entfernt | rot (`ungültige Referenzen`) |
|  | Großbuchstaben zugelassen | rot (`ungültige Referenzen`) |
|  | Kontext vor Referenz | rot (`ungültige Referenzen`) |
|  | Präfixprüfung entfernt | **grün** (F-1) |
| `Begin` Grenze | Grenze 0 zugelassen / Kontext zuerst | rot (`Begin mit Grenze 0 und -1`) |
| Grenze / `ErrZuGross` | `>` zu `>=` | rot (`Grenze genau erreicht …`) |
|  | Datei bleibt nach Überschreiten | rot (`Grenze genau erreicht …`) |
| `Commit` Hash und Länge | Länge +1 | rot (sieben Teiltests über `lege`, auch `Commit liefert Hash und Länge`) |
| Gleicher Inhalt, gleichzeitig | Zähler im Dateinamen | rot, aber in sieben Teiltests (F-8) |
|  | Mutex in `Commit` entfernt | **grün** (F-8) |
| `Commit` heilt | vorhandene Datei bleibt | rot (`Commit ersetzt eine beschädigte Datei`) |
| Genau ein Abschluss | Zustandsprüfung in `Write` / `Commit` entfernt | rot (`genau ein Abschluss`) |
| `Abort` | temporäre Datei bleibt | rot (`Abort räumt auf …`, `genau ein Abschluss`) |
|  | `Abort` löscht den Blob nach `Commit` | rot (`genau ein Abschluss`) |
|  | `Abort` nach `Abort` ist Fehler | rot (`Abort räumt auf …`) |
| Writer an Kontext gebunden | Kontext in `Write` / `Commit` ignoriert | rot (`Abbruch des Kontexts …`) |
| Sichtbar erst nach `Commit` | Schreiben direkt auf den Zielnamen | rot nur in `TestRechteUnabhaengigVonUmask` (liest `tmp/`); der genannte Fänger `Datei erst nach Commit sichtbar` bleibt **grün**, weil der Zufallsname keine gültige Referenz ist (F-13) |
| Rechte `0600`/`0700` | Datei `0644`, Verzeichnis `0755` | rot (`TestRechteUnabhaengigVonUmask`) |
|  | beide `Chmod` entfernt (`umask`-Unabhängigkeit) | **grün** (F-2) |
| `Open` prüft Hash | Hashprüfung entfernt | rot (`Open und Verify …`, `Commit ersetzt …`) |
|  | Fehler fehlt/beschädigt vertauscht | rot (drei Teiltests) |
|  | Nicht-Datei nicht geprüft | rot (`Open und Verify …`) |
|  | Leser nicht auf Anfang gesetzt | rot (`Open und Verify …`) |
| `Verify` wie `Open` | prüft nur Existenz | rot (`Open und Verify …`) |
| `List` | Reihenfolge umgekehrt | rot (`List sortiert …`) |
|  | Fremdes gelistet | rot (`List sortiert …`) |
|  | `IsRegular` entfernt | **grün** (F-7) |
| `Remove` | Fehler bei unbekannt | rot (`Remove, unbekannt …`) |
| Erzeugen räumt Reste | Aufräumen entfernt | rot (`TestErzeugenLoeschtReste`) |
| Verzeichnis ungültig | `New` legt Verzeichnis an | rot (`TestErzeugenMitUngueltigemVerzeichnis`) |
|  | `tmp` blockiert (Code bricht Zusage) | Test fehlt (F-4) |
| Abgebrochener Kontext | Prüfung in `Open`, `List`, `Begin` entfernt | rot (`abgebrochener Kontext …`) |
| nicht in der Tabelle: `fsync` | `Sync` / Verzeichnis-`Sync` entfernt | **grün** (F-3) |
| nicht in der Tabelle: Nr. 9 „ein Adapter" | zweites `New` | Zusage ohne Fänger (F-6) |
| Hash über `p` statt `p[:n]` | (nur bei Teilschreibung wirksam) | grün, nicht erreichbar ohne fehlschlagendes `Write` des Dateisystems; kein Finding |

## Prüfung der Fragen aus dem Auftrag

- **Nebenläufigkeit:** 200 Durchläufe `Remove` gegen `Commit` gleicher Bytes und 300 `Verify` gegen heilende `Commit`: keine Fehler, kein Leser sah eine halbe Datei (Umbenennen ist atomar, ein geöffneter Leser behält seinen Inode). Gleichzeitiges `Commit` gleicher Bytes: eine Datei. `-race` war im Container nicht lauffähig (kein cgo); die Nebenläufigkeit ist damit nur durch Wiederholung, nicht durch den Race-Detektor belegt (F-8, F-12 b).
- **Rename über Verzeichnis:** F-5.
- **Symlink:** `Open` meldet `ErrBlobBeschaedigt` (`Lstat`), `List` zeigt ihn nicht, `Commit` ersetzt den Verweis durch die geprüften Bytes, die Zieldatei außerhalb bleibt unberührt, `Remove` löscht nur den Verweis. Ohne Befund.
- **Absturz zwischen Schritten:** Temporäre Dateien liegen getrennt in `tmp/` und werden beim Erzeugen gelöscht; ein Absturz vor dem Umbenennen hinterlässt nie eine sichtbare Teildatei. Absturz nach dem Umbenennen vor dem Verzeichnis-`fsync` ist durch `SPEC-043` gedeckt. Der Fänger für Nr. 9 stellt den Rest per Hand her, nicht per Absturz (kein Befund). `fsync` selbst: F-3.
- **Pfad-Injektion über Referenzen:** Der Pfad entsteht ausschließlich aus den geprüften Hexziffern (`hexVon`, bytegenau, ohne Normalisierung); `../`, Null-Byte und Unicode scheitern vor jedem Dateizugriff. `List` baut Referenzen aus Dateinamen, die dieselbe Prüfung bestehen müssen. Ohne Befund (F-1 betrifft nur das Präfix, nicht den Pfad).
- **Zusagen im Spec-Text ohne Fänger:** F-1, F-3, F-4, F-6, F-7; Nr. 10 (keine Prüfung auf Redaction) ist eine Nicht-Zusage.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/ports/outbound/blob_repository.go`, `errors.go` | geprüft, ohne Befund (Schnittstellen und Fehler entsprechen `SPEC-106`; Kommentare tragen Zusage und Kopplung) |
| `src/hexagon/ports/outbound/outboundtest/blob_vertrag.go` | geprüft, Befunde F-1, F-2 (Vertrag trägt die `umask`-Prüfung nicht), F-7, F-8 |
| `src/adapters/driven/blobs/blobs.go` | geprüft, Befunde F-1, F-4, F-5, F-6, F-7; Pfadbildung und Symlink ohne Befund |
| `src/adapters/driven/blobs/writer.go` | geprüft, Befunde F-3, F-5, F-11, F-12; Zustandsautomat und Mutex ohne Befund |
| `src/adapters/driven/blobs/blobs_test.go` | geprüft, Befunde F-2, F-4 |
| `spec/spezifikation.md` (`SPEC-106`) | geprüft, Nr. 1–8 und 10 wie zugesagt umgesetzt; Kopf teilweise verletzt (F-1 Fänger, F-4 Code); Nr. 9 ohne Fänger (F-6) |
| Plan `done/slice-blob-ablage.md` | geprüft, Befunde F-9, F-10; Kopf, §1, §3 nachgezogen (`AGENTS.md` §3.9); Adresse in `slice-workspace-state-persistenz` (§3.13) nimmt den Abgleich an |
| `.golangci.yml`, `//nolint` | geprüft, ohne Befund (keine Suppression, keine Lockerung, `AGENTS.md` §3.2/§3.6) |
| `.a-check.yml` | geprüft, ohne Befund (Adapter importiert `model` und `outbound`, sonst Standardbibliothek; Plan-Zusage „nur Standardbibliothek" gilt) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 6 |
| LOW | 4 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Zusage ohne rot werdende Mutation · Test prüft den Fall, in dem die Zusage nicht wirkt · Zusage ohne Fänger · Randform nicht entschieden · Mutation rot aus anderem Grund · Fänger im Plan unter anderem Namen · Marke überschritten, nicht vermerkt

## Verdikt

**Merge-blockierend:** ja — kein HIGH, aber sechs MEDIUM. F-4 ist ein verifizierter Bruch von `SPEC-106` Kopf, F-5 und F-6 sind nicht entschiedene Randformen mit rohen Fehlern, F-1 bis F-3 Zusagen, deren Mutation grün bleibt (`AGENTS.md` §3.15: Zusage ohne Fänger ist ein MEDIUM). Die Klasse „Zusage ohne rot werdende Mutation" tritt zum zweiten Mal in Folge auf (zuvor F-4 im Report zu `slice-session-recording-persistenz`); beim dritten Mal greift der Steering-Loop.

**Übergabe:** Findings gehen an den Implementer; F-5, F-6, F-11, F-12 zusätzlich an den Architect (Randformen vor dem Code, `AGENTS.md` §3.12), F-10 an Planner/Architect. Die Finding-Klassen gehen in die Slice-Closure §7. Dieser Report ersetzt keine Verifikation gegen die DoD.
