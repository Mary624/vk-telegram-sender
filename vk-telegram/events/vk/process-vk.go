package vk

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"vk-telegram/clients/telegram"
	"vk-telegram/clients/vk"
	"vk-telegram/events"
	"vk-telegram/lib/e"
	"vk-telegram/storage"
)

type Processor struct {
	vkClient       *vk.Client
	telegramClient *telegram.Client
	storage        *storage.StorageClient
}

type Meta struct {
	Attachments []vk.Attachment
	Ts          int
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

func New(vkClient *vk.Client, telegramClient *telegram.Client, hostStorage, pwStorage string) *Processor {
	return &Processor{
		vkClient:       vkClient,
		telegramClient: telegramClient,
		storage:        storage.NewRedisClient(hostStorage, pwStorage),
	}
}

func (p *Processor) Fetch(ts int) ([]events.Event, error) {
	updates, err := p.vkClient.Updates(ts)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates.Updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates.Updates))

	for _, u := range updates.Updates {
		ts, _ := strconv.Atoi(updates.Ts)
		res = append(res, event(u, ts))
	}

	return res, nil
}

func (p *Processor) Process(event events.Event) error {
	meta, err := GetMeta(event)
	if "/start" == event.Text {
		err = p.vkClient.SendMessage(msgHello, event.ChatID, meta.Ts)
		if err != nil {
			return err
		}
		return nil
	}
	if "/help" == event.Text {
		err = p.vkClient.SendMessage(msgHelp, event.ChatID, meta.Ts)
		if err != nil {
			return err
		}
		return nil
	}
	if ok, _ := regexp.MatchString("\\/setKey\\s[A-z0-9]+", event.Text); ok {
		// мб кэшировать
		if _, err := p.storage.GetKey(strconv.Itoa(event.ChatID)); err != nil {
			key := strings.Replace(event.Text, "/setKey ", "", -1)
			err = p.storage.AddKey(strconv.Itoa(event.ChatID), key)
			if err != nil {
				return err
			}

			err = p.storage.AddKey(key, "")
			if err != nil {
				return err
			}

			err = p.vkClient.SendMessage(msgCreated, event.ChatID, meta.Ts)
			if err != nil {
				return err
			}
			return nil
		}

		p.vkClient.SendMessage(msgKeyExist, event.ChatID, meta.Ts)
		if err != nil {
			return err
		}
		return nil
	}
	if res, err := p.storage.GetKey(strconv.Itoa(event.ChatID)); res == "" {
		err = p.vkClient.SendMessage(msgCantGetKey, event.ChatID, meta.Ts)
		if err != nil {
			return err
		}
		return nil
	}

	//ПО CHAT_ID TG!!! НУЖНО ADDKEY(KEY, CHAT_ID)
	// ОТПРАВКА В ТЕЛЕГУ
	key, err := p.storage.GetKey(strconv.Itoa(event.ChatID))
	chatIdTg, err := p.storage.GetKey(key)
	if err != nil || chatIdTg == "" {
		err = p.vkClient.SendMessage(msgCantFindTgKey, event.ChatID, meta.Ts)
		if err != nil {
			return err
		}
		return nil
	}
	id, err := strconv.Atoi(chatIdTg)
	if err != nil {
		return err
	}
	if event.Text != "" {
		p.telegramClient.SendMessage(id, event.Text)
	}
	if len(meta.Attachments) != 0 {
		for _, att := range meta.Attachments {
			if att.Type == "photo" {
				urlPhoto := att.Photo.Sizes[len(att.Photo.Sizes)-1]
				p.telegramClient.SendMessage(id, urlPhoto.URL)
			}
		}
	}

	return nil
}

func GetMeta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func event(u vk.Update, ts int) events.Event {
	res := events.Event{
		Text:   u.Object.Message.Text,
		ChatID: u.Object.Message.ChatID,
	}

	res.Meta = Meta{
		Attachments: u.Object.Message.Attachments,
		Ts:          ts,
	}

	return res
}
