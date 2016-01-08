package config

import "github.com/miekg/dns"

var List []string = []string{
	"www.taobao.com",
	"www.baidu.com",
	"www.qq.com",
	"www.meituan.com",
	"www.sina.com.cn",
	"api.weibo.cn",
	"weibo.cn",
	"ww2.sinaimg.cn"}

func IsLocalMysqlBackend(d string) bool {
	for _, x := range List {
		if dns.Fqdn(d) == dns.Fqdn(x) {
			return true
		}
	}
	return false
}
