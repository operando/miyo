package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	crowi "github.com/b4b4r07/go-crowi"
	"github.com/nlopes/slack"
)

const SLACK_API string = "https://slack.com/api/"

// TODO:作成終わったらページを開くようにする(オプション)
// TODO:Threadの開始のメッセージがわかりやすいようにする

func main() {
	configPath := flag.String("cf", "", "Configuration file path")
	channel := flag.String("c", "", "Slack channel to fetch thread from")
	thread_ts := flag.String("t", "", "Unique identifier of a thread's parent message(thread_ts)")
	pagePath := flag.String("p", "", "Create page path")
	flag.Parse()

	var config Config
	_, err := LoadConfig(*configPath, &config)
	if err != nil {
		fmt.Println(err)
		return
	}
	api := slack.New(config.SlackToken)

	values := url.Values{
		"token":     {config.SlackToken},
		"channel":   {*channel},
		"thread_ts": {*thread_ts},
	}

	response, err := threadRequest("channels.replies", values, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Print(response.Messages)

	boby := bytes.NewBufferString("")
	uMap := map[string]slack.User{}
	for _, m := range response.Messages {
		//fmt.Println(m)
		u, ok := uMap[m.User]
		if ok {
			//fmt.Println("hit cache")
			//fmt.Println(u)

			boby.WriteString("## ")
			boby.WriteString("![](")
			boby.WriteString(u.Profile.Image32)
			boby.WriteString(")")
			boby.WriteString(" ")
			boby.WriteString(u.Name)
			boby.WriteString(" ")
			writeTime(boby, &m)
		} else {
			u, err := api.GetUserInfo(m.User)
			if err == nil {
				uMap[m.User] = *u
				//fmt.Println(u)

				boby.WriteString("## ")
				boby.WriteString("![](")
				boby.WriteString(u.Profile.Image32)
				boby.WriteString(")")
				boby.WriteString(" ")
				boby.WriteString(u.Name)
				boby.WriteString(" ")
				writeTime(boby, &m)
			}
		}

		boby.WriteString("\n\n")
		boby.WriteString(m.Text)
		boby.WriteString("\n\n")

	}
	//fmt.Println(boby)

	cfg := crowi.Config{
		URL:   config.Crowi.ApiUrl,
		Token: config.Crowi.Token,
	}
	client, err := crowi.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	res, err := client.Pages.Create(context.Background(), *pagePath, boby.String())
	if err != nil {
		panic(err)
	}

	if !res.OK {
		log.Printf("[ERROR] %s", res.Error)
	}

	fmt.Print("SUCCESS!!")
}

type Thread struct {
	Messages []slack.Message `json:"messages"`
	slack.SlackResponse
}

func threadRequest(path string, values url.Values, debug bool) (*Thread, error) {
	response := &Thread{}
	err := post(path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, err
	}
	return response, nil
}

func parseResponseBody(body io.ReadCloser, intf *interface{}, debug bool) error {
	response, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	// FIXME: will be api.Debugf
	if debug {
		log.Printf("parseResponseBody: %s\n", string(response))
	}

	err = json.Unmarshal(response, &intf)
	if err != nil {
		return err
	}

	return nil
}

func postForm(endpoint string, values url.Values, intf interface{}, debug bool) error {
	resp, err := http.PostForm(endpoint, values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return parseResponseBody(resp.Body, &intf, debug)
}

func post(path string, values url.Values, intf interface{}, debug bool) error {
	return postForm(SLACK_API+path, values, intf, debug)
}

func writeTime(b *bytes.Buffer, m *slack.Message) {
	b.WriteString("<span style=\"color: #999999; font-size: 16px; font-weight:lighter;\">")
	b.WriteString(getTime(m))
	b.WriteString("</span>")
}

func getTime(m *slack.Message) string {
	i, _ := strconv.ParseFloat(m.Timestamp, 64)
	t := time.Unix(int64(i), 0)
	return t.Format("2006/1/2 15:04:05")
}
