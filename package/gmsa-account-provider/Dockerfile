FROM registry.suse.com/bci/bci-micro:15.4.17.1
ARG OS
ARG ARCH

COPY ./dist/gmsa-account-provider-linux-$ARCH ./

ENTRYPOINT ["gmsa-account-provider-linux-$ARCH"]
