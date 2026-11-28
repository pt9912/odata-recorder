# Slice slice-ci-github-actions: CI mit GitHub Actions — `make gates` auf einem frischen Runner bei Push und Pull Request

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — ein einzelner Slice, dessen Closure-Bedingung seine DoD ist (ein grüner und ein roter Lauf auf dem Runner); kein repo-weiter Beleg geht darüber hinaus (Baseline-Regelwerk `modul-06-roadmap.md` §Wann Arbeit eine Welle braucht). Er trägt keinen Meilenstein: die Erstabnahme (M7) liegt vor seinem Start. Er ist ein **Kettenslice** (Linie C, Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)): er ändert mindestens `harness/README.md`.

**Bezug:** `LH-QA-06` (Messmethode: containerbasierter CI-Test), `LH-RB-03` (Entwicklung und Ausführung in Docker-basierten CI-Umgebungen), [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md) (Gegenproben außerhalb der Gate-Kette), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (`test-race` ist Gate).

**Berührte Spec-Stellen:** `LH-RB-03.a` (Linux und CI) · `SPEC-004` (Migrationsregel: Satz über CI und seinen Ersatz) · `SPEC-122` (`make test-race`) · [SPEC-211](../../../../harness/vertraege/SPEC-211-container-tests-in-der-kette.md) Nr. 5 e (Abbruch durch TERM, offene Sendung mit Trigger „eine CI“) · `SPEC-323` (Neustarttest, kommt mit [slice-container-image](../in-progress/slice-container-image.md)) · neuer Vertrag der CI unter `harness/vertraege/` (Kennung vergibt der Architect aus dem Bereich der Linie C).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner, Auftrag des Users vom 2026-11-28: planen; Umsetzung erst nach der Erstabnahme). **Datum:** 2026-11-28.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Bei Push und Pull Request fährt GitHub Actions `make gates` unverändert auf einem frischen, von GitHub gestellten Linux-Runner und meldet rot, wenn ein Gate rot ist; dieser Lauf ist der containerbasierte CI-Test, den die Messmethode von `LH-QA-06` verlangt und den heute ein Lauf über `make` ersetzt.

**Rahmen** (aus dem Auftrag und den harten Regeln, nicht zur Wahl):

- Docker-only (`AGENTS.md` §3.1): der Workflow installiert keine Toolchain und ruft `make`; die Prüfläufe bleiben netzlos und ohne schreibbaren Mount.
- Actions und Images sind gepinnt (Actions auf die volle Commit-SHA, Images wie heute auf den Digest).
- Keine Geheimnisse; die Rechte des Runner-Tokens sind minimal (nur lesen).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Gate-Stempel und der volle Lauf ohne Stempel (Slice (a) aus [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md)) — anderer Vorgang: die ADR ist `Proposed`, und der Slice ändert `tools/harness/working-tree-hash.sh`, wofür die Zustimmung des Users fehlt.
- Nachtlauf und Hash-Prüfung je Push (Slice (b) aus [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md)) — **vorläufig** außen: die ADR ordnet sie hinter die Stempel. Ob diese CI ohne beides genügt, ob sie einen Teil von (b) vorwegnimmt oder ob (b) hierher gehört, entscheidet Frage 2 in §6; der Planner entscheidet es nicht. Ändert die Antwort diesen Ausschluss, wird der Plan vor dem Start geändert und neu geschnitten.
- Gegenproben, Lasttest und Mutationsläufe auf dem Runner — **vorläufig** außen: keiner ist Teil von `make gates` ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), `SPEC-216`); Frage 6 in §6.
- Änderungen an Gates, Schwellen, Fristen, Tests oder Produktcode, damit der Runner grün wird — eine Lockerung braucht eine ADR (`AGENTS.md` §3.6); ein Fund ist eine Rückführung (§4) oder ein Carveout, kein Beifang.
- Änderungen am Lastenheft — Sache des fachlichen Owners; nach `Accepted` nur über einen Change Request ([Lastenheft](../../../../spec/lastenheft.md) §Historie). §6 stellt die Frage.
- Registry, Veröffentlichung, Release, Ablage von Artefakten — anderer Vorgang; sie brauchen Schreibrechte und Geheimnisse, die dieser Slice ausschließt.
- Einstellungen des Repos auf GitHub (Branch-Protection, Pflicht-Status) — liegen nicht im Repo; Entscheidung des Users (§6).
- Ersatz des lokalen Gate-Laufs vor dem Handoff — `AGENTS.md` §6 und der Stop-Hook bleiben, wie sie sind; die CI kommt hinzu.
- Der Arbeitscontainer für Entwickler in der CI — [welle-dev-container](../welle-dev-container.md); sie schließt ihrerseits eine CI aus (§6 dort).

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

Entwurf; der Architect schärft die drei Liefer-Punkte mit dem Vertrag vor dem ersten Code-Commit.

- [ ] **Lauf** (`LH-QA-06`, `LH-RB-03`): der Workflow fährt `make gates` bei Push und Pull Request auf einem frischen Runner, ohne Toolchain-Installation und ohne ausgelassenes Gate; ein grüner Lauf ist belegt mit Commit, Adresse des Laufs, Runner-Label, Dauer und Messendem, darin der Neustarttest (`test-image`, `SPEC-323`) und `test-race`.
- [ ] **Rot ist sichtbar** (`AGENTS.md` §3.10, §3.15): je Zusage und Randform des Vertrags ein Fänger; mindestens ein auf dem Runner gesehener Lauf mit absichtlich gebrochenem Gate ist rot, aus dem benannten Grund, und der gebrochene Stand gelangt nicht nach `main`.
- [ ] **Pins und Rechte:** jede Action auf der vollen Commit-SHA, Runner-Label ohne `latest`, Rechte des Tokens nur lesend, kein Geheimnis, kein Token im Arbeitsbaum nach dem Checkout, kein Auslöser, der fremdem Code Schreibrechte oder Geheimnisse gibt; je Zusage ein Träger (Prüfung oder benanntes Negativ mit Grund).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: jede Stelle, die „keine CI“ oder den Ersatz durch `make` trägt, zieht im liefernden Commit nach (`AGENTS.md` §3.9; Liste in §3).
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
| `.github/workflows/gates.yml` (Name ist Vorschlag) | neu | der Workflow; es gibt heute kein Verzeichnis `.github/` |
| Vertrag der CI unter `harness/vertraege/`, samt Zeile im Index dort | neu | Auslöser, Checkout, Rechte, Pins, Randformen (§6), Tabelle *Zusage · Fänger*; Architect, vor dem ersten Code-Commit (`AGENTS.md` §3.12) |
| Fänger der Zusagen des Workflows (Ort und Form offen) | neu | Frage 12 in §6; heute liest kein Gate die YAML-Datei außer `make zeichen-check` |
| `harness/README.md` | update | Zeile „eine CI gibt es noch nicht“, §Sensors „Nicht behauptet“ zur Verdrahtung der CI, Safety-Grenzen |
| `harness/vertraege/` (`SPEC-211` Nr. 5 e, `SPEC-323` Festlegung und Nr. 1) | update | beide begründen mit „ohne CI“ |
| `spec/spezifikation.md` (`SPEC-004` Migrationsregel, gegebenenfalls `LH-RB-03.a`) | update | der Satz „Solange es keine CI gibt“; nach Antwort auf Frage 13 (Architect, Historie der Spezifikation) |
| `Makefile`, `harness/mk/`, `Dockerfile` | nur nach Entscheidung | nur, wenn der Vertrag es nennt (etwa Parallelität, Cache); sonst unberührt |

Weitere Stellen, die „keine CI“ nennen und nur der Planner im Integrationsschritt ändert (Übergangsregeln Nr. 4): [welle-dev-container](../welle-dev-container.md) §6, [welle-betrieb](../welle-betrieb.md) §2 (historische Bedingung, falls die Welle dann noch offen ist). `docs/plan/planning/rueckverfolgung.md` ist ein datierter Auslesestand. Dateien in `done/` bleiben. [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md) nennt „eine CI gibt es nicht“ im Kontext; solange sie `Proposed` ist, führt sie der Architect (Frage 2).
Die Liste ist der Stand vom 2026-11-28; der Planner sucht beim Start neu.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`), alle Bedingungen zugleich:

1. Der Kopf von `spec/lastenheft.md` trägt den Status `Accepted` (Erstabnahme; Auftrag des Users vom 2026-11-28). Heute: `Draft`.
2. [slice-container-image](../in-progress/slice-container-image.md) liegt in `done/` (Vorgänger in der Gate-Kette: erst mit ihm trägt `make gates` den Neustarttest, den die Messmethode von `LH-QA-06` verlangt).
3. Kein anderer Kettenslice liegt in `in-progress/`. Die Stellung zu den übrigen offenen Kettenslices legt der Start fest.
4. Die Fragen 1 bis 13 in §6 sind beantwortet; eine Antwort, die den Schnitt ändert, ändert den Plan vor dem Start. Vertrag, Raster und Tabelle *Zusage · Fänger* stehen vor dem ersten Code-Commit.
5. Der User hat zugestimmt, dass Zweige dieses Slice nach GitHub gepusht werden: ohne Push gibt es keinen Lauf und keinen Beleg. Gepusht wird in diesem Repo nur auf Anweisung.

**Größenmarke:** noch nicht gesetzt, ohne Block; der Planner setzt sie beim Start, vor dem ersten Code-Commit (`AGENTS.md` §3.14, `SPEC-128` Nr. 1), und prüft dabei den Stand dieses Plans gegen das Repo (Zahl der Gates, Laufzeit, Stellen aus §3).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): grün wird der Runner nur mit Änderungen an Produktcode, Tests, Fristen oder Gate-Rezepten über das hinaus, was der Vertrag nennt — diese Änderung ist ein eigener Slice; oder die Antwort auf Frage 2 zieht Stempel oder Nachtlauf herein. Für die Marke gilt `SPEC-128` Nr. 10 c.
- `in-progress` → `open` (blockiert — Carveout?): ein Gate ist nur auf dem Runner rot (Carveout mit Trigger und Folge-Slice, nie ein stilles Rot, ein übersprungenes Gate oder ein automatischer Wiederholungslauf); oder eine Entscheidung des Users steht aus (Push, Kosten, Zugriff auf Images).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, darin ein grüner und ein absichtlich roter Lauf auf GitHub am Stand des Slice-Zweigs, §7 nennt beide mit Commit und Adresse; Review und Verifikation ohne offenen Befund über LOW; Closure-Notiz mit Lerneintrag.
Den ersten Lauf auf `main` trägt der Integrationsschritt nach dem Merge und dem Push nach; ist er rot, ist das ein Befund mit Carveout oder Folge-Slice, kein Nachtrag im Stillen.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

### Fragen an den Architect (alle offen)

1. **ADR.** Braucht die CI eine ADR (neuer externer Dienst, zweiter Ort der Gate-Kette, Pins und Rechte als dauerhafte Setzung), oder genügt der Vertrag unter `harness/vertraege/`? — **Ausgang:** offen, Adresse Architect.
2. **Verhältnis zu [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md)** (`Proposed`). Sie sieht „CI mit Nachtlauf und Hash-Prüfung“ als Folge-Slice (b) nach den Gate-Stempeln (a) vor. Braucht diese CI Stempel oder Nachtlauf, oder nicht? Ist dieser Slice (b), ein Vorläufer davon, oder davon unabhängig, und muss der Entwurf angepasst werden? — **Ausgang:** offen, Adresse Architect; die Annahme der ADR steht beim User.
3. **Laufzeit und Aufbau des Jobs.** Ein voller Lauf dauert lokal rund 20 Minuten (Angabe des Auftrags; belegt zum Beispiel 1033 s frisch, [slice-admin-inspektion](../done/slice-admin-inspektion.md) §7, Host mit 20 Kernen); ein Runner mit weniger Kernen braucht eher länger, gemessen ist das nicht. Ein Job oder mehrere, `make` mit oder ohne `-j` (Speicher: die Container-Tests laufen dann zugleich), welche Zeitgrenze des Jobs? Bleibt bei einer Aufteilung sichtbar, dass kein Gate aus `GATE_CHECKS` fehlt? — **Ausgang:** offen, Adresse Architect.
4. **Docker des Runners.** `test-image` startet Container im Daemon des Hosts (`SPEC-323`), der hier der Daemon des Runners ist; welche Voraussetzungen setzen die Rezepte (Plattform-Argumente, Bau-Optionen)? Das Image des Runners und sein Docker sind nicht auf einen Digest pinnbar: benanntes Negativ oder anderer Weg? — **Ausgang:** offen, Adresse Architect.
5. **Netz.** Netz brauchen die Stufe `deps` (Module) und der Bezug der Images (Docker Hub und `ghcr.io`). Wie verhält sich der Lauf bei Bezugsgrenzen oder Ausfall? Sind alle gepinnten Images ohne Anmeldung beziehbar (nicht geprüft)? Wäre eine Anmeldung nötig, steht sie gegen „keine Geheimnisse“ und geht an den User. — **Ausgang:** offen, Adresse Architect.
6. **Gegenproben.** Fährt die CI `make gegenproben` (je Push, geplant, gar nicht)? Die Fälle reißen ihre Frist von 120 s unter der eigenen Last schon auf dem Entwicklungs-Host (`BEO-REPO/gegenproben-erzeugen-die-last-selbst`; [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md)). — **Ausgang:** offen, Adresse Architect.
7. **Paketlaufzeit unter dem Detektor.** `test-race` hat zehn Minuten je Paket; `src/bootstrap` brauchte zuletzt 415,9 s bei 4 CPUs (`BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze`, ein Lauf). Reicht ein kleiner Runner? Wenn nicht: Welcher Weg ist ohne Lockerung offen? — **Ausgang:** offen, Adresse Architect; ein größerer Runner kostet und geht an den User.
8. **Flackern unter Last.** Tests und Stufen mit Frist werden unter Last rot und im nächsten Lauf grün (`BEO-REPO/test-fremden-pakets-flackert-unter-last`). Gilt jedes Rot als Befund, und wer darf neu starten? — **Ausgang:** offen, Adresse Architect.
9. **Cache.** Läuft jeder Lauf ohne Cache der Docker-Schichten, oder wird er gehalten? Mit Cache kommen `test` und `test-race` bei gleichem Inhalt aus dem Cache und sind kein neuer Lauf (`BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache`). — **Ausgang:** offen, Adresse Architect.
10. **Checkout und geprüfter Stand.** `make groessenmarke` braucht die Historie bis zur `basis` eines Plans in Arbeit (`SPEC-128` Nr. 7); ein flacher Klon endet dort rot. Bei Pull Requests: prüft der Lauf den Kopf des Zweigs oder den Merge mit dem Ziel, und was misst die Größenmarke dann? — **Ausgang:** offen, Adresse Architect.
11. **Auslöser, Doppel- und Abbruchläufe.** Welche Zweige bei Push; Pull Requests aus Forks; Push und Pull Request desselben Standes; Abbruch eines überholten Laufs. Übernommene Sendung (`AGENTS.md` §3.13): [SPEC-211](../../../../harness/vertraege/SPEC-211-container-tests-in-der-kette.md) Nr. 5 e lässt offen, ob Rezept oder Treiber auf TERM allein an `make` reagieren, Adresse Planner, Trigger „eine CI oder ein `timeout` um `make`“. Dieser Slice nimmt sie an: der Vertrag entscheidet, welches Signal Abbruch und Zeitgrenze des Jobs senden, und ob eine Änderung folgt oder das Negativ mit neuem Grund bleibt. — **Ausgang:** offen, Adresse Architect.
12. **Fänger des Workflows.** Womit wird rot, wenn Pin, Rechte, Auslöser oder der Aufruf von `make gates` sich ändern? Der Workflow läuft nicht lokal; ein Beweis auf dem Runner braucht Push. d-check führt ein Modul `workflows`, das hier nicht aktiv ist (Eignung nicht geprüft). Wer hebt die Pins der Actions an, und wann? — **Ausgang:** offen, Adresse Architect.
13. **Texte der Spezifikation.** `SPEC-004` sagt, die CI muss auch die Migrationskette auf einer Kopie bestehender Datenbanken fahren; mit einer CI entfällt der Ersatz durch `make schema-check` dem Wortlaut nach, die Kette gibt es aber erst mit einer zweiten Schemaversion. Was sagt die Spezifikation dann? — **Ausgang:** offen, Adresse Architect; der Ersatz war eine Entscheidung des fachlichen Owners.

### Fragen an den fachlichen Owner und den User

- **Lastenheft, `LH-QA-06`.** Der Satz „Solange es keine CI gibt, erkennt der fachliche Owner einen Lauf über `make` als Ersatz für den CI-Test an“ verliert mit der CI seine Wirkung. Zu entscheiden: bleibt er stehen, wird er gestrichen, oder gilt der Lauf über `make` weiter, wenn die CI ausfällt oder nicht läuft? Eine Textänderung nach `Accepted` braucht einen Change Request; vor `Accepted` wäre sie ohne ihn möglich. Der Planner ändert nichts. — **Ausgang:** offen, Adresse fachlicher Owner.
- **GitHub.** Zustimmung zum Push (§4 Nr. 5); Sichtbarkeit des Repos, Größe des Runners und Kosten oder Kontingent der Laufminuten (nicht geprüft); ob ein roter Lauf das Mergen sperrt (Branch-Protection). — **Ausgang:** offen, Adresse User.

### Randformen (Vorentwurf des Rasters, `AGENTS.md` §3.16)

Die Zellen nennen die Form und die Frage, die sie entscheidet; Ort der Entscheidung wird der Vertrag der CI. Vor dem ersten Code-Commit ersetzt der Architect jeden Verweis auf eine Frage durch die Nummer im Vertrag.

| Operation | fehlend oder leer | unbekannt oder fremd | Abbruch | Wiederholung |
|---|---|---|---|---|
| Auslöser | Push ohne Änderung an Prüfeingaben (reine Doku, Tag, gelöschter Zweig): Frage 11 | Pull Request aus einem Fork: Frage 11 | gilt nicht: vor dem Start gibt es keinen Lauf | Push und Pull Request desselben Standes: Frage 11 |
| Checkout | Historie ohne `basis`: Frage 10 | Merge-Stand eines Pull Request statt Zweigkopf: Frage 10 | Fehler des Checkout ist rot: Frage 12 | gilt nicht: ein Checkout je Lauf |
| Bezug von Images und Modulen | Image oder Modul nicht beziehbar: Frage 5 | anderer Inhalt hinter einem Pin: Images stehen auf Digest, Module prüft die Stufe `deps` gegen `go.sum` (Bestand); für Actions Frage 12 | Netzfehler mitten im Bezug: Frage 5 | Bezugsgrenze bei vielen Läufen: Frage 5 |
| Lauf `make gates` | ein Gate fehlt im Lauf: Frage 3 und 12 | Docker oder Runner mit anderem Verhalten: Frage 4 | Zeitgrenze oder Abbruch des Jobs: Frage 11 | neuer Start nach Rot: Frage 8 |
| Ergebnis | kein Lauf zu einem Commit: Frage 12 und Branch-Protection (User) | gilt nicht: das Ergebnis ist der Status des Jobs, ein anderes liest niemand | abgebrochener Lauf ist nicht grün: Frage 11 | ein grüner zweiter Lauf überdeckt einen roten ersten: Frage 8 |

### Risiken

- Ein Gate ist lokal grün und auf dem Runner rot (weniger Kerne, Fristen von 120 s je Stufe, Paketlaufzeit unter `-race`). — **Ausgang:** offen, Adresse Architect (Fragen 3, 7, 8); tritt es ein, Carveout oder Folge-Slice.
- Der Speicherplatz des Runners reicht für Images und Zwischenstände nicht (nicht gemessen). — **Ausgang:** offen, Adresse Implementer: die erste Messung auf dem Runner nennt den Stand.
- Ein roter Lauf auf `main` bleibt unbemerkt, weil nichts ihn liest und nichts am Merge hängt (ein Trigger ohne Wächter). — **Ausgang:** offen, Adresse User (Branch-Protection) und Planner (Pflicht im Integrationsschritt).
- Die CI hängt an einem Dienst außerhalb des Repos; fällt er aus oder ändert er das Runner-Image, ist das Ergebnis ohne Änderung am Stand anders. — **Ausgang:** offen, Adresse Architect (Fragen 1, 4) und fachlicher Owner (`LH-QA-06`).
- Zwischen Plan und Start liegt die Erstabnahme; Gates, Laufzeiten und Textstellen ändern sich bis dahin. — **Ausgang:** offen, Adresse Planner beim Start (§4).

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

Noch nicht begonnen.

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

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`).

**Vorgelagert — offene Beobachtungen sichten** (Stand 2026-11-28, Zahl der Dateien unter `evidence/`; beim Start neu zu sichten):

- `BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (3): Frage 7.
- `BEO-REPO/test-fremden-pakets-flackert-unter-last` (4): Frage 8.
- `BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache` (20): Frage 9.
- `BEO-REPO/gegenproben-erzeugen-die-last-selbst` (1): Frage 6.

Mit der Planung fällt kein Beleg an, kein Eintrag erreicht durch sie 3×; die Einträge ab 3 ohne Zielort liest die Welle-Closure.

**Modus-Begründungsblock:** entfällt, alle berührten Sub-Areas GF.
