# Slice slice-fallback-policy: Fallback-Policy, Prüfung der Einträge und New

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

**Berührte Spec-Stellen:** `LH-FA-12.a` · `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080` (die Einstellungen, deren Einträge die Policy prüft) · `SPEC-110` Nr. 2 (die Schlüssel, die diese Version noch ablehnt) · `ARC-002` · `architecture.md` §2.5 (Modus-Routing) · `SPEC-261` Nr. 1, 5 a, 6 a und b, 7, 8, 9 c (Vertrag der Policy, Architect; der Satz zu `*` in Nr. 6 b trägt Nr. 6 c und 6 e mit); die Entscheidung (`LH-FA-12.b`, `SPEC-261` Nr. 2 bis 4, 5 b bis e, 6 c bis e, 9 a und b, 10) trägt `slice-fallback-policy-entscheidung`.
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

**Ziel:** Ein Paket `src/hexagon/services/hybrid/policy` (der Ort ist Sache des Architect, `SPEC-261` Nr. 1) nimmt die typisierten Werte der Policy (Hauptschalter, Allowlist der Hosts, Routenfreigaben für mutierende Methoden, Kategorien des Replay-Fallbacks), prüft die Syntax jedes Eintrags und die Zusammenhänge der Werte und liefert die geparste, unveränderliche `Policy` oder einen `*KonfigFehler` mit Schlüssel, Nummer des Eintrags und Ursache. Die Entscheidung über die Policy (`Entscheide`) schreibt `slice-fallback-policy-entscheidung` auf der geparsten Form dieses Slice.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `Entscheide`, die Auslöser nach `ODR`-Code, die Methodenklassen, der Vergleich von Zielhost und Pfad mit den geparsten Einträgen und der Replay-Fallback — `slice-fallback-policy-entscheidung`; der Schnitt trennt die Prüfung der Einträge von der Entscheidung, weil die Schätzung des ungeschnittenen Plans über der Schnittfrage lag (§4); der zweite Teil liest die geparste Form und schreibt keinen zweiten Parser.
- Der Aufruf von `MockUseCase`, `ReplayUseCase` und `UpstreamClient`, Strom und Zustand — `slice-hybrid-service`; die Policy ist eine Funktion ohne Port, damit sie ohne Fakes vollständig prüfbar ist und kein Zweig von einem Fake abhängt.
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

- [ ] Eintrag von `FALLBACK_ALLOW_HOSTS`: Syntax (Name oder IPv6 in Klammern, optional Port) mit den Ursachen `leer` und `host` nach `SPEC-261` Nr. 5 a, geparst abgelegt für den Vergleich in `slice-fallback-policy-entscheidung`.
- [ ] Eintrag von `FALLBACK_WRITE_ALLOW` und `HYBRID_REPLAY_FALLBACK`: Syntax von Methode, Pfadmuster und Kategorie mit den Ursachen `leer`, `trenner`, `methode`, `muster` und `kategorie` in der Rangfolge nach `SPEC-261` Nr. 6 a und b und Nr. 7, geparst abgelegt.
- [ ] `New` und `*KonfigFehler`: je Schlüssel der erste Verstoß, Reihenfolge der Schlüssel, `errors.Join`, kein Wert eines Eintrags im Text, Zusammenhang `ohne-hosts` (`SPEC-261` Nr. 8); `New` ruft nichts und liest keine Umgebung, das Paket importiert nur `model` und die Standardbibliothek (`SPEC-261` Nr. 9 c).
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
| `src/hexagon/services/hybrid/policy/` | neu | `Policy` mit geparsten Einträgen, `New(Werte)`, `*KonfigFehler` als reine Funktionen über `model`-Typen und typisierte Werte (Ort und Namen: Architect) |
| `spec/spezifikation.md` | update (Architect) | `SPEC-261` (Vertrag der Policy samt Randformen) neu; Nachzug an jeder Stelle, die dieselbe Aussage trägt (`AGENTS.md` §3.9): `LH-FA-12.a`, `LH-FA-12.b`, `SPEC-062`, `SPEC-079`, `SPEC-080`; `SPEC-010` und `SPEC-110` gelten unverändert |
| `spec/architecture.md` | update (Architect) | Verzeichnisbaum um `hybrid/policy/`; §2.5 nennt die Policy als reine Entscheidung ohne Port |
| Tests | neu | Syntax der Einträge als Mengen mit Soll-Ursache; Fehler von `New` als Raster; die Policy importiert weder Port noch Adapter (`make a-check`) |

Der Vertrag steht in `SPEC-261` (Architect, 2026-11-08). Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15; die Spalte *Nr.* trägt die Kennung der Zeile und den Ort im Vertrag) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstanden mit dem Vertrag **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Zeilen tragen die Kennungen des ungeschnittenen Plans (P01 bis P24); die Zeilen P01 bis P10, P12, P15, P16 und P22 bis P24 führt `slice-fallback-policy-entscheidung`, und die Zeile P01 ihres Fängers wegen (die Zusage ist über `Entscheide` beobachtbar), hier steht dafür die Zeile P01a. Die Tabelle hat neun Zeilen.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| P01a (Nr. 1 a bis c, Teil von P01) | `New` hält keine Referenz auf die Slices; Listen sind Mengen (Reihenfolge und Doppelte ohne Wirkung); die Policy ist nach `New` unveränderlich | akzeptiertes Negativ: Grund: ohne `Entscheide` hat die Policy in diesem Slice keine Wirkung, die ein Fänger beobachten könnte; Folge: bis `slice-fallback-policy-entscheidung` bindet kein Code die Policy; ein Kopierfehler in `New` fiele erst dort auf und wäre dort zu beheben; Fänger ist P01 in `slice-fallback-policy-entscheidung` (Mengen 9: Liste × Eingriff, gleiche Entscheidungen wie ohne Eingriff) | keine (akzeptiertes Negativ) | andere |
| P11 (Nr. 5 a) | Eintrag von `Hosts`: Name oder IPv6 in Klammern (auch die IPv4-abgebildete Form), optional Port ohne führende Null von 1 bis 65535; Ursache `leer` oder `host` | Menge 30: 10 Gutfälle (`h`, `h.example`, `h-1_x`, `h:1`, `h:65535`, `[::1]`, `[::1]:8080`, `[2001:db8::1]`, `H`, `[::ffff:1.2.3.4]`; dazu `[::ffff:1.2.3.4]` und `[::ffff:102:304]` mit gleicher kanonischer Form) und 20 Verstöße (leer mit `leer`; `http://h`, `h/x`, `h/`, `u@h`, `*.h`, `h*`, ` h`, `h x`, `h:0`, `h:65536`, `h:`, `h:+1`, `h:01`, `h:1x`, `[::1`, `[]`, `[1.2.3.4]`, `[fe80::1%25eth0]`, ein Name mit Byte ab 0x80 mit `host`), Schlüssel `FALLBACK_ALLOW_HOSTS` | Port `0`, `65536`, `01` oder leer angenommen; Zone angenommen; `[1.2.3.4]` angenommen; Schema angenommen; Platzhalter angenommen; `[::ffff:1.2.3.4]` abgewiesen (je Element rot) | andere |
| P13 (Nr. 6 a) | Eintrag von `Schreibfreigaben`: Methode, ein Leerzeichen, Muster; Ursachen `leer`, `trenner`, `methode`, `muster` in dieser Rangfolge | Menge 19: 5 Gutfälle (je mutierende Methode mit `/svc/Orders`) und 14 Verstöße (leer; `POST`; `/svc/x`; `GET`; `post /svc/x`; `GET /svc/x`; `HEAD /svc/x`; `OPTIONS /svc/x`; ` /svc/x`; `get x`; `POST  /svc/x`; `POST /svc/x ` mit Leerzeichen am Ende; `POST x`; `POST `), Soll-Ursache je Fall, Schlüssel `FALLBACK_WRITE_ALLOW`, `Eintrag` 1 | `muster` vor `methode` geprüft; lesende Methode angenommen; Kleinschreibung angenommen; zwei Leerzeichen angenommen (je Element rot) | andere |
| P14 (Nr. 6 b, f) | Muster: Syntax und Ausschlüsse; höchstens ein `*` je Segment (kein Zurücksetzen im Vergleich); kein Segment mit `*`, das ein Punkt-Segment deckte (sechs Muster, fünf Segmente: Nr. 6 b, e) | Menge 37 und Menge 169 (Klasse der Segmente mit `*`: `A*B`, A und B je null bis zwei Teile aus `.`, `%2e`, `%2E`, also 13 × 13; 6 Verstöße `.*`, `*.`, `%2e*`, `*%2e`, `%2E*`, `*%2E`, 163 Gutfälle, darunter `*`, `..*`, `.*.`, `%2e*%2e`; je Muster als Segment in `/a/<S>`, `/<S>/b` und `/a/<S>/**`); Menge 37: 9 Gutfälle (`/a`, `/a/b`, `/a/*`, `/a/**`, `/a/Orders(*)`, `/a/*b`, `/a/a*`, `/a/x%2Fy`, `/a/%2e%2ex`) und 28 Verstöße je Regel (kein `/` am Anfang; `/`; `//`; `/a//b`; abschließendes `/`; `?`; `#`; `\`; `,`; Byte 0x1F; Byte 0x7F; Byte ab 0x80; `%`; `%2`; `%G0`; `%2G`; `/a/*b*`; `/a/**b`; `/a/***`; `/**/a`; `/a/x**`; `/a/.`; `/a/..`; `/a/%2e`; `/a/%2E`; `/a/.%2e`; `/a/%2e%2E`; `/a/%2e./b`), Ursache `muster` | die Prüfung auf `.` und `..` entfällt; `%2E` nicht erkannt; zwei `*` im Segment erlaubt; `**` mitten erlaubt; `,` erlaubt; `%` ohne Hexziffern erlaubt; die Prüfung des Segments mit `*` entfällt, gilt nur für `.`, nur für `%2e` ohne `%2E`, nur vor oder nur hinter dem `*`, oder weist auch `..*`, `.*.` oder `*` allein ab (je Element rot) | andere |
| P17 (Nr. 7) | Eintrag von `ReplayKategorien`: genau `unsupported` oder `coverage_insufficient`, Schreibweise zählt | Menge 10: 2 Gutfälle und 8 Verstöße (leer mit `leer`; `Unsupported`, `UNSUPPORTED`, `none`, `coverage-insufficient`, `unsupported ` mit Leerzeichen am Ende, `ODR-E6003`, `all` mit `kategorie`), Schlüssel `HYBRID_REPLAY_FALLBACK` | Schreibweise ignoriert; `none` angenommen; Leerraum entfernt | andere |
| P18 (Nr. 8 a, b) | Fehler von `New`: je Schlüssel der erste Verstoß (ab 1), Reihenfolge der Schlüssel, `errors.Join`, `errors.As` je Fehler, keine Policy; die Syntax wird auch bei `Upstream` aus geprüft | Menge 19: Liste {3} allein × `Upstream` {aus, an} (6); je Liste zwei Verstöße an den Stellen 1 und 3 (3); ein gültiger Eintrag vor dem Verstoß (3); Paare von Listen (3); alle drei (1); je Liste zwei Verstöße an den Stellen 2 und 3 (3); Schleife über `errors.As` zählt die Fehler, prüft Schlüssel, `Eintrag` und Reihenfolge, `Policy` ist nil | der letzte statt des ersten Verstoßes; `Eintrag` ab 0; Abbruch nach dem ersten Schlüssel; Reihenfolge der Schlüssel vertauscht; Prüfung nur bei `Upstream` an (je Mutation rot) | andere |
| P19 (Nr. 8 b) | Der Text eines Fehlers nennt Schlüssel, Nummer und Ursache und nie den Wert eines Eintrags | Menge 5: je ein Verstoß mit dem Marker `GEHEIM-7f3a` im Wert für die Ursachen `host`, `trenner`, `methode`, `muster`, `kategorie`; `Error()` enthält Schlüssel, Nummer und Ursache, nicht den Marker | der Text trägt den Wert des Eintrags (`%q`) | andere |
| P20 (Nr. 8 c) | Zusammenhänge: `Upstream` an ohne `Hosts` ist `ohne-hosts` (Schlüssel `HYBRID_FALLBACK`, `Eintrag` 0), entfällt bei ungültigem Eintrag in `Hosts`; `Upstream` aus mit gesetzten Listen und eine Freigabe ohne Host sind keine Fehler | Menge 8: `Upstream` {aus, an} × `Hosts` {leer, gültig, ungültig} (6); aus mit `Schreibfreigaben` und leeren `Hosts` (1); an mit gültigen `Hosts` und einer `Schreibfreigabe` ohne Host im Eintrag (1); Soll aus der Tabelle des Vertrags | `ohne-hosts` auch bei aus; fehlt bei an; zusätzlich bei ungültigem Eintrag; eine Freigabe ohne Host als Fehler (je Mutation rot) | andere |
| P21 (Nr. 8 d, Nr. 9 c) | `New` ruft nichts und liest keine Umgebung; kein Zustand, keine Sperre, keine Goroutine, keine Variable auf Paketebene; Importe nur `model` und die Standardbibliothek | Test über den Quelltext des Pakets (`go/parser`): Menge der Importe ohne `os`, `sync`, `time`, `context`, `net` (außer `net/url` und `net/netip`) und ohne ein Paket des Hexagons außer `model`; kein `go`-Statement; keine `var`-Deklaration auf Paketebene; Methoden nur auf `Policy` und `KonfigFehler` | Import eines Ports; `sync.Mutex`; `os.Getenv`; `net.LookupHost`; `go func`; eine Variable auf Paketebene (je Mutation rot) | andere |

Die Tabelle hat neun Zeilen, alle der Klasse `andere`: die Policy hat keine Nebenläufigkeit (§6), und P21 fängt Zustand, Sperre und Goroutine im Quelltext, nicht mit dem Detektor.

Akzeptierte Negative von P21 (Architect, 2026-11-08; die Tabelle ist ihre Adresse): der Quelltext-Test liest nur die Dateien im Wurzelverzeichnis des Pakets, kennt kein `init`-Verbot, erlaubt `reflect` und `runtime` und hält feste Untergrenzen der Zählprüfung (3 Dateien, 4 Importe, 8 Funktionen). Grund: Zustand braucht eine Variable auf Paketebene oder eine Methode, und beides fängt der Test; ein `init` ohne Variable hält keinen Zustand der Policy; `reflect` und `runtime` legen keinen an; ein Unterverzeichnis ist ein eigenes Paket, dessen Import-Richtung `make a-check` hält (`ADR-0002`). Folge: ein `init` mit Wirkung auf ein importiertes Paket bliebe unbemerkt, und mehr Dateien ändern die Untergrenzen nicht; zu wenige gelesene Dateien meldet der Test als eigenen Fehler, nicht stumm. Der Test sagt das in einem Kommentar der Klasse *Grenze* (`AGENTS.md` §3.7); eine eigene Folgepflicht gibt es nicht.

**Umsetzung und Mutationen** (Implementer, 2026-11-08): Das Paket `src/hexagon/services/hybrid/policy` hält `policy.go` (`Werte`, `Policy`, `New`, `*KonfigFehler`, Kategorien), `host.go` (Eintrag von `Hosts`) und `freigabe.go` (Eintrag von `Schreibfreigaben` und Muster); die Tests liegen daneben als externes Paket `policy_test` (`host_test.go`, `freigabe_test.go`, `policy_test.go`, `quelltext_test.go`, `hilfen_test.go`; `export_test.go` gibt die geparsten Formen als einfache Werte heraus, weil `testpackage` im Lint-Profil nur diese Datei ausnimmt). Die Policy importiert allein `errors`, `fmt`, `net/netip` und `strings`. Je Zeile der Tabelle ein Test der obersten Ebene; jede Mutation läuft mit `sh tools/harness/mutationslauf.sh` (Image `odata-recorder:test`, `--memory=4g`, ein Prozess je Mutation, Klasse `andere`, deshalb keine Rate: der Test ist deterministisch und läuft in jedem Lauf rot oder grün). Von 151 Mutationen (140 der ersten Runde, 11 der Fix-Runde zu Review M1 und I1) sind 150 rot im benannten Test, eine ist äquivalent (`host-grenze-nach-9`: `:` erreicht die Namensprüfung nie, weil der erste `:` den Namen vom Port trennt); drei Mutationen liefen im ersten Lauf nicht übersetzbar (unbenutzter Import oder Variable) und wurden mit übersetzbarer Form wiederholt und rot gesehen; eine Mutation (`c < 0x20` statt `c < 0x21` im Muster) war im Lauf nur im Paket rot, nicht im benannten Test, bis die Tabelle von `TestMusterEintraege` die Fälle mit Leerzeichen bekam. Die Liste der Mutationen liegt nicht im Repo (Arbeitsverzeichnis des Laufs); der Verifier schreibt sie aus der Spalte *Mutationen* neu oder misst selbst.

| Zeile | Zusage | Mutationen (Zahl, Beispiele) | roter Test |
|---|---|---|---|
| P11 | Eintrag von `Hosts` (Nr. 5 a, b) | 36: `[::ffff:1.2.3.4]` abgewiesen; Port `0`, `01`, `65536`, leer, Überlauf, Ziffern ober- und unterhalb; Zone, `[1.2.3.4]`, Rest hinter `]`, Name leer mit Port; Platzhalter, Pfad, `@`, Leerraum, Byte ab `0x80` im Namen; `_`, `-`, `.` abgewiesen; keine Kleinschreibung, IPv6 nicht kanonisch; zwölf Grenzen von `A-Z`, `a-z`, `0-9` | `TestHostEintraege` (24), `TestHostGenerator` (11, eine äquivalent) |
| P13 | Eintrag von `Schreibfreigaben` (Nr. 6 a) | 14: je mutierende Methode abgewiesen (5), `GET`, `HEAD`, `OPTIONS` angenommen, Kleinschreibung, zwei Leerzeichen, letztes statt erstes Leerzeichen, `muster` vor `methode`, `methode` vor `trenner`, `leer` als `trenner` | `TestFreigabeEintraege` |
| P14 | Muster (Nr. 6 b) | 40: Prüfung eines Segments mit `*` auf ein Punkt-Segment entfällt, erkennt nur `.`, nur `%2e` oder nur `%2E`, prüft nur den Anfang oder nur das Ende, weist `*` allein, `..*`, `.*.` oder `a.*` ab (10); Leerzeichen, `0x1F`, `0x7F`, Byte ab `0x80`, `?`, `#`, `\`, `,` erlaubt; `%` ungeprüft, erste und zweite Ziffer, sechs Grenzen der Hexziffern; leeres Segment, zwei `*`, `**` mitten, `**b`, `x**`; Prüfung auf `.`, `%2e`, `%2E`, `.` allein, `..` allein entfällt; Seiten des `*` vertauscht; ohne `/` am Anfang | `TestMusterEintraege` (23), `TestFreigabeGenerator` (7), `TestMusterStern` (10) |
| P17 | `ReplayKategorien` (Nr. 7) | 7: Flags vertauscht, `coverage_insufficient` abgewiesen, `leer` als `kategorie`, `none`, Schreibweise, Leerraum, Code und `all` angenommen | `TestKategorien` |
| P18 | Fehler von `New` (Nr. 8 a, b) | 17: `Eintrag` ab 0 (3 Listen), letzter statt erster Verstoß (3), Abbruch nach `Hosts`, `Schreibfreigaben`, vor dem Zusammenhang, Reihenfolge der Schlüssel (2), nur bei `Upstream` an (3), `errors.Join` nur des ersten, Policy trotz Fehler | `TestNewFehler` (15), `TestZusammenhaenge` (2) |
| P19 | Text ohne Wert (Nr. 8 b) | 6: Wert an die Ursache gehängt (3 Listen), Text ohne Nummer, ohne Schlüssel, ohne Ursache | `TestFehlerText` |
| P20 | Zusammenhänge (Nr. 8 c) | 7: `ohne-hosts` auch bei aus, fehlt bei an, bei ungültigem Eintrag, bei Freigabe ohne Host, `Eintrag` 1, falscher Schlüssel, falsche Ursache; dazu Reihenfolge nach `HYBRID_REPLAY_FALLBACK` | `TestZusammenhaenge` |
| P21 | Quelltext (Nr. 8 d, 9 c) | 14 + 8: Import von `os`, `sync`, `sync/atomic`, `time`, `context`, `net`, `net/http`, `os/exec`, einem Port, einem Dienst; `go`-Statement; `var` und `var (…)` auf Paketebene; Methode auf fremdem Typ; die Liste der verbotenen Importe selbst (`os`, `time`, `context`, `net`, `net/`, `sync/`, Fremdbibliothek, nur `model`) | `TestQuelltext` (14), `TestImportErlaubt` (8) |
| — | Schlüssel und Ursachen im Wortlaut | 2: ein Buchstabe im Schlüssel, ein Buchstabe in einer Ursache | `TestKonstanten` |

Mengen (Schritt 14 von [implement-slice](../../../../.claude/commands/implement-slice.md), Generator und Orakel unabhängig vom Code, Zählprüfung der Menge im Test): `Hosts` 4 Grundformen (`h.example:8080`, `[2001:db8::1]:8080`, `[::1]`, `H-1_x`) × jedes Byte 0 bis 255 an jeder Stelle einfügen, ersetzen, löschen (je Grundform `(n+1)·256 + n·256 + n`; zusammen 22 570 Einträge), Orakel: reguläre Ausdrücke und `net.ParseIP`; `Schreibfreigaben` 5 Grundformen (`POST /svc/Orders(*)/**`, `DELETE /a/*b/x%2Fy`, `MERGE /a/%2e%2ex/*`, `PATCH /a/b`, `PUT /a/.x`) × dieselben drei Eingriffe (zusammen 40 781 Einträge), Orakel: reguläre Ausdrücke über die Byte-Menge und Regeln je Segment, die Regel zu Punkt-Segmenten aus dem Wortlaut von Nr. 6 b und 6 e (ein Segment mit `*` ist ungültig genau dann, wenn eine Einsetzung des `*` durch Punkte nach dem Ersetzen von `%2e` und `%2E` `.` oder `..` ergibt; das ganze Segment `*` bleibt gültig), nicht aus `parseSegment`; `ReplayKategorien` 2 Grundformen × dieselben Eingriffe (zusammen 16 928 Einträge), Orakel: Gleichheit. Die Orakel stimmen mit dem Code in jedem erzeugten Eintrag überein (`net.ParseIP` und `netip.ParseAddr` weichen in keinem der 22 570 Einträge ab). Die Klasse der Segmente mit `*` (Review M1, Menge nach dem Wortlaut von Nr. 6 b und 6 e): `A*B`, A und B je null bis zwei Teile aus `.`, `%2e`, `%2E` (1 + 3 + 9 = 13 Seiten, 13 × 13 = 169 Segmente), jedes als Segment in `/a/<S>`, `/<S>/b` und `/a/<S>/**` (507 Muster, über `ParseMuster` und `New` mit Ursache `muster`); das Orakel ist die Einsetzung des `*` durch einen oder zwei Punkte (ungültig, wenn A und B nicht beide leer sind und das Ergebnis nach dem Ersetzen `.` oder `..` ist); Zählprüfung 169 Segmente, 6 Verstöße, 163 Gutfälle; Handfälle `/svc/.*`, `/svc/*.`, `/svc/%2e*`, `/svc/*%2e`, `/svc/%2E*`, `/svc/*%2E` (Verstöße) und `*`, `..*`, `.*.`, `%2e*%2e`, `a.*` (gültig). Eine Mutation an einem Element, das der Befund nicht nannte, in der Fix-Runde rot gesehen: „prüft nur das Ende" (rot an `.*`), „weist `.*.` ab" und „weist `a.*` ab" (rot an Gutfällen). Die handgeschriebenen Tabellen (30, 19, 37, 10 Fälle) bleiben als Goldens daneben. Die Test-Eingaben mit Bytes über `0x7F` (`h\xc3\xa4`, `/a/\x80`) stehen als `\x`-Folgen, keine Folge aus Backslash, `u` und vier Hexziffern; die Dateien sind mit `LC_ALL=C grep -P` auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

Entschieden und gezählt (Architect, 2026-11-08): `[::ffff:1.2.3.4]` ist eine gültige IPv6-Adresse (`SPEC-261` Nr. 5 a) und hat in P11 einen eigenen Fall; der Eintrag gibt nur den Host frei, den er nennt, die Allowlist schützt den Namen, nicht die Adresse (Nr. 5 e). Die Go-Namen `Schluessel*` und `Ursache*` gegengelesen: der Vertrag bindet die Werte (Signatur, Nr. 8 b und c), die Namen bleiben Entwurf des Implementers und gelten, ohne Eintrag in `SPEC-261`.

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
7. Replay-Fallback: für alle bekannten Methoden, unabhängig vom Hauptschalter — Nr. 4 b; Frage an den User in §6 von `slice-fallback-policy-entscheidung`.
8. Tabelle, Raster und Nachzug: die Tabelle (24 Zeilen, nach dem Schnitt neun hier und 16 in `slice-fallback-policy-entscheidung`), §6 und die genannten Stellen.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen und am 2026-11-08 nach dem Schnitt erneut: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstelle): *von diesem Slice an* `slice-bootstrap-m4` (Lesen der vier Schlüssel und Aufruf der Syntax-Prüfung): angenommen, §1 *Übernimmt* und DoD Punkt 1 von `slice-bootstrap-m4` tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* dort trifft sie. *Von diesem Slice an* `slice-fallback-policy-entscheidung` (Entscheidung, Vergleich von Host und Pfad, Replay-Fallback): angenommen, §1 *Übernimmt* dort trägt sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* dort trifft sie. Die Sendungen aus `SPEC-261` Nr. 10 (Aufrufer) gibt `slice-fallback-policy-entscheidung` weiter, nicht dieser Slice. *An diesen Slice:* keine.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein Slice muss in `done/` liegen. Die Spezifikation trägt den Vertrag samt Randformen und der Syntax von Pfadmuster und Host-Vergleich, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Code-Abhängigkeit: keine.** Die Policy liest nur `model` (die Codes `CodeE5001` und `CodeE6001` bis `CodeE6003` stehen seit `slice-fehlermodell-konfiguration` in `src/hexagon/model/fehler.go`) und typisierte Werte; sie ruft keinen Port.

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`:** die Policy braucht weder `MockUseCase` noch `ReplayUseCase` noch `UpstreamClient` und kann auf dem heutigen Stand von `src/` entstehen. Der Trigger der Welle (`welle-hybrid` §2: `welle-mock` in `done/`, M3) ist seit der Closure von `welle-mock` eingetreten; der Start ist keine Umplanung mehr.

**Stand der Bedingungen beim Start** (Planner, 2026-11-08, `main` = `686612a2`): `welle-mock` liegt in `done/` (M3 erreicht), `slice-fehlermodell-konfiguration` liefert die Codes `CodeE5001` und `CodeE6001` bis `CodeE6003`. In `in-progress/` laufen parallel `slice-etag-versionen` und `slice-admin-lebenszyklus`; keiner berührt `src/hexagon/services/hybrid/`, und dieser Slice ändert keinen Pfad der Gate-Kette (Übergangsregel 7, §6), also kein Konflikt. Die Verträge der Policy schreibt der Architect im Zweig (Aufträge 1 bis 8 in §3), bevor der erste Code-Commit entsteht.

**Größenmarke und Schnitt** (Planner, 2026-11-08, `SPEC-128` Nr. 1; ersetzt die vorläufige Marke 2600 der Prosa-Form, die auf angenommenen 20 Zeilen beruhte): Der Plan stand mit der Tabelle des Architect bei 24 Zeilen (*a* = 0, *b* = 24, *f* = 0) und *p* = 500: *S* = 80·24 + 500 = 2420, *M* = 3300, *S* ≥ 2223, also Schnittfrage. *p* = 500 prüft der Planner an den Teilen der Policy (eigene Schätzung: Parser für Host, Routenfreigabe und Muster rund 220 Zeilen, `New` und `*KonfigFehler` rund 110, Entscheidung samt Vergleich von Host und Muster rund 170) und lässt die Erwartung stehen: sie liegt in der Größenordnung, und die Formel überschätzt Tests. **Entscheidung: geschnitten**, vor dem Code, an der Naht, die der Architect nennt: Prüfung der Einträge und `New` (dieser Slice) gegen die Entscheidung (`slice-fallback-policy-entscheidung`). Die Teile teilen die geparste Form der Einträge, aber nicht den Mechanismus: dieser Slice definiert sie und prüft ihre Syntax, der zweite liest sie und schreibt keinen zweiten Parser (Code-Abhängigkeit in der Reihenfolge, kein Code in beiden); der Test der Zusage „Listen sind Mengen“ (P01) liegt wegen seines Fängers im zweiten Teil, hier steht die Zeile P01a als akzeptiertes Negativ. Dieser Slice: 9 Zeilen (P01a, P11, P13, P14, P17 bis P21), *a* = 0, *b* = 9, *f* = 0, *p* = 300: *S* = 85·0 + 80·9 + 200·0 + 300 = 1020, *M* = 1400, keine Schnittfrage (*S* < 2223). Der zweite Teil: 16 Zeilen, *p* = 200, *S* = 1480, *M* = 2000. Die Summe der Schätzungen (2500) liegt um die Zeile P01a (80) über der des ungeschnittenen Plans (2420), die der Marken (3400) um 100 über 3300. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Basis der Messung: `9310c219` (der Commit, der den Vertrag `SPEC-261` abschließt; er ändert unter dem Pfad des Blocks nichts).

**Teil des Architect zur Größenmarke** (2026-11-08, `SPEC-128` Nr. 3 k; `produktcode` ist die Erwartung für Auswahl, Host, Muster, Prüfung und Fehler des ungeschnittenen Plans, hier auf den Anteil dieses Slice gesetzt) **und Teil des Planner** (`basis`, `schaetzung`, `marke` nach der Tabelle in §3, 9 Zeilen, alle `andere`, keine Fakes):

```groessenmarke
pfad: src/hexagon/services/hybrid
produktcode: 300
fakes: 0
basis: 9310c219
schaetzung: 1020
marke: 1400
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: 1400** (siehe *Größenmarke und Schnitt* oben).
- `in-progress` → `open` (blockiert — Carveout?): eine Zelle des Rasters (Pfadmuster, Host-Syntax) hat keinen Ort in der Spezifikation, und der Architect entscheidet sie nicht vor dem Code (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): jede Zelle nennt ihren Ort in `SPEC-261` (Architect, 2026-11-08); der Fänger steht in der Tabelle *Zusage · Fänger* in §3. Die Operationen der Entscheidung und des Vergleichs führt `slice-fallback-policy-entscheidung`.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Prüfung eines Eintrags von `FALLBACK_WRITE_ALLOW` (Methode und Pfadmuster) | kein Leerzeichen, kein Muster: Nr. 6 a | der leere Eintrag, das leere Muster: Nr. 6 a | unbekannte oder lesende Methode, Zeichen außerhalb der Menge, zwei `*` in einem Segment, ein Segment mit `*`, das ein Punkt-Segment deckte (`.*`, `%2e*`): Nr. 6 a, Nr. 6 b, Nr. 6 e | gilt nicht: reine Funktion ohne `ctx` | Doppelte Einträge sind eine Menge: Nr. 1 a |
| Prüfung eines Eintrags von `FALLBACK_ALLOW_HOSTS` | Port ohne Name, Name ohne Klammern bei IPv6: Nr. 5 a | der leere Eintrag: Nr. 5 a | Schema, Pfad, Platzhalter, Zone, Port außerhalb 1 bis 65535: Nr. 5 a | gilt nicht: reine Funktion ohne `ctx` | Nr. 1 a |
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

Der Modus des Hauptschalters in der Entscheidung (Zeile der Zusatztabelle des ungeschnittenen Plans) steht in `slice-fallback-policy-entscheidung`.

- Pfadmuster und Host-Vergleich haben in `SPEC-062` und `SPEC-080` keine Syntax — **Ausgang:** entfallen: entschieden in `SPEC-261` Nr. 5 und Nr. 6 (Architect, 2026-11-08).
- Methoden außerhalb von GET, HEAD, POST, PUT, PATCH und DELETE (Auftrag 4) und der Replay-Fallback für mutierende Methoden (Auftrag 7) haben keinen Ort — **Ausgang:** entfallen: entschieden in `SPEC-261` Nr. 2 und Nr. 4 b; die Frage, ob der Replay-Fallback mutierender Methoden gewollt ist, steht in §6 von `slice-fallback-policy-entscheidung`.
- Offene Fragen an den User (Replay-Fallback auch für mutierende Methoden, bis zur Antwort gilt `SPEC-261` Nr. 4 b; `ODR-E6006` als Auslöser, bis zur Antwort nein; Lücke des Lastenhefts zum Replay-Fallback, [Abgleichsbericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §8 Nr. 2) — **Ausgang:** weiter offen → §6 von `slice-fallback-policy-entscheidung` (Gegenstand der Entscheidung, nicht der Prüfung der Einträge); Adresse: User.
- Eine Policy, die in einem Zustand fehlt, den die Matrix nicht nennt, erlaubt nie einen Fallback — **Ausgang:** entfallen für diesen Slice: Gegenstand der Entscheidung, Fänger P03, P06 und P10 in `slice-fallback-policy-entscheidung`.
- Wechselwirkung (`SPEC-261` Nr. 4 e) und die Sendungen aus Nr. 10 an `slice-hybrid-service` und `slice-bootstrap-m4` (`AGENTS.md` §3.13) — **Ausgang:** entfallen für diesen Slice: sie liegen bei `slice-fallback-policy-entscheidung` (Geber); die Nehmer tragen sie seit dem Schnitt (Planner, 2026-11-08, Commit `c484eefd` auf `main`).
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): die Policy hat keine Nebenläufigkeit, der Detektor hat nichts zu melden; P21 fängt Zustand, Sperre und Goroutine im Quelltext — **Ausgang:** entfallen: reine Funktion ohne Zustand, Zählprüfung der Fälle statt eines Laufs unter dem Detektor.
- Review-Befund M1 (Muster `/svc/.*`, `*.`, `%2e*`, `*%2e` und ihre `%2E`-Formen deckten ein Punkt-Segment): **Ausgang:** entfallen: in der Syntax geschlossen, `SPEC-261` Nr. 6 b, c, e (Architect, 2026-11-08); Fänger P14 (Menge 169); die Marke bleibt, weil P14 eine Zeile bleibt (*S* = 1020, *M* = 1400, Messung 1221 vor dem Fix).
- Größe: Schnitt vor dem Code entschieden und der Block gesetzt (§4: *S* = 1020, *M* = 1400) — **Ausgang:** entfallen: Planner, 2026-11-08, vor dem ersten Code-Commit.
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
