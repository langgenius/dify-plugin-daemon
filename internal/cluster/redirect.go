package cluster

import (
	"errors"
	"fmt"
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

// constructLocalRedirectUrl constructs a URL for localhost redirection
func constructLocalRedirectUrl(port uint16, request *http.Request) string {
	url := "http://localhost:" + fmt.Sprintf("%d", port) + request.URL.Path
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

// redirectRequestToLocal redirects request to localhost
func redirectRequestToLocal(port uint16, request *http.Request) (int, http.Header, io.ReadCloser, error) {
	url := constructLocalRedirectUrl(port, request)

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
	// If redirecting to current node, use localhost
	if node_id == c.id {
		return redirectRequestToLocal(c.port, request)
	}

	node, ok := c.nodes.Load(node_id)
	if !ok {
		return 0, nil, nil, errors.New("node not found")
	}

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

	// If all IPs failed, try to refresh node information and retry once
	if err := c.updateNodeStatus(); err == nil {
		// Reload node information after update
		if updatedNode, ok := c.nodes.Load(node_id); ok {
			updatedIps := c.SortIps(updatedNode)
			for _, ip := range updatedIps {
				statusCode, header, body, err := redirectRequestToIp(ip, request)
				if err == nil {
					return statusCode, header, body, nil
				}
				lastErr = err
			}
		}
	}

	return 0, nil, nil, lastErr
}
