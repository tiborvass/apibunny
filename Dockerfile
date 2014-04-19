FROM crosbymichael/golang

RUN apt-get update && apt-get install -y graphviz

ADD main.go /go/src/apibunny/main.go

RUN go install apibunny
RUN /go/bin/apibunny > /apibunny.dot; cat -n /apibunny.dot
RUN dot -Tpdf /apibunny.dot > apibunny.pdf

CMD ["cat", "/apibunny.pdf"]
