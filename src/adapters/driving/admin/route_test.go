package admin_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeichenmengen des Vertrags (SPEC-238 Nr. 2 b) als Literale: Literal, erstes Zeichen und Rest eines Namens.
const (
	zeichenLiteral = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._~-"
	zeichenName1   = "abcdefghijklmnopqrstuvwxyz"
	zeichenNameN   = "abcdefghijklmnopqrstuvwxyz_"
)

// routenFall ist eine Route des Generators mit der Erwartung des Orakels.
type routenFall struct {
	name    string
	route   admin.Route
	gueltig bool
}

// z1Achsen sind die Elemente der Achsen von M-Route: ungültige Methode, festes ungültiges Muster, Bytewert im Literal,
// im ersten und im weiteren Zeichen eines Namens, nil-Handler.
type z1Achsen struct {
	methoden, muster []string
	faelle           []routenFall
}

func z1Generator() z1Achsen {
	a := z1Achsen{
		methoden: []string{"", "get", "HEAD", "OPTIONS", "GET "},
		muster:   []string{"", "x", "/", "/a/", "//a", "/a//b", "/{}", "/{a", "/a}", "/{a}{b}", "/a{b}", "/a/{b}/{b}"},
	}
	h := neueRoute().handle
	for _, m := range a.methoden {
		a.faelle = append(a.faelle, routenFall{"Methode " + fmt.Sprintf("%q", m), admin.Route{Methode: m, Muster: "/a", Handler: h}, false})
	}
	for _, m := range a.muster {
		a.faelle = append(a.faelle, routenFall{"Muster " + fmt.Sprintf("%q", m), admin.Route{Methode: "POST", Muster: m, Handler: h}, false})
	}
	for i := 0; i < 256; i++ {
		c := string([]byte{byte(i)})
		a.faelle = append(a.faelle,
			routenFall{fmt.Sprintf("Literal %#x", i), admin.Route{Methode: "POST", Muster: "/a" + c, Handler: h}, strings.Contains(zeichenLiteral, c)},
			routenFall{fmt.Sprintf("Name1 %#x", i), admin.Route{Methode: "POST", Muster: "/{" + c + "}", Handler: h}, strings.Contains(zeichenName1, c)},
			routenFall{fmt.Sprintf("NameN %#x", i), admin.Route{Methode: "POST", Muster: "/{a" + c + "}", Handler: h}, strings.Contains(zeichenNameN, c)},
		)
	}
	a.faelle = append(a.faelle, routenFall{"nil-Handler", admin.Route{Methode: "POST", Muster: "/a"}, false})
	return a
}

// TestRoutenGrammatik hält Z1 (M-Route): jede Route außerhalb der Grammatik von SPEC-238 Nr. 2 a bis c ist ErrUngueltig
// ohne Server, jede innerhalb gültig. Das Orakel sind die Zeichenmengen des Vertrags; die Zählprüfung hält die Menge
// gegen die Summe der Achsen und die gültigen Fälle gegen die Längen der Zeichenmengen.
func TestRoutenGrammatik(t *testing.T) {
	a := z1Generator()
	if len(zeichenLiteral) != 66 || len(zeichenName1) != 26 || len(zeichenNameN) != 27 {
		t.Fatalf("Zeichenmengen %d, %d, %d, erwartet 66, 26, 27", len(zeichenLiteral), len(zeichenName1), len(zeichenNameN))
	}
	if soll := len(a.methoden) + len(a.muster) + 3*256 + 1; len(a.faelle) != soll || soll != 786 {
		t.Fatalf("%d Fälle, erwartet die Summe der Achsen %d (786)", len(a.faelle), soll)
	}
	gueltig := 0
	for _, f := range a.faelle {
		k, _ := konfigMit(f.route)
		srv, err := admin.New(k)
		if f.gueltig {
			gueltig++
		}
		if f.gueltig != (err == nil && srv != nil) || (!f.gueltig && (!errors.Is(err, model.ErrUngueltig) || srv != nil)) {
			t.Errorf("%s: Fehler %v, Server %v, erwartet gültig = %v", f.name, err, srv, f.gueltig)
		}
	}
	if soll := len(zeichenLiteral) + len(zeichenName1) + len(zeichenNameN); gueltig != soll {
		t.Errorf("%d gültige Fälle, erwartet %d", gueltig, soll)
	}
}

// TestRoutenGutfaelle hält Z1: alle fünf Methoden, mehrere Parameter, Unterstrich im Namen, ein Muster mit führenden
// Unterstrichen und eine leere Liste der Routen sind gültig.
func TestRoutenGutfaelle(t *testing.T) {
	h := neueRoute().handle
	var faelle [][]admin.Route
	for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		faelle = append(faelle, []admin.Route{{Methode: m, Muster: "/a", Handler: h}})
	}
	faelle = append(faelle,
		[]admin.Route{{Methode: "POST", Muster: "/__admin/sessions/{id}/close", Handler: h}},
		[]admin.Route{{Methode: "GET", Muster: "/{a}/{b}/{c}", Handler: h}},
		[]admin.Route{{Methode: "POST", Muster: "/{a_b}", Handler: h}},
		nil, []admin.Route{})
	if len(faelle) != 10 {
		t.Fatalf("%d Gutfälle, erwartet 10 (fünf Methoden, drei Muster, nil und leere Liste)", len(faelle))
	}
	for i, routen := range faelle {
		k, _ := konfigMit(routen...)
		if srv, err := admin.New(k); err != nil || srv == nil {
			t.Errorf("Gutfall %d: Fehler %v, Server %v", i, err, srv)
		}
	}
}

// TestRoutenTextOhneWert hält Z1 und Z3: der Text des Fehlers nennt weder Methode noch Muster; der Canary steht in
// beiden und in keinem Text.
func TestRoutenTextOhneWert(t *testing.T) {
	h := neueRoute().handle
	c := canary()
	faelle := map[string][]admin.Route{
		"Methode":            {{Methode: c, Muster: "/a", Handler: h}},
		"Muster":             {{Methode: "POST", Muster: "/" + c + "!", Handler: h}},
		"Methode und Muster": {{Methode: c, Muster: "/" + c + "!", Handler: h}},
		"Handler":            {{Methode: "POST", Muster: "/" + c}},
		"Überlappung":        {{Methode: "POST", Muster: "/" + c, Handler: h}, {Methode: "POST", Muster: "/" + c, Handler: h}},
		"Name doppelt":       {{Methode: "POST", Muster: "/" + c + "/{a}/{a}", Handler: h}},
	}
	for name, routen := range faelle {
		k, _ := konfigMit(routen...)
		srv, err := admin.New(k)
		if !errors.Is(err, model.ErrUngueltig) || srv != nil {
			t.Errorf("%s: Fehler %v, Server %v", name, err, srv)
			continue
		}
		if strings.Contains(err.Error(), c) {
			t.Errorf("%s: der Text nennt den Wert: %q", name, err)
		}
	}
}

// z3Muster sind die neun Muster von M-Überlappung; /b/c unterscheidet sich von /a/c allein im ersten Segment.
func z3Muster() []string {
	return []string{"/a/b", "/a/c", "/a/{x}", "/{x}/c", "/a/{y}", "/a", "/a/b/c", "/healthz", "/b/c"}
}

// modellUeberlappt ist das Modell des Tests je Segment (SPEC-238 Nr. 2 d): gleiche Zahl der Segmente, an jeder Stelle
// gleiches Literal oder mindestens ein Parameter.
func modellUeberlappt(a, b string) bool {
	x, y := strings.Split(a[1:], "/"), strings.Split(b[1:], "/")
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		parameter := strings.HasPrefix(x[i], "{") || strings.HasPrefix(y[i], "{")
		if !parameter && x[i] != y[i] {
			return false
		}
	}
	return true
}

// TestRoutenUeberlappung hält Z3 (M-Überlappung): 45 Paare aus neun Mustern, je mit gleicher und mit verschiedener
// Methode; gleiche Methode überlappt genau, wenn das Modell es sagt, verschiedene nie. Dazu vier Fälle der eingebauten Route.
func TestRoutenUeberlappung(t *testing.T) {
	muster := z3Muster()
	h := neueRoute().handle
	faelle := 0
	for i := range muster {
		for j := i; j < len(muster); j++ {
			for _, gleich := range []bool{true, false} {
				faelle++
				m2 := "PUT"
				if gleich {
					m2 = "POST"
				}
				k, _ := konfigMit(admin.Route{Methode: "POST", Muster: muster[i], Handler: h}, admin.Route{Methode: m2, Muster: muster[j], Handler: h})
				srv, err := admin.New(k)
				soll := !(gleich && modellUeberlappt(muster[i], muster[j]))
				if soll != (err == nil && srv != nil) || (!soll && !errors.Is(err, model.ErrUngueltig)) {
					t.Errorf("%s %s mit %s %s: Fehler %v, erwartet gültig = %v", "POST", muster[i], m2, muster[j], err, soll)
				}
			}
		}
	}
	if n := len(muster); faelle != n*(n+1)/2*2 || faelle != 90 {
		t.Errorf("%d Paarfälle, erwartet n(n+1)/2 × 2 = 90", faelle)
	}
}

// TestEingebauteRouteUeberlappt hält Z3: die eingebaute Route GET /healthz zählt als Route.
func TestEingebauteRouteUeberlappt(t *testing.T) {
	h := neueRoute().handle
	faelle := []struct {
		methode, muster string
		gueltig         bool
	}{{"GET", "/{x}", false}, {"GET", "/healthz", false}, {"POST", "/healthz", true}, {"GET", "/healthz/x", true}}
	for _, f := range faelle {
		k, _ := konfigMit(admin.Route{Methode: f.methode, Muster: f.muster, Handler: h})
		srv, err := admin.New(k)
		if f.gueltig != (err == nil && srv != nil) || (!f.gueltig && !errors.Is(err, model.ErrUngueltig)) {
			t.Errorf("%s %s: Fehler %v, erwartet gültig = %v", f.methode, f.muster, err, f.gueltig)
		}
	}
	if len(faelle) != 4 {
		t.Errorf("%d Fälle, erwartet 4", len(faelle))
	}
}
