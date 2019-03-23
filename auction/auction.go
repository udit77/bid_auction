package auction

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"net"
	"github.com/bid_auction/bidding"
	"github.com/bid_auction/web"
	"github.com/eapache/go-resiliency/breaker"
	"github.com/gorilla/mux"
)

type auctionModule struct {
	api *web.Web
	cb  *breaker.Breaker
}
type headerBid struct {
	score int
	err   error
}

func GetModule(web *web.Web) *auctionModule {
	module := new(auctionModule)
	module.api = web
	module.cb = breaker.New(web.Config.CircuitBreaker.ErrorThreshold, web.Config.CircuitBreaker.SuccessThreshold,
		time.Duration(web.Config.CircuitBreaker.Timeout)*time.Millisecond)
	return module
}

func makeBidRequest(adPlacementID string) ([]byte, error) {
	uri := "http://127.0.0.1:9000/bidding/"+adPlacementID
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{
		Timeout: 200 * time.Millisecond,
	}
	res, err := httpClient.Do(req)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return []byte{}, errors.New("timeout occurred in calling bidding service")
		}
		return []byte{}, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return []byte{}, err
		}
		return responseBody, nil
	} else {
		return []byte{}, nil
	}
}

func (mod *auctionModule) getBid(adRequest *bidding.AdRequest) (*bidding.AdObject, error) {
	var (
		apiResponse []byte
		err         error
	)

	if mod.api.Config.CircuitBreaker.Enabled {
		resBreaker := mod.cb.Run(func() error {
			apiResponse, err = makeBidRequest(adRequest.AdPlacementID)
			return err
		})
		if resBreaker == breaker.ErrBreakerOpen {
			return &bidding.AdObject{}, errors.New("circuit breaker is open")
		} else if resBreaker != nil {
			return &bidding.AdObject{}, resBreaker
		}
	} else {
		apiResponse, err = makeBidRequest(adRequest.AdPlacementID)
		if err != nil {
			return &bidding.AdObject{}, err
		}
	}

	response := &bidding.AdObject{}
	err = json.Unmarshal(apiResponse, response)
	if err != nil {
		return &bidding.AdObject{}, err
	}
	return response, nil
}

func (mod *auctionModule) getHeaderBiddingForAd(adRequest *bidding.AdRequest) (int, error) {
	var (
		wg            sync.WaitGroup
		maxBid        int
		validResponse bool
	)
	activeServices := mod.api.Config.ActiveBiddingServices
	c := make(chan headerBid, activeServices)
	wg.Add(activeServices)
	for i := 0; i < activeServices; i++ {
		go func(i int) {
			defer wg.Done()
			response, err := mod.getBid(adRequest)
			c <- headerBid{response.BidPrice, err}
		}(i)
	}
	wg.Wait()
	close(c)
	for val := range c {
		if val.score > maxBid {
			maxBid = val.score
		}
		if val.err == nil {
			validResponse = true
		}
	}
	if !validResponse {
		return maxBid, errors.New("no valid request bid found")
	}
	return maxBid, nil
}

func NewHTTPHandler(mod *auctionModule) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		adRequest := new(bidding.AdRequest)
		adRequest.AdPlacementID = params["ad_placement_id"]

		headerBid, err := mod.getHeaderBiddingForAd(adRequest)
		if err != nil {
			http.Error(w, "", http.StatusNoContent)
			return
		}
		apiResponse := make(map[string]interface{})
		apiResponse["bid_price"] = headerBid
		apiResponse["ad_placement_id"] = adRequest.AdPlacementID
		response, err := json.Marshal(apiResponse)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Write(response)
		return
	})
}
