package mock_test

import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// geloeschtWs ist ein Workspace-Repository, dem der Workspace in der Transaktion fehlt.
type geloeschtWs struct{ outbound.WorkspaceRepository }

func (geloeschtWs) Get(context.Context, string) (model.Workspace, error) {
	return model.Workspace{}, outbound.ErrNichtGefunden
}

// zweigM24 ist ein Zweig der Z-Liste (M24): Aufbau, Request je Lauf und Soll (Code, Status oder Kontextfehler).
type zweigM24 struct {
	name    string
	prep    func(u *umg)
	anf     func(idx int) inbound.Anfrage
	code    model.Code
	status  int
	kontext bool
}

// seedM24 legt die Einträge an, die die Zweige der Erfolge brauchen: je Lauf ein eigener Schlüssel (idx 0 bis 259), vorhandene
// Einträge ab 11, Tombstones ab 2000, partielle Einträge ab 3000, dazu 5 (Tombstone) und 6 (partial).
func seedM24(u *umg) {
	es := []model.EntityState{voll(1, "a"), voll(2, "b"), voll(3, "c")}
	tomb := voll(5, "t")
	tomb.Deleted, tomb.Version = true, 3
	es = append(es, tomb, teilEintrag("6", 2))
	for i := 0; i < 260; i++ {
		es = append(es, voll(11+i, "v"))
		t := voll(2000+i, "t")
		t.Deleted, t.Version = true, 3
		es = append(es, t, teilEintrag(strconv.Itoa(3000+i), 2))
	}
	u.en.setze(setProducts, es...)
}

// zweigeM24 sind die 36 Zweige der Z-Liste (M24) in den fünf Gruppen von SPEC-192; jeder erreicht eine Goroutine bei einem Aufruf
// je Lauf nur einmal.
func zweigeM24() []zweigM24 {
	je := func(m string, basis int) func(int) inbound.Anfrage {
		return func(idx int) inbound.Anfrage { return dynAnfrage(m, strconv.Itoa(basis+idx), "neu", wsID) }
	}
	fest := func(m, key string) func(int) inbound.Anfrage {
		return func(int) inbound.Anfrage { return dynAnfrage(m, key, "neu", wsID) }
	}
	leserFehler := func(ur model.EingabeUrsache) func(u *umg) {
		return func(u *umg) { u.p.setzeFehler("Eingabe.LeseEntity", &model.EingabeFehler{Ursache: ur}) }
	}
	nichts := func(*umg) {}
	return []zweigM24{
		// Methode und Ressource (4)
		{name: "Paarung", prep: func(u *umg) {
			u.c.url.f = func(string, string) (model.ODataURL, error) {
				return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceEntity, EntitySet: setProducts, Schluessel: []model.SchluesselTeil{zahl("", "2")}}}, nil
			}
		}, anf: fest("POST", "2"), code: e6002},
		{name: "$filter", prep: func(u *umg) {
			u.c.url.f = func(string, string) (model.ODataURL, error) {
				return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceEntity, EntitySet: setProducts, Schluessel: []model.SchluesselTeil{zahl("", "2")}}, Optionen: model.Optionen{Filter: idGroesser("1")}}, nil
			}
		}, anf: fest("PATCH", "2"), code: e6002},
		{name: "ohne Eingabe", prep: func(u *umg) { s := u.c.satz(false, "4.0"); s.Eingabe = nil; u.k.Dienste[0].Codecs = s }, anf: fest("POST", "7"), code: e6002},
		{name: "NurLesend", prep: func(u *umg) { u.k.Dienste[0].Codecs = u.c.satz(true, "4.0") }, anf: fest("POST", "7"), code: e6006},
		// Zuordnung und Ressourcen (8)
		{name: "kein Dienst", prep: nichts, anf: func(int) inbound.Anfrage { a := dynAnfrage("POST", "7", "n", wsID); a.Pfad = "/anderer/x"; return a }, code: model.CodeE5005},
		{name: "Header fehlt", prep: nichts, anf: func(int) inbound.Anfrage {
			a := dynAnfrage("POST", "7", "n", wsID)
			delete(a.Header, "X-Recorder-Workspace")
			return a
		}, code: model.CodeE5003},
		{name: "Workspace fremd", prep: func(u *umg) {
			u.ws.setze(model.Workspace{ID: wsID, ServiceID: "anderer", BaselineSnapshotID: baselineID})
		}, anf: fest("POST", "7"), code: model.CodeE5004},
		{name: "URL nicht unterstützt", prep: func(u *umg) {
			u.c.url.f = func(string, string) (model.ODataURL, error) { return model.ODataURL{}, model.ErrURLNichtUnterstuetzt }
		}, anf: fest("POST", "7"), code: model.CodeE6001},
		{name: "Blob fehlt", prep: func(u *umg) { u.p.setzeFehler("Blobs.Open", outbound.ErrBlobFehlt) }, anf: fest("POST", "7"), code: model.CodeE3001},
		{name: "Metadaten ungültig", prep: func(u *umg) { u.p.setzeFehler("MetadatenCodec.ParseMetadata", model.ErrMetadatenUngueltig) }, anf: fest("POST", "7"), code: model.CodeE6004},
		{name: "Set unbekannt", prep: func(u *umg) {
			u.c.url.f = func(string, string) (model.ODataURL, error) {
				return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: "Unbekannt"}}, nil
			}
		}, anf: fest("POST", "7"), code: e6008},
		{name: "Schlüssel ungültig", prep: func(u *umg) {
			u.c.url.f = func(string, string) (model.ODataURL, error) {
				return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceEntity, EntitySet: setProducts, Schluessel: []model.SchluesselTeil{zahl("Falsch", "2")}}}, nil
			}
		}, anf: fest("PATCH", "2"), code: e6005},
		// Body (5)
		{name: "Fehler ungueltig", prep: leserFehler(model.EingabeSyntax), anf: fest("POST", "7"), code: e6007},
		{name: "Fehler nicht_unterstuetzt", prep: leserFehler(model.EingabeNavigation), anf: fest("POST", "7"), code: e6002},
		{name: "Widerspruch", prep: nichts, anf: func(int) inbound.Anfrage {
			a := dynAnfrage("PATCH", "2", "neu", wsID)
			a.Body, a.Laenge = io.NopCloser(strings.NewReader("3|neu")), 5
			return a
		}, code: e6007},
		{name: "reservierter Text", prep: nichts, anf: func(int) inbound.Anfrage { return dynAnfrage("PATCH", "2", model.Marker, wsID) }, code: e6007},
		{name: "leerer Schlüssel", prep: nichts, anf: func(int) inbound.Anfrage {
			a := dynAnfrage("POST", "7", "neu", wsID)
			a.Body, a.Laenge = io.NopCloser(strings.NewReader("|neu")), 4
			return a
		}, code: e9001},
		// Transaktion (13)
		{name: "Workspace gelöscht", prep: func(u *umg) {
			u.tx.innen = func(s outbound.WorkspaceStores) outbound.WorkspaceStores { s.Workspaces = geloeschtWs{}; return s }
		}, anf: fest("PATCH", "2"), code: e7001},
		{name: "Duplikat", prep: nichts, anf: fest("POST", "2"), code: e6009},
		{name: "Tombstone", prep: nichts, anf: fest("PATCH", "5"), code: e6008},
		{name: "fehlend complete", prep: nichts, anf: fest("DELETE", "7000"), code: e6008},
		{name: "fehlend partial", prep: func(u *umg) { u.co.setze(baselineID, setProducts, model.CoveragePartial) }, anf: fest("DELETE", "7000"), code: e6003},
		{name: "PATCH auf partial", prep: nichts, anf: fest("PATCH", "6"), code: e6003},
		{name: "Typ", prep: func(u *umg) {
			u.ei.f = func(_ context.Context, _ io.Reader, _ []string, _ string, _ model.EingabeArt) (model.EntityEingabe, error) {
				e := eingabeIDName("2", "neu")
				e.Typ = typSpecial
				return e, nil
			}
		}, anf: fest("PATCH", "2"), code: e6007},
		{name: "Coverage-Status ungültig", prep: func(u *umg) { u.co.setze(baselineID, setProducts, model.CoverageStatus("gibt-es-nicht")) }, anf: fest("DELETE", "7000"), code: e9001},
		{name: "ErrRevisionVeraltet", prep: func(u *umg) { u.p.setzeFehler("Workspaces.Advance", outbound.ErrRevisionVeraltet) }, anf: fest("PATCH", "2"), code: e9001},
		{name: "Rollback nach Advance", prep: func(u *umg) { u.p.setzeFehler("Schreiber.SchreibeAnzahl", model.ErrUngueltig) }, anf: fest("PATCH", "2"), code: e9001},
		{name: "leeres Kennzeichen", prep: func(u *umg) { u.c.sch.kopf = model.Darstellung{} }, anf: fest("PATCH", "2"), code: e9001},
		{name: "Kontextabbruch in f", prep: func(u *umg) { u.p.setzeFehler("Entities.Get", context.Canceled) }, anf: fest("PATCH", "2"), kontext: true},
		{name: "Fehler von Put", prep: func(u *umg) { u.p.setzeFehler("Entities.Put", outbound.ErrBeschaedigt) }, anf: fest("PATCH", "2"), code: model.CodeE3001},
		// Erfolg (6)
		{name: "POST", prep: nichts, anf: je("POST", 1000), status: 201},
		{name: "PATCH", prep: nichts, anf: je("PATCH", 11), status: 204},
		{name: "PUT", prep: nichts, anf: je("PUT", 11), status: 204},
		{name: "DELETE", prep: nichts, anf: je("DELETE", 11), status: 204},
		{name: "POST über Tombstone", prep: nichts, anf: je("POST", 2000), status: 201},
		{name: "PUT auf partial", prep: nichts, anf: je("PUT", 3000), status: 204},
	}
}

// passt sagt, ob ein Ergebnis das Soll des Zweigs erfüllt.
func (z zweigM24) passt(a inbound.Antwort, err error) bool {
	switch {
	case z.kontext:
		return err != nil && fehlercode(err) == 0 && (err == context.Canceled || err == context.DeadlineExceeded)
	case z.code != 0:
		f, ok := err.(*inbound.Fehler)
		return ok && f.Code == z.code && f.Status == z.code.Stati()[0]
	}
	return err == nil && a.Status == z.status
}

// TestZweigeGleichzeitigMutation (W75, M24): jeder der 36 Zweige läuft in Runden aus 64 Goroutinen gleichzeitig gegen einen Dienst,
// die Antwort ist gleich der des Einzelaufrufs; 4 Runden, 256 Aufrufe je Zweig und Prozess. `make test-race` zählt das Rennen je
// Prozess; der Wirkungs-Fänger zählt abweichende Antworten.
func TestZweigeGleichzeitigMutation(t *testing.T) {
	zweige := zweigeM24()
	if len(zweige) != 36 {
		t.Fatalf("%d Zweige, erwartet 36", len(zweige))
	}
	for _, z := range zweige {
		u := schreibUmg(t)
		u.dynamisch()
		seedM24(u)
		z.prep(u)
		m := u.mocker()
		a, err := m.Mock(u.ctx, z.anf(256))
		if !z.passt(a, err) {
			t.Errorf("%s: Einzelaufruf %d %v, kein Soll", z.name, a.Status, err)
			continue
		}
		var falsch atomic.Int64
		for runde := 0; runde < 4; runde++ {
			aufEinmal(64, func(i int) {
				if a, err := m.Mock(u.ctx, z.anf(runde*64+i)); !z.passt(a, err) {
					falsch.Add(1)
				}
			})
		}
		if n := falsch.Load(); n != 0 {
			t.Errorf("%s: %d von 256 Antworten weichen vom Einzelaufruf ab", z.name, n)
		}
	}
}
