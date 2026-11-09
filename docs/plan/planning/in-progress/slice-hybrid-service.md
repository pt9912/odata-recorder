# Slice slice-hybrid-service: HybridUseCase mit Entscheidungsmatrix nach ODR-Code

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, `LH-FA-04`, `LH-FA-11`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` (Entscheidungsmatrix) · `LH-FA-04.c` (Adressierung, `X-Recorder-*`-Header) · `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080` · `SPEC-020`, `SPEC-021`, `SPEC-024` · `SPEC-120` (`UpstreamClient`, Fehler) · `SPEC-092` (`LINK_REWRITE`) · `ARC-002`, `ARC-003` · `architecture.md` §2.2, §2.5, §4.6.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-hybrid-service`, Worktree `/Development/odata-recorder-wt-hybrid-service`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `HybridUseCase` (Inbound Port in `src/hexagon/ports/inbound`, Dienst in `src/hexagon/services/hybrid`) ruft `MockUseCase` über dessen Inbound Port und entscheidet bei einem Fehler des Mock nach dem `ODR`-Code (nie nach dem HTTP-Status) mit der Policy aus `slice-fallback-policy`: Antwort oder Fehler bleiben (Erfolg, fachliche `404`, `409`, `412`, ungültige Anfrage), `ReplayUseCase` wird nur für die freigegebenen Kategorien gerufen, der `UpstreamClient` nur nach der Policy; ein Replay-Fehler außer `ODR-E5001`, insbesondere ein erschöpfter Key (`ODR-E5002`), wird unverändert ausgeliefert und löst nie einen Upstream-Zugriff aus (`LH-FA-12.b`).

**Übernimmt:** das Risiko der Upstream-Links einer Fallback-Antwort aus `welle-hybrid` §6 (`LINK_REWRITE` ist nach `SPEC-092` nur `off`); es steht als Risiko in §6 dieses Plans. Dazu die Pflichten des Aufrufers aus `SPEC-261` Nr. 10 (Sendung von `slice-fallback-policy-entscheidung`, dessen §1 *Ausdrücklich NICHT* diesen Slice nennt): der Dienst gibt der Policy dieselbe Zeichenkette als `Ziel`, die er sendet, und liest sie mit `url.Parse` wie die Policy (`SPEC-261` Nr. 10 a, Nr. 5 c; Sendung von `slice-fallback-policy-entscheidung`, Architect 2026-11-08), Fänger: der Referenzdienst empfängt unter dem Host, den die Policy freigab; er ruft `Entscheide` vor jedem Senden an den Upstream; er sendet `X-HTTP-Method`, `X-HTTP-Method-Override` und `X-Method-Override` nie (`SPEC-261` Nr. 10 c, Auftrag 9 entschieden); und der Replay-Fallback einer mutierenden Methode ändert den Zustand des Mock nicht (der Client sieht eine aufgezeichnete Erfolgsantwort, der Mock bleibt unverändert).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Policy (Hauptschalter, Allowlist, Routenfreigaben, Kategorien) und ihre Syntax — `slice-fallback-policy` (Prüfung der Einträge und `New`) und `slice-fallback-policy-entscheidung` (`Entscheide`); der Dienst ruft sie und entscheidet nichts selbst.
- Der Modus `hybrid` im Composition Root, die Konfiguration und der Listener — `slice-bootstrap-m4`; der Dienst läuft hier nur gegen Fakes der Ports.
- Speichern einer Fallback-Antwort und Datenabgleich zwischen Mock und Upstream — Out-of-Scope von `LH-FA-12`.
- Umschreiben absoluter Upstream-Links (`LINK_REWRITE`) — `SPEC-092` kennt nur `off`; das Risiko steht in §6.
- Spans, Audit und Metriken des Fallbacks — `welle-betrieb`; es gibt noch keinen Port `Telemetry`.

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

- [ ] Mock zuerst: Erfolg und jeder Fehler ohne Fallback-Auslöser (fachliche `404`, `409`, `412`, ungültige Anfrage, Fehler ohne `ODR`-Code) gehen unverändert an den Aufrufer; die Zuordnung (`X-Recorder-*`) geht unverändert an Mock und Replay.
- [ ] Replay-Fallback nur für die Kategorien der Policy (`ODR-E6001` und `ODR-E6002` als `unsupported`, `ODR-E6003` als `coverage_insufficient`): ein Treffer wird ausgeliefert, `ODR-E5001` geht weiter zur Entscheidung über den Upstream, jeder andere Replay-Fehler (Liste in `LH-FA-12.b`) unverändert; ein erschöpfter Key geht nie an den Upstream; ein Replay-Fallback einer mutierenden Methode lässt den Zustand des Mock unverändert (`SPEC-261` Nr. 10, Sendung von `slice-fallback-policy-entscheidung`).
- [ ] Upstream-Fallback über `UpstreamClient` nur nach Freigabe der Policy: alle `X-Recorder-*`-Header entfernt, Body des Requests nach der Entscheidung des Architect, Antwort-Strom genau einmal geschlossen, ohne Freigabe kein Aufruf des Upstream (Zähler 0); `Entscheide` läuft vor jedem Senden und bekommt als `Ziel` dieselbe Zeichenkette, die der Dienst sendet, und die drei Header zum Überschreiben der Methode gehen nie an den Upstream, Fänger: Schleife über die drei Namen × Schreibweisen (kanonisch, klein, groß) am Referenzdienst, der die empfangenen Header hält (`SPEC-261` Nr. 10 c, Sendung von `slice-fallback-policy-entscheidung`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Modus `hybrid` erreicht es erst mit `slice-bootstrap-m4`).
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
| `src/hexagon/ports/inbound/` | update | `HybridUseCase` |
| `src/hexagon/services/hybrid/` | update | Dienst neben der Policy aus `slice-fallback-policy` und `slice-fallback-policy-entscheidung` |
| `spec/spezifikation.md` | update (Architect) | Vertrag des Dienstes samt Randformen; Nachzug `LH-FA-12.b`, `architecture.md` §2.2 und §4.6 |
| Tests | neu | Matrix `LH-FA-12.b` × Zustände der Policy; die Fakes von `MockUseCase`, `ReplayUseCase` und `UpstreamClient` halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

| Zusage | Fänger |
|---|---|
| *(Architect: eine Zeile je Zusage, Randform und Fall des Vertrags)* | *(ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates)* |

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Vertrag von `HybridUseCase`:** Signatur in der Form der Use Cases der anderen Modi, Fehlerabbildung nach `SPEC-109`, Ort in der Spezifikation.
2. **Ablauf und Rangfolge:** Mock, `ODR`-Code, Policy, Replay, Policy, Upstream; welche Stufe welcher Code auslöst (`LH-FA-12.b`), Fehler ohne `ODR`-Code, Abbruch des Kontextes zwischen den Stufen.
3. **Body des Requests:** der Mock liest den Strom, ein Upstream-Fallback eines Requests mit Body braucht ihn erneut. Entscheiden: Pufferung (nur wenn die Policy einen Fallback erlauben kann; Grenze und Fehlercode bei Überschreitung) oder Fallback nur für Requests ohne Body.
4. **Zuordnung:** `X-Recorder-Workspace` geht unverändert an Mock und Replay; `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session` sind Pflicht, sobald ein Replay-Fallback konfiguriert ist (`LH-FA-04.c`): wann `ODR-E5003` entsteht (vor dem Aufruf des Mock oder erst im Fallback); alle `X-Recorder-*`-Header fallen vor dem Upstream weg.
5. **Antwort des Fallbacks:** ob sie sich als Fallback kennzeichnet (Header), welche Header der Upstream-Antwort weitergehen, und der Umgang mit absoluten Upstream-Links in der Antwort (`SPEC-092`, Risiko in §6).
6. **Fehler des Upstream im Fallback** (`ODR-E4001`, `SPEC-120` Nr. 8): unverändert ausliefern, kein Wiederholen, kein Rückfall auf Replay.
7. **Ströme:** die verworfene Antwort des Mock und jede Antwort, die nicht ausgeliefert wird, schließt der Dienst genau einmal, auch nach Abbruch.
8. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.
9. **Header des Upstream-Fallbacks** (`SPEC-261` Nr. 10, *Ungeprüft*): welche Request-Header der Dienst sendet, darunter die Header zum Überschreiben der Methode (`X-HTTP-Method-Override` und Verwandte): **Entschieden** (Architect, 2026-11-08): entfernen; `SPEC-261` Nr. 10 c nennt die drei Header, *Ungeprüft* nennt andere Namen und die Query.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen; ergänzt am 2026-11-08): *an diesen Slice* aus `slice-fallback-policy-entscheidung` (dieselbe Zeichenkette als `Ziel`, Aufruf vor jedem Senden, Header-Auswahl, Zustand des Mock beim mutierenden Replay-Fallback): angenommen, §1 *Übernimmt* und DoD (Punkte 2 und 3) tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie (der Ausschluss der Policy betrifft deren Syntax und Entscheidung, nicht den Aufruf). *Von diesem Slice an* `slice-bootstrap-m4` (Verdrahtung des Dienstes im Composition Root): angenommen, §1 und DoD Punkt 1 von `slice-bootstrap-m4` tragen sie. *An diesen Slice:* das Risiko der Upstream-Links aus `welle-hybrid` §6, angenommen (§1 *Übernimmt*, §6).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-fallback-policy`, `slice-fallback-policy-entscheidung`, `slice-mock-zuordnung`, `slice-mock-lesen`, `slice-mock-schreiben` und `slice-replay-service` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum diese sechs:**

- `slice-fallback-policy` und `slice-fallback-policy-entscheidung` (der Policy-Slice ist vor dem Code geschnitten, der zweite baut auf dem ersten): Typen, `New` und Entscheidungsfunktion, die der Dienst ruft (*Code-Abhängigkeit*).
- `slice-mock-zuordnung` und `slice-mock-lesen` (die beiden Teile des geschnittenen Plans): `slice-mock-zuordnung` legt den Inbound Port `MockUseCase` an (er liegt heute nicht in `src/hexagon/ports/inbound/`) und liefert `ODR-E6001`; `slice-mock-lesen` liefert `ODR-E6003`, nach dem der Dienst mit `ODR-E6001` entscheidet (*Code-Abhängigkeit, Port*).
- `slice-mock-schreiben`: legt fest, welche Mutationen der Mock mit `ODR-E6002` als nicht unterstützt abweist; ohne diesen Vertrag hat die Zeile „POST auf eine nicht unterstützte Operation“ der Matrix keinen Boden (*Vertrags-Abhängigkeit*, keine Typ-Abhängigkeit).
- `slice-replay-service`: legt den Inbound Port `ReplayUseCase` an (heute nicht vorhanden) und liefert `ODR-E5001`, `ODR-E5002` und `ODR-E5003` (*Code-Abhängigkeit, Port*).

**Stand der Bedingungen beim Start** (Planner, 2026-11-08, `main` = `33d3ba55`): alle sechs Start-Slices liegen in `done/` ([slice-fallback-policy](../done/slice-fallback-policy.md), [slice-fallback-policy-entscheidung](../done/slice-fallback-policy-entscheidung.md), [slice-mock-zuordnung](../done/welle-mock/slice-mock-zuordnung.md), [slice-mock-lesen](../done/welle-mock/slice-mock-lesen.md), [slice-mock-schreiben](../done/welle-mock/slice-mock-schreiben.md), [slice-replay-service](../done/welle-record-replay/slice-replay-service.md)). **Noch offen vor dem ersten Code-Commit:** der Vertrag des Architect in der Spezifikation (Aufträge 1 bis 8 in §3, die Randformen des Rasters in §6), die Tabelle *Zusage · Fänger* in §3 und, daraus folgend, der Block der Größenmarke (`SPEC-128` Nr. 3); der Architect arbeitet im Zweig nach dem Start-Commit (Nr. 3 der Übergangsregeln). **Gate-Kette (Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)):** kein Kettenslice in `in-progress/` (dort liegen nur `slice-if-match` und dieser Slice); dieser Slice ändert `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` und den Stempel nicht (Pfade in §3: `src/hexagon/ports/inbound/`, `src/hexagon/services/hybrid/`, `spec/spezifikation.md`, Tests), Linie B, kein Kettenslice. Parallel läuft `slice-if-match` (`src/hexagon/services/mock/`, `src/hexagon/model/`, `src/adapters/driving/httpodata/`); kein gemeinsamer Pfad aus §3 außer `spec/spezifikation.md` und gegebenenfalls `src/hexagon/ports/inbound/` (Port `MockUseCase` unverändert zu lassen ist Voraussetzung dieses Slice; ändert `slice-if-match` ihn, nimmt der spätere Zweig `main` vor dem Handoff auf). Kennungen: Linie B vergibt aus `SPEC-260` bis `SPEC-299` (zuletzt `SPEC-282` belegt, `SPEC-261` ist die Policy); weitere Nummern für den Vertrag von `HybridUseCase` vergibt der Architect aus diesem Bereich.

**Größenmarke** (Planner, 2026-11-08, `SPEC-128` Nr. 1, **vorläufig, Prosa, kein Block**): die Tabelle *Zusage · Fänger* fehlt; der Block mit `pfad`, `produktcode`, `fakes`, `basis`, `schaetzung` und `marke` entsteht, wenn der Architect die Zeilen nach Klassen und die Produktcode-Erwartung nennt, und der Planner setzt die Marke vor dem ersten Code-Commit mit `make groessenmarke` (*S* = 85·*a* + 80·*b* + 200·*f* + *p*; *M* = 100 · ⌈135·*S* / 10000⌉; Schnittfrage genau dann, wenn *S* ≥ 2223). Orientierung ohne Bindung: drei Fakes der Ports (`MockUseCase`, `ReplayUseCase`, `UpstreamClient`) und ein Referenzdienst mit Verbindungszähler und Header-Haltepunkt als Test-Server, also *f* = 4; Raster 4 Zeilen × 5 Spalten, Zusatztabelle über Matrix × Zustände der Policy und Stufen, 25 bis 35 Zeilen erwartet, davon 2 nebenläufig (gleichzeitige Aufrufe, Schließen der Ströme nach Abbruch). Bei 25 Zeilen (*a* = 2, *b* = 23, *f* = 4, *p* = 500) wäre *S* = 3310 und *M* = 4500, bei 35 Zeilen (*a* = 2, *b* = 33, *f* = 4, *p* = 600) *S* = 4210 und *M* = 5700, **beide mit Schnittfrage**; selbst bei 16 Zeilen (*a* = 2, *b* = 14), *f* = 3 und *p* = 400 läge *S* bei 2290 und über der Grenze. Die Schnittfrage ist damit praktisch sicher: Beantwortet der Architect sie mit „ja“, wird **vor dem Code** geschnitten (`AGENTS.md` §3.14). Ein Schnitt entlang der Stufen ist ein Vorschlag, den der Architect prüft: (1) Mock zuerst und Replay-Fallback (Dienst, Port `HybridUseCase`, Fakes von `MockUseCase` und `ReplayUseCase`) gegen (2) Upstream-Fallback (Body, `X-Recorder-*`, Header der Methode, Strom der Antwort, Fake von `UpstreamClient` und Referenzdienst); sie teilen die Entscheidung nach `ODR`-Code und den Aufruf von `Entscheide`, und teilen sie einen Mechanismus, den der zweite Teil neu schriebe, steht der Grund in diesem Abschnitt. Der Block des ersten Teils braucht dann eine `basis` nach dem Start-Commit; die Sendungen aus `slice-fallback-policy-entscheidung` verteilen sich nach Stufe (Pflicht 1 bis 3 auf Teil 2, Mutations-Zustand auf Teil 1).

**Nicht gebraucht:** `slice-bootstrap-m3`, `slice-snapshot-use-case`, `slice-mock-abnahme` und M3, denn der Dienst läuft gegen Fakes der Ports und der Composition Root gehört `slice-bootstrap-m4`; auch M2 (`slice-bootstrap-m2`) braucht der Dienst nicht selbst, nur das, was `slice-replay-service` und `slice-mock-lesen` von ihr voraussetzen. **Ohne die Ports des Mock startet der Slice nicht.**

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt (*Größenmarke* oben: vorläufig, Prosa, Schnittfrage praktisch sicher).
- `in-progress` → `open` (blockiert — Carveout?): der Body eines Requests lässt sich für den Upstream-Fallback nicht erhalten, ohne den Port `MockUseCase` zu ändern (Auftrag 3), oder die Zuordnung der Replay-Header im Hybrid-Modus ist nicht entschieden (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): „(Vertrag, offen)“ heißt, der Architect trägt den Ort in der Spezifikation vor dem Übergang `next` → `in-progress` ein; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Aufruf des Mock und Entscheidung nach dem Fehlercode | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Replay-Fallback (Zuordnungs-Header, Kategorie, Replay-Fehler) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Upstream-Fallback (Body, `X-Recorder-*`, Antwort des Upstream) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Schließen der Ströme (verworfene Mock-Antwort, ausgelieferte Antwort) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- Der Body eines Requests lässt sich nur einmal lesen; der Mock liest ihn, ein Upstream-Fallback braucht ihn erneut — **Ausgang:** offen; Adresse: Architect (Auftrag 3).
- Eine Fallback-Antwort kann absolute Upstream-Links tragen, `LINK_REWRITE` kennt nur `off` (`SPEC-092`; Risiko aus `welle-hybrid` §6) — **Ausgang:** offen; Adresse: Architect (Auftrag 5): Kennzeichnung der Antwort und Hinweis im Handbuch-Kapitel von `slice-bootstrap-m4`.
- Die Zuordnung der Replay-Header im Hybrid-Modus (`ODR-E5003` vor dem Mock-Aufruf oder erst im Fallback) hat keinen Ort in `LH-FA-04.c` — **Ausgang:** offen; Adresse: Architect (Auftrag 4).
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): die Fakes von `MockUseCase`, `ReplayUseCase` und `UpstreamClient` tragen jede Schnittstelle des Originals (der Body als `ReadCloser`, `Close` gezählt), zählen jeden Aufruf und halten jedes Argument fest; ein Fake, der panikt, verdeckt den benannten Test nicht — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (17×): die Matrix hat viele Zweige, der Dienst hält keinen Zustand; gleichzeitige Aufrufe laufen als Fall unter `make test-race` — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- **FRAGEN AN DEN USER (drei, aus `slice-fallback-policy-entscheidung`; der Planner entscheidet sie nicht, sie bremsen den Start nicht, die Antwort ändert `SPEC-261` und, soweit sie den Dienst färbt, Raster und Tabelle dieses Slice):**
  - **FRAGE 1 (User): Replay-Fallback auch für mutierende Methoden?** Bis zur Antwort gilt `SPEC-261` Nr. 4 b = **ja**. Folge für diesen Slice: ein Replay-Fallback einer mutierenden Methode liefert eine aufgezeichnete Erfolgsantwort, der Mock bleibt unverändert (DoD Punkt 2, Fänger zählt die Schreibzugriffe des Mock-Fakes). Antwortet der User „nur lesend“, entfällt die Zeile im Replay-Fallback; Zeilen von Raster und Tabelle ziehen nach.
  - **FRAGE 2 (User): `ODR-E6006` (V2: Eigenschaft außerhalb des Umfangs) als Auslöser?** Bis zur Antwort **nein**, kein Fallback (`SPEC-261` Nr. 3 d). Folge: der Dienst liefert `ODR-E6006` unverändert aus; bei „ja“ kommt eine Stufe der Matrix hinzu.
  - **FRAGE 3 (User): Lücke im Lastenheft?** Der Replay-Fallback hat dort weder Anforderung noch Akzeptanzkriterium (Change Request vor `Accepted` oder Bestätigung); der Dienst baut ihn aus `LH-FA-12.a` und `LH-FA-12.b` der Spezifikation.
  - **Ausgang:** weiter offen; Adresse: User; Beleg: `slice-fallback-policy-entscheidung` §6 und §7.
- Sendungen beim Start (Planner, 2026-11-08, `AGENTS.md` §3.13): *von* `slice-fallback-policy-entscheidung` (`SPEC-261` Nr. 10 a bis c, Mock-Zustand beim mutierenden Replay-Fallback) und *von* `slice-fallback-policy` (Fänger: der Referenzdienst empfängt unter dem Host, den die Policy freigab) — **Ausgang:** eingetreten als Sendung, angenommen: §1 *Übernimmt* und DoD Punkte 2 und 3 tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie; Fänger-Zeilen der Sendung (Host-Fänger am Referenzdienst, Schleife 3 Namen × Schreibweisen, Zähler der Schreibzugriffe des Mock) gehören in die Tabelle des Architect. Adresse: Architect.
- Merge-Reihenfolge: `slice-if-match` läuft parallel; gemeinsam sind nur `spec/spezifikation.md` und, falls `slice-if-match` den Port `MockUseCase` berührt, `src/hexagon/ports/inbound/` — **Ausgang:** offen; Adresse: Planner (wer später merged, nimmt `main` vor dem Handoff auf).
- Gate-Kette (Nr. 7 der Übergangsregeln): der Slice ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette in §3.
- Größe: vorläufige Prosa-Marke, Schnittfrage praktisch sicher (§4) — **Ausgang:** offen; Adresse: Architect (Zeilen nach Klassen, Produktcode-Erwartung, Schnitt nach Stufen prüfen, vor dem Code), danach Planner (Block und Marke vor dem ersten Code-Commit).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-08; Zähler sind die Dateien unter `evidence/`). Der Slice hat drei Fakes und viele Zweige: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×) sind verkörpert (`.claude/commands/implement-slice.md` Schritt 14, `.harness/skills/reviewer.md`) und gelten für die Fakes (§6); `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (17×, offen, Schwelle erreicht, Vorlage für den Reviewer-Skill beim User) für die Zweige der Matrix, die kein Lauf ausführt (§6). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (57×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (15×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (35×, `AGENTS.md` §3.16), `BEO-REPO/zustand-des-arguments-als-dimension` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (14×) für die Pufferung des Bodys (Auftrag 3). Offen und mit dem Slice berührt: `BEO-REPO/sendung-nennt-nicht-alle-pflichten-des-vertrags` (1×; der Plan trägt die vier Pflichten aus Nr. 10 und prüft sie gegen jede Nummer des Vertrags, die eine Pflicht des Aufrufers trägt) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (17×, geplant in `slice-guard-schreibwege`; die Betriebsregeln in §3 gelten bis dahin). Der Slice schreibt keine Zähler, er trägt Evidenz-Dateien bei der Closure bei.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
