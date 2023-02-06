package headless

import "github.com/go-rod/rod/lib/launcher"

// newLauncher creates a new launcher with default options.
func newLauncher(path string) *launcher.Launcher {
	if path == "" {
		var found bool

		path, found = launcher.LookPath()
		if !found {
			path = launcher.NewBrowser().MustGet()
		}
	}

	return launcher.New().
		Bin(path).
		// recommended flags to run in serverless environments
		// see https://github.com/alixaxel/chrome-aws-lambda/blob/master/source/index.ts
		Set("allow-running-insecure-content").
		Set("autoplay-policy", "user-gesture-required").
		Set("disable-component-update").
		Set("disable-domain-reliability").
		Set("disable-features", "AudioServiceOutOfProcess", "IsolateOrigins", "site-per-process").
		Set("disable-print-preview").
		Set("disable-setuid-sandbox").
		Set("disable-site-isolation-trials").
		Set("disable-speech-api").
		Set("disable-web-security").
		Set("disk-cache-size", "33554432").
		Set("enable-features", "SharedArrayBuffer").
		Set("hide-scrollbars").
		Set("ignore-gpu-blocklist").
		Set("in-process-gpu").
		Set("mute-audio").
		Set("no-default-browser-check").
		Set("no-pings").
		Set("no-sandbox").
		Set("no-zygote").
		Set("single-process").
		Set("use-gl", "swiftshader").
		Set("window-size", "1920", "1080")
}

// newLambdaLauncher creates a new launcher with default options for AWS Lambda.
func newLambdaLauncher(path string) *launcher.Launcher {
	if path == "" {
		var found bool

		path, found = launcher.LookPath()
		if !found {
			path = launcher.NewBrowser().MustGet()
		}
	}

	return launcher.New().
		Bin(path).
		// recommended flags to run in serverless environments
		// see https://github.com/alixaxel/chrome-aws-lambda/blob/master/source/index.ts
		Set("allow-running-insecure-content").
		Set("autoplay-policy", "user-gesture-required").
		Set("disable-component-update").
		Set("disable-domain-reliability").
		Set("disable-features", "AudioServiceOutOfProcess", "IsolateOrigins", "site-per-process").
		Set("disable-print-preview").
		Set("disable-setuid-sandbox").
		Set("disable-site-isolation-trials").
		Set("disable-speech-api").
		Set("disable-web-security").
		Set("disk-cache-size", "33554432").
		Set("enable-features", "SharedArrayBuffer").
		Set("hide-scrollbars").
		Set("ignore-gpu-blocklist").
		Set("in-process-gpu").
		Set("mute-audio").
		Set("no-default-browser-check").
		Set("no-pings").
		Set("no-sandbox").
		Set("no-zygote").
		Set("single-process").
		Set("use-gl", "swiftshader").
		Set("window-size", "1920", "1080")
}
