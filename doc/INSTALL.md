
### install go

	wget latermoon:8000/go1.1.2.linux-amd64.tar.gz
	tar zxvf go1.1.2.linux-amd64.tar.gz
	mkdir /home/server/gopath

	vi /etc/profile
	export PATH=$PATH:/home/server/go/bin
	export GOROOT=/home/server/go/
	export GOPATH=/home/server/gopath/

### git

	yum install git
	git clone -b GoRedisDev https://github.com/latermoon/GoRedis.git GoRedisDev
	git fetch

	yum install hg

### gorocks

	yum install bzip2-devel

### run

/usr/lib64/libsnappy.so.1.1.4
/usr/lib64/libstdc++.so.6.0.13
/usr/lib64/libgflags.so.0.0.0



