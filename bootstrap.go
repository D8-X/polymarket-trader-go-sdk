package polytrade

import (
	"context"
	"fmt"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type BootstrapResult struct {
	Creds           *L2Credentials
	EOAAddress      string
	SafeAddress     string
	DeployResponse  *RelayerResponse
	ApproveResponse *RelayerResponse
}

func Bootstrap(ctx context.Context, privateKeyHex string, relayerCreds *RelayerCredentials) (*BootstrapResult, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("bootstrap: invalid private key: %w", err)
	}
	eoaAddress := crypto.PubkeyToAddress(pk.PublicKey).Hex()

	creds, err := DeriveL2Credentials(privateKeyHex, PolygonChainID)
	if err != nil {
		creds, err = CreateL2Credentials(privateKeyHex, PolygonChainID)
		if err != nil {
			return nil, fmt.Errorf("bootstrap: get/create L2 credentials: %w", err)
		}
	}

	safeAddr, deployResp, err := EnsureSafeAddressSync(ctx, eoaAddress, privateKeyHex, relayerCreds)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: ensure safe: %w", err)
	}

	approveResp, err := ApproveSafeTokens(ctx, eoaAddress, privateKeyHex, CTFExchange, relayerCreds)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: approve tokens: %w", err)
	}

	return &BootstrapResult{
		Creds:           creds,
		EOAAddress:      eoaAddress,
		SafeAddress:     safeAddr,
		DeployResponse:  deployResp,
		ApproveResponse: approveResp,
	}, nil
}
