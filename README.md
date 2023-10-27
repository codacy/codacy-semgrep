# Codacy Semgrep

This is the docker engine we use at Codacy to have [Semgrep](https://github.com/returntocorp/semgrep) support.

## Usage

You can create the docker by doing:

  ```bash
  docker build -t codacy-semgrep:latest .
  ```

The docker is ran with the following command:

  ```bash
  docker run -it -v $srcDir:/src codacy-semgrep:latest
  ```

## Generate Docs

 1. Update the version in `go.mod`
 2. Install the dependencies:

```bash
go mod download
```

 3. Run the DocGenerator:

```bash
go run ./doc-generator.go &&\
scala-cli doc-generator.sc
```

## Test

We use the [codacy-plugins-test](https://github.com/codacy/codacy-plugins-test) to test our external tools integration.
You can follow the instructions there to make sure your tool is working as expected.

## What is Codacy?

[Codacy](https://www.codacy.com/) is an Automated Code Review Tool that monitors your technical debt, helps you improve your code quality, teaches best practices to your developers, and helps you save time in Code Reviews.

### Among Codacy’s features

- Identify new Static Analysis issues
- Commit and Pull Request Analysis with GitHub, BitBucket/Stash, GitLab (and also direct git repositories)
- Auto-comments on Commits and Pull Requests
- Integrations with Slack, HipChat, Jira, YouTrack
- Track issues in Code Style, Security, Error Proneness, Performance, Unused Code and other categories

Codacy also helps keep track of Code Coverage, Code Duplication, and Code Complexity.

Codacy supports PHP, Python, Ruby, Java, JavaScript, and Scala, among others.

### Free for Open Source

Codacy is free for Open Source projects.
