# Slice slice-admin-listener-token: Admin-Listener und Token-Authentisierung

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-QA-04`, `LH-RB-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Admin-Authentisierung, Token-Datei, Listener) · `LH-QA-04.a` · `SPEC-005` (Zeile `GET /healthz`, Schlusssatz) · `SPEC-006` (Fehlerobjekt) · `SPEC-008`, `SPEC-056`, `SPEC-064` · `SPEC-109` (Abbildung der Port-Fehler) · `SPEC-124` *Ungeprüft* (Code der HTTP-Antwort auf `model.ErrUngueltig`) · `SPEC-114` (Korrelations-ID) · `ARC-007`, `ARC-021` · `architecture.md` §2.2, §5 (Sicherheit).
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

**Ziel:** Der Driving Adapter `src/adapters/driving/admin` bedient einen eigenen HTTP-Listener (Default `127.0.0.1:8081`), verlangt auf jedem Endpunkt außer `GET /healthz` `Authorization: Bearer <token>` und vergleicht zeitkonstant mit dem Token aus `ADMIN_TOKEN_FILE` (fehlt die Datei, erzeugt der Prozess ein Token mit mindestens 256 Bit und Rechten `0600`; eine ungültige Datei verhindert den Start mit `ODR-E2001`); fehlendes oder falsches Token ergibt `ODR-E2101` (401) ohne Zustandsänderung, mit Audit-Eintrag ohne Tokenwert; jede Antwort trägt Fehlerobjekt und Korrelations-ID, und der Rahmen (Routen, Abbildung der Port-Fehler) steht für die Endpunkt-Slices bereit.

**Übernimmt:** den Code der HTTP-Antwort auf ein formal ungültiges Argument (`model.ErrUngueltig`, etwa leerer `service_id` oder ungültige Form der `id`): `SPEC-109` kennt keinen, `SPEC-124` *Ungeprüft* hält ihn als benannte Spec-Lücke ohne Adresse, bis der Vertrag des Admin-Adapters ihn festlegt (Sendung von `slice-session-use-case`, `AGENTS.md` §3.13). Er gilt für alle Endpunkte und steht hier, weil dieser Slice den Rahmen der Fehlerabbildung liefert.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Endpunkte außer `GET /healthz` — `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces`; der Rahmen nimmt Routen entgegen, die Tests registrieren eine Testroute.
- TLS, `ADMIN_TLS_*`, `ADMIN_ALLOW_PLAINTEXT` und die Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` — `slice-admin-tls`; bis dahin bindet kein Composition Root den Listener (`slice-bootstrap-m6` verdrahtet ihn erst nach TLS), also entsteht keine ungeschützte Erreichbarkeit.
- Das Lesen von `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` aus der Umgebung und das Aufheben ihrer Ablehnung in `SPEC-110` Nr. 2 — `slice-bootstrap-m6`; ein gelesener, aber nirgends verdrahteter Schlüssel wäre eine still wirkungslose Sicherheitseinstellung. Der Adapter nimmt typisierte Werte.
- Der Port `AuditLog` und sein Adapter — `welle-betrieb` (`slice-audit-log`); der Slice ruft den Port.
- Token-Rotation, mehrere Token, Benutzer, Rollen, OIDC — Out-of-Scope der Welle (§6): die Spezifikation kennt ein Token.

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

- [ ] Listener und Rahmen: eigener Listener (Default `127.0.0.1:8081`), getrennt vom OData-Listener; `GET /healthz` ohne Token; jede Antwort mit Fehlerobjekt (`SPEC-006`) und Korrelations-ID aus `IDGenerator`; Abbildung der Port-Fehler nach `SPEC-109` samt dem Code für `model.ErrUngueltig` (Übernahme).
- [ ] Authentisierung: `Authorization: Bearer <token>`, zeitkonstanter Vergleich; fehlendes oder falsches Token ergibt `ODR-E2101` (401) ohne Aufruf des Use Case und ohne Zustandsänderung, mit Audit-Eintrag ohne Tokenwert; Randformen des Headers (leer, anderes Schema, mehrfach, Leerraum) nach Vertrag.
- [ ] Token-Datei: fehlt sie, erzeugt der Prozess ein Token (mindestens 256 Bit, Rechte `0600`, nie in einem Log); eine nicht lesbare Datei, weniger als 32 Zeichen, Gruppen- oder Weltrechte, ein Verzeichnis, keine reguläre Datei nach Auflösung eines Symlinks, mehr als eine Zeile und ein nicht beschreibbares Verzeichnis der fehlenden Datei sind `ODR-E2001`; ein Symlink auf eine reguläre Datei (gemountetes Secret) ist gültig.
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
| `src/adapters/driving/admin/` | neu | Listener, Routing-Rahmen, Authentisierung, Fehlerobjekt, Korrelations-ID, Abbildung der Port-Fehler, Token-Datei (Ort der Token-Datei-Logik: Architect) |
| `spec/spezifikation.md` | update (Architect) | Vertrag des Rahmens: Header-Randformen, Reihenfolge der Prüfungen der Token-Datei, Audit-Eintrag des Abweisens, Code für `model.ErrUngueltig`; Nachzug `LH-FA-13.a`, `SPEC-005`, `SPEC-064` |
| Tests | neu | Adapter gegen einen `httptest`-Server und Fakes von `AuditLog` und Use Cases; Token-Dateien in einem Testverzeichnis mit gesetzten Rechten; der Adapter importiert weder Adapter noch Service (`make a-check`) |

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

1. **Vertrag des Rahmens:** Routen entgegennehmen, Reihenfolge der Schichten (Korrelations-ID vor Authentisierung), Fehlerobjekt nach `SPEC-006` mit dem Korrelationsfeld, das `SPEC-006` nur „zusätzlich“ nennt (Feldname, Header).
2. **Authentisierung:** Randformen von `Authorization` (fehlt, leer, Schema in anderer Schreibweise, `Basic`, mehrere Header, Token mit Leerraum), `GET /healthz` mit falschem Token und mit anderen Methoden, unbekannte Pfade (`401` vor `404`, damit ein Client ohne Token die Routen nicht erfährt, oder umgekehrt).
3. **„Der Vergleich erfolgt zeitkonstant“** (`LH-FA-13.a`): ein Fänger, der aus dem benannten Grund rot wird (Mutation: Vergleich mit `==` oder `bytes.Equal`), mit Haltepunkt an einem Argument, das der Test übergibt, oder ein akzeptiertes Negativ mit Grund und Folge (`AGENTS.md` §3.15; Register-Sichtung der Welle §5).
4. **Token-Datei:** Reihenfolge der Prüfungen bei mehreren Verstößen, Erzeugen (Rechte von Anfang an, atomar, Wettlauf zweier Prozesse), Plattform ohne POSIX-Rechte, Auflösung eines Symlinks, die Grenze „weniger als 32 Zeichen“ nach Entfernen von Leerraum; die Meldung nennt Schlüssel und Grund, nie den Wert (`SPEC-110` Nr. 12 als Muster).
5. **Audit des Abweisens:** Felder (kein Tokenwert, auch kein Teilstück), Verhalten bei vielen Abweisungen, und ob der Adapter jeden Fehlercode dem Port `Telemetry` meldet (Zähler der Diagnose); wenn ja, braucht der Slice `slice-telemetry-port` im Start-Trigger (§4).
6. **Code der HTTP-Antwort auf `model.ErrUngueltig`** (Übernahme), für alle Endpunkte.
7. **Grenzen des Servers:** Lese- und Schreib-Timeouts, größte Header und Bodies, langsame Clients; das Beenden einer Session kann länger dauern als jedes Schreib-Timeout (`SPEC-124` *Ungeprüft*); zweistufiges Herunterfahren des Listeners wie in `slice-http-adapter`.
8. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* der Code für `model.ErrUngueltig` aus `slice-session-use-case` (`done/`, nicht änderbar): angenommen, §1 *Übernimmt* und DoD Punkt 1 tragen ihn mit der Kennung des Gebers. *Von diesem Slice an:* das Lesen der `ADMIN_*`-Schlüssel an `slice-bootstrap-m6` (§1 *Übernimmt* und DoD von `slice-bootstrap-m6` tragen es), TLS an `slice-admin-tls` (§1 dort), die Endpunkte an `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces` (deren §1 nennt den Rahmen als Geber); der Port `AuditLog` ist keine Sendung, sondern eine Abhängigkeit von `slice-audit-log` (§4).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-audit-log` (`welle-betrieb`) liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** `LH-FA-13.a` verlangt für jedes Abweisen mit `ODR-E2101` einen Audit-Eintrag ohne Tokenwert; der Port `AuditLog` liegt heute nicht in `src/hexagon/ports/support/` (*Code-Abhängigkeit*). Meldet der Adapter jeden Fehlercode dem Port `Telemetry` (Auftrag 5), kommt `slice-telemetry-port` hinzu.

**Unabhängig von `welle-mock`, M2 und dem HTTP-Adapter** (`slice-http-adapter`): `ARC-007` ist ein eigener Driving Adapter, Adapter importieren einander nicht (Architektur §2); er braucht `IDGenerator` (`done/`) für die Korrelations-ID und keinen Composition Root, die Tests bauen den Server selbst. Ob `slice-audit-log` vor dem Mock starten darf, entscheidet die Planung von `welle-betrieb`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14. **Vorab benannter Teil für einen Schnitt beim Handoff** (`AGENTS.md` §3.14): die Token-Datei (DoD Punkt 3); sie wird zuletzt in eigenem Commit geschrieben, vorher gemessen, und teilt keinen Mechanismus mit Listener, Rahmen und Authentisierung.
- `in-progress` → `open` (blockiert — Carveout?): `slice-audit-log` liefert den Port nicht in der Form, die der Eintrag des Abweisens braucht, oder die Token-Datei verlangt eine Plattformregel, die der Architect nicht festlegt (siehe §6).

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
| Authentisierung (Header `Authorization`) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Token-Datei laden (Pfad, Inhalt, Rechte) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron beim Start ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| Token-Datei erzeugen (fehlende Datei) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron beim Start ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| `GET /healthz` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Abbildung von Port-Fehlern und `model.ErrUngueltig` auf Status und Fehlerobjekt | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Herunterfahren des Listeners (zweistufig) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- „Der Vergleich erfolgt zeitkonstant“ (`LH-FA-13.a`) hat keinen Fänger, der aus dem benannten Grund rot wird — **Ausgang:** offen; Adresse: Architect (Auftrag 3): Fänger mit Haltepunkt an einem Argument des Tests oder akzeptiertes Negativ mit Grund und Folge im Plan.
- `LH-FA-13.a` nennt keine Begrenzung von Fehlversuchen; ein vom Betreiber gesetztes Token von nur 32 Zeichen wäre durch Raten angreifbar — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User (Owner); kein Code, bis entschieden.
- Die Prüfung der Rechte der Token-Datei gilt nach `LH-FA-13.a` „auf POSIX-Systemen“; das Verhalten auf anderen Systemen und im Test ohne gesetzte Rechte hat keinen Ort (Auftrag 4) — **Ausgang:** offen; Adresse: Architect.
- Der Code der HTTP-Antwort auf `model.ErrUngueltig` (Übernahme) hat keinen Ort in `SPEC-109` (Auftrag 6) — **Ausgang:** offen; Adresse: Architect, vor dem ersten Code-Commit; er blockiert alle drei Endpunkt-Slices.
- `slice-audit-log` (`welle-betrieb`) hat keine Datei; die Form des Ports `AuditLog` (Felder, Fehler des Anhängens, Verhalten bei Fehler des Audits) entscheidet dessen Vertrag — **Ausgang:** offen; Adresse: Planner von `welle-betrieb` und Architect (der Eintrag des Abweisens ist ein Kunde des Ports, `AGENTS.md` §3.13).
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): die Zweige der Authentisierung und der Token-Datei sind viele; jede Zeile des Rasters hat einen Lauf — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×): die Fakes von `AuditLog` und der Use Cases tragen jede Schnittstelle des Originals und zählen jeden Aufruf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: der Rahmen wächst mit Fehlerabbildung, Timeouts und Token-Datei; die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** offen; Adresse: Planner beim Start (§4), mit der Token-Datei als vorab benanntem Teil.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×, Zweige der Authentisierung und der Token-Datei), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes, außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (7×, `AGENTS.md` §3.10; Token-Datei bei zwei Prozessen), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15; der zeitkonstante Vergleich als Negativ) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
