# Review slice-groessenformel-neu (Text-Slice, Beschluss V0)

Reviewer, frischer Kontext. Stand HEAD adb4978d, Basis ec1ab4ef. Diff: `harness/vertraege/SPEC-128-groessenmarke.md` (+13/-4), `docs/plan/planning/in-progress/slice-groessenformel-neu.md` (+204/-43); `git diff --numstat ec1ab4ef..adb4978d`. Keine andere Datei.
Gegen Plan (§1, §3, §4, §6), SPEC-128, AGENTS.md §3.6 bis §3.18, [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md), [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md), [MR-001](../../harness/conventions/MR-001-harness-vertraege-ausserhalb-der-spezifikation.md) und die Register-Einträge `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`, `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen`.

Alle Mutationen liefen auf Kopien im Scratchpad (`groessenformel-reviewer/arch`, `clone`, `m-*`); der Worktree ist unverändert (`git status` leer, HEAD adb4978d).

## Findings

### MEDIUM

**M1 — Verweis-Richtung für Werkzeugverträge falsch angenommen; Zeile 4a trägt einen Grund, der das Gegenteil sagt**
- quelle: [MR-001](../../harness/conventions/MR-001-harness-vertraege-ausserhalb-der-spezifikation.md) Nr. 5; AGENTS.md §3.15 (Negativ ohne tragenden Grund), §3.7 (Herkunft in ein auflösbares Feld)
- pfad: Plan · „Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie“ (Kopf); Plan · Tabellenzeile `| 4a |` („Richtung nach oben, `AGENTS.md` §3.4 für die Spec-Straten“; Grund „stehen nach `harness/README.md` und MR-001 neben den Straten“); Plan §3 „Auftrag an die Closure“ („trägt **kein** Feld `liegt in`“); Plan §6 letzter Punkt („`SPEC-128` darf den Slice nicht nennen“); SPEC-128 Nr. 2 „Stand 2026-11-20; Beschluss des Users“
- befund: MR-001 Nr. 5 sagt ausdrücklich, `harness/vertraege/` gehöre nicht zur Klasse `spec-straten` und ein Vertrag dürfe auf ADRs und Slice-Pläne zeigen; `.d-check.yml` Zeile 173 führt `spec-straten` nur für die drei Dateien unter `spec/`. AGENTS.md §3.4 gilt für `spec/architecture.md`. Das Verbot „SPEC-128 nennt den Slice nie“ hat damit keine Quelle; Zeile 4a begründet das fehlende Fänger-Paar mit MR-001, das die Zusage widerlegt. Folgen im Plan: Closure-Form „benannte Spec-Lücke ohne `liegt in`“ und der offene Punkt in §6 an den User sind auf dieser Annahme gebaut; im Vertrag steht die Herkunft der 62 Punkte als „Beschluss des Users“ und Datum, nicht als auflösbares Feld, und Nr. 9 sagt „die übrigen Punkte stehen in den Plänen“, ohne dass der Vertrag den Plan nennt, in dem die 62-Zeilen-Tabelle liegt (ein einziger Plan; 35 Punkte haben eine Messpunkt-Zeile, 27 stehen in Closure-Tabellen und Evidence).
- Mutation (auf Kopie, `make docs-check` gegen die Kopie): (a) `slice-groessenformel-neu` im Text von SPEC-128: `971 Datei(en) geprüft, 0 Befund(e)`; (b) Link auf den Plan in SPEC-128: 0 Befunde; zur Gegenprobe (c) toter Link: `target-missing`, (d) eine ADR-Kennung ohne Link: `id-unlinked`. Das Gate prüft Verträge also auf Links und Kennungen, und es lässt Slice-Verweise zu, wie MR-001 es sagt.
- verifizierbar: ja (die vier Läufe oben, `.d-check.yml` Zeile 173, MR-001 Nr. 5)
- klasse: Referenz-Richtung der Spec-Straten auf Werkzeugverträge übergeneralisiert (verwandt: Register-Eintrag `architect-text-mit-verweis-nach-unten-zwingt-implementer`, dort nur `spec/`)

### LOW

**L1 — Nr. 9: Kennzahlen der Auslastung rechnen mit neu gerechneter Marke, der Vertragstext sagt es dort nicht**
- quelle: AGENTS.md §3.15 (Zeile 7: Zahlen nachrechenbar), §3.11
- pfad: SPEC-128 · „von diesen 61 Punkten melden 7 (5 unter 0,45, 2 über 1,00); bei 0,60, dem Median, wären es 33“ und „(0,47 bei 61 Punkten)“; Nr. 9 definiert „Auslastung = Messung am Ende ÷ Marke“
- befund: Die Tabelle im Plan führt *M* nach Nr. 1 aus der *S* am Ende, bei sieben Plänen weicht sie von der Marke des Plans ab. Mit der Marke des Plans liegt `handbuch-geruest` bei 6956 ÷ 6900 = 1,008 und `bootstrap-m2` bei 3075 ÷ 2900 = 1,06 (Plan §3), beide über 1,00; die Zahl 7 gilt also nur für die neu gerechnete Marke. Nr. 2 nennt „Die Marke nach Nr. 1 hält 59 von 61“, Nr. 9 nicht. Die Entscheidung ändert sich nicht (mehr als 7, rund 15 %).
- verifizierbar: ja (Nachrechnen, Messpunkt-Zeilen der Pläne unter `done/`)
- klasse: Kennzahl ohne genannte Rechenbasis

**L2 — Nr. 10 a (3): Tempus behauptet ein Ereignis; Nachprüfpfad des Vertrags unvollständig**
- quelle: AGENTS.md §3.7 (beschreibt, was da ist), §3.15
- pfad: SPEC-128 · „ein systematischer Lesefehler zeigte sich beim Nachrechnen aus den Plänen“
- befund: Die Vorlage im Plan (§3 Teil 1, Punkt 7) sagt „wäre über die Tabelle oben nachprüfbar“; der Vertrag sagt „zeigte sich“ und liest sich als Bericht, ein solcher Fehler sei aufgetreten. Ich habe 21 Messpunkt-Zeilen aus `done/` gegen die Tabelle gelesen: Abweichungen allein durch Handoff-Stand und neu gerechnete Marke (siehe L1), kein Lesefehler. Die Tabelle, gegen die man nachrechnet, steht im Vertrag nicht.
- verifizierbar: nein
- klasse: Aussage über Vergangenes statt über den Zustand

**L3 — Median und Perzentile ohne Definition; einzelne Zahlen weichen bei üblicher Rechnung ab**
- quelle: AGENTS.md §3.15 (Zeile 7: der Reviewer rechnet nach)
- pfad: SPEC-128 · „18 Punkte mit Median 1,15 und Maximum 1,66“; „Nr. 1 1,88; … mit Zuschlag 1,40: 1,89“; Plan §3 · „0,75 (*S* unter 1500, 12 Punkte)“ und „0,71 (ab 6000, 8)“
- befund: Mit Rang-Perzentilen (nächster Rang, unterer Median bei gerader Zahl) stimmen alle Zahlen; mit dem üblichen Median bei gerader Zahl und linearer Interpolation ergeben sich: Median der 18 Punkte 1,157 (also 1,16), p75 der 20 Punkte 1,38 und p90 1,75 (statt 1,35 und 1,66), V1-Verhältnis p90 ÷ p10 = 1,877 (also 1,88, wie V0 mit 1,878), Mediane der Größenklassen 0,78 und 0,76 (im Plan 0,75 und 0,71). Die Definition steht nirgends; keine Aussage kippt.
- verifizierbar: ja
- klasse: Kennzahl ohne Definition des Verfahrens

**L4 — Plan §1 nennt die alte Schwelle als geltend**
- quelle: AGENTS.md §3.9 (wer eine Aussage berichtigt, zieht jede Stelle nach)
- pfad: Plan · „unter der Schwelle 0,60 von `SPEC-128` Nr. 9“ (Absatz *Messpunkte nach den 14*)
- befund: Der Satz steht in der Gegenwart und nennt eine Schwelle, die Nr. 9 nicht mehr trägt; der Nachzug-Beleg in §3 zählt die Treffer in diesem Plan als „Zahlen des Vorschlags, Messpunkte“ und lässt diese Stelle stehen.
- verifizierbar: ja (`grep`)
- klasse: Berichtigung erreicht nicht jede Stelle

### INFO

- **I1 (Zeile 1, äquivalente Mutanten):** `zuschlagProzent*s > schnittGrenze` auf `>=` bleibt grün; zu Recht, denn 135·*S* = 300000 hat keine ganzzahlige Lösung. Dasselbe gilt für `schnittGrenze` 300100 (135·2222 = 299970, 135·2223 = 300105). Kein Mangel des Fängers.
- **I2 (§3.6):** Verdikt „greift nicht“ bestätigt. `grep` nach `Auslastung`, `0,60`, `0,45` in `tools/` (Code und Tests), `harness/mk/`, `.claude/`, `.harness/skills/`, `AGENTS.md`, `harness/README.md`, `spec/`: kein Treffer außer SPEC-128, Plan und Register-Evidence. Die Marke wird nicht lockerer (Nr. 1 byte-gleich: die Hunks des Diffs beginnen bei Nr. 2). [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) trägt keine Zahl der Formel oder Schwelle. Sachlich sinkt die Zahl der Meldungen von 33 auf 7 von 61, also die Rückkopplung der Neumessung; das ist die dokumentierte Entscheidung des Users, kein Gate.
- **I3 (Geltungsgrenze):** Satz 1 („überwiegend Test mit Code“) und Satz 2 (Handläufe, Gate-Ausgaben, Doku-Gate) sind nicht komplementär: Zeilen mit Diff oder akzeptiertem Negativ als Fänger (Zeilen 2, 4a, 5, 6, 7 dieses Plans) stehen in keiner der drei Klassen von Satz 2. Die Entscheidung „kein Block“ trägt Satz 1 (2 von 8 Zeilen), Nr. 3 m/4 a und Nr. 10 k; sie ist richtig, der Wortlaut lässt den gemischten Fall dem Urteil des Planner (Raster Zeile 2).
- **I4 (Rundung):** Wie die Auslastung vor dem Vergleich mit 0,45 und 1,00 gerundet wird, steht nicht in Nr. 9; das Raster verweist für „genau 0,45“ auf den Wortlaut „unter“. Besteht seit der Fassung mit 0,60.
- **I5 (Nr. 10 d):** `abfrage` gegen den Plan (Image aus HEAD, `--network none`, Mount `:ro`): `keine-messung`, Hinweis `kein Block \`groessenmarke\`, der Plan wird nicht gemessen`, Exit 0; die Tabelle des Plans wird dabei fehlerfrei gelesen (ein Eingabefehler der Tabelle hätte Exit 2 gegeben).

## Auflösung der Tabelle *Zusage · Fänger* (Plan §3)

| Zeile | Mutation (Kopie) | Ergebnis |
|---|---|---|
| 1, genannt | `zuschlagProzent` 135 auf 140 | `TestMarkeRundung` rot (dazu `TestSchnittfrageGrenze` und 15 weitere) |
| 1, nicht genannt | `faktorNebenlaeufig` 85 auf 80 | `TestSchaetzungGitter` rot (dazu `TestMesspunkte`, `TestMesspunktZeile`) |
| 1, nicht genannt | `faktorAndere` 80 auf 85 | `TestSchaetzungGitter` rot (und 11 weitere) |
| 1, nicht genannt | `faktorFake` 200 auf 190 | `TestSchaetzungGitter` rot (dazu `TestMesspunkte`, `TestMesspunktZeile`) |
| 1, nicht genannt | `schnittGrenze` 300000 auf 310000 | `TestSchnittfrageGrenze` rot (und 5 weitere) |
| 1, nicht genannt | Aufrundung `+ 9999` auf `+ 9900` und auf `+ 10000` | `TestMarkeRundung` rot, je Mutation |
| 1, nicht genannt | Zuschlag 134 und 136 | `TestMarkeRundung` und `TestSchnittfrageGrenze` rot |
| 1, nicht genannt | Rundung auf 50 statt 100 | `TestMarkeRundung` und `TestSchnittfrageGrenze` rot |
| 1, nicht genannt | `+ p` auf `+ p + p/1500` | `TestSchaetzungGitter` rot |
| 1 | `>` auf `>=` in `schnittfrage` | grün, äquivalent (I1) |
| 2 | `zuschlagProzent` 140 in `formel.go`, committet in Klon | `git diff --stat ec1ab4ef..HEAD` über `EINGABEN`: 1 Datei; am Worktree und nach dem Zurücksetzen: 0 Zeilen |
| 3 | Zeile 84 `hinweis(… "kein Block …")` entfällt | `abfrage Hinweis kein Block` rot („Hinweis: kein Block“ fehlt im Strom stderr; dazu `TestZustand`, `TestM3ProsaZeile`) |
| 3 | zusätzlich `return exitEingabe` nach dem Hinweis | rot (`TestZustand`, `TestM10StroemeDesProgramms`, `TestM3ProsaZeile`) |
| 4 | NBSP aus Bytewerten (`sed 's/…/\xc2\xa0…/'`, gegengelesen mit `LC_ALL=C grep -P '\xC2\xA0'`: 0 vorher, 1 nachher) in SPEC-128 | `make zeichen-check` Exit 2: `SPEC-128-groessenmarke.md:125:11: U+00A0 Zs NBSP`; unmutiert grün; nach `git checkout --` 0 Treffer |
| 4a | Name und Link auf den Plan in SPEC-128 | `make docs-check` grün (Gate fängt nicht, wie der Plan sagt); das Verbot selbst ist unbegründet (M1) |
| 5, 6 | kein Fänger | akzeptiertes Negativ in Nr. 10 k mit Grund und Folge; Maß fehlt in Nr. 10 k (INFO: keine Zahl der betroffenen Slices, Plan Teil 1 nannte „2 von 63“) |
| 7 | Nachrechnen (siehe unten) | Zahlen stimmen bis auf L1 und L3 |

Zeile 7, nachgerechnet aus den 62 Zeilen der Tabelle (awk, auf Kopie): *a* = (*S* − 80*z* − 200*f* − *p*) ÷ 5 ist in allen 62 Zeilen eine nichtnegative ganze Zahl; *M* = 100·⌈135·*S*/10000⌉, *E* ÷ *S*, *E* ÷ *M* und *E* ÷ *H* der Tabelle stimmen in allen Zeilen. Kennzahlen n = 61: *E* ÷ *S* Median 0,826, p10 0,642, p90 1,205; *E* ÷ *M* Median 0,599, p10 0,472, p90 0,853; 49 von 61 unter der Schätzung; 59 von 61 unter der Marke (`telemetry-port` 1,031, `bootstrap-m4-konfig` 1,921); unter 0,60: 31, unter 0,50: 8, unter 0,45: 5, unter 0,40: 3; Meldungen bei 0,45: 5 + 2 = 7, bei 0,60: 31 + 2 = 33; *E* ÷ *H* bei 20 Paaren Median 1,16, p25 1,09, p75 1,35, p90 1,66, Maximum 3,09; ohne `bootstrap-m4-konfig` und `gates-ohne-worktrees` Maximum 1,66; Handoff ab 0,70 der Marke: zwei Paare, eines reißt. V1 hält an 32 jüngeren Punkten 27 (Marke auf Hundert gerundet), die an *alt* angepasste Form 26, V0 31; V2 hält 30. 62 Punkte: 60 Go, 1 Skript, 1 Handlauf; *alt* 29, *neu* 32 (ohne Handlauf).

## Entscheidung „kein Block“ gegen SPEC-128

Nr. 3 m: `spec/`, `docs/`, `harness/vertraege/` sind als Pfad unzulässig; Nr. 4 a schließt `harness/vertraege/` immer aus der Messung aus; jeder zulässige Pfad misst für diesen Plan 0 (Block bliebe nach Nr. 5 a „vor dem ersten Code-Commit“). Nr. 2 *Geltungsgrenze* (2 von 8 Zeilen mit Test), Nr. 3 p, Nr. 8 (grün ohne Block), Nr. 10 k: konsistent. Plan enthält weder eine Zeile `**Größenmarke**` noch einen Zaun `groessenmarke` (Zustand `kein Block`, nicht Prosa-Form). Bestätigt.

## Rand und Beleg-Fragen des Auftrags

- Zahlen in Nr. 2 und 9 gegen die Tabelle: übereinstimmend (62/61 Punkte, Median 0,83, p10 0,64, p90 1,20, 0,47, 7 statt 33, 1,16, 1,66) bis auf L1 und L3.
- Vertrag nennt den Slice nicht: zutreffend (kein Treffer von `slice-groessenformel` in SPEC-128); ein Verbot dafür besteht nicht (M1).
- Verweise lösen auf: `make docs-check` am Worktree `971 Datei(en) geprüft, 0 Befund(e)`; `make zeichen-check` `ok, 2006 Datei(en)`; Rohbytes beider Dateien: 0 Treffer für BOM, NBSP, Nullbreite, Zeilentrenner, Backslash-u.
- Vier akzeptierte Negative (Nr. 10 a 1 bis 4): Grund und Folge je Punkt, Maß bei 1 (p10 bis p90, 1 von 30), 2 (drei Punkte), 3 (20 Paare, 10 zurückgerechnet); bei 4 ohne Zahl (Wellen, ein Architect), Folge „Größenordnungen“.
- Plan §3.9/§3.13: Kopf (`Bezug`, `Berührte Spec-Stellen`), §1, §3, §6 tragen den Beschluss; offene Punkte stehen in §6 mit Ausgang *offen* und Adresse (Reviewer, User, Planner/Integrationsschritt, Planner/Closure); Roadmap und Register nach Nr. 4 der Übergangsregeln nicht im Zweig.

## Geprüft, ohne Befund (Negativbefund-Zeilen)

- `harness/vertraege/SPEC-128-groessenmarke.md`: Nr. 1 byte-gleich (Hunks ab Nr. 2); Zahlen von Nr. 2, 9, 10 a, 10 k gegen die Tabelle nachgerechnet; Befunde M1, L1 bis L3, I3, I4.
- `docs/plan/planning/in-progress/slice-groessenformel-neu.md`: DoD, §1, §3, §4, §6 gegen SPEC-128 und AGENTS.md gelesen; Befunde M1, L4.
- `tools/groessenmarke/`, `tools/gegenproben/`, `tools/harness/groessenmarke.sh`, `Dockerfile`, `.dockerignore`, `harness/mk/`, `harness/gegenproben.stempel`: im Diff nicht berührt, `git diff --stat` über `EINGABEN` leer; geprüft, ohne Befund.
- `spec/`, `AGENTS.md`, `harness/README.md`, `.claude/`, `.harness/skills/`: im Diff nicht berührt, keine Zahl der Schwelle; geprüft, ohne Befund.
- [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md), [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md): nicht berührt; [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) trägt keine Zahl der Formel oder der Schwelle; [ADR-0010](../plan/adr/0010-groesse-der-guide-dateien-als-gate.md) betrifft Guide-Dateien, SPEC-128 gehört nicht dazu; geprüft, ohne Befund.
- `docs/plan/planning/observations/` (beide Einträge): nicht im Diff; `state.md` nennt noch „Beschluss steht aus“, in §6 des Plans mit Adresse geführt; geprüft, ohne Befund am Diff.
