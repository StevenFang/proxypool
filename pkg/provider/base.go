package provider

import (
	"fmt"
	"github.com/Sansui233/proxypool/pkg/healthcheck"
	"strings"

	"github.com/Sansui233/proxypool/pkg/proxy"
)

type Provider interface {
	Provide() string
}

type Base struct {
	Proxies    *proxy.ProxyList `yaml:"proxies"`
	Types      string           `yaml:"type"`
	Country    string           `yaml:"country"`
	NotCountry string           `yaml:"not_country"`
}

// 根据子类的的Provide()传入的信息筛选节点，结果会改变传入的proxylist。
func (b *Base) preFilter() {
	proxies := make(proxy.ProxyList, 0)

	needFilterType := true
	needFilterCountry := true
	needFilterNotCountry := true
	if b.Types == "" || b.Types == "all" {
		needFilterType = false
	}
	if b.Country == "" || b.Country == "all" {
		needFilterCountry = false
	}
	if b.NotCountry == "" {
		needFilterNotCountry = false
	}
	types := strings.Split(b.Types, ",")
	countries := strings.Split(b.Country, ",")
	notCountries := strings.Split(b.NotCountry, ",")

	bProxies := *b.Proxies
	for _, p := range bProxies {
		if needFilterType {
			typeOk := false
			for _, t := range types {
				if p.TypeName() == t {
					typeOk = true
					break
				}
			}
			if !typeOk {
				goto exclude
			}
		}

		if needFilterNotCountry {
			for _, c := range notCountries {
				if strings.Contains(p.BaseInfo().Name, c) {
					goto exclude
				}
			}
		}

		if needFilterCountry {
			countryOk := false
			for _, c := range countries {
				if strings.Contains(p.BaseInfo().Name, c) {
					countryOk = true
					break
				}
			}
			if !countryOk {
				goto exclude
			}
		}

		proxies = append(proxies, p)
	exclude:
	}

	b.Proxies = &proxies
}

func checkSpeedResult(p proxy.Proxy) proxy.Proxy {
	if healthcheck.SpeedResults == nil {
		return p
	}
	if speed, ok := healthcheck.SpeedResults[p.Identifier()]; ok {
		pp := p.Clone()
		pp.AddToName(fmt.Sprintf("_%5.2fMb", speed))
		return pp
	}
	return p
}
