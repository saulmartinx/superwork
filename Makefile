build: fmt
	@go build

build_linux:
	env GOOS=linux go build -o superwork_linux

dist:
	rm -rf dist
	mkdir -p dist/css
	cp public/css/lib.css dist/css
	minify --output dist/css/app.css public/css/app.css
	mkdir -p dist/js
	cp public/js/lib.js dist/js
	minify --output dist/js/app.js public/js/app.js
	cp -r public/fonts dist/
	cp -r public/documents dist/
	html-minifier \
		--remove-attribute-quotes \
		--remove-comments \
		--collapse-whitespace \
		public/index.html -o dist/index.html
	cp -r public/images dist/

deploy: build_linux dist
	ssh -X root@superwork 'mkdir -p /home/deploy/superwork/public'
	ssh -X root@superwork 'chown -R deploy /home/deploy/superwork/public'
	rsync -r db superwork_linux deploy@superwork:~/superwork/.
	rsync -r -a dist/ deploy@superwork:~/superwork/public
	ssh -X root@superwork 'chown -R www-data /home/deploy/superwork/public'
	ssh -X deploy@superwork 'sudo stop superwork; sudo start superwork'
	rm -rf dist

tail:
	ssh -X deploy@superwork 'tail -f /home/deploy/superwork/superwork.log'

test:
	@go test

fmt:
	@go fmt

run: build
	./superwork

clean:
	rm -rf superwork superwork_linux dist

lint:
	jslint -nomen public/js/app.js
	@html-validator --file=public/index.html --verbose
	csslint public/css/app.css

golint:
	gometalinter

vet:
	go vet

minify:
	minify --output public/js/lib.js public/js/jquery.min.js public/js/bootstrap.min.js \
		public/js/bootstrap-datepicker.js public/js/ie10-viewport-bug-workaround.js \
		public/js/underscore-min.js public/js/backbone-min.js public/js/moment.min.js \
		public/js/js.cookie.js public/js/jquery-ui.js public/js/moment-duration-format.js \
		public/js/et.js
	minify --output public/css/lib.css public/css/bootstrap.min.css public/css/bootstrap-theme.min.css \
		public/css/bootstrap-datepicker3.css public/css/ie10-viewport-bug-workaround.css public/css/theme.css \
		public/css/font-awesome.css public/css/bootstrap-social.css public/css/jquery-ui.css

coverage:
	@go test -coverprofile=coverage.out
	@go tool cover -html=coverage.out
