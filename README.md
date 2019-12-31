# Documentation

`github-developer-exporter` is a prometheus exporter which talks to GitHub or GitHub Enterprise API to get information about `Organization`, `Repository`, `Issue` and `Pull Request` connected to `Assignee`, `Requested Reviewers`, `State` etc. 

# Why do we create it ?

One of the problem in developer teams is that someone has a lot of tasks comparing with other members.
Unbalanced assign of tasks or of reviews makes them less productivity.
And we have no solution or tool to track time-series task assignee or requested reviewers.
Thus we create it to check developer team conditions through time-series analysis.

# What I can do

We can visualize time-series `Issue` or `Pull Request` with labels via Prometheus.

[Prometheus](https://prometheus.io/) is a open source systems monitoring with time series data.
[Grafana](https://grafana.com/) is a open source analytics and monitoring solution for every database. 

# Installation

# Metrics

| Metric name | Metric type | Labels/tags | Status
| :--- | :--- | :--- | :--- |