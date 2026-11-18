package hybrid_test

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// TestHostFaenger hält Z20: der Referenzdienst empfängt unter dem Host (Verbindung und Host-Header), den die Policy
// freigab: GET ergibt eine Verbindung mit Methode GET und Query, POST mit Route eine mit Methode POST und dem Body
// byte-gleich; der Referenzdienst des anderen Dienstes bleibt ohne Verbindung, und eine Anfrage an den Dienst, dessen
// Host nicht freigegeben ist, erreicht keinen.
func TestHostFaenger(t *testing.T) {
	a, b := neuerReferenz(t), neuerReferenz(t)
	dienste := []hybrid.Dienst{
		{Root: "/odata/svc", UpstreamURL: "http://" + a.adresse()},
		{Root: "/odata/other", UpstreamURL: "http://" + b.adresse()},
	}
	pol := uPolicy(t, []string{a.adresse()}, []string{"POST /odata/svc/**", "POST /odata/other/**"})
	inhalt := muster(6, 4)

	g := rigMitClient(t, pol, dienste)
	g.mockE6003(t)
	got, err := g.v.Hybrid(context.Background(), anfrage("GET"))
	if err != nil || got.Status != 200 {
		t.Fatalf("GET: Fehler %v, Status %d", err, got.Status)
	}
	text, _ := io.ReadAll(got.Body)
	_ = got.Body.Close()
	if string(text) != "antwort" || !reflect.DeepEqual(kennzeichen(got), []string{"upstream"}) {
		t.Errorf("GET: Body %q, Kennzeichnung %v", text, kennzeichen(got))
	}
	r := a.alle()
	if a.verbindungen.Load() != 1 || len(r) != 1 || r[0].Methode != "GET" || r[0].Host != a.adresse() || r[0].RawQuery != "$top=1" || r[0].Pfad != "/Products" {
		t.Errorf("GET: Verbindungen %d, Requests %+v", a.verbindungen.Load(), r)
	}

	g = rigMitClient(t, pol, dienste)
	g.mockE6003(t)
	p := mitBody("POST", neuerStrom(inhalt), -1)
	if _, err := g.v.Hybrid(context.Background(), p); err != nil {
		t.Fatalf("POST: %v", err)
	}
	r = a.alle()
	if a.verbindungen.Load() != 2 || len(r) != 2 || r[1].Methode != "POST" || !reflect.DeepEqual(r[1].Body, inhalt) {
		t.Errorf("POST: Verbindungen %d, Requests %+v", a.verbindungen.Load(), r)
	}

	g = rigMitClient(t, pol, dienste)
	g.mockE6003(t)
	q := anfrage("GET")
	q.Pfad = "/odata/other/Products"
	if _, err := g.v.Hybrid(context.Background(), q); err == nil {
		t.Errorf("Host des anderen Dienstes ist nicht freigegeben, erwartet den Fehler von M")
	}
	if a.verbindungen.Load() != 2 || b.verbindungen.Load() != 0 {
		t.Errorf("Verbindungen A %d, B %d; erwartet 2 und 0", a.verbindungen.Load(), b.verbindungen.Load())
	}
}

// zustand ist ein Zustand der Policy der Sicherheitsachse; sendet ist das Orakel „darf senden“ nach SPEC-261 Nr. 4,
// unabhängig vom Code des Dienstes.
type zustand struct {
	name   string
	werte  func(m, host string) policy.Werte
	sendet func(m string, c string) bool
}

func routeFuer(m, pfad string) []string {
	if mutierendOrakel(m) {
		return []string{m + " " + pfad}
	}
	return nil
}

// TestKeineAnfrageOhneFreigabe hält Z21: unter den Policy-Zuständen ohne Freigabe erreicht keine Anfrage den
// Referenzdienst (Verbindungszähler), für jede Methode ohne lesende Freigabe und jeden Code; die Positivkontrolle
// (Route passt) sendet genau bei mutierender Methode und den drei Auslösern von M. Override-Header ändern die Methode
// nicht.
func TestKeineAnfrageOhneFreigabe(t *testing.T) {
	ref := neuerReferenz(t)
	host := ref.adresse()
	dienste := []hybrid.Dienst{{Root: "/odata/svc", UpstreamURL: "http://" + host}}
	nie := func(string, string) bool { return false }
	zustaende := []zustand{
		{"alle Schlüssel leer", func(m, h string) policy.Werte { return policy.Werte{} }, nie},
		{"Hauptschalter aus", func(m, h string) policy.Werte {
			return policy.Werte{Hosts: []string{h}, Schreibfreigaben: routeFuer(m, standardP)}
		}, nie},
		{"fremder Host", func(m, h string) policy.Werte {
			return policy.Werte{Upstream: true, Hosts: []string{"fremd.example"}, Schreibfreigaben: routeFuer(m, standardP)}
		}, nie},
		{"Route für anderen Pfad", func(m, h string) policy.Werte {
			return policy.Werte{Upstream: true, Hosts: []string{h}, Schreibfreigaben: routeFuer(m, "/odata/svc/Other")}
		}, nie},
		{"Route für andere Methode", func(m, h string) policy.Werte {
			anders := "POST"
			if m == "POST" {
				anders = "PUT"
			}
			return policy.Werte{Upstream: true, Hosts: []string{h}, Schreibfreigaben: routeFuer(anders, standardP)}
		}, nie},
		{"Replay-Kategorie gesetzt", func(m, h string) policy.Werte {
			return policy.Werte{
				Upstream: true, Hosts: []string{h}, Schreibfreigaben: routeFuer(m, standardP),
				ReplayKategorien: []string{katUnsupported, katCoverage},
			}
		}, nie},
		{"Route passt (Positivkontrolle)", func(m, h string) policy.Werte {
			return policy.Werte{Upstream: true, Hosts: []string{h}, Schreibfreigaben: routeFuer(m, standardP)}
		}, func(m, c string) bool {
			return mutierendOrakel(m) && (c == "ODR-E6001" || c == "ODR-E6002" || c == "ODR-E6003")
		}},
	}
	methoden := []string{"POST", "PUT", "PATCH", "DELETE", "MERGE", "OPTIONS", "TRACE", "get"}
	codes := []string{"ODR-E6001", "ODR-E6002", "ODR-E6003", "ODR-E5001"}
	n, sendeten := 0, 0
	for _, z := range zustaende {
		for _, m := range methoden {
			for _, c := range codes {
				w := z.werte(m, host)
				g := rigMitClient(t, neuePolicy(t, mitMiss(w, c)), dienste)
				if c == missCode {
					g.mockE6003(t)
					g.replayFehler(code(codeNamed(t, missCode)))
				} else {
					g.mockFehler(code(codeNamed(t, c)))
					g.replayFehler(code(codeNamed(t, missCode)))
				}
				vorher := ref.verbindungen.Load()
				_, _ = g.v.Hybrid(context.Background(), mitBody(m, neuerStrom(muster(3, 2)), 3))
				n++
				soll := z.sendet(m, c)
				if soll {
					sendeten++
				}
				if got := ref.verbindungen.Load() - vorher; (got == 1) != soll || got > 1 {
					t.Errorf("%s, %s, %s: %d Verbindungen, erwartet %v", z.name, m, c, got, soll)
				}
			}
		}
	}
	if n != len(zustaende)*len(methoden)*len(codes) || sendeten != 5*3 {
		t.Errorf("Zählprüfung: %d Fälle, %d mit Verbindung", n, sendeten)
	}

	// Override-Header ändern die Methode nicht.
	vorher := ref.verbindungen.Load()
	g := rigMitClient(t, uPolicy(t, []string{host}, nil), dienste)
	g.mockE6003(t)
	a := mitBody("POST", neuerStrom(muster(3, 2)), 3)
	a.Header["X-HTTP-Method"] = []string{"GET"}
	_, _ = g.v.Hybrid(context.Background(), a)
	if got := ref.verbindungen.Load() - vorher; got != 0 {
		t.Errorf("POST ohne Route mit X-HTTP-Method: GET: %d Verbindungen", got)
	}
	g = rigMitClient(t, uPolicy(t, []string{host}, nil), dienste)
	g.mockE6003(t)
	b := anfrage("GET")
	b.Header["X-HTTP-Method"] = []string{"DELETE"}
	if _, err := g.v.Hybrid(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	if r := ref.alle(); r[len(r)-1].Methode != "GET" {
		t.Errorf("GET mit X-HTTP-Method: DELETE kam als %s an", r[len(r)-1].Methode)
	}
}

// mitMiss ergänzt für den Fall ODR-E5001 (Fehler von R nach einem Mock-Fehler ODR-E6003) die Kategorie coverage, damit
// der Weg über R führt.
func mitMiss(w policy.Werte, c string) policy.Werte {
	if c == missCode && !slices.Contains(w.ReplayKategorien, katCoverage) {
		w.ReplayKategorien = append(slices.Clone(w.ReplayKategorien), katCoverage)
	}
	return w
}

// TestMethodenHeaderNieAmUpstream hält Z22: die drei Header zum Überschreiben der Methode gehen in keiner Schreibweise
// an den Referenzdienst, bei GET und bei POST mit Route; andere Header kommen mit allen Werten in ihrer Reihenfolge an.
func TestMethodenHeaderNieAmUpstream(t *testing.T) {
	ref := neuerReferenz(t)
	host := ref.adresse()
	dienste := []hybrid.Dienst{{Root: "/odata/svc", UpstreamURL: "http://" + host}}
	namen := []string{"X-HTTP-Method", "X-HTTP-Method-Override", "X-Method-Override"}
	andere := map[string][]string{
		"Authorization": {"Bearer a", "Bearer b"}, "Accept": {"application/json"}, "If-Match": {"W/\"1\""},
		"Content-Type": {"text/plain"}, "Accept-Encoding": {"br"},
	}
	n := 0
	for _, name := range namen {
		for _, schreibweise := range schreibweisen(name) {
			for _, m := range []string{"GET", "POST"} {
				n++
				g := rigMitClient(t, uPolicy(t, []string{host}, []string{"POST " + standardP}), dienste)
				g.mockE6003(t)
				a := anfrage(m)
				a.Header[schreibweise] = []string{"DELETE"}
				for k, v := range andere {
					a.Header[k] = v
				}
				if m == "POST" {
					a.Body, a.Laenge = neuerStrom(muster(3, 1)), 3
				}
				vorher := len(ref.alle())
				if _, err := g.v.Hybrid(context.Background(), a); err != nil {
					t.Errorf("%s %s: %v", m, schreibweise, err)
					continue
				}
				r := ref.alle()
				if len(r) != vorher+1 {
					t.Errorf("%s %s: %d Requests", m, schreibweise, len(r)-vorher)
					continue
				}
				e := r[len(r)-1]
				for _, nm := range namen {
					if _, da := e.Header[http.CanonicalHeaderKey(nm)]; da {
						t.Errorf("%s %s: Header %s kam an", m, schreibweise, nm)
					}
				}
				for k, v := range andere {
					if !reflect.DeepEqual(e.Header[http.CanonicalHeaderKey(k)], v) {
						t.Errorf("%s %s: %s kam als %v an, erwartet %v", m, schreibweise, k, e.Header[http.CanonicalHeaderKey(k)], v)
					}
				}
				if e.Methode != m {
					t.Errorf("%s %s: Methode %s", m, schreibweise, e.Methode)
				}
			}
		}
	}
	if n != 3*4*2 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}
