package binance

import (
	"context"
	"encoding/json"
	"net/http"
)

// PayTransactionService retrieve the fiat deposit/withdraw history
type MarginSubscriptionTokenService struct {
	c          *Client
	symbol     *string
	isIsolated *bool
	validity   *int64
}

// Symbol set Symbol
func (s *MarginSubscriptionTokenService) Symbol(symbol string) *MarginSubscriptionTokenService {
	s.symbol = &symbol
	return s
}

// IsIsolated set IsIsolated
func (s *MarginSubscriptionTokenService) IsIsolated(isIsolated bool) *MarginSubscriptionTokenService {
	s.isIsolated = &isIsolated
	return s
}

// Validity set Validity
func (s *MarginSubscriptionTokenService) Validity(validity int64) *MarginSubscriptionTokenService {
	s.validity = &validity
	return s
}

// Do send request
func (s *MarginSubscriptionTokenService) Do(ctx context.Context, opts ...RequestOption) (*MarginSubscriptionToken, error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/sapi/v1/userListenToken",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	if s.isIsolated != nil {
		r.setParam("isIsolated", *s.isIsolated)
	}
	if s.validity != nil {
		r.setParam("validity", *s.validity)
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res := MarginSubscriptionToken{}
	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type MarginSubscriptionToken struct {
	Token          string `json:"token"`
	ExpirationTime int64  `json:"expirationTime"`
}
