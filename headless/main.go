package main

import "os/exec"
import "fmt"
import "errors"

type FinderOptions struct {
	Url string // Required
	MatchBy string // Required, Either [ regex, css, xpath, js ]
	Matcher string // Required, grammar/code to match against
	Secret string // Optional, if using MatchBy regex or js, otherwise, required.
	Timeout uint // Optional, defaults to 10 seconds.
	Strategy string // Optional, defaults to fallback, Either [ fallback, static, webdriver ]
}



func  NewFinder(Url string, MatchBy string, Matcher string, Secret string, Timeout uint, Strategy string) (*FinderOptions, error) {
	hop := FinderOptions{Url, MatchBy, Matcher, Secret, Timeout, Strategy}
	if hop.Url == "" {
		return nil, errors.New("Url must be assigned a non-empty value.")
	}
	if hop.MatchBy == "" || (hop.MatchBy != "regex" && hop.MatchBy != "css" && hop.MatchBy != "xpath" && hop.MatchBy != "js") {
		return nil, errors.New(`
           MatchBy must be assigned a valid value.
            Valid values: [ "regex", "css", "xpath", "js" ]`)
	}
	if hop.Matcher == "" {
		return nil, errors.New("Matcher required.")
	}
	if hop.Secret == "" && hop.MatchBy != "regex" && hop.MatchBy != "js" {
		return nil, errors.New("Secret required.")
	}
	if hop.Strategy == "" || (hop.Strategy != "fallback" && hop.Strategy != "static" && hop.Strategy != "webdriver")  {
		hop.Strategy = "fallback"
	}
	if hop.Timeout == 0 {
		hop.Timeout = 5
	}
	if hop.MatchBy == "js" || hop.MatchBy == "xpath" {
		hop.Strategy = "webdriver"
	}
	return &hop, nil
}


func find(opt *FinderOptions) (string) {
	out, bberr := exec.Command("bb", "src/headless/find.cljc",
		"--url", opt.Url,
		"--match-by", opt.MatchBy,
		"--matcher", opt.Matcher,
		"--secret", opt.Secret,
		"--timeout", string(opt.Timeout),
		"--strategy", opt.Strategy).Output()
	if bberr != nil {
		fmt.Print(bberr.Error())
	}
	fmt.Print(string(out))
	return string(out) // Last line contains the JSON encoded output

}

func main() {
	finder_opts, arg_err := NewFinder("https://clojure.org", "css", "div.clj-header-message", "Clojure is a robust, practical, and fast programming language with a set of useful features that together form a simple, coherent, and powerful tool.", 10, "")
	if arg_err != nil {
		fmt.Print(arg_err.Error())
	}
	
	output := find(finder_opts)

}
