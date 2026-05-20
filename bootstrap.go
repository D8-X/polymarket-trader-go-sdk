package polytrade

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type DepositWalletBootstrapResult struct {
	Creds                *L2Credentials
	EOAAddress           string
	DepositWalletAddress string
	DeployResponse       *RelayerResponse
	BatchResponse        *RelayerResponse
}

func BootstrapDepositWallet(ctx context.Context, privateKeyHex, depositWalletAddress string, wrapAmount *big.Int, relayerCreds *RelayerCredentials) (*DepositWalletBootstrapResult, error) {
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

	deployResp, err := DeployDepositWallet(ctx, eoaAddress, relayerCreds)
	if err != nil {
		return nil, fmt.Errorf("bootstrap deposit wallet: deploy: %w", err)
	}
	if deployResp != nil && deployResp.TransactionID != "" {
		_, _ = WaitForRelayerTransaction(ctx, deployResp.TransactionID)
		time.Sleep(2 * time.Second)
	}

	var batchResp *RelayerResponse
	if wrapAmount != nil {
		r, err := WrapAndApproveDepositWallet(ctx, eoaAddress, privateKeyHex, depositWalletAddress, wrapAmount, relayerCreds)
		if err != nil {
			return nil, fmt.Errorf("bootstrap deposit wallet: wrap+approve: %w", err)
		}
		batchResp = r
	}

	return &DepositWalletBootstrapResult{
		Creds:                creds,
		EOAAddress:           eoaAddress,
		DepositWalletAddress: depositWalletAddress,
		DeployResponse:       deployResp,
		BatchResponse:        batchResp,
	}, nil
}
