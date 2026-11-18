package urlparser_test

import (
	"strings"
	"testing"
)

// Die Dekodierung in jeder Stellung (SPEC-262 Nr. 4): zehn Escapes (`%2E`, `%2e`, `%2F`, `%2f`, `%2C`, `%2c`, `%2A`, `%24`, `%28`,
// `%29`) in 21 Stellungen eines Elements, kurz und nach 120 Segmenten. Die Eingabe steht roh, das Orakel `exOrakel` bekommt den
// dekodierten Wert. Offen bleibt: ein anderes Escape als diese zehn.

// exEscapeZahl ist die Zahl der Eingaben: 10 Escapes * 21 Stellungen * 2 Längen.
const exEscapeZahl = 420

func TestExpandEscapeInJederStellung(t *testing.T) {
	p := exParser()
	escapes := [][2]string{{"%2E", "."}, {"%2e", "."}, {"%2F", "/"}, {"%2f", "/"}, {"%2C", ","}, {"%2c", ","}, {"%2A", "*"}, {"%24", "$"}, {"%28", "("}, {"%29", ")"}}
	formen := []string{"E", "Ea", "aE", "aEa", "EE", "aEEa", "a/E", "E/a", "a.E", "E.a", "a/$ref/E", "$refE", "E$ref", "a/E$ref", "a,E", "E,a", "a,E,b", "EEE", "a/Ea/b", "Ns.T/E", "a/b.E/c"}
	vorlauf := strings.Repeat("a/", 120)
	n := 0
	for _, e := range escapes {
		for _, form := range formen {
			roh := strings.ReplaceAll(form, "E", e[0])
			dek := strings.ReplaceAll(form, "E", e[1])
			exPruefe(t, p, exQ(roh), exOrakel(dek))
			exPruefe(t, p, exQ(vorlauf+roh), exOrakel(vorlauf+dek))
			n += 2
		}
	}
	exZaehlt(t, "TestExpandEscapeInJederStellung", n, exEscapeZahl)
}
