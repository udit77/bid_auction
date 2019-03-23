package bidding

import(
	"github.com/bid_auction/web"
	"net/http"
	"encoding/json"
	"time"
	"math/rand"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"fmt"
	"github.com/gorilla/mux"
)

type AdRequest struct{
	AdPlacementID string
}

type AdObject struct{
	AdSlot string `json:"ad_slot"`
	AdID string	`json:"ad_id"`
	BidPrice int	`json:"bid_price"`
}


type biddingModule struct{
	api *web.Web
}

const (
	min = 100
	max = 1000
)


func GetModule(web *web.Web) *biddingModule{
	module := new(biddingModule)
	module.api = web
	return module
}


func requestBid(adPlacementId string)( *AdObject, error){
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(max - min) + min
	if num % 5 == 0 {
		return &AdObject{}, errors.New("no bidding found for the ad slot")
	}
	return &AdObject{
		AdID:fmt.Sprintf("%v",uuid.NewV4()),
		BidPrice:num*3,
		AdSlot:adPlacementId,
	}, nil
}


func NewHTTPHandler(mod *biddingModule) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		adKey := params["ad_placement_id"]
		bidResponse , err := requestBid(adKey)
		if err != nil {
			http.Error(w,"",http.StatusNoContent)
			return
		}
		apiResponse, err := json.Marshal(bidResponse)
		if err != nil{
			http.Error(w,"",http.StatusInternalServerError)
			return
		}
		w.Write(apiResponse)
		return
	})
}
