# Verifikations-Report: slice-lint-stufe — 2026-10-19

**Frage:** Bauen wir es richtig? Gegen DoD 1–4 und die Closure-Trigger (§5) des Plans, [SPEC-101](../../harness/vertraege/SPEC-101-lint-profil.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), `LH-QA-06`.

**Gegenstand:** `git diff 7f425d3..HEAD` (Stand `862961e`), Plan `docs/plan/planning/in-progress/slice-lint-stufe.md`. Eingang: Review `docs/reviews/2026-10-19-slice-lint-stufe-review.md`; F-1, F-2, F-4 im letzten Commit behoben, F-3, F-5, F-6 bewusst offen.

**Rolle/Datum:** Verifier · 2026-10-19. Das Repo blieb unverändert (`git status` sauber nach der Mutation).

## Urteil

**DoD 1–4: erfüllt. Closure-Trigger §5: Sensor-Teile erfüllt, Closure-Notiz mit Lerneintrag fehlt noch (§7 „Wird bei der Closure geschrieben"); DoD 5–10 nicht Gegenstand dieser Verifikation und offen.** Keine Verifier-only-Klasse (DoD-Verletzung, Spec-Abweichung) gefunden; zwei Hinweise unten. Der Slice kann nach Review-Abschluss und Closure-Notiz nach `done/`.

## Selbst gefahrene Sensoren (keine Übernahme von Behauptungen)

| Sensor | Ergebnis |
|---|---|
| `make gates` | Exit 0; `lint-profil` und `lint` stehen in der Kette (Build-Zeile `--target lint`, `RUN --network=none sh /usr/local/bin/lint.sh` Stufe 7/7); `schema-check` bis `integrity_check` ok |
| Probe: `// nolint:unused // probe` in `src/bootstrap/bootstrap.go` angehängt, `make lint` | rot: `lint.sh: ./src/bootstrap/bootstrap.go:38: Direktive nolint oder lint:ignore`, Build-Fehler Exit 1; danach `git checkout src/bootstrap/bootstrap.go`, Arbeitsbaum sauber |
| Gegenprobe F-1: Profil mit `    rules: []` gegen `lint.sh` (Wegwerfkopie im Scratchpad) | kein Befund der Prüfung 2 (Randform 3 eingehalten); golangci-lint dort nicht vorhanden, nur der Skript-Teil bewertet |
| Rückführungs-Marke §4 (500 Zeilen, Summe `git diff --numstat 7f425d3..HEAD` über die Pfade aus §3) | 131; Marke nicht gerissen |

## Findings je Stufe

### DoD 1 — Dockerfile-Stufe `lint` — erfüllt

- Stufe `lint` läuft golangci-lint nach `.golangci.yml` auf `/go/pkg/mod` aus `deps` und `src/` aus `source`, `RUN --network=none`.
- Image per `ARG LINT_IMAGE`, gesetzt aus `GOLANGCI_LINT_IMAGE` in `harness/mk/lint-profil.mk`; der Digest steht an einer Stelle (beide Targets nutzen ihn).
- `.dockerignore` lässt `.golangci.yml` und `tools/harness/lint.sh` zu; Kommentar nach F-2 angepasst („Build, Test und Lint").
- Hinweis (F-3, offen, INFO): `ARG LINT_IMAGE=scratch` als Vorgabe; nur der direkte Aufruf ohne `make` betroffen. Kein DoD-Bruch.

### DoD 2 — `tools/harness/lint.sh` — erfüllt

Gegen SPEC-101 geprüft:

| SPEC-101 | Code | Urteil |
|---|---|---|
| Prüfung 1: Direktiven `nolint`/`lint:ignore`, überall, auch Tests | `grep -rnE --include='*.go' '//[[:space:]]*(nolint\|lint:ignore)'`; Probe rot (s. o.); F-4 (mehrere Leerzeichen) behoben | konform |
| Prüfung 2: `# Why:` je Regel | awk-Block, Leerzeile löst den Block | konform |
| Prüfung 3: `warn-unused: true` gesetzt, Warnung bricht nicht | awk setzt Befund bei Fehlen; Warnung kommt von golangci-lint | konform |
| Prüfung 4: mindestens eine `*.go` | `find … -print -quit` | konform |
| Randform 1: Profil fehlt, Exit 2 | `[ -r … ] \|\| exit 2` | konform (am Skript; über `make lint` nicht beobachtbar, F-5) |
| Randform 3: `rules: []` kein Befund | Gegenprobe oben | konform (F-1 behoben) |
| Zwei Teile laufen beide, Exit 1 bei jedem Befund | `rc=1` und `golangci-lint … \|\| rc=1` | konform |

- ADR-0003 / `AGENTS.md` §3.2: das Profil ist nicht gelockert (Diff berührt `.golangci.yml` nur im Kopfkommentar).
- Hinweis (F-6, offen, INFO): Einrückung des Why-Blocks nicht entschieden; kein Bruch einer Zusage.

### DoD 3 — `make lint` an `GATE_CHECKS`, Sensors-Zeile — erfüllt

- `GATE_CHECKS += lint-profil lint`; `make lint-profil` bleibt als Schema-Prüfung bestehen und läuft in `make gates`.
- `harness/README.md` §Sensors führt die Zeile `make lint` mit Bindung [ADR-0003](../plan/adr/0003-lint-profil-golangci.md); „Nicht behauptet" nennt die Gegenprobe zu `make lint` (`SPEC-102`/`slice-gegenproben`) und das Coverage-Gate.
- Sensors-Zeile nennt keine Prüfung, die es nicht gibt.

### DoD 4 — `make gates` grün — erfüllt

Selbst gefahren, Exit 0 (siehe Tabelle).

### Closure-Trigger §5

| Teil | Stand |
|---|---|
| `make gates` grün mit `make lint` in der Kette | erfüllt |
| Probe-Befund macht `make lint` rot, danach zurückgenommen | erfüllt (selbst gefahren) |
| DoD vollständig | offen: DoD 5–10 (Review-Abschluss, Doku-Update, Closure-Notiz, Register, Risiko-Ausgänge, Paarungen) stehen auf `[ ]`; DoD 1–4 sind im Plan noch nicht abgehakt |
| Closure-Notiz mit Lerneintrag | fehlt (§7 leer) |

## Hinweise für den Planner

1. **Behauptung ohne Beleg (§3.10):** Im Repo liegt kein Implementer-Bericht „Zusage · Mutation · roter Test" zu `lint.sh`; die Mutationen belegt nur der Review-Report. Die Gegenprobe-Fläche selbst ist `slice-gegenproben` zugewiesen und trägt sich dort; für die Closure sollte §7 den Beleg auf den Review verweisen.
2. **Plan-Nachzug (§3.9):** Der Fix-Commit `862961e` änderte `.dockerignore`, `lint.sh`, SPEC-101; §3 des Plans nennt diese Pfade bereits, §6-Risiken sind ohne Ausgang-Änderung korrekt. Kein Nachzug nötig. §6 „weiter offen" (Schwellen) und F-3/F-5/F-6 gehören mit Ausgang in die Closure-Notiz.

## Nicht geprüft

Bedarf gegen die Realität (Validator), Code-Qualität im Diff (Review), DoD 5–10.
