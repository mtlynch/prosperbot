# ProsperBot

[![Build Status](https://travis-ci.org/mtlynch/prosperbot.svg?branch=master)](https://travis-ci.org/mtlynch/prosperbot)
[![Coverage Status](https://coveralls.io/repos/github/mtlynch/prosperbot/badge.svg?branch=master)](https://coveralls.io/github/mtlynch/prosperbot?branch=master)
[![GoDoc](https://godoc.org/github.com/mtlynch/prosperbot?status.svg)](https://godoc.org/github.com/mtlynch/prosperbot)
[![Go Report Card](https://goreportcard.com/badge/github.com/mtlynch/prosperbot)](https://goreportcard.com/report/github.com/mtlynch/prosperbot)

## Overview

ProsperBot is an automated lending bot for the peer to peer lending platform, [Prosper](https://www.prosper.com).

ProsperBot is primarily a proof-of-concept for using [gofn-prosper](https://github.com/mtlynch/gofn-prosper), a set of Go language bindings I wrote for the [Prosper API](https://developers.prosper.com/docs/investor/). Others can use ProsperBot, but it will probably require some tinkering to make it do what you want.

## Related Repositories

* [gofn-prosper](https://github.com/mtlynch/gofn-prosper): The Go bindings that ProsperBot uses to communicate with the [Prosper API](https://developers.prosper.com/docs/investor/).
* [ansible-role-prosperbot](https://github.com/mtlynch/ansible-role-prosperbot): An Ansible role for installing ProsperBot to an Ubuntu server
* [prosperbot-frontend](https://github.com/mtlynch/prosperbot-frontend): A web app front-end for ProsperBot that displays ProsperBot's state and activity.

## Requirements

* Go 1.5 or above
* Redis 2.x or above
