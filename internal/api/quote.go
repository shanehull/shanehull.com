package api

import (
	"bytes"
	"log"
	"math/rand/v2"
	"net/http"
)

func QuoteHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idx := rand.N(len(quotes))

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
