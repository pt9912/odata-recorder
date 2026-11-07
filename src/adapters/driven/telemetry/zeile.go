package telemetry

import (
	"log/slog"
	"slices"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

const (
	maxMsg        = 256
	zeitFormat    = "2006-01-02T15:04:05.000000000Z"
	hexZiffern    = "0123456789abcdef"
	zeilenTrenner = 0x2028
	absatzTrenner = 0x2029
)

// logFolge ist die Reihenfolge der Felder der Menge Log hinter timestamp, level und msg
// (SPEC-231 Nr. 1).
func logFolge() []string {
	return []string{
		"mode", "service_id", "session_id", "exchange_id", "workspace_id", "correlation_id",
		"error_code", "duration_ms", "ereignis", "richtung", "grund", "blob", "befunde",
	}
}

// eintrag prüft die Attribute, zählt die Verwürfe und schreibt ab MinLevel eine Zeile; die Uhr
// wird nur für eine geschriebene Zeile gerufen (SPEC-231 Nr. 2).
func (a *Adapter) eintrag(level slog.Level, msg string, attrs []support.Attr) {
	angenommen, verworfen := filtern(opLog, attrs)
	a.verwerfen(verworfen...)
	if level < a.minLevel {
		return
	}
	folge := logFolge()
	slices.SortFunc(angenommen, func(x, y support.Attr) int {
		return slices.Index(folge, x.Key) - slices.Index(folge, y.Key)
	})
	a.zeileSchreiben(a.clock.Now(), level, msg, angenommen)
}

// zeileSchreiben bildet die Zeile aus Zeit, Stufe, Text und den Feldern in der Reihenfolge der
// Ausgabe und schickt sie in einem Write unter der Sperre der Zeile hinaus; ein Fehler der Senke
// oder weniger Bytes verwerfen die Zeile und zählen senke (SPEC-232 Nr. 3, 5 c, 5 g).
func (a *Adapter) zeileSchreiben(t time.Time, level slog.Level, msg string, felder []support.Attr) {
	z := make([]byte, 0, 256)
	z = append(z, `{"timestamp":"`...)
	z = append(z, t.UTC().Format(zeitFormat)...)
	z = append(z, `","level":"`...)
	z = append(z, stufe(level)...)
	z = append(z, `","msg":`...)
	z = anfuehren(z, msgText(msg))
	for _, f := range felder {
		z = append(z, ',')
		z = anfuehren(z, f.Key)
		z = append(z, ':')
		switch f.Art {
		case support.ArtInt:
			z = strconv.AppendInt(z, f.Ganz, 10)
		case support.ArtFloat:
			z = strconv.AppendFloat(z, f.Gleit, 'f', 3, 64)
		default:
			z = anfuehren(z, f.Text)
		}
	}
	z = append(z, '}', '\n')

	a.zeilenMu.Lock()
	n, err := a.writer.Write(z)
	a.zeilenMu.Unlock()
	if err != nil || n < len(z) {
		a.verwerfen(support.VerwurfSenke)
	}
}

// stufe bildet den Wert des Levels auf einen der vier Namen ab (SPEC-231 Nr. 2).
func stufe(l slog.Level) string {
	switch {
	case l < slog.LevelInfo:
		return "DEBUG"
	case l < slog.LevelWarn:
		return "INFO"
	case l < slog.LevelError:
		return "WARN"
	}
	return "ERROR"
}

// msgText ersetzt je ungültigem Byte durch U+FFFD und kürzt dann auf den längsten Anfang bis
// 256 Bytes an einer Zeichengrenze (SPEC-231 Nr. 3).
func msgText(s string) string {
	out := make([]byte, 0, min(len(s), maxMsg)+utf8.UTFMax)
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		if len(out)+utf8.RuneLen(r) > maxMsg {
			break
		}
		out = utf8.AppendRune(out, r)
		i += n
	}
	return string(out)
}

// anfuehren hängt s als JSON-Text an z; s ist gültiges UTF-8. Maskiert werden allein ", \, die Zeichen
// unter U+0020 und die Zeilentrenner U+2028 und U+2029 (SPEC-231 Nr. 1).
func anfuehren(z []byte, s string) []byte {
	z = append(z, '"')
	for _, r := range s {
		switch {
		case r == '"' || r == '\\':
			z = append(z, '\\', byte(r))
		case r < ' ' || r == zeilenTrenner || r == absatzTrenner:
			z = append(z, '\\', 'u', hexZiffern[r>>12&15], hexZiffern[r>>8&15], hexZiffern[r>>4&15], hexZiffern[r&15])
		default:
			z = utf8.AppendRune(z, r)
		}
	}
	return append(z, '"')
}
