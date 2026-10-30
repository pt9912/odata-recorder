# Slice slice-etag-versionen: ETag aus dem Versionszähler der Entity

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-etag](../welle-etag.md).

**Bezug:** `LH-FA-15`, `LH-QA-02`, `LH-FA-07`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-15.a` (Entity-Versionen als ETag; ohne aktivierte Unterstützung kein scheinbarer Schutz) · `LH-QA-02.a` (ETags leiten sich aus dem Versionszähler ab) · `SPEC-003` (`entity_version`) · `SPEC-049` (Zeile GET, Spalte Soll: ETag) · `LH-FA-07.b` (Antworten auf POST, PATCH, PUT) · `SPEC-108` Nr. 1 (`WithWorkspaceTransaction`) · `ARC-002`, `ARC-003`, `ARC-004`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Mock liefert je Entity ein Versionskennzeichen (ETag), das allein aus dem Versionszähler `entity_version` der Entity im Workspace folgt: eine unveränderte Entity trägt bei jedem Lesen denselben ETag, jede erfolgreiche Änderung der Entity (POST, PATCH, PUT) macht einen neuen sichtbar, ein fehlgeschlagener Vorgang lässt Zähler und ETag unverändert, und derselbe Zustand ergibt nach Reset und Restore denselben ETag (`LH-QA-02`). Der ETag steht im Header und in der Darstellung der Entity (Entity-Abfrage, Elemente einer Collection, Antworten auf POST, PATCH und PUT mit Entity). Ohne aktivierte ETag-Unterstützung steht keines von beiden in der Antwort: der Mock behauptet keinen Concurrency-Schutz (`LH-FA-15.a`).

**Übernimmt (Zähler und Aktivierung):** Die Entscheidung, wer `entity_version` erhöht, solange kein ETag ausgeliefert wird (offener Punkt in §6 von `slice-mock-schreiben`, Sendung von `slice-mock-schreiben`, `AGENTS.md` §3.13), liegt dort; dieser Slice baut auf ihr auf und liefert die Auslieferung des ETag. Was *aktiviert* heißt (Schlüssel, immer an), entscheidet der Architect in diesem Slice (§3).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `If-Match` und `If-None-Match`, der Vergleich mit dem Zähler und `412` — `slice-if-match`; dieser Slice liefert den ETag, prüft aber keinen Request-Header.
- `Prefer: return=minimal|representation` — `slice-prefer-return`; bis dahin gilt das Muss-Verhalten (POST `201` mit Entity, PATCH und PUT `204`).
- ETags im Record und Replay — aufgezeichnete ETags bleiben Header der Exchanges (`welle-etag` §6); der Mock erzeugt ETags nur aus seinem Zähler.
- Verteilte Versionskoordination über mehrere Instanzen — Out-of-Scope von `LH-FA-15`.

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

- [ ] Bildung und Auslieferung: der ETag folgt allein aus `entity_version` (Form nach Architect), er steht als Header und in der Darstellung der Entity in jeder Antwort, die eine Entity trägt (GET auf Key, Elemente einer Collection, POST, PATCH, PUT mit Entity); gleicher Zustand ergibt denselben ETag, auch nach Reset und Restore.
- [ ] Zähler: `entity_version` steigt bei jeder erfolgreichen Mutation der Entity in derselben Workspace-Transaktion; ein fehlgeschlagener Vorgang, andere Entities und andere Workspaces lassen ihn unverändert; ein Schlüssel, der nach DELETE neu angelegt wird, trägt keinen ETag, den die frühere Entity trug (Randform in §6).
- [ ] Aktivierung: ohne aktivierte ETag-Unterstützung tragen Antworten weder Header noch `@odata.etag` (`LH-FA-15.a`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün.
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
| `src/hexagon/services/mock/` | update | Zähler bei Mutationen, ETag in den Ergebnissen von Lesen und Schreiben |
| `src/hexagon/model/` | update | Typ für das Versionskennzeichen, falls der Architect ihn im Domain-Modell verlangt |
| `src/adapters/driven/odata/jsoncodec/` | update | `@odata.etag` in der Darstellung der Entity (Schreiber aus `slice-odata-json-schreiben`) |
| `src/adapters/driving/httpodata/` | update | Header `ETag` in der Antwort |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-15` (Happy Path und Boundary) je Antwortart; Wirkungs-Fänger für den Zähler unter gleichzeitigen Änderungen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis in der Spezifikation, Neues unter einer künftigen Nummer, die der Architect vergibt):

1. Form des ETag: stark oder schwach, Kodierung des Zählers, Anführungszeichen; ob Service, Entity-Set oder Schlüssel eingehen (zwei Entities mit gleichem Zähler tragen gleiche oder verschiedene ETags).
2. Aktivierung: ein Konfigurationsschlüssel oder immer an; was „ohne aktivierte Unterstützung“ in der Antwort bedeutet (`LH-FA-15.a`), und ob ein `If-Match`-Header dann ignoriert oder abgewiesen wird (die Prüfung selbst liefert `slice-if-match`, die Wahl trifft der Architect hier).
3. Startwert von `entity_version` bei der Materialisierung der Baseline, bei POST und nach DELETE mit erneutem POST desselben Schlüssels (ein früherer ETag darf eine neue Entity nicht treffen).
4. Wer `entity_version` erhöht (Entscheidung aus `slice-mock-schreiben`, §6 dort lesen) und dass Restore und Reset den Zähler des Zustands wiederherstellen, nicht fortzählen.
5. Ob eine Collection-Antwort je Element einen ETag trägt und ob die Collection selbst einen hat.
6. Ob `@odata.etag` im Body oder nur der Header ausgeliefert wird (Metadaten-Stufe der Antwort).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect mit dem Vertrag, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Trigger der Welle). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-mock-schreiben` (der Zähler `entity_version` wird in der Workspace-Transaktion einer Mutation erhöht, und der Slice liest die Entscheidung dazu aus dessen §6), `slice-mock-lesen` und `slice-odata-json-schreiben` (der ETag erscheint in der Antwort des Lesens, `@odata.etag` ist ein Teil des Schreibers), `slice-snapshot-use-case` (Restore stellt Zähler und ETag wieder her; der Fänger dafür setzt den Use Case voraus). Kein Typ und kein Port dieser Slices liegt heute in `done/`: der Slice ist **nicht** vor `welle-mock` startbar.
- **Binary und Handbuch:** Der Modus `mock` erreicht das Binary erst mit `slice-bootstrap-m3`; das Handbuch-Beispiel der DoD läuft gegen dieses Binary, `slice-handbuch-meldungskatalog` liefert das Kataloge-Gate. Beide liegen mit `welle-mock` in `done/`.
- **Was der Slice nicht braucht:** `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-expand-batch` und `welle-odata-v2` (Reihenfolge nach den Muss-Wellen ist Priorisierung, keine Abhängigkeit); `slice-if-match` und `slice-prefer-return` (die Folge-Slices bauen auf diesem, nicht umgekehrt). M2 ist über `welle-mock` Voraussetzung, nicht durch einen eigenen Typ dieses Slice.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil der Schätzung, und der Slice hat keinen Leser fremder Bytes.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): Form des ETag, Aktivierung oder Startwert des Zählers sind nicht entschieden (siehe §6), oder `slice-mock-schreiben` hat die Frage nach dem Zähler anders beantwortet, als dieser Plan annimmt.

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

- Ein Schlüssel, der gelöscht und neu angelegt wird, könnte denselben ETag wie die frühere Entity tragen; ein veralteter `If-Match` träfe dann die neue Entity (die Folge zeigt sich in `slice-if-match`) — **Ausgang:** offen; Architect (Auftrag 3).
- Der Zähler unter gleichzeitigen Änderungen: zwei Mutationen derselben Entity dürfen nicht denselben Zähler erzeugen; der Detektor meldet nur, was läuft — **Ausgang:** offen; Wirkungs-Fänger, der die verschiedenen ETags nach N gleichzeitigen Änderungen zählt, mit Rate je Mutation und je Lauf von `go test` samt Zahl der Prozesse und Messendem (`AGENTS.md` §3.10; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` 2×, `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` 1×).
- Restore und Reset: der Zähler des Zustands, nicht ein fortlaufender, sonst unterscheidet sich der ETag nach Reset von dem des ersten Laufs (`LH-QA-02`) — **Ausgang:** offen; Architect (Auftrag 4), Fänger über Reset und Restore.
- Fakes der Ports (Repository, Schreiber) halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Entity vorhanden | Entity fehlt oder Tombstone | Unterstützung nicht aktiviert | Abbruch (Kontext) | Wiederholung |
|---|---|---|---|---|---|
| GET Key | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| GET Collection (je Element) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| POST | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| PATCH, PUT | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| DELETE, danach POST desselben Schlüssels | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Reset, Restore | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice ist nebenläufig (Zähler unter gleichzeitigen Änderungen): `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) gelten für die Rate des Wirkungs-Fängers und für Zweige, die kein Lauf ausführt; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für die Fakes; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Verkörpert: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (`AGENTS.md` §3.10, Sensor `make test-race`). Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
