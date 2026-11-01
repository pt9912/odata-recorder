package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Hochfahren führt Wartung, Services anlegen, erste Sessions (nur record) und den Listener in
// dieser Reihenfolge aus (SPEC-139 Nr. 6). Vor jedem Schritt prüft es ctx; ein abgebrochener
// Kontext und jeder Fehler eines Schritts beenden es an der Stelle.
func (h *Handle) Hochfahren(ctx context.Context) error {
	h.mu.Lock()
	if h.zustand != gebaut {
		h.mu.Unlock()
		return ungueltig("Hochfahren läuft einmal je Handle")
	}
	h.zustand = hochgefahren
	h.mu.Unlock()
	schritte := []func(context.Context) error{h.startWartung, h.legeDiensteAn}
	if h.k.Modus == modusRecord {
		schritte = append(schritte, h.oeffneErsteSessions)
	}
	schritte = append(schritte, h.starteListener)
	for _, schritt := range schritte {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := schritt(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handle) startWartung(ctx context.Context) error {
	_, err := h.wartung.Startup(ctx)
	return err
}

// legeDiensteAn legt jeden Service an, den die Ablage nicht kennt; ein gespeicherter mit anderem
// service_root oder anderer odata_version verhindert den Start (SPEC-107 Nr. 14).
func (h *Handle) legeDiensteAn(ctx context.Context) error {
	for _, d := range h.k.Dienste {
		err := h.dienste.Create(ctx, model.Service{ID: d.ID, ServiceRoot: d.Root, ODataVersion: odataVersion})
		if err == nil {
			continue
		}
		if !errors.Is(err, outbound.ErrBereitsVorhanden) {
			return err
		}
		gespeichert, err := h.dienste.Get(ctx, d.ID)
		if err != nil {
			return err
		}
		if schluessel := abweichungen(gespeichert, d); len(schluessel) > 0 {
			zeilen := make([]string, len(schluessel))
			for i, s := range schluessel {
				h.log.Error("start: gespeicherter Service weicht von der Konfiguration ab",
					"ereignis", "dienst_abweichend", "error_code", model.CodeE2001.String(), "service_id", d.ID)
				zeilen[i] = konfigZeile(s, "gespeicherter Service gleicher id hat "+artikel(s))
			}
			return &KonfigFehler{Zeilen: zeilen}
		}
	}
	return nil
}

// abweichungen nennt die Schlüssel, in denen der gespeicherte Service von der Konfiguration abweicht,
// service_root vor odata_version (SPEC-139 Nr. 6).
func abweichungen(gespeichert model.Service, d Dienst) []string {
	var schluessel []string
	if gespeichert.ServiceRoot != d.Root {
		schluessel = append(schluessel, "service_root")
	}
	if gespeichert.ODataVersion != odataVersion {
		schluessel = append(schluessel, "odata_version")
	}
	return schluessel
}

func artikel(schluessel string) string {
	if schluessel == "service_root" {
		return "einen anderen service_root"
	}
	return "eine andere odata_version"
}

func konfigZeile(schluessel, grund string) string {
	return fmt.Sprintf("%s %s: %s", model.CodeE2001, schluessel, grund)
}

func konfigFehler(schluessel, grund string) error {
	return &KonfigFehler{Zeilen: []string{konfigZeile(schluessel, grund)}}
}

func (h *Handle) oeffneErsteSessions(ctx context.Context) error {
	h.mu.Lock()
	h.sessionBegonnen = true
	h.mu.Unlock()
	for _, d := range h.k.Dienste {
		if _, err := h.sessions.OpenInitialSession(ctx, d.ID); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handle) starteListener(ctx context.Context) error {
	l, err := h.b.Listen(ctx, "tcp", h.k.Listen)
	if err != nil {
		if kontextFehler(err) {
			return err
		}
		return konfigFehler("LISTEN", "Adresse nicht bindbar")
	}
	srv, err := h.b.NeuerServer(httpodata.Konfig{Verarbeiten: h.verarbeiten, IDs: h.ids, Listener: l})
	if err != nil {
		_ = l.Close()
		return err
	}
	ende := make(chan struct{})
	h.mu.Lock()
	h.server, h.adresse, h.serveEnd = srv, l.Addr().String(), ende
	h.mu.Unlock()
	go func() {
		defer close(ende)
		defer close(h.fehler)
		if err := srv.Serve(); err != nil {
			h.fehler <- err
		}
	}()
	return nil
}

func kontextFehler(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// Herunterfahren bringt zurück, was Hochfahren geöffnet hat: Anhalten, CloseAll (record, nach einer
// begonnenen ersten Session), Schliessen mit dem Rest der Frist, Warten auf Serve, Ablage schließen
// (SPEC-139 Nr. 7). Es läuft einmal; jeder weitere und gleichzeitige Aufruf liefert dasselbe Ergebnis.
func (h *Handle) Herunterfahren(ctx context.Context) error {
	h.einmal.Do(func() { h.ergebnis = h.faheHerunter(context.WithoutCancel(ctx)) })
	return h.ergebnis
}

func (h *Handle) faheHerunter(ctx context.Context) error {
	h.mu.Lock()
	srv, ende, begonnen := h.server, h.serveEnd, h.sessionBegonnen
	h.zustand = beendet
	h.mu.Unlock()
	frist := h.k.ShutdownTimeout
	var fehler []error
	var angehalten time.Time
	if srv != nil {
		srv.Anhalten()
		angehalten = h.uhr.Now()
	}
	rest := frist
	if h.k.Modus == modusRecord && begonnen {
		bericht, err := h.sessions.CloseAll(ctx)
		if err != nil {
			fehler = append(fehler, err)
		}
		if bericht.Fehlgeschlagen > 0 {
			fehler = append(fehler, errors.New("Sessions nicht beendet"))
		}
		if srv != nil {
			rest = max(frist-h.uhr.Now().Sub(angehalten), 0)
		}
	}
	if srv != nil {
		if err := srv.Schliessen(rest); err != nil {
			fehler = append(fehler, err)
		}
		<-ende
	}
	if err := h.ablage.Close(); err != nil {
		fehler = append(fehler, err)
	}
	return errors.Join(fehler...)
}
