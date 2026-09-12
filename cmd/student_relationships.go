package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"zion-english/frontend"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/models"
	"zion-english/internal/studentgraph"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

type studentRelationshipGraphResponse struct {
	Nodes []studentgraph.Node `json:"nodes"`
	Links []studentgraph.Link `json:"links"`
}

func handleStudentRelationships(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.StudentRelationships(frontend.StudentRelationshipsData{
		GraphAPIURL:      utils.URL("/api/student-relationships/graph"),
		StudentSearchURL: utils.URL("/student-relationships/partials/student-search"),
	}).Render(r.Context(), w); err != nil {
		logs.Log().Error("render student relationships", zap.Error(err))
	}
}

func handleGetStudentRelationshipGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	q := dbRO.GetQueries()

	studentRows, err := q.GetStudentsForRelationshipGraph(ctx)
	if err != nil {
		logs.Log().Error("fetch students for relationship graph", zap.Error(err))
		HttpError(w, "Failed to load student data", http.StatusInternalServerError)
		return
	}

	edgeRows, err := q.GetStudentRelationshipEdges(ctx)
	if err != nil {
		logs.Log().Error("fetch student relationship edges", zap.Error(err))
		HttpError(w, "Failed to load relationship data", http.StatusInternalServerError)
		return
	}

	graph := studentgraph.Build(mapStudentsForGraph(studentRows), mapEdgesForGraph(edgeRows), studentgraph.BuildOptions{
		StudentViewURL: func(studentID int64) string {
			return utils.URL(fmt.Sprintf("/students/%d/view", studentID))
		},
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(studentRelationshipGraphResponse{
		Nodes: graph.Nodes,
		Links: graph.Links,
	}); err != nil {
		logs.Log().Error("encode student relationship graph", zap.Error(err))
	}
}

func handleStudentRelationshipStudentSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := firstQueryParam(r, "studentQ", "q")
	w.Header().Set("Content-Type", "text/html")
	if q == "" {
		frontend.StudentDiagramSearchResults(nil).Render(r.Context(), w)
		return
	}

	students, err := dbRO.GetQueries().SearchStudentsByName(r.Context(), sql.NullString{String: q, Valid: true})
	if err != nil {
		HttpError(w, "Failed to search students", http.StatusInternalServerError)
		return
	}

	var studentResponses []models.StudentAPIResponse
	for _, s := range students {
		studentResponses = append(studentResponses, models.StudentAPIResponse{
			ID:           s.ID,
			Name:         s.Name,
			Currency:     s.Currency,
			RatePerClass: s.RatePerClass,
		})
	}

	if err := frontend.StudentDiagramSearchResults(studentResponses).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func mapStudentsForGraph(rows []queries.GetStudentsForRelationshipGraphRow) []studentgraph.StudentRow {
	students := make([]studentgraph.StudentRow, len(rows))
	for i, row := range rows {
		students[i] = studentgraph.StudentRow{
			ID:            row.ID,
			Name:          row.Name,
			ParentName:    nullStringValue(row.ParentName),
			AssignedColor: row.AssignedColor,
			Status:        row.Status,
		}
	}
	return students
}

func mapEdgesForGraph(rows []queries.GetStudentRelationshipEdgesRow) []studentgraph.EdgeRow {
	edges := make([]studentgraph.EdgeRow, len(rows))
	for i, row := range rows {
		edges[i] = studentgraph.EdgeRow{
			StudentID:        row.StudentID,
			RelatedStudentID: row.RelatedStudentID,
			Relationship:     nullStringValue(row.Relationship),
		}
	}
	return edges
}
