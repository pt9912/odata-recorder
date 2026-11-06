# Review-Bericht: slice-test-integration — 2026-11-06

**Rolle:** Reviewer (Modul 10) · **Frage:** Ist der Diff gegen Plan, Entscheidungen und Hard Rules wartbar und richtig? — nicht gegen die DoD (Verifier).

**Gegenstand:** `slice-test-integration`, Zweig `slice-test-integration`, HEAD `d30d04f`, Diff `9a6fb92..d30d04f` (23 Dateien, 1953 Einfügungen, 12 Löschungen). Bezug: `LH-FA-01`, `LH-FA-02`, `LH-FA-03`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-RB-03`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md).

**Geprüft gegen:** Plan `slice-test-integration` (§3 Tabelle mit 29 Zeilen, Mutationsabsatz, §4, §6 mit Raster und Entscheidungen), Vertrag `SPEC-037` Nr. 1 bis 10, `AGENTS.md` §3.2, §3.6, §3.7, §3.9, §3.10, §3.11, §3.15, §3.17, §3.18.

**Methode.** Einmal gefahren (Durchstich vor Härtung, keine Wiederholungsschleife). Der Bericht des Implementers ist nicht übernommen. Selbst gemessen, Docker nur unter `flock /tmp/odata-recorder-heavy.lock`, `--memory 2g`, je Mutation eine Wegwerfkopie des Arbeitsbaums im Scratchpad, ein Image, N Container (`-test.count=1`, je Lauf ein Prozess). Gemessen hat der Reviewer; die Raten unten sind seine, nicht die des Plans.

**Grenze dieses Laufs (ehrlich).** Nicht alle 29 Zeilen sind aufgelöst. Aufgelöst sind die sechs nebenläufigen Zeilen 7, 13, 14, 15, 18, 28 (mit je einem Element, das der Plan nicht nannte, bei 14, 15 und 28) sowie zwei Mutationen an Elementen der Klasse *Geheimnis* und *Ort der Suche* (Befunde M-1, M-2). Die deterministischen Zeilen 1 bis 6, 8 bis 12, 16, 17, 19 bis 27 und 29 sind gelesen, nicht mutiert; ihre Mutationen stehen im Plan und sind hier nicht nachgefahren. Die unmutierte Stufe ist mit einem vollen Lauf von `make test-integration` (7 von 7 grün, 7 s mit warmem Cache) gesehen, nicht in 20 oder 30 Läufen.

---

## Befunde

### Findings nach Schwere

**HIGH: keine.**

Zu (a), Schutz-Szenario Zeile 28: Am Produkt ergab sich kein Befund gegen `LH-FA-14`. Mit der Mutation `Authorization` aus der Standardmenge von `REDACT_HEADERS` wird `TestE2ESchutz/Speicher_voll` rot (30 von 30, siehe unten), das Geheimnis liegt dann im `-wal` vor und in der Datenbank nach dem Herunterfahren; unmutiert steht kein Header-Geheimnis in einer Datei. Für das Body-Feld `password` trägt der Test das Urteil nicht (M-1), also ist „das Produkt hält `password` bei vollem Speicher aus Dateien heraus" hier **nicht belegt**, aber auch **nicht widerlegt**; ein Befund am Produkt ist es nicht.

---

#### M-1 (MEDIUM): Das Body-Feld `password` in Zeile 28 hat keinen Fänger, der rot wird

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Zusage ohne Fänger, Fänger schmaler als Zusage), §3.17; `LH-FA-14` und `LH-FA-14.a`; `SPEC-037` Nr. 7 (Schutz, Negative) und Nr. 9
- `pfad`: `tests/e2e/stufe_schutz_test.go` · `post := postJSON(`{"Name":"x","password":"` + passwort + `"}`)` und `geheimnisse := []string{geheimA, geheimD, passwort}`
- `befund`: Zeile 28 und `SPEC-037` Nr. 7 sagen zu, dass das Body-Feld `password` bei vollem Speicher weder vor noch nach dem Herunterfahren in einer Datei steht. Die Mutation „`$..password` aus der Standardmenge von `REDACT_JSON_PATHS` entfernt" (`src/bootstrap/konfig_werte.go`, Zeile `"$..password", ...`) bleibt grün: 0 von 30 Läufen in `TestE2ESchutz/Speicher_voll`, 0 von 3 Läufen der ganzen Stufe `TestE2ESchutz`. Bei `ENOSPC` wird der Body-Blob nicht geschrieben, also liegt der Wert auch ohne Redaktion in keiner Datei; die Vorbedingung der Zusage verhindert die Beobachtung, die der Fänger sucht. Kein anderer Test der Stufe führt ein Body-Feld mit sensiblem Namen, die Redaktion von JSON-Feldern ist über die Prozessgrenze damit ungeprüft. Die Mutationstabelle des Plans (Zeile 28) nennt für `password` keine Mutation, nur `Authorization`.
- `verifizierbar`: ja (Mutation oben, `make test-integration` auf Wegwerfkopie)
- `klasse`: Fänger beobachtet nur, was seine Vorbedingung ohnehin verhindert

#### M-2 (MEDIUM): Bei vollem Speicher suchen die Fänger die Geheimnisse nicht in stdout und stderr

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.17; `SPEC-037` Nr. 9 (*„für ihn gilt die Suche vor und nach dem Herunterfahren in jedem Fall, auch bei vollem Speicher"*; gesucht wird *„in jeder Datei ... und in stdout und stderr"*); `LH-FA-14`, `LH-QA-04` (Geheimnis in Log)
- `pfad`: `tests/e2e/stufe_schutz_test.go` · `func speicherVollSchutz` (der Fänger ruft `durchsuche`, nie `sucheStroeme`)
- `befund`: `speicherVollSchutz` durchsucht nur Dateien. Mutation: in `schliesseAb` (`src/hexagon/services/record/ende.go`) vor `x.protokolliereAblage(actx, err)` jeden Wert von `x.anfrage.Header` mit `println` auf stderr ausgeben: 0 von 10 Läufen rot. Dass die Mutation wirkt, ist gesondert belegt: ein Probe-Lauf mit zusätzlichem `t.Errorf` auf `p.fehl` zeigt `Bearer GEHEIM-A-4f7c1e` und `sid=GEHEIM-B-9d2a60` auf stderr. Ein Geheimnis im Fehlerpfad der Ablage, der nur bei vollem Speicher läuft, erreicht eine diagnostische Senke, und kein Test der Stufe führt diesen Pfad mit einer Suche in den Strömen (`geheimnisse` sucht Ströme, läuft aber mit gelingender Ablage).
- `verifizierbar`: ja
- `klasse`: Allaussage über Orte (Dateien und Ströme) als Fänger über nur einen Ort

---

#### L-1 (LOW): Plan §6 trägt einen wörtlichen Zeilenumbruch als Zeichenfolge

- `kategorie`: LOW
- `quelle`: Maintainability; Plan §6 verlangt je Risiko genau einen Ausgang
- `pfad`: `docs/plan/planning/done/slice-test-integration.md` · `(Neurechnung nur, wenn sie unter 0,60 liegt).\n- Netzlos mit zwei Servern`
- `befund`: Zwei Aufzählungspunkte (Auslastung 0,47 und *Netzlos mit zwei Servern*) stehen in einer Zeile, getrennt durch die Zeichen Backslash, `n`, Minus; im gerenderten Text ist der zweite Punkt Teil des ersten, und seine Zeile *Ausgang: entfallen* hängt am Ausgang *offen* des ersten.
- `verifizierbar`: ja (`grep -n '\\n- '` auf die Datei)
- `klasse`: Escape-Folge im Schreibwerkzeug nicht gegengelesen (`AGENTS.md` §3.18)

#### L-2 (LOW): Zwei Kommentare der Stufe Schutz sagen mehr zu, als der Code hält

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.7, §3.11
- `pfad`: `tests/e2e/stufe_schutz_test.go` · `// secrets sind die drei Geheimnisse mit dem Ort, an dem sie erlaubt sind`; ebenda `// speicherVollSchutz: bei vollem Speicher nach Beginn der Antwort bleibt der Status, und der Marker im`
- `befund`: `secrets()` liefert nur drei Zeichenketten, keinen Ort. Der Kopf von `speicherVollSchutz` nennt allein den Marker im Request-Header als nicht abgelegt; der Marker ist seit der Entscheidung des Architect ein Spurwert (vor dem Herunterfahren `t.Logf`, nach ihm rot), die geprüften Geheimnisse (`Authorization`, `password`) fehlen im Kopf.
- `verifizierbar`: nein (kein Gate liest Kommentare)
- `klasse`: Kommentar sagt mehr oder anderes zu, als die Prüfung fängt

---

#### I-1 (INFO): Pflichtlauf ohne Träger

`harness/README.md` (Sensors-Tabelle, Zeile `make test-integration`) und Plan §6 nennen den Pflichtlauf vor jeder Welle-Closure und bei Änderungen an `src/bootstrap` und `src/adapters/driving`. Nichts erzwingt ihn, solange das Ziel neben `make gates` liegt (Entscheidung User offen, Plan §6 *Gate-Kette oder daneben*). Zuständig: User/Planner. Kein Befund am Diff; `AGENTS.md` §3.6 ist nicht berührt: `GATE_CHECKS` ist unverändert, kein Gate gelockert.

#### I-2 (INFO): `tests/e2e` läuft in `make test-race` nur als Übersetzung

`SPEC-037` Nr. 10 g nennt es mit Grund und Folge. Die Tests teilen Zähler und Puffer zwischen Goroutinen (`atomic`, `puffer` mit Sperre) und sind nie unter dem Detektor gelaufen; `p.exit` wird nach `close(p.ende)` gelesen und ist geordnet. Kein Befund, nur dass „Race-sauber" hier ungeprüft ist.

#### I-3 (INFO): Dockerfile, Stufe `integration`

Zwischen der Stufe `integration` und dem Kommentar `# --- handbuch` fehlt die Leerzeile, die die übrigen Stufen trennt; die Stufe davor hat zwei. Ohne Konventions-Anker, kein Finding.

---

## Beantwortung der benannten Prüfpunkte

**(b) Treiber `tools/harness/test-integration.sh`.** Selbst gefahren: Argument `x` und `''` ergeben `Aufruf:` und Exit 2 ohne Container (zwei Läufe). `SIGTERM` und `SIGINT` an den Treiber während einer Stufe: Exit 2, `docker ps` danach ohne Container der Stufe, Wegwerfverzeichnis entfernt. Stufe grün nur bei Exit 0, Zeile `--- PASS: <Name> ` und ohne `--- SKIP`/`--- FAIL` (gelesen; Mutationen des Plans 2 bis 4 nicht nachgefahren). `--network none`, `--memory 2g`, `--read-only`, zwei `tmpfs`, kein Mount: im Skript, so gelesen; die Verdrahtung außer dem Netz-Fall ist akzeptiertes Negativ (`SPEC-037` Nr. 10 f). Kein Befund.

**(c) Ausnahme in `.golangci.yml`.** Muster `^tests/e2e/uebergang/main\.go$` mit Text `os\.Std(out|err)`: eine Datei, kein Verzeichnis, mit `# Why:`-Block. `forbidigo` gilt für alles andere unverändert. Keine Lockerung nach `AGENTS.md` §3.6; Entscheidung des Architect im Plan §6 stimmt mit dem Diff überein. `uebergang.go` schreibt über übergebene `io.Writer`. Kein Befund.

**(d) Raten (vom Reviewer gemessen, je Lauf ein Container).**

| Zeile | Mutation (Element) | roter Test, Ausgabe | Rate |
|---|---|---|---|
| 7 | `SIGTERM` nicht angemeldet (`cli.go`, auf `SIGUSR1` umgestellt; die Streichung allein übersetzt nicht) | `TestE2EProzess/SIGTERM`: `SIGTERM: Exit -1, erwartet 0` | 20 von 20 |
| 13 | Aufruf von `protokolliereAblage` entfernt (`ende.go`) | `TestE2EAufzeichnung/Speicher_voll`: `0 von 2 Zeilen auf stderr mit ["ereignis=ablage" "error_code=ODR-"]` | 20 von 20 |
| 14 | Replay lässt die erste Aufnahme der Session aus (Element, das der Plan nicht nannte, `replay.go`) | `TestE2EPersistenz/Neustart`: `/odata/$metadata: Wiedergabe weicht ab: Status` | 20 von 20 |
| 15 | Replay der leeren Session liefert `ODR-E5003` statt `ODR-E5001` (`replay.go`; andere Mutation als im Plan) | `TestE2EPersistenz/leer`: `Replay /odata/Products(1): 400 ..., erwartet 404 mit ODR-E5001` | 20 von 20 |
| 18 | Replay ruft den Upstream bei fehlendem Treffer (`replay.go`, eigene Fassung) | `TestE2EReplay/Fehlerfälle`: `Zähler (Verbindungen, Requests) [2 9], vor dem Replay [1 8]` und alle weiteren Fälle | 20 von 20 |
| 28 | `Authorization` aus `REDACT_HEADERS` (`konfig_werte.go`), ganze Stufe | `Geheimnisse` und `Speicher_voll`: `GEHEIM-A-4f7c1e ... vor dem Herunterfahren in odata-recorder.db-wal` | 30 von 30 |
| 28 | dieselbe Mutation, nur `TestE2ESchutz/Speicher_voll` | `Geheimnis GEHEIM-A-4f7c1e` und `GEHEIM-D-2c71` vor dem Herunterfahren im `-wal`; `GEHEIM-A-4f7c1e liegt nach dem Herunterfahren in odata-recorder.db` | 30 von 30 |
| 28 | `$..password` aus `REDACT_JSON_PATHS` | **grün** (M-1) | 0 von 30 |
| 28 | Geheimnis auf stderr im Fehlerpfad der Ablage | **grün** (M-2) | 0 von 10 |

Die Zahlen 20 von 20 und 30 von 30 des Plans sind für diese Zeilen bestätigt. Die unmutierten Läufe (20 und 30) habe ich nicht wiederholt.

**(e) Gate nicht gelockert.** `GATE_CHECKS` unverändert, `make test-integration` steht daneben (Fragment `harness/mk/test-integration.mk`, Sensors-Tabelle als *kein Gate*); die Zeile *Nicht behauptet* in `harness/README.md` nennt die ungeprüfte Verdrahtung und die fehlenden Szenarien. Kein Befund.

**Rohbytes (`AGENTS.md` §3.18).** Kein `\uXXXX` in `tests/e2e`, `tools/harness/test-integration.sh`; kein BOM, NBSP oder Zeilentrenner in Tests, Treiber, Vertrag, Plan (`LC_ALL=C grep -P`). Die Fixtures bilden keine Werte aus Escape-Folgen.

**Raster (Plan §6).** Gelesen gegen `SPEC-037`; jede Zelle trägt Nr. oder „gilt nicht" mit Grund. Zellen *Geheimnisse suchen · Wiederholung* (vor und nach dem Herunterfahren) und *Platte füllen · Abbruch* (Zeilen 13, 28) sind die, die M-1 und M-2 berühren: der Ort ist benannt, der Fänger deckt ihn für Dateien und nur für Header-Geheimnisse.

---

## Geprüft, ohne Befund

- `tools/harness/` (Treiber `test-integration.sh`): geprüft, ohne Befund.
- `harness/mk/` und `repo.mk`: geprüft, ohne Befund.
- `Dockerfile` (Stufe `integration`): geprüft, ohne Befund (I-3 ist Hinweis).
- `.golangci.yml`: geprüft, ohne Befund.
- `harness/README.md`: geprüft, ohne Befund (I-1 Hinweis).
- `tests/e2e/uebergang/` und `tests/e2e/fixture/`: geprüft, ohne Befund.
- `tests/e2e/` (Stufen Prozess, Weiterleitung, Aufzeichnung, Persistenz, Replay, Matching): gelesen, Zeilen 7, 13, 14, 15, 18 mutiert, ohne Befund; Stufe Schutz: M-1, M-2, L-2.
- `harness/vertraege/SPEC-037-test-integration.md`, `docs/plan/planning/e2e-abdeckung.md` (`make e2e-abdeckung`: unverändert, 7 Stufen, 7 Deklarationen): geprüft, ohne Befund.
- `docs/plan/planning/done/slice-test-integration.md`: L-1.
- Produktcode `src/`: nicht geändert, kein Befund am Produkt (kein Geheimnis in einer Datei bei unmutiertem Lauf).

## Steering-Loop (Zähler)

- *Fänger beobachtet nur, was seine Vorbedingung verhindert* (M-1): erstes Auftreten.
- *Allaussage über Orte, Fänger über einen Ort* (M-2): schärft *Fänger führt nur eine Form der Eingabe* (Reviewer-Skill, MEDIUM, `AGENTS.md` §3.15); Zähler zählt mit.
- *Escape-Folge nicht gegengelesen* (L-1): `AGENTS.md` §3.18 trägt die Regel, Plan-Datei nicht vom Gegenlesen erfasst.
