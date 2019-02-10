package citygame

import (
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"git.fleta.io/fleta/city_game/template"
)

var (
	libPath string
)

func init() {
	var pwd string
	{
		pc := make([]uintptr, 10) // at least 1 entry needed
		runtime.Callers(1, pc)
		f := runtime.FuncForPC(pc[0])
		pwd, _ = f.FileLine(pc[0])

		path := strings.Split(pwd, "/")
		pwd = strings.Join(path[:len(path)-1], "/")
	}

	libPath = pwd
}

var t *template.Template

func Start() {
	t = template.NewTemplate(&template.TemplateConfig{
		TemplatePath: libPath + "/html/pages/",
		LayoutPath:   libPath + "/html/layout/",
	})

	http.HandleFunc("/", pageHandler)

	panic(http.ListenAndServe(":8080", nil))

	select {}
}

// Handle HTTP request to either static file server or page server
func pageHandler(w http.ResponseWriter, r *http.Request) {
	//remove first "/" character
	urlPath := r.URL.Path[1:]

	//if the path is include a dot direct to static file server
	if strings.Contains(urlPath, ".") {
		// define your static file directory
		staticFilePath := libPath + "/html/resource/"
		//other wise, let read a file path and display to client
		http.ServeFile(w, r, staticFilePath+urlPath)
	} else {
		data, err := t.Route(r, urlPath)
		// data, err := e.routePath(r, urlPath)
		if err != nil {
			handleErrorCode(500, "Unable to retrieve file", w)
		} else {
			w.Write(data)
		}
	}
}

// Generate error page
func handleErrorCode(errorCode int, description string, w http.ResponseWriter) {
	w.WriteHeader(errorCode)                    // set HTTP status code (example 404, 500)
	w.Header().Set("Content-Type", "text/html") // clarify return type (MIME)

	data, _ := ioutil.ReadFile(libPath + "/html/errors/error-1.html")

	w.Write(data)
}
