package bot

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"k8s.io/client-go/util/workqueue"
)

// Bot struct
type Bot struct {
	secret string
	git    *github.Client
	queue  workqueue.RateLimitingInterface
	cmds   map[string]func(*command) bool
}

// InitOptions struct
type InitOptions struct {
	Token  string
	Secret string
}

// Initialize bot
func (b *Bot) Initialize(opts InitOptions) {
	b.secret = opts.Secret

	// initialize Github client
	b.git = initializeGitClient(opts.Token)

	// initialize working queue
	b.queue = workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
		100*time.Millisecond,
		5*time.Second,
	))

	// initialize command handlers
	b.cmds = map[string]func(*command) bool{
		"/close":       b.cmdClose,
		"/assign":      b.cmdAssign,
		"/unassign":    b.cmdAssign,
		"/cc":          b.cmdCc,
		"/uncc":        b.cmdCc,
		"/hold":        b.cmdHold,
		"/wip":         b.cmdWip,
		"/kind":        b.cmdLabel,
		"/remove-kind": b.cmdLabel,
		"/area":        b.cmdLabel,
		"/remove-area": b.cmdLabel,
	}

	// register webhook handlers
	b.registerHandlers()

	glog.Info("webhook server initialized.")
}

// Run starts the webhook server.
// stopCh channel is used to send interrupt signal to stop it.
func (b *Bot) Run(stopCh <-chan struct{}) {
	go func() {
		<-stopCh
		shutdown = true
		glog.Info("receiving stop signal, shutting down server...")
		for {
			if atomic.LoadInt32(&processing) == 0 {
				glog.Fatalln("webhook server terminated by signal.")
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	go b.worker()

	glog.Info("webhook server started, listening on 0.0.0.0:11111")
	err := http.ListenAndServe(":11111", nil)
	glog.Fatalf("webhook server terminated: %v", err)
}

func (b *Bot) registerHandlers() {
	http.HandleFunc("/webhook", b.handleWebhook)
}

func (b *Bot) worker() {
	for {
		b.processNextItem()
	}
}

// processNextItem fetches a command from queue and reacts.
func (b *Bot) processNextItem() {
	// wait until there is new item in the working queue
	item, _ := b.queue.Get()
	defer b.queue.Done(item)

	c := item.(*command)
	f, ok := b.cmds[c.cmd]
	if !ok {
		// invalid command, ignore
		b.queue.Forget(item)
		return
	}

	// try to run corresponding command
	if f(c) {
		b.queue.Forget(item)
		return
	}

	// failed running command, retry
	if b.queue.NumRequeues(item) < 10 {
		b.queue.AddRateLimited(item)
	} else {
		b.queue.Forget(item)
	}
}

func initializeGitClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return github.NewClient(tc)
}
