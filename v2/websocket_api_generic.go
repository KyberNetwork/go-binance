package binance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	websocket2 "github.com/KyberNetwork/go-binance/v2/common/websocket"
	"github.com/gorilla/websocket"
)

type Logger interface {
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

var ErrNoConnection = errors.New("no connection")

type GenericWSResponse struct {
	ID         string          `json:"id"`
	Status     int             `json:"status"`
	Result     json.RawMessage `json:"result,omitempty"`
	RateLimits []RateLimit     `json:"rateLimits,omitempty"`
	Error      *RPCError       `json:"error,omitempty"`
}

func (r GenericWSResponse) ParseJSON(out interface{}) error {
	return json.Unmarshal(r.Result, out)
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ServerTime int64 `json:"serverTime"`
		RetryAfter int64 `json:"retryAfter"`
	} `json:"data"`
}

// Future holds the response channel for an outstanding request.
type Future struct {
	Response <-chan GenericWSResponse // receive exactly one value
	// internal:
	doneCh chan GenericWSResponse
}

type GenericWSClient struct {
	conn websocket2.Connection
	//conn      *websocket.Conn
	writeMu   sync.Mutex // serialize writes to websocket
	pendingMu sync.Mutex
	pending   map[string]chan GenericWSResponse

	idCounter uint64
	// closed when read loop or Close() triggers termination
	terminateOnce sync.Once
	terminateErr  error
	closeCh       chan struct{}
	apiKey        string
	apiSecret     string
	keyType       string
	logger        Logger
	blocker       *RateLimitBlocker
}

func NewGenericWSClient(wsURL, apiKey, apiSecret, keyType string, header http.Header, logger Logger, blocker *RateLimitBlocker) (*GenericWSClient, error) {
	conn, err := websocket2.NewConnection(func() (*websocket.Conn, error) {
		Dialer := websocket.Dialer{
			Proxy:             http.ProxyFromEnvironment,
			HandshakeTimeout:  45 * time.Second,
			EnableCompression: false,
		}
		c, _, err := Dialer.Dial(wsURL, header)
		if err != nil {
			return nil, err
		}

		return c, nil
	}, true, WebsocketPongTimeout)
	if err != nil {
		return nil, err
	}
	c := &GenericWSClient{
		conn:      conn,
		pending:   make(map[string]chan GenericWSResponse),
		closeCh:   make(chan struct{}),
		apiKey:    apiKey,
		apiSecret: apiSecret,
		keyType:   keyType,
		logger:    logger,
		blocker:   blocker,
	}
	go c.readLoop()
	return c, nil
}

func (c *GenericWSClient) Close() error {
	c.terminateOnce.Do(func() {
		c.terminateErr = errors.New("client closed")
		close(c.closeCh)
		_ = c.conn.Close()
		// notify pending
		c.pendingMu.Lock()
		for id, ch := range c.pending {
			// send an error-like response
			ch <- GenericWSResponse{
				ID:    id,
				Error: &RPCError{Code: -1, Message: "client closed"},
			}
			close(ch)
			delete(c.pending, id)
		}
		c.pendingMu.Unlock()
	})
	return c.terminateErr
}
func (c *GenericWSClient) SendRequest(ctx context.Context, method string, params map[string]interface{}, maxTimeout time.Duration) (GenericWSResponse, error) {
	if err := c.blocker.Try(); err != nil {
		return GenericWSResponse{}, err
	}
	// create a context with timeout derived from provided ctx
	ctxReq, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()
	id := strconv.FormatUint(atomic.AddUint64(&c.idCounter, 1), 10)
	rawData, err := websocket2.CreateRequest(
		websocket2.NewRequestData(id, c.apiKey, c.apiSecret, 0, c.keyType),
		websocket2.WsApiMethodType(method), params)
	if err != nil {
		return GenericWSResponse{}, fmt.Errorf("can not create request: %w", err)
	}

	// prepare response channel
	respCh := make(chan GenericWSResponse, 1)

	// store pending
	c.pendingMu.Lock()
	c.pending[id] = respCh
	c.pendingMu.Unlock()

	// ensure cleanup: remove pending on exit
	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
	}()
	if err := c.conn.WriteMessage(websocket.TextMessage, rawData); err != nil {
		return GenericWSResponse{}, fmt.Errorf("write failed: %w", err)
	}
	// wait for response, client close, or timeout/cancel
	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return resp, fmt.Errorf("rpc error: code=%d message=%s", resp.Error.Code, resp.Error.Message)
		}
		return resp, nil
	case <-ctxReq.Done():
		return GenericWSResponse{}, errors.New("connection closed")
	}
}

// SendRequestAsync returns a Future you can wait on later. The future will deliver exactly one GenericWSResponse.
func (c *GenericWSClient) SendRequestAsync(method string, params map[string]interface{}) (*Future, error) {
	if err := c.blocker.Try(); err != nil {
		return nil, err
	}
	id := strconv.FormatUint(atomic.AddUint64(&c.idCounter, 1), 10)
	rawData, err := websocket2.CreateRequest(
		websocket2.NewRequestData(id, c.apiKey, c.apiSecret, 0, c.keyType),
		websocket2.WsApiMethodType(method), params)
	if err != nil {
		return nil, fmt.Errorf("can not create request: %w", err)
	}
	respCh := make(chan GenericWSResponse, 1)
	c.pendingMu.Lock()
	c.pending[id] = respCh
	c.pendingMu.Unlock()

	if err := c.conn.WriteMessage(websocket.TextMessage, rawData); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, err
	}

	done := make(chan GenericWSResponse, 1)
	future := &Future{Response: done, doneCh: done}

	// forward the response across and close channel
	go func() {
		resp := <-respCh
		done <- resp
		close(done)
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
	}()

	return future, nil
}

func (c *GenericWSClient) readLoop() {
	for {
		select {
		case <-c.closeCh:
			return
		default:
			// proceed to read
		}
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.failAllPending(fmt.Errorf("read error: %w", err))
			_ = c.conn.Close()
			return
		}
		// read raw message
		var resp GenericWSResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			c.logger.Errorw("json unmarshal error", "data", string(data), "err", err)
			continue
		}
		if resp.ID == "" {
			continue
		}
		// find pending
		c.pendingMu.Lock()
		ch, ok := c.pending[resp.ID]
		c.pendingMu.Unlock()
		if !ok {
			c.logger.Errorw("unknown id", "id", resp.ID)
			continue
		}
		c.updateRateLimit(resp)
		select {
		case ch <- resp:
		default:
		}
		close(ch)
	}
}

func (c *GenericWSClient) failAllPending(err error) {
	c.terminateOnce.Do(func() {
		c.terminateErr = err
		close(c.closeCh) // signal Close
		c.pendingMu.Lock()
		defer c.pendingMu.Unlock()
		for id, ch := range c.pending {
			ch <- GenericWSResponse{
				ID:    id,
				Error: &RPCError{Code: -1, Message: err.Error()},
			}
			close(ch)
			delete(c.pending, id)
		}
	})
}

func (c *GenericWSClient) Wait() {
	select {
	case <-c.closeCh:
	}
}

func (c *GenericWSClient) updateRateLimit(resp GenericWSResponse) {
	switch resp.Status {
	case http.StatusTooManyRequests:
		for _, v := range resp.RateLimits {
			if v.Limit == v.Count {
				rt, err := GetRetryAfter(v.Interval, v.IntervalNum)
				if err != nil {
					c.logger.Errorw("get retry after error", "error", err)
					return
				}
				c.blocker.SetRetryAfter(rt)
				return
			}
		}
	case http.StatusTeapot:
		c.blocker.SetRetryAfter(time.UnixMilli(resp.Error.Data.RetryAfter))
	default:
	}
}

func GetRetryAfter(unit string, intervalVal int64) (time.Time, error) {
	switch unit {
	case "SECOND":
		duration := time.Duration(intervalVal) * time.Second
		return time.Now().Add(duration).Truncate(duration), nil
	case "MINUTE":
		duration := time.Duration(intervalVal) * time.Minute
		return time.Now().Add(duration).Truncate(duration), nil
	case "HOUR":
		duration := time.Duration(intervalVal) * time.Hour
		return time.Now().Add(duration).Truncate(duration), nil
	case "DAY":
		duration := time.Duration(intervalVal) * time.Hour * 24
		return time.Now().Add(duration).Truncate(duration), nil
	default:
		return time.Time{}, fmt.Errorf("unknown unit: %s", unit)
	}
}

type RateLimitBlocker struct {
	retryAfter time.Time
	lock       sync.Mutex
}

func (r *RateLimitBlocker) Try() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if time.Now().After(r.retryAfter) {
		return nil
	}
	return fmt.Errorf("retry after %v", r.retryAfter)
}

func (r *RateLimitBlocker) SetRetryAfter(at time.Time) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.retryAfter = at
}

func NewRateLimitBlocker() *RateLimitBlocker {
	return &RateLimitBlocker{}
}

type WSGenericClientSession struct {
	retryBlocker *RateLimitBlocker
	wsURL        string
	apiKey       string
	apiSecret    string
	keyType      string
	header       http.Header
	client       *GenericWSClient
	logger       Logger
	lock         sync.Mutex
}

func NewWSGenericClientSession(wsURL, apiKey, apiSecret, keyType string, header http.Header, rateLimitBlocker *RateLimitBlocker, logger Logger) *WSGenericClientSession {
	return &WSGenericClientSession{
		retryBlocker: rateLimitBlocker,
		wsURL:        wsURL,
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		keyType:      keyType,
		header:       header,
		logger:       logger,
	}
}

func (s *WSGenericClientSession) SendRequest(ctx context.Context, method string, params map[string]interface{}, maxTimeout time.Duration) (GenericWSResponse, error) {
	client := s.getClient()
	if client == nil {
		return GenericWSResponse{}, ErrNoConnection
	}
	return client.SendRequest(ctx, method, params, maxTimeout)
}

func (s *WSGenericClientSession) SendRequestAsync(method string, params map[string]interface{}) (*Future, error) {
	client := s.getClient()
	if client == nil {
		return nil, ErrNoConnection
	}
	return client.SendRequestAsync(method, params)
}

func (s *WSGenericClientSession) Run() {
	for {
		client, err := NewGenericWSClient(s.wsURL, s.apiKey, s.apiSecret, s.keyType, s.header, s.logger, s.retryBlocker)
		if err != nil {
			s.logger.Errorw("new client error", "err", err)
			time.Sleep(time.Second)
			continue
		}
		s.logger.Infow("client connected", "url", s.wsURL)
		s.setClient(client)
		s.client.Wait()
		s.setClient(nil)
	}
}

func (s *WSGenericClientSession) setClient(c *GenericWSClient) {
	s.lock.Lock()
	s.client = c
	s.lock.Unlock()
}

func (s *WSGenericClientSession) getClient() *GenericWSClient {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.client
}
