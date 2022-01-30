# record插件
record plugin for monibuca

实现了录制Flv文件的功能，并且支持再次使用录制好的Flv文件作为发布者进行发布。

## 默认配置

```toml
[Record]
Path = "resource/[appName]/[streamName]"
AutoRecord = false
AppendFlv  = false
```

- `Path` 表示要保存的Flv文件的根路径，可以使用相对路径或者绝对路径，默认为流路径
- `AutoRecord` 是否自动开启录制
- `AppendFlv` 当存在历史录制文件时，是否追加写入，如果不存在历史录制文件，则总是新创建文件

## API

- `/api/record/flv/list` 罗列所有录制的flv文件
- `/api/record/flv?streamPath=live/rtc` 开始录制某个流
- `/api/record/flv/stop?streamPath=live/rtc` 停止录制某个流
- `/api/record/flv/play?streamPath=live/rtc` 将某个flv文件读取并发布成一个直播流
- `/api/record/flv/delete?streamPath=live/rtc` 删除某个flv文件

## 点播功能

访问 http://[HOST]:[Gateway Port]/vod/live/rtc.flv 将会读取对应的flv文件