package polytrade

import (
	"net/http"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/auth"
)

func DeriveL2Credentials(privateKeyHex string, chainID int) (*L2Credentials, error) {
	return auth.DeriveCredentials(privateKeyHex, chainID)
}

func CreateL2Credentials(privateKeyHex string, chainID int) (*L2Credentials, error) {
	return auth.CreateCredentials(privateKeyHex, chainID)
}

func SignL2Request(creds *L2Credentials, method, path string, body []byte) (*L2Headers, error) {
	return auth.SignRequest(creds, method, path, body)
}

func signL2Message(secret, ts, method, path string, body []byte) (string, error) {
	return auth.SignMessage(secret, ts, method, path, body)
}

func ApplyL2Headers(req *http.Request, h *L2Headers) {
	auth.ApplyHeaders(req, h)
}
