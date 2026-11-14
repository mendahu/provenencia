package shortlist

import "unsafe"

// FiledKeysShared reports whether every key a handle is filed under points at
// the bytes of the posting map's own key, not at a copy.
func (x *Index) FiledKeysShared() bool {
	for _, f := range x.filed {
		for _, k := range f.keys {
			shared := false
			for mk := range x.byKey[f.kind] {
				if mk == k && unsafe.StringData(mk) == unsafe.StringData(k) {
					shared = true
				}
			}
			if !shared {
				return false
			}
		}
	}
	return true
}
