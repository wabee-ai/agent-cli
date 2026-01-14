# Minimal Docker image for wabee CLI
FROM scratch

COPY wabee /wabee

ENTRYPOINT ["/wabee"]
