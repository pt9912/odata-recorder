package httpupstream_test

import (
	"slices"
	"strings"
	"testing"
)

// fehlerFall ist ein Fall, in dem Senden mit einem Fehler endet. lauf baut Server und Client,
// prüft, was der Fall außer der Art zusagt, und liefert den Fehler; art ist der Fehler, den
// errors.Is erreichen muss. Die Tests A2, A5, A8, A9, A12, A13, A16 und A18 fahren ihre Fälle aus
// dieser Tabelle, A14 alle.
type fehlerFall struct {
	name string
	art  error
	lauf func(t *testing.T) error
}

func alleFehlerFaelle() []fehlerFall {
	return slices.Concat(argumentFaelle(), argumentGrenzFaelle(), rahmungFehlerFaelle(), fristFehlerFaelle(),
		nichtErreichbarFaelle(), kontextFehlerFaelle(), bodyFehlerFaelle(), rangfolgeFaelle(), blockiertFehlerFaelle())
}

// fahre fährt die Fälle und prüft je die Art des Fehlers.
func fahre(t *testing.T, faelle []fehlerFall) {
	t.Helper()
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			pruefeArt(t, f.lauf(t), f.art)
		})
	}
}

// TestFehlerOhneWerte: kein Fehlertext von Senden nennt den Marker. mitGeheim setzt ihn in den Pfad
// von Ziel, Methode, Pfad, Query und einen Header; jeder Verstoß gegen Ziel trägt ihn im Pfad von Ziel
// (außer dem leeren Ziel), jeder Fall mit gelesenem Kopf (101, ctx, Frist) in Status-Text und
// Header-Wert, 101 auch in Upgrade; einzelne Fälle in weitere Werte der Anfrage, Bytes der Antwort
// oder den Text eines Fehlers von Body oder des Fakes (A14, SPEC-120 Nr. 8).
func TestFehlerOhneWerte(t *testing.T) {
	for _, f := range alleFehlerFaelle() {
		t.Run(f.name, func(t *testing.T) {
			err := f.lauf(t)
			if err == nil {
				t.Fatal("kein Fehler")
			}
			if strings.Contains(err.Error(), geheim) {
				t.Errorf("Fehlertext nennt einen Wert: %q", err.Error())
			}
		})
	}
}
