# Verifikation slice-audit-verdrahtung

**Rolle:** Verifier (einmal). **Datum:** 2026-11-07. **Zweig:** `slice-audit-verdrahtung`, HEAD `6fd5090` (Fix-Runde nach Review `a269473`), Basis `8667f9d`.
**Frage:** Bauen wir es richtig, gegen Plan, DoD und Spezifikation (`SPEC-139` Nr. 2 bis 4, 6, 7, 10 und *Ungeprüft*; `SPEC-235` Nr. 1, 2, 6 bis 8; `SPEC-234` Nr. 3, 8), Bezug `LH-RB-04`, `LH-QA-04`, `LH-FA-13`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md).

## Urteil

**Bestanden mit Auflagen.** Kein Befund der Klasse MEDIUM oder höher. Drei LOW-Auflagen an den Planner (unten), keine verlangt Code. Produktcode 0 ist bestätigt. Der Slice darf nach Nachzug der Auflage 1 in die Closure.

## Belege statt Behauptung

| Zusage des Implementers | Eigene Prüfung | Ergebnis |
|---|---|---|
| `make gates` grün | selbst gefahren (Log: `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/verif-av/gates.log`, 676 Zeilen, `exit 0`) | grün; `d-check: 747 Datei(en) geprüft, 0 Befund(e)`; `groessenmarke slice-audit-verdrahtung.md: ok, Messung 1288 von Marke 1900`; Race, Lint, Handbuch, Zeichen, Schema, Migration grün |
| Produktcode 0 | `git diff --name-only 8667f9d..HEAD` | nur `docs/plan/.../slice-audit-verdrahtung.md`, `.../open/slice-bootstrap-m6.md`, die Review-Datei und zwei Testdateien unter `src/bootstrap/`; kein Pfad unter `src/` außer `*_test.go`, `export_test.go` unverändert |
| Größenmarke | `git diff --numstat 8667f9d..HEAD` über die drei Pfade des Blocks | 411 + 877 + 0 = 1288 ≤ 1900 (Gate bestätigt); Schätzung 1360, `fakes: 2` (`AuditAufzeichnung`, `SessionsMitFehlertext`) |
| Änderung am Review-Report nur die zwei Links | `git diff a269473..HEAD -- docs/reviews/` | bestätigt: eine Zeile (Bezug), die zwei nackten ADR-Kennungen wurden zu Links; sonst nichts geändert |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -nP '\\u[0-9A-Fa-f]{4}\|BOM\|NBSP'` über beide Testdateien und den Plan | kein Treffer; Werte des Canary stehen als `[]byte{…}` bzw. Bytefolgen aus ASCII-Literalen |
| Sendung an `slice-bootstrap-m6` | `git diff` der Datei gegen `8667f9d`; `git log main` der Datei | auf `main` mit Commit `0595204` eingetragen (Zweig hat es per Merge): §1 *Übernimmt außerdem* (Geber `slice-audit-verdrahtung`), DoD Punkt 2 (Fänger mit dem Tokenwert als Canary, auch nach abgewiesener Anfrage), Prüfung der Sendungen; kein *Ausdrücklich NICHT* des Nehmers trifft sie (gelesen: `HYBRID_*`/`FALLBACK_*`, `LISTEN_TLS_*`, neue Endpunkte, Modus zur Laufzeit). Der Nehmer liegt in `open/`, ist also Adresse. Inhalt Zweig = `main`-Stand (Diff leer). |
| Vertrag in der Spezifikation vor dem Code | `SPEC-139` Nr. 2, 3, 6, 7, 10 gelesen | Wortlaut trägt: Eine Instanz je Prozess mit `audit.jsonl`; `Hochfahren` Schritt (1) als Auslöser der Aufbewahrung; `Herunterfahren` schreibt keinen Eintrag und legt die Datei nicht an, auch wiederholt |

## DoD gegen Beleg

| DoD-Punkt (§2) | Beleg | Stand |
|---|---|---|
| Datei im Betrieb, beide Modi, V01a bis V01d, V06 | `TestAuditEintraegeJeFall` (F1 bis F7), `TestAuditSchreibfehlerAendertNichts` | belegt |
| Neustart, leere Ablage, abgelehnte Löschung: V02a, V02b, V05a bis V05c | `TestAuditNeustart`, `TestAuditNeustartHaengtAn`, `TestAuditOhneLoeschungKeineDatei`, `TestAuditAblehnungenKeinEintrag`, `TestAuditBuildFehlerOhneDatei` | belegt |
| Kein Geheimnis (V03), Tokenwert nicht Gegenstand | `TestAuditCanary`, `TestAuditCanaryOrakel`; Tokenwert als benanntes Negativ mit Grund und Folge, Adresse `slice-bootstrap-m6` | belegt; Sendung angenommen |
| `make gates` grün | von mir gefahren | belegt |
| Review, Report liegt vor | `docs/reviews/2026-11-07-slice-audit-verdrahtung-review.md` | belegt |
| Handbuch | Begründung in §6 (kein gebundenes Beispiel möglich), Adresse `slice-admin-sessions` | zugesagt, Ausgang *weiter offen* mit Adresse; zulässig |
| Closure-Notiz, Register, Ausgänge, Paarungen | §7 *Wird bei der Closure gefüllt* | offen, Sache der Closure (Planner) |

V07 (Aufrufstellen) steht in der Tabelle, aber in keinem DoD-Punkt als eigener Liefer-Punkt; er ist durch Punkt 1 (Zahl der Zeilen) gedeckt. Kein Befund.

## Plan gegen Code (Tabelle *Zusage · Fänger*)

Alle zwölf Zeilen (V01a bis V07) haben einen Test mit dem im Plan genannten Namen; die Mengen M1 (7), M2 (172), M3 (7), M4 (3) tragen eine Zählprüfung im Test (`len`, `n != 172`, `len(ids) != 9`, `len(faelle) != 7`). Die Zweig-Menge W (vier Aufrufstellen) wird per Syntaxbaum gezählt. Abweichung Plan gegen Code: nur die Zeile in Auflage 1.

## Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Menge aus dem Wortlaut der Zusage, nicht aus dem Befund: *Felder der Zeile, deren Wert Root, Wrapper oder ID-Quelle beeinflusst* (acht Felder) × *Kodierung des Canary*. Eigene Mutationen an Elementen, die der Befund und die Fix-Tabelle nicht nannten (Werkzeug `tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, 4g, unter `flock /tmp/odata-recorder-heavy.lock`, Klasse `andere`, 1 Prozess je Mutation, `last` 3,5 bis 4,3 von 4; Messender: Verifier; Liste `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/verif-av/liste2.json`, Ausgabe `.../verif-av/mut2.out`):

| Mutation | Feld · Kodierung | Fänger | Ergebnis |
|---|---|---|---|
| `sess-svc-hex-pw` | `service_id` (Aufrufstelle Session) · Hex klein des Passworts | `TestAuditCanary` | rot |
| `maint-svc-hex-ben-upper` | `service_id` (Aufrufstelle Wartung) · Hex groß des Benutzers | `TestAuditCanary` | rot |
| `imp-id-b64raw-pw` | `import_id` · Base64 ohne Füllung des Passworts | `TestAuditCanary` | rot |
| `cor-hex-ben-fest` | `correlation_id` · `cor-` + Hex des Benutzers (anderer Wert als die Befund-Mutation) | `TestAuditCanary` | rot |
| `cor-hex-ben-upper` | `correlation_id` · Hex groß | `TestAuditCanary` (Form `^cor-[0-9a-f]{32}$`) | rot |
| `cor-b64-pw` | `correlation_id` · Base64 | `TestAuditCanary` | rot |
| `cor-teilstueck-je-zeile` | `correlation_id` · je Zeile verschieden, Form `cor-` + 32 Hex, 8 Byte des Passworts als Hex plus 16 Zufallsziffern | `TestAuditCanary` | **grün** |
| `ws-art-vertauscht` | `art` | `TestAuditEintraegeJeFall` | rot |
| `sess-extra-feld` | Feld außerhalb der Pflichtfelder der Art | `TestAuditEintraegeJeFall` | rot |
| `v07-alias-importer` | Alias von `k.Audit` in `baselineimport/importer.go` (anderer Ort als beim Review) | `TestAuditAufrufstellen` | rot |
| `v07-paren` | `(w.k.Audit).Append` in `maintenance/startup.go` (Aufruf, den der erste Zähler nicht sah) | `TestAuditAufrufstellen` | rot |

10 von 11 rot, 1 grün wie vorhergesagt. Grenze meiner Prüfung: das Werkzeug gibt keine Meldung des roten Tests aus; ich habe den Grund der Rotfärbung nicht je Mutation gelesen. Für die Felder, die M1 auf den erwarteten Wert festlegt (`service_id`, `session_id`, `workspace_id`, `import_id`), färbt auch M1 rot; das sagt der Plan selbst (Prüfgrenze, Folge). Diskriminierend sind nur die `correlation_id`-Zeilen, und dort ist der Grund an der Mutation ablesbar (Hex klein: Orakelform; Groß: Form; Base64: Form und Base64-Alphabet).

### Prüfgrenze `correlation_id` (Folge ehrlich?)

Die grüne Mutation `cor-teilstueck-je-zeile` bestätigt die Prüfgrenze im Plan: ein je Zeile verschiedener Wert in der Form `cor-` + 32 Hex bleibt ungefangen. Die Folge im Plan ist im Kern ehrlich (»das Geheimnis wäre in der Datei lesbar«). Zwei Feinheiten, die keine Änderung am Code verlangen: (a) der Satz trifft auf Teilstück und Abwandlung zu, nicht auf einen Hash (ein Hash macht das Geheimnis nicht lesbar, nur vergleichbar); (b) `timestamp` kann nur Ziffern tragen, deshalb ist die Prüfgrenze dort enger als der Plan nahelegt, aber ungefährlich. Kein Befund, Auflage 2 nur zur Schärfung des Wortlauts.

### V07, Syntaxbaum-Robustheit

Der Zähler erlaubt `Audit` nur als Empfänger eines Aufrufs von `Append` und im Vergleich mit `nil`; jede andere Verwendung des Selektors (Alias, Methodenwert, Übergabe, Klammer) zählt als Fund. Eigene Mutationen: Alias in einer anderen Datei, Klammer um den Selektor: beide rot. Bekannte Grenze, im Plan benannt: Reflexion und Übergabe an eine Hilfsfunktion ohne den Selektor `.Audit`; ein Eintrag in `src/hexagon/services/` außerhalb von `_test.go` ist erfasst. Die Zählung gilt nur für `src/hexagon/services/`; ein Aufruf von `Append` aus einem anderen Verzeichnis unter `src/hexagon/` wäre ungezählt. Der Plan nennt als Gegenstand »Code der Dienste«; das deckt es, die Grenze steht aber nicht als Negativ im Plan (Auflage 3).

### Register `zweig-ohne-lauf-unter-dem-detektor`

Die Modell-Menge W/N in §3 ist vollständig gegen die Zeilen V01 bis V06 abgebildet (W1 bis W4, N1 bis N8 mit Fänger; die Zweige `ErrReferenziert`, Kontextfehler, gleicher Hash als benanntes Negativ mit Grund und Folge). Die Abdeckungsangabe (ungedeckt: die Rückgaben der Konstruktoren in `verdrahte` und `verdrahteRecord`) habe ich **nicht** selbst erneut gemessen: mein Lauf eines Wegwerf-Containers mit `-coverpkg` wurde von der Sicherheitsprüfung der Umgebung abgewiesen (`docker rm -f` im Skript). Sie bleibt eine Angabe des Implementers; der Reviewer hat sie nicht widersprochen. Das benannte Negativ »Fehlerzweig eines Konstruktors im Root ohne Lauf« mit Registervorschlag trägt Grund und Folge (§3.15 erfüllt).

## Befunde und Auflagen an den Planner

1. **LOW, §3.9 (Plan folgt jeder Korrektur): zwei Stellen nennen die Zahl 6 und die Achsen ohne die siebte Ablehnung.** Die Zeile V05b der Tabelle sagt »Schleife über sechs Ablehnungen« und listet »Session von einer Baseline referenziert« (`ODR-E7004`) nicht; die Achse *Grund* der Menge M3 nennt »referenziert« nicht, obwohl die Zeile »7 (…die sieben…)« und der Test (`len(faelle) != 7`) es tragen. Das ist dieselbe Aussage an mehreren Stellen, nur eine (die Mengenzeile und §3 »Mengen«) wurde nachgezogen. Nachzug im Plan vor der Closure; kein Code.
2. **LOW, Wortlaut der Prüfgrenze »Weitere Kodierungen«:** die Folge sollte zwischen Teilstück und Abwandlung (Geheimnis lesbar) und Hash (nur vergleichbar) unterscheiden, und sagen, dass der Fehlertext-Canary nur in der rohen Form geprüft wird (eine Form in `TestAuditCanary`; Hex, Base64 und Prozent-Kodierung gelten für Benutzer, Passwort und Query-Wert). Im Betrieb folgt: der Fehlertext erreicht die Datei nur über den Eintrag, und kein Dienst legt ihn in ein Feld (F7 rot sonst); ein Root, der ihn kodiert einsetzte, bliebe ungefangen.
3. **LOW, benannte Grenze von V07:** der Zähler läuft allein über `src/hexagon/services/`. Als akzeptiertes Negativ mit Folge ergänzen (ein Aufruf von `Append` auf der Instanz an anderer Stelle von `src/hexagon/` oder in `src/bootstrap/` außer dem Root bliebe ungezählt; gefangen nur, weil die Dateien der Datei-Fänger V01 bis V06 die Wirkung sehen).

Keine Auflage an den Implementer; kein Rücksprung an Architect.

## Nicht behauptet

Nicht erneut gemessen: die Abdeckung der berührten Pfade; die 20 Mutationen des Implementers und die 5 des Reviewers (Stichprobe von elf eigenen genügt der Klasse, nicht der Zahl); `make test-integration` und `make test-determinism` (kein Gate dieser Kette, nicht Gegenstand des Slice). Der Grund jeder roten Mutation ist nicht je Lauf aus der Meldung gelesen (Werkzeug liefert nur das Ergebnis).

## Rückführungs-Hinweis (Planner)

Closure-Pflichten, die offen sind und nicht Sache dieser Rolle: §7 Closure-Notiz mit Lerneintrag, Beobachtungs-Register (Vorschlag »Fehlerzweig eines Konstruktors im Root ohne Lauf«, Prüfgrenze der Kodierungen), Ausgänge der §6-Risiken *weiter offen* mit Adressen (`slice-admin-sessions`, `slice-bootstrap-m6`, `slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-mock`).
