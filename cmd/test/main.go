package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/thetechnick/prusa-hub/internal/linkclient"
)

func main() {
	c := linkclient.NewClient(
		linkclient.WithEndpoint("http://mini-1.home.nico-schieder.de/api"),
		linkclient.WithAPIKey("yN4PCLXP9ihWroq"))

	ctx := context.Background()
	printer, err := c.GetPrinter(ctx)
	if err != nil {
		panic(err)
	}
	j, _ := json.MarshalIndent(printer, "", "  ")
	fmt.Println(string(j))
}
