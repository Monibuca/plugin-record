package record

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// sqlite数据库初始化，用来存放视频的关键帧等信息
func initSqliteDB(sqliteDbPath string) *gorm.DB {
	// 打开数据库连接
	sqlitedb, err := gorm.Open(sqlite.Open(sqliteDbPath), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	err = sqlitedb.AutoMigrate(&FLVKeyframe{})
	err = sqlitedb.AutoMigrate(&EventRecord{})
	err = sqlitedb.AutoMigrate(&Exception{})
	if err != nil {
		log.Fatal(err)
	}
	return sqlitedb
}

// 保存Recorder到数据库中
func (r *Recorder) SaveToDB() {
	startTime := time.Now().Format("2006-01-02 15:04:05")
	endTime := time.Now().Format("2006-01-02 15:04:05")
	fileName := r.FileName
	if r.FileName == "" {
		fileName = strings.ReplaceAll(r.Stream.Path, "/", "-") + "-" + time.Now().Format("2006-01-02-15-04-05")
	}
	filepath := RecordPluginConfig.Mp4.Path + "/" + r.Stream.Path + "/" + fileName + r.Ext //录像文件存入的完整路径（相对路径）
	eventRecord := EventRecord{StreamPath: r.Stream.Path, RecId: r.ID, RecordMode: "0", BeforeDuration: "0",
		AfterDuration: fmt.Sprintf("%.0f", r.Fragment.Seconds()), CreateTime: startTime, StartTime: startTime,
		EndTime: endTime, Filepath: filepath, Filename: fileName + r.Ext, Urlpath: "record/" + strings.ReplaceAll(r.filePath, "\\", "/"), Fragment: fmt.Sprintf("%.0f", r.Fragment.Seconds()), Type: r.Ext}
	err = db.Create(&eventRecord).Error
	if err != nil {
		r.Error("save to db error", zap.Error(err))
	} else {
		r.Info("save to db success", zap.String("filepath", filepath))
	}
}

// 更新录像文件表中记录，包括录像文件的大小、结束时间以及录像状态
func (r *Recorder) RemoveRecordById() {
	endTime := time.Now().Format("2006-01-02 15:04:05")
	fileName := r.FileName
	if r.FileName == "" {
		fileName = strings.ReplaceAll(r.Stream.Path, "/", "-") + "-" + time.Now().Format("2006-01-02-15-04-05")
	}
	filepath := RecordPluginConfig.Mp4.Path + "/" + r.Stream.Path + "/" + fileName + r.Ext //录像文件存入的完整路径（相对路径）

	eventRecord := EventRecord{Filepath: filepath, RecordMode: "1", BeforeDuration: "0",
		AfterDuration: fmt.Sprintf("%.0f", r.Fragment.Seconds()),
		EndTime:       endTime, IsDelete: "1"}

	err = db.Where("filepath = ? AND is_delete = ?", filepath, "0").Updates(eventRecord).Error
	if err != nil {
		r.Error("update to db error", zap.Error(err))
	} else {
		r.Info("update to db success", zap.String("filepath", filepath))
	}
}
