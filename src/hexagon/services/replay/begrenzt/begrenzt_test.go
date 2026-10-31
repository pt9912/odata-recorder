package begrenzt_test

import (
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/services/replay/begrenzt"
)

// TestGrenze hält die Zusage des Puffers: genau die Grenze wird angenommen, ein Byte mehr ist ErrGrenze und wird
// nicht angenommen, auch über mehrere Writes; ohne Bytes ist Bytes nil.
func TestGrenze(t *testing.T) {
	for max := int64(1); max <= 5; max++ {
		for _, teile := range [][]int{{int(max)}, {int(max) + 1}, {1, int(max)}, {int(max), 1}} {
			p := begrenzt.Neu(max)
			var summe int64
			var err error
			for _, n := range teile {
				var k int
				k, err = p.Write(make([]byte, n))
				summe += int64(k)
				if err != nil {
					break
				}
			}
			ueber := int64(0)
			for _, n := range teile {
				ueber += int64(n)
			}
			if ueber <= max && (err != nil || summe != ueber) || ueber > max && (!errors.Is(err, begrenzt.ErrGrenze) || summe != max) {
				t.Errorf("max %d, Writes %v: angenommen %d, Fehler %v", max, teile, summe, err)
			}
			if int64(len(p.Bytes())) != summe {
				t.Errorf("max %d, Writes %v: Bytes %d, angenommen %d", max, teile, len(p.Bytes()), summe)
			}
		}
	}
	if begrenzt.Neu(3).Bytes() != nil {
		t.Error("leerer Puffer liefert nicht nil")
	}
}
