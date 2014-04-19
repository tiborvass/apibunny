This is a quick hack in Go trying to explore the maze at http://apibunny.com.

# Running

## With Docker

	docker build -t apibunny .
	docker run -i apibunny > apibunny.pdf

Open apibunny.pdf

## Without Docker

You'll need to have Go and graphviz installed

	go run main.go > apibunny.dot
	dot -Tpdf apibunny.dot > apibunny.pdf

Open apibunny.pdf
