package binance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type CreateSubscriptionToken func() (*MarginSubscriptionToken, error)

// WsApiMarginUserDataServe only accepts Ed25519 key type.
func WsApiMarginUserDataServe(apiKey, secretKey string, createSubscription CreateSubscriptionToken, logger Logger, handler WsUserDataHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	initSubscription, err := createSubscription()
	if err != nil {
		return nil, nil, err
	}
	sessionTerminated := make(chan struct{})
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		event := new(WsUserDataEvent)

		err = json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}

		switch UserDataEventType(j.Get("e").MustString()) {
		case UserDataEventTypeOutboundAccountPosition:
			err = json.Unmarshal(message, &event.AccountUpdate)
			if err != nil {
				errHandler(err)
				return
			}
		case UserDataEventTypeBalanceUpdate:
			err = json.Unmarshal(message, &event.BalanceUpdate)
			if err != nil {
				errHandler(err)
				return
			}
		case UserDataEventTypeExecutionReport:
			err = json.Unmarshal(message, &event.OrderUpdate)
			if err != nil {
				errHandler(err)
				return
			}
		case UserDataEventTypeListStatus:
			err = json.Unmarshal(message, &event.OCOUpdate)
			if err != nil {
				errHandler(err)
				return
			}
		case EventStreamTerminated:
			errHandler(errors.New("session terminated"))
			close(sessionTerminated)
		}
		handler(event)
	}

	var wsc *WsApiService
	if wsc, err = NewWsApiService(apiKey, secretKey); err != nil {
		return
	}
	doneC, stopC, err = wsc.wsServe(wsHandler, errHandler)
	if err != nil {
		return
	}
	var (
		subRes UserMarginStreamSubscriptionResponse
	)
	if subRes, err = wsc.subscribeUserMarginDataStream(context.Background(), initSubscription.Token); err != nil {
		stopC <- struct{}{}
		return nil, doneC, err
	}
	logger.Infow("session valid", "until", time.UnixMilli(initSubscription.ExpirationTime))
	tilExpire := time.Until(time.UnixMilli(initSubscription.ExpirationTime)) - time.Minute*5
	t := time.NewTimer(tilExpire)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-doneC:
				return
			case <-sessionTerminated:
				stopC <- struct{}{}
				return
			case <-t.C:
				newSub, err := createSubscription()
				if err != nil {
					errHandler(fmt.Errorf("create new session subscription: %w", err))
					stopC <- struct{}{}
					return
				}
				subRes, err = wsc.subscribeUserMarginDataStream(context.Background(), newSub.Token)
				if err != nil {
					errHandler(fmt.Errorf("extend session failed: %w", err))
					stopC <- struct{}{}
					return
				}
				tilExpire := time.Until(time.UnixMilli(subRes.Result.ExpirationTime)) - time.Minute*5
				t.Reset(tilExpire)
				logger.Infow("extend session", "dur", tilExpire, "to", time.UnixMilli(subRes.Result.ExpirationTime))
			}
		}
	}()
	return
}
