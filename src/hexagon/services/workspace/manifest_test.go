package workspace_test

import (
	"strconv"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// fehlt löscht das Glied statt es zu setzen.
type fehltT int

const fehlt fehltT = 0

func setze(m obj, k string, v any) {
	if v == fehlt {
		delete(m, k)
		return
	}
	m[k] = v
}

// mitBaum liefert die Bytes des Standard-Manifests (drei Einträge), nachdem f den Baum geändert hat.
func mitBaum(f func(m obj)) []byte {
	m := manifestBaum("bsl-1", "svc", standardEnts())
	f(m)
	return []byte(schreibe(m))
}

// mitEintrag ändert den Eintrag i des Standard-Manifests; n ist die Zahl der Einträge (1 oder 3).
func mitEintrag(n, i int, f func(e obj)) []byte {
	m := manifestBaum("bsl-1", "svc", standardEnts()[:n])
	f(eintraege(m)[i].(obj))
	return []byte(schreibe(m))
}

// manifestFall ist ein Manifest, das der Dienst abweist (Verstoß) oder annimmt.
type manifestFall struct {
	name     string
	manifest []byte
}

func gueltig() []byte { return mitBaum(func(obj) {}) }

// syntaxVerstoesse (Z20): alle echten Präfixe eines gültigen Manifests und die Formen, die kein Objekt,
// nicht genau ein Wert oder kein gültiges UTF-8 sind.
func syntaxVerstoesse() []manifestFall {
	g := gueltig()
	var r []manifestFall
	for n := 0; n < len(g); n++ {
		r = append(r, manifestFall{"Präfix " + strconv.Itoa(n), g[:n]})
	}
	formen := map[string][]byte{
		"Array": []byte("[]"), "null": []byte("null"), "Zahl": []byte("1"), "String": []byte(`"s"`), "true": []byte("true"),
		"Rest x": append(append([]byte{}, g...), 'x'), "Rest Objekt": append(append([]byte{}, g...), "{}"...),
		"zwei Werte": append(append([]byte{}, g...), g...), "BOM": append([]byte{0xEF, 0xBB, 0xBF}, g...),
		"Rest Komma": append(append([]byte{}, g...), ','),
	}
	for n, b := range formen {
		r = append(r, manifestFall{n, b})
	}
	for _, f := range [][2]string{{"Milk", "Mi\xffk"}, {"Milk", "Mi\xc3k"}, {"Milk", "Mi\xed\xa0\x80k"}, {"Food", "Fo\x80d"}} {
		r = append(r, manifestFall{"ungültiges UTF-8 " + strconv.Quote(f[1]), []byte(replace(string(g), f[0], f[1]))})
	}
	return r
}

func replace(s, alt, neu string) string {
	for i := 0; i+len(alt) <= len(s); i++ {
		if s[i:i+len(alt)] == alt {
			return s[:i] + neu + s[i+len(alt):]
		}
	}
	panic("Stelle fehlt: " + alt)
}

// kopfVerstoesse (Z21): je eines der fünf Glieder des Kopfes mal seine Formen.
func kopfVerstoesse() []manifestFall {
	n := len(standardEnts())
	formen := map[string][]any{
		"manifest_version":     {fehlt, "1", 2, 0, roh("1.0"), roh("01"), roh("1e0"), roh("-1"), true, nil, []any{1}, obj{}},
		"snapshot_kind":        {fehlt, 1, nil, "workspace", "Baseline", "", []any{"baseline"}},
		"service_id":           {fehlt, 7, nil, "other", "", "SVC", "svc "},
		"baseline_snapshot_id": {fehlt, 7, nil, "other", "", "BSL-1", "bsl-1 "},
		"entity_count": {fehlt, roh(strconv.Itoa(n + 1)), roh(strconv.Itoa(n - 1)), "3", roh("3.0"), roh("3e0"), roh("03"), roh("-3"),
			true, nil, roh("99999999999999999999"), roh("0"), []any{3}},
	}
	var r []manifestFall
	for k, vs := range formen {
		for i, v := range vs {
			r = append(r, manifestFall{k + " Form " + strconv.Itoa(i), mitBaum(func(m obj) { setze(m, k, v) })})
		}
	}
	leer := manifestBaum("bsl-1", "svc", nil)
	leer["entity_count"] = roh("-0")
	return append(r, manifestFall{"entity_count -0 bei leerer Liste", []byte(schreibe(leer))})
}

// eintragVerstoesse (Z22): je Glied der Einträge mal seine Formen an erstem, mittlerem und letztem Element
// eines Manifests mit drei und mit einem Element; dazu die Formen von entities und seiner Elemente.
func eintragVerstoesse() []manifestFall {
	formen := map[string][]any{
		"entity_set":          {fehlt, nil, "", 7, []any{}, true, obj{}},
		"typed_key":           {fehlt, nil, "", 7, []any{}, true, obj{}},
		"completeness":        {fehlt, nil, "", "unknown", "Complete", "COMPLETE", "partial ", 1, []any{"complete"}},
		"payload":             {fehlt, nil, []any{}, 7, "s", true, roh(`"{}"`)},
		"selected_properties": {fehlt, nil, obj{}, "x", 7, []any{""}, []any{"a", "a"}, []any{1}, []any{nil}, []any{"a", ""}, []any{obj{}}, []any{[]any{}}, true},
	}
	var r []manifestFall
	for _, n := range []int{3, 1} {
		for _, i := range []int{0, 1, 2} {
			if i >= n {
				continue
			}
			for k, vs := range formen {
				for j, v := range vs {
					r = append(r, manifestFall{"Eintrag " + strconv.Itoa(i) + " von " + strconv.Itoa(n) + " " + k + " Form " + strconv.Itoa(j),
						mitEintrag(n, i, func(e obj) { setze(e, k, v) })})
				}
			}
		}
	}
	for j, v := range []any{fehlt, nil, obj{}, "s", 7, true} {
		r = append(r, manifestFall{"entities Form " + strconv.Itoa(j), mitBaum(func(m obj) { setze(m, "entities", v) })})
	}
	for j, v := range []any{fehlt, nil} {
		r = append(r, manifestFall{"entities Form " + strconv.Itoa(j) + " bei entity_count 0",
			mitBaum(func(m obj) { setze(m, "entities", v); m["entity_count"] = 0 })})
	}
	for i := 0; i < 3; i++ {
		for j, v := range []any{nil, 1, "s", []any{}, true} {
			r = append(r, manifestFall{"Element " + strconv.Itoa(i) + " Form " + strconv.Itoa(j),
				mitBaum(func(m obj) { eintraege(m)[i] = v })})
		}
	}
	return r
}

// doppelte (Z23): ein Paar (entity_set, typed_key) zweimal, benachbart und getrennt.
func doppelte() []manifestFall {
	var r []manifestFall
	for _, p := range [][2]int{{0, 1}, {0, 2}, {1, 2}} {
		r = append(r, manifestFall{"Paar " + strconv.Itoa(p[0]) + "," + strconv.Itoa(p[1]), mitBaum(func(m obj) {
			a, b := eintraege(m)[p[0]].(obj), eintraege(m)[p[1]].(obj)
			b["entity_set"], b["typed_key"] = a["entity_set"], a["typed_key"]
		})})
	}
	return r
}

// gueltigeFormen sind Manifeste, die der Dienst annimmt, und die Zahl ihrer Einträge.
func gueltigeFormen() map[string]struct {
	manifest []byte
	n        int
} {
	type f = struct {
		manifest []byte
		n        int
	}
	return map[string]f{
		"Kontrolle":                   {gueltig(), 3},
		"Leerraum dahinter":           {append(gueltig(), " \n\t\r "...), 3},
		"Leerraum davor":              {append([]byte(" \n"), gueltig()...), 3},
		"selected_properties leer":    {mitEintrag(3, 1, func(e obj) { e["selected_properties"] = []any{} }), 3},
		"payload leeres Objekt":       {mitEintrag(3, 0, func(e obj) { e["payload"] = roh("{}") }), 3},
		"unbekanntes Glied":           {mitEintrag(3, 0, func(e obj) { e["extra"] = 1 }), 3},
		"unbekanntes Glied Kopf":      {mitBaum(func(m obj) { m["extra"] = "x"; m["coverage"] = "keine Liste"; m["provenance"] = 5 }), 3},
		"gleicher Schlüssel, Sets":    {mitEintrag(3, 2, func(e obj) { e["typed_key"] = "Id=1" }), 3},
		"Schreibweise des Schlüssels": {mitEintrag(3, 1, func(e obj) { e["typed_key"] = "id=1"; e["entity_set"] = "Products" }), 3},
		"Typ ungleich Set":            {mitEintrag(3, 0, func(e obj) { e["payload"] = roh(`{"@odata.type":"#Model.Special","Id":"1"}`) }), 3},
	}
}

// alleVerstoesse sind alle Manifeste, die der Dienst mit ODR-E3001 abweist.
func alleVerstoesse() []manifestFall {
	var r []manifestFall
	r = append(r, syntaxVerstoesse()...)
	r = append(r, kopfVerstoesse()...)
	r = append(r, eintragVerstoesse()...)
	r = append(r, doppelte()...)
	return r
}

// manifestAbgewiesen: die Operation endet mit ODR-E3001 und ErrBeschaedigt, der Blob ist einmal geöffnet
// und einmal geschlossen, keine Transaktion lief.
func manifestAbgewiesen(t *testing.T, w *welt, err error, name string) {
	t.Helper()
	pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
	if w.p.anzahl("Blobs.Open") != 1 || w.p.anzahl("Strom.Close") != 1 || w.p.anzahl(lblTx) != 0 {
		t.Errorf("%s: Open %d, Close %d, Transaktion %d", name, w.p.anzahl("Blobs.Open"), w.p.anzahl("Strom.Close"), w.p.anzahl(lblTx))
	}
}
