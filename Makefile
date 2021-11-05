#export GOPATH=/home/server/goprojects3

all: gate center sdk game match battle chat rank video campaign
logic: game match battle chat rank video campaign

gate:
	@echo 'building gate ...'
	@go build -o build/gate ./gopuppy/apps/gate
	@echo 'build gate done'
.PHONY: gate

center:
	@echo 'building center ...'
	@go build -o build/center ./gopuppy/apps/center
	@echo 'build center done'
.PHONY: center

game:
	@echo 'building game ...'
	@go build -o build/game ./apps/game
	@echo 'build game done'
.PHONY: game

match:
	@echo 'building match ...'
	@go build -o build/match ./apps/match
	@echo 'build match done'
.PHONY: match

battle:
	@echo 'building battle ...'
	@go build -o build/battle ./apps/battle
	@echo 'build battle done'
.PHONY: battle

chat:
	@echo 'building chat ...'
	@go build -o build/chat ./apps/chat
	@echo 'build chat done'
.PHONY: chat

rank:
	@echo 'building rank ...'
	@go build -o build/rank ./apps/rank
	@echo 'build rank done'
.PHONY: rank

video:
	@echo 'building video ...'
	@go build -o build/video ./apps/video
	@echo 'build video done'
.PHONY: video

campaign:
	@echo 'building campaign ...'
	@go build -o build/campaign ./apps/campaign
	@echo 'build campaign done'
.PHONY: campaign

sdk:
	@echo 'building sdk ...'
	@go build -o build/sdk ./apps/sdk
	@echo 'build sdk done'
.PHONY: sdk

robot:
	@echo 'building robot ...'
	@go build -o build/robot ./apps/robot
	@echo 'build robot done'
.PHONY: video

