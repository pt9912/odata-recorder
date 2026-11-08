# Slice slice-admin-authentisierung: Admin-Rahmen, Teil 3 — Ablauf, Authentisierung und Antwort

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-QA-04`, `LH-RB-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Bearer-Token, zeitkonstanter Vergleich, 401 ohne Zustandsänderung) · `LH-QA-04.a` · `SPEC-238` Nr. 1 (`New` kopiert das Token; `Vergleich` nil), Nr. 2 (`Header` ohne `Authorization`), Nr. 3 (Ablauf: a, c, d, e, f), Nr. 4 (Authentisierung), Nr. 5 (Header und Erfolgsantwort), Nr. 13 (a, b: das Token und der Kandidat erreichen weder Body noch Header) · `SPEC-005` (Zeile `GET /healthz`, Schlusssatz) · `SPEC-064` (`ODR-E2101`) · `SPEC-006` (Fehlerobjekt) · `SPEC-114` Nr. 6 (Korrelations-ID) · `ARC-007`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie. `SPEC-238` steht auf `main` (seit dem Merge von `slice-admin-listener-token`).

**Verantwortlich:** pt9912 (Zweig `slice-admin-authentisierung`, Worktree `/Development/odata-recorder-wt-admin-authentisierung`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07. **Architect:** Tabelle, Raster und Randformen gegen `SPEC-238`, `SPEC-240` und den Code von `slice-admin-rahmen` geprüft und geschärft am 2026-11-08 (Änderungen gegenüber dem Planner-Stand: Zeilen K5, U10, U11; U1 mit 20 statt 14 Formen; A7 auf die Stelle gelegt, an der sie erreichbar ist; Mengen und Raster in §3 und §6).

**Herkunft:** Teil 3 des Schnitts von [slice-admin-listener-token](../done/slice-admin-listener-token.md) (siehe dort und [slice-admin-rahmen](../done/slice-admin-rahmen.md) für die Reihenfolge der sechs Teile). Der Plan trägt die Zeilen K4, A1, A3 bis A6, U1 bis U9 und E2 bis E4 der Tabelle des Architect (18 von 102); sie stehen im Wortlaut in §3, geschärft um K5, A7 (Sendung), U10 und U11 (22 Zeilen). Gegenüber dem Vorschlag des Architect liegen A2 (Zeile für den angehaltenen Server) bei `slice-admin-lebenszyklus`, E1 und E5 (Schleifen über die Fehler der Routen) bei `slice-admin-fehlerabbildung` und N1 (gleichzeitige Requests) bei `slice-admin-abweisen-audit`: jede dieser Zeilen braucht Code, den erst der genannte Teil liefert.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Rahmen aus `slice-admin-rahmen` bedient den Ablauf eines Requests nach `SPEC-238` Nr. 3: `GET /healthz` ohne Token (nur `GET`, nur genau dieser Pfad), die Authentisierung vor der Zuordnung (`Authorization: Bearer <token>`, genau eine Zeile, Schema in jeder ASCII-Schreibweise, Kandidat Byte für Byte, `Konfig.Vergleich` mit `vergleiche` auf SHA-256 und `subtle.ConstantTimeCompare`; nil heißt `vergleiche`), fehlendes oder falsches Token als 401 `ODR-E2101` mit `WWW-Authenticate: Bearer` ohne Aufruf der Route und ohne Zustandsänderung, den Aufruf des `Handler` genau einmal je gültigem Request und die Antwort nach Nr. 5 (Header `Content-Type`, `Content-Length`, `X-Correlation-Id`, `Allow`, `WWW-Authenticate`; Erfolgsantwort der Route). `New` kopiert das übergebene Token. Das 401 schreibt in diesem Teil noch keinen Audit-Eintrag und zählt nicht. **Übernimmt (Sendung von `slice-admin-rahmen`, Verifier-Befund 1, 2026-11-08, `AGENTS.md` §3.13):** den Fänger zu Z8 des Rahmens für `Authorization` mit mehreren Zeilen — die gehaltene `Anfrage` trägt auch dann keine `Authorization`-Zeile, wenn der Request zwei Zeilen führt (Zeile A7). **Wo A7 erreichbar ist (Architect):** Ein Request mit zwei Zeilen erreicht nach diesem Teil keine Route mehr (401, U1); der Filter ist deshalb allein auf der Funktion `anfrage` beobachtbar, die die Anfrage bildet. A7 ruft sie über `export_test.go` direkt; die Prüfung über `ServeHTTP` führen A7 und Z8 mit einer gültigen kanonischen Zeile und weiteren Schlüsseln in anderer Schreibweise.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der angehaltene Server (503 `ODR-E5010` vor der Authentisierung), `Serve`, `Anhalten`, `Schliessen`, Grenzen des `http.Server` — `slice-admin-lebenszyklus`; ohne `Anhalten` hat die Zeile A2 keinen Aufrufer.
- Abbildung der Fehler der Routen, Panik, Zählen (`CountError`, auch für 401, 404 und 405) und die Schleifen über die Fehlerobjekte (E1, E5) — `slice-admin-fehlerabbildung`; der Teil schreibt das Fehlerobjekt für 401 über die eine Stelle aus `slice-admin-rahmen`. Die Route mit Fehler (500 aus dem Code des Rahmens) gehört als Requestart zu *R* und zu den Zellen von E2, weil dieselbe Stelle ihre Header schreibt; ihre Abbildung prüft dieser Teil nicht.
- Der Audit-Eintrag zu `ODR-E2101` (Art `admin_abgewiesen`, `SPEC-238` Nr. 8) und die Zeile N1 (gleichzeitige Requests, die `Append` und `CountError` einschließt) — `slice-admin-abweisen-audit`; bis dahin schreibt der Adapter keinen Eintrag (Risiko in §6, nichts erreicht den Listener).
- Konfiguration, Routen, Zuordnung, Code `ODR-E2102` — `slice-admin-rahmen` (Voraussetzung dieses Teils); die Token-Datei — `slice-admin-listener-token` (Teil 1); dieser Teil ruft `LadeToken` nie und liest keine Datei.
- Die Begrenzung von Fehlversuchen — das Lastenheft nennt sie nicht (`LH-FA-13.a`); benannte Lücke, Adresse User (§6), kein Code.
- Die Endpunkte, TLS und das Lesen von `ADMIN_*` — `slice-admin-sessions`, `slice-admin-imports`, `slice-admin-workspaces`, `slice-admin-tls`, `slice-bootstrap-m6`; das Binary bleibt unverändert.

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

- [ ] Ablauf: `KorrelationsID` genau einmal je Request vor allem anderen (A1), `GET /healthz` ohne Token und nur `GET` auf genau diesem Pfad (A3, A4), Authentisierung vor Zuordnung (A5), `Handler` genau einmal je Request mit gültiger Zuordnung und Token (A6); die Schleifen über die Requestarten *R* tragen die Formen dieses Teils (healthz, 20 ungültige Formen von `Authorization`, falsches Token, Route mit 2xx, Route mit Fehler, 404, 405: 26).
- [ ] Authentisierung: `Authorization: Bearer <token>` nach `SPEC-238` Nr. 4 (genau eine Zeile, Schema in jeder ASCII-Schreibweise, Kandidat Byte für Byte), `Vergleich` genau einmal für einen nichtleeren Kandidaten und allein sein Ergebnis entscheidet, nil heißt `vergleiche` (K5), `vergleiche` mit `subtle.ConstantTimeCompare` auf SHA-256 und sonst kein Vergleich des Tokens im Paket, die Authentisierung für jede Methode und jeden Pfad außer `GET /healthz`, 401 `ODR-E2101` mit `WWW-Authenticate: Bearer` ohne Aufruf der Route und ohne Zustandsänderung; weder Token noch Kandidat erreichen Body oder Header (U11); `New` kopiert das Token (K4).
- [ ] Sendung von `slice-admin-rahmen` (Befund 1 des Verifiers): `Authorization` mit mehreren Zeilen (auch gleichlautend, auch mit leerem Wert, auch in jeder Schreibweise des Schlüssels) erreicht keine Route, die `Anfrage` trägt keine Zeile davon (Zeile A7, auf der Funktion `anfrage` und über `ServeHTTP`).
- [ ] Antwort: `Allow` nur bei 405, `WWW-Authenticate` nur bei 401, `X-Correlation-Id` an jeder Antwort mit der ID des Requests, Erfolgsantwort der Route mit Status, `Content-Type` und `Content-Length` (E2 bis E4); die Zählprüfung der 27 Zellen von E2 erreicht `slice-admin-fehlerabbildung` (dieser Teil: 18).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Für diesen Slice: kein neuer Code (`ODR-E2101` steht in `harness/katalog-ausnahmen.txt`); das Binary bleibt unverändert (kein Aufruf von `admin` aus `bootstrap`, `src/bootstrap` ohne Änderung), in §7 vermerkt.
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
| `src/adapters/driving/admin/` | update | Ablauf von `ServeHTTP` (healthz, Authentisierung vor Zuordnung, Aufruf des `Handler`), Feld `Konfig.Vergleich`, `vergleiche`, Header der Antwort (`X-Correlation-Id` an jeder, `WWW-Authenticate`, `Content-Type`/`Content-Length` der Erfolgsantwort) und Erfolgsantwort (`SPEC-238` Nr. 1, 3 a, c, d, e, f, Nr. 4, Nr. 5) |
| Tests | neu / update | Adapter gegen `ServeHTTP` mit `httptest`; die Fakes `zaehlendeIDs` und `routeFake` aus `slice-admin-rahmen` (der `useCaseFake` hinter der Testroute zählt jede Operation, U8; er entsteht in diesem Teil, `src/adapters/driving/admin/auth_helfer_test.go`); `Vergleich` als zählende Funktion des Tests; AST-Test für `vergleiche` und die Verwendung des Tokens (U6); `export_test.go` führt `anfrage` und `vergleiche` für A7 und U5 heraus; der Adapter importiert weder Adapter noch Service (`make a-check`). **Bestandstests:** `anfrageMit` setzt schon `Authorization: Bearer <32 mal t>`, die Tests des Rahmens erreichen ihre Route also weiter; zu ändern ist der Fall von Z8, der `Authorization` als einzigen Träger mit dem Canary überschreibt (er bekommt die gültige kanonische Zeile und den Canary in einem Schlüssel anderer Schreibweise), und jeder Test, der 404 oder 405 ohne Header erwartet (jetzt 401, A5). Das Token des Fixtures (`tokenGueltig`, 32 mal `t`) ist für U3 zu schwach (kein Großbuchstabe, kein innerer Leerraum): U3 und U5 bekommen Tokens aus Bytewerten (`AGENTS.md` §3.18) mit Großbuchstaben, Ziffern, innerem Leerzeichen und einem Token mit BOM-Präfix (Bytes EF BB BF). |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../done/slice-admin-listener-token.md); geschärft 2026-11-08, siehe Kopf); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut (A Ablauf, U Authentisierung, E Antwort, K Konfiguration). **Wachsende Mengen:** Die Requestarten *R* (A1, E4) und die Antwortarten (E2) wachsen mit den Teilen. *R* dieses Teils (26): healthz (1), jede der 20 Formen von U1, falsches Token (`Vergleich` falsch, 1), 404, 405, Route mit 2xx, Route mit Fehler (500); `slice-admin-lebenszyklus` ergänzt 503, `slice-admin-fehlerabbildung` die Route mit Fehlercode und die Route mit Panik. Die Antwortarten von E2 dieses Teils: 401, 404, 405, 200, healthz, 500 × die drei Header (WWW-Authenticate, Allow, Connection) = 18 Zellen; 503 (Lebenszyklus) sowie 400 und Fehler der Route mit Code (Fehlerabbildung) kommen dazu, die Zählprüfung 9 × 3 = 27 gilt am Ende. Jeder Teil trägt seine Formen in die Schleifen ein; die Elemente der Vorgänger rechnet seine Zählprüfung mit.

**Grenzen der Zählung in diesem Teil:** Die Tests dieses Teils behaupten für 401, 404, 405 und 500 weder 0 noch 1 Aufruf von `Audit.Append` oder `CountError`, denn `slice-admin-abweisen-audit` und `slice-admin-fehlerabbildung` machen daraus 1; eine Behauptung 0 müssten sie umkehren. Fest bei 0 bleibt beides für `GET /healthz` (A3).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| K4 | (Nr. 1) `New` kopiert das `Token` | Quellslice nach `New` mit 0-Bytes überschrieben: das ursprüngliche Token authentisiert (200), das überschriebene nicht (401) | Kopie entfernt | andere |
| K5 | (Nr. 1, 4 e) `Vergleich` nil heißt `vergleiche`: das richtige Token authentisiert, jede Abweichung nicht | `Konfig` ohne `Vergleich`: richtiges Token 200; Kandidaten Token + ein Byte, Token ohne letztes Byte, letztes Byte geändert, Präfix des Tokens, ein Byte: je 401 | nil führt zu `true`, zu `false` oder zu einem Aufruf von nil | andere |
| A1 | (Nr. 3 a) `KorrelationsID` wird genau einmal je Request gerufen, vor allem anderen; die anderen Operationen von `IDs` nie | `zaehlendeIDs` (fünf Operationen gezählt); Schleife über Requestarten *R* (26, Liste oben): Zähler 1 und 0; Haltepunkt in `KorrelationsID` sieht `Vergleich` und Route noch bei 0 | ID erst nach der Authentisierung; ein zweiter Aufruf im Fehlerpfad | andere |
| A3 | (Nr. 3 c) `GET /healthz` antwortet 200 mit `{"status":"ok"}` und liest `Authorization` nicht | Schleife über die 20 Formen von M-Header und das richtige Token (21): 200 und Literal; `Vergleich` 0, Audit 0, `CountError` 0 | healthz hinter der Authentisierung | andere |
| A4 | (Nr. 3 c) Nur `GET` auf genau `/healthz` ist frei; die Query ist gleichgültig; `/healthz/`, `/HEALTHZ`, `/healthz/x`, `/healthz%2F` und jede andere Methode gehen durch die Authentisierung | M-healthz: Methoden (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE, `get`) × Pfade (`/healthz`, `/healthz/`, `/HEALTHZ`, `/healthz/x`, `/healthz%2F`, `/healthzx`, `/xhealthz`, `/health`, `//healthz`) ohne Token: 401, außer `GET /healthz` mit Query (keine, `?a=1`, `?`) 200; mit Token: 405 mit `Allow: GET` für jede andere Methode auf `/healthz`, sonst 404; Zählprüfung 81 und 3 | `HasPrefix`, `HasSuffix` oder `Contains` statt Gleichheit; Methode nicht oder ohne Beachtung der Schreibweise geprüft | andere |
| A5 | (Nr. 3 d, e) Die Authentisierung liegt vor der Zuordnung: ohne Token und mit falschem Token 401 für registrierten und unbekannten Pfad und falsche Methode, mit Token 404 und 405 | Menge von Z6 (32) × (ohne Token, mit falschem Token, mit Token): Orakel Status als Literal | Zuordnung vor der Authentisierung | andere |
| A6 | (Nr. 3 f) Der `Handler` läuft genau einmal je Request mit gültiger Zuordnung und Token | `routeFake` zählt: Schleife über alle Routen des Test-Servers (mindestens 5, eine je Methode GET, POST, PUT, PATCH, DELETE): 1 | Wiederholung nach einem Fehler | andere |
| A7 | (Nr. 2 `Header` ohne `Authorization`; Sendung von `slice-admin-rahmen`, Z8, Befund 1) Die `Anfrage` trägt keine Zeile von `Authorization`, in keiner Schreibweise des Schlüssels und bei keiner Anzahl und Form der Werte; jeder andere Header bleibt mit allen Werten | **(a) Funktion `anfrage`** (über `export_test.go`): 8192 Schreibweisen des Schlüssels × 8 Wertformen (ein Wert; zwei verschiedene; zwei gleichlautende; zwei, der erste leer; zwei, der zweite leer; drei; ein leerer Wert; leere Liste) = 65536 Fälle, daneben die 48 Nachbarnamen aus Z8 (darunter `Proxy-Authorization`, `X-Authorization`, `Authorization-Info`) × dieselben 8 Wertformen = 384 Fälle: Orakel unabhängig vom Code, kein Schlüssel der Schreibweisen-Menge in `Header`, jeder Nachbar mit gleicher Wertliste (auch leer); Zählprüfung 65536 und 384. **(b) `ServeHTTP`:** die gültige kanonische Zeile plus ein Schlüssel jeder 8192 Schreibweisen mit dem Canary (die Route läuft, `Header` trägt weder Schlüssel noch Canary noch Token), die Nachbarn mit leerem Wert und leerer Liste bleiben (Z8 des Rahmens, erweitert); mit zwei kanonischen Zeilen 401 und Route 0 (U1) | Filter entfernt nur einzeilige `Authorization` (rot allein in (a): ein Request mit zwei Zeilen erreicht die Route nie); Filter entfernt nur bei nichtleerem Wert; Slice der Länge 2, 3 oder 0 nicht gefiltert; Filter verwirft jeden Header mit leerem Wert (rot an den Nachbarn mit leerem Wert) | andere |
| U1 | (Nr. 4 a, b) Jede ungültige Form von `Authorization` ergibt 401 ohne Aufruf von `Vergleich` und ohne Route | Menge M-Header (20): 1 fehlt (kein Schlüssel); 2 Schlüssel mit leerer Liste; 3 leerer Wert; 4 nur Leerzeichen als Wert (direkter Aufruf, `net/http` kürzt sie auf der Leitung); 5 `Bea` (kürzer als das Schema); 6 `Bearer` allein; 7 `Bearer` mit drei Leerzeichen dahinter (leerer Kandidat, direkter Aufruf); 8 `Basic <token>`; 9 `Token <token>`; 10 `Bearer<token>`; 11 Schema, Tab, Token; 12 Token ohne Schema; 13 zwei gleichlautende gültige Zeilen; 14 zwei gültige Zeilen verschieden geschrieben (`Bearer`, `bearer`); 15 gültige Zeile, danach `Basic x`; 16 `Basic x`, danach gültige Zeile; 17 leere Zeile, danach gültige; 18 drei gültige Zeilen; 19 `X-Authorization` statt `Authorization`; 20 der Schlüssel `authorization` (nichtkanonisch) allein mit gültigem Wert (direkter Aufruf; `net/http` liefert nur den kanonischen Schlüssel, die Authentisierung liest allein diesen: 401 fehlt); Zählprüfung 20; Orakel Literal des Vertrags; die Formen 13 bis 18 sind die Wirkung von Nr. 4 a, die Form 20 hält die Lesart fest. **Klasse „anderes Schema“ (Implementer, Mutationslauf):** die Formen 8 und 9 (`Basic `, `Token ` mit fünf Buchstaben) scheitern schon an der Prüfung des Leerzeichens an Stelle 6 und fangen „Schema nicht geprüft“ nicht; der Generator `TestSchemaJedesByte` führt an jeder der 6 Stellen des Schemas jedes Byte außer den beiden Schreibweisen des Buchstabens (6 × 254 = 1524) und jedes Byte hinter dem Schema (256, allein das Leerzeichen authentisiert): 401, `Vergleich` 0 | Prüfung auf genau eine Zeile zu `>= 1`; zu `!= 2`; Schema nicht geprüft; Länge des Wertes nicht gegen das Schema geprüft (Panik bei `Bea`); Schlüssel ohne Beachtung der Schreibweise gesucht (Form 20) | andere |
| U2 | (Nr. 4 b) Das Schema `Bearer` gilt in jeder ASCII-Schreibweise und mit einem oder mehreren Leerzeichen | Schema in allen 64 Schreibweisen (2 hoch 6) × Leerzeichen (1, 2, 5): 200, Kandidat gleich dem Token; Zählprüfung 192 | Schema exakt verglichen; nur ein Leerzeichen; Schreibweise nur für `bearer`, `BEARER` | andere |
| U3 | (Nr. 4 b) Der Kandidat ist der Rest Byte für Byte: Groß- und Kleinschreibung, innerer Leerraum, Präfix, Suffix, Bytes jenseits von ASCII und das letzte Byte zählen | M-Kandidat (7), Token aus Bytewerten mit Großbuchstaben: (1) Token + `x`; (2) ohne letztes Byte; (3) erstes Byte klein statt groß; (4) Rest mit führendem Tab (`Bearer`, Leerzeichen, Tab, Token); (5) Rest `Bearer <token>` (Schema doppelt); (6) Token mit innerem Leerzeichen als erwartetes Token, gleicher Kandidat: 200; (7) erwartetes Token mit BOM-Präfix: Kandidat mit BOM 200, ohne BOM 401; Orakel Literale; die Formen 1 bis 5 und 7 (ohne BOM) rufen `Vergleich` einmal, mit dem Rest als Kandidat; Klasse „Byte des Kandidaten“ (Implementer): `TestKandidatAlleBytes`, ein Token aus allen 256 Bytewerten authentisiert (200), und jede der 256 Stellen um eins erhöht ergibt 401; `TestTokenLaengen` führt die Längen 32, 33 und 8192 mit je einem gleichen, einem um ein Byte längeren, einem kürzeren und einem im letzten Byte geänderten Kandidaten | Kandidat getrimmt oder in Kleinbuchstaben; Rest hinter dem Tab abgeschnitten; Bytes jenseits von ASCII verworfen | andere |
| U4 | (Nr. 4 c) `Vergleich(Token, Kandidat)` wird für einen nichtleeren Kandidaten genau einmal gerufen, und allein sein Ergebnis entscheidet; für einen Verstoß aus (a) und (b) nie | `Vergleich` als zählende Funktion des Tests: Rückgabe `true` bei falschem Kandidaten ergibt 200, `false` bei richtigem 401; die Argumente (erwartet, gegeben) Byte für Byte; ein Aufruf; Kandidat der Länge 1 (`Bearer x`) ruft einmal mit `x`, der leere Kandidat (U1 Form 7) nie; für die 20 Formen von U1 0 Aufrufe | der Aufruf durch `==` ersetzt (Zähler 0); die Argumente vertauscht; Rückgabe mit einem eigenen Vergleich verknüpft (`&&`, `\|\|`); Grenze `>= 1` statt `> 0` am Kandidaten verschoben | andere |
| U5 | (Nr. 4 e) `vergleiche` liefert `true` genau bei gleichen Werten | Generator über Formen (gleich; kürzer; länger; gleich lang an erster, mittlerer, letzter Stelle abweichend: 6) × Längen 32, 33, 64 (18 Fälle) und leer/leer, leer/nichtleer, nil/leer (3): Orakel `string(a) == string(b)` im Test; Zählprüfung 21 | Rückgabe invertiert; nur die Länge verglichen; nur das erste Byte verglichen | andere |
| U6 | (Nr. 4 e) `vergleiche` ruft `subtle.ConstantTimeCompare` auf `sha256.Sum256` beider Werte, und das Paket vergleicht Token und Kandidat sonst nirgends | AST-Test (`go/parser`) über die Dateien des Pakets ohne `_test.go`: (1) in `vergleiche` genau 1 Aufruf von `ConstantTimeCompare` und 2 von `Sum256`; (2) im Paket 0 Aufrufe von `bytes.Equal`, `bytes.Compare`, `bytes.EqualFold`, `strings.Compare`, `slices.Equal`, `slices.Compare`, `reflect.DeepEqual`, `subtle.ConstantTimeEq`, und `strings.EqualFold` allein mit dem Literal `"Authorization"` als Argument (der Filter von `anfrage`); (3) kein Ausdruck `==` oder `!=` mit einem Operanden, dessen Bezeichner oder Feld `token` oder `kandidat` heißt (ohne Beachtung der Schreibweise), auch unter `string(…)` und `[]byte(…)`; Zählprüfung gegen die Quelle: die Zahl der geprüften Dateien gleich der Zahl der Dateien ohne `_test.go` im Verzeichnis | `bytes.Equal` statt `ConstantTimeCompare`; `Sum256` entfernt; `kandidat == string(s.token)` neben dem Aufruf | andere |
| U7 | (Nr. 4 d) Die Authentisierung gilt für jede Methode und jeden Pfad außer `GET /healthz` | Methoden (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE) × Pfade (registriert, unbekannt, `/`, `/__admin`, `*` bei OPTIONS) ohne Token: 401; Zählprüfung 8 × 5 | Ausnahme für HEAD oder OPTIONS | andere |
| U8 | (Nr. 4; `LH-FA-13.a`) Ein fehlendes oder falsches Token ergibt 401 ohne Zustandsänderung und ohne Aufruf des Use Case | `routeFake` und der `useCaseFake` hinter der Testroute (zählt jede Operation, Lesen wie Schreiben) nach jeder Form von U1, jeder Zelle von U7 und `Vergleich` gleich `false`: Summe 0 | die Route vor der Prüfung gerufen | andere |
| U9 | (Nr. 4, 5) Das 401 trägt das Fehlerobjekt `ODR-E2101` und `WWW-Authenticate: Bearer` | Fall mit den Literalen `{"error":{"code":"ODR-E2101","message":"Admin authentication missing or invalid","target":"admin"}}`, Status 401, `Content-Type: application/json`, `Content-Length` als Zahl des Literals (fest im Test, nicht `len` des Codes) und Header `WWW-Authenticate` gleich `Bearer` | Header `WWW-Authenticate` fehlt; Wert `Bearer realm=…` | andere |
| U10 | (Nr. 4 b) Leerraum am Rand des Wertes entfernt `net/http`; zwei `Authorization`-Zeilen auf der Leitung kommen als zwei Werte an | Server über `httptest.NewServer(srv)` auf Loopback mit rohem TCP-Client (kein Client, der Zeilen zusammenfasst): (1) `Bearer <token>` mit Leerzeichen und Tab dahinter: 200; (2) zwei `Authorization`-Zeilen, beide gültig: 401 und Route 0; (3) `authorization` in Kleinschreibung auf der Leitung mit gültigem Wert: 200 (der Server kanonisiert) | Annahme zur Kanonisierung bricht (Wechsel des Servers oder der Lesart): (3) und (2) werden rot | andere |
| U11 | (Nr. 13 a, b) Weder das Token noch der Kandidat erreichen den Body oder einen Header einer Antwort, auch nicht als Teilstück ab 8 Bytes | Schleife über die 20 Formen von U1 mit dem Canary als Kandidat, den Fall `Vergleich` falsch und die Erfolgsantwort: Header (Namen und Werte) und Body der Antwort enthalten weder das Token noch den Kandidat noch ein Teilstück von 8 Bytes Länge an jeder Stelle des Wertes; Zählprüfung 22 | 401 mit `WWW-Authenticate: Bearer realm=<Kandidat>`; Kandidat im Body | andere |
| E2 | (Nr. 5) `WWW-Authenticate` nur bei 401, `Allow` nur bei 405, `Connection: close` nur bei 503 | Antwortarten dieses Teils (401, 404, 405, 200, healthz, 500) × die drei Header: 18 Zellen, Orakel Literale; die Zellen 503, 400 und Fehler der Route mit Code folgen mit ihren Teilen (27 am Ende) | `Connection: close` an jeder Fehlerantwort; `WWW-Authenticate` an jeder Fehlerantwort | andere |
| E3 | (Nr. 5) Eine Erfolgsantwort trägt den Status der Route; ein nichtleerer `Body` mit `Content-Type: application/json` und `Content-Length`, ein leerer ohne Body | Fälle 200 mit Body, 201 mit Body, 204 ohne Body, 200 mit leerem Body, 200 mit `nil` als Body, `GET /healthz` (Body `{"status":"ok"}`, `Content-Length` 15); Orakel Header als Literale | `Content-Type` fehlt; Body bei 204 geschrieben; `Content-Length` fehlt | andere |
| E4 | (Nr. 5; `LH-FA-13.a`; `SPEC-114` Nr. 6) Jede Antwort trägt `X-Correlation-Id` gleich der ID des Requests | Schleife über *R* (26): der Header ist die Kennung, die `zaehlendeIDs` für diesen Request vergab; zwei Requests tragen verschiedene | Header fehlt bei Erfolg, healthz, 404, 405 oder der Route mit Fehler | andere |

Zeilen gesamt 22: *a* = 0, *b* = 22. In A1 und E2 nennen die Zellen die Formen dieses Teils (siehe *Wachsende Mengen*), `slice-admin-lebenszyklus` und `slice-admin-fehlerabbildung` ergänzen den Rest. Die Zeilen K5, U10 und U11 sind gegenüber dem Planner-Stand neu (Architect, 2026-11-08): K5 und U11 tragen Sätze des Vertrags, die kein Fänger trug (`Vergleich` nil, Nr. 13 a und b); U10 hält die Annahme über `net/http`, auf der Nr. 4 b und die Form 20 ruhen.

**Akzeptierte Negative dieses Teils** (Grund und Folge, `AGENTS.md` §3.15):

- N-1 Kodierte Schreibweise von `/healthz`: ob `/%68ealthz` frei ist (Lesart `URL.Path`) oder die Authentisierung durchläuft (Lesart `EscapedPath`), entscheidet `SPEC-238` Nr. 3 c nicht, und kein Fänger legt es fest. Grund: der Body der Route ist ein Literal ohne Geheimnis, beide Lesarten sind zulässig; die Zuordnung (Nr. 2) liest den dekodierten Pfad, die natürliche Umsetzung ist dieselbe Lesart. Folge: eine Änderung der Lesart bleibt grün; die Tabelle A4 führt nur die Formen, in denen die Lesarten übereinstimmen (`/healthz%2F` ist dekodiert `/healthz/` und geht in beiden durch die Authentisierung).
- N-2 Eine Route, die zu Status 204 einen nichtleeren `Body` liefert: `net/http` verwirft ihn, der Adapter prüft es nicht. Grund: keine Route des Plans liefert es, `SPEC-238` Nr. 15 d bindet Daten an den Status 2xx. Folge: ein Endpunkt, der es täte, erhielte ein 204 ohne Body und einen `Content-Length`, den `net/http` selbst entscheidet; der Vertrag des Endpunkts nennt den Status.
- N-3 Namen von Vergleichsfunktionen außerhalb der Liste in U6 (andere Pakete, eigene Hilfsfunktionen) und ein Vergleich des Tokens unter einem Bezeichner, der weder `token` noch `kandidat` heißt: der AST-Test sieht sie nicht. Grund: die Menge ist unbeschränkt (`SPEC-238` *Ungeprüft*: das Review liest die eine Funktion). Folge: ein früh endender Vergleich unter fremdem Namen bleibt grün; Prüfpunkt jedes Reviews, das `admin.go` ändert, ist `vergleiche` und die Stelle, die sie ruft.
- N-4 Aus N-10 des Rahmens bleiben Namen ab 20 Zeichen und die Namen jenseits der 48 Nachbarn ohne Fänger (Grund und Folge dort); die Wertformen (leerer erster Wert, leere Liste) fängt A7 jetzt.

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, Auftrag C; das Orakel trägt Statuscodes, Codes, Header und Meldungstexte als Literale aus dem Vertrag, nie als Konstanten des Codes):

| Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|---|
| M-Header (U1, A3, U4, U8, U11) | 20 ungültige Formen von `Authorization` (Liste in U1) | Liste im Test | 401, Literal `ODR-E2101`, 0 Aufrufe von `Vergleich`, Route, `useCaseFake` | 20 gleich der Zahl der Formen in U1; `SPEC-238` Nr. 4 (a, b) nennt keine Zahl, jede Form ist ein Fall der Sätze (a) oder (b) |
| M-Schema (U2) | Schreibweise 64 × Leerzeichen 3 gleich 192 | Schleife | 200, Kandidat gleich dem Token | 192 |
| M-Kandidat (U3) | 7 Formen (Liste in U3) | Liste im Test | Literale (200 oder 401), Argumente von `Vergleich` | 7 |
| M-vergleiche (U5) | 6 Formen × Längen 32, 33, 64 und 3 Randfälle | Schleife | `string(a) == string(b)` im Test | 21 |
| M-Methode × Pfad (U7, A5) | Methoden 8 × Pfade 5 (U7); die Menge Methoden 8 × Pfade 4 aus `slice-admin-rahmen` (Z6) für A5 | Schleife | Status und `Allow` als Literale | 40 und 32 |
| M-healthz (A4) | Methoden 9 × Pfade 9; Query 3 | Schleife | 200 für `GET /healthz`, sonst 401 ohne Token, 405 mit `Allow: GET` oder 404 mit Token | 81 und 3 |
| M-Anfrage (A7) | Schreibweisen 8192 × Wertformen 8; Nachbarn 48 × Wertformen 8 | Schleife | kein Schlüssel der Schreibweisen-Menge, jeder Nachbar mit gleicher Liste | 65536 und 384 |

**Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** die Länge des Wertes von `Authorization` (Prüfgrenze 32, 33 und 8192 Bytes; Grund: ab 33 Bytes entscheidet allein die Gleichheit; Folge: ein Wert nahe 64 KiB wird von `net/http` mit 431 abgewiesen, nicht vom Adapter). U2 und U3 führen Tokens der Längen 32 und 33; ein Fall (200) führt ein Token der Länge 8192 und seinen Kandidaten gleicher Länge.

**Zweige als Menge, Runden unter dem Detektor** (Architect, Auftrag D; Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat null Blöcke mit Zähler 0, der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| Ablauf: `GET /healthz`, Authentisierung (fehlt, mehrere, Schema, leer, `Vergleich` falsch, `Vergleich` wahr), 404, 405, Route | A1, A3 bis A6, U1 bis U4 | – |
| `vergleiche`, `Vergleich` nil | U5, U6, K5 | – |
| Erfolg leer und nichtleer | E3 | – |
| Filter von `anfrage` | A7 | – |

**Aufrufstellen und zählende Fakes** (Architect, Auftrag E; „jeder Fall je Aufrufstelle“ gilt für diese Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| genau eine, vor allem anderen | `IDs.KorrelationsID` | A1, E4 (Schleife über *R*) | `zaehlendeIDs`: zählt alle fünf Operationen, hält die vergebene Kennung |
| genau eine, nach der Bildung des Kandidaten | `Vergleich` | U4, K5 (nil) | zählende Funktion des Tests: hält beide Argumente, Rückgabe einstellbar |
| genau eine je Route | `Handler` | A6, U8, A7 (b) | `routeFake`: zählt, hält `Anfrage` und `ctx`; der `useCaseFake` dahinter zählt jede Operation (U8) |
| keine in diesem Teil | `Audit.Append`, `Telemetrie.CountError` | A3 (0 für healthz) | `FakeAudit`, `FakeTelemetry` zählen jede Operation, auch die, die der Teil nie rufen darf |

**Zeitkonstanter Vergleich** (Architect, Auftrag H): Haltepunkt am Argument: `Konfig.Vergleich` (U4) belegt, dass der Adapter die Funktion ruft und allein ihr Ergebnis gilt; die Quelle der Funktion belegt der AST-Test U6 (`subtle.ConstantTimeCompare` auf SHA-256, keine verbotenen Vergleiche; Grenze N-3); das Zeitverhalten selbst misst kein Test (akzeptiertes Negativ in `SPEC-238` *Ungeprüft*, Grund und Folge dort).

**Größe der Sicht des Architect** (`SPEC-128` Nr. 1, Aufteilung durch den Planner): 22 Zeilen (*a* = 0, *b* = 22; der Planner-Stand vom 2026-11-07 zählte 18), *f* = 0 (die Fakes liegen seit `slice-admin-rahmen`), *p* = 100. Der Planner setzt danach *S* = 80·22 + 100 = 1860 und *M* = 2600 in den Block von §4 (§6, Größe).

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18). Für diesen Teil gilt das für das Token mit BOM (U3 Form 7) und jedes Token mit Bytes jenseits von ASCII.
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.
- Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (`BEO-REPO/report-der-rolle-faellt-docs-check`); der Planner reicht den Satz an beide weiter.

**Stand der Umsetzung** (Implementer, 2026-11-08; Quelle `src/adapters/driving/admin/admin.go`, Tests in den Dateien `auth_helfer_test.go`, `authentisierung_test.go`, `anfrage_filter_test.go`, `quelle_test.go`, `leitung_test.go` und `export_test.go` desselben Verzeichnisses; Bestandstest `anfrage_test.go`: Z8 trägt die gültige kanonische Zeile, der Canary steht in einem Schlüssel anderer Schreibweise). Mutationen fuhr der Implementer mit `tools/harness/mutationslauf.sh` (Klasse `andere`, Image `odata-recorder:test`, `--memory=4g`, `--cpus=4`, unter `flock`; ein Prozess je Mutation und Stufe, Last des Hosts 5 bis 9 bei 4 CPUs durch fremde Läufe, Messender: das Werkzeug). Die Klasse `andere` gilt für diesen Teil, weil er keine Zusage über Nebenläufigkeit trägt (Nr. 12 und Zeile N1 gehören zu `slice-admin-abweisen-audit`); eine Rate ist deshalb keine zugesagt, jede Mutation ist deterministisch (rot in 1 von 1 Prozessen).

Ergebnis: 109 Mutationen, 108 rot (davon `A7-serve-nur-einzeilig` allein im Fänger (a) `TestAnfrageFilterFunktion`, wie die Zeile A7 es sagt), eine äquivalent (`E2-www-authenticate-bearer-an-jedem-fehler`: der Wert ist hinter der Bedingung `www != ""` immer `Bearer`, das Verhalten bleibt gleich). Der erste Lauf (99 Mutationen) ließ `U1-schema-nicht-geprueft` grün; Ursache: die Formen 8 und 9 von U1 scheitern an der Prüfung des Leerzeichens an Stelle 6, nicht am Schema. Der Fix deckt die Klasse (U1, `TestSchemaJedesByte`), nicht die eine Mutation: 6 Stellen × 254 Bytes + 256 Bytes hinter dem Schema; rot gesehen an sechs Mutationen, die der Befund nicht nannte (`U1-schema-letzte-stelle-ungeprueft`, `U1-schema-erste-stelle-ungeprueft`, `U1-schema-stelle-2-ungeprueft`, `U1-schema-stelle-3-ungeprueft`, `U1-byte-hinter-schema-tab-oder-leerzeichen`, `U1-byte-hinter-schema-whitespace`).

| Zeile | Mutation (Name im Lauf; Wirkung) | roter Test |
|---|---|---|
| K4 | `K4-kopie-entfernt` (`token: k.Token`) | `TestNewKopiertToken` |
| K5 | `K5-nil-true`, `K5-nil-false`, `K5-nil-bleibt-nil` (Aufruf von nil) | `TestVergleichNil` |
| A1 | `A1-id-nach-authentisierung`, `A1-zweiter-aufruf-im-401`, `A1-andere-operation` (`SessionID` zusätzlich) | `TestKorrelationsIDJeRequestart` |
| A3 | `A3-healthz-hinter-authentisierung`, `A3-healthz-liest-authorization`, `A3-healthz-zaehlt` (`CountError`), `A3-healthz-audit` (`Append`) | `TestHealthzMitFormen` |
| A4 | `A4-hasprefix`, `A4-hassuffix`, `A4-contains`, `A4-pfad-ohne-beachtung-der-schreibweise`, `A4-query-zaehlt`, `A4-methode-nicht-geprueft`, `A4-methode-ohne-beachtung-der-schreibweise` | `TestHealthzMethodeUndPfad` |
| A5 | `A5-zuordnung-vor-authentisierung`, `A5-404-vor-authentisierung`, `A5-405-vor-authentisierung` | `TestAuthentisierungVorZuordnung` |
| A6 | `A6-wiederholung-nach-fehler`, `A6-handler-zweimal` | `TestHandlerGenauEinmal` |
| A7 | `A7-nur-einzeilig`, `A7-nur-nichtleer`, `A7-laenge-2-nicht-gefiltert`, `A7-laenge-3-nicht-gefiltert`, `A7-laenge-0-nicht-gefiltert`, `A7-verwirft-leere-werte`, `A7-nur-kanonischer-schluessel`, `A7-verwirft-praefix`, `A7-verwirft-suffix` | `TestAnfrageFilterFunktion`; ebenso `A7-serve-verwirft-leere-werte` in `TestAnfrageFilterServeHTTP`, `A7-z8-nur-kanonisch` in `TestHeaderOhneAuthorization` |
| U1 | `U1-genau-eine-zeile-ge1`, `…-ne2`, `…-eq0`, `U1-schema-nicht-geprueft` (rot erst in `TestSchemaJedesByte`), `U1-laenge-nicht-geprueft` (Panik bei `Bea`), `U1-schluessel-ohne-beachtung-der-schreibweise`, `U1-leerer-kandidat`, `U1-tab-zaehlt`, `U1-leerzeichen-nicht-verlangt` und die sechs weiteren Mutationen am Schema und am Byte hinter dem Schema (siehe oben) | `TestFormenUngueltigerAuthorization`, `TestSchemaJedesByte` |
| U2 | `U2-schema-exakt`, `U2-ein-leerzeichen`, `U2-nur-drei-schreibweisen` | `TestSchemaUndLeerzeichen` |
| U3 | `U3-kandidat-getrimmt`, `U3-kandidat-klein`, `U3-rest-hinter-tab-abgeschnitten`, `U3-innerer-leerraum-entfernt`, `U3-bytes-jenseits-ascii-verworfen`, `U3-letztes-byte-ignoriert`; `U3-kappung-bei-4096` (Token 8192); `U3-nul-byte-schneidet-ab`, `U3-bytes-ab-0x80-verworfen`, `U3-steuerzeichen-verworfen`, `U3-fuehrendes-nul-gestrichen` | `TestKandidatByteFuerByte`, `TestTokenLaengen`, `TestKandidatAlleBytes` |
| U4 | `U4-aufruf-durch-gleichheit`, `U4-argumente-vertauscht`, `U4-und-eigener-vergleich`, `U4-oder-eigener-vergleich`, `U4-grenze-verschoben`, `U4-zweimal-gerufen`, `U4-ergebnis-ignoriert` | `TestVergleichEntscheidetAllein` |
| U5 | `U5-invertiert`, `U5-nur-laenge`, `U5-nur-erstes-byte` | `TestVergleiche` |
| U6 | `U6-slices-equal`, `U6-strings-compare`, `U6-sum256-entfernt`, `U6-token-neben-dem-aufruf`, `U6-token-ungleich-neben-dem-aufruf` | `TestVergleichNurUeberConstantTimeCompare` (der Zähler selbst an Proben: `TestQuelleProbenAmZaehler`) |
| U7 | `U7-ausnahme-head`, `U7-ausnahme-options` | `TestAuthentisierungJedeMethodeJederPfad` |
| U8 | `U8-route-vor-der-pruefung`, `U8-route-vor-der-pruefung-falsches-token` | `TestFormenUngueltigerAuthorization`, `TestFalschesTokenOhneZustand` |
| U9 | `U9-www-authenticate-fehlt`, `U9-www-authenticate-mit-realm`, `U9-code`, `U9-status` | `TestUnauthorized` |
| U10 | `U10-schluessel-nur-klein`, `U10-erste-zeile-genuegt` | `TestLeitungAnnahmenZuNetHTTP` |
| U11 | `U11-kandidat-im-realm`, `U11-token-im-401-header`, `U11-authorization-im-erfolgsheader` | `TestKeinGeheimnisInDerAntwort` |
| E2 | `E2-connection-close-an-jedem-fehler`, `E2-www-authenticate-an-jedem-fehler`, `E2-allow-an-jedem-fehler` | `TestHeaderJeAntwortart` |
| E3 | `E3-content-type-fehlt`, `E3-content-length-fehlt`, `E3-header-auch-bei-leerem-body`, `E3-status-fest`, `E3-healthz-body` | `TestErfolgsantwort` |
| E4 | `E4-header-fehlt`, `E4-header-fest`, `E4-header-fehlt-bei-healthz`, `…-bei-erfolg`, `…-bei-fehler-der-route`, `…-bei-404`, `…-bei-405`, `…-bei-401` | `TestKorrelationsHeader` |

Mengen (Generator · Orakel · Zahl, die der Test zählt): *R* 26 (Liste; Literale des Vertrags) · M-Header 20 (Liste; `401` und `ODR-E2101`) · M-Schema 64 × 3 = 192 (Schleife über die Buchstabenmaske; `200` und Kandidat gleich dem Token) · M-Kandidat 7 Formen in 8 Fällen (Liste; Literale und die Argumente von `Vergleich`) · M-vergleiche 21, beide Reihenfolgen (`string(a) == string(b)` im Test) · M-Methode × Pfad 8 × 5 = 40 (U7) und 32 Paare in je drei Modi (A5) · M-healthz 81 in zwei Modi (ohne und mit Token) und 3 Queries · M-Anfrage 65536 und 384 über `anfrage`, dazu 8192 und 96 über `ServeHTTP` · Schema-Bytes 1524 und 256 · Kandidat-Bytes 256 · Token-Längen 32, 33, 8192 mit je vier Kandidaten · U11 22 · E2 18 · E3 6 · A6 fünf Methoden in zwei Modi. Zweige: das Profil von `go test -race -coverprofile` (`odata-recorder:race`) über das Paket hat drei Blöcke mit Zähler 0, keinen in `admin.go`: `src/adapters/driving/admin/route.go` Zeile 128 (Fehler der Dekodierung eines Segments, das `EscapedPath` nie liefert, `slice-admin-rahmen`) und `src/adapters/driving/admin/tokendatei.go` Zeilen 108 und 132 (`slice-admin-listener-token`).

Gegengelesen auf Rohbytes (`LC_ALL=C grep -P`, `AGENTS.md` §3.18): alle Dateien des Verzeichnisses `src/adapters/driving/admin` auf die Folgen Backslash-u-vier-Hex-Ziffern, BOM (EF BB BF) und NBSP (C2 A0): kein Treffer; das Token mit BOM entsteht in `bom()` aus Bytewerten (`src/adapters/driving/admin/auth_helfer_test.go`), die Dateien tragen jenseits von ASCII nur Umlaute in Kommentaren und Meldungen.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-admin-rahmen](../done/slice-admin-rahmen.md) liegt in `done/` (Code-Abhängigkeit: `Konfig`, Zuordnung, die eine Stelle für das Fehlerobjekt, die Fakes `zaehlendeIDs` und `routeFake`). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-238` Nr. 3 bis 5) seit dem Zweig von `slice-admin-listener-token` (`AGENTS.md` §3.12). Stand der Bedingungen beim Schnitt (Planner, 2026-11-07, `main` = `aadbef0`): `slice-audit-log` und `slice-telemetry-port` liegen in `done/`; der Teil ruft beide Ports nicht.

**Größenmarke** (Planner, 2026-11-08, `SPEC-128` Nr. 1; vollständiger Block unten, vor dem ersten Code-Commit): 22 Zeilen der Tabelle (Architect, 2026-11-08), alle Klasse `andere`: *a* = 0, *b* = 22, *f* = 0, *p* = 100: *S* = 85·0 + 80·22 + 200·0 + 100 = 1860, *M* = 2600, keine Schnittfrage (*S* < 2223), es wird vor dem Code nicht geschnitten. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfad des Blocks: `src/adapters/driving/admin/`; `produktcode` 100, `fakes` 0. Die Schätzung folgt allein der Formel, ohne Eigenkalibrierung. Stand der Bedingungen beim Start (`main` = `c469ef1e`): `slice-admin-rahmen` liegt in `done/`; `slice-audit-log` und `slice-telemetry-port` liegen in `done/`; in `in-progress/` liegen sonst nur `slice-handbuch-katalog-schluessel-kommandos` (Linie der Gate-Kette) und der parallel startende `slice-bootstrap-m3`; dieser Slice ändert keinen Pfad der Gate-Kette (Übergangsregel 7, §6), also kein Konflikt. Basis der Messung: `c469ef1e` (die Start-Commits ändern allein Pläne; Messung vor dem Code 0).

```groessenmarke
pfad: src/adapters/driving/admin/
produktcode: 100
fakes: 0
basis: c469ef1e
schaetzung: 1860
marke: 2600
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 2600 (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein Teil, der sich ohne geteilten Mechanismus abtrennen lässt, steht dann vor dem Code in §4, sonst steht die Messung mit Grund im Plan (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert — Carveout?): der Architect ändert `SPEC-238` Nr. 3 oder 4 (Rückgabe an den Planner), oder `slice-admin-rahmen` liefert `Konfig` oder die Zuordnung nicht in der Form, die der Ablauf braucht. Die Schärfung vom 2026-11-08 ändert die Spezifikation nicht: jede Randform ist aus Nr. 1 bis 5 und 13 oder durch ein akzeptiertes Negativ (N-1 bis N-4) entschieden.

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07, geschärft 2026-11-08, Zeilen dieses Teils): jede Zelle nennt eine Nummer von `SPEC-238` (Rahmen) oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Authentisierung (Header `Authorization`) | `SPEC-238` Nr. 4 a: kein Header und ein Schlüssel ohne Wert sind ein Verstoß, 401 | Nr. 4 b: leerer Wert, nur Leerraum, `Bearer` ohne Kandidat, `Bearer` mit Leerzeichen und leerem Kandidat; `Vergleich` wird nicht gerufen (Nr. 4 c) | Nr. 4 b und c: anderes Schema, mehrere Zeilen (auch gleichlautende, auch drei), Kandidat ungleich dem Token (`Vergleich` falsch), Kandidat mit Bytes jenseits von ASCII (Byte für Byte); Schlüssel in nichtkanonischer Schreibweise: gilt nicht für den Betrieb (`net/http` kanonisiert den Schlüssel, U10), im direkten Aufruf von `ServeHTTP` ist der Header nicht da, 401 nach Nr. 4 a (U1 Form 20) | Nr. 8: `Append` hat keinen `ctx` und läuft zu Ende; der Client kann trennen, der Eintrag steht (`SPEC-234` Nr. 5 c); der Eintrag gehört zu `slice-admin-abweisen-audit` | Nr. 4 und Nr. 12: jeder Request für sich, kein Zustand; Nr. 8: jede Abweisung ein Eintrag (ebenfalls dort) |
| `GET /healthz` | Nr. 3 c: kein `Authorization` nötig, 200 | Nr. 3 c: leere Query und leerer Header gleichgültig, 200 | Nr. 3 c und d: falsches Token gleichgültig; andere Methode (auch `get`), `/healthz/`, `/healthz/x`, `/healthzx`, andere Schreibweise der Buchstaben gehen durch die Authentisierung; kodierte Schreibweise der Buchstaben (`/%68ealthz`): gilt nicht als Vertrag, N-1 (§3) | Nr. 3 b und Nr. 9: angehalten ist 503, ein trennender Client lässt den Request enden (beides `slice-admin-lebenszyklus`) | Nr. 3 c und Nr. 12: zustandslos, je Request eine ID, nichts gezählt, nichts auditiert |
| Bildung der `Anfrage` (kein `Authorization` in `Header`) | Nr. 2: ein Request ohne `Authorization` ist nicht möglich (401), die Funktion liefert `Header` ohne Schlüssel | Nr. 2: der Schlüssel fehlt in `Header` unabhängig vom Wert, auch bei leerem Wert und leerer Liste (A7); Nachbarn mit leerem Wert und leerer Liste bleiben | Nr. 2: jede Schreibweise des Schlüssels, jede Zahl von Zeilen; Nachbarnamen (`Proxy-Authorization`, `X-Authorization`, `Authorization-Info`) bleiben mit allen Werten; Namen ab 20 Zeichen: N-4 | gilt nicht: die Funktion hat keinen `ctx`, sie liest den Request nur | Nr. 2 und Nr. 12: `Header` ist eine Kopie, zwei Requests teilen nichts (Z8 des Rahmens) |
| Erfolgsantwort der Route (Header, Status, Body) | Nr. 5: `Body` nil oder leer, dann wird kein Body geschrieben und `Content-Type` und `Content-Length` bleiben aus | Nr. 5: dasselbe für `Body` mit Länge 0 | Status außerhalb von 200 bis 299: Nr. 6 d (`slice-admin-fehlerabbildung`); Status 204 mit nichtleerem Body: gilt nicht als Vertrag, N-2 (§3) | Nr. 9: Kontextfehler ohne Antwort (`slice-admin-fehlerabbildung`) | Nr. 12: je Request eine eigene Antwort und eine eigene ID (E4) |

**Zusatztabelle: Paare von Verstößen am selben Schlüssel und Modi** (`AGENTS.md` §3.16; die Fänger sind Zeilen der Tabelle in §3):

| Schlüssel und Paar oder Modus | Ergebnis (Ort) | Fänger |
|---|---|---|
| Header: fehlender Header und unbekannter Pfad | 401, nicht 404 (`SPEC-238` Nr. 3 d, e) | A5 |
| Header: falsches Token und nicht vorgesehene Methode | 401, nicht 405 (Nr. 3 d, e) | A5 |
| Header: zwei Zeilen und falsches Schema | 401 (Nr. 4 a); der eine Eintrag gehört zu `slice-admin-abweisen-audit` (D1) | U1 (Formen 15, 16) |
| Header: zwei Zeilen und leerer Wert in der ersten | 401, `Vergleich` nicht gerufen (Nr. 4 a, c) | U1 (Form 17), U4 |
| Header: Schema `Bearer` und Kandidat, der selbst `Bearer <token>` lautet | `Vergleich` genau einmal mit dem ganzen Rest, 401 (Nr. 4 b, c) | U3 (Form 5) |
| Header: leerer Kandidat und Kandidat der Länge 1 | der leere: 401 ohne Aufruf, der der Länge 1: ein Aufruf (Nr. 4 b, c) | U4 |
| Modus `GET /healthz` ohne Token, mit falschem, mit richtigem Token, mit jeder Form von M-Header | 200 in allen (Nr. 3 c) | A3 |
| Modus `GET /healthz?x=1` | 200 (Nr. 3 c) | A4 |
| Modus `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`, `TRACE`, `get` auf `/healthz` ohne Token | 401 (Nr. 3 c, d) | A4, U7 |
| Modus dieselben Methoden auf `/healthz` mit Token | 405 mit `Allow: GET` (Nr. 3 d, e, Nr. 5) | A4, Z6 (`slice-admin-rahmen`) |
| Modus jede der fünf Methoden einer Route mit Token | der `Handler` der Route, genau einmal (Nr. 3 f) | A6 |
| Modus `Vergleich` nil | `vergleiche` entscheidet (Nr. 1, 4 e) | K5 |
| Modus Route liefert Erfolg mit `Authorization` im Request | die `Anfrage` trägt keine Zeile davon (Nr. 2) | A7 (b) |

Die Paare „ungültige Form und angehaltener Server“ und „`GET /healthz` bei angehaltenem Server“ gehören zu `slice-admin-lebenszyklus` (A2).

- „Der Vergleich erfolgt zeitkonstant“ (`LH-FA-13.a`) hat keinen Fänger, der aus dem benannten Grund rot wird — **Ausgang:** entschieden (Architect, 2026-11-07, `SPEC-238` Nr. 4 e): `Vergleich` ist ein Feld von `Konfig` (Haltepunkt am Argument, Fänger U4), die Funktion `vergleiche` ruft `subtle.ConstantTimeCompare` auf SHA-256 beider Werte, die Quelle belegt der AST-Test U6; das Zeitverhalten selbst misst kein Test (akzeptiertes Negativ mit Grund und Folge in `SPEC-238` *Ungeprüft*; die Grenze der Namensliste von U6 ist N-3).
- Das 401 schreibt in diesem Teil keinen Audit-Eintrag und zählt nicht, obwohl `LH-FA-13.a` einen Eintrag ohne Tokenwert verlangt und `SPEC-238` Nr. 7 das Zählen. Grund: der Eintrag ist Zusage von `slice-admin-abweisen-audit` (Art `admin_abgewiesen` in `support.Arten()`, Zeilen D1 bis D7), das Zählen aller Fehler (401, 404, 405, 500) die von `slice-admin-fehlerabbildung`. Folge: bis diese Teile in `done/` liegen, ist die Anforderung für das Abweisen nicht belegt; der Listener ist nirgends verdrahtet (kein Composition Root vor `slice-bootstrap-m6`), also entsteht keine erreichbare Abweisung ohne Eintrag — **Ausgang:** entschieden (Planner, Schnitt): Fänger in `slice-admin-abweisen-audit` und `slice-admin-fehlerabbildung`.
- `LH-FA-13.a` nennt keine Begrenzung von Fehlversuchen; ein vom Betreiber gesetztes Token von nur 32 Zeichen wäre durch Raten angreifbar — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User (Owner); kein Code, bis entschieden.
- Die Schleifen über *R* und die Antwortarten (A1, E2, E4) wachsen mit den Teilen — **Ausgang:** entschieden (Planner, Schnitt): jeder Teil trägt seine Formen ein (§3, *Wachsende Mengen*); die volle Zählprüfung (27 Zellen in E2) gilt mit `slice-admin-fehlerabbildung`.
- Der Filter der `Anfrage` (A7, Sendung des Rahmens) ist über `ServeHTTP` nicht mehr beobachtbar für Requests mit mehreren Zeilen, denn sie enden vor der Route im 401 — **Ausgang:** entschieden (Architect, 2026-11-08): A7 (a) prüft die Funktion `anfrage` direkt über `export_test.go`, A7 (b) die Schreibweisen des Schlüssels neben einer gültigen kanonischen Zeile; die Authentisierung liest allein den kanonischen Schlüssel (`Header.Values("Authorization")`), die Formen 13 bis 18 und 20 von U1 halten die Lesart fest, U10 die Annahme über `net/http`. Keine Änderung der Spezifikation: Nr. 2 sagt „ohne `Authorization`“, Nr. 4 a „mehrere Zeilen“.
- Randformen ohne Ort in `SPEC-238` (kodierte Schreibweise von `/healthz`, Status 204 mit Body, Namen von Vergleichsfunktionen) — **Ausgang:** entschieden (Architect, 2026-11-08): akzeptierte Negative N-1 bis N-3 in §3 mit Grund und Folge; kein Code entscheidet sie ohne diese Zeilen.
- Der Fund zur Token-Datei (`BEO-REPO/token-datei-wettlauf-fehlt-gegen-symlink-ohne-ziel`, Wettlauf zwischen `Stat` und `Lstat` in `LadeToken`) betrifft diesen Teil nicht: er ruft `LadeToken` nie und liest keine Datei (`SPEC-238` Einleitung) — **Ausgang:** entfallen für diesen Slice; die Adresse (Architect, danach ein Fix-Slice des Planners für `tokendatei.go`) bleibt im Register.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (12×): jeder Zweig der Authentisierung hat einen Lauf unter dem Detektor (Z-Liste in §3) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): die Fakes `routeFake` und `useCaseFake` zählen jede Operation, auch die, die der Rahmen nie rufen darf — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/akzeptiertes-negativ-ohne-folge` (15×): das Zeitverhalten des Vergleichs ist ein Negativ mit Grund und Folge in `SPEC-238` *Ungeprüft*, der Haltepunkt ist U4; N-1 bis N-4 tragen Grund und Folge — **Ausgang:** weiter offen → `BEO-REPO/akzeptiertes-negativ-ohne-folge` im Register.
- Größe: der Block in §4 trägt für die 22 Zeilen der Tabelle *S* = 1860, *M* = 2600, Basis `c469ef1e`, unter der Schnittgrenze 2223 — **Ausgang:** gesetzt (Planner, 2026-11-08, vor dem ersten Code-Commit); `make groessenmarke` grün.
- Gate-Kette (Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch`): der Teil ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).
- Abweichung vom Plan (Implementer, 2026-11-08): die Zeile U1 nannte die Formen 8 (`Basic `) und 9 (`Token `) als Fänger zu „Schema nicht geprüft“; beide scheitern an der Prüfung des Leerzeichens an Stelle 6, die Mutation blieb im ersten Mutationslauf grün. Der Fänger ist die Klasse „anderes Schema“ als Generator (`TestSchemaJedesByte`, 1524 + 256 Fälle), die Zeilen U1 und U3 in §3 sind nachgezogen; keine Zelle des Rasters ohne Ort, keine Randform im Code entschieden — **Ausgang:** entschieden (Implementer, Fänger ergänzt); Adresse: Reviewer, der die Klasse mit einer Mutation an einem anderen Element auflöst.
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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Schnitt nachgezogen (Planner, 2026-11-07, `main` = `aadbef0`, Zähler aus den Dateien unter `evidence/`): einschlägig sind `zweig-ohne-lauf-unter-dem-detektor` 12×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `akzeptiertes-negativ-ohne-folge` 15×, `betriebsregel-fuer-repo-dateien-verletzt` 9×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 44×, `randform-erst-im-review-entschieden` 34×, `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3×, `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11×, `schaetzformel-ueberschaetzt-den-slice` 21×, `report-der-rolle-faellt-docs-check` 5× (Verkörperung, soweit die Schwelle erreicht ist, beim Lese-Schritt der Welle-Closure); kein neues Verzeichnis, kein Zähler gesetzt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
