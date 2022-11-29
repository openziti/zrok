#!/bin/bash

# ctrl-01.zrok.io
ssh -i ~/.ssh/nf-zrok-ubuntu ctrl-01.dev.zrok.io sudo systemctl stop zrok-ctrl
scp -i ~/.ssh/nf-zrok-ubuntu ~/local/zrok/bin/zrok ctrl-01.dev.zrok.io:local/zrok/bin/zrok
ssh -i ~/.ssh/nf-zrok-ubuntu ctrl-01.dev.zrok.io sudo systemctl start zrok-ctrl

# ctrl-02.zrok.io
ssh -i ~/.ssh/nf-zrok-ubuntu ctrl-02.dev.zrok.io sudo systemctl stop zrok-ctrl
scp -i ~/.ssh/nf-zrok-ubuntu ~/local/zrok/bin/zrok ctrl-02.dev.zrok.io:local/zrok/bin/zrok
ssh -i ~/.ssh/nf-zrok-ubuntu ctrl-02.dev.zrok.io sudo systemctl start zrok-ctrl

# in-01.zrok.io
ssh -i ~/.ssh/nf-zrok-ubuntu in-01.dev.zrok.io sudo systemctl stop zrok-http-frontend
scp -i ~/.ssh/nf-zrok-ubuntu ~/local/zrok/bin/zrok in-01.dev.zrok.io:local/zrok/bin/zrok
ssh -i ~/.ssh/nf-zrok-ubuntu in-01.dev.zrok.io sudo systemctl start zrok-http-frontend

# in-02.zrok.io
ssh -i ~/.ssh/nf-zrok-ubuntu in-02.dev.zrok.io sudo systemctl stop zrok-http-frontend
scp -i ~/.ssh/nf-zrok-ubuntu ~/local/zrok/bin/zrok in-02.dev.zrok.io:local/zrok/bin/zrok
ssh -i ~/.ssh/nf-zrok-ubuntu in-02.dev.zrok.io sudo systemctl start zrok-http-frontend
