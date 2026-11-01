# Slice slice-v2-query: Query-Semantik des V2-Umfangs und expliziter Kompatibilitätsfehler

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-QA-05`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-170` · `SPEC-173` · `SPEC-174` · `SPEC-177` · `SPEC-109` Nr. 2 und 5 · `SPEC-230` Nr. 8 · `LH-FA-17.a` (versionierter Adapter für Query-Semantik, keine Vermischung von V2 und V4) · `LH-FA-17` Negative (nicht unterstützte V2-Eigenschaft: expliziter Kompatibilitätsfehler) · `LH-FA-08.a` und `LH-FA-08.b` (V4-Semantik der Abfrageoptionen, auf die der AST abbildet) · `SPEC-022` (`ODR-E6001`) · `SPEC-049` · `SPEC-002` · `ARC-001`, `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
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

**Ziel:** Der Parser der V2-URL hinter dem Port der ADR überführt Pfad und `raw_query` einer V2-Abfrage im dokumentierten V2-Umfang (Standardabfrage: Collection, Key, `$count`, `$filter`, `$orderby`, `$top`, `$skip`, `$select`, `$inlinecount` und die V2-Schreibweisen der Literale) in den Query-AST des Domain-Modells, so dass der Query-Interpreter aus `welle-mock` ihn unverändert ausführt; jede V2-Eigenschaft außerhalb des Umfangs (etwa `$expand`, `$links`, `$format=atom`, Funktionsimporte, `$batch`, Mutationen) ergibt einen expliziten Kompatibilitätsfehler mit eigenem Code und nie eine scheinbar gültige Antwort (`LH-FA-17`, Negative; `LH-RB-01`). Der Slice liest Text und kennt das Modell nicht (`URLCodec.ParseURL(pfad, rawQuery)`, `SPEC-160` Nr. 9, `SPEC-173`); die Prüfung gegen das Modell ist Sache des Dienstes, der die Query ausführt. Er schreibt keine Antwort. Der Kompatibilitätsfehler ist `ODR-E6006` (HTTP 501, `SPEC-174`, `SPEC-177`); die Ursache `methode` erzeugt der Mock (`slice-v2-bootstrap`), nicht der Parser.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Metadaten-Leser und Payload-Codec — `slice-v2-metadaten` und `slice-v2-payload`; dieser Slice braucht das V2-Modell und liest und schreibt keine Payloads.
- Wahl des Adapters je Service, Konfiguration, Route und Rundlauftest — `slice-v2-bootstrap`.
- Ausführung der Abfrage — der Query-Interpreter aus `slice-query-interpreter` (`welle-mock`); ein V2-Konstrukt, das der AST nicht tragen kann (V4-Semantik), ist hier ein Kompatibilitätsfehler und ändert den Interpreter nicht (V4-Pfad bleibt unberührt, `welle-odata-v2` §6).
- V2-Mutationen (POST, MERGE, PUT, DELETE) — `welle-odata-v2` §6; was die Spezifikation nicht nennt, ergibt den Kompatibilitätsfehler (der Parser erkennt die Methode und lehnt sie ab).
- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`.

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

- [ ] Parser: Standardabfrage des V2-Umfangs (Ressource, Key, Systemoptionen, `$filter` mit den V2-Literalen und -Operatoren des Umfangs, `SPEC-173`) wird zum Query-AST des Domain-Modells; ungültig `ODR-E6005` (400). Der Test über beide Seiten (Parser plus Interpreter) liegt im Rundlauf von `slice-v2-bootstrap`.
- [ ] Kompatibilitätsfehler: jede V2-Eigenschaft außerhalb des Umfangs (Option, Operator, Funktion, Methode, Pfadsegment) ergibt den Fehler `ODR-E6006` (`*model.KompatFehler`, `SPEC-174`), nie eine scheinbar gültige Antwort und nie `404` oder die V4-Meldung; der Fehler nennt die Eigenschaft.
- [ ] Randformen des Lesers fremder Bytes (leere und doppelte Option, Prozent-Kodierung, V4-Syntax am V2-Parser, Literale in V2-Schreibweise, Länge, Tiefe, Wiederholung): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der URLs.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary (der Code des Kompatibilitätsfehlers): Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Ist `slice-handbuch-meldungskatalog` vorher in `done/`, trägt dieser Slice den Eintrag von `ODR-E6006` im Handbuch; sonst trägt ihn `slice-v2-bootstrap`, das auch das Kapitel zu V2 trägt (der Planner legt es beim Start nach der Reihenfolge fest und trägt es beim Nehmer zuerst ein).
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
| `src/adapters/driven/odata/urlparserv2/` | neu | Parser (`SPEC-173`), `urlparserv2.New(opts ...Option)`, `WithFilter` |
| `src/hexagon/model/` | update | Fehlerwert `KompatFehler` und `ErrKompatibilitaet`, Code `ODR-E6006` in `model.AlleCodes()` (36); der Query-AST bleibt unverändert |
| `src/hexagon/ports/outbound/abbildung.go` und die Tests der Zahl der Codes | update | Abbildung `ErrKompatibilitaet` → `ODR-E6006` (`SPEC-109` Nr. 5), 36 Reihen von `odr_errors_total` (`SPEC-230` Nr. 8) |
| Tests | neu | Fälle von `LH-FA-17` je Option und je Fehlerklasse; Generator über die Klasse der URLs |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Parser liest fremde Bytes (den Text der URL): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); URLs mit Escape-Folgen entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Auftrag an den Architect:** erledigt ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), `Accepted`). Auftrag 1 steht in `SPEC-170` Nr. 2 und 3, 2 in `SPEC-173` Nr. 3 bis 5, 3 in `SPEC-174` Nr. 1 und 4 (Code `ODR-E6006`, Status 501), 4 in `SPEC-173` Nr. 4 und 5, 5 in `SPEC-174` Nr. 6. Der Text der ursprünglichen Fragen bleibt zur Nachprüfung stehen:

1. Der V2-Umfang der Abfrage (Teil des dokumentierten V2-Umfangs aus `slice-v2-metadaten`, Auftrag 1): Optionen, Operatoren, Funktionen (`substringof`, `startswith`, …), Literale (`datetime'…'`, `guid'…'`, Suffixe `L` und `M`), `$inlinecount`, `$count` als Pfad, Key-Syntax.
2. Abbildung auf den Query-AST: welche V2-Konstrukte auf bestehende Knoten fallen und welche der AST nicht trägt (sie sind Kompatibilitätsfehler); wo die Typregeln der V2-Typen liegen (`slice-edm-ausdruecke` ist V4-typisiert, die Abbildung der V2-Typen darauf steht in der Spezifikation, nicht im Code).
3. Kompatibilitätsfehler: Code, Status (vorgeschlagen ist die Linie von `ODR-E6001`, `501`; die Entscheidung trifft der Architect), Meldungstext mit der Eigenschaft, Rangfolge gegen `400` bei mehreren Verstößen.
4. Grenzen (Länge, Tiefe der Ausdrücke) mit Grund und Folge im Betrieb; Umgang mit V4-Syntax am V2-Parser.
5. Version in der Anfrage: `DataServiceVersion`, `MaxDataServiceVersion` und `OData-Version` und was bei einem Widerspruch zur Version des Service gilt.

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

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` (erfüllt), und der Vertrag samt Randformen steht in `SPEC-173` und `SPEC-174` (erfüllt); die Tabelle *Zusage · Fänger* füllt der Architect vor dem ersten Code-Commit (`AGENTS.md` §3.12). Der Slice ist startbar ohne `welle-mock`.

- **Code-Abhängigkeit, mit Grund:** `slice-odata-url-pfad` und `slice-odata-url-filter` (beide in `done/`: Gerüst des Parsers und Query-AST, auf die V2 abbildet), Modell und Fehlermodell (`done/`). Der Parser kennt das Modell nicht (`SPEC-173`); `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-edm-ausdruecke` und `slice-query-interpreter` sind keine Voraussetzung, der Test über beide Seiten (Parser plus Interpreter) wandert in den Rundlauf von `slice-v2-bootstrap`. Der Schnitt von Pfad und Optionen (`SPEC-173` Nr. 1 bis 4, 6, 7) und `$filter` (Nr. 5), die allein die Literale (Nr. 3) und die Einhängestelle `WithFilter` teilen, ist bei Bedarf möglich, nach dem Vorbild von `slice-odata-url-pfad` und `slice-odata-url-filter`.
- **Was der Slice nicht braucht — und warum:** `welle-mock` ganz, `MockUseCase`, Workspace und HTTP-Adapter (der Parser liest Text; der Rundlauf gehört `slice-v2-bootstrap`), M2 und `slice-bootstrap-m3` (das Binary ändert sich erst mit der Wahl des Adapters), `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`. Der V2-Parser bildet auf den AST der Mock-Slices ab und hängt an dessen Form, nicht am Use Case.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, Leser fremder Bytes haben in der Formel keinen Zuschlag mehr, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`, oder der AST trägt ein Konstrukt des Umfangs nicht.

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

- Die Menge der URLs ist unbeschränkt (jede Zeichenfolge) — **Ausgang:** offen; Generator über die Klasse mit Orakel und Zählprüfung oder Prüfgrenze als akzeptiertes Negativ mit Grund und Folge im Betrieb (`AGENTS.md` §3.17).
- Der AST der Mock-Slices ist beim Schreiben dieses Plans nicht entschieden; ein V2-Konstrukt, das er nicht trägt, wird Kompatibilitätsfehler oder verlangt einen neuen Knoten (Änderung am V4-Pfad, `welle-odata-v2` §6) — **Ausgang:** entfallen: der AST liegt in `done/` (`SPEC-165`), `SPEC-173` Nr. 5 bildet auf ihn ab; ein Konstrukt ohne Knoten ist `ODR-E6006`.
- Zahl der Fehlerfälle: der Kompatibilitätsfehler ist eine Allaussage über jede V2-Eigenschaft außerhalb des Umfangs — **Ausgang:** offen; der Fänger läuft als Schleife über Optionen, Operatoren, Funktionen, Methoden und Pfadsegmente der V2-Spezifikation, die der Umfang nicht nennt, nicht als Liste der Fälle, die einfielen (`AGENTS.md` §3.15, §3.17).
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Option fehlt oder leer | unbekannte Option, Funktion oder Methode (außerhalb des Umfangs) | fremde Syntax (V4 am V2-Parser, anderes Format) | Literale in V2-Schreibweise, Typfehler | Wiederholung (doppelte Option), über Grenze (Länge, Tiefe) | Abbruch (Kontext) |
|---|---|---|---|---|---|---|
| Ressource: Collection, Key, `$count` | `/` allein: `pfad` (`SPEC-173` Nr. 2); keine Query: gültig | Rest-Segment, `$links`, `$batch`: `pfad` (`SPEC-173` Nr. 2) | `Set(1)/$count`: `pfad_form`; `$value` als erstes Segment: `segment` (`SPEC-173` Nr. 2) | `SPEC-173` Nr. 3 (Tabelle, `literal_typ`, `literal`) | `SPEC-173` Nr. 1 (`SPEC-160` Nr. 6: Teile, Länge) | gilt nicht: `ParseURL` hat keinen Kontext (`SPEC-160` Nr. 1) |
| Option `$filter` | `$filter=` leer: `wert` (`SPEC-173` Nr. 4, `SPEC-162`) | Funktion, Arithmetik, Pfad: `filter` (`SPEC-173` Nr. 5) | `contains(`, `any(`: `filter`; Literal ohne Präfix: `wert` (`SPEC-173` Nr. 5) | `SPEC-173` Nr. 3 und 5 | doppelt: `option_doppelt`; Tiefe und Länge: `SPEC-173` Nr. 1 (`SPEC-164`, `SPEC-160` Nr. 6) | gilt nicht (wie oben) |
| Optionen `$orderby`, `$top`, `$skip`, `$select`, `$inlinecount` | leer: `wert` (`SPEC-173` Nr. 4) | `$expand`, `$skiptoken`, `$callback`: `option`; `$format=xml` und `atom`: `format` (`SPEC-173` Nr. 4) | `$count`, `$search`, `$apply` …: `option` (`SPEC-173` Nr. 4) | `$top=abc`, `$inlinecount=foo`: `wert` (`SPEC-173` Nr. 4) | doppelt: `option_doppelt`; Grenzen `SPEC-160` Nr. 6 (`SPEC-173` Nr. 1) | gilt nicht (wie oben) |
| Kompatibilitätsfehler | gilt nicht: er hat immer eine Ursache (`SPEC-174` Nr. 2) | die Menge: `SPEC-170` Nr. 3, Ursachen `SPEC-174` Nr. 2 | V4-Syntax am V2-Parser: `SPEC-173` Nr. 4 und 5 | `literal_typ` (`SPEC-174` Nr. 2, `SPEC-173` Nr. 3) | Rangfolge `SPEC-174` Nr. 4 und `SPEC-173` Nr. 6 | gilt nicht (wie oben) |

**Zusatztabelle** (je Paar von Verstößen an derselben URL; `AGENTS.md` §3.16): Kompatibilitätsfehler und `400` an derselben URL, zwei Eigenschaften außerhalb des Umfangs, Version der Anfrage gegen Version des Service.

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| Kompatibilitätsfehler und `400` an derselben URL | `ODR-E6005` geht vor, gleich wo er steht (`SPEC-173` Nr. 1 und 6) | ein Fall je Paar (Eigenschaft der Menge mal Verstoß der Klasse `ungueltig`) |
| zwei Eigenschaften außerhalb des Umfangs | die erste in der Lesefolge (`SPEC-173` Nr. 1 und 6) | Schleife über Paare aus Pfad, Option, Filter und Literal |
| Version der Anfrage gegen Version des Service | die Kopfzeilen ändern nichts (`SPEC-174` Nr. 6, `SPEC-173` Nr. 7) | gleiche Antwort mit und ohne `OData-Version: 4.0`, `MaxDataServiceVersion: 1.0` |

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
