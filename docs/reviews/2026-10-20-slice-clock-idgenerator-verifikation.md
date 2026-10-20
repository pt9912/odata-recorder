# Verifikation: slice-clock-idgenerator

* Datum: 2026-10-20
* Rolle: verifier (frischer Kontext; kein Code geändert, kein Commit)
* Gegenstand: `git diff fe372f7..HEAD`, Stand 72f2b12
* Plan: `docs/plan/planning/in-progress/slice-clock-idgenerator.md`
* Maßstab: DoD (§2), `SPEC-114`, Tabelle Zusage·Fänger (§3), Raster (§6), `AGENTS.md` §3.7, §3.11, §3.13 bis §3.16, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); Bezug `LH-QA-02`, `LH-FA-13`
* Eingang: Review `docs/reviews/2026-10-20-slice-clock-idgenerator-review.md` (F1 bis F8)
* Belege selbst gefahren: `make gates` (Exit 0, Arbeitsbaum danach sauber); zwei Mutationen in Wegwerfkopie im Image `odata-recorder:test`, netzlos, Kopie nur lesend gemountet; `git diff --numstat`.

## Urteil

**Bestanden, mit zwei LOW-Befunden und der Liste dessen, was für die Closure noch fehlt.** Kein Befund verletzt `SPEC-114` oder eine DoD-Zusage des Codes; die offenen Punkte sind Plan-Hygiene und Closure-Pflichten.

## 1. DoD gegen Code und Tests

| DoD-Punkt | Befund |
|---|---|
| Ports `Clock`/`IDGenerator` unter `src/hexagon/ports/support/`, Adapter `Systemuhr`/`ZufallsIDs` (UTC, nicht fallend, Form Nr. 6) | erfüllt: `support.go`, `clock.go`; Zusicherungen `var _ support.Clock`, `var _ support.IDGenerator`; Form per `IDVertrag` |
| `FestUhr`, `FestIDs`, `UhrVertrag`, `IDVertrag` in `supporttest`, gegen echte und feste Implementierung; Eindeutigkeit gleichzeitig | erfüllt: `TestVertrag` (clock) und `TestFestVertrag` rufen dieselben Vertragsfunktionen; `eindeutig` in `IDVertrag` (64 × 500 je Operation) |
| `make gates` grün | **selbst gefahren: grün** (Exit 0; `baseline-verify` OK, `a-check`, `lint`, `pruefung test` ohne Fehlermeldung) |
| Review durchgeführt, Report liegt vor | erfüllt (Report vom selben Tag); Häkchen in der DoD fehlt noch (siehe Closure) |
| Doku-Update (öffentlicher Vertrag) | erfüllt: `SPEC-114` samt Historie-Zeilen, `LH-QA-02.a` verweist darauf (`spec/spezifikation.md` Zeile 422) |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen, siehe unten |

Die Häkchen der DoD in §2 sind alle noch leer; das ist vor der Closure zu setzen (Inhalt vor `git mv`, `AGENTS.md` §3.3).

## 2. SPEC-114 gegen Code, alle Nummern

| Nr. | Zusage | Code / Beleg |
|---|---|---|
| 1 | UTC, ohne monotone Ablesung, keine Rundung | `clock.go` `.UTC()`; `pruefeForm` (Ort, `z == z.Round(0)`); `TestSystemuhrOhneRundung` |
| 2 | nicht fallend, gleicher Wert erlaubt, gleichzeitig sicher | `hoechst`/`gesehen` unter `mu`; `TestSystemuhrRuecksprung`, `TestSystemuhrNebenlaeufig`, `UhrVertrag` |
| 3 | Bereich/Nullzeit unverändert durchgereicht, Nr. 2 gilt | `TestSystemuhrBereichUngeprueft` (Null, Jahr 0, Null nach `t`, Jahr 10000 wird Höchstwert) |
| 4 | `Clock` ohne Warten/Timer, Port ohne Argument/Kontext/Fehler | `TestPortSignaturen` (genau eine Methode `Now`, alle `IDGenerator`-Methoden `func() string`) |
| 5 | `FestUhr`: Start nach UTC, Panik Nullzeit/Bereich, `Advance` | `fest.go`; `TestFestUhrStart`, `TestFestUhrAdvance`, `TestFestUhrNebenlaeufig`; Reihenfolge Nullzeit vor Bereich im Code (`IsZero` zuerst) |
| 6 | fünf Operationen, Präfixe, Form 36 Zeichen | `IDVertrag` (Reflect in beide Richtungen, Regex `^[a-z]{3}-[0-9a-f]{32}$`, Präfix-Tabelle) |
| 7 | 128 Bit `crypto/rand`, kein Seed | `zufall`; `TestZufallsIDsUnabhaengig`, `TestZufallsIDsLiestCryptoRand` |
| 8 | Kollision vernachlässigbar; kein Wiederholen | `eindeutig`; Verbraucher-Teil benanntes Negativ (kein Verbraucher) |
| 9 | Client-Kennungen unberührt | strukturell durch `TestPortSignaturen`; Form liegt bei den Nehmern (Punkt 8 unten) |
| 10 | `FestIDs`: Zähler je Operation, dezimal, Instanzen unabhängig, lückenlos | `fest.go` `%032d`; `TestFestIDsFolge` (bis 12, Literale), `TestFestIDsNebenlaeufig` (1 bis 32000) |

Keine Abweichung zwischen Spezifikation und Code gefunden. *Ungeprüft* in `SPEC-114` deckt die benannten Negative (Plattformuhr, Güte der Quelle, Ausfall der Quelle, Datenrennen ohne `-race`).

## 3. Zusage·Fänger-Tabelle (§3), Raster (§6), Allaussagen

Alle in der Tabelle genannten Fänger sind aufgelöst gegen die Testnamen (`grep ^func Test`): `TestVertrag`, `TestFestVertrag`, `TestSystemuhrOhneRundung`, `TestSystemuhrRuecksprung`, `TestSystemuhrNebenlaeufig`, `TestSystemuhrQuelleSerialisiert`, `TestSystemuhrBereichUngeprueft`, `TestPortSignaturen`, `TestFestUhrStart`, `TestFestUhrAdvance`, `TestFestUhrNebenlaeufig`, `TestZufallsIDsUnabhaengig`, `TestZufallsIDsLiestCryptoRand`, `TestFestIDsFolge`, `TestFestIDsNebenlaeufig`, `IDVertrag`/`UhrVertrag` und `make a-check` existieren.

### V1 LOW — Fänger-Name ohne Test
* pfad: Plan §3, Zeile *Kollision bei 128 Bit* · `TestZufallsIDsEindeutig (in IDVertrag)`
* befund: Einen Test dieses Namens gibt es nicht; die Prüfung ist die Hilfsfunktion `eindeutig` in `vertrag.go`, aufgerufen aus `IDVertrag`. Die Zeile löst nur über die Klammer „in `IDVertrag`" auf. Der Plan erlaubt dem Implementer, den vergebenen Namen im Bericht zu nennen; im Plan nachgezogen ist er nicht (`AGENTS.md` §3.9).
* Folge: ein Leser sucht einen Test, den es nicht gibt. Behebung: Zeile auf `IDVertrag` (Hilfsfunktion `eindeutig`) umstellen.

Raster §6: keine leere Zelle; jede Zelle nennt Nummer oder „gilt nicht" mit Grund. Die Zelle *Zufallsquelle fällt aus* trägt jetzt ein ausdrücklich akzeptiertes Negativ mit Grund und der offenen Mutation, und `SPEC-114` *Ungeprüft* nennt es.

Allaussagen (geschärfte §3.16): Nr. 6 (alle Operationen) wird über die Methodenmenge per Reflect in beide Richtungen gefangen; Nr. 10 (jede Operation zählt für sich, Paar) über die Schleife über alle fünf Operationen in `TestFestIDsFolge`; Nr. 4/9 (alle Methoden `func() string`) über Reflect. Als Menge ungefangen bleibt „Präfixe sind verschieden" (Review F6, INFO, siehe unten).

Akzeptierte Negative der Tabelle, jeweils mit Grund: Datenrennen in `FestUhr`/`FestIDs` ohne `-race`, Verbraucher trifft `ErrBereitsVorhanden`, Client-Kennung durch Generator ersetzt, Ausfall der Zufallsquelle, Güte von `crypto/rand`. Alle begründet.

## 4. Review-Befunde F1 bis F8

| F | Stufe | Stand | Beleg |
|---|---|---|---|
| F1 | MEDIUM | erledigt | `TestSystemuhrQuelleSerialisiert` vorhanden. **Selbst gemutet:** `u.quelle()` vor `u.mu.Lock()` ergibt „Quelle lief mit 8 gleichzeitigen Aufrufen, erwartet 1", rot aus dem benannten Grund. Plan §3 trennt jetzt die Reihenfolge-Zusage von `TestSystemuhrNebenlaeufig` |
| F2 | MEDIUM | erledigt | `TestZufallsIDsLiestCryptoRand` vorhanden. **Selbst gemutet:** `rand.Read(b[:8])` ergibt fünf rote Zeilen mit halb gefüllten Kennungen. Die paketweite `math/rand`-Form fällt nach Konstruktion durch (Kennung stammt nicht aus dem ersetzten Leser); nicht separat gefahren |
| F3 | LOW | erledigt | Plan §3 nennt die gemessene Wirkung (20 Läufe, 11-mal rot, nur über `TestVertrag`); §6 benennt „requires cgo" |
| F4 | LOW | erledigt (Historie) | Nr. 10 in Spec, Plan §3 und Raster nachgezogen; kein offener Zustand. Gehört in den Lerneintrag (Beobachtung `BEO-REPO/randform-erst-im-review-entschieden`, Wiederholung) |
| F5 | LOW | erledigt | Zelle als akzeptiertes Negativ mit Grund; *Ungeprüft* in `SPEC-114` ergänzt; Kommentar von `zufall` grammatisch berichtigt |
| F6 | INFO | unverändert, akzeptiert | Präfix-Eindeutigkeit nur über Literale; keine Aktion |
| F7 | INFO | Teil (b) erledigt | Nullwert `Systemuhr{}` paniert weiterhin; Plan-Zahl nachgezogen (siehe 5) |
| F8 | INFO | unverändert | `a-check` hier grün; ungeschichtete `tools/`-Dateien unabhängig vom Diff |

## 5. Metrik §4

`git diff --numstat fe372f7..HEAD` über `src/hexagon/ports/support/`, `src/adapters/driven/clock/`, `spec/spezifikation.md`: 4+2, 72, 189, 6, 19, 25, 91, 150, 125 = **683** Zeilen (Summe aus Einfügungen und Löschungen). Plan nennt 683; die Zahl stimmt. Marke 1000, Schätzung 700: keine Rückführung. `harness/gegenproben.stempel` und der Review-Report liegen nicht unter den Plan-Pfaden und zählen nicht. Wächst die Datei noch durch Nachzüge, ist die Zahl bei der Closure erneut zu messen.

## 6. Kommentare (§3.7, §3.11)

Die Kommentare von Ports, `Systemuhr`, `FestUhr`, `FestIDs` und der Vertragsfunktionen sagen nur, was ein Test prüft (UTC/Rundung, Höchstwert, Panik ohne Standänderung, Zähler, Quelle unter der Sperre). Keine verworfene Alternative, keine Chronik.

### V2 LOW — Laufzeit-Zusage im Kommentar von `zufall` ohne Test
* pfad: `src/adapters/driven/clock/clock.go` · Kommentar zu `zufall` („gibt keinen Fehler zurück und beendet bei einem Fehler der Quelle das Programm")
* befund: Die Aussage beschreibt die Go-Laufzeit; kein Test prüft sie, und `SPEC-114` *Ungeprüft* sagt dasselbe offen. Als Kopplung an die Laufzeit ist sie zulässig, trägt aber ein Verhalten zu, das §3.11 nur mit Fänger oder ausdrücklichem Negativ erlaubt. Der Kommentar nennt das Negativ nicht.
* Folge: gering; der Kommentar könnte auf „Ausfall der Quelle ist nicht getestet (`SPEC-114` *Ungeprüft*)" ergänzt werden. Kein Blocker.

## 7. Handover-Sendungen (§3.13)

Jeder Nehmer führt die Sendung in §1 (*Übernimmt*) mit der Kennung des Gebers; alle sechs liegen in `open/`, keiner in `done/`, und kein Ausschluss trifft die Sendung:

| Nehmer | Sendung | Stelle |
|---|---|---|
| `slice-http-adapter` | Korrelations-ID | §1 Zeile 36 |
| `slice-record-service` | Exchange-Kennung, Zeit | §1 Zeile 36; Start nennt `slice-clock-idgenerator` in `done/` |
| `slice-session-wartung` | Session-Kennung, Zeit, Form einer gelieferten `id` | §1 Zeile 38; Start nennt den Geber in `done/` |
| `slice-workspace-use-case` | Workspace-Kennung, Zeit | §1 Zeile 38 |
| `slice-baseline-import` | Snapshot-Kennung, Zeit | §1 Zeile 38 |
| `slice-bootstrap-m2` | eine Instanz je Prozess | §1 Zeile 36 |

Die Ausschlüsse von `slice-workspace-use-case` und `slice-session-wartung` nennen weder Zeit noch Kennungen. Die DoD der Nehmer wurde nicht einzeln ausgewertet; §1 genügt der Regel („§1 oder DoD").

## 8. Nr. 9, Client-Kennungen

Adresse vorhanden: `slice-session-wartung` §1 legt die Form einer vom Aufrufer gelieferten `id` (leer, überlang, Zeichen) „mit dem Architect vor dem Code" fest, mit Verweis auf `SPEC-114` Nr. 9 und Kennung des Gebers. `slice-workspace-use-case` führt die Form einer mitgegebenen Workspace-`id` in §1 und als offenes Risiko in §6 (Zeile 122, Ausgang „offen; Raster-Zelle"). Beide Nehmer liegen in `open/` und nehmen an. Die `import_id` ist in `slice-baseline-import` §1 beim Client belassen.

## Was für die Closure fehlt

1. Häkchen der DoD (§2) setzen; Inhalt vor dem reinen `git mv` nach `done/` (`AGENTS.md` §3.3).
2. §7 Closure-Notiz mit Steering-Loop-Lerneintrag füllen (Klassen des Reviews: Zusage ohne Fänger, Plan nennt Fänger, der nicht fängt, Randform im Code vor der Spezifikation entschieden, Raster-Zelle ohne Fänger); `liegt in` nur, wenn etwas verkörpert wurde.
3. Beobachtungs-Register fortschreiben oder „keine Beobachtung" in §7 notieren. Das Risiko `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` steht bei 2×; ein Eintrag zum dritten Auftreten (ein Test, der ohne `-race` mal grün, mal rot ist) ist mit dem nur teilweise gefangenen „Sperre entfernen" in `Systemuhr` und `FestUhr` zu entscheiden und zu benennen.
4. Alle Ausgänge in §6 auf eine der drei Formen bringen: *Nebenläufige Tests … weiter offen bis zur Closure* und *Form von Client-Kennungen … weiter offen* brauchen bei Closure einen Ausgang bzw. die Bestätigung der Adresse (Träger `slice-workspace-use-case`, `slice-session-wartung`); „eingetreten und benannt" beim `-race`-Risiko ist bereits ein Ausgang.
5. V1 im Plan nachziehen (Fänger-Name); V2 optional.
6. Die Zahl 683 beim letzten Commit erneut messen und die Closure-Zahl nennen.
7. Die drei Paarungen (Anker · Folge-Slice · Register) bleiben der nächsten Welle-Closure vorbehalten; die Folge-Slice-Seite ist durch Abschnitt 7 getragen.

## Nicht geprüft / Grenzen

* Kein `-race`: das Image kann es nicht („requires cgo"); Datenrennen sind nur über ihre Wirkung sichtbar (so im Plan zugesagt).
* Nicht jede Mutation des Reviews (rund 45) wiederholt; nachgefahren sind die für F1 und F2 neu hinzugekommenen Fänger.
* Die DoD der sechs Nehmer-Slices wurde nicht Zeile für Zeile gelesen.
