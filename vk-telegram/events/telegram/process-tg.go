package telegram

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"vk-telegram/clients/telegram"
	"vk-telegram/events"
	"vk-telegram/lib/e"

	"vk-telegram/storage"
)

type Processor struct {
	tg      *telegram.Client
	offset  int
	storage *storage.StorageClient
}

type Meta struct {
	Username string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

func New(client *telegram.Client, hostStorage, pwStorage string) *Processor {
	return &Processor{
		tg:      client,
		storage: storage.NewRedisClient(hostStorage, pwStorage),
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(event events.Event) error {
	if "/start" == event.Text {
		err := p.tg.SendMessage(event.ChatID, msgHello)
		if err != nil {
			return err
		}
		return nil
	}
	if "/help" == event.Text {
		err := p.tg.SendMessage(event.ChatID, msgHelp)
		if err != nil {
			return err
		}
		return nil
	}
	if ok, _ := regexp.MatchString("\\/setKey\\s[A-z0-9]+", event.Text); ok {
		key := strings.Replace(event.Text, "/setKey ", "", -1)
		if _, err := p.storage.GetKey(key); err != nil {
			err = p.tg.SendMessage(event.ChatID, msgWrongKey)
			if err != nil {
				return err
			}
			return nil
		}
		err := p.storage.AddKey(key, strconv.Itoa(event.ChatID))
		if err != nil {
			return err
		}
		err = p.tg.SendMessage(event.ChatID, msgGoodKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func event(u telegram.Update) events.Event {
	res := events.Event{
		ChatID: u.Message.Chat.ID,
		Text:   u.Message.Text,
	}

	res.Meta = Meta{
		Username: u.Message.From.Username,
	}

	return res
}
