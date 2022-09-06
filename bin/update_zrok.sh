#!/bin/bash

ssh -i ~/.ssh/nf-zrok-ubuntu ctrl-01.zrok.io sudo systemctl stop zrok-ctrl
scp -i ~/.ssh/nf-zrok-ubuntu ~/local/zrok/bin/zrok ctrl-01.zrok.io:local/zrok/bin/zrok
ssh -i ~/.ssh/nf-zrok-ubuntu ctrl-01.zrok.io sudo systemctl start zrok-ctrl

ssh -i ~/.ssh/nf-zrok-ubuntu in-01.zrok.io sudo systemctl stop zrok-http-frontend
scp -i ~/.ssh/nf-zrok-ubuntu ~/local/zrok/bin/zrok in-01.zrok.io:local/zrok/bin/zrok-ctrl
ssh -i ~/.ssh/nf-zrok-ubuntu in-01.zrok.io sudo systemctl start zrok-http-frontend