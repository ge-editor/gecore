package remote

import (
	"encoding/json"
	"io"
)

func encode(w io.Writer, v any) error { return json.NewEncoder(w).Encode(v) }

func decode(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) }
