// Package depictionvaluetest holds depiction documents for test fixtures.
package depictionvaluetest

const (
	// Locator is a Citation locator that cites page 1.
	Locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`
	// Region is a depiction region on the page Locator cites.
	Region = `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`
)
