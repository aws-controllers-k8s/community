🟩 [View the live docs](https://aws.github.io/aws-controllers-k8s/)

# Writing and publishing docs

Our docs are written in [MkDocs](https://www.mkdocs.org/) using the [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) theme. 

MkDocs is a static site generator. It converts markdown files to static html pages. Edit the markdown files, and view the rendered site with MkDocs. 

## Automatic deployments with github actions 

Commits that change a file under the `docs/` directory trigger a github action to build the site, and deploy it to github pages. 

The github action is [mkdocs-deploy-gh-pages](https://github.com/mhausenblas/mkdocs-deploy-gh-pages).

🟧 Running `mkdocs gh-deploy` locally will have no effect, the github action will overwrite it. 

## Build the docs locally

### Install mkdocs-material

🟧 The ACK docs require the mkdocs-material theme. 

To ensure you are using a compatible version of mkdocs and the mkdocs-material theme, use a python virtual environment. Then, install mkdocs-material.

```
python3 -m venv venv 
source venv/bin/activate
pip install mkdocs-material
```

Visit the python docs to learn more about [virutal environments (venv)](https://docs.python.org/3/library/venv.html).

Optionally, use a global `.gitignore` to hide the `venv` directory. 

### Local preview 

🟧 Test locally, including both the content and the navigation structure. 

The docs are just Markdown files and in order to see the rendered preview locally (before PRing the repo), do:

```
 $ mkdocs serve
INFO    -  Building documentation...
INFO    -  Cleaning site directory
WARNING -  A relative path to 'user-docs.md' is included in the 'nav' configuration, which is not found in the documentation files
WARNING -  A relative path to 'dev-docs.md' is included in the 'nav' configuration, which is not found in the documentation files
WARNING -  A relative path to 'discussions.md' is included in the 'nav' configuration, which is not found in the documentation files
[I 200630 14:56:59 server:296] Serving on http://127.0.0.1:8000
[I 200630 14:56:59 handlers:62] Start watching changes
[I 200630 14:56:59 handlers:64] Start detecting changes

```

Problems with the local preview? Check the [Material for MkDocs changelog](https://squidfunk.github.io/mkdocs-material/upgrading/).

### Writing 

The [MkDocs reference](https://www.mkdocs.org/user-guide/writing-your-docs/) includes information on the structure of the `docs/` folder, and writing in markdown.

Review the [Material for MkDocs reference](https://squidfunk.github.io/mkdocs-material/reference/formatting/) for information on the theme and formatting.

### Publish

To publish updated docs, commit changes to the markdown files and open a pull request. When your commits are merged, a github action will automatically build and deploy the site.

[View the status of github action runs.](https://github.com/aws/aws-controllers-k8s/actions)