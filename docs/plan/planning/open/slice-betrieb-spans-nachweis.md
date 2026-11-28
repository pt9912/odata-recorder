# Slice slice-betrieb-spans-nachweis: Ein Test belegt Record, Replay und Mock mit den fünf Spans und den Zählern

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md) — Kriterium der Closure (§3 dort), das kein Slice der Welle als einen Test geliefert hat.

**Bezug:** `LH-QA-05`, `LH-QA-04`, `LH-FA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md).

**Berührte Spec-Stellen:** `SPEC-026` bis `SPEC-030` (Spans und Pflicht-Attribute) · `SPEC-230` (Zähler je Fehlercode) · `SPEC-294` Nr. 5 (welche Fehler gezählt werden). Der Slice ändert keine Stelle.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner, Vorbereitung der Closure von `welle-betrieb`). **Datum:** 2026-11-28.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar.

**Ziel:** Ein Test fährt über den Composition Root die Modi `record`, `replay` und `mock`, findet die Spans `SPEC-026` bis `SPEC-030` mit ihren Pflicht-Attributen ohne Query-Literale und Bodies und sieht je ausgelöstem Fehlercode den Zähler des Adapters der Telemetrie steigen.

**Anlass** (Planner, 2026-11-28, gelesen, nicht gefahren): Die Teile sind einzeln belegt, in keinem Test zusammen. Spans je Modus: `TestWurzelGesamtspanJedeAnfrage`, `TestWurzelCaptureLiegtImGesamtspan`, `TestWurzelReplayMatchLiegtImGesamtspan`, `TestWurzelMockSpanLiegtImGesamtspan`, `TestWurzelMutationSpanZeile` (alle `src/bootstrap/`). Zähler: `TestFehlerzaehlerJederFall` (`src/adapters/driving/httpodata/`, Fake der Telemetrie); mit dem echten Adapter über den Root lesen `TestWurzelHybridUpstreamNichtErreichbar` (`ODR-E4001`, Modus `hybrid`) und `TestAdminAbweisenBegrenzt` (`ODR-E2101`, Admin-Listener) den Zähler, kein Test liest ihn für einen Code, den ein Client am OData-Listener in `record`, `replay` oder `mock` erhält.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Eine Änderung am Produktcode oder an der Spezifikation — der Slice belegt vorhandenes Verhalten; zeigt der Test eine Abweichung, hält der Implementer an (§4).
- Der Modus `hybrid` — das Kriterium der Welle nennt ihn nicht; seine Spans und Zähler tragen die Tests von [`slice-telemetry-einbau-hybrid`](../done/slice-telemetry-einbau-hybrid.md).
- Ein Lauf am gebauten Binary — das Kriterium verlangt „ein Test“; der Stand der Zähler ist im Prozess am Adapter lesbar.
- Ersatz oder Kürzung der Tests oben — sie bleiben als Fänger ihrer Verträge.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Ein Test über `record`, `replay` und `mock` findet jeden der fünf Spans mit jedem Pflicht-Attribut; ein Canary als Query-Literal und in Request- und Response-Body steht in keiner Zeile der Senke.
- [ ] Derselbe Test vergleicht je Modus den Stand der Zähler mit genau den Codes, die sein Client als Fehlerobjekt erhielt; mindestens ein Code je Modus.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben; keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/bootstrap/telemetrie_betrieb_test.go` | neu | eine Testfunktion; Helfer der vorhandenen Dateien `telemetrie_*_test.go` (`baueWurzelHaken`, `mockDaten`, `replayDaten`, `fakeUp`) |

Der Bericht nennt je Zusage der DoD eine Mutation (ein Span fehlt, ein Pflicht-Attribut fehlt, der Canary als Attribut, ein Zähler steigt nicht) und den roten Lauf.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein Code-Vorgänger; die Slices der Welle liegen in `done/`. Kein Kettenslice.

**Größenmarke:** Den Block nach `SPEC-128` Nr. 1 setzt der Planner beim Start; erwartet ist eine neue Testdatei und kein Produktcode.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß): die Marke liegt nach `SPEC-128` über der Schnittgrenze (nicht zu erwarten).
- `in-progress` → `open` (blockiert): der Test findet ein Verhalten, das den Spec-Stellen widerspricht; Rückgabe an den Planner.

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

Kein neuer Vertrag; ein Raster der Randformen entfällt (`AGENTS.md` §3.12 gilt für Slices mit neuem Vertrag).

- Das Paket `bootstrap` liegt unter dem Detektor bei 415,9 s von 600 s (`BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze`); der Test kommt hinzu — **Ausgang:** weiter offen; der Implementer nennt die Zeit des Tests mit und ohne Detektor.
- Der Capture-Span endet erst mit dem Schließen des Body; ein Test, der vorher liest, flackert — **Ausgang:** weiter offen; Muster in `TestWurzelGesamtspanJedeAnfrage` (Herunterfahren vor dem Lesen der Senke).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten** (Anlage, 2026-11-28): `BEO-REPO/paketlaufzeit-unter-race-naehert-sich-der-zeitgrenze` (Risiko in §6), `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (Mutationen in `src/bootstrap` laufen mit eigenem Treiber), `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (je Attribut ein eigener Wert). Der Planner sichtet beim Start erneut.
