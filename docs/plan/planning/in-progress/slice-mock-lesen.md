# Slice slice-mock-lesen: MockUseCase, Teil Lesen mit Coverage-Prüfung

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-06`, `LH-FA-08`, `LH-FA-11`, `LH-FA-09`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-06.a` Nr. 4 und 5 · `LH-FA-11.a` · `LH-FA-07.b` (Lesefälle) · `SPEC-044` · `SPEC-011` · `SPEC-024`, `SPEC-070` · `LH-FA-04.c` (Adressierung) · `ARC-002`, `ARC-003` · `architecture.md` §4.4, §2.4 (Zustand *Mock*).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-mock-lesen)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `MockUseCase` (Inbound Port, Paket `src/hexagon/services/mock`) ordnet einen Request Service und Workspace zu, liefert `$metadata` aus den gespeicherten Metadaten und beantwortet Collection, Key-Lookup und `/$count` aus dem Workspace-Zustand über Parser, Coverage-Prüfung, Interpreter und Codec; die Coverage wird vor der Query-Ausführung geprüft (`ODR-E6003` bei `STRICT_COVERAGE=true`, sonst Antwort mit Header `X-Recorder-Coverage`), und ein `404` entsteht nur, wo das Nichtvorhandensein abgesichert ist.

**Übernimmt (Sendung von `slice-v2-bootstrap`, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted`):** die Konfiguration von `MockUseCase` nimmt je Dienst `Codecs *outbound.CodecSatz` (URL-Leser und Schreiber) entgegen; `nil` bedeutet die V4-Implementierungen der Konfig, und der Dienst verzweigt nie nach der Version (`SPEC-175` Nr. 3; ein Feld des Satzes, das der Dienst nicht braucht, darf `nil` sein). Kein Ausschluss dieses Slice trifft sie.

**Übernimmt (Sendung von `slice-odata-json-schreiben`, `SPEC-168` Nr. 7 und 10, `SPEC-169`):** der Dienst ruft über den Satz `URL` und `Schreiber`, setzt `Content-Type` und `Kennzeichen` unverändert aus der `Darstellung` jeder Ausgabe (auch der Anzahl und des Fehlerobjekts) in die Antwort und kennt die Namen der Kennzeichen nicht (`SPEC-168` Nr. 7; `SPEC-175` Nr. 7); der Dienst reicht `Accept` des Requests an den Schreiber und handelt Stufe und Zahlform nicht selbst aus (`SPEC-169` Nr. 3). **`Wurzel`** bildet der Dienst aus der Konfiguration (`PUBLIC_BASE_URL`, `SPEC-055`) und dem Service-Root, **nie** aus dem `Host`-Header der Anfrage (`SPEC-168` Nr. 10); der Schreiber prüft nur die Form (`SPEC-169` Nr. 2). Die Quelle von `Wurzel` ist damit diesem Slice zugewiesen (Geber: `slice-odata-json-schreiben` §6); `slice-bootstrap-m3` reicht `PUBLIC_BASE_URL` aus dem Start an die Konfiguration des Dienstes. Kein Ausschluss dieses Slice trifft diese Punkte.

**Übernimmt (Sendung von `slice-query-interpreter`, `SPEC-190`, `SPEC-191`; Geber: `slice-query-interpreter` §6):** der Dienst ruft `model.FuehreAus` und erfüllt die fünf Pflichten des Aufrufers: (1) die Maske `beobachtet` aus den bekannten Properties der Baseline des Workspace (`SPEC-190` Nr. 1 d); (2) `/$count` nur mit `Filter` im Aufruf, damit `@odata.count` und `/$count` dieselbe Zahl tragen (Nr. 2); (3) jede Entity trägt die Key-Properties in `Properties` (Nr. 3 f); (4) `Schluessel` ist je Entity eindeutig und nichtleer (Nr. 3 e); (5) kein Wert der Art `redigiert` in einer Property, die `$filter` oder `$orderby` nennt, sonst endet der Aufruf mit `500` und `ODR-E9001` (Nr. 3 g). Kein Ausschluss dieses Slice trifft diese Punkte.

**Übernimmt (Sendung von `slice-odata-url-pfad`, §6 und §7 dort, `SPEC-160`; Geber: `slice-odata-url-pfad`; Eintrag beim Start, Planner, 2026-11-07):** (a) der Pfad kommt relativ zum Service-Root beim Parser an, das Abschneiden des Präfixes ist Sache des Dienstes (`SPEC-160` Nr. 1; `/` für den Root, auch ohne abschließenden `/`); (b) welche Option zu welcher Art der Ressource gehört (`/$metadata?$top=1`, `/Set(1)?$orderby=Id`; `SPEC-160` Nr. 9 d; `/$count` ignoriert `$top`, `$skip`, `$orderby`); (c) ein eigener Fake von `URLCodec` in den Tests des Dienstes (a-check verbietet den Import des Adapters); (d) der Ende-zu-Ende-Fall „der Schlüssel der URL findet die importierte Entity, je Typ der acht“ (die Gleichheit der kanonischen Texte von URL-Parser und JSON-Leser ist *Ungeprüft* in `SPEC-160`; `BEO-REPO/regel-in-zwei-adaptern-ohne-sensor`); (e) für `/` liefert der Parser die Ursache `service_dokument` (`501`), die der Dienst abbildet, solange keine Mindestfassung entschieden ist (deckt sich mit dem Ausschluss des Service-Dokuments unten). Dazu der Verdrahtungs-Fall: der Parser ist bis hierher unverdrahtet, ein Defekt im Zusammenspiel von Parser und Interpreter zeigt sich erst im Dienst (Risiko-Ausgang *weiter offen* bei `slice-mock-lesen`, §6 dort). Kein Ausschluss dieses Slice trifft diese Punkte.

**Übernimmt (Erwartung von `slice-hybrid-service`, dessen §4 *Code-Abhängigkeit, Port*; Geber: `slice-hybrid-service`, Eintrag beim Start, Planner, 2026-11-07):** `MockUseCase` liegt als Inbound Port in `src/hexagon/ports/inbound/` und trägt den Lesepfad so, dass `slice-mock-schreiben` ihn erweitert und `HybridUseCase` ihn ruft; die Fehler `ODR-E6001` (nicht unterstützte Abfrage) und `ODR-E6003` (Coverage) kommen als Fehler des Ports an, nach denen der Aufrufer entscheidet. Kein Ausschluss dieses Slice trifft den Punkt.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Mutationen — `slice-mock-schreiben`; Konstruktor, Routing und Zuordnung sind gemeinsam, die Zustandsänderung nicht.
- Fallback auf Replay oder Upstream — `welle-hybrid`; `MockUseCase` kennt weder `ReplayUseCase` noch `UpstreamClient`.
- Spans und Metriken (`SPEC-029`) — `welle-betrieb`; es gibt noch keinen Port `Telemetry`.
- Das Service-Dokument (`GET` auf den Service-Root) — nicht im Muss-Umfang von `welle-mock` (Entscheidung des Users, 2026-10-30); kein Vertrag in der Spezifikation, es bleibt `501`. **Auftrag an den Architect vor dem ersten Code-Commit:** Er prüft am Referenz-Client des M3-Tests (`slice-bootstrap-m3`), ob dieser das Service-Dokument anfordert, und hält den Beleg (den Request oder sein Fehlen) in §6 fest. Fordert der Client es an, schreibt der Architect eine Mindestfassung mit eigenem Vertrag in die Spezifikation, vor dem ersten Code-Commit (`AGENTS.md` §3.12), und gibt den Slice an den Planner zurück (DoD, Marke, Handbuch-Kapitel); sonst bleibt es bei `501`.

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

- [ ] Zuordnung und `$metadata`: Service und Workspace aus dem Request (`ODR-E5005`, `ODR-E5003`, `ODR-E7001`), `$metadata` aus den gespeicherten Metadaten, ein EntitySet ohne Instanzen bleibt im Modell (`LH-FA-06`, Boundary).
- [ ] Collection, Key-Lookup und `/$count` über Parser, Interpreter und Codec; `404` nur bei abgesichertem Nichtvorhandensein, sonst Coverage-Fehler; `Content-Type` und `Kennzeichen` aus der `Darstellung` des Schreibers, `Wurzel` aus `PUBLIC_BASE_URL` und Service-Root, nie aus dem `Host`-Header (Fänger: ein anderer `Host` ändert Kontext und `@odata.id` nicht; Sendung von `slice-odata-json-schreiben`, `SPEC-168` Nr. 7 und 10); Aufruf des Interpreters mit den fünf Pflichten des Aufrufers aus `SPEC-190` Nr. 1 d, 2, 3 e, 3 f und 3 g (Maske aus den Baseline-Properties, `/$count` nur mit `Filter`, Key-Properties je Entity, `Schluessel` eindeutig und nichtleer, kein Wert `redigiert` in gefilterten oder geordneten Properties; Sendung von `slice-query-interpreter`); die Sendung von `slice-odata-url-pfad` (§1: Präfix, Optionen je Art der Ressource, Fake von `URLCodec`, Ende-zu-Ende-Fall des Schlüssels je Typ der acht, `service_dokument` als `501`) und der Port für `slice-hybrid-service`.
- [ ] Coverage vor der Query-Ausführung, gebunden an die Baseline des Workspace: `ODR-E6003` bei `STRICT_COVERAGE=true`, sonst Header `X-Recorder-Coverage`; null beobachtete Entities sind kein Leerheitsbeweis (`LH-FA-11`).
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
| `src/hexagon/ports/inbound/` | update | `MockUseCase` |
| `src/hexagon/services/mock/` | neu | Zuordnung, Coverage-Prüfung, Lesen, Ausführung des Interpreters, Serialisierung |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, Zuordnung im Modus `mock` (Architect) |
| Tests | neu | Fälle von `LH-FA-06`, `LH-FA-08` und `LH-FA-11` je Coverage-Zustand und je Modus von `STRICT_COVERAGE`; der Dienst importiert keinen Upstream-Typ und kein Paket eines anderen Service |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Die Fakes der Ports (Repositories, Codec) halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15).

**Aufträge an den Architect** (Start, Planner, 2026-11-07; vor dem ersten Code-Commit, `AGENTS.md` §3.12; Ausgang jeweils *offen* in §6; der Architect arbeitet im Zweig `slice-mock-lesen`):

1. **Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15, Spalten nach `SPEC-128` Nr. 3 a) und **Raster *Operation × Argumentart*** (`AGENTS.md` §3.16) in §6, jede Zelle mit Nummer in der Spezifikation oder „gilt nicht“ mit Grund. Operationen: Zuordnung (Service, Workspace), `$metadata`, Collection, Key-Lookup, `/$count`; Argumentarten: Header fehlt, leer, mehrfach, fremd (gehört zu einem anderen Service), unbekannt; Coverage je Zustand (`complete`, `partial`, `unknown`) × Methode × Modus von `STRICT_COVERAGE`; Abbruch (`ctx`); Wiederholung. Eine Aufrufstelle je Port (Service-, Workspace-, Entity-, Metadaten-Repository, `URLCodec`, Schreiber, Interpreter): der Fake hält jedes Argument fest und zählt jede Operation, auch die verbotenen (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 4×; `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test`, 4×).
2. **Die Sendungen des Aufrufers als Zeilen der Tabelle** (je Zusage ein Fänger, nie ein Haken im Plan): `Wurzel` aus `PUBLIC_BASE_URL` und Service-Root, ein anderer `Host` ändert `@odata.context` und `@odata.id` nicht (`SPEC-168` Nr. 10); `Content-Type` und `Kennzeichen` unverändert aus der `Darstellung` jeder Ausgabe, auch der Anzahl und des Fehlerobjekts (`SPEC-168` Nr. 7), der Dienst kennt keinen Namen eines Kennzeichens; `Accept` unverändert an den Schreiber (`SPEC-169` Nr. 3); die **fünf Pflichten** aus `SPEC-190` (Nr. 1 d Maske aus den Baseline-Properties, Nr. 2 `/$count` nur mit `Filter`, Nr. 3 f Key-Properties in `Properties`, Nr. 3 e `Schluessel` eindeutig und nichtleer, Nr. 3 g kein `redigiert` in gefilterten oder geordneten Properties, sonst `500` mit `ODR-E9001`) je mit einem Fänger, der ohne die Pflicht rot wird (`BEO-REPO/sendung-nennt-nicht-alle-pflichten-des-vertrags`); die Sendung von `slice-odata-url-pfad` (a) bis (e) aus §1; `Codecs` mit `nil` ist die V4-Implementierung der Konfig, der Dienst verzweigt nie nach der Version (`SPEC-175` Nr. 3), ein Feld des Satzes, das der Dienst nicht braucht, darf `nil` sein. Dazu die **Fehlerabbildung** der Ursachen von Parser und Interpreter auf `ODR-E6001`, `ODR-E6002`, `ODR-E6005`, `ODR-E6003`, `ODR-E9001` mit Status (Menge im Modell, nicht aus der Erinnerung).
3. **Pflicht-Mengen** (`AGENTS.md` §3.17, `implement-slice` Schritt 14) **mit Achsen**: je Klasse die Menge, die Zahl ihrer Elemente, die Achsen (zum Beispiel acht Typen des Schlüssels × Coverage-Zustand × Modus von `STRICT_COVERAGE` × Art der Ressource × Option), der Generator über die Klasse, das Orakel (nie der Dienst selbst) und die Prüfgrenze der unbeschränkten Mengen als akzeptiertes Negativ mit Grund und Folge im Betrieb (`BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze`, 11×; `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation`, 6×).
4. **Zweig- und Rückgaben-Menge mit Zählprüfung gegen eine Modell-Menge** (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, 5×; `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, 28×): die Menge der Zweige des Dienstes, die unter dem Detektor von `make test-race` laufen müssen, und die Menge der Rückgaben (Fehlerursachen, Coverage-Zustände, Arten der Ressource) werden aus einer Menge im Modell oder im Vertrag gebildet, und ein Test zählt, dass die Menge des Tests so groß ist wie diese Modell-Menge (Muster: `model.AlleEingabeUrsachen` aus `slice-odata-json-eingabe`); ein neuer Wert im Modell macht den Zähler rot. Die nebenläufige Zusage (Lesesicht bei gleichzeitiger Mutation, §6) bekommt einen Wirkungs-Fänger, der die Wirkung zählt, oder ein benanntes Negativ mit Grund; Raten mit Zahl der Prozesse und Messendem (`AGENTS.md` §3.10).
5. **Schnittfrage** nach `SPEC-128` Nr. 1: der Dienst bringt mehrere Fakes (Richtwert wenigstens sechs, *f* allein trägt 1200 zu *S* bei), *S* liegt voraussichtlich über 2223 (§4). Der Architect beantwortet sie vor der Tabelle: Kandidaten sind (A) Zuordnung und `$metadata` (DoD-Punkt 1) gegen (B) Collection, Key-Lookup und `/$count` mit Coverage-Prüfung (DoD-Punkte 2 und 3); sie teilen Konstruktor, Port und Zuordnung (`slice-mock-schreiben` baut auf allen dreien auf), und die Coverage-Prüfung liegt vor der Query-Ausführung, also in B. Schneidet er, entsteht ein Slice vor diesem und der Architect gibt den Schnitt mit Plan-Entwurf an den Planner zurück; schneidet er nicht, steht der Grund im Block `schnitt-grund` (Vorbild: `slice-v2-payload-lesen`, `slice-odata-url-filter`).
6. **Kennungen und Dateien neben der Linie B** (neben diesem Slice laufen `slice-v2-payload-schreiben` und `slice-v4-schreiber-marker-text`, beide Linie B): von `SPEC-160` bis `SPEC-199` sind `SPEC-192` bis `SPEC-199` frei, **dieser Slice nimmt `SPEC-195` bis `SPEC-199`**; `SPEC-192` bis `SPEC-194` bleiben für die beiden Nachbarn (`slice-v2-payload-schreiben` hat bis zum Start keine genommen, `SPEC-172` Nr. 6 h bis o und Nr. 7 hat er in `SPEC-172` ergänzt). Reichen fünf Nummern nicht, geht das als Rückgabe an den Planner (die Zähl­räume der Linien A und C sind getrennt; der Bereich von Linie A, `SPEC-129` bis `SPEC-159`, ist nicht ausgeschöpft, weil Linie A leer steht; die Zuweisung ändert nur der Planner in den Übergangsregeln). `ARC` ab 030 und `ADR` ab 0020 teilt die Linie B; braucht der Slice eine `ARC` oder `ADR`, geht das als Rückgabe an den Planner. **Dateien der Nachbarn:** dieser Slice ändert weder `codec_satz.go` noch den Wortlaut von `SPEC-168`, `SPEC-169` oder `SPEC-172` und schreibt in `src/adapters/driven/odata/` keine Datei (er ruft die Ports; `jsoncodec` und `jsoncodecv2` gehören den Nachbarn); er schreibt in `src/hexagon/services/mock/`, `src/hexagon/ports/inbound/`, die Tests und `spec/spezifikation.md` (neue Nummern ab `SPEC-195`, `SPEC-044`, `SPEC-011`, `SPEC-024`, `SPEC-070` nur lesend). Findet er in den Nachbarn eine Lücke (Schreiber), geht sie als Rückgabe an den Planner, nicht als Änderung dort. Die Merge-Reihenfolge entscheidet der erste Verifier-Pass; der zweite Zweig nimmt `main` per Merge auf.
7. **Fehlercodes und Katalog** (`SPEC-201` Nr. 3 und 11, `ODR-E6007` ist das Vorbild): bringt der Slice einen Code in `src/hexagon/model/fehler.go` (heute sind `ODR-E5003`, `ODR-E5005`, `ODR-E6001` bis `ODR-E6007`, `ODR-E7001` angelegt), setzt er ihn im selben Slice in `harness/katalog-ausnahmen.txt` (mit Grund: der Modus `mock` erreicht das Binary erst mit `slice-bootstrap-m3`, dessen Handbuch-Eintrag er nennt) und führt ihn in `§4` der Spezifikation; `make katalog-check` ist grün (Beleg im Bericht). Erreicht ein Code das Binary doch, geht der Handbuch-Eintrag mit einem Beispiel als Datei in den Slice (DoD). Die Datei der Ausnahmen gehört nicht zur Gate-Kette (Nr. 7 der Übergangsregeln nennt `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` und den Stempel); ändert der Slice eine dieser Dateien, auch `.a-check.yml` oder `.golangci.yml`, geht das als Rückgabe an den Planner.
8. **Das Service-Dokument** (Auftrag aus §1): den Beleg am Referenz-Client des M3-Tests vor dem ersten Code-Commit. Stand beim Start (Planner): weder `slice-bootstrap-m3` noch `welle-mock` §3 benennen einen Client, der M3-Test ruft `$metadata` und Abfragen, kein Service-Dokument; der Architect prüft den Wortlaut des Tests in `welle-mock` §3 und hält den Beleg (Request oder Fehlen) in §6 fest, bei Bedarf nach §1 (Mindestfassung, Rückgabe an den Planner), sonst bleibt es bei `501`.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Schleifen aus dem Wortlaut der Zusage, nie eine Liste der Fälle, die einfielen (`AGENTS.md` §3.17).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-query-interpreter` und `slice-odata-json-schreiben` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Stand beim Start** (Planner, 2026-11-07): erfüllt, `slice-query-interpreter`, `slice-odata-json-schreiben`, `slice-handbuch-meldungskatalog`, `welle-record-replay` (mit M2) und `slice-odata-json-eingabe` liegen in `done/`; `slice-v2-bootstrap` ist nicht Voraussetzung (die Konfig nimmt `Codecs` mit `nil` entgegen). Der Slice liegt in der Linie B neben `slice-v2-payload-schreiben` und `slice-v4-schreiber-marker-text` (Auftrag 6 in §3) und berührt keine Datei der Gate-Kette (Übergangsregeln Nr. 7; kein Slice der Linie C in `in-progress/`; `src/hexagon/services/**` ist in `.a-check.yml` als Schicht `services` schon geführt, die Ausnahmen-Datei des Katalogs gehört nicht zur Kette, Auftrag 7); ändert Vertrag oder Code eine Datei der Gate-Kette, geht das als Rückgabe an den Planner (§6).
- **Größenmarke** (`SPEC-128`; Faktoren allein dort; Stand beim Start, Planner). Der Block mit `basis`, `schaetzung` und `marke` steht auf `main` nicht (ein Block ohne Zahlen der Tabelle wäre ein Platzhalter und machte `make groessenmarke` rot); der Architect legt ihn im Zweig an, `basis` ist der letzte Commit, der Vertrag, Tabelle und Raster abschließt, und der Planner setzt Schätzung und Marke in eigenem Commit vor dem ersten Code-Commit. Pfade: `src/hexagon/services/mock/` und `src/hexagon/ports/inbound/` (Tests eingeschlossen). Keine Rechnung ohne Tabelle; die Schnittfrage stellt sich voraussichtlich (Auftrag 5 in §3).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt (Block im Zweig, Auftrag des Planners im Block der Größenmarke oben).
- `in-progress` → `open` (blockiert — Carveout?): die Zuordnung von Service und Workspace im Modus `mock` (welche Header, welcher Fehler) sind nicht entschieden, oder der Referenz-Client des M3-Tests fordert das Service-Dokument an und dessen Mindestfassung steht nicht in der Spezifikation (siehe §6).

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

- Zuordnung im Modus `mock` (Header fehlt, leer, mehrfach, gehört zu einem anderen Service): `LH-FA-04.c` regelt sie für Service und Workspace, das Raster prüft jede Zelle — **Ausgang:** offen; Zelle ohne Ort geht an den Architect.
- Das Service-Dokument (`GET` auf den Service-Root) hat in der Spezifikation keinen Vertrag und gehört nicht zum Muss-Umfang (Entscheidung des Users, 2026-10-30) — **Ausgang:** offen; Adresse: Architect (Auftrag in §1): Beleg am Referenz-Client des M3-Tests vor dem ersten Code-Commit; fordert er es an, Mindestfassung mit eigenem Vertrag und Rückgabe an den Planner, sonst `501`.
- Lesesicht bei gleichzeitiger Mutation („Workspace-Snapshot plus lokale Änderungen“, `LH-FA-07.a`): welcher Zustand ein Lesezugriff sieht — **Ausgang:** offen; Architect, Wirkungs-Fänger.
- Einzel-Lookup auf einem EntitySet mit Coverage `partial`: `404` oder `ODR-E6003` (`SPEC-044`, `LH-FA-07.b`) — **Ausgang:** offen; Raster, eine Zelle je Coverage-Zustand und Methode.
- Schnittfrage (Auftrag 5 in §3): *S* liegt voraussichtlich über 2223 — **Ausgang:** offen; Adresse: Architect (Schnitt oder `schnitt-grund`), Marke setzt der Planner nach der Tabelle.
- Die Sendungen des Aufrufers (fünf Pflichten aus `SPEC-190`, `Wurzel`, `Darstellung`, `Accept`) und die Sendung von `slice-odata-url-pfad` (a) bis (e) als Zeilen der Tabelle, je mit Fänger (Auftrag 2 in §3) — **Ausgang:** offen; Adresse: Architect (Zeilen), Implementer (Fänger).
- Zweig- und Rückgaben-Menge mit Zählprüfung gegen eine Modell-Menge (Auftrag 4 in §3; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`) — **Ausgang:** offen; Adresse: Architect (Menge und Zählprüfung im Vertrag), Implementer (Test).
- Der Parser ist bis zu diesem Slice unverdrahtet (Sendung von `slice-odata-url-pfad`, §6 dort): ein Defekt im Zusammenspiel von Parser, Interpreter und Dienst zeigt sich erst hier — **Ausgang:** weiter offen; Adresse: dieser Slice (Ende-zu-Ende-Fall je Typ der acht).
- Linie B neben `slice-v2-payload-schreiben` und `slice-v4-schreiber-marker-text` (Auftrag 6 in §3): Kennungen (`SPEC-195` bis `SPEC-199`), Dateien der Nachbarn, Reihenfolge der Merges — **Ausgang:** offen; Adresse: Architect, bei einer Berührung Planner.
- Fehlercodes und Katalog (Auftrag 7 in §3, `SPEC-201` Nr. 3 und 11): ein neuer Code ohne Zeile in `harness/katalog-ausnahmen.txt` oder Handbuch macht `make katalog-check` rot — **Ausgang:** offen; Adresse: Architect (Codes im Vertrag), Implementer (Ausnahme oder Kapitel).
- Gate-Kette (Stand beim Start in §4): ein Bedarf an einer Änderung der Gate-Kette — **Ausgang:** weiter offen (nicht zu erwarten); Adresse: Planner.
- `Codecs` und Satz: `nil` heißt die V4-Implementierungen der Konfig (`SPEC-175` Nr. 3); wie die Konfig sie bildet, bleibt `slice-bootstrap-m3` — **Ausgang:** offen; Raster, Zelle „`Codecs` gleich `nil`“.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-07; Zähler = Zahl der Dateien in `evidence/`). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, `AGENTS.md` §3.15), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×, `AGENTS.md` §3.15 und §3.16), `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (28×, `AGENTS.md` §3.17) und `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (21×, `AGENTS.md` §3.9), `BEO-REPO/zustand-des-arguments-als-dimension` (6×, `implement-slice` Schritt 19; Achse Zustand der Coverage), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14), `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×, Reviewer-Skill), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (8×, beide `AGENTS.md` §3.10) für die Lesesicht bei gleichzeitiger Mutation, `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (19×, `AGENTS.md` §3.14). Offen und für diesen Slice einschlägig: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) für die Fakes der Ports; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (5×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure von `welle-mock`) für die Zweig-Menge mit Zählprüfung (Auftrag 4 in §3, Mittel der Fix-Runde aus `slice-odata-json-eingabe`); `BEO-REPO/sendung-nennt-nicht-alle-pflichten-des-vertrags` (1×, Fund bei `slice-query-interpreter`: drei statt fünf Pflichten) für die Sendungen in §1 (die Tabelle führt je Nummer des Vertrags eine Zeile); `BEO-REPO/regel-in-zwei-adaptern-ohne-sensor` (1×) für den Ende-zu-Ende-Fall des Schlüssels je Typ (Sendung (d) aus `slice-odata-url-pfad`); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (13×, `geplant`: Neukalibrierung) für die Marke, die trotzdem der Formel folgt; `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (5×) für die Gegenproben bei wachsendem Produktcode. Der Slice hat Nebenläufigkeit (Lesesicht, Zähler) und viele Fakes; mit ihm erreicht kein offener Eintrag ohne Ausgang 3×, außer den genannten.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
