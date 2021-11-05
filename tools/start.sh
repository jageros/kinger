#! /bin/sh
#
# start.sh
# Copyright (C) 2018 Register <registerdedicated(at)gmail.com>
#
# Distributed under terms of the GPLv3 license.
#

mkdir -p logs
touch logs/kingwar.log
mkdir -p pids
export GOPATH=/home/server/goprojects2
nohup ./build/center --id 1 2 >> logs/kingwar.log & echo $! > pids/center1.pid
sleep 1s
nohup ./build/game --id 1 2 >> logs/kingwar.log & echo $! > pids/game1.pid
#nohup ./build/game --id 2 2 >> logs/kingwar.log & echo $! > pids/game2.pid
#nohup ./build/game --id 3 2 >> logs/kingwar.log & echo $! > pids/game3.pid
nohup ./build/battle --id 1 2 >> logs/kingwar.log & echo $! > pids/battle1.pid
nohup ./build/match --id 1 2 >> logs/kingwar.log & echo $! > pids/match1.pid
nohup ./build/chat --id 1 2 >> logs/kingwar.log & echo $! > pids/chat1.pid
nohup ./build/video --id 1 2 >> logs/kingwar.log & echo $! > pids/video1.pid
nohup ./build/rank --id 1 2 >> logs/kingwar.log & echo $! > pids/rank1.pid
nohup ./build/campaign --id 1 2 >> logs/kingwar.log & echo $! > pids/campaign1.pid
nohup ./build/gate --id 1 2 >> logs/kingwar.log & echo $! > pids/gate1.pid
ps aux|grep ./build

