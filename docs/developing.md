# Developing on AWS Controller for Kubernetes

## Setup git repository

1. Fork the [upstream source repository](github.com/aws/aws-service-operator-k8s) to your private Github account
2. On your local workstation and run:

```bash
cd go/src/github.com/aws
git clone git@github.com:$GITHUB_ID/aws-service-operator-k8s
cd aws-service-operator-k8s
git remote add upstream git@github.com:aws/aws-service-operator-k8s
```

## Create your local branch

```bash
git fetch --all && git checkout -b $BRANCH_NAME upstream/mvp
```

## Make your changes, commit and push

```bash
git add . && git commit
git push origin $BRANCH_NAME
```

example output:
```bash
Enumerating objects: 6, done.
Counting objects: 100% (6/6), done.
Delta compression using up to 8 threads
Compressing objects: 100% (4/4), done.
Writing objects: 100% (4/4), 710 bytes | 710.00 KiB/s, done.
Total 4 (delta 2), reused 0 (delta 0)
remote: Resolving deltas: 100% (2/2), completed with 2 local objects.
remote: This repository moved. Please use the new location:
remote:   git@github.com:$GITHUB_ID/aws-service-operator-k8s.git
remote: 
remote: Create a pull request for 'docs' on GitHub by visiting:
remote:      https://github.com/$GITHUB_ID/aws-service-operator-k8s/pull/new/docs
remote: 
To github.com:a-hilaly/aws-service-operator-k8s
 * [new branch]      docs -> docs
```

## Submit a pull request against [upstream source repository](github.com/aws/aws-service-operator-k8s)

Either the link that show up as in the example above orgo to the upstream source repository and open the Pull Request. You'll see a link like the image below:

![image](./images/github-pr.png)