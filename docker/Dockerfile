FROM ubuntu:16.04

# 安装基础依赖
RUN apt-get clean && apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    tzdata

# 修改时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /gitbot

COPY bot /gitbot
COPY preset_labels.json /gitbot

ENTRYPOINT ["/gitbot/bot", "webhook"]
