# Slice slice-v2-payload: Codec für V2-Payloads, lesen (Import), schreiben (Mock), leeres Ergebnis

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-17.a` (versionierter Adapter für V2-Payloads, keine Vermischung von V2 und V4) · `LH-FA-17` Boundary (leeres Ergebnis bleibt im V2-Antwortformat konsistent) · `SPEC-126` (Vertrag des V4-Lesers als Vorbild; Schreibweisen der Typen Nr. 10) · `SPEC-127` Nr. 8 (Auswahl der Antworten durch den Import) · `SPEC-002` · `SPEC-006` (Fehlerobjekt) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
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

**Ziel:** Der Codec für V2-Payloads hinter den Ports, die die ADR zum versionierten Adapter festlegt, **liest** die Antworten eines V2-Dienstes auf Entity- und Collection-Abfragen (für den Import) in typisierte Entities gegen das V2-Modell aus `slice-v2-metadaten` und **schreibt** Antworten des Mock im V2-Format: Entity, Collection und Anzahl, das leere Ergebnis in der Form, die das V2-Antwortformat für eine leere Menge verlangt (`LH-FA-17`, Boundary), und das Fehlerobjekt. Lesen und Schreiben sind Gegenstücke: was der Schreiber ausgibt, liest der Leser zurück. Das Format der V2-Payloads (JSON, Atom) steht im dokumentierten V2-Umfang des Architect; was er nicht nennt, ergibt den Kompatibilitätsfehler aus `slice-v2-query`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Metadaten-Leser für V2 und V2-Modell — `slice-v2-metadaten` (Voraussetzung); dieser Slice liest und schreibt Entities gegen das Modell, er bildet es nicht.
- Query-Semantik und Kompatibilitätsfehler für Abfragen — `slice-v2-query`.
- Wahl des Adapters je Service, Konfiguration und die Anpassung des Imports an einen Codec je Version (Auswahl der Antworten, Ketten über Folgelinks) — `slice-v2-bootstrap` bzw. ein Slice davor, den der Planner dort anlegt (§6); die Auswahl und die Ketten des Imports liegen in `done/` (`SPEC-127` Nr. 8 und 12) und sind keine Adresse für diesen Slice.
- V2-Mutationen und das Lesen von Client-Bodies für V2 — `welle-odata-v2` §6; was die Spezifikation nicht nennt, ergibt den Kompatibilitätsfehler.
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

- [ ] Lesen: Antworten auf Entity- und Collection-Abfrage in typisierte Entities gegen das V2-Modell (Hülle, Typen in der Schreibweise von V2, Anzahl, Folgelink, nicht gedeutete Konstrukte als übersprungen gemeldet); kanonische Schreibweise je Typ; V4-Leser und V4-Schreiber unberührt (ihre Tests laufen unverändert grün).
- [ ] Schreiben: Entity, Collection, Anzahl und Fehlerobjekt im V2-Format; leeres Ergebnis konsistent im V2-Format; Rundlauf Schreiben → Lesen liefert dieselben Entities.
- [ ] Randformen des Lesers fremder Bytes (leer, falsche Hülle, V4-Antwort am V2-Leser, Typ-Schreibweisen, Zahlen jenseits von `float64`, doppelte Namen, Größe, Tiefe, Zerlegung des Stroms in Reads): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der Bodies.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (ein V2-Dienst erreicht es erst mit `slice-v2-bootstrap`, das auch das Kapitel zu V2 trägt).
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
| `src/hexagon/ports/outbound/` | update | Schnittstellen der Codec-Gruppe je Version nach der ADR |
| `src/adapters/driven/odata/` (neues Paket für den V2-Codec; Name nach Architect) | neu | Leser und Schreiber der V2-Payloads |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect) |
| Tests | neu | Fälle von `LH-FA-17` (Boundary, Happy Path) für Lesen und Schreiben; Generator über die Klasse der Bodies; Rundlauf |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Leser liest fremde Bytes (ein Format, das dieses Repo nicht schreibt): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Bodies mit Escape-Folgen oder Rohbytes entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Auftrag an den Architect** (vor dem ersten Code-Commit; die Aufteilung der Ports je Version steht in der ADR, Status `Accepted`; Neues in der Spezifikation unter einer künftigen Nummer, die der Architect vergibt):

1. Format der V2-Payloads im dokumentierten V2-Umfang (JSON, Atom oder beide) und die Verhandlung über `Accept`; was darüber hinausgeht, ist Kompatibilitätsfehler.
2. Hülle und Mitglieder der Antworten (Entity, Collection, Anzahl, Folgelink, Verweise auf Navigationen), ihre Schreibweisen je Typ und kanonische Form; Zuordnung zu den Werten des Modells, so dass der Import (`SPEC-127`) sie wie die V4-Werte verarbeitet.
3. Leeres Ergebnis: die Form der leeren Menge im Antwortformat und das Verhalten für `$count` und Anzahl.
4. Fehlerobjekt im V2-Format und Abbildung der Fehler des Mock (`SPEC-006`); Header der V2-Antwort (Versionskennzeichen, `Content-Type`) und wer sie setzt: der Codec liefert die Kennzeichnung der Darstellung, der Use Case oder der Adapter schreibt den Header.
5. Grenzen und Fehler des Lesers (Größe aus `MAX_CAPTURE_BODY`, Tiefe, Länge von Zahlen; Ursachen und Code, Vorbild `SPEC-126` Nr. 5 und 14), Rangfolge und Ort.
6. Wie ein V4-Body am V2-Leser (und umgekehrt) endet: Fehler der Seite, die ihn liest, ohne Vermischung der Modelle.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect mit dem Vertrag, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (V2 schreibt Zeitpunkte als `\/Date(…)\/`) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-v2-metadaten` und `slice-odata-json-schreiben` liegen in `done/`, und die ADR des Architect zum versionierten Adapter ist `Accepted`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-v2-metadaten` (das V2-Modell ist der Parameter von Lesen und Schreiben; ohne es gibt es nichts, wogegen zu typisieren ist), `slice-odata-json-schreiben` (der Port zum Schreiben von OData-JSON und die Darstellung der Typen, die der V2-Schreiber ersetzt bzw. neben die er tritt). Der **Lese-Teil** braucht nur `slice-v2-metadaten` und den Port `ODataCodec` (`done/`); der **Schreib-Teil** braucht den Port aus `welle-mock`. Ein Schnitt in zwei Slices ist nicht vorab benannt: Lesen und Schreiben teilen die Abbildung der V2-Typen (kanonische Schreibweise gegen Schreibweise im Body), das Teil wäre kein Teil „ohne geteilten Mechanismus“ (`AGENTS.md` §3.14); der Planner prüft es beim Start, wenn die Tabelle *Zusage · Fänger* vorliegt.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Workspace, Composition Root, M2 und `slice-bootstrap-m3` (der Codec arbeitet auf Bytes, Modell und Entities; der Rundlauf über Import und Mock gehört `slice-v2-bootstrap`), `slice-v2-query`, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`. Das heißt: **der Lese-Teil ist bedingt ohne `welle-mock` und M2 startbar (nach `slice-v2-metadaten` und der ADR), der Schreib-Teil nicht.**
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, Leser fremder Bytes haben in der Formel keinen Zuschlag mehr, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil; der V4-Leser `json` (Zeile *json* in `SPEC-128` Nr. 2) ist die Messlatte für die Erwartung des Lesens.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt; der Schnitt vor dem Code, falls die Schnittfrage von `SPEC-128` ihn verlangt, läge zwischen Lesen und Schreiben, sofern sie keinen Mechanismus teilen (sonst steht der Grund in §4).
- `in-progress` → `open` (blockiert — Carveout?): Format der Payloads oder Header der V2-Antwort sind nicht entschieden (Auftrag 1 und 4), oder die ADR teilt die Ports anders, als dieser Plan annimmt.

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

- Der Import eines V2-Dienstes braucht mehr als den Leser: die Auswahl der Antworten (`SPEC-127` Nr. 8: heute `application/json` mit `odata.metadata`) und die Ketten über Folgelinks und Anzahl (`SPEC-127` Nr. 12) sind V4-Regeln in `done/`; kein Slice der Welle trägt ihre Anpassung — **Ausgang:** offen; Adresse: Architect (Entscheidung in der ADR) und Planner (Slice für die Anpassung des Imports vor `slice-v2-bootstrap`, `AGENTS.md` §3.13).
- Die Menge der Bodies ist unbeschränkt (jede Bytefolge) — **Ausgang:** offen; Generator über die Klasse mit Orakel und Zählprüfung oder Prüfgrenze als akzeptiertes Negativ mit Grund und Folge im Betrieb (`AGENTS.md` §3.17).
- Das Format der V2-Antworten (JSON oder Atom) steht im Lastenheft nicht; der Referenzdienst der Welle legt es fest — **Ausgang:** offen; Architect (Auftrag 1).
- Zeit- und Zahlenschreibweisen von V2 (Zeitpunkte als Millisekunden, `Edm.Int64` und `Edm.Decimal` als Zeichenketten): jede Abbildung auf die kanonische Form ist verlustbehaftet oder nicht — **Ausgang:** offen; Raster, eine Zelle je Typ (Schleife über die Typen des V2-Umfangs).
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Body leer oder keine Menge | falsche Hülle (V4-Body am V2-Leser) | unbekannte Property, nicht unterstützter Typ | Typ-Schreibweisen, Zahlen jenseits `float64` | doppelter Name, über Grenze (Größe, Tiefe) | Wiederholung, Abbruch (Kontext, Strom) |
|---|---|---|---|---|---|---|
| Lesen einer Entity | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Lesen einer Collection | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Schreiben einer Entity | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Schreiben einer Collection, Anzahl, leeres Ergebnis | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Schreiben des Fehlerobjekts | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für Fakes der Codec-Ports und Hüllen von Strömen (jede Schnittstelle des Originals, jeder Read gezählt); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
