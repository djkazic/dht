package token

import (
	"crypto/sha1"
	"encoding/binary"
	"net"
	"time"

	"github.com/bradfitz/iter"
)

// Server manages creation and validation of tokens issued to querying nodes.
type Server struct {
	secret           []byte
	interval         time.Duration
	maxIntervalDelta int
	timeNow          func() time.Time
}

// NewServer returns a token server.
func NewServer(secret []byte, interval time.Duration, maxIntervalDelta int, timeNow func() time.Time) *Server {
	if timeNow == nil {
		timeNow = time.Now
	}
	ret := &Server{
		interval:         interval,
		maxIntervalDelta: maxIntervalDelta,
		timeNow:          timeNow,
	}
	copy(ret.secret, secret)
	return ret
}

// NewDefault returns a pre configured token server.
func NewDefault(secret []byte) *Server {
	return NewServer(secret, 5*time.Minute, 2, time.Now)
}

// CreateToken for given addr.
func (s Server) CreateToken(addr *net.UDPAddr) string {
	return s.createToken(addr, s.timeNow())
}

func (s Server) createToken(addr *net.UDPAddr, t time.Time) string {
	h := sha1.New()
	ip := addr.IP.To16()
	if len(ip) != 16 {
		panic(ip)
	}
	h.Write(ip)
	ti := t.UnixNano() / int64(s.interval)
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(ti))
	h.Write(b[:])
	h.Write(s.secret)
	return string(h.Sum(nil))
}

// ValidToken for given address.
func (s *Server) ValidToken(token string, addr *net.UDPAddr) bool {
	t := s.timeNow()
	for range iter.N(s.maxIntervalDelta + 1) {
		if s.createToken(addr, t) == token {
			return true
		}
		t = t.Add(-s.interval)
	}
	return false
}
