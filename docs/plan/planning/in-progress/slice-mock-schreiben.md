# Slice slice-mock-schreiben: MockUseCase, Teil Schreiben: POST, PATCH, PUT, DELETE

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-07`, `LH-FA-09`, `LH-QA-02`, `LH-QA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-07.a` · `LH-FA-07.b` · `LH-FA-09.a` · `SPEC-044` (PATCH und PUT auf partieller Entity) · `SPEC-108` Nr. 1 (`WithWorkspaceTransaction`) · `SPEC-023`, `SPEC-024` · `ARC-002`, `ARC-003` · `architecture.md` §2.4 (Zustand *Mock*).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-mock-schreiben`, Worktree `/Development/odata-recorder-wt-mock-schreiben`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `MockUseCase` führt POST (`201` mit Entity und `Location`, Duplicate Key `409`), PATCH (nur übergebene Properties), PUT (ersetzt, fehlende erforderliche Property `400`, kein Upsert) und DELETE (`204`) auf dem Workspace aus; jede Mutation wird vor dem Schreiben validiert, Schreibvorgänge eines Workspace laufen serialisiert in einer Transaktion, ein fehlgeschlagener Vorgang lässt den Zustand unverändert, und PATCH auf einer partiellen Entity ergibt `ODR-E6003`.

**Übernimmt (Sendung von `slice-odata-json-schreiben`, `SPEC-168` Nr. 8 und 10, `SPEC-169` Nr. 2):** die `Location` der `201`-Antwort setzt der Dienst aus `Adresse` des Schreibers im Satz `Codecs` (`CodecSatz.Schreiber`, dieselbe Zeichenfolge wie `@odata.id` der geschriebenen Entity; eine reine Funktion ohne `ctx`), mit der `Wurzel`, die `slice-mock-lesen` aus `PUBLIC_BASE_URL` und dem Service-Root bildet (nie aus dem `Host`-Header); `Codecs` und `Satz` selbst übernimmt `slice-mock-zuordnung`. Kein Ausschluss dieses Slice trifft diesen Punkt.

**Übernimmt (Sendung von `slice-odata-json-eingabe`, `SPEC-180` Nr. 6, `SPEC-109` Nr. 5):** der Dienst liest den Body von POST, PATCH und PUT über `CodecSatz.Eingabe` (`ODataEingabe.LeseEntity`) und ruft den Leser an drei Aufrufstellen {`POST`, `PATCH`, `PUT`} des Request (die drei Teile eines Changeset trägt `slice-batch-changesets`; zusammen die sechs Stellen der Menge). Je Aufrufstelle reicht er die acht Zusagen von `SPEC-180` Nr. 6 a bis h unverändert weiter (`ctx` des Aufrufs, `src`, `contentType` des Request, `entitySet`, `art`, `grenze` gleich `MAX_CAPTURE_BODY`, Fehlerabbildung über `SPEC-109` Nr. 5: `ErrEingabeUngueltig` → `ODR-E6007` mit `400`, `ErrEingabeNichtUnterstuetzt` → `ODR-E6002` mit `501`, Zustand unverändert bei jedem Fehler); Fänger je Aufrufstelle, der Fake des Ports hält jedes Argument fest. Der Handbuch-Eintrag von `ODR-E6007` kommt mit `slice-bootstrap-m3` (der Code erreicht einen Client erst mit dem Modus `mock`). Kein Ausschluss dieses Slice trifft diesen Punkt.

**Übernimmt (Sendung von `slice-odata-json-eingabe`, Review F-15, `SPEC-180` Nr. 3):** `Error()` eines Fehlers des Eingabe-Lesers trägt den Namen einer unbekannten Property im Ort, und ein Name kann bis zu `grenze` Bytes lang sein (gemessen: ein Name von 1 MiB ergibt einen Ort von 1 MiB). Der Architect entscheidet vor dem Code (Auftrag beim Start, `AGENTS.md` §3.12), ob der Ort im Fehlertyp begrenzt wird (Änderung von `SPEC-180` Nr. 3) oder ob die Länge als akzeptiertes Negativ mit Grund und Folge in *Ungeprüft* von `SPEC-180` steht; der Dienst schreibt den Ort nicht ungekürzt in Log oder Span, solange das nicht entschieden ist. Fänger: ein Name von `grenze` Bytes ergibt in Log und Span keinen Eintrag über der gewählten Grenze. Kein Ausschluss dieses Slice trifft diesen Punkt.

**Übernimmt (Erwartung von `slice-mock-zuordnung` für Punkt (a) und von `slice-mock-lesen` für die Punkte (b) bis (d), `SPEC-195` Nr. 7, `SPEC-196` Nr. 4 a und `SPEC-199`; Geber: `slice-mock-zuordnung` und `slice-mock-lesen`, die beiden Teile des Schnitts; Eintrag beim Start, Architect, 2026-11-07; Geber nachgezogen beim Schnitt, Planner, 2026-11-07):** (a) die Mutationen `POST`, `PATCH`, `PUT` und `DELETE` ersetzen die Nummer 7 von `SPEC-195` für diese vier Methoden (bis dahin endet jede mit `ODR-E6002`, bei einem Dienst, dessen Satz nur liest, mit `ODR-E6006`; jede andere Methode bleibt `ODR-E6002`); (b) ein Tombstone behält `KnownProperties` des Eintrags, denn die Maske des Lesens ist die Vereinigung über alle Einträge des Sets einschließlich der Tombstones (`SPEC-196` Nr. 4 a); (c) das `404` bei abgesichert fehlender Entity trägt den Code `ODR-E6008` (`SPEC-199`), das Nichtvorhandensein ist abgesichert bei einem Tombstone oder bei Coverage `complete` der Baseline (`SPEC-197` Nr. 3), sonst `ODR-E6003`; (d) die Mutation schreibt über `WithWorkspaceTransaction` und ändert die Coverage der Baseline nie (`SPEC-197` Nr. 6). Kein Ausschluss dieses Slice trifft diese Punkte.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- ETag, `If-Match` und `Prefer` — `welle-etag`; bis dahin gilt das Muss-Verhalten ohne `Prefer` (POST `201`, PATCH und PUT `204`).
- Navigations- und Referenzregeln über den eindeutig definierten Modellumfang hinaus — sie ergeben `ODR-E6002` (`LH-FA-07.a`).
- Snapshots und Restore — `slice-snapshot-use-case`; die Revision, die eine Mutation erhöht, ist hier der Gegenstand, der Snapshot nicht.
- Der Endpunkt der Admin-API für Mutationen — gibt es nicht; Mutationen laufen nur über den OData-Listener.

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

- [ ] POST und DELETE (der Body von POST, PATCH und PUT kommt über `ODataEingabe`, Sendung von `slice-odata-json-eingabe`, `SPEC-180` Nr. 6: an den drei Aufrufstellen {`POST`, `PATCH`, `PUT`} je ein Fänger für die acht Zusagen a bis h, Abbildung `ODR-E6007` mit `400` und `ODR-E6002` mit `501`, Zustand unverändert bei Fehler): Key aus dem Body (der Mock vergibt keinen), `201` mit `Location` aus `Adresse` des Schreibers (gleich dem `@odata.id` der geschriebenen Entity; Sendung von `slice-odata-json-schreiben`, `SPEC-168` Nr. 8), Duplicate Key `409`, `404` bei vollständiger Kenntnis, sonst `ODR-E6003`; `204` bei DELETE.
- [ ] PATCH und PUT: PATCH ändert nur übergebene Properties, PUT ersetzt (fehlende nullable Property wird `null`, fehlende erforderliche `400`), kein Upsert; PATCH auf partieller Entity `ODR-E6003`; unbekannte Property `400`.
- [ ] Atomarität: Schreibvorgänge je Workspace serialisiert in einer Workspace-Transaktion, ein fehlgeschlagener Vorgang ändert den Zustand nicht, die Revision steigt nur bei Erfolg; andere Workspaces bleiben unberührt (`LH-FA-09`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Modus `mock` erreicht es erst mit `slice-bootstrap-m3`).
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
| `src/hexagon/services/mock/` | update | Mutationen in der Workspace-Transaktion |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, Pflichtfelder, Fehlerklassen (Architect) |
| Tests | neu | Fälle von `LH-FA-07` je Methode und je Fehlerklasse; Wirkungs-Fänger für die Serialisierung der Schreibvorgänge |
| `src/hexagon/model/fehler.go`, `harness/katalog-ausnahmen.txt`, die Tests, die die Codetabelle listen, und die Reihen von `SPEC-230` Nr. 8 | update, nur bei neuem Code (Auftrag K) | Code und Status des Duplicate Key `409` und der Mutationsfehler, falls die Spezifikation keinen bestehenden nennt; `ODR-E6008` ist der 38. Code |

**Aufträge an den Architect zum Start** (Start, Planner, 2026-11-07, `main` = `1f39ca2`; vor dem ersten Code-Commit, `AGENTS.md` §3.12; der Architect arbeitet im Zweig `slice-mock-schreiben`; Ausgang jeweils *offen* in §6). Wo etwas vollständig ist, genügt die Bestätigung mit Nummer.

- **A. Tabelle mit allen Sätzen jeder Vertragsnummer (`AGENTS.md` §3.15).** Jeder Satz jeder Nummer, die der Vertrag führt (`LH-FA-07.a`, `LH-FA-07.b`, `LH-FA-09.a`, `SPEC-044`, `SPEC-195` Nr. 7, `SPEC-108` Nr. 1, die neuen Nummern), auch jeder Nebensatz („kein Upsert“, „der Mock vergibt keinen Key“, „ändert die Coverage nie“, „Zustand unverändert“), hat eine Zeile mit Nr., Zusage, Fänger, Mutation und Klasse (`nebenläufig` oder `andere`, `SPEC-128` Nr. 3) oder ein benanntes Negativ mit Grund und Folge; die Tabelle trägt die Spalten, die das Werkzeug aus `SPEC-128` liest.
- **B. Raster (§6, `AGENTS.md` §3.16).** Operation {POST, PATCH, PUT, DELETE} × Argumentart (fehlender, leerer, unbekannter oder fremder Wert; Abbruch; Wiederholung; Maskenzustand der Property: beobachtet, nicht beobachtet, Tombstone; Coverage `complete` und nicht) jede Zelle mit Ort oder „gilt nicht“ mit Grund; Zusatztabelle mit Fänger je Paar von Verstößen am selben Schlüssel (Body-Fehler und Duplicate Key, unbekannte Property und fehlende erforderliche Property, Key in URL und Key im Body widersprüchlich) und je Modus (Service nur lesend `ODR-E6006`, V2-Service, Methode ausserhalb der vier).
- **C. Pflicht-Mengen mit Achsen, Generator, Orakel und Prüfgrenze, gezählt gegen die Quelle** (Register `menge-im-plan-weicht-vom-orakel-des-tests-ab`, `unbeschraenkte-menge-ohne-benannte-pruefgrenze`): Methoden × Entity-Zustand (vorhanden, Tombstone, nie gesehen bei `complete` und bei `partial`) × Body-Art; die acht Zusagen a bis h der Eingabe × drei Aufrufstellen; Property-Arten und Edm-Typen (`model`-Mengen, nicht eine feste Zahl); die Fehlerabbildung (`SPEC-109` Nr. 5). Das Orakel trägt Statuscodes, Codes, Header und Meldungstexte als **Literale aus dem Vertrag**, nicht als Konstanten des Pakets.
- **D. Zweige und Rückgaben als Menge, in Runden unter dem Detektor** (Register `zweig-ohne-lauf-unter-dem-detektor`, 13×): Z-Liste mit Fänger je Zweig, Zählprüfung gegen die Quelle; für jeden Zweig, den eine Goroutine nur einmal erreicht (Abbruch des `ctx` in der Transaktion, Rollback, Wettlauf zweier Schreiber, Fehler des Repository mitten im Vorgang), die Runde (ein Lauf je Zweig unter `go test -race -coverprofile`); null Blöcke ohne Treffer oder Negativ mit Grund und Folge.
- **E. „Menge ohne ihre Positionen“** (neue Beobachtung `BEO-REPO/menge-ohne-ihre-positionen`): jede Menge, die der Vertrag nennt, führt auch ihre Positionen und Reihenfolgen (Collection mit mehreren Einträgen: Mutation am ersten, mittleren, letzten; Reihenfolge der Collection nach POST, PUT und DELETE; mehrere Properties, Verstoß in der ersten, mittleren, letzten); **Fixture gleicht nie dem Wert der Mutation** (Register `fixture-gleicht-dem-wert-der-mutation`, 5×). Allaussagen („jede Mutation“, „jeder Fehler lässt den Zustand unverändert“) als Schleife über beide Mengen (`AGENTS.md` §3.15, §3.17).
- **F. Gegenstand je Aufrufstelle, zählende Fakes, Bindung an den Port** (Register `hilfe-ersetzt-argument-verdeckt-eigenschaften`, `panik-eines-fakes-verdeckt-den-benannten-test` je 4×, `dienst-erfuellt-den-port-ohne-bindung` 2×, Lerneintrag V-2): was der Vertrag je Aufruf zusagt (`ctx` des Aufrufs, Fehlerabbildung, Kontextfehler, die acht Argumente der Eingabe) hat einen Fall je Aufrufstelle der Kette (Workspace-Repository, Entity-State, Transaktion, `ODataEingabe`, Schreiber), nicht am ersten Port; jeder Fake hält jedes Argument fest und zählt jede Operation, auch die, die der Dienst nie rufen darf (Schreiben bei Fehler, Lesen der Upstream-Ports, `Persist` nach Rollback); eine Zusicherung zur Übersetzungszeit bindet Fake und Dienst an den Port. Der Architect nennt, wo die Bindung steht.
- **G. Serialisierung und Fenster der Sperre** (`AGENTS.md` §3.10; Register `rate-ohne-detektor-schwankt-mit-der-last` 8×, `sperrfenster-zwischen-sammlungen-ohne-faenger`, `nebenlaeufiger-lauf-nur-im-glueck-gruen` 13×): je Fenster der Sperre (zwischen Prüfen und Schreiben, zwischen Validieren und Revision, zwischen zwei Workspaces) eine Zeile mit Wirkungs-Fänger oder benanntem Negativ; vor einem Negativ ein Haltepunkt an einem Argument, das der Test übergibt (`ctx`, Fake, Transaktionsfunktion); Raten mit Zahl der Prozesse und Messendem. Wirkungs-Fänger zählen, was das Fenster zuließe, statt auf ein Datenrennen zu warten.
- **H. Entscheidung vor dem Code (§6 des Plans): Fehlercodes des Duplicate Key `409` und der Mutationsfehler, `entity_version`, Referenzregeln und Navigationen (`LH-FA-07.a`), Maske bei PUT/PATCH.** Jede im Vertrag mit Ort; der Implementer entscheidet keine.
- **I. F-15 aus `slice-odata-json-eingabe` (Länge des Ortes im `Error()`).** Der Architect entscheidet **vor dem Code**: Begrenzung in `SPEC-180` Nr. 3 (Änderung der Eingabe-Nummer, Mutation und Fänger im Paket `jsoncodec` ist dann ein Teil dieses Slice oder geht mit Grund an den Planner) oder akzeptiertes Negativ mit Grund und Folge in *Ungeprüft* von `SPEC-180`. Bis zur Entscheidung schreibt der Dienst den Ort nie ungekürzt in Log oder Span; die Menge der Ausgabewege (Log, Span, Fehlerobjekt, Audit) trägt je Element einen Fänger mit einem Namen von `grenze` Bytes.
- **J. Eingabe-Satz, V4 und V2.** Der Dienst ruft `CodecSatz.Eingabe` an den drei Aufrufstellen; ist `Eingabe` im Satz eines V2-Service nicht belegt (heute führt kein Plan in `open/` oder `done/` einen V2-Leser für Client-Bodies, `slice-v2-bootstrap` nennt keinen), entscheidet der Architect die Randform (nil-Eingabe, Fehlercode `ODR-E6006` oder `ODR-E6002`, kein Panik) und trägt sie in Raster und Tabelle; entsteht daraus eine Sendung an einen anderen Slice, geht sie als Rückgabe an den Planner, nicht als Eintrag in fremde Pläne. Body-Fehler `ODR-E6007` (`400`) und `ODR-E6002` (`501`) wie in §1.
- **K. Kennungen und Katalog-Pflicht (`SPEC-201` Nr. 3 und 11).** Der Architect nimmt nur die freien Nummern `SPEC-192` bis `SPEC-194`; `SPEC-195` bis `SPEC-199` sind vergeben (`slice-mock-lesen`), `SPEC-230` bis `SPEC-232` und `SPEC-234` bis `SPEC-237` Linie D, ab `SPEC-238` Linie D/Admin. Der Bereich 160 bis 199 der Linie B ist fast erschöpft, die Kennungsfrage ist beim User offen: braucht der Vertrag mehr als drei Nummern, bevorzugt er neue Sätze in bestehenden Nummern des Mock (`SPEC-195` Nr. 7 wird ersetzt, nicht dupliziert), und eine vierte Nummer geht als Rückgabe an den Planner. Ein neuer Code: `src/hexagon/model/fehler.go` samt Test, §4 der Spezifikation, Reihen von `SPEC-230` Nr. 8 (Metrik `odr_errors_total`, heute 38) samt Tests der Telemetrie, `harness/katalog-ausnahmen.txt` (das Binary erreicht den Modus `mock` erst mit `slice-bootstrap-m3`, also Ausnahme mit Grund) und `make katalog-check` im Bericht; ein Code, der im Katalog des Binary landet, kommt mit Handbuch-Kapitel im selben Slice. Nächste freie Nummer ist `ODR-E6009`; `slice-admin-listener-token` (Auftrag J dort) kann denselben Code beanspruchen: der Architect liest den Zweig und die Spezifikation von `main`, nennt die Nummer im Bericht, und bei Zusammenstoß geht die Frage an den Planner.
- **L. Schnittfrage und Marke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1). Der Architect meldet *a*, *b*, *f* und *p* aus der Tabelle und setzt im Block `groessenmarke` weder `basis` noch `schaetzung` noch `marke`; den Block setzt der Planner nach der Bestätigung in eigenem Commit vor dem ersten Code-Commit. Wird *S* ≥ 2223, wird **vor dem Code** geschnitten (oder `schnitt-grund` in §4, wenn kein Teil ohne geteilten Mechanismus abtrennbar ist); Kandidat: POST und DELETE gegen PATCH und PUT (die Mutationen teilen Transaktion und Zustand, der Schnitt muss einen grünen Stand zwischen den Teilen lassen, Register `schnitt-laesst-keinen-gruenen-stand-zwischen-den-slices`). Die Auslastung der letzten Teile lag bei 0,47 bis 0,60 (Register `schaetzformel-ueberschaetzt-den-slice`, 22×): Produktcode wird überschätzt, Tests unterschätzt; der Architect benennt das in der Meldung, ohne die Formel zu ändern.
- **M. Pfade neben den Zweigen.** `slice-telemetry-spans` schreibt `src/adapters/driven/telemetry/` und `spec/spezifikation.md`; `slice-admin-listener-token` schreibt `src/adapters/driving/admin/` (neu), `spec/spezifikation.md` und bei einem neuen Code `model/fehler.go`, `harness/katalog-ausnahmen.txt`. Dieser Slice schreibt `src/hexagon/services/mock/`, `spec/spezifikation.md` und bei einem neuen Code die genannten Dateien samt der Tests der Telemetrie zu `SPEC-230` Nr. 8 (`regeln_test.go`); die Überschneidung ist `spec/spezifikation.md` (nur eigene Nummern und Zeilen) und der eine neue Code. Der Architect bestätigt die Liste und nennt jede Datei darüber hinaus (Abweichung an den Planner).
- **N. Reports der Rollen.** Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (Register `report-der-rolle-faellt-docs-check`); der Architect nimmt den Satz in seinen Bericht auf, der Planner reicht ihn an Reviewer und Verifier weiter. Der Implementer schreibt Repo-Dateien nur mit `Edit` und `Write` (Register `betriebsregel-fuer-repo-dateien-verletzt`, 9×).

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Nebenläufig: die Zeilen der Tabelle zählen mit Faktor 85 (`AGENTS.md` §3.14); „unter einer Sperre, ohne sie dazwischen freizugeben“ bekommt je Fenster der Sperre eine Zeile mit Fänger oder benanntem Negativ (`AGENTS.md` §3.10).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-mock-zuordnung`, `slice-mock-lesen` (beide Teile des geschnittenen Plans) und `slice-odata-json-eingabe` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).
**Stand beim Start** (Planner, 2026-11-07, `main` = `1f39ca2`): erfüllt. `slice-mock-zuordnung`, `slice-mock-lesen` und `slice-odata-json-eingabe` liegen in `done/`; `slice-odata-json-schreiben` (Schreiber, `Adresse`) ebenfalls. §1 *Ausdrücklich NICHT* und DoD der Nachbarn gelesen (`AGENTS.md` §3.13): `slice-mock-zuordnung` und `slice-mock-lesen` nennen die Sendungen R4 (a) bis (d) als Erwartung an diesen Slice; `slice-etag-versionen` (Zähler `entity_version`), `slice-snapshot-use-case` (Serialisierung je Workspace), `slice-hybrid-service` (Mutationen mit `ODR-E6002`) und `slice-batch-changesets` (drei Aufrufstellen der Eingabe im Changeset) bauen auf diesem Slice auf und führen die Punkte selbst; kein Ausschluss eines Nehmers trifft eine Sendung. In `in-progress/` laufen daneben `slice-telemetry-spans` (Worktree `/Development/odata-recorder-wt-telemetry-spans`) und `slice-admin-listener-token` (Worktree `/Development/odata-recorder-wt-admin-listener-token`); Pfade siehe Auftrag M. Die Gate-Kette bleibt unberührt (Nr. 7 der Übergangsregeln). Speicher beim Start: 16 GB verfügbar.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt** (Start am 2026-11-07, Prosa, kein Block auf `main`); eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt (Auftrag L in §3): danach setzt der Planner in eigenem Commit vor dem ersten Code-Commit den Block `groessenmarke` mit `basis`, `schaetzung` und `marke`; bis dahin bindet kein Gate. Liegt die Schätzung × 1,25 über 3000, wird vor dem ersten Code-Commit geschnitten (zum Beispiel an POST und DELETE gegen PATCH und PUT).
- `in-progress` → `open` (blockiert — Carveout?): Referenzregeln des Modells und die Frage, wer `entity_version` erhöht, sind nicht entschieden (siehe §6).

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

- Länge des Ortes im `Error()` des Eingabe-Lesers (Sendung von `slice-odata-json-eingabe`, Review F-15, siehe §1) — **Ausgang:** offen; Adresse: Architect (Begrenzung in `SPEC-180` Nr. 3 oder Negativ mit Grund und Folge) vor dem ersten Code-Commit.
- Serialisierung der Schreibvorgänge je Workspace ist eine Zusage über Nebenläufigkeit: der Detektor meldet nur, was läuft — **Ausgang:** offen; Wirkungs-Fänger, der zählt, was das Fenster zuließe, mit der Rate je Mutation und je Lauf von `go test` samt Zahl der Prozesse und Messendem (`AGENTS.md` §3.10; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`).
- Umfang der Referenzregeln und Navigationen, „soweit der Modellumfang sie eindeutig definiert“ (`LH-FA-07.a`) — **Ausgang:** offen; Architect nennt die Fälle, alles andere ist `ODR-E6002`.
- Wer `entity_version` erhöht, solange kein ETag ausgeliefert wird (die Spalte existiert im Schema) — **Ausgang:** offen; Architect entscheidet es in diesem Slice, so dass `welle-etag` darauf aufbaut.
- PUT und PATCH mit nicht beobachteten Properties (Property-Maske, `SPEC-044`) — **Ausgang:** offen; Raster, eine Zelle je Methode und Maskenzustand.
- Aufträge A bis N an den Architect zum Start (§3): Sätze je Vertragsnummer, Raster, Pflicht-Mengen mit Zählprüfung, Zweige und Runden unter dem Detektor, Menge mit Positionen und Fixture, Gegenstand je Aufrufstelle und Bindung an den Port, Fenster der Sperre, Entscheidungen vor dem Code, F-15, Eingabe-Satz für V2, Kennungen und Katalog, Schnitt und Marke, Pfade, Reports — **Ausgang:** offen; Adresse: Architect (Bestätigung in diesem Abschnitt oder Ergänzung), Implementer (Test), Reviewer und Verifier (Auftrag N).
- Kennungen: nur `SPEC-192` bis `SPEC-194` sind frei, die Kennungsfrage der Linie B (Bereich 160 bis 199) liegt beim User — **Ausgang:** offen; Adresse: Planner (Rückgabe des Architect bei mehr als drei Nummern).
- Code-Nummer eines neuen Fehlers (`ODR-E6009`) gegen `slice-admin-listener-token` (dort Auftrag J) — **Ausgang:** offen; Adresse: Architect (Bericht), Integrationsschritt (der zweite Merge zieht `main` nach).
- Pfade neben den laufenden Zweigen (`spec/spezifikation.md`, bei neuem Code `model/fehler.go`, `harness/katalog-ausnahmen.txt`, Tests der Telemetrie) — **Ausgang:** offen; Adresse: Integrationsschritt.
- Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch` (Gate-Kette): der Slice ändert weder `GATE_CHECKS` noch ein Gate; `harness/katalog-ausnahmen.txt` ist Eingabe von `make katalog-check`, kein Gate-Fragment — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice ist nebenläufig: `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) sind offen und gelten für die Rate des Wirkungs-Fängers und für Zweige, die kein Lauf ausführt; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes. Verkörpert: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (7×, `AGENTS.md` §3.10, Sensor `make test-race`). Keiner der offenen Einträge erreicht mit diesem Slice 3×.
Beim Start nachgezogen (Planner, 2026-11-07, `main` = `1f39ca2`, Zähler aus den Dateien unter `evidence/`): `zweig-ohne-lauf-unter-dem-detektor` 13× (Schwelle lange erreicht; Verkörperung beim Lese-Schritt der Welle-Closure, Auftrag D), `rate-ohne-detektor-schwankt-mit-der-last` 8× und `nebenlaeufiger-lauf-nur-im-glueck-gruen` 13× (Auftrag G), `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `dienst-erfuellt-den-port-ohne-bindung` 2× (Auftrag F), `menge-ohne-ihre-positionen` 1× und `fixture-gleicht-dem-wert-der-mutation` 5× (Auftrag E), `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3× und `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11× (Auftrag C), `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 44× (Auftrag A), `randform-erst-im-review-entschieden` 34× (Auftrag H), `fix-deckt-die-genannte-mutation-nicht-die-klasse` 34×, `akzeptiertes-negativ-ohne-folge` 15×, `schaetzformel-ueberschaetzt-den-slice` 22× (Auftrag L), `fehlertext-nennt-ort-und-wert-ohne-faenger` 1× (Auftrag I), `betriebsregel-fuer-repo-dateien-verletzt` 9× (Auftrag N).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
