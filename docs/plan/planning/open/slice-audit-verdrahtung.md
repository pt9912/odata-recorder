# Slice slice-audit-verdrahtung: Audit-Log im Composition Root, Datei im Datenverzeichnis und Canary

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-RB-04`, `LH-FA-13`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)

**Berührte Spec-Stellen:** `LH-RB-04.a` · `LH-QA-04.a` (kein Geheimnis in den Einträgen) · `SPEC-139` Nr. 2 bis 4 (Build, eine Instanz je Prozess, Modi) · `SPEC-077` (`DATA_DIR`, Datei `audit.jsonl`) · `SPEC-234` Nr. 8 (Konfig der Dienste, eine Instanz) · `SPEC-235` Nr. 1, 2, 7, 8 (Pfad, Datei entsteht mit dem ersten Eintrag, Neustart, Name der Datei) · `ARC-020`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel (Teil V des Schnitts von `slice-audit-log` in drei Teile):** Der Composition Root (`src/bootstrap/`) erzeugt je Prozess genau einen `audit.New` über `<DATA_DIR>/audit.jsonl` mit der einen Uhr, dem einen Generator und dem einen Logger und gibt diese eine Instanz als `Audit` in `maintenance.Konfig`, `session.Konfig`, `workspace.Konfig` und `baselineimport.Konfig` (`SPEC-139` Nr. 3, `SPEC-234` Nr. 8). Nach einer Löschung steht der Eintrag in der Datei, im Modus `record` mit allen vier Arten, im Modus `replay` für Aufbewahrung und Workspace; ein zweites `Build` auf demselben Datenverzeichnis hängt an; ein Canary (Token, Userinfo der `UpstreamURL`, Fehlertext) steht in keinem Byte der Datei; ohne Löschung entsteht die Datei nicht (`SPEC-235` Nr. 2).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Port, Typen und Adapter — `slice-audit-log`; die Aufrufstellen und das Feld `Audit` der vier Konfigs — `slice-audit-dienste`. Grund: dieser Slice verdrahtet fertige Teile und schreibt keinen Dienst um.
- Audit-Einträge der Admin-Schreibendpunkte und der Eintrag zu `ODR-E2101` — `welle-admin-api` (`SPEC-005`, `LH-FA-13.a`); die Nehmer tragen die Entscheidung `SPEC-234` Nr. 7.
- Das Handbuch-Kapitel zur Datei `audit.jsonl` — mit `slice-admin-sessions` (erster Slice, dessen Binary eine Löschung auslöst); hier steht in §7, was das Binary nach der Verdrahtung tut und warum kein gebundenes Beispiel die Datei erzeugen kann (siehe §6).
- Die Abbildung auf die Telemetrie (`audit_fehlgeschlagen`, Adapter `audit`) — `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` (`SPEC-231` Nr. 4 und 7).
- Rotation, Lesen und Export der Datei — kein Abnahmeumfang (`SPEC-235`, Entscheidung).

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

- [ ] Verdrahtung: eine Instanz von `audit.New` über `<DATA_DIR>/audit.jsonl` in allen vier Diensten, in beiden Modi, wo der Dienst besteht (`SPEC-139` Nr. 3, 4; Zeilen V01, V04).
- [ ] Datei im Datenverzeichnis: nach einer Löschung steht der Eintrag in der Ablage, ein zweites `Build` hängt an und ändert Vorhandenes nicht, ohne Löschung entsteht die Datei nicht (`SPEC-235` Nr. 2, 7; Zeilen V01, V02, V05).
- [ ] Kein Geheimnis: ein Canary-Secret (Token, Userinfo der `UpstreamURL`, Fehlertext) erscheint in keinem Byte der Datei (`LH-QA-04`; Zeile V03).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: die Datei `audit.jsonl` ist für den Anwender sichtbar, aber kein gebundenes Beispiel gegen das gebaute Binary kann sie erzeugen (das Binary kennt `version` und `start`; ein Admin-Endpunkt fehlt); das Kapitel folgt mit `slice-admin-sessions`. In §7 steht, was das Binary nach der Verdrahtung tut.
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
| `src/bootstrap/` | update | Verdrahtung: eine Instanz des Adapters in allen vier Diensten (`SPEC-139` Nr. 3, 4) |
| `src/hexagon/`, `src/adapters/driven/audit/` | entfällt | Port, Typen, Dienste und Adapter liefern `slice-audit-log` und `slice-audit-dienste` |
| `tools/schema/schema.yaml`, `src/adapters/driven/sqlite/migrations/` | entfällt | die Ablage ist eine Datei (`SPEC-235`): kein Schema, keine Migration |
| `spec/spezifikation.md` | update (Architect, getan) | `SPEC-139` Nr. 2 bis 4, `SPEC-077` |
| Tests | update, neu | Tests des Roots gegen das `Handle`; die Datei gelesen und geparst |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation vor dem ersten Code-Commit (`AGENTS.md` §3.12).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3 a; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Zeilen V01 bis V05 sind die des Architect aus dem ungeschnittenen Plan von `slice-audit-log`, unverändert.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| V01 | Über das `Handle` stehen die Einträge in `<DATA_DIR>/audit.jsonl`: im Modus `record` alle vier Arten, im Modus `replay` Aufbewahrung und Workspace | Schleife über beide Modi und ihre Arten: Löschung über `Sessions()`, `Workspaces()`, fehlgeschlagener Import über `Import()`, Aufbewahrung mit `RetentionMaxAge` beim zweiten `Build`; die Datei gelesen und geparst | Wartung ohne `Audit` verdrahtet | andere |
| V02 | Neustart: ein zweites `Build` auf demselben `DATA_DIR` findet die Zeilen unverändert und hängt an | zwei `Build` nacheinander, Datei vor und nach | Datei beim `Build` abgeschnitten | andere |
| V03 | Ein Canary (Token, Userinfo der `UpstreamURL`, Fehlertext) steht in keinem Byte der Datei | Datei nach allen Canary durchsucht, nach allen vier Arten | Wert aus der Konfiguration als Feld | andere |
| V04 | Alle vier Dienste erhalten dieselbe Instanz des Adapters | Haken (`NeueWartung`, `NeuerImport`, `NeueWorkspaces`, Session-Dienst): `Audit` ist je Konfig dieselbe Instanz | zweite Instanz für den Workspace-Dienst | andere |
| V05 | Ohne Löschung entsteht die Datei nicht, auch nicht beim Hochfahren und Herunterfahren | `DATA_DIR` nach `Hochfahren` und `Herunterfahren` ohne Löschung | Datei in `Build` angelegt | andere |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Art, jeder Modus, jeder der vier Dienste), führt schon der Handoff die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-audit-dienste` und `slice-bootstrap-m2` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-audit-log` (`welle-betrieb`): der Adapter `audit.Datei` besteht erst dort (*Code-Abhängigkeit*, mittelbar über `slice-audit-dienste`).
- `slice-audit-dienste` (`welle-betrieb`): die Felder `Audit` der vier Konfigs bestehen erst dort; ohne sie ließe sich die Instanz nicht übergeben (*Code-Abhängigkeit*).
- `slice-bootstrap-m2` (`welle-record-replay`): `slice-bootstrap-m2` ändert `src/bootstrap/` für die Verdrahtung der Dienste und der Listener; zwei Slices an denselben Dateien laufen nacheinander (Übergangsregeln, Welle-Datei §5). Läuft dieser Slice nach ihm, ergänzt er dessen Verdrahtung um die eine Instanz des Audit.
- Handbuch-Punkt der DoD: siehe §6, Adresse Planner.

**Start-Ausnahme je Slice** (Linie D, `welle-betrieb` §2): wie `slice-audit-log`; `welle-mock` ist nicht Bedingung, M2 ist es über `slice-bootstrap-m2`.

**Größenmarke** (`SPEC-128`; Faktoren allein dort): Der Block mit `basis`, `schaetzung` und `marke` entsteht beim Start, weil `basis` der Commit ist, der den Stand nach `slice-audit-dienste` und `slice-bootstrap-m2` abschließt; bis dahin steht hier die Rechnung als Prosa. Pfad: `src/bootstrap`. Erwarteter Produktcode (Architect): 40. Fakes: 0. Zeilen: 5, alle andere (keine nebenläufige). Nach `SPEC-128` Nr. 1: *S* 440, *M* 600. `basis` muss hinter dem letzten Commit von `slice-bootstrap-m2` liegen, damit dessen Zeilen unter `src/bootstrap` nicht in die Messung dieses Slice fallen (`SPEC-128` Nr. 10 g). Der Slice liefert keinen Leser fremder Bytes.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke ohne gedeckten Eintrag.
- `in-progress` → `open` (blockiert): `slice-bootstrap-m2` ändert den Aufbau von `Build` so, dass die Verdrahtung des Audit eine Entscheidung des Architect verlangt (Reihenfolge der Instanzen, `SPEC-139` Nr. 3), oder eine Änderung der Gate-Kette (nicht zu erwarten).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor. Mit diesem Slice liegt `slice-audit-log` vollständig geliefert vor (Teile A, E, V): die Welle-Bedingung „jede Löschung und jeder fehlgeschlagene Import schreibt einen Audit-Eintrag“ ist belegt.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- **Handbuch** (DoD): Die Datei `audit.jsonl` im Datenverzeichnis ist eine für den Anwender sichtbare Datei, entsteht aber erst bei einer Löschung, und kein gebundenes Beispiel gegen das gebaute Binary kann sie erzeugen (das Binary kennt `version` und `start`; eine Löschung ohne Admin-Endpunkt löst nur die Aufbewahrung bei vorhandenen Daten aus). Eine Aussage im Handbuch ohne Beispiel verstieße gegen `SPEC-200` Nr. 1 (K1) — **Ausgang:** weiter offen; Adresse: `slice-admin-sessions` (trägt das Kapitel in seiner DoD; Geber: dieser Slice und `slice-audit-log`).
- Reihenfolge der Verdrahtung gegenüber `slice-bootstrap-m2` (beide ändern `src/bootstrap/`) — **Ausgang:** entfallen durch den Start-Trigger (§4): dieser Slice startet nach ihm.
- Ein zweiter `Build` im selben Prozess auf demselben `DATA_DIR` ist nicht unterstützt (`SPEC-106`, `SPEC-139` *Ungeprüft*), V02 fährt zwei `Build` nacheinander nach dem Herunterfahren — **Ausgang:** weiter offen (bei der Closure); der Fänger V02 ruft `Herunterfahren` zwischen den Läufen.
- Telemetrie-Abbildung: `SPEC-231` Nr. 4 und 7 brauchen später `audit_fehlgeschlagen` und den Adapter `audit` — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock`.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster unten steht vor dem ersten Code-Commit, jede Zelle mit Ort oder „gilt nicht“ und Grund; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen (bis zum Review).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Die Operation ist `Build` über `DATA_DIR`; das Argument ist der Pfad des Datenverzeichnisses und der Modus.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch oder Schreibfehler der Ablage | Wiederholung | Geheimnis im Gegenstand |
|---|---|---|---|---|---|---|
| `Build` mit der Instanz des Audit | gilt nicht: `DATA_DIR` hat einen Default (`SPEC-077`), `Build` legt `audit.jsonl` nicht an (`SPEC-139` Nr. 2) | `DATA_DIR` leer oder Verzeichnis fehlt: Fehler der Ablage kommt unverändert zurück (`SPEC-139` Nr. 2) | Modus außer `record` und `replay`: `model.ErrUngueltig` (`SPEC-139` Nr. 2); im Modus `replay` kein Session-Dienst und kein Import (`SPEC-139` Nr. 4) | Schreibfehler der Datei erst mit dem ersten Eintrag, ohne Wirkung auf den Dienst: `SPEC-235` Nr. 5, `SPEC-234` Nr. 5 b | zweites `Build` nach `Herunterfahren` hängt an (`SPEC-235` Nr. 7) | `SPEC-235` Nr. 6, `SPEC-234` Nr. 3 (V03) |

**Zusatztabelle** (`AGENTS.md` §3.16: mehrere Verstöße und Abhängigkeit vom Modus), mit der Spalte *Fänger*:

| Fall | Ort | Fänger |
|---|---|---|
| Modus `record`: alle vier Arten | `SPEC-139` Nr. 4 (Session-Dienst und Import nur im Modus `record`) | V01, Modus `record` |
| Modus `replay`: Aufbewahrung und Workspace, kein Session-Dienst und kein Import | `SPEC-139` Nr. 4 | V01, Modus `replay` |

Die Fenster der Sperre und der Abbruch gehören zu `slice-audit-log`, die Aufrufstellen zu `slice-audit-dienste`.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-06, aus dem Plan von `slice-audit-log` übernommen); der Planner zieht Treffer und Zähler beim Start nach. Dieser Slice hat keinen Fake. Offen und einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) für den Schreibfehler der Ablage; offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner der offenen erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
