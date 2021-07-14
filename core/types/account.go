package types

import (
	"bytes"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/common"
)

//type Account struct {
//	Address common.Address `json:"address" yaml:"address"`
//	Auth    string         `json:"auth" yaml:"auth"`
//	KeyData []byte         `json:"-" yaml:"-"` // stores encrypted key
//	key     *keystore.Key
//}

// Balance represents registered account balance response
type Balance struct {
	Address common.Address `json:"address" yaml:"address"`
	Balance uint64         `json:"balance" yaml:"balance"`
	Nonce   uint64         `json:"nonce" yaml:"nonce"`
}

func (b *Balance) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(*b); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (b *Balance) Deserialize(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(b)
}
