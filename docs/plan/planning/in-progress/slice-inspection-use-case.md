# Slice slice-inspection-use-case: InspectionUseCase, Exchanges, Coverage und Diagnose lesen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-FA-11`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Exchanges, Coverage, Diagnose) · `SPEC-005` (Zeilen `GET …/exchanges`, `GET /__admin/coverage`, `GET /__admin/diagnostics`) · `LH-FA-11.a` · `SPEC-044`, `SPEC-043` (Coverage und Evidenz) · `SPEC-105`, `SPEC-107` (Repositories, aus denen gelesen wird) · `SPEC-109` (Fehlerabbildung) · `SPEC-071`, `SPEC-112` · `ARC-002`, `ARC-003` · `architecture.md` §2.2 (Zeile `InspectionUseCase`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-inspection-use-case)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `InspectionUseCase` (Inbound Port in `src/hexagon/ports/inbound`, Dienst in `src/hexagon/services/inspection`) liest die Exchanges einer Session, die Coverage einer Baseline und die Diagnose (Konfigurationszusammenfassung ohne Geheimnisse, Modus, Zähler je Fehlercode) paginiert und ohne Geheimnisse; er verändert nichts und kennt kein HTTP.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Endpunkte, Pfade und Statuscodes — `slice-admin-sessions`; der Use Case kennt kein HTTP, der Adapter bildet ab.
- Anlegen, Beenden und Löschen von Sessions — `SessionUseCase` (`done/`); `InspectionUseCase` liest nur.
- Die Ports `Telemetry` und `TelemetryReader` und die Zähler selbst — `welle-betrieb` (`slice-telemetry-port`, `done/`); der Dienst liest die Zähler über `TelemetryReader`, er erzeugt sie nicht und erhält nie `Telemetry` oder den Adapter (`SPEC-230` Nr. 9). Das Zählen in den Diensten ist Einbau (`slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-mock`).
- Export und Abruf der Metriken und Spans — `welle-betrieb` (Welle §6).
- Audit des Lesens — `SPEC-005` auditiert die Schreibendpunkte; ein Lesezugriff schreibt keinen Eintrag.
- Den Import-Report lesen — `ImportUseCase.Get` (`done/`); der Endpunkt liegt in `slice-admin-imports`.

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

- [ ] Exchanges: die Exchanges einer Session paginiert (Seitenform nach Vertrag des Architect); eine Session ohne Exchange ergibt eine leere Seite, eine unbekannte Session `ODR-E7003`; die Felder enthalten keine Geheimnisse (Werte von Header, Query und Body nur redigiert oder nicht enthalten, nach Vertrag).
- [ ] Coverage: die Einträge je EntitySet einer Baseline mit Status und Evidenz aus `CoverageRepository`; Parameter und Fehlerfall einer unbekannten Baseline nach Vertrag.
- [ ] Diagnose: Konfigurationszusammenfassung aus einer Allowlist von Schlüsseln (nie Token, Zugangsdaten oder Inhalte von Geheimnisdateien), Modus und Zähler je Fehlercode über den Lese-Port `TelemetryReader` (`Snapshot().Errors`, ein Eintrag je Code, auch mit 0; `SPEC-230` Nr. 9).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/hexagon/ports/inbound/` | update | `InspectionUseCase` |
| `src/hexagon/services/inspection/` | neu | Lesen der Exchanges, der Coverage und der Diagnose; Seitenbildung |
| `src/hexagon/ports/outbound/`, `src/adapters/driven/sqlite/` | update, nur wenn der Architect eine Seitenabfrage für `RecordingRepository` entscheidet | `ListBySession` liefert heute alle Exchanges einer Session (§6) |
| `spec/spezifikation.md` | update (Architect) | Vertrag samt Seitenform, Feldern der Exchanges, Parametern der Coverage und Allowlist der Diagnose |
| Tests | neu | Fälle je Operation; die Fakes der Repositories und `FakeTelemetry` (beide Ports, aus `slice-telemetry-port`) halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

| Zusage | Fänger |
|---|---|
| *(Architect: eine Zeile je Zusage, Randform und Fall des Vertrags)* | *(ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates)* |

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Paralleler Zweig (Linie D): die Übergangsregeln in [`README.md`](../README.md) §Parallelbetrieb gelten für jede Rolle. Schwere Läufe (`make gates`, `make test-race`, `make gegenproben`, `make groessenmarke`, Mutationsläufe in Containern, jeder `docker build`) laufen als `flock /tmp/odata-recorder-heavy.lock <befehl>`, ein Lauf, dessen Wartezeit die Zeitgrenze des Aufrufs übersteigen kann, als Hintergrundlauf; vor einem großen Lauf `free -g`. Der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet (nach einem Lauf `git restore harness/gegenproben.stempel`). Roadmap, Welle-Pläne, Beobachtungs-Register, ADR-Index, `AGENTS.md`, Commands, Rollen-Typen und Skills schreibt dieser Zweig nicht; wer dort etwas ändern müsste, trägt es in §6 mit Ausgang *offen* und Adresse Planner ein. In `open/`-Pläne anderer Slices schreibt dieser Zweig nichts.
- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Vertrag von `InspectionUseCase`:** Operationen (Exchanges, Coverage, Diagnose), Argumente und Rückgabetypen, Fehlerabbildung als Anwendung von `SPEC-109`, Ort in der Spezifikation.
2. **Paginierung:** `SPEC-005` und die Architektur sagen „paginiert“, kein Vertrag nennt Parameter (Seitengröße, Zeiger oder Versatz), Standard und Obergrenze, Reihenfolge, Verhalten bei ungültigem oder zu großem Wert, Stabilität bei gleichzeitigem Anhängen, Gesamtzahl. `RecordingRepository.ListBySession` liefert alle Exchanges: erweitert der Architect den Port um eine Seitenabfrage (berührt `SPEC-105` und den SQLite-Adapter und damit die Größe) oder schneidet der Dienst im Speicher (Grenze für große Sessions)?
3. **Felder eines Exchange in der Liste:** Kopfdaten (Kennung, Sequenz, Methode, Pfad, Status, `capture_status`, Zeitpunkte, Vermerke der Redaction) und ob Header-Werte, Query und Bodies (oder nur Größe und Hash) erscheinen; „ohne Geheimnisse“ als Allowlist der Felder, nicht als Liste der entfernten.
4. **Coverage:** `SPEC-005` nennt für `GET /__admin/coverage` keinen Parameter. Auf welche Baseline bezieht sich der Aufruf (Baseline-Kennung, Workspace, Service, alle), Form der Antwort je EntitySet mit Evidenz, Fehlerfall einer unbekannten Baseline (Code?).
5. **Diagnose:** Allowlist der Schlüssel der Konfigurationszusammenfassung (nie Token, Pfade von Geheimnisdateien als Inhalt, Zugangsdaten), Form der Zähler je Fehlercode, und die Anwendung von `TelemetryReader.Snapshot().Errors` (entschieden in `SPEC-230` Nr. 8 und 9 und im Satz hinter der Tabelle von `SPEC-005`; der Architect prüft Form, Reihenfolge und Fehlerfall, entscheidet nichts neu); die Zusammenfassung kommt als typisierter Wert vom Composition Root, der Dienst sieht nie die Umgebung.
6. **`ODR-E7012`:** `SPEC-112` kennt „Exchange unbekannt“ (Admin), `SPEC-005` hat keinen Endpunkt für einen einzelnen Exchange: Zeile ergänzen oder den Code anders verorten.
7. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Aufträge an den Architect, Ergänzung beim Start** (Planner, 2026-11-07; vor dem ersten Code-Commit, `AGENTS.md` §3.12; Ausgang jeweils *offen* in §6):

8. **Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15) gegen den Vertrag aus Auftrag 1 bis 6: je Zusage, Randform und Fall eine Zeile, auch „Lesen schreibt nichts“ (Fake jedes Repositories zählt jede Operation, auch `Create`, `Update`, `Delete`, `Append`), „kein Geheimnis in Exchanges und Diagnose“ (Allaussage, Fänger über alle Felder der Allowlist, `BEO-REPO/universalaussage-im-vertrag-ungeprueft`, 14×), der `ctx` und der Kontextfehler je Aufrufstelle (Exchanges: Session und Exchanges; Coverage; Diagnose: Zusammenfassung und `Snapshot`), Gleichheit zweier verglichener Werte an jeder Grenze der Seite (Seitengröße gleich Obergrenze, Versatz gleich Zahl der Exchanges). `FakeTelemetry` ist `TelemetryReader` und `Telemetry`; der Dienst darf `Telemetry` nie rufen, der Fake zählt es.
9. **Pflicht-Mengen mit Achsen** (`AGENTS.md` §3.15, §3.17, `implement-slice` Schritt 14; Register `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, 29×, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`, 6×, `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze`, 11×): je Klasse die **Achsen**, die Zahl der Elemente als Produkt, der Generator, das **Orakel** (unabhängig vom Dienst: Modell der Seiten aus der Menge der Exchanges, nie der Dienst selbst) und die Prüfgrenze der unbeschränkten Menge als akzeptiertes Negativ mit Grund und Folge im Betrieb. Mindestens die Achsen **Seite** (Größe 0, 1, Standard, Obergrenze minus 1, Obergrenze, Obergrenze plus 1, negativ, sehr groß; Versatz oder Zeiger an jeder Grenze: 0, letzte, hinter der letzten, ungültig), **Bestand** (0, 1, genau eine Seite, eine Seite plus 1, mehrere Seiten; Anhängen zwischen zwei Seiten), **Feld** (jedes Feld der Allowlist der Exchange-Zeile; jede Art, in der ein Geheimnis eingebracht sein kann: Header, Query, Body, Pfad), **Coverage** (jeder Status, Evidenz vorhanden und fehlend, 0, 1, viele EntitySets; unbekannte, leere und fremde Baseline), **Diagnose** (jeder Schlüssel der Zusammenfassung gegen die Allowlist; jeder Modus; jeder Code der Tabelle von `SPEC-109` mit Zählerstand 0 und größer 0) und **Fehler** (jeder Fehler jedes Ports aus `SPEC-109` Nr. 5 an jeder Aufrufstelle, `ctx` abgebrochen vor und nach jedem Aufruf). Die Zählprüfung hält die Zahl der Elemente gegen die Menge der Quelle (`model.AlleCodes()`, die Menge der Statuswerte, die Felder des Typs), nie gegen eine feste Zahl im Test: `slice-mock-zuordnung` (Linie B) fügt `ODR-E6008` hinzu (Tabelle 37 auf 38 Codes; der Zähler `Snapshot().Errors` hat dann 38 Einträge).
10. **Zweige und Rückgaben als Menge mit Zählprüfung** (Register `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, 7×, Schwelle erreicht): der Architect nennt die **Modell-Menge** der Wege von `Exchanges`, `Coverage` und `Diagnose` (jeder Verwurf nach Grund, jede Seitengrenze, jede Fehlerabbildung je Port, jedes Feld der Allowlist) und die Tabelle führt je Element eine Zeile mit Fänger; der Test zählt die Menge gegen diese Modell-Menge (Zahl der Zweige im Code gleich Zahl der Zeilen), ein Zweig, den kein Lauf unter dem Detektor erreicht, ist ein benanntes Negativ mit Grund und Folge. Trägt der Vertrag eine Zusage über Nebenläufigkeit (Anhängen während des Lesens, Stand der Zähler), gilt `AGENTS.md` §3.10 je Fenster, sonst steht „gilt nicht“ mit Grund.
11. **Schnittfrage** nach `SPEC-128` Nr. 1: der Planner schätzt die Größe nach der Tabelle des Architect (Block `groessenmarke` im Zweig); über der Schnittgrenze wird vor dem Code geschnitten. Der Plan liefert drei Operationen mit eigener Seiten- und Allowlist-Regel; der Architect nennt vor der Tabelle einen Schnitt (etwa Exchanges mit Seitenabfrage, Coverage, Diagnose; jeder Teil ohne geteilten Mechanismus) oder den `schnitt-grund`, sonst schneidet der Planner nach der Tabelle. Entscheidet der Architect die Seitenabfrage in `RecordingRepository` und im SQLite-Adapter (Auftrag 2), liegen `src/hexagon/ports/outbound/` und `src/adapters/driven/sqlite/` im Block (`pfad`) und der Schnitt trennt diesen Teil zuerst.
12. **Kennungen und Dateien** (Linie D, Bereich `SPEC` 230 bis 259, `ARC` 040 bis 049, `ADR` 0040 bis 0049, `CO` 060 bis 079; letzte vergebene `SPEC` der Linie: `SPEC-235`): neue Nummern ab `SPEC-236`; `slice-telemetry-log` (Zweig) kann `SPEC-236` ebenfalls brauchen, der Zusammenstoß ist der des Integrationsschritts, der Architect nennt die Nummer im Plan. Der Slice ändert in `spezifikation.md` die Zeilen der Admin-API (`SPEC-005` ff.), `SPEC-105` und `SPEC-107`, soweit der Wortlaut es trägt, `SPEC-112` und hängt eine Zeile an die Änderungstabelle am Dateiende. Neben den Zweigen: `slice-mock-zuordnung` (Linie B) ändert `SPEC-160`, `SPEC-109` Nr. 5 und 6, `SPEC-174`, `SPEC-175`, `SPEC-195` bis `SPEC-199`, die Zeile `odr_errors_total` der Tabelle der Metriken und die Änderungstabelle; `slice-telemetry-log` (Linie D) ändert `SPEC-231`, `SPEC-232`, `SPEC-230` und die Änderungstabelle. Dieser Slice berührt `src/adapters/driven/odata/**`, `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/mock.go`, `src/adapters/driven/telemetry`, `src/hexagon/ports/support/` und `src/bootstrap/` nicht; `src/hexagon/ports/inbound/` nur um eine neue Datei `inspection.go`. Eine Änderung an `SPEC-109` oder `SPEC-230` geht als Rückgabe an den Planner.
13. **Katalog-Pflicht** (`SPEC-201` Nr. 3 und 11): erwartet ist kein Eintrag im Katalog des Binary (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`); bringt der Vertrag doch einen neuen Fehlercode, Schlüssel oder ein Kommando (etwa einen Code für Seitenparameter oder eine unbekannte Baseline), steht er im selben Slice in `harness/katalog-ausnahmen.txt` und im Handbuch und in §4 des Plans, oder der Architect gibt die Frage an den Planner zurück (Rückführung `in-progress` → `open`, §4).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-telemetry-port` (`welle-betrieb`) liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, Seitenform und Allowlist, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand beim Start** (Planner, 2026-11-07): erfüllt. `slice-telemetry-port` liegt in `done/` und liefert `TelemetryReader` mit `Snapshot().Errors` (33 Einträge, auch mit 0; `SPEC-230` Nr. 9) und `FakeTelemetry`; `slice-telemetry-log` und `slice-telemetry-spans` sind nicht Voraussetzung. Der Slice berührt keine Datei der Gate-Kette (Übergangsregeln Nr. 7: weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` noch den Stempel); in `in-progress/` liegen `slice-telemetry-log` (Linie D, allein `src/adapters/driven/telemetry`), `slice-mock-zuordnung` (Linie B, `src/adapters/driven/odata/**`, `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/mock.go`), beide keine Gate-Kette. Er berührt keine Datei der Zweige: `src/hexagon/ports/inbound/` nur mit einer neuen Datei `inspection.go`, `src/hexagon/services/inspection/` neu; `src/hexagon/ports/outbound/` und `src/adapters/driven/sqlite/` nur, wenn der Architect die Seitenabfrage entscheidet (kein anderer Zweig berührt sie). Gemeinsam bleiben nur `spec/spezifikation.md` (Aufträge 12 in §3) und die Pläne. Ändert Vertrag oder Code eine Datei der Gate-Kette, `src/bootstrap/` oder der Linie B (auch `.a-check.yml`, `.golangci.yml`), geht das als Rückgabe an den Planner (§6). Zweig `slice-inspection-use-case`, Worktree `/Development/odata-recorder-wt-inspection-use-case`.

**Warum, und für welchen Teil:** nur der Teil *Diagnose* liest Zähler je Fehlercode über `TelemetryReader` (`src/hexagon/ports/support/`, geliefert von `slice-telemetry-port`): *Code-Abhängigkeit* für diesen Teil. Die Teile *Exchanges* und *Coverage* lesen `RecordingRepository`, `SessionRepository` und `CoverageRepository`, die seit `welle-persistenz` in `src/hexagon/ports/outbound/` liegen, und brauchen nichts davon.

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`** (Mock, `WorkspaceUseCase`, HTTP-Adapter, Composition Root): der Dienst kennt weder HTTP noch einen Modus. Starteten Exchanges und Coverage vor dem Port, wäre das ein Schnitt der Diagnose in einen eigenen Slice (Umplanung durch den Planner); dieser Plan hält den Slice ganz und nennt den Port als Trigger. Ob `slice-telemetry-port` selbst vor dem Mock starten darf, entscheidet die Planung von `welle-betrieb`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.

**Größenmarke** (Stand beim Start, Planner): Der Block mit `basis`, `schaetzung` und `marke` steht auf `main` nicht (ein Block ohne diese Werte wäre ein Platzhalter). Rollenfolge: Der Architect legt den Block im Zweig an (`pfad`, `produktcode`, `fakes`; `basis` ist der Commit, der Vertrag, Tabelle und Raster abschließt), der Planner setzt `basis`, `schaetzung` und `marke` in eigenem Commit vor dem ersten Code-Commit nach `SPEC-128` Nr. 1 (Schätzung *S* = 85·*a* + 80·*b* + 200·*f* + *p*, Marke *M* = 100·⌈135·*S*/10000⌉; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Die Schnittfrage (*S* ≥ 2223) stellt der Planner nach der Tabelle; zur Frage an den Architect siehe Auftrag 11 in §3.
- `in-progress` → `open` (blockiert — Carveout?): `TelemetryReader` reicht für die Zähler der Diagnose nicht (Form oder Fehlerfall, Auftrag 5; Rückgabe an den Planner, der Port gehört `slice-telemetry-port` in `done/`), oder die Seitenabfrage verlangt eine Änderung von `RecordingRepository` und SQLite-Adapter, die die Marke sprengt (siehe §6).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): „(Vertrag, offen)“ heißt, der Architect trägt den Ort in der Spezifikation vor dem Übergang `next` → `in-progress` ein; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Exchanges einer Session lesen (Session, Seitenparameter) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Coverage lesen (Baseline oder Workspace) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Diagnose lesen (Konfiguration, Modus, Zähler) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- Die Paginierung hat keinen Vertrag (Auftrag 2) — **Ausgang:** offen; Adresse: Architect, vor dem ersten Code-Commit.
- Eine Seitenabfrage für `RecordingRepository` berührt `SPEC-105`, den Port und den SQLite-Adapter und damit Größe und Schnitt — **Ausgang:** offen; Adresse: Architect (Auftrag 2), danach der Planner für die Marke.
- Lesen der Zähler: `slice-telemetry-port` (`done/`) hat entschieden, dass `TelemetryReader` mit `Snapshot().Errors` liest, nicht `Telemetry` (`SPEC-230` Nr. 9, Sendung des Gebers an diesen Slice); der Nehmer trägt sie seit dem Start in Ziel, DoD, Auftrag 5 und §4 (`AGENTS.md` §3.13) — **Ausgang:** entfallen (die Frage nach einer Leseoperation ist entschieden); bleibt eine Lücke der Form oder des Fehlerfalls, steht sie hier neu.
- Auftrag 8 bis 13 (Tabelle, Pflicht-Mengen mit Achsen, Zweig-Menge mit Zählprüfung, Schnittfrage, Kennungen und Dateien, Katalog-Pflicht) — **Ausgang:** offen; Adresse: Architect (Tabelle, Raster, Schnitt, Kennungen), Implementer (Zählprüfung; Bericht je Befund mit Menge, Zahl der Elemente und einer Mutation an einem Element, das der Befund nicht nannte, `AGENTS.md` §3.17); bei Berührung der Gate-Kette, von `src/bootstrap/` oder der Linie B: Planner.
- Die Parameter von `GET /__admin/coverage` und der Code für eine unbekannte Baseline stehen nicht in `SPEC-005` (Auftrag 4), ebenso `ODR-E7012` ohne Endpunkt (Auftrag 6) — **Ausgang:** offen; Adresse: Architect.
- Geheimnisse in Exchanges und Diagnose: die Allowlist der Felder (Aufträge 3 und 5) ist der Schutz; der Fänger über alle Artefakte ist der Canary-Test in `slice-test-security` — **Ausgang:** offen bis zur Closure; Adresse: Architect.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): die Fakes der Repositories und `FakeTelemetry` tragen jede Schnittstelle des Originals und zählen jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** offen; Adresse: Planner nach der Tabelle des Architect (§4, Auftrag 11).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-07; Zähler = Zahl der Dateien in `evidence/`). Offen und einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (7×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure) für die Zweig-Menge (Auftrag 10); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (16×) für die Marke. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×) für die Fakes der Repositories und `FakeTelemetry`; `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (29×, `AGENTS.md` §3.17), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14) für die Mengen; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, beide `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.16, für „ohne Geheimnisse“). Mit diesem Slice erreicht kein offener Eintrag mit weniger als 3× die Schwelle.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
