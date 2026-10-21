package httpupstream

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

// tlsTransport setzt TLS und Protokoll des Transports (SPEC-120 Nr. 1, 7): bei https prüft er Kette
// und Hostnamen gegen die Anker des Systems und CAPEM, bei Pruefen=false nichts; er spricht nur
// HTTP/1.1, auch über TLS. CAPEM wird auch bei Pruefen=false gelesen.
func tlsTransport(tr *http.Transport, pruefen bool, capem []byte) error {
	anker, err := leseAnker(capem)
	if err != nil {
		return err
	}
	var nurHTTP1 http.Protocols
	nurHTTP1.SetHTTP1(true)
	tr.Protocols = &nurHTTP1
	tr.TLSClientConfig = &tls.Config{RootCAs: anker, InsecureSkipVerify: !pruefen}
	return nil
}

// leseAnker liest CAPEM: leer heißt die Anker des Systems allein (nil); sonst die Anker des Systems
// und jedes lesbare Zertifikat aus CAPEM, mindestens eines, nach der Regel von x509, die auch der Start
// an UPSTREAM_CA_FILE anwendet. Grenze: liefert das System keinen Bestand, beginnt der Pool leer
// (SPEC-120 Ungeprüft); kein Test fängt diesen Zweig.
func leseAnker(capem []byte) (*x509.CertPool, error) {
	if len(capem) == 0 {
		return nil, nil
	}
	anker, err := x509.SystemCertPool()
	if err != nil {
		anker = x509.NewCertPool()
	}
	if !anker.AppendCertsFromPEM(capem) {
		return nil, ungueltig("CAPEM")
	}
	return anker, nil
}
