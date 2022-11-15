(ns find (:require [babashka.pods :as pods]
                   [babashka.cli :as cli]
                   [etaoin.api :as e] ;; WebDriver API
                   [ruuter.core :as ruuter] ;; HTTP router
                   [cheshire.core :as json]
                   [org.httpkit.server :as http]
                   [org.httpkit.client :refer [get]]
                   [clojure.core.match :refer [match]] ;; Pattern Matching
                   [clojure.java.io :as io]))

(pods/load-pod "./bb/finder")
(require '[pod.jaydeesimon.jsoup :as jsoup]) ;; jsoup css selectors library

(def CSS jsoup/select)

(def chrome-driver-opts {:capabilities {:chromeOptions {:args ["--headless" "--no-sandbox"]}}})


;; Finder API Strategies
(def static-css-strategy
  (fn [url selector]
    (-> @(get url)
        :body
        (CSS selector)
        first
        :text)))

(def static-regex-strategy
  (fn [url regex]
    (let [src (-> @(get url) :body)
          ptrn (re-pattern regex)]
          (re-find ptrn src))))

(def webdriver-js-strategy
  (fn [url js]
    (e/with-driver :chrome chrome-driver-opts driver
      (e/go driver url)
      (e/js-execute driver js))))

(def webdriver-xpath-strategy
  (fn [url xpath]
    (e/with-driver :chrome chrome-driver-opts driver
      (e/go driver url)
      (e/get-element-text driver {:xpath xpath}))))

(def webdriver-css-strategy
  (fn [url selector]
    (e/with-driver :chrome chrome-driver-opts driver
      (e/go driver url)
      (e/get-element-text driver {:css selector}))))

(def webdriver-regex-strategy
  (fn [url regex]
    (e/with-driver :chrome chrome-driver-opts driver
      (e/go driver url)
      (let [src (e/get-source driver)
            ptrn (re-pattern regex)]
        (re-find ptrn src)))))

;; Fallback Strategies (static first, webdriver last)
(defn fallback-regex
  ([url matcher] (let [static-res? (boolean (static-regex-strategy url matcher))]
                   (if static-res?
                     true
                     (boolean (webdriver-regex-strategy url matcher)))))
  ([url matcher secret] (let [static-res? (= secret (static-regex-strategy url matcher))]
                          (if static-res?
                            true
                            (= secret (webdriver-regex-strategy url matcher))))))

(defn fallback-css [url matcher secret]
  (let [static-res? (= secret (static-css-strategy url matcher))]
    (if static-res?
      true
      (= secret (webdriver-css-strategy url matcher)))
    ))

;; Option Validation and Strategy Dispatch
(defn dispatch [opts]
  (when (or
         (= "" (opts :url)) (nil? (opts :url))
         (= "" (opts :matcher)) (nil? (opts :matcher))
         (= "" (opts :match-by)) (nil? (opts :match-by)))
    {:status 400 :body ((if (opts :batch?) identity json/generate-string) {:match false :message (str "[error] url, matcher, and match-by must all contain valid values.")})})
  (let [{:keys [url matcher secret batch?]} opts
        result (match [(merge opts (if (nil? secret) {:secret ""} {}))]
                      [{:strategy "webdriver" :match-by "js" :secret ""}] (boolean (webdriver-js-strategy url matcher))
                      [{:strategy "fallback" :match-by "regex" :secret ""}] (fallback-regex url matcher)
                      [{:strategy "static" :match-by "regex" :secret ""}] (boolean (static-regex-strategy url matcher))
                      [{:strategy "webdriver" :match-by "regex" :secret ""}] (boolean (webdriver-regex-strategy url matcher))
                      [{:secret ""}] "Finder: Invalid configuration, secret must be non empty"
                      [{:strategy "fallback" :match-by "css"}] (fallback-css url matcher secret)
                      [{:strategy "fallback" :match-by "regex"}] (fallback-regex url matcher secret)
                      [{:strategy "static" :match-by "regex"}] (= secret (static-regex-strategy url matcher))
                      [{:strategy "static" :match-by "css"}] (= secret (static-css-strategy url matcher))
                      [{:strategy "webdriver" :match-by "regex"}] (= secret (webdriver-regex-strategy url matcher))
                      [{:strategy "webdriver" :match-by "css"}] (= secret (webdriver-css-strategy url matcher))
                      [{:strategy "webdriver" :match-by "xpath"}] (= secret (webdriver-xpath-strategy url matcher))
                      [{:strategy "webdriver" :match-by "js"}] (= secret (webdriver-js-strategy url matcher))
                      :else "Finder: Invalid configuration, unknown." )]
        (if (string? result)
          {:status 400 :body ((if batch? identity json/generate-string) {:match false :message (str "[error] " result)})}
          {:status 200 :body ((if batch? identity json/generate-string) {:match result :message (if result "match found" "no match")})})
    ))

;; Modified dispatch and find handlers for batch processing
(defn batch-dispatch [opts]
  (:body (dispatch (merge opts {:batch? true}))))

(defn find-batch [opt-arr]
 { :status 200 :body (json/generate-string (doall (pmap batch-dispatch opt-arr)))})

;; Find api CLI & HTTP entrypoint
(defn find [opts] (if (contains? opts :batch) (find-batch (opts :batch)) (dispatch opts)))


;; HTTP Request middleware
(defn json->edn [reader] (json/parse-stream reader true))
(defn parse-json-middleware [request]
  (-> request
      :body
      io/reader
      json->edn))

;; HTTP routes
(def routes
  [{:path "/health"
    :method :get
    :response {:status  200
               :headers {"Content-Type" "text/html"}
               :body "Ok"}}
   {:path "/v1/find"
    :method :get
    :response (fn [req]
                (-> req
                    parse-json-middleware
                    find))
    }])

;; CLI Server entrypoint
(defn start-server [{:keys [port address] :or {port 8080 address "0.0.0.0"}}]
  (http/run-server
   #(ruuter/route routes %)
   {:port (int port)
    :address address})
  (println "Server started on " address ":" port)
  @(promise))

;; Main entrypoint
(defn -main [args] (let [cli-arg (cli/parse-opts args)]
                       (if (contains? cli-arg :server)
                         (start-server cli-arg)
                         (:body (find cli-arg)))))
(-main *command-line-args*)


