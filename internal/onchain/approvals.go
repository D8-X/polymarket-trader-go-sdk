package onchain

import (
	"context"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

func allowance(ctx context.Context, eth ContractCaller, token, owner, spender common.Address) (*big.Int, error) {
	sel := ethutil.Keccak256([]byte("allowance(address,address)"))[:4]
	data := append([]byte{}, sel...)
	data = append(data, ethutil.PadTo32(owner.Bytes())...)
	data = append(data, ethutil.PadTo32(spender.Bytes())...)
	out, err := eth.CallContract(ctx, ethereum.CallMsg{To: &token, Data: data}, nil)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(out), nil
}

func isApprovedForAll(ctx context.Context, eth ContractCaller, token, owner, operator common.Address) (bool, error) {
	sel := ethutil.Keccak256([]byte("isApprovedForAll(address,address)"))[:4]
	data := append([]byte{}, sel...)
	data = append(data, ethutil.PadTo32(owner.Bytes())...)
	data = append(data, ethutil.PadTo32(operator.Bytes())...)
	out, err := eth.CallContract(ctx, ethereum.CallMsg{To: &token, Data: data}, nil)
	if err != nil {
		return false, err
	}
	return len(out) >= 32 && out[31] == 1, nil
}

func IsFullyApproved(ctx context.Context, eth ContractCaller, depositWallet common.Address) (bool, error) {
	pusd := common.HexToAddress(consts.PUSDAddress)
	ct := common.HexToAddress(consts.ConditionalTokens)
	ctf := common.HexToAddress(consts.CTFExchange)
	negRisk := common.HexToAddress(consts.NegRiskCTFExchange)

	threshold := new(big.Int).Lsh(big.NewInt(1), 128)

	for _, spender := range []common.Address{ctf, negRisk} {
		a, err := allowance(ctx, eth, pusd, depositWallet, spender)
		if err != nil {
			return false, err
		}
		if a.Cmp(threshold) < 0 {
			return false, nil
		}
	}
	for _, operator := range []common.Address{ctf, negRisk} {
		ok, err := isApprovedForAll(ctx, eth, ct, depositWallet, operator)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
