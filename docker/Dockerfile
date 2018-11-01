FROM ubuntu:16.04

# 安装基础依赖
RUN apt-get clean && apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    tzdata

# 修改时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ADD bot /usr/local/bin

ENTRYPOINT ["/usr/local/bin/bot", "webhook"]
