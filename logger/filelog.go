package logger

import (
	"fmt"
	"os"
	"time"
)

type FileLogWriter struct {
	rec chan *LogRecord
	rot chan bool

	filename string
	file     *os.File

	format string

	header, trailer string

	maxlines          int
	maxlines_curlines int

	maxsize         int
	maxsize_cursize int

	daily          bool
	daily_opendate int

	rotate bool
}

func (w *FileLogWriter) LogWrite(rec *LogRecord) {
	w.rec <- rec
}

func (w *FileLogWriter) Close() {
	close(w.rec)
	w.file.Sync()
}

func NewFileLogWriter(fname string, rotate bool) *FileLogWriter {
	w := &FileLogWriter{
		rec:      make(chan *LogRecord, LogBufferLength),
		rot:      make(chan bool),
		filename: fname,
		format:   "[%D %T] [%L] (%S) %M",
		rotate:   rotate,
	}

	if err := w.intRotate(); err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return nil
	}

	go func() {
		defer func() {
			if w.file != nil {
				fmt.Fprintf(w.file, FormatLogRecord(w.trailer, &LogRecord{Created: time.Now()}))
				w.file.Close()
			}
		}()

		for {
			select {
			case <-w.rot:
				if err := w.intRotate(); err != nil {
					fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
					return
				}
			case rec, ok := <-w.rec:
				if !ok {
					return
				}
				now := time.Now()
				if (w.maxlines > 0 && w.maxlines_curlines >= w.maxlines) ||
					(w.maxsize > 0 && w.maxsize_cursize >= w.maxsize) ||
					(w.daily && now.Day() != w.daily_opendate) {
					if err := w.intRotate(); err != nil {
						fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
						return
					}
				}

				n, err := fmt.Fprint(w.file, FormatLogRecord(w.format, rec))
				if err != nil {
					fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
					return
				}

				w.maxlines_curlines++
				w.maxsize_cursize += n
			}
		}
	}()

	return w
}

func (w *FileLogWriter) Rotate() {
	w.rot <- true
}

func (w *FileLogWriter) intRotate() error {
	if w.file != nil {
		fmt.Fprint(w.file, FormatLogRecord(w.trailer, &LogRecord{Created: time.Now()}))
		w.file.Close()
	}

	if w.rotate {
		_, err := os.Lstat(w.filename)
		if err == nil {
			num := 1
			fname := ""
			// 如果设置按日切换，并且当前时间不等于上次文件打开时间
			if w.daily && time.Now().Day() != w.daily_opendate {
				// 获得昨天日期
				yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
				for ; err == nil && num <= 999; num++ {
					// 新文件名 ＝ 文件名.昨日日期.num
					fname = w.filename + fmt.Sprintf(".%s.%03d", yesterday, num)
					_, err = os.Lstat(fname)
				}
			} else {
				// 如果不是按日，则使用当前时间
				for ; err == nil && num <= 999; num++ {
					fname = w.filename + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), num)
					_, err = os.Lstat(fname)
				}
			}

			// 如果err 为空，则表示 有文件不存在
			if err == nil {
				return fmt.Errorf("Rotate: Canot find free log number to rename %s\n", w.filename)
			}
			w.file.Close()

			err = os.Rename(w.filename, fname)
			if err != nil {
				return fmt.Errorf("Rotate: %s\n", err)
			}
		}
	}

	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	w.file = fd

	now := time.Now()
	fmt.Fprint(w.file, FormatLogRecord(w.header, &LogRecord{Created: now}))

	w.daily_opendate = now.Day()
	w.maxlines_curlines = 0
	w.maxsize_cursize = 0

	return nil
}

func (w *FileLogWriter) SetFormat(format string) *FileLogWriter {
	w.format = format
	return w
}

func (w *FileLogWriter) SetHeadFoot(head, foot string) *FileLogWriter {
	w.header, w.trailer = head, foot
	if w.maxlines_curlines == 0 {
		fmt.Fprint(w.file, FormatLogRecord(w.header, &LogRecord{Created: time.Now()}))
	}
	return w
}

func (w *FileLogWriter) SetRotateLines(maxlines int) *FileLogWriter {
	w.maxlines = maxlines
	return w
}

func (w *FileLogWriter) SetRotateSize(maxsize int) *FileLogWriter {
	w.maxsize = maxsize
	return w
}

func (w *FileLogWriter) SetRotateDaily(daily bool) *FileLogWriter {
	w.daily = daily
	return w
}

func (w *FileLogWriter) SetRotate(rotate bool) *FileLogWriter {
	w.rotate = rotate
	return w
}

func NewXMLLogWriter(fname string, rotate bool) *FileLogWriter {
	return NewFileLogWriter(fname, rotate).SetFormat(
		`	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%S</source>
		<message>%M</message>
	</record>`).SetHeadFoot("<log created=\"%D %T\">", "</log>")
}
