# Slice slice-absturztest: Prozessende mitten in Mutation, Reset, Restore und Snapshot hinterlässt nur vorher oder nachher

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung über die DoD hinaus. Der Gegenstand (`LH-QA-03`) gehört zu `welle-mock` (in `done/`), und kein Closure-Kriterium von [welle-betrieb](../welle-betrieb.md) nennt ihn. Gebraucht wird er von M7 (Erstabnahme): die [Roadmap](../in-progress/roadmap.md) führt ihn dort als Bedingung.

**Bezug:** `LH-QA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-048` (Absturztest, Nr. 1 bis 5) · `SPEC-049` Nr. 1 (Aufbau) · `LH-QA-03.a`. Geändert wird allein der Vermerk *Stand* beim Liefern.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-absturztest`, Worktree `/Development/odata-recorder-wt-absturz`; der Zweig entsteht von `main` nach dem Start-Commit, der Start liegt auf `main`)

**Autor:** pt9912 (Planner, Schnitt von [slice-abnahme-belege](../done/slice-abnahme-belege.md) vor dem Code). **Datum:** 2026-11-27.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Test im Paket `bootstrap_test` belegt, dass ein Prozessende mit `SIGKILL` mitten in einer Mutation, einem Reset, einem Restore oder einem Snapshot nach dem Neustart nur den Stand davor oder den Stand danach hinterlässt (`LH-QA-03` Messmethode *Fehler-Injektion*, `SPEC-048`: Kindprozess, Haltepunkte an Port-Dekoratoren, 10 Haltepunkte).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Erweiterungen auf dem Record-Pfad und `ODV4-Q-11` b — [slice-abnahme-belege](../done/slice-abnahme-belege.md).
- Produktcode und ein Haken im Produkt, ein Schlaf als Orakel, `t.Skip` — der Vertrag schließt sie aus (`SPEC-048` Nr. 1); zeigt der Test eine Abweichung, greift die Rückführung in §4.
- Absturz zwischen fsync und Rename, während `COMMIT`, Stromausfall, Anhängen eines Exchange und Import, das gebaute Binary — benannte Negative des Vertrags (`SPEC-048` Nr. 5; Zeilen N1 bis N4).

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] `LH-QA-03`: der Absturztest `src/bootstrap/absturz_test.go` nach `SPEC-048` Nr. 1 bis 3 führt die 10 Haltepunkte mit der Prüfung nach jedem (T1 bis T17); er läuft in `make test` und unter `make test-race`.
- [ ] Die Mutationen nach `SPEC-048` Nr. 4 (Transaktion geteilt; Mutation außerhalb der Transaktion; `Create` des Snapshots vor dem `Commit` des Writers) sind je rot gesehen, am benannten Haltepunkt.
- [ ] Der Vermerk *Stand* von `SPEC-048` im liefernden Commit nachgezogen.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/bootstrap/absturz_test.go` | neu | Kindprozess, Port-Dekoratoren, 10 Haltepunkte, Lage je Haltepunkt, Prüfung nach dem Neustart samt Gegenproben der Prüfung (`TestAbsturzEndeOhneHaltepunkt`, `TestAbsturzEndeAuswertung`, `TestAbsturzStartBefunde`, `TestAbsturzPruefungErkennt`); Fehler-Injektion nach `LH-QA-03`. Er nutzt den Aufbau von `abnahme_aufbau_test.go`, ohne ihn zu ändern |
| `spec/spezifikation.md` | update | Vermerk *Stand* von `SPEC-048` |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Planner aus `SPEC-048`, weil der Vertrag dort steht und der Slice nur Test liefert; „Nr.“ meint `SPEC-048`; der Implementer trägt Mutationen und Raten mit Zahl der Prozesse, Last und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| T1 | Mutation (`PATCH`), Ende in der Transaktion vor `COMMIT`: Stand vorher | Haltepunkt, Nr. 2 | eine Mutation außerhalb der Transaktion | andere |
| T2 | Mutation, Ende nach `COMMIT` vor der Rückkehr: Stand nachher | Haltepunkt, Nr. 2 | die Transaktion in zwei geteilt | andere |
| T3 | Reset, Ende vor `COMMIT`: vorher | Haltepunkt | ein Teil des Reset außerhalb der Transaktion | andere |
| T4 | Reset, Ende nach `COMMIT`: nachher | Haltepunkt | Transaktion des Reset geteilt | andere |
| T5 | Restore, Ende vor `COMMIT`: vorher | Haltepunkt | ein Teil des Restore außerhalb der Transaktion | andere |
| T6 | Restore, Ende nach `COMMIT`: nachher | Haltepunkt | Transaktion des Restore geteilt | andere |
| T7 | Snapshot, Ende vor `Begin` des Blobs: ohne neuen Snapshot | Haltepunkt | `Create` vor `Begin` | andere |
| T8 | Snapshot, Ende im `Write` nach einem Teil der Bytes: ohne neuen Snapshot, `BlobsFehlend` leer | Haltepunkt | `Create` vor dem `Commit` des Writers | andere |
| T9 | Snapshot, Ende nach `Commit` des Writers vor `Create`: ohne neuen Snapshot; der verwaiste Blob ist bis zur Wartung zulässig | Haltepunkt | wie T8 | andere |
| T10 | Snapshot, Ende nach `Create` vor der Rückkehr: mit dem neuen, wiederherstellbaren Snapshot | Haltepunkt | `Create` ohne abgeschlossenen Blob | andere |
| T11 | Die Zahl der gefahrenen Haltepunkte ist das Literal 10 | Zählprüfung | ein Haltepunkt entfällt | andere |
| T12 | Der Kindprozess endet durch `SIGKILL` mit der Marke seiner Stelle; ein nicht erreichter Haltepunkt ist rot, ebenso eine Race-Meldung in seiner Ausgabe ([ADR-0007](../../adr/0007-race-detektor-im-gate.md)) | Prüfung des Exit, Nr. 3; `TestAbsturzEndeAuswertung`, `TestAbsturzEndeOhneHaltepunkt` | der Haltepunkt sendet nicht und der Kindprozess endet regulär | andere |
| T13 | `Startup` nach dem Neustart: kein Fehler, kein Befund der Datenbank, `BlobsFehlend` und `BlobsBeschaedigt` leer | nach jedem Haltepunkt | der Bericht von `Startup` wird nicht gewertet | andere |
| T14 | Entity-State, Revision und Cursors von `ws-1` und die Liste der Snapshots sind genau vorher oder genau nachher; die Orakel stammen aus demselben Aufbau ohne Absturz | Vergleich nach jedem Haltepunkt | der Vergleich lässt einen Teil (Cursors) aus | andere |
| T15 | `ws-2` ist unverändert | Vergleich | die Operation trifft beide Workspaces | andere |
| T16 | Nach `Startup` trägt die Ablage keinen unreferenzierten Blob und keine temporäre Datei | Auflisten der Ablage | die Bereinigung des Starts entfällt | andere |
| T17 | `Restore` von `vor` gelingt, und eine weitere Mutation gelingt | nach jedem Haltepunkt | eine Sperre oder ein beschädigter Blob bleibt zurück | andere |
| N1 | Negativ: Ende zwischen fsync und Rename und während `COMMIT` (kein Port-Aufruf liegt dazwischen). Folge: Atomarität von Rename und Transaktion tragen (2 Punkte ohne Fänger) | Negativ | — | andere |
| N2 | Negativ: Stromausfall, Absturz des Betriebssystems (`SIGKILL` lässt den Seiten-Cache). Folge: Dauerhaftigkeit über einen Ausfall des Hosts nicht belegt | Negativ | — | andere |
| N3 | Negativ: Anhängen eines Exchange und Import (`LH-QA-03` nennt Mutationen und Snapshots). Folge: ein Fehler nur beim Prozessende auf diesen Wegen bliebe ungefangen | Negativ | — | andere |
| N4 | Negativ: das gebaute Binary (Haltepunkt verlangte einen Haken). Folge: eine Verdrahtung ohne Startwartung fängt der Test nicht | Negativ | — | andere |

**Mutationen und Raten** (Implementer, 2026-11-28; gemessen von ihm in einem Container ohne Netz, `--cpus=4 --memory=4g`). Je Mutation eine Wegwerfkopie und ein Lauf `go test -run '^TestAbsturz' -count=1`, jede aus dem genannten Grund rot, 31 von 31. Die Haltepunkte sind Port-Aufrufe ohne Schlaf, die Mutationen sind keine Nebenläufigkeit.

| Zusage | Mutation | Roter Test (Grund) |
|---|---|---|
| T1 | `Advance` des Mock vor der Transaktion | `TestAbsturz/mutation-vor-commit` (Revision) |
| T2 | `Advance` des Mock in einer zweiten Transaktion | `…/mutation-nach-commit` (Revision), `…/mutation-vor-commit` (Lage) |
| T3, T5 | `Advance` in Reset bzw. Restore vor der Transaktion | `…/reset-vor-commit` bzw. `…/restore-vor-commit` (Revision) |
| T4, T6 | Transaktion von Reset bzw. Restore geteilt (`Advance` in der zweiten) | `…/reset-nach-commit` bzw. `…/restore-nach-commit` (Revision), je auch `vor-commit` (Lage) |
| T7 | `CreateWorkspace` vor `Begin` | `…/snapshot-vor-begin` (`BlobsFehlend`, Snapshot-Liste) und die drei anderen |
| T8 | `CreateWorkspace` zwischen `Begin` und `Write` | `…/snapshot-im-write` (`BlobsFehlend`) |
| T9 | `CreateWorkspace` zwischen `Write` und `Commit` | `…/snapshot-nach-writer-commit` (Snapshot-Liste); `im-write` bleibt grün |
| T10 | `Abort` statt `Commit`, dann `CreateWorkspace` | `…/snapshot-nach-create` (`BlobsFehlend`, Restore); alle anderen Haltepunkte mit, weil `vor` aus demselben Code entsteht |
| T11 | ein Haltepunkt aus der Tabelle entfernt | `TestAbsturz` (9 statt 10) |
| T12 | Dekorator sendet nach `COMMIT` nicht; Auswertung nimmt ein reguläres Ende an | die drei `…/…-nach-commit` (nicht erreicht); `TestAbsturzEndeOhneHaltepunkt` |
| T13 | Bericht von `Startup` nicht gewertet; `Befunde` aus der Auswertung gestrichen | `TestAbsturzPruefungErkennt` (4 Schäden); `TestAbsturzStartBefunde` |
| T14 | Vergleich ohne Cursors; ohne Snapshot-Liste | `…PruefungErkennt/Cursor_von_ws-1`; die vier `snapshot-*` (Orakel nachher gleich vorher) und `…/Snapshot_von_ws-1` |
| T15 | `ResetAll` ohne Einschränkung auf den Workspace; Vergleich ohne Stand bzw. Liste von `ws-2` | `…/reset-nach-commit` (`ws-2`); `…/Cursor_von_ws-2` bzw. `…/Snapshot_von_ws-2` |
| T16 | Abgleich in `Startup` entfernt; Aufräumen von `tmp` in `blobs.New` entfernt; Auflisten der Ablage entfernt | `…/snapshot-nach-writer-commit`; `…/snapshot-im-write`; `…PruefungErkennt/temporäre_Datei` |
| T17 | `ReferencedBlobs` ohne Workspace-Snapshots; Restore entfernt; weitere Mutation entfernt | alle 10 Haltepunkte (Restore `vor`); `…PruefungErkennt` (Blob fehlt, beschädigt, Metadaten); `…/Metadaten_fehlen` |
| Lage je Haltepunkt (`SPEC-048` Nr. 2) | Ende vor f statt danach | die drei `…-vor-commit` |
| Lage je Haltepunkt | im `Write` nichts bzw. alles geschrieben; Ende nach statt vor `Begin`; Ende vor statt nach dem `Commit` des Writers | `…/snapshot-im-write` (beide); `…/snapshot-vor-begin`; `…/snapshot-nach-writer-commit` |

Raten (je Lauf von `go test`, ein Prozess samt Kindprozessen): ohne Detektor 300 von 300 Haltepunkt-Läufen grün (`-count=30`, `TestAbsturz` 5,68 bis 6,06 s) und `-count=5` auf der letzten Fassung grün; mit Detektor 100 von 100 (`-count=10`) und 50 von 50 auf der letzten Fassung (`-count=5`), dazu `-count=10 -cpu=1` grün; keine Race-Meldung eines Prozesses (`GORACE=log_path`, keine Datei). Last des Hosts (1 min) vor und nach den Läufen: 7,49 und 3,06 bzw. 2,64 und 3,34.

Negative des Tests, je mit Grund und Folge: dass `absturzLaufe` die Lage und die Prüfung ruft, hat keinen eigenen Fänger (2 von 2 Aufrufen; Grund: Testcode ohne Test darüber; Folge: fiele ein Aufruf weg, blieben die Mutationen oben grün). Ein Befund der Datenbank wird am Aufruf nicht hergestellt (1 von 3 Befundlisten; Grund: er verlangte rohes SQL im Test; Folge: ihn trägt allein die Auswertung, `TestAbsturzStartBefunde`).

Negative der Auswertung des Prozessendes (`absturzWerte`), je mit Grund und Folge:

- In der Prüfung des Signals haben 2 von 3 Teilbedingungen keinen Fänger: `!ok` der Typzusicherung (`Sys()` ist auf Unix immer `syscall.WaitStatus`) und `!status.Signaled()` (gleichwertig: ohne Signal liefert `Signal()` -1, das nicht `SIGKILL` ist; die Mutation blieb grün). Folge: keine.
- Exit-Codes und fremde Signale sind eine große Menge, der Code vergleicht einen Wert (`SIGKILL`). Der Fänger führt 3 von 255 Exit-Codes (1; 9 als Nummer von `SIGKILL`; 137 als Schreibweise der Shell) und 2 fremde Signale (`SIGTERM`, `SIGUSR1`). Folge: eine Änderung, die genau einen anderen Wert annimmt, bliebe unbemerkt.
- Die Race-Meldung erkennt der Test am Text `WARNING: DATA RACE` in der Ausgabe (1 Text). Dass der Detektor ihn vor `SIGKILL` dorthin schreibt, hat keinen dauerhaften Fänger (1 von 1 Kette; Grund: er verlangte ein Datenrennen im Testcode, das `make test-race` rot machte); gesehen ist es mit einem eingesetzten Rennen (unten). Folge: ändert eine Go-Version den Wortlaut, oder lenkt `GORACE` des Aufrufers die Meldung mit `log_path` in eine Datei (kein Gate setzt es), bleibt die Prüfung still.

Zwei Teile der Prüfung nach dem Neustart sind doppelt gedeckt und haben keinen eigenen Fänger (Reviewer-Befund F-7; beide Mutationen blieben grün): die Bedingung auf das Verzeichnis `sha256` in `absturzAblage` (1 von 2 Teilbedingungen; Grund: kein temporärer Name gleicht einem Hash; Folge ohne sie: eine Datei außerhalb von `sha256/` unter dem Namen eines referenzierten Hashes bliebe unentdeckt) und der Zweig *der Snapshot `vor` fehlt* in `absturzWiederherstellen` (1 von 3 Meldungen; Grund: der Vergleich der Snapshot-Liste meldet dasselbe; Folge ohne ihn: keine, solange der Vergleich die Liste führt).

Hinweis (F-6): ein Lauf mit Filter auf einen Untertest (`-run '^TestAbsturz$/^mutation-vor-commit$'`) ist ohne Mutation rot, weil die Zählprüfung 10 Haltepunkte verlangt. Mutationsbelege laufen mit `-run '^TestAbsturz'`.

**Auswertung des Prozessendes** (Implementer, Fix-Runde, 2026-11-28; gemessen von ihm in einem Container ohne Netz, `--cpus=4 --memory=4g`). `absturzWerte` weist in 6 Zweigen ab; `TestAbsturzEndeAuswertung` führt 16 Enden (1 angenommen, 15 abgewiesen), die die Shell ohne Eingriff in den Kindprozess erzeugt (beim gescheiterten Start ein fehlendes Programm); `TestAbsturzEndeOhneHaltepunkt` führt dazu den echten Kindprozess mit unbekannter Operation (Exit 1). Je Mutation ein Lauf `go test -run '^TestAbsturz' -count=1`:

| Zweig (Enden) | Mutation | Roter Test |
|---|---|---|
| Race-Meldung (stderr, stdout, Exit 66) | Prüfung abgeschaltet; nur eine längere Zeile am Anfang der Ausgabe gilt | `…EndeAuswertung`, je die drei `Race-Meldung_*` |
| Frist (1) | Prüfung abgeschaltet | `…EndeAuswertung/Frist` |
| kein Exit-Fehler (Start scheitert, reguläres Ende) | Zweig abgeschaltet; reguläres Ende angenommen; gescheiterter Start angenommen | `TestAbsturzEndeOhneHaltepunkt` (Panik); `TestAbsturzEndeOhneHaltepunkt` und `…/reguläres_Ende`; `…/Start_scheitert` |
| nicht `SIGKILL` (Exit 1, 9, 137, `SIGTERM`, `SIGUSR1`) | Prüfung entfernt; nur die Signalnummer; nur das Signal ohne Exit; `SIGUSR1` zusätzlich angenommen; Exit 137 zusätzlich angenommen | die Enden des Zweigs; entfernt bzw. ohne Exit auch `TestAbsturzEndeOhneHaltepunkt` (unbekannte Operation); einzeln `…/SIGUSR1` bzw. `…/Exit_137` |
| Marke fehlt (1) | Fehler beim Lesen übergangen | `…/Marke_fehlt` |
| Marke nennt nicht die Stelle (ohne Stelle, fremd, längerer Name) | Prüfung entfernt; Präfix ohne Leerzeichen | die drei `…/Marke_*`; `…/Marke_mit_längerem_Namen`, `…/SIGKILL_mit_Marke` und die drei `…-vor-commit` |
| Race-Meldung am echten Detektor (F-2) | Datenrennen in `absturzKind` eingesetzt, `-race` | 10 von 10 Haltepunkten rot (*Datenrennen im Kindprozess*); ohne die Prüfung bleibt derselbe Lauf grün |

Raten: grün in 3 von 3 Wiederholungen (`-count=3`) ohne und mit Detektor, 5 von 5 mit Detektor (`-count=5 -v`), `TestAbsturzEnde*` 20 von 20 ohne Detektor; `TestAbsturzEndeAuswertung` braucht 1,02 s (die Frist von 1 s). Last des Hosts (1 min) vor und nach dem Container-Lauf: 3,49 und 2,71. Der Fall *Frist* hält den Prozess mit `SIGSTOP` an und enthält keinen Schlaf.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; „Nr.“ meint `SPEC-048`):

| Operation | fehlend | leer | unbekannt oder fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Mutation, Reset, Restore | Haltepunkt nicht erreicht: rot, Nr. 3 | gilt nicht: fester Ausgangsstand, Nr. 1 | fremder Workspace `ws-2` unverändert: Nr. 3 | Prozessende an den Haltepunkten, Nr. 2 | nach dem Neustart eine weitere Mutation, Nr. 3 |
| Snapshot erzeugen | wie oben | gilt nicht: der Ausgangsstand ist nicht leer, Nr. 1 | gilt nicht: keine Wahl des Namens im Test | Nr. 2 (vier Haltepunkte) | `Restore` von `vor` gelingt, Nr. 3 |
| Neustart (`Startup`) | fehlender Blob: rot, Nr. 3 | gilt nicht: `Startup` nimmt allein den Kontext, und jeder Neustart öffnet den nicht leeren Ausgangsstand, Nr. 1 | beschädigter Blob oder Befund: rot, Nr. 3 | gilt nicht: `Startup` wird nicht abgebrochen (Nr. 5 nennt keinen Fall; ein Abbruch hier ist kein Gegenstand) | gilt nicht: je Haltepunkt ein Neustart |

**Kollision und Namen** (Planner, 2026-11-28, an `main` und den Zweigen `slice-admin-abweisen-begrenzen` und `slice-container-image` samt Arbeitsbaum gelesen): `src/bootstrap/absturz_test.go` gibt es auf keinem der drei Stände. Alle Namen der Datei auf oberster Ebene beginnen mit `absturz`, `Absturz` oder `TestAbsturz` (Paket `bootstrap_test`; auf `main` und den zwei Zweigen steht keiner mit diesem Anfang; `ab` gehört der Abnahme). Der Test braucht keinen Export des Pakets `bootstrap`: er baut die Dienste aus den exportierten Konstruktoren. Braucht er doch einen, ist `export_test.go` der einzige zulässige Ort (eine eigene Datei mit `package bootstrap` weist das Lint-Profil ab, `BEO-REPO/test-export-in-eigener-datei-verletzt-testpackage`), und `slice-admin-abweisen-begrenzen` ändert diese Datei: der Implementer hält an und gibt an den Planner zurück (Block und Kollision). [slice-container-image](../in-progress/slice-container-image.md) ändert das `Dockerfile` und keinen Pfad unter `src/`; keiner der zwei Slices ändert `spec/spezifikation.md`.

**Betriebsregeln für alle Rollen:** wie in [slice-abnahme-belege](../done/slice-abnahme-belege.md) §3; im Scratchpad nur unter `absturztest-<rolle>/`.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): der Vertrag steht auf `main` (`2e289d4d`, erfüllt); ein Worktree-Platz ist frei; der Planner hat die Schnittfrage unten entschieden. Kein Kettenslice (Übergangsregeln Nr. 7). Kein Code-Vorgänger: der Slice teilt keine Datei mit [slice-abnahme-belege](../done/slice-abnahme-belege.md).

**Stand der Bedingungen** (Planner, 2026-11-28): erfüllt. `2e289d4d` ist Vorfahr von `main` (Stand `e95cc484`); der Vertrag steht mit Tabelle (T1 bis T17, N1 bis N4) und Raster ohne Platzhalter; ein Worktree-Platz ist frei (zwei von drei belegt); die Schnittfrage ist entschieden. Offene Fragen an den Architect: keine. Der Vertrag ist gegen die Bausteine gelesen: die temporäre Datei eines abgebrochenen Writers entfernt das Erzeugen der Blob-Ablage (`SPEC-106` Nr. 1), den unreferenzierten Blob die Startwartung (`SPEC-123` Nr. 7 b); Mock, Reset und Restore laufen über den Port der Transaktion, den ein Dekorator umhüllt.

**Größenmarke** (`SPEC-128` Nr. 1, `AGENTS.md` §3.14; Planner): 21 Zeilen (T1 bis T17, N1 bis N4), *a* = 0, *b* = 21, *f* = 3 (Dekoratoren um `WorkspaceTransactor`, `BlobRepository` mit seinem Writer und `SnapshotRepository`), *p* = 0: *S* = 80·21 + 600 = 2280, Marke 3100, **Schnittfrage** (*S* ≥ 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. `basis`, `schaetzung` und `marke` setzt der Planner beim Start, vor dem ersten Code-Commit. Der Teil des Architect (`pfad`, `produktcode`, `fakes`) ist vom Planner gesetzt.

**Beim Start gesetzt** (Planner, 2026-11-28, vor dem ersten Code-Commit): `basis` `dac6775a` (reiner Move nach `in-progress/`, Vorfahr von `HEAD`), `schaetzung` 2280, `marke` 3100 (135·2280 = 307800, ⌈30,78⌉ = 31). Messbasis nach `SPEC-128` Nr. 10 g: die Datei des Blocks gibt es an der Basis nicht, fremde Zeilen 0. Der Vermerk *Stand* in `spec/spezifikation.md` liegt außerhalb des Blocks und zählt nicht.

**Schnittfrage entschieden** (Planner, 2026-11-28): **kein Schnitt**, auch kein Handoff-Schnitt; der Grund steht als `schnitt-grund` im Block. Geprüft ist die Naht *Transaktion* (T1 bis T6) gegen *Snapshot* (T7 bis T10). Rechnerisch lägen beide Teile unter der Schnittgrenze (erster mit T11 bis T17, N1 bis N4 und dem einen Dekorator rund 1560, zweiter mit zwei Dekoratoren rund 720). Abtrennbar ist trotzdem keiner (`AGENTS.md` §3.14): beide teilen den Mechanismus aus Kindprozess (Rolle und Haltepunkt aus der Umgebung), Aufbau, Orakel *vorher* und *nachher* und Prüfung nach dem Neustart (T12 bis T17); der zweite Teil bestünde allein aus zwei Dekoratoren an diesem Gerüst und wäre ohne den ersten nicht lieferbar. Dazu verlangt der Vertrag eine Testfunktion mit der Zahl 10 als Literal (T11, `SPEC-048` Nr. 1 und 2): ein erster Teil führte 6 und widerspräche ihm, bis der zweite liegt. Ein Handoff-Schnitt ist aus demselben Grund nicht vorab benannt.

```groessenmarke
pfad: src/bootstrap/absturz_test.go
produktcode: 0
fakes: 3
basis: dac6775a
schaetzung: 2280
marke: 3100
schnitt-grund: Kein Teil ist ohne geteilten Mechanismus abtrennbar. Die Naht Transaktion gegen Snapshot trennt nur die Haltepunkte; Kindprozess, Aufbau, Orakel vorher und nachher und die Pruefung nach dem Neustart (T12 bis T17) gelten allen 10, und der Vertrag verlangt eine Testfunktion, deren Zaehlpruefung die Zahl 10 als Literal fuehrt (T11). Ein Teil Snapshot allein waere ohne das Geruest des anderen nicht lieferbar.
```

**Rückführungen — vorab benannt:**

- `in-progress` → `open` (blockiert): ein Haltepunkt ist ohne Haken im Produkt nicht erreichbar, oder der Test zeigt einen Zustand, der weder vorher noch nachher ist. Der Slice ändert keinen Produktcode; er gibt an den Planner und den Architect zurück.
- `in-progress` → `next` (zu groß): über der Marke vor dem Handoff und ohne vorab benannten Teil.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, Review bestanden, `make gates` grün, je Zusage eine Mutation rot gesehen (`AGENTS.md` §3.10), Closure-Notiz mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der Test findet einen echten Fehler (verwaiste Zustände, nicht wiederherstellbarer Snapshot) — **Ausgang:** weiter offen: erste Rückführung in §4, Adresse Planner und Architect. Stand beim Handoff (Implementer): kein Fund an den 10 Haltepunkten.
- Ein Kindprozess unter `make test-race` verlängert den Lauf von `bootstrap` über eine Zeitgrenze (`BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze`) — **Ausgang:** weiter offen; Adresse: Implementer (misst mit Last und Messendem). Stand (Implementer, Fix-Runde, 2026-11-28, ein Container mit 4 CPUs und 4 GiB, ein Lauf `go test -race -count=1`): das Paket `bootstrap` braucht 410 s bei einer Grenze von `go test` von 10 min je Paket (Last 6,87 vor, 4,25 nach dem Lauf); daran haben die Tests `TestAbsturz*` zusammen 16,3 s (8,46, 1,78, 1,02, 0,00 und 5,03 s); die Kindprozesse laufen mit `atexit_sleep_ms=0`. Der Reviewer maß 404 s in einem Lauf auf dem Stand `5a67a1e1` (Last 5,40 vor, 3,08 nach dem Lauf; Report 2026-11-28-slice-absturztest-review). Je ein Lauf, kein Vergleich mit dem Paket ohne den Test.
- Die Schnittfrage (§4) — **Ausgang:** entfallen (Planner, 2026-11-28: kein Schnitt, `schnitt-grund` im Block).
- `SPEC-048` Nr. 4 nennt für *`Create` vor dem `Commit` des Writers* rot am Haltepunkt *im `Write`*. Das trifft für `Create` vor `Write` (T8); zwischen `Write` und `Commit` bleibt dieser Haltepunkt grün, rot werden *nach `Commit` des Writers* und *nach `Create`* (T9; Reviewer-Befund F-4 im Report 2026-11-28-slice-absturztest-review) — **Ausgang:** offen; Adresse: Architect (Wortlaut von Nr. 4). Der Test ändert sich dadurch nicht.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, der Default `*` (Kürzel `REPO`); es gibt keine feinere.

**Vorgelagert — offene Beobachtungen sichten:** Register beim Start nachgezogen (Planner, 2026-11-28; Zahl der Dateien unter `evidence/`). Zusätzlich berührt: `BEO-REPO/test-export-in-eigener-datei-verletzt-testpackage` (1×; §3 *Kollision und Namen*), `BEO-REPO/test-nicht-wiederholbar-im-selben-container` (1×; je Lauf und Haltepunkt ein eigenes `t.TempDir`, die Raten je Lauf verlangen `-count`), `BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last` (7×; ein rotes `TestTransaktionVertrag` gehört dem Adapter), `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (33×; die Zeilen teilen ein Gerüst, die Marke liegt voraussichtlich weit über der Messung). Offen und berührt wie bei der Anlage: `BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache` (20×), `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (17×), `BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (1×), `BEO-REPO/zustand-unveraendert-nach-fehler-ohne-ruecklesen` (1×; T14 liest zurück), `BEO-REPO/orakel-prueft-nur-die-zaehler-des-testaufbaus` (1×; Orakel nach Nr. 3 aus dem Aufbau ohne Absturz).

Alle berührten Sub-Areas GF.
