package bootstrap

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"slices"
	"strings"
	"time"
)

// Die fünf festen Gründe von ladeTLS; der sechste, Rang 1, trägt den Namen des anderen Schlüssels (SPEC-291 Nr. 2).
const (
	grundNichtLesbar    = "Datei nicht lesbar"
	grundKeinZertifikat = "Datei enthält kein PEM-Zertifikat"
	grundKeinSchluessel = "Datei enthält keinen PEM-Schlüssel"
	grundAusserhalb     = "Zertifikat außerhalb seiner Gültigkeit"
	grundPasstNicht     = "Schlüssel ungültig oder passt nicht zum Zertifikat"
)

// tlsDateien nennt die zwei Dateien eines TLS-Paares und die Namen, unter denen der Aufrufer sie meldet (SPEC-291 Nr. 2).
type tlsDateien struct {
	certSchluessel, keySchluessel, certPfad, keyPfad string
}

// tlsSeite ist der Stand der Prüfung einer der beiden Dateien; grund ist leer, solange sie bestanden hat.
type tlsSeite struct {
	name, anderer, pfad string
	roh                 []byte
	grund               string
}

// ladeTLS lädt und prüft Zertifikat und Schlüssel nach SPEC-291 Nr. 2 bis 5 und liefert das Paar als *tls.Config oder die
// Verstöße, nach Schlüssel (Bytefolge) sortiert, höchstens einen je Schlüssel. Sind beide Pfade leer, liefert es (nil, nil)
// und liest nichts. Es liest keine Uhr, keinen Kontext und keinen Zustand und schreibt nichts (SPEC-291 Nr. 12).
func ladeTLS(d tlsDateien, jetzt time.Time) (*tls.Config, []verstoss) {
	if d.certPfad == "" && d.keyPfad == "" {
		return nil, nil
	}
	zert := &tlsSeite{name: d.certSchluessel, anderer: d.keySchluessel, pfad: d.certPfad}
	schl := &tlsSeite{name: d.keySchluessel, anderer: d.certSchluessel, pfad: d.keyPfad}
	blatt := zert.liesBlatt()
	schl.lies(istSchluesselTyp, grundKeinSchluessel)
	// Der Rang 4 des Schlüssels läuft, wenn beide den Rang 3 bestanden haben, auch bei einem Verstoß der Gültigkeit.
	beideLesbar := blatt != nil && schl.grund == ""
	if blatt != nil && (jetzt.Before(blatt.NotBefore) || jetzt.After(blatt.NotAfter)) {
		zert.grund = grundAusserhalb
	}
	var paar tls.Certificate
	if beideLesbar {
		var err error
		if paar, err = tls.X509KeyPair(zert.roh, schl.roh); err != nil {
			schl.grund = grundPasstNicht
		}
	}
	var verstoesse []verstoss
	for _, s := range []*tlsSeite{zert, schl} {
		if s.grund != "" {
			verstoesse = append(verstoesse, verstoss{s.name, s.grund})
		}
	}
	if verstoesse != nil {
		slices.SortStableFunc(verstoesse, func(a, b verstoss) int { return strings.Compare(a.schluessel, b.schluessel) })
		return nil, verstoesse
	}
	return &tls.Config{Certificates: []tls.Certificate{paar}, MinVersion: tls.VersionTLS12, NextProtos: []string{"http/1.1"}}, nil
}

// liesBlatt prüft die Ränge 1 bis 3 des Zertifikats und liefert das Blatt, den ersten Block `CERTIFICATE`; liefert es nil,
// steht der Grund in s.grund.
func (s *tlsSeite) liesBlatt() *x509.Certificate {
	b := s.lies(func(typ string) bool { return typ == "CERTIFICATE" }, grundKeinZertifikat)
	if b == nil {
		return nil
	}
	blatt, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		s.grund = grundKeinZertifikat
		return nil
	}
	return blatt
}

// lies prüft die Ränge 1 bis 3: leerer Pfad, Datei nicht lesbar, und kein Block, dessen Typ passt annimmt (Grund: kein). Es
// liefert den ersten Block, dessen Typ passt annimmt, oder nil und hält den Inhalt in s.roh.
func (s *tlsSeite) lies(passt func(typ string) bool, kein string) *pem.Block {
	if s.pfad == "" {
		s.grund = "erforderlich, wenn " + s.anderer + " gesetzt ist"
		return nil
	}
	var err error
	if s.roh, err = os.ReadFile(s.pfad); err != nil {
		s.grund = grundNichtLesbar
		return nil
	}
	rest := s.roh
	for {
		var b *pem.Block
		if b, rest = pem.Decode(rest); b == nil {
			s.grund = kein
			return nil
		}
		if passt(b.Type) {
			return b
		}
	}
}

// istSchluesselTyp nimmt den Typ `PRIVATE KEY` und jeden Typ an, der auf ` PRIVATE KEY` endet (SPEC-291 Nr. 4).
func istSchluesselTyp(typ string) bool {
	return typ == "PRIVATE KEY" || strings.HasSuffix(typ, " PRIVATE KEY")
}
