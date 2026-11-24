# ADR-0031: Der Arbeitscontainer läuft privilegiert mit eigenem Docker-Daemon, unter festen Limits, im eigenen Netz und ohne Zugriff auf Daemon und Dateien des Hosts

**Status:** Accepted

**Datum:** 2026-11-24

**Autor:** pt9912

**Bezug:** [`LH-QA-06`](../../../spec/lastenheft.md#lh-qa-06--portabilität-und-reproduzierbarkeit), [`LH-RB-03`](../../../spec/lastenheft.md#lh-rb-03--lokale-entwicklungsumgebung), [ADR-0001](0001-implementierungssprache-go.md), [ADR-0004](0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md)

**Schärft:** [`SPEC-215`](../../../harness/vertraege/SPEC-215-arbeitscontainer.md)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Die Gates bauen Images mit festen Tags im Daemon des Hosts, und die Läufe mehrerer Rollen teilen ihn: Sie stören sich (Tags, Fristen der Gegenproben unter Last), und Mutationsläufe haben den Host mehrfach in den Speichermangel getrieben.
Der Arbeitscontainer soll diese Läufe in einen eigenen Daemon unter festem Speicher- und CPU-Limit holen; er fährt `make gates` und `make gegenproben` (Folgepflicht aus [ADR-0005](0005-gegenproben-ausserhalb-der-gate-kette.md)), und der Neustarttest im Container (`LH-QA-06.a`) braucht später einen Docker-Daemon im Testlauf.
Ein Docker-Daemon im Container braucht Rechte, die ein gewöhnlicher Container nicht hat: mounten, Namensräume und cgroups anlegen, Netzregeln setzen.
Die Entscheidung berührt die Sicherheit des Hosts; der Architect hat sie entworfen, der User hat am 2026-11-24 entschieden (Antwort „ok deine Empfehlungen“ auf die Empfehlungen des Architect, alle fünf Fragen des Entwurfs zu dieser ADR).
Gemessen am Host dieses Entwurfs: Docker 29.10.0, cgroup v2, Speichertreiber `overlayfs`, 33 362 579 456 Byte Speicher; 20 Kerne und 31 GB nennen die Übergangsregeln für `u-boot`.
Annahmen: (1) Im Container läuft Code dieses Repos, seine Abhängigkeiten nach `go.sum`, Images nach Digest und die Befehle der Rollen, kein Code Dritter. (2) Die Rollen erreichen den Daemon des Hosts heute schon: alles läuft über `make` mit `docker`, und ein Zugang zu dessen Socket ist dem Root des Hosts gleichzusetzen; der Container weitet die Rechte der Rollen nicht aus, er weitet aus, was ein Ausbruch aus einem Prüflauf erreichen könnte. (3) Der Host ist Linux mit cgroup v2.
Kippt (1), gilt die Entscheidung nicht (Re-Evaluierungs-Trigger); kippt (3), startet der Container nicht (Vertrag, Randform).

## Entscheidung

Wir führen den Arbeitscontainer als **privilegierten Container mit eigenem Docker-Daemon** (Option A) unter folgenden Grenzen:

1. **Privilegierung.** `--privileged`, aber ohne jede Verbindung zu Daemon und Dateien des Hosts: kein Docker-Socket, kein Bind-Mount (auch nicht lesend), nur benannte Volumes; kein Host-Netz, kein Host-PID- und kein Host-IPC-Namensraum, kein veröffentlichter Port; der innere Daemon hört allein auf dem Unix-Socket im Container, kein TCP-Port.
2. **Limits.** Der Ort ist die Befehlszeile des Außencontainers (seine cgroup trifft jeden Prozess darin, auch BuildKit des inneren Daemons und jeden inneren Container): Speicher 10 GiB ohne Swap, 8 CPUs, 8192 Prozesse. Die Werte sind Vorgaben von Variablen des Werkzeugs; der Wert 0 oder leer ist abgewiesen, ein Limit nicht kleiner als der Speicher des Hosts ebenso. Auf einem Host laufen höchstens zwei Container zugleich (Regel, die kein Werkzeug prüft). Rechnung: zwei Container zugleich (Closure-Trigger der Welle) belegen höchstens 20 von 31 GB und 16 von 20 Kernen. Bei Überschreitung des Speichers wählt der Kernel den Prozess: er beendet Prozesse **im** Container (Exit 137 des betroffenen Prozesses, im inneren Bau ein roter Schritt); der Host bleibt unberührt. Der Container endet nur, wenn der Kernel dessen ersten Prozess wählt; ein Ende der ganzen Gruppe bei jeder Überschreitung (`memory.oom.group`) gehört nicht zur Entscheidung.
3. **Netz.** Ein eigenes Bridge-Netz je Container, Ausgang per NAT, keine veröffentlichten Ports. Die Prüfläufe im inneren Daemon bleiben `--network none`, nur `deps` holt aus dem Netz; der innere Daemon bekommt einen Adressbereich, der den des Außennetzes nicht überlappt.

Die Zahlen sind die Rechnung aus dem Kontext; der Slice misst Zeit bis zur Bereitschaft und Spitze des Speichers auf `u-boot` und trägt sie ein; ändert eine Messung einen Wert, ist das eine Folge-Entscheidung (Re-Evaluierungs-Trigger).
Formen der Werte, Meldungen und Randformen stehen in `SPEC-215`, nicht hier.

**Einordnung nach `AGENTS.md` §3.1.** Der Wortlaut „Host braucht nur Docker und GNU make“ und „Prüfläufe netzlos und ohne schreibbaren Mount“ gilt für Prüfläufe; der Arbeitscontainer ist keiner, und die Prüfläufe in ihm bleiben es (innere Container ohne Netz, ohne Mount). Der Host braucht weiter nur Docker, `bash`, `git` und GNU `make`. Eine Ergänzung der harten Regel ist nicht nötig und nicht Teil dieser Entscheidung; wer sie dennoch will, braucht eine eigene ADR. **Nach §3.6:** Keine Schwelle sinkt, die Gates laufen unverändert.

## Verglichene Alternativen

Privilegierung:

| Option | Pro | Contra |
|---|---|---|
| **A — `--privileged`, Daemon im Container (gewählt)** | der übliche Weg für einen Daemon im Container; nichts auf dem Host zu installieren (Docker-only); cgroup-Limits der Gates (`--memory 2g` im inneren Container) und BuildKit funktionieren; wenig Eigenbau | Root im Container ist dem Root des Hosts nahe (Geräte, Mounts, Kernel-Schnittstellen); die Grenzen oben schützen vor Fehlern und Ausreißern, nicht vor einem gezielten Ausbruch |
| B — gezielte Fähigkeiten statt `--privileged` (`SYS_ADMIN`, `NET_ADMIN` und weitere, Profile für seccomp und AppArmor gelockert, keine Host-Geräte) | weniger Fläche (kein Zugriff auf Geräte des Hosts) | die kleinste Menge an Fähigkeiten ist nicht gemessen, `SYS_ADMIN` allein ist weit; mehr Aufwand ohne belegten Gewinn; als Nachzug nach Messung offen (Trigger) |
| C — Daemon ohne Root im Container (rootless, unprivilegiert) | ein Ausbruch erreicht nur einen unprivilegierten Benutzer des Hosts | ob der Container die Limits der inneren Container setzen darf (cgroup-Delegation), ob BuildKit und Overlay-Dateisystem laufen, ist nicht gemessen; Netz über Hilfsprozess; die Limits der Gates wären gefährdet |
| D — Laufzeit mit eigenem Kernel-Namensraum (etwa Sysbox, gVisor, Kata) | stärkste Trennung ohne `--privileged` | eine Laufzeit auf dem Host installieren und im Daemon eintragen: widerspricht „Host braucht nur Docker und GNU make“; Pflege und Lieferkette auf `u-boot` |
| E — kein Container, eine Maschine (VM, zweiter Rechner) | stärkste Trennung, eigene Limits der Maschine | außerhalb von Repo und `make`; Aufwand auf `u-boot` und Pflege der Maschine; die Welle liefert dann kein Gerüst |
| F — Socket des Host-Daemons einhängen | einfach | derselbe Daemon: Tags und Speicher bleiben geteilt, das Ziel der Welle entfällt; Root des Hosts per Socket |
| G — nichts tun (Sperre `flock`, Daemon des Hosts) | keine neue Fläche | Speichermangel, Tag-Konflikt und Fristen unter Last bleiben (Anlass der Welle) |

Limits (Ort):

| Option | Pro | Contra |
|---|---|---|
| **L1 — Befehlszeile des Außencontainers (gewählt)** | trifft jeden Prozess im Container samt BuildKit; der Host setzt die Grenze, der Container kann sie nicht lockern | ein Limit gilt für alles zusammen, nicht je Lauf im Container |
| L2 — nur im inneren Daemon (Konfiguration, Limits je innerem Container) | feiner | der Daemon selbst und BuildKit bleiben unbegrenzt; der Host sieht nichts |
| L3 — beides | feine Grenzen und Hülle | die inneren Grenzen der Gates (`--memory 2g`) gibt es ohnehin; keine neue Entscheidung nötig |

Netz:

| Option | Pro | Contra |
|---|---|---|
| **N1 — eigenes Bridge-Netz je Container, Ausgang per NAT, keine Ports (gewählt)** | zwei Container sehen einander nicht; kein Dienst des Containers von außen erreichbar; `deps` kommt an Registry und Modulproxy | der Container erreicht das Netz des Hosts und dessen LAN über NAT; eine Sperre bräuchte Regeln am Host (Daemon-Konfiguration, Firewall), außerhalb von Docker und `make` |
| N2 — Standard-Bridge des Hosts | nichts anzulegen | Container auf derselben Bridge sehen einander |
| N3 — kein Netz, Spiegel und Proxy auf dem Host | Container ohne Ausgang | ein Dienst auf dem Host (Registry-Spiegel, Modulproxy) ist Infrastruktur über Docker und `make` hinaus |
| N4 — Netz des Hosts | keine Adressfragen | der Container teilt Ports und Dienste des Hosts; verworfen |

## Konsequenzen

- Positiv: Der Daemon des Hosts trägt vom Lauf nichts (nur Image, Container, Volumes und Netz des Arbeitscontainers); Speicher und CPU der Läufe sind begrenzt; die Gegenproben bekommen ihren ersten Lauf unter festem Limit.
- Negativ: Ein privilegierter Container ist ein Weg zum Root des Hosts, wenn Code darin ihn gezielt sucht; die Annahme (1) trägt die Entscheidung. Das Limit schützt den Host vor Speichermangel, nicht vor einem Ausbruch. Das LAN des Hosts bleibt über NAT erreichbar. Zwei Container können zusammen mehr Speicher zusagen, als der Host hat (die Summe prüft kein Werkzeug; Regel: höchstens zwei).
- Folgepflicht: `SPEC-215` (Vertrag) trägt die Formen; die Zeile des Ziels kommt in `harness/README.md`; ein Lauf auf `u-boot` belegt Bereitschaft, Limit und „Host unberührt“ mit Zahlen. Wer §3.1 ändern will, schreibt eine eigene ADR.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `tools/gegenproben` mit Ersatz für `docker` | jeder Start trägt die Pflichtoptionen (Limits, Namen, Netz) und keine der verbotenen Verbindungen zum Host (Socket, Bind-Mount, Host-Namensräume, Port) | `make gegenproben` |
| Bereitschaftsprobe | der innere Daemon ist ein anderer als der des Hosts, das Limit gilt im Container, kein TCP-Port des inneren Daemons | `make dev-probe` (ein Lauf auf `u-boot`, kein Gate) |

## Entscheidungen des Users (2026-11-24)

Der User hat die Empfehlungen des Architect angenommen; jede Frage des Entwurfs ist entschieden, keine bleibt offen.

1. **Privilegierung:** Option A mit den Grenzen oben. Die Annahme (1) bleibt Annahme dieser Entscheidung: ist Fremdcode im Container denkbar (Pull Requests Dritter), ist D oder E der Weg (Re-Evaluierungs-Trigger).
2. **`AGENTS.md` §3.1:** bleibt unberührt; kein Satz zur Ausnahme „Arbeitscontainer ist kein Prüflauf“.
3. **Werte:** 10 GiB, kein Swap, 8 CPUs, 8192 Prozesse je Container, höchstens zwei Container je Host.
4. **Überschreitung des Speichers:** der Kernel wählt den Prozess; der Host bleibt unberührt, der Lauf wird rot. „Der Container endet“ gilt nur, wenn der Kernel dessen ersten Prozess wählt (ein Ende der Gruppe, `memory.oom.group`, ist nicht entschieden).
5. **Netz:** N1; das LAN des Hosts wird nicht gesperrt (keine Regeln am Host).

Der Zugang von außen (Schlüssel) ist nicht Gegenstand dieser Entscheidung: [ADR-0032](0032-schluesselverwaltung-des-arbeitscontainers.md).

## Re-Evaluierungs-Trigger

Wenn Code Dritter im Container laufen soll (Annahme 1); wenn eine Messung zeigt, dass eine kleinere Menge an Fähigkeiten als `--privileged` den Daemon trägt (Option B); wenn zwei Container auf `u-boot` in den Speichermangel führen (Werte); wenn ein Docker-Release den Daemon im Container ohne `--privileged` trägt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-11-23 | Proposed | Entwurf des Architect; die Entscheidung steht beim User aus |
| 2026-11-24 | Accepted | Entscheidung des User vom 2026-11-24 (Antwort „ok deine Empfehlungen“ auf die Empfehlungen des Architect zu den fünf Fragen); Vorbehalt A von `SPEC-215` entfällt |

Nach `Accepted` wird diese Datei **nicht mehr inhaltlich überschrieben**.
Spätere Korrekturen oder Schärfungen entstehen als neue ADR mit
`Supersedes ADR-NNNN` (Baseline-Regelwerk `modul-04-adrs.md`
§Hard Rule für Accepted-ADRs).
