# Slice slice-admin-abweisen-begrenzen: Einträge des Abweisens begrenzen (höchstens 20 Einträge `admin_abgewiesen` je Fenster von 60 s, `ODR-E2101` bleibt gezählt)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md) (der Planner trägt den Slice in §4 und §5 der Welle nach).

**Bezug:** `LH-FA-13`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-301` (Begrenzung der Einträge des Abweisens, Architect, 2026-11-24) · `SPEC-238` Nr. 1, 3 d, 8, 12 und *Ungeprüft* (nachgezogen) · `SPEC-234` Nr. 2 und 7 · `SPEC-300` Nr. 5 d, 12 b und *Ungeprüft* · `SPEC-299` *Ungeprüft* · `LH-FA-13.a` der Spezifikation (der Satz zu Audit und Token) · `SPEC-041` Nr. 7 und 7 c (Zählung am Netzwerkzugriffstest) · `SPEC-114` (Uhr) · `SPEC-128` Nr. 1 (Größenmarke).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Architect; Entscheidung des User vom 2026-11-24: Variante A der Entscheidungsvorlage zu R5 in §6 von [slice-bootstrap-m6](slice-bootstrap-m6.md)). **Datum:** 2026-11-24.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem** Abschnitt des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Admin-Adapter schreibt höchstens 20 Einträge `admin_abgewiesen` je festem Fenster von 60 s (der Beginn ist die Abweisung, die das Fenster eröffnet); jede weitere Abweisung bleibt 401 `ODR-E2101` mit `X-Correlation-Id`, schreibt keinen Eintrag und steht im Zähler `ODR-E2101`; `admin.Konfig` trägt dafür die Uhr, und der Composition Root übergibt sie (`SPEC-301`).

**Übernimmt:** den Gegenstand *Begrenzung der Fehlversuche* (Rückgabe R5 des Architect vom 2026-11-07, Geber `slice-admin-listener-token`; Entscheidungsvorlage in §6 von [slice-bootstrap-m6](slice-bootstrap-m6.md), dessen §1 *Ausdrücklich NICHT* den Punkt als eigenen Folge-Slice mit dem Vorschlag dieser Kennung ausschließt: der Nehmer nimmt an, §3.13 von `AGENTS.md`). Der Planner zieht in [slice-bootstrap-m6](slice-bootstrap-m6.md) §6 den Ausgang der Vorlage von *offen* auf *entschieden* nach (die Pläne der Teile schreibt der Architect nicht).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Sperre je Quelle (Variante B der Vorlage), ein Code `ODR-E2103` mit 429, ein Header `Retry-After` — Entscheidung des User (2026-11-24): Variante A; B sperrt hinter NAT und auf Loopback den Betreiber selbst aus.
- Die Mindestlänge von 64 Zeichen für das Token und die Rotation von `audit.jsonl` (Variante C) — verworfen: C1 verhindert den Start bestehender Installationen und belegt keinen Zufall, C2 spült die Einträge der Änderungen im Angriff aus der Generation.
- Ein Eintrag „n weitere“ für die gedrosselten Abweisungen — brauchte eine neue Art und ein Feld (`SPEC-234` Nr. 2 ist geschlossen); die Zahl steht im Zähler `ODR-E2101` (`SPEC-301` Nr. 4).
- Ein Schlüssel `ADMIN_*` für 20 und 60 s — feste Werte des Pakets (`SPEC-301` Nr. 1): kein Betreiber braucht den Schlüssel, und jeder Schlüssel kostet Tabelle, Katalog und Handbuch.
- Eine Änderung der Teile von [slice-bootstrap-m6](slice-bootstrap-m6.md) oder ihrer Tests — sie liegen vor diesem Slice in `done/` (Start, §4); ihre Tests zählen höchstens 18 Abweisungen je Handle (`TestM6AbweisungOhneAenderung`, 9 Routen × 2 Anlässe im Modus `record`) und bleiben mit der Grenze von 20 grün. Wächst eine ihrer Mengen über 20 je Handle, ist das ein Nachzug dieses Slice (§6).
- Das Kapitel zum Admin-Listener im Handbuch — das Kapitel liefert Zeile B17 von [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md); dieser Slice fügt ihm zwei Sätze hinzu (`SPEC-301` Nr. 8) und baut es nicht.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Das Fenster im Adapter und die Verdrahtung: `admin.Konfig.Uhr`, die Zulassung unter einer Sperre, gedrosselte Abweisungen ohne `Append`, der Root übergibt die eine Uhr (Zeilen W01 bis W17 der Tabelle in §3; `LH-FA-13`, `SPEC-301` Nr. 1 bis 7).
- [ ] Der Nachzug der Zählung: der Netzwerkzugriffstest zählt 1008 Einträge mit vorrückender Uhr und 20 ohne (Zeile W18, `SPEC-041` Nr. 7 c), und jeder Test der Pakete `admin` und `bootstrap`, der eine Zeile je Abweisung erwartet, hält die Grenze von 20 je Fenster ein oder rückt die Uhr vor.
- [ ] Das Handbuch: im Kapitel zum Admin-Listener die zwei Sätze von `SPEC-301` Nr. 8 (höchstens 20 Einträge je 60 s und der Zähler; ein selbst gesetztes Token ab 32 Zeichen ist nur gegen Raten geschützt, wenn es ein Zufallswert ist); `make handbuch-check` grün.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo mit Wellen von der nächsten Welle-Closure.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/adapters/driving/admin/admin.go` | update | `Konfig.Uhr` (nil ist `model.ErrUngueltig`), das Fenster als Zustand von `Server`, die Zulassung vor `Append` an der einen Stelle der Abweisung (`SPEC-301` Nr. 1, 2, 3) |
| `src/adapters/driving/admin/abweisen_fenster.go` | neu | der Zustand (Beginn, Zahl, Sperre) und die eine Funktion der Zulassung; die Konstanten 20 und 60 s; kein Timer, keine Goroutine (`SPEC-301` Nr. 2, 5, 6) |
| `src/adapters/driving/admin/abweisen_fenster_test.go`, `abweisen_fenster_gleichzeitig_test.go` | neu | die Fänger der Zeilen W01 bis W16 (die nebenläufigen W10 bis W12 in der zweiten Datei) |
| `src/adapters/driving/admin/abweisen_audit_test.go`, `abweisen_gleichzeitig_test.go`, `konfig_test.go`, `fakes_test.go`, `helfer_test.go`, `export_test.go` | update | jede Hilfe, die `Konfig` bildet, setzt `Uhr`; die Zählungen „je Abweisung ein Eintrag“ rücken die Uhr oder bleiben unter 20 (§6) |
| Netzwerkzugriffstest des Pakets `admin` (Datei nach dem Namen aus [slice-test-security](../done/slice-test-security.md)) | update | Zeile W18 (`SPEC-041` Nr. 7 und 7 c) |
| `src/bootstrap/wurzel.go` | update | die eine Zeile `Uhr: <die Uhr des Prozesses>` in `admin.Konfig` (`SPEC-300` Nr. 5 d) |
| `src/bootstrap/admin_abweisen_test.go` (neu), `src/bootstrap/export_test.go` | neu und update | Zeile W17: 21 Abweisungen am Listener des `Handle`, 20 Zeilen, Zähler 21, je Modus |
| Handbuch (`docs/user/benutzerhandbuch.md`, Kapitel zum Admin-Listener) | update | die zwei Sätze von `SPEC-301` Nr. 8; Text ohne Planung, Chronik und Kennung (`SPEC-200`) |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-041-test-security.md` | geliefert (Architect, 2026-11-24) | `SPEC-301` neu, Nachzug von `SPEC-238`, `SPEC-234`, `SPEC-292`, `SPEC-299`, `SPEC-300`, `LH-FA-13.a` und `SPEC-041`; dieser Slice schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen vor dem Code (`AGENTS.md` §3.12); die Randformen entscheidet der Architect in `SPEC-301` Nr. 5, nicht der Implementer (Architect, 2026-11-24; jede Zeile nennt die Nummer des Vertrags). Der Implementer trägt je Zeile Mutation und roten Test, die Raten der nebenläufigen Zeilen mit Zahl der Prozesse und Messendem und das Ergebnis der Mutation in einen Absatz von §3 ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| W01 | (`SPEC-301` Nr. 1) `New` weist eine nil-`Uhr` ab (`model.ErrUngueltig`, kein Wert im Text), ruft die Uhr nicht und startet keine Goroutine; `Konfig` führt für Zahl und Dauer kein Feld | `TestNewUhr`: nil-`Uhr` ergibt `ErrUngueltig` ohne `Server`, Text ohne Wert; die Fake-Uhr hat nach `New` 0 Aufrufe; `reflect` über `Konfig`: genau 8 Felder (die 7 von `SPEC-238` Nr. 1 und `Uhr`) | nil gilt als Systemuhr; `New` ruft `Now()`; ein Feld für die Dauer | andere |
| W02 | (Nr. 2 c, 5 a) Die ersten 20 Abweisungen eines Fensters sind zugelassen, ab der 21. gedrosselt | `TestFensterZahl`: Schleife n von 1 bis 25 Abweisungen bei stehender Uhr: `Append` min(n, 20) mal; die Zeilen n = 19, 20, 21, 22 mit Literal | Grenze 21; `<=` statt `<`; die Zahl steigt auch gedrosselt (bei n = 25 steigt sie über 20, das Fenster verliert den Stand) | andere |
| W03 | (Nr. 2 b, 5 b) Fenstergrenze: bei 60 s − 1 ns gedrosselt, bei 60 s zugelassen mit neuem Beginn; die Dauer ist 60 s | `TestFensterGrenze`: nach 20 Zulassungen `Advance(60 s − 1 ns)`: die 21. gedrosselt; `Advance(1 ns)`: zugelassen, danach 19 weitere zugelassen und die 21. des neuen Fensters gedrosselt; die Dauer als Literal (59 s und 61 s sind je eine Zeile) | `>` statt `≥` (Gleichheit); Dauer 59 s oder 61 s; die Dauer in ganzen Sekunden abgeschnitten (`Seconds()`) | andere |
| W04 | (Nr. 2 b, 5 c) Der Beginn ist die Zeit der eröffnenden Abweisung (nicht `New`, kein Vielfaches von 60 s); eine beliebig lange Ruhe öffnet ein neues Fenster | `TestFensterBeginn`: Uhr bei `New` auf *t*, erste Abweisung bei *t* + 5 min 17 s, danach 20 zugelassen; die 21. bei *t* + 5 min 17 s + 59,999999999 s gedrosselt, bei +60 s zugelassen; ein Sprung von einem Tag öffnet ein Fenster | Beginn bei `New`; Beginn auf volle Minuten gerundet; der Beginn des neuen Fensters nicht gesetzt | andere |
| W05 | (Nr. 5 c) Die Uhr liefert einen Wert vor dem Beginn, die Nullzeit oder eine Zeit außerhalb von 1 bis 9999: das Fenster bleibt offen, keine Panik, kein Überlauf; „noch kein Fenster“ ist ein eigener Zustand | `TestFensterUhrWerte`: eine Test-Uhr liefert nach 20 Zulassungen 1 ns, 1 Jahr, die Nullzeit und das Jahr 10000 vor dem Beginn (4 Werte): alle gedrosselt, keine Panik; als erste Zeit die Nullzeit: 20 Abweisungen zugelassen, die 21. gedrosselt (kein neues Fenster je Abweisung); ein Wert nach dem Beginn (Jahr 9999) öffnet ein Fenster | Betrag der Differenz; `Beginn.IsZero()` als „kein Fenster“; vorzeichenlose Rechnung | andere |
| W06 | (Nr. 3) Die gedrosselte Antwort gleicht der zugelassenen bis auf `X-Correlation-Id`: Status, Body, Header und ihre Werte, kein `Retry-After` | `TestGedrosseltAntwort`: Menge der Anlässe (kein Header; leerer Wert; falsches Token; Schema `Basic`; eine Route mit `*inbound.Fehler` `ODR-E2101`) × (zugelassen, gedrosselt) = 5 × 2 Zellen: Body-Bytes und die Menge der Header ohne `X-Correlation-Id` gleich | `Retry-After` an der gedrosselten; Status 429; ein Header `X-Gedrosselt` | andere |
| W07 | (Nr. 3, 4) Jede Abweisung, auch die gedrosselte, ruft `CountError(ODR-E2101)` genau einmal und `IDs.KorrelationsID()` genau einmal | `TestGedrosseltZaehlt`: 25 Abweisungen: `CountError` 25 mal mit `ODR-E2101`, sonst kein Aufruf des Fakes; `KorrelationsID` 25 mal, die ID jeder Antwort ist die des n-ten Aufrufs | früher Rücksprung bei Drosselung vor `CountError`; `CountError` nur bei Zulassung | andere |
| W08 | (Nr. 3) Gedrosselt: kein `Append`, kein anderer Aufruf eines Ports | `TestGedrosseltOhnePort`: nach 20 Zulassungen bleibt `FakeAudit.Aufrufe()` bei 20 für 5 weitere Abweisungen; der Fake jeder Route hat 0 Aufrufe; die 20 Einträge tragen die Art und den Code, die ID der Antwort und sonst kein Feld | `Append` mit leerem Eintrag für die gedrosselte; ein Zählereintrag anderer Art | andere |
| W09 | (Nr. 2, 5 g, 5 h) Allein die Antwort 401 holt eine Zulassung; 404, 405, 400, 503, 500, 200 und `GET /healthz` rufen die Uhr nicht und ändern das Fenster nie, auch bei vollem Fenster; beide Wege zu 401 teilen ein Fenster | `TestFensterNurAbweisung`: Menge der 7 Antworten × (Fenster leer, Fenster voll) = 14 Zellen: 0 Aufrufe der Uhr, danach 1. Abweisung zugelassen bzw. die 21. gedrosselt; Wege: 10 aus der Authentisierung und 10 aus der Route (zusammen 20 zugelassen), danach die 21. aus jedem Weg gedrosselt (2 Zellen) | `Now()` bei jeder Antwort; 404 verbraucht eine Zulassung; je Weg ein Fenster | andere |
| W10 | (Nr. 5 d) Bei vielen gleichzeitigen Abweisungen im Fenster sind genau 20 zugelassen | `TestFensterGleichzeitig`: 64 Goroutinen hinter einem Startsignal bei stehender Uhr: `Append` genau 20, `CountError` 64; je Prozess unter `make test-race` (`-test.count=1`) und mit `-count=20` | die Sperre des Fensters entfällt (Detektor); die Zahl wird ohne Sperre erhöht | nebenläufig |
| W11 | (Nr. 2, 5 d, 6) Der Zustand zwischen dem Lesen der Zahl und ihrem Erhöhen ist für keinen Request sichtbar (Wirkungs-Fänger: der Detektor sieht ein Freigeben der Sperre dazwischen nicht) | `TestFensterZulassungAtomar`: eine Uhr, die den ersten Aufruf von `Now()` an einem Treffpunkt hält; ein zweiter Request startet, solange der erste hält: seine Uhr wird nicht gerufen und sein `Append` steht nicht vor der Freigabe; nach der Freigabe sind bei 64 Requests genau 20 zugelassen; die Rate je Mutation und Prozess im Absatz von §3 | `Now()` vor der Sperre; Prüfen und Erhöhen in zwei Abschnitten mit je einer Sperre | nebenläufig |
| W12 | (Nr. 2, 5 e) Die Sperre wird vor `Append` freigegeben: eine zweite Abweisung läuft, solange `Append` der ersten hängt | `TestFensterOhneSperreUeberAppend`: `FakeAudit.Vor` hält den ersten `Append`; die zweite Abweisung ist beantwortet und ihr `Append` gerufen, bevor der Test die erste freigibt | die Sperre über `Append` gehalten | nebenläufig |
| W13 | (Nr. 5 e) Eine erteilte Zulassung ist verbraucht, gleich was `Append` liefert; die Antwort bleibt 401 | `TestFensterAppendFehler`: `FakeAudit.Fehler` gesetzt: 25 Abweisungen: `Append` 20 mal, alle 25 Antworten 401, `CountError` 25 mal | die Zulassung zählt erst nach erfolgreichem `Append` | andere |
| W14 | (Nr. 5 f) Das Fenster gehört der Instanz: zwei `Server` teilen keinen Zustand, ein neuer `Server` beginnt ohne Fenster | `TestFensterJeServer`: A verbraucht 20; B (eigene Fakes, eigene Uhr) lässt 20 zu; A wird geschlossen, C lässt 20 zu | das Fenster als Variable des Pakets | andere |
| W15 | (Nr. 5 g) Ein angehaltener Server holt keine Zulassung und ruft die Uhr nicht; ein vor `Anhalten` zugelassener Request schreibt seinen Eintrag | `TestFensterAngehalten`: nach `Anhalten` 3 Requests ohne Token: 503 `ODR-E5010`, 0 Aufrufe der Uhr, `Append` unverändert; ein zugelassener Request, der in `Append` hängt (`Vor`), schreibt seinen Eintrag nach `Anhalten` | die Zulassung vor der Prüfung des Zustands | andere |
| W16 | (Nr. 8) Der Vertrag führt keine Art, kein Feld von `Eintrag`, keinen Code, keinen Schlüssel und kein Kommando ein | `support.Arten()` hat 12 Zeilen und `Eintrag` 7 Felder (`reflect`), `model.AlleCodes()` hat 41 Elemente; `make katalog-check` grün | eine Art `admin_abgewiesen_gedrosselt`; ein Feld `Anzahl` in `Eintrag` | andere |
| W17 | (Nr. 1, 8; `SPEC-300` Nr. 5 d, 12 b) Der Root übergibt die Uhr: am `Handle` ergeben 21 Abweisungen im Fenster 20 Zeilen `admin_abgewiesen` in `audit.jsonl`, und der Zähler `ODR-E2101` der Diagnose steht auf 21 | `TestAdminAbweisenBegrenzt` (Paket `bootstrap`, je Modus `record`, `replay`, `mock`, `hybrid` = 4 Zellen): 21 Anfragen ohne Token an den Admin-Listener: 20 Zeilen mit den vier Schlüsseln in der Folge von `SPEC-238` Nr. 8; `GET /__admin/diagnostics` mit Token zeigt 21 für `ODR-E2101` | der Root übergibt keine Uhr (`New` scheitert, `Build` rot); keine Begrenzung am Root (21 Zeilen) | andere |
| W18 | (`SPEC-041` Nr. 7 c) Der Netzwerkzugriffstest zählt 1008 Einträge mit vorrückender Uhr und 20 ohne | `TestSicherheitKeineAenderung` (Nachzug): `Advance(60 s)` vor jeder der 2 · 504 Abweisungen: 1008 Einträge; Kontrolle ohne Vorrücken: 20 Einträge und 1008 `CountError` | das Vorrücken entfällt (20 statt 1008); die Kontrolle zählt 1008 | andere |

**Mengen, Generatoren und Zählprüfungen** (Orakel als Literal im Test; `AGENTS.md` §3.15 und §3.17): die Menge der Zahlen n von 1 bis 25 mit den Grenzzeilen 19 bis 22 (W02); die Fensterdauern 59 s − 1 ns, 60 s − 1 ns, 60 s, 61 s und die Ruhen 5 min 17 s und 1 Tag (W03, W04); 4 Werte der Uhr vor dem Beginn und 1 danach (W05); 5 Anlässe × 2 (W06); 7 Antworten × 2 Zustände des Fensters und 2 Wege (W09); 4 Modi (W17). Die Randformen 5 a bis 5 h von `SPEC-301` haben je einen Fänger: 5 a W02, 5 b W03, 5 c W04 und W05, 5 d W10 und W11, 5 e W12 und W13, 5 f W14, 5 g W15, 5 h W09 (8 Randformen, 11 Zeilen); die Nr. 1 bis 4 und 8 des Vertrags tragen die übrigen 7 Zeilen W01, W06 bis W08 und W16 bis W18. Die Aussage „jede Antwort außer 401 holt keine Zulassung“ ist eine Schleife über die 7 Antworten, nicht eine Liste der eingefallenen Fälle.

**Zählprüfung der Zeilen:** 18 Zeilen; *a* = 3 (W10 bis W12), *b* = 15, *f* = 1 (die Test-Uhr mit Treffpunkt und den Werten vor dem Beginn), *p* = 100 (admin.go, `abweisen_fenster.go`, eine Zeile im Root).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): alle vier Teile von [slice-bootstrap-m6](slice-bootstrap-m6.md) liegen in `done/` ([slice-bootstrap-m6-endpunkte](slice-bootstrap-m6-endpunkte.md), [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md), [slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md) und der Teil D selbst, der die Kennung des ungeteilten Slice trägt) und [slice-test-security](../done/slice-test-security.md) liegt in `done/`; die Spezifikation trägt `SPEC-301` und den Nachzug von `SPEC-041` vor dem ersten Code-Commit (`AGENTS.md` §3.12; geliefert am 2026-11-24).

**Warum die Reihenfolge nach allen vier Teilen** (Architect, 2026-11-24): (1) Der Aufruf von `admin.New` im Root entsteht in Teil B; vorher ist keine Zeile zu verdrahten. (2) Alle vier Teile ändern `src/bootstrap/` (`wurzel.go`, `export_test.go`) und laufen nacheinander; dieser Slice ändert dieselben zwei Dateien um eine Zeile und einen Test und läuft danach, nie zugleich (*Code-Abhängigkeit*). (3) Die Tests der Teile B und D (Zeile B15, `TestM6Rundlauf` mit 2 Abweisungen, `TestM6AbweisungOhneAenderung` mit höchstens 18 je Handle) setzen „jede Abweisung ein Eintrag“ voraus; sie bleiben mit der Grenze von 20 unverändert grün, und die M6-Prüfung wartet auf diesen Slice nicht. Die Closure von [welle-admin-api](../welle-admin-api.md) trägt ihn als Bedingung (der Planner zieht Welle und Roadmap nach).

**Größenmarke** (`SPEC-128` Nr. 1 und 10 g, `AGENTS.md` §3.14; Vorschlag des Architect 2026-11-24; **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 18 Zeilen der Tabelle (W01 bis W18), *a* = 3, *b* = 15, *f* = 1, *p* = 100: *S* = 85·3 + 80·15 + 200·1 + 100 = 255 + 1200 + 200 + 100 = 1755, 135·1755 = 236925, **Marke 2400** = 100 · ⌈23,6925⌉; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`. Die Entscheidungsvorlage rechnete mit 8 Zeilen (*S* = 795, Marke 1100); die Tabelle führt nach dem Raster von §6 jede Randform von `SPEC-301` Nr. 5 mit einem Fänger und hat 18 Zeilen, die Marke folgt der Tabelle. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

```groessenmarke
pfad: src/adapters/driving/admin/admin.go
pfad: src/adapters/driving/admin/abweisen_fenster.go
pfad: src/adapters/driving/admin/abweisen_fenster_test.go
pfad: src/adapters/driving/admin/abweisen_fenster_gleichzeitig_test.go
pfad: src/adapters/driving/admin/abweisen_audit_test.go
pfad: src/adapters/driving/admin/abweisen_gleichzeitig_test.go
pfad: src/adapters/driving/admin/konfig_test.go
pfad: src/adapters/driving/admin/fakes_test.go
pfad: src/adapters/driving/admin/helfer_test.go
pfad: src/adapters/driving/admin/export_test.go
pfad: src/bootstrap/wurzel.go
pfad: src/bootstrap/export_test.go
pfad: src/bootstrap/admin_abweisen_test.go
produktcode: 100
fakes: 1
```

Der Block nennt die Dateien einzeln (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`); der Netzwerkzugriffstest von [slice-test-security](../done/slice-test-security.md) steht nicht im Block, weil sein Dateiname erst mit jenem Slice entsteht: der Planner nimmt ihn beim Start einzeln auf (`SPEC-128` Nr. 10 g), und der Implementer nennt jede weitere Datei, die er ändert oder anlegt, im selben Commit im Block. Die Dateien des Handbuchs und der Spezifikation stehen nicht im Block (`SPEC-128` Nr. 3 m). Die neuen Dateien (`abweisen_fenster.go`, die Tests, `admin_abweisen_test.go`) gibt es noch nicht (die Namen sind ein Vorschlag des Architect; die Messung zählt dort 0, bis sie entstehen).

**Kollisionen** (Architect, 2026-11-24, gelesen an den Plänen; `SPEC-128` Nr. 10 g): (1) Die vier Teile von [slice-bootstrap-m6](slice-bootstrap-m6.md) ändern `src/bootstrap/wurzel.go` und `export_test.go`: sie liegen vor diesem Slice in `done/` (Start), ihre Zeilen gehören zur Basis. (2) [slice-test-security](../done/slice-test-security.md) legt den Netzwerkzugriffstest im Paket `admin` an und liegt vor Teil A in `done/`. (3) [slice-admin-tls-betrieb](slice-admin-tls-betrieb.md) (Start nach allen vier Teilen) ändert das Handbuch (Kapitel zu TLS), nicht das Paket `admin` und nicht `wurzel.go`; die Reihenfolge gegen diesen Slice ist frei, sie ändern verschiedene Abschnitte des Handbuchs. (4) Die Gate-Kette (Nr. 7 der Übergangsregeln): kein Pfad aus `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`; der Planner prüft es beim Start.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein vorab benannter Schnitt ist der Nachzug der Tests (W18 und die Zählungen der Pakete `admin` und `bootstrap`) als eigener letzter Commit, ohne gemeinsamen Mechanismus mit dem Fenster.
- `in-progress` → `open` (blockiert — Carveout?): ein Slice der Teile von M6 ändert `admin.New` oder `Konfig` anders, als `SPEC-301` annimmt, oder die Dateien des Netzwerkzugriffstests fehlen.

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-24; jede Zelle nennt ihren Ort in `SPEC-301` oder den Verträgen davor oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `New` mit dem Feld `Uhr` | nil: `model.ErrUngueltig`, kein Server (`SPEC-301` Nr. 1, `SPEC-238` Nr. 1) | gilt nicht: ein Port hat keinen leeren Wert; ein Typed-nil steht in `SPEC-238` *Ungeprüft* | eine Uhr, die rückwärts oder die Nullzeit liefert: Nr. 5 c (Zeile W05) | gilt nicht: kein `ctx`, `New` ruft keinen Port (`SPEC-301` Nr. 1) | jedes `New` hat ein eigenes Fenster (Nr. 5 f, W14) |
| Zulassung einer Antwort 401 (Lesen der Uhr, Vergleich mit Beginn und Zahl, Erhöhen) | noch kein Fenster eröffnet: die erste Abweisung eröffnet eines (Nr. 2 b) | Zahl 0 im neuen Fenster: die eröffnende Abweisung ist zugelassen (Nr. 2 b, c) | eine andere Antwort als 401: keine Zulassung (Nr. 5 h, W09); Zeit außerhalb von 1 bis 9999: Nr. 5 c | der Client trennt während `Append`: die Zulassung ist erteilt, `Append` läuft ohne `ctx` weiter (Nr. 5 e, `SPEC-238` Nr. 8) | die 20. und die 21. Abweisung, die Gleichheit an der Fenstergrenze (Nr. 5 a, b; W02, W03) |
| Fenster (Zustand `Beginn` und Zahl) | ohne Fenster: Nr. 2 b | Zahl 0 nur im neuen Fenster, nie nach der Zulassung (Nr. 2 c) | Sprung der Uhr nach vorn: neues Fenster; nach hinten: offen (Nr. 5 c) | Neustart des Prozesses und neuer `Server`: ohne Fenster (Nr. 5 f) | zwei gleichzeitige Abweisungen: genau 20 (Nr. 5 d, W10, W11) |
| `Append` des zugelassenen Eintrags | gilt nicht: die Feldmenge ist fest (`SPEC-238` Nr. 8, `SPEC-234` Nr. 2) | gilt nicht: wie links | der Fehler der Ablage: die Zulassung ist verbraucht, die Antwort bleibt 401 (Nr. 5 e, W13); ein `Append`, das hängt: hält nur diesen Request (W12) | Herunterfahren: ein zugelassener Request schreibt seinen Eintrag (Nr. 5 g, W15) | jede Abweisung ein eigener Aufruf, höchstens 20 je Fenster (Nr. 5 a) |
| `Anhalten` und `Schliessen` mit dem Fenster | gilt nicht: kein Timer, keine Goroutine (Nr. 5 g) | gilt nicht: wie links | ein angehaltener Server holt keine Zulassung, ruft die Uhr nicht (Nr. 5 g, W15) | gilt nicht: das Fenster hat nichts zu beenden (Nr. 5 g) | wiederholtes `Anhalten`: wie `SPEC-238` Nr. 11 |

**Zusatztabelle** (`AGENTS.md` §3.16: die Operation meldet mehrere Verstöße, und Zusagen hängen vom Modus ab; je Paar von Verstößen am selben Schlüssel und je Modus eine Zeile):

| Paar oder Modus | Ergebnis | Fänger |
|---|---|---|
| Verstoß der Authentisierung und `ODR-E2101` einer Route im selben Fenster | beide Wege teilen das Fenster (`SPEC-301` Nr. 5 h) | W09 |
| fehlender und falscher Anlass (`SPEC-238` Nr. 4 a bis c) | jeder zählt gleich, dieselbe Antwort (Nr. 3, 5 h) | W06, W09 |
| jeder Modus (`record`, `replay`, `mock`, `hybrid`) | derselbe Weg, dieselbe Grenze (`SPEC-300` Nr. 5, `SPEC-301` Nr. 8) | W17 mit der Achse Modus |

**Akzeptierte Negative** (`AGENTS.md` §3.15: Grund, Folge und Maß):

1. **Die Grenze des Fensters in Echtzeit am Binary.** Grund: die Tests setzen die Uhr; kein Test wartet 60 s. Folge: ob die Uhr des Prozesses das Fenster öffnet und schließt, belegt allein ein Handlauf über 60 s. Maß: 1 von 2 Seiten der Grenze am Prozess (die Zahl 20 und 21 ja, der Ablauf nein); am Adapter sind beide Seiten gefangen (W03).
2. **Die zwei Sätze des Handbuchs** (`SPEC-301` Nr. 8). Grund: kein Gate prüft Prosa (`make handbuch-check` prüft Beispiele, Kopf und Kataloge). Folge: ein Satz kann fehlen oder veralten, ohne dass etwas rot wird; Reviewer und Verifier lesen ihn. Maß: 2 von 2 Sätzen ohne Fänger.
3. **Die eine Uhr** (Instanz des Roots). Grund: eine zweite Uhr desselben Typs liefert dieselbe Zeit, kein Verhalten unterscheidet sie. Folge: übergibt der Root eine andere Instanz derselben Uhr, bleibt alles grün. Maß: 1 von 2 Mutationen am Root ungefangen (W17 fängt die fehlende Uhr, nicht die zweite).
4. **Der Zeitkanal.** Grund: die zugelassene Antwort wartet auf `Sync` von `Append`, die gedrosselte nicht (`SPEC-301` *Ungeprüft*). Folge: ein Client unterscheidet beide an der Antwortzeit; sie verrät nichts über das Token. Maß: 0 Tests, 1 Kanal.
5. **Ein Datenrennen, das kein Test auslöst.** Grund und Maß: `AGENTS.md` §3.10, die nebenläufigen Zeilen W10 bis W12 tragen Wirkungs-Fänger; ungefangen bleibt ein Zugriff auf das Fenster außerhalb der drei Stellen, den kein Test ausführt (0 bekannte).

Offene Punkte und Risiken:

- Der Netzwerkzugriffstest (`SPEC-041` Nr. 7) zählt 1008 Einträge je Server; mit der Grenze zählte er 20, und die Zeile bliebe grün, ohne das zu belegen, was sie belegen soll — **Ausgang:** entschieden (Architect, 2026-11-24): `SPEC-041` Nr. 7 und 7 c tragen das Vorrücken der Uhr und eine Kontrolle ohne Vorrücken; Zeile W18.
- Tests der Pakete `admin` und `bootstrap`, die eine Zeile je Abweisung erwarten (`TestM6AbweisungOhneAenderung` mit 18 Abweisungen je Handle, zwei Reserven bis 20): wächst eine Menge über 20 je Handle in 60 s Echtzeit, zählt der Test Zeilen nach `SPEC-301` — **Ausgang:** weiter offen; Adresse: Planner beim Start (er liest die Mengen der vier Teile und der Stufe 11 `TestE2EAdmin` gegen 20 je Handle); der Implementer rückt die Uhr vor oder begrenzt die Menge und meldet es im Bericht (`AGENTS.md` §3.13). Gelesen bei der Closure von [slice-admin-workspaces-audit](../done/slice-admin-workspaces-audit.md) (Planner, 2026-11-25): die Tests der Workspace-Endpunkte im Paket `admin` schicken je Server höchstens eine Abweisung (`TestWS14Authentisierung`: 143 Fälle, je Fall ein frischer Server mit genau einem Eintrag `admin_abgewiesen`; kein anderer Test der Dateien `workspaces*_test.go` benutzt `ohneAuth` oder die Verstöße von `wsAuthVerstoesse`); die Menge ist unter 20 je Handle und braucht keinen Nachzug. Die Datei `abweisen_audit_test.go` aus dem Block dieses Slice ändert [slice-admin-workspaces-audit](../done/slice-admin-workspaces-audit.md) (`TestAppendEineAufrufstelle` zählt vier Stellen in `workspaces.go`); ihre Zeilen gehören nach dem Merge nach `main` zur Basis dieses Slice, keine fremden.
- Die Marke: die Vorlage rechnete mit 8 Zeilen, die Tabelle hat 18 — **Ausgang:** entschieden (Architect, 2026-11-24): *S* = 1755, Marke 2400 aus der Tabelle (§4); der Planner bestätigt sie beim Start mit `make groessenmarke`.
- Ein Angreifer verbraucht die 20 Einträge je Fenster, und die Spur eines späteren Angreifers im selben Fenster fehlt im Audit (`SPEC-301` Nr. 9) — **Ausgang:** entschieden (Architect, 2026-11-24, User 2026-11-24): die Folge der Variante A; der Zähler `ODR-E2101` trägt die Zahl.
- Das Raten eines selbst gesetzten Tokens bleibt unbegrenzt (G2) — **Ausgang:** entschieden (User, 2026-11-24): der Satz im Handbuch (`SPEC-301` Nr. 8), keine Änderung des Starts.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks).

*(wird bei der Closure geschrieben)*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** berührt wird die Sub-Area `*` (Default des Repos, Kürzel `REPO`, Modus Greenfield, `harness/conventions.md`); der Slice liegt im Paket `src/adapters/driving/admin` und in einer Zeile von `src/bootstrap`; die Schwelle von zwei Achsen ist für eine zweite Sub-Area nicht erfüllt.

**Vorgelagert — offene Beobachtungen sichten:** das Register (`../observations/`) ist durchgegangen; Treffer zu den Fehlversuchen des Admin-Listeners gibt es nicht (die Frage ist Rückgabe R5, kein Eintrag des Registers); `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` betrifft die Raten der nebenläufigen Zeilen W10 bis W12 (Maß je Prozess, `AGENTS.md` §3.10).

**Modus-Begründungsblock:** alle berührten Sub-Areas GF.
