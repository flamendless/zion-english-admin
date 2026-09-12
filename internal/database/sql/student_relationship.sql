-- name: InsertStudentRelationship :exec
INSERT INTO tbl_student_relationships (student_id, related_student_id, relationship)
VALUES (?, ?, ?);

-- name: GetRelationshipsByStudentID :many
SELECT sr.student_id, sr.related_student_id, sr.relationship, rs.name AS related_student_name
FROM tbl_student_relationships sr
INNER JOIN tbl_students rs ON sr.related_student_id = rs.id
WHERE sr.student_id = ?
ORDER BY rs.name ASC;

-- name: GetAllStudentRelationships :many
SELECT sr.student_id, sr.relationship, rs.name AS related_student_name
FROM tbl_student_relationships sr
INNER JOIN tbl_students rs ON sr.related_student_id = rs.id
ORDER BY rs.name ASC;

-- name: DeleteStudentRelationshipsByStudentID :exec
DELETE FROM tbl_student_relationships WHERE student_id = ?;

-- name: DeleteStudentRelationship :exec
DELETE FROM tbl_student_relationships
WHERE student_id = ? AND related_student_id = ?;

-- name: GetStudentRelationshipEdges :many
SELECT
	sr.student_id,
	s.name AS student_name,
	sr.related_student_id,
	rs.name AS related_student_name,
	sr.relationship
FROM tbl_student_relationships sr
INNER JOIN tbl_students s ON s.id = sr.student_id AND s.status != 'deleted'
INNER JOIN tbl_students rs ON rs.id = sr.related_student_id AND rs.status != 'deleted'
ORDER BY s.name ASC, rs.name ASC;

-- name: GetStudentsForRelationshipGraph :many
SELECT id, name, parent_name, assigned_color, status
FROM tbl_students
WHERE status != 'deleted'
	AND (
		TRIM(COALESCE(parent_name, '')) != ''
		OR id IN (SELECT student_id FROM tbl_student_relationships)
		OR id IN (SELECT related_student_id FROM tbl_student_relationships)
	)
ORDER BY name ASC;
