package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/featureflags"
	"zion-english/internal/logs"
	"zion-english/internal/notifications"
	"zion-english/internal/storage"
	"zion-english/internal/teacherintrovideo"

	"go.uber.org/zap"
)

type introVideoUploadJobParams struct {
	VideoID          int64
	User             auth.User
	StagedPath       string
	Ext              string
	MimeType         string
	OriginalFilename string
	OriginalFileSize int64
	CompressPreset   string
	Settings         constants.IntroVideoEncodeSettings
}

func runIntroVideoUploadJob(params introVideoUploadJobParams) {
	ctx := context.Background()
	user := params.User
	originalBase := filepath.Base(params.OriginalFilename)

	processed, err := teacherintrovideo.ProcessStagedFile(ctx, params.StagedPath, params.Ext, params.MimeType, params.Settings)
	if err != nil {
		logs.Log().Error("intro video background processing failed",
			zap.Error(err),
			zap.Int64("video_id", params.VideoID),
			zap.Int64("teacher_id", user.ID),
			zap.String("original_filename", originalBase),
			zap.Int64("original_file_size", params.OriginalFileSize),
			zap.String("compress_preset", params.CompressPreset),
			zap.Int("encode_timeout_secs", params.Settings.TimeoutSecs),
			zap.Bool("skip_compress", params.Settings.SkipCompress),
		)
		failIntroVideoUploadJob(ctx, user, params, originalBase, err)
		return
	}
	defer processed.Cleanup()

	uploadFile, err := os.Open(processed.Path)
	if err != nil {
		logs.Log().Error("open processed intro video", zap.Error(err))
		failIntroVideoUploadJob(ctx, user, params, originalBase, err)
		return
	}
	defer uploadFile.Close()

	storedFilename := fmt.Sprintf("%d_%d%s", user.ID, time.Now().UnixNano(), processed.Ext)
	store := storage.Default()
	if err := store.Put(ctx, storage.CategoryIntroVideos, storedFilename, uploadFile, processed.MimeType); err != nil {
		logs.Log().Error("write intro video file", zap.Error(err))
		failIntroVideoUploadJob(ctx, user, params, originalBase, err)
		return
	}

	if err := dbRW.GetQueries().CompleteTeacherIntroVideoUpload(ctx, queries.CompleteTeacherIntroVideoUploadParams{
		StoredFilename:   sql.NullString{String: storedFilename, Valid: true},
		MimeType:         sql.NullString{String: processed.MimeType, Valid: true},
		FileSize:         sql.NullInt64{Int64: processed.Size, Valid: true},
		OriginalFileSize: sql.NullInt64{Int64: processed.OriginalSize, Valid: true},
		ID:               params.VideoID,
	}); err != nil {
		_ = store.Delete(ctx, storage.CategoryIntroVideos, storedFilename)
		logs.Log().Error("complete teacher intro video upload", zap.Error(err), zap.Int64("video_id", params.VideoID))
		failIntroVideoUploadJob(ctx, user, params, originalBase, err)
		return
	}

	insertUploadLog(ctx, user, uploadLogEntry{
		Module:         "profile",
		Outcome:        constants.UploadLogOutcomeSucceeded,
		Kind:           constants.UploadLogKindIntroVideo,
		Summary:        fmt.Sprintf("intro video file for teacher '%s' (id %d), file '%s'", user.Name, user.ID, originalBase),
		Filename:       originalBase,
		FileSize:       processed.OriginalSize,
		FileSizeValid:  true,
		CompressPreset: params.CompressPreset,
	})
	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("submitted intro video file for teacher '%s'", user.Name))
	notifyTeacher(ctx, user.ID, user.Name, notifications.SystemUser(), notifications.KindIntroVideoProcessed,
		"Your intro video was submitted and is pending administrator review.", "")
	notifySuperuser(ctx, user, notifications.KindIntroVideoSubmitted,
		fmt.Sprintf("Teacher '%s' submitted an intro video file", user.Name), "")
}

func failIntroVideoUploadJob(ctx context.Context, user auth.User, params introVideoUploadJobParams, originalBase string, err error) {
	insertUploadLog(ctx, user, uploadLogEntry{
		Module:         "profile",
		Outcome:        constants.UploadLogOutcomeFailed,
		Kind:           constants.UploadLogKindIntroVideo,
		Summary:        fmt.Sprintf("intro video processing failed for teacher '%s' (id %d), file '%s': %v", user.Name, user.ID, originalBase, err),
		Filename:       originalBase,
		FileSize:       params.OriginalFileSize,
		FileSizeValid:  params.OriginalFileSize > 0,
		CompressPreset: params.CompressPreset,
	})
	if delErr := dbRW.GetQueries().DeleteTeacherIntroVideoByID(ctx, params.VideoID); delErr != nil {
		logs.Log().Error("delete failed intro video processing row", zap.Error(delErr), zap.Int64("video_id", params.VideoID))
	}
	notifyTeacher(ctx, user.ID, user.Name, notifications.SystemUser(), notifications.KindIntroVideoFailed,
		fmt.Sprintf("Intro video upload failed: %s", introVideoSubmitErrorMessage(err)), "")
}

func startIntroVideoUploadJob(params introVideoUploadJobParams) {
	go runIntroVideoUploadJob(params)
}

func introVideoEncodeSettings(ctx context.Context) constants.IntroVideoEncodeSettings {
	return featureflags.IntroVideoCompressSettings(ctx, dbRO)
}
