package model_test

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func TestFehlerobjektJSON(t *testing.T) {
	obj, err := model.NeuesFehlerobjekt(model.CodeE5001)
	if err != nil {
		t.Fatal(err)
	}
	roh, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	var aussen map[string]map[string]any
	if err := json.Unmarshal(roh, &aussen); err != nil {
		t.Fatal(err)
	}
	if len(aussen) != 1 || aussen["error"] == nil {
		t.Fatalf("Schlüssel der Oberfläche: %s", roh)
	}
	var schluessel []string
	for k, v := range aussen["error"] {
		if _, ok := v.(string); !ok {
			t.Errorf("Feld %s ist keine Zeichenkette", k)
		}
		schluessel = append(schluessel, k)
	}
	slices.Sort(schluessel)
	if !slices.Equal(schluessel, []string{"code", "message", "target"}) {
		t.Fatalf("Felder %v, erwartet code, message, target (kein Status-Feld)", schluessel)
	}
	if aussen["error"]["code"] != "ODR-E5001" || aussen["error"]["target"] != "replay" {
		t.Errorf("Werte %s", roh)
	}
	wieder, _ := json.Marshal(obj)
	if string(wieder) != string(roh) {
		t.Error("zweite Serialisierung weicht ab")
	}
}

func TestFehlerobjektUngueltigerCode(t *testing.T) {
	for _, c := range []model.Code{0, -1, 999} {
		if _, err := model.NeuesFehlerobjekt(c); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("Code %d: %v", int(c), err)
		}
		if c.Gueltig() || c.String() != "" || c.StatusErlaubt(0) || c.StatusErlaubt(500) || c.Message() != "" {
			t.Errorf("Code %d trägt Eigenschaften", int(c))
		}
	}
}

func TestCodeTabelle(t *testing.T) {
	// Spiegel von SPEC-109 Nr. 2; dass er der Spezifikation entspricht, gleicht der Review ab.
	soll := map[string]struct {
		stati []int
		ziel  model.Zielbereich
	}{
		"ODR-E2001": {[]int{0}, "config"}, "ODR-E2101": {[]int{401}, "admin"}, "ODR-E2102": {[]int{400, 404, 405}, "admin"}, "ODR-E2401": {[]int{0}, "record"},
		"ODR-E3001": {[]int{500}, "storage"}, "ODR-E4001": {[]int{502, 504}, "upstream"},
		"ODR-E5001": {[]int{404}, "replay"}, "ODR-E5002": {[]int{410}, "replay"}, "ODR-E5003": {[]int{400}, "replay"},
		"ODR-E5004": {[]int{409}, "replay"}, "ODR-E5005": {[]int{404}, "service"}, "ODR-E5006": {[]int{400}, "replay"},
		"ODR-E5007": {[]int{409}, "replay"}, "ODR-E5008": {[]int{422}, "replay"}, "ODR-E5009": {[]int{409}, "record"},
		"ODR-E5010": {[]int{503}, "service"},
		"ODR-E6001": {[]int{501}, "mock"}, "ODR-E6002": {[]int{501}, "mock"}, "ODR-E6003": {[]int{422}, "mock"},
		"ODR-E6004": {[]int{422, 0}, "import"}, "ODR-E6005": {[]int{400}, "mock"}, "ODR-E6006": {[]int{501}, "mock"}, "ODR-E6007": {[]int{400}, "mock"}, "ODR-E6008": {[]int{404}, "mock"}, "ODR-E6009": {[]int{409}, "mock"}, "ODR-E6010": {[]int{412}, "mock"}, "ODR-E6101": {[]int{422}, "import"}, "ODR-E6102": {[]int{409}, "import"},
		"ODR-E7001": {[]int{404}, "workspace"}, "ODR-E7002": {[]int{404}, "workspace"}, "ODR-E7003": {[]int{404}, "session"},
		"ODR-E7004": {[]int{409}, "session"}, "ODR-E7005": {[]int{404}, "import"}, "ODR-E7006": {[]int{409}, "workspace"},
		"ODR-E7007": {[]int{409}, "workspace"}, "ODR-E7008": {[]int{422}, "workspace"}, "ODR-E7009": {[]int{409}, "session"},
		"ODR-E7010": {[]int{409}, "session"}, "ODR-E7011": {[]int{409}, "session"}, "ODR-E7012": {[]int{404}, "session"},
		"ODR-E9001": {[]int{500}, "internal"},
	}
	alle := model.AlleCodes()
	if len(alle) != len(soll) {
		t.Fatalf("%d Codes im Modell, %d in §4", len(alle), len(soll))
	}
	// ODR-E2102 steht hinter ODR-E2101, mit dem festen Text und den drei Status 400, 404 und 405 (SPEC-240); jeder andere
	// ganzzahlige Status von 0 bis 999 ist ausgeschlossen.
	if i := slices.Index(alle, model.CodeE2101); alle[i+1] != model.CodeE2102 || model.CodeE2102.Message() != "Admin request invalid" {
		t.Errorf("ODR-E2102 steht nicht hinter ODR-E2101 oder trägt die message %q", model.CodeE2102.Message())
	}
	for s := 0; s < 1000; s++ {
		if model.CodeE2102.StatusErlaubt(s) != (s == 400 || s == 404 || s == 405) {
			t.Errorf("ODR-E2102: Status %d erlaubt = %v", s, s != 400 && s != 404 && s != 405)
		}
	}
	if model.CodeE2102.String() != "ODR-E2102" || model.CodeE2102.Zielbereich() != "admin" {
		t.Errorf("ODR-E2102: Text %q Ziel %q", model.CodeE2102.String(), model.CodeE2102.Zielbereich())
	}
	if obj, err := model.NeuesFehlerobjekt(model.CodeE2102); err != nil ||
		obj.Error.Code != "ODR-E2102" || obj.Error.Message != "Admin request invalid" || obj.Error.Target != "admin" {
		t.Errorf("Fehlerobjekt zu ODR-E2102: %+v (%v)", obj, err)
	}
	// ODR-E6009 steht hinter ODR-E6008 (W67), mit dem festen Text und dem einen Status 409 des Duplicate Key.
	if i := slices.Index(alle, model.CodeE6008); alle[i+1] != model.CodeE6009 || model.CodeE6009.Message() != "Entity key already exists" {
		t.Errorf("ODR-E6009 steht nicht hinter ODR-E6008 oder trägt die message %q", model.CodeE6009.Message())
	}
	for s := 0; s < 1000; s++ {
		if model.CodeE6009.StatusErlaubt(s) != (s == 409) {
			t.Errorf("ODR-E6009: Status %d erlaubt = %v", s, s != 409)
		}
	}
	if model.CodeE6009.String() != "ODR-E6009" || model.CodeE6009.Zielbereich() != "mock" {
		t.Errorf("ODR-E6009: Text %q Ziel %q", model.CodeE6009.String(), model.CodeE6009.Zielbereich())
	}
	// ODR-E6010 steht hinter ODR-E6009 (F22), mit dem festen Text und dem einen Status 412 der Vorbedingung (SPEC-283 Nr. 7); die Zahl
	// der Codes ist 41.
	if i := slices.Index(alle, model.CodeE6009); alle[i+1] != model.CodeE6010 || model.CodeE6010.Message() != "Precondition failed" {
		t.Errorf("ODR-E6010 steht nicht hinter ODR-E6009 oder trägt die message %q", model.CodeE6010.Message())
	}
	for s := 0; s < 1000; s++ {
		if model.CodeE6010.StatusErlaubt(s) != (s == 412) {
			t.Errorf("ODR-E6010: Status %d erlaubt = %v", s, s != 412)
		}
	}
	if model.CodeE6010.String() != "ODR-E6010" || model.CodeE6010.Zielbereich() != "mock" || len(alle) != 41 {
		t.Errorf("ODR-E6010: Text %q Ziel %q, %d Codes, erwartet 41", model.CodeE6010.String(), model.CodeE6010.Zielbereich(), len(alle))
	}
	if obj, err := model.NeuesFehlerobjekt(model.CodeE6010); err != nil ||
		obj.Error.Code != "ODR-E6010" || obj.Error.Message != "Precondition failed" || obj.Error.Target != "mock" {
		t.Errorf("Fehlerobjekt zu ODR-E6010: %+v (%v)", obj, err)
	}
	// ODR-E6008 steht hinter ODR-E6007, mit dem festen Text der Entscheidung.
	if i := slices.Index(alle, model.CodeE6007); alle[i+1] != model.CodeE6008 || model.CodeE6008.Message() != "Entity or entity set not found" {
		t.Errorf("ODR-E6008 steht nicht hinter ODR-E6007 oder trägt die message %q", model.CodeE6008.Message())
	}
	// ODR-E6007 steht hinter ODR-E6006, mit dem festen Text der Entscheidung.
	if i := slices.Index(alle, model.CodeE6006); alle[i+1] != model.CodeE6007 || model.CodeE6007.Message() != "Invalid request body" {
		t.Errorf("ODR-E6007 steht nicht hinter ODR-E6006 oder trägt die message %q", model.CodeE6007.Message())
	}
	// ODR-E6006 steht hinter ODR-E6005, mit dem festen Text der Entscheidung; CodeE9001 ist der letzte Code der Tabelle.
	if i := slices.Index(alle, model.CodeE6005); alle[i+1] != model.CodeE6006 || model.CodeE6006.Message() != "Unsupported OData V2 feature" {
		t.Errorf("ODR-E6006 steht nicht hinter ODR-E6005 oder trägt die message %q", model.CodeE6006.Message())
	}
	if alle[len(alle)-1] != model.CodeE9001 || (model.CodeE9001 + 1).Gueltig() {
		t.Error("ODR-E9001 ist nicht der letzte Code, oder der Code dahinter ist gültig")
	}
	for _, c := range alle {
		erwartet, ok := soll[c.String()]
		if !ok {
			t.Errorf("%s steht nicht in der Tabelle des Tests", c)
			continue
		}
		if !slices.Equal(c.Stati(), erwartet.stati) || c.Zielbereich() != erwartet.ziel {
			t.Errorf("%s: Status %v Ziel %s, erwartet %v %s", c, c.Stati(), c.Zielbereich(), erwartet.stati, erwartet.ziel)
		}
		if again, ok := model.ParseCode(c.String()); !ok || again != c {
			t.Errorf("ParseCode(%s) = %v", c, again)
		}
		for _, fremd := range []int{200, 418, 507} {
			if c.StatusErlaubt(fremd) {
				t.Errorf("%s erlaubt Status %d", c, fremd)
			}
		}
	}
	if !model.CodeE4001.StatusErlaubt(502) || !model.CodeE4001.StatusErlaubt(504) || model.CodeE4001.StatusErlaubt(500) || model.CodeE4001.StatusErlaubt(0) {
		t.Error("ODR-E4001 trägt nur 502 oder 504")
	}
	if _, ok := model.ParseCode("ODR-E0000"); ok {
		t.Error("ODR-E0000 ist kein Code")
	}
}

func TestMessageFest(t *testing.T) {
	gesehen := map[string]model.Code{}
	for _, c := range model.AlleCodes() {
		a, err := model.NeuesFehlerobjekt(c)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := model.NeuesFehlerobjekt(c)
		m := a.Error.Message
		if strings.TrimSpace(m) == "" || m != b.Error.Message || m != c.Message() {
			t.Errorf("%s: message %q nicht fest oder leer", c, m)
		}
		if strings.ContainsAny(m, "%{}<>\n") {
			t.Errorf("%s: message %q sieht nach einem eingesetzten Wert aus", c, m)
		}
		if vorher, doppelt := gesehen[m]; doppelt {
			t.Errorf("%s und %s teilen die message %q", c, vorher, m)
		}
		gesehen[m] = c
	}
}
