# recordplugin
record plugin for monibuca

实现了录制Flv文件的功能，并且支持再次使用录制好的Flv文件作为发布者进行发布。在Monibuca的web界面的控制台中提供了对房间进行录制的操作按钮，以及列出所有已经录制的文件的界面。

## 插件名称

Record

### 配置
配置中的Path 表示要保存的Flv文件的根路径，可以使用相对路径或者绝对路径
```toml
[Record]
Path="resource"
```