# Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/imxw/miniokr.

ARG BASE_IMAGE=debian:latest
FROM ${BASE_IMAGE}
LABEL maintainer="<ixw1991@126.com>"

WORKDIR /opt/miniokr


# 更新 CA 证书
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    mkdir -p /var/log/miniokr

COPY miniokr /opt/miniokr/bin/
COPY frontend /opt/miniokr/frontend

ENTRYPOINT ["/opt/miniokr/bin/miniokr"]
CMD ["-c", "/etc/miniokr/miniokr.yaml"]
