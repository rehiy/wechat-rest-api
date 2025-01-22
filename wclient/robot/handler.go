package robot

import (
	"sort"
	"strings"

	"github.com/opentdp/wrest-chat/dbase/keyword"
	"github.com/opentdp/wrest-chat/dbase/setting"
	"github.com/opentdp/wrest-chat/wcferry"
)

var handlers = []*Handler{}
var handlerPre = []*Handler{}
var handlerMap = map[string]*Handler{}

type HandlerFunc func(*wcferry.WxMsg) string

type Handler struct {
	Level    int32       // 0:不限制 7:群管理 9:创始人
	Order    int32       // 排序，越小越靠前
	Roomid   string      // 使用场景 [*:所有,-:私聊,+:群聊,其他:群聊]
	Command  string      // 指令
	Describe string      // 指令的描述信息
	PreCheck HandlerFunc // 前置检查，可拦截文本聊天内容
	Callback HandlerFunc // 指令回调，返回回复内容
}

func GetHandlers() []*Handler {

	return handlers

}

func ResetHandlers() {

	hlst := []*Handler{}
	hpre := []*Handler{}
	hmap := map[string]*Handler{}

	// 获取指令列表
	hlst = append(hlst, aiHandler()...)
	hlst = append(hlst, apiHandler()...)
	hlst = append(hlst, badHandler()...)
	hlst = append(hlst, banHandler()...)
	hlst = append(hlst, cmddHandler()...)
	hlst = append(hlst, keyworddHandler()...)
	hlst = append(hlst, helpHandler()...)
	hlst = append(hlst, revokeHandler()...)
	hlst = append(hlst, roomHandler()...)
	hlst = append(hlst, topHandler()...)
	hlst = append(hlst, webhookHandler()...)

	// 指令列表排序
	sort.Slice(hlst, func(i, j int) bool {
		return hlst[i].Order < hlst[j].Order
	})

	// 获取指令映射
	for _, v := range hlst {
		hmap[v.Command] = v // 重名会覆盖
		if v.PreCheck != nil {
			hpre = append(hpre, v)
		}
	}

	// 获取别名数据
	kws, err := keyword.FetchAll(&keyword.FetchAllParam{Group: "handler"})
	if err == nil && len(kws) > 0 {
		for _, v := range kws {
			if hmap[v.Target] != nil {
				hmap[v.Phrase+"@"+v.Roomid] = hmap[v.Target]
				if v.Roomid == "*" {
					hmap[v.Target].Level = v.Level
				}
			}
		}
	}

	// 更新全局数据
	handlers, handlerPre, handlerMap = hlst, hpre, hmap

}

func ApplyHandlers(msg *wcferry.WxMsg) string {

	// 前置钩子
	for _, v := range handlerPre {
		if txt := v.PreCheck(msg); txt != "" {
			return txt
		}
	}

	// 清理空白
	content := strings.TrimSpace(msg.Content)
	content = strings.ReplaceAll(content, " ", " ")
	if content == "" {
		return ""
	}

	// 解析指令
	params := strings.SplitN(content, " ", 2)
	handler := handlerMap[params[0]] // 默认
	if handler == nil {
		if msg.IsGroup { // 群聊
			handler = handlerMap[params[0]+"@"+msg.Roomid]
			if handler == nil {
				handler = handlerMap[params[0]+"@+"]
			}
		} else { // 私聊
			handler = handlerMap[params[0]+"@-"]
		}
		if handler == nil { // 全局
			handler = handlerMap[params[0]+"@*"]
			if handler == nil {
				if content[0] == '/' {
					return setting.InvalidHandler
				}
				return ""
			}
		}
	}

	// 验证权限
	if groupLimit(msg, handler.Level, handler.Roomid) {
		return setting.InvalidHandler
	}

	// 重写消息
	if len(params) > 1 {
		msg.Content = strings.TrimSpace(params[1])
	} else {
		msg.Content = ""
	}

	// 执行指令
	return handler.Callback(msg)

}
