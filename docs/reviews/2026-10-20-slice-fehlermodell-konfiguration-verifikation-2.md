# Nachtrag 2: Verifikation slice-fehlermodell-konfiguration

* Datum: 2026-10-20
* Rolle: verifier (frischer Kontext), kein Code geändert, kein Commit
* Gegenstand: HEAD 2baecef, Plan `docs/plan/planning/done/slice-fehlermodell-konfiguration.md`
* Eingang: `...-review.md` (F-1..F-9), `...-verifikation.md` (V-1), `...-review-re-review.md` (N-1..N-6)
* Sensor selbst gefahren: `make gates`, Exit 0 (Ausgabe bis `migration-check: ok`). Arbeitsbaum danach sauber bis auf diesen Bericht.
* Nicht selbst gefahren: die im Plan genannten Mutationen (N-1, N-2). Mutieren hieße Code ändern; ihr Rot-Sehen bleibt Beleg des Implementers bzw. eines Reviewer-Laufs.

## Urteil

**Bestanden, ohne Befund über LOW.** Alle Befunde sind im Code, in der Spezifikation und im Plan erledigt und haben einen Fänger. Die Closure fehlt noch (siehe Schluss).

## (1) Befunde gegen Code, Plan, Spec

| Befund | Stand | Beleg |
|---|---|---|
| F-1 Default `DATA_DIR` | erledigt | Fänger `TestKonfigDefaults`, Plan-Zeile Defaults |
| F-2 Randformen im Code entschieden | erledigt | `SPEC-110` Nr. 1, 6, 7, 9, 12; Plan §6 trägt sie, Historie-Zeile 2026-10-20 |
| F-3 Marke §4 | erledigt | §4 nennt Metrik, Zahl, Begründung (siehe 4) |
| F-4 Kommentar | erledigt | Re-Review N-4; Kommentare siehe 5 |
| F-5 Sortierung | erledigt | `konfig.go` `fehler()`: Schlüssel, dann Grund (`cmp.Or`), `Compact`; Fänger `TestKonfigMeldung`, `TestKonfigMeldungSortierung` |
| V-1 falscher Typ, zwei Zeilen | erledigt | `dateiEintrag`: `text` liefert `ok=false`, danach keine Weiterprüfung; `SPEC-110` Nr. 9 Absatz „steht in der Datei und fehlt nicht“; Fänger `TestKonfigDienstefeldTyp` (Plan-Zeile 119) |
| N-1 Typfehler `upstream_url` nur in `record` gefangen | erledigt | `TestKonfigDienstefeldTyp` enthält jetzt `upstream_url_replay` und `upstream_url_mock`; Plan-Zeile nennt sie samt Mutation |
| N-2 `VERIFY=yes` meldete zwei Zeilen | erledigt | `liesTLS` nutzt `optionalerWahrheitswert` (Wert, gesetzt, gültig); Widerspruch nur bei zwei gültigen Werten; `SPEC-110` Nr. 10 zweiter Absatz; Fänger `TestKonfigTLSUngueltig` mit genauer Zeilenmenge (8 Fälle, auch „braucht“/„widerspricht“ ausgeschlossen) |
| N-3 Testkommentar | erledigt | Suche nach „umordnen“ und „'.' <“ in `konfig_test.go`: keine Treffer |
| N-4..N-6 | INFO, keine Handlung | |

## (2) Zusage · Fänger, SPEC-110 Nr. 9/10/12

* Alle 40 `Test…`-Namen der Tabelle §3 (Zeilen 93 bis 126) existieren als `func` unter `src/` (per Schleife über die Namen geprüft, keine Ausgabe „MISSING“). Der Inhalt der drei geänderten Zeilen (119, 122, 124) steht in den Tests (Zeilenmengen, Modus `replay`/`mock`).
* Behauptung „jeder Schlüssel höchstens ein Grund“ (Nr. 12): aktiv gegen den Code gesucht. Alle `melde`-Stellen durchgegangen:
  * `MODE`: ungültig oder nicht unterstützt, durch frühen Rücksprung exklusiv.
  * Nicht gelieferte Schlüssel: Präfixe (`ADMIN_`, `HYBRID_`, `FALLBACK_`, `LISTEN_TLS_`) und Namen überschneiden sich mit keinem geprüften Schlüssel (`LISTEN` ist kein Präfixtreffer).
  * Wahrheitswerte: Form (einmal) oder Widerspruch (nur bei zwei gültigen Werten); `UPSTREAM_TLS_*` wird nur in `liesTLS` gelesen.
  * Dauer, Größe, Aufzählung, `LISTEN`, `SERVICE_ID`, `SERVICE_ROOT_PATH`: je ein `melde` pro Pfad.
  * Listen: `liste` meldet leeren Eintrag oder `none` und liefert dann `nil`, `pruefeEintraege` meldet höchstens einmal (Rücksprung); `none,Bad Header` meldet nur `none`.
  * `UPSTREAM_URL`: „erforderlich“ oder „keine absolute URL“, exklusiv; bei `SERVICES_FILE` „zugleich gesetzt“ und der Einzeldienst wird nicht gelesen.
  * `SERVICES_FILE[i]`: „kein Objekt“ endet mit Rücksprung, „unbekanntes Feld“ nur ohne diesen; `.id`/`.service_root`: `pflichtfeld` ist ein Switch, „doppelt“/„überlappt“ setzt einen gültigen, nichtleeren Wert voraus; Typfehler überspringt die Folgeprüfung.
  * `DATA_DIR`: NUL-Zeichen, oder (nur ohne andere Verstöße) Anlegefehler, exklusiv.
  * `UPSTREAM_CA_FILE`: Switch.
  Ergebnis: kein weiterer Fall mit zwei Gründen je Schlüssel gefunden; die Aussage hält. Grenze: Prüfung per Lektüre, kein Fuzz.
* Fänger der Aussage (Plan Zeile 124: `TestKonfigDienstefeldTyp`, `TestKonfigTLSUngueltig`) decken die zwei in der Spezifikation genannten Fälle; für die übrigen Schlüssel trägt sie die Code-Struktur (Switch/Rücksprung), nicht ein eigener Test. Akzeptiert, im Plan als Negativ der Sortierung nach Grund benannt.

## (3) Raster §6

Neun Zeilen × sechs Spalten, keine leere Zelle; jede Zelle nennt `SPEC-110`/`SPEC-109` Nr. oder „gilt nicht“ mit Grund. Die Zellen zu falschem Typ und ungültigem Wahrheitswert stimmen mit Code und Spezifikation überein.

## (4) Zahl in §4

`git diff --numstat 89db02c..HEAD` über `src/hexagon/model src/hexagon/ports/outbound src/adapters/driven/sqlite src/bootstrap spec/spezifikation.md`: Einfügungen 2357 + Löschungen 12 = **2369**. Ohne `_test.go` und ohne Spezifikation: **1103**. Beide Zahlen stimmen mit §4 überein. Marke 800 überschritten, Begründung und Metrik stehen (`AGENTS.md` §3.14). Die Zahl wandert mit jedem weiteren Commit; die Closure-Notiz sollte den letzten Commit nennen.

## (5) Kommentare (§3.7, §3.11)

Suche nach Konjunktiv und Chronik-Wörtern in `konfig*.go`, `model/fehler*.go`, `outbound/abbildung*.go`, `sqlite/fehler_test.go`: drei Treffer, alle Indikativ über den Zustand (`konfig_test.go:27` „spätere Einträge überschreiben frühere“, `:748` „obwohl [2] mit [0] und [1] kollidiert“, `fehler_test.go:101`). Kein Kommentar beschreibt eine verworfene Alternative. Zusagen im Kopf: `liesTLS`, `dateiEintrag`, `TestKonfigDienstefeldTyp` (nennt „hybrid ist nicht geprüft“), `TestKonfigTLSUngueltig` entsprechen dem, was die Tests fangen.

## (6) DoD

| Punkt | Stand |
|---|---|
| Fehlermodell, Abbildung, Test über alle Port-/Adapter-Fehler | erfüllt (`TestAbbildungVollstaendig`, `TestSqliteFehlerVollstaendig`) |
| Konfiguration, Defaults, `ODR-E2001` | erfüllt, Nr. 1 bis 13 im Code wiedergefunden |
| Tests der Randformen §6 | erfüllt |
| `make gates` grün | selbst gefahren, Exit 0 |
| Review mit Report | Report und Re-Review liegen vor; dessen MEDIUM N-1/N-2 sind behoben (HEAD 2baecef), ein dritter Lauf dafür liegt nicht vor |
| Doku-Update | Spezifikation vor dem Code bzw. mit ihm geändert; Historie-Zeile steht; `docs-check` grün |
| Closure-Notiz, Register, Risiko-Ausgänge, Häkchen | offen |

## Befund

### L-1 LOW — Der Start-Befehl hat keine benannte Adresse (`AGENTS.md` §3.13)
§1 schließt den Start-Befehl aus („gehört nicht hierher“), nennt aber keinen Folge-Slice. `slice-bootstrap-m2` (open) trägt im Ziel „verdrahtet Konfiguration“, übernimmt `LadeOderMelde` und Exit-Code 1 aber nicht ausdrücklich in §1 oder DoD. Bis zur Closure ist entweder §1 von `slice-fehlermodell-konfiguration` um die Kennung zu ergänzen und die Sendung in `slice-bootstrap-m2` einzutragen, oder der Punkt geht an den Planner. Bis dahin ist „Start scheitert“ nur auf Funktionsebene belegt (V-H-3 des ersten Berichts).

## Offen für die Closure

1. §7 Closure-Notiz mit Steering-Loop-Lerneintrag füllen (`liegt in` nur, wenn etwas verkörpert wurde). Kandidat: zweite Wiederkehr „Zusage mit Moduszusatz, Fänger deckt einen Modus“ und „Vertragsaussage über Beobachtbarkeit falsch“; Schwelle 3× prüfen gegen das Register.
2. Beobachtungs-Register `docs/plan/planning/observations/` fortschreiben (`BEO-REPO/randform-erst-im-review-entschieden` hat hier weitere Belege: Randformen im Review F-2, V-1, N-2) oder „keine Beobachtung“ in §7 notieren; keinen Zähler setzen.
3. DoD-Häkchen in §2 setzen (alle zehn Punkte).
4. Risiko-Ausgänge in §6: drei stehen auf „weiter offen“ (Abbildung/SQLite-Adapter, Setzungen JSON und 1 GiB, `ErrRevisionVeraltet`); je einen Ausgang eintragen (der Plan nennt „bis zur Closure“).
5. L-1 auflösen (Adresse für den Start-Befehl, `AGENTS.md` §3.13).
6. Zahl in §4 mit dem letzten Commit des Slice neu messen; erst Inhalt, dann reiner `git mv` nach `done/` (`AGENTS.md` §3.3), Roadmap und Verweise (`make slice-mv`) nachziehen; Folge-Slices mit Start „liegt in `done/`“ bleiben dann auflösbar.
7. Entscheidung des Planners, ob ein dritter enger Review für N-1/N-2 nötig ist; die Fänger sind benannt und vorhanden, ihr Rot-Sehen bleibt Behauptung des Implementers.
