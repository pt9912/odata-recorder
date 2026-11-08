# Slice slice-fallback-policy: Fallback-Policy als reine Entscheidung

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` (Entscheidungsmatrix) · `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080` · `SPEC-020`, `SPEC-021`, `SPEC-022`, `SPEC-023`, `SPEC-024` (die Codes, nach denen die Matrix entscheidet) · `SPEC-110` Nr. 2 (die Schlüssel, die diese Version noch ablehnt) · `ARC-002` · `architecture.md` §2.5 (Modus-Routing) · `SPEC-261` (Vertrag der Policy, Architect).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-fallback-policy`, Worktree `/Development/odata-recorder-wt-fallback-policy`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Eine reine Entscheidung (Paket unter `src/hexagon/services/hybrid`, der Ort ist Sache des Architect) legt aus Methode, Zielhost, Pfad, dem `ODR`-Code des Ergebnisses und den typisierten Werten der Policy (Hauptschalter, Allowlist der Hosts, Routenfreigaben für mutierende Methoden, Kategorien des Replay-Fallbacks) fest, ob die Antwort bleibt, ein Replay-Fallback oder ein Upstream-Fallback erlaubt ist; ohne ausdrückliche Freigabe lautet jede Entscheidung „kein Fallback“, und sie hängt am `ODR`-Code, nie am HTTP-Status (`LH-FA-12.b`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Aufruf von `MockUseCase`, `ReplayUseCase` und `UpstreamClient`, Strom und Zustand — `slice-hybrid-service`; die Policy ist eine Funktion ohne Port, damit die Matrix ohne Fakes vollständig prüfbar ist und kein Zweig von einem Fake abhängt.
- Das Lesen der Schlüssel `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` aus der Umgebung und das Aufheben ihrer Ablehnung in `SPEC-110` Nr. 2 — `slice-bootstrap-m4`; solange kein Composition Root die Policy verdrahtet, bliebe ein gelesener Schlüssel eine still wirkungslose Sicherheitseinstellung. Die Policy nimmt typisierte Werte, nie die Umgebung; Syntax und Prüfung eines Eintrags der Routenfreigabe (Methode, Pfadmuster) gehören dagegen zu ihr, `slice-bootstrap-m4` ruft sie beim Lesen der Liste.
- Das Senden an den Upstream und das Speichern einer Fallback-Antwort — `slice-hybrid-service` (Senden); Speichern und Datenabgleich sind Out-of-Scope von `LH-FA-12`.
- Spans, Audit und Metriken der Entscheidung — `welle-betrieb`; es gibt noch keinen Port `Telemetry`.

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

- [ ] Entscheidungsmatrix `LH-FA-12.b`: je Zeile (Erfolg, fachliche `404`/`409`/`412`, Coverage unvollständig, nicht unterstützte Operation, ungültige Anfrage, Replay-Miss) und je Zustand der Policy das Ergebnis; Auslöser ist der `ODR`-Code, nie der HTTP-Status; unter den Replay-Fehlern löst nur `ODR-E5001` einen Fallback aus.
- [ ] Upstream-Freigabe: nur bei Hauptschalter `enabled`, nur zu einem Host aus `FALLBACK_ALLOW_HOSTS`, der der Host der konfigurierten `UPSTREAM_URL` ist, lesend nur mit GET und HEAD, mutierend nur mit passendem Eintrag aus `FALLBACK_WRITE_ALLOW` (Methode und Pfadmuster); sonst nichts.
- [ ] Replay-Fallback: nur für die Kategorien `unsupported` und `coverage_insufficient` aus `HYBRID_REPLAY_FALLBACK`, unabhängig vom Hauptschalter; Syntax des Eintrags der Routenfreigabe und Vergleich von Host und Pfad nach dem Vertrag des Architect.
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
| `src/hexagon/services/hybrid/` | neu | Policy als reine Funktion über `model`-Typen und typisierte Werte (Ort und Namen: Architect) |
| `spec/spezifikation.md` | update (Architect) | `SPEC-261` (Vertrag der Policy samt Randformen) neu; Nachzug an jeder Stelle, die dieselbe Aussage trägt (`AGENTS.md` §3.9): `LH-FA-12.a`, `LH-FA-12.b`, `SPEC-062`, `SPEC-079`, `SPEC-080`; `SPEC-010` und `SPEC-110` gelten unverändert |
| `spec/architecture.md` | update (Architect) | Verzeichnisbaum um `hybrid/policy/`; §2.5 nennt die Policy als reine Entscheidung ohne Port |
| Tests | neu | Matrix `LH-FA-12.b` als Tabelle × Zustände der Policy; Generator über Methoden, Hosts und Muster mit unabhängigem Orakel; die Policy importiert weder Port noch Adapter (`make a-check`) |

Der Vertrag steht in `SPEC-261` (Architect, 2026-11-08). Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15; die Spalte *Nr.* trägt die Kennung der Zeile und den Ort im Vertrag) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Mengen: *C* = `model.AlleCodes()` (*N* Codes) plus der Nullwert, −1 und *N* + 1 (*N* + 3 Werte); *M* = 13 Methoden (die sieben bekannten, `OPTIONS`, `TRACE`, `CONNECT`, `get`, die leere und `PROPFIND`); *Z* = 112 Zustände der Policy (`Upstream` aus: `Hosts` {leer, passend, fremd, passend mit falschem Port} × `Schreibfreigaben` {keine, passend, andere Methode, anderes Muster} × `ReplayKategorien` {keine, `unsupported`, `coverage_insufficient`, beide} = 64; `Upstream` an: `Hosts` ohne die leere = 48).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| P01 (Nr. 1 a bis c) | Listen sind Mengen (Reihenfolge und Doppelte ohne Wirkung); `New` hält keine Referenz auf die Slices; die Policy ist nach `New` unveränderlich | Menge 9: Liste {`Hosts`, `Schreibfreigaben`, `ReplayKategorien`} × Eingriff {Reihenfolge umgekehrt, jeder Eintrag verdoppelt, Slice nach `New` überschrieben und gekürzt}; gleiche Entscheidungen auf 3 Codes × 5 Methoden wie ohne Eingriff, kein Fehler von `New` | `New` speichert die Slice ohne Kopie (je Liste rot); die Entscheidung hängt an der Position (nur der erste Eintrag gilt) | andere |
| P02 (Nr. 2 a bis c) | Methodenklassen: `GET`, `HEAD` lesend; `POST`, `PUT`, `PATCH`, `DELETE`, `MERGE` mutierend; jede andere Zeichenkette nie freigegeben, weder Replay noch Upstream | Menge 52: *M* (13) × Code {`ODR-E6001`, `ODR-E6002`, `ODR-E6003`, `ODR-E5001`} im Zustand voll offen (`Upstream` an, Host passend, je mutierender Methode ein Eintrag mit Muster `/svc/Orders`, beide Kategorien); die 6 Nie-Methoden ergeben in allen 24 Fällen `Bleiben`, die Soll-Werte der übrigen kommen aus dem Orakel von P06; Zählprüfung 13 × 4 | `OPTIONS` als lesend; `MERGE` nicht mutierend; `get` per `ToUpper` gleich `GET`; die leere Methode als lesend (je Element rot) | andere |
| P03 (Nr. 3 a bis d) | Auslöser: genau `ODR-E6001`, `ODR-E6002`, `ODR-E6003` (Mock) und `ODR-E5001` (Replay-Miss, nur Upstream); jeder andere Code, der Nullwert und ein Wert außerhalb der Tabelle lösen nichts aus, auch `ODR-E6006` | Schleife über *C* mit `GET` im Zustand voll offen: genau 3 Codes ergeben `Replay` (Kategorien beide), 4 Codes mit leeren Kategorien `Upstream`, alle übrigen `Bleiben`; Zählprüfung *C* = *N* + 3 und Auslöser = 4 | Standardzweig der Auswahl liefert `Upstream`; `ODR-E6006`, `ODR-E6005` oder `ODR-E5002` als Auslöser; der Nullwert als Auslöser (je Code rot) | andere |
| P04 (Nr. 3, Kopf) | Der HTTP-Status ist keine Eingabe; `Anfrage` trägt genau Methode, Pfad, Ziel und Code | `TestAnfrageFelder` liest `reflect.TypeOf(Anfrage{})`: die Feldmenge ist {Methode, Pfad, Ziel, Code} mit den Typen `string`, `string`, `string`, `model.Code` | ein Feld `Status int` ergänzt; `Code` durch `Status` ersetzt | andere |
| P05 (Nr. 4 b) | Replay-Fallback: Mock-Auslöser mit gesetzter Kategorie (`ODR-E6001`, `ODR-E6002`: `unsupported`; `ODR-E6003`: `coverage_insufficient`) für lesende wie mutierende Methoden, unabhängig von Hauptschalter, Hosts und Ziel; `ODR-E5001` nie | Menge 336: Kategorien (4) × Code {`ODR-E6001`, `ODR-E6002`, `ODR-E6003`} × 7 bekannte Methoden × `Upstream` {aus, an mit passendem Host} × Ziel {passend, leer}; `Replay` genau bei gesetzter Kategorie des Codes; dazu 28 Fälle `ODR-E5001` × 4 Kategorien × 7 Methoden ohne `Replay` | `unsupported` und `coverage_insufficient` vertauscht; `ODR-E6002` der Kategorie `coverage_insufficient` zugerechnet; der Hauptschalter in der Replay-Regel gelesen; das leere Ziel verhindert Replay; Replay nur für lesende Methoden (je Mutation rot) | andere |
| P06 (Nr. 4 a bis d) | Gesamtmatrix `LH-FA-12.b`: Upstream nur bei Hauptschalter an, Auslöser, freigegebenem Host und (lesend, oder mutierend mit Mock-Auslöser und Routenfreigabe); sonst `Bleiben`; ohne ausdrückliche Freigabe nie ein Fallback | Schleife über *C* × *M* × *Z* = (*N* + 3) × 13 × 112 gegen ein Orakel, das die Zeilen von `LH-FA-12.b` als Wertetabelle trägt (Zeile, Codes, Spalte Replay, Spalte Upstream) und die Regeln des Codes nicht nachbaut; Zählprüfung: Zahl der Fälle und die der Ergebnisse `Replay` und `Upstream` je gleich der aus der Tabelle gezählten | Standardzweig erlaubt; Hauptschalter nicht geprüft; Host nicht geprüft; mutierend wie lesend behandelt; mutierende Freigabe ohne Mock-Auslöser; Oder statt Und zwischen Host und Methode (je Mutation rot) | andere |
| P07 (Nr. 4 c, Nr. 6 c) | Eine mutierende Methode erreicht den Upstream nur mit einer Routenfreigabe, die Methode und Pfad deckt | Menge 135: 5 mutierende Methoden × Freigabe {keine, passend, jede der 4 anderen Methoden mit passendem Muster, passende Methode mit Muster mit anderem Segment, mit zusätzlichem Segment, mit fehlendem Segment} (9) × Code {`ODR-E6001`, `ODR-E6002`, `ODR-E6003`}; `Upstream` genau bei „passend“ | Methode des Eintrags nicht verglichen; Muster nicht verglichen (jede Freigabe genügt); die Freigabe einer anderen Methode gilt (je Mutation rot) | andere |
| P08 (Nr. 4 c, e) | Replay-Miss `ODR-E5001`: Upstream nur für `GET` und `HEAD` zum freigegebenen Host; eine mutierende Methode nie, auch nicht mit passender Routenfreigabe | Menge 7: die 7 bekannten Methoden mit `ODR-E5001`, `Upstream` an, Host passend, für jede mutierende Methode ein passender Eintrag: `GET`, `HEAD` ergeben `Upstream`, die 5 mutierenden `Bleiben` | der mutierende Zweig gilt auch für `ODR-E5001`; `ODR-E5001` wie ein Mock-Auslöser behandelt (je Mutation rot) | andere |
| P09 (Nr. 4 e) | Sind Replay und Upstream offen, gilt der Replay; erst `ODR-E5001` führt zu `Upstream` | Menge 9: Mock-Auslöser (3) × {Kategorie des Codes gesetzt: `Replay`; Kategorie leer: `Upstream`; derselbe Zustand mit `ODR-E5001`: `Upstream`}, `GET`, `Upstream` an, Host passend | Reihenfolge von Replay und Upstream vertauscht | andere |
| P10 (Nr. 4 f) | Ist der Hauptschalter aus, erreicht keine Anfrage den Upstream, weder lesend noch mit Freigabe; `Hosts` und `Schreibfreigaben` bleiben ohne Wirkung | Menge 168: `Upstream` aus × `Hosts` {leer, passend, fremd} × `Schreibfreigaben` {keine, passend} × 7 Methoden × Code {`ODR-E6001`, `ODR-E6002`, `ODR-E6003`, `ODR-E5001`}: nie `Upstream`; Gegenprobe: dieselben Fälle mit `Upstream` an und passendem Host liefern für `GET` und `HEAD` `Upstream` (8 Fälle) | der Hauptschalter nur im lesenden Zweig geprüft; nur im mutierenden Zweig geprüft (je Zweig rot) | andere |
| P11 (Nr. 5 a) | Eintrag von `Hosts`: Name oder IPv6 in Klammern, optional Port ohne führende Null von 1 bis 65535; Ursache `leer` oder `host` | Menge 29: 9 Gutfälle (`h`, `h.example`, `h-1_x`, `h:1`, `h:65535`, `[::1]`, `[::1]:8080`, `[2001:db8::1]`, `H`) und 20 Verstöße (leer mit `leer`; `http://h`, `h/x`, `h/`, `u@h`, `*.h`, `h*`, ` h`, `h x`, `h:0`, `h:65536`, `h:`, `h:+1`, `h:01`, `h:1x`, `[::1`, `[]`, `[1.2.3.4]`, `[fe80::1%25eth0]`, ein Name mit Byte ab 0x80 mit `host`), Schlüssel `FALLBACK_ALLOW_HOSTS` | Port `0`, `65536`, `01` oder leer angenommen; Zone angenommen; `[1.2.3.4]` angenommen; Schema angenommen; Platzhalter angenommen (je Element rot) | andere |
| P12 (Nr. 5 b bis e) | Vergleich: Namen ohne Beachtung der Schreibweise, IPv6 kanonisch; Eintrag ohne Port gleich Standardport des Schemas; nur Gleichheit (kein Teil, keine Endung, kein Platzhalter); ein nicht lesbares `Ziel` gibt nichts frei; ein Eintrag ohne Treffer gibt nichts frei | Raster Ziel × Eintrag mit Soll je Paar: 10 gültige Ziele (`http://h`, `https://h`, `https://h:443`, `https://h:8443`, `http://h:80`, `http://H`, `https://[::1]`, `https://[::1]:8443`, `https://h/pfad`, `https://h?q`) × 9 Einträge (`h`, `H`, `h:443`, `h:80`, `h:8443`, `[::1]`, `[0:0:0:0:0:0:0:1]`, `[::1]:8443`, `x`); dazu 16 nicht lesbare oder fremde Ziele (leer, `h`, `ftp://h`, `http://`, `http://u@h`, `http://h:0`, `http://h:99999`, `http://h:`, `http://[::1`, Zone, Byte ab 0x80, `http://h%41`, `h.` gegen Eintrag `h`, `xh`, `a.h`, `h.example`) mit allen Einträgen: immer `Bleiben`; Zählprüfung der Paare | Teilstring oder Endung verglichen; Port ignoriert; Standardports vertauscht; Schreibweise zählt; Benutzerteil toleriert; Schema nicht geprüft; IPv6 textuell verglichen (je Element rot) | andere |
| P13 (Nr. 6 a) | Eintrag von `Schreibfreigaben`: Methode, ein Leerzeichen, Muster; Ursachen `leer`, `trenner`, `methode`, `muster` in dieser Rangfolge | Menge 19: 5 Gutfälle (je mutierende Methode mit `/svc/Orders`) und 14 Verstöße (leer; `POST`; `/svc/x`; `GET`; `post /svc/x`; `GET /svc/x`; `HEAD /svc/x`; `OPTIONS /svc/x`; ` /svc/x`; `get x`; `POST  /svc/x`; `POST /svc/x ` mit Leerzeichen am Ende; `POST x`; `POST `), Soll-Ursache je Fall, Schlüssel `FALLBACK_WRITE_ALLOW`, `Eintrag` 1 | `muster` vor `methode` geprüft; lesende Methode angenommen; Kleinschreibung angenommen; zwei Leerzeichen angenommen (je Element rot) | andere |
| P14 (Nr. 6 b, f) | Muster: Syntax und Ausschlüsse; höchstens ein `*` je Segment (kein Zurücksetzen im Vergleich) | Menge 37: 9 Gutfälle (`/a`, `/a/b`, `/a/*`, `/a/**`, `/a/Orders(*)`, `/a/*b`, `/a/a*`, `/a/x%2Fy`, `/a/%2e%2ex`) und 28 Verstöße je Regel (kein `/` am Anfang; `/`; `//`; `/a//b`; abschließendes `/`; `?`; `#`; `\`; `,`; Byte 0x1F; Byte 0x7F; Byte ab 0x80; `%`; `%2`; `%G0`; `%2G`; `/a/*b*`; `/a/**b`; `/a/***`; `/**/a`; `/a/x**`; `/a/.`; `/a/..`; `/a/%2e`; `/a/%2E`; `/a/.%2e`; `/a/%2e%2E`; `/a/%2e./b`), Ursache `muster` | die Prüfung auf `.` und `..` entfällt; `%2E` nicht erkannt; zwei `*` im Segment erlaubt; `**` mitten erlaubt; `,` erlaubt; `%` ohne Hexziffern erlaubt (je Element rot) | andere |
| P15 (Nr. 6 c, d) | Vergleich von Muster und Pfad: Segmente, Menge W, kein Dekodieren, leere Segmente passen nie, Segmentzahl gleich außer bei `**`, ein Pfad ohne führendes `/` passt nie | Generator: 301 Muster (Folgen der Länge 1 bis 3 aus {`Orders`, `Orders(1)`, `*`, `Orders(*)`, `a*`, `*b`}, das letzte Segment aus diesen sechs oder `**`: 7 + 42 + 252) × 1117 Pfade der Grammatik (Länge 0 bis 3 aus {`Orders`, `Orders(1)`, `Orders(1,2)`, `ab`, `xb`, `a`, `b`, das leere Segment, `.`, `..`} mit führendem `/`: 1 + 10 + 100 + 1000 = 1111, dazu 6 Pfade ohne führendes `/`); Orakel: aus dem Muster gebaute `regexp` mit der Klasse W; Zählprüfung der Paare und der Treffer je Muster | `*` deckt null Bytes; `*` deckt `/`; `**` deckt null Segmente; `**` nicht nur am Ende; Anfang und Ende überlappen (`a*a` gegen `a`); Anfang per `HasPrefix`; Groß- und Kleinschreibung ignoriert; Segmentzahl nicht verglichen (je Mutation rot) | andere |
| P16 (Nr. 6 c, e) | Nichts wird dekodiert oder normalisiert; ein Platzhalter deckt weder `/` noch `%` noch `\` noch `?` noch `#` und kein Segment `.` oder `..`; ein Muster ohne Platzhalter gleicht byte-gleich | Menge 53: 17 Pfadendungen unter `/svc/Orders/` (`%2e%2e/Geheim`, `%2E%2E`, `%2e`, `..`, `.`, `a%2Fb`, `a%5Cb`, `a` `\` `b`, `a?b`, `a#b`, `a//b`, `a/`, Byte 0x80, NUL, `ä`, `a%zz`, `Orders('a%20b')`) × Muster {`/svc/Orders/*`, `/svc/Orders/**`, `/svc/Orders/Orders(*)`}: kein Treffer (51); dazu das Muster `/svc/Orders/a%2Fb` gegen die Pfade `a%2Fb` (Treffer) und `a%2fb` (kein Treffer) | `%` in W; `\` in W; `..` von `*` gedeckt; `%2e%2e` vor dem Vergleich dekodiert; der Pfad vor dem Vergleich mit `path.Clean` bereinigt (je Mutation rot) | andere |
| P17 (Nr. 7) | Eintrag von `ReplayKategorien`: genau `unsupported` oder `coverage_insufficient`, Schreibweise zählt | Menge 10: 2 Gutfälle und 8 Verstöße (leer mit `leer`; `Unsupported`, `UNSUPPORTED`, `none`, `coverage-insufficient`, `unsupported ` mit Leerzeichen am Ende, `ODR-E6003`, `all` mit `kategorie`), Schlüssel `HYBRID_REPLAY_FALLBACK` | Schreibweise ignoriert; `none` angenommen; Leerraum entfernt | andere |
| P18 (Nr. 8 a, b) | Fehler von `New`: je Schlüssel der erste Verstoß (ab 1), Reihenfolge der Schlüssel, `errors.Join`, `errors.As` je Fehler, keine Policy; die Syntax wird auch bei `Upstream` aus geprüft | Menge 19: Liste {3} allein × `Upstream` {aus, an} (6); je Liste zwei Verstöße an den Stellen 1 und 3 (3); ein gültiger Eintrag vor dem Verstoß (3); Paare von Listen (3); alle drei (1); je Liste zwei Verstöße an den Stellen 2 und 3 (3); Schleife über `errors.As` zählt die Fehler, prüft Schlüssel, `Eintrag` und Reihenfolge, `Policy` ist nil | der letzte statt des ersten Verstoßes; `Eintrag` ab 0; Abbruch nach dem ersten Schlüssel; Reihenfolge der Schlüssel vertauscht; Prüfung nur bei `Upstream` an (je Mutation rot) | andere |
| P19 (Nr. 8 b) | Der Text eines Fehlers nennt Schlüssel, Nummer und Ursache und nie den Wert eines Eintrags | Menge 5: je ein Verstoß mit dem Marker `GEHEIM-7f3a` im Wert für die Ursachen `host`, `trenner`, `methode`, `muster`, `kategorie`; `Error()` enthält Schlüssel, Nummer und Ursache, nicht den Marker | der Text trägt den Wert des Eintrags (`%q`) | andere |
| P20 (Nr. 8 c) | Zusammenhänge: `Upstream` an ohne `Hosts` ist `ohne-hosts` (Schlüssel `HYBRID_FALLBACK`, `Eintrag` 0), entfällt bei ungültigem Eintrag in `Hosts`; `Upstream` aus mit gesetzten Listen und eine Freigabe ohne Host sind keine Fehler | Menge 8: `Upstream` {aus, an} × `Hosts` {leer, gültig, ungültig} (6); aus mit `Schreibfreigaben` und leeren `Hosts` (1); an mit gültigen `Hosts` und einer `Schreibfreigabe` ohne Host im Eintrag (1); Soll aus der Tabelle des Vertrags | `ohne-hosts` auch bei aus; fehlt bei an; zusätzlich bei ungültigem Eintrag; eine Freigabe ohne Host als Fehler (je Mutation rot) | andere |
| P21 (Nr. 8 d, Nr. 9 c) | `New` ruft nichts und liest keine Umgebung; kein Zustand, keine Sperre, keine Goroutine, keine Variable auf Paketebene; Importe nur `model` und die Standardbibliothek | Test über den Quelltext des Pakets (`go/parser`): Menge der Importe ohne `os`, `sync`, `time`, `context`, `net` (außer `net/url` und `net/netip`) und ohne ein Paket des Hexagons außer `model`; kein `go`-Statement; keine `var`-Deklaration auf Paketebene; Methoden nur auf `Policy` und `KonfigFehler` | Import eines Ports; `sync.Mutex`; `os.Getenv`; `net.LookupHost`; `go func`; eine Variable auf Paketebene (je Mutation rot) | andere |
| P22 (Nr. 9 a, b) | `Entscheide` ist total und wiederholbar; die nil-Policy und `Policy{}` antworten `Bleiben` | Menge 3 Policies (nil, `Policy{}`, voll offen) × 6000 Eingaben für Methode, Pfad, Ziel (Zufall mit festem Seed und Grenzformen: leer, NUL, ungültiges UTF-8, ein Pfad von 1 MiB, `%` am Ende, 10000 Segmente) × *C*: keine Panik, nil und `Policy{}` stets `Bleiben`; jede Eingabe zweimal mit 1000 anderen Aufrufen dazwischen, gleiche Entscheidung | Index auf den leeren Pfad (`Pfad[0]`); ein Zähler oder Cache im Receiver; Panik bei nil-Receiver (je Mutation rot) | andere |
| P23 (Nr. 10 a, b) | Aufrufer: derselbe `Ziel`-String wie gesendet, Aufruf vor jedem Senden, Abbildung von `*KonfigFehler` auf `ODR-E2001`, `enabled` ohne `UPSTREAM_URL` | akzeptiertes Negativ: Grund: Hybrid-Dienst und Root entstehen in `slice-hybrid-service` und `slice-bootstrap-m4`; Folge: bis dahin bindet kein Code die Policy und kein Betrieb hängt an ihr; Fänger der Aufrufer ist der M4-Test am Verbindungszähler | keine (akzeptiertes Negativ) | andere |
| P24 (*Ungeprüft* von `SPEC-261`) | Vier akzeptierte Negative mit Grund und Folge im Wortlaut des Vertrags: Name statt Adresse (DNS), Header zum Überschreiben der Methode, Abweichung zwischen den Lesern von `Ziel`, Zustand des Mock bei einem Replay-Fallback einer mutierenden Methode | kein Test in diesem Slice; Folgen: ein Name der Allowlist auf unerwarteter Adresse wird erreicht; eine freigegebene Route kann im Upstream als andere Methode laufen, wenn der Hybrid-Dienst den Header weiterreicht; der Mock bleibt unverändert bei Erfolgsantwort; Fänger der beiden mittleren: `slice-hybrid-service` | keine (akzeptiertes Negativ) | andere |

Die Tabelle hat 24 Zeilen, alle der Klasse `andere`: die Policy hat keine Nebenläufigkeit (§6), und P21 fängt Zustand, Sperre und Goroutine im Quelltext, nicht mit dem Detektor.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Vertrag der Policy** in der Spezifikation: Eingaben (Methode, Zielhost, Pfad, `ODR`-Code oder Kategorie des Ergebnisses, Werte der Policy), Ergebnis (bleiben, Replay-Fallback, Upstream-Fallback) und die Fehler der Prüfung eines Eintrags; Ort und Namen der Typen (Paket unter `src/hexagon/services/hybrid`, `model` oder ein eigenes Paket), die `HybridUseCase` und der Composition Root rufen.
2. **Pfadmuster von `FALLBACK_WRITE_ALLOW`** (`SPEC-080` nennt nur „Methode und Pfadmuster“): Form eines Eintrags, Platzhalter, Bezugspfad (OData-Pfad nach dem Service-Root oder voller Pfad, mit oder ohne Query), Behandlung von Escapes, Schreibweise der Methode, doppelte und leere Einträge.
3. **Host-Vergleich** (`SPEC-062`: nur zum Host der konfigurierten `UPSTREAM_URL`, und nur wenn er in `FALLBACK_ALLOW_HOSTS` steht): Form eines Eintrags (Host oder Host mit Port), Schreibweise, IPv6, und was ein Eintrag bedeutet, der nicht der Host der `UPSTREAM_URL` ist.
4. **Methodenklassen:** lesend sind GET und HEAD (`LH-FA-12.b`); OPTIONS, TRACE, CONNECT und unbekannte Methoden gelten als lesend, mutierend oder nie freigegeben, mit Grund.
5. **Auslöser:** welche `ODR`-Codes die Kategorien `unsupported` (`ODR-E6001`, `ODR-E6002`) und `coverage_insufficient` (`ODR-E6003`) bilden; ein Ergebnis ohne `ODR`-Code (ungültige Anfrage, Kontextfehler, `ODR-E9001`) löst keinen Fallback aus; die Zeile „Replay-Miss“ gilt nur für `ODR-E5001`.
6. **Zusammenhänge der Werte** für `slice-bootstrap-m4`: Hauptschalter `disabled` bei gesetzten `FALLBACK_*`-Schlüsseln, `enabled` mit leerer Allowlist, ein Eintrag von `FALLBACK_WRITE_ALLOW` ohne Host in der Allowlist; Verhalten der Policy und ob der Start sie als `ODR-E2001` abweist (`SPEC-110` Nr. 10 als Muster).
7. **Replay-Fallback und Methode:** `LH-FA-12.b` nennt den Replay-Fallback „nur explizite Policy“ ohne Bezug auf die Methode; gilt er für alle Methoden oder nur für lesende? Die Zelle hat heute keinen Ort in `SPEC-079`.
8. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Ausgang der Aufträge** (Architect, 2026-11-08; alle in `SPEC-261`, Nachzug in `LH-FA-12.a`, `LH-FA-12.b`, den Zeilen `SPEC-062`, `SPEC-079`, `SPEC-080` und `architecture.md` §2.5; die Zeilen `SPEC-010` und `SPEC-110` blieben unverändert, weil sie gelten):

1. Vertrag: Paket `src/hexagon/services/hybrid/policy` (eigenes Paket, damit „kein Port, kein Adapter“ am Quelltext prüfbar bleibt, wenn `HybridUseCase` im Paket `hybrid` hinzukommt), `New(Werte)` und `Entscheide(Anfrage)`, Ergebnis `Bleiben`, `Replay`, `Upstream`, Fehler `*KonfigFehler` — Nr. 1, 4, 8 und der Kopf.
2. Pfadmuster: `<Methode> <Muster>`, voller Pfad samt Service-Root, ohne Query, `*` (ein Segment oder ein Teil davon, höchstens eines je Segment), `**` (Rest, nur am Ende), nichts dekodiert, `%`, `\`, `?`, `#` und `..` von Platzhaltern nie gedeckt; Methode in Großschreibung, Doppelte sind eine Menge, leere Einträge ungültig — Nr. 6.
3. Host: Name oder `[IPv6]`, optional Port; Schreibweise ohne Beachtung der ASCII-Groß- und Kleinschreibung; Eintrag ohne Port gleich Standardport des Schemas; ein Eintrag, der nicht der Host der `UPSTREAM_URL` ist, gibt nichts frei und ist kein Fehler — Nr. 5.
4. Methodenklassen: lesend `GET`, `HEAD`; mutierend `POST`, `PUT`, `PATCH`, `DELETE`, `MERGE`; alles andere nie freigegeben — Nr. 2.
5. Auslöser: `ODR-E6001`, `ODR-E6002` (`unsupported`), `ODR-E6003` (`coverage_insufficient`); Replay-Miss nur `ODR-E5001`; `ODR-E6006` bewusst keiner — Nr. 3.
6. Zusammenhänge: `enabled` ohne `Hosts` ist `ODR-E2001` (`ohne-hosts`); `disabled` mit gesetzten Listen und eine Freigabe ohne Host sind kein Fehler; die Syntax wird immer geprüft; der Root weist `enabled` ohne `UPSTREAM_URL` ab — Nr. 8 c, Nr. 10 b.
7. Replay-Fallback: für alle bekannten Methoden, unabhängig vom Hauptschalter — Nr. 4 b; Frage an den User in §6.
8. Tabelle, Raster und Nachzug: diese Tabelle (24 Zeilen), §6 und die genannten Stellen.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstelle): *von diesem Slice an* `slice-bootstrap-m4` (Lesen der vier Schlüssel und Aufruf der Syntax-Prüfung): angenommen, §1 *Übernimmt* und DoD Punkt 1 von `slice-bootstrap-m4` tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* dort trifft sie. *An diesen Slice:* keine.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein Slice muss in `done/` liegen. Die Spezifikation trägt den Vertrag samt Randformen und der Syntax von Pfadmuster und Host-Vergleich, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Code-Abhängigkeit: keine.** Die Policy liest nur `model` (die Codes `CodeE5001` und `CodeE6001` bis `CodeE6003` stehen seit `slice-fehlermodell-konfiguration` in `src/hexagon/model/fehler.go`) und typisierte Werte; sie ruft keinen Port.

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`:** die Policy braucht weder `MockUseCase` noch `ReplayUseCase` noch `UpstreamClient` und kann auf dem heutigen Stand von `src/` entstehen. Der Trigger der Welle (`welle-hybrid` §2: `welle-mock` in `done/`, M3) ist seit der Closure von `welle-mock` eingetreten; der Start ist keine Umplanung mehr.

**Stand der Bedingungen beim Start** (Planner, 2026-11-08, `main` = `686612a2`): `welle-mock` liegt in `done/` (M3 erreicht), `slice-fehlermodell-konfiguration` liefert die Codes `CodeE5001` und `CodeE6001` bis `CodeE6003`. In `in-progress/` laufen parallel `slice-etag-versionen` und `slice-admin-lebenszyklus`; keiner berührt `src/hexagon/services/hybrid/`, und dieser Slice ändert keinen Pfad der Gate-Kette (Übergangsregel 7, §6), also kein Konflikt. Die Verträge der Policy schreibt der Architect im Zweig (Aufträge 1 bis 8 in §3), bevor der erste Code-Commit entsteht.

**Größenmarke** (Planner, 2026-11-08, `SPEC-128` Nr. 1; **vorläufig, Prosa-Form**; die Tabelle *Zusage · Fänger* fehlt noch, der Block mit `basis`, `schaetzung` und `marke` steht erst nach ihr und vor dem ersten Code-Commit): angenommen 20 Zeilen der Tabelle, *a* = 0, *b* = 20, *f* = 0, *p* = 300 (reine Funktion, kein Port, kein Fake): *S* = 85·0 + 80·20 + 200·0 + 300 = 1900, *M* = 2600, keine Schnittfrage (*S* < 2223). Die Schnittfrage entsteht erst bei 25 oder mehr Zeilen (bei *p* = 300); der Planner rechnet *S* und *M* neu, sobald die Tabelle des Architect vorliegt, und ersetzt diese Zeile mit Grund. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfad des Blocks: `src/hexagon/services/hybrid/`. Basis der Messung: `686612a2` (die Start-Commits ändern allein Pläne).

**Teil des Architect zur Größenmarke** (2026-11-08, `SPEC-128` Nr. 3 k; `basis`, `schaetzung` und `marke` setzt der Planner nach der Tabelle in §3, 24 Zeilen, alle `andere`, keine Fakes; `produktcode` ist die Erwartung für Auswahl, Host, Muster, Prüfung und Fehler):

```groessenmarke
pfad: src/hexagon/services/hybrid
produktcode: 500
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: vorläufig 2600** (siehe *Größenmarke* oben); die endgültige Zahl steht, wenn die Tabelle *Zusage · Fänger* vorliegt.
- `in-progress` → `open` (blockiert — Carveout?): eine Zelle des Rasters (Pfadmuster, Host-Vergleich, Methodenklasse) hat keinen Ort in der Spezifikation, und der Architect entscheidet sie nicht vor dem Code (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt ihren Ort in `SPEC-261` (Architect, 2026-11-08); der Fänger steht in der Tabelle *Zusage · Fänger* in §3 (Zeilen P01 bis P24).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Entscheidung über Replay-Fallback (Code, Kategorien, Methode) | der Nullwert von `Code`, die nil-Policy und `Policy{}`: Nr. 3 c, Nr. 9 a | leere Methode, leere `ReplayKategorien`: Nr. 2 c, Nr. 4 b | ein Code außerhalb der Tabelle, eine unbekannte Methode: Nr. 3 c, Nr. 2 c | gilt nicht: reine Funktion ohne `ctx` | gleiche Eingabe, gleiche Entscheidung: Nr. 9 b |
| Entscheidung über Upstream-Fallback (Methode, Host, Pfad, Code, Werte) | leeres `Ziel`, leerer `Pfad`, Nullwert von `Code`: Nr. 5 c, Nr. 6 d, Nr. 3 c | leere `Hosts` (nichts freigegeben), leere `Schreibfreigaben` (mutierend bleibt): Nr. 5 d, Nr. 4 c | fremder Host, unbekannte Methode, unbekannter Code: Nr. 5 e, Nr. 2 c, Nr. 3 c | gilt nicht: reine Funktion ohne `ctx` | Nr. 9 b |
| Prüfung eines Eintrags von `FALLBACK_WRITE_ALLOW` (Methode und Pfadmuster) | kein Leerzeichen, kein Muster: Nr. 6 a | der leere Eintrag, das leere Muster: Nr. 6 a | unbekannte oder lesende Methode, Zeichen außerhalb der Menge, zwei `*` in einem Segment: Nr. 6 a, Nr. 6 b | gilt nicht: reine Funktion ohne `ctx` | Doppelte Einträge sind eine Menge: Nr. 1 a |
| Prüfung eines Eintrags von `FALLBACK_ALLOW_HOSTS` | Port ohne Name, Name ohne Klammern bei IPv6: Nr. 5 a | der leere Eintrag: Nr. 5 a | Schema, Pfad, Platzhalter, Zone, Port außerhalb 1 bis 65535: Nr. 5 a | gilt nicht: reine Funktion ohne `ctx` | Nr. 1 a |
| Vergleich des Hosts mit `FALLBACK_ALLOW_HOSTS` und `UPSTREAM_URL` | `Ziel` ohne Port (Standardport des Schemas): Nr. 5 d | `Ziel` leer oder ohne Hostnamen: Nr. 5 c | anderer Name, Port oder Schema, nicht lesbares `Ziel`: Nr. 5 c, Nr. 5 d, Nr. 5 e | gilt nicht: reine Funktion ohne `ctx` | Nr. 9 b |
| Vergleich von Pfad und Muster | Pfad ohne führendes `/`: Nr. 6 d | der leere Pfad, ein leeres Segment: Nr. 6 d, Nr. 6 c | `%`, `\`, `?`, `#`, `..`, Bytes außerhalb der Menge W: Nr. 6 c, Nr. 6 e | gilt nicht: reine Funktion ohne `ctx` | Nr. 9 b |
| Prüfung eines Eintrags von `HYBRID_REPLAY_FALLBACK` | gilt nicht: eine Liste ohne Eintrag ist der Default (`SPEC-110` Nr. 1) | der leere Eintrag: Nr. 7 | eine andere Schreibweise oder Kategorie (`none`, `Unsupported`): Nr. 7 | gilt nicht: reine Funktion ohne `ctx` | Nr. 1 a |
| `New` (Zusammenhang der Werte) | `Upstream` an ohne `Hosts`: Nr. 8 c | alle Listen leer bei `Upstream` aus: Nr. 8 c | `Hosts` ohne passenden Service, `Schreibfreigaben` ohne Host: Nr. 8 c, Nr. 5 e | gilt nicht: reine Funktion ohne `ctx` | `New` mehrfach mit denselben Slices: Nr. 1 b |

Zusatztabelle (mehrere Verstöße, Abhängigkeit vom Modus):

| Fall | Erwartung (`SPEC-261`) | Fänger |
|---|---|---|
| zwei Verstöße in derselben Liste | der erste, `Eintrag` ab 1: Nr. 8 b | P18 (Stellen 1 und 3, 2 und 3) |
| Verstöße in zwei oder drei Listen | je Schlüssel ein Fehler, Reihenfolge `FALLBACK_ALLOW_HOSTS`, `FALLBACK_WRITE_ALLOW`, `HYBRID_REPLAY_FALLBACK`: Nr. 8 b | P18 (Paare, alle drei) |
| Verstoß bei `Upstream` aus | Fehler wie bei an: Nr. 8 a | P18 (Liste allein × aus) |
| `Upstream` an ohne `Hosts` zusammen mit einem Verstoß in einer anderen Liste | beide Fehler, der des Zusammenhangs zuletzt: Nr. 8 b, Nr. 8 c | P20, P18 |
| `Upstream` an mit ungültigem Eintrag in `Hosts` | nur der Fehler des Eintrags: Nr. 8 c | P20 |
| Modus des Hauptschalters in der Entscheidung | aus: nie `Upstream`; an: nach Nr. 4 c; `Replay` unabhängig: Nr. 4 b, Nr. 4 f | P05, P06, P10 |

- Pfadmuster und Host-Vergleich haben in `SPEC-062` und `SPEC-080` keine Syntax — **Ausgang:** entfallen: entschieden in `SPEC-261` Nr. 5 und Nr. 6 (Architect, 2026-11-08).
- Methoden außerhalb von GET, HEAD, POST, PUT, PATCH und DELETE (Auftrag 4) und der Replay-Fallback für mutierende Methoden (Auftrag 7) haben keinen Ort — **Ausgang:** entfallen: entschieden in `SPEC-261` Nr. 2 und Nr. 4 b; die Frage, ob der Replay-Fallback mutierender Methoden gewollt ist, steht unten.
- Der Replay-Fallback hat im Lastenheft weder Anforderung noch Akzeptanzkriterium ([Abgleichsbericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §8 Nr. 2); die Policy leitet ihn aus `LH-FA-12.a` und `LH-FA-12.b` der Spezifikation ab — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User (Change Request vor `Accepted` oder Bestätigung).
- **Frage (Architect, 2026-11-08):** Soll der Replay-Fallback auch für mutierende Methoden gelten (`SPEC-261` Nr. 4 b: ja, weil `LH-FA-12.b` ihn nicht an die Methode bindet und er keinen Upstream kontaktiert), mit der Folge, dass der Mock unverändert bleibt, während der Client eine aufgezeichnete Erfolgsantwort sieht? Die Alternative „nur lesend“ ist eine Zeile in Nr. 4 b und die Zeilen P05 und P06 — **Ausgang:** weiter offen; Adresse: User; bis zur Antwort gilt Nr. 4 b.
- **Frage (Architect, 2026-11-08):** Soll `ODR-E6006` (V2: Eigenschaft außerhalb des Umfangs) ein Auslöser werden? `SPEC-261` Nr. 3 d sagt nein, weil `LH-FA-12.b` den Code nicht nennt; ein V2-Service fiele sonst bei einer nicht unterstützten Eigenschaft auf den Upstream zurück. Das wäre eine Änderung der Matrix im Lastenheft — **Ausgang:** weiter offen; Adresse: User; bis zur Antwort kein Fallback (die sichere Seite).
- Eine Policy, die in einem Zustand fehlt, den die Matrix nicht nennt (unbekannter Code, unbekannte Methode, leere Werte), erlaubt nie einen Fallback; der Fänger ist die Schleife über Methoden × Zustände × Codes mit Zählprüfung (P03, P06, P10), die Mutation „Standardzweig erlaubt“ färbt sie rot — **Ausgang:** offen bis zur Closure.
- Wechselwirkung (`SPEC-261` Nr. 4 e): wer für eine Kategorie den Replay-Fallback setzt, schaltet den schreibenden Upstream-Fallback für deren Codes ab, weil die Zeile „Replay-Miss“ der Matrix nur `GET` und `HEAD` kennt — **Ausgang:** weiter offen; Adresse: Planner (Sendung an `slice-bootstrap-m4`: das Handbuch nennt die Wechselwirkung; vor der Nennung liest der Planner dessen §1 *Ausdrücklich NICHT* und die DoD, `AGENTS.md` §3.13).
- Sendungen aus Nr. 10 von `SPEC-261` (`AGENTS.md` §3.13): an `slice-hybrid-service` — `Ziel` ist dieselbe Zeichenkette, die gesendet wird; `Entscheide` vor jedem Senden; welche Header der Fallback sendet, darunter die Header zum Überschreiben der Methode (*Ungeprüft* von `SPEC-261`); der Replay-Fallback einer mutierenden Methode ändert den Mock nicht. An `slice-bootstrap-m4` — `*KonfigFehler` wird `ODR-E2001`, `enabled` ohne `UPSTREAM_URL` wird `ODR-E2001`, die Reihenfolge der Schlüssel in der Meldung. Nehmer prüfen *Ausdrücklich NICHT* und DoD, bevor der Planner die Sendung einträgt — **Ausgang:** offen; Adresse: Planner.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): die Policy hat keine Nebenläufigkeit, der Detektor hat nichts zu melden; jeder Zweig der Matrix hat einen Fall der Tabelle; P21 fängt Zustand, Sperre und Goroutine im Quelltext — **Ausgang:** entfallen: reine Funktion ohne Zustand, Zählprüfung der Fälle gegen die Zeilen von `LH-FA-12.b` statt eines Laufs unter dem Detektor.
- Größe: die Marke steht vorläufig als Prosa in §4 (*S* = 1900, *M* = 2600, angenommen 20 Zeilen); der Block mit dem Teil des Architect steht in §4. Die Tabelle hat 24 Zeilen, alle `andere`; mit `produktcode` 500 und `fakes` 0 ergibt `SPEC-128` Nr. 1 *S* = 80·24 + 500 = 2420 und *M* = 3300, und *S* ≥ 2223 stellt die Schnittfrage. Die natürliche Naht liegt zwischen der Prüfung der Einträge und `New` (P11, P13, P14, P17 bis P20) und der Entscheidung (P02 bis P10, P12, P15, P16); beide teilen den Mechanismus der geparsten Einträge in `Policy` — **Ausgang:** offen; Adresse: Planner (Rechnung, Schnittfrage oder Grund in §4 und Block mit `basis`, `schaetzung` und `marke` in einem Commit vor dem ersten Code-Commit).
- Gate-Kette (Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch`): der Slice ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Reine Funktion mit großer Matrix: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×, offen) gilt nur eingeschränkt (§6). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `AGENTS.md` §3.16, für „kein Fall …“ der Matrix) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
