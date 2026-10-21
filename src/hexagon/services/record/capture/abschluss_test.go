package capture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// Die fünf Ergebnisse je Richtung, die Abschluss unterscheidet.
func ergebnisVariante(r capture.Richtung, v byte) capture.Ergebnis {
	praefix := map[capture.Richtung]string{capture.Request: "sha256:req", capture.Response: "sha256:resp"}[r]
	switch v {
	case 'A':
		return capture.Ergebnis{Richtung: r, Ausgang: capture.Vollstaendig, Body: model.BodyAbsent, Redaction: model.RedactionNone, Integrity: model.IntegrityAbsent}
	case 'N':
		return capture.Ergebnis{Richtung: r, Ausgang: capture.Vollstaendig, Body: model.BodyPresent, Blob: praefix + "N", Redaction: model.RedactionNone, Integrity: model.IntegrityVerified}
	case 'P':
		return capture.Ergebnis{Richtung: r, Ausgang: capture.Vollstaendig, Body: model.BodyPresent, Blob: praefix + "P", Redaction: model.RedactionApplied, Integrity: model.IntegrityVerified}
	case 'X':
		return capture.Ergebnis{Richtung: r, Ausgang: capture.Abbruch, Grund: capture.GrundAbbruch, Body: model.BodyDiscarded, Redaction: model.RedactionUnsafe, Integrity: model.IntegrityAbsent}
	}
	return capture.Ergebnis{Richtung: r, Ausgang: capture.Verstoss, Grund: capture.GrundSyntax, Body: model.BodyDiscarded, Redaction: model.RedactionUnsafe, Integrity: model.IntegrityAbsent}
}

type abschlussFall struct {
	paar        string
	capture     model.CaptureStatus
	key         bool
	integrity   model.BodyIntegrity
	equivalence model.ReplayEquivalence
}

// abschlussFaelle sind die 25 Paarungen von Hand: A absent, N present/none, P present/applied,
// X Abbruch, V Verstoß; erst Request, dann Response.
func abschlussFaelle() []abschlussFall {
	return []abschlussFall{
		{"AA", model.CaptureComplete, true, model.IntegrityAbsent, model.EquivalenceByte},
		{"AN", model.CaptureComplete, true, model.IntegrityVerified, model.EquivalenceByte},
		{"AP", model.CaptureComplete, true, model.IntegrityVerified, model.EquivalenceSemantic},
		{"AX", model.CaptureIncomplete, true, model.IntegrityAbsent, model.EquivalenceNone},
		{"AV", model.CaptureFailed, true, model.IntegrityAbsent, model.EquivalenceNone},
		{"NA", model.CaptureComplete, true, model.IntegrityVerified, model.EquivalenceByte},
		{"NN", model.CaptureComplete, true, model.IntegrityVerified, model.EquivalenceByte},
		{"NP", model.CaptureComplete, true, model.IntegrityVerified, model.EquivalenceSemantic},
		{"NX", model.CaptureIncomplete, true, model.IntegrityVerified, model.EquivalenceNone},
		{"NV", model.CaptureFailed, true, model.IntegrityVerified, model.EquivalenceNone},
		{"PA", model.CaptureComplete, true, model.IntegrityVerified, model.EquivalenceByte},
		{"PN", model.CaptureComplete, true, model.IntegrityVerified, model.EquivalenceByte},
		{"PP", model.CaptureComplete, true, model.IntegrityVerified, model.EquivalenceSemantic},
		{"PX", model.CaptureIncomplete, true, model.IntegrityVerified, model.EquivalenceNone},
		{"PV", model.CaptureFailed, true, model.IntegrityVerified, model.EquivalenceNone},
		{"XA", model.CaptureFailed, false, model.IntegrityAbsent, model.EquivalenceNone},
		{"XN", model.CaptureFailed, false, model.IntegrityVerified, model.EquivalenceNone},
		{"XP", model.CaptureFailed, false, model.IntegrityVerified, model.EquivalenceNone},
		{"XX", model.CaptureFailed, false, model.IntegrityAbsent, model.EquivalenceNone},
		{"XV", model.CaptureFailed, false, model.IntegrityAbsent, model.EquivalenceNone},
		{"VA", model.CaptureFailed, false, model.IntegrityAbsent, model.EquivalenceNone},
		{"VN", model.CaptureFailed, false, model.IntegrityVerified, model.EquivalenceNone},
		{"VP", model.CaptureFailed, false, model.IntegrityVerified, model.EquivalenceNone},
		{"VX", model.CaptureFailed, false, model.IntegrityAbsent, model.EquivalenceNone},
		{"VV", model.CaptureFailed, false, model.IntegrityAbsent, model.EquivalenceNone},
	}
}

func TestAbschluss(t *testing.T) {
	if len(abschlussFaelle()) != 25 {
		t.Fatalf("%d Paarungen, erwartet 25", len(abschlussFaelle()))
	}
	for _, f := range abschlussFaelle() {
		t.Run(f.paar, func(t *testing.T) {
			req, resp := ergebnisVariante(capture.Request, f.paar[0]), ergebnisVariante(capture.Response, f.paar[1])
			z, err := capture.Abschluss(req, resp)
			if err != nil {
				t.Fatal(err)
			}
			if z.Capture != f.capture || z.KeyMoeglich != f.key || z.Integrity != f.integrity || z.Equivalence != f.equivalence {
				t.Errorf("Zustand = %v/%v/%v/%v, erwartet %v/%v/%v/%v", z.Capture, z.KeyMoeglich, z.Integrity, z.Equivalence,
					f.capture, f.key, f.integrity, f.equivalence)
			}
			if z.RequestBody != req.Body || z.RequestBlob != req.Blob || z.RequestRedact != req.Redaction ||
				z.ResponseBody != resp.Body || z.ResponseBlob != resp.Blob || z.ResponseRedact != resp.Redaction {
				t.Errorf("Bodies, Blobs oder Redactions weichen vom Ergebnis ab: %+v", z)
			}
		})
	}
}

func TestAbschlussGueltig(t *testing.T) {
	kopf := model.Exchange{
		ID: "e1", SessionID: "s1", SequenceNo: 1, Method: "GET", Path: "/x",
		StartedAt: time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC), ResponseStatus: 200,
	}
	exchange := func(z capture.Zustand) model.Exchange {
		e := kopf
		if z.KeyMoeglich {
			e.MatchKey = "k"
		}
		e.RequestBody, e.RequestBlob, e.RequestRedact = z.RequestBody, z.RequestBlob, z.RequestRedact
		e.ResponseBody, e.ResponseBlob, e.ResponseRedact = z.ResponseBody, z.ResponseBlob, z.ResponseRedact
		e.Capture, e.Integrity, e.Equivalence = z.Capture, z.Integrity, z.Equivalence
		return e
	}
	for _, f := range abschlussFaelle() {
		t.Run(f.paar, func(t *testing.T) {
			z, err := capture.Abschluss(ergebnisVariante(capture.Request, f.paar[0]), ergebnisVariante(capture.Response, f.paar[1]))
			if err != nil {
				t.Fatal(err)
			}
			if err := exchange(z).Validate(); err != nil {
				t.Errorf("Validate: %v", err)
			}
		})
	}
	t.Run("Gegenprobe: complete mit verworfenem Body ist ungültig", func(t *testing.T) {
		z, _ := capture.Abschluss(ergebnisVariante(capture.Request, 'N'), ergebnisVariante(capture.Response, 'X'))
		z.Capture = model.CaptureComplete
		if err := exchange(z).Validate(); err == nil {
			t.Error("Validate lässt complete bei verworfenem Body zu")
		}
	})
}

func TestAbschlussArgumente(t *testing.T) {
	req, resp := ergebnisVariante(capture.Request, 'N'), ergebnisVariante(capture.Response, 'N')
	faelle := []struct {
		name      string
		req, resp capture.Ergebnis
	}{
		{"Nullwert als Request", capture.Ergebnis{}, resp},
		{"Nullwert als Response", req, capture.Ergebnis{}},
		{"vertauscht", resp, req},
		{"beide Request", req, req},
		{"beide Response", resp, resp},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			z, err := capture.Abschluss(f.req, f.resp)
			if !errorsIsUngueltig(err) || z != (capture.Zustand{}) {
				t.Errorf("Abschluss = %+v, %v; erwartet ErrUngueltig", z, err)
			}
		})
	}
	// Ein Ergebnis fremder Herkunft prüft Abschluss nur auf die Richtung (SPEC-119 Nr. 12).
	t.Run("fremde Ergebnisse", func(t *testing.T) {
		fremd := []struct {
			name      string
			req, resp capture.Ergebnis
			capture   model.CaptureStatus
			key       bool
		}{
			{"present ohne Blob", capture.Ergebnis{Richtung: capture.Request, Ausgang: capture.Vollstaendig, Body: model.BodyPresent}, resp, model.CaptureComplete, true},
			{"Request ohne Ausgang", capture.Ergebnis{Richtung: capture.Request}, resp, model.CaptureFailed, false},
			{"Response ohne Ausgang", req, capture.Ergebnis{Richtung: capture.Response}, model.CaptureComplete, true},
		}
		for _, f := range fremd {
			z, err := capture.Abschluss(f.req, f.resp)
			if err != nil || z.Capture != f.capture || z.KeyMoeglich != f.key {
				t.Errorf("%s: Abschluss = %+v, %v; erwartet kein Fehler, %v und Key %v", f.name, z, err, f.capture, f.key)
			}
		}
	})
	t.Run("rein", func(t *testing.T) {
		req2, resp2 := req, resp
		a, _ := capture.Abschluss(req, resp)
		b, _ := capture.Abschluss(req, resp)
		if a != b || !reflect.DeepEqual(req, req2) || !reflect.DeepEqual(resp, resp2) {
			t.Errorf("Abschluss ist nicht rein: %+v, %+v", a, b)
		}
	})
}

// testNamen sammelt die Namen der Testfunktionen des Pakets.
func testNamen(t *testing.T) map[string]bool {
	t.Helper()
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	namen := map[string]bool{}
	for _, e := range eintraege {
		if !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "Test") {
				namen[fn.Name.Name] = true
			}
		}
	}
	return namen
}

// TestFixturesSPEC053 führt die Sicherheits-Fixtures aus SPEC-053 und prüft, dass die Tests
// existieren, die sie im Einzelnen fangen. Eine Fixture ist nicht Vollstaendig, außer dem Token
// über Chunk-Grenzen: es ist vollständig und redigiert.
func TestFixturesSPEC053(t *testing.T) {
	geheimBody := `{"password":"` + geheim + `","a":1}`
	namen := testNamen(t)
	fixtures := []struct {
		name  string
		tests []string
		lauf  func(t *testing.T) (*umg, capture.Ergebnis)
		aus   capture.Ausgang
		grund capture.Grund
	}{
		{"Token über Chunk-Grenzen", []string{"TestZerlegungOhneWirkung"}, func(t *testing.T) (*umg, capture.Ergebnis) {
			u := neu(t)
			return u, u.lauf(zerlegt([]byte(geheimBody), 1), antwort())
		}, capture.Vollstaendig, ""},
		{"gzip-Bombe", []string{"TestBombe", "TestEntpacktGrenze"}, func(t *testing.T) (*umg, capture.Ergebnis) {
			u := neu(t, mitMax(8<<20))
			return u, u.lauf(quelleBytes(gz(jsonString(1<<20))), mitEncoding(antwort(), "gzip"))
		}, capture.Verstoss, capture.GrundBombe},
		{"Kette", []string{"TestKette", "TestKetteReihenfolge"}, func(t *testing.T) (*umg, capture.Ergebnis) {
			u := neu(t)
			return u, u.lauf(quelleBytes(gz(geheimBody)), mitEncoding(antwort(), "gzip,gzip,gzip"))
		}, capture.Verstoss, capture.GrundEncoding},
		{"MIME-Mismatch", []string{"TestGrundTabelle"}, func(t *testing.T) (*umg, capture.Ergebnis) {
			u := neu(t)
			return u, u.lauf(quelle(geheimBody), xmlMetadata())
		}, capture.Verstoss, capture.GrundMismatch},
		{"Body-Abbruch", []string{"TestAbbruchUrsachen"}, func(t *testing.T) (*umg, capture.Ergebnis) {
			u := neu(t)
			return u, u.lauf(quelleMit(errQuelle, geheimBody[:25]), antwort())
		}, capture.Abbruch, capture.GrundAbbruch},
		{"langsamer Consumer", []string{"TestReadWartetNichtAufSchreiber"}, func(t *testing.T) (*umg, capture.Ergebnis) {
			u := neu(t, mitQueue(100))
			u.blobs.haltWrite = 1
			a := u.tee(quelle(geheimBody[:30], wiederholt(" ", 400)), antwort())
			_, _ = a.Read(make([]byte, 64))
			warteAufAnkunft(t, u.blobs.imWrite)
			_, _ = a.Read(make([]byte, 400))
			close(u.blobs.freiWrite)
			return u, abwarten(t, a)
		}, capture.Verstoss, capture.GrundUeberlauf},
		{"Queue-Überlauf", []string{"TestQueueHoechststand"}, func(t *testing.T) (*umg, capture.Ergebnis) {
			f := captureFaelle()
			for _, c := range f {
				if c.name == "ueberlauf" {
					return fuehreAus(t, c)
				}
			}
			t.Fatal("Zeile ueberlauf fehlt")
			return nil, capture.Ergebnis{}
		}, capture.Verstoss, capture.GrundUeberlauf},
		{"Redaction-Fehler", []string{"TestGrundTabelle"}, func(t *testing.T) (*umg, capture.Ergebnis) {
			u := neu(t)
			return u, u.lauf(quelle(`{"password":1}`), antwort())
		}, capture.Verstoss, capture.GrundNichtstring},
		{"Middleware-Logging", []string{"TestKeineDateiUndKeinLog"}, nil, "", ""},
	}
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			for _, n := range f.tests {
				if !namen[n] {
					t.Errorf("der Test %s existiert nicht", n)
				}
			}
			if f.lauf == nil {
				return
			}
			u, e := f.lauf(t)
			if e.Ausgang != f.aus || e.Grund != f.grund {
				t.Errorf("Ergebnis = %v/%q, erwartet %v/%q", e.Ausgang, e.Grund, f.aus, f.grund)
			}
			if strings.Contains(u.blobs.alleAngeboten(), geheim) {
				t.Errorf("der Writer sah den geheimen Wert")
			}
		})
	}
}
