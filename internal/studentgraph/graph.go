package studentgraph

import (
	"fmt"
	"strings"

	"zion-english/internal/constants"
)

const (
	NodeKindStudent = "student"
	NodeKindParent  = "parent"

	ParentEdgeLabel = "child"
)

type StudentRow struct {
	ID            int64
	Name          string
	ParentName    string
	AssignedColor string
	Status        string
}

type EdgeRow struct {
	StudentID          int64
	RelatedStudentID   int64
	Relationship       string
}

type Node struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Color   string `json:"color,omitempty"`
	Status  string `json:"status,omitempty"`
	ViewURL string `json:"viewUrl,omitempty"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

type BuildOptions struct {
	StudentViewURL func(studentID int64) string
}

func Build(students []StudentRow, edges []EdgeRow, opts BuildOptions) Graph {
	if opts.StudentViewURL == nil {
		opts.StudentViewURL = func(studentID int64) string {
			return fmt.Sprintf("/students/%d/view", studentID)
		}
	}

	nodes := make([]Node, 0, len(students))
	links := make([]Link, 0, len(edges)+len(students))
	nodeIndex := make(map[string]int, len(students))

	for _, student := range students {
		id := studentNodeID(student.ID)
		color := strings.TrimSpace(student.AssignedColor)
		if color == "" {
			color = constants.DefaultAssignedColor
		}
		nodes = append(nodes, Node{
			ID:      id,
			Name:    student.Name,
			Kind:    NodeKindStudent,
			Color:   color,
			Status:  student.Status,
			ViewURL: opts.StudentViewURL(student.ID),
		})
		nodeIndex[id] = len(nodes) - 1
	}

	parentGroups := make(map[string][]StudentRow)
	for _, student := range students {
		parentName := strings.TrimSpace(student.ParentName)
		if parentName == "" {
			continue
		}
		key := parentKey(parentName)
		parentGroups[key] = append(parentGroups[key], student)
	}

	for key, group := range parentGroups {
		parentName := strings.TrimSpace(group[0].ParentName)
		parentID := parentNodeID(key)
		if _, ok := nodeIndex[parentID]; !ok {
			nodes = append(nodes, Node{
				ID:   parentID,
				Name: parentName,
				Kind: NodeKindParent,
			})
			nodeIndex[parentID] = len(nodes) - 1
		}
		for _, student := range group {
			links = append(links, Link{
				Source: parentID,
				Target: studentNodeID(student.ID),
				Label:  ParentEdgeLabel,
			})
		}
	}

	for _, edge := range edges {
		sourceID := studentNodeID(edge.StudentID)
		targetID := studentNodeID(edge.RelatedStudentID)
		if _, ok := nodeIndex[sourceID]; !ok {
			continue
		}
		if _, ok := nodeIndex[targetID]; !ok {
			continue
		}
		links = append(links, Link{
			Source: sourceID,
			Target: targetID,
			Label:  strings.TrimSpace(edge.Relationship),
		})
	}

	return Graph{
		Nodes: nodes,
		Links: links,
	}
}

func studentNodeID(id int64) string {
	return fmt.Sprintf("%d", id)
}

func parentNodeID(key string) string {
	return "parent:" + key
}

func parentKey(parentName string) string {
	return strings.ToLower(strings.TrimSpace(parentName))
}
