package main

import (
	"context"
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

func hasTrelloCard(ctx context.Context, client *gapi.Client, owner, repo string, number int) bool {
	opt := &gapi.IssueListCommentsOptions{
		ListOptions: gapi.ListOptions{PerPage: 10},
	}
	for {
		res, resp, err := client.Issues.ListComments(ctx, owner, repo, number, opt)
		if err != nil {
			log.Println("Error getting comments:", err)
		}
		for _, c := range res {
			log.Println(c.GetBody())
			if strings.Contains(c.GetBody(), "trello.com/c/") {
				log.Printf("Issue #%d is attached to a trello card", number)
				return true
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return false
}

func processPR(ctx context.Context, client *gapi.Client, owner, repo string, number int) {
	if !hasTrelloCard(ctx, client, owner, repo, number) {
		return
	}

	log.Printf("Issue #%d is attached to a trello card", number)

	pr, _, err := client.PullRequests.Get(ctx, owner, repo, number)
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
		owner,
		repo,
		pr.GetHead().GetSHA(),
		status,
	)
	if err != nil {
		log.Println("Couldn't clear status on PR", number, err)
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
		payload, err := hook.Parse(r,
			github.IssueCommentEvent,
			github.PullRequestEvent,
			github.PushEvent,
			github.PullRequestReviewEvent,
		)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
				return
			}
		}

		switch payload.(type) {

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			processPR(r.Context(), client, pullRequest.PullRequest.Base.Repo.Owner.Login, pullRequest.PullRequest.Base.Repo.Name, int(pullRequest.PullRequest.Number))
		case github.IssueCommentPayload:
			p := payload.(github.IssueCommentPayload)
			processPR(r.Context(), client, p.Repository.Owner.Login, p.Repository.Name, int(p.Issue.Number))
		case github.PullRequestReviewPayload:
			p := payload.(github.PullRequestReviewPayload)
			processPR(r.Context(), client, p.Repository.Owner.Login, p.Repository.Name, int(p.PullRequest.Number))
		case github.PullRequestReviewCommentPayload:
			p := payload.(github.PullRequestReviewCommentPayload)
			processPR(r.Context(), client, p.Repository.Owner.Login, p.Repository.Name, int(p.PullRequest.Number))
		case github.PushPayload:
			p := payload.(github.PushPayload)

			res, _, err := client.PullRequests.ListPullRequestsWithCommit(r.Context(), p.Repository.Owner.Login, p.Repository.Name, p.After, nil)
			if err != nil {
				return
			}
			for _, pr := range res {
				processPR(r.Context(), client, p.Repository.Owner.Login, p.Repository.Name, *pr.Number)
			}
		}
	})
	http.ListenAndServe(":3000", nil)
}
