package helpers

import "math/big"

func EncodeBytes(src []byte) string {
	var n big.Int

	n.SetBytes(src)
	return n.Text(62)
}
