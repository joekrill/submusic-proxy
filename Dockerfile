FROM scratch
ARG TARGETPLATFORM
ENTRYPOINT ["/usr/bin/subsonic-sanitizer"]
COPY $TARGETPLATFORM/subsonic-sanitizer /usr/bin/