package main

import (
	"log"
	"fmt"
	"net/http"
	"github.com/nthskyradiated/stocks-api-go-postgres/router"
)

func main(){
	r := router.Router()
	fmt.Println("server starting on port 8080")

	log.Fatal(http.ListenAndServe(":8080", r))
}