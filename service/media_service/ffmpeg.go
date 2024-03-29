package media_service

import (
	"fmt"
	"os"
	"os/exec"

	"lucy/models"
	"lucy/pkg/log"
	"lucy/pkg/setting"
)

const LiveRecordPath = "live-record"

// ffmpegRecord use ffmpeg to record live video
func ffmpegRecord(streamUrl string, liveName string) {
	go func() {
		owner, token, ok := ParseLiveName(liveName)
		if !ok {
			return
		}
		ownerDir := fmt.Sprintf("%s/%s/%s", setting.AppSetting.AppRoot, LiveRecordPath, owner)
		os.MkdirAll(ownerDir, os.ModePerm)
		outFileName := token + ".mp4"
		relativePath := fmt.Sprintf("%s/%s/%s", LiveRecordPath, owner, outFileName)
		localFilePath := fmt.Sprintf("%s/%s", ownerDir, outFileName)
		log.Debug("ffmpegRecord", "outFileName", outFileName,
			"relativePath", relativePath, "localFilePath", localFilePath)
		cmd := exec.Command("ffmpeg", "-i", streamUrl, localFilePath, "-y")
		err := cmd.Run()
		l := models.Live{}
		if err != nil {
			log.Error("ffmpeg record failed", "err", err.Error())
			err = models.Db().Model(&l).Where(models.Live{Name: liveName}).
				Update("record_status", noRecord).Error
			if err != nil {
				log.Debug("update database failed", "err", err)
			}
			return
		}
		err = models.Db().Model(&l).Where(models.Live{Name: liveName}).
			Updates(models.Live{
				RecordStatus: available,
				RecordPath:   relativePath,
			}).Error
		if err != nil {
			log.Warn("update database failed", "err", err)
		} else {
			log.Info("record success", "streamUrl", streamUrl,
				"liveName", liveName, "recordPath", localFilePath)
		}
	}()
}
