# Slice slice-v4-schreiber-marker-text: Der Text `[REDACTED]` eines `Edm.String` im V4-Schreiber

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung außer der DoD dieses Slice.

**Bezug:** `LH-FA-07`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-169` Nr. 1 und 4 (Zuordnung der Arten, Rundlauf Nr. 8) · `SPEC-168` Nr. 4 (Wert; `redigiert` als Art ist verboten, der Text `[REDACTED]` eines `Edm.String` nicht) · `SPEC-126` Nr. 10 f (der Leser liest den Rohtext `[REDACTED]` als Marker) · `SPEC-172` Nr. 6 n (die Entscheidung der Version 2 als Vorbild).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner). **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Rundlauf von `SPEC-169` Nr. 8 gilt für jeden Text eines `Edm.String`, auch für `[REDACTED]`: der Schreiber `jsoncodec` (Version 4) schreibt den Text so, dass `ODataCodec` ihn nicht als Marker liest (Escape für `[` nach dem Vorbild von `SPEC-172` Nr. 6 n) oder weist ihn mit `ErrUngueltig` ab; welche der beiden Formen gilt, entscheidet der Architect vor dem ersten Code-Commit in `SPEC-169` und `SPEC-168` Nr. 4.

**Übernimmt:** `slice-v2-payload-schreiben` (Fund, Rückgabe 1 des Architect an den Planner, 2026-11-06: der V4-Schreiber schreibt `[REDACTED]` roh, `ODataCodec` liest den Rohtext als Marker; V2 löst es mit `SPEC-172` Nr. 6 n).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der V2-Schreiber `jsoncodecv2` — `slice-v2-payload-schreiben` löst den Fall dort selbst (Entscheidung 3, Zeile 31 der Tabelle); dieser Slice ändert keine Datei dort.
- Der Leser `ODataCodec` und die Markerform von `SPEC-126` Nr. 10 f — Vertrag in `done/`; die Lücke liegt im Schreiber, und eine Änderung des Lesers wäre eine Änderung des Formats, das der Import aufgezeichnet hat.
- Weitere Texte, die der Leser besonders liest, soweit der Architect keine findet — Suchauftrag des Architect (§3), kein eigener Punkt; findet er welche, gehören sie in diesen Slice, solange sie den Schreiber `jsoncodec` betreffen.
- Die Verdrahtung des Schreibers im Composition Root — `slice-bootstrap-m3`; der Slice ändert den Wert, den `NewSchreiber()` liefert, nicht die Stelle, die ihn einhängt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Vertrag: `SPEC-169` (und, falls nötig, `SPEC-168` Nr. 4) nennt die Form für den Text `[REDACTED]` und die Randformen daneben (`[REDACTED`, `[REDACTED]x`, `[redacted]`, Text mit Leerraum davor und danach, Text in `Collection`); der Architect hat sie vor dem ersten Code-Commit entschieden.
- [ ] Schreiber: `jsoncodec` setzt den Vertrag um; Rundlauf mit `ODataCodec` über die Klasse der Texte (Schleife über alle Elemente, `AGENTS.md` §3.17), Fänger rot ohne die Änderung.
- [ ] Der Vertragstext von `SPEC-169` Nr. 8 und die *Ungeprüft*-Absätze nennen keine Ausnahme mehr für diesen Text.
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
| `src/adapters/driven/odata/jsoncodec/` | update | Schreiben des Textes eines `Edm.String` |
| `spec/spezifikation.md` | update | `SPEC-169` Nr. 1, 4, 8 und *Ungeprüft*, `SPEC-168` Nr. 4 (Architect) |
| Tests | neu | Rundlauf mit `ODataCodec` über die Klasse der Texte; Fälle von `SPEC-169` Nr. 8 |

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis in der Spezifikation):

1. Die Form: Escape für `[` (wie `SPEC-172` Nr. 6 n, mit Prüfung, dass der Leser den Escape als gewöhnlichen String liest) oder `ErrUngueltig`; Grund und Alternative in der Entscheidung.
2. Die Klasse der Texte, die der Leser besonders liest (`SPEC-126` Nr. 10), aus dem Wortlaut gebildet, nicht aus dem Fund: jeder dieser Texte hat eine Zeile in der Tabelle *Zusage · Fänger*.
3. Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und Raster (§6, `AGENTS.md` §3.16) vor dem Übergang `next` → `in-progress`; die Größenmarke setzt der Planner danach (`SPEC-128`).

**Tabelle *Zusage · Fänger*** (füllt der Architect):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `next` → `in-progress`): `slice-odata-json-eingabe` liegt in `done/` (der Zweig ändert `jsoncodec` und `SPEC-125`/`SPEC-168` zugleich; ein Start davor erzeugt Konflikte im selben Paket). Der Vertrag steht in der Spezifikation, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie nach dem Vertrag des Architect vor dem ersten Code-Commit nach `SPEC-128` Nr. 1 (Block `groessenmarke` in diesem Abschnitt); erwartet ist ein kleiner Slice.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): die Form des Textes ist nicht entschieden (siehe §6).

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

- Der Schreiber `jsoncodec` und der Leser `ODataCodec` sind nicht als Paar über die Klasse der Texte geprüft (Beobachtung „Schreiber und Leser nicht als Paar geprüft“; das Register legt `slice-v2-payload-schreiben` bei seiner Closure an, dort `BEO-REPO/` mit dem Beleg dieses Fundes) — **Ausgang:** offen; Fänger: Rundlauf als Schleife über die Klasse (DoD, zweiter Punkt).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- Ein Escape ändert die Bytes, die ein bereits aufgezeichneter oder ausgelieferter Wert trägt — **Ausgang:** offen; Adresse: Architect (Entscheidung zu Nr. 1 des Auftrags nennt die Folge im Betrieb).

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Text `[REDACTED]` | Text mit Rest davor oder dahinter | anderer Text, den der Leser besonders liest | Text in `Collection` |
|---|---|---|---|---|
| Entity schreiben | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: keine Collection-Property in `SPEC-169` Nr. 4, falls der Architect es bestätigt |
| Collection von Entities schreiben | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: wie oben |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-06): `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (`AGENTS.md` §3.17) und `BEO-REPO/randform-erst-im-review-entschieden` (`AGENTS.md` §3.16) gelten; der Planner zieht die Zähler beim Start nach.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
