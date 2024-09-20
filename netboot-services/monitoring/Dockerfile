FROM telegraf:1.32.0

RUN export DEBIAN_FRONTEND=noninteractive && apt-get -qq update && apt-get -qq upgrade -y && apt-get -qq install jq atftp -y

RUN adduser --system nonroot
USER nonroot
WORKDIR /etc/telegraf

COPY monitor_serveravailability.sh /etc/telegraf/monitor_serveravailability.sh
COPY monitor_tftp.sh /etc/telegraf/monitor_tftp.sh
COPY telegraf.conf /etc/telegraf/telegraf.conf
COPY run.sh /etc/telegraf/run.sh

CMD ["/etc/telegraf/run.sh"]
