package jsoncodec_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ outbound.ODataEingabe = (*jsoncodec.EingabeLeser)(nil)

// eGrenze ist die Grenze der meisten Fälle (die Standardgrenze von MAX_CAPTURE_BODY).
const eGrenze = int64(8 << 20)

func eArten() []model.EingabeArt {
	return []model.EingabeArt{model.EingabePost, model.EingabePatch, model.EingabePut}
}

// ---- Fakes -----------------------------------------------------------------------------------------------------------

// eSchritt ist ein Read von src: die Bytes und der Fehler, der mit dem letzten Byte kommt.
type eSchritt struct {
	daten []byte
	err   error
}

// eStrom ist der Fake von src: er hält jeden Read fest (angeforderte und gelieferte Bytes) und zählt jede Operation, die der
// Leser nie rufen darf. Das Ereignisprotokoll teilt er mit dem Fake von ctx.
type eStrom struct {
	log                                         *[]string
	schritte                                    []eSchritt
	i, off                                      int
	readN, nachEnde                             int
	angefordert, geliefert                      []int
	summe                                       int64
	closeN, writeToN, readByteN, seekN, readAtN int
	beendet                                     bool
}

func eNeuerStrom(schritte ...eSchritt) *eStrom {
	return &eStrom{log: &[]string{}, schritte: schritte}
}

// eStromGeteilt liefert body in den Stücken der Längen laengen (der Rest als letztes Stück) und endet danach mit io.EOF.
func eStromGeteilt(body []byte, laengen ...int) *eStrom {
	var s []eSchritt
	for _, n := range laengen {
		if n > len(body) {
			n = len(body)
		}
		s = append(s, eSchritt{daten: body[:n]})
		body = body[n:]
	}
	if len(body) > 0 {
		s = append(s, eSchritt{daten: body})
	}
	return eNeuerStrom(s...)
}

func eStromText(body string) *eStrom { return eStromGeteilt([]byte(body)) }

func (s *eStrom) Read(p []byte) (int, error) {
	s.readN++
	*s.log = append(*s.log, "Read")
	s.angefordert = append(s.angefordert, len(p))
	s.summe += int64(len(p))
	if s.beendet {
		s.nachEnde++
	}
	if s.i >= len(s.schritte) {
		s.beendet = true
		s.geliefert = append(s.geliefert, 0)
		return 0, io.EOF
	}
	st := s.schritte[s.i]
	n := copy(p, st.daten[s.off:])
	s.off += n
	var err error
	if s.off >= len(st.daten) {
		err = st.err
		s.i++
		s.off = 0
		if err != nil {
			s.beendet = true
		}
	}
	s.geliefert = append(s.geliefert, n)
	return n, err
}

func (s *eStrom) Close() error                      { s.closeN++; return nil }
func (s *eStrom) WriteTo(io.Writer) (int64, error)  { s.writeToN++; return 0, nil }
func (s *eStrom) ReadByte() (byte, error)           { s.readByteN++; return 0, io.EOF }
func (s *eStrom) Seek(int64, int) (int64, error)    { s.seekN++; return 0, nil }
func (s *eStrom) ReadAt([]byte, int64) (int, error) { s.readAtN++; return 0, io.EOF }

// verboten meldet jede Operation, die der Leser nie rufen darf, und jeden Read nach dem Ende.
func (s *eStrom) verboten(t *testing.T, was string) {
	t.Helper()
	if s.closeN+s.writeToN+s.readByteN+s.seekN+s.readAtN+s.nachEnde != 0 {
		t.Errorf("%s: Close %d, WriteTo %d, ReadByte %d, Seek %d, ReadAt %d, Read nach dem Ende %d", was, s.closeN, s.writeToN, s.readByteN, s.seekN, s.readAtN, s.nachEnde)
	}
}

// eKontext ist der Fake von ctx: er zählt Err, Done und Deadline; ab ist die Nummer des Aufrufs von Err, ab dem er err
// liefert (0: nie).
type eKontext struct {
	log                 *[]string
	err                 error
	ab                  int
	errN, doneN, deadlN int
}

func (*eKontext) Value(any) any { return nil }

func (k *eKontext) Err() error {
	k.errN++
	*k.log = append(*k.log, "Err")
	if k.ab > 0 && k.errN >= k.ab {
		return k.err
	}
	return nil
}

func (k *eKontext) Done() <-chan struct{} {
	k.doneN++
	*k.log = append(*k.log, "Done")
	return nil
}

func (k *eKontext) Deadline() (time.Time, bool) {
	k.deadlN++
	*k.log = append(*k.log, "Deadline")
	return time.Time{}, false
}

// ---- Aufruf ----------------------------------------------------------------------------------------------------------

// eAufruf sind die Argumente von LeseEntity.
type eAufruf struct {
	ctx    context.Context
	src    io.Reader
	ct     []string
	m      model.ServiceModel
	set    string
	art    model.EingabeArt
	grenze int64
}

var eM = eModell()

func eJSONCT() []string { return []string{"application/json"} }

// eA ist der gültige Aufruf mit dem Body.
func eA(set string, art model.EingabeArt, body string) eAufruf {
	return eAufruf{ctx: context.Background(), src: strings.NewReader(body), ct: eJSONCT(), m: eM, set: set, art: art, grenze: eGrenze}
}

func (a eAufruf) los() (model.EntityEingabe, error) {
	return jsoncodec.NewEingabe().LeseEntity(a.ctx, a.src, a.ct, a.m, a.set, a.art, a.grenze)
}

// ---- Prüfungen -------------------------------------------------------------------------------------------------------

// eFehler prüft, dass der Fehler ein *EingabeFehler mit Ursache und Ort ist, den Sentinel genau seiner Klasse und den Code
// seiner Klasse trägt, und dass das Ergebnis der Nullwert ist.
func eFehler(t *testing.T, was string, got model.EntityEingabe, err error, u model.EingabeUrsache, ort string) {
	t.Helper()
	ef, ok := err.(*model.EingabeFehler)
	if !ok {
		t.Errorf("%s: Fehler %T %v, erwartet *EingabeFehler %s %q", was, err, err, u, ort)
		return
	}
	if ef.Ursache != u || ef.Ort != ort {
		t.Errorf("%s: Ursache %q Ort %q, erwartet %q %q", was, ef.Ursache, ef.Ort, u, ort)
	}
	nicht := u.NichtUnterstuetzt()
	if errors.Is(err, model.ErrEingabeNichtUnterstuetzt) != nicht || errors.Is(err, model.ErrEingabeUngueltig) == nicht {
		t.Errorf("%s: Sentinel der Klasse falsch (nicht_unterstuetzt %v)", was, nicht)
	}
	if (nicht && ef.Code() != model.CodeE6002) || (!nicht && ef.Code() != model.CodeE6007) {
		t.Errorf("%s: Code %v", was, ef.Code())
	}
	if got.Typ != "" || got.Schluessel != "" || got.Properties != nil {
		t.Errorf("%s: Ergebnis %+v, erwartet der Nullwert", was, got)
	}
}

// eOK prüft ein Ergebnis ohne Fehler.
func eOK(t *testing.T, was string, got model.EntityEingabe, err error, soll model.EntityEingabe) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: Fehler %v", was, err)
		return
	}
	if soll.Properties == nil {
		soll.Properties = []model.PropertyWert{}
	}
	if got.Properties == nil || !reflect.DeepEqual(got, soll) {
		t.Errorf("%s: Ergebnis %+v, erwartet %+v", was, got, soll)
	}
}

// eProg prüft einen Programmfehler: model.ErrUngueltig ohne EingabeFehler und ohne Code, Nullwert.
func eProg(t *testing.T, was string, got model.EntityEingabe, err error) {
	t.Helper()
	var ef *model.EingabeFehler
	var coded interface{ Code() model.Code }
	switch {
	case !errors.Is(err, model.ErrUngueltig):
		t.Errorf("%s: kein ErrUngueltig: %v", was, err)
	case errors.As(err, &ef), errors.Is(err, model.ErrEingabeUngueltig), errors.Is(err, model.ErrEingabeNichtUnterstuetzt):
		t.Errorf("%s: ist ein EingabeFehler: %v", was, err)
	case errors.As(err, &coded):
		t.Errorf("%s: trägt einen Code: %v", was, err)
	}
	if got.Typ != "" || got.Schluessel != "" || got.Properties != nil {
		t.Errorf("%s: Ergebnis %+v", was, got)
	}
}

// ---- Bodys -----------------------------------------------------------------------------------------------------------

// eQuote schreibt s als JSON-String: Anführungszeichen, Backslash und Steuerzeichen als Escape, alles andere roh. Die
// Escape-Folge entsteht aus Bytewerten (AGENTS.md §3.18).
func eQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := range len(s) {
		switch c := s[i]; {
		case c == '"' || c == '\\':
			b.WriteString(bs + string(c))
		case c < 0x20:
			b.WriteString(esc(fmt.Sprintf("%04X", c)))
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// eObj baut ein Objekt aus Paaren Name, Wert; der Name wird als JSON-String geschrieben, der Wert ist rohes JSON.
func eObj(paare ...string) string {
	var b strings.Builder
	b.WriteByte('{')
	for i := 0; i < len(paare); i += 2 {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(eQuote(paare[i]) + ":" + paare[i+1])
	}
	b.WriteByte('}')
	return b.String()
}

// ePW ist ein Eintrag der Properties.
func ePW(name string, art model.WertArt, text string) model.PropertyWert { return pw(name, art, text) }

// ---- Modell ----------------------------------------------------------------------------------------------------------

func eP(i int) string { return strconv.Itoa(i) }

// eModell ist das Modell der Tests: Z (vier Zustände), W (Werte), M und MO (Mitglieder), die Hierarchie, R (Pflicht), St und
// StO (Stellen), K0 bis K7, KZ2, KZ3 (Schlüssel) und OhneKey.
func eModell() model.ServiceModel {
	var typen []model.EntityType
	var sets []model.EntitySet
	add := func(set string, t model.EntityType) {
		typen = append(typen, t)
		if set != "" {
			sets = append(sets, model.EntitySet{Name: set, EntityType: t.Name})
		}
	}
	edm := model.UnterstuetzteEDMTypen()
	zp := []model.Property{prop("K", "Edm.Int32", false)}
	for i, ty := range edm {
		zp = append(zp, prop("P"+eP(i), ty, false))
	}
	for i, ty := range edm {
		zp = append(zp, prop("Q"+eP(i), ty, true))
	}
	add("Zs", model.EntityType{Name: "Ns.Z", Keys: []string{"K"}, Properties: zp})

	wp := []model.Property{prop("Id", "Edm.Int32", false)}
	for i, n := range eWNamen() {
		wp = append(wp, prop(n, edm[i], true))
	}
	add("Ws", model.EntityType{Name: "Ns.W", Keys: []string{"Id"}, Properties: wp})

	mp := []model.Property{prop("Id", "Edm.Int32", false), prop("Name", "Edm.String", true), prop("Dbl", "Edm.Double", true),
		prop("Bin", "Edm.Binary", true), prop("Col", "Collection(Edm.String)", true), prop("Kom", "Ns.Komplex", true), prop("Auf", "Ns.Farbe", true)}
	navs := []model.NavigationProperty{{Name: "Einzel", Target: "Ns.M"}, {Name: "Liste", Target: "Ns.M", Collection: true}}
	add("Ms", model.EntityType{Name: "Ns.M", Keys: []string{"Id"}, Properties: mp, NavigationProperties: navs})
	add("Os", model.EntityType{Name: "Ns.MO", Open: true, Keys: []string{"Id"}, Properties: mp, NavigationProperties: navs})

	idp, namep, extrap := prop("Id", "Edm.Int32", false), prop("Name", "Edm.String", true), prop("Extra", "Edm.String", true)
	add("Wurzeln", model.EntityType{Name: "Ns.Wurzel", Abstract: true, Keys: []string{"Id"}, Properties: []model.Property{idp}})
	add("Mitten", model.EntityType{Name: "Ns.Mitte", BaseType: "Ns.Wurzel", Keys: []string{"Id"}, Properties: []model.Property{idp, namep}})
	add("Blaetter", model.EntityType{Name: "Ns.Blatt", BaseType: "Ns.Mitte", Keys: []string{"Id"}, Properties: []model.Property{idp, namep, extrap}})
	add("", model.EntityType{Name: "Ns.Blatt2", BaseType: "Ns.Blatt", Keys: []string{"Id"}, Properties: []model.Property{idp, namep, extrap}})
	add("", model.EntityType{Name: "Ns.Abs", BaseType: "Ns.Mitte", Abstract: true, Keys: []string{"Id"}, Properties: []model.Property{idp, namep}})
	add("Fremde", model.EntityType{Name: "Ns.Fremd", Keys: []string{"Id"}, Properties: []model.Property{idp}})

	add("Rs", model.EntityType{Name: "Ns.R", Keys: []string{"Id"}, Properties: eRProps()})

	sp := []model.Property{prop("Id", "Edm.String", false), prop("an", "Edm.Int32", true), prop("bn", "Edm.Int32", true),
		prop("ap", "Edm.Int32", false), prop("bp", "Edm.Int32", false), prop("ad", "Edm.Double", true), prop("bd", "Edm.Double", true),
		prop("n1", "Edm.Int32", true), prop("n2", "Edm.Int32", true), prop("n3", "Edm.Int32", true)}
	sn := []model.NavigationProperty{{Name: "av", Target: "Ns.St"}, {Name: "bv", Target: "Ns.St"}}
	add("Sts", model.EntityType{Name: "Ns.St", Keys: []string{"Id"}, Properties: sp, NavigationProperties: sn})
	add("Sto", model.EntityType{Name: "Ns.StO", Open: true, Keys: []string{"Id"}, Properties: sp, NavigationProperties: sn})

	for i, ty := range edm {
		add("K"+eP(i), model.EntityType{Name: "Ns.K" + eP(i), Keys: []string{"K"}, Properties: []model.Property{prop("K", ty, false)}})
	}
	add("KZ2", model.EntityType{Name: "Ns.KZ2", Keys: []string{"Region", "Id"}, Properties: []model.Property{prop("Id", "Edm.Int32", false), prop("Region", "Edm.String", false)}})
	add("KZ3", model.EntityType{Name: "Ns.KZ3", Keys: []string{"A", "B", "C"}, Properties: []model.Property{prop("A", "Edm.Int32", false), prop("B", "Edm.String", false), prop("C", "Edm.Guid", false)}})
	add("OhneKey", model.EntityType{Name: "Ns.OhneKey", Properties: []model.Property{prop("X", "Edm.String", true)}})
	sets = append(sets, model.EntitySet{Name: "OhneTyp", EntityType: "Ns.Fehlt"})
	return model.NeuesServiceModel("4.0", []string{"Ns"}, typen, sets, nil)
}

func eWNamen() []string { return []string{"S", "B", "I32", "I64", "Dec", "G", "D", "T"} }

// eRProps sind die Properties des Typs R in der Reihenfolge des Modells: der Schlüssel Id (Nullable wahr), die acht nicht
// nullbaren R0 bis R7, die acht nullbaren N0 bis N7 und die zwei nicht unterstützten, nicht nullbaren U0 und U1, gemischt.
func eRProps() []model.Property {
	edm := model.UnterstuetzteEDMTypen()
	ps := []model.Property{prop("Id", "Edm.Int32", true)}
	for i, ty := range edm {
		ps = append(ps, prop("N"+eP(i), ty, true), prop("R"+eP(i), ty, false))
		if i == 0 {
			ps = append(ps, prop("U0", "Edm.Double", false))
		}
		if i == 2 {
			ps = append(ps, prop("U1", "Edm.Binary", false))
		}
	}
	return ps
}

// eRErforderlich sind die neun erforderlichen Properties von R in der Reihenfolge des Modells mit einem gültigen Wert.
func eRErforderlich() [][2]string {
	werte := []string{`"x"`, `true`, `7`, `8`, `9.5`, `"0f8fad5b-d9cb-469f-a165-70867728950e"`, `"2020-01-01"`, `"2020-01-01T00:00:00Z"`}
	out := [][2]string{{"Id", "1"}}
	for i := range werte {
		out = append(out, [2]string{"R" + eP(i), werte[i]})
	}
	return out
}
