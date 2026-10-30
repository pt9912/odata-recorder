# Slice slice-guide-groessen-gate: Größe der Guide-Dateien als Gate

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-sensoren](../welle-sensoren.md) — zweiter Slice der Folge; Start, wenn `slice-unsichtbare-zeichen` in `done/` liegt.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an Gates), Retirement-Check `docs/reviews/2026-10-25-retirement-check-hard-rules.md` §5 und §7 (Vorschlag `slice-guide-groessen-gate`, Aufwand S), Baseline-Regelwerk `grundlagen-klassifikation.md` §Guide-Datei-Wildwuchs (Obergrenze `file.max-lines`/`max-bytes`), [ADR-0002](../../adr/0002-architektur-gate-a-check.md) und [ADR-0003](../../adr/0003-lint-profil-golangci.md) (Muster: ein Gate mit Bindung an eine ADR), `AGENTS.md` §3.6 (Gates werden nicht ohne ADR gelockert).

**Berührte Spec-Stellen:** `.d-check.yml` (Konfiguration des Doku-Gates; Block des Moduls `file`) · `SPEC-040` (Vertrag von `docs-check`, Nachzug durch den Architect) · `harness/README.md` §Sensors (Zeile von `make docs-check`, Spalte Vertrag auf höchstens 200 Zeichen, sonst Prosa nach `harness/sensors/`). Produkt-Spezifikation und Architektur-Sicht bleiben unberührt.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Gate in der Kette `make gates` (das d-check-Modul `file` unter `make docs-check`, oder ein anderer Träger nach der Entscheidung des Architect) wird rot, wenn eine Guide-Datei (`AGENTS.md`, die Commands, die Rollen-Typen, die Skills und weitere vom Architect benannte) ihre Obergrenze an Zeilen oder Bytes überschreitet; die Grenzen frieren den gemessenen Ist-Stand ein und sinken mit der Verschlankung, und die Gegenprobe zeigt das Rot.

**Messwerte des Planners am Stand `0c524e2`** (Bytes aus `wc -c`, Zeilen aus `wc -l`, zur Orientierung; der Implementer misst am Start neu): `AGENTS.md` 24 648 Bytes und 337 Zeilen; `.claude/commands/implement-slice.md` 16 992 Bytes; `.claude/commands/close-welle.md` 11 369; `.claude/commands/plan-welle.md` 6 554; `.harness/skills/reviewer.md` 7 397; `.harness/skills/closure-note-reviewer.md` 4 144; die sechs Rollen-Typen unter `.claude/agents/` zwischen 2 166 und 3 759; `harness/README.md` 10 382; `harness/conventions.md` 16 065. Die Zahlen sind Bytes; der Retirement-Check zählt Zeichen (24 193 für `AGENTS.md`); die Einheit des Gates entscheidet der Architect (Auftrag 1).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Verschlankung von `AGENTS.md` — Folgearbeit nach der Welle (Weg 2 des Users: erst Sensoren, dann Messwerte, [welle-sensoren](../welle-sensoren.md) §6); das Gate friert ein und senkt die Grenze erst nach einer Verschlankung, die ein anderer Vorgang des Users ist. Die Richtgröße des Retirement-Checks (17 000 Bytes, 300 Zeilen für `AGENTS.md`) ist eine Schätzung nach der Verschlankung und kein Wert dieses Slice.
- Eine Änderung an `AGENTS.md`, Commands, Rollen-Typen oder Skills — sie ändert nur der User; auch ein Hinweis auf das neue Gate in `AGENTS.md` (Zeiger) geht als Diff-Vorschlag in der Closure.
- Dateien, die kein Guide sind: Pläne (`docs/plan/planning/`), ADRs, Berichte, Spezifikation, Lastenheft, die vendorte Baseline — sie wachsen mit dem Gegenstand und werden nicht von jedem Lauf geladen; eine Obergrenze für sie ist eine andere Entscheidung (`AGENTS.md` §3.14 begrenzt den Plan über die Größenmarke des Codes, nicht über seine Länge).
- Die Grenze je Hard Rule (`structure` mit `max-lines` je Abschnitt `### 3.<n>`) — nur, wenn der Architect sie in Auftrag 4 als tragbar belegt; sonst Folgearbeit nach der Verschlankung. Der Handbuchtext nennt sie, erprobt ist sie nicht.
- Ein Gate gegen den Inhalt der Guide-Dateien (Duplikate zwischen `AGENTS.md`, Commands und Skills) — nicht berechenbar; das ist Review und Verschlankung.

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

- [ ] Das Modul `file` (oder der vom Architect gewählte Träger) ist aktiv für die benannte Dateimenge, mit Grenzen in Zeilen und Bytes je Datei oder Gruppe; die Grenzen stehen auf dem gemessenen Ist-Stand, jede Grenze mit ihrer Messung im Kommentar der Konfiguration (Klasse *Grenze* und *Rang-Zeiger* nach `AGENTS.md` §3.7), und `make docs-check` ist grün.
- [ ] Die Wirkung des Moduls ist erprobt und belegt, nicht aus dem Handbuch geschlossen: ein Fall der Gegenproben (oder, wenn der Architect es anders entscheidet, ein beschriebener Lauf mit Ausgabe im Vertrag) zeigt, dass eine um ein Byte und um eine Zeile verlängerte Guide-Datei das Gate rot macht und dass eine Datei an der Grenze grün bleibt; die Zusatzzeit der Gegenproben ist gemessen und in §7 berichtet.
- [ ] Die Regel zum Anheben und Senken der Grenze steht dort, wo der Architect sie entscheidet (Vertrag, Kommentar der Konfiguration oder ADR, Auftrag 3), und die Spalte *Vertrag* der Zeile von `make docs-check` in `harness/README.md` nennt das Modul, ohne die 200 Zeichen zu überschreiten.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `SPEC-040` (Nachzug durch den Architect), Zeile von `make docs-check` in `harness/README.md`; die Closure legt den Diff-Vorschlag für `AGENTS.md` bei (Zeiger auf das Gate; der User ändert).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

Der Planner schneidet; der Architect entscheidet (Aufträge unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.d-check.yml` | update | `file` in `modules:`, Block mit Dateimenge und Grenzen (Zeilen, Bytes); der Kommentarblock nennt je Grenze Messung, Rang-Zeiger und die Regel zum Anheben und Senken |
| `harness/vertraege/SPEC-040-docs-check.md` | update (Architect) | die neue Zusage von `docs-check`: Dateimenge, Einheit, Verhalten bei fehlender Datei und leerem Treffer, *Ungeprüft* |
| `harness/README.md` | update | Zeile von `make docs-check` in §Sensors (Spalte Vertrag ≤ 200 Zeichen, `structure`-Regel von `.d-check.yml`), Zeile in *Nicht behauptet*, falls der Architect Reste benennt (z. B. `.harness/**` außerhalb der Prüfmenge) |
| `docs/plan/adr/` | neu (Architect), falls Auftrag 3 oder 6 es verlangt | Entscheidung über die Regel zum Anheben der Grenze (Verhältnis zu `AGENTS.md` §3.6) und über die Bindung des Gates; Zeile im ADR-Index |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh`, `harness/mk/gegenproben.mk` | update | Fall(e) zu `SPEC-102`, nummeriert nach dem letzten Fall: eine Kopie des Baums, in der eine Guide-Datei an die Grenze und über sie gebracht wird, und `docs-check` auf der Kopie ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); `EINGABEN` um die Pfade der Konfiguration; Stempel nach dem letzten Commit unter den Eingabepfaden |
| Dateien unter `.claude/` und `.harness/skills/` | **nicht geändert** | nur gemessen; ob das Gate sie erreicht (`scan.ignore` nennt `.harness/**`), ist Auftrag 1 |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Wirkung des Moduls am Image gelesen und belegt, Aufträge 1 bis 7, Tabelle *Zusage · Fänger*, Raster, Nachzüge, erwartete Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorschlag an den User).

**Betriebsregeln für alle Rollen** (Wortlaut und Gründe: [welle-sensoren](../welle-sensoren.md) §4):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container gleichzeitig; `make gates` nie parallel zu Mutationsläufen; Docker sonst nur über `make` (`AGENTS.md` §3.1). Gate-Läufe laufen im Hauptbaum, weil die Images feste Tags tragen. Das Doku-Gate selbst fährt in einem per Digest gepinnten Image mit schreibgeschütztem Mount (`:ro`).
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- **Eine Grenze wird nie angehoben, um den Slice grün zu bekommen** (`AGENTS.md` §3.6): wächst eine Guide-Datei während des Slice (der Slice selbst ändert keine), meldet der Implementer es dem Planner, statt die Grenze zu setzen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen.
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; Wirkung des Moduls: das Handbuch des gepinnten Images, das der Architect über `make docs-check` und die Hilfe des Images liest, nicht aus dem Gedächtnis; der Vertrag steht in `harness/vertraege/`, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), `AGENTS.md` §3.8):

1. **Die Wirkung des Moduls `file` belegen.** Schlüssel der Konfiguration (`max-lines`, `max-bytes`, Dateimenge als Pfad oder Glob), Fehlercodes (`file-lines-exceeded`, `file-bytes-exceeded`), Einheit (Bytes oder Zeichen; Umlaute zählen verschieden: 24 648 Bytes gegen 24 193 Zeichen bei `AGENTS.md`), Zeilenzählung (letzte Zeile ohne Zeilenende), und vor allem: Erreicht das Modul Dateien unter `.claude/` (verstecktes Verzeichnis) und unter `.harness/skills/`, obwohl `scan.ignore` in `.d-check.yml` `.harness/**` ausnimmt? Ist die Wirkung am Pin nicht so, wie das Handbuch sagt, oder erreicht das Modul diese Dateien nicht, entscheidet der Architect mit Grund und Folge: anderer Träger (ein Skript unter `tools/harness/`, ein Ziel in `GATE_CHECKS`), oder ein benannter Rest unter *Nicht behauptet*. Das Ergebnis steht vor dem ersten Code-Commit im Vertrag (`AGENTS.md` §3.12).
2. **Die Dateimenge** der Guide-Dateien mit Grund je Gruppe: `AGENTS.md`, `.claude/commands/*.md`, `.claude/agents/*.md`, `.harness/skills/*.md`, und ob `harness/README.md` und `harness/conventions.md` (jeder Lauf liest sie laut Leseordnung) dazugehören. Eine Datei, die dazugehören müsste und nicht erreicht wird, ist ein benannter Rest, keine Lücke im Stillen.
3. **Die Grenzen und ihre Regel.** Ist-Stand genau oder mit Spielraum (je Datei), Zeilen und Bytes beide oder nur eines; wie eine Grenze wandert: **Senken** ist eine Verschärfung (ein Commit mit der Messung, der Verschlankung nachgelagert), **Anheben** ist eine Lockerung im Sinn von `AGENTS.md` §3.6 (ADR) oder, wie der User es für die Welle vorgesehen hat, ein sichtbarer Commit mit Begründung — der Architect löst den Widerspruch beider Lesarten auf und nennt Grund und Folge: eine ADR je Anhebung, eine ADR einmal für die Ratsche, oder ein Carveout mit Frist. Ob die Grenze am Ist-Stand eine Verschlankung erzwingt, bevor jemand `AGENTS.md` um eine Zeile erweitern kann: ja, und das ist gewollt; die Folge (jeder Diff-Vorschlag des Users, der `AGENTS.md` verlängert, bricht das Gate) steht im Vertrag.
4. **Grenze je Hard Rule (`structure` mit `max-lines` je Abschnitt `### 3.<n>`)** — tragbar oder nicht, am Image belegt (Handbuchtext: „nicht erprobt“, Retirement-Check Abschnitt 5); wenn ja, mit Wert und Grund, wenn nein, als benannter Rest mit Folge.
5. **Der Nachweis des Rots.** Kopie des Baums mit verlängerter Guide-Datei in den Gegenproben ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) — der Fall fährt `docs-check` (Docker, gepinntes Image, `:ro`, netzlos) auf der Kopie: wie, ohne dass die Stufe `gegenproben` Docker aufruft (Treiber auf dem Host wie bei `groessenmarke-docker.sh`/`schema.mk`, oder ein anderer Weg), und was der Fall kostet (Zeitbudget in `SPEC-102`; `docs-check` allein dauert wenige Sekunden, ein Image-Start je Fall zählt). Das Fehlen einer Gegenprobe zu `make docs-check` ist heute in `harness/README.md` *Nicht behauptet* genannt; dieser Slice schließt den Teil, den er trägt, und nennt den Rest.
6. **Tabelle *Zusage · Fänger*** (§3), **Raster** (§6), **Nachzüge** nach `AGENTS.md` §3.9 (`SPEC-040`, `harness/README.md`, Gegenproben-Eingaben) und ob eine ADR zum Gate nötig ist (Muster: Bindung-Spalte; neue Gates sind erlaubt, eine Regel zum Anheben der Grenze ist eine Entscheidung, Auftrag 3).
7. **Erwartete Größe** (Zeilen Konfiguration und Skript, falls einer entsteht; Fall der Gegenprobe) für die Marke des Planners; ob ein Teil nebenläufig ist (zählt × 85).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 6) | — | — |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-unsichtbare-zeichen` liegt in `done/` (Reihenfolge der Welle: beide berühren `harness/README.md`, die Gegenproben und den Stempel), und der Architect hat die Aufträge 1 bis 7 entschieden, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner schreibt nach dem Architect den Block `groessenmarke` ([SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 3) in diese Sektion und setzt die Marke mit `make groessenmarke` im eigenen Commit vor dem ersten Code-Commit. Aufwand nach Einschätzung des Planners: S.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Schätzung des Architect mal 1,25 liegt über 3000 (`AGENTS.md` §3.14); oder der Fall der Gegenprobe braucht einen neuen Treiber, der Docker aus dem Host heraus fährt und mehr als die Hälfte der Größe trägt (dann ein eigener Slice für den Treiber).
- `in-progress` → `open` (blockiert): das Modul `file` erreicht die Guide-Dateien nicht und der Architect findet keinen Träger ohne neue Host-Abhängigkeit über `bash`, `git` und `docker` hinaus (`harness/conventions.md` §Baseline); dann geht die Frage an den User.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, `make gates` und `make gegenproben` grün im Hauptbaum, Review- und Verifikations-Report unter `docs/reviews/`, Closure-Notiz in §7. Der Lerneintrag nennt den Sensor mit `liegt in` (`.d-check.yml` oder das Ziel, mit dem Anker `seit welle-sensoren`), die Größen der Guide-Dateien zum Stand der Closure (Eingabe der Verschlankung) und den Diff-Vorschlag an den User.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Das Modul `file` ist am Pin nicht erprobt; es kann die Dateien unter `.claude/` und `.harness/skills/` nicht erreichen (`scan.ignore`, verstecktes Verzeichnis) oder anders zählen als das Handbuch sagt — **Ausgang:** offen bis zur Closure; entschieden im Vertrag (Auftrag 1), der Beleg ist ein gelaufener Fall.
- Ratsche ohne Entscheidung des Users über das Anheben: die Grenze am Ist-Stand färbt jede Erweiterung von `AGENTS.md` rot, auch eine berechtigte; ein Gate, das niemand anhebt, wird umgangen (`AGENTS.md` §3.2) oder gelockert — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 3, die Folge für die Diff-Vorschläge des Users steht im Vertrag.
- Die Messung von Zeilen und Bytes ist Handarbeit, sobald die Guide-Dateien wachsen (Bytes in `.d-check.yml` führen den Stand des Tages): der Kommentar der Konfiguration trägt die Messung; ein Kommentar, der den alten Stand nennt, ist falsch (`AGENTS.md` §3.7) — **Ausgang:** offen bis zur Closure; der Implementer trägt den Stand ein, die Closure misst neu.
- Laufzeit der Gegenproben: ein Fall mit Image-Start verlängert den Lauf, dessen Re-Evaluierungs-Trigger ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): Lauf über 30 s bei geänderter Eingabe; gemessen 56,9 s Schritt) beim User unentschieden ist — **Ausgang:** offen bis zur Closure; der Slice misst die Zusatzzeit und berichtet sie, er entscheidet den Trigger nicht; weiter offen: → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (Adresse: User).
- Eine Gegenprobe, die `docs-check` fährt, braucht Docker auf dem Host in der Gegenproben-Kette; heute ruft die Stufe `gegenproben` kein Docker — **Ausgang:** offen bis zur Closure; entschieden in Auftrag 5, ein benannter Rest unter *Nicht behauptet*, wenn kein Weg ohne neue Host-Abhängigkeit bleibt.

**Raster der Randformen** (`AGENTS.md` §3.16; füllt der Architect vor dem ersten Code-Commit: jede Zelle nennt ihren Ort im Vertrag oder „gilt nicht“ mit Grund; eine leere Zelle ist ein Fehler des Plans; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Messen einer Datei gegen die Grenze | (Vertrag, offen) Datei der Menge fehlt | (Vertrag, offen) leere Datei; Datei nur aus Zeilenende | (Vertrag, offen) Grenze −1, genau an der Grenze, Grenze +1 (Zeilen und Bytes je); Datei ohne Zeilenende am Schluss; Datei mit Mehrbyte-Zeichen | (Vertrag, offen) Lesefehler | (Vertrag, offen) zweiter Lauf auf unverändertem Baum |
| Bestimmen der Dateimenge (Pfad oder Glob) | (Vertrag, offen) Glob trifft nichts | (Vertrag, offen) leere Liste | (Vertrag, offen) Datei unter `scan.ignore`; Datei in verstecktem Verzeichnis; Symlink | (Vertrag, offen) gilt nicht, falls reine Konfiguration | (Vertrag, offen) dieselbe Datei in zwei Gruppen mit zwei Grenzen |
| Lesen der Grenze (Konfiguration) | (Vertrag, offen) Grenze nicht gesetzt | (Vertrag, offen) Grenze 0 | (Vertrag, offen) negativer oder nicht numerischer Wert; unbekannter Schlüssel | (Vertrag, offen) gilt nicht, falls reine Konfiguration | (Vertrag, offen) gleicher Schlüssel zweimal |
| Anheben und Senken der Grenze | (Vertrag, offen) Anhebung ohne Begründung im Commit | (Vertrag, offen) Anhebung auf den Ist-Stand | (Vertrag, offen) Senken unter den Ist-Stand (Gate rot an der eigenen Datei) | (Vertrag, offen) gilt nicht | (Vertrag, offen) Anhebung desselben Werts zweimal |
| Meldung und Ausgang | (Vertrag, offen) keine Überschreitung | (Vertrag, offen) Überschreitung nur einer der beiden Grenzen | (Vertrag, offen) beide Grenzen derselben Datei, mehrere Dateien | (Vertrag, offen) Ausgabe bricht | (Vertrag, offen) gleiche Meldung beim zweiten Lauf |

**Zusatztabelle Paare und Modi** (Spalte *Fänger*; füllt der Architect: Zeilengrenze und Byte-Grenze derselben Datei, eine Datei in zwei Gruppen, `file` gegen `structure` an derselben Datei):

| Paar oder Modus | Entscheidung (Ort im Vertrag) | Fänger |
|---|---|---|
| (Architect) | (Vertrag, offen) | — |

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

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: die Größen aller Guide-Dateien (Bytes und Zeilen) zum Stand der Closure neben den Grenzen des Gates, die gemessene Zusatzzeit der Gegenproben (Schritt und Gesamtlauf, vorher und nachher, Zahl der Läufe, Zustand des Hosts) und den Diff-Vorschlag an den User.

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, Stand 2026-10-30, 19 Einträge; Zahlen aus den Dateien unter `evidence/`). Für diesen Slice zählen:
- `gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, offen), `stempel-gehoert-zu-keinem-committeten-stand` (1×, offen), `betriebsregel-fuer-repo-dateien-verletzt` (1×, offen): Gegenprobe ohne Produkt in der Eingabe, Stempel nach dem letzten Commit, Betriebsregeln in §3.
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×), `akzeptiertes-negativ-ohne-folge` (10×), `randform-erst-im-review-entschieden` (20×), `universalaussage-im-vertrag-ungeprueft` (9×): Tabelle, Raster und der Satz „die Grenze gilt für jede Guide-Datei der Menge“ mit Fänger über alle Elemente der Menge.
- `berichtigung-erreicht-nicht-jede-stelle` (11×, §3.9): die Größen stehen in der Konfiguration, im Kommentar und in diesem Plan; ein Stand, der sich ändert, zieht alle Stellen im selben Commit nach.
- `hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×, offen): trifft zu, falls die Gegenprobe `docs-check` durch ein Skript ersetzt oder umhüllt; die Hülle trägt jede Schnittstelle des Originals.
- `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen): trifft nur zu, falls ein Fänger nur mit Wahrscheinlichkeit rot wird; ein Treffer wäre der dritte und macht aus der Notiz eine Lücke (eigener Slice).
Die Retirement-Prüfung der Regeln mit Herkunfts-Anker (`AGENTS.md` §3.6; `grundlagen-traceability.md` Sensor 2) trifft diesen Slice: er lockert keine Regel; er fügt ein Gate hinzu. Unter der Schwelle ohne Bezug: die übrigen Einträge.

**Modus-Begründungsblock — Umfang.** Alle berührten Sub-Areas GF (Default `*`, Greenfield: Neues Repo ohne Bestandscode; die Spec führt, der Code folgt).
