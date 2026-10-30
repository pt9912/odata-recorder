# Slice slice-dev-container-zugang: Zugang zum Arbeitscontainer per SSH und Schlüsselpaar

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-RB-04` (Schutz von Daten und Zugangsdaten: kein Geheimnis im Repo oder Image), `LH-RB-03` (Entwicklung in Docker-Umgebungen), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Vertrag unter `harness/vertraege/`). Eine neue ADR zur Schlüsselverwaltung kommt vom Architect (Auftrag 1).

**Berührte Spec-Stellen:** der Vertrag des Gerüsts aus `slice-dev-container-geruest` (Bereitschaftsprobe, erweitert) · ein neuer Vertrag oder ein neuer Abschnitt dort für den Zugang (Nummer vergibt der Architect) · `harness/README.md` (Werkzeuge, „Nicht behauptet“).
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

**Ziel:** Der laufende Arbeitscontainer nimmt SSH-Verbindungen an, die sich mit einem hinterlegten Schlüssel ausweisen, und weist jede andere Anmeldung ab; kein privater Schlüssel liegt im Repo, im Image oder im Build-Kontext, und die Bereitschaftsprobe des Gerüsts umfasst den SSH-Dienst.

**Auftrag an den Architect** (Entscheidungen vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Planner entscheidet sie nicht):

1. **Schlüsselverwaltung** (ADR-Kandidat, sicherheitsrelevant): wer das Schlüsselpaar erzeugt und wo der private Teil liegt (beim Entwickler, nie im Container, nie im Repo), wie der öffentliche Teil in den Container kommt (Mount, Datei, Umgebung; Wirkung auf Image-Schichten und Verlauf), ob der Host-Schlüssel des Containers über Neustarts stabil ist (Fingerabdruck, keine Warnung beim Entwickler) oder je Start neu entsteht, und wie ein Schlüssel ersetzt wird. Eine Anleitung, die `StrictHostKeyChecking=no` oder Ähnliches zum Umgehen der Warnung nennt, ist keine Lösung (Geist von `AGENTS.md` §3.2).
2. **Anmeldung**: nur Schlüssel, kein Passwort, kein Root; welcher Benutzer; was er darf (Shell oder nur Git-Befehle); Port und Bindeadresse auf `u-boot` (Schleife oder Netz) und was bei belegtem Port geschieht (zwei Container auf einem Host).
3. **Probe**: Nachzug des Vertrags aus `slice-dev-container-geruest` um die Bedingung „SSH-Dienst nimmt Verbindungen an“; die Probe des Gerüsts ist dafür erweiterbar (Auftrag 3 dort). Der Nachzug steht im selben Commit wie die Änderung (`AGENTS.md` §3.9).
4. **Raster und Tabelle**: Raster *Operation × Argumentart* in §6 (Operationen Anmelden, Schlüssel hinterlegen; Argumente Schlüssel fehlend, leer, ungültig, falscher Typ, mehrere, zu lange Zeile; Rechte der Dateien falsch; Port belegt; Wiederholung) mit Ort je Zelle, und die Tabelle *Zusage · Fänger* in §3 (`AGENTS.md` §3.15) vor dem Übergang `next` → `in-progress`; die Abweisung ist eine Allaussage über Anmeldearten (kein Passwort, kein falscher Schlüssel, kein Root, kein leerer Schlüssel) und bekommt einen Fänger als Schleife über alle Arten (`AGENTS.md` §3.15).
5. **Prüfbarkeit**: Was ein Gate fängt (die Konfiguration des SSH-Dienstes als Text, `sshd -t`, ein Fall der Gegenproben gegen eine Kopie der Konfiguration; Anmeldeversuche gegen einen Wegwerf-Container ohne Privilegien) und was nur ein Lauf auf `u-boot` belegt (akzeptiertes Negativ mit Grund und Folge).
6. **Größe**: Schätzung und Marke wie in `slice-dev-container-geruest` Auftrag 9.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Bare-Repo, Klon und Einlieferung — `slice-dev-container-einlieferung`; sie nutzen den Zugang, sie begründen ihn nicht.
- Ein Geheimnis-System (Tresor, Schlüsselserver) — nicht gefordert; die Welle liefert keine Infrastruktur über Docker und `make` hinaus (`LH-QA-06` sagt „ohne verpflichtende externe Infrastruktur“ für das Produkt zu; der Arbeitscontainer folgt derselben Haltung).
- Mehrere Benutzer, Rollen-Trennung im Container, Rechte pro Rolle — nicht gefordert; ein Benutzer genügt dem Anlass der Welle. Folge: zwei Entwickler teilen den Zugang, bis ein Slice es ändert.
- Absicherung des Netzes um `u-boot` (Firewall, VPN) — Sache des Hosts, kein Inhalt des Repos.
- Eine Schlüssel-Rotation mit Ablaufdatum — der Architect nennt das Ersetzen (Auftrag 1); ein Ablauf ist Betrieb.

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

- [ ] Die Entscheidungen 1 und 2 des Architect stehen vor dem Code in einer ADR (Schlüsselverwaltung) mit Status `Accepted` und im Vertrag.
- [ ] Anmeldung mit dem hinterlegten Schlüssel gelingt; jede andere Art (kein Schlüssel, fremder Schlüssel, Passwort, Root, leerer Schlüssel) wird abgewiesen — ein Fall je Element, als Schleife (`AGENTS.md` §3.15).
- [ ] Kein privater Schlüssel in Repo, Image-Schichten oder Build-Kontext (Fänger nach Auftrag 5); die Probe des Gerüsts umfasst den SSH-Dienst; zu jeder Zusage ein Fänger in §3, `make gegenproben` gefahren.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Vertrag des Gerüsts (Probe) und `harness/README.md` im selben Commit nachgezogen (`AGENTS.md` §3.9).
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
| `docs/plan/adr/` (ADR zur Schlüsselverwaltung) und ADR-Index | neu / update | Entscheidung 1, Status `Accepted` vor dem Code |
| `harness/vertraege/` (Vertrag des Gerüsts, Zugang) | update / neu | Zusagen des Zugangs, Probe um den SSH-Dienst erweitert |
| Image, Konfiguration des SSH-Dienstes, Start-Skript (Orte nach dem Gerüst) | update / neu | Dienst, Schlüssel-Aufnahme, Host-Schlüssel |
| `tools/gegenproben/` und `SPEC-102` | update | Fälle zu den Zusagen, die ein Gate fängt (Auftrag 5) |
| `harness/README.md` | update | Zeile des Ziels, „Nicht behauptet“ für den Teil, den nur `u-boot` belegt |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) entsteht beim Architect **vor** dem Übergang `next` → `in-progress`; sie ist noch leer.

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`, höchstens zwei Container zugleich; Schlüssel für Tests entstehen im Test in einem Wegwerfverzeichnis, nie als Datei im Repo; Werte mit Escape-Folge oder Rohbytes aus Bytewerten und gegengelesen (`AGENTS.md` §3.18).


## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-dev-container-geruest` liegt in `done/`; die ADR zur Schlüsselverwaltung hat den Status `Accepted`; Vertrag, Tabelle *Zusage · Fänger* und Raster aus §6 stehen vor dem ersten Code-Commit (Architect, `AGENTS.md` §3.12). Danach setzt der Planner die Marke.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner setzt sie nach dem Architect auf die Schätzung × 1,25, aufgerundet auf volle Hundert, mit `make groessenmarke` (`SPEC-128`), falls `slice-groessenmarke-werkzeug` dann in `done/` liegt, sonst nach der Form von `AGENTS.md` §3.14; die Metrik zählt die Pfade aus §3, Tests und Fälle eingeschlossen.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke; offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten) und die Grenze 3000 vor dem Code.
- `in-progress` → `open` (blockiert): der Weg des Gerüsts erlaubt keinen SSH-Dienst im Container (Befund an den Planner; das Gerüst liegt dann in `done/`, keine Adresse, `AGENTS.md` §3.13).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, je Zusage des Vertrags ist eine Mutation rot gesehen (`AGENTS.md` §3.10), und eine Anmeldung auf `u-boot` mit Schlüssel und je Art der Abweisung ist belegt; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Schlüsselleck: ein privater Schlüssel gerät in eine Image-Schicht, den Build-Kontext oder das Repo — **Ausgang:** offen; Fänger nach Auftrag 5 (Suche in Schichten und Kontext), ein Treffer färbt das Gate rot.
- Warnung des Host-Schlüssels führt zum Umgehen der Prüfung beim Entwickler — **Ausgang:** offen; die ADR nennt den Weg ohne Umgehung (Auftrag 1), das Dokument des letzten Slice wiederholt ihn.
- Der Zugang erweitert die Angriffsfläche des Containers (Port am Host) — **Ausgang:** offen; Bindung und Rechte nach Auftrag 2, Wirkung in der ADR.
- Erweiterung der Probe bricht den Vertrag des Gerüsts (in `done/`): zulässig nur, weil das Gerüst die Probe als erweiterbar festlegt (Auftrag 3 dort) — **Ausgang:** offen; hat das Gerüst es nicht so entschieden, ist es ein Befund an den Planner (`AGENTS.md` §3.13).
- Zwei Container auf einem Host kollidieren am Port — **Ausgang:** offen; Zeile im Raster.
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* als Gerüst; jede Zelle nennt die Frage, der Ort ist der Vertrag des Architect; eine Zelle ohne Ort hält den Implementer an — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Schlüssel hinterlegen | keine Datei: Start verweigert oder Zugang ohne Schlüssel? (Vertrag, offen) | leere Datei: (Vertrag, offen) | ungültige Zeile, falscher Schlüsseltyp, mehrere Schlüssel, Zeile mit Optionen: (Vertrag, offen) | Abbruch beim Schreiben: halbe Datei? (Vertrag, offen) | gleicher Schlüssel zweimal: (Vertrag, offen) |
| Anmelden | kein Schlüssel angeboten: abgewiesen (Vertrag, offen) | leerer Benutzername: (Vertrag, offen) | fremder Schlüssel, Passwort, Root, unbekannter Benutzer: abgewiesen (Vertrag, offen) | Verbindung bricht während der Anmeldung: (Vertrag, offen) | Wiederholung nach Abweisung: gleiche Antwort (Vertrag, offen) |
| Host-Schlüssel bereitstellen | kein Schlüssel vorhanden: wird erzeugt? (Vertrag, offen) | leere Datei: (Vertrag, offen) | Rechte falsch, Schlüsseltyp unzulässig: (Vertrag, offen) | Abbruch beim Erzeugen: (Vertrag, offen) | Neustart: gleicher Fingerabdruck? (Vertrag, offen) |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen:

| Paar | Entscheidung | Fänger |
|---|---|---|
| gültiger Schlüssel und falscher Benutzer | offen (Vertrag): abgewiesen | offen |
| Schlüssel mit Optionszeile und Befehlsbeschränkung | offen (Vertrag) | offen |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30; keiner steht bei 3× und offen). Für diesen Slice zählt: `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — die Betriebsregeln stehen in §3, ein Auftreten wird als Beleg eingetragen. Sonst keine Treffer.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
