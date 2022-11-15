FROM fedora:latest AS builder

ARG TARGETARCH

RUN yum install gcc glibc-devel zlib-devel libstdc++-static curl git -y
RUN yum install zlib-devel zlib-static -y
RUN yum install glibc-static -y

# Install Enterprise GraalVM and Dependencies (needed to compile native image with JNI invocations)
RUN cd / && mkdir graal

COPY resources/graalvm-ee-java19-${TARGETARCH}.tar.gz /graal/graal.tar
COPY resources/graal-config /graal/config

WORKDIR /graal

RUN tar -xzf graal.tar && rm graal.tar

ENV PATH="${PATH}:/graal/graalvm-ee-java19-22.3.0/bin"
ENV JAVA_HOME="/graal/graalvm-ee-java19-22.3.0"

RUN gu install native-image -A --config /graal/config

COPY resources/pod-jaydeesimon-jsoup-0.1-standalone.jar jsoup-pod.jar
RUN native-image -jar jsoup-pod.jar -H:Name=finder --static -H:+ReportExceptionStackTraces -J-Dclojure.spec.skip-macros=true -J-Dclojure.compiler.direct-linking=true -H:Log=registerResource: --initialize-at-build-time -H:EnableURLProtocols=http,https --verbose --allow-incomplete-classpath --no-fallback --no-server -J-Xmx8g

FROM debian:stretch AS runner
USER root

# Install chromedriver
RUN apt-get update && apt install chromium-driver curl -y

# Install script interpreter
RUN curl -sLO https://raw.githubusercontent.com/babashka/babashka/master/install
RUN chmod a+x install
RUN ["/bin/bash", "./install", "--static"]

COPY --from=builder /graal /graal
RUN mkdir babashka
WORKDIR /babashka

COPY src/find.clj src/find.clj
COPY --from=builder /graal/finder bb/finder
COPY bb.edn bb.edn
ENV PATH="${PATH}:/graal/graalvm-ee-java19-22.3.0/bin"
ENV JAVA_HOME="/graal/graalvm-ee-java19-22.3.0"

RUN bb prepare
WORKDIR /babashka
ENTRYPOINT ["bb", "./src/find.clj"]
