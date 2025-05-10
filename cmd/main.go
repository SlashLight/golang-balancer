package main

import (
	"fmt"

	"github.com/SlashLight/golang-balancer/internal/config"
)

func main() {
	// TODO [x] загрузить конфиг
	cfg := config.MustLoad()

	fmt.Println(cfg)
	// TODO [] настроить логгер

	// TODO [] создать балансировщик

	// TODO [] запустить сервер
}
