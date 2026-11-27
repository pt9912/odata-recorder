# Slice slice-absturztest: Prozessende mitten in Mutation, Reset, Restore und Snapshot hinterlässt nur vorher oder nachher

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung über die DoD hinaus. Der Gegenstand (`LH-QA-03`) gehört zu `welle-mock` (in `done/`), und kein Closure-Kriterium von [welle-betrieb](../welle-betrieb.md) nennt ihn. Gebraucht wird er von M7 (Erstabnahme): die [Roadmap](../in-progress/roadmap.md) führt ihn dort als Bedingung.

**Bezug:** `LH-QA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-048` (Absturztest, Nr. 1 bis 5) · `SPEC-049` Nr. 1 (Aufbau) · `LH-QA-03.a`. Geändert wird allein der Vermerk *Stand* beim Liefern.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** — (bis zur Priorisierung).

**Autor:** pt9912 (Planner, Schnitt von [slice-abnahme-belege](../next/slice-abnahme-belege.md) vor dem Code). **Datum:** 2026-11-27.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Test im Paket `bootstrap_test` belegt, dass ein Prozessende mit `SIGKILL` mitten in einer Mutation, einem Reset, einem Restore oder einem Snapshot nach dem Neustart nur den Stand davor oder den Stand danach hinterlässt (`LH-QA-03` Messmethode *Fehler-Injektion*, `SPEC-048`: Kindprozess, Haltepunkte an Port-Dekoratoren, 10 Haltepunkte).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Erweiterungen auf dem Record-Pfad und `ODV4-Q-11` b — [slice-abnahme-belege](../next/slice-abnahme-belege.md).
- Produktcode und ein Haken im Produkt, ein Schlaf als Orakel, `t.Skip` — der Vertrag schließt sie aus (`SPEC-048` Nr. 1); zeigt der Test eine Abweichung, greift die Rückführung in §4.
- Absturz zwischen fsync und Rename, während `COMMIT`, Stromausfall, Anhängen eines Exchange und Import, das gebaute Binary — benannte Negative des Vertrags (`SPEC-048` Nr. 5; Zeilen N1 bis N4).

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

- [ ] `LH-QA-03`: der Absturztest `src/bootstrap/absturz_test.go` nach `SPEC-048` Nr. 1 bis 3 führt die 10 Haltepunkte mit der Prüfung nach jedem (T1 bis T17); er läuft in `make test` und unter `make test-race`.
- [ ] Die Mutationen nach `SPEC-048` Nr. 4 (Transaktion geteilt; Mutation außerhalb der Transaktion; `Create` des Snapshots vor dem `Commit` des Writers) sind je rot gesehen, am benannten Haltepunkt.
- [ ] Der Vermerk *Stand* von `SPEC-048` im liefernden Commit nachgezogen.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
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
| `src/bootstrap/absturz_test.go` | neu | Kindprozess, Port-Dekoratoren, 10 Haltepunkte, Prüfung nach dem Neustart; Fehler-Injektion nach `LH-QA-03`. Er nutzt den Aufbau von `abnahme_aufbau_test.go`, ohne ihn zu ändern |
| `spec/spezifikation.md` | update | Vermerk *Stand* von `SPEC-048` |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Planner aus `SPEC-048`, weil der Vertrag dort steht und der Slice nur Test liefert; „Nr.“ meint `SPEC-048`; der Implementer trägt Mutationen und Raten mit Zahl der Prozesse, Last und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| T1 | Mutation (`PATCH`), Ende in der Transaktion vor `COMMIT`: Stand vorher | Haltepunkt, Nr. 2 | eine Mutation außerhalb der Transaktion | andere |
| T2 | Mutation, Ende nach `COMMIT` vor der Rückkehr: Stand nachher | Haltepunkt, Nr. 2 | die Transaktion in zwei geteilt | andere |
| T3 | Reset, Ende vor `COMMIT`: vorher | Haltepunkt | ein Teil des Reset außerhalb der Transaktion | andere |
| T4 | Reset, Ende nach `COMMIT`: nachher | Haltepunkt | Transaktion des Reset geteilt | andere |
| T5 | Restore, Ende vor `COMMIT`: vorher | Haltepunkt | ein Teil des Restore außerhalb der Transaktion | andere |
| T6 | Restore, Ende nach `COMMIT`: nachher | Haltepunkt | Transaktion des Restore geteilt | andere |
| T7 | Snapshot, Ende vor `Begin` des Blobs: ohne neuen Snapshot | Haltepunkt | `Create` vor `Begin` | andere |
| T8 | Snapshot, Ende im `Write` nach einem Teil der Bytes: ohne neuen Snapshot, `BlobsFehlend` leer | Haltepunkt | `Create` vor dem `Commit` des Writers | andere |
| T9 | Snapshot, Ende nach `Commit` des Writers vor `Create`: ohne neuen Snapshot; der verwaiste Blob ist bis zur Wartung zulässig | Haltepunkt | wie T8 | andere |
| T10 | Snapshot, Ende nach `Create` vor der Rückkehr: mit dem neuen, wiederherstellbaren Snapshot | Haltepunkt | `Create` ohne abgeschlossenen Blob | andere |
| T11 | Die Zahl der gefahrenen Haltepunkte ist das Literal 10 | Zählprüfung | ein Haltepunkt entfällt | andere |
| T12 | Der Kindprozess endet durch `SIGKILL`; ein nicht erreichter Haltepunkt ist rot | Prüfung des Exit, Nr. 3 | der Haltepunkt sendet nicht und der Kindprozess endet regulär | andere |
| T13 | `Startup` nach dem Neustart: kein Fehler, kein Befund der Datenbank, `BlobsFehlend` und `BlobsBeschaedigt` leer | nach jedem Haltepunkt | der Bericht von `Startup` wird nicht gewertet | andere |
| T14 | Entity-State, Revision und Cursors von `ws-1` und die Liste der Snapshots sind genau vorher oder genau nachher; die Orakel stammen aus demselben Aufbau ohne Absturz | Vergleich nach jedem Haltepunkt | der Vergleich lässt einen Teil (Cursors) aus | andere |
| T15 | `ws-2` ist unverändert | Vergleich | die Operation trifft beide Workspaces | andere |
| T16 | Nach `Startup` trägt die Ablage keinen unreferenzierten Blob und keine temporäre Datei | Auflisten der Ablage | die Bereinigung des Starts entfällt | andere |
| T17 | `Restore` von `vor` gelingt, und eine weitere Mutation gelingt | nach jedem Haltepunkt | eine Sperre oder ein beschädigter Blob bleibt zurück | andere |
| N1 | Negativ: Ende zwischen fsync und Rename und während `COMMIT` (kein Port-Aufruf liegt dazwischen). Folge: Atomarität von Rename und Transaktion tragen (2 Punkte ohne Fänger) | Negativ | — | andere |
| N2 | Negativ: Stromausfall, Absturz des Betriebssystems (`SIGKILL` lässt den Seiten-Cache). Folge: Dauerhaftigkeit über einen Ausfall des Hosts nicht belegt | Negativ | — | andere |
| N3 | Negativ: Anhängen eines Exchange und Import (`LH-QA-03` nennt Mutationen und Snapshots). Folge: ein Fehler nur beim Prozessende auf diesen Wegen bliebe ungefangen | Negativ | — | andere |
| N4 | Negativ: das gebaute Binary (Haltepunkt verlangte einen Haken). Folge: eine Verdrahtung ohne Startwartung fängt der Test nicht | Negativ | — | andere |

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; „Nr.“ meint `SPEC-048`):

| Operation | fehlend | leer | unbekannt oder fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Mutation, Reset, Restore | Haltepunkt nicht erreicht: rot, Nr. 3 | gilt nicht: fester Ausgangsstand, Nr. 1 | fremder Workspace `ws-2` unverändert: Nr. 3 | Prozessende an den Haltepunkten, Nr. 2 | nach dem Neustart eine weitere Mutation, Nr. 3 |
| Snapshot erzeugen | wie oben | gilt nicht: der Ausgangsstand ist nicht leer, Nr. 1 | gilt nicht: keine Wahl des Namens im Test | Nr. 2 (vier Haltepunkte) | `Restore` von `vor` gelingt, Nr. 3 |
| Neustart (`Startup`) | fehlender Blob: rot, Nr. 3 | gilt nicht | beschädigter Blob oder Befund: rot, Nr. 3 | gilt nicht: `Startup` wird nicht abgebrochen (Nr. 5 nennt keinen Fall; ein Abbruch hier ist kein Gegenstand) | gilt nicht: je Haltepunkt ein Neustart |

**Betriebsregeln für alle Rollen:** wie in [slice-abnahme-belege](../next/slice-abnahme-belege.md) §3; im Scratchpad nur unter `absturztest-<rolle>/`.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): der Vertrag steht auf `main` (`2e289d4d`, erfüllt); ein Worktree-Platz ist frei; der Planner hat die Schnittfrage unten entschieden. Kein Kettenslice (Übergangsregeln Nr. 7). Kein Code-Vorgänger: der Slice teilt keine Datei mit [slice-abnahme-belege](../next/slice-abnahme-belege.md).

**Größenmarke** (`SPEC-128` Nr. 1, `AGENTS.md` §3.14; Planner): 21 Zeilen (T1 bis T17, N1 bis N4), *a* = 0, *b* = 21, *f* = 3 (Dekoratoren um `WorkspaceTransactor`, `BlobRepository` mit seinem Writer und `SnapshotRepository`), *p* = 0: *S* = 80·21 + 600 = 2280, Marke 3100, **Schnittfrage** (*S* ≥ 2223). `basis`, `schaetzung` und `marke` setzt der Planner beim Start, vor dem ersten Code-Commit. **Offen bis zum Start** (Adresse Planner): ob an der Naht *Transaktion* (T1 bis T6) gegen *Snapshot* (T7 bis T10) geschnitten wird; der zweite Teil riefe allein das Gerüst des ersten (Kindprozess, Aufbau, Prüfung). Wird nicht geschnitten, steht der Grund als `schnitt-grund` im Block. Der Teil des Architect (`pfad`, `produktcode`, `fakes`) ist vom Planner gesetzt.

```groessenmarke
pfad: src/bootstrap/absturz_test.go
produktcode: 0
fakes: 3
```

**Rückführungen — vorab benannt:**

- `in-progress` → `open` (blockiert): ein Haltepunkt ist ohne Haken im Produkt nicht erreichbar, oder der Test zeigt einen Zustand, der weder vorher noch nachher ist. Der Slice ändert keinen Produktcode; er gibt an den Planner und den Architect zurück.
- `in-progress` → `next` (zu groß): über der Marke vor dem Handoff und ohne vorab benannten Teil.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, Review bestanden, `make gates` grün, je Zusage eine Mutation rot gesehen (`AGENTS.md` §3.10), Closure-Notiz mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der Test findet einen echten Fehler (verwaiste Zustände, nicht wiederherstellbarer Snapshot) — **Ausgang:** weiter offen: erste Rückführung in §4, Adresse Planner und Architect.
- Ein Kindprozess unter `make test-race` verlängert den Lauf von `bootstrap` über eine Zeitgrenze (`BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze`) — **Ausgang:** weiter offen; Adresse: Implementer (misst mit Last und Messendem).
- Die Schnittfrage ist offen (§4) — **Ausgang:** weiter offen; Adresse: Planner (beim Start).

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

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, der Default `*` (Kürzel `REPO`); es gibt keine feinere.

**Vorgelagert — offene Beobachtungen sichten:** Register bei der Anlage gesichtet (2026-11-27; der Planner zieht beim Start nach). Offen und berührt: `BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache` (20×), `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (17×), `BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (1×), `BEO-REPO/zustand-unveraendert-nach-fehler-ohne-ruecklesen` (1×; T14 liest zurück), `BEO-REPO/orakel-prueft-nur-die-zaehler-des-testaufbaus` (1×; Orakel nach Nr. 3 aus dem Aufbau ohne Absturz).

Alle berührten Sub-Areas GF.
