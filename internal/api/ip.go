package api

import (
	"net"
	"net/http"
	"strings"
)

func GetIpFromRequest(r *http.Request) (string, error) {
	userIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	userIP = strings.Replace(userIP, ":", "_", -1)
	return userIP, nil
}
