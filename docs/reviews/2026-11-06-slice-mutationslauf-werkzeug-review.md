# Review-Bericht: slice-mutationslauf-werkzeug — 2026-11-06

**Rolle:** Reviewer (Modul 10) · **Frage:** Ist der Diff gegen Plan, Entscheidungen und Hard Rules wartbar und richtig? — nicht gegen die DoD (Verifier).

**Gegenstand:** `slice-mutationslauf-werkzeug`, Zweig `slice-mutationslauf-werkzeug`, HEAD `b0eee68`, Diff `9c9b406..b0eee68` (17 Dateien, 3177 Einfügungen, 6 Löschungen; Code `ee83825`, `7f4bc8c`, Plan `b0eee68`). Bezug: `LH-QA-06`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0004](../plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md).

**Geprüft gegen:** Plan `slice-mutationslauf-werkzeug` (§1 Ziel und Abgrenzung, §2 DoD, §3 Tabelle Z1 bis Z13 und Absatz *Umsetzung des Implementers*, §4 Messung, §5, §6 Raster und Risiken), Vertrag `SPEC-210` Nr. 1 bis 12, `SPEC-128` Nr. 5 d, `AGENTS.md` §3.1, §3.2, §3.6, §3.7, §3.9, §3.10, §3.11, §3.12, §3.14, §3.15, §3.16, §3.17, §3.18, Reviewer-Skill.

**Methode.** Einmal gefahren (Durchstich vor Härtung, keine Wiederholungsschleife). Der Bericht des Implementers ist nicht übernommen; jede Zahl unten stammt aus eigenen Läufen des Reviewers. Docker nur unter `flock /tmp/odata-recorder-heavy.lock`, Container `odata-recorder:race` mit `--memory=2g --memory-swap=2g --cpus=4 --network none`, Wegwerfkopien per `docker cp` (kein Mount), alle Container und Kopien nach Gebrauch entfernt (`docker ps -a` mit meinem Label: 0). Das Modul `tools/mutationslauf` fährt `make gates` nicht (SPEC-210 Nr. 12 d); der Reviewer fuhr seine Tests selbst. `make gates` und `make groessenmarke` habe ich nicht gefahren; `make docs-check` und `make zeichen-check` auf dem Stand mit diesem Report: grün.

- **Grundlauf** (Image `odata-recorder:race`): `gofmt -l` leer, `go vet ./...` ohne Befund, Go 1.27.1. Die schnellen Tests (ohne `TestHaengerNachGnade`, `TestAbbruchRaeumtAuf`, `TestTreiberFrist`) grün in 29 s, die Tests mit echten Prozessen (`TestKopieJeMutation`, `TestGleichwertigWieEinzellauf`, `TestBaumUnveraendert`, `TestUebersetzungsfehlerIstKeinRot`, `TestNullprobe`, `TestCacheAendertAussageNicht`) grün in 83 s, `TestAbbruchRaeumtAuf/SIGTERM` in 0 von 20 Prozessen rot.
- **Mutationen:** 89 Mutationen an einer Wegwerfkopie des Codes (Programm und Treiber), je einmal; zuerst die schnellen Tests, bei Überleben zusätzlich die echten (`-failfast`). 6 Mutationen übersetzen selbst nicht und zählen nicht (K6, M4, E5, L15, S21, S29). Von 83 gültigen blieben 47 grün (darunter äquivalente und die unten als „kein Finding“ genannten), 36 wurden rot. Die Tabelle steht unter *Aufgelöste Zeilen und Mutationen*. Die Mutationen sind meine eigenen Konstruktionen; die Buchstaben K, M, E, T, L, S, A ordnen sie nach Datei (Kopie, Main, Echt, Treiber, Lauf, Liste, Anker).
- **Raten** (§3.10, nebenläufige Zeilen Z6, Z8, Z12): selbst gemessen, Messender der Reviewer, siehe *Raten*.
- **Messung des Gewinns:** der Treiber end-to-end auf dem Satz A des Implementers (Liste nur in dessen Scratchpad), drei Läufe; das *bisherige Vorgehen* mit dessen Skript `baseline.sh` nachgefahren, drei Läufe. Last des Hosts je Zahl genannt, siehe *Messung des Gewinns*.

**Grenze dieses Laufs (ehrlich).** Nicht aufgelöst sind: Z2 (die acht Arten gegen den Einzellauf) außer über die Mutationen der Tabelle unten und den Grundlauf; Z9, Z10, Z11 nur über die Mutationen T1 bis T5, T11 (die des Implementers zu Z9 bis Z11 habe ich nicht wiederholt); `TestHaengerNachGnade` und `TestTreiberFrist` (60 s und 2,5 Minuten Wanduhr, nicht gefahren); ein Lauf mit 200 Mutationen oder mit einem Paket, dessen Bau teuer ist; SIGHUP und SIGINT am Treiber (nur gelesen); die Zeile Z12 *Start neben einem Bau* (nur *zwei Baue zugleich* gefahren). Der Host war nicht ruhig: die Last (Minutenwert) lag im Verlauf zwischen 4 und 37; die Werte stehen je Zahl.

---

## Befunde

### Findings nach Schwere

**HIGH: keine.** Zu (b): die Gate-Kette ist unberührt (`git diff --name-only 9c9b406..b0eee68` nennt außer `tools/mutationslauf/*`, `tools/harness/mutationslauf.sh` und dem Slice-Plan keine Datei; kein Ziel in `harness/mk/`, keine Stufe im `Dockerfile`, kein `GATE_CHECKS`, `harness/README.md`, Sensors-Tabelle und `.golangci.yml` unberührt). `AGENTS.md` §3.1: der Treiber legt `create` mit `--network none` an, kein `-v`, `--volume`, `--mount`, `--tmpfs`, die Eingaben gehen per `docker cp`, der Treiber schreibt nichts auf den Host. §3.6: keine Schwelle gesenkt; §3.2: keine Suppression-Direktive im Diff.

---

#### M-1 (MEDIUM): Der Umfang der Wegwerfkopie (SPEC-210 Nr. 4 a) hat weder eine Tabellenzeile noch einen Fänger

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Zusage ohne Fänger, Fänger führt nur eine Form), §3.17; `SPEC-210` Nr. 4 a
- `pfad`: `tools/mutationslauf/kopie.go` · `for _, d := range []string{"src", "tests"} {` und `return fs.SkipDir` und `case d.Type()&fs.ModeSymlink != 0:`
- `befund`: Nr. 4 a sagt zu: Wurzelmodul gleich `go.mod`, `go.sum`, `src`, `tests`; verschachtelte Module nicht Teil der Kopie. Fünf Mutationen an `kopie.go` bleiben mit allen Tests grün (je rund 100 bis 125 Sekunden Lauf der schnellen und der echten Tests): `tests` nicht mitkopieren (K3), `go.sum` nicht mitkopieren (K4), verschachtelte Module mitkopieren (K1, `SkipDir` durch `MkdirAll`), Symlinks nicht übernehmen (K2), Dateimodus der Kopie auf `0o644` begrenzen (K6b, Ausführbit verloren). Die Tabelle Z1 bis Z13 hat für Nr. 4 a keine Zeile; die Fixture hat weder `tests/`, noch `go.sum`, noch ein verschachteltes Modul im kopierten Baum, noch einen Symlink.
- `verifizierbar`: ja (Mutation oben auf Wegwerfkopie, `go test -count=1 .` im Image `odata-recorder:race`)
- `klasse`: Zusage ohne Fänger — Menge der Kopie (Elemente einzeln ungefangen)

#### M-2 (MEDIUM): Der echte Cache-Adapter (Größe messen, Cache leeren) hat keinen Fänger; Z12 prüft nur den Fake

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger schmaler als Zusage), §3.10; `SPEC-210` Nr. 8 d, Nr. 9 (Speicher als Anforderung des Users)
- `pfad`: `tools/mutationslauf/echt.go` · `summe += i.Size()` und `"go", "clean", "-cache"`
- `befund`: Nr. 8 d verlangt: über 1,5 GiB ruft das Programm `go clean -cache`. Z12 fängt die Schwelle am Fake (`Cache 1610612737: 0-mal geleert`). Der echte Adapter `echt.CacheGroesse`/`CacheLeeren` hat keinen Test: `summe += 1` statt `i.Size()` (E6), `IsRegular()` entfernt (E10) und `go clean -modcache` statt `-cache` (E7) bleiben grün. E7 löscht im Container den Modul-Cache; die folgenden Mutationen bauen ohne Netz nicht mehr (`uebersetzungsfehler` oder `abbruch` statt einer Aussage), und der Fehler zeigt sich erst im Betrieb nach dem Überschreiten der Grenze.
- `verifizierbar`: ja (Mutationen E6, E7, E10)
- `klasse`: Zusage nur am Fake gefangen — echter Adapter ungeprüft

#### M-3 (MEDIUM): Optionen des Containers und des Baus aus SPEC-210 Nr. 3 b, 5 b, 9 a sind überwiegend nicht gepinnt

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.10; `SPEC-210` Nr. 3 b (`--cpus=4 --pids-limit=1024 --label`), Nr. 5 b und 9 a (`-trimpath -buildvcs=false -p 2`), Nr. 10 c (Label trägt das Aufräumen von Resten)
- `pfad`: `tools/harness/mutationslauf.sh` · `--cpus=4 --pids-limit=1024` und `--label odata-recorder.mutationslauf=1 -e GOMEMLIMIT`; `tools/mutationslauf/echt.go` · `"-trimpath", "-buildvcs=false", "-p", "2"`
- `befund`: Von den Optionen des Containers pinnen `TestTreiberOptionen` und `TestTreiberUmgebung` nur `--network none`, `--memory`, `--memory-swap`, `GOMEMLIMIT` und die Abwesenheit von Mounts. Grün bleiben: `--pids-limit` entfernt (T1), `--cpus=8` (T2), `--label` aus `create` entfernt (T3; dann findet der nächste Aufruf die Reste nach Nr. 10 c nie), `-e MUTATIONSLAUF_GRENZE` entfernt (T4; das Programm nähme 3600 Sekunden, der Treiber tötet nach `grenze + 60` ohne Ergebniszeilen), `-e MUTATIONSLAUF_KALT` entfernt (T5), die Umgebung des Baus im `befehl` (`CGO_ENABLED=0 GOFLAGS=-mod=readonly GOTOOLCHAIN=local GOPROXY=off`, T11). Im Programm bleiben grün: `-trimpath` (E1), `-buildvcs=false` (E2) und `-p 2` (E3) entfernt, `GOPROXY=off` (M6), `GOTOOLCHAIN=local` (M7), `GOFLAGS=-mod=readonly` (M8) entfernt. `-p 2` ist der Speicherschutz der Bau-Spitze, den der User als Anforderung gestellt hat (Plan §6 *Speicher und Platz*); `-trimpath` bestimmt, ob der Cache des Images trifft.
- `verifizierbar`: ja
- `klasse`: Zusage nur am Fake gefangen — Optionen von Container und Bau ungepinnt

#### M-4 (MEDIUM): Die Diagnose (SPEC-210 Nr. 1 c) ist nur für `rot-paket` gefangen, nicht für die drei anderen Ergebnisse

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe), §3.17; `SPEC-210` Nr. 1 c
- `pfad`: `tools/mutationslauf/lauf.go` · `case ergUebersetz, ergRotPaket, ergFlattern, ergAbbruch:`
- `befund`: Nr. 1 c: bei `uebersetzungsfehler`, `rot-paket`, `flatternd`, `abbruch` stehen die ersten 40 Zeilen auf stderr. `ergRotPaket` aus der Liste zu streichen macht `TestDiagnose` rot (L1); `ergFlattern` (L2), `ergAbbruch` (L3) und `ergUebersetz` (L4) zu streichen bleibt grün. Der Fehlertext des Compilers, der einzige Hinweis auf die Ursache eines `uebersetzungsfehler`, würde nie ausgegeben. Zusätzlich bleibt grün, bei `flatternd` nur den Text des Arms mit Detektor auszugeben (`mit.text` statt `mit.text + ohne.text`, L8).
- `verifizierbar`: ja (Mutationen L2, L3, L4, L8)
- `klasse`: Fänger führt nur ein Element der Menge (Ergebnisse mit Diagnose)

#### M-5 (MEDIUM): Ein Signal während der Nullprobe ist nicht gefangen — Exit 2 statt Exit 3

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15, §3.16 (Abbruch × jede Operation); `SPEC-210` Nr. 10 a, Raster *Test fahren × Abbruch*
- `pfad`: `tools/mutationslauf/lauf.go` · `len(befunde) > 0 && ctx.Err() == nil`
- `befund`: Nr. 10 a: auf SIGTERM/SIGINT endet das Programm mit Exit 3. `TestAbbruchRaeumtAuf` sendet das Signal erst, wenn der hängende Test der ersten Mutation läuft. Ohne die Bedingung `ctx.Err() == nil` (L13) endet ein Signal in der Nullprobe, der längsten Phase (rund 20 Sekunden bei Satz A), mit Exit 2 und der Meldung `Ausgangsstand rot` statt mit Exit 3 und `abbruch`; kein Test wird rot. Die Zelle des Rasters (*Test fahren × Abbruch*) nennt Nr. 10 a ohne Fänger für diese Phase.
- `verifizierbar`: ja (Mutation L13)
- `klasse`: Randform-Zelle ohne Fänger (Abbruch in einer Phase)

#### M-6 (MEDIUM): Die Zeichenmenge von `name` und `test` ist nicht als Schleife über alle Elemente gefangen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (ein Fall je Element einer Menge), §3.17; `SPEC-210` Nr. 2 b
- `pfad`: `tools/mutationslauf/liste.go` · `r == '.' || r == '_' || r == '-'` und `testZeichen(r rune) bool { return nameZeichen(r) && r != '.' && r != '-' }` und `strings.HasPrefix(m.Test, "Test")`
- `befund`: Nr. 2 b erlaubt für `name` die Zeichen `A-Za-z0-9._-`. Die Ränder der drei Bereiche sind gefangen (S1 bis S6 rot; S1 und S3 allerdings nur vom 83-Sekunden-Test `TestGleichwertigWieEinzellauf`, weil dessen Namen `z` und `Z` enthalten, nicht von `TestListeForm`). Die Zeichen `.` (S7), `_` (S8) und `-` (S9) aus `name` zu streichen bleibt grün; für `test` bleibt grün, dass `.` zulässig wird (S11, im Muster `^Test.A$` ein Platzhalter), und dass nur `T` statt `Test` als Präfix gilt (S12). `TestListeForm` führt kein Name-Beispiel mit diesen Zeichen. `-` in `test` und die Länge sind gefangen.
- `verifizierbar`: ja (Mutationen S7, S8, S9, S11, S12)
- `klasse`: Fänger führt nur einen Teil der Zeichenmenge

#### M-7 (MEDIUM): Das Paar `alt` leer und `neu` leer ist im Code und im Test entschieden, nicht im Vertrag

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.12, §3.16 (eine im Code entschiedene Zelle ist ein MEDIUM); `SPEC-210` Nr. 2 b, 2 d, Nr. 11 Zusatztabelle
- `pfad`: `tools/mutationslauf/liste.go` · `m.Neu == m.Alt && m.Alt != ""`
- `befund`: Nr. 2 b nennt `neu` gleich `alt` als Befund `neu ändert nichts`; `alt` leer ist ein eigener Befund. Bei `alt` und `neu` beide leer liefert der Code nur `alt leer`; `TestListeZusatztabelle` (Fall *alle Felder einer Mutation*) hält es fest. Die Zusatztabelle des Vertrags hat die Zeile nicht, die Auslegungen des Plans (§3) nennen sie nicht. Fällt die Bedingung `&& m.Alt != ""`, wird der Test rot (S28): der Fänger ist da, die Entscheidung steht an keinem Ort der Spezifikation.
- `verifizierbar`: ja (Mutation S28 rot)
- `klasse`: Paar von Verstößen im Code entschieden

#### M-8 (MEDIUM): Die Messung des Gewinns (DoD Punkt 2) ist aus dem Repo nicht wiederholbar

- `kategorie`: MEDIUM
- `quelle`: Plan `slice-mutationslauf-werkzeug` §2 (*benannter, wiederholbarer Satz von Mutationen*), §5 (*derselbe Satz*); `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`
- `pfad`: Plan · `*Satz A:* fünf Mutationen in` und `Liste und Skripte nicht im Repo`
- `befund`: Satz A, Satz B und das *bisherige Vorgehen* (Nachbau nach §1 Befund 1 und 2) liegen nur im Scratchpad des Implementers (`messsatz.json`, `baseline.sh`, `subst`); der Plan beschreibt sie in Prosa (Datei, Zahl, Klasse), nennt aber weder die `alt`/`neu`-Texte noch die Tests je Mutation noch das Skript. Ein Verifier oder Planner kann den Satz nicht neu aufbauen, sobald die Scratchpads weg sind; der Closure-Trigger in §5 (*derselbe Satz mit dem bisherigen Vorgehen und mit dem Werkzeug*) ist dann nicht prüfbar. Ich konnte die Messung nur wiederholen, weil die Dateien im Scratchpad noch lagen.
- `verifizierbar`: nein (kein Gate; Beleg ist das Fehlen der Liste im Repo)
- `klasse`: Messung nicht wiederholbar

---

#### L-1 (LOW): Der Kopf des Treibers sagt „in jedem Fall“, der Vertrag nennt SIGKILL als Ausnahme

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.11
- `pfad`: `tools/harness/mutationslauf.sh` · `Der Container wird in jedem Fall entfernt.`
- `befund`: Die Tests fangen Erfolg, Exit 2, scheiternden `cp` und HUP/INT/TERM. Bei SIGKILL des Treibers bleibt der Container (SPEC-210 Nr. 10 c: höchstens einer mit Label, der nächste Aufruf räumt ihn); der Kopf nennt die Grenze nicht.
- `verifizierbar`: nein
- `klasse`: Kopf sagt mehr zu, als der Test fängt

#### L-2 (LOW): Die Tabelle Z1 bis Z13 hat keine Zeile für Aufruf und Umgebung (Nr. 1 a), Image (Nr. 3 a), Kopie-Umfang (Nr. 4 a)

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15 (jede Zusage eine Zeile)
- `pfad`: Plan · `| Z13 | Die Liste wird vollständig geprüft`
- `befund`: Für Nr. 1 a und 3 a gibt es Tests (`TestRunEingabe`, `TestImageAbgleich`, `TestTreiberAufrufUndImage`, `TestUmgebung`), im Absatz des Implementers fehlt die Zuordnung zu einer Zeile; für Nr. 4 a fehlen Zeile und Test (M-1).
- `verifizierbar`: nein
- `klasse`: Tabelle unvollständig gegen den Vertrag

#### L-3 (LOW): Das Binary der Detektor-Art bleibt bis zum Ende des Aufrufs liegen; ein Bau für einen leeren Arm läuft trotzdem

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.15; `SPEC-210` Nr. 4 b, 7 b, Nr. 9 (Speicher und Platz)
- `pfad`: `tools/mutationslauf/lauf.go` · `for _, p := range []string{kopie, kopie + ".test", kopie + ".test.race"} {` und `if m.NDetektor > 0 {`
- `befund`: Das Entfernen von `kopie + ".test.race"` je Mutation (L11) und der Wegfall der Bedingung `m.NDetektor > 0` (L9, Bau mit Detektor auch bei Arm 0) lassen alle Tests grün; ebenso das Entfernen des Race-Binary der Nullprobe (L12). Im ersten Fall wächst die Ablage um ein Race-Binary je nebenläufiger Mutation bis zum Ende des Aufrufs (die Ablage liegt im Dateisystem des Containers, kein tmpfs), im zweiten kostet ein Arm mit 0 einen Bau. Z12 prüft die Ablage nach dem Lauf, wenn `lauf` entfernt ist.
- `verifizierbar`: ja (L9, L11, L12)
- `klasse`: Zusage nur am Ende des Laufs geprüft

#### L-4 (LOW): Randform *`null` als Wurzel, Liste oder Eintrag* ist im Code entschieden

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.12 (Beispiel `null`), §3.16
- `pfad`: `tools/mutationslauf/liste.go` · `err != nil || wurzel == nil` und `err != nil || elemente == nil` und `err != nil || obj == nil`
- `befund`: Die drei Prüfungen auf `null` zu entfernen (S24, S25, S26) lässt alle Tests grün; nur der Wortlaut des Grundes ändert sich (`kein JSON-Objekt` oder `falscher Typ` gegen `Liste leer`, `Feld mutationen fehlt`, `Mutation n: kein Objekt`), Exit 2 bleibt. Der Vertrag nennt keinen Grund für `null`; der Plan führt die Gründe als Auslegungen (§6).
- `verifizierbar`: ja
- `klasse`: Randform im Code entschieden (Wortlaut)

#### I-1 (INFO): Zu (a) — der Zweck „billiger“ ist mit der Messung nicht belegt; die DoD verlangt keinen Gewinn

- `kategorie`: INFO
- `quelle`: Plan §1 (Ziel: *vier Ansatzpunkte umsetzen*), §2 Punkt 2, §5; `SPEC-210` Kopf (*Es macht den Lauf billiger*)
- `pfad`: Plan · `**Ergebnis: kein Gewinn an Laufzeit für Satz A und B**`
- `befund`: §2 Punkt 2 verlangt (a) bis (d) *umgesetzt oder je Punkt mit Messung und Grund als nicht tragend benannt*, §5 verlangt bei fehlendem Gewinn die Zahl in §7: ein Gewinn ist keine Bedingung der DoD, und der Plan sagt es offen. Meine Läufe bestätigen die Richtung und die Größe (siehe *Messung des Gewinns*): Satz A kostet das Werkzeug rund ein Viertel mehr Wanduhr (30,2 bis 30,5 s gegen 23,7 bis 24,4 s). Die Messung des Implementers ist dünn (ein Lauf je Zahl, Streuung 20,8 bis 35,1 s unter fremder Last), ihr Ergebnis trägt aber; meine drei Läufe je Seite streuen unter 1 s. Der feste Teil des Aufrufs (rund 20 s, die Nullprobe nach Nr. 5 a mit zwei Bauen und einem Lauf des ganzen Pakets je Art) übersteigt, was die Mutationen sparen (Mutation mit Detektor 1,9 s gegen 16,4 s). Nicht gemessen und damit offen: der kalte `GOCACHE` je Reviewer (§1 Befund 3), Listen mit vielen Mutationen je Paket, viele Pakete. Der Kopf des Vertrags („Es macht den Lauf billiger“) sagt mehr zu, als die Messung trägt; er gehört dem Architect.
- `verifizierbar`: nein (Messung, kein Gate)
- `klasse`: Zweck nicht belegt (Planner, Architect, User)

#### I-2 (INFO): Zu (c) — GNU-Formen im Treiber

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.1 (*Host braucht nur Docker und GNU make*), `harness/conventions.md` (Host: `bash`, `git`, `docker`)
- `pfad`: `tools/harness/mutationslauf.sh` · `du -sk --apparent-size --exclude=.git` und `tar -C "$wurzel" --exclude=./.git --transform`
- `befund`: `tar --exclude` und `--transform` sind GNU-Formen; `tools/harness/zeichen.sh` (`tar -c --null --no-recursion -T -`) setzt GNU `tar` schon voraus, `tar` ist also keine neue Abhängigkeit. `du --apparent-size` ist neu und GNU-spezifisch (BusyBox und BSD kennen es nicht). Der Vertrag (Kopf) nennt `bash`, `tar`, `docker`, der Treiber ist `sh`. Die Entscheidung, ob GNU gefordert wird, gehört dem Architect (Plan §6, offen).
- `verifizierbar`: nein
- `klasse`: Host-Abhängigkeit über Docker und make hinaus

#### I-3 (INFO): Zu (d) — Selbstanzeige zur Betriebsregel

- `kategorie`: INFO
- `quelle`: Betriebsregel *auf Repo-Dateien nur `Edit` und `Write`*; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`
- `pfad`: Auftrag des Reviews · Selbstanzeige des Implementers (kein Text im Plan)
- `befund`: Der Implementer legte `go.mod`, `port.go` und den Treiber anfangs per Heredoc, `sed` und `touch` an. Der Zustand der Dateien im Diff ist unauffällig: `gofmt -l` leer, `go vet` ohne Befund, `LC_ALL=C grep -P` auf BOM, NBSP, Zeilentrenner und Nullbreitenzeichen über `tools/mutationslauf`, `tools/harness/mutationslauf.sh` und den Plan leer, keine Escape-Folge aus Backslash, `u` und vier Hex-Ziffern im Code. Nur vermerkt, kein Finding.
- `verifizierbar`: nein
- `klasse`: Betriebsregel für Repo-Dateien verletzt

#### I-4 (INFO): Zu (e) — die offenen Auslegungen aus §6, eingeordnet (nicht entschieden)

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.12, §3.13; Plan §6
- `pfad`: Plan · `**Tabellenzeile Z4 gegen SPEC-210 Nr. 5 a:**`
- `befund`: (1) *Z4, Präfix-Fall:* Mit `$` im Muster trifft `TestA` genau einen Test; der Vertrag verlangt dann keinen Fehler. Die Lesart des Implementers (`rot-paket`, wenn nur `TestAB` rot wird) folgt dem Vertrag; die Tabelle ist zu berichtigen (Architect). Den fehlenden Anker fängt der Test: ohne `$` im `-test.run` wird `TestNullprobe` rot (A1), ohne `$` im `-test.list` `TestKaltSetztCache` (A2). (2) *Befundtexte:* die Gründe von `Liste nicht lesbar: <Grund>` und `<Feld> ungültig` sind Auslegungen; ihr Wortlaut ist in `TestListeBefunde` fixiert (L-4 und M-7 berühren sie). (3) *Testnaht `<Ablage>`/`<Image>`:* zwei wahlfreie Argumente des Programms, nicht des Treibers; der Treiber nimmt genau ein Argument (SPEC-210 Nr. 1 a); kein Befund. (4) *`last=<l>/<k>`:* im Lauf des Reviewers stand `last=4.19/4` bei einem Minutenwert des Hosts von 4,17 und 20 CPUs: `last` ist die Last des Hosts, `k` die `--cpus` des Containers; die Zahlen sind nicht vergleichbar, wie der Plan sagt. Die Adresse bleibt der Architect (Hinweis im Vertrag).
- `verifizierbar`: nein
- `klasse`: Auslegung ohne Ort im Vertrag (Architect)

#### I-5 (INFO): Zu (f) — Größenüberschreitung 3129 gegen Marke 3000

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.14; `SPEC-128` Nr. 5 d
- `pfad`: Plan §4 · `ueberschreitung: 3200 | Stand beim Handoff 3129 Zeilen`
- `befund`: Nachgemessen: `git diff --numstat 9c9b406..b0eee68` über `tools/mutationslauf` und `tools/harness/mutationslauf.sh` gibt 3129 Zeilen (1245 Programm und Treiber, 1884 Tests, 0 Löschungen). Der Eintrag `ueberschreitung: 3200` mit Grund deckt die Messung nach `SPEC-128` Nr. 5 d (N ≥ Messung); der Weg ist der des Gates. Zwei Hinweise: (1) Der Plan nennt in §4 für *Messung nach dem Code über der Marke* die Rückführung nach `next`; der Implementer nahm sie nicht vor und gab die Entscheidung an den Planner (§6, offen) — keine Lockerung, aber ein Abweichen vom Wortlaut des Plans. (2) Der Spielraum des Eintrags beträgt 71 Zeilen; schon der Fänger zu M-1 (Fixture mit `tests`, `go.sum`, verschachteltem Modul, Symlink) und zu M-4 (vier Ergebnisse × Diagnose) dürfte ihn übersteigen; der Eintrag braucht dann einen neuen Wert mit eigenem Grund (SPEC-128 Nr. 5 d). Das Gate selbst habe ich nicht gefahren.
- `verifizierbar`: ja (`make groessenmarke`)
- `klasse`: Größenmarke überschritten

#### I-6 (INFO): Parallele Aufrufe ohne `flock` — Randform fehlt im Raster

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.16; `SPEC-210` Nr. 10 c
- `pfad`: `tools/harness/mutationslauf.sh` · `--filter status=exited --filter status=created --filter status=dead`
- `befund`: Der Aufruf entfernt jeden Container mit dem Label, der nicht läuft, einschließlich `created`. Ein zweiter Aufruf ohne `flock`, der zwischen `docker create` und `docker start` ist (die Zeit des `docker cp` des Baums), verlöre seinen Container. Der Vertrag legt den `flock` in die Hand des Aufrufers (Nr. 3 e); die Zelle *Aufräumen × Wiederholung* nennt parallele Aufrufe nicht.
- `verifizierbar`: nein
- `klasse`: Randform paralleler Aufruf nicht im Raster

#### I-7 (INFO): `SPEC-210` Zeile *Stand* ist überholt

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.7 (Zustandsfelder)
- `pfad`: `harness/vertraege/SPEC-210-mutationslauf.md` · `**Stand:** der Vertrag steht vor dem Code; nichts davon ist geliefert.`
- `befund`: Der Code liegt im Zweig; die Zeile sagt weiter *nichts davon ist geliefert*. Die Datei gehört dem Architect und ist nicht Teil des Diffs; vermerkt für den Nachzug bei der Closure.
- `verifizierbar`: nein
- `klasse`: Zustandsfeld überholt (Architect)

---

## Aufgelöste Zeilen und Mutationen

Messender: der Reviewer; Image `odata-recorder:race`, Container `--memory=2g --memory-swap=2g --cpus=4 --network none`, Wegwerfkopie. „rot“ nennt den ersten roten Test. Die Namen der Mutationen stehen in den Findings; „grün“ heißt: alle schnellen und alle echten Tests des Moduls grün.

| Mutation (Zeile) | Ergebnis |
|---|---|
| Z1/Z4 Kopie: K1 verschachtelte Module kopieren · K2 Symlinks weglassen · K3 `tests` weglassen · K4 `go.sum` weglassen · K6b Dateimodus | grün (M-1) |
| Z1 Kopie: K5 fehlendes `tests` nicht tolerieren | rot `TestErgebniszeile` |
| Z5 Ausgabe: L1 Diagnose `rot-paket` gestrichen | rot `TestDiagnose` |
| Z5: L2 `flatternd`, L3 `abbruch`, L4 `uebersetzungsfehler` aus der Diagnose gestrichen · L8 Text des Arms ohne Detektor | grün (M-4) |
| Z2: L5 Stufe 2 auch bei rot (`andere`) · L6 auch bei `rot` (Rate) · L7 auch bei `flatternd` | rot `TestErgebniszeile` |
| Z2: L10 Binary mit und ohne Detektor an einem Pfad | rot `TestGleichwertigWieEinzellauf` (`rennen: Werkzeug rot, erwartet flatternd`) |
| Z2/Z4: E4 `CGO_ENABLED=0` für den Detektor-Bau · A5 `-test.timeout=1h` | rot `TestGleichwertigWieEinzellauf` (Exit 2; `endlos: Werkzeug abbruch, erwartet rot`) |
| Z4: A1 `$` im `-test.run` fehlt | rot `TestNullprobe` |
| Z4: A2 `$` im `-test.list` fehlt | rot `TestKaltSetztCache` |
| Z4: A4 `-test.count=1` nicht übergeben | grün (äquivalent: der Vorgabewert des Test-Binary ist 1; kein Finding) |
| Z7: E8 Frist des Baus 1 Sekunde | rot `TestKaltSetztCache` (`Ausgangsstand rot`) |
| Z12: E6 Größe zählt 1 je Datei · E10 zählt alle Einträge · E7 `go clean -modcache` | grün (M-2) |
| Z12: L11 Race-Binary je Mutation nicht entfernt · L9 Bau für Arm 0 · L12 Race-Binary der Nullprobe nicht entfernt | grün (L-3) |
| Z8: L13 Signal in der Nullprobe (Bedingung `ctx.Err() == nil` entfällt) | grün (M-5) |
| Z8: L14 `go clean -cache` mit Kontext des Abbruchs · L16 Nullprobe mit Detektor auch für Mutationen der Klasse `andere` | grün (L16 äquivalent in der Wirkung; L14 wirkt nur, wenn der Abbruch in das Leeren fällt; kein Finding) |
| Z10/Z12: T1 `--pids-limit` · T2 `--cpus=8` · T3 `--label` · T4 `-e MUTATIONSLAUF_GRENZE` · T5 `-e MUTATIONSLAUF_KALT` · T11 Umgebung des Baus | grün (M-3) |
| Z12: E1 `-trimpath` · E2 `-buildvcs=false` · E3 `-p 2` · M6 `GOPROXY=off` · M7 `GOTOOLCHAIN=local` · M8 `GOFLAGS=-mod=readonly` | grün (M-3) |
| Z12: E9 Grenze der Ausgabe 64 KiB aufgehoben · M10 `ladeLast` liest nicht den ersten Wert | grün (kein Finding: die Grenze der Ausgabe ist keine Zusage des Vertrags; `ladeLast` liest `/proc/loadavg`, den Rand des Programms) |
| Z12/Z13: M1 `strconv.Itoa(n) != v` · M2 `KALT` leer ungültig · M3 `go.sum` im Abgleich · M5 Zahl der Argumente · M9 Grenze 0 | rot (`TestUmgebung`, `TestRunBefundeStartenNichts`, `TestImageAbgleich`, `TestRunEingabe`) |
| Z13: S1 bis S6 Ränder der Bereiche · S10 · S13 · S16 · S17 · S22 · S23 · S27 · S28 · S30 · S31 · S32 · S33 · S35 · S36 | rot (`TestGleichwertigWieEinzellauf` für S1, S3; `TestListeBefunde`, `TestListeForm`, `TestListeZusatztabelle`, `TestRateNurFuerNebenlaeufig`, `TestSpeicherUndGleichzeitigkeit`, `TestErgebniszeile` sonst) |
| Z13: S7 `.` · S8 `_` · S9 `-` im Namen · S11 `.` im Test · S12 Präfix `Test` | grün (M-6) |
| Z13: S14 NUL im Pfad · S15 Backslash · S18 leeres Segment · S19 führender `/` · S20 Endung `.go` auf `o` gekürzt | grün (S19 äquivalent: das leere erste Segment fängt dasselbe; S14, S15, S18 ohne Wirkung auf das Ergebnis; S20: nur Endungen auf `o` ohne `.go` ungefangen, Rest der Zeichenklasse aus M-6, kein eigenes Finding) |
| Z13: S24, S25, S26 `null` als Liste, Wurzel, Eintrag | grün (L-4) |
| Z13: S34 Länge in Bytes statt Zeichen | grün (äquivalent: die Zeichenmenge verbietet alles außer ASCII) |
| übersetzen nicht: K6, M4, E5, L15, S21, S29 | nicht gewertet |

## Raten

Messender: der Reviewer. Test-Binary aus `go test -c` im Image `odata-recorder:race` (Container 2g, 4 CPUs, ohne Netz, unter `flock`); je Mutation 20 Prozesse nacheinander mit `-test.count=1` und ein Prozess mit `-test.count=20`. Last des Hosts (Minutenwert) zu Beginn der Reihe 8,5, im Verlauf 7 bis 37; 20 CPUs. Die Zahl bei `-test.count=20` ist die Zahl der Zeilen `--- FAIL` (40 Zeilen, zwei je Wiederholung: Teiltest und Elterntest), keine Zahl der Wiederholungen; die Rate ist an den 20 Prozessen abzulesen.

| Mutation (Zeile) | Test | 20 Prozesse, rot | 1 Prozess `-count=20` |
|---|---|---|---|
| ohne Mutation | `TestRateNurFuerNebenlaeufig` | 0 von 20 | grün |
| Z6 Deckel 5 | dasselbe | 20 von 20 | rot, 40 Zeilen `FAIL` |
| Z6 Deckel 3 | dasselbe | 20 von 20 | rot, 40 Zeilen |
| Z6 Arm mit 0 startet doch einen Prozess (`for range max(n, 1)`) | dasselbe | 20 von 20 | rot, 40 Zeilen |
| ohne Mutation | `TestSpeicherUndGleichzeitigkeit`, `TestKopieAbbruch`, `TestKopieNichtEntfernt` | 0 von 20 | grün |
| Z12 zwei Baue zugleich (Bau in einer Goroutine neben dem Race-Bau, ein Element, das weder Tabelle noch Implementer nannten) | dieselben | 20 von 20 | rot, 40 Zeilen |
| Z8 Kopie bleibt (kein `Entfernen` je Mutation) | dieselben | 20 von 20 | rot, 40 Zeilen |
| ohne Mutation | `TestAbbruchRaeumtAuf/SIGTERM` | 0 von 20 | grün |
| Z8 `lauf` bleibt (kein Entfernen der Ablage am Ende) | dasselbe | 20 von 20 | rot, 40 Zeilen |
| Z8 nur der Prozess, nicht die Gruppe endet (`Kill(pid)` statt `Kill(-pid)`) | dasselbe | 20 von 20 | rot, 40 Zeilen |
| ohne Mutation, `-race` | `TestRateNurFuerNebenlaeufig`, `TestErgebniszeile` | 0 von 20 | grün |
| Z6 Zähler ohne Sperre (`mu.Lock` und `mu.Unlock` um den Zähler entfernt), `-race`, `CGO_ENABLED=1` | dieselben | 20 von 20 mit Detektor | 9 Zeilen `FAIL` (der Detektor meldet dasselbe Rennen je Prozess einmal; die Rate zählt der Prozess, nicht `-count`) |

Die Zeilen Z6, Z8 und Z12 sind für die gefahrenen Mutationen in jedem Lauf rot. Nicht gemessen: Z12 *Start neben einem Bau*, Z8 am Treiber (kein Rate-Thema, Ersatz-Skript für `docker`), das Verhalten des Fängers `Baue` bei einer Last über der des Laufs.

## Messung des Gewinns

Messender: der Reviewer. Satz A: die fünf Mutationen des Implementers in `src/hexagon/services/record/anhaengen.go` (eine `nebenläufig` mit 50 und 50 Prozessen, vier `andere`, davon eine äquivalente); Host mit 20 CPUs; ein Lauf nach dem anderen unter `flock`; der Treiber aus dem Worktree, Image `odata-recorder:race`.

| | Werkzeug (Treiber, ganzer Aufruf) | Bisheriges Vorgehen (Nachbau des Implementers, `baseline.sh`) |
|---|---|---|
| Wanduhr, Lauf 1 / 2 / 3 | 30,5 / 30,2 / 30,2 s | 24,4 / 23,7 / 24,2 s |
| im Container | `zusammenfassung dauer` 24,5 / 24,3 / 24,2 s | 19,5 / 19,0 / 19,1 s |
| Mutation mit Detektor (100 und 101 Prozesse) | 1,9 / 2,1 / 1,9 s | 16,6 / 16,4 / 16,4 s |
| je Mutation `andere` | 0,4 bis 0,6 s | 0,6 bis 0,8 s |
| Summe der `dauer` der fünf Mutationen | 4,1 / 4,3 / 3,9 s | — |
| Last des Hosts, Minutenwert zu Beginn | 4,2 / 4,2 / 4,2 | 5,0 / 5,9 / 7,6 |
| Last des Hosts, Fünfzehnminutenwert | 6,4 bis 6,6 | 14,4 bis 14,9 |
| Ergebnis je Mutation | 4 rot, 1 grün; 50 von 50 und 50 von 50 | 4 rot, 1 grün; 50 von 50 und 50 von 50 |

Folgerung: Das Werkzeug ist auf Satz A reproduzierbar rund 6 s (ein Viertel) langsamer. Es verschiebt die Kosten der Mutation mit Detektor (16,4 s im Nachbau) in einen festen Teil des Aufrufs von rund 20 s (Aufruf minus Summe der Mutationen; die Nullprobe nach Nr. 5 a, der Bau des Programms und die Kopie des Baums, nicht einzeln gemessen). Je Mutation der Klasse `andere` ist das Werkzeug in meinen Läufen nicht teurer (0,4 bis 0,6 s gegen 0,6 bis 0,8 s). Die Fünfzehnminutenlast der beiden Reihen unterscheidet sich (6,5 gegen 14,5); die Reihen sind darum nicht besser vergleichbar als auf einige Sekunden. Die Ergebnisse je Mutation gleichen sich (Z2). Nicht beantwortet: der kalte `GOCACHE` je Reviewer, Listen mit vielen Mutationen je Paket.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tools/mutationslauf/` Formatierung und Vet | geprüft, ohne Befund: `gofmt -l` leer, `go vet ./...` ohne Befund, Image `odata-recorder:race`, Go 1.27.1 |
| Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `GATE_CHECKS`, Sensors-Tabelle, `.golangci.yml`) | geprüft, ohne Befund: im Diff nicht berührt |
| `AGENTS.md` §3.1 (Docker-only, netzlos, ohne Mount) | geprüft, ohne Befund: `--network none`, kein Mount, `docker cp`; der Treiber schreibt nichts auf den Host (Lesen und `TestTreiberSchreibtNichts`); Ende-zu-Ende dreimal gelaufen, danach kein Container mit dem Label |
| `AGENTS.md` §3.2 (Suppression) | geprüft, ohne Befund: keine `//nolint`-Direktive im Diff |
| `AGENTS.md` §3.6 (Gate nicht gelockert) | geprüft, ohne Befund: kein Gate geändert; Überschreitung der Marke über den Eintrag des Gates (I-5) |
| `AGENTS.md` §3.7 (Kommentare) | geprüft, ohne Befund: keine Konjunktive über verworfene Alternativen, kein abwesender Text, kein abgebrochener Satz in Programm, Treiber und Tests; L-1 betrifft die Zusage, nicht die Klasse |
| `AGENTS.md` §3.18 (Rohbytes) | geprüft, ohne Befund: keine Escape-Folge und kein BOM, NBSP, Zeilentrenner oder Nullbreitenzeichen in Programm, Treiber und Plan (`LC_ALL=C grep -P`) |
| Raster (Plan §6) gegen SPEC-210 Nr. 11 | geprüft: jede Zelle trägt eine Nummer oder *gilt nicht* mit Grund; Lücken in M-5, M-7, I-6 |
| `harness/README.md` und `harness/vertraege/README.md` | geprüft, ohne Befund: `harness/README.md` unberührt (der Plan verschiebt es nach `slice-test-determinism`); die Vertragszeile für SPEC-210 steht |
| Nebenläufigkeit im Programm (`arm`, Deckel von vier, Sperre um den Zähler) | geprüft, ohne Befund: Z6, Z8, Z12 rot in jedem gefahrenen Lauf, siehe *Raten* |
| Ende-zu-Ende des Treibers | geprüft, ohne Befund: dreimal Exit 0, Ergebniszeilen im Format von Nr. 1 b, `last=…/4`, Zusammenfassung stimmt mit den Zeilen |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 8 |
| LOW | 4 |
| INFO | 7 |

**Finding-Klassen dieses Laufs:** Zusage ohne Fänger — Menge der Kopie · Zusage nur am Fake gefangen — echter Adapter und Optionen ungepinnt · Fänger führt nur ein Element der Menge · Randform-Zelle ohne Fänger (Abbruch in einer Phase) · Fänger führt nur einen Teil der Zeichenmenge · Paar von Verstößen im Code entschieden · Messung nicht wiederholbar · Kopf sagt mehr zu, als der Test fängt · Zweck nicht belegt

## Verdikt

**Merge-blockierend:** nein für HIGH (keines); die acht MEDIUM-Findings sind Zusagen ohne Fänger, Randformen ohne Ort und eine nicht wiederholbare Messung — nach `AGENTS.md` §3.15 (*Eine Zusage ohne Fänger ist ein MEDIUM*) vor der Verifikation zu schließen. M-8 hat keinen Gate-Fänger und braucht entweder die Liste und das Skript im Repo oder die Aussage des Planners, dass die Messung kein Beleg des Slice ist. Die Frage (a) ist keine Bedingung der DoD.

**Übergabe:** Findings gehen an den Implementer (M-1 bis M-7, L-1 bis L-4; Fixes gegen die Klasse nach `AGENTS.md` §3.17, nicht gegen die genannten Mutationen); M-7, I-1, I-2, I-4, I-6, I-7 an Architect, Planner oder User (Auslegungen, GNU, Vertragskopf, Zweck); I-5 an den Planner (Eintrag der Größe). Die Finding-Klassen gehen zusätzlich in die Slice-Closure §7. Dieser Report ist ein Lauf-Beleg und ersetzt keine Verifikation — DoD und Spec prüft der Verifier.
