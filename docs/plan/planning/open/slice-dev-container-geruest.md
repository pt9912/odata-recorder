# Slice slice-dev-container-geruest: Gerüst des Arbeitscontainers

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-QA-06` (gleiches Testverhalten bei gleicher Konfiguration, Container), `LH-RB-03` (Entwicklung unter Linux und in Docker-Umgebungen), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Folgepflicht: der Arbeitscontainer fährt `make gegenproben`), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Vertrag unter `harness/vertraege/`), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Stufen des `Dockerfile`). Eine neue ADR zur Privilegierung kommt vom Architect (Auftrag 1).

**Berührte Spec-Stellen:** ein neuer Vertrag unter `harness/vertraege/` (Nummer vergibt der Architect, Auftrag 3) · `AGENTS.md` §3.1 (Docker-only; Verhältnis siehe Auftrag 1) · `SPEC-102` (Gegenproben, nur falls der Architect das Gerüst dort prüfen lässt).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

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

**Ziel:** Auf dem Host `u-boot` startet ein Container aus einem gepinnten Image mit eigenem Docker-Daemon unter festem Speicher- und CPU-Limit und meldet erst nach einer Bereitschaftsprobe, dass der Daemon in ihm Befehle annimmt; ein Start, der nicht bereit wird, endet nach einer Frist rot.

**Auftrag an den Architect** (Entscheidungen vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Planner entscheidet sie nicht):

1. **Privilegierung** (ADR-Kandidat, sicherheitsrelevant): Ein Docker-Daemon im Container braucht erweiterte Rechte. Der Architect vergleicht die Wege (privilegierter Container, Daemon ohne Root im Container, Laufzeit mit eigenem Kernel-Namensraum, Verzicht auf den Container zugunsten einer Maschine) nach Wirkung auf den Host, Aufwand auf `u-boot` und dem, was `AGENTS.md` §3.1 verlangt (kein lokaler Toolchain-Install; der Host braucht nur Docker und GNU make) — und nennt, ob §3.1 ergänzt werden muss (das wäre eine Änderung der harten Regel mit ADR, nicht stillschweigend). Zu entscheiden ist auch, was der Container auf dem Host sehen darf (Sockets, Mounts, Netz) und was nicht.
2. **Limits**: Wert und Ort von Speicher-, Swap- und CPU-Grenze (Befehlszeile des Außencontainers oder Konfiguration des Daemons), und dass die Grenze auch BuildKit im inneren Daemon trifft, der sich vom Host aus nicht begrenzen lässt; das Verhalten bei Überschreitung (Container endet, Host unberührt) ist eine Zusage mit Fänger.
3. **Vertrag** unter `harness/vertraege/`: Start, Bereitschaftsprobe (was „bereit“ heißt, Frist, Exit-Codes, Wiederholung, Ausgabe), Stopp, wiederholter Start, Abbruch durch Signal. Die Probe ist **erweiterbar**: spätere Slices der Welle fügen Bedingungen hinzu (Zugang), ohne den Vertrag zu ersetzen.
4. **Image und Pins**: Basis per Digest, Versionen der Pakete, Ort der Stufe oder des Skripts (`Dockerfile`-Stufe, `harness/mk/`-Fragment, `tools/harness/`); Bindung in `harness/README.md` (Sensors-Tabelle oder Werkzeuge-Tabelle; ein Ziel, das privilegiert startet, gehört nicht in `GATE_CHECKS` — der Architect entscheidet und begründet es wie [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)).
5. **Zustand**: Was ein Neustart behält (Daemon-Verzeichnis mit Image-Cache für `deps`, Bare-Repos) und was nicht; ein frischer Start ohne Zustand muss möglich sein (Closure-Kriterium der Welle).
6. **Netz**: Die Prüfläufe im inneren Daemon sind netzlos (`--network none`), nur `deps` holt aus dem Netz; welches Netz der Außencontainer bekommt (Spiegel, Proxy, nur Ausgang) ist eine Entscheidung mit Grund.
7. **Raster und Tabelle**: das Raster *Operation × Argumentart* (Operationen Start, Probe, Stopp; Argumente Limit fehlend, leer, null, unbekannt, größer als der Host; Name des Containers schon vergeben; Daemon startet nicht; Abbruch während des Starts; Wiederholung) in §6, jede Zelle mit Ort im Vertrag (`AGENTS.md` §3.16), und die Tabelle *Zusage · Fänger* in §3 (`AGENTS.md` §3.15) — beides vor dem Übergang `next` → `in-progress`.
8. **Prüfbarkeit**: wie ein Fänger ohne privilegierten Start im Gate lebt (Ersatz für `docker` in einer Gegenprobe nach dem Muster der Fälle 103 bis 114 von `SPEC-102`, Fälle in `tools/gegenproben/`) und welcher Teil nur ein Lauf auf `u-boot` belegt; Letzteres steht als akzeptiertes Negativ mit Grund und Folge in §3 und in „Nicht behauptet“ von `harness/README.md`.
9. **Größe**: Schätzung und Marke nach `AGENTS.md` §3.14 (Zeilen der Tabelle, Klassen, Port-Fakes entfallen hier; der Fake ist ein Ersatz-Skript), die der Planner nach dem Vertrag mit dem Werkzeug setzt; wird die Grenze 3000 überschritten, schneidet der Planner vor dem Code.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- SSH-Server, Schlüsselpaar und Host-Schlüssel — `slice-dev-container-zugang`; das Gerüst läuft ohne Zugang von außen, die Schlüsselverwaltung ist eine eigene Entscheidung mit eigener Prüfbarkeit.
- Bare-Repo, Klon und Einlieferung — `slice-dev-container-einlieferung`; sie brauchen den Zugang.
- Lauf von `make gates` und `make gegenproben` im Container — `slice-dev-container-lauf`; dieser Slice beweist, dass der Daemon antwortet, nicht, dass die Gates in ihm bestehen.
- Das Dokument für Entwickler — `slice-dev-container-dokument`; es beschreibt den ganzen Weg und lässt sich erst nach ihm schreiben.
- Ein Eingriff in die vorhandenen Gate-Stufen (`Dockerfile`: `test`, `build`, `race`, `lint`, `gegenproben`) — Bestand: der Container fährt sie, wie sie sind; ein Eingriff verschöbe die Prüfung ihrer Gates.

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

- [ ] Die Entscheidungen 1 bis 6 des Architect stehen vor dem Code in einer ADR (Privilegierung, Netz) und im Vertrag unter `harness/vertraege/`; die ADR hat den Status `Accepted`, bevor der erste Code-Commit entsteht.
- [ ] Start mit Limit und Bereitschaftsprobe nach dem Vertrag: ein frischer Container wird bereit und meldet es; ein Daemon, der nicht bereit wird, endet nach der Frist rot; ein Lauf über dem Speicherlimit beendet den Container, nicht den Host (gemessen auf `u-boot`, Zahl im Bericht).
- [ ] Zu jeder Zusage des Vertrags ein Fänger in der Tabelle *Zusage · Fänger* (§3) und, wo ein Skript oder Gate sie hält, ein Fall der Gegenproben; `make gegenproben` ist gefahren.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Zeile in `harness/README.md` (Werkzeuge oder Sensors, nach Entscheidung des Architect) und, falls die ADR es verlangt, `AGENTS.md` §3.1.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `harness/vertraege/` (neuer Vertrag) | neu | Zusagen des Gerüsts, vom Architect vor dem Code |
| `docs/plan/adr/` (ADR zur Privilegierung und zum Netz) und ADR-Index | neu / update | Entscheidungen 1 und 6, Status `Accepted` vor dem Code |
| Image, Start- und Probe-Skript, Make-Fragment (Orte nach Entscheidung 4) | neu | Gerüst; Ziele außerhalb von `GATE_CHECKS`, falls der Architect es so entscheidet |
| `tools/gegenproben/` und `SPEC-102` | update | Fälle zu den Zusagen, die ein Ersatz-Skript für `docker` fängt (Entscheidung 8) |
| `harness/README.md` | update | Zeile des Ziels und „Nicht behauptet“ für den Teil, den nur `u-boot` belegt |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) entsteht beim Architect **vor** dem Übergang `next` → `in-progress`; sie ist noch leer.

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`, höchstens zwei Container zugleich (Erfahrung vom 2026-10-24: Mutationskopien belegten ein Vielfaches des Bedarfs); Werte mit Escape-Folge oder Rohbytes aus Bytewerten und gegengelesen (`AGENTS.md` §3.18).


## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-fundament` liegt in `done/`; die ADR zur Privilegierung und zum Netz hat den Status `Accepted`; der Vertrag, die Tabelle *Zusage · Fänger* und das Raster aus §6 stehen vor dem ersten Code-Commit (Architect, `AGENTS.md` §3.12). Danach setzt der Planner die Marke.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner setzt sie nach dem Architect auf die Schätzung × 1,25, aufgerundet auf volle Hundert; liegt `slice-groessenmarke-werkzeug` dann in `done/`, mit `make groessenmarke` (Vertrag `SPEC-128`), sonst nach der Form von `AGENTS.md` §3.14. Der Slice liefert Skripte und ein Image, keinen Go-Code; die Metrik zählt die Pfade aus §3, Tests und Fälle der Gegenproben eingeschlossen.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke; offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten) und die Grenze 3000 vor dem Code.
- `in-progress` → `open` (blockiert): die ADR zur Privilegierung wird `Rejected` oder `Superseded`, oder `u-boot` erlaubt den gewählten Weg nicht (Befund an den Planner, die Welle wird umgeplant, Drift-Log).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, je Zusage des Vertrags ist eine Mutation rot gesehen (`AGENTS.md` §3.10), und ein Start auf `u-boot` ist mit Zahlen belegt (Zeit bis zur Bereitschaft, Limit, Speicher beim Lauf); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Prüfbarkeit: Der privilegierte Start läuft nicht in `make gates` (Docker-only, netzlos, ohne schreibbaren Mount); was nur `u-boot` belegt, ist ohne Sensor — **Ausgang:** offen; der Architect benennt Fänger und akzeptiertes Negativ mit Grund und Folge (Auftrag 8), das Negativ steht in „Nicht behauptet“.
- Sicherheit: Ein privilegierter Container mit eigenem Daemon vergrößert die Angriffsfläche des Hosts — **Ausgang:** offen; die ADR zur Privilegierung benennt Wirkung, Alternativen und Begrenzung (Auftrag 1); ohne `Accepted` kein Code.
- Konflikt mit `AGENTS.md` §3.1: der Wortlaut „Host braucht nur Docker und GNU make“ gilt für Prüfläufe; ob der Arbeitscontainer ihn ergänzt oder nicht berührt, entscheidet die ADR — **Ausgang:** offen; Änderung der harten Regel nur durch ADR (`AGENTS.md` §3.6 sinngemäß, Regel nicht still).
- Gemeinsamer Daemon-Zustand: Image-Tags `odata-recorder:*` kollidieren auf dem Host, wenn der Container den Daemon des Hosts nutzt statt eines eigenen — **Ausgang:** offen; der Vertrag sagt zu, dass der innere Daemon allein baut; Fänger: nach dem Lauf trägt der Host-Daemon kein Image des Laufs (Auftrag 7 und 8).
- Nebenläufigkeit zweier Container auf einem Host (Port, Name, Volume): Randform des Rasters — **Ausgang:** offen; Zeile im Raster.
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* als Gerüst; jede Zelle nennt die Frage, der Ort ist der Vertrag des Architect; eine Zelle ohne Ort hält den Implementer an — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Start | Limit nicht gesetzt: Start verweigert oder Standard? (Vertrag, offen) | Limit leer oder `0`: wie „nicht gesetzt“? (Vertrag, offen) | Limit größer als der Speicher des Hosts, nicht parsbarer Wert: (Vertrag, offen) | Signal während des Starts: Aufräumen? (Vertrag, offen) | Start bei laufendem oder angelegtem Container gleichen Namens: (Vertrag, offen) |
| Bereitschaftsprobe | Daemon antwortet nie: rot nach Frist (Vertrag, offen) | Antwort ohne Inhalt: (Vertrag, offen) | Daemon einer fremden Version: (Vertrag, offen) | Abbruch der Probe: Exit-Code (Vertrag, offen) | Probe wiederholt auf bereitem Container: gleiches Ergebnis (Vertrag, offen) |
| Stopp | Container nicht vorhanden: (Vertrag, offen) | gilt nicht: Stopp hat keinen Wert | gilt nicht: Stopp hat nur den Namen | Signal während des Stopps: (Vertrag, offen) | Stopp eines gestoppten Containers: (Vertrag, offen) |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen:

| Paar | Entscheidung | Fänger |
|---|---|---|
| Limit gesetzt und Daemon nicht bereit | offen (Vertrag): welcher Fehler zuerst gemeldet wird | offen |
| Start abgebrochen und Name schon vergeben | offen (Vertrag) | offen |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30; keiner steht bei 3× und offen). Für diesen Slice zählen: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, `offen`) — die Messungen im Container laufen unter einem festen Limit; Raten und Zeiten nennen Limit, Last und Messenden, und erreicht ein Fund mit dieser Welle 3×, braucht der Eintrag einen Ausgang (Welle-Closure); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — die Betriebsregeln stehen in §3, ein Auftreten wird als Beleg eingetragen. Die Erfahrung vom 2026-10-24 (Speichermangel des Hosts durch Mutationskopien, Container ohne Limit) hat **keinen Eintrag im Register**; sie liegt als Notiz des Users vor und ist Anlass dieser Welle, nicht gezählt — die Closure entscheidet, ob ein Eintrag mit Beleg `evidence/slice-dev-container-geruest.md` entsteht.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
