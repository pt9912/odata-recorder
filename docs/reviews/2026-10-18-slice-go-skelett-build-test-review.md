# Review: slice-go-skelett-build-test (Commit 703f174)

* Rolle: Reviewer (Modul 10), Skill `.harness/skills/reviewer.md`
* Gegenstand: `git show 703f174` gegen `docs/plan/planning/done/slice-go-skelett-build-test.md`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)/0002/0003, AGENTS.md §3, `spec/architecture.md` §2
* Bezug: `LH-QA-06`, `LH-RB-03`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
* Gefahrene Gates: `make gates` (grün), `make build` (grün; Binary gibt `odata-recorder dev` aus)
* Vorherige Findings am Modul: keine (docs/reviews/ trägt nur Spec-Straten-Reviews)

Ergebnis: 0 HIGH, 3 MEDIUM, 4 LOW, 2 INFO.

## Findings

### F1
- kategorie: MEDIUM
- quelle: AGENTS.md §3.12 (Randformen vor dem Code), `LH-QA-06`
- pfad: `docs/plan/planning/done/slice-go-skelett-build-test.md` · „## 6. Risiken und offene Punkte"; `harness/vertraege/SPEC-036-test.md`
- befund: §6 nennt keine Randformen von `Run`, und SPEC-036 entscheidet sie nicht. Das betrifft leere args, `version` mit Zusatzargument, unbekanntes Kommando, leeres Argument `[""]`, Flags wie `--help` und nil-Writer. Der Code entscheidet sie selbst (Exit 2, Meldung nur zu `args[0]`). Ein nil-`io.Writer` paniert in `fmt.Fprintf`; keine Stelle entscheidet oder testet das.
- verifizierbar: nein (kein Gate prüft, ob §6 Randformen nennt)
- klasse: Randform im Code statt vor dem Code entschieden

### F2
- kategorie: MEDIUM
- quelle: AGENTS.md §3.10 (Zusage ↔ Mutation), SPEC-036
- pfad: `harness/vertraege/SPEC-036-test.md` · „mit der Version aus dem Build-Argument `VERSION`" und „reproduzierbar (`-trimpath`, ohne VCS-Stempel)"
- befund: Der Vertrag sagt die Versionsübergabe per `VERSION` und die Reproduzierbarkeit zu. Kein Test und kein Gate fängt eine Mutation daran: `main.version` umbenennen oder `-X` entfernen (Binary meldet still `dev`), oder `-trimpath` streichen. Die Unit-Tests übergeben die Version direkt an `Run` und berühren `main.go` und das Dockerfile-`build` nicht. Das Commit-Body nennt Mutationen nur für Exit-Code, Ausgabe, Zusatzargument, gofmt und vet.
- verifizierbar: ja (Mutation fahren; `make gates` bleibt grün)
- klasse: Vertragszusage ohne fangende Mutation

### F3
- kategorie: MEDIUM
- quelle: SPEC-036, `LH-QA-06`
- pfad: `harness/mk/go-gates.mk` · „$(DOCKER_BUILD) --target build -t odata-recorder:build ."
- befund: `make build` reicht das Build-Argument `VERSION` nicht durch. Der Vertrag sagt „Version aus dem Build-Argument `VERSION`" zu, über `make` ist das aber nicht erreichbar. Beobachtet: das gebaute Binary meldet immer `odata-recorder dev`. Zusätzlich liegt das Binary nur im Image `odata-recorder:build`; kein Target liefert es aus.
- verifizierbar: ja (`make build`, Binary starten)
- klasse: Vertrag sagt mehr zu, als das Target erreichbar macht

### F4
- kategorie: LOW
- quelle: AGENTS.md §3.9 (Slice-Plan folgt der Korrektur)
- pfad: `docs/plan/planning/done/slice-go-skelett-build-test.md` · „`Run(args, stdout, stderr) int` gibt die Version aus"
- befund: §3 nennt die Signatur ohne Parameter `version`; der Code hat `Run(args, version, stdout, stderr)`. §3 beschreibt außerdem „Test der Version und der Fehlerausgabe (Happy, Negative)", der Test deckt aber auch den Zusatzargument-Fall ab. Die Plan-Zeile ist nicht nachgezogen.
- verifizierbar: nein
- klasse: Slice-Plan nicht mit Code nachgezogen

### F5
- kategorie: LOW
- quelle: AGENTS.md §3.11 (Kommentar sagt nur zu, was geprüft wird), §3.7
- pfad: `Dockerfile` · „deps: Module aus go.mod (und go.sum, sobald es Abhängigkeiten gibt), geprüft gegen go.sum"
- befund: Das Dockerfile kopiert in `deps` nur `COPY go.mod ./`. `go.sum` gelangt nicht in die Stufe, auch wenn es existiert; `go mod verify` prüft damit nicht gegen `go.sum`. Die Zusage im Kommentar hat im Code keinen Träger. Außerdem stehen „Eingangs-Images" im Plural bei einem einzigen `FROM`-Image.
- verifizierbar: ja (go.sum anlegen, `deps` ansehen)
- klasse: Kommentar-Zusage ohne Träger

### F6
- kategorie: LOW
- quelle: AGENTS.md §3.7 (Kommentar beschreibt, was da ist)
- pfad: `src/bootstrap/bootstrap.go` · „// Exit-Codes von Run."
- befund: Der Const-Block trägt neben den Exit-Codes auch `kommandoName`, der kein Exit-Code ist. Der Kommentar beschreibt den Block ungenau.
- verifizierbar: nein
- klasse: Kommentar ungenau

### F7
- kategorie: LOW
- quelle: AGENTS.md §3.9, Maintainability
- pfad: `docs/plan/planning/in-progress/roadmap.md` · Löschung von „Nichts in Arbeit."
- befund: Die Zeile ist ersatzlos entfernt und hinterlässt Leerzeilen an der Stelle. Die Roadmap nennt weiter nur den Link zu `welle-fundament`, der Slice als Inhalt der Welle fehlt. Die Änderung hängt am Slice, steht aber nicht in dessen §3.
- verifizierbar: nein
- klasse: Änderung außerhalb der Plan-Tabelle

### F8
- kategorie: INFO
- quelle: `.golangci.yml`, [ADR-0003](../plan/adr/0003-lint-profil-golangci.md)
- pfad: `src/bootstrap/main/main.go` · „var version = "dev""
- befund: `gochecknoglobals` nimmt `version` aus. `os.Std*` ist für `src/bootstrap/` per Ausnahme freigegeben. `testpackage` (Paket `bootstrap_test`) und `forbidigo` sind erfüllt. Ein Verstoß gegen das Profil ist nicht erkennbar; Lauf nicht verifizierbar, solange `make lint` kein Gate ist.
- verifizierbar: nein (Lint-Gate fehlt, Folge: slice-lint-stufe)
- klasse: Lint-Konformität nicht maschinell belegt

### F9
- kategorie: INFO
- quelle: `spec/architecture.md` §2, [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md)
- pfad: `src/bootstrap/main/main.go` · „import (\"os\" ..."
- befund: `src/bootstrap/main` liegt im Glob `src/bootstrap/**` und importiert nur `bootstrap`; kein Importgraph-Verstoß. Verifikation gegen die DoD ist Sache des Verifiers.
- verifizierbar: ja (`make a-check`)
- klasse: Importgraph konform

## Negativbefunde (geprüft, ohne Befund)

- `.dockerignore`: Allowlist `go.mod`, `go.sum`, `src/` stimmt mit §3 überein; geprüft, ohne Befund.
- `Dockerfile`: Digest-Pin auf dem einzigen `FROM`, `RUN --network=none` in `test` und `build`, Netz nur in `deps`, `-trimpath -buildvcs=false`, `GOTOOLCHAIN=local`; geprüft, ohne Befund (außer F5).
- `harness/mk/go-gates.mk`: `GATE_CHECKS += build test`, Kommentare tragen Kopplung und Grenze; geprüft, ohne Befund (außer F3).
- `harness/README.md` (Sensors-Zeilen, „Nicht behauptet"): Bindung [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), Zeilen stimmen mit den Targets überein; geprüft, ohne Befund.
- `harness/vertraege/README.md`: Index-Zeile nennt beide Targets; geprüft, ohne Befund.
- `go.mod`: keine Fremdabhängigkeit; geprüft, ohne Befund.
- `src/bootstrap/bootstrap_test.go`: Table-Test deckt Happy, unbekanntes Kommando und Zusatzargument; geprüft, ohne Befund außer F1 und F2.
- Hard Rules §3.1 (Docker-only: Build und Test nur über das Dockerfile), §3.2 (keine Suppression), §3.14 (Marke in §4 nennt die Metrik `git diff --numstat`): geprüft, ohne Befund.
- ADR-Verstöße ([ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md)): geprüft, ohne Befund. Geheimnis-, Persistenz- und Zustandsmodell-HIGH-Regeln: nicht berührt.
