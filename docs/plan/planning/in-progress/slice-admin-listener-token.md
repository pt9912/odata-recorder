# Slice slice-admin-listener-token: Admin-Listener und Token-Authentisierung, Teil 1 — die Token-Datei

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

**Berührte Spec-Stellen:** `LH-FA-13.a` (Token-Datei: fehlt sie, erzeugt der Prozess ein Token mit mindestens 256 Bit und Rechten `0600`; eine ungültige Datei verhindert den Start mit `ODR-E2001`; „nie in einem Log“) · `LH-QA-04.a` · `SPEC-056` (Zeile `ADMIN_TOKEN_FILE`) · `SPEC-239` (Token-Datei, `admin.LadeToken`, neu, Architect) · `SPEC-110` Nr. 12 und 13 (Muster der Meldung, synchron ohne Kontext) · `SPEC-238` Nr. 13 (Geheimnisse) · `architecture.md` §5 (Sicherheit). Der Vertrag des ganzen Rahmens (`SPEC-238`) und der Code `ODR-E2102` (`SPEC-240`) stehen seit dem Architect in diesem Zweig; die Teile 2 bis 6 tragen sie (siehe *Herkunft*).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-admin-listener-token`, Worktree `/Development/odata-recorder-wt-admin-listener-token`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

**Herkunft:** Dieser Slice ist seit dem Schnitt vor dem Code (Planner, 2026-11-07, nach dem Vorschlag des Architect, Rückgabe R1, `SPEC-128` Nr. 1, `AGENTS.md` §3.14) der **Teil 1** von sechs, die Token-Datei (der vorab benannte Teil). Die Teile 2 bis 6 sind Pläne in `open/` und folgen nacheinander: [slice-admin-rahmen](../open/slice-admin-rahmen.md) (Konfiguration, Routen, Zuordnung, Code `ODR-E2102`), [slice-admin-authentisierung](../open/slice-admin-authentisierung.md), [slice-admin-lebenszyklus](../open/slice-admin-lebenszyklus.md), [slice-admin-fehlerabbildung](../open/slice-admin-fehlerabbildung.md) und [slice-admin-abweisen-audit](../open/slice-admin-abweisen-audit.md). Dieser Plan trägt die Zeilen T1, T2, T4, T5 und T7 bis T26 der Tabelle des Architect (24 von 102).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem** Abschnitt des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Paket `src/adapters/driving/admin` lädt das Token des Admin-Listeners aus `ADMIN_TOKEN_FILE` (`admin.LadeToken`, `SPEC-239`): fehlt die Datei, erzeugt der Prozess ein Token mit 64 Zeichen `0-9a-f` (256 Bit aus `crypto/rand`) und Rechten `0600` von Anfang an (Hilfsdatei und `Link`, vollständig oder gar nicht, von zwei Prozessen gewinnt einer); eine ungültige Datei verhindert den Start mit `ODR-E2001 ADMIN_TOKEN_FILE: <Grund>` mit einem von sieben Gründen (nicht lesbar, Verzeichnis, keine reguläre Datei, Gruppen- oder Weltrechte, nicht anlegbar, mehr als eine Zeile, weniger als 32 Zeichen), ohne Inhalt und Pfad zu nennen; ein Symlink auf eine reguläre Datei (gemountetes Secret) ist gültig; eine vorhandene Datei wird nie geändert; nichts erreicht ein Log, `stdout` oder `stderr`.

**Übernahmen des ungeschnittenen Plans.** Die Sendungen, die der Plan vor dem Schnitt trug, liegen seit dem Schnitt bei den Teilen, die den Code liefern; die Adresse der Geber (`slice-session-use-case`, `slice-inspection-use-case`) bleibt `slice-admin-listener-token`, die Teile tragen die DoD (`AGENTS.md` §3.13):

- der Code der HTTP-Antwort auf `model.ErrUngueltig` und auf einen ungültigen Seitenparameter (Entscheidung des Architect: `ODR-E2102` mit 400, `SPEC-238` Nr. 6, `SPEC-240`) — die Zeile der Codetabelle in `slice-admin-rahmen`, die Abbildung in `slice-admin-fehlerabbildung`;
- das Zählen der Fehler mit Code für den Admin-Listener (`SPEC-230` Nr. 8; eine Stelle schreibt das Fehlerobjekt und zählt, neun Schreibstellen) — `slice-admin-fehlerabbildung`.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Alles außer der Token-Datei: Listener, Konfiguration, Routen und Zuordnung (`slice-admin-rahmen`), Authentisierung und `GET /healthz` (`slice-admin-authentisierung`), `Serve`, `Anhalten`, `Schliessen` (`slice-admin-lebenszyklus`), Fehlerabbildung und Zählen (`slice-admin-fehlerabbildung`), Audit des Abweisens (`slice-admin-abweisen-audit`) — die Token-Datei teilt keinen Mechanismus mit dem Rahmen (sie ruft keinen Port und kein HTTP) und ist einzeln lieferbar.
- Die Endpunkte außer `GET /healthz` — `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces`.
- TLS, `ADMIN_TLS_*`, `ADMIN_ALLOW_PLAINTEXT` und die Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` — `slice-admin-tls`.
- Das Lesen von `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` aus der Umgebung, der Default `<DATA_DIR>/admin.token`, der Aufruf von `LadeToken` beim Start, das Einsortieren der Meldung unter den Schlüssel und das Aufheben der Ablehnung in `SPEC-110` Nr. 2 — `slice-bootstrap-m6`; ein gelesener, aber nirgends verdrahteter Schlüssel wäre eine still wirkungslose Sicherheitseinstellung. `LadeToken` nimmt einen Pfad.
- Token-Rotation, mehrere Token, Benutzer, Rollen, OIDC — Out-of-Scope der Welle (§6): die Spezifikation kennt ein Token.

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

- [ ] Token-Datei (`SPEC-239`): fehlt sie, erzeugt der Prozess ein Token (mindestens 256 Bit, Rechte `0600`, nie in einem Log); eine nicht lesbare Datei, weniger als 32 Zeichen, Gruppen- oder Weltrechte, ein Verzeichnis, keine reguläre Datei nach Auflösung eines Symlinks, mehr als eine Zeile und ein nicht beschreibbares Verzeichnis der fehlenden Datei sind `ODR-E2001` mit dem Grund aus der Rangfolge der sieben Gründe; ein Symlink auf eine reguläre Datei (gemountetes Secret) ist gültig; eine vorhandene Datei bleibt unverändert; von zwei gleichzeitigen Aufrufen auf eine fehlende Datei gewinnt einer.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Für diesen Teil: kein neuer Code und kein Schlüssel (`ODR-E2001` besteht; `ADMIN_TOKEN_FILE` liest erst `slice-bootstrap-m6`); das Binary bleibt unverändert (kein Aufruf von `admin` aus `bootstrap`, `src/bootstrap` ohne Änderung), in §7 vermerkt. Der Code `ODR-E2102` kommt mit `slice-admin-rahmen`, samt Ausnahme in `harness/katalog-ausnahmen.txt`.
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
| `src/adapters/driving/admin/tokendatei.go` | neu | `LadeToken(pfad)`, `TokenFehler`, Hilfsdatei und `Link`, sieben Gründe in der Rangfolge von `SPEC-239` Nr. 6 |
| `src/adapters/driving/admin/` (Tests) | neu | Tests der Token-Datei in einem Testverzeichnis mit gesetzten Rechten (`os.Chmod`, unabhängig von der Umask); AST-Tests über das Paket; der Adapter importiert weder Adapter noch Service (`make a-check`) |
| `spec/spezifikation.md` | geliefert (Architect, in diesem Zweig) | `SPEC-238` bis `SPEC-240` für den ganzen Rahmen, Nachzug `LH-FA-13.a`, `SPEC-005`, `SPEC-064`, `SPEC-109`, `SPEC-230` Nr. 8, `SPEC-234` Nr. 2, die *Ungeprüft*-Zeilen zu `model.ErrUngueltig`, Historie; dieser Teil schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem 2026-11-07 (Architect) mit dem Vertrag in der Spezifikation (`SPEC-238` bis `SPEC-240`) vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind entschieden, der Implementer entscheidet keine im Code. Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest. Dieser Teil führt die Zeilen T des Architect im Wortlaut; `Nr.` ist eine Kennung, keine Zählung (T Token-Datei); die Lücken T3 und T6 sind Zeilen, die in T2 und T5 aufgingen (beide Seiten derselben Grenze sind ein Fänger). Die Zeilen der übrigen Nummern des Vertrags stehen in den Plänen der Teile 2 bis 6.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| T1 | (`SPEC-239` Nr. 1) Eine gültige Datei liefert den Rest nach dem Entfernen von ASCII-Leerraum an beiden Enden; innerer Leerraum, Schreibweise und Bytes außerhalb von ASCII bleiben | Generator: Kern (32 Bytes, 64 Hex-Zeichen, mit innerem Leerzeichen, 32 Bytes UTF-8, 100 Bytes) × Rand (keiner, `\n`, `\r\n`, Leerzeichen, `\t`, `\v`, `\f`, `\r`) links × rechts; Orakel der erwartete Kern als Literal je Fall, nicht aus `bytes.Trim`; Zählprüfung 5 × 8 × 8 | Trim nur von `\n`; Unicode-Trim (U+00A0 am Rand muss bleiben) | andere |
| T2 | (Nr. 1 d) Die Rechte `0400`, `0500`, `0600`, `0700` sind gültig; jedes Bit von Gruppe und Welt ist der Verstoß `Gruppen- oder Weltrechte` | Schleife über die 4 gültigen Modi und über die sechs Bits (040, 020, 010, 004, 002, 001) einzeln sowie 0644, 0660, 0777, mit `os.Chmod` (unabhängig von der Umask); Orakel Literal; die Datei danach unverändert | Prüfung auf genau `0600`; Maske `0o070` statt `0o077`; `0o027` | andere |
| T4 | (Nr. 1 d) Es zählen die Rechte der aufgelösten Datei, nicht die des Symlinks | Symlink auf Datei 0644 ist Verstoß; Symlink auf Datei 0400 gültig; Kette aus 2 Stufen | `Lstat` statt `Stat` | andere |
| T5 | (Nr. 1 g) Weniger als 32 Bytes nach dem Entfernen sind `weniger als 32 Zeichen`, 32 und 33 sind gültig; Leerraum am Rand zählt nicht, innerer zählt; eine leere Datei und eine Datei nur aus Leerraum werden nie akzeptiert | Fälle 31, 32, 33 Bytes rein; 31 + `\n`; 30 + 2 Leerzeichen am Rand; 31 mit einem inneren Leerzeichen (32, gültig); Menge (0 Bytes; `\n`; 31 Leerzeichen; 40 Leerzeichen; 40 Tabs; `\r\n\r\n`) | `<= 32`; Länge vor dem Entfernen; leeres Token zugelassen | andere |
| T7 | (Nr. 1 f) Ein `\n` oder `\r` im Rest ist `mehr als eine Zeile`; nur abschließender Umbruch und umgebender Leerraum werden entfernt | Verstöße `a\nb`, `a\r\nb`, `a\rb`, `a\n\nb` (jeder Teil 40 Bytes) ; gültig: Token + `\n\n`, `\n` + Token, Token + `\r\n`, Leerzeichen + Token + `\n` | nur `\n` geprüft | andere |
| T8 | (Nr. 1 b) Ein Verzeichnis, auch als Ziel eines Symlinks, ist `ist ein Verzeichnis` | 2 Fälle | `IsDir` auf `Lstat` | andere |
| T9 | (Nr. 1 c) Keine reguläre Datei ist `keine reguläre Datei`, ohne zu blockieren: FIFO, `/dev/null`, Symlink auf `/dev/null`, Symlink ohne Ziel, Unix-Socket | Schleife über 5 Formen; der Aufruf läuft in einer Goroutine mit Zeitgrenze (ein Hängen ist rot); Orakel Literal | `ReadFile` vor der Prüfung der Art (der FIFO blockiert) | andere |
| T10 | (Nr. 1 a) Ein Fehler von `Stat` außer „fehlt“ ist `Datei nicht lesbar` und erzeugt nichts | Fälle: ein Pfadelement ist eine reguläre Datei; Symlinkschleife; NUL im Pfad; danach entstand keine Datei | jeder Fehler von `Stat` als „fehlt“ gelesen (es würde erzeugt) | andere |
| T11 | (Nr. 1, 6) Bei mehreren Verstößen am selben Pfad gilt der kleinste Rang | Zusatztabelle in §6 (6 Zeilen): Pfad × erwarteter Grund als Literal | Reihenfolge von Rechten und Zeilen vertauscht | andere |
| T12 | (Nr. 1) Ein leerer `pfad` ist `model.ErrUngueltig`, ein relativer gilt zum Arbeitsverzeichnis | Fälle: `""` (kein `TokenFehler`); `rel/x` nach Wechsel des Arbeitsverzeichnisses in ein Testverzeichnis | leerer Pfad als „fehlt“ gelesen | andere |
| T13 | (Nr. 1) Ein Symlink auf eine reguläre Datei mit gültigem Inhalt (gemountetes Secret) und eine Kette von Symlinks sind gültig; die Rückgabe ist der Inhalt des Ziels | Fälle: ein Symlink, Kette 2, Kette 3, Ziel `0400` | `Lstat` statt `Stat` | andere |
| T14 | (Nr. 3 a) Das erzeugte Token hat 64 Zeichen `0-9a-f`; die Datei ist Token + `\n`; die Rückgabe das Token ohne `\n` | 100 Läufe in je neuem Verzeichnis: Muster `^[0-9a-f]{64}$`, alle 100 verschieden | 63 Zeichen; Großbuchstaben; ein fester Wert | andere |
| T15 | (Nr. 3 a) Die Zufallsquelle ist `crypto/rand`, nicht `math/rand` | AST-Test über die Importe des Pakets ohne `_test.go`: `crypto/rand` ja, `math/rand` und `math/rand/v2` nein | Import `math/rand` | andere |
| T16 | (Nr. 3 b) Die Datei trägt von Anfang an keine weiteren Rechte als `0600` und am Ende genau `0600`, bei jeder Umask | Schleife über Umask 000, 022, 077, 0277 (`syscall.Umask`, wiederhergestellt, nicht parallel): `Perm()` gleich 0600; Haltepunkt nach dem Anlegen der Hilfsdatei: `Perm() & 077` gleich 0 und Größe 0 | Anlegen mit 0644 und spätes `Chmod`; `Chmod` entfernt (Umask 0277 ergibt 0400) | andere |
| T17 | (Nr. 3 b bis d) Die Datei erscheint vollständig oder gar nicht | Haltepunkt (Parameter von `lade`) nach dem Schreiben, vor dem `Link`: `pfad` fehlt, die Hilfsdatei hat 65 Bytes; nach der Rückkehr hat `pfad` 65 Bytes | direkt in `pfad` geschrieben | andere |
| T18 | (Nr. 3 d; Nr. 4) Von zwei Aufrufen auf eine fehlende Datei gewinnt einer, beide liefern dasselbe Token, im Verzeichnis bleibt allein `pfad` | Haltepunkt nach „fehlt“ und vor dem `Link` ruft einen inneren `lade` zu Ende: die Rückgaben sind gleich, die Datei trägt sie, genau eine Datei im Verzeichnis | `os.Rename` statt `os.Link`; `O_TRUNC` auf `pfad` | nebenläufig |
| T19 | (Nr. 3 d; Nr. 4) Gleichzeitige Aufrufe auf eine fehlende Datei erzeugen genau eine Datei | 16 Goroutinen in 20 Läufen (neues Verzeichnis je Lauf) unter dem Detektor: alle Rückgaben gleich, genau eine Datei, `0600`; Rate je Prozess im Plan (ein deterministischer Fänger ist T18) | wie T18 | nebenläufig |
| T20 | (Nr. 3 e) Die Hilfsdatei wird auf jedem Rückkehrpfad entfernt | nach Erzeugen, nach dem Wettlauf (T18) und nach `Datei nicht anlegbar`: die Liste der Dateien im Verzeichnis ist `pfad` oder leer | `Remove` im Zweig `fs.ErrExist` vergessen | andere |
| T21 | (Nr. 3 f) Fehlt das Verzeichnis oder ist es eine Datei, ist es `Datei nicht anlegbar`; es entsteht weder Datei noch Verzeichnis | Fälle: Eltern fehlt (2 Ebenen tief), Eltern ist eine reguläre Datei; danach fehlt der Pfad weiterhin | `MkdirAll` eingefügt | andere |
| T22 | (Nr. 4) Ein zweiter Aufruf nach dem Erzeugen liefert dasselbe Token und lässt die Datei unverändert | Fall: Bytes, Modus und Änderungszeit gleich | die Datei neu geschrieben | andere |
| T23 | (Nr. 4) Ein Aufruf ändert, kürzt, löscht und korrigiert eine vorhandene Datei nie, auch bei einem Verstoß nicht | Schleife über alle Verstöße von T2, T5, T7, T8, T9: Bytes und Modus vorher und nachher gleich | `Chmod` auf 0600 bei einem Rechteverstoß | andere |
| T24 | (Nr. 5, 6) Der Text ist `ODR-E2001 ADMIN_TOKEN_FILE: <Grund>` mit einem der sieben Gründe und nennt weder Inhalt noch Pfad | Schleife über die sieben Gründe als Literale; Zählprüfung gegen die Quelle: die Zahl der Gründe im Quelltext (`go/parser`) gleich 7 gleich den Zeilen von Nr. 6; ein Canary im Inhalt der Datei und im Namen des Verzeichnisses steht nicht im Text; `errors.As(*TokenFehler)` | der Pfad im Text | andere |
| T25 | (Nr. 5; `LH-FA-13.a`: „nie in einem Log“) `LadeToken` schreibt weder Log noch `stdout` noch `stderr` | `log.SetOutput`, `os.Stdout`, `os.Stderr` auf Pipes: nach Erzeugen und nach jedem Verstoß 0 Bytes | `fmt.Println(token)` | andere |
| T26 | (Nr. 1, 3) Das Lesen von `ADMIN_TOKEN_FILE` aus der Umgebung, der Default `<DATA_DIR>/admin.token`, der Aufruf beim Start und das Einsortieren der Zeile liegen bei `slice-bootstrap-m6` | **Akzeptiertes Negativ.** Grund: `SPEC-110` Nr. 2 lehnt `ADMIN_*` ab, ein gelesener, aber nicht verdrahteter Schlüssel wäre eine still wirkungslose Sicherheitseinstellung. Folge: bis `slice-bootstrap-m6` erzeugt der Prozess kein Token; der Admin-Listener ist nicht erreichbar (§1) | – | andere |

Zeilen gesamt 24: *a* = 2 (T18, T19), *b* = 22 (das akzeptierte Negativ T26 zählt als `andere`, `SPEC-128` Nr. 3 e).

**Entscheidungen des Architect für diesen Teil** (2026-11-07; Aufträge 4 und 8, A bis D, G, L, N und O; die Aufträge der übrigen Nummern stehen in den Plänen der Teile 2 bis 6):

- **Prüfungen und Reihenfolge der Token-Datei.** Die Reihenfolge der Prüfungen (sieben Gründe, Rang in `SPEC-239` Nr. 6), das Erzeugen (Hilfsdatei und `Link`, Rechte von Anfang an), die Plattform (Rechteprüfung außer unter `windows`), der Symlink und die Grenze von 32 Zeichen (nach dem Entfernen von ASCII-Leerraum, Maß Bytes) sind entschieden: `SPEC-239`. Das Lesen von `ADMIN_TOKEN_FILE` aus der Umgebung ist Sache von `slice-bootstrap-m6`.
- **Pflicht-Mengen.** Jede Menge hat Achsen, Generator, Orakel und Zählprüfung; das Orakel trägt Gründe, Meldungstexte und Rechte als Literale aus dem Vertrag, nie als Konstanten des Codes.

  | Menge | Achsen und Elemente | Generator | Orakel | Zählprüfung gegen die Quelle |
  |---|---|---|---|---|
  | M-Token-Inhalt (T1) | Kern 5 × Rand links 8 × Rand rechts 8 gleich 320 | Generator | Kern als Literal je Fall | 320 |
  | M-Rechte (T2) | 4 gültige, 6 einzelne Bits, 3 Kombinationen gleich 13 | Liste | gültig oder `Gruppen- oder Weltrechte` | 13 |
  | M-Nicht-regulär (T9) | FIFO, `/dev/null`, Symlink darauf, Symlink ohne Ziel, Socket gleich 5 | Liste | `keine reguläre Datei` | 5 |
  | M-Gründe (T24) | 7 Gründe der Token-Datei | Liste | Literale | 7 gleich der Zahl im Quelltext (`go/parser`) und den Zeilen von `SPEC-239` Nr. 6 |
  | M-Umask (T16) | 000, 022, 077, 0277 | Liste | `Perm()` gleich 0600 | 4 |

  **Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** die Größe der Token-Datei (Prüfgrenze 1 MiB; Grund: `os.ReadFile` liest alles; Folge: eine Datei von 1 GiB wird ganz gelesen, der Betreiber legt sie an).
- **Zweige als Menge, Runden unter dem Detektor** (Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat null Blöcke mit Zähler 0, außer den drei benannten Negativen Z-N1 bis Z-N3; der Implementer nennt sie mit Datei und Funktion im Bericht):

  | Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
  |---|---|---|
  | `LadeToken`: Stat-Fehler, Verzeichnis, nicht regulär, Rechte, mehr als eine Zeile, unter 32, gültig, Symlink, Erzeugen, Hilfsdatei anlegen scheitert, `Link` gewinnt, `Link` verliert (`fs.ErrExist`), Aufräumen | T1, T2, T4, T5, T7 bis T10, T13, T14, T16 bis T21 | `Link` verliert (T18, T19): ein Lauf je Zweig |
  | Z-N1: Lesefehler von `os.ReadFile` (`SPEC-239` Nr. 1 e) | Negativ: der Superuser des Images liest jede Datei; Folge: der Zweig ist ohne Fänger, ein Betreiber ohne Leserecht sieht `Datei nicht lesbar` | – |
  | Z-N2: Fehler von `Write`, `Sync`, `Chmod`, `Link` außer `fs.ErrExist` | Negativ: `os.File` ohne Naht; Folge: `Datei nicht anlegbar` ohne Test | – |
  | Z-N3: der Zweig unter `windows` ohne Rechteprüfung | Negativ: das Image läuft auf Linux; Folge: unter Windows prüft niemand Rechte | – |
- **Aufrufstellen und Fakes.** Die Token-Datei ruft keinen Port; sie hat keine Aufrufstelle für „jeder Fall je Aufrufstelle“ und braucht keinen Fake (*f* = 0). Das Dateisystem ist das echte (Testverzeichnisse); der Haltepunkt am Argument ist der Parameter von `lade` (T16, T17, T18).
- **Token-Handling und Canary.** Gefangen: Fehlertexte (T24) und `stdout`, `stderr` und `log.Default()` (T25) mit einem Canary aus Bytewerten (`AGENTS.md` §3.18) im Inhalt der Datei und im Namen des Verzeichnisses. Die Schleife über alle acht Ausgabewege des Rahmens (`SPEC-238` Nr. 13) liegt in `slice-admin-abweisen-audit` (H1).
- **Größe.** Aus den Zeilen dieses Teils: *a* = 2, *b* = 22, *f* = 0, *p* = 150 (Produktcode der Token-Datei etwa 150); der Block in §4.
- **Kennungen.** `SPEC-238` (Rahmen), `SPEC-239` (Token-Datei), `SPEC-240` (Zeile `ODR-E2102`) sind vergeben; frei ab `SPEC-241`; keine Nummer aus den Räumen von `slice-mock-lesen`, Linie D und der Reserve `SPEC-237`.
- **Pfade neben den Zweigen.** Dieser Teil schreibt allein `src/adapters/driving/admin/` (neu, kein Zweig berührt es). Die Dateien, die der Zweig mit anderen Zweigen teilt (Tabelle von `SPEC-109`, die Zahl in `SPEC-230` Nr. 8, `fehler.go`, die Tests des Audit und der Telemetrie, `harness/katalog-ausnahmen.txt`), schreibt `slice-admin-rahmen`; der Hinweis zu ihrem Merge steht dort in §6 (Rückgabe R4).
- **Reports.** Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (Register `report-der-rolle-faellt-docs-check`); der Planner reicht den Satz an beide weiter.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Raten der Fänger** (`AGENTS.md` §3.10): T18 und T19 sind Zusagen über Nebenläufigkeit (zwei Aufrufe auf eine fehlende Datei); T18 ist der deterministische Fänger (Haltepunkt), T19 der Lauf mit 16 Goroutinen. Jede Mutation muss in jedem Lauf rot werden; der Implementer trägt die Tabelle *Zusage · Mutation · roter Test* und die Raten je Mutation (Zahl der Prozesse, Messender, Last) in einen Absatz von §3 ein, bevor er an den Reviewer übergibt.

**Schnitt des Planner** (2026-11-07, vor dem Code, `AGENTS.md` §3.14, `SPEC-128` Nr. 1). Die Rechnung des Architect ergab *S* = 85·12 + 80·90 + 200·4 + 570 = 9590 und *M* = 13000, Schnittfrage ja. Der Vorschlag war sechs Teile; die Zeilen nach den Nummern des Vertrags geordnet. Der Planner hat ihn gegen die Tabelle geprüft und vier Zeilen und zwei Fakes verschoben, weil Fänger Code voraussetzen, den ein späterer Teil liefert:

| Teil | Slice | Zeilen der Tabelle | *a* | *b* | *f* | *p* | *S* | *M* |
|---|---|---|---|---|---|---|---|---|
| 1 | `slice-admin-listener-token` (dieser Plan) | T1, T2, T4, T5, T7 bis T26 (24) | 2 | 22 | 0 | 150 | 2080 | 2900 |
| 2 | [slice-admin-rahmen](../open/slice-admin-rahmen.md) | K1 bis K3, Z1, Z3 bis Z8, Y2 bis Y4 (13) | 0 | 13 | 3 | 110 | 1750 | 2400 |
| 3 | [slice-admin-authentisierung](../open/slice-admin-authentisierung.md) | K4, A1, A3 bis A6, U1 bis U9, E2 bis E4 (18) | 0 | 18 | 0 | 100 | 1540 | 2100 |
| 4 | [slice-admin-lebenszyklus](../open/slice-admin-lebenszyklus.md) | A2, K5, S1 bis S10, X2, G1 bis G3 (16) | 7 | 9 | 1 | 100 | 1615 | 2200 |
| 5 | [slice-admin-fehlerabbildung](../open/slice-admin-fehlerabbildung.md) | B1 bis B9, C1 bis C5, E1, E5, X1, H2 (18) | 0 | 18 | 0 | 70 | 1510 | 2100 |
| 6 | [slice-admin-abweisen-audit](../open/slice-admin-abweisen-audit.md) | D1 bis D7, H1, N1, Y1, Y5, Y6, P1 (13) | 3 | 10 | 0 | 40 | 1095 | 1500 |

Summe 102 Zeilen, *a* = 12, *b* = 90, *f* = 4, *p* = 570, *S* = 9590. Jeder Teil liegt unter 2223 (kein `schnitt-grund` nötig). Abweichungen vom Vorschlag des Architect, je mit Grund: **K4** (Kopie des Tokens) nach Teil 3, weil der Fänger die Authentisierung braucht (200 und 401); **K5** (Server teilen nichts) und **A2** (angehaltener Server) nach Teil 4, weil beide `Anhalten` rufen; **E1** und **E5** nach Teil 5, weil beide Schleifen über die Fehler laufen, die eine Route liefert (Abbildung in Teil 5); **N1** nach Teil 6, weil die Schleife `Append` und `CountError` einschließt; **H1** (acht Ausgabewege) nach Teil 6, weil der Weg des Audit-Eintrags und die Fehlertexte von `Serve` und `Schliessen` erst dort bestehen; **G1** bleibt in Teil 4 (der Teil baut den `http.Server`). Die Fakes: der Vorschlag verteilte *f* auf die Teile 2, 3 und 6; `zaehlendeIDs`, `routeFake` und `listenerFake` braucht schon Teil 2 (K3, Z4, Z7), der Test-Server mit Client erst Teil 4. Die Teile folgen nacheinander (Reihenfolge, keine Code-Abhängigkeit zu Teil 1): alle schreiben in `src/adapters/driving/admin/`, und die Messung nach `SPEC-128` Nr. 4 g zählt jede Änderung an den Pfaden der Marke, auch die eines parallelen Zweigs. Schleifen über *R* (Requestarten) und die Antwortarten wachsen mit den Teilen; die Pläne nennen die Formen, die ihr Code erzeugt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-audit-log` (`welle-betrieb`) und `slice-telemetry-port` liegen in `done/`; für die Token-Datei selbst braucht der Teil keinen von beiden (er ruft keinen Port), die Bedingung gilt für den ungeschnittenen Slice und seine späteren Teile. Geprüft am 2026-11-07 (Planner, `main` = `02f252f`): beide liegen in `done/`; `slice-audit-dienste` und `slice-audit-verdrahtung` ebenfalls (nicht Trigger dieses Slice); `slice-telemetry-einbau-gateway` und `slice-listener-tls` sind keine Kanten dieses Slice (`welle-admin-api` §5), und der Slice ändert `src/bootstrap/` nicht, `SPEC-110` Nr. 2 und die Reihenfolge mit `slice-bootstrap-m4` bleiben unberührt. **Trigger der Welle:** `welle-admin-api` startet nicht; der Slice ist eine Ausnahme je Slice wie `slice-inspection-use-case` (Linie D, [welle-admin-api](../welle-admin-api.md) §2): sein Trigger (M5) ist Priorisierung, keine Code-Abhängigkeit, und er gilt, solange das Binary unverändert bleibt (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`); sonst Rückführung `in-progress` → `open`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** `LH-FA-13.a` verlangt eine Token-Datei mit Erzeugen und Prüfen; sie ist der einzige Teil des Rahmens, der kein HTTP und keinen Port braucht (*keine Code-Abhängigkeit*). Teil 1 hängt an nichts und ist die Reihenfolge der Teile voraus.

**Größenmarke** (Planner, 2026-11-07, `SPEC-128` Nr. 1): der Block unten trägt `pfad`, `produktcode` und `fakes` des Architect und `basis`, `schaetzung` und `marke` des Planner, in einem Commit nach dem Schnitt und vor dem ersten Code-Commit gesetzt. 24 Zeilen der Tabelle, *a* = 2, *b* = 22, *f* = 0, *p* = 150: *S* = 85·2 + 80·22 + 200·0 + 150 = 2080, *M* = 2900, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Die Teile 2 bis 6 tragen ihre Marke als Prosa; ihren Block setzt der Planner beim Start des jeweiligen Slice.

```groessenmarke
pfad: src/adapters/driving/admin/
produktcode: 150
fakes: 0
basis: 20c02c538f1d6a0b2910cf0a5831e9ea4263c932
schaetzung: 2080
marke: 2900
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Block in diesem Abschnitt trägt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Ein Teil, der sich ohne geteilten Mechanismus abtrennen lässt, steht dann vor dem Code in diesem Abschnitt, sonst steht die Messung mit Grund im Plan (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert — Carveout?): die Token-Datei verlangt eine Plattformregel, die der Architect nicht festlegt (siehe §6), oder das Binary ändert sich (der Teil ruft `bootstrap` und liest `ADMIN_*`).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07, Zeilen dieses Teils): jede Zelle nennt eine Nummer von `SPEC-239` (Token-Datei) oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12). Die Raster der übrigen Operationen stehen in den Plänen der Teile 2 bis 6.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Token-Datei laden (Pfad, Inhalt, Rechte) | `SPEC-239` Nr. 3: fehlt die Datei, wird erzeugt; Nr. 1: leerer `pfad` ist `ErrUngueltig` | Nr. 1 g: leere Datei und Datei nur aus Leerraum, `weniger als 32 Zeichen` | Nr. 1 a bis f und Nr. 6: Verzeichnis, Symlink ohne Ziel, FIFO, Gerät, Rechte von Gruppe oder Welt, mehrere Zeilen | gilt nicht: synchron beim Start ohne Kontext (`SPEC-110` Nr. 13) | Nr. 4: kein Zustand, dasselbe Ergebnis, die Datei bleibt unverändert |
| Token-Datei erzeugen (fehlende Datei) | Nr. 3 f: Verzeichnis fehlt, `Datei nicht anlegbar`, nichts entsteht | Nr. 1: leerer `pfad` ist `ErrUngueltig` | Nr. 3 d und f: das Verzeichnis ist eine Datei; ein anderer Aufruf legt die Datei zwischen Prüfung und `Link` an (`fs.ErrExist`) | gilt nicht: synchron beim Start ohne Kontext (`SPEC-110` Nr. 13) | Nr. 3 d und Nr. 4: zwei gleichzeitige Aufrufe liefern dasselbe Token, ein zweiter Aufruf lässt die Datei unverändert |

**Zusatztabelle: Paare von Verstößen am selben Schlüssel** (`AGENTS.md` §3.16): Mehrere Verstöße am selben Pfad entscheidet die Rangfolge von `SPEC-239` Nr. 6; die Fänger sind Zeilen der Tabelle in §3.

| Schlüssel und Paar oder Modus | Ergebnis (Ort) | Fänger |
|---|---|---|
| Token-Datei: Verzeichnis mit Rechten 0777 | `ist ein Verzeichnis` (`SPEC-239` Nr. 6, Rang 2 vor 4) | T11 |
| Token-Datei: FIFO mit Rechten 0666 | `keine reguläre Datei`, ohne zu blockieren (Rang 3 vor 4) | T11, T9 |
| Token-Datei: Rechte 0644 und zwei Zeilen | `Gruppen- oder Weltrechte` (Rang 4 vor 6) | T11 |
| Token-Datei: Rechte 0644 und weniger als 32 Zeichen | `Gruppen- oder Weltrechte` (Rang 4 vor 7) | T11 |
| Token-Datei: zwei Zeilen und zusammen weniger als 32 Zeichen | `mehr als eine Zeile` (Rang 6 vor 7) | T11 |
| Token-Datei: Rechte 0644, zwei Zeilen, weniger als 32 Zeichen | `Gruppen- oder Weltrechte` (Rang 4) | T11 |

- Die Prüfung der Rechte der Token-Datei gilt nach `LH-FA-13.a` „auf POSIX-Systemen“; das Verhalten auf anderen Systemen und im Test ohne gesetzte Rechte hat keinen Ort (Auftrag 4) — **Ausgang:** entschieden (Architect, 2026-11-07, `SPEC-239` Nr. 2): die Prüfung entfällt allein unter `windows`; die Tests setzen Rechte mit `os.Chmod`, unabhängig von der Umask (T2, T16); der Zweig unter `windows` ist ein Negativ (Z-Liste, Z-N3).
- `LH-FA-13.a` nennt keine Begrenzung von Fehlversuchen; ein vom Betreiber gesetztes Token von nur 32 Zeichen wäre durch Raten angreifbar — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User (Owner); kein Code, bis entschieden. Sichtbar gehalten im Meilenstein M6 der Roadmap und in `welle-admin-api` §5 (Rückgabe R5 des Architect: jede Abweisung mit `ODR-E2101` schreibt einen Eintrag ohne Drosselung, `slice-admin-abweisen-audit`).
- `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×): die Token-Datei bei zwei Prozessen (T18, T19) — T18 ist der deterministische Fänger, T19 der Lauf mit Rate je Prozess im Plan — **Ausgang:** weiter offen → `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` im Register.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (12×): jeder Zweig von `LadeToken` hat einen Lauf unter dem Detektor (Z-Liste in §3, drei benannte Negative) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/akzeptiertes-negativ-ohne-folge` (15×): T26 und die Z-Negative Z-N1 bis Z-N3 tragen Grund und Folge im Plan — **Ausgang:** weiter offen → `BEO-REPO/akzeptiertes-negativ-ohne-folge` im Register.
- Fänger gegen einen Tokenwert in `<DATA_DIR>/audit.jsonl` des Prozesses (Sendung von `slice-audit-verdrahtung`) — **Ausgang:** entschieden; liegt bei `slice-bootstrap-m6` (dort §1 *Übernimmt außerdem*, am 2026-11-07 gelesen); dieser Slice fängt am Fehlertext und an den Pipes (T24, T25).
- Größe: die Tabelle liegt vor, der Slice ist vor dem Code geschnitten (*S* der Teile 1095 bis 2080, Tabelle in §3); der Block in §4 trägt `basis`, `schaetzung` und `marke` — **Ausgang:** entfallen (Marke gesetzt, Planner, vor dem ersten Code-Commit).
- Rückgaben R1 bis R6 des Architect (Schnitt, Test der Arten, Pflichten der Endpunkte, geteilte Dateien, Entscheidung des Users, Verdrahtung): R1 — **Ausgang:** entschieden (Planner, 2026-11-07): sechs Teile, Tabelle in §3. R2 — **Ausgang:** entschieden (Planner): die Änderung der einen Testdatei `src/bootstrap/audit_betrieb_test.go` verletzt „Binary unverändert“ nicht; sie gehört zu `slice-admin-abweisen-audit`. R3 — **Ausgang:** entschieden (Planner): die Pflichten von `SPEC-238` Nr. 15 stehen in §1 und DoD von `slice-admin-sessions`, `slice-admin-imports` und `slice-admin-workspaces` mit diesem Slice als Geber. R4 — **Ausgang:** entschieden (Planner): der Hinweis zum Merge der geteilten Dateien steht in §6 von `slice-admin-rahmen` (der Teil, der `fehler.go` ändert); dieser Teil berührt keine. R5 — **Ausgang:** weiter offen (User, siehe oben). R6 — **Ausgang:** entschieden (Planner): `slice-bootstrap-m6` trägt Aufruf von `LadeToken`, Meldung, `Konfig`, die Zelle `admin_abgewiesen`, die Reihenfolge des Herunterfahrens und den Canary-Fänger (§1 und DoD dort).
- Gate-Kette (Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch`): der Teil ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).

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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Start nachgezogen (Planner, 2026-11-07, `main` = `02f252f`, Zähler aus den Dateien unter `evidence/`) und beim Schnitt für die Token-Datei (`main` = `aadbef0`): einschlägig sind `nebenlaeufiger-lauf-nur-im-glueck-gruen` 13× (Token-Datei bei zwei Prozessen, `AGENTS.md` §3.10), `zweig-ohne-lauf-unter-dem-detektor` 12× (Zweige von `LadeToken`), `akzeptiertes-negativ-ohne-folge` 15× (T26, Z-N1 bis Z-N3), `betriebsregel-fuer-repo-dateien-verletzt` 9×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 44×, `randform-erst-im-review-entschieden` 34×, `menge-im-plan-weicht-vom-orakel-des-tests-ab` 3×, `unbeschraenkte-menge-ohne-benannte-pruefgrenze` 11×, `schaetzformel-ueberschaetzt-den-slice` 21×, `report-der-rolle-faellt-docs-check` 5× (Verkörperung, soweit die Schwelle erreicht ist, beim Lese-Schritt der Welle-Closure); kein neues Verzeichnis, kein Zähler gesetzt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
