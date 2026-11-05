package replay_test

import (
	"context"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestRequestTextOhneZiel hält Z14 (SPEC-135 Nr. 7 c): der Request-Body als `text/plain` ist `typ`
// und ODR-E5001, auch auf den Pfaden der Ziele und unter beiden Modi des Matchings. Der Austausch
// mit dem Key, der bei einem gesetzten Ziel entstünde, liegt in der Ablage: setzte der Replay das
// Ziel, fände er ihn.
func TestRequestTextOhneZiel(t *testing.T) {
	zahl := 0
	for _, odata := range []string{"/Products/$count", "/Products(1)/Orders/$count", "/Products(1)/Name/$value", "/Products(1)/password/$value"} {
		for _, modus := range []model.BodyModus{model.BodyBytes, model.BodyJSONCanonical} {
			for _, body := range []string{"12", "S3CR3T"} {
				zahl++
				u := neu(t)
				u.k.Regeln.BodyMode = modus
				u.k.Pfade = pfadPassword(t)
				a := mitBodyAnfrage([]byte(body), "text/plain")
				a.Pfad = root + odata
				zaehlung, wert, name := model.TextZiel(odata)
				if !zaehlung && !wert {
					t.Fatalf("%s: der Pfad ist kein Ziel", odata)
				}
				var aus []byte
				if zaehlung || name != "password" {
					aus = []byte(body)
				} else {
					aus = []byte(model.Marker)
				}
				s, err := model.BildeSchluessel(model.SchluesselEingabe{
					Method: a.Methode, ServiceID: svcID, Path: odata, Header: a.Header, Body: aus, RequestRedaction: model.RedactionNone,
				}, u.k.Regeln)
				if err != nil {
					t.Fatalf("BildeSchluessel: %v", err)
				}
				u.lege(u.austausch(1, s.MatchKey, true))
				_, err = u.replay(context.Background(), a)
				erwarteCode(t, odata+" "+string(modus)+" "+body, err, model.CodeE5001)
				if _, da := u.cursorStand(); da {
					t.Errorf("%s: Cursor", odata)
				}
			}
		}
	}
	if zahl != 16 {
		t.Errorf("%d Fälle", zahl)
	}
}
