build:
	docker build -t darkhelmetlive/tinderizer:latest .

push:
	docker push darkhelmetlive/tinderizer:latest

shell:
	docker run -it darkhelmetlive/tinderizer:latest /bin/bash
