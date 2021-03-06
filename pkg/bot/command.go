package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"

	"github.com/dastanng/gitbot/pkg/bot/labels"
)

type command struct {
	owner     string // repo owner
	ownerType string // type of repo owner

	repo   string // repo name
	number int    // number of issue (or pullrequest)
	author string // author of issue (or pullrequest)
	user   string // command user

	cmd  string   // command name
	args []string // command arguments. optional

	event interface{} // github event
}

func (c *command) succeed() string {
	return fmt.Sprintf("%s - succeed!", c.info())
}

func (c *command) invalid() string {
	return fmt.Sprintf("%s - invalid!", c.info())
}

func (c *command) failed() string {
	return fmt.Sprintf("%s - failed!", c.info())
}

func (c *command) info() string {
	return fmt.Sprintf("[%s/%s #%d(%s)] %s: %s %s",
		c.owner, c.repo, c.number, c.author,
		c.user, c.cmd, strings.Join(c.args, " "),
	)
}

// argsToUsers parses user list from command args
// if empty args, return c.user
func (c *command) argsToUsers() []string {
	var users []string
	for _, arg := range c.args {
		if u := strings.TrimPrefix(arg, "@"); len(u) > 0 {
			users = append(users, u)
		}
	}
	if len(users) == 0 {
		users = append(users, c.user)
	}
	return users
}

// cmdClose handles command /close
func (b *Bot) cmdClose(c *command) bool {
	// check command syntax
	if len(c.args) != 0 {
		glog.Info(c.invalid())
		return true
	}

	ctx := context.Background()

	// close command can only be used by authors or collaborators
	if c.user != c.author {
		isCollab, _, err := b.git.Repositories.IsCollaborator(ctx, c.owner, c.repo, c.user)
		if err != nil {
			glog.Errorf("%s err: %v", c.failed(), err)
			return false
		}
		if !isCollab {
			glog.Infof("%s user is neither author nor a collaborator, ignore.", c.failed())
			return true
		}
	}

	// close issue as user requested
	state := new(string)
	*state = "closed"
	if _, _, err := b.git.Issues.Edit(ctx, c.owner, c.repo, c.number, &github.IssueRequest{State: state}); err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}
	glog.Info(c.succeed())
	return true
}

// cmdAssign handles command /[un]assign [[@]...]
func (b *Bot) cmdAssign(c *command) bool {
	// check command syntax
	if len(c.args) > 1 {
		glog.Info(c.invalid())
		return true
	}

	// ignore assign command when repo owner is not an organization
	if c.ownerType != "Organization" {
		glog.Infof("repo owner is not an organization, ignore.")
		return true
	}

	ctx := context.Background()
	assignee := c.user
	if len(c.args) == 1 {
		assignee = strings.TrimPrefix(c.args[0], "@")
	}

	// TODO(dunjut) check membership

	// assign/unassign issue to/from assignee as requested.
	var err error
	if c.cmd == "/assign" {
		_, _, err = b.git.Issues.AddAssignees(ctx, c.owner, c.repo, c.number, []string{assignee})
	} else { // /unassign
		_, _, err = b.git.Issues.RemoveAssignees(ctx, c.owner, c.repo, c.number, []string{assignee})
	}
	if err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}
	glog.Info(c.succeed())
	return true
}

// cmdCc handles command /[un]cc [[@]...]
func (b *Bot) cmdCc(c *command) bool {
	var err error
	ctx := context.Background()

	var validUsers []string
	for _, usr := range c.argsToUsers() {
		// author is not allowed to be a reviewer
		if usr == c.author {
			continue
		}
		// validates if user is a 'member' or 'collaborator' of owner/repo
		isMember, err := b.isMember(c.owner, c.repo, usr)
		if err != nil {
			glog.Errorf("%s err: %v", c.failed(), err)
			return false
		}
		if isMember {
			validUsers = append(validUsers, usr)
		}
	}

	if len(validUsers) == 0 {
		return true
	}

	reviewersRequest := github.ReviewersRequest{Reviewers: validUsers}
	if c.cmd == "/cc" {
		_, _, err = b.git.PullRequests.RequestReviewers(ctx, c.owner, c.repo, c.number, reviewersRequest)
	} else { // /uncc
		_, err = b.git.PullRequests.RemoveReviewers(ctx, c.owner, c.repo, c.number, reviewersRequest)
	}

	if err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}

	glog.Info(c.succeed())
	return true
}

// isMember validates if user is a 'member' or 'collaborator' of owner/repo
func (b *Bot) isMember(owner, repo, user string) (bool, error) {
	ctx := context.Background()

	// make sure user is a member of an organization
	isMember, _, err := b.git.Organizations.IsMember(ctx, owner, user)
	if err != nil {
		return false, err
	}

	if !isMember {
		// make sure user is a collaborator of a repo
		isCollab, _, err := b.git.Repositories.IsCollaborator(ctx, owner, repo, user)
		if err != nil {
			return false, err
		}
		if !isCollab {
			return false, nil
		}
	}
	return true, nil
}

// cmdHold handles command /hold [cancel]
func (b *Bot) cmdHold(c *command) bool {
	var err error
	ctx := context.Background()

	// check command syntax
	if len(c.args) > 1 {
		glog.Info(c.invalid())
		return true
	}

	var isCancel bool
	if len(c.args) == 1 {
		if c.args[0] != "cancel" {
			glog.Info(c.invalid())
			return true
		}
		isCancel = true
	}

	if !isCancel { // /hold
		_, _, err = b.git.Issues.AddLabelsToIssue(ctx, c.owner, c.repo, c.number, []string{labels.Hold})
	} else { // /hold cancel
		_, err = b.git.Issues.RemoveLabelForIssue(ctx, c.owner, c.repo, c.number, labels.Hold)
	}

	if err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}

	glog.Info(c.succeed())
	return true
}

// cmdWip handles command /wip [cancel]
func (b *Bot) cmdWip(c *command) bool {
	var err error
	ctx := context.Background()

	// check command syntax
	if len(c.args) > 1 {
		glog.Info(c.invalid())
		return true
	}

	var isCancel bool
	if len(c.args) == 1 {
		if c.args[0] != "cancel" {
			glog.Info(c.invalid())
			return true
		}
		isCancel = true
	}

	if !isCancel { // /wip
		_, _, err = b.git.Issues.AddLabelsToIssue(ctx, c.owner, c.repo, c.number, []string{labels.WorkInProgress})
	} else { // /wip cancel
		_, err = b.git.Issues.RemoveLabelForIssue(ctx, c.owner, c.repo, c.number, labels.WorkInProgress)
	}

	if err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}

	glog.Info(c.succeed())
	return true
}

// cmdLabel handles command /[remove-](kind|area|task)
func (b *Bot) cmdLabel(c *command) bool {
	var err error
	ctx := context.Background()

	// check command syntax
	if len(c.args) != 1 {
		glog.Info(c.invalid())
		return true
	}

	if len(c.args[0]) == 0 {
		glog.Info(c.invalid())
		return true
	}

	// user should add / remove label from available repo labels
	recognizedLabels, err := b.getRepoLabels(c.owner, c.repo)
	if err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}

	// remove command type prefix '/' and 'remove-'
	cmdSuffix := strings.TrimPrefix(c.cmd[1:], "remove-")

	// do not add new label from cmd args
	cmdLabel := strings.ToLower(fmt.Sprintf("%s/%s", cmdSuffix, c.args[0]))
	if _, ok := recognizedLabels[cmdLabel]; !ok {
		return true
	}

	isRemove := false
	if len(cmdSuffix) != len(strings.TrimPrefix(c.cmd, "/")) {
		isRemove = true
	}

	if isRemove {
		_, err = b.git.Issues.RemoveLabelForIssue(ctx, c.owner, c.repo, c.number, cmdLabel)
	} else {
		_, _, err = b.git.Issues.AddLabelsToIssue(ctx, c.owner, c.repo, c.number, []string{cmdLabel})
	}

	if err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}

	glog.Info(c.succeed())
	return true
}

// getRepoLabels returns labels from repo
func (b *Bot) getRepoLabels(owner, repo string) (map[string]*github.Label, error) {
	ctx := context.Background()

	lables := make(map[string]*github.Label)
	opt := &github.ListOptions{Page: 1, PerPage: 100}
	for opt.Page > 0 {
		list, resp, err := b.git.Issues.ListLabels(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}

		for _, l := range list {
			lables[strings.ToLower(*l.Name)] = l
		}
		opt.Page = resp.NextPage
	}
	return lables, nil
}

// cmdLgtm handles command /lgtm [cancel]
func (b *Bot) cmdLgtm(c *command) bool {
	var err error
	ctx := context.Background()

	// check command syntax
	if len(c.args) > 1 {
		glog.Info(c.invalid())
		return true
	}

	// validates if user is a 'member' or 'collaborator' of owner/repo
	isMember, err := b.isMember(c.owner, c.repo, c.user)
	if err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}

	if !isMember {
		glog.Infof("user %s is not a member or collaborator of %s/%s, ignore.", c.user, c.owner, c.repo)
		return true
	}

	var isCancel bool
	if len(c.args) == 1 {
		if c.args[0] != "cancel" {
			glog.Info(c.invalid())
			return true
		}
		isCancel = true
	}

	if !isCancel { // /lgtm
		_, _, err = b.git.Issues.AddLabelsToIssue(ctx, c.owner, c.repo, c.number, []string{labels.LGTM})
	} else { // /lgtm cancel
		_, err = b.git.Issues.RemoveLabelForIssue(ctx, c.owner, c.repo, c.number, labels.LGTM)
	}

	if err != nil {
		glog.Errorf("%s err: %v", c.failed(), err)
		return false
	}

	glog.Info(c.succeed())
	return true
}
