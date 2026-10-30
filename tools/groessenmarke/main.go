// Command groessenmarke rechnet aus der Tabelle Zusage · Fänger eines Slice-Plans die Schätzung
// und prüft die Marke des Blocks gegen die Formel und die Messung (SPEC-128):
// `groessenmarke abfrage <Plan>` und `groessenmarke pruefen <Plan> <Messung>`. Es liest nur
// Dateien und schreibt nur nach stdout und stderr.
package main

import (
	"fmt"
	"io"
	"os"
)

const (
	exitOK      = 0
	exitBefund  = 1
	exitEingabe = 2

	aufrufText = "Aufruf: groessenmarke abfrage <Plan>\n       groessenmarke pruefen <Plan> <Messung>\n"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// aufrufGueltig ist wahr für `abfrage <Plan>` und `pruefen <Plan> <Messung>` ohne leeres Argument.
func aufrufGueltig(args []string) bool {
	for _, a := range args {
		if a == "" {
			return false
		}
	}
	switch {
	case len(args) == 2:
		return args[0] == "abfrage"
	case len(args) == 3:
		return args[0] == "pruefen"
	}
	return false
}

func run(args []string, stdout, stderr io.Writer) int {
	if !aufrufGueltig(args) {
		_, _ = fmt.Fprint(stderr, aufrufText)
		return exitEingabe
	}
	text, err := os.ReadFile(args[1])
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "groessenmarke: Plan fehlt: %s\n", args[1])
		return exitEingabe
	}
	if args[0] == "abfrage" {
		return abfrage(args[1], string(text), stdout, stderr)
	}
	m, err := os.ReadFile(args[2])
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "groessenmarke: Messung fehlt: %s\n", args[2])
		return exitEingabe
	}
	return pruefen(args[1], string(text), m, stdout, stderr)
}

// hinweis schreibt einen Hinweis in der Form von SPEC-128 Nr. 6 a.
func hinweis(w io.Writer, name, text string) {
	_, _ = fmt.Fprintf(w, "groessenmarke %s: Hinweis: %s\n", name, text)
}

// eingabefehler meldet die Eingabefehler eines Plans oder einer Messung auf stderr.
func eingabefehler(w io.Writer, name string, fehler []string) int {
	for _, f := range fehler {
		_, _ = fmt.Fprintf(w, "groessenmarke %s: Eingabe: %s\n", name, f)
	}
	return exitEingabe
}

// abfrage sagt, ob der Plan gemessen wird (SPEC-128 Nr. 7 a): stdout `keine-messung` oder
// `messen` mit Basis und Pfaden; die Hinweise stehen auf stderr.
func abfrage(name, text string, stdout, stderr io.Writer) int {
	p := lesePlan(text)
	if len(p.fehler) > 0 {
		return eingabefehler(stderr, name, p.fehler)
	}
	switch p.zustand {
	case keinBlock:
		hinweis(stderr, name, "kein Block `"+blockInfo+"`, der Plan wird nicht gemessen")
	case prosaForm:
		hinweis(stderr, name, "Prosa-Form der Marke ohne Block, der Plan wird nicht gemessen")
	case markeAusstehend:
		hinweis(stderr, name, fmt.Sprintf("Marke steht aus: Schätzung %d, Marke %d",
			p.tabellenSchaetzung(), markeZu(p.tabellenSchaetzung())))
	case vollstaendig:
		_, _ = fmt.Fprintf(stdout, "messen\nbasis %s\n", p.blk.basis)
		for _, pf := range p.blk.pfade {
			_, _ = fmt.Fprintf(stdout, "pfad %s\n", pf)
		}
		return exitOK
	}
	_, _ = fmt.Fprintln(stdout, "keine-messung")
	return exitOK
}

// pruefen wendet SPEC-128 Nr. 5 und 6 an: Hinweise und die Zeile ok auf stdout, dazu die
// Messpunkt-Zeile auch bei einem Befund, die Befunde auf stderr.
func pruefen(name, text string, messungText []byte, stdout, stderr io.Writer) int {
	p := lesePlan(text)
	if len(p.fehler) > 0 {
		return eingabefehler(stderr, name, p.fehler)
	}
	if p.zustand != vollstaendig {
		return eingabefehler(stderr, name, []string{"kein vollständiger Block, pruefen misst nur Pläne, für die abfrage messen sagt"})
	}
	m := leseMessung(messungText, p.blk)
	if m.fehler != "" {
		return eingabefehler(stderr, name, []string{m.fehler})
	}
	r := pruefe(p, m.summe)
	if m.hinweis != "" {
		hinweis(stdout, name, m.hinweis)
	}
	for _, h := range r.hinweise {
		hinweis(stdout, name, h)
	}
	if len(r.befunde) == 0 {
		_, _ = fmt.Fprintf(stdout, "groessenmarke %s: ok, Messung %d von Marke %d\n", name, m.summe, p.blk.marke)
	}
	_, _ = fmt.Fprintf(stdout, "Messpunkt %s: zeilen=%d nebenlaeufig=%d fakes=%d produktcode=%d schaetzung=%d marke=%d messung=%d\n",
		name, p.tab.nebenlaeufig+p.tab.andere, p.tab.nebenlaeufig, p.blk.fakes, p.blk.produktcode,
		p.tabellenSchaetzung(), p.blk.marke, m.summe)
	for _, b := range r.befunde {
		_, _ = fmt.Fprintf(stderr, "groessenmarke %s: %s\n", name, b)
	}
	if len(r.befunde) > 0 {
		return exitBefund
	}
	return exitOK
}
