package web

import (
	"github.com/bid_auction/config"
)

type Web struct{
	Config    *config.Config
}

func New(cfg *config.Config) *Web {
	w := &Web{
		Config: cfg,
	}
	return w
}

