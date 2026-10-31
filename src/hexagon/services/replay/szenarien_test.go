package replay_test

import (
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// szenario baut für einen Code der Menge aus SPEC-135 Nr. 15 den Zustand, in dem Replay mit ihm endet, und
// liefert den Request. strat ist die Strategie der Konfig.
type szenario struct {
	code model.Code
	// status ist der Status des Codes (SPEC-109 Nr. 2), unabhängig vom Dienst notiert.
	status int
	baue   func(u *umg, strat model.ReplayStrategy) inbound.Anfrage
}

const canary = "CANARY-7f3a"

// szenarien sind die 13 erreichbaren Codes (SPEC-135 Nr. 15), je mit dem Zustand, der ihn auslöst.
func szenarien() []szenario {
	nurA := func(u *umg) { u.lege(u.austausch(1, u.keyA(), true)) }
	return []szenario{
		{model.CodeE3001, 500, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage {
			nurA(u)
			u.p.setzeFehler("Blobs.Open", errors.Join(outbound.ErrBlobFehlt, errors.New(canary)))
			return anf()
		}},
		{model.CodeE5001, 404, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage { return anf() }},
		{model.CodeE5002, 410, func(u *umg, s model.ReplayStrategy) inbound.Anfrage {
			nurA(u)
			u.cursor(s, 1)
			return anf()
		}},
		{model.CodeE5003, 400, func(_ *umg, _ model.ReplayStrategy) inbound.Anfrage {
			a := anf()
			delete(a.Header, "X-Recorder-Workspace")
			return a
		}},
		{model.CodeE5004, 409, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage {
			u.sess.setze(model.Session{ID: sesID, ServiceID: "svc-b", CreatedAt: zeit(), Status: model.SessionComplete})
			return anf()
		}},
		{model.CodeE5005, 404, func(_ *umg, _ model.ReplayStrategy) inbound.Anfrage {
			a := anf()
			a.Pfad = "/x"
			return a
		}},
		{model.CodeE5006, 400, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage {
			u.k.Regeln.InvalidEncoding = model.KodierungReject
			a := anf()
			a.RawQuery = "a=%zz"
			return a
		}},
		{model.CodeE5007, 409, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage {
			u.lege(u.austausch(1, u.keyB(), true), u.austausch(2, u.keyA(), true))
			u.cursor(model.StrategyStrict)
			return anf()
		}},
		{model.CodeE5008, 422, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage {
			u.lege(u.austausch(1, u.keyA(), false))
			return anf()
		}},
		{model.CodeE7001, 404, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage {
			z := u.st.stand()
			delete(z.ws, wsID)
			u.st.setze(z)
			return anf()
		}},
		{model.CodeE7003, 404, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage {
			u.sess.weg(sesID)
			return anf()
		}},
		{model.CodeE7009, 409, func(u *umg, _ model.ReplayStrategy) inbound.Anfrage {
			u.sess.setze(model.Session{ID: sesID, ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionOpen})
			return anf()
		}},
		{model.CodeE9001, 500, func(_ *umg, _ model.ReplayStrategy) inbound.Anfrage {
			a := anf()
			a.Pfad = root + "/a?b"
			return a
		}},
	}
}
