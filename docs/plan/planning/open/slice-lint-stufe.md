# Slice slice-lint-stufe: Lint-Stufe für den Code samt eigener Prüfungen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-fundament](../welle-fundament.md).

**Bezug:** `LH-QA-06`, [ADR-0003](../../adr/0003-lint-profil-golangci.md), [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-101` (Lint-Profil).
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

**Ziel:** `make lint` prüft den Go-Code nach `.golangci.yml` und mit eigenen Prüfungen und ist ein Gate.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Gegenprobe des Lint-Gates — eigener Slice [slice-gegenproben](slice-gegenproben.md); sie braucht dieses Gate als Gegenstand.
- Neue Linter oder gelockerte Schwellen — das Profil ist mit [ADR-0003](../../adr/0003-lint-profil-golangci.md) beschlossen; eine Änderung ist eine neue ADR (`AGENTS.md` §3.6).
- Zulassung von Fremdbibliotheken in `gomodguard_v2` — jede Bibliothek bekommt ihre ADR in welle-persistenz.
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

- [ ] `Dockerfile`-Stufe `lint`: golangci-lint nach `.golangci.yml` auf den Modulen der Stufe `deps`, ohne Netz; das Image kommt per Build-Argument aus `harness/mk/lint-profil.mk`, damit der Pin an einer Stelle steht.
- [ ] `tools/harness/lint.sh` mit den eigenen Prüfungen: keine Direktive `//nolint` und `//lint:ignore` im Code, jede Regel unter `exclusions.rules` mit `# Why:`, `warn-unused` für eine Regel ohne Wirkung.
- [ ] `make lint` hängt an `GATE_CHECKS`; Zeile in `harness/README.md` §Sensors. `make lint-profil` bleibt als Schema-Prüfung bestehen.
- [ ] `make gates` grün.
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
| `Dockerfile` | update | Stufe `lint` aus `source`, Image per Build-Argument |
| `tools/harness/lint.sh` | neu | eigene Prüfungen nach [ADR-0003](../../adr/0003-lint-profil-golangci.md) und `AGENTS.md` §3.2 |
| `harness/mk/lint-profil.mk` | update | Target `lint` neben `lint-profil`, beide in `GATE_CHECKS` |
| `harness/README.md`, `harness/vertraege/SPEC-101-lint-profil.md` | update | Sensors-Zeile; `SPEC-101` nennt die Code-Prüfung |

- Das Profil verlangt `go.mod` und Pakete; der Slice läuft deshalb erst nach `slice-go-skelett-build-test`.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-go-skelett-build-test](../next/slice-go-skelett-build-test.md) liegt in `done/`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 500 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert): das Profil lässt sich im Container nicht gegen das Skelett fahren, ohne ein Fremdpaket zuzulassen.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün mit `make lint` in der Kette, und ein Probe-Befund im Skelett (etwa eine `//nolint`-Direktive) macht `make lint` rot, bevor er zurückgenommen wird; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen des Profils und der eigenen Prüfungen (`AGENTS.md` §3.12): `.golangci.yml` fehlt oder ist nicht lesbar, das Profil weicht von der gelesenen Form ab, eine Direktive `//nolint` steht in einer Testdatei, ein Verzeichnis enthält keine Go-Dateien — **Ausgang:** entfallen: entschieden in `harness/vertraege/SPEC-101-lint-profil.md` vor dem ersten Code-Commit.
- Die Schwellen des Profils (Komplexität 15, Funktionslänge 100) beißen am Skelett nicht — **Ausgang:** weiter offen: wird mit der ersten Fachlogik sichtbar und dann als Beobachtung geführt.
- `lint.sh` liest das Profil in fester Form; eine Umformatierung bricht es — **Ausgang:** entfallen: die Form steht im Kopfkommentar des Profils und die Gegenprobe in [slice-gegenproben](slice-gegenproben.md) hält sie.

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
