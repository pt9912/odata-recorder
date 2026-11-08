# Slice slice-handbuch-mock: Kapitel zum Modus mock mit gebundenem Beispiel und die Umgebung im Beispiel-Format

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md) (aus `welle-mock` verlegt, Entscheidung des Users, 2026-11-08).

**Bezug:** `LH-QA-05` (Fehlerdiagnose), `LH-FA-06` (Mock), `LH-FA-12` (Betriebsarten), `LH-FA-15` (ETag, Beispiel zu `ETAG_ENABLED`), `LH-RB-01` (unterstützter Teilumfang), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-200` Nr. 1, 6 d, 7 b und die zugehörigen Nummern des Beispiel-Formats und des Laufs (`harness/vertraege/SPEC-200-handbuch.md`; der Architect erweitert sie vor dem Code) · `SPEC-110` (Startkonfiguration, Zeilen `STRICT_COVERAGE` und `PUBLIC_BASE_URL`) · `SPEC-138` (Signalverhalten des Dauerlaufs) · `SPEC-201` (Kataloge im Abgleich mit dem Handbuch).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Handbuch beschreibt den Modus `mock` (Start ohne Upstream, Antworten aus dem Zustand eines Workspace, die Wirkung von `STRICT_COVERAGE`, `PUBLIC_BASE_URL` und `ETAG_ENABLED`) in einem Kapitel mit mindestens einem gebundenen Beispiel; die Tabelle *Einstellungen* trägt die Zeilen `STRICT_COVERAGE` und `PUBLIC_BASE_URL`. Dafür erweitert der Slice den Vertrag `SPEC-200` und das Werkzeug `pruefung handbuch` um eine Umgebung im Beispiel-Format und um einen Lauf, der einen Dauerlauf mit einem Signal beendet.

**Übernimmt (Sendung von `slice-bootstrap-m3`, Architect 2026-11-08, `AGENTS.md` §3.13):** Das Kapitel zum Modus `mock` samt Beispiel als Datei, die Zeilen `STRICT_COVERAGE` und `PUBLIC_BASE_URL` der Tabelle *Einstellungen* und die Erweiterung von `SPEC-200` Nr. 6 und 7 um Umgebung und einen Lauf, der den Dauerlauf mit einem Signal beendet. Grund beim Geber: `SPEC-200` Nr. 6 d und 7 b erlauben keine Umgebung, ein `start` mit leerer Umgebung läuft im Modus `mock` bis zur Zeitgrenze von 10 s, ein Verstoß gegen `STRICT_COVERAGE` oder `PUBLIC_BASE_URL` braucht eine Umgebung, und `SPEC-200` Nr. 1 lässt ein Verhalten nur als gebundenes Beispiel zu. Geliefert hat der Geber die acht Zeilen des Mock in der Tabelle *Fehlercodes*. Die Adresse nimmt an: kein Ausschluss eines Nachbarn trifft die Sendung (gelesen: `slice-handbuch-abnahme` §1 schließt Kapitel über Mock aus, weil sein Probelauf prüft, was das Handbuch zum Zeitpunkt der Closure sagt, und trifft damit ein Kapitel eines anderen Slice nicht; `slice-bootstrap-m3-e2e-mock` §1 und DoD schließen das Handbuch aus, ihr Binary ändert für den Anwender nichts; `welle-benutzerhandbuch` §6 schloss Kapitel über Mock aus und führt seit der Entscheidung des Users vom 2026-11-08 diesen Slice in §4; `welle-mock` schließt ohne ihn).

**Übernimmt (zweite Sendung, von `slice-etag-versionen`, Architect 2026-11-08, `AGENTS.md` §3.13):** das gebundene Beispiel zu `ETAG_ENABLED` (Start mit `ETAG_ENABLED=true`, Lesen, Ändern, neuer ETag; Geber `slice-etag-versionen`). Grund beim Geber: dasselbe Format-Problem wie bei der ersten Sendung, `SPEC-200` Nr. 6 kennt keine Umgebung, und `SPEC-200` Nr. 1 lässt die Aussage, `ETAG_ENABLED=true` liefere ETags, nur als gebundenes Beispiel zu. Die Zeile `ETAG_ENABLED` der Tabelle *Einstellungen* trägt der Geber, nicht dieser Slice; dieser Slice trägt das Beispiel als Datei und braucht dafür `slice-etag-versionen` in `done/`. Gesamt-Sendung an diesen Slice damit: ein Kapitel mit Beispielen, die eine Umgebung brauchen (`STRICT_COVERAGE`, `PUBLIC_BASE_URL`, `ETAG_ENABLED`; der Dauerlauf endet mit einem Signal), und die Erweiterung von `SPEC-200` Nr. 6 und 7, die alle drei trägt. Gelesen: §1 *Ausdrücklich NICHT* dieses Slice schließt das Beispiel nicht aus (kein Punkt nennt ETag, Lesen oder Ändern). **Grenze, die der Architect entscheidet:** „Lesen, Ändern, neuer ETag“ braucht gegen den laufenden Dauerlauf Anfragen an den Listener, und `SPEC-200` Nr. 6 d kennt weder Eingabe noch Listener. Ob das unter den letzten Ausschluss („Neue Gate-Klassen über Umgebung und Signal hinaus“) fällt, entscheidet der Architect vor dem Code: trägt er Anfragen auf Loopback (kein Netz nach außen, kein Mehrprozess-Lauf) in die Erweiterung, bleibt es in diesem Slice; hält er sie für eine eigene Klasse, geht das Beispiel zu `ETAG_ENABLED` als Rückgabe an den Planner (§6), und `ETAG_ENABLED` bleibt eine Zeile der Einstellungen ohne Wirkungsaussage.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Fehlercodes des Mock in der Tabelle *Fehlercodes* und die acht Zeilen in `harness/katalog-ausnahmen.txt` — geliefert von `slice-bootstrap-m3` (A16); dieser Slice setzt sie voraus.
- Binary, Verdrahtung, Konfiguration des Modus `mock` — `slice-bootstrap-m3`; ein Unterschied zwischen Spezifikation und Binary ist ein Befund an den Planner.
- Kapitel über Hybrid, Fallback, Admin-API, Snapshots über HTTP, Telemetrie, TLS — Folgewellen; das Kapitel entsteht in der Welle, die das Verhalten liefert (`welle-benutzerhandbuch` §6; der Modus `mock` ist davon ausgenommen und gehört in diesen Slice).
- Der Probelauf durch eine Rolle ohne Vorwissen — `slice-handbuch-abnahme`; dieser Slice belegt Beispiele durch das Gate, nicht die Verständlichkeit.
- Szenarien des gebauten Binary in `make test-integration` — `slice-bootstrap-m3-e2e-mock`; das Handbuch-Gate fährt seine eigene Stufe `handbuch`.
- Neue Gate-Klassen über Umgebung und Signal hinaus (Mehrprozess-Läufe, Netz, Upstream) — das Beispiel des Mock braucht sie nicht; was der Architect darüber hinaus für nötig hält, geht als Rückgabe an den Planner.

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

- [ ] `SPEC-200` und `pruefung handbuch` kennen eine Umgebung im Beispiel-Format und einen Lauf, der den Dauerlauf mit einem Signal beendet (Vertrag vom Architect vor dem Code; Tabelle *Zusage · Fänger* in §3, Raster in §6); die Gegenproben zu `pruefung` fangen je Zusage eine Mutation (`LH-QA-05`).
- [ ] Kapitel zum Modus `mock` im Handbuch mit mindestens einem gebundenen Beispiel (Ist-Zustand, belegt durch eine Datei gegen das gebaute Binary in der Stufe `handbuch`); die Aussagen zur Wirkung von `STRICT_COVERAGE` und `PUBLIC_BASE_URL` stehen nur als gebundenes Beispiel (`SPEC-200` Nr. 1) (`LH-FA-06`, `LH-RB-01`); dazu das gebundene Beispiel zu `ETAG_ENABLED` (Start mit `ETAG_ENABLED=true`, Lesen, Ändern, neuer ETag; Geber `slice-etag-versionen`), vorbehaltlich der Entscheidung des Architect zu Anfragen im Beispiel-Format (§1, §6) (`LH-FA-15`).
- [ ] Die Zeilen `STRICT_COVERAGE` und `PUBLIC_BASE_URL` stehen in der Tabelle *Einstellungen*; der Abgleich der Schlüssel nach `SPEC-201` ist grün (`LH-QA-05`).
- [ ] `make gates` grün, `make gegenproben` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Handbuch ist das Update; Kopf (Stand, Software-Version) nachgezogen; es entsteht kein Änderungsverlauf (`welle-benutzerhandbuch` §1).
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
| `harness/vertraege/SPEC-200-handbuch.md` | update | Umgebung im Beispiel-Format (Nr. 6), Lauf mit Signal (Nr. 7); der Architect schreibt sie vor dem Code |
| `tools/` (Programm `pruefung`, Stufe `handbuch`) | update | Umgebung und Signal im Lauf gegen das Binary; Pfade nennt der Architect |
| `tools/gegenproben/` und `harness/gegenproben.stempel` | update | je Zusage der Erweiterung ein Fall der Gegenprobe (Nr. 7 der Übergangsregeln: Kettenpfade) |
| `docs/user/benutzerhandbuch.md`, `docs/user/beispiele/` | update, neu | Kapitel zum Modus `mock`, Zeilen `STRICT_COVERAGE` und `PUBLIC_BASE_URL` der Tabelle *Einstellungen* (die Zeile `ETAG_ENABLED` trägt `slice-etag-versionen`), Beispiel-Datei(en) auch zu `ETAG_ENABLED` |

Der Architect füllt die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und ergänzt die Zeilen oben; sie ist beim Start leer.

| Zusage | Fänger |
|---|---|
| — | — (Architect) |

**Betriebsregeln für die Rollen dieses Slice:** Dateien des Repos nur mit `Edit` und `Write`, kein `sed -i`, kein Heredoc auf Repo-Dateien (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`). Werte mit Backslash-u-Folge, BOM, NBSP oder Zeichen jenseits von ASCII entstehen aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18). Docker-Läufe mit `--memory` und wenige parallel.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): `slice-etag-versionen` liegt in `done/` (Schlüssel `ETAG_ENABLED` im Binary), `slice-bootstrap-m3` liegt in `done/` (der Modus `mock`, `STRICT_COVERAGE` und `PUBLIC_BASE_URL` und die acht Zeilen der Tabelle *Fehlercodes* liegen damit in `main`) und kein anderer Slice der Gate-Kette liegt in `in-progress/` (Nr. 7 der Übergangsregeln). Der Architect hat `SPEC-200` erweitert und die Tabelle in §3 sowie das Raster in §6 gefüllt, bevor der erste Code-Commit entsteht. Beim Start liest der Planner §1 *Ausdrücklich NICHT* und die DoD der Nachbarn neu (`AGENTS.md` §3.13).

**Gate-Kette (Nr. 7 der Übergangsregeln):** Linie C, Kettenslice. Er ändert `tools/` (Programm `pruefung`, Stufe `handbuch`), `tools/gegenproben/` und den Stempel `harness/gegenproben.stempel`, also Pfade aus Nr. 7; er läuft nie zugleich mit einem anderen Kettenslice. Vor ihm gehalten: `slice-test-contract` (in `in-progress/` seit 2026-11-08); es warten außerdem `slice-gate-integration-determinism` (zurückgestellt in `open/`) und `slice-guard-schreibwege`. Der Planner ordnet die Reihenfolge beim Start. Ändert der Architect das `Dockerfile` oder `Makefile` zusätzlich, steht das in §3.

**Größenmarke** (`AGENTS.md` §3.14, vorläufig als Prosa, keine Rechnung des Architect): grob S rund 900 und M rund 1250, nach der Art der Änderung (Erweiterung eines Werkzeugs samt Gegenproben, Handbuch zählt nicht, `SPEC-128` Nr. 3 m), keine Schnittfrage erwartet. Der Block `groessenmarke` fehlt bewusst; der Planner setzt `basis`, `schaetzung` und `marke` nach `SPEC-128` Nr. 1 aus der Tabelle des Architect in eigenem Commit vor dem ersten Code-Commit (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Plans, Tests eingeschlossen). Das Beispiel zu `ETAG_ENABLED` ist ein Beispiel mehr und ändert die Prosa-Marke nicht; ob Anfragen im Beispiel-Format sie ändern, nennt der Architect (§6), die Marke setzt dann der Planner. Nebenläufig zählt der Lauf mit Signal, weil das Werkzeug ein Binary startet und beendet.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem Code über der Schnittgrenze nach `SPEC-128` Nr. 1; vorab benannter Teil: das Kapitel samt Beispiel als eigener Commit nach dem Werkzeug.
- `in-progress` → `open` (blockiert): der Architect findet, dass die Erweiterung ein Gate lockert oder das `Dockerfile` ändern muss, während `slice-gate-integration-determinism` oder ein anderer Kettenslice die Kette hält.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, `make handbuch-check` fährt das Beispiel des Mock gegen das gebaute Binary, und je Zusage der Erweiterung ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor. Der Slice gehört zu `welle-benutzerhandbuch` (Kapitel zum Modus `mock`, Closure-Trigger dort §3) und ist kein Kriterium von `welle-mock` mehr.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Erweiterung von `SPEC-200` (Umgebung im Beispiel-Format, Lauf mit Signal) braucht den Architect vor dem Code; sie ändert `pruefung handbuch`, seine Gegenproben und das Format der Beispiel-Datei und ist keine Entscheidung des Implementers (`AGENTS.md` §3.12) — **Ausgang:** offen; Adresse: Architect (vor dem ersten Code-Commit).
- Das Beispiel zu `ETAG_ENABLED` braucht Anfragen an den laufenden Listener, die das Format nicht kennt (`SPEC-200` Nr. 6 d); der Architect entscheidet vor dem Code, ob sie in die Erweiterung gehören (dann nennt er den Zuwachs für die Marke, der Planner setzt sie nach `SPEC-128` Nr. 1 neu) oder ob das Beispiel zurück an den Planner geht; die Prosa-Marke in §4 gilt für Umgebung und Signal. Der Slice startet erst, wenn `slice-etag-versionen` in `done/` liegt (der Schlüssel muss im Binary stehen) — **Ausgang:** offen; Adresse: Architect (Entscheidung), Planner (Start, Marke).
- Die Gate-Kette ist gehalten (`slice-test-contract`; `slice-gate-integration-determinism`, `slice-guard-schreibwege` warten), der Slice startet erst, wenn Nr. 7 frei ist — **Ausgang:** offen; Adresse: Planner (Reihenfolge beim Start).
- Das Handbuch sagt mehr zu als ein Test prüft (`BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `verkörpert` in `AGENTS.md` §3.15) — **Ausgang:** offen; Fänger: Tabelle *Zusage · Fänger* und Mutationen am Beispiel.
- Der Abgleich der Schlüssel liest die Zeile `Geprüft werden` von `SPEC-110` Nr. 2 (K 29, im Code 28 nach `slice-bootstrap-m3`); eine Zahl, die zwischen Geber und diesem Slice weiter wandert (etwa durch `slice-admin-authentisierung`), bricht die Zählstellen — **Ausgang:** offen; Adresse: Planner (Integrationsschritt beim Start, Zahl gegen `main` prüfen).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster unten als Gerüst; die Zellen füllt der Architect vor dem ersten Code-Commit, eine leere Zelle ist ein Fehler des Plans — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Umgebung im Beispiel-Format lesen | Architect | Architect | Architect | Architect | Architect |
| Lauf mit Signal beenden | Architect | Architect | Architect | Architect | Architect |
| Beispiel des Mock binden (Handbuch gegen Datei) | Architect | Architect | Architect | gilt nicht: reiner Lesevorgang, Architect bestätigt | Architect |

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

**Vorgelagert — offene Beobachtungen sichten:** Der Planner liest das Register (`../observations/`) beim Start neu, denn bis dahin schreiben andere Slices es fort; zu erwarten sind `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
