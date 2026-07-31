package notify_pusher

import (
	"context"
	"strings"

	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/artalkjs/artalk/v2/internal/dao"
	"github.com/artalkjs/artalk/v2/internal/entity"
	"github.com/artalkjs/artalk/v2/internal/log"
	"github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/dingding"
	"github.com/nikoksr/notify/service/line"
	"github.com/nikoksr/notify/service/slack"
	"github.com/nikoksr/notify/service/telegram"
)

type NotifyPusherConf struct {
	config.AdminNotifyConf
	Dao *dao.Dao

	// Provide a custom function to bridge the gap between Notify pusher and Email pusher
	EmailPush func(notify *entity.Notify) error
}

type NotifyPusher struct {
	conf     *NotifyPusherConf
	dao      *dao.Dao
	ctx      context.Context
	helper   *notify.Notify
	shoutrrr *router.ServiceRouter
}

func NewNotifyPusher(conf *NotifyPusherConf) *NotifyPusher {
	pusher := &NotifyPusher{
		conf:   conf,
		dao:    conf.Dao,
		ctx:    context.Background(),
		helper: notify.New(),
	}

	pusher.loadHelper()

	return pusher
}

func (pusher *NotifyPusher) loadHelper() {
	var (
		helper = pusher.helper
		conf   = pusher.conf
	)

	// Telegram
	tgConf := conf.Telegram
	if tgConf.Enabled {
		if telegramService, err := telegram.New(tgConf.ApiToken); err == nil {
			telegramService.AddReceivers(tgConf.Receivers...)
			helper.UseServices(telegramService)
		} else {
			log.Error("[Notify] Telegram service init error: ", err)
		}
	}

	// 钉钉
	dingTalkConf := conf.DingTalk
	if dingTalkConf.Enabled {
		dingTalkService := dingding.New(&dingding.Config{Token: dingTalkConf.Token, Secret: dingTalkConf.Secret})
		helper.UseServices(dingTalkService)
	}

	// Slack
	slackConf := conf.Slack
	if slackConf.Enabled {
		slackService := slack.New(slackConf.OauthToken)
		slackService.AddReceivers(slackConf.Receivers...)
		helper.UseServices(slackService)
	}

	// LINE
	LINEConf := conf.LINE
	if LINEConf.Enabled {
		if lineService, err := line.New(pusher.conf.LINE.ChannelSecret, pusher.conf.LINE.ChannelAccessToken); err == nil {
			lineService.AddReceivers(LINEConf.Receivers...)
			helper.UseServices(lineService)
		} else {
			log.Error("[Notify] LINE service init error: ", err)
		}
	}

	// Shoutrrr
	shoutrrrConf := conf.Shoutrrr
	if shoutrrrConf.Enabled {
		urls := normalizeShoutrrrURLs(shoutrrrConf.URLs)
		if len(urls) == 0 {
			log.Warn("[Notify] Shoutrrr service init skipped: no URLs configured")
		} else if shoutrrrSender, err := shoutrrr.CreateSender(urls...); err == nil {
			pusher.shoutrrr = shoutrrrSender
		} else {
			log.Error("[Notify] Shoutrrr service init error: ", err)
		}
	}
}

func normalizeShoutrrrURLs(rawURLs []string) []string {
	urls := make([]string, 0, len(rawURLs))
	for _, rawURL := range rawURLs {
		if url := strings.TrimSpace(rawURL); url != "" {
			urls = append(urls, url)
		}
	}

	return urls
}
