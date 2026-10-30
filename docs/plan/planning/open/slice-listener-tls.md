# Slice slice-listener-tls: TLS des OData-Listeners (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`) als wiederverwendbarer Baustein

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-RB-02`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-RB-02.a` (HTTP/TLS-Konfiguration, Nachweis: Integrationstests für HTTP, HTTPS und Zertifikatsfehler) · `SPEC-088` (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`) · `SPEC-089` (Admin, Gegenstück) · `SPEC-110` Nr. 2 (heute: `LISTEN_TLS_*` gesetzt ist `ODR-E2001`) · `SPEC-110` (Prüfung der Startkonfiguration) · `SPEC-059` (Herunterfahren) · `ARC-005`, `ARC-012`.
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

**Ziel:** Der OData-Listener spricht HTTPS, wenn `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` beide gesetzt sind, und HTTP, wenn keiner gesetzt ist; ein Start mit genau einem der beiden Schlüssel oder mit unlesbarem, unpassendem oder abgelaufenem Zertifikat scheitert mit `ODR-E2001` vor dem Öffnen des Listeners; die Startkonfiguration nimmt `LISTEN_TLS_*` an (der Nachzug von `SPEC-110` Nr. 2 durch den Architect). Die TLS-Konfiguration steckt in einem Baustein, den der Admin-Listener (`welle-admin-api`) mit `ADMIN_TLS_*` wiederverwendet, ohne ihn zu ändern. Der HTTPS-Test am Listener zeigt: gültiges Zertifikat wird bedient, ein Client mit unbekannter CA lehnt ab (`LH-RB-02.a`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- TLS des Admin-Listeners, `ADMIN_TLS_*`, `ADMIN_ALLOW_PLAINTEXT` und die Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` — `welle-admin-api` (`slice-admin-tls` der Tabelle dort); dieser Slice liefert nur den Baustein. Der Nehmer hat noch keine Datei (siehe §6).
- TLS zum Upstream (`UPSTREAM_TLS_VERIFY`, `UPSTREAM_CA_FILE`) — Bestand, `slice-upstream-tls` in `done/`; Gegenstand ist die Clientseite, hier die Serverseite.
- Neuladen von Zertifikat oder Schlüssel im Betrieb, Client-Zertifikate (mTLS), Zertifikatsverwaltung — kein Anker im Lastenheft (`LH-RB-02` verlangt HTTPS und Zertifikatsprüfung, nicht Betriebsverfahren); der Start liest die Dateien einmal.
- Die Zusage „Transportverschlüsselung und Zertifikatsprüfung dürfen nicht ohne ausdrückliche Konfiguration abgeschaltet werden“ für den Upstream — Bestand (`UPSTREAM_TLS_INSECURE_SKIP_VERIFY`, `slice-upstream-tls`).

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

- [ ] Baustein und Listener: der Listener bedient HTTPS mit dem Zertifikat aus `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY`; ein Client mit unbekannter CA lehnt den Handshake ab, ein Client mit der CA erhält die Antwort; ohne beide Schlüssel bleibt es HTTP (`LH-RB-02`, `LH-RB-02.a`).
- [ ] Startkonfiguration: genau einer der beiden Schlüssel, ein leerer Wert, eine unlesbare Datei, ein Zertifikat ohne passenden Schlüssel, ein abgelaufenes oder nicht zu parsendes Zertifikat endet mit `ODR-E2001` und Meldung nach `SPEC-110`, bevor ein Listener öffnet; `SPEC-110` Nr. 2 lehnt `LISTEN_TLS_*` nicht mehr ab (Nachzug des Architect), `ADMIN_*` bleibt abgelehnt.
- [ ] Der Baustein ist ohne Änderung für den Admin-Listener nutzbar (eigene Schlüsselnamen als Parameter, kein Bezug auf `LISTEN_`); ein Test belegt die zweite Verwendung mit anderen Schlüsselnamen.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Die Schlüssel `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` kommen in den Katalog der Konfigurationsschlüssel, ein Start mit nur einem trägt `ODR-E2001`.
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
| `src/adapters/driving/httpodata/` (oder ein Paket daneben, Ort nach Auftrag 1) | update, neu | Baustein: Laden von Zertifikat und Schlüssel, `tls.Config`, Listener mit TLS |
| `src/bootstrap/` (`konfig*.go`, Verdrahtung) | update | Schlüsselpaar prüfen, Meldung `ODR-E2001`, Übergabe an den Listener |
| `spec/spezifikation.md` | update | `SPEC-110` Nr. 2 und Prüfung der Schlüssel, Vertrag des Bausteins mit Randformen (Architect) |
| `.a-check.yml` | update, falls nötig | der Baustein unter den Regeln des Hexagons |
| Benutzerhandbuch (Ort und Form legt `slice-handbuch-geruest` fest) | update | Kapitel und Katalog (Pflicht Handbuch) |
| Tests | neu | Zertifikate und Schlüssel als Fixtures zur Laufzeit erzeugt (kein Geheimnis im Repo), HTTPS-Test |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Aufträge an den Architect:**

1. **Ort und Schnitt des Bausteins.** In welchem Paket der Baustein liegt, damit Admin- und OData-Listener ihn nutzen, ohne dass ein Driving Adapter den anderen importiert (`ARC-005`–`ARC-007`, `.a-check.yml`); Parameter (Pfade und Namen der Schlüssel für die Meldung).
2. **Prüfung beim Start.** Was „gültig“ heißt: Zertifikatskette im PEM, Schlüssel passt (`tls.X509KeyPair`), Ablauf, Mindestversion von TLS, Cipher-Suiten nach Vorgabe der Standardbibliothek oder eigener Liste; Dateien sind Betreibereingaben (Größe, Rechte des Schlüssels analog zur Token-Datei in `LH-FA-13.a`?); Meldung je Verstoß ohne Inhalt der Schlüsseldatei.
3. **Nachzug `SPEC-110`.** `LISTEN_TLS_*` aus der Liste der abgelehnten Schlüssel nehmen, Reihenfolge der Prüfung (`ODR-E2001` vor dem Öffnen), Zusammenspiel mit `PUBLIC_BASE_URL` und `LISTEN` (Schema `https` in Links, falls `PUBLIC_BASE_URL` abgeleitet wird; `SPEC-055`).
4. **Protokolle.** HTTP/1.1 und HTTP/2 über TLS (`SPEC-033`: „HTTP/2 nach Transportfähigkeit“); ob ALPN HTTP/2 anbietet und was der Adapter an Zusagen für Ströme (`slice-http-adapter`) dann hält.
5. **Herunterfahren.** Das zweistufige Herunterfahren des Listeners (`slice-http-adapter`) gilt mit TLS unverändert; Verhalten bei Handshake, der beim Herunterfahren läuft.
6. **Fixtures.** Zertifikate und Schlüssel entstehen im Test (Hierarchie, Ablauf, falscher Hostname); keine Datei mit Schlüssel im Repo (`LH-RB-04`).
7. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** für die Marke des Planners.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger (Test, Fall der Gegenprobe, Ausgabe des Gates) | Mutation |
|---|---|---|---|
| — | (Architect, Auftrag 7) | — | — |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei. PEM-Fixtures mit Zeilenenden, Leerraum und Fremdtext vor und nach dem Block entstehen aus Bytewerten.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Fehlform von Zertifikat und Schlüssel, jede Kombination der beiden Schlüssel gesetzt/leer/fehlend), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-http-adapter`, `slice-bootstrap-m2` und `slice-handbuch-meldungskatalog` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen und den Nachzug von `SPEC-110`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-http-adapter` (`welle-record-replay`): der Listener, dessen Transport der Slice ändert; `src/adapters/driving/httpodata` besteht heute nicht.
- `slice-bootstrap-m2` (`welle-record-replay`): der Composition Root, der Konfiguration lädt und den Listener öffnet. Der Slice ändert die Startkonfiguration und verdrahtet den Baustein; ohne `LadeOderMelde` im Startbefehl (heute ruft `Run` es nicht) gibt es keinen Ort für die Meldung `ODR-E2001` am Binary und keinen Prozess für den HTTPS-Test.
- `slice-handbuch-meldungskatalog` (`welle-benutzerhandbuch`, setzt `slice-handbuch-geruest` voraus): die DoD verlangt das Handbuch-Kapitel und ein grünes Kataloge-Gate; zwei neue Schlüssel (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`) und ein Code-Fall brauchen das Gate.

**Unabhängig von `welle-mock` / M3, weil** der Listener modusunabhängig ist und der HTTPS-Test den Modus `record` oder `replay` nutzt; weder `MockUseCase` noch der Modus `mock` werden gebraucht. **Nicht unabhängig von M2:** der Baustein hängt am Listener des HTTP-Adapters und am Start im Composition Root von `slice-bootstrap-m2`. Ein Teilstart mit dem Baustein allein (nach `slice-http-adapter`, ohne Verdrahtung) ist möglich, ergibt aber nicht die DoD und wird hier nicht geplant; schneidet der Architect den Baustein ab, geht die Verdrahtung an einen Folge-Slice und der Planner passt diese Datei an. `slice-container-image`, die Telemetrie und `slice-lasttest` sind nicht Bedingung.

**Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (`SPEC-128`), sobald die Tabelle *Zusage · Fänger* vorliegt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Der Slice liest fremde Bytes (PEM von Betreibern): der Planner wendet dafür die Regel für Leser fremder Bytes aus `AGENTS.md` §3.14 an, soweit der Architect die Prüfung der Dateien selbst schreibt und nicht der Standardbibliothek überlässt.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten (`AGENTS.md` §3.14), die Naht ist dann Baustein (Laden, Prüfen) gegen Verdrahtung und Startkonfiguration.
- `in-progress` → `open` (blockiert): der Nachzug von `SPEC-110` fehlt oder der Baustein lässt sich nur mit einem Import des anderen Driving Adapters teilen (siehe §6).

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

- Der Baustein soll für den Admin-Listener ohne Änderung taugen; ohne Zuschnitt (Parameter statt Konstanten, Ort neben beiden Adaptern) entsteht ein Import eines Driving Adapters im anderen, den `make a-check` abweist — **Ausgang:** offen; Adresse: Architect (Auftrag 1).
- Sendung ohne Nehmer mit Datei: `ADMIN_TLS_*` und die Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` gehören `welle-admin-api` (`slice-admin-tls`), dessen Datei noch nicht besteht — **Ausgang:** offen; Adresse: Planner beim Anlegen der Slices von `welle-admin-api` (`AGENTS.md` §3.13).
- Start mit einer Schlüsseldatei, die Gruppen- oder Weltrechte trägt: die Spezifikation verlangt es für die Token-Datei, nicht für den TLS-Schlüssel — **Ausgang:** offen; Adresse: Architect (Auftrag 2), Randform vor dem Code.
- Geheimnis in Fixtures: ein privater Schlüssel im Repo ist ein Befund gegen `LH-RB-04` — **Ausgang:** offen; Auftrag 6, die Fixtures entstehen zur Laufzeit.
- HTTP/2 über TLS ändert das Verhalten von Strömen und Abbruch gegenüber dem Listener ohne TLS — **Ausgang:** offen; Adresse: Architect (Auftrag 4), Fänger mit beiden Protokollen oder benanntes Negativ.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund):

| Operation | fehlender Wert (nur einer der beiden Schlüssel) | leerer Wert | unbekannter oder fremder Wert (Datei fehlt, kein PEM, Schlüssel passt nicht, abgelaufen) | Abbruch (Herunterfahren während des Handshakes) | Wiederholung (zweiter Start, zweiter Handshake) | Modus (`record`, `replay`, `mock`) |
|---|---|---|---|---|---|---|
| Startkonfiguration prüfen (`ODR-E2001`) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Zertifikat und Schlüssel laden | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Listener mit TLS öffnen und bedienen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Handshake des Clients (gültig, unbekannte CA, falscher Hostname) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Zweite Verwendung des Bausteins mit anderen Schlüsselnamen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice liest fremde Bytes (PEM-Dateien des Betreibers): `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14, `.claude/agents/implementer.md` und `.harness/skills/reviewer.md`) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (5×, verkörpert in `AGENTS.md` §3.17) gelten für „jede Fehlform“ von Zertifikat und Schlüssel; `BEO-REPO/randform-erst-im-review-entschieden` (20×, verkörpert in `AGENTS.md` §3.16) für das Raster. Offen: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für Testhilfen, die Verbindungen und Ströme umhüllen, und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
