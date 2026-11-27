package bootstrap

import (
	"crypto/tls"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/audit"
	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// Die fünf Schlüssel des Admin-Listeners, die diese Version liest (SPEC-110 Nr. 2, SPEC-300 Nr. 2), und die Namen, unter
// denen die Token-Datei in DATA_DIR liegt. admintls.go meldet die Verstöße unter diesen Namen und führt sie nie selbst als
// Literal (SPEC-299 Nr. 1).
const (
	schluesselAdminListen = "ADMIN_LISTEN"
	schluesselAdminToken  = "ADMIN_TOKEN_FILE"
	schluesselAdminCert   = "ADMIN_TLS_CERT"
	schluesselAdminKey    = "ADMIN_TLS_KEY"
	schluesselAdminFlag   = "ADMIN_ALLOW_PLAINTEXT"
	dateiAdminToken       = "admin.token"
	standardAdminListen   = "127.0.0.1:8081"
)

// adminPruefung ist die Form von pruefeAdminTransport: Laden und Build rufen die Prüfung über einen Wert dieses Typs.
type adminPruefung func(adminEingabe, time.Time) (adminTransport, []verstoss)

// grundDateiDesDienstes ist der Grund einer Token-Datei, die der Dienst selbst schreibt (SPEC-300 Nr. 4).
const grundDateiDesDienstes = "Datei, die der Dienst selbst schreibt"

// adminPraefix ist der Anfang von schluesselAdminListen bis einschließlich dem ersten `_`; jeder andere Schlüssel mit diesem
// Anfang ist abgelehnt (SPEC-110 Nr. 2). Er steht nie als Literal in konfig*.go (SPEC-201 Nr. 4).
func adminPraefix() string {
	return schluesselAdminListen[:strings.IndexByte(schluesselAdminListen, '_')+1]
}

// liesAdmin liest die fünf Schlüssel in jedem Modus und lehnt jeden anderen gesetzten Schlüssel mit dem Präfix ab
// (SPEC-300 Nr. 2); die Pfade des Paars trägt k ohne Leerraum am Rand. Die Rückgabe heißt: ADMIN_LISTEN und
// ADMIN_ALLOW_PLAINTEXT haben ihre Form bestanden (SPEC-299 Nr. 5, 6). Es liest und legt keine Datei an; DataDir ist
// vorher gelesen.
func (p *pruefer) liesAdmin(k *Konfiguration) (formGueltig bool) {
	vorher := len(p.verstoesse)
	k.AdminListen = p.listen(schluesselAdminListen, standardAdminListen)
	k.AdminAllowPlaintext = p.wahrheitswert(schluesselAdminFlag, false)
	formGueltig = len(p.verstoesse) == vorher
	k.AdminTLSCert, _ = p.wert(schluesselAdminCert)
	k.AdminTLSKey, _ = p.wert(schluesselAdminKey)
	if w, ok := p.wert(schluesselAdminToken); ok {
		k.AdminTokenFile = w
	} else {
		k.AdminTokenFile = filepath.Join(k.DataDir, dateiAdminToken)
	}
	gelesen := []string{schluesselAdminListen, schluesselAdminToken, schluesselAdminCert, schluesselAdminKey, schluesselAdminFlag}
	for name, v := range p.env {
		if strings.HasPrefix(name, adminPraefix()) && !slices.Contains(gelesen, name) && kappe(v) != "" {
			p.melde(name, grundNichtUnterstuetzt)
		}
	}
	return formGueltig
}

// adminEingabeAus bildet die Eingabe der Prüfung aus den Werten der Konfiguration, unverändert (SPEC-300 Nr. 3).
func adminEingabeAus(k Konfiguration, formGueltig bool) adminEingabe {
	return adminEingabe{
		Listen: k.AdminListen, CertPfad: k.AdminTLSCert, KeyPfad: k.AdminTLSKey,
		Klartext: k.AdminAllowPlaintext, FormGueltig: formGueltig,
	}
}

// pruefeAdmin ist der Aufruf der Prüfung am Laden, nach den reinen Prüfungen und mit den Dateien von LISTEN_TLS_*, zur
// Zeit jetzt (SPEC-299 Nr. 8 a); die Verstöße gehen in die eine Meldung des Ladens, der Transport verfällt.
func (p *pruefer) pruefeAdmin(k *Konfiguration, formGueltig bool, jetzt time.Time, pruefe adminPruefung) {
	_, vs := pruefe(adminEingabeAus(*k, formGueltig), jetzt)
	p.verstoesse = append(p.verstoesse, vs...)
}

// ladeTransport ist der erste Schritt von Build: das Paar des TLS des Listeners (ladeListenTLS) und, mit nichtleerem
// AdminListen, der Transport des Admin-Listeners (SPEC-299 Nr. 8 b); die Zeilen beider stehen nach Schlüssel sortiert in
// einem Fehler (SPEC-139 Nr. 2). Ohne AdminListen ruft es die Prüfung nicht, und der Transport ist der Nullwert
// (SPEC-300 Nr. 11).
func ladeTransport(k Konfiguration, jetzt time.Time, pruefe adminPruefung) (*tls.Config, adminTransport, error) {
	cfg, err := ladeListenTLS(k, jetzt)
	if k.AdminListen == "" {
		return cfg, adminTransport{}, err
	}
	transport, vs := pruefe(adminEingabeAus(k, true), jetzt)
	if len(vs) == 0 {
		return cfg, transport, err
	}
	var zeilen []string
	var kf *KonfigFehler
	if errors.As(err, &kf) {
		zeilen = slices.Clone(kf.Zeilen)
	}
	for _, v := range vs {
		zeilen = append(zeilen, konfigZeile(v.schluessel, v.grund))
	}
	slices.Sort(zeilen)
	return nil, adminTransport{}, &KonfigFehler{Zeilen: zeilen}
}

// ladeAdminToken lädt oder erzeugt die Token-Datei (SPEC-300 Nr. 4, SPEC-239): ein Pfad auf eine Datei, die der Dienst selbst
// schreibt, wird vor LadeToken abgewiesen; ein TokenFehler ist die Zeile am Schlüssel ADMIN_TOKEN_FILE, jeder andere Fehler
// kommt unverändert zurück.
func ladeAdminToken(dataDir, pfad string) ([]byte, error) {
	if dateiDesDienstes(dataDir, pfad) {
		return nil, konfigFehler(schluesselAdminToken, grundDateiDesDienstes)
	}
	token, err := admin.LadeToken(pfad)
	var tf *admin.TokenFehler
	if errors.As(err, &tf) {
		return nil, konfigFehler(schluesselAdminToken, tf.Grund)
	}
	return token, err
}

// dateiDesDienstes meldet, ob pfad das Audit-Log oder eine Datei der Datenbank in dataDir benennt: der letzte Namensteil ist
// `audit.jsonl` oder beginnt mit dem Namen der Datenbank, und das Verzeichnis davor ist nach os.SameFile dataDir. Das
// Verzeichnis davor wird nicht lexikalisch gekürzt, damit `..` hinter einem Verweis wie beim Öffnen wirkt.
func dateiDesDienstes(dataDir, pfad string) bool {
	verzeichnis, name := filepath.Split(pfad)
	if name != audit.DateiName && !strings.HasPrefix(name, dateiDB) {
		return false
	}
	if verzeichnis == "" {
		verzeichnis = "."
	}
	a, errA := os.Stat(verzeichnis)
	b, errB := os.Stat(dataDir)
	return errA == nil && errB == nil && os.SameFile(a, b)
}
