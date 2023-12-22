package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/soulteary/google-gemini-openai-proxy/define"
	"github.com/soulteary/google-gemini-openai-proxy/util"
)

func ProxyWithConverter(requestConverter RequestConverter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS, POST")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, x-requested-with")
			c.Status(200)
			return
		}
		Proxy(c, requestConverter)
	}
}

var maskURL = regexp.MustCompile(`key=.+`)

// Proxy Azure OpenAI
func Proxy(c *gin.Context, requestConverter RequestConverter) {
	director := func(req *http.Request) {
		if req.Body == nil {
			util.SendError(c, errors.New("request body is empty"))
			return
		}
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		var openaiPayload OpenAIPayload

		if err := json.Unmarshal(body, &openaiPayload); err != nil {
			util.SendError(c, errors.Wrap(err, "parse payload error"))
			return
		}

		// get model from origin request body
		model := strings.TrimSpace(openaiPayload.Model)
		if model == "" {
			model = define.DEFAULT_REST_API_MODEL_NAME
		}

		var payload GoogleGeminiPayload
		for _, data := range openaiPayload.Messages {
			var message GeminiPayloadContents
			if strings.ToLower(data.Role) == "user" {
				message.Role = "user"
			} else {
				message.Role = "model"
			}
			message.Parts = append(message.Parts, GeminiPayloadParts{
				Text: strings.TrimSpace(data.Content),
			})
			payload.Contents = append(payload.Contents, message)
		}

		// set default safety settings
		var safetySettings []GeminiSafetySettings
		safetySettings = append(safetySettings, GeminiSafetySettings{
			Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
			Threshold: "BLOCK_ONLY_HIGH",
		})
		payload.SafetySettings = safetySettings

		// set default generation config
		payload.GenerationConfig.StopSequences = []string{"Title"}
		payload.GenerationConfig.Temperature = openaiPayload.Temperature
		payload.GenerationConfig.MaxOutputTokens = openaiPayload.MaxTokens
		payload.GenerationConfig.TopP = openaiPayload.TopP
		// payload.GenerationConfig.TopK = openaiPayload.TopK

		// get deployment from request
		deployment, err := GetDeploymentByModel(model)
		if err != nil {
			util.SendError(c, err)
			return
		}

		// get auth token from header or deployemnt config
		token := deployment.ApiKey
		if token == "" {
			rawToken := req.Header.Get("Authorization")
			token = strings.TrimPrefix(rawToken, "Bearer ")
		}
		if token == "" {
			util.SendError(c, errors.New("token is empty"))
			return
		}
		req.Header.Set("Authorization", token)

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			util.SendError(c, errors.Wrap(err, "Error converting to JSON"))
			return
		}

		originURL := req.URL.String()
		req, err = requestConverter.Convert(req, deployment, payloadJSON)
		if err != nil {
			util.SendError(c, errors.Wrap(err, "convert request error"))
			return
		}

		log.Printf("proxying request [%s] %s -> %s", model, originURL, maskURL.ReplaceAllString(req.URL.String(), "key=******"))
	}

	proxy := &httputil.ReverseProxy{Director: director}
	transport, err := util.NewProxyFromEnv()
	if err != nil {
		util.SendError(c, errors.Wrap(err, "get proxy error"))
		return
	}
	if transport != nil {
		proxy.Transport = transport
	}

	proxy.ServeHTTP(c.Writer, c.Request)

	// issue: https://github.com/Chanzhaoyu/chatgpt-web/issues/831
	if c.Writer.Header().Get("Content-Type") == "text/event-stream" {
		if _, err := c.Writer.Write([]byte{'\n'}); err != nil {
			log.Printf("rewrite response error: %v", err)
		}
	}

	if c.Writer.Status() != 200 {
		// log.Printf("encountering error with body: %s", string(bodyBytes))
		log.Printf("encountering error with body: %s", string("1"))
	}
}

func GetDeploymentByModel(model string) (*DeploymentConfig, error) {
	deploymentConfig, exist := ModelDeploymentConfig[model]
	if !exist {
		return nil, errors.New(fmt.Sprintf("deployment config for %s not found", model))
	}
	return &deploymentConfig, nil
}
