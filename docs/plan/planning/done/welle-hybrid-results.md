# Welle welle-hybrid — Hybrid-Modus mit Fallback-Policy — Closure-Notiz

> **Zitier-Form** *(bleibt stehen — Norm, kein Ausfüll-Hinweis).* Dieses
> Artefakt friert ein; was es zitiert, bewegt sich weiter. Deshalb: **Kennung,
> nicht Adresse** — `slice-<Kennung>` statt seines Lifecycle-Pfads, `make <target>`
> statt eines Links auf die Sensor-Datei, eine Baseline-Stelle als
> `v<X.Y.Z>` · `regelwerk/<datei>.md` §<Abschnitt> statt als Link
> (Baseline-Regelwerk `grundlagen-harness-dateien.md` §harness/README.md als
> Einstiegspunkt). Adoptierter Tag: `v6.18.1`.

**Welle:** welle-hybrid
**Abschluss:** 2026-11-19
**Verantwortlich:** pt9912

## Was wurde geliefert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — *was gelernt wurde*: geliefert · was
funktionierte · was anders lief. Mit ID-Bezug, wo es einen gibt.

Sechs Slices der Welle, alle in `done/`: `slice-fallback-policy`, `slice-fallback-policy-entscheidung`, `slice-hybrid-service`, `slice-hybrid-service-upstream`, `slice-bootstrap-m4-konfig`, `slice-bootstrap-m4`. Die Welle trägt den Meilenstein M4.

- **Fallback-Policy** (`LH-FA-12`, `LH-FA-12.a`, `LH-FA-12.b`; `slice-fallback-policy` und `slice-fallback-policy-entscheidung`, Vertrag `SPEC-261`): Prüfung der Einträge (Host, Routenfreigabe, Kategorien) mit `New` und `*KonfigFehler`, dazu die Entscheidung `Entscheide` nach `ODR`-Code, Methodenklasse, Hauptschalter und Vergleich von Zielhost und Pfadmuster; total, wiederholbar, ein `nil` oder leeres `Policy` antwortet `Bleiben`.
- **`HybridUseCase`** (`LH-FA-12`, `LH-FA-04`, `LH-FA-11`; `slice-hybrid-service` und `slice-hybrid-service-upstream`, Vertrag `SPEC-284`): Mock zuerst, Entscheidung am Code des Fehlers (nie am HTTP-Status), Replay-Fallback mit der Kennzeichnung `X-Recorder-Fallback: replay`, Upstream-Fallback nur nach Freigabe der Policy über `UpstreamClient` mit Filter der Header, Kennzeichnung `upstream` und Fehlern `ODR-E4001` und `ODR-E9001` ohne Wiederholen; Pufferung des Bodys nur für fallback-fähige Anfragen bis `MAX_CAPTURE_BODY`.
- **Konfiguration und Verdrahtung** (`LH-FA-12`, `LH-QA-02`; `slice-bootstrap-m4-konfig` und `slice-bootstrap-m4`, `SPEC-110`, `SPEC-139`, `SPEC-037`): die Schlüssel `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` mit jedem `*KonfigFehler` als `ODR-E2001` ohne Wert in der Meldung; der Composition Root bindet im Modus `hybrid` genau `HybridUseCase`; `TestM4` belegt die Matrix am Referenzdienst, der Verbindungen und Requests je Methode zählt; `TestE2EHybrid` ist Stufe 10 von `make test-integration` am gebauten Binary.
- **Handbuch** (`LH-QA-06`): das Kapitel `Modus hybrid` und die Zeilen der Einstellungen (`MODE`, `UPSTREAM_URL`, `STRICT_COVERAGE`, `ETAG_ENABLED`, `PUBLIC_BASE_URL` und die vier Schlüssel) stehen im Benutzerhandbuch; `make handbuch-check` und `make katalog-check` sind grün. Das Kapitel steht als Prosa und Tabelle, ohne Beispiel-Datei (benannte Abweichung unten).
- **M4 erreicht:** `TestM4` (`tests/integration/m4_test.go`, Teil von `make test`) belegt am Verbindungszähler des Referenzdienstes, dass im Modus `hybrid` nur freigegebene lesende Anfragen den Upstream erreichen und ein POST ohne Freigabe nie.

## Was hat funktioniert?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3.

- **Drei Schnitte vor dem Code** (Drift-Log der Roadmap, 2026-11-08 und 2026-11-18): die Policy (Prüfung der Einträge, dann Entscheidung), der Dienst (Mock und Replay, dann Upstream) und die Verdrahtung (Konfiguration, dann Composition Root) liefen als sechs Slices statt dreien, jeder mit Raster und Tabelle *Zusage · Fänger* vor dem ersten Code-Commit. Die Entscheidungen des Architect vor dem Code (Dienstwahl, Start und Herunterfahren im Modus `hybrid`, Stufe 10, `UPSTREAM_CA_FILE` in `hybrid`) standen in `SPEC-261`, `SPEC-284`, `SPEC-139` und `SPEC-037`; die drei Fragen, die der Implementer von `slice-bootstrap-m4` im Code nicht entschied, standen mit Ausgang *offen* und Adresse im selben Commit in §6 und waren am 2026-11-19 entschieden.
- **Orakel ohne Code des Prüflings:** die Sicherheitsprobe der Entscheidung (616 Policies, danach 57,7 Mio. Fälle eines eigenen Generators) fand keine Abweichung in der gefährlichen Richtung; die Probe des Dienstes (187 200 Fälle) keinen Zugriff auf den Upstream ohne Freigabe; der Verbindungszähler des Referenzdienstes trägt die Zusage „ein POST ohne Eintrag in `FALLBACK_WRITE_ALLOW` erreicht den Upstream nie“ am echten Adapter `httpupstream`, nicht an einem Fake (`TestM4/Mutierender_Fallback`).
- **Wirkungs-Fänger mit Haltepunkt für die Zähler** (`slice-bootstrap-m4` F3): die Raten der Zähler unter Gleichzeitigkeit (`TestM4/Gleichzeitig`) sind in §3 des Plans mit Zahl der Prozesse ausgewertet; das Fehlen der Sperre fängt `make test-race`.
- **Die Verifikation fand jeweils, was Review und Plan nicht trugen,** und fuhr Mutationen an Elementen, die keiner nannte: 10 eigene, alle rot (`slice-fallback-policy`), 20 eigene (`slice-bootstrap-m4`: 18 rot, eine äquivalent, ein Übersetzungsfehler), 24 von 24 rot (`slice-bootstrap-m4-konfig`).

## Was ging anders als geplant?

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 — jede Zeile möglichst mit der Konsequenz,
die daraus schon gezogen wurde (Folge-Slice, Spec-Version).

- **Benannte Abweichung beim Closure-Kriterium `412` (Welle-Datei §3, vierter Punkt).** Das Kriterium wurde bei der Closure von `slice-bootstrap-m4` auf „ohne Kennzeichnung, 0 Verbindungen und 0 Requests am Referenzdienst“ gesenkt; die Hälfte „Zustand unverändert“ der Sendung von `welle-etag` (Geber `slice-if-match`, Fall F28) prüft der M4-Test nicht: `TestM4/Ohne_Auslöser_kein_Upstream` (Fall `fachliche 412`) liest die Entity nach dem `PATCH` nicht zurück und fährt mit `If-Match: "nope"`, nicht mit dem ETag einer früheren Version. Die Verifikation fand es als V-1 (LOW), gegen die Zeile M12 der Tabelle *Zusage · Fänger*. *Folge:* die Hälfte trägt allein die Transaktion des Mock mit den Tests von `slice-if-match` (in `done/`, 20 gleichzeitige Änderungen, genau eine gelingt); im Hybrid-Betrieb ist ein Zustand, der sich nach einem `412` ändert, deshalb nicht belegt, sondern aus dem Verhalten des Mock gefolgert. Der Rest (ein Fall in `tests/integration/m4_test.go`, der nach dem `412` zurückliest und mit dem ETag einer früheren Version fährt) hat keinen Nehmer-Slice und steht im Register als `BEO-REPO/zustand-unveraendert-nach-fehler-ohne-ruecklesen` (1×, Adresse Planner, Sichtungs-Schritt der nächsten Planung eines Slice, der `tests/integration/m4_test.go` oder die Fälle des Mock bei `412` berührt). Die Entscheidung des Users lautete: keine Rückgabe an den Implementer.
- **Das Handbuch deckt den Modus `hybrid` ohne gebundenes Beispiel** (Welle-Datei §3, fünfter Punkt, und deren *Pflicht Handbuch*): `SPEC-200` Nr. 6 d und Nr. 7 b kennen im Format der Beispiel-Datei keine Umgebung, `MODE=hybrid` und die vier Schlüssel sind dort nicht ausdrückbar, und ein `start` ohne Umgebung ist der Modus `mock`. Entscheidung des Architect (`SPEC-200` Nr. 6, keine Beispiel-Datei für `hybrid`): Kapitel und Zeilen stehen als Prosa und Tabelle; der Katalog der Schlüssel und Codes ist gegen den Quelltext geprüft (`make katalog-check`), der Inhalt der Sätze nicht (`SPEC-200` Nr. 14 b). *Folge im Betrieb:* ein Satz des Kapitels kann dem Ist-Zustand widersprechen, ohne dass ein Gate rot wird; der Reviewer las den Inhalt gegen `TestM4`. Dieselbe Lücke wie bei `welle-etag` (Verhalten über HTTP, Umgebung), schon in der Spezifikation benannt.
- **Vier Review-Runden in `slice-bootstrap-m4-konfig`** statt einer, jede mit einer Klasse, die der Fix der Vorrunde nicht ganz deckte (Leerraum je Stelle, Position, Modus, Länge, Gegenseite je Zeichen); zwei Fixes erzeugten ein Fail-open (die Konstante `leerraum` statt `TrimSpace` ließ U+00A0, U+0085 und VT im Namen eines Query-Parameters durch). Ursache: die Vergleichsform eines Eintrags (Name gegen den dekodierten Schlüssel, Host gegen eine ASCII-Allowlist) war kein Orakel des Fängers; erst `KonfigRedactionTrifftJedesZeichen` (230 Zeichen, Konfiguration gegen Redaktor) deckte die Klasse in einer Runde. Gezählt unter `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` und neu unter `BEO-REPO/eintrag-trifft-nie-weil-namensform-und-vergleichsform-auseinanderliegen` (1×).
- **Zwei Runden in `slice-hybrid-service-upstream`:** die Zusage „der Dienst schließt den Body der Anfrage nie“ lief als Zeile Z16 von Teil 1 „unverändert weiter“, aber der Rig von Z16 hat kein `Ziel`, der Weg U war nie erreicht; die Mutation „Body vor `Senden` geschlossen“ fand erst der Verifier (V-1). Konsequenz: der Plan des Nachfolgers führt je Zusage des Vorgängers, die der neue Weg trifft, eine Zeile mit dem Fänger, der den Weg erreicht (gezählt unter `BEO-REPO/menge-ohne-ihre-positionen`, Achse Weg zu U). Dazu F-6: die Randform „beide Fehler des Ports zugleich“ stand nicht in der Spezifikation; der Implementer meldete sie in §6 mit Ausgang *offen*, hielt aber nicht an (`AGENTS.md` §3.12); die Entscheidung (`SPEC-284` Nr. 11) kam nach dem Code, der Code stimmte überein.
- **Die Marke von `slice-bootstrap-m4-konfig` riss** (1872 von 900, mit 143 fremden Zeilen aus `main`; eigene 1729): der Grund steht im Plan (Fänger der Klassen nach `AGENTS.md` §3.17, kein vorab benannter Teil, Konstante, Funktion und Fänger teilen den Mechanismus), keine Rückführung. Ende ÷ Handoff 2,58 liegt über der Spanne von `SPEC-128` Nr. 2 und allen Werten des Registers (`BEO-REPO/auslastung-am-handoff-als-endwert-gelesen`, Ausgang *geplant*: Auftrag an den Architect).
- **Eine Marke, die der Planner beim Start nicht setzte** (`slice-hybrid-service-upstream`): sichtbar erst bei der Closure; `make groessenmarke` weist darauf hin. Gezählt unter `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×).
- **Diese Closure selbst:** der Planner zog drei Evidence-Dateien des Registers mit einem `sed -i` auf die Kennung ohne Endung (eine) und sonst mit `Edit` (zwei); das `sed -i` sperrte der Command-Guard nicht (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`, 21×, Träger `slice-guard-schreibwege` liegt in `open/`). Die Vorprüfung von `make archive-welle` sperrte mit `[haenger]` an diesen drei Dateien (`BEO-REPO/verweis-auf-review-report-sperrt-die-archivierung`, jetzt 3×).

## Messpunkte der Größenmarke

Regeln dieser Sektion: `SPEC-128` Nr. 9 — Auslastung gleich Messung am Ende geteilt durch die Marke; über 1,00 oder unter 0,60 meldet der Planner dem Architect. *S* die Schätzung nach `SPEC-128` Nr. 1, *E* die Messung am Ende nach `git diff --numstat`, aus den Abschnitten *Größe* der Pläne.

| Slice | *S* | Marke | *E* | *E* ÷ *S* | *E* ÷ Marke |
|---|---|---|---|---|---|
| `slice-fallback-policy` | 1020 | 1400 | 1311 | 1,29 | 0,94 |
| `slice-fallback-policy-entscheidung` | 1480 | 2000 | 1706 | 1,15 | 0,85 |
| `slice-hybrid-service` | 2820 | 3900 | 2241 | 0,79 | 0,58 |
| `slice-hybrid-service-upstream` | 2220 | 3000 | 2126 | 0,96 | 0,71 |
| `slice-bootstrap-m4-konfig` | 630 | 900 | 1872 (eigene 1729, 143 fremd) | 2,97 (eigen 2,74) | 2,08 (eigen 1,92) |
| `slice-bootstrap-m4` | 1820 | 2500 | 2222 | 1,22 | 0,89 |

Lesart: eine Marke gerissen (`slice-bootstrap-m4-konfig`, Grund im Plan, siehe oben; dem Architect über `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen` und `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` gemeldet, beide mit Ausgang *geplant*); `slice-hybrid-service` liegt mit 0,575 unter 0,60 (dort schon gemeldet, Faktoren neu rechnen). Die Messung der Welle ist die der Pläne in `done/`; sie wurde in dieser Closure nicht neu gefahren, `make groessenmarke` meldet ohne Slice in Arbeit „kein Slice in Arbeit“.

## Steering-Loop-Einträge

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 3 (hier stehen **nur** Beobachtungen, die im
Register 3× erreicht haben; jeder Eintrag nennt seine `BEO-<KUERZEL>/<slug>`) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Lese-Schritt: 34 Beobachtungen haben 3× erreicht (gezählt: Verzeichnisse mit mindestens drei Dateien in `evidence/`; 80 Verzeichnisse mit 504 Belegen am Tag der Closure, bei der Closure von welle-etag waren es 75 mit 475 und 31 über der Schwelle). Neu über der Schwelle in dieser Closure: `BEO-REPO/verweis-auf-review-report-sperrt-die-archivierung` (3×, Beleg dieser Closure). **In dieser Closure ist keine Regel verkörpert worden**: Der Planner schreibt weder `AGENTS.md` noch `.claude/` noch `.harness/skills/`; die Einträge unten tragen deshalb kein Feld `liegt in` und keinen Anker `seit welle-hybrid`, und die Paarung der Anker prüft nichts. Eine neue Hard Rule oder eine Änderung an einem Gate wurde nicht vorgenommen (`AGENTS.md` §3.5, §3.6; [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) steht weiter auf `Proposed`).

Von den 34 tragen 29 einen Ausgang (`verkörpert`, `geplant` oder `behoben`); **fünf stehen ohne Ausgang** (die Zeilen *offen* der Tabelle), die der Planner nicht selbst verkörpert, sondern mit Diff-Vorschlag oder Adresse an den User gibt (Abschnitt *Entscheidungen beim User*). Bei zwei weiteren der sieben Zeilen (`zweig-ohne-lauf-unter-dem-detektor`, `report-der-rolle-faellt-docs-check`) hat die Entscheidung des Users vom 2026-11-08 den Träger schon benannt (`slice-regeln-verkoerpern`, `open/`); ihre `state.md` trägt jetzt den Ausgang *geplant* statt „beim User offen“ (Lese-Schritt: sieben Zeilen ohne Ausgang vor, fünf nach dieser Closure).

| Beobachtung | Belege | Belege dieser Welle | Ausgang |
|---|---|---|---|
| `zweig-ohne-lauf-unter-dem-detektor` | 18 | `slice-fallback-policy-entscheidung`, `slice-hybrid-service` | **geplant** in `slice-regeln-verkoerpern` (`open/`, Vorlage 1 der Closure von welle-mock); nicht verkörpert |
| `report-der-rolle-faellt-docs-check` | 5 | keiner | **geplant** in `slice-regeln-verkoerpern` (`open/`, Vorlage 2); nicht verkörpert |
| `menge-ohne-ihre-positionen` | 6 | `slice-hybrid-service-upstream`, `slice-bootstrap-m4-konfig`, `slice-bootstrap-m4` | **offen**, kein Zielort; Diff-Vorschlag E1 |
| `verweis-auf-review-report-sperrt-die-archivierung` | 3 | `welle-hybrid` | **offen**, kein Zielort; Diff-Vorschlag E2 |
| `nullprobe-grenze-unter-der-paketlaufzeit` | 5 | `slice-bootstrap-m4-konfig` | **offen**, kein Zielort; Auftrag an den Architect E3 |
| `laufzeitmessung-widerspricht-sich-unter-last` | 3 | keiner | **offen**; Diff-Vorschlag im Bericht der Closure von `slice-gegenproben-zeitbudget`, E4 |
| `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` | 3 | keiner | **offen**; Diff-Vorschlag im Bericht der Closure von `slice-gegenproben-zeitbudget`, E4 |

- **Ausgang bleibt, Zähler gewachsen** (Retirement-Check: keine Regel wurde entfernt oder gelockert; keine Verkörperung ist hier berührt):

| Beobachtung | Belege dieser Welle | Ausgang |
|---|---|---|
| `fix-deckt-die-genannte-mutation-nicht-die-klasse` (55) | `slice-fallback-policy`, `slice-fallback-policy-entscheidung`, `slice-hybrid-service`, `slice-hybrid-service-upstream`, `slice-bootstrap-m4-konfig` | verkörpert, `AGENTS.md` §3.17 (geschärft seit welle-etag), `implement-slice`, `verifier.md`; die Regel trug in `slice-bootstrap-m4-konfig` vier Runden nicht (Wiederauftreten, keine Lücke des Wortlauts; Herkunft `welle-record-replay`, die Beobachtung ist seither wieder aufgetreten, die Regel bleibt) |
| `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (62) | `slice-fallback-policy-entscheidung`, `slice-hybrid-service` | verkörpert, `AGENTS.md` §3.15 und Reviewer-Skill |
| `randform-erst-im-review-entschieden` (36) | `slice-hybrid-service-upstream` | verkörpert, `AGENTS.md` §3.12 und §3.16 |
| `berichtigung-erreicht-nicht-jede-stelle` (30), `akzeptiertes-negativ-ohne-folge` (17), `rate-ohne-detektor-schwankt-mit-der-last` (9), `fixture-gleicht-dem-wert-der-mutation` (8) | `slice-bootstrap-m4` | verkörpert (`AGENTS.md` §3.9, §3.15, §3.10) |
| `groessenmarke-ueberschritten-ohne-rueckfuehrung` (20) | `slice-bootstrap-m4-konfig` | verkörpert, `harness/mk/groessenmarke.mk` und `AGENTS.md` §3.14 |
| `unbeschraenkte-menge-ohne-benannte-pruefgrenze` (18) | `slice-hybrid-service` | verkörpert, `implement-slice` Schritt 14, Reviewer-Skill |
| `betriebsregel-fuer-repo-dateien-verletzt` (21) | `slice-bootstrap-m4-konfig`, `welle-hybrid` | geplant in `slice-guard-schreibwege` (`open/`); E5 |

- **Unter der Schwelle, neu in dieser Welle:** `BEO-REPO/zustand-unveraendert-nach-fehler-ohne-ruecklesen` (1×, mit dem Rest), `BEO-REPO/eintrag-trifft-nie-weil-namensform-und-vergleichsform-auseinanderliegen` (1×), `BEO-REPO/marke-beim-start-nicht-gesetzt` (1×); offen, Sichtung bei `/plan-welle`.
- **Benannte Spec-Lücke, schon geführt:** das Format der Beispiel-Datei kennt keine Umgebung (`SPEC-200` Nr. 6 d, Nr. 7 b), Verhalten über HTTP und Betriebsarten mit Schlüsseln sind damit im Handbuch nicht als gebundenes Beispiel belegbar (Prosa bleibt ungeprüft nach Nr. 14 b). Sie steht in der versionierten Spezifikation, hier nur gezählt (zweiter Fall nach `welle-etag`). Bezug `LH-QA-06`; Adresse User, kein neues Format ohne Entscheidung.

### Entscheidungen beim User

Der Planner hat nichts davon umgesetzt. Die Diff-Vorschläge sind Eingabe; die Größengrenzen der Guide-Dateien sind ein Gate ([ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md)), `AGENTS.md` steht auf 326 von 326 Zeilen und 23633 von 23633 Bytes, `implement-slice.md` auf 17081 von 17081 Bytes, `reviewer.md` auf 145 von 145 Zeilen: jede Verkörperung trägt ihre Kürzung im selben Diff (`slice-regeln-verkoerpern` ist der Ort dafür, sein §1 schließt „weitere Beobachtungen über der Schwelle“ aus; ein weiterer Träger ist Entscheidung des Users).

1. **E1 — `menge-ohne-ihre-positionen` (6×): decken `AGENTS.md` §3.15 und §3.17 und der Reviewer-Eintrag *Menge unbeschränkt oder ohne Achsenliste* die Lücke?** Befund des Planners: nein, nicht für den ersten Wurf. §3.17 nennt „jede Position“ nur für den *Fix*; §3.15 verlangt „ein Fall je Element einer Menge“, aber nicht, dass der Plan die Achsen der Menge nennt (Stelle des Elements im Wert, Abstand eines Paares, Modus, Weg zum Aufruf); der Reviewer-Eintrag fängt die Lücke erst im Review. Sechs Vorgänge, drei in dieser Welle (Weg zu U, Position bis 17 in acht Listen, Achse Modus). Vorschlag (nicht verkörpert, Ziel `AGENTS.md` §3.15, nach dem Satz „… der Fall, in dem zwei verglichene Werte gleich sind.“; Anker `seit welle-hybrid` erst bei der Verkörperung; Kürzung im selben Diff nötig, etwa eine der beiden *Falsch/Richtig*-Paare von §3.15 oder §3.14):

   ```diff
   -Eine Fixture gleicht nie dem Wert der Mutation.
   +Eine Menge im Plan nennt ihre Achsen (Stelle des Elements im Wert, Abstand eines Paares, Modus, Weg zum Aufruf); der Fänger führt jede Achse, oder das Negativ nennt die ungeführte. Eine Fixture gleicht nie dem Wert der Mutation.
   ```

   Alternative: *gedeckt* feststellen (Ausgang `gestrichen` mit Begründung ist dafür nicht passend, die Beobachtung tritt weiter auf); dann bliebe nur das Zählen.
2. **E2 — `verweis-auf-review-report-sperrt-die-archivierung` (3×, neu über der Schwelle):** eine Zeile in `.claude/commands/implement-slice.md` Schritt 25 (Beleg-Form) und in den Anweisungen von Reviewer und Verifier, dass ein Report in Evidence-Dateien und Plänen mit seiner Kennung ohne Pfad und ohne Endung steht. Vorschlag (nicht verkörpert, Ziel `implement-slice.md` Schritt 25, hinter „kein Freitext, eine Datei je Auftreten.“; Kürzung im selben Diff, die Datei steht auf ihrer Grenze):

   ```diff
   +Ein Review-Report steht in Evidence und Plan mit seiner Kennung ohne Pfad und ohne Endung (`2026-11-18-slice-x-review-2`), nie als Dateiname oder Link: `make archive-welle` sperrt mit `[haenger]`, solange eine Datei den Pfad nennt.
   ```

   Alternative: ein Sensor, der die Form prüft (ein neues Gate ist erlaubt, `AGENTS.md` §3.6), oder die Sperre des Werkzeugs durch ein Ziehen der Verweise im Werkzeug selbst ersetzen (Entscheidung beim Träger des Werkzeugs, `ai-harness-init`).
3. **E3 — `nullprobe-grenze-unter-der-paketlaufzeit` (5×): der Auftrag an den Architect ergeht noch nicht.** Lesart von `SPEC-210` Nr. 5 a und 9 c (feste Zeitgrenze `-test.timeout=2m` der Nullprobe; `src/bootstrap` braucht rund 150 s ohne Detektor); ob die Grenze ein Feld der Liste wird oder aus der Paketlaufzeit folgt, ist eine Vertragsfrage, kein Planner-Entscheid. Zu entscheiden: der Auftrag jetzt (dann folgt ein Kettenslice der Linie C, nicht produktiv) oder weiter offen.
4. **E4 — `laufzeitmessung-widerspricht-sich-unter-last` und `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (je 3×):** die Diff-Vorschläge (a) und (b) stehen unverändert im Bericht der Closure von `slice-gegenproben-zeitbudget` (§7, Steering-Loop-Eintrag); zu entscheiden: verkörpern (Ort: `AGENTS.md` §3.15 und eine Regel nach §3.10, mit Kürzung) oder als geplant einem Träger zuweisen; kein neuer Beleg in dieser Welle.
5. **E5 — `betriebsregel-fuer-repo-dateien-verletzt` (21×, Ausgang *geplant* seit der Closure von welle-record-replay):** der Träger `slice-guard-schreibwege` liegt weiter in `open/`; in dieser Closure schrieb der Planner selbst einmal mit `sed -i` (siehe oben). Zu entscheiden: der Start des Kettenslice (Reihenfolge der Linie C und Nr. 7 der Übergangsregeln; die Roadmap führt ihn als zweiten der Kette).
6. **E6 — `zweig-ohne-lauf-unter-dem-detektor` (18×) und `report-der-rolle-faellt-docs-check` (5×):** Träger `slice-regeln-verkoerpern` (`open/`, dritter der Kette); die Vorlagen stehen in `welle-mock-results` (Vorlage 1 und 2). Zu entscheiden: der Start; der Planner hat nur den Ausgang *geplant* in den `state.md` nachgetragen, den die Entscheidung vom 2026-11-08 trägt (Drift-Log der Roadmap), keine Regel geschrieben.
7. **E7 — die drei Fragen aus der Eröffnung der Welle, weiter beim User** (bremsen nichts; Stand der Policy bis zur Antwort in `SPEC-261` Nr. 3 d und 4 b): Replay-Fallback auch für mutierende Methoden (bis dahin ja), `ODR-E6006` als Auslöser (bis dahin nein), Lücke im Lastenheft (der Replay-Fallback hat keine Anforderung; Beleg §6 und §7 des Plans von `slice-fallback-policy-entscheidung`).
8. **E8 — Status von [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) (`Proposed`) und der Pflichtlauf in `.claude/commands/close-welle.md`:** die ADR ist als Gate seit mehreren Wellen in Kraft, die Annahme steht aus (offen wie bei `welle-etag`); der Absatz „Pflichtlauf neben der Gate-Kette“ des Commands nennt Läufe, die seit `slice-gate-integration-determinism` in `GATE_CHECKS` hängen (in dieser Closure liefen sie als Teil von `make gates`); ob der Absatz entfällt, ist Entscheidung des Users (der Planner schreibt `.claude/` nicht).
9. **E9 — Sendung der Welle ohne Nehmer: Spans und Zähler des Hybrid-Fallbacks.** §6 der Welle-Datei weist sie `welle-betrieb` zu; kein Slice von `welle-betrieb` nimmt sie an, und `slice-hybrid-service` schließt sie in seinem §1 aus (keine Adresse, `AGENTS.md` §3.13). Der Architect empfahl in §6 von `slice-telemetry-port` einen neuen Slice `slice-telemetry-einbau-hybrid` in `welle-betrieb`, nach `slice-hybrid-service` und `slice-telemetry-spans`; der Planner legt ihn nicht an, bevor der User entschieden hat (derselbe Punkt steht in §6 von `slice-telemetry-einbau-gateway` mit Ausgang *offen*). Die Welle schließt damit mit einer Sendung ohne Nehmer, die ihr Closure-Trigger nicht verlangt (Out-of-Scope §6 der Welle-Datei).

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

- `slice-bootstrap-m6` (`open/`, `welle-admin-api`): zieht den Stand von `slice-bootstrap-m4` (`src/bootstrap/`, `SPEC-110` Nr. 2) beim Start nach; nie zugleich mit einem anderen Slice, der `src/bootstrap/` ändert.
- `slice-guard-schreibwege` (`open/`, ohne Welle, Linie C): Träger von `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`; unverändert.
- `slice-regeln-verkoerpern` (`open/`, ohne Welle, Linie C): Träger von `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` und `BEO-REPO/report-der-rolle-faellt-docs-check`; unverändert.
- `slice-groessenformel-neu` (`open/`, ohne Welle, Linie C): Träger der Neukalibrierung (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`, `BEO-REPO/auslastung-am-handoff-als-endwert-gelesen`); der Datenpunkt von `slice-bootstrap-m4-konfig` (Ende ÷ Handoff 2,58) steht im Register.
- `slice-gegenproben-eichung` (`open/`, ohne Welle, Linie C): Träger der Eichung zum Zeitbudget; unverändert, braucht einen Termin ohne schwere Läufe.
- Kein Slice, ein Rest: der Fall „`412` im Modus `hybrid` liest den Zustand zurück“ (`BEO-REPO/zustand-unveraendert-nach-fehler-ohne-ruecklesen`, Adresse Planner, Sichtungs-Schritt); kein offener Slice berührt `tests/integration/m4_test.go` oder die Fälle des Mock bei `412`.
- Kein Slice, ein Auftrag: die Lesart von `SPEC-210` Nr. 5 a und 9 c für die Zeitgrenze der Nullprobe (Adresse Architect, Entscheidung E3), Träger `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`.

## Verifikation

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Schritt 1 — keine Behauptung ohne nachprüfbaren
Anker (Hash, Lauf, Zahl).

Stand der Läufe: `main` auf `5f5e6fe4`, Arbeitsbaum ohne Änderung an Code, Spezifikation und Gates (die Änderungen dieser Closure betreffen Planungs-Dokumente und das Register). Läufe unter `flock` auf der Sperre der Arbeitsbäume. Last des Hosts beim Start: Loadavg über eine Minute 2,94 bei 3 CPUs.

- **Alle Slices der Welle in `done/`:** 6 von 6 Dateien `slice-<Kennung>.md` der Tabelle in §4 der Welle-Datei liegen unter `done/` (`slice-fallback-policy`, `slice-fallback-policy-entscheidung`, `slice-hybrid-service`, `slice-hybrid-service-upstream`, `slice-bootstrap-m4-konfig`, `slice-bootstrap-m4`); Verifikation je Slice bestanden (`slice-bootstrap-m4-konfig` und `slice-hybrid-service-upstream` nach mehreren Review-Runden, `slice-fallback-policy-entscheidung` mit vier Auflagen: zwei erledigt, eine als Sendung in §6, eine als Hinweis).
- **`make gates`:** Exit 0 in 34 s. Aus dem Log: `a-check` ohne Befund, `baseline-verify` v6.18.1 OK (54 Dateien), `docs-check` 973 Dateien ohne Befund, `groessenmarke` „kein Slice in Arbeit“, `zeichen-check` ok (2007 Dateien). Die Stufen `test`, `race`, `lint`, `build`, `handbuch` und `katalog` kamen aus dem Docker-Zwischenspeicher (Quellen gleich dem letzten Bau), die Laufzeit der Tests ist damit nicht neu gemessen; deshalb der frische Lauf unten.
- **Pflichtlauf neben der Gate-Kette** (`close-welle` Schritt 1): gedeckt, denn `make test-integration`, `make test-determinism` und `make test-contract` hängen an `GATE_CHECKS` und liefen innerhalb von `make gates` (keine Zwischenspeicher-Stufen, Skripte unter `tools/harness/`): `test-contract` `/out/bootstrap.test` und `/out/integration.test` grün; `test-determinism` „Läufe: 100, abweichend: 0“ (`TestE2EDeterminismus`); `test-integration` zehn Stufen grün (`TestE2EProzess` bis `TestE2EMockNeustart`, Stufe 10 `TestE2EHybrid` mit der Abdeckung `LH-FA-12`, `LH-FA-04`).
- **Closure-Kriterium §3, dritter Punkt (M4-Test), frisch gefahren** (`go test ./tests/integration/ -run 'TestM4$' -count=1 -v` im Image `odata-recorder:test`, `--memory 3g`, ohne Netz; Exit 0, `PASS`, 1,5 s): `TestM4/Ohne_Auslöser_kein_Upstream` (Treffer des Mock, fachliche `404`, `409`, `412`, ungültige Anfrage: 0 Verbindungen), `TestM4/Lesender_Fallback` (genau eine lesende Verbindung bei unvollständiger Coverage mit `HYBRID_FALLBACK=enabled` und freigegebenem Host), `TestM4/Ohne_Freigabe` (`ODR-E6003` und 0), `TestM4/Mutierender_Fallback` (POST auf eine nicht unterstützte Operation: ohne Eintrag in `FALLBACK_WRITE_ALLOW` 0, mit passendem Eintrag eine), `TestM4/Replay-Miss` (`ODR-E5001`, Opt-in nur für `GET` und `HEAD` zum freigegebenen Host), `TestM4/Erschöpfter_Key` (`ODR-E5002`, 0), `TestM4/Replay-Fallback`, `TestM4/Strikte_Abdeckung`, `TestM4/Gleichzeitig`. `TestM4` ist Teil von `make test` (Paket `tests/integration`).
- **Closure-Kriterium §3, vierter Punkt (`412`):** erfüllt in der gesenkten Fassung (0 Verbindungen, 0 Requests, ohne Kennzeichnung; `TestM4/Ohne_Auslöser_kein_Upstream`); die Hälfte „Zustand unverändert“ ist nicht belegt, siehe *Was ging anders als geplant*, benannte Abweichung mit Folge.
- **Closure-Kriterium §3, fünfter Punkt (Handbuch):** das Kapitel `Modus hybrid` und die Zeilen der Einstellungen stehen im Benutzerhandbuch; `make handbuch-check` und `make katalog-check` sind Teil des grünen `make gates`; ohne gebundenes Beispiel für `hybrid` (benannte Abweichung).
- **`make gegenproben`:** grün. Der Aufruf meldet „Eingaben unverändert seit dem letzten grünen Lauf (39ea6316…) — nichts läuft neu“; der Stempel `harness/gegenproben.stempel` trägt den Stand des Merge von `slice-expand-parser` (Commit `5f5e6fe4`), der Code von `src/`, `tests/` und `tools/` hat sich seither nicht geändert. In dieser Closure ist kein neuer Lauf der Gegenproben gefahren; das Kriterium stützt sich auf den Stempel, nicht auf eine Messung der Laufzeit.
- **Trigger-Audit, Carveouts:** `docs/plan/carveouts/` führt keine Datei; 0 offen, keine Welle schließt mit einem Carveout, kein Gate ist rot.
- **Trigger-Audit, bootstrap-aware Gates:** kein Gate trägt eine Reifestufe; die Schwellen sind feste Grenzen (Größengrenzen der Guide-Dateien nach [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md), Größenmarke nach `SPEC-128`). Keine Verschiebung in dieser Closure. 0 offen.
- **Trigger-Audit, Entscheidungen** (Re-Evaluierungs-Trigger je ADR gegen den Stand der Closure):
  - [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0002](../../adr/0002-architektur-gate-a-check.md), [ADR-0003](../../adr/0003-lint-profil-golangci.md), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md): `a-check`, `lint` und `lint-profil` grün, kein Anlass; Pin `v6.18.1`, `baseline-verify` OK. Nicht eingetreten.
  - [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md): siehe `make gegenproben` oben; kein gültiger Lauf mit gerissener Frist, der Trigger von [ADR-0013](../../adr/0013-zeitbudget-der-gegenproben-als-soll-mit-messvorschrift.md) ist nicht eingetreten. Die Gegenproben bleiben außerhalb der Kette.
  - [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md): `make test` und `make migration-check` sind Teil des grünen `make gates`; keine Auffälligkeit der Bibliothek in den Plänen der Welle. Nicht eingetreten.
  - [ADR-0007](../../adr/0007-race-detektor-im-gate.md): `make test-race` ist Teil des grünen `make gates` (aus dem Zwischenspeicher), die Verifikation von `slice-bootstrap-m4` fuhr es frisch ohne Zwischenspeicher; kein Rennen gemeldet. Der Trigger nennt 10 Minuten Paketlaufzeit unter dem Detektor: in dieser Closure nicht neu gemessen. Nicht eingetreten.
  - [ADR-0008](../../adr/0008-blob-bereinigung-im-betrieb-waechter-und-zurueckstellen.md): kein neuer Schreiber von Blobs in dieser Welle (die Fallback-Antwort wird nicht gespeichert, `LH-FA-12`). Nicht eingetreten.
  - [ADR-0009](../../adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md): `make groessenmarke` ohne Slice in Arbeit grün; eine Marke wurde gerissen (`slice-bootstrap-m4-konfig`, Eintrag mit Grund im Plan, das Gate nahm ihn an). Status bleibt `Proposed`, die Annahme ist Entscheidung des Users (E8).
  - [ADR-0010](../../adr/0010-groesse-der-guide-dateien-als-gate.md): keine Verschiebung der Grenzen in dieser Welle, kein Guide-Datei-Diff in dieser Closure. Nicht eingetreten.
  - [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md), [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md): von dieser Welle nicht berührt (kein V2-Konstrukt; `make a-check` ohne Befund). Nicht eingetreten.
  - [ADR-0030](../../adr/0030-eingecheckte-gate-staempel-mit-voller-gegenprobe.md): `Proposed`, von dieser Welle nicht berührt.
- **Archivierung:** der Träger liegt (`.harness/state/bin/ai-harness-init`); die Vorprüfung vor dieser Notiz nannte 6 Mitglieder der Welle, 3 wellenlose Slices seit der Closure von welle-etag, 28 fremde, 25 Review-Reports und 3 Sperren (`[ergebnisnotiz]`, `[kein-plan]` fallen mit dieser Notiz und dem Move der Welle-Datei, `[haenger]` an drei Evidence-Dateien, behoben durch Ziehen auf die Kennung). Die Vorprüfung nach Move und Link-Reconciliation meldete „Sperren: keine“ (6 Mitglieder, 1 wellenloser Slice, 2 weitere wellenlose „bleibt liegen (nach der Grenze)“, 28 fremde, 20 Review-Reports). `make archive-welle WELLE=welle-hybrid` lief als Werkzeug, kein Handlauf, in zwei Commits: der reine Move von 7 Slices und dem Welle-Plan nach `done/welle-hybrid/` (Commit `82ef366a`) und der Inhalt (Commit `19ccae81`: `archiv.zip` mit 311 632 Bytes, 8 Stubs, 20 Review-Reports entfernt, Verweis-Nachzug in 5 Dateien). Eingesammelt wurde neben den sechs Slices der Welle der wellenlose `slice-gegenproben-zeitbudget`. Danach `make docs-check` ohne Befund (956 Dateien) und `make gates` Exit 0 in 33 s. Die Commit-Nachrichten des Werkzeugs nennen keine `LH-*`- oder `ADR-*`-Kennung, die Kennungen der Closure tragen die übrigen Commits.
- **Drei Paarungen** (nach dem Schreiben dieser Notiz): *Anker*: kein Eintrag trägt das Feld `liegt in`, es gibt nichts zu prüfen (nichts verkörpert). *Folge-Slice*: `slice-bootstrap-m6`, `slice-guard-schreibwege`, `slice-regeln-verkoerpern`, `slice-groessenformel-neu` und `slice-gegenproben-eichung` liegen als Dateien in `open/`. *Register*: alle genannten Kennungen `BEO-REPO/…` existieren als Verzeichnisse, und alle 80 Verzeichnisse tragen ein nicht leeres `evidence/`.
