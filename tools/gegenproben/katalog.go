package main

import "strings"

// Fälle 174 bis 204 (SPEC-102, SPEC-201): `katalog <Wurzel>` im Wegwerfverzeichnis des Falls. Der Baum
// trägt fehler.go mit vier Codes, die Spezifikation mit denselben, die Ausnahmen mit einem Code, konfig.go mit
// zwei Schlüsseln, cli.go mit zwei Kommandos und das Handbuch mit einer Zeile für jeden Namen aus H. Jeder Fall
// verändert einen Baustein dieses Baums. Zeichen jenseits von ASCII stehen als Bytewerte (AGENTS.md §3.18).

const (
	kaFehler    = "src/hexagon/model/fehler.go"
	kaSpez      = "spec/spezifikation.md"
	kaAusnahmen = "harness/katalog-ausnahmen.txt"
	kaHandbuch  = "docs/user/benutzerhandbuch.md"
	kaKonfig    = "src/bootstrap/konfig.go"
	kaCli       = "src/adapters/driving/cli/cli.go"
	kaParagraf  = "\xc2\xa74"
	kaKopfH     = "| Code | Bedeutung |\n|---|---|\n"
	kaSchl      = "Schl\xc3\xbcssel"
	kaKopfS     = "| " + kaSchl + " | Bedeutung |\n|---|---|\n"
	kaKopfK     = "| Kommando | Bedeutung |\n|---|---|\n"
)

// kaCodes sind die vier Codes des sauberen Baums; der zweite steht in den Ausnahmen.
func kaCodes() []string { return []string{"ODR-E2001", "ODR-E2101", "ODR-E2201", "ODR-E2301"} }

func kaFehlerText(codes ...string) string {
	var b strings.Builder
	b.WriteString("package model\n\nvar tabelle = []zeile{\n")
	for _, c := range codes {
		b.WriteString("\t{\"" + c + "\", 1},\n")
	}
	return b.String() + "}\n"
}

func kaSpezText(codes ...string) string {
	return kaSpezMit(kaSpec110("`MODE`, `LISTEN`"), codes...)
}

// kaSpec110 ist der Abschnitt SPEC-110 mit der Zeile der Schlüssel, deren Spans der Text nennt.
func kaSpec110(spans string) string {
	return "### SPEC-110 \xe2\x80\x94 Start\n\n1. **Leer.** Text.\n2. **" + kaSchl + " dieser Version.** Gepr\xc3\xbcft werden " + spans +
		".\n   Gesetzt (nichtleer) und nicht geliefert sind `ADMIN_*`.\n3. **Weiter.** Text.\n"
}

func kaSpezMit(spec110 string, codes ...string) string {
	var b strings.Builder
	b.WriteString("# Spezifikation\n\n## 4. Fehler-Codes und Logging-Felder\n\n| ID | Code | Bedingung | Aktion |\n|---|---|---|---|\n")
	for i, c := range codes {
		b.WriteString("| `SPEC-00" + string(rune('1'+i)) + "` | `" + c + "` | Bedingung | Aktion |\n")
	}
	return b.String() + "\n## 5. Metriken\n\nText.\n\n" + spec110
}

// kaKonfigText liest je Name ein Literal; `nichtGelieferte` nennt die Präfixe aus n.
func kaKonfigText(namen []string, n string) string {
	var b strings.Builder
	b.WriteString("package bootstrap\n\nfunc lade(p *pruefer) {\n")
	for _, s := range namen {
		b.WriteString("\tp.wert(\"" + s + "\")\n")
	}
	return b.String() + "}\n\nfunc nichtGelieferte() (praefixe, namen []string) {\n\treturn []string{" + n + "}, nil\n}\n"
}

// kaCliText trägt je Kommando eine Konstante.
func kaCliText(kommandos ...string) string {
	s := "package cli\n\nconst (\n\tkommandoName = \"odata-recorder\"\n"
	for i, k := range kommandos {
		s += "\tk" + string(rune('a'+i)) + " = \"" + k + "\"\n"
	}
	return s + ")\n"
}

func kaZeilen(kopf string, namen []string) string {
	s := kopf
	for _, n := range namen {
		s += "| `" + n + "` | Die Bedeutung. |\n"
	}
	return s
}

func kaHandbuchText(zeilen []string, davor, danach string) string {
	return kaHandbuchMit(zeilen, davor, danach, []string{"MODE", "LISTEN"}, []string{"start", "version"})
}

// kaHandbuchMit trägt hinter dem Abschnitt Fehlercodes die Abschnitte Einstellungen und Kommandos.
func kaHandbuchMit(zeilen []string, davor, danach string, schl, komm []string) string {
	var b strings.Builder
	b.WriteString("# Handbuch\n\n## Fehlercodes\n\n" + davor + kaKopfH)
	for _, c := range zeilen {
		b.WriteString("| `" + c + "` | Die Bedeutung. |\n")
	}
	return b.String() + danach + "\n## Einstellungen\n\nText.\n\n" + kaZeilen(kaKopfS, schl) + "\n## Kommandos\n\nText.\n\n" + kaZeilen(kaKopfK, komm)
}

// kaBaum ist der saubere Baum G1; die Fälle ersetzen einzelne Dateien.
func kaBaum() map[string]string {
	c := kaCodes()
	return map[string]string{
		kaFehler:    kaFehlerText(c...),
		kaSpez:      kaSpezText(c...),
		kaAusnahmen: c[1] + " | Grund\n",
		kaHandbuch:  kaHandbuchText([]string{c[0], c[2], c[3]}, "Text.\n\n", ""),
		kaKonfig:    kaKonfigText([]string{"MODE", "LISTEN"}, "\"ADMIN_\""),
		kaCli:       kaCliText("start", "version"),
	}
}

// kaFall baut einen Fall der Gruppe: das Programm mit dem Wegwerfverzeichnis als Wurzel; die Dateien aus
// aender ersetzen die des sauberen Baums, ein leerer Inhalt löscht die Datei aus dem Baum. Ein Text mit
// `1>` steht auf stdout, einer mit `!` darf nicht erscheinen, sonst steht er auf stderr.
func kaFall(name string, aender map[string]string, exit int, want ...string) Fall {
	baum := kaBaum()
	for d, inhalt := range aender {
		if inhalt == "" {
			delete(baum, d)
		} else {
			baum[d] = inhalt
		}
	}
	f := gmFall(name, gateKatalog, baum, []string{"{dir}"}, exit, want...)
	f.Name = "katalog/" + name
	return f
}

// katalogFaelle sind die Fälle G1 bis G31 (SPEC-102, Fälle 174 bis 204).
func katalogFaelle() []Fall {
	return append(katalogFaelleCodes(), katalogFaelleSchluessel()...)
}

// katalogFaelleCodes sind die Fälle G1 bis G15 zu den Fehlercodes.
func katalogFaelleCodes() []Fall {
	c := kaCodes()
	h := func(zeilen ...string) string { return kaHandbuchText(zeilen, "Text.\n\n", "") }
	return []Fall{
		kaFall("g1-sauber", nil, 0, "1>katalog: ok", "!katalog: Fehlercodes"),
		kaFall("g2-code-ohne-spezifikation", map[string]string{kaFehler: kaFehlerText(append(kaCodes(), "ODR-E9901")...),
			kaHandbuch: h(c[0], c[2], c[3], "ODR-E9901")}, 1,
			"katalog: Fehlercodes: ODR-E9901 in fehler.go (Zeile 8), nicht in "+kaParagraf, "!fehlt im Handbuch"),
		kaFall("g3-code-ohne-zeile", map[string]string{kaHandbuch: h(c[0], c[3])}, 1,
			"katalog: Fehlercodes: ODR-E2201 fehlt im Handbuch", "!nicht in "+kaParagraf),
		kaFall("g4-handbuch-nennt-unbekannt", map[string]string{kaHandbuch: kaHandbuchText([]string{c[0], c[2], c[3]}, "Siehe ODR-E9999.\n\n", "")}, 1,
			"katalog: Fehlercodes: Handbuch Zeile 5 nennt ODR-E9999, der nicht in fehler.go steht"),
		kaFall("g5-handbuch-nennt-ausnahme", map[string]string{kaHandbuch: kaHandbuchText([]string{c[0], c[2], c[3]}, "Siehe ODR-E2101.\n\n", "")}, 1,
			"katalog: Fehlercodes: Handbuch Zeile 5 nennt ODR-E2101, der in den Ausnahmen steht", "!nicht in fehler.go"),
		kaFall("g6-ausnahme-ohne-grund", map[string]string{kaAusnahmen: c[1] + " |\n"}, 1,
			"katalog: Fehlercodes: Ausnahme ODR-E2101 (Zeile 1) ohne Grund"),
		kaFall("g7-ausnahme-ohne-code", map[string]string{kaAusnahmen: c[1] + " | Grund\nODR-E9902 | Grund\n"}, 1,
			"katalog: Fehlercodes: Ausnahme ODR-E9902 (Zeile 2) steht nicht in fehler.go"),
		kaFall("g8-code-doppelt", map[string]string{kaFehler: kaFehlerText(append(kaCodes(), c[0])...)}, 1,
			"katalog: Fehlercodes: ODR-E2001 doppelt in fehler.go, erstes Vorkommen Zeile 4"),
		kaFall("g9-form", map[string]string{kaFehler: kaFehlerText(append(kaCodes(), "ODR-E123")...)}, 1,
			"katalog: Fehlercodes: Form: fehler.go Zeile 8: \"ODR-E123\" hat nicht die Form"),
		kaFall("g10-abschnitt-fehlt", map[string]string{kaHandbuch: "# Handbuch\n\n## Anderes\n\nText.\n"}, 2,
			"katalog: Eingabe: Handbuch: Abschnitt Fehlercodes fehlt", "!katalog: Fehlercodes"),
		kaFall("g11-fehler-go-fehlt", map[string]string{kaFehler: ""}, 2,
			"katalog: Eingabe: fehler.go fehlt: src/hexagon/model/fehler.go", "!katalog: Fehlercodes"),
		kaOhneArgument(),
		kaFall("g13-k-ohne-c", map[string]string{kaSpez: kaSpezText(append(kaCodes(), "ODR-E9903")...)}, 0, "1>katalog: ok"),
		kaFall("g14-code-nur-im-zaun", map[string]string{kaHandbuch: kaHandbuchText([]string{c[0], c[3]}, "Text.\n\n", "\n```\n"+c[2]+"\n```\n")}, 1,
			"katalog: Fehlercodes: ODR-E2201 fehlt im Handbuch", "!nicht in fehler.go"),
		kaFall("g15-ausnahme-mit-zeile", map[string]string{kaHandbuch: h(c[0], c[1], c[2], c[3])}, 1,
			"katalog: Fehlercodes: Handbuch Zeile 10 nennt ODR-E2101, der in den Ausnahmen steht"),
	}
}

// kaOhneArgument ist G12: der Aufruf ohne Argument gibt die Aufrufzeile aus und liest nichts.
func kaOhneArgument() Fall {
	f := kaFall("g12-ohne-argument", nil, 2, "Aufruf: katalog <Wurzel>", "!katalog: Eingabe", "!katalog: Fehlercodes")
	f.Argumente = []string{}
	return f
}
