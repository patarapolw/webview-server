FROM karalabe/xgo-latest
RUN apt-get update
RUN apt-get install -y build-essential libgtk-3-dev libwebkit2gtk-4.0-dev