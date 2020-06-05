package main

import (
	"bisq-add-on/server"
	"log"
	"net/http"
)

func main() {
	service := server.InitService()

	http.HandleFunc("/buy", service.BuyHandle)
	http.HandleFunc("/sell", service.SellHandle)
	http.HandleFunc("/check-offer", service.CheckOfferHandle)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
