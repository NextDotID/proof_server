(ns pod.jaydeesimon.jsoup
  (:refer-clojure :exclude [read read-string])
  (:require [bencode.core :as bencode]
            [clojure.edn :as edn])
  (:import (org.jsoup Jsoup)
           (java.io PushbackInputStream)
           (org.jsoup.nodes Element Attribute))
  (:gen-class))

(set! *warn-on-reflection* true)

(def stdin (PushbackInputStream. System/in))

(defn write [v]
  (bencode/write-bencode System/out v)
  (.flush System/out))

(defn read-string [^"[B" v]
  (String. v))

(defn read []
  (bencode/read-bencode stdin))

(defn select [html css-query]
  (let [elements (-> (Jsoup/parse html)
                     (.select ^String css-query))]
    (map (fn [element]
           {:id (.id ^Element element)
            :class-names (.classNames ^Element element)
            :tag-name (.normalName ^Element element)
            :attrs (->> (.attributes ^Element element)
                        .iterator
                        iterator-seq
                        (map (juxt (memfn ^Attribute getKey) (memfn ^Attribute getValue)))
                        (into {}))
            :own-text (.ownText ^Element element)
            :text (.text ^Element element)
            :whole-text (.wholeText ^Element element)
            :inner-html (.html ^Element element)
            :outer-html (.outerHtml ^Element element)})
         elements)))

(def lookup
  {'pod.jaydeesimon.jsoup/select select})

;; Copied from https://github.com/babashka/pod-babashka-hsqldb/blob/master/src/pod/babashka/hsqldb.clj#L33
(defn -main [& _args]
  (loop []
    (let [message (try (read)
                       (catch java.io.EOFException _
                         ::EOF))]
      (when-not (identical? ::EOF message)
        (let [op (get message "op")
              op (read-string op)
              op (keyword op)
              id (some-> (get message "id")
                         read-string)
              id (or id "unknown")]
          (case op
            :describe (do (write {"format" "edn"
                                  "namespaces" [{"name" "pod.jaydeesimon.jsoup"
                                                 "vars" [{"name" "select"}]}]
                                  "id" id
                                  "ops" {"shutdown" {}}})
                          (recur))
            :invoke (do (try
                          (let [var  (-> (get message "var")
                                         read-string
                                         symbol)
                                args (get message "args")
                                args (read-string args)
                                args (edn/read-string args)]
                            (if-let [f (lookup var)]
                              (let [value (pr-str (apply f args))
                                    reply {"value" value
                                           "id" id
                                           "status" ["done"]}]
                                (write reply))
                              (throw (ex-info (str "Var not found: " var) {}))))
                          (catch Throwable e
                            (binding [*out* *err*]
                              (println e))
                            (let [reply {"ex-message" (ex-message e)
                                         "ex-data" (pr-str
                                                     (assoc (ex-data e)
                                                       :type (class e)))
                                         "id" id
                                         "status" ["done" "error"]}]
                              (write reply))))
                        (recur))
            :shutdown (System/exit 0)
            (recur)))))))

(comment

  ;; Run these commands in Babashka
  (require '[babashka.pods :as pods])

  ;; for the uberjar
  (pods/load-pod ["java" "-jar" "target/uberjar/pod-jaydeesimon-jsoup-0.1-standalone.jar"])

  ;; for the graalvm compiled binary
  (pods/load-pod "./pod-jaydeesimon-jsoup")

  (require '[pod.jaydeesimon.jsoup :as jsoup])

  (-> (curl/get "https://clojure.org")
      :body
      (jsoup/select "div.clj-header-message")
      first
      :text)

  )
