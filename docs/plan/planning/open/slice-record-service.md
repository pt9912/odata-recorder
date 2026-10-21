# Slice slice-record-service: RecordUseCase und Upstream-Adapter

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-FA-02`, `LH-RB-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-01.a` (Proxy) · `LH-FA-02.a`, `LH-FA-02.b`, `LH-FA-02.c` · `LH-FA-04.c` (Service-Root) · `SPEC-012`, `SPEC-054`, `SPEC-057`, `SPEC-086`, `SPEC-087`, `SPEC-092` · `SPEC-048` · `ARC-002` · `SPEC-118`, `SPEC-119` (Nr. 12 für Sendung (g)), `ARC-003`, `ARC-004`, `ARC-008` · `LH-FA-13.a` (Status der Session) · `architecture.md` §4.1.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `RecordUseCase` bestimmt Service und offene Session aus dem Rohrequest, leitet über den Outbound Port `UpstreamClient` und dessen Adapter `httpupstream` an den Upstream weiter, zeichnet über die Capture-Pipeline auf, hängt den Exchange mit `sequence_no` atomar an und kann eine Session anhalten (`Drain`).

**Übernimmt:** Zeit (`started_at`, `finished_at`) kommt aus `Clock.Now()`, die Exchange-Kennung aus `IDGenerator.ExchangeID()` (`SPEC-114`); die Tests setzen `supporttest.FestUhr` und `FestIDs` (Sendung von `slice-clock-idgenerator`).

**Übernimmt (Redaction und Match-Key):** Der Dienst ruft `BildeSchluessel` (`SPEC-116`) mit dem unredigierten Request und dem bereits redigierten Request-Body, persistiert `RawQuery` und `Header` des Ergebnisses und `KeyRedacted`, redigiert die Response-Header mit `RedigiereHeader` (`SPEC-115` Nr. 2, 3) und übergibt als `Path` den Pfad nach dem Service-Root; bei `ErrKodierung` (`MATCH_INVALID_ENCODING=reject`) antwortet er `ODR-E5006` vor jeder Weiterleitung und hängt keinen Exchange an, bei `redaction.request=unsafe` bleibt der Key `null` (`SPEC-116` Nr. 7, 9); ein Body, der als JSON ungültig ist, ergibt im Modus `json-canonical` den Byte-Hash (Art 2) und keinen Fehler, auch bei `reject` (`SPEC-116` Nr. 6; Sendung von `slice-redaction-match-key`).
**Übernimmt (Redaktor):** Der Request-Body geht durch `RedigiereBody` (über die Capture-Pipeline, `SPEC-118`), und der Key wird aus dessen Ausgabe gebildet, nie aus den Eingabebytes (`SPEC-118` Nr. 10, `SPEC-116` Einleitung); bei einem Fehler des Redaktors ist `redaction.request=unsafe`, der Key `null` und der Exchange `failed` (`SPEC-001`); die Aufzeichnung des Response-Bodys und die Weiterleitung laufen weiter (Sendung von `slice-streaming-redaction`).
**Übernimmt (Capture-Pipeline):** Sendung von `slice-capture-pipeline` (Vertrag `SPEC-119`): (a) Der Dienst gibt jeden Body mit einem Strom durch `Tee` und liest ihn aus der `Aufnahme` (Nr. 2, 3); Antworten, die nach HTTP keinen Body haben (`HEAD`, 1xx, 204, 304), gibt er nicht hinein (Nr. 6); `Metadata` setzt er nur für die Antwort auf `<service_root>/$metadata`. (b) Er holt `Ergebnis()` je Richtung, bildet `Abschluss`, setzt die Felder des Exchange daraus und bildet den Key nur bei `KeyMoeglich` aus `Ergebnis.Redigiert` und `RequestRedact` des Zustands, als `Body` und `RequestRedaction` von `SchluesselEingabe` (Nr. 12, 13). (c) Er beendet jede Aufnahme, deren Austausch endet, ohne dass der Strom zu Ende gelesen wurde (Upstream liest den Request-Body nicht zu Ende, Client-Abbruch, Frist von `Drain`), mit `Abbrechen` (Nr. 14). (d) Er protokolliert `Ergebnis.Fehler` mit `ODR-E2401` ohne Payload; die Pipeline protokolliert nichts (Nr. 11). (e) Er entscheidet, ob er `Accept-Encoding` zum Upstream auf `gzip, deflate` einschränkt (ein Upstream mit `br` oder `zstd` ergibt einen `failed`-Austausch, Nr. 6), und den Zustand eines Austauschs ohne Antwort-Strom (Transportfehler vor dem ersten Byte): die Pipeline kennt keinen Austausch ohne Strom. (f) Der Fault-Injection-Punkt zwischen gelungenem Blob-`Commit` und dem Anhängen des Exchange liegt bei diesem Slice (Nr. 8, 15): ein gelungener `Commit` wird nie zurückgenommen, der verwaiste redigierte Blob ist zulässig (`SPEC-048`). (g) Offene Randform aus V20 der Nachverifikation von `slice-capture-pipeline`, vor dem Code durch den Architect im Vertrag zu entscheiden: Ein Exchange mit beiden Seiten `Vollstaendig` beim `Drain` oder beim Transportfehler nach dem ersten Byte — gilt er als laufend (`incomplete` oder `failed`) oder als `complete` (`SPEC-119` Nr. 12)? `Drain` vor dem ersten Response-Byte: es gibt kein Response-Ergebnis, und `Abschluss` braucht eines. Daran hängt die Session-Regel (`LH-FA-13.a`).
**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- HTTP-Server, Header-Mapping des Eingangs und Listener-Lebenszyklus — `slice-http-adapter`; der Use Case erhält Methode, Pfad, `raw_query`, Header und einen Body-Strom.
- Beenden und Anlegen von Sessions als Use Case — `slice-session-wartung`; `Drain` ist die Operation, die es aufruft.
- Umschreiben von Links und Weiterleitungen — `LINK_REWRITE` darf nur `off` sein, Record speichert und liefert unverändert.
- Upstream-Fallback und Allowlist (`FALLBACK_*`) — Hybrid, nicht in der Welle.

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

- [ ] Outbound Port `UpstreamClient` und Adapter `httpupstream`: Methode, Pfad, Query, Body und End-to-End-Header weiterleiten, Hop-by-Hop-Header behandeln, `X-Recorder-*` entfernen, Timeout, TLS-Optionen des Upstreams, Streaming in beide Richtungen.
- [ ] `RecordUseCase`: Service aus dem Pfad (`ODR-E5005`), offene Session (`ODR-E5009`, `ODR-E5003`), Weiterleiten, Aufzeichnen über die Pipeline, Anhängen des Exchange mit Match-Key aus der redigierten Darstellung, Fehlermodi vor und nach dem ersten gesendeten Byte (`ODR-E4001`, `ODR-E2401`), `Drain` mit Frist.
- [ ] Fault-Injection-Punkte im Commit und nach dem Commit vor der Antwort (`SPEC-048`): kein unredigierter persistenter Body, kein `complete` für einen Teil-Exchange.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/inbound/` | neu | `RecordUseCase` (Rohrequest als Domain-DTO mit Body-Strom) |
| `src/hexagon/ports/outbound/` | update | `UpstreamClient` |
| `src/hexagon/services/record/` | neu | Service, ruft `capture` |
| `src/adapters/driven/httpupstream/` | neu | HTTP/TLS-Client zum Upstream |
| `spec/spezifikation.md` | update | Vertrag von `UpstreamClient` und `RecordUseCase` mit Randformen |
| Tests der Pakete | neu | Service gegen Fakes der Ports, Adapter gegen `httptest`; Vertragstest `UpstreamClient` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

- Der Use Case ist nebenläufig: mehrere Exchanges je Session laufen gleichzeitig, `sequence_no` ist monoton; die Vergabe geschieht, wo die Reihenfolge der Abschlüsse feststeht (Entscheidung des Architect, siehe §6).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-capture-pipeline` und `slice-fehlermodell-konfiguration` und `slice-clock-idgenerator` und `slice-redaction-match-key` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 1700 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 1200 Zeilen in derselben Metrik.
- `in-progress` → `open` (blockiert): `sequence_no` lässt sich bei gleichzeitigen Exchanges nicht ohne Wartezeit monoton vergeben (siehe §6).

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

- `sequence_no` bei gleichzeitigen Exchanges: Beginn oder Abschluss entscheidet die Reihenfolge — und damit, was `sequential`-Replay als Erstes liefert. `LH-FA-02.a` sagt „monotone `sequence_no`“ und „Beginn und Ende jedes Austauschs erfassen“ — **Ausgang:** offen; Architect entscheidet vor dem Code.
- Zwei offene Sessions desselben Service und der Header `X-Recorder-Recording-Session` (`ODR-E5003`) — **Ausgang:** offen; Raster-Zelle der Session-Wahl.
- `HEAD`, `OPTIONS`, `CONNECT`, Anfragen mit `Expect: 100-continue`, `Upgrade`, `Transfer-Encoding: chunked` im Request — **Ausgang:** offen; Raster *Methode × Body*, was weitergeleitet, aufgezeichnet oder abgewiesen wird.
- `$batch` (`multipart/mixed`) ist ein unbekannter Typ und fail-closed für die Aufzeichnung; ob der Request dennoch weitergeleitet wird, ist nur in `SPEC-046` Nr. 1 angedeutet — **Ausgang:** offen; Architect.
- Nebenläufigkeit: `Drain` gegen laufende Exchanges, Client-Abbruch, Upstream-Timeout — **Ausgang:** offen; das Gate fährt kein `-race` (`CGO_ENABLED=0`): die nebenläufigen Fänger ordnen mit Kanälen, und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50`.
- Status eines Exchange mit beiden Seiten `Vollstaendig` beim `Drain` oder beim Transportfehler nach dem ersten Byte, und `Drain` vor dem ersten Response-Byte (§1, Sendung (g) von `slice-capture-pipeline`) — **Ausgang:** offen; der Architect entscheidet vor dem Code in der Spezifikation, und die Session-Regel in `LH-FA-13.a` folgt der Entscheidung.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (über der Schwelle, ohne Ausgang); der Slice ist nebenläufig und trägt die meisten Randformen der Welle. Der Slice ist nebenläufig: das Gate fährt kein `-race` (`CGO_ENABLED=0`), die nebenläufigen Fänger ordnen mit Kanälen, jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3), und ein drittes Auftreten dieser Beobachtung in dieser Welle ist kein Befund mehr, sondern ein eigener Slice.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
