# 集合

## hashmap

- 链表什么时候树化？
jdk1.8开始才有树化: 
    1. 链表的长度大于8
    2. hashmap的容量大于等于64

如果树的元素少于6，树会退化为链表

- 为什么hashmap的容量都是2的幂？

因为计算元素放在哪个槽，需要对元素的hashcode进行取模，如果容量是2的幂的话，计算取模直接 hashcode & (cap - 1)也能算出位置，并且性能比直接取模更好

- hashmap的扩容

当存放的元素 / 容量 > 负载因子时发生扩容
创建一个新数组，容量是老数组的两倍，
jdk1.8之前: 将元素全部rehash然后放到新位置
jdk1.8之后: 将链表低位留在原地，高位直接迁移到 旧index + 旧cap 的位置，减少了重新计算slot的开销


- 解决hash冲突的方法

  1. 拉链法: 像hashmap这样冲突了把元素挂在一个槽下面
  2. 重hash法：负载因子太高，进行扩容并且重新计算元素的hashcode (ps: hashmap在jdk1.8之后不用重新计算slot位置、redis的哈希表采用渐进式hash，防止长时间阻塞)
  3. 开放寻址法


- hashmap的并发安全问题

1.7 死循环
1.8 
TODO

- concurrentHashMap
1.7时，采用分段锁，多个slot对应一个segment
1.8 对每个slot用synchronized上锁，插入空slot时采用cas优化


