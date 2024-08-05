package cluster

import (
	"errors"
	"io"
	"net/http"
	"strconv"
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
	redirected_request, err := http.NewRequest(
		request.Method,
		"http://"+ip.Address+":"+strconv.FormatUint(uint64(c.port), 10)+request.URL.Path,
		request.Body,
	)

	if err != nil {
		return 0, nil, nil, err
	}

	// copy headers
	for key, values := range request.Header {
		for _, value := range values {
			redirected_request.Header.Add(key, value)
		}
	}

	client := http.DefaultClient
	resp, err := client.Do(redirected_request)

	if err != nil {
		return 0, nil, nil, err
	}

	return resp.StatusCode, resp.Header, resp.Body, nil
}
