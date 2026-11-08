package jsoncodecv2_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestEtagNieGeschrieben (E13, SPEC-169 Nr. 5, SPEC-281 Nr. 3): jsoncodecv2 liest das Feld Etag nicht; mit gesetztem und mit ungültigem Etag
// schreibt es dieselben Bytes wie ohne, als Entity und als Collection.
func TestEtagNieGeschrieben(t *testing.T) {
	etags := []string{`W/"0"`, `W/"7"`, `W/"9223372036854775807"`, `7`, `W/"07"`, `W/"` + bs + `"`, `w/"1"`}
	ohne := entityBody(t, "ohne", gutEntity())
	ohneKol := collectionBody(t, "ohne", gutCollection())
	for _, etag := range etags {
		ea := gutEntity()
		ea.Entity.Etag = etag
		if got := entityBody(t, etag, ea); got != ohne {
			t.Errorf("Entity %q: %s, will %s", etag, got, ohne)
		}
		ca := gutCollection()
		for i := range ca.Entities {
			ca.Entities[i] = model.Entity{Typ: ca.Entities[i].Typ, Schluessel: ca.Entities[i].Schluessel, Properties: ca.Entities[i].Properties, Etag: etag}
		}
		if got := collectionBody(t, etag, ca); got != ohneKol {
			t.Errorf("Collection %q: %s, will %s", etag, got, ohneKol)
		}
	}
}
