package admin_test

import (
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// specZeile ist eine Zeile der Tabelle von SPEC-109 Nr. 2 als Literal des Vertrags: der Code, die Status, die der Code
// als HTTP-Antwort tragen darf (nil heißt: der Code wird nie ausgeliefert), und der Zielbereich.
type specZeile struct {
	code   string
	stati  []int
	target string
}

// specCodes ist die Tabelle von SPEC-109 Nr. 2 in der Reihenfolge der Codes; das Orakel von E1, B1 und B2.
func specCodes() []specZeile {
	return []specZeile{
		{"ODR-E2001", nil, "config"}, {"ODR-E2101", []int{401}, "admin"}, {"ODR-E2102", []int{400, 404, 405}, "admin"},
		{"ODR-E2401", nil, "record"}, {"ODR-E3001", []int{500}, "storage"}, {"ODR-E4001", []int{502, 504}, "upstream"},
		{"ODR-E5001", []int{404}, "replay"}, {"ODR-E5002", []int{410}, "replay"}, {"ODR-E5003", []int{400}, "replay"},
		{"ODR-E5004", []int{409}, "replay"}, {"ODR-E5005", []int{404}, "service"}, {"ODR-E5006", []int{400}, "replay"},
		{"ODR-E5007", []int{409}, "replay"}, {"ODR-E5008", []int{422}, "replay"}, {"ODR-E5009", []int{409}, "record"},
		{"ODR-E5010", []int{503}, "service"}, {"ODR-E6001", []int{501}, "mock"}, {"ODR-E6002", []int{501}, "mock"},
		{"ODR-E6003", []int{422}, "mock"}, {"ODR-E6004", []int{422}, "import"}, {"ODR-E6005", []int{400}, "mock"},
		{"ODR-E6006", []int{501}, "mock"}, {"ODR-E6007", []int{400}, "mock"}, {"ODR-E6008", []int{404}, "mock"},
		{"ODR-E6009", []int{409}, "mock"}, {"ODR-E6010", []int{412}, "mock"}, {"ODR-E6101", []int{422}, "import"},
		{"ODR-E6102", []int{409}, "import"}, {"ODR-E7001", []int{404}, "workspace"}, {"ODR-E7002", []int{404}, "workspace"},
		{"ODR-E7003", []int{404}, "session"}, {"ODR-E7004", []int{409}, "session"}, {"ODR-E7005", []int{404}, "import"},
		{"ODR-E7006", []int{409}, "workspace"}, {"ODR-E7007", []int{409}, "workspace"}, {"ODR-E7008", []int{422}, "workspace"},
		{"ODR-E7009", []int{409}, "session"}, {"ODR-E7010", []int{409}, "session"}, {"ODR-E7011", []int{409}, "session"},
		{"ODR-E7012", []int{404}, "session"}, {"ODR-E9001", []int{500}, "internal"},
	}
}

// modellCode liefert den Code des Modells zur Zeile i der Tabelle; die Zählprüfung hält beide Folgen gleich.
func modellCode(i int) model.Code { return model.AlleCodes()[i] }

// alleStati liefert die Vereinigung der Status der Tabelle, aufsteigend und ohne Doppelte.
func alleStati() []int {
	var out []int
	for _, z := range specCodes() {
		for _, s := range z.stati {
			if !slices.Contains(out, s) {
				out = append(out, s)
			}
		}
	}
	slices.Sort(out)
	return out
}

// fehlerBody liefert das Fehlerobjekt der Zeile z als Text des Vertrags (SPEC-238 Nr. 5): `message` ist der feste Text
// des Codes (SPEC-109 Nr. 3), dessen Wortlaut kein Vertrag ist.
func fehlerBody(z specZeile, c model.Code) string {
	return `{"error":{"code":"` + z.code + `","message":"` + c.Message() + `","target":"` + z.target + `"}}`
}

// TestSpecCodesGegenQuelle hält M-Code × Status gegen die Quelle: die Literal-Tabelle und model.AlleCodes() sind gleich
// lang und gleich in Code, Status (ohne 0) und Zielbereich; der Text des Codes ist `ODR-E` und vier Ziffern.
func TestSpecCodesGegenQuelle(t *testing.T) {
	spec, alle := specCodes(), model.AlleCodes()
	if len(spec) != len(alle) {
		t.Fatalf("Tabelle des Tests %d Zeilen, model.AlleCodes() %d", len(spec), len(alle))
	}
	for i, z := range spec {
		c := modellCode(i)
		var stati []int
		for _, s := range c.Stati() {
			if s != 0 {
				stati = append(stati, s)
			}
		}
		if c.String() != z.code || !slices.Equal(stati, z.stati) || string(c.Zielbereich()) != z.target {
			t.Errorf("Zeile %d: Modell %s %v %s, Tabelle %s %v %s", i, c.String(), stati, c.Zielbereich(), z.code, z.stati, z.target)
		}
	}
	if got := alleStati(); !slices.Equal(got, []int{400, 401, 404, 405, 409, 410, 412, 422, 500, 501, 502, 503, 504}) {
		t.Errorf("Vereinigung der Status %v", got)
	}
}
