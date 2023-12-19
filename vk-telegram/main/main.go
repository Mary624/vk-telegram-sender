package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	clientTelegram "vk-telegram/clients/telegram"
	clientVK "vk-telegram/clients/vk"
	eventconsumer "vk-telegram/consumer/event-consumer"
	eventsTg "vk-telegram/events/telegram"
	eventsVK "vk-telegram/events/vk"

	"github.com/joho/godotenv"
)

func init() {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fullPath := filepath.Join(path, ".env")
	err = godotenv.Load(fullPath)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	hostVk := os.Getenv("HOST_VK")
	tokenVk := os.Getenv("TOKEN_VK")
	v := os.Getenv("V")
	group_id, _ := strconv.Atoi(os.Getenv("GROUP_ID"))
	tgBotHost := os.Getenv("TG_BOT_HOST")
	tokenTg := os.Getenv("TOKEN_TG")
	hostDB := os.Getenv("HOST_DB")
	pwDB := os.Getenv("PW_DB")
	clientVk := clientVK.NewClient(tokenVk, hostVk, v)
	clientTg := clientTelegram.New(tgBotHost, tokenTg)
	ts, err := clientVk.Connect(group_id)
	if err != nil {
		panic("can't connect to VK")
	}
	processorVk := eventsVK.New(clientVk, clientTg, hostDB, pwDB)
	processorTg := eventsTg.New(clientTg, hostDB, pwDB)
	consumer := eventconsumer.New(processorVk, processorVk, processorTg, processorTg)
	consumer.Start(ts)
}
