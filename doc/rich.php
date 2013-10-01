<?php


redis.richset('user:100422:profile', 'setting.mute', true)
redis.richget('user:100422:profile', 'name,age,photos,setting.mute,device.uid,device.client')

redis.richmerge('user:100422:profile', 'setting.mute')


?>