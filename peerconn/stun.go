package peerconn

import (
	"io/ioutil"
	"net/http"
)

// GetDefaultStunHosts ...
func GetDefaultStunHosts() (string, error) {
	resp, err := http.Get("https://signaling.irieda.com/stun")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
