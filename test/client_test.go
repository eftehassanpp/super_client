package test

import (
	"testing"

	"github.com/eftehassanpp/super_client"
)

func Test(t *testing.T) {
	client, err := super_client.NewSuperClient(&super_client.ClientConfig{BaseProxy: "http://SofagJ-cc-ES-pool-p2p-sessionid-change:ZHlQm1@proxy.zettaproxies.io:8888"})
	if err != nil {
		return
	}
	// client.SetProxy("http://SofagJ-cc-ES-pool-p2p-sessionid-WpKgt6P6barF:ZHlQm1@proxy.zettaproxies.io:8888")
	t.Logf("Current Proxy: %s", client.GetProxy())
	err = client.SetNewProxy()
	if err != nil {
		t.Log(err)
	}
	var output struct {
		IP   string `json:"ip"`
		City string `json:"city"`
	}
	response := client.Get("https://ipinfo.io/json", nil, nil, &output)
	t.Log(response.StatusCode)
	t.Log(output)
	t.Logf("Current Proxy: %s", client.GetProxy())
}
