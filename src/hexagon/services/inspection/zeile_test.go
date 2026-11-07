package inspection_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// enthalten sind die Felder der Allowlist, ausgeschlossen die übrigen von model.Exchange (SPEC-236 Nr. 5).
func enthalten() []string {
	return []string{
		"ID", "SequenceNo", "Method", "Path", "StartedAt", "FinishedAt", "KeyRedacted", "RequestBody", "ResponseBody",
		"ResponseStatus", "Capture", "RequestRedact", "ResponseRedact", "Integrity", "Equivalence",
	}
}

func ausgeschlossen() []string {
	return []string{"SessionID", "RawQuery", "MatchKey", "RequestHeaders", "ResponseHeaders", "RequestBlob", "ResponseBlob"}
}

// felder liefert Name und Typ der Felder eines Struct-Typs.
func felder(typ reflect.Type) map[string]reflect.Type {
	m := map[string]reflect.Type{}
	for i := range typ.NumField() {
		m[typ.Field(i).Name] = typ.Field(i).Type
	}
	return m
}

// TestAllowlistIstMengeDerFelder (X11): enthaltene und ausgeschlossene Felder des Tests sind genau die
// Felder von model.Exchange, ExchangeZeile führt genau die enthaltenen mit demselben Typ. Ein neues
// Feld in model.Exchange ohne Entscheidung macht den Test rot.
func TestAllowlistIstMengeDerFelder(t *testing.T) {
	ex, ze := felder(reflect.TypeFor[model.Exchange]()), felder(reflect.TypeFor[inbound.ExchangeZeile]())
	alle := append(slices.Clone(enthalten()), ausgeschlossen()...)
	if len(alle) != len(ex) || len(enthalten()) != len(ze) {
		t.Fatalf("Felder des Tests %d, von model.Exchange %d; enthalten %d, ExchangeZeile %d", len(alle), len(ex), len(enthalten()), len(ze))
	}
	for _, n := range enthalten() {
		if typ, da := ze[n]; !da || typ != ex[n] {
			t.Errorf("enthaltenes Feld %s: in ExchangeZeile %v, in model.Exchange %v", n, typ, ex[n])
		}
	}
	for _, n := range ausgeschlossen() {
		if _, da := ex[n]; !da {
			t.Errorf("ausgeschlossenes Feld %s fehlt in model.Exchange", n)
		}
		if _, da := ze[n]; da {
			t.Errorf("ausgeschlossenes Feld %s steht in ExchangeZeile", n)
		}
	}
}

// setzeWert setzt v auf einen Wert, der vom Nullwert und von dem anderer Felder verschieden ist (i).
func setzeWert(v reflect.Value, i int) {
	switch v.Kind() {
	case reflect.String:
		v.SetString(fmt.Sprintf("wert-%d", i))
	case reflect.Int:
		v.SetInt(int64(100 + i))
	case reflect.Bool:
		v.SetBool(true)
	default:
		// Gerades i: lokale Zeit mit monotoner Ablesung; ungerades i: Zone ungleich UTC (In verwirft die
		// monotone Ablesung). Ein Dienst, der die Zeit umrechnet oder rundet, fällt in einer der Formen auf.
		if i%2 == 0 {
			v.Set(reflect.ValueOf(time.Now().Add(time.Duration(i) * time.Hour)))
		} else {
			v.Set(reflect.ValueOf(time.Now().In(time.FixedZone("test", (i+1)*3600))))
		}
	}
}

// erwartet bildet die Zeile aus e, indem es jedes enthaltene Feld nach Namen kopiert.
func erwartet(e model.Exchange) inbound.ExchangeZeile {
	var z inbound.ExchangeZeile
	for _, n := range enthalten() {
		reflect.ValueOf(&z).Elem().FieldByName(n).Set(reflect.ValueOf(e).FieldByName(n))
	}
	return z
}

// TestFeldUnveraendert (X11): jedes enthaltene Feld erscheint mit seinem Wert, und kein anderes Feld
// der Zeile erscheint mit einem; die Basis trägt nur eine sequence_no, FinishedAt ist die Nullzeit
// und ResponseStatus 0, wo sie nicht das geprüfte Feld sind.
func TestFeldUnveraendert(t *testing.T) {
	// Zwei Runden: i ändert in der zweiten Runde seine Parität, die Zeitfelder tragen beide Formen.
	for i, name := range slices.Concat(enthalten(), enthalten()) {
		e := model.Exchange{SessionID: sid, SequenceNo: 1}
		setzeWert(reflect.ValueOf(&e).Elem().FieldByName(name), i)
		f := neuesFake()
		f.sitzungen[sid] = []model.Exchange{e}
		z := lies(t, neuerDienst(t, f), sid, inbound.Seite{})
		if len(z.Eintraege) != 1 {
			t.Fatalf("%s: %d Einträge", name, len(z.Eintraege))
		}
		wert := reflect.ValueOf(e).FieldByName(name).Interface()
		if got := reflect.ValueOf(z.Eintraege[0]).FieldByName(name).Interface(); !reflect.DeepEqual(got, wert) {
			t.Errorf("%s: Wert %v, erwartet %v", name, got, wert)
		}
		if !reflect.DeepEqual(z.Eintraege[0], erwartet(e)) {
			t.Errorf("%s: Zeile %+v, erwartet %+v", name, z.Eintraege[0], erwartet(e))
		}
	}
}

// ort ist ein Einbringungsort eines ausgeschlossenen Feldes (Menge M-Feld).
type ort struct {
	name, feld string
	setze      func(e *model.Exchange, c string)
}

func orte() []ort {
	kopf := func(name bool, c string) map[string][]string {
		if name {
			return map[string][]string{c: {"x"}}
		}
		return map[string][]string{"X-Kopf": {c}}
	}
	return []ort{
		{"SessionID", "SessionID", func(e *model.Exchange, c string) { e.SessionID = c }},
		{"RawQuery", "RawQuery", func(e *model.Exchange, c string) { e.RawQuery = c }},
		{"MatchKey", "MatchKey", func(e *model.Exchange, c string) { e.MatchKey = c }},
		{"RequestHeaders-Name", "RequestHeaders", func(e *model.Exchange, c string) { e.RequestHeaders = kopf(true, c) }},
		{"RequestHeaders-Wert", "RequestHeaders", func(e *model.Exchange, c string) { e.RequestHeaders = kopf(false, c) }},
		{"ResponseHeaders-Name", "ResponseHeaders", func(e *model.Exchange, c string) { e.ResponseHeaders = kopf(true, c) }},
		{"ResponseHeaders-Wert", "ResponseHeaders", func(e *model.Exchange, c string) { e.ResponseHeaders = kopf(false, c) }},
		{"RequestBlob", "RequestBlob", func(e *model.Exchange, c string) { e.RequestBlob = c }},
		{"ResponseBlob", "ResponseBlob", func(e *model.Exchange, c string) { e.ResponseBlob = c }},
	}
}

// kanarienvogel ist der Wert, der an einem Ort steht und nirgends in der Antwort erscheinen darf; er
// entsteht aus Bytewerten.
func kanarienvogel(name string) string { return string([]byte{0x4b, 0x4e, 0x52, 0x2d}) + name }

// alleStrings sammelt jede Zeichenkette unter v: Felder, Elemente, Schlüssel und Werte.
func alleStrings(v reflect.Value, aus *[]string) {
	switch v.Kind() {
	case reflect.String:
		*aus = append(*aus, v.String())
	case reflect.Struct:
		for i := range v.NumField() {
			alleStrings(v.Field(i), aus)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			alleStrings(v.Index(i), aus)
		}
	case reflect.Map:
		for it := v.MapRange(); it.Next(); {
			alleStrings(it.Key(), aus)
			alleStrings(it.Value(), aus)
		}
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			alleStrings(v.Elem(), aus)
		}
	}
}

// kodierungen sind die Darstellungen einer Seite, in denen der Canary gesucht wird.
func kodierungen() map[string]func(z inbound.ExchangeSeite) string {
	return map[string]func(z inbound.ExchangeSeite) string{
		"encoding/json": func(z inbound.ExchangeSeite) string { b, _ := json.Marshal(z); return string(b) },
		"%+v":           func(z inbound.ExchangeSeite) string { return fmt.Sprintf("%+v", z) },
		"Zeichenketten": func(z inbound.ExchangeSeite) string {
			var s []string
			alleStrings(reflect.ValueOf(z), &s)
			return strings.Join(s, "\n")
		},
	}
}

// seitenMit liest alle Seiten (Limit 1) einer Session aus drei Exchanges, die setze verändert.
func seitenMit(t *testing.T, id string, setze func(e *model.Exchange)) []inbound.ExchangeSeite {
	t.Helper()
	f := neuesFake()
	f.legeAn(id, 3)
	for i := range f.sitzungen[id] {
		setze(&f.sitzungen[id][i])
	}
	seiten := folge(t, neuerDienst(t, f), id, 1, 0)
	if len(seiten) != 3 {
		t.Fatalf("%d Seiten statt 3", len(seiten))
	}
	return seiten
}

// TestKeinGeheimnis (X12, Menge M-Feld: 9 Orte × 3 Kodierungen): an keinem Einbringungsort eines
// ausgeschlossenen Feldes erscheint der Canary in einer Seite, gleich in welcher Kodierung. Die
// Orte sind genau die ausgeschlossenen Felder (die Header mit zwei Orten); die Gegenprobe hält
// den Canary im Pfad als sichtbar fest.
func TestKeinGeheimnis(t *testing.T) {
	felderDerOrte := map[string]int{}
	for _, o := range orte() {
		felderDerOrte[o.feld]++
	}
	if len(felderDerOrte) != len(ausgeschlossen()) || felderDerOrte["RequestHeaders"] != 2 || felderDerOrte["ResponseHeaders"] != 2 {
		t.Fatalf("Orte je Feld %v, ausgeschlossen %v", felderDerOrte, ausgeschlossen())
	}
	for _, n := range ausgeschlossen() {
		if felderDerOrte[n] == 0 {
			t.Errorf("ausgeschlossenes Feld %s ohne Ort", n)
		}
	}
	for _, o := range orte() {
		c := kanarienvogel(o.name)
		id := sid
		if o.name == "SessionID" {
			id = c
		}
		seiten := seitenMit(t, id, func(e *model.Exchange) { o.setze(e, c) })
		for kname, k := range kodierungen() {
			for _, z := range seiten {
				if strings.Contains(k(z), c) {
					t.Errorf("Ort %s, Kodierung %s: der Canary steht in der Antwort", o.name, kname)
				}
			}
		}
	}
	c := kanarienvogel("Path")
	for kname, k := range kodierungen() {
		seiten := seitenMit(t, sid, func(e *model.Exchange) { e.Path = c })
		if !strings.Contains(k(seiten[0]), c) {
			t.Errorf("Gegenprobe, Kodierung %s: der Canary im Pfad ist nicht zu sehen", kname)
		}
	}
}
