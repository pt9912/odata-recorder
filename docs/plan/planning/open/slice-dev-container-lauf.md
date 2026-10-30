# Slice slice-dev-container-lauf: Lauf von make gates und make gegenproben im Container

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-QA-06` (gleiche Konfiguration und Eingabe, gleiches Testverhalten), `LH-QA-03` (nebenläufige Tests, Isolation: zwei Läufe stören sich nicht), `LH-RB-03` (Entwicklung in Docker-Umgebungen), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Folgepflicht: der Arbeitscontainer fährt `make gegenproben`), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (`make test-race` in der Kette), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Vertrag unter `harness/vertraege/`).

**Berührte Spec-Stellen:** `SPEC-102` (Gegenproben: Eingaben, Stempel, Frist je Fall; nur die Stellen, die der Container berührt) · `SPEC-122` (`make test-race`, Zeitgrenze) · ein neuer Vertrag oder Abschnitt für den Lauf im Container (Nummer vergibt der Architect) · die Verträge der Vorgänger (Probe, falls der Lauf eine Bedingung hinzufügt).
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

**Ziel:** Auf einem eingelieferten Stand bestehen `make gates` und `make gegenproben` im Container unter dessen Limit, zwei Container mit verschiedenen Zweigen bestehen nebeneinander, und der Daemon des Hosts legt dabei kein Image und keinen Container des Laufs an.

**Auftrag an den Architect** (Entscheidungen vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Planner entscheidet sie nicht):

1. **Ablauf**: wie ein Lauf ausgelöst wird (Befehl über SSH, Wrapper, Auslöser beim Eingang — Anschluss an Entscheidung 2 der Einlieferung), Ausgabe und Exit-Code zum Entwickler, Ort der Protokolle, was bei abgebrochener Verbindung geschieht (Lauf endet oder läuft weiter), Verhältnis zu `record-gates`, `.harness/stop-gate-streng` und dem Stop-Hook (der Nachweis des Laufs liegt im Arbeitsstand des Containers, nicht im Klon des Entwicklers).
2. **Netz und Cache**: Umsetzung der Netz-Entscheidung des Gerüsts (`deps` holt aus dem Netz, die Prüfläufe sind netzlos); ob Image- und Modul-Cache Läufe überleben, ohne eine Prüfung zu lockern oder zu überspringen.
3. **Zeit und Limit**: Gemessen im Container, einzeln und zu zweit: Zeit je Fall der Gegenproben (Frist je Fall nach `SPEC-102`), Zeitgrenze von `make test-race` (`SPEC-122`), Spitze des Speichers. Welche Messreihe gilt (Zahl der Läufe, Last, Messender; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Ein Fall, der unter dem Limit reißt, ist ein Befund über den Zuschnitt des Falls oder über das Limit, **nie** eine längere Frist (`AGENTS.md` §3.6; Erfahrung aus `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
4. **Stempel**: ob der Container den Stempel `harness/gegenproben.stempel` schreibt oder nur prüft; ob die Prüfsumme im Container für denselben Commit gleich der am Host ist (Dateirechte, Zeilenenden, ungetrackte Dateien, Pins); wohin ein neuer Stempel gelangt (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`). Die Zusage „der Stempel reist mit dem Commit in andere Klone und Container“ aus `SPEC-102` bekommt hier ihren ersten Nachweis im Container.
5. **Host unberührt**: die Zusage, dass der Daemon des Hosts nichts vom Lauf sieht (kein Image `odata-recorder:*`, kein Container, kein Volume), und ihr Fänger (Vergleich der Listen vor und nach dem Lauf auf `u-boot`).
6. **Raster und Tabelle**: Raster *Operation × Argumentart* in §6 (Operation Lauf auslösen; Argumente Zweig nicht eingeliefert, Arbeitsstand schmutzig, Daemon nicht bereit, Limit unterschritten, Abbruch durch Signal oder Verbindung, Wiederholung, zweiter Lauf im selben Container) mit Ort je Zelle, und die Tabelle *Zusage · Fänger* in §3 vor dem Übergang `next` → `in-progress`; Zusagen über „kein Image des Laufs auf dem Host“ als Schleife über Images, Container und Volumes.
7. **Prüfbarkeit**: was ein Fall der Gegenproben mit Ersatz-Skripten fängt (Ablauf, Exit-Codes, Weitergabe der Argumente) und was nur ein Lauf auf `u-boot` belegt (Zeiten, Spitze des Speichers, Parallelität) — letzteres als Messung mit Zahl der Läufe, Last und Messendem im Plan, nicht als „hoch“ oder „stabil“ (`AGENTS.md` §3.10).
8. **Größe**: Schätzung und Marke wie in `slice-dev-container-geruest` Auftrag 9.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung an Gates, Schwellen oder Fristen, damit der Lauf im Container besteht — `AGENTS.md` §3.6; ein Fall, der reißt, ist ein Befund an den Planner (der Zuschnitt des Falls gehört dem Slice, der ihn schrieb, und liegt dann in `done/`: keine Adresse, `AGENTS.md` §3.13).
- Eine CI, die den Container aufruft — Welle §6.
- Ein Coverage-Gate oder der Lasttest zu `LH-QA-01` — Welle `welle-record-replay` §6; der Lauf im Container fährt, was die Gate-Kette hat.
- Das Dokument für Entwickler — `slice-dev-container-dokument`; es beschreibt, was dieser Slice liefert, und kann erst danach stehen.
- Der Ablauf der Rollen (wer wann einliefert und fährt) — Welle §6; der Slice liefert den Lauf, nicht das Protokoll der Rollen.

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

- [ ] Die Entscheidungen 1 bis 5 des Architect stehen vor dem Code im Vertrag unter `harness/vertraege/`.
- [ ] Auf einem eingelieferten Stand bestehen `make gates` und `make gegenproben` im Container; der Daemon des Hosts trägt danach kein Image, keinen Container und kein Volume des Laufs (Listen vor und nach dem Lauf im Bericht); ein Lauf über dem Speicherlimit beendet den Container, nicht den Host.
- [ ] Zwei Container mit verschiedenen Zweigen bestehen zugleich; Zeit je Fall der Gegenproben, Zeit von `make test-race` und Spitze des Speichers stehen mit Zahl der Läufe, Last und Messendem im Plan (§3), neben dem Einzellauf; keine Frist wurde verlängert; zu jeder Zusage ein Fänger in §3.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Vertrag, `harness/README.md` (Werkzeuge, „Nicht behauptet“) und, falls die Folgepflicht aus [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) damit erfüllt ist, `SPEC-102` im selben Commit (`AGENTS.md` §3.9); die ADR bleibt unverändert (`AGENTS.md` §3.5).
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
| `harness/vertraege/` (Vertrag des Laufs im Container) | neu | Auslösen, Ausgabe, Exit-Code, Stempel, Host unberührt |
| Wrapper für den Lauf im Container, Make-Fragment (Orte nach dem Gerüst) | neu | Ablauf nach Entscheidung 1 |
| `tools/gegenproben/` und `SPEC-102` | update | Fälle mit Ersatz-Skripten (Auftrag 7); `SPEC-102` nur dort, wo der Container die Zusagen des Stempels berührt |
| `harness/README.md` | update | Zeile des Ziels; „Nicht behauptet“ für Zeiten, Spitze des Speichers und Parallelität, die nur `u-boot` belegt |
| Messung (Absatz in §3 dieses Plans) | update | Zahl der Läufe, Last, Messender; Einzellauf neben Parallellauf |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) entsteht beim Architect **vor** dem Übergang `next` → `in-progress`; sie ist noch leer. Der Absatz der Messung (Zeiten, Spitze, Raten) entsteht beim Implementer (`AGENTS.md` §3.10).

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`; **keine Messreihe zugleich mit einem Gate-Lauf im Hauptbaum** (Last verfälscht die Zeit je Fall); Werte mit Escape-Folge oder Rohbytes aus Bytewerten und gegengelesen (`AGENTS.md` §3.18).


## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-dev-container-einlieferung` liegt in `done/`; Vertrag, Tabelle *Zusage · Fänger* und Raster aus §6 stehen vor dem ersten Code-Commit (Architect, `AGENTS.md` §3.12). Danach setzt der Planner die Marke.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner setzt sie nach dem Architect auf die Schätzung × 1,25, aufgerundet auf volle Hundert, mit `make groessenmarke` (`SPEC-128`), falls `slice-groessenmarke-werkzeug` dann in `done/` liegt, sonst nach der Form von `AGENTS.md` §3.14.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke; offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten) und die Grenze 3000 vor dem Code.
- `in-progress` → `open` (blockiert): ein Gate oder eine Gegenprobe besteht im Container nicht, obwohl sie am Host besteht, und die Ursache liegt im Gate, nicht im Container (Befund an den Planner; der Slice des Gates liegt in `done/`, keine Adresse, `AGENTS.md` §3.13).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind im Hauptbaum und im Container grün, je Zusage des Vertrags ist eine Mutation rot gesehen (`AGENTS.md` §3.10), die Messung steht mit Zahlen in §3; die Closure-Notiz mit Lerneintrag liegt vor und nennt den Ausgang von `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` und `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` aus dem Lauf im Container.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Eine Frist je Fall reißt im Container (weniger CPU, ein Daemon mehr dazwischen), und die naheliegende Reaktion ist die längere Frist — **Ausgang:** offen; `AGENTS.md` §3.6: der Befund geht an den Planner, die Frist bleibt; die Ursache ist zu messen, nicht zu vermuten (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`).
- Eine Messung nennt eine Rate oder Zeit als Einzelzahl — **Ausgang:** offen; Zahl der Läufe, Last und Messender stehen dabei (`BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, 2×); erreicht ein Fund mit dieser Welle 3×, braucht der Eintrag in der Closure einen Ausgang.
- Der Stempel gilt am Host und gilt nicht im Container (Prüfsumme weicht ab) oder umgekehrt — **Ausgang:** offen; Entscheidung 4, Fänger: dieselbe Prüfsumme für denselben Commit an beiden Orten, eine Mutation ändert eine Eingabedatei.
- Der innere Daemon fällt auf den Daemon des Hosts zurück (Socket durchgereicht, Umgebungsvariable) und verfälscht „Host unberührt“ — **Ausgang:** offen; Fänger: Listen des Hosts vor und nach dem Lauf, eine Mutation lässt den Lauf den Host-Daemon rufen.
- Ein Lauf am Limit (Speicher) beendet den Container mitten im Gate, und der Stempel oder der Nachweis bleibt halb — **Ausgang:** offen; Zusage: ein roter Lauf lässt den Stempel unverändert (`SPEC-102`), Fänger im Raster.
- Nebenläufigkeit: zwei Container teilen Host-Ressourcen (Platte, CPU) und stören sich doch — **Ausgang:** offen; Messung zu zweit, nicht Zusicherung; ein akzeptiertes Negativ nennt Grund und Folge (`AGENTS.md` §3.15).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* als Gerüst; jede Zelle nennt die Frage, der Ort ist der Vertrag des Architect; eine Zelle ohne Ort hält den Implementer an — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Lauf auslösen | Zweig nicht angegeben: (Vertrag, offen) | leerer Zweigname: (Vertrag, offen) | Zweig nicht eingeliefert, Zweig mit anderem Stand als der Klon: (Vertrag, offen) | Signal, Verbindung bricht: Lauf endet oder läuft weiter? Stempel? (Vertrag, offen) | Lauf auf demselben Commit: Stempel entscheidet (Vertrag, offen) |
| Lauf im Arbeitsstand | Daemon nicht bereit: Probe rot, Lauf verweigert (Vertrag, offen) | Arbeitsstand leer: (Vertrag, offen) | Arbeitsstand schmutzig (ungetrackte Dateien): (Vertrag, offen) | Limit überschritten, Container endet: (Vertrag, offen) | zweiter Lauf im selben Container zugleich: (Vertrag, offen) |
| Stempel behandeln | Stempel fehlt: Lauf erzwungen (`SPEC-102`) | leerer Stempel: (Vertrag, offen) | Stempel eines anderen Standes: Lauf erzwungen (`SPEC-102`) | roter Lauf: Stempel unverändert (`SPEC-102`) | grüner Lauf zweimal: zweiter entfällt (`SPEC-102`) |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen:

| Paar | Entscheidung | Fänger |
|---|---|---|
| Frist je Fall gerissen und Limit des Speichers erreicht | offen (Vertrag): welche Ursache gemeldet wird | offen |
| Stempel stimmt und Arbeitsstand schmutzig | offen (Vertrag): Lauf entfällt nicht | offen |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30; keiner steht bei 3× und offen). Für diesen Slice zählen: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, `offen`) — die Frist je Fall unter dem Limit des Containers; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, `offen`) — Messungen mit Zahl, Last und Messendem; ein drittes Auftreten in dieser Welle macht aus ihm eine Lücke mit eigenem Ausgang in der Welle-Closure; `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (1×, `offen`) — der Stempel im Container (Entscheidung 4); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — Betriebsregeln in §3.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
