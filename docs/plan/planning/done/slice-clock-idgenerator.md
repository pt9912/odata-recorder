# Slice slice-clock-idgenerator: Zeit und Kennungen hinter Querschnitts-Ports

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-QA-02`, `LH-FA-13`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-QA-02.a` (Zeit und IDs aus `Clock`/`IDGenerator`) · `SPEC-114` (Vertrag, mit diesem Plan geschrieben) · `SPEC-105` (Zeit und Speicherung) · `ARC-019` · `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Querschnitts-Ports `Clock` und `IDGenerator`, ihr Driven Adapter `src/adapters/driven/clock` (`Systemuhr`, `ZufallsIDs`) und die Test-Implementierungen `FestUhr` und `FestIDs` samt Vertragstests in `src/hexagon/ports/support/supporttest/` liefern Zeit und Kennungen (Session, Exchange, Korrelation, Workspace, Snapshot) nach `SPEC-114`, und die Tests der Services ersetzen sie durch feste Werte.

**Übergibt** (jeder Nehmer führt die Sendung mit der Kennung dieses Slice in seinem §1; `AGENTS.md` §3.13): `slice-http-adapter` (Korrelations-ID), `slice-record-service` (Exchange-Kennung, Zeit), `slice-session-wartung` (Session-Kennung, Zeit, Form einer vom Aufrufer gelieferten `id`), `slice-workspace-use-case` (Workspace-Kennung, Zeit), `slice-baseline-import` (Snapshot-Kennung, Zeit), `slice-bootstrap-m2` (Verdrahtung, eine Instanz je Prozess).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `AuditLog` und `Telemetry` als Querschnitts-Ports — Folgewelle (Welle §6); ohne Audit-Endpunkte und Spans gibt es keinen Aufrufer.
- Kennungen, die der Client vergibt (Replay-Session-ID, `import_id`, mitgegebene Workspace- oder Session-`id`) — `LH-QA-02.a` legt sie beim Client, der Port liefert nur serverseitige (`SPEC-114` Nr. 9); ihre Form legt der Use Case fest, der sie annimmt (`slice-workspace-use-case` §6, `slice-session-wartung`).
- Aufrufe der Ports in den Services und die Verdrahtung im Composition Root — die Slices der Zeile *Übergibt*; hier entstehen Ports, Adapter und Fakes, kein Aufrufer.
- Dauer messen (`duration_ms`) und Timer — `Clock` kennt nur `Now()` (`SPEC-114` Nr. 4); Messung gehört zu `Telemetry` (Folgewelle, Welle §6).

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

- [x] Ports `Clock` und `IDGenerator` unter `src/hexagon/ports/support/` und der Adapter `Systemuhr`/`ZufallsIDs`: UTC-Zeit, nicht fallend, Kennungen der Form `SPEC-114` Nr. 6.
- [x] `FestUhr`, `FestIDs` und die Vertragstests `UhrVertrag`, `IDVertrag` in `supporttest`, die gegen echte und feste Implementierung laufen; Test der Eindeutigkeit unter gleichzeitigem Aufruf.
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/support/` | neu | `Clock` (`Now() time.Time`), `IDGenerator` (`SessionID`, `ExchangeID`, `KorrelationsID`, `WorkspaceID`, `SnapshotID`, je `func() string`), ohne Argument, Kontext und Fehler (`SPEC-114`) |
| `src/hexagon/ports/support/supporttest/` | neu | `FestUhr` (`NewFestUhr`, `Now`, `Advance`), `FestIDs` (`NewFestIDs`), die Vertragsfunktionen `UhrVertrag(t, uhr)` und `IDVertrag(t, gen)` nach dem Muster von `outboundtest`; die Vertragsfunktionen zählen die Methoden der Ports mit `reflect` auf |
| `src/adapters/driven/clock/` | neu | `Systemuhr` (`NewSystemuhr`, unexportiert `neuMit(quelle)` für die Tests der Rücksprünge) und `ZufallsIDs` (`NewZufallsIDs`, 16 Byte aus `crypto/rand`, hex); je eine Zusicherung `var _ support.Clock = …` |
| `spec/spezifikation.md` | update | **vor dem Code geschrieben:** `SPEC-114` samt Historie-Zeile; `LH-QA-02.a` verweist darauf |
| Tests der Pakete | neu | Tabellen- und Vertragstests nach der Tabelle unten |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16); beide gehören zum Vertrag `SPEC-114`, der vor dem ersten Code-Commit in der Spezifikation steht (`AGENTS.md` §3.12). Die Fänger sind Namen geplanter Tests; der Implementer legt sie unter diesem Namen an oder nennt den Namen, den er stattdessen vergibt, im Bericht. Der Race-Detektor gehört nicht zum Gate (`pruefung test` fährt `go test` ohne `-race`); wo ein Fänger davon abhängt, steht das in der Zeile.

**Zusage · Fänger** (jede Zeile löst der Reviewer mit einer Mutation auf):

| Zusage | Fänger |
|---|---|
| `Now()` liefert UTC (Ort `time.UTC`), `SPEC-114` Nr. 1 | `UhrVertrag` (`TestVertrag`, `TestFestVertrag`) für `Systemuhr` und `FestUhr`: `Location() == time.UTC`; für `FestUhr` auch mit Start in `+02:00`. Mutation: `.Local()` statt `.UTC()`: rot |
| Keine monotone Ablesung, keine Rundung, Nanosekunden bleiben (Nr. 1) | `TestSystemuhrOhneRundung`: Quelle liefert `time.Now()` mit Nanosekunden ≠ 0; Ergebnis hat `t == t.Round(0)` und dieselben Nanosekunden. Mutation: `Truncate(time.Second)`: rot; `UTC()` entfernt: rot (`t == t.Round(0)` und Ort). Ein `Round(0)` trägt die Uhr nicht: `UTC()` entfernt die monotone Ablesung selbst (Dokumentation von `time.Time`), eine Mutation „ohne `Round(0)`“ gibt es daher nicht. Dass der Speicher die Nanosekunden hält, trägt `outboundtest` (`123456789`, `SPEC-105`) — hier nicht doppelt geprüft |
| Nicht fallend; Rücksprung bleibt auf dem Höchstwert; gleicher Wert erlaubt (Nr. 2) | `TestSystemuhrRuecksprung`: Quelle `t, t-1h, t-1h, t+1s, t` ergibt `t, t, t, t+1s, t+1s`. Mutation: Klammer entfernen: rot |
| Dieselbe Zusage bei gleichzeitigem Aufruf (Nr. 2, Einleitung) | `TestSystemuhrNebenlaeufig`: 64 Goroutinen, Quelle fällt mit jedem Aufruf (atomarer Zähler); die Folge jeder einzelnen Goroutine ist nicht fallend. Mutation: Klammer entfernen: **sicher** rot. Die Reihenfolge „Quelle unter der Sperre lesen“ ändert die Monotonie nicht (die Klammer hält sie auch bei Ablesung vor der Sperre) und wird von diesem Test **nicht** gefangen; sie hat einen eigenen Fänger, `TestSystemuhrQuelleSerialisiert`: die Quelle zählt die gleichzeitig laufenden Aufrufe (atomarer Zähler hoch, `time.Sleep(time.Millisecond)`, Höchststand merken, Zähler runter); 8 Goroutinen × 20 `Now`; der Höchststand ist 1. Mutation: `u.quelle()` vor `u.mu.Lock()`: rot (Höchststand > 1); Sperre entfernen: ebenfalls rot. Der Test ist ohne Race-Detektor nicht mathematisch zwingend (eine Planung ohne Überlappung bliebe grün), aber ein falsch-grüner Lauf ist durch die schlafende Quelle bei 8 Goroutinen praktisch ausgeschlossen und ein falsch-roter unmöglich. Das Image hat keinen Race-Detektor (`-race` bricht mit „requires cgo“ ab; Review F3), er wird nicht zugesagt. Gemessen: „Sperre entfernen“ ohne diesen Test ist nur über `TestVertrag` (Teil `Quelle in +02:00`) rot, in 20 Läufen 11-mal; `TestSystemuhrNebenlaeufig` bleibt dabei grün |
| Keine Bereichsprüfung, aber Nr. 2 gilt: Nullzeit und Jahr 10000 gehen unverändert durch; Nullzeit nach gültigem Wert bleibt auf dem Höchstwert; ein Wert hinter Jahr 9999 wird Höchstwert (Nr. 3) | `TestSystemuhrBereichUngeprueft`: frische Uhr mit Quelle `0001-01-01T00:00:00Z` liefert sie; Quelle `t, Null` liefert `t, t`; Quelle `t, 10000-01-01, t` liefert `t, 10000-01-01, 10000-01-01`. Mutation: Ersatzwert für Null: rot. Ein Wert vor dem Jahr 1 an einer frischen Uhr geht ebenfalls unverändert durch (kein Vorgänger, Nr. 3); Mutation: Höchstwert ohne „gesehen“-Marke: rot |
| `Clock` kennt weder Warten noch Timer; Port ohne Argument, Kontext, Fehler (Einleitung, Nr. 4) | `TestPortSignaturen`: `reflect` auf `Clock` und `IDGenerator`: `Now` ist `func() time.Time`, jede Methode von `IDGenerator` ist `func() string`, `Clock` hat genau eine Methode. Mutation: Parameter `ctx` oder eine zweite Methode: rot |
| `NewFestUhr`: Nullzeit und Jahr außerhalb 1 bis 9999 nach Umrechnung sind Panik, die Grenzen selbst gültig (Nr. 5) | `TestFestUhrStart`: gültig `0001-01-01T00:00:00.000000001Z`, `9999-12-31T23:59:59.999999999Z`, `2026-10-20T10:00:00+02:00` (ergibt 08:00 UTC); Panik bei `time.Time{}`, `time.Time{}.In(+02:00)`, `10000-01-01T00:00:00Z`, `9999-12-31T23:00:00-02:00` (Jahr 10000 in UTC), `0001-01-01T01:00:00+02:00` (Jahr 0 in UTC). Mutation: Jahr der Ortszeit prüfen: rot |
| `FestUhr.Now` ändert den Stand nicht; `Advance` rückt vor, `0` erlaubt, negativ und Ergebnis nach Jahr 9999 sind Panik ohne Änderung (Nr. 5) | `TestFestUhrAdvance`: zweimal `Now` gleich; `Advance(1ns)` plus 1 ns; `Advance(0)` unverändert; `Advance(-1)` und `Advance` über `9999-12-31T23:59:59.999999999Z` Panik, danach unveränderter Stand. Mutation: negatives `Advance` zulassen: rot; Stand vor der Panik ändern: rot |
| Gleichzeitige `Advance` summieren sich (Nr. 5) | `TestFestUhrNebenlaeufig`: 64 Goroutinen je 1000 × `Advance(1ns)` ergeben Start + 64000 ns. Mutation: Sperre entfernen: im Review rot gesehen, ohne Race-Detektor nicht mathematisch sicher — **akzeptiertes Negativ**, Grund: das Gate fährt kein `-race` (siehe oben), und ein Datenrennen lässt sich im Test nicht erzwingen |
| Jede Operation von `IDGenerator` liefert `<Präfix>-<32 × 0-9a-f>`, die Präfixe sind `ses`, `exc`, `cor`, `wsp`, `snp` und verschieden (Nr. 6) — **Allaussage über alle Operationen** | `IDVertrag` für `ZufallsIDs` und `FestIDs`: zählt die Methoden von `IDGenerator` mit `reflect` auf, ruft jede, prüft den regulären Ausdruck und das Präfix aus einer Tabelle im Test; die Tabelle und die Methodenmenge müssen sich in beiden Richtungen decken (eine neue Methode ohne Zeile ist rot, eine Zeile ohne Methode ebenso). Mutation: Präfix einer Operation ändern: rot; eine sechste Methode mit falscher Form: rot |
| Kennungen kommen aus `crypto/rand`, ohne Seed, keine feste Folge (Nr. 7) | `TestZufallsIDsUnabhaengig`: zwei Instanzen liefern verschiedene erste Kennungen je Operation. Mutation: `math/rand` mit bei **jedem Aufruf** neu gesetztem festem Seed: rot. `TestZufallsIDsLiestCryptoRand`: tauscht `crypto/rand.Reader` gegen einen Leser, der die Bytes `0x00, 0x01, …` aufsteigend liefert (Wiederherstellung per `t.Cleanup`, kein `t.Parallel`), ruft die fünf Operationen nacheinander und erwartet je Aufruf genau die nächsten 16 Bytes in Hex (`ses-000102030405060708090a0b0c0d0e0f`, dann `exc-101112…1f` usw.). Mutation: ein **paketweiter** mutexgeschützter `math/rand`-Generator mit festem Seed: rot (die Kennung stammt nicht aus dem Leser); `rand.Read(b[:8])`: rot. Der Implementer prüft, dass `rand.Read` den ersetzten `Reader` verwendet (Go-Dokumentation: „Read calls io.ReadFull on Reader“); trägt das nicht, gibt er es an den Architect zurück. Akzeptiertes Negativ bleibt die Güte des Standard-`Reader` (Betriebssystem, `SPEC-114` *Ungeprüft*): von außen nicht messbar, und die Kennung ist keine Geheimhaltung |
| Kollision bei 128 Bit vernachlässigbar, gleichzeitig sicher (Nr. 8, Einleitung); **gesagt ist kein „kollisionsfrei“** | `IDVertrag`, Hilfsfunktion `eindeutig` (läuft in `TestVertrag` und `TestFestVertrag`; ein Test dieses Namens gibt es nicht): 64 Goroutinen × 500 Aufrufe je Operation, alle Kennungen paarweise verschieden. Mutation: nur 1 Byte Zufall: rot; feste Kennung: rot |
| Erzeugte Kennung trifft `ErrBereitsVorhanden`: `SPEC-109` wie jede andere, kein Wiederholen (Nr. 8) | kein Fänger in diesem Slice — **akzeptiertes Negativ**, Grund: eine Aussage über den Verbraucher, den es hier nicht gibt; die Verbraucher-Slices führen die Zelle nicht, weil der Fall bei 128 Bit nicht eintritt |
| Client-Kennungen berührt der Generator nicht (Nr. 9) | `TestPortSignaturen` (strukturell): jede Methode ist `func() string`, der Generator nimmt also keine Eingabe an; eine Allaussage „berührt nie“ folgt daraus, nicht aus einem Verhaltenstest, und ist damit ausreichend gefangen. Mutation: ein Parameter an einer Operation: rot. Dass ein Verbraucher eine Client-Kennung durch den Generator ersetzte, kann hier nicht eintreten (es gibt keinen Aufrufer) und gehört den Nehmern (§1 *Übergibt*) — **akzeptiertes Negativ**; die Form legt der Use Case fest |
| `FestIDs`: Zähler je Operation ab 1, Instanzen unabhängig, gleichzeitig lückenlos und ohne Doppelung (Nr. 10) | `TestFestIDsFolge`: erste drei Kennungen je Operation enden auf `01`, `02`, `03`; die Folge geht über die Zehnergrenze: `…09` gefolgt von `…10` (dezimal, nicht `…0a`); eine zweite Instanz beginnt bei `01`; `TestFestIDsNebenlaeufig`: 64 × 500 Aufrufe einer Operation ergeben genau die Nummern 1 bis 32000. Mutation: ein Zähler für alle Operationen oder globaler Zähler: rot; `%032x` statt `%032d`: rot an `…10`; Sperre entfernen: im Review rot gesehen (Absturz durch Map-Schreiben), ohne Race-Detektor nicht mathematisch sicher — **akzeptiertes Negativ**, Grund: das Gate fährt kein `-race` (siehe oben) |
| `FestUhr`/`FestIDs` erfüllen `UhrVertrag`/`IDVertrag` wie die echten (Einleitung) | derselbe Aufruf von `UhrVertrag(t, …)` und `IDVertrag(t, …)` für beide Implementierungen im Test des Pakets `clock` bzw. `supporttest`. Mutation: Fake mit anderer Form: rot |
| Schichten: Adapter und Fakes importieren nur Modell und `ports/support` | `make a-check`: Kante `adapters-driven` → `ports-support` ist erlaubt, `adapters-driven` → `adapters-driven` nicht. Mutation: Import von `adapters/driven/sqlite` in `clock`: `a-check` rot |

Diese Aussagen stehen in `SPEC-114` und haben bewusst keinen Fänger hier, weil sie nur Verbraucher binden: eine Kennung ordnet nichts (Nr. 6, die Ordnung tragen `SPEC-105` Nr. 4, 11), `finished_at` liegt nie vor `started_at` (Folge von Nr. 2; die Prüfung trägt `SPEC-105` Nr. 10 im Modell).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` ist eröffnet; kein Vorgänger-Slice. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 1000 Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3 einschließlich `spec/spezifikation.md`, Tests und `supporttest` eingeschlossen; Basis ist der Commit vor dem ersten Code-Commit). Die Schätzung liegt bei rund 700 Zeilen in derselben Metrik: Produktcode etwa 130 (Ports 40, Adapter 90), `supporttest` etwa 250, Tests der Pakete etwa 250, Spezifikation etwa 25 (der Plan selbst liegt nicht unter den Pfaden). Der Vorgänger lag beim Dreifachen seiner Schätzung (2369 gegen 600); die Marke trägt das ein. Gemessen mit `git diff --numstat fe372f7` über `src/hexagon/ports/support/`, `src/adapters/driven/clock/` und `spec/spezifikation.md`: 683 Zeilen, nach dem Stand mit `TestSystemuhrQuelleSerialisiert` und `TestZufallsIDsLiestCryptoRand`.
- `in-progress` → `open` (blockiert): `a-check` meldet für `src/hexagon/ports/support/`, `supporttest` oder `src/adapters/driven/clock/` eine Kante, die `.a-check.yml` heute nicht kennt (die Schicht `ports-support` und ihre Kanten liegen vor); die Regel wird dort ergänzt, nicht umgangen (`AGENTS.md` §3.6), und der Plan zieht nach.

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

- Die Form der Kennungen ist nicht festgelegt (Präfix, Länge, Alphabet) — **Ausgang:** entfallen; entschieden in `SPEC-114` Nr. 6 (`<Präfix>-<32 × 0-9a-f>`, 36 Zeichen) und Nr. 7 (128 Bit `crypto/rand`).
- Zeit mit Auflösung unter einer Sekunde gegen die Speicherung in SQLite (`SPEC-105`: Jahresbereich, Nullzeit) — **Ausgang:** entfallen; entschieden in `SPEC-114` Nr. 1 bis 3: Nanosekunden ohne Rundung (der Speicher hält neun Stellen, `outboundtest` belegt es), Bereich und Nullzeit der Systemuhr werden nicht ersetzt, sondern vom Port abgewiesen.
- Rücksprung der Systemzeit lässt `finished_at` vor `started_at` fallen (`SPEC-105` Nr. 10) — **Ausgang:** entfallen; `SPEC-114` Nr. 2: die Uhr ist nicht fallend, `TestSystemuhrRuecksprung`.
- Nebenläufige Tests laufen nur im Glück grün (`BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen`, 2×) — **Ausgang:** eingetreten und benannt; die Zeilen der Tabelle in §3 trennen sichere von nicht mathematisch sicheren Fängern, die nicht mathematisch sicheren stehen als akzeptiertes Negativ mit Grund. Der Fänger „Sperre in `Systemuhr` entfernen“ war ohne Wirkungs-Fänger nur in 11 von 20 Läufen rot (Review F3); die Beobachtung ist im Register als weiterer Beleg geführt (§7). Ein Gate mit `-race` gibt es nicht (nächstes Risiko).
- Das Gate fährt kein `-race`; Datenrennen in `FestUhr`, `FestIDs`, `Systemuhr` sind nur über ihre Wirkung sichtbar — **Ausgang:** eingetreten und benannt; das Image kann kein `-race` („requires cgo“), die Zeilen der Tabelle in §3 tragen einen Wirkungs-Fänger (`TestSystemuhrQuelleSerialisiert`) oder ein akzeptiertes Negativ; ein Gate mit `-race` wäre ein eigener Vorschlag, keine Lockerung (`AGENTS.md` §3.6), und gehört nicht in diesen Slice.
- Randformen (`AGENTS.md` §3.12, §3.16) — **Ausgang:** entfallen; das Raster unten steht vor dem ersten Code-Commit, jede Zelle nennt ihren Ort. Bemerkt der Implementer eine Zelle ohne Ort, hält er an und gibt sie an den Architect zurück.
- Die Form von Client-Kennungen (leer, überlang, fremde Zeichen) ist in `SPEC-114` Nr. 9 an die Use Cases der Nehmer gegeben, nicht entschieden — **Ausgang:** übergeben an `slice-workspace-use-case` (§1 *Übernimmt* und §6) und `slice-session-wartung` (§1 *Übernimmt*), beide in `open/` und mit der Kennung dieses Slice als Geber (Verifikation, Abschnitt 8); sie berühren diesen Slice nicht, weil der Generator keine Eingabe sieht.

**Raster Operation × Argumentart** (`AGENTS.md` §3.16). Jede Zelle nennt Nummer der Spezifikation (`SPEC-114`) oder „gilt nicht“ mit Grund. Die Operationen haben kein Argument; die Zeilen je Modus (echt, fest) und die Paare stehen trotzdem einzeln.

| Operation | fehlender Wert | leerer Wert | unbekannter Wert | fremder / ungültiger Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| `Systemuhr.Now` | gilt nicht: kein Argument (Einleitung) | gilt nicht: kein Argument | gilt nicht: kein Argument | Systemzeit springt zurück: bleibt auf dem Höchstwert (Nr. 2); Nullzeit oder Jahr außerhalb 1 bis 9999: unverändert durchgereicht, der Port weist ab (Nr. 3) | gilt nicht: kein Kontext, blockiert nie (Einleitung, Nr. 4) | nicht fallend, gleicher Wert erlaubt (Nr. 2); gleichzeitig sicher (Einleitung) |
| `Systemuhr.Now`, Paar Rücksprung und Bereich | gilt nicht | gilt nicht | gilt nicht | Nullzeit nach gültigem Wert: bleibt auf dem Höchstwert; Wert hinter Jahr 9999 nach gültigem Wert: wird Höchstwert und bleibt es bis zum Neustart (Nr. 3) | gilt nicht | wie oben |
| `FestUhr.Now` | gilt nicht: kein Argument | gilt nicht | gilt nicht | gilt nicht: der Stand ist gültig, sonst entsteht die Uhr nicht (Nr. 5) | gilt nicht: kein Kontext | gleicher Wert bis `Advance` (Nr. 5) |
| `NewFestUhr(start)` | Nullzeit `time.Time{}`: Panik (Nr. 5) | wie fehlend: die Nullzeit ist der leere Wert (Nr. 5) | gilt nicht: kein Katalog | andere Zone: nach UTC umgerechnet; Jahr außerhalb 1 bis 9999 nach der Umrechnung: Panik (Nr. 5) | gilt nicht: synchron, kein Kontext | jede Instanz hat ihren eigenen Stand (Nr. 5) |
| `NewFestUhr`, Paar Nullzeit und Bereich | gilt nicht: die Nullzeit liegt im Jahr 1 und im Bereich, beide Verstöße schließen sich aus; die Reihenfolge Nullzeit, dann Bereich ist gesetzt (Nr. 5) | gilt nicht | gilt nicht | gilt nicht | gilt nicht | gilt nicht |
| `FestUhr.Advance(d)` | gilt nicht: Pflichtargument vom Typ `Duration` | `d == 0`: erlaubt, Stand unverändert (Nr. 5) | gilt nicht: kein Katalog | `d < 0`: Panik, Stand unverändert; Ergebnis nach Jahr 9999: Panik, Stand unverändert (Nr. 5); beide Verstöße gleichzeitig sind ausgeschlossen, ein negatives `d` führt nie über 9999 | gilt nicht: kein Kontext | Wiederholung und gleichzeitiger Aufruf summieren sich (Nr. 5) |
| `ZufallsIDs.SessionID`, `ExchangeID`, `KorrelationsID`, `WorkspaceID`, `SnapshotID` | gilt nicht: kein Argument | gilt nicht: kein Argument | gilt nicht: kein Argument | Zufallsquelle fällt aus: Prozess endet, kein Fehlerwert (Nr. 7); **kein Fänger, akzeptiertes Negativ:** das Verhalten gehört der Go-Laufzeit (`rand.Read` füllt den Puffer vollständig, gibt keinen Fehler zurück und beendet bei einem Fehler des `Reader` das Programm unwiederbringlich), ein Test brauchte einen Unterprozess und prüfte nur die Laufzeit. Offen bleibt die Mutation „`io.ReadFull(rand.Reader, …)` mit ignoriertem Fehler“, die bei Ausfall Nullbyte lieferte; kein Test fängt sie, der Ausfall tritt auf den unterstützten Plattformen nicht ein (`SPEC-114` *Ungeprüft*) | gilt nicht: kein Kontext, blockiert nie | jede Kennung neu, Kollision vernachlässigbar (Nr. 8); gleichzeitig sicher (Einleitung) |
| Verbraucher: erzeugte Kennung trifft `ErrBereitsVorhanden` | gilt nicht | gilt nicht | gilt nicht | `SPEC-109` wie jede Kennung, kein Wiederholen (Nr. 8) | gilt nicht | gilt nicht: tritt bei 128 Bit nicht ein |
| `FestIDs.SessionID`, `ExchangeID`, `KorrelationsID`, `WorkspaceID`, `SnapshotID` | gilt nicht: kein Argument | gilt nicht | gilt nicht | gilt nicht: der Zähler hat keine Fehlerform | gilt nicht: kein Kontext | Zähler je Operation +1 ab 1, dezimal geschrieben, `…09` folgt `…10` (Nr. 10); zwei Instanzen unabhängig; gleichzeitig lückenlos (Nr. 10) |
| `FestIDs`, Paar zweier Operationen | gilt nicht | gilt nicht | gilt nicht | gilt nicht | gilt nicht | jede Operation zählt für sich, `ses-…01` und `exc-…01` stören sich nicht (Nr. 10) |
| Client-Kennung (Replay-Session, `import_id`, mitgegebene `id`) | Generator sieht sie nicht (Nr. 9) | wie fehlend; die Form legt der Use Case der Nehmer fest (Nr. 9) | wie fehlend | wie fehlend | gilt nicht | gilt nicht |

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

- **Was hat funktioniert:** Der Vertrag `SPEC-114` stand vor dem Code, mit Tabelle *Zusage · Fänger* und Raster. Das Review (`docs/reviews/2026-10-20-slice-clock-idgenerator-review.md`, rund 45 Mutationen, keine HIGH) und die Verifikation (`docs/reviews/2026-10-20-slice-clock-idgenerator-verifikation.md`, Urteil bestanden, `make gates` selbst gefahren, grün) lösten jede Zeile mit einer Mutation auf. `IDVertrag` zählt die Methoden von `IDGenerator` mit `reflect` in beide Richtungen auf; eine sechste Methode oder eine Tabellenzeile ohne Methode wird rot. Ein Wirkungs-Fänger (`TestSystemuhrQuelleSerialisiert`: die Quelle zählt gleichzeitige Aufrufe) ersetzt dort den Race-Detektor, wo das Image keinen hat.
- **Was ging anders als geplant:** (1) Zwei Zusagen des Plans hatten keinen Fänger: „die Quelle wird unter der Sperre gelesen“ (F1) und „keine feste Folge“ (F2); beide haben jetzt einen. (2) Der Plan nannte einen Lauf („`-count=20` und unter `-race` rot gesehen“), den das Image nicht trägt, und ordnete die Wirkung dem falschen Test zu (F3); gemessen: „Sperre entfernen“ ist nur über `TestVertrag` und nur in 11 von 20 Läufen rot. (3) Die Darstellung des Zählers von `FestIDs` entschied zuerst der Code, dann `SPEC-114` Nr. 10 (F4). (4) Eine Raster-Zelle (Ausfall der Zufallsquelle) hatte weder Fänger noch benanntes Negativ (F5). (5) Die Verifikation fand einen Fänger-Namen ohne Test (V1): die Eindeutigkeit prüft die Hilfsfunktion `eindeutig`, aufgerufen aus `IDVertrag`; der Plan ist berichtigt. (6) Marke §4: `git diff --numstat fe372f7..ceed127` (letzter Commit mit Änderung an einem Pfad aus §3) über `src/hexagon/ports/support/`, `src/adapters/driven/clock/` und `spec/spezifikation.md` ergibt 683 gegen die Marke 1000 und die Schätzung 700 (Summe aus Einfügungen und Löschungen, Tests eingeschlossen: 681 + 2; davon `supporttest` 366, die Spezifikation 6, weil `SPEC-114` schon im Basis-Commit stand); nach den Commits dieser Closure, die nur `docs/` berühren, unverändert. Keine Rückführung.
- **Steering-Loop-Eintrag:** Regelvorschlag, nicht verkörpert und deshalb ohne Zielort (der Auftraggeber entscheidet; `AGENTS.md` bleibt unberührt): Wo ein Vertrag eine Zusage über Nebenläufigkeit macht (Sperre, Zähler, Reihenfolge), gilt sie nur als gefangen, wenn die Mutation in einer Wiederholung (`-count=20`) in jedem Lauf rot wird; sonst bekommt sie einen Wirkungs-Fänger, der die Wirkung zählt statt auf ein Datenrennen zu warten, oder steht als akzeptiertes Negativ mit Grund im Plan. Der Plan nennt keinen Lauf (`-race`), den das Image nicht fährt. Vorgesehener Ort wäre ein Satz in `AGENTS.md` §3.10 (Mutation „vom Implementer selbst rot gesehen“). Auslöser: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (slice-gegenproben, slice-sqlite-adapter-migration, slice-clock-idgenerator — 3×). Zuordnung zur Beobachtung ist ein Urteil: die ersten beiden Belege sind ein Lauf, der nur im Glück grün war, der dritte ein Fänger, der nur im Glück rot wird; die Ursache ist dieselbe (Nebenläufigkeit zeigt sich ohne `-race` nur zufällig). Ab 3× verlangt das Register einen Ausgang (verkörpert oder geplant); er ist hier noch nicht gesetzt.
  Weitere Lernkandidaten, jeweils nur gezählt: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (F1, F2, F5; bereits verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (F4; bereits verkörpert in `AGENTS.md` §3.16). Die Randform trat trotz Raster auf, weil der Zähler eine Darstellung (Basis) ist, kein Wert-Fall; das Raster Operation × Argumentart führt keine Zeile „Darstellung des Ergebnisses“. Das ist ein Hinweis für den Architect, keine Regel.
- **Beobachtungs-Register (`../observations/`):** `evidence/slice-clock-idgenerator.md` in `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen/` (jetzt 3×), `BEO-REPO/randform-erst-im-review-entschieden/` und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft/` ergänzt; keine neue Beobachtung. Die Zähler folgen aus den Dateien.
- **Folge-Slices:** `slice-http-adapter`, `slice-record-service`, `slice-session-wartung`, `slice-workspace-use-case`, `slice-baseline-import` und `slice-bootstrap-m2` — alle sechs sind Dateien in `open/` und führen die Sendung (Zeit, Kennungen, Verdrahtung) mit der Kennung dieses Slice in §1. Die Form einer vom Client gelieferten Kennung tragen `slice-workspace-use-case` und `slice-session-wartung`.
- **Risiken aus §6:** alle mit Ausgang — Form der Kennungen, Auflösung der Zeit, Rücksprung, Randformen (entfallen mit Grund), nebenläufige Tests und `-race` (eingetreten und benannt), Client-Kennungen (übergeben).
- **Drei Paarungen:** Anker (kein Anker, nichts verkörpert) · Folge-Slice (sechs Nehmer in `open/`) · Register (drei Verzeichnisse oben) — von der nächsten Welle-Closure geprüft; der letzte DoD-Punkt bleibt bis dahin offen.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (2×); die Randformen stehen als Raster in §6. Der Slice ist nebenläufig: Wiederholungsläufe (`-race`, `-count`) gehören vor die Übergabe, und ein drittes Auftreten dieser Beobachtung in dieser Welle ist kein Befund mehr, sondern ein eigener Slice.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
