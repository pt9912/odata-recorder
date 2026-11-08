# Slice slice-fallback-policy-entscheidung: Entscheidung der Fallback-Policy (Matrix, Host-Vergleich, Muster-Vergleich)

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

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` (Entscheidungsmatrix) · `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080` · `SPEC-020`, `SPEC-021`, `SPEC-022`, `SPEC-023`, `SPEC-024` (die Codes, nach denen die Matrix entscheidet) · `ARC-002` · `architecture.md` §2.5 (Modus-Routing) · `SPEC-261` Nr. 1 (Listen sind Mengen), 2 bis 4 (Methodenklassen, Auslöser, Matrix), 5 b bis e und 6 c bis e (Vergleich von Host und Pfad), 9 und 10 (Totalität, Aufrufer) (Vertrag der Policy, Architect).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `Entscheide` auf der Policy, die `slice-fallback-policy` aus den geprüften Einträgen baut (Paket `src/hexagon/services/hybrid/policy`), legt aus Methode, Zielhost, Pfad und dem `ODR`-Code des Ergebnisses fest, ob die Antwort bleibt, ein Replay-Fallback oder ein Upstream-Fallback erlaubt ist; ohne ausdrückliche Freigabe lautet jede Entscheidung „kein Fallback“, und sie hängt am `ODR`-Code, nie am HTTP-Status (`LH-FA-12.b`). Der Slice ist die zweite Hälfte des geschnittenen Plans der Policy (Schnittfrage `SPEC-128` Nr. 1: Schätzung 2420 für den ungeschnittenen Plan, §4).

**Übernimmt:** von `slice-fallback-policy` (dessen §1 *Ausdrücklich NICHT* diesen Slice nennt) die Entscheidung über die geparsten Einträge: die Policy, ihre Typen, die Syntax und die geparste Form von Host, Methode und Muster, `New` und `*KonfigFehler` liegen dort in `done/`; dieser Slice schreibt nur die Entscheidung und die Vergleiche darauf und keinen zweiten Parser (*Code-Abhängigkeit*).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Syntax und Prüfung der Einträge von `FALLBACK_ALLOW_HOSTS`, `FALLBACK_WRITE_ALLOW` und `HYBRID_REPLAY_FALLBACK`, `New`, `*KonfigFehler` und die Zusammenhänge der Werte (`SPEC-261` Nr. 5 a, 6 a und b, 7, 8) — `slice-fallback-policy`; die Entscheidung liest die geparste Form und prüft nichts neu.
- Der Aufruf von `MockUseCase`, `ReplayUseCase` und `UpstreamClient`, Strom und Zustand — `slice-hybrid-service`; die Policy ist eine Funktion ohne Port, damit die Matrix ohne Fakes vollständig prüfbar ist und kein Zweig von einem Fake abhängt.
- Das Lesen der Schlüssel `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` aus der Umgebung und das Aufheben ihrer Ablehnung in `SPEC-110` Nr. 2 — `slice-bootstrap-m4`; solange kein Composition Root die Policy verdrahtet, bliebe ein gelesener Schlüssel eine still wirkungslose Sicherheitseinstellung.
- Das Senden an den Upstream und das Speichern einer Fallback-Antwort — `slice-hybrid-service` (Senden); Speichern und Datenabgleich sind Out-of-Scope von `LH-FA-12`.
- Spans, Audit und Metriken der Entscheidung — `welle-betrieb`; es gibt noch keinen Port `Telemetry`.

**Sendungen** (`AGENTS.md` §3.13; die Nehmer tragen sie seit dem Schnitt mit der Kennung dieses Gebers in §1 und DoD): an `slice-hybrid-service` — `Ziel` ist dieselbe Zeichenkette, die gesendet wird; `Entscheide` vor jedem Senden; Auswahl der gesendeten Header, darunter die Header zum Überschreiben der Methode (*Ungeprüft* von `SPEC-261`); der Replay-Fallback einer mutierenden Methode ändert den Mock nicht. An `slice-bootstrap-m4` — `*KonfigFehler` wird `ODR-E2001`, `enabled` ohne `UPSTREAM_URL` wird `ODR-E2001`, das Handbuch nennt die Wechselwirkung aus `SPEC-261` Nr. 4 e.

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

- [ ] Entscheidungsmatrix `LH-FA-12.b`: je Zeile (Erfolg, fachliche `404`/`409`/`412`, Coverage unvollständig, nicht unterstützte Operation, ungültige Anfrage, Replay-Miss) und je Zustand der Policy das Ergebnis; Auslöser ist der `ODR`-Code, nie der HTTP-Status; unter den Replay-Fehlern löst nur `ODR-E5001` einen Fallback aus; `Entscheide` ist total, wiederholbar und antwortet auf die nil-Policy und `Policy{}` mit `Bleiben`.
- [ ] Upstream-Freigabe: nur bei Hauptschalter `enabled`, nur zu einem Host aus `FALLBACK_ALLOW_HOSTS`, der der Host der konfigurierten `UPSTREAM_URL` ist (Vergleich nach `SPEC-261` Nr. 5 b bis e), lesend nur mit GET und HEAD, mutierend nur mit passendem Eintrag aus `FALLBACK_WRITE_ALLOW` (Methode und Vergleich des Pfadmusters nach `SPEC-261` Nr. 6 c bis e); sonst nichts.
- [ ] Replay-Fallback: nur für die Kategorien `unsupported` und `coverage_insufficient` aus `HYBRID_REPLAY_FALLBACK`, unabhängig vom Hauptschalter und von Host und Ziel; Listen sind Mengen (Reihenfolge und Doppelte ohne Wirkung).
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
| `src/hexagon/services/hybrid/policy/` | update | `Entscheide`, Methodenklassen, Auslöser, Vergleich von Host und Pfadmuster auf der geparsten Form aus `slice-fallback-policy` |
| Tests | neu | Matrix `LH-FA-12.b` als Tabelle × Zustände der Policy; Generator über Methoden, Hosts und Muster mit unabhängigem Orakel; die Policy importiert weder Port noch Adapter (`make a-check`) |

Der Vertrag steht in `SPEC-261` (Architect, 2026-11-08); die Spezifikation ist unverändert zu führen, ein Nachzug entfällt, solange der Slice den Vertrag nicht berichtigt. Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster in §6 (`AGENTS.md` §3.16) liegen vor dem Start; die Randformen entscheidet der Architect, nicht der Implementer. Die Zeilen tragen die Kennungen des ungeschnittenen Plans (P01 bis P24); die Zeilen P11, P13, P14 und P17 bis P21 führt `slice-fallback-policy`, dazu dort die Zeile P01a, die P01 als akzeptiertes Negativ bis zu diesem Slice trägt (dessen Fänger P01 steht hier). Die Mengen: *C* = `model.AlleCodes()` (*N* Codes) plus der Nullwert, −1 und *N* + 1 (*N* + 3 Werte); *M* = 13 Methoden (die sieben bekannten, `OPTIONS`, `TRACE`, `CONNECT`, `get`, die leere und `PROPFIND`); *Z* = 112 Zustände der Policy (`Upstream` aus: `Hosts` {leer, passend, fremd, passend mit falschem Port} × `Schreibfreigaben` {keine, passend, andere Methode, anderes Muster} × `ReplayKategorien` {keine, `unsupported`, `coverage_insufficient`, beide} = 64; `Upstream` an: `Hosts` ohne die leere = 48).

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
| P12 (Nr. 5 b bis e) | Vergleich: Namen ohne Beachtung der Schreibweise, IPv6 kanonisch; Eintrag ohne Port gleich Standardport des Schemas; nur Gleichheit (kein Teil, keine Endung, kein Platzhalter); ein nicht lesbares `Ziel` gibt nichts frei; ein Eintrag ohne Treffer gibt nichts frei | Raster Ziel × Eintrag mit Soll je Paar: 10 gültige Ziele (`http://h`, `https://h`, `https://h:443`, `https://h:8443`, `http://h:80`, `http://H`, `https://[::1]`, `https://[::1]:8443`, `https://h/pfad`, `https://h?q`) × 9 Einträge (`h`, `H`, `h:443`, `h:80`, `h:8443`, `[::1]`, `[0:0:0:0:0:0:0:1]`, `[::1]:8443`, `x`); dazu 16 nicht lesbare oder fremde Ziele (leer, `h`, `ftp://h`, `http://`, `http://u@h`, `http://h:0`, `http://h:99999`, `http://h:`, `http://[::1`, Zone, Byte ab 0x80, `http://h%41`, `h.` gegen Eintrag `h`, `xh`, `a.h`, `h.example`) mit allen Einträgen: immer `Bleiben`; Zählprüfung der Paare | Teilstring oder Endung verglichen; Port ignoriert; Standardports vertauscht; Schreibweise zählt; Benutzerteil toleriert; Schema nicht geprüft; IPv6 textuell verglichen (je Element rot) | andere |
| P15 (Nr. 6 c, d) | Vergleich von Muster und Pfad: Segmente, Menge W, kein Dekodieren, leere Segmente passen nie, Segmentzahl gleich außer bei `**`, ein Pfad ohne führendes `/` passt nie | Generator: 301 Muster (Folgen der Länge 1 bis 3 aus {`Orders`, `Orders(1)`, `*`, `Orders(*)`, `a*`, `*b`}, das letzte Segment aus diesen sechs oder `**`: 7 + 42 + 252) × 1117 Pfade der Grammatik (Länge 0 bis 3 aus {`Orders`, `Orders(1)`, `Orders(1,2)`, `ab`, `xb`, `a`, `b`, das leere Segment, `.`, `..`} mit führendem `/`: 1 + 10 + 100 + 1000 = 1111, dazu 6 Pfade ohne führendes `/`); Orakel: aus dem Muster gebaute `regexp` mit der Klasse W; Zählprüfung der Paare und der Treffer je Muster | `*` deckt null Bytes; `*` deckt `/`; `**` deckt null Segmente; `**` nicht nur am Ende; Anfang und Ende überlappen (`a*a` gegen `a`); Anfang per `HasPrefix`; Groß- und Kleinschreibung ignoriert; Segmentzahl nicht verglichen (je Mutation rot) | andere |
| P16 (Nr. 6 c, e) | Nichts wird dekodiert oder normalisiert; ein Platzhalter deckt weder `/` noch `%` noch `\` noch `?` noch `#` und kein Segment `.` oder `..`; ein Muster ohne Platzhalter gleicht byte-gleich | Menge 53: 17 Pfadendungen unter `/svc/Orders/` (`%2e%2e/Geheim`, `%2E%2E`, `%2e`, `..`, `.`, `a%2Fb`, `a%5Cb`, `a` `\` `b`, `a?b`, `a#b`, `a//b`, `a/`, Byte 0x80, NUL, `ä`, `a%zz`, `Orders('a%20b')`) × Muster {`/svc/Orders/*`, `/svc/Orders/**`, `/svc/Orders/Orders(*)`}: kein Treffer (51); dazu das Muster `/svc/Orders/a%2Fb` gegen die Pfade `a%2Fb` (Treffer) und `a%2fb` (kein Treffer) | `%` in W; `\` in W; `..` von `*` gedeckt; `%2e%2e` vor dem Vergleich dekodiert; der Pfad vor dem Vergleich mit `path.Clean` bereinigt (je Mutation rot) | andere |
| P22 (Nr. 9 a, b) | `Entscheide` ist total und wiederholbar; die nil-Policy und `Policy{}` antworten `Bleiben` | Menge 3 Policies (nil, `Policy{}`, voll offen) × 6000 Eingaben für Methode, Pfad, Ziel (Zufall mit festem Seed und Grenzformen: leer, NUL, ungültiges UTF-8, ein Pfad von 1 MiB, `%` am Ende, 10000 Segmente) × *C*: keine Panik, nil und `Policy{}` stets `Bleiben`; jede Eingabe zweimal mit 1000 anderen Aufrufen dazwischen, gleiche Entscheidung | Index auf den leeren Pfad (`Pfad[0]`); ein Zähler oder Cache im Receiver; Panik bei nil-Receiver (je Mutation rot) | andere |
| P23 (Nr. 10 a, b) | Aufrufer: derselbe `Ziel`-String wie gesendet, Aufruf vor jedem Senden, Abbildung von `*KonfigFehler` auf `ODR-E2001`, `enabled` ohne `UPSTREAM_URL` | akzeptiertes Negativ: Grund: Hybrid-Dienst und Root entstehen in `slice-hybrid-service` und `slice-bootstrap-m4`; Folge: bis dahin bindet kein Code die Policy und kein Betrieb hängt an ihr; Fänger der Aufrufer ist der M4-Test am Verbindungszähler | keine (akzeptiertes Negativ) | andere |
| P24 (*Ungeprüft* von `SPEC-261`) | Vier akzeptierte Negative mit Grund und Folge im Wortlaut des Vertrags: Name statt Adresse (DNS), Header zum Überschreiben der Methode, Abweichung zwischen den Lesern von `Ziel`, Zustand des Mock bei einem Replay-Fallback einer mutierenden Methode | kein Test in diesem Slice; Folgen: ein Name der Allowlist auf unerwarteter Adresse wird erreicht; eine freigegebene Route kann im Upstream als andere Methode laufen, wenn der Hybrid-Dienst den Header weiterreicht; der Mock bleibt unverändert bei Erfolgsantwort; Fänger der beiden mittleren: `slice-hybrid-service` | keine (akzeptiertes Negativ) | andere |

Die Tabelle hat 16 Zeilen, alle der Klasse `andere`: die Policy hat keine Nebenläufigkeit (§6), und P21 in `slice-fallback-policy` fängt Zustand, Sperre und Goroutine im Quelltext, nicht mit dem Detektor; P21 bleibt in diesem Slice grün (`Entscheide` liefert weder Variable auf Paketebene noch Goroutine noch einen verbotenen Import).

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect:** keine offenen; der Vertrag, die Tabelle und das Raster stehen (`SPEC-261`, Architect, 2026-11-08). Eine Randform, die der Implementer in der Entscheidung bemerkt und die dort nicht steht, geht an den Architect zurück (`AGENTS.md` §3.12).

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-11-08 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *von diesem Slice an* `slice-hybrid-service` (vier Pflichten des Aufrufers) und `slice-bootstrap-m4` (`*KonfigFehler`, `enabled` ohne `UPSTREAM_URL`, Wechselwirkung im Handbuch): angenommen, beide Nehmer tragen die Sendung in §1 *Übernimmt* und in der DoD mit der Kennung dieses Gebers (Commit vor diesem Plan), kein *Ausdrücklich NICHT* dort trifft sie. *An diesen Slice:* von `slice-fallback-policy` die geparste Form der Einträge (Code-Abhängigkeit, kein Auftrag).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-fallback-policy` liegt in `done/` (Merge nach `main`). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-261`) bereits.

**Warum:** `slice-fallback-policy` liefert `Policy`, die geparste Form von Host, Methode und Muster, `New` und `*KonfigFehler`; die Entscheidung liest sie (*Code-Abhängigkeit*, dieselben Dateien unter `src/hexagon/services/hybrid/policy/`). Keine Abhängigkeit von `welle-mock`, `welle-record-replay` und M2: die Entscheidung braucht weder `MockUseCase` noch `ReplayUseCase` noch `UpstreamClient`.

**Größenmarke** (Planner, 2026-11-08, `SPEC-128` Nr. 1; Schnitt von `slice-fallback-policy`, dessen §4 die Rechnung des ungeschnittenen Plans trägt: 24 Zeilen, *p* = 500, *S* = 2420, *M* = 3300, Schnittfrage ja): 16 Zeilen der Tabelle (P01 bis P10, P12, P15, P16, P22 bis P24), *a* = 0, *b* = 16, *f* = 0, *p* = 200 (Auswahl, Methodenklassen, Vergleich von Host und Muster auf der geparsten Form): *S* = 85·0 + 80·16 + 200·0 + 200 = 1480, *M* = 2000, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Planner setzt `basis`, `schaetzung` und `marke` beim Start; **Basis** ist der letzte Commit von `slice-fallback-policy` unter `src/hexagon/services/hybrid`, damit die Messung allein diesen Slice zählt (Pfad `src/hexagon/services/hybrid` ist der des Architect). Der erste Teil (`slice-fallback-policy`) hat 9 Zeilen (die Zeile P01a als akzeptiertes Negativ zählt mit), *p* = 300, *S* = 1020, *M* = 1400, Basis `9310c219`; die Summe der Schätzungen (2500) liegt um diese Zeile (80) über der des ungeschnittenen Plans (2420), die der Marken (3400) um 100 über 3300.

**Teil des Architect zur Größenmarke** (2026-11-08, `SPEC-128` Nr. 3 k; für den Teil der Entscheidung übernommen aus dem Block von `slice-fallback-policy`, `produktcode` auf den Anteil dieses Slice gesetzt vom Planner mit dem Schnitt):

```groessenmarke
pfad: src/hexagon/services/hybrid
produktcode: 200
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). **Marke: 2000** (siehe *Größenmarke* oben).
- `in-progress` → `open` (blockiert — Carveout?): eine Zelle des Rasters (Vergleich von Host oder Muster, Methodenklasse) hat keinen Ort in der Spezifikation, und der Architect entscheidet sie nicht vor dem Code (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt ihren Ort in `SPEC-261` (Architect, 2026-11-08); der Fänger steht in der Tabelle *Zusage · Fänger* in §3. Die Operationen der Prüfung von Einträgen und `New` führt `slice-fallback-policy`.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Entscheidung über Replay-Fallback (Code, Kategorien, Methode) | der Nullwert von `Code`, die nil-Policy und `Policy{}`: Nr. 3 c, Nr. 9 a | leere Methode, leere `ReplayKategorien`: Nr. 2 c, Nr. 4 b | ein Code außerhalb der Tabelle, eine unbekannte Methode: Nr. 3 c, Nr. 2 c | gilt nicht: reine Funktion ohne `ctx` | gleiche Eingabe, gleiche Entscheidung: Nr. 9 b |
| Entscheidung über Upstream-Fallback (Methode, Host, Pfad, Code, Werte) | leeres `Ziel`, leerer `Pfad`, Nullwert von `Code`: Nr. 5 c, Nr. 6 d, Nr. 3 c | leere `Hosts` (nichts freigegeben), leere `Schreibfreigaben` (mutierend bleibt): Nr. 5 d, Nr. 4 c | fremder Host, unbekannte Methode, unbekannter Code: Nr. 5 e, Nr. 2 c, Nr. 3 c | gilt nicht: reine Funktion ohne `ctx` | Nr. 9 b |
| Vergleich des Hosts mit `FALLBACK_ALLOW_HOSTS` und `UPSTREAM_URL` | `Ziel` ohne Port (Standardport des Schemas): Nr. 5 d | `Ziel` leer oder ohne Hostnamen: Nr. 5 c | anderer Name, Port oder Schema, nicht lesbares `Ziel`: Nr. 5 c, Nr. 5 d, Nr. 5 e | gilt nicht: reine Funktion ohne `ctx` | Nr. 9 b |
| Vergleich von Pfad und Muster | Pfad ohne führendes `/`: Nr. 6 d | der leere Pfad, ein leeres Segment: Nr. 6 d, Nr. 6 c | `%`, `\`, `?`, `#`, `..`, Bytes außerhalb der Menge W: Nr. 6 c, Nr. 6 e | gilt nicht: reine Funktion ohne `ctx` | Nr. 9 b |

Zusatztabelle (Abhängigkeit vom Modus):

| Fall | Erwartung (`SPEC-261`) | Fänger |
|---|---|---|
| Modus des Hauptschalters in der Entscheidung | aus: nie `Upstream`; an: nach Nr. 4 c; `Replay` unabhängig: Nr. 4 b, Nr. 4 f | P05, P06, P10 |

Die Zusatztabelle der Verstöße am selben Schlüssel (mehrere Verstöße je Liste und über Listen) führt `slice-fallback-policy`.

- Der Replay-Fallback hat im Lastenheft weder Anforderung noch Akzeptanzkriterium ([Abgleichsbericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §8 Nr. 2); die Policy leitet ihn aus `LH-FA-12.a` und `LH-FA-12.b` der Spezifikation ab — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User (Change Request vor `Accepted` oder Bestätigung).
- **Frage (Architect, 2026-11-08):** Soll der Replay-Fallback auch für mutierende Methoden gelten (`SPEC-261` Nr. 4 b: ja, weil `LH-FA-12.b` ihn nicht an die Methode bindet und er keinen Upstream kontaktiert), mit der Folge, dass der Mock unverändert bleibt, während der Client eine aufgezeichnete Erfolgsantwort sieht? Die Alternative „nur lesend“ ist eine Zeile in Nr. 4 b und die Zeilen P05 und P06 — **Ausgang:** weiter offen; Adresse: User; bis zur Antwort gilt Nr. 4 b.
- **Frage (Architect, 2026-11-08):** Soll `ODR-E6006` (V2: Eigenschaft außerhalb des Umfangs) ein Auslöser werden? `SPEC-261` Nr. 3 d sagt nein, weil `LH-FA-12.b` den Code nicht nennt; ein V2-Service fiele sonst bei einer nicht unterstützten Eigenschaft auf den Upstream zurück. Das wäre eine Änderung der Matrix im Lastenheft — **Ausgang:** weiter offen; Adresse: User; bis zur Antwort kein Fallback (die sichere Seite).
- Eine Policy, die in einem Zustand fehlt, den die Matrix nicht nennt (unbekannter Code, unbekannte Methode, leere Werte), erlaubt nie einen Fallback; der Fänger ist die Schleife über Methoden × Zustände × Codes mit Zählprüfung (P03, P06, P10), die Mutation „Standardzweig erlaubt“ färbt sie rot — **Ausgang:** offen bis zur Closure.
- Wechselwirkung (`SPEC-261` Nr. 4 e): wer für eine Kategorie den Replay-Fallback setzt, schaltet den schreibenden Upstream-Fallback für deren Codes ab, weil die Zeile „Replay-Miss“ der Matrix nur `GET` und `HEAD` kennt — **Ausgang:** eingetreten als Sendung: `slice-bootstrap-m4` nimmt sie an (DoD Punkt 7, Handbuch nennt die Wechselwirkung; Planner, 2026-11-08, §3.13 geprüft).
- Sendungen aus Nr. 10 von `SPEC-261` (`AGENTS.md` §3.13): an `slice-hybrid-service` — `Ziel` ist dieselbe Zeichenkette, die gesendet wird; `Entscheide` vor jedem Senden; Auswahl der gesendeten Header, darunter die Header zum Überschreiben der Methode (*Ungeprüft* von `SPEC-261`); der Replay-Fallback einer mutierenden Methode ändert den Mock nicht. An `slice-bootstrap-m4` — `*KonfigFehler` wird `ODR-E2001`, `enabled` ohne `UPSTREAM_URL` wird `ODR-E2001`, die Reihenfolge der Schlüssel in der Meldung — **Ausgang:** eingetreten als Sendung: beide Nehmer tragen sie in §1 und DoD (Commit vor diesem Plan, Planner 2026-11-08).
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): die Policy hat keine Nebenläufigkeit, der Detektor hat nichts zu melden; jeder Zweig der Matrix hat einen Fall der Tabelle — **Ausgang:** entfallen: reine Funktion ohne Zustand, Zählprüfung der Fälle gegen die Zeilen von `LH-FA-12.b` statt eines Laufs unter dem Detektor.
- Größe: *S* = 1480 und *M* = 2000 (§4), keine Schnittfrage; der Block mit `basis`, `schaetzung` und `marke` entsteht beim Start vor dem ersten Code-Commit — **Ausgang:** offen; Adresse: Planner beim Start.
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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-08, aus dem Plan von `slice-fallback-policy` übernommen); der Planner zieht Treffer und Zähler beim Start nach. Reine Funktion mit großer Matrix: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×, offen) gilt nur eingeschränkt (§6). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `AGENTS.md` §3.16, für „kein Fall …“ der Matrix) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
