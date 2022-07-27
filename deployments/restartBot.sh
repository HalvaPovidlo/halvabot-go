#!/bin/bash

cd $(dirname "$0")

while getopts "bmw" option; do
   case $option in
      b) # botapp
         killall botapp
         echo "botapp restarted"
         nohup ./botapp 1>bot.log 2>&1 &
         ;;
      m) # mock
         killall botmock
         echo "botmock restarted"
         nohup ./botmock 1>mock.log 2>&1 &
         ;;
      w) # webhost
         killall webhost
         echo "webhost restarted"
         nohup ./webhost 1>web.log 2>&1 &
         ;;
   esac
done

echo "script finished"
