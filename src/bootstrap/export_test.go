package bootstrap

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// EtagFuerSatz ist die Entscheidung des Composition Root, ob ein Dienst mit diesem Satz ETags liefert (SPEC-280 Nr. 2).
func EtagFuerSatz(satz outbound.CodecSatz, an bool) bool { return etagFuerSatz(satz, an) }

// Leerraum sind die Zeichen, die die Konfiguration am Rand entfernt (SPEC-110 Nr. 1).
const Leerraum = leerraum

// Kappe entfernt Leerraum an beiden Rändern, wie die Konfiguration es an jeder Stelle tut (SPEC-110 Nr. 1).
func Kappe(s string) string { return kappe(s) }

// IstQueryName ist die Prüfung eines Eintrags von REDACT_QUERY_PARAMS (SPEC-110 Nr. 6).
func IstQueryName(s string) bool { return istQueryName(s) }

// TLSVerstoss ist ein Verstoß von ladeTLS: der Schlüssel, unter dem der Aufrufer meldet, und der Grund (SPEC-291 Nr. 2).
type TLSVerstoss struct{ Schluessel, Grund string }

// LadeTLS ist ladeTLS mit den Namen und Pfaden des Aufrufers (SPEC-291 Nr. 2); es kennt keinen Namen von sich aus.
func LadeTLS(certSchluessel, keySchluessel, certPfad, keyPfad string, jetzt time.Time) (*tls.Config, []TLSVerstoss) {
	cfg, vs := ladeTLS(tlsDateien{certSchluessel, keySchluessel, certPfad, keyPfad}, jetzt)
	var aus []TLSVerstoss
	for _, v := range vs {
		aus = append(aus, TLSVerstoss{v.schluessel, v.grund})
	}
	return cfg, aus
}

// Die Naht der Funktionen des Admin-Listeners (SPEC-299, admintls.go).

// AdminEingabe, AdminTransport und AdminWeg sind die Typen von pruefeAdminTransport.
type (
	AdminEingabe   = adminEingabe
	AdminTransport = adminTransport
	AdminWeg       = adminWeg
)

// Die vier Wege des Admin-Listeners (SPEC-299 Nr. 3).
const (
	WegKeiner   = wegKeiner
	WegTLS      = wegTLS
	WegLoopback = wegLoopback
	WegKlartext = wegKlartext
)

// AdminLoopback ist adminLoopback (SPEC-299 Nr. 2).
func AdminLoopback(listen string) bool { return adminLoopback(listen) }

// PruefeAdminTransport ist pruefeAdminTransport mit den Verstößen als TLSVerstoss (SPEC-299 Nr. 3, 4, 6).
func PruefeAdminTransport(e AdminEingabe, jetzt time.Time) (AdminTransport, []TLSVerstoss) {
	t, vs := pruefeAdminTransport(e, jetzt)
	var aus []TLSVerstoss
	for _, v := range vs {
		aus = append(aus, TLSVerstoss{v.schluessel, v.grund})
	}
	return t, aus
}

// AdminUmhuelle ist umhuelle von t (SPEC-299 Nr. 8 c).
func AdminUmhuelle(t AdminTransport, l net.Listener) net.Listener { return t.umhuelle(l) }

// AdminWarne ist warne von t (SPEC-299 Nr. 9).
func AdminWarne(t AdminTransport, log *slog.Logger) { t.warne(log) }

// LadeKonfigurationZu ist LadeKonfiguration zu einer festen Zeit des Starts (SPEC-291 Nr. 3).
func LadeKonfigurationZu(environ []string, jetzt time.Time) (Konfiguration, error) {
	return ladeKonfigurationZu(environ, jetzt)
}

// Die Naht der Aufrufstellen der Prüfung des Admin-Transports am Laden und in Build (SPEC-299 Nr. 8 a, b).

// AdminPruefung ist die Form der Prüfung mit den Verstößen als TLSVerstoss; PruefeAdminTransport hat sie.
type AdminPruefung = func(AdminEingabe, time.Time) (AdminTransport, []TLSVerstoss)

func intern(p AdminPruefung) adminPruefung {
	return func(e adminEingabe, jetzt time.Time) (adminTransport, []verstoss) {
		t, vs := p(e, jetzt)
		var aus []verstoss
		for _, v := range vs {
			aus = append(aus, verstoss{v.Schluessel, v.Grund})
		}
		return t, aus
	}
}

// LadeKonfigurationMitPruefung ist LadeKonfigurationZu mit p an der Stelle Laden.
func LadeKonfigurationMitPruefung(environ []string, jetzt time.Time, p AdminPruefung) (Konfiguration, error) {
	return ladeKonfigurationMit(environ, jetzt, intern(p))
}

// SetzePruefung setzt p an die Stelle Build von b.
func SetzePruefung(b *Bauteile, p AdminPruefung) { b.PruefeAdmin = intern(p) }

// AdminTransportVon liefert den Transport des Admin-Listeners, den Build im Handle ablegt (SPEC-299 Nr. 8 b).
func AdminTransportVon(h *Handle) AdminTransport { return h.transport }

// ListenTLSPraefix ist der Anfang der Schlüssel `LISTEN_TLS_*`, den die Konfiguration aus den zwei Namen bildet.
func ListenTLSPraefix() string { return listenTLSPraefix() }

// SatzVersion4 ist der Satz der Codecs der Version 4.0 des Mock.
func SatzVersion4() outbound.CodecSatz { return satzVersion4() }

// SatzTabelle ist die Tabelle Version → Satz des Root (SPEC-139 Nr. 11 b).
func SatzTabelle() map[string]outbound.CodecSatz { return satzTabelle() }

// Die Naht der Tests: Das Testpaket bootstrap_test setzt die Haken von Bauteile und ruft Start und
// Build mit ihnen; der Betrieb nutzt standardBauteile.

// Bauteile ist die Naht der Verdrahtung (SPEC-139).
type Bauteile = bauteile

// Server ist, was der Root vom Listener-Adapter ruft.
type Server = server

// StandardBauteile liefert die Konstruktoren des Betriebs.
func StandardBauteile() Bauteile { return standardBauteile() }

// StartMit ist Start mit gegebener Umgebung und gegebenen Bauteilen.
func StartMit(ctx context.Context, environ []string, stderr io.Writer, b Bauteile) int {
	return start(ctx, environ, stderr, b)
}

// BuildMit ist Build mit gegebenen Bauteilen.
func BuildMit(ctx context.Context, k Konfiguration, log *slog.Logger, b Bauteile) (*Handle, error) {
	return build(ctx, k, log, io.Discard, b)
}

// AdminRouten liefert die Routen des Admin-Adapters, die Build im Handle ablegt (SPEC-300 Nr. 13 a).
func AdminRouten(h *Handle) []admin.Route { return h.routen }

// TelemetrieVon liefert den Adapter des Prozesses (SPEC-139 Nr. 13).
func TelemetrieVon(h *Handle) *telemetry.Adapter { return h.tel }

// BuildMitSenke ist Build mit gegebenen Bauteilen und der Senke des Adapters des Prozesses (SPEC-139 Nr. 13).
func BuildMitSenke(ctx context.Context, k Konfiguration, log *slog.Logger, senke io.Writer, b Bauteile) (*Handle, error) {
	return build(ctx, k, log, senke, b)
}

// IDsVon liefert den IDGenerator des Handle.
func IDsVon(h *Handle) support.IDGenerator { return h.ids }

// Warte ist das Warten von Start auf den Abbruch von ctx oder das Ende von Serve.
func Warte(ctx context.Context, fehler <-chan error) error { return warte(ctx, fehler) }

// ErrServeBeendet ist der Ausfall, wenn Serve ohne Anhalten und ohne Abbruch von ctx zurückkehrt.
var ErrServeBeendet = errServeBeendet

// Meldung ist die Zeile, die Start zu einem Fehler schreibt.
func Meldung(err error) string { return meldung(err) }

// LiteralHandle liefert den Handle eines hochgefahrenen Prozesses aus Fakes; bediene ruft Serve jedes Servers wie Hochfahren,
// und Herunterfahren wartet auf ihre Rückkehr.
func LiteralHandle(modus string, frist time.Duration, uhr support.Clock, sessions inbound.SessionUseCase, srv Server, ablage io.Closer) *Handle {
	return LiteralHandleZwei(modus, frist, uhr, sessions, srv, nil, ablage)
}

// LiteralHandleZwei ist LiteralHandle mit dem Admin-Server adm (nil ohne).
func LiteralHandleZwei(modus string, frist time.Duration, uhr support.Clock, sessions inbound.SessionUseCase, srv, adm Server, ablage io.Closer) *Handle {
	h := &Handle{
		k: Konfiguration{Modus: modus, ShutdownTimeout: frist}, uhr: uhr, sessions: sessions, ablage: ablage,
		fehler: neuerFehlerKanal(), sessionBegonnen: true, zustand: hochgefahren,
	}
	h.bediene(srv, adm)
	return h
}

// AuditVon liefert das AuditLog des Prozesses.
func AuditVon(h *Handle) support.AuditLog { return h.audit }

// AusfallDanach ist das Nachlesen von Start nach dem Abbruch von ctx.
func AusfallDanach(ctx context.Context, fehler <-chan error) error {
	return ausfallDanach(ctx, fehler)
}
