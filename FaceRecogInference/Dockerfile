FROM golang

RUN apt-get -y update
RUN apt-get -y install build-essential \
					cmake \
					pkg-config \
					libx11-dev \
					libatlas-base-dev \
					libgtk-3-dev \
					libboost-python-dev \
					libjpeg-dev \
					git \
					wget \
					tar \
					&& apt-get clean \
					&& rm -rf /tmp/* /var/tmp/*

RUN cd ~ && \
	mkdir dlibInstall && \
	cd dlibInstall && \
	wget http://dlib.net/files/dlib-19.21.tar.bz2 && \
	tar xvf dlib-19.21.tar.bz2 && \
	cd dlib-19.21 && \
	mkdir build && \
	cd build && \
	cmake .. && \
	cmake --build . --config Release && \
	make install && \
	ldconfig && \
	cd ../../../

WORKDIR /app
ADD go.mod go.sum /app/
RUN go mod download
ADD . /app/
RUN go build -o facerec server.go
ENTRYPOINT ["./facerec"]