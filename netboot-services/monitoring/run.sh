#!/bin/bash
set -e

allSet=true

# The shellcheck disable directives are used, as we expect those variables to be replaced by the pipeline. If they are not replaced, we run telegraf in a --test mode, which causes the container to fail (and be restarted by docker-compose)
# shellcheck disable=SC2154
if [[ "$datadog_secret" == "#{datadog_secret}#" ]]
then
  echo "Please set the variable datadog_secret"
  allSet=false
fi

if [[ "$allSet" == true ]]
then
  exec telegraf --config /etc/telegraf/telegraf.conf
else
  exec telegraf --config /etc/telegraf/telegraf.conf --test
fi
