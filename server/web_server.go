package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/labstack/echo"
)

type WebServer struct {
	path            string
	hasWatch        bool
	templates       *template.Template
	echo            *echo.Echo
	isRequireReload bool
	sync.Mutex
}

func NewWebServer(echo *echo.Echo, path string) *WebServer {
	web := &WebServer{
		echo: echo,
		path: path,
	}

	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		WebPath, err := filepath.Abs(path)
		if err != nil {
			log.Fatalln(err)
		}

		NewFileWatcher(WebPath, func(ev string, path string) {
			if strings.HasPrefix(filepath.Ext(path), ".htm") {
				web.isRequireReload = true
			}
		})
		web.hasWatch = true
	}
	web.UpdateRender()

	return web
}

func (web *WebServer) CheckWatch() {
	if web.isRequireReload {
		web.Lock()
		if web.isRequireReload {
			err := web.UpdateRender()
			if err != nil {
				log.Println(err)
			} else {
				web.isRequireReload = false
			}
		}
		web.Unlock()
	}
}

func (web *WebServer) UpdateRender() error {
	tp := template.New("").Delims("<%", "%>").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	})
	if web.hasWatch {
		filepath.Walk(web.path, func(path string, fi os.FileInfo, err error) error {
			if strings.HasPrefix(filepath.Ext(path), ".htm") {
				rel, err := filepath.Rel(web.path, path)
				if err != nil {
					return err
				}
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				rel = filepath.ToSlash(rel)
				template.Must(tp.New(rel).Parse(string(data)))
			}
			return nil
		})
	} else {
		for path, v := range staticFiles {
			if strings.HasPrefix(filepath.Ext(path), ".htm") {
				var data []byte
				if v.size == 0 {
					data = []byte(v.data)
				} else {
					br := bytes.NewReader([]byte(v.data))
					r, err := gzip.NewReader(br)
					body, err := ioutil.ReadAll(r)
					if err != nil {
						panic(err)
					}
					data = body
				}
				template.Must(tp.New(path).Parse(string(data)))
			}
		}
	}
	web.templates = tp

	return nil
}

func (web *WebServer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return web.templates.ExecuteTemplate(w, name, data)
}

func (web *WebServer) SetupStatic(e *echo.Echo, prefix string, root string) {
	h := func(c echo.Context) error {
		fname := c.Param("*")
		upath := path.Join(prefix, fname)[1:]
		if web.hasWatch {
			fpath := path.Join(root, fname)
			return c.File(fpath)
		} else {
			if file, has := staticFiles[upath]; !has {
				return c.NoContent(http.StatusNotFound)
			} else {
				rw := c.Response().Writer
				header := c.Response().Header()
				req := c.Request()

				if file.hash != "" {
					if hash := req.Header.Get("If-None-Match"); hash == file.hash {
						c.Response().WriteHeader(http.StatusNotModified)
						return nil
					}
					header.Set("ETag", file.hash)
				}
				if !file.mtime.IsZero() {
					if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil && file.mtime.Before(t.Add(1*time.Second)) {
						c.Response().WriteHeader(http.StatusNotModified)
						return nil
					}
					header.Set("Last-Modified", file.mtime.UTC().Format(http.TimeFormat))
				}

				header.Set("Content-Type", file.mime)
				bUnzip := false
				if file.size > 0 {
					if header.Get("Content-Encoding") == "" && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
						header.Set("Content-Length", strconv.Itoa(len(file.data)))
						header.Set("Content-Encoding", "gzip")
					} else {
						header.Set("Content-Length", strconv.Itoa(file.size))
						bUnzip = true
					}
				} else {
					header.Set("Content-Length", strconv.Itoa(len(file.data)))
				}
				c.Response().WriteHeader(http.StatusOK)
				if bUnzip {
					reader, err := gzip.NewReader(strings.NewReader(file.data))
					if err != nil {
						return err
					}
					defer reader.Close()
					io.Copy(rw, reader)
				} else {
					io.WriteString(rw, file.data)
				}
				return nil
			}
		}
	}
	e.GET(prefix, h)
	if prefix == "/" {
		e.GET(prefix+"*", h)
	} else {
		e.GET(prefix+"/*", h)
	}
}
