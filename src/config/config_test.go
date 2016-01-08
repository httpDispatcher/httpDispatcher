package config

import "testing"

func TestGet(t *testing.T) {
	x := []string{"api.weibo.cn", "weibo.cn", "ww2.sinaimg.cn", "ww3.sinaimg.cn"}
	//	x := []string{"weibo.cn"}
	for _, xx := range x {
		ok := IsLocalMysqlBackend(xx)
		if ok {
			t.Log("Got : ", xx)
		} else {
			t.Log("Not got : ", xx, ok)
		}
	}
}
