# Slice slice-handbuch-meldungskatalog: Abgleich der Kataloge (Fehlercodes, Konfigurationsschlüssel, Kommandos) zwischen Binary, Spezifikation und Handbuch

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md).

**Bezug:** `LH-QA-05` (stabile Fehlerklassen, unterscheidbar für den Anwender), `LH-RB-01` (nicht unterstützte Funktionen dürfen nicht stillschweigend als unterstützt erscheinen; der Katalog der Schlüssel und Kommandos hält das Handbuch an das, was das Binary kennt), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (der Vertrag des Abgleichs liegt in `harness/vertraege/`), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben außerhalb der Gate-Kette; Trigger der Laufzeit), [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (Vorbild für ein Programm im Gate mit Eingaben vom Host).

**Berührte Spec-Stellen:** `SPEC-110` (Startkonfiguration: Schlüssel, Nr. 2 die nicht gelieferten) · `spezifikation.md §3` (Defaults und Konstanten, Spalte *Name*) · `spezifikation.md §4` (Katalog der Fehler-Codes, Spalte *Code*) · `SPEC-109` (Abbildung auf Codes) · `SPEC-113` (Rückfall `ODR-E9001`) · `SPEC-102` (Gegenproben) · `SPEC-128` (Vorbild eines Werkzeug-Vertrags mit Treiber und Gate-Anschluss) · der neue Vertrag in `harness/vertraege/`, dessen Kennung der Architect vergibt.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** pt9912 (Zweig slice-handbuch-meldungskatalog)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Namensentscheidung:** Der Name `slice-handbuch-meldungskatalog` bleibt, §1 benennt die drei Kataloge ausdrücklich. Grund: Alle drei sind Kataloge im Sinn des Vorbilds (eine geschlossene Menge von Namen, die Code, Spezifikation und Handbuch gleichlautend führen müssen); eine Umbenennung in `slice-handbuch-abgleich` kostete einen reinen `git mv` als eigenen Commit (`AGENTS.md` §3.3) und den Nachzug jedes Verweises (`AGENTS.md` §3.9), ohne dass ein Leser den Gegenstand besser fände. Eine Zeile „Kataloge“ in Titel und Ziel genügt.

**Ziel:** Ein Gate hält drei Kataloge gegen das Handbuch: (1) die Fehlercodes `ODR-E…` (Code-Tabelle in `src/hexagon/model/fehler.go`, Katalog in `spezifikation.md §4`, Handbuch), (2) die Schlüssel der Startkonfiguration (`SPEC-110`; Quelle im Code: die Schlüssel, die `src/bootstrap/konfig*.go` liest, Spezifikation, Handbuch) und (3) die Kommandos und Optionen des Binary (Quelle: der Adapter `src/adapters/driving/cli`, Menge K nach `SPEC-138` Nr. 2: `version` und `start` (nur mit gesetztem `Start`); Handbuch). Regel (Ist-Zustand, Entscheidung des Users, Welle §1): Jeder Fehlercode, jeder Konfigurationsschlüssel und jedes Kommando, das das Binary kennt, kommt im Handbuch vor, und das Handbuch nennt nichts, was das Binary nicht kennt; sonst ist das Gate rot. Das Handbuch trägt dafür je Katalog einen Abschnitt mit einer Tabelle; das Gerüst hat „Aufruf und Version“ und „Fehlerbehebung“ angelegt, „Einstellungen“ und die Tabellen legt dieser Slice selbst an (Sendung von `slice-handbuch-geruest`). Vorbild ist der Abgleich `meldungskatalog-check` des Nachbar-Repos pgwire-recorder (Vertrag SPEC-054 dort), nicht übernommen, sondern auf dieses Repo geschnitten.

**Auftrag an den Architect** (vor dem ersten Code-Commit, im neuen Vertrag unter `harness/vertraege/`; `AGENTS.md` §3.12, §3.15, §3.16):

1. **Vertrag und Form der Quellen.** C ist eine Struktur-Tabelle (`{CodeE2001, "ODR-E2001", []int{0}, ZielConfig, "…"}`, 36 Einträge auf `main` am 2026-11-06 (37 mit `ODR-E6007` aus dem Zweig von `slice-odata-json-eingabe`; der Architect zählt am Stand des Starts neu), seit `slice-http-lebenszyklus` den Code `ODR-E5010` (503) einführt; der Code gehört in die Menge H, weil das Binary nach M2 ihn ausliefert), keine Liste von Konstanten wie im Vorbild; K hat als erste Zelle die Kennung `SPEC-nnn` und als zweite den Code. Der Architect legt die Leseregeln fest (welche Zeile ist ein Code, Form `^ODR-E[0-9]{4}$`, doppelt, Sperren bei fehlender Quelle, Ausgabe und Ausgang) und die Grenze (was nicht geprüft ist: Spalten Bedingung, Aktion, Bedeutung).
2. **Die Menge H — die eigentliche Entscheidung, je Katalog.** Das Handbuch sagt nur zu, was das Binary zeigt (Entscheidung des Users, Welle §1). Der Katalog der Fehlercodes führt aber Codes, die erst mit Folgewellen erreichbar sind (Admin-API: `ODR-E7005`, `ODR-E7006`, `ODR-E7012` …; Mock: `ODR-E6001` …; V2: `ODR-E6006`, ein Code, den nur ein V2-Dienst erreicht, Sendung von `slice-v2-query`, `AGENTS.md` §3.13: die Entscheidung des Architect zu H nennt ihn ausdrücklich, und sein Eintrag im Handbuch kommt mit `slice-v2-bootstrap`, dem Slice, der einem V2-Dienst den Weg ins Binary öffnet; ebenso `ODR-E6007`, Sendung von `slice-odata-json-eingabe`, `AGENTS.md` §3.13: der Code steht in `fehler.go`, bevor ihn ein Client erreicht, die Entscheidung zu H nennt ihn ausdrücklich, und sein Eintrag im Handbuch kommt mit `slice-bootstrap-m3`); die Spezifikation führt Schlüssel, die das Binary ausdrücklich ablehnt (`ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`: `nichtGelieferte` in `konfig_werte.go`). Optionen: (a) H führt alle Einträge der Quelle des Codes und sagt nur die Bedeutung nach dem Katalog — nach der Entscheidung des Users ausgeschlossen, soweit der Eintrag im Binary nicht erreichbar ist; (b) H führt genau die Einträge, die das gebaute Binary kennt, und der Code trägt ein Merkmal dafür (Fehlercodes: die Spalte *Ziel* in `fehler.go` ist ein Anfang; Schlüssel: die Menge, die `LadeKonfiguration` liest, minus `nichtGelieferte`); (c) Klassen statt Einträge, mit dem Verlust, dass ein einzelner Eintrag ungeprüft bleibt. Der Planner neigt zu (b) wegen des Ist-Zustands, sieht aber, dass (b) ein Merkmal im Code verlangt, das sich mit jeder Welle ändert, und dass die Schlüssel heute als Literale an vielen Stellen von `konfig*.go` stehen (keine Tabelle). Der Architect entscheidet mit Grund und nennt die Folge: ein Code-Umbau zu einer Schlüssel-Tabelle (dann Folge-Slice mit Kennung, `AGENTS.md` §3.13) oder eine Leseregel über die Literale (mit Grenze und Fänger über alle Stellen, an denen `p.wert`, `p.aufzaehlung`, `p.dauer` und die übrigen Leser einen Schlüssel nennen). Dasselbe für die Kommandos: Quelle ist der Aufruf-Zweig in `Run`; der Architect entscheidet, ob er maschinell lesbar ist oder eine Tabelle im Code verlangt. Folge für die Folgewellen: Jede Welle, die ein Kommando, einen Schlüssel oder einen Fehlercode erreichbar macht, zieht das Handbuch im selben Slice nach, sonst ist das Gate rot (das ist die Folge der Regel, und der Planner trägt sie als Hinweis an die Planer der Folgewellen).
3. **Code außerhalb der Tabelle.** Literale `"ODR-…"` stehen heute auch in `src/hexagon/model/konflikt.go`, `daten_fehler.go`, `redaktor.go` und `src/bootstrap/konfig.go`; das Vorbild verbietet sie außerhalb der Tabelle. Der Architect entscheidet, ob der Abgleich sie als Befund meldet (dann ist das ein Code-Umbau und gehört nicht in diesen Slice, sondern als Folge-Slice benannt) oder sie ausdrücklich zulässt (Grenze).
4. **Ort in der Kette und Laufzeit.** Docker-only (`AGENTS.md` §3.1): ein Programm in Go unter `tools/` in einer Stufe des `Dockerfile` oder ein Treiber mit Eingaben vom Host wie bei `SPEC-128`; der Architect wählt die Form, nennt Eingaben (Pfade von C, K, H) und was in den Kontext der Stufe kopiert wird. Hängt das Ziel an `GATE_CHECKS` und die Gegenprobe an die Liste der Gegenproben, wie `SPEC-102` es für andere Werkzeuge hält? Die Stufung folgt dem Vertrag der Gegenproben und [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md).
5. **Gegenproben.** Ein Fall je Punkt des Vertrags (Fehlform in einem Wegwerf-Baum): fehlender Eintrag in K, fehlender in H, Eintrag nur in H, Form, doppelt, Sperren, Ausgang — und das je Katalog (Fehlercodes, Schlüssel, Kommandos), denn die Kataloge sind eine Menge (`AGENTS.md` §3.15): ein Fall je Katalog, je Quelle und je Richtung der Differenz.
6. **Zeit.** Der Planner verlangt die Messung der Zusatzzeit von Gate und Gegenproben (Schritt `make gates` und `make gegenproben`, vorher und nachher, Zahlen im Plan). Die Frage, ob [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Trigger „Lauf über 30 s bei geänderter Eingabe“; gemessen 56,9 s im Schritt) neu zu fassen ist, entscheidet dieser Slice nicht (siehe §6).
7. **Form je Katalog.** Je Katalog Leseregeln und Befunde nach Nr. 1 (Zeile, Form des Namens, doppelt, Sperren): Fehlercodes `^ODR-E[0-9]{4}$`; Schlüssel in Großbuchstaben, Ziffern und Unterstrich; Kommandos als Wort. Die Richtungen der Mengen (Code ohne Handbuch, Handbuch ohne Code) gelten je Katalog, die Quelle K (Spezifikation) nur dort, wo sie führt (Fehlercodes `§4`, Schlüssel `§3` und `SPEC-110`); für Kommandos ist K seit `SPEC-138` Nr. 2 die Spezifikation (Menge `start` nur mit gesetztem `Start`, und `version`); der Architect entscheidet, ob die Kommando-Tabelle im Code steht.
8. **Wo die Optionen des Binary stehen.** Optionen (Flags) kennt das Binary heute nicht (`cli.Run` wertet nur das erste Argument aus); der Katalog der Kommandos führt, was der Adapter `cli` kennt, und die Optionen, sobald eine Welle sie liefert, nicht vorher. Der Architect hält fest, wie das Gate mit einem Kommando umgeht, das Optionen trägt (Zeile je Kommando und Option).
9. **Fänger für „Geplantes“ und „Chronik“: entfällt.** Der Fänger ist `make handbuch-check` (`SPEC-200`, geliefert von `slice-handbuch-geruest`, Geber; Nachzug beim Start, Planner, 2026-11-06). Die Tabellen dieses Slice im Handbuch sind Prosa im Sinn der Mengen A, B und C von `SPEC-200` Nr. 14: ein Satz wie „noch nicht unterstützt“ in einer Zelle ist rot, die Zellen tragen die Bedeutung nach dem Katalog in Worten, die keine Menge trifft (Rückgabe an den Planner, wenn eine Bedeutung der Spezifikation ein Wort der Mengen trifft).
10. **Nachzüge.** `harness/README.md` (Sensors-Zeile samt Bindung, höchstens 200 Zeichen in den Zellen, `structure` in `.d-check.yml`), `AGENTS.md` nur als Diff-Vorschlag an den User; Tabelle *Zusage · Fänger* (§3), Raster (§6), Größe.
11. **Codes, die nur der Handle erzeugt (Sendung von `slice-workspace-use-case`, `AGENTS.md` §3.13).** Ob `ODR-E7001`, `ODR-E7007` und weitere Codes, die allein der Composition Root erzeugt, in der Menge H stehen, entscheidet der Architect in Nr. 2 mit Grund je Code (Schleife über die Codes der Ziele `config` und `workspace`, nicht eine Liste der Fälle, die einfielen).
12. **Zweig `slice-odata-json-eingabe` läuft zugleich (Linie B).** Er führt `ODR-E6007` in `fehler.go` ein (36 → 37 Einträge); die Entscheidung zu H nennt den Code ausdrücklich (Nr. 2), der Eintrag im Handbuch kommt mit `slice-bootstrap-m3`. Der Abgleich liest `fehler.go` nur; ändert der Vertrag die Datei, geht das als Rückgabe an den Planner (Konflikt mit dem Zweig der Linie B). Welcher der beiden Zweige zuerst nach `main` geht, entscheidet der erste Verifier-Pass; der zweite nimmt `main` per Merge auf und zählt neu.
13. **Schnittfrage zuerst.** Die Rechnung in §4 liegt voraussichtlich über der Schnittgrenze (*S* ≥ 2223, `SPEC-128` Nr. 1); der Architect beantwortet die Frage vor Tabelle und Raster: Schnitt an einem Katalog ohne geteilten Mechanismus (Kommandos und Schlüssel als zweiter Slice nach den Fehlercodes) oder der Grund dagegen, und meldet den neuen Plan an den Planner.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Texte um die Tabellen (Aufgaben, Fehlerbilder mit Ursache und Lösung, Beispiele) — `slice-handbuch-record-replay`; dieser Slice bringt Abschnitte, Tabellen, Form und Gate und füllt die Tabellen mit den Einträgen, die das Binary nach M2 kennt, mit der Bedeutung nach dem Katalog.
- Kapitel zu Verhalten, das das Binary nicht kennt, und Hinweise „folgt“ — Entscheidung des Users (Welle §1); das Gate macht sie rot, statt sie zu dulden.
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

- [ ] Vertrag des Abgleichs in `harness/vertraege/` und das Programm samt Treiber und Ziel in `GATE_CHECKS`, für alle drei Kataloge (Fehlercodes, Konfigurationsschlüssel, Kommandos); je Katalog ein Abschnitt mit Tabelle im Handbuch; jede Zusage des Vertrags hat einen Fänger, vom Implementer rot gesehen (`LH-QA-05`, `LH-RB-01`).
- [ ] Die Regel „Jeder Fehlercode, jeder Konfigurationsschlüssel und jedes Kommando des Binary kommt im Handbuch vor, und das Handbuch nennt nichts, was das Binary nicht kennt“ ist durch Fänger über beide Richtungen und alle Elemente der drei Mengen gefangen (Schleife über die Mengen, `AGENTS.md` §3.15, §3.17).
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
- [ ] Sendung von `slice-odata-json-eingabe` (Geber, `AGENTS.md` §3.13): wer von beiden Zweigen als zweiter nach `main` geht, trägt `ODR-E6007` ein. Der Code steht in `fehler.go` und in §4 der Spezifikation (`SPEC-180`); das Binary zeigt ihn erst mit `slice-bootstrap-m3`. Deshalb steht er als Zeile `ODR-E6007 | <Grund>` in die Datei der Ausnahmen (harness/katalog-ausnahmen.txt), das Handbuch nennt ihn nicht (`SPEC-201` Nr. 3 c, Befund *Handbuch nennt Ausnahme*), und `make katalog-check` ist auf dem Stand nach dem Merge grün. Mit `slice-bootstrap-m3` wandert der Code aus den Ausnahmen in den Fehlercodes-Abschnitt des Handbuchs (`SPEC-201` Nr. 11).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| Vertrag unter `harness/vertraege/` | neu | Leseregeln, Mengen, Sperren, Ausgabe, Grenze (Auftrag Nr. 1 bis 3) |
| Programm unter `tools/`, Treiber, Stufe im `Dockerfile` | neu | Ort und Form nach Nr. 4; drei Kataloge (Fehlercodes, Schlüssel, Kommandos) |
| `src/bootstrap/` (Schlüssel- oder Kommando-Tabelle) | update, nur falls der Architect in Nr. 2 die Tabelle im Code verlangt | sonst Folge-Slice mit Kennung |
| Fragment unter `harness/mk/` | neu | Ziel und Anbindung an `GATE_CHECKS` und die Gegenproben |
| Tests des Programms | neu | ein Fänger je Zusage, Mutationen vom Implementer rot gesehen |
| Gegenproben (`tools/gegenproben/`) | update | ein Fall je Punkt des Vertrags (Nr. 5) |
| docs/user/benutzerhandbuch.md | update | Abschnitt der Fehlercodes mit Tabelle |
| `harness/README.md` | update | Sensors-Zeile mit Bindung |

Der Architect füllt die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und ergänzt die Zeilen oben; sie ist beim Start leer.

| Zusage | Fänger |
|---|---|
| Jeder Fehlercode des Binary steht im Handbuch, und das Handbuch nennt keinen anderen | — (Architect; Schleife über C und H) |
| Jeder Konfigurationsschlüssel des Binary steht im Handbuch, und das Handbuch nennt keinen anderen | — (Architect; Schleife über die Schlüssel und H) |
| Jedes Kommando des Binary steht im Handbuch, und das Handbuch nennt keines, das der Adapter `cli` nicht kennt | — (Architect; Schleife über die Kommandos und H) |
| Das Gate ist grün auf dem Stand des Binary und rot, wenn ein Eintrag fehlt oder überzählig ist | — (Architect; Gegenprobe je Katalog) |

**Betriebsregeln für die Rollen dieses Slice:** Dateien des Repos nur mit `Edit` und `Write`, kein `sed -i`, kein Heredoc auf Repo-Dateien (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`). Fixtures mit Backslash-u-Folge, BOM, NBSP oder Zeichen jenseits von ASCII entstehen aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18). Der Handoff ist ein Commit. Docker-Läufe mit `--memory` und wenige parallel. Änderungen an `AGENTS.md`, Commands, Rollen-Typen und Skills gehen als Diff-Vorschlag an den User.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-handbuch-geruest` und `slice-bootstrap-m2` liegen in `done/` (das Gerüst trägt den Ort der Abschnitte; erst nach M2 liest das Binary Schlüssel und erzeugt Fehlercodes, der Adapter `cli` (`cli.Run`) ruft heute keinen Start, also weder `LadeKonfiguration` noch `LadeOderMelde`); der Architect hat die Aufträge aus §1 beantwortet, die Tabelle in §3 und das Raster in §6 gefüllt, bevor der erste Code-Commit entsteht.

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7, Start, Planner, 2026-11-06): Dieser Slice ändert `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md` und `tools/gegenproben/` (Linie C) und ist damit der eine Besitzer der Gate-Kette. Stand beim Start: kein anderer Slice der Gate-Kette in `in-progress/` (dort liegen `slice-odata-json-eingabe`, Linie B, und dieser; der pausierte Zweig `adr-gate-laufzeit` ist kein Slice in `in-progress/`). Solange dieser Slice läuft, startet kein weiterer Slice der Gate-Kette (`slice-gate-integration-determinism`, `slice-guard-schreibwege`, die Slices von Teil 2 von `welle-sensoren`, `slice-worktree-parallel`); Kennungen aus dem Bereich der Linie C (`SPEC` 200 bis 229, `ADR` 0030 bis 0039, `CO` 040 bis 059, `MR` nur Linie C), nächste freie `SPEC`-Nummer: der Architect liest sie am Stand des Starts. Gate-Zeit der Sperre: der Slice fährt `make gates` und `make gegenproben` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Linien zugleich.

**Größenmarke** (`AGENTS.md` §3.14): Stand beim Start (Planner, 2026-11-06). Faktoren, Zuschlag, Aufrundung und Schnittgrenze stehen allein in `SPEC-128` Nr. 1 (einen Zuschlag für Leser fremder Bytes kennt sie nicht mehr). Der Block mit `basis`, `schaetzung` und `marke` steht auf `main` nicht (ein Block ohne Zahlen der Tabelle wäre ein Platzhalter und machte `make groessenmarke` rot); der Architect legt ihn im Zweig an, `basis` ist der letzte Commit, der Vertrag, Tabelle und Raster abschließt, und der Planner setzt Schätzung und Marke in eigenem Commit vor dem ersten Code-Commit. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen. **Vorläufige Rechnung ohne Tabelle, nur Richtwert, nicht bindend:** drei Kataloge mit je rund 10 Zeilen der Tabelle (30 Zeilen, keine nebenläufig, *a* = 0, *b* = 30), *f* = 0, *p* rund 700 (Programm, Treiber, Fragment, Gegenproben ohne Tests), also *S* = 80·30 + 700 = 3100 und *M* = 100 · ⌈135·3100 / 10000⌉ = 4200; die Schnittfrage stellt sich (*S* ≥ 2223, Auftrag Nr. 13). Die Aufwandsschätzung der Welle-Zeile steht von M auf L (drei Kataloge statt einem). Die bindende Rechnung steht erst mit der Tabelle des Architect.

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
- Die Schlüssel stehen als Literale an vielen Stellen von `konfig*.go`, nicht in einer Tabelle; ein Leser des Quelltexts übersieht einen Schlüssel, den ein neuer Leser einführt — **Ausgang:** offen; Auftrag Nr. 2 (Tabelle im Code als Folge-Slice oder Leseregel mit Grenze und Fänger über alle Leser).
- Das Binary kennt heute nur `version`; Schlüssel und Fehlercodes werden erst durch `slice-bootstrap-m2` im Binary erreichbar (Start-Trigger) — **Ausgang:** offen; entfällt, sobald der Slice in `done/` liegt und der Architect den Stand am Binary geprüft hat.
- Die Regel des Gates („nichts im Handbuch, was das Binary nicht kennt“) fängt Namen, nicht die Aussage über sie; ein falsch beschriebener Schlüssel bleibt grün — **Ausgang:** weiter offen; Prüfung von Bedeutung und Beispiel tragen die belegten Beispiele (`slice-handbuch-record-replay`) und der Probelauf (`slice-handbuch-abnahme`); akzeptiertes Negativ mit Grund und Folge im Vertrag (Grenze).
- Die Menge H (Auftrag Nr. 2) kollidiert mit dem Ist-Zustand, wenn H Codes führt, die das Binary noch nicht erzeugt — **Ausgang:** offen; Entscheidung des Architect vor dem Code.
- Literale `"ODR-…"` außerhalb der Tabelle (Auftrag Nr. 3) — **Ausgang:** offen; Entscheidung des Architect, bei Verbot ein Folge-Slice mit Kennung, den der Planner anlegt.
- Zählstelle der Codes: die Tabelle in `fehler.go` wächst mit dem Zweig `slice-odata-json-eingabe` (36 → 37), und der Plan, die Fänger und das Handbuch nennen eine Zahl; die Kennung `ODR-E6007` gehört in die Entscheidung zu H (Nr. 2, Nr. 12) — **Ausgang:** offen; Adresse: Architect (zählt am Stand des Starts), Integrationsschritt (zählt nach dem zweiten Merge neu, `AGENTS.md` §3.9).
- Nr. 11 (Codes, die nur der Handle erzeugt) und Nr. 12 (Zweig der Linie B) sind Sendungen und Rückgaben ohne Entscheidung — **Ausgang:** offen; Adresse: Architect.
- Schnittfrage (Nr. 13): *S* liegt voraussichtlich über 2223 — **Ausgang:** offen; Adresse: Architect (Schnitt oder Grund dagegen), Planner setzt die Marke nach der Tabelle.
- Der Abgleich liest Texte, die andere schreiben (Quelltext, Spezifikation, Handbuch): Eine Randform, die der Vertrag nicht nennt, wird im Review gefunden (`BEO-REPO/randform-erst-im-review-entschieden`, 20×, `verkörpert` in `AGENTS.md` §3.16) — **Ausgang:** offen; das Raster unten ist das Gerüst, die Zellen füllt der Architect vor dem ersten Code-Commit.
- Eine Allaussage des Vertrags („jeder Code in C steht in K“) ohne Fänger über alle Fälle (`BEO-REPO/universalaussage-im-vertrag-ungeprueft`, 9×, `verkörpert` in `AGENTS.md` §3.16) — **Ausgang:** offen; Fänger als Schleife über alle Codes der Tabelle.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Quelle C lesen | Architect | Architect | Architect | Architect | Architect |
| Quelle K lesen | Architect | Architect | Architect | Architect | Architect |
| Quelle H lesen | Architect | Architect | Architect | Architect | Architect |
| Schlüssel aus `konfig*.go` lesen (Katalog 2) | Architect | Architect | Architect | Architect | Architect |
| Kommandos aus `src/adapters/driving/cli` lesen (Katalog 3) | Architect | Architect | Architect | Architect | Architect |
| Mengen vergleichen (je Katalog) | Architect | Architect | Architect | gilt nicht: reiner Lesevorgang, Architect bestätigt | Architect |
| Befunde ausgeben | Architect | Architect | Architect | Architect | Architect |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Verstößen am selben Code:

| Paar | Entscheidung | Fänger |
|---|---|---|
| Code fehlt in K und in H | Architect | Architect |
| Code doppelt in H und fehlt in C | Architect | Architect |
| Form verletzt und Code doppelt | Architect | Architect |
| Schlüssel im Handbuch, den das Binary ablehnt (`nichtGelieferte`) und der in K steht | Architect | Architect |
| Kommando im Handbuch, das der Adapter `cli` nicht kennt, und Kommando im Adapter `cli` ohne Handbuch | Architect | Architect |
| Eintrag in zwei Katalogen mit gleichem Namen (ein Wort als Kommando und als Schlüssel) | Architect | Architect |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-06; Zähler = Zahl der Dateien in `evidence/`). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/randform-erst-im-review-entschieden` (33×, `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×, `AGENTS.md` §3.15), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.16), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (21×, `AGENTS.md` §3.9) für den Nachzug der Zahl der Codes, `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, `implement-slice` Schritt 14) für die Leser fremder Texte, `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×, `implement-slice` Schritt 14) und `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (19×, `geplant` in `slice-groessenmarke-werkzeug`) für die Marke. Offen mit Bezug: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (4×, `geplant` in `slice-coverage-gate`; der Slice schreibt Gegenproben, die Eingabe der Fälle sind Wegwerf-Bäume, nicht die Pakete des Produkts), `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (11×, `geplant`: Neukalibrierung der Faktoren, Träger noch kein Slice; die Marke folgt der Formel, ein Hinweis zur Überschätzung gehört in die Messung), `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (7×, `geplant` in `slice-guard-schreibwege`; Betriebsregeln in §3 gelten) und `BEO-REPO/schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices` (1×, für den Schnitt aus Nr. 13). Kein Eintrag erreicht mit diesem Slice ohne Ausgang 3×; die Zahlen stehen unter `evidence/` des jeweiligen Eintrags.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
