# Slice slice-token-datei-wettlauf: Die Token-Datei lädt eine Datei, die ein gleichzeitiger Aufruf zwischen `Stat` und `Lstat` anlegt (Fix zu `slice-admin-listener-token`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md) (Fix an einem Teil der Welle, `slice-admin-listener-token`; die Slice-Liste und das Drift-Log der Welle pflegt der Planner beim Start).

**Bezug:** `LH-FA-13` (Buchstabe a: Token-Datei), `LH-QA-06` (gleiche Konfiguration und gleiche Eingabe führen zu gleichem Testverhalten: der Test ist heute ein Flake), [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-239` Nr. 1 a (Entscheidung nach dem fehlenden `Stat`), Nr. 3 d (Verlierer des `Link`) und Nr. 4 (Fenster der Nebenläufigkeit) mit *Ungeprüft* · `SPEC-110` Nr. 13 (synchron, ohne Kontext; nicht geändert).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-token-datei-wettlauf`, Worktree `/Development/odata-recorder-wt-token-datei-wettlauf`; der Zweig entsteht von `main` nach dem Start-Commit, der Implementer nimmt `main` vor dem ersten Code-Commit per `git merge main` auf)

**Autor:** pt9912 (Architect, aus dem Register `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel`). **Datum:** 2026-11-23.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem** Abschnitt des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `admin.LadeToken` meldet nie `keine reguläre Datei` für eine Datei, die ein gleichzeitiger Aufruf zwischen dem `Stat` und dem `Lstat` des Verlierers anlegt: findet `Lstat` nach einem fehlenden `Stat` einen Eintrag, behandelt der Aufruf ihn wie einen, der schon beim ersten `Stat` da war (er erzeugt nichts, lädt ihn mit neuem `Stat` und den Schritten (b) bis (g), `SPEC-239` Nr. 1 a), statt ihn über `statFehler` auf `keine reguläre Datei` abzubilden. Ein neuer Haltepunkt `nachStat` führt das Fenster deterministisch; `TestWettlaufNebenlaeufig` bleibt unverändert und ist danach in jedem Lauf grün (heute färbt er `pruefung test` bei `--cpus=4` in 3 von 4 Läufen rot und macht jedes Gate unzuverlässig, `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel`).

**Vertrag (Architect, 2026-11-23):** `SPEC-239` Nr. 1 a und *Ungeprüft* sind im selben Commit wie dieser Plan geschärft (Historie der Spezifikation, 2026-11-23); der Planner bringt den Commit nach `main`, bevor er den Slice startet, der Vertrag steht damit vor dem ersten Code-Commit (`AGENTS.md` §3.12). Die Entscheidung (der Verlierer lädt die Datei des Gewinners und meldet nie `keine reguläre Datei` für eine reguläre Datei) trägt `SPEC-239` Nr. 1 a und Nr. 4; eine Fallunterscheidung Symlink gegen Nicht-Symlink gibt es nicht (§3, *Entscheidung der Form*).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Alles außer `src/adapters/driving/admin/tokendatei.go` und den Tests dazu (`tokendatei*_test.go`): der Wettlauf liegt allein in `lade`; `admin.go`, `lebenszyklus.go`, `export_test.go`, `bindung_test.go` und die Dateien der anderen Admin-Slices im Paket (`imports*`, `sessions*`, `workspaces*`, `inspektion*`) bleiben unberührt.
- Ein neuer Fehlercode, Grund, Schlüssel oder Kommando — `SPEC-239` Nr. 6 führt sieben Gründe, ihr Text ist Vertrag, und der Fix fügt keinen hinzu; das Binary bleibt unverändert, ein Handbuch-Kapitel entfällt (das Kataloge-Gate bleibt grün, ohne dass der Plan es berührt).
- Eine Lockerung des Last-Tests (weniger Goroutinen oder Läufe, eine Wiederholung bei Rot, `t.Skip`): `TestWettlaufNebenlaeufig` bleibt unverändert und ist das Ziel, nicht der Weg; ein Gate wird nicht gelockert (`AGENTS.md` §3.6).
- Der Wegfall des `Lstat` (jeder fehlende `Stat` ginge in `erzeuge`, der `Link` meldete `fs.ErrExist`): verworfen, weil `SPEC-239` Nr. 1 a und `TestSymlinkOhneZielOhneHilfsdatei` (T31) zusagen, dass ein Symlink ohne Ziel keine Hilfsdatei mit dem Token anlegt; der Fix hält diese verifizierte Zusage und schließt allein das Fenster.
- Handlungen anderer Prozesse und die Atomarität von `Link` zwischen Prozessen — `SPEC-239` *Ungeprüft* (zwei Prozesse auf derselben Datei, Wechsel zwischen `Stat` und dem Lesen, Datei nach dem `Lstat` gelöscht): der Test läuft in einem Prozess, und ein Haltepunkt zwischen `Lstat` und dem neuen `Stat` kostet ein weiteres Feld für einen Ausgang, den der Ausgang des ersten Fehlers schon führt (TW7).
- Windows (`SPEC-239` Nr. 2): das Image läuft auf Linux, der Fix ändert keine Plattformregel.

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

- [ ] Fix und Fänger: `lade` lädt nach einem fehlenden `Stat` und einem gelingenden `Lstat` über `ladeVorhandene` (neues `Stat`, Schritte (b) bis (g)) und erzeugt nichts; der Haltepunkt `nachStat` zwischen `Stat` und `Lstat` ist die Naht des Tests; die Zeilen TW1 bis TW6 der Tabelle in §3 haben je ihren Fänger, die Menge TW2 (11 Einträge × 3 Zugänge = 33 Zellen) führt jede Zelle mit Orakel als Literal, und je Zeile ist die Mutation vom Implementer selbst rot gesehen (`AGENTS.md` §3.10, §3.15); `SPEC-239` ist mit diesem Plan geschärft (Architect) und der Quelltext folgt ihr.
- [ ] Der Flake ist behoben, belegt in jedem Lauf: `TestWettlaufNebenlaeufig` ist grün in jedem Lauf der Messung (kein „meist“): mindestens fünf frische Läufe von `pruefung test` (je ein frischer Container, `--cpus=4`, unter der Sperre) und mindestens 30 Prozesse `go test -count=1 -run '^TestWettlaufNebenlaeufig$'` im selben Container-Zuschnitt; der Bericht nennt je Messung Messenden, Last des Hosts vor und nach jedem Lauf, Zahl der Prozessoren und n von m (Stand heute: 3 von 4 `pruefung test`-Läufen rot, Messender Orchestrator, 2026-11-23). Ein Rot in einem Lauf heißt: der Fix gilt nicht, die Zeile geht an den Architect (§6). Die Messung allein beweist den Fix nicht (30 von 30 grün schließt eine Rate unter 10 % nicht aus); der Beweis ist TW1, in jedem Lauf rot ohne den Fix.
- [ ] Doku und Gate: das Binary und das Handbuch bleiben unverändert (§7 hält es fest); `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben: `evidence/slice-token-datei-wettlauf.md` in `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel/` und die Stand-Zeile dort auf den Beleg gesetzt; **kein Zähler wird gesetzt**, er folgt aus den Dateien.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/adapters/driving/admin/tokendatei.go` | update | `lade`: nach `Stat` mit `fs.ErrNotExist` ruft es `Lstat`; scheitert der, `erzeuge`; gelingt er, `ladeVorhandene` (kein `statFehler` mehr für diesen Fall); jeder andere Fehler von `Stat` ist `Datei nicht lesbar`; `istFehlend` entfällt (die Entscheidung steht in `lade`); `haltepunkte` bekommt `nachStat func()` (Aufruf nach dem fehlenden `Stat`, vor `Lstat`), `LadeToken` übergibt keinen; kein neuer Grund, keine andere Signatur |
| `src/adapters/driving/admin/tokendatei_halt_test.go` | neu (Paket `admin`) | Test-Export `LadeMitStatHalt(pfad, nachStat)` ruft `lade` mit dem neuen Haltepunkt; `export_test.go` bleibt unberührt (es wird von anderen Slices im Paket mitgeschrieben): `LadeMitHaltepunkten` setzt `nachStat` nicht, der Wert `nil` heißt darum kein Halt; ändert der Implementer doch `export_test.go`, nennt er die Datei im selben Commit im Block (§4) |
| `src/adapters/driving/admin/tokendatei_fenster_test.go` | neu (`//go:build unix`) | TW1 bis TW3 und TW5: Fenster, Menge Eintrag × Zugang, Unverändertheit; Orakel als Literal (Gründe `gKeineDatei` usw. aus `tokendatei_test.go`, nie aus dem Code); FIFO und Schleife laufen unter `mitFrist` |
| `src/adapters/driving/admin/tokendatei_ausgabe_test.go` | update | TW6: der Weg im Fenster (zwei Fälle) in `TestKeineAusgabe`, Zählstelle 41 → 43 |
| `src/adapters/driving/admin/tokendatei_test.go`, `src/adapters/driving/admin/tokendatei_zusagen_test.go` | nur falls nötig | Helfer wiederverwendet (`schnappschuss`, `namen`, `mitFrist`, `erwarteGrund`); `TestWettlaufNebenlaeufig`, `TestWettlaufDeterministisch`, `TestZweiterAufruf` und `TestSymlinkOhneZielOhneHilfsdatei` bleiben unverändert |
| `spec/spezifikation.md` | geliefert (Architect, mit diesem Plan) | `SPEC-239` Nr. 1 a, *Ungeprüft*, Historie; dieser Slice schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidung der Form** (Architect, 2026-11-23; Ort: `SPEC-239` Nr. 1 a). Ein Eintrag, den `Lstat` nach einem fehlenden `Stat` findet, geht in `ladeVorhandene`, gleich welcher Art er ist; ein Symlink ohne Ziel bleibt `keine reguläre Datei`, weil das neue `Stat` wieder „fehlt“ meldet, `Lstat` gelingt und `statFehler` den Grund bildet. *Alternative:* `lade` unterscheidet am Modus von `Lstat`: ein Symlink bleibt `keine reguläre Datei`, jeder andere Eintrag geht in `ladeVorhandene`. *Verworfen:* beide Formen geben für jeden Eintrag denselben Ausgang außer für einen Symlink, dessen Ziel erst im Fenster entsteht (`keine reguläre Datei` statt des Tokens); diesen Fall hatte `SPEC-239` *Ungeprüft* ausgenommen, ein Zweig, den nur ein ungeprüfter Fall unterscheidet, ist eine Zusage ohne Fänger (`AGENTS.md` §3.15, `BEO-REPO/aequivalente-mutation-hinter-doppeltem-waechter`), und der Ausgang hinge vom Zeitpunkt ab, in dem der Symlink entsteht, ohne Gewinn für den Betreiber. TW2 prüft die Gleichheit „Eintrag im Fenster gleich Eintrag vor dem ersten `Stat`“ je Eintrag über alle drei Zugänge. *Folge im Betrieb:* ein Symlink auf ein Secret, der genau zwischen den zwei Aufrufen entsteht, wird geladen statt abgewiesen.

**Fenster der Anlage** (`SPEC-239` Nr. 4: ein Aufruf sieht die Datei eines anderen zuerst vor dem `Stat`, zwischen `Stat` und `Lstat` oder beim `Link`; jedes Fenster hat einen Haltepunkt im Test und einen Fall):

| Fenster | Haltepunkt | Fall | Fänger |
|---|---|---|---|
| vor dem `Stat` | keiner nötig: die Datei liegt vor dem Aufruf | zwei Aufrufe nacheinander; jeder Eintrag der Menge vor dem Aufruf | `TestZweiterAufruf` (T22), TW2 Zugang 1 |
| zwischen `Stat` und `Lstat` | `nachStat` (neu) | der Haltepunkt ruft einen zweiten `LadeToken` zu Ende (legt an), entfernt einen Symlink ohne Ziel (löscht) oder tut nichts | TW1, TW2 Zugang 2 |
| beim `Link` (zwischen Anlegen der Hilfsdatei und `Link`, auch vor dem Schreiben) | `vorLink` (Bestand), `nachAnlegen` (Bestand) | der Haltepunkt ruft einen zweiten `LadeToken` zu Ende; jeder Eintrag der Menge wird angelegt | `TestWettlaufDeterministisch` (T18), TW2 Zugang 3 |
| nach dem `Link` | keiner: nach `Link` ist die Datei vollständig (`SPEC-239` Nr. 3, *Entscheidung*); ein Aufruf, der danach beginnt, ist der zweite Aufruf von Nr. 4 und sein Fenster ist „vor dem `Stat`“ | siehe erste Zeile | T22 |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-23; jede Zeile trägt ihre Menge; Kennung `TW` = Token-Datei-Wettlauf):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| TW1 | (`SPEC-239` Nr. 1 a, Nr. 4) Legt ein gleichzeitiger Aufruf die Datei zwischen `Stat` und `Lstat` an, erzeugt der Verlierer nichts und liefert deren Token; er meldet nie `keine reguläre Datei`. Entfernt der Haltepunkt dort einen Symlink ohne Ziel, erzeugt der Aufruf die Datei (`Lstat` meldet „fehlt“) | Menge Aktion im Fenster = legt an (ein innerer `LadeToken` läuft zu Ende), löscht (der Symlink ohne Ziel, der beim `Stat` lag, ist weg), nichts: 3 Zellen mit `LadeMitStatHalt`. Orakel als Literal: legt an: die Rückgabe außen gleicht der innen, 64 Zeichen `0-9a-f`, die Datei trägt sie, `0600`, im Verzeichnis allein `admin.token`, `nachAnlegen` und `vorLink` des äußeren Aufrufs rufen `t.Error` (er legt keine Hilfsdatei an und versucht keinen `Link`); löscht: Token mit 64 Zeichen, Datei `0600`; nichts: wie der Aufruf ohne Haltepunkt (T14). Der Haltepunkt wird in jeder Zelle genau einmal gerufen | altes Verhalten (nach `Lstat`-Erfolg `statFehler`): rot in der Zelle legt an (`keine reguläre Datei`), in jedem Lauf; der Verlierer fällt nach `Lstat` in `erzeuge` (`nachAnlegen` wird gerufen, der `Link` meldet `fs.ErrExist`, das Token stimmt, nur der Haltepunkt sieht es); ein Fehler von `Lstat` als Eintrag gelesen (Zelle löscht: `Datei nicht lesbar`); der Haltepunkt zweimal gerufen oder gar nicht | nebenläufig |
| TW2 | (`SPEC-239` Nr. 1 a, b bis g, Nr. 6) Ein Eintrag, den `Lstat` nach einem fehlenden `Stat` findet, hat den Ausgang, den derselbe Eintrag vor dem ersten `Stat` hätte: der Aufruf erzeugt nichts und lädt ihn mit neuem `Stat` und den Schritten (b) bis (g); der Ausgang ist je Eintrag derselbe in jedem Zugang | Menge Eintrag × Zugang. Eintrag (11): e1 gültige Datei `0600` (Token), e2 Verzeichnis (`ist ein Verzeichnis`), e3 FIFO (`keine reguläre Datei`, ohne zu blockieren), e4 Datei `0644` mit gültigem Inhalt (`Gruppen- oder Weltrechte`), e5 Datei `0600` mit zwei Zeilen (`mehr als eine Zeile`), e6 leere Datei `0600` (`weniger als 32 Zeichen`), e7 Datei `0644` mit zwei Zeilen (`Gruppen- oder Weltrechte`, das Paar Rechte × Zeilen: kleinster Rang), e8 Symlink auf eine gültige Datei `0600` (Token), e9 Symlink ohne Ziel (`keine reguläre Datei`), e10 Symlink auf ein Verzeichnis (`ist ein Verzeichnis`), e11 Schleife von Symlinks (`Datei nicht lesbar`). Zugang (3): vor dem `Stat` (der Eintrag liegt vor dem Aufruf; der Haltepunkt `nachStat` wird genau für e9 gerufen, sonst nie), zwischen `Stat` und `Lstat` (der Haltepunkt legt den Eintrag an), beim `Link` (`vorLink` legt den Eintrag an). 11 × 3 = 33 Zellen; Orakel das Literal je Eintrag, gleich über die drei Zugänge; je Zelle zusätzlich: keine Hilfsdatei bleibt, im Verzeichnis liegen allein der Eintrag und seine Hilfsdateien des Tests. Zählprüfung: 11 Einträge, 3 Zugänge, 33 gelaufene Zellen; die sechs Ausgänge Token, `ist ein Verzeichnis`, `keine reguläre Datei`, `Gruppen- oder Weltrechte`, `mehr als eine Zeile`, `weniger als 32 Zeichen`, `Datei nicht lesbar` kommen je mindestens zweimal vor (die Gründe 1 e und `Datei nicht anlegbar` sind über einen Eintrag nicht erreichbar, `SPEC-239` *Ungeprüft*) | altes Verhalten: rot in Zugang 2 für e1 bis e8, e10 und e11 (e9 bleibt grün: derselbe Grund); `ladeVorhandene` nur bei regulärer Datei (sonst `keine reguläre Datei`): rot für e2, e8, e10, e11; die alte Unterscheidung (Symlink nach `Lstat` gibt `keine reguläre Datei`, sonst laden): rot für e8, e10, e11; `pruefeUndLies` mit der Auskunft von `Lstat` statt eines neuen `Stat`: rot für e8, e10; der Haltepunkt `nachStat` auch nach einem gelingenden `Stat` gerufen (Zugang 1, e1: der Aufruf zählt 1 statt 0): rot; Elemente, die der Befund nicht nannte: Verzeichnis, FIFO, die Paare e7 und e11 | andere |
| TW3 | (`SPEC-239` Nr. 4 Satz 3) Der Aufruf ändert, kürzt, löscht und korrigiert den Eintrag im Fenster nie, auch bei einem Verstoß nicht | Menge Eintrag (e1 bis e11) × Zugang (zwischen `Stat` und `Lstat`, beim `Link`) = 22 Zellen: Art, Modus, Größe und Bytes des Eintrags (der Symlink: sein Ziel) vor und nach dem Aufruf gleich (`schnappschuss`); die Datei `0644` bleibt `0644`, die leere leer | `Chmod` auf `0600` bei einem Rechteverstoß in den neuen Zweig (rot für e4 und e7, die Fixture `0644` gleicht nie dem Wert `0600` der Mutation); `os.Remove` des Eintrags bei einem Verstoß (rot für e4 bis e7); Schreiben eines Bytes in die Datei bei `weniger als 32 Zeichen` (rot für e6; ein Kürzen bliebe an der leeren Datei ohne Wirkung und wäre keine Mutation) | andere |
| TW4 | (`SPEC-239` Nr. 3 d, Nr. 4) Bestand, der bleiben muss: von zwei Aufrufen gewinnt einer, beide liefern dasselbe Token, im Verzeichnis bleibt allein `pfad` (Fenster beim `Link`); ein zweiter Aufruf nach dem Erzeugen liefert dasselbe Token und lässt die Datei unverändert (Fenster vor dem `Stat` und nach dem `Link`) | `TestWettlaufDeterministisch` (T18), `TestZweiterAufruf` (T22), `TestHilfsdateiEntfernt` (T20), unverändert; sie laufen mit dem Fix und werden nicht umgeschrieben | `os.Rename` statt `os.Link` (T18, vom Implementer einmal erneut rot zu sehen, weil der Fix `lade` umbaut); Datei neu geschrieben beim zweiten Aufruf (T22) | nebenläufig |
| TW5 | (`SPEC-239` Nr. 4) Gleichzeitige Aufrufe auf eine fehlende Datei erzeugen genau eine Datei, jede Rückgabe ist deren Inhalt, in jedem Lauf | `TestWettlaufNebenlaeufig` (T19) unverändert: 16 Goroutinen in 20 Läufen unter dem Detektor von `make test-race` und in `pruefung test`; Menge Läufe 20 × 16 Aufrufe je Prozess; die Rate steht in der DoD (n von m) mit Messendem, Last und Prozessoren; der deterministische Fänger dieser Zusage ist TW1 | altes Verhalten: rot in 3 von 4 `pruefung test`-Läufen bei `--cpus=4` (Messender Orchestrator, 2026-11-23, Last nicht festgehalten; der Implementer misst die Rate des alten Verhaltens in derselben Weise vor dem Fix und nennt n von m) | nebenläufig |
| TW6 | (`SPEC-239` Nr. 5) Auch der neue Weg schreibt weder Log noch `stdout` noch `stderr` und nennt weder Inhalt noch Pfad | `TestKeineAusgabe` wächst um den Weg im Fenster mit zwei Fällen (eine gültige Datei und eine Datei mit Rechteverstoß erscheinen zwischen `Stat` und `Lstat`; Haltepunkt `nachStat`): die Zählstelle 41 wird 43, die Erfassung bleibt leer, der Grund jedes Aufrufs gezählt | `println(pfad)` und `fmt.Fprintln(os.Stderr, token)` im neuen Zweig von `lade` (rot für beide Fälle) | andere |
| TW7 | (`SPEC-239` *Ungeprüft*) Akzeptiertes Negativ: eine Datei, die ein anderer Prozess nach dem `Lstat` und vor dem neuen `Stat` löscht, gibt den Grund des neuen `Stat` | kein Fänger. Haltepunkt gesucht: `ladeVorhandene` hat kein Argument, das der Test übergibt (`pfad` allein), und ein Haltepunkt dort ist ein weiteres Feld von `haltepunkte` für einen Ausgang, dessen Zeile TW2 e11 führt. Grund: eine Handlung außerhalb von `LadeToken`, kein gleichzeitiger Aufruf (`LadeToken` löscht nie). Folge: der Aufruf meldet `Datei nicht lesbar`, oder `keine reguläre Datei`, ersetzt ein Symlink ohne Ziel die Datei; der nächste Start lädt oder erzeugt die Datei. Maß: 1 Bedingung von 1 („auch `Lstat` meldet jetzt fehlt“, in `statFehler`) ohne Fänger, der Zweig selbst ist geführt | nicht anwendbar | andere |

Zeilen: TW1 bis TW7 gleich 7 (*a* = 3: TW1, TW4, TW5; *b* = 4: TW2, TW3, TW6, TW7); TW7 zählt als `andere` (`SPEC-128` Nr. 3 e).

**Mutationen und Raten** (Implementer; `AGENTS.md` §3.10, §3.15, §3.17): *wird beim Handoff gefüllt.* Je Zeile die Mutationen und der Test, der aus dem benannten Grund rot wurde; die Raten der Klasse `nebenläufig` (TW1, TW4, TW5) je Mutation mit Zahl der Prozesse, Messendem, Zahl der Prozessoren und Last; ein Fänger mit Wahrscheinlichkeit nennt seine Rate mit Zahl. TW1 ist deterministisch (rot in jedem Prozess); TW5 hat eine Rate, und der Bericht nennt sie, das alte Verhalten und den Fix. Der Implementer trägt die Tabelle und die Raten in einen Absatz dieser Stelle ein (`AGENTS.md` §3.10), denn Reviewer und Verifier haben den Bericht nicht.

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, 2026-11-23; Orakel als Literal, nie aus dem Code, nie als Aufruf des Lesers; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15): *M-Eintrag-Zugang* (TW2, TW3) mit den Achsen *Art des Eintrags* (Datei, Verzeichnis, FIFO, Symlink mit und ohne Ziel, Symlink auf ein Verzeichnis, Schleife; Inhalt gültig, leer, zwei Zeilen; Rechte `0600`, `0644`), *Zeitpunkt, zu dem der Eintrag erscheint* (vor dem `Stat`, zwischen `Stat` und `Lstat`, beim `Link`) und *Weg zum Aufruf* (`LadeToken` mit dem Haltepunkt des Tests); *M-Aktion* (TW1) mit der Achse *was das Fenster tut* (legt an, löscht, nichts). Der Modus (`record`, `replay`, `mock`, `hybrid`) ist keine Achse: das Paket liest keine Umgebung und kennt keinen Modus (`SPEC-239` Einleitung). Zählprüfung der Zweige: das Profil von `go test -coverprofile` über das Paket `admin` hat in `tokendatei.go` keinen Block mit Zähler 0 außer denen, die `SPEC-239` *Ungeprüft* nennt (Lesen schlägt fehl, 1 e; `Write`, `Sync`, `Chmod` der Hilfsdatei); der Implementer nennt jeden Ausreißer mit Funktion im Bericht.

**Unbeschränkte Mengen und ihre Prüfgrenze (akzeptierte Negative; Grund, Folge, Maß):** (1) *Zahl der gleichzeitigen Aufrufe und der Läufe* (TW5): Prüfgrenze 16 Goroutinen in 20 Läufen je Prozess; Grund: mehr Überlappung kostet Laufzeit, und das Fenster ist durch TW1 deterministisch geführt; Folge: ein Fenster, das nur bei mehr als 16 gleichzeitigen Aufrufen trifft, bliebe grün, es gibt aber nach `SPEC-239` Nr. 4 keine weiteren Fenster als die drei; Maß: der Implementer trägt n von m ein. (2) *Zwei Prozesse auf derselben Datei* und der Wechsel zwischen `Stat` und dem Lesen: `SPEC-239` *Ungeprüft*, unverändert. TW7 steht in der Tabelle.

**Größe und Schnitt** (Architect, 2026-11-23; `SPEC-128` Nr. 1, `AGENTS.md` §3.14): *a* = 3, *b* = 4, *f* = 0 (kein Port, kein Fake; der Haltepunkt ist ein Rückruf), *p* = 30 (Umbau von `lade` und `istFehlend` rund 20 Zeilen, das Feld `nachStat` mit Kommentar und Aufruf rund 10), also *S* = 85·3 + 80·4 + 200·0 + 30 = 255 + 320 + 0 + 30 = 605; 135·605 = 81675, *M* = 100 · ⌈8,1675⌉ = 900; keine Schnittfrage (*S* < 2223, die Schnittgrenze liegt 1618 darüber). Ein weiterer Schnitt ist nicht vorgesehen.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18). Die Fixtures dieses Slice sind ASCII; der Inhalt von e1 bis e7 wird aus Bytewerten gebaut.
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Die Rate von TW5 und des alten Verhaltens wird mit Last des Hosts vor und nach jedem Lauf und mit Messendem berichtet; eine Einzelzahl gilt als Größenordnung (`AGENTS.md` §3.10).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Prüfung der Sendungen** (`AGENTS.md` §3.13): dieser Plan nimmt keine Sendung an und gibt keine ab; die Rückgabe vom 2026-11-22 (Randform ohne Ort, Adresse Architect) ist mit `SPEC-239` entschieden, der Ausgang steht in §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-admin-listener-token](../done/slice-admin-listener-token.md) liegt in `done/` (er liefert `tokendatei.go` und die Tests, die dieser Slice ändert); die Spezifikation trägt den Vertrag samt Randformen auf `main`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): der Commit mit diesem Plan und der Schärfung von `SPEC-239` ist gemerged. Der Slice ist ein Fix mit Vorrang: er macht `make test` und damit jedes Gate wieder zuverlässig und startet als Ausnahme je Slice, unabhängig von der Reihenfolge der Slice-Liste von [welle-admin-api](../welle-admin-api.md) (der Planner trägt ihn dort und in der Roadmap beim Start ein).

**Stand der Bedingungen beim Start** (Planner, 2026-11-23, `main` = `09d995de` nach den Moves): `slice-admin-listener-token` liegt in `done/` (geprüft an der Datei; er liefert `tokendatei.go` mit `lade`, `haltepunkte`, `istFehlend`, `statFehler` und `ladeVorhandene` und die Tests T1 bis T31, `export_test.go` mit `LadeMitHaltepunkten` liegt im Paket), ebenso alle seine Teile (`slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung`, `slice-admin-abweisen-audit`); `SPEC-239` Nr. 1 a, Nr. 4 und *Ungeprüft* tragen den Vertrag samt Randformen auf `main` (Merge `0feeb705`), Tabelle TW1 bis TW7, Raster, Zusatztabelle und Mengen stehen in §3 und §6. Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`. Der Start gilt als Ausnahme je Slice wie bei den Admin-Slices davor ([welle-admin-api](../welle-admin-api.md) §2, Absatz zu diesem Slice); die Freigabe des Users zum Start der Admin-Slices liegt vor, und das Binary bleibt unverändert. **Kollisionen** (`git diff --name-only main...slice-admin-imports-audit`, 15 Dateien): `slice-admin-imports-audit` ändert im Paket `admin` allein `imports.go`, sechs Dateien `imports_*_test.go` und `abweisen_audit_test.go`, dazu Dateien unter `src/adapters/driven/audit/`, `src/hexagon/ports/support/`, `src/bootstrap/audit_betrieb_test.go` und seinen Plan samt Review-Report; keine davon ist eine Datei des Blocks, und weder `export_test.go` noch `bindung_test.go` noch eine `tokendatei*`-Datei steht dort. Weitere Slices in `in-progress/` gibt es nicht. Die Sendungen: der Plan nimmt keine an und gibt keine ab (§3, *Prüfung der Sendungen*); aus `done/` nennen `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-inspektion` und `slice-admin-workspaces` den Fix-Slice des Planners als Adresse des Befunds `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel` (gelesen gegen §1 *Ausdrücklich NICHT* und DoD: der Befund ist der Gegenstand dieses Slice, kein Ausschluss trifft ihn; die Evidenz-Dateien der vier liegen im Register), sonst weist kein Plan diesem Slice etwas zu.

**Abhängigkeiten — Grund je Eintrag:**

- `slice-admin-listener-token` (Teil 1 des Schnitts): *Code-Abhängigkeit*. `LadeToken`, `lade`, `haltepunkte`, `LadeMitHaltepunkten` (in `export_test.go`) und die Tests T1 bis T31 liegen dort.

**Nicht gebraucht:** die Teile `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` (sie rufen `LadeToken` nicht und ändern `tokendatei.go` nicht), `slice-bootstrap-m6` (der Aufrufer im Composition Root bleibt, die Signatur von `LadeToken` ändert sich nicht).

**Reihenfolge und Konflikte** (Architect, 2026-11-23; der Planner prüft sie beim Start): [slice-admin-imports-audit](../in-progress/slice-admin-imports-audit.md) läuft im selben Paket `src/adapters/driving/admin/`; sein Block nennt allein `imports*`-Dateien, `abweisen_audit_test.go` und die Dateien von `audit*`, kein `tokendatei*` und kein `export_test.go`; die Dateien beider Slices sind disjunkt, die Zweige laufen nebeneinander und mergen unabhängig. Die Marke dieses Slice nennt die Dateien einzeln (`tokendatei*`), damit die Zeilen fremder Slices im Verzeichnis nicht mitzählen (`SPEC-128` Nr. 10 g; `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`); `bindung_test.go` und `export_test.go` stehen nicht im Block. Der Flake trifft jeden Lauf von `make test` im Paket, auch die Gates der laufenden Slices; je früher dieser Fix auf `main` liegt, desto früher ist `make gates` der anderen Linien zuverlässig. Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-23): *S* = 605 gleich der Tabelle (7 Zeilen: *a* = 3, *b* = 4; *f* = 0; *p* = 30, Rechenweg in §3 *Größe und Schnitt*), **Marke 900** gleich der Formel (135·605 = 81675, ⌈8,1675⌉ = 9); keine Schnittfrage. Der Block trägt den Teil des Architect (`pfad`, `produktcode`, `fakes`); **der Planner hat `basis`, `schaetzung` (605) und `marke` (900) beim Start am 2026-11-23 vor dem ersten Code-Commit gesetzt** (`AGENTS.md` §3.14; `BEO-REPO/marke-beim-start-nicht-gesetzt`) und dabei die Tabelle neu gerechnet: sieben Zeilen (TW1, TW4, TW5 `nebenläufig`, also *a* = 3; TW2, TW3, TW6, TW7 `andere`, also *b* = 4), *f* = 0, *p* = 30, *S* = 85·3 + 80·4 + 200·0 + 30 = 605, *M* = 100 · ⌈135·605 / 10000⌉ = 100 · ⌈8,1675⌉ = 900, keine Schnittfrage (605 < 2223); das Werkzeug (`make groessenmarke`) bestätigt Schätzung gleich Tabelle und Marke gleich Formel. **Messbasis** `09d995de` (der Commit des Verweis-Nachzugs nach dem Move nach `in-progress/` auf `main`, der letzte Commit von `main` vor dem Code-Start; die Messung ist 0): die vier vorhandenen Dateien des Blocks liegen unverändert vor der Basis (jüngster Commit unter ihnen `58bd1f2a`), `tokendatei_fenster_test.go` und `tokendatei_halt_test.go` gibt es auf `main` nicht, also trägt die Basis keine fremde Zeile. Eine ältere Marke gibt es nicht (kein Eintrag `ersetzt`). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Legt der Implementer eine weitere Datei an oder ändert er eine weitere Datei, nennt er sie im selben Commit im Block.

```groessenmarke
pfad: src/adapters/driving/admin/tokendatei.go
pfad: src/adapters/driving/admin/tokendatei_test.go
pfad: src/adapters/driving/admin/tokendatei_zusagen_test.go
pfad: src/adapters/driving/admin/tokendatei_ausgabe_test.go
pfad: src/adapters/driving/admin/tokendatei_fenster_test.go
pfad: src/adapters/driving/admin/tokendatei_halt_test.go
produktcode: 30
fakes: 0
basis: 09d995de
schaetzung: 605
marke: 900
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Ein weiterer Schnitt ist nicht vorgesehen; die Schätzung liegt weit unter der Schnittgrenze.
- `in-progress` → `open` (blockiert): ein Lauf von `TestWettlaufNebenlaeufig` ist trotz des Fix rot, oder der Implementer bemerkt eine Randform ohne Ort in `SPEC-239` (`AGENTS.md` §3.12); beides geht mit Datum und Adresse Architect nach §6, der Implementer entscheidet es nicht im Code.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, je Zeile der Tabelle ist eine Mutation rot gesehen (`AGENTS.md` §3.10), `TestWettlaufNebenlaeufig` ist in jedem Lauf der Messung von §2 grün; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt eine Nummer von `SPEC-239` (Architect, 2026-11-23) oder „gilt nicht“ mit Grund; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12). Das Argument ist der Eintrag an `pfad` (`LadeToken` hat kein weiteres); „fehlender Wert“ heißt: nichts liegt an `pfad`, „leerer Wert“: die Datei ist leer, „unbekannter oder fremder Wert“: ein Eintrag anderer Art oder mit Verstoß.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Entscheidung nach `Stat` und `Lstat` (Nr. 1 a), auch im Fenster | Nr. 1 a (`Lstat` meldet „fehlt“: Erzeugen, Nr. 3); Fänger TW1 (Zellen löscht und nichts) | Nr. 1 a und 1 g (die leere Datei im Fenster gibt `weniger als 32 Zeichen`); ein leerer `pfad` ist `model.ErrUngueltig` (Nr. 1), vor dem `Stat`, im Fenster nicht erreichbar | Nr. 1 a mit b bis g und Nr. 6 (Verzeichnis, FIFO, Symlink mit und ohne Ziel, Schleife, Datei mit Verstoß: der Ausgang des Eintrags vor dem `Stat`) | gilt nicht: `LadeToken` läuft synchron ohne Kontext (Einleitung, `SPEC-110` Nr. 13) und legt vor dem `Lstat` nichts an; ein Prozessende im Fenster hinterlässt nichts | Nr. 4 (der zweite Aufruf liefert dasselbe Token, die Datei bleibt unverändert) |
| Laden eines Eintrags, den ein anderer anlegte (`ladeVorhandene`; Nr. 1 a im Fenster, Nr. 3 d) | Nr. 1 a *Ungeprüft* (die Datei verschwindet vor dem neuen `Stat`: Grund des neuen `Stat`; TW7) | Nr. 1 g (`weniger als 32 Zeichen`) | Nr. 1 b bis g und Nr. 6 (Rang bei mehreren Verstößen, TW2 e7) | gilt nicht: wie oben (kein Kontext); ein Absturz lässt keine Hilfsdatei zurück, die der Fall des Fensters anlegt | Nr. 4 (jeder Aufruf liest neu, `LadeToken` hält keinen Zustand) |
| Erzeugen (Hilfsdatei und `Link`, Nr. 3 b bis d; Fenster beim `Link`) | Nr. 3 f (das Verzeichnis fehlt: `Datei nicht anlegbar`); Nr. 3 d (die Hilfsdatei fehlt beim `Link`: `Datei nicht anlegbar`, T30) | gilt nicht: `LadeToken` legt nie eine leere `pfad` an, die Datei erscheint vollständig (Nr. 3, *Entscheidung*); eine leere Datei von außen: Nr. 1 g | Nr. 3 d (`fs.ErrExist`: der Aufruf lädt den Eintrag nach Nr. 1 b bis g, TW2 Zugang 3) | Nr. 3 e und *Ungeprüft* (ein Absturz zwischen 3 b und 3 e lässt die Hilfsdatei zurück; kein Test beendet den Prozess) | Nr. 4 (gleichzeitige Aufrufe erzeugen genau eine Datei, TW4, TW5) |

**Zusatztabelle: Paare von Verstößen und Wege** (`AGENTS.md` §3.16; der Modus ist keine Achse, siehe §3):

| Nr. | Paar | Entscheidung | Ort | Fänger |
|---|---|---|---|---|
| 1 | Eintrag im Fenster mit zwei Verstößen (Rechte `0644` und zwei Zeilen) | der kleinste Rang gilt: `Gruppen- oder Weltrechte` | Nr. 1 a mit Nr. 6 | TW2 (e7 in drei Zugängen) |
| 2 | derselbe Eintrag in jedem der drei Zugänge (vor dem `Stat`, zwischen `Stat` und `Lstat`, beim `Link`) | derselbe Ausgang; ein Symlink ohne Ziel `keine reguläre Datei` in jedem Zugang | Nr. 1 a, Nr. 3 d, Nr. 4 | TW2 (33 Zellen) |
| 3 | gleichzeitiger Aufruf legt an × Symlink ohne Ziel liegt vor dem `Stat` und ist im Fenster entfernt | der Gewinner lädt das Token des anderen, der Aufruf mit dem entfernten Symlink erzeugt (Nr. 1 a, Nr. 3) | Nr. 1 a, Nr. 4 | TW1 (löscht, legt an) |
| 4 | Fenster × Modus (`record`, `replay`, `mock`, `hybrid`) | gilt nicht: das Paket liest keine Umgebung und kennt keinen Modus | Einleitung von `SPEC-239` | gilt nicht |

- Der Flake bleibt, weil ein weiteres Fenster existiert, das `SPEC-239` Nr. 4 nicht nennt (etwa ein Fenster, in dem `Stat` einen Fehler außer „fehlt“ meldet) — **Ausgang:** weiter offen; Adresse: Architect (der Implementer hält an, trägt die Zeile mit Datum in diese Stelle ein und gibt den roten Lauf mit Meldung und Zahl an den Architect; `AGENTS.md` §3.12, §3.13); Register: `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel`.
- Die Rate des Last-Tests schwankt mit der Last des Hosts, und 30 grüne Prozesse schließen eine Rate unter 10 % nicht aus (§2) — **Ausgang:** weiter offen → `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` im Register; der Beweis ist TW1, die Rate nur der Zusatzbeleg.
- Der Fix macht den Ausgang von Handlungen außerhalb (ein Symlink mit Ziel, der im Fenster entsteht, wird geladen statt abgewiesen) — **Ausgang:** entschieden (Architect, 2026-11-23, §3 *Entscheidung der Form*): Folge im Betrieb im Vertrag benannt, kein Eintrag nötig.
- Die Fixtures gleichen dem Wert einer Mutation (`0644` gegen `0600`, ein gültiger Inhalt gegen den erzeugten Token) — **Ausgang:** weiter offen → `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` im Register; TW3 nennt die Probe, der Reviewer löst sie auf.
- Ein Fänger, der Elemente zählt, die der Befund nicht nannte: die Menge TW2 bildet der Implementer aus dem Wortlaut von `SPEC-239` Nr. 1 a und b bis g (11 Einträge × 3 Zugänge), nicht aus dem Befund (nur „Datei“ im Fenster) — **Ausgang:** weiter offen → `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` im Register (`AGENTS.md` §3.17); der Bericht nennt je Befund die Menge, die Zahl und eine Mutation an einem ungenannten Element (e8, e10 oder e11).
- Die Marke steht im Block erst beim Start, weil die Messbasis vorher nicht feststeht — **Ausgang:** entfallen (Planner, 2026-11-23, beim Start): `basis` (`09d995de`), `schaetzung` (605) und `marke` (900) stehen im Block (§4) vor dem ersten Code-Commit (`BEO-REPO/marke-beim-start-nicht-gesetzt`).
- Zeilen fremder Slices an den Pfaden des Blocks: ein anderer Slice, der auf `main` landet, bevor dieser Slice endet, könnte eine der sechs Dateien ändern, und die Messung zählte sie mit (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`, 6×; dieser Slice wäre ein siebter Vorgang, wenn es geschieht) — **Ausgang:** weiter offen; beim Start geprüft (Planner, 2026-11-23): die Basis trägt keine fremde Zeile, und `slice-admin-imports-audit` (in Arbeit, derselbe Paket-Ordner) ändert keine Datei des Blocks (§4, *Stand der Bedingungen beim Start*); Adresse: Implementer (misst beim Handoff `git diff --numstat <basis>..<main-Elter des letzten Merge von main> -- <Pfade des Blocks>`, nennt die fremden Zeilen in der Messpunkt-Zeile und setzt `ueberschreitung` nur bei Überschreitung).
- Akzeptierte Negative (TW7; §3: (1) Zahl der Läufe von TW5, (2) zwei Prozesse und Wechsel zwischen `Stat` und Lesen) — **Ausgang:** weiter offen; Adresse: Reviewer und Verifier lösen sie gegen die Klassen auf (`AGENTS.md` §3.15, §3.17); der Implementer trägt die Maße (n von m) ein.

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

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register am 2026-11-23 gesichtet und beim Start erneut (84 Einträge; der Zähler ist die Zahl der Dateien unter `evidence/`, Treffer und Zähler beim Start nachgezogen). Offen und für den Slice einschlägig: `BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel` (4×: `slice-admin-authentisierung`, `slice-admin-inspektion`, `slice-admin-rahmen`, `slice-admin-workspaces`; dieser Slice ist sein Träger), `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×) und `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×) für die Größenmarke (§4), `BEO-REPO/aequivalente-mutation-hinter-doppeltem-waechter` (1×) und `BEO-REPO/aequivalent-eingestuft-ohne-umzuordnende-eingabe` (2×) für die Wahl der Form in §3. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (20×, `AGENTS.md` §3.10), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×, `AGENTS.md` §3.10), `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (9×), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (70×, `AGENTS.md` §3.17), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (19×, `AGENTS.md` §3.15) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (22×).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
