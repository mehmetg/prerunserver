package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"net/http"
	"path/filepath"
	"os"
	"io"
	"log"
	"path"
	"archive/zip"
	"os/exec"
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"
	"strings"
	"runtime"
)

func main() {

	lnx64 := "https://dl.ngrok.com/ngrok_2.0.19_linux_amd64.zip"
	lnx32 := "https://dl.ngrok.com/ngrok_2.0.19_linux_386.zip"
	win32 := "https://dl.ngrok.com/ngrok_2.0.19_windows_386.zip"
	win64 := "https://dl.ngrok.com/ngrok_2.0.19_windows_amd64.zip"
	mac32 := "https://dl.ngrok.com/ngrok_2.0.19_darwin_386.zip"
	mac64 := "https://dl.ngrok.com/ngrok_2.0.19_darwin_amd64.zip"
	var link string
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "386" {
			link = lnx32
		} else if runtime.GOARCH == "arm64" {
			link = lnx64
		}
	case "darwin":
		if runtime.GOARCH == "386" {
			link = mac32
		} else if runtime.GOARCH == "arm64" {
			link = mac64
		}
	case "windows":
		if runtime.GOARCH == "386" {
			link = win32
		} else if runtime.GOARCH == "arm64" {
			link = win64
		}
	}

	fmt.Printf("Downloading Ngrok Client for %q - %q !", runtime.GOOS, runtime.GOARCH)
	arch := DownloadFile(link)
	file := Unzip(arch)
	fmt.Println("Starting Ngrok!")
	defer ExecuteBinary(file, []string{"http", "4444", ">", "/dev/null", "&"}).Wait()
	fmt.Println("Starting server!")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/ls", GetFileList)
	router.HandleFunc("/file", GetFile)
	router.HandleFunc("/tunnel", GetTunnel)

	log.Fatal(http.ListenAndServe(":4444", router))

	/*tun, err := tunnel.ListenNgrokIOHTTP()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Serving at: %v\n", tun.Addr())
	err = http.Serve(tun, router)
	if err != nil {
		panic(err)
	}*/
}
func ExecuteBinary(bin string, args []string) *exec.Cmd{
	path := os.Getenv("PATH")
	pwd, err :=  os.Getwd()
	CheckError(err)
	err = os.Setenv("PATH", path + ":" + pwd)
	CheckError(err)
	binary, err := exec.LookPath(bin)
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start command
	err = cmd.Start()
	CheckError(err)
	return cmd
}

func CheckError(err error){
	if err != nil{
		panic(err)
	}
}

func DownloadFile(link string) (string){
	fmt.Println(link)
	_, file := path.Split(link)
	out, err := os.Create(file)
	if err != nil{
		panic(err)
	}
	defer out.Close();

	resp, err := http.Get(link)
	if err != nil{
		panic(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil{
		panic(err)
	}
	return file
}

func Unzip(src string) (string) {
	var fileName string
	r, err := zip.OpenReader(src)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "ngrok" {
			fileName = f.Name
			rc, err := f.Open()
			if err != nil {
				panic(err)
			}
			defer rc.Close()

			f, err := os.OpenFile(
				f.Name, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, f.Mode())
			if err != nil {
				panic(err)
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				panic(err)
			}
		}
	}

	return fileName
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

func GetJson(url string) ([] byte) {
	r, err := http.Get(url)
	retries := 0
	for retries < 3 && (err != nil || r.StatusCode != http.StatusOK)  {
		r, err = http.Get(url)
		retries++
		time.Sleep(500 * time.Millisecond)
	}
	CheckError(err)
	if r.StatusCode != http.StatusOK {
		panic(errors.New("Ngrok did not respond!"))
	}
	defer r.Body.Close()
	contents, err := ioutil.ReadAll(r.Body)
	CheckError(err)
	return contents
}

func FindTunnel() string{
	var res interface{}
	response := GetJson("http://localhost:4040/api/tunnels")
	err := json.Unmarshal(response, &res)
	CheckError(err)
	tunnels := res.(map[string]interface{})
	for _, v := range tunnels {
		switch vv := v.(type) {
		case []interface{}:
			for _, u := range vv {
				tunnel := u.(map[string]interface{})
				for _, val := range tunnel{
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
