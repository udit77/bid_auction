package main

import (
	"github.com/bid_auction/config"
	"github.com/bid_auction/web"
	"github.com/gorilla/mux"
	"github.com/bid_auction/auction"
	"github.com/bid_auction/bidding"
	"log"
	"net/http"
)


func init(){
	config.Init()
}

func main(){
	web := web.New(config.Get())
	router := mux.NewRouter()
	log.Println("[Main] Initializing bidding and auction services")
	router.Handle("/auction/{ad_placement_id}",auction.NewHTTPHandler(auction.GetModule(web))).Methods("GET")
	router.Handle("/bidding/{ad_placement_id}",bidding.NewHTTPHandler(bidding.GetModule(web))).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:9000",
	}
	log.Fatal(srv.ListenAndServe())
}