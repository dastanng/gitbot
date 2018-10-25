package bot

import (
	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type Bot struct {
	git *github.Client
}

type InitOptions struct {
	Token string
}

func (b *Bot) Initialize(opts InitOptions) {
	// TODO(dunjut)
}

// Run starts the webhook server.
// stopCh channel is used to send interrupt signal to stop it.
func (b *Bot) Run(stopCh <-chan struct{}) {
	// TODO(dunjut)
	glog.Info("starting webhook server...")
}
