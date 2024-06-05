package server

import (
	"net/http"

	"github.com/alioygur/gores"
)

// SimpleMessage contains a simple message for return.
type SimpleMessage struct {
	Message string `json:"Message"`
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	_ = gores.JSON(w, http.StatusOK, SimpleMessage{Message: "All is well."})
}
