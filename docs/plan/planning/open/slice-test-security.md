# Slice slice-test-security: Sicherheitsnachweis mit Canary-Secrets und Netzwerkzugriff

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-QA-04`, `LH-RB-04`, `LH-FA-14`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-QA-04.a` (Canary-Secrets, Netzwerkzugriffstest) · `SPEC-041` (Vertrag von `make test-security`, `harness/vertraege/`) · `LH-FA-13.a` (Token, Transportschutz) · `LH-FA-14.a` (Redaction) · `SPEC-046`, `SPEC-053` (Bedrohungsmodell) · `SPEC-005` (die Endpunkte, über die die Tabelle läuft) · `SPEC-060`, `SPEC-084`, `SPEC-085` (Default-Listen der Redaction) · `SPEC-064` · `ARC-007`.
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

**Ziel:** `make test-security` weist `LH-QA-04.a` nach: Canary-Secrets in Headern, Query-Parametern und Body-Feldern werden über den Record-Pfad aufgezeichnet und in allen Speicher- und Log-Artefakten (Datenbank samt WAL, Blobs, Logs, Audit) nicht gefunden, ebenso nicht der Wert des Admin-Tokens; jeder Admin-Endpunkt außer `GET /healthz` weist ohne und mit falschem Token mit `ODR-E2101` ab; der Admin-Listener ist nur über Loopback oder mit TLS erreichbar. Das Ziel gehört zur Gate-Kette.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Maßnahmen selbst (Redaction, Token, TLS) — `slice-redaction-match-key`, `slice-streaming-redaction`, `slice-capture-pipeline` (alle `done/`), `slice-admin-listener-token`, `slice-admin-tls`; der Slice belegt sie und ändert sie nicht, ein Fund ist ein Befund an den Slice der Maßnahme.
- Die Fallback-Verbote des Hybrid-Modus — `SPEC-041` nennt sie in der Festlegung, aber `welle-admin-api` ist von `welle-hybrid` unabhängig (Welle §5); den Beleg trägt der M4-Test aus `slice-bootstrap-m4`. Der Architect schneidet die Festlegung von `SPEC-041` entsprechend zu (Auftrag 1, §6).
- Ein Geheimnis, das kein Test einspielt, und seine Formen außerhalb der eingespielten Canary-Werte (Kodierungen, Teilstücke) — Grenze von `SPEC-041`; der Vertrag nennt sie unter *Ungeprüft*.
- Ein Zugriff von einem anderen Host — der Prüflauf ist netzlos (`AGENTS.md` §3.1); der Test prüft Start-Prüfung, gebundene Adresse und Handshake (Grenze, §6).

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

- [ ] Canary-Test: je Eintrag der Default-Listen der Redaction (`SPEC-060`, `SPEC-084`, `SPEC-085`) ein Canary-Wert in Header, Query-Parameter und Body-Feld wird aufgezeichnet; die Suche über Datenbank samt WAL und SHM, Blobs, Logs und Audit findet keinen; eine Kontrollprobe belegt, dass die Suche einen Wert findet, den die Redaction nicht entfernt (sonst ist ein leerer Befund wertlos).
- [ ] Netzwerkzugriffstest: für jede geschützte Zeile der Tabelle in `SPEC-005` (Schleife über die registrierten Routen, Zählprüfung gegen die Tabelle) ergibt der Aufruf ohne Token und mit falschem Token `ODR-E2101` ohne Zustandsänderung, `GET /healthz` ohne Token `200`; ein nicht-loopback `ADMIN_LISTEN` ohne TLS und ohne Plaintext-Flag startet nicht, Loopback und TLS sind erreichbar.
- [ ] `make test-security` in der Gate-Kette: Vertrag `SPEC-041` (Architect), die Zeile in `harness/README.md` §Sensors im selben Commit wie das Ziel, ein Fall der Gegenproben, wenn der Vertrag ihn verlangt ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-041`, Sensors-Tabelle in `harness/README.md`).
- [ ] Handbuch: nicht berührt; der Slice ändert das Binary nicht (§7 hält es fest). `slice-handbuch-record-replay` verlangt für Aussagen zum Schutz der Daten einen Test mit Canary-Werten als Beleg; im Repo gibt es heute keinen, ihn liefert dieser Slice (§6).
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
| `tests/` | neu | Canary-Test (Prozess im Modus `record` über den Composition Root aus `slice-bootstrap-m2`, Referenzdienst mit den Canary-Werten) und Netzwerkzugriffstest (Admin-Adapter an einem echten Socket) |
| `harness/vertraege/SPEC-041-test-security.md` | update (Architect) | Vertrag statt Platzhalter: Festlegung, Artefaktmenge, Ausgabe, Exit-Codes, *Ungeprüft* |
| `harness/mk/`, `Dockerfile`, `.dockerignore` | neu, update | Ziel `make test-security` am Ende von `GATE_CHECKS`, Stufe ohne Netz (Muster `pruefung` und `schema.mk`); Form nach dem Vertrag |
| `harness/README.md` | update | Zeile des Ziels in §Sensors mit Bindung und der Satz in *Nicht behauptet*, im selben Commit wie das Ziel (sonst meldet `make docs-check` ein Target ohne Zeile) |
| `tools/gegenproben/` | update, nur wenn der Vertrag es verlangt | ein Fall, der den Test rot färbt, wenn die Redaction fehlt ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

| Zusage | Fänger |
|---|---|
| *(Architect: eine Zeile je Zusage, Randform und Fall des Vertrags)* | *(ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates)* |

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18). Das gilt für die Canary-Werte selbst: ein Canary mit solcher Folge entsteht aus Bytewerten.
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Vertrag von `make test-security`** (`SPEC-041` trägt heute nur einen Platzhalter): Festlegung zugeschnitten auf `LH-QA-04.a` (Redaction, Admin-Bindings, Credential-Leak-Checks), die Fallback-Verbote entweder mit Verweis auf den M4-Test von `slice-bootstrap-m4` oder als eigene Zeile (dann hängt der Slice an `welle-hybrid`, §4), Ausgabe, Exit-Codes, *Ungeprüft*.
2. **Artefaktmenge** („alle Speicher- und Log-Artefakte“): Datenbank samt WAL und SHM, Blob-Verzeichnis, Ausgabe auf stdout und stderr, Logdatei, Ablage des Audits, temporäre Dateien in `DATA_DIR`; byte-genaue Suche; welche Kodierungen des Canary gesucht werden (roh, URL-kodiert, JSON-escaped).
3. **Menge der Orte** als Schleife aus den Default-Listen (`SPEC-060`, `SPEC-084`, `SPEC-085`) und dem Wortlaut von `LH-QA-04.a` (Header, Query-Parameter, Body-Felder), Orakel und Zählprüfung (Schritt 14); die Kontrollprobe, die die Suche selbst fängt.
4. **„Jeder Admin-Endpunkt“:** Quelle der Menge (die Routen des Adapters gegen die Tabelle in `SPEC-005`), Zählprüfung, Pfade mit Platzhaltern.
5. **Netzwerkzugriffstest unter `--network none`** (`AGENTS.md` §3.1): was geprüft wird (gebundene Adresse, Startverweigerung, Handshake) und was nicht (Zugriff von einem anderen Host); als akzeptiertes Negativ mit Grund und Folge im Vertrag.
6. **Ort in der Kette:** Laufzeit des Ziels (Prozess und Image-Bau), Stufe im `Dockerfile`, Gegenprobe nach [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md).
7. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* der Nachweis `LH-QA-04.a` aus `slice-streaming-redaction` und `slice-capture-pipeline` (beide `done/`): deren §1 weist den Canary-Test nicht ausdrücklich zu, Redaction und Capture tragen eigene Tests auf Bausteinebene. *Von diesem Slice an:* der Canary-Nachweis, den `slice-handbuch-record-replay` (`open/`) für seine Aussagen zum Schutz der Daten verlangt; sein Start-Trigger (`slice-bootstrap-m2`) liegt vor diesem Slice und nennt ihn nicht, das ist keine Sendung an ihn, aber ein Hinweis an den Planner von `welle-benutzerhandbuch` (§6, Ausgang *offen*).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token`, `slice-admin-tls`, `slice-admin-sessions`, `slice-admin-imports`, `slice-admin-workspaces`, `slice-bootstrap-m2`, `slice-audit-log` und `slice-telemetry-einbau-gateway` (die beiden letzten `welle-betrieb`) liegen in `done/`. Der Vertrag `SPEC-041` trägt Festlegung und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** der Netzwerkzugriffstest läuft über **jeden** Admin-Endpunkt (`LH-QA-04.a`); die Schleife über die registrierten Routen deckt nur, was existiert, also braucht die Aussage die drei Endpunkt-Slices (*Aussage-Abhängigkeit*; `welle-admin-api` §5 nennt nur Listener, Token und TLS). Listener, Token und TLS sind *Code-Abhängigkeit*. Der Canary-Test fährt den Record-Pfad über den Composition Root (`slice-bootstrap-m2`, *Code-Abhängigkeit*) und sucht in allen Log-Artefakten; Audit-Ablage (`slice-audit-log`) und die Logs des Gateways (`slice-telemetry-einbau-gateway`) müssen bestehen, sonst bliebe ein Artefakt ungesucht (*Aussage-Abhängigkeit*).

**Nicht gebraucht:** `welle-mock`, `slice-bootstrap-m3`, `slice-bootstrap-m6` (der Test baut den Server selbst) und `welle-hybrid`. Der Slice liegt vor `slice-bootstrap-m6` (`welle-admin-api` §5).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): der Test findet ein Geheimnis in einem Artefakt (ein Befund an den Slice der Maßnahme, kein Carveout: ein Geheimnis im Speicher wird nicht mit Frist zugelassen), oder die Artefaktmenge ist nicht entschieden (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): „(Vertrag, offen)“ heißt, der Architect trägt den Ort in `harness/vertraege/` vor dem Übergang `next` → `in-progress` ein; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Canary einspielen (Ort: Header, Query, Body-Feld) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Artefakte durchsuchen (Datenbank, WAL, Blobs, Logs, Audit) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Admin-Endpunkt ohne und mit falschem Token aufrufen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Erreichbarkeit des Admin-Listeners prüfen (Loopback, TLS, Klartext) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Ausgabe und Exit-Code von `make test-security` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- `SPEC-041` nennt in der Festlegung auch „Fallback-Verbote“; `welle-admin-api` ist von `welle-hybrid` unabhängig — **Ausgang:** offen; Adresse: Architect (Auftrag 1): Verweis auf den M4-Test oder eigene Zeile mit Abhängigkeit von `welle-hybrid`.
- Der Prüflauf ist netzlos (`AGENTS.md` §3.1): die Erreichbarkeit von einem anderen Host wird nicht beobachtet, nur Bindung, Start-Prüfung und Handshake — **Ausgang:** weiter offen als Grenze im Vertrag unter *Ungeprüft* (Auftrag 5); Adresse: Architect.
- Ein Geheimnis in einer Form, die kein Canary-Wert trägt (Kodierung, Teilstück), wird nicht gefunden — **Ausgang:** weiter offen als Grenze im Vertrag unter *Ungeprüft* (Auftrag 2); Adresse: Architect.
- Die Artefaktmenge wächst mit `welle-betrieb` (neue Logs, Audit-Ablage): eine Ablage außerhalb der durchsuchten Menge bliebe ungesucht — **Ausgang:** offen bis zur Closure; Fänger: Zählprüfung aller Dateien unter `DATA_DIR` und der Ausgabeströme gegen die durchsuchte Menge.
- `slice-handbuch-record-replay` (`welle-benutzerhandbuch`) verlangt einen Canary-Test als Beleg seiner Aussagen zum Schutz der Daten; im Repo gibt es keinen, und sein Start-Trigger liegt vor diesem Slice — **Ausgang:** offen; Adresse: Planner von `welle-benutzerhandbuch` (die Aussagen entfallen oder das Kapitel wartet auf diesen Slice).
- Laufzeit des Ziels in der Gate-Kette (Prozessstart, Image) und der Re-Evaluierungs-Trigger von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) bei einem neuen Fall der Gegenproben — **Ausgang:** offen; Adresse: User (Entscheidung nach Messung, Auftrag 6).
- `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×): Prozess, Sockets und Referenzdienst laufen unter Last; Kanäle statt Schlaf, Raten mit Zahl der Läufe — **Ausgang:** weiter offen → `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` im Register.
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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×), `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×), `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (1×) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, `.claude/commands/implement-slice.md` Schritt 14; die Orte der Canary-Werte), `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×, `AGENTS.md` §3.18), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (5×, `AGENTS.md` §3.17) und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
