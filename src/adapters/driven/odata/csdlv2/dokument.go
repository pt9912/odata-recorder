package csdlv2

import "github.com/pt9912/odata-recorder/src/hexagon/model"

type art int

// Die Elemente, die der Leser auswertet, und kOpak für den Inhalt, der nur gezählt wird.
const (
	kEdmx art = iota + 1
	kDataServices
	kSchema
	kEntityType
	kKey
	kPropertyRef
	kProperty
	kNavProp
	kAssoc
	kAssocEnd
	kContainer
	kSet
	kAssocSet
	kAssocSetEnd
	kOpak
)

// rahmen ist ein offenes Element; ort ist sein Ort nach SPEC-171 Nr. 4.
type rahmen struct {
	art   art
	ort   string
	typ   *rawType
	assoc *rawAssoc
	aset  *rawAssocSet
}

// slot ist eine Stelle im Bericht: ein Fund oder, für eine Property, die Stelle ihres Fundes
// `property_typ`, den erst das Auflösen entscheidet.
type slot struct {
	fund model.Fund
	mem  *member
}

// member ist eine Property oder Navigation-Property; die Felder ab edm füllt das Auflösen.
type member struct {
	nav      bool
	name     string
	typ      string
	nullable bool
	rel      string
	von      string
	nach     string
	// aufgelöst
	edm        string
	fundName   string
	collection bool
	assoc      *rawAssoc
}

type rawType struct {
	namespace string
	name      string
	qn        string
	hasBase   bool
	baseRaw   string
	base      *rawType
	abstract  bool
	keys      int
	refs      []string
	refSet    map[string]bool
	members   []*member
	namen     map[string]bool
}

// rawEnd ist ein End einer Association.
type rawEnd struct {
	typ  string
	role string
	mult string
}

type rawAssoc struct {
	qn   string
	ends []rawEnd
}

// end liefert das End mit der Rolle.
func (a *rawAssoc) end(rolle string) (rawEnd, bool) {
	for _, e := range a.ends {
		if e.role == rolle {
			return e, true
		}
	}
	return rawEnd{}, false
}

// rawSetEnd ist ein End einer AssociationSet.
type rawSetEnd struct {
	role string
	set  string
}

type rawAssocSet struct {
	name  string
	ort   string
	assoc string
	ends  []rawSetEnd
	ziel  *rawAssoc
}

// set liefert das EntitySet des End mit der Rolle.
func (s *rawAssocSet) set(rolle string) (string, bool) {
	for _, e := range s.ends {
		if e.role == rolle {
			return e.set, true
		}
	}
	return "", false
}

type rawSet struct {
	name string
	typ  string
	ort  string
}

// kindRef ist ein EntitySet oder eine AssociationSet in der Reihenfolge der Kinder des Containers.
type kindRef struct {
	set  *rawSet
	aset *rawAssocSet
}

type rawContainer struct {
	qn        string
	reihe     []kindRef
	setNamen  map[string]bool
	asetNamen map[string]bool
}

// dokument sammelt, was das Lesen findet, in der Reihenfolge des Dokuments.
type dokument struct {
	namespaces []string
	abbild     map[string]string
	typen      []*rawType
	index      map[string]*rawType
	assocs     map[string]*rawAssoc
	deklariert map[string]bool
	container  *rawContainer
	// navs sind die Navigation-Properties je Association in der Reihenfolge des Dokuments; bindungen ist die laufende
	// Zahl der Bindings, die das Auflösen der AssociationSet gezählt hat.
	navs      map[*rawAssoc][]*member
	bindungen int
}

func neuesDokument() dokument {
	return dokument{abbild: map[string]string{}, index: map[string]*rawType{}, assocs: map[string]*rawAssoc{}, deklariert: map[string]bool{}}
}
