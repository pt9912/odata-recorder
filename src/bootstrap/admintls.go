package bootstrap

import (
	"crypto/tls"
	"log/slog"
	"net"
	"strings"
	"time"
)

// Die Namen der Schlüssel, unter denen die Funktionen dieser Datei melden (SPEC-299 Nr. 1, 4, 6); sie stehen hier als Konstanten, nicht in konfig*.go.
const (
	adminTLSCert = "ADMIN_TLS_CERT"
	adminTLSKey  = "ADMIN_TLS_KEY"
	adminListen  = "ADMIN_LISTEN"
)

// grundAdminZusammenhang ist der feste Grund des Verstoßes am Schlüssel ADMIN_LISTEN, ohne dessen Wert (SPEC-299 Nr. 6).
const grundAdminZusammenhang = "außerhalb von Loopback verlangt TLS: " + adminTLSCert + " und " + adminTLSKey

// adminWeg ist der Weg des Admin-Listeners (SPEC-299 Nr. 3); der Nullwert ist kein Weg.
type adminWeg int

// Die Wege: kein Weg (Verstoß oder ungültige Form), HTTPS, Klartext auf Loopback und Klartext mit Flag.
const (
	wegKeiner adminWeg = iota
	wegTLS
	wegLoopback
	wegKlartext
)

// adminEingabe trägt die Werte der Konfiguration, aus denen pruefeAdminTransport den Weg bestimmt (SPEC-299 Nr. 3 bis 6).
type adminEingabe struct {
	Listen, CertPfad, KeyPfad string
	Klartext, FormGueltig     bool
}

// adminTransport ist der Weg des Listeners und, bei wegTLS, die Konfiguration seiner Verbindungen (SPEC-299 Nr. 3, 7).
type adminTransport struct {
	TLS *tls.Config
	Weg adminWeg
}

// adminLoopback meldet, ob der Host von listen, der Text vor dem letzten `:`, bytegleich `127.0.0.1`, `[::1]` oder `localhost` lautet;
// ohne `:` gibt es keinen Host (SPEC-299 Nr. 2). Es löst keinen Namen auf.
func adminLoopback(listen string) bool {
	i := strings.LastIndex(listen, ":")
	if i < 0 {
		return false
	}
	switch listen[:i] {
	case "127.0.0.1", "[::1]", "localhost":
		return true
	}
	return false
}

// pruefeAdminTransport lädt das Paar mit ladeTLS unter den Namen ADMIN_TLS_* und bestimmt den Weg nach SPEC-299 Nr. 3, 4 und 6. Ein Paar mit
// Verstoß liefert keinen Weg; bei ungültiger Form entfällt der Zusammenhang, die Dateien eines gesetzten Paares werden dennoch geprüft.
// Es liest keine Uhr, keinen Kontext und keinen Zustand und schreibt nichts (SPEC-299 Nr. 13).
func pruefeAdminTransport(e adminEingabe, jetzt time.Time) (adminTransport, []verstoss) {
	cfg, vs := ladeTLS(tlsDateien{adminTLSCert, adminTLSKey, e.CertPfad, e.KeyPfad}, jetzt)
	switch {
	case vs != nil || !e.FormGueltig:
		return adminTransport{}, vs
	case cfg != nil:
		return adminTransport{TLS: cfg, Weg: wegTLS}, nil
	case adminLoopback(e.Listen):
		return adminTransport{Weg: wegLoopback}, nil
	case e.Klartext:
		return adminTransport{Weg: wegKlartext}, nil
	}
	return adminTransport{}, []verstoss{{adminListen, grundAdminZusammenhang}}
}

// umhuelle liefert bei wegTLS den Listener mit TLS, sonst l selbst (SPEC-299 Nr. 8 c); es ruft keine Operation von l.
func (t adminTransport) umhuelle(l net.Listener) net.Listener {
	if t.Weg == wegTLS {
		return tls.NewListener(l, t.TLS)
	}
	return l
}

// warne schreibt bei wegKlartext genau einen Eintrag der Stufe WARN mit festem Text und dem einen Attribut `ereignis`; bei jedem anderen Weg und
// bei log nil schreibt es nichts (SPEC-299 Nr. 9).
func (t adminTransport) warne(log *slog.Logger) {
	if t.Weg == wegKlartext && log != nil {
		log.Warn("start: Admin-Listener ohne TLS erreichbar", "ereignis", "admin_klartext")
	}
}
