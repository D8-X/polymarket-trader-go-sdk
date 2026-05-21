package polytrade

import (
	"context"
	"net/http"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/relayer"
)

func signRelayerHMAC(secret string, timestamp int64, method, path string, body []byte) string {
	return relayer.SignHMAC(secret, timestamp, method, path, body)
}

func applyRelayerHeaders(req *http.Request, creds *RelayerCredentials, method, path string, body []byte) {
	relayer.ApplyHeaders(req, creds, method, path, body)
}

func getRelayerTransaction(ctx context.Context, transactionID string) (*RelayerTransaction, error) {
	return relayer.GetTransaction(ctx, transactionID)
}

func waitForRelayerTransaction(ctx context.Context, transactionID string) (*RelayerTransaction, error) {
	return relayer.WaitForTransaction(ctx, transactionID)
}
