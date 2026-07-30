FROM registry.suse.com/bci/bci-micro:15.6.54.14
ARG OS
ARG ARCH

COPY ./dist/ccg-plugin-installer-linux-$ARCH ./

ENTRYPOINT ["ccg-plugin-installer-linux-$ARCH"]
