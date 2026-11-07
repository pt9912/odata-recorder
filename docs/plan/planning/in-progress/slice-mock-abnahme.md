# Slice slice-mock-abnahme: Abnahmetests des Mock: Matrix, Determinismus, Isolation, Neustart

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-QA-02`, `LH-QA-03`, `LH-RB-01`, `LH-FA-08`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-QA-02.a` · `LH-QA-03.a` · `LH-RB-01.a` · `SPEC-049` (Abnahmematrix) · `SPEC-048` (Fault Injection, nur Neustart-Teil) · `architecture.md` §6 (Qualitätssicherung).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-mock-abnahme`, Worktree `/Development/odata-recorder-wt-mock-abnahme`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Tests in `tests/` belegen den Mock gegen die Zusagen des Lastenhefts, die kein Slice allein trägt: die Muss-Zeilen der Abnahmematrix (`SPEC-049`, `ODV4-*`) als Golden-Fixtures, 100 identische Läufe nach Reset mit gleichen normalisierten Antworten, 20 parallele Workspaces ohne Beobachtung fremden Zustands und Cursors, und ein regulärer Neustart, nach dem abgeschlossene Mutationen und Snapshots vorhanden sind.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Soll-Zeilen der Matrix (`Prefer`, `If-Match`, `412`, zusätzliche Funktionen) — `welle-etag` und `welle-expand-batch`.
- Die Fault-Injection-Matrix der Persistenz (`SPEC-048`) — Bestand von `welle-persistenz`; hier prüft nur der Neustart-Test die Folge für Mutationen und Snapshots.
- Der Lasttest zu `LH-QA-01` — `welle-betrieb`; er braucht die Spans.
- Neue Produktfunktion — der Slice ändert `src/` nur, wo ein Test einen Fehler findet, und meldet ihn dann an den Slice, dem die Zusage gehört.

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

- [ ] Abnahmematrix: die Muss-Zeilen von `SPEC-049` (`ODV4-GET`, `-CRUD`, `-Q`, `-T`, `-H`, `-E`) als Golden-Fixtures, je Fixture-ID ein Test (`LH-RB-01`).
- [ ] Determinismus: 100 identische Läufe nach Reset ergeben gleiche normalisierte Antworten (`LH-QA-02`), Vergleich ohne `VOLATILE_HEADERS`.
- [ ] Isolation und Neustart: mindestens 20 parallele Workspaces beobachten keinen fremden Zustand und keine fremden Cursor; nach regulärem Neustart sind abgeschlossene Mutationen und Snapshots vorhanden (`LH-QA-03`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `tests/` | neu | Golden-Fixtures der Matrix, Läufe über 100 Wiederholungen, Paralleltest, Neustarttest; Pakete im Quellstand der Stufe `source`, damit `make test-race` sie fährt (Muster von `slice-bootstrap-m2`) |
| `spec/spezifikation.md` | update | Fixture-IDs von `SPEC-049` als Testadressen mit Inhalt (Architect) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster entstehen mit der Festlegung der Fixtures in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Fixtures sind „geplante Testadressen“ (`SPEC-049`) und werden vom Architect festgelegt, nicht vom Implementer.

- Die Läufe sind Teil von `make test`: wie lange die 100 Läufe dauern dürfen, ist eine Randform des Vertrags (§6).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-snapshot-use-case` liegt in `done/`. Die Spezifikation trägt die Fixtures der Matrix `SPEC-049` mit Inhalt, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand beim Start** (Planner, 2026-11-07, `main` = `08843060`): Trigger Teil 1 erfüllt, `slice-snapshot-use-case` liegt seit dem Merge `56b87d09` in `done/`. Trigger Teil 2 (Fixtures der Matrix `SPEC-049` mit Inhalt in der Spezifikation) **nicht erfüllt**: `SPEC-049` führt die Adressen `ODV4-GET-01..03`, `-CRUD-01..06`, `-Q-01..12`, `-T-01..08`, `-H-01..03`, `-E-01..04` als „geplante Testadressen“ ohne Inhalt; er ist Bedingung für den ersten Code-Commit, nicht für den Start, und liegt beim Architect (Aufträge M1 bis M7). §1 *Ausdrücklich NICHT* und DoD der Sendungen von `slice-snapshot-use-case` gelesen (`AGENTS.md` §3.13): `welle-mock` §5 und `done/slice-snapshot-use-case.md` §7 weisen diesem Slice allein die Bedingung „Trigger in `done/`“ und den Prüfgegenstand Reset-Determinismus, Isolation und Neustart mit Snapshots zu; kein Ausschluss in §1 dieses Slice trifft sie (Snapshots und Restore sind Gegenstand der Neustart-Zeile, nicht ausgeschlossen). Codes: `model.AlleCodes()` zählt auf `main` 39; dieser Slice fügt keinen hinzu, solange kein Test einen Fehler findet (§1, vierter Punkt); `slice-admin-rahmen` (40. Code) und `slice-sqlite-interrupt-transaktion` mergen nach `main`, wer als zweiter merged, gleicht die Zahl an (`welle-mock` §5, Merge-Nähe). Gate-Kette (Nr. 7 der Übergangsregeln): `tests/` ist neu, `Makefile`, `harness/mk/`, `Dockerfile` und Stempel bleiben unberührt; braucht `tests/` im Build-Kontext eine Änderung des `Dockerfile`, geht das als Rückgabe an den Planner. Speicher beim Start: 14 GB verfügbar, Load 13.

**Größenmarke** (Planner, 2026-11-07; Prosa-Form, kein Block): **nicht gesetzt.** `SPEC-128` Nr. 1 rechnet auf die Tabelle *Zusage · Fänger*, und diese entsteht erst mit den Fixtures (M1); eine Zahl vor der Tabelle wäre geraten (§4 Rückführungen). Der Planner setzt Marke und Block (`basis`, `schaetzung`, `marke`, Pfade aus §3) in eigenem Commit nach der Tabelle und vor dem ersten Code-Commit und beantwortet dann die Schnittfrage.

**Aufträge an den Architect zum Start** (Planner, 2026-11-07, `main` = `08843060`; vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Architect arbeitet im Zweig `slice-mock-abnahme`; Ausgang jeweils *offen* in §6; Namen M1 bis M7):

- **M1. Fixtures der Matrix `SPEC-049` mit Inhalt.** Für jede Adresse der Muss-Zeilen (`ODV4-GET-01..03`, `-CRUD-01..06`, `-Q-01..12`, `-T-01..08`, `-H-01..03`, `-E-01..04`: 36 Adressen) legt er Request, erwartete Antwort (Status, Header, Body) und Ausgangsdaten fest; dazu die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) mit den Spalten, die `SPEC-128` Nr. 3 liest, und die Mutation je Zeile. Die Soll-Zeilen bleiben draußen (§1).
- **M2. Determinismus.** Was „normalisiert“ ist (`VOLATILE_HEADERS` und weitere flüchtige Felder), was ein Lauf umfasst (Reset, Anfragenfolge, Vergleichsform) und das Zeitbudget der 100 Läufe in `make test` als Randform mit Ort.
- **M3. Isolation.** Welche Beobachtungen „fremder Zustand“ und „fremder Cursor“ heißen (lesen, mutieren, Reset, Snapshot, Restore), die Menge der 20 Workspaces mit Zählprüfung, und ob Cursor im Mock-Modus überhaupt vorkommen (`LH-QA-03.a`).
- **M4. Neustart.** Im Test-Binary oder als Prozess mit `SIGTERM` (`LH-FA-02.c`); was „abgeschlossen“ bei Mutationen heißt und was nach dem Neustart zu sehen ist (Mutationen, Snapshots, Restore). Die Fault-Injection-Matrix (`SPEC-048`) bleibt draußen.
- **M5. Raster und Zusatztabelle (`AGENTS.md` §3.16)** für Reset, Mutation, Snapshot, Restore und Neustart: je Zelle ein Ort oder „gilt nicht“ mit Grund.
- **M6. Nebenläufigkeit.** Fänger der Zusagen über 20 Workspaces mit Wirkungs-Fänger oder benanntem Negativ (`AGENTS.md` §3.10; Register `rate-ohne-detektor-schwankt-mit-der-last`, `zweig-ohne-lauf-unter-dem-detektor`), Zweige mit Zählprüfung unter `-race -coverprofile`, Raten mit Zahl der Prozesse, Last und Messendem.
- **M7. Pfade, Merge-Nähe, Reports.** Pfade des Blocks (§3: `tests/`, `spec/spezifikation.md`; weitere nur mit Meldung an den Planner) und die Faktoren *a*, *b*, *f*, *p* für die Marke; in `spec/spezifikation.md` nur die Stellen von `SPEC-049` und was sie berühren, kein Umstellen daneben. Reports von Reviewer und Verifier tragen ADR-Kennungen nie nackt, sondern als auflösenden Link (Register `report-der-rolle-faellt-docs-check`); der Implementer schreibt Repo-Dateien nur mit `Edit` und `Write` (Register `betriebsregel-fuer-repo-dateien-verletzt`, 13×).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt; die Gruppe der Matrix lässt sich vorab von den drei übrigen Gruppen trennen (kein geteilter Mechanismus), wenn die Schätzung × 1,25 über 3000 liegt.
- `in-progress` → `open` (blockiert — Carveout?): die Fixtures der Matrix sind vom Architect nicht festgelegt (siehe §6).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Fixture-IDs `ODV4-*` sind „geplante Testadressen“ ohne Inhalt (`SPEC-049`) — **Ausgang:** offen; Adresse: Architect (M1); er legt Inhalt und Zahl der Fixtures vor dem ersten Code-Commit fest, sonst entscheidet der Implementer, was „abgenommen“ heißt.
- Laufzeit der 100 Läufe und des Paralleltests in `make test` und die Frist der Gegenproben, die Pakete des Produkts mitfahren (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`) — **Ausgang:** offen; Adresse: Architect (M2), Zeitbudget als Randform im Vertrag.
- Raten der nebenläufigen Fänger (20 Workspaces) — **Ausgang:** offen; Adresse: Architect (M6) und Implementer; Rate je Lauf mit Zahl der Prozesse und Messendem (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`).
- Der Neustarttest: im Test-Binary oder als Prozess (`SIGTERM`, `LH-FA-02.c`) — **Ausgang:** offen; Adresse: Architect (M4).
- Aufträge M1 bis M7 an den Architect zum Start (§4): Fixtures mit Inhalt und Tabelle, Determinismus, Isolation, Neustart, Raster, Nebenläufigkeit, Pfade und Reports — **Ausgang:** offen; Adresse: Architect (Bestätigung oder Ergänzung in diesem Abschnitt), Implementer (Test).
- Größe: die Marke ist nicht gesetzt (§4), weil die Tabelle fehlt — **Ausgang:** offen; Adresse: Planner nach M1 (Block mit `basis`, `schaetzung`, `marke` in einem Commit vor dem ersten Code-Commit).
- Gate-Kette (Nr. 7 der Übergangsregeln): ändert `tests/` den Build-Kontext (`Dockerfile`, `.dockerignore`), ist das ein Pfad der Gate-Kette — **Ausgang:** offen; Adresse: Planner (Rückgabe, nicht im Code entscheiden).
- Zahl der Codes: ein Fehler, den ein Test findet, kann einen Code verlangen, und `slice-admin-rahmen` und `slice-sqlite-interrupt-transaktion` mergen anders auf `main` — **Ausgang:** offen; Adresse: Planner (Integrationsschritt, `welle-mock` §5).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-07, `main` = `08843060`, Zähler aus den Dateien unter `evidence/`). Einschlägig: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (8×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (16×) für den Paralleltest, `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (5×) für die Laufzeit der Läufe in der Gate-Kette, `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (13×), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (48×), `BEO-REPO/randform-erst-im-review-entschieden` (34×), `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (26×) und `BEO-REPO/report-der-rolle-faellt-docs-check` (5×). Verkörpert: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×, `AGENTS.md` §3.10). Das Register führt sie als offen oder verkörpert nach eigenem Stand; dieser Start ändert keinen Zustand.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
