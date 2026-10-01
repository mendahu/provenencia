package deleteimpact

// Reserved inbound vias (tables not in the catalog yet). Count/List return 0
// so Observation Impact is ready for Claims / Narrative without a schema change.
const (
	ViaReconciliationEvidence = "reconciliation_claim_evidence.observation_id"
	ViaNarrativeTarget        = "narrative.target"
)
