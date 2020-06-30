# Writing and publishing docs

Our docs are written in [MkDocs](https://www.mkdocs.org/) using the [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) theme. Once you've installed `mkdocs` you can preview the docs locally and publish them live, subsequently.

## Local preview 

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

## Publish

When you're done and have committed the changes, you can publish them using the `mkdocs gh-deploy` command.
