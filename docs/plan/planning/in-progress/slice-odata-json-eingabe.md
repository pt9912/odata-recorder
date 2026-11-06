# Slice slice-odata-json-eingabe: OData-Codec zum Lesen von Client-Bodies für POST, PATCH und PUT

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-07`, `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-07.a` · `LH-FA-07.b` · `LH-FA-08.b` (Typen) · `SPEC-044` (fehlend und `null`) · `SPEC-023` · `SPEC-126` Nr. 10 (Schreibweisen der acht Typen) · `ARC-004`, `ARC-011`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-odata-json-eingabe)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Schnittstelle der Codec-Gruppe zum Lesen von Client-Bodies überführt den JSON-Body eines POST, PATCH oder PUT in eine typisierte Entity-Eingabe gegen das Modell: acht Typen, `null` und fehlende Property bleiben unterscheidbar, eine unbekannte Property ergibt `400`, eine Property nicht unterstützten Typs `501` mit `ODR-E6002`, eine fehlende erforderliche Property bei POST und PUT `400`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Fachregeln wie Duplicate Key (`409`), Existenz der Entity und Statuscodes — `slice-mock-schreiben`; der Codec liest Bytes und entscheidet nichts über den Zustand.
- Das Lesen von OData-Antworten für den Import — vorhanden als `ODataCodec` (`SPEC-126`); dieser Slice liest Requests, deren Gestalt (Teilmengen bei PATCH, keine `@odata.context`) eine andere ist.
- Deep Insert und Navigationen im Body — nicht im MVP (`SPEC-049`).

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

- [ ] Typisierte Eingabe gegen das Modell: acht Typen, `null` und fehlende Property unterscheidbar, unbekannte Property `400`.
- [ ] Nicht unterstützter Typ `501` mit `ODR-E6002`; fehlende erforderliche Property bei POST und PUT `400`.
- [ ] Randformen des Bodys (kein Objekt, leerer Body, doppelte Property, falscher `Content-Type`, Größe über der Grenze): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger.
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
| `src/hexagon/ports/outbound/` | update | Schnittstelle der Codec-Gruppe zum Lesen von Client-Bodies |
| `src/adapters/driven/odata/jsoncodec/` | update | Leser der Entity-Eingabe |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-07.a` je Typ und je Fehlerklasse; Generator über die Klasse der Body-Formen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Leser liest fremde Bytes: die Produktcode-Größe des Architect ist ein fester Anteil der Schätzung, einen Zuschlag für Leser fremder Bytes kennt `SPEC-128` Nr. 1 nicht mehr (`AGENTS.md` §3.14); Bodies mit Escape-Folgen entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Aufträge an den Architect** (Start, Planner, 2026-11-06; vor dem ersten Code-Commit, `AGENTS.md` §3.12; Ausgang jeweils *offen* in §6):

1. **Vertrag** in `spec/spezifikation.md` aus dem Kennungsbereich der Linie B (`SPEC-180` bis `SPEC-189` sind für diesen Slice bestimmt, `SPEC-190` und folgende trägt schon `slice-query-interpreter` im Zweig; `SPEC-160` bis `SPEC-179` sind vergeben; `ARC` ab 030, ADR ab 0020; Übergangsregeln Nr. 6) mit Schnittstelle (Port in `src/hexagon/ports/outbound/`, Anbindung an `CodecSatz` neben `Leser` und `Schreiber` aus `SPEC-168` Nr. 10, Eingabetyp der Entity mit der Unterscheidung fehlend / `null` / Wert nach `SPEC-044`), Fehlerklassen (`400`, `501` mit `ODR-E6002`, Abbildung auf Codes der Tabelle der Fehlercodes; ein neuer Code geht als Rückgabe an den Planner, `welle-mock` §4 Handbuch-Pflicht) und den Randformen aus §6.
2. **Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15, Spalten nach `SPEC-128` Nr. 3 a) und **Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Operationen mindestens POST, PATCH, PUT; Argumentarten: Body fehlend, leer, kein Objekt, unbekannte Property, doppelte Property, `null` bei nicht nullbarer Property, falscher Typ, falscher `Content-Type`, zu groß, Abbruch des `ctx`, Wiederholung) mit Ort je Zelle.
3. **Entscheidungen zu den Risiken in §6**: Grenze für die Größe eines Request-Bodys im Mock (Herkunft der Zahl, Konfigurationsschlüssel oder fester Wert; ein neuer Schlüssel ist ein Handbuch-Fall und geht an den Planner), `@odata.*` im Body (ignorieren oder abweisen, `@odata.context` gesondert), doppelte Property, Zahlen als Zeichenkette für `Edm.Int64` und `Edm.Decimal`, Zeitpunkte und `Edm.Guid` in den Schreibweisen von `SPEC-126` Nr. 10.
4. **Pflicht-Menge-Tabelle** (`AGENTS.md` §3.17, `implement-slice` Schritt 14): je Klasse der Spezifikation (acht Typen × Wertformen, Body-Formen als JSON-Wert, Ziffernklassen von Zahlen und Zeitpunkten, Escape-Folgen und Rohbytes in Namen und Werten, Stellen eines Fehlers) die Menge, die Zahl ihrer Elemente, der Generator über die Klasse, das Orakel (nie der Leser) und die Prüfgrenze der unbeschränkten Menge als akzeptiertes Negativ mit Grund und Folge im Betrieb. `null` und fehlend sind zwei Elemente jeder Menge der Properties; das leere Element steht neben dem fehlenden (`BEO-REPO/leere-zeile-ist-nicht-fehlende-zeile`).
5. **Wächter für fremde Bytes** nach dem Muster von `SPEC-160` Nr. 6 (Grenzen mit beiden Seiten, zählbare Größe der Zuteilung je Aufruf, Eingabe über der Grenze wird nicht weiter gelesen) und **Schnittfrage** nach `SPEC-128` Nr. 1: die Tabelle ergibt voraussichtlich *S* ≥ 2223; ist ein Teil ohne geteilten Mechanismus abtrennbar (etwa Body-Form und Grenzen gegen Typprüfung der acht Typen), schneidet der Architect vor dem Code und meldet den neuen Plan an den Planner, sonst steht der Grund (`schnitt-grund`) im Block.
6. **Aufrufstellen für `slice-mock-schreiben`** (Sendung, `AGENTS.md` §3.13): Welche Zusagen des Vertrags je Aufruf gelten (`ctx`, Fehlerabbildung, Kontextfehler), steht als Menge der Aufrufstellen im Vertrag; der Planner trägt sie bei der Closure in §1 und DoD von `slice-mock-schreiben` und `slice-batch-changesets` ein.

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7, Start, Planner): Dieser Slice berührt laut §3 keine Datei der Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel) und hat die Linie B. Ändert Vertrag oder Code eine dieser Dateien (auch `.a-check.yml` und `.golangci.yml`), geht das als Rückgabe an den Planner (§6, Ausgang *offen*), nicht in den Zweig. Stand beim Start: kein Slice der Gate-Kette in `in-progress/` (dort liegen `slice-audit-dienste`, Linie D, und `slice-query-interpreter`, Linie B, beide ohne Datei der Gate-Kette).

**Größenmarke** (`SPEC-128`; Faktoren allein dort; Stand beim Start, Planner, 2026-11-06). Der Block mit `basis`, `schaetzung` und `marke` steht auf `main` nicht (ein Block ohne Zahlen der Tabelle wäre ein Platzhalter und machte `make groessenmarke` rot); der Architect legt ihn im Zweig an, `basis` ist der letzte Commit, der Vertrag, Tabelle und Raster abschließt. Pfade: Dateien statt Verzeichnisse dort, wo `slice-query-interpreter` dasselbe Verzeichnis ändert (`src/hexagon/model/`), sonst die Verzeichnisse `src/hexagon/ports/outbound/` (nur neue Dateien und `codec_satz.go`) und `src/adapters/driven/odata/jsoncodec/` (nur neue Dateien); Tests eingeschlossen. **Vorläufige Rechnung ohne Tabelle, nur Richtwert, nicht bindend:** rund 45 Zeilen (2 nebenläufig, 43 andere), *f* = 0 bis 1, *p* rund 1400 (Vergleich: der Leser `jsoncodec` 1640 am Ende bei 1500 erwartet, der Schreiber 1000 erwartet), also *S* = 85·2 + 80·43 + 200·0 + 1400 = 5010 und *M* = 100 · ⌈135·5010 / 10000⌉ = 6800; die Schnittfrage stellt sich (*S* ≥ 2223). Die bindende Rechnung steht erst mit der Tabelle des Architect; der Planner setzt dann Schätzung und Marke in eigenem Commit vor dem ersten Code-Commit (`AGENTS.md` §3.14).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/`. Stand beim Start (Planner, 2026-11-06): erfüllt, `welle-record-replay` ist geschlossen; zugleich gilt der Trigger von Teil 1 von [welle-mock](../welle-mock.md) (§2: `slice-baseline-import` in `done/`) und die empfohlene Reihenfolge (Kette, dann `slice-odata-json-schreiben`, dann dieser Slice; `slice-odata-json-schreiben` liegt in `done/`). `slice-query-interpreter` (Linie B) läuft zugleich: Entscheidung des Users, höchstens drei Worktrees, die Regel „höchstens eine Linie von Teil 1 zugleich“ bleibt als Hinweis stehen. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): die Grenze für die Größe eines Request-Bodys im Mock und der Umgang mit `@odata.*`-Annotationen im Body sind nicht entschieden (siehe §6).

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

- Grenze für die Größe eines Request-Bodys im Mock: `MAX_CAPTURE_BODY` gilt dem Capture, nicht dem Mock — **Ausgang:** offen; Architect (Auftrag 3 in §3).
- `@odata.*`-Annotationen und `@odata.context` im Request-Body (ignorieren oder abweisen) — **Ausgang:** offen; Architect (Auftrag 3 in §3).
- Doppelte Property in einem JSON-Objekt und Zahlen als Zeichenkette (`Edm.Int64`, `Edm.Decimal`) — **Ausgang:** offen; Raster-Zellen mit Ort (Aufträge 2 und 3 in §3).
- Die Menge der Body-Formen ist unbeschränkt (jedes JSON-Wert) — **Ausgang:** offen; Prüfgrenze und ihre Folge im Betrieb als akzeptiertes Negativ oder Generator über die Klasse (`AGENTS.md` §3.17; Auftrag 4 in §3).
- Gate-Kette (Bedingung in §3): ein Bedarf an einer Änderung der Gate-Kette — **Ausgang:** weiter offen (nicht zu erwarten); Adresse: Planner, falls der Architect oder der Implementer ihn meldet.
- Konflikt mit `slice-query-interpreter` (Linie B, zugleich in `in-progress/`) an `src/hexagon/model/` (Eingabetyp der Entity gegen die „Menge typisierter Entities“ des Interpreters) und an der Spezifikation (Kennungsbereich: hier `SPEC-180` bis `SPEC-189`, dort ab `SPEC-190`) — **Ausgang:** offen; Adresse: Architect (nutzt der Eingabetyp den Entity-Typ des Modells, geht jede Änderung daran als Rückgabe an den Planner; die Pfade der Marke nennen Dateien statt `src/hexagon/model/`).
- Größenmarke: die Schnittfrage stellt sich nach der vorläufigen Rechnung (*S* rund 5010 gegen 2223) — **Ausgang:** offen; Adresse: Architect (Auftrag 5 in §3), Marke setzt der Planner nach der Tabelle.
- Zweig ohne Lauf unter dem Detektor (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, 3×, offen, Lese-Schritt bei der Closure von `welle-mock` oder des nächsten Slice mit Nebenläufigkeit) — **Ausgang:** offen; Adresse: Architect, falls der Vertrag eine Zusage über Nebenläufigkeit macht (Wiederholung, Gleichzeitigkeit der Aufrufe), sonst entfallen.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-06; Zähler = Zahl der Dateien in `evidence/`). Der Slice liest fremde Bytes. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×, `.claude/commands/implement-slice.md` Schritt 14, `.claude/agents/implementer.md`, `.harness/skills/reviewer.md`), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (25×, `AGENTS.md` §3.17), `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (5×, `AGENTS.md` §3.18), `BEO-REPO/randform-erst-im-review-entschieden` (33×, `AGENTS.md` §3.16), `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, `implement-slice` Schritt 14 und Reviewer-Skill) für die Menge der Body-Formen, `BEO-REPO/ziffern-klasse-ohne-schleife` (5×, `implement-slice` Schritt 14) für Zahlen und Zeitpunkte als Zeichenkette, `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, `implement-slice` Schritt 14), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.16), `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (3×, `implement-slice` Schritt 19) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (13×, `AGENTS.md` §3.15). Offen: `BEO-REPO/leere-zeile-ist-nicht-fehlende-zeile` (1×, Treffer für leer gegen fehlend, Auftrag 4 in §3), `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (3×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure, §6), `BEO-REPO/spezifikation-widerspricht-sich-innerhalb-einer-nummer` (1×, Treffer für den Vertrag mit vielen Randformen) und `BEO-REPO/lint-anpassung-verengt-den-faenger` (1×, Treffer für Umbauten nach dem Lint). `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (7×) ist geplant in `slice-guard-schreibwege`; bis dahin gilt die Betriebsregel „nur `Edit` und `Write`“ (Übergangsregeln Nr. 8). Zu Schätzung und Marke: `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (11×, geplant: Neukalibrierung von `SPEC-128` durch den Architect). Keiner der offenen erreicht mit diesem Slice zusätzlich die Schwelle ohne weiteren Vorgang.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
