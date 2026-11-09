package hybrid_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
)

// TestNew hält Z01: New weist jedes Feld der Konfig ab (kein Dienst, leerer Root an erster und an späterer Stelle,
// nil in Mock, Replay, Policy, MaxBytes 0 und -1) und nimmt MaxBytes 1 an; es ruft keinen Port.
func TestNew(t *testing.T) {
	m, r := &stufe{}, &stufe{}
	gut := func() hybrid.Konfig {
		return hybrid.Konfig{
			Dienste:  []hybrid.Dienst{{Root: "/a"}},
			Mock:     fakeMock{m},
			Replay:   fakeReplay{r},
			Policy:   mitKategorien(t),
			MaxBytes: 1,
		}
	}
	faelle := []struct {
		name string
		aend func(*hybrid.Konfig)
		bad  bool
	}{
		{"gut, MaxBytes 1", func(*hybrid.Konfig) {}, false},
		{"gut, MaxBytes groß", func(k *hybrid.Konfig) { k.MaxBytes = 1 << 40 }, false},
		{"kein Dienst nil", func(k *hybrid.Konfig) { k.Dienste = nil }, true},
		{"kein Dienst leer", func(k *hybrid.Konfig) { k.Dienste = []hybrid.Dienst{} }, true},
		{"Root leer", func(k *hybrid.Konfig) { k.Dienste = []hybrid.Dienst{{}} }, true},
		{"Root später leer", func(k *hybrid.Konfig) { k.Dienste = []hybrid.Dienst{{Root: "/a"}, {Root: "/b"}, {}} }, true},
		{"Mock nil", func(k *hybrid.Konfig) { k.Mock = nil }, true},
		{"Replay nil", func(k *hybrid.Konfig) { k.Replay = nil }, true},
		{"Policy nil", func(k *hybrid.Konfig) { k.Policy = nil }, true},
		{"MaxBytes 0", func(k *hybrid.Konfig) { k.MaxBytes = 0 }, true},
		{"MaxBytes -1", func(k *hybrid.Konfig) { k.MaxBytes = -1 }, true},
	}
	for _, f := range faelle {
		k := gut()
		f.aend(&k)
		v, err := hybrid.New(k)
		if f.bad {
			if !errors.Is(err, model.ErrUngueltig) || v != nil {
				t.Errorf("%s: Dienst %v, Fehler %v; erwartet model.ErrUngueltig ohne Dienst", f.name, v, err)
			}
			continue
		}
		if err != nil || v == nil {
			t.Errorf("%s: Fehler %v", f.name, err)
		}
	}
	if m.zahl() != 0 || r.zahl() != 0 {
		t.Errorf("New rief einen Port: Mock %d, Replay %d", m.zahl(), r.zahl())
	}
}

// TestUngueltigeAnfrage hält Z02: ctx nil, Methode leer, Laenge -2 und Body nil bei Laenge 1 ergeben
// model.ErrUngueltig ohne Code, ohne Aufruf einer Stufe, ohne Lesen des Bodys und ohne Wert im Text; Laenge -1 und 0
// sind gültig.
func TestUngueltigeAnfrage(t *testing.T) {
	faelle := []struct {
		name string
		ctx  func() context.Context
		mod  func(*inbound.Anfrage)
		bad  bool
	}{
		{"ctx nil", func() context.Context { return nil }, func(*inbound.Anfrage) {}, true},
		{"Methode leer", context.Background, func(a *inbound.Anfrage) { a.Methode = "" }, true},
		{"Laenge -2", context.Background, func(a *inbound.Anfrage) { a.Laenge = -2 }, true},
		{"Laenge -2 ohne Body", context.Background, func(a *inbound.Anfrage) { a.Laenge, a.Body = -2, nil }, true},
		{"Body nil bei Laenge 1", context.Background, func(a *inbound.Anfrage) { a.Laenge, a.Body = 1, nil }, true},
		{"Body nil bei Laenge 99", context.Background, func(a *inbound.Anfrage) { a.Laenge, a.Body = 99, nil }, true},
		{"Laenge -1 mit Leser", context.Background, func(a *inbound.Anfrage) { a.Laenge = -1 }, false},
		{"Laenge 0 mit Leser", context.Background, func(a *inbound.Anfrage) { a.Laenge = 0 }, false},
		{"Laenge -1 ohne Body", context.Background, func(a *inbound.Anfrage) { a.Laenge, a.Body = -1, nil }, false},
		{"Laenge 0 ohne Body", context.Background, func(a *inbound.Anfrage) { a.Laenge, a.Body = 0, nil }, false},
	}
	for _, f := range faelle {
		g := neuesRig(t, offenste(t), standardMax)
		s := neuerStrom(muster(4, 1))
		a := anfrage("POST")
		a.Pfad = "/odata/svc/GeheimerWert"
		a.Body, a.Laenge = s, 4
		f.mod(&a)
		_, err := g.v.Hybrid(f.ctx(), a)
		if !f.bad {
			if err != nil || g.m.zahl() != 1 {
				t.Errorf("%s: Fehler %v, Mock %d-mal; erwartet gültig", f.name, err, g.m.zahl())
			}
			continue
		}
		var fe *inbound.Fehler
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &fe) {
			t.Errorf("%s: Fehler %v; erwartet model.ErrUngueltig ohne Code", f.name, err)
		}
		if err != nil && (strings.Contains(err.Error(), "GeheimerWert") || strings.Contains(err.Error(), "POST")) {
			t.Errorf("%s: der Text nennt einen Wert des Requests: %q", f.name, err.Error())
		}
		g.zaehlen(t, f.name, 0, 0)
		if s.reads.Load() != 0 {
			t.Errorf("%s: Body gelesen (%d)", f.name, s.reads.Load())
		}
	}
}
