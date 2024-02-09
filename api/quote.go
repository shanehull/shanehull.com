package api

import (
	"bytes"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func genRandIndex() int {
	return globalRand.Intn(len(quotes))
}

func QuoteHandler(w http.ResponseWriter, r *http.Request) {

	idx := genRandIndex()

	component := quote(quotes[idx].Text, quotes[idx].Author)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := component.Render(r.Context(), buf); err != nil {
		log.Print("failed to render component:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, err := w.Write(buf.Bytes())
	if err != nil {
		log.Print("failed to write response:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
