package healthcheck

import (
	"net/http"
	"time"
)

func IsAlive(address string, timeout time.Duration) (bool, error) {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(address)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false, err
	}

	return true, nil
}
