package accounts

import (
	"github.com/mr-tron/base58"
)

// DigestLength is for signature digest length
const DigestLength = 32

func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}

func Base58Decode(input []byte) ([]byte, error) {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		return nil, err
	}

	return decode, nil
}
