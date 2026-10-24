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

**Berührte Spec-Stellen:** `LH-FA-18.a` · `LH-FA-11.a` · `SPEC-043` · `SPEC-044` · `SPEC-047` (Use-Case-Teil) · `SPEC-107` · `SPEC-004` (Manifest) · `SPEC-051`, `SPEC-068`, `SPEC-069`, `SPEC-097` · `ARC-002`, `ARC-003` · `architecture.md` §4.2, §4.7.
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

- [ ] Lesen, Interpretieren und Zusammenführen: nur `complete`-Exchanges mit Status 200 und passender Metadatenversion, Deduplizierung identischer Payloads, Teilprojektionen nur widerspruchsfrei, Konflikte stoppen mit `ODR-E6101`, redigierte Nicht-Key-Properties als `unknown` maskiert, redigierte Key-Felder `ODR-E6101`; Provenienz je Entity.
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
| Tests | neu | Fixtures aus aufgezeichneten Antworten des Referenzdienstes; Fälle der Abnahme von `LH-FA-18` und `LH-FA-11` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

- Die Kettenauflösung bildet den Match-Key mit der Funktion aus `slice-redaction-match-key`; der Import nutzt damit dieselbe Funktion wie Record und Replay.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-odata-codec-csdl` und `slice-odata-codec-json` und `slice-redaction-match-key` und `slice-session-wartung` und `slice-session-use-case` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 2100 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 1500 Zeilen in derselben Metrik (vorläufig, vor Vertrag und Tabelle; der Planner setzt die Marke beim Start nach der Tabelle des Architect neu, `AGENTS.md` §3.14, und ersetzt diese mit Grund).
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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, Ausgang `verkörpert` in `AGENTS.md` §3.10); die Randformen von `LH-FA-18.a` sind zahlreich; das Raster steht vor dem Code. Der Slice ist nebenläufig: das Gate fährt `-race` (`make test-race`, `SPEC-122`), die nebenläufigen Fänger ordnen weiter mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8b), und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
