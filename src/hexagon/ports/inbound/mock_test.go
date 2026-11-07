package inbound_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestMockPort hält Z23: MockUseCase hat die eine Methode Mock mit der Signatur (ctx, Anfrage) → (Antwort, error); der
// Port wächst nur additiv.
func TestMockPort(t *testing.T) {
	typ := reflect.TypeFor[inbound.MockUseCase]()
	if typ.NumMethod() != 1 || typ.Method(0).Name != "Mock" {
		t.Fatalf("Methoden des Ports: %d, erste %q; erwartet genau Mock", typ.NumMethod(), typ.Method(0).Name)
	}
	soll := reflect.TypeFor[func(context.Context, inbound.Anfrage) (inbound.Antwort, error)]()
	if got := typ.Method(0).Type; got != soll {
		t.Errorf("Signatur %v, erwartet %v", got, soll)
	}
}
