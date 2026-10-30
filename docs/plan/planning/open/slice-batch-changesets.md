# Slice slice-batch-changesets: Ausführung der Einzelrequests und atomare Changesets, Antwort des Batch

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-QA-03`, `LH-FA-07`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-16.a` (atomare Changesets, Executor- und Serializer-Contract-Test für `$batch`) · `LH-FA-07.a` (Zustand bleibt bei Fehler unverändert) · `LH-FA-09.a` (Isolation je Workspace) · `SPEC-108` Nr. 1 (`WithWorkspaceTransaction`) · `SPEC-049` (Lastenheft §5: `$batch` nicht im ersten Abnahmeumfang) · `SPEC-022` (`ODR-E6001`) · `SPEC-006` · `ARC-002`, `ARC-003`, `ARC-005` · `architecture.md` §7 (Erweiterungspunkte).
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

**Ziel:** Der Mock beantwortet `POST /$batch`: er führt die Einzelrequests eines geprüften Batch (aus `slice-batch-parser`) in der Reihenfolge der Teile aus, Lesen über den Lesepfad und Änderungen über den Schreibpfad von `MockUseCase`, und serialisiert die Antwort mit einem Einzelergebnis je Teil. Ein Changeset läuft in genau einer Workspace-Transaktion: gelingen alle Requests, wird sein Zustand übernommen, scheitert einer, wird nichts davon übernommen (Zustand und Revision wie vor dem Changeset). Ein leerer Batch ergibt die gültige leere Antwort; Verweise über `Content-ID` zwischen Requests eines Changeset werden aufgelöst.

**Übernimmt (Route):** Die Route `POST /$batch` im HTTP-Adapter und im Composition Root — `slice-odata-url-pfad` schließt `$batch` aus und nennt `welle-expand-batch` (Sendung von `slice-odata-url-pfad`, `AGENTS.md` §3.13); `slice-bootstrap-m3` bindet nur die Routen des Muss-Umfangs.

**Übernimmt (Fall `If-Match`):** Liegt `welle-etag` vor `welle-expand-batch` in `done/`, trägt dieser Slice den Fall „PATCH mit `If-Match` in einem Changeset“ als Zeile der Tabelle *Zusage · Fänger*; liegt `welle-expand-batch` zuerst, ergänzt `slice-if-match` ihn (Sendung von `slice-if-match`, `AGENTS.md` §3.13, Berührung ohne Kante nach `welle-etag` §5).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Parser und Validator des Batch — `slice-batch-parser`; dieser Slice liest die geprüfte Struktur und prüft keine Bytes der Hülle.
- Parser der URL und Leser der Client-Bodies — vorhanden aus `welle-mock` (`slice-odata-url-pfad`, `slice-odata-json-eingabe`); die Ausführung ruft sie je Einzelrequest.
- Aufzeichnen und Abspielen von Batch-Requests — der Record-Pfad leitet sie als HTTP unverändert weiter, der Replay-Pfad bedient sie unabhängig vom Parser (`welle-expand-batch` §6, `LH-FA-16.a`).
- Actions, Functions, Deep Insert, Batch in Batch und herstellerspezifische Batch-Erweiterungen — Lastenheft §5 bzw. Out-of-Scope von `LH-FA-16`; sie ergeben `501`.
- Ausführung im Modus `hybrid` mit Upstream-Fallback — `welle-hybrid`; der Mock kennt keinen Upstream-Port.

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

- [ ] Ausführung: Einzelrequests (Lesen, POST, PATCH, PUT, DELETE) in der Reihenfolge der Teile über den Pfad von `MockUseCase`, `Content-ID`-Verweise aufgelöst, Antwort mit einem Einzelergebnis je Teil in der Form des Architect; leerer Batch ergibt die gültige leere Antwort; Route `POST /$batch`.
- [ ] Atomarität: ein Changeset in einer Workspace-Transaktion; scheitert ein Request an jeder Stelle und mit jeder Art von Fehler, bleiben Zustand, Revision und Zähler der Entities wie vor dem Changeset, und die Antwort nennt den Fehler; Changesets und Requests anderer Workspaces bleiben unberührt (`LH-FA-09`).
- [ ] Executor- und Serializer-Contract-Test für `$batch` (`LH-FA-16.a`); ein Test im Modus `mock` über die Abnahme der Welle: gültiger Batch mit Einzelergebnissen, leerer Batch, fehlerhaftes Changeset ohne Teiländerung.
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
| `src/hexagon/services/mock/` | update | Ausführung des Batch; die Schreibschritte laufen als Einheiten innerhalb einer Transaktion, die der Aufrufer hält |
| `src/hexagon/ports/inbound/` | update | Eingang für `$batch`, falls der Architect ihn nicht im Port von `MockUseCase` führt |
| `src/adapters/driving/httpodata/` | update | Route `POST /$batch`, Antwort als Strom |
| `src/adapters/driven/odata/` (Paket des Batch-Lesers) | update | Schreiber der Batch-Antwort (Serializer) |
| `src/bootstrap/` | update | Verdrahtung der Route im Modus `mock` |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-16` (Happy Path, Boundary, Negative); Schleife über Stelle und Art des Fehlers im Changeset; Wirkungs-Fänger für die Serialisierung |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Nebenläufig: die Zeilen der Tabelle zählen im Block (`SPEC-128`) als nebenläufig; „unter einer Sperre, ohne sie dazwischen freizugeben“ bekommt je Fenster der Sperre (zwischen dem ersten und dem letzten Request eines Changeset, zwischen letztem Request und Commit) eine Zeile mit Fänger oder benanntem Negativ (`AGENTS.md` §3.10).
- Die Zusage „kein Teilzustand nach einem Fehler“ ist eine Allaussage über Stellen und Arten: der Test läuft als Schleife über jede Position im Changeset und jede Art des Fehlers (`400`, `404`, `409`, `501`, Abbruch des Kontexts), nicht als Liste der Fälle, die einfielen (`AGENTS.md` §3.15, §3.17).

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis in der Spezifikation, Neues unter einer künftigen Nummer, die der Architect vergibt):

1. Zerlegung der Schreibpfade von `slice-mock-schreiben` in Schritte, die in der Transaktion eines Aufrufers laufen (`SPEC-108` Nr. 1 kennt die Form `WithWorkspaceTransaction`, eine Einzelmutation öffnet ihre eigene); die Form der Schritte und wer die Transaktion hält. Der Refactor am Code von `slice-mock-schreiben` ist Teil dieses Slice (§3), ein Vertrag in `done/` ändert sich dadurch nicht.
2. Revision: eine Erhöhung je Changeset oder je Request; Zähler `entity_version` je Request; Sichtbarkeit der Zwischenstände innerhalb des Changeset (spätere Requests sehen frühere) und nach außen (keine).
3. Fehler in einem Changeset: Form der Antwort (ein Einzelergebnis für das Changeset mit dem Fehler des auslösenden Requests), Status der Hülle, Verhalten der Requests außerhalb von Changesets nach einem Fehler (weiter oder Abbruch, `continue-on-error`).
4. `Content-ID`: Auflösung von `$<id>` in URL und Body gegen die Entity, die ein früherer Request angelegt hat; Verweis auf einen gescheiterten oder unbekannten Request.
5. Antwortform: Hülle, Teile, `Content-ID` der Antwort, Header je Teil, Streaming der Antwort gegen das Halten im Speicher, Größe.
6. Coverage und `STRICT_COVERAGE` für Lese-Requests im Batch (derselbe Maßstab wie einzeln) und die Wirkung eines Abbruchs des Kontexts mitten im Batch.
7. Zugeordnete Workspace: ein Workspace je Batch (die Zuordnung des Hüllen-Requests gilt für alle Teile) oder je Teil.

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

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Trigger der Welle) und `slice-batch-parser` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-batch-parser` (die geprüfte Struktur ist die Eingabe), `slice-mock-schreiben` (Schreibpfad und `WithWorkspaceTransaction`, deren Schritte der Changeset in einer Transaktion zusammenfasst), `slice-mock-lesen` (Lese-Requests im Batch), `slice-odata-url-pfad` und `slice-odata-json-eingabe` (Ziele und Bodies der Einzelrequests), `slice-odata-json-schreiben` (Bodies der Einzelergebnisse), `slice-http-adapter` (Route und Strom der Antwort). Keiner davon außer den Teilen aus `done/` (Transaktion im SQLite-Adapter, Modell) liegt heute vor: der Slice ist **nicht** vor `welle-mock` startbar.
- **Binary und Handbuch:** die Route erreicht das Binary im Modus `mock` mit `slice-bootstrap-m3`; Beispiel und Kataloge-Gate wie in `slice-etag-versionen`.
- **Was der Slice nicht braucht:** `slice-expand-parser` und `slice-expand-ausfuehrung` (ein Lese-Request mit `$expand` im Batch läuft über den Lesepfad und ist ein Fall des Rasters, keine Abhängigkeit), `welle-etag` (Berührung ohne Kante, siehe §1), `welle-hybrid`, `welle-odata-v2`, `welle-betrieb`, `welle-admin-api`.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil; der Slice liest keine fremden Bytes und schreibt das Antwortformat.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt; stellt die Schnittfrage von `SPEC-128` den Schnitt, wird vor dem ersten Code-Commit geschnitten, etwa an Einzelrequests ohne Changeset gegen Changesets, falls beide keinen Mechanismus teilen (sonst steht der Grund in §4).
- `in-progress` → `open` (blockiert — Carveout?): der Schreibpfad von `slice-mock-schreiben` lässt sich nicht ohne Vertragsänderung in einer Transaktion des Aufrufers ausführen (Auftrag 1, siehe §6).

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

- Atomarität und Serialisierung sind Zusagen über Nebenläufigkeit: der Detektor meldet nur, was läuft, und sieht kein Fenster bei lauter geordneten Zugriffen — **Ausgang:** offen; je Fenster der Sperre eine Zeile der Tabelle mit Wirkungs-Fänger (eine gleichzeitige Einzelmutation während des Changeset erscheint nicht zwischen dessen Requests), mit Rate je Mutation und je Lauf von `go test` samt Zahl der Prozesse und Messendem (`AGENTS.md` §3.10; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` 2×, `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` 1×).
- Der Schreibpfad aus `slice-mock-schreiben` öffnet je Mutation eine eigene Transaktion; ein Changeset braucht dieselben Schritte in einer — **Ausgang:** offen; Architect (Auftrag 1) beim Start, der Refactor liegt in diesem Slice; geht er nur mit Vertragsänderung in `done/`, geht der Punkt an den Planner (Slice davor, `AGENTS.md` §3.13).
- Größe der Antwort und des Requests: der Request-Body-Limit des HTTP-Adapters (`slice-http-adapter`, Risiko *Request-Größe*) und die Grenze des Batch-Lesers — **Ausgang:** offen; Architect (Auftrag 5, `slice-batch-parser` Auftrag 3).
- Ein Abbruch des Kontexts mitten im Changeset darf keinen Teilzustand hinterlassen — **Ausgang:** offen; Fänger an jeder Position (Schleife), der Kontext des Aufrufs erreicht jeden Aufruf der Kette (`AGENTS.md` §3.15).
- Fakes der Ports halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | leerer Batch, leeres Changeset | Fehler an erster, mittlerer, letzter Stelle | `Content-ID` unbekannt oder auf gescheiterten Request | nicht unterstützte Operation im Teil | Abbruch (Kontext) mitten im Changeset | Wiederholung (derselbe Batch, Duplicate Key) |
|---|---|---|---|---|---|---|
| Batch ausführen (Einzelrequests außerhalb eines Changeset) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Changeset ausführen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Antwort serialisieren | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Route `POST /$batch` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

**Zusatztabelle** (je Paar von Verstößen am selben Changeset und je Modus eine Zeile; `AGENTS.md` §3.16): zwei Fehler im selben Changeset (erster entscheidet), Fehler im Changeset und Fehler eines Requests außerhalb, `STRICT_COVERAGE` an und aus für Lese-Requests, mit und ohne `Content-ID`.

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| (Vertrag, offen) | (Vertrag, offen) | (offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice ist nebenläufig: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) sind offen und gelten für die Rate des Wirkungs-Fängers und für Zweige, die kein Lauf ausführt; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für die Fakes; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Verkörpert: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (`AGENTS.md` §3.10, Sensor `make test-race`). Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
