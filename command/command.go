package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/eryajf/chatgpt-dingtalk/pkg/cache"
	"github.com/eryajf/chatgpt-dingtalk/pkg/dingbot"
	"github.com/eryajf/chatgpt-dingtalk/pkg/logger"
	"github.com/eryajf/chatgpt-dingtalk/pkg/process"
	"github.com/eryajf/chatgpt-dingtalk/public"
)

const (
	CommandChat  = "#chat"
	CommandChats = "#chats"
	CommandReset = "#reset"
	CommandHelp  = "#help"
	CommandTmpl  = "#tmpl"
	CommandImage = "#image"
)

type Command struct {
	Desc string
	Cmd  string
}

func Commands() map[string]Command {
	return map[string]Command{
		CommandChat:  {Desc: "单独聊天", Cmd: CommandChat},
		CommandChats: {Desc: "带上下文聊天", Cmd: CommandChats},
		CommandReset: {Desc: "重置带上下文聊天", Cmd: CommandReset},
		CommandHelp:  {Desc: "显示帮助信息", Cmd: CommandHelp},
		CommandTmpl:  {Desc: "内置的prompt", Cmd: CommandTmpl},
		CommandImage: {Desc: "生成图片", Cmd: CommandImage},
	}
}

func Welcome() string {
	var welcomes []string = []string{"Commands:", "================================="}
	for k, v := range Commands() {
		welcomes = append(welcomes, fmt.Sprintf("%s -> %s", k, v.Desc))
	}
	welcomes = append(welcomes, "=================================")
	welcomes = append(welcomes, "🚜 例：@我发送 空 或 #help 将返回此帮助信息")
	return strings.Join(welcomes, "\n")
}

// 收到消息
func ReceiveMsg(msgObj *dingbot.ReceiveMsg) error {
	// 打印钉钉回调过来的请求明细
	logger.Info(fmt.Sprintf("dingtalk callback parameters: %#v", msgObj))

	content := strings.TrimSpace(msgObj.Text.Content)

	switch content {
	case CommandChat:
		return ReceiveChat(msgObj)
	case CommandChats:
		return ReceiveChats(msgObj)
	case CommandReset:
		return ReceiveReset(msgObj)
	case "":
		fallthrough
	case CommandHelp:
		return ReceiveWelcome(msgObj)
	case CommandTmpl:
		return ReceiveTmpl(msgObj)
	default:
	}

	if !public.CheckRequest(msgObj) {
		return errors.New("超出使用限制，请联系管理员")
	}

	if strings.HasPrefix(content, CommandImage) {
		return ReceiveImage(msgObj)
	}

	var err error
	msgObj.Text.Content, err = process.GeneratePrompt(content)
	if err != nil {
		return err
	}

	if public.FirstCheck(msgObj) {
		return process.Do(cache.UserModeChat, msgObj)
	} else {
		return process.Do(cache.UserModeChats, msgObj)
	}

}

// 发送欢迎信息
func ReceiveWelcome(msg *dingbot.ReceiveMsg) error {
	_, err := msg.ReplyToDingtalk(string(dingbot.TEXT), Welcome())
	return err
}

func ReceiveChat(msg *dingbot.ReceiveMsg) error {
	public.UserService.SetUserMode(msg.SenderStaffId, CommandChat)
	_, err := msg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("=====现在进入与👉%s👈单聊的模式 =====", msg.SenderNick))
	if err != nil {
		logger.Warning(fmt.Errorf("send message error: %v", err))
	}
	return err
}

func ReceiveChats(msg *dingbot.ReceiveMsg) error {
	public.UserService.SetUserMode(msg.SenderStaffId, CommandChats)
	_, err := msg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("=====现在进入与👉%s👈串聊的模式 =====", msg.SenderNick))
	if err != nil {
		logger.Warning(fmt.Errorf("send message error: %v", err))
	}
	return err
}

func ReceiveReset(msg *dingbot.ReceiveMsg) error {
	public.UserService.ClearUserMode(msg.SenderStaffId)
	public.UserService.ClearUserSessionContext(msg.SenderStaffId)
	_, err := msg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("=====已重置与👉%s👈的对话模式，可以开始新的对话=====", msg.SenderNick))
	if err != nil {
		logger.Warning(fmt.Errorf("send message error: %v", err))
	}
	return err
}

func ReceiveHelp(msg *dingbot.ReceiveMsg) error {
	return ReceiveHelp(msg)
}

func ReceiveTmpl(msg *dingbot.ReceiveMsg) error {
	var title string
	for _, v := range *public.Prompt {
		title = title + v.Title + " | "
	}
	_, err := msg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("%s 您好，当前程序内置集成了这些prompt：\n====================================\n| %s \n====================================\n你可以选择某个prompt开头，然后进行对话。\n以周报为例，可发送 #周报 我本周用Go写了一个钉钉集成ChatGPT的聊天应用", msg.SenderNick, title))
	if err != nil {
		logger.Warning(fmt.Errorf("send message error: %v", err))
	}
	return err
}

func ReceiveImage(msg *dingbot.ReceiveMsg) error {
	_, err := msg.ReplyToDingtalk(string(dingbot.MARKDOWN), "发送以 **#image** 开头的内容，将会触发绘画能力，图片生成之后，将会保存在程序根目录下的 **images目录** \n 如果你绘图没有思路，可以在这两个网站寻找灵感。\n - [https://lexica.art/](https://lexica.art/)\n- [https://www.clickprompt.org/zh-CN/](https://www.clickprompt.org/zh-CN/)")
	if err != nil {
		logger.Warning(fmt.Errorf("send message error: %v", err))
	}
	return process.ImageGenerate(msg)
}

func useTmpl(msg string)

// case "余额":
// 	cacheMsg := public.UserService.GetUserMode("system_balance")
// 	if cacheMsg == "" {
// 		rst, err := public.GetBalance()
// 		if err != nil {
// 			logger.Warning(fmt.Errorf("get balance error: %v", err))
// 			return err
// 		}
// 		t1 := time.Unix(int64(rst.Grants.Data[0].EffectiveAt), 0)
// 		t2 := time.Unix(int64(rst.Grants.Data[0].ExpiresAt), 0)
// 		cacheMsg = fmt.Sprintf("💵 已用: 💲%v\n💵 剩余: 💲%v\n⏳ 有效时间: 从 %v 到 %v\n", fmt.Sprintf("%.2f", rst.TotalUsed), fmt.Sprintf("%.2f", rst.TotalAvailable), t1.Format("2006-01-02 15:04:05"), t2.Format("2006-01-02 15:04:05"))
// 	}

// 	_, err := rmsg.ReplyToDingtalk(string(dingbot.TEXT), cacheMsg)
// 	if err != nil {
// 		logger.Warning(fmt.Errorf("send message error: %v", err))
// 	}
