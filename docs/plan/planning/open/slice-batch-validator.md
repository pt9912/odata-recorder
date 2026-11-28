# Slice slice-batch-validator: Prüfung der Struktur eines `$batch` (Methoden, `Content-ID`, Verweise)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-16.a` (eigener Validator-Contract-Test, unbekannte Operationen `501`) · `SPEC-184` (Vertrag dieses Slice) · `SPEC-182` Nr. 4 und 5 (Fehlerwert, Ursachen, Rangfolge hinter dem Leser) · `SPEC-183` Nr. 4 c (Form der `Content-ID`) und Nr. 8 (Beispiel, Zeilen der Prüfung) · `SPEC-022` (`ODR-E6001`) · `SPEC-180` (`ODR-E6007`) · `ARC-004`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-28.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die reinen Funktionen `model.PruefeBatch` und `model.BatchRequest.Verweis` im Paket `src/hexagon/model` prüfen die Struktur eines gelesenen Batch nach `SPEC-184`: erlaubte Methoden je Ebene, `Content-ID` je Changeset eindeutig, Verweise `$<id>` nur auf einen früheren Request desselben Changeset; ein Verstoß ist ein `*model.BatchFehler` mit Ursache und Ort (`ODR-E6007` bzw. `501` mit `ODR-E6001`). Beide Funktionen lesen nur die Domain-Werte und rufen keinen Port (`LH-FA-16.a`: Validator-Contract-Test).

**Übernimmt (Schnitt von `slice-batch-parser`, Geber `slice-batch-parser`, `AGENTS.md` §3.13):** die Prüfung der Struktur samt `Verweis`, die Zeilen 37 bis 43 der Tabelle *Zusage · Fänger*, die zwei Zeilen des Rasters und die drei Paare der Zusatztabelle zur Prüfung und den Validator-Contract-Test, die der frühere Plan von `slice-batch-parser` führte.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Leser, Port `ODataBatch`, Domain-Werte, Fehlerwert `BatchFehler` mit seinen Ursachen und die Zeilen der Abbildung — `slice-batch-parser`; dieser Slice liest sie und ändert sie nicht, auch die Ursachen der Prüfung stehen dort schon in der Liste.
- Der Aufruf von `PruefeBatch` vor der ersten Ausführung, die Auflösung eines Verweises gegen den früheren Request und die Ausführung — `slice-batch-changesets` (dessen §1 trägt die Sendung); dieser Slice liefert die Funktionen, nicht ihren Aufrufer.
- Parsen von Ziel und Body eines Einzelrequests und Verweise im Body — `slice-batch-changesets`; `SPEC-184` Nr. 4 regelt beides nicht.
- Die Form der Werte eines von Hand gebauten Batch — nicht Gegenstand der Prüfung (`SPEC-184` *Ungeprüft*); der Leser ist der einzige Erzeuger.

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

- [ ] Prüfung (`SPEC-184`): Methoden je Ebene (außerhalb und innerhalb eines Changeset), `Content-ID` je Changeset eindeutig, `Verweis`, Verweise nur rückwärts; Verstoß mit Ursache und Ort, nicht Unterstütztes `501` mit `ODR-E6001`; die Funktionen ändern den Wert nicht.
- [ ] Randformen: je Zelle des Rasters (§6) ein Ort in der Spezifikation und ein Fänger der Tabelle (§3); Validator-Contract-Test (`LH-FA-16.a`); `make a-check` hält die Richtung (`model` importiert keinen Port und keinen Adapter).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (die Funktionen haben keinen Aufrufer vor `slice-batch-changesets`).
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
| `src/hexagon/model/batch_pruefen.go` | neu | `PruefeBatch` und `Verweis` (`SPEC-184`) |
| `src/hexagon/model/batch_pruefen_test.go` | neu | Zeilen 37 bis 43, Validator-Contract-Test (`LH-FA-16.a`) |

Die Randformen sind in `SPEC-184` entschieden (`AGENTS.md` §3.12); der Implementer entscheidet keine.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Zeilen und Testnamen des Architect, aus dem Plan von `slice-batch-parser` mit ihren Nummern übernommen; Dateien unter `src/hexagon/model/`; der Implementer fährt je Zeile die Mutation selbst):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 37 | Methoden: je Ebene die erlaubten; `GET` im Changeset `methode`, jede andere `methode_nicht_unterstuetzt` | `TestPruefeBatchMethode`: Schleife über {die fünf, `HEAD`, `OPTIONS`, `MERGE`, `PROPFIND`, `get`, leer} × beide Ebenen × Stelle | `GET` im Changeset erlaubt | andere |
| 38 | `Content-ID` je Changeset eindeutig, bytegleich; verschiedene Changesets, Einzelteile und fehlende stoßen nicht an | `TestPruefeBatchContentID`: alle Paare von Stellen in Changesets der Länge 2 bis 5, Schreibweise, über Changesets | Eindeutigkeit über den ganzen Batch; nur Nachbarn verglichen | andere |
| 39 | `Verweis` liefert `<id>` genau nach `SPEC-184` Nr. 3 | `TestBatchVerweis`: Ziele über {`$1`, `$1/x`, `$1?a`, `$`, `/$1`, `%241`, `$1(2)`, die fünf Wörter, Form-Grenzen 128 und 129} | `$metadata` ist ein Verweis | andere |
| 40 | Verweis nur auf einen früheren Request desselben Changeset, sonst `referenz` | `TestPruefeBatchReferenz`: früher, selbst, später, anderes Changeset, Einzelteil davor und danach, außerhalb eines Changeset | Spätere zählen mit | andere |
| 41 | Rangfolge: erster Teil, erster Request, dann Methode, Eindeutigkeit, Auflösbarkeit; jeder Verstoß des Lesers geht vor | `TestPruefeBatchRang`: alle Paare der drei Prüfungen am selben und an zwei Requests; ein Fall mit Leser-Verstoß hinter einem Prüfverstoß | Eindeutigkeit vor Methode | andere |
| 42 | `PruefeBatch` ändert `b` nicht und ist wiederholbar; ein leerer Batch und ein leeres Changeset sind `nil` | `TestPruefeBatchRein`: tiefer Vergleich vor und nach, zweimal; leere Formen | `b` wird sortiert | andere |
| 43 | Das unveränderte Beispiel von `SPEC-183` Nr. 8 ergibt `nil`, und die Zeilen der Prüfung in seiner Tabelle ergeben das Genannte (Validator-Contract-Test) | `TestPruefeBatchContract` | Eine Antwort wird vertauscht | andere |

```groessenmarke
pfad: src/hexagon/model/batch_pruefen.go
pfad: src/hexagon/model/batch_pruefen_test.go
produktcode: 100
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

**Start** (`next` → `in-progress`): `slice-batch-parser` liegt in `done/`.

- **Code-Abhängigkeit, mit Grund:** `slice-batch-parser` liefert die Domain-Werte (`model.Batch`) und den Fehlerwert (`model.BatchFehler` mit Ursachen und Sentinels), die beide Funktionen lesen und liefern. Wie für ihn ist der Welle-Trigger (`welle-mock` in `done/`) Reihenfolge-Vorsicht, keine Code-Abhängigkeit: ein Start davor ist eine Entscheidung des Users, nicht dieses Plans.
- **Was der Slice nicht braucht — und warum:** `MockUseCase`, Workspace, Parser der OData-URL und der Client-Bodies, HTTP-Adapter, Composition Root (die Funktionen sehen Werte, keine Bytes und keinen Zustand).
- **Größenmarke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1; die Faktoren stehen allein dort; Planner, 2026-11-28, mit dem Schnitt): gezählt ist die Tabelle in §3 mit 7 Zeilen, *a* = 0, *b* = 7, *f* = 0, *p* = 100 (Setzung des Planner aus den 600 Zeilen, die der Architect für Leser und Prüfung zusammen erwartete). *S* = 660, Marke *M* = **900**; keine Schnittfrage. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Block trägt `basis`, `schaetzung` und `marke` noch nicht (`make groessenmarke` meldet *Marke steht aus*); der Planner setzt sie beim Start, vor dem ersten Code-Commit.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 900 Zeilen (Metrik wie oben). Eine Überschreitung schneidet nicht (`SPEC-128` Nr. 10 c); sie bekommt einen Eintrag `ueberschreitung` mit Grund im Block.
- `in-progress` → `open` (blockiert — Carveout?): eine Randform fehlt in `SPEC-184`, oder die Prüfung braucht einen anderen Domain-Wert oder Fehlerwert, als `slice-batch-parser` geliefert hat (dann an den Architect; `slice-batch-parser` liegt in `done/` und ist keine Adresse, `AGENTS.md` §3.13).

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

- Die Regel der Form einer `Content-ID` (`SPEC-183` Nr. 4 c) braucht der Leser aus `slice-batch-parser` und `Verweis` hier; ob sie an einem Ort im Domain-Modell steht oder zweimal, sagt der Vertrag nicht, und Zeile 39 führt nur die Längengrenzen, nicht das Alphabet (`AGENTS.md` §3.17) — **Ausgang:** offen; Adresse Architect, vor dem ersten Code-Commit von `slice-batch-parser` (dort derselbe Punkt in §6).
- Zeile 41 verlangt „ein Fall mit Leser-Verstoß hinter einem Prüfverstoß“: der Fall ruft den Leser, und ein Test in `model` darf keinen Adapter importieren (`make a-check`); die Datei dieses Falls liegt dann außerhalb der Pfade des Blocks — **Ausgang:** offen; Adresse Architect, vor dem ersten Code-Commit (Ort des Fängers und Pfad im Block).
- Kleine Schätzung (*S* = 660): ein kleiner Slice mit mehreren Prüfrunden liegt eher über der Marke (`SPEC-128` Nr. 10 a) — **Ausgang:** offen; Eintrag `ueberschreitung` mit Grund.
- Zusammenspiel mit der Ausführung: Aufruf vor der ersten Ausführung und Auflösung des Verweises — **Ausgang:** offen; `slice-batch-changesets` §1 führt die Sendung.
- Randformen (`AGENTS.md` §3.12, §3.16) — **Ausgang:** offen; Raster unten, eine Zelle ohne Ort geht an den Architect.

**Raster *Operation × Argumentart*** (Zellen des Architect, aus dem Plan von `slice-batch-parser`; die Spalten sind dieselben wie dort):

| Operation | Body fehlt oder leer | Grenze fehlt, leer oder ungültig | unbekannter Teiltyp oder Methode | fremde Form (JSON, absolute URL, anderer Host) | über Grenze (Größe, Teile, Tiefe) | Wiederholung (`Content-ID`, Changeset) | Abbruch (Kontext, Strom) |
|---|---|---|---|---|---|---|---|
| Validator: Methoden und Changesets | `SPEC-184` Kopf, `SPEC-183` Nr. 5 (leer ist ohne Verstoß) | gilt nicht: er sieht Werte, keine Bytes | `SPEC-184` Nr. 1 | gilt nicht: er prüft kein Ziel außer dem Verweis (`SPEC-184` Nr. 3) | gilt nicht: die Grenzen prüft der Leser (`SPEC-183` Nr. 7) | `SPEC-184` Kopf (reine Funktion) | gilt nicht: reine Funktion ohne Kontext und Strom (`SPEC-184` Kopf) |
| Validator: `Content-ID` und Referenzen | `SPEC-182` Nr. 3, `SPEC-183` Nr. 4 c, `SPEC-184` Nr. 2 | gilt nicht: er sieht Werte, keine Bytes | `SPEC-184` Nr. 3, 4 (unbekannter Verweis, reservierte Wörter) | `SPEC-184` Nr. 3 (`/$1`, `%241`) | `SPEC-183` Nr. 4 c (Länge der `Content-ID`) | `SPEC-184` Nr. 2 | gilt nicht: reine Funktion ohne Kontext und Strom (`SPEC-184` Kopf) |

**Zusatztabelle** (je Paar von Verstößen; `AGENTS.md` §3.16). Eine Zeile je Modus gilt nicht: die Prüfung kennt keinen Modus (`SPEC-184`).

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| unbekannte Methode · Verstoß des Lesers dahinter | der Leser (`SPEC-182` Nr. 5 h) | Zeile 41 |
| doppelte `Content-ID` · unzulässige Methode am selben Request | Methode (`SPEC-184` Nr. 5) | Zeile 41 |
| doppelte `Content-ID` · nicht auflösbarer Verweis am selben Request | `content_id_doppelt` (`SPEC-184` Nr. 5) | Zeile 41 |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-28); die Treffer sind verkörpert und gelten als Regel: `BEO-REPO/menge-ohne-ihre-positionen` (`AGENTS.md` §3.15) für die Mengen der Zeilen 37 bis 40 (Ebene, Stelle, Abstand des Paares), `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` für die Betriebsregel „nur `Edit` und `Write`“. Der Code ist nicht nebenläufig und hat keine Fakes. Kein offener Eintrag erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
