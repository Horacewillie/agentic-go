package main

import (
	"fmt"

	"github.com/horacewillie/agent-starter/ollama"
)

// "context"
// "fmt"
// "log"
// "time"

func main() {
	ollama.WgSimulation()

	in := ollama.Gen(1, 2, 3, 4)

	c1 := ollama.Sq(in)
	c2 := ollama.Sq(in)

	// out := ollama.Merge(c1, c2)

	// fmt.Println(<-out)

	for n := range ollama.Merge(c1, c2) {
		fmt.Println(n)
	}
	// client := ollama.CreateClient("http://localhost:11434", 5*time.Minute)

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// defer cancel()

	// resp, err := client.Chat(ctx, ollama.ChatRequest{
	// 	Model: "qwen3.5:4b",
	// 	Messages: []ollama.Message{
	// 		{
	// 			Role:    "user",
	// 			Content: "Why is the sky blue?",
	// 		},
	// 	},
	// 	Stream: false,
	// 	Think:  false,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(resp.Message.Content)
}
