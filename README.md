# Rollify [![Build Status][ci-image]][ci-url] [![Go Report Card][goreport-image]][goreport-url]

[ci-image]: https://github.com/rollify/rollify/workflows/CI/badge.svg
[ci-url]: https://github.com/rollify/rollify/actions
[goreport-image]: https://goreportcard.com/badge/github.com/rollify/rollify
[goreport-url]: https://goreportcard.com/report/github.com/rollify/rollify

## Introduction

Rollify is a web applicaiton manage your dice rolls when you play online tabletop role games with your friends.

It exposes a REST API for this. Main features:

- Create Rooms
- Throw dice rolls in the room that will be received by all the connected users.
- Dice roll history for the room (by date, user...)
- No registration required (only a room link needs to be shared).
- Compatible dice: d4, d6, d8, d10, d12, d20.
- Different dice combinations.
- Open source
- Available online in https://rollify.app

## App

This is the rollify backend, it has all the logic, it exposes a REST API that the frontends and CLIs can use to manage dice rolls, rooms...

## Why this?

During the 2020 COVID19 pandemic, some friends and I started playing tabletop role games ([Shadowrun]) online, the current dice apps for multiple connected users at the same time are not the best..., so I decided to make this app for multiple reasons:

- Play with my friends to tabletop role games online using a proper dice rolling app.
- Improve the online tabletop role playing experience for other people.
- Learn a little bit of fronted (I'm more of black screens :P).

Apart from this, some people reach me in the past that they don't know where to start to create a production ready application in Go, well, there you go! :) If you check the source code you will find:

- Decouple and clean source code.
- Simple and flexible implementation.
- Easy to understand structure.
- Metrics.
- Tests.
- SQL storage usage.
- Alternative implementation for fast development.
- And more! ...

## Where is running https://rollify.app

Is running on my personal Kubernetes tiny cluster, depending on the usage of the app, I'll find a bigger home for Rollify.

## Architecture

### Storage

Rollify supports multiple storage types

#### Memory

By default it will run with memory based storage, this is useful for development, because you dpon't need to set up a MySQL database (although if you are developing MySQL storage features you will need to use this storage).

If you want to run a cheap rollify for you and your friends with ephemeral dice rolls, you can run this with a single instance and will do the job.

#### MySQL

You have the schema in [schema][db-schema].

### Scalability

At this moment using MySQL you can scale horizontally rollify with multiple instances, the frontend uses polling to get the dice rolls every few seconds (websockets usage is being developed at this moment).

### Metrics

It has Prometheus metrics for almost all the internal components on the `:8081/metrics` endpoint:

- HTTP API.
- Application services (domain logic).
- Storage.
- Dice roller (this important we want evenlly distributed dice rolls).
- Event stream (events, websockets...).

## Development

Download the source code and run, this should be enough for most of the cases:

```bash
go run ./cmd/rollify/ --development --debug
```

You can use:

- `make test`: Runs all the tests.
- `make check`: Checks the source code.
- `make gen`: Generates everything required by the app (e.g mocks).

## Deploy on Kubernetes

TODO.

[shadowrun]: https://en.wikipedia.org/wiki/Shadowrun
[db-schema]: schema
