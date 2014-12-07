// Avoid procrastination and speedup productive pages
// Author: Izidor Matušov <izidor.matusov@gmail.com>
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"
)

const HOSTS_FILE = "/etc/hosts"
const UPSTART_FILE = "/etc/init/local_server.conf"
const LISTEN_IP = "127.0.42.42"

var REDIRECTS = map[string]string {
	"c": "https://www.google.com/calendar/render",
	"d": "https://drive.google.com",
	"m": "https://inbox.google.com",
}

var ANTI_PROCRASTINATION = []string{
	"facebook.com",
	"news.ycombinator.com",
	"9gag.com",
	"twitter.com",
}

// Check if we have a root user
func isRoot() bool {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Can't get the current user", err)
	}
	return usr.Uid == "0" || usr.Gid == "0" || usr.Username == "root"
}

func findAliases(r io.Reader, aliases []string) (
		foundAliases []string, err error) {
	comments := regexp.MustCompile("#.*")
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Remove comments
		line = comments.ReplaceAllString(line, "")

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) < 2 {
			return nil, errors.New("Invalid hostfile line " + line)
		}
		ip, hosts := fields[0], fields[1:]
		isLocal := ip == LISTEN_IP
		for _, host := range hosts {
			for _, alias := range aliases {
				if host == alias {
					if !isLocal {
						return nil, errors.New("Alias " + host +
							" is assigned non-local IP " + ip)
					}
					foundAliases = append(foundAliases, host)
				}
			}
		}
	}
	return
}

func getExpectedAliases() (aliases []string) {
	for alias, _ := range REDIRECTS {
		aliases = append(aliases, alias)
	}
	for _, alias := range ANTI_PROCRASTINATION {
		if !strings.HasPrefix(alias, "www.") {
			aliases = append(aliases, "www." + alias)
		}
		aliases = append(aliases, alias)
	}
	return aliases
}

func setUpHostsFile() {
	f, err := os.Open(HOSTS_FILE)
	if err != nil {
		log.Fatal("Can't open", HOSTS_FILE, err)
	}

	expectedAliases := getExpectedAliases()
	aliases, err := findAliases(f, expectedAliases)
	if err != nil {
		log.Fatal("Can't parse", HOSTS_FILE, ":", err)
	}

	// Generate missing aliases
	missingAliases := []string{}
	for _, expectedAlias := range expectedAliases {
		found := false
		for _, alias := range aliases {
			if expectedAlias == alias {
				found = true
				break
			}
		}
		if !found {
			missingAliases = append(missingAliases, expectedAlias)
		}
	}

	if len(missingAliases) > 0 {
		content := fmt.Sprintf(`
# Local server config
%s	%s
`, LISTEN_IP, strings.Join(missingAliases, " "))
		log.Printf("Adding hostfile content %q", content)

		hostfile, err := os.OpenFile(HOSTS_FILE, os.O_RDWR | os.O_APPEND, 0660)
		if err != nil {
			log.Fatal("Can't open", HOSTS_FILE, "for appending")
		}
		defer hostfile.Close()
		if _, err := hostfile.Write([]byte(content)); err != nil {
			log.Fatal("Can't write content to", HOSTS_FILE)
		}
	}
}

func setUpDaemon() {
	if _, err := os.Stat(UPSTART_FILE); os.IsNotExist(err) {
		upstart_file, err := os.OpenFile(UPSTART_FILE, os.O_RDWR | os.O_CREATE, 0640)
		if err != nil {
			log.Fatal("Can't create", UPSTART_FILE)
		}
		log.Print("Creating ", UPSTART_FILE)
		defer upstart_file.Close()
		content := `
description "Local server redirection"
author      "Izidor Matušov"
 
start on (local-filesystems and net-device-up)
stop on runlevel [!2345]
 
respawn
exec /usr/local/bin/local_server
`
		if _, err := upstart_file.Write([]byte(content)); err != nil {
			log.Fatal("Can't write content to", UPSTART_FILE)
		}
	}
}

// Install app itself
func install() {
	if !isRoot() {
		log.Fatal("Needs to be root for installation")
	}
	setUpHostsFile()
	setUpDaemon()
}

func antiProcrastinationHandler(w http.ResponseWriter, r *http.Request) {
	domain := r.Host

	var content string

	n := rand.Intn(100)
	switch {
		case n <= 10:
			content = `
<p>Look what it does to you&hellip;</p>
<p><iframe width="560" height="315" src="//www.youtube.com/embed/Naj5NIVl4mw" frameborder="0" allowfullscreen></iframe></p>`
		case n <= 20:
			content = `
<p>You don't want to look like this, do you?</p>
<p><img src="http://img-9gag-ftw.9cache.com/photo/ad6eZdQ_700b.jpg"></p>`
		case n <= 25:
			content = `
<p>Warm kitty, soft kitty</p>
<p><img src="http://placekitten.com/g/200/300"></p>`
		case n <= 75:
			content = `
<p>Husky instead?</p>
<p><img src="/__pin/?query=husky"></p>`
		default:
			content = `
<p>Why not enjoy this nice picture from reddit instead?</p>
<p><img src="/__reddit/"></p>`
	}

	fmt.Fprintf(w, `
<html><head>
<title>No %s</title>
<style>
body { background-color: black; color: white; }
img { max-width:1000px; max-height:500px;}</style>
</head>
<body>
<center>
<h1>No need to procrastinate on <em>%s</em></h1>
%s
</center>
</body></html>`, domain, domain, content)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Request to ", r.Host, r.URL.String())

	redirectUrl, ok := REDIRECTS[r.Host]
	if ok {
		http.Redirect(w, r, redirectUrl, http.StatusFound)
		log.Print(redirectUrl)
		return
	}

	isAntiProcrastination := false
	for _, host := range ANTI_PROCRASTINATION {
		if host == r.Host || strings.HasSuffix(r.Host, "." + host) {
			isAntiProcrastination = true
			break
		}
	}

	if isAntiProcrastination {
		antiProcrastinationHandler(w, r)
		return
	}

	message := "Uknown address "
	if ! r.URL.IsAbs() {
		message = message + r.Host
	}
	message = message + r.URL.String()
	http.Error(w, message,  http.StatusNotFound)
}

// Webscrape pinterest for a random image
func pinterestHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if len(query) == 0 {
		log.Print("Missing query")
		http.NotFound(w, r)
		return
	}

	log.Printf("Searching pinterest for %q", query)

	url := "https://www.pinterest.com/search/?q=" + query
	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Can't get response from %q: %v", url, err)
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("Can't read pinterest response: %v", err)
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}

	images := []string{}
	pinsRegexp := regexp.MustCompile("<img src=\"(.*?)\".*?class=\"pinImg")
	matches := pinsRegexp.FindAllStringSubmatch(string(contents), -1)
	for _, match := range matches {
		images = append(images, match[1])
	}

	finalImage := images[rand.Intn(len(images))]
	log.Print("Chosen image: ", finalImage)
	http.Redirect(w, r, finalImage, http.StatusFound)
}


type ReditListing struct {
	Data struct {
		Children []struct {
			Data struct {
				Url string
			}
		}
	}
}

// Debug function to show all images
func dumpImageList(w http.ResponseWriter, name string, images []string) {
	fmt.Fprintf(w, "<html><style>img{max-width:1000px;max-height:500px;}</style>")
	fmt.Fprintf(w, "<h1>%s</h1>", name)
	for _, image := range images {
		fmt.Fprintf(w, "<p><img src='%s'>", image)
	}
}

func redditHandler(w http.ResponseWriter, r *http.Request) {
	REDDITS := []string{
		"Pictures",
		"funny",
		"EarthPorn",
		"CityPorn",
		"AnimalPorn",
		"itookapicture",
	}

	reddit := REDDITS[rand.Intn(len(REDDITS))]
	url := "http://www.reddit.com/r/" + reddit + "/top.json"
	log.Print("Fetching top reddit from ", url)
	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Can't get response from %q: %v", url, err)
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("Can't read reddit response: %v", err)
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}

	listing := ReditListing{}
	err = json.Unmarshal(contents, &listing)
	if err != nil {
		msg := fmt.Sprintf("Invalid reddit JSON: %v", err)
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	images := []string{}
	for _, entry := range listing.Data.Children {
		image := entry.Data.Url
		// Find direct link to imgur image
		imgurRegexp := regexp.MustCompile("^https?://(m.)?imgur.com/(gallery/)?([^/]+)$")
		image = imgurRegexp.ReplaceAllString(image, "http://i.imgur.com/$3.jpg")

		isImage := (
			strings.HasSuffix(image, ".jpg") ||
			strings.HasSuffix(image, ".jpeg") ||
			strings.HasSuffix(image, ".png") ||
			strings.HasSuffix(image, ".gif"))

		// Ignore links to main imgur page
		if strings.HasPrefix(image, "http://i.imgur.com") || isImage  {
			images = append(images, image)
		} else {
			log.Print("Rejecting picture URL ", image)
		}
	}

	if len(images) == 0 {
		msg := "No images found"
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	finalImage := images[rand.Intn(len(images))]
	log.Print("Chosen image: ", finalImage)
	http.Redirect(w, r, finalImage, http.StatusFound)
}

func main() {
	wantsInstall := flag.Bool("install", false, "Install server on the machine")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	if *wantsInstall {
		install()
	} else {
		http.HandleFunc("/", requestHandler)
		http.HandleFunc("/__pin/", pinterestHandler)
		http.HandleFunc("/__reddit/", redditHandler)
		log.Print("Starting server")
		log.Fatal(http.ListenAndServe(LISTEN_IP + ":443", nil))
	}
}
