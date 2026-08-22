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
