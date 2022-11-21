(ns build
  (:require
   [clojure.tools.build.api :as b]))

(def class-dir ".holy-lambda/target/classes")
(def basis (b/create-basis {:project "deps.edn"}))

(defn clean [_]
  (b/delete {:path ".holy-lambda/target"})
  (b/delete {:path ".holy-lambda/build"}))


;; Compile clojure code into standalone jar 
(defn uberjar [_]
  (b/copy-dir {:src-dirs ["src" "resources"]
               :target-dir class-dir})
  (b/compile-clj {:basis basis
                  :src-dirs ["src"]
                  :ns-compile ['headless.find] ;; Only package this namespace
                  :class-dir class-dir})
  (b/uber {:class-dir class-dir
           :main 'headless.find
           :basis basis
           :uber-file ".holy-lambda/build/output.jar"}))