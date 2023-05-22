#!/bin/bash

helm package fate-exchange
helm package fate

base64 -i fate-exchange-v*.tgz > fate-exchange-base64.txt
base64 -i fate-v*.tgz > fate-base64.txt
