FROM openjdk:27-ea-trixie

RUN useradd -m -u 1001 runner \
 && mkdir -p /work \
 && chown runner:runner /work

USER runner
WORKDIR /work