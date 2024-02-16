package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/dapr/dapr/pkg/apphealth"
	"github.com/dapr/dapr/pkg/channel"
	"github.com/dapr/dapr/pkg/config"
	diag "github.com/dapr/dapr/pkg/diagnostics"
	invokev1 "github.com/dapr/dapr/pkg/messaging/v1"
	httpMiddleware "github.com/dapr/dapr/pkg/middleware/http"
	auth "github.com/dapr/dapr/pkg/runtime/security"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	StatusCode int32             `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// Channel is an Lambda implementation of an AppChannel.
type Channel struct {
	client   *lambda.Lambda
	funcName string

	ch                     chan struct{}
	tracingSpec            config.TracingSpec
	appHeaderToken         string
	maxResponseBodySizeMB  int
	appHealthCheckPath     string
	appHealth              *apphealth.AppHealth
	pipeline               httpMiddleware.Pipeline
	controlPlatformAddress string
}

// ChannelConfiguration is the configuration used to create an HTTP AppChannel.
type ChannelConfiguration struct {
	Client                 *lambda.Lambda
	FuncName               string
	MaxConcurrency         int
	TracingSpec            config.TracingSpec
	MaxRequestBodySizeMB   int
	ControlPlatformAddress string
}

// CreateLocalChannel creates an HTTP AppChannel.
func CreateLocalChannel(config ChannelConfiguration) (channel.AppChannel, error) {
	c := &Channel{
		client:                 config.Client,
		funcName:               config.FuncName,
		tracingSpec:            config.TracingSpec,
		appHeaderToken:         auth.GetAppToken(),
		maxResponseBodySizeMB:  config.MaxRequestBodySizeMB,
		controlPlatformAddress: config.ControlPlatformAddress,
	}

	if config.MaxConcurrency > 0 {
		c.ch = make(chan struct{}, config.MaxConcurrency)
	}

	return c, nil
}
func (c *Channel) GetAppConfig() (*config.ApplicationConfig, error) {
	return nil, nil
}

// --mode host --config /Users/yu/.dapr/config.yaml --app-protocol lambda --log-level info --app-max-concurrency -1 --components-path /Users/yu/.dapr/components --dapr-http-max-request-size -1 --dapr-http-read-buffer-size -1 --app-id helloworld --app-port 8086 --dapr-http-port 3506 --dapr-grpc-port 64873 --profile-port -1 --metrics-port 64874 --dapr-internal-grpc-port 64875 --app-host 127.0.0.1 --region ningxia --control-platform-address 127.0.0.1:7777 --function-name tt-xbyu-test --provider lambda
func (c *Channel) InvokeMethod(ctx context.Context, req *invokev1.InvokeMethodRequest) (*invokev1.InvokeMethodResponse, error) {

	if c.ch != nil {
		c.ch <- struct{}{}
	}
	defer func() {
		if c.ch != nil {
			<-c.ch
		}
	}()

	// Emit metric when request is sent
	diag.DefaultHTTPMonitoring.ClientRequestStarted(ctx, "GET", req.Message().Method, int64(len(req.Message().Data.GetValue())))
	startRequest := time.Now()
	payload, err := req.RawDataFull()
	if err != nil {
		return nil, fmt.Errorf("read request body failed, err:%s", err)
	}
	// 调用 Lambda 函数
	resp, err := c.client.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(c.funcName), // 替换为你的 Lambda 函数名
		Payload:      payload,
	})
	if err != nil {
		return nil, fmt.Errorf("calling Lambda function failed, err:%s", err)
	}

	elapsedMs := float64(time.Since(startRequest) / time.Millisecond)

	var rsp Response
	if err := json.Unmarshal(resp.Payload, &rsp); err != nil {
		diag.DefaultHTTPMonitoring.ClientRequestCompleted(ctx, "GET", req.Message().GetMethod(), strconv.Itoa(http.StatusInternalServerError), 0, elapsedMs)
		return nil, fmt.Errorf("unmarshal resp failed, err:%s", err)
	}
	contentType := "application/json"
	if _, ok := rsp.Headers["content-type"]; ok {
		contentType = rsp.Headers["content-type"]
	}
	// Limit response body if needed
	var body io.ReadCloser
	//if h.maxResponseBodySizeMB > 0 {
	//	body = streamutils.LimitReadCloser(channelResp., int64(h.maxResponseBodySizeMB)<<20)
	//} else {
	body = io.NopCloser(strings.NewReader(rsp.Body))
	//}
	var contentLength int64
	if _, ok := rsp.Headers["content-length"]; ok {
		contentLength, _ = strconv.ParseInt(rsp.Headers["content-length"], 10, 64)
	} else {
		contentLength = int64(len(rsp.Body))
	}

	// Convert status code
	invokeResp := invokev1.
		NewInvokeMethodResponse(int32(*resp.StatusCode), "", nil).
		WithRawData(body).
		WithContentType(contentType)
	if err != nil {
		diag.DefaultHTTPMonitoring.ClientRequestCompleted(ctx, "GET", req.Message().GetMethod(), strconv.Itoa(http.StatusInternalServerError), contentLength, elapsedMs)
		return nil, err
	}

	//rsp, err := c.parseChannelResponse(req, resp)
	//if err != nil {
	//	diag.DefaultHTTPMonitoring.ClientRequestCompleted(ctx, "GET", req.Message().GetMethod(), strconv.Itoa(http.StatusInternalServerError), contentLength, elapsedMs)
	//	return nil, err
	//}

	diag.DefaultHTTPMonitoring.ClientRequestCompleted(ctx, "GET", req.Message().GetMethod(), strconv.Itoa(int(invokeResp.Status().Code)), contentLength, elapsedMs)

	return invokeResp, nil
}

func (c *Channel) HealthProbe(ctx context.Context) (bool, error) {
	return true, nil
}

func (c *Channel) SetAppHealth(ah *apphealth.AppHealth) {
	c.appHealth = ah
}
