package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/PullRequestInc/go-gpt3"
)

type payload struct {
	Text string `json:"text"`
}

type response struct {
	EOF   bool   `json:"eof"`
	Error string `json:"error"`
	Text  string `json:"text"`
}

func doJson(client gpt3.Client, r io.Reader, w io.Writer) error {
	enc := json.NewEncoder(w)
	dec := json.NewDecoder(r)
	for {
		var p payload
		err := dec.Decode(&p)
		if err != nil {
			return err
		}

		err = client.ChatCompletionStream(
			context.Background(),
			gpt3.ChatCompletionRequest{
				Messages: []gpt3.ChatCompletionRequestMessage{
					{Role: "user", Content: p.Text},
				},
				MaxTokens:   3000,
				Temperature: gpt3.Float32Ptr(0),
			}, func(resp *gpt3.ChatCompletionStreamResponse) {
				enc.Encode(response{EOF: false, Text: resp.Choices[0].Delta.Content})
			})
		if err != nil {
			err = enc.Encode(response{Error: err.Error()})
			if err != nil {
				return err
			}
			continue
		}

		err = enc.Encode(response{EOF: true})
		if err != nil {
			return err
		}
	}
}

func main() {
	var j bool
	flag.BoolVar(&j, "json", false, "json input/output")
	flag.Parse()

	apiKey := os.Getenv("CHATGPT_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing CHATGPT_API_KEY")
	}

	option := gpt3.WithDefaultEngine(gpt3.GPT3Dot5Turbo)
	client := gpt3.NewClient(apiKey, option)

	if j {
		log.Fatal(doJson(client, os.Stdin, os.Stdout))
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()

		err := client.ChatCompletionStream(
			context.Background(),
			gpt3.ChatCompletionRequest{
				Messages: []gpt3.ChatCompletionRequestMessage{
					{Role: "user", Content: text},
				},
				MaxTokens:   3000,
				Temperature: gpt3.Float32Ptr(0),
			}, func(resp *gpt3.ChatCompletionStreamResponse) {
				fmt.Print(resp.Choices[0].Delta.Content)
			})
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println()
	}
}
