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

This is the rollify backend and frontend (A monolith), it has all the logic, it exposes a REST API to integrate with and also a very simple and clean frontend using [HTMX](https://htmx.org/) and [PicoCSS](https://picocss.com/).

### API

You can chek the API docs [here](https://rollify.app/api/v1/apidocs.json). You can use the [Swagger editor online](https://editor-next.swagger.io/) and import that URL to have readable docs.

## Where is running Rollify

Is running on my personal Kubernetes tiny cluster, depending on the usage of the app, I'll find a bigger home for Rollify.

## Architecture

### Storage

Rollify supports multiple storage types

#### Memory

By default it will run with memory based storage, this is useful for development, because you don't need to set up a MySQL database (although if you are developing MySQL storage features you will need to use this storage).

If you want to run a cheap rollify for you and your friends with ephemeral dice rolls, you can run this with a single instance and will do the job.

#### MySQL

You have the schema in [schema][db-schema].

### Events

Rollify uses websockets to notify the clients using events, for example when there is a new dice roll.

#### Memory

By default it will run with a memory based event system, this is useful for development, because you don't need to set up a NATS server.

As with the memory database, this can be used for cheap rollify deployments that only have one instance and don't need to communicate with other rollify instances.

#### NATS

In order that all the users connected to the different Rollify instances receive the events, the publisher needs to notify all the instances, this is done with a NATS pubsub system.

### Scalability

At this moment Rollify scales horizontally, for this you will need:

- MySQL storage for the database.
- NATS message system for the events delivery.

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

[db-schema]: schema
