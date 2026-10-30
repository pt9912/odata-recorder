# Slice slice-expand-parser: $expand in der OData-URL, Parser und Validator gegen die Navigationen des Modells

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-16.a` (Erweiterungspunkt `$expand`, eigener Parser- und Validator-Contract-Test, unbekannte Operationen `501`) · `SPEC-049` (Zeile GET: Navigation-Expand, Spalte „Nicht in MVP“) · `SPEC-022` (`ODR-E6001`) · `SPEC-002`, `SPEC-125` Nr. 11 (Navigation-Properties des Modells) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkte: separate AST-Knoten, Validatoren und Executors).
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

**Ziel:** Der Parser der OData-URL (`urlparser`, hinter dem Port der Codec-Gruppe aus `slice-odata-url-pfad`) liest die Systemoption `$expand` aus `raw_query` und überführt sie in eigene Knoten des Query-AST des Domain-Modells; der Validator prüft jede genannte Navigation gegen die Navigation-Properties des Entity-Typs im Modell des Dienstes (auch geerbte). Eine gültige Form im unterstützten Umfang ergibt den AST, eine ungültige `400`, eine gültige, aber nicht unterstützte (Umfang nach Entscheidung des Architect) `501` mit `ODR-E6001`. Der Slice liest nur Text und prüft gegen das Modell; er greift auf keine Daten zu (`LH-FA-16.a`: Parser- und Validator-Contract-Test).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Auflösen der Navigation gegen Daten, Datenbasis der Navigationen und Ausgabe der eingebetteten Entities — `slice-expand-ausfuehrung` (Executor- und Serializer-Contract-Test von `LH-FA-16.a` liegen dort); dieser Slice liefert den geprüften AST.
- Verschachtelte Optionen im `$expand` (`$filter`, `$select`, `$orderby`, `$top`, `$skip`, `$count`, `$levels`, `$expand`) — der unterstützte Umfang ist Entscheidung des Architect vor dem Start (`welle-expand-batch` §6); was er nicht nennt, ergibt `501`. Nennt er `$filter`, braucht der Slice zusätzlich `slice-odata-url-filter` und `slice-edm-ausdruecke` (§4).
- `$apply`, `$search`, Actions, Functions, Deep Insert — nicht im ersten Abnahmeumfang (Lastenheft §5); sie bleiben `501`.
- `$batch` — `slice-batch-parser`; die beiden Äste sind unabhängig.
- Herstellerspezifische Navigations-Erweiterungen — Out-of-Scope von `LH-FA-16`.

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

- [ ] Grammatik und AST: `$expand` (Liste von Navigationen im Umfang des Architect, Prozent-Kodierung, Leerraum, Trennzeichen) wird zu Knoten des Query-AST; die Option erscheint neben den Systemoptionen aus `slice-odata-url-pfad`, ohne deren Verhalten zu ändern.
- [ ] Validator gegen das Modell: Navigation vorhanden (auch geerbt), Property ist eine Navigation, Zieltyp aufgelöst; unbekannt oder keine Navigation `400`, gültig, aber nicht unterstützt `501` mit `ODR-E6001`; kein Zugriff auf Daten.
- [ ] Randformen des Optionstextes (leer, doppelte Option, doppelte Navigation, Pfad über mehrere Segmente, `*`, `$ref`, Tiefe, Länge, Groß- und Kleinschreibung): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Parser- und Validator-Contract-Test (`LH-FA-16.a`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Modus `mock` erreicht es erst mit `slice-bootstrap-m3`; die Antwort auf `$expand` zeigt sich erst mit `slice-expand-ausfuehrung`).
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
| `src/hexagon/model/` | update | Knoten des Query-AST für `$expand` (technologieunabhängig) |
| `src/hexagon/ports/outbound/` | update | Schnittstelle der Codec-Gruppe zur OData-URL, falls der Architect sie für den Validator erweitert |
| `src/adapters/driven/odata/urlparser/` | update | Parser und Validator der Option |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-16` je Navigation und je Fehlerklasse; Generator über die Klasse der Optionstexte |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Parser liest fremde Bytes (den Text der URL): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Optionstexte mit Escape-Folgen entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis in der Spezifikation, Neues unter einer künftigen Nummer, die der Architect vergibt):

1. Unterstützter Umfang von `$expand`: eine Navigation, Liste, Pfad über mehrere Navigationen (`a/b`), `*`, `$ref`, `$levels`, Typ-Cast; was darüber hinausgeht, ist `501`.
2. Verschachtelte Optionen: welche (wenn überhaupt) und mit welcher Grammatik; die Entscheidung bestimmt, ob `slice-odata-url-filter` und `slice-edm-ausdruecke` Voraussetzung sind (§4).
3. Grammatik der Option (Trennzeichen, Klammern, Leerraum, Kodierung) und Grenzen (Tiefe, Länge, Zahl der Navigationen) mit Grund und Folge im Betrieb.
4. Fehlerklassen und Rangfolge: `400` gegen `501`, unbekannte Navigation gegen Property ohne Navigation, Navigation eines Subtyps, doppelte Navigation; Fehlerobjekt und Codes (bestehende Codes oder neue, die der Architect vergibt).
5. Form des AST-Knotens und seine Stelle im Query-AST, so dass `slice-expand-ausfuehrung` ihn ohne Änderung des Parsers liest.

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

**Start** (`next` → `in-progress`): `slice-odata-url-pfad` liegt in `done/`; nennt der Architect `$filter` in verschachtelten Optionen, auch `slice-odata-url-filter` und `slice-edm-ausdruecke`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-odata-url-pfad` (Typen des Query-AST, Paket `urlparser`, Zerlegung von `raw_query` in Systemoptionen und die Fehlerbildung `400`/`501`; dieser Slice erweitert sie und kann sie nicht vorwegnehmen). Das Modell mit den Navigation-Properties (`SPEC-002`, `SPEC-125` Nr. 11) liegt in `done/` (`slice-odata-codec-csdl`). Der Welle-Trigger (`welle-mock` in `done/`) ist für diesen Slice Reihenfolge-Vorsicht (Priorisierung nach den Muss-Wellen), keine Code-Abhängigkeit: ein Start davor ist eine Entscheidung des Users, nicht dieses Plans.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Query-Interpreter, JSON-Schreiber, Workspace und Composition Root (der Slice liest Text und prüft gegen das Modell, keine Daten), M2 und `slice-bootstrap-m3` (das Binary ändert sich erst mit der Ausführung), `slice-expand-ausfuehrung` (Folge-Slice), `slice-batch-parser` und die anderen Wellen. Das heißt: **unabhängig von `MockUseCase` und von M2**; es genügt `slice-odata-url-pfad` aus `welle-mock`, und der braucht seinerseits nur das Modell (`done/`). Die Entscheidung des Architect zur Datenbasis (`welle-expand-batch` §6) trifft diesen Slice nicht.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, Leser fremder Bytes haben in der Formel keinen Zuschlag mehr, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): der Umfang von `$expand` (Auftrag 1 und 2) ist nicht entschieden, oder `slice-odata-url-pfad` hat die Form des Query-AST anders geschnitten, als dieser Plan annimmt.

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

- Die Menge der Optionstexte ist unbeschränkt (jede Zeichenfolge) — **Ausgang:** offen; Prüfgrenze und ihre Folge im Betrieb als akzeptiertes Negativ oder Generator über die Klasse (`AGENTS.md` §3.17).
- Umfang der verschachtelten Optionen: ein Umfang über `$select`, `$orderby`, `$top`, `$skip`, `$count` hinaus zieht `slice-odata-url-filter` und `slice-edm-ausdruecke` als Voraussetzung nach und verschiebt den Start dieses Slice hinter sie — **Ausgang:** offen; Architect (Auftrag 2), der Start-Trigger in §4 wird dann berichtigt.
- Form des Query-AST aus `slice-odata-url-pfad` ist beim Schreiben dieses Plans nicht entschieden — **Ausgang:** offen; Architect (Auftrag 5), Abgleich beim Start.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Option fehlt | leerer Wert | unbekannte Navigation | fremder Wert (Property ohne Navigation, anderer Typ) | Pfad, `*`, `$ref`, verschachtelte Optionen | Wiederholung (Option, Navigation) | Abbruch (Kontext) |
|---|---|---|---|---|---|---|---|
| Parser | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Validator gegen das Modell | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für die Fakes der Codec-Ports; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
