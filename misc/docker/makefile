BUILD_SERVICE=build
APP_SERVICE=unix-logger
all: build push

push: 
	docker-compose push ${APP_SERVICE}

build:
	docker-compose up ${BUILD_SERVICE} &&\
	docker-compose build ${APP_SERVICE}