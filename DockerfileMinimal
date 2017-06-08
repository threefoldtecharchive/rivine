FROM scratch

ARG binaries_location=dist/linux

COPY $binaries_location/rivined /rivined
COPY $binaries_location/rivinec /rivinec

EXPOSE 23112

ENTRYPOINT ["/rivined"]
