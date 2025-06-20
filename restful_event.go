package record

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"m7s.live/engine/v4/util"
)

var mu sync.Mutex

func errorJsonString(args map[string]interface{}) string {
	resultJsonData := make(map[string]interface{})
	for field, value := range args {
		resultJsonData[field] = value
	}
	jsonString, _ := json.Marshal(resultJsonData)
	return string(jsonString)
}

// 根据事件id拉取录像流
func (conf *RecordConfig) API_event_pull(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	resultJsonData := make(map[string]interface{})
	resultJsonData["code"] = -1
	if token == "" || token != "m7s" {
		resultJsonData["msg"] = "token错误"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	var postData map[string]interface{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resultJsonData["msg"] = "Unable to read request body "
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	err = json.Unmarshal(body, &postData)
	id := int(postData["id"].(float64))
	if id > 0 {
		var eventRecord EventRecord
		result := db.First(&eventRecord, id) // 根据主键查询
		if result.Error != nil {
			log.Println("Error finding eventrecord:", result.Error)
		}
		fileUrlPath := eventRecord.Urlpath
		if fileUrlPath != "" {
			// 判断协议类型
			scheme := "http"
			if r.TLS != nil {
				scheme = "https" // 如果 r.TLS 不为空，则表示是 HTTPS
			}
			// 获取完整请求的地址：包含协议、主机和端口
			requestHost := fmt.Sprintf("%s://%s", scheme, r.Host)
			fileUrlPath = requestHost + "/" + fileUrlPath
			newStreamPath := eventRecord.StreamPath + "/" + strings.TrimSuffix(eventRecord.Filename, ".flv")
			requestFullPath := fmt.Sprintf(requestHost+"/hdl/api/pull?streamPath=%s&target=%s", newStreamPath, fileUrlPath)
			resp, err := http.Get(requestFullPath)
			if err != nil {
				log.Fatal("Error while making GET request:", err)
			}
			defer resp.Body.Close() // 确保函数结束后关闭响应体

			// 读取响应体
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("Error reading response body:", err)
			}

			// 打印响应内容
			fmt.Println("Response Status:", resp.Status)
			fmt.Println("Response Body:", string(body))
			resultJsonData["streamPath"] = newStreamPath
			resultJsonData["code"] = 0
			resultJsonData["msg"] = ""
			util.ReturnError(util.APIErrorNone, errorJsonString(resultJsonData), w, r)
		}
	}
}

// 录像报警列表
func (conf *RecordConfig) API_alarm_list(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	resultJsonData := make(map[string]interface{})
	resultJsonData["code"] = -1
	if token == "" || token != "m7s" {
		resultJsonData["msg"] = "token错误"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	//TODO 用token验证用户信息是否有效，并获取用户信息换取userid
	var postData map[string]interface{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resultJsonData["msg"] = "Unable to read request body "
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	err = json.Unmarshal(body, &postData)
	pageNum := postData["pageNum"].(float64)
	if pageNum <= 0 {
		resultJsonData["msg"] = "pageNum error"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	pageSize := postData["pageSize"].(float64)
	if pageSize <= 0 {
		resultJsonData["msg"] = "pageSize error"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	exceptions, totalCount, err := paginate(db, Exception{}, int(pageNum), int(pageSize), postData)
	if err != nil {
		resultJsonData["msg"] = err.Error()
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	resultJsonData["totalCount"] = totalCount
	resultJsonData["pageNum"] = pageNum
	resultJsonData["pageSize"] = pageSize
	resultJsonData["list"] = exceptions
	resultJsonData["code"] = 0
	resultJsonData["msg"] = ""
	util.ReturnError(util.APIErrorNone, errorJsonString(resultJsonData), w, r)
}

// 事件录像列表
func (conf *RecordConfig) API_event_list(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	resultJsonData := make(map[string]interface{})
	resultJsonData["code"] = -1
	if token == "" || token != "m7s" {
		resultJsonData["msg"] = "token错误"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	//TODO 用token验证用户信息是否有效，并获取用户信息换取userid
	var postData map[string]interface{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resultJsonData["msg"] = "Unable to read request body "
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	err = json.Unmarshal(body, &postData)
	pageNum := postData["pageNum"].(float64)
	if pageNum <= 0 {
		resultJsonData["msg"] = "pageNum error"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	pageSize := postData["pageSize"].(float64)
	if pageSize <= 0 {
		resultJsonData["msg"] = "pageSize error"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	eventRecords, totalCount, err := paginate(db, EventRecord{}, int(pageNum), int(pageSize), postData)
	if err != nil {
		resultJsonData["msg"] = err.Error()
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	resultJsonData["totalCount"] = totalCount
	resultJsonData["pageNum"] = pageNum
	resultJsonData["pageSize"] = pageSize
	resultJsonData["list"] = eventRecords
	resultJsonData["code"] = 0
	resultJsonData["msg"] = ""
	util.ReturnError(util.APIErrorNone, errorJsonString(resultJsonData), w, r)
}

// 事件录像
func (conf *RecordConfig) API_event_start(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	token := r.Header.Get("token")
	resultJsonData := make(map[string]interface{})
	resultJsonData["code"] = -1
	if token == "" || token != "m7s" {
		resultJsonData["msg"] = "token错误"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	//TODO 用token验证用户信息是否有效，并获取用户信息换取userid
	var eventRecordModel EventRecord
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resultJsonData["msg"] = "Unable to read request body "
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	// 解析JSON数据到map
	err = json.Unmarshal(body, &eventRecordModel)
	if err != nil {
		resultJsonData["msg"] = "Invalid JSON format "
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	streamPath := eventRecordModel.StreamPath
	if streamPath == "" {
		resultJsonData["msg"] = "no streamPath"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}

	//TODO 获取到磁盘容量低，磁盘报错的情况下需要报异常，并且根据事件类型做出处理
	// if getDisckException(streamPath) {
	// 	resultJsonData["msg"] = "disk is full"
	// 	util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
	// 	return
	// }

	eventId := eventRecordModel.RecId
	if eventId == "" {
		resultJsonData["msg"] = "no eventId"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	t := eventRecordModel.Type
	if t == "" {
		t = "flv"
	}
	//recordMode := eventRecordModel.RecordMode
	//if recordMode == "" {
	//	resultJsonData["msg"] = "no recordMode"
	//	util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
	//	return
	//}
	eventName := eventRecordModel.EventName
	if eventName == "" {
		resultJsonData["msg"] = "no eventName"
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	beforeDuration := eventRecordModel.BeforeDuration
	if beforeDuration == "" {
		beforeDuration = strconv.Itoa(conf.beforeDuration)
	}
	afterDuration := eventRecordModel.AfterDuration
	if afterDuration == "" {
		afterDuration = strconv.Itoa(conf.afterDuration)
	}
	recordTime := time.Now().Format("2006-01-02 15:04:05")
	fileName := strings.ReplaceAll(streamPath, "/", "-") + "-" + time.Now().Format("2006-01-02-15-04-05")
	startTime := time.Now().Add(-30 * time.Second).Format("2006-01-02 15:04:05")
	endTime := time.Now().Add(30 * time.Second).Format("2006-01-02 15:04:05")
	//切片大小
	fragment := eventRecordModel.Fragment
	//var id string
	irecorder := NewFLVRecorder(EventMode)
	recorder := irecorder.GetRecorder()
	recorder.FileName = fileName
	recorder.append = false
	irecorder.SetId(streamPath)
	filepath := conf.Flv.Path + "/" + streamPath + "/" + fileName + recorder.Ext //录像文件存入的完整路径（相对路径）
	urlpath := "record/" + streamPath + "/" + fileName + recorder.Ext            //网络拉流的地址
	if fragment != "" {
		if f, err := time.ParseDuration(fragment); err == nil {
			recorder.Fragment = f
		}
	} else {
		recorder.Fragment = 0
	}
	found := false
	if recordtmp, ok := conf.recordings.Load(recorder.ID); ok {
		found = true
		irecorder = recordtmp.(*FLVRecorder)
	}
	if found {
		irecorder.UpdateTimeout(30 * time.Second)
	} else {
		err = irecorder.StartWithDynamicTimeout(streamPath, fileName, 30*time.Second)
	}
	if err != nil {
		exceptionChannel <- &Exception{AlarmType: "record", AlarmDesc: "录像失败", StreamPath: streamPath}
		resultJsonData["msg"] = err.Error()
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	var outid uint
	var eventRecord EventRecord
	if found {
		var oldeventRecord EventRecord
		// 定义 User 结构体作为查询条件
		queryRecord := EventRecord{StreamPath: streamPath}
		db.Where(&queryRecord).Order("id DESC").First(&oldeventRecord)
		eventRecord = EventRecord{StreamPath: streamPath, RecId: eventId, RecordMode: "1", EventName: eventName, BeforeDuration: beforeDuration,
			AfterDuration: afterDuration, CreateTime: recordTime, StartTime: startTime, EndTime: endTime, Filepath: oldeventRecord.Filepath, Filename: oldeventRecord.Filename,
			EventDesc: eventRecordModel.EventDesc, Urlpath: oldeventRecord.Urlpath, Type: t}
	} else {
		eventRecord = EventRecord{StreamPath: streamPath, RecId: eventId, RecordMode: "1", EventName: eventName, BeforeDuration: beforeDuration,
			AfterDuration: afterDuration, CreateTime: recordTime, StartTime: startTime, EndTime: endTime, Filepath: filepath, Filename: fileName + recorder.Ext, EventDesc: eventRecordModel.EventDesc, Urlpath: urlpath, Type: t}
	}
	err = db.Omit("id", "fragment", "isDelete").Create(&eventRecord).Error
	outid = eventRecord.Id
	if err != nil {
		exceptionChannel <- &Exception{AlarmType: "record", AlarmDesc: "录像失败", StreamPath: streamPath}

		resultJsonData["msg"] = err.Error()
		util.ReturnError(-1, errorJsonString(resultJsonData), w, r)
		return
	}
	resultJsonData["id"] = outid
	resultJsonData["code"] = 0
	resultJsonData["msg"] = ""
	util.ReturnError(util.APIErrorNone, errorJsonString(resultJsonData), w, r)
}
