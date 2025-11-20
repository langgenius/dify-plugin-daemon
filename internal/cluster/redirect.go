package cluster

import (
	"errors"
	"io"
	"net/http"
	"time"
)

func constructRedirectUrl(ip address, request *http.Request) string {
	url := "http://" + ip.fullAddress() + request.URL.Path
	if request.URL.RawQuery != "" {
		url += "?" + request.URL.RawQuery
	}
	return url
}

// basic redirect request
func redirectRequestToIp(ip address, request *http.Request) (int, http.Header, io.ReadCloser, error) {
	url := constructRedirectUrl(ip, request)

	// create a new request
	redirectedRequest, err := http.NewRequest(
		request.Method,
		url,
		request.Body,
	)

	if err != nil {
		return 0, nil, nil, err
	}

	// copy headers
	for key, values := range request.Header {
		for _, value := range values {
			redirectedRequest.Header.Add(key, value)
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(redirectedRequest)

	if err != nil {
		return 0, nil, nil, err
	}

	return resp.StatusCode, resp.Header, resp.Body, nil
}

// RedirectRequest redirects the request to the specified node
func (c *Cluster) RedirectRequest(
	node_id string, request *http.Request,
) (int, http.Header, io.ReadCloser, error) {
	node, ok := c.nodes.Load(node_id)
	if !ok {
		return 0, nil, nil, errors.New("node not found")
	}

	// Sort IPs by voting results to try the most likely healthy address first.
	// See voteAddresses/SortIps for the voting mechanism.
	ips := c.SortIps(node)
	if len(ips) == 0 {
		return 0, nil, nil, errors.New("no available ip found")
	}

	// Try each IP until we find a working one
	var lastErr error
	for _, ip := range ips {
		statusCode, header, body, err := redirectRequestToIp(ip, request)
		if err == nil {
			return statusCode, header, body, nil
		}
		lastErr = err
	}

	return 0, nil, nil, lastErr
}
