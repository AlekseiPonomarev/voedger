#!/usr/bin/env bash
 # Copyright (c) 2023 Sigma-Soft, Ltd.
 # @author Aleksei Ponomarev

utils_SSH_OPTS() {
  echo "-q -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o LogLevel=ERROR"
}

utils_ssh() {
  local ssh_options_string=$(utils_SSH_OPTS)

  # Split the string into an array
  IFS=' ' read -r -a ssh_options <<< "$ssh_options_string"

  # Pass options as separate arguments
  ssh "${ssh_options[@]}" "$@"
}

utils_scp() {
  local ssh_options_string=$(utils_SSH_OPTS)

  IFS=' ' read -r -a ssh_options <<< "$ssh_options_string"
  scp "${ssh_options[@]}" "$@"
}