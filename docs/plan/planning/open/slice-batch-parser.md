# Slice slice-batch-parser: Parser und Validator für $batch-Requests, leerer Batch

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

**Berührte Spec-Stellen:** `LH-FA-16.a` (Erweiterungspunkt `$batch` mit atomaren Changesets, eigener Parser- und Validator-Contract-Test, unbekannte Operationen `501`) · `SPEC-049` (Lastenheft §5: `$batch` nicht im ersten Abnahmeumfang) · `SPEC-022` (`ODR-E6001`) · `SPEC-109` (Fehlermodell) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkte).
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

**Ziel:** Ein Port der Codec-Gruppe liest den Body eines `$batch`-Requests (Hülle nach dem Format, das der Architect festlegt, mindestens `multipart/mixed` des OData-V4-Batch) in typisierte Domain-Werte: eine Folge von Teilen, jeder Teil entweder ein Einzelrequest (Methode, Ziel, Header, Body-Bytes, `Content-ID`) oder ein Changeset (Folge von Änderungsrequests). Der Validator prüft die Struktur (erlaubte Methoden je Ebene, eindeutige und auflösbare `Content-ID`, Grenzen); ein leerer Batch ergibt eine gültige leere Folge, ein Verstoß einen Fehler mit Ursache, eine gültige, aber nicht unterstützte Form `501` mit `ODR-E6001`. Der Parser liest Bytes und entscheidet nichts über Zustand, Ziel oder Body der Einzelrequests (`LH-FA-16.a`: Parser- und Validator-Contract-Test).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ausführung der Einzelrequests, die atomare Workspace-Transaktion eines Changesets, die Antwort des Batch und die Route `POST /$batch` — `slice-batch-changesets` (Executor- und Serializer-Contract-Test von `LH-FA-16.a` liegen dort); dieser Slice liefert die geprüfte Struktur.
- Parsen der URL und des Bodys eines Einzelrequests — die Ausführung nutzt dafür `urlparser` und den Leser der Client-Bodies aus `welle-mock`; der Parser hält Ziel und Body als Text und Bytes.
- Aufzeichnen und Abspielen von Batch-Requests — der Record-Pfad leitet sie als HTTP unverändert weiter, der Replay-Pfad bedient sie unabhängig vom Parser (`welle-expand-batch` §6, `LH-FA-16.a`).
- Herstellerspezifische Batch-Erweiterungen, Actions und Functions in Batch-Teilen — Out-of-Scope von `LH-FA-16` bzw. Lastenheft §5; sie ergeben `501`.

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

- [ ] Lesen der Hülle: Grenze (`boundary`) aus dem `Content-Type`, Teile, Changesets, Einzelrequest mit Methode, Ziel, Header, Body und `Content-ID`; ein leerer Batch ergibt die leere Folge; Ergebnis und Fehler hängen nicht davon ab, wie der Strom in Reads zerlegt ist.
- [ ] Validator: erlaubte Methoden je Ebene (außerhalb und innerhalb eines Changeset), `Content-ID` eindeutig und nur rückwärts referenzierbar, Grenzen (Größe, Zahl der Teile, Tiefe) mit Grund und Folge im Betrieb; Verstoß mit Ursache, nicht Unterstütztes `501` mit `ODR-E6001`.
- [ ] Randformen der Hülle (fehlende oder leere Grenze, Teil ohne Content-Type, Teil ohne Request, Kopfzeilen mit Falten, Zeilenenden, Epilog, Wiederholung): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Parser- und Validator-Contract-Test (`LH-FA-16.a`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (die Route `/$batch` erreicht es erst mit `slice-batch-changesets`).
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
| `src/hexagon/model/` | update | Typen für Batch, Teil, Changeset, Einzelrequest und der Fehler |
| `src/hexagon/ports/outbound/` | update | Schnittstelle der Codec-Gruppe zum Lesen von Batch-Bodies |
| `src/adapters/driven/odata/` (neues Paket für den Batch-Leser; Name nach Architect) | neu | Leser und Validator |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-16` je Teil und je Fehlerklasse; Generator über die Klasse der Hüllen, Zerlegung des Stroms in Reads |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Leser liest fremde Bytes (ein Format, das dieses Repo nicht schreibt): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Bodies mit Escape-Folgen oder Rohbytes entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis in der Spezifikation, Neues unter einer künftigen Nummer, die der Architect vergibt):

1. Format: nur `multipart/mixed` (OData V4.0) oder auch der JSON-Batch (V4.01); was darüber hinausgeht, ist `501`.
2. Leser: die Standardbibliothek (`mime/multipart`, `net/http`) oder ein eigener Leser; die Zusagen dürfen nicht am Verhalten der Bibliothek hängen (Vorbild `SPEC-125`, `SPEC-126`: Zerlegung des Stroms, Größe, doppelte Namen und Zeilenenden prüft der Leser selbst). Gilt Fremdbibliothek, wäre das eine ADR.
3. Grenzen mit Grund und Folge im Betrieb: Größe des Bodys (Quelle der Grenze, eine eigene oder `MAX_CAPTURE_BODY`), Zahl der Teile, Tiefe (Changeset im Changeset), Länge der Kopfzeilen.
4. Validator: Methoden je Ebene, `Content-ID` (Eindeutigkeit, Referenz `$<id>` nur rückwärts, Auflösung erst bei der Ausführung), Verhalten bei `Prefer: respond-async`, `continue-on-error` und Abhängigkeiten (`dependsOn`).
5. Fehler: Ursachen und Codes (bestehende oder neue, die der Architect vergibt), Rangfolge bei mehreren Verstößen, Ort in der Datei (Teil und Zeile ohne Abhängigkeit vom Lexer).
6. Form der Domain-Werte, so dass `slice-batch-changesets` sie ohne Änderung des Parsers liest.

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

**Start** (`next` → `in-progress`): keine Slice-Abhängigkeit über `done/` hinaus. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** keine auf einen offenen Slice. Der Leser braucht das Fehlermodell (`SPEC-109`, `slice-fehlermodell-konfiguration` in `done/`), die Konvention der Codec-Ports (`SPEC-125`, `SPEC-126` als Vorbild für Zerlegung des Stroms, Grenzen und Fehler, `done/`) und das Domain-Modell; er ruft weder `urlparser` noch Workspace, Repositories, Interpreter oder `MockUseCase`. Der Welle-Trigger (`welle-mock` in `done/`) ist für diesen Slice Reihenfolge-Vorsicht (Priorisierung nach den Muss-Wellen), keine Code-Abhängigkeit: ein Start davor ist eine Entscheidung des Users, nicht dieses Plans.
- **Was der Slice nicht braucht — und warum:** `MockUseCase` und `WithWorkspaceTransaction` (die Ausführung liegt in `slice-batch-changesets`), die Parser der OData-URL und der Client-Bodies (Ziele und Bodies bleiben hier Text und Bytes), HTTP-Adapter, Composition Root, M2 und `slice-bootstrap-m3` (das Binary ändert sich erst mit der Route), `slice-expand-parser` und die anderen Wellen. Das heißt: **unabhängig von `welle-mock` und von M2**, soweit der Code betroffen ist.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, Leser fremder Bytes haben in der Formel keinen Zuschlag mehr, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil und bei einem Hüllenformat der größere Teil der Schätzung.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt; ein Schnitt vor dem Code läge zwischen Hülle (Teile und Grenzen) und Validator (Methoden, `Content-ID`), sofern sie keinen Mechanismus teilen.
- `in-progress` → `open` (blockiert — Carveout?): Format, Leser oder Grenzen (Auftrag 1 bis 3) sind nicht entschieden, oder der Architect verlangt eine Fremdbibliothek ohne ADR.

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

- Die Menge der Hüllen ist unbeschränkt (jede Bytefolge, jede Zerlegung des Stroms in Reads) — **Ausgang:** offen; Generator über die Klasse (Hüllen, Zerlegungen, Zeilenenden) mit Orakel und Zählprüfung oder Prüfgrenze als akzeptiertes Negativ mit Grund und Folge im Betrieb (`AGENTS.md` §3.17).
- Das Verhalten einer Bibliothek (`mime/multipart`) als Grundlage einer Zusage: Zeilenenden, Kopfzeilen mit Falten, doppelte Grenze — **Ausgang:** offen; Architect (Auftrag 2), Fänger je Zusage, die der Leser selbst prüft.
- Größe eines Batch-Bodys: der Request-Body-Limit im HTTP-Adapter (`slice-http-adapter`, Risiko dort *Request-Größe*) und die Grenze des Lesers sind zwei Quellen — **Ausgang:** offen; Architect (Auftrag 3), eine Quelle benennen.
- Zusammenspiel mit der Ausführung: Form der Domain-Werte (Auftrag 6) — **Ausgang:** offen; Abgleich mit `slice-batch-changesets` beim Start des zweiten Slice.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation, auch `Close` und jeden Read (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Body fehlt oder leer | Grenze fehlt, leer oder ungültig | unbekannter Teiltyp oder Methode | fremde Form (JSON, absolute URL, anderer Host) | über Grenze (Größe, Teile, Tiefe) | Wiederholung (`Content-ID`, Changeset) | Abbruch (Kontext, Strom) |
|---|---|---|---|---|---|---|---|
| Parser: Hülle und Teile | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Parser: Einzelrequest (Zeile, Header, Body) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Validator: Methoden und Changesets | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Validator: `Content-ID` und Referenzen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

**Zusatztabelle** (je Paar von Verstößen am selben Teil; `AGENTS.md` §3.16): ungültige Grenze und ungültiger Teil, doppelte `Content-ID` und unzulässige Methode, Größe über der Grenze und Verstoß dahinter.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für Fakes und Hüllen von Strömen (jede Schnittstelle des Originals, jeder Read gezählt); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
