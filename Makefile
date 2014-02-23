deps:
	godep save -copy=false

deploy:
	git push heroku master

.PHONY: deps deploy
