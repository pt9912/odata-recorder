# Slice slice-odata-codec-csdl: OData-Codec, Teil Metadaten (CSDL)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-06`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-06.a` (Metadatenmodell) · `SPEC-002` (OData-Modell) · `SPEC-049` (EDM-Typen) · `SPEC-068` (`ODR-E6004`) · `SPEC-107` (Service-Metadaten) · `LH-RB-01.a` · `ARC-004` (`ODataCodec`) · `ARC-011`.
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

**Ziel:** Der Outbound Port `ODataCodec` (Teil Metadaten) und sein Driven Adapter `src/adapters/driven/odata/csdl` lesen EDMX/CSDL V4 in ein technologieunabhängiges Modell aus Schemas, Entity-Typen, Schlüsseln, Properties, Entity-Sets und Navigation-Properties, lösen Referenzen und Vererbung auf und weisen nicht Unterstütztes im Validierungsbericht aus.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Entities und Collections im JSON — `slice-odata-codec-json`; der Teil braucht dieses Modell.
- OData-URL-Parsing und Query-AST — gehört zum Mock; der Port trägt hier nur die Operationen für Metadaten.
- Ausliefern der Metadaten über HTTP — Mock-Modus, nicht in der Welle.

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

- [ ] Port `ODataCodec` (Metadaten) und das Domain-Modell des OData-Dienstes; kein Parsertyp in Signaturen (`architecture.md` §2).
- [ ] CSDL-Adapter: Parsen, Referenzen, Vererbung, EDM-Typen der Abnahmematrix; `ODR-E6004` bei nicht parsebarem CSDL, EDMX-Version ungleich 4.0, Entity-Typ ohne Schlüssel, nicht auflösbarer Referenz und Schlüssel-Property mit nicht unterstütztem Typ.
- [ ] Validierungsbericht mit `unsupported_features` für Properties und Annotationen, die das Modell gültig lassen.
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
| `src/hexagon/model/` | update | Modell des OData-Dienstes |
| `src/hexagon/ports/outbound/` | update | `ODataCodec` |
| `src/adapters/driven/odata/csdl/` | neu | CSDL-Leser |
| `spec/spezifikation.md` | update | Vertrag von `ODataCodec` (Metadaten) mit Randformen, Form des Validierungsberichts |
| Tests | neu | Fixtures echter CSDL-Dokumente und Negativfälle je Fehlerbedingung; Vertragstest am Port |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-fehlermodell-konfiguration` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 1700 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 1200 Zeilen in derselben Metrik.
- `in-progress` → `open` (blockiert): eine Bibliothek für CSDL wird nötig und hat keine ADR (siehe §6).

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

- Bibliothek: Die Standardbibliothek (`encoding/xml`) reicht für EDMX, eine Drittbibliothek braucht eine ADR vor dem Code (`AGENTS.md` §3.6 gilt sinngemäß, wie bei [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)) — **Ausgang:** offen; Architect entscheidet vor dem Code.
- Wer die Metadaten eines Service ablegt (`MetadataRepository.SetMetadata`) ist nicht festgelegt: Die Spezifikation nennt das Erfassen von `$metadata` im Modus `record` und die Prüfung beim Import, aber keinen Aufrufer und keinen Zeitpunkt — **Ausgang:** offen; Architect entscheidet vor dem Code und weist den Aufrufer einem Slice der Welle zu (Kandidaten: `slice-record-service`, `slice-baseline-import`).
- Unbekannte Namensräume, mehrere Schemas, Aliase, Zyklen in der Vererbung, abstrakte Typen, offene Typen, Composite-Schlüssel, Schlüssel aus Navigation — **Ausgang:** offen; Raster *Konstrukt × Behandlung* in §6.
- Größe und Tiefe: ein Dokument mit sehr vielen Typen oder tiefer Vererbung — **Ausgang:** offen; Grenzen für Eingabegröße und Tiefe benennt die Spezifikation oder die Zelle sagt „gilt nicht“ mit Grund.
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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (2×); der Codec hat viele Randformen aus einem fremden Format; das Raster entsteht vor dem Code.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
