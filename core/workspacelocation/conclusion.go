package workspacelocation

import "github.com/mendahu/provenencia/core/vocabkey"

// Conclusion maps a primary conclusion type key to a Person, Event or Place
// workspace location. Association types return false.
func Conclusion(typeKey, entityID, ref, title string) (Location, bool) {
	switch typeKey {
	case vocabkey.Person:
		return Location{Section: Persons, EntityID: entityID, Ref: ref, Title: title}, true
	case vocabkey.Event:
		return Location{Section: Events, EntityID: entityID, Ref: ref, Title: title}, true
	case vocabkey.Place:
		return Location{Section: Places, EntityID: entityID, Ref: ref, Title: title}, true
	default:
		return Location{}, false
	}
}
