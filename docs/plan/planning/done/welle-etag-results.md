# Welle welle-etag — ETags, bedingte Änderungen und Prefer (Soll) — Closure-Notiz

> **Zitier-Form** *(bleibt stehen — Norm, kein Ausfüll-Hinweis).* Dieses
> Artefakt friert ein; was es zitiert, bewegt sich weiter. Deshalb: **Kennung,
> nicht Adresse** — `slice-<Kennung>` statt seines Lifecycle-Pfads, `make <target>`
> statt eines Links auf die Sensor-Datei, eine Baseline-Stelle als
> `v<X.Y.Z>` · `regelwerk/<datei>.md` §<Abschnitt> statt als Link
> (Baseline-Regelwerk `grundlagen-harness-dateien.md` §harness/README.md als
> Einstiegspunkt). Adoptierter Tag: `v6.18.1`.

**Welle:** welle-etag
**Abschluss:** 2026-11-18
**Verantwortlich:** pt9912

## Was wurde geliefert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — *was gelernt wurde*: geliefert · was
funktionierte · was anders lief. Mit ID-Bezug, wo es einen gibt.

Drei Slices der Welle, alle in `done/`: `slice-etag-versionen`, `slice-if-match`, `slice-prefer-return`. Die Welle trägt keinen Meilenstein (Soll-Umfang, nicht Teil der Erstabnahme).

- **ETag aus dem Versionszähler** (`LH-FA-15`, `LH-QA-02`; `slice-etag-versionen`, Vertrag `SPEC-280` bis `SPEC-282`): `W/"<n>"` allein aus `entity_version`, stabil beim Lesen, neu bei jeder Änderung, in der Kopfzeile `ETag` und als `@odata.etag` der Stufe `minimal`; der Schlüssel `ETAG_ENABLED` (Vorgabe aus, nur Modus `mock`, nur Dienste der Version 4.0). Ohne ihn behauptet der Mock keinen Schutz (`LH-FA-15.a`).
- **Bedingte Änderungen** (`LH-FA-15`; `slice-if-match`, Vertrag `SPEC-283`): `If-Match` und `If-None-Match` bei `PATCH`, `PUT` und `DELETE` mit schwachem Vergleich; ein veralteter Wert ergibt `412` mit dem Code `ODR-E6010` (41. Code) ohne Änderung von Zustand, Zähler und Revision; Vergleich und Schreiben liegen in derselben Workspace-Transaktion, von 20 gleichzeitigen Änderungen mit demselben `If-Match` gelingt genau eine.
- **`Prefer`** (`LH-FA-07.b`, Zeile Header in `SPEC-049`; `slice-prefer-return`, Vertrag `SPEC-285` bis `SPEC-289`): `return=minimal` ergibt `204` mit `OData-EntityId` (bei `POST` mit `Location`) und `Preference-Applied`, `return=representation` einen Body byte-gleich dem folgenden `GET`; die Auswertung der Kopfzeile ist eine reine Funktion `model.LiesRueckgabe` über der ganzen Grammatik.
- **Handbuch** (`LH-QA-06`): die Zeile `ETAG_ENABLED` der Einstellungen und die Zeile `ODR-E6010` der Fehlercodes stehen im Benutzerhandbuch; das Kataloge-Gate ist grün.
- **Rundlauf des Closure-Kriteriums** (Welle-Datei §3, dritter Punkt): kein einzelner Test fährt alle Schritte hintereinander; die Schritte tragen benannte Tests, die in dieser Closure frisch gefahren sind (Abschnitt *Verifikation*).

## Was hat funktioniert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3.

- Die Reihenfolge `slice-etag-versionen` → `slice-if-match` mit den Entscheidungen des Architect vor dem Code (Form des ETag, schwacher Vergleich, Rang der Prüfungen, Neuanlage nach `DELETE`) samt Raster und Tabelle *Zusage · Fänger*: in keinem der drei Slices entschied der Code eine Randform; die zwei offenen Zellen von `slice-if-match` (`Workspaces.Get` mit 2, Rang von `ODR-E9001`) kamen als Rückgabe und wurden vom Architect entschieden.
- Wirkungs-Fänger mit Haltepunkt am Fake für Zusagen über Nebenläufigkeit (Zähler unter Gleichzeitigkeit, die vier Fenster der Transaktion): in jedem Prozess rot (20 von 20 und 10 von 10, vom Verifier nachgemessen), weil `make test-race` die Mutationen nicht sieht (`AGENTS.md` §3.10). Der SQLite-Test über die echte Transaktion steht ausdrücklich als zweite, probabilistische Absicherung (11 bis 19 von 20), nicht als Fänger der Zusage.
- Der Fänger über die Klasse „Backslash an jeder Stelle der Grammatik“ in `slice-prefer-return` ist am Ende erschöpfend: jede Zeile über dem Alphabet der Grammatik, 6.742.207 Zeilen in drei Mengen, Orakel aus regulären Ausdrücken ohne Code des Lesers; die Nachverifikation fand an neun eigenen Mutationen kein ungefangenes Element.
- Die bedingte Sendung der siebenten Zählstelle (`policy`, 41 Codes) stand in §3 und §6 von `slice-if-match` und fing den Merge: vier Tests von `policy` wurden rot, die Zahlen wurden laut Plan nachgezogen.

## Was ging anders als geplant?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — jede Zeile möglichst mit der Konsequenz,
die daraus schon gezogen wurde (Folge-Slice, Spec-Version).

- **Vier Review-Runden in `slice-prefer-return`** statt einer. Die Runden 2, 3 und 4 fanden je eine Achse der Klasse „Backslash an jeder Stelle“ (`SPEC-285` Nr. 2), die der Fix der Vorrunde nicht führte: Folgen von Backslashes im Paar eines fremden Elements (Runde 2), der Backslash hinter dem Schlusszeichen eines Paares (Runde 3), das Zeichen vor der Folge und der Kontext des Paares (Runde 4). Ursache jedes Mal: die Menge war aus den Beispielen des Befunds gebildet, nicht aus dem Alphabet der Grammatik. Konsequenz: erschöpfende Aufzählung im Fix der Runde 4; die Regel dazu stand in `AGENTS.md` §3.17 und trug nicht, ihre Bauart bei einer unbeschränkten Klasse steht seither dort (siehe *Steering-Loop-Einträge*).
- **Das Handbuch deckt HTTP-Verhalten nicht, Weg B** (Entscheidung des Architect): `SPEC-200` Nr. 1 lässt Verhalten im Handbuch nur als gebundenes Beispiel zu, und ein HTTP-Request ist in dem Format nicht ausdrückbar. Das gebundene Beispiel zu `ETAG_ENABLED` strich der Architect aus `slice-etag-versionen` (Zeile E24) und aus der Sendung an `slice-handbuch-mock`; ETag, `If-Match` und `Prefer` haben deshalb kein Kapitel, ihr Verhalten tragen `SPEC-280` bis `SPEC-289` und die Tests. Das Closure-Kriterium der Welle-Datei wurde bei `slice-prefer-return` darauf nachgezogen (Handbuch deckt Schlüssel und Fehlercode). Die Prosa der Zeile `ETAG_ENABLED` im Handbuch beschreibt Verhalten, das kein Gate prüft (`SPEC-200` Nr. 14 b); die Zeile auf die Einstellung zu kürzen ist `slice-handbuch-etag-zeile`.
- **Eine Kürzung nahm der Zusage ihren Träger** (`slice-etag-versionen`, V1): die Fix-Runde kürzte die Handbuch-Zeile um den Satz zu `If-Match`, `SPEC-280` Nr. 3 und E24 sagten ihn weiter zu; kein Sensor schlug an. Der Planner strich den Halbsatz in beiden und führte die Aussage in der Sendung. Gezählt unter `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`.
- **Die Nullprobe von `tools/harness/mutationslauf.sh` trägt `src/bootstrap` nicht** (Paketlaufzeit über der festen Grenze): die Mutationen von `slice-etag-versionen` und `slice-if-match` liefen mit eigenen Treibern. Gezählt unter `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (4×, siehe unten).
- **Fremde Zeilen in der Messung** (`slice-if-match`): 800 Zeilen der `policy`-Tests von `slice-fallback-policy-entscheidung` zählten nach dem Merge in den Block (`SPEC-128` Nr. 10 g); neues Verzeichnis `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (1×), der Plan nennt künftig die eigene Zahl neben der Messung.
- **Die Laufzeit eines Tests unter dem Detektor nähert sich der Zeitgrenze** (`slice-prefer-return`, N-V1): `TestLiesRueckgabeErschoepfend` 133 s bei 4 CPUs, 352 s bei 1 CPU, Grenze 600 s; neues Verzeichnis `BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (1×).
- **Der Fall `412` im Modus `hybrid`** (Welle-Datei §3, vierter Punkt) hat in dieser Welle keinen Test: `welle-hybrid` liegt nicht in `done/`, die Welle schloss zuerst. Konsequenz: die Sendung steht im M4-Test von `welle-hybrid` (§3) und in der DoD von `slice-bootstrap-m4`, Geber `slice-if-match` F28. Der Fall „`PATCH` mit `If-Match` in einem Changeset“ liegt bei `slice-batch-changesets` (§1 *Übernimmt*), weil `welle-etag` vor `welle-expand-batch` schloss.

## Messpunkte der Größenmarke

Regeln dieser Sektion: `SPEC-128` Nr. 9 — Auslastung gleich Messung am Ende geteilt durch die Marke; über 1,00 oder unter 0,60 meldet der Planner dem Architect. *S* die Schätzung nach `SPEC-128` Nr. 1, *E* die Messung am Ende nach `git diff --numstat`, aus den Abschnitten *Größe* der Pläne.

| Slice | *S* | Marke | *E* | *E* ÷ *S* | *E* ÷ Marke |
|---|---|---|---|---|---|
| `slice-etag-versionen` | 2190 | 3000 | 1999 | 0,91 | 0,67 |
| `slice-if-match` | 2460 | 3400 | 2454 (eigene Zeilen 1654) | 1,00 (eigen 0,67) | 0,72 (eigen 0,49) |
| `slice-prefer-return` | 3065 | 3500 | 2399 | 0,78 | 0,69 |

Lesart: keine Marke gerissen, keine Auslastung über 1,00; `slice-if-match` liegt ohne den Fremdanteil unter 0,60 und ist dem Architect gemeldet (Messpunkt-Reihe im Plan).
## Steering-Loop-Einträge

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 (hier stehen **nur** Beobachtungen, die im
Register 3× erreicht haben; jeder Eintrag nennt seine `BEO-<KUERZEL>/<slug>`) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Lese-Schritt: 31 Beobachtungen haben 3× erreicht (gezählt: Verzeichnisse mit mindestens drei Dateien in `evidence/`, 75 Verzeichnisse mit 475 Belegen am Tag der Closure); bei der Closure von welle-mock waren es 29. Neu über der Schwelle seit damals: `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (4×, davon `slice-etag-versionen` und `slice-if-match` in dieser Welle) und `BEO-REPO/menge-ohne-ihre-positionen` (3×, aus `slice-hybrid-service-upstream`, nicht aus dieser Welle). **In dieser Closure ist eine Regel geschärft worden** (der Eintrag unten mit Zielort); die zwei neuen Einträge über der Schwelle tragen keinen Zielort, ihr Ausgang steht aus. Eine neue Hard Rule oder eine Änderung an einem Gate wurde nicht vorgenommen ([ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) steht auf `Proposed`; `AGENTS.md` §3.5, §3.6).

- **`AGENTS.md` §3.17 geschärft:** bei einer unbeschränkten Klasse ist die Menge des Fixes das Alphabet der Grammatik bis zur Prüfgrenze (Länge), das zu prüfende Zeichen ein Symbol, Zeichen- und Strukturmenge getrennt, das Orakel ohne Code des Lesers — liegt in `AGENTS.md §3.17`. Auslöser: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (53×, in dieser Welle `slice-etag-versionen`, `slice-if-match`, `slice-prefer-return`). Die Entscheidung nach dem Register: Der Kandidat aus `slice-prefer-return` §7 gehört zu dieser Beobachtung, deren Schwelle 3× längst überschritten ist, und die Bauart selbst ist dreifach belegt: der Fix nahm das Alphabet der Klasse je Stelle in `slice-odata-json-schreiben` (alle 128 ASCII-Bytes je Stelle, Evidenz und `state.md` der Beobachtung), in `slice-replay-service` (256 Bytes je Stelle, 196 608 Elemente, Evidenz von `BEO-REPO/universalaussage-im-vertrag-ungeprueft`) und in `slice-prefer-return` (6.742.207 Zeilen in drei Mengen, Orakel aus regulären Ausdrücken, drei Review-Runden in Folge vor dem Fix). Die Prosa ersetzt nichts: der Zähler zeigt, dass §3.17 die Wiederkehr der Lücke nicht hielt; die Schärfung nennt die Bauart, die in der dritten Runde hielt. **Kürzung im selben Diff** (die Grenzen der Guide-Datei sind ein Gate, [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md)): in `AGENTS.md` §1 entfielen zwei Wiederholungen von Angaben, die `harness/conventions.md` §Adoptierte Konventions-Quellen führt (Regelwerk und Templates parallel entpackt, `SHA256SUMS`, Verweis auf `modul-02-harness-bootstrap.md`; „netzlos, weil parallel“) und der Satz „Derivativ: bei Konflikt gelten die kanonischen Quellen“ (derselbe Inhalt steht im ersten Absatz dieser Datei); `AGENTS.md` steht danach auf 326 Zeilen und 23633 Bytes (vorher 329 und 23648), die Grenzen in `.d-check.yml` stehen auf diesem Ist (Senken ist frei, `SPEC-040` Nr. 8). Keine Regel wurde gelockert oder entfernt.
- **Ausgang bleibt, Zähler gewachsen** (Beobachtungen mit Belegen aus dieser Welle; Retirement-Check: keine Regel wurde entfernt oder gelockert). Die Spalte *Ausgang* nennt den Stand aus `state.md`.

| Beobachtung | Belege ohne diese Welle → jetzt | Belege dieser Welle | Ausgang |
|---|---|---|---|
| `fix-deckt-die-genannte-mutation-nicht-die-klasse` | 50 → 53 | `slice-etag-versionen`, `slice-if-match`, `slice-prefer-return` | verkörpert, `AGENTS.md` §3.17 (geschärft seit welle-etag), `implement-slice`, `verifier.md`; Werkzeug `tools/harness/mutationslauf.sh` (seit welle-record-replay) |
| `vertrag-sagt-mehr-zu-als-ein-gate-prueft` | 59 → 61 | `slice-etag-versionen`, `slice-if-match` | verkörpert, `AGENTS.md` §3.15 und Reviewer-Skill (seit welle-record-replay) |
| `unbeschraenkte-menge-ohne-benannte-pruefgrenze` | 15 → 16 | `slice-if-match` | verkörpert, `implement-slice` Schritt 14, Reviewer-Skill (seit welle-record-replay) |
| `betriebsregel-fuer-repo-dateien-verletzt` | 17 → 18 | `slice-prefer-return` (Heredoc und `sed -i` an einem Report) | geplant in `slice-guard-schreibwege` (`open/`) |
| `nullprobe-grenze-unter-der-paketlaufzeit` | 2 → 4 | `slice-etag-versionen`, `slice-if-match` | **offen, Schwelle erreicht**, kein Zielort; Adresse: Architect (Lesart von `SPEC-210` Nr. 5 a und 9 c, Zeitgrenze der Nullprobe), kein Nehmer-Slice |
| `menge-ohne-ihre-positionen` | 3 → 3 | keiner | **offen, Schwelle erreicht**, kein Zielort; nahe an `AGENTS.md` §3.15 und §3.17 und am Reviewer-Eintrag *Menge ohne Achsenliste*, die verkörpert sind; ob die Lücke dort schon gedeckt ist, entscheidet der User |
| unter der Schwelle, neu in dieser Welle: `bedingter-blockpfad-zaehlt-fremde-zeilen` (1×), `paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (1×) | – | `slice-if-match`, `slice-prefer-return` | offen; Sichtung bei `/plan-welle` |

Die übrigen 24 Beobachtungen über der Schwelle sind in dieser Welle ohne neuen Beleg geblieben; ihr Ausgang steht unverändert in der `state.md` je Verzeichnis (zuletzt gelesen in der Closure von welle-mock).

- **Benannte Spec-Lücke, schon geführt:** Verhalten über HTTP lässt sich im Handbuch nicht als gebundenes Beispiel belegen (`SPEC-200` Nr. 1, Prosa über Verhalten bleibt ungeprüft nach Nr. 14 b); sie steht in der versionierten Spezifikation, hier nur gezählt. Bezug `LH-QA-06`; Adresse User, weiter offen (kein neues Format ohne Entscheidung).

## Beobachtungs-Register (Zeiger)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register — der Zähler wird **nicht** hier gepflegt; diese
Sektion ist ein Zeiger und trägt keine Daten.

Der Zähler steht in [`../observations/`](../observations/).
Was in dieser Welle **3×** erreicht hat, steht oben unter
*Steering-Loop-Einträge*.

## Folge-Slices

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — **derivativ**: Diese Liste zeigt nur,
das Original ist die Slice-Datei. Jeder genannte Folge-Slice muss als Datei im
Planning-Lifecycle existieren; genannt ohne angelegt ist dieselbe Klasse wie
ein halluziniertes Gate.

- `slice-handbuch-etag-zeile` (`open/`, `welle-benutzerhandbuch`, Linie C): die Zeile `ETAG_ENABLED` der Einstellungen im Handbuch auf die Einstellung kürzen; Geber `slice-prefer-return`.
- `slice-bootstrap-m4` (`open/`, `welle-hybrid`): trägt den Fall „`412` im Modus `hybrid` löst keinen Fallback aus“ im M4-Test (DoD Punkt 2); Geber `welle-etag` §3, Fall F28 aus `slice-if-match`.
- `slice-batch-changesets` (`open/`, `welle-expand-batch`): trägt den Fall „`PATCH` mit `If-Match` in einem Changeset“ in der Tabelle *Zusage · Fänger* (§1 *Übernimmt*); Geber `welle-etag`.
- `slice-guard-schreibwege` (`open/`, ohne Welle, Linie C): Träger von `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`; unverändert.
- Kein Slice, ein Auftrag: die Lesart von `SPEC-210` Nr. 5 a und 9 c für die Zeitgrenze der Nullprobe (Adresse Architect), Träger `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`.

## Verifikation

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 1 — keine Behauptung ohne nachprüfbaren
Anker (Hash, Lauf, Zahl).

Stand der Läufe: `main` auf `a9287eb9`, Zweig `welle-etag-closure`, Arbeitsbaum ohne Änderung an Code, Spezifikation und Gates (die Änderungen dieser Closure betreffen Planungs-Dokumente). Alle schweren Läufe unter `flock` auf der Sperre der Arbeitsbäume, nacheinander. Last des Hosts während der Läufe: Loadavg über eine Minute 40 bis 217 bei 3 CPUs durch fremde Läufe; die Läufe messen keine Leistung.

- **Alle Slices der Welle in `done/`:** 3 von 3 Dateien `slice-<Kennung>.md` der Tabelle in §4 der Welle-Datei liegen unter `done/` (`slice-etag-versionen`, `slice-if-match`, `slice-prefer-return`); Verifikation je Slice bestanden, bei `slice-prefer-return` mit Nachverifikation.
- **`make gates`:** Exit 0. Aus dem Log: `a-check` 0 Befunde, `baseline-verify` v6.18.1 OK (54 Dateien), `docs-check` 917 Dateien ohne Befund, `groessenmarke` „kein Slice in Arbeit“, `migration-check` ok, `zeichen-check` ok (1906 Dateien), die Schema-Prüfung ohne Befund, `lint` 0 issues. Die Stufen `test` und `race` kamen aus dem Docker-Zwischenspeicher (Quellen gleich dem letzten Bau; die Prüfsumme über `src/` und `tests/` im Image `odata-recorder:test` gleicht der des Arbeitsbaums), die Laufzeit der Tests ist damit nicht neu gemessen; deshalb die frischen Läufe unten.
- **Pflichtlauf neben der Gate-Kette** (`close-welle` Schritt 1): gedeckt, denn `make test-integration`, `make test-determinism` und `make test-contract` hängen an `GATE_CHECKS` und liefen innerhalb von `make gates`: `test-integration` neun Stufen grün (`TestE2EProzess` bis `TestE2EMockNeustart`), `test-determinism` „Läufe: 100, abweichend: 0“ (`TestE2EDeterminismus`), `test-contract` `/out/bootstrap.test` und `/out/integration.test` grün.
- **Closure-Kriterium §3, dritter Punkt (Rundlauf im Modus `mock`), frisch gefahren** (`go test … -count=1 -v` im Image `odata-recorder:test`, `--memory 3g`, ohne Netz; Exit 0, 0 `--- FAIL`): zweimal Lesen mit gleichem ETag, ETag und Zähler nach Änderungen: `TestEtagEinLesen`, `TestEtagMutationen`, `TestEtagAusUndAn` (Paket `mock`); Änderung mit aktuellem `If-Match`, neuer ETag sichtbar, dieselbe Änderung mit dem alten ETag `412` und `ODR-E6010` ohne Zustandsänderung: `TestIfMatchHttpEndeZuEnde` (echtes HTTP, echte SQLite-Transaktion), `TestIfMatchPasst`, `TestIfMatchVeraltet`, `TestIfMatchStern`; `Prefer: return=minimal` und `return=representation` für `POST`, `PATCH` und `PUT`: `TestPreferEntityGleichGet` (36 Fälle), `TestPreferMinimalEntityId` (6), `TestPreferWiederholungGleichGet` (8), Dienst `TestPrefer*` (Paket `mock`, 14 Tests); die Zeilen der Matrix `SPEC-049` mit ETag: `TestAbnahmeMatrixMitEtag` (36 IDs, 92 Schritte, PASS in 7,17 s); Reset und Restore stellen den ETag her: `TestEtagResetStelltHer`, `TestEtagRestoreStelltHer`. Hybrid: `TestPrefer*` im Paket `hybrid` (3 Tests, `Prefer` geht unverändert in Fallback und Upstream).
- **Closure-Kriterium §3, vierter Punkt (`412` im Modus `hybrid`):** nicht anwendbar, `welle-hybrid` liegt nicht in `done/`; die Sendung steht bei `slice-bootstrap-m4` (siehe *Folge-Slices*).
- **Closure-Kriterium §3, fünfter Punkt (Handbuch):** Schlüssel `ETAG_ENABLED` und Fehlercode `ODR-E6010` stehen im Benutzerhandbuch; `make handbuch-check` und `make katalog-check` sind Teil des grünen `make gates`.
- **`make gegenproben`: nicht grün, kein gültiger Lauf — das Kriterium „`make gegenproben` ist grün“ (Welle-Datei §3, zweiter Punkt) ist in dieser Closure nicht belegt.** Der Stempel gehört zum Stand nach `slice-if-match`, `src/` ist Eingabe und änderte sich danach (`slice-prefer-return`), der Lauf war also nötig und lief zweimal bei geänderter Eingabe: beide rot mit 9 von 216 Fällen, beide Male dieselben neun Fälle `lint/basis`, `lint/direktive-nolint`, `lint/ausnahme-ohne-why`, `lint/ausnahme-mit-why`, `lint/direktive-mit-leerzeichen`, `lint/why-falsch-eingerueckt`, `lint/zweites-modul-sauber`, `lint/befund-im-zweiten-modul`, `lint/modul-ohne-go-datei` mit „Zeit überschritten“ (Frist je Fall 120 s, `fallFrist` in `tools/gegenproben/main.go`); 207 Fälle ok, darunter alle Fälle zu `a-check`, `docs-check` (einschließlich der Ausnahme `.wt/**`), `groessenmarke`, `katalog`, `zeichen`, `migration`, `test`, `build` und `handbuch`. Laufzeit des Schritts 131 s und 124 s, des ganzen Aufrufs 151 s und 144 s. Last des Hosts (Loadavg über eine Minute, 3 CPUs): 40 und 17 beim Start, 217 und 181 am Ende, durch fremde Läufe anderer Arbeitsbäume. Der Code änderte sich zwischen den Läufen nicht; der Stempel blieb unberührt (er wird nur bei Grün geschrieben). **Wertung:** [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) macht einen Lauf außerhalb der Lastgrenze ungültig statt rot und verbietet das Wiederholen bis grün; die Lastgrenze steht noch nicht in `SPEC-102` (sie liefert `slice-gegenproben-zeitbudget`, `open/`), also gibt es für diesen Host keinen Maßstab für „gültig“. Die Ursache (Last oder ein Fall, der auch ohne Last über 120 s braucht) ist nicht belegt; ein Rot der Gruppe `lint` ohne Änderung an `tools/harness/lint.sh`, am Profil oder am Code der Gruppe ist ein Befund für den Planner und den Architect, kein Anlass, die Frist zu ändern (`AGENTS.md` §3.6). Das Kriterium bleibt offen, Adresse: User (Lauf auf ruhigem Host, Eingabe für `slice-gegenproben-zeitbudget`).
- **Trigger-Audit, Carveouts:** `docs/plan/carveouts/` führt keine Datei; 0 offen, keine Welle schließt mit einem Carveout, kein Gate ist rot.
- **Trigger-Audit, bootstrap-aware Gates:** kein Gate trägt eine Reifestufe; die Schwellen sind feste Grenzen (Größengrenzen der Guide-Dateien nach [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md), Größenmarke nach `SPEC-128`). Keine Verschiebung in dieser Closure, `AGENTS.md` und die übrigen Guide-Dateien blieben unberührt. 0 offen.
- **Trigger-Audit, Entscheidungen** (Re-Evaluierungs-Trigger je ADR gegen den Stand der Closure):
  - [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0002](../../adr/0002-architektur-gate-a-check.md), [ADR-0003](../../adr/0003-lint-profil-golangci.md), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md): `a-check`, `lint` und `lint-profil` grün, kein Anlass; die Baseline nennt keinen neuen Ort für Werkzeugverträge (Pin `v6.18.1`, `baseline-verify` OK). Nicht eingetreten.
  - [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md): der Trigger „Lauf bei geänderter Eingabe über 30 s“ ist seit der Closure von welle-record-replay eingetreten und von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) (`Accepted`) beantwortet; umgesetzt wird er von `slice-gegenproben-zeitbudget` (`open/`). Stand in dieser Closure siehe `make gegenproben` oben. Die Gegenproben bleiben außerhalb der Kette.
  - [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): `TestIfMatchHttpEndeZuEnde` lief frisch über die echte Transaktion (siehe oben), `TestEtagGleichzeitigSQLite` und `TestIfMatchWettlaufSQLite` sind Teil des grünen `make test`; keine Auffälligkeit der Bibliothek. Nicht eingetreten.
  - [ADR-0007](../../adr/0007-race-detektor-im-gate.md): `make test-race` ist Teil des grünen `make gates`; kein Rennen gemeldet. Der Trigger nennt 10 Minuten Paketlaufzeit unter dem Detektor: der längste Test der Welle, `TestLiesRueckgabeErschoepfend`, braucht 352 s bei einer CPU (N-V1), unter der Grenze; er ist als `BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (1×) geführt, Adresse der Planner beim nächsten Slice, der `src/hexagon/model` oder die Mengen des Tests ändert. Nicht eingetreten, nahe.
  - [ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md): kein neuer Schreiber von Blobs in dieser Welle (ETag, Bedingungen und `Prefer` schreiben Entities). Nicht eingetreten.
  - [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md): `make groessenmarke` ohne Slice in Arbeit grün; die Messungen der drei Slices lagen alle auf oder unter ihrer Marke, das Gate war an keinem Stand ohne Fehler rot. Status bleibt `Proposed`, die Annahme ist Entscheidung des User (offen, wie bei der Closure von welle-mock).
  - [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md): keine Verschiebung der Grenzen in dieser Welle; nicht eingetreten.
  - [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md): von dieser Welle nicht berührt (kein V2-Konstrukt, keine neue Kante eines Prüfpakets). Nicht eingetreten.
  - [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md): Trigger „Frist je Fall reißt in einem gültigen Lauf“: siehe `make gegenproben` oben.
- **Archivierung:** nicht ausgeführt, die Start-Bedingung ist nicht eingetreten. `make archive-welle WELLE=welle-etag` endet mit Exit 0 und der Meldung „der Traeger liegt nicht (`.harness/state/bin/ai-harness-init`) — dieses Repo archiviert seine Wellen nicht“; der Träger liegt im gitignorierten Zustands-Bereich und fehlt in diesem Arbeitsbaum und im Hauptbaum, ein erneuter Lauf des Werkzeugs legt ihn wieder ab. Ein `done/*/archiv.zip` existiert (`altbestand`, `welle-fundament`, `welle-persistenz`, `welle-record-replay`, `welle-mock`), die Vorprüfung `[untergrenze]` sperrt also nicht. Es ist kein Handlauf gefahren worden. Die Zeitdokumente der Welle bleiben flach in `done/` liegen: `slice-etag-versionen`, `slice-if-match`, `slice-prefer-return` und die zehn Reports unter `docs/reviews/` dieser drei Slices; welche wellenlosen Slices seit der Closure von welle-mock das Werkzeug mit einsammelt, sagt erst seine Vorprüfung. Die Archivierung führt, wer den Träger wieder ablegt, nach dieser Closure (Welle-Datei und Notiz liegen dann in `done/`, die Sperren `[ergebnisnotiz]` und `[kein-plan]` entfallen); der Geltungsbereich der Sensoren ist vorher zu prüfen (was auf `done/*.md` keilt, sieht die Stubs unter `done/<welle-id>/` nicht; die Link- und Kennungspflichten des Doku-Gates gelten im Stub), er wurde bei den Archivierungen von `welle-mock` und davor belegt und ist in dieser Closure nicht neu gemessen.
- **Drei Paarungen** (nach dem Schreiben dieser Notiz): *Anker*: ein Eintrag trägt das Zielort-Feld, `AGENTS.md §3.17`; die Datei existiert, und die Überschrift des Abschnitts trägt `seit welle-etag` (`### 3.17 … (seit welle-record-replay, geschärft seit welle-etag)`). *Folge-Slice*: `slice-handbuch-etag-zeile`, `slice-bootstrap-m4`, `slice-batch-changesets` und `slice-guard-schreibwege` liegen als Dateien in `open/`. *Register*: die genannten Kennungen `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse`, `vertrag-sagt-mehr-zu-als-ein-gate-prueft`, `unbeschraenkte-menge-ohne-benannte-pruefgrenze`, `betriebsregel-fuer-repo-dateien-verletzt`, `nullprobe-grenze-unter-der-paketlaufzeit`, `menge-ohne-ihre-positionen`, `bedingter-blockpfad-zaehlt-fremde-zeilen` und `paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` existieren als Verzeichnisse, und alle 75 Verzeichnisse tragen ein nicht leeres `evidence/`.
