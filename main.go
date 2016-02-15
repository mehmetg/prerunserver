package main

import (
	"fmt"
	services "github.com/mehmetg/prerunserver/services"

)

func main(){
	fmt.Println("Starting!")
	fmt.Println("Launching Ngrok!")
	services.NgrokService()
	fmt.Println("Launching PreRunService!")
	services.PreRunService()
	fmt.Println("Done!")
}
