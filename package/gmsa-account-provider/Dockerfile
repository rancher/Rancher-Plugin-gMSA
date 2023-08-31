FROM registry.suse.com/bci/bci-micro:15.4.17.1
ARG OS
ARG ARCH

COPY ./dist/artifacts/gmsa-account-provider-$OS-$ARCH ./

ENTRYPOINT ["gmsa-account-provider-$OS-$ARCH"]
