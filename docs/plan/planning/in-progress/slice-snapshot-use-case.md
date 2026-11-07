# Slice slice-snapshot-use-case: SnapshotUseCase und Restore des Workspace

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-10`, `LH-FA-09`, `LH-QA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-10.a` · `LH-FA-09.b` (Abgrenzung zum Reset) · `SPEC-108` Nr. 12 bis 16 (Workspace-Snapshot) · `SPEC-025`, `SPEC-065`, `SPEC-094`, `SPEC-096` · `ARC-002`, `ARC-003` · `architecture.md` §4.5.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-snapshot-use-case`, Worktree `/Development/odata-recorder-wt-snapshot-use-case`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `SnapshotUseCase` erzeugt benannte Snapshots, die ausschließlich den Entity-State eines Workspace festhalten, und listet sie; `WorkspaceUseCase.Restore` ersetzt den Entity-State atomar durch den eines Snapshots, erhöht die Revision und lässt alle Replay-Cursor unverändert; unbekannte oder fremde Snapshot-IDs, ein Snapshot eines anderen Service, ein vergebener Name und die Art `baseline` ergeben die Codes der Spezifikation, ohne dass ein Teilzustand übernommen wird.

**Übernimmt:** den Punkt „Snapshots erzeugen und wiederherstellen“ aus `slice-workspace-use-case` (dessen §1 *Ausdrücklich NICHT* nennt `SnapshotUseCase` und `Restore` als Sendung an den Mock).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Endpunkte `GET` und `POST …/snapshots` und `POST …/restore` — `welle-admin-api`; der Use Case ist über den Composition Root erreichbar.
- Löschen einzelner Snapshots — `SPEC-108` kennt es nicht (ein Snapshot geht nur mit seinem Workspace).
- Versionskontrolle mit Merge von Snapshot-Zweigen — Out-of-Scope von `LH-FA-10`.

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

- [ ] Erzeugen und Auflisten: Name im Workspace eindeutig (`ODR-E7006`), nur der Entity-State im Manifest, Cursor nie Bestandteil; ein leerer Workspace ergibt einen leeren Snapshot (`LH-FA-10`, Boundary).
- [ ] Restore: atomar, neue Revision, Cursor unverändert; unbekannte oder fremde ID `ODR-E7002`, anderer Service `ODR-E5004`, Art `baseline` `ODR-E7008`, jeweils ohne Teilzustand.
- [ ] Restore und Mutation desselben Workspace sind serialisiert; die Zusage hat einen Wirkungs-Fänger, nicht nur den Detektor (`AGENTS.md` §3.10).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (Snapshots erreichen den Anwender erst über die Admin-API in `welle-admin-api`).
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
| `src/hexagon/ports/inbound/` | update | `SnapshotUseCase`, `Restore` am `WorkspaceUseCase` |
| `src/hexagon/services/snapshot/` | neu | Erzeugen und Auflisten, Manifest über `BlobRepository` |
| `src/hexagon/services/workspace/` | update | `Restore` ohne Zugriff auf `CursorRepository` |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-10` (Happy, leerer Workspace mit fortgeschrittenem Cursor, ungültige ID); der Fake von `CursorRepository` zählt jede Operation, damit „Restore berührt keinen Cursor“ fangbar ist |

**Aufträge an den Architect zum Start** (Start, Planner, 2026-11-07, `main` = `da89fbae`; vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Architect arbeitet im Zweig `slice-snapshot-use-case`; Ausgang jeweils *offen* in §6). Wo etwas vollständig ist, genügt die Bestätigung mit Nummer.

- **A. Tabelle mit allen Sätzen jeder Vertragsnummer (`AGENTS.md` §3.15).** Jeder Satz jeder Nummer, die der Vertrag führt (`LH-FA-10.a`, `LH-FA-09.b`, `SPEC-108` Nr. 12 bis 16, `SPEC-025`, `SPEC-065`, `SPEC-094`, `SPEC-096`, die Absätze der Spezifikation zu Restore und Snapshot um Zeile 270, die neuen Sätze), auch jeder Nebensatz („ohne dass ein Teilzustand übernommen wird“, „Cursor unverändert“, „nur der Entity-State“, „ein leerer Workspace ergibt einen leeren Snapshot“), hat eine Zeile mit Nr., Zusage, Fänger, Mutation und Klasse (`nebenläufig` oder `andere`, `SPEC-128` Nr. 3) oder ein benanntes Negativ mit Grund und Folge; die Tabelle trägt die Spalten, die das Werkzeug aus `SPEC-128` liest. Jedes Test-Symbol der Tabelle existiert (Register `tabelle-nennt-test-der-nicht-existiert`).
- **B. Raster (§6, `AGENTS.md` §3.16).** Operation {Erzeugen, Auflisten, Restore} × Argumentart (fehlender, leerer, unbekannter oder fremder Wert; Snapshot eines anderen Workspace, eines anderen Service, Art `baseline`; vergebener Name; Abbruch; Wiederholung; Maskenzustand der Entities: beobachtet, nicht beobachtet, Tombstone; leerer Workspace; Cursor vorhanden und fortgeschritten) jede Zelle mit Ort oder „gilt nicht“ mit Grund. Zusatztabelle mit Fänger je Paar von Verstößen am selben Schlüssel (unbekannte ID und Art `baseline`, fremder Service und fremder Workspace, vergebener Name und unbekannter Workspace) und je Modus (Service nur lesend, V2-Service, Restore einer Baseline als Snapshot-ID); die Prüfreihenfolge ist im Vertrag, nicht im Code.
- **C. Pflicht-Mengen mit Achsen, Generator, Orakel und Prüfgrenze, gezählt gegen die Quelle** (Register `menge-im-plan-weicht-vom-orakel-des-tests-ab`, `unbeschraenkte-menge-ohne-benannte-pruefgrenze`): Entity-Zustand (vorhanden, Tombstone, partial/complete, Maske) × Snapshot-Zeitpunkt (vor und nach einer Mutation, nach Reset) × Aktion; die Fehlerabbildung (`SPEC-109` Nr. 5) je Aufrufstelle; die Felder des Manifests und des Snapshot-Datensatzes. Das Orakel trägt Statuscodes, Codes und Meldungstexte als **Literale aus dem Vertrag**, nicht als Konstanten des Pakets.
- **D. Zweige und Rückgaben als Menge, in Runden unter dem Detektor** (Register `zweig-ohne-lauf-unter-dem-detektor`, 16×): Z-Liste mit Fänger je Zweig, Zählprüfung gegen die Quelle; für jeden Zweig, den eine Goroutine nur einmal erreicht (Abbruch des `ctx` mitten im Erzeugen oder im Restore, Rollback, Fehler des Blob-Speichers oder des Repository zwischen Blob und Referenz, Wettlauf von Restore und Mutation), die Runde (ein Lauf je Zweig unter `go test -race -coverprofile`); null Blöcke ohne Treffer oder Negativ mit Grund und Folge. Ein Fenster ohne Haltepunkt am Argument: Masse mit Rate und Messendem, Wirkungs-Fänger in jedem Lauf.
- **E. „Menge ohne ihre Positionen“** (Register `menge-ohne-ihre-positionen`, 2×): jede Menge, die der Vertrag nennt, führt auch Positionen und Reihenfolgen (Entities im Manifest: Mutation am ersten, mittleren, letzten; Reihenfolge der Liste der Snapshots nach `created_at`, dann Kennung, bei gleicher Zeit; mehrere Entity-Sets; mehrere Cursor); **Fixture gleicht nie dem Wert der Mutation** (Register `fixture-gleicht-dem-wert-der-mutation`, 6×). Allaussagen („jeder Fehler lässt den Zustand unverändert“, „Cursor nie Bestandteil“) als Schleife über beide Mengen (`AGENTS.md` §3.15, §3.17).
- **F. `ctx`-Identität, Gegenstand je Aufrufstelle, zählende Fakes, Bindung an den Port** (Register `hilfe-ersetzt-argument-verdeckt-eigenschaften`, `panik-eines-fakes-verdeckt-den-benannten-test` je 4×, `dienst-erfuellt-den-port-ohne-bindung` 2×): der `ctx` des Aufrufs (Identität, nicht nur `Value`), die Fehlerabbildung und der Kontextfehler haben einen Fall je Aufrufstelle der Kette (Workspace-Repository, Snapshot-Repository, Blob-Speicher, Entity-State, Transaktion, Uhr, Generator), nicht am ersten Port; jeder Fake hält jedes Argument fest und zählt jede Operation, auch die, die der Dienst nie rufen darf (jede Operation von `CursorRepository` bei Restore und Erzeugen, `Put` des Entity-State bei Fehler, `Advance` bei Fehler, Löschen eines Snapshots); eine Zusicherung zur Übersetzungszeit bindet Fake und Dienst an den Port. Zählprüfung des Quelltexts: jede Aufrufstelle eines Ports im Dienst steht mit Zahl in der Tabelle, nicht nur die Aufrufstellen mit Zählung im Test.
- **G. Serialisierung und Fenster der Sperre** (`AGENTS.md` §3.10; Register `rate-ohne-detektor-schwankt-mit-der-last` 8×, `sperrfenster-zwischen-sammlungen-ohne-faenger`, `nebenlaeufiger-lauf-nur-im-glueck-gruen` 13×): Restore, Mutation und Reset desselben Workspace sind serialisiert (Zusage aus `slice-mock-schreiben`); je Fenster der Sperre (zwischen Lesen des Manifests und Ersetzen des State, zwischen Ersetzen und `Advance`, zwischen Erzeugen des Manifest-Blobs und Referenz, zwischen zwei Workspaces) eine Zeile mit Wirkungs-Fänger oder benanntem Negativ; vor einem Negativ ein Haltepunkt an einem Argument, das der Test übergibt (`ctx`, Fake, Transaktionsfunktion). Wirkungs-Fänger zählen, was das Fenster zuließe; Raten mit Zahl der Prozesse und Messendem.
- **H. Entscheidungen vor dem Code (§6 des Plans):** Manifest und `state_hash` (deterministisch, `LH-QA-02`; Reihenfolge der Entities, Darstellung der Maske und der Tombstones, ob Tombstones im Manifest stehen), `metadata_hash`, Dauer der Transaktion beim Erzeugen (Blob-Schreiben außerhalb oder innerhalb), Bedeutung der Revision beim Restore (neue Revision auch bei gleichem State), Restore einer Maske (`SPEC-044`: Unbekanntes wird nie zu `null`), Prüfreihenfolge der Fehler, Namen (leer, Länge, Zeichen, Groß-/Kleinschreibung), Obergrenzen. Jede im Vertrag mit Ort; der Implementer entscheidet keine.
- **I. Kennungen und Katalog-Pflicht (`SPEC-201` Nr. 3 und 11).** Der Architect nimmt **nur freie Nummern**; `SPEC-192` bis `SPEC-199` sind vergeben (Mock), `SPEC-236` Inspection, `SPEC-237` Reserve, `SPEC-238` bis `SPEC-240` Admin; der Bereich 160 bis 199 der Linie B ist erschöpft, die Kennungsfrage ist beim User offen. Der Vertrag steht deshalb als neue Sätze in bestehenden Nummern (`SPEC-108` Nr. 12 bis 16, die Absätze zu Restore, Workspace-Dienst-Nummern); braucht er eine neue Nummer, geht die Frage als Rückgabe an den Planner, bevor er sie schreibt. Codes: `ODR-E7006`, `ODR-E7002`, `ODR-E5004`, `ODR-E7008` stehen in `src/hexagon/model/fehler.go` und, außer `ODR-E5004` (im Binary über Replay erreichbar), in `harness/katalog-ausnahmen.txt` mit Grund (Snapshots verwaltet nur die Admin-API); der Architect bestätigt, dass kein neuer Code entsteht (der nächste wäre der 41. nach `ODR-E2102`, Zahl in `SPEC-230` Nr. 8), sonst folgt `fehler.go` samt Tests der Telemetrie und Ausnahmen, und `make katalog-check` steht im Bericht.
- **J. Schnitt und Marke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1). Der Architect meldet *a*, *b*, *f* und *p* aus der Tabelle und setzt im Block `groessenmarke` weder `basis` noch `schaetzung` noch `marke`; den Block setzt der Planner nach der Bestätigung in eigenem Commit vor dem ersten Code-Commit. Wird *S* ≥ 2223, wird **vor dem Code** geschnitten (Kandidat: Erzeugen und Auflisten gegen Restore; der Schnitt muss einen grünen Stand lassen, Register `schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices`) oder `schnitt-grund` in §4. Die Auslastung der letzten Teile lag bei 0,41 bis 0,71 (Register `schaetzformel-ueberschaetzt-den-slice`, 25×): Produktcode wird überschätzt, Tests unterschätzt; der Architect benennt das in der Meldung, ohne die Formel zu ändern.
- **K. Pfade neben den Zweigen.** Dieser Slice schreibt `src/hexagon/ports/inbound/`, `src/hexagon/services/snapshot/` (neu), `src/hexagon/services/workspace/` und `spec/spezifikation.md` (nur eigene Sätze); `tests/` und der Composition Root, falls der Use Case dort gebunden wird, nennt der Architect. Der Architect bestätigt die Liste und nennt jede Datei darüber hinaus (Abweichung an den Planner).
- **L. Reports der Rollen und Betriebsregeln.** Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (Register `report-der-rolle-faellt-docs-check`); der Planner reicht den Satz an Reviewer und Verifier weiter. Der Implementer schreibt Repo-Dateien nur mit `Edit` und `Write` (Register `betriebsregel-fuer-repo-dateien-verletzt`, 12×), und er prüft vor dem Handoff, dass jedes Test-Symbol der Tabelle existiert.

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Fake eines Ports hält jedes Argument fest und zählt jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-mock-schreiben` liegt in `done/` (erst dort entstehen die Änderungen, die ein Snapshot festhält). **Stand beim Start** (Planner, 2026-11-07, `main` = `da89fbae`): erfüllt, `slice-mock-schreiben` liegt in `done/` (Zweig-Merge nach `main`). §1 *Ausdrücklich NICHT* und DoD der Nachbarn gelesen (`AGENTS.md` §3.13): `slice-mock-abnahme` (prüft Reset-Determinismus, Isolation und Neustart mit Snapshots, startet nach diesem Slice), `slice-bootstrap-m3` (verdrahtet, was dieser Slice liefert, keine neue Fachlogik), `slice-batch-changesets` (nennt weder Snapshot noch Restore) und `slice-telemetry-einbau-mock` (Mutations-Span von Reset und Restore entscheidet dessen Auftrag 1; der Dienst muss dann bestehen) schließen keine Sendung dieses Slice aus. Die Codes `ODR-E7006`, `ODR-E7002`, `ODR-E7008` und `ODR-E5004` stehen in `src/hexagon/model/fehler.go`, die ersten drei in `harness/katalog-ausnahmen.txt` (Auftrag I). Die Gate-Kette bleibt unberührt (Nr. 7 der Übergangsregeln). Speicher beim Start: 12 GB verfügbar. **Marke: noch nicht gesetzt** (Prosa, kein Block auf `main`); eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt (Auftrag J): danach setzt der Planner in eigenem Commit vor dem ersten Code-Commit den Block `groessenmarke` mit `basis`, `schaetzung` und `marke`; bis dahin bindet kein Gate. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): Form des Manifests und des `state_hash` sind nicht entschieden (siehe §6).

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

- Abbruch zwischen Schreiben des Manifest-Blobs und dem Commit des Snapshots (Blob zuerst, dann Referenz; ein verwaister Blob ist zulässig, `SPEC-048`) — **Ausgang:** offen; Fänger durch Fehler-Injektion an den Haltepunkten des Ports, nicht am `ctx`-Argument allein.
- Bildung von `state_hash` und Manifest (deterministisch, `LH-QA-02`; Reihenfolge der Entities, Darstellung der Maske) — **Ausgang:** offen; Architect.
- Restore gleichzeitig mit Mutation und mit Reset desselben Workspace — **Ausgang:** offen; Serialisierung je Workspace aus `slice-mock-schreiben`, Wirkungs-Fänger mit Rate je Lauf (`AGENTS.md` §3.10).
- Restore einer Maske: ein Snapshot hält Entities mit Property-Maske; Wiederherstellung darf Unbekanntes nicht zu `null` machen (`SPEC-044`) — **Ausgang:** offen; Raster-Zelle.

- Aufträge A bis L an den Architect zum Start (§3): Sätze je Vertragsnummer, Raster, Pflicht-Mengen mit Zählprüfung, Zweige und Runden unter dem Detektor, Menge mit Positionen und Fixture, `ctx`-Identität und Gegenstand je Aufrufstelle, zählende Fakes und Bindung, Fenster der Sperre, Entscheidungen vor dem Code, Kennungen und Katalog, Schnitt und Marke, Pfade, Reports — **Ausgang:** offen; Adresse: Architect (Bestätigung in diesem Abschnitt oder Ergänzung), Implementer (Test), Reviewer und Verifier (Auftrag L).
- Kennungen: keine freie `SPEC`-Nummer der Linie B; die Kennungsfrage (Bereich 160 bis 199) liegt beim User — **Ausgang:** offen; Adresse: Planner (Rückgabe des Architect, wenn der Vertrag eine neue Nummer braucht).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Offen und für diesen Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes von `CursorRepository`, `SnapshotRepository` und `BlobRepository`; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für Restore unter Mutation. Verkörpert: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15). Keiner der offenen Einträge erreicht mit diesem Slice 3×.
Beim Start nachgezogen (Planner, 2026-11-07, `main` = `da89fbae`, Zähler aus den Dateien unter `evidence/`): `zweig-ohne-lauf-unter-dem-detektor` 16× (Schwelle lange erreicht; Verkörperung beim Lese-Schritt der Welle-Closure, Auftrag D), `rate-ohne-detektor-schwankt-mit-der-last` 8× und `nebenlaeufiger-lauf-nur-im-glueck-gruen` 13× (Auftrag G), `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `dienst-erfuellt-den-port-ohne-bindung` 2× (Auftrag F), `menge-ohne-ihre-positionen` 2× und `fixture-gleicht-dem-wert-der-mutation` 6× (Auftrag E), `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3× und `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11× (Auftrag C), `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 47× und `tabelle-nennt-test-der-nicht-existiert` 1× (Auftrag A), `randform-erst-im-review-entschieden` 34× (Auftrag H), `schaetzformel-ueberschaetzt-den-slice` 25× (Auftrag J), `sqlite-transaktionsvertrag-rot-unter-last` 6× (erstes `make gates` unter Last), `betriebsregel-fuer-repo-dateien-verletzt` 12× (Auftrag L).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
