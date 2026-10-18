# Slice slice-go-skelett-build-test: Go-Skelett mit Build und Test in Containern

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-fundament](../welle-fundament.md).

**Bezug:** `LH-QA-06` (Reproduzierbarkeit), `LH-RB-03` (lokale Entwicklungsumgebung), [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-036` (`make test`) · `ARC-012` (Composition Root).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912

**Autor:** pt9912. **Datum:** 2026-10-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Go-Modul mit Composition Root unter `src/bootstrap` baut und testet in Containern, und `make build` und `make test` sind Gates.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Fachlogik, Ports und Adapter — das Skelett trägt nur Composition Root und Version; die Schichten entstehen mit ihren Wellen, sonst prüfte `a-check` leere Verzeichnisse.
- Coverage-Schwelle und Integrationstests — eine Schwelle über einem Skelett misst nichts; sie kommen mit der ersten Fachlogik (Folge: welle-persistenz).
- Lint — eigener Slice [slice-lint-stufe](../open/slice-lint-stufe.md), weil er ein zweites Werkzeug samt eigener Prüfungen einführt.
- Produkt-Image und Release — erst sinnvoll, wenn der Prozess etwas tut.
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

- [x] Go-Modul `github.com/pt9912/odata-recorder` mit Composition Root unter `src/bootstrap` (Paket und Startpunkt, Version über `-ldflags`) und einem Test (Paket `bootstrap_test`).
- [x] `Dockerfile` mit den Stufen `deps`, `source`, `test` (`gofmt`, `go vet`, `go test`) und `build` (reproduzierbar, `-trimpath`); Eingangs-Image per Digest gepinnt, Stufen nach `deps` ohne Netz.
- [x] `make build` und `make test` laufen über das `Dockerfile` und hängen an `GATE_CHECKS`; Zeilen in `harness/README.md` §Sensors.
- [x] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review.
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; kein Zähler wird gesetzt. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `go.mod` | neu | Modul `github.com/pt9912/odata-recorder` ([ADR-0001](../../adr/0001-implementierungssprache-go.md)); keine Fremdabhängigkeit |
| `src/bootstrap/bootstrap.go` | neu | Composition Root: `Run(args, stdout, stderr) int` gibt die Version aus; kein Fremdpaket |
| `src/bootstrap/main/main.go` | neu | Startpunkt, ruft `bootstrap.Run` und ist die einzige Stelle mit `os.Exit` |
| `src/bootstrap/bootstrap_test.go` | neu | Test der Version und der Fehlerausgabe (Happy, Negative) nach `LH-QA-06` |
| `Dockerfile`, `.dockerignore` | neu | Stufen `deps`, `source`, `test`, `build` nach dem Muster von `pgwire-recorder` |
| harness/mk/go-gates.mk | neu | `build` und `test`, beide in `GATE_CHECKS` |
| `.dockerignore` | neu | Allowlist des Build-Kontexts: `go.mod`, `go.sum`, `src/` |
| `harness/vertraege/SPEC-036-test.md`, `harness/vertraege/README.md` | update | Vertrag nennt `make build` und `make test` mit dem, was sie zusagen |
| `harness/README.md` | update | Sensors-Zeilen für `make build` und `make test` |

- Der Startpunkt liegt unter `src/bootstrap/main`, weil Go ein `main`-Paket braucht und `src/bootstrap` das Composition-Root-Glob von `.a-check.yml` ist.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): Die ADRs [0001](../../adr/0001-implementierungssprache-go.md), [0002](../../adr/0002-architektur-gate-a-check.md) und [0003](../../adr/0003-lint-profil-golangci.md) sind `Accepted` (Trigger der Welle).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 400 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert): kein per Digest pinnbares Go-Image ist verfügbar.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün und `make build` und `make test` laufen aus einem frischen Klon ohne Host-Toolchain; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Das Go-Toolchain-Image altert still — **Ausgang:** weiter offen: wird mit der Pin-Hebung des Dockerfiles als Beobachtung geführt (kein Slice nötig).
- `src/bootstrap/main` weicht von der Baum-Skizze der Architektur-Sicht ab — **Ausgang:** entfallen: die Sicht nennt `src/bootstrap` als Composition Root, ein Unterverzeichnis davon bleibt innerhalb des Globs.

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

**Vorgelagert — offene Beobachtungen sichten:** Das [Beobachtungs-Register](../observations/README.md) trägt nur seine `README.md`; keine Treffer.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
