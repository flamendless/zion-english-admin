-- name: InsertStudentRelationship :exec
INSERT INTO tbl_student_relationships (student_id, related_student_id, relationship)
VALUES (?, ?, ?);

-- name: GetRelationshipsByStudentID :many
SELECT sr.student_id, sr.relationship, rs.name AS related_student_name
FROM tbl_student_relationships sr
INNER JOIN tbl_students rs ON sr.related_student_id = rs.id
WHERE sr.student_id = ?
ORDER BY rs.name ASC;

-- name: GetAllStudentRelationships :many
SELECT sr.student_id, sr.relationship, rs.name AS related_student_name
FROM tbl_student_relationships sr
INNER JOIN tbl_students rs ON sr.related_student_id = rs.id
ORDER BY rs.name ASC;
