# Slice slice-regeln-verkoerpern: Die Pflicht der Folgewellen und zwei Vorlagen der Closure von welle-mock in die Guide-Dateien

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — kein Produkt-Slice und keine Closure-Bedingung, die von der DoD verschieden wäre (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Linie C der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln): er ändert `AGENTS.md`, `.harness/skills/`, `.claude/commands/`, `.claude/agents/` und `.d-check.yml`, nicht die Gate-Kette; er ist trotzdem ein Kettenslice (Nr. 7), weil die Größengrenzen der Guide-Dateien ein Gate sind (`make docs-check`) und seine Verschiebung eine Änderung der Gate-Konfiguration ist.

**Bezug:** `LH-QA-06` (Reproduzierbarkeit der Prüfläufe; das Lastenheft führt keine Anforderung an die Regeln der Rollen, der Bezug ist der nächste, den es gibt), [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (Größe der Guide-Dateien als Gate; Anheben ist eine Lockerung, `AGENTS.md` §3.6), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Verträge der Werkzeuge, `SPEC-201`), Register-Einträge `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×), `BEO-REPO/report-der-rolle-faellt-docs-check` (5×), `BEO-REPO/menge-ohne-ihre-positionen` (6×), `BEO-REPO/verweis-auf-review-report-sperrt-die-archivierung` (3×), `BEO-REPO/ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (4×) und `BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last` (3×), alle über der Schwelle; die ersten beiden stehen in der Closure von `welle-mock` als Vorlage ohne `liegt in`, die übrigen vier in der Closure von `welle-hybrid` (*Entscheidungen beim User*).

**Berührte Spec-Stellen:** `SPEC-201` Nr. 11 (Pflicht, die dieser Slice in die Guide-Dateien trägt; der Vertrag bleibt unverändert, Grund in §3 Abschnitt B) · `SPEC-040` Nr. 8 b (Verschiebung der Größengrenzen, Messung aller Dateien) · `SPEC-040` Nr. 9 (Abschnitte von `AGENTS.md`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** pt9912 (Zweig slice-regeln-verkoerpern)

**Autor:** pt9912 (Planner der Closure von `welle-mock`). **Datum:** 2026-11-08. Entwurf des Architect: 2026-11-20 (§3 Abschnitte A bis H, Raster in §6). Größenmarke vom Planner geprüft: 2026-11-20 (§4).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Regeln, die die Closures von `welle-mock`, `welle-hybrid` und `slice-guard-schreibwege` als Vorlage oder Diff-Vorschlag ohne Träger hinterließen, stehen dort, wo jeder Lauf sie liest, und kein Gate wird gelockert. Der Umfang ist seit der Entscheidung des Users vom 2026-11-20 (Wortlaut unten):

1. die Pflicht der Folgewellen aus `SPEC-201` Nr. 11 (wer einen Fehlercode, einen Schlüssel oder ein Kommando erreichbar macht, ändert Ausnahmen und Handbuch im selben Slice) · 2. Vorlage 1 der Closure von `welle-mock` (Zweig ohne Lauf unter dem Detektor) · 3. Vorlage 2 (Report der Rolle besteht `make docs-check`) · 4. E1 (`menge-ohne-ihre-positionen`) · 5. E2 (`verweis-auf-review-report-sperrt-die-archivierung`) · 6. E4 (`ein-negativ-benennt-die-ungefangene-klasse-zu-eng` und `laufzeitmessung-widerspricht-sich-unter-last`) · 7. Rückgabe 2 aus §7 von [slice-guard-schreibwege](../done/slice-guard-schreibwege.md) (`Edit` im Typ `architect`) · 8. Rückgabe 3 (Satz zu den Scratchpad-Namen in den sechs Rollen-Typen).

Jeder Diff, der eine Guide-Datei verlängert, kürzt im selben Diff eine Datei desselben Maßes um mindestens denselben Betrag oder verschiebt die Grenze nach `SPEC-040` Nr. 8 b; die Summe der Grenzen steigt nicht.

**Entscheidung des Users (2026-11-20, Auswahl in der Sitzung, Wortlaut des Auftrags an den Architect):**

> 1. Der Slice darf `AGENTS.md`, `.claude/commands/`, `.claude/agents/` und `.harness/skills/` im Worktree ändern; der Implementer schreibt dort. Nr. 4 der Übergangsregeln ist damit bestätigt.
> 2. Zum Umfang gehören zusätzlich zu den drei Regeln aus welle-mock: E1 (menge-ohne-ihre-positionen, ein Satz in `AGENTS.md` §3.15), E2 (verweis-auf-review-report-sperrt-die-archivierung: Schritt 25 in `implement-slice.md` plus Reviewer-/Verifier-Anweisungen), E4 (ein-negativ-benennt-die-ungefangene-klasse-zu-eng und laufzeitmessung-widerspricht-sich-unter-last: `AGENTS.md` §3.15 und eine Regel nach §3.10) sowie die Rückgaben 2 und 3 aus slice-guard-schreibwege §7 (Edit im Architect-Typ; Scratchpad-Namen-Satz in den sechs `.claude/agents/*.md`).

Die beiden Rückgaben heißen in §7 von `slice-guard-schreibwege` unter *Entscheidungen beim User* Nr. 1 (Rückgabe 2) und Nr. 2 (Rückgabe 3); gemeint sind genau diese beiden Punkte. Dessen Nr. 3 und 4 (Nachzug von `SPEC-214`, Wartezeit des Hooks) sind nicht gewählt; Nr. 5 (Zuweisung von `ein-negativ-benennt-die-ungefangene-klasse-zu-eng`) erledigt E4.

**Übernimmt:** die Sendung aus der Closure von `welle-mock` (Geber `welle-mock`, Abschnitt *Steering-Loop-Einträge*: Vorlage 1, Vorlage 2 und die Pflicht aus `SPEC-201` Nr. 11, Entscheidung des Users vom 2026-11-08, Punkte lösen); die Punkte E1, E2 und E4 aus *Entscheidungen beim User* von `welle-hybrid-results` und die Rückgaben 2 und 3 aus §7 von `slice-guard-schreibwege` (Entscheidung des Users vom 2026-11-20).

**Wer legt was fest.** Der Architect hat vor dem Code festgelegt (§3 Abschnitte A bis H, Raster in §6): Ort und Wortlaut jeder Regel, die Kürzungen mit ihrem Träger, die Messung und die Verschiebung der Grenzen, das Raster, die Tabelle *Zusage · Fänger* und die erwartete Größe. Der Planner entscheidet nichts davon; der Implementer führt den Wortlaut aus und ändert ihn nicht (`AGENTS.md` §3.12).

**Ausgangsmaße** (am Start gemessen, Grenzen in `.d-check.yml`; sie stehen auf ihren Grenzen, außer der Zeilengrenze der Commands, die eine Zeile darüber steht): `AGENTS.md` 326 Zeilen und 23633 Bytes; `.claude/commands/implement-slice.md` 220 Zeilen und 17081 Bytes; `.claude/agents/architect.md` 53 Zeilen und 3759 Bytes; `.harness/skills/reviewer.md` 145 Zeilen und 8110 Bytes; `.claude/agents/verifier.md` 43 Zeilen und 2639 Bytes. Der Implementer misst vor dem ersten Edit neu.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Anhebung einer Grenze ohne gleichzeitige Senkung (Nr. 8 b von `SPEC-040`) — das wäre eine Lockerung des Gates und braucht eine ADR (`AGENTS.md` §3.6); wer die Verkörperung nicht mit Kürzung tragen kann, gibt den Punkt an den Planner zurück, statt die Grenze zu heben.
- Ein Sensor für die Pflicht aus `SPEC-201` Nr. 11 (Erreichbarkeit eines Codes gegen die Ausnahmen, `SPEC-201` Nr. 9 c) und ein Sensor für die Form des Report-Verweises aus E2 — ein Fänger am Binary oder am Werkzeug ist ein anderer Vorgang; die Verkörperung ist die Gegenmaßnahme, die `SPEC-201` selbst nennt, und E2 nennt den Sensor als Alternative (*Entscheidungen beim User* Nr. 2 von `welle-hybrid-results`), die der User nicht gewählt hat.
- Die übrigen Punkte der *Entscheidungen beim User* von `welle-hybrid-results` (E3, E5 bis E9) und die Nr. 3 und 4 aus §7 von `slice-guard-schreibwege` — der User hat sie nicht in den Umfang genommen; sie bleiben bei ihm.
- Die übrigen Beobachtungen über der Schwelle — der Umfang nennt die sechs oben.
- Die Größenformel und der Zuschlag (`SPEC-128` Nr. 1) — Gegenstand von `slice-groessenformel-neu`.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung.

- [x] Die acht Regeln des Umfangs (§1) stehen am Ort und im Wortlaut aus §3 (Abschnitt B), jede Regel aus einer Closure mit dem Anker `seit welle-mock` bzw. `seit welle-hybrid` (`grundlagen-traceability.md` §Herkunfts-Anker) und, soweit sie einen Zielort hat, dem Feld `liegt in` in der Notiz des Slice (§7; Zielorte in §3 Abschnitt B); die Pflicht aus `SPEC-201` Nr. 11 trägt den Zeiger auf ihren Vertrag.
- [x] Die Größengrenzen sind gehalten: jede berührte Guide-Datei liegt auf oder unter ihrer Grenze, keine Grenze ist ohne Senkung einer anderen im selben Maß gehoben (Messung vor und nach dem Diff in §7), `make docs-check` ist grün.
- [x] `make gates` grün (Verifier, Zweig, HEAD `47e1f32d`, Exit 0; der Lauf im Hauptbaum nach dem Merge ist der Integrationsschritt, §5).
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update: `SPEC-201` Nr. 11 bleibt unverändert (§3 Abschnitt B); das Register: die sechs Einträge des Bezugs bekommen den Ausgang *verkörpert* mit Zielort.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag (Einträge mit `liegt in <Zielort>`).
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen (nach dem Move geprüft: die Zielorte tragen `seit welle-mock` bzw. `seit welle-hybrid`, kein Folge-Slice genannt, alle sechs Kennungen des Registers liegen mit nicht leerem `evidence/`).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

Der Planner schneidet; der Architect entscheidet (Abschnitte A bis H unten), der Implementer erweitert diese Liste in seinem ersten Lauf. Die Zeilen sind Kandidaten.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.harness/skills/reviewer.md` | update | Vorlage 1, Prüfzeile zur Pflicht, Vorlage 2 mit E2; Kürzung K5; 145/8110 auf 151/8668 |
| `.claude/agents/verifier.md` | update | Vorlage 2 mit E2, Scratchpad-Satz; 43/2639 auf 48/3046 |
| `.claude/agents/architect.md` | update | `Edit`, Scratchpad-Satz, Rest der Vorlage `ANPASSEN` entfällt; 53/3759 auf 50/3614 |
| `.claude/agents/{implementer,reviewer,validator,planner}.md` | update | Scratchpad-Satz, je ein Satz hinter dem Absatz *Dein Kontext-Zuschnitt* |
| `.claude/commands/implement-slice.md` | update | Pflicht aus `SPEC-201` Nr. 11 (Schritt 17), E2 (Schritt 25); Kürzungen K3 und K4; 220/17081 auf 212/16180 |
| `AGENTS.md` | update | E1, E4 (§3.15), E4 (§3.10); Kürzungen K1 und K2; 326/23633 auf 323/23905 |
| `.d-check.yml` | update (Verschiebung nach `SPEC-040` Nr. 8 b) | Grenzen auf die neue Messung, Summe sinkt je Maß |
| `docs/plan/planning/observations/BEO-REPO/<sechs Einträge>/` | update (Closure, Planner) | `state.md`: Ausgang *verkörpert* mit Zielort |

`harness/vertraege/SPEC-201-handbuch-katalog.md` bleibt unverändert (Abschnitt B).

**Reihenfolge der Rollen nach dem Start**, jede in frischem Kontext: Architect (Texte, Orte, Kürzungen, Messung, Tabelle *Zusage · Fänger*, Raster) → Implementer → Reviewer → Fixes → Verifier → Planner (Closure). Eine Größenmarke führt der Slice nicht: der Architect hat es in Abschnitt G entschieden, der Planner hat es gegen `SPEC-128` geprüft und in §4 bestätigt.

**Betriebsregeln für alle Rollen:** `Edit` und `Write` auf Repo-Dateien (kein `sed -i`, kein Heredoc, kein `>>`); Commit-Messages per `git commit -F <datei>`; schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`; jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`; Zeichen jenseits von ASCII aus Bytewerten und auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

### A. Umfang und Rechte (Entscheidung des Users, 2026-11-20)

Der Wortlaut steht in §1. Der Implementer schreibt `AGENTS.md`, `.claude/commands/`, `.claude/agents/` und `.harness/skills/` in diesem Zweig; Nr. 4 der Übergangsregeln ist für diesen Slice bestätigt und wird nicht geändert (die Datei `docs/plan/planning/README.md` schreibt der Integrationsschritt). Keine ADR: nichts wird gelockert, die Verschiebung ist nach `SPEC-040` Nr. 8 b ohne ADR zulässig, [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) gilt unverändert.

### B. Orte und Wortlaut (Auftrag 1)

| Regel | Ort | Anker | `liegt in` (Zielort) |
|---|---|---|---|
| Pflicht aus `SPEC-201` Nr. 11 | `.claude/commands/implement-slice.md` Schritt 17 (Doku und Vertrag); Prüfung durch die Zeile *Code erreichbar gemacht, …* der MEDIUM-Liste des Reviewer-Skills | `seit welle-mock` | `.claude/commands/implement-slice.md` Schritt 17 und `.harness/skills/reviewer.md` |
| Vorlage 1 (Zweig ohne Lauf) | `.harness/skills/reviewer.md`, MEDIUM-Liste hinter *Fixture, Hilfe oder Kommentar verdeckt, was der Test prüft* | `seit welle-mock` | `.harness/skills/reviewer.md` |
| Vorlage 2 (Report besteht `make docs-check`) | `.harness/skills/reviewer.md` §Output-Schema und `.claude/agents/verifier.md` | `seit welle-mock` | beide Dateien |
| E1 | `AGENTS.md` §3.15, eine Zeile hinter *Ein Fänger führt jede Form der Eingabe …* | `seit welle-hybrid` (Überschrift) | `AGENTS.md` §3.15 |
| E2 | `implement-slice.md` Schritt 25; Skill §Output-Schema; `verifier.md` | `seit welle-hybrid` | die drei Dateien |
| E4, Negativ | `AGENTS.md` §3.15, der Satz *Jedes akzeptierte Negativ …* wird erweitert | `seit welle-hybrid` (Überschrift) | `AGENTS.md` §3.15 |
| E4, Laufzeit | `AGENTS.md` §3.10, eine Zeile hinter dem Satz *Eine Rate steht mit Last des Hosts …* | `seit welle-hybrid` (Überschrift) | `AGENTS.md` §3.10 |
| Rückgabe 2 | `.claude/agents/architect.md`, Zeile `tools:` | kein Anker (Werkzeugliste, keine Regel) | — |
| Rückgabe 3 | die sechs Typen unter `.claude/agents/`, Ende des Absatzes *Dein Kontext-Zuschnitt* | Zeiger *Übergangsregeln Nr. 8* als Herkunftsfeld | — |

**Entscheidungen zu Ort und Träger:**

- **Pflicht in `implement-slice.md`, nicht in `AGENTS.md`.** Sie bindet den Implementer an einer Stelle seines Ablaufs (Schritt 17: Doku und Vertrag), kostet dort eine Kürzung, die ohnehin entfällt (K3, K4), und braucht keinen Abschnitt `### 3.<n>` (höchstens 29 Zeilen). Eine Hard Rule hieße beim Reviewer HIGH (`ADR-Verstoß (Layer, Tool, Hard Rule)`); die Folge einer Auslassung ist ein MEDIUM.
- **Prüfer ist der Reviewer-Skill (MEDIUM).** Die Zeile trägt keinen Wortlaut der Pflicht, nur den Verstoß und den Zeiger auf `SPEC-201` Nr. 11; damit steht die Pflicht an **einem** Ort (Schritt 17) und die Prüfung an einem anderen mit anderem Inhalt, kein Drift der Pflicht.
- **`SPEC-201` Nr. 11 bleibt unverändert.** Der Vertrag trägt Wortlaut und Folge der Pflicht, nicht ihre Träger; ein Zeiger vom Vertrag auf Dateien unter `.claude/` wäre ein zweiter Ort, der beim Umnummerieren eines Schritts driftet. Die Zeiger laufen von den Guide-Dateien zum Vertrag (Richtung nach oben). `SPEC-201` Nr. 9 c (*ein Fänger fehlt*) bleibt wahr: gemeint ist ein Fänger am Binary; der Reviewer prüft von Hand.
- **E4, Laufzeit „nach §3.10":** der Satz steht in §3.10 hinter dem Satz zur Rate, den er verallgemeinert (Streuung unter Last); ein eigener Abschnitt `### 3.<n>` kostete vier Zeilen Kopf und Falsch/Richtig ohne Gewinn. §3.10 hat danach 21 von 29 Zeilen, §3.15 15 (Messung von `make docs-check`, Abschnitt I).
- **E4, Negativ:** Der Diff-Vorschlag aus dem Bericht der Closure von `slice-gegenproben-zeitbudget` liegt nicht im Repo (der Bericht ist nicht committet); der Wortlaut folgt aus dem Text der Beobachtung (`BEO-REPO/ein-negativ-benennt-die-ungefangene-klasse-zu-eng/observation.md`: *Klasse und Maß fehlen*) und der Lehre 1 in §7 von `slice-gegenproben-zeitbudget`. Der Haltepunkt am Argument vor einem Negativ steht schon in §3.10; er wird nicht wiederholt.

**Wortlaut je Datei.** Der Implementer übernimmt ihn wörtlich; die Zeilenumbrüche der Blöcke unten (rund 100 Zeichen, Einrückung der Liste) bleiben wie angegeben. Frei ist allein der Umbruch des Scratchpad-Satzes in den fünf Typen ohne Block (siehe `.claude/agents/`, Nr. 2 und 3); die Fänger der Tabelle F zählen dort ein Token, das kein Umbruch trennt.

`.harness/skills/reviewer.md`

1. Hinter dem Eintrag *Fixture, Hilfe oder Kommentar verdeckt, was der Test prüft* (letzte Zeile `mehr Elemente als der Test (`AGENTS.md` §3.11)  (seit welle-record-replay)`) zwei Einträge einfügen:

```
- **Zweig ohne Lauf unter dem Detektor** — ein Test sagt zu, `make test-race` melde jedes
  ungeordnete Paar, und die Menge der Zweige stammt aus der Erinnerung des Autors, nicht aus dem
  Abdeckungsprofil der Detektor-Tests oder aus einer Menge im Modell mit Zählprüfung; ein Zweig
  ohne Lauf bleibt ungemeldet (`AGENTS.md` §3.10)  (seit welle-mock)
- **Code erreichbar gemacht, Ausnahmen und Handbuch unverändert** — der Diff macht einen
  Fehlercode, einen Schlüssel oder ein Kommando erreichbar, ohne ihn aus
  `harness/katalog-ausnahmen.txt` zu entfernen und im Handbuch zu führen (`SPEC-201` Nr. 11)
  (seit welle-mock)
```

2. §Output-Schema, hinter dem Satz, der mit `neue Datei statt Überschreibung.` endet, drei Zeilen einfügen:

```
Eine `ADR-*`-Kennung steht im Report als auflösender Link, nie nackt; nach dem Schreiben läuft
`make docs-check`  (seit welle-mock). Den eigenen Report nennt er in Plan und Evidence mit
Kennung ohne Pfad und Endung  (seit welle-hybrid).
```

3. Entfernen (K5): der Block `> **Pflicht beim Ausfüllen (Modul 10 §Übungen):** …` bis `… keines deiner Repo-HIGHs zur Anwendung.` (vier Zeilen) samt der Leerzeile dahinter.

`.claude/commands/implement-slice.md`

1. Schritt 17 ersetzen (eine Zeile wird fünf):

```
17. Doku, ADR-Index und README aktualisieren, falls ein öffentlicher Vertrag berührt ist. Macht der
    Slice einen Fehlercode, einen Schlüssel oder ein Kommando erreichbar, entfernt er ihn im
    selben Slice aus `harness/katalog-ausnahmen.txt` und führt ihn im Handbuch; einen neuen Code in
    `fehler.go` setzt er in die Ausnahmen oder ins Handbuch (`SPEC-201` Nr. 11; kein Gate fängt die
    Erreichbarkeit)  (seit welle-mock).
```

2. Schritt 25: die Zeile `kein Freitext, eine Datei je Auftreten. Geschrieben wird er **vor** dem Move aus Schritt 24 — die` ersetzen durch:

```
    kein Freitext, eine Datei je Auftreten. Ein Review-Report steht in Evidence und Plan mit
    seiner Kennung ohne Pfad und ohne Endung (`2026-11-18-slice-x-review-2`), nie als Dateiname
    oder Link: `make archive-welle` sperrt mit `[haenger]`, solange eine Datei den Pfad nennt
    (seit welle-hybrid). Geschrieben wird der Beleg **vor** dem Move aus Schritt 24 — die
```

3. K3: der Punkt `- **Gate-Nachweis + Stop-Hook.** …` (zehn Zeilen, bis `ist.` hinter `STOP_GATE_STRENG=1 gesetzt`) wird ersetzt durch:

```
- **Gate-Nachweis + Stop-Hook.** `make gates` endet mit `record-gates` (Content-Hash des Working
  Tree, Commit-SHA von HEAD). Der Stop-Hook verweigert das Turn-Ende, sobald HEAD ein anderer ist
  als beim letzten grünen Lauf **und** der Inhalt nicht gedeckt ist — **ein Commit oder
  `git mv`-Commit ohne frischen `make gates`-Lauf lässt ihn rot.** Wann er streng blockiert, steht
  im Kopf von `.claude/hooks/stop-require-gates.sh`.
```

4. K4: der Punkt `- **Commit-Kennung.** …` (fünfzehn Zeilen, bis `… fällt sie.`) wird ersetzt durch:

```
- **Commit-Kennung.** Eine Commit-Message ohne Kennung (`ADR-NNNN`, `LH-XX-NN`, `MR-NNN`; die
  Menge steht in der Zeile `patterns=` von `tools/harness/commit-msg-traceability.sh`) weist
  `.githooks/commit-msg` ab, auch bei `make slice-mv` und `make archive-welle`; `make hooks-install`
  aktiviert ihn, `git commit --no-verify` umgeht ihn. Geprüft wird die Anwesenheit einer Kennung,
  nicht ihre Wahrheit; das Doku-Update bei berührtem öffentlichem Vertrag bleibt deine Arbeit.
```

`AGENTS.md`

1. K1: in §1 die fünf Zeilen von `**Modul-Bundle** (`README.md` = Index). Beim Bootstrap wird das` bis `(Quelle/Stand in [`harness/conventions.md`](harness/conventions.md) §Baseline).` werden:

```
**Modul-Bundle** (`README.md` = Index). Es liegt **vendored** unter
`.harness/baseline/<tag>/{regelwerk,templates}/` (Quelle und Stand:
[`harness/conventions.md`](harness/conventions.md) §Baseline).
```

2. K2: in §1 der Absatz `Die **Skelett-Vorlagen** der Baseline liegen **vendored** unter` bis `statt frei zu formulieren.` (sechs Zeilen) wird:

```
Die **Skelett-Vorlagen** der Baseline liegen vendored unter `.harness/baseline/<tag>/templates/`;
ein neues Artefakt (ADR, Slice, Welle, Carveout, Review-Report) wird aus ihnen **kopiert und
ausgefüllt**, nicht frei formuliert.
```

3. §3.10: die Überschrift endet auf `(geschärft seit welle-record-replay und seit welle-hybrid)`; hinter dem Satz `Eine Rate steht mit Last des Hosts und Messendem; … nicht als fester Wert.` eine Zeile:

```
Eine Laufzeit steht als Gewinn oder Verlust nur mit mehreren Läufen je Seite und der Last des Hosts vor und nach jedem Lauf, und nur, wenn der Unterschied die Streuung übersteigt; ein Lauf je Seite beweist keinen von beiden.
```

4. §3.15: die Überschrift endet auf `(seit welle-fundament, geschärft seit welle-record-replay und seit welle-hybrid)`; hinter dem Satz `Ein Fänger führt jede Form der Eingabe, … gleich sind.` eine Zeile:

```
Eine Menge im Plan nennt ihre Achsen (Stelle des Elements im Wert, Abstand eines Paares, Modus, Weg zum Aufruf); der Fänger führt jede Achse, oder das Negativ nennt die ungeführte.
```

und der Satz `Jedes akzeptierte Negativ nennt Grund und Folge, sonst ist es eine Zusage ohne Fänger.` lautet:

```
Jedes akzeptierte Negativ nennt Grund und Folge und die ungefangene Klasse mit ihrem Maß (n von m), nicht eine Aufzählung oder ein Fenster als Zahl, das die nächste Messung überholt; sonst ist es eine Zusage ohne Fänger.
```

`.claude/agents/`

1. `architect.md`: Zeile `tools: Read, Write, Bash` wird `tools: Read, Write, Edit, Bash`; der Absatz *Dein Kontext-Zuschnitt* endet auf `Entscheidungslage, **bevor** Code existiert.`, die Zeile wird fortgesetzt:

```
Entscheidungslage, **bevor** Code existiert. Im gemeinsamen Scratchpad legst du Dateien und
Verzeichnisse nur unter `<kurz>-<rolle>/` an und löschst nur dort (Übergangsregeln Nr. 8).
```

   K6: die Leerzeile vor `<!-- ANPASSEN: …` und die drei Zeilen des Kommentars entfallen (vier Zeilen).
2. `implementer.md`, `reviewer.md`, `validator.md`, `planner.md`: derselbe Satz, angehängt an die letzte Zeile des Absatzes *Dein Kontext-Zuschnitt* und auf zwei Zeilen umbrochen (`Im gemeinsamen Scratchpad legst du Dateien und` / `Verzeichnisse nur unter `<kurz>-<rolle>/` an und löschst nur dort (Übergangsregeln Nr. 8).`). Der Umbruch richtet sich nach der Zeilenlänge der Datei und ist frei, auch zwischen `Übergangsregeln` und `Nr. 8`; die Zeilenzahl wächst um eine, höchstens um zwei. Fest ist das Token `<kurz>-<rolle>/` (ein Wort ohne Leerzeichen, der Umbruch trennt es nie); der Fänger der Tabelle F, Zeile 10, zählt dieses Token.
3. `verifier.md`: der Scratchpad-Satz hängt an `und für das Review.` (Ende des Absatzes *Dein Kontext-Zuschnitt — und die Falle, für die es dich gibt.*) und ist wie in Nr. 2 umbrochen. Hinter dem Absatz *Einen Fix löst du gegen die Klasse auf …* (endet `(seit welle-record-replay).`) eine Leerzeile und:

```
**Dein Report besteht `make docs-check`:** eine `ADR-*`-Kennung steht als auflösender Link, nie
nackt, und der Lauf folgt dem Schreiben des Reports  (seit welle-mock). Den Report nennst du in
Plan und Evidence mit Kennung ohne Pfad und Endung  (seit welle-hybrid).
```

**Wortlaut ist Eingabe, kein Spielraum:** wo eine Zeile beim Einfügen die Umbruchbreite der Datei sprengt, bleibt der Wortlaut, nicht der Umbruch. Weicht der Implementer inhaltlich ab, gibt er den Punkt an den Architect zurück (§3.12).

### C. Kürzungen (Auftrag 2)

Jede Kürzung nennt die Regel, deren Text entfällt, und wo sie weiter steht. Keine der Streichungen trug einen Anker `seit …`, keine wird von einer anderen Datei des Repos zitiert (`grep` über `*.md`, `*.yml`, `*.sh` am 2026-11-20, ohne die vendorte Baseline und Zeitdokumente: kein Treffer auf die Überschriften der gekürzten Stellen).

| Nr. | Datei, Text | Was entfällt | Wo es weiter steht (Träger) |
|---|---|---|---|
| K1 | `AGENTS.md` §1, Satz zum Release-ZIP | Adresse des ZIP und der Satz *Beim Bootstrap wird … committet vendored* | `harness/conventions.md` §Adoptierte Konventions-Quellen führt dieselbe Adresse und den Pfad `.harness/baseline/<tag>/…`; §1 behält den Zeiger auf §Baseline |
| K2 | `AGENTS.md` §1, Absatz *Skelett-Vorlagen* | die Doppelrolle *Referenz-Form … Ziel-Form* | `harness/README.md` §Guides, Zeile der Vorlagen (derselbe Satz); *kopiert und ausgefüllt* bleibt im Absatz und in `implement-slice.md` (*Neue Artefakte per `cp`*) |
| K3 | `implement-slice.md`, Punkt *Gate-Nachweis + Stop-Hook* | die Zweige des Hooks (Turn-Ende ohne Commit, fehlender Stempel, streng über `.harness/stop-gate-streng` und `STOP_GATE_STRENG=1`) | Kopf von `.claude/hooks/stop-require-gates.sh` (DEFAULT, STRENG, GRENZEN, in der Sache wortgleich); der Punkt behält die Zusage *Commit ohne frischen `make gates`-Lauf lässt ihn rot* und den Zeiger |
| K4 | `implement-slice.md`, Punkt *Commit-Kennung* | die Beschreibung des Bootstraps (*An diesem Pfad ist das Werkzeug ein Gast*, skip-if-present, *reist mit dem Klon*) | Kopf von `.githooks/commit-msg`, `harness/mk/hooks-install.mk` und `tools/harness/commit-msg-traceability.sh`; die Zusagen *Anwesenheit, nicht Wahrheit*, *auch Commits der Repo-Werkzeuge*, *`--no-verify` umgeht*, *Doku-Update bleibt deine Arbeit* bleiben im Punkt |
| K5 | `reviewer.md`, Block *Pflicht beim Ausfüllen* | die Anweisung an den Autor des Skills, die HIGH-Liste mit mindestens zwei Repo-Regeln zu füllen | die Vorlage `reviewer.template.md` der vendorten Baseline führt sie wörtlich; die HIGH-Liste trägt mehr als zwei Repo-Regeln (Spec-Straten, Geheimnis, Zustandsmodell, Zustandsfeld, Kommentar); eine Laufregel ist es nicht |
| K6 | `architect.md`, Rest der Vorlage `ANPASSEN` | der Kommentar, der zum Eintragen von Quellen auffordert | der Absatz *Warum dieser Typ existiert — und woran er hängt* sagt dasselbe über den Namen im Kopf; die fünf anderen Typen behalten den Kommentar (N5) |

Keine Kürzung schwächt ein Gate; die Zusagen, die gekürzte Texte nannten, haben in den Spalten oben ihren Träger (`AGENTS.md` §3.11).

### D. Messung und Verschiebung (Auftrag 3)

Entwurf im Scratchpad gemessen (Zeilen wie `wc -l`, Bytes wie `wc -c`); der Implementer misst nach seinem Edit neu und setzt die gemessenen Werte.

| Datei / Regel | vorher (Zeilen / Bytes) | nach dem Diff (Entwurf) | Grenze vorher | Grenze nachher | Änderung der Grenze |
|---|---|---|---|---|---|
| `AGENTS.md` | 326 / 23633 | 323 / 23905 | 326 / 23633 | 323 / 23905 | −3 Zeilen, +272 Bytes |
| `.claude/commands/*.md` (Maximum `implement-slice.md`) | 220 / 17081 | 212 / 16180 | 221 / 17081 | 212 / 16180 | −9 Zeilen, −901 Bytes |
| `.claude/agents/*.md` (Maximum `architect.md`) | 53 / 3759 | 50 / 3614 | 53 / 3759 | 50 / 3614 | −3 Zeilen, −145 Bytes |
| `.harness/skills/*.md` (Maximum `reviewer.md`) | 145 / 8110 | 151 / 8668 | 145 / 8110 | 151 / 8668 | +6 Zeilen, +558 Bytes |
| `harness/conventions.md` | 228 / 16065 | unverändert | 228 / 16065 | 228 / 16065 | 0 |
| **Summe der Grenzen** | | | 973 / 68648 | 964 / 68432 | −9 Zeilen, −216 Bytes |

Weitere Dateien nach dem Diff (gemessen): `verifier.md` 48 / 3046, `implementer.md` 49 / 3209, `reviewer.md` (Typ) 42 / 2682, `validator.md` 43 / 2633, `planner.md` 38 / 2306; alle unter der neuen Grenze der Agenten (50 / 3614). Die Abschnitte `### 3.<n>` von `AGENTS.md` bleiben unter 29 Zeilen (§3.7 mit 29 ist das Maximum und unverändert).

**Verschiebung.** Ein Commit trägt die Edits der fünf Guide-Gruppen und `.d-check.yml` (`SPEC-040` Nr. 8 b: die Hebung `AGENTS.md` Bytes und `.harness/skills/*.md` Zeilen und Bytes steht im selben Commit wie die Senkung der Commands und der Agenten). Die Summe sinkt je Maß, also keine Lockerung. Der Kommentar von `.d-check.yml` nennt Messung vorher und nachher je Datei und den Grund; derselbe Text steht in der Commit-Nachricht. Der Kommentar nannte bisher die Summe 976 Zeilen und 68663 Bytes; die Blöcke ergeben 973 und 68648. Der Implementer trägt die gemessene Summe ein; den früheren Wert nennt der Kommentar nicht (`AGENTS.md` §3.7), die Commit-Nachricht und dieser Plan halten ihn fest.

**Die Gegenproben:** `.d-check.yml` steht in `EINGABEN` von `tools/harness/gegenproben-eingabe.sh`; die Prüfsumme ändert sich, der Stempel `harness/gegenproben.stempel` wird in diesem Zweig nicht committet, der Integrationsschritt fährt `make gegenproben` nach dem Merge (Nr. 2 der Übergangsregeln). Die Fälle 156 bis 170 und 213 lesen die Grenzen der echten Datei, aber mit kleinen Dateien; die neuen Grenzen ändern ihren Ausgang nicht.

### E. Randformen

Das Raster *Operation × Argumentart* steht in §6 (`AGENTS.md` §3.16), samt den drei Randformen aus Auftrag 4 (Regel in zwei Dateien, gelöschter Anker, Verschiebung und Kürzung im selben Commit).

### F. Tabelle *Zusage · Fänger* (Auftrag 5; `AGENTS.md` §3.15)

Dieser Slice hat keinen Go-Code und keine Nebenläufigkeit; alle Zeilen sind `andere`. Fänger sind `make docs-check`, `make zeichen-check` und Handproben des Verifiers (Zählung mit `grep -c`, `test -e`, `wc`); wo kein Gate fängt, steht das Negativ (N1 bis N6 unten). Die Anker-Zählung ist eine Schleife über Dateien und Wellen; Soll je Datei (`welle-mock` / `welle-hybrid`): `AGENTS.md` 0 / 2, `implement-slice.md` 1 / 1, `reviewer.md` (Skill) 3 / 1, `verifier.md` 1 / 1, alle anderen Guide-Dateien 0 / 0.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | Die Pflicht aus `SPEC-201` Nr. 11 steht in Schritt 17 von `implement-slice.md`, nennt den Vertrag und trägt `seit welle-mock` | Anker-Zählung der Schleife (Datei `implement-slice.md`, Welle `welle-mock`, Soll 1); Handprobe, N1 | Satz in Schritt 17 gelöscht: Zählung 0 | andere |
| 2 | Die Prüfzeile *Code erreichbar gemacht, Ausnahmen und Handbuch unverändert* steht in der MEDIUM-Liste des Skills, ohne den Wortlaut der Pflicht | Anker-Zählung (Skill, `welle-mock`, Soll 3); `grep -c 'SPEC-201'` im Skill gleich 1; N1 | Zeile gelöscht: Zählung 2 | andere |
| 3 | Vorlage 1 (*Zweig ohne Lauf unter dem Detektor*) steht in der MEDIUM-Liste hinter *Fixture, Hilfe oder Kommentar verdeckt, was der Test prüft* | Anker-Zählung (Skill, `welle-mock`) und `grep -n`: die Zeile des Eintrags liegt hinter der des Fixture-Eintrags und vor der Zeile `**LOW**` (beide Grenzen); N1, N9 | Eintrag in den Abschnitt LOW verschoben: Reihenfolge falsch, Zählung bleibt 3; Eintrag vor den Fixture-Eintrag: untere Grenze rot | andere |
| 4 | Vorlage 2 steht in zwei Dateien: im Skill §Output-Schema und in `verifier.md` | Schleife über beide Dateien: `grep -c 'nach dem Schreiben'` im Skill und `grep -c 'Dein Report besteht'` in `verifier.md`, je 1 (beide Wortfolgen stehen im Wortlaut von Abschnitt B in einer Zeile); Anker-Zählung; N1 | Satz nur in `verifier.md` gelöscht (nicht im Skill, nicht im Befund genannt): Zählung dort 0 | andere |
| 5 | E1: der Satz zu den Achsen einer Menge steht in `AGENTS.md` §3.15 vor *Eine Fixture gleicht nie dem Wert der Mutation.* | `grep -n` (Zeile des Satzes liegt zwischen den beiden genannten Sätzen); Anker-Zählung `AGENTS.md`/`welle-hybrid` gleich 2; `make docs-check` für den Abschnitt (15 von 29 Zeilen); N1 | Satz gelöscht: Zählung des Anfangs `Eine Menge im Plan nennt ihre Achsen` 0 | andere |
| 6 | E2 steht in drei Dateien mit je eigenem Wortlaut: `implement-slice.md` Schritt 25, Skill §Output-Schema, `verifier.md` | Schleife über die drei Dateien: `grep -c 'ohne Pfad und'` je 1 (die Wortfolge steht im Wortlaut von Abschnitt B in einer Zeile) und Anker-Zählung `welle-hybrid`; N1 | Satz nur im Skill gelöscht: Zählung dort 0 | andere |
| 7 | E4, Negativ: der Satz zu Grund, Folge und ungefangener Klasse mit Maß steht in §3.15 | `grep -c 'ungefangene Klasse mit ihrem Maß'` in `AGENTS.md` gleich 1; Länge von §3.15 von `make docs-check` gedeckt | Rückfall auf die alte Fassung (`nennt Grund und Folge, sonst`): Zählung 0 | andere |
| 8 | E4, Laufzeit: der Satz steht in §3.10 hinter dem Satz zur Rate | `grep -n`: die Zeile des Satzes ist die Zeile von `Eine Rate steht mit Last` plus 1 (unmittelbar); §3.10 hat höchstens 29 Zeilen (`make docs-check`; Ist 21); N9 | Satz ans Ende des Abschnitts verschoben: Zeile 189 statt 182, rot; Satz um 9 Zeilen gestreckt (Gegenprobe des Gates für die Länge; 8 Zeilen ergäben genau 29 und blieben grün): `section-lines-exceeded` | andere |
| 9 | Rückgabe 2: `architect.md` trägt `Edit` in `tools:` | `grep -c '^tools: Read, Write, Edit, Bash$'` gleich 1; Handprobe, N1 | `Edit` gestrichen: Zählung 0 | andere |
| 10 | Rückgabe 3: der Scratchpad-Satz steht in allen sechs Typen | Schleife über `.claude/agents/*.md`: `grep -c '<kurz>-<rolle>/'` je Datei 1, Summe 6 (das Token steht im Satz an einer Stelle und steht in keiner Datei vor der Änderung; der Umbruch trennt es nie) | Satz in `planner.md` gelöscht (nicht im Befund genannt): Summe 5 | andere |
| 11 | Jede berührte Guide-Datei liegt auf oder unter ihrer Grenze (Zeilen und Bytes) | `make docs-check` (`file-lines-exceeded`, `file-bytes-exceeded`, `SPEC-040` Nr. 5) | je Regel (`AGENTS.md`, Commands, Agenten, Skills) eine Zeile angehängt: jeweils rot, 4 von 4 | andere |
| 12 | Kein Abschnitt `### 3.<n>` von `AGENTS.md` trägt mehr als 29 Zeilen | `make docs-check` (`section-lines-exceeded`, Nr. 9) | in §3.10 14 Zeilen und in §3.15 15 Zeilen angehängt (Ist 21 und 15 von 29; 14 Zeilen in §3.15 ergäben genau 29 und blieben grün): zwei Befunde, 2 von 2 | andere |
| 13 | Die Summe der Grenzen steigt nicht, jede Grenze steht auf dem Ist | `make docs-check` fängt nur eine Grenze unter dem Ist (Nr. 8 a); Summe und Gleichheit rechnen Reviewer und Verifier mit `wc` und dem Kommentar von `.d-check.yml` (N4) | Grenze um 1 unter das Ist gesetzt: `file-bytes-exceeded`; eine Grenze über dem Ist bleibt grün (Negativ) | andere |
| 14 | Pfade und `ADR-*`-Kennungen der neuen Texte lösen auf (`harness/katalog-ausnahmen.txt`, `ADR-*` als Muster); der Zeiger auf `SPEC-201` nur per Handprobe (N8) | `make docs-check` (`codepaths`, `ids`) für `AGENTS.md`, Commands, Agenten; im Skill Handprobe `test -e harness/katalog-ausnahmen.txt` (N3) | Pfad in `implement-slice.md` um einen Buchstaben geändert: `codepath-missing` | andere |
| 15 | Keine Escape-Folge und kein unsichtbares Zeichen in den geschriebenen Dateien | `make zeichen-check` (`SPEC-233`) und Gegenlesen mit `LC_ALL=C grep -P` auf die Folge Backslash, `u`, vier Hex-Ziffern, BOM, NBSP, Nullbreite (`AGENTS.md` §3.18); der Bericht nennt die gegengelesenen Dateien | ein NBSP in `reviewer.md` eingefügt: `zeichen-check` rot | andere |
| 16 | Die Kürzungen K1 bis K6 nehmen keine Zusage mit, deren Träger entfällt | Handprobe je Zeile von Abschnitt C: der Träger der Spalte *Wo es weiter steht* trägt den Satz noch (`grep` im Kopf von `stop-require-gates.sh`, `commit-msg`, `conventions.md`, `harness/README.md`, `reviewer.template.md`) | K3 ohne die Zusage *Commit ohne frischen Lauf lässt ihn rot*: die Handprobe `grep -c 'lässt ihn rot'` in `implement-slice.md` gleich 0 | andere |

Der Implementer fährt jede Mutation einmal selbst und nennt je Zeile Mutation und roten Befund (`AGENTS.md` §3.10); die Mutationen der Zeilen 1 bis 10 und 16 sind Handproben (der Zähler der Schleife ändert sich), nicht Läufe eines Gates, die Zeilen 11, 12, 14 und 15 sind Läufe. Der Implementer trägt die Tabelle und die Ergebnisse in einen Absatz von §3 ein.

**Akzeptierte Negative** (`AGENTS.md` §3.15: Grund, Folge, ungefangene Klasse mit Maß):

- **N1 — das Löschen eines Satzes in einer Guide-Datei fängt kein Gate.** Grund: `make docs-check` misst Größe, Verweise und Richtung, nicht Inhalt (`SPEC-040` *Grenze*); ein gelöschter Satz verkürzt die Datei und bleibt grün. Folge: eine spätere Kürzung kann eine dieser Regeln still entfernen; Gegenmaßnahmen sind der Retirement-Check am Anker `seit welle-…` (`grundlagen-traceability.md`, Sensor 2) und der Reviewer am Diff. Klasse und Maß: Löschung eines der zehn Sätze der Zeilen 1 bis 10 der Tabelle, die in neun Dateien stehen (`AGENTS.md`, `implement-slice.md`, der Skill und die sechs Typen unter `.claude/agents/`; `verifier.md` und `architect.md` tragen je zwei der Sätze); der Implementer löscht je Zeile einen Satz in einer Kopie und nennt, ob `make docs-check` grün bleibt: erwartet 10 von 10 Löschungen grün (jede verkürzt eine Datei und verletzt keine Grenze).
- **N2 — ob eine Rolle die Regel befolgt, fängt kein Test.** Grund: Prosa in einer Guide-Datei wirkt über den Lauf der Rolle. Folge: die Regel kann wirkungslos bleiben; das Register zählt das Wiederauftreten (Beispiel: `fix-deckt-die-genannte-mutation-nicht-die-klasse`, die Regel §3.17 trug in `slice-bootstrap-m4-konfig` vier Runden nicht); der Retirement-Check fragt bei jeder Lockerung nach.
- **N3 — Pfade und Kennungen im Skill prüft kein Gate.** Grund: `.harness/**` steht in `scan.ignore` von `.d-check.yml`; das Modul `file` erreicht die Datei, `links`, `ids` und `codepaths` nicht. Folge: ein toter Pfad im Skill bleibt grün; die Handprobe `test -e` des Verifiers fängt den Pfad der Prüfzeile (Zeile 14). Menge: ein Pfad (`harness/katalog-ausnahmen.txt`) und ein Zeiger (`SPEC-201` Nr. 11), 1 von 2 ohne Handprobe.
- **N4 — eine Grenze über dem Ist und eine Hebung ohne Senkung fängt kein Gate** (`SPEC-040` Nr. 10 a, 8 e). Grund: Gleich dem Ist als Prüfung zwänge das Senken in jeden Kürzungs-Commit. Folge: eine zu hoch gesetzte Grenze passiert; Reviewer und Verifier rechnen die Summe und die Gleichheit mit dem Ist nach (Abschnitt D, Zeile *Summe*).
- **N5 — fünf Typen behalten den Kommentar `ANPASSEN`.** Grund: nur `architect.md` steht auf der Grenze und braucht den Platz (K6). Folge: die Typen sind uneinheitlich; kein Lauf liest den Kommentar als Regel.
- **N6 — der Zeiger *Übergangsregeln Nr. 8* ist ohne Link.** Grund: die Typen verweisen nie mit Pfad in `docs/plan/planning/`; der Zeiger steht als Text. Folge: baut der Planner die Übergangsregeln um (Ablösung durch `slice-worktree-parallel`, *Nr. 3 bis 8 bleiben*), verweist der Satz auf eine Nummer, die es nicht mehr gibt; der Auslöser ist der Umbau der Übergangsregeln, die Adresse der Planner.
- **N7 — die Fänger der Tabelle F zählen ein Fragment der Zusage, nicht ihren ganzen Satz** (Befund F-3 des Reviews `2026-11-20-slice-regeln-verkoerpern-review`). Grund: ein Gate liest den Inhalt eines Satzes in einer Guide-Datei nicht (N1); die Handproben zählen je Zusage eine Wortfolge (`nach dem Schreiben`, `ohne Pfad und`, `ungefangene Klasse mit ihrem Maß`, `<kurz>-<rolle>/`, `lässt ihn rot`, `SPEC-201`). Folge: eine Umformulierung oder ein Teilverlust im selben Satz (ein Halbsatz, eine Achse der Liste, `(n von m)`, `und Endung`, der Wortlaut der Pflicht in der Prüfzeile) lässt die Zählung grün; Gegenmaßnahmen sind der Reviewer am Diff und der Retirement-Check am Anker. Klasse und Maß: Teilmutation eines zugesagten Satzes, 10 von 16 Zeilen der Tabelle (1, 2, 3, 4, 5, 6, 7, 8, 10, 16) lassen mindestens eine Teilmutation grün (acht grüne Mutationen des Reviews an den Zeilen 2, 4, 5, 6, 7 (zwei), 10 und 16; die Verifikation fand drei weitere an den Zeilen 1, 3 und 8: Halbsatz mit `harness/katalog-ausnahmen.txt` in Schritt 17, `oder aus einer Menge im Modell mit Zählprüfung` in Vorlage 1, `und nur, wenn der Unterschied die Streuung übersteigt` im Laufzeit-Satz; Berichtigung V-1 der Verifikation `2026-11-20-slice-regeln-verkoerpern-verifikation`).
- **N8 — der Zeiger auf `SPEC-201` Nr. 11 prüft kein Gate** (Befund F-4). Grund: `ids` verlangt Links nur für `LH-`, `ADR-` und `MR-`, nicht für `SPEC-`; `codepaths` prüft Pfade, nicht Kennungen. Folge: ein toter Zeiger (`SPEC-999`) in Schritt 17 oder in der Prüfzeile bleibt grün; der Zeiger ist die Zusage *die Pflicht trägt den Zeiger auf ihren Vertrag* (§2). Klasse und Maß: Zeiger auf `SPEC-201` in den Guide-Dateien, 2 von 2 ohne Gate (Schritt 17 und Prüfzeile; die Prüfzeile liegt zudem unter N3); Handprobe `grep -c 'SPEC-201. Nr. 11'` in Schritt 17 gleich 1 und `test -e harness/vertraege/SPEC-201-handbuch-katalog.md`.
- **N9 — die Fänger der Ortszusagen (Zeilen 3, 5, 8) sind Handproben ohne Gate** (Befund F-5). Grund: kein Gate liest die Reihenfolge von Sätzen oder Einträgen; die Probe ist `grep -n` mit den Grenzen der Zeile (3: hinter dem Fixture-Eintrag und vor `**LOW**`; 5: zwischen den beiden genannten Sätzen; 8: unmittelbar hinter der Rate, Zeile plus 1). Folge: wer die Probe nur als *nach* liest, lässt eine Verschiebung an eine andere Stelle des Abschnitts grün (der Eintrag der Zeile 3 an einer anderen Stelle der MEDIUM-Liste bleibt grün, solange er hinter dem Fixture-Eintrag und vor `**LOW**` steht). Klasse und Maß: Ortszusage ohne Gate, 3 von 3 (Zeilen 3, 5, 8); die Grenzen stehen in der Tabelle, Reviewer und Verifier fahren die Probe.

### G. Größe und Marke (Auftrag 5)

**Entscheidung nach `SPEC-128`: dieser Slice führt keinen Block `groessenmarke` und wird nicht gemessen** (`SPEC-128` Nr. 3 p: *kein Block*, Hinweis und Exit 0; Nr. 10 d). Gründe: (1) Es gibt keinen Go-Code; die Faktoren der Formel (80 je andere Zeile) stammen aus drei Go-Slices und überschätzen einen Diff aus Guide-Dateien um das Zehnfache; eine Auslastung weit unter 0,60 löste nach Nr. 9 eine Neurechnung der Faktoren aus einem Slice aus, der kein Messpunkt der Faktoren ist. (2) Das Wachstum der Guide-Dateien begrenzt schon `make docs-check` je Datei und Maß (`SPEC-040`); die Marke fügte keine Grenze hinzu. Folge im Betrieb: ein Hinweis *kein Block* steht in der Ausgabe von `make groessenmarke`; die Rolle, die übergibt oder prüft, nennt den Lauf in ihrem Bericht (`SPEC-128` Nr. 10 d).

**Erwartung der Größe** (für den Planner und den Verifier): der Diff über die Guide-Dateien und `.d-check.yml` liegt bei 77 Einfügungen und 71 Löschungen (`git diff --numstat` der Entwürfe: `AGENTS.md` +11 −14, `implement-slice.md` +18 −26, Skill +11 −5, `architect.md` +3 −6, `verifier.md` +6 −1, die vier anderen Typen je +2 −1, also +8 −4, `.d-check.yml` geschätzt +20 −15; Summe der Teile +77 −71), also rund 148 Zeilen vor Review-Fixes; mit Fixes höchstens 250. Wünscht der Planner trotzdem einen Block: Pfade `AGENTS.md`, `.claude/commands/implement-slice.md`, `.claude/agents`, `.harness/skills/reviewer.md`, `.d-check.yml`; `produktcode: 0`, `fakes: 0`; die Tabelle F hat 16 Zeilen der Klasse `andere`, also *S* = 80·16 = 1280 und *M* = 1800 nach `SPEC-128` Nr. 1; die Auslastung läge bei rund 0,1 und ist kein Messpunkt der Faktoren. Der Architect empfiehlt den Block nicht; der Planner hat die Entscheidung am 2026-11-20 gegen den Wortlaut von `SPEC-128` geprüft und bestätigt (§4).

### H. Reihenfolge für den Implementer

1. Vor dem ersten Edit: alle Guide-Dateien messen (`wc -l`, `wc -c`), gegen Abschnitt D halten; Abweichung seit dem Start gibt der Implementer an den Architect zurück.
2. Die Edits aus Abschnitt B wörtlich ausführen, jede Kürzung aus Abschnitt C mit `grep` auf Anker und eingehende Verweise vorab prüfen (K1 bis K6).
3. Messen; `.d-check.yml` setzen (Grenzen auf das gemessene Ist, Kommentar mit Messung vorher und nachher, Summe, Begründung); alles in **einem** Commit, Nachricht per `git commit -F` mit `LH-QA-06` und [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md) (die Kennung genügt der Prüfung `commit-msg-traceability.sh`), beide Messungen und der Grund der Verschiebung stehen darin.
4. Die Mutationen aus Abschnitt F einmal selbst fahren, `make docs-check` und `make zeichen-check` grün, Rohbytes der geschriebenen Dateien gegenlesen (`LC_ALL=C grep -P`), `make gates` unter der Sperre. `harness/gegenproben.stempel` nicht committen (Abschnitt D).
5. Bericht und die Tabelle *Zusage · Mutation · roter Test* mit den Zahlen in §3 nachziehen (`AGENTS.md` §3.10), vor dem Handoff an den Reviewer.

### I. Bericht des Implementers (Handoff an den Reviewer)

**Messung (Zeilen / Bytes, `wc -l` und `wc -c`), vor dem ersten Edit gleich Abschnitt D und am Start:** `AGENTS.md` 326 / 23633 auf 323 / 23905; `implement-slice.md` 220 / 17081 auf 212 / 16180; `architect.md` 53 / 3759 auf 50 / 3614; Skill `reviewer.md` 145 / 8110 auf 151 / 8668; `harness/conventions.md` 228 / 16065 unverändert. Die übrigen Dateien nach dem Diff: `implementer.md` 49 / 3209, `verifier.md` 48 / 3046, `reviewer.md` (Typ) 42 / 2682, `validator.md` 43 / 2633, `planner.md` 38 / 2306 (zwei Zeilen statt einer: die Zeile des Absatzes ist lang); alle unter der Grenze der Agenten (50 / 3614). Summe der Grenzen: 973 / 68648 auf 964 / 68432 (−9, −216), in **einem** Commit mit den Edits; `.d-check.yml` nennt die gemessene Summe (der frühere Kommentar nannte 976 / 68663, seine Blöcke ergaben 973 / 68648; der Kommentar trägt den früheren Wert nicht). Der Stempel `harness/gegenproben.stempel` ist nicht committet. Diff nach §4 (`git diff --numstat 8433db48..HEAD` über die fünf Pfade): im Bericht an den Reviewer. `make groessenmarke` meldet *kein Block* (§4).

**Gegengelesen auf Rohbytes (`LC_ALL=C grep -P` auf Backslash-`u`-vier-Hex, BOM, NBSP, Nullbreite, Steuerzeichen):** `AGENTS.md`, `.claude/commands/implement-slice.md`, die sechs `.claude/agents/*.md`, `.harness/skills/reviewer.md`, `.d-check.yml` — kein Treffer außer dem Beispiel mit der Folge Backslash, `u`, `005B` in §3.18 von `AGENTS.md`, das schon vor dem Slice dort stand und nicht berührt ist. `make zeichen-check` grün.

**Fänger der Zeilen 1 bis 10 und 16** sind Zählungen und Zeilenfolgen mit `grep -c` und `grep -n` über die Dateien (Schleifen über Dateien, Wellen und die sechs Typen); je Zeile steht die Mutation unten, rot gesehen auf einer Kopie des Stands, nie im Zweig. Handlauf, kein Gate, kein Zeitbudget: deterministisch, keine Rate. Vor der Änderung (Stand `63d367bc`) sind 28 der Prüfungen der Zeilen 1 bis 10 rot, nach ihr keine.

| Nr. | Zusage | Mutation (Kopie) | roter Test |
|---|---|---|---|
| 1 | Pflicht in Schritt 17 | die vier Zeilen hinter *Macht der* in Schritt 17 gelöscht | `grep -c 'seit welle-mock'` in `implement-slice.md` 0 statt 1 und `grep -c 'SPEC-201. Nr. 11'` 0 statt 1 |
| 2 | Prüfzeile im Skill, ohne Wortlaut der Pflicht | Prüfzeile gelöscht | `seit welle-mock` im Skill 2 statt 3, `SPEC-201` im Skill 0 statt 1 |
| 3 | Vorlage 1 hinter dem Fixture-Eintrag | Eintrag hinter den Abschnitt LOW verschoben | Zeilenfolge Fixture 82, Zweig 95, LOW 92 rot; die Zählung bleibt 3 und fängt es nicht (die Zeilenfolge ist der Fänger) |
| 4 | Vorlage 2 in Skill und `verifier.md` | (a) Absatz nur in `verifier.md` gelöscht; (b) drei Zeilen nur im Skill gelöscht | (a) `Dein Report besteht` 0 statt 1; (b) `nach dem Schreiben` 0 statt 1 |
| 5 | E1 zwischen den beiden Sätzen | Satz gelöscht | Zeilenfolge 240, leer, 241 rot; Anfang `Eine Menge im Plan nennt ihre Achsen` 0 statt 1 |
| 6 | E2 in drei Dateien | je Datei die Zeile `ohne Pfad und` gelöscht (Schleife, drei Läufe) | je Datei `ohne Pfad und` 0 statt 1, 3 von 3 rot; für Skill und `verifier.md` zusätzlich `seit welle-hybrid` 0 statt 1 |
| 7 | E4 Negativ | Rückfall auf *nennt Grund und Folge, sonst* | `ungefangene Klasse mit ihrem Maß` 0 statt 1 |
| 8 | E4 Laufzeit hinter dem Satz zur Rate | Satz ans Ende des Abschnitts verschoben | Zeile der Rate 181, Zeile des Satzes 189 statt 182 rot |
| 9 | Rückgabe 2 | `Edit` gestrichen | `grep -c '^tools: Read, Write, Edit, Bash$'` 0 statt 1 |
| 10 | Rückgabe 3 in sechs Typen | je Typ die Zeile mit `<kurz>-<rolle>/` gelöscht (Schleife, sechs Läufe) | je Typ 0 statt 1, 6 von 6 rot; Summe 5 statt 6 (bei `planner.md`) |
| 11 | Größe je Regel | je Regel `x` angehängt (vier Läufe von `d-check` auf einer Kopie) | vier mal `file-lines-exceeded` und `file-bytes-exceeded`, 4 von 4 (`AGENTS.md` 324 / 23907, Commands 213 / 16182, Agenten 51 / 3616, Skills 152 / 8670) |
| 12 | Länge der Abschnitte | in §3.10 und §3.15 je 14 Zeilen angehängt, dann 9 und 15 | mit 14: §3.10 trägt 35 Zeilen (rot), §3.15 29 (grün); mit 9 und 15: beide 30, zwei `section-lines-exceeded`, 2 von 2 (Ist 21 und 15, nicht 22 und 16: Berichtigung in Zeile 12 und Abschnitt B) |
| 13 | Grenze gleich Ist | `max-bytes` von `AGENTS.md` auf 23904 | `file-bytes-exceeded` (Datei hat 23905, erlaubt 23904); 23906 bleibt grün (N4) |
| 14 | Pfade lösen auf | Pfad in Schritt 17 um einen Buchstaben (`.txs`) | `codepath-missing` an Zeile 118; im Skill dieselbe Änderung grün (N3), `test -e harness/katalog-ausnahmen.txt` ja |
| 15 | keine unsichtbaren Zeichen | NBSP an `reviewer.md` angehängt (Kopie mit `git init`) | `make zeichen-check` rot: `.harness/skills/reviewer.md:152:2: U+00A0 Zs NBSP`, 1 Befund |
| 16 | Kürzungen nehmen keine Zusage mit | K3 ohne *lässt ihn rot* | `grep -c 'lässt ihn rot'` 0 statt 1; die Träger von K1 bis K6 tragen ihren Satz noch (`grep` an `harness/conventions.md`, `harness/README.md`, `stop-require-gates.sh` mit DEFAULT, STRENG, GRENZEN, `commit-msg`, `hooks-install.mk`, `architect.md`, Vorlage `reviewer.template.md`), 9 von 9 |

**N1 gemessen:** die zehn Löschungen der Zeilen 1 bis 10 (je eine in einer Kopie, bei Zeile 3 die Verschiebung) lassen `make docs-check` grün: 10 von 10, wie erwartet. **N3 gemessen:** ein toter Pfad im Skill bleibt grün (`scan.ignore`), 1 von 1. Die Läufe von `make docs-check` auf Kopien nutzten dasselbe Image mit `--network none`, schreibgeschützt und `--memory=4g`.

**Abweichungen vom Plan (keine inhaltliche):** (a) Zeile 12 nannte *je 14 Zeilen*; §3.15 trägt 15 statt 16 Zeilen (§3.10: 21 statt 22), 14 Zeilen in §3.15 blieben grün; Zeile 12, Zeile 5 und die Entscheidung zu E4 in Abschnitt B sind berichtigt. (b) `planner.md` wächst um zwei Zeilen statt einer (38 statt 37 Zeilen, Bytes gleich): der Satz läuft in drei Zeilen aus, weil die Zeile des Absatzes lang ist; der Umbruch ist nach Abschnitt B frei, und die Datei liegt unter der Grenze. (c) Wortlaut und Orte aller acht Regeln und der Kürzungen K1 bis K6 sind wörtlich ausgeführt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Closure-Notiz `welle-mock-results`); kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln); Reihenfolge der Kette seit 2026-11-08: `slice-gate-integration-determinism`, `slice-guard-schreibwege`, dieser Slice, `slice-groessenformel-neu`. Es laufen höchstens drei Slice-Worktrees zugleich. Der Slice braucht `slice-guard-schreibwege` nicht im Code; er ändert aber dieselben Rollen-Dateien nicht, sodass die Reihenfolge der Kette eine Wahl des Planners ist, kein Blocker. Der User kann ihn vorziehen.

**Start belegt (2026-11-19, Planner):** `welle-mock-results` liegt in `done/`; `in-progress/` enthielt nur `roadmap.md`, kein anderer Slice der Kette lief (`slice-gate-integration-determinism`, `slice-guard-schreibwege` und `slice-gates-ohne-worktrees` liegen in `done/`); `git worktree list` nannte vor dem Start nur den Hauptbaum, der Worktree dieses Slice ist der erste (Verzeichnis odata-recorder-wt-regeln neben dem Hauptbaum, Zweig `slice-regeln-verkoerpern`, Basis `main`). Die Ausgangsmaße aus §1 sind am Start nachgemessen und gleich: `reviewer.md` 145 Zeilen und 8110 Bytes, `implement-slice.md` 220 Zeilen und 17081 Bytes, `AGENTS.md` 326 Zeilen und 23633 Bytes, `architect.md` 53 Zeilen und 3759 Bytes, `verifier.md` 43 Zeilen und 2639 Bytes (Grenzen und Summen in `.d-check.yml`, Kommentar zur Verschiebung vom 2026-11-06); der Implementer misst vor dem ersten Edit trotzdem neu.

**Größenmarke:** kein Block. Der Architect hat es am 2026-11-20 entschieden (§3 Abschnitt G); der Planner hat die Entscheidung vor dem ersten Code-Commit gegen den Wortlaut von `SPEC-128` geprüft und hält sie. Der Vertrag verlangt keinen Block: ein Plan ohne vollständigen Block wird nicht gemessen, Hinweis und Exit 0 (Nr. 3 p), ein Stand ohne Block ist grün (Nr. 8), Code vor der Marke bleibt ungesehen (Nr. 10 d); `AGENTS.md` §3.14 bindet nur eine Marke, die der Plan nennt.

Gründe: (1) Die Schätzung nach Nr. 1 ergäbe *S* = 80·16 = 1280 und *M* = 1800 (Schnittfrage nein, 1280 liegt unter 2223); bei einer Erwartung von rund 148 bis 250 Zeilen (Abschnitt G) wäre die Auslastung rund 0,1. Eine Marke, die das Siebenfache der Erwartung zulässt, fängt keine Überschreitung, und nach Nr. 9 (Auslastung unter 0,60) meldete der Planner sie dem Architect, der die Faktoren aus allen Messpunkten neu rechnet; Nr. 2 leitet sie aber aus drei Go-Slices ab, dieser Slice hat keinen Go-Code und ist kein Messpunkt dafür. (2) Das Wachstum der Guide-Dateien begrenzt `make docs-check` je Datei und Maß (`SPEC-040`), und die Summe der Grenzen steigt nicht (§3 Abschnitt D); eine Marke fügte keine Grenze hinzu.

Folge im Betrieb (Nr. 10 d): `make groessenmarke` meldet den Hinweis *kein Block* und bleibt grün (am 2026-11-20 gefahren); Implementer, Reviewer und Verifier nennen den Lauf mit diesem Hinweis in ihrem Bericht. Ein Block käme nur mit einem Beschluss des Planner vor dem ersten Code-Commit (Pfade, Schätzung und Marke wie in Abschnitt G) und bliebe ein Messpunkt außerhalb der Faktoren.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): der Diff liegt mit den Review-Fixes über 400 Zeilen (Metrik nach `AGENTS.md` §3.14: Summe aus Einfügungen und Löschungen von `git diff --numstat 8433db48..<letzter Commit des Slice>` über `AGENTS.md`, `.claude/commands/implement-slice.md`, `.claude/agents`, `.harness/skills/reviewer.md` und `.d-check.yml`, Tests eingeschlossen; die Basis ist der Commit des Architect, an diesen Pfaden ist die Messung dort 0; 400 ist das 1,6fache der Obergrenze von 250 aus Abschnitt G; keine Marke nach `SPEC-128`, kein Gate fängt sie, §6) oder ein Schnitt entlang der Regeln (Regeln aus `welle-mock` · Regeln aus `welle-hybrid` und die Rückgaben) wäre ohne geteilten Mechanismus möglich; beide Teile teilen die Verschiebung in `.d-check.yml`, ein Schnitt braucht darum zwei Verschiebungen.
- `in-progress` → `open` (blockiert): die Kürzungen lassen sich nicht ohne Verlust eines Trägers finden und der User muss über eine ADR zur Anhebung einer Grenze entscheiden.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag (je Regel des Umfangs ein Eintrag mit `liegt in <Zielort>` und Anker `seit welle-mock` bzw. `seit welle-hybrid`, Zielorte in §3 Abschnitt B, oder eine benannte Lücke, wenn eine Regel nicht ohne Verlust gekürzt werden konnte).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Eine Kürzung entfernt eine Regel, die ein Fänger oder ein Anker trägt — **Ausgang:** entfallen: der Architect hat K1 bis K6 mit Träger belegt (§3 Abschnitt C, kein Anker, kein eingehender Verweis, Zusage jeweils an anderer Stelle); Reviewer und Verifier haben jeden Träger selbst nachgelesen, keine Zusage ging verloren (Tabelle F, Zeile 16); der Retirement-Check gilt weiter (`grundlagen-traceability.md`).
- Die Pflicht aus `SPEC-201` Nr. 11 steht danach in zwei Dateien und driftet — **Ausgang:** entfallen: ein Ort (Schritt 17 von `implement-slice.md`), die Prüfzeile des Skills trägt nur den Zeiger auf den Vertrag; der Vertrag selbst bleibt unverändert (§3 Abschnitt B); Review und Verifikation fanden den Wortlaut der Pflicht nur in Schritt 17. Dass ein Fänger den Wortlaut in der Prüfzeile nicht bemerkte, ist N7 (Zeile 2).
- Die Verschiebung einer Grenze wird als Lockerung gelesen — **Ausgang:** entfallen: die Summe sinkt je Maß (973/68648 auf 964/68432, §3 Abschnitt D), jede Grenze steht auf dem Ist; Reviewer und Verifier haben den Diff von `.d-check.yml` gegen `SPEC-040` Nr. 8 b und 8 e gerechnet; N4 bleibt für spätere Verschiebungen.
- Raster der Randformen (Auftrag 4) ist noch leer — **Ausgang:** entfallen: das Raster steht unten, gefüllt vom Architect am 2026-11-20 vor dem ersten Code-Commit.
- Größe: die Marke steht als Prosa in §4 und ist am Start nicht gesetzt — **Ausgang:** entfallen: entschieden in §3 Abschnitt G und vom Planner gegen `SPEC-128` bestätigt (§4: kein Block, Gründe und Folge); der Planner trägt nichts nach, `make groessenmarke` meldet *kein Block*.
- Die Rückführungsschwelle von 400 Zeilen in §4 ist keine Marke nach `SPEC-128`, und kein Gate misst sie — **Ausgang:** entfallen: der Verifier maß `git diff --numstat 8433db48..HEAD` über die Pfade aus §4 am Stand `47e1f32d`: 81 Einfügungen und 75 Löschungen, Summe 156, unter 400 (der Review am Stand davor: 157); die Schwelle wurde nicht erreicht. Ohne Block fängt kein Gate sie (`SPEC-128` Nr. 10 d); für einen späteren Slice dieser Art bleibt das die Folge.
- Rückgaben des Planner an den Architect aus der Prüfung des Plans (2026-11-20) — **Ausgang:** entfallen, vom Architect am 2026-11-20 behoben: (a) Abschnitt G nennt jetzt 77 Einfügungen und 71 Löschungen (148 Zeilen), die Summe der Teile (die Teile standen richtig, die Summe 105 und 55 nicht); §4 nennt rund 148 bis 250 Zeilen; die Messung des Diffs bleibt die der Entwürfe, `.d-check.yml` ist geschätzt (+20 −15) und nicht im Scratchpad gemessen. (b) Das Maß von N1 lautet jetzt *10 von 10 Löschungen* (eine je Zeile 1 bis 10 der Tabelle F, in neun Dateien) statt *4 von 4 Dateien*. (c) Der Fänger der Zeile 10 zählt das Token `<kurz>-<rolle>/` (ein Wort ohne Leerzeichen, im Satz einmal, vor der Änderung in keiner Datei unter `.claude/` und in `AGENTS.md`); der Umbruch des Satzes in den fünf Typen ohne Block ist frei und nennt §3 Abschnitt B (`.claude/agents/`, Nr. 2) ausdrücklich. Die Fänger der Zeilen 4 und 6 zählen Wortfolgen, die der feste Wortlaut der Blöcke in einer Zeile führt.
- Der Slice-Zweig schreibt `AGENTS.md`, `.harness/skills/reviewer.md`, `.claude/agents/*.md` und `.claude/commands/implement-slice.md`; Nr. 4 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) behält diese Dateien dem Integrationsschritt vor — **Ausgang:** entfallen: der User hat am 2026-11-20 bestätigt, dass der Slice sie ändert und der Implementer dort schreibt (Wortlaut §1); ein Rollen-Lauf, dessen Rechte `.claude/` nicht schreiben lassen, gibt den Punkt zurück, statt es zu umgehen.
- `.d-check.yml` steht in den Eingaben der Gegenproben; der Stempel `harness/gegenproben.stempel` wird mit der geänderten Datei veraltet — **Ausgang:** offen; Adresse: Planner oder Hauptagent im Integrationsschritt (nach dem Merge `make gegenproben`, Stempel allein committen, Nr. 2 der Übergangsregeln); der Implementer committet den Stempel nicht, Reviewer und Verifier fuhren `make gegenproben` nicht (F-9, V-4).
- Der Kommentar in `.d-check.yml` nennt die Summe 976 Zeilen und 68663 Bytes, die Blöcke ergeben 973 und 68648 — **Ausgang:** entfallen: der Kommentar trägt die gemessene Summe und nennt den früheren Wert nicht (Befund F-1 des Reviews `2026-11-20-slice-regeln-verkoerpern-review`, behoben im Fix-Commit; §3 Abschnitt D).
- Die Längen der Abschnitte §3.10 und §3.15 im Plan (22 und 16 von 29) und die Mutation der Zeile 12 (je 14 Zeilen) stimmten nicht mit der Messung von `make docs-check` überein (21 und 15; 14 Zeilen in §3.15 ergeben 29 und bleiben grün) — **Ausgang:** entfallen: der Implementer hat Zeile 5, Zeile 12 und die Entscheidung zu E4 in §3 berichtigt und im Absatz I benannt; keine Entscheidung betroffen.
- Für E4 liegt kein Diff-Vorschlag im Repo (der Bericht der Closure von `slice-gegenproben-zeitbudget` ist nicht committet); der Architect hat den Wortlaut aus den Beobachtungen abgeleitet — **Ausgang:** weiter offen, Einspruch möglich: Adresse User (§3 Abschnitt B, *E4*); der Reviewer konnte den Wortlaut nur gegen den Plan prüfen, nicht gegen eine Vorlage (Befund F-8 des Reviews `2026-11-20-slice-regeln-verkoerpern-review`).
- Der Zeiger *Übergangsregeln Nr. 8* in den sechs Typen kann bei einem Umbau der Übergangsregeln auf eine nicht mehr vorhandene Nummer zeigen — **Ausgang:** weiter offen; akzeptiertes Negativ N6 (§3 Abschnitt F); Adresse: Planner beim Umbau der Übergangsregeln.
- Die Regel E4 (Negativ mit Klasse und Maß) steht nur in `AGENTS.md` §3.15; der Prüfer *Akzeptiertes Negativ ohne Grund und Folge* im Skill (`.harness/skills/reviewer.md`) und Schritt 14 von `implement-slice.md` fragen weiter nur nach Grund und Folge, kein Konsument prüft Klasse und Maß (V-2 der Verifikation `2026-11-20-slice-regeln-verkoerpern-verifikation`) — **Ausgang:** offen, als Frage; Adresse: User oder Architect (der Umfang des Users nannte allein §3.15 und §3.10, der Slice hat die Prüfzeile nicht berührt); Folge im Betrieb: Die Regel kann ohne Prüfer folgenlos bleiben; gegengehalten wird allein das Register, das das Wiederauftreten von `BEO-REPO/ein-negativ-benennt-die-ungefangene-klasse-zu-eng` zählt.
- Schritt 17 und die Prüfzeile sagen, wer einen Fehlercode, einen Schlüssel oder ein Kommando erreichbar mache, entferne *ihn* aus `harness/katalog-ausnahmen.txt`; die Datei führt nur Codes (F-7, Wortlaut von `SPEC-201` Nr. 11) — **Ausgang:** offen; Adresse: Architect, falls der Vertrag nachgeschärft wird; kein Nehmer unter den offenen Slices.

**Raster der Randformen (`AGENTS.md` §3.16; Auftrag 4, Architect 2026-11-20).** Dieser Slice liefert keinen Vertrag nach §3.10; das Raster hält die Randformen der Operationen auf Guide-Dateien fest. Orte sind `SPEC-040` und der Plan (§3), wo die Spezifikation nichts regelt.

| Operation | fehlend | leer | unbekannt oder fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Satz einfügen | Datei fehlt: `file-no-match` (`SPEC-040` Nr. 6 a); Abschnitt `### 3.<n>` fehlt: `section-missing` (Nr. 9); Einfügestelle nicht auffindbar: der Implementer hält an und gibt an den Architect zurück (§3.12) | gilt nicht: der Wortlaut steht in §3 Abschnitt B, der Implementer setzt keinen Wert; eine leere Datei ist 0/0 und grün (Nr. 6 b) | Pfad oder Kennung im Wortlaut löst nicht auf: `codepaths` und `ids` von `make docs-check` für `AGENTS.md`, Commands und Agenten; im Skill ungeprüft (`scan.ignore`, N3) | Lauf endet vor dem Commit: der Stop-Hook (`implement-slice.md`, Punkt *Gate-Nachweis + Stop-Hook*) verlangt `make gates`; ein Teil-Edit bleibt ungedeckt im Arbeitsbaum | derselbe Satz in zwei Dateien: §3 Abschnitt B (ein Ort je Pflicht; E2 steht je Rolle mit eigenem Wortlaut); Zählprobe in Tabelle F, Zeile 1 bis 8 |
| Text kürzen | Text nicht mehr da (Datei seit dem Start geändert): der Implementer misst vor dem ersten Edit neu und gibt zurück | gilt nicht: jede Kürzung ersetzt durch den Wortlaut aus Abschnitt B oder streicht einen Block ganz (K5, K6) | Kürzung trifft Anker oder eingehenden Verweis: K1 bis K6 geprüft (Abschnitt C); der Implementer wiederholt `grep` vor dem Edit | Kürzung ohne Ergänzung: die Grenze liegt über dem Ist, Senken ist frei (`SPEC-040` Nr. 8 a) | gilt nicht: jede Kürzung einmal |
| Grenze verschieben | Regel ohne Datei: `file-no-match` (Nr. 6 a) | Maß fehlt oder negativ: Exit 2 (Nr. 7) | Grenze unter dem Ist: rot (Nr. 8 a); Grenze über dem Ist oder Hebung ohne Senkung: kein Gate (Nr. 10 a, 8 e), Reviewer und Verifier rechnen (N4) | Hebung und Senkung in getrennten Commits: ein Zwischenstand verstößt gegen Nr. 8 b; entschieden: ein Commit (Abschnitt D) | zwei Verschiebungen im Slice: gilt nicht, eine; jede weitere trüge Messung und Begründung (Nr. 8 b) |
| Anker oder Zeiger setzen | Regel ohne `seit welle-<Kennung>`: Zählprobe, Tabelle F, Zeile 1 bis 8 | gilt nicht: der Anker steht im Wortlaut | Welle ohne `done/welle-<Kennung>-results.md`: beide Dateien liegen (`welle-mock-results.md`, `welle-hybrid-results.md`), Probe `test -e` | gilt nicht: ein Anker ist Text, kein Vorgang mit Zwischenzustand | zwei Wellen in einer Überschrift: erlaubt, Form wie in §3.16 (`geschärft seit … und seit …`) |

Weitere Randformen aus Auftrag 4: **Regel in zwei Dateien** (Drift, `AGENTS.md` §1): entschieden in §3 Abschnitt B (Pflicht an einem Ort, Prüfzeile ohne Wortlaut der Pflicht). **Eine Kürzung löscht einen Anker `seit welle-<Kennung>`:** keine der sechs Kürzungen trägt einen (§3 Abschnitt C); der Retirement-Check (`grundlagen-traceability.md`) bleibt für spätere Kürzungen in Kraft (N1). **Verschiebung und Kürzung im selben Commit:** ja, ein Commit (§3 Abschnitt D).

**Offene Aufträge an den Architect** (Stand Start 2026-11-19): (1) bis (5) sind am 2026-11-20 entschieden, siehe §3 Abschnitte B bis G und das Raster oben; (6) entschieden: der User hat E1, E2, E4 und die Rückgaben 2 und 3 gewählt, die Zeilen und Bytes und die Kürzungen stehen in §3 Abschnitt C und D. Offen bleibt keiner.

**Aufgenommene Ergänzungen (Entscheidung des Users, 2026-11-20):** E1, E2, E4 und die Rückgaben 2 und 3 aus der Tabelle des Planners vom Start; der Umfang in §1 und die DoD in §2 tragen sie. Nicht aufgenommen und beim User geblieben: E3 (Auftrag an den Architect zur Nullprobe), E5, E7 bis E9 aus `welle-hybrid-results` sowie die Nr. 3 und 4 aus §7 von `slice-guard-schreibwege`; `zeit-zusage-an-einer-form-und-einer-locale` und `allaussage-an-jeder-stelle-mit-faenger-an-erster-und-zweiter-stelle` (je 1×) bleiben unter der Schwelle gezählt.

**Sendungen am Start geprüft (`AGENTS.md` §3.13, 2026-11-19; fortgeschrieben 2026-11-20):** (a) *Eingang:* die Sendung dieses Slice (Vorlage 1, Vorlage 2, Pflicht aus `SPEC-201` Nr. 11) steht in §1 *Übernimmt* und in der DoD; kein Ausschluss in §1 trifft sie. Die Sendungen E1, E2, E4 und die Rückgaben 2 und 3 stehen am 2026-11-20 in §1 *Übernimmt* und in der DoD (Geber `welle-hybrid` bzw. `slice-guard-schreibwege`, Entscheidung des Users). Der Geber `welle-mock-results` führt die Vorlagen als „beim User offen“ ohne Nehmer (das Zeitdokument liegt in `done/` und bleibt); den Träger nennen das Drift-Log der Roadmap (2026-11-08), `welle-hybrid-results` (*Folge-Slices*) und die `state.md` der Einträge, die dieser Plan am Start auf `in-progress/` nachzieht. (b) *Ausgang:* der Ausschluss „Größenformel und Zuschlag“ nennt `slice-groessenformel-neu` (`open/`, Kette, vierter); dessen §1 führt die Formel als Ziel und schließt sie nicht aus, die Adresse trägt. Der frühere Ausschluss „Werkzeugliste der Rollen-Typen“ entfällt: sie ist seit 2026-11-20 Umfang. Ein Sensor für `SPEC-201` Nr. 9 c und für die Form des Report-Verweises (§1) hat keinen Nehmer unter den offenen Slices (geprüft am 2026-11-19: `grep` über `open/` und `next/` nach „Erreichbarkeit“ und `SPEC-201`: `slice-handbuch-mock` und `slice-handbuch-etag-zeile` nennen `SPEC-201` nur als Abgleich der Schlüssel, `slice-test-security` meint die Erreichbarkeit des Admin-Listeners; keiner nimmt einen Sensor an); er bleibt beim User. (c) *Dritte:* kein anderer Slice nennt diesen Slice als Nehmer außer der Roadmap und `welle-hybrid-results`.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Entstanden aus der Closure von `welle-mock` (Entscheidung des Users vom 2026-11-08), am 2026-11-20 um die Punkte aus `welle-hybrid` und `slice-guard-schreibwege` erweitert (Entscheidung des Users); geschlossen am 2026-11-20 vom Planner im Zweig `slice-regeln-verkoerpern` (HEAD vor der Closure `47e1f32d`, Basis des Diffs der Commit des Architect `8433db48`). Review (`2026-11-20-slice-regeln-verkoerpern-review`: 1 HIGH, 3 MEDIUM, 2 LOW, 3 INFO) und Verifikation (`2026-11-20-slice-regeln-verkoerpern-verifikation`: freigegeben mit zwei Auflagen, 0 HIGH, 1 MEDIUM, 1 LOW, 2 INFO) liefen je einmal; Durchstich vor Härtung: keine Wiederholungsschleife. Merge nach `main`, `make gates` im Hauptbaum nach dem Merge und `make gegenproben` samt Stempel sind der Integrationsschritt, nicht Teil dieser Notiz.

- **Was hat funktioniert:** Der Architect legte Ort, Wortlaut, Kürzungen, Messung, Raster und Tabelle *Zusage · Fänger* vor dem ersten Edit fest; der Implementer führte den Wortlaut aus, Reviewer und Verifier verglichen ihn Zeichen für Zeichen mit Abschnitt B und fanden keine inhaltliche Abweichung. Die sechs Kürzungen K1 bis K6 nahmen keine Zusage mit: beide Rollen lasen jeden Träger selbst nach, kein eingehender Verweis auf eine gekürzte Überschrift. Die Verschiebung der Grenzen stand in **einem** Commit mit den Edits, die Summe sank je Maß, kein Gate wurde gelockert, keine ADR war nötig. Der Verifier mutete an Elementen, die der Plan nicht nannte (je Fragment der Sätze), und fand so das zu kleine Maß von N7.
- **Messung** (Zeilen / Bytes, `wc -l` und `wc -c`, vor dem Slice und am Stand `47e1f32d`; Grenzen gleich Ist):

  | Datei / Regel | vorher | nachher | Änderung |
  |---|---|---|---|
  | `AGENTS.md` | 326 / 23633 | 323 / 23905 | −3 / +272 |
  | `.claude/commands/*.md` (Maximum `implement-slice.md`) | 220 / 17081 | 212 / 16180 | −8 / −901 (Grenze 221 auf 212: −9) |
  | `.claude/agents/*.md` (Maximum `architect.md`) | 53 / 3759 | 50 / 3614 | −3 / −145 |
  | `.harness/skills/*.md` (Maximum `reviewer.md`) | 145 / 8110 | 151 / 8668 | +6 / +558 |
  | `harness/conventions.md` | 228 / 16065 | 228 / 16065 | 0 |
  | **Summe der Grenzen** | 973 / 68648 | 964 / 68432 | −9 / −216 |

  Die übrigen Typen nach dem Diff: `implementer.md` 49 / 3209, `verifier.md` 48 / 3046, `reviewer.md` (Typ) 42 / 2682, `validator.md` 43 / 2633, `planner.md` 38 / 2306. Diff nach §4 (`git diff --numstat 8433db48..HEAD` über `AGENTS.md`, `implement-slice.md`, `.claude/agents`, `.harness/skills/reviewer.md`, `.d-check.yml`, Verifier am Stand `47e1f32d`): 81 Einfügungen und 75 Löschungen, Summe 156 (Erwartung 148, Obergrenze mit Fixes 250, Rückführungsschwelle 400). `make groessenmarke`: Hinweis *kein Block*, Exit 0 (`SPEC-128` Nr. 10 d; Verifier). `make docs-check`: 968 Dateien, 0 Befunde; `make zeichen-check`: 2003 Dateien ok (Verifier, Stand `47e1f32d`). `make gates`: Exit 0 am Stand `47e1f32d` im Zweig, sauberer Baum, unter der Sperre (Verifier; der Implementer nannte keinen Lauf, V-3); die Closure ändert nur Plan, Register, Reports und Roadmap.
- **Was ging anders als geplant:** (a) Der Implementer überschrieb beim Einfügen des Berichts die Überschrift `## 4. Trigger` (F-2); behoben. (b) Der Kommentar in `.d-check.yml` beschrieb abwesenden Text, die Zahl des früheren Kommentars (F-1, HIGH nach `AGENTS.md` §3.7); der Auftrag dazu stand im Plan, behoben mit der gemessenen Summe. (c) Entwurfswerte von `verifier.md` und `planner.md` standen in §3 weiter, nachdem der Bericht sie berichtigt hatte (F-6); behoben, `AGENTS.md` §3.9. (d) Das Maß von N7 war zu klein (V-1): *7 von 16* statt *10 von 16*, in dem Slice, der die Regel *Klasse und Maß* schreibt. Befund-Bilanz mit Ausgang je Befund:
  - Review: F-1 (HIGH, Kommentar) behoben; F-2 (MEDIUM, Überschrift) behoben; F-3 (MEDIUM, Fänger an Wortfragment) akzeptiertes Negativ N7, Maß berichtigt (V-1); F-4 (MEDIUM, Zeiger auf `SPEC-201`) N8; F-5 (LOW, Ortszusagen) N9; F-6 (LOW, Entwurfswerte) behoben; F-7 (INFO, Wortlaut spricht von Schlüssel und Kommando, die Ausnahmen führen Codes) §6, offen, Adresse Architect; F-8 (INFO, E4 ohne Vorlage) §6, offen, Adresse User; F-9 (INFO, Stempel) §6, offen, Integrationsschritt.
  - Verifikation: V-1 (MEDIUM, Maß von N7) behoben in dieser Notiz und in N7; V-2 (LOW, kein Konsument für Klasse und Maß im Skill und in Schritt 14) §6, offen, Frage an User oder Architect; V-3 (INFO, kein Lauf von `make gates` im Bericht des Implementers) mit dem Lauf des Verifiers belegt, siehe Messung; V-4 (INFO, Stempel) §6, offen, Integrationsschritt.
- **Akzeptierte Negative** (Grund, Folge, Klasse und Maß; Wortlaut in §3 Abschnitt F): N1 Löschen eines Satzes in einer Guide-Datei fängt kein Gate — Grund: `make docs-check` misst Größe, Verweise und Richtung, nicht Inhalt; Folge: eine spätere Kürzung kann eine Regel still entfernen, Reviewer am Diff und Retirement-Check am Anker halten dagegen; Maß: 10 von 10 Löschungen bleiben grün (Implementer gemessen, Verifier an Kopien bestätigt). N2 ob eine Rolle die Regel befolgt, fängt kein Test — Grund: Prosa wirkt über den Lauf der Rolle; Folge: eine Regel kann wirkungslos bleiben, das Register zählt das Wiederauftreten; Maß: 8 von 8 Regeln ohne Fänger der Befolgung. N3 Pfade und Kennungen im Skill prüft kein Gate — Grund: `.harness/**` steht in `scan.ignore`; Folge: ein toter Pfad im Skill bleibt grün, die Handprobe `test -e` fängt ihn; Maß: 1 von 1 gemessen grün, 1 von 2 Stellen ohne Handprobe. N4 eine Grenze über dem Ist und eine Hebung ohne Senkung fängt kein Gate — Grund: Gleichheit mit dem Ist zwänge das Senken in jeden Kürzungs-Commit; Folge: eine zu hoch gesetzte Grenze passiert, Reviewer und Verifier rechnen nach; Maß: 5 von 5 Grenzen von Hand gerechnet, 1 von 1 gemutet (330 Zeilen für `AGENTS.md`) grün. N5 fünf Typen behalten den Kommentar `ANPASSEN` — Grund: nur `architect.md` steht auf der Grenze; Folge: uneinheitliche Typen, kein Lauf liest den Kommentar als Regel; Maß: 5 von 6 Typen. N6 der Zeiger *Übergangsregeln Nr. 8* ist ohne Link — Grund: die Typen verweisen nie mit Pfad in `docs/plan/planning/`; Folge: nach einem Umbau der Übergangsregeln zeigt der Satz auf eine fehlende Nummer; Maß: 6 von 6 Typen, Auslöser der Umbau, Adresse Planner. N7 die Fänger der Tabelle F zählen ein Fragment der Zusage, nicht den Satz — Grund: ein Gate liest keinen Satzinhalt; Folge: Umformulierung oder Teilverlust im selben Satz bleibt grün, Reviewer am Diff und Retirement-Check halten dagegen; Maß: **10 von 16** Zeilen (1, 2, 3, 4, 5, 6, 7, 8, 10, 16), berichtigt von 7 (V-1). N8 der Zeiger auf `SPEC-201` Nr. 11 prüft kein Gate — Grund: `ids` verlangt Links nur für `LH-`, `ADR-` und `MR-`; Folge: ein toter Zeiger (`SPEC-999`) bleibt grün, Handprobe `grep` und `test -e`; Maß: 2 von 2 Stellen (Schritt 17, Prüfzeile). N9 die Fänger der Ortszusagen sind Handproben — Grund: kein Gate liest die Reihenfolge von Sätzen; Folge: wer die Probe nur als *nach* liest, lässt eine Verschiebung grün; Maß: 3 von 3 Zeilen (3, 5, 8).
- **Steering-Loop-Eintrag** (je Regel ein Eintrag; Anker in den Zielorten, Auflösung nach `grundlagen-traceability.md` §Herkunfts-Anker):
  - Wer einen Fehlercode, einen Schlüssel oder ein Kommando erreichbar macht, entfernt den Code im selben Slice aus den Ausnahmen und führt ihn im Handbuch (`SPEC-201` Nr. 11); der Reviewer prüft es mit einer MEDIUM-Zeile ohne Wortlaut der Pflicht — liegt in `.claude/commands/implement-slice.md` (Schritt 17, die Pflicht) und `.harness/skills/reviewer.md` (die Prüfzeile); Anker `seit welle-mock` in beiden.
  - Eine Zusage, die `make test-race` melde jedes ungeordnete Paar, nennt die Zweige aus dem Abdeckungsprofil der Detektor-Tests oder aus einer Menge im Modell mit Zählprüfung, nicht aus der Erinnerung des Autors (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, 18×) — liegt in `.harness/skills/reviewer.md` (MEDIUM-Liste, Anker `seit welle-mock`).
  - Ein Report der Rolle besteht `make docs-check`: eine `ADR-*`-Kennung steht als auflösender Link, der Lauf folgt dem Schreiben (`BEO-REPO/report-der-rolle-faellt-docs-check`, 5×) — liegt in `.harness/skills/reviewer.md` und `.claude/agents/verifier.md` (Anker `seit welle-mock`; zuvor in `state.md` als Vorschlag mit dem Anker `seit welle-betrieb`, verkörpert ist `seit welle-mock`).
  - Eine Menge im Plan nennt ihre Achsen, der Fänger führt jede, oder das Negativ nennt die ungeführte (`BEO-REPO/menge-ohne-ihre-positionen`, 6×) — liegt in `AGENTS.md §3.15` (Anker `seit welle-hybrid` in der Überschrift).
  - Ein Review-Report steht in Plan und Evidence mit seiner Kennung ohne Pfad und Endung, nie als Dateiname oder Link, sonst sperrt `make archive-welle` mit `[haenger]` (`BEO-REPO/verweis-auf-review-report-sperrt-die-archivierung`, 3×) — liegt in `.claude/commands/implement-slice.md` (Schritt 25), `.harness/skills/reviewer.md` (Output-Schema) und `.claude/agents/verifier.md` (Anker `seit welle-hybrid`). Diese Notiz hält die Regel selbst ein: Review und Verifikation stehen mit Kennung, ohne Pfad und Endung.
  - Ein akzeptiertes Negativ nennt die ungefangene Klasse mit ihrem Maß (n von m), nicht eine Aufzählung oder ein Fenster als Zahl, das die nächste Messung überholt (`BEO-REPO/ein-negativ-benennt-die-ungefangene-klasse-zu-eng`, jetzt 5×) — liegt in `AGENTS.md §3.15` (Anker `seit welle-hybrid` in der Überschrift). Der Slice hat die Regel am eigenen Negativ N7 nicht gehalten (V-1); Konsument im Skill steht aus (§6, V-2).
  - Eine Laufzeit steht als Gewinn oder Verlust nur mit mehreren Läufen je Seite, der Last des Hosts vor und nach jedem Lauf und nur, wenn der Unterschied die Streuung übersteigt (`BEO-REPO/laufzeitmessung-widerspricht-sich-unter-last`, 3×) — liegt in `AGENTS.md §3.10` (Anker `seit welle-hybrid` in der Überschrift).
  - Ohne Anker, kein Eintrag nach der Paarung: die Rückgabe 2 (`Edit` in der Werkzeugliste des Typs `architect`, Werkzeugliste statt Regel) und die Rückgabe 3 (Satz zu den Scratchpad-Namen in den sechs Typen; Herkunft ist der Zeiger *Übergangsregeln Nr. 8*, Anker `seit slice-guard-schreibwege` steht dort).
  - Lehre dieser Closure: Die Regel *Klasse und Maß* hat im selben Slice ihr Negativ zu eng gefasst; der Verifier fand es durch Mutation je Fragment an Zeilen, die der Review nicht genannt hatte. Ein Fänger, der ein Wortfragment zählt, deckt den Satz nicht (N7); der Konsument der Regel steht aus (V-2).
- **Entscheidungen beim User:** (1) *E4-Wortlaut:* Einspruch gegen den vom Architect aus den Beobachtungen abgeleiteten Wortlaut von §3.15 und §3.10, der Diff-Vorschlag der Closure von `slice-gegenproben-zeitbudget` liegt nicht im Repo (§6, F-8). (2) *Konsument von E4 (V-2):* soll der Prüfer *Akzeptiertes Negativ ohne Grund und Folge* im Skill und Schritt 14 von `implement-slice.md` zusätzlich nach Klasse und Maß fragen? Ohne diese Zeile hat E4 keinen Konsumenten (Grundsatz *Jedes Artefakt hat einen Konsumenten*, `grundlagen-harness-dateien.md`); die Aussage stünde dann an drei Stellen (`AGENTS.md` §3.9, §3.17). Adresse User oder Architect. Die nicht gewählten Punkte E3, E5, E7 bis E9 und Nr. 3 und 4 aus §7 von `slice-guard-schreibwege` bleiben beim User, unverändert.
- **Beobachtungs-Register (`../observations/`):** `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, `report-der-rolle-faellt-docs-check`, `menge-ohne-ihre-positionen`, `verweis-auf-review-report-sperrt-die-archivierung`, `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` und `laufzeitmessung-widerspricht-sich-unter-last` stehen in `state.md` auf *verkörpert* mit Zielort und Anker (Nr. 4 der Übergangsregeln: Register-Fortschreibung im Integrationsschritt, hier auftragsgemäß mit dem Closure-Commit im Zweig). Neuer Beleg: `evidence/slice-regeln-verkoerpern.md` in `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (fünftes Auftreten: das Maß des Negativs N7 war zu klein, V-1); V-2 steht dort im `state.md` als offen mit Adresse. *Keine Beobachtung angefallen* trifft nicht zu.
- **Folge-Slices:** keine. Offene Punkte ohne Slice: Stempel (Integrationsschritt), E4-Wortlaut und Konsument von E4 (User, Architect), Zeiger *Übergangsregeln Nr. 8* (N6, Planner), Wortlaut zu Schlüssel und Kommando in `SPEC-201` Nr. 11 (F-7, Architect); alle in §6 mit Adresse (`AGENTS.md` §3.13: kein offener Slice nennt die Guide-Dateien als Gegenstand; `slice-groessenformel-neu` schließt Regeln der Rollen aus seinem Umfang nicht ein).
- **Gegenlesen** (`AGENTS.md` §3.18): die Dateien, die diese Closure schreibt (der Plan, die beiden Reports unter `docs/reviews/`, die sechs `state.md` und der neue Beleg unter `docs/plan/planning/observations/BEO-REPO/`, die Roadmap), sind auf Rohbytes gegengelesen (`LC_ALL=C grep -P`: BOM, NBSP, Nullbreite, Zeilentrenner, Steuerzeichen, CR und die Folge Backslash, `u`, vier Hex-Ziffern; Ergebnis im Bericht an den Auftraggeber); `make zeichen-check` deckt dieselbe Menge im Gate.
- **Gegenproben:** Eingabe berührt: `.d-check.yml` steht in `EINGABEN` von `tools/harness/gegenproben-eingabe.sh`, der Stempel `harness/gegenproben.stempel` ist am Zweig veraltet; `make gegenproben` und den Stempel allein committen ist der Integrationsschritt (Nr. 2 der Übergangsregeln), die Fälle 156 bis 170 und 213 ändern ihren Ausgang laut Plan nicht (nicht gefahren, §6).
- **Risiken aus §6:** jedes mit genau einem Ausgang, siehe §6 (entfallen: Kürzung, Pflicht an einem Ort, Lockerung, Raster, Marke, Rückführungsschwelle, Rückgaben an den Architect, Rechte, Längen, Kommentar der Summe; weiter offen: Stempel, E4-Wortlaut, Zeiger *Übergangsregeln Nr. 8*, Konsument von E4, Wortlaut zu Schlüssel und Kommando).
- **Drei Paarungen:** nach dem Move, siehe das letzte DoD-Häkchen in §2.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×) und `BEO-REPO/report-der-rolle-faellt-docs-check` (5×) sind der Gegenstand der Closure von `welle-mock`; seit 2026-11-20 gehören `menge-ohne-ihre-positionen` (6×), `verweis-auf-review-report-sperrt-die-archivierung` (3×), `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (4×) und `laufzeitmessung-widerspricht-sich-unter-last` (3×) dazu (Entscheidung des Users). `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` und `BEO-REPO/kommentar-nennt-mehr-elemente-als-der-test-prueft` berühren Auftrag 2 (Kürzung darf nichts zusagen, was kein Fänger trägt; Tabelle F, Zeile 16). Register am Start gesichtet (2026-11-19, Zähler aus den Dateien unter `evidence/`). `zeit-zusage-an-einer-form-und-einer-locale` und `allaussage-an-jeder-stelle-mit-faenger-an-erster-und-zweiter-stelle` (je 1×) bleiben unter der Schwelle gezählt.
