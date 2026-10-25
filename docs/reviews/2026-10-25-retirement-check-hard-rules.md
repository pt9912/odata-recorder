# Retirement-Check der Hard Rules AGENTS.md §3.9 bis §3.18

**Rolle:** Planner (frischer Kontext). **Stand:** `5bcf983` (HEAD beim Start), 2026-10-25.
**Art:** reine Analyse; geändert wurde weder `AGENTS.md` noch Regelwerk, Command, Typ oder Skill.
**Anlass:** Frage des Users, ob Regeln in `AGENTS.md`, die nicht funktioniert haben, wieder entfallen sollen; dazu die Fragen, ob §3.14 in `AGENTS.md` einen besseren Platz hat und ob ein Gate dazu fehlt. Der User hat dem Retirement-Check zugestimmt.
**Bezug:** `LH-FA-18`; Register-Eintrag `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`.

## 1. Zweck und Maßstab

Der Check beantwortet je Regel die Frage des Retirement-Checks (`grundlagen-traceability.md`, Sensor 2): Ist die Beobachtung, aus der die Regel entstand, seither wieder aufgetreten, und ist der Grund der Regel entfallen oder von einem anderen Träger gedeckt? Daraus folgt eine Empfehlung aus vier Ausgängen.

Maßstab aus dem Regelwerk (`.harness/baseline/v6.18.1/regelwerk/`):

- `grundlagen-klassifikation.md`: Faustregel „so weit links und oben wie möglich“ (computational vor inferential); Steering Loop 3× heißt „Guide oder Sensor nachziehen“; der Abschnitt *Guide-Datei-Wildwuchs* nennt zwei Mittel, die Obergrenze `file.max-lines`/`max-bytes` und den Index-plus-Datei-Schnitt.
- `grundlagen-harness-dateien.md` §Konventionsspeicher: kurze Zeile in `AGENTS.md`, Volltext in `harness/rules/<name>.md`.
- `modul-09-implementierung.md` §Hard Rules: `AGENTS.md` trägt Negativregeln mit Falsch/Richtig, keine Verfahren.
- `grundlagen-traceability.md` Sensor 2: Eine Regel mit Herkunfts-Anker wird nur entfernt oder gelockert, wenn die Herkunft konsultiert und das Ergebnis dokumentiert ist. Das ist mit diesem Bericht geschehen; ein Ausgang D (Entfallen) kommt dabei nicht vor (siehe 4.1).

Die vier Ausgänge:

| | Ausgang |
|---|---|
| A | bleibt in `AGENTS.md` |
| B | Volltext wandert in Rollen-Typ, Command, Skill oder `harness/rules/`; in `AGENTS.md` bleiben Überschrift, ein Satz und ein Zeiger (Index-plus-Datei) |
| C | wird Sensor oder Gate (Vorschlag mit Slice-Kennung); die Regel bleibt als Zeiger |
| D | entfällt |

## 2. Methode (womit gezählt)

Alle Zahlen stammen aus dem Repo-Stand `5bcf983`; Quelle in Klammern.

- **Herkunft:** `git log -S"### 3.<n> " -- AGENTS.md` für die erste Fassung; die Zahl der Fassungen aus einem Lauf über alle Commits, der den Abschnitt je Commit hasht (zählt Änderungen des Abschnitts, nicht des Gesamtdokuments). Register: `docs/plan/planning/observations/BEO-REPO/<slug>/` mit `state.md` (Zielort) und der Zahl der Dateien in `evidence/` (= Vorkommen, je Vorgang eine).
- **Zeitpunkt der Slices:** Datum des ersten Reports in `docs/reviews/` je Slice; für die Slices der Wellen `welle-fundament` und `welle-persistenz` (ohne Report in dieser Form) das Datum des Abschlusses laut `git log`. Das Anlegedatum einer Evidence-Datei ist **kein** Maß: mehrere wurden am 2026-10-21 gesammelt nachgetragen.
- **(a) diagnostiziert:** Vorgänge im Register, deren Slice **nach** der ersten Fassung (bei Schärfung: nach der letzten) liegt; dazu **Befundzeilen**: Tabellenzeilen in `docs/reviews/2026-10-2*.md` mit Kennung `F-…`/`V-…`/`N-…`, die `§3.<n>` oder `AGENTS.md 3.<n>` nennen (`grep -E`). Obergrenze der Wirkung: Ein Reviewer, der die Regel als Anker nennt, hat sie als Maßstab benutzt, ob sie verletzt war oder nur zitiert wurde, zeigt die Zeile nicht.
- **(b) verhindert:** nur gezählt, wenn ein Bericht oder Plan es belegt. Das Repo hält kaum Belege dafür; zwei Proxys sind gekennzeichnet: (i) Zahl der Plan- und Wellen-Dateien, die die Regel zitieren (`grep` über `open/`, `done/`, `welle-record-replay.md`; 25 Dateien); (ii) Form-Befolgung in den 16 Plänen unter `done/` (Tabelle *Zusage · Fänger*, Raster, `numstat`-Metrik).
- **(c) unklar:** Slices nach Einführung ohne Vorkommen. Ursache nicht trennbar: Regel hielt, Slice berührte sie nicht, oder der Reviewer sah nichts.
- **Leser:** `grep` nach `§3.<n>` und `AGENTS.md 3.<n>` in `.claude/`, `.harness/skills/`, `harness/`, `tools/`, `src/`, `docs/plan/planning/`.
- **Größe:** `LC_ALL=C.UTF-8 awk` über `AGENTS.md` (Zeichen) und `LC_ALL=C awk` (Bytes); Abschnitt = Überschrift bis zur nächsten, einschließlich Leerzeile.
- **Sensor-Bestand:** `harness/mk/*.mk`, `.claude/hooks/`, `tools/harness/`, `.d-check.yml`; Fähigkeit von d-check am Pin v0.86.1 aus dem Benutzerhandbuch des Images (`d-check -manual "Benutzerhandbuch: d-check"`, netzlos).
- **Nicht gemessen:** Verbrauch in Tokens je Lauf (`make span-report`: „0 von 224 Agent-Läufen trugen Verbrauchs-Zähler“). Aufwand S/M/L ist eine Schätzung des Planners, nicht gemessen.

Grenze, die für alle Zahlen gilt: Das Register zählt Vorgänge, nicht Wirkung. „Nach Einführung wieder aufgetreten“ widerlegt die Regel nicht (sie kann die Schwere gesenkt haben), und „nicht aufgetreten“ belegt sie nicht. Die Befundzeilen sind kein Trend der Schwere. Wo ich eine Ursache nenne, steht sie als Lesart des Planners, nicht als Messung.

## 3. Tabelle

Slices insgesamt: 25 (16 unter `done/`, 9 archiviert in `done/welle-fundament/` und `done/welle-persistenz/`); mit Report in `docs/reviews/` 16. „n/N“ heißt n Slices mit Vorkommen von N Slices nach der Einführung.

| Regel | Herkunft | (a) diagnostiziert | (b) verhindert (belegt / Proxy) | (c) | Leser außerhalb | Sensor möglich | Größe (Zeichen/Zeilen) | Empf. |
|---|---|---|---|---|---|---|---|---|
| §3.9 Plan folgt Korrektur | `e2a4898` (2026-10-18), geschärft `7484d1c` (2026-10-23, „seit welle-record-replay“); Register `berichtigung-erreicht-nicht-jede-stelle` (11), verkörpert §3.9 | 11 von 24 Slices nach erster Fassung; 6 von 6 Slices seit der Schärfung; 30 Befundzeilen in 12 von 16 Slices; zuletzt `slice-baseline-import` (4 Stellen) | belegt: keiner; Proxy: 11 von 25 Plänen zitieren sie | 13 | Reviewer-Skill: nein; Commands: nein | nur Teil: Prüfung „Commit ändert `src/` oder `harness/` eines laufenden Slice, aber nicht dessen Plan“ (M, Hook oder `commit-msg`; Fehlalarm bei Fix-Runden); „Zahl im Plan = Zahl im Test“ nicht berechenbar | 674 / 8 | A |
| §3.10 Mutation je Zusage | `e2a4898`; vier Fassungen (`ae4e2a6`, `7484d1c`, `98ad360`); Register `bericht-der-rolle-liegt-nicht-im-repo` (3), `nebenlaeufiger-lauf-nur-im-glueck-gruen` (7); offen ohne Verkörperung: `rate-ohne-detektor-…` (2), `panik-eines-fakes-…` (2), `zweig-ohne-lauf-…` (1) | Bericht-Teil: 3 Vorgänge (2 von 3 vor der Fassung vom 10-23); Nebenläufigkeit: 5 vor der Klausel (10-21 21:56), 2 danach (`slice-race-gate`, `slice-baseline-import`); 34 Befundzeilen in 9 von 16 Slices | belegt: für Datenrennen trägt `make test-race` ([ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)); Proxy: 22 von 25 Plänen zitieren sie | nicht ermittelt | Reviewer-Skill (MEDIUM „Rate nicht benannt“), `implement-slice` Schritt 19 (ohne Verweis auf §3.10, sondern auf §3.6, siehe 6.3), `implementer.md` | Datenrennen: ja, `make test-race` besteht; Mutationstest: L (kein Mutationswerkzeug im Image, `deps` brauchen Netz), nicht empfohlen | 2512 / 20 | B |
| §3.11 Kommentar sagt nur zu, was ein Test prüft | `e2a4898`; eine Fassung; kein Register-Eintrag (die Vorgänge laufen unter `vertrag-sagt-mehr-zu-als-ein-gate-prueft`) | 49 Befundzeilen in 11 von 16 Slices, zuletzt `slice-baseline-import`; Schwere überwiegend LOW | belegt: Verifikationen führen „Kommentare (§3.7, §3.11) ohne Befund“ als Prüfzeile (z. B. `2026-10-20-slice-redaction-match-key-verifikation.md`); Proxy: 7 von 25 Plänen | nicht ermittelt | Reviewer-Skill (Kategorie Kommentare), `implement-slice` Schritt 20 (nennt §3.7) | nein (inferential) | 364 / 7 | A |
| §3.12 Randformen vor dem Code | `e2a4898`; eine Fassung; Register `randform-erst-im-review-entschieden` (20), verkörpert §3.16 | 19 von 24 Slices nach Einführung (das Register zählt ab `slice-go-skelett-build-test`) | Proxy: 24 von 25 Plänen zitieren sie | 5 | Reviewer-Skill: nein; `architect.md`: nein | wie §3.16 | 722 / 9 | B |
| §3.13 Adresse nimmt an | `e2a4898`, Schärfung `ae69cae` (2026-10-25); Register `befund-an-rolle-ohne-artefakt` (3), verkörpert §3.13 | 3 von 24 Slices (`slice-record-service`, `slice-upstream-adapter`, `slice-odata-codec-json`); 3 Befundzeilen in 3 Slices; keine Befundzeile nach 10-23 | belegt: `slice-groessenmarke-werkzeug` §1 („§3.13 geprüft“) und die offenen Pläne `slice-protokolltreue`, `slice-http-adapter`, `slice-workspace-use-case`, `slice-bootstrap-m2` zitieren sie; Proxy: 17 von 25 | 21 | keiner | nein; ein Wechselseitigkeits-Test („A nennt B, B nennt A“) hat zu viel erlaubte einseitige Verweise (M, niedriger Nutzen bei 3 Vorgängen) | 1155 / 13 | B |
| §3.14 Größenmarke | `e2a4898` (Fassung „Marke nennt Metrik“), fünf Fassungen (`d143752`, `482fdcb`, `ae4e2a6`, `ae69cae`); Register `groessenmarke-ueberschritten-ohne-rueckfuehrung` (16), `geplant` → `slice-groessenmarke-werkzeug` | Überschreitung bei 11 von 17 Slices vor der Fassung vom 10-21 21:56 (65 %), bei 5 von 8 danach (62 %: `slice-session-use-case`, `slice-session-wartung`, `slice-odata-codec-csdl`, `slice-odata-codec-json`, `slice-baseline-import`); zuletzt 10027 gegen Marke 8200; 17 Befundzeilen in 10 von 16 Slices | belegt: Form wirkt, 16 von 16 Plänen unter `done/` nennen die Metrik (`numstat`) und eine Marke; die Rückführung greift nicht | 3 Slices unter der Marke seit 10-21 | `slice-groessenmarke-werkzeug` (nennt die Prosa in §1 und §3); `harness/vertraege/` nach [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) soll die Faktoren tragen | ja: `slice-groessenmarke-werkzeug` in `open/`, L (Plan ohne Zeilen-Schätzung in §4; Architect setzt Formel und Ort in der Kette) | 2644 / 19 | C |
| §3.15 Zusage und Fänger als Paar | `a8d44ea` (2026-10-19, „seit welle-fundament“), fünf Fassungen; Register `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25), `akzeptiertes-negativ-ohne-folge` (10) | Vorkommensrate je Slice unverändert: 4 von 4 Slices vor der Regel, 21 von 21 danach; 146 Befundzeilen in 13 von 16 Slices (kein Rückgang: `slice-odata-codec-csdl` 25, `slice-baseline-import` 15) | belegt: Form 16 von 16 Plänen (Tabelle *Zusage · Fänger*); Proxy: 24 von 25 Plänen | 0 | Reviewer-Skill (drei MEDIUM-Zeilen), `implement-slice` (Schritt 14 „Tabelle *Zusage · Fänger*“) | Form ja, Inhalt nein: `structure` von d-check (Spalte *Fänger* nicht leer, S, nicht erprobt); ob der Fänger greift, entscheidet nur die Mutation | 1657 / 15 | B |
| §3.16 Randform-Raster | `f855f01` (2026-10-20, „seit welle-persistenz“), drei Fassungen; Register `randform-erst-im-review-entschieden` (20), `universalaussage-im-vertrag-ungeprueft` (9) | Rate unverändert: Randform im Review entschieden bei 7 von 9 Slices vor der Regel (78 %), 13 von 16 danach (81 %); Allaussage: 9 von 9 Vorkommen liegen nach dem Start der Regel; 56 Befundzeilen in 12 von 16 Slices | belegt: Form: das Wort „Raster“ steht in 16 von 16 Plänen, eine Tabelle in §6 in 15 von 16 (ohne `slice-gegenproben-build-test`); Proxy: 24 von 25 Plänen zitieren die Regel | 3 | Reviewer-Skill (MEDIUM „Paar von Ursachen ohne Fänger“) | Teil: `structure` auf Plan-§6 („keine leere Zelle“, Spalte *Fänger*), S bis M, nicht erprobt; „Ort der Zelle trägt die Entscheidung“ nicht berechenbar | 1364 / 11 | B |
| §3.17 Fix deckt die Klasse | `ad11657` (2026-10-24); eine Fassung; Register `fix-deckt-die-genannte-mutation-nicht-die-klasse` (5), zugleich verkörpert in `implement-slice` Schritt 19 und `verifier.md` | 2 von 2 Slices nach der Regel (`slice-odata-codec-json` teils, `slice-baseline-import`: drei Runden, 10 genannte und 6 nicht genannte Mutationen); 17 Befundzeilen in 2 Slices | belegt: dritte Runde von `slice-baseline-import` deckte die Klasse (Zustandsautomat über 1 022 490 Fälle) | 0 | `implement-slice` Schritt 19, `verifier.md` (Duplikat des Inhalts) | nein (inferential) | 1544 / 11 | B |
| §3.18 Rohbytes statt Escape | `ae69cae` (2026-10-25, „seit welle-record-replay“); Register `werkzeug-eingabe-setzt-escape-in-zeichen-um` (4) | 1 Vorgang nach Einführung (`slice-baseline-import`), und dort „galt und hielt“ (Quelltexte gegengelesen, Fehlgriff nur im Probe-Paket des Reviewers); 9 Zitate, 0 Befundzeilen | belegt: 1 (`slice-baseline-import`, Evidence-Datei); Proxy: 3 von 25 Plänen, 2 Test-Dateien in `src/` zitieren sie | 0 | `src/hexagon/services/baselineimport/anfrage_test.go`, `manifest_test.go` | Teil: Prüfung auf unsichtbare Zeichen (BOM, NBSP, U+2028/2029, U+200B) in `src/**/*.go` und `tools/**` in `tools/harness/lint.sh` (S); die Umsetzung eines gemeinten Escape in ASCII ist nicht erkennbar | 953 / 9 | A |

Gerundete Teilsummen: §3.9 bis §3.18 = 13 589 Zeichen in 122 Zeilen; `AGENTS.md` insgesamt 24 193 Zeichen (24 648 Bytes) in 337 Zeilen; die Abschnitte §3.1 bis §3.8 zusammen 4 783 Zeichen.

### 3.1 Kurz: §3.1 bis §3.8 (nur ob sie ein Gate oder einen Hook trägt)

| Regel | Träger im Repo | Befund |
|---|---|---|
| §3.1 Docker-only | `.claude/hooks/pretooluse-command-guard.sh` (nennt die Regel im Kopf), Makefile und Dockerfile; nicht erfasst sind `perl`, `/usr/bin/env python3` (laut Kopf des Hooks) | computational getragen, Regel bleibt Zeiger |
| §3.2 Suppression-Verbot | `tools/harness/lint.sh` (`//nolint`, `# Why:`), `nolintlint` in `.golangci.yml` ([ADR-0003](../plan/adr/0003-lint-profil-golangci.md)); `d-check:ignore` mit Grund | computational getragen |
| §3.3 git mv und Inhalt | `make slice-mv` (`tools/harness/slice-mv.sh`, reiner Move); kein Gate prüft Mischcommits | teils getragen |
| §3.4 Architektur meilensteinfrei | `spec-straten` mit `direction: no-downward` und Klasse `aussen` in `.d-check.yml`, in `make docs-check` | computational getragen |
| §3.5 ADR nach Accepted immutable | d-check-Ziel `doc-immutable` mit `history-range-guard`; **nicht** in `GATE_CHECKS`, weil die Range der Aufrufer setzt (`harness/mk/doc-gate.mk`) | Träger vorhanden, kein Gate der Kette |
| §3.6 Gates nicht lockern | kein Träger; `gegenproben` fangen Wächter-Fälle, nicht die Senkung einer Schwelle | inferential |
| §3.7 Kommentar beschreibt, was da ist | `implement-slice` Schritt 20, Reviewer-Skill | inferential |
| §3.8 ADR ohne Regeldetails | kein Träger (Architect-Typ nennt sie nicht) | inferential |

Für diese acht Regeln empfehle ich nichts; sie stehen nicht im Auftrag.

## 4. Auswertung

### 4.1 Hat eine Regel den Retirement-Test (Ausgang D) bestanden?

Nein. Bei **keiner** der zehn Regeln ist der Grund entfallen oder von einem anderen Träger gedeckt:

- Neun von zehn Regeln haben nach der Einführung neue Vorgänge im Register oder Befundzeilen. Ausnahme ist §3.18: ein Vorgang, den die Regel nach Bericht abgefangen hat, und keine Befundzeile. Bei §3.13 ist das letzte Vorkommen `slice-odata-codec-json` (2026-10-24/25), also vor der Schärfung vom 2026-10-25; seither keines.
- Wo ein Sensor besteht, deckt er einen Teil (§3.10 Datenrennen durch `make test-race`), nicht die Zusage.
- Die Regeln, die den Maßstab des Reviews bilden (§3.15, §3.16, §3.11), werden in den Reports hundertfach als Anker benutzt (188, 113, 101 Zitate). Entfiele die Regel, verlöre der Reviewer den Anker für LOW/MEDIUM-Befunde (`.harness/skills/reviewer.md`, LOW verlangt einen Konventions-Anker).

### 4.2 Was „hat nicht funktioniert“ hier heißt

Für drei Regeln ist die Rate **unverändert** nach der Einführung, gemessen als Anteil der Slices mit Vorkommen:

| Regel | vor | nach | Lesart |
|---|---|---|---|
| §3.15 | 4 von 4 Slices | 21 von 21 | Das Paar *Zusage · Fänger* steht in jedem Plan (Form 16 von 16); ob der Fänger greift, findet weiter erst der Reviewer. Ursache: Regel ist Urteilssache, kein Sensor fängt „Fänger greift nicht“. |
| §3.16 | 7 von 9 (78 %) | 13 von 16 (81 %) | Das Raster steht in 15 von 16 Plänen; die Zellen werden trotzdem im Code entschieden oder tragen keinen Ort. Ursache: Form erzwungen, Inhalt nicht. |
| §3.14 | 11 von 17 (65 %) | 5 von 8 (62 %) | Die Prosa schätzt und misst, erzwingt keine Rückführung; die Formel trifft nicht (json 120,6 %, import 122 % der Marke). Ursache: Regel schreibt eine Handlung vor, die nur ein Werkzeug rechnen kann. |

Für §3.9 gibt es kein „vor“ (die erste Fassung ist die erste Handlung); 6 von 6 Slices seit der Schärfung zeigen das Vorkommen weiter. Ursache nach Lesart des Planners: Die Regel verlangt, alle Stellen einer Aussage zu finden; das ist der Fall, den ein Prüfer in frischem Kontext findet und der Autor nicht.

Für §3.17 und §3.18 ist der Zeitraum zu kurz (1 bis 2 Slices); §3.18 zeigt einen belegten Treffer.

Die **Ursachen** teilen sich so (Lesart des Planners aus den Evidence-Dateien, nicht gezählt):

- *Regel nicht befolgt*: selten belegt. Die Formen (Tabelle, Metrik; Raster bei 15 von 16) stehen in den Plänen; ein Verstoß ist meist die Lücke im Inhalt, nicht das Fehlen der Form.
- *Regel zu unscharf*: je vier Schärfungen von §3.14 und §3.15 (je fünf Fassungen) zeigen, dass die Regel nachgeschärft wurde, ohne die Rate zu senken.
- *Regel nicht am Arbeitsort sichtbar*: die Regeln stehen in `AGENTS.md`; `implement-slice` Schritt 3 liest die Datei, der Reviewer-Skill verweist auf sie, `architect.md` und `planner.md` nennen keine der Regeln §3.12 bis §3.16, obwohl der Architect die Tabelle *Zusage · Fänger* einfügt und die Randformen entscheidet. Ob die Datei dort dennoch im Kontext liegt, gilt nur für meinen eigenen Lauf: `AGENTS.md` war als Projektanweisung geladen.

### 4.3 Wer liest die Regeln (außerhalb von `AGENTS.md`)

Treffer nach `grep` in `.claude/`, `.harness/skills/`, `harness/`, `tools/`, `src/`:

- §3.10: `.harness/skills/reviewer.md` Zeile 75.
- §3.15: `.harness/skills/reviewer.md` Zeilen 67 und 78.
- §3.16: `.harness/skills/reviewer.md` Zeile 65.
- §3.17: `.claude/commands/implement-slice.md` Zeile 163, `.claude/agents/verifier.md` Zeile 27.
- §3.18: `src/hexagon/services/baselineimport/anfrage_test.go` Zeile 231, `manifest_test.go` Zeile 229.
- §3.9, §3.11, §3.12, §3.13, §3.14: **kein** Treffer.

Inhaltlich (ohne Paragraphen-Nummer) stehen in den Rollen-Dateien: Mutation je Zusage (`implementer.md`, `implement-slice` Schritt 19), Generator-Pflicht (Schritt 14, eigenes Register-Ergebnis), Kommentar-Prüfung (Schritt 20). Der Implementer-Typ nennt nicht §3.9, §3.12 bis §3.16.

## 5. Sensor-Bestand und Lücken

Vorhanden und einschlägig: `make test-race` (Datenrennen, §3.10), `make docs-check` (Straten-Richtung, Kennungen, Struktur-Zellen der Sensors-Tabelle in `harness/README.md`), `make lint` (§3.2), `.claude/hooks/pretooluse-command-guard.sh` (§3.1), `tools/harness/commit-msg-traceability.sh` (Kennung im Commit).

**Kein Gate misst die Größe von `AGENTS.md`, von `implement-slice.md` (16 992 Zeichen, 221 Zeilen) oder des Reviewer-Skills (7 397 Zeichen).** Das Modul `file` von d-check (Zeilen- und Byte-Obergrenze einer ganzen Datei, `file-lines-exceeded`, `file-bytes-exceeded`) gibt es seit d-check v0.78.0, der Pin ist v0.86.1; `.d-check.yml` aktiviert es nicht (`modules:` nennt `links, anchors, ids, matrix, codepaths, spans, planning, structure, targets`). Das Handbuch nennt außerdem für `structure` die Bedingung `max-lines` je Abschnitt (`sections: each`), die jeden Abschnitt `### 3.<n>` einzeln deckelt. Beides ist hier **nicht erprobt**; die Wirkung folgt aus dem Handbuchtext des Images.

Vorschlag für Sensoren (Planner-Kennungen, noch nicht angelegt):

| Kennung | Inhalt | Regel | Aufwand | Ort der Wirkung |
|---|---|---|---|---|
| `slice-groessenmarke-werkzeug` (besteht in `open/`) | Schätzung, Marke, Messung, Eintrag der Überschreitung | §3.14 | L | Gate oder Handoff-Prüfung; Ort entscheidet der Architect (Plan §6) |
| `slice-guide-groessen-gate` | `file`-Modul für `AGENTS.md`, Commands und Reviewer-Skill; optional `structure max-lines` je Hard Rule | alle (Guide-Datei-Wildwuchs) | S | `make docs-check`, in `GATE_CHECKS` |
| `slice-plan-form-sensor` | `structure`-Regel auf Slice-Pläne: Tabelle *Zusage · Fänger* und Raster in §6 mit Spalte *Fänger*, keine leere Zelle | §3.15, §3.16 (Form, nicht Inhalt) | M (die Wirkung bei mehreren Tabellen je Abschnitt ist ungeprüft; Archiv-Stubs und Pläne ohne Vertrag brauchen `skip-pattern`/`exempt-paths`) | `make docs-check` |
| `slice-unsichtbare-zeichen` | Prüfung auf BOM, NBSP, U+2028, U+2029, U+200B in `src/**/*.go`, `tools/**` und Fixtures | §3.18 (Teil) | S | `make lint` (eigene Prüfungen in `tools/harness/lint.sh`) |

Nicht vorgeschlagen: ein Mutationswerkzeug (§3.10, L, bräuchte Netz für `deps`, widerspricht der Netzlosigkeit der Prüfläufe), ein Test „Zahl im Plan = Zahl im Test“ (§3.9, nicht berechenbar), ein Wächter für das Lesen fremder Berichte (§3.11, inferential).

## 6. Empfehlung je Regel

| Regel | Ausgang | Begründung |
|---|---|---|
| §3.9 | A | Kurz (674 Zeichen), gilt für jede Rolle, die einen Plan oder Code ändert (Planner, Architect, Implementer; Verifier findet), und kein anderer Ort liest sie. Wirkung nicht belegt (6 von 6 Slices seit der Schärfung), aber kein Träger sonst: Entfernen ließe 30 Befundzeilen ohne Anker. Ein Teil-Sensor (Plan im selben Commit) bleibt Option, ohne eigenen Slice. |
| §3.10 | B | Größte Regel (2512 Zeichen) mit drei Teilen: Kern (Mutation je Zusage, Bericht), Nebenläufigkeit (Detektor, Raten, Fenster), Plan-Absatz. Kern bleibt als ein Satz; der Rest gehört dorthin, wo der Implementer (Schritt 19 in `implement-slice`) und Reviewer (MEDIUM-Zeile im Skill) schon arbeiten. Der Detektor besteht als Gate. |
| §3.11 | A | 364 Zeichen; Schärfung von §3.7; kein anderer Ort außer dem Skill; 101 Zitate zeigen den Gebrauch als Reviewer-Anker. Zusammenlegen mit §3.7 würde Verweise brechen, ohne nennenswert zu sparen. |
| §3.12 | B | Von §3.16 geschärft (Raster *Operation × Argumentart*). Überschrift und ein Satz bleiben; der Volltext gehört mit §3.16 zusammen in die Datei für die Rollen, die Pläne schreiben (Architect, Planner) und prüfen. Die Nummer bleibt, 24 von 25 Plänen zitieren sie. |
| §3.13 | B | Planner-Regel (Sendung an andere Slices): Rollen Planner und Architect. Seltenes Vorkommen (3 Vorgänge, keine Befundzeile nach 10-23), in 17 von 25 Plan-Dateien zitiert, also am Arbeitsort des Planners sichtbar. Volltext nach `plan-welle.md` und `architect.md` (oder `harness/rules/`). |
| §3.14 | C | Die Prosa hat bei 16 Überschreitungen die Rückführung nicht erzwungen (Rate 65 % vor, 62 % nach der Schärfung). Eine Handlung, die Zahlen rechnet, gehört in ein Werkzeug (`slice-groessenmarke-werkzeug`); die Faktoren gehören nach [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) und Modul `MR-001` in `harness/vertraege/`. Bis dahin Zwischenschritt B: Volltext in `plan-welle.md` (der Planner setzt die Marke) und `implement-slice` (der Implementer misst); in `AGENTS.md` bleibt der Zeiger. Der Slice steht in `open/`, nicht in `next/`. |
| §3.15 | B | Kern des Review-Maßstabs, Rate unverändert (21 von 21). Die Regel trägt Architect (fügt die Tabelle ein), Implementer, Reviewer, Verifier; der Reviewer-Skill hält drei MEDIUM-Zeilen. Volltext in `harness/rules/` mit §3.10, §3.12, §3.16, §3.17 als ein Block; `AGENTS.md` behält Überschrift, Satz, Zeiger. Sensor für die Form (S bis M) als eigener Slice. |
| §3.16 | B | Wie §3.15; Raster-Tabelle in §6 bei 15 von 16 Plänen, Inhalt: 13 von 16 Slices mit Vorkommen. `structure`-Sensor deckt leere Zellen, nicht den Ort. |
| §3.17 | B | Inhalt steht bereits in `implement-slice` Schritt 19 und `verifier.md` (Duplikat); `AGENTS.md` hält die längste Fassung (1544 Zeichen) mit einem OData-Beispiel. Ein Satz plus Zeiger genügt. Beachte: jede Verkürzung muss die Berichtspflicht („Menge, Zahl der Elemente, eine Mutation an nicht genanntem Element, Prüfgrenze bei unbeschränkter Menge“) im Ziel tragen, sonst geht sie verloren. |
| §3.18 | A | Tool-Falle, die jede schreibende Rolle trifft (alle sechs Typen schreiben Dateien); 953 Zeichen; bisher der einzige Fall, in dem ein Bericht „galt und hielt“ sagt. Der Teil-Sensor (`slice-unsichtbare-zeichen`, S) ergänzt, ersetzt nicht. |

### 6.1 Antwort auf „warum §3.14 in `AGENTS.md`, gibt es keinen besseren Platz, gibt es kein Gate“

- **Besserer Platz:** ja. Die Regel richtet sich an den Planner (setzt die Marke vor dem ersten Code-Commit) und den Implementer (misst am Ende). Beide Rollen haben einen Anweisungssatz (`plan-welle`, `implement-slice`); die Faktoren sind Werkzeug-Vertrag und gehören nach [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) in `harness/vertraege/`, nicht in die Hard Rules. Sie steht in `AGENTS.md`, weil dort bei der Einführung am 2026-10-18 alle Slice-Regeln standen.
- **Gate:** keines. `slice-groessenmarke-werkzeug` liegt in `open/` (Plan vom 2026-10-25, ohne Zeilen-Schätzung); Feedforward hat bei 11 von 17 und 5 von 8 Slices die Überschreitung nicht verhindert, und eine Handlung („Rückführung“) ist keine Prüfung.
- **Empfehlung:** C, mit Zwischenschritt B, wenn der Slice nicht sofort läuft.

### 6.2 Doppelte Verkörperung (Duplikate)

- §3.17 steht in `AGENTS.md`, `implement-slice` Schritt 19 und `verifier.md`. In `AGENTS.md` ist es die längste Fassung.
- §3.10 (Rate) und §3.15 (Akzeptiertes Negativ, Fänger nur eine Form) stehen zusätzlich als MEDIUM-Zeilen im Reviewer-Skill.
- §3.14: die Faktoren (85/40/65/200/1,25) stehen heute nur in der Prosa; das Werkzeug wird sie als Vertrag führen und die Prosa damit doppeln.
- Hook und Regel: §3.1 steht im Kopf von `.claude/hooks/pretooluse-command-guard.sh` (Zeiger, kein Duplikat).

### 6.3 Beifang (nicht geändert)

`implement-slice.md` Schritt 19 nennt `AGENTS.md` §3.6 als Quelle für „zu jedem Wächter die Mutation benennen“. §3.6 ist „Gates dürfen nicht ohne ADR gelockert werden“; die Mutationsregel steht in §3.10. Der Zeiger passt vermutlich nicht. Das ist eine Änderung an einem Command und gehört nicht in diesen Auftrag; sie steht hier als Fund.

## 7. Gesamtbild

- **Größe:** `AGENTS.md` hat 24 193 Zeichen (24 648 Bytes, 337 Zeilen), davon §3.9 bis §3.18 mit 13 589 Zeichen, also 56 %. Kein Gate misst die Größe.
- **Nach den Empfehlungen** (B und C kürzen sieben Regeln — §3.10, §3.12, §3.13, §3.14, §3.15, §3.16, §3.17 — auf je Überschrift, ein bis drei Sätze und einen Zeiger, geschätzt 350 bis 400 Zeichen): etwa 15 500 Bytes und etwa 280 Zeilen (Schätzung des Planners aus 11 834 Bytes alt und 7 × 380 neu; nicht gemessen). Die Ersparnis liegt bei rund 9 100 Zeichen, 38 % der Datei.
- **Ersparnis je Lauf:** gilt nur, wenn jeder Agentenlauf die Datei lädt (für meinen Lauf bestätigt; für die anderen nicht geprüft). Der Bestand zählt 224 Agent-Läufe (`make span-report`, 2026-10-18 bis 2026-10-25); eine Ersparnis in Tokens lässt sich nicht angeben (kein Verbrauchs-Zähler im Bestand).
- **Zusatzlast je Rolle:** Die Ersparnis ist eine Obergrenze. Wer die Regeln anwendet, liest sie weiter: Implementer, Reviewer und Verifier lesen den Block *Vertrag und Fänger* (rund 7 800 Zeichen für §3.10, §3.12, §3.15, §3.16, §3.17) zusätzlich über den Zeiger und sparen daher kaum; Architect liest den Block für §3.12, §3.15, §3.16 (rund 3 700); Planner liest §3.13 und §3.14 (rund 3 800) in `plan-welle`. Gewinner sind Läufe ohne Vertragsbezug: Validator, Planner bei Closure und Beobachtungs-Läufe wie dieser.
- **Gate für die Größe:** vorhanden im Werkzeug, nicht aktiviert (siehe 5). Vorschlag nach der Verschlankung: `file` mit `max-bytes: 17000` und `max-lines: 300` für `AGENTS.md` (Ist plus etwa 10 %, Ratchet nach dem Muster der Baseline-Vorlage `templates/.d-check.yml`; Wert ist eine Schätzung auf Basis der Zahlen oben und vor dem Setzen zu messen). Ohne die Verschlankung hieße die Grenze 25 000 Bytes und 340 Zeilen und würde die Datei nur einfrieren. Dieselbe Obergrenze für `implement-slice.md` (Ist 16 992 Zeichen) wäre die zweite Zeile; ihr Wert muss der User setzen.

## 8. Fragen an den User

1. Soll die Verschlankung der Hard Rules (B für sieben Regeln, C für §3.14) als eigener Slice laufen, und soll der Volltext nach `harness/rules/<name>.md` (Form des Regelwerks) oder in die Rollen-Dateien? Hinweis: Das berührt `AGENTS.md`, Commands und Typen; sie ändert nur der User.
2. Soll `slice-groessenmarke-werkzeug` von `open/` nach `next/` vor die Verschlankung rücken, oder gilt der Zwischenschritt B für §3.14 (Volltext in `plan-welle`/`implement-slice`)?
3. Soll das `file`-Modul von d-check als Gate für `AGENTS.md` (und `implement-slice.md`, Reviewer-Skill) in `.d-check.yml` aktiviert werden, und mit welchen Grenzen (Vorschlag 17 000 Bytes und 300 Zeilen für `AGENTS.md` nach der Verschlankung)?
4. Sollen die Form-Sensoren `slice-plan-form-sensor` (M) und `slice-unsichtbare-zeichen` (S) als Slices angelegt werden, oder reicht der Reviewer-Anker?
5. Soll §3.9 trotz „nicht verhindert“ (6 von 6 Slices seit der Schärfung) in `AGENTS.md` bleiben (Empfehlung A), oder wünscht der User einen Teil-Sensor (Plan im selben Commit wie die Korrektur, Fehlalarm bei Fix-Runden)?
6. Die Zeiger in `implement-slice` Schritt 19 (`§3.6`) passt vermutlich nicht zur Regel; soll sie mit der Verschlankung berichtigt werden?
7. Soll §3.12 künftig als Satz plus Zeiger auf §3.16 stehen (Nummern bleiben, 24 von 25 Plänen zitieren sie), damit die Randform-Regel nur einmal vollständig steht?
8. Maßstab: Zählt für „hat nicht funktioniert“ die Vorkommens-Rate je Slice (hier §3.14, §3.15, §3.16 unverändert), oder die Schwere? Letzteres lässt sich aus dem Bestand nicht messen; das Register müsste die Schwere je Vorgang führen.

## 9. Gegenlesen

Gegengelesen wurde diese Datei auf Rohbytes mit `LC_ALL=C grep -P`: keine Folge aus Backslash, `u` und vier Hex-Ziffern (0 Treffer), kein BOM, kein NBSP, kein U+2028, U+2029 oder U+200B (0 Treffer). `make gates` lief vor dem Commit grün (Exit 0).
