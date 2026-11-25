package httpapi

import (
	"net/http"
	"strconv"
)

func playerID(r *http.Request) (int64, bool) {
	raw := r.Header.Get("X-Player-ID")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
