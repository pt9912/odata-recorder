# Slice slice-handbuch-meldungskatalog: Abgleich der Fehlercodes zwischen Code, Spezifikation und Handbuch

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md).

**Bezug:** `LH-QA-05` (stabile Fehlerklassen, unterscheidbar für den Anwender), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (der Vertrag des Abgleichs liegt in `harness/vertraege/`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben außerhalb der Gate-Kette; Trigger der Laufzeit), [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Vorbild für ein Programm im Gate mit Eingaben vom Host).

**Berührte Spec-Stellen:** `spezifikation.md §4` (Katalog der Fehler-Codes, Spalte *Code*) · `SPEC-109` (Abbildung auf Codes) · `SPEC-113` (Rückfall `ODR-E9001`) · `SPEC-102` (Gegenproben) · `SPEC-128` (Vorbild eines Werkzeug-Vertrags mit Treiber und Gate-Anschluss) · der neue Vertrag in `harness/vertraege/`, dessen Kennung der Architect vergibt.
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

**Ziel:** Ein Gate gleicht die Fehlercodes `ODR-E…` dreier Quellen ab — die Code-Tabelle in `src/hexagon/model/fehler.go` (C), den Katalog in `spezifikation.md §4` (K) und die Tabellen des Benutzerhandbuchs (H) — und wird rot, wenn sie auseinanderlaufen; das Handbuch trägt dafür einen Abschnitt mit der Tabelle der Fehlercodes. Vorbild ist der Abgleich `meldungskatalog-check` des Nachbar-Repos pgwire-recorder (Vertrag SPEC-054 dort), nicht übernommen, sondern auf dieses Repo geschnitten.

**Auftrag an den Architect** (vor dem ersten Code-Commit, im neuen Vertrag unter `harness/vertraege/`; `AGENTS.md` §3.12, §3.15, §3.16):

1. **Vertrag und Form der Quellen.** C ist eine Struktur-Tabelle (`{CodeE2001, "ODR-E2001", []int{0}, ZielConfig, "…"}`, 34 Einträge), keine Liste von Konstanten wie im Vorbild; K hat als erste Zelle die Kennung `SPEC-nnn` und als zweite den Code. Der Architect legt die Leseregeln fest (welche Zeile ist ein Code, Form `^ODR-E[0-9]{4}$`, doppelt, Sperren bei fehlender Quelle, Ausgabe und Ausgang) und die Grenze (was nicht geprüft ist: Spalten Bedingung, Aktion, Bedeutung).
2. **Die Menge H — die eigentliche Entscheidung.** Das Handbuch sagt nur zu, was das Binary zeigt (Slice `slice-handbuch-geruest`, Auftrag Nr. 3). Der Katalog führt aber Codes, die erst mit Folgewellen erreichbar sind (Admin-API: `ODR-E7005`, `ODR-E7006`, `ODR-E7012` …; Mock: `ODR-E6001` …). Optionen: (a) H führt alle Codes von C und sagt zu jedem nur die Bedeutung nach dem Katalog, ohne Zusage über den Weg dorthin; (b) H führt genau die Codes, die das gebaute Binary erzeugen kann, und C trägt ein Merkmal dafür (die Spalte *Ziel* in `fehler.go` ist ein Anfang); (c) H führt Codes nach Klassen und der Abgleich prüft nur die Klassen. Der Planner neigt zu (b) wegen des Ist-Zustands, sieht aber, dass (b) ein neues Merkmal in der Code-Tabelle verlangt, das sich bei jeder Welle ändert; der Architect entscheidet mit Grund und nennt die Folge für die Folgewellen (deren Kapitel erweitern H).
3. **Code außerhalb der Tabelle.** Literale `"ODR-…"` stehen heute auch in `src/hexagon/model/konflikt.go`, `daten_fehler.go`, `redaktor.go` und `src/bootstrap/konfig.go`; das Vorbild verbietet sie außerhalb der Tabelle. Der Architect entscheidet, ob der Abgleich sie als Befund meldet (dann ist das ein Code-Umbau und gehört nicht in diesen Slice, sondern als Folge-Slice benannt) oder sie ausdrücklich zulässt (Grenze).
4. **Ort in der Kette und Laufzeit.** Docker-only (`AGENTS.md` §3.1): ein Programm in Go unter `tools/` in einer Stufe des `Dockerfile` oder ein Treiber mit Eingaben vom Host wie bei `SPEC-128`; der Architect wählt die Form, nennt Eingaben (Pfade von C, K, H) und was in den Kontext der Stufe kopiert wird. Hängt das Ziel an `GATE_CHECKS` und die Gegenprobe an die Liste der Gegenproben, wie `SPEC-102` es für andere Werkzeuge hält? Die Stufung folgt dem Vertrag der Gegenproben und [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md).
5. **Gegenproben.** Ein Fall je Punkt des Vertrags (Fehlform in einem Wegwerf-Baum): fehlender Code in K, fehlender in H, Code nur in H, Form, doppelt, Sperren, Ausgang. Die Aufrufstellen sind eine Menge (`AGENTS.md` §3.15): ein Fall je Quelle und je Richtung der Differenz.
6. **Zeit.** Der Planner verlangt die Messung der Zusatzzeit von Gate und Gegenproben (Schritt `make gates` und `make gegenproben`, vorher und nachher, Zahlen im Plan). Die Frage, ob [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Trigger „Lauf über 30 s bei geänderter Eingabe“; gemessen 56,9 s im Schritt) neu zu fassen ist, entscheidet dieser Slice nicht (siehe §6).
7. **Nachzüge.** `harness/README.md` (Sensors-Zeile samt Bindung, höchstens 200 Zeichen in den Zellen, `structure` in `.d-check.yml`), `AGENTS.md` nur als Diff-Vorschlag an den User; Tabelle *Zusage · Fänger* (§3), Raster (§6), Größe.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Inhalt der Tabelle im Handbuch über Codes, die das Binary erst nach M2 erzeugt — `slice-handbuch-record-replay` erweitert H um die Fehlerbilder, die M2 erreichbar macht; dieser Slice bringt Abschnitt, Form und Gate.
- Umbau von Code, der Literale `"ODR-…"` außerhalb der Tabelle trägt — andere Schicht; Folge-Slice nur, wenn der Architect in Nr. 3 sie verbietet (dann Befund an den Planner mit Kennung des Nehmers, `AGENTS.md` §3.13).
- Prüfung von Bedeutung, Maßnahme und Klasse einer Zeile gegen den Text — der Abgleich liest Mengen, er urteilt nicht über Inhalt (Grenze des Vorbilds); Inhalt prüft der Probelauf in `slice-handbuch-abnahme`.
- Änderung von [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) oder des Triggers der Gegenproben — Entscheidung des Users (§6).

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

- [ ] Vertrag des Abgleichs in `harness/vertraege/` und das Programm samt Treiber und Ziel in `GATE_CHECKS`; Abschnitt mit Tabelle im Handbuch; jede Zusage des Vertrags hat einen Fänger, vom Implementer rot gesehen (`LH-QA-05`).
- [ ] Gegenproben zu allen Punkten des Vertrags laufen in `make gegenproben`.
- [ ] Zusatzzeit gemessen (Gate und Gegenproben, Zahlen im Plan); die ADR-Frage ist nicht entschieden, sondern dem User vorgelegt.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `harness/README.md` (Sensors-Zeile mit Bindung), Handbuch; `AGENTS.md` als Diff-Vorschlag.
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
| Vertrag unter `harness/vertraege/` | neu | Leseregeln, Mengen, Sperren, Ausgabe, Grenze (Auftrag Nr. 1 bis 3) |
| Programm unter `tools/`, Treiber, Stufe im `Dockerfile` | neu | Ort und Form nach Nr. 4 |
| Fragment unter `harness/mk/` | neu | Ziel und Anbindung an `GATE_CHECKS` und die Gegenproben |
| Tests des Programms | neu | ein Fänger je Zusage, Mutationen vom Implementer rot gesehen |
| Gegenproben (`tools/gegenproben/`) | update | ein Fall je Punkt des Vertrags (Nr. 5) |
| docs/user/benutzerhandbuch.md | update | Abschnitt der Fehlercodes mit Tabelle |
| `harness/README.md` | update | Sensors-Zeile mit Bindung |

Der Architect füllt die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und ergänzt die Zeilen oben; sie ist beim Start leer.

| Zusage | Fänger |
|---|---|
| — | — (Architect) |

**Betriebsregeln für die Rollen dieses Slice:** Dateien des Repos nur mit `Edit` und `Write`, kein `sed -i`, kein Heredoc auf Repo-Dateien (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`). Fixtures mit Backslash-u-Folge, BOM, NBSP oder Zeichen jenseits von ASCII entstehen aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18). Der Handoff ist ein Commit. Docker-Läufe mit `--memory` und wenige parallel. Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills gehen als Diff-Vorschlag an den User.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-handbuch-geruest` und `slice-bootstrap-m2` liegen in `done/` (das Gerüst trägt den Ort der Abschnitte; erst nach M2 liest das Binary Schlüssel und erzeugt Fehlercodes, `Run` ruft heute weder `LadeKonfiguration` noch `LadeOderMelde`); der Architect hat die Aufträge aus §1 beantwortet, die Tabelle in §3 und das Raster in §6 gefüllt, bevor der erste Code-Commit entsteht.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner setzt sie nach dem Architect mit `make groessenmarke` (`SPEC-128`) aus der Tabelle in §3, Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen. Der Slice liefert einen Leser fremder Bytes (Quelltext-Tabelle, Spezifikation und Handbuch als Texte, die ein anderer schreibt): Zeilen der Tabelle × 65 statt × 40, dazu die Produktcode-Größe des Architect, × 1,25.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nur vor dem Code über 3000 nach der Formel; ein Schnitt beim Handoff nur an dem Teil, den §4 vor dem Code benennt. Vorab benannter Teil, wenn der Architect ihn bestätigt: die Gegenproben (wie bei `slice-groessenmarke-werkzeug` Teil G), zuletzt in eigenem Commit, sofern sie keinen Mechanismus mit dem Rest teilen.
- `in-progress` → `open` (blockiert): Der Architect verlangt in Nr. 3 den Umbau von Code mit Literalen außerhalb der Tabelle; das ist ein Befund an den Planner (neuer Slice, `AGENTS.md` §3.13), nicht Code in diesem Slice.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Laufzeit: Der Abgleich und seine Gegenproben verlängern `make gates` und `make gegenproben`; [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) setzt den Trigger bei einem Lauf über 30 s bei geänderter Eingabe, gemessen waren 56,9 s im Schritt, und die Entscheidung des Users dazu ist offen — **Ausgang:** offen; Adresse: der User. Der Slice misst die Zusatzzeit (DoD) und entscheidet die ADR-Frage nicht; eine neue oder geänderte ADR ist ein Vorschlag an den User.
- Die Menge H (Auftrag Nr. 2) kollidiert mit dem Ist-Zustand, wenn H Codes führt, die das Binary noch nicht erzeugt — **Ausgang:** offen; Entscheidung des Architect vor dem Code.
- Literale `"ODR-…"` außerhalb der Tabelle (Auftrag Nr. 3) — **Ausgang:** offen; Entscheidung des Architect, bei Verbot ein Folge-Slice mit Kennung, den der Planner anlegt.
- Der Abgleich liest Texte, die andere schreiben (Quelltext, Spezifikation, Handbuch): Eine Randform, die der Vertrag nicht nennt, wird im Review gefunden (`BEO-REPO/randform-erst-im-review-entschieden`, 20×, `verkörpert` in `AGENTS.md` §3.16) — **Ausgang:** offen; das Raster unten ist das Gerüst, die Zellen füllt der Architect vor dem ersten Code-Commit.
- Eine Allaussage des Vertrags („jeder Code in C steht in K“) ohne Fänger über alle Fälle (`BEO-REPO/universalaussage-im-vertrag-ungeprueft`, 9×, `verkörpert` in `AGENTS.md` §3.16) — **Ausgang:** offen; Fänger als Schleife über alle Codes der Tabelle.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Quelle C lesen | Architect | Architect | Architect | Architect | Architect |
| Quelle K lesen | Architect | Architect | Architect | Architect | Architect |
| Quelle H lesen | Architect | Architect | Architect | Architect | Architect |
| Mengen vergleichen | Architect | Architect | Architect | gilt nicht: reiner Lesevorgang, Architect bestätigt | Architect |
| Befunde ausgeben | Architect | Architect | Architect | Architect | Architect |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Verstößen am selben Code:

| Paar | Entscheidung | Fänger |
|---|---|---|
| Code fehlt in K und in H | Architect | Architect |
| Code doppelt in H und fehlt in C | Architect | Architect |
| Form verletzt und Code doppelt | Architect | Architect |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30). Für diesen Slice zählen: `BEO-REPO/randform-erst-im-review-entschieden` (20×, `verkörpert` in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `verkörpert` in `AGENTS.md` §3.15), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `verkörpert` in `AGENTS.md` §3.16), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `verkörpert` in `AGENTS.md` §3.15), `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) für die Gegenproben und `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (16×, `geplant` in `slice-groessenmarke-werkzeug`) für die Marke. Kein Eintrag erreicht mit dieser Welle 3× ohne Ausgang; die Zahlen stehen unter `evidence/` des jeweiligen Eintrags.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
