# Slice slice-telemetry-einbau-hybrid: Span des Fallbacks und Zähler der verworfenen Versuche im Hybrid-Dienst und Verdrahtung des Adapters des Prozesses im Modus `hybrid`

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

**Berührte Spec-Stellen** (Kandidaten des Planners; der Architect legt sie mit dem Vertrag fest): `SPEC-284` (Hybrid-Dienst: Konfig, Ablauf, Replay- und Upstream-Fallback) · `SPEC-261` (Fallback-Policy: wann ein Fallback beginnt) · `SPEC-296` Nr. 5, 7 e und *Ungeprüft* sowie `SPEC-290` Nr. 7 g und *Ungeprüft* (der Versuch trägt seinen Span; der Span des Fallbacks und der Zähler der verworfenen Versuche stehen dort als *Ungeprüft*) · `SPEC-294` Nr. 5 und 7 g (Zähler je Fehlercode: was der Hybrid-Dienst verwirft, zählt nirgends; der Versuch liegt im Gesamtspan) · `SPEC-230` Nr. 5 und 8 (Menge der Spans als Daten in `support.Spans()`, Metriken) · `SPEC-231` Nr. 6 (Span-Zeile) · `SPEC-232` Nr. 2 (sensible Inhalte) · `SPEC-139` Nr. 12 a und 13 (Verdrahtung des Hybrid-Dienstes, ein Adapter je Prozess) · `SPEC-136` (der HTTP-Adapter schreibt das Fehlerobjekt, das der Zähler zählt) · `spezifikation.md` §5 (Tabelle der Spans) · `SPEC-320` bis `SPEC-339` (Bereich der Linie F, vom Architect vor dem Code zu vergeben, Übergangsregeln Nr. 6).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Entscheidung des Users vom 2026-11-23 zu [welle-betrieb](../welle-betrieb.md) §5, Punkt b, nach der Empfehlung des Architect in §6 von [slice-telemetry-port](../done/slice-telemetry-port.md)). **Datum:** 2026-11-23.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Hybrid-Dienst meldet seinen Fallback. Im Modus `hybrid` antwortet der Mock zuerst, und nur bei einem Code, den die Fallback-Policy freigibt, folgt der Replay-Fallback oder der Upstream-Fallback (`SPEC-284`, `SPEC-261`). Der Versuch (der Aufruf von `Mock`, der Aufruf von `Replay`) trägt schon seinen Span (`SPEC-290` Nr. 7 g, `SPEC-296` Nr. 7 e). Es fehlen zwei Dinge, die kein Vertrag dieser Version trägt: ein Span des Fallbacks selbst und der Zähler der verworfenen Versuche (`SPEC-296` Nr. 5 und *Ungeprüft*). Der Slice liefert beide nach dem Vertrag, den der Architect vor dem Start schreibt (Name und Pflicht-Attribute des Spans mit Zeile in `spezifikation.md` §5 und Eintrag in `support.Spans()`, Art und Schlüssel des Zählers, `Telemetrie` im `Konfig` des Hybrid-Dienstes, Aufrufstellen), und verdrahtet den Adapter des Prozesses in `verdrahteHybrid` (`SPEC-139` Nr. 12 a und 13). Kein Query-Literal, kein Body und keine Adresse des Upstreams wird Attribut.

**Übernimmt (Sendung von `welle-hybrid` §6, Spans und Zähler des Hybrid-Fallbacks, ohne Nehmer bis zur Entscheidung des Users vom 2026-11-23; Geber: `welle-hybrid`, in den Plänen [slice-telemetry-port](../done/slice-telemetry-port.md), [slice-telemetry-spans](../done/slice-telemetry-spans.md), [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md), [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md), [slice-telemetry-einbau-mock-mutation](../done/slice-telemetry-einbau-mock-mutation.md) und [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md) mit Ausgang *weiter offen* geführt; §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen, kein Ausschluss trifft die Sendung):** der Span des Fallbacks, der Zähler der verworfenen Versuche und die Verdrahtung. Die Pläne der Geber liegen in `done/` und nennen den Nehmer nicht; der Nehmer ist diese Datei (`AGENTS.md` §3.13).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Gesamtspan, der Zähler je Fehlercode und die Übergabe des Adapters des Prozesses an den HTTP-Adapter — [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md) (liegt bei diesem Start in `done/`). Nach Entscheidung des Users vom 2026-11-23 (Variante A) zählt allein der HTTP-Adapter, was ein Client als Fehlerobjekt erhielt (`SPEC-294` Nr. 5); dieser Slice ändert das nicht: der Zähler der verworfenen Versuche ist eine andere Zahl als `odr_errors_total`, oder der Architect legt die Frage dem User vor (§6).
- Der Query-Span des Versuchs im Mock und der Match-Span des Versuchs im Replay — [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md) und [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md); beide liegen in `done/`, dieser Slice ändert die Pakete `mock` und `replay` nicht.
- Der Capture-Span und die Einträge der Dienste `record`, `session` und `maintenance` — [slice-telemetry-einbau-record](slice-telemetry-einbau-record.md); der Dienst `record` gehört nicht zum Modus `hybrid` (`SPEC-295` Nr. 7 f).
- Port und Adapter `Telemetry` als Bauteil: Zeile, Stufen, Senke, `Handler`, Verteilung und Ring — [slice-telemetry-port](../done/slice-telemetry-port.md), [slice-telemetry-log](../done/slice-telemetry-log.md), [slice-telemetry-handler](../done/slice-telemetry-handler.md), [slice-telemetry-spans](../done/slice-telemetry-spans.md). Dieser Slice fügt dem Adapter allein hinzu, was der Vertrag nennt (eine Zeile der Menge der Spans, gegebenenfalls eine Reihe der Metriken), und ändert keinen bestehenden Vertrag des Adapters ohne den Architect.
- Das Handbuch-Kapitel über die Logzeilen des Prozesses — [slice-handbuch-telemetrie](slice-handbuch-telemetrie.md); die Zeile des Fallbacks braucht einen Upstream oder importierte Austausche und ist als gebundenes Beispiel nicht erreichbar (`SPEC-200` Nr. 7 b); der Slice schreibt kein Kapitel und führt keinen Fehlercode, Schlüssel oder Kommando ein (siehe DoD und §6).
- Die Binary-Verdrahtung über `verdrahteHybrid` hinaus (Konfiguration, Start, Kommandos) — [slice-bootstrap-m4](../done/welle-hybrid/slice-bootstrap-m4.md) hat sie geliefert; der Slice ändert nur, welchen Adapter der Hybrid-Dienst erhält, und fügt keinen Schlüssel hinzu.
- Latenz des Fallbacks im Lasttest — [slice-lasttest](slice-lasttest.md) nimmt den Hybrid-Fallback aus (`LH-QA-01.a`); der Span dieses Slice ist keine Bedingung des Lasttests.
- Die Art der Policy und ihre Entscheidungen (`SPEC-261`) und die Antwort des Fallbacks (`SPEC-284` Nr. 7 bis 9) — Bestand von [slice-fallback-policy](../done/welle-hybrid/slice-fallback-policy.md) und [slice-hybrid-service](../done/welle-hybrid/slice-hybrid-service.md); der Span beobachtet sie und ändert kein Ergebnis.

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

- [ ] Span des Fallbacks im Hybrid-Dienst nach dem Vertrag des Architect: Beginn, Ende und Attribute, `Konfig.Telemetrie` mit nil-Prüfung, das `ctx` des Aufrufs unverändert an `StartSpan` und an jeden Port hinter dem Beginn; kein Query-Literal, kein Body und keine Adresse des Upstreams als Attribut (Zeilen der Tabelle in §3, vom Architect).
- [ ] Zähler der verworfenen Versuche nach dem Vertrag des Architect (Replay-Fallback und Upstream-Fallback; ein Versuch, dessen Ergebnis der Hybrid-Dienst verwirft); `odr_errors_total` und `SPEC-294` Nr. 5 bleiben, wie sie sind, sofern der User nichts anderes entscheidet.
- [ ] Verdrahtung im Composition Root: der Adapter des Prozesses ist die `Telemetrie` des Hybrid-Dienstes (`SPEC-139` Nr. 13, ein Adapter je Prozess); ein Test durch den Root mit dem echten Adapter findet die Zeile des Spans als Kind des Gesamtspans (Zeilen der Tabelle in §3, vom Architect).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: der Slice führt keinen Fehlercode, keinen Schlüssel und kein Kommando ein; die Logzeile des Fallbacks beschreibt kein Kapitel (§6); das Kataloge-Gate ist grün, und §7 sagt, dass kein Fehlercode, Schlüssel oder Kommando des Binary dazukam.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

**Vertrag fehlt, Auftrag Architect** (`AGENTS.md` §3.12, §3.15, §3.16): Die Tabelle *Zusage · Fänger*, das Raster *Operation × Argumentart* (§6), die Mengen mit Zählprüfung und die Nummern aus `SPEC-320` bis `SPEC-339` entstehen mit dem Vertrag vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit; bis dahin steht hier keine Tabelle. Aufträge:

1. **Span des Fallbacks:** Name (Konstante, Zeile in `spezifikation.md` §5, Eintrag in `support.Spans()`; die Menge der Spans ist nach `SPEC-230` Nr. 5 geschlossen: fünf Namen, elf Attribute), Pflicht-Attribute und ihre Werte (Fallback-Art Replay oder Upstream, Ergebnis), Beginn und Ende (nach der Entscheidung der Policy, vor der Antwort, auf jedem Rückweg genau einmal), Eltern-Kind zum Gesamtspan und zum Span des Versuchs (`SPEC-294` Nr. 3 e).
2. **Zähler der verworfenen Versuche:** Name, Art, Einheit und Schlüssel (`SPEC-230` Nr. 8) und das Verhältnis zu `SPEC-294` Nr. 5. Die Entscheidung des Users vom 2026-11-23 (Variante A) gilt: kein Dienst zählt zusätzlich in `odr_errors_total`, und `ODR-E2401` und die Codes von `ablage` stehen nicht darin. Ein Zähler, der einen verworfenen Code in `odr_errors_total` zählte, wäre eine Änderung dieser Zusage; der Architect schneidet ihn als eigene Reihe oder gibt die Frage an den Planner zurück.
3. **`hybrid.Konfig.Telemetrie`** (nil-Prüfung, `SPEC-284` Nr. 1), Aufrufstellen mit Zahl je Aufruf (Muster `SPEC-295` und `SPEC-296` Nr. 6), `ctx`-Regel und Abbruch (Kontextfehler vor und nach dem Fallback, Abbruch des Clients bei wartendem Upstream, Herunterfahren), Verhalten je Fallback-Art, ohne Fallback (kein Span) und bei Pufferung über `MaxBytes` (`SPEC-284` Nr. 5 e).
4. **Sensible Inhalte** (`SPEC-232` Nr. 2): Quellen der Attribute und Fänger je Stelle, insbesondere keine Adresse, kein Host und kein Header des Upstreams.
5. **Nachzug im Vertrag:** `SPEC-296` Nr. 5, 7 e und *Ungeprüft*, `SPEC-290` Nr. 7 g und *Ungeprüft*, `SPEC-294` Nr. 5 (der Satz „was er verwirft, zählt nirgends“) und `SPEC-230` Nr. 5 (Zahl der Spans und Attribute) tragen nach der Lieferung nicht mehr, dass kein Vertrag den Span des Fallbacks trägt; die Änderung an `spec/` schreibt der Architect.
6. **Eingaben der Marke** (`SPEC-128` Nr. 3 k): der Teil des Architect im Block (`pfad`, `produktcode`, `fakes`) und die Klassen der Tabelle (`SPEC-128` Nr. 3 e).

Pfad-Kandidaten (der Architect bestätigt oder ändert sie):

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/services/hybrid/` | update | `Konfig.Telemetrie`, Span des Fallbacks, Zähler der verworfenen Versuche |
| `src/hexagon/ports/support/` und `src/adapters/driven/telemetry/` | update | die Zeile in `support.Spans()` und gegebenenfalls die Reihe der Metriken; die Tests, die fünf Spans und elf Attribute festhalten, ziehen um (§6) |
| `src/bootstrap/` | update | `verdrahteHybrid` gibt dem Dienst den Adapter des Prozesses; Tests der Verdrahtung durch den Root |
| `spec/spezifikation.md` | Architect, auf `main` | Vertrag aus `SPEC-320` bis `SPEC-339`, Nachzug der Nummern aus Auftrag 5; dieser Slice schreibt sie nicht |
| Tests | neu, update | Journal-Fake der Telemetrie im Paket `hybrid`, Tests durch den Composition Root im Modus `hybrid` mit dem echten Adapter |

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

**Start** (`next` → `in-progress`): [slice-hybrid-service](../done/welle-hybrid/slice-hybrid-service.md) und [slice-hybrid-service-upstream](../done/welle-hybrid/slice-hybrid-service-upstream.md) (der Hybrid-Dienst mit beiden Fallbacks), [slice-bootstrap-m4](../done/welle-hybrid/slice-bootstrap-m4.md) (`verdrahteHybrid`), [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md), [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md), [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md) und [slice-telemetry-einbau-record](slice-telemetry-einbau-record.md) liegen in `done/`; der Vertrag samt Tabelle, Raster und Mengen steht in `spec/spezifikation.md` und in §3 und §6 dieses Plans auf `main`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Vertrag ist Auftrag an den Architect (§3).

**Abhängigkeiten — Grund je Eintrag:**

- [slice-hybrid-service](../done/welle-hybrid/slice-hybrid-service.md) und [slice-hybrid-service-upstream](../done/welle-hybrid/slice-hybrid-service-upstream.md): *Code-Abhängigkeit*. Der Hybrid-Dienst mit Replay-Fallback und Upstream-Fallback ist der Ort des Spans und des Zählers (`src/hexagon/services/hybrid/`).
- [slice-bootstrap-m4](../done/welle-hybrid/slice-bootstrap-m4.md): *Code-Abhängigkeit*. `verdrahteHybrid` in `src/bootstrap/wurzel.go` bindet den Hybrid-Dienst.
- [slice-telemetry-einbau-gateway](../done/slice-telemetry-einbau-gateway.md) (Gesamtspan, Adapter des Prozesses) und [slice-telemetry-einbau-replay](../done/slice-telemetry-einbau-replay.md) (ein Adapter je Prozess in `hybrid`, Match-Span des Versuchs) und [slice-telemetry-einbau-mock](../done/slice-telemetry-einbau-mock.md) (Query-Span des Versuchs): *Code-Abhängigkeit*. Der Span des Fallbacks ist Kind des Gesamtspans und Nachbar der Spans der Versuche.
- [slice-telemetry-einbau-record](slice-telemetry-einbau-record.md): *Reihenfolge*, keine Code-Abhängigkeit. Beide ändern `src/bootstrap/wurzel.go`; sie laufen nacheinander (Planner, 2026-11-23, `welle-betrieb` §5: wo zwei Slices dieselben Dateien in `src/bootstrap/` ändern, laufen sie nacheinander); der Hybrid-Slice startet nach dem Teil 3.
- [slice-telemetry-spans](../done/slice-telemetry-spans.md) und [slice-telemetry-port](../done/slice-telemetry-port.md): Port, Fake und Adapter mit der Menge der Spans, die dieser Slice um eine Zeile erweitert.

**Größenmarke:** steht aus bis zum Vertrag. Der Plan hat keine Tabelle und keinen Block (`SPEC-128` Nr. 3 p: Zustand *kein Block*); der Architect liefert mit dem Vertrag die Tabelle und den Teil des Architect im Block (`pfad`, `produktcode`, `fakes`), der Planner setzt `basis`, `schaetzung` und `marke` beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke`. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt (`AGENTS.md` §3.14); ergibt die Tabelle des Architect *S* ≥ 2223, schneidet der Planner vor dem Code.
- `in-progress` → `open` (blockiert): der Vertrag lässt sich nicht ohne Änderung der Zusage von `SPEC-294` Nr. 5 schneiden (siehe §6) und der User hat nicht entschieden, oder der Span des Fallbacks lässt sich nicht beenden, ohne dass er die Antwort des Fallbacks ändert (`SPEC-284` Nr. 7 bis 9).

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

- Der Vertrag fehlt: weder der Span des Fallbacks noch der Zähler der verworfenen Versuche noch `hybrid.Konfig.Telemetrie` steht in der Spezifikation; ohne Vertrag gibt es keine Tabelle, kein Raster und keine Marke (`AGENTS.md` §3.12) — **Ausgang:** weiter offen; Adresse: Architect (Aufträge 1 bis 6 in §3; Nummern aus `SPEC-320` bis `SPEC-339`, Übergangsregeln Nr. 6), vor dem Start dieses Slice.
- Der Zähler der verworfenen Versuche und die Zusage von `SPEC-294` Nr. 5 (Variante A, Entscheidung des Users vom 2026-11-23: der Adapter zählt, was ein Client als Fehlerobjekt erhielt, kein Dienst zählt zusätzlich) stehen gegeneinander, wenn der Zähler `odr_errors_total` meint — **Ausgang:** weiter offen; Adresse: Architect (Auftrag 2; ein Zähler als eigene Reihe widerspricht der Zusage nicht, ein Eintrag in `odr_errors_total` ginge an den Planner und den User zurück).
- Die Menge der Spans ist geschlossen (fünf Namen, elf Attribute, `SPEC-230` Nr. 5 und `support.Spans()`); ein sechster Name ändert die Tests, die die Zahlen festhalten (die Zahl 5 und 11 im Paket `support`, die Menge der Zeilen im Adapter), und die Aussage in §5 von `spezifikation.md` — **Ausgang:** weiter offen; Adresse: Architect (der Vertrag nennt die Stellen, der Implementer sucht sie mit `git grep` und trägt jede weitere in §3 nach, `AGENTS.md` §3.9).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** weiter offen; Adresse: Architect (Raster *Operation × Argumentart* mit den Randformen des Fallbacks: Abbruch des Clients bei wartendem Upstream, Herunterfahren, Kontextfehler zwischen Versuch und Fallback, kein Fallback, Pufferung über `MaxBytes`, je Fallback-Art; jede Zelle trägt einen Ort oder `gilt nicht` mit Grund).
- Der Einbau trifft einen nebenläufigen Pfad (Upstream-Fallback mit wartender Anfrage, `Schliessen` wartet höchstens `rest`) und Fakes, die `ctx` umhüllen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×), `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (8×) — **Ausgang:** weiter offen; Adresse: Architect (Wirkungs-Fänger oder akzeptiertes Negativ je Zusage über Nebenläufigkeit), Raten je Prozess misst der Implementer (`AGENTS.md` §3.10).
- Handbuch: Mit dem Slice schreibt das Binary eine weitere Span-Zeile im Modus `hybrid`; sie braucht einen Upstream oder importierte Austausche und ist als gebundenes Beispiel nicht erreichbar (`SPEC-200` Nr. 6 d, 7 b); das Kapitel sagt sie nicht (`SPEC-200` Nr. 1) — **Ausgang:** entschieden (Planner, 2026-11-23): wie die Zeilen von `odata.record.capture` und `odata.replay.match` in [slice-handbuch-telemetrie](slice-handbuch-telemetrie.md) (§1 dort): kein Kapitel, kein Fehlercode, Schlüssel oder Kommando in diesem Slice; ob das Kapitel später gebundene Beispiele für Zeilen mit Upstream bekommt, ist die Frage in `welle-betrieb` §5, Punkt c (Adresse User).
- Schnitt und Größenmarke: Plan ohne Tabelle, keine Schätzung — **Ausgang:** weiter offen; Adresse: Planner (setzt beim Start den Block mit `basis`, `schaetzung` und `marke` nach der Tabelle des Architect, `AGENTS.md` §3.14; ergibt *S* ≥ 2223, schneidet der Planner vor dem Code).
- Gemeinsame Datei `src/bootstrap/wurzel.go` mit [slice-telemetry-einbau-record](slice-telemetry-einbau-record.md) und fremde Zeilen an den Pfaden des Blocks (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`) — **Ausgang:** entschieden für die Reihenfolge (Planner, 2026-11-23): die beiden Slices laufen nacheinander, dieser nach Teil 3 (§4); die Dateien eines zugleich laufenden Slice vergleicht der Planner beim Start und nennt den Block mit Dateien einzeln, wo `src/bootstrap/` fremde Zeilen trüge.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Anlegens gesichtet (2026-11-23; der Zähler ist die Zahl der Dateien unter `evidence/`, der Planner zieht Treffer und Zähler beim Start nach). Der Einbau trifft einen nebenläufigen Pfad und Fakes, die `ctx` umhüllen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (18×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (8×); `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (10×) gilt für jede Wirkung, die der Detektor nicht meldet; `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×) für `src/bootstrap/` (§4). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (77×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16), `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (72×, `AGENTS.md` §3.17), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (43×, `AGENTS.md` §3.9) für den Nachzug der Nummern in §3 Auftrag 5 und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (22×). Keine Beobachtung steht mit diesem Slice offen unter 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
