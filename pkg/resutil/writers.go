package resutil

import (
	"bytes"
	"encoding/json"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io"
)

func WriteJSON(w io.Writer, lg *zap.SugaredLogger, v interface{}) error {
	payload, err := json.Marshal(v)
	if err != nil {
		lg.Errorf("unable to marshal payload: %s", err)
		return err
	}

	out := bytes.NewBuffer(nil)
	if err := json.Indent(out, payload, "", "	"); err != nil {
		lg.Errorf("Unable to indent json payload: %s", err)
		return err
	}

	return write(w, out.Bytes())
}

func WriteYAML(w io.Writer, lg *zap.SugaredLogger, v interface{}) error {
	payload, err := yaml.Marshal(v)
	if err != nil {
		lg.Errorf("unable to marshal payload: %s", err)
		return err
	}

	return write(w, payload)
}

func write(w io.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}
