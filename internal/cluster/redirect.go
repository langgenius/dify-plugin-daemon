package cluster

import (
	"errors"
	"io"
	"net/http"
)

// RedirectRequest redirects the request to the specified node
func (c *Cluster) RedirectRequest(
	node_id string, request *http.Request,
) (int, http.Header, io.ReadCloser, error) {
	node, ok := c.nodes.Load(node_id)
	if !ok {
		return 0, nil, nil, errors.New("node not found")
	}

	ips := c.SortIps(node)
	if len(ips) == 0 {
		return 0, nil, nil, errors.New("no available ip found")
	}

	ip := ips[0]

	// create a new request
	redirectedRequest, err := http.NewRequest(
		request.Method,
		"http://"+ip.fullAddress()+request.URL.Path,
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

	client := http.DefaultClient
	resp, err := client.Do(redirectedRequest)

	if err != nil {
		return 0, nil, nil, err
	}

	return resp.StatusCode, resp.Header, resp.Body, nil
}
