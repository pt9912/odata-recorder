# Slice slice-handbuch-katalog-schluessel-kommandos: Kataloge der Konfigurationsschlüssel und der Kommandos im Abgleich mit dem Handbuch

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md).

**Bezug:** `LH-QA-05`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-201` (Vertrag unter `harness/vertraege/`, Abgleich der Kataloge) Nr. 4, 5, 6, 7 (Art 4, 8, 10), 8, 10 a · `SPEC-200` (Handbuch gegen das gebaute Binary) · `SPEC-102` (Gegenproben) · `SPEC-110` Nr. 2 (Schlüssel dieser Version) · `SPEC-138` Nr. 2 (Kommandos des Adapters).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-handbuch-katalog-schluessel-kommandos`, Worktree `/Development/odata-recorder-wt-handbuch-katalog`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912 (Planner). **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make katalog-check` hält auch die Kataloge der Schlüssel und der Kommandos gegen das Handbuch (`SPEC-201` Nr. 4, 5): jeder Schlüssel, den die Startkonfiguration liest, und jedes Kommando des Binary steht im Handbuch (Abschnitte *Einstellungen* und *Kommandos*), und das Handbuch nennt keinen, den das Binary nicht kennt oder ablehnt.

**Übernimmt:** `slice-handbuch-meldungskatalog` (Rückgabe des Architect an den Planner, 2026-11-06: der Slice ist an der Grenze der Kataloge geschnitten; die Kataloge der Schlüssel und der Kommandos nach `SPEC-201` Nr. 4, 5, 6, 7 (Art 4, 8, 10), 8 und 10 a, die beiden Tabellen des Handbuchs, die Gegenproben zu ihren Befunden und der Nachzug von `SPEC-102` gehen hierher).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Programm `katalog`, Stufen `katalog-bin` und `katalog`, Ziel `make katalog-check`, Gate-Kette und der Katalog der Fehlercodes samt Ausnahmen-Datei — `slice-handbuch-meldungskatalog`; dieser Slice erweitert das Programm, er legt es nicht an (Start-Trigger, §4).
- Die Entscheidung der ADR-Frage zur Laufzeit ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) — der User, vorgelegt in `slice-handbuch-meldungskatalog` §6; dieser Slice misst die Zusatzzeit seiner Erweiterung und trägt sie in §7 nach.
- Bedeutung und Wahrheit der Zeilen der Tabellen (`SPEC-201` Nr. 9 a) — akzeptiertes Negativ des Vertrags; die belegten Beispiele (`slice-handbuch-record-replay`) und der Probelauf (`slice-handbuch-abnahme`) tragen es.
- Optionen des Binary, Namen hinter `Gesetzt (nichtleer) und nicht geliefert sind` in `SPEC-110` Nr. 2 und der Leser für `SERVICE_ODATA_VERSION` — der Vertrag führt keine Optionen (`SPEC-201` Nr. 5 b) und vergleicht die Menge N nicht (Nr. 9 c); den Leser liefert die Welle `welle-odata-v2` (`slice-v2-bootstrap`).
- Eine Änderung von `AGENTS.md` — nur als Diff-Vorschlag an den User (Pflicht der Folgewellen, `SPEC-201` Nr. 11), den schon `slice-handbuch-meldungskatalog` vorlegt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Kataloge im Programm `katalog`: Schlüssel (C aus `konfig*.go`, K aus `SPEC-110` Nr. 2, N der abgelehnten Namen, W aus der Tabelle *Einstellungen*) und Kommandos (C aus `cli.go`, W aus der Tabelle *Kommandos*) nach `SPEC-201` Nr. 4 bis 8 und 10 a; jede Zusage hat einen Fänger, vom Implementer rot gesehen (`LH-QA-05`, `LH-RB-01`).
- [ ] Handbuch: die Abschnitte *Einstellungen* (27 Schlüssel von `SPEC-110` Nr. 2, davon 26 im Code gelesen) und *Kommandos* (`start`, `version`) mit den Tabellen nach `SPEC-201`; die Regel „Jeder Schlüssel und jedes Kommando, das das Binary zeigt, kommt im Handbuch vor, und das Handbuch nennt nichts anderes“ ist durch Fänger über beide Richtungen und alle Elemente der Mengen gefangen (Schleife über die Mengen, `AGENTS.md` §3.15, §3.17): 26 Schlüssel, 6 abgelehnte Namen und Präfixe, 2 Kommandos.
- [ ] Gegenproben zu den Befunden von Schlüsseln und Kommandos laufen in `make gegenproben`; `SPEC-102` trägt ihre Fallnummern nach.
- [ ] Zusatzzeit gemessen (Gate und Gegenproben, Zahlen in §7).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `harness/README.md` (Sensors-Zeile von `make katalog-check` nennt alle drei Kataloge), Handbuch.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tools/katalog/` | update | Leser von `konfig*.go`, `SPEC-110` Nr. 2 und `cli.go`; Befunde Art 4, 8, 10 für Schlüssel und Kommandos |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` | update | Fälle zu Befunden von Schlüsseln und Kommandos |
| `harness/README.md` | update | Zeile *Sensors* von `make katalog-check` |
| `harness/vertraege/SPEC-102-gegenproben.md` | update | Fallnummern der neuen Fälle |
| `docs/user/benutzerhandbuch.md` | update | Abschnitte *Einstellungen* und *Kommandos* mit Tabellen |
| Tests | neu | Mengen aus dem Wortlaut: 26 Schlüssel, 6 Namen und Präfixe der Menge N, 2 Kommandos |

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis im Vertrag `SPEC-201` und in der Tabelle unten):

1. Tabelle *Zusage · Fänger* und Raster beim Start (Zusagen aus `SPEC-201` Nr. 4 bis 8 und 10 a). Hinweise des Planners: Mengen aus dem Wortlaut: 27 Schlüssel von `SPEC-110` Nr. 2 (heute 26 in `konfig*.go`, `SERVICE_ODATA_VERSION` fehlt dort, K ohne C ist kein Befund), 6 abgelehnte Namen und Präfixe (`ADMIN_`, `HYBRID_`, `FALLBACK_`, `LISTEN_TLS_`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`), 2 Kommandos (`start`, `version`); die Formen sind disjunkt (Schlüssel in Großbuchstaben, Kommandos in Kleinbuchstaben): ein Eintrag in zwei Katalogen mit gleichem Namen gilt nicht.
2. Der Fänger der Leseregel über die Literale in `konfig*.go` über alle Leser (`SPEC-201` Nr. 4 a, 9 e): Sendung von `slice-handbuch-meldungskatalog` (Risiko zur Tabelle der Schlüssel).
3. Größenmarke: Block `groessenmarke` (Teil des Architect) in §4; Hinweis des Planners: *S* grob 80·14 + 400 = 1520 (14 Zeilen), keine Schnittfrage.

**Tabelle *Zusage · Fänger*** (füllt der Architect beim Start):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write` (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-handbuch-meldungskatalog` liegt in `done/` (das Programm `katalog`, die Stufen, das Ziel `make katalog-check` und die Gate-Kette sind sein Gegenstand). Der Architect hat Tabelle und Raster gefüllt, bevor der erste Code-Commit entsteht.

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7, Planner, 2026-11-06): Dieser Slice ändert dieselben Dateien der Gate-Kette wie `slice-handbuch-meldungskatalog` (`Dockerfile`, dem Fragment `katalog.mk` unter `harness/mk/`, `harness/README.md`, `tools/gegenproben/`) und **startet nicht daneben, sondern erst nach dessen `done/`**. Kein weiterer Slice der Gate-Kette steht in `in-progress/`, solange dieser läuft. Kennungen aus dem Bereich der Linie C (`SPEC` 200 bis 229, `ADR` 0030 bis 0039, `CO` 040 bis 059): die nächste freie `SPEC`-Nummer der Linie C ist `SPEC-202`. Gate-Zeit: `make gates` und `make gegenproben` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Linien zugleich.

**Stand beim Start** (Planner, 2026-11-08, `main` = `e5c0f8b9`): erfüllt. `slice-handbuch-meldungskatalog` liegt in `done/`, das Programm `katalog`, die Stufen und `make katalog-check` stehen auf `main`. Nr. 7 der Übergangsregeln: kein anderer Slice der Gate-Kette in `in-progress/` (`slice-admin-rahmen` und `slice-mock-abnahme` ändern weder `GATE_CHECKS` noch ein Gate); `slice-admin-rahmen` ändert `harness/katalog-ausnahmen.txt` und `model.AlleCodes()`, Eingaben des Fehlercode-Katalogs, nicht die Kataloge der Schlüssel und Kommandos.

**Größenmarke** (Planner, 2026-11-08, `SPEC-128` Nr. 1; Prosa-Form, der Block folgt vom Architect und wird vom Planner in eigenem Commit vor dem ersten Code-Commit bestätigt): 14 Zeilen der Tabelle, *a* = 0, *b* = 14, *f* = 0, *p* = 400: *S* = 85·0 + 80·14 + 200·0 + 400 = 1520, *M* = 100·⌈135·1520/10000⌉ = 2100, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Architect legt den Block `groessenmarke` (Pfade, `produktcode`, `fakes`) beim Start an; ändert sich die Zeilenzahl der Tabelle, rechnet der Planner neu. `basis`, `schaetzung` und `marke` setzt der Planner.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt.
- `in-progress` → `open` (blockiert — Carveout?): der Vertrag `SPEC-201` Nr. 4 oder 5 verlangt eine Änderung von `konfig*.go` oder `cli.go`; das ist nach dem Vertrag nicht vorgesehen (der Abgleich liest, er ändert nichts).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Schlüssel stehen als Literale an vielen Stellen von `konfig*.go`, nicht in einer Tabelle (Sendung von `slice-handbuch-meldungskatalog`, Leseregel `SPEC-201` Nr. 4 a, 9 e) — **Ausgang:** offen; Fänger über alle Leser (Auftrag 2 des Architect).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster entsteht beim Start vor dem ersten Code-Commit, jede Zelle mit Ort in `SPEC-201` oder „gilt nicht“ mit Grund; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- Eine Allaussage des Vertrags („jeder Schlüssel in C steht in K“) ohne Fänger über alle Fälle (`BEO-REPO/universalaussage-im-vertrag-ungeprueft`) — **Ausgang:** offen; Schleifen über die Mengen (DoD, zweiter Punkt).
- Eine Welle macht einen Schlüssel oder ein Kommando erreichbar und ändert Ausnahmen und Handbuch nicht (Pflicht der Folgewellen, `SPEC-201` Nr. 11) — **Ausgang:** weiter offen (akzeptiertes Negativ des Vertrags, Nr. 9 c); Adresse: der User über den Diff-Vorschlag zu `AGENTS.md` aus `slice-handbuch-meldungskatalog`.

**Raster *Operation × Argumentart*** (füllt der Architect beim Start; Stand: alle `(Vertrag, offen)`):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Schlüssel lesen (C, K, N) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Kommandos lesen (C) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Handbuch lesen (Einstellungen, Kommandos) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Mengen vergleichen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

*Wird bei der Closure gefüllt.*

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-06): `BEO-REPO/randform-erst-im-review-entschieden` (`AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (`AGENTS.md` §3.15, §3.16) gelten; der Planner zieht die Zähler beim Start nach.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
