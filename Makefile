build:
	docker build -t darkhelmetlive/tinderizer:latest .

shell:
	docker run -it darkhelmetlive/tinderizer:latest /bin/bash
