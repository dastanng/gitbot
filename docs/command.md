# Command 列表

| Command                                | Example                                  | Description                              | Who can use                              |
| :------------------------------------- | :--------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| /approve [no-issue\|cancel]            | `/approve`<br />`/approve no-issue`      | Approves a pull request                  | Users listed as 'approvers' in appropriate OWNERS files. |
| /[un]cc [[@]...]                       | `/cc`<br />`/uncc`<br />`/cc @dunjut`    | Requests a review from the user(s).      | Anyone can use the command, but the target user must be a member of the org that owns the repository. |
| /lgtm [cancel] or Github Review action | `/lgtm` <br />`/lgtm cancel`<br />['Approve' or 'Request Changes'](https://help.github.com/articles/about-pull-request-reviews/) | Adds or removes the 'lgtm' label which is typically used to gate merging. | Collaborators on the repository. '/lgtm cancel' can be used additionally by the PR author. |
| /[un]assign [[@]...]                   | `/assign`<br />`/unassign`<br />`/assign @dunjut` | Assigns an assignee to the PR.           | Anyone can use the command, but the target user must be a member of the org that owns the repository. |
| /hold [cancel]                         | `/hold`<br />`/hold cancel`              | Adds or removes the `do-not-merge/hold` Label which is used to indicate that the PR should not be automatically merged. | Anyone can use the /hold command to add or remove the 'do-not-merge/hold' Label. |
| /close                                 | `/close`                                 | Closes an issue or PR.                   | Authors and collaborators on the repository can trigger this command. |
| /wip [cancel]                          | `/wip`<br />`/wip cancel`                | Adds or removes the `work-in-progress` label which is used to indicate that the PR is not ready for reviewing or merging. | Only authors can trigger this command.   |
