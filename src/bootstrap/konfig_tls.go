package bootstrap

import (
	"crypto/tls"
	"strings"
	"time"
)

// Die zwei Schlüssel des TLS des OData-Listeners (SPEC-110 Nr. 2, SPEC-291 Nr. 7).
const (
	schluesselListenCert = "LISTEN_TLS_CERT"
	schluesselListenKey  = "LISTEN_TLS_KEY"
)

// listenTLSPraefix ist der Anfang von schluesselListenCert bis einschließlich dem letzten `_`; jeder andere Schlüssel mit diesem
// Anfang ist abgelehnt (SPEC-110 Nr. 2). Es steht nie als Literal in konfig*.go, denn dort gilt jedes Literal der Form eines
// Schlüssels als Schlüssel (SPEC-201 Nr. 4).
func listenTLSPraefix() string {
	return schluesselListenCert[:strings.LastIndexByte(schluesselListenCert, '_')+1]
}

// mitListenTLS heißt: beide Pfade sind gesetzt, ob ihre Dateien gültig sind oder nicht (SPEC-291 Nr. 7 d).
func (k *Konfiguration) mitListenTLS() bool { return k.ListenTLSCert != "" && k.ListenTLSKey != "" }

// listenTLSDateien nennt dem Baustein die zwei Pfade der Konfiguration und die Namen, unter denen er sie meldet (SPEC-291 Nr. 7 c).
func (k *Konfiguration) listenTLSDateien() tlsDateien {
	return tlsDateien{
		certSchluessel: schluesselListenCert, keySchluessel: schluesselListenKey,
		certPfad: k.ListenTLSCert, keyPfad: k.ListenTLSKey,
	}
}

// liesListenTLS liest die Pfade von LISTEN_TLS_CERT und LISTEN_TLS_KEY in jedem Modus (SPEC-291 Nr. 7 a) und lehnt jeden anderen
// gesetzten Schlüssel mit dem Präfix ab (SPEC-110 Nr. 2); die Dateien prüft pruefeListenTLS.
func (p *pruefer) liesListenTLS(k *Konfiguration) {
	k.ListenTLSCert, _ = p.wert(schluesselListenCert)
	k.ListenTLSKey, _ = p.wert(schluesselListenKey)
	for name, v := range p.env {
		if strings.HasPrefix(name, listenTLSPraefix()) && name != schluesselListenCert && name != schluesselListenKey && kappe(v) != "" {
			p.melde(name, grundNichtUnterstuetzt)
		}
	}
}

// pruefeListenTLS lädt das Paar zur Zeit jetzt und meldet jeden Verstoß unter dem Namen seines Schlüssels (SPEC-291 Nr. 7 b, c);
// ohne einen der beiden Pfade meldet der Baustein den Schlüssel ohne Pfad und prüft den anderen weiter (SPEC-291 Nr. 2).
func (p *pruefer) pruefeListenTLS(k *Konfiguration, jetzt time.Time) {
	_, verstoesse := ladeTLS(k.listenTLSDateien(), jetzt)
	for _, v := range verstoesse {
		p.melde(v.schluessel, v.grund)
	}
}

// ladeListenTLS lädt das Paar von Build (SPEC-291 Nr. 8): nil ohne jeden Pfad, sonst die Konfiguration der Verbindungen
// oder ein *KonfigFehler mit einer Zeile je Verstoß, nach Schlüssel sortiert; bei genau einem Pfad ist es der Fehler.
func ladeListenTLS(k Konfiguration, jetzt time.Time) (*tls.Config, error) {
	cfg, verstoesse := ladeTLS(k.listenTLSDateien(), jetzt)
	if len(verstoesse) == 0 {
		return cfg, nil
	}
	zeilen := make([]string, len(verstoesse))
	for i, v := range verstoesse {
		zeilen[i] = konfigZeile(v.schluessel, v.grund)
	}
	return nil, &KonfigFehler{Zeilen: zeilen}
}
