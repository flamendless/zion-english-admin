-- name: GetReportTeacherSummaries :many
SELECT
    t.id AS teacher_id,
    t.name AS teacher_name,
    COUNT(cr.id) AS total_classes,
    COALESCE(SUM(CASE WHEN cr.status = 'conducted' THEN 1 ELSE 0 END), 0) AS conducted_classes
FROM tbl_teachers t
LEFT JOIN tbl_class_records cr ON cr.teacher_id = t.id
    AND cr.date >= ? AND cr.date <= ?
WHERE t.status = 'approved' AND t.deleted = 0
  AND (
    ? = ''
    OR t.name LIKE '%' || ? || '%'
    OR EXISTS (
      SELECT 1 FROM tbl_class_records cr2
      JOIN tbl_students s ON cr2.student_id = s.id
      WHERE cr2.teacher_id = t.id
        AND cr2.date >= ? AND cr2.date <= ?
        AND s.name LIKE '%' || ? || '%'
    )
  )
GROUP BY t.id, t.name
ORDER BY t.name ASC;

-- name: GetReportTeacherEarnings :many
SELECT cr.teacher_id, cr.currency, COALESCE(SUM(cr.rate), 0) AS total_rate
FROM tbl_class_records cr
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.date >= ? AND cr.date <= ?
  AND cr.status = 'conducted'
  AND t.status = 'approved' AND t.deleted = 0
  AND (
    ? = ''
    OR t.name LIKE '%' || ? || '%'
    OR EXISTS (
      SELECT 1 FROM tbl_class_records cr2
      JOIN tbl_students s ON cr2.student_id = s.id
      WHERE cr2.teacher_id = t.id
        AND cr2.date >= ? AND cr2.date <= ?
        AND s.name LIKE '%' || ? || '%'
    )
  )
GROUP BY cr.teacher_id, cr.currency;

-- name: GetTeacherReportClassRecords :many
SELECT cr.id, cr.student_id, cr.teacher_id, cr.date, cr.start_time, cr.end_time,
       cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.reason, cr.notes,
       cr.created_at, cr.updated_at, cr.recorded_by_role,
       s.name AS student_name, t.name AS teacher_name
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.teacher_id = ? AND cr.date >= ? AND cr.date <= ?
ORDER BY cr.date ASC, s.name ASC, cr.rate ASC, cr.start_time ASC;

-- name: GetClassRecordFingerprintRows :many
SELECT cr.id, cr.student_id, cr.date, cr.start_time, cr.end_time,
       cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.updated_at
FROM tbl_class_records cr
WHERE cr.teacher_id = ? AND cr.date >= ? AND cr.date <= ?
ORDER BY cr.id ASC;

-- name: GetClassRecordFingerprintRowsForRange :many
SELECT cr.id, cr.teacher_id, cr.student_id, cr.date, cr.start_time, cr.end_time,
       cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.updated_at
FROM tbl_class_records cr
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.date >= ? AND cr.date <= ?
  AND t.status = 'approved' AND t.deleted = 0
ORDER BY cr.teacher_id ASC, cr.id ASC;

-- name: GetReportGeneration :one
SELECT id, teacher_id, start_date, end_date, content_hash, output_path, record_count, generated_at
FROM tbl_report_generations
WHERE teacher_id = ? AND start_date = ? AND end_date = ?;

-- name: GetReportGenerationsForRange :many
SELECT id, teacher_id, start_date, end_date, content_hash, output_path, record_count, generated_at
FROM tbl_report_generations
WHERE start_date = ? AND end_date = ?;

-- name: UpsertReportGeneration :exec
INSERT INTO tbl_report_generations (teacher_id, start_date, end_date, content_hash, output_path, record_count, generated_at)
VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
ON CONFLICT(teacher_id, start_date, end_date) DO UPDATE SET
  content_hash = excluded.content_hash,
  output_path = excluded.output_path,
  record_count = excluded.record_count,
  generated_at = datetime('now');

-- name: GetReportGenerationByOutputBasename :one
SELECT rg.id, rg.teacher_id, rg.start_date, rg.end_date, rg.content_hash, rg.output_path, rg.record_count, rg.generated_at,
       t.name AS teacher_name
FROM tbl_report_generations rg
JOIN tbl_teachers t ON rg.teacher_id = t.id
WHERE rg.output_path = ?;
