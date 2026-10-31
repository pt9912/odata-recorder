# Slice slice-bootstrap-m4: Composition Root für den Modus hybrid und der M4-Test

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, `LH-QA-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` · `LH-FA-04.c` (Pflicht der `X-Recorder-*`-Header im Modus `hybrid`) · `SPEC-110` Nr. 2, 8, 9 (Schlüssel dieser Version, Modus, Services) · `SPEC-010`, `SPEC-062`, `SPEC-075`, `SPEC-076`, `SPEC-079`, `SPEC-080` · `SPEC-055` (`PUBLIC_BASE_URL` in `hybrid`) · `ARC-005`, `ARC-012` · `architecture.md` §2.5 (Modus-Routing).
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

**Ziel:** Der Composition Root bindet im Modus `hybrid` genau `HybridUseCase` an den OData-Listener, liest die Schlüssel `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` (Syntax und Zusammenhänge nach `SPEC-110`), verlangt bei freigegebenem Upstream-Fallback die `UPSTREAM_URL` des Service, und der M4-Test aus §3 der Welle läuft grün in `make test`: ein Referenzdienst zählt Verbindungen und Methoden und belegt die Entscheidungsmatrix.

**Übernimmt:** den Nachzug von `SPEC-110` Nr. 2 und 8 für den Modus `hybrid` und die Schlüssel `HYBRID_*` und `FALLBACK_*`, den `slice-bootstrap-m3` für den Modus `mock` hinterlässt (dessen §1 *Ausdrücklich NICHT* weist beides diesem Slice zu), und das Lesen der Policy-Werte aus der Umgebung (Sendung von `slice-fallback-policy`, dessen §1 *Ausdrücklich NICHT* diesen Slice nennt).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Neue Fachlogik — der Slice verdrahtet, was `slice-fallback-policy` und `slice-hybrid-service` liefern.
- Die Schlüssel `ADMIN_*` und `LISTEN_TLS_*` — `welle-admin-api` und `welle-betrieb`; sie bleiben `ODR-E2001`, damit keine Sicherheitseinstellung still wirkungslos ist (`SPEC-110` Nr. 2).
- Das Umschalten des Modus zur Laufzeit — `LH-FA-13` schließt es aus dem ersten Abnahmeumfang aus.
- Die Anzeige des Modus in der Diagnose — `welle-admin-api`; die Admin-API braucht den Hybrid-Betrieb nicht.
- `LINK_REWRITE` — bleibt `off` (`SPEC-092`).

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

- [ ] Verdrahtung und Konfiguration: der Modus `hybrid` bindet `HybridUseCase`; die vier Schlüssel kommen aus dem Wert des Starts bis in die Policy (Fänger, dass der Root sie weiterreicht und den Nullwert nicht stehen lässt; Muster aus `slice-upstream-tls`); `UPSTREAM_URL` ist bei `HYBRID_FALLBACK=enabled` Pflicht; Zusammenhänge und Randformen nach `SPEC-110`.
- [ ] M4-Test aus §3 der Welle (dazu als Stufe von `make test-integration` am gebauten Binary das Szenario des Verbindungszählers, `SPEC-037`, für die Kriterien, die die [Rückverfolgungs-Matrix](../rueckverfolgung.md) diesem Slice zuweist): ein Treffer des Mock ergibt 0 Verbindungen; eine lesende Anfrage bei unvollständiger Coverage mit Freigabe genau eine lesende Verbindung, ohne Freigabe `ODR-E6003` und 0; ein POST auf eine nicht unterstützte Operation ohne Eintrag in `FALLBACK_WRITE_ALLOW` 0, mit passendem Eintrag eine; ein Replay-Miss `ODR-E5001` mit Opt-in nur für GET und HEAD zum freigegebenen Host, ein erschöpfter Key `ODR-E5002` und fachliche `404` und `409` 0.
- [ ] Nicht gelieferte Schlüssel (`ADMIN_*`, `LISTEN_TLS_*`) bleiben `ODR-E2001`, und die Spezifikation führt den Nachzug von `SPEC-110` Nr. 2 und 8.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand: Modus `hybrid`, Schlüssel `HYBRID_*` und `FALLBACK_*`, die Fehlercodes des Fallbacks, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Voraussetzung ist `slice-handbuch-meldungskatalog` in `done/` (Trigger von `welle-mock`).
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
| `src/bootstrap/` | update | Wahl des Use Case je Modus um `hybrid` erweitert; die vier Schlüssel lesen, prüfen und an die Policy geben; `UPSTREAM_URL` als Pflicht bei freigegebenem Upstream-Fallback |
| `src/adapters/driving/httpodata/` | — | kein Änderungsbedarf erwartet, der Adapter kennt den Modus nicht; ein Bedarf geht an den Architect (§6) |
| `tests/` | neu | M4-Test mit dem Referenzdienst aus `slice-bootstrap-m2`, der Verbindungen und Methoden zählt (ob er nach Methode zählt, steht in §6) |
| `spec/spezifikation.md` | update (Architect) | Nachzug von `SPEC-110` Nr. 2 und 8, Syntax und Zusammenhänge der vier Schlüssel |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

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

1. **Nachzug von `SPEC-110`:** Nr. 2 (die vier Schlüssel wandern aus der Ablehnung in die geprüften), Nr. 8 (`hybrid` lieferbar) und Syntax nach Nr. 3 und 6: `HYBRID_FALLBACK` (`disabled` oder `enabled`), `HYBRID_REPLAY_FALLBACK` (Liste der Kategorien), `FALLBACK_ALLOW_HOSTS` (Liste), `FALLBACK_WRITE_ALLOW` (Liste von Einträgen in der Syntax aus `slice-fallback-policy`); Randformen je Schlüssel (leer, unbekannte Kategorie, Doppelte, ungültiger Eintrag).
2. **Zusammenhänge:** die Zusammenhänge aus Auftrag 6 von `slice-fallback-policy`; `UPSTREAM_URL` je Service Pflicht bei `HYBRID_FALLBACK=enabled` (`SPEC-076`); gesetzte `HYBRID_*` und `FALLBACK_*` in den Modi `record`, `replay` und `mock`: ignoriert oder `ODR-E2001`.
3. **`PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid`** (`SPEC-055`, `SPEC-011`): Pflicht und Default wie im Modus `mock`, oder eigene Regel.
4. **Aufbau des M4-Tests:** Referenzdienst, der Verbindungen und Methoden zählt; die Fälle aus §3 der Welle als Tabelle; Fixture aus Import (Baseline) und Aufzeichnung (Recording-Session und Header für den Replay-Fallback); Ordnung unter dem Detektor (Kanäle statt Schlaf, Muster aus `slice-bootstrap-m2`).
5. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* aus `slice-bootstrap-m3` (Nachzug für `hybrid`, `HYBRID_*`, `FALLBACK_*`) und `slice-fallback-policy` (Lesen der Schlüssel): angenommen, §1 *Übernimmt* trägt sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie. *Von diesem Slice an:* keine.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-hybrid-service` und `slice-bootstrap-m3` liegen in `done/`. Die Spezifikation trägt den Nachzug von `SPEC-110`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** `slice-hybrid-service` liefert `HybridUseCase`, den der Root bindet (*Code-Abhängigkeit*). `slice-bootstrap-m3` liefert den Mock-Pfad im Composition Root (Bau von `MockUseCase`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`, den Nachzug von `SPEC-110` für `mock`) und die Fixture aus Import, Workspace und Referenzdienst, die der M4-Test erweitert (*Code-Abhängigkeit*: dieselben Dateien in `src/bootstrap/`). `ReplayUseCase` und der Upstream-Adapter liegen über `slice-bootstrap-m2` im Root vor.

**Nicht ohne den Mock:** der Slice braucht `welle-mock` ganz (über `slice-bootstrap-m3`) und M3. **Reihenfolge zu `slice-bootstrap-m6`:** beide ändern `src/bootstrap/` und `SPEC-110` Nr. 2; sie laufen nacheinander und nicht parallel, welcher zuerst, ist Priorisierung (`welle-admin-api` ist von `welle-hybrid` unabhängig), und der zweite zieht den Stand des ersten nach (`AGENTS.md` §3.9).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): der Nachzug von `SPEC-110` für `hybrid` liegt nicht vor, oder der Referenzdienst lässt sich nicht nach Methode zählen, ohne die Fixture aus `slice-bootstrap-m2` zu ändern (siehe §6).

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
| Lesen von `HYBRID_FALLBACK` und `HYBRID_REPLAY_FALLBACK` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| Lesen von `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| Zusammenhänge der Schlüssel und `UPSTREAM_URL` im Modus `hybrid` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| Start und Herunterfahren im Modus `hybrid` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- Der Referenzdienst aus `slice-bootstrap-m2` zählt Verbindungen, ob er nach Methode zählt, steht nicht in dessen Plan — **Ausgang:** offen; Adresse: Architect (Auftrag 4); die Fixture ist dann in diesem Slice zu erweitern.
- Gesetzte `HYBRID_*` und `FALLBACK_*` in den Modi `record`, `replay` und `mock` (Auftrag 2) — **Ausgang:** offen; Adresse: Architect.
- Der M4-Test läuft mit Referenzdienst, Recorder und Client im selben Prozess unter dem Detektor (`make test-race`) — **Ausgang:** offen; Muster und Ordnung der Fänger wie in `slice-bootstrap-m2` (Kanäle statt Schlaf).
- `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×): der Hybrid-Pfad im Composition Root erweitert die Pakete, deren Laufzeit Fälle der Gegenproben mitfahren — **Ausgang:** weiter offen → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` im Register.
- Gleichzeitige Änderung von `src/bootstrap/` und `SPEC-110` Nr. 2 durch `slice-bootstrap-m6` — **Ausgang:** offen; Adresse: Planner (Reihenfolge beim Start, keine parallelen Arbeitsbäume auf diesen Dateien).
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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, Prozess, Referenzdienst und Client unter Last) sowie `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `AGENTS.md` §3.9; der Nachzug von `SPEC-110` und die Welle-Datei sind Stellen, die eine Berichtigung tragen) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
