#!/bin/bash

killall botapp

cd $(dirname "$0")
nohup ./botapp 1>bot.log 2>&1 &
