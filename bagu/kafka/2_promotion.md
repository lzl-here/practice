# 进阶

## 高可用
副本
每个leader都维护一个ISR列表，存储了基本和leader同步的follower



request.required.asked: 
- 0: 单向发送
- 1: leader 不等follower同步就ack
- -1: 消息给leader后，ISR列表中的follwer全部同步完再ack (ISR中有两个节点才能保证数据不丢失,也就是需要副本)


## 脑裂



## 高性能

- 顺序写
- 刷盘策略: 异步刷盘/同步刷盘
- 消息缓存，批量发送
- 零拷贝: sendfile, mmap
- 消息压缩


## rebanlance

## 水位
