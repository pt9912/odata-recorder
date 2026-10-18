# Lastenheft — OData Recorder

**Version:** 0.7.0  
**Status:** Draft  
**Autor:** Projektentwurf  
**Datum:** 2026-10-18

---

## 1. Zweck und Geltungsbereich

Der OData Recorder ist ein Werkzeug für Softwareentwickler und Testautomatisierung, das die Kommunikation zwischen OData-Clients und bestehenden OData-Diensten aufzeichnet, zu einem späteren Zeitpunkt reproduzierbar wiedergibt und einen zustandsbehafteten OData-Mockdienst auf Grundlage verfügbarer Metadaten und aufgezeichneter Daten bereitstellt. Integrationstests sollen ohne dauerhaft erreichbaren Produktiv- oder Testserver ausführbar sein.

Zum System gehören die Erfassung und Verwaltung von HTTP/OData-Austauschen, ein deterministischer Replay-Betrieb, die Bereitstellung eines lokal veränderbaren OData-Datenbestands, die Verwaltung von Testzuständen sowie eine bedienbare Schnittstelle für Konfiguration und Teststeuerung. Das System darf nicht voraussetzen, dass eine Aufzeichnung einen vollständigen Datenbestand enthält. Unvollständige Abdeckung muss erkennbar bleiben.

Primäres Einsatzszenario ist die lokale Entwicklung und automatisierte CI-Testausführung. Der erste Abnahmeumfang umfasst OData V4 mit JSON und XML-Metadaten; OData V2 ist eine mögliche spätere Erweiterung. Eine vollständige Simulation beliebiger Geschäftslogik wird nicht verlangt.

**Priorisierung:** Anforderungen in diesem Entwurf sind mit **Muss** (erste abnahmefähige Version), **Soll** (geplanter Ausbau) oder **Kann** (optional) gekennzeichnet. Der Status `Draft` bedeutet, dass Umfang und Grenzwerte noch abgestimmt werden können.

## 2. Stakeholder

| Stakeholder | Rolle | Erwartung |
|---|---|---|
| Anwendungsentwickler | Hauptnutzer | OData-Integrationen lokal und offline entwickeln und prüfen |
| Testautomatisierung / QA | Hauptnutzer | Reproduzierbare, isolierte Tests mit definierter Ausgangslage |
| Plattform-/CI-Team | Betreiber | Einfach startbarer, konfigurierbarer und diagnostizierbarer Dienst |
| API-Verantwortliche | Schnittstellengeber | Protokollkonforme Kommunikation und keine unbeabsichtigten Backend-Mutationen |
| Informationssicherheit | Prüfer | Vertrauliche Zugangsdaten und sensible Nutzdaten geschützt behandeln |

## 3. Funktionale Anforderungen

### LH-FA-01 — Transparente Weiterleitung (Muss)

**Beschreibung:** Das System leitet HTTP-Anfragen an einen konfigurierten OData-Upstream weiter und gibt dessen Antworten an den Client zurück, ohne deren fachliche Bedeutung zu verändern.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein erreichbarer Upstream, when ein Client `GET /Products` sendet, then erhält er Status, Nutzdaten und relevante HTTP-Header der Upstream-Antwort.
- **Boundary:** Given eine leere erfolgreiche Antwort, when sie weitergeleitet wird, then bleiben Status und leerer Body erhalten.
- **Negative:** Given ein nicht erreichbarer Upstream, when eine Anfrage weitergeleitet wird, then erhält der Client einen definierten Gateway-Fehler und das Ereignis wird diagnostizierbar protokolliert.

**Out-of-Scope:** Veränderung der Geschäftslogik des Upstreams.

---

### LH-FA-02 — Aufzeichnung von HTTP-Austauschen (Muss)

**Beschreibung:** Das System erfasst pro Austausch HTTP-Methode, Pfad, Query, relevante Header, Request- und Response-Body, Antwortstatus und Reihenfolge in einer identifizierbaren Recording-Session.

**Akzeptanzkriterien:**
- **Happy Path:** Given aktiver Record-Modus, when ein GET- und ein POST-Austausch erfolgreich stattfinden, then sind beide mit Request und Response in korrekter Reihenfolge abrufbar.
- **Boundary:** Given ein Austausch ohne Request- oder Response-Body, when er erfasst wird, then bleibt die leere Nutzlast eindeutig von einer fehlgeschlagenen Erfassung unterscheidbar.
- **Negative:** Given ein Speicherfehler, when eine Aufzeichnung geschrieben werden soll, then wird der Fehler sichtbar gemeldet und kein vollständig gespeicherter Austausch vorgetäuscht.

**Out-of-Scope:** Rekonstruktion nicht beobachteter Backend-Aufrufe.

---

### LH-FA-03 — Persistenz und Wiederverwendung (Muss)

**Beschreibung:** Aufzeichnungen bleiben über Prozessneustarts erhalten und können für spätere Replay- und Mock-Läufe ausgewählt werden.

**Akzeptanzkriterien:**
- **Happy Path:** Given eine gespeicherte Session, when der Dienst neu gestartet wird, then sind ihre Austausche weiterhin vollständig lesbar.
- **Boundary:** Given eine leere Session, when sie geöffnet wird, then wird sie als leer angezeigt, ohne erfundene Austausche.
- **Negative:** Given beschädigte gespeicherte Daten, when sie gelesen werden, then wird ein nachvollziehbarer Integritätsfehler ausgegeben.

**Out-of-Scope:** Replikation über mehrere Recorder-Instanzen.

---

### LH-FA-04 — Deterministisches Replay (Muss)

**Beschreibung:** Das System beantwortet Anfragen aus einer ausdrücklich zugeordneten Recording-Session und einem isolierten Workspace, ohne den Upstream zu kontaktieren. Wiederholte identische Requests werden nach einer dokumentierten, konfigurierbaren Sequenzregel behandelt.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein aufgezeichneter Request mit Response, when derselbe Request im Replay-Modus eingeht, then wird die passende gespeicherte Response mit Status und Nutzdaten geliefert.
- **Boundary:** Given zwei identische aufgezeichnete Requests mit unterschiedlichen Responses und expliziter Workspace-/Session-Zuordnung, when sie im Sequenzmodus zweimal angefragt werden, then werden die Responses in aufgezeichneter Reihenfolge geliefert.
- **Negative:** Given kein passender Austausch oder eine fehlende/mehrdeutige Workspace-/Session-Zuordnung, when Replay angefragt wird, then wird ein eindeutiger Fehler gemeldet und kein Upstream kontaktiert.

**Zusätzliche verbindliche Abnahmeregel:** Given zwei parallele Replay-Sessions desselben Workspaces, when beide dieselbe Recording-Session abspielen, then bleiben ihre Cursor unabhängig; ein Workspace-Reset setzt alle zugehörigen Replay-Cursor auf ihren Anfangszustand.

**Out-of-Scope:** Automatisches Erfinden fehlender Antworten.

---

### LH-FA-05 — Konfigurierbares Request-Matching (Muss)

**Beschreibung:** Das System bietet ein dokumentiertes HTTP-Matching, das semantisch irrelevante Unterschiede in der Reihenfolge von Query-Parametern und deren äquivalenter URL-Kodierung toleriert. Relevante Unterschiede dürfen nicht unbemerkt zusammenfallen.

**Akzeptanzkriterien:**
- **Happy Path:** Given zwei Requests mit gleicher Methode, gleichem Pfad und gleichen Query-Parametern in anderer Reihenfolge, when sie verglichen werden, then gelten sie als passend.
- **Boundary:** Given zwei Requests mit demselben Pfad, aber unterschiedlichen `$filter`-Werten, when sie verglichen werden, then bleiben sie unterscheidbar.
- **Negative:** Given ein syntaktisch ungültiger Request für eine aktivierte Normalisierungsregel, when Matching erfolgt, then wird ein definierter Fehler oder ein dokumentierter konservativer Fallback verwendet, aber keine falsche Gleichheit behauptet.

**Out-of-Scope:** Vollständige semantische Äquivalenz beliebiger OData-Ausdrücke in der ersten Version.

---

### LH-FA-06 — OData-Metadaten bereitstellen (Muss)

**Beschreibung:** Der Recorder kann die Metadaten eines unterstützten OData-V4-Dienstes erfassen und im Mock-Modus bereitstellen. Der Mock nutzt daraus mindestens EntitySets, Schlüssel und Eigenschaftstypen.

**Akzeptanzkriterien:**
- **Happy Path:** Given gültige OData-V4-Metadaten, when der Mock gestartet und `$metadata` abgefragt wird, then wird ein konsistentes Metadatendokument geliefert und das enthaltene EntitySet ist abfragbar.
- **Boundary:** Given ein EntitySet ohne aufgezeichnete Instanzen, when Metadaten geladen werden, then bleibt das EntitySet im Modell vorhanden.
- **Negative:** Given ungültige oder nicht unterstützte Metadaten, when sie geladen werden, then wird ein expliziter Validierungsfehler ausgegeben.

**Out-of-Scope:** Vollständige Interpretation herstellerspezifischer Metadaten-Erweiterungen.

---

### LH-FA-07 — Zustandsbehaftete CRUD-Operationen (Muss)

**Beschreibung:** Der Mockserver ermöglicht das Lesen, Erstellen, Aktualisieren und Löschen von Entitäten in einem lokalen, pro Testlauf isolierten Datenbestand. Er respektiert definierte Schlüssel und grundlegende Typ-/Pflichtfeldregeln.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein gültiges EntitySet, when eine Entität per POST angelegt, per GET gelesen, per PATCH verändert und per DELETE gelöscht wird, then sind die Zustandsübergänge in nachfolgenden GET-Anfragen sichtbar.
- **Boundary:** Given ein nachweislich vollständig erfasstes, leeres EntitySet, when die Collection abgefragt wird, then liefert der Mock eine gültige leere Collection.
- **Negative:** Given ein ungültiger Schlüssel oder ein Verstoß gegen eine erforderliche Eigenschaft, when eine Mutation versucht wird, then wird eine passende OData-Fehlerantwort geliefert und der Zustand bleibt unverändert.

**Out-of-Scope:** Frei programmierbare serverseitige Geschäftsregeln.

---

### LH-FA-08 — OData-Abfrageoptionen (Muss)

**Beschreibung:** Der Mock unterstützt für den vereinbarten OData-V4-Teilumfang `$filter` (Vergleichsoperatoren und logische Verknüpfungen), `$select`, `$orderby`, `$top`, `$skip` und `$count` mit dokumentierter Semantik.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein EntitySet mit mindestens fünf Datensätzen, when eine kombinierte Filter-, Sortier- und Paging-Abfrage erfolgt, then entspricht die Antwort dem erwarteten Teilbestand und der Sortierreihenfolge.
- **Boundary:** Given `$top=0`, when eine Collection abgefragt wird, then werden keine Elemente ausgegeben; ein angeforderter Count bezieht sich auf die gefilterte Menge vor Paging.
- **Negative:** Given ein nicht unterstützter Operator, when die Abfrage ausgeführt wird, then folgt ein expliziter Fehler statt einer ungefilterten Antwort.

**Out-of-Scope:** `$expand`, `$apply`, Delta-Abfragen und vollständige OData-Funktionsbibliothek in der ersten Version.

---

### LH-FA-09 — Isolierte Workspaces und Reset (Muss)

**Beschreibung:** Testläufe können aus derselben Ausgangslage unabhängige Zustände erhalten. Ein Reset stellt den Baseline-Entity-State und sämtliche Replay-Sequenzpositionen des Workspaces auf ihre initialen Werte zurück. Ein Snapshot-Restore stellt ausschließlich den Entity-State wieder her und lässt Replay-Cursor unverändert.

**Akzeptanzkriterien:**
- **Happy Path:** Given zwei aus derselben Baseline erstellte Workspaces, when Workspace A eine Entität verändert, then bleibt Workspace B unverändert.
- **Boundary:** Given ein Workspace ohne Änderungen, when Reset ausgelöst wird, then bleiben Entity-State und Replay-Cursor identisch zur Baseline.
- **Negative:** Given eine unbekannte Workspace-ID, when ein Reset verlangt wird, then erfolgt eine definierte Fehlermeldung und kein anderer Workspace wird verändert.

**Out-of-Scope:** Gemeinsamer schreibbarer Zustand mehrerer voneinander unabhängiger Recorder-Instanzen.

---

### LH-FA-10 — Snapshot-Verwaltung (Muss)

**Beschreibung:** Nutzer können benannte konsistente Zustands-Snapshots erstellen, auswählen und wiederherstellen.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein veränderter Workspace, when ein Snapshot erstellt, weitere Änderungen vorgenommen und der Snapshot wiederhergestellt wird, then entspricht der Entity-State dem Zeitpunkt der Snapshot-Erstellung; Replay-Cursor bleiben auf ihrer aktuellen Position.
- **Boundary:** Given ein leerer Workspace und ein fortgeschrittener Replay-Cursor, when ein Snapshot erstellt und wiederhergestellt wird, then bleibt der Entity-State leer und der Replay-Cursor unverändert.
- **Negative:** Given eine ungültige Snapshot-ID, when Wiederherstellung angefordert wird, then wird kein Teilzustand übernommen.

**Out-of-Scope:** Versionskontrolle mit Merge von Snapshot-Zweigen.

---

### LH-FA-11 — Erkennbarkeit unvollständiger Daten (Muss)

**Beschreibung:** Das System unterscheidet zwischen nachweislich vollständig, teilweise und unbekannt erfassten Datenbeständen. Es darf unvollständige Daten nicht stillschweigend als vollständiges Ergebnis ausgeben, wenn der strikte Abdeckungsmodus aktiviert ist.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein ausdrücklich als vollständig validierter EntitySet-Bestand, when eine unterstützte Collection-Abfrage erfolgt, then wird sie aus dem Bestand beantwortet.
- **Boundary:** Given eine Aufzeichnung mit `$top=10` ohne Vollständigkeitsnachweis, when sie importiert wird, then lautet die Abdeckung `partial` oder `unknown`, nicht `complete`.
- **Negative:** Given ein unvollständiger oder unbekannter Bestand im strikten Modus, when eine nicht sicher beantwortbare Collection-Abfrage erfolgt, then wird vor Query-Ausführung eine eindeutige Coverage-Fehlermeldung geliefert; null beobachtete Entitäten gelten nicht als Nachweis einer leeren Collection.

**Out-of-Scope:** Automatischer Vollständigkeitsbeweis aus beliebigen partiellen Antworten.

---

### LH-FA-12 — Sichere Betriebsarten und Fallback (Muss)

**Beschreibung:** Nutzer können Record, Replay und Mock ausdrücklich auswählen. Ein kombinierter Betrieb darf nur nach expliziten Regeln auf den Upstream zugreifen; schreibende Requests dürfen niemals implizit weitergereicht werden.

**Akzeptanzkriterien:**
- **Happy Path:** Given Replay- oder Mock-Modus ohne Upstream-Fallback, when Requests eingehen, then entstehen keine Upstream-Verbindungen.
- **Boundary:** Given ein erlaubter Fallback für GET, when ein lokaler Treffer fehlt, then wird nur der explizit freigegebene Upstream-Zugriff ausgeführt.
- **Negative:** Given ein POST ohne ausdrückliche Freigabe für schreibenden Fallback, when der lokale Modus keine Antwort liefern kann, then wird die Anfrage nicht an den Upstream gesendet.

**Out-of-Scope:** Automatischer produktiver Datenabgleich zwischen Mock und Upstream.

---

### LH-FA-13 — Management und Diagnose (Muss)

**Beschreibung:** Das System bietet eine von den OData-Routen getrennte Verwaltung für Sessions, Workspaces, Snapshots, Coverage und Fehlerdiagnose. Der Betriebsmodus wird beim Prozessstart gewählt; eine Laufzeit-Modusumschaltung gehört nicht zum ersten Abnahmeumfang.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein gestarteter Recorder, when ein Nutzer Sessions und Workspaces über die Verwaltungsschnittstelle abfragt, then erhält er deren Kennungen und Status.
- **Boundary:** Given keine Sessions, when die Übersicht abgefragt wird, then wird eine leere Ergebnisliste ausgegeben.
- **Negative:** Given eine unberechtigte oder fehlerhafte Management-Anfrage, when sie eingeht, then wird sie abgewiesen, ohne Recorder-Zustand zu verändern.

**Out-of-Scope:** Grafische Web-Oberfläche in der ersten Version.

---

### LH-FA-14 — Schutz sensibler Aufzeichnungsdaten (Muss)

**Beschreibung:** Geheimnisse in Headern und konfigurierbaren Body-Feldern werden vor persistenter Speicherung und vor Diagnoseausgaben gemäß definierten Regeln entfernt oder ersetzt. Live-HTTP-Transparenz und Vollständigkeit des gespeicherten Recordings sind getrennte Zusagen: Ein Capture-Fehler nach Beginn des Response-Streamings ändert die bereits gesendete HTTP-Antwort nicht nachträglich; der Austausch wird als unvollständig markiert. Redigierte Bodies sind nicht zwingend byteidentisch replayfähig.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein Request mit `Authorization` und `Cookie`, when er aufgezeichnet wird, then sind diese Werte weder in der Recording-Datei noch im regulären Log im Klartext vorhanden.
- **Boundary:** Given ein Body ohne sensible Felder, when er erfasst wird, then bleibt seine für Replay benötigte fachliche Nutzlast erhalten.
- **Negative:** Given eine fehlgeschlagene erforderliche Redaktionsregel oder einen Speicherfehler nach Beginn der Client-Antwort, when persistiert werden soll, then werden keine ungeschützten Bytes abgelegt, der Austausch als unvollständig markiert und der Fehler diagnostiziert, ohne den bereits gesendeten Status rückwirkend zu ändern.

**Out-of-Scope:** Automatische zuverlässige Erkennung sämtlicher personenbezogener Daten ohne Konfiguration.

---

### LH-FA-15 — Konkurrierende Änderungen und ETags (Soll)

**Beschreibung:** Der Mock kann Versionskennzeichen bereitstellen und bedingte Änderungen anhand von `If-Match` prüfen.

**Akzeptanzkriterien:**
- **Happy Path:** Given eine Entität mit ETag, when PATCH mit aktuellem `If-Match` erfolgt, then wird die Änderung übernommen und ein neuer Versionsstand sichtbar.
- **Boundary:** Given eine unveränderte Entität, when sie mehrfach gelesen wird, then bleibt ihr ETag stabil.
- **Negative:** Given ein veraltetes ETag, when PATCH erfolgt, then wird die Änderung mit passender Precondition-Fehlermeldung abgewiesen.

**Out-of-Scope:** Verteilte Versionskoordination über mehrere Instanzen.

---

### LH-FA-16 — Erweiterte OData-Funktionen (Soll)

**Beschreibung:** Das System soll Navigationen mit `$expand` und OData-`$batch` einschließlich atomarer Changesets in einem definierten unterstützten Umfang verarbeiten können.

**Akzeptanzkriterien:**
- **Happy Path:** Given eine gültige Navigation und einen gültigen Batch, when die jeweilige Funktion ausgeführt wird, then sind die verknüpften Daten beziehungsweise Einzelergebnisse korrekt sichtbar.
- **Boundary:** Given ein leerer Batch oder eine leere Navigation, when verarbeitet, then entsteht eine gültige leere Antwort gemäß unterstützter OData-Version.
- **Negative:** Given ein fehlerhaftes Changeset, when es ausgeführt wird, then werden keine Teiländerungen daraus übernommen.

**Out-of-Scope:** Alle herstellerspezifischen Batch- und Navigations-Erweiterungen.

---

### LH-FA-17 — Weitere OData-Versionen (Kann)

**Beschreibung:** OData V2 soll optional als eigener unterstützter Kompatibilitätsumfang ergänzt werden können.

**Akzeptanzkriterien:**
- **Happy Path:** Given ein als V2 unterstützt deklarierter Dienst, when seine Metadaten und eine Standardabfrage verarbeitet werden, then erfolgt die Verarbeitung gemäß dokumentiertem V2-Umfang.
- **Boundary:** Given ein V2-Dienst mit leeren Ergebnissen, when er abgefragt wird, then bleibt das V2-Antwortformat konsistent.
- **Negative:** Given eine nicht unterstützte V2-Eigenschaft, when sie verwendet wird, then erfolgt ein expliziter Kompatibilitätsfehler.

**Out-of-Scope:** Zusicherung vollständiger Kompatibilität zu allen SAP-spezifischen Erweiterungen.

### LH-FA-18 — Recording-Daten in Mock-Baseline importieren (Muss)

**Beschreibung:** Der Nutzer kann geeignete aufgezeichnete OData-V4-Entity- und Collection-Antworten ausdrücklich in eine neue, unveränderliche Mock-Baseline importieren. Die Übernahme ist anhand der zugehörigen Metadaten typisiert, nachvollziehbar und atomar; Konflikte und unvollständige Daten werden ausgewiesen. Ein Import geschieht nicht stillschweigend während des Recordings.

**Akzeptanzkriterien:**
- **Happy Path:** Given gültige Metadaten und eine aufgezeichnete Collection mit zwei eindeutig geschlüsselten Entitäten, when der Import für diese Session gestartet wird, then enthält die neue Baseline beide Entitäten mit Herkunftsnachweis und ein daraus erzeugter Workspace kann sie lesen.
- **Boundary:** Given eine paginierte oder durch `$top` begrenzte Antwort ohne Vollständigkeitsnachweis, when sie importiert wird, then werden die beobachteten Entitäten übernommen, die Coverage bleibt `partial` oder `unknown` und ein leerer Import gilt nicht als Beweis für eine leere Collection.
- **Negative:** Given ein ungültiger Key, widersprüchliche Payloads für denselben Key oder nicht passende Metadaten, when der Import gestartet wird, then scheitert er nachvollziehbar und die bisherige Baseline bleibt unverändert.

**Zusätzliche verbindliche Abnahmeregeln:**
- **Baseline-gebundene Coverage:** Given zwei unterschiedliche Baselines desselben Dienstes, von denen nur eine nachweislich vollständig ist, when beide abgefragt werden, then kann die vollständige Coverage nicht auf die andere Baseline übertragen werden.
- **Teilprojektion:** Given eine importierte Entity mit `$select`-Teilprojektion, when eine nicht nachweislich vorhandene Property gelesen oder die Entity verändert werden soll, then wird der Vorgang entweder nach expliziter, verlustfreier Regel behandelt oder mit Coverage-Fehler abgewiesen; fehlend und `null` bleiben unterscheidbar.
- **Import-Autorisierung:** Given ein nicht autorisierter Admin-Client, when er einen Import anfordert, then wird der Import ohne Änderung abgewiesen.

**Out-of-Scope:** Automatisches Crawling aller nicht beobachteten Backend-Daten und stillschweigende Bereinigung fachlicher Datenkonflikte.

---

## 4. Nichtfunktionale Anforderungen und Randbedingungen

### LH-QA-01 — Replay-Performance

- **Anforderung:** Bei lokalem Speicher, 2 vCPU, 4 GiB RAM, 10.000 aufgezeichneten Austauschen, Antwort-Bodies bis 64 KiB und 20 Requests/s beträgt die zusätzliche Verarbeitungszeit des Replay-Dienstes p95 höchstens 100 ms. Der Wert gilt für Antworten, die der Dienst allein aus lokalem Speicher bildet (Replay und Mockserver); im Record- und im Pass-through-Betrieb bestimmt der Upstream die Antwortzeit, dort wird keine Latenz zugesagt. Vorläufiger, nicht freigegebener Abnahmezielwert. Fachlicher Owner muss p95 ≤ 100 ms oder einen alternativen Zielwert vor Status Accepted explizit bestätigen; ohne Entscheidung keine fachliche Freigabe.
- **Messmethode:** Reproduzierbarer Lasttest mit 10 Minuten Messdauer nach Warm-up; Messung zwischen Request-Eingang und Beginn der Response-Ausgabe.

### LH-QA-02 — Determinismus

- **Anforderung:** Gleiche Recording-Session, Matching-Konfiguration und initialer Workspace-Zustand liefern bei identischer sequenzieller Request-Folge dieselben Statuscodes, fachlichen Bodies und relevanten Header. Volatile Header sind explizit zu definieren.
- **Messmethode:** 100 identische Testläufe nach Reset; Vergleich normalisierter Antworten.

### LH-QA-03 — Isolation und Konsistenz

- **Anforderung:** Parallele Tests dürfen keine Zustandsänderungen oder Replay-Cursor anderer Workspaces beobachten; abgeschlossene Mutationen und Snapshots bleiben nach regulärem Neustart konsistent.
- **Messmethode:** Paralleltest mit mindestens 20 Workspaces sowie Neustart- und Fehler-Injektionstests.

### LH-QA-04 — Datensicherheit

- **Anforderung:** Konfigurierte Geheimnisse dürfen nicht in persistenten Recordings oder regulären Logs erscheinen; Verwaltungsschnittstellen sind standardmäßig nicht extern ungeschützt erreichbar.
- **Messmethode:** Canary-Secrets in Headern und Body, anschließend Prüfung aller Speicher- und Log-Artefakte; Netzwerkzugriffstest auf Management-Endpunkte.

### LH-QA-05 — Bedienbarkeit und Fehlerdiagnose

- **Anforderung:** Fehlende Replay-Treffer, ungültige Metadaten, nicht unterstützte Query-Optionen und unvollständige Datenabdeckung müssen anhand stabiler Fehlerklassen unterscheidbar sein.
- **Messmethode:** Automatisierte Negativtests prüfen Fehlerklasse, Status und diagnostischen Kontext ohne Geheimnisoffenlegung.

### LH-QA-06 — Portabilität und Reproduzierbarkeit

- **Anforderung:** Der Dienst läuft auf Linux als Container und kann ohne verpflichtende externe Infrastruktur gestartet werden; gleiche Konfiguration und gleiche Eingabedaten führen zu gleichem Testverhalten.
- **Messmethode:** Containerbasierter CI-Test mit frischem Datenverzeichnis und anschließendem Persistenz-/Neustarttest.

### LH-QA-07 — Protokolltreue

- **Anforderung:** Der Record-Modus darf gültige unterstützte OData-Austausche nicht fachlich verändern; unbekannte OData-Erweiterungen dürfen die reine HTTP-Aufzeichnung nicht grundsätzlich verhindern.
- **Messmethode:** Vergleich von Direktzugriff und Proxy-Zugriff anhand eines Referenzdienstes, einschließlich Status, Body und relevanter Header.

### LH-RB-01 — OData-Kompatibilität

- **Art:** Technisch
- **Vorgabe:** Der erste Abnahmeumfang orientiert sich an OData V4.0. Der konkret unterstützte Teilumfang ist verbindlich zu dokumentieren; nicht unterstützte Funktionen dürfen nicht stillschweigend als unterstützt erscheinen.
- **Nachweis:** Konformitätsorientierte Tests mit Referenzanfragen und Negativfällen.

### LH-RB-02 — HTTP-Schnittstelle

- **Art:** Technisch
- **Vorgabe:** Der Recorder kommuniziert mit Clients und Upstream über HTTP(S); Transportverschlüsselung und Zertifikatsprüfung dürfen nicht ohne ausdrückliche Konfiguration abgeschaltet werden.
- **Nachweis:** Integrationstests für HTTP und HTTPS sowie Zertifikatsfehler.

### LH-RB-03 — Lokale Entwicklungsumgebung

- **Art:** Technisch
- **Vorgabe:** Entwicklung und Ausführung müssen unter Linux und in Docker-basierten CI-Umgebungen möglich sein.
- **Nachweis:** Reproduzierbarer Container-Build und automatisierter End-to-End-Test auf Linux.

### LH-RB-04 — Schutz von Daten und Zugangsdaten

- **Art:** Rechtlich / organisatorisch
- **Vorgabe:** Der Einsatz mit personenbezogenen oder vertraulichen Daten muss durch geeignete Zugriffskontrolle, Redaktionsregeln und Aufbewahrungsvorgaben abgesichert werden; die Verantwortung für zulässige Testdaten verbleibt beim Betreiber.
- **Nachweis:** Dokumentierte Schutzmaßnahmen, Konfigurationsprüfung und Test auf Geheimnisoffenlegung.

**Erstabnahme:** Muss-Anforderungen sind abnahmebindend; Soll- und Kann-Anforderungen sind dokumentierter Ausbau und nicht Teil der Erstabnahme. Die QA-Grenzwerte müssen vor Freigabe des Lastenhefts bestätigt werden.

## 5. Globale Out-of-Scope-Punkte

- Kein vollwertiger Ersatz für einen beliebigen produktiven OData-Server.
- Keine Garantie auf fachlich vollständige Antworten aus unvollständigen Aufzeichnungen.
- Keine automatische Ausführung unbekannter Geschäftslogik, Actions oder Functions im ersten Abnahmeumfang.
- Keine vollständige OData-V2-Unterstützung im ersten Abnahmeumfang.
- Keine verteilte Hochverfügbarkeits- oder Multi-Node-Architektur.
- Keine grafische Administrationsoberfläche im ersten Abnahmeumfang.
- Kein impliziter schreibender Fallback auf reale Backend-Systeme.
- Kein verpflichtender Cloud-Dienst oder externes Datenbanksystem.
- Kein transparentes Mitschneiden verschlüsselten Verkehrs ohne explizite Proxy-Konfiguration des Clients.

## 6. Glossar

| Begriff | Bedeutung im Lastenheft |
|---|---|
| OData | Standardisiertes Protokoll für datenorientierte HTTP-Schnittstellen |
| Upstream | Echter OData-Dienst, an den Requests im Record-Modus weitergeleitet werden |
| Recording | Persistierte Folge beobachteter HTTP-Request-/Response-Austausche |
| Session | Identifizierbare Gruppe zusammengehöriger Recordings |
| Replay | Auslieferung zuvor aufgezeichneter Antworten ohne echten Backend-Aufruf |
| Mockserver | Lokal antwortender Dienst, der OData-Abfragen auf eigenem Datenzustand ausführt |
| EntitySet | Über den OData-Service adressierbare Menge von Entitäten eines Typs |
| Metadata | Maschinenlesbare Beschreibung von Entitäten, Schlüsseln, Typen und Beziehungen |
| Workspace | Isolierter Testkontext mit eigenem veränderbarem Zustand und Replay-Position |
| Snapshot | Wiederherstellbarer konsistenter Stand eines Workspace |
| Matching | Zuordnung einer eingehenden Anfrage zu einem aufgezeichneten Austausch |
| Replay-Miss | Kein passender aufgezeichneter Austausch vorhanden |
| Coverage | Aussage darüber, ob ein lokaler Datenbestand vollständig, teilweise oder unbekannt erfasst ist |
| Strict Coverage | Betriebsweise, die nicht sicher beantwortbare Abfragen auf unvollständigem Datenbestand abweist |
| ETag | Versionskennzeichen für bedingte HTTP-Zugriffe |
| Changeset | Gruppe zusammengehöriger Änderungen innerhalb eines OData-Batchs |

## 7. Historie

Ab Status `Accepted` sind inhaltliche Änderungen nur über einen dokumentierten Change Request zulässig. Der Versionssprung und der externe Änderungsverweis werden dann hier eingetragen. Vor `Accepted` sind Änderungen am Entwurf ohne Change Request möglich. Tatsachenberichtigungen ohne Anforderungsänderung sind ausdrücklich als solche zu kennzeichnen.

| Version | Datum | Änderung | Verweis |
|---|---|---|---|
| 0.1.0 | 2026-10-17 | Initialer Entwurf auf Basis der Kursvorlage | — |
| 0.2.0 | 2026-10-17 | Mock-Import, Coverage und Replay präzisiert | — |
| 0.3.0 | 2026-10-17 | Session-Isolation, partielle Entities und sichere Capture-Pfade | — |
| 0.4.0 | 2026-10-17 | Neutrales d-migrate-Schema eingeführt | — |
| 0.5.0 | 2026-10-17 | Integritätsanforderungen an die Persistenz, Freigabekriterien | — |
| 0.6.0 | 2026-10-18 | Entwurf vor Accepted: Tatsachenberichtigung der Historie (technische Lösungsdetails gehören in die Spezifikation); keine Änderung des Anforderungsumfangs | — |
| 0.7.0 | 2026-10-18 | Entwurf vor Accepted: `LH-QA-01` auf Antworten aus lokalem Speicher begrenzt, Upstream-abhängige Betriebsarten ausgenommen; Zielwert und Freigabe durch den Owner unverändert offen | — |
