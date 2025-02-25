# 杂项

## count

### count原理

count(1)：
不会返回字段给服务层，而是返回1

count(*)：
把*翻译成0，返回0给服务层

count(id):
扫描聚簇索引

count(name):
统计非NULL的字段，效率最差，建了索引遍历索引，没建遍历聚簇索引（这种情况最慢）



如果有二级索引，count遍历二级索引来求count
如果没有才遍历聚簇索引
因为聚簇索引还要存放行数据，一个节点大小就是一页，所以聚簇索引的节点更多，需要更多次的io，速度更慢
MySQL对count(1)和count(*)有优化：如果有多个二级索引的时候，优化器会使用key_len 最小的二级索引进行扫描

### MyISAM引擎

MyISAM引擎每张表都维护了总行数，count(*)直接读取
而InnoDB不支持统计行数，因为InnoDB需要支持事务，对于每个事务读到的数据都可能不一样，count出的结果都不一样，像myisam这样操作count出的数值会是脏数据

### 怎么优化count？

explain中可以统计一个模糊值，对于不需要精确数据的场景可以这样做，效率很高

冗余维护一份count数据


## sql的执行流程

读请求：
1. 建立tcp连接
2. 发送sql语句
3. 计算sql哈希值，尝试走查询缓存。命中直接返回结果
4. 语法解析器解析语法
5. 预处理：*转化为全部字段、检查字段是否存在、检查表是否存在....
6. 优化器优化执行计划
7. 执行器执行sql，从存储引擎读取数据，返回给客户端


## 数据类型

varchar：变长字符串类型，大小限制在2^16-1字节以内，varchar(n)中的n表示字符串长度，所以varchar能存多少字符取决于编码


## InnoDB和MyISAM
- InnoDB支持聚簇和非聚簇索引，MyISAM所有数据都单独存储，MyISAM的索引叶子存储的是数据地址，回查比InnoDB要高效
- InnoDB支持事务，MyISAM不支持
- InnoDB支持行锁，基于索引实现的，MyISAM不支持，只能锁表, InnoDB的写入性能更快
- InnoDB有自己的内存管理机制：buffer pool，对内存的利用更多，而MyISAM在这一块很少
- InnoDB支持redolog和undolog，MyISAM只支持binlog（服务层）
- 因为redolog，InnoDB具备容灾恢复能力，MyISAM不具备
- InnoDB支持外键，MyISAM不支持
- MyISAM表存了总数，InnoDB因为需要支持多版本快照读（MVCC），无法存总数，count需要实时计算，而MyISAM直接读取






