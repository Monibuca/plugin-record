package record

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
)

type RecordConfig struct {
	config.Subscribe
	config.HTTP
	Flv                         Record `desc:"flv录制配置"`
	Mp4                         Record `desc:"mp4录制配置"`
	Fmp4                        Record `desc:"fmp4录制配置"`
	Hls                         Record `desc:"hls录制配置"`
	Raw                         Record `desc:"视频裸流录制配置"`
	RawAudio                    Record `desc:"音频裸流录制配置"`
	recordings                  sync.Map
	beforeDuration              int           `desc:"事件前缓存时长"`
	afterDuration               int           `desc:"事件后缓存时长"`
	MysqlDSN                    string        `desc:"mysql数据库连接字符串"`
	ExceptionPostUrl            string        `desc:"第三方异常上报地址"`
	SqliteDbPath                string        `desc:"sqlite数据库路径"`
	DiskMaxPercent              float64       `desc:"硬盘使用百分之上限值，超过后报警"`
	LocalIp                     string        `desc:"本机IP"`
	RecordFileExpireDays        int           `desc:"录像自动删除的天数,0或未设置表示不自动删除"`
	RecordPathNotShowStreamPath bool          `desc:"录像路径中是否包含streamPath，默认true"`
	Storage                     StorageConfig `desc:"MINIO 配置"`
}

//go:embed default.yaml
var defaultYaml DefaultYaml
var ErrRecordExist = errors.New("recorder exist")
var RecordPluginConfig = &RecordConfig{
	Flv: Record{
		Path:          "record/flv",
		Ext:           ".flv",
		GetDurationFn: getFLVDuration,
	},
	Fmp4: Record{
		Path: "record/fmp4",
		Ext:  ".mp4",
	},
	Mp4: Record{
		Path: "record/mp4",
		Ext:  ".mp4",
	},
	Hls: Record{
		Path: "record/hls",
		Ext:  ".m3u8",
	},
	Raw: Record{
		Path: "record/raw",
		Ext:  ".", // 默认h264扩展名为.h264,h265扩展名为.h265
	},
	RawAudio: Record{
		Path: "record/raw",
		Ext:  ".", // 默认aac扩展名为.aac,pcma扩展名为.pcma,pcmu扩展名为.pcmu
	},
	beforeDuration:              30,
	afterDuration:               30,
	MysqlDSN:                    "",
	ExceptionPostUrl:            "http://www.163.com",
	SqliteDbPath:                "./m7sv4.db",
	DiskMaxPercent:              80.00,
	LocalIp:                     getLocalIP(),
	RecordFileExpireDays:        0,
	RecordPathNotShowStreamPath: true,
}

var plugin = InstallPlugin(RecordPluginConfig, defaultYaml)
var exceptionChannel = make(chan *Exception)
var db *gorm.DB

func (conf *RecordConfig) OnEvent(event any) {
	switch v := event.(type) {
	case FirstConfig, config.Config:
		//if conf.MysqlDSN == "" {
		//	plugin.Error("mysqlDSN 数据库连接配置为空，无法运行，请在config.yaml里配置")
		//}

		go func() { //处理所有异常，录像中断异常、录像读取异常、录像导出文件中断、磁盘容量低于阈值异常、磁盘异常
			for exception := range exceptionChannel {
				SendToThirdPartyAPI(exception)
			}
		}()
		if conf.MysqlDSN == "" {
			plugin.Info("sqliteDb filepath is" + conf.SqliteDbPath)
			db = initSqliteDB(conf.SqliteDbPath)
		} else {
			plugin.Info("mysqlDSN is" + conf.MysqlDSN)
			db = initMysqlDB(conf.MysqlDSN)
		}

		if conf.RecordFileExpireDays > 0 { //当有设置录像文件自动删除时间时，则开始运行录像自动删除的进程
			//主要逻辑为
			//搜索event_records表中event_level值为1的（非重要）数据，并将其create_time与当前时间比对，大于RecordFileExpireDays则进行删除，数据库标记is_delete为1，磁盘上删除录像文件
			go func() {
				for {
					var eventRecords []EventRecord
					expireTime := time.Now().AddDate(0, 0, -conf.RecordFileExpireDays)
					// 创建包含查询条件的 EventRecord 对象
					// queryRecord := EventRecord{
					// 	IsDelete: "0", // 查询条件：is_delete = 1
					// }
					fmt.Printf(" 进行录像文件自动删除： 即将删除创建时间小于 %s 的录像文件。\n", expireTime.Format("2006-01-02 15:04:05"))
					err = db.Where("create_time < ?", expireTime).Find(&eventRecords).Error
					if err == nil {
						if len(eventRecords) > 0 {
							for _, record := range eventRecords {
								fmt.Printf("执行删除 录像ID: %d, 创建时间: %s, 录像文件: %s\n", record.RecId, record.CreateTime, record.Filepath)
								err = os.Remove(record.Filepath)
								if err != nil {
									fmt.Println("error is " + err.Error())
								}
								err = db.Delete(record).Error
								if err != nil {
									fmt.Println("error is " + err.Error())
								}
							}
						}
					}

					// 等待 1 分钟后继续执行
					<-time.After(1 * time.Minute)
				}
			}()
		}
		//检查录像任务是否存在，不存在则启动
		conf.CheckRecordDB()

		conf.Flv.Init()
		conf.Mp4.Init()
		conf.Fmp4.Init()
		conf.Hls.Init()
		conf.Raw.Init()
		conf.RawAudio.Init()
	case SEpublish:
		streamPath := v.Target.Path
		if conf.Flv.NeedRecord(streamPath) {
			go NewFLVRecorder(OrdinaryMode).Start(streamPath)
		}
		if conf.Mp4.NeedRecord(streamPath) {
			go NewMP4Recorder().Start(streamPath)
		}
		if conf.Fmp4.NeedRecord(streamPath) {
			go NewFMP4Recorder().Start(streamPath)
		}
		if conf.Hls.NeedRecord(streamPath) {
			go NewHLSRecorder().Start(streamPath)
		}
		if conf.Raw.NeedRecord(streamPath) {
			go NewRawRecorder().Start(streamPath)
		}
		if conf.RawAudio.NeedRecord(streamPath) {
			go NewRawAudioRecorder().Start(streamPath)
		}
	}
}
func (conf *RecordConfig) getRecorderConfigByType(t string) (recorder *Record) {
	switch t {
	case "flv":
		recorder = &conf.Flv
	case "mp4":
		recorder = &conf.Mp4
	case "fmp4":
		recorder = &conf.Fmp4
	case "hls":
		recorder = &conf.Hls
	case "raw":
		recorder = &conf.Raw
	case "raw_audio":
		recorder = &conf.RawAudio
	}
	return
}

func getFLVDuration(file io.ReadSeeker) uint32 {
	_, err := file.Seek(-4, io.SeekEnd)
	if err == nil {
		var tagSize uint32
		if tagSize, err = util.ReadByteToUint32(file, true); err == nil {
			_, err = file.Seek(-int64(tagSize)-4, io.SeekEnd)
			if err == nil {
				_, timestamp, _, err := codec.ReadFLVTag(file)
				if err == nil {
					return timestamp
				}
			}
		}
	}
	return 0
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}
