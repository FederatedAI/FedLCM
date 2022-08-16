# Contributing

## Welcome
This guide provides FedLCM project contribution guidelines for open source contributors. **Please leave comments / suggestions if you find something is missing or incorrect.**

When contributing to this repository, please first discuss the change you wish to make via issue, email, or any other method with the owners of this repository before making a change. 

Please note we have a code of conduct, please follow it in all your interactions with the project. Please also follow the general FATE community [CONTRIBUTING.md](https://github.com/FederatedAI/FATE-Community/blob/master/CONTRIBUTING.md) guide.

## Contribute Workflow
PR are always welcome, even if they only contain small fixes like typos or a few lines of codes. If there will be a significant effort, please:
1. Document it as a proposal;
2. PR it to https://github.com/FederatedAI/FATE-Community/tree/master/proposal;
3. Get a discussion by creating issues in this project and refer to the proposal;
4. Implement the feature when the proposal got 2+ maintainer approved, refer to the next sections on how to do it;
5. PR to this project's **current develop** branch.

### Fork and clone
First, fork this repository on GitHub to your personal account and clone the code to your local workspace.

Set user to match your GitHub profile name and clone the project:
```
USER={your github profile name}
git clone https://github.com/FederatedAI/FedLCM.git && cd FedLCM

# add your fork
git remote add $USER https://github.com/$USER/FedLCM.git

git fetch -av
```

### Branch
Changes should be made on your own fork in a new branch. The branch should be named XXX-description where XXX is the number of the issue. PR should be rebased on top of main branch without multiple branches mixed into the PR. If your PR do not merge cleanly, use commands listed below to get it up to date.

```
#Suppose `origin` is the origin upstream

git fetch origin
git checkout main
git rebase origin/main
```
Branch from the updated main branch:
```
git checkout -b my_feature main
```

### Keep sync with upstream
Once your branch gets out of sync with the origin/main branch, use the following commands to update:
```
git checkout my_feature
git fetch -a
git rebase origin/main
Please use fetch / rebase (as shown above) instead of git pull. git pull does a merge, which leaves merge commits. These make the commit history messy and violate the principle that commits ought to be individually understandable and useful (see below). You can also consider changing your .git/config file via git config branch.autoSetupRebase always to change the behavior of git pull.
```

### Update the APIs and related documents
Our RESTful APIs are documented with [Swagger](https://swagger.io/)
If your commit that changes the RESTful APIs, make sure to run `make swag` to update the Swagger documents.

### Commit
As FedLCM has integrated the [DCO (Developer Certificate of Origin)](https://probot.github.io/apps/dco/) check tool, contributors are required to sign-off that they adhere to those requirements by adding a Signed-off-by line to the commit messages. Git has even provided a -s command line option to append that automatically to your commit messages, please use it when you commit your changes.
```
$ git commit -s -m 'This is my commit message'
```
Commit your changes if they're ready:
```
git add -A
git commit -s
git push --force-with-lease $USER my_feature
The commit message should follow the convention on How to Write a Git Commit Message. Be sure to include any related GitHub issue references in the commit message. 
```

### Push and Create PR
When ready for review, push your branch to your fork repository on github.com:
```
git push --force-with-lease $USER my_feature
```
Then visit your fork at https://github.com/$USER/FedLCM and click the `Compare & Pull Request` button next to your `my_feature` branch to create a new pull request (PR). The PR should:
1. Ensure all unit test passed;
2. The tittle of PR should highlight what it solves briefly. The description of PR should refer to all the issues that it addresses. Ensure to put a reference to issues (such as `Close #xxx` and `Fixed #xxx`)  Please refer to the [PULL_REQUEST_TEMPLATE.md](./PULL_REQUEST_TEMPLATE.md).

Once your pull request has been opened it will be assigned to one or more reviewers. Those reviewers will do a thorough code review, looking for correctness, bugs, opportunities for improvement, documentation and comments, and style.

Commit changes made in response to review comments to the same branch on your fork.