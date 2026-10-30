# Slice slice-handbuch-abnahme: Probelauf des Handbuchs durch eine Rolle ohne Vorwissen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md).

**Bezug:** `LH-QA-05` (Bedienbarkeit und Fehlerdiagnose: der Anwender unterscheidet Fehler anhand stabiler Klassen), Standard `docs/user/benutzerhandbuch-standard.md` Abschnitt 13 (Qualität prüfen: „Gib das Handbuch jemandem, der die Software nicht kennt“).

**Berührte Spec-Stellen:** —

**Hinweis zur Rolle:** Die Validierung prüft gegen den Bedarf außerhalb der Artefakte und läuft in eigener Kadenz (Baseline-Regelwerk `modul-01-entwicklungszyklus.md`); die kanonische Rolle dafür ist `validator`. Der Probelauf ist ein Rollenlauf mit der Rolle `validator`.
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

**Ziel:** Eine Rolle ohne Vorwissen führt jede Aufgabe des Handbuchs allein nach dem Handbuch am gebauten Binary aus; die Befunde sind behoben oder tragen einen Ausgang, und Kopf (Handbuch-Version, Software-Version, Stand) und Änderungsverlauf stimmen mit dem gebauten Binary überein.

**Auftrag an den Architect** (vor dem Lauf; `AGENTS.md` §3.12, §3.16):

1. **Form des Probelaufs.** Eingabe der Rolle: nur das Handbuch und das gebaute Binary samt Docker (kein Quelltext, keine Spezifikation, kein Plan, kein Zugriff auf `spec/`, `src/`, `docs/plan/`); Ausgabe: ein Bericht je Aufgabe (ausgeführt ohne Rückfrage ja oder nein, Schritt der ersten Rückfrage, Abweichung zwischen Handbuch und Binary). Der Architect legt das Schema des Berichts und den Ablageort fest und nennt, wie der Lauf die Eingabe begrenzt (ein Rollenlauf mit eigenem Arbeitsverzeichnis, der nur das Handbuch enthält, ist der Vorschlag des Planners).
2. **Prüfliste.** Standard Abschnitt 13 als Liste von Prüfpunkten, je Punkt Quelle des Urteils (Gate, Probelauf oder Mensch): beschriebene Funktionen vorhanden, Befehle und Ausgaben stimmen, Fehlermeldungen richtig beschrieben, tote Links (Gate), sensible Daten entfernt, Sprache einheitlich. Punkte, die der Standard für eine Oberfläche nennt (Button-Namen, Screenshots, PDF, Suche), entfallen mit Grund nach der Anpassung aus `slice-handbuch-geruest`.
3. **Behandlung der Befunde.** Ein Befund über das Handbuch behebt der Implementer in diesem Slice; ein Befund über das Binary (das Handbuch ist richtig, das Binary weicht von der Spezifikation ab) ist ein Befund an den Planner mit der Kennung des Nehmers in `done/` (`AGENTS.md` §3.13), kein Code in diesem Slice.
4. **Pflege von Kopf und Verlauf.** Wer Version und Stand setzt, und woran ein Nachzug scheitert, wenn das Binary eine andere Version meldet (Beleg: Ausgabe von `version`).
5. Raster *Operation × Argumentart* (§6) und Tabelle *Zusage · Fänger* (§3) füllen, Größe angeben.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Kapitel über Mock, Hybrid, Admin-API — Folgewellen; der Probelauf prüft, was das Handbuch zum Zeitpunkt der Closure sagt.
- Code- oder Binary-Änderungen — andere Schicht; ein Binary-Befund geht an den Planner.
- Ein wiederkehrender Probelauf als Gate — kostet einen Rollenlauf pro Änderung und ist kein deterministischer Sensor; der Lauf gehört zur Closure der Welle und, mit Beschluss der Folgewellen, zu deren Closure (Vorschlag an den User in der Closure-Notiz der Welle).
- Eine Übersetzung oder Barrierefreiheitsprüfung nach Standard Abschnitt 11 — der Recorder hat keine Oberfläche; was bleibt (Überschriftenstruktur), prüft der Probelauf mit.

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

- [ ] Probelauf nach Auftrag Nr. 1 gefahren, Bericht je Aufgabe liegt im Repo (Ablage nach Architect); jeder Befund ist im Handbuch behoben oder trägt einen Ausgang in §6 (`LH-QA-05`).
- [ ] Prüfliste nach Nr. 2 abgearbeitet; Kopf und Änderungsverlauf stimmen mit der Ausgabe von `version` überein.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Handbuch nach den Befunden nachgezogen.
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
| Bericht des Probelaufs | neu | Ablage und Schema nach Auftrag Nr. 1 |
| docs/user/benutzerhandbuch.md | update | Behebung der Befunde, Kopf und Änderungsverlauf |

Der Architect füllt die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und ergänzt die Zeilen oben; sie ist beim Start leer.

| Zusage | Fänger |
|---|---|
| — | — (Architect) |

**Betriebsregeln für die Rollen dieses Slice:** Dateien des Repos nur mit `Edit` und `Write`, kein `sed -i`, kein Heredoc auf Repo-Dateien (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`). Der Bericht entsteht im Repo, nicht nur im Rollenlauf (`BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`). Docker-Läufe mit `--memory` und wenige parallel. Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills gehen als Diff-Vorschlag an den User.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-handbuch-record-replay` liegt in `done/` (alle Kapitel des Handbuchs stehen); der Architect hat die Aufträge aus §1 beantwortet, die Tabelle in §3 und das Raster in §6 gefüllt.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner setzt sie nach dem Architect mit `make groessenmarke` (`SPEC-128`), Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen. Erwartet wird ein Slice ohne Code; die Marke folgt der Formel mit den Zeilen der Tabelle.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem Code über 3000 nach der Formel; für diesen Slice nicht erwartet.
- `in-progress` → `open` (blockiert): Der Probelauf findet einen Binary-Befund, der eine Aufgabe des Handbuchs unausführbar macht; das ist ein Befund an den Planner, und der Slice wartet auf den Nehmer.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und die Closure-Notiz mit Lerneintrag liegt vor; der Lerneintrag nennt, was der Probelauf gefunden hat, das weder Test noch Gate gefunden hätte (geschärfte Regel, neuer Sensor oder benannte Spec-Lücke).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Rolle hat Vorwissen (Kontext aus dem Repo, Training über das Werkzeug) und liest das Handbuch deshalb zu großzügig — **Ausgang:** offen; Auftrag Nr. 1 begrenzt die Eingabe, der Bericht nennt, woraus die Rolle einen Schritt ableitete.
- Der Lauf braucht Docker und das gebaute Binary; ein parallel laufender Rollenlauf belastet den Daemon (Speicherlimit der Rollenläufe) — **Ausgang:** offen; Läufe mit `--memory` und wenige parallel.
- Der Bericht der Rolle liegt nur im Lauf (`BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo`, 3×, Ausgang in `state.md` des Eintrags) — **Ausgang:** offen; der Bericht ist ein Artefakt im Repo (Auftrag Nr. 1).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster unten als Gerüst; die Zellen füllt der Architect, eine leere Zelle ist ein Fehler des Plans — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Aufgabe des Handbuchs ausführen | Architect | Architect | Architect | Architect | Architect |
| Kopf gegen `version` prüfen | Architect | Architect | Architect | gilt nicht: reiner Vergleich, Architect bestätigt | Architect |
| Befund einem Nehmer zuordnen | Architect | Architect | Architect | Architect | Architect |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30). Für diesen Slice zählen: `BEO-REPO/bericht-der-rolle-liegt-nicht-im-repo` (3×) und `BEO-REPO/befund-an-rolle-ohne-artefakt` (3×) für den Bericht des Probelaufs und die Zuordnung der Befunde, `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) für §3. Der Planner liest den Stand dieser Einträge beim Start neu.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
