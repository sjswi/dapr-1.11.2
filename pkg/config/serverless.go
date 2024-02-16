package config

type Provider string

const (
	AWSLambda Provider = "lambda"
	AliyunFc  Provider = "fc"
)

type AWSConfig struct {
	AccessKeyID    string `mapstructure:"accessKeyID" json:"accessKeyID"`
	AccessSecretID string `mapstructure:"accessSecretID" json:"accessSecretID"`
}

type Response struct {
	Rows  AWSConfig `json:"rows"`
	Error string    `json:"error"`
}

type AliyunConfig struct {
	AccessKeyID    string `mapstructure:"accessKeyID" json:"accessKeyID"`
	AccessSecretID string `mapstructure:"accessSecretID" json:"accessSecretID"`
}
