package replay

import "github.com/pt9912/odata-recorder/src/hexagon/model"

// Hilfe für die externen Tests: die Auswahl, die TestCursorTripelErreichbar über alle Bestände ruft.

// Waehle ruft waehle.
func Waehle(s model.ReplayStrategy, key string, liste []model.Exchange, redigierteBodies bool, consumed []int64, position int64) (model.Exchange, error) {
	return waehle(s, key, liste, redigierteBodies, consumed, position)
}
