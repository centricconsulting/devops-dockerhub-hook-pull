// This utility listens for incoming POSTs in order to initiate an action like a
// repository pull or a notification.
package main

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// Determine current operating environment.
var (
	m          *martini.Martini
	configFile string
	err        error
	out        []byte
	apiv       string
)

type ExecuteKeys struct {
	Keys []string
}

// init runs before everything else.
func init() {
	// Set the API version.
	apiv = "1.00"
	// Check credentials to make sure this is a legit request.
	configFile = "conf.json"
	_, err := os.Stat(configFile)
	// If there is a problem with the file, err on the side of caution and
	// reject the request.
	if err != nil {
		log.Printf("Error: Could not find configuration file %s", configFile)
		os.Exit(1)
	}

	m = martini.New()
	// Setup Routes
	r := martini.NewRouter()
	r.Post(`/app/:app_id/env/:env/port/:port/pull/:pull_string`, RunPullScript)
	r.Get(`/system/ping`, PingTheApi)
	r.Get(`/system/version`, GetDHHPApiVersion)

	// Add the router action
	m.Action(r.Handle)
} // func

func GetDHHPApiVersion() (int, string) {
	return http.StatusOK, apiv
} // func

func PingTheApi() (int, string) {
	return http.StatusOK, "PONG"
} // func

// RunPullScript will validate the request, then execute the command to pull
// the new Docker build from the repository.
func RunPullScript(params martini.Params) (int, string) {

	// Validate the pull string.  If we can't, stop here.
	if !ValidateKey(params["pull_string"]) {
		log.Printf("Error: Could not validate supplied pull string")
		return http.StatusUnauthorized, "Invalid pull string."
	}

	// Depending on the application id sent in, execute a different command. If
	// the app id doesn't match anything here, stop here.
	switch params["app_id"] {
	case "bbapi","bbweb","tm":
		out, err = exec.Command("/root/"+params["app_id"]+"/deploy.sh", params["env"], params["app_id"], params["port"]).Output()
	default:
		log.Printf("Error: Invalid application Id")
		return http.StatusBadRequest, "Invalid application Id."
	}

	// If there was an error issuing any of the shell commands, stop here.
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return http.StatusBadRequest, err.Error()
	}

	// Let the caller know everything looks good and that the job has been run.
	log.Printf("Info: Pull request completed for the %s environment", params["env"])
	return http.StatusOK, string(out[:])
} // func

// ValidateKey makes sure the string sent in with the request is valid.  This
// hopefully keeps people from messing with it randomly.
func ValidateKey(pkey string) bool {
	Keychain := ExecuteKeys{}

	// Check credentials to make sure this is a legit request.
	file, err := os.Open(configFile)
	// If there is a problem with the file, err on the side of caution and
	// reject the request.
	if err != nil {
		return false
	}

	// Decode the json into something we can process.
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Keychain)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return false
	}
	// Now that we have the keychain, is the requested key in it.
	for _, key := range Keychain.Keys {
		// If the keychain key matches the requested key, we're good.
		if key == pkey {
			return true
		}
	}
	// Everything was cool, but the supplied key simply doesn't match anything.
	return false
} // func

func main() {
	// Let's go!  You can change the listening port to whatever you want.
	m.RunOnAddr(":1966")
} // func
