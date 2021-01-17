package main



// source tutorial https://medium.com/@trongdan_tran/circuit-breaker-and-retry-64830e71d0f6
//  how to run : "go run main.go" for success
//  "SERVER_ERROR=1 go run main.go" for failed and triger circuit breaker

import (
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"fmt"
	"log"
	"net/http"
	"os"
)
const commandName = "produce_api"

func main(){

	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		Timeout: 500,
		MaxConcurrentRequests: 100,
		ErrorPercentThreshold: 50,
		RequestVolumeThreshold: 3,
		SleepWindow: 1000,
	})

	http.HandleFunc("/",logger(handle))
	log.Println("listening on : 8081")
	http.ListenAndServe(":8081",nil)
}

func handle(w http.ResponseWriter, r *http.Request){

	output := make (chan bool, 1)
	errors := hystrix.Go(commandName, func() error{
		// talk to other services
		err:= callChargeProducerAPI()
		if err == nil {
			output <- true
		}
		return err
	},nil)

	select {
	case out := <- output:
		// success
		log.Printf("success %v", out)
	case err := <- errors:
		//failure
		log.Printf("failed %s",err)
	}
}

func logger(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path, r.Method)
		fn(w, r)
	}
}

func callChargeProducerAPI() error {
	fmt.Println(os.Getenv("SERVER_ERROR"))
	if os.Getenv("SERVER_ERROR") == "1" {
		return errors.New("503 error")
	}
	return nil
}
