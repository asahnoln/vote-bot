package tg_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// TODO: Check if anything missing in checking
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
			Result: tg.Message{
				MessageID: 234,
			},
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

	// NOTE: Tested separately
	_, err := d.Send(context.Background(), want)
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

func TestDestinationEditText(t *testing.T) {
	var gotURL string
	tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.Path

		resp := tg.Response{
			OK: true,
		}

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			t.Fatalf("tg srv resp err: %v", err)
		}
	}))
	defer tgSrvStub.Close()

	d := tg.NewDest(tgSrvStub.URL)

	want := tg.EditMessageText{
		ChatID: 42,
	}
	_, err := d.Send(context.Background(), want)
	if err != nil {
		t.Fatalf("send err: got %v; want nil", err)
	}

	if got, want := gotURL, "/editMessageText"; got != want {
		t.Errorf("tg srv request path: got %q; want %q", got, want)
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

	_, err := d.Send(context.Background(), tg.SendRichMessage{})
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

	_, err := d.Send(context.Background(), tg.SendRichMessage{})
	if got, want := err, io.EOF; !errors.Is(got, want) {
		t.Errorf("got err %v; want %q", got, want)
	}
}

// TODO: Change string errors to own wrapped types
func TestDestRequestError(t *testing.T) {
	d := tg.NewDest("")

	_, err := d.Send(context.Background(), tg.SendRichMessage{})
	if got, want := err.Error(), `http client do: Post "/sendRichMessage": unsupported protocol scheme ""`; got != want {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestDestNoopForNil(t *testing.T) {
	tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer tgSrvStub.Close()
	d := tg.NewDest(tgSrvStub.URL)

	_, err := d.Send(context.Background(), nil)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}
}

func TestDestinationReturnsWhateverRequestsNeeds(t *testing.T) {
	wantResult := tgResultStub{
		SomeData: 1,
	}
	tgSrvStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := tg.Response{
			OK:     true,
			Result: wantResult,
		}

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			t.Fatalf("tg srv resp err: %v", err)
		}
	}))
	defer tgSrvStub.Close()

	d := tg.NewDest(tgSrvStub.URL)

	r, err := d.Send(context.Background(), &requestStub{})
	if err != nil {
		t.Fatalf("send err: got %v; want nil", err)
	}

	got, ok := r.(tgResultStub)
	if got, want := ok, true; got != want {
		t.Fatalf("type of result must be tgResultStub: got %v; want %v", got, want)
	}

	if got, want := got.SomeData, wantResult.SomeData; got != want {
		t.Fatalf("data in stub: got %v; want %v", got, want)
	}
}

type tgResultStub struct {
	SomeData int
}

type requestStub struct {
	err error
}

func (r *requestStub) URL() string {
	return "requestStub"
}

func (r *requestStub) UnmarshalResult(data []byte) (any, error) {
	result := tgResultStub{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return result, r.err
}
