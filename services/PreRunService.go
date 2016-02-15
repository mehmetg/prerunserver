package services
import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"html"
	"path/filepath"
	"os"
	"encoding/json"
	"strings"
	utilities "github.com/mehmetg/prerunserver/utilities"
	"syscall"
	"errors"
)

func PreRunService() {
	fmt.Println("Starting server!")
	ret, _, errNo := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if errNo != 0{
		panic(errors.New(fmt.Sprintf("Fork failed err: %d ", errNo)))
	} else {
		return
	}
	switch ret {
	case 0:
		break
	default:
		os.Exit(0)
	}
	sid, err := syscall.Setsid()
	utilities.CheckError(err)
	if sid == -1 {
		os.Exit(-1)
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/ls", GetFileList)
	router.HandleFunc("/file", GetFile)
	router.HandleFunc("/tunnel", GetTunnel)

	defer log.Fatal(http.ListenAndServe(":5922", router))
	fmt.Println("WTF!")
}

func Index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintf(w, "Hello %q", html.EscapeString(r.URL.Path))
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetFileList(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		path := r.URL.Query().Get("path")
		pattern := r.URL.Query().Get("pattern")
		if len(pattern) == 0 {
			pattern = "*"
		}
		files, _ := filepath.Glob(path + pattern)
		for _, f := range files {
			fmt.Fprintf(w, "%q", html.EscapeString(f))
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fileFullPath := r.URL.Query().Get("filefullpath")
		if _, err := os.Stat(fileFullPath); err == nil {
			http.ServeFile(w, r, fileFullPath)
		} else {
			http.Error(w, "File not found", http.StatusNotFound)
		}

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tunnel := FindTunnel()
		fmt.Fprint(w, html.EscapeString(tunnel))
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func FindTunnel() string {
	var res interface{}
	response := utilities.HttpGetJson("http://localhost:4040/api/tunnels")
	err := json.Unmarshal(response, &res)
	utilities.CheckError(err)
	tunnels := res.(map[string]interface{})
	for _, v := range tunnels {
		switch vv := v.(type) {
		case []interface{}:
			for _, u := range vv {
				tunnel := u.(map[string]interface{})
				for _, val := range tunnel {
					switch val.(type) {
					case string:
						strval, ok := val.(string)
						if ok && strings.Contains(strval, "https://") {
							//fmt.Println(tunnel["public_url"])
							return strval
						}
					}
				}
			}
		}
	}
	return ""
}