package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestServerSideEvent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/.servus", nil)
	res := httptest.NewRecorder()

	events := make(chan fsnotify.Event, 1)
	handler := serverSideEvent(log.Default(), &fsnotify.Watcher{Events: events})

	go func() {
		time.Sleep(250 * time.Millisecond)
		events <- fsnotify.Event{
			Op:   fsnotify.Write,
			Name: "test.html",
		}
	}()

	handler(res, req)

	if actual := res.Result().Status; actual != "200 OK" {
		t.Logf("Wrong status; expected='200 OK', actual='%s'", actual)
		t.Fail()
	}

	for key, expected := range map[string]string{
		"Content-Type":  "text/event-stream",
		"Connection":    "keep-alive",
		"Cache-Control": "no-cache",
	} {
		if actual := res.Result().Header[key][0]; actual != expected {
			t.Logf("Wrong header '%s'; expected='%s',actual='%s'", key, expected, actual)
			t.Fail()
		}
	}

	body, err := io.ReadAll(res.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	match, err := regexp.Match("data: servus pid=\\d+ WRITE         \"test\\.html\"", body)
	if err != nil {
		t.Fatal(err)
	}
	if match == false {
		t.Logf("Unexpected body: %s", body)
		t.Fail()
	}
}

func TestServeFile(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test.html", nil)
	req.SetPathValue("file", "test.html")
	res := httptest.NewRecorder()

	handler := serveFile(log.Default())

	handler(res, req)

	if actual := res.Result().Status; actual != "200 OK" {
		t.Logf("Wrong status; expected='200 OK', actual='%s'", actual)
		t.Fail()
	}
}
