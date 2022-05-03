#!/bin/bash

killall botapp
killall webhost

cd $(dirname "$0")
nohup ./botapp 1>bot.log 2>&1 &
nohup ./webhost 1>web.log 2>&1 &
