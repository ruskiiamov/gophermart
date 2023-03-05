package accrualsystem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ruskiiamov/gophermart/internal/logger"
)

const (
	timeout           = 1 * time.Second
	retryAfterHeader  = "Retry-After"
	defaultRetryAfter = 2
)

type Order struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type ErrNotAvailable struct {
	AvailableSince time.Time
}

func (e *ErrNotAvailable) Error() string {
	return fmt.Sprintf("accrual system not available until %s", e.AvailableSince.Format(time.RFC3339))
}

type Connector struct {
	address        string
	client         http.Client
	availableSince time.Time
}

func NewConnector(address string) *Connector {
	return &Connector{
		address:        address,
		client:         http.Client{Timeout: timeout},
		availableSince: time.Now(),
	}
}

func (c *Connector) GetAccrual(ctx context.Context, orderID int) (status string, accrual int, err error) {
	if time.Now().Before(c.availableSince) {
		return "", 0, &ErrNotAvailable{AvailableSince: c.availableSince}
	}

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
		retryAfter := getRetryAfter(response)
		c.availableSince = time.Now().Add(time.Duration(retryAfter) * time.Second)
		return "", 0, &ErrNotAvailable{AvailableSince: c.availableSince}
	}

	if response.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("get accrual request not ok: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
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

func getRetryAfter(response *http.Response) int {
	retryAfterValue := response.Header.Get(retryAfterHeader)
	if retryAfterValue == "" {
		return defaultRetryAfter
	}

	retryAfter, err := strconv.Atoi(retryAfterValue)
	if err != nil {
		logger.Error(fmt.Sprintf("retry-after atoi error: %s", err))
		return defaultRetryAfter
	}

	return retryAfter
}
