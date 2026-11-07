# Slice slice-inspection-use-case: InspectionUseCase, Exchanges, Coverage und Diagnose lesen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-FA-11`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Exchanges, Coverage, Diagnose) · `SPEC-005` (Zeilen `GET …/exchanges`, `GET /__admin/coverage`, `GET /__admin/diagnostics`) · `LH-FA-11.a` · `SPEC-044`, `SPEC-043` (Coverage und Evidenz) · `SPEC-105`, `SPEC-107` (Repositories, aus denen gelesen wird) · `SPEC-109` (Fehlerabbildung) · `SPEC-071`, `SPEC-112` · `ARC-002`, `ARC-003` · `architecture.md` §2.2 (Zeile `InspectionUseCase`).
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

**Ziel:** `InspectionUseCase` (Inbound Port in `src/hexagon/ports/inbound`, Dienst in `src/hexagon/services/inspection`) liest die Exchanges einer Session, die Coverage einer Baseline und die Diagnose (Konfigurationszusammenfassung ohne Geheimnisse, Modus, Zähler je Fehlercode) paginiert und ohne Geheimnisse; er verändert nichts und kennt kein HTTP.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Endpunkte, Pfade und Statuscodes — `slice-admin-sessions`; der Use Case kennt kein HTTP, der Adapter bildet ab.
- Anlegen, Beenden und Löschen von Sessions — `SessionUseCase` (`done/`); `InspectionUseCase` liest nur.
- Der Port `Telemetry` und die Zähler selbst — `welle-betrieb` (`slice-telemetry-port`); der Dienst liest die Zähler über den Port, er erzeugt sie nicht.
- Export und Abruf der Metriken und Spans — `welle-betrieb` (Welle §6).
- Audit des Lesens — `SPEC-005` auditiert die Schreibendpunkte; ein Lesezugriff schreibt keinen Eintrag.
- Den Import-Report lesen — `ImportUseCase.Get` (`done/`); der Endpunkt liegt in `slice-admin-imports`.

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

- [ ] Exchanges: die Exchanges einer Session paginiert (Seitenform nach Vertrag des Architect); eine Session ohne Exchange ergibt eine leere Seite, eine unbekannte Session `ODR-E7003`; die Felder enthalten keine Geheimnisse (Werte von Header, Query und Body nur redigiert oder nicht enthalten, nach Vertrag).
- [ ] Coverage: die Einträge je EntitySet einer Baseline mit Status und Evidenz aus `CoverageRepository`; Parameter und Fehlerfall einer unbekannten Baseline nach Vertrag.
- [ ] Diagnose: Konfigurationszusammenfassung aus einer Allowlist von Schlüsseln (nie Token, Zugangsdaten oder Inhalte von Geheimnisdateien), Modus und Zähler je Fehlercode über den Port `Telemetry`.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/hexagon/ports/inbound/` | update | `InspectionUseCase` |
| `src/hexagon/services/inspection/` | neu | Lesen der Exchanges, der Coverage und der Diagnose; Seitenbildung |
| `src/hexagon/ports/outbound/`, `src/adapters/driven/sqlite/` | update, nur wenn der Architect eine Seitenabfrage für `RecordingRepository` entscheidet | `ListBySession` liefert heute alle Exchanges einer Session (§6) |
| `spec/spezifikation.md` | update (Architect) | Vertrag samt Seitenform, Feldern der Exchanges, Parametern der Coverage und Allowlist der Diagnose |
| Tests | neu | Fälle je Operation; die Fakes der Repositories und des Ports `Telemetry` halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

| Zusage | Fänger |
|---|---|
| *(Architect: eine Zeile je Zusage, Randform und Fall des Vertrags)* | *(ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates)* |

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Vertrag von `InspectionUseCase`:** Operationen (Exchanges, Coverage, Diagnose), Argumente und Rückgabetypen, Fehlerabbildung als Anwendung von `SPEC-109`, Ort in der Spezifikation.
2. **Paginierung:** `SPEC-005` und die Architektur sagen „paginiert“, kein Vertrag nennt Parameter (Seitengröße, Zeiger oder Versatz), Standard und Obergrenze, Reihenfolge, Verhalten bei ungültigem oder zu großem Wert, Stabilität bei gleichzeitigem Anhängen, Gesamtzahl. `RecordingRepository.ListBySession` liefert alle Exchanges: erweitert der Architect den Port um eine Seitenabfrage (berührt `SPEC-105` und den SQLite-Adapter und damit die Größe) oder schneidet der Dienst im Speicher (Grenze für große Sessions)?
3. **Felder eines Exchange in der Liste:** Kopfdaten (Kennung, Sequenz, Methode, Pfad, Status, `capture_status`, Zeitpunkte, Vermerke der Redaction) und ob Header-Werte, Query und Bodies (oder nur Größe und Hash) erscheinen; „ohne Geheimnisse“ als Allowlist der Felder, nicht als Liste der entfernten.
4. **Coverage:** `SPEC-005` nennt für `GET /__admin/coverage` keinen Parameter. Auf welche Baseline bezieht sich der Aufruf (Baseline-Kennung, Workspace, Service, alle), Form der Antwort je EntitySet mit Evidenz, Fehlerfall einer unbekannten Baseline (Code?).
5. **Diagnose:** Allowlist der Schlüssel der Konfigurationszusammenfassung (nie Token, Pfade von Geheimnisdateien als Inhalt, Zugangsdaten), Form der Zähler je Fehlercode, und die Operation des Ports `Telemetry`, die Zähler liest (heute gibt es den Port nicht; Anforderung an `slice-telemetry-port`, §6); die Zusammenfassung kommt als typisierter Wert vom Composition Root, der Dienst sieht nie die Umgebung.
6. **`ODR-E7012`:** `SPEC-112` kennt „Exchange unbekannt“ (Admin), `SPEC-005` hat keinen Endpunkt für einen einzelnen Exchange: Zeile ergänzen oder den Code anders verorten.
7. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-telemetry-port` (`welle-betrieb`) liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, Seitenform und Allowlist, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum, und für welchen Teil:** nur der Teil *Diagnose* liest Zähler je Fehlercode über den Port `Telemetry`, den es heute nicht gibt (`src/hexagon/ports/support/` führt `Clock` und `IDGenerator`): *Code-Abhängigkeit* für diesen Teil. Die Teile *Exchanges* und *Coverage* lesen `RecordingRepository`, `SessionRepository` und `CoverageRepository`, die seit `welle-persistenz` in `src/hexagon/ports/outbound/` liegen, und brauchen nichts davon.

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`** (Mock, `WorkspaceUseCase`, HTTP-Adapter, Composition Root): der Dienst kennt weder HTTP noch einen Modus. Starteten Exchanges und Coverage vor dem Port, wäre das ein Schnitt der Diagnose in einen eigenen Slice (Umplanung durch den Planner); dieser Plan hält den Slice ganz und nennt den Port als Trigger. Ob `slice-telemetry-port` selbst vor dem Mock starten darf, entscheidet die Planung von `welle-betrieb`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): der Port `Telemetry` bietet keine Operation, die Zähler liest, oder die Seitenabfrage verlangt eine Änderung von `RecordingRepository` und SQLite-Adapter, die die Marke sprengt (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): „(Vertrag, offen)“ heißt, der Architect trägt den Ort in der Spezifikation vor dem Übergang `next` → `in-progress` ein; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Exchanges einer Session lesen (Session, Seitenparameter) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Coverage lesen (Baseline oder Workspace) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Diagnose lesen (Konfiguration, Modus, Zähler) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- Die Paginierung hat keinen Vertrag (Auftrag 2) — **Ausgang:** offen; Adresse: Architect, vor dem ersten Code-Commit.
- Eine Seitenabfrage für `RecordingRepository` berührt `SPEC-105`, den Port und den SQLite-Adapter und damit Größe und Schnitt — **Ausgang:** offen; Adresse: Architect (Auftrag 2), danach der Planner für die Marke.
- Der Port `Telemetry` (`slice-telemetry-port`, `welle-betrieb`) nennt „Zähler je Fehlercode“, aber kein Lesen der Zähler über den Port; die Anforderung steht noch nicht im Plan des Gebers, dessen Datei fehlt — **Ausgang:** offen; Adresse: Planner von `welle-betrieb` (die Sendung gilt erst als angenommen, wenn dessen §1 oder DoD sie trägt, `AGENTS.md` §3.13).
- Die Parameter von `GET /__admin/coverage` und der Code für eine unbekannte Baseline stehen nicht in `SPEC-005` (Auftrag 4), ebenso `ODR-E7012` ohne Endpunkt (Auftrag 6) — **Ausgang:** offen; Adresse: Architect.
- Geheimnisse in Exchanges und Diagnose: die Allowlist der Felder (Aufträge 3 und 5) ist der Schutz; der Fänger über alle Artefakte ist der Canary-Test in `slice-test-security` — **Ausgang:** offen bis zur Closure; Adresse: Architect.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×): die Fakes der Repositories und des Ports `Telemetry` tragen jede Schnittstelle des Originals und zählen jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** offen; Adresse: Planner beim Start (§4).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes der Repositories und des Ports `Telemetry`, außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `AGENTS.md` §3.16, für „ohne Geheimnisse“). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
