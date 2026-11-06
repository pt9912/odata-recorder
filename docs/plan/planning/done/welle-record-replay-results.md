# Welle welle-record-replay — Aufzeichnen und Wiederabspielen — Closure-Notiz

> **Zitier-Form** *(bleibt stehen — Norm, kein Ausfüll-Hinweis).* Dieses
> Artefakt friert ein; was es zitiert, bewegt sich weiter. Deshalb: **Kennung,
> nicht Adresse** — `slice-<Kennung>` statt seines Lifecycle-Pfads, `make <target>`
> statt eines Links auf die Sensor-Datei, eine Baseline-Stelle als
> `v<X.Y.Z>` · `regelwerk/<datei>.md` §<Abschnitt> statt als Link
> (Baseline-Regelwerk `grundlagen-harness-dateien.md` §harness/README.md als
> Einstiegspunkt). Adopierter Tag: `v6.18.1`.

**Welle:** welle-record-replay
**Abschluss:** 2026-11-06
**Verantwortlich:** pt9912

## Was wurde geliefert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — *was gelernt wurde*: geliefert · was
funktionierte · was anders lief. Mit ID-Bezug, wo es einen gibt.

27 Slices der Welle und der wellenlose `slice-gegenproben-build-test` (die Zahl aus dem Lauf von `archive-welle --vorschau`: 27 Mitglieder, 1 wellenlos).

- **Der Durchstich M2** (`LH-FA-01`, `LH-FA-02`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`): `TestDurchstichM2` zeichnet über den Recorder gegen einen Referenzdienst auf, beendet die Session, importiert sie in eine Baseline, legt einen Workspace an und spielt nach dem Ende des Referenzdienstes byte-gleich ab (`slice-m2-referenzdienst`, `slice-text-antworten`); `TestProtokolltreue` vergleicht Direktzugriff und Zugriff über den Recorder über 16 Antwortarten (`LH-QA-07`, `slice-protokolltreue`).
- **Record-Pfad:** Redaction von Headern und Query mit Match-Key, Streaming-Redaction fail-closed, Capture-Pipeline, `RecordUseCase`, Upstream-Adapter mit TLS (`LH-FA-02`, `LH-FA-05`, `LH-FA-14`, `LH-RB-02`; `slice-redaction-match-key`, `slice-streaming-redaction`, `slice-capture-pipeline`, `slice-record-service`, `slice-upstream-adapter`, `slice-upstream-tls`).
- **Sessions und Wartung:** `MaintenanceUseCase`, `SessionUseCase` in Verwaltung und Beenden (`LH-FA-02`, `LH-FA-03`, `LH-FA-13`; `slice-session-wartung`, `slice-session-use-case`, `slice-session-beenden`).
- **Codec, Import, Workspace, Replay:** OData-Codec zum Lesen (CSDL und JSON), `ImportUseCase` mit Coverage-Beweis, `WorkspaceUseCase`, `ReplayUseCase` mit Zuordnung, Strategien und Cursor (`LH-FA-04`, `LH-FA-09`, `LH-FA-11`, `LH-FA-18`; `slice-odata-codec-csdl`, `slice-odata-codec-json`, `slice-baseline-import`, `slice-workspace-use-case`, `slice-replay-service`).
- **Treiber und Wurzel:** HTTP-Adapter in Kern und Lebenszyklus, CLI-Adapter, Composition Root (`LH-FA-12`; `slice-http-adapter`, `slice-http-lebenszyklus`, `slice-cli-adapter`, `slice-bootstrap-m2`).
- **Nachweise am Binary:** `make test-integration` (sieben Stufen, das gebaute Binary im Container ohne Netz, `SPEC-037`) und `make test-determinism` (100 Läufe nach Reset, `SPEC-038`), beide neben der Gate-Kette (`slice-test-integration`, `slice-test-determinism`).
- **Werkzeuge im Gate:** `make test-race` ([ADR-0007](../../adr/0007-race-detektor-im-gate.md), `slice-race-gate`) und `make groessenmarke` (`SPEC-128`, `slice-groessenmarke-werkzeug`; die Lage in der Kette ist [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md), Status `Proposed`).
- Verträge in `spec/spezifikation.md` und `harness/vertraege/` (darunter `SPEC-037`, `SPEC-038`, `SPEC-122`, `SPEC-128`, `SPEC-136`, `SPEC-138`).

## Was hat funktioniert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3.

- Der Schnitt vor dem Code: `slice-record-service`, `slice-upstream-adapter`, `slice-session-wartung`, `slice-session-use-case`, `slice-http-adapter` und `slice-bootstrap-m2` wurden vor dem ersten Code-Commit oder an einem vorab benannten Teil geschnitten (Drift-Log der Roadmap); keine Rückführung nach dem Code wurde nötig.
- Implementer, Reviewer und Verifier in frischem Kontext, die Verifier mit eigenen Probe-Paketen aus Generatoren über die Klasse: In keinem Slice war der Produktcode der Befund, die Kosten waren Fänger für Mengen aus dem Wortlaut der Zusagen.
- Der Durchstich vor der Härtung (Entscheidung des Users vom 2026-10-31): M2 stand mit `slice-m2-referenzdienst`, die Absicherung (`slice-text-antworten`, `slice-protokolltreue`, `slice-test-integration`, `slice-test-determinism`) folgte danach.
- Der Parallelbetrieb in Worktrees mit der Sperre für schwere Läufe und den Übergangsregeln: vier Linien liefen nebeneinander, die Gate-Kette hatte nie zwei Besitzer.
- Das Beobachtungs-Register: 278 Belege in 43 Beobachtungen am Tag der Closure (gezählt mit `find … evidence`); die Verkörperung ab 3× hat 27 Einträge auf eine Regel, einen Sensor oder eine benannte Adresse geführt.

## Was ging anders als geplant?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — jede Zeile möglichst mit der Konsequenz,
die daraus schon gezogen wurde (Folge-Slice, Spec-Version).

- Die Welle wuchs von 14 auf 27 Slices (Zeilen im Drift-Log der Roadmap): Schnitte vor dem Code, ein Race-Gate, das Werkzeug der Größenmarke, der Vergleichstest, die Nachweise am Binary. Konsequenz: die Wellen-Datei trägt die Reihenfolge; keine Rückführung.
- **Die Formel der Größenmarke widersprach sich in zwei Quellen** (`AGENTS.md` §3.14 mit 40/85 und 1,25, `SPEC-128` Nr. 1 mit 80/85 und 1,35) und überschätzte danach die meisten Slices. Konsequenz: `AGENTS.md` §3.14 trägt keine Zahl mehr und zeigt auf `SPEC-128` Nr. 1; die Neukalibrierung ist ein Auftrag an den Architect (unten, *Messpunkte der Größenmarke*).
- Der Pflichtlauf von `make test-integration` und `make test-determinism` vor der Closure hatte keinen Träger; er stand als Rat im Plan. Konsequenz: Entscheidung des Users, beide kommen in `GATE_CHECKS` (`slice-gate-integration-determinism`); bis dahin ist der Lauf ein Schritt in `close-welle`.
- Rückgaben an den Architect und den Planner standen mehrfach nur im Bericht der Rolle; die nächste Prüfrunde fand sie. Konsequenz: `AGENTS.md` §3.13 (offene Entscheidung im selben Commit in §6 mit Ausgang *offen*), `implement-slice` Schritt 21 (jeder Handoff trägt einen Commit mit Plan-Nachzug).
- Ein Report einer Rolle machte `make gates` rot (Kennung ohne Link), weil niemand nach dem Report `make docs-check` fuhr; die Review-Reports lassen sich bei der Archivierung nur löschen, wenn kein Verweis auf sie bleibt (Sperre `[haenger]`). Konsequenz: 63 Verweise in 56 Dateien tragen vor der Archivierung die Kennung des Reports ohne Pfad.
- Auslastungen der Marke zwischen 0,39 und 1,06 (Ende gegen Marke): drei von 17 Messpunkten rissen die Marke, acht lagen unter 0,60 (Tabelle unten).

## Messpunkte der Größenmarke

Regeln dieser Sektion: `SPEC-128` Nr. 9 — Auslastung gleich Messung am Ende geteilt durch die Marke; über 1,00 oder unter 0,60 meldet der Planner dem Architect. Die Tabelle ist aus den Plänen unter `done/` von Hand nachgerechnet (*S* die Schätzung nach `SPEC-128` Nr. 1 aus der Tabelle des Plans, *E* die Messung am Ende nach `git diff --numstat`, *p* der erwartete Produktcode); Slices anderer Wellen sind mitgezählt, soweit sie einen Messpunkt tragen. Bei `slice-http-lebenszyklus` ist *E* aus der Auslastung 0,64 des Plans berechnet (0,64 mal 2700), die Messpunkt-Zeile im Plan trägt den Handoff-Wert 1161.

| Slice | *p* | *S* | Marke | *E* | *E* ÷ *S* | *E* ÷ Marke |
|---|---|---|---|---|---|---|
| `slice-v2-import-codecs` | 120 | 1045 | 1500 | 758 | 0,73 | 0,51 |
| `slice-http-lebenszyklus` | 140 | 1990 | 2700 | 1728 | 0,87 | 0,64 |
| `slice-test-integration` | 330 | 2960 | 4000 | 1931 | 0,65 | 0,48 |
| `slice-protokolltreue` | 0 | 1965 | 2700 | 1517 | 0,77 | 0,56 |
| `slice-test-determinism` | 80 | 1280 | 1800 | 694 | 0,54 | 0,39 |
| `slice-m2-referenzdienst` | 40 | 1190 | 1700 | 980 | 0,82 | 0,58 |
| `slice-v2-query-filter` | 400 | 3370 | 4600 | 2546 | 0,76 | 0,55 |
| `slice-bootstrap-m2` | 550 | 2445 | 2900 | 3075 | 1,26 | 1,06 |
| `slice-telemetry-port` | 580 | 2150 | 3000 | 3093 | 1,44 | 1,03 |
| `slice-groessenmarke-werkzeug` | 750 | 4670 | 6400 | 5342 | 1,14 | 0,83 |
| `slice-replay-service` | 800 | 6925 | 9200 | 6017 | 0,87 | 0,65 |
| `slice-workspace-use-case` | 950 | 7815 | 10600 | 6540 | 0,84 | 0,62 |
| `slice-handbuch-geruest` | 1000 | 9815 | 6900 | 6956 | 0,71 | 1,01 |
| `slice-v2-query` | 1000 | 5130 | 7000 | 3781 | 0,74 | 0,54 |
| `slice-v2-payload-lesen` | 1400 | 5970 | 8100 | 4085 | 0,68 | 0,50 |
| `slice-v2-metadaten` | 1500 | 6070 | 8200 | 4962 | 0,82 | 0,61 |
| `slice-odata-url-pfad` | 1700 | 6790 | 9200 | 6517 | 0,96 | 0,71 |

Lesart: 17 Messpunkte. Median *E* ÷ *S* gesamt 0,82; mit *p* unter 350 (sechs Punkte) 0,75, mit *p* ab 400 (elf Punkte) 0,84. Auslastung unter 0,60 in acht, über 1,00 in drei Punkten (`slice-bootstrap-m2`, `slice-telemetry-port`, `slice-handbuch-geruest`). Die Vorbereitung dieser Closure nannte 0,77 und zehn von 17; der Punktesatz ist ein anderer, die Neurechnung aus allen Messpunkten ist Sache des Architect. Der Zuwachs vom Handoff zum Ende lag in den Plänen bei 1,19 bis 1,49.

**Auftrag an den Architect** (Entscheidung des Users vom 2026-11-06, Träger ist noch kein Slice): die Faktoren von `SPEC-128` Nr. 1 und 2 nach Klassen getrennt neu rechnen (mit und ohne Produktcode, nebenläufig und andere), `SPEC-128` Nr. 9 auf den Endwert statt auf den Handoff-Wert stellen, den Vertrag und die Gegenproben nachziehen. Register: `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` und `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen` stehen auf `geplant`.

## Steering-Loop-Einträge

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 (hier stehen **nur** Beobachtungen, die im
Register 3× erreicht haben; jeder Eintrag nennt seine `BEO-<KUERZEL>/<slug>`) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

27 Beobachtungen haben 3× erreicht; zwölf waren schon verkörpert und tragen ihren Anker, fünfzehn bekamen mit dieser Closure einen Ausgang.

- **Hard Rule** geschärft: Zusage und Fänger als Paar, gleiche verglichene Werte, Aufrufstellen der Ports als Menge, Allaussage als Schleife.
  — liegt in `AGENTS.md §3.15`.
  Auslöser: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (41×).
- **Reviewer-Skill** ergänzt: MEDIUM-Zeilen für Paar von Ursachen ohne Fänger, Fänger mit einer Form der Eingabe, akzeptiertes Negativ ohne Grund und Folge, Fänger schmaler als Zusage.
  — liegt in `.harness/skills/reviewer.md`.
  Auslöser: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (41×) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (13×).
- **Hard Rule** geschärft: Randformen als Raster Operation × Argumentart, Allaussage des Vertrags mit Fänger über alle Fälle.
  — liegt in `AGENTS.md §3.16`.
  Auslöser: `BEO-REPO/randform-erst-im-review-entschieden` (30×) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×).
- **Hard Rule** ergänzt: ein Fix deckt die Klasse der Zusage, nicht die genannte Mutation; der Bericht führt je Befund die Menge, ihre Achsen (Zeichen, Länge, Folge, Seite, Stelle, Zustand), die Elementzahl je Achse und eine nicht genannte Mutation je Achse.
  — liegt in `AGENTS.md §3.17`.
  Auslöser: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (21×).
- **Command** geschärft: die Achsentabelle je Befund, Felder derselben Referenz tragen in der Fixture verschiedene Werte (Schritt 19).
  — liegt in `.claude/commands/implement-slice.md`.
  Auslöser: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (21×), `BEO-REPO/zustand-des-arguments-als-dimension` (slice-handbuch-geruest, slice-odata-url-pfad, slice-replay-service, slice-telemetry-port — 4×), `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (slice-bootstrap-m2, slice-odata-url-pfad, slice-workspace-use-case — 3×).
- **Command** geschärft: eine Dezimalzahl ist die Menge Ziffer × Stelle; eine Testhilfe trägt alles, was ihr Original trägt; eine unbeschränkte Menge und ein Fänger schmaler als die Zusage stehen im Plan als akzeptiertes Negativ mit Grund und Folge oder werden verbreitert; Klassen generierbar im Handoff (Schritt 14).
  — liegt in `.claude/commands/implement-slice.md`.
  Auslöser: `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×), `BEO-REPO/ziffern-klasse-ohne-schleife` (slice-handbuch-geruest, slice-replay-service, slice-text-antworten, slice-v2-query-filter, slice-workspace-use-case — 5×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (slice-baseline-import, slice-groessenmarke-werkzeug, slice-odata-codec-json, slice-workspace-use-case — 4×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (slice-cli-adapter, slice-groessenmarke-werkzeug, slice-http-adapter, slice-http-lebenszyklus, slice-odata-url-filter, slice-odata-url-pfad, slice-telemetry-port, slice-v2-query-filter — 8×).
- **Command** geschärft: jeder Handoff trägt einen Commit, der den Bericht in §3 des Plans nachzieht (Schritt 21).
  — liegt in `.claude/commands/implement-slice.md`.
  Auslöser: `BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo` (6×, davor verkörpert in `AGENTS.md` §3.10).
- **Reviewer-Skill** ergänzt: MEDIUM-Zeilen für Fänger schmaler als Zusage, Menge unbeschränkt oder ohne Achsenliste, Fixture, Hilfe oder Kommentar, die verdecken, was der Test prüft; Hinweis zu `panic:` und Zeitlimit statt `--- FAIL`.
  — liegt in `.harness/skills/reviewer.md`.
  Auslöser: `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (slice-session-beenden, slice-session-use-case, slice-v2-metadaten, slice-workspace-use-case — 4×), `BEO-REPO/menge-im-plan-weicht-vom-orakel-des-tests-ab` (slice-groessenmarke-werkzeug, slice-handbuch-geruest, slice-v2-query — 3×), `BEO-REPO/kommentar-nennt-mehr-elemente-als-der-test-prueft` (slice-replay-service, slice-telemetry-port, slice-v2-import-codecs — 3×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/fixture-gleicht-dem-wert-der-mutation`.
- **Hard Rule** geschärft: die Rate eines nebenläufigen Fängers steht mit Last des Hosts und Messendem, eine Einzelzahl gilt als Größenordnung; der Detektor von `make test-race` zählt je Prozess.
  — liegt in `AGENTS.md §3.10`.
  Auslöser: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (slice-http-lebenszyklus, slice-race-gate, slice-session-wartung, slice-telemetry-port, slice-text-antworten, slice-v2-import-codecs — 6×).
- **Sensor** neu: der Race-Detektor als Glied der Gate-Kette.
  — liegt in `harness/mk/go-gates.mk` (Ziel `make test-race`, Zeile mit dem Anker).
  Auslöser: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×).
- **Sensor** neu: die Größenmarke der Pläne mit Block als Gate.
  — liegt in `harness/mk/groessenmarke.mk` (Ziel `make groessenmarke`, Zeile mit dem Anker).
  Auslöser: `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (19×).
- **Hard Rule** gekürzt: `AGENTS.md` §3.14 behält die Metrik, die Marke vom Planner vor dem Code und die Schnittregel und trägt keine Zahl mehr; Formel, Faktoren, Zuschlag und Schnittgrenze stehen allein in `SPEC-128` Nr. 1. Die Neukalibrierung ist der Auftrag an den Architect oben.
  — liegt in `AGENTS.md §3.14`.
  Auslöser: `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (slice-protokolltreue, slice-test-determinism, slice-test-integration, slice-v2-import-codecs, slice-v2-metadaten, slice-v2-payload-lesen, slice-v2-query, slice-v2-query-filter — 8×) und `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen` (slice-bootstrap-m2, slice-groessenmarke-werkzeug, slice-http-lebenszyklus — 3×); beide Ausgänge sind `geplant` (Auftrag an den Architect, noch ohne Slice).
- **Hard Rule** geschärft: Berichtigungen erreichen jede Stelle, die die Aussage trägt.
  — liegt in `AGENTS.md §3.9`.
  Auslöser: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (20×).
- **Hard Rule** geschärft: eine offene Entscheidung steht im selben Commit wie ihr Bericht in §6 mit Ausgang *offen* und Adresse.
  — liegt in `AGENTS.md §3.13`.
  Auslöser: `BEO-REPO/befund-an-rolle-ohne-artefakt` (slice-record-service, slice-upstream-adapter, slice-odata-codec-json, slice-http-lebenszyklus — 4×).
- **Hard Rule** ergänzt: ein Wert mit Escape-Folge oder Rohbytes wird aus Bytewerten gebaut und nach dem Schreiben gegengelesen.
  — liegt in `AGENTS.md §3.18`.
  Auslöser: `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (slice-redaction-match-key, slice-streaming-redaction, slice-odata-codec-json, slice-unsichtbare-zeichen, slice-v2-metadaten — 5×).
- **Übergangsregel** mit Anker: der Stempel der Gegenproben gehört dem Integrationsschritt und steht nach dem letzten Eingabe-Commit.
  — liegt in `docs/plan/planning/README.md`.
  Auslöser: `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (slice-baseline-import, slice-guide-groessen-gate, slice-odata-url-filter, slice-unsichtbare-zeichen — 4×); ein Sensor fehlt (benannte Grenze).
- **Command** ergänzt: der Pflichtlauf von `make test-integration` und `make test-determinism` als Schritt vor der Closure, bis die beiden Ziele an `GATE_CHECKS` hängen.
  — liegt in `.claude/commands/close-welle.md`.
  Auslöser: `BEO-REPO/pflichtlauf-neben-der-gate-kette-ohne-traeger` (slice-test-integration, slice-test-determinism — 2×, unter der Schwelle; Ausgang `geplant` in `slice-gate-integration-determinism` auf Entscheidung des Users).
- **Ausgang `geplant`, noch kein Zielort:**
  `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (slice-baseline-import, slice-unsichtbare-zeichen, slice-v2-metadaten, slice-workspace-use-case — 4×) geplant in `slice-guard-schreibwege` (Wächter im Command-Guard für `sed -i` und Heredoc auf Repo-Dateien; `git commit -F -` bleibt erlaubt);
  `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (slice-groessenmarke-werkzeug, slice-guide-groessen-gate, slice-odata-codec-json, slice-unsichtbare-zeichen — 4×) geplant in `slice-coverage-gate` (Auftrag 5 an den Architect).
- **Sensor benannt, Prosa ausgeschöpft:** `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (21×) hat in `AGENTS.md` §3.17, `implement-slice` und `verifier.md` alle Prosa, die sie tragen kann; der Träger des Sensors ist `slice-mutationslauf-werkzeug` (Mutationsläufe der Rollen in einem Container-Lauf), Stand in `in-progress/`.
- **Offen, Schwelle erreicht, ohne Zielort:** `BEO-REPO/sqlite-transaktionsvertrag-rot-unter-last` (slice-http-adapter, slice-odata-url-pfad, slice-v2-metadaten, slice-workspace-use-case — 4×): die Ursache ist ungeklärt (Last oder Zustand des Adapters), kein Lauf dieser Closure zeigte das Rot; Adresse Architect, Auslöser ist ein weiteres Rot von `TestTransaktionVertrag`.

- **Benannte Spec-Lücken** (verkörpert in einer versionierten Spec oder mit Adresse; sie tragen kein `liegt in`):
  - `SPEC-136` Nr. 12, Randformen (a) und (b) (Listener, dessen `Close` nach einem äußeren Schließen `nil` liefert; Fehler des Accept nach `Anhalten`): im Spec-Stand der Closure im Wortlaut der Nr. 12 entschieden, der Rückgabe-Punkt aus `slice-http-lebenszyklus` ist erledigt.
  - `SPEC-138`: die Kanalgröße „genau 2“ steht in Nr. 1, das Negativ (kein Lauf trennt Puffer 1 von 2) steht nicht in *Ungeprüft*; die übrigen Punkte aus `slice-cli-adapter` (genau einmal, Signale, Pflichten des Starts) tragen Wortlaut in Nr. 6 bis 8. Offen: die Kanalgröße als Eintrag in *Ungeprüft*, Adresse Architect.
  - `SPEC-037` Nr. 10 h: die Markierung `capture_status` von `LH-FA-14` (Negative) hat am Binary keinen Fänger, als akzeptiertes Negativ mit Grund, Folge und Adresse `slice-bootstrap-m6` geführt.
  - `SPEC-162` Nr. 7 (Rangfolge von Elementzahl und Prüfung der Elemente in `$orderby`, aus `slice-v2-query`): gehört zu `welle-odata-v2`, Adresse Architect, Zeile *Offen aus `slice-v2-query`* in der Welle-Datei.
  - Die Faktoren von `SPEC-128` Nr. 1 und 2 (Zuschlag und Klassen): Auftrag an den Architect, siehe oben.

Unter der Schwelle bleiben 16 Beobachtungen (`zweig-ohne-lauf-unter-dem-detektor` und `report-der-rolle-faellt-docs-check` mit 2×, `pflichtlauf-neben-der-gate-kette-ohne-traeger` mit 2×, dreizehn mit 1×).

**Akzeptierte Negative, nicht als Folge-Slice benannt** (Entscheidung des Users vom 2026-11-06): die Fänger V1 bis V4 von `slice-protokolltreue` (geordnete Paare der Kettenklasse, Status 5xx und 1xx, Zeilenanzahl über drei, Schreibweise für Nicht-ASCII) und die V-Befunde von `slice-v2-query`, `slice-v2-query-filter`, `slice-v2-payload-lesen` und `slice-text-antworten` sind akzeptierte Negative mit Grund und Folge: der Code ist laut Probe-Paketen in allen Fällen richtig, Durchstich vor Härtung. Sie werden bei der Closure von `welle-odata-v2` neu bewertet.

## Beobachtungs-Register (Zeiger)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register — der Zähler wird **nicht** hier gepflegt; diese
Sektion ist ein Zeiger und trägt keine Daten.

Der Zähler steht in [`../observations/`](../observations/).
Was in dieser Welle **3×** erreicht hat, steht oben unter
*Steering-Loop-Einträge*.

## Folge-Slices

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — **derivativ**.

- `slice-gate-integration-determinism` (neu, `open/`, Linie C): `make test-integration` und `make test-determinism` an `GATE_CHECKS`; Trigger: diese Welle in `done/`.
- `slice-guard-schreibwege` (neu, `open/`, Linie C): Wächter im Command-Guard für die Schreibwege auf Repo-Dateien.
- `slice-mutationslauf-werkzeug` (`in-progress/`, ohne Welle): Sensor-Träger für `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`.
- `slice-coverage-gate` (`open/`, [welle-sensoren](../welle-sensoren.md) Teil 2): trägt `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`.
- `slice-bootstrap-m6` (`open/`, [welle-admin-api](../welle-admin-api.md)): trägt den Fänger der Markierung `capture_status` (`SPEC-037` Nr. 10 h).
- Kein Slice, ein Auftrag: die Neukalibrierung der Faktoren von `SPEC-128` durch den Architect.

## Verifikation

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 1 — keine Behauptung ohne nachprüfbaren
Anker (Hash, Lauf, Zahl).

- **Trigger-Audit, Carveouts:** `docs/plan/carveouts/` führt keine Datei; 0 offen, keine Welle schließt mit einem Carveout und keines der Gates ist rot.
- **Trigger-Audit, bootstrap-aware Gates:** kein Gate trägt eine Reifestufe; die Schwellen sind feste Grenzen (Größengrenzen der Guide-Dateien nach [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md), Größenmarke nach `SPEC-128`). Die Größengrenzen verschoben sich in dieser Closure nach `SPEC-040` Nr. 8 b: Summe der Zeilen gleich (976), Summe der Bytes von 68861 auf 68663 gesenkt; der Trigger von [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (zwei Verschiebungen in Folge ohne Senkung) ist nicht eingetreten. 0 offen.
- **Trigger-Audit, Entscheidungen** (Re-Evaluierungs-Trigger je ADR gegen den Stand der Closure):
  - [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): der Trigger „Lauf bei geänderter Eingabe über 30 s“ **ist eingetreten**: der Schritt der Fälle dauerte 84,8 s, der Export des Images 38,8 s (Messung unten). Die Gegenproben bleiben außerhalb der Kette; die Entscheidung über Folge (Beschleunigung, anderer Zuschnitt, Kette) liegt beim User, `welle-sensoren` trägt dieselbe Messung (56,9 s Schritt, 87 bis 89 s Gesamt) als Eingabe. Kein stilles Rot: der Lauf ist grün.
  - [ADR-0007](../../adr/0007-race-detektor-im-gate.md): `pruefung race` brauchte 231,5 s, weit unter der Standardgrenze von 10 Minuten; kein Rennen gemeldet, die Plattform trägt den Detektor, das Image bringt den Compiler mit. Nicht eingetreten.
  - [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md): `make groessenmarke` 3,0 s ohne Änderung der Eingaben (Grenze 10 s), das Gate war in keinem Lauf der Closure rot ohne Fehler. Trigger nicht eingetreten; der Status bleibt `Proposed`, die Annahme ist Entscheidung des **Users** (Adresse User, offen).
  - [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md): siehe die Verschiebung der Grenzen oben; nicht eingetreten.
  - [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Codec in Go geliefert), [ADR-0002](../../adr/0002-architektur-gate-a-check.md) und [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) (`make a-check` 0 Befunde), [ADR-0003](../../adr/0003-lint-profil-golangci.md) (`make lint` grün), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Baseline unverändert `v6.18.1`), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) (kein Lasttest, `LH-QA-01` ungeprüft; keine Beobachtung zu Geschwindigkeit, Pflege oder Lizenz), [ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md) und [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) (kein V2-Konstrukt mit neuem Element des Modells): Trigger nicht eingetreten.
- **`make gates`:** Exit 0 auf Stand `5a235db`, unter `flock /tmp/odata-recorder-heavy.lock` (Gesamtzeit 2651 s einschließlich Wartezeit auf die Sperre). Aus dem Log: `a-check` 0 Befunde, `baseline-verify` v6.18.1 OK (54 Dateien), `docs-check` 650 Dateien ohne Befund, `migration-check` ok, `handbuch-check` ok, `zeichen-check` ok (1301 Dateien), `make groessenmarke` ohne Befund (Hinweis zu den drei Plänen in `in-progress/` ohne Block). Die Stufen `test` und `race` kamen aus dem Docker-Zwischenspeicher (Quellen gleich dem letzten Bau).
- **`TestDurchstichM2`:** einzeln im Image der Stufe `test` (`go test ./tests/integration -run '^TestDurchstichM2$' -count=1 -v`) PASS in 0,55 s; er ist Teil von `make test` (Paket `tests/integration`).
- **Laufzeiten ohne Wartezeit** (Stand `a5cdcdd`, Last des Hosts 5,5 bis 5,8 bei 20 Kernen, fremde Läufe zugleich; keine Messung der Leistung, `LH-QA-01` bleibt ungeprüft): `pruefung test` 92,6 s, `pruefung race` 231,5 s, `make groessenmarke` 3,0 s, `make test-integration` 5,9 s und `make test-determinism` 2,1 s bei vorhandenen Images. Der Zuwachs der Kette durch beide Ziele liegt damit bei rund 8 s mit warmem Zwischenspeicher; die Zeit eines kalten Baus der Stufe `integration` ist nicht gemessen (Eingabe für den Auftrag 3 von `slice-gate-integration-determinism`).
- **`make gegenproben`:** Exit 0, 173 Fälle, 0 rot; der Stempel war veraltet (Eingaben geändert), der Lauf dauerte bei geänderter Eingabe 84,8 s für die Fälle und 38,8 s für den Export (796 s einschließlich Wartezeit); der neue Stempel steht in einem eigenen Commit (`a5cdcdd`), eine erneute Prüfung auf diesem Stand ließ den Lauf entfallen (1,1 s).
- **`make test-integration`:** Exit 0, sieben Stufen grün (`TestE2EProzess`, `TestE2EWeiterleitung`, `TestE2EAufzeichnung`, `TestE2EPersistenz`, `TestE2EReplay`, `TestE2EMatching`, `TestE2ESchutz`); Pflichtlauf dieser Closure, im ersten Lauf 446 s einschließlich Wartezeit und Bau.
- **`make test-determinism`:** Exit 0, 100 Läufe, 0 abweichend (`TestE2EDeterminismus`).
- **Archivierung:** nicht ausgeführt. `make archive-welle WELLE=welle-record-replay` bricht in der Vorprüfung mit `[unsauber]` ab (eine ungetrackte, leere `.gitignore` im Wurzelverzeichnis); die Sperre `[haenger]` ist aufgelöst. Die Archivierung führt der User nach Entfernen oder Festschreiben der Datei; bis dahin liegen Slice-Pläne und Review-Reports flach unter `done/` und `docs/reviews/`.
