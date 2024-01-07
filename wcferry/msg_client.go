package wcferry

import (
	"errors"
	"strings"

	"github.com/clbanning/mxj"
	"github.com/opentdp/go-helper/logman"
)

type MsgClient struct {
	*pbSocket               // RPC 客户端
	receiving bool          // 接收消息中
	callbacks []MsgCallback // 消息回调函数
}

type MsgPayload struct {
	*WxMsg                    // 消息原始数据
	XmlMap     map[string]any `json:",omitempty"`
	ContentMap map[string]any `json:",omitempty"`
}

// 消息回调函数
type MsgCallback func(msg *MsgPayload)

// 关闭 RPC 连接
// return error 错误信息
func (c *MsgClient) Destroy(force bool) error {
	if !force && len(c.callbacks) > 0 {
		return errors.New("callbacks not empty")
	}
	c.callbacks = []MsgCallback{}
	c.receiving = false
	return c.close()
}

// 创建消息接收器
// param fn ...MsgCallback 消息回调函数
func (c *MsgClient) Register(fn ...MsgCallback) error {
	if !c.receiving {
		// 连接消息服务
		if err := c.init(0); err != nil {
			logman.Error("msg receiver", "error", err)
			return err
		}
		// 开始接收消息
		c.receiving = true
		go c.listener()
	}
	c.callbacks = append(c.callbacks, fn...)
	return nil
}

// 消息接收器循环
func (c *MsgClient) listener() {
	defer c.Destroy(true)
	for c.receiving {
		if resp, err := c.recv(); err == nil {
			res := &MsgPayload{resp.GetWxmsg(), nil, nil}
			// 解析 XML 内容
			if res.ContentMap, err = convertMsgContent(res.Content); err == nil {
				res.Content = ""
			}
			if res.XmlMap, err = mxj.NewMapXml([]byte(res.Xml)); err == nil {
				res.Xml = ""
			}
			// 批量回调
			for _, f := range c.callbacks {
				go f(res)
			}
		} else {
			logman.Error("msg receiver", "error", err)
		}
	}
	logman.Warn("msg receiver stopped")
}

func convertMsgContent(str string) (mxj.Map, error) {
	str = strings.TrimSpace(str)
	xmlPrefixes := []string{"<?xml", "<sysmsg", "<msg"}
	for _, prefix := range xmlPrefixes {
		if strings.HasPrefix(str, prefix) {
			return mxj.NewMapXml([]byte(str))
		}
	}
	return nil, errors.New("skip")
}
