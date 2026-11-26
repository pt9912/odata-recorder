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

// Die zwei Schlüssel des Admin-Listeners, die diese Version liest (SPEC-110 Nr. 2, SPEC-300 Nr. 2), und die Namen, unter
// denen die Token-Datei in DATA_DIR liegt.
const (
	schluesselAdminListen = "ADMIN_LISTEN"
	schluesselAdminToken  = "ADMIN_TOKEN_FILE"
	dateiAdminToken       = "admin.token"
	standardAdminListen   = "127.0.0.1:8081"
)

// grundDateiDesDienstes ist der Grund einer Token-Datei, die der Dienst selbst schreibt (SPEC-300 Nr. 4).
const grundDateiDesDienstes = "Datei, die der Dienst selbst schreibt"

// adminPraefix ist der Anfang von schluesselAdminListen bis einschließlich dem ersten `_`; jeder andere Schlüssel mit diesem
// Anfang ist abgelehnt (SPEC-110 Nr. 2). Er steht nie als Literal in konfig*.go (SPEC-201 Nr. 4).
func adminPraefix() string {
	return schluesselAdminListen[:strings.IndexByte(schluesselAdminListen, '_')+1]
}

// liesAdmin liest ADMIN_LISTEN und ADMIN_TOKEN_FILE in jedem Modus, lehnt jeden anderen gesetzten Schlüssel mit dem Präfix ab
// und prüft den Zusammenhang von ADMIN_LISTEN ohne Paar und ohne Flag, wenn dessen Form gilt (SPEC-300 Nr. 2, SPEC-299 Nr. 6).
// Die Token-Datei liest und legt es nicht an; DataDir ist vorher gelesen.
func (p *pruefer) liesAdmin(k *Konfiguration, jetzt time.Time) {
	vorher := len(p.verstoesse)
	k.AdminListen = p.listen(schluesselAdminListen, standardAdminListen)
	formGueltig := len(p.verstoesse) == vorher
	if w, ok := p.wert(schluesselAdminToken); ok {
		k.AdminTokenFile = w
	} else {
		k.AdminTokenFile = filepath.Join(k.DataDir, dateiAdminToken)
	}
	for name, v := range p.env {
		if strings.HasPrefix(name, adminPraefix()) && name != schluesselAdminListen && name != schluesselAdminToken && kappe(v) != "" {
			p.melde(name, grundNichtUnterstuetzt)
		}
	}
	_, vs := pruefeAdminTransport(adminEingabe{Listen: k.AdminListen, FormGueltig: formGueltig}, jetzt)
	p.verstoesse = append(p.verstoesse, vs...)
}

// ladeTransport ist der erste Schritt von Build: das Paar des TLS des Listeners (ladeListenTLS) und, mit Admin-Listener, der
// Zusammenhang von ADMIN_LISTEN ohne Paar und ohne Flag; die Zeilen beider stehen nach Schlüssel sortiert in einem Fehler
// (SPEC-139 Nr. 2).
func ladeTransport(k Konfiguration, jetzt time.Time) (*tls.Config, error) {
	cfg, err := ladeListenTLS(k, jetzt)
	if k.AdminListen == "" {
		return cfg, err
	}
	_, vs := pruefeAdminTransport(adminEingabe{Listen: k.AdminListen, FormGueltig: true}, jetzt)
	if len(vs) == 0 {
		return cfg, err
	}
	p := &pruefer{verstoesse: vs}
	var zeilen []string
	var kf *KonfigFehler
	if errors.As(err, &kf) {
		zeilen = kf.Zeilen
	}
	if errors.As(p.fehler(), &kf) {
		zeilen = append(slices.Clone(kf.Zeilen), zeilen...)
	}
	slices.Sort(zeilen)
	return nil, &KonfigFehler{Zeilen: zeilen}
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
