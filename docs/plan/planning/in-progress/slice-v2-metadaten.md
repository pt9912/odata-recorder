# Slice slice-v2-metadaten: Metadaten-Leser für V2 (EDMX/CSDL V2) mit Validierungsbericht und Version je Service

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-170` · `SPEC-171` · `SPEC-178` · `SPEC-179` · `LH-FA-17.a` (versionierter Adapter für V2-Metadaten, keine Vermischung von V2 und V4 in einem Service-Modell) · `LH-RB-01` (nicht Unterstütztes erscheint nie als unterstützt) · `SPEC-002` (OData-Modell) · `SPEC-125` (Vertrag des V4-Metadatenlesers als Vorbild; Nr. 6 a: ein EDMX der Versionen 1 bis 3 ist dort `version`) · `SPEC-107` Nr. 14 und 15 (`odata_version` des Service, Metadaten durch den Import) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
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

**Ziel:** Ein Leser für das EDMX/CSDL einer V2-Quelle (Teilumfang nach dem dokumentierten V2-Umfang des Architect) liefert das Modell des Dienstes als `model.ServiceModel` mit `Version()` gleich `2.0` (`SPEC-171`, Paket `csdlv2`) und einen Validierungsbericht: ausgewertet wird der Teilumfang, alles andere erscheint als Fund und nie als unterstützt (`LH-RB-01`). Das Modell trägt die Version des Dienstes (die „Version je Service“ des Slice-Titels: die Version, die das Modell nennt, nicht ihre Deklaration in der Konfiguration), und ein V2-Dokument vermischt sich nicht mit dem V4-Modell. Der V4-Leser (`csdl`) und das Verhalten eines V2-Dokuments an ihm (`ODR-E6004`, Ursache `version`) bleiben unverändert.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Payload-Codec für V2 — `slice-v2-payload-lesen` und `slice-v2-payload-schreiben`.
- Query-Semantik und Kompatibilitätsfehler — `slice-v2-query`.
- Deklaration der Version eines Service in der Konfiguration und Wahl des Adapters im Composition Root — `slice-v2-bootstrap`.
- Anpassung des Imports an einen Satz von Codecs je Dienst — `slice-v2-import-codecs` (der Nehmer trägt es zuerst ein, `AGENTS.md` §3.13; die Änderung am Vertrag von `SPEC-127` Nr. 1 liegt dort).
- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`; sie erscheinen im Bericht als Fund.
- V2-Mutationen und ihre Metadaten — `welle-odata-v2` §6; was die Spezifikation nicht nennt, ergibt den Kompatibilitätsfehler.

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

- [ ] Leser und Modell: EDMX/CSDL V2 im Teilumfang des Architect (Wurzel, Schema, Entity-Typen, Schlüssel, Properties, Assoziationen und Navigation-Properties, Container, Entity-Sets) in `model.ServiceModel` mit Version `2.0`; Namen aufgelöst; V4-Modell und V4-Leser unberührt (die Tests des V4-Lesers laufen unverändert grün).
- [ ] Validierungsbericht: jedes nicht ausgewertete Konstrukt (Funktionsimporte, komplexe Typen soweit nicht im Umfang, herstellerspezifische Annotationen und Attribute, nicht unterstützte Property-Typen) ist ein Fund; nichts Unterstütztes wird verschwiegen und nichts Unterstütztes behauptet (`LH-RB-01`).
- [ ] Verträge und Randformen eines Lesers fremder Bytes (Kodierung, Größe, Tiefe, ungültiges XML, falsche Wurzel oder Version, Verweise ohne Ziel, Wiederholung, Zerlegung des Stroms in Reads): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der Dokumente.
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
| `src/adapters/driven/odata/csdlv2/` | neu | Leser und Validierungsbericht (`SPEC-171`); kein eigenes Modell und kein neuer Port ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)) |
| Tests | neu | Fälle von `LH-FA-17` (Happy Path) für die Metadaten; Generator über die Klasse der Dokumente; Fixtures aus einem echten V2-Dienst |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Leser liest fremde Bytes (ein Format, das dieses Repo nicht schreibt): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); Dokumente mit Escape-Folgen, BOM oder Rohbytes entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Auftrag an den Architect:** erledigt ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), `Accepted`). Auftrag 1 (Umfang) steht in `SPEC-170`; 2 (Modell) in [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) und `SPEC-171` Nr. 8 (derselbe Typ mit Version `2.0`); 3 (Abbildung) in `SPEC-170` Nr. 4, `SPEC-171` Nr. 5 und 6; 4 (Dokument) in `SPEC-171` Nr. 2 und 3; 5 (Bericht) in `SPEC-171` Nr. 4, 9 und `SPEC-125` Nr. 13; 6 (Ort der Version) in `SPEC-170` Nr. 1 und `SPEC-175` Nr. 1.

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

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` und der V2-Umfang steht in der Spezifikation (`SPEC-170`, `SPEC-171`; beides erfüllt). Die Tabelle *Zusage · Fänger* füllt der Architect vor dem ersten Code-Commit (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** keine auf einen offenen Slice. Der Leser steht hinter dem Port des Metadatenlesers (`MetadataCodec`, `SPEC-125`), den `slice-odata-codec-csdl` in `done/` geliefert hat, und braucht das Modell und das Fehlermodell (`done/`); die ADR ändert den Port nicht, und das Modell ist derselbe Typ. Er ruft weder `MockUseCase` noch Workspace, Interpreter oder URL-Parser. Der Welle-Trigger (`welle-mock` in `done/`) trägt für diesen Slice nichts; er ist Reihenfolge-Vorsicht (Priorisierung nach den Muss-Wellen).
- **Was der Slice nicht braucht — und warum:** `welle-mock` und M2 (Leser und Bericht arbeiten auf Bytes und Modell; der Rundlauf über Import und Mock gehört `slice-v2-bootstrap`), `slice-v2-payload-lesen`, `slice-v2-payload-schreiben` und die späteren Slices der Welle, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`. Das heißt: **startbar ohne `welle-mock` und ohne M2, sobald [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted` ist.**
- **Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`), vorläufig aus der Grobschätzung des Architect (grob 5700 bis 6100 für *S* ohne Hebung des XML-Rahmens, gerechnet am oberen Ende): *S* ≈ 6100, Marke *M* = 100 · ⌈135·6100 / 10000⌉ = **8300**. Die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/adapters/driven/odata/csdlv2/` und die Tests dort, Tests eingeschlossen. Der Block nach `SPEC-128` Nr. 3 h folgt mit der Tabelle *Zusage · Fänger* des Architect (Teil des Architect: `pfad`, `produktcode`, `fakes`; Teil des Planner: `basis`, `schaetzung`, `marke`); bis dahin ist die Marke eine Prosa-Zeile und das Gate misst sie nicht. **Schnittfrage** (*S* ≥ 2223): Lesen und Auflösen des Dokuments teilen den Baum (die Namen werden im Baum gelesen und im Baum aufgelöst), ein Teil ohne geteilten Mechanismus lässt sich nicht abtrennen (Architect); der Grund steht damit hier. Der Architect darf beim Start die Hebung des versionsfreien XML-Rahmens als Vorlauf wählen ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) erlaubt sie, die V4-Tests bleiben unverändert); dann wird dieser Teil vor dem Code benannt und die Marke neu gesetzt.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Zahl der Marke steht in der Größenmarke oben.
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`.

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

- Der dokumentierte V2-Umfang war weder im Lastenheft noch in der Spezifikation beschrieben — **Ausgang:** entfallen: er steht in `SPEC-170`.
- Die Menge der V2-Dokumente ist unbeschränkt (jede Bytefolge, jede Zerlegung) — **Ausgang:** offen; Generator über die Klasse mit Orakel und Zählprüfung oder Prüfgrenze als akzeptiertes Negativ mit Grund und Folge im Betrieb (`AGENTS.md` §3.17).
- Echte V2-Dienste liefern Konstrukte, die der Umfang nicht kennt (herstellerspezifische Annotationen, Funktionsimporte); sie dürfen den Bericht füllen, nicht das Lesen brechen — **Ausgang:** offen; Fänger je Art des Fundes.
- Das V4-Modell darf durch den Leser nicht berührt werden (`LH-FA-17.a`): die Tests von `slice-odata-codec-csdl` laufen unverändert, und ein V2-Dokument am V4-Leser bleibt `ODR-E6004` — **Ausgang:** offen; Fänger an beiden Lesern.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation, auch `Close` und jeden Read (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Dokument leer oder ohne Schema | falsche Wurzel oder Version (V4 am V2-Leser) | unbekannter oder nicht unterstützter Typ | fremde Erweiterung (herstellerspezifisch) | Verweis ohne Ziel, Wiederholung (doppelter Name) | über Grenze (Größe, Tiefe, Summe) | Abbruch (Kontext, Strom) |
|---|---|---|---|---|---|---|---|
| Lesen des Dokuments | leer, nur Leerraum, nur BOM: `syntax`; Wurzel ohne `Schema`: gültig, leeres Modell (`SPEC-171` Nr. 1, `SPEC-125` Nr. 4) | `SPEC-171` Nr. 2 (`wurzel`, `version`) | Property bleibt im Modell, Fund `property_typ` (`SPEC-171` Nr. 4, 5) | Fund `erweiterung` (`SPEC-171` Nr. 4) | doppelter Name: `struktur` (`SPEC-171` Nr. 1, `SPEC-125` Nr. 8 e) | `SPEC-171` Nr. 1 (`SPEC-125` Nr. 5: `zu_gross`, `tiefe`, `kette`, `zu_viele_properties`) | `SPEC-171` Nr. 1 (`SPEC-125` Nr. 1, 2) |
| Auflösen der Namen und Assoziationen | leeres Modell: nichts aufzulösen, gültig (`SPEC-171` Nr. 1) | gilt nicht: Wurzel und Version stehen vor dem Auflösen fest (`SPEC-171` Nr. 2) | nicht auflösbarer Name `nicht_aufloesbar`, Schlüsseltyp `schluesseltyp` (`SPEC-171` Nr. 5, 7) | gilt nicht: Erweiterungen werden gemeldet, nicht aufgelöst (`SPEC-171` Nr. 4) | `SPEC-171` Nr. 6 und 7 (Relationship, Rollen, Typ, Sets: `nicht_aufloesbar`; `Multiplicity`: `struktur`) | `SPEC-171` Nr. 7 (`SPEC-125` Nr. 5, 9 a, 9 b) | `SPEC-171` Nr. 1 und *Ungeprüft*: das Auflösen ist nicht abbrechbar (`SPEC-125` Nr. 1) |
| Bericht | leere Menge, nicht `nil` (`SPEC-171` Nr. 9) | Fehler: Bericht Nullwert (`SPEC-171` Nr. 9) | Fund `property_typ` (`SPEC-171` Nr. 4) | Fund `erweiterung`, Ausnahme der drei `m:`-Attribute (`SPEC-171` Nr. 4) | gilt nicht: ein Fehler ist kein Fund (`SPEC-171` Nr. 9) | Fehler: Bericht Nullwert (`SPEC-171` Nr. 9) | Fehler: Bericht Nullwert (`SPEC-171` Nr. 9) |

**Zusatztabelle** (je Paar von Verstößen am selben Dokument; `AGENTS.md` §3.16): ungültiges XML und falsche Version, Verweis ohne Ziel und nicht unterstützter Typ, Größe über der Grenze und Verstoß dahinter.

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| ungültiges XML und falsche Version | die Wurzel entscheidet, sobald ihr Start-Tag gelesen ist; Müll davor ist `syntax` (`SPEC-171` Nr. 2) | ein Fall je Paar im Rangfolge-Test des Lesers (Name beim Start) |
| Verweis ohne Ziel und nicht unterstützter Typ | `nicht_aufloesbar` gewinnt; der Fund `property_typ` erscheint nicht, der Bericht ist Nullwert (`SPEC-171` Nr. 7, 9) | ein Fall je Paar |
| Größe über der Grenze und Verstoß dahinter | `zu_gross` am Byte `grenze + 1`, davor entscheidet der Verstoß (`SPEC-171` Nr. 1, `SPEC-125` Nr. 5) | ein Fall je Paar |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für Fakes und Hüllen von Strömen; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig; der Leser hat keinen Zustand zwischen Aufrufen. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
