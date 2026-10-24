package outboundtest

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Teil Z des Vertrags (SPEC-126 Nr. 10 c, d): die kanonischen Formen von Edm.Guid, Edm.Date und Edm.DateTimeOffset.

const guidKlein = "0f8fad5b-d9cb-469f-a165-70867728950e"

func jsonString(s string) string { return `"` + s + `"` }

// jsonZeichen ist das Zeichen c als Inhalt eines JSON-Strings.
func jsonZeichen(c rune) string {
	switch {
	case c == '"':
		return `\"`
	case c == '\\':
		return `\\`
	case c < 0x20:
		return esc(fmt.Sprintf("%04x", c))
	}
	return string(c)
}

func jfF6() []jfall {
	gut := func(in, soll string) wertFall { return gutW(jsonString(in), model.ArtGUID, soll) }
	werte := []wertFall{
		gut(guidKlein, guidKlein), gut(strings.ToUpper(guidKlein), guidKlein), gut("0F8fAd5B-d9CB-469f-A165-70867728950E", guidKlein),
		gut("00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000"),
		gut("ffffffff-ffff-ffff-ffff-ffffffffffff", "ffffffff-ffff-ffff-ffff-ffffffffffff"),
	}
	for _, schlecht := range []string{
		"{" + guidKlein + "}", strings.ReplaceAll(guidKlein, "-", ""), " " + guidKlein, guidKlein + " ", guidKlein[:35], guidKlein + "0",
		"0f8fad5bd-9cb-469f-a165-70867728950e", "0f8fad5b-d9cb-469f-a1657-0867728950e", "urn:uuid:" + guidKlein, "", "0f8fad5b_d9cb_469f_a165_70867728950e",
		strings.Replace(guidKlein, "0", string(rune(0xff10)), 1), strings.Replace(guidKlein, "a", string(rune(0x430)), 1),
	} {
		werte = append(werte, schlechtW(jsonString(schlecht), model.DatenWert))
	}
	for pos := range 36 {
		if pos == 8 || pos == 13 || pos == 18 || pos == 23 {
			continue
		}
		for c := range rune(0x80) {
			zeichen := guidKlein[:pos] + string(c) + guidKlein[pos+1:]
			json := jsonString(guidKlein[:pos] + jsonZeichen(c) + guidKlein[pos+1:])
			if hex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F'); hex {
				werte = append(werte, gutW(json, model.ArtGUID, strings.ToLower(zeichen)))
			} else {
				werte = append(werte, schlechtW(json, model.DatenWert))
			}
		}
	}
	return wertFaelle("F6", "Edm.Guid", werte, true)
}

func jF6(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF6()) }

// datumGueltig ist das Orakel für Edm.Date: time.Date normiert einen Tag außerhalb des Monats, die Rundreise erkennt es.
func datumGueltig(jahr, monat, tag int) bool {
	if jahr < 1 || jahr > 9999 || monat < 1 || monat > 12 || tag < 1 {
		return false
	}
	t := time.Date(jahr, time.Month(monat), tag, 0, 0, 0, 0, time.UTC)
	return t.Year() == jahr && int(t.Month()) == monat && t.Day() == tag
}

func jfF7() []jfall {
	var werte []wertFall
	for _, d := range []string{"0001-01-01", "9999-12-31", "2024-02-29", "2000-02-29", "1900-02-28", "2023-12-31"} {
		werte = append(werte, gutW(jsonString(d), model.ArtDate, d))
	}
	ziffer0 := string(rune(0xff10))
	for _, d := range []string{
		"0000-01-01", "10000-01-01", "2023-02-29", "1900-02-29", "2100-02-29", "2024-13-01", "2024-00-10", "2024-04-31", "2024-01-32", "2024-01-00",
		"2024-1-5", "20240105", "2024-01-05T00:00:00Z", " 2024-01-05", "2024-01-05 ", "+2024-01-05", "-2024-01-05", "2024/01/05", "",
		ziffer0 + "024-01-05", "2024-" + ziffer0 + "1-05", "2024-01-" + ziffer0 + "5", "2024-01-5 ", "2024-01-05Z",
		"2024/01-05", "2024-01/05", "2024:01-05", "2024-01:05", "2024 01-05", "2024-01 05", " 024-01-05", "2024- 1-05", "2024-01- 5",
	} {
		werte = append(werte, schlechtW(jsonString(d), model.DatenWert))
	}
	for _, jahr := range []int{0, 1, 4, 100, 400, 1900, 2000, 2023, 2024, 2100, 9999, 10000} {
		for monat := range 14 {
			for _, tag := range []int{0, 1, 27, 28, 29, 30, 31, 32} {
				text := fmt.Sprintf("%04d-%02d-%02d", jahr, monat, tag)
				if datumGueltig(jahr, monat, tag) {
					werte = append(werte, gutW(jsonString(text), model.ArtDate, text))
				} else {
					werte = append(werte, schlechtW(jsonString(text), model.DatenWert))
				}
			}
		}
	}
	return wertFaelle("F7", "Edm.Date", werte, true)
}

func jF7(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF7()) }

// tageSeit1970 und zivilVonTagen sind der Kalender des Orakels von F8 ohne das Paket time (Algorithmus von Howard Hinnant).
func tageSeit1970(j, m, d int) int64 {
	if m <= 2 {
		j--
	}
	era := j / 400
	if j < 0 {
		era = (j - 399) / 400
	}
	jahrDerEra := j - era*400
	ersatzMonat := m - 3
	if m <= 2 {
		ersatzMonat = m + 9
	}
	tagDesJahres := (153*ersatzMonat+2)/5 + d - 1
	tagDerEra := jahrDerEra*365 + jahrDerEra/4 - jahrDerEra/100 + tagDesJahres
	return int64(era)*146097 + int64(tagDerEra) - 719468
}

func zivilVonTagen(z int64) (j, m, d int) {
	z += 719468
	era := z / 146097
	if z < 0 {
		era = (z - 146096) / 146097
	}
	tagDerEra := z - era*146097
	jahrDerEra := (tagDerEra - tagDerEra/1460 + tagDerEra/36524 - tagDerEra/146096) / 365
	tagDesJahres := tagDerEra - (365*jahrDerEra + jahrDerEra/4 - jahrDerEra/100)
	mp := (5*tagDesJahres + 2) / 153
	d = int(tagDesJahres - (153*mp+2)/5 + 1)
	m = int(mp + 3)
	if mp >= 10 {
		m = int(mp - 9)
	}
	j = int(jahrDerEra + era*400)
	if m <= 2 {
		j++
	}
	return j, m, d
}

// zeitpunktOrakel liefert die kanonische Schreibweise eines Zeitpunkts aus seinen Teilen und ob er im Bereich 0001 bis 9999 liegt.
func zeitpunktOrakel(j, mo, d, h, mi, s int, bruch string, versatzMinuten int) (string, bool) {
	sekunden := tageSeit1970(j, mo, d)*86400 + int64(h*3600+mi*60+s) - int64(versatzMinuten*60)
	tage := sekunden / 86400
	rest := sekunden % 86400
	if rest < 0 {
		tage, rest = tage-1, rest+86400
	}
	uj, um, ud := zivilVonTagen(tage)
	if uj < 1 || uj > 9999 {
		return "", false
	}
	text := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d", uj, um, ud, rest/3600, rest%3600/60, rest%60)
	if bruch = strings.TrimRight(bruch, "0"); bruch != "" {
		text += "." + bruch
	}
	return text + "Z", true
}

// zeitpunktGenerator ist das Orakel von F8 über Datum × Zeit × Bruchteil × Versatz.
func zeitpunktGenerator() []wertFall {
	var werte []wertFall
	versaetze := []struct {
		text string
		min  int
	}{{"Z", 0}, {"+00:00", 0}, {"-00:00", 0}, {"+01:00", 60}, {"-05:30", -330}, {"+23:59", 23*60 + 59}, {"-23:59", -(23*60 + 59)}}
	for _, dat := range []struct{ j, m, d int }{{1, 1, 1}, {1, 1, 2}, {1969, 12, 31}, {1970, 1, 1}, {2020, 2, 29}, {2023, 12, 31}, {9999, 12, 30}, {9999, 12, 31}} {
		for _, zeit := range []struct {
			text     string
			h, mi, s int
			sek      bool
		}{{"00:00", 0, 0, 0, false}, {"23:59", 23, 59, 0, false}, {"00:00:00", 0, 0, 0, true}, {"12:34:56", 12, 34, 56, true}, {"23:59:59", 23, 59, 59, true}} {
			bruchteile := []string{""}
			if zeit.sek {
				bruchteile = append(bruchteile, "5", "50", "123456789012", "000000000000", "000000000001")
			}
			for _, bruch := range bruchteile {
				for _, v := range versaetze {
					text := fmt.Sprintf("%04d-%02d-%02dT%s", dat.j, dat.m, dat.d, zeit.text)
					if bruch != "" {
						text += "." + bruch
					}
					soll, ok := zeitpunktOrakel(dat.j, dat.m, dat.d, zeit.h, zeit.mi, zeit.s, bruch, v.min)
					werte = append(werte, zeitpunktFall(text+v.text, soll, ok))
				}
			}
		}
	}
	return werte
}

// zeitpunktFall ist der Fall zu einem Text: gültig mit dem Text des Orakels oder `wert`.
func zeitpunktFall(text, soll string, gueltig bool) wertFall {
	if gueltig {
		return gutW(jsonString(text), model.ArtDateTimeOffset, soll)
	}
	return schlechtW(jsonString(text), model.DatenWert)
}

func jfF8() []jfall {
	ok := func(in, soll string) wertFall { return gutW(jsonString(in), model.ArtDateTimeOffset, soll) }
	const utc = "2020-01-01T00:00:00Z"
	werte := []wertFall{
		ok(utc, utc), ok("2020-01-01T00:00Z", utc), ok("2020-01-01T01:00:00+01:00", utc), ok("2019-12-31T19:00:00-05:00", utc),
		ok("2020-01-01T00:00:00.0Z", utc), ok("2020-01-01T00:00:00.000Z", utc),
		ok("2020-01-01T00:00:00.5Z", "2020-01-01T00:00:00.5Z"), ok("2020-01-01T00:00:00.50Z", "2020-01-01T00:00:00.5Z"),
		ok("2020-01-01T00:00:00.500000000000Z", "2020-01-01T00:00:00.5Z"), ok("2020-01-01T00:00:00.000000000001Z", "2020-01-01T00:00:00.000000000001Z"),
		ok("2020-01-01T00:00:00+00:00", utc), ok("2020-01-01T00:00:00-00:00", utc), ok("2020-01-01T00:00:00+23:59", "2019-12-31T00:01:00Z"),
		ok("2020-01-01T00:00:00-23:59", "2020-01-01T23:59:00Z"),
		ok("0001-01-01T00:00:00Z", "0001-01-01T00:00:00Z"), ok("0001-01-01T00:00:00-00:01", "0001-01-01T00:01:00Z"),
		ok("9999-12-31T23:59:59Z", "9999-12-31T23:59:59Z"), ok("9999-12-31T23:59:59.999999999999Z", "9999-12-31T23:59:59.999999999999Z"),
		ok("2024-03-01T00:30:00+01:00", "2024-02-29T23:30:00Z"), ok("2024-01-01T00:30:00+01:00", "2023-12-31T23:30:00Z"),
		ok("2023-03-01T00:30:00+01:00", "2023-02-28T23:30:00Z"), ok("2020-01-01T00:00:00-12:00", "2020-01-01T12:00:00Z"),
	}
	for _, schlecht := range []string{
		"0001-01-01T00:00:00+00:01", "9999-12-31T23:59:59-00:01", "2020-01-01t00:00:00Z", "2020-01-01T00:00:00z", "2020-01-01 00:00:00Z",
		"2020-01-01T00:00:00", "2020-01-01T00:00", "2020-01-01T00:00:60Z", "2020-01-01T00:60:00Z", "2020-01-01T24:00:00Z", "2020-01-01T00Z",
		"2020-01-01T00:00.5Z", "2020-01-01T00:00:00+1:00", "2020-01-01T00:00:00+0100", "2020-01-01T00:00:00 Z", " 2020-01-01T00:00:00Z",
		"2020-01-01T00:00:00Z ", "2020-01-01T00:00:00.0000000000000Z", "2020-01-01T00:00:00.Z", "2020-01-01T00:00:00+24:00",
		"2020-01-01T00:00:00+01:60", "2020-01-01T00:00:00+01:00:00", "2020-02-30T00:00:00Z", "2020-01-01T00:00:00,5Z", "2020-01-01", "",
		"2020-01-01T0:00:00Z", "2020-01-01T00:00:00Zz", "2020-01-01T00:00:00+01", "2020-01-01T00:00:00 +01:00", "10000-01-01T00:00:00Z",
		"2020-01-01T" + string(rune(0xff10)) + "0:00:00Z", "2020-01-01T00:00:", "2020-01-01T00:00:Z", "2020-01-01T00:00:0Z", "2020-01-01T00:00:00.5",
		"2020/01-01T00:00:00Z", "2020-01/01T00:00:00Z", "2020-01-01T00-00:00Z", "2020-01-01T00:00-00Z", "2020-01-01T 0:00:00Z", "2020-01-01T00: 0:00Z",
		"2020-01-01T00:00: 0Z", "2020-01-01T00:00:00+01:0", "2020-01-01T00:00:00+01:000", "2020-01-01T00:00:00+01-00", "2020-01-01T00:00:00+01000",
		"2020-01-01T00:00:00 01:00", "2020-01-01T00:00:00+ 1:00", "2020-01-01T00:00:00+01: 0",
	} {
		werte = append(werte, schlechtW(jsonString(schlecht), model.DatenWert))
	}
	return verbinde(wertFaelle("F8", "Edm.DateTimeOffset", werte, false), wertFaelle("F8 Generator", "Edm.DateTimeOffset", zeitpunktGenerator(), true))
}

func jF8(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF8()) }

// jfZ1: der Composite-Schlüssel des Dienstes `composite.xml` mit Edm.Guid, Edm.Date und Edm.DateTimeOffset als Properties;
// jede Reihenfolge der Properties im Body ergibt denselben Schlüssel, der Guid steht in Kleinbuchstaben.
func jfZ1() []jfall {
	var out []jfall
	teile := []string{`"OrderID":"` + strings.ToUpper(guidKlein) + `"`, `"LineNo":2`, `"Shipped":"2024-02-29"`, `"At":"2020-01-01T01:00:00+01:00"`}
	for _, p := range permutationen(len(teile)) {
		var ms []string
		for _, i := range p {
			ms = append(ms, teile[i])
		}
		fs := beideGut(fmt.Sprintf("Z1 Reihenfolge %v", p), "KZ", "Lines", strings.Join(ms, ","))
		for i := range fs {
			fs[i].schluessel = "OrderID=" + guidKlein + ",LineNo=2"
		}
		out = append(out, fs...)
	}
	return append(out,
		beide("Z1 Guid mit Klammern im Schlüssel", "KZ", "Lines", `"OrderID":"{`+guidKlein+`}","LineNo":2`, model.DatenWert, "/OrderID")...)
}

func jZ1(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfZ1()) }
