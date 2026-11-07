package main

import "strings"

// Fälle 174 bis 188 (SPEC-102, SPEC-201): `katalog <Wurzel>` im Wegwerfverzeichnis des Falls. Der Baum
// trägt fehler.go mit vier Codes, die Spezifikation mit denselben, die Ausnahmen mit einem Code und das
// Handbuch mit einer Zeile für jeden Code aus H. Jeder Fall verändert einen Baustein dieses Baums.

const (
	kaFehler    = "src/hexagon/model/fehler.go"
	kaSpez      = "spec/spezifikation.md"
	kaAusnahmen = "harness/katalog-ausnahmen.txt"
	kaHandbuch  = "docs/user/benutzerhandbuch.md"
	kaParagraf  = "\xc2\xa74"
	kaKopfH     = "| Code | Bedeutung |\n|---|---|\n"
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
	var b strings.Builder
	b.WriteString("# Spezifikation\n\n## 4. Fehler-Codes und Logging-Felder\n\n| ID | Code | Bedingung | Aktion |\n|---|---|---|---|\n")
	for i, c := range codes {
		b.WriteString("| `SPEC-00" + string(rune('1'+i)) + "` | `" + c + "` | Bedingung | Aktion |\n")
	}
	return b.String() + "\n## 5. Metriken\n\nText.\n"
}

func kaHandbuchText(zeilen []string, davor, danach string) string {
	var b strings.Builder
	b.WriteString("# Handbuch\n\n## Fehlercodes\n\n" + davor + kaKopfH)
	for _, c := range zeilen {
		b.WriteString("| `" + c + "` | Die Bedeutung. |\n")
	}
	return b.String() + danach
}

// kaBaum ist der saubere Baum G1; die Fälle ersetzen einzelne Dateien.
func kaBaum() map[string]string {
	c := kaCodes()
	return map[string]string{
		kaFehler:    kaFehlerText(c...),
		kaSpez:      kaSpezText(c...),
		kaAusnahmen: c[1] + " | Grund\n",
		kaHandbuch:  kaHandbuchText([]string{c[0], c[2], c[3]}, "Text.\n\n", ""),
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

// katalogFaelle sind die Fälle G1 bis G15 (SPEC-102, Fälle 174 bis 188).
func katalogFaelle() []Fall {
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
