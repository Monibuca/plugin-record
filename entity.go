package record

import "time"

// mysql数据库eventrecord表
type EventRecord struct {
	Id             uint   `json:"id" desc:"自增长id" gorm:"primaryKey;autoIncrement"`
	StreamPath     string `json:"streamPath" desc:"流路径" gorm:"type:varchar(255);comment:流路径"`
	RecId          string `json:"RecId" desc:"录制编号" gorm:"type:varchar(255);comment:录制编号"`
	RecordMode     string `json:"recordMode" desc:"事件类型,0=连续录像模式，1=事件录像模式" gorm:"type:varchar(255);comment:事件类型,0=连续录像模式，1=事件录像模式"`
	EventName      string `json:"eventName" desc:"事件名称" gorm:"type:varchar(255);comment:事件名称"`
	BeforeDuration string `json:"beforeDuration" desc:"事件前缓存时长" gorm:"type:varchar(255);comment:事件前缓存时长"`
	AfterDuration  string `json:"afterDuration" desc:"事件后缓存时长" gorm:"type:varchar(255);comment:事件后缓存时长"`
	CreateTime     string `json:"createTime" desc:"录像时间" gorm:"type:varchar(255);comment:录像时间"`
	StartTime      string `json:"startTime" desc:"录像开始时间" gorm:"type:varchar(255);comment:录像开始时间"`
	EndTime        string `json:"endTime" desc:"录像结束时间" gorm:"type:varchar(255);comment:录像结束时间"`
	Filepath       string `json:"filePath" desc:"录像文件物理路径" gorm:"type:varchar(255);comment:录像文件物理路径"`
	Urlpath        string `json:"urlPath" desc:"录像文件下载URL路径" gorm:"type:varchar(255);comment:录像文件下载URL路径"`
	IsDelete       string `json:"isDelete" desc:"是否删除，0表示正常，1表示删除，默认0" gorm:"type:varchar(255);default:'0';comment:是否删除，0表示正常，1表示删除，默认0"`
	UserId         string `json:"useId" desc:"用户id" gorm:"type:varchar(255);comment:用户id"`
	Filename       string `json:"fileName" desc:"文件名" gorm:"type:varchar(255);comment:文件名"`
	Fragment       string `json:"fragment" desc:"切片大小" gorm:"type:varchar(255);comment:切片大小;default:'0'"`
	EventDesc      string `json:"eventDesc" desc:"事件描述" gorm:"type:varchar(255);comment:事件描述"`
	Type           string `json:"type" desc:"录像文件类型" gorm:"type:varchar(255);comment:录像文件类型,flv,mp4,raw,fmp4,hls"`
	EventLevel     string `json:"eventLevel" desc:"事件级别" gorm:"type:varchar(255);comment:事件级别,0表示重要事件，无法删除且表示无需自动删除,1表示非重要事件,达到自动删除时间后，自动删除;default:'1'"`
}

//// TableName 返回自定义的表名
//func (EventRecord) TableName() string {
//	return "eventrecord"
//}

// mysql数据库里Exception 定义异常结构体
type Exception struct {
	CreateTime string `json:"createTime" gorm:"type:varchar(50)"`
	AlarmType  string `json:"alarmType" gorm:"type:varchar(50)"`
	AlarmDesc  string `json:"alarmDesc" gorm:"type:varchar(50)"`
	ServerIP   string `json:"serverIP" gorm:"type:varchar(50)"`
	StreamPath string `json:"streamPath" gorm:"type:varchar(50)"`
	FileName   string `json:"fileName" gorm:"type:varchar(100)"`
}

// sqlite数据库用来存放每个flv文件的关键帧对应的offset及abstime数据
type FLVKeyframe struct {
	FLVFileName  string    `gorm:"not null"`
	FrameOffset  int64     `gorm:"not null"`
	FrameAbstime uint32    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
