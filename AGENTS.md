# Goal
This is to be an application that converts markdown with codeblocks to good looking pdf files. We will use the Go programming language to create them.

# Stages
1. Write a plan to PLAN.md and refer to it from this document — **✅ Done: see [PLAN.md](./PLAN.md)**
2. Implement after user verification

# Features
- builds through a devcontainer on github actions
- statically compiles to a binary without dependencies
- uses semver and releases binaries for OSX, Linux, Windows both AMD as well as ARM.
- start with a 0.0.x version until ready for release
- supports embedded images in PNG or SVG format
- supports embedded Mermaid and D2 diagrams
- assures that the generated diagrams look good and have the proper "light" look for a pdf document
- assures that the generated diagrams fit well within the pdf document

# Working Directory
Project root: `/Users/dirk/Documents/projects/markdown2pdf`
All shell commands must use this as the working directory.

# Go Version
This project uses **Go 1.25**. The devcontainer image is `mcr.microsoft.com/devcontainers/go:1.25` and `GOTOOLCHAIN=auto` is set so toolchain downloads happen automatically. Do not upgrade to a newer Go version without explicit approval.
