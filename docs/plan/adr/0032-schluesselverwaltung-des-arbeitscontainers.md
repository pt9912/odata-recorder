# ADR-0032: Der Entwickler erzeugt das Schlüsselpaar, der Arbeitscontainer hält einen stabilen Host-Schlüssel im Zustand und lauscht nur auf der Schleife des Hosts

**Status:** Accepted

**Datum:** 2026-11-24

**Autor:** pt9912

**Bezug:** [`LH-RB-04`](../../../spec/lastenheft.md#lh-rb-04--schutz-von-daten-und-zugangsdaten), [`LH-RB-03`](../../../spec/lastenheft.md#lh-rb-03--lokale-entwicklungsumgebung), [ADR-0004](0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0031](0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md)

**Schärft:** der Vertrag des Zugangs zum Arbeitscontainer (noch nicht angelegt; die Nummer ist die nächste freie im Bereich der Linie C, und dessen Architect vergibt sie)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Ein Klon außerhalb des Containers liefert einen Zweig per SSH in ein Bare-Repo des Containers ein (Ziel der Welle).
Dafür braucht der SSH-Dienst im Container zwei Schlüssel: einen öffentlichen des Entwicklers, der Anmeldungen erlaubt, und einen eigenen Host-Schlüssel, an dem der Entwickler den Container erkennt.
Zwei Fehlerbilder bestimmen die Entscheidung: ein privater Schlüssel, der in Repo, Image oder Build-Kontext gerät (`LH-RB-04`), und ein Host-Schlüssel, der sich bei jedem Start ändert; die Warnung verführt dazu, die Prüfung abzuschalten (`StrictHostKeyChecking=no`), und das ist keine Lösung (Geist von `AGENTS.md` §3.2).
Annahmen: Der Klon des Entwicklers liegt auf `u-boot` selbst (die Arbeitsbäume liegen dort) oder erreicht es per SSH; ein Klon auf einem anderen Rechner geht über einen Tunnel durch `u-boot`, und die Entscheidung hängt nicht davon ab, wo der Klon liegt. Wer Zugriff auf den Docker-Daemon des Hosts hat, sieht den Inhalt des Containers ohnehin (`docker exec`), der öffentliche Schlüssel ist kein Geheimnis. Soll ein Entwickler ohne Tunnel von einem anderen Rechner anmelden können, braucht der Zugang eine Bindung am Netz des Hosts (Option B2): das ist eine eigene Entscheidung.
Die Privilegierung des Containers entscheidet [ADR-0031](0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md); diese ADR gilt für den Zugang, den der Zugang-Slice baut, und baut auf deren Entscheidung auf (keine Bind-Mounts, nur benannte Volumes).

## Entscheidung

Wir halten fest (Entscheidung des User vom 2026-11-24):

1. **Das Schlüsselpaar erzeugt der Entwickler** (Typ `ed25519`, ein eigenes Paar für diesen Zweck). Der private Teil bleibt bei ihm: nie im Container, nie im Repo, nie im Image, nie im Build-Kontext. Kein Werkzeug des Repos erzeugt, speichert oder überträgt einen privaten Schlüssel.
2. **Der öffentliche Teil kommt beim Start als Wert in den Container** (der Inhalt der `.pub`-Datei, vom Werkzeug auf dem Host gelesen und als Umgebungswert übergeben), nicht per Mount, nicht im Image. Er landet beim Start in der Liste der erlaubten Schlüssel. Ersetzen heißt: Stopp und neuer Start mit dem neuen Wert; der Zustand (Bare-Repos, Host-Schlüssel) bleibt.
3. **Der Host-Schlüssel entsteht beim ersten Start des Zustands und liegt im Zustand** (ein benanntes Volume des Containers), nicht im Image. Er überlebt Stopp und Neustart; `make dev-start` und `make dev-probe` geben seinen Fingerabdruck auf dem Host aus, und der Entwickler trägt ihn einmal in seine `known_hosts` ein. Wer den Zustand löscht, bekommt einen neuen Schlüssel und eine erwartete Warnung; das Dokument für Entwickler nennt den Weg ohne Umgehung.
4. **Der Dienst lauscht nur auf der Schleife des Hosts** (`127.0.0.1`, ein Port je Container, ein belegter Port macht den Start rot). Ein Entwickler von einem anderen Rechner geht über einen Tunnel durch `u-boot`; die Bindung am Netz des Hosts (B2) ist nicht entschieden und braucht eine eigene ADR.
5. **Anmeldung nur mit Schlüssel**: kein Passwort, kein Root. Benutzer, Shell oder nur Git-Befehle und die Formen der Schlüsselzeile (leer, mehrere, mit Optionen) entscheidet der Vertrag des Zugangs.

## Verglichene Alternativen

Erzeugung und Ort des privaten Teils:

| Option | Pro | Contra |
|---|---|---|
| **E1 — der Entwickler erzeugt, der private Teil verlässt ihn nie (gewählt)** | kein Geheimnis in einem Repo-Werkzeug; nichts zu übertragen | der Entwickler führt einen Schritt aus (`ssh-keygen`), das Dokument beschreibt ihn |
| E2 — ein Ziel erzeugt das Paar auf dem Host und übergibt den privaten Teil | ein Schritt weniger | ein privater Schlüssel entsteht und liegt auf `u-boot`, der Transport zum Entwickler ist ein neues Geheimnis-Problem |
| E3 — Schlüssel im Image oder im Repo | kein Schritt | verworfen: widerspricht `LH-RB-04` |
| E4 — Passwort statt Schlüssel | kein Schlüsselpaar | verworfen: Passwort im Befehl oder in einer Datei, kein Fänger gegen ein schwaches |

Öffentlicher Teil:

| Option | Pro | Contra |
|---|---|---|
| **Ü1 — Wert beim Start (gewählt)** | kein Mount, nichts im Image; atomar mit dem Start | der Wert steht in `docker inspect` (öffentlicher Schlüssel, kein Geheimnis); mehrere Schlüssel brauchen eine Form (Vertrag) |
| Ü2 — Bind-Mount einer Datei des Hosts, lesend | übliche Form | bricht die Grenze „kein Bind-Mount“ aus [ADR-0031](0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md); eine Ausnahme je Datei wäre eine neue Entscheidung |
| Ü3 — `docker cp` nach dem Start | kein Mount | ein Schritt nach dem Start; zwischen Start und Kopie läuft der Dienst ohne Schlüssel |
| Ü4 — im Image oder Build-Kontext | einfach | der öffentliche Schlüssel eines Entwicklers im Image macht das Image personengebunden; der Kontext trägt dann Schlüsseldateien |

Host-Schlüssel:

| Option | Pro | Contra |
|---|---|---|
| H1 — bei jedem Start neu | nichts zu speichern | Warnung bei jedem Start; der Entwickler lernt, sie zu umgehen |
| **H2 — stabil im Zustand, Fingerabdruck ausgegeben (gewählt)** | eine Warnung nur nach Löschen des Zustands; der Fingerabdruck kommt auf einem Weg, der nicht der SSH-Weg ist | der Schlüssel liegt im Volume (wer den Daemon des Hosts erreicht, liest ihn) |
| H3 — vom Entwickler vorgegeben | feste Identität über Zustände | ein privater Schlüssel auf dem Weg in den Container (wie E2) |
| H4 — eine Zertifizierungsstelle | keine Einzel-Einträge | Infrastruktur über Docker und `make` hinaus |

Erreichbarkeit:

| Option | Pro | Contra |
|---|---|---|
| **B1 — Schleife des Hosts (gewählt)** | kein Dienst von außen erreichbar | Entwickler auf anderen Rechnern brauchen einen Tunnel |
| B2 — Netz des Hosts | direkte Anmeldung von anderen Rechnern | ein Dienst mit Schlüssel-Anmeldung im LAN; Firewall und Port am Host, außerhalb von Docker und `make` |
| B3 — kein SSH, Einlieferung per `docker exec` als Transport | kein Dienst, kein Schlüssel, kein Port | jeder Einlieferer braucht den Docker-Zugang zum Host (dem Root gleich); verfehlt das Ziel der Welle (SSH) und taugt nicht für andere Rechner |

## Konsequenzen

- Positiv: Kein privater Schlüssel im Repo, im Image oder im Kontext (ein Fänger kann den Kontext und die Schichten durchsuchen); die Host-Warnung tritt nur auf, wo sich die Identität ändert; der Container bleibt ersetzbar, der Zustand trägt Repos und Host-Schlüssel.
- Negativ: Der Entwickler führt einen einmaligen Schritt aus (Paar erzeugen, Fingerabdruck eintragen). Ein Schlüssel im Volume ist für jeden lesbar, der den Daemon des Hosts erreicht. Der öffentliche Schlüssel steht in `docker inspect`.
- Folgepflicht: Der Zugang-Slice schreibt den Vertrag und verlangt `Accepted` dieser ADR vor dem Code; die Probe des Arbeitscontainers bekommt die Bedingung „SSH-Dienst nimmt Verbindungen an“; das Dokument für Entwickler nennt Erzeugung, Eintrag des Fingerabdrucks und den Weg nach einem Zustandsverlust.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `tools/gegenproben` | in Repo, Build-Kontext und Schichten des Images steht kein privater Schlüssel (Suche nach dem Kopf einer Schlüsseldatei); die Konfiguration des Dienstes erlaubt nur Schlüssel und kein Root | `make gegenproben` |

## Entscheidungen des Users (2026-11-24)

Der User hat die Empfehlungen des Architect angenommen: B1 (die Schleife des Hosts, ein Port je Container; ein Entwickler von einem anderen Rechner geht über einen Tunnel durch `u-boot`, B2 wäre eine eigene ADR) und E1 mit Ü1 und H2 (der Entwickler erzeugt das Paar, der öffentliche Teil kommt beim Start als Wert in den Container, ohne Mount und ohne Image; der Host-Schlüssel liegt stabil im Zustand, sein Fingerabdruck wird auf dem Host ausgegeben, keine Umgehung).

1. **Ort des Klons:** nicht ausschlaggebend; die Schleife genügt für einen Klon auf `u-boot`, jeder andere geht über den Tunnel.
2. **Wert beim Start** (Ü1), nicht eine Datei per Mount: keine Ausnahme von [ADR-0031](0031-arbeitscontainer-mit-eigenem-daemon-privilegiert-unter-festen-limits.md).
3. **Host-Schlüssel stabil im Zustand** (H2) mit Fingerabdruck auf dem Host, ohne Umgehung der Prüfung des Host-Schlüssels.
4. **Ein Entwickler oder mehrere** je Container: die ADR legt keine Zahl fest; die Form der Schlüsselzeile (leer, mehrere, mit Optionen) entscheidet der Vertrag des Zugangs (Entscheidung 5). Mehr als ein Entwickler im Betrieb ist ein Re-Evaluierungs-Trigger.

## Re-Evaluierungs-Trigger

Wenn Entwickler von anderen Rechnern ohne Tunnel arbeiten sollen (Option B2); wenn der Schlüssel eines Entwicklers im Betrieb oft wechselt; wenn der Container mehr als einen Entwickler bedient.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-23 | Proposed | Entwurf des Architect; die Entscheidung steht beim User aus |
| 2026-11-24 | Accepted | Entscheidung des User vom 2026-11-24 (Antwort „ok“ auf die Empfehlungen: B1 und E1 mit Ü1 und H2); die Frage nach dem Ort des Klons ist nicht mehr ausschlaggebend |

Nach `Accepted` wird diese Datei **nicht mehr inhaltlich überschrieben**.
Spätere Korrekturen oder Schärfungen entstehen als neue ADR mit
`Supersedes ADR-NNNN` (Baseline-Regelwerk `modul-04-adrs.md`
§Hard Rule für Accepted-ADRs).
