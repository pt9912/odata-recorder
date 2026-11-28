# Verifikation slice-batch-parser (Leser für `$batch`) — 2026-11-28

**Rolle:** Verifier (genau ein Lauf). **Stand:** Zweig `slice-batch-parser`, HEAD `46c12249`. **Eingang:** Plan `slice-batch-parser` (DoD, §3 mit Handoff und Fix-Runde, §6), [Review](2026-11-28-slice-batch-parser-review.md) (Gegenstand `866d90ec`), Fix-Runde `380ac0fd`.
**Bezug:** `LH-FA-16`, `LH-FA-16.a`; `SPEC-182`, `SPEC-183`, `SPEC-184` (allein der Leser), `SPEC-126` Nr. 2, `SPEC-109` Nr. 5; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md). **Messender:** der Verifier.

## Urteil

**Bestätigt mit Auflagen.** DoD 1 bis 3 und der Gate-Nachweis sind erfüllt, Produktcode und Vertrag stimmen überein, die Fix-Runde deckt die Klassen F-1 bis F-4. Kein HIGH, kein MEDIUM, ein LOW, drei INFO. Keine Auflage verlangt eine Code-Änderung. Der Slice ist nicht Teil der Erstabnahme; `LH-FA-16` ist mit ihm nicht erfüllt (Ausführung und Antwort liegen bei `slice-batch-changesets`).

## Auflagen (an den Planner, vor `done/`)

1. §6 berichtigen (V-1) und jedem Risiko seinen Ausgang geben; mehrere stehen auf *offen*.
2. §7 zum Binary genau fassen (V-2).
3. Closure-Pflichten: Häkchen der DoD (alle noch leer), Closure-Notiz mit Lerneintrag, Register, drei Paarungen.
4. Integration: `main` ist 10 Commits voraus (`7dc8b3e7`; kein Pfad unter `src/`, `git merge-tree` ohne Konflikt). Der gemergte Stand braucht seinen eigenen Gate-Nachweis; `harness/gegenproben.stempel` gehört zum Integrationsschritt (Plan, Betriebsregeln) und ist hier nicht geprüft.

## Findings

| ID | Kategorie | Befund | Quelle |
|---|---|---|---|
| V-1 | LOW | §6 sagt zum Risiko *Bestehende Tests* weiter, „die acht Dateien ändern allein Listen und Zahlen“. Seit der Fix-Runde trägt `suche_test.go` eine neue Zählfunktion, `nenntEmpfaenger` und drei neue Zusicherungen. §3 ist nachgezogen, §6 nicht. Kein Fänger ist gelockert: die Zusicherung „13“ steht weiter. | `AGENTS.md` §3.9 |
| V-2 | INFO | Die DoD verlangt ohne neuen Code den Satz in §7, „dass das Binary unverändert blieb“. Wörtlich trifft er nicht zu: `abbildung.go` hat zwei Zeilen mehr und `model` neue Werte, beide Pakete liegen im Binary. Zutreffend ist: kein neuer Code, Schlüssel oder Kommando, kein Verhalten geändert, kein Produktcode importiert `batchcodec`. | DoD, Zeile zum Katalog |
| V-3 | INFO | Zeile 28 nennt „unverwechselbare Texte in jeder Verstoß-Fixture“; die Fixture zu `zu_viele` trägt allein `geheim[:1]`. Ohne Wirkung auf den Fang: `wantFehler` vergleicht Ursache und Ort genau, `Error()` entsteht nur aus ihnen. | `AGENTS.md` §3.11 |
| V-4 | INFO | `SPEC-126` Nr. 2 nennt „ein Kontextfehler von `src`“; `TestBatchStromfehler` führt allein `context.Canceled`, nicht `context.DeadlineExceeded` (M4 des Reviews, dort grün und ohne Befund; hier nicht nachgefahren). Der Code vergleicht allein `io.EOF`. | `SPEC-182` Nr. 2 |

## DoD-Konformität

| DoD | Ergebnis | Beleg |
|---|---|---|
| 1 Lesen der Hülle, leerer Batch, Zerlegung in Reads | erfüllt | Produktcode gegen `SPEC-182` Nr. 1 bis 3, 5, 6 und `SPEC-183` Nr. 1 bis 7 gelesen; Tests der Zeilen 3 bis 27, 34, 35; Gate-Stempel |
| 2 Port, Domain-Werte, `ContentIDGueltig`, `BatchFehler`, zwei Sentinels, zwei Zeilen | erfüllt | Signaturen gleich `SPEC-182`; 16 Ursachen mit Klasse und Code gleich der Tabelle in Nr. 4; Abbildung `E6007` und `E6001` wie `SPEC-109` Nr. 5 |
| 3 Raster, Fänger je Zelle, Parser-Contract-Test | erfüllt | 14 Zellen und 7 Paare der Zusatztabelle haben Ort und Fänger; 36 Zeilen, jeder genannte Test steht im Baum (Zeile 36 als Negativ mit Grund, Folge, Maß); `TestBatchContract` |
| 4 `make gates` grün | erfüllt | `gates-passed.head` gleich `46c12249`, `gates-passed.diffsha` gleich dem Hash des Arbeitsbaums (`6fb7a182…`, 12:57:16). Nicht neu gefahren; die Ausgabe des Laufs hat der Verifier nicht gesehen |
| Review | erfüllt | Report liegt vor. Die Fix-Runde ist nicht erneut reviewt; ihre Deckung steht unten |
| Doku-Update | erfüllt | `SPEC-182` bis `SPEC-184`, `SPEC-109` Nr. 5 und die Sicht (`ARC-004`, `ARC-011`) tragen den Vertrag |
| Katalog und Handbuch | erfüllt, Satz in §7 offen | `ODR-E6001` und `ODR-E6007` bestanden schon; `katalog-check` über den Gate-Stempel; V-2 |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen | Sache der Closure |

## Fix-Runde gegen die Klassen des Reviews

Ein Lauf von `tools/harness/mutationslauf.sh` unter `flock` (ein Container, ohne Netz, 4g, 4 CPU; Last 6,00 beim Start, 14 GiB verfügbar): **12 von 12 rot**, Nullprobe grün, kein Übersetzungsfehler, kein Abbruch, je Mutation 0,4 bis 8,2 s. Keine Mutation ist eine, die Review oder §3 nennen. Bei `rot` gibt das Werkzeug keine Ausgabe; dass der Grund der benannte ist, folgt aus dem Zuschnitt der Mutation, nicht aus einer gelesenen Fehlerzeile.

| Befund, Menge | Eigene Mutation | Roter Test |
|---|---|---|
| F-1: Grenze mal Ende des Stroms | Fehler von `src` geht der Größe vor; `io.EOF` mit Daten geht ab der Länge des Lesepuffers der Größe vor | `TestBatchGroesse` (beide) |
| F-2: Prüfstelle mal Art des Endes | nach dem letzten Read wird allein die Frist übergangen; nach dem letzten Read allein bei gültigem Inhalt geprüft; vor weiteren Reads der Fehler umhüllt | `TestBatchAbbruch` (alle drei) |
| F-3: Paare am Changeset | alle Köpfe der Teile eines Changeset vor dem ersten Request; offenes Changeset mit zu vielen Teilen ist `zu_viele`; geschachtelt `verschachtelt` vor den Kopfzeilen | `TestBatchRang` (alle drei) |
| F-4: Typen mit `Unwrap` und `Is` auf einem Zeiger | `Unwrap` im Verzeichnis `driven` (`sqlite`) ohne Nennung des Empfängers; neues `Is` auf `*DatenFehler` | `TestTypenMitUnwrapAufZeiger` (beide) |

Stichproben außerhalb der Befunde: ein Read mit `0, nil` als Ende (Zeile 6) rot in `TestBatchZerlegung`; ein Puffer als Variable des Pakets (Zeile 32, `SPEC-182` Nr. 6) rot in `TestBatchGleichzeitig`, mit Detektor 20 von 20, ohne 20 von 20 Prozessen (Last 5,52, vom Werkzeug gemeldet). Die übrigen Zeilen der Tabelle sind gegen den Vertrag gelesen, nicht einzeln mutiert.

Zum Maß von F-4: die Zahlen (13 `Unwrap`, davon 9 mit Empfänger; 3 `Is`, davon 3) stimmen mit der Suche im Quelltext überein.

## Plan-vs-Code

- Die Commits `866d90ec` und `380ac0fd` ändern allein Pfade der Tabelle in §3 und den Plan. `spec/spezifikation.md` und `spec/architecture.md` (Zeile *Architect*) stehen schon vor der Basis. Unverändert, wie zugesagt: `.a-check.yml`, `tools/`, `Dockerfile`, `harness/`, `src/bootstrap/`, `CodecSatz`.
- Bestehende Tests (`866d90ec`, acht Dateien): allein Listen und Zahlen; `28*28` zu `28*30` und `4*(28+4+1)` zu `4*(30+4+1)` bleiben Gleichheiten. Zur Fix-Runde siehe V-1.
- Importe des Lesers: Domain, Ports, Standardbibliothek; weder `mime/multipart` noch `net/http` noch `net/textproto`.
- **Größenmarke** (selbst gemessen, `git diff --numstat 7fb0f523..46c12249` über die Pfade des Blocks): 3417 von 4600, davon Tests 2776, Produktcode 641; gleich dem Messpunkt in §3. Produktcode über der Erwartung von 500, keine Überschreitung der Marke.
- Sendungen: `slice-batch-validator` und `slice-batch-changesets` führen die Übernahmen mit der Kennung des Gebers.
- Rohbytes: die Go-Dateien in `batchcodec` und `model/batch*` tragen kein Steuerzeichen außer Tab und `LF`.

**Entscheidungen:** [ADR-0001](../plan/adr/0001-implementierungssprache-go.md): Go, keine neue Abhängigkeit; ohne Befund. Der eigene Leser ist die Entscheidung in `SPEC-182` und eingehalten.

## Nicht gefahren

`make gates` (Stempel passt), `make gegenproben` (Auftrag), `make test`, `make test-race`, `make lint`, `make groessenmarke` einzeln. Die Raten des Implementers zu Zeile 32 (andere Mutation: Feld des Lesers) und die Zuteilung zu Zeile 33 sind nicht nachgemessen. Neben dem Container des Mutationslaufs lief allein nacheinander der von `make docs-check` für diesen Report.
