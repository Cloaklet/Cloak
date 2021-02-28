FROM ubuntu:bionic
RUN apt update && \
    apt install --no-recommends -y libappindicator3-dev gcc libgtk-3-dev libxapp-dev