package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nlopes/slack"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const SLACK_API string = "https://slack.com/api/"

// TODO:tokenやchannelなどをconfig or パラムで設定できるようにする
// TODO:Crowiのページ作成 + 作成終わったらページを開くようにする(オプション)
// TODO:作成終わったらページを開くようにする(オプション)
// TODO:投稿時間を表示する
// TODO:Threadの開始のメッセージがわかりやすいようにする
// TODO:CrowiのページPathの作成方法を考える

func main() {
	token := flag.String("t", "", "token")
	channel := flag.String("c", "", "channel")
	thread_ts := flag.String("ts", "", "thread_ts")
	flag.Parse()

	api := slack.New(*token)

	values := url.Values{
		"token":     {*token},
		"channel":   {*channel},
		"thread_ts": {*thread_ts},
	}

	response, err := threadRequest("channels.replies", values, true)
	if err != nil {
		return
	}

	fmt.Print(response.Messages)

	boby := bytes.NewBufferString("")
	uMap := map[string]slack.User{}
	for _, m := range response.Messages {
		fmt.Println(m)
		u, ok := uMap[m.User]
		if ok {
			fmt.Println("hit cache")
			fmt.Println(u)

			boby.WriteString("## ")
			boby.WriteString(u.Name)
			boby.WriteString(" ")
			boby.WriteString("![](")
			boby.WriteString(u.Profile.Image24)
			boby.WriteString(")")
		} else {
			u, err := api.GetUserInfo(m.User)
			if err == nil {
				uMap[m.User] = *u
				fmt.Println(u)

				boby.WriteString("## ")
				boby.WriteString(u.Name)
				boby.WriteString(" ")
				boby.WriteString("![](")
				boby.WriteString(u.Profile.Image24)
				boby.WriteString(")")
			}
		}

		boby.WriteString("\n\n")
		boby.WriteString(m.Text)
		boby.WriteString("\n\n")

	}
	fmt.Println(boby)
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
