package record

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
)

type FileWr interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

var WritingFiles sync.Map

type FileWriter struct {
	filePath string
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
	bufw *bufio.Writer
}

func (f *FileWriter) Seek(offset int64, whence int) (int64, error) {
	if f.bufw != nil {
		f.bufw.Flush()
	}
	return f.Seeker.Seek(offset, whence)
}

func (f *FileWriter) Close() error {
	WritingFiles.Delete(f.filePath)
	return f.Closer.Close()
}

type VideoFileInfo struct {
	Path     string
	Size     int64
	Duration uint32
}

type Record struct {
	Ext           string        `desc:"文件扩展名"`       //文件扩展名
	Path          string        `desc:"存储文件的目录"`     //存储文件的目录
	AutoRecord    bool          `desc:"是否自动录制"`      //是否自动录制
	Filter        config.Regexp `desc:"录制过滤器"`       //录制过滤器
	Fragment      time.Duration `desc:"分片大小，0表示不分片"` //分片大小，0表示不分片
	fs            http.Handler
	CreateFileFn  func(filename string, append bool) (FileWr, error) `json:"-" yaml:"-"`
	GetDurationFn func(file io.ReadSeeker) uint32                    `json:"-" yaml:"-"`
}

func (r *Record) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.fs.ServeHTTP(w, req)
}

func (r *Record) NeedRecord(streamPath string) bool {
	return r.AutoRecord && (!r.Filter.Valid() || r.Filter.MatchString(streamPath))
}

func (r *Record) Init() {
	os.MkdirAll(r.Path, 0766)
	r.fs = http.FileServer(http.Dir(r.Path))
	r.CreateFileFn = func(filename string, append bool) (file FileWr, err error) {
		filePath := filepath.Join(r.Path, filename)
		if err = os.MkdirAll(filepath.Dir(filePath), 0766); err != nil {
			return file, err
		}
		fw := &FileWriter{filePath: filePath}
		if !append {
			if _, loaded := WritingFiles.LoadOrStore(filePath, fw); loaded {
				return file, ErrRecordExist
			}
		}
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|util.Conditoinal(append, os.O_APPEND, os.O_TRUNC), 0666)
		if err == nil && !append {
			fw.Reader = file
			fw.Writer = file
			fw.Seeker = file
			fw.Closer = file
			return fw, nil
		}
		return
	}
}

func (r *Record) Tree(dstPath string, level int) (files []*VideoFileInfo, err error) {
	var dstF *os.File
	dstF, err = os.Open(dstPath)
	if err != nil {
		return
	}
	defer dstF.Close()
	fileInfo, err := dstF.Stat()
	if err != nil {
		return
	}
	if !fileInfo.IsDir() { //如果dstF是文件
		if r.Ext == "." || path.Ext(fileInfo.Name()) == r.Ext {
			p := strings.TrimPrefix(dstPath, r.Path)
			p = strings.ReplaceAll(p, "\\", "/")
			var duration uint32
			if r.GetDurationFn != nil {
				duration = r.GetDurationFn(dstF)
			}
			files = append(files, &VideoFileInfo{
				Path:     strings.TrimPrefix(p, "/"),
				Size:     fileInfo.Size(),
				Duration: duration,
			})
		}
		return
	} else { //如果dstF是文件夹
		var dir []os.FileInfo
		dir, err = dstF.Readdir(0) //获取文件夹下各个文件或文件夹的fileInfo
		if err != nil {
			return
		}
		for _, fileInfo = range dir {
			var _files []*VideoFileInfo
			_files, err = r.Tree(filepath.Join(dstPath, fileInfo.Name()), level+1)
			if err != nil {
				return
			}
			files = append(files, _files...)
		}
		return
	}

}
