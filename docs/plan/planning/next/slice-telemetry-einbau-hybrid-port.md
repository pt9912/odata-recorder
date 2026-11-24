# Slice slice-telemetry-einbau-hybrid-port: Port `FallbackTelemetry`, Span `odata.hybrid.fallback` und Reihe `odr_hybrid_discarded_total` im Port und im Adapter von `Telemetry` (Teil 1 des Schnitts von `slice-telemetry-einbau-hybrid`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-FA-12`, `LH-QA-04`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen** (Architect, 2026-11-23; Teil 1 des Schnitts, die Stellen des Dienstes und der Verdrahtung trägt [slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md)): `SPEC-320` (Zeile des Spans `odata.hybrid.fallback` in `spezifikation.md` §5) · `SPEC-230` Nr. 3, 4, 5 (sechs Namen, dreizehn Pflicht-Attribute), 8, 9, 10 und neue Nr. 11 (Reihe `odr_hybrid_discarded_total`, Port `FallbackTelemetry`, `Snapshot.Verworfen`) · `SPEC-232` Nr. 5 a und g (`CountDiscarded` neben `CountError`) · `SPEC-231` Nr. 6 (Bestand: Span-Zeile) · `SPEC-321` Nr. 5 (der Zähler, den der Dienst in Teil 2 ruft) · `SPEC-200` Nr. 1 und `SPEC-201` (Handbuch und Katalog: kein neuer Fehlercode, Schlüssel oder Kommando) · `spezifikation.md` §5.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md), Vorschlag des Architect vom 2026-11-23). **Datum:** 2026-11-23.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Port und Adapter von `Telemetry` tragen, was der Vertrag `SPEC-320` und `SPEC-321` (Architect, 2026-11-23) für den Hybrid-Fallback braucht, und der Hybrid-Dienst darf es in Teil 2 rufen: die Konstante `SpanHybridFallback` und die Zeile `odata.hybrid.fallback` in `support.Spans()` (sechs Namen, dreizehn Pflicht-Attribute, `recorder.fallback_type` und `recorder.fallback_result`), der Port `FallbackTelemetry` (`Telemetry` und die Operation `CountDiscarded`), `support.VerworfeneCodes()` (`ODR-E5001`, `ODR-E6001`, `ODR-E6002`, `ODR-E6003`), `Snapshot.Verworfen` samt `Kopie()`, `FakeTelemetry.CountDiscarded` und im Adapter die Reihe `odr_hybrid_discarded_total` mit `CountDiscarded` und den zwei Schlüsseln in `regeln.go` (`SPEC-230` Nr. 3 bis 5, 8 bis 11). `Telemetry` bleibt bei drei Operationen, weil eine vierte fünf Fakes der Tests in vier Paketen träfe (`mock`, `replay`, `httpodata`, zweimal `admin`), die der Hybrid-Dienst nicht braucht (Zeile P7); `odr_errors_total` und `SPEC-294` Nr. 5 bleiben, wie sie sind (Entscheidung des Users vom 2026-11-23, Variante A). Das Binary ändert sich nicht: kein Dienst ruft `CountDiscarded` und keiner den Span, bis [slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md) (Teil 2) liegt. **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-23): die Schätzung des ganzen Plans (*S* = 2495) liegt über der Schnittgrenze von `SPEC-128` Nr. 1; Teil 1 (diese Datei) liefert Port und Adapter (Zeilen P1 bis P9), Teil 2 den Dienst und die Verdrahtung (Zeilen D1 bis D18).

**Übernimmt (Sendung von `welle-hybrid` §6, Spans und Zähler des Hybrid-Fallbacks, Teil Port und Adapter; Geber: `welle-hybrid`, in den Plänen [slice-telemetry-port](../done/slice-telemetry-port.md), [slice-telemetry-spans](../done/slice-telemetry-spans.md), [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md), [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md), [slice-telemetry-einbau-mock-mutation](../done/slice-telemetry-einbau-mock-mutation.md) und [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md) mit Ausgang *weiter offen* geführt; Eintrag des Planners vom 2026-11-23 beim Schnitt, `AGENTS.md` §3.13; §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen, kein Ausschluss trifft die Sendung):** die Zeile des Spans in der Menge der Spans, der Port `FallbackTelemetry`, die Reihe und der Stand `Snapshot.Verworfen`. Die Pläne der Geber liegen in `done/` und nennen den Nehmer nicht; der Nehmer des Dienstes und der Verdrahtung ist [slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md) (Teil 2).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Dienst: `hybrid.Konfig.Telemetrie`, die Aufrufe von `CountDiscarded`, `StartSpan`, `Set` und `End` an den Fallbacks, der Journal-Fake im Paket `hybrid` und die Verdrahtung des Adapters in `verdrahteHybrid` — [slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md) (Teil 2, Zeilen D1 bis D18; §1 und DoD dort tragen die Sendung, `AGENTS.md` §3.13). Teil 2 importiert `support.FallbackTelemetry` und `SpanHybridFallback`; ohne diesen Teil übersetzt das Paket `hybrid` nicht. Das Paket `hybrid` ändert dieser Teil nicht.
- Der Gesamtspan, der Zähler je Fehlercode und die Übergabe des Adapters des Prozesses an den HTTP-Adapter — [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md) (liegt in `done/`). Nach Entscheidung des Users vom 2026-11-23 (Variante A) zählt allein der HTTP-Adapter, was ein Client als Fehlerobjekt erhielt (`SPEC-294` Nr. 5); die Reihe `odr_hybrid_discarded_total` ist eine eigene Reihe, `CountError` und `Snapshot.Errors` bleiben, wie sie sind (Zeile P7).
- Der Query-Span des Versuchs im Mock und der Match-Span des Versuchs im Replay — [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md) und [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md); beide liegen in `done/`, dieser Slice ändert die Pakete `mock` und `replay` nicht.
- Der Capture-Span und die Einträge der Dienste `record`, `session` und `maintenance` — [slice-telemetry-einbau-record](../done/slice-telemetry-einbau-record.md) (liegt bei diesem Start in `done/`); der Dienst `record` gehört nicht zum Modus `hybrid` (`SPEC-295` Nr. 7 f).
- Port und Adapter `Telemetry` als Bauteil: Zeile, Stufen, Senke, `Handler`, Verteilung und Ring — [slice-telemetry-port](../done/slice-telemetry-port.md), [slice-telemetry-log](../done/slice-telemetry-log.md), [slice-telemetry-handler](../done/slice-telemetry-handler.md), [slice-telemetry-spans](../done/slice-telemetry-spans.md). Dieser Teil fügt allein hinzu, was `SPEC-230` nennt, und ändert keinen bestehenden Vertrag des Adapters: `Telemetry` bleibt bei drei Operationen (Zeile P7).
- Das Handbuch-Kapitel über die Logzeilen des Prozesses — [slice-handbuch-telemetrie](slice-handbuch-telemetrie.md); der Slice schreibt kein Kapitel und führt keinen Fehlercode, Schlüssel oder Kommando ein (`SPEC-200` Nr. 1; siehe DoD und §6).
- Latenz des Fallbacks im Lasttest — [slice-lasttest](slice-lasttest.md) nimmt den Hybrid-Fallback aus (`LH-QA-01.a`).
- Die Art der Policy und ihre Entscheidungen (`SPEC-261`) — Bestand von [slice-fallback-policy](../done/welle-hybrid/slice-fallback-policy.md); der Span beobachtet sie und ändert kein Ergebnis.

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

- [ ] Menge der Spans mit dem Span des Fallbacks: `SpanHybridFallback`, die sechste Zeile in `support.Spans()` (sechs Namen, dreizehn Pflicht-Attribute), die Schlüssel `recorder.fallback_type` und `recorder.fallback_result` mit der Form Token und je Span zulässig, der Adapter nimmt, verteilt und schreibt die Zeile; die Tests, die die Zahlen der Menge festhalten, ziehen um (Zeilen P1 bis P3, vom Architect).
- [ ] Zähler der verworfenen Versuche im Port und im Adapter: der Port `FallbackTelemetry` (`Telemetry` und `CountDiscarded`), `VerworfeneCodes()`, `Snapshot.Verworfen` samt `Kopie()`, `FakeTelemetry.CountDiscarded` und die Reihe `odr_hybrid_discarded_total` mit der Sperre nach `SPEC-232` Nr. 5 a und g; `Telemetry` bleibt bei drei Operationen, `CountError` und `Snapshot.Errors` bleiben, wie sie sind (Zeilen P4 bis P9, vom Architect). Den Dienst und die Verdrahtung liefert Teil 2, [slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-230`, `SPEC-232`, `SPEC-320` stehen vor dem Start auf `main`; dieser Teil schreibt keine Spezifikation).
- [ ] Handbuch: der Slice führt keinen Fehlercode, keinen Schlüssel und kein Kommando ein; das Binary schreibt mit ihm keine neue Zeile (kein Dienst ruft den Span); das Kataloge-Gate ist grün, und §7 sagt, dass kein Fehlercode, Schlüssel oder Kommando des Binary dazukam.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/support/` (mit `supporttest/`) und `src/adapters/driven/telemetry/` | update | Konstante `SpanHybridFallback` und Zeile in `support.Spans()`, `FallbackTelemetry`, `VerworfeneCodes()`, `Snapshot.Verworfen` samt `Kopie()`, `FakeTelemetry.CountDiscarded`; Adapter: Reihe `odr_hybrid_discarded_total`, `CountDiscarded`, die zwei Schlüssel in `regeln.go`; die Tests der Zahlen (unten) ziehen um |
| `spec/spezifikation.md` und `spec/architecture.md` | geliefert (Architect, 2026-11-23) | `SPEC-320`, `SPEC-321` und Nachzug (`SPEC-230`, `SPEC-232`); dieser Plan schreibt sie nicht |
| Tests | neu, update | Tests am Adapter, am Fake und an der Menge der Spans (Zeilen P1 bis P9); die Tests des Dienstes und durch den Root liegen in Teil 2 |

Der Vertrag steht in der Spezifikation (`SPEC-320` und `SPEC-321` samt Nachzug, Architect, 2026-11-23, Zweig `spec-hybrid-vertrag`; er liegt auf `main`). Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15; Zeilen P1 bis P9 des ungeschnittenen Plans), die Mengen mit Zählprüfung und die akzeptierten Negative stehen unten, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16), vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidungen des Architect** (2026-11-23; die Aufträge 1 bis 6 des ungeschnittenen Plans mit ihrem Ort in `spec/spezifikation.md`; hier die Teile, die Port und Adapter betreffen, der Rest steht in Teil 2):

1. **Span des Fallbacks** (`SPEC-320`): `odata.hybrid.fallback` (Konstante `SpanHybridFallback`, Zeile in §5, Eintrag in `support.Spans()`: sechs Namen, dreizehn Attribute) mit `recorder.fallback_type` (`replay`, `upstream`; Token) und `recorder.fallback_result` (`antwort`, `fehler`, `abgebrochen`; Token). Beginn, Ende und die Aufrufstellen im Dienst: Teil 2.
2. **Zähler der verworfenen Versuche** (`SPEC-230` Nr. 8 und 11): eigene Reihe `odr_hybrid_discarded_total` (Zähler, Einheit 1), Schlüssel die vier Codes von `support.VerworfeneCodes()` (`ODR-E5001`, `ODR-E6001`, `ODR-E6002`, `ODR-E6003`, die Auslöser der Policy), im Stand `Snapshot.Verworfen`. Die Operation `CountDiscarded` steht hinter dem neuen Port `FallbackTelemetry` (`Telemetry` und diese Operation); `Telemetry` bleibt bei drei Operationen, weil eine vierte fünf Fakes der Tests in vier Paketen träfe (`mock`, `replay`, `httpodata`, zweimal `admin`), die der Hybrid-Dienst nicht braucht. **`odr_errors_total` und `SPEC-294` Nr. 5 bleiben unverändert** (Entscheidung des Users vom 2026-11-23, Variante A): ein Fehler mit Code steigt in genau einer der beiden Reihen. Wer `CountDiscarded` ruft: Teil 2.
3. **Nachzug im Vertrag** (geliefert, `spezifikation.md` §7 mit Datum 2026-11-23): `SPEC-230` Nr. 3, 4, 5, 8, 9, 10 und 11, `SPEC-232` Nr. 5 a und g, `spezifikation.md` §5 und die Zeilen zum Port in der Sicht (`spec/architecture.md`, Tabelle der Ports und `ARC-021`); die übrigen Stellen des Nachzugs (`SPEC-284`, `SPEC-139`, `SPEC-290`, `SPEC-294`, `SPEC-296`) betreffen den Dienst (Teil 2).
4. **Eingaben der Marke** (`SPEC-128` Nr. 3 k): der Teil des Architect im Block und die Klassen der Tabelle stehen in §4 und unten (Spalte *Klasse*).

**Tests, die die Zahlen der Menge der Spans festhalten** (`git grep` auf `main` am 2026-11-23; der Implementer sucht erneut und trägt jede weitere hier nach, `AGENTS.md` §3.9): `src/hexagon/ports/support/telemetry_test.go` (Zeilen 65 und 66: 5 Spans, 11 Attribute; Zeile 87: 6 Slices), `src/adapters/driven/telemetry/span_hilfen_test.go` (Zeilen 25, 45 und 115: elf), `src/adapters/driven/telemetry/regeln_test.go` (Zeile 175: 24 Schlüssel, 144 Paare, 11 in Spans), `src/adapters/driven/telemetry/telemetry_test.go` (Zeilen 49 und 106: Zahl der Namen und der fremden Schlüssel) und `src/hexagon/ports/support/supporttest/telemetrie.go` (leerer Stand mit vier Sammlungen, Zeilen 121 bis 123). Der sechste Test, `src/bootstrap/telemetrie_replay_test.go` (`TestWurzelReplayMatchLiegtImGesamtspan`, Fall `hybrid`: 3 Zeilen, nach dem Einbau 5), gehört Teil 2 (Zeile D12).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-23, Zeilen P1 bis P9 des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). `Nr.` ist die Kennung des ungeschnittenen Plans: **P** gehört zu Teil 1 (Port und Adapter), **D** zu Teil 2 ([slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md)). Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| P1 | (`SPEC-320`, `SPEC-230` Nr. 5) `support.Spans()` führt sechs Spans mit zusammen dreizehn Pflicht-Attributen, der sechste `odata.hybrid.fallback` mit `recorder.fallback_type` und `recorder.fallback_result`; die Konstante `SpanHybridFallback` trägt den Namen | Test im Paket `support`: `Spans()` gleich dem Literal aus §5 (6 Namen, 13 Attribute, Reihenfolge der Tabelle), Konstante gleich dem Literal, frische Slices (1 + 6 gleich 7) | Zeile fehlt; Attribut vertauscht; Name anders geschrieben; Konstante abweichend | andere |
| P2 | (Nr. 5; `SPEC-231` Nr. 6) Der Adapter nimmt `odata.hybrid.fallback` mit beiden Attributen ohne Verwurf an, zeichnet Verteilung und Eintrag auf und schreibt die Zeile in der Reihenfolge der Tabelle | Test am echten Adapter: Beginn mit `fallback_type`, ein `Set` mit `fallback_result`, `End`: `Dropped` alle 0, `Durations` mit 6 Einträgen zu je 12 Eimern, Zeile als Literal (`timestamp`, `level`, `msg`, `ereignis`, `span_id`, `parent_id`, `duration_ms`, beide Attribute); Zählprüfung 6 gegen §5 | Schlüssel nicht in der Menge des Spans; Verteilung für den Namen nicht angelegt; Reihenfolge der Attribute vertauscht | andere |
| P3 | (`SPEC-230` Nr. 3 bis 5) Die zwei Schlüssel haben die Form Token, ein Verstoß zählt `wert` (am Ende `pflicht`), und jeder Schlüssel ist allein am eigenen Span zulässig | Menge M-Schlüssel: 2 Schlüssel × 3 Verstöße (leer, Großbuchstabe, Byte ab 0x80) gleich 6 mit `wert` und je `pflicht`; Matrix 6 Spans × 13 Schlüssel gleich 78 Paare, 13 zulässig, 65 fremd (`schluessel`); Paar von Verstößen (beide Werte ungültig: `wert` 2, `pflicht` 2); Orakel Literal aus §5; Zählprüfung 13, 6 und 78 | die zwei Schlüssel als Form Kennung; ein Schlüssel des Fallback-Spans am Span `odata.mock.query` zulässig | andere |
| P4 | (`SPEC-230` Nr. 11) `CountDiscarded(c)` erhöht den Zähler von `c` aus `support.VerworfeneCodes()` um 1; jeder andere Code erhöht keinen Zähler und zählt `fehlercode`; `Errors` bleibt unverändert | Menge M-Codes: die 4 Codes der Menge (je 1, dann 2) und 37 Codes der Tabelle außerhalb, `0`, `-1` und `42` gleich 44 Eingaben; Orakel: `Verworfen` und `Dropped[fehlercode]` als Literal, `Errors` mit 41 Nullen; Zählprüfung 41 gleich 4 + 37, 4 gleich `len(VerworfeneCodes())` und die Reihenfolge `ODR-E5001`, `ODR-E6001`, `ODR-E6002`, `ODR-E6003` | ein Code der Tabelle außerhalb als Zähler zugelassen; `CountDiscarded` ruft `CountError`; Reihenfolge der Codes vertauscht | andere |
| P5 | (Nr. 9) `Snapshot().Verworfen` hat je Code der Menge einen Eintrag, auch mit 0, ist nie nil und eine tiefe Kopie; der leere Stand von `FakeTelemetry` hat fünf leere, nicht-nil Sammlungen | Test: Stand eines frischen Adapters (4 Einträge, alle 0), nach einer Zählung, eine Änderung der Kopie wirkt nicht zurück, `Kopie()` lässt eine nil-Sammlung nil, leerer Stand des Fakes mit 5 Sammlungen; Zählprüfung 5 und 4 | Eintrag nur nach einer Zählung; `Verworfen` teilt die Map mit dem Adapter; `Kopie()` ohne `Verworfen` | andere |
| P6 | (Nr. 10) `FakeTelemetry` erfüllt `FallbackTelemetry`, hält jeden Aufruf von `CountDiscarded` unverändert fest und zählt ihn | Test im Paket `supporttest`: Aufruf, Argument (Code), Reihenfolge im Protokoll neben `CountError`, Zähler der Operation, kopierte Aufzeichnung; `var _ support.FallbackTelemetry` für den Fake und für `*telemetry.Adapter` | Aufruf nicht festgehalten; Code verändert; Operation nicht gezählt | andere |
| P7 | (Nr. 11, `SPEC-230` Nr. 1) `Telemetry` bleibt bei drei Operationen; die beiden Reihen sind getrennt | Test: `reflect` zählt 3 Methoden an `support.Telemetry` und 4 an `support.FallbackTelemetry`; `CountError` ändert `Verworfen` nie (41 Codes) und `CountDiscarded` ändert `Errors` nie (4 Codes); Zählprüfung 3, 4, 41 und 4 | vierte Operation in `Telemetry`; `CountError` zählt in `Verworfen` | andere |
| P8 | (`SPEC-232` Nr. 5 a) `CountDiscarded` verliert keinen Zuwachs unter gleichzeitigem Aufruf | Wirkungs-Fänger und Detektor: 16 Goroutinen × 4 Codes × 1000 Zählungen, Summe je Code 16000, mit gleichzeitigem `Snapshot`; unter `make test-race` je Prozess (`-test.count=1`, rot in jedem Prozess) und `-count=20` für die Summe; Rate je Prozess | Zähler ohne Sperre; `Verworfen` im `Snapshot` ohne Sperre gelesen | nebenläufig |
| P9 | (`SPEC-232` Nr. 5 g) `CountDiscarded` nimmt die Sperre der Zeile nicht: eine Senke, die blockiert, hält `CountDiscarded` nicht an | Test: eine blockierende Senke hält ein `Log` an der Schranke, `CountDiscarded` kehrt für alle 4 Codes zurück; das Muster der Prüfung für `CountError`; Rate je Prozess | `CountDiscarded` unter der Sperre der Zeile | nebenläufig |

Zeilen: P1 bis P9 gleich 9 (*a* = 2: P8, P9; *b* = 7); die akzeptierten Negative stehen unter der Tabelle und zählen nicht als Zeile (`SPEC-128` Nr. 3 e gilt für Zeilen). Die Zeilen D1 bis D18 des ungeschnittenen Plans stehen in Teil 2; die Summe beider Teile ist die Tabelle des Architect mit 27 Zeilen (*a* = 5, *b* = 22).

**Mengen mit Zählprüfung** (Generator oder Schleife, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15):

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| M-Spans (P1 bis P3) | 6 Spans, 13 Attribute, 78 Paare Span × Schlüssel (13 zulässig) | Literal aus §5 | 6, 13 und 78 |
| M-Codes (P4) | 4 Codes der Menge + 37 außerhalb + 3 Randwerte gleich 44 | `Verworfen`, `Dropped[fehlercode]` | 41 gleich 4 + 37 |
| M-Zuwachs (P8, P9) | 16 Goroutinen × 4 Codes × 1000; 4 Codes an der Schranke | Summe 16000 je Code | 4 |

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; die Zeilen zu Modi und Paaren am Dienst stehen in Teil 2):

| Paar oder Modus | Erwartung | Fänger |
|---|---|---|
| beide Attribute ungültig (Form) | `wert` 2, am Ende `pflicht` 2, der Span wird aufgezeichnet | P3 |

**Akzeptierte Negative dieses Teils** (Grund, Folge und Maß; die Nummern des ungeschnittenen Plans in Klammern, die übrigen Negative führt Teil 2): (6) Die Zahlen der Mengen dieses Teils (6 Namen und 13 Attribute der Menge der Spans, 4 Codes von `VerworfeneCodes()`) gegen die Tabellen der Spezifikation prüft kein Test, weil die Stufe `source` des Images `spec/` nicht trägt; Folge: ändert jemand `SPEC-230` Nr. 5 oder 11 ohne den Test, sieht es erst ein Review; Maß: 3 Zahlen, vom Architect am 2026-11-23 von Hand gezählt (die Zahlen des Dienstes stehen in Teil 2). (7) Der Zähler hat im Produkt keinen Leser (`SPEC-321` *Ungeprüft*): wie jede Metrik von `SPEC-230` (es gibt keinen Abrufweg); Folge: ein Betreiber sieht verworfene Versuche an den Zeilen der Spans; Maß: 0 von 1 Leser außer `TelemetryReader`. (8) Das Fenster in `Snapshot` zwischen dem Kopieren von `Errors` und `Verworfen` (beide unter `mu`): kein Vertrag verbindet beide Stände (anders als `Count` und `Counts`, `SPEC-230` Nr. 9), so dass ein Freigeben dazwischen keine Zusage bricht; Maß: 1 von 1 Fenstern ohne Zusage; Haltepunkt: unter der Sperre liegt kein Aufruf eines übergebenen Arguments. (9) Ein Datenrennen, das kein Lauf auslöst (`SPEC-122` Nr. 8 und 9): ein Zugriff, den kein Lauf ausführt, und Prüfen und Handeln in getrennten Abschnitten bei lauter geordneten Zugriffen; Folge wie `SPEC-232` *Ungeprüft*; Maß: der Adapter hat einen Abschnitt je Operation unter `mu`; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×).

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-telemetry-spans](../done/slice-telemetry-spans.md) und [slice-telemetry-port](../done/slice-telemetry-port.md) (Port, Fake und Adapter mit der Menge der Spans) liegen in `done/`, und [slice-telemetry-einbau-record](../done/slice-telemetry-einbau-record.md) liegt in `done/` (*Reihenfolge*: sein Block nennt `src/adapters/driven/telemetry/` ebenfalls, nur Tests; dazu ändert es `wurzel.go`, das Teil 2 danach ändert). Der Vertrag samt Tabelle, Raster und Mengen steht in `spec/spezifikation.md` (`SPEC-320`, `SPEC-321`, Nachzug) und in §3 und §6 dieses Plans auf `main`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Teil 1 hängt an keinem Code des Hybrid-Dienstes; die Bedingungen des Dienstes ([slice-hybrid-service](../done/welle-hybrid/slice-hybrid-service.md), [slice-hybrid-service-upstream](../done/welle-hybrid/slice-hybrid-service-upstream.md), [slice-bootstrap-m4](../done/welle-hybrid/slice-bootstrap-m4.md), die Einbau-Slices von Gateway, Mock und Replay) gelten für Teil 2.

**Abhängigkeiten — Grund je Eintrag:**

- [slice-telemetry-spans](../done/slice-telemetry-spans.md) und [slice-telemetry-port](../done/slice-telemetry-port.md): *Code-Abhängigkeit*. Port, Fake und Adapter mit der Menge der Spans, die dieser Teil um eine Zeile und um den Port `FallbackTelemetry` erweitert.
- [slice-telemetry-einbau-record](../done/slice-telemetry-einbau-record.md): *Reihenfolge*, keine Code-Abhängigkeit. Sein Block führt `src/adapters/driven/telemetry/` (nur Tests, `handler_test.go`); dieser Teil ändert dort Dateien und Tests des Adapters, und `wurzel.go` ändert erst Teil 2. Reihenfolge nach dem Vorschlag des Architect: `slice-telemetry-einbau-record` → Teil 1 → Teil 2 → Teil A von [slice-bootstrap-m6](slice-bootstrap-m6-endpunkte.md).
- Teil 1 → Teil 2: *Code-Abhängigkeit*. [slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md) importiert `support.FallbackTelemetry` und die Konstante `SpanHybridFallback`; ohne diesen Teil übersetzt das Paket `hybrid` nicht. Teil 2 startet erst, wenn dieser Teil in `done/` liegt.

**Größenmarke** (`SPEC-128` Nr. 1 und 10 g, `AGENTS.md` §3.14; Teil des Architect 2026-11-23 im Block unten; der Planner hat die Zahlen aus der Tabelle nachgerechnet, **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 9 Zeilen der Tabelle (P1 bis P9), *a* = 2 (P8, P9), *b* = 7, *f* = 0 (der Fake `FakeTelemetry` besteht, seine Operation ist Produktcode des Testpakets, kein neuer Port-Fake und kein Test-Server), *p* = 60 (Konstante, Zeile in `Spans()`, Port, `VerworfeneCodes()`, `Snapshot.Verworfen` und `Kopie()`, `FakeTelemetry.CountDiscarded`, Reihe, `CountDiscarded` und zwei Schlüssel im Adapter): *S* = 85·2 + 80·7 + 200·0 + 60 = 170 + 560 + 0 + 60 = 790, 135·790 = 106650, **Marke 1100** = 100 · ⌈10,665⌉; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`. Der Teil liegt 1433 unter der Schnittgrenze; jede weitere Zeile der Tabelle kostet 80. Summe der Teile: 790 + 1705 (Teil 2) = 2495, gleich der Schätzung des ungeschnittenen Plans. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

```groessenmarke
pfad: src/hexagon/ports/support/
pfad: src/adapters/driven/telemetry/
produktcode: 60
fakes: 0
```

**Pfade und fremde Zeilen** (`SPEC-128` Nr. 4 g und Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`): Der Block nennt zwei Verzeichnisse. `src/hexagon/ports/support/` (mit `supporttest/`) nennt kein anderer Plan in `open/`, `next/` oder `in-progress/` in seinem Block (am 2026-11-23 gelesen); `src/adapters/driven/telemetry/` führt [slice-telemetry-einbau-record](../done/slice-telemetry-einbau-record.md) ebenfalls im Block (nur Tests); es liegt bei diesem Start in `done/` und seine Zeilen gehören zur Basis, keine fremden. Landet ein anderer Slice vor diesem Teil auf `main` und ändert eine Datei des Blocks, zählt seine Zeile mit; der Implementer misst vor dem Merge, nennt die fremden Zeilen in der Messpunkt-Zeile und trägt, wenn die Messung über der Marke liegt, `ueberschreitung` mit Grund ein (`SPEC-128` Nr. 5 d), keine Rückführung. Legt der Implementer eine Datei außerhalb der zwei Verzeichnisse an oder ändert er eine, nennt er sie im selben Commit im Block.

**Reihenfolge und Kollisionen** (Planner, 2026-11-23, nach dem Vorschlag des Architect; Prüfung `git diff --name-only main...<Zweig>`): (1) [slice-telemetry-einbau-record](../done/slice-telemetry-einbau-record.md) (in Arbeit; Zweig `slice-telemetry-einbau-record`): ändert unter `src/adapters/driven/telemetry/` allein `handler_test.go`; dieser Teil startet nach seinem Merge. (2) [slice-admin-tls](../done/slice-admin-tls.md) (in Arbeit): ändert keine Datei der zwei Verzeichnisse. (3) Die Gate-Kette (Übergangsregeln Nr. 7) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`. Der Planner prüft beim Start erneut.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start nach dem Schnitt setzt (`AGENTS.md` §3.14); ein weiterer Schnitt ist nicht vorgesehen (der Teil liegt 1433 unter der Schnittgrenze).
- `in-progress` → `open` (blockiert): `FakeTelemetry` oder die Menge der Spans lassen sich nicht erweitern, ohne einen bestehenden Vertrag der Fakes der Tests zu brechen (`SPEC-230` Nr. 1: `Telemetry` bleibt bei drei Operationen, so dass kein Fake der Pakete `mock`, `replay`, `httpodata` und `admin` ändert).

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

- Der Vertrag fehlt: weder der Span des Fallbacks noch der Zähler der verworfenen Versuche steht in der Spezifikation; ohne Vertrag gibt es keine Tabelle, kein Raster und keine Marke (`AGENTS.md` §3.12) — **Ausgang:** entfallen (Architect, 2026-11-23): `SPEC-320` und `SPEC-321` stehen mit Nachzug auf `main` (Kopf und §3 nennen die Stellen), Tabelle, Mengen und Negative stehen in §3, das Raster unten, die Marke in §4.
- Der Zähler der verworfenen Versuche und die Zusage von `SPEC-294` Nr. 5 (Variante A, Entscheidung des Users vom 2026-11-23: der Adapter zählt, was ein Client als Fehlerobjekt erhielt, kein Dienst zählt zusätzlich) stehen gegeneinander, wenn der Zähler `odr_errors_total` meint — **Ausgang:** entschieden (Architect, 2026-11-23; Verdikt: die Entscheidung gilt und der Plan hat nichts Falsches behauptet): der Zähler ist eine eigene Reihe `odr_hybrid_discarded_total` mit eigener Operation hinter dem Port `FallbackTelemetry` (`SPEC-230` Nr. 11), `CountError` und `Snapshot.Errors` bleiben; Zeile P7 trägt die Trennung der Reihen, die Zeile D9 (Teil 2) die Zusage, dass der Dienst `CountError` nie ruft. Eine Frage an den User entsteht nicht.
- Die Menge der Spans ist geschlossen (fünf Namen, elf Attribute, `SPEC-230` Nr. 5 und `support.Spans()`); ein sechster Name ändert die Tests, die die Zahlen festhalten, und die Aussage in §5 von `spezifikation.md` — **Ausgang:** entschieden (Architect, 2026-11-23): sechs Namen, dreizehn Attribute; §5 und `SPEC-230` Nr. 5 sind nachgezogen, die Tests, die die Zahlen festhalten, stehen in §3 (Liste mit Stand `main`, 2026-11-23); der Implementer sucht sie mit `git grep` erneut und trägt jede weitere in §3 nach (`AGENTS.md` §3.9).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-23): das Raster unten trägt in jeder Zelle einen Ort oder `gilt nicht` mit Grund; eine Randform, die der Implementer im Code findet, geht an den Architect (`AGENTS.md` §3.12).
- Der Adapter trifft einen nebenläufigen Pfad (Zähler und Verteilung unter einer Sperre): `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (10×), `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (8×) — **Ausgang:** entschieden (Architect, 2026-11-23): jede Zusage über Nebenläufigkeit hat einen Wirkungs-Fänger (P8, P9) oder steht als akzeptiertes Negativ mit Grund, Folge und Maß in §3 (die Negative 8 und 9); den Fänger gegen das Fehlen der Sperre trägt der Detektor von `make test-race` je Prozess; die Raten je Prozess misst der Implementer (`AGENTS.md` §3.10). Die Fakes der Tests melden eine Verletzung mit `t.Errorf` oder `t.Fatalf`, statt zu paniken, und jede Schranke trägt eine Frist, so dass eine Mutation `--- FAIL` des benannten Tests erzeugt und nicht `panic:` oder das Zeitlimit des Testprogramms (P9).
- Handbuch: Teil 1 ändert das Binary nicht (kein Dienst ruft den Span oder den Zähler); die Zeile des Fallbacks braucht einen Upstream oder importierte Austausche und ist als gebundenes Beispiel nicht erreichbar (`SPEC-200` Nr. 6 d, 7 b) — **Ausgang:** entschieden (Planner, 2026-11-23): wie die Zeilen von `odata.record.capture` und `odata.replay.match` in [slice-handbuch-telemetrie](slice-handbuch-telemetrie.md) (§1 dort): kein Kapitel, kein Fehlercode, Schlüssel oder Kommando in diesem Slice; ob das Kapitel später gebundene Beispiele für Zeilen mit Upstream bekommt, ist die Frage in `welle-betrieb` §5, Punkt c (Adresse User).
- Schnitt und Größenmarke: Plan vor dem Schnitt ohne Tabelle, keine Schätzung — **Ausgang:** entschieden (Planner, 2026-11-23): die Tabelle des Architect ergibt *S* = 2495 und die Schnittfrage (*S* ≥ 2223); der Planner hat Teil 1 (diese Datei, *S* = 790, Marke 1100) und Teil 2 ([slice-telemetry-einbau-hybrid](slice-telemetry-einbau-hybrid.md), *S* = 1705, Marke 2400) aus der Tabelle nachgerechnet, die Zeilen P1 bis P9 und D1 bis D18 im Wortlaut verteilt (jede Zeile in genau einem Teil) und trägt beide Teile in `welle-betrieb` §4 und §5 und in die Roadmap ein; `basis`, `schaetzung` und `marke` setzt er je Teil beim Start vor dem ersten Code-Commit (`AGENTS.md` §3.14).
- Gemeinsame Datei oder Verzeichnis mit anderen Slices (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`): `src/adapters/driven/telemetry/` führt [slice-telemetry-einbau-record](../done/slice-telemetry-einbau-record.md) im Block (nur Tests) — **Ausgang:** entschieden (Planner, 2026-11-23): Start-Bedingung ist `slice-telemetry-einbau-record` in `done/` (§4); der Zweig `slice-admin-tls` ändert keine Datei der Verzeichnisse.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23, die Zeilen des Ports und des Adapters; die Zeilen des Dienstes stehen in Teil 2; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund; „Nr.“ ohne Nennung heißt `SPEC-321`, bei Zeilen des Adapters `SPEC-230`):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch (Client, `ctx`, `Schliessen`) | Wiederholung | Herunterfahren |
|---|---|---|---|---|---|---|
| `CountDiscarded` am Adapter (`SPEC-230` Nr. 11) | Nr. 11 (der Nullwert `0` ist kein Code der Menge: `fehlercode`, kein Zähler steigt) | gilt nicht: der Code ist eine Zahl, die Operation hat kein anderes Argument | Nr. 11 (jeder andere Code, auch einer der Tabelle von `SPEC-109`, negativ, größer als der letzte: `fehlercode`) | gilt nicht: die Operation hat kein `ctx`, liest keine Uhr und kehrt immer zurück (Nr. 11) | Nr. 11 (jeder Aufruf erhöht um 1, kein Zusammenfassen) | gilt nicht: der Adapter hält keinen Hintergrundlauf; eine blockierende Senke hält die Operation nicht an (`SPEC-232` Nr. 5 g) |
| Span im Adapter und `support.Spans()` (`SPEC-230` Nr. 5, `SPEC-320`) | Nr. 5 (fehlt beim `End` ein Pflicht-Attribut: `pflicht` je fehlendes, der Span wird aufgezeichnet) | Nr. 3 und 4 (ein leerer Wert ist ein Verstoß der Form: `wert`, das Attribut entfällt) | Nr. 5 (ein Name außerhalb der sechs: `spanname`; ein Schlüssel, der kein Pflicht-Attribut des Spans ist: `schluessel`) | Nr. 1 (`End` wirkt auch bei beendetem `ctx`) | Nr. 7 (ein zweites `End` und jedes `Set` nach dem `End` zählen `zustand`) | gilt nicht: der Adapter kennt kein Herunterfahren; ein Span, der nie endet, wird nicht aufgezeichnet (Nr. 7) |

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

*Wird bei der Closure gefüllt.*

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-23; der Zähler ist die Zahl der Dateien unter `evidence/`, der Planner zieht Treffer und Zähler beim Start nach). Der Adapter trifft einen nebenläufigen Pfad: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (8×); `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (10×) gilt für jede Wirkung, die der Detektor nicht meldet. Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (77×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (72×, `AGENTS.md` §3.17), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (43×, `AGENTS.md` §3.9) für den Nachzug der Zahlen in §3 und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (22×). Keine Beobachtung steht mit diesem Slice offen unter 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
