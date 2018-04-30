FROM alpine

USER 1001
COPY openshift-lb-controller .
ENTRYPOINT ["./openshift-lb-controller"]
