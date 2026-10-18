# Reviewer-Skill — odata-recorder

* Status: Accepted
* Bezug: AGENTS.md §3 (Harte Regeln), `harness/README.md` §Source precedence
* Gilt für: die Review-Rolle (Agent-Typ `reviewer`)

## Kontext-Eingang (Pflicht)

Was der Reviewer *immer* mitbringt, bevor er den Diff liest:

- Diff des PR
- `spec/lastenheft.md` (für referenzierte `LH-*`-IDs)
- ADRs, deren ID im PR oder in der Commit-Message vorkommt
- `AGENTS.md` §"Hard Rules"
- vorherige Findings am gleichen Modul (letzte ~5 PRs)

Ohne diesen Block sieht der Reviewer den Code, aber nicht *die Verträge, gegen
die er prüft*.

## Klassifikation

Jeder Anker HIGH/MEDIUM/LOW hat eine *konkrete* Liste — nicht generisch. INFO ist
bewusst kurz (Ergänzungs-Kanal, nicht Hauptkanal).

**HIGH** — eines der folgenden:
- ADR-Verstoß (Layer, Tool, Hard Rule)
- Sicherheits-Anti-Pattern (Injection, fehlende Auth-Prüfung)
- Korrektheitsfehler im *kritischen* Pfad (Capture und Redaction vor der Persistenz, Replay-Cursor, Workspace-State-Mutation, Admin-Authentisierung)
- Suppression eines Gates (`#noqa`, `//nolint`, `[SuppressMessage]`) ohne ADR
- **Norm nur im Template-Kommentar** — eine Regel steht im `<!-- -->`-Block
  eines `.template.md` und nirgends sonst. Sie ist beim Adopter weg, sobald er
  die Kommentare entfernt. Kein Gate fängt das (siehe Baseline-Regelwerk
  `grundlagen-harness-dateien.md` §Template-Schichtung)
- **Kommentar trägt keine der Kommentar-Klassen** — ein Kommentar in Code, Config
  oder Skript beschreibt die verworfene Alternative („Ohne X wäre …"), einen
  abwesenden Text („früher stand hier …") oder bricht mitten im Satz ab, weil
  eine Teilersetzung den Rest stehen ließ. Kein Gate fängt das (siehe
  Baseline-Regelwerk `grundlagen-harness-dateien.md` §Was ein Kommentar trägt)
- **Zustandsfeld trägt Chronik** — eine `Stand`-/`Status`-Zelle (Roadmap,
  Beobachtungs-Register, Meilenstein) erzählt, wie der Zustand entstand, statt
  Zustand und Beleg als Anker zu nennen; oder ein Drift-Log protokolliert
  Schließungen und erreichte Meilensteine. Kein Gate fängt das (siehe
  Baseline-Regelwerk `grundlagen-harness-dateien.md` §Was ein Kommentar trägt,
  *Dieselbe Regel für Zustandsfelder*)
- **Spec-Straten widersprechen sich** — eine Aussage in `spec/spezifikation.md` oder
  `spec/architecture.md` weicht von `spec/lastenheft.md` ab, oder dieselbe
  normative Aussage steht in zwei Straten mit verschiedenem Inhalt (Beispiel: Cursor-Wirkung
  von Snapshot-Restore). Bei Konflikt gewinnt das höhere Stratum (Source Precedence);
  die Architektur-Sicht führt keine neue Anforderung und keine Wellen, Slices, ADR-Bezüge
  oder Sprachbezüge (AGENTS.md §3.4)
- **Geheimnis erreicht eine persistente oder diagnostische Senke** — Rohbytes, unredigierte
  Header oder Bodies, Tokens oder Query-Werte gelangen in Recording, Blob, Log, Span-Attribut
  oder Fehlerantwort; oder ein Admin-Endpunkt außer `GET /healthz` ist ohne Token erreichbar
  (`LH-FA-14`, `LH-FA-13`, `LH-QA-04`)
- **Zustandsmodell verletzt** — ein Reset oder Restore ändert Cursor oder Entity-State anders als
  `LH-FA-09`/`LH-FA-10` festlegen, oder `complete`-Coverage entsteht ohne Vollständigkeitsbeweis
  (`LH-FA-11`, `LH-FA-18`)

**MEDIUM** — eines der folgenden:
- unklare Fehlerbehandlung am Rand des Spec-Bereichs
- fehlende Negativtests bei neuem öffentlichem Vertrag
- Wiederholung eines Musters, das schon zweimal LOW war

**LOW** — *mit Konventions-Anker* (ADR, Hard Rule, Linter-Regel, Eintrag im
Reviewer-Skill): stilistisch unschön ohne semantische Auswirkung, einmalige
Tippfehler, unbenutzte Imports.

**INFO** — Hinweis ohne erwartete Aktion (z. B. „diese Stelle hat ein passendes
ArchUnit-Pendant, das du nicht kennst").

> **Pflicht beim Ausfüllen (Modul 10 §Übungen):** Die HIGH-Liste muss mindestens
> *zwei* repo-spezifische Regeln nennen, die ein generischer Skill nicht abdeckt.
> Ist der Skill ohne sie, ist er noch nicht scharf genug — dann kommt bei einem
> Lauf auf einem realen Diff keines deiner Repo-HIGHs zur Anwendung.

## Was dieser Skill NICHT macht

- Keine Lösungsvorschläge („schreib das so") — Reviewer kategorisiert,
  Implementer entscheidet.
- Kein Refactoring-Vorschlag, der über den Diff hinausgeht.
- Keine Verifikation gegen DoD — das ist Verifier-Aufgabe (Modul 11).
- Keine Validation gegen reale Bedürfnisse — das ist Validator-Aufgabe.
- **Kein Stil-Polizist:** Formatierung oder Benennung ohne Konventions-Anker ist
  kein Finding.
- **Kein HIGH- oder MEDIUM-Finding ohne Failure-Szenario:** was sich nicht als
  konkretes Versagen erzählen lässt, wird nicht als HIGH oder MEDIUM gemeldet.

Wenn etwas auffällt, das in diese Kategorien gehört: ein INFO-Finding mit Verweis
auf die zuständige Rolle.

## Output-Schema

Jedes Finding:

- `kategorie`: HIGH | MEDIUM | LOW | INFO
- `quelle`: ADR-ID, `LH-*`-ID, Hard-Rule-Name oder „Maintainability"
- `pfad`: Datei · wörtliches, in der Datei eindeutig auffindbares Kurzzitat
  der Stelle als Anker; die Zeile darf als Lesehilfe dazu, ist aber nicht der Anker
- `befund`: 1–2 Sätze, beobachtbar, ohne Lösungsvorschlag
- `verifizierbar`: ja/nein — gibt es einen Gate-Lauf, der es bestätigen würde?
- `klasse`: stabile Kurz-Bezeichnung des Fehlermusters, z. B. „Tie-Break in
  sortierender Operation nicht dokumentiert" — speist den
  Steering-Loop-Zähler (siehe §Pflege)

Zusätzlich am Ende: eine Zeile „geprüft, ohne Befund" pro betrachtetem
Verzeichnis (Negativbefund-Zeile — sonst ist „keine Findings" nicht von „nicht
geprüft" unterscheidbar). Report-Gerüst für den ganzen Lauf:
`.harness/baseline/<tag>/templates/docs/reviews/review-report.template.md`, ein Report pro Lauf, Folgeläufe als
neue Datei statt Überschreibung.

## Pflege (Steering-Loop)

Bei dreimaligem Auftreten desselben Findings:

- ist die Kategorie noch richtig? → Klassifikation schärfen
- gibt es einen ADR/`AGENTS.md`-Eintrag, der das verhindert hätte?
  → Folge-ADR oder `AGENTS.md`-Update
- gibt es eine Fitness Function, die das prüfen würde? → Modul 13, Gate hinzufügen

Diese Skill-Datei wird **nicht** überschrieben, sondern versioniert
(ADR-Hard-Rule, Modul 4).
