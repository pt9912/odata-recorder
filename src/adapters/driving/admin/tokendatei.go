// Package admin ist der Driving Adapter des Admin-Listeners (SPEC-238). Die Token-Datei lädt und
// erzeugt LadeToken (SPEC-239); der Adapter importiert weder einen anderen Adapter noch einen Service.
package admin

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die sieben Gründe der Token-Datei (SPEC-239 Nr. 6); der Text ist Vertrag.
const (
	grundNichtLesbar   = "Datei nicht lesbar"
	grundVerzeichnis   = "ist ein Verzeichnis"
	grundKeineDatei    = "keine reguläre Datei"
	grundRechte        = "Gruppen- oder Weltrechte"
	grundMehrZeilen    = "mehr als eine Zeile"
	grundZuKurz        = "weniger als 32 Zeichen"
	grundNichtAnlegbar = "Datei nicht anlegbar"
)

const (
	// tokenMinBytes ist die kleinste Länge des Tokens nach dem Entfernen von Leerraum (SPEC-239 Nr. 1 g).
	tokenMinBytes = 32
	// leerraum ist der ASCII-Leerraum, der an beiden Enden entfällt (SPEC-239 Nr. 1 f).
	leerraum = " \t\n\v\f\r"
	// zufallBytes sind die 256 Bit des erzeugten Tokens (SPEC-239 Nr. 3 a).
	zufallBytes = 32
	// hilfsSuffixBytes bestimmt die 16 Hex-Zeichen im Namen der Hilfsdatei (SPEC-239 Nr. 3 b).
	hilfsSuffixBytes = 8
)

// TokenFehler ist ein Verstoß der Token-Datei (SPEC-239 Nr. 5, 6); Grund ist der Text hinter dem Doppelpunkt.
type TokenFehler struct{ Grund string }

// Error liefert `ODR-E2001 ADMIN_TOKEN_FILE: <Grund>`; weder Inhalt noch Pfad stehen im Text.
func (e *TokenFehler) Error() string { return "ODR-E2001 ADMIN_TOKEN_FILE: " + e.Grund }

func fehler(grund string) error { return &TokenFehler{Grund: grund} }

// haltepunkte sind die Nähte, an denen ein Test zwischen zwei Schritten des Erzeugens eingreift (SPEC-239 Nr. 3);
// beide erhalten den Pfad der Hilfsdatei.
type haltepunkte struct {
	nachAnlegen func(hilfsdatei string)
	vorLink     func(hilfsdatei string)
}

func keinHalt(string) {}

// LadeToken lädt das Token aus der Datei pfad oder erzeugt die fehlende Datei (SPEC-239). Ein leerer pfad
// ist model.ErrUngueltig; ein Verstoß der Datei ist *TokenFehler.
func LadeToken(pfad string) ([]byte, error) {
	return lade(pfad, haltepunkte{nachAnlegen: keinHalt, vorLink: keinHalt})
}

func lade(pfad string, h haltepunkte) ([]byte, error) {
	if pfad == "" {
		return nil, fmt.Errorf("%w: Pfad der Token-Datei fehlt", model.ErrUngueltig)
	}
	fi, err := os.Stat(pfad)
	if err != nil {
		if istFehlend(pfad, err) {
			return erzeuge(pfad, h)
		}
		return nil, statFehler(pfad, err)
	}
	return pruefeUndLies(pfad, fi)
}

// istFehlend: die Datei fehlt, und es liegt auch kein Symlink ohne Ziel an ihrer Stelle (SPEC-239 Nr. 1 a).
func istFehlend(pfad string, err error) bool {
	if !errors.Is(err, fs.ErrNotExist) {
		return false
	}
	_, lerr := os.Lstat(pfad)
	return lerr != nil
}

// statFehler bildet einen Fehler von Stat, der nicht „fehlt“ ist, auf seinen Grund ab (SPEC-239 Nr. 1 a).
func statFehler(pfad string, err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		if _, lerr := os.Lstat(pfad); lerr == nil {
			return fehler(grundKeineDatei)
		}
	}
	return fehler(grundNichtLesbar)
}

// pruefeUndLies führt die Schritte b bis g von SPEC-239 Nr. 1 aus; die Art wird vor dem Öffnen geprüft.
func pruefeUndLies(pfad string, fi fs.FileInfo) ([]byte, error) {
	switch {
	case fi.IsDir():
		return nil, fehler(grundVerzeichnis)
	case !fi.Mode().IsRegular():
		return nil, fehler(grundKeineDatei)
	case runtime.GOOS != "windows" && fi.Mode().Perm()&0o077 != 0:
		return nil, fehler(grundRechte)
	}
	roh, err := os.ReadFile(pfad)
	if err != nil {
		return nil, fehler(grundNichtLesbar)
	}
	rest := bytes.Trim(roh, leerraum)
	if bytes.ContainsAny(rest, "\r\n") {
		return nil, fehler(grundMehrZeilen)
	}
	if len(rest) < tokenMinBytes {
		return nil, fehler(grundZuKurz)
	}
	return bytes.Clone(rest), nil
}

// erzeuge legt die fehlende Datei über eine Hilfsdatei und einen Hardlink an (SPEC-239 Nr. 3): die Datei
// erscheint vollständig oder gar nicht, und von gleichzeitigen Aufrufen gewinnt einer.
func erzeuge(pfad string, h haltepunkte) ([]byte, error) {
	token := zufallHex(zufallBytes)
	hilfsdatei := pfad + ".tmp-" + zufallHex(hilfsSuffixBytes)
	f, err := os.OpenFile(hilfsdatei, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fehler(grundNichtAnlegbar)
	}
	defer func() { _ = os.Remove(hilfsdatei) }()
	h.nachAnlegen(hilfsdatei)
	if err := schreibe(f, token+"\n"); err != nil {
		return nil, fehler(grundNichtAnlegbar)
	}
	h.vorLink(hilfsdatei)
	err = os.Link(hilfsdatei, pfad)
	switch {
	case err == nil:
		return []byte(token), nil
	case errors.Is(err, fs.ErrExist):
		return ladeVorhandene(pfad)
	}
	return nil, fehler(grundNichtAnlegbar)
}

// ladeVorhandene lädt die Datei, die ein anderer Aufruf angelegt hat (SPEC-239 Nr. 3 d).
func ladeVorhandene(pfad string) ([]byte, error) {
	fi, err := os.Stat(pfad)
	if err != nil {
		return nil, statFehler(pfad, err)
	}
	return pruefeUndLies(pfad, fi)
}

// schreibe setzt die Rechte auf 0600, schreibt, synchronisiert und schließt die Hilfsdatei.
func schreibe(f *os.File, inhalt string) error {
	err := f.Chmod(0o600)
	if err == nil {
		_, err = f.WriteString(inhalt)
	}
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// zufallHex liefert n Bytes aus crypto/rand als 2n Zeichen 0-9a-f; crypto/rand.Read meldet keinen Fehler.
func zufallHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
