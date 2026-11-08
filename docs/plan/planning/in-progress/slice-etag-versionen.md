# Slice slice-etag-versionen: ETag aus dem Versionszähler der Entity

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-etag](../welle-etag.md).

**Bezug:** `LH-FA-15`, `LH-QA-02`, `LH-FA-07`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-15.a` (Entity-Versionen als ETag; ohne aktivierte Unterstützung kein scheinbarer Schutz) · `LH-QA-02.a` (ETags leiten sich aus dem Versionszähler ab) · `SPEC-003` (`entity_version`) · `SPEC-049` (Zeile GET, Spalte Soll: ETag) · `LH-FA-07.b` (Antworten auf POST, PATCH, PUT) · `SPEC-108` Nr. 1 (`WithWorkspaceTransaction`) · `ARC-002`, `ARC-003`, `ARC-004` · `SPEC-280` bis `SPEC-282` (Architect, 2026-11-08, Nachzüge in `SPEC-049`, `SPEC-110`, `SPEC-168`, `SPEC-169`, `SPEC-193`, `SPEC-194`, `SPEC-198`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-etag-versionen`, Worktree `/Development/odata-recorder-wt-etag-versionen`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Mock liefert je Entity ein Versionskennzeichen (ETag), das allein aus dem Versionszähler `entity_version` der Entity im Workspace folgt: eine unveränderte Entity trägt bei jedem Lesen denselben ETag, jede erfolgreiche Änderung der Entity (POST, PATCH, PUT) macht einen neuen sichtbar, ein fehlgeschlagener Vorgang lässt Zähler und ETag unverändert, und derselbe Zustand ergibt nach Reset und Restore denselben ETag (`LH-QA-02`). Der ETag steht im Header und in der Darstellung der Entity (Entity-Abfrage, Elemente einer Collection, Antworten auf POST, PATCH und PUT mit Entity). Ohne aktivierte ETag-Unterstützung steht keines von beiden in der Antwort: der Mock behauptet keinen Concurrency-Schutz (`LH-FA-15.a`).

**Übernimmt (Zähler und Aktivierung):** Die Entscheidung, wer `entity_version` erhöht, solange kein ETag ausgeliefert wird (offener Punkt in §6 von `slice-mock-schreiben`, Sendung von `slice-mock-schreiben`, `AGENTS.md` §3.13), liegt dort; dieser Slice baut auf ihr auf und liefert die Auslieferung des ETag. Was *aktiviert* heißt (Schlüssel, immer an), entscheidet der Architect in diesem Slice (§3).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `If-Match` und `If-None-Match`, der Vergleich mit dem Zähler und `412` — `slice-if-match`; dieser Slice liefert den ETag, prüft aber keinen Request-Header.
- `Prefer: return=minimal|representation` — `slice-prefer-return`; bis dahin gilt das Muss-Verhalten (POST `201` mit Entity, PATCH und PUT `204`).
- ETags im Record und Replay — aufgezeichnete ETags bleiben Header der Exchanges (`welle-etag` §6); der Mock erzeugt ETags nur aus seinem Zähler.
- Verteilte Versionskoordination über mehrere Instanzen — Out-of-Scope von `LH-FA-15`.
- Das gebundene Handbuch-Beispiel zu `ETAG_ENABLED` (Start mit `ETAG_ENABLED=true`, Lesen, Ändern, neuer ETag) und die Umgebung im Beispiel-Format von `SPEC-200` Nr. 6 — `slice-handbuch-mock` (Architect, 2026-11-08: dessen Vertragserweiterung um die Umgebung ist dieselbe Voraussetzung; dieser Slice trägt nur die Zeile der Einstellungen; die Adresse nimmt an, sobald der Planner sie in dessen §1 und DoD einträgt, §6).

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

- [ ] Bildung und Auslieferung: der ETag folgt allein aus `entity_version` (Form nach Architect), er steht als Header und in der Darstellung der Entity in jeder Antwort, die eine Entity trägt (GET auf Key, Elemente einer Collection, POST, PATCH, PUT mit Entity); gleicher Zustand ergibt denselben ETag, auch nach Reset und Restore.
- [ ] Zähler: `entity_version` steigt bei jeder erfolgreichen Mutation der Entity in derselben Workspace-Transaktion; ein fehlgeschlagener Vorgang, andere Entities und andere Workspaces lassen ihn unverändert; ein Schlüssel, der nach DELETE neu angelegt wird, trägt keinen ETag, den die frühere Entity trug (Randform in §6).
- [ ] Aktivierung: ohne aktivierte ETag-Unterstützung tragen Antworten weder Header noch `@odata.etag` (`LH-FA-15.a`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch im selben Slice nachgezogen (Zeile `ETAG_ENABLED` der Einstellungen, Ist-Zustand; das Beispiel als Datei liegt bei `slice-handbuch-mock`, §1 und §6); das Kataloge-Gate ist grün.
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
| `src/hexagon/services/mock/` | update | `Dienst.Etag`; `EtagFuer(Version)` in den gedeuteten Entities (Lesen) und in `204`/`201` der Mutationen (Header `ETag`), ein Lesen je Antwort (`SPEC-281` Nr. 1) |
| `src/hexagon/model/` | update | `EtagFuer(version int64) string`, Feld `Entity.Etag`, `FuehreAus` behält es bei der Projektion (`abfrage.go`, `SPEC-281` Nr. 3) |
| `src/adapters/driven/odata/jsoncodec/` | update | `@odata.etag` hinter `@odata.id` in der Stufe `minimal`, Prüfung der Form von `Entity.Etag` (`SPEC-168` Nr. 3 e, `SPEC-169` Nr. 5) |
| `src/adapters/driving/httpodata/` | update | nur Tests, falls der Adapter die Header der `Antwort` unverändert durchreicht (Fänger E7 über den Adapter); sonst die Durchreichung von `ETag` |
| `src/bootstrap/` (`konfig.go`, `konfig_werte.go`, `wurzel.go`) | update | Schlüssel `ETAG_ENABLED` (`SPEC-282`, `SPEC-110` Nr. 15), `Dienst.Etag` nur für Dienste der Version `4.0` (`SPEC-280` Nr. 2); neu gegenüber dem Entwurf des Plans, vom Architect ergänzt |
| `spec/spezifikation.md` | geliefert (Architect, im Zweig) | `SPEC-280` bis `SPEC-282`, Nachzüge `SPEC-110`, `SPEC-168`, `SPEC-169`, `SPEC-193`, `SPEC-194`, `SPEC-198`, `SPEC-049`; dieser Slice schreibt sie nicht |
| Tests | neu | Zeilen E1 bis E24 der Tabelle; Wirkungs-Fänger für den Zähler unter gleichzeitigen Änderungen (E18, E19) |
| `src/hexagon/services/mock/kein_zustand_test.go`, `fakes_test.go` | update | `TestMockerFelder` zählt vier Felder von `Dienst` (ID, Root, Codecs, `Etag`); der Fake von `Entities` kennt `wachsend` (E8), der Fake des Schreibers `zeigeEtag` (E19) |
| `src/adapters/driven/odata/jsoncodecv2/` | Tests | `TestEtagNieGeschrieben` (E13); kein Pfad des Blocks |
| `docs/user/benutzerhandbuch.md` | update | Zeile `ETAG_ENABLED` der Einstellungen (E24; das Beispiel als Datei liegt bei `slice-handbuch-mock`, §6) |

**Vom Architect entschieden** (2026-11-08, im Zweig `slice-etag-versionen`; Bezug bleibt `LH-FA-15`, `LH-QA-02`, `LH-FA-07`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), keine neue ADR: die Entscheidungen stehen mit Alternativen und Folge in `SPEC-280` und `SPEC-281`): (1) Form `W/"<n>"`, schwach, allein aus `entity_version`, ohne Set, Schlüssel, Service, Workspace (`SPEC-280` Nr. 1). (2) Schlüssel `ETAG_ENABLED`, Vorgabe `false`; aus: weder `ETag` noch `@odata.etag`, `If-Match` ignoriert (`SPEC-280` Nr. 2 und 3, `SPEC-282`). (3) Startwert 0 bei Baseline und Reset, erste Mutation 1; Neuanlage nach `DELETE` trägt Tombstone-Version + 1 und damit nie einen früheren ETag des Schlüssels (`SPEC-280` Nr. 4 c). (4) Zähler: nur `Put` des Dienstes in der Mutationstransaktion; Reset und Restore stellen ihn wieder her, ohne Epoche (`SPEC-280` Nr. 4 d). (5) Collection: je Element `@odata.etag`, kein Header (`SPEC-281` Nr. 1). (6) `@odata.etag` im Body der Stufe `minimal`, nie in `none`; der Header steht in beiden (`SPEC-281` Nr. 1, 4 e).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-08; der Ort im Vertrag steht in der Zelle *Nr.*; die Kennung ist keine Zählung; Tests der Pakete unter den Pfaden aus §3, soweit kein anderes Paket genannt ist):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| E1 (`SPEC-280` Nr. 1) | `EtagFuer(n)` ist `W/"` + Dezimalzahl ohne führende Null + `"` (`0` für null), schwach | `model`: Tabelle 0, 1, 9, 10, 99, 100, 2147483647, 4294967296, 9223372036854775807 mit Literal und Schleife 0 bis 300 gegen einen unabhängigen Aufbau der Zeichenfolge | `W/` weggelassen; `"` weggelassen; `%x` statt `%d`; Auffüllen mit führender Null | andere |
| E2 (`SPEC-280` Nr. 1) | Der ETag hängt allein von der Version ab: zwei Einträge mit gleicher Version in anderem Set, mit anderem Schlüssel, in anderem Service und anderem Workspace tragen denselben ETag | `mock`: Schleife über 2 Sets × 2 Schlüssel × 2 Workspaces, Version je Eintrag gleich; ETags gleich; dazu andere Version, anderer ETag | Set, Schlüssel oder Workspace in den Text gemischt | andere |
| E3 (`SPEC-282`, `SPEC-110` Nr. 15) | `ETAG_ENABLED`: ungesetzt, leer und `false` ergeben aus, `true` an; `1`, `True`, `yes` sind der Verstoß `ETAG_ENABLED` (Wert nicht in der Meldung), in jedem Modus geprüft, nur im Modus `mock` wirksam; die Konfiguration trägt den wirksamen Wert | `bootstrap`: Schleife über 6 Werte × die Modi `record`, `replay`, `mock`; Meldung enthält den Wert nicht; Nullwert von `mock.Dienst` ist aus (Test auf Konfig ohne das Feld) | Vorgabe `true`; Prüfung nur im Modus `mock`; Wert in der Meldung | andere |
| E4 (`SPEC-280` Nr. 2) | Der Composition Root setzt `Dienst.Etag` nur für einen Dienst der Version `4.0`; ein Dienst der Version `2.0` bleibt aus, auch bei `ETAG_ENABLED=true` | `bootstrap`: die Hilfsfunktion, die `Dienst.Etag` bildet, über (`4.0`, `2.0`) × (aus, an); Verdrahtung des Modus `mock` an: `Etag` am Dienst des Satzes `4.0` wahr | Versionsprüfung entfernt; `Etag` immer aus | andere |
| E5 (`SPEC-280` Nr. 2) | Aus: keine Antwort trägt `ETag` oder `@odata.etag`, in jeder Art (GET Entity, HEAD Entity, Collection, `$count`, `$metadata`, POST, PATCH, PUT, DELETE, jeder Fehler); Bodies und Header gleich dem Stand ohne diesen Vertrag | `mock`: Schleife über die zehn Arten × 2 Dienste (Satz `nil`, eigener Satz) mit Orakel; die 92 Schritte der Abnahme (`TestVertragMatrix`, 36 IDs) bleiben unverändert grün | Prüfung von `Dienst.Etag` beim Header oder beim Setzen von `Entity.Etag` entfernt | andere |
| E6 (`SPEC-280` Nr. 3) | `If-Match` und `If-None-Match` ändern nichts, bei aus und bei an: gleiche Antwort und gleicher Zustand wie ohne Header, auch bei `*`, fremdem und veraltetem Wert; der Dienst liest beide Header nie | `mock`: Schleife über Methode (PATCH, PUT, DELETE) × Header (`If-Match`, `If-None-Match`) × Wert (`*`, `W/"99"`, aktueller ETag, leer) × aus/an: Antwort gleich der ohne Header, Zustand gleich; kein Zähler auf `Header` (akzeptiertes Negativ, `SPEC-280` Nr. 3) | Auswertung von `If-Match` eingebaut (`412` bei Abweichung) | andere |
| E7 (`SPEC-281` Nr. 1) | An: `GET` und `HEAD` einer Entity tragen `ETag: W/"<Version>"` und `@odata.etag` mit demselben Text; ein wiederholtes Lesen derselben Entity gibt denselben ETag (`LH-FA-15` Boundary); sonst nur die bisherigen Header | `mock` und `httpodata`: Versionen 0, 1, 5 je Methode (GET, HEAD); 10 Lesungen derselben Entity gleich; Header-Menge gleich der von aus plus `ETag` | Version + 1 im Header; Header ohne Mitglied oder umgekehrt; zufälliger Anteil | andere |
| E8 (`SPEC-281` Nr. 1) | Header und Mitglied entstehen aus **einem** Lesen des Eintrags; der Dienst liest für den ETag nie ein zweites Mal | `mock`: Fake von `Entities` zählt jede Operation und liefert bei jedem `Get` eine höhere Version: ein `Get`, Header gleich Mitglied | zweiter `Get` für den Header | andere |
| E9 (`SPEC-281` Nr. 1, 3) | Collection: jedes Element trägt `@odata.etag` gleich dem ETag, den `GET` derselben Entity liefert; kein Header `ETag`; `X-Recorder-Coverage` unverändert; ein Tombstone erscheint nicht | `mock`: Schleife über alle Elemente des Ergebnisses je Abfrage (ohne Option, `$top`, `$skip`, `$orderby`, `$filter`, `$select`, `$count=true`), 8 Entities mit gemischten Versionen und einem Tombstone; Elemente gegen `GET` Key | Header `ETag` auch an der Collection; `Etag` nur am ersten Element | andere |
| E10 (`SPEC-281` Nr. 3) | `FuehreAus` behält `Etag` bei jeder Auswahl, Ordnung und Seite, auch bei der Projektion mit `$select` | `model`: Entities mit gesetztem `Etag` durch `Select` (alle, einzelne, mit und ohne Schlüssel-Property), `Orderby`, `Top`, `Skip`: jedes Ergebnis trägt dasselbe `Etag` | `Etag` in der Projektion nicht kopiert (`abfrage.go`) | andere |
| E11 (`SPEC-281` Nr. 4 e) | Stufe `none`: Kopfzeile `ETag` bleibt, kein `@odata.etag` im Body; `Accept` ändert den ETag nie | `mock` und `jsoncodec`: Schleife über `Accept` (`none`, `minimal`, `full`, `IEEE754Compatible=true`, `text/html`, fehlend) × Entity und Collection; Wert des Headers gleich über alle | Mitglied auch in `none`; ETag hängt von `Accept` ab | andere |
| E12 (`SPEC-281` Nr. 1, 4 a) | Kein ETag an `/$count`, `/$metadata`, Antworten mit Fehler (darunter `GET` auf Tombstone und fehlenden Eintrag, `ODR-E6008`, `ODR-E6003`) | `mock`: Schleife über die Arten und Fehlercodes, jede Antwort ohne `ETag`; Fehlerfälle als `*inbound.Fehler` ohne Antwort | ETag bei `$count` gesetzt | andere |
| E13 (`SPEC-168` Nr. 3 e, `SPEC-169` Nr. 5, 8) | Der Schreiber schreibt `@odata.etag` hinter `@odata.id` in `minimal`, nie in `none`, nie für leeres `Etag`; ein ungültiges `Etag` ist `ErrUngueltig` und schreibt nichts; `jsoncodecv2` schreibt es nie; der Leser liest den Body zu denselben Entities zurück | `jsoncodec`: gültig 0 bis 9223372036854775807; Menge ungültiger Formen (ohne `W/`, `w/`, ohne `"`, Leerraum, führende Null, `+`, `-`, Buchstabe, 20 Ziffern, 9223372036854775808, `"`, Backslash und Steuerzeichen im Text, nur `W/""`); Rundlauf mit `ODataCodec`; `jsoncodecv2` mit gesetztem `Etag` schreibt gleiche Bytes wie ohne | Prüfung der Form entfernt; Mitglied vor `@odata.id`; V2 schreibt es | andere |
| E14 (`SPEC-281` Nr. 2) | `POST`: `201` mit `ETag` und `@odata.etag` der neuen Version (`1` ohne Vorgänger), `Location` und die übrigen Header unverändert | `mock` und `httpodata`: POST auf neuen Schlüssel in Baseline-Set und Set ohne Eintrag; Header-Menge = bisherige + `ETag` | ETag der Version 0; `Location` fehlt | andere |
| E15 (`SPEC-281` Nr. 2) | `PATCH`, `PUT`: `204` mit `ETag` der Version nach der Mutation, auch bei `{}` und bei gleichem Inhalt, kein Body; `DELETE`: `204` ohne ETag | `mock`: Schleife (PATCH, PUT, DELETE) × Version vorher 0, 1, 4; ETag der Antwort = ETag des folgenden `GET` und ≠ dem vorherigen; PATCH `{}` und PUT gleiches Objekt ändern den ETag | ETag der Version vor `Put`; `{}` zählt nicht; DELETE trägt ETag | andere |
| E16 (`SPEC-280` Nr. 4 b) | Eine fehlgeschlagene Mutation lässt Zähler und ETag unverändert und trägt keinen ETag, auch bei Fehler nach `Put` (Rollback) | `mock`: Schleife über die Fehlerarten (`ODR-E6009`, `ODR-E6008`, `ODR-E6003`, `ODR-E6007` je Ursache, `ODR-E7001`, Fehler von `Advance`, des Schreibers nach `Put`, `context.Canceled` und `context.DeadlineExceeded` als Fehler von `Put` und `Advance` in der Transaktion): ETag des folgenden `GET` gleich dem vorher | Version vor der Prüfung erhöht; Rollback fehlt | andere |
| E17 (`SPEC-280` Nr. 4 b) | Andere Einträge und andere Workspaces behalten ihre Version bei einer Mutation | `mock`: 3 Entities in 2 Workspaces, Mutation einer, ETags der fünf anderen unverändert (auch gleicher Schlüssel in `ws-2`) | `Put` ohne Workspace-Kennung; gemeinsamer Zähler | andere |
| E18 (`SPEC-280` Nr. 5) | *n* gleichzeitige erfolgreiche `PATCH` derselben Entity mit Version *v* geben *n* verschiedene ETags, genau `EtagFuer(v+1)` bis `EtagFuer(v+n)`; der Endstand ist `EtagFuer(v+n)` | `mock` (Wirkungs-Fänger, zählt die verschiedenen ETags der Antworten): *n* = 2, 8, 32 Goroutinen über die echte SQLite-Transaktion (`bootstrap`) und über den Fake; Rate je Mutation und je Lauf von `go test` mit Zahl der Prozesse und Messendem im Plan | Version vor der Transaktion gelesen; Antwort liest den Eintrag erneut nach `Put` | nebenläufig |
| E19 (`SPEC-280` Nr. 5) | Ein Lesen gleichzeitig mit Mutationen zeigt Header und Mitglied aus einer Version, die vor oder nach einer Mutation liegt, nie eine Mischung | `mock`: *m* Mutationen gegen *l* Leser (`GET` Entity und Collection), jede Antwort: Header gleich Mitglied, Version aus *v* bis *v+m*, je Leser nicht fallend; Rate wie E18 | zweites Lesen für den Header (zeigt E8 nur ohne Gleichzeitigkeit) | nebenläufig |
| E20 (`SPEC-280` Nr. 4 c) | Löschen und Neuanlage desselben Schlüssels: die Neuanlage trägt keinen ETag, den eine frühere Version des Schlüssels trug | `mock`: Schleife über Vorgeschichte (Baseline-Version 0; 1 bis 3 `PATCH`; schon einmal gelöscht und neu angelegt) × 1 bis 3 Wiederholungen von `DELETE` und `POST`: alle gelieferten ETags des Schlüssels paarweise verschieden und streng steigend | `DELETE` erhöht die Version nicht (`POST` setzt `1`); `POST` über Tombstone setzt `1` | andere |
| E21 (`SPEC-280` Nr. 4 a, d) | Reset: derselbe Zustand hat denselben ETag wie im ersten Lauf (Version 0); `LH-QA-02` | `bootstrap` (echte SQLite): 100 Läufe aus Mutation, Reset, Lesen: ETags aller Läufe gleich dem des ersten Lesens; Tombstone nach Reset nicht mehr sichtbar | Reset lässt die Version stehen; Version zählt fort | andere |
| E22 (`SPEC-280` Nr. 4 d) | Restore stellt die Version des Snapshots her, auch für einen später gelöschten Eintrag und einen später angelegten | `bootstrap` (echte SQLite): Snapshot bei Version 2, danach `PATCH`, `DELETE`, `POST` anderer Schlüssel, Restore: ETags gleich denen beim Anlegen des Snapshots, gelöschter Eintrag wieder sichtbar mit Version 2 | Restore setzt `0`; Restore zählt fort | andere |
| E23 (`SPEC-281` Nr. 1) | Ein Dienst mit `Etag` an unterscheidet sich von aus allein um die Kopfzeile `ETag` und das Mitglied `@odata.etag`: alle Schritte der Abnahme (36 IDs, 92 Schritte), vom ETag befreit, gleich den Fixtures | `bootstrap` (`abnahme_etag_test.go`): `TestVertragMatrix` mit `Etag` an, vor dem Vergleich Header `ETag` und Mitglied `@odata.etag` entfernt; zusätzlich Menge der Header-Namen = Fixture + `ETag` bei Entity, `201`, `204` (nicht `DELETE`) | zusätzlicher Header (`Cache-Control`); Reihenfolge der Mitglieder geändert | andere |
| E24 (DoD, `SPEC-280` Nr. 2) | Das Handbuch nennt `ETAG_ENABLED`, Form, Auslieferung und dass `If-Match` bis zur Auswertung ohne Wirkung ist; kein Beispiel als Datei in diesem Slice (liegt bei `slice-handbuch-mock`, §6) | `make handbuch-check` (Kopf, Regel des Handbuchs) und `make katalog-check` (Schlüsselkatalog nennt `ETAG_ENABLED`) | Zeile `ETAG_ENABLED` aus der Tabelle der Einstellungen entfernt (`make katalog-check` rot) | andere |

Zeilen gesamt 24: *a* = 2 (E18, E19), *b* = 22. Zusammenfassungen des Architect: Konfiguration (Wert, Modus, Meldung) in E3; Schreiber (Form, Platz, Rundlauf, V2) in E13; die drei Mutationen mit `204` in E15; Reset und Restore sind zwei Zeilen, weil ihre Zustände verschieden entstehen.

**Umsetzung und Belege** (Implementer, 2026-11-08; `AGENTS.md` §3.9, §3.10). Jede Mutation lief auf einer Kopie des Baums (Quellen schreibgeschützt eingehängt), je Mutation ein Container ohne Netz mit `--memory=4g --cpus=4`, `go test -count=1 -run <Test>` des Pakets; *rot* heißt: der genannte Test schlägt an, der Aufbau baut. Die Mutationen des Pakets `mock` und der Abnahme liefen nach der letzten Änderung des Codes (Auslagerung von `kopf204`, Lint) ein zweites Mal, alle wieder rot (die bei E20 genannte Ausnahme „`DELETE` erhöht die Version nicht“ blieb grün, wie dort benannt). Treiber ist ein Skript im Scratchpad des Implementers (`sed` auf der Kopie), nicht `tools/harness/mutationslauf.sh`: dessen Nullprobe-Grenze trägt `src/bootstrap` nicht.

| Nr. | Zusage · Mutation | roter Test |
|---|---|---|
| E1 | Form von `EtagFuer`: ohne `W/`; ohne schließendes `"`; hexadezimal; mit Auffüllen durch Nullen | `TestEtagFuer` (model) |
| E2 | allein aus der Version: Schlüssel, Set oder Workspace in den Wert gemischt | `TestEtagNurAusVersion`, `TestEtagWiederholung`, `TestEtagAusUndAn` |
| E3 | `ETAG_ENABLED`: Vorgabe `true`; Wert in der Meldung; Prüfung nur im Modus `mock` | `TestKonfigEtagEnabled` |
| E4 | Version `4.0` allein: Prüfung entfernt; `Etag` immer aus | `TestEtagFuerSatz`; bei aus auch `TestEtagVerdrahtung` |
| E5 | aus ohne ETag: Prüfung von `Dienst.Etag` entfernt | `TestEtagAusUndAn`, `TestEtagPost` |
| E6 | `If-Match` ausgewertet; `If-None-Match` ausgewertet | `TestEtagHeaderKeinEinfluss` |
| E7 | Header: Version + 1; kein Header; HEAD ohne Header; Adapter verwirft `ETag` | `TestEtagWiederholung`, `TestEtagAusUndAn`, `TestAntwortEtag` (httpodata) |
| E8 | ein Lesen: zweites Lesen für den Header; Antwort liest nach `Put` erneut | `TestEtagEinLesen` |
| E9 | Collection: Header `ETag` auch an ihr; Mitglied nur am ersten Element | `TestEtagCollection`, `TestEtagAccept` |
| E10 | Projektion kopiert `Etag` nicht | `TestFuehreAusBehaeltEtag` (model) |
| E11 | Mitglied auch in der Stufe `none`; ETag hängt von `Accept` ab | `TestEtagMitglied` (jsoncodec); `TestEtagAccept` |
| E12 | ETag an `$count` | `TestEtagKeinETagBeiFehler`, `TestEtagAusUndAn` |
| E13 | Schreiber: Prüfung der Form entfällt; Mitglied vor `@odata.id`; Mitglied für leeres `Etag`; führende Null erlaubt; Wert über 2^63-1 erlaubt; Zeichen vor der Zahl (`+`, `-`) erlaubt; Endzeichen `"` nicht verlangt; `jsoncodecv2` schreibt das Mitglied | `TestEtagUngueltig`, `TestEtagMitglied`, `TestEtagGenerator`, `TestEtagNieGeschrieben` (jsoncodecv2) |
| E14 | `POST`: ETag der Version vor dem Eintrag | `TestEtagPost`, `TestEtagAusUndAn`, `TestEtagEinLesen` |
| E15 | `PATCH` trägt den ETag der Version vor `Put`; `{}` zählt nicht; `DELETE` trägt einen ETag | `TestEtagMutationen` |
| E16 | Fehler der Antwort nach `Put` verschluckt; Transaktion bestätigt trotz Fehler | `TestEtagFehlschlag` |
| E17 | Zähler gemeinsam für alle Einträge; `Put` auch in einen anderen Workspace | `TestEtagAndereBleiben` |
| E18 | Version vor der Transaktion gelesen | `TestEtagGleichzeitigeAenderungen` (Fake), `TestEtagGleichzeitigSQLite` |
| E19 | zweites Lesen für den Header | `TestEtagLesenWaehrendAenderung`, `TestEtagEinLesen` |
| E20 | `POST` über Tombstone setzt `1`; `DELETE` setzt die Version `0`; (`DELETE` erhöht die Version nicht: fängt keiner der Tests dieses Slice, denn der ETag der Neuanlage bleibt streng steigend; mindestens sechs vorhandene Tests des Pakets, darunter `TestDeleteWettlauf` und `TestZustandsmatrix`, fangen es) | `TestEtagLoeschenNeuanlage`; vorhandene Tests |
| E21 | Reset ohne `ReplaceAll` | `TestEtagResetStelltHer` |
| E22 | Restore setzt die Version `0` | `TestEtagRestoreStelltHer` |
| E23 | zusätzlicher Header `Cache-Control`; Mitglied vor `@odata.id` | `TestAbnahmeMatrixMitEtag`, `TestEtagAusUndAn` |
| E24 | Zeile der Einstellungen; Beispiel siehe §6 (Abweichung E24) | — |

Nicht gefahren, also ohne roten Beleg: E14 „`Location` fehlt“, E22 „Restore zählt fort“, E13 „Rundlauf“ (die Mutation „Rundlauf bricht“ hat kein Gegenstück im Code), E7 „zufälliger Anteil“. Das zweite Lesen ohne Gleichzeitigkeit (E8, E19) fängt allein der Fake mit wachsender Version (`entities.wachsend`).

**Generatoren und Mengen** (Schritt 14 von `implement-slice`): `EtagFuer`: Schleife 0 bis 300 (301 Elemente) und 9 Literale, dazu `TestEtagFuerKlasse` über die Klasse „nichtnegative 63-Bit-Zahl“: jede Zehnerpotenz 10^k mit 10^k−1 und 10^k+1 (k = 1 bis 18), 2^k mit den Nachbarn (k = 1 bis 62), 2^63−1 und 4000 Zahlen einer Folge mit festem Seed (mindestens 4000 verschiedene Elemente, Zählprüfung im Test); Orakel eine Ziffernrechnung ohne `strconv` (`etagOrakel`). Form von `Entity.Etag`: `TestEtagGenerator` erzeugt aus fünf gültigen Formen durch Löschen, Einfügen, Ersetzen, Verdoppeln und Abschneiden (16 Zeichen, darunter Backslash, `"`, Vorzeichen, Leerraum, NUL, Nicht-ASCII aus Bytewerten) und aus den Zehnerpotenzen und Neunerfolgen von 1 bis 25 Stellen 1827 Formen, Orakel ein regulärer Ausdruck mit `big.Int`-Grenze 2^63-1 (165 gültig, 1662 ungültig); dazu Listen von 60 gültigen und 42 ungültigen Formen an vier Stellen (Entity, drei Plätze in der Collection) in zwei Stufen. `FuehreAus`: 14 Auswahlen × 4 Ordnungen × 2 Filter × 4 `$top` × 5 `$skip` = 1120 Fälle. Arten der Antwort: 10 Arten × 2 Formen des Dienstes × aus und an = 40. `If-Match` und `If-None-Match`: 2 × 3 × 2 × 6 = 72. Schreiben: 15 Läufe aus Vorgeschichte (5) × Wiederholung (3). Reset: 100 Läufe; Abnahme mit ETag: 36 IDs, 92 Schritte.

**Raten der Wirkungs-Fänger** (E18, E19; `AGENTS.md` §3.10). Kein Datenrennen: jeder Zugriff liegt unter einer Sperre des Fakes oder der Transaktion, der Detektor von `make test-race` fängt keine der Mutationen von E18 und E19 (das Lesen vor der Transaktion und das zweite Lesen sind geordnete Zugriffe); der Fänger zählt die Wirkung. Messender: der Implementer, ein `go test -c` je Mutation, dann *N* Prozesse des Test-Binary mit `-test.count=1` im Container (`--cpus=4 --memory=4g`); ein Prozess ist rot, wenn eine seiner Wiederholungen rot ist (E18: 3 Größen *n* × 20 Wiederholungen, E19: 20 Wiederholungen mit 8 Änderungen und 4 Lesern; SQLite: 3 Größen × 5 Wiederholungen). Last des Hosts (Lastmittel der letzten Minute vor und nach dem Lauf): 5,65 bis 12,93, Läufe anderer Slices zugleich.

| Mutation · Test | Prozesse rot / *N* |
|---|---|
| Version vor der Transaktion gelesen · `TestEtagGleichzeitigeAenderungen` (Fake) | 20 / 20 |
| Version vor der Transaktion gelesen · `TestEtagGleichzeitigSQLite` | 10 / 10 |
| Version vor der Transaktion gelesen · `TestEtagGleichzeitigeAenderungen` unter `-race` | 10 / 10 (Wirkung, nicht Detektor) |
| zweites Lesen für den Header · `TestEtagLesenWaehrendAenderung` | 20 / 20 |
| Antwort liest nach `Put` erneut · `TestEtagGleichzeitigeAenderungen` | 0 / 20: in der Transaktion äquivalent; `TestEtagEinLesen` fängt sie ohne Gleichzeitigkeit |
| ohne Mutation: Fake, beide Tests · SQLite | 0 / 20 und 0 / 20 · 0 / 5 |
| ohne Mutation: beide Tests unter `-race` (kein Datenrennen gemeldet) | 0 / 10 |

**Prüfgrenzen** (akzeptierte Negative, `AGENTS.md` §3.15, §3.17; Implementer, Fix-Runde nach dem Review):
- `EtagFuer` ist nur über die Stichproben der Klasse geprüft (oben), nicht über alle 2^63 Versionen. Grund: die Menge ist nicht aufzählbar. Folge im Betrieb: eine Abweichung in einem Band zwischen den Stichproben, das weder Zehnerpotenz, Zweierpotenz, deren Nachbar noch eine der Folgezahlen trifft, bliebe ungefangen; die Funktion ist eine Zeile `strconv.FormatInt` ohne Verzweigung, und der Zähler erreicht im Betrieb eines Testwerkzeugs diese Bänder nicht. Die Mutationen „Band 5001 bis 1999999999 plus 1“, Überlauf bei 2^32, Umwandlung in `int32`, Abschneiden bei 10 Ziffern und „Bit 62 gelöscht“ sind rot (`TestEtagFuerKlasse`).
- Die Zellen *Abbruch (Kontext)* der Zeilen GET Key und GET Collection des Rasters gelten bei `ETAG_ENABLED` unverändert zum Stand ohne ETag (`SPEC-281` Nr. 4 c): der Diff fügt keinen Aufruf eines Ports hinzu, der Kontext erreicht dieselben Ports wie vorher, und die vorhandenen Tests der Lesepfade (`ctx` an jedem Port, beendeter Kontext) laufen mit `Etag` aus. Es gibt keinen Fänger bei `Etag` an; Folge: ein Dienst, der bei an einen eigenen Pfad für einen beendeten Kontext einführte, bliebe ungefangen. Die Aussage „Kontext beendet vor … der Transaktion“ in E16 ist zurückgenommen; gefangen sind die Kontextfehler als Fehler von `Put` und `Advance`.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Trigger der Welle). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-mock-schreiben` (der Zähler `entity_version` wird in der Workspace-Transaktion einer Mutation erhöht, und der Slice liest die Entscheidung dazu aus dessen §6), `slice-mock-lesen` und `slice-odata-json-schreiben` (der ETag erscheint in der Antwort des Lesens, `@odata.etag` ist ein Teil des Schreibers), `slice-snapshot-use-case` (Restore stellt Zähler und ETag wieder her; der Fänger dafür setzt den Use Case voraus). Kein Typ und kein Port dieser Slices liegt heute in `done/`: der Slice ist **nicht** vor `welle-mock` startbar.
- **Binary und Handbuch:** Der Modus `mock` erreicht das Binary erst mit `slice-bootstrap-m3`; das Beispiel zu `ETAG_ENABLED` läuft gegen dieses Binary erst bei `slice-handbuch-mock` (§6), `slice-handbuch-meldungskatalog` liefert das Kataloge-Gate. Beide liegen mit `welle-mock` in `done/`.
- **Was der Slice nicht braucht:** `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-expand-batch` und `welle-odata-v2` (Reihenfolge nach den Muss-Wellen ist Priorisierung, keine Abhängigkeit); `slice-if-match` und `slice-prefer-return` (die Folge-Slices bauen auf diesem, nicht umgekehrt). M2 ist über `welle-mock` Voraussetzung, nicht durch einen eigenen Typ dieses Slice.
- **Stand der Bedingungen beim Start** (Planner, 2026-11-08, `main` = `686612a2`): `welle-mock` liegt in `done/` (M3 erreicht), ebenso `slice-mock-schreiben`, `slice-mock-lesen`, `slice-odata-json-schreiben`, `slice-snapshot-use-case`, `slice-bootstrap-m3` und `slice-handbuch-meldungskatalog`. In `in-progress/` laufen parallel `slice-fallback-policy` (`src/hexagon/services/hybrid/`) und `slice-admin-lebenszyklus` (`src/adapters/driving/admin/`); beide teilen keinen Pfad aus §3, und dieser Slice ändert keinen Pfad der Gate-Kette (Übergangsregel 7, §6), also kein Konflikt. Berührung für später: `slice-if-match` und `slice-prefer-return` bauen auf demselben Pfad und starten nach diesem Slice. Die Verträge schreibt der Architect im Zweig, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).
- **Größenmarke** (Teil des Architect, 2026-11-08, `SPEC-128` Nr. 1): Die Tabelle hat 24 Zeilen, *a* = 2, *b* = 22; erwartet *f* = 0 und *p* = 260 (Aufteilung in §6). Der Block unten trägt `pfad`, `produktcode` und `fakes` (Architect) sowie `basis`, `schaetzung` und `marke` (Planner, 2026-11-08, vor dem ersten Code-Commit gesetzt). Rechnung nach `SPEC-128` Nr. 1: *S* = 85·2 + 80·22 + 200·0 + 260 = 2190; *M* = 100 · ⌈135·2190 / 10000⌉ = 100 · ⌈29,565⌉ = 3000; keine Schnittfrage (*S* < 2223, 33 Zeilen darunter); die vorläufige Prosa-Marke (20 Zeilen, *p* = 300, *S* = 1910) ist damit ersetzt. Basis ist `af1f1bac` (Merge-Stand von `main` beim Start, `git merge-base HEAD main`); der Architect-Commit ändert keinen Pfad des Blocks, die Messung ist vor dem Code 0. *p* = 260 bleibt: die fünf Posten des Architect (Modell und Projektion 30, Dienst 80, Schreiber 50, Konfiguration und Wurzel 60, Hilfen 40) decken alle Pfade des Blocks, und der Wert ist eine Schätzung (±10 % wie bei den gemessenen Slices ergäben *S* 2164 bis 2216, also keine Schnittfrage; erst *p* > 293 löste sie aus). **Schnittfrage ehrlich:** *S* liegt nahe der Grenze, aber *M* = 3000 liegt nicht darüber; ein abtrennbarer Teil ohne geteilten Mechanismus besteht nicht: Konfiguration und Wurzel (`ETAG_ENABLED`, `Dienst.Etag` nur für `4.0`) teilen das Feld `Dienst.Etag` und `EtagFuer` mit Modell, Dienst und Schreiber, ein Teil ohne sie liefert keinen auslieferbaren ETag (kein einzeln lieferbarer Wert), und das Handbuch liegt nicht in den Pfaden des Blocks. Geschnitten wird nicht; überschreitet die Messung 3000, steht ein `ueberschreitung`-Eintrag mit Grund, keine Rückführung nach dem Code. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Slice hat keinen Leser fremder Bytes. Die Pfade von `src/bootstrap/` sind einzelne Dateien, weil das Verzeichnis zu den Zweigen anderer Slices gehört.

```groessenmarke
pfad: src/hexagon/services/mock/
pfad: src/hexagon/model/
pfad: src/adapters/driven/odata/jsoncodec/
pfad: src/adapters/driving/httpodata/
pfad: src/bootstrap/konfig.go
pfad: src/bootstrap/konfig_werte.go
pfad: src/bootstrap/wurzel.go
pfad: src/bootstrap/konfig_test.go
pfad: src/bootstrap/etag_ende_zu_ende_test.go
pfad: src/bootstrap/etag_wurzel_test.go
pfad: src/bootstrap/abnahme_etag_test.go
produktcode: 260
fakes: 0
basis: af1f1bac52851893d572bf96336d8f5bad64c346
schaetzung: 2190
marke: 3000
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Marke ist 3000 (*Größenmarke* oben: *S* = 2190, Basis `af1f1bac`).
- `in-progress` → `open` (blockiert — Carveout?): Form des ETag, Aktivierung oder Startwert des Zählers sind entgegen `SPEC-280` nicht tragfähig (siehe §6), oder `slice-mock-schreiben` hat die Frage nach dem Zähler anders beantwortet, als dieser Plan annimmt.

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

- Ein Schlüssel, der gelöscht und neu angelegt wird, könnte denselben ETag wie die frühere Entity tragen; ein veralteter `If-Match` träfe dann die neue Entity (die Folge zeigt sich in `slice-if-match`) — **Ausgang:** offen (Entscheidung des Architect: nein, `SPEC-280` Nr. 4 c, Fänger E20; der Ausgang *entfallen* steht bei der Closure, wenn E20 rot gesehen ist).
- Der Zähler unter gleichzeitigen Änderungen: zwei Mutationen derselben Entity dürfen nicht denselben Zähler erzeugen; der Detektor meldet nur, was läuft — **Ausgang:** offen; Wirkungs-Fänger E18 und E19, der die verschiedenen ETags nach N gleichzeitigen Änderungen zählt, mit Rate je Mutation und je Lauf von `go test` samt Zahl der Prozesse und Messendem (`AGENTS.md` §3.10; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` 2×, `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` 1×).
- Restore und Reset: der Zähler des Zustands, nicht ein fortlaufender, sonst unterscheidet sich der ETag nach Reset von dem des ersten Laufs (`LH-QA-02`) — **Ausgang:** offen (Entscheidung des Architect: `SPEC-280` Nr. 4 d, Fänger E21 und E22).
- Fakes der Ports (Repository, Schreiber) halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×); E8 setzt einen `Entities`-Fake voraus, der bei jedem `Get` eine höhere Version liefert und jeden Aufruf zählt — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten ist gefüllt (Architect, 2026-11-08); jede Zelle nennt ihren Ort oder „gilt nicht“ mit Grund; die Zusatztabelle führt die Modi — **Ausgang:** offen bis zur Bestätigung des Reviewers (eine Zelle ohne Ort ginge zurück an den Architect).
- Größe: die Tabelle hat 24 Zeilen (*a* = 2, *b* = 22), der Architect erwartet *f* = 0 (E8 erweitert den vorhandenen Fake von `Entities`, kein neuer) und *p* = 260 (Modell und Projektion etwa 30, Dienst 80, Schreiber 50, Konfiguration und Wurzel 60, Hilfen 40); *S* = 85·2 + 80·22 + 200·0 + 260 = 2190, keine Schnittfrage (*S* < 2223); der Planner setzt *basis* `af1f1bac`, *schaetzung* 2190 und *marke* 3000 vor dem ersten Code-Commit (§4), *p* unverändert; ein Teil ohne geteilten Mechanismus besteht nicht (§4) — **Ausgang:** offen bis zur Messung beim Handoff (Messpunkt-Zeile von `make groessenmarke` in den Plan; über 3000 nur mit Eintrag samt Grund). Messung des Implementers beim Handoff (Stand `Handoff`, Commit `7783f4f5`): `make groessenmarke` meldet `Messpunkt slice-etag-versionen.md: zeilen=24 nebenlaeufig=2 fakes=0 produktcode=260 schaetzung=2190 marke=3000 messung=1954` und `ok, Messung 1954 von Marke 3000` (Auslastung 0,65; 0,89 der Schätzung 2190); der von Hand gemessene Produktcode ist 110 (`git diff --numstat af1f1bac` über `etag.go`, `entity.go`, `abfrage.go`, `deuten.go`, `lesen.go`, `mock.go`, `schreiben.go`, `schreiber_etag.go`, `schreiber_bau.go`, `konfig.go`, `konfig_werte.go`, `wurzel.go`) gegen *p* = 260; kein Eintrag `ueberschreitung`. Außerhalb der Pfade des Blocks liegen die Tests `wurzel_test.go`, `export_test.go` (Bootstrap) und `jsoncodecv2/schreiber_etag_test.go`, ungezählt. Die Messpunkt-Zeile in den Plan trägt der Planner (`SPEC-128` Nr. 9).
- Pfad `src/bootstrap/` (Konfiguration, Wurzel) kam durch den Schlüssel `ETAG_ENABLED` hinzu (§3); er liegt in keinem Pfad der Gate-Kette, kollidiert aber mit Zweigen, die dieselben Dateien ändern (`slice-fallback-policy`, `slice-admin-lebenszyklus`: Verdrahtung in `wurzel.go`) — **Ausgang:** offen; Adresse: Planner (Reihenfolge der Merges, Konflikt in `wurzel.go` und `konfig.go` absehbar, klein). Planner, 2026-11-08: Merge-Reihenfolge `slice-fallback-policy`, dann `slice-admin-lebenszyklus`, dann dieser Slice; wer später merged, zieht `main` vor dem Handoff in den Zweig und löst die Konflikte in `wurzel.go` und `konfig.go` dort. Die Basis `af1f1bac` bleibt; was ein gemergtes `main` in den sieben Dateien von `src/bootstrap/` an Zeilen Fremder bringt, zählt die Messung mit (`SPEC-128` Nr. 4 g), und der Eintrag `ueberschreitung` nennt es dann als Grund. Sendung an `slice-if-match` (`open/`): in dessen §1 angenommen (Commit `a33588f0`), Geber `slice-etag-versionen`.
- `slice-if-match` erbt drei Entscheidungen (`SPEC-280` Nr. 1 und 3): die schwache Form `W/"<n>"` (Vergleich von `If-Match` legt dort der Architect fest), `If-Match` bei aus ohne Wirkung, und dass `SPEC-192` *Ungeprüft* dort fällt; die Sendung steht im selben Commit in §1 von `slice-if-match` (`AGENTS.md` §3.13) — **Ausgang:** offen; Adresse: Planner von `slice-if-match` (nimmt die Sendung an oder gibt sie zurück).
- Akzeptierte Negative (Architect, Grund und Folge in der Spezifikation, keine Folgepflicht): ein ETag aus der Geschichte vor Reset oder Restore kann einen anderen Zustand treffen (`SPEC-280` Nr. 4, `LH-QA-02` geht vor, Reset und Restore sind Handlungen des Testers); ein Dienst der Version `2.0` liefert keinen ETag (`SPEC-280` Nr. 2, er ist `NurLesend`) — **Ausgang:** entfallen als Risiko, weil entschieden und im Vertrag benannt.
- Gate-Kette (Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch`): der Slice ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).
- Abweichung E24 (Implementer, 2026-11-08; entschieden vom Architect, 2026-11-08): das gebundene Handbuch-Beispiel „Start mit `ETAG_ENABLED=true`, Lesen, Ändern, neuer ETag“ ist im Format von `SPEC-200` Nr. 6 nicht ausdrückbar (leere Umgebung, kein Listener; Nr. 6 d: wer ein Kapitel mit einer solchen Eingabe braucht, erweitert diese Nummer). Entscheidung: das Beispiel ist aus E24 und der DoD gestrichen, `SPEC-200` wird in diesem Slice nicht erweitert (die kleinere Änderung; der Slice bleibt lieferbar). Dieser Slice trägt die Zeile `ETAG_ENABLED` der Einstellungen (Schlüsselkatalog, `make katalog-check`); das Beispiel ist Teil von `slice-handbuch-mock`, das die Umgebung im Beispiel-Format und einen Lauf mit Signal ohnehin liefert (derselbe Befund wie bei `slice-bootstrap-m3`). Die Aussage, `ETAG_ENABLED=true` liefere ETags, steht im Handbuch nur, soweit ein gebundenes Beispiel sie trägt (`SPEC-200` Nr. 1), sonst als Zeile der Einstellungen ohne Wirkungsaussage — **Ausgang:** entschieden; die Sendung ist offen, bis der Nehmer sie annimmt; Adresse: Planner (trägt im Slice `slice-handbuch-mock` in `open/` auf `main` in §1 *Übernimmt* und in der DoD ein: gebundenes Beispiel zu `ETAG_ENABLED` — Start mit `ETAG_ENABLED=true`, Lesen, Ändern, neuer ETag, Geber `slice-etag-versionen`; gelesen: dessen §1 *Ausdrücklich NICHT* schließt es nicht aus, der Slice liegt in `open/`, kein Zusatz der Spezifikation nötig außer der Erweiterung von `SPEC-200`, die er ohnehin führt; geht der Nehmer nicht an, bleibt das Beispiel eine offene Beobachtung des Planners).
- Abweichung E6 (Implementer, 2026-11-08; entschieden vom Architect, 2026-11-08): der Fänger kann „der Dienst liest beide Header nie“ nicht über einen Zähler der Zugriffe fangen; `inbound.Anfrage.Header` ist eine einfache Map. Entscheidung: akzeptiertes Negativ, Grund und Folge stehen in `SPEC-280` Nr. 3 (Port nicht geändert: ein Typ mit Haken wäre eine Änderung aller Adapter für einen Fall, den die Wirkung ohnehin deckt); gefangen ist die Wirkung (72 Fälle: Antwort und Zustand gleich der Anfrage ohne Header). Ein Dienst, der einen Header liest und folgenlos verwirft, bleibt ungefangen; er ändert nichts, was ein Client sieht — **Ausgang:** entfallen (entschieden als akzeptiertes Negativ, keine Folgepflicht).
- Randform ohne ausdrücklichen Ort (Implementer, 2026-11-08; entschieden vom Architect, 2026-11-08): die Lesart des Implementers gilt — die Konfiguration trägt den gelesenen, gültigen Wert von `ETAG_ENABLED` in jedem Modus (wie `STRICT_COVERAGE`), der Composition Root liest ihn nur im Modus `mock`; `SPEC-110` Nr. 15 sagt es seither ausdrücklich. Fänger bleibt `TestKonfigEtagEnabled` (3 Modi × 10 Werte), er prüft den Wert in allen drei Modi; keine Änderung am Code — **Ausgang:** entfallen (entschieden, kein Test-Auftrag).

**Raster *Operation × Argumentart*** (Architect, 2026-11-08; jede Zelle: Ort im Vertrag oder „gilt nicht“ mit Grund; alle Zellen gelten bei an, die Spalte *Unterstützung nicht aktiviert* nennt aus):

| Operation | Entity vorhanden | Entity fehlt oder Tombstone | Unterstützung nicht aktiviert | Abbruch (Kontext) | Wiederholung |
|---|---|---|---|---|---|
| GET Key | `SPEC-281` Nr. 1 (Header und Mitglied, eine Version) | `SPEC-281` Nr. 4 a (Fehler von `SPEC-196` Nr. 7, ohne ETag) | `SPEC-280` Nr. 2 (weder Header noch Mitglied) | `SPEC-281` Nr. 4 c (Fehler ohne Code, keine Antwort) | `SPEC-280` Nr. 5 (derselbe ETag) |
| GET Collection (je Element) | `SPEC-281` Nr. 1 und 3 (Mitglied je Element, kein Header) | `SPEC-281` Nr. 4 a (Tombstone erscheint nicht; fehlend gibt kein Element) | `SPEC-280` Nr. 2 | `SPEC-281` Nr. 4 c | `SPEC-280` Nr. 5 |
| POST | `SPEC-193` Nr. 2 (`ODR-E6009`, kein ETag, Zähler unverändert; `SPEC-281` Nr. 2) | `SPEC-281` Nr. 2 (`1` ohne Vorgänger, Tombstone + 1, `SPEC-280` Nr. 4 c) | `SPEC-280` Nr. 2 (Zähler läuft, `201` ohne ETag) | `SPEC-192` Nr. 7 (Rollback), `SPEC-281` Nr. 4 c | `SPEC-193` Nr. 2 (zweiter `POST` desselben Schlüssels ist `ODR-E6009`) |
| PATCH, PUT | `SPEC-281` Nr. 2 (ETag der neuen Version, auch bei `{}`) | `SPEC-193` Nr. 2 (`ODR-E6008`, `ODR-E6003`, kein ETag; `SPEC-281` Nr. 2) | `SPEC-280` Nr. 2 (`204` ohne ETag) | `SPEC-192` Nr. 7, `SPEC-280` Nr. 4 b (Zähler unverändert) | `SPEC-280` Nr. 4 a (jede erfolgreiche Wiederholung ergibt einen neuen ETag) |
| DELETE, danach POST desselben Schlüssels | `SPEC-281` Nr. 1 (`DELETE`: `204` ohne ETag; Tombstone mit Version + 1, `SPEC-193` Nr. 3 e) | `SPEC-280` Nr. 4 c (`POST` über Tombstone: Tombstone + 1; `DELETE` auf Tombstone: `ODR-E6008`) | `SPEC-280` Nr. 2 | `SPEC-192` Nr. 7 | `SPEC-193` Nr. 2 (zweites `DELETE` ist `ODR-E6008`, zweites `POST` `ODR-E6009`) |
| Reset, Restore | `SPEC-280` Nr. 4 d (Version des Zustands: `0` bzw. des Manifests) | `SPEC-280` Nr. 4 d (Reset entfernt Tombstones, Restore stellt sie mit ihrer Version her, `SPEC-129` Nr. 17) | `SPEC-280` Nr. 2 (der Zähler wird wiederhergestellt, ausgeliefert wird nichts) | `SPEC-129` Nr. 11 (eine Transaktion, bei Abbruch unverändert) | `SPEC-129` Nr. 10 (Reset ist für den Inhalt des Entity-State idempotent), `SPEC-280` Nr. 4 d |

**Zusatztabelle (Modus der Zusage)** (`AGENTS.md` §3.16: die Zusage hängt vom Modus ab; je Modus eine Zeile mit Fänger):

| Modus | Zusage | Fänger |
|---|---|---|
| Dienst ohne das Feld (Nullwert) und `ETAG_ENABLED=false` | kein ETag, Antworten wie bisher (`SPEC-280` Nr. 2) | E5, E3 und die 92 Schritte der Abnahme |
| `ETAG_ENABLED=true`, Dienst der Version `4.0` | ETag nach `SPEC-281` | E7 bis E15, E23 |
| `ETAG_ENABLED=true`, Dienst der Version `2.0` | kein ETag (`SPEC-280` Nr. 2) | E4 |
| `ETAG_ENABLED=true` in den Modi `record` und `replay` | keine Wirkung (`SPEC-110` Nr. 15) | E3 |
| `If-Match` oder `If-None-Match` gesetzt, aus und an | ohne Wirkung (`SPEC-280` Nr. 3) | E6 |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice ist nebenläufig (Zähler unter gleichzeitigen Änderungen): `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) gelten für die Rate des Wirkungs-Fängers und für Zweige, die kein Lauf ausführt; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für die Fakes; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Verkörpert: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (`AGENTS.md` §3.10, Sensor `make test-race`). Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
