package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	lastCommits = make(map[string]string, 5)
	secondTimer int
)

func init() {
	flag.IntVar(&secondTimer, "interval", 60, "the amount of seconds between checks")
	flag.Parse()
}

func main() {
	readConfigOrPanic()
	readLastCommits()

	upstreamTimer := time.NewTicker(time.Duration(secondTimer) * time.Second)

	channel := make(chan bool)
	for {
		expect := 0
		for proj, repos := range botConfig.Repositories {
			for _, repo := range repos {
				go update(proj, repo, channel)
				expect++
			}
		}

		for expect > 0 {
			<-channel
			expect--
		}
		err := writeLastCommits()
		if err != nil {
			log.Println("could not write last commits", err)
		}
		<-upstreamTimer.C
	}
}

func readLastCommits() {
	if _, err := os.Stat("lastcommits.json"); err != nil {
		return // None; we'll just set the current as newest
	}

	body, err := ioutil.ReadFile("lastcommits.json")
	if err != nil {
		log.Fatalln("could not read last commits:", err)
	}

	err = json.Unmarshal(body, &lastCommits)
	if err != nil {
		log.Fatalln("could not unmarshal last commits:", err)
	}
}

func writeLastCommits() error {
	body, err := json.Marshal(lastCommits)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("lastcommits.json", body, 0755)
}

func update(proj, repo string, channel chan bool) {
	defer func() { channel <- true }()
	commits, err := getRepoCommits(proj, repo)
	if err != nil {
		log.Println("could not read commits!", err)
		return
	}

	if commits.Size == 0 {
		log.Println("commits were size 0?")
		return
	}
	lastCommit, existed := lastCommits[proj+"/"+repo]
	lastCommits[proj+"/"+repo] = commits.Values[0].DisplayId
	if !existed {
		log.Println("new project found, assigning commit", commits.Values[0].DisplayId, "to", proj+"/"+repo)
		return
	}
	if lastCommit == commits.Values[0].DisplayId {
		log.Println("no changes for", proj+"/"+repo)
		return
	}

	var lastIndex uint = 0
	for commits.Values[lastIndex].DisplayId != lastCommit {
		lastIndex++
		if lastIndex >= commits.Size {
			// Uh oh, we've exhausted all commits; let's break it here.
			lastIndex--
			break
		}
	}
	log.Println("found", lastIndex, "new commits to", proj+"/"+repo)

	body := strings.Builder{}
	body.WriteString(proj + "/" + repo + " - Updates found!\n")
	for i := uint(0); i < lastIndex; i++ {
		commit := commits.Values[i]
		commit.Message = strings.SplitN(commit.Message, "\n", 2)[0]
		body.WriteString(commit.DisplayId + " (" + commit.Author.Name + "): " + commit.Message + "\n")
	}

	type WebhookBody struct {
		Content string `json:"content"`
	}
	send, err := json.Marshal(WebhookBody{
		Content: "```\n" + body.String() + "\n```",
	})
	if err != nil {
		log.Println("could not marshal body:", err)
		return
	}
	for _, hook := range botConfig.Webhooks {
		resp, err := http.Post(hook, "application/json", bytes.NewReader(send))
		if err != nil {
			log.Println("could not post to discord:", err)
			continue
		}
		defer resp.Body.Close()
	}
}

// vim: set ff=unix autoindent ts=4 sw=4 tw=0 noet :
