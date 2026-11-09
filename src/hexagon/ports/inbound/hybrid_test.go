package inbound_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestHybridPort hält Z01 für den Port: HybridUseCase hat die eine Methode Hybrid mit der Signatur (ctx, Anfrage) →
// (Antwort, error); der Port wächst nur additiv.
func TestHybridPort(t *testing.T) {
	typ := reflect.TypeFor[inbound.HybridUseCase]()
	if typ.NumMethod() != 1 || typ.Method(0).Name != "Hybrid" {
		t.Fatalf("Methoden des Ports: %d, erste %q; erwartet genau Hybrid", typ.NumMethod(), typ.Method(0).Name)
	}
	soll := reflect.TypeFor[func(context.Context, inbound.Anfrage) (inbound.Antwort, error)]()
	if got := typ.Method(0).Type; got != soll {
		t.Errorf("Signatur %v, erwartet %v", got, soll)
	}
}
