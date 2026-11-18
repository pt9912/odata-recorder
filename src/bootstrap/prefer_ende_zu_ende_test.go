package bootstrap_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// preferMock baut den Prüfling über der Datei und der Ablage von u mit dem echten Schreiber, der echten Eingabe und der echten SQLite-Transaktion; etag setzt Dienst.Etag.
func preferMock(t *testing.T, u *abUmgebung, etag bool) *mock.Mocker {
	t.Helper()
	m, err := mock.New(mock.Konfig{
		Dienste: []mock.Dienst{{ID: "svc", Root: "/odata", Etag: etag}}, Metadaten: u.db.Metadata(), Workspaces: u.db.Workspaces(), Entities: u.db.Entities(),
		Coverage: u.db.Coverage(), Blobs: u.store, MetadatenCodec: csdl.New(), URLCodec: urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser())),
		Schreiber: jsoncodec.NewSchreiber(), Eingabe: jsoncodec.NewEingabe(), Transaktion: u.db, PublicBaseURL: "http://recorder.test", MaxBytes: 8388608,
	})
	abMust(t, err)
	return m
}

// TestPreferEntityGleichGet (Z18, SPEC-288 Nr. 1): der Body einer Antwort mit Entity ist, bei gleichem `Accept` und ohne `$select`, byte-gleich dem Body eines GET derselben Entity
// unmittelbar danach, auch der ETag der Kopfzeile; er entsteht aus dem geschriebenen Eintrag, nicht aus dem Body der Anfrage. Fall 6 (`PATCH` mit Teilbody, `PATCH` mit `{}`,
// `PUT` mit fehlender nullbarer Property, `PUT` mit allen Properties, `POST`, `POST` mit `return=representation`) × Accept 3 × Etag 2 = 36 mit echtem Schreiber.
func TestPreferEntityGleichGet(t *testing.T) {
	accepts := [][]string{nil, {"application/json;odata.metadata=none"}, {"application/json;odata.metadata=minimal"}}
	faelle := []struct {
		name, methode, pfad, body, prefer, get string
	}{
		{"PATCH Teilbody", "PATCH", "/Products(1)", `{"Name":"neu"}`, "return=representation", "/Products(1)"},
		{"PATCH leer", "PATCH", "/Products(2)", `{}`, "return=representation", "/Products(2)"},
		{"PUT ohne nullbare Property", "PUT", "/Products(3)", `{"Id":3,"Name":"p"}`, "return=representation", "/Products(3)"},
		{"PUT mit allen Properties", "PUT", "/Products(4)", `{"Id":4,"Name":"q","Active":true,"Stock":1,"Price":2.5}`, "return=representation", "/Products(4)"},
		{"POST ohne Wahl", "POST", "/Products", `{"Id":20,"Name":"n"}`, "", "/Products(20)"},
		{"POST mit Wahl", "POST", "/Products", `{"Id":21,"Name":"n"}`, "return=representation", "/Products(21)"},
	}
	n := 0
	for _, etag := range []bool{false, true} {
		for _, acc := range accepts {
			u := abFrisch(t)
			m := preferMock(t, u, etag)
			for _, f := range faelle {
				n++
				kopf := map[string][]string{}
				if acc != nil {
					kopf["Accept"] = acc
				}
				if f.prefer != "" {
					kopf["Prefer"] = []string{f.prefer}
				}
				ist := abSende(context.Background(), m, "ws-1", abAnfrage{Methode: f.methode, Pfad: f.pfad, Body: f.body, Header: kopf})
				get := abSende(context.Background(), m, "ws-1", abAnfrage{Methode: "GET", Pfad: f.get, Header: map[string][]string{"Accept": acc}})
				kontext := f.name + " Etag " + map[bool]string{false: "aus", true: "an"}[etag] + " Accept " + strings.Join(acc, ",")
				if ist.Err != nil || get.Err != nil || get.Status != 200 {
					t.Errorf("%s: Fehler %v, GET %d %v", kontext, ist.Err, get.Status, get.Err)
					continue
				}
				wantStatus := map[string]int{"POST": 201, "PATCH": 200, "PUT": 200}[f.methode]
				if ist.Status != wantStatus || ist.Body != get.Body {
					t.Errorf("%s: Status %d, erwartet %d; Body\n%s\nGET danach\n%s", kontext, ist.Status, wantStatus, ist.Body, get.Body)
				}
				if etag && (len(ist.Header["Etag"]) != 1 || !strings.EqualFold(strings.Join(ist.Header["Etag"], ""), strings.Join(get.Header["Etag"], ""))) {
					t.Errorf("%s: ETag %v, GET danach %v", kontext, ist.Header["Etag"], get.Header["Etag"])
				}
				if !etag && (len(ist.Header["Etag"]) != 0 || strings.Contains(ist.Body, "@odata.etag")) {
					t.Errorf("%s: ETag trotz Dienst.Etag aus", kontext)
				}
			}
		}
	}
	if n != 36 {
		t.Errorf("Menge %d, erwartet 36", n)
	}
}

// TestPreferMinimalEntityId (SPEC-286 Nr. 1): `return=minimal` ergibt `204` ohne Body mit `OData-EntityId` (und bei `POST` `Location`) in derselben Zeichenfolge, die ein folgendes GET als
// `@odata.id` nennt; der Wert beginnt mit der öffentlichen Basis des Dienstes. Methode 3 × Etag 2 = 6 mit echtem Schreiber.
func TestPreferMinimalEntityId(t *testing.T) {
	n := 0
	for _, etag := range []bool{false, true} {
		u := abFrisch(t)
		m := preferMock(t, u, etag)
		for _, f := range []struct{ methode, pfad, body, id string }{
			{"PATCH", "/Products(1)", `{"Name":"x"}`, "/Products(1)"}, {"PUT", "/Products(2)", `{"Id":2,"Name":"y"}`, "/Products(2)"},
			{"POST", "/Products", `{"Id":30,"Name":"z"}`, "/Products(30)"},
		} {
			n++
			ist := abSende(context.Background(), m, "ws-1", abAnfrage{Methode: f.methode, Pfad: f.pfad, Body: f.body, Header: map[string][]string{"Prefer": {"return=minimal"}}})
			get := abSende(context.Background(), m, "ws-1", abAnfrage{Methode: "GET", Pfad: f.id, Header: map[string][]string{"Accept": {"application/json;odata.metadata=minimal"}}})
			want := "http://recorder.test/odata" + f.id
			if ist.Err == nil && ist.Antwort.Body == nil {
				t.Errorf("%s: Body nil, der leere Strom ist der Body der 204 (SPEC-286 Nr. 1)", f.methode)
			}
			if ist.Err != nil || ist.Status != 204 || ist.Body != "" || !strings.Contains(get.Body, `"@odata.id":"`+want+`"`) {
				t.Errorf("%s: %d %q %v; GET %q", f.methode, ist.Status, ist.Body, ist.Err, get.Body)
			}
			if got := ist.Header["Odata-Entityid"]; len(got) != 1 || got[0] != want {
				t.Errorf("%s: OData-EntityId %v, erwartet %s", f.methode, got, want)
			}
			if got := ist.Header["Location"]; (f.methode == "POST") != (len(got) == 1 && got[0] == want) || f.methode != "POST" && got != nil {
				t.Errorf("%s: Location %v", f.methode, got)
			}
			if got := ist.Header["Preference-Applied"]; len(got) != 1 || got[0] != "return=minimal" {
				t.Errorf("%s: Preference-Applied %v", f.methode, got)
			}
		}
	}
	if n != 6 {
		t.Errorf("Menge %d, erwartet 6", n)
	}
}

// TestPreferWiederholungGleichGet (SPEC-287 Nr. 7, SPEC-289 Nr. 1): dieselbe Mutation zweimal sind zwei Mutationen, und die Antwort mit Entity zeigt je den Eintrag ihrer Mutation:
// der Body ist byte-gleich dem GET unmittelbar danach und trägt den Namen der eigenen, nicht der vorigen Mutation. Methode (PATCH, PUT) × Etag 2 × Mutation 2 = 8 mit echtem Schreiber.
func TestPreferWiederholungGleichGet(t *testing.T) {
	n := 0
	for _, etag := range []bool{false, true} {
		u := abFrisch(t)
		m := preferMock(t, u, etag)
		for _, methode := range []string{"PATCH", "PUT"} {
			for _, name := range []string{"erste", "zweite"} {
				n++
				body := `{"Name":"` + name + `"}`
				if methode == "PUT" {
					body = `{"Id":1,"Name":"` + name + `"}`
				}
				kontext := methode + " " + name + " Etag " + map[bool]string{false: "aus", true: "an"}[etag]
				ist := abSende(context.Background(), m, "ws-1", abAnfrage{Methode: methode, Pfad: "/Products(1)", Body: body, Header: map[string][]string{"Prefer": {"return=representation"}}})
				get := abSende(context.Background(), m, "ws-1", abAnfrage{Methode: "GET", Pfad: "/Products(1)"})
				if ist.Err != nil || get.Err != nil || ist.Status != 200 || ist.Body != get.Body || !strings.Contains(ist.Body, `"Name":"`+name+`"`) {
					t.Errorf("%s: %d %q %v; GET %q", kontext, ist.Status, ist.Body, ist.Err, get.Body)
				}
			}
		}
	}
	if n != 8 {
		t.Errorf("Menge %d, erwartet 8", n)
	}
}
