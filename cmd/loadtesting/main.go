package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/tester"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/generic_invoke"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/test_utils"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

var (
	flagSupportAutoScale    = flag.Bool("auto-scale", false, "support auto scale")
	flagMaxInstances        = flag.Int("max-instances", 1, "max instances")
	flagMinInstances        = flag.Int("min-instances", 1, "min instances")
	flagStartConcurrency    = flag.Int("start-concurrency", 10, "starting concurrency level")
	flagMaxConcurrency      = flag.Int("max-concurrency", 1000, "maximum concurrency level to test")
	flagConcurrencyStep     = flag.Int("concurrency-step", 10, "step to increase concurrency in each round")
	flagStabilizationTime   = flag.Int("stabilization-time", 5, "seconds to wait for system stabilization between levels")
	flagDeclineThreshold    = flag.Float64("decline-threshold", 0.1, "threshold for throughput decline to determine max capacity")
	flagConsecutiveDeclines = flag.Int("consecutive-declines", 3, "number of consecutive declines to confirm max capacity")
	flagShowLog             = flag.Bool("show-log", true, "show log")
	flagMaxLatency          = flag.Int64("max-latency", 0, "maximum acceptable latency in milliseconds (0 means no limit)")
	flagMaxErrorRate        = flag.Float64("max-error-rate", 0, "maximum acceptable error rate (0-100, 0 means no limit)")
	flagPlateau             = flag.Int("plateau", 3, "number of consecutive rounds with no throughput improvement to stop testing")
)

type ConcurrencyResult struct {
	Concurrency int     // concurrency level
	Throughput  float64 // requests per second
	AvgLatency  int64   // average latency in milliseconds
	ErrorRate   float64 // percentage of failed requests
}

func main() {
	flag.Parse()

	if !*flagShowLog {
		log.SetShowLog(false)
	}

	routine.InitPool(100000)

	runtime, err := test_utils.GetOpenAIRuntime(*flagSupportAutoScale, *flagMaxInstances, *flagMinInstances)
	if err != nil {
		log.Panic("Failed to get runtime: ", err)
	}
	defer runtime.Stop()

	port, cancel := test_utils.StartFakeOpenAIServer()
	defer cancel()

	// wait for 10 seconds for the auto scale to be ready
	time.Sleep(10 * time.Second)

	results := []ConcurrencyResult{}
	declineCount := 0
	plateauCount := 0
	maxThroughput := 0.0
	optimalConcurrency := *flagStartConcurrency

	fmt.Printf("\nConcurrency Testing Results with %d instances:\n", *flagMaxInstances)
	fmt.Println("----------------------------------------------------------")
	fmt.Printf("| %-12s | %-12s | %-12s | %-12s |\n", "Concurrency", "Throughput", "Avg Latency", "Error Rate")
	fmt.Println("----------------------------------------------------------")

	// Test increasing concurrency levels
	for concurrency := *flagStartConcurrency; concurrency <= *flagMaxConcurrency; concurrency += *flagConcurrencyStep {
		// Run test at current concurrency level
		result := runConcurrencyTest(runtime, port, concurrency, concurrency*10)
		results = append(results, result)

		// Print result in table format
		fmt.Printf("| %-12d | %-12.2f | %-12d | %-12.2f |\n",
			result.Concurrency, result.Throughput, result.AvgLatency, result.ErrorRate)

		// Check if we've reached maximum acceptable latency
		if *flagMaxLatency > 0 && result.AvgLatency > *flagMaxLatency {
			fmt.Printf("\nStopping test: Maximum acceptable latency (%d ms) exceeded\n", *flagMaxLatency)
			break
		}

		// Check if we've reached maximum acceptable error rate
		if *flagMaxErrorRate > 0 && result.ErrorRate > *flagMaxErrorRate {
			fmt.Printf("\nStopping test: Maximum acceptable error rate (%.2f%%) exceeded\n", *flagMaxErrorRate)
			break
		}

		// Check if we've found the optimal concurrency
		if len(results) > 1 {
			prevThroughput := results[len(results)-2].Throughput

			// If throughput declined by more than the threshold
			if result.Throughput < prevThroughput*(1-*flagDeclineThreshold) {
				declineCount++
				if declineCount >= *flagConsecutiveDeclines {
					// We've confirmed the decline, optimal was before the declines started
					optimalConcurrency = results[len(results)-declineCount-1].Concurrency
					maxThroughput = results[len(results)-declineCount-1].Throughput
					fmt.Println("\nStopping test: Consistent throughput decline detected")
					break
				}
			} else {
				// Reset decline count if throughput improved
				declineCount = 0

				// Check for plateau (no significant improvement)
				if result.Throughput <= maxThroughput*1.02 {
					plateauCount++
					if plateauCount >= *flagPlateau {
						fmt.Println("\nStopping test: Throughput has plateaued")
						break
					}
				} else {
					plateauCount = 0
				}

				// Update max throughput if current is better
				if result.Throughput > maxThroughput {
					maxThroughput = result.Throughput
					optimalConcurrency = concurrency
				}
			}
		} else {
			maxThroughput = result.Throughput
		}

		// Wait for system to stabilize before next test
		time.Sleep(time.Duration(*flagStabilizationTime) * time.Second)
	}

	fmt.Println("----------------------------------------------------------")
	fmt.Printf("\nOptimal Concurrency: %d with Throughput: %.2f req/s\n\n", optimalConcurrency, maxThroughput)

	// Generate summary of results
	fmt.Println("Performance Summary:")
	fmt.Println("As concurrency increases:")

	// Find patterns in the data
	latencyTrend := analyzeTrend(results, func(r ConcurrencyResult) float64 { return float64(r.AvgLatency) })
	throughputTrend := analyzeTrend(results, func(r ConcurrencyResult) float64 { return r.Throughput })

	fmt.Printf("- Latency: %s\n", latencyTrend)
	fmt.Printf("- Throughput: %s\n", throughputTrend)
	fmt.Printf("- System reached maximum capacity at %d concurrent requests\n", optimalConcurrency)
}

func runConcurrencyTest(runtime *local_runtime.LocalPluginRuntime, port int, concurrency int, requestCount int) ConcurrencyResult {
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	totalLatency := int64(0)
	successCount := 0
	errorCount := 0

	lock := sync.Mutex{}
	startTime := time.Now()

	for i := 0; i < requestCount; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			start := time.Now()
			runOnce(runtime, requests.RequestInvokeLLM{
				BaseRequestInvokeModel: requests.BaseRequestInvokeModel{
					Provider: "openai",
					Model:    "gpt-3.5-turbo",
				},
				Credentials: requests.Credentials{
					Credentials: map[string]any{
						"openai_api_key":  "test",
						"openai_api_base": fmt.Sprintf("http://localhost:%d", port),
					},
				},
				InvokeLLMSchema: requests.InvokeLLMSchema{
					ModelParameters: map[string]any{
						"temperature": 0.5,
					},
					PromptMessages: []model_entities.PromptMessage{
						{
							Role:    "user",
							Content: "Hello, world!",
						},
					},
					Tools:  []model_entities.PromptMessageTool{},
					Stop:   []string{},
					Stream: true,
				},
				ModelType: model_entities.MODEL_TYPE_LLM,
			})

			latency := time.Since(start).Milliseconds()

			lock.Lock()
			totalLatency += latency
			successCount++
			lock.Unlock()
		}()
	}

	wg.Wait()
	duration := time.Since(startTime).Seconds()

	// Calculate metrics
	totalRequests := successCount + errorCount
	avgLatency := int64(0)
	if totalRequests > 0 {
		avgLatency = totalLatency / int64(totalRequests)
	}

	throughput := float64(successCount) / duration
	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(errorCount) / float64(totalRequests) * 100
	}

	return ConcurrencyResult{
		Concurrency: concurrency,
		Throughput:  throughput,
		AvgLatency:  avgLatency,
		ErrorRate:   errorRate,
	}
}

func analyzeTrend(results []ConcurrencyResult, extractor func(ConcurrencyResult) float64) string {
	if len(results) < 3 {
		return "insufficient data for trend analysis"
	}

	// Simple trend analysis
	increasing := 0
	decreasing := 0

	for i := 1; i < len(results); i++ {
		current := extractor(results[i])
		previous := extractor(results[i-1])

		if current > previous {
			increasing++
		} else if current < previous {
			decreasing++
		}
	}

	if increasing > decreasing*2 {
		return "consistently increasing"
	} else if decreasing > increasing*2 {
		return "consistently decreasing"
	} else if increasing > decreasing {
		return "generally increasing with fluctuations"
	} else if decreasing > increasing {
		return "generally decreasing with fluctuations"
	} else {
		return "fluctuating with no clear trend"
	}
}

func runOnce(
	runtime *local_runtime.LocalPluginRuntime,
	request requests.RequestInvokeLLM,
) {
	session := session_manager.NewSession(
		session_manager.NewSessionPayload{
			UserID:                 "test",
			TenantID:               "test",
			PluginUniqueIdentifier: plugin_entities.PluginUniqueIdentifier(""),
			ClusterID:              "test",
			InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_MODEL,
			Action:                 access_types.PLUGIN_ACCESS_ACTION_INVOKE_LLM,
			Declaration:            nil,
			BackwardsInvocation:    tester.NewMockedDifyInvocation(),
			IgnoreCache:            true,
		},
	)
	session.BindRuntime(runtime)

	response := stream.NewStream[model_entities.LLMResultChunk](1024)

	listener, err := runtime.Listen(session.ID)
	if err != nil {
		log.Panic("Failed to listen: ", err)
	}
	listener.Listen(func(chunk plugin_entities.SessionMessage) {
		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[model_entities.LLMResultChunk](chunk.Data)
			if err != nil {
				response.WriteError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "unmarshal_error",
					"message":    fmt.Sprintf("unmarshal json failed: %s", err.Error()),
				})))
				response.Close()
				return
			} else {
				response.Write(chunk)
			}
		case plugin_entities.SESSION_MESSAGE_TYPE_END:
			response.Close()
		case plugin_entities.SESSION_MESSAGE_TYPE_ERROR:
			e, err := parser.UnmarshalJsonBytes[plugin_entities.ErrorResponse](chunk.Data)
			if err != nil {
				break
			}
			response.WriteError(errors.New(e.Error()))
			response.Close()
		default:
			response.WriteError(errors.New(parser.MarshalJson(map[string]string{
				"error_type": "unknown_stream_message_type",
				"message":    "unknown stream message type: " + string(chunk.Type),
			})))
			response.Close()
		}
	})

	// close the listener if stream outside is closed due to close of connection
	response.OnClose(func() {
		listener.Close()
	})

	session.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
		session.Action,
		generic_invoke.GetInvokePluginMap(
			session,
			request,
		),
	)

	for response.Next() {
		response.Read()
	}
}
