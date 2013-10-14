yum install git
git clone -b GoRedisDev https://github.com/latermoon/GoRedis.git GoRedisDev
git fetch

vi /etc/profile
export PATH=$PATH:/home/server/go/bin
export GOROOT=/home/server/go/
export GOPATH=/home/server/gopath/

vi /etc/hosts
207.97.227.239 github.com
204.232.175.94 gist.github.com
207.97.227.243 raw.github.com
203.208.46.176  code.google.com

yum install hg

go get github.com/syndtr/goleveldb/leveldb

