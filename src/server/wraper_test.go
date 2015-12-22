package server

import (
	"testing"
	"time"
)

func TestGetSOARecord(t *testing.T) {
	soa, e := GetSOARecord("www.a.shifen.com")
	t.Log(soa)
	t.Log(e)
	time.Sleep(10 * time.Second)
	soa, e = GetSOARecord("www.a.shifen.com")
	t.Log(soa)
	t.Log(e)
}
