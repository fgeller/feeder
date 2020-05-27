ARTIFACT=feeder

clean:
	rm -fv ${ARTIFACT}

build: clean
	go build -o ${ARTIFACT} .

run: build
	./${ARTIFACT} -config ~/.config/feeder/config.json
