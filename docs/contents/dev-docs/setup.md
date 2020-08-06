# Setup 

We walk you now through the setup to start contributing to the AWS Controller
for Kubernetes (ACK). No matter if you're contributing code or docs, follow 
the steps below.

!!! tip "Issue before PR"
    Of course we're happy about code drops via PRs, however, in order to give
    us time to plan ahead and also to avoid disappointment, consider creating
    an issue first and submit a PR later. This also helps us to coordinate
    between different contributors and should in general help keeping everyone
    happy.

## Fork the upstream repository

First, fork the [upstream source repository](https://github.com/aws/aws-controllers-k8s) 
into your personal GitHub account. Then, in `$GOPATH/src/github.com/aws/`,
clone your repo and add the upstream like so:

```
git clone git@github.com:$GITHUB_ID/aws-controllers-k8s && \
cd aws-controllers-k8s && \
git remote add upstream git@github.com:aws/aws-controllers-k8s
```

!!! note "Go version"
    We recommend to use a Go version of `1.14` or above for development.

## Create your local branch

Next, you create a local branch where you work on your feature or bug fix.
Let's say you want to enhance the docs, so set `BRANCH_NAME=docs-improve` and
then:

```
git fetch --all && git checkout -b $BRANCH_NAME upstream/main
```

## Commit changes

Make your changes locally, commit and push using:

```
git commit -a -m "improves the docs a lot"

git push origin $BRANCH_NAME
```

With an example output:

```bash
Enumerating objects: 6, done.
Counting objects: 100% (6/6), done.
Delta compression using up to 8 threads
Compressing objects: 100% (4/4), done.
Writing objects: 100% (4/4), 710 bytes | 710.00 KiB/s, done.
Total 4 (delta 2), reused 0 (delta 0)
remote: Resolving deltas: 100% (2/2), completed with 2 local objects.
remote: This repository moved. Please use the new location:
remote:   git@github.com:$GITHUB_ID/aws-controllers-k8s.git
remote: 
remote: Create a pull request for 'docs' on GitHub by visiting:
remote:      https://github.com/$GITHUB_ID/aws-controllers-k8s/pull/new/docs
remote: 
To github.com:a-hilaly/aws-controllers-k8s
 * [new branch]      docs -> docs
```

## Create a pull request

Finally, submit a pull request against the upstream source repository.

Use either the link that show up as in the example above or to the upstream 
source repository and there open the pull request as depicted below:

![images](../images/github-pr.png)

We monitor the GitHub repo and try to follow up with comments within a working
day.

