package studentgraph

import (
	"testing"
)

func TestBuildParentGroupingAndRelationshipEdges(t *testing.T) {
	students := []StudentRow{
		{ID: 1, Name: "Jane", ParentName: "Smith", AssignedColor: "#90C020", Status: "active"},
		{ID: 2, Name: "John", ParentName: "Smith", AssignedColor: "#B9D283", Status: "active"},
		{ID: 3, Name: "Alex", ParentName: "", AssignedColor: "#90C020", Status: "inactive"},
	}
	edges := []EdgeRow{
		{StudentID: 1, RelatedStudentID: 2, Relationship: "siblings"},
		{StudentID: 3, RelatedStudentID: 1, Relationship: "cousin"},
	}

	graph := Build(students, edges, BuildOptions{
		StudentViewURL: func(studentID int64) string {
			return "/students/" + studentNodeID(studentID) + "/view"
		},
	})

	if len(graph.Nodes) != 4 {
		t.Fatalf("expected 4 nodes (3 students + 1 parent), got %d", len(graph.Nodes))
	}

	parentID := parentNodeID(parentKey("Smith"))
	foundParent := false
	for _, node := range graph.Nodes {
		if node.ID == parentID {
			foundParent = true
			if node.Kind != NodeKindParent || node.Name != "Smith" {
				t.Fatalf("unexpected parent node: %+v", node)
			}
		}
		if node.Kind == NodeKindStudent && node.ViewURL == "" {
			t.Fatalf("student node missing view URL: %+v", node)
		}
	}
	if !foundParent {
		t.Fatal("parent node not found")
	}

	if len(graph.Links) != 4 {
		t.Fatalf("expected 4 links (2 parent-child + 2 relationship), got %d", len(graph.Links))
	}

	var parentLinks, relLinks int
	for _, link := range graph.Links {
		switch {
		case link.Source == parentID && link.Label == ParentEdgeLabel:
			parentLinks++
		case link.Source == "1" && link.Target == "2" && link.Label == "siblings":
			relLinks++
		case link.Source == "3" && link.Target == "1" && link.Label == "cousin":
			relLinks++
		default:
			t.Fatalf("unexpected link: %+v", link)
		}
	}
	if parentLinks != 2 || relLinks != 2 {
		t.Fatalf("unexpected link counts: parent=%d rel=%d", parentLinks, relLinks)
	}
}

func TestBuildSkipsEdgesForMissingStudents(t *testing.T) {
	students := []StudentRow{
		{ID: 1, Name: "Jane", Status: "active"},
	}
	edges := []EdgeRow{
		{StudentID: 1, RelatedStudentID: 99, Relationship: "siblings"},
	}

	graph := Build(students, edges, BuildOptions{})
	if len(graph.Links) != 0 {
		t.Fatalf("expected dangling edge to be skipped, got %d links", len(graph.Links))
	}
}

func TestBuildUsesDefaultColorWhenMissing(t *testing.T) {
	graph := Build([]StudentRow{
		{ID: 1, Name: "Jane", Status: "active"},
	}, nil, BuildOptions{})

	if graph.Nodes[0].Color != "#90C020" {
		t.Fatalf("expected default color, got %s", graph.Nodes[0].Color)
	}
}

func TestParentKeyIsCaseInsensitive(t *testing.T) {
	students := []StudentRow{
		{ID: 1, Name: "Jane", ParentName: "Smith", Status: "active"},
		{ID: 2, Name: "John", ParentName: "smith", Status: "active"},
	}

	graph := Build(students, nil, BuildOptions{})
	parentCount := 0
	for _, node := range graph.Nodes {
		if node.Kind == NodeKindParent {
			parentCount++
		}
	}
	if parentCount != 1 {
		t.Fatalf("expected one parent node for case-insensitive names, got %d", parentCount)
	}
}
