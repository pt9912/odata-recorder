package bootstrap_test

import (
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestSatzVersion4OhneExpand (SPEC-262 Nr. 1): der Codec-Satz der Version 4.0, den der Aufbau des Dienstes trägt, liest `$expand`
// nicht: jeder Wert ist `option_nicht_unterstuetzt`, und `$filter` bleibt lesbar. Die Tabelle der Abnahme baut ihren Dienst selbst und
// sieht diese Zeile des Aufbaus nicht. Offen: ob der Dienst die Antwort `501` aus dieser Ursache bildet (Zeile `ODV4-E-03` c).
func TestSatzVersion4OhneExpand(t *testing.T) {
	url := bootstrap.SatzVersion4().URL
	for _, query := range []string{"$expand=Foo", "$expand=", "$expand=a,b", "$expand=*"} {
		_, err := url.ParseURL("/Products", query)
		var uf *model.URLFehler
		if !errors.As(err, &uf) || uf.Ursache != model.URLOptionNichtUnterstuetzt || uf.Ort != "query/1" {
			t.Errorf("%q: %v, erwartet option_nicht_unterstuetzt", query, err)
		}
	}
	if _, err := url.ParseURL("/Products", "$filter=Id eq 1"); err != nil {
		t.Errorf("$filter: %v", err)
	}
}
