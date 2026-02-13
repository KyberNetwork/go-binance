package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type MarginWSTokenService struct {
	c      *Client
	symbol *string
}

func (s *MarginWSTokenService) Symbol(symbol string) *MarginWSTokenService {
	s.symbol = &symbol
	return s
}

type MarginWSTokenResponse struct {
	Token          string `json:"token"`
	ExpirationTime int64  `json:"expirationTime"`
}

func (s *MarginWSTokenService) Do(ctx context.Context, opts ...RequestOption) (res MarginWSTokenResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/sapi/v1/userListenToken",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	r.setParam("timestamp", time.Now().UnixMilli())

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return MarginWSTokenResponse{}, err
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return MarginWSTokenResponse{}, err
	}
	return res, nil
}
