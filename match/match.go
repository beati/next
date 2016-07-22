package match

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

type Matcher struct {
	log           bool
	currentUserID int64
	poolLock      sync.Mutex
	userPool      map[*user]struct{}
	turnSecret    []byte
}

func NewMatcher(log bool, turnSecret string) *Matcher {
	matcher := &Matcher{}
	matcher.log = log
	matcher.userPool = make(map[*user]struct{})
	matcher.turnSecret = []byte(turnSecret)
	return matcher
}

var upgrader = websocket.Upgrader{}

func (matcher *Matcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := atomic.AddInt64(&matcher.currentUserID, 1)
	logOutput := ioutil.Discard
	if matcher.log {
		logOutput = os.Stderr
	}
	logger := log.New(logOutput, "userID: "+strconv.FormatInt(userID, 10)+" ", log.Lshortfile)

	logger.Printf("new user connected")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Printf("%v %v", "websocket upgrade error:", err)
		return
	}
	defer conn.Close()

	userInfos := struct {
		Name string `json:"name"`
	}{}
	err = conn.ReadJSON(&userInfos)
	if err != nil {
		logger.Printf("%v %v", "user infos retrieval error:", err)
		return
	}
	if userInfos.Name == "" {
		logger.Printf("%v", "empty user name")
		return
	}
	if utf8.RuneCountInString(userInfos.Name) > 15 {
		logger.Printf("%v", "user name too long")
		return
	}

	logger.Printf("user infos retrieved, name: %v", userInfos.Name)

	self := &user{
		name:   userInfos.Name,
		conn:   conn,
		userID: userID,
		logger: logger,
	}

	matcher.put(self)

	for {
		m := message{}
		err := conn.ReadJSON(&m)
		if err != nil {
			logger.Printf("receive error, disconnecting: %v", err)
			self.handleDisconnection(matcher)
			return
		}

		switch m.Type {
		case "start":
			fallthrough
		case "end":
			logger.Printf("%v", "start or end sent by client")
			return
		case "next":
			logger.Printf("next received, matchID: %v", m.MatchID)
			self.handleNext(matcher, m.MatchID)
		default:
			self.stateLock.Lock()
			if self.remotePeer != nil && m.MatchID == self.match.matchID {
				_ = self.remotePeer.send(&m)
			}
			self.stateLock.Unlock()
		}
	}
}

func (matcher *Matcher) put(u *user) {
SeekPeer:
	u.logger.Printf("seeking a peer")
	var peer *user
	matcher.poolLock.Lock()
	for peer = range matcher.userPool {
		delete(matcher.userPool, peer)
		break
	}
	matcher.poolLock.Unlock()
	if peer != nil {
		u.logger.Printf("peer found, matching with %v", peer.userID)

		var matchID uint32
		err := binary.Read(rand.Reader, binary.BigEndian, &matchID)
		if err != nil {
			log.Fatal(err)
		}
		match := match{matchID: matchID}

		match.Lock()
		peer.stateLock.Lock()
		if peer.disconnected {
			u.logger.Printf("peer %v was disconnected", peer.userID)
			peer.stateLock.Unlock()
			goto SeekPeer
		}
		peer.match = &match
		peer.remotePeer = u
		peer.stateLock.Unlock()

		u.stateLock.Lock()
		u.match = &match
		u.remotePeer = peer
		u.stateLock.Unlock()

		turnUsername, turnPassword := matcher.newTurnCreds()
		_ = u.sendStart(peer.name, matchID, true, turnUsername, turnPassword)
		turnUsername, turnPassword = matcher.newTurnCreds()
		_ = peer.sendStart(u.name, matchID, false, turnUsername, turnPassword)
		match.Unlock()
	} else {
		u.logger.Printf("no peer found, put in wait queue")
		matcher.poolLock.Lock()
		matcher.userPool[u] = struct{}{}
		matcher.poolLock.Unlock()
	}
}

func (matcher *Matcher) delete(u *user) {
	matcher.poolLock.Lock()
	delete(matcher.userPool, u)
	matcher.poolLock.Unlock()
	u.logger.Printf("deleted from pool")
}

type user struct {
	name         string
	connLock     sync.Mutex
	conn         *websocket.Conn
	stateLock    sync.Mutex
	match        *match
	remotePeer   *user
	disconnected bool
	userID       int64
	logger       *log.Logger
}

func (u *user) send(m interface{}) error {
	defer u.connLock.Unlock()
	u.connLock.Lock()
	return u.conn.WriteJSON(m)
}

func (u *user) sendStart(peerName string, matchID uint32, offer bool, turnUsername, turnPassword string) error {
	return u.send(struct {
		messageHeader
		PeerName     string `json:"peerName"`
		Offer        bool   `json:"offer"`
		TurnUsername string `json:"turnUsername,omitempty"`
		TurnPassword string `json:"turnPassword,omitempty"`
	}{
		messageHeader{
			"start",
			matchID,
		},
		peerName,
		offer,
		turnUsername,
		turnPassword,
	})
}

func (u *user) sendEnd(matchID uint32) error {
	return u.send(messageHeader{
		"end",
		matchID,
	})
}

func (u *user) handleDisconnection(matcher *Matcher) {
	u.stateLock.Lock()
	u.disconnected = true
	u.stateLock.Unlock()

	match := u.getMatch()
	if match != nil {
		match.Lock()
		u.unMatch(match.matchID)
		match.Unlock()
	} else {
		matcher.delete(u)
	}
}

func (u *user) handleNext(matcher *Matcher, matchID uint32) {
	match := u.getMatch()
	if match != nil && matchID == match.matchID {
		match.Lock()
		u.unMatch(match.matchID)
		match.Unlock()

		matcher.put(u)
	}
}

func (u *user) getMatch() *match {
	defer u.stateLock.Unlock()
	u.stateLock.Lock()
	return u.match
}

func (u *user) unMatch(matchID uint32) {
	u.stateLock.Lock()
	remotePeer := u.remotePeer
	u.match = nil
	u.remotePeer = nil
	u.stateLock.Unlock()

	if remotePeer != nil {
		remotePeer.stateLock.Lock()
		remotePeer.remotePeer = nil
		_ = remotePeer.sendEnd(matchID)
		remotePeer.stateLock.Unlock()
	}
}

type match struct {
	sync.Mutex
	matchID uint32
}

type messageHeader struct {
	Type    string `json:"type"`
	MatchID uint32 `json:"matchID"`
}

type message struct {
	messageHeader
	Data json.RawMessage `json:"data"`
}

func (matcher *Matcher) newTurnCreds() (string, string) {
	username := ""
	password := ""

	if len(matcher.turnSecret) > 0 {
		b := make([]byte, 6)
		c, err := rand.Read(b)
		if err != nil || c != 6 {
			log.Fatal(err)
		}
		name := base64.StdEncoding.EncodeToString(b)

		timestamp := time.Now().Unix() + 10
		username = fmt.Sprintf("%d:%s", timestamp, name)
		mac := hmac.New(sha1.New, matcher.turnSecret)
		_, _ = mac.Write([]byte(username))
		password = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}

	return username, password
}
