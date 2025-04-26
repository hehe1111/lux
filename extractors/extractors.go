package extractors

import (
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/utils"
)

var lock sync.RWMutex
var extractorMap = make(map[string]Extractor)

// Register registers an Extractor.
func Register(domain string, e Extractor) {
	lock.Lock()
	extractorMap[domain] = e
	lock.Unlock()
}
// * 注册一个 Extractor，用于将特定的域名与对应的 Extractor 关联起来
// * 每个 Extractor 都会在自己的 init 函数中调用 Register 函数将自己注册到 extractorMap 中
// * 示例： extractors/bilibili/bilibili.go 的 init 函数
// * 在 extractorMap 中，key 是域名，value 是 Extractor 的实例
// * 在 Extract 函数中，会根据 URL 的域名从 extractorMap 中获取对应的 Extractor 实例

// Extract is the main function to extract the data.
func Extract(u string, option Options) ([]*Data, error) {
	u = strings.TrimSpace(u)
	var domain string

	// * 匹配 av、BV、ep 开头的短链接，如： BV... ，而不是 https://www.bilibili.com/video/BV.../
	bilibiliShortLink := utils.MatchOneOf(u, `^(av|BV|ep)\w+`)
	// * 如果匹配到了，则说明是 bilibili 的短链接，需要将其转换为完整的 B 站 URL
	if len(bilibiliShortLink) > 1 {
		bilibiliURL := map[string]string{
			"av": "https://www.bilibili.com/video/",
			"BV": "https://www.bilibili.com/video/",
			"ep": "https://www.bilibili.com/bangumi/play/",
		}
		domain = "bilibili"
		u = bilibiliURL[bilibiliShortLink[1]] + u
	} else {
		// * 把 https://www.bilibili.com/video/BV.../ 解析成 {Scheme: "https", Host: "www.bilibili.com", Path: "/video/BV.../"}
		u, err := url.ParseRequestURI(u)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if u.Host == "haokan.baidu.com" {
			domain = "haokan"
		} else if u.Host == "xhslink.com" {
			domain = "xiaohongshu"
		} else {
			// * 从 www.bilibili.com 提取出 bilibili
			domain = utils.Domain(u.Host)
		}
	}
	extractor := extractorMap[domain]
	if extractor == nil {
		extractor = extractorMap[""]
	}
	videos, err := extractor.Extract(u, option)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, v := range videos {
		v.FillUpStreamsData()
	}
	return videos, nil
}
