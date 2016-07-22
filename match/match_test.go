package match

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMatch(t *testing.T) {
	matcher := NewMatcher(false, "")

	srv := httptest.NewServer(matcher)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")

	next := make(chan error)

	const clientCount = 100
	for i := 0; i < clientCount; i++ {
		go client(url, next)
	}

	const messageCount = 10000
	for i := 0; i < messageCount; i++ {
		err := <-next
		if err != nil {
			t.Error(err)
			return
		}
	}
}

type testMessageIn struct {
	messageHeader
	PeerName     string `json:"peerName"`
	Offer        bool   `json:"offer"`
	TurnUsername string `json:"turnUsername,omitempty"`
	TurnPassword string `json:"turnPassword,omitempty"`
	Data         string `json:"data"`
}

type testMessageOut struct {
	messageHeader
	Data string `json:"data"`
}

func client(url string, next chan error) {
	dialer := websocket.Dialer{}
	conn, response, err := dialer.Dial(url, http.Header{})
	if err != nil {
		next <- err
	}
	if response.StatusCode != 101 {
		next <- errors.New(response.Status)
	}

	b := make([]byte, 6)
	c, err := rand.Read(b)
	if err != nil || c != 6 {
		next <- err
	}
	name := base64.StdEncoding.EncodeToString(b)

	err = conn.WriteJSON(struct {
		Name string `json:"name"`
	}{name})
	if err != nil {
		next <- err
	}

	matched := false
	var matchID uint32
	var peerName string
	var offer bool

	for {
		m := testMessageIn{}
		err = conn.ReadJSON(&m)
		if err != nil {
			next <- err
		}

		if m.Type == "start" {
			if matched {
				next <- errors.New("Already matched")
			}
			matched = true
			matchID = m.MatchID
			peerName = m.PeerName
			offer = m.Offer
			if offer {
				err = conn.WriteJSON(testMessageOut{
					messageHeader{
						"offer",
						matchID,
					},
					peerName,
				})
				if err != nil {
					next <- err
				}
			}
		} else {
			if !matched || m.MatchID != matchID {
				next <- errors.New("Not matched or wrong matchID")
			}

			switch m.Type {
			case "end":
				matched = false
				err = conn.WriteJSON(messageHeader{
					"next",
					matchID,
				})
				if err != nil {
					next <- err
				}
			case "offer":
				if m.Data != name {
					next <- errors.New("Names does not match")
				}
				if offer {
					next <- errors.New("Sent and reveived offer")
				}
				err = conn.WriteJSON(testMessageOut{
					messageHeader{
						"answer",
						matchID,
					},
					peerName,
				})
				if err != nil {
					next <- err
				}
			case "answer":
				if !offer {
					next <- errors.New("Sent and reveived answer")
				}
				matched = false
				err = conn.WriteJSON(messageHeader{
					"next",
					matchID,
				})
				if err != nil {
					next <- err
				}
				next <- nil
			}
		}
	}
}
