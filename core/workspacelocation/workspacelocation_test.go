package workspacelocation

import "testing"

func TestGraph(t *testing.T) {
	got := Graph("src-id", "graph-id", "SRC-1", "1881 Census")
	want := Location{
		Section: Sources, SourceID: "src-id", GraphID: "graph-id",
		SourceSurface: SurfaceGraph, Ref: "SRC-1",
		Title: "1881 Census", SourceTitle: "1881 Census",
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if untitled := Graph("src-id", "graph-id", "SRC-1", ""); untitled.Title != "SRC-1" || untitled.SourceTitle != "" {
		t.Fatalf("untitled Source: %+v", untitled)
	}
}
