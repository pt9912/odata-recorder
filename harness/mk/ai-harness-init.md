# Gate-Index — Teil des Werkzeugs ai-harness-init

Diese Datei gehört dem Werkzeug `ai-harness-init`. Jeder Lauf des Werkzeugs (Bootstrap und
`add-lang`) schreibt sie kanonisch neu; eine Änderung von Hand geht beim nächsten Lauf verloren.
Was das Repo über ein Target dieser Datei entscheidet — einen Carveout, eine Sensor-Datei, eine
eigene Bindung —, steht in der Zeile von `harness/README.md` §Sensors, die diese Datei verlinkt.

**Was hier steht.** Je Make-Target der Make-Dateien, die das Werkzeug schreibt, eine Zeile:
`Makefile`, `d-check.mk`, `a-check.mk` (nur mit Arch-Gate) und alle Dateien unter
`harness/mk/`. Gate ist, was ein Fragment an `GATE_CHECKS` hängt, dazu `gates`; die
übrigen stehen in der zweiten Tabelle mit `kein Gate`. Ein Target, das `harness/README.md` beim
Lauf schon als Tabellenzeile führt, steht hier nicht. Die Spalte Vertrag bzw. Tut was trägt den
Hilfetext hinter `##` der Regel-Zeile.

**Wer das prüft.** Das Doku-Gate des Ziels (`.d-check.yml`, Modul `targets`) führt diese
Datei und `harness/README.md` gemeinsam als Autorität: ein Target ohne Zeile in einer der beiden
meldet `gate-undocumented`, eine Zeile ohne Target `gate-phantom`. Die eigenen Targets des Repos
stehen in `repo.mk` und brauchen ihre Zeile in `harness/README.md`.

**Grenzen.**

- Kein Target steht in zwei Teilen: so verlangt es die adoptierte Kurs-Fassung (Regelwerk
  `grundlagen-harness-dateien.md` §harness/README.md als Einstiegspunkt). Das Doku-Gate prüft die Disjunktheit, wenn in `.d-check.yml` das Modul `targets` in `modules` steht, `targets.authority` beide Teile als Liste nennt und `targets.authority-disjoint: true` gesetzt ist; dann meldet es ein Target, das in beiden Teilen eine Zeile hat, als `gate-declared-twice`. Sonst bleibt eine Doppelung still. Der Sensor sieht eine Doppelung erst, wenn das Repo sie angelegt hat: legt es nach einem Lauf eine eigene Zeile für ein Target dieser Datei an, steht es bis zum nächsten Lauf in beiden Teilen.
- Eine Datei, die das Repo selbst unter `harness/mk/` ablegt, liest der Lauf mit; ihre Targets
  stehen dann hier statt in `harness/README.md`.

## Gates

| Target | Vertrag | Bindung |
|---|---|---|
| `make baseline-verify` | Vendored Baseline netzlos verifizieren | — |

## Werkzeuge — kein Gate

| Target | Tut was | Bindung |
|---|---|---|
| `make archive-welle` | Zeitdokumente einer geschlossenen Welle archivieren (WELLE=<welle-id> \| WELLE=altbestand KENNUNG=<K>) — KEIN Gate | kein Gate |
| `make doc-commits` | Commit-Message-Traceability via Modul commits; RANGE=base..head | kein Gate |
| `make doc-complete` | Vollständigkeits-Gate: Requirements-Waise ⇒ Exit 1 | kein Gate |
| `make doc-doctor` | erklärende Diagnose mit Fix-Kandidaten | kein Gate |
| `make doc-help` | diese Liste der doc-*-Targets | kein Gate |
| `make doc-immutable` | Doc-/ADR-Immutabilität via git-Diff (Modul vcs); RANGE=base..head oder STAGED=1 | kein Gate |
| `make doc-planning` | Planning-Lifecycle-Konsistenz (Roadmap <-> in-progress) via Modul planning; hermetisch, ohne Range | kein Gate |
| `make doc-repair` | Reparatur-Patch (unified diff) auf stdout, git-apply-rein | kein Gate |
| `make doc-structure` | Struktur-Invarianten innerhalb der Dokumente via Modul structure; hermetisch, ohne Range | kein Gate |
| `make doc-targets` | Deklarations-Konsistenz Doku<->Build-Targets via Modul targets; hermetisch, ohne Range | kein Gate |
| `make doc-trace` | Requirements Traceability Matrix auf stdout (advisory) | kein Gate |
| `make doc-tracked` | Getrackt-Status aufloesbarer Referenz-Ziele via Modul tracked; braucht .git im Mount, ohne Range | kein Gate |
| `make doc-usage` | Aufruf und Optionen von d-check selbst (--help) | kein Gate |
| `make e2e-abdeckung` | die E2E-Abdeckungs-Sicht aus den Stufen-Deklarationen erzeugen — KEIN Gate | kein Gate |
| `make help` | Diese Hilfe | kein Gate |
| `make history-range-guard` | Vorlauf-Waechter: RANGE muss aufloesbar UND nicht leer sein (STAGED=1 prueft den Index; den STAGED-Zweig fuehrt nur doc-immutable) | kein Gate |
| `make hooks-install` | den git-eigenen commit-msg-Traeger im Klon aktivieren (core.hooksPath) — KEIN Gate | kein Gate |
| `make record-gates` | Gate-Nachweis stempeln (nur nach gruenen Checks) | kein Gate |
| `make selbstpruefung` | die emittierte Durchsetzungsschicht auf einem frischen Klon selbst pruefen — KEIN Gate | kein Gate |
| `make slice-mv` | Lifecycle-Wechsel eines Slice inkl. Verweis-Nachzug (SLICE=slice-<Kennung> TO=<open\|next\|in-progress\|done>) — KEIN Gate | kein Gate |
| `make span-clean` | Span-Bestand entfernen (ausdruecklich, kein Automatismus) — KEIN Gate | kein Gate |
| `make span-report` | Token-Bilanz je Rolle aus dem Span-Bestand — KEIN Gate (Bericht, kein Sensor) | kein Gate |
| `make traeger-fetch` | Traeger aus dem gepinnten Release nachholen (braucht Netz) — NICHT in gates | kein Gate |
