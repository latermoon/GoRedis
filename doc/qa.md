如果Redis从库同步速度太慢，“overcoming of output buffer limits”，Master会断开连接。

[16815] 07 Oct 18:23:27.121 * Slave ask for synchronization
[16815] 07 Oct 18:23:27.121 * Starting BGSAVE for SYNC
[16815] 07 Oct 18:23:27.259 * Background saving started by pid 9437
[16815] 07 Oct 18:24:10.091 * Background saving terminated with success
[16815] 07 Oct 18:29:24.064 # Client addr=10.80.102.91:26127 fd=725 name= age=357 idle=357 flags=S db=0 sub=0 psub=0 multi=-1 qbuf=0 qbuf-free=0 obl=16376 oll=3195 omem=80765752 events=rw cmd=sync scheduled to be closed ASAP for overcoming of output buffer limits.
[16815] 07 Oct 18:30:01.514 * Background saving started by pid 10630
[16815] 07 Oct 18:30:45.858 * Background saving terminated with success