# Slice slice-baseline-import: ImportUseCase mit Coverage-Beweis

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-18`, `LH-FA-11`, `LH-FA-14`, `LH-FA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-18.a` · `LH-FA-11.a` · `SPEC-043` · `SPEC-044` · `SPEC-047` (Use-Case-Teil) · `SPEC-107` · `SPEC-126` (Aufrufer von `ODataCodec`) · `SPEC-004` (Manifest) · `SPEC-051`, `SPEC-068`, `SPEC-069`, `SPEC-097` · `SPEC-125` (Aufrufer von `MetadataCodec`, Modell Nr. 11) · `SPEC-116` (`BildeSchluessel`) · `SPEC-109` Nr. 5 (Abbildung der Port-Fehler) · `SPEC-114` Nr. 9 (`import_id`) · `LH-FA-14.a` (Maske `unknown`, Match-Key) · `LH-FA-03` (Session als Quelle) · `ARC-002`, `ARC-003` · `architecture.md` §2 (Zeile `ImportUseCase`), §4.2, §4.7 (Sequenz Baseline-Import).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `ImportUseCase` liest die abgeschlossenen Exchanges einer Session, interpretiert sie über den Codec gegen das Modell, führt Entities zusammen oder weist Konflikte ab, belegt Coverage nur durch den Vollständigkeitsbeweis und veröffentlicht Manifest, Import-Eintrag, Quellen und Coverage atomar als unveränderliche Baseline; ein wiederholter Import mit gleicher `import_id` ist idempotent.

**Übernimmt:** Die Übergaben aus `welle-persistenz` zur Coverage: die Durchsetzung von `LH-FA-11` („`complete` nur mit Beweis“, `SPEC-107` *Ungeprüft*) und die Form der Coverage-Evidenz (`evidence_json` ist dort frei geformt).

**Übernimmt (Zeit und Kennungen):** Die Snapshot-Kennung der Baseline kommt aus `IDGenerator.SnapshotID()`, `created_at` aus `Clock.Now()` (`SPEC-114`); die `import_id` vergibt der Client und der Generator berührt sie nicht (`SPEC-114` Nr. 9; Sendung von `slice-clock-idgenerator`).

**Übernimmt (Match-Key):** Die Auflösung von `@odata.nextLink` bildet den Key mit `BildeSchluessel` (`SPEC-116`); bei `ErrKodierung` gilt der Link als nicht auflösbar (die Kette bricht ab, die Coverage ist nicht `complete`), und `key_redacted` verhindert `complete` (Sendung von `slice-redaction-match-key`).
**Übernimmt (Metadaten):** Der Import liest das `$metadata` der Session mit `MetadataCodec` (`SPEC-125`) in das Modell, gegen das er die Entities prüft, und ein `ODR-E6004` endet ihn, bevor etwas geschrieben ist; nach jedem erfolgreichen `Publish`, auch dem idempotenten ohne neuen Import, ruft er `MetadataRepository.SetMetadata(service_id, ref, ref)` mit der Blob-Referenz des Dokuments als Wert und Hash, und ein Fehler von `SetMetadata` ist der Fehler des Imports bei bestehender Baseline (`SPEC-107` Nr. 15; Sendung von `slice-odata-codec-csdl`, `AGENTS.md` §3.13). Offen für den Vertrag dieses Slice, ausdrücklich nicht dort entschieden: welche Antwort der Session das `$metadata` ist (mehrere, keine, nicht `complete`, nicht 200) und was bei einem `metadata_hash` der Anfrage ungleich der Referenz gilt. Der Fake des Ports `MetadataCodec` (Zählung jeder Operation, jedes Argument festgehalten, Modell, Bericht und Fehler als Vorgabe) liefert dieser Slice mit seinen Tests; `slice-odata-codec-csdl` liefert die Vertragssuite des Ports, aber keinen Fake.

**Übernimmt (Entities):** Der Import liest die Antworten der Exchanges mit `ODataCodec` (`SPEC-126`) in typisierte Entities (Sendung von `slice-odata-codec-json`, `AGENTS.md` §3.13). Entschieden dort, hier umzusetzen: (a) **Aufruf:** `ParseCollection` für die Antwort auf `GET <Set>`, `ParseEntity` für die auf `GET <Set>(<Schlüssel>)`, nur für `complete`-Exchanges mit Status 200, Body `present` und JSON; das Modell ist das des `$metadata` derselben Session, `ctx` der des Imports, `grenze` `MAX_CAPTURE_BODY` (`SPEC-013`), `src` der Strom des Blobs; die Auswahl der Antworten vor dem Aufruf trifft der Import (eine Antwort mit `odata.metadata=none`, auf `$count`, einen Property-Pfad oder eine Navigation ruft er nicht, denn sie hat keinen Kontext der Form `SPEC-126` Nr. 7 und der Fehler `kontext` beendete den Import). (b) **Fehlerabbildung:** ein `*model.DatenFehler` wird zu `ODR-E6101` (`SPEC-109` Nr. 5, Gegenstand Import), ein Kontextfehler bleibt ohne Code, ein Fehler von `src` wird nach `SPEC-109` Nr. 5 abgebildet; `ODR-E6004` bleibt dem Lesen des `$metadata`. (c) **Fake:** der Fake des Ports `ODataCodec` (Zählung jeder Operation, jedes Argument festgehalten, Ergebnis und Fehler als Vorgabe) liefert dieser Slice mit seinen Tests, 200 Zeilen in seiner Marke (`AGENTS.md` §3.14, der Planner setzt sie neu); `slice-odata-codec-json` liefert die Vertragssuite des Ports, aber keinen Fake. (d) **Aufrufstellen als Menge:** jede Stelle des Imports, die `ParseEntity` oder `ParseCollection` ruft, bekommt dort ihre Fälle für `ctx`, Fehlerabbildung, Kontextfehler und „kein Aufruf, den der Dienst nie rufen darf“ (`AGENTS.md` §3.15). (e) **`@odata.nextLink`:** der Import löst auf (`SPEC-116`, `BildeSchluessel`); `FolgeLink.Redigiert` heißt, die Kette ist nicht belegbar (die Coverage bleibt `partial`); ein fehlender Link auf der letzten Seite ist ein Teil des Beweises, nicht der Beweis. (f) **Maske und Report:** `Wert.Art` gleich `redigiert` wird `unknown` (`LH-FA-14.a`); ein redigierter Schlüssel kommt als `ODR-E6101` vom Codec; `Entity.Uebersprungen` fließt in den Report (`unsupported_features`); `Anzahl` ist Evidenz der Coverage, kein Beweis; zwei Elemente mit demselben `Schluessel` liefert der Codec beide, Zusammenführen und Konflikte entscheidet dieser Slice (`LH-FA-18.a` Nr. 4). Das *Ausdrücklich NICHT* unten trifft diese Sendung nicht.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Endpunkt `POST /__admin/imports` samt Authentisierung — Folgewelle (Welle §6); `SPEC-047` bindet den Use Case, der Adapter kommt später.
- Lesen der Baseline im Mock (Coverage-Prüfung vor Collection-Reads, `STRICT_COVERAGE`) — Mock-Modus, nicht in der Welle; hier entsteht die Evidenz, die er liest.
- Import aus mehreren Sessions mit Konfliktstrategie außer `reject` — `SPEC-047` lässt im MVP nur `conflict_policy=reject` und `target=new_baseline` zu.

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

- [ ] Lesen, Interpretieren und Zusammenführen: nur `complete`-Exchanges mit Status 200 und passender Metadatenversion, Entities über `ODataCodec` lesen (Aufruf je Antwortart, Fehlerabbildung, Fake mit Zählung; Sendung von `slice-odata-codec-json`, §1), Deduplizierung identischer Payloads, Teilprojektionen nur widerspruchsfrei, Konflikte stoppen mit `ODR-E6101`, redigierte Nicht-Key-Properties als `unknown` maskiert, redigierte Key-Felder `ODR-E6101`; Provenienz je Entity.
- [ ] Coverage: `complete` nur nach Vollständigkeitsbeweis (ungefilterte Collection-Abfrage, geschlossene `@odata.nextLink`-Kette mit Auflösung über den Match-Key, kein `key_redacted`), sonst `partial` oder `unknown`; Form der Evidenz in der Spezifikation; ein leerer Import ist kein Beweis einer leeren Collection.
- [ ] Manifest (`manifest_version=1`) als Blob, atomares Veröffentlichen über `ImportRepository.Publish`, Idempotenz über den Hash der Anfrage, Fehlercodes `ODR-E7003`, `ODR-E7009`, `ODR-E6004`, `ODR-E6101`, `ODR-E6102`; das `$metadata` der Session mit `MetadataCodec` vor dem ersten Schreiben lesen und nach jedem erfolgreichen `Publish` `SetMetadata` rufen (Sendung von `slice-odata-codec-csdl`, §1).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/inbound/` | update | `ImportUseCase` |
| `src/hexagon/services/baselineimport/` | neu | Interpretation, Zusammenführen, Coverage, Veröffentlichen |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, Form der Evidenz, Form des Manifests und der Reports |
| `spec/architecture.md` | update, falls der Vertrag sie ändert | Zeile `ImportUseCase` in §2, Sequenz Baseline-Import in §4 (Berichtigung an jeder Stelle, `AGENTS.md` §3.9; ohne Slices und ADR-Bezüge, `AGENTS.md` §3.4) |
| `src/hexagon/model/` | update, falls der Vertrag Typen verlangt | Import-Report, Coverage-Evidenz, Manifest-Typen (Ort und Namen entscheidet der Architect mit dem Vertrag) |
| `src/hexagon/ports/outbound/` (Fakes `ODataCodec` und `MetadataCodec`) | neu | Ort der Fakes entscheidet der Architect (Muster der Fakes anderer Ports) |
| Tests | neu | Fixtures aus aufgezeichneten Antworten des Referenzdienstes, aus Bytewerten gebaut, wo ein Zeichen oder eine Kodierung die Eingabe ist; Fälle der Abnahme von `LH-FA-18` und `LH-FA-11`; die Fakes von `MetadataCodec` und `ODataCodec` (Zählung jeder Operation, jedes Argument festgehalten) |

Der Vertrag, die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen in der Spezifikation (Produkt, kein Werkzeug) vor dem ersten Code-Commit (`AGENTS.md` §3.12), nicht vor dem Start; die Randformen, die sie tragen müssen, nennt §6.

- Die Kettenauflösung bildet den Match-Key mit der Funktion aus `slice-redaction-match-key`; der Import nutzt damit dieselbe Funktion wie Record und Replay.

**Reihenfolge der Rollen nach dem Start**, jede erst, wenn die vorige übergeben hat, jede in frischem Kontext: Architect (Vertrag, Entscheidungen aus den Aufträgen unten, Tabelle, Raster, Paare, Grenzen, Nachzüge, Sendungen an Nehmer, erwartete Produktcode-Größe) → Planner (Marke aus der Tabelle und der Produktcode-Erwartung, §4, eigener Commit, vor dem ersten Code-Commit; Prüfung der Adressen nach `AGENTS.md` §3.13; Schnittfrage, wenn die Schätzung × 1,25 über 3000 liegt) → Implementer (erster Code-Commit erst nach der Marke) → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure; Änderungen an `AGENTS.md` nur nach Diff-Vorlage an den User). Der letzte Commit des Architect, der Vertrag, Tabelle und Raster abschließt, ist die `<Basis>` der Marke (§4). Hält der Implementer an einer Zelle ohne Ort an (`AGENTS.md` §3.12), geht sie an den Architect, nicht in den Code.

**Betriebsregeln für alle Rollen** (Lehren der Läufe vom 2026-10-24: OOM durch Mutationen ohne Grenze, Zeitgrenzen der Gegenproben unter Last):

- Jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4`; höchstens zwei Container gleichzeitig; Mutationen nacheinander, nie parallel. Eine Mutation, die ein Limit entfernt, läuft mit `--memory=2g` und `timeout`. Vor einem großen Lauf `free -g`. Docker sonst nur über `make` (`AGENTS.md` §3.1).
- `make gates` nie parallel zu Mutationsläufen: die Zeitgrenzen der Gegenproben und von `pruefung` sind für einen ruhigen Host bemessen, ein Rot unter Last ist kein Befund, bevor es auf ruhigem Host wiederholt ist (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
- Auf Repo-Dateien kein `sed -i`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (Scratch außerhalb des Repos).
- Das Werkzeug setzt eine Folge aus Backslash, `u` und vier Hex-Ziffern still um (`AGENTS.md` §3.18): Fixtures, Pläne und Berichte mit solchen Folgen, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen; der Bericht nennt die gegengelesene Datei.
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (`AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen. Eine Hilfe, die ein Argument umhüllt (Strom, Fake, `ctx`), darf keine Eigenschaft des Originals verdecken: sie trägt jede Schnittstelle des Originals, zählt je Aufruf, und ein Fänger prüft die Hülle selbst (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13), nicht nur im Bericht. Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen.

**Prüfung der Sendungen an diesen Slice durch den Planner** (`AGENTS.md` §3.13, beim Start gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 dieses Plans; Nehmer ist dieser Slice, er liegt jetzt in `in-progress/`): (1) `slice-odata-codec-json` (`ParseEntity`, `ParseCollection`, `DatenFehler`, `FolgeLink.Redigiert`, `Entity.Uebersprungen`, `Anzahl`, `Wert.Art`, Fake von `ODataCodec`): angenommen; §1 *Übernimmt (Entities)* (a) bis (f) und DoD Punkt 1 tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie (Endpunkt, Lesen im Mock, Konfliktstrategie außer `reject`). (2) `slice-odata-codec-csdl` (`MetadataCodec`, `SetMetadata` nach `Publish`, Fake von `MetadataCodec`, offene Wahl der Antwort und Abweichung des Hashes): angenommen; §1 *Übernimmt (Metadaten)* und DoD Punkt 3. (3) `slice-redaction-match-key` (`BildeSchluessel`, `ErrKodierung`, `key_redacted`) und `slice-clock-idgenerator` (Snapshot-Kennung, `created_at`, `import_id` des Clients): angenommen in §1. (4) `slice-session-wartung`, `slice-session-use-case`: keine Sendung, nur Start-Trigger; alle fünf Geber des Start-Triggers liegen in `done/`. (5) Teil Z von `slice-odata-codec-json` blieb im Slice (kein Folge-Slice nötig, der Start-Trigger bleibt wie er ist). (6) Der Zweifel aus der Fachentscheidung zu `slice-odata-codec-json` (ob der Import eine Projektion ohne Schlüssel überspringen *darf*; der Codec lehnt sie mit `schluessel` ab) ist ein Auftrag an den Architect (Auftrag 3 unten), weil das Lastenheft ihn nicht beantwortet. Offen im Sinne von §3.13 ist nichts, was dieser Slice an andere sendet; was er selbst an Folge-Slices weist, nennt der Architect im Vertrag (Auftrag 9).

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; Pfade `spec/spezifikation.md`, `spec/architecture.md`; Kennungen fortlaufend je Datei; Muster sind `SPEC-125` und `SPEC-126`; eine ADR nur, wenn eine Entscheidung sie verlangt, sonst Spezifikation, `AGENTS.md` §3.8):

1. **Vertrag des Imports** (Port `ImportUseCase` in `inbound/`, Typen, Signatur, Fehlerform) und **alle Randformen vor dem Code** (`AGENTS.md` §3.12): Eingabe (`import_id`, `service_id`, `source_session_ids`, `metadata_hash`, `conflict_policy`, `target` nach `SPEC-047`; `ctx` nil und beendet; leere, doppelte, unbekannte Werte), Rückgabe (Report, `baseline_snapshot_id`, `coverage[]`, `entity_count`; neu gegen Wiederholung), Fehler je Verstoß mit Code (`ODR-E7003`, `ODR-E7009`, `ODR-E6004`, `ODR-E6101`, `ODR-E6102`, Speicherfehler nach `SPEC-109` Nr. 5) und Rangfolge, wenn mehrere zugleich gelten. Form des Manifests (`manifest_version=1`, `SPEC-043`, `SPEC-004`), der Evidenz (`evidence_json`, Übergabe aus `welle-persistenz`) und des Reports (`unsupported_features`, `SPEC-051`): vor dem Code in der Spezifikation, ohne Rest-Offenheit (Risiko in §6).
2. **Welche Antwort der Session ist das `$metadata`** (offen aus `slice-odata-codec-csdl`, `SPEC-107` Nr. 15): Kriterium der Zuordnung in der Session (Pfad, Methode, Status, `capture_status`, Body-Zustand; Fall keine, mehrere, mehrere gleiche und mehrere abweichende Antworten, nicht `complete`, nicht 200, nicht lesbar, Redaction-Zustand des Bodys), und was bei einem `metadata_hash` der Anfrage ungleich der Blob-Referenz des gewählten Dokuments gilt (Code, Rang, ob der Hash die Auswahl bestimmt oder nur prüft; Zusammenspiel mit der Idempotenz über den Hash der Anfrage). Das gewählte Dokument ist auch das Modell, gegen das alle Entities der Session gelesen werden.
3. **Auswahl der Antworten vor dem Aufruf des Codecs** (Sendung (a)): welche Exchanges zählen als Collection-, Entity-, `$metadata`- oder Fremdantwort (`$count`, Property-Pfad, Navigation, `$ref`, `odata.metadata=none`, Query-Optionen `$select`, `$filter`, `$top`, `$skip`, `$expand`, `$orderby`), Status ungleich 200, `capture_status` nicht `complete`, Body `absent`, `truncated` oder redigiert, nicht JSON; je Art, ob der Import sie überspringt (und wie der Report es ausweist), ob sie die Coverage verhindern oder ob sie den Import mit einem Code beenden. Dazu der Zweifel aus `slice-odata-codec-json`: eine Projektion ohne Schlüssel ruft der Import nicht, oder der Import endet mit `ODR-E6101`; entscheiden, mit der Folge im Betrieb.
4. **Zusammenführen, Deduplizieren, Konflikte** (`LH-FA-18.a` Nr. 4): Identität (`entity_set` und Schlüssel), Gleichheit zweier Payloads über die kanonischen Werte (`SPEC-126` Nr. 10), Teilprojektionen nur widerspruchsfrei (Vereinigung der Properties; `null` gegen nicht selektiert gegen `unknown`), wann zwei Werte widersprechen (inkl. `Art`-Wechsel, anderer Typ, `@odata.type`), Reihenfolge der Entscheidung und Determinismus über die Reihenfolge der Exchanges, Fehlercode und Ort bei Konflikt (`ODR-E6101`, `SPEC-109`/`SPEC-051` Ursache), zwei Elemente mit gleichem Schlüssel in einer Antwort (`SPEC-126` Nr. 12), Provenienz je Entity (`SPEC-043`, `provenance[]`).
5. **Coverage und Beweis** (`LH-FA-11`, `SPEC-044`, `SPEC-107` *Ungeprüft*): wann `complete`, `partial`, `unknown`; Form der Evidenz; ungefilterte Collection-Abfrage (welche Query-Optionen sie ausschließen), **Auflösung von `@odata.nextLink`** (`SPEC-116`: Eingabe von `BildeSchluessel` aus Link und Request der Seite, absolut gegen relativ, andere Authority oder Basis, Schleife, doppelte Seite, Link auf eine Seite außerhalb der Session, `ErrKodierung`, `key_redacted`, Reihenfolge der Kette gegen Reihenfolge der Aufzeichnung), `FolgeLink.Redigiert` und fehlender Link der letzten Seite, `Anzahl` als Evidenz (Vergleich mit den Elementen: gleich, kleiner, größer, fehlt) und nicht als Beweis, ein leerer Import ist kein Beweis einer leeren Collection, mehrere Beweise für dieselbe Collection (zwei Ketten, zwei Sessions), Teilketten.
6. **Maske `unknown` und Redaction-Zustände** (`LH-FA-14.a`): Zustände von `Wert.Art` (`redigiert`, `null`, nicht selektiert) und von Body und Header der Exchange (`none`, `applied`, `truncated`, `key_redacted`, fail-closed-Reste) je Art der Property und des Schlüssels; was `unknown` im Payload heißt (Form in `payload`/`selected_properties`), welche Zustände `complete` verhindern, `Entity.Uebersprungen` im Report (`unsupported_features`), redigierter Schlüssel (`ODR-E6101` vom Codec) gegen redigierter `@odata.nextLink`.
7. **Abbildung der Fehler und `ctx` je Aufrufstelle** (`AGENTS.md` §3.15, §3.17): die Aufrufstellen des Imports als Menge — jede Stelle, die `MetadataCodec.ParseMetadata`, `ODataCodec.ParseEntity`, `ODataCodec.ParseCollection`, `BildeSchluessel`, Repository-, Blob-, Clock- oder Generator-Operationen ruft —, und je Stelle: `ctx` unverändert weitergereicht, Kontextfehler ohne Code und unverändert (`==`), Fehler des Ports nach `SPEC-109` Nr. 5, Fehler des Stroms `src` (Blob fehlt, beschädigt), Größengrenze (`grenze` ist `MAX_CAPTURE_BODY`, `SPEC-013`; für `ParseMetadata` entscheiden), und was der Dienst nie rufen darf (Zählung jeder Operation der Fakes, auch derer, die er nie rufen darf: kein `Publish` vor dem letzten Lesen, kein `SetMetadata` nach einem Fehler vor `Publish`, kein Codec-Aufruf für eine übersprungene Antwort). Fehler von `SetMetadata` nach `Publish` (Nr. 15 von `SPEC-107`: Fehler des Imports bei bestehender Baseline) mit Wiederholung als Heilungsweg.
8. **Idempotenz, Nebenläufigkeit und Randformen des Veröffentlichens:** Hash der Anfrage (kanonische Form, Reihenfolge von `source_session_ids`), gleiche `import_id` und gleicher Hash (Wiederholung, `SetMetadata` trotzdem), gleiche `import_id` und anderer Inhalt (`ODR-E6102`), fehlgeschlagener Import und Wiederverwendung der `import_id`, gleichzeitige Imports derselben `import_id` und derselben Session (`SPEC-047`: der Primärschlüssel entscheidet), Session wird während des Imports gelöscht oder beendet, Session im Status `open` (`ODR-E7009`), unbekannte Session (`ODR-E7003`), Session ohne Exchanges, Session mit anderem Service als `service_id`, mehrere Sessions in `source_session_ids` (der Wortlaut von `SPEC-047` lässt `conflict_policy=reject`, §1 *Ausdrücklich NICHT* schließt andere Strategien aus, nicht mehrere Sessions: entscheiden). **Nebenläufigkeit:** ob der Import selbst Goroutinen startet (etwa Lesen der Blobs oder Codec-Aufrufe), oder sequenziell liest; die Entscheidung bestimmt den Faktor 85 gegen 40 der Marke und ob Zusagen zu gleichzeitigen Imports einen Wirkungs-Fänger unter `make test-race` brauchen (Rate je Prozess und Zahl, `AGENTS.md` §3.10; ein Rennen ohne ungeordneten Zugriff meldet der Detektor nicht).
9. **Raster als §6-Zellen** *Operation × Argumentart* (`AGENTS.md` §3.16) mit Zusatztabelle der Paare von Verstößen (Dimension *Art der Antwort* und *Position* in der Kette/Session; je Paar der Fänger, die Zahl der Paare aus der Menge gezählt) und die **Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15) mit der Menge je Zeile aus dem Wortlaut der Zusage (`AGENTS.md` §3.17), die Mutation je Zeile und der Fall, in dem zwei verglichene Werte gleich sind; je Klasse, ob sie generierbar ist (Probe-Paket des Verifiers, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`). Jedes akzeptierte Negativ mit Grund und Folge. **Sendungen** nach `AGENTS.md` §3.13: wer etwas an einen Folge-Slice (`slice-workspace-use-case`: Lesen der Baseline; `slice-replay-service`; `slice-bootstrap-m2`: Composition Root, Verdrahtung von `ImportUseCase`; `slice-http-adapter` und Folgewelle: Endpunkt) zuweist, liest vorher dessen §1 *Ausdrücklich NICHT* und DoD und trägt die Sendung im selben Commit beim Nehmer ein; ein Nehmer in `done/` oder mit passendem Ausschluss ist keine Adresse, der Punkt geht an den Planner.
10. **Nachzug an jeder Stelle, die dieselbe Aussage trägt** (`AGENTS.md` §3.9): `LH-FA-18.a`, `LH-FA-11.a`, `SPEC-043`, `SPEC-044`, `SPEC-047`, `SPEC-107` (Nr. 15, *Ungeprüft*), `SPEC-125` Kopf (Auswahl des `$metadata`), `SPEC-126` (Verweise auf den Vertrag des Imports), `SPEC-109`/`SPEC-051` (Ursache des `ODR-E6101` im Import, falls ein Feld hinzukommt), die Zeile `ImportUseCase` in `architecture.md` §2 und die Sequenz Baseline-Import (ohne Slices, Wellen, ADR; `AGENTS.md` §3.4). Ein Nachzug, der einen Satz berichtigt, sucht alle Stellen, die ihn tragen.
11. **Größe:** der Architect nennt zum Vertrag die **erwartete Produktcode-Größe des Imports** in Zeilen (`baselineimport`, Typen, Evidenz, Manifest; mit Begründung, Spanne und Vergleich mit dem Produktcode von `slice-session-beenden` oder `slice-record-service` als Dienste); der **Leser-Zuschlag aus `AGENTS.md` §3.14 gilt nicht** (jede andere Zeile × 40 statt × 65), denn der Import liest keine fremden Bytes selbst, er ruft die Codecs. Fakes: +200 für `ODataCodec`, +200 für `MetadataCodec`. Der Architect benennt, wenn es einen gibt, einen **Teil ohne geteilten Mechanismus**, der beim Handoff geschnitten werden dürfte (etwa Coverage und Beweis gegen Interpretieren und Zusammenführen, wie das Risiko in §6 ihn vorbereitet: nur wenn Evidenz und Ketten keinen Zustand mit dem Zusammenführen teilen), damit §4 ihn vor dem Code benennen kann (`AGENTS.md` §3.14). Er zählt die Zeilen der Tabelle nach der Zählregel von `slice-odata-codec-csdl` §4 (nebenläufig, wenn der Code eine Goroutine startet oder der Fänger eine Goroutine, eine Barriere oder einen Halt trägt).

**Der Planner hat nichts davon entschieden.** Die Aufträge benennen die Entscheidungen und Mengen, die der Architect vor dem Code schließen muss; jede offene Entscheidung steht in §6 mit dem Ausgang *offen*.

**Aufträge an den Implementer** (nach der Marke): (a) der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); (b) der Absatz in §3 mit *Zusage · Mutation · roter Test* und den Raten (Zahl der Prozesse und Messender, `AGENTS.md` §3.10); (c) jeder Fix bildet die Menge aus dem Wortlaut der Zusage, schreibt den Fänger als Schleife und nennt je Befund Menge, Zahl der Elemente und eine Mutation an einem nicht genannten Element (`AGENTS.md` §3.17); (d) Messung `git diff --numstat <Basis>..HEAD` über die Pfade der Marke vor der Übergabe, die Zahl im Bericht; (e) die Fakes von `ODataCodec` und `MetadataCodec` zählen jede Operation (auch die, die der Dienst nie rufen darf) und halten jedes Argument fest; (f) die Betriebsregeln oben.

**Aufträge an Reviewer und Verifier** (jeder in frischem Kontext): jede Zeile der Tabelle und das Raster mit je einer Mutation auflösen, Fixes gegen die **Klasse** (`AGENTS.md` §3.17), nicht gegen die genannte Mutation; der Verifier fährt zusätzlich ein **Probe-Paket aus Generatoren** über die Klassen des Vertrags (Antwortarten × Zustände, Ketten, Paare von Antworten mit gleichem Schlüssel), einmal gegen den Code, dann gegen die Mutationen, und legt es im Report ab (Pfad und Ergebnis). Die Fakes werden auf die Hülle-Falle geprüft (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`). Betriebsregeln oben, besonders für Mutationen, die ein Limit entfernen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-odata-codec-csdl` und `slice-odata-codec-json` und `slice-redaction-match-key` und `slice-session-wartung` und `slice-session-use-case` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand des Starts (2026-10-25):** Der Start-Trigger ist erfüllt: `slice-odata-codec-csdl`, `slice-odata-codec-json`, `slice-redaction-match-key`, `slice-session-wartung` und `slice-session-use-case` liegen in `done/`. Der Vertrag, die Tabelle und das Raster stehen noch aus (Architect, §3); die Marke ist **noch nicht neu gesetzt**, die vorläufige 2100 unten bleibt bis dahin unangetastet.

**Größenmarke** (`AGENTS.md` §3.14): **noch nicht gesetzt.** Der Planner setzt sie nach dem Vertrag des Architect und vor dem ersten Code-Commit auf die ganze Schätzung × 1,25, aufgerundet auf volle Hundert, und ersetzt die vorläufige Marke 2100 dabei mit Grund. Die Schätzung nach der Regel: Zeilen der Tabelle *Zusage · Fänger* (× 85 nebenläufig, × 40 jede andere; der Leser-Zuschlag × 65 gilt nicht, der Import liest keine fremden Bytes selbst) plus 200 für den Fake von `ODataCodec` plus 200 für den Fake von `MetadataCodec` plus die vom Architect genannte erwartete Produktcode-Größe als fester Anteil (§3 Auftrag 11); beide Zahlen stehen getrennt in §4. Liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten; lässt sich kein Teil ohne geteilten Mechanismus abtrennen, steht der Grund hier. Ein beim Handoff erlaubter Schnitt ist nur der Teil, den der Architect benennt und §4 vor dem Code festhält. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der letzte Commit des Architect, der Vertrag, Tabelle und Raster abschließt (Dokumente zählen nicht mit). Messpunkte der Vorgänger, die der Planner vor Augen hat (Hinweis, keine Regel): `slice-odata-codec-csdl` Formel 1850, Ende 4333; `slice-odata-codec-json` Formel 2050, Ende 7236 (Dienste wie `slice-session-beenden`: Marke 3000, gemessen 2998).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 2100 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 1500 Zeilen in derselben Metrik (vorläufig, vor Vertrag und Tabelle; der Planner setzt die Marke beim Start nach der Tabelle des Architect neu, `AGENTS.md` §3.14, und ersetzt diese mit Grund). Hinweis für die Neusetzung (Sendung von `slice-odata-codec-json`): zur Formel kommen +200 für den Fake von `ODataCodec` und +200 für den Fake von `MetadataCodec` sowie der Produktcode des Imports als fester Anteil, den die Formel nicht zählt (Lehre aus `slice-odata-codec-csdl` und `slice-odata-codec-json`).
- `in-progress` → `open` (blockiert): die Form des Manifests oder der Evidenz bleibt ohne Entscheidung des Architect offen.

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

- Der Slice ist der größte der Welle; sein Schnitt in „Interpretieren und Zusammenführen“ und „Coverage“ ist vorbereitet — **Ausgang:** offen; wird die Marke überschritten, geht der Teil Coverage und Beweis als `slice-import-coverage` zurück zur Zerlegung.
- Form der Evidenz (`evidence_json`) und des Reports: nicht festgelegt, beide fließen später in den Mock und in die Admin-Antworten — **Ausgang:** offen; Architect legt sie vor dem Code in der Spezifikation fest.
- Gleichzeitige Imports derselben Session und derselben `import_id`; Session wird beim Import gelöscht — **Ausgang:** offen; Raster-Zelle, Wiederholungsläufe.
- Der Beweis für `complete` setzt `key_redacted=false` für jede Seite voraus; Replay-Requests mit zusammenfallendem Key sind gewollt (`LH-FA-14.a`) — **Ausgang:** offen; Fänger in der Tabelle *Zusage · Fänger*.
- Metadaten des Service (`metadata_hash`) zum Zeitpunkt des Imports: wer sie setzt und was der Hash ist, ist entschieden (`SPEC-107` Nr. 15: der Import, nach `Publish`, die Blob-Referenz); offen bleiben die Wahl der Antwort der Session und die Abweichung des Hashes der Anfrage (§1) — **Ausgang:** offen; Architect legt beides vor dem Code im Vertrag des Imports fest.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- Auswahl der Antworten vor dem Aufruf des Codecs (Antwortarten, Query-Optionen, Status, Body-Zustände) und die Projektion ohne Schlüssel (Zweifel aus `slice-odata-codec-json`) — **Ausgang:** offen; Adresse: Architect (§3 Auftrag 3), Entscheidung mit Folge im Betrieb vor dem Code.
- Zusammenführen, Gleichheit zweier Payloads, Konflikte, Teilprojektionen und `unknown` (Zustände von `Wert.Art`, Body und Header) — **Ausgang:** offen; Adresse: Architect (§3 Auftrag 4 und 6), Zellen im Raster und Paare in der Zusatztabelle.
- Auflösung von `@odata.nextLink` und Form des Beweises (Schleife, doppelte Seite, andere Basis, `ErrKodierung`, `key_redacted`, `Anzahl` gegen Elemente) — **Ausgang:** offen; Adresse: Architect (§3 Auftrag 5), Fänger als Schleife über Linkformen und Kettenpositionen.
- Aufrufstellen des Imports als Menge (`ctx`, Fehlerabbildung, Kontextfehler, Größengrenze, „kein Aufruf, den der Dienst nie rufen darf“) und die Fakes von `ODataCodec` und `MetadataCodec` — **Ausgang:** offen; Adresse: Architect (§3 Auftrag 7), Implementer (Fakes zählen jede Operation, Hülle-Falle).
- Nebenläufigkeit des Imports und gleichzeitige Imports derselben `import_id` (Faktor 85 gegen 40 der Marke, Wirkungs-Fänger gegen Detektor, `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen`) — **Ausgang:** offen; Adresse: Architect (§3 Auftrag 8).
- Größe (`BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung`, 15×, verkörpert in `AGENTS.md` §3.14): die vorläufige Marke 2100 (Schätzung rund 1500) kennt weder Tabelle noch Produktcode noch zwei Fakes — **Ausgang:** offen; Adresse: Planner setzt die Marke nach dem Vertrag (§4); Architect nennt die Produktcode-Erwartung und einen möglichen Handoff-Schnitt (§3 Auftrag 11).
- Mehrere Sessions in `source_session_ids` (der Wortlaut von `SPEC-047` nennt eine Liste, §1 *Ausdrücklich NICHT* schließt nur andere Strategien als `reject` aus) — **Ausgang:** offen; Adresse: Architect (§3 Auftrag 8); schließt er mehrere Sessions aus dem MVP aus, trägt er das als Ausschluss in §1 ein.
- Bericht des Implementers (`BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`, 3×, verkörpert in `AGENTS.md` §3.10): Tabelle *Zusage · Mutation · roter Test* und Raten in §3 — **Ausgang:** offen; fehlt der Absatz im Review, geht der Beleg an den User.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Register unter `../observations/BEO-REPO/` trägt beim Start siebzehn Einträge (Sub-Area `*`; Stand 2026-10-25 nach der Closure von `slice-odata-codec-json`); Zahlen aus den Dateien unter `evidence/`, Ausgänge aus den `state.md`. Die Randformen von `LH-FA-18.a` und `LH-FA-11` sind zahlreich; das Raster steht vor dem Code. Der Slice liefert einen Vertrag über Zusammenführen und Beweis und ist der erste Aufrufer zweier Codec-Ports:
- `randform-erst-im-review-entschieden` (19×, verkörpert in `AGENTS.md` §3.16) und `universalaussage-im-vertrag-ungeprueft` (8×, §3.16): Raster in §6 vor dem ersten Code-Commit, Allaussagen („kein Fall außer … ist `ODR-E6101`“, „jede Antwortart“) nennen ihren Fänger über alle Fälle (§3 Auftrag 9).
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (24×, §3.15) und `akzeptiertes-negativ-ohne-folge` (9×, §3.15): Tabelle mit Menge je Zeile, Aufrufstellen als Menge, Negative mit Grund und Folge (§3 Auftrag 7, 9).
- `fix-deckt-die-genannte-mutation-nicht-die-klasse` (4×, verkörpert in `AGENTS.md` §3.17): Mengen aus dem Wortlaut der Zusage, Fänger als Schleife; Auftrag an Implementer, Reviewer, Verifier (§3).
- `suite-ohne-generator-fuer-klassen-der-spezifikation` (2×, offen): Probe-Paket aus Generatoren; der Architect nennt die generierbaren Klassen (§3 Auftrag 9).
- `hilfe-ersetzt-argument-verdeckt-eigenschaften` (1×, offen): trifft zu, denn die Fakes umhüllen keine Argumente, aber Hüllen um Ströme der Blobs und `ctx` kommen in den Fängern vor (Betriebsregeln, §3).
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (15×, §3.14): Marke nach Tabelle, Fakes und Produktcode-Erwartung, vor dem ersten Code-Commit (§4).
- `berichtigung-erreicht-nicht-jede-stelle` (10×, §3.9): Stellen in §3 Auftrag 10; Sendungen an Folge-Slices in Auftrag 9.
- `befund-an-rolle-ohne-artefakt` (3×, verkörpert in `AGENTS.md` §3.13): Betriebsregel in §3, jede offene Entscheidung steht in §6 mit Ausgang *offen* und Adresse.
- `nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, §3.10, Sensor `make test-race`), `rate-ohne-detektor-schwankt-mit-der-last` (2×, offen), `zweig-ohne-lauf-unter-dem-detektor` (1×, offen) und `panik-eines-fakes-verdeckt-den-benannten-test` (2×, offen): Bezug über die Nebenläufigkeit des Imports und gleichzeitige Imports (§3 Auftrag 8); die Fakes `ODataCodec` und `MetadataCodec` laufen im Aufrufer, eine Panik in einer Goroutine verdeckt den benannten Test, falls der Import nebenläufig liest.
- `bericht-der-rolle-liegt-nicht-im-repo` (3×, §3.10): Risiko in §6.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (3×, verkörpert in `AGENTS.md` §3.18): Fixtures mit Escape-Folgen, BOM und Kodierungen aus Bytewerten, gegengelesen (Betriebsregeln, §3).
- `gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, offen): `make gates` nie parallel zu Mutationsläufen (Betriebsregeln, §3).

**Stand des Starts (2026-10-25):** Unter der Schwelle (kein Ausgang zu vergeben): `panik-eines-fakes-verdeckt-den-benannten-test`, `rate-ohne-detektor-schwankt-mit-der-last`, `suite-ohne-generator-fuer-klassen-der-spezifikation` (je 2×), `hilfe-ersetzt-argument-verdeckt-eigenschaften`, `gegenprobe-haengt-an-der-laufzeit-des-produkts`, `zweig-ohne-lauf-unter-dem-detektor` (je 1×); ein dritter Beleg löst beim Lese-Schritt der Closure die Verkörperung aus. Alle übrigen sind verkörpert. Der Slice ist nebenläufig, soweit der Architect es entscheidet (§3 Auftrag 8): das Gate fährt `-race` (`make test-race`, `SPEC-122`), nebenläufige Fänger ordnen mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8b), und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
