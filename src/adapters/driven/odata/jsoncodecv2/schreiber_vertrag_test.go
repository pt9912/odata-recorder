package jsoncodecv2_test

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestSchreibeNilArgumente hält Zeile 1: Operation × Argument (ctx, dst nil).
func TestSchreibeNilArgumente(t *testing.T) {
	var nilCtx context.Context
	var nilDst io.Writer
	n := 0
	for _, op := range operationen() {
		for _, arg := range []string{"ctx", "dst"} {
			n++
			scheitert(t, op.name+" "+arg+" nil", func(c context.Context, d io.Writer) (model.Darstellung, error) {
				if arg == "ctx" {
					return op.gut(nilCtx, d)
				}
				return op.gut(c, nilDst)
			})
		}
	}
	if n != 8 {
		t.Errorf("Menge %d, will 8", n)
	}
}

// TestArgumentVorKontext hält Zeile 2: ein ungültiges Argument ist ErrUngueltig, auch bei beendetem Kontext.
func TestArgumentVorKontext(t *testing.T) {
	abgebrochen, cancel := context.WithCancel(context.Background())
	cancel()
	frist, cancel2 := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer cancel2()
	n := 0
	for _, op := range operationen() {
		for zustand, ctx := range map[string]context.Context{"abgebrochen": abgebrochen, "Frist": frist} {
			n++
			z := neuesZiel()
			d, err := op.schlecht(ctx, z)
			istUngueltig(t, err, op.name+" "+zustand)
			if d != (model.Darstellung{}) {
				t.Errorf("%s %s: Darstellung %+v", op.name, zustand, d)
			}
			z.leer(t, op.name+" "+zustand)
		}
		// Paar aus nil-Ziel und beendetem Kontext: (a) vor (c), SPEC-168 Nr. 1
		var nilDst io.Writer
		for zustand, ctx := range map[string]context.Context{"abgebrochen": abgebrochen, "Frist": frist} {
			n++
			d, err := op.gut(ctx, nilDst)
			istUngueltig(t, err, op.name+" dst nil und "+zustand)
			if d != (model.Darstellung{}) {
				t.Errorf("%s dst nil und %s: Darstellung %+v", op.name, zustand, d)
			}
		}
	}
	if n != 16 {
		t.Errorf("Menge %d, will 16 (4 Operationen × 2 Zustände × {ungültiges Argument, dst nil})", n)
	}
}

// TestKontextBeendet hält Zeile 3: der Fehler des Kontexts unverändert, dst leer; der aktive Kontext schreibt.
func TestKontextBeendet(t *testing.T) {
	abgebrochen, cancel := context.WithCancel(context.Background())
	cancel()
	frist, cancel2 := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer cancel2()
	ursache := errors.New("ursache")
	mitUrsache, cancel3 := context.WithCancelCause(context.Background())
	cancel3(ursache)
	faelle := []struct {
		name string
		ctx  func() context.Context
		want error
	}{
		{"Canceled", func() context.Context { return abgebrochen }, context.Canceled},
		{"DeadlineExceeded", func() context.Context { return frist }, context.DeadlineExceeded},
		{"Ursache", func() context.Context { return mitUrsache }, context.Canceled},
	}
	n := 0
	for _, op := range operationen() {
		for _, f := range faelle {
			n++
			z := neuesZiel()
			d, err := op.gut(f.ctx(), z)
			if err != f.want || err == ursache {
				t.Errorf("%s %s: Fehler %v (%T), will %v unverändert", op.name, f.name, err, err, f.want)
			}
			var coded interface{ Code() model.Code }
			if errors.As(err, &coded) {
				t.Errorf("%s %s: trägt einen Code", op.name, f.name)
			}
			if d != (model.Darstellung{}) {
				t.Errorf("%s %s: Darstellung %+v", op.name, f.name, d)
			}
			z.leer(t, op.name+" "+f.name)
		}
		z := neuesZiel()
		if _, err := op.gut(context.Background(), z); err != nil {
			t.Errorf("%s aktiv: %v", op.name, err)
		}
		z.nurEinWrite(t, op.name+" aktiv")
	}
	if n != 12 {
		t.Errorf("Menge %d, will 12", n)
	}
}

// TestKontextEinmalVorWrite hält Zeile 4: Err genau einmal, nach Argumenten und Aufbau, unmittelbar vor dem Write.
func TestKontextEinmalVorWrite(t *testing.T) {
	n := 0
	for _, op := range operationen() {
		z := neuesZiel()
		k := neuerKontext(z, nil)
		if _, err := op.gut(k, z); err != nil {
			t.Fatalf("%s: %v", op.name, err)
		}
		if got := strings.Join(*z.log, ","); got != "Err,Write" || k.doneN != 0 || k.deadlN != 0 {
			t.Errorf("%s gültig: Ereignisse %q, Done %d, Deadline %d, will Err,Write", op.name, got, k.doneN, k.deadlN)
		}
		z2 := neuesZiel()
		k2 := neuerKontext(z2, nil)
		if _, err := op.schlecht(k2, z2); err == nil {
			t.Errorf("%s ungültig: kein Fehler", op.name)
		}
		if len(*z2.log) != 0 || k2.errN != 0 || k2.doneN != 0 || k2.deadlN != 0 {
			t.Errorf("%s ungültig: Ereignisse %v", op.name, *z2.log)
		}
		n += 2
	}
	if n != 8 {
		t.Errorf("Menge %d, will 8", n)
	}
}

// TestModellVersion hält Zeile 5: nur die Version 2.0 gilt, an Entity, Collection und Adresse.
func TestModellVersion(t *testing.T) {
	n := 0
	for _, v := range []string{"4.0", "", "3.0", "2.0 "} {
		m := modellVersion(v)
		if v == "" {
			m = model.ServiceModel{}
		}
		scheitert(t, "Entity "+v, func(c context.Context, d io.Writer) (model.Darstellung, error) {
			return schreiber().SchreibeEntity(c, d, m, gutEntity())
		})
		scheitert(t, "Collection "+v, func(c context.Context, d io.Writer) (model.Darstellung, error) {
			return schreiber().SchreibeCollection(c, d, m, gutCollection())
		})
		if a, err := schreiber().Adresse(m, wurzel, "Produkte", "1"); a != "" {
			t.Errorf("Adresse %q: %q", v, a)
		} else {
			istUngueltig(t, err, "Adresse "+v)
		}
		n += 3
	}
	if n != 12 {
		t.Errorf("Menge %d, will 12", n)
	}
	if a, err := schreiber().Adresse(modellVersion("2.0"), wurzel, "Produkte", "1"); err != nil || a != wurzel+"/Produkte(1)" {
		t.Errorf("Gegenfall 2.0: %q, %v", a, err)
	}
}

// TestSetUngueltig hält Zeile 6: Set leer, unbekannt, andere Schreibweise, mit Namensraum, ohne auflösbaren Typ, Nullwert-Modell.
func TestSetUngueltig(t *testing.T) {
	formen := []struct {
		name, set string
		m         model.ServiceModel
	}{
		{"leer", "", testModell()}, {"unbekannt", "Unbekannt", testModell()}, {"Schreibweise", "produkte", testModell()},
		{"Namensraum", "Ns.Produkte", testModell()}, {"ohne Typ", "OhneTyp", testModell()}, {"Nullwert-Modell", "Produkte", model.ServiceModel{}},
		{"leer, im Modell vorhanden", "", setModell("")},
	}
	n := 0
	for _, f := range formen {
		e, c := gutEntity(), gutCollection()
		e.Set, c.Set = f.set, f.set
		scheitert(t, "Entity "+f.name, func(ctx context.Context, d io.Writer) (model.Darstellung, error) {
			return schreiber().SchreibeEntity(ctx, d, f.m, e)
		})
		scheitert(t, "Collection "+f.name, func(ctx context.Context, d io.Writer) (model.Darstellung, error) {
			return schreiber().SchreibeCollection(ctx, d, f.m, c)
		})
		a, err := schreiber().Adresse(f.m, wurzel, f.set, "1")
		if a != "" {
			t.Errorf("Adresse %s: %q", f.name, a)
		}
		istUngueltig(t, err, "Adresse "+f.name)
		n += 3
	}
	if n != 21 {
		t.Errorf("Menge %d, will 21 (7 Formen, 3 Operationen)", n)
	}
}

// TestFehlerobjektTabelle hält Zeile 18: genau NeuesFehlerobjekt(c) zu einem Code der Tabelle.
func TestFehlerobjektTabelle(t *testing.T) {
	codes := model.AlleCodes()
	if len(codes) < 36 {
		t.Fatalf("Tabelle der Codes hat %d Zeilen, am Start 36", len(codes))
	}
	gueltig, abgelehnt := 0, 0
	for _, c := range codes {
		f, err := model.NeuesFehlerobjekt(c)
		if err != nil {
			t.Fatal(err)
		}
		z := neuesZiel()
		if _, err := schreiber().SchreibeFehler(context.Background(), z, f); err != nil {
			t.Errorf("%s: %v", c, err)
		}
		gueltig++
		abweichungen := map[string]model.Fehlerobjekt{}
		g := f
		g.Error.Message += "x"
		abweichungen["message angehängt"] = g
		g = f
		g.Error.Message = ""
		abweichungen["message leer"] = g
		g = f
		g.Error.Target += "x"
		abweichungen["target angehängt"] = g
		g = f
		g.Error.Target = ""
		abweichungen["target leer"] = g
		g = f
		g.Error.Code = strings.ToLower(f.Error.Code)
		abweichungen["code klein"] = g
		g = f
		g.Error.Code += " "
		abweichungen["code mit Leerzeichen"] = g
		for was, a := range abweichungen {
			if a == f {
				continue
			}
			scheitert(t, c.String()+" "+was, func(ctx context.Context, d io.Writer) (model.Darstellung, error) {
				return schreiber().SchreibeFehler(ctx, d, a)
			})
			abgelehnt++
		}
	}
	if gueltig != len(codes) || abgelehnt < 4*len(codes) {
		t.Errorf("gültig %d von %d, abgelehnt %d", gueltig, len(codes), abgelehnt)
	}
	f := fehlerCode()
	faelle := map[string]model.Fehlerobjekt{
		"leer":                  {},
		"unbekannter Code":      {Error: model.FehlerInhalt{Code: "ODR-E0000", Message: f.Error.Message, Target: f.Error.Target}},
		"message der Anfrage":   {Error: model.FehlerInhalt{Code: f.Error.Code, Message: "GET /Produkte?$filter=x", Target: f.Error.Target}},
		"nur Code":              {Error: model.FehlerInhalt{Code: f.Error.Code}},
		"Code in Kleinbuchstab": {Error: model.FehlerInhalt{Code: strings.ToLower(f.Error.Code), Message: f.Error.Message, Target: f.Error.Target}},
	}
	for was, a := range faelle {
		scheitert(t, was, func(ctx context.Context, d io.Writer) (model.Darstellung, error) {
			return schreiber().SchreibeFehler(ctx, d, a)
		})
	}
}

// TestNichtsBeiVerstoss hält Zeile 19: ein Verstoß an irgendeiner Stelle lässt dst leer, auch hinter gültigen Teilen.
func TestNichtsBeiVerstoss(t *testing.T) {
	schlecht := map[string]func(e *model.Entity){
		"Typ":  func(e *model.Entity) { e.Typ = "Ns.Unbekannt" },
		"Wert": func(e *model.Entity) { e.Properties[1].Wert.Text = "\x80" },
	}
	n := 0
	for stelle := range 3 {
		for was, f := range schlecht {
			a := gutCollection()
			a.Entities = append([]model.Entity{}, a.Entities...)
			a.Entities[stelle].Properties = append([]model.PropertyWert{}, a.Entities[stelle].Properties...)
			f(&a.Entities[stelle])
			collectionScheitert(t, "Element "+string(rune('1'+stelle))+" "+was, a)
			n++
		}
	}
	letzte := gutCollection()
	letzte.Entities[2].Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "3"), pw("Name", model.ArtString, "C"), pw("Zeit", model.ArtDateTimeOffset, "2023-02-29T00:00:00Z")}
	collectionScheitert(t, "letzte Property des letzten Elements", letzte)
	paar := gutEntity()
	paar.Set = "Unbekannt"
	paar.Entity.Typ = "Ns.Abstrakt"
	entityScheitert(t, "Set und Typ", paar)
	wurzelWert := gutEntity()
	wurzelWert.Wurzel = "ftp://h"
	wurzelWert.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "01")}
	entityScheitert(t, "Wurzel und Wert", wurzelWert)
	scheitert(t, "Version und Set", func(c context.Context, d io.Writer) (model.Darstellung, error) {
		a := gutEntity()
		a.Set = "Unbekannt"
		return schreiber().SchreibeEntity(c, d, modellVersion("4.0"), a)
	})
	if n+4 != 10 {
		t.Errorf("Menge %d, will 10", n+4)
	}
}

// TestEinWrite hält Zeile 20: genau ein Write, sonst keine Methode; die Ausgabe ist nie leer.
func TestEinWrite(t *testing.T) {
	gross := gutCollection()
	gross.Entities = nil
	for i := range 1000 {
		gross.Entities = append(gross.Entities, produkt(strconv.Itoa(i+1), strings.Repeat("x", 100)))
	}
	grossEntity := gutEntity()
	grossEntity.Entity.Properties[1].Wert.Text = strings.Repeat("y", 200<<10)
	n := 0
	groessen := map[string]func(z *ziel) (model.Darstellung, error){
		"Entity klein": func(z *ziel) (model.Darstellung, error) {
			return schreiber().SchreibeEntity(context.Background(), z, testModell(), gutEntity())
		},
		"Entity groß": func(z *ziel) (model.Darstellung, error) {
			return schreiber().SchreibeEntity(context.Background(), z, testModell(), grossEntity)
		},
		"Collection klein": func(z *ziel) (model.Darstellung, error) {
			return schreiber().SchreibeCollection(context.Background(), z, testModell(), gutCollection())
		},
		"Collection groß": func(z *ziel) (model.Darstellung, error) {
			return schreiber().SchreibeCollection(context.Background(), z, testModell(), gross)
		},
		"Anzahl klein": func(z *ziel) (model.Darstellung, error) {
			return schreiber().SchreibeAnzahl(context.Background(), z, 5)
		},
		"Anzahl groß": func(z *ziel) (model.Darstellung, error) {
			return schreiber().SchreibeAnzahl(context.Background(), z, 1<<62)
		},
		"Fehler klein": func(z *ziel) (model.Darstellung, error) {
			return schreiber().SchreibeFehler(context.Background(), z, fehlerCode())
		},
		"Fehler letzter Code": func(z *ziel) (model.Darstellung, error) {
			f, _ := model.NeuesFehlerobjekt(model.AlleCodes()[len(model.AlleCodes())-1])
			return schreiber().SchreibeFehler(context.Background(), z, f)
		},
	}
	for was, f := range groessen {
		z := neuesZiel()
		if _, err := f(z); err != nil {
			t.Fatalf("%s: %v", was, err)
		}
		z.nurEinWrite(t, was)
		if len(z.writes[0]) == 0 {
			t.Errorf("%s: leere Ausgabe", was)
		}
		if strings.HasSuffix(was, "groß") && strings.HasPrefix(was, "Collection") && len(z.writes[0]) < 100<<10 {
			t.Errorf("%s: nur %d Bytes", was, len(z.writes[0]))
		}
		n++
	}
	if n != 8 {
		t.Errorf("Menge %d, will 8", n)
	}
}

// TestDstFehler hält Zeile 21: der Fehler unverändert, ein kurzer Write ist io.ErrShortWrite, Darstellung der Nullwert.
func TestDstFehler(t *testing.T) {
	sentinel := errors.New("dst kaputt")
	n := 0
	for _, op := range operationen() {
		faelle := []struct {
			name string
			z    *ziel
			want error
		}{
			{"Fehler", &ziel{log: &[]string{}, kurz: -1, err: sentinel}, sentinel},
			{"kurz 0", &ziel{log: &[]string{}, kurz: 0}, io.ErrShortWrite},
			{"kurz Länge-1", &ziel{log: &[]string{}, kurz: -2}, io.ErrShortWrite},
		}
		for _, f := range faelle {
			d, err := op.gut(context.Background(), f.z)
			if err != f.want {
				t.Errorf("%s %s: Fehler %v, will %v unverändert", op.name, f.name, err, f.want)
			}
			if d != (model.Darstellung{}) {
				t.Errorf("%s %s: Darstellung %+v", op.name, f.name, d)
			}
			if len(f.z.writes) != 1 {
				t.Errorf("%s %s: Write %d-mal", op.name, f.name, len(f.z.writes))
			}
			n++
		}
		z := &ziel{log: &[]string{}, kurz: -1, err: sentinel}
		_, err := op.schlecht(context.Background(), z)
		istUngueltig(t, err, op.name+" ungültig und dst-Fehler")
		z.leer(t, op.name+" ungültig und dst-Fehler")
		abgebrochen, cancel := context.WithCancel(context.Background())
		cancel()
		z = &ziel{log: &[]string{}, kurz: -1, err: sentinel}
		if _, err := op.gut(abgebrochen, z); err != context.Canceled {
			t.Errorf("%s beendet und dst-Fehler: %v", op.name, err)
		}
		z.leer(t, op.name+" beendet und dst-Fehler")
	}
	if n != 12 {
		t.Errorf("Menge %d, will 12", n)
	}
}

// TestDarstellungskopf hält Zeile 22: Content-Type und Kennzeichen jeder Ausgabe als Literal; Accept hat keine Wirkung und keinen Fehler.
func TestDarstellungskopf(t *testing.T) {
	kennzeichen := model.Kopfzeile{Name: "DataServiceVersion", Wert: "2.0"}
	json := "application/json;charset=utf-8"
	n := 0
	d, err := schreiber().SchreibeEntity(context.Background(), neuesZiel(), testModell(), gutEntity())
	if err != nil || d != (model.Darstellung{ContentType: json, Kennzeichen: kennzeichen}) {
		t.Errorf("Entity: %+v, %v", d, err)
	}
	d, err = schreiber().SchreibeCollection(context.Background(), neuesZiel(), testModell(), gutCollection())
	if err != nil || d != (model.Darstellung{ContentType: json, Kennzeichen: kennzeichen}) {
		t.Errorf("Collection: %+v, %v", d, err)
	}
	d, err = schreiber().SchreibeAnzahl(context.Background(), neuesZiel(), 1)
	if err != nil || d != (model.Darstellung{ContentType: "text/plain;charset=utf-8", Kennzeichen: kennzeichen}) {
		t.Errorf("Anzahl: %+v, %v", d, err)
	}
	d, err = schreiber().SchreibeFehler(context.Background(), neuesZiel(), fehlerCode())
	if err != nil || d != (model.Darstellung{ContentType: json, Kennzeichen: kennzeichen}) {
		t.Errorf("Fehler: %+v, %v", d, err)
	}
	n += 4
	formen := [][]string{
		nil, {""}, {"application/json"}, {"application/json;odata.metadata=none"}, {"application/atom+xml"}, {"*/*"},
		{"odata.metadata=none"}, {"application/atom+xml", "application/json;odata.metadata=none"}, {"application/json;;odata.metadata=none"},
	}
	ohneE, ohneC := entityBody(t, "ohne Accept", gutEntity()), collectionBody(t, "ohne Accept", gutCollection())
	for _, accept := range formen {
		e, c := gutEntity(), gutCollection()
		e.Accept, c.Accept = accept, accept
		ze, zc := neuesZiel(), neuesZiel()
		de, err1 := schreiber().SchreibeEntity(context.Background(), ze, testModell(), e)
		dc, err2 := schreiber().SchreibeCollection(context.Background(), zc, testModell(), c)
		if err1 != nil || err2 != nil || de != (model.Darstellung{ContentType: json, Kennzeichen: kennzeichen}) || dc != de {
			t.Errorf("Accept %q: %+v %+v, %v, %v", accept, de, dc, err1, err2)
		}
		if ze.body() != ohneE || zc.body() != ohneC {
			t.Errorf("Accept %q ändert die Bytes: %s | %s", accept, ze.body(), zc.body())
		}
		n += 2
	}
	if len(formen) != 9 || n != 22 {
		t.Errorf("Formen %d, Menge %d; will 9, 22", len(formen), n)
	}
}
