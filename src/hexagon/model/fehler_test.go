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
		"ODR-E2001": {[]int{0}, "config"}, "ODR-E2101": {[]int{401}, "admin"}, "ODR-E2401": {[]int{0}, "record"},
		"ODR-E3001": {[]int{500}, "storage"}, "ODR-E4001": {[]int{502, 504}, "upstream"},
		"ODR-E5001": {[]int{404}, "replay"}, "ODR-E5002": {[]int{410}, "replay"}, "ODR-E5003": {[]int{400}, "replay"},
		"ODR-E5004": {[]int{409}, "replay"}, "ODR-E5005": {[]int{404}, "service"}, "ODR-E5006": {[]int{400}, "replay"},
		"ODR-E5007": {[]int{409}, "replay"}, "ODR-E5008": {[]int{422}, "replay"}, "ODR-E5009": {[]int{409}, "record"},
		"ODR-E6001": {[]int{501}, "mock"}, "ODR-E6002": {[]int{501}, "mock"}, "ODR-E6003": {[]int{422}, "mock"},
		"ODR-E6004": {[]int{422, 0}, "import"}, "ODR-E6005": {[]int{400}, "mock"}, "ODR-E6101": {[]int{422}, "import"}, "ODR-E6102": {[]int{409}, "import"},
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
		for _, fremd := range []int{200, 418, 503} {
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
