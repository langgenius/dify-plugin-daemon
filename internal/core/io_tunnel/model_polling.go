package io_tunnel

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func StartPolling(
	session *session_manager.Session,
	request *requests.RequestStartPolling,
) (*model_entities.ModelPollingResult, error) {
	return invokeSingleResponse[requests.RequestStartPolling, model_entities.ModelPollingResult](session, request)
}

func CheckPolling(
	session *session_manager.Session,
	request *requests.RequestCheckPolling,
) (*model_entities.ModelPollingResult, error) {
	return invokeSingleResponse[requests.RequestCheckPolling, model_entities.ModelPollingResult](session, request)
}

func invokeSingleResponse[Req any, Rsp any](
	session *session_manager.Session,
	request *Req,
) (*Rsp, error) {
	response, err := GenericInvokePlugin[Req, Rsp](session, request, 1)
	if err != nil {
		return nil, err
	}
	defer response.Close()

	var (
		got  bool
		item Rsp
	)

	for response.Next() {
		value, readErr := response.Read()
		if readErr != nil {
			return nil, readErr
		}
		item = value
		got = true
	}

	if !got {
		return nil, errors.New("plugin returned empty polling response")
	}

	return &item, nil
}
