package record

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// 查询所有正在录制中的记录
func (conf *RecordConfig) CheckRecordDB() {
	go func() {
		for {
			plugin.Info("启动录制检查器，开始检查未在录制的视频流信息：", zap.Time("start", time.Now()))
			var outRecordings []any
			var eventRecords []EventRecord

			var recordings []any
			conf.recordings.Range(func(key, value any) bool {
				recordings = append(recordings, value)
				return true
			})
			err = db.Where("is_delete = ?", "0").Find(&eventRecords).Error
			if err != nil {
				plugin.Error("查询数据库失败 %s", zap.Error(err))
				return
			} else {
				plugin.Info("DB中处于录制中的数量：", zap.Int("recordings", len(eventRecords)))
				// 查询出在eventRecords中fileName字段包含recordings中的filename的记录
				// 遍历eventRecords，判断fileName字段是否包含recordings中的filename
				for _, record := range eventRecords {
					plugin.Info("DB中处于录制中的流内容：", zap.String("recording-Id", record.RecId), zap.String("record-Filepath", record.Filepath), zap.String("record-Filename", record.Filename))
					// 如果 recordings 为空，将所有的录制流都添加到 outRecordings
					if len(recordings) == 0 {
						outRecordings = append(outRecordings, record)
					} else {
						// 遍历 recordings，判断是否有包含当前 record.Filename 的录制流
						found := false
						for _, recording := range recordings {
							if strings.Contains(recording.(IRecorder).GetRecorder().ID, record.RecId) {
								found = true
								break
							}
						}
						// 如果没有找到，则添加到 outRecordings
						if !found {
							outRecordings = append(outRecordings, record)
						}
					}
				}
				// 打印未在录制的视频流
				for _, outRecording := range outRecordings {
					streamPath := outRecording.(EventRecord).StreamPath
					fileName := outRecording.(EventRecord).Filename
					plugin.Info("已停止录制的视频流", zap.Any("StreamPath", streamPath), zap.Any("Filename", fileName))

					exception := Exception{AlarmType: "Recording-Stop", AlarmDesc: "Recording Stopped Exception", StreamPath: streamPath, FileName: fileName}
					callback(&exception)
				}
			}

			// 等待 1 分钟后继续执行
			<-time.After(60 * time.Second)
		}
	}()

}

// 向第三方发送异常报警
func callback(exception *Exception) {
	exception.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	exception.ServerIP = RecordPluginConfig.LocalIp
	data, err := json.Marshal(exception)
	if err != nil {
		fmt.Println("Error marshalling exception:", err)
		return
	}
	err = db.Create(&exception).Error
	if err != nil {
		fmt.Println("异常数据插入数据库失败:", err)
		return
	}
	resp, err := http.Post(RecordPluginConfig.ExceptionPostUrl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error sending exception to third party API:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to send exception, status code:", resp.StatusCode)
	} else {
		fmt.Println("Exception sent successfully!")
	}
}
