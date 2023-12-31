package define

const (
	ENV_GEMINI_ENDPOINT   = "GEMINI_ENDPOINT"
	ENV_GEMINI_API_KEY    = "GEMINI_API_KEY"
	ENV_GEMINI_MODEL_NAME = "GEMINI_MODEL"

	ENV_GEMINI_HTTP_PROXY  = "GEMINI_HTTP_PROXY"
	ENV_GEMINI_SOCKS_PROXY = "GEMINI_SOCKS_PROXY"
)

const (
	DEFAULT_REST_API_VERSION_SHIM = "/v1"
	DEFAULT_REST_API_VERSION      = "/v1beta"
	DEFAULT_REST_API_ENTRYPOINT   = "https://generativelanguage.googleapis.com"
	DEFAULT_REST_API_MODEL_NAME   = "gemini-pro"
)
