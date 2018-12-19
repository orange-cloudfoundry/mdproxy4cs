#!/bin/bash

# add ip 169.254.169.254 to loopback interface
ip addr show dev lo | grep -q 'inet 169.254.169.254/32' || {
  ip addr add 169.254.169.254/32 dev lo
}
