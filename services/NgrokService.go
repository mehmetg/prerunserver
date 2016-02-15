package services
import (
	"fmt"
	utilities "github.com/mehmetg/prerunserver/utilities"
	"runtime"
)

func NgrokService(){
	ops := runtime.GOOS
	arch := runtime.GOARCH
	fmt.Printf("Downloading Ngrok Client for %q - %q !", ops, arch)
	archive := utilities.DownloadFile(utilities.GetNgrokLink(ops, arch))
	file := utilities.Unzip(archive)
	fmt.Println("Starting Ngrok!")
	utilities.ExecuteBinary(file, []string{"http", "5922"})
}
