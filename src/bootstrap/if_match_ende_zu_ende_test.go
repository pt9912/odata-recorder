package bootstrap_test

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// mitKopf ist die Anfrage a mit der Kopfzeile name und dem Wert wert.
func mitKopf(a abAnfrage, name, wert string) abAnfrage {
	a.Header = map[string][]string{name: {wert}}
	return a
}

// TestIfMatchHttpEndeZuEnde (F27, SPEC-283, SPEC-136 Nr. 4 und 7 a): über echtes HTTP und die echte SQLite-Transaktion: ein veralteter Tag ergibt 412 mit
// ODR-E6010 im Fehlerobjekt und ohne ETag, Zustand und ETag bleiben; der aktuelle Tag ergibt 204 mit dem neuen ETag; dieselbe Anfrage ein zweites Mal
// ist 412; If-None-Match * bei vorhandenem Eintrag ist 412; If-Match * lässt DELETE zu.
func TestIfMatchHttpEndeZuEnde(t *testing.T) {
	u := abFrisch(t)
	m, _ := abEtagMocks(t, u)
	basis := vertragServer(t, "if-match", &vertragSpiegel{}, m)
	c := vertragClient(t)
	sende := func(a abAnfrage) drahtAntwort {
		t.Helper()
		r, err := drahtSende(c, basis, a)
		if err != nil {
			t.Fatalf("%s %s: %v", a.Methode, a.Pfad, err)
		}
		return r
	}
	etag := func() string {
		t.Helper()
		r := sende(abG("/Products(1)", ""))
		if r.Status != 200 {
			t.Fatalf("GET Products(1): Status %d", r.Status)
		}
		return r.Header.Get("ETag")
	}
	abgewiesen := func(was string, r drahtAntwort) {
		t.Helper()
		if r.Status != 412 || vertragCode(r.Body) != "ODR-E6010" || len(r.Header.Values("ETag")) != 0 {
			t.Errorf("%s: Status %d Code %q ETag %v, erwartet 412 und ODR-E6010 ohne ETag", was, r.Status, vertragCode(r.Body), r.Header.Values("ETag"))
		}
	}
	if got := etag(); got != etagLiteral(0) {
		t.Fatalf("ETag vor der Änderung %q", got)
	}
	abgewiesen("PATCH mit veraltetem Tag", sende(mitKopf(abM("PATCH", "/Products(1)", `{"Name":"x"}`), "If-Match", etagLiteral(7))))
	if got := etag(); got != etagLiteral(0) {
		t.Errorf("ETag nach der abgewiesenen Änderung %q, erwartet %q", got, etagLiteral(0))
	}
	ok := sende(mitKopf(abM("PATCH", "/Products(1)", `{"Name":"x"}`), "If-Match", etagLiteral(0)))
	if ok.Status != 204 || ok.Header.Get("ETag") != etagLiteral(1) {
		t.Errorf("PATCH mit aktuellem Tag: Status %d ETag %q, erwartet 204 und %q", ok.Status, ok.Header.Get("ETag"), etagLiteral(1))
	}
	abgewiesen("dieselbe Anfrage ein zweites Mal", sende(mitKopf(abM("PATCH", "/Products(1)", `{"Name":"x"}`), "If-Match", etagLiteral(0))))
	abgewiesen("PUT mit If-None-Match *", sende(mitKopf(abM("PUT", "/Products(1)", `{"Id":1,"Name":"p"}`), "If-None-Match", "*")))
	abgewiesen("DELETE mit If-None-Match auf den eigenen Tag", sende(mitKopf(abM("DELETE", "/Products(1)", ""), "If-None-Match", etagLiteral(1))))
	if got := etag(); got != etagLiteral(1) {
		t.Errorf("ETag nach den abgewiesenen Anfragen %q, erwartet %q", got, etagLiteral(1))
	}
	if del := sende(mitKopf(abM("DELETE", "/Products(1)", ""), "If-Match", "*")); del.Status != 204 {
		t.Errorf("DELETE mit If-Match *: Status %d, erwartet 204", del.Status)
	}
}

// wettlaufSQLite fährt n Anfragen gleichzeitig gegen Products(1) über die echte SQLite-Transaktion und liefert die Zahl je Status und Code.
func wettlaufSQLite(t *testing.T, n int, anfrage func(k int) abAnfrage) (stati map[int]int, codes map[model.Code]int, u *abUmgebung) {
	t.Helper()
	u = abFrisch(t)
	m, _ := abEtagMocks(t, u)
	var mu sync.Mutex
	stati, codes = map[int]int{}, map[model.Code]int{}
	p := &abParallel{N: n}
	p.lauf(func(k int) {
		ist := abSende(context.Background(), m, "ws-1", anfrage(k))
		mu.Lock()
		defer mu.Unlock()
		if ist.F != nil {
			codes[ist.F.Code]++
			return
		}
		stati[ist.Status]++
	})
	return stati, codes, u
}

// TestIfMatchWettlaufSQLite (F15, F16, F18, SPEC-283 Nr. 8 a, b, d) über der echten SQLite-Transaktion: von 20 gleichzeitigen Änderungen mit demselben
// If-Match gelingt genau eine, die übrigen sind ODR-E6010 (bei DELETE ODR-E6008); der ETag danach ist der der Version 1. Form 3 (PATCH, PUT, Mischung) + DELETE,
// je 5 frische Aufbauten.
func TestIfMatchWettlaufSQLite(t *testing.T) {
	const n = 20
	formen := map[string]func(k int) abAnfrage{
		"PATCH": func(k int) abAnfrage {
			return mitKopf(abM("PATCH", "/Products(1)", `{"Name":"n`+strconv.Itoa(k)+`"}`), "If-Match", etagLiteral(0))
		},
		"PUT": func(k int) abAnfrage {
			return mitKopf(abM("PUT", "/Products(1)", `{"Id":1,"Name":"n`+strconv.Itoa(k)+`"}`), "If-Match", etagLiteral(0))
		},
		"Mischung": func(k int) abAnfrage {
			if k%2 == 0 {
				return mitKopf(abM("PATCH", "/Products(1)", `{"Name":"n`+strconv.Itoa(k)+`"}`), "If-Match", etagLiteral(0))
			}
			return mitKopf(abM("PUT", "/Products(1)", `{"Id":1,"Name":"n`+strconv.Itoa(k)+`"}`), "If-Match", etagLiteral(0))
		},
		"DELETE": func(int) abAnfrage { return mitKopf(abM("DELETE", "/Products(1)", ""), "If-Match", etagLiteral(0)) },
	}
	for name, anfrage := range formen {
		fehler := model.CodeE6010
		if name == "DELETE" {
			fehler = model.CodeE6008
		}
		for wdh := 0; wdh < 5; wdh++ {
			stati, codes, u := wettlaufSQLite(t, n, anfrage)
			if stati[204] != 1 || len(stati) != 1 || codes[fehler] != n-1 || len(codes) != 1 {
				t.Errorf("%s Wiederholung %d: Status %v Codes %v, erwartet ein 204 und %d × %s", name, wdh, stati, codes, n-1, fehler)
			}
			if name == "DELETE" {
				continue
			}
			m, _ := abEtagMocks(t, u)
			if got := etagVonAntwort(t, abSende(context.Background(), m, "ws-1", abG("/Products(1)", ""))); got != etagLiteral(1) {
				t.Errorf("%s Wiederholung %d: ETag danach %q, erwartet %q", name, wdh, got, etagLiteral(1))
			}
		}
	}
}
