package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type CreateMarginSpecialKeyService struct {
	c              *Client
	apiName        string
	symbol         *string
	ips            *string
	publicKey      *string
	permissionMode *string
	recvWindow     *int64
}

// APIName sets the quoteAsset parameter.
func (s *CreateMarginSpecialKeyService) APIName(apiName string) *CreateMarginSpecialKeyService {
	s.apiName = apiName
	return s
}

// Symbol sets the symbol parameter.
func (s *CreateMarginSpecialKeyService) Symbol(symbol string) *CreateMarginSpecialKeyService {
	s.symbol = &symbol
	return s
}

// Ips sets the ips parameter (comma-separated IP list).
func (s *CreateMarginSpecialKeyService) Ips(ips string) *CreateMarginSpecialKeyService {
	s.ips = &ips
	return s
}

// PublicKey sets the publicKey parameter.
func (s *CreateMarginSpecialKeyService) PublicKey(publicKey string) *CreateMarginSpecialKeyService {
	s.publicKey = &publicKey
	return s
}

// PermissionMode sets the permissionMode parameter.
func (s *CreateMarginSpecialKeyService) PermissionMode(permissionMode string) *CreateMarginSpecialKeyService {
	s.permissionMode = &permissionMode
	return s
}

// RecvWindow sets the recvWindow parameter.
func (s *CreateMarginSpecialKeyService) RecvWindow(recvWindow int64) *CreateMarginSpecialKeyService {
	s.recvWindow = &recvWindow
	return s
}

// Do sends the request.
func (s *CreateMarginSpecialKeyService) Do(ctx context.Context, opts ...RequestOption) (res MarginSpecialKey, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/sapi/v1/margin/apiKey",
		secType:  secTypeSigned,
	}
	if s.apiName != "" {
		r.setParam("apiName", s.apiName)
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	if s.ips != nil {
		r.setParam("ip", *s.ips)
	}
	if s.publicKey != nil {
		r.setParam("publicKey", *s.publicKey)
	}
	if s.permissionMode != nil {
		r.setParam("permissionMode", *s.permissionMode)
	}
	if s.recvWindow != nil {
		r.setParam("recvWindow", *s.recvWindow)
	}
	r.setParam("timestamp", time.Now().UnixMilli())

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return MarginSpecialKey{}, err
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return MarginSpecialKey{}, err
	}
	return res, nil
}

// MarginSpecialKey defines the response of CreateMarginSpecialKeyService
type MarginSpecialKey struct {
	APIKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
	Type      string `json:"type"`
}

type DeleteMarginSpecialKeyService struct {
	c          *Client
	apiName    *string
	apiKey     *string
	symbol     *string
	recvWindow *int64
}

// APIName sets the quoteAsset parameter.
func (s *DeleteMarginSpecialKeyService) APIName(apiName string) *DeleteMarginSpecialKeyService {
	s.apiName = &apiName
	return s
}

// Symbol sets the symbol parameter.
func (s *DeleteMarginSpecialKeyService) Symbol(symbol string) *DeleteMarginSpecialKeyService {
	s.symbol = &symbol
	return s
}

// APIKey sets the APIKey parameter.
func (s *DeleteMarginSpecialKeyService) APIKey(apiKey string) *DeleteMarginSpecialKeyService {
	s.apiKey = &apiKey
	return s
}

// RecvWindow sets the recvWindow parameter.
func (s *DeleteMarginSpecialKeyService) RecvWindow(recvWindow int64) *DeleteMarginSpecialKeyService {
	s.recvWindow = &recvWindow
	return s
}

// Do sends the request.
func (s *DeleteMarginSpecialKeyService) Do(ctx context.Context, opts ...RequestOption) (err error) {
	r := &request{
		method:   http.MethodDelete,
		endpoint: "/sapi/v1/margin/apiKey",
		secType:  secTypeSigned,
	}
	if s.apiName != nil {
		r.setParam("apiName", *s.apiName)
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	if s.apiKey != nil {
		r.setParam("apiKey", *s.apiKey)
	}
	if s.recvWindow != nil {
		r.setParam("recvWindow", *s.recvWindow)
	}
	r.setParam("timestamp", time.Now().UnixMilli())

	_, err = s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ##############

type UpdateMarginSpecialKeyService struct {
	c          *Client
	apiKey     string
	symbol     *string
	ips        *string
	recvWindow *int64
}

// APIName sets the quoteAsset parameter.
func (s *UpdateMarginSpecialKeyService) APIName(apiName string) *UpdateMarginSpecialKeyService {
	s.apiKey = apiName
	return s
}

// Symbol sets the symbol parameter.
func (s *UpdateMarginSpecialKeyService) Symbol(symbol string) *UpdateMarginSpecialKeyService {
	s.symbol = &symbol
	return s
}

// APIKey sets the APIKey parameter.
func (s *UpdateMarginSpecialKeyService) IP(ips string) *UpdateMarginSpecialKeyService {
	s.ips = &ips
	return s
}

// RecvWindow sets the recvWindow parameter.
func (s *UpdateMarginSpecialKeyService) RecvWindow(recvWindow int64) *UpdateMarginSpecialKeyService {
	s.recvWindow = &recvWindow
	return s
}

// Do sends the request.
func (s *UpdateMarginSpecialKeyService) Do(ctx context.Context, opts ...RequestOption) (err error) {
	r := &request{
		method:   http.MethodPut,
		endpoint: "/sapi/v1/margin/apiKey",
		secType:  secTypeSigned,
	}
	if s.apiKey != "" {
		r.setParam("apiKey", s.apiKey)
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	if s.ips != nil {
		r.setParam("ip", *s.ips)
	}

	if s.recvWindow != nil {
		r.setParam("recvWindow", *s.recvWindow)
	}
	r.setParam("timestamp", time.Now().UnixMilli())

	_, err = s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ###

type ListMarginSpecialKeyService struct {
	c          *Client
	symbol     *string
	recvWindow *int64
}

// Symbol sets the symbol parameter.
func (s *ListMarginSpecialKeyService) Symbol(symbol string) *ListMarginSpecialKeyService {
	s.symbol = &symbol
	return s
}

// RecvWindow sets the recvWindow parameter.
func (s *ListMarginSpecialKeyService) RecvWindow(recvWindow int64) *ListMarginSpecialKeyService {
	s.recvWindow = &recvWindow
	return s
}

type MarginAPIKey struct {
	APIName        string `json:"apiName"`
	APIKey         string `json:"apiKey"`
	IP             string `json:"ip"`
	Type           string `json:"type"`
	PermissionMode string `json:"permissionMode"`
}

// Do sends the request.
func (s *ListMarginSpecialKeyService) Do(ctx context.Context, opts ...RequestOption) (res []MarginAPIKey, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/sapi/v1/margin/api-key-list",
		secType:  secTypeSigned,
	}
	if s.symbol != nil {
		r.setParam("symbol", *s.symbol)
	}
	if s.recvWindow != nil {
		r.setParam("recvWindow", *s.recvWindow)
	}
	r.setParam("timestamp", time.Now().UnixMilli())

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
