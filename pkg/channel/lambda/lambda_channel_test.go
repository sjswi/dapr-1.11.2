package lambda

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/dapr/dapr/pkg/config"
	"io"
	nethttp "net/http"
	"testing"
)

func TestLambdaInvoke(t *testing.T) {
	resp, err := nethttp.Get(fmt.Sprintf("http://%s/tt/api/v1/secret/getSecret?provider=aws", "127.0.0.1:7777"))
	if err != nil {
		panic(err)
	}
	var awsConfig config.Response
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(all, &awsConfig)
	if err != nil {
		panic(err)
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("cn-northwest-1"), // 可以根据你的区域进行调整
		Credentials: credentials.NewStaticCredentials(awsConfig.Rows.AccessKeyID, awsConfig.Rows.AccessSecretID, ""),
	})
	if err != nil {
		panic(err)
	}
	lambdaClient := lambda.New(sess)
	resp1, err := lambdaClient.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String("tt-xbyu-test"), // 替换为你的 Lambda 函数名
		Payload:      []byte(`{"name": "value1"}`),
	})
	if err != nil {
		panic(err)
	}
	var rsp Response
	if err := json.Unmarshal(resp1.Payload, &rsp); err != nil {
		panic(err) // 或使用更优雅的错误处理方式
	}

	fmt.Println(string(resp1.Payload))
	fmt.Printf("%v\n", rsp)
}
