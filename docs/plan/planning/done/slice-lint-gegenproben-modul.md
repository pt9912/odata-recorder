# Slice slice-lint-gegenproben-modul: `tools/gegenproben` unter dem Lint-Gate

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-fundament](../welle-fundament.md).

**Bezug:** `LH-QA-06`, [ADR-0003](../../adr/0003-lint-profil-golangci.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-101` (Lint-Profil) · `SPEC-102` (Gegenproben).
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

**Ziel:** `make lint` prüft auch das Go-Modul `tools/gegenproben` nach `.golangci.yml`, und das Programm ist frei von Befunden.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Neue Linter oder gelockerte Schwellen — das Profil bleibt, wie es mit [ADR-0003](../../adr/0003-lint-profil-golangci.md) beschlossen ist; der Code passt sich an (`AGENTS.md` §3.6).
- Weitere Gegenproben für andere Gates — [slice-gegenproben](../done/slice-gegenproben.md) liegt in `done/`; dieser Slice fügt nur Fälle zu `lint.sh` hinzu, soweit seine Zusagen (Modul-Schleife, Prüfung 4) eine fangende Mutation brauchen.
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

- [x] `tools/gegenproben` ohne Befund des Profils: die Globals werden zu einer übergebenen Umgebung, `RemoveAll` wird geprüft, `greife` zerlegt, die Bedingung zu `go.sum` ausgelagert; die Zeichenkette `//nolint` in den Fällen wird aus zwei Teilen gebildet.
- [x] `make lint` prüft beide Module (Stufe `lint` bekommt `tools/gegenproben` in den Kontext, `lint.sh` ruft `golangci-lint` je Modul); `main.go` des Programms erhält die Ausnahme für `os.Stdout` und `os.Stderr` im Profil mit `# Why:`.
- [x] `SPEC-101` und `SPEC-102` ziehen nach (die Grenze „außerhalb von `make lint`" entfällt); `make gegenproben` bleibt grün.
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
| `tools/gegenproben/main.go`, `faelle.go`, `main_test.go` | refactor | Umgebung statt Globals, `RemoveAll` geprüft, `greife` zerlegt, Bedingung zu `go.sum` ausgelagert, Zeichenkette `//nolint` aus zwei Teilen |
| `tools/harness/lint.sh` | update | `golangci-lint` je Modul (Wurzel und `tools/gegenproben`) |
| `Dockerfile` | update | Stufe `lint` bekommt `tools/gegenproben` |
| `.golangci.yml` | update | Ausnahme für `os.Stdout` und `os.Stderr` in `tools/gegenproben/main.go` mit `# Why:` |
| `harness/vertraege/SPEC-101-lint-profil.md`, `harness/vertraege/SPEC-102-gegenproben.md`, `harness/README.md` | update | Grenze „außerhalb von `make lint`" entfällt |

- `lint.sh` und `.golangci.yml` stehen schon in den Eingaben der Gegenproben; der Stempel ist nach dem Umbau neu zu schreiben.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-lint-stufe](../done/slice-lint-stufe.md) und [slice-gegenproben](../done/slice-gegenproben.md) liegen in `done/`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 300 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert): das Profil lässt sich nicht erfüllen, ohne einen Linter zu lockern.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün mit beiden Modulen unter `make lint`, ein Probe-Befund in `tools/gegenproben` macht `make lint` rot, bevor er zurückgenommen wird, und `make gegenproben` ist grün; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen des Umbaus (`AGENTS.md` §3.12): ein Modul ohne Go-Datei, ein Modul mit eigener `go.mod`, ein Befund im zweiten Modul — **Ausgang:** entfallen: entschieden in `SPEC-101` vor dem ersten Code-Commit.
- Der Umbau ändert das Verhalten des Programms — **Ausgang:** entfallen: die Fälle des Programms bleiben grün (zwei ändern nur die Schreibweise der Direktiv-Zeichenkette, die Tests werden an die Umgebung angepasst und um die Randform 8 erweitert).

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

- **Was hat funktioniert:** `make lint` prüft beide Module (Wurzel und `tools/gegenproben`, je „0 issues"); die Randformen standen vor dem Code in `SPEC-101`, und der Umbau des Programms (Umgebung statt Globals) ließ alle Tests und Fälle grün; je Zusage wurde eine Mutation rot gesehen (Modul-Schleife, Prüfung 4, Ausnahme für `main.go`, Aufräum-Fehler, Umgebung).
- **Was ging anders als geplant:** Das Review fand die Modul-Schleife ohne fangende Gegenprobe (F-1) und die Behandlung des Aufräum-Fehlers als im Code entschiedene Randform (F-2); die Verifikation fand Prüfung 4 im Modul-Zweig und den Ausschluss von `tools/` nur halb gefangen (V-1). Fünf neue Fälle in `SPEC-102` (zwölf bis sechzehn) fangen sie, `SPEC-102` Randform 8 nennt den Aufräum-Fehler. Die Zahl „elf Fälle" in [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) ist als Tatsachenberichtigung auf „die Fälle" gesetzt, damit sie nicht altert.
- **Steering-Loop-Eintrag:** Nichts verkörpert; gezählt, nicht verkörpert: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` steht jetzt bei 4× (der Eintrag geht in die Closure von welle-fundament) und `BEO-REPO/randform-erst-im-review-entschieden` bei 2×.
- **Beobachtungs-Register (`../observations/`):** je eine weitere Datei `evidence/slice-lint-gegenproben-modul.md` in beiden Einträgen.
- **Folge-Slices:** keine; offen bleiben die Enge der Ausnahme für `main.go` (kein Test, im Vertrag nicht zugesagt) und Module tiefer als `tools/<Name>` (Grenze in `SPEC-101`).
- **Risiken aus §6:** beide mit dem Ausgang *entfallen*, siehe §6.
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
