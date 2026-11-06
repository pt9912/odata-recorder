package jsoncodec_test

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
	}
	if n != 8 {
		t.Errorf("Menge %d, will 8", n)
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
		ctx  context.Context
		want error
	}{
		{"Canceled", abgebrochen, context.Canceled},
		{"DeadlineExceeded", frist, context.DeadlineExceeded},
		{"Ursache", mitUrsache, context.Canceled},
	}
	n := 0
	for _, op := range operationen() {
		for _, f := range faelle {
			n++
			z := neuesZiel()
			d, err := op.gut(f.ctx, z)
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

// TestModellVersion hält Zeile 5: nur die Version 4.0 gilt, an Entity, Collection und Adresse.
func TestModellVersion(t *testing.T) {
	n := 0
	for _, v := range []string{"2.0", "", "3.0", "4.0 "} {
		m := modellVersion(v)
		if v == "" {
			m = model.ServiceModel{}
		}
		scheitert(t, "Entity "+v, func(c context.Context, d io.Writer) (model.Darstellung, error) {
			return sch.SchreibeEntity(c, d, m, gutEntity())
		})
		scheitert(t, "Collection "+v, func(c context.Context, d io.Writer) (model.Darstellung, error) {
			return sch.SchreibeCollection(c, d, m, gutCollection())
		})
		if a, err := sch.Adresse(m, wurzel, "Produkte", "1"); a != "" {
			t.Errorf("Adresse %q: %q", v, a)
		} else {
			istUngueltig(t, err, "Adresse "+v)
		}
		n += 3
	}
	if n != 12 {
		t.Errorf("Menge %d, will 12", n)
	}
	if a, err := sch.Adresse(modellVersion("4.0"), wurzel, "Produkte", "1"); err != nil || a != wurzel+"/Produkte(1)" {
		t.Errorf("Gegenfall 4.0: %q, %v", a, err)
	}
}

// TestSetUngueltig hält Zeile 6: Set leer, unbekannt, andere Schreibweise, mit Namensraum, ohne auflösbaren Typ, Nullwert-Modell.
func TestSetUngueltig(t *testing.T) {
	formen := []struct {
		name, set string
		m         model.ServiceModel
	}{
		{"leer", "", modell}, {"unbekannt", "Unbekannt", modell}, {"Schreibweise", "produkte", modell},
		{"Namensraum", "Ns.Produkte", modell}, {"ohne Typ", "OhneTyp", modell}, {"Nullwert-Modell", "Produkte", model.ServiceModel{}},
		{"leer, im Modell vorhanden", "", setModell("")},
	}
	n := 0
	for _, f := range formen {
		e, c := gutEntity(), gutCollection()
		e.Set, c.Set = f.set, f.set
		scheitert(t, "Entity "+f.name, func(ctx context.Context, d io.Writer) (model.Darstellung, error) {
			return sch.SchreibeEntity(ctx, d, f.m, e)
		})
		scheitert(t, "Collection "+f.name, func(ctx context.Context, d io.Writer) (model.Darstellung, error) {
			return sch.SchreibeCollection(ctx, d, f.m, c)
		})
		a, err := sch.Adresse(f.m, wurzel, f.set, "1")
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

// TestFehlerobjektTabelle hält Zeile 14: genau NeuesFehlerobjekt(c) zu einem Code der Tabelle.
func TestFehlerobjektTabelle(t *testing.T) {
	codes := model.AlleCodes()
	if len(codes) < 36 {
		t.Fatalf("Tabelle der Codes hat %d Zeilen, am Start 36", len(codes))
	}
	for _, c := range codes {
		f, err := model.NeuesFehlerobjekt(c)
		if err != nil {
			t.Fatal(err)
		}
		z := neuesZiel()
		if _, err := sch.SchreibeFehler(context.Background(), z, f); err != nil {
			t.Errorf("%s: %v", c, err)
		}
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
				return sch.SchreibeFehler(ctx, d, a)
			})
		}
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
			return sch.SchreibeFehler(ctx, d, a)
		})
	}
}

// TestNichtsBeiVerstoss hält Zeile 15: ein Verstoß an irgendeiner Stelle lässt dst leer, auch hinter gültigen Teilen.
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
	letzte.Entities[2].Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "3"), pw("Name", model.ArtString, "C"), pw("Tag", model.ArtDate, "2023-02-29")}
	collectionScheitert(t, "letzte Property des letzten Elements", letzte)
	paar := gutEntity()
	paar.Set = "Unbekannt"
	paar.Entity.Typ = "Ns.Abstrakt"
	entityScheitert(t, "Set und Typ", paar)
	if n+2 != 8 {
		t.Errorf("Menge %d, will 8", n+2)
	}
}

// TestEinWrite hält Zeile 16: genau ein Write, sonst keine Methode; die Ausgabe ist nie leer.
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
			return sch.SchreibeEntity(context.Background(), z, modell, gutEntity())
		},
		"Entity groß": func(z *ziel) (model.Darstellung, error) {
			return sch.SchreibeEntity(context.Background(), z, modell, grossEntity)
		},
		"Collection klein": func(z *ziel) (model.Darstellung, error) {
			return sch.SchreibeCollection(context.Background(), z, modell, gutCollection())
		},
		"Collection groß": func(z *ziel) (model.Darstellung, error) {
			return sch.SchreibeCollection(context.Background(), z, modell, gross)
		},
		"Anzahl klein": func(z *ziel) (model.Darstellung, error) { return sch.SchreibeAnzahl(context.Background(), z, 5) },
		"Anzahl groß":  func(z *ziel) (model.Darstellung, error) { return sch.SchreibeAnzahl(context.Background(), z, 1<<62) },
		"Fehler klein": func(z *ziel) (model.Darstellung, error) {
			return sch.SchreibeFehler(context.Background(), z, fehlerCode())
		},
		"Fehler letzter Code": func(z *ziel) (model.Darstellung, error) {
			f, _ := model.NeuesFehlerobjekt(model.AlleCodes()[len(model.AlleCodes())-1])
			return sch.SchreibeFehler(context.Background(), z, f)
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

// TestDstFehler hält Zeile 17: der Fehler unverändert, ein kurzer Write ist io.ErrShortWrite, Darstellung der Nullwert.
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

// TestDarstellungskopf hält Zeile 18: Content-Type und Kennzeichen jeder Ausgabe, als Literal.
func TestDarstellungskopf(t *testing.T) {
	kennzeichen := model.Kopfzeile{Name: "OData-Version", Wert: "4.0"}
	faelle := []struct {
		accept []string
		want   string
	}{
		{nil, "application/json;odata.metadata=minimal;charset=utf-8"},
		{[]string{"application/json;odata.metadata=none"}, "application/json;odata.metadata=none;charset=utf-8"},
		{[]string{"application/json;IEEE754Compatible=true"}, "application/json;odata.metadata=minimal;IEEE754Compatible=true;charset=utf-8"},
		{[]string{"application/json;odata.metadata=none;IEEE754Compatible=true"}, "application/json;odata.metadata=none;IEEE754Compatible=true;charset=utf-8"},
	}
	n := 0
	for _, f := range faelle {
		e, c := gutEntity(), gutCollection()
		e.Accept, c.Accept = f.accept, f.accept
		d, err := sch.SchreibeEntity(context.Background(), neuesZiel(), modell, e)
		if err != nil || d.ContentType != f.want || d.Kennzeichen != kennzeichen {
			t.Errorf("Entity %v: %+v, %v", f.accept, d, err)
		}
		d, err = sch.SchreibeCollection(context.Background(), neuesZiel(), modell, c)
		if err != nil || d.ContentType != f.want || d.Kennzeichen != kennzeichen {
			t.Errorf("Collection %v: %+v, %v", f.accept, d, err)
		}
		n += 2
	}
	d, err := sch.SchreibeAnzahl(context.Background(), neuesZiel(), 1)
	if err != nil || d != (model.Darstellung{ContentType: "text/plain;charset=utf-8", Kennzeichen: kennzeichen}) {
		t.Errorf("Anzahl: %+v, %v", d, err)
	}
	d, err = sch.SchreibeFehler(context.Background(), neuesZiel(), fehlerCode())
	if err != nil || d != (model.Darstellung{ContentType: "application/json;odata.metadata=minimal;charset=utf-8", Kennzeichen: kennzeichen}) {
		t.Errorf("Fehler: %+v, %v", d, err)
	}
	if n+2 != 10 {
		t.Errorf("Menge %d, will 10", n+2)
	}
}
