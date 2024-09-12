package core

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/tidwall/pretty"
)

func PrintPeerAddr(pa peer.AddrInfo) {
	paddr, _ := pa.MarshalJSON()
	PrintBytes(paddr)
}

func PrintBytes(out []byte) {
	fmt.Print(string(pretty.Color(pretty.Pretty(out), nil)))
	pp.Printf("\n")
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
const base = 36

func EncodeToSafeString(data []byte) string {
	num := new(big.Int).SetBytes(data)
	encoded := ""
	zero := big.NewInt(0)
	base := big.NewInt(base)

	for num.Cmp(zero) > 0 {
		mod := new(big.Int)
		num.DivMod(num, base, mod)
		encoded = string(charset[mod.Int64()]) + encoded
	}

	if encoded == "" {
		return string(charset[0])
	}
	return encoded
}

func DecodeToBytes(s string) ([]byte, error) {
	num := new(big.Int)
	base := big.NewInt(base)

	for _, c := range s {
		index := int64(-1)
		for i, char := range charset {
			if c == char {
				index = int64(i)
				break
			}
		}
		if index == -1 {
			return nil, errors.New("invalid character in input")
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(index))
	}

	return num.Bytes(), nil
}
