# record插件

对流进行录制的功能插件，提供Flv和fmp4格式的录制功能。
## 配置

- 配置中的path 表示要保存的文件的根路径，可以使用相对路径或者绝对路径
- filter 代表要过滤的StreamPath正则表达式，如果不匹配，则表示不录制。为空代表不进行过滤
```yaml
record:
  subscribe:
      subaudio: true
      subvideo: true
      iframeonly: false
      waittimeout: 10
  flv:
      ext: .flv
      path: ./flv
      autorecord: false
      filter: ""
  mp4:
      ext: .mp4
      path: ./mp4
      autorecord: false
      filter: ""
  hls:
      ext: .m3u8
      path: ./hls
      autorecord: false
      filter: ""
```

## API

- `/record/api/list?type=flv` 罗列所有录制的flv文件
- `/record/api/start?type=flv&streamPath=live/rtc` 开始录制某个流
- `/record/api/stop?type=flv&streamPath=live/rtc` 停止录制某个流

其中将type值改为mp4则录制成fmp4格式。
## 点播功能

访问格式：
 [http/https]://[host]:[port]/record/[streamPath].[flv/mp4]

例如：
- `http://localhost:8080/record/live/test.flv` 将会读取对应的flv文件
- `http://localhost:8080/record/live/test.mp4` 将会读取对应的fmp4文件

