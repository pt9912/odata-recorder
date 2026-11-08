package main

// katalogFaelleSchluessel sind die Fälle G16 bis G31 (SPEC-102, Fälle 189 bis 204; SPEC-201): je Befundart der
// Schlüssel und der Kommandos und je Familie der Eingabefehler ein Fall. Die Zeilen der Befunde folgen dem Baum aus
// kaBaum: konfig.go nennt seine Schlüssel ab Zeile 4, das Handbuch die Zeilen der Tabelle Einstellungen ab Zeile 19
// und die der Tabelle Kommandos ab Zeile 28 (eine dritte Zeile in Einstellungen ist Zeile 21).
func katalogFaelleSchluessel() []Fall {
	c := kaCodes()
	cod := []string{c[0], c[2], c[3]}
	hb := func(schl, komm []string) string { return kaHandbuchMit(cod, "Text.\n\n", "", schl, komm) }
	standardS, standardK := []string{"MODE", "LISTEN"}, []string{"start", "version"}
	mitS := func(extra string) string { return hb(append(append([]string{}, standardS...), extra), standardK) }
	schl := "katalog: " + kaSchl + ": "
	return []Fall{
		kaFall("g16-schluessel-ohne-spezifikation", map[string]string{
			kaKonfig:   kaKonfigText([]string{"MODE", "LISTEN", "NEU_X"}, "\"ADMIN_\""),
			kaHandbuch: mitS("NEU_X")}, 1,
			schl+"NEU_X in konfig*.go (Zeile 6), nicht in SPEC-110", "!fehlt im Handbuch"),
		kaFall("g17-schluessel-ohne-zeile", map[string]string{kaHandbuch: hb([]string{"MODE"}, standardK)}, 1,
			schl+"LISTEN fehlt im Handbuch", "!nicht in SPEC-110"),
		kaFall("g18-handbuch-nennt-unbekannt", map[string]string{kaHandbuch: mitS("UNBEKANNT_X")}, 1,
			schl+"Handbuch Zeile 21 nennt UNBEKANNT_X, der nicht in konfig*.go steht", "!den das Binary ablehnt"),
		kaFall("g19-handbuch-nennt-praefix", map[string]string{kaHandbuch: mitS("ADMIN_TOKEN")}, 1,
			schl+"Handbuch Zeile 21 nennt ADMIN_TOKEN, den das Binary ablehnt", "!der nicht in konfig*.go steht"),
		kaFall("g20-handbuch-nennt-abgelehnt", map[string]string{
			kaKonfig:   kaKonfigText([]string{"MODE", "LISTEN"}, "\"ADMIN_\", \"STRICT_COVERAGE\""),
			kaHandbuch: mitS("STRICT_COVERAGE")}, 1,
			schl+"Handbuch Zeile 21 nennt STRICT_COVERAGE, den das Binary ablehnt", "!der nicht in konfig*.go steht"),
		kaFall("g21-name-in-nichtgelieferte", map[string]string{kaKonfig: kaKonfigText([]string{"MODE", "LISTEN"}, "\"ADMIN_\", \"HYBRID_\"")}, 0,
			"1>katalog: ok", "!katalog: "+kaSchl),
		kaFall("g22-k-ohne-c", map[string]string{kaSpez: kaSpezMit(kaSpec110("`MODE`, `LISTEN`, `SERVICE_NEU`"), c...)}, 0, "1>katalog: ok"),
		kaFall("g23-k-ohne-c-mit-zeile", map[string]string{
			kaSpez:     kaSpezMit(kaSpec110("`MODE`, `LISTEN`, `SERVICE_NEU`"), c...),
			kaHandbuch: mitS("SERVICE_NEU")}, 1,
			schl+"Handbuch Zeile 21 nennt SERVICE_NEU, der nicht in konfig*.go steht", "!nicht in SPEC-110"),
		kaFall("g24-kommando-ohne-zeile", map[string]string{kaCli: kaCliText("start", "version", "stop")}, 1,
			"katalog: Kommandos: stop fehlt im Handbuch"),
		kaFall("g25-handbuch-nennt-kommando-unbekannt", map[string]string{kaHandbuch: hb(standardS, []string{"start", "version", "halt"})}, 1,
			"katalog: Kommandos: Handbuch Zeile 30 nennt halt, der nicht in cli.go steht"),
		kaFall("g26-kommando-in-einstellungen", map[string]string{kaHandbuch: mitS("start")}, 1,
			schl+"Form: Handbuch Zeile 21: \"start\" hat nicht die Form"),
		kaFall("g27-einstellungen-fehlt", map[string]string{kaHandbuch: "# Handbuch\n\n## Fehlercodes\n\nText.\n\n" + kaKopfH + kaZeilen("", cod) +
			"\n## Kommandos\n\nText.\n\n" + kaZeilen(kaKopfK, standardK)}, 2,
			"katalog: Eingabe: Handbuch: Abschnitt Einstellungen fehlt", "!katalog: "+kaSchl),
		kaFall("g28-kommandos-fehlt", map[string]string{kaHandbuch: "# Handbuch\n\n## Fehlercodes\n\nText.\n\n" + kaKopfH + kaZeilen("", cod) +
			"\n## Einstellungen\n\nText.\n\n" + kaZeilen(kaKopfS, standardS)}, 2,
			"katalog: Eingabe: Handbuch: Abschnitt Kommandos fehlt", "!katalog: Kommandos"),
		kaFall("g29-konfig-verzeichnis-fehlt", map[string]string{kaKonfig: ""}, 2,
			"katalog: Eingabe: konfig*.go fehlt: src/bootstrap", "!katalog: "+kaSchl),
		kaFall("g30-spec110-ohne-zeile", map[string]string{kaSpez: kaSpezMit("### SPEC-110 \xe2\x80\x94 Start\n\n1. **Leer.** Text.\n", c...)}, 2,
			"katalog: Eingabe: SPEC-110: "+kaSchl+" fehlen", "!katalog: "+kaSchl+":"),
		kaFall("g31-cli-ohne-konstante", map[string]string{kaCli: "package cli\n\nfunc run() {}\n"}, 2,
			"katalog: Eingabe: cli.go: kein Kommando gefunden", "!katalog: Kommandos"),
	}
}
