package bot

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

var (
	shutdown   bool
	processing int32
)

// handleWebhook calls serve() to actually process webhook requests.
// after shutdown signal is caught, it stops handling.
func (b *Bot) handleWebhook(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&processing, 1)
	defer atomic.AddInt32(&processing, -1)
	if shutdown {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("server shutting down"))
		return
	}
	b.serve(w, r)
}

// handleAddPresetLabels adds preset labels to owner's repo
func (b *Bot) handleAddPresetLabels(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		owner := r.URL.Query().Get("owner")
		if len(owner) < 1 {
			http.Error(w, "Url Param 'owner' is missing", http.StatusBadRequest)
			return
		}
		repo := r.URL.Query().Get("repo")
		if len(repo) < 1 {
			http.Error(w, "Url Param 'repo' is missing", http.StatusBadRequest)
			return
		}
		if err := b.addPresetLabels(owner, repo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		return
	}

	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	return
}

// serve validates and dispatches webhook events to corresponding plugins.
func (b *Bot) serve(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(b.secret))
	if err != nil {
		glog.Infof("validate payload failed: %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		glog.Infof("parse webhook failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid payload data"))
		return
	}
	switch e := event.(type) {
	case *github.IssueCommentEvent:
		if *e.Action != "created" {
			return
		}
		var (
			owner  = *e.Repo.Owner.Login
			repo   = *e.Repo.Name
			number = *e.Issue.Number
			author = *e.Issue.User.Login
			user   = *e.Comment.User.Login
			cmds   = parseCommentBody(*e.Comment.Body)
		)
		for _, c := range cmds {
			c.owner = owner
			c.ownerType = *e.Repo.Owner.Type
			c.repo = repo
			c.number = number
			c.author = author
			c.user = user
			c.event = e
			// add command to working queue
			b.queue.Add(c)
		}
	case *github.PullRequestReviewCommentEvent:
		if *e.Action != "created" {
			return
		}
		var (
			owner  = *e.Repo.Owner.Login
			repo   = *e.Repo.Name
			number = *e.PullRequest.Number
			author = *e.PullRequest.User.Login
			user   = *e.Comment.User.Login
			cmds   = parseCommentBody(*e.Comment.Body)
		)
		for _, c := range cmds {
			c.owner = owner
			c.ownerType = *e.Repo.Owner.Type
			c.repo = repo
			c.number = number
			c.author = author
			c.user = user
			c.event = e
			// add command to working queue
			b.queue.Add(c)
		}
	default:
	}

}

func parseCommentBody(comment string) []*command {
	if !strings.HasPrefix(comment, "/") {
		return nil
	}

	var cmds []*command
	for _, line := range strings.Split(comment, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if !strings.HasPrefix(line, "/") {
			return nil
		}
		cmdargs := strings.Split(line, " ")
		cmds = append(cmds, &command{
			cmd:  cmdargs[0],
			args: cmdargs[1:],
		})
	}
	return cmds
}
