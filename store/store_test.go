package store_test

import (
	"testing"

	"github.com/asahnoln/vote-bot/store"
)

func TestModeString(t *testing.T) {
	tests := []struct {
		mode store.Mode
		want string
	}{
		{store.OptionalMode, "опциональный"},
		{store.RealMode, "реальный"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			got := tt.mode.String()
			if got, want := got, tt.want; got != want {
				t.Errorf("got %v; want %v", got, want)
			}
		})
	}
}
