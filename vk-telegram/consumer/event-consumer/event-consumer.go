package eventconsumer

import (
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
	"vk-telegram/events"
	eventsVK "vk-telegram/events/vk"
)

const batchSize = 100

type Consumer struct {
	fetcherSender     events.Fetcher
	processorSender   events.Processor
	fetcherReceiver   events.Fetcher
	processorReceiver events.Processor
	logC              *slog.Logger
}

func (c Consumer) Start(ts int) {
	go c.startSender(ts) // СДЕЛАТЬ УНИВЕРСАЛЬНЫМ ДЛЯ ОТПРАВИТЕЛЯ
	go c.startReceiver() // СДЕЛАТЬ УНИВЕРСАЛЬНЫМ ДЛЯ ПОЛУЧАТЕЛЯ
	select {}
}

func New(fetcherSender events.Fetcher, processorSender events.Processor, fetcherReceiver events.Fetcher, processorReceiver events.Processor) Consumer {
	return Consumer{
		fetcherSender:     fetcherSender,
		processorSender:   processorSender,
		fetcherReceiver:   fetcherReceiver,
		processorReceiver: processorReceiver,
		logC: slog.New(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}
}

func (c Consumer) startSender(ts int) {
	for {
		gotEvents, err := c.fetcherSender.Fetch(ts)
		if err != nil {
			scds := 1
			for i := 0; i < 3; i++ {
				time.Sleep(time.Duration(scds) * time.Second)
				if gotEvents, err = c.fetcherSender.Fetch(ts); err == nil {
					break
				}
				scds *= 2
			}
			if err != nil {
				slog.Error("[ERR] consumer VK:", err)

				continue
			}
		}
		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)

			continue
		}

		metaVk, err := eventsVK.GetMeta(gotEvents[len(gotEvents)-1])
		if err != nil {
			c.logC.Error("[ERR] consumer VK:", err)
			continue
		}
		ts = metaVk.Ts
		if err := c.handleEvents(c.processorSender, gotEvents); err != nil {
			log.Print(err)

			continue
		}

	}
}

func (c Consumer) startReceiver() {
	for {
		gotEvents, err := c.fetcherReceiver.Fetch(batchSize)
		if err != nil {
			scds := 1
			for i := 0; i < 3; i++ {
				time.Sleep(time.Duration(scds) * time.Second)
				if gotEvents, err = c.fetcherReceiver.Fetch(batchSize); err == nil {
					break
				}
				scds *= 2
			}
			if err != nil {
				slog.Error("[ERR] consumer TG:", err)

				continue
			}
		}
		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)

			continue
		}
		if err := c.handleEvents(c.processorReceiver, gotEvents); err != nil {
			log.Print(err)

			continue
		}

	}

}

func (c *Consumer) handleEvents(procerssor events.Processor, evs []events.Event) error {
	var wg sync.WaitGroup
	wg.Add(len(evs))
	for _, event := range evs {
		go func(ev events.Event) {
			c.logC.Info("got new event: ", ev.Text, nil)

			if err := procerssor.Process(ev); err != nil {
				c.logC.Info("can't handle event: ", err)
			}
			wg.Done()
		}(event)
	}
	wg.Wait()
	return nil
}
