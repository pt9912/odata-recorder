# Slice slice-expand-validator: Validator von `$expand` gegen die Navigationen des Modells

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

**Berührte Spec-Stellen:** `LH-FA-16.a` (eigener Validator-Contract-Test, unbekannte Operationen `501`) · `SPEC-002`, `SPEC-125` Nr. 9 d und Nr. 11 (Navigation-Properties des Modells, wirksam einschließlich geerbter; `Target` aufgelöst) · `SPEC-160` Nr. 9 (was der Dienst prüft, nicht der Parser) · `SPEC-263` (Vertrag von `PruefeExpand`) · `SPEC-167` (Muster: reine Funktion in `model`, eigener Fehlertyp) · `SPEC-022` (`ODR-E6001`), `SPEC-163` (`ODR-E6005`) · `SPEC-049` (Zeile GET: Navigation-Expand) · `ARC-004`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Eine reine Funktion im Paket `src/hexagon/model` (Muster `PruefeFilter`, `SPEC-167`) prüft jede Navigation des Knotens `$expand` (geliefert von `slice-expand-parser`) gegen die Navigation-Properties des Entity-Typs im Modell des Dienstes, die geerbten eingeschlossen: eine vorhandene Navigation ist gültig, ein unbekannter Name oder eine Property ohne Navigation ist `400` (`ODR-E6005`), eine gültige, aber nicht unterstützte Form `501` (`ODR-E6001`), nach `SPEC-263` (§6 V1 bis V7). Die Funktion liest nur den Knoten und das Modell, sie greift auf keine Daten zu und kennt weder Kontext noch Port (`LH-FA-16.a`: Validator-Contract-Test).

**Übernimmt (Schnitt von `slice-expand-parser`, Geber `slice-expand-parser`, `AGENTS.md` §3.13):** den Validator gegen das Modell samt Fehlertyp, Raster und Validator-Contract-Test, die der frühere Plan von `slice-expand-parser` führte (Navigation vorhanden, auch geerbt; Property ist eine Navigation; Zieltyp aufgelöst; unbekannt oder keine Navigation `400`, gültig, aber nicht unterstützt `501`; kein Zugriff auf Daten).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Parser, Grammatik, Knoten und Einhängestelle von `$expand` — `slice-expand-parser`; dieser Slice liest den Knoten und ändert ihn nicht (`SPEC-160` Nr. 3 gilt für den Parser).
- Der Aufruf des Validators im Dienst und die Verdrahtung — `slice-expand-ausfuehrung` (dessen §1 trägt die Sendung); dieser Slice liefert die Funktion, nicht ihren Aufrufer.
- Auflösen gegen Daten, Datenbasis der Navigationen, Ausgabe — `slice-expand-ausfuehrung`.
- Navigation ohne `NavigationBinding` am Set und fehlende Daten der Navigation — `slice-expand-ausfuehrung` (dessen §1 trägt die Sendung, `AGENTS.md` §3.13); die Funktion prüft den Typ, nicht die Bindings (`SPEC-263` Nr. 5).
- Änderungen am Metadatenleser und am Modell (`SPEC-125`): `Path` der `NavigationPropertyBinding` wird nicht aufgelöst (`SPEC-125` *Ungeprüft*), und die Funktion liest, was das Modell liefert; eine Änderung des Modells ist ein Vertrag in `done/` und keine Adresse für eine Sendung (`AGENTS.md` §3.13).
- Herstellerspezifische Navigations-Erweiterungen — Out-of-Scope von `LH-FA-16`.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste.

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Funktion und Fehlertyp: jede Navigation des Knotens gegen die wirksamen Navigation-Properties des Typs (auch geerbt); Property ohne Navigation, unbekannter Name und die nicht unterstützte Form mit Ursache, Klasse, Code und Status nach dem Vertrag des Architect; kein Teil der Eingabe im Fehler; kein Zugriff auf Daten.
- [ ] Randformen (Programmfehler des Aufrufers, Wiederholung, Nebenläufigkeit, Reihenfolge mehrerer Verstöße, Subtyp, Navigation ohne Binding): je Zelle des Rasters (§6) ein Ort in der Spezifikation und ein Fänger der Tabelle (§3), als Schleife über alle Navigationen und alle Properties des Fixture-Modells.
- [ ] Validator-Contract-Test (`LH-FA-16.a`): ein eigener Test fährt die Funktion gegen die Beispiele der Spezifikation; `make a-check` hält die Richtung (`model` importiert keinen Port und keinen Dienst).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (entschieden: `SPEC-109` bleibt, `ExpandFehler` trägt die Sentinel der URL-Klassen, `SPEC-263` Nr. 3).
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen; das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (die Funktion hat keinen Aufrufer vor `slice-expand-ausfuehrung`).
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
| `src/hexagon/model/` | neu | `PruefeExpand` und der Fehlertyp (Muster `filter_typ.go`, `filter_typ_fehler.go`), technologieunabhängig |
| `src/hexagon/ports/outbound/abbildung.go` | unverändert | `ExpandFehler` trägt die Sentinel der URL-Klassen, die `SPEC-109` Nr. 5 schon abbildet (`SPEC-263` Nr. 3); keine Zeile neu |
| `spec/spezifikation.md` | update | Vertrag `SPEC-263` mit Randformen (Architect, geschrieben) |
| Tests | neu | Fälle je Navigation und je Fehlerklasse; Schleife über alle Navigationen und Properties des Fixture-Modells |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Größenmarke setzt der Planner beim Start nach `SPEC-128` aus dieser Tabelle und der Erwartung des Architect: `produktcode` rund 120 (die Funktion rund 50 Zeilen, der Fehlertyp rund 70 nach dem Muster von `filter_typ_fehler.go`, 76 Zeilen). Die Tabelle *Zusage · Fänger* füllt der Planner aus `SPEC-263` Nr. 1 bis 6 vor dem Start.

**Tabelle *Zusage · Fänger*** (füllt der Architect mit dem Vertrag, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen; die Menge der Fälle des Validators sind alle Navigationen und alle Properties jedes Typs des Fixture-Modells, einschließlich der geerbten.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-expand-parser` liegt in `done/` (der Knoten und seine Form, Entscheidung A8 des Architect dort). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-expand-parser` (Typ des Knotens in `src/hexagon/model/`; dieser Slice liest ihn und kann ihn nicht vorwegnehmen). Das Modell mit den Navigation-Properties (`SPEC-002`, `SPEC-125` Nr. 11) liegt in `done/` (`slice-odata-codec-csdl`). Der Welle-Trigger (`welle-mock` in `done/`) ist Reihenfolge-Vorsicht, keine Code-Abhängigkeit.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Query-Interpreter, JSON-Schreiber, Workspace, Composition Root, M2, `slice-bootstrap-m3` (die Funktion hat keinen Aufrufer), `slice-expand-ausfuehrung` (Folge-Slice), `slice-batch-parser`.
- **Größenmarke:** steht aus. Der Planner setzt sie beim Start, vor dem ersten Code-Commit, nach `SPEC-128` aus der Tabelle des Architect; die Faktoren stehen allein dort, die Produktcode-Erwartung des Architect ist ein fester Anteil. Schnittgrenze und Metrik wie `AGENTS.md` §3.14.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): die Form des Knotens aus `slice-expand-parser` trägt die Prüfung nicht, oder das Modell liefert die Navigationen eines Typs nicht so, wie `SPEC-125` Nr. 11 sie beschreibt.

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

- Die Menge der Fälle ist die Menge der Navigationen und Properties aller Typen eines Modells (unbeschränkt) — **Ausgang:** offen; Schleife über die Menge des Fixture-Modells mit Subtypen, Ketten und offenem Typ, Prüfgrenze und Folge als akzeptiertes Negativ (`AGENTS.md` §3.17).
- Fakes der Ports: die Funktion hat keinen Port — **Ausgang:** entfallen (`PruefeExpand` hat keinen Port, `SPEC-263`).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Raster unten, `SPEC-263`); die Tabelle *Zusage · Fänger* füllt der Planner.

**Entscheidungen des Architect — Ausgang *entschieden*, Ort `SPEC-263`** (`AGENTS.md` §3.13). Keine ADR: die Entscheidungen schärfen die Spezifikation (`AGENTS.md` §3.8) und ändern keine angenommene ADR.

- **V1 Signatur.** Entschieden (`SPEC-263` Kopf, Nr. 1): `PruefeExpand(typ EntityType, o *ExpandOption) error`; `typ` ist der Typ des Entity-Sets mit den wirksamen Properties und Navigation-Properties, den geerbten eingeschlossen; das Modell wird nicht gebraucht (die Zieltypen sind beim Lesen aufgelöst, `SPEC-125` Nr. 9 d); `o` ist `Optionen.Expand`, ein Aufrufer ohne `$expand` ruft nicht.
- **V2 Fehlerwerte.** Entschieden (`SPEC-263` Nr. 3): `model.ExpandFehler{Ursache}` mit `navigation_unbekannt` und `keine_navigation` (`ODR-E6005`, 400) und `navigation_nicht_unterstuetzt` (`ODR-E6001`, 501), Sentinel `ErrURLUngueltig` bzw. `ErrURLNichtUnterstuetzt`; kein Teil der Eingabe im Fehler; `abbildung.go` und `SPEC-109` bleiben.
- **V3 Reihenfolge.** Entschieden (`SPEC-263` Nr. 2): der erste Verstoß von links entscheidet, gleich welcher Klasse (wie `SPEC-167` Nr. 3); Programmfehler gehen allem vor (Nr. 4).
- **V4 Subtyp und Typ-Cast.** Entschieden (`SPEC-263` Nr. 2 d): eine Navigation, die nur ein abgeleiteter Typ trägt, ist `navigation_unbekannt` (`typ` ist der Typ des Sets); der Typ-Cast im Pfad ist schon im Parser `501`; ein abstrakter Typ des Sets verhält sich wie jeder andere.
- **V5 Navigation ohne Binding.** Entschieden (`SPEC-263` Nr. 5): gültig, die Funktion kennt Sets und Bindings nicht. Die Behandlung in der Ausführung trägt `slice-expand-ausfuehrung` (dessen §1, Sendung von `slice-expand-validator`).
- **V6 Property ohne Navigation.** Entschieden (`SPEC-263` Nr. 2 b, c): jede Property des Typs, auch mit nicht unterstütztem Typ, ist `keine_navigation` (400); ein Name im offenen Typ, der in keiner der beiden Listen steht, ist `navigation_nicht_unterstuetzt` (501, wie `SPEC-167` Nr. 5).
- **V7 Programmfehler.** Entschieden (`SPEC-263` Nr. 4): `o` nil, leere Liste, leerer Name, `*` und doppelter Name sind `model.ErrUngueltig` ohne Code, für den ganzen Knoten vor der Prüfung der Namen.

**Raster *Operation × Argumentart*** (Zellen gefüllt vom Architect; der Ort steht in `SPEC-263`):

| Operation | Knoten fehlt oder leer | unbekannte Navigation | fremder Wert (Property ohne Navigation, anderer Typ, Subtyp) | geerbte Navigation | Wiederholung (Aufruf, Name) | Nebenläufigkeit |
|---|---|---|---|---|---|---|
| `PruefeExpand` | `SPEC-263` Nr. 4 | `SPEC-263` Nr. 2 d und Nr. 3 | `SPEC-263` Nr. 2 b, c, d (Property des Typs, Name im offenen Typ, Navigation eines Subtyps, Name eines anderen Typs) | `SPEC-263` Nr. 1 und Nr. 2 a (wirksame Navigationen, geerbte eingeschlossen) | `SPEC-263` Nr. 6 (Aufruf) und Nr. 4 (doppelter Name) | `SPEC-263` Nr. 6 |

**Zusatztabelle: mehrere Verstöße in einem Aufruf** (je Paar das Ergebnis nach `SPEC-263` Nr. 2 und 4; die Spalte *Fänger* nennt die Zeile der Tabelle in §3, die der Planner aus `SPEC-263` Nr. 1 bis 6 füllt):

| Paar (links, rechts im Knoten) | Ergebnis | Fänger |
|---|---|---|
| unbekannte Navigation, Property ohne Navigation | `navigation_unbekannt` (der erste von links, Nr. 2) | Zeile der Tabelle in §3 (Planner) |
| Property ohne Navigation, unbekannte Navigation | `keine_navigation` (der erste von links, Nr. 2) | Zeile der Tabelle in §3 (Planner) |
| Name im offenen Typ, Property ohne Navigation | `navigation_nicht_unterstuetzt` (der erste von links, gleich welcher Klasse) | Zeile der Tabelle in §3 (Planner) |
| Property ohne Navigation, Name im offenen Typ | `keine_navigation` (der erste von links) | Zeile der Tabelle in §3 (Planner) |
| Verstoß, danach doppelter Name oder `*` | `model.ErrUngueltig` (Programmfehler vor jedem Verstoß, Nr. 4) | Zeile der Tabelle in §3 (Planner) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-18); der Planner zieht Treffer und Zähler beim Start nach. Die Treffer von `slice-expand-parser` §6 (`menge-ohne-ihre-positionen`, `menge-im-plan-weicht-vom-orakel-des-tests-ab`, `suite-ohne-generator-fuer-klassen-der-spezifikation`, `fix-deckt-die-genannte-mutation-nicht-die-klasse`) gelten für die Schleife über Navigationen und Properties. Der Slice ist nicht nebenläufig, bis auf die Zusage der gleichzeitigen Aufrufe (`SPEC-167` Nr. 8 als Muster).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
