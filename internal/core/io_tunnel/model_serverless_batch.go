package io_tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/serverless_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

type serverlessRequestBytesLimiter interface {
	MaxServerlessRequestBytes() int
}

type textRequestBatch[Request any] struct {
	request         Request
	payloadBytes    int
	maxRequestBytes int
	itemOffset      int
	itemCount       int
}

type serverlessBatchResultLog struct {
	batchIndex      int
	batchCount      int
	itemOffset      int
	itemCount       int
	payloadBytes    int
	maxRequestBytes int
	startedAt       time.Time
	outcome         string
	errorType       string
}

type serverlessBatchResultMismatchError struct {
	Action        string
	BatchIndex    int
	ExpectedItems int
	ActualItems   int
	ResponseCount int
	Reason        string
}

func (e *serverlessBatchResultMismatchError) Error() string {
	message := fmt.Sprintf(
		"ServerlessBatchResultMismatch: action=%s batch_index=%d expected_items=%d actual_items=%d response_count=%d",
		e.Action,
		e.BatchIndex,
		e.ExpectedItems,
		e.ActualItems,
		e.ResponseCount,
	)
	if e.Reason != "" {
		message += " reason=" + e.Reason
	}
	return message
}

func invokeTextEmbeddingServerlessAware(
	session *session_manager.Session,
	request *requests.RequestInvokeTextEmbedding,
) (*stream.Stream[model_entities.TextEmbeddingResult], error) {
	runtime := session.Runtime()
	if runtime == nil || runtime.Type() != plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS {
		return GenericInvokePlugin[requests.RequestInvokeTextEmbedding, model_entities.TextEmbeddingResult](
			session,
			request,
			1,
		)
	}

	limiter, ok := runtime.(serverlessRequestBytesLimiter)
	if !ok || limiter.MaxServerlessRequestBytes() <= 0 {
		return nil, errors.New("serverless runtime request byte limit is unavailable")
	}
	maxRequestBytes := limiter.MaxServerlessRequestBytes()
	batches, err := splitTextRequest(
		session,
		request.Texts,
		maxRequestBytes,
		func(start, end int) requests.RequestInvokeTextEmbedding {
			candidate := *request
			candidate.Texts = request.Texts[start:end]
			return candidate
		},
	)
	if err != nil {
		return nil, err
	}
	if len(batches) <= 1 {
		return GenericInvokePlugin[requests.RequestInvokeTextEmbedding, model_entities.TextEmbeddingResult](
			session,
			request,
			1,
		)
	}
	observeServerlessBatchPlan(session, len(request.Texts), maxRequestBytes, batches)

	response := stream.NewStream[model_entities.TextEmbeddingResult](1)
	recorder := newPluginInvocationRecorder(session)
	outcome := &pluginInvocationOutcomeTracker{}
	ctx, cancel := context.WithCancel(session.RequestContext())
	response.OnClose(cancel)
	response.OnClose(func() {
		recorder.record(outcome.outcome())
	})

	go coordinateTextEmbeddingBatches(ctx, session, request, batches, response, outcome)
	return response, nil
}

func getTextEmbeddingNumTokensServerlessAware(
	session *session_manager.Session,
	request *requests.RequestGetTextEmbeddingNumTokens,
) (*stream.Stream[model_entities.GetTextEmbeddingNumTokensResponse], error) {
	runtime := session.Runtime()
	if runtime == nil || runtime.Type() != plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS {
		return GenericInvokePlugin[requests.RequestGetTextEmbeddingNumTokens, model_entities.GetTextEmbeddingNumTokensResponse](
			session,
			request,
			1,
		)
	}

	limiter, ok := runtime.(serverlessRequestBytesLimiter)
	if !ok || limiter.MaxServerlessRequestBytes() <= 0 {
		return nil, errors.New("serverless runtime request byte limit is unavailable")
	}
	maxRequestBytes := limiter.MaxServerlessRequestBytes()
	batches, err := splitTokenCountRequest(session, request, maxRequestBytes)
	if err != nil {
		return nil, err
	}
	if len(batches) <= 1 {
		return GenericInvokePlugin[requests.RequestGetTextEmbeddingNumTokens, model_entities.GetTextEmbeddingNumTokensResponse](
			session,
			request,
			1,
		)
	}
	observeServerlessBatchPlan(session, len(request.Texts), maxRequestBytes, batches)

	response := stream.NewStream[model_entities.GetTextEmbeddingNumTokensResponse](1)
	recorder := newPluginInvocationRecorder(session)
	outcome := &pluginInvocationOutcomeTracker{}
	ctx, cancel := context.WithCancel(session.RequestContext())
	response.OnClose(cancel)
	response.OnClose(func() {
		recorder.record(outcome.outcome())
	})

	go coordinateTokenCountBatches(ctx, session, request, batches, response, outcome)
	return response, nil
}

func splitTokenCountRequest(
	session *session_manager.Session,
	request *requests.RequestGetTextEmbeddingNumTokens,
	maxRequestBytes int,
) ([]textRequestBatch[requests.RequestGetTextEmbeddingNumTokens], error) {
	return splitTextRequest(
		session,
		request.Texts,
		maxRequestBytes,
		func(start, end int) requests.RequestGetTextEmbeddingNumTokens {
			return tokenCountRequestSlice(request, start, end)
		},
	)
}

func splitTextRequest[Request any](
	session *session_manager.Session,
	texts []string,
	maxRequestBytes int,
	sliceRequest func(start, end int) Request,
) ([]textRequestBatch[Request], error) {
	emptyRequest := sliceRequest(0, 0)
	emptyPayloadBytes := requestPayloadBytes(session, &emptyRequest)
	if emptyPayloadBytes > maxRequestBytes {
		return nil, &serverless_runtime.ServerlessPayloadTooLargeError{
			Action:          session.Action,
			PayloadBytes:    emptyPayloadBytes,
			MaxRequestBytes: maxRequestBytes,
		}
	}

	batches := make([]textRequestBatch[Request], 0)
	start := 0
	estimatedPayloadBytes := emptyPayloadBytes
	for itemIndex, text := range texts {
		encodedText, err := json.Marshal(text)
		if err != nil {
			return nil, fmt.Errorf("marshal text at item_index=%d: %w", itemIndex, err)
		}

		nextPayloadBytes := estimatedPayloadBytes + len(encodedText)
		if itemIndex > start {
			nextPayloadBytes++
		}
		if nextPayloadBytes <= maxRequestBytes {
			estimatedPayloadBytes = nextPayloadBytes
			continue
		}

		if itemIndex == start {
			return nil, oversizedTextItemError(session, sliceRequest, itemIndex, maxRequestBytes)
		}
		batchRequest := sliceRequest(start, itemIndex)
		payloadBytes := requestPayloadBytes(session, &batchRequest)
		if payloadBytes > maxRequestBytes {
			return nil, fmt.Errorf(
				"serverless batch size estimate mismatch: payload_bytes=%d max_request_bytes=%d",
				payloadBytes,
				maxRequestBytes,
			)
		}
		batches = append(batches, textRequestBatch[Request]{
			request:         batchRequest,
			payloadBytes:    payloadBytes,
			maxRequestBytes: maxRequestBytes,
			itemOffset:      start,
			itemCount:       itemIndex - start,
		})
		start = itemIndex
		estimatedPayloadBytes = emptyPayloadBytes + len(encodedText)
		if estimatedPayloadBytes > maxRequestBytes {
			return nil, oversizedTextItemError(session, sliceRequest, itemIndex, maxRequestBytes)
		}
	}

	if start < len(texts) {
		batchRequest := sliceRequest(start, len(texts))
		payloadBytes := requestPayloadBytes(session, &batchRequest)
		if payloadBytes > maxRequestBytes {
			return nil, fmt.Errorf(
				"serverless batch size estimate mismatch: payload_bytes=%d max_request_bytes=%d",
				payloadBytes,
				maxRequestBytes,
			)
		}
		batches = append(batches, textRequestBatch[Request]{
			request:         batchRequest,
			payloadBytes:    payloadBytes,
			maxRequestBytes: maxRequestBytes,
			itemOffset:      start,
			itemCount:       len(texts) - start,
		})
	}

	return batches, nil
}

func observeServerlessBatchPlan[Request any](
	session *session_manager.Session,
	originalItemCount int,
	maxRequestBytes int,
	batches []textRequestBatch[Request],
) {
	payloadBytes := make([]int, len(batches))
	for i := range batches {
		payloadBytes[i] = batches[i].payloadBytes
	}
	recordServerlessBatchPlan(session, payloadBytes)
	log.InfoContext(
		session.RequestContext(),
		"serverless text request split into serial batches",
		"runtime_type", plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		"action", session.Action,
		"session_id", session.ID,
		"original_item_count", originalItemCount,
		"batch_count", len(batches),
		"max_request_bytes", maxRequestBytes,
	)
}

func logServerlessBatchResult(
	session *session_manager.Session,
	result serverlessBatchResultLog,
) {
	log.InfoContext(
		session.RequestContext(),
		"serverless text batch completed",
		"runtime_type", plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		"action", session.Action,
		"session_id", session.ID,
		"batch_count", result.batchCount,
		"batch_index", result.batchIndex,
		"batch_item_offset", result.itemOffset,
		"batch_item_count", result.itemCount,
		"payload_bytes", result.payloadBytes,
		"max_request_bytes", result.maxRequestBytes,
		"duration_ms", time.Since(result.startedAt).Milliseconds(),
		"outcome", result.outcome,
		"error_type", result.errorType,
	)
}

func oversizedTextItemError[Request any](
	session *session_manager.Session,
	sliceRequest func(start, end int) Request,
	itemIndex int,
	maxRequestBytes int,
) error {
	singleRequest := sliceRequest(itemIndex, itemIndex+1)
	payloadBytes := requestPayloadBytes(session, &singleRequest)
	recordServerlessOversizeItem(session)
	log.WarnContext(
		session.RequestContext(),
		"serverless text item exceeds request payload limit",
		"runtime_type", plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		"action", session.Action,
		"session_id", session.ID,
		"item_index", itemIndex,
		"payload_bytes", payloadBytes,
		"max_request_bytes", maxRequestBytes,
		"error_type", "ServerlessPayloadTooLarge",
	)
	return &serverless_runtime.ServerlessPayloadTooLargeError{
		Action:          session.Action,
		PayloadBytes:    payloadBytes,
		MaxRequestBytes: maxRequestBytes,
		ItemIndex:       &itemIndex,
	}
}

func requestPayloadBytes[Request any](
	session *session_manager.Session,
	request *Request,
) int {
	return len(session.Message(
		session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
		GetInvokePluginMap(session, request),
	))
}

func tokenCountRequestSlice(
	request *requests.RequestGetTextEmbeddingNumTokens,
	start int,
	end int,
) requests.RequestGetTextEmbeddingNumTokens {
	candidate := *request
	candidate.Texts = request.Texts[start:end]
	return candidate
}

func processServerlessBatches[Request any, Response any](
	ctx context.Context,
	session *session_manager.Session,
	batches []textRequestBatch[Request],
	consume func(batchIndex int, itemCount int, results []Response) error,
) error {
	for batchIndex := range batches {
		if err := ctx.Err(); err != nil {
			return err
		}

		batch := &batches[batchIndex]
		batchStartedAt := time.Now()
		batchResults, err := invokePluginBatch[Request, Response](ctx, session, &batch.request)
		if err == nil {
			err = consume(batchIndex, batch.itemCount, batchResults)
		}
		if err != nil {
			outcome, errorType := serverlessBatchErrorDetails(err)
			var mismatch *serverlessBatchResultMismatchError
			if errors.As(err, &mismatch) {
				recordServerlessBatchMismatch(session, mismatch.Reason)
			}
			logServerlessBatchResult(
				session,
				serverlessBatchResultLog{
					batchIndex:      batchIndex,
					batchCount:      len(batches),
					itemOffset:      batch.itemOffset,
					itemCount:       batch.itemCount,
					payloadBytes:    batch.payloadBytes,
					maxRequestBytes: batch.maxRequestBytes,
					startedAt:       batchStartedAt,
					outcome:         outcome,
					errorType:       errorType,
				},
			)
			return err
		}

		logServerlessBatchResult(
			session,
			serverlessBatchResultLog{
				batchIndex:      batchIndex,
				batchCount:      len(batches),
				itemOffset:      batch.itemOffset,
				itemCount:       batch.itemCount,
				payloadBytes:    batch.payloadBytes,
				maxRequestBytes: batch.maxRequestBytes,
				startedAt:       batchStartedAt,
				outcome:         pluginInvocationOutcomeSuccess,
			},
		)
	}
	return nil
}

func serverlessBatchErrorDetails(err error) (outcome string, errorType string) {
	switch {
	case errors.Is(err, context.Canceled):
		return pluginInvocationOutcomeCanceled, "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return pluginInvocationOutcomeCanceled, "deadline_exceeded"
	}

	var mismatch *serverlessBatchResultMismatchError
	if errors.As(err, &mismatch) {
		return pluginInvocationOutcomeError, "ServerlessBatchResultMismatch"
	}
	return pluginInvocationOutcomeError, "runtime_error"
}

func closeServerlessBatchResponse[Response any](
	ctx context.Context,
	response *stream.Stream[Response],
	outcome *pluginInvocationOutcomeTracker,
	err error,
) {
	if ctx.Err() == nil {
		outcome.markError()
	}
	response.CloseWithError(err)
}

func coordinateTokenCountBatches(
	ctx context.Context,
	session *session_manager.Session,
	originalRequest *requests.RequestGetTextEmbeddingNumTokens,
	batches []textRequestBatch[requests.RequestGetTextEmbeddingNumTokens],
	response *stream.Stream[model_entities.GetTextEmbeddingNumTokensResponse],
	outcome *pluginInvocationOutcomeTracker,
) {
	mergedNumTokens := make([]int, 0, len(originalRequest.Texts))
	err := processServerlessBatches(
		ctx,
		session,
		batches,
		func(
			batchIndex int,
			itemCount int,
			batchResults []model_entities.GetTextEmbeddingNumTokensResponse,
		) error {
			if len(batchResults) == 1 && len(batchResults[0].NumTokens) == itemCount {
				mergedNumTokens = append(mergedNumTokens, batchResults[0].NumTokens...)
				return nil
			}

			actualItems := 0
			if len(batchResults) == 1 {
				actualItems = len(batchResults[0].NumTokens)
			}
			return &serverlessBatchResultMismatchError{
				Action:        string(session.Action),
				BatchIndex:    batchIndex,
				ExpectedItems: itemCount,
				ActualItems:   actualItems,
				ResponseCount: len(batchResults),
				Reason:        "token_count",
			}
		},
	)
	if err != nil {
		closeServerlessBatchResponse(ctx, response, outcome, err)
		return
	}

	if len(mergedNumTokens) != len(originalRequest.Texts) {
		recordServerlessBatchMismatch(session, "final_token_count")
		outcome.markError()
		response.CloseWithError(&serverlessBatchResultMismatchError{
			Action:        string(session.Action),
			BatchIndex:    len(batches),
			ExpectedItems: len(originalRequest.Texts),
			ActualItems:   len(mergedNumTokens),
			ResponseCount: 1,
		})
		return
	}

	response.WriteBlocking(model_entities.GetTextEmbeddingNumTokensResponse{
		NumTokens: mergedNumTokens,
	})
	if ctx.Err() != nil {
		response.Close()
		return
	}
	outcome.markSuccess()
	response.Close()
}

func coordinateTextEmbeddingBatches(
	ctx context.Context,
	session *session_manager.Session,
	originalRequest *requests.RequestInvokeTextEmbedding,
	batches []textRequestBatch[requests.RequestInvokeTextEmbedding],
	response *stream.Stream[model_entities.TextEmbeddingResult],
	outcome *pluginInvocationOutcomeTracker,
) {
	var mergedResult *model_entities.TextEmbeddingResult
	err := processServerlessBatches(
		ctx,
		session,
		batches,
		func(batchIndex int, itemCount int, batchResults []model_entities.TextEmbeddingResult) error {
			if len(batchResults) != 1 {
				return &serverlessBatchResultMismatchError{
					Action:        string(session.Action),
					BatchIndex:    batchIndex,
					ExpectedItems: itemCount,
					ResponseCount: len(batchResults),
					Reason:        "response_count",
				}
			}
			return mergeTextEmbeddingResult(
				session,
				batchIndex,
				itemCount,
				&mergedResult,
				batchResults[0],
			)
		},
	)
	if err != nil {
		closeServerlessBatchResponse(ctx, response, outcome, err)
		return
	}

	if mergedResult == nil || len(mergedResult.Embeddings) != len(originalRequest.Texts) {
		actualItems := 0
		if mergedResult != nil {
			actualItems = len(mergedResult.Embeddings)
		}
		recordServerlessBatchMismatch(session, "final_embedding_count")
		outcome.markError()
		response.CloseWithError(&serverlessBatchResultMismatchError{
			Action:        string(session.Action),
			BatchIndex:    len(batches),
			ExpectedItems: len(originalRequest.Texts),
			ActualItems:   actualItems,
			ResponseCount: 1,
			Reason:        "final_embedding_count",
		})
		return
	}

	response.WriteBlocking(*mergedResult)
	if ctx.Err() != nil {
		response.Close()
		return
	}
	outcome.markSuccess()
	response.Close()
}

func mergeTextEmbeddingResult(
	session *session_manager.Session,
	batchIndex int,
	expectedItems int,
	mergedResult **model_entities.TextEmbeddingResult,
	batchResult model_entities.TextEmbeddingResult,
) error {
	if len(batchResult.Embeddings) != expectedItems {
		return &serverlessBatchResultMismatchError{
			Action:        string(session.Action),
			BatchIndex:    batchIndex,
			ExpectedItems: expectedItems,
			ActualItems:   len(batchResult.Embeddings),
			ResponseCount: 1,
			Reason:        "embedding_count",
		}
	}

	usage := batchResult.Usage
	if usage.Tokens == nil ||
		usage.TotalTokens == nil ||
		usage.Currency == nil ||
		usage.Latency == nil {
		return &serverlessBatchResultMismatchError{
			Action:        string(session.Action),
			BatchIndex:    batchIndex,
			ExpectedItems: expectedItems,
			ActualItems:   len(batchResult.Embeddings),
			ResponseCount: 1,
			Reason:        "missing_usage_field",
		}
	}

	if *mergedResult == nil {
		tokens := *usage.Tokens
		totalTokens := *usage.TotalTokens
		currency := *usage.Currency
		latency := *usage.Latency
		result := batchResult
		result.Embeddings = append([][]float64(nil), batchResult.Embeddings...)
		result.Usage.Tokens = &tokens
		result.Usage.TotalTokens = &totalTokens
		result.Usage.Currency = &currency
		result.Usage.Latency = &latency
		*mergedResult = &result
		return nil
	}

	merged := *mergedResult
	if batchResult.Model != merged.Model {
		return embeddingBatchMismatch(session, batchIndex, expectedItems, "model")
	}
	if !usage.UnitPrice.Equal(merged.Usage.UnitPrice) {
		return embeddingBatchMismatch(session, batchIndex, expectedItems, "unit_price")
	}
	if !usage.PriceUnit.Equal(merged.Usage.PriceUnit) {
		return embeddingBatchMismatch(session, batchIndex, expectedItems, "price_unit")
	}
	if *usage.Currency != *merged.Usage.Currency {
		return embeddingBatchMismatch(session, batchIndex, expectedItems, "currency")
	}

	merged.Embeddings = append(merged.Embeddings, batchResult.Embeddings...)
	*merged.Usage.Tokens += *usage.Tokens
	*merged.Usage.TotalTokens += *usage.TotalTokens
	merged.Usage.TotalPrice = merged.Usage.TotalPrice.Add(usage.TotalPrice)
	*merged.Usage.Latency += *usage.Latency
	return nil
}

func embeddingBatchMismatch(
	session *session_manager.Session,
	batchIndex int,
	expectedItems int,
	reason string,
) error {
	return &serverlessBatchResultMismatchError{
		Action:        string(session.Action),
		BatchIndex:    batchIndex,
		ExpectedItems: expectedItems,
		ActualItems:   expectedItems,
		ResponseCount: 1,
		Reason:        reason,
	}
}

func invokePluginBatch[Request any, Response any](
	ctx context.Context,
	session *session_manager.Session,
	request *Request,
) ([]Response, error) {
	batchOutcome := &pluginInvocationOutcomeTracker{}
	response, err := invokePluginOnce[Request, Response](ctx, session, request, 1, batchOutcome, nil)
	if err != nil {
		return nil, err
	}
	return readPluginBatch(ctx, response)
}

func readPluginBatch[Response any](
	ctx context.Context,
	response *stream.Stream[Response],
) ([]Response, error) {
	stopCloseOnCancel := context.AfterFunc(ctx, response.Close)
	defer stopCloseOnCancel()

	var results []Response
	for response.Next() {
		result, err := response.Read()
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
