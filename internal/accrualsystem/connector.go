package accrualsystem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/bonus"
)

const (
	timeout           = 1 * time.Second
	retryAfterHeader  = "Retry-After"
	defaultRetryAfter = 2
)

type Order struct {
	Order   int     `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

var _ bonus.AccrualSystemConnector = (*connector)(nil)

type connector struct {
	address string
	client  http.Client
	lock    chan struct{}
	mu      sync.Mutex
}

func NewConnector(address string) *connector {
	lock := make(chan struct{})
	close(lock)

	return &connector{
		address: address,
		client: http.Client{
			Timeout: timeout,
		},
		lock: lock,
	}
}

func (c *connector) GetAccrual(ctx context.Context, orderID int) (status string, accrual int, err error) {
	<-c.lock

	url := c.address + "/api/orders/" + strconv.Itoa(orderID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("new request error: %w", err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return "", 0, fmt.Errorf("do request error: %w", err)
	}

	if response.StatusCode == http.StatusTooManyRequests {
		c.lockByTooManyRequests(response)
		return "", 0, bonus.ErrAccrualNotReady
	}

	if response.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("get accrual request not ok: %s", response.Status)
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return "", 0, fmt.Errorf("read body error: %w", err)
	}

	var order Order
	err = json.Unmarshal(body, &order)
	if err != nil {
		return "", 0, fmt.Errorf("json unmarshall error: %w", err)
	}

	accrual = int(order.Accrual * 100)

	return order.Status, accrual, nil
}

func (c *connector) lockByTooManyRequests(response *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.lock:
		var retryAfter int
		var err error
		retryAfterValue := response.Header.Get(retryAfterHeader)
		if retryAfterValue != "" {
			retryAfter, err = strconv.Atoi(retryAfterValue)
			if err != nil {
				log.Error().Msgf("retry-after atoi error: %s", err)
				retryAfter = defaultRetryAfter
			}
		} else {
			retryAfter = defaultRetryAfter
		}
		c.lock = make(chan struct{})
		time.AfterFunc((time.Duration(retryAfter) * time.Second), func() {
			close(c.lock)
		})
	default:
	}
}
