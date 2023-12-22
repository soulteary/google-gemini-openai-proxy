package gemini

import (
	"net/http"
	"net/url"
)

type DeploymentConfig struct {
	ModelName   string   `yaml:"model_name" json:"model_name" mapstructure:"model_name"` // corresponding model name in openai
	Endpoint    string   `yaml:"endpoint" json:"endpoint" mapstructure:"endpoint"`       // deployment endpoint
	ApiKey      string   `yaml:"api_key" json:"api_key" mapstructure:"api_key"`          // secrect key1 or 2
	EndpointUrl *url.URL // url.URL form deployment endpoint
}

type RequestConverter interface {
	Name() string
	Convert(req *http.Request, config *DeploymentConfig, payload []byte) (*http.Request, error)
}

type StripPrefixConverter struct {
	Prefix string
}

type OpenAIPayloadMessages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIPayload struct {
	MaxTokens       int                     `json:"max_tokens"`
	Model           string                  `json:"model"`
	Temperature     float64                 `json:"temperature"`
	TopP            float64                 `json:"top_p"`
	PresencePenalty float64                 `json:"presence_penalty"`
	Messages        []OpenAIPayloadMessages `json:"messages"`
	Stream          bool                    `json:"stream"`
}

type GoogleGeminiPayload struct {
	Contents         []GeminiPayloadContents  `json:"contents"`
	SafetySettings   []GeminiSafetySettings `json:"safetySettings"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig"`
}

type GeminiSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiGenerationConfig struct {
	StopSequences   []string `json:"stopSequences"`
	Temperature     float64  `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
}

type GeminiPayloadContents struct {
	Parts []GeminiPayloadParts `json:"parts"`
	Role  string               `json:"role"`
}

type GeminiPayloadParts struct {
	Text string `json:"text"`
}
