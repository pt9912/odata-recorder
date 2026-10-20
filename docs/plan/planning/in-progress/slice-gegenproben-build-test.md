# Slice slice-gegenproben-build-test: Gegenproben zu make build und make test

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — der Closure-Trigger ist die DoD dieses Slice; er schließt den Meilenstein M1 der Roadmap.

**Bezug:** `LH-QA-06`, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-036` (Build und Unit-Tests, jetzt mit dem Prüfprogramm `pruefung`) · `SPEC-102` (Gegenproben).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-19.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make gegenproben` weist nach, dass die Prüfungen von `make test` (`gofmt`, `go vet`, Unit-Test) und `make build` (Version, `-trimpath`) einen Verstoß rot melden und einen erlaubten Fall nicht.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Gegenproben für `docs-check`, `schema-check` und `baseline-verify` — sie haben ihre Gegenprobe zum Teil in den Gates selbst; sie folgen bei Bedarf als eigener Slice.
- Änderungen am Verhalten von `make build` und `make test` — der Umbau verschiebt die Prüfungen in ein Go-Programm (`pruefung`), ändert aber ihre Wirkung nicht (`AGENTS.md` §3.6); die Meldungstexte dürfen sich ändern.
- Ein Docker-Build je Fall — die Fälle laufen wie die übrigen in einem Prozess ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), `SPEC-102`).

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

- [ ] Die Prüfungen von `make test` und `make build` liegen als Go-Programm `tools/pruefung` (`pruefung test`, `pruefung build`, Vertrag in `SPEC-036`); das `Dockerfile` ruft es, `go-gates.mk` bleibt ohne eigene Prüfung, und die Wirkung der Gates bleibt unverändert.
- [ ] Fälle 25 bis 33 in `tools/gegenproben` und `SPEC-102`: Basis, nicht formatierte Datei, `go vet`-Befund, roter Unit-Test, Binary ohne `-trimpath`, Binary mit falscher Version, fehlendes Binary, falscher Aufruf; jeder Verstoß-Fall nennt den Grund in der Ausgabe. Alles in Go, kein neues Shell-Skript.
- [ ] `make gegenproben` ist grün im Zeitbudget aus `SPEC-102`; M1 steht in der Roadmap auf *erreicht*.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/pruefung/` | neu | Go-Modul ohne Abhängigkeiten: die Prüfungen aus `Dockerfile` und `harness/mk/go-gates.mk` als `pruefung test` und `pruefung build` |
| `Dockerfile`, `harness/mk/go-gates.mk`, `.dockerignore` | update | Stufe `pruefung` (gofmt, vet, Test, Bau), Aufruf `pruefung test /src` in der Stufe `test` und `pruefung build /out/odata-recorder "$VERSION"` in der Stufe `build`; die zwei Prüfzeilen nach dem `docker build` in `go-gates.mk` entfallen; die Lint-Stufe kopiert `tools/pruefung` und prüft es als zweites Modul; die Stufe `gegenproben` bekommt `pruefung` |
| `.golangci.yml` | update | Die Ausnahme für `os.Stdout` und `os.Stderr` im Einstieg gilt auch für `tools/pruefung/main.go` (mit `# Why:`); keine Schwelle sinkt |
| `tools/harness/gegenproben-eingabe.sh` | update | `tools/pruefung` gehört zu den Eingaben; die Pfadliste steht einmal (Bash-Array) und dient der Existenzprüfung und `git ls-files` |
| `tools/gegenproben/` | update | Fälle 25 bis 33 und ihre Befehle; Fälle mit `Bau` erzeugen das Binary mit `go build` in der Kopie; `Umgebung.Programme` nennt `pruefung`, `go` und `gofmt`; die Eingabenliste des Programms bleibt (`pruefung` ist das Programm, keine Eingabe der Kopie), die Prüfsumme in `gegenproben-eingabe.sh` nennt `tools/pruefung` |
| `harness/vertraege/SPEC-102-gegenproben.md`, `SPEC-036-test.md`, `harness/README.md` | update | Fälle, Prüfprogramm und Randformen stehen vor dem Code im Vertrag; `harness/README.md` nennt `pruefung` bei den Sensors |
| `harness/mk/gegenproben.mk` | update | Kopf und `make help`-Text nennen `pruefung` (test, build) |

**Randform-Raster** (`AGENTS.md` §3.16; Ort in `SPEC-036`, Randformen von `pruefung`):

| Operation | fehlend | leer | unbekannt | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `pruefung test` | Wurzel fehlt: Nr. 2; Argument fehlt: Nr. 1 | leeres Argument: Nr. 1; ohne Go-Dateien: Nr. 8 | unbekanntes Kommando: Nr. 1; Programm nicht aufrufbar: Nr. 7 | Nr. 10 | gilt nicht: ohne Zustand, jeder Lauf ist neu (Nr. 10) |
| `pruefung build` | Binary fehlt: Nr. 3 | leere Version: Nr. 1 | kein Go-Binary: Nr. 5; startet nicht: Nr. 4; Version mit Zusatztext: Nr. 6 | Nr. 10 | gilt nicht: ohne Zustand (Nr. 10) |

**Zusage · Fänger** (`AGENTS.md` §3.15):

| Zusage | Fänger |
|---|---|
| `pruefung test` meldet eine nicht formatierte Datei | Fall `test/nicht-formatiert`; Mutation: `gofmt`-Prüfung entfernt |
| `pruefung test` meldet einen `go vet`-Befund | Fall `test/vet-befund`; Mutation: `go vet` entfernt |
| `pruefung test` meldet einen roten Unit-Test | Fall `test/roter-test`; Mutation: `go test` entfernt |
| `pruefung test` ohne Befund endet mit Exit 0 | Fall `test/basis` |
| `pruefung build` meldet fehlendes `-trimpath` | Fall `build/ohne-trimpath`; Mutation: Prüfung entfernt |
| `pruefung build` meldet falsche Version | Fall `build/falsche-version`; Mutation: Prüfung entfernt |
| `pruefung build` meldet ein fehlendes Binary mit Exit 2 | Fall `build/binary-fehlt`; Mutation: Prüfung entfernt |
| `pruefung` meldet falschen Aufruf mit Exit 2 | Fall `test/aufruf` fängt nur den Aufruf ohne Argumente (Mutation: der Zweig für null Argumente entfernt); die falsche Argumentzahl von `test` und `build`, ein unbekanntes und ein leeres Argument fängt `TestAufrufFehlerStartenNichts` in `tools/pruefung` (Mutation: `len(args) >= 2` bzw. `len(args) == 3` durch `true` ersetzt) |
| `pruefung build` ohne Befund endet mit Exit 0 | Fall `build/basis` |
| Version mit Zusatztext, Binary startet nicht, kein Go-Binary, Wurzel fehlt (Randformen Nr. 2, 4, 5, 6) | Tests in `tools/pruefung`; Mutation: jeweilige Prüfung entfernt |
| Version mit abschließendem oder führendem Leerzeichen und mit `\r\n` (Randform Nr. 6) | `TestBuildVersionBefund`; Mutation: `strings.TrimRight(…, "\n")` durch `strings.TrimSpace(…)` ersetzt |
| Befunde und Fehler nach stderr, `ok` nach stdout (Randform Nr. 11) | `TestStroeme`; Mutation: `ok`-Zeile nach stderr, Befundtext nach stdout |
| Argumente mit führendem `-` gehen unverändert an `go` (Randform Nr. 13) | `TestTestMusterMitStrichGehenUnveraendertAnGo`; Mutation: Argument mit Strich verworfen |
| Exit-Codes 0, 1, 2 (Randform Nr. 14) | Gegenproben: Exit 2 bei `test/aufruf` und `build/binary-fehlt`, Exit 1 bei den Verstoß-Fällen; die Unit-Tests vergleichen gegen die Konstanten und pinnen die Zahlen nicht |
| Falscher Aufruf, `Binary` fehlt, Programm nicht aufrufbar, Verzeichnis ohne Go-Dateien, kein Zustand (Randformen Nr. 1, 3, 7, 8, 10) | Tests in `tools/pruefung`; Mutation: jeweilige Prüfung entfernt, bei Nr. 8 eine eigene Prüfung auf Go-Dateien hinzugefügt, bei Nr. 10 eine Datei in `Wurzel` geschrieben |
| Die Eingabenliste nennt `tools/pruefung` (Existenz) | `gegenproben-eingabe.sh` endet mit Exit 1, wenn `tools/pruefung` fehlt (in einer Wegwerfkopie gesehen); dass die Prüfsumme die Dateien darin erfasst, fängt nichts (`SPEC-102` §Grenze) |
| Die Stufen `test` und `build` rufen `pruefung` | Gate: ein roter Verstoß im Quellstand lässt `make test` bzw. `make build` rot werden — Beleg im Bericht des Implementers, kein Fall (Grenze in `SPEC-102`) |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-lint-gegenproben-modul](../done/welle-fundament/slice-lint-gegenproben-modul.md) liegt in `done/`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

*Marke überschritten:* 974 Zeilen (Summe aus `git diff --numstat 2091ba6..HEAD` über die Pfade aus §3 plus `.golangci.yml`), davon der größere Teil Tests von `pruefung` und der Gegenproben sowie der Nachzug nach dem Review. Der Slice bleibt ungeteilt: ein Schnitt hätte `pruefung` von seinen Fällen getrennt, und keine Hälfte wäre einzeln lieferbar.

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 900 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; Schätzung 600, weil `pruefung` ein Programm mit eigenen Tests ist).
- `in-progress` → `open` (blockiert): `pruefung` lässt sich im Image der Stufe `gegenproben` nicht fahren, ohne ein weiteres Image zu pinnen.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, jede Mutation aus der Tabelle wurde rot gesehen und M1 steht in der Roadmap auf *erreicht*; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen von `pruefung` (`AGENTS.md` §3.12, §3.16): Raster in §3, entschieden in `SPEC-036` und `SPEC-102` vor dem ersten Code-Commit — **Ausgang:** wird bei der Closure eingetragen.
- Die Fälle verlangen ein Binary, das nur im Image der Stufe `build` entsteht — **Ausgang:** weiter offen: der Slice baut es im Prozess mit `go build` in der Kopie; weicht das vom Image ab, steht es als Beobachtung im Register.
- `pruefung` ändert das Verhalten der Gates unbemerkt — **Ausgang:** wird bei der Closure eingetragen; Beleg: `make gates` grün, die Mutation je Zusage und ein roter Verstoß im Quellstand je Ziel.
- Zeitbudget: neun Fälle mit `go`-Aufrufen können die 30 s aus `SPEC-102` sprengen — **Ausgang:** weiter offen: Budget für 33 Fälle unbelegt (Messung unten), kein Lockern im Code.
  Messung bei Last (Lastmittel 12 bis 83 auf 20 Kernen, `time make gegenproben`, Images lokal): 38,5 s in der Stufe, 51 s gesamt mit den neun Fällen; die 24 früheren Fälle allein brauchen im Container 33 bis 55 s Wanduhr und 7,3 bis 7,8 Minuten CPU, die neun Fälle fügen 1,3 bis 1,7 Minuten CPU hinzu. Die 30 s sind auf dieser Maschine auch ohne die neun Fälle nicht zu halten; der Punkt ist beim Planner: `SPEC-102` nennt jetzt, dass das Budget für 33 Fälle nicht belegt ist; eine Messung auf ruhiger Maschine steht aus.
- `<Binary> version` endet mit Exit ungleich 0 und gibt die erwartete Version aus: `pruefung` meldet es als Befund `Binary startet nicht:` (Exit 1); der Test `TestBuildBinaryStartetNicht` fängt es — **Ausgang:** entfallen: vom Planner entschieden und in `SPEC-036` Randform Nr. 4 eingetragen.

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

Wird bei der Closure geschrieben.

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

**Vorgelagert — offene Beobachtungen sichten:** Im Beobachtungs-Register sind `BEO-REPO/randform-erst-im-review-entschieden` und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` verkörpert (`AGENTS.md` §3.16 und §3.15); beides fließt als Raster und Tabelle in §3 ein. `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (2×) betrifft die nebenläufigen Fälle.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
