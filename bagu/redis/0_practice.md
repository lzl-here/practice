# 实践

## 缓存雪崩

- 过期时间加上随机值，保证数据不会同一时间大量过期
- 对于短时间的大流量，让缓存时间设置长一些，比如秒杀，把数据都预热到缓存中，过期时间比活动还长，保证流量不会打到db

## 缓存穿透

- 限流
- 空key: 空key可以单独维护起来，不和正常业务数据放在一起防止影响正常业务，先查空key集合，查到了就返回空
- 布隆过滤器: 删除了db中的数据怎么办？
    - 定期生成新的布隆过滤器，数据不一致问题
    - 计数过滤器，


## 缓存击穿

- 锁 + 双重检查
- 逻辑过期：返回旧数据，异步读取新数据


## 数据一致性

### 删cache，写db
一写一读场景下会有问题： 
**先删cache，再写db**：
请求1删cache,请求2发现cache没数据，读db，把数据加载到cache，然后返回旧数据，请求1写db

**先写db，再删cache**
开始cache没数据，请求1读db，请求2写db，删cache，请求2写cache

两种方式都有不一致可能，但是第二种概率要小很多：
1. 要出现不一致：请求1删cache中间请求2读cache，然后发起db读读取，读cache是很快的
2. 请求2要在请求1把数据写入cache之前完成读取db和删cache完成，这个概率是很小的，因为读db很慢，大概率在读db之前请求2就已经把cache写入了


### 写cache，写db
无论是先写cache，再写db，还是先写db，再写cache
并发写场景下都有严重的不一致问题，不建议使用

### 延迟双删
因为cache aside策略仍然可能出现不一致问题，所以先删缓存，再写db，开个异步任务过了一段时间再删缓存，防止中间被其他请求污染缓存
异步任务开销、db访问率上升、编写困难
(感觉没啥用，一点也不现实，没见过企业这样用)

## hotkey

- 静态资源可以通过 nginx、cdn来缓存
- 热点探测：通过一些工具探测到访问频率比较高的key，或者如果提前知道了热点数据
    1. 找到后进行热点打散，拆分成hotkey_1、hotkey_2... 分片到不同redis节点上
    2. 或者存储到本地缓存，对于写的场景有数据丢失风险，还有数据不一致性问题 


美团热点解决工具：
热点机器单独部署，监控到了热key后将key转移到热点key，因为只处理热key并且机器配置一般更高，对热key的承载能力更强

超出预期的的请求进行限流、降级来保护服务器，保证服务器不完全不可用


## bigkey

### 大key坏处：
- 内存占用，可能造成节点的内存倾斜
- 网络io很慢
- 序列化对cpu消耗很大
- 操作阻塞主进程


### 解决方案:
大key拆分：
- 普通拆分，1000个元素的list，每次都要全量获取，拆成10个小list，每次mget或者pipeline获取
- 根据业务属性拆分，比如：key拼接时间，适用于对于需要按时间查找的场合

数据压缩: 把数据压缩后再写入redis，缺点是服务器压缩和解压缩消耗cpu资源


### 删除bigkey
不能使用del，会阻塞主进程，应该使用unlink命令，后台进行删除

