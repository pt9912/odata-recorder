# Slice slice-expand-ausfuehrung: `$expand` im Mock ausführen und im OData-JSON ausgeben

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-FA-11`, `LH-QA-02`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-265` (Ausführen im Mock) · `SPEC-264` Nr. 3, 4 und `SPEC-266` Nr. 1, 2, 5 (als Aufrufer) · `LH-FA-16.a`, `LH-FA-11.a` · `SPEC-044`, `SPEC-262`, `SPEC-263` · die Zeiger in `SPEC-192`, `SPEC-193`, `SPEC-195` bis `SPEC-198`, `SPEC-290` · `SPEC-049` (Zeile `ODV4-E-03` c), `SPEC-042` Nr. 5 · `ARC-001`, `ARC-004`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

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

**Ziel:** Der Mock führt `$expand` aus (`SPEC-265`): er verknüpft die Entities nach `model.Beziehungen` über die Referenzbedingungen des `$metadata` (`SPEC-264`) und setzt sie als Erweiterungen, die der Schreiber eingebettet ausgibt (`SPEC-266`). Eine leere Navigation ergibt `[]` oder `null`, ein nicht belegbares Ergebnis `ODR-E6003`, eine nicht auflösbare Navigation `ODR-E6001`, und derselbe Zustand ergibt dieselbe Antwort (`LH-QA-02`).

**Übernimmt (Interpreter und Ausgabe):** Navigationen der Abfrage — `slice-query-interpreter` schließt `$expand` und Navigationen aus und nennt `welle-expand-batch` (Sendung von `slice-query-interpreter`, `AGENTS.md` §3.13); eingebettete Entities in der Antwort des Mock — `slice-odata-json-schreiben` schließt sie aus und nennt `welle-expand-batch` (Sendung von `slice-odata-json-schreiben`). Beide nimmt dieser Slice an. Den Schreiber der eingebetteten Entities, die Datenbasis im Modell und den Serializer-Contract-Test (auch Sendung von `slice-expand-parser`) hat er beim Schnitt vor dem Code an [`slice-expand-bausteine`](slice-expand-bausteine.md) weitergegeben (§1 dort); hier bleibt, dass der Dienst sie ruft und füllt.

**Pflichten als Aufrufer** der Bausteine aus [`slice-expand-bausteine`](slice-expand-bausteine.md), mit den Nummern des Vertrags:

- `model.Beziehungen` läuft erst nach `PruefeExpand` (`SPEC-264` Nr. 3, `SPEC-265` Nr. 2); `*model.ExpandFehler` wird abgebildet, `model.ErrUngueltig` ist `ODR-E9001` (`SPEC-264` Nr. 4). Zeile `C3`.
- Allein der Mock setzt `Erweiterungen`, nach `FuehreAus` (`SPEC-266` Nr. 1); sie bestehen die Prüfung von `SPEC-266` Nr. 2 (Name, Art, Set, höchstens ein Element ohne `Sammlung`, keine zweite Ebene). Zeilen `C12`, `C19`.
- An der Version 2.0 setzt der Dienst keine (`SPEC-266` Nr. 5). Zeile `C1`.

**Übernimmt (Verdrahtung, Sendung von `slice-expand-parser` und `slice-expand-validator`, `AGENTS.md` §3.13):** das Einschalten von `$expand` im Parser (`urlparser.WithExpand`), den Aufruf von `model.PruefeExpand` vor der Ausführung, die Abbildung seiner Fehler und die Änderung der Fixtures, deren Antwort sich damit ändert: `ODV4-E-03` c fragt `$expand=*` und bleibt `501` (`SPEC-049`, `abnahme_matrix_test.go`; `$expand=Foo` wäre danach `400`), `SPEC-042` Nr. 5 und `tests/integration/vertrag_fehlerklassen_test.go` fragen `$expand=GEHEIM-Q(` und bleiben `501` (`$expand=GEHEIM-Q` wäre danach `400`), und `TestSatzVersion4OhneExpand` in `src/bootstrap/expand_wurzel_test.go` hält statt der Abweisung das Lesen fest. Spezifikation, Werkzeug-Vertrag und Tests ändern sich im selben Commit (`AGENTS.md` §3.9). Bis zur Verdrahtung antwortet das Binary auf `$expand` mit `501`.

**Übernimmt (Sendung von `slice-expand-validator`, `AGENTS.md` §3.13):** die Navigation ohne `NavigationBinding` und die fehlenden Daten einer Navigation. Ohne Binding ist sie nicht auflösbar (`SPEC-264` Nr. 3 a): die Bestimmung liegt in [`slice-expand-bausteine`](slice-expand-bausteine.md), die Antwort `501` hier (`SPEC-265` Nr. 2); fehlende Daten entscheidet `SPEC-265` Nr. 5.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Parser und Validator von `$expand` — `slice-expand-parser`, `slice-expand-validator` liegen in `done/`; dieser Slice liest den geprüften Knoten.
- Die Angaben `Partner` und `Bedingungen` im Metadatenleser, `model.Beziehungen` (`SPEC-264`), `model.Erweiterung` und der Schreiber eingebetteter Entities (`SPEC-266`) — [`slice-expand-bausteine`](slice-expand-bausteine.md), vor diesem (§4).
- Import, Entity-State, Schema, Leser der Antworten — die Datenbasis ändert keinen davon (`SPEC-264` Nr. 1).
- Verschachtelte Optionen, Pfade, `*`, `$ref`, Tiefe über eins — der Parser weist sie mit `501` ab (`SPEC-262` Nr. 3 b, 4).
- Navigation als Pfadsegment, Schreiben von Navigationen (`@odata.bind`), Deep Insert — kein Gegenstand von `LH-FA-16`; `501` (`SPEC-049`).
- `$apply`, `$search`, Actions, Functions — bleiben `501` (Lastenheft §5).
- `$expand` für Services der Version 2.0 — `SPEC-173` führt die Option nicht; `501`.
- Aufzeichnen und Abspielen von `$expand`-Requests — Record und Replay behandeln sie als HTTP, unabhängig vom Parser (`LH-FA-16.a`).

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

- [ ] Ausführen samt Verdrahtung (`SPEC-265` Nr. 1 bis 5, 8, 9; Sendungen aus §1): Satz der Version 4 liest `$expand`, Prüfung und Bestimmung vor jedem Zugriff auf Daten, Coverage und Lesen der Zielmengen, Auflösen, Fixtures nach §1.
- [ ] Antwort (`SPEC-265` Nr. 6, 7, 10): eingebettete Entities (Objekt, `null`, Array), Property-Maske und ETag der verknüpften Entities, Header `X-Recorder-Coverage`, `@odata.context` unverändert, dieselben Bytes bei gleichem Zustand.
- [ ] Executor-Contract-Test für `$expand` (`LH-FA-16.a`) und ein Test im Modus `mock` über die Abnahme der Welle: verknüpfte Daten sichtbar, leere Navigation leer.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary), darin der unterstützte Umfang (`LH-RB-01`): welche Navigationen auflösbar sind und dass die Zielmenge selbst aufgezeichnet sein muss; das Kataloge-Gate ist grün.
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
| `spec/spezifikation.md` | update | Vertrag `SPEC-265` (Architect, geschrieben); Zeile `ODV4-E-03` c mit der Verdrahtung |
| `src/hexagon/services/mock/` | update | Zeilen C: Ausführen (`SPEC-265`) |
| `src/bootstrap/`, `tests/integration/vertrag_fehlerklassen_test.go`, `harness/vertraege/SPEC-042-test-contract.md` | update | Verdrahtung, Abnahme im Modus `mock`, Fixtures |
| `docs/user/` | update | Handbuch-Kapitel `$expand` |

**Bestehende Tests, die mitziehen** (geprüft am Stand `40546730`): `TestStellenImQuelltext` (`src/hexagon/services/mock/schreiben_weitere_test.go`) zählt 22 Aufrufstellen der Ports im Quelltext des Pakets und wird mit S27 und S28 rot, bis die Listen der Fänger für Kontext- und Portfehler nachgezogen sind; die Anpassung prüft mindestens so streng wie vorher (`BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger`).

**Entscheidungen zum Vertrag** (Architect; Ort in der Spezifikation, keine ADR berührt): Coverage der Zielmengen (`SPEC-265` Nr. 3); Tombstone, `null`, fehlendes Ziel, Reihenfolge (`SPEC-265` Nr. 5); verknüpfte Entities, Header, Kontext (`SPEC-265` Nr. 6, 7); Mutationen pflegen nichts, die Beziehung folgt den Werten (`SPEC-264` Nr. 1). Ein neuer Fehlercode und eine neue Ursache entstehen nicht. Die Entscheidungen zu Datenbasis und Form der Ausgabe stehen in [`slice-expand-bausteine`](slice-expand-bausteine.md) §3.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, aus `SPEC-265` und dem Raster in §6; `C` Ausführen, die Zeilen `A` und `B` des ungeschnittenen Plans führt [`slice-expand-bausteine`](slice-expand-bausteine.md); Testnamen sind Vorgaben, die der Implementer im Plan nachzieht; Orakel und Fixture-Modell entstehen im Test ohne Code der Funktion):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| C1 | Der Satz der Version 4 liest `$expand`, der der Version 2.0 nicht (`SPEC-265` Nr. 1) | `expand_wurzel_test.go`, neu gefasst, je Version | `WithExpand` fehlt im Aufbau | andere |
| C2 | `$expand` an `$metadata`, `/$count` (`GET`, `HEAD`) und an den vier Mutationen ist `501` ohne weiteren Port; Methode, Paarung und `$filter` gehen bei Mutationen vor (`SPEC-265` Nr. 1) | `TestExpandRessourcen`: alle Methoden × alle Arten | Eine Mutation ignoriert `$expand` | andere |
| C3 | Namen vor Beziehungen, beide vor Schlüssel, Coverage und Entities; `400` vor `501` an jeder Stelle (`SPEC-265` Nr. 2) | `TestExpandPruefungVorDaten`: Fake zählt die Ports | `Beziehungen` läuft vor `PruefeExpand` | andere |
| C4 | Coverage je Zielmenge einmal; strikt ist eine nicht vollständige Zielmenge einer Collection-Navigation `422` vor jedem `ListSet`, bei der Collection vor der Ausführung, beim Lookup nach ihm; Navigation auf eine Entity ohne Vorprüfung (`SPEC-265` Nr. 3) | `TestExpandCoverageZiel`: Zustände × Modus × Art × Collection und Lookup × gleiches und fremdes Set | Die Vorprüfung entfällt | andere |
| C5 | Je Zielmenge ein `ListSet`, die Liste des eigenen Sets wird in der Collection wiederverwendet, auch bei leerem Ergebnis (`SPEC-265` Nr. 4) | `TestExpandListSet`: zwei Navigationen auf ein Set, Selbstbezug, leere Seite | Je Navigation ein Aufruf | andere |
| C6 | Gelöschte Einträge erscheinen nie; ein verdorbener Eintrag oder fehlender Schlüssel einer Zielmenge ist `ODR-E9001` (`SPEC-265` Nr. 4) | `TestExpandZielEintraege` | Tombstones werden verknüpft | andere |
| C7 | Auflösen über die Werte vor der Projektion; `$select` ohne Fremdschlüssel ändert die Verknüpfung nicht (`SPEC-265` Nr. 5) | `TestExpandSelect` | Die projizierte Entity wird verknüpft | andere |
| C8 | Unbekannte oder redigierte `Quelle` (Entity-Navigation) bzw. `Ziel` (Collection, jeder Eintrag, auch bei leerem Ergebnis) ist `422` in beiden Modi (`SPEC-265` Nr. 5 a) | `TestExpandUnbekannt`: unbekannt und redigiert × Seite × Modus | Die unbekannte Property gilt als `null` | andere |
| C9 | Passen: alle Paare nicht `null` und gleich; Collection nach Schlüssel geordnet, keine Treffer `[]`; zusammengesetzter Schlüssel, Selbstbezug (`SPEC-265` Nr. 5 b, c) | `TestExpandPassen`: Generator über Daten mit 1 und 2 Paaren, `null`, Reihenfolge von `ListSet` | `null` passt zu `null` | andere |
| C10 | Eine Entity ohne Treffer: `Quelle` `null` ist `null`; Coverage `complete` ist `null`; Tombstone des Schlüssels ist `null`; sonst `422`, in beiden Modi (`SPEC-265` Nr. 5 d) | `TestExpandOhneTreffer`: Fälle × Zustände × Modus | Ohne Treffer immer `null` | andere |
| C11 | `$filter`, `$orderby`, `$top`, `$skip`, `$count` wirken allein auf das Set der Anfrage (`SPEC-265` Nr. 6) | `TestExpandOptionen` | `$top` kürzt die eingebettete Collection | andere |
| C12 | Verknüpfte Entity: bekannte Properties, nie `null` für unbekannte, nichts redigiertes, partiell ohne Fehler, ETag bei an (`SPEC-265` Nr. 6) | `TestExpandVerknuepfteEntity`: ETag an und aus | Ein partieller Eintrag ist `422` | andere |
| C13 | Header: Collection schwächster Zustand aus Set und Collection-Zielen, Entity nur mit Collection-Navigation, Entity-Ziele zählen nicht; Kontext und Zahl unverändert (`SPEC-265` Nr. 7) | `TestExpandHeader`: Zustände × Zustände × Art der Navigation × Collection und Lookup | Der Header nennt nur das Set der Anfrage | andere |
| C14 | S27 und S28: Argumente, `ctx`, Fehlerabbildung, Kontextfehler ohne weiteren Aufruf, kein Schreiben (`SPEC-265` Nr. 8) | `TestExpandAufrufstellen`: Schleife über die Stellen × Fehlerarten | S28 bildet mit Gegenstand `Sonstiges` ab | andere |
| C15 | Der Span nennt die Coverage des Sets der Anfrage (`SPEC-265` Nr. 8) | `TestExpandSpan` | Die Coverage des Ziels überschreibt das Attribut | andere |
| C16 | Rangfolge nach `SPEC-265` Nr. 9 | `TestExpandRang`: die Paare der Zusatztabelle in §6 | Auflösen läuft vor dem Deuten aller Zielmengen | andere |
| C17 | Dieselbe Anfrage gegen denselben Zustand ergibt dieselben Bytes; `HEAD` gleicht `GET` (`SPEC-265` Nr. 10) | `TestExpandWiederholung`: permutierte `ListSet`-Antworten | Die eingebettete Folge ist die von `ListSet` | andere |
| C18 | Kein gemeinsamer Stand über Mengen und keine Transaktion: akzeptiertes Negativ, Grund und Folge in `SPEC-265` Nr. 10; kein Zustand im Dienst | `kein_zustand_test.go` (Zustand); Fake zählt `Transaktion` null | Der Dienst merkt sich Beziehungen zwischen Aufrufen | andere |
| C19 | Im Modus `mock` liefert `$expand` verknüpfte Daten und eine leere Navigation leer; die Fixtures aus §1 bleiben `501` (`LH-FA-16`) | Abnahmetest in `src/bootstrap/` am HTTP-Eingang; `make test-contract` | Der Dienst setzt `Erweiterungen` nicht | andere |
| C20 | Die Laufzeit der Verknüpfung ist nicht zugesagt: akzeptiertes Negativ (`SPEC-265` *Ungeprüft*); Grund: `LH-QA-01.a` misst ohne `$expand`; Folge: eine quadratische Verknüpfung bleibt bei grünen Gates unbemerkt, Maß 0 von 1 Klasse gefangen | Review des Diffs | Verknüpfung durch doppelte Schleife | andere |

**Größenmarke, Teil des Architect** (Pfade der Zeilen `C`; `produktcode` ist der Anteil, den der Planner beim Schnitt aus den 590 Zeilen des Architect gesetzt hat, §4):

```groessenmarke
pfad: src/hexagon/services/mock
pfad: src/bootstrap
pfad: tests/integration/vertrag_fehlerklassen_test.go
produktcode: 290
fakes: 0
```

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Trigger der Welle), `slice-expand-parser`, `slice-expand-validator` und [`slice-expand-bausteine`](slice-expand-bausteine.md) liegen in `done/`. Der Vertrag samt Randformen steht in `SPEC-264` bis `SPEC-266`.

- **Voraussetzungen, geprüft am Stand des Vertrags:** `welle-mock` liegt in `done/` (Interpreter, Schreiber, `MockUseCase`, Coverage, Workspace-Zugriff vorhanden); Parser (`urlparser.WithExpand`) und Validator (`model.PruefeExpand`) liegen vor und sind nicht verdrahtet. Datenbasis im Modell und Schreiber der eingebetteten Entities liefert [`slice-expand-bausteine`](slice-expand-bausteine.md).
- **Binary und Handbuch:** Modus `mock` und Kataloge-Gate liegen vor.
- **Was der Slice nicht braucht:** `slice-batch-parser`, `slice-batch-validator`, `slice-batch-changesets`.
- **Schnitt vor dem Code** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1; Planner, 2026-11-28, nach dem Vorschlag des Architect): Der ungeschnittene Plan hatte 40 Zeilen und `produktcode` 590, *S* = 3790, über der Schnittgrenze. Abgetrennt sind die Zeilen `A` und `B` als [`slice-expand-bausteine`](slice-expand-bausteine.md) (*S* = 1900), hier bleiben die Zeilen `C`; die Teile teilen keinen Mechanismus, die Naht sind die Typen `model.Beziehung` und `model.Erweiterung`. Den `produktcode` hat der Planner als 300 und 290 verteilt (Setzung des Planner, der Architect nennt nur die Summe).
- **Größenmarke:** 20 Zeilen, *a* = 0, *b* = 20, *f* = 0, *p* = 290, *S* = 1890, Marke *M* = **2600**; keine Schnittfrage. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Block trägt `basis`, `schaetzung` und `marke` noch nicht (`make groessenmarke` meldet *Marke steht aus*); der Planner setzt sie beim Start, vor dem ersten Code-Commit.
- **Kollisionen** (geprüft am Stand `40546730`): Unter dem Verzeichnis `src/bootstrap` des Blocks nennt der Block von `slice-v2-rundlauf` die Datei `src/bootstrap/v2_ursache_test.go` (`SPEC-128` Nr. 11, noch nicht im Programm); `slice-batch-changesets` (noch ohne Block) schreibt Dienst und Composition Root. Der Planner prüft beim Start, dass keiner der beiden zugleich läuft. Kein Kettenslice.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 2600 Zeilen (Metrik wie oben). Eine Überschreitung schneidet nicht (`SPEC-128` Nr. 10 c); sie bekommt einen Eintrag `ueberschreitung` mit Grund im Block.
- `in-progress` → `open` (blockiert): eine Randform ohne Ort in `SPEC-264` bis `SPEC-266`, die der Architect nicht im laufenden Slice entscheiden kann.

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

- Der Vertrag verlangt einen vorgelagerten Slice (Änderung an `SPEC-125`, dessen liefernder Slice in `done/` liegt) und einen Schnitt nach der Marke (§4) — **Ausgang:** eingetreten: [`slice-expand-bausteine`](slice-expand-bausteine.md); [welle-expand-batch](../welle-expand-batch.md) §2, §4, §5 und §6 sind nachgezogen.
- Der unterstützte Umfang ist eng: Beziehungen nur aus `ReferentialConstraint` und `Partner` des `$metadata` samt Binding; eingebettete Entities aufgezeichneter `$expand`-Antworten werden nicht Daten des Mock (`SPEC-264` Nr. 1). Wer allein mit `$expand` aufzeichnet, muss die Zielmengen zusätzlich aufzeichnen — **Ausgang:** offen; Adresse: User (fachlicher Owner). Frage: reicht dieser Umfang für `LH-FA-16`? Empfehlung des Architect: ja; ein Lernen aus Antworten wäre ein eigener Vorgang (Import, Ablage, Schema), erst bei einem benannten Dienst, den der Umfang nicht trägt. Der Planner legt die Frage vor dem Start von [`slice-expand-bausteine`](slice-expand-bausteine.md) vor.
- Die Fixtures `ODV4-E-03` c und die Klasse *nicht unterstützte Query-Option* (`SPEC-042` Nr. 5) gehören zur Abnahme der Erstversion; sie ändern allein die Query, Status und Code bleiben (§1). Sie ziehen allein mit der Verdrahtung in diesem Slice, [`slice-expand-bausteine`](slice-expand-bausteine.md) ändert keine. Läuft die Release-Arbeit an diesen Dateien, startet die Verdrahtung danach — **Ausgang:** offen; Planner.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Raster und Zusatztabelle unten; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*:**

| Operation | Fehlender, leerer, unbekannter, fremder Wert | Navigation leer | Ziel fehlt oder gelöscht | Ziel nicht nachweislich vollständig | Selbstbezug, Zyklus, Tiefe | Property-Maske, abgeleiteter Typ | Abbruch (Kontext), Wiederholung |
|---|---|---|---|---|---|---|---|
| Auflösen: Navigation auf eine Entity | `null` und unbekannt: `SPEC-265` Nr. 5 a, d | `SPEC-265` Nr. 5 d | `SPEC-265` Nr. 5 d | `SPEC-265` Nr. 5 d | eine Ebene: `SPEC-265` Einleitung; Selbstbezug Nr. 4 | `SPEC-265` Nr. 5 a, 6 | `SPEC-265` Nr. 8, 10 |
| Auflösen: Navigation auf eine Collection | `SPEC-265` Nr. 5 a, b | `SPEC-265` Nr. 5 c | gelöschte erscheinen nie: `SPEC-265` Nr. 4 | `SPEC-265` Nr. 3, 7 | wie oben | `SPEC-265` Nr. 5 a, 6 | `SPEC-265` Nr. 8, 10 |
| Lesen einer Collection mit `$expand` | `SPEC-265` Nr. 2 | leeres Ergebnis liest die Zielmengen: `SPEC-265` Nr. 4 | `SPEC-265` Nr. 5 | `SPEC-265` Nr. 3, `SPEC-197` Nr. 2 | `SPEC-265` Nr. 4 | `$select`: `SPEC-265` Nr. 5, 7; abgeleitet `SPEC-266` Nr. 3 | `SPEC-265` Nr. 8, 10 |
| Lesen einer Entity mit `$expand` | `SPEC-265` Nr. 2; Schlüssel `SPEC-196` Nr. 7 | gilt nicht: die Anfrage trifft eine Entity | Entity der Anfrage `SPEC-196` Nr. 7; Ziel `SPEC-265` Nr. 5 | `SPEC-265` Nr. 3, 7 | `SPEC-265` Nr. 4, 10 | `SPEC-196` Nr. 7 e, `SPEC-265` Nr. 6 | `SPEC-265` Nr. 8, 10 |
| `$metadata`, `/$count`, Mutationen, Version 2.0 | `SPEC-265` Nr. 1 | gilt nicht: keine Ausführung | gilt nicht: keine Ausführung | gilt nicht: keine Ausführung | gilt nicht: keine Ausführung | gilt nicht: keine Ausführung | `SPEC-265` Nr. 1: kein Port |

Die Zeilen des Rasters zu den Angaben des `$metadata`, zur Bestimmung der Beziehung und zur Ausgabe führt [`slice-expand-bausteine`](slice-expand-bausteine.md) §6.

**Zusatztabelle** (zwei Verstöße an einer Anfrage und Modi, `SPEC-265` Nr. 9):

| Paar oder Modus | Ergebnis | Fänger |
|---|---|---|
| Form von `$expand` und Methode nicht ausgeführt | Fehler des Lesens der URL | `C16` |
| Mutation mit `$filter` und `$expand` | `ODR-E6002` | `C2` |
| unbekannte und nicht auflösbare Navigation | `ODR-E6005` | `C3` |
| nicht auflösbare Navigation und ungültiger Schlüssel | `ODR-E6001` | `C16` |
| Set der Anfrage und Zielmenge unvollständig, strikt | `ODR-E6003` ohne Lesen der Zielcoverage | `C4` |
| unvollständige Zielmenge und ungültiger `$filter`, strikt, Collection | `ODR-E6003` | `C16` |
| Entity der Anfrage `partial` und Ziel ohne Beleg | `ODR-E6003` der Entity, keine Zielmenge gelesen | `C16` |
| verdorbener Eintrag in einer Zielmenge und unbekannte Property in einer früheren | `ODR-E9001` | `C16` |
| `STRICT_COVERAGE` falsch, Collection-Ziel unvollständig | bekannte Entities, Header nennt den Zustand | `C13` |
| `ETAG_ENABLED` an und aus | verknüpfte Entities mit und ohne `@odata.etag` | `C12` |
| Version 2.0 | `501` | `C1` |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) gelten für die Fakes von Repository, Schreiber und Interpreter; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Beim Schnitt (2026-11-28) dazu: `BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger` (2×, ohne Ausgang) für den Test aus §3. Der Slice ist nicht nebenläufig. Kein offener Eintrag verlangt vor diesem Slice einen eigenen Slice.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
