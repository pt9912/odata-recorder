# Slice slice-gegenproben: Gegenproben für a-check und Lint, außerhalb der Gate-Kette

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-fundament](../welle-fundament.md).

**Bezug:** `LH-QA-06`, [ADR-0002](../../adr/0002-architektur-gate-a-check.md), [ADR-0003](../../adr/0003-lint-profil-golangci.md)

**Berührte Spec-Stellen:** `SPEC-039` (`make a-check`) · `SPEC-101` (Lint-Profil) · `SPEC-102` (Gegenproben).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make gegenproben` weist in Kopien des Arbeitsbaums nach, dass `a-check` und `make lint` einen Verstoß rot melden und einen erlaubten Fall nicht, und ist schnell: ein Go-Programm fährt alle Fälle nebenläufig in einem Container, ohne je Fall einen Docker-Build oder -Lauf, und er läuft nur, wenn sich eine Eingabe der Gegenproben geändert hat (Prüfsumme als versionierter Stempel).
Zu den Fällen für `make lint` gehören die Randformen aus `SPEC-101`, die über das Gate nicht erreichbar sind oder nur ein zweiter Linter deckt: ein fehlendes Profil, eine Direktive mit zwei Leerzeichen nach `//` und die Einrückung des `# Why:`-Blocks (aus [slice-lint-stufe](../done/slice-lint-stufe.md), Review-Befunde F-4 bis F-6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Gegenproben anderer Gates, die ein Docker-Image je Fall brauchen — eine Gegenprobe ohne Zeitbudget ist in `pgwire-recorder` aus der Gate-Kette gefallen; dieser Slice setzt das Budget (§4) und liefert nur die zwei Gegenproben.
- Gegenproben für `docs-check`, `schema-check` und `baseline-verify` — sie haben ihre Gegenprobe zum Teil in den Gates selbst, und kein neuer Verstoß ist offen; sie folgen bei Bedarf als eigener Slice.
- Aufnahme der Gegenproben in `make gates` — sie laufen länger und prüfen die Gates, nicht den Stand; die Entscheidung hält die ADR in diesem Slice fest.
- Neue Gate-Regeln — die Gegenprobe prüft die vorhandenen Regeln und fügt keine hinzu.
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

- [x] Gegenprobe `a-check-negativ` (Go-Programm `tools/gegenproben`, ein Prozess, Fälle nebenläufig): In einer Kopie des Arbeitsbaums lehnt `a-check` einen Import des Domain-Kerns auf einen Driven Adapter und einen Import zwischen zwei Adaptern ab und lässt den erlaubten Weg über einen Port zu.
- [x] Gegenprobe `lint-gegenprobe`: In einer Kopie lehnt `make lint` eine `//nolint`-Direktive und eine Ausnahme ohne `# Why:` ab und lässt eine begründete Ausnahme zu.
- [x] `make gegenproben` läuft nur bei geänderter Eingabe (Prüfsumme gegen den versionierten Stempel; Änderung an Doku, Plan oder Reports löst keinen Lauf aus, ein roter Lauf lässt den Stempel stehen) und fährt beide außerhalb von `GATE_CHECKS` in einem Docker-Build der Stufe `gegenproben` und bleibt im Zeitbudget aus §4; die Entscheidung steht als ADR (Gegenproben verlassen die Gate-Kette), Zeilen in `harness/README.md` §Sensors.
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review.
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; kein Zähler wird gesetzt. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/gegenproben/` (eigenes Go-Modul mit `go.mod`) | neu | ein Programm, das je Fall eine Kopie des Baums in einem Wegwerfverzeichnis anlegt, den Verstoß einfügt und `a-check` bzw. `lint.sh` und `golangci-lint` als Prozess ruft; die Fälle laufen nebenläufig, die Kopie ist ein Dateikopieren im Speicherdateisystem, kein `docker cp`; ein eigenes Modul, damit `go test ./...` und `make lint` des Produkts es nicht mitnehmen |
| `Dockerfile` | update | Stufe `gegenproben` auf dem golangci-lint-Image mit dem `a-check`-Binary aus dessen Image (`COPY --from`), baut das Programm einmal und führt es aus; ohne Netz, die Stufe kopiert nur die Eingaben der Gegenproben; `--no-cache-filter` lässt sie bei geänderter Eingabe immer laufen, damit ihre Ausgabe erscheint |
| `tools/harness/gegenproben-eingabe.sh`, `harness/gegenproben.stempel` | neu | Prüfsumme der Eingaben und Pins; der Stempel ist versioniert und trägt die Prüfsumme des letzten grünen Laufs |
| `.dockerignore` | update | lässt `.a-check.yml` und `tools/gegenproben` in den Kontext (`go.sum` steht schon dort) |
| `tools/harness/lint.sh`, `.golangci.yml`, `harness/vertraege/SPEC-101-lint-profil.md` | update | der `# Why:`-Block der Regel ist wie sie eingerückt (Review-Befund F-6 aus dem Folge-Slice-Eintrag) |
| harness/vertraege/SPEC-102-gegenproben.md | neu | Vertrag der Gegenproben: Namensregel, Liste, Kopie, Verstoßformen, Exit-Codes |
| harness/mk/gegenproben.mk | neu | Ziel `gegenproben` (ein `docker build --target gegenproben`); nicht in `GATE_CHECKS` |
| docs/plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md (neu), `docs/plan/adr/README.md` | neu / update | Entscheidung mit drei Alternativen; Index mitziehen |
| `harness/README.md` | update | Sensors-Zeilen für die Gegenproben |

- Die Gegenproben arbeiten in einer Kopie des Arbeitsbaums (Wegwerfverzeichnis im Container), nie im Arbeitsbaum selbst.
- Geschwindigkeit ist Teil des Vertrags: kein Docker-Build und kein Container je Fall, ein gemeinsamer Go-Build-Cache für die Fälle, und die Fälle laufen nebenläufig.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-go-skelett-build-test](../done/slice-go-skelett-build-test.md) und [slice-lint-stufe](../done/slice-lint-stufe.md) liegen in `done/`.

**Zeitbudget:** `time make gegenproben` bei geänderten Eingaben und kaltem Docker-Build-Cache höchstens 30 s, bei unveränderten Eingaben höchstens 1 s (Wanduhr, Entwicklungsrechner; gemessen 13,5 s und 0,04 s); `SPEC-102` trägt die Zahlen.

**Marke überschritten:** Die Summe aus `git diff --numstat` über die Pfade aus §3 liegt am Ende bei mehr als 600 Zeilen, getragen von Test- und Gegenprobe-Code; der Slice bleibt ungeteilt, weil Programm, Fälle, Stempel und Vertrag einzeln nichts liefern.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 600 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen) (Gegenprobe-Code zählt etwa doppelt).
- `in-progress` → `open` (blockiert): das Skelett bietet keinen Ort, an dem ein Verstoß ohne Fremdpaket einzufügen ist.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, `make gegenproben` ist grün und die ADR ist `Accepted` samt Index-Eintrag; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen der Gegenproben (`AGENTS.md` §3.12): die Kopie des Baums scheitert, Docker fehlt, die einzufügende Datei fehlt, ein Verstoß wird nicht gefunden — **Ausgang:** entfallen: entschieden in SPEC-102 vor dem ersten Code-Commit.
- Der Docker-Cache fehlt (neuer Klon, `docker builder prune`) — **Ausgang:** entfallen: die Prüfsumme im versionierten Stempel entscheidet, nicht der Cache; der kalte Lauf hält das Budget.
- Der versionierte Stempel kann von Hand geschrieben werden und belegt dann nichts — **Ausgang:** weiter offen: kein Gate liest ihn (`SPEC-102`, Grenze); die Pflicht zum Fahren steht in `SPEC-102`.
- Das Zeitbudget hält nicht, weil `golangci-lint` je Fall seinen Cache neu füllt — **Ausgang:** entfallen: gemessen 13,5 s bei geänderter Eingabe; die Fälle teilen den Cache, aber nicht die Sperre von `golangci-lint` (eigenes `TMPDIR` je Fall, `SPEC-102` Randform 11), sonst brechen erlaubte Fälle nebenläufig ab.
- Eine Gegenprobe prüft nur die eingefügten Verstoßformen — **Ausgang:** weiter offen: weitere Formen kommen mit den Bibliotheks-Grenzen (`tech`) in welle-persistenz.
- Die Gegenproben laufen nicht in `make gates` und altern unbemerkt — **Ausgang:** entfallen: `SPEC-102` legt den Lauf am Ende jeder Welle und vor dem Release fest, und welle-dev-container fährt ihn automatisch in einem eigenen Container.

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

- **Was hat funktioniert:** Die Fälle und Randformen standen vor dem Code in `SPEC-102`; ein Go-Programm fährt elf Fälle nebenläufig in einem Docker-Build (13,5 s bei geänderter, 0,04 s bei unveränderter Eingabe), und je Zusage wurde eine Mutation rot gesehen (Kern-Rolle, Adapter-Glob, Port-Kante, jede Prüfung von `lint.sh`, Prozessgruppen-Kill, Aufräumen, veränderte Eingabe, `TMPDIR` je Fall).
- **Was ging anders als geplant:** Das Review fand drei zugewiesene Lint-Fälle ohne Entsprechung im Vertrag (F-1) und eine wirkungslose Frist (F-2); die Verifikation fand die Sperre von golangci-lint, die nebenläufige Fälle abbrechen ließ (V-1, HIGH); alles ist behoben, ein Nachreview fand 0 HIGH. Die Rückführungs-Marke von 600 Zeilen ist überschritten und im Plan benannt.
- **Steering-Loop-Eintrag:** Nichts verkörpert; gezählt, nicht verkörpert: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` steht jetzt bei 3× und geht in die Steering-Loop-Einträge der Closure von welle-fundament; `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` 1×.
- **Beobachtungs-Register (`../observations/`):** `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft/` fortgeschrieben (`evidence/slice-gegenproben.md`), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen/` neu angelegt; `BEO-REPO/randform-erst-im-review-entschieden` trat nicht auf.
- **Folge-Slices:** keine; offen bleiben Randform 10 (fehlende Eingabe im Prüfsummen-Skript, nur im Handlauf belegt), `tools/gegenproben` außerhalb von `make lint` (Review F-8; übernimmt [slice-lint-gegenproben-modul](../open/slice-lint-gegenproben-modul.md)) und die Gegenproben von `docs-check`, `schema-check` und `baseline-verify` (§1).
- **Risiken aus §6:** Randformen *entfallen*; Verstoßformen *weiter offen*; Altern der Gegenproben *entfallen*; Docker-Cache *entfallen*; Stempel von Hand *weiter offen*; Zeitbudget *entfallen*, siehe §6.
- **Drei Paarungen:** von der Closure der Welle welle-fundament geprüft.

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

**Vorgelagert — offene Beobachtungen sichten:** Das [Beobachtungs-Register](../observations/README.md) trägt nur seine `README.md`; keine Treffer.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
