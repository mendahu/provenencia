package audit

// sqlInsertScope stores one scope row. Record calls it for each id of each
// effects.Scopes entry; the scope types themselves are declared in effects.
const sqlInsertScope = `INSERT OR IGNORE INTO audit_transaction_scopes (audit_transaction_id, scope_type, scope_id)
	VALUES (?, ?, ?)`
