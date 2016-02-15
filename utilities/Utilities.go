package utilities
import (
	"os/exec"
	"os"
	"fmt"
	"path"
	"net/http"
	"io"
	"time"
	"errors"
	"io/ioutil"
	"archive/zip"
)

func ExecuteBinary(bin string, args []string) *exec.Cmd{
	path := os.Getenv("PATH")
	pwd, err :=  os.Getwd()
	CheckError(err)
	err = os.Setenv("PATH", path + ":" + pwd)
	CheckError(err)
	binary, err := exec.LookPath(bin)
	cmd := exec.Command(binary, args...)
	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start command
	err = cmd.Start()
	CheckError(err)
	//err = cmd.Process.Release()
	//CheckError(err)
	return cmd
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

func HttpGetJson(url string) ([] byte) {
	r, err := http.Get(url)
	retries := 0
	for retries < 3 && (err != nil || r.StatusCode != http.StatusOK) {
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

func CheckError(err error){
	if err != nil{
		panic(err)
	}
}

func GetNgrokLink(ops, arch string) (string){
	lnx64 := "https://dl.ngrok.com/ngrok_2.0.19_linux_amd64.zip"
	lnx32 := "https://dl.ngrok.com/ngrok_2.0.19_linux_386.zip"
	win32 := "https://dl.ngrok.com/ngrok_2.0.19_windows_386.zip"
	win64 := "https://dl.ngrok.com/ngrok_2.0.19_windows_amd64.zip"
	mac32 := "https://dl.ngrok.com/ngrok_2.0.19_darwin_386.zip"
	mac64 := "https://dl.ngrok.com/ngrok_2.0.19_darwin_amd64.zip"
	var link string
	switch ops {
	case "linux":
		if arch == "386" {
			link = lnx32
		} else if arch == "amd64" {
			link = lnx64
		}
	case "darwin":
		if arch == "386" {
			link = mac32
		} else if arch== "amd64" {
			link = mac64
		}
	case "windows":
		if arch == "386" {
			link = win32
		} else if arch == "amd64" {
			link = win64
		}
	default:
		link = "None"
	}
	return link
}