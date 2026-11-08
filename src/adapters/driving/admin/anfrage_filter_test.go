package admin_test

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// a7Wertform ist eine Form der Werteliste eines Headers.
type a7Wertform struct {
	name  string
	werte []string
}

// a7Wertformen sind die 8 Wertformen von M-Anfrage: ein Wert, zwei verschiedene, zwei gleichlautende, zwei mit leerem
// ersten, zwei mit leerem zweiten, drei, ein leerer Wert und die leere Liste.
func a7Wertformen() []a7Wertform {
	return []a7Wertform{
		{"ein Wert", []string{"v"}},
		{"zwei verschiedene", []string{"a", "b"}},
		{"zwei gleichlautende", []string{"a", "a"}},
		{"zwei, der erste leer", []string{"", "b"}},
		{"zwei, der zweite leer", []string{"a", ""}},
		{"drei", []string{"a", "b", "c"}},
		{"ein leerer Wert", []string{""}},
		{"leere Liste", []string{}},
	}
}

// requestMit bildet einen Request mit genau diesem Header, ohne Netz und ohne Kanonisierung.
func requestMit(kopf http.Header) *http.Request {
	return &http.Request{Method: "GET", URL: &url.URL{Path: "/a"}, Header: kopf}
}

// TestAnfrageFilterFunktion hält A7 (a) auf der Funktion anfrage: in keiner der 8192 Schreibweisen des Schlüssels und bei
// keiner der 8 Wertformen steht eine Zeile von Authorization in Header (65536 Fälle); jeder der 48 Nachbarnamen bleibt mit
// derselben Wertliste, auch der leeren (384 Fälle). Orakel: der Schlüssel in Kleinbuchstaben gegen den Namen.
func TestAnfrageFilterFunktion(t *testing.T) {
	schreibweisen, nachbarn, formen := z8Schreibweisen(), z8Nachbarn(), a7Wertformen()
	faelleAuth, faelleNachbar := 0, 0
	for _, s := range schreibweisen {
		for _, f := range formen {
			faelleAuth++
			a := admin.BildeAnfrage(requestMit(http.Header{s: f.werte, "X-Keep": {"1"}}), "id", map[string]string{})
			for k := range a.Header {
				if strings.ToLower(k) == "authorization" {
					t.Fatalf("Schlüssel %q, %s: Header trägt %q", s, f.name, k)
				}
			}
			if len(a.Header) != 1 || !slices.Equal(a.Header["X-Keep"], []string{"1"}) {
				t.Fatalf("Schlüssel %q, %s: Header %v, erwartet allein X-Keep", s, f.name, a.Header)
			}
		}
	}
	for _, n := range nachbarn {
		for _, f := range formen {
			faelleNachbar++
			a := admin.BildeAnfrage(requestMit(http.Header{n: f.werte, "Authorization": {"Bearer x"}}), "id", map[string]string{})
			werte, da := a.Header[n]
			if !da || !slices.Equal(werte, f.werte) || len(a.Header) != 1 {
				t.Fatalf("Nachbar %q, %s: Header %v, erwartet allein %q mit %q", n, f.name, a.Header, n, f.werte)
			}
		}
	}
	if len(schreibweisen) != 8192 || len(formen) != 8 || faelleAuth != 65536 || len(nachbarn) != 48 || faelleNachbar != 384 {
		t.Errorf("%d Schreibweisen, %d Wertformen, %d und %d Fälle, erwartet 8192, 8, 65536 und 384", len(schreibweisen),
			len(formen), faelleAuth, faelleNachbar)
	}
}

// TestAnfrageFilterServeHTTP hält A7 (b) über ServeHTTP: neben der gültigen kanonischen Zeile trägt der Request in jeder
// der 8192 Schreibweisen einen Schlüssel mit dem Canary; die Route läuft, Header trägt weder Schlüssel noch Canary noch
// Token. Die 48 Nachbarn mit leerem Wert und mit leerer Liste bleiben.
func TestAnfrageFilterServeHTTP(t *testing.T) {
	gueltig := bearer(string(tokenGueltig()))
	faelle := 0
	for _, s := range z8Schreibweisen() {
		faelle++
		route := neueRoute()
		u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		kopf := map[string][]string{"Authorization": gueltig["Authorization"]}
		if s != "Authorization" {
			kopf[s] = []string{"Bearer " + canary()}
		}
		rec := u.rufe(ohneAuth("GET", "/a", kopf))
		a, ok := route.einzige()
		if !ok || rec.Code != 200 {
			t.Fatalf("Schlüssel %q: Status %d, Route %d Aufrufe", s, rec.Code, route.aufrufe())
		}
		if len(a.Header) != 0 {
			t.Fatalf("Schlüssel %q: Header %v, erwartet leer", s, a.Header)
		}
	}
	if faelle != 8192 {
		t.Errorf("%d Schreibweisen, erwartet 8192", faelle)
	}
	for _, n := range z8Nachbarn() {
		for _, f := range []a7Wertform{{"leerer Wert", []string{""}}, {"leere Liste", []string{}}} {
			route := neueRoute()
			u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
			u.rufe(ohneAuth("GET", "/a", map[string][]string{"Authorization": gueltig["Authorization"], n: f.werte}))
			a, ok := route.einzige()
			werte, da := a.Header[n]
			if !ok || !da || !slices.Equal(werte, f.werte) || len(a.Header) != 1 {
				t.Errorf("Nachbar %q, %s: Header %v, erwartet allein %q mit %q", n, f.name, a.Header, n, f.werte)
			}
		}
	}
}
