package support_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

func TestPortSignaturen(t *testing.T) {
	clock := reflect.TypeFor[support.Clock]()
	if clock.NumMethod() != 1 {
		t.Fatalf("Clock hat %d Methoden, erwartet genau Now", clock.NumMethod())
	}
	if m, ok := clock.MethodByName("Now"); !ok || m.Type != reflect.TypeFor[func() time.Time]() {
		t.Errorf("Now hat die Form %v, erwartet func() time.Time", m.Type)
	}
	ids := reflect.TypeFor[support.IDGenerator]()
	for i := range ids.NumMethod() {
		if m := ids.Method(i); m.Type != reflect.TypeFor[func() string]() {
			t.Errorf("%s hat die Form %v, erwartet func() string", m.Name, m.Type)
		}
	}
}
