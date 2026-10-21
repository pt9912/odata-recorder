package capture_test

import (
	"context"
	"errors"
	"io"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

func TestNewArgumente(t *testing.T) {
	gueltig := func() capture.Konfig {
		return capture.Konfig{
			Blobs: neueBlobs(), Pfade: []model.JSONPfad{pfad(t, "$..password")},
			MaxBytes: 5, QueueBytes: 5, Frist: time.Second, MaxVerhaeltnis: 5,
		}
	}
	faelle := []struct {
		name string
		aend func(*capture.Konfig)
	}{
		{"Blobs nil", func(k *capture.Konfig) { k.Blobs = nil }},
		{"MaxBytes 0", func(k *capture.Konfig) { k.MaxBytes = 0 }},
		{"MaxBytes -1", func(k *capture.Konfig) { k.MaxBytes = -1 }},
		{"QueueBytes 0", func(k *capture.Konfig) { k.QueueBytes = 0 }},
		{"QueueBytes -1", func(k *capture.Konfig) { k.QueueBytes = -1 }},
		{"MaxVerhaeltnis 0", func(k *capture.Konfig) { k.MaxVerhaeltnis = 0 }},
		{"MaxVerhaeltnis -1", func(k *capture.Konfig) { k.MaxVerhaeltnis = -1 }},
		{"Frist 0", func(k *capture.Konfig) { k.Frist = 0 }},
		{"Frist -1 s", func(k *capture.Konfig) { k.Frist = -time.Second }},
		{"Nullwert in Pfade", func(k *capture.Konfig) { k.Pfade = append(k.Pfade, model.JSONPfad{}) }},
		{"Nullwert allein in Pfade", func(k *capture.Konfig) { k.Pfade = []model.JSONPfad{{}} }},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			k := gueltig()
			f.aend(&k)
			p, err := capture.New(k)
			if !errors.Is(err, model.ErrUngueltig) || p != nil {
				t.Errorf("New = %v, %v; erwartet ErrUngueltig und keine Pipeline", p, err)
			}
		})
	}
	gut := []struct {
		name string
		aend func(*capture.Konfig)
	}{
		{"Pfade nil", func(k *capture.Konfig) { k.Pfade = nil }},
		{"Pfade leer", func(k *capture.Konfig) { k.Pfade = []model.JSONPfad{} }},
		{"alle Grenzen genau 1", func(k *capture.Konfig) {
			k.MaxBytes, k.QueueBytes, k.MaxVerhaeltnis, k.Frist = 1, 1, 1, time.Nanosecond
		}},
		{"nur MaxBytes 1", func(k *capture.Konfig) { k.MaxBytes = 1 }},
		{"nur QueueBytes 1", func(k *capture.Konfig) { k.QueueBytes = 1 }},
		{"nur MaxVerhaeltnis 1", func(k *capture.Konfig) { k.MaxVerhaeltnis = 1 }},
		{"nur Frist 1 ns", func(k *capture.Konfig) { k.Frist = time.Nanosecond }},
		{"QueueBytes über MaxBytes", func(k *capture.Konfig) { k.QueueBytes = 1 << 40 }},
	}
	for _, f := range gut {
		t.Run(f.name, func(t *testing.T) {
			k := gueltig()
			f.aend(&k)
			if p, err := capture.New(k); err != nil || p == nil {
				t.Errorf("New = %v, %v; erwartet eine Pipeline", p, err)
			}
		})
	}
}

func TestStandard(t *testing.T) {
	blobs := neueBlobs()
	pfade := []model.JSONPfad{pfad(t, "$..password"), pfad(t, "$..token")}
	k := capture.Standard(blobs, pfade, 7)
	if k.QueueBytes != 16<<20 || k.Frist != 30*time.Second || k.MaxVerhaeltnis != 20 {
		t.Errorf("Standard setzt %d, %v, %d; erwartet 16 MiB, 30 s, 20", k.QueueBytes, k.Frist, k.MaxVerhaeltnis)
	}
	if k.Blobs != blobs || !slices.Equal(k.Pfade, pfade) || k.MaxBytes != 7 {
		t.Errorf("Standard reicht Blobs, Pfade oder MaxBytes nicht durch: %+v", k)
	}
	if _, err := capture.New(k); err != nil {
		t.Errorf("New(Standard) = %v", err)
	}
}

func TestTeeArgumente(t *testing.T) {
	var keinCtx context.Context
	faelle := []struct {
		name string
		tee  func(p *capture.Pipeline, src io.Reader) (*capture.Aufnahme, error)
	}{
		{"ctx nil", func(p *capture.Pipeline, src io.Reader) (*capture.Aufnahme, error) {
			return p.Tee(keinCtx, src, antwort())
		}},
		{"src nil", func(p *capture.Pipeline, _ io.Reader) (*capture.Aufnahme, error) {
			return p.Tee(t.Context(), nil, antwort())
		}},
		{"Richtung Nullwert", func(p *capture.Pipeline, src io.Reader) (*capture.Aufnahme, error) {
			return p.Tee(t.Context(), src, capture.Eingang{})
		}},
		{"Richtung 99", func(p *capture.Pipeline, src io.Reader) (*capture.Aufnahme, error) {
			return p.Tee(t.Context(), src, capture.Eingang{Richtung: capture.Richtung(99)})
		}},
		{"Metadata bei Request", func(p *capture.Pipeline, src io.Reader) (*capture.Aufnahme, error) {
			return p.Tee(t.Context(), src, capture.Eingang{Richtung: capture.Request, Metadata: true})
		}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			p := u.pipe()
			frei := p.AnhaltenAmStart()
			defer frei()
			src := quelle(`{"a":1}`)
			vorher := runtime.NumGoroutine()
			a, err := f.tee(p, src)
			if !errors.Is(err, model.ErrUngueltig) || a != nil {
				t.Errorf("Tee = %v, %v; erwartet ErrUngueltig und keine Aufnahme", a, err)
			}
			if n := src.anzahlReads(); n != 0 {
				t.Errorf("src wurde %d-mal gelesen", n)
			}
			if st := u.blobs.stand(); st.begins != 0 {
				t.Errorf("Begin %d-mal gerufen", st.begins)
			}
			if nachher := runtime.NumGoroutine(); nachher > vorher {
				t.Errorf("Goroutinen %d -> %d: Tee startete eine Goroutine", vorher, nachher)
			}
		})
	}
	gut := []struct {
		name string
		ein  capture.Eingang
	}{
		{"Metadata bei Response", capture.Eingang{Richtung: capture.Response, ContentType: []string{"application/xml"}, Metadata: true}},
		{"Request ohne Metadata", anfrage()},
	}
	for _, f := range gut {
		t.Run(f.name, func(t *testing.T) {
			u := neu(t)
			a, err := u.pipe().Tee(context.Background(), quelle(), f.ein)
			if err != nil || a == nil {
				t.Fatalf("Tee = %v, %v", a, err)
			}
			_, _ = io.Copy(io.Discard, a)
			_ = abwarten(t, a)
		})
	}
}

func TestTeeCtxBeendet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	u := neu(t)
	a, err := u.pipe().Tee(ctx, quelle(`{"a":1}`), antwort())
	if err != nil || a == nil {
		t.Fatalf("Tee mit beendetem ctx = %v, %v; erwartet eine Aufnahme", a, err)
	}
	e := abwarten(t, a)
	verworfen(t, e, capture.Response, capture.Abbruch, capture.GrundAbbruch)
	if !errors.Is(e.Fehler, context.Canceled) {
		t.Errorf("Ursache %v ist nicht context.Canceled", e.Fehler)
	}
	if st := u.blobs.stand(); st.begins != 0 {
		t.Errorf("Begin %d-mal gerufen", st.begins)
	}

	// encoding geht einem Abbruch von ctx vor Tee vor, auch bei leerem Body (SPEC-119 Nr. 9).
	for _, src := range []*skript{quelle(`{"a":1}`), quelle()} {
		b, err := neu(t).pipe().Tee(ctx, src, mitEncoding(antwort(), "br"))
		if err != nil {
			t.Fatalf("Tee: %v", err)
		}
		verworfen(t, abwarten(t, b), capture.Response, capture.Verstoss, capture.GrundEncoding)
	}
}
