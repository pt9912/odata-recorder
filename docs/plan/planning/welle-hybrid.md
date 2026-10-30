# Welle welle-hybrid: Hybrid-Modus mit Fallback-Policy

**Lifecycle:** Diese Datei entsteht bei der **Eröffnung** der Welle und liegt
flach unter `docs/plan/planning/`; bei Closure wandert sie per `git mv` nach
`done/` (neben ihre `welle-<Kennung>-results.md`). Der Zustand ist die
Verzeichnis-Position — kein Status-Feld. Diese Welle ist vor ihrem Start angelegt
(Entscheidung des Users vom 2026-10-30); bis ihr Trigger eintritt steht sie in der
Roadmap unter *Nächste Wellen*. Ihre Slice-Dateien liegen seit
2026-10-30 in `open/`; die Pläne tragen ihre Start-Trigger mit Abhängigkeiten.

**Zielmeilenstein:** M4 (siehe [Roadmap](in-progress/roadmap.md)).

**Verantwortlich:** pt9912. **Datum:** 2026-10-30.

---

## 1. Welle-Ziel

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht.

Ein Test startet den Recorder im Modus `hybrid` gegen einen Referenzdienst mit Verbindungszähler und belegt die Entscheidungsmatrix (`LH-FA-12`): ein Treffer des Mock entsteht ohne Upstream-Verbindung; bei unvollständiger Coverage oder nicht unterstützter Operation geschieht nur, was die Policy ausdrücklich freigibt (Replay-Fallback für die konfigurierten Kategorien, Upstream-Fallback nur für lesende Methoden zu einem Host aus `FALLBACK_ALLOW_HOSTS`); ein POST ohne Eintrag in `FALLBACK_WRITE_ALLOW` erreicht den Upstream nie; fachliche `404`, `409`, `412` und ungültige Anfragen lösen keinen Fallback aus, und ein erschöpfter Replay-Key fällt nie still auf den Upstream durch.
Die Welle liefert keinen neuen Rechenkern: `HybridUseCase` ruft `MockUseCase` und `ReplayUseCase` über deren Inbound Ports und den Upstream über `UpstreamClient`.
Eine Welle, weil der Closure-Trigger die Zusage „schreibende Requests werden nie implizit weitergereicht“ über alle Zweige des Zusammenspiels von Mock, Replay und Upstream-Port beobachtet; die DoD keines einzelnen Slice sieht alle Zweige.

## 2. Trigger (Welle startet)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Regeln — ein Trigger ist **beobachtbar** dann, wenn ein *anderer*
Mensch ohne Rückfrage sagen kann, ob er eingetreten ist; ein Datum darf erwähnt
werden, aber nie Trigger sein. Und der **Start**-Trigger ist **kein Ergebnis
dieser Welle**: Steht er in der Slice-Liste unten, ist er falsch platziert.

- `welle-mock` liegt in `done/`: `HybridUseCase` ruft `MockUseCase` über dessen Inbound Port, und `ReplayUseCase` liegt seit `welle-record-replay` vor.
- M3 ist erreicht: der M3-Test aus `slice-bootstrap-m3` ist grün und Teil von `make test`.

**Ausnahme je Slice (Code-Unabhängigkeit):** `slice-fallback-policy` hat keine Code-Abhängigkeit von `welle-mock`, `welle-record-replay` und M2 (reine Entscheidung ohne `MockUseCase`, `ReplayUseCase` und `UpstreamClient`); der Trigger der Welle hält sie formal zurück. Ein Start davor ist eine Umplanung des Planners mit Eintrag im Drift-Log der Roadmap und bleibt, solange sie nicht vorliegt, ausgeschlossen. `slice-hybrid-service` und `slice-bootstrap-m4` hängen an `welle-mock` (§5).

## 3. Closure-Trigger (Welle schließt)

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wann Arbeit eine Welle braucht — der Trigger muss das *Mehr* gegenüber den
einzelnen Slice-DoDs benennen; kann er das nicht, liegt keine Welle vor.

- Alle Slices der Welle liegen in `done/`.
- `make gates` ist grün; `make gegenproben` ist grün.
- Der M4-Test aus `slice-bootstrap-m4` ist grün und Teil von `make test`; der Referenzdienst zählt seine Verbindungen und Methoden: ein Treffer des Mock ergibt 0 Verbindungen; eine lesende Anfrage bei unvollständiger Coverage mit `HYBRID_FALLBACK=enabled` und freigegebenem Host genau eine lesende Verbindung, dieselbe Anfrage ohne Freigabe `ODR-E6003` und 0; ein POST auf eine nicht unterstützte Operation ohne Eintrag in `FALLBACK_WRITE_ALLOW` 0, mit passendem Eintrag eine; ein Replay-Miss `ODR-E5001` mit Opt-in nur für GET und HEAD zum freigegebenen Host, ein erschöpfter Key `ODR-E5002` und fachliche `404` und `409` 0.
- Das Handbuch deckt alles ab, was das Binary dieser Welle kann (Modus `hybrid`, Schlüssel, Fehlercodes); das Kataloge-Gate aus `slice-handbuch-meldungskatalog` ist grün.
- Die Closure-Notiz liegt als `welle-hybrid-results.md` im Ruheort `done/`; ihre *Steering-Loop-Einträge* tragen je verkörperter Regel mit dem Anker `seit welle-hybrid` das Feld `liegt in`.

## 4. Slices in dieser Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Lifecycle als State Machine — der Zustand eines Slice ist sein
Lifecycle-Verzeichnis und wird hier **nicht** gespiegelt.

| Slice | Titel | Bezug |
|---|---|---|
| slice-fallback-policy | Fallback-Policy als reine Entscheidung: Hauptschalter, Allowlist des Zielhosts, lesende Methoden, Routenfreigaben (Methode und Pfadmuster) für mutierende Methoden, Kategorien des Replay-Fallbacks | `LH-FA-12`; `LH-FA-12.a`, `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080` |
| slice-hybrid-service | `HybridUseCase`: Mock zuerst, Entscheidungsmatrix nach `ODR`-Code (nie nach HTTP-Status), Replay-Fallback nur bei `ODR-E5001`, Weiterleitung über `UpstreamClient` nur nach Policy | `LH-FA-12`, `LH-FA-04`, `LH-FA-11`; `LH-FA-12.b`, `SPEC-020`, `SPEC-024` |
| slice-bootstrap-m4 | Modus `hybrid` im Composition Root, Konfiguration (`HYBRID_*`, `FALLBACK_*`), M4-Test mit Verbindungszähler am Referenzdienst | `LH-FA-12`, `LH-QA-02`; `SPEC-110`, `SPEC-075` |

Die Slice-Dateien liegen seit 2026-10-30 in `open/`; die Pläne tragen ihre Start-Trigger mit Abhängigkeiten, und `plan-welle` gleicht sie beim Start gegen den Ist-Stand ab. Marken und Betriebsregeln setzt der Planner beim Start (`SPEC-128`).

**Pflicht Handbuch:** Jeder Slice, dessen Ergebnis ein Anwender am Binary sieht (neues Kommando, Konfigurationsschlüssel, Fehlercode, HTTP-Schnittstelle, Betriebsart), trägt in seiner DoD den Punkt „Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün“. In dieser Welle trägt ihn `slice-bootstrap-m4` (Modus `hybrid`, Schlüssel `HYBRID_*` und `FALLBACK_*`), `slice-fallback-policy` und `slice-hybrid-service` nur, soweit sie einen Fehlercode in den Katalog bringen.

## 5. Abhängigkeiten

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Roadmap-Struktur: fünf Abschnitte.

- Blockiert: keine Welle (`welle-admin-api` zeigt den Modus in der Diagnose an, braucht den Hybrid-Betrieb aber nicht).
- Wird blockiert von: `welle-mock` (`MockUseCase`), mittelbar `welle-record-replay` (`ReplayUseCase`, `UpstreamClient`); `slice-fallback-policy` ausgenommen (keine Code-Abhängigkeit, §2).
- Unabhängig: `welle-betrieb`, `welle-admin-api`, `welle-etag`, `welle-expand-batch`, `welle-odata-v2`; die Reihenfolge nach `welle-mock` ist Priorisierung.
- Reihenfolge der Slices (Pfeil heißt: braucht): `slice-fallback-policy` → `slice-hybrid-service` → `slice-bootstrap-m4`. Kanten nach außen: `slice-fallback-policy` hat keine; `slice-hybrid-service` braucht zusätzlich `slice-mock-lesen`, `slice-mock-schreiben` und `slice-replay-service`; `slice-bootstrap-m4` braucht zusätzlich `slice-bootstrap-m3`. `slice-bootstrap-m4` und `slice-bootstrap-m6` (`welle-admin-api`) ändern beide `src/bootstrap/` und `SPEC-110` Nr. 2 und laufen nacheinander, nie zugleich; welcher zuerst läuft, ist Priorisierung, und der zweite zieht den Stand des ersten nach.
- **Register-Sichtung (Eröffnung Schritt 2):** `observations/` gesichtet, Stand dieses Plans. Treffer, beim Start in den Slice-Plänen (§8) mit dem dann aktuellen Zähler als Risiko zu führen: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes von `UpstreamClient`, `MockUseCase` und `ReplayUseCase` (jede Operation zählen, jedes Argument festhalten, `AGENTS.md` §3.15); `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×), weil die Matrix viele Zweige hat und der Dienst keinen Zustand halten soll; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Kein offener Eintrag erreicht mit dieser Welle 3×.

## 6. Out-of-Scope für diese Welle

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Wellen-Closure-Prozedur, Eröffnung Schritt 1 — Out-of-Scope gehört zur
Zielsetzung: Was nicht ausdrücklich ausgeschlossen ist, dehnt die Welle, bis
der Closure-Trigger unerreichbar wird.

- Speichern einer Fallback-Antwort (als Exchange oder Entity) und jeder automatische Datenabgleich zwischen Mock und Upstream — Out-of-Scope von `LH-FA-12`.
- Umschreiben absoluter Links einer Upstream-Antwort (`LINK_REWRITE`) — nach `SPEC-092` nur `off`; eine Fallback-Antwort kann Upstream-Links tragen, das Risiko steht beim Start im Plan von `slice-hybrid-service`.
- Umschalten des Modus zur Laufzeit — `LH-FA-13` schließt es aus der Erstabnahme aus.
- Endpunkte der Admin-API für den Hybrid-Betrieb — `welle-admin-api`.
- Spans, Audit und Metriken des Fallbacks — `welle-betrieb`.
- Upstream-Zugriff außerhalb des Fallbacks (Aufzeichnen, Weiterleiten) — Modus `record`, `welle-record-replay`.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-traceability.md`
§Herkunfts-Anker für Steering-Loop-Regeln — dort die **Ruheort-Regel**: Die
beiden Zeiger unten sind so zu schreiben, wie sie vom Ruheort `done/` auflösen,
nicht vom Schreibort.

Ergebnis: `welle-hybrid-results.md` (Geschwister im Ruheort `done/`).
Zähler: das Beobachtungs-Register eine Ebene über dem Ruheort (das Verzeichnis observations/).
