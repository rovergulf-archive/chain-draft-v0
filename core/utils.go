package core

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func onCtxDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("mining cancelled. %s", ctx.Err())
	default:
	}

	return nil
}
