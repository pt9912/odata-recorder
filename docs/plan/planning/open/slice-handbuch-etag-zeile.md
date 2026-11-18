# Slice slice-handbuch-etag-zeile: Zeile `ETAG_ENABLED` des Handbuchs auf die Einstellung kürzen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md).

**Bezug:** `LH-FA-15` (ETag), `LH-QA-06` (Handbuch), [ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)

**Berührte Spec-Stellen:** `SPEC-200` Nr. 1 (Verhalten des Binary im Handbuch nur als gebundenes Beispiel) und Nr. 14 b (Prosa über Verhalten bleibt ungeprüft) · `SPEC-201` (Schlüsselkatalog nennt `ETAG_ENABLED`) · `SPEC-280`, `SPEC-283` (das Verhalten, das die Zeile heute beschreibt).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner). **Datum:** 2026-11-18.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Zeile `ETAG_ENABLED` der Tabelle *Einstellungen* in `docs/user/benutzerhandbuch.md` nennt den Schlüssel und seine Wirkung als Einstellung (Wert, Modus, Vorgabe, was `true` einschaltet) und beschreibt kein HTTP-Verhalten ohne gebundenes Beispiel mehr (Form des ETag, Kopfzeile, `@odata.etag`, Vergleich, `412`); das Verhalten tragen `SPEC-280` bis `SPEC-283`.

**Übernimmt (Sendung von `slice-prefer-return`, Planner 2026-11-18, `AGENTS.md` §3.13; Entscheidung des Architect zum Handbuch von `Prefer`, Weg B):** das Kürzen der Zeile `ETAG_ENABLED`. Grund beim Geber: die Zeile weicht von `SPEC-200` Nr. 1 ab; `slice-etag-versionen` und `slice-if-match` liegen in `done/`, und `slice-handbuch-mock` schließt das Kürzen in §1 aus (die Zeile trug der Geber).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Das gebundene Beispiel zu `ETAG_ENABLED` und das Kapitel zum Modus `mock` — `slice-handbuch-mock` (Sendung von `slice-etag-versionen`); dieser Slice schreibt kein Kapitel und kein Beispiel.
- Eine Beispiel-Form für HTTP in `SPEC-200` — Entscheidung des Architect (Weg B, `slice-prefer-return` §6): kein Eingriff in den Vertrag.
- Andere Zeilen der Tabelle *Einstellungen* und die Zeile `ODR-E6010` der Tabelle *Fehlercodes* — der Katalog-Abgleich (`SPEC-201`) trägt sie; eine Prüfung weiterer Prosa über Verhalten wäre ein eigener Auftrag.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Die Zeile `ETAG_ENABLED` trägt Schlüssel, Wert, Modus, Vorgabe und die Wirkung als Einstellung; keine Aussage über Header, Form des ETag oder `412` ohne gebundenes Beispiel.
- [ ] `make katalog-check` und `make handbuch-check` bleiben grün; `make gates` grün.
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
| `docs/user/benutzerhandbuch.md` | update | Zeile `ETAG_ENABLED` der Tabelle *Einstellungen* kürzen (`SPEC-200` Nr. 1) |

Kein Code, kein Vertrag, kein neuer Fänger: Die Zeile ist Prosa über Verhalten, die `SPEC-200` Nr. 14 b ungeprüft lässt; der Schlüsselkatalog (`make katalog-check`) hält den Schlüssel, nicht den Wortlaut.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein Trigger außer der Priorisierung; die Änderung hängt von keinem Slice ab. Liegt `slice-handbuch-mock` beim Start in `in-progress/`, schreiben beide dasselbe Dokument; der Planner legt die Reihenfolge fest (die gekürzte Zeile zuerst, das Kapitel mit dem Beispiel danach).

**Größenmarke:** Prosa, kein Block; die Änderung ist eine Tabellenzeile. Der Planner setzt beim Start den Block nach `SPEC-128` Nr. 1, falls das Gate ihn verlangt.

**Rückführungen — vorab benannt:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): die Marke liegt nach `SPEC-128` über der Schnittgrenze (nicht zu erwarten).
- `in-progress` → `open` (blockiert): der Architect entscheidet, dass die Zeile die Aussagen behalten soll (Änderung von `SPEC-200` Nr. 1 statt der Zeile).

## 5. Closure-Trigger

DoD vollständig, Review und Verifikation bestanden, `make gates` grün im Hauptbaum, Closure-Notiz in §7 mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst.

- Die gekürzte Zeile verliert Aussagen (`If-Match`, `412`), die kein gebundenes Beispiel trägt; ein Anwender findet sie nur in der Spezifikation — **Ausgang:** weiter offen; Folge: das Beispiel von `slice-handbuch-mock` deckt Lesen, Ändern und neuen ETag, nicht `412`; ein HTTP-Beispiel bräuchte eine Erweiterung von `SPEC-200` (Entscheidung des Architect, Weg B, nicht getroffen).
- Der Wortlaut der Zeile (was „Wirkung als Einstellung“ ist) ist nicht entschieden — **Ausgang:** weiter offen; Adresse: Architect (nimmt den Wortlaut vor dem Start an).

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln.

Beim Anlegen entstanden aus der Sendung von `slice-prefer-return` (Review Runde 2 N-5, Verifikation V-2); noch nicht begonnen.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, `*` (Kürzel `REPO`); alle berührten Sub-Areas GF.

**Vorgelagert — offene Beobachtungen sichten:** nicht gesichtet beim Anlegen; der Planner sichtet beim Start.
