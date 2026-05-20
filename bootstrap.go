package polytrade

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type depositWalletBootstrapResult struct {
	Creds                *L2Credentials
	EOAAddress           string
	DepositWalletAddress string
	DeployResponse       *RelayerResponse
	BatchResponse        *RelayerResponse
}

func bootstrapDepositWallet(ctx context.Context, eth ReceiptFetcher, privateKeyHex string, wrapAmount *big.Int, relayerCreds *RelayerCredentials) (*depositWalletBootstrapResult, error) {
	if eth == nil {
		return nil, fmt.Errorf("bootstrap deposit wallet: nil ReceiptFetcher")
	}
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("bootstrap deposit wallet: invalid private key: %w", err)
	}
	eoaAddress := crypto.PubkeyToAddress(pk.PublicKey).Hex()

	creds, err := DeriveL2Credentials(privateKeyHex, PolygonChainID)
	if err != nil {
		creds, err = CreateL2Credentials(privateKeyHex, PolygonChainID)
		if err != nil {
			return nil, fmt.Errorf("bootstrap deposit wallet: get/create L2 credentials: %w", err)
		}
	}

	depositWalletAddress, deployResp, _, err := deployAndResolveDepositWallet(ctx, eth, eoaAddress, relayerCreds)
	if err != nil {
		return nil, fmt.Errorf("bootstrap deposit wallet: deploy: %w", err)
	}
	time.Sleep(2 * time.Second)

	var batchResp *RelayerResponse
	if wrapAmount != nil {
		r, err := wrapAndApproveDepositWallet(ctx, eoaAddress, privateKeyHex, depositWalletAddress, wrapAmount, relayerCreds)
		if err != nil {
			return nil, fmt.Errorf("bootstrap deposit wallet: wrap+approve: %w", err)
		}
		batchResp = r
	}

	return &depositWalletBootstrapResult{
		Creds:                creds,
		EOAAddress:           eoaAddress,
		DepositWalletAddress: depositWalletAddress,
		DeployResponse:       deployResp,
		BatchResponse:        batchResp,
	}, nil
}
