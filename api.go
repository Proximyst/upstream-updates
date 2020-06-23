package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type responseCommits struct {
	Size   uint `json:"size"`
	Values []struct {
		DisplayId string `json:"displayId"`
		Author    struct {
			Name string `json:"name"`
		} `json:"author"`
		Message       string `json:"message"`
		CommitterTime uint64 `json:"committerTimestamp"`
	} `json:"values"`
}

func getRepoCommits(project, repo string) (retResp responseCommits, retErr error) {
	url := fmt.Sprintf(botConfig.Endpoint, project, repo)
	resp, err := http.Get(url)
	if err != nil {
		retErr = err
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		retErr = errors.New("status code for " + project + "/" + repo + " is: " + strconv.Itoa(resp.StatusCode))
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		retErr = err
		return
	}

	retErr = json.Unmarshal(body, &retResp)
	return
}

// vim: set ff=unix autoindent ts=4 sw=4 tw=0 noet :
