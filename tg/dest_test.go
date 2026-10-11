package tg_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestDestination(t *testing.T) {
	got := tg.SendRichMessage{}
	tgSrvStubCalled := false
	tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tgSrvStubCalled = true

		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("tg srv request method: got %q; want %q", got, want)
		}

		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("tg srv request Content-Type: got %q; want %q", got, want)
		}

		if got, want := r.URL.Path, "/sendRichMessage"; got != want {
			t.Errorf("tg srv request path: got %q; want %q", got, want)
		}

		err := json.NewDecoder(r.Body).Decode(&got)
		if err != nil {
			t.Errorf("tg srv request unmarshal error: %v", err)
		}
		defer r.Body.Close()

		resp := tg.Response{
			OK: true,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			t.Fatalf("tg srv resp err: %v", err)
		}
	}))
	defer tgSrvStub.Close()

	d := tg.NewDest(tgSrvStub.URL)

	want := tg.SendRichMessage{
		ChatID: 52,
	}
	err := d.Send(context.Background(), want)
	if err != nil {
		t.Fatalf("send err: got %v; want nil", err)
	}

	if got, want := tgSrvStubCalled, true; got != want {
		t.Fatalf("tg srv stub called: got %v; want %v", got, want)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}
}

func TestDestTgSrvNotOK(t *testing.T) {
	tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := tg.Response{
			OK:          false,
			Description: "something is not ok",
		}

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			t.Fatalf("tg srv resp err: %v", err)
		}
	}))
	defer tgSrvStub.Close()

	d := tg.NewDest(tgSrvStub.URL)

	err := d.Send(context.Background(), nil)
	if got, want := err.Error(), "tg srv not ok: something is not ok"; got != want {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestDestTgSrvWrongJSON(t *testing.T) {
	tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(``))
	}))
	defer tgSrvStub.Close()

	d := tg.NewDest(tgSrvStub.URL)

	err := d.Send(context.Background(), nil)
	if got, want := err, io.EOF; !errors.Is(got, want) {
		t.Errorf("got err %v; want %q", got, want)
	}
}
