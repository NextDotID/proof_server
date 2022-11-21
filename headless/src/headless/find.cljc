 #?(:bb (do
          (println "Running in bb")
          (require '[babashka.pods :as pods])
          (pods/load-pod ["java" "-jar" "dist/jar/pod-jaydeesimon-jsoup-0.1-standalone.jar"])
          )
        :clj (import (org.jsoup Jsoup)
                     (org.jsoup.select Elements)
                     (org.jsoup.nodes Element)
                     (java.lang NullPointerException)))
(ns headless.find
  (:gen-class) 
  (:require
   [babashka.cli :as cli]
   [etaoin.api :as e] ;; WebDriver API
   [ruuter.core :as ruuter] ;; HTTP router
   [cheshire.core :as json]
   [org.httpkit.server :as http]
   [org.httpkit.client :refer [get]]
   [clojure.core.match :refer [match]] ;; Pattern Matching
   [clojure.java.io :as io]))

;; Debug variables
;; (set! *warn-on-reflection* true)

;; JSOUP Interop 
#?(:bb nil
   :clj (defn get-page [l]
  (.get (Jsoup/connect l))))
#?(:bb nil
   :clj (defn get-elems [page css]
  (.select page css)))

;; Default driver opts (Currently only chrome/chromium)
(def chrome-driver-opts {:capabilities {:chromeOptions {:args ["--headless" "--no-sandbox"]}}})

;; TODO: refactor strategies to work independent of the driver implementation
;;   - Example driver with driver opts used as defaults
;;     (defn start-default-chrome [& args] (e/chrome chrome-driver-opts) )

;; Finder API Strategies - Static
(def static-css-strategy
  #?(:bb (fn [url selector]
           (try (-> 
                 (-> @(get url) :body)
                 (pod.jaydeesimon.jsoup/select selector)
                 first
                 :text)
                (catch Exception _ (do (println (str "Error - Failed to get url -> " url)) false))))
     :clj (fn [url selector]
    (try (let [html  (get-page url)
               el (first (get-elems html selector))
               res (if (nil? el) false (.text el))]
           res)
         (catch Exception e (do (println (str "Error - Failed to get url -> " url)) false)))))) ;; TODO: Allow for custom Element attribute after selector instead of only Element.text

(def static-regex-strategy
  (fn [url regex]
    (let [src (-> @(get url) :body)
          ptrn (re-pattern regex)
          res (try (re-find ptrn src) (catch NullPointerException e false))]
      (if (nil? res) false res))))

;; Finder API Strategies - WebDriver
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

;; Finder API Strategies - Fallback (static first, webdriver last)
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
      (= secret (webdriver-css-strategy url matcher)))))

;; Option Validation and Dispatch
(defn dispatch [opts]
  ;; Option Validation
  (when (or
         (= "" (opts :url)) (nil? (opts :url))
         (= "" (opts :matcher)) (nil? (opts :matcher))
         (= "" (opts :match-by)) (nil? (opts :match-by)))
    {:status 400 :body ((if (opts :batch?) identity json/generate-string) {:match false :message (str "[error] url, matcher, and match-by must all contain valid values.")})})
  ;; Pattern Matching Dispatch
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
                 :else "Finder: Invalid configuration, unknown.")]
    ;; Result Coercion
    (if (string? result)
      {:status 400 :body ((if batch? identity json/generate-string) {:match false :message (str "[error] " result)})}
      {:status 200 :body ((if batch? identity json/generate-string) {:match result :message (if result "match found" "no match")})})))

;; Batch vars and helpers
(def default-batch-size 5)
(defn safe-parse-int [s d]
  (try (Integer/parseInt s)
       (catch Exception _ d)))

;; Modified dispatch and find handlers for batch processing
(defn batch-dispatch [opts]
  (:body (dispatch (merge opts {:batch? true}))))
(defn find-batch [{:keys [batch batch-size] :or {batch [] batch-size default-batch-size}}] ;; Allow specifying batch-size, and integrate browser instance pooling 
  (let [;; Validate batch-size arg
        bs (if (int? batch-size) batch-size (safe-parse-int batch-size default-batch-size))
         ;; Start parallel batches from single thread, maybe parallelizable with fold? 
        proc (flatten (transduce
                       (comp
                        (partition-all bs)
                        (map (fn [v]
                               (doall (pmap batch-dispatch v)))))
                       conj
                       batch))
        ;; Stringify result to JSON
        body (json/generate-string proc)]
    ;; HTTP Response
    {:status 200
     :body body}))

;; Find api CLI & HTTP entrypoint
(defn find [opts]
  (try
    (if (contains? opts :batch) (find-batch opts) (dispatch opts))
    (catch Exception e
      (println (str "[DEBUG] find crash, error -> " (pr-str e)))
      {:status 500 :body "Internal Server Error."})))

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
                    find))}])

;; Server entrypoint
(defn start-server [{:keys [port address] :or {port 8080 address "0.0.0.0"}}]
  (http/run-server
   #(ruuter/route routes %)
   {:port (int port)
    :address address})
  (println "Server started on " address ":" port)
  @(promise) ;; Hold thread for HTTP server
  )

;; Lambda entrypoint
(defn lambda [{:keys [event ctx]}]
  (println (str "Event " (pr-str event) "\nCtx " (pr-str ctx))) 
  {:statusCode 200
   :headers {"content-type" "application/json"}
   :body (:body (headless.find/find event))})

;; Main (manifold) entrypoint 
(defn -main [& args]
  (when-not (nil? (first args))
    (try
      (if (or (contains? (first args) :cli) (string? (first args)))
        (let [cli-arg (if (contains? (first args) :cli) ((first args) :cli) (cli/parse-opts args))]
          (if (contains? cli-arg :server)
            (start-server cli-arg)
            (:body (find cli-arg))))
        (lambda (first args)))
      (catch Exception e
        (println (str "[DEBUG] main crash, error -> " (pr-str e)))
        (if (string? (first args))
          {:status 500 :body (str "Internal Server Error." (pr-str e))}
          "CLI Internal Error.")))))

#?(:bb (-main {:cli (cli/parse-opts *command-line-args*)}))