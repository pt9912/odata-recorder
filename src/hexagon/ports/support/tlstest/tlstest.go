// Package tlstest erzeugt Aussteller, Blätter und PEM-Dateien für Tests von TLS (SPEC-291) zur Laufzeit:
// nur die Standardbibliothek, kein Produktcode, kein Geheimnis im Repo (LH-RB-04). Ein Fehler beim Erzeugen
// beendet den Test mit t.Fatalf.
package tlstest

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Art ist die Art des Schlüssels eines Ausstellers oder Blatts.
type Art int

// Die Arten, die die Tests brauchen; ECDSA (P-256) ist der Nullwert.
const (
	ECDSA Art = iota
	RSA
	Ed25519
)

// Form ist die Form, in der ein Schlüssel als PEM-Block steht.
type Form int

// Die Formen: PKCS8 (Block `PRIVATE KEY`, jede Art), PKCS1 (Block `RSA PRIVATE KEY`, nur RSA) und
// SEC1 (Block `EC PRIVATE KEY`, nur ECDSA); PKCS8 ist der Nullwert.
const (
	PKCS8 Form = iota
	PKCS1
	SEC1
)

// Optionen sind die Eigenschaften eines Zertifikats; der Nullwert gilt eine Stunde vor und nach dem Aufruf
// (das Zertifikat trägt Sekunden), heißt `localhost` und trägt einen neuen Schlüssel der Art ECDSA.
type Optionen struct {
	// Von und Bis sind NotBefore und NotAfter; null heißt: eine Stunde vor bzw. nach jetzt.
	Von, Bis time.Time
	// Name ist der Common Name und, beim Blatt, der einzige DNS-Name; leer heißt `localhost`.
	Name string
	// Art ist die Art des neuen Schlüssels.
	Art Art
}

// Aussteller ist eine Wurzel oder ein Zwischenaussteller mit dem Schlüssel, der Blätter und Zwischenaussteller
// unterschreibt.
type Aussteller struct {
	Zertifikat *x509.Certificate
	Schluessel crypto.Signer

	wurzel *x509.Certificate
	kette  [][]byte
}

// Blatt ist ein Zertifikat mit seinem Schlüssel; Kette sind die DER-Bytes der Zwischenzertifikate hinter dem
// Blatt, die Wurzel gehört nicht dazu.
type Blatt struct {
	Zertifikat *x509.Certificate
	Schluessel crypto.Signer
	Kette      [][]byte
}

// Ergebnis ist der Ausgang eines Handshakes im Prozess; Version und Protokoll gelten für den Client.
type Ergebnis struct {
	ServerFehler, ClientFehler error
	Version                    uint16
	Protokoll                  string
}

func ok(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("tlstest: %v", err)
	}
}

func muss[T any](t testing.TB, wert T, err error) T {
	t.Helper()
	ok(t, err)
	return wert
}

func neuerSchluessel(t testing.TB, art Art) crypto.Signer {
	t.Helper()
	switch art {
	case RSA:
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		return muss(t, k, err)
	case Ed25519:
		_, k, err := ed25519.GenerateKey(rand.Reader)
		return muss(t, k, err)
	default:
		k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		return muss(t, k, err)
	}
}

// vorlage bildet das Zertifikat nach o; ein Aussteller ist eine CA, ein Blatt trägt seinen DNS-Namen.
func (o Optionen) vorlage(t testing.TB, aussteller bool, standard string) *x509.Certificate {
	t.Helper()
	von, bis, name := o.Von, o.Bis, o.Name
	if von.IsZero() {
		von = time.Now().Add(-time.Hour)
	}
	if bis.IsZero() {
		bis = time.Now().Add(time.Hour)
	}
	if name == "" {
		name = standard
	}
	nummer, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	v := &x509.Certificate{
		SerialNumber: muss(t, nummer, err), Subject: pkix.Name{CommonName: name}, NotBefore: von, NotAfter: bis,
		BasicConstraintsValid: true, IsCA: aussteller, KeyUsage: x509.KeyUsageDigitalSignature,
	}
	if aussteller {
		v.KeyUsage = x509.KeyUsageCertSign
	} else {
		v.DNSNames = []string{name}
		v.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}
	return v
}

// unterschreibe stellt vorlage für den Schlüssel pub aus, unterschrieben von signer unter eltern (bei einer
// Wurzel eltern gleich vorlage).
func unterschreibe(t testing.TB, vorlage, eltern *x509.Certificate, pub crypto.PublicKey, signer crypto.Signer) *x509.Certificate {
	t.Helper()
	der, err := x509.CreateCertificate(rand.Reader, vorlage, eltern, pub, signer)
	ok(t, err)
	z, err := x509.ParseCertificate(der)
	return muss(t, z, err)
}

// NeueWurzel liefert eine selbst unterschriebene Wurzel, gültig um jetzt, mit einem ECDSA-Schlüssel.
func NeueWurzel(t testing.TB) *Aussteller {
	t.Helper()
	o := Optionen{}
	k := neuerSchluessel(t, o.Art)
	v := o.vorlage(t, true, "tlstest-wurzel")
	z := unterschreibe(t, v, v, k.Public(), k)
	return &Aussteller{Zertifikat: z, Schluessel: k, wurzel: z}
}

// Zwischen liefert einen Zwischenaussteller, den a unterschreibt.
func (a *Aussteller) Zwischen(t testing.TB, o Optionen) *Aussteller {
	t.Helper()
	k := neuerSchluessel(t, o.Art)
	z := unterschreibe(t, o.vorlage(t, true, "tlstest-zwischen"), a.Zertifikat, k.Public(), a.Schluessel)
	return &Aussteller{Zertifikat: z, Schluessel: k, wurzel: a.wurzel, kette: append([][]byte{z.Raw}, a.kette...)}
}

// Blatt liefert ein Blatt, das a unterschreibt; seine Kette sind die Zwischenzertifikate von a bis zur Wurzel.
func (a *Aussteller) Blatt(t testing.TB, o Optionen) *Blatt {
	t.Helper()
	k := neuerSchluessel(t, o.Art)
	z := unterschreibe(t, o.vorlage(t, false, "localhost"), a.Zertifikat, k.Public(), a.Schluessel)
	return &Blatt{Zertifikat: z, Schluessel: k, Kette: a.kette}
}

// Klient liefert die Konfiguration eines Clients, der allein der Wurzel von a vertraut und den Namen name erwartet.
func (a *Aussteller) Klient(name string) *tls.Config {
	pool := x509.NewCertPool()
	pool.AddCert(a.wurzel)
	return &tls.Config{RootCAs: pool, ServerName: name}
}

// Block liefert einen PEM-Block des Typs typ mit den Bytes der, Zeilenende LF.
func Block(typ string, der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der})
}

// ZertifikatPEM liefert den Block des Blatts allein.
func (b *Blatt) ZertifikatPEM() []byte { return Block("CERTIFICATE", b.Zertifikat.Raw) }

// KettePEM liefert die Blöcke des Blatts und, danach, seiner Zwischenzertifikate.
func (b *Blatt) KettePEM() []byte {
	aus := b.ZertifikatPEM()
	for _, der := range b.Kette {
		aus = append(aus, Block("CERTIFICATE", der)...)
	}
	return aus
}

// SchluesselPEM liefert den Block des Schlüssels des Blatts in der Form f.
func (b *Blatt) SchluesselPEM(t testing.TB, f Form) []byte { return SchluesselPEM(t, b.Schluessel, f) }

// SchluesselPEM liefert den Block des Schlüssels k in der Form f; PKCS1 setzt einen RSA-Schlüssel voraus und SEC1
// einen ECDSA-Schlüssel, sonst bricht der Aufruf ab.
func SchluesselPEM(t testing.TB, k crypto.Signer, f Form) []byte {
	t.Helper()
	switch f {
	case PKCS1:
		return Block("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(k.(*rsa.PrivateKey)))
	case SEC1:
		der, err := x509.MarshalECPrivateKey(k.(*ecdsa.PrivateKey))
		return Block("EC PRIVATE KEY", muss(t, der, err))
	default:
		der, err := x509.MarshalPKCS8PrivateKey(k)
		return Block("PRIVATE KEY", muss(t, der, err))
	}
}

// Datei schreibt inhalt in die Datei name im Verzeichnis verz (angelegt, wenn es fehlt) mit den Rechten modus,
// unabhängig von umask, und liefert den Pfad.
func Datei(t testing.TB, verz, name string, inhalt []byte, modus os.FileMode) string {
	t.Helper()
	ok(t, os.MkdirAll(verz, 0o700))
	pfad := filepath.Join(verz, name)
	ok(t, os.WriteFile(pfad, inhalt, modus))
	ok(t, os.Chmod(pfad, modus))
	return pfad
}

// Handshake führt einen Handshake zwischen server und client im Prozess über eine Verbindung auf dem Loopback
// (die Verbindung nimmt kein Listener des Produkts an) und liefert beide Fehler und das Verhandelte des Clients;
// scheitert eine Seite, schließt sie die Verbindung, damit die andere zurückkehrt. Der Kernel puffert die
// Verbindung: ein Client, der das Zertifikat vor dem Ende des Fluges des Servers ablehnt, schreibt seinen Alert,
// ohne dass der Server lesen muss.
func Handshake(t testing.TB, server, client *tls.Config) Ergebnis {
	t.Helper()
	ctx := t.Context()
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	ok(t, err)
	defer func() { _ = ln.Close() }()
	cc, err := (&net.Dialer{}).DialContext(ctx, "tcp", ln.Addr().String())
	ok(t, err)
	defer func() { _ = cc.Close() }()
	sc, err := ln.Accept()
	ok(t, err)
	defer func() { _ = sc.Close() }()
	fertig := make(chan error, 1)
	go func() {
		err := tls.Server(sc, server).HandshakeContext(ctx)
		if err != nil {
			_ = sc.Close()
		}
		fertig <- err
	}()
	c := tls.Client(cc, client)
	cerr := c.HandshakeContext(ctx)
	if cerr != nil {
		_ = cc.Close()
	}
	st := c.ConnectionState()
	return Ergebnis{ServerFehler: <-fertig, ClientFehler: cerr, Version: st.Version, Protokoll: st.NegotiatedProtocol}
}
