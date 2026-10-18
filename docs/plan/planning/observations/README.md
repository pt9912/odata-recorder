# Beobachtungs-Register

Regeln dieser Ablage: Baseline-Regelwerk `modul-06-roadmap.md` §Das
Beobachtungs-Register.

## Form

Eine Beobachtung ist ein Verzeichnis, keine Tabellenzeile:

```text
BEO-<KUERZEL>/<slug>/
├── observation.md            die Identität — einmal geschrieben
├── state.md                  der veränderliche Stand
└── evidence/<vorgangs-id>.md je Auftreten eine Datei
```

`<KUERZEL>` wird nachgeschlagen, nicht erfunden: es ist das Sub-Area-Kürzel
aus der Modus-Deklaration in `harness/conventions.md`. Diese Spalte trägt
jedes Repo für diesen Zweck — unabhängig davon, ob seine ADR- oder
Slice-Kennungen selbst ein Bereichssegment führen: die Kennung einer
Beobachtung braucht das Kürzel so oder so. `<slug>` ist lowercase
Kebab-Case. Die Kennung der Beobachtung ist der Pfad `BEO-<KUERZEL>/<slug>`
selbst.

## Wer schreibt

Eingetragen wird bei der Slice-Closure: ein neues Verzeichnis für eine neue
Beobachtung, oder eine weitere Datei im vorhandenen `evidence/` für ein
Wiederauftreten. Erhöht wird nichts — der Zähler ist die Zahl der
Evidence-Dateien und folgt aus ihnen.

## Wer liest

An zwei Stellen: Die Welle-Closure liest, was 3× erreicht hat
(Lese-Schritt) — ohne Wellen-Betrieb löst die jeweilige Slice-Closure diesen
Schritt selbst aus. Die Slice-Planung liest, was darunter liegt
(Sichtungs-Schritt).

## Beleg-Form

`evidence/<vorgangs-id>.md` — der Dateiname ist die Kennung eines
abgeschlossenen Vorgangs (Regelfall: eine Slice-ID), kein Freitext. Ein
Vorgang zählt einmal; das erzwingt das Dateisystem, nicht die Disziplin.

## Die drei Ausgänge

| Ausgang | Was dazugehört |
|---|---|
| `verkörpert` | Zielort **und** Herkunfts-Anker |
| `geplant` | Kennung des Slice oder der Welle, die die Regel schreibt |
| `gestrichen` | die Begründung, warum die Beobachtung nicht mehr auftreten kann |

`verkörpert` und `geplant` werden ab 3× zugewiesen. `gestrichen` ist an
diese Schwelle nicht gebunden: fällt die Ursache vorher weg, wandert die
Zeile mit Begründung dorthin, unabhängig vom Zähler. Unterhalb der Schwelle
ist `offen` der Normalzustand, kein Ausgang. `gestrichen` heißt nicht
gelöscht: das Verzeichnis bleibt liegen, mit der Begründung in seinem
`state.md`.

## Eine leere Ablage

Trägt dieses Verzeichnis nur diese Datei, ist noch nichts beobachtet — nicht:
das Register wurde nie geführt. Git trackt kein leeres Verzeichnis; diese
Datei ist der Unterschied zwischen den beiden Aussagen.
