# Slice slice-bootstrap-m4-konfig: Konfiguration des Modus hybrid (Schlüssel `HYBRID_*` und `FALLBACK_*`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` · `SPEC-110` Nr. 2, 8, 9 (Schlüssel dieser Version, Modus, Services) · `SPEC-010`, `SPEC-062`, `SPEC-076`, `SPEC-079`, `SPEC-080` · `SPEC-055` (`PUBLIC_BASE_URL` in `hybrid`) · `SPEC-261` Nr. 4 e, 8, 10 b (Fehler der Policy).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-bootstrap-m4)

**Autor:** pt9912. **Datum:** 2026-11-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Startkonfiguration liest und prüft die Schlüssel `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` samt ihren Zusammenhängen (`SPEC-110`), macht jeden `*KonfigFehler` der Policy zu `ODR-E2001`, verlangt bei `HYBRID_FALLBACK=enabled` die `UPSTREAM_URL` jedes Service und bietet die Funktion, die aus den gelesenen Werten die Policy baut; `hybrid` ist ein gültiger Wert von `MODE`.

**Übernimmt** (Teil 1 des Plans `slice-bootstrap-m4`, der vor dem Code geschnitten ist, §4): den Nachzug von `SPEC-110` Nr. 2 und 8 für die Schlüssel `HYBRID_*` und `FALLBACK_*` und den Wert `hybrid` von `MODE`, den `slice-bootstrap-m3` für den Modus `mock` hinterlässt (dessen §1 *Ausdrücklich NICHT* weist beides `slice-bootstrap-m4` zu); das Lesen der Policy-Werte aus der Umgebung (Sendung von `slice-fallback-policy`, dessen §1 *Ausdrücklich NICHT* `slice-bootstrap-m4` nennt); aus dem Vertrag der Policy (`SPEC-261` Nr. 4 e, Nr. 8 c, Nr. 10 b; Sendung von `slice-fallback-policy-entscheidung`, dessen §1 *Ausdrücklich NICHT* `slice-bootstrap-m4` nennt): ein `*KonfigFehler` der Policy wird `ODR-E2001` (die Reihenfolge der Schlüssel in der Meldung entscheidet der Architect, §6 (b)) und `HYBRID_FALLBACK=enabled` ohne `UPSTREAM_URL` wird `ODR-E2001`; die Zeilen der vier Schlüssel in der Tabelle der Einstellungen des Handbuchs, ohne die der Katalog der Schlüssel (`SPEC-201` Nr. 4) rot wird, sobald `konfig*.go` sie nennt (Folge dieses Schnitts, §6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Bindung von `HybridUseCase` im Composition Root, `hybrid.Konfig`, der M4-Test, die Stufe 10 und das Handbuch-Kapitel — `slice-bootstrap-m4` (Teil 2): Der Modus `hybrid` startet in diesem Slice nicht; `vorbereiten` in `src/bootstrap/wurzel.go` lehnt ihn weiter ab, Teil 2 ruft die Funktion, die hier die Policy baut, und schreibt `konfig*.go` nicht um.
- Neue Fachlogik — der Slice verdrahtet nur die Schlüssel mit dem, was `slice-fallback-policy` und `slice-fallback-policy-entscheidung` liefern.
- Die Schlüssel `ADMIN_*` und `LISTEN_TLS_*` — `welle-admin-api` und `welle-betrieb`; sie bleiben `ODR-E2001`, damit keine Sicherheitseinstellung still wirkungslos ist (`SPEC-110` Nr. 2).
- Das Umschalten des Modus zur Laufzeit — `LH-FA-13` schließt es aus dem ersten Abnahmeumfang aus.
- `LINK_REWRITE` — bleibt `off` (`SPEC-092`).

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Lesen und Prüfen: die vier Schlüssel kommen aus der Umgebung in die Konfiguration (Syntax, Leerraum, Doppelte und Randformen nach `SPEC-110` Nr. 1, 3, 6); ein `*KonfigFehler` der Policy wird `ODR-E2001`, ohne den Wert in der Meldung; `UPSTREAM_URL` ist bei `HYBRID_FALLBACK=enabled` je Service Pflicht und fehlt sie, ist es `ODR-E2001`; die Funktion, die aus den gelesenen Werten `policy.New` ruft und die Werte weiterreicht, steht (Fänger, dass sie den Nullwert nicht stehen lässt; Muster aus `slice-upstream-tls`).
- [ ] Spezifikation und Modus: `hybrid` ist ein gültiger Wert von `MODE` und die Konfiguration meldet dafür keinen Verstoß; die Zusammenhänge der Schlüssel zu den Modi und `PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid` gelten nach dem Nachzug von `SPEC-110` Nr. 2 und 8 und 9 (Architect); `ADMIN_*` und `LISTEN_TLS_*` bleiben `ODR-E2001`.
- [ ] Handbuch: die Zeilen der vier Schlüssel stehen in der Tabelle der Einstellungen (Ist-Zustand), das Kataloge-Gate (`make katalog-check`) ist grün; Kapitel und Beispiele zum Modus `hybrid` liefert `slice-bootstrap-m4`.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `src/bootstrap/` (`konfig*.go`) | update | die vier Schlüssel lesen, prüfen und in die Konfiguration legen; `hybrid` als gültiger Wert von `MODE`; die vier Familien aus `nichtGelieferte` nehmen; `UPSTREAM_URL` als Pflicht bei `enabled`; die Funktion, die `policy.New` ruft und `*KonfigFehler` auf `ODR-E2001` abbildet |
| `src/bootstrap/wurzel.go` | — | kein Änderungsbedarf: `vorbereiten` lehnt `hybrid` weiter ab (Teil 2 bindet ihn); ein Bedarf geht an den Architect (§6, (f)) |
| `src/bootstrap/` (Tests) | neu | `m4_konfig_test.go` (Lesen, Syntax, Zusammenhänge, `*KonfigFehler`); Muster `m3_konfig_test.go`; kein Fake, kein Referenzdienst |
| `docs/user/` | update | Zeilen der vier Schlüssel in der Tabelle der Einstellungen; kein Pfad der Größenmarke (`SPEC-128` Nr. 3 m) |
| `spec/spezifikation.md` | update (Architect) | Nachzug von `SPEC-110` Nr. 2, 8, 9 und der Schlüsselzeilen; Verträge und Pläne zählen nicht in die Marke |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. **Stand der Tabelle** (Planner, 2026-11-18, Schnitt aus `slice-bootstrap-m4`; die Zeilen M04 bis M09 tragen dort dieselben Nummern): Entwurf, jede Zeile nennt Zusage, Fänger, Mutation und Klasse; der Architect prüft jede gegen den Nachzug von `SPEC-110`, den er noch liefert (Nr. 2 lehnt die vier Schlüsselfamilien und Nr. 8 den Modus `hybrid` heute noch ab), und trägt in §6 die Orte nach. Zeilen, deren Vertragsstelle noch fehlt, tragen „(Nachzug)“ und sind bis dahin Entwurf. Ein Fänger nennt seinen Test mit dem Kurznamen ohne `Test`. Jede Mutation, die ein Fänger fängt, fährt der Implementer selbst (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| M04 · `SPEC-261` Nr. 8 b, 10 b (Nachzug: Reihenfolge in der Meldung) | Jeder `*KonfigFehler` der Policy wird `ODR-E2001`, der Start scheitert, kein Listener öffnet; die Meldung nennt Schlüssel und Grund, nie den Wert | Schleife über die sieben Ursachen × die drei Listenschlüssel und `HYBRID_FALLBACK` (Zählprüfung gegen die Ursachen von `SPEC-261`), dazu je Paar von Verstößen an zwei Schlüsseln die Reihenfolge | Abbildung auf `ODR-E9001`; Wert in der Meldung; nur der erste Fehler gemeldet; Listener vor der Prüfung geöffnet | andere |
| M05 · `SPEC-261` Nr. 10 b, `SPEC-076` | `HYBRID_FALLBACK=enabled` ohne `UPSTREAM_URL` in irgendeinem Service ist `ODR-E2001`; `disabled` ohne `UPSTREAM_URL` startet | Tabellentest: ein Service; zwei Services, nur einer ohne URL, beide Reihenfolgen; `SERVICES_FILE` mit Eintrag ohne `upstream_url`; `disabled` und `enabled` je mit und ohne | Prüfung nur am ersten Service; Prüfung auch bei `disabled` | andere |
| M06 · `SPEC-110` Nr. 1, 3, 6 (Nachzug) | Syntax der vier Schlüssel: leer gilt als ungesetzt; `HYBRID_FALLBACK` nur `enabled` oder `disabled` in dieser Schreibweise; Listen durch Komma, Leerraum am Rand entfernt, leerer Eintrag und `none` ungültig, Doppelte ohne Wirkung | Schleife Randform × vier Schlüssel mit Zählprüfung gegen die Tabelle des Architect (Raster §6, Zeilen 1 und 2) | je Randform eine (Leerraum nicht entfernt; `Enabled` angenommen; leerer Eintrag angenommen) | andere |
| M07 · `SPEC-110` Nr. 2, 14 (Nachzug) | Gesetzte `HYBRID_*` und `FALLBACK_*` in `record`, `replay` und `mock`, `PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid`: wie der Architect entscheidet | Schleife Modus (4) × Schlüsselfamilie (4) und die zwei Schlüssel im Modus `hybrid` | je Zelle eine | andere |
| M08 · `SPEC-110` Nr. 2 | `ADMIN_*` und `LISTEN_TLS_*` bleiben `ODR-E2001` („nicht unterstützt“) in jedem Modus, auch `hybrid`; die vier geprüften Familien sind es nicht mehr | Schleife Modus (4) × Familie (6 Familien: `ADMIN_*`, `LISTEN_TLS_*`, die vier Schlüssel) mit Zählprüfung; die Meldung nennt keinen Wert | Ablehnung nur außerhalb von `hybrid`; `HYBRID_*` weiter abgelehnt | andere |
| M09 · `SPEC-110` Nr. 8 (Nachzug) | `hybrid` ist ein gültiger Wert von `MODE`: die Konfiguration meldet dafür keinen Verstoß; ein ungültiger Wert nennt die lieferbaren Werte (alle vier); die Syntax aller Schlüssel wird mitgemeldet. Der Start im Modus `hybrid` gelingt erst mit der Bindung (Teil 2, dort Zeile M01) | Test der Konfiguration: `MODE=hybrid` ohne Verstoß; ungültiger Wert mit zugleich ungültigem Schlüssel (Text im Katalog) | `hybrid` bleibt Verstoß; Meldung nennt drei Werte | andere |

**Entwurf und Rest.** Die Zeilen M06 bis M09 hängen am Nachzug von `SPEC-110` (Aufträge 1 bis 3); M04 an der Reihenfolge der Meldung (§6, Zeile *Reihenfolge der Meldung*). Die Mutationen sind Vorschläge des Planners aus dem Wortlaut der Zusagen; der Implementer ergänzt sie je Fänger und nennt die roten Tests im Bericht. Mengen für Schleifen (Generator-Pflicht): Ursachen aus `SPEC-261` Nr. 8, Modi (4), Schlüsselfamilien (6); Orakel ist eine unabhängige Fassung der Regel, nicht der Code der Policy.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem ersten Code-Commit, `AGENTS.md` §3.12):

1. **Nachzug von `SPEC-110`:** Nr. 2 (die vier Schlüssel wandern aus der Ablehnung in die geprüften), Nr. 8 (`hybrid` ist gültiger Wert; Wortlaut des Zwischenstands, solange der Start im Modus `hybrid` an `vorbereiten` scheitert, §6 (f)) und Syntax nach Nr. 3 und 6: `HYBRID_FALLBACK` (`disabled` oder `enabled`), `HYBRID_REPLAY_FALLBACK` (Liste der Kategorien), `FALLBACK_ALLOW_HOSTS` (Liste), `FALLBACK_WRITE_ALLOW` (Liste von Einträgen in der Syntax aus `slice-fallback-policy`); Randformen je Schlüssel (leer, unbekannte Kategorie, Doppelte, ungültiger Eintrag); Reihenfolge der Meldung.
2. **Zusammenhänge:** die Zusammenhänge aus Auftrag 6 von `slice-fallback-policy`; `UPSTREAM_URL` je Service Pflicht bei `HYBRID_FALLBACK=enabled` (`SPEC-076`), Schlüssel der Meldung bei `SERVICES_FILE`; gesetzte `HYBRID_*` und `FALLBACK_*` in den Modi `record`, `replay` und `mock`: ignoriert oder `ODR-E2001`.
3. **`PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid`** (`SPEC-055`, `SPEC-011`): Pflicht und Default wie im Modus `mock`, oder eigene Regel.
4. **Handbuch-Zeilen** der vier Schlüssel (`SPEC-200`, Tabelle der Einstellungen): Wortlaut für den Zwischenstand (§6, (f)).
5. Tabelle *Zusage · Fänger* und Raster in §6 bestätigen, Orte nachtragen.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-11-18 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* aus `slice-bootstrap-m3` (Nachzug für `hybrid`, `HYBRID_*`, `FALLBACK_*`), `slice-fallback-policy` (Lesen der Schlüssel) und `slice-fallback-policy-entscheidung` (`*KonfigFehler` wird `ODR-E2001`, `enabled` ohne `UPSTREAM_URL` wird `ODR-E2001`): die Geber liegen in `done/` und nennen `slice-bootstrap-m4`; mit dem Schnitt am 2026-11-18 tragen §1 *Übernimmt* und DoD Punkte 1 und 2 dieses Slice sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie; der Teil 2 (`slice-bootstrap-m4`) führt sie nicht mehr. *Von diesem Slice an:* die Rückgaben an den Architect (§6) und an den Planner; die Funktion, die die Policy baut, nimmt `slice-bootstrap-m4` (Teil 2, dessen §1 *Übernimmt* und DoD Punkt 1 sie mit der Kennung dieses Gebers tragen).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`, hier am 2026-11-18 mit dem Schnitt): `slice-fallback-policy`, `slice-fallback-policy-entscheidung` und `slice-bootstrap-m3` liegen in `done/`. Die Spezifikation trägt den Nachzug von `SPEC-110`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): **Stand 2026-11-18 (Planner): fehlt** (Nr. 2 lehnt die vier Schlüsselfamilien und Nr. 8 den Modus `hybrid` ab); **kein Code-Commit vor ihm**, Adresse Architect (§6).

**Warum:** Der Teil entstand aus dem Schnitt von `slice-bootstrap-m4` (Schätzung 2370, Schnittfrage ja nach `SPEC-128` Nr. 1; Entscheidung des Users am 2026-11-18). Er liefert die Konfiguration, die `slice-bootstrap-m4` (Teil 2) im Composition Root ruft (*Code-Abhängigkeit*: Teil 2 braucht Teil 1, schreibt dessen Dateien `konfig*.go` aber nicht um; die Naht ist die Funktion, die die Policy baut). Er braucht kein `HybridUseCase` und keinen Referenzdienst: die Tests laufen allein auf der Konfiguration und dieser Funktion, ohne Fake. Ohne Bindung ändert sich am Binary: die vier Schlüssel und der Wert `hybrid` werden gelesen und geprüft; ein Start im Modus `hybrid` scheitert weiter in `vorbereiten` (`ODR-E2001`, „Betriebsmodus nicht lieferbar“; Wortlaut des Zwischenstands §6, (f)).

**Reihenfolge zu `slice-bootstrap-m6`:** beide ändern `src/bootstrap/` und `SPEC-110` Nr. 2; sie laufen nacheinander und nicht parallel, welcher zuerst, ist Priorisierung, und der zweite zieht den Stand des ersten nach (`AGENTS.md` §3.9).

**Größenmarke** (Planner, 2026-11-18, vor dem ersten Code-Commit; Tabelle *Zusage · Fänger* in §3 mit 6 Zeilen): *a* = 0, *b* = 6, *f* = 0, Produktcode rund 150 (Vorschlag des Planners, der Architect bestätigt oder ändert ihn: Lesen und Prüfen der vier Schlüssel mit Zusammenhängen und die Funktion der Policy). *S* = 85·0 + 80·6 + 200·0 + 150 = **630**, *M* = 100 · ⌈135·630 / 10000⌉ = **900**. Keine Schnittfrage (135·630 = 85050 ≤ 300000, `SPEC-128` Nr. 1). **Basis** ist der Commit vor dem ersten Code-Commit: `7a5170d9` (reiner Move von `slice-bootstrap-m4` nach `open/`, Planner, 2026-11-18); die Pfade des Blocks sind bis dahin unberührt. Ein Merge von `main` in den Zweig, der diese Pfade ändert, zählt in die Messung (`SPEC-128` Nr. 4 g); der Merge steht dann als `ueberschreitung` mit Grund oder als neue Basis mit `ersetzt`.

```groessenmarke
pfad: src/bootstrap/
produktcode: 150
fakes: 0
basis: 7a5170d9da222d9f13bd2106554e196a85114455
schaetzung: 630
marke: 900
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 900 (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)).
- `in-progress` → `open` (blockiert — Carveout?): der Nachzug von `SPEC-110` liegt nicht vor.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Planner, 2026-11-18, gegen den Stand von `spec/spezifikation.md` am Start des Zweigs; Zeilen 1 bis 3 des Rasters von `slice-bootstrap-m4`, die Zeile *Start und Herunterfahren* liegt bei Teil 2): „(offen, Auftrag n)“ heißt, der Architect trägt den Ort in der Spezifikation vor dem ersten Code-Commit ein, die Adresse steht in der Zeile *Offene Zellen* unten; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12). Orte mit „(Nachzug)“ tragen den Wortlaut erst, wenn `SPEC-110` Nr. 2 den Schlüssel nennt; bis dahin lehnt Nr. 2 ihn ab.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Lesen von `HYBRID_FALLBACK` und `HYBRID_REPLAY_FALLBACK` | `SPEC-110` Nr. 1 mit den Defaults `SPEC-010` (`disabled`) und `SPEC-079` (leer) | `SPEC-110` Nr. 1: leer gilt als ungesetzt | `HYBRID_FALLBACK`: `SPEC-010` und `SPEC-110` Nr. 3 (Schreibweise zählt; Nachzug); `HYBRID_REPLAY_FALLBACK`: Eintrag `SPEC-261` Nr. 7 (Ursache `kategorie`, `leer`), Liste `SPEC-110` Nr. 6 (Nachzug); Doppelte `SPEC-261` Nr. 1 a | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-110` Nr. 13 |
| Lesen von `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` | `SPEC-110` Nr. 1 mit den Defaults `SPEC-062` und `SPEC-080` (leer) | `SPEC-110` Nr. 1: leer gilt als ungesetzt; ein leerer Eintrag in der Liste (`a,,b`): (offen, Auftrag 1) | Eintrag: `SPEC-261` Nr. 5 a (Host, Ursache `host`, `leer`) und Nr. 6 a, b (Methode, Muster, Ursachen `trenner`, `methode`, `muster`); Doppelte `SPEC-261` Nr. 1 a; Zerlegung der Liste und Leerraum `SPEC-261` Kopf (Werte), `SPEC-110` Nr. 6 (Nachzug) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-110` Nr. 13 |
| Zusammenhänge der Schlüssel und `UPSTREAM_URL` im Modus `hybrid` | `enabled` ohne `UPSTREAM_URL` in irgendeinem Service: `SPEC-261` Nr. 10 b, `SPEC-076`; der Schlüssel der Meldung bei `SERVICES_FILE` und die Bedingung von `upstream_url` im Modus `hybrid`: (offen, Auftrag 2) | `enabled` mit leerer Allowlist: `SPEC-261` Nr. 8 c (Ursache `ohne-hosts`, Schlüssel `HYBRID_FALLBACK`) | `disabled` mit gesetzten Listen, Freigabe ohne Host, Host ohne Service: kein Fehler, `SPEC-261` Nr. 8 c, 5 e; `ADMIN_*` und `LISTEN_TLS_*`: `SPEC-110` Nr. 2; gesetzte `HYBRID_*` und `FALLBACK_*` in `record`, `replay`, `mock`: (offen, Auftrag 2); `PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid`: (offen, Auftrag 3); Reihenfolge der Meldung: (offen, Auftrag 1) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-110` Nr. 13 |

**Offene Zellen** (Rückgabe an den Architect, Planner, `AGENTS.md` §3.13; Ausgang *offen*, Adresse Architect, vor dem ersten Code-Commit):
(a) leerer Eintrag in einer Liste der vier Schlüssel: `SPEC-110` Nr. 6 nennt ihn ungültig, `SPEC-261` Nr. 5 a, 6 a, 7 nennen ihn als Ursache `leer` der Policy; wer meldet ihn, und gilt Nr. 12 (ein Grund je Schlüssel) in beiden Fällen;
(b) Reihenfolge der Meldung: `SPEC-110` Nr. 12 sortiert nach Schlüssel in Bytefolge (`FALLBACK_ALLOW_HOSTS`, `FALLBACK_WRITE_ALLOW`, `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`), `SPEC-261` Nr. 8 b fügt die Fehler in der Reihenfolge `FALLBACK_ALLOW_HOSTS`, `FALLBACK_WRITE_ALLOW`, `HYBRID_REPLAY_FALLBACK`, dann `HYBRID_FALLBACK`; §1 dieses Plans sagt „entscheidet der Architect“; welche Reihenfolge gilt für die Meldung;
(c) Schlüssel der Meldung für `enabled` ohne `UPSTREAM_URL` (`UPSTREAM_URL` oder `SERVICES_FILE[<i>].upstream_url`) und ob `upstream_url` im Modus `hybrid` nur bei `enabled` Pflicht ist (`SPEC-110` Nr. 9 kennt die Pflicht nur im Modus `record`);
(d) gesetzte `HYBRID_*` und `FALLBACK_*` in den Modi `record`, `replay`, `mock`: ignoriert oder `ODR-E2001` (Auftrag 2), und `PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid` (Auftrag 3);
(f) Zwischenstand bis zur Bindung: die Konfiguration nimmt `hybrid` und die vier Schlüssel an, `vorbereiten` (`src/bootstrap/wurzel.go`) lehnt den Modus mit „Betriebsmodus nicht lieferbar“ ab; welchen Wortlaut trägt `SPEC-110` Nr. 8 dafür, und gelten die Schlüssel in `mock`, `record` und `replay` als wirkungslos und damit nach dem Zweck von Nr. 2 (keine still wirkungslose Einstellung) als `ODR-E2001` (hängt an (d));
((e), Start und Herunterfahren im Modus `hybrid`, liegt bei `slice-bootstrap-m4`;)
(g) Handbuch im Zwischenstand: `make katalog-check` verlangt die vier Schlüssel in der Tabelle der Einstellungen, sobald `konfig*.go` sie nennt (`SPEC-201` Nr. 4); Wortlaut der Zeilen, solange der Modus `hybrid` nicht startet (`SPEC-200`, Auftrag 4).

**Zusatztabelle** (Operation meldet mehrere Verstöße, Zusage hängt vom Modus ab; `AGENTS.md` §3.16):

| Fall | Ort | Fänger |
|---|---|---|
| Zwei Verstöße an zwei Schlüsseln der Policy (Paare der vier Schlüssel) | `SPEC-110` Nr. 12, `SPEC-261` Nr. 8 b; Reihenfolge (offen, (b)) | M04, Schleife über alle Paare (6) |
| Zwei Verstöße am selben Schlüssel (leerer Eintrag und ungültiger Eintrag in `FALLBACK_ALLOW_HOSTS`; `enabled` mit leerer Allowlist und fehlender `UPSTREAM_URL`) | `SPEC-110` Nr. 12 (höchstens ein Grund je Schlüssel), `SPEC-261` Nr. 8 b (der erste je Schlüssel), `SPEC-261` Nr. 8 c (`ohne-hosts` entfällt bei ungültigem Eintrag); (offen, (a), (c)) | M04, je Paar ein Fall |
| Modus `record`, `replay`, `mock`, `hybrid` × Schlüsselfamilie | `SPEC-110` Nr. 2 (`ADMIN_*`, `LISTEN_TLS_*`); die vier Familien: (offen, (d)) | M07, M08, Schleife 4 × 6 |
| Modus `hybrid` mit und ohne `STRICT_COVERAGE`, `PUBLIC_BASE_URL` | (offen, (d)) | M07 |

- Gesetzte `HYBRID_*` und `FALLBACK_*` in den Modi `record`, `replay` und `mock` (Auftrag 2) — **Ausgang:** offen; Adresse: Architect ((d)).
- **Start-Bedingung nicht erfüllt** (Planner, 2026-11-18): `SPEC-110` Nr. 2 und 8 tragen den Nachzug nicht, die Tabelle ist Entwurf, die offenen Zellen (a) bis (d), (f), (g) haben keinen Ort — **Ausgang:** offen; Adresse: Architect (Aufträge 1 bis 5); kein Code-Commit vor dem Nachzug (`AGENTS.md` §3.12).
- **Zwischenstand und Handbuch-Zeilen** ((f), (g)): zwischen diesem Slice und `slice-bootstrap-m4` liest das Binary die Schlüssel, ohne dass `hybrid` startet — **Ausgang:** offen; Adresse: Architect; beide Slices gehen nacheinander in denselben Zweig, kein Release dazwischen.
- Gleichzeitige Änderung von `src/bootstrap/` und `SPEC-110` Nr. 2 durch `slice-bootstrap-m6` — **Ausgang:** offen; Adresse: Planner (Reihenfolge beim Start, keine parallelen Arbeitsbäume auf diesen Dateien).
- Größe: die ungeschnittene Schätzung 2370 ergab eine Schnittfrage — **Ausgang:** eingetreten (Schnitt vor dem Code in diesen Slice und `slice-bootstrap-m4`, Entscheidung des Users 2026-11-18); dieser Teil: *S* 630, Marke 900.
- Sendungen aus `slice-bootstrap-m3`, `slice-fallback-policy`, `slice-fallback-policy-entscheidung` (§3, Prüfung der Sendungen) — **Ausgang:** eingetreten (angenommen, §1 *Übernimmt* und DoD Punkte 1 und 2).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-18, aus der Sichtung von `slice-bootstrap-m4`, 2026-10-30, übernommen; der Planner zieht Treffer und Zähler beim Start nach). Offen und für den Slice einschlägig: `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `AGENTS.md` §3.9; der Nachzug von `SPEC-110` ist eine Stelle, die eine Berichtigung trägt) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
