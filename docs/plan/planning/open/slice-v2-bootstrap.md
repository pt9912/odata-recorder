# Slice slice-v2-bootstrap: Wahl des Adapters je Service-Version im Composition Root und Rundlauftest

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-FA-12`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-17.a` (versionierter Adapter, keine Vermischung von V2 und V4) · `LH-FA-12.a` (Betriebsarten) · `SPEC-054` (`SERVICE_ROOT_PATH`, Adressierung je Service) · `SPEC-091` (`SERVICES_FILE`: heute `id`, `service_root`, `upstream_url` je Eintrag, keine Version) · `SPEC-107` Nr. 14 (der Root legt jeden Service mit der Version `4.0` an) · `SPEC-110` (Startkonfiguration) · `SPEC-127` Nr. 1 (Konfiguration des Imports) · `ARC-005`, `ARC-012` · `architecture.md` §2.5 (Modus-Routing), §7 (Erweiterungspunkt OData V2).
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

**Ziel:** Der Composition Root wählt je Service den Adapter seiner OData-Version: die Version steht je Service in der Startkonfiguration (ohne Angabe `4.0`), der Root legt den Service mit dieser Version an, und Import und Mock erhalten Metadatenleser, Payload-Codec und URL-Parser dieser Version; der V4-Pfad bleibt unverändert, und V2 und V4 mischen sich in keinem Service-Modell. Der Rundlauftest der Welle belegt es: er zeichnet einen Referenzdienst mit V2-Antworten auf, importiert die Session, legt einen Workspace an und beantwortet im Modus `mock` die Metadaten und eine Standardabfrage nach dem dokumentierten V2-Umfang; ein leeres Ergebnis bleibt im V2-Format, eine nicht unterstützte V2-Eigenschaft ergibt den Kompatibilitätsfehler, und ein V4-Dienst verhält sich wie vorher.

**Übernimmt (Konfiguration und Anlage des Service):** die Angabe der Version je Service (`SERVICES_FILE` kennt heute keine, `SPEC-091`) und die Anlage mit der konfigurierten statt der festen Version (`SPEC-107` Nr. 14); die Ableitung der Version aus dem Modell trägt `slice-v2-metadaten` (Auftrag 6 dort).

**Übernimmt (Handbuch zu V2):** das Kapitel zu V2 mit Beispiel gegen das gebaute Binary und den Katalogeintrag des Kompatibilitätsfehlers, soweit `slice-v2-metadaten`, `slice-v2-payload` und `slice-v2-query` sie nicht im selben Slice tragen konnten, weil ein V2-Dienst das Binary erst hier erreicht (Sendung von `slice-v2-metadaten`, `slice-v2-payload` und `slice-v2-query`, `AGENTS.md` §3.13).

**Übernimmt (Import je Version):** Die Anpassung des Imports an einen Codec je Version — `baselineimport.New` nimmt genau ein Paar von Codecs, die Auswahl der Antworten (`SPEC-127` Nr. 8) und die Ketten über Folgelinks (Nr. 12) sind V4-Regeln — entscheidet der Architect in der ADR; verlangt sie eine Änderung an Verträgen in `done/`, legt der Planner dafür vor diesem Slice einen eigenen an (§6), und dieser Slice trägt nur die Verdrahtung.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Metadaten-Leser, Payload-Codec und URL-Parser für V2 — `slice-v2-metadaten`, `slice-v2-payload` und `slice-v2-query`; dieser Slice verdrahtet sie und ändert keinen von ihnen.
- Aufzeichnen im Modus `record`: die Aufzeichnung ist von der Version unabhängig (`welle-odata-v2` §1); kein Teil des Record-Pfads ändert sich.
- V2-Mutationen und der Modus `hybrid` für einen V2-Dienst — `welle-odata-v2` §6 bzw. `welle-hybrid`; für einen V2-Dienst im Modus `hybrid` gilt, was der Architect in der ADR festlegt, sonst ein Fehler beim Start.
- Änderungen am V4-Pfad (Mock, Replay, Record) über die Ports hinaus, hinter denen der Adapter steht — `welle-odata-v2` §6.
- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`.

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

- [ ] Wahl des Adapters: Version je Service in der Startkonfiguration (fehlend `4.0`, unbekannt oder im Widerspruch zum gespeicherten Service ein Start-Fehler `ODR-E2001`, wie `SPEC-107` Nr. 14 es für den Root sagt), Anlage des Service mit dieser Version, Metadatenleser, Payload-Codec und URL-Parser je Version in Import und Mock verdrahtet; ein V4-Dienst nutzt unverändert die V4-Adapter.
- [ ] Rundlauftest in `make test`: V2-Referenzdienst aufzeichnen, Session beenden, importieren, Workspace anlegen, im Modus `mock` Metadaten und Standardabfrage beantworten; leeres Ergebnis im V2-Format; nicht unterstützte V2-Eigenschaft ergibt den Kompatibilitätsfehler; die Tests von `welle-mock` laufen unverändert grün.
- [ ] Randformen der Konfiguration und der Wahl (Version fehlt, leer, unbekannt, Widerspruch zum gespeicherten Service, Services verschiedener Versionen nebeneinander, Neustart): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün.
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
| `src/bootstrap/` | update | Konfiguration der Version je Service, Wahl und Verdrahtung der Adapter |
| `src/hexagon/services/` (Import, Mock) | update | Entgegennahme der Adapter je Version, soweit die ADR sie nicht im Root bündelt |
| `src/adapters/driving/httpodata/` | update | Kennzeichnung der V2-Antwort (Header), soweit die ADR sie dem Adapter zuweist |
| `spec/spezifikation.md` | update | Schlüssel der Konfiguration, Anlage des Service, Vertrag mit Randformen (Architect) |
| Tests | neu | Rundlauftest V2 gegen einen Referenzdienst als Test-Server; Fälle der Konfiguration |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Slice liefert einen Test-Server (den Referenzdienst mit V2-Antworten) und Fakes der Adapter: sie zählen im Block (`SPEC-128`) als Port-Fakes und Test-Server; der Rundlauftest ist ein Test mit Prozessen und Fristen, die Betriebsregel zu `docker` und `flock` gilt dafür besonders.

**Auftrag an den Architect** (vor dem ersten Code-Commit; die Aufteilung der Ports je Version steht in der ADR, Status `Accepted`; Neues in der Spezifikation unter einer künftigen Nummer, die der Architect vergibt):

1. Konfiguration: Schlüssel der Version je Service (Eintrag der `SERVICES_FILE`, Einzelschlüssel für den einen Service), Werte, Default, Prüfung (`SPEC-110`); Verhalten bei einer Version, die der Speicher anders kennt (`SPEC-107` Nr. 14: `ODR-E2001`).
2. Wahl: wo der Root die Adapter bündelt (je Service ein Satz Adapter, ein Verteiler hinter den Ports, der nach Version entscheidet) und wie der Mock-Kern und der Import sie erhalten; Zusage, dass kein Adapter einer Version ein Modell der anderen sieht (`LH-FA-17.a`).
3. Import je Version: Auswahl der Antworten und Ketten für V2 (Anpassung von `SPEC-127`), und ob sie ein Slice vor diesem verlangt (Adresse: Planner, `AGENTS.md` §3.13).
4. Modi: welche Modi ein V2-Dienst kennt (`record`, `replay`, `mock`; `hybrid` ausgeschlossen oder nicht) und der Fehler beim Start für den Rest.
5. Kennzeichnung der V2-Antwort im Header (Version, `Content-Type`) und wer sie setzt (Auftrag 4 von `slice-v2-payload`).
6. Der Referenzdienst V2 des Tests: was er liefert (Metadaten, eine Collection, ein leeres Ergebnis, eine nicht unterstützte Eigenschaft) und welche Antworten der Import nach der Auswahl aus Auftrag 3 erkennt.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect mit dem Vertrag, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-v2-query` liegt in `done/` (und damit `slice-v2-payload` und `slice-v2-metadaten`), `welle-mock` liegt in `done/` (Trigger der Welle), und die ADR des Architect zum versionierten Adapter ist `Accepted`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** die drei V2-Slices davor (die Adapter, die der Root wählt), `slice-bootstrap-m2` (der Composition Root, in dem Record, Import und die Anlage der Services verdrahtet sind, `SPEC-107` Nr. 14), `slice-bootstrap-m3` (der Modus `mock` im Root und der M3-Test, der die Grundlage des Rundlaufs ist), `slice-handbuch-meldungskatalog` (Kataloge-Gate für die Handbuch-Pflicht). Der Slice ist **nicht** vor `welle-mock` und nicht ohne M2 startbar.
- **Was der Slice nicht braucht:** `welle-etag`, `welle-expand-batch`, `welle-hybrid`, `welle-betrieb`, `welle-admin-api`; die Reihenfolge nach den Muss-Wellen ist Priorisierung.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, die Produktcode-Erwartung des Architect (Zeile `produktcode` im Block) ist ein fester Anteil; der Slice liest keine fremden Bytes im Sinne eines Formats, aber der Referenzdienst V2 und die Fakes sind Test-Server und Port-Fakes der Schätzung.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt; ein Schnitt vor dem Code läge zwischen Konfiguration und Wahl einerseits und dem Rundlauftest andererseits, sofern sie keinen Mechanismus teilen (sonst steht der Grund in §4).
- `in-progress` → `open` (blockiert — Carveout?): die Anpassung des Imports je Version (Auftrag 3) ist nicht entschieden oder ohne Slice vor diesem nicht zu tragen (siehe §6).

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

- Der Import eines V2-Dienstes braucht mehr als einen Codec je Version: Auswahl der Antworten, Ketten und Folgelinks sind V4-Regeln in `done/` (`SPEC-127` Nr. 8 und 12), und kein Slice der Welle trägt ihre Anpassung — **Ausgang:** offen; Adresse: Architect (Entscheidung in der ADR, Auftrag 3) und Planner (Slice für die Anpassung des Imports vor diesem, `AGENTS.md` §3.13); ohne ihn ist der Rundlauftest der Welle nicht erreichbar.
- Der Rundlauf ist ein Test mit Prozessen, Fristen und einem Referenzdienst; er erbt die Zeitgrenzen von `make test` und `make gegenproben` — **Ausgang:** offen; Zeit gemessen und im Bericht genannt (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
- Die Tests von `welle-mock` müssen unverändert grün laufen (`welle-odata-v2` §3); ein Fänger vergleicht V4-Antworten vor und nach dem Slice über die Ports — **Ausgang:** offen.
- Mehrere Services verschiedener Versionen im selben Prozess dürfen sich nicht beeinflussen (Roots, Metadaten, Caches) — **Ausgang:** offen; Fänger mit einem V2- und einem V4-Dienst nebeneinander.
- Fakes der Ports und Adapter halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Version fehlt | Version leer | unbekannte oder fremde Version | Widerspruch zum gespeicherten Service | Services verschiedener Versionen nebeneinander | Abbruch (Signal im Start), Wiederholung (Neustart) |
|---|---|---|---|---|---|---|
| Prüfung der Startkonfiguration | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Anlage des Service | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Wahl der Adapter in Import und Mock | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Rundlauf Aufzeichnen, Importieren, Abfragen | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

**Zusatztabelle** (je Modus eine Zeile; `AGENTS.md` §3.16): V2-Dienst in `record`, `replay`, `mock` und `hybrid`; V4-Dienst unverändert.

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| (Vertrag, offen) | (Vertrag, offen) | (offen) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für Fakes der Codec-Ports; `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) für den Rundlauftest und die Gegenproben bei wachsendem Produktcode; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig im Sinne einer Zusage über Sperren; der Rundlauftest läuft unter `-race` im Gate (`make test-race`). Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
