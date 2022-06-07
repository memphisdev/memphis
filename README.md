<p align="center">
  <a href="https://memphis.dev" target="_blank">
    <img alt="memphis.dev-logo" height="70" alt="memphis.dev Logo" src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/graphics+for+github/color+logo.svg">
  </a>
 </p>
 <p align="center">
  <a href="https://memphis.dev/docs/">Docs</a> - <a href="https://twitter.com/Memphis_Dev">Twitter</a> - <a href="https://www.youtube.com/channel/UCVdMDLCSxXOqtgrBaRUHKKg">YouTube</a>
</p>

<p align="center">
  <a href="https://discord.gg/WZpysvAeTf"><img src="https://img.shields.io/discord/963333392844328961?color=6557ff&label=discord" alt="Discord"></a> <a href=""><img src="https://img.shields.io/github/issues-closed/memphisdev/memphis-broker?color=6557ff"></a> <a href="https://github.com/memphisdev/memphis-broker/blob/master/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Code%20of%20Conduct-v1.0-ff69b4.svg?color=ffc633" alt="Code Of Conduct"></a> <a href="https://github.com/memphisdev/memphis-broker/blob/master/LICENSE"><img src="https://img.shields.io/github/license/memphisdev/memphis-broker?color=ffc633" alt="License"></a> <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/memphisdev/memphis-broker?color=61dfc6"> <img src="https://img.shields.io/github/last-commit/memphisdev/memphis-broker?color=61dfc6&label=last%20commit">
</p>

### Probably the easiest message broker in the world.

**[Memphis{dev}](https://memphis.dev)** is a modern replacement for Apache Kafka.<br>A message broker for developers made out of devs' struggles with using message brokers,<br>building complex data/event-driven apps, and troubleshooting them.<br><br>Allowing developers to achieve all other message brokers' benefits in a fraction of the time.<br>

# Features
**Current**
- Fully optimized message broker in under 3 minutes
- Easy-to-use UI, CLI, and SDKs
- Data-level observability
- Runs on your Docker or Kubernetes

**Coming soon**
- Embedded schema registry using dbt
- Message Journey - Real-time messages tracing
- More SDKs
- Inline processing
- Ready-to-use connectors and analysis functions

# Getting Started
[Installation Videos](https://www.youtube.com/playlist?list=PL_7iYjqhtXpWpZT2U0zDYo2eGOoGmg2mm)<br><br>
Helm for Kubernetes
```shell
helm repo add memphis https://k8s.memphis.dev/charts/ && \
helm install my-memphis memphis/memphis --create-namespace --namespace memphis
```
Docker Compose
```shell
curl -s https://memphisdev.github.io/memphis-docker/docker-compose.yml -o docker-compose.yml && \
docker compose -f docker-compose.yml -p memphis up
```

[An event-driven demo app](https://medium.com/memphis-dev/how-to-build-your-own-wolt-app-b220d738bb71)

# High-Level Architecture
![]( "Architecture")
<img alt="memphis.dev-logo" height="500" alt="memphis.dev Architecture" src="https://memphis-public-files.s3.eu-central-1.amazonaws.com/graphics+for+github/Architecture.png">

# Local access
### Via Kubernetes
```shell
To access Memphis UI from localhost, run the below commands:
  1. kubectl port-forward service/memphis-ui 9000:80 --namespace memphis > /dev/null &

To access Memphis using CLI or SDK from localhost, run the below commands:
  2. kubectl port-forward service/memphis-cluster 7766:7766 6666:6666 5555:5555 --namespace memphis > /dev/null &

Dashboard: http://localhost:9000
Memphis broker: localhost:5555 (Management Port) / 7766 (Data Port) / 6666 (TCP Port)
```
**For Production Environments**
Please expose the UI, Cluster, and Control-plane via k8s ingress / load balancer / nodeport

### Via Docker
Dashboard - http://localhost:9000<br>
Broker - localhost:7766<br>
Control-Plane - localhost:5555/6666<br>

# Beta
Memphis{dev} is currently in Beta version. This means that we are still working on essential features like real-time messages tracing,<br>
Schema registry, and inline processing, as well as making more SDKs and supporting materials.

How does it affect you? Well... mostly it doesn't.<br>
(a) The core of memphis broker is highly stable<br>
(b) We learn&fix fast<br><br>
But we need your love, and any help we can get by stars, PR, feedback, issues, and enhancments.<br>
Read more on https://memphis.dev/docs

# Support

## Ask a question about Memphis{dev} or related

You can ask questions, and participate in discussions about Amplication-related topics in the Amplication Discord channel.

<a href="https://discord.gg/WZpysvAeTf"><img src="https://amplication.com/images/discord_banner_purple.svg" /></a>

## Create a bug report

If you see an error message or run into an issue, please [create bug report](https://github.com/memphisdev/memphis-broker/issues/new?assignees=&labels=type%3A%20bug&template=bug_report.md&title=). This effort is valued and it will help all Memphis{dev} users.


## Submit a feature request

If you have an idea, or you're missing a capability that would make development easier and more robust, please [Submit feature request](https://github.com/memphisdev/memphis-broker/issues/new?assignees=&labels=type%3A%20feature%20request&template=feature_request.md&title=).

If a similar feature request already exists, don't forget to leave a "+1".
If you add some more information such as your thoughts and vision about the feature, your comments will be embraced warmly :)
