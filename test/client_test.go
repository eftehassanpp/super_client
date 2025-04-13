package test

import (
	"testing"

	"github.com/eftehassanpp/super_client"
)

func Test(t *testing.T) {
	client, err := super_client.NewSuperClient(&super_client.ClientConfig{})
	if err != nil {
		return
	}
	var output struct {
		IP   string `json:"ip"`
		City string `jons:"city"`
	}
	response := client.Get("https://ipinfo.io/json", nil, nil, &output)
	t.Log(response.StatusCode)
	t.Log(output)
}
