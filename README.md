# Documentation

`github-developer-exporter` is a prometheus exporter which talks to GitHub or GitHub Enterprise API to get information about `Organization`, `Repository`, `Issue` and `Pull Request` connected to `Assignee`, `Requested Reviewers`, `State` etc. 

# Why do we create it ?

One of the problem in developer teams is that someone has a lot of tasks comparing with other members.
Unbalanced assign of issues and of reviews makes them less productive.
And we have no solution or tool to track time-series task assignee or requested reviewers.
So we create it to check developer team conditions through time-series analysis via prometheus.

# What I can do

We can visualize time-series `Organization`, `Repository`, `Issue` or `Pull Request` with labels as a Prometheus Exporter.

[Prometheus](https://prometheus.io/) is a open source systems monitoring with time series data.
[Grafana](https://grafana.com/) is a open source analytics and monitoring solution for every database. 

# Installation

This exporter is written in [Go](https://golang.org/), making it easy to build and deploy as a static binary.
You can clone this repository and build yourself or pull image from [DockerHub](https://hub.docker.com/repository/docker/kyobad/github-developer-exporter).

# Metrics

| Metric name | Metric type | Labels/tags | Status
| :--- | :--- | :--- | :--- |
| org_info | gauge | `login`=<\login-field\><br>`name`=<\organization-name\><br>`url`=\<url\><br>`email`=<\organization-email\><br>`blog`=<\blog-url\><br>`created_at`=<\created timestamp\><br>`updated_at`=<\last update timestamp\> | STABLE |
| org_total_repos_count | gauge | `login`=<\login-field\><br>`name`=<\organization-name\><br>`url`=<\url\><br>`email`=<\organization-email\><br>`blog`=<\blog-url\><br>`created_at`=<\created timestamp\><br>`updated_at`=<\last update timestamp\> | STABLE |
| org_public_repos_count | gauge | `login`=<\login-field\><br>`name`=<\organization-name\><br>`url`=<\url\><br>`email`=<\organization-email\><br>`blog`=<\blog-url\><br>`created_at`=<\created timestamp\><br>`updated_at`=<\last update timestamp\> | STABLE |
| org_private_repos_count | gauge | `login`=<\login-field\><br>`name`=<\organization-name\><br>`url`=<\url\><br>`email`=<\organization-email\><br>`blog`=<\blog-url\><br>`created_at`=<\created timestamp\><br>`updated_at`=<\last update timestamp\> | STABLE |
| repo_info | gauge | `org_name`=<\organization-name\><br>`name`=<\repository-name\><br>`full_name`=<\fullname\><br>`owner`=<\organization-owner\><br>`url`=<\repository-url\><br>`default_branch`=<\default-branch\><br>`archived`=<\true or false\><br>`laungage`=<\mainly used laungage\><br>`created_at`=<\created timestamp\><br>`updated_at`=<\last updated timestamp\><br>`pushed_at`=<\last pushed timestamp\> | STABLE |
| repo_open_issue_count | gauge | `org_name`=<\organization-name\><br>`name`=<\repository-name\><br>`full_name`=<\fullname\><br>`owner`=<\organization-owner\><br>`url`=<\repository-url\><br>`default_branch`=<\default-branch\><br>`archived`=<\true or false\><br>`laungage`=<\mainly used laungage\><br>`created_at`=<\created timestamp\><br>`updated_at`=<\last updated timestamp\><br>`pushed_at`=<\last pushed timestamp\> | STABLE |
| issue_info | gauge | `org_name`=<\organization-name\><br>`repo_name`=<\repository-name\><br>`state`=<\open or close\><br>`title`=<\issue-title\><br>`created_at`=<\creation timestamp\><br>`updated_at`=<\last updated timestamp\><br>`closed_at`=<\If not closed, it returns ""\><br>`assignee`=<\if not assigned, it returns ""\><br>`label`=<\labels joined with comma. e.g. "good first issue,help wanted"\> | STABLE |
| pull_request_info | gauge | `org_name`=<\organization-name\><br>`repo_name`=<\repository-name\><br>`state`=\<\open or close\\><br>`title`=<\issue-title\><br>`created_at`=<\creation timestamp\><br>`updated_at`=<\last updated timestamp\><br>`closed_at`=<\If not closed, it returns "".\><br>`assignee`=<\If not assigned, it returns "".\><br>`reviewer`=<\If someone finished review, it does not return them.\><br>`label`=<labels joined with comma. e.g. "good first issue,help wanted"\> | STABLE |
