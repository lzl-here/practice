# 订单系统

每日订单增量十几二十万，根据订单id做了分库分表
业务上指定最多只能查到6个月之内的数据，6 * 30 * 15w = 2700w，现在MySQL热库的数据量在2-3千万左右
6个月之前的数据全都归档到冷库中
- 冷库采用什么？
- MySQL？HBase?
- 怎么同步的？

根据userID进行订单的分表
- 对于用户：提供分页查看自己订单的接口（在某个店铺下，所有店铺下），因为根据userID分片，直接命中
- 对于商家：提供多维度的复杂查询，比如,使用es查询
- 


业内其他做法：
- 基因法：把userID后n位拼接到订单后，根据这后n位查询，这样就能根据orderID和userID查询
- 路由表：根据orderID分片，建立 userID -> orderID 的映射关系，然后根据userID查询出orderID，再去对应的订单表查 （性能慢）
- 冗余存储：根据orderID分片，冗余存储一份全量数据，根据userID分片，（性能好，存储成本高，数据一致性保证更复杂）

更加复杂的聚合查询，如：报表数据，用户画像等等，通过大数据部门提供接口，我们这做数据的聚合汇总，和前台交互



## 大商家

标记法特库特表

一，不以商户id routing，改为时间。业务上，查一个月或特定时间段单据本来就是常见的
二，routing配置routing_partition_size。
三，shard个数设计为素数个




