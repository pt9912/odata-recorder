package bootstrap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Hochfahren führt, mit Admin-Listener, die Token-Datei, dann Wartung, Services anlegen, erste Sessions (nur record) und die
// Listener in dieser Reihenfolge aus (SPEC-139 Nr. 6, SPEC-300 Nr. 4, 7). Vor jedem Schritt prüft es ctx; ein abgebrochener
// Kontext und jeder Fehler eines Schritts beenden es an der Stelle.
func (h *Handle) Hochfahren(ctx context.Context) error {
	h.mu.Lock()
	if h.zustand != gebaut {
		h.mu.Unlock()
		return ungueltig("Hochfahren läuft einmal je Handle")
	}
	h.zustand = hochgefahren
	h.mu.Unlock()
	var schritte []func(context.Context) error
	if h.k.AdminListen != "" {
		schritte = append(schritte, h.ladeToken)
	}
	schritte = append(schritte, h.startWartung, h.legeDiensteAn)
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

// ladeToken ist der erste Schritt mit Admin-Listener (SPEC-300 Nr. 4); er läuft ohne Kontext (SPEC-239).
func (h *Handle) ladeToken(context.Context) error {
	token, err := ladeAdminToken(h.k.DataDir, h.k.AdminTokenFile)
	h.token = token
	return err
}

func (h *Handle) startWartung(ctx context.Context) error {
	_, err := h.wartung.Startup(ctx)
	return err
}

// legeDiensteAn legt jeden Service mit der Version seines Dienstes an, den die Ablage nicht kennt; ein gespeicherter mit anderem
// service_root oder anderer odata_version als die des Dienstes verhindert den Start (SPEC-107 Nr. 14).
func (h *Handle) legeDiensteAn(ctx context.Context) error {
	for _, d := range h.k.Dienste {
		err := h.dienste.Create(ctx, model.Service{ID: d.ID, ServiceRoot: d.Root, ODataVersion: d.Version})
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

// abweichungen nennt die Schlüssel, in denen der gespeicherte Service von dem des Dienstes abweicht,
// service_root vor odata_version; die Version geht gegen die des Dienstes (SPEC-139 Nr. 6).
func abweichungen(gespeichert model.Service, d Dienst) []string {
	var schluessel []string
	if gespeichert.ServiceRoot != d.Root {
		schluessel = append(schluessel, "service_root")
	}
	if gespeichert.ODataVersion != d.Version {
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

// starteListener bindet LISTEN und, mit Admin-Listener, ADMIN_LISTEN, umhüllt beide, baut beide Server und bedient sie erst
// danach (SPEC-300 Nr. 7); den Admin-Listener umhüllt allein der Weg TLS, den Build lud (SPEC-299 Nr. 8 c). Ein Fehler der
// Bindung oder eines Konstruktors schließt jeden gebundenen Listener; kein Server steht dann im Handle, und beide Adressen
// bleiben leer. Zwischen Bindung und Serve prüft es ctx nicht.
func (h *Handle) starteListener(ctx context.Context) (err error) {
	l, al, err := h.binde(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			schliesse(l, al)
		}
	}()
	if h.tls != nil {
		l = tls.NewListener(l, h.tls)
	}
	// Ohne Admin-Listener ist al nil und der Weg wegKeiner: umhuelle gibt nil zurück.
	al = h.transport.umhuelle(al)
	srv, err := h.b.NeuerServer(httpodata.Konfig{
		Verarbeiten: h.verarbeiten, IDs: h.ids, Listener: l, Telemetrie: h.tel, Modus: h.k.Modus,
	})
	if err != nil {
		return err
	}
	var adm server
	adresse := ""
	if al != nil {
		adm, err = h.b.NeuerAdmin(admin.Konfig{
			Listener: al, Token: h.token, IDs: h.ids, Audit: h.audit, Telemetrie: h.tel, Routen: h.routen, Uhr: h.uhr,
		})
		if err != nil {
			return err
		}
		adresse = al.Addr().String()
	}
	h.mu.Lock()
	h.adresse, h.adminAdresse = l.Addr().String(), adresse
	h.mu.Unlock()
	h.bediene(srv, adm)
	return nil
}

// binde öffnet LISTEN und, mit Admin-Listener, ADMIN_LISTEN. Ein Kontextfehler kommt unverändert zurück; sonst ist jeder
// gescheiterte Bind eine Zeile am Schlüssel, beide nach Schlüssel sortiert. Bei einem Fehler ist nichts offen.
func (h *Handle) binde(ctx context.Context) (net.Listener, net.Listener, error) {
	adressen := []struct{ schluessel, adresse string }{{"LISTEN", h.k.Listen}}
	if h.k.AdminListen != "" {
		adressen = append(adressen, struct{ schluessel, adresse string }{schluesselAdminListen, h.k.AdminListen})
	}
	offen := make([]net.Listener, 2)
	p := &pruefer{}
	for i, a := range adressen {
		l, err := h.b.Listen(ctx, "tcp", a.adresse)
		if kontextFehler(err) {
			schliesse(offen...)
			return nil, nil, err
		}
		if err != nil {
			p.melde(a.schluessel, "Adresse nicht bindbar")
			continue
		}
		offen[i] = l
	}
	if err := p.fehler(); err != nil {
		schliesse(offen...)
		return nil, nil, err
	}
	return offen[0], offen[1], nil
}

func schliesse(listener ...net.Listener) {
	for _, l := range listener {
		if l != nil {
			_ = l.Close()
		}
	}
}

// bediene legt beide Server im Handle ab und ruft Serve jedes Servers in einer eigenen Goroutine; adm ist nil ohne
// Admin-Listener. Der erste Ausfall geht einmal in fehler: ein Fehler von Serve oder eine Rückkehr mit nil, bevor
// Herunterfahren begonnen hat (errServeBeendet). fehler schließt erst, wenn jedes Serve zurückgekehrt ist (SPEC-300 Nr. 9).
func (h *Handle) bediene(srv, adm server) {
	server := []interface{ Serve() error }{srv}
	if adm != nil {
		server = append(server, adm)
	}
	ende := make(chan struct{})
	h.mu.Lock()
	h.server, h.adminServer, h.serveEnd = srv, adm, ende
	h.mu.Unlock()
	var laufend sync.WaitGroup
	var einmal sync.Once
	for _, s := range server {
		laufend.Go(func() {
			err := s.Serve()
			if err == nil {
				if h.heruntergefahren() {
					return
				}
				err = errServeBeendet
			}
			einmal.Do(func() { h.fehler <- err })
		})
	}
	go func() {
		laufend.Wait()
		close(h.fehler)
		close(ende)
	}()
}

// neuerFehlerKanal ist der Kanal von Fehler: er fasst den einen Wert, damit kein Serve darauf blockiert.
func neuerFehlerKanal() chan error { return make(chan error, 1) }

// heruntergefahren meldet, ob Herunterfahren begonnen hat.
func (h *Handle) heruntergefahren() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.zustand == beendet
}

func kontextFehler(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// Herunterfahren bringt zurück, was Hochfahren geöffnet hat: Anhalten des OData-Servers, unmittelbar danach des Admin-Servers,
// CloseAll (record, nach einer begonnenen ersten Session), Schliessen des OData-Servers und danach des Admin-Servers je mit dem
// Rest der Frist, Warten auf beide Serve, Ablage schließen (SPEC-139 Nr. 7, SPEC-300 Nr. 8). Es läuft einmal; jeder weitere und
// gleichzeitige Aufruf liefert dasselbe Ergebnis.
func (h *Handle) Herunterfahren(ctx context.Context) error {
	h.einmal.Do(func() { h.ergebnis = h.faheHerunter(context.WithoutCancel(ctx)) })
	return h.ergebnis
}

func (h *Handle) faheHerunter(ctx context.Context) error {
	h.mu.Lock()
	srv, adm, ende, begonnen := h.server, h.adminServer, h.serveEnd, h.sessionBegonnen
	h.zustand = beendet
	h.mu.Unlock()
	frist := h.k.ShutdownTimeout
	var fehler []error
	var angehalten time.Time
	if srv != nil {
		srv.Anhalten()
		angehalten = h.uhr.Now()
	}
	if adm != nil {
		adm.Anhalten()
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
		if adm != nil {
			if err := adm.Schliessen(max(frist-h.uhr.Now().Sub(angehalten), 0)); err != nil {
				fehler = append(fehler, err)
			}
		}
		<-ende
	}
	if err := h.ablage.Close(); err != nil {
		fehler = append(fehler, err)
	}
	return errors.Join(fehler...)
}
