This is a very simple app to listen for requests from DockerHub that a new build has been pushed for your application and you need to issue an update.  Ordinarily this was done manually and was the last piece in making a completely automated build cycle.

## System Requirements
Go needs to be installed in order to compile the script.

On Ubuntu/Debian, `apt-get install gccgo-go` will install a copy of Go for you.

Then you need to setup the GOPATH.  Make sure you make it permanent on the OS so you don't have to reset it every time you want to build the app.
`export GOPATH=/usr/local/go`

This is the only outside dependency needed to run this utility.  It's essentially a service router/framework we use on BlueBeak (think Sinatra on Ruby).  You could easily use Go's built-in http library but I was lazy.
`go get	github.com/go-martini/martini`

## Install
Put `main.go` someplace convenient like `/usr/local/go/src`. Then build it with `go build -o /usr/local/bin/dockerhub-hook-pull`.  This will install it in `/usr/local/bin`.  You will also need to put the `conf.json` file in that directory to as the server looks for it on start (see below).  If the config file doesn't exist, the server will complain and exit when you try to start the server.

### conf.json
This is a simple file that acts as a keychain for your hook.  It adds a bit of security to the call so that random people cannot call the endpoint and always update your software.  It's a JSON file that holds a comma separated array of keys.  When you call the endpoint, simply specify one of the keys as your *pull string*.  If the pull string you specify does not exist in the `conf.json` file, the pull request will be ignored.

## Set Script to Run as a Service
Make sure the hook server is set to startup automatically in case the VM you're on gets rebooted.  For Ubuntu:

    cd /etc/init.d
    vi <service_name> blah, blah
    chmod 755 <service_name>

Now it should start automatically the next time the VM is cycled.

## Execute the Hook
This is the URL setup in DockerHub that is called whenever a successful build occurs.  This is always a `POST` operation.

`http://youdomain.com:1966/app/:app_id/env/:env/port/:port/pull/:pull_string`

DockerHub provides a "Test" button that will execute the hook so you can make sure it's working.

## Example Pull Script
This is the script that is executed by the server.  It essentially initiates a pull from the repository, kills any previously running instances then starts up the new one in its place.  Obviously you script will be different and you will probably have different parameters being fed by the server, but hopefully you get the idea.

    # Usage: deploy.sh [dev|qa|demo|stage|prod] [app_name] [port]

    # Pull the repo down.
    echo "Pulling BlueBeak Web repository..."
    docker pull centric/bluebeak-web:$1
    # Stop the currently executing container.
    echo "Stopping any running containers..."
    docker stop bb_web_$1
    # Remove all of the orphaned containers.
    echo "Remove all orphaned and exited containers..."
    docker rm $(docker ps -q -f status=exited)
    # Start the application.
    echo "Starting the application..."
    docker run -e "RACK_ENV=$1" -v /root/$2/log/home/julieweb/log -d -p $3:4567 --name bb_web_$1 centric/bluebeak-web:$1

## TODO

* Have the script notify Slack or similar when the build fails or is successful.
