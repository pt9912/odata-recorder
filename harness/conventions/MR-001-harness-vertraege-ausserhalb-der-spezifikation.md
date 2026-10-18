# MR-001 — Verträge der Harness-Werkzeuge im Ordner harness, nicht in der Produktspezifikation

Regeln dieser Datei: Pflichtfelder sind Datum, Geltungsbereich,
**Ersetzt-Baseline-Regel**, Adaption, Begründung und Auflösungs-Trigger;
`Löst auf` und `Ausgelöst durch Baseline-Stand` nur, wenn dieser Eintrag einen
früheren ablöst. `Ersetzt-Baseline-Regel` nennt **genau eine** Regel der
Baseline, an deren Stelle dieser Eintrag tritt — als Link mit
Abschnitts-Anker in die vendored Fassung; ein Datei-Link benennt keine Regel.
**Dieser Link trägt zwei Dinge, die sich bewegen, und für beide gibt es einen
Wächter statt einer Formregel:** die *Tiefe* (nach dem `git mv` nach `done/`
zeigt der relative Pfad eine Ebene zu hoch — der Umzug zieht die
Pfad-Berichtigung nach sich, als eigener Commit; die Existenzprüfung des Links
meldet sie, wenn sie ausbleibt) und die *Version* (jeder Baseline-Bump
entwertet `<tag>` — der adoptierte Stand steht einmal im Adaptions-Block, ein
Versions-Sensor prüft jeden Pin dagegen; Muster in `.d-check.yml`).
Wer mehrere Regeln ersetzen will, schreibt mehrere Einträge. Ein Eintrag, der
keine benannte Regel ersetzt, ist ein **Fork**, keine Adaption.

- **Datum:** 2026-10-18
- **Geltungsbereich:** `harness/vertraege/` und `spec/spezifikation.md`, dazu `harness/README.md` §Source precedence und `AGENTS.md` §Kanonische Quellen
- **Ersetzt-Baseline-Regel:** [`grundlagen-referenz-richtung.md` §Spec-Straten](../../.harness/baseline/v6.18.1/regelwerk/grundlagen-referenz-richtung.md#spec-straten-mehr-als-ein-spec-dokument) — der Absatz *Auch die Werkzeuge des Harness treffen technische Festlegungen*, der als Ort dieser Festlegungen die Spezifikation nennt
- **Adaption:**
  1. **Ort.** Die Festlegungen eines Harness-Werkzeugs stehen in `harness/vertraege/`, eine Datei je Kennung, `SPEC-<NNN>-<kurztitel>.md`; `spec/spezifikation.md` führt keine.
  2. **Stratum.** Es bleibt das Technik-Stratum: fortschreibbar ohne Change Request, das Lastenheft ist die Decke, eine Festlegung erweitert es nie.
  3. **Kennung.** `SPEC-<NNN>` bleibt; die Reihe zählt über beide Orte fortlaufend.
  4. **Rang.** Rang 2 der Source Precedence trägt beide Orte: `spec/spezifikation.md` für das Produkt, `harness/vertraege/` für die Werkzeuge.
  5. **Referenz-Richtung.** `harness/vertraege/` gehört nicht zur Klasse `spec-straten`; ein Vertrag darf auf ADRs und Slice-Pläne zeigen, die Spezifikation nennt keinen Vertrag mit Link.
- **Begründung:** Die Produktspezifikation legt nur das Produkt fest. Werkzeugverträge wachsen mit jedem Harness-Slice und haben andere Leser (Implementer, Reviewer eines Gates) als die Produktspezifikation.
- **Auflösungs-Trigger:** Die Baseline nennt einen Ort außerhalb der Produktspezifikation für Werkzeugverträge, oder das Repo führt wieder einen Abschnitt für sie in `spec/spezifikation.md` ein.
