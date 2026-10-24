package csdl

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
	kContainer
	kSet
	kBinding
	kAnnotations
	kOpak
)

// rahmen ist ein offenes Element; ort ist sein Ort nach Nr. 7.
type rahmen struct {
	art art
	ort string
	typ *rawType
	set *rawSet
}

// slot ist eine Stelle im Bericht: ein Fund oder, für eine Property, die Stelle ihres Fundes
// `property_typ`, den erst das Auflösen entscheidet.
type slot struct {
	fund model.Fund
	mem  *member
}

type member struct {
	nav      bool
	name     string
	typ      string
	nullable bool
	// aufgelöst
	edm        string
	collection bool
}

type rawType struct {
	namespace string
	name      string
	qn        string
	hasBase   bool
	baseRaw   string
	base      *rawType
	abstract  bool
	open      bool
	keys      int
	refs      []string
	refSet    map[string]bool
	members   []*member
	namen     map[string]bool
}

type rawBinding struct {
	path   string
	target string
}

type rawSet struct {
	name     string
	typ      string
	ort      string
	bindings []rawBinding
}

type rawContainer struct {
	qn         string
	sets       []*rawSet
	setNamen   map[string]bool
	singletons map[string]bool
}

// dokument sammelt, was das Lesen findet, in der Reihenfolge des Dokuments.
type dokument struct {
	namespaces []string
	abbild     map[string]string
	typen      []*rawType
	index      map[string]*rawType
	deklariert map[string]bool
	container  *rawContainer
}

func neuesDokument() dokument {
	return dokument{abbild: map[string]string{}, index: map[string]*rawType{}, deklariert: map[string]bool{}}
}
