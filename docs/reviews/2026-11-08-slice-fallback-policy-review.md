# Review slice-fallback-policy (Teil 1: Prüfung der Einträge und New)

* Reviewer, ein Lauf, 2026-11-08; Basis 9310c219, HEAD 6dc3d6dc, Zweig slice-fallback-policy
* Geprüft gegen: Plan `slice-fallback-policy` (P11, P13, P14, P17 bis P21, P01a, §3, §6), `SPEC-261` Nr. 1 bis 9, `AGENTS.md` §3.2, §3.6, §3.7, §3.10 bis §3.18, [LH-FA-12](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
* Nicht geprüft: die DoD (Verifier), `Entscheide` (nicht im Diff)

## Zusammenfassung

HIGH 0 · MEDIUM 1 · LOW 2 · INFO 3.

Mutationslauf (Liste des Implementers, `tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, unter `flock`, Last 7 bis 13 beim Lauf, Messender: Reviewer): 140 Mutationen, 139 rot, 1 grün, 0 flatternd, 0 Übersetzungsfehler, Laufzeit der Mutationen 61,6 s. Die grüne (`host-grenze-nach-9`, `c <= '9'` zu `c <= ':'` in `nameByte`) ist äquivalent: `:` wird vor `nameByte` am Port abgetrennt und erreicht die Funktion nie. Die Liste wurde nicht als einziger Beleg übernommen: die Zeilen und Zellen sind unten gegen den Code aufgelöst; zusätzliche Mutationen habe ich nicht gefahren (siehe LOW-2 zur Grenze).

## Findings

### MEDIUM

**M1**
- kategorie: MEDIUM
- quelle: `SPEC-261` Nr. 6 b, c, e; [LH-FA-12](../../spec/lastenheft.md)
- pfad: `src/hexagon/services/hybrid/policy/freigabe.go` · `func parseSegment` (Zweig `case n == 1:`)
- befund: Das Muster `/svc/.*` (ebenso `/svc/*.`, `/svc/%2e*`, `/svc/*%2e`) ist nach Nr. 6 b gültig und wird angenommen, weil die Prüfung auf `.` und `..` nur Segmente ohne `*` trifft; die Ausnahme in Nr. 6 c ("ein ganz von `*` gedecktes Segment ist weder `.` noch `..`") greift bei einem teilweise gedeckten Segment nicht, sodass `.*` den Pfadsegment `..` und `%2e*` das Segment `%2e.` decken kann. Ein Upstream löst beide als Elternverzeichnis auf; die Freigabe wäre breiter als der Wortlaut, den Nr. 6 e als Grund der Menge W nennt.
- verifizierbar: ja (Fall der Eintragsprüfung: `POST /a/.*` wird angenommen; die Entscheidung selbst liegt in `slice-fallback-policy-entscheidung`)
- klasse: Muster-Syntax lässt Platzhalter zu, der Punkt-Segmente deckt
- Zuständig: Architect (Wortlaut von Nr. 6 b/c) oder Teil 2; dieser Diff setzt den Wortlaut richtig um, der Orakel-Test (`orakelMuster`) bildet dieselbe Lücke ab. Failure-Szenario setzt einen Betreiber voraus, der ein solches Muster schreibt; ohne Eintrag dieser Form entsteht keine Freigabe.

### LOW

**L1**
- kategorie: LOW
- quelle: `AGENTS.md` §3.7 (Zusage nur, was da ist)
- pfad: `src/hexagon/services/hybrid/policy/policy.go` · "Das Paket importiert model und die Standardbibliothek"
- befund: Der Paketkommentar sagt im Indikativ, das Paket importiere `model`; kein Import von `model` steht im Paket (nur Standardbibliothek). `quelltext_test.go` erlaubt `model`, verlangt ihn aber nicht.
- verifizierbar: ja (Import-Liste)
- klasse: Kommentar nennt mehr als der Code hat

**L2**
- kategorie: LOW
- quelle: `AGENTS.md` §3.15 (akzeptiertes Negativ mit Grund und Folge), Maintainability
- pfad: `src/hexagon/services/hybrid/policy/quelltext_test.go` · `if dateien < 3 || importe < 4 || funktionen < 8`
- befund: Die Schwellen des Zählwächters sind feste Zahlen; ein Paket mit mehr Dateien ohne Befund ändert sie nicht, ein Entfernen von Dateien unter die Schwelle bricht sie als "zu wenig gelesen" (benannt, nicht stumm). Die Grenze des Quelltext-Tests (nur Wurzel des Paketverzeichnisses, keine Unterverzeichnisse, kein `init`-Verbot, `reflect`/`runtime` erlaubt) steht weder im Test noch im Plan als Negativ.
- verifizierbar: nein
- klasse: Grenze eines Quelltext-Fängers nicht benannt

### INFO

**I1** (Sicherheitsbewertung, Bitte um Gegenlesen)
- pfad: `host.go` · `adr.Is6()` in `zerlegeKlammer`
- befund: `[::ffff:1.2.3.4]` (IPv4-abgebildet) wird angenommen und zu `::ffff:1.2.3.4` kanonisiert; `[::ffff:102:304]` ergibt dieselbe Form. Das ist nach Nr. 5 a ("IPv6-Adresse ohne Zone") der Wortlaut, ein Fänger dafür fehlt in Gutfall- und Verstoßlisten (der Generator führt das Orakel `net.ParseIP` ohne Abbildung). Bewertung "bewusst offen": der Eintrag gibt nur den Host frei, den er nennt (Vergleich mit `Ziel` byte-gleich nach Kanonisierung), nicht mehr; die Allowlist ist Namens-, nicht Adressschutz (Nr. 5 e, *Ungeprüft*). Keine breitere Freigabe als der Eintrag. Empfohlen: ein Fall in der Tabelle von P11, damit die Entscheidung gezählt ist (Nr. 5 b ändern sie nicht).
- klasse: Randform ohne Fall, Wortlaut deckt sie

**I2** (Namen gegen `SPEC-261`)
- befund: `Werte`, `KonfigFehler{Schluessel, Eintrag, Ursache}` und die sieben Ursachen-Strings stimmen wörtlich mit der Signatur und Nr. 5 a, 6 a, 7, 8 c überein (`TestKonstanten`). Die Konstanten `Schluessel*` und `Ursache*` sowie `SchluesselUpstream` (für `ohne-hosts`) nennt der Vertrag nicht als Namen; der Wert ist durch Nr. 8 b/c vorgegeben, die Namen sind Entwurf des Implementers. Architect soll gegenlesen. Reihenfolge der Schlüssel in `New` (Hosts, Schreibfreigaben, Replay-Kategorien, dann Zusammenhang) entspricht Nr. 8 b.
- klasse: Namen ohne Vorgabe im Vertrag

**I3** (Größenmarke, nur bewertet)
- befund: Produktcode `policy.go` 118 + `freigabe.go` 116 + `host.go` 110 = 344 Zeilen gegen Block 300 (rund 15 Prozent darüber); Messung `git diff --numstat` ohne Plan 1221, unter Marke 1400. Kein Verstoß gegen `SPEC-128`, die Marke hält.
- klasse: Schätzung unter Messung

## Auflösung der Tabelle (Zeile · Mutation am Code · Fänger)

| Zeile | Gelesen am Code | Ergebnis |
|---|---|---|
| P01a | akzeptiertes Negativ mit Grund, Folge und Fänger-Adresse (P01 in `slice-fallback-policy-entscheidung`); `New` kopiert in `hostEintrag`/`freigabe` (Strings unveränderlich), hält keinen Slice | ok |
| P11 | Mutationen Port `0`/`01`/`65536`/leer, Zone, `[1.2.3.4]`, Rest hinter `]`, Platzhalter, Grenzen `A-Z`/`a-z`/`0-9` der Liste rot (139/140); Generator (Orakel unabhängig: Regex, `net.ParseIP`) läuft über vier Grundformen mit allen 256 Bytes | ok |
| P13 | Rangfolge `leer`, `trenner`, `methode`, `muster` am Code (`IndexByte` am ersten Leerzeichen); zweites Leerzeichen und Leerzeichen am Ende fallen in `musterBytes` (`< 0x21`), Fall vorhanden | ok |
| P14 | Zeichenmenge, `%`-Prüfung (`i+2 >= len(m)`), `**` nur als letztes ganzes Segment, ein `*` je Segment, Punkt-Segmente nach `%2e`/`%2E`; Menge 37 plus Generator mit fünf Grundformen | ok bis M1 |
| P17 | `pruefeKategorie` exakt, Generator über beide Namen (n+1)·256+n·256+n | ok |
| P18 | erste Verletzung je Schlüssel (`break`), `Eintrag` ab 1, Reihenfolge, `errors.Join`, keine Policy, Prüfung auch bei `Upstream` aus: je eine Mutation in der Liste rot | ok |
| P19 | `Error()` ohne Wert; Marker für fünf Ursachen (`host`, `trenner`, `methode`, `muster`, `kategorie`) | ok; kein Leck: `KonfigFehler` enthält keinen Eintragstext, kein Pfad gibt ihn an `fmt` |
| P20 | `ohne-hosts` nur bei `Upstream` und `len(Hosts) == 0`; Zeilen Upstream × Hosts vollständig | ok |
| P21 | `TestQuelltext` liest das Paket per `go/parser`; Importe, `go`, `var`, Methoden; `TestImportErlaubt` fängt die Liste selbst; Fehlerverhalten: Lesefehler und Parsefehler sind `t.Fatal`, zu wenig Gelesenes ist `t.Fatalf`; Panik möglich nur bei leerer `Recv.List` (kann in gültigem Go nicht auftreten) | ok, siehe L2 |

Raster §6 (Operation × Argumentart): für die Eintragsprüfung und `New` sind fehlender, leerer, unbekannter und fremder Wert, Abbruch (kein Kontext, gilt nicht) und Wiederholung (`New` ruft nichts, keine Umgebung) durch P11 bis P20 belegt; keine Zelle im Code entschieden, die der Vertrag nicht trägt.

## Sicherheit, gezielt

- Platzhalter: `*` und `**` werden in `Hosts` abgelehnt (`nameByte`), in `Schreibfreigaben` je Segment geparst; `%`, `\` , `?`, `#`, `,`, Leerzeichen und alle Bytes außerhalb von `0x21` bis `0x7E` fallen in `musterBytes`. Dekodiert oder normalisiert wird nichts. Die Breite eines Platzhalters entscheidet erst `Entscheide` (nicht im Diff); M1 betrifft die Syntax.
- Groß-/Kleinschreibung: Host-Namen werden zu ASCII-Kleinschreibung gefaltet, Muster byte-gleich; entspricht Nr. 5 b und 6 c.
- IPv6: `netip.ParseAddr` mit Abweisung von Zone und `Is4()`; Kanonisierung über `String()`; siehe I1.
- Geheimnisse: der Text von `*KonfigFehler` trägt Schlüssel, Nummer und Ursache; kein Pfad der Fehler trägt den Wert des Eintrags.

## geprüft, ohne Befund

- `src/hexagon/services/hybrid/policy` (`host.go`, `policy.go` außer L1, `export_test.go`): ohne Befund; `export_test.go` ist die einzige interne Testdatei und über `testpackage.skip-regexp` in `.golangci.yml` gedeckt (unverändert gegenüber der Basis).
- `.golangci.yml`, Gates: nicht verändert, keine `//nolint`-Direktive im Diff (§3.2, §3.6).
- Plan `slice-fallback-policy.md`: §3 und Messung (1221 gegen Marke 1400) nachgezogen (§3.9, §3.14); Tabelle und Raster tragen Fänger oder benanntes Negativ (§3.15, §3.16).
- Quelltext der Dateien auf Escape-Folgen (§3.18): keine Folge aus Backslash, `u` und Hex im Produktcode; Testwerte mit Rohbytes sind als `\x..` geschrieben.
