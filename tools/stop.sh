for pid in `cat pids/gate1.pid`; do kill -2 $pid; done
sleep 1s
for pid in `cat pids/chat1.pid`; do kill -2 $pid; done
for pid in `cat pids/video1.pid`; do kill -2 $pid; done
for pid in `cat pids/rank1.pid`; do kill -2 $pid; done
for pid in `cat pids/match1.pid`; do kill -2 $pid; done
for pid in `cat pids/game1.pid`; do kill -2 $pid; done
for pid in `cat pids/game2.pid`; do kill -2 $pid; done
for pid in `cat pids/game3.pid`; do kill -2 $pid; done
for pid in `cat pids/battle1.pid`; do kill -2 $pid; done
for pid in `cat pids/battle2.pid`; do kill -2 $pid; done
for pid in `cat pids/campaign1.pid`; do kill -2 $pid; done
sleep 1s
for pid in `cat pids/center1.pid`; do kill -2 $pid; done
for pid in `cat pids/center2.pid`; do kill -2 $pid; done
ps aux|grep ./build
