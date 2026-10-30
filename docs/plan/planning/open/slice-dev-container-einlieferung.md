# Slice slice-dev-container-einlieferung: Bare-Repo, Klon und Einlieferung per Zweig

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-dev-container](../welle-dev-container.md).

**Bezug:** `LH-RB-03` (Entwicklung in Docker-Umgebungen), `LH-QA-06` (gleiche Eingabe, gleiches Verhalten: der Container sieht genau den eingelieferten Commit), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (der Stempel reist mit dem Commit), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md) (Vertrag unter `harness/vertraege/`).

**Berührte Spec-Stellen:** der Vertrag des Gerüsts und des Zugangs aus den Vorgängern (Probe, falls die Einlieferung eine Bedingung hinzufügt) · ein neuer Vertrag oder Abschnitt für Bare-Repo und Einlieferung (Nummer vergibt der Architect) · `SPEC-102` (Eingaben und Stempel, nur falls der Architect sie berührt).
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

**Ziel:** Im Container liegt ein Bare-Repo; ein Klon außerhalb des Containers (auch ein Arbeitsbaum, `git worktree`) liefert einen Zweig per SSH ein, und im Container entsteht daraus ein Arbeitsstand, der genau den eingelieferten Commit trägt — ohne ungetrackte oder fremde Dateien.

**Auftrag an den Architect** (Entscheidungen vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Planner entscheidet sie nicht):

1. **Bare-Repo**: Ort (Zustand, der den Neustart überlebt, oder ersetzbar), Anlage beim Start (idempotent), Eigentümer und Rechte, Name.
2. **Einlieferung**: Namensschema der Zweige (je Rolle, je Arbeitsbaum, je Slice), was erlaubt ist (neuer Zweig, Fortschreiben, Überschreiben, Löschen, `main`) und was abgewiesen wird mit welcher Meldung; ob ein Hook im Bare-Repo prüft oder nur eine Konvention gilt; ob der Eingang einen Lauf auslöst (Entscheidung, die `slice-dev-container-lauf` trägt).
3. **Arbeitsstand im Container**: Klon oder Arbeitsbaum je Zweig, Anlage, Aufräumen, Wiederholung auf demselben Zweig; Garantie „genau der Commit, sonst nichts“ (Verhältnis zu `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`: ein ungetrackter Rest verfälscht die Prüfsumme der Gegenproben).
4. **Rückweg**: wie Ergebnisse (Berichte unter `docs/reviews/`, ein Stempel) zum Entwickler kommen — als Zweig im Bare-Repo, den der Klon abholt, oder gar nicht; der Container hat keinen Zugang zu einem fremden Remote und keine Zugangsdaten (Grenze, kein Push nach außen).
5. **Raster und Tabelle**: Raster *Operation × Argumentart* in §6 (Operationen Einliefern, Abholen, Arbeitsstand anlegen; Argumente Zweigname fehlend, leer, unbekannt, ungültig — Leerzeichen, `..`, führender Bindestrich, Sonderzeichen —, Zweig schon vorhanden, Zweig gleich dem Stand, Abbruch mitten im Push, Wiederholung, zwei Einlieferungen zugleich, derselbe Zweig zweimal zugleich) mit Ort je Zelle, und die Tabelle *Zusage · Fänger* in §3 vor dem Übergang `next` → `in-progress`; Zusagen über Zweignamen als Schleife über alle Formen (`AGENTS.md` §3.15).
6. **Prüfbarkeit**: Fälle der Gegenproben mit dem echten `git` des Images in Wegwerf-Repos (Muster der Fälle 115 bis 120 in `SPEC-102`) und was nur ein Lauf auf `u-boot` belegt (akzeptiertes Negativ mit Grund und Folge).
7. **Größe**: Schätzung und Marke wie in `slice-dev-container-geruest` Auftrag 9.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Lauf von `make gates` und `make gegenproben` — `slice-dev-container-lauf`; dieser Slice liefert den Stand, nicht die Prüfung.
- Merge, Rebase und Review im Container — der Container ist ein Ort für Läufe, kein zweiter Hauptbaum; der Entwickler merged im eigenen Klon (Welle §6).
- Ein Push des Containers in ein fremdes Remote (GitHub oder Ähnliches) — Grenze: der Container trägt keine Zugangsdaten nach außen; der Rückweg endet im Bare-Repo (Auftrag 4).
- Große Dateien, Submodule, Hooks des Entwicklers — das Repo hat keine; ein Fall wäre ein Befund.
- Schutz von `main` über die Abweisung hinaus (Berechtigungen, Reviews) — keine Plattform im Container; die Abweisung ist die ganze Zusage.

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

- [ ] Die Entscheidungen 1 bis 4 des Architect stehen vor dem Code im Vertrag unter `harness/vertraege/`.
- [ ] Ein Zweig wird von außen per SSH eingeliefert; Bare-Repo und Arbeitsstand im Container tragen denselben Commit (gleicher Hash, Baum ohne ungetrackte Dateien); jede abgewiesene Form des Rasters (§6) endet mit der Meldung des Vertrags und ändert das Bare-Repo nicht.
- [ ] Zu jeder Zusage ein Fänger in §3 (Zweignamen als Schleife über alle Formen), `make gegenproben` gefahren; der Rückweg arbeitet nach Entscheidung 4.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: Vertrag und `harness/README.md` im selben Commit (`AGENTS.md` §3.9).
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
| `harness/vertraege/` (Vertrag der Einlieferung) | neu | Zusagen von Bare-Repo, Einlieferung, Arbeitsstand, Rückweg |
| Image, Start-Skript, Hook im Bare-Repo (Orte nach dem Gerüst) | update / neu | Anlage des Bare-Repos, Abweisung nach Entscheidung 2, Arbeitsstand |
| `tools/gegenproben/` und `SPEC-102` | update | Fälle in Wegwerf-Repos mit echtem `git` (Auftrag 6) |
| `harness/README.md` | update | Zeile des Ziels, „Nicht behauptet“ für den Teil, den nur `u-boot` belegt |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) entsteht beim Architect **vor** dem Übergang `next` → `in-progress`; sie ist noch leer.

Betriebsregeln für alle Rollen dieses Slice: Repo-Dateien nur mit `Edit` und `Write` (kein `sed -i`, `awk`, Python, Heredoc in die Datei; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`); jeder `docker run` einer Rolle mit `--memory` und `--memory-swap`, höchstens zwei Container zugleich; Zweignamen mit Sonderzeichen oder Rohbytes aus Bytewerten gebaut und gegengelesen (`AGENTS.md` §3.18).


## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-dev-container-zugang` liegt in `done/`; Vertrag, Tabelle *Zusage · Fänger* und Raster aus §6 stehen vor dem ersten Code-Commit (Architect, `AGENTS.md` §3.12). Danach setzt der Planner die Marke.

**Größenmarke** (`AGENTS.md` §3.14): noch nicht gesetzt. Der Planner setzt sie nach dem Architect auf die Schätzung × 1,25, aufgerundet auf volle Hundert, mit `make groessenmarke` (`SPEC-128`), falls `slice-groessenmarke-werkzeug` dann in `done/` liegt, sonst nach der Form von `AGENTS.md` §3.14.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke; offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten) und die Grenze 3000 vor dem Code.
- `in-progress` → `open` (blockiert): der Zugang des Vorgängers trägt keine Git-Befehle (Befund an den Planner; der Vorgänger liegt dann in `done/`, keine Adresse, `AGENTS.md` §3.13).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, je Zusage des Vertrags ist eine Mutation rot gesehen (`AGENTS.md` §3.10), und eine Einlieferung von außen auf `u-boot` ist mit dem Hash auf beiden Seiten belegt; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Ein Arbeitsstand mit ungetrackten Dateien oder aus einem früheren Lauf verfälscht die Prüfsumme der Gegenproben (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`, 1×) — **Ausgang:** offen; Fänger: der Arbeitsstand ist nach der Anlage rein (`git status` leer, Hash gleich), eine Mutation legt eine ungetrackte Datei ab.
- Gleichzeitige Einlieferungen zweier Zweige oder desselben Zweigs zweimal — **Ausgang:** offen; Zeile im Raster, Wirkungs-Fänger, der den Zustand des Bare-Repos zählt (kein Datenrennen im Go-Sinn, `make test-race` fängt hier nichts; akzeptiertes Negativ nur mit Grund und Folge, `AGENTS.md` §3.10).
- Die Abweisung von `main` wird umgangen (Umbenennung, `refs/…`-Formen, löschen) — **Ausgang:** offen; die Klasse „jede Ref-Form, die `main` trifft“ ist eine Menge, die der Vertrag zählt, der Fänger läuft als Schleife über sie (`AGENTS.md` §3.17).
- Schlüssel- oder Rechteprobleme des Zugangs zeigen sich erst bei Git — **Ausgang:** offen; sind sie ein Fehler des Vorgängers, ist es ein Befund an den Planner (`AGENTS.md` §3.13).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* als Gerüst; jede Zelle nennt die Frage, der Ort ist der Vertrag des Architect; eine Zelle ohne Ort hält den Implementer an — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Einliefern | Zweigname fehlt: (Vertrag, offen) | leerer Name, leerer Push: (Vertrag, offen) | ungültiger Name (Leerzeichen, `..`, führender Bindestrich), `main`, fremde Ref-Form: abgewiesen (Vertrag, offen) | Abbruch mitten im Push: Bare-Repo unverändert? (Vertrag, offen) | gleicher Stand erneut, Überschreiben, Löschen: (Vertrag, offen) |
| Abholen (Rückweg) | Zweig nicht vorhanden: (Vertrag, offen) | Zweig ohne Commit: (Vertrag, offen) | Zweig eines anderen Entwicklers: (Vertrag, offen) | Abbruch beim Abholen: (Vertrag, offen) | Wiederholung: (Vertrag, offen) |
| Arbeitsstand anlegen | Zweig nicht eingeliefert: Meldung (Vertrag, offen) | leerer Zweig: (Vertrag, offen) | Zweig mit abweichender Historie: (Vertrag, offen) | Abbruch beim Anlegen: halber Stand aufgeräumt? (Vertrag, offen) | Zweig schon ausgecheckt: (Vertrag, offen) |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen:

| Paar | Entscheidung | Fänger |
|---|---|---|
| gültiger Name und Rechte des Bare-Repos falsch | offen (Vertrag) | offen |
| zwei Einlieferungen am selben Zweig, eine abgewiesen | offen (Vertrag) | offen |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle 19 Einträge, Stand 2026-10-30; keiner steht bei 3× und offen). Für diesen Slice zählen: `BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand` (1×, `offen`) — der Arbeitsstand ist genau der Commit (Risiko oben); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×, `offen`) — Betriebsregeln in §3.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
