package onchain

import (
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
)

func EncodeApproveCalldata(spender string, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("approve(address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(spender).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return data
}

func EncodeTransferCalldata(to string, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("transfer(address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(to).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return data
}

func EncodeOnrampWrapCalldata(asset, to string, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("wrap(address,address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(asset).Bytes())...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(to).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return data
}

func EncodeOfframpUnwrapCalldata(asset, to string, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("unwrap(address,address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(asset).Bytes())...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(to).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return data
}

func EncodeSetApprovalForAllCalldata(operator string, approved bool) []byte {
	selector := ethutil.Keccak256([]byte("setApprovalForAll(address,bool)"))[:4]
	flag := big.NewInt(0)
	if approved {
		flag = big.NewInt(1)
	}
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(operator).Bytes())...)
	data = append(data, ethutil.PadTo32(flag.Bytes())...)
	return data
}

func EncodeSplitPositionCalldata(collateral, conditionID string, partition []int64, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("splitPosition(address,bytes32,bytes32,uint256[],uint256)"))[:4]
	data := make([]byte, 0, 4+5*32+32+len(partition)*32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(collateral).Bytes())...)
	data = append(data, make([]byte, 32)...)
	data = append(data, ParseBytes32Hex(conditionID)...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(5*32)).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(len(partition))).Bytes())...)
	for _, p := range partition {
		data = append(data, ethutil.PadTo32(big.NewInt(p).Bytes())...)
	}
	return data
}

func EncodeMergePositionsCalldata(collateral, conditionID string, partition []int64, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("mergePositions(address,bytes32,bytes32,uint256[],uint256)"))[:4]
	data := make([]byte, 0, 4+5*32+32+len(partition)*32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(collateral).Bytes())...)
	data = append(data, make([]byte, 32)...)
	data = append(data, ParseBytes32Hex(conditionID)...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(5*32)).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(len(partition))).Bytes())...)
	for _, p := range partition {
		data = append(data, ethutil.PadTo32(big.NewInt(p).Bytes())...)
	}
	return data
}

func EncodeRedeemPositionsCalldata(collateral, conditionID string, indexSets []int64) []byte {
	selector := ethutil.Keccak256([]byte("redeemPositions(address,bytes32,bytes32,uint256[])"))[:4]
	data := make([]byte, 0, 4+4*32+32+len(indexSets)*32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(collateral).Bytes())...)
	data = append(data, make([]byte, 32)...)
	data = append(data, ParseBytes32Hex(conditionID)...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(4*32)).Bytes())...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(len(indexSets))).Bytes())...)
	for _, i := range indexSets {
		data = append(data, ethutil.PadTo32(big.NewInt(i).Bytes())...)
	}
	return data
}

func ParseBytes32Hex(s string) []byte {
	stripped := ethutil.StripHexPrefix(s)
	b := common.FromHex("0x" + stripped)
	out := make([]byte, 32)
	if len(b) >= 32 {
		copy(out, b[len(b)-32:])
	} else {
		copy(out[32-len(b):], b)
	}
	return out
}
