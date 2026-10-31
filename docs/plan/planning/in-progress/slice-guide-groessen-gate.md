# Slice slice-guide-groessen-gate: Größe der Guide-Dateien als Gate

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-sensoren](../welle-sensoren.md) — zweiter Slice der Folge, zweiter von Teil 1; Start, wenn `slice-unsichtbare-zeichen` in `done/` liegt. Er liegt vor `slice-workspace-use-case` (Reihenfolge, kein Blocker, [welle-sensoren](../welle-sensoren.md) §5).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an Gates), Retirement-Check `docs/reviews/2026-10-25-retirement-check-hard-rules.md` §5 und §7 (Vorschlag `slice-guide-groessen-gate`, Aufwand S), Baseline-Regelwerk `grundlagen-klassifikation.md` §Guide-Datei-Wildwuchs (Obergrenze `file.max-lines`/`max-bytes`), [ADR-0002](../../adr/0002-architektur-gate-a-check.md) und [ADR-0003](../../adr/0003-lint-profil-golangci.md) (Muster: ein Gate mit Bindung an eine ADR), [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (Proposed, Entwurf des Architect: Träger und Regel zum Anheben der Grenze), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), `AGENTS.md` §3.6 (Gates werden nicht ohne ADR gelockert).

**Berührte Spec-Stellen:** `.d-check.yml` (Konfiguration des Doku-Gates; Module `file` und `structure`) · `SPEC-040` (Vertrag von `docs-check`, vom Architect geschrieben) · `SPEC-102` (Fälle 156 bis 170, Eingaben, Aufbau) · [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) mit Index-Zeile · `harness/README.md` §Sensors (Zeilen von `make docs-check` und `make gegenproben`, Spalte Vertrag und *Tut was* auf höchstens 200 Zeichen; *Nicht behauptet*). Produkt-Spezifikation und Architektur-Sicht bleiben unberührt.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** pt9912 (Zweig slice-guide-groessen-gate)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Gate in der Kette `make gates` (die d-check-Module `file` und `structure` unter `make docs-check`, `SPEC-040` Nr. 1, kein neues Ziel) wird rot, wenn eine Guide-Datei der Prüfmenge (`AGENTS.md`, die Commands, die Rollen-Typen, die Skills und `harness/conventions.md`: 13 Dateien in 5 Regeln, `SPEC-040` Nr. 2) ihre Obergrenze an Zeilen oder Bytes überschreitet oder ein Abschnitt `### 3.<n>` von `AGENTS.md` seine Obergrenze an Zeilen (`SPEC-040` Nr. 9); die Grenzen frieren den gemessenen Ist-Stand ein und sinken mit der Verschlankung (Nr. 8), und die Gegenprobe zeigt das Rot (Fälle 156 bis 170 von `SPEC-102`, das Binary `d-check` aus dem Image des Pins, ohne Docker auf dem Host).

**Messwerte des Architect am Stand `4e36c01`** (Zeilen aus `wc -l`, Bytes aus `wc -c`; der Implementer misst am Start neu): `AGENTS.md` 337 Zeilen und 24 648 Bytes; `.claude/commands/` `close-welle.md` 142/11 369, `implement-slice.md` 221/16 992, `plan-welle.md` 99/6 554; `.claude/agents/` zwischen 36/2 166 (`planner.md`) und 53/3 759 (`architect.md`); `.harness/skills/` `closure-note-reviewer.md` 90/4 144, `reviewer.md` 137/7 397; `harness/conventions.md` 228/16 065; `harness/README.md` 137/11 673 (nicht in der Menge, `SPEC-040` Nr. 10 d). Die Gruppen stehen auf dem Maximum ihrer Dateien; die Summe der Grenzen ist 976 Zeilen und 68 861 Bytes (Bezugsgröße der Verschiebung, `SPEC-040` Nr. 8 b). Gemessen werden Bytes; der Retirement-Check zählt Zeichen (24 193 für `AGENTS.md`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Verschlankung von `AGENTS.md` — Folgearbeit nach der Welle (Weg 2 des Users: erst Sensoren, dann Messwerte, [welle-sensoren](../welle-sensoren.md) §6); das Gate friert ein und senkt die Grenze erst nach einer Verschlankung, die ein anderer Vorgang des Users ist. Die Richtgröße des Retirement-Checks (17 000 Bytes, 300 Zeilen für `AGENTS.md`) ist eine Schätzung nach der Verschlankung und kein Wert dieses Slice.
- Eine Änderung an `AGENTS.md`, Commands, Rollen-Typen oder Skills — sie ändert nur der User. Ein Zeiger auf das Gate in `AGENTS.md` entfällt nach der Empfehlung des Architect (`AGENTS.md` §4: der Gate-Index steht einmal in `harness/README.md`; eine Zeile mehr wäre bei Grenze gleich Ist rot, `SPEC-040` Nr. 8 d); die Closure legt dem User den Satz bei, der das sagt (§5).
- Dateien, die kein Guide sind, und die Guide-Dateien außerhalb der Menge: Pläne (`docs/plan/planning/`), ADRs, Berichte, Spezifikation, Lastenheft, die vendorte Baseline, `harness/README.md` (wächst mit jedem Gate, `SPEC-040` Nr. 10 d), Hooks, `README.md` — sie wachsen mit dem Gegenstand oder sind kein Guide; eine Obergrenze für sie ist eine andere Entscheidung (`AGENTS.md` §3.14 begrenzt den Plan über die Größenmarke des Codes, nicht über seine Länge). Adresse der Wiedervorlage von `harness/README.md`: der Nutzer (§6, offener Punkt 3).
- Die Annahme von [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (Status Proposed): sie steht beim Nutzer (§6, offener Punkt 1); dieser Slice setzt sie nicht.
- Bytes je Hard Rule und Zeichen oder Token als Einheit — das Werkzeug zählt Bytes nur je Datei; `SPEC-040` Nr. 10 e und f nennen Grund und Folge.
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

- [ ] Die Module `file` und `structure` sind aktiv in `.d-check.yml` für die Prüfmenge nach `SPEC-040` Nr. 2 (13 Dateien in 5 Regeln) und die Hard Rules von `AGENTS.md` (Nr. 9), mit Grenzen in Zeilen und Bytes je Gruppe auf dem gemessenen Ist-Stand; jede Grenze trägt ihre Messung im Kommentar der Konfiguration (Klasse *Grenze* und *Rang-Zeiger* nach `AGENTS.md` §3.7), und `make docs-check` ist grün.
- [ ] Die Wirkung ist erprobt und belegt: die Fälle 156 bis 170 von `SPEC-102` zeigen, dass eine um ein Byte und um eine Zeile verlängerte Guide-Datei das Gate rot macht und eine Datei an der Grenze grün bleibt, über das Binary `d-check` aus dem Image des Pins (ohne Docker auf dem Host); die Zusatzzeit der Gegenproben ist gemessen und in §7 berichtet.
- [ ] Die Regel zum Anheben und Senken der Grenze steht in `SPEC-040` Nr. 8 und in [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (Status Proposed, die Annahme beim Nutzer); die Spalte *Vertrag* der Zeile von `make docs-check` in `harness/README.md` nennt die Module, ohne die 200 Zeichen zu überschreiten.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `SPEC-040`, `SPEC-102` und [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) samt Index-Zeile (vom Architect, vor dem ersten Code-Commit), Zeilen von `make docs-check` und `make gegenproben` und *Nicht behauptet* in `harness/README.md` (Wortlaut `SPEC-040` Nr. 11 und §3 unten); die Closure legt dem User den Satz bei, dass der Zeiger in `AGENTS.md` entfällt und jede Erweiterung einer Guide-Datei eine Verschiebung braucht (`SPEC-040` Nr. 8 b); ein Diff-Vorschlag für `AGENTS.md` entfällt nach der Empfehlung des Architect.
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
| `.d-check.yml` | update | `file` am Ende von `modules:`; Block `file:` mit fünf Regeln (13 Dateien) und Eintrag in `structure:` (`### 3.<n>`, `max-lines`); Werte am Start neu gemessen, Kommentar nennt je Gruppe Datei, Zeilen, Bytes und Stand (`SPEC-040` Nr. 4 b); kein `hint`; der Kommentar zur Position des Moduls `structure` nennt die zweite Regel mit |
| `harness/vertraege/SPEC-040-docs-check.md`, `harness/vertraege/SPEC-102-gegenproben.md`, `docs/plan/adr/0010-groesse-der-guide-dateien-als-gate.md`, `docs/plan/adr/README.md` | **geliefert (Architect, Planner-Commit vor dem Code)** | der Vertrag, die Fälle 156 bis 170, die ADR (Proposed) und die Index-Zeile; der Implementer zieht `Stand` und `Grenze` von `SPEC-040` und `SPEC-102` nach, wenn sie geliefert sind |
| `harness/README.md` | update | Zeile `make docs-check` (Vertrag: `Doku-Referenzen, Spec-Straten-Richtung und Größengrenzen der Guide-Dateien (d-check: file, structure; SPEC-040)`, Bindung [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) als Link); Zeile `make gegenproben` (Spalte *Tut was* ≤ 200 Zeichen, nennt `docs-check`; der Implementer zählt); *Nicht behauptet*: Gegenproben zu `docs-check` außer den Größengrenzen (`SPEC-040` Nr. 11), `harness/README.md` und die vendorte Baseline außerhalb der Größengrenzen (`SPEC-040` Nr. 10 d) |
| `Dockerfile`, `.dockerignore` | update | `ARG D_CHECK_IMAGE=scratch`, Stufe `d-check-bin` neben `a-check-bin`, in der Stufe `gegenproben` `COPY --from=d-check-bin /d-check /usr/local/bin/d-check`; die Zeile `COPY go.mod go.sum* .a-check.yml .golangci.yml ./` nimmt `.d-check.yml` mit; `!.d-check.yml`; Kommentar der Stufe nennt `docs-check` |
| `harness/mk/gegenproben.mk`, `tools/harness/gegenproben-eingabe.sh` | update | `--build-arg D_CHECK_IMAGE=$(DCHECK_REF)`; `"$(DCHECK_REF)"` als dritter Pin an `gegenproben-eingabe.sh`; `.d-check.yml` in `EINGABEN` (Pflicht-Datei); Kopf- und Hilfetext nennen `docs-check` |
| `tools/gegenproben/` | update | `gateDCheck` in `faelle.go`/`main.go`, `guide.go` mit den Fällen 156 bis 170, `guide_test.go` (Zahl, Namen, Reihenfolge; die Fälle 166 bis 170 verwenden `{dir}` und die echte Konfiguration; Muster `zeichen_test.go`); stdout für Befunde, stderr für die Summenzeile getrennt |
| `harness/gegenproben.stempel` | update | nach dem letzten Commit unter den Eingabepfaden |
| Dateien unter `.claude/` und `.harness/skills/` | **nicht geändert** | nur gemessen; das Modul erreicht sie trotz `scan.ignore` (`SPEC-040` Nr. 6 d, am Pin belegt) |

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Wirkung des Moduls am Image gelesen und belegt, Aufträge 1 bis 7, Tabelle *Zusage · Fänger*, Raster, Nachzüge, erwartete Größe) → Planner (Marke mit `make groessenmarke`, eigener Commit, vor dem ersten Code-Commit) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure; Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills nur als Diff-Vorschlag an den User).

**Betriebsregeln für alle Rollen** (Wortlaut und Gründe: [welle-sensoren](../welle-sensoren.md) §4):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`, höchstens zwei Container gleichzeitig; `make gates` nie parallel zu Mutationsläufen; Docker sonst nur über `make` (`AGENTS.md` §3.1). Gate-Läufe laufen im Hauptbaum, weil die Images feste Tags tragen. Das Doku-Gate selbst fährt in einem per Digest gepinnten Image mit schreibgeschütztem Mount (`:ro`).
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- **Eine Grenze wird nie angehoben, um den Slice grün zu bekommen** (`AGENTS.md` §3.6): wächst eine Guide-Datei während des Slice (der Slice selbst ändert keine), meldet der Implementer es dem Planner, statt die Grenze zu setzen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen.
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice.

**Entscheidungen des Architect (Aufträge 1 bis 7, alle beantwortet am 2026-10-31, Stand der Messungen `4e36c01`):**

1. **Träger:** die Module `file` und `structure` in `.d-check.yml` unter `make docs-check`; kein neues Ziel, kein Skript (`SPEC-040` Nr. 1, [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md)). Das Modul erreicht `.claude/**` und `.harness/skills/**` trotz `scan.ignore`; Einheit Bytes, Zeilen gleich `wc -l` plus unvollständige Schlusszeile; gleich der Grenze grün, eins darüber rot (Nr. 3, 5, 6).
2. **Menge:** `AGENTS.md`, `.claude/commands/*.md`, `.claude/agents/*.md`, `.harness/skills/*.md`, `harness/conventions.md`; `harness/README.md` nicht, mit Grund und Folge (Nr. 2, 10 d).
3. **Grenzen:** Ist-Stand ohne Spielraum, beide Maße, die Gruppen auf dem Maximum ihrer Dateien; Senken frei; Anheben nur als Verschiebung (die Summe der Grenzen desselben Maßes steigt nicht) oder durch eigene ADR; die Klasse „Verschiebung“ genehmigt [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) einmal (Nr. 4, 8).
4. **Hard Rules:** tragbar, `structure` mit `max-lines` je Abschnitt `### 3.<n>` von `AGENTS.md` (Maximum heute §3.7 mit 29 Zeilen, 29 grün und 28 rot belegt); Anheben nur per ADR (Nr. 9).
5. **Nachweis des Rots:** das Binary `/d-check` des gepinnten Images wird wie `a-check` in die Stufe `gegenproben` kopiert; kein Docker auf dem Host, kein Treiber; der Pin kommt in die Eingaben der Prüfsumme; Zeitbudget Schätzung 0 bis 2 s (`SPEC-102`).
6. **Tabelle, Raster, Nachzüge, ADR:** unten und in §6; die ADR ist [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (Proposed).
7. **Größe:** a = 0, b = 15, f = 0, p = 70 (§4).
- **Diff-Vorschlag für `AGENTS.md`:** keiner empfohlen (`AGENTS.md` §4 hält den Gate-Index einmal in `harness/README.md`; eine Zeile mehr wäre über der Grenze).

**Aufträge an den Architect, Wortlaut der Frage** (beantwortet oben; Entscheidungen und Texte, die kein Plan trifft; Wirkung des Moduls: das Handbuch des gepinnten Images, das der Architect über `make docs-check` und die Hilfe des Images liest, nicht aus dem Gedächtnis; der Vertrag steht in `harness/vertraege/`, [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), `AGENTS.md` §3.8):

1. **Die Wirkung des Moduls `file` belegen.** Schlüssel der Konfiguration (`max-lines`, `max-bytes`, Dateimenge als Pfad oder Glob), Fehlercodes (`file-lines-exceeded`, `file-bytes-exceeded`), Einheit (Bytes oder Zeichen; Umlaute zählen verschieden: 24 648 Bytes gegen 24 193 Zeichen bei `AGENTS.md`), Zeilenzählung (letzte Zeile ohne Zeilenende), und vor allem: Erreicht das Modul Dateien unter `.claude/` (verstecktes Verzeichnis) und unter `.harness/skills/`, obwohl `scan.ignore` in `.d-check.yml` `.harness/**` ausnimmt? Ist die Wirkung am Pin nicht so, wie das Handbuch sagt, oder erreicht das Modul diese Dateien nicht, entscheidet der Architect mit Grund und Folge: anderer Träger (ein Skript unter `tools/harness/`, ein Ziel in `GATE_CHECKS`), oder ein benannter Rest unter *Nicht behauptet*. Das Ergebnis steht vor dem ersten Code-Commit im Vertrag (`AGENTS.md` §3.12).
2. **Die Dateimenge** der Guide-Dateien mit Grund je Gruppe: `AGENTS.md`, `.claude/commands/*.md`, `.claude/agents/*.md`, `.harness/skills/*.md`, und ob `harness/README.md` und `harness/conventions.md` (jeder Lauf liest sie laut Leseordnung) dazugehören. Eine Datei, die dazugehören müsste und nicht erreicht wird, ist ein benannter Rest, keine Lücke im Stillen.
3. **Die Grenzen und ihre Regel.** Ist-Stand genau oder mit Spielraum (je Datei), Zeilen und Bytes beide oder nur eines; wie eine Grenze wandert: **Senken** ist eine Verschärfung (ein Commit mit der Messung, der Verschlankung nachgelagert), **Anheben** ist eine Lockerung im Sinn von `AGENTS.md` §3.6 (ADR) oder, wie der User es für die Welle vorgesehen hat, ein sichtbarer Commit mit Begründung — der Architect löst den Widerspruch beider Lesarten auf und nennt Grund und Folge: eine ADR je Anhebung, eine ADR einmal für die Ratsche, oder ein Carveout mit Frist. Ob die Grenze am Ist-Stand eine Verschlankung erzwingt, bevor jemand `AGENTS.md` um eine Zeile erweitern kann: ja, und das ist gewollt; die Folge (jeder Diff-Vorschlag des Users, der `AGENTS.md` verlängert, bricht das Gate) steht im Vertrag.
4. **Grenze je Hard Rule (`structure` mit `max-lines` je Abschnitt `### 3.<n>`)** — tragbar oder nicht, am Image belegt (Handbuchtext: „nicht erprobt“, Retirement-Check Abschnitt 5); wenn ja, mit Wert und Grund, wenn nein, als benannter Rest mit Folge.
5. **Der Nachweis des Rots.** Kopie des Baums mit verlängerter Guide-Datei in den Gegenproben ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) — der Fall fährt `docs-check` (Docker, gepinntes Image, `:ro`, netzlos) auf der Kopie: wie, ohne dass die Stufe `gegenproben` Docker aufruft (Treiber auf dem Host wie bei `groessenmarke-docker.sh`/`schema.mk`, oder ein anderer Weg), und was der Fall kostet (Zeitbudget in `SPEC-102`; `docs-check` allein dauert wenige Sekunden, ein Image-Start je Fall zählt). Das Fehlen einer Gegenprobe zu `make docs-check` ist heute in `harness/README.md` *Nicht behauptet* genannt; dieser Slice schließt den Teil, den er trägt, und nennt den Rest.
6. **Tabelle *Zusage · Fänger*** (§3), **Raster** (§6), **Nachzüge** nach `AGENTS.md` §3.9 (`SPEC-040`, `harness/README.md`, Gegenproben-Eingaben) und ob eine ADR zum Gate nötig ist (Muster: Bindung-Spalte; neue Gates sind erlaubt, eine Regel zum Anheben der Grenze ist eine Entscheidung, Auftrag 3).
7. **Erwartete Größe** (Zeilen Konfiguration und Skript, falls einer entsteht; Fall der Gegenprobe) für die Marke des Planners; ob ein Teil nebenläufig ist (zählt × 85).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; vom Architect, 15 Zeilen, alle Klasse `andere`: keine Zeile ist nebenläufig, kein Fänger hat eine Rate. Fänger ist ein Fall der Gegenprobe von `SPEC-102` (Nummer), ein akzeptiertes Negativ oder eine benannte Ausgabe des Gates. Die Fälle 156 bis 165 prüfen die Wirkung des Moduls am Pin mit einer Konfiguration des Falls und kleinen Grenzen (5 Zeilen, 40 Bytes, Hard Rule 6 Zeilen); die Fälle 166 bis 170 prüfen die echte Konfiguration mit Stubs. Die Spalte *Mutation* nennt die Mutation, die den Fänger rot machen muss; der Implementer trägt sie mit Raten, Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Die Module `file` und `structure` sind in `.d-check.yml` aktiv und tragen ihre Blöcke (`SPEC-040` Nr. 1 a) | Fall 170 (die echte Zeile `modules:` nennt `file` und `structure` als Wörter, sonst rot mit Text) und Fall 167 (ohne Block kein Befund) | `file` aus der Zeile `modules:` genommen; Block `file:` gelöscht; Eintrag in `structure:` gelöscht | andere |
| 2 | Jede Datei der Prüfmenge (13 Dateien in 5 Regeln) ist an beide Grenzen gebunden (Nr. 2, 3, 5) | Fall 167: 13 Stubs (`AGENTS.md`, drei Commands, sechs Typen, zwei Skills, `conventions.md`), jeder mit 50 000 Zeilen zu 16 Bytes; die Erwartung ist eine Schleife über die 13 Pfade mal die zwei Codes, je als Text `<Pfad>:1<TAB><Glob>...<Code>` | Regel `.claude/agents/*.md` entfernt; `max-lines` einer Regel gelöscht; Glob `.harness/skills/*.md` auf `reviewer.md` verengt; Glob `harness/conventions.md` auf `harness/*.md` ohne Treffer | andere |
| 3 | Eine neue Datei in einem Glob erbt die Grenze der Regel (Nr. 6 e) | Fall 163 (Fixture: `neu-gross.md` mit 41 Bytes rot, `neu-klein.md` mit 40 grün in drei Gruppen) und Fall 168 (echt: `neu.md` mit 50 000 Zeilen in `.claude/commands/`, `.claude/agents/`, `.harness/skills/` rot, `neu-klein.md` kein Befund) | `.claude/commands/*.md` durch die drei Dateinamen ersetzt | andere |
| 4 | Gleich der Grenze ist grün, ein Byte darüber rot, eine Zeile darüber rot, und das Maß ist im Code genannt (Nr. 3, 4 c, 5) | Fall 156 (vier Dateien exakt 5 Zeilen und 40 Bytes: Exit 0), Fall 157 (41 Bytes bei 5 Zeilen: nur `file-bytes-exceeded`, Text `Datei hat 41 Bytes, erlaubt sind 40`, nicht `file-lines-exceeded`), Fall 158 (6 Zeilen bei 36 Bytes: nur `file-lines-exceeded`) | Pin auf eine Version, die `>=` vergleicht; Pin, der die Maße vertauscht oder eines nicht prüft | andere |
| 5 | Die Einheit ist Byte, nicht Zeichen (Nr. 3) | Fall 160: Datei aus 21 Zeichen U+00E4 (Bytes `C3 A4`) und Zeilenvorschub, 43 Bytes bei 22 Zeichen und Grenze 40: `Datei hat 43 Bytes, erlaubt sind 40` | Pin, der Zeichen zählt | andere |
| 6 | Die Zeilenzählung nimmt die letzte Zeile ohne Zeilenvorschub mit; die leere Datei hat 0 Zeilen, `"\n"` eine (Nr. 3, 6 b, 6 c) | Fall 161: vier Dateien in vier Gruppen (5 Zeilen ohne Zeilenvorschub am Ende grün, 6 Zeilen ohne Zeilenvorschub rot mit `Datei hat 6 Zeilen`, leere Datei grün, `"\n"` grün): genau ein Befund, Summenzeile `1 Befund(e)` | Pin, der nur Zeilenvorschübe zählt (6 Zeilen ohne Schluss-LF grün) | andere |
| 7 | Überschreitet eine Datei beide Maße, sind es zwei Befunde; überschreiten mehrere Dateien einer Regel, ist es ein Befund je Datei und Maß (Nr. 5) | Fall 159: zwei Dateien in `.claude/commands/`, beide über beiden Maßen: vier Befunde, `4 Befund(e)` | Pin, der nach dem ersten Befund einer Datei oder einer Regel abbricht | andere |
| 8 | Fehlt die Datei einer Regel, ist das Ziel rot (Nr. 6 a) | Fall 162: `AGENTS.md` fehlt, `.claude/agents/` ist leer: zwei Befunde `file-no-match`, Exit 1 | Pin ohne Nullmengen-Härte (Regel ohne Treffer grün) | andere |
| 9 | Dateien in versteckten Verzeichnissen und unter `scan.ignore` gehören zur Menge (Nr. 6 d) | Fälle 157 und 158 (Pfade `.claude/...` und `.harness/skills/...` bei `ignore: [".harness/**"]`: je ein Befund für alle vier Pfade) und Fall 167 (echte `scan`-Zeile mit `.harness/**`, Skills rot) | `.harness/skills/*.md` im Pin durch `scan.ignore` ausgeblendet; verstecktes Verzeichnis übersprungen | andere |
| 10 | Die Meldung nennt Ist und Grenze; kein `hint` ersetzt sie (Nr. 4 d, 5) | Fälle 157, 158, 160, 161 (Text mit beiden Zahlen) und Fall 167 (`erlaubt sind` in jeder der 26 Zeilen) | `hint: "..."` in einer Regel | andere |
| 11 | Jeder Abschnitt `### 3.<n>` von `AGENTS.md` ist an `max-lines` gebunden: gleich grün, eins darüber rot, Befund auf der Überschriftszeile, andere Abschnitte (`## 4.`) nicht gezählt (Nr. 9) | Fall 164 (Fixture mit `max-lines: 6`: `### 3.1` mit 6 Zeilen grün, `### 3.2` mit 7 rot, `### 3.3` mit 3 Zeilen grün, `## 4.` mit 500 Zeilen ohne Befund: genau ein Befund, Fundstelle ist die Zeile von `### 3.2`) und Fall 167 (Stub-Abschnitt mit 50 000 Zeilen rot, neben dem `file`-Befund derselben Datei im selben Lauf) | `sections: one` statt `each`; Muster auf `^### 3\.1 `; `max-lines` entfernt; Muster `^#{2,3} 3\.` (trifft `## 4.` nicht, aber `### 3.x` anders) | andere |
| 12 | Trifft das Muster keinen Abschnitt, ist das Ziel rot (Nr. 9) | Fall 165 (Fixture) und Fall 169 (echt): `AGENTS.md` ohne `### 3.<n>`: `section-missing` | Muster auf eine Überschrift, die nicht vorkommt, ohne dass etwas rot wird (Pin ohne Härte) | andere |
| 13 | Die Echt-Konfiguration ist im Ausgangszustand grün auf kleinen Stubs (kein Dauerrot durch den Aufbau) | Fall 166: 13 Stubs je `x` plus `### 3.1` in `AGENTS.md` und ein `harness/README.md` mit der Tabelle `Target · Vertrag · Tut was · Bindung`: Exit 0, `0 Befund(e)` | Stub-Aufbau, der die Spalte `Tut was` auslässt (`section-column-missing`); eine Regel, die auf kleinen Dateien rot ist (Grenze 0) | andere |
| 14 | akzeptiertes Negativ: die Höhe der Grenze und die Regel zum Anheben prüft kein Gate (Nr. 8 e, 10 a) | grob Fall 167 (eine Grenze über 50 000 Zeilen oder 800 000 Bytes macht ihn rot); sonst nur der Reviewer am Diff von `.d-check.yml`; Folge: eine hoch gesetzte oder nicht gedeckte Anhebung unter dieser Schwelle passiert; Grund: ein Gate „Grenze gleich Ist“ erzwänge das Senken im Commit der Kürzung | Grenze von `AGENTS.md` ohne Verschiebung um 100 Bytes angehoben (kein Fall rot) | andere |
| 15 | akzeptierte Negative: Verdrahtung der Gegenprobe (Binary, `EINGABEN`), Verhalten des Pins bei Konfigurationsfehlern und im zweiten Lauf, Datei in zwei Regeln, Symlinks (Nr. 10 b, c, g, h) | kein Fall; Grund und Folge im Vertrag; Beleg beim Handoff: die Prüfsumme von `gegenproben-eingabe.sh` ändert sich mit `.d-check.yml`; ohne Binary sind alle 15 Fälle rot | `.d-check.yml` aus `EINGABEN` genommen; `COPY --from=d-check-bin` entfernt (alle Fälle rot) | andere |

Absatz *Mutationen und Raten* (`AGENTS.md` §3.10): vom Implementer vor der Übergabe zu füllen (Zusage · Mutation · roter Test, Raten mit Zahl der Prozesse und Messendem; keine Zeile ist nebenläufig, ein Fänger ist ein Lauf des Pins und deterministisch).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-unsichtbare-zeichen` liegt in `done/` (Reihenfolge der Welle: beide berühren `harness/README.md`, die Gegenproben und den Stempel), und der Architect hat die Aufträge 1 bis 7 entschieden, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Größenmarke** (`AGENTS.md` §3.14): Metrik: die Summe aus Einfügungen und Löschungen von `git diff --numstat <basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` zählt nie ([SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md) Nr. 4 a). Der Block unten trägt den Teil des Architect (`pfad`, `produktcode`, `fakes`): a = 0 Zeilen `nebenläufig`, b = 15 Zeilen `andere` (die zwei akzeptierten Negative, Nr. 14 und 15 der Tabelle in §3, zählen als Zeilen, `SPEC-128` Nr. 3 e), f = 0 (keine Port-Fakes, kein Test-Server), p = 70 (Produktcode des Architect: `.d-check.yml` rund 45 Zeilen samt Kommentar, `Dockerfile` 6, `.dockerignore` 1, `harness/mk/gegenproben.mk` 5, `tools/harness/gegenproben-eingabe.sh` 2, `harness/README.md` 6, Stempel 2; kein Programm; die Fälle in Go, rund 250 Zeilen, deckt die Formel über b). Den Teil des Planner (`basis`, `schaetzung`, `marke`, `SPEC-128` Nr. 3 k) hat der Planner am 2026-10-31 im eigenen Commit vor dem ersten Code-Commit gesetzt: `basis` ist der Plan-Commit davor (Spec, ADR und Plan), mit `make groessenmarke` geprüft; a und b hat er in der Tabelle in §3 nachgezählt (0 und 15 Zeilen, wie der Architect): *S* = 85·0 + 80·15 + 200·0 + 70 = 1270, *M* = 100·⌈135·1270 / 10000⌉ = 100·⌈17,145⌉ = 1800; keine Schnittfrage (135·1270 = 171 450 ≤ 300 000). Die Marke ist nach dem ersten Code-Commit fest (`SPEC-128` Nr. 5 c). Aufwand nach Einschätzung des Planners: S.

```groessenmarke
pfad: .d-check.yml
pfad: Dockerfile
pfad: .dockerignore
pfad: harness/mk/gegenproben.mk
pfad: harness/README.md
pfad: harness/gegenproben.stempel
pfad: tools/gegenproben
pfad: tools/harness/gegenproben-eingabe.sh
produktcode: 70
fakes: 0
basis: 3309cb4e56b5210b6496104c9520b3e91a01c5ee
schaetzung: 1270
marke: 1800
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Schätzung mal 1,35 liegt über 3000 (`AGENTS.md` §3.14, `SPEC-128` Nr. 1 Schnittfrage); sie liegt bei 1270 und ist weit darunter.
- Die frühere Bedingung „der Fall der Gegenprobe braucht einen Treiber, der Docker aus dem Host fährt“ entfällt: das Binary `d-check` liegt in der Stufe `gegenproben` (Architect, Entscheidung 5). Ebenso entfällt die Rückführung nach `open` („das Modul erreicht die Guide-Dateien nicht“): es erreicht sie (`SPEC-040` Nr. 6 d, am Pin belegt).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, `make gates` und `make gegenproben` grün im Hauptbaum, Review- und Verifikations-Report unter `docs/reviews/`, Closure-Notiz in §7. Der Lerneintrag nennt den Sensor mit `liegt in` (`.d-check.yml` oder das Ziel, mit dem Anker `seit welle-sensoren`), die Größen der Guide-Dateien zum Stand der Closure (Eingabe der Verschlankung; der Planner misst neu, offener Punkt 6 in §6) und den Satz an den User anstelle eines Diff-Vorschlags: „Der Zeiger entfällt (`AGENTS.md` §4); jede Erweiterung einer Guide-Datei braucht eine Verschiebung (`SPEC-040` Nr. 8)“.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Das Modul `file` ist am Pin nicht erprobt; es kann die Dateien unter `.claude/` und `.harness/skills/` nicht erreichen (`scan.ignore`, verstecktes Verzeichnis) oder anders zählen als das Handbuch sagt — **Ausgang:** entfallen (Architect am 2026-10-31): das Modul erreicht `.claude/**` und `.harness/skills/**` trotz `scan.ignore`, die Zählung stimmt mit `wc` überein (gleich der Grenze grün, eins darüber rot); Ort `SPEC-040` Nr. 3 und 6, Beleg die Läufe des Architect am Pin und die Fälle 156 bis 158 und 167.
- Ratsche ohne Entscheidung des Users über das Anheben: die Grenze am Ist-Stand färbt jede Erweiterung von `AGENTS.md` rot, auch eine berechtigte; ein Gate, das niemand anhebt, wird umgangen (`AGENTS.md` §3.2) oder gelockert — **Ausgang:** weiter offen. Entschieden ist die Regel (`SPEC-040` Nr. 8, [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) Proposed; Folge für Diff-Vorschläge Nr. 8 d: jede Verlängerung von `AGENTS.md` braucht eine Verschiebung). Offen bleibt die Annahme von [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md); Adresse: der User, offener Punkt 1 unten.
- Die Messung von Zeilen und Bytes ist Handarbeit, sobald die Guide-Dateien wachsen (Bytes in `.d-check.yml` führen den Stand des Tages): der Kommentar der Konfiguration trägt die Messung; ein Kommentar, der den alten Stand nennt, ist falsch (`AGENTS.md` §3.7) — **Ausgang:** entschieden (`SPEC-040` Nr. 4 b: die Werte und ihre Messung stehen einmal, im Kommentar neben dem Block, kein Zahlenwert in der Grenze des Vertrags); das Gate macht einen Kommentar mit zu kleinen Zahlen nicht rot (Nr. 10 a). Der Implementer misst am Start neu und trägt den Stand ein; die Closure misst neu und trägt die Zahlen in §7 neben die Grenzen ein. Möglich: eingetreten, falls der User eine Datei der Menge in `main` vor dem Merge ändert (offener Punkt 6).
- Laufzeit der Gegenproben: ein Fall mit Image-Start verlängert den Lauf, dessen Re-Evaluierungs-Trigger ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): Lauf über 30 s bei geänderter Eingabe; gemessen 56,9 s Schritt) beim User unentschieden ist — **Ausgang:** weiter offen. Entschieden ist der Weg: kein Image-Start je Fall (das Binary liegt in der Stufe), Schätzung 0 bis 2 s im Schritt (`SPEC-102` Zeitbudget); der Implementer misst vorher und nachher und berichtet in §7, er entscheidet den Trigger nicht. Neu: `.d-check.yml` ist Eingabe der Prüfsumme, jede Änderung daran, auch ein Kommentar, kostet einen vollen Lauf der Stufe. Adresse: der User → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (offener Punkt 5).
- Eine Gegenprobe, die `docs-check` fährt, braucht Docker auf dem Host in der Gegenproben-Kette; heute ruft die Stufe `gegenproben` kein Docker — **Ausgang:** entfallen (Architect): das statisch gelinkte Binary `/d-check` kommt aus dem Image des Pins in die Stufe `gegenproben`, die Fälle rufen es im Wegwerfverzeichnis; der Treiber entfällt. Die Verdrahtung (Kopie des Binary, `.d-check.yml` in `EINGABEN`) fängt kein Fall, ein akzeptiertes Negativ (Tabelle Nr. 15, `SPEC-040` Nr. 10 h).
- Neu: ändert der User eine Datei der Prüfmenge in `main` vor dem Merge, ist `docs-check` am Merge rot, bis der Planner neu misst — **Ausgang:** weiter offen; Adresse: der Planner (die Closure misst neu, `AGENTS.md` §3.9).

**Raster der Randformen** (`AGENTS.md` §3.16; vom Architect vor dem ersten Code-Commit gefüllt: jede Zelle nennt ihre Nummer in `SPEC-040` oder „gilt nicht“ mit Grund; eine leere Zelle ist ein Fehler des Plans; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Messen einer Datei gegen die Grenze | `SPEC-040` Nr. 6 a (`file-no-match`) | Nr. 6 b (0 Zeilen, 0 Bytes, grün) | Nr. 3, 4 c, 5, 6 c, 6 g (Grenze minus 1, gleich, plus 1 je Maß; Datei ohne Zeilenvorschub am Ende; Mehrbyte-Zeichen; jede Dateiart) | Nr. 6 a (eine unlesbare Datei ist `file-no-match`) | Nr. 10 g (zweiter Lauf: Eigenschaft des Pins, akzeptiertes Negativ) |
| Bestimmen der Dateimenge (Glob) | Nr. 6 a (Glob trifft nichts) | Nr. 7 und Nr. 1 a (Block `file:` ohne Eintrag oder fehlend prüft nichts: die Fälle 167 und 170 werden rot) | Nr. 6 d (Datei unter `scan.ignore`, in verstecktem Verzeichnis), Nr. 6 f und 10 c (Symlink), Nr. 6 e (neue Datei) | gilt nicht: reine Konfiguration, kein Lauf mit Zustand (ein Lesefehler ist Nr. 6 a) | Nr. 10 b (dieselbe Datei in zwei Regeln: beide gelten) |
| Lesen der Grenze (Konfiguration) | Nr. 7 (Exit 2) und Nr. 3 (beide Maße gesetzt) | Nr. 7 (Grenze 0 heißt: nur die leere Datei ist grün) | Nr. 7 (negativ, nicht numerisch, unbekannter Schlüssel: Exit 2) | gilt nicht: reine Konfiguration | Nr. 7 (gleicher Glob zweimal: Exit 2) |
| Anheben und Senken der Grenze | Nr. 8 b (Anhebung ohne Verschiebung oder ADR; kein Gate, Nr. 8 e, 10 a) | Nr. 8 a (Grenze gleich Ist ist der Normalfall, keine Anhebung) | Nr. 8 a (Senken unter den Ist-Stand färbt das Ziel an der eigenen Datei rot) | gilt nicht: ein Commit, kein Lauf | Nr. 8 b (derselbe Wert zweimal ist keine Anhebung; die Summe zählt den Zustand, nicht die Folge der Commits) |
| Meldung und Ausgang | Nr. 5 (keine Überschreitung: Exit 0) | Nr. 5 (nur ein Maß über der Grenze: ein Befund) | Nr. 5 (beide Maße; mehrere Dateien; je Regel ein Befund) | gilt nicht: d-check schreibt Befunde auf stdout, der Exit trägt das Ergebnis; ein Abbruch von `make` ist kein Befund dieses Vertrags | Nr. 10 g (gleiche Meldung beim zweiten Lauf) |
| Hard-Rule-Abschnitt (`structure`) | Nr. 9 (`section-missing`) | Nr. 9 (Abschnitt ohne Text zählt 1 Zeile, am Pin gemessen, grün) | Nr. 9 und 10 e (Abschnitt anderer Form wie `## 4.` oder `#### 3.1.1` bleibt ungezählt oder gehört zum Abschnitt darüber) | gilt nicht: reine Konfiguration | Nr. 9 (`each`: jeder Abschnitt gleichen Musters einzeln) |

**Zusatztabelle Paare und Modi** (Spalte *Fänger*; vom Architect: Zeilengrenze und Byte-Grenze derselben Datei, eine Datei in zwei Gruppen, `file` gegen `structure` an derselben Datei):

| Paar oder Modus | Entscheidung (Ort im Vertrag) | Fänger |
|---|---|---|
| Zeilen- und Byte-Grenze derselben Datei überschritten | zwei Befunde, kein Vorrang (Nr. 5) | Fall 159 (vier Befunde), Fall 167 (26 Zeilen) |
| nur eine der beiden Grenzen überschritten | genau ein Befund mit dem Code des Maßes (Nr. 5) | Fälle 157 (nur Bytes), 158 (nur Zeilen), 161 (nur Zeilen) |
| mehrere Dateien einer Regel über der Grenze | ein Befund je Datei und Maß (Nr. 5) | Fall 159 (zwei Dateien), Fall 167 (sechs Rollen-Typen) |
| dieselbe Datei in zwei Regeln | beide gelten; keine Regel dieser Konfiguration überlappt (Nr. 10 b) | akzeptiertes Negativ (Nr. 10 b): kein Fänger, weil keine Regel ihn auslöst; Folge: keine, solange kein Glob überlappt |
| `file` gegen `structure` an derselben Datei (`AGENTS.md`) | unabhängige Module, je eigener Befund: `AGENTS.md:1` und `AGENTS.md:<Überschriftszeile>` (Nr. 5, 9) | Fall 167 (beide im selben Lauf) |
| Modus mit und ohne `scan.ignore` | das Modul liest `scan.ignore` nicht (Nr. 6 d) | Fälle 157 (Fixture-`ignore`), 167 (echte `ignore`-Zeile) |
| Gruppen-Glob gegen Einzeldatei | Glob steht auf dem Maximum seiner Dateien, gilt für jede (Nr. 4 a) | Fall 167 (alle sechs Typen, beide Skills, alle drei Commands rot), Fall 163 |

**Offene Punkte mit Adresse** (`AGENTS.md` §3.13; Ausgang jeweils *offen*, bis die Adresse antwortet):

1. **[ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) annehmen** (Status Proposed wie [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md)) — Adresse: der User. Bis dahin trägt die Bindung in `harness/README.md` eine Proposed-ADR; Folge keine für das Gate.
2. **Diff-Vorschlag an `AGENTS.md`** — Empfehlung des Architect: keiner (Duplikat des Gate-Index nach `AGENTS.md` §4; eine Zeile mehr wäre über der Grenze). Wünscht der User einen Zeiger, braucht er eine Verschiebung nach `SPEC-040` Nr. 8 b. Adresse: der User; die Closure legt den Satz bei (der Planner).
3. **`harness/README.md` außerhalb der Menge** (`SPEC-040` Nr. 10 d) — wächst um rund eine Zeile je Gate (`slice-abdeckung-deklaration` und `slice-coverage-gate` fügen zwei hinzu; ihre §1 nehmen die Sendung nicht an, es ist eine Auskunft, kein Auftrag). Wiedervorlage: wenn die Datei über rund 160 Zeilen wächst oder der User sie unter ein Gate stellen will (dann mit Spielraum, ADR). Adresse: der User; ein Register-Eintrag erst beim zweiten Auffallen.
4. **Die zwei vendorten Baseline-Dateien in `.claude/rules/`** (19 877 Bytes, 28 % der Sitzungslast von 72 263 Bytes) sind nicht kürzbar und stehen außerhalb; ob sie weiter per Symlink geladen werden, ist ein Vorgang des Users (Entscheidung über `.claude/rules/`), nicht dieses Slice. Adresse: der User.
5. **Zeitbudget und Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md):** der Implementer misst (Schritt und Gesamtlauf, vorher und nachher, Zahl der Läufe, Zustand des Hosts); der Trigger bleibt beim User (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
6. **Nummern der Fälle und Merge-Zeitpunkt:** die Fälle 156 bis 170 von `SPEC-102` nummeriert der Stand beim Handoff (heute 155 in `main`); kollidiert eine Nummer beim Merge, nimmt der später mergende die nächste. [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) ist in keinem Zweig vergeben. Ändert der User eine Datei der Prüfmenge vor dem Merge, misst der Planner neu. Adresse: der Planner.
7. **Pin-Wert der Fälle:** der Wortlaut `Datei hat N Bytes, erlaubt sind M` und die Reihenfolge der Spalten gehören dem Pin; ein Pin-Wechsel kann die Fälle rot machen, das ist gewollt (der Pin steht in den Eingaben) und kein Befund.

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

*Wird bei der Closure gefüllt.* Die Closure trägt zusätzlich: die Größen aller Guide-Dateien (Bytes und Zeilen) zum Stand der Closure neben den Grenzen des Gates, die gemessene Zusatzzeit der Gegenproben (Schritt und Gesamtlauf, vorher und nachher, Zahl der Läufe, Zustand des Hosts) und den Satz an den User anstelle eines Diff-Vorschlags für `AGENTS.md` (§5).

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
