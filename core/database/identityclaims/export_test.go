package identityclaims

// BridgeHeadsSQL, AssociationSQL and the conds expose the bridge queries to the plan tests.
var (
	BridgeHeadsSQL = bridgeHeadsSQL
	FileCond       = fileCond
	GraphCond      = graphCond
	UnfileCond     = touchesEnd
)

// AssociationSQL is associationSQL for ends named by property key and entity id.
func AssociationSQL(typeKey string, propertyKeys []string, entityID []byte, termKey string, termID []byte) (string, []any) {
	var ends []endpoint
	for _, k := range propertyKeys {
		ends = append(ends, endpoint{propertyKey: k, entityID: entityID})
	}
	return associationSQL(typeKey, ends, termKey, termID)
}
