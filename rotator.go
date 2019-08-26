// Copyright © 2019 Moises P. Sena <moisespsena@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logrotate

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/phayes/permbits"

	"github.com/dustin/go-humanize"

	"github.com/moisespsena-go/logging"
	path_helpers "github.com/moisespsena-go/path-helpers"
)

var log = logging.GetOrCreateLogger(path_helpers.GetCalledDir())

const T_FORMAT = "%Y%M%DT%h%m%s%Z"

var MaxSize int64 = 1024 * 1024 * 10
var fmtRe, _ = regexp.Compile("^\\d{8}T\\d{6}Z$")

type Rotator struct {
	Path    string
	Options Options

	dir, name, ext string
	f              *os.File
	mu             sync.RWMutex
	wg             sync.WaitGroup
	lastRotation   Control
}

func New(path string, options ...Options) *Rotator {
	var opt Options
	for _, opt = range options {
	}
	dir := filepath.Dir(path)

	if opt.HistoryDir == "" {
		opt.HistoryDir = path + ".glogrotation"
	}

	name := filepath.Base(path)
	ext := filepath.Ext(name)
	opt.HistoryPath = filepath.Join(opt.HistoryPath, strings.TrimSuffix(name, ext)+"_"+T_FORMAT+ext)

	return &Rotator{
		Path:    path,
		Options: opt,
		dir:     dir,
		name:    name,
		ext:     ext,
	}
}

func (this *Rotator) Close() (err error) {
	this.Wait()
	if this.f != nil {
		this.mu.Lock()
		defer this.mu.Unlock()
		err = this.f.Close()
		this.f = nil
	}
	return nil
}

func (this *Rotator) Send(p []byte) (n int, err error) {
	if this.f == nil {
		_, err = this.Open()
		if err != nil {
			return
		}
	}
	if _, err = this.AutoRotate(len(p)); err != nil {
		return
	}
	return this.f.Write(p)
}

func (this *Rotator) Write(p []byte) (n int, err error) {
	return this.Send(p)
}

func (this *Rotator) Open() (f *os.File, err error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.f != nil {
		err = fmt.Errorf("is open")
		return
	}

	mode := this.Options.FileMode
	if mode == 0 {
		mode = 0600
		this.Options.FileMode = mode
	}

	if this.Options.DirMode == 0 {
		bits := permbits.FileMode(mode)
		bits.SetUserExecute(true)
		this.Options.DirMode = os.FileMode(bits)
	}

	if _, err = os.Stat(this.Options.HistoryDir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(this.Options.HistoryDir, this.Options.DirMode)
		} else {
			return
		}
	} else {
		os.Chmod(this.Options.HistoryDir, this.Options.DirMode)
	}

	if _, err = os.Stat(this.dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(this.dir, this.Options.DirMode)
		} else {
			return
		}
	} else {
		os.Chmod(this.dir, this.Options.DirMode)
	}

	if err = this.loadControlOrCreate(); err != nil {
		return
	}

	if _, err = os.Stat(this.Path); err == nil {
		os.Chmod(this.Path, this.Options.FileMode)
	} else if !os.IsNotExist(err) {
		return
	}

	this.f, err = os.OpenFile(this.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, mode)
	return
}

func (this *Rotator) File() *os.File {
	return this.f
}

func (this *Rotator) RotateOptions() Options {
	return this.Options
}

func (this *Rotator) AutoRotate(increaseSize int) (entry *FileHistroyEntry, err error) {
	if ms := this.Options.MaxSize; ms >= 0 {
		if ms == 0 {
			ms = MaxSize
		}
		var s int64
		if s, err = this.f.Seek(0, io.SeekEnd); err != nil {
			return
		} else if s += int64(increaseSize); s > this.Options.MaxSize {
			return this.Rotate()
		}
	}
	now := time.Now()
	switch this.Options.Duration {
	case Minutely:
		layout := "2006-01-02 15:04"
		if this.lastRotation.Last.Format(layout) != now.Format(layout) {
			return this.Rotate()
		}
	case Hourly:
		layout := "2006-01-02 15"
		if this.lastRotation.Last.Format(layout) != now.Format(layout) {
			return this.Rotate()
		}
	case Daily:
		layout := "2006-01-02"
		if this.lastRotation.Last.Format(layout) != now.Format(layout) {
			return this.Rotate()
		}
	case Weekly:
		if fmt.Sprint(this.lastRotation.Last.ISOWeek()) != fmt.Sprint(now.ISOWeek()) {
			return this.Rotate()
		}
	case 0, Monthly:
		layout := "2006-01"
		if this.lastRotation.Last.Format(layout) != now.Format(layout) {
			return this.Rotate()
		}
	case Yearly:
		if this.lastRotation.Last.Year() != now.Year() {
			return this.Rotate()
		}
	}
	return nil, nil
}

func (this *Rotator) NewNameT(t time.Time) (name, pth string) {
	name = TFormat(t, this.Options.HistoryPath)
	return name, filepath.Join(this.Options.HistoryDir, name)
}

func (this *Rotator) NewName() (name, pth string) {
	return this.NewNameT(time.Now().UTC())
}

func (this *Rotator) controlPath() string {
	return filepath.Join(this.Options.HistoryDir, "."+this.name+".rtr")
}

func (this *Rotator) loadControlOrCreate() (err error) {
	var f *os.File
	this.lastRotation = Control{}
	if f, err = os.Open(this.controlPath()); err != nil {
		if os.IsNotExist(err) {
			this.lastRotation.Last = time.Now()
			return this.saveControl()
		}
		return
	}
	defer f.Close()
	var i int64
	if err = binary.Read(f, binary.BigEndian, &i); err != nil {
		return
	}
	this.lastRotation.Last = time.Unix(i, 0).In(time.Local)
	return
}

func (this *Rotator) saveControl() (err error) {
	var f *os.File
	if f, err = os.OpenFile(this.controlPath(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, this.Options.FileMode); err != nil {
		return
	}
	defer f.Close()
	return binary.Write(f, binary.BigEndian, this.lastRotation.Last.Unix())
}

func (this *Rotator) Rotate() (entry *FileHistroyEntry, err error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	name := this.name
	ext := filepath.Ext(name)
	now := time.Now()
	hRelPath, hPath := this.NewNameT(now.UTC())
	this.lastRotation.Last = now
	if err := this.saveControl(); err != nil {
		log.Errorf("logging.rotator: save control of %q failed: %s", this.Path, err.Error())
	}

	old := this.f
	old.Close()
	if d := filepath.Dir(hPath); d != "." {
		if _, err = os.Stat(d); os.IsNotExist(err) {
			err = os.MkdirAll(d, this.Options.DirMode)
		}
		if err == nil {
			if err = os.Rename(this.Path, hPath); err != nil {
				doerr := func() {
					log.Errorf("logging.rotator: failed to create history entry log file of %q to %q: %s", this.Path, hRelPath, err.Error())
				}
				doerr()
				hPath = filepath.Join(this.Path, TFormat(now, strings.TrimSuffix(name, ext)+"_"+T_FORMAT+ext))
				if err = os.Rename(this.Path, filepath.Join(this.Path, strings.TrimSuffix(name, ext)+"_"+T_FORMAT+ext)); err != nil {
					doerr()
				}
			}
		} else {
			log.Errorf("logging.rotator: failed to create history entry log dir of %q: %s", this.Path, err.Error())
		}
	}

	if err == nil {
		this.f, err = os.OpenFile(this.Path, os.O_RDWR|os.O_CREATE, this.Options.FileMode)
		if err == nil {
			entry = &FileHistroyEntry{at: now, path: hRelPath, root: this.Options.HistoryDir}
			this.wg.Add(1)
			go func() {
				defer this.wg.Done()
				r, err := os.Open(hPath)
				if err != nil {
					log.Errorf("logging.rotator: failed to open for compress history log entry %q: %s", hRelPath, err.Error())
					return
				}
				w, err := os.OpenFile(hPath+".gz", os.O_RDWR|os.O_CREATE|os.O_TRUNC, this.Options.FileMode)
				if err != nil {
					log.Errorf("logging.rotator: failed to compress history log entry %q: %s", hRelPath, err.Error())
					return
				}

				if size, err := io.Copy(gzip.NewWriter(w), r); err == nil {
					log.Debugf("logging.rotator: history log entry %q compressed: size=%s", hRelPath, humanize.Bytes(uint64(size)))
					os.Remove(hPath)
				} else {
					log.Error(err)
					return
				}

				if s := this.Options.HistoryCount; s > 0 {
					var t time.Time
					history, _ := this.History(t, t, 0)
					if len(history) > s {
						for _, e := range history[s:] {
							if err := os.Remove(e.AbsPath()); err != nil {
								log.Error(err)
							} else if d := filepath.Dir(e.path); d != "." {
								if err := path_helpers.Parents(d, string(filepath.Separator), func(sub string) (err error) {
									d := filepath.Join(this.Options.HistoryDir, sub)
									if df, err := os.Open(d); err != nil {
										log.Error(err)
									} else if _, err := df.Readdirnames(1); err != nil {
										if err == io.EOF {
											err = os.Remove(d)
										}
									}
									return
								}); err != nil && err != io.EOF {
									log.Error(err)
								}
							}
						}
					}
				}
			}()
		}
	} else {
		this.f, err = os.OpenFile(this.Path, os.O_RDWR|os.O_APPEND, this.Options.FileMode)
	}
	return
}

func (this *Rotator) Each(cb func(name, info string, finfo os.FileInfo) error) (err error) {
	prefix := this.name
	if this.ext != "" {
		prefix = strings.TrimSuffix(prefix, this.ext)
	}
	prefix += "_"
	return filepath.Walk(this.Options.HistoryDir, func(path string, finfo os.FileInfo, err error) error {
		if os.IsNotExist(err) {
			return filepath.SkipDir
		}
		name := finfo.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".gz") {
			info := name[len(prefix) : len(name)-len(this.ext)-3]
			if info[8] == 'T' && info[15] == 'Z' {
				rel, _ := filepath.Rel(this.Options.HistoryDir, path)
				if err = cb(rel, info, finfo); err != nil {
					return err
				}
			}
		}
		return err
	})
}

func (this *Rotator) History(from, to time.Time, count int64) (events Entries, err error) {
	if err = this.Each(func(pth, info string, finfo os.FileInfo) error {
		if fmtRe.MatchString(info) {
			if t, err := time.Parse("20060102T150405Z", info); err == nil {
				events = append(events, &FileHistroyEntry{
					at:   t.In(time.Local),
					path: pth,
					root: this.Options.HistoryDir,
				})
			} else {
				log.Errorf("bad entry name %q:", pth, err)
			}
		}
		return nil
	}); err != nil {
		return
	}

	if !from.IsZero() {
		events = events.filter(func(entry *FileHistroyEntry) bool {
			return entry.at.Equal(from) || entry.at.After(from)
		})
	}
	if !to.IsZero() {
		events = events.filter(func(entry *FileHistroyEntry) bool {
			return entry.at.Equal(to) || entry.at.Before(to)
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].at.After(events[j].at)
	})
	return
}

func (this *Rotator) Wait() {
	this.wg.Wait()
}

func TFormat(t time.Time, fmt string) string {
	if !strings.Contains(fmt, "%Z") {
		t = t.UTC()
	}

	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						strings.ReplaceAll(
							strings.ReplaceAll(fmt,
								"%Y", t.Format("2006")),
							"%M", t.Format("01")),
						"%D", t.Format("02")),
					"%h", t.Format("15")),
				"%m", t.Format("04")),
			"%s", t.Format("05")),
		"%Z", t.Format("Z0700"),
	)
}

type FileHistroyEntry struct {
	at         time.Time
	root, path string
}

func (this *FileHistroyEntry) At() time.Time {
	return this.at
}

func (this *FileHistroyEntry) Path() string {
	return this.path
}

func (this *FileHistroyEntry) AbsPath() string {
	return filepath.Join(this.root, this.path)
}

func (this *FileHistroyEntry) Reader() (r io.ReadCloser, err error) {
	var f *os.File
	pth := this.AbsPath()
	if f, err = os.Open(pth); os.IsNotExist(err) {
		if f, err = os.Open(pth + ".gz"); err == nil {
			return gzip.NewReader(r)
		}
	}
	r = f
	return
}

type Entries []*FileHistroyEntry

func (this Entries) filter(f func(entry *FileHistroyEntry) bool) (r Entries) {
	for _, e := range this {
		if f(e) {
			r = append(r, e)
		}
	}
	return
}
