package gemini

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/soulteary/google-gemini-openai-proxy/define"
	"github.com/spf13/viper"
)

func (c *StripPrefixConverter) Name() string {
	return "StripPrefix"
}

func (c *StripPrefixConverter) Convert(req *http.Request, config *DeploymentConfig, payload []byte) (*http.Request, error) {
	req.Host = config.EndpointUrl.Host
	req.URL.Scheme = config.EndpointUrl.Scheme
	req.URL.Host = config.EndpointUrl.Host
	req.URL.Path = fmt.Sprintf("%s/models/%s:generateContent", define.DEFAULT_REST_API_VERSION, config.ModelName)
	req.URL.RawPath = req.URL.EscapedPath()

	query := req.URL.Query()
	query.Add("key", config.ApiKey)
	req.URL.RawQuery = query.Encode()
	req.Body = io.NopCloser(bytes.NewBuffer(payload))
	req.ContentLength = int64(len(payload))
	return req, nil
}

func NewStripPrefixConverter(prefix string) *StripPrefixConverter {
	return &StripPrefixConverter{
		Prefix: prefix,
	}
}

func GetOptionFromEnv(key string) string {
	return strings.TrimSpace(viper.GetString(key))
}

func GetInstance() (err error) {
	var config DeploymentConfig
	endpoint := GetOptionFromEnv(define.ENV_GEMINI_ENDPOINT)
	apikey := GetOptionFromEnv(define.ENV_GEMINI_API_KEY)
	modelName := GetOptionFromEnv(define.ENV_GEMINI_MODEL_NAME)

	if endpoint == "" {
		endpoint = define.DEFAULT_REST_API_ENTRYPOINT
	}

	if modelName == "" {
		modelName = define.DEFAULT_REST_API_MODEL_NAME
	}

	config.ApiKey = apikey
	config.Endpoint = endpoint
	config.ModelName = modelName

	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse endpoint error: %w", err)
	}
	config.EndpointUrl = u

	ModelDeploymentConfig[modelName] = config

	return nil
}
