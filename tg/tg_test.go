package tg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
)

func TestDefaultNotFound(t *testing.T) {
	r := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	b := tg.Bot{}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("bot response code: got %v; want %v", got, want)
	}
}
