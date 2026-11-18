package hybrid

import (
	"bytes"
	"io"
	"math"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// fehlerLeser liefert bei jedem Read den Fehler, mit dem das Lesen des Bodys beim Puffern endete.
type fehlerLeser struct{ err error }

func (f fehlerLeser) Read([]byte) (int, error) { return 0, f.err }

// probeCodes sind die Codes, mit denen jeder Weg zu R oder U beginnt (SPEC-284 Nr. 5 a).
func probeCodes() [4]model.Code {
	return [4]model.Code{model.CodeE6001, model.CodeE6002, model.CodeE6003, model.CodeE5001}
}

// fallbackFaehig sagt, ob die Policy für die Anfrage bei mindestens einem der Codes einen Fallback erlaubt; das Ziel
// ist das der Anfrage (SPEC-284 Nr. 5 a).
func (v *Verteiler) fallbackFaehig(a inbound.Anfrage, ziel string) bool {
	for _, c := range probeCodes() {
		if v.policy.Entscheide(policy.Anfrage{Methode: a.Methode, Pfad: a.Pfad, Ziel: ziel, Code: c}) != policy.Bleiben {
			return true
		}
	}
	return false
}

// stufen sind die Anfragen für M, R und U und die Aussage, ob ein Fallback erlaubt bleibt.
type stufen struct {
	mock, replay, upstream inbound.Anfrage
	fallback               bool
}

// vorbereiten liefert die Anfragen für M, R und U und sagt, ob ein Fallback erlaubt bleibt (SPEC-284 Nr. 5). Ohne Body,
// bei einer nicht fallback-fähigen Anfrage und über der Grenze geht die Anfrage unverändert an beide Stufen und ohne
// Pufferung an M; gepuffert erhält jede Stufe einen eigenen Leser über die vollständigen Bytes. U erhält ohne Body
// weder Leser noch Länge (nil und 0).
func (v *Verteiler) vorbereiten(a inbound.Anfrage, ziel string) stufen {
	ohneBody := a
	ohneBody.Body, ohneBody.Laenge = nil, 0
	switch {
	case a.Body == nil || a.Laenge == 0:
		return stufen{mock: a, replay: a, upstream: ohneBody, fallback: true}
	case a.Laenge > v.maxBytes:
		return stufen{mock: a, replay: a, upstream: a}
	case !v.fallbackFaehig(a, ziel):
		return stufen{mock: a, replay: a, upstream: a, fallback: true}
	}
	grenze := v.maxBytes
	if grenze < math.MaxInt64 {
		grenze++
	}
	puffer, err := io.ReadAll(io.LimitReader(a.Body, grenze))
	switch {
	case err != nil:
		m := a
		m.Body = io.MultiReader(bytes.NewReader(puffer), fehlerLeser{err: err})
		return stufen{mock: m, replay: a, upstream: a}
	case int64(len(puffer)) > v.maxBytes:
		m := a
		m.Body = io.MultiReader(bytes.NewReader(puffer), a.Body)
		return stufen{mock: m, replay: a, upstream: a}
	}
	m, r, u := a, a, a
	m.Body, m.Laenge = bytes.NewReader(puffer), int64(len(puffer))
	r.Body, r.Laenge = bytes.NewReader(puffer), int64(len(puffer))
	u.Body, u.Laenge = bytes.NewReader(puffer), int64(len(puffer))
	return stufen{mock: m, replay: r, upstream: u, fallback: true}
}
