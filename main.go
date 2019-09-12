package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	gapi "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"gopkg.in/go-playground/webhooks.v5/github"
)

var (
	path, webhookSecret, personalAccessToken string
)

func init() {
	path = os.Getenv("WEBHOOK_PATH")
	if path == "" {
		path = "/trellobot"
	}

	webhookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Fatal("GITHUB_WEBHOOK_SECRET must be specified!")
	}

	personalAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	if personalAccessToken == "" {
		log.Fatal("GITHUB_ACCESS_TOKEN must be specified!")
	}
}

func main() {
	hook, _ := github.New(github.Options.Secret(webhookSecret))

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: personalAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := gapi.NewClient(tc)

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.IssueCommentEvent, github.PullRequestEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
			}
		}

		switch payload.(type) {

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)

			status := &gapi.RepoStatus{
				State: func() *string {
					s := "pending"
					return &s
				}(),
				Context: func() *string {
					s := "trello/attached-card"
					return &s
				}(),
			}

			_, _, err := client.Repositories.CreateStatus(
				context.Background(),
				pullRequest.PullRequest.Head.Repo.Owner.Login,
				pullRequest.PullRequest.Head.Repo.Name,
				pullRequest.PullRequest.Head.Sha,
				status,
			)
			if err != nil {
				fmt.Println(err)
				return
			}
		case github.IssueCommentPayload:
			p := payload.(github.IssueCommentPayload)

			if strings.Contains(p.Comment.Body, "trello.com/c/") {
				log.Printf("Issue #%d is attached to a trello card",
					p.Issue.Number)

				pr, _, err := client.PullRequests.Get(context.Background(), p.Repository.Owner.Login, p.Repository.Name, int(p.Issue.Number))
				if err != nil {
					log.Println("Couldn't get PR", err)
					return
				}

				status := &gapi.RepoStatus{
					State: func() *string {
						s := "success"
						return &s
					}(),
					Context: func() *string {
						s := "trello/attached-card"
						return &s
					}(),
				}

				_, _, err = client.Repositories.CreateStatus(
					context.Background(),
					p.Repository.Owner.Login,
					p.Repository.Name,
					pr.GetHead().GetSHA(),
					status,
				)
				if err != nil {
					log.Println("Couldn't clear status on PR", p.Issue.Number, err)
					return
				}
			}
		}
	})
	http.ListenAndServe(":3000", nil)
}
