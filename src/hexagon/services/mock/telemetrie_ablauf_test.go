package mock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// ohneSpanFall ist eine Anfrage, die an einem der Schritte (a) bis (f) von SPEC-195 Nr. 3 endet oder `$metadata` ist: kein Span.
type ohneSpanFall struct {
	name    string
	bau     func(u *umg, a *inbound.Anfrage)
	keinCtx bool
	// code ist der Code des Fehlers; 0 mit ok ist die Antwort, 0 ohne ok ErrUngueltig oder ein Kontextfehler.
	code model.Code
	ok   bool
}

// ohneSpanFaelle ist die Menge M-OhneSpan (SPEC-290 Nr. 7 a): Schritt (a) 2, (b) 1, (c) 3, (d) 4, (e) 5, (f) 4, `$metadata` 2 gleich 21.
func ohneSpanFaelle() []ohneSpanFall {
	urlFehler := func(e error) func(u *umg, a *inbound.Anfrage) {
		return func(u *umg, _ *inbound.Anfrage) {
			u.c.url.f = func(string, string) (model.ODataURL, error) { return model.ODataURL{}, e }
		}
	}
	methode := func(m string, bau func(u *umg)) func(u *umg, a *inbound.Anfrage) {
		return func(u *umg, a *inbound.Anfrage) { a.Methode = m; bau(u) }
	}
	wsFehler := func(e error) func(u *umg, a *inbound.Anfrage) {
		return func(u *umg, _ *inbound.Anfrage) { u.p.setzeFehler("Workspaces.Get", e) }
	}
	kopf := func(werte ...string) func(u *umg, a *inbound.Anfrage) {
		return func(_ *umg, a *inbound.Anfrage) { a.Header[headerName] = werte }
	}
	nichts := func(*umg) {}
	return []ohneSpanFall{
		{name: "(a) ctx nil", bau: func(*umg, *inbound.Anfrage) {}, keinCtx: true},
		{name: "(a) Methode leer", bau: func(_ *umg, a *inbound.Anfrage) { a.Methode = "" }},
		{name: "(b) kein Dienst", bau: func(_ *umg, a *inbound.Anfrage) { a.Pfad = "/nirgends" }, code: model.CodeE5005},
		{name: "(c) Header fehlt", bau: func(_ *umg, a *inbound.Anfrage) { delete(a.Header, headerName) }, code: model.CodeE5003},
		{name: "(c) Header leer", bau: kopf(""), code: model.CodeE5003},
		{name: "(c) Header zwei Werte", bau: kopf(wsID, wsID), code: model.CodeE5003},
		{name: "(d) URL ODR-E6005", bau: urlFehler(&model.URLFehler{Ursache: model.URLSegment}), code: model.CodeE6005},
		{name: "(d) URL ODR-E6001", bau: urlFehler(&model.URLFehler{Ursache: model.URLServiceDokument}), code: model.CodeE6001},
		{name: "(d) URL ODR-E6006", bau: urlFehler(&model.KompatFehler{Ursache: model.KompatMethode}), code: model.CodeE6006},
		{name: "(d) Art außerhalb", bau: urlFehler(nil), code: model.CodeE9001},
		{name: "(e) OPTIONS", bau: methode("OPTIONS", nichts), code: model.CodeE6002},
		{name: "(e) NurLesend mit POST", bau: methode("POST", func(u *umg) { u.k.Dienste[0].Codecs = u.c.satz(true, "") }), code: model.CodeE6006},
		{name: "(e) POST an Entity", bau: methode("POST", func(u *umg) { u.entity(model.Optionen{}, "1") }), code: model.CodeE6002},
		{name: "(e) $filter an Mutation", bau: methode("PATCH", func(u *umg) { u.entity(model.Optionen{Filter: idGroesser("1")}, "1") }), code: model.CodeE6002},
		{name: "(e) Dienst ohne Leser", bau: methode("POST", func(u *umg) { u.menge(model.Optionen{}); u.k.Eingabe = nil }), code: model.CodeE6002},
		{name: "(f) Workspace unbekannt", bau: func(u *umg, _ *inbound.Anfrage) { u.ws.entferne(wsID) }, code: model.CodeE7001},
		{name: "(f) Workspace fremd", bau: func(u *umg, _ *inbound.Anfrage) { u.ws.setze(model.Workspace{ID: wsID, ServiceID: "anderer"}) }, code: model.CodeE5004},
		{name: "(f) Workspaces.Get mit Fehler", bau: wsFehler(errors.New("roh")), code: model.CodeE9001},
		{name: "(f) Workspaces.Get mit Kontextfehler", bau: wsFehler(context.Canceled)},
		{name: "$metadata GET", bau: func(*umg, *inbound.Anfrage) {}, ok: true},
		{name: "$metadata HEAD", bau: func(_ *umg, a *inbound.Anfrage) { a.Methode = "HEAD" }, ok: true},
	}
}

// TestKeinSpan hält Q4 (SPEC-290 Nr. 2 c, 7 a): `$metadata` und jede Anfrage, die an den Schritten (a) bis (f) endet, rufen Telemetrie nie.
func TestKeinSpan(t *testing.T) {
	faelle := ohneSpanFaelle()
	if len(faelle) != 21 {
		t.Fatalf("%d Fälle, M-OhneSpan nennt 21", len(faelle))
	}
	for _, f := range faelle {
		u := neueUmg(t)
		a := metadatenAnfrage()
		f.bau(u, &a)
		ctx := u.ctx
		if f.keinCtx {
			ctx = keinKontext()
		}
		antwort, err := u.mocker().Mock(ctx, a)
		switch {
		case f.ok:
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
				continue
			}
			lies(t, antwort)
		case f.code != 0:
			_ = erwarteFehler(t, f.name, err, f.code)
		case err == nil || fehlercode(err) != 0:
			t.Errorf("%s: Fehler %v, erwartet einen ohne Code", f.name, err)
		}
		if aufrufe := u.tel.Aufrufe(); len(aufrufe) != 0 {
			t.Errorf("%s: %d Aufrufe von Telemetrie, erwartet 0", f.name, len(aufrufe))
		}
	}
}

// rueckweg ist eine Aufrufstelle S3 bis S11 oder S14 bis S16 hinter dem Beginn der Query (SPEC-290 Nr. 6): die Art, die sie erreicht,
// die Operation, an der der Fehler eingespeist wird, und der Code (S14 bis S16 sind Fehler des Interpreters).
type rueckweg struct {
	stelle string
	op     string
	bau    func(u *umg)
	code   model.Code
}

func rueckwege() []rueckweg {
	menge := func(u *umg) { u.menge(model.Optionen{}) }
	entity := func(u *umg) { u.entity(model.Optionen{}, "1") }
	anzahl := func(u *umg) { u.anzahl(model.Optionen{}) }
	port := model.CodeE9001
	unbekannt := &model.SelectOption{Namen: []string{"Unbekannt"}}
	return []rueckweg{
		{"S3", "Metadaten.Get", menge, port}, {"S4", "Blobs.Open", menge, port}, {"S5", "MetadatenCodec.ParseMetadata", menge, port},
		{"S6", "Coverage.Get", menge, port}, {"S7", "Entities.ListSet", menge, port}, {"S8", "Entities.Get", entity, port},
		{"S9", "Schreiber.SchreibeEntity", entity, port}, {"S10", "Schreiber.SchreibeCollection", menge, port},
		{"S11", "Schreiber.SchreibeAnzahl", anzahl, port},
		{"S14", "", func(u *umg) { u.menge(model.Optionen{OrderBy: []model.Sortierung{{Name: "Ref"}}}) }, model.CodeE6001},
		{"S15", "", func(u *umg) { u.entity(model.Optionen{Select: unbekannt}, "1") }, model.CodeE6005},
		{"S16", "", func(u *umg) {
			u.anzahl(model.Optionen{Filter: filterOpt(vergleich(model.VergleichEq, "Unbekannt", model.Literal{Art: model.LiteralZahl, Text: "1"}))})
		}, model.CodeE6005},
	}
}

// TestSpanEndeJederRueckweg hält Q6 (SPEC-290 Nr. 3 b): genau ein End auf jedem Rückweg der Query, mit Fehler mit Code (12 Stellen) und
// mit Kontextfehler (9 Stellen, S3 bis S11, die Stellen mit ctx); das Ergebnis bleibt der Code bzw. der Kontextfehler ohne Code.
//
// Der Span trägt am Ende die Werte, die an der Stelle feststehen (SPEC-290 Nr. 4 a und b; Orakel je Stelle als Literal): vor der Prüfung
// des Sets (S3 bis S5) `_.` und `nicht_gelesen`, danach der Name des Sets aus dem Modell, und der Status, sobald die Coverage gelesen ist
// (S7, S10, S11, S14, S16 bei Collection und Anzahl), sonst `nicht_gelesen`. Dazu das Paar „Kontextfehler und Fehler der Tabelle zugleich“
// (`errors.Join`, SPEC-109 Nr. 6) an den 9 Stellen mit ctx: Kontextfehler ohne Code, ein End. Fälle: 12 + 9 + 9 = 30.
func TestSpanEndeJederRueckweg(t *testing.T) {
	// endwerte: Name des Sets und Status am Ende, je Stelle.
	endwerte := map[string][2]string{
		"S3": {"_.", "nicht_gelesen"}, "S4": {"_.", "nicht_gelesen"}, "S5": {"_.", "nicht_gelesen"}, "S6": {setProducts, "nicht_gelesen"},
		"S7": {setProducts, "complete"}, "S8": {setProducts, "nicht_gelesen"}, "S9": {setProducts, "nicht_gelesen"},
		"S10": {setProducts, "complete"}, "S11": {setProducts, "complete"}, "S14": {setProducts, "complete"},
		"S15": {setProducts, "nicht_gelesen"}, "S16": {setProducts, "complete"},
	}
	stellen := rueckwege()
	n, mitCtx := 0, 0
	for _, s := range stellen {
		n++
		u := leseUmg(t)
		s.bau(u)
		if s.op != "" {
			u.p.setzeFehler(s.op, errors.New("roh"))
		}
		_, err := u.leseGET()
		_ = erwarteFehler(t, s.stelle+" mit Code", err, s.code)
		erwarteEinEnde(t, s.stelle+" mit Code", u)
		erwarteEndwerte(t, s.stelle+" mit Code", u, endwerte[s.stelle])
		if s.op == "" {
			continue
		}
		mitCtx++
		for _, form := range []struct {
			name string
			err  error
		}{{"Kontextfehler", context.Canceled}, {"Kontextfehler und Fehler der Tabelle", errors.Join(context.Canceled, outbound.ErrBeschaedigt)}} {
			n++
			u = leseUmg(t)
			s.bau(u)
			u.p.setzeFehler(s.op, form.err)
			_, err = u.leseGET()
			if !errors.Is(err, context.Canceled) || fehlercode(err) != 0 {
				t.Errorf("%s mit %s: Fehler %v, erwartet Canceled ohne Code", s.stelle, form.name, err)
			}
			erwarteEinEnde(t, s.stelle+" mit "+form.name, u)
			erwarteEndwerte(t, s.stelle+" mit "+form.name, u, endwerte[s.stelle])
		}
	}
	if len(stellen) != 12 || len(endwerte) != 12 || mitCtx != 9 || n != 30 {
		t.Errorf("%d Stellen, %d Orakel, %d mit ctx, %d Fälle; erwartet 12, 12, 9 und 30 (SPEC-195 Nr. 11)", len(stellen), len(endwerte), mitCtx, n)
	}
}

// erwarteEndwerte prüft Name des Sets und Status am Ende des einen Spans.
func erwarteEndwerte(t *testing.T, kontext string, u *umg, soll [2]string) {
	t.Helper()
	b := buecher(u.tel.FakeTelemetry)
	if len(b) != 1 {
		return
	}
	if set, _ := b[0].wert(attrSet); set != soll[0] {
		t.Errorf("%s: %s %q, erwartet %q", kontext, attrSet, set, soll[0])
	}
	if st, _ := b[0].wert(attrCov); st != soll[1] {
		t.Errorf("%s: %s %q, erwartet %q", kontext, attrCov, st, soll[1])
	}
}

// erwarteEinEnde prüft, dass ein Span begann und genau einmal endete.
func erwarteEinEnde(t *testing.T, kontext string, u *umg) {
	t.Helper()
	if b := buecher(u.tel.FakeTelemetry); len(b) != 1 || b[0].enden != 1 {
		t.Errorf("%s: Bücher %+v, erwartet ein Span mit einem End", kontext, b)
	}
}

// TestSpanCtxUnveraendert hält Q7 (SPEC-290 Nr. 3 d): StartSpan und jeder Port der Query erhalten das ctx des Aufrufs, nie das des
// Spans (der Fake liefert ein abgeleitetes). Elf Stellen S2 bis S12, vier Anfragen (Collection, Entity, Anzahl, $metadata), dazu StartSpan.
func TestSpanCtxUnveraendert(t *testing.T) {
	art := map[string]func(u *umg){
		"menge": func(u *umg) { u.menge(model.Optionen{}) }, "entity": func(u *umg) { u.entity(model.Optionen{}, "1") },
		"anzahl": func(u *umg) { u.anzahl(model.Optionen{}) }, "metadaten": func(*umg) {},
	}
	stellen := []struct{ stelle, op, art string }{
		{"S2", "Workspaces.Get", "menge"}, {"S3", "Metadaten.Get", "menge"}, {"S4", "Blobs.Open", "menge"},
		{"S5", "MetadatenCodec.ParseMetadata", "menge"}, {"S6", "Coverage.Get", "menge"}, {"S7", "Entities.ListSet", "menge"},
		{"S8", "Entities.Get", "entity"}, {"S9", "Schreiber.SchreibeEntity", "entity"}, {"S10", "Schreiber.SchreibeCollection", "menge"},
		{"S11", "Schreiber.SchreibeAnzahl", "anzahl"}, {"S12", "Schreiber.SchreibeAnzahl", "metadaten"},
	}
	if len(stellen) != 11 {
		t.Fatalf("%d Stellen, erwartet 11 (SPEC-195 Nr. 11: S2 bis S12)", len(stellen))
	}
	aufbauten := map[string]*umg{}
	for name, bau := range art {
		u := leseUmg(t)
		bau(u)
		if _, err := u.leseGET(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		aufbauten[name] = u
	}
	for _, s := range stellen {
		u := aufbauten[s.art]
		ev := u.p.von(s.op)
		if len(ev) == 0 {
			t.Errorf("%s: %s nicht gerufen", s.stelle, s.op)
		}
		for _, e := range ev {
			if e.ctx != u.ctx {
				t.Errorf("%s: %s erhielt nicht das ctx des Aufrufs", s.stelle, s.op)
			}
		}
	}
	for name, u := range aufbauten {
		if name == "metadaten" {
			continue
		}
		starts := 0
		for _, a := range u.tel.Aufrufe() {
			if a.Op == supporttest.OpStartSpan {
				starts++
				if any(a.Ctx()) != any(u.ctx) {
					t.Errorf("%s: StartSpan erhielt nicht das ctx des Aufrufs", name)
				}
			}
		}
		if starts != 1 {
			t.Errorf("%s: %d StartSpan, erwartet 1", name, starts)
		}
	}
}

// codeSzenario ist ein Lauf mit einem Code der Menge von SPEC-195 Nr. 10; span ist die Zahl der StartSpan: 1 für die Query und für die
// Mutation hinter dem Beginn (SPEC-290 Nr. 7 b), 0 für jeden Weg, der vor dem Beginn endet.
type codeSzenario struct {
	name string
	code model.Code
	span int
	lauf func(t *testing.T) (*umg, error)
}

// codeSzenarien sind die 16 Codes je ein Szenario und ein zweites für die sechs Codes, die vor und nach dem Beginn entstehen
// (ODR-E5005, ODR-E6001, ODR-E6002, ODR-E6005, ODR-E7001, ODR-E9001): 22 Szenarien.
func codeSzenarien() []codeSzenario {
	ohne := func(c model.Code, name string, bau func(u *umg, a *inbound.Anfrage)) codeSzenario {
		return codeSzenario{name, c, 0, func(t *testing.T) (*umg, error) { return laufeFall(t, fall{name, bau, c}) }}
	}
	mitSpan := func(c model.Code, name string, bau func(u *umg)) codeSzenario {
		return codeSzenario{name, c, 1, func(t *testing.T) (*umg, error) {
			u := leseUmg(t)
			bau(u)
			_, err := u.leseGET()
			return u, err
		}}
	}
	mutation := func(c model.Code, name, methode, id string, bau func(u *umg)) codeSzenario {
		return codeSzenario{name, c, 1, func(t *testing.T) (*umg, error) {
			u := schreibUmg(t)
			bau(u)
			_, err := u.mutiere(methode, id, eingabeIDName(id, "x"))
			return u, err
		}}
	}
	nichts := func(*umg) {}
	urlFehler := func(e error) func(u *umg, a *inbound.Anfrage) {
		return func(u *umg, _ *inbound.Anfrage) {
			u.c.url.f = func(string, string) (model.ODataURL, error) { return model.ODataURL{}, e }
		}
	}
	return []codeSzenario{
		mitSpan(model.CodeE3001, "E3001", func(u *umg) { u.menge(model.Optionen{}); u.p.setzeFehler("Blobs.Open", outbound.ErrBlobFehlt) }),
		ohne(model.CodeE5003, "E5003", func(_ *umg, a *inbound.Anfrage) { delete(a.Header, headerName) }),
		ohne(model.CodeE5004, "E5004", func(u *umg, _ *inbound.Anfrage) { u.ws.setze(model.Workspace{ID: wsID, ServiceID: "anderer"}) }),
		ohne(model.CodeE5005, "E5005 vor dem Beginn", func(_ *umg, a *inbound.Anfrage) { a.Pfad = "/nirgends" }),
		mitSpan(model.CodeE5005, "E5005 nach dem Beginn", func(u *umg) { u.menge(model.Optionen{}); u.md.entferne(dienstID) }),
		ohne(model.CodeE6001, "E6001 vor dem Beginn", urlFehler(&model.URLFehler{Ursache: model.URLServiceDokument})),
		mitSpan(model.CodeE6001, "E6001 nach dem Beginn", func(u *umg) { u.entity(model.Optionen{Filter: idGroesser("1")}, "1") }),
		ohne(model.CodeE6002, "E6002 vor dem Beginn", func(_ *umg, a *inbound.Anfrage) { a.Methode = "OPTIONS" }),
		mutation(model.CodeE6002, "E6002 nach dem Beginn", "POST", "7", func(u *umg) { u.p.setzeFehler("Eingabe.LeseEntity", model.ErrEingabeNichtUnterstuetzt) }),
		mitSpan(model.CodeE6003, "E6003", func(u *umg) {
			u.k.StrictCoverage = true
			u.co.setze(baselineID, setProducts, model.CoveragePartial)
			u.menge(model.Optionen{})
		}),
		mitSpan(model.CodeE6004, "E6004", func(u *umg) {
			u.menge(model.Optionen{})
			u.p.setzeFehler("MetadatenCodec.ParseMetadata", model.ErrMetadatenUngueltig)
		}),
		ohne(model.CodeE6005, "E6005 vor dem Beginn", urlFehler(&model.URLFehler{Ursache: model.URLSegment})),
		mitSpan(model.CodeE6005, "E6005 nach dem Beginn", func(u *umg) {
			u.ressource(model.RessourceEntity, setProducts, model.Optionen{}, zahl(kanarienwert, "1"))
		}),
		ohne(model.CodeE6006, "E6006", urlFehler(&model.KompatFehler{Ursache: model.KompatMethode})),
		mutation(model.CodeE6007, "E6007", "POST", "7", func(u *umg) {
			u.p.setzeFehler("Eingabe.LeseEntity", &model.EingabeFehler{Ursache: model.EingabeSyntax})
		}),
		mitSpan(model.CodeE6008, "E6008", func(u *umg) { u.ressource(model.RessourceMenge, kanarienwert+"Set", model.Optionen{}) }),
		mutation(model.CodeE6009, "E6009", "POST", "2", nichts),
		{"E6010", model.CodeE6010, 1, func(t *testing.T) (*umg, error) {
			u := etagUmg(t, true, 1, 1, 1)
			_, err := u.mutiereMit("DELETE", "2", map[string][]string{"If-Match": {`W/"0"`}})
			return u, err
		}},
		ohne(model.CodeE7001, "E7001 vor dem Beginn", func(u *umg, _ *inbound.Anfrage) { u.ws.entferne(wsID) }),
		mitSpan(model.CodeE7001, "E7001 nach dem Beginn", func(u *umg) {
			u.menge(model.Optionen{})
			u.p.setzeFehler("Entities.ListSet", outbound.ErrNichtGefunden)
		}),
		ohne(model.CodeE9001, "E9001 vor dem Beginn", urlFehler(nil)),
		mitSpan(model.CodeE9001, "E9001 nach dem Beginn", func(u *umg) {
			u.en.setze(setProducts, eintrag(setProducts, "1", `{"Id":1,"`+kanarienwert+`":"x"}`, alleNamen()...))
			u.menge(model.Optionen{})
		}),
	}
}

// TestKeinZaehlerImDienst hält Q2 (SPEC-290 Nr. 1, 5): der Dienst ruft von Telemetrie nie Log und nie CountError, für keinen der 16 Codes
// (22 Szenarien) und keinen Kontextfehler (je Art einer: 6 Lesezugriffe und 4 Mutationen); StartSpan zählt nur, wo der Span beginnt.
func TestKeinZaehlerImDienst(t *testing.T) {
	szenarien := codeSzenarien()
	codes := map[string]bool{}
	for _, s := range szenarien {
		u, err := s.lauf(t)
		_ = erwarteFehler(t, s.name, err, s.code)
		codes[s.code.String()] = true
		if got := u.tel.Anzahl(supporttest.OpLog); got != 0 {
			t.Errorf("%s: Log %d-mal", s.name, got)
		}
		if got := u.tel.Anzahl(supporttest.OpCountError); got != 0 {
			t.Errorf("%s: CountError %d-mal", s.name, got)
		}
		if got := u.tel.Anzahl(supporttest.OpStartSpan); got != s.span {
			t.Errorf("%s: StartSpan %d-mal, erwartet %d", s.name, got, s.span)
		}
		if s.span == 1 {
			erwarteEinEnde(t, s.name, u)
		}
	}
	if len(szenarien) != 22 {
		t.Errorf("%d Szenarien, erwartet 22 (16 Codes und 6 zweite Szenarien)", len(szenarien))
	}
	soll := spec195Nr10()
	if len(soll) != 16 || len(codes) != 16 {
		t.Fatalf("Liste %d Codes, Szenarien %d verschiedene; erwartet 16", len(soll), len(codes))
	}
	for _, c := range soll {
		if !codes[c] {
			t.Errorf("%s hat kein Szenario", c)
		}
	}
}

// TestKeinZaehlerBeiKontextfehler hält Q2 für den Kontextfehler: einer je Art der Operation, ohne Code, ohne Log und ohne CountError.
func TestKeinZaehlerBeiKontextfehler(t *testing.T) {
	n := 0
	for _, methode := range []string{"GET", "HEAD"} {
		for _, art := range []struct {
			name, op string
			bau      func(u *umg)
		}{
			{"Collection", "Coverage.Get", func(u *umg) { u.menge(model.Optionen{}) }},
			{"Entity", "Entities.Get", func(u *umg) { u.entity(model.Optionen{}, "1") }},
			{"Anzahl", "Coverage.Get", func(u *umg) { u.anzahl(model.Optionen{}) }},
		} {
			n++
			u := leseUmg(t)
			art.bau(u)
			u.p.setzeFehler(art.op, context.Canceled)
			_, err := u.rufe(anfrage(methode, dienstRoot+"/x"))
			pruefeKontextfehler(t, methode+" "+art.name, u, err, 1)
		}
	}
	for _, methode := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		n++
		u := schreibUmg(t)
		u.p.setzeFehler("Entities.Get", context.Canceled)
		_, err := u.mutiere(methode, "2", eingabeIDName("2", "x"))
		pruefeKontextfehler(t, methode, u, err, 1)
	}
	if n != 10 {
		t.Errorf("%d Operationen, erwartet 10 (6 Lesezugriffe und 4 Mutationen)", n)
	}
}

func pruefeKontextfehler(t *testing.T, kontext string, u *umg, err error, span int) {
	t.Helper()
	if !errors.Is(err, context.Canceled) || fehlercode(err) != 0 {
		t.Errorf("%s: Fehler %v, erwartet Canceled ohne Code", kontext, err)
	}
	log, zaehler, starts := u.tel.Anzahl(supporttest.OpLog), u.tel.Anzahl(supporttest.OpCountError), u.tel.Anzahl(supporttest.OpStartSpan)
	if log != 0 || zaehler != 0 || starts != span {
		t.Errorf("%s: Log %d, CountError %d, StartSpan %d (erwartet %d)", kontext, log, zaehler, starts, span)
	}
	if span == 1 {
		erwarteEinEnde(t, kontext, u)
	}
}
