package httpupstream

import "time"

// starteFrist startet die Frist vom Aufruf bis zum Kopf der Antwort (SPEC-120 Nr. 6); läuft sie ab,
// ruft sie f. Die gelieferte Funktion beendet die Frist und meldet, ob das gelang: falsch heißt,
// sie lief ab.
func (c *Client) starteFrist(f func()) func() bool {
	if c.neueFrist != nil {
		return c.neueFrist(c.frist, f)
	}
	return time.AfterFunc(c.frist, f).Stop
}
