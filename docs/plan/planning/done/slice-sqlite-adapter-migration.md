# Slice slice-sqlite-adapter-migration: SQLite-Adapter mit Migration aus dem Schema

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-persistenz](../welle-persistenz.md).

**Bezug:** `LH-FA-03`, `LH-QA-06`, [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0003](../../adr/0003-lint-profil-golangci.md)

**Berührte Spec-Stellen:** `SPEC-004` (Normatives Schema) · `SPEC-014` (`SQLITE_BUSY_TIMEOUT`) · `SPEC-103` (Öffnen und Migrieren) · `SPEC-104` (Migrations-Gate) · `SPEC-102` (Gegenproben) · `ARC-009`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-19.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Driven Adapter `src/adapters/driven/sqlite` öffnet eine SQLite-Datei, wendet die eingebettete Migration `0001` an, und ein Gate hält diese Migration gleich dem DDL, das `tools/schema/schema.yaml` erzeugt.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Repositories und Ports — folgen in [slice-session-recording-persistenz](../next/slice-session-recording-persistenz.md) und den weiteren Slices der Welle; dieser Slice liefert nur Öffnen, Migration und Integritätsprüfung.
- Migrationskette für eine zweite Schemaversion — das Schema hat eine Version, die Kette braucht eine zweite (`harness/README.md`, *Nicht behauptet*).
- Die Abbildung von `ErrBeschaedigt` auf `ODR-E3001` — sie gehört der Schicht, die Fehler dem Nutzer meldet (Services und Adapter der Folgewellen); es gibt dafür noch keinen Slice, der Punkt geht an den Planner und wird beim Planen von welle-record-replay vergeben (`AGENTS.md` §3.13).
- Entscheidung über den Treiber — [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) ist angenommen; dieser Slice setzt sie um.

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

- [x] Treiber `modernc.org/sqlite` in `go.mod` und `go.sum` gepinnt, `tech`-Regel in `.a-check.yml` und Eintrag in `gomodguard_v2`; `sqlite.Open` setzt `foreign_keys`, `busy_timeout` und den Journal-Modus und wendet die Migration an.
- [x] Gate `migration-check` (an `GATE_CHECKS`): die eingebettete `0001_init.sql` ist byte-gleich dem DDL aus `make schema-generate`; ein Unterschied färbt das Gate rot.
- [x] Tests am Adapter: leere Datei wird migriert, `foreign_key_check` und `integrity_check` sind leer, ein zweites Öffnen ändert nichts, die Randformen aus §6 sind getestet.
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `go.mod`, `go.sum`, `Dockerfile` | update | Treiber in der Stufe `deps`; Netz bleibt auf diese Stufe beschränkt |
| `.a-check.yml`, `.golangci.yml` | update | `tech`-Regel und `gomodguard_v2`-Eintrag für `modernc.org/sqlite` ([ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)) |
| `src/adapters/driven/sqlite/` | neu | `Open`, `Migrate`, eingebettete `migrations/0001_init.sql` |
| `harness/mk/schema.mk`, `tools/harness/migration-check.sh`, `harness/vertraege/SPEC-104-migration-check.md` | update / neu | Gate `migration-check` und sein Vertrag (Randformen vor dem Code) |
| `tools/gegenproben/`, `Dockerfile`, `.dockerignore`, `tools/harness/gegenproben-eingabe.sh`, `harness/vertraege/SPEC-102-gegenproben.md`, `harness/README.md` | update | Fälle 17 bis 24 für `migration-check.sh` und die Bibliotheksgrenze; Modulcache und das Skript gehören in die Stufe `gegenproben`, die neuen Dateien in die Eingaben |
| `spec/spezifikation.md` | update | `SPEC-103`: Öffnen und Migrieren mit den Randformen (leerer Pfad, Verzeichnis, fremde Datenbank, höhere Schemaversion, beschädigte Datei, gleichzeitiges Öffnen) |

- Die Migration ist das DDL aus dem Schema, kein Handtext: `make schema-generate` liefert es, das gewählte Verfahren zum Einbetten steht im Vertrag des Gates.
- Eine beschädigte Datei liefert `ErrBeschaedigt`; die Abbildung auf `ODR-E3001` (Recording beschädigt) gehört der meldenden Schicht (siehe §1).

**Zusage · Fänger** (`AGENTS.md` §3.15):

| Zusage | Fänger |
|---|---|
| `Open` setzt `foreign_keys=ON` | Test: Fremdschlüsselverletzung wird abgelehnt; Mutation: Pragma entfernt |
| Die Migration läuft genau einmal | Test: zweites Öffnen, `user_version` bleibt; Mutation: Migration bei jedem Öffnen |
| Höhere Schemaversion wird verweigert | `TestNeuereSchemaversionWirdVerweigert`; Mutation: Zweig entfernt |
| `busy_timeout` und `journal_mode` auf jeder Verbindung | `TestPragmasJederVerbindung`, `TestBusyTimeoutAusDerOption`; Mutation: Wert fest |
| Fremde Datenbank wird verweigert | `TestFremdeDatenbankWirdVerweigert`; Mutation: Prüfung abgeschaltet |
| Keine SQLite-Datei und abgeschnittener Rumpf sind `ErrBeschaedigt` | `TestKeineSQLiteDateiIstBeschaedigt`; Mutation: Klassifizierung abgeschaltet |
| Ungültige Pfade, kein Anlegen von Verzeichnissen | `TestUngueltigePfade` |
| Gleichzeitiges Öffnen migriert einmal | `TestGleichzeitigesOeffnenMigrierteinmal`; Mutation: ohne Transaktion |
| Migration ist atomar | `TestMigrationLaeuftInEinerTransaktion`; Mutation: ohne `BEGIN` |
| Zweites `Close` ist kein Fehler | `TestCloseIstIdempotent` |
| Jedes Öffnen stellt WAL sicher (auch bei Modus `delete`) | `TestOeffnenStelltWalWiederHer`; Mutation: Aufruf entfernt |
| Das Umstellen wiederholt sich bei `SQLITE_BUSY` bis zur Frist, nur dann | `TestWalUmstellenWiederholtBeiBelegt`, `TestWalUmstellenGibtNachDerFristAuf`, `TestWalUmstellenWiederholtNurBeiBelegt`; Mutationen: keine Wiederholung, falsche Codes, jeder Fehler als belegt |
| Umstellen nach der Frist scheitert, ein späteres Öffnen holt nach | `TestWalUmstellenScheitertNachDerFristUndWirdNachgeholt`; Mutation: Aufruf entfernt |
| Kontextabbruch beendet die Wiederholung | `TestWalUmstellenEndetBeiKontextabbruch`; Mutation: Abbruch liefert `nil` |
| `migration-check.sh` lehnt jede falsche Argumentzahl ab | Fälle `migration/falsche-aufrufform`, `migration/zu-viele-argumente`; Mutation: `-eq` zu `-ge` |
| Gleichzeitiges Öffnen gelingt dauerhaft | 500 Wiederholungen von `TestGleichzeitigesOeffnenMigrierteinmal` einmalig grün (347 s), statistisch, keine Mutation |
| Andere Fehler bleiben unklassifiziert | `TestAndereFehlerBleibenUnklassifiziert`; Mutation: jeder Fehler beschädigt |
| Abgelehnte Datei bleibt byte-gleich | `zustand`-Prüfung in `TestNeuereSchemaversionWirdVerweigert`, `TestFremdeDatenbankWirdVerweigert`, `TestKeineSQLiteDateiIstBeschaedigt`, `TestAbgeschnittenerRumpfIstBeschaedigt` |
| Die Bibliothek bleibt im Adapter | Fall `a-check/bibliothek-in-fremdem-adapter` (`tech-leak`) |
| `migration-check.sh` meldet fehlende Dateien und falschen Aufruf | Fälle `migration/migration-fehlt`, `migration/erzeugtes-ddl-fehlt`, `migration/falsche-aufrufform` |
| `migration-check` meldet Abweichung und weitere Dateien | Fälle `migration/abweichung`, `migration/weitere-datei`; Mutationen: `cmp` und Dateischleife entfernt |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md) ist `Accepted` und `welle-fundament` liegt in `done/`.

**Marke überschritten:** Die Summe aus `git diff --numstat` über die Pfade aus §3 liegt bei 1202 Zeilen (alle Pfade aus §3, go.sum eingeschlossen); getragen von der erzeugten Migration (163 Zeilen), den Tests des Adapters und den Fällen der Gegenproben. Der Slice bleibt ungeteilt, weil Adapter, Gate und Gegenprobe einzeln nichts liefern.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 400 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert): der Treiber lässt sich ohne C-Toolchain nicht im Container bauen.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün mit `migration-check` in der Kette, eine geänderte Migration färbt es rot, bevor sie zurückgenommen wird, und `make gegenproben` ist grün; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen des Öffnens und des Gates (`AGENTS.md` §3.12): fremde Datei, neuere Schemaversion, beschädigte Datei, gleichzeitiges Öffnen, fehlende Migration — **Ausgang:** entfallen: entschieden in `SPEC-103` und `SPEC-104` vor dem ersten Code-Commit; das Review fand dennoch zwei Abweichungen von `SPEC-103` (WAL vor der Versionsprüfung, flackernde Gleichzeitigkeit), beide behoben.
- Der Treiber zieht Abhängigkeiten, die `go mod verify` oder das netzlose Bauen stören — **Ausgang:** entfallen: `make test`, `make lint`, `make gegenproben` und `make gates` laufen mit dem Treiber; nur die Stufe `deps` braucht Netz.
- Das Einbetten der Migration weicht vom erzeugten DDL ab, ohne dass jemand es merkt — **Ausgang:** entfallen: `migration-check` und sein Gegenprobe-Fall.

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

- **Was hat funktioniert:** Der Adapter öffnet und migriert die Datei nach `SPEC-103`, das Gate `migration-check` hält die Migration gleich dem DDL aus dem Schema, und die Fälle 17 bis 24 belegen es; je Zusage wurden mindestens zwei Mutationen rot gesehen, die Gleichzeitigkeit mit 2000 grünen Wiederholungen.
- **Was ging anders als geplant:** Das Review fand zwei HIGH (WAL vor der Versionsprüfung, flackernde Gleichzeitigkeit), die erste Verifikation fiel durch, weil das Umstellen auf WAL unter Last mit `SQLITE_BUSY` scheiterte; die Lösung ist das Sicherstellen von WAL bei jedem erfolgreichen Öffnen mit Wiederholung bis zur Frist. `SPEC-103` bekam dafür eine Randform (Wiederholung, Kontextabbruch), `SPEC-104` Fänger für Meldung und `diff`. Die Rückführungs-Marke ist überschritten und in §4 benannt.
- **Steering-Loop-Eintrag:** Nichts verkörpert; gezählt, nicht verkörpert: `BEO-REPO/randform-erst-im-review-entschieden` steht jetzt bei 3× (der Eintrag geht in die Steering-Loop-Einträge der Closure von welle-persistenz), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` bei 2× und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` bei 5× (verkörpert in `AGENTS.md` §3.15).
- **Beobachtungs-Register (`../observations/`):** je eine weitere Datei `evidence/slice-sqlite-adapter-migration.md` in den drei Einträgen.
- **Folge-Slices:** keine; die Abbildung von `ErrBeschaedigt` auf `ODR-E3001` steht in der Roadmap bei welle-record-replay, die Bibliotheksgrenze für die Composition Root bleibt offen (a-check nimmt sie aus), und die Zusage zur Gleichzeitigkeit ist statistisch belegt.
- **Risiken aus §6:** Randformen und Treiber *entfallen*, Migrations-Einbettung *entfallen*, siehe §6.
- **Drei Paarungen:** von der Closure der Welle welle-persistenz geprüft.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (2×), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (1×) und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15); sie fließen als Randformen in §6 und als Tabelle *Zusage · Fänger* in §3 ein.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
