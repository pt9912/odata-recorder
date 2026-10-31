# Slice slice-http-adapter: HTTP-Driving-Adapter für Record und Replay

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-FA-04`, `LH-FA-02`, `LH-QA-05`, `LH-QA-07`, `LH-RB-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-01.a` · `LH-FA-02.b` · `LH-FA-04.b` · `LH-QA-05.a` · `LH-QA-07.a` · `SPEC-006` · `SPEC-109` Nr. 1, 2 · `SPEC-114` Nr. 6 · `SPEC-121` · `SPEC-135` · `SPEC-136` Nr. 1 bis 10, 12 · `ARC-005` · `architecture.md` §2.2, §2.5.
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

**Ziel:** Der Driving Adapter `src/adapters/driving/httpodata` nimmt HTTP-Requests des OData-Listeners an, bildet Methode, Pfad, `raw_query`, Header und Body-Strom unverändert auf `inbound.Anfrage` ab (`SPEC-136` Nr. 4), ruft den Dienst des Betriebsmodus als Funktionswert (`Verarbeiten`, `SPEC-136` Nr. 2), streamt die Antwort (Nr. 5, 6), schließt den Strom genau einmal, vergibt die Korrelations-ID und schreibt Fehler als Fehlerobjekt (Nr. 7, 8); ein Lesefehler oder ein Kontextfehler bricht die Verbindung ab (Nr. 9). `SPEC-136` und `SPEC-137` schreibt der Zweig beim Start des Slice nach dem Vorabentwurf des Architect (Linie A, `SPEC-129` bis `SPEC-159`); der Lebenszyklus (Anhalten, Schließen, 503, Grenzen des Servers) ist vor dem Code in [slice-http-lebenszyklus](slice-http-lebenszyklus.md) geschnitten.

**Übernimmt:** Die Korrelations-ID kommt aus `IDGenerator.KorrelationsID()` (`SPEC-114` Nr. 6), nicht aus eigenem Code des Adapters; die Tests setzen `supporttest.FestIDs` (Sendung von `slice-clock-idgenerator`).

**Übernimmt (Record):** Der Adapter füllt `inbound.Anfrage` unverändert (`SPEC-121` Kopf): `Pfad` mit den Escapes des Clients ohne Query, `RawQuery` ohne `?`, alle Header außer `Host`, `Body` (nil ohne Body), `Laenge` (`-1`, `0` oder die angekündigte Länge) und `KorrelationsID`. Er schreibt `Antwort.Status` und `Antwort.Header`, kopiert `Antwort.Body` und ruft `Antwort.Body.Close` genau einmal, auch nach einem Fehler beim Schreiben an den Client und nach dessen Abbruch (`SPEC-121` Nr. 7). Vom Aufruf von `Record` bis zur Rückkehr von `Close` (ohne `Antwort`: bis zur Rückkehr von `Record`) ändert er weder `Anfrage.Header` noch eine seiner Wertlisten und liest `Body` nicht (`SPEC-121` Kopf). Einen `*inbound.Fehler` schreibt er mit dessen `Code` und `Status` als Fehlerobjekt (`SPEC-109` Nr. 1, `SPEC-121` Nr. 14); ein Kontextfehler ohne Code bekommt keine Antwort, `model.ErrUngueltig` ist `ODR-E9001` (Sendung von `slice-record-service`).

**Übernimmt (Replay):** (Sendung von `slice-replay-service`, `SPEC-135`) Die `Anfrage` wird unverändert gefüllt, auch `X-Recorder-*`; der Dienst entnimmt die Zuordnung (`SPEC-052`). Länge und Coding der Antwort setzt der Adapter selbst: der Dienst liefert weder `Content-Length` noch `Content-Encoding`, bei `HEAD` keine Länge. Ein Fehler ohne `*Fehler` und ohne Code (Kontextfehler, Abbruch des Clients, `model.ErrUngueltig`) wird wie in `SPEC-121` Nr. 14 behandelt. `X-Recorder-Match-Redacted` kennt der Adapter nicht; der Dienst liefert es in den Headern, und der Adapter schreibt sie unverändert.

**Übernimmt (Race-Detektor):** Die nebenläufigen Tests des Pakets laufen im Ziel mit `go test -race`, das `slice-race-gate` als `make test-race` an `make gates` hängt (Sendung von `slice-race-gate`); dessen Vertrag führt keine Liste der Pakete, sondern fährt alle Pakete der Stufe `source` (`SPEC-122` Nr. 2), so dass `httpodata` ohne Eintrag darunter fällt und dieser Slice dafür nichts ändert.

**Herunterfahren geht weiter:** Die Sendung von `slice-session-beenden` (zwei Schritte des Listeners, 503 auf bestehenden Verbindungen, keine doppelte Frist; `SPEC-124` Nr. 14, `LH-FA-02.c`) und das Risiko „Request-Größe und Zeitgrenzen“ liegen bei [slice-http-lebenszyklus](slice-http-lebenszyklus.md), das sie in seinem §1 annimmt; dieser Slice trägt sie nicht mehr.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Entscheidungen über Cursor, Match-Key, Coverage oder CRUD — der Adapter entscheidet sie nicht (`architecture.md` §2.2).
- Die Wahl des Use Case je Modus — Composition Root (`slice-bootstrap-m2`); der Adapter kennt den Modus nicht.
- Span `odata.gateway.request` und Metriken — es gibt noch keinen `Telemetry`-Port (Welle §6); der Adapter vergibt die Korrelations-ID, die der Port später trägt.
- TLS des Listeners — Folgewelle.
- Der Lebenszyklus (Anhalten, Schließen mit Frist, 503 `ODR-E5010` auf bestehenden Verbindungen, feste Grenzen des Servers) — [slice-http-lebenszyklus](slice-http-lebenszyklus.md): der einzige Teil ohne geteilten Mechanismus mit dem Kern, vor dem Code geschnitten (`AGENTS.md` §3.14).
- Spans und Zähler je Fehlercode am Adapter — `slice-telemetry-einbau-gateway` (nimmt an, siehe dessen §1).
- Absolute-Form-Requests (Entscheidung des Architect in `SPEC-136` Nr. 4c) — der Kern liefert dafür keinen Pfad.

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

- [ ] Abbildung und Streaming: Request und Antwort ohne Veränderung der Bytes, geliehene kontextgebundene Ströme, `Close` genau einmal, Abbruch durch Kontext und Lesefehler; Transportfehler vor dem ersten Byte als Gateway-Status, danach keine Änderung des Status (`LH-FA-02.b`); Länge und Coding der Antwort setzt der Adapter (bei `HEAD` keine Länge), `X-Recorder-*` der Anfrage gehen unverändert an den Dienst (Sendung von `slice-replay-service`).
- [ ] Fehlerobjekt mit Code, Status und `X-Correlation-Id` für jeden Code der Tabelle, `ODR-E9001` für Verträge, die der Dienst bricht (Vorschlag des Architect; Wortlaut und Zählung der Codes: `SPEC-136` Nr. 7, 8, beim Start).
- [ ] Tests über `httptest` und rohen Client gegen Fakes der Inbound Ports, ohne SQLite; die nebenläufigen laufen im Ziel mit `-race` aus `slice-race-gate` (Sendung von `slice-race-gate`) und mit `-count=50`.
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
| `src/adapters/driving/httpodata/` | neu | `New`, `Serve`, Handler, Abbildung, Fehlerobjekt, Streaming, Abbruch (der Lebenszyklus: `slice-http-lebenszyklus`) |
| `spec/spezifikation.md` | update | Vertrag `SPEC-136` mit Randformen: Header-Behandlung, Hop-by-Hop, Abbruch, Fehlerobjekt (Vorabentwurf des Architect; Nr. 11 bis 15 und `SPEC-137` gehören zu `slice-http-lebenszyklus`) |
| Tests des Pakets | neu | `httptest` mit Fake-Ports; Fälle je Fehlercode aus `LH-QA-05.a` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

- Der Adapter importiert nur Inbound Ports, Querschnitts-Ports und Domain (`a-check`); die Ports `RecordUseCase` und `ReplayUseCase` liefern die beiden Slices davor.
- Der Auftrag an den Architect zur Reihenfolge des Herunterfahrens (Sendung von `slice-session-beenden`) ist beim Vorabentwurf erledigt und liegt bei [slice-http-lebenszyklus](slice-http-lebenszyklus.md) (`SPEC-136` Nr. 13 bis 15).
- Modus-Routing: der Adapter kennt den Modus nicht (`architecture.md` §2.5); `Konfig.Verarbeiten` ist ein Funktionswert, den der Composition Root aus `rec.Record` (Modus `record`) oder `rep.Replay` (Modus `replay`) bindet, beide mit der Signatur `func(ctx, inbound.Anfrage) (inbound.Antwort, error)`; Mock und Hybrid binden dieselbe Signatur später (Vorabentwurf des Architect, `SPEC-136` Nr. 2).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-record-service` und `slice-replay-service` und `slice-race-gate` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 4800 Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/adapters/driving/httpodata/`, Tests eingeschlossen; Basis ist der letzte Commit des Architect). Schätzung S = 3525 (a=5, b=29, f=2, p=380 nach der Tabelle des Architect), Marke 100 · aufrunden(135·3525/10000) = 4800; die Schnittfrage ist gestellt und der Lebenszyklus vor dem Code geschnitten (`slice-http-lebenszyklus`); Grund für den Kern: er ist ein Ablauf, den Handler, Fake des Dienstes und Test-Server teilen, und ein weiterer Schnitt ließe den Adapter ohne Fehlerpfad oder ohne Abbruchsemantik. Die ältere Erwartung „rund 700 Zeilen“ und die Rückführung „über 1000“ trugen keine Tabelle und entfallen. Den Block nach `SPEC-128` Nr. 3 setzt der Planner beim Start.
- `in-progress` → `open` (blockiert): entfällt — `RequestURI` und `r.Body` genügen, kein Typ von `net/http` erreicht den Port (Vorabentwurf des Architect).

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

- Header-Behandlung: mehrfach vorkommende Header, Groß-/Kleinschreibung, `Host`, `Connection`-Liste, `Transfer-Encoding` — **Ausgang:** offen; Raster *Header-Art × Behandlung*, Entscheidung in der Spezifikation.
- Client bricht ab, bevor der Use Case antwortet; Antwort bricht ab nach dem ersten Byte — **Ausgang:** offen; die nebenläufigen Fänger ordnen mit Kanälen und laufen im Ziel mit `-race` aus `slice-race-gate`, und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50`.
- Pfad ohne Service-Root, Pfad mit Prozent-Kodierung, `raw_query` mit ungültiger Kodierung — **Ausgang:** offen; Raster-Zelle, der Adapter reicht unverändert, der Use Case entscheidet (`ODR-E5005`).
- Request-Größe und Zeitgrenzen des Servers (`ReadHeaderTimeout`, Idle) — **Ausgang:** übergeben an [slice-http-lebenszyklus](slice-http-lebenszyklus.md) (entschieden vom Architect im Vorabentwurf, Fänger L1, L2 dort); der Kern setzt keine Grenze.
- Herunterfahren in zwei Schritten (Sendung von `slice-session-beenden`) — **Ausgang:** übergeben an [slice-http-lebenszyklus](slice-http-lebenszyklus.md), das die Sendung in seinem §1 annimmt; die Reihenfolge im Composition Root trägt `slice-bootstrap-m2`.
- Der Vorabentwurf des Architect nennt fünf Annahmen über `net/http`, die als erste Tests laufen (ungültiger Escape ergibt 400 vor dem Handler, `ErrAbortHandler` schließt ohne Abschluss, `Content-Type` nil unterdrückt das Sniffing, `Flush` sendet sofort, `ctx` bei ungelesenem Body) — **Ausgang:** offen; Adresse Implementer, bei Abweichung die Zelle an den Architect (`AGENTS.md` §3.12).
- Replay-Antworten gehen chunked (keine `Content-Length`, `Flush` je Stück) — **Ausgang:** akzeptiertes Negativ: `LH-QA-07.a` zählt `Content-Length` nicht zu den relevanten Headern, und `SPEC-135` Nr. 13 b nimmt die Länge bewusst aus; Folge: ein Client sieht `Transfer-Encoding: chunked` statt einer Länge.
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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (9×, Ausgang `verkörpert` in `AGENTS.md` §3.10); der Slice ist nebenläufig (Abbrüche, Herunterfahren): die nebenläufigen Fänger ordnen mit Kanälen und laufen im Ziel mit `-race` aus `slice-race-gate`, und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
