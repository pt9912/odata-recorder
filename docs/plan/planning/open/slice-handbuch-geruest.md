# Slice slice-handbuch-geruest: Gerüst des Benutzerhandbuchs, Anpassung des Standards, Kapitel über Belegtes

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md).

**Bezug:** `LH-RB-01` (der unterstützte Teilumfang ist verbindlich zu dokumentieren), `LH-RB-03` und `LH-QA-06` (Bauen und Start unter Linux und Docker ohne externe Infrastruktur), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, Docker-only).

**Berührte Spec-Stellen:** `SPEC-110` (Startkonfiguration: Syntax der Werte und Prüfung) · `spezifikation.md §3` (Defaults und Konstanten) — der Architect ergänzt, was die Kapitel über Belegtes sonst tragen.
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

**Ziel:** docs/user/benutzerhandbuch.md liegt als Gerüst nach dem Standard (`docs/user/benutzerhandbuch-standard.md`, Abschnitt 14) vor, mit Kopf (Handbuch-Version, Software-Version, Stand, Autor, Gültigkeitsbereich) und Kapiteln nur über Gegenstände, die das gebaute Binary heute zeigt und ein Test oder Gate belegt; die Anpassung des Standards an dieses Repo ist entschieden und niedergelegt.

**Warum kein eigener Slice für den Standard:** Der Standard existiert (`docs/user/benutzerhandbuch-standard.md`, vom User bereitgestellt, Kopie des Standards eines Nachbar-Repos). Er ist allgemein gehalten (Oberfläche, Anmeldung, Rollen, Screenshots, Support, FAQ) und passt auf ein Kommandozeilen- und HTTP-Werkzeug nur teilweise. Die Anpassung ist klein und gehört zum Gerüst, weil das Gerüst sie als Erstes braucht.

**Auftrag an den Architect** (vor dem ersten Code-Commit, in der Spezifikation oder, wo er den Ort wählt, in `docs/user/` oder `harness/vertraege/`; `AGENTS.md` §3.12 und §3.16):

1. **Zielgruppen und Begriffe.** Hauptnutzer nach dem Lastenheft (§2): Anwendungsentwickler und Testautomatisierung/QA; Betreiber: das Plattform-/CI-Team. Der Architect legt fest, ob ein Handbuch beide bedient (Vorschlag des Planners: ja, ein Handbuch, Kapitel nach Aufgaben, Betreiber-Aufgaben als eigene Kapitel) und welche Begriffe das Handbuch einführt (Aufzeichnung, Session, Workspace, Baseline, Match-Key; das Glossar von `harness/conventions.md` ist der Anfang).
2. **Anpassung des Standards.** Je Abschnitt des Standards (1 bis 14): gilt, gilt angepasst (Wortlaut), gilt nicht (Grund). Die Abschnitte 6 (Screenshots), 7 (Rollen und Rechte), 11 (Barrierefreiheit: Bilder, PDF) und Teile von 3 (Anmeldung, Support) beziehen sich auf eine Oberfläche, die der Recorder nicht hat. Ort der Anpassung: ein Abschnitt am Ende des Standards oder ein Vertrag; die Datei des Standards bleibt ansonsten unverändert, damit der Abgleich mit der Quelle möglich bleibt. Der Planner empfiehlt den Abschnitt im Standard, weil Autor und Prüfer dort ohnehin lesen.
3. **Ist-Zustand und Nachweis je Aussage.** Die Regel des Nachbar-Repos (das Handbuch beschreibt nur das gebaute Binary, keine Chronik, kein Zielstand, kein Verweis auf Spezifikation, ADRs, Slices, Wellen oder Reviews; ein Beispiel gilt erst als belegt, wenn es als Datei gegen das gebaute Binary läuft) ist als Regel dieses Repos zu entscheiden: Wortlaut, Klassen von Aussagen und der Beleg je Klasse (Test, Gate, Probe). Die Änderung an `AGENTS.md` (§3.11) ist ein Diff-Vorschlag an den User, kein Edit des Slice.
4. **Kopf und Pflege.** Quelle der Software-Version (Ausgabe von `version` des gebauten Binary), Regel für Handbuch-Version und Stand, Gültigkeitsbereich (Binary aus `make build`), Form des Änderungsverlaufs (Standard Abschnitt 9: je Version eine Zeile in Betreibersicht, ohne interne Kennungen).
5. **Welche Kapitel schon belegt sind.** Tabelle Kapitel · Gegenstand · Beleg, am Code geprüft: Voraussetzungen (Docker und GNU `make`, `LH-RB-03`), Bauen und Version, Start und Startkonfiguration (`SPEC-110`: Umgebungsvariablen, Defaults, Fehler des Starts mit `ODR-E2001`), Glossar, Änderungsverlauf. Kapitel ohne Beleg stehen nicht im Gerüst.
6. **Ort der belegten Beispiele.** Beispiele (Konfigurationen, Aufrufe) liegen als Dateien vor und laufen gegen das gebaute Binary. Der Build-Kontext der Stufe `test` enthält womöglich `docs/` nicht; der Architect nennt den Ort der Beispiel-Dateien und den Test, der sie fährt, und die Bindung an das Handbuch (Beispiel im Handbuch gleich Datei).
7. **Doku-Gate.** Wie das Handbuch das Doku-Gate (`make docs-check`) passiert: Verzeichnis ist gescannt, die Klasse ist `aussen` (kein Stratum darf es nennen); Entscheidung, ob das Handbuch Kennungen (`LH-*`, `ADR-*`) trägt (Planner: nein, Standard Abschnitt 9 und Ist-Zustand) und ob es Pfade in Inline-Code nennt (`codepaths` prüft die Wurzeln `spec`, `docs`, `harness`).
8. **Rang.** `docs/user/*` steht in der Source Precedence auf Rang 6 (`harness/README.md`, `AGENTS.md` §2); der Architect bestätigt, dass daraus nichts folgt, was `harness/README.md` ändern muss, und nennt sonst die Stelle.
9. Raster *Operation × Argumentart* (§6) und Tabelle *Zusage · Fänger* (§3) in diesem Plan füllen, Größe angeben.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Kapitel über Aufzeichnen und Wiedergeben — `slice-handbuch-record-replay`; sein Gegenstand existiert für Anwender erst nach M2 (`slice-bootstrap-m2`).
- Tabellen der Fehlercodes (`### Fehlercodes`, `### Warnungen`) — `slice-handbuch-meldungskatalog`; die Form der Tabelle liegt im Vertrag des Abgleichs. Das Gerüst lässt Fehlerbehebung als Überschrift nur dort stehen, wo ein Beleg sie trägt.
- Kapitel über Mock, Hybrid, Admin-API, Snapshots, TLS des Listeners — Folgewellen (Welle §6).
- Änderungen an `AGENTS.md`, Commands, Typen, Skills — Diff-Vorschlag an den User, den der Architect oder Planner liefert.
- Änderungen an Code oder Binary — Schicht: das Gerüst beschreibt, was da ist; findet der Architect eine Lücke zwischen Spezifikation und Binary, geht sie als Befund an den Planner.

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

- [ ] docs/user/benutzerhandbuch.md mit Kopf und den Kapiteln aus dem Auftrag Nr. 5; jede Aussage über Verhalten hat ihren Beleg nach Nr. 3, jedes Beispiel liegt als Datei vor und läuft in der Stufe `test` gegen das gebaute Binary (`LH-RB-01`, `LH-RB-03`, `LH-QA-06`).
- [ ] Anpassung des Standards an dieses Repo niedergelegt (Auftrag Nr. 2); der Diff-Vorschlag für `AGENTS.md` liegt dem User vor (Nr. 3).
- [ ] `make gates` grün; `make docs-check` meldet das Handbuch ohne Befund.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (hier: die Kette `docs/user/` ist selbst das Update; `harness/README.md` nur, wenn Nr. 8 eine Stelle nennt).
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
| docs/user/benutzerhandbuch.md | neu | Kopf und Kapitel über Belegtes (Auftrag Nr. 4, 5) |
| `docs/user/benutzerhandbuch-standard.md` | update | Anpassung an dieses Repo, soweit Nr. 2 sie dort ansiedelt |
| Beispiel-Dateien des Handbuchs und Test | neu | Ort und Bindung nach Nr. 6; ein Fall je Beispiel gegen das gebaute Binary |
| Spec-Stelle oder Vertrag der Regel | update oder neu | Ist-Zustand und Nachweis je Aussage (Nr. 3), falls der Architect sie dort ansiedelt |

Der Architect füllt die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und ergänzt die Zeilen oben; sie ist beim Start leer.

| Zusage | Fänger |
|---|---|
| — | — (Architect) |

**Betriebsregeln für die Rollen dieses Slice:** Dateien des Repos nur mit `Edit` und `Write`, kein `sed -i`, kein Heredoc auf Repo-Dateien (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`). Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII (Konfigurationsbeispiele, Texte) entsteht aus Bytewerten und die geschriebene Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18). Der Handoff ist ein Commit. Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills gehen als Diff-Vorschlag an den User.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): Die Welle ist gestartet (`slice-groessenmarke-werkzeug` in `done/`); der Architect hat die Aufträge aus §1 beantwortet und die Tabelle in §3 sowie das Raster in §6 gefüllt, bevor der erste Code-Commit entsteht.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner setzt sie nach dem Architect mit `make groessenmarke` (`SPEC-128`) aus der Tabelle in §3, Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; wie ein Leser fremder Bytes zählt das Handbuch nicht, wohl aber ein Test, der Beispiel-Dateien liest.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem Code über 3000 nach der Formel (`AGENTS.md` §3.14); ein Schnitt beim Handoff nur an dem Teil, den §4 vor dem Code benennt. Vorab benannter Teil: der Test der Beispiel-Dateien, wenn er keinen Mechanismus mit dem Rest teilt.
- `in-progress` → `open` (blockiert): Der Architect findet eine Lücke zwischen Spezifikation und Binary (ein Kapitel hätte keinen Beleg); das ist ein Befund an den Planner, kein Kapitel mit „noch nicht“.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und die Closure-Notiz mit Lerneintrag liegt vor. `make gates` fährt in diesem Slice die Beispiel-Dateien gegen das gebaute Binary; der Lauf der Gates im Hauptbaum ist die Bedingung (die Welle wurde in einem Worktree geplant, in dem `make gates` nicht lief).

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Das Binary zeigt heute nur Start und Version (Skelett aus Composition Root und Version): Das Kapitel „Konfiguration“ beschreibt nur, was der Start prüft und meldet; ein Satz über ein Verhalten, das erst `slice-bootstrap-m2` liefert, wäre Chronik oder Zielstand — **Ausgang:** offen; der Architect legt je Kapitel den Beleg fest (Auftrag Nr. 5), ein Kapitel ohne Beleg entfällt.
- Der Standard passt nur teilweise (Oberfläche, Rollen, Screenshots) — **Ausgang:** offen; Auftrag Nr. 2.
- Der Build-Kontext der Stufe `test` enthält `docs/` womöglich nicht, ein Beispiel läge dann außerhalb des Tests — **Ausgang:** offen; Auftrag Nr. 6 nennt den Ort der Dateien.
- Randformen (`AGENTS.md` §3.12, §3.16): Raster unten als Gerüst; die Zellen füllt der Architect vor dem ersten Code-Commit, eine leere Zelle ist ein Fehler des Plans — **Ausgang:** offen.
- Die Regel „Ist-Zustand“ liegt nur im Slice, bis der User den Diff für `AGENTS.md` annimmt — **Ausgang:** offen; Adresse: der User (Diff-Vorschlag, Auftrag Nr. 3).
- Kennungen im Handbuch brechen `docs-check` (`ADR-*` ist Link-Pflicht) oder verletzen den Standard (Abschnitt 9) — **Ausgang:** offen; Auftrag Nr. 7.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Beispiel-Datei gegen das Binary fahren | Architect | Architect | Architect | Architect | Architect |
| Kopf lesen (Software-Version gegen `version`) | Architect | Architect | Architect | gilt nicht: reiner Text, Architect bestätigt | Architect |
| Konfigurationswert im Handbuch nennen | Architect | Architect | Architect | gilt nicht: Start prüft vor jeder Wirkung, Architect bestätigt | Architect |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30). Für diesen Slice zählen: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `verkörpert` in `AGENTS.md` §3.15; für ein Handbuch: ein Satz sagt mehr zu als die Prüfung), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `verkörpert` in `AGENTS.md` §3.9; das Handbuch ist eine weitere Stelle, die eine Berichtigung trägt), `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×, `verkörpert` in `AGENTS.md` §3.18; Konfigurationsbeispiele) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`; Betriebsregeln in §3). Kein Eintrag erreicht mit dieser Welle 3× ohne Ausgang; die Zahlen stehen unter `evidence/` des jeweiligen Eintrags.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
