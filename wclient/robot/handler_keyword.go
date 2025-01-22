package robot

import (
	"strings"

	"github.com/opentdp/wrest-chat/dbase/keyword"
	"github.com/opentdp/wrest-chat/wcferry"
	"github.com/opentdp/wrest-chat/wclient"
)

func keyworddHandler() []*Handler {

	cmds := []*Handler{}

	keywords, err := keyword.FetchAll(&keyword.FetchAllParam{Group: "keyword"})
	if err != nil {
		return cmds
	}

	keywordprecise, err := keyword.FetchAll(&keyword.FetchAllParam{Group: "keywordprecise"})
	if err != nil {
		return cmds
	}

	for k, v := range keywords {
		v := v // copy
		if v.Remark == "" {
			v.Remark = "神秘指令"
		}

		cmds = append(cmds, &Handler{
			Level:    v.Level,
			Order:    410 + int32(k),
			Roomid:   v.Roomid,
			Command:  v.Phrase,
			Describe: v.Remark,
			PreCheck: func(msg *wcferry.WxMsg) string {
				arr := strings.Split(v.Phrase, "|")
				for _, key := range arr {
					if strings.Contains(msg.Content, key) {
						wclient.SendFlexMsg(v.Target, msg.Sender, msg.Roomid)
						return ""
					}
				}

				return ""
			},
		})
	}
	for k, v := range keywordprecise {
		v := v // copy
		if v.Remark == "" {
			v.Remark = "神秘指令"
		}

		cmds = append(cmds, &Handler{
			Level:    v.Level,
			Order:    410 + int32(k),
			Roomid:   v.Roomid,
			Command:  v.Phrase,
			Describe: v.Remark,
			PreCheck: func(msg *wcferry.WxMsg) string {
				arr := strings.Split(v.Phrase, "|")
				for _, key := range arr {
					if msg.Content == key {
						wclient.SendFlexMsg(v.Target, msg.Sender, msg.Roomid)
						return ""
					}
				}

				return ""
			},
		})
	}

	return cmds

}
