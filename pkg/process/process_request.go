package process

import (
	"fmt"
	"strings"

	// "time"

	"github.com/eryajf/chatgpt-dingtalk/pkg/cache"
	"github.com/eryajf/chatgpt-dingtalk/pkg/dingbot"
	"github.com/eryajf/chatgpt-dingtalk/pkg/logger"
	"github.com/eryajf/chatgpt-dingtalk/public"
	"github.com/solywsh/chatgpt"
)

// 执行处理请求
func Do(mode string, rmsg *dingbot.ReceiveMsg) error {
	// 先把模式注入
	public.UserService.SetUserMode(rmsg.SenderStaffId, mode)
	switch mode {
	case cache.UserModeChat:
		reply, err := chatgpt.SingleQa(rmsg.Text.Content, rmsg.SenderStaffId)
		if err != nil {
			logger.Info(fmt.Errorf("gpt request error: %v", err))
			if strings.Contains(fmt.Sprintf("%v", err), "maximum text length exceeded") {
				public.UserService.ClearUserSessionContext(rmsg.SenderStaffId)
				_, err = rmsg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("请求openai失败了，错误信息：%v，看起来是超过最大对话限制了，已自动重置您的对话", err))
				if err != nil {
					logger.Warning(fmt.Errorf("send message error: %v", err))
					return err
				}
			} else {
				_, err = rmsg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("请求openai失败了，错误信息：%v", err))
				if err != nil {
					logger.Warning(fmt.Errorf("send message error: %v", err))
					return err
				}
			}
		}
		if reply == "" {
			logger.Warning(fmt.Errorf("get gpt result falied: %v", err))
			return nil
		} else {
			reply = strings.TrimSpace(reply)
			reply = strings.Trim(reply, "\n")
			// 回复@我的用户
			msgType := string(dingbot.TEXT)
			if strings.Contains(reply, "```") {
				msgType = string(dingbot.MARKDOWN)
			}
			_, err = rmsg.ReplyToDingtalk(msgType, reply)
			if err != nil {
				logger.Warning(fmt.Errorf("send message error: %v", err))
				return err
			}
		}
	case cache.UserModeChats:
		cli, reply, err := chatgpt.ContextQa(rmsg.Text.Content, rmsg.SenderStaffId)
		if err != nil {
			logger.Info(fmt.Sprintf("gpt request error: %v", err))
			if strings.Contains(fmt.Sprintf("%v", err), "maximum text length exceeded") {
				public.UserService.ClearUserSessionContext(rmsg.SenderStaffId)
				_, err = rmsg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("请求openai失败了，错误信息：%v，看起来是超过最大对话限制了，已自动重置您的对话", err))
				if err != nil {
					logger.Warning(fmt.Errorf("send message error: %v", err))
					return err
				}
			} else {
				_, err = rmsg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("请求openai失败了，错误信息：%v", err))
				if err != nil {
					logger.Warning(fmt.Errorf("send message error: %v", err))
					return err
				}
			}
		}
		if reply == "" {
			logger.Warning(fmt.Errorf("get gpt result falied: %v", err))
			return nil
		} else {
			reply = strings.TrimSpace(reply)
			reply = strings.Trim(reply, "\n")
			// 回复@我的用户
			msgType := string(dingbot.TEXT)
			if strings.Contains(reply, "```") {
				msgType = string(dingbot.MARKDOWN)
			}
			_, err = rmsg.ReplyToDingtalk(msgType, reply)
			if err != nil {
				logger.Warning(fmt.Errorf("send message error: %v", err))
				return err
			}
			_ = cli.ChatContext.SaveConversation(rmsg.SenderStaffId)
		}
	default:

	}
	return nil
}

func ImageGenerate(rmsg *dingbot.ReceiveMsg) error {
	reply, err := chatgpt.ImageQa(rmsg.Text.Content, rmsg.SenderStaffId)
	if err != nil {
		logger.Info(fmt.Errorf("gpt request error: %v", err))
		_, err = rmsg.ReplyToDingtalk(string(dingbot.TEXT), fmt.Sprintf("请求openai失败了，错误信息：%v", err))
		if err != nil {
			logger.Warning(fmt.Errorf("send message error: %v", err))
			return err
		}
	}
	if reply == "" {
		logger.Warning(fmt.Errorf("get gpt result falied: %v", err))
		return nil
	} else {
		reply = strings.TrimSpace(reply)
		reply = strings.Trim(reply, "\n")
		// 回复@我的用户
		_, err = rmsg.ReplyToDingtalk(string(dingbot.MARKDOWN), fmt.Sprintf(">点击图片可旋转或放大。\n![](%s)", reply))
		if err != nil {
			logger.Warning(fmt.Errorf("send message error: %v", err))
			return err
		}
	}
	return nil
}
