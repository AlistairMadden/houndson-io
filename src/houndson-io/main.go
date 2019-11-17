package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/coreos/go-systemd/activation"
	"github.com/gorilla/mux"
)

type siteHandler struct {
	commonPageData
}

func (handler siteHandler) homepageHandler(w http.ResponseWriter, r *http.Request) {
	data := todoPageData{
		CommonPageData: handler.commonPageData,
		HeadTitle:      "Houndson - Home",
		PageTitle:      "My TODO list",
		Todos: []todo{
			{Title: "Task 1", Done: false},
			{Title: "Task 2", Done: true},
			{Title: "Task 3", Done: true},
		},
	}
	err := index.ExecuteTemplate(w, "index", data)
	if err != nil {
		fmt.Println(err)
	}
}

func (handler siteHandler) aboutHandler(w http.ResponseWriter, r *http.Request) {
	time := time.Now().String()
	data := aboutPageData{
		HeadTitle: "Houndson - About",
		Time:      time,
	}
	about.ExecuteTemplate(w, "index", data)
}

type commonPageData struct {
	BaseAddress string
}

type todo struct {
	Title string
	Done  bool
}

type todoPageData struct {
	CommonPageData commonPageData
	HeadTitle      string
	PageTitle      string
	Todos          []todo
}

type aboutPageData struct {
	CommonPageData commonPageData
	HeadTitle      string
	Time           string
}

var index *template.Template
var about *template.Template

func main() {

	// Specify project root dir
	rootDir := "/home/ajm/workspaces/houndson-io/"

	// Gather template files
	commonDirs := []string{"common"}
	homepageDirs := append(commonDirs, []string{"homepage"}...)
	aboutDirs := append(commonDirs, []string{"about"}...)

	var err error
	var homepageFiles []string
	var aboutFiles []string
	homepageFiles, err = getFilesInDirectory(rootDir, homepageDirs)
	aboutFiles, err = getFilesInDirectory(rootDir, aboutDirs)

	println(homepageDirs)

	if err != nil {
		panic(err)
	}

	// Build them
	index = template.Must(template.ParseFiles(homepageFiles...))
	about = template.Must(template.ParseFiles(aboutFiles...))

	// Specify directory for serving static files
	var dir string
	flag.StringVar(&dir, "dir", "static/", "the directory to serve files from. Defaults to the static dir")
	flag.Parse()

	baseDir := "/home/ajm/workspaces/houndson-io"
	staticDir := filepath.Join(baseDir, "static")

	println(staticDir)

	commonPageData := commonPageData{
		BaseAddress: "https://houndson.io",
	}
	siteHandler := siteHandler{
		commonPageData: commonPageData,
	}

	// Register routes
	var router = mux.NewRouter()
	router.HandleFunc("/", siteHandler.homepageHandler)
	router.HandleFunc("/about", siteHandler.aboutHandler)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Create server
	server := &http.Server{
		Handler: router,
	}

	// Get socket file descriptor from systemd
	listeners, err := activation.Listeners()
	if err != nil {
		panic(err)
	}

	if len(listeners) != 1 {
		println(len(listeners))
		panic("Unexpected number of socket activation fds")
	}

	err = server.Serve(listeners[0])
	println(err)
}

func getFilesInDirectory(rootDir string, fileDirs []string) ([]string, error) {
	var files []string
	var err error
	for _, fileDir := range fileDirs {
		err = filepath.Walk(rootDir + fileDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			files = append(files, path)
			return nil
		})
	}

	return files, err
}
