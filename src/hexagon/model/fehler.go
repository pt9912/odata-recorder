package model

import (
	"errors"
	"fmt"
	"slices"
)

// Code ist ein Fehlercode aus §4 der Spezifikation (SPEC-109). Der Nullwert ist kein Code;
// einen Code außerhalb von §4 stellt der Typ nicht dar.
type Code int

// Die Codes aus §4 der Spezifikation.
const (
	CodeE2001 Code = iota + 1
	CodeE2101
	CodeE2401
	CodeE3001
	CodeE4001
	CodeE5001
	CodeE5002
	CodeE5003
	CodeE5004
	CodeE5005
	CodeE5006
	CodeE5007
	CodeE5008
	CodeE5009
	CodeE6001
	CodeE6002
	CodeE6003
	CodeE6004
	CodeE6101
	CodeE6102
	CodeE7001
	CodeE7002
	CodeE7003
	CodeE7004
	CodeE7005
	CodeE7006
	CodeE7007
	CodeE7008
	CodeE7009
	CodeE7010
	CodeE7011
	CodeE7012
	CodeE9001
)

// Zielbereich ist das Feld target des Fehlerobjekts (SPEC-109 Nr. 2).
type Zielbereich string

// Die Zielbereiche.
const (
	ZielConfig    Zielbereich = "config"
	ZielAdmin     Zielbereich = "admin"
	ZielRecord    Zielbereich = "record"
	ZielStorage   Zielbereich = "storage"
	ZielUpstream  Zielbereich = "upstream"
	ZielReplay    Zielbereich = "replay"
	ZielService   Zielbereich = "service"
	ZielMock      Zielbereich = "mock"
	ZielImport    Zielbereich = "import"
	ZielWorkspace Zielbereich = "workspace"
	ZielSession   Zielbereich = "session"
	ZielInternal  Zielbereich = "internal"
)

type codeZeile struct {
	code    Code
	text    string
	stati   []int
	ziel    Zielbereich
	message string
}

// codeZeilen ist die Tabelle aus SPEC-109 Nr. 2; Status 0 heißt, der Code wird nie als
// HTTP-Antwort ausgeliefert.
func codeZeilen() []codeZeile {
	return []codeZeile{
		{CodeE2001, "ODR-E2001", []int{0}, ZielConfig, "Invalid configuration"},
		{CodeE2101, "ODR-E2101", []int{401}, ZielAdmin, "Admin authentication missing or invalid"},
		{CodeE2401, "ODR-E2401", []int{0}, ZielRecord, "Capture or redaction failed"},
		{CodeE3001, "ODR-E3001", []int{500}, ZielStorage, "Stored data damaged or not loadable"},
		{CodeE4001, "ODR-E4001", []int{502, 504}, ZielUpstream, "Upstream unavailable"},
		{CodeE5001, "ODR-E5001", []int{404}, ZielReplay, "No matching recorded exchange"},
		{CodeE5002, "ODR-E5002", []int{410}, ZielReplay, "Recorded exchanges exhausted"},
		{CodeE5003, "ODR-E5003", []int{400}, ZielReplay, "Workspace or replay assignment missing or ambiguous"},
		{CodeE5004, "ODR-E5004", []int{409}, ZielReplay, "Assignment belongs to another service"},
		{CodeE5005, "ODR-E5005", []int{404}, ZielService, "No service configured for the path"},
		{CodeE5006, "ODR-E5006", []int{400}, ZielReplay, "Invalid request encoding"},
		{CodeE5007, "ODR-E5007", []int{409}, ZielReplay, "Next expected exchange has a different match key"},
		{CodeE5008, "ODR-E5008", []int{422}, ZielReplay, "Matched exchange is not replayable"},
		{CodeE5009, "ODR-E5009", []int{409}, ZielRecord, "No recording session accepts the request"},
		{CodeE6001, "ODR-E6001", []int{501}, ZielMock, "Unsupported query or read access"},
		{CodeE6002, "ODR-E6002", []int{501}, ZielMock, "Unsupported mutation or write access"},
		{CodeE6003, "ODR-E6003", []int{422}, ZielMock, "Incomplete coverage"},
		{CodeE6004, "ODR-E6004", []int{422, 0}, ZielImport, "Metadata invalid or unsupported"},
		{CodeE6101, "ODR-E6101", []int{422}, ZielImport, "Baseline import invalid or inconsistent"},
		{CodeE6102, "ODR-E6102", []int{409}, ZielImport, "Import id reused with different content"},
		{CodeE7001, "ODR-E7001", []int{404}, ZielWorkspace, "Workspace unknown"},
		{CodeE7002, "ODR-E7002", []int{404}, ZielWorkspace, "Snapshot unknown"},
		{CodeE7003, "ODR-E7003", []int{404}, ZielSession, "Recording session unknown"},
		{CodeE7004, "ODR-E7004", []int{409}, ZielSession, "Session is referenced by a baseline"},
		{CodeE7005, "ODR-E7005", []int{404}, ZielImport, "Import unknown"},
		{CodeE7006, "ODR-E7006", []int{409}, ZielWorkspace, "Snapshot name already taken"},
		{CodeE7007, "ODR-E7007", []int{409}, ZielWorkspace, "Workspace id already taken"},
		{CodeE7008, "ODR-E7008", []int{422}, ZielWorkspace, "Snapshot kind does not fit the operation"},
		{CodeE7009, "ODR-E7009", []int{409}, ZielSession, "Session is still open"},
		{CodeE7010, "ODR-E7010", []int{409}, ZielSession, "Session id already taken"},
		{CodeE7011, "ODR-E7011", []int{409}, ZielSession, "Session is not open"},
		{CodeE7012, "ODR-E7012", []int{404}, ZielSession, "Exchange unknown"},
		{CodeE9001, "ODR-E9001", []int{500}, ZielInternal, "Internal error"},
	}
}

func (c Code) zeile() (codeZeile, bool) {
	for _, z := range codeZeilen() {
		if z.code == c {
			return z, true
		}
	}
	return codeZeile{}, false
}

// AlleCodes liefert jeden Code aus §4 in der Reihenfolge der Tabelle.
func AlleCodes() []Code {
	zeilen := codeZeilen()
	out := make([]Code, len(zeilen))
	for i, z := range zeilen {
		out[i] = z.code
	}
	return out
}

// ParseCode liest die Schreibweise `ODR-E2001`; ein Text außerhalb von §4 ist kein Code.
func ParseCode(text string) (Code, bool) {
	for _, z := range codeZeilen() {
		if z.text == text {
			return z.code, true
		}
	}
	return 0, false
}

// Gueltig meldet, ob c ein Code aus §4 ist.
func (c Code) Gueltig() bool {
	_, ok := c.zeile()
	return ok
}

// String liefert `ODR-E2001`; ein ungültiger Code liefert den leeren Text.
func (c Code) String() string {
	z, _ := c.zeile()
	return z.text
}

// Zielbereich liefert das Feld target; ein ungültiger Code liefert den leeren Text.
func (c Code) Zielbereich() Zielbereich {
	z, _ := c.zeile()
	return z.ziel
}

// Message liefert den festen Text des Codes (SPEC-109 Nr. 3); ein ungültiger Code liefert den leeren Text.
func (c Code) Message() string {
	z, _ := c.zeile()
	return z.message
}

// Stati liefert die HTTP-Status, die der Code tragen darf; 0 heißt „ohne Status“ (SPEC-109 Nr. 2).
func (c Code) Stati() []int {
	z, _ := c.zeile()
	return slices.Clone(z.stati)
}

// StatusErlaubt meldet, ob der Code den Status tragen darf; ein ungültiger Code trägt keinen.
func (c Code) StatusErlaubt(status int) bool {
	z, _ := c.zeile()
	return slices.Contains(z.stati, status)
}

// Fehlerobjekt ist der Body einer Fehlerantwort (SPEC-006, SPEC-109 Nr. 1).
type Fehlerobjekt struct {
	Error FehlerInhalt `json:"error"`
}

// FehlerInhalt trägt die drei Felder des Fehlerobjekts.
type FehlerInhalt struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target"`
}

// NeuesFehlerobjekt bildet das Fehlerobjekt zu c; ein ungültiger Code ist ErrUngueltig.
func NeuesFehlerobjekt(c Code) (Fehlerobjekt, error) {
	z, ok := c.zeile()
	if !ok {
		return Fehlerobjekt{}, errors.Join(ErrUngueltig, fmt.Errorf("Code %d steht nicht in §4 der Spezifikation", int(c)))
	}
	return Fehlerobjekt{Error: FehlerInhalt{Code: z.text, Message: z.message, Target: string(z.ziel)}}, nil
}
